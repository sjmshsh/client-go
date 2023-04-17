package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 加载k8s的config配置
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err)
	}
	// 获取clientset客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	// get pods -n gitlab
	podsClient := clientset.CoreV1().Pods("gitlab")
	podList, err := podsClient.List(context.Background(), metav1.ListOptions{Limit: 500})
	if err != nil {
		panic(err)
	}
	for _, v := range podList.Items {
		fmt.Printf("NAMESPACE:%v \t NAME:%v \t STATU:%v\n", v.Namespace, v.Name, v.Status.Phase)
	}
}
