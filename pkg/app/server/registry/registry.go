package registry

import "net"

const (
	DefaultWeight = 10
)

// Registry is extension interface of service registry.
type Registry interface {
	Register(info *Info) error
	Deregister(info *Info) error
}

// Info is used for registry.
// The fields are just suggested, which is used depends on design.
type Info struct {
	// ServiceName will be set in hertz by default
	ServiceName string
	// Addr will be set in hertz by default
	Addr net.Addr
	// Weight will be set in hertz by default
	Weight int
	// extend other infos with Tags.
	Tags map[string]string
}

// NoopRegistry is an empty implement of Registry
var NoopRegistry Registry = &noopRegistry{}

// NoopRegistry
type noopRegistry struct{}

func (e noopRegistry) Register(*Info) error {
	return nil
}

func (e noopRegistry) Deregister(*Info) error {
	return nil
}
