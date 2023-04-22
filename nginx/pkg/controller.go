package pkg

import (
	"context"
	v14 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informer "k8s.io/client-go/informers/core/v1"
	v1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	coreLister "k8s.io/client-go/listers/core/v1"
	netInformer "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/scale/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	"time"
)

const (
	workNum  = 5
	maxRetry = 10
)

type controller struct {
	client        kubernetes.Interface
	ingressLister netInformer.IngressLister
	serviceLister coreLister.ServiceLister
	// 该方法用于向任务队列中添加一个任务。任务可以是任何类型的对象，只要实现了 runtime.Object 接口即可。
	// 在任务添加到队列后，工作线程会从队列中取出任务并进行处理。
	queue workqueue.RateLimitingInterface
}

func (c *controller) updateService(oldObj interface{}, newObj interface{}) {
	// todo 比较annotation
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}
	// 如果新对象和老对象不相等的话，那么就说明需要更新了，把新对象放入到workqueue里面去
	c.enqueue(newObj)
}

func (c *controller) addService(obj interface{}) {
	c.enqueue(obj)
}

func (c *controller) enqueue(obj interface{}) {
	// k8s中用于生成缓存键值的函数
	// 在 Kubernetes 中，许多资源的信息都存储在内存中的缓存中，例如 Pod、Service、Endpoint 等。
	// 这些缓存数据通常以键值对的形式存储，其中键是一个字符串，值是一个 Kubernetes 对象。
	//
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.queue.Add(key)
}

// 如果Service的annotation是存在的，而且是我们所期望的，需要对它进行重建
func (c *controller) deleteIngress(obj interface{}) {
	ingress := obj.(*v12.Ingress)
	// 可以获取到ingress对应的service
	ownerReference := v13.GetControllerOf(ingress)
	if ownerReference != nil {
		return
	}
	if ownerReference.Kind != "Service" {
		return
	}
	// 因为ingres的名字和service的名字是一样的，所以我们还需要一个namespace，否则系统无法辨别
	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}

func (c *controller) Run(stopCh chan struct{}) {
	for i := 0; i < workNum; i++ {
		// 用于启动一个工作线程，并定期执行该线程中的任务
		// 然后使用 <-stopCh 语句等待停止信号，如果收到信号，则停止任务并返回 nil。
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

// 不停的从workqueue里面获取东西然后去进行处理
func (c *controller) worker() {
	// 从workqueue里面去获取东西
	for c.processNextItem() {

	}
}

func (c *controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	// 处理完成之后你需要把这个移除掉
	defer c.queue.Done(item)

	key := item.(string)

	// 得到item之后协调资源的状态
	// 常会执行一些与资源对象相关的操作，比如获取资源对象的最新状态，比较当前状态与期望状态的差异，根据差异进行调整等
	// 在任务处理完成后，需要使用 c.queue.Done(item) 方法通知任务队列任务已完成，以便队列可以接受下一个任务。
	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}

func (c *controller) syncService(key string) error {
	// 用于将资源对象的键（key）拆分为命名空间和名称两个部分。在 Kubernetes 二次开发中，
	// 该方法通常用于从任务队列中获取任务的键，并根据名称和命名空间获取资源对象。
	namespaceKey, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	// 删除
	// 根据命名空间和名称获取Service对象
	service, err := c.serviceLister.Services(namespaceKey).Get(name)
	// 如果获取到的错误是errors.IsNotFound则说明Service对象不存在
	if errors.IsNotFound(err) {
		return nil
	}
	// 否则，如果获取到的错误不是nil，则说明出现了其他错误，也直接返回
	if err != nil {
		return err
	}

	// 新增和删除
	_, ok := service.GetAnnotations()["ingress/http"]
	// c.ingressLister是一个Ingress对象的缓存列表器
	// 用于从内存中快速获取Ingress对象的信息，而无需从API Server发送请求
	//
	ingress, err := c.ingressLister.Ingresses(namespaceKey).Get(name)
	// 如果不存在这个对象，则直接返回
	if err != nil && errors.IsNotFound(err) {
		return err
	}

	// 如果service可以获取到标签但是ingress在缓存里面没有获取到
	if ok && errors.IsNotFound(err) {
		// create ingress
		ig := c.ConstructIngress(service)
		// ig是要创建的ingress对象
		_, err := c.client.NetworkingV1().Ingresses(namespaceKey).Create(context.Background(), ig, v13.CreateOptions{})
		if err != nil {
			return err
		}
		// 如果serivce没有获取到标签，ingress在缓存里面也没有，就可以删除ingress了
	} else if !ok && ingress != nil {
		// delete ingress
		err := c.client.NetworkingV1().Ingresses(namespaceKey).Delete(context.Background(), name, v13.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// 如果出错了就需要把key重新塞到队列里面去
func (c *controller) handlerError(key string, err error) {
	// 获取指定对象重新如队列的次数
	if c.queue.NumRequeues(key) <= 10 {
		// 传入到限速队列里面去
		// 限速队列是一个用于控制处理速度的队列，它可以防止控制器一次性处理过多的对象，导致系统负载过高
		c.queue.AddRateLimited(key)
	}
	runtime.HandleError(err)
	// 将指定的对象从控制器的队列中删除
	c.queue.Forget(key)
}

func (c *controller) ConstructIngress(service *v14.Service) *v12.Ingress {
	ingress := v12.Ingress{}

	// 指定OwnerReferences
	ingress.ObjectMeta.OwnerReferences = []v13.OwnerReference{
		*v13.NewControllerRef(service, scheme.SchemeGroupVersion.WithKind("Service")),
	}

	ingress.Name = service.Name
	ingress.Namespace = service.Namespace
	pathType := v12.PathTypePrefix
	icn := "nginx"
	ingress.Spec = v12.IngressSpec{
		IngressClassName: &icn,
		Rules: []v12.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: v12.IngressRuleValue{
					HTTP: &v12.HTTPIngressRuleValue{
						Paths: []v12.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: v12.IngressBackend{
									Service: &v12.IngressServiceBackend{
										Name: service.Name,
										Port: v12.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &ingress
}

func NewController(client kubernetes.Interface, serviceInformer informer.ServiceInformer,
	ingressInformer v1.IngressInformer) controller {
	c := controller{
		client: client,
		// ingressInformer.Lister() 是 Kubernetes 中 Ingress Informer 对象的一个方法，用于获取 Ingress 资源的 Lister 接口。
		// Lister 接口可以方便地从内存缓存中获取 Ingress 资源的信息，而不需要向 Kubernetes API Server 发送请求。
		ingressLister: ingressInformer.Lister(),
		serviceLister: serviceInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingressManager"),
	}

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
	})

	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})

	return c
}
