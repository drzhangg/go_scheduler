package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
)

func main() {
	config := clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}}
	client, err := clientv3.New(config)
	if err != nil {
		fmt.Println("clientv3.New failed, err:", err)
	}

	getRsp, err := client.Get(context.TODO(), "/test/etcd/comment_service", clientv3.WithPrefix())
	if err != nil {
		fmt.Println("client Get failed, err:", err)
	}

	for _, k := range getRsp.Kvs {
		fmt.Println(string(k.Key), string(k.Value))
	}
}
