// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	providerconfig "github.com/stuttgart-things/xplane-provider-vspherevm/internal/controller/cluster/providerconfig"
	machineclass "github.com/stuttgart-things/xplane-provider-vspherevm/internal/controller/cluster/virtual/machineclass"
	machinesnapshot "github.com/stuttgart-things/xplane-provider-vspherevm/internal/controller/cluster/virtual/machinesnapshot"
	virtualmachine "github.com/stuttgart-things/xplane-provider-vspherevm/internal/controller/cluster/virtualmachine/virtualmachine"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.Setup,
		machineclass.Setup,
		machinesnapshot.Setup,
		virtualmachine.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.SetupGated,
		machineclass.SetupGated,
		machinesnapshot.SetupGated,
		virtualmachine.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
