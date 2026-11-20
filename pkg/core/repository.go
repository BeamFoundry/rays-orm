package core

import (
	"context"
	"errors"
	"reflect"

	"github.com/timandy/routine"
	"gorm.io/gorm"

	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
)

// GormTransactional represent the interface of a Transactional [gorm.DB]
type GormTransactional interface {
	Transactional[gorm.DB]
}

// GormTransactionalResolver can be implemented by a repository struct embedding [OrmRepository]
// to return a [GormTransactional] for a given [context.Context] (for example, using an injected
// [Foundry] of [GormTransactional]).
type GormTransactionalResolver interface {
	GormTransactional(ctx context.Context) GormTransactional
}

// OrmEntityTypeProvider is implemented by [OrmRepository] (and promoted to any embedding repository)
// to expose the managed entity's [reflect.Type].
type OrmEntityTypeProvider interface {
	EntityType() reflect.Type
}

// OrmRepositoryTypeFromContext extracts the embedding repository's [reflect.Type] (e.g. *UserRepository)
// from a Prototype [context.Context] (checking [PrototypeContextProxyType] first, then [PrototypeContextParentType]).
func OrmRepositoryTypeFromContext(ctx context.Context) reflect.Type {
	if t, ok := ctx.Value(PrototypeContextProxyType).(reflect.Type); ok {
		return t
	}
	if t, ok := ctx.Value(PrototypeContextParentType).(reflect.Type); ok {
		return t
	}
	return nil
}

// OrmEntityTypeFromContext extracts the entity's [reflect.Type] (e.g. User) from a Prototype [context.Context]
// (checking [PrototypeContextProxyInstance] first, then [PrototypeContextParentInstance]).
func OrmEntityTypeFromContext(ctx context.Context) reflect.Type {
	if p, ok := ctx.Value(PrototypeContextProxyInstance).(OrmEntityTypeProvider); ok {
		return p.EntityType()
	}
	if p, ok := ctx.Value(PrototypeContextParentInstance).(OrmEntityTypeProvider); ok {
		return p.EntityType()
	}
	return nil
}

type OrmRepositoryInterface[T any, K comparable] interface {
	// Query() lets you execute queries on <T>, calling gorm Model() *gorm.DB first.
	Query(queries ...OrmQueryFunc) *gorm.DB
	// FindById() lets you find an optional model pointer by its Id or type <ID>.
	FindById(id K) Optional[*T]
	// Delete() lets you delete a model <T> pointer, previously returned.
	Delete(object *T) Error
	// DeleteById() lets you delete a model using its id of type <ID>.
	DeleteById(id K) Error
	// Save() saves a model <T> pointer, returning the pointer and/or an Error.
	Save(object *T) (*T, Error)
	// SaveAll saves a slice of <T> pointers, returning it and/or an Error.
	SaveAll(object []*T) ([]*T, Error)
}

type txContextEntry struct {
	//nolint:containedctx // Caches resolved GormTransactional for the active goroutine context
	ctx context.Context
	tx  GormTransactional
}

type OrmRepository[T any, K comparable] struct {
	Repository
	txManager          TransactionManager `@:"Inject"`
	transactionManager GormTransactional  `@:"Inject"`
	logger             Logging            `@:"Inject"`
	resolver           GormTransactionalResolver
	txCache            routine.ThreadLocal[*txContextEntry]
}

// EntityType returns the [reflect.Type] of the entity <T> managed by this repository.
func (this *OrmRepository[T, ID]) EntityType() reflect.Type {
	return reflect.TypeFor[T]()
}

func (this *OrmRepository[T, ID]) tx() GormTransactional {
	if this.resolver == nil {
		return this.transactionManager
	}
	ctx := this.txManager.GetContext()
	comparableCtx := reflect.TypeOf(ctx).Comparable()
	if this.txCache != nil && comparableCtx {
		if entry := this.txCache.Get(); entry != nil && entry.tx != nil && entry.ctx == ctx {
			return entry.tx
		}
	}
	resolved := this.resolver.GormTransactional(ctx)
	if resolved == nil {
		return this.transactionManager
	}
	if this.txCache != nil && comparableCtx {
		this.txCache.Set(&txContextEntry{ctx: ctx, tx: resolved})
	}
	return resolved
}

func (this *OrmRepository[T, ID]) Query(queries ...OrmQueryFunc) *gorm.DB {
	var t T
	db := this.tx().Get().Model(&t)
	if len(queries) > 0 {
		db = db.Scopes(queries...)
	}
	return db
}

func (this *OrmRepository[T, ID]) FindById(id ID) Optional[*T] {
	tx := this.tx()
	var t T
	if err := tx.Get().First(&t, id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.SetError(NewErrorFrom(err))
		}
		return OptionalEmpty[*T]()
	}
	return OptionalOf(&t)
}

func (this *OrmRepository[T, ID]) Delete(object *T) Error {
	tx := this.tx()
	if err := tx.Get().Delete(object).Error; err != nil {
		e := NewErrorFrom(err)
		tx.SetError(e)
		return e
	}
	return nil
}

func (this *OrmRepository[T, ID]) DeleteById(id ID) Error {
	tx := this.tx()
	var t T
	if err := tx.Get().Delete(&t, id).Error; err != nil {
		e := NewErrorFrom(err)
		tx.SetError(e)
		return e
	}
	return nil
}

func (this *OrmRepository[T, ID]) Save(object *T) (*T, Error) {
	tx := this.tx()
	if err := tx.Get().Save(object).Error; err != nil {
		e := NewErrorFrom(err)
		tx.SetError(e)
		return nil, e
	}
	return object, nil
}

func (this *OrmRepository[T, ID]) SaveAll(objects []*T) ([]*T, Error) {
	tx := this.tx()
	if err := tx.Get().Save(objects).Error; err != nil {
		e := NewErrorFrom(err)
		tx.SetError(e)
		return nil, e
	}
	return objects, nil
}

func (this *OrmRepository[T, ID]) PostConstruct(parent any) Error {
	this.txCache = routine.NewThreadLocal[*txContextEntry]()
	if resolver, ok := parent.(GormTransactionalResolver); ok {
		this.resolver = resolver
	}
	reflectType := reflect.TypeOf(parent)
	queryBuilder := gormQueryBuilder[T, ID]{}.New(this.logger)
	for field := range reflectType.Elem().Fields() {
		if field.Type.Kind() != reflect.Func {
			continue
		}
		function := &field

		gormQuery, err := queryBuilder.Build(function)
		if err != nil {
			err.LogError()
			continue
		}
		f := reflect.ValueOf(parent).Elem().FieldByName(function.Name)
		if !function.IsExported() {
			f = UnprotectValue(f)
		}
		query := gormQuery.Query(this.tx)
		f.Set(reflect.MakeFunc(function.Type, query))
	}
	return nil
}
