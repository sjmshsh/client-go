package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err)
	}
	// 带缓存的 k8s集群 压力瓶颈 就在api-server
	discoveryClient, err := disk.NewCachedDiscoveryClientForConfig(config, "", "", 10*time.Minute)
	_, lists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		panic(err)
	}
	for _, list := range lists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			panic(err)
		}
		for _, resource := range list.APIResources {
			fmt.Printf("name:%v,group:%v,version:%v\n", resource.Name, gv.Group, gv.Version)
		}
	}
}
