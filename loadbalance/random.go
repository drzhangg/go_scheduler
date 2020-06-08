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

	index := rand.Intn(len(nodes))
	node = nodes[index]
	return
}
