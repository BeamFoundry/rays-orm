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
Just like the core Rays framework, it is highly recommended to use Go's dot import feature for the ORM package
(`import . "github.com/BeamFoundry/rays-orm"`). This allows you to use types like `OrmRepository`,
`OrmQueryFunc`, and `GormTransactional` directly without package prefixes, keeping your domain code exceptionally clean and readable.

### Core Architectural Features

1. **Auto-Generated Repository Methods:** By simply defining a function signature like `findByName func(string) *User`
  on your repository struct, Rays ORM will automatically generate the underlying GORM query logic and
  bind it to that field during the container's wiring phase.
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

1.  **Prefix:** Must be `find`, `Find`, `query`, or `Query`.
2.  **Subject:** (Optional) Capitalized words describing what is being found (e.g., `ActiveUsers`).
3.  **Separator:** Must include `By`.
4.  **Conditions:** Field names followed by an optional Operator, separated by `And` or `Or`.
5.  **OrderBy:** (Optional) The keyword `OrderBy` followed by one or more sorting rules. Each rule consists of a field name
and an optional direction (`Asc` or `Desc`, defaulting to `Desc`). Multiple sorting rules must be separated by the `And` keyword.

*Example:* `findUsersByAgeGreaterThanAndNameLikeOrderByNameAscAndCreatedAtDesc`

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
* **Multiple Entities:** `[]*T`
* **Query Builder:** `*gorm.DB` (Returns the query before mapping to an object)
* **Error Handling:** You may append `Error` as a second return parameter to any of the above to catch database faults.

#### The Escape Hatch: `OrmQueryFunc`

If you need dynamic query logic that cannot be expressed in the method name, you can add a variadic parameter of type `...OrmQueryFunc`
(which is an alias for `func(*gorm.DB) *gorm.DB`) as the final argument.

See **Chapter 7: Advanced Querying** for details on using this for preloads and custom SQL.

### 3. Bringing it Together

```go
package repositories

import (
    "time"
    . "github.com/BeamFoundry/rays-orm"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type User struct {
    ID        uint
    Name      string
    Age       uint
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 1. Define the public contract
type UserRepositoryInterface interface {
    OrmRepositoryInterface[User, uint]

    FindByName(name string) *User
    FindActiveUsers(status string, minAge int) ([]*User, Error)
    FindByAgeRangeDynamic(min, max int, customFilters ...OrmQueryFunc) []*User
}

// 2. Define the implementation
type UserRepository struct {
    OrmRepository[User, uint]

    findByName                        func(string) *User
    findByStatusAndAgeGreaterThan     func(string, int) ([]*User, Error)
    findByAgeBetweenOrderByNameAscAndIdDesc func(int, int, ...OrmQueryFunc) []*User
}

// 3. Expose the generated functions
func (this *UserRepository) FindByName(name string) *User {
    return this.findByName(name)
}

func (this *UserRepository) FindActiveUsers(status string, minAge int) ([]*User, Error) {
    return this.findByStatusAndAgeGreaterThan(status, minAge)
}

func (this *UserRepository) FindByAgeRangeDynamic(min, max int, customFilters ...OrmQueryFunc) []*User {
    return this.findByAgeBetweenOrderByNameAscAndIdDesc(min, max, customFilters...)
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
    . "github.com/BeamFoundry/rays-orm"
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
    return s.txManager.Transactional(ctx, func() Error {

        newUser := &User{Name: name}

        // 1. This save natively participates in the active transaction
        _, err := s.userRepo.Save(newUser)
        if err != nil {
            return err // Triggers rollback
        }

        // 2. This custom query also participates in the transaction
        auditLog := &Audit{Action: "USER_CREATED", TargetID: newUser.Id()}
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
    . "github.com/BeamFoundry/rays-orm"
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

The auto-generated repository methods (`findBy...`) are excellent for standard CRUD operations.
However, enterprise applications frequently require eager loading (preloads), pagination, and
complex joins.

Rays ORM handles this complexity using the `OrmQueryFunc` escape hatch, keeping your repository
interfaces clean and putting data hydration logic squarely in the hands of the business service layer.

### 1. Relationships and Preloading

Instead of embedding eager-loading logic into the method name string, Rays ORM provides helper
functions (`WithPreload` and `WithJoins`) that return an `OrmQueryFunc`.

You can pass these helpers as variadic arguments to any auto-generated repository method to dynamically
shape the returned graph.

```go
package services

import (
    . "github.com/BeamFoundry/rays-orm"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type DashboardService struct {
    Service
    userRepo *UserRepository `@:"Inject"`
}

func (s *DashboardService) GetUserData(userId uint) Error {

    // 1. Fetching just the user (Lightweight)
    user, _ := s.userRepo.FindById(userId).Get()

    // 2. Fetching the user WITH their Profile relation (Eager Loading)
    userWithProfile, _ := s.userRepo.FindById(userId, WithPreload("Profile")).Get()

    // 3. Complex Nested Preloading
    detailedUser, _ := s.userRepo.FindById(userId,
        WithPreload("Profile"),
        WithPreload("Posts", "status = ?", "published"), // Conditional preload
        WithPreload("Posts.Comments"),                   // Nested preload
    ).Get()

    return nil
}
```

### 2. Pagination and Limits

Because `OrmQueryFunc` is simply an alias for `func(*gorm.DB) *gorm.DB`, you can easily write
your own helpers for features like pagination.

```go
package utils

import (
    "gorm.io/gorm"
    . "github.com/BeamFoundry/rays-orm"
)

// A custom pagination helper
func Paginate(page int, pageSize int) OrmQueryFunc {
    return func(db *gorm.DB) *gorm.DB {
        offset := (page - 1) * pageSize
        return db.Offset(offset).Limit(pageSize)
    }
}
```

You can then apply this helper to your repository calls dynamically:

```go
func (s *UserService) GetRecentUsers(page int) []*User {
    // The repository method remains generic, but the service dictates the limits!
    users, _ := s.userRepo.FindAllByOrderByCreatedAtDesc(
        utils.Paginate(page, 20),
    )
    return users
}
```

### 3. Raw SQL and Complex Queries

If you hit the limits of the generated query builder (e.g., complex subqueries, massive data aggregations,
or highly specific database vendor features), you can bypass the magic method generation entirely.

Every repository embedding `OrmRepository` exposes a `Query()` method. This returns the raw `*gorm.DB` object
mapped to the current Model, correctly bound to any active `ThreadLocal` transaction.

```go
package repositories

import (
    . "github.com/BeamFoundry/rays-orm"
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

