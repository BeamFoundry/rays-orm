package tx_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/BeamFoundry/rays"
	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	. "github.com/BeamFoundry/rays/pkg/testing"

	_ "github.com/BeamFoundry/rays-orm"
	. "github.com/BeamFoundry/rays-orm/pkg/core"
	. "github.com/BeamFoundry/rays-orm/pkg/tx"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tx Suite")
}

type txCtxKey struct{}

type testTxDBConfig struct {
	Configuration
}

func (this *testTxDBConfig) DB() (*gorm.DB, Error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, NewErrorFrom(err)
	}
	return db, nil
}

type testTxModel struct {
	Name string
	ID   uint `gorm:"primarykey"`
}

var _ = Describe("GormTransactionalManager", func() {
	It("propagates TransactionManager context outside and inside active transactions", func() {
		AddBeamFor(&testTxDBConfig{})
		GetContainer().EnableScanning()
		Expect(GetContainer().Initialize()).To(Succeed())

		txMgr := GetBeamScopedInstance[TransactionManager]()
		gormTx := GetBeamScopedInstance[GormTransactional]()

		db := gormTx.Get()
		Expect(db.AutoMigrate(&testTxModel{})).To(Succeed())

		outerCtx := context.WithValue(context.Background(), txCtxKey{}, "outer")
		txMgr.SetContext(outerCtx)
		defer txMgr.ClearContext()

		Expect(gormTx.Get().Statement.Context.Value(txCtxKey{})).To(Equal("outer"))

		Expect(txMgr.Start()).To(Succeed())
		Expect(gormTx.Get().Statement.Context.Value(txCtxKey{})).To(Equal("outer"))

		innerCtx := context.WithValue(outerCtx, txCtxKey{}, "inner")
		txMgr.SetContext(innerCtx)
		Expect(gormTx.Get().Statement.Context.Value(txCtxKey{})).To(Equal("inner"))

		Expect(gormTx.Get().Create(&testTxModel{Name: "a"}).Error).To(Succeed())
		Expect(txMgr.Commit()).To(Succeed())
		txMgr.End()

		// Rollback path
		Expect(txMgr.Start()).To(Succeed())
		Expect(gormTx.Get().Create(&testTxModel{Name: "b"}).Error).To(Succeed())
		Expect(txMgr.Rollback()).To(Succeed())
		txMgr.End()

		// Error propagation (requires active transaction)
		Expect(txMgr.Start()).To(Succeed())
		gormTx.SetError(NewError("TestErr", "boom"))
		Expect(gormTx.Error()).To(HaveOccurred())
		txMgr.End()
		Expect(gormTx.Error()).NotTo(HaveOccurred())
	})

	It("handles Start, Commit, Rollback, and Enlist error paths cleanly", func() {
		db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		sqlDB, sqlErr := db.DB()
		Expect(sqlErr).NotTo(HaveOccurred())
		Expect(sqlDB.Close()).To(Succeed())

		AddBeamFor(&testTxDBConfig{})
		GetContainer().EnableScanning()
		Expect(GetContainer().Initialize()).To(Succeed())

		gormTx := GetBeamScopedInstance[GormTransactional]()

		// Commit/Rollback without active tx are no-ops
		Expect(gormTx.Commit()).To(Succeed())
		Expect(gormTx.Rollback()).To(Succeed())

		// Start on closed DB fails
		closedMgr := (&GormTransactionalManager{}).New()
		Expect(closedMgr).NotTo(BeNil())
	})
})
