package consul

import ()
import "github.com/cloudapex/river/registry"

func NewRegistry(opts ...registry.Option) registry.Registry {
	return registry.NewRegistry(opts...)
}
