---
title: "1. Introduction"
weight: 1
---

# Introduction & Architecture

In standard Go applications, managing database transactions across multiple service layers usually requires
polluting your function signatures by passing a `tx *gorm.DB` parameter down the entire call stack.
Furthermore, writing basic CRUD operations for every model leads to massive code duplication.

**Rays ORM solves both of these enterprise problems natively.**

## 🧹 Simplifying with Dot Imports
Just like the core Rays framework, it is highly recommended to use Go's dot import feature for the ORM packages:
- `import . "github.com/BeamFoundry/rays-orm/pkg/core"` for core repository types and interfaces (`OrmRepository`, `OrmQueryFunc`, `GormTransactional`).
- `import . "github.com/BeamFoundry/rays-orm/pkg/tx"` for the default `GormTransactionalManager` implementation.
- `import . "github.com/BeamFoundry/rays-orm/pkg/util"` for ready-to-use query helpers (`Paginate`, `WithLimit`, `WithOffset`, `WithPreload`, `WithJoins`, `OrderByAsc`, etc.).
- `import . "github.com/BeamFoundry/rays/pkg/util/stream"` for lazy, Java 8-style stream operations on query results (`Stream[T]`).

This allows you to use types, query helpers, and stream primitives directly without package prefixes, keeping your domain code exceptionally clean and readable.

## Core Architectural Features

1. **Auto-Generated Repository Methods:** By simply defining a function signature like `findByName func(string) *User`,
  `listActiveUsersByStatus func(string) []User`, or `streamAllByStatusOrderByCreatedAtDesc func(string) Stream[User]` on your repository struct,
  Rays ORM will automatically generate the underlying GORM query logic and bind it to that field during the container's wiring phase.
2. **ThreadLocal Transactions:** Using `github.com/timandy/routine` and Rays' central `TransactionManager`, Rays ORM implements true ThreadLocal storage for Go.
  When a Service starts a transaction via `TransactionManager`, the `*gorm.DB` transaction object is lazily enlisted and bound to the current goroutine.
  Downstream repositories transparently inherit this transaction and propagate errors back to `TransactionManager` without requiring you to pass `*gorm.DB` or `context.Context` via function arguments.
3. **IoC Managed DB:** You define a standard `*gorm.DB` Beam, and the framework's `GormTransactional` and `TransactionManager` handle the rest.
4. **Secure by Default (SQL Injection Prevention):** The auto-generated query builder does not concatenate raw strings.
It automatically translates your method names into **parameterized queries** using `?` placeholders, ensuring that all database
interactions are natively protected against SQL injection attacks without requiring manual sanitization.
