package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	gvs := schema.GroupVersionResource{
		Group:    "core",
		Version:  "v1",
		Resource: "pods",
	}
	unstructuredList, err := dynamicClient.Resource(gvs).
		Namespace("gitlab").
		List(context.Background(), metav1.ListOptions{Limit: 500})
	if err != nil {
		panic(err)
	}
	podList := &corev1.PodList{}
	// 通过runtime的函数将unstructured.UnstructuredList转换成PodList
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredList.UnstructuredContent(), podList)
	if err != nil {
		panic(err)
	}
	for _, v := range podList.Items {
		fmt.Printf("NAMESPACE:%v \t NAME:%v \t STATU:%v\n", v.Namespace, v.Name, v.Status.Phase)
	}
}
