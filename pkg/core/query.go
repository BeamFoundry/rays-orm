package core

import (
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	. "github.com/BeamFoundry/rays/pkg/lang"
	. "github.com/BeamFoundry/rays/pkg/util"
)

type OrmQueryFunc = func(*gorm.DB) *gorm.DB

/*
func WithPreload(query string, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, args...)
	}
}

func WithJoins(query string, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, args...)
	}
}
*/

// Where Mode
type whereModeEnum struct {
	AND,
	OR whereMode
}

type whereMode = Enum[whereModeEnum]

var whereModes whereModeEnum = whereMode{}.New()

// where Operator
type operatorEnum struct {
	IS, EQUALS, NOT,
	GREATERTHAN, AFTER,
	LESSTHAN, BEFORE,
	GREATERTHANOREQUAL, LESSTHANOREQUAL,
	ISNULL, NULL,
	ISNOTNULL, NOTNULL,
	BETWEEN,
	LIKE, IN,
	NOTLIKE, NOTIN operator
}

type operator = Enum[operatorEnum]

var operators operatorEnum = operator{}.New() // Initialize the struct instance

// order sort
type orderSortingEnum struct {
	ASC,
	DESC orderSorting
}

type orderSorting = Enum[orderSortingEnum]

var orderSortings orderSortingEnum = orderSorting{}.New()

type OrmQuery interface {
	Query(args ...any) OrmQueryFunc
}

type whereQuery struct {
	field    *reflect.StructField // One of T member
	mode     whereMode
	operator operator // Operator to use, See above.
}

func (this *whereQuery) NumArgs() int {
	switch this.operator {
	case operators.ISNULL, operators.ISNOTNULL, operators.NULL, operators.NOTNULL:
		return 0
	case operators.BETWEEN:
		return 2
	default:
		return 1
	}
}

// operator will be translated from the text parsed and put in uppercase into SQL.
//
//   - "IS", "EQUALS" : translates to SQL operator "=" // The default operator if none is specified
//   - "NOT : translates to SQL operator "<>"
//   - "GREATERTHAN" or "AFTER" : translates to ">"
//   - "LESSTHAN", "BEFORE" : translates to "<"
//   - "GREATERTHANOREQUAL" : translates to ">="
//   - "LESSTHANOREQUAL" : translates to "<="
//   - "ISNULL", "NULL" : translates to "IS NULL"
//   - "ISNOTNULL", "NOTNULL" : translates to "NOT NULL"
//   - "LIKE", "IN", "BETWEEN": translates to themself
//   - "NOTLIKE" : "NOT LIKE"
//   - "NOTIN": "NOT IN"
func (this *whereQuery) Operator() string {
	switch this.operator {
	case operators.IS, operators.EQUALS:
		return "="
	case operators.NOT:
		return "<>"
	case operators.GREATERTHAN, operators.AFTER:
		return ">"
	case operators.GREATERTHANOREQUAL:
		return ">="
	case operators.LESSTHAN, operators.BEFORE:
		return "<"
	case operators.LESSTHANOREQUAL:
		return "<="
	case operators.ISNULL, operators.NULL:
		return "IS NULL"
	case operators.ISNOTNULL, operators.NOTNULL:
		return "NOT NULL"
	case operators.NOTLIKE:
		return "NOT LIKE"
	default:
		return this.operator.String()
	}
}

func (this *whereQuery) Query(args ...any) OrmQueryFunc {
	field := CaseUtils(this.field.Name).SnakeCase().String()
	operator := this.Operator()
	return func(db *gorm.DB) *gorm.DB {
		var mode func(any, ...any) *gorm.DB
		switch this.mode {
		case whereModes.OR:
			mode = db.Or
		default:
			mode = db.Where
		}
		sql := String(field + " " + operator)
		switch this.NumArgs() {
		case 1:
			sql = sql.Concat(" ?")
			return mode(sql.String(), args[0])
		case 2:
			sql = sql.Concat(" ? AND ?")
			return mode(sql.String(), args[0], args[1])
		}
		return mode(sql.String())
	}
}

type orderByQuery struct {
	field *reflect.StructField // One of T member
	order orderSorting         // ASC or DESC
}

/*
func newOrderByQuery(field, order string) orderByQuery {
	return orderByQuery{field, order}
}
*/

func (this *orderByQuery) Query(_ ...any) OrmQueryFunc {
	field := CaseUtils(this.field.Name).SnakeCase().String()
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: field},
			Desc:   this.order.Compare(orderSortings.DESC) == 0,
		})
	}
}

/*
type preloadQuery struct {
	field   string
	orderby string
}

func newPreloadQuery(field, orderby string) preloadQuery {
	return preloadQuery{field, orderby}
}

func (this preloadQuery) Query(_ ...any) OrmQueryFunc {
	field := CaseUtils(this.field).SnakeCase().String()
	if this.orderby == "" {
		return func(db *gorm.DB) *gorm.DB {
			return db.Preload(this.field)
		}
	}
	order := addDefaultOperator(this.orderby, "Asc", reOrderDefaultOperator)
	orders := reExtractOrder.FindAllStringSubmatch(order, -1)
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(this.field, func(db *gorm.DB) *gorm.DB {
			columns := []clause.OrderByColumn{}
			for _, order := range orders {
				columns = append(columns, clause.OrderByColumn{
					Column: clause.Column{Name: field + "." + CaseUtils(order[1]).SnakeCase().String()},
					Desc:   String(order[2]).ToUpperCase().Equals("DESC"),
				})
			}
			return db.Order(clause.OrderBy{Columns: columns})
		})
	}
}
*/
