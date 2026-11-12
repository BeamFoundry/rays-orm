---
title: "7. Advanced Querying"
weight: 7
---

# Advanced Querying

Auto-generated repository methods (`findBy...`) are fast, convenient, and handle standard CRUD operations effortlessly.
However, enterprise applications frequently require dynamic query shaping: eager loading (preloads), pagination,
custom sorting, ad-hoc filters, and complex joins.

Rays ORM handles this complexity using the **`OrmQueryFunc`** escape hatch. By accepting variadic `...OrmQueryFunc`
parameters, repository methods keep their interfaces clean and place data hydration logic squarely in the hands
of the business service layer.

The `github.com/BeamFoundry/rays-orm/pkg/util` package provides a rich set of built-in `OrmQueryFunc` helpers ready
for production use:

```go
import (
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays-orm/pkg/util"
)
```

---

## 1. Quick Reference: Built-in `OrmQueryFunc` Helpers

All helpers return an `OrmQueryFunc` (an alias for `func(*gorm.DB) *gorm.DB`) that can be passed to `repo.Query(...)` or to repository methods declaring a variadic `...OrmQueryFunc` parameter (methods whose signatures do not declare `...OrmQueryFunc` cannot accept query helpers):

| Category | Function | Signature | Description |
|---|---|---|---|
| **Pagination** | `Paginate` | `Paginate(page, pageSize int) OrmQueryFunc` | Sets 1-indexed pagination (`offset = (page - 1) * pageSize`, `limit = pageSize`). Safe defaults for non-positive inputs, overflow safe. |
| **Limits** | `WithLimit` / `Limit` | `WithLimit(limit int) OrmQueryFunc` | Restricts max records; `WithLimit(0)` explicitly enforces zero rows when combined. |
| **Offset** | `WithOffset` | `WithOffset(offset int) OrmQueryFunc` | Skips records; `WithOffset(0)` cancels any previously chained offset. |
| **Relations** | `WithPreload` | `WithPreload(query string, args ...any) OrmQueryFunc` | Eagerly loads relationships or nested associations with optional conditions. |
| **Joins** | `WithJoins` | `WithJoins(query string, args ...any) OrmQueryFunc` | Appends a SQL JOIN condition for eager loading or query filtering. |
| **Sorting** | `OrderByAsc` | `OrderByAsc(column string) OrmQueryFunc` | Sorts results by `column ASC`. |
| **Sorting** | `OrderByDesc` | `OrderByDesc(column string) OrmQueryFunc` | Sorts results by `column DESC`. |
| **Sorting** | `WithOrder` | `WithOrder(value any) OrmQueryFunc` | Applies arbitrary sorting expressions (e.g. `"priority DESC, created_at ASC"`). |
| **Projections** | `WithSelect` | `WithSelect(query any, args ...any) OrmQueryFunc` | Selects specific columns to hydrate from the database. |
| **Projections** | `WithOmit` | `WithOmit(columns ...string) OrmQueryFunc` | Excludes specific columns from model hydration. |
| **Projections** | `WithDistinct` | `WithDistinct(args ...any) OrmQueryFunc` | Enforces `DISTINCT` row selection. |
| **Filtering** | `WithWhere` | `WithWhere(query any, args ...any) OrmQueryFunc` | Adds a dynamic parameterized `WHERE` condition to the query. |
| **Filtering** | `WithOr` | `WithOr(query any, args ...any) OrmQueryFunc` | Adds a dynamic parameterized `OR` condition to the query. |
| **Grouping** | `WithGroup` | `WithGroup(name string) OrmQueryFunc` | Groups records by `GROUP BY name`. |
| **Grouping** | `WithHaving` | `WithHaving(query any, args ...any) OrmQueryFunc` | Adds a `HAVING` aggregation filter clause. |
| **Soft Deletes** | `WithUnscoped` | `WithUnscoped() OrmQueryFunc` | Bypasses soft-delete filtering to include soft-deleted records. |
| **Tables** | `WithTable` | `WithTable(name string, args ...any) OrmQueryFunc` | Overrides the target database table for partition or view queries. |
| **Composition** | `Combine` | `Combine(funcs ...OrmQueryFunc) OrmQueryFunc` | Chains multiple query functions into a single reusable scope. |

---

## 2. Eager Loading & Relations

Instead of embedding eager-loading logic into the method name string, use `WithPreload` and `WithJoins` to dynamically specify association loading on models with defined relationships (such as `User.Profile`, `User.Posts`, and `Post.Comments`):

```go
package services

import (
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays-orm/pkg/util"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type DashboardService struct {
    Service
    userRepo *UserRepository `@:"Inject"`
}

func (s *DashboardService) GetUserData(userId uint) Error {
    // 1. Fetching the user WITH their Profile relation (Eager Loading via Query with WithPreload)
    var userWithProfile User
    if err := s.userRepo.Query(WithPreload("Profile")).First(&userWithProfile, userId).Error; err != nil {
        return NewErrorFrom(err)
    }

    // 2. Complex Nested Preloading
    var detailedUser User
    if err := s.userRepo.Query(
        WithPreload("Profile"),
        WithPreload("Posts", "status = ?", "published"), // Conditional preload
        WithPreload("Posts.Comments"),                   // Nested preload
    ).First(&detailedUser, userId).Error; err != nil {
        return NewErrorFrom(err)
    }

    return nil
}
```

### `WithJoins(query string, args ...any)`

Appends a SQL JOIN statement to the query. Ideal for filtering records by relationship fields or eager-loading 1-to-1 / belongs-to associations:

```go
// Fetch users who have an active subscription via SQL JOIN
activeUsers, err := s.userRepo.ListByStatus("ACTIVE",
    WithJoins("JOIN subscriptions ON subscriptions.user_id = users.id AND subscriptions.status = ?", "ACTIVE"),
)
```

---

## 3. Pagination & Windowing

### `Paginate(page, pageSize int)`

`Paginate` converts a 1-indexed page number and page size into SQL `OFFSET` and `LIMIT` clauses:
$$\text{offset} = (\text{page} - 1) \times \text{pageSize}$$

Safety defaults & overflow protection:
* If `page <= 0`, it automatically defaults to `page = 1`.
* If `pageSize <= 0`, it defaults to `DefaultPageSize` (10).
* If `(page - 1) * pageSize` cannot be represented (overflows `math.MaxInt`), it safely attaches a query error (`PaginationOverflowError` with `"page"` and `"pageSize"` attributes) to the query rather than applying an offset that wraps around to a non-positive value.

```go
func (s *UserService) ListUsers(page, pageSize int) ([]User, Error) {
    users, err := s.userRepo.ListByStatus(
        "ACTIVE",
        Paginate(page, pageSize),
        OrderByDesc("created_at"),
    )
    if err != nil {
        return nil, err
    }
    return users, nil
}
```

* **`Paginate(page, pageSize int)`**: Converts a 1-indexed page number and page size to `OFFSET (page-1)*pageSize LIMIT pageSize`. Non-positive page numbers default to `1`, and non-positive page sizes default to `DefaultPageSize` (10). If the offset overflows integer limits, it safely attaches a `PaginationOverflowError`.
* **`WithLimit(limit int)` / `Limit(limit int)`**: Restricts returned rows. Combining scopes with `WithLimit(0)` explicitly enforces 0 rows while preserving any active offset.
* **`WithOffset(offset int)`**: Skips records. Passing `WithOffset(0)` explicitly cancels any previously set offset.

### `WithLimit(limit int)` and `Limit(limit int)`

Restricts the number of rows returned without altering the offset:

```go
// Fetch the top 5 newest registered users
latestUsers := s.userRepo.FindByRole("MEMBER",
    OrderByDesc("created_at"),
    Limit(5), // or WithLimit(5)
)
```

### `WithOffset(offset int)`

Skips a fixed number of rows:

```go
// Skip the first 10 rows and return the next records
users := s.userRepo.FindByRole("MEMBER",
    OrderByAsc("id"),
    WithOffset(10),
    Limit(20),
)
```

---

## 4. Sorting & Ordering

### `OrderByAsc(column string)` & `OrderByDesc(column string)`

Provides clean, explicit directional ordering:

```go
// Sort users by age ascending
youngestUsers, err := s.userRepo.ListByStatus("ACTIVE", OrderByAsc("age"))

// Sort users by registration date descending
newestUsers, err := s.userRepo.ListByStatus("ACTIVE", OrderByDesc("created_at"))
```

### `WithOrder(value any)`

Allows arbitrary SQL order clauses, including multi-column sorting:

```go
// Multi-column ordering with custom directions
users := s.userRepo.FindByRole("ENGINEER",
    WithOrder("role ASC, age DESC, name ASC"),
)
```

---

## 5. Projections & Field Selection

### `WithSelect(query any, args ...any)`

Restricts the SQL `SELECT` list to specific columns. This avoids hydrating large unused text or JSON fields when performing lightweight queries:

```go
// Hydrate only id, name, and role fields
summaryUsers, err := s.userRepo.ListByStatus("ACTIVE",
    WithSelect("id, name, role"),
)
```

### `WithOmit(columns ...string)`

Excludes specific columns from model hydration:

```go
// Exclude sensitive or bulky fields
publicUsers, err := s.userRepo.ListByStatus("ACTIVE",
    WithOmit("password_hash", "security_token", "bio"),
)
```

### `WithDistinct(args ...any)`

Applies the `DISTINCT` keyword to query results:

```go
// Retrieve unique roles
var roles []string
s.userRepo.Query(WithDistinct("role")).Pluck("role", &roles)
```

---

## 6. Dynamic Filtering & Aggregations

While method names like `FindByNameAndAgeGreaterThan` handle static criteria, business logic often requires dynamic conditions based on optional user filters.

### `WithWhere(query any, args ...any)` & `WithOr(query any, args ...any)`

Injects parameterized WHERE/OR clauses safely:

```go
func (s *UserService) SearchUsers(role string, minAge *uint, maxAge *uint) []User {
    filters := []OrmQueryFunc{}

    if minAge != nil {
        filters = append(filters, WithWhere("age >= ?", *minAge))
    }
    if maxAge != nil {
        filters = append(filters, WithWhere("age <= ?", *maxAge))
    }

    return s.userRepo.FindByRole(role, filters...)
}
```

### `WithGroup(name string)` & `WithHaving(query any, args ...any)`

Enables SQL aggregation pipelines:

```go
type RoleCount struct {
    Role  string
    Count int
}

var results []RoleCount
s.userRepo.Query(
    WithSelect("role, count(*) as count"),
    WithGroup("role"),
    WithHaving("count(*) >= ?", 5),
).Scan(&results)
```

---

## 7. Soft Deletes & Table Overrides

### `WithUnscoped()`

GORM automatically filters out soft-deleted records (`WHERE deleted_at IS NULL`). When building administrative recovery tools or auditing workflows, use `WithUnscoped()` to include deleted records:

```go
// Find a user including soft-deleted entries
var deletedUser User
if err := s.userRepo.Query(WithUnscoped()).First(&deletedUser, userId).Error; err != nil {
    // handle error
}

// Find all soft-deleted records
allHistoricalUsers := s.userRepo.FindByRole("MEMBER",
    WithUnscoped(),
    WithWhere("deleted_at IS NOT NULL"),
)
```

### `WithTable(name string, args ...any)`

Redirects the query to an alternate table or partition while retaining the model's schema mapping:

```go
// Query an archive table for older records
archivedUsers, err := s.userRepo.ListByStatus("INACTIVE",
    WithTable("users_archive_2025"),
)
```

---

## 8. Composing Reusable Query Specifications (`Combine`)

`Combine` merges multiple `OrmQueryFunc` helpers into a single composite function. This lets you encapsulate common business scopes into clean, reusable functions:

```go
package scopes

import (
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays-orm/pkg/util"
)

// ActiveRecentUsers encapsulates three distinct query behaviors
func ActiveRecentUsers(page int) OrmQueryFunc {
    return Combine(
        WithWhere("status = ?", "ACTIVE"),
        OrderByDesc("created_at"),
        Paginate(page, 20),
    )
}

// In the service layer:
func (s *UserService) GetDashboardUsers(role string, page int) []User {
    return s.userRepo.FindByRole(role, ActiveRecentUsers(page))
}
```

---

## 9. Direct GORM Querying (`repo.Query()`)

If you require raw SQL, subqueries, or complex joins that exceed auto-generated methods, call `repo.Query(queries ...OrmQueryFunc)`.

This returns a native GORM `*gorm.DB` instance bound to the model and automatically enrolled in any active **ThreadLocal** transaction:

```go
package repositories

import (
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays-orm/pkg/util"
)

type AnalyticsRepository struct {
    OrmRepository[User, uint]
}

func (r *AnalyticsRepository) GetAverageAgeByRole(role string) float64 {
    var avg float64

    // r.Query() returns *gorm.DB scoped to User and active transactions
    r.Query(
        WithWhere("role = ?", role),
        WithUnscoped(),
    ).
    Select("AVG(age)").
    Scan(&avg)

    return avg
}
```
