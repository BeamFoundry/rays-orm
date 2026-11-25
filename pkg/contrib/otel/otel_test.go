package otel_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/BeamFoundry/rays"
	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	. "github.com/BeamFoundry/rays/pkg/testing"

	. "github.com/BeamFoundry/rays-orm/pkg/contrib/otel"

	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func TestOtel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Otel Suite")
}

type testItem struct {
	Name string
	ID   uint `gorm:"primarykey"`
}

type testDBConfig struct {
	Configuration
}

func (this *testDBConfig) DB() (*gorm.DB, Error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, NewErrorFrom(err)
	}
	return db, nil
}

type wrappedTP struct{ oteltrace.TracerProvider }

type testTPConfig struct {
	Configuration
}

func (this *testTPConfig) TracerProvider() oteltrace.TracerProvider {
	return &wrappedTP{TracerProvider: nooptrace.NewTracerProvider()}
}

type testCustomizer struct {
	called bool
}

var _ GormOtelCustomizer = &testCustomizer{}

func (this *testCustomizer) CustomizeGormOtel(options []tracing.Option) []tracing.Option {
	this.called = true
	return append(options, tracing.WithAttributes(attribute.String("custom", "yes")))
}

type failingPluginDB struct {
	Configuration
}

func (this *failingPluginDB) DB() (*gorm.DB, Error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		return nil, NewErrorFrom(err)
	}
	// Pre-register the plugin so a second db.Use returns ErrRegistered
	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, NewErrorFrom(err)
	}
	return db, nil
}

var _ = Describe("GormOtelAutoConfiguration", func() {
	It("instruments *gorm.DB with TracerProvider, config options, and customizers", func() {
		Expect(GetContainer().LoadConfigAt(map[string]any{
			"withoutMetrics":        true,
			"withoutQueryVariables": true,
			"withoutServerAddress":  true,
			"recordStackTrace":      true,
			"dbSystem":              "sqlite",
		}, "rays.orm.otel")).NotTo(HaveOccurred())

		customizer := &testCustomizer{}
		AddBeamFor(&testDBConfig{})
		AddBeamFor(&testTPConfig{})
		AddBeamFor(customizer)
		GetContainer().EnableScanning()
		Expect(GetContainer().Initialize()).To(Succeed())

		Expect(customizer.called).To(BeTrue())

		db := GetBeamScopedInstance[*gorm.DB]()
		Expect(db.AutoMigrate(&testItem{})).To(Succeed())
		Expect(db.WithContext(context.Background()).Create(&testItem{Name: "one"}).Error).To(Succeed())
	})

	It("returns an error when db.Use fails", func() {
		AddBeamFor(&failingPluginDB{})
		GetContainer().EnableScanning()
		err := GetContainer().Initialize()
		Expect(err).To(HaveOccurred())
	})
})
