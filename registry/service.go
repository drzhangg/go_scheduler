package registry

//服务节点信息
type Service struct {
	Name  string  `json:"name"`
	Nodes []*Node `json:"node"`
}

//节点信息
type Node struct {
	Id     string `json:"id"`
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"` //负载均衡权重
}
