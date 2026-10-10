package orm

import (
	"context"

	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	"github.com/timandy/routine"
	"gorm.io/gorm"
)

type GormTxManager struct {
	TxManager[gorm.DB]
	db          *gorm.DB
	threadLocal routine.ThreadLocal[*gorm.DB]
	logger      Logging `@:"Inject"`
}

func (this *GormTxManager) Start(ctx context.Context) *gorm.DB {
	this.logger.Trace("Start Transaction", "db", this)
	tx := this.db.WithContext(ctx).Begin()
	this.threadLocal.Set(tx)
	return tx
}

func (this *GormTxManager) Get() *gorm.DB {
	if tx := this.threadLocal.Get(); tx != nil {
		return tx
	}
	return this.db
}

func (this *GormTxManager) End() {
	this.logger.Trace("Ending Transaction", "db", this)
	this.threadLocal.Set(nil)
}

func (this *GormTxManager) Transactional(ctx context.Context, action func() Error) {
	db := this.Start(ctx)
	err := Try(func() {
		RaiseErr(action())
	})
	if err != nil {
		db.Rollback()
	} else {
		db.Commit()
	}
	this.threadLocal.Set(nil)
}

type GormRepositoryConfigurator struct {
	Configuration
	driver *gorm.DB `@:"Inject"`
	//
}

func (this *GormRepositoryConfigurator) GormTxManager() *GormTxManager {
	return &GormTxManager{db: this.driver, threadLocal: routine.NewThreadLocalWithInitial(func() *gorm.DB { return this.driver })}
}

var _, _ = any(&GormRepositoryConfigurator{}).(Stereotype)
