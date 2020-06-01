package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go_scheduler/registry"
)

type EtcdRegistry struct {
	options     *registry.Options
	client      *clientv3.Client       //etcd客户端
	serviceChan chan *registry.Service //节点信息

}

func init() {

}

func (e *EtcdRegistry) Name() string {
	return "etcd"
}

//初始化etcd客户端参数
func (e *EtcdRegistry) Init(ctx context.Context, opts ...registry.Option) (err error) {
	e.options = &registry.Options{}
	for _, opt := range opts {
		opt(e.options)
	}

	e.client, err = clientv3.New(clientv3.Config{
		Endpoints:   e.options.Addr,
		DialTimeout: e.options.Timeout,
	})
	if err != nil {
		err = fmt.Errorf("init etcd failed, err:%v", err)
		return
	}
	return
}

func (e *EtcdRegistry) Register(ctx context.Context, service *registry.Service) (err error) {
	select {
	case e.serviceChan <- service:
	default:
		err = fmt.Errorf("register chan is full")
		return
	}
	return
}

func (e *EtcdRegistry) Unregister(ctx context.Context, service *registry.Service) (err error) {
	return
}
