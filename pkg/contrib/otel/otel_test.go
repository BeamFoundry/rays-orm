package otel_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/BeamFoundry/rays"
	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	. "github.com/BeamFoundry/rays/pkg/testing"

	. "github.com/BeamFoundry/rays-orm/pkg/contrib/otel"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
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
	db *gorm.DB
}

func (this *testDBConfig) DB(
	_ struct {
		Primary
	},
) (*gorm.DB, Error) {
	if this.db == nil {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			return nil, NewErrorFrom(err)
		}
		this.db = db
	}
	return this.db, nil
}

func (this *testDBConfig) SharedDB(
	_ struct {
		Qualifier `Name:"shared"`
	},
) (*gorm.DB, Error) {
	db, err := this.DB(struct{ Primary }{})
	if err != nil {
		return nil, err
	}
	return db.Session(&gorm.Session{}), nil
}

var recordedTP oteltrace.TracerProvider

type testTPConfig struct {
	Configuration
}

func (this *testTPConfig) TracerProvider() oteltrace.TracerProvider {
	return recordedTP
}

type testDBHolder struct {
	Service
	db *gorm.DB `@:"Inject,Qualifier{dB}"`
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
	// Pre-register a different plugin instance so db.Use returns ErrRegistered
	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, NewErrorFrom(err)
	}
	return db, nil
}

var _ = Describe("GormOtelAutoConfiguration", func() {
	It("instruments *gorm.DB with TracerProvider, config options, customizers, and skips shared registry duplicates", func() {
		Expect(GetContainer().LoadConfigAt(map[string]any{
			"withoutMetrics":        true,
			"withoutQueryVariables": true,
			"withoutServerAddress":  true,
			"recordStackTrace":      true,
			"dbSystem":              "sqlite",
		}, "rays.orm.otel")).NotTo(HaveOccurred())

		sr := tracetest.NewSpanRecorder()
		tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
		recordedTP = tp
		DeferCleanup(func() {
			recordedTP = nil
			Expect(tp.Shutdown(context.Background())).To(Succeed())
		})

		customizer := &testCustomizer{}
		AddBeamFor(&testDBConfig{})
		AddBeamFor(&testDBHolder{})
		AddBeamFor(&testTPConfig{})
		AddBeamFor(customizer)
		GetContainer().EnableScanning()
		Expect(GetContainer().Initialize()).To(Succeed())

		Expect(customizer.called).To(BeTrue())

		db := GetBeamScopedInstance[*testDBHolder]().db
		Expect(db).NotTo(BeNil())
		Expect(db.AutoMigrate(&testItem{})).To(Succeed())
		Expect(db.WithContext(context.Background()).Create(&testItem{Name: "secret-query-value"}).Error).To(Succeed())

		Expect(tp.ForceFlush(context.Background())).To(Succeed())
		spans := sr.Ended()
		Expect(spans).NotTo(BeEmpty())

		var foundCustomAttr bool
		for _, sp := range spans {
			for _, attr := range sp.Attributes() {
				if attr.Key == "custom" && attr.Value.AsString() == "yes" {
					foundCustomAttr = true
				}
				Expect(strings.Contains(attr.Value.AsString(), "secret-query-value")).To(BeFalse())
			}
		}
		Expect(foundCustomAttr).To(BeTrue())
	})

	It("returns an error when db.Use fails with a different registered plugin", func() {
		AddBeamFor(&failingPluginDB{})
		GetContainer().EnableScanning()
		err := GetContainer().Initialize()
		Expect(err).To(HaveOccurred())
	})

	It("surfaces non-KeyNotFoundError from rays.orm.otel config", func() {
		Expect(GetContainer().LoadConfigAt(map[string]any{
			"withoutMetrics": "not-a-bool",
		}, "rays.orm.otel")).NotTo(HaveOccurred())

		AddBeamFor(&testDBConfig{})
		GetContainer().EnableScanning()
		err := GetContainer().Initialize()
		Expect(err).To(HaveOccurred())
	})
})
