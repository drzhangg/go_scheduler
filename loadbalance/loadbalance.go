package loadbalance

import (
	"context"
	"errors"
	"go_scheduler/registry"
)

var (
	ErrNotHaveNodes = errors.New("not have node")
)

type LoadBalance interface {
	Name() string
	Select(ctx context.Context, nodes []*registry.Node) (node *registry.Node, err error)
}
