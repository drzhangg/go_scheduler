registry：etcd实现的服务注册、服务发现
#### loadbalance：负载均衡设计
    * 负载均衡算法大概分几种：
        * 随机算法  random.go
        * 权重算法  
        * 轮询算法
        * 权重随机算法
        * 权重轮询算法


#### 简单构思下
做个日志收集系统
仿造老男孩项目的思路
elk日志收集：elasticsearch + logstash + kibana
使用gin，etcd，nsq

##### 想要使用的技术：
web框架：gin
orm框架：gorm
redis
docker
docker-compose


grpc
etcd
protobuf
elasticsearch
nsq
go-micro
k8s

从上到下一步一步来

通过sh脚本，上传，更新