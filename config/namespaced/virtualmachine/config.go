package virtualmachine

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures the vsphere_virtual_machine resource
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("vsphere_virtual_machine", func(r *ujconfig.Resource) {
		r.ShortGroup = "virtualmachine"
		r.Kind = "VirtualMachine"
	})
}
