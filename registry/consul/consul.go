package consul

import (
	"github.com/szxby/myMqant/registry"
)

func NewRegistry(opts ...registry.Option) registry.Registry {
	return registry.NewRegistry(opts...)
}
