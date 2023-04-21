package main

import (
	"fmt"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

func main() {

	// create config
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}

	// create client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// get informer
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Core().V1().Pods().Informer()

	// add workqueue
	rateLimitingQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "controller")

	// add event
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("ADD Event")
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				panic(err)
			}
			rateLimitingQueue.AddRateLimited(key)

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				panic(err)
			}
			rateLimitingQueue.AddRateLimited(key)
			fmt.Println("Update Event")
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("Delete Event")
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				panic(err)
			}
			rateLimitingQueue.AddRateLimited(key)
		},
	})
	// start informer
	stopChan := make(chan struct{})
	factory.Start(stopChan)
	factory.WaitForCacheSync(stopChan)
	<-stopChan
}
