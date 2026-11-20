package tx

import (
	"github.com/timandy/routine"
	"gorm.io/gorm"

	. "github.com/BeamFoundry/rays-orm/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
)

var activeTransactions = routine.NewThreadLocal[map[*gorm.DB]*gorm.DB]()

type GormTransactionalManager struct {
	driver    *gorm.DB           `@:"Inject"`
	txManager TransactionManager `@:"Inject"`
	logger    Logging            `@:"Inject"`
}

var _ GormTransactional = &GormTransactionalManager{}

func (this GormTransactionalManager) New() *GormTransactionalManager {
	return &this
}

func (this *GormTransactionalManager) getActiveTx() *gorm.DB {
	if m := activeTransactions.Get(); m != nil {
		return m[this.driver]
	}
	return nil
}

func (this *GormTransactionalManager) setActiveTx(tx *gorm.DB) {
	m := activeTransactions.Get()
	if m == nil {
		m = make(map[*gorm.DB]*gorm.DB)
		activeTransactions.Set(m)
	}
	m[this.driver] = tx
}

func (this *GormTransactionalManager) clearActiveTx() {
	if m := activeTransactions.Get(); m != nil {
		delete(m, this.driver)
		if len(m) == 0 {
			activeTransactions.Remove()
		}
	}
}

func (this *GormTransactionalManager) Start() Error {
	this.logger.Trace("Start Transaction", "db", this)
	tx := this.driver.WithContext(this.txManager.GetContext()).Begin()
	if tx.Error != nil {
		return NewErrorFrom(tx.Error)
	}
	this.setActiveTx(tx)
	return nil
}

func (this *GormTransactionalManager) Commit() Error {
	this.logger.Trace("Commit Transaction", "db", this)
	if tx := this.getActiveTx(); tx != nil {
		if err := tx.Commit().Error; err != nil {
			return NewErrorFrom(err)
		}
	}
	return nil
}

func (this *GormTransactionalManager) Rollback() Error {
	this.logger.Trace("Rollback Transaction", "db", this)
	if tx := this.getActiveTx(); tx != nil {
		if err := tx.Rollback().Error; err != nil {
			return NewErrorFrom(err)
		}
	}
	return nil
}

func (this *GormTransactionalManager) Get() *gorm.DB {
	if tx := this.getActiveTx(); tx != nil {
		return tx
	}
	if this.txManager.IsActive() {
		if err := this.txManager.Enlist(this); err != nil {
			db := this.driver.WithContext(this.txManager.GetContext())
			//nolint:errcheck // GORM AddError stores error on db.Error and returns it
			db.AddError(err)
			return db
		}
		return this.getActiveTx()
	}
	return this.driver.WithContext(this.txManager.GetContext())
}

func (this *GormTransactionalManager) End() {
	this.logger.Trace("Ending Transaction", "db", this)
	this.clearActiveTx()
}

func (this *GormTransactionalManager) SetError(err Error) {
	this.txManager.SetError(err)
}

func (this *GormTransactionalManager) Error() Error {
	return this.txManager.Error()
}
