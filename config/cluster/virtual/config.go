package virtual

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures the vsphere_virtual_machine_snapshot and
// vsphere_virtual_machine_class resources for cluster scope.
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("vsphere_virtual_machine_snapshot", func(r *ujconfig.Resource) {
		r.ShortGroup = "virtual"
		r.Kind = "MachineSnapshot"
		r.References["virtual_machine_uuid"] = ujconfig.Reference{
			TerraformName: "vsphere_virtual_machine",
		}
	})

	p.AddResourceConfigurator("vsphere_virtual_machine_class", func(r *ujconfig.Resource) {
		r.ShortGroup = "virtual"
		r.Kind = "MachineClass"
	})
}
