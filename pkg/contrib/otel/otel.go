// Package otel provides OpenTelemetry auto-configuration for GORM databases
// managed by rays-orm.
//
// Underscore-importing this package registers [GormOtelAutoConfiguration],
// which instruments every [*gorm.DB] beam with
// [gorm.io/plugin/opentelemetry/tracing] using the container's
// [go.opentelemetry.io/otel/trace.TracerProvider] when available.
package otel

import (
	_ "github.com/BeamFoundry/rays" // rays
	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"

	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// GormOtelCustomizer can be implemented by any Beam to customize the
// [tracing.Option] slice used when instrumenting [*gorm.DB] instances.
type GormOtelCustomizer interface {
	CustomizeGormOtel(options []tracing.Option) []tracing.Option
}

type gormOtelConfig struct {
	dbSystem              string
	withoutMetrics        bool
	withoutQueryVariables bool
	withoutServerAddress  bool
	recordStackTrace      bool
}

// GormOtelAutoConfiguration automatically registers the OpenTelemetry GORM
// tracing plugin on every [*gorm.DB] beam in the container.
type GormOtelAutoConfiguration struct {
	Component
	env            Environment                        `@:"Inject"`
	logger         Logging                            `@:"Inject"`
	tracerProvider Optional[oteltrace.TracerProvider] `@:"Inject"`
	dbs            []*gorm.DB                         `@:"Inject"`
	customizers    []GormOtelCustomizer               `@:"Inject"`
}

var _, _ = any(&GormOtelAutoConfiguration{}).(Stereotype)

// PostConstruct instruments each [*gorm.DB] beam with the OpenTelemetry GORM plugin.
func (this *GormOtelAutoConfiguration) PostConstruct() Error {
	var cfg gormOtelConfig
	//nolint:errcheck // Optional config section; missing key leaves zero-value defaults
	_ = this.env.ConfigUnmarshall("rays.orm.otel", &cfg)

	var opts []tracing.Option
	if tp, ok := this.tracerProvider.Get(); ok {
		opts = append(opts, tracing.WithTracerProvider(tp))
	}
	if cfg.withoutMetrics {
		opts = append(opts, tracing.WithoutMetrics())
	}
	if cfg.withoutQueryVariables {
		opts = append(opts, tracing.WithoutQueryVariables())
	}
	if cfg.withoutServerAddress {
		opts = append(opts, tracing.WithoutServerAddress())
	}
	if cfg.recordStackTrace {
		opts = append(opts, tracing.WithRecordStackTrace())
	}
	if cfg.dbSystem != "" {
		opts = append(opts, tracing.WithDBSystem(cfg.dbSystem))
	}
	for _, customizer := range this.customizers {
		opts = customizer.CustomizeGormOtel(opts)
	}

	plugin := tracing.NewPlugin(opts...)
	for _, db := range this.dbs {
		if err := db.Use(plugin); err != nil {
			return NewErrorFrom(err)
		}
		this.logger.Debug("Registered OpenTelemetry GORM plugin", "plugin", plugin.Name())
	}
	return nil
}
