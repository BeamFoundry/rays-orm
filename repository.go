package orm

import (
	"reflect"

	"gorm.io/gorm"

	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
)

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

type OrmRepository[T any, K comparable] struct {
	Repository
	transactionManager TxManager[gorm.DB] `@:"Inject"`
	logger             Logging            `@:"Inject"`
}

func (this *OrmRepository[T, ID]) Query(queries ...OrmQueryFunc) *gorm.DB {
	var t T
	db := this.transactionManager.Get().Model(&t)
	if len(queries) > 0 {
		db = db.Scopes(queries...)
	}
	return db
}

func (this *OrmRepository[T, ID]) FindById(id ID) Optional[*T] {
	var t T
	if err := this.transactionManager.Get().First(&t, id).Error; err != nil {
		return OptionalEmpty[*T]()
	}
	return OptionalOf(&t)
}

func (this *OrmRepository[T, ID]) Delete(object *T) Error {
	if err := this.transactionManager.Get().Delete(object).Error; err != nil {
		return NewErrorFrom(err)
	}
	return nil
}

func (this *OrmRepository[T, ID]) DeleteById(id ID) Error {
	var t T
	if err := this.transactionManager.Get().Delete(&t, id).Error; err != nil {
		return NewErrorFrom(err)
	}
	return nil
}

func (this *OrmRepository[T, ID]) Save(object *T) (*T, Error) {
	if err := this.transactionManager.Get().Save(object).Error; err != nil {
		return nil, NewErrorFrom(err)
	}
	return object, nil
}

func (this *OrmRepository[T, ID]) SaveAll(objects []*T) ([]*T, Error) {
	if err := this.transactionManager.Get().Save(objects).Error; err != nil {
		return nil, NewErrorFrom(err)
	}
	return objects, nil
}

func (this *OrmRepository[T, ID]) PostConstruct(parent any) Error {
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
		query := gormQuery.Query(this.transactionManager)
		f.Set(reflect.MakeFunc(function.Type, query))
	}
	return nil
}
