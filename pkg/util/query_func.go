package util

import (
	"math"

	. "github.com/BeamFoundry/rays-orm/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DefaultPageSize is the default number of items per page when pageSize <= 0 in Paginate.
const DefaultPageSize = 10

// Paginate returns an OrmQueryFunc that calculates offset and limit from a 1-indexed page number and page size.
// If page <= 0, page defaults to 1.
// If pageSize <= 0, pageSize defaults to DefaultPageSize.
// If (page - 1) * pageSize exceeds the maximum integer value, a query error is set on the database query.
func Paginate(page, pageSize int) OrmQueryFunc {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	return func(db *gorm.DB) *gorm.DB {
		p := page - 1
		if p > math.MaxInt/pageSize {
			tx := db.Session(&gorm.Session{})
			err := NewError("PaginationOverflowError", "pagination offset overflow", "page", page, "pageSize", pageSize)
			if addErr := tx.AddError(err); addErr != nil {
				return tx
			}
			return tx
		}
		offset := p * pageSize
		if offset == 0 {
			return db.Offset(-1).Limit(pageSize)
		}
		return db.Offset(offset).Limit(pageSize)
	}
}

// WithLimit returns an OrmQueryFunc that limits the number of results returned by the query.
func WithLimit(limit int) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		tx := db.Limit(limit)
		if c, ok := tx.Statement.Clauses["LIMIT"]; ok {
			if l, ok := c.Expression.(clause.Limit); ok {
				l.Limit = &limit
				c.Expression = l
				tx.Statement.Clauses["LIMIT"] = c
			}
		}
		return tx
	}
}

// Limit is an alias for WithLimit.
func Limit(limit int) OrmQueryFunc {
	return WithLimit(limit)
}

// WithOffset returns an OrmQueryFunc that sets the number of records to skip before starting to return records.
func WithOffset(offset int) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		if offset == 0 {
			return db.Offset(-1)
		}
		return db.Offset(offset)
	}
}

// WithPreload returns an OrmQueryFunc that eagerly loads an association or relationship.
func WithPreload(query string, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, args...)
	}
}

// WithJoins returns an OrmQueryFunc that specifies Joins conditions for eager loading or query filtering.
func WithJoins(query string, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, args...)
	}
}

// WithOrder returns an OrmQueryFunc that specifies ordering rules for retrieving records.
func WithOrder(value any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(value)
	}
}

// OrderByAsc returns an OrmQueryFunc that sorts results by the given column in ascending order.
func OrderByAsc(column string) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column},
			Desc:   false,
		})
	}
}

// OrderByDesc returns an OrmQueryFunc that sorts results by the given column in descending order.
func OrderByDesc(column string) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column},
			Desc:   true,
		})
	}
}

// WithSelect returns an OrmQueryFunc that specifies fields to retrieve when querying the database.
func WithSelect(query any, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(query, args...)
	}
}

// WithOmit returns an OrmQueryFunc that specifies columns to omit from the query.
func WithOmit(columns ...string) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Omit(columns...)
	}
}

// WithDistinct returns an OrmQueryFunc that specifies distinct fields from the query.
func WithDistinct(args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Distinct(args...)
	}
}

// WithWhere returns an OrmQueryFunc that adds WHERE conditions to the query.
func WithWhere(query any, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	}
}

// WithOr returns an OrmQueryFunc that adds OR conditions to the query.
func WithOr(query any, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Or(query, args...)
	}
}

// WithGroup returns an OrmQueryFunc that specifies the GROUP BY clause for the query.
func WithGroup(name string) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Group(name)
	}
}

// WithHaving returns an OrmQueryFunc that specifies the HAVING clause for the query.
func WithHaving(query any, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Having(query, args...)
	}
}

// WithUnscoped returns an OrmQueryFunc that ignores soft-deleted conditions when querying.
func WithUnscoped() OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}
}

// WithTable returns an OrmQueryFunc that specifies the table to query.
func WithTable(name string, args ...any) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		return db.Table(name, args...)
	}
}

// Combine combines multiple OrmQueryFunc into a single OrmQueryFunc executed sequentially.
func Combine(funcs ...OrmQueryFunc) OrmQueryFunc {
	return func(db *gorm.DB) *gorm.DB {
		for _, f := range funcs {
			if f != nil {
				db = f(db)
			}
		}
		return db
	}
}
