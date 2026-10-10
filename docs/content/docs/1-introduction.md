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
Just like the core Rays framework, it is highly recommended to use Go's dot import feature for the ORM package
(`import . "github.com/BeamFoundry/rays-orm"`). This allows you to use types like `OrmRepository`,
`OrmQueryFunc`, and `GormTxManager` directly without package prefixes, keeping your domain code exceptionally clean and readable.

## Core Architectural Features

1. **Auto-Generated Repository Methods:** By simply defining a function signature like `findByName func(string) *User`
  on your repository struct, Rays ORM will automatically generate the underlying GORM query logic and
  bind it to that field during the container's wiring phase.
2. **ThreadLocal Transactions:** Using `github.com/timandy/routine`, Rays ORM implements true ThreadLocal storage for Go.
  When a Service starts a transaction, the `*gorm.DB` transaction object is bound to the current goroutine.
  Downstream repositories transparently inherit this transaction without requiring you to pass it via function arguments.
3. **IoC Managed DB:** You define a standard `*gorm.DB` Beam, and the framework's `GormTxManager` handles the rest.
4. **Secure by Default (SQL Injection Prevention):** The auto-generated query builder does not concatenate raw strings.
It automatically translates your method names into **parameterized queries** using `?` placeholders, ensuring that all database
interactions are natively protected against SQL injection attacks without requiring manual sanitization.
