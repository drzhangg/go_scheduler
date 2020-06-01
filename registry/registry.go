package registry

import "context"

//服务注册插件的接口
type Registry interface {
	//插件名字
	Name() string
	//初始化插件
	Init(ctx context.Context, opt ...Option) (err error)
	//服务注册	service存放节点信息
	Register(ctx context.Context, service *Service) (err error)
	//取消服务注册
	Unregister(ctx context.Context, service *Service) (err error)
}
