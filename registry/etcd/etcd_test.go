package etcd

import (
	"context"
	"fmt"
	"go_scheduler/registry"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	registryInst, err := registry.InitRegistry(context.TODO(), "etcd", registry.WithAddr([]string{"127.0.0.1:2379"}), registry.WithTimeout(time.Second), registry.WithRegistryPath("/test/etcd/"), registry.WithHeartBeat(5))
	if err != nil {
		fmt.Errorf("init registry failed, err:%v", err)
	}

	service := &registry.Service{
		Name: "comment_service",
	}
	fmt.Println("service:", service)

	service.Nodes = append(service.Nodes, &registry.Node{
		IP:   "127.0.0.1",
		Port: 8801,
	}, &registry.Node{
		IP:   "127.0.0.2",
		Port: 8802,
	})
	err = registryInst.Register(context.TODO(), service)
	if err != nil {
		fmt.Errorf("Register registry failed, err:%v", err)
	}

	for {

		service, err := registryInst.GetService(context.TODO(), "comment_service")
		if err != nil {
			fmt.Errorf("GetService err:%v", err)
			return
		}
		fmt.Println("service:", service)
		time.Sleep(1 * time.Second)
	}
}
