package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	// 创建clientset对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	// 创建stopCh对象，用于在程序退出之前通知Informer退出，因为Informer是一个持久运行的goroutine
	stopCh := make(chan struct{})
	defer close(stopCh)

	// 实例化ShareInformer对象，一共参数是clientset，另一个是time.Minute用于设置多久进行一次resync(重新同步)
	// resync会周期性的执行List操作，将所有的资源存放在Informer Store中，如果参数为0，则禁止resync操作
	sharedInformers := informers.NewSharedInformerFactory(clientset, time.Minute)
	// 得到具体Pod资源的informer对象
	informer := sharedInformers.Core().V1().Pods().Informer()
	// 为Pod资源添加资源事件回调方法，支持AddFunc，UpdateFunc和DelteFunc
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			myObj := obj.(metav1.Object)
			log.Println(myObj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oObj := oldObj.(metav1.Object)
			nObj := newObj.(metav1.Object)
			log.Println(oObj.GetName())
			log.Println(nObj.GetName())
		},
	})
	//通过Run函数运行当前Informer,内部为Pod资源类型创建Informer
	informer.Run(stopCh)
}
