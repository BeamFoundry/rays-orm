---
title: "7. Advanced Querying"
weight: 7
---

# Advanced Querying

The auto-generated repository methods (`findBy...`) are excellent for standard CRUD operations.
However, enterprise applications frequently require eager loading (preloads), pagination, and
complex joins.

Rays ORM handles this complexity using the `OrmQueryFunc` escape hatch, keeping your repository
interfaces clean and putting data hydration logic squarely in the hands of the business service layer.

## 1. Relationships & Preloading

Instead of embedding eager-loading logic into the method name string, Rays ORM provides helper
functions (`WithPreload` and `WithJoins`) that return an `OrmQueryFunc`.

You can pass these helpers as variadic arguments to any auto-generated repository method to dynamically
shape the returned graph.

```go
package services

import (
    . "github.com/BeamFoundry/rays-orm/pkg/core"
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

## 2. Pagination & Limits

Because `OrmQueryFunc` is simply an alias for `func(*gorm.DB) *gorm.DB`, you can easily write
your own helpers for features like pagination.

```go
package utils

import (
    "gorm.io/gorm"
    . "github.com/BeamFoundry/rays-orm/pkg/core"
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

## 3. Raw SQL & Complex Queries

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
