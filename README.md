# BeamFoundry Rays ORM ⛁

[![beamfoundry](https://rays.bfdy.dev/rays-orm/rays-orm-bfdy.jpeg)](https://rays.bfdy.dev/rays-orm/)

[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat&logo=go)](https://golang.org)
[![codecov](https://codecov.io/gh/BeamFoundry/rays-orm/graph/badge.svg?token=9YA0XCI0U7)](https://codecov.io/gh/BeamFoundry/rays-orm)

**This is an extract of the full documentation available at [rays.bdfy.dev](https://rays.bfdy.dev/rays-orm/).**

**Rays ORM** is the official database integration package for the **BeamFoundry Rays** framework,
bringing the magic of Spring Data-style repositories to Go using [GORM](https://gorm.io/).

By bridging Rays' Inversion of Control (IoC) container with GORM and ThreadLocal storage,
this package completely eliminates database connection boilerplate, automatic repository method
generation, and context-passing for transactions.

## Introduction and Architecture

In standard Go applications, managing database transactions across multiple service layers usually requires
polluting your function signatures by passing a `tx *gorm.DB` parameter down the entire call stack.
Furthermore, writing basic CRUD operations for every model leads to massive code duplication.

**Rays ORM solves both of these enterprise problems natively.**

### 🧹 Simplifying with Dot Imports
Just like the core Rays framework, it is highly recommended to use Go's dot import feature for the ORM packages:
- `import . "github.com/BeamFoundry/rays-orm/pkg/core"` for core repository types, interfaces, and transaction management (`OrmRepository`, `OrmQueryFunc`, `GormTransactional`).
- `import . "github.com/BeamFoundry/rays-orm/pkg/util"` for ready-to-use query helpers (`Paginate`, `WithLimit`, `WithOffset`, `WithPreload`, `WithJoins`, `OrderByAsc`, etc.).
- `import . "github.com/BeamFoundry/rays/pkg/util/stream"` for lazy, Java 8-style stream operations on query results (`Stream[T]`).

This allows you to use types, query helpers, and stream primitives directly without package prefixes, keeping your domain code exceptionally clean and readable.

### Core Architectural Features

1. **Auto-Generated Repository Methods:** By simply defining a function signature like `findByName func(string) *User`,
  `listActiveUsersByStatus func(string) []User`, or `streamAllByStatusOrderByCreatedAtDesc func(string) Stream[User]` on your repository struct,
  Rays ORM will automatically generate the underlying GORM query logic and bind it to that field during the container's wiring phase.
2. **ThreadLocal Transactions:** Using `github.com/timandy/routine`, Rays ORM implements true ThreadLocal storage for Go.
  When a Service starts a transaction, the `*gorm.DB` transaction object is bound to the current goroutine.
  Downstream repositories transparently inherit this transaction without requiring you to pass it via function arguments.
3. **IoC Managed DB:** You define a standard `*gorm.DB` Beam, and the framework's `GormTransactional` handles the rest.
4. **Secure by Default (SQL Injection Prevention):** The auto-generated query builder does not concatenate raw strings.
It automatically translates your method names into **parameterized queries** using `?` placeholders, ensuring that all database
interactions are natively protected against SQL injection attacks without requiring manual sanitization.

## Quickstart and Setup

To use Rays ORM, you must define a Configuration Beam that provides a raw `*gorm.DB` instance,
and optionally a Migrator service to initialize your schema.

### 1. Defining the Database Beam

Use a standard Rays `Configuration` struct to open your database connection and provide it to the container.

```go
package config

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type DBConfig struct {
    Configuration
}

// Provides the raw *gorm.DB Beam to the container
func (*DBConfig) DB() (*gorm.DB, Error) {
    if db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{}); err == nil {
        return db, nil
    } else {
        return nil, NewErrorFrom(err)
    }
}

```

### 2. Auto-Migrations

Because the database is a standard Beam, you can inject it into a `Service` to handle auto-migrations during
the `PostConstruct` lifecycle phase. You can utilize the Rays `Try / RaiseErr` pattern for clean error handling.

```go
package migrations

import (
    "gorm.io/gorm"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type DbMigrator struct {
    Service
    db *gorm.DB `@:"Inject"`
}

func (this *DbMigrator) PostConstruct() Error {
    return Try(func() {
        RaiseErr(this.db.AutoMigrate(&User{}))
        RaiseErr(this.db.AutoMigrate(&Todo{}))
    })
}

```

### 3. The Entrypoint

In your `main.go`, you simply need to underscore import the `rays-orm` package. This ensures the framework detects
the `GormTransactional` and prepares the repository generation hooks.

```go
package main

import (
    "os"
    "github.com/BeamFoundry/rays"

    // Essential: Registers the Rays ORM hooks and Transaction Manager
    _ "github.com/BeamFoundry/rays-orm"
)

func main() {
    ctx := rays.ApplicationContext()

    if err := ctx.Initialize(); err != nil {
        err.LogErrorWithStackTrace()
        os.Exit(1)
    }

    ctx.Run()
}
```

## Auto-Generated Repositories

Rays ORM eliminates CRUD boilerplate by automatically implementing interface methods based on their function signatures.


### 1. The Base Repository

Rays provides `OrmRepository[T, ID]`, which automatically provides implementations for the standard operations
defined by the `OrmRepositoryInterface`:

* **`Query(queries ...OrmQueryFunc) *gorm.DB`**: Lets you execute queries on `<T>`, calling gorm `Model() *gorm.DB` first.
* **`FindById(id ID) Optional[*T]`**: Lets you find an optional model pointer by its Id of type `<ID>`.
* **`Delete(object *T) Error`**: Lets you delete a model `<T>` pointer, previously returned.
* **`DeleteById(id ID) Error`**: Lets you delete a model using its id of type `<ID>`.
* **`Save(object *T) (*T, Error)`**: Saves a model `<T>` pointer, returning the pointer and/or an `Error`.
* **`SaveAll(object []*T) ([]*T, Error)`**: Saves a slice of `<T>` pointers, returning it and/or an `Error`.

### 2. The Query Builder Syntax

To add custom queries, you define an unexported `func` field on your struct. During container wiring, Rays ORM parses
the name of the function field and **automatically generates the GORM query code** to fulfill it.

#### Anatomy of a Method Name

A query method name must follow a strict grammatical structure:
`[Prefix][Subject]By[Conditions]OrderBy[SortRules]`

1.  **Prefix:** Must be `find`, `Find`, `query`, `Query`, `list`, `List`, `stream`, or `Stream`.
    * Use `find`/`Find` or `query`/`Query` for single entities (`*T`, `Optional[*T]`) or collections (`[]T`).
    * Use `list`/`List` for retrieving multiple entities into a slice (`[]T`).
    * Use `stream`/`Stream` for returning lazy, Java 8-style streams (`Stream[T]`).
2.  **Subject:** (Optional) Capitalized words describing what is being found (e.g., `ActiveUsers`).
3.  **Separator:** Must include `By`.
4.  **Conditions:** Field names followed by an optional Operator, separated by `And` or `Or`.
5.  **OrderBy:** (Optional) The keyword `OrderBy` followed by one or more sorting rules. Each rule consists of a field name
and an optional direction (`Asc` or `Desc`, defaulting to `Desc`). Multiple sorting rules are chained directly without separators.

*Examples:*
* `findUsersByAgeGreaterThanAndNameLikeOrderByNameAscCreatedAtDesc`
* `listUsersByStatusAndAgeGreaterThan(string, uint) ([]User, Error)`
* `streamAllByStatusOrderByCreatedAtDesc(string) Stream[User]`

#### Supported Operators

| Operator Keyword | Equivalent SQL | Arguments Required | Example |
| :--- | :--- | :--- | :--- |
| *(None)*, `Is`, `Equals` | `=` | 1 (Field Type) | `findByName(string)` |
| `Not` | `!=` | 1 (Field Type) | `findByStatusNot(string)` |
| `GreaterThan`, `After` | `>` | 1 (Field Type) | `findByAgeGreaterThan(int)` |
| `LessThan`, `Before` | `<` | 1 (Field Type) | `findByCreatedAtBefore(time.Time)` |
| `GreaterThanOrEqual` | `>=` | 1 (Field Type) | `findByAgeGreaterThanOrEqual(int)` |
| `LessThanOrEqual` | `<=` | 1 (Field Type) | `findByAgeLessThanOrEqual(int)` |
| `IsNull`, `Null` | `IS NULL` | 0 | `findByDeletedAtIsNull()` |
| `IsNotNull`, `NotNull` | `IS NOT NULL` | 0 | `findByEmailNotNull()` |
| `Between` | `BETWEEN x AND y` | 2 (Field Type) | `findByAgeBetween(int, int)` |
| `Like` | `LIKE` | 1 (String) | `findByNameLike(string)` |
| `NotLike` | `NOT LIKE` | 1 (String) | `findByNameNotLike(string)` |
| `In` | `IN (...)` | 1 (Slice of Field Type) | `findByRoleIn([]string)` |
| `NotIn` | `NOT IN (...)` | 1 (Slice of Field Type) | `findByStatusNotIn([]int)` |

#### Security: SQL Injection Prevention

It is crucial to note that Rays ORM does **not** concatenate raw strings to build queries. Under the framework hood, your method
names are translated into **parameterized queries** using `?` placeholders (e.g., `.Where("name = ?", name)`).

The arguments provided to your function are passed directly to GORM's execution engine. This ensures that your application is
entirely protected against SQL injection attacks by default.

#### Return Types

Your generated function can define the following return signatures:
* **Single Entity:** `*T` or `Optional[*T]`
* **Multiple Entities:** `[]T` or `Stream[T]` (Java 8-style lazy Stream from `github.com/BeamFoundry/rays/pkg/util/stream`)
* **Query Builder:** `*gorm.DB` (Returns the query before mapping to an object)
* **Error Handling:** You may append `Error` as a second return parameter to any of the above to catch database faults.

When returning a `Stream[T]`, you can leverage functional transformations like `.Filter(...)`, `.Map(...)`, `.SortedBy(...)`, and `.ToSlice()`:

```go
stream, err := userRepo.StreamUsersByAgeGreaterThan(18)
if err != nil {
    return err
}

adultNames := stream.
    Filter(func(u User) bool { return u.Age >= 21 }).
    Map(func(u User) string { return u.Name }).
    ToSlice()
```

#### The Escape Hatch: `OrmQueryFunc`

If you need dynamic query logic that cannot be expressed in the method name, you can add a variadic parameter of type `...OrmQueryFunc`
(which is an alias for `func(*gorm.DB) *gorm.DB`) as the final argument.

See **Chapter 7: Advanced Querying** for details on using this for preloads and custom SQL.

### 3. Bringing it Together

```go
package repositories

import (
    "time"
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
    . "github.com/BeamFoundry/rays/pkg/util/stream"
)

type Profile struct {
    ID        uint
    UserID    uint
    Bio       string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Comment struct {
    ID        uint
    PostID    uint
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Post struct {
    ID        uint
    UserID    uint
    Title     string
    Status    string
    Comments  []Comment
    CreatedAt time.Time
    UpdatedAt time.Time
}

type User struct {
    ID        uint
    Name      string
    Status    string
    Role      string
    Age       uint
    Profile   Profile
    Posts     []Post
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 1. Define the public contract
type UserRepositoryInterface interface {
    OrmRepositoryInterface[User, uint]

    FindByName(name string) *User
    ListActiveUsers(status string, minAge uint) ([]User, Error)
    ListByStatus(status string, customFilters ...OrmQueryFunc) ([]User, Error)
    FindByRole(role string, customFilters ...OrmQueryFunc) []User
    StreamUsersByAgeGreaterThan(minAge uint) (Stream[User], Error)
    FindByAgeRangeDynamic(min, max uint, customFilters ...OrmQueryFunc) []User
}

// 2. Define the implementation
type UserRepository struct {
    OrmRepository[User, uint]

    findByName                              func(string) *User
    listByStatusAndAgeGreaterThan           func(string, uint) ([]User, Error)
    listByStatus                            func(string, ...OrmQueryFunc) ([]User, Error)
    findByRole                              func(string, ...OrmQueryFunc) []User
    streamByAgeGreaterThan                  func(uint) (Stream[User], Error)
    findByAgeBetweenOrderByNameAscIDDesc    func(uint, uint, ...OrmQueryFunc) []User
}

// 3. Expose the generated functions
func (this *UserRepository) FindByName(name string) *User {
    return this.findByName(name)
}

func (this *UserRepository) ListActiveUsers(status string, minAge uint) ([]User, Error) {
    return this.listByStatusAndAgeGreaterThan(status, minAge)
}

func (this *UserRepository) ListByStatus(status string, customFilters ...OrmQueryFunc) ([]User, Error) {
    return this.listByStatus(status, customFilters...)
}

func (this *UserRepository) FindByRole(role string, customFilters ...OrmQueryFunc) []User {
    return this.findByRole(role, customFilters...)
}

func (this *UserRepository) StreamUsersByAgeGreaterThan(minAge uint) (Stream[User], Error) {
    return this.streamByAgeGreaterThan(minAge)
}

func (this *UserRepository) FindByAgeRangeDynamic(min, max uint, customFilters ...OrmQueryFunc) []User {
    return this.findByAgeBetweenOrderByNameAscIDDesc(min, max, customFilters...)
}
```

## ThreadLocal Transactions

A common pain point in Go is passing database transaction objects (`*gorm.DB`) explicitly as arguments to
every repository function.

Rays ORM utilizes `github.com/timandy/routine` to establish **ThreadLocal** state for your goroutines.
When you start a transaction via the `GormTransactional`, the transaction is stored in the local thread context.

When your auto-generated repositories execute queries, they transparently check the ThreadLocal context.
If a transaction exists, they use it. If not, they use the global `*gorm.DB`.

### Using the `GormTransactional` Closure

Inject the `GormTransactional` into your business logic services to orchestrate complex database operations safely.
The `Transactional` closure accepts a `context.Context` (which is passed to GORM's `WithContext`) and manages commits
and rollbacks automatically based on the returned error.

```go
package services

import (
    "context"
    "gorm.io/gorm"
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type UserService struct {
    Service

    txManager GormTransactional  `@:"Inject"`
    userRepo  *UserRepository `@:"Inject"`
    auditRepo *AuditRepository`@:"Inject"`
}

func (s *UserService) CreateUserWithAudit(ctx context.Context, name string) Error {

    // Execute a block of code within a ThreadLocal transaction
    return s.txManager.Transactional(ctx, func(_ *gorm.DB) Error {

        newUser := &User{Name: name}

        // 1. This save natively participates in the active transaction
        _, err := s.userRepo.Save(newUser)
        if err != nil {
            return err // Triggers rollback
        }

        // 2. This custom query also participates in the transaction
        auditLog := &Audit{Action: "USER_CREATED", TargetID: newUser.ID}
        _, err = s.auditRepo.Save(auditLog)
        if err != nil {
            return err // Triggers rollback
        }

        return nil // Triggers commit
    })
}
```

### Manual Transaction Management

If you prefer or require granular control over the transaction lifecycle, you can manually start, commit,
and rollback transactions.

When using this approach, you **must** call `defer this.transactionManager.End()` immediately after starting
the transaction to ensure the ThreadLocal state is properly cleaned up, even in the event of a panic.

```go
package services

import (
    "context"
    "errors"
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/core"
)

type UserServiceImpl struct {
    Service
    transactionManager GormTransactional  `@:"Inject"`
    userRepository     *UserRepository `@:"Inject"`
}

func (this *UserServiceImpl) AddUser(ctx context.Context, username string) (*User, error) {
    // Start the transaction and bind it to the current thread
    tx := this.transactionManager.Start(ctx)

    // ALWAYS defer End() to clean up the ThreadLocal storage
    defer this.transactionManager.End()

    if _, ok := this.userRepository.FindByUsername(username).Get(); ok {
        tx.Rollback()
        return nil, errors.New("User already exists")
    }

    user := &User{ Name: username }

    u, err := this.userRepository.Save(user)
    if err != nil {
        tx.Rollback()
        return nil, err
    }

    tx.Commit()
    return u, nil
}
```

## Advanced Querying

The auto-generated repository methods (`findBy...`, `listBy...`, `streamBy...`) are excellent for standard CRUD operations.
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

### Quick Reference: Built-in `OrmQueryFunc` Helpers

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

### 1. Relationships and Preloading

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

### 2. Pagination, Limits & Offsets

Using the built-in `Paginate`, `WithLimit`, and `WithOffset` helpers from `pkg/util`:

```go
func (s *UserService) GetRecentUsers(page int) ([]User, Error) {
    // Paginate(page, pageSize) handles 1-indexed pagination with safe defaults & overflow protection
    users, err := s.userRepo.ListByStatus(
        "ACTIVE",
        Paginate(page, 20),
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

### 3. Scope Composition with `Combine`

You can compose multiple query functions into a reusable scope:

```go
activeAuditedScope := Combine(
    WithWhere("status = ?", "ACTIVE"),
    WithPreload("Posts"),
    OrderByDesc("created_at"),
)

users := s.userRepo.FindByRole("ADMIN", activeAuditedScope)
```

### 4. Raw SQL and Complex Queries

If you hit the limits of the generated query builder (e.g., complex subqueries, massive data aggregations,
or highly specific database vendor features), you can bypass the magic method generation entirely.

Every repository embedding `OrmRepository` exposes a `Query()` method. This returns the raw `*gorm.DB` object
mapped to the current Model, correctly bound to any active `ThreadLocal` transaction.

```go
package repositories

import (
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/core"
)

type AnalyticsRepository struct {
    OrmRepository[User, uint]
}

// Manually implementing a complex method
func (r *AnalyticsRepository) GetAverageAgeByRole(role string) float64 {
    var average float64

    // r.Query() returns *gorm.DB mapped to the User table
    r.Query().
        Select("AVG(age)").
        Where("role = ?", role).
        Scan(&average)

    return average
}
```

