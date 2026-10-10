---
title: "2. Quickstart & Setup"
weight: 2
---

# Quickstart & Setup

To use Rays ORM, you must define a Configuration Beam that provides a raw `*gorm.DB` instance,
and optionally a Migrator service to initialize your schema.

## 1. Defining the Database Beam

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

## 2. Auto-Migrations

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

## 3. The Entrypoint

In your `main.go`, you simply need to underscore import the `rays-orm` package. This ensures the framework detects
the `GormTxManager` and prepares the repository generation hooks.

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
