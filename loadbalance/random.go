package loadbalance

import (
	"context"
	"go_scheduler/registry"
	"math/rand"
)

type RandomBalance struct {
}

func (r *RandomBalance) Name() string {
	return "random"
}

func (r *RandomBalance) Select(ctx context.Context, nodes []*registry.Node) (node *registry.Node, err error) {
	if len(nodes) == 0 {
		err = ErrNotHaveNodes
		return
	}

	//加权随机算法
	var totalWeight int
	for _, val := range nodes {
		totalWeight += val.Weight
	}

	curWeight := rand.Intn(totalWeight)
	curIndex := -1
	for index, node := range nodes {
		curWeight -= node.Weight
		if curWeight < 0 {
			curIndex = index
			break
		}
	}

	if curIndex == -1 {
		return
	}

	node = nodes[curIndex]
	return
}
