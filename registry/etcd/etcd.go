package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go_scheduler/registry"
	"path"
	"sync"
	"sync/atomic"
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

	value              atomic.Value //本地缓存数据
	lock               sync.Mutex
	registryServiceMap map[string]*RegistryService //通过map存放节点信息
}

//etcd租约信息
type RegistryService struct {
	id          clientv3.LeaseID                        //租约id
	service     *registry.Service                       //服务信息
	registered  bool                                    //是否注册过
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse //续租chan
}

//通过map缓存数据
type AllServiceInfo struct {
	serviceMap map[string]*registry.Service
}

var (
	etcdRegistry = &EtcdRegistry{
		serviceChan:        make(chan *registry.Service, MaxServiceNum),
		registryServiceMap: make(map[string]*RegistryService, MaxServiceNum),
	}
)

func init() {
	//先把数据存到缓存中
	allServiceInfo := &AllServiceInfo{
		serviceMap: make(map[string]*registry.Service, MaxServiceNum),
	}
	etcdRegistry.value.Store(allServiceInfo)

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
		fmt.Println("把service放到chan中:", service)
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
	//ticker := time.NewTicker(MaxSyncServiceInterval)
	for {
		select {
		//从chan中读取数据
		case service := <-e.serviceChan:
			fmt.Println("run中的chan：", service)
			registryService, ok := e.registryServiceMap[service.Name]
			fmt.Println(ok)
			//说明存在
			if ok {
				fmt.Println("1231231")
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
		//case <-ticker.C:
		//	fmt.Println("进入ticker----")
		//e.syncServiceFromEtcd()
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

func (e *EtcdRegistry) servicePath(name string) string {
	return path.Join(e.options.RegistryPath, name)
}

//服务发现
func (e *EtcdRegistry) GetService(ctx context.Context, name string) (service *registry.Service, err error) {
	//先从缓存中读取
	service, ok := e.getServiceFromCache(ctx, name)
	if ok {
		return
	}

	//加上锁，一次只有一个请求进入
	e.lock.Lock()
	defer e.lock.Unlock()
	//从缓存中读取，如果有数据，直接返回数据
	service, ok = e.getServiceFromCache(ctx, name)
	if ok {
		return
	}

	//缓存中没有数据，从etcd中读取
	key := e.servicePath(name)

	resp, err := e.client.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return
	}
	fmt.Println("Getservice 里：", resp)

	service = &registry.Service{
		Name: name,
	}

	//将从etcd中读取的数据放到service里
	for _, kv := range resp.Kvs {

		fmt.Sprintf("version:%d,key:%s,value:%s\n", kv.Version, kv.Key, kv.Value)
		value := kv.Value
		var tmpService registry.Service
		err = json.Unmarshal(value, &tmpService)
		if err != nil {
			return
		}

		//将etcd中node节点数据append进service
		for _, node := range tmpService.Nodes {
			service.Nodes = append(service.Nodes, node)
		}
	}

	//从缓存中加载数据
	allServiceInfoOld := e.value.Load().(*AllServiceInfo)
	var allServiceNew = &AllServiceInfo{
		serviceMap: make(map[string]*registry.Service, MaxServiceNum),
	}

	for key, val := range allServiceInfoOld.serviceMap {
		allServiceNew.serviceMap[key] = val
	}

	allServiceNew.serviceMap[name] = service
	//将etcd中读取的数据放到缓存中
	e.value.Store(allServiceNew)
	return
}

//从缓存中获取数据
func (e *EtcdRegistry) getServiceFromCache(ctx context.Context, name string) (service *registry.Service, ok bool) {
	allServiceInfo := e.value.Load().(*AllServiceInfo)
	service, ok = allServiceInfo.serviceMap[name]
	return
}

func (e *EtcdRegistry) syncServiceFromEtcd() {

	var allServiceInfoNew = &AllServiceInfo{
		serviceMap: make(map[string]*registry.Service, MaxServiceNum),
	}

	ctx := context.TODO()
	allServiceInfo := e.value.Load().(*AllServiceInfo)
	fmt.Println("缓存加载----", allServiceInfo.serviceMap["comment_service"])

	//对于缓存的每一个服务，都需要从etcd中进行更新
	for _, service := range allServiceInfo.serviceMap {
		key := e.servicePath(service.Name)
		resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
		if err != nil {
			allServiceInfoNew.serviceMap[service.Name] = service
			continue
		}

		serviceNew := &registry.Service{
			Name: service.Name,
		}

		for _, kv := range resp.Kvs {
			value := kv.Value
			var tmpService registry.Service
			err = json.Unmarshal(value, &tmpService)
			if err != nil {
				fmt.Printf("unmarshal failed, err:%v value:%s", err, string(value))
				return
			}

			for _, node := range tmpService.Nodes {
				serviceNew.Nodes = append(serviceNew.Nodes, node)
			}
		}
		allServiceInfoNew.serviceMap[serviceNew.Name] = serviceNew
	}

	e.value.Store(allServiceInfoNew)
}
