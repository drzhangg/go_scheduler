package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go_scheduler/registry"
	"path"
	"time"
)

const (
	MaxServiceNum          = 10               //最大service数量
	MaxSyncServiceInterval = time.Second * 10 //最大异步service数量
)

type EtcdRegistry struct {
	options     *registry.Options
	client      *clientv3.Client       //etcd客户端
	serviceChan chan *registry.Service //节点信息

	registryServiceMap map[string]*RegistryService //通过map存放节点信息
}

//etcd租约信息
type RegistryService struct {
	id          clientv3.LeaseID                        //租约id
	service     *registry.Service                       //服务信息
	registered  bool                                    //是否注册过
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse //续租chan
}

var (
	etcdRegistry = &EtcdRegistry{
		serviceChan:        make(chan *registry.Service, MaxServiceNum),
		registryServiceMap: make(map[string]*RegistryService, MaxServiceNum),
	}
)

func init() {
	//注册服务
	registry.RegisterPlugin(etcdRegistry)
	go etcdRegistry.run()
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
		Endpoints:   e.options.Addrs,
		DialTimeout: e.options.Timeout,
	})
	if err != nil {
		err = fmt.Errorf("init etcd failed, err:%v", err)
		return
	}
	return
}

//注册服务，将service放到e.serviceChan中
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

//这个方法不停的从chan中读取数据，将数据通过map形式进行存储
func (e *EtcdRegistry) run() {
	for {
		select {
		//从chan中读取数据
		case service := <-e.serviceChan:
			registryService, ok := e.registryServiceMap[service.Name]
			//说明存在
			if ok {
				//将
				for _, node := range service.Nodes {
					registryService.service.Nodes = append(registryService.service.Nodes, node)
				}
				registryService.registered = false
				break
			}
			//如果不存在
			registryService = &RegistryService{
				service: service,
			}
			e.registryServiceMap[service.Name] = registryService
		default:
			//
			e.registerOrKeepAlive()
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (e *EtcdRegistry) registerOrKeepAlive() {
	for _, registryService := range e.registryServiceMap {
		//注册过了，保持续租
		if registryService.registered {
			e.keepAlive(registryService)
			continue
		}
		//没有注册过，进行注册
		e.registerService(registryService)
	}
}

func (e *EtcdRegistry) keepAlive(registryService *RegistryService) {
	select {
	case resp := <-registryService.keepAliveCh:
		if resp == nil {
			registryService.registered = false
			return
		}
	}
	return
}

//注册服务
func (e *EtcdRegistry) registerService(registryService *RegistryService) {
	resp, err := e.client.Grant(context.TODO(), e.options.HeartBeat)
	if err != nil {
		return
	}
	registryService.id = resp.ID

	for _, node := range registryService.service.Nodes {
		tmp := &registry.Service{
			Name:  registryService.service.Name,
			Nodes: []*registry.Node{node},
		}

		data, err := json.Marshal(tmp)
		if err != nil {
			continue
		}

		key := e.serviceNodePath(tmp)
		fmt.Printf("register key:%s\n", key)
		_, err = e.client.Put(context.TODO(), key, string(data), clientv3.WithLease(resp.ID))
		if err != nil {
			continue
		}

		ch, err := e.client.KeepAlive(context.TODO(), resp.ID)
		if err != nil {
			continue
		}
		registryService.keepAliveCh = ch
		registryService.registered = true
	}
}

func (e *EtcdRegistry) serviceNodePath(service *registry.Service) string {
	nodeIp := fmt.Sprintf("%s:%d", service.Nodes[0].IP, service.Nodes[0].Port)
	return path.Join(e.options.RegistryPath, service.Name, nodeIp)
}
