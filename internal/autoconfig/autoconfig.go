package autoconfig

import (
	"context"

	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	"github.com/timandy/routine"
	"gorm.io/gorm"
)

type GormTransactionalManager struct {
	Component
	driver      *gorm.DB `@:"Inject"`
	logger      Logging  `@:"Inject"`
	threadLocal routine.ThreadLocal[*gorm.DB]
}

var _ Transactional[gorm.DB] = &GormTransactionalManager{}

var _, _ = any(&GormTransactionalManager{}).(Stereotype)

func (this *GormTransactionalManager) PostConstruct() Error {
	this.threadLocal = routine.NewThreadLocalWithInitial(func() *gorm.DB { return this.driver })
	return nil
}

func (this *GormTransactionalManager) Start(ctx context.Context) *gorm.DB {
	this.logger.Trace("Start Transaction", "db", this)
	tx := this.driver.WithContext(ctx).Begin()
	this.threadLocal.Set(tx)
	return tx
}

func (this *GormTransactionalManager) Get() *gorm.DB {
	if tx := this.threadLocal.Get(); tx != nil {
		return tx
	}
	return this.driver
}

func (this *GormTransactionalManager) End() {
	this.logger.Trace("Ending Transaction", "db", this)
	this.threadLocal.Set(nil)
}

func (this *GormTransactionalManager) Transactional(ctx context.Context, action func(tx *gorm.DB) Error) {
	db := this.Start(ctx)
	err := Try(func() {
		RaiseErr(action(db))
	})
	if err != nil {
		db.Rollback()
	} else {
		db.Commit()
	}
	this.threadLocal.Set(nil)
}

/*
type GormRepositoryConfigurator struct {
	Configuration
	driver *gorm.DB `@:"Inject"`
	//
}

func (this *GormRepositoryConfigurator) GormTxManager() *GormTransactionalManager {
	return &GormTransactionalManager{driver: this.driver, threadLocal: routine.NewThreadLocalWithInitial(func() *gorm.DB { return this.driver })}
}

var _, _ = any(&GormRepositoryConfigurator{}).(Stereotype)
*/
