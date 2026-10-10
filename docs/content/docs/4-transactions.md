---
title: "4. ThreadLocal Transactions"
weight: 4
---

# ThreadLocal Transactions

A common pain point in Go is passing database transaction objects (`*gorm.DB`) explicitly as arguments to
every repository function.

Rays ORM utilizes `github.com/timandy/routine` to establish **ThreadLocal** state for your goroutines.
When you start a transaction via the `GormTxManager`, the transaction is stored in the local thread context.

When your auto-generated repositories execute queries, they transparently check the ThreadLocal context.
If a transaction exists, they use it. If not, they use the global `*gorm.DB`.

## Using the `GormTxManager` Closure

Inject the `*GormTxManager` into your business logic services to orchestrate complex database operations safely.
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

    txManager *GormTxManager  `@:"Inject"`
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

## Manual Transaction Management

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
    transactionManager *GormTxManager  `@:"Inject"`
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
