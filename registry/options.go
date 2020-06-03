package registry

import "time"

type Options struct {
	Addrs         []string      //地址集合
	Timeout      time.Duration //超时时间
	RegistryPath string        //注册地址
	HeartBeat    int64         //心跳
}

type Option func(options *Options)

func WithAddr(adds []string) Option {
	return func(options *Options) {
		options.Addrs = adds
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(options *Options) {
		options.Timeout = timeout
	}
}

func WithRegistryPath(path string) Option {
	return func(options *Options) {
		options.RegistryPath = path
	}
}

func WithHeartBeat(heartBeat int64) Option {
	return func(options *Options) {
		options.HeartBeat = heartBeat
	}
}
