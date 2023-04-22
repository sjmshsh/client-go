package main

import (
	"client-go/nginx/pkg"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func main() {
	// 1. config
	config, err := clientcmd.BuildConfigFromFlags("", "D:\\a计算机相关\\基础笔记\\云原生\\config")
	if err != nil {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalln(err)
		}
		config = inClusterConfig
	}
	// 2. client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}
	//
	factory := informers.NewSharedInformerFactory(clientset, 0)
	// 3. informer
	// ServiceInformer是Kubernetes客户端库中的一个Informer对象
	// 用于监控Kubernetes中的Service资源的变化，并将这些变化同步到本地缓存中
	// 使用 ServiceInformer 可以方便地从内存缓存中获取 Service 资源的信息，
	// 而不需要向 Kubernetes API Server 发送请求。在 Kubernetes 二次开发中，如果需要使用 Service 资源，可以通过如下方式使用 ServiceInformer：
	serviceInformer := factory.Core().V1().Services()
	ingressInformer := factory.Networking().V1().Ingresses()

	controller := pkg.NewController(clientset, serviceInformer, ingressInformer)
	stopCh := make(chan struct{})
	// 通过factory启动informer
	factory.Start(stopCh)
	// factory.WaitForCacheSync(stopCh) 是 Kubernetes 中用于等待缓存同步完成的函数。
	// 在 Kubernetes 中，许多资源的信息都存储在内存中的缓存中，例如 Pod、Service、Endpoint 等。
	// 这些缓存数据通常在启动时从 Kubernetes API Server 中获取，然后在运行时进行更新。
	// 在二次开发中，如果需要使用这些缓存数据，需要确保缓存已经同步完成。
	// factory.WaitForCacheSync(stopCh) 函数就是用来等待缓存同步完成的。
	// 它会一直阻塞，直到所有缓存都已经同步完成，或者在 stopCh 接收到信号时返回错误。
	factory.WaitForCacheSync(stopCh)

	controller.Run(stopCh)
}
