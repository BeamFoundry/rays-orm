---
title: "4. ThreadLocal Transactions"
weight: 4
---

# ThreadLocal Transactions

A common pain point in Go is passing `context.Context` and database transaction objects (`*gorm.DB`) explicitly as arguments to every service and repository function.

Rays and Rays ORM utilize `github.com/timandy/routine` to establish **ThreadLocal** state for your goroutines:

- **Services** inject the core `TransactionManager` (`github.com/BeamFoundry/rays/pkg/core`) to start, commit, or roll back transactions—and optionally bind a goroutine-local `context.Context`.
- **Repositories** (`OrmRepository`) inject `GormTransactional` (`GormTransactionalManager` from `github.com/BeamFoundry/rays-orm/pkg/tx`), which is registered as a **Prototype** Beam by default.
- **Lazy Enlistment & Error Propagation**: When a repository executes a query while a transaction is active on `TransactionManager`, `GormTransactional` lazily enlists itself in the current goroutine's transaction and opens a `*gorm.DB` transaction using `TransactionManager.GetContext()`. Any database error encountered by a repository is reported back to `GormTransactional.SetError(err)` and forwarded to `TransactionManager`, which immediately rolls back the active transaction (as well as any other participating `Transactional[T]` providers on that goroutine).

---

## Using the `TransactionManager` Closure

Inject `TransactionManager` into your business logic services to orchestrate database operations safely. The `Transactional` closure manages `Start()`, `Commit()`, `Rollback()`, panic recovery, and `End()` automatically:

```go
package services

import (
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type UserService struct {
    Service

    txManager TransactionManager `@:"Inject"`
    userRepo  *UserRepository    `@:"Inject"`
    auditRepo *AuditRepository   `@:"Inject"`
}

func (s *UserService) CreateUserWithAudit(name string) Error {
    // Execute a block of code within a coordinated ThreadLocal transaction
    return s.txManager.Transactional(func() Error {
        newUser := &User{Name: name}

        // 1. Lazily enlists GormTransactional and participates in the active transaction
        _, err := s.userRepo.Save(newUser)
        if err != nil {
            return err // Triggers rollback
        }

        // 2. Reuses the active ThreadLocal *gorm.DB transaction
        auditLog := &Audit{Action: "USER_CREATED", TargetID: newUser.ID}
        _, err = s.auditRepo.Save(auditLog)
        if err != nil {
            return err // Triggers rollback
        }

        return nil // Triggers commit
    })
}
```

### Binding a `context.Context`

If you need to attach a request context, timeout, or cancellation signal to the current goroutine's queries and transactions, call `SetContext(ctx)` and `defer ClearContext()` on `TransactionManager`:

```go
func (s *UserService) CreateUserWithContext(ctx context.Context, name string) Error {
    s.txManager.SetContext(ctx)
    defer s.txManager.ClearContext()

    return s.txManager.Transactional(func() Error {
        _, err := s.userRepo.Save(&User{Name: name})
        return err
    })
}
```

---

## Manual Transaction Management

If you require granular imperative control over the transaction lifecycle, you can manually call `Start()`, `Commit()`, `Rollback()`, and `End()` on `TransactionManager`.

When using this approach, always call `defer this.transactionManager.End()` immediately after `Start()` to ensure the ThreadLocal state is cleaned up (and automatically rolled back if `Commit()` was not reached):

```go
package services

import (
    "errors"

    . "github.com/BeamFoundry/rays/pkg/core"
)

type UserServiceImpl struct {
    Service
    transactionManager TransactionManager `@:"Inject"`
    userRepository     *UserRepository    `@:"Inject"`
}

func (this *UserServiceImpl) AddUser(username string) (*User, error) {
    // Start the transaction on the current goroutine
    if err := this.transactionManager.Start(); err != nil {
        return nil, err
    }

    // ALWAYS defer End() to clean up ThreadLocal storage
    defer this.transactionManager.End()

    if _, ok := this.userRepository.FindByUsername(username).Get(); ok {
        _ = this.transactionManager.Rollback()
        return nil, errors.New("user already exists")
    }

    user := &User{Name: username}

    u, err := this.userRepository.Save(user)
    if err != nil {
        _ = this.transactionManager.Rollback()
        return nil, err
    }

    if err := this.transactionManager.Commit(); err != nil {
        return nil, err
    }
    return u, nil
}
```

---

## Multi-Database Routing: Type Splitting & Context Splitting

Because `GormTransactional` is a **Prototype** Beam and `GormTransactionalManager` tracks active goroutine transactions per `*gorm.DB` driver instance:

- Repositories that resolve to the **same** `*gorm.DB` driver on a goroutine automatically share the same active SQL transaction.
- Repositories that resolve to **different** `*gorm.DB` drivers on a goroutine each enlist their own `*gorm.DB` transaction in `TransactionManager`, providing **best-effort** multi-database commit and rollback coordination (rather than an atomic distributed transaction).

When a Prototype `GormTransactional` injects a Prototype `*gorm.DB` (a `Configuration` method accepting `ctx context.Context`), Rays automatically populates `ctx` with:
- `PrototypeContextProxyType`: The `reflect.Type` of the repository (e.g. `*UserRepository`).
- `PrototypeContextProxyInstance`: The repository instance (which implements `OrmEntityTypeProvider` via `OrmRepository[T, ID]`).
- All runtime context values set on `TransactionManager.SetContext(ctx)`.

Rays ORM provides two helper functions in `github.com/BeamFoundry/rays-orm/pkg/core` to inspect the target repository or entity from `ctx`:

| Helper Function | Return Value | Description |
| :--- | :--- | :--- |
| `OrmEntityTypeFromContext(ctx)` | `reflect.Type` | Returns the `reflect.Type` of entity `T` (e.g. `reflect.TypeFor[Animal]()`). |
| `OrmRepositoryTypeFromContext(ctx)` | `reflect.Type` | Returns the `reflect.Type` of the repository pointer (e.g. `reflect.TypeFor[*AnimalRepository]()`). |

---

### 1. Type Splitting (by Entity Type `T` or Repository Type)

To route different entities or repositories to separate `*gorm.DB` connections without modifying your repository structs, define a Prototype `DB(ctx context.Context) *gorm.DB` method in a `Configuration` Beam:

```go
package config

import (
    "context"
    "reflect"

    "gorm.io/gorm"

    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/core"
)

type MultiDBConfig struct {
    Configuration
    primaryDB *gorm.DB
    auditDB   *gorm.DB
}

// Accepting context.Context makes *gorm.DB a Prototype Beam evaluated per repository
func (this *MultiDBConfig) DB(ctx context.Context) *gorm.DB {
    // Option A: Route by OrmRepository Entity Type T
    if OrmEntityTypeFromContext(ctx) == reflect.TypeFor[AuditLog]() {
        return this.auditDB
    }

    // Option B: Route by Repository Type, Name, or Package Path
    if repoType := OrmRepositoryTypeFromContext(ctx); repoType != nil {
        if repoType == reflect.TypeFor[*AuditRepository]() {
            return this.auditDB
        }
    }

    return this.primaryDB
}
```

When a service runs `this.txManager.Transactional(...)` and calls both `userRepo` (`primaryDB`) and `auditRepo` (`auditDB`), both databases lazily enlist in `TransactionManager` for **best-effort** coordination: if an error occurs before commit, all enlisted databases are rolled back together; at completion, each enlisted database is committed sequentially.

> **Important:** Multi-database coordination is **not** an atomic two-phase commit (2PC). If one database commits and a subsequent database's commit fails, the already-committed changes remain durable while the failing and remaining uncommitted databases are rolled back. Application-level recovery or compensation is required to handle partial commit failures.

---

### 2. Context Splitting (Multi-Tenancy & Dynamic Routing)

When a Controller or Middleware binds runtime parameters (such as a tenant ID or shard key) to `TransactionManager.SetContext(ctx)`, a repository can dynamically resolve its `GormTransactional` from the active `context.Context` by:

1. Injecting `Foundry[GormTransactional]` into the repository struct.
2. Implementing the `GormTransactionalResolver` interface (`GormTransactional(ctx context.Context) GormTransactional`).

`OrmRepository` automatically invokes `GormTransactional(ctx)` with `TransactionManager.GetContext()` and caches the resolved `GormTransactional` for that context on the current goroutine:

```go
package repositories

import (
    "context"

    "gorm.io/gorm"

    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type TenantKeyType string

const TenantKey TenantKeyType = "tenant"

// 1. Repository implements GormTransactionalResolver using an injected Foundry
type UserRepository struct {
    OrmRepository[User, uint]
    foundry    Foundry[GormTransactional] `@:"Inject"`
    findByName func(string) *User
}

var _ GormTransactionalResolver = &UserRepository{}

func (this *UserRepository) GormTransactional(ctx context.Context) GormTransactional {
    if ctx.Value(TenantKey) == nil {
        return nil // Intentional default route: falls back to the default injected GormTransactional
    }
    // Stop the operation and propagate any Forge routing failure rather than falling back to defaultDB
    return RaiseVal(this.foundry.Forge(ctx))
}

func (this *UserRepository) FindByName(name string) *User {
    return this.findByName(name)
}

// 2. Configuration selects the *gorm.DB based on the runtime context.Context
type TenantDBConfig struct {
    Configuration
    defaultDB *gorm.DB
    tenantADB *gorm.DB
    tenantBDB *gorm.DB
}

func (this *TenantDBConfig) DB(ctx context.Context) (*gorm.DB, Error) {
    switch ctx.Value(TenantKey) {
    case nil:
        return this.defaultDB, nil
    case "tenant-a":
        return this.tenantADB, nil
    case "tenant-b":
        return this.tenantBDB, nil
    default:
        return nil, NewError("UnknownTenantError", "unrecognized tenant in context")
    }
}
```

In your Controller or Service, simply bind the tenant context to `TransactionManager`:

```go
ctx := context.WithValue(req.Context(), TenantKey, "tenant-a")
this.txManager.SetContext(ctx)
defer this.txManager.ClearContext()

err := this.txManager.Transactional(func() Error {
    // Automatically forged and executed against tenantADB
    _, err := this.userRepo.Save(&User{Name: "Alice"})
    return err
})
```
