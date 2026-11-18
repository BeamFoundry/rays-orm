package autoconfig

import (
	"context"

	_ "github.com/BeamFoundry/rays" // rays

	. "github.com/BeamFoundry/rays-orm/pkg/core"
	. "github.com/BeamFoundry/rays-orm/pkg/tx"
	. "github.com/BeamFoundry/rays/pkg/core"
)

type GormConfiguration struct {
	Configuration
}

var _, _ = any(&GormConfiguration{}).(Stereotype)

//nolint:revive // Rays Configuration methods require the anonymous condition struct as first argument
func (this *GormConfiguration) GormTransactional(
	_ struct {
		ConditionalOnMissingBeam
	},
	_ context.Context,
) GormTransactional {
	return GormTransactionalManager{}.New()
}
