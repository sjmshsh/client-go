package main

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 加载k8s的config配置
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err)
	}
	// /api/v1/namespaces/{namespaces}/pods
	// 设置config.APIPath请求的HTTP路径
	config.APIPath = "api"
	// 设置config.GroupVersion请求的资源组/资源版本
	config.GroupVersion = &corev1.SchemeGroupVersion
	// 设置config.NegotiatedSerializer数据的解码器
	config.NegotiatedSerializer = scheme.Codecs
	// 构建rest客户端
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}
	result := &corev1.PodList{}
	// 发请求
	err = restClient.Get().
		Namespace("gitlab").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
		Do(context.Background()).
		Into(result)
	if err != nil {
		panic(err)
	}
	for _, d := range result.Items {
		fmt.Printf("NAMESPACE:%v \t NAME:%v \t STATU:%v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}
