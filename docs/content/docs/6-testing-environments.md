---
title: "6. Testing & Environments"
weight: 6
---

# Testing & Environments

Enterprise applications need to seamlessly switch databases based on the environment: a robust PostgreSQL instance for
Production, and a fast, ephemeral SQLite instance for local testing.

By combining Rays ORM with the core framework's `ConditionalOnProfile` and `ConfigurationProperties`, you can create a
highly scalable environment strategy.

## 1. The Configuration Files

Define your connection strings in your standard YAML configuration files.

```yaml
# config/application.yaml (Default / Testing)
app:
  database:
    dsn: "file::memory:"
```

```yaml
# config/application-prod.yaml (Loaded only when 'prod' profile is active)
app:
  database:
    dsn: "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable"
```

## 2. The Dialector Strategy

Create an interface to abstract the GORM Dialector opening logic, and implement it for both SQLite and Postgres.

```go
package config

import (
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    . "github.com/BeamFoundry/rays/pkg/core"
)

// 1. The Abstraction
type DBOpener interface {
    Open(string) gorm.Dialector
}

// 2. The SQLite Implementation
type SQLiteDB struct{}

func (this *SQLiteDB) Open(dsn string) gorm.Dialector {
    return sqlite.Open(dsn)
}

// 3. The Postgres Implementation
type PostgresDB struct{}

func (this *PostgresDB) Open(dsn string) gorm.Dialector {
    return postgres.Open(dsn)
}

// 4. The Conditional Dialector Configuration
type DialectorConfig struct {
    Configuration
}

// Loads Postgres ONLY if the 'prod' profile is active
func (this *DialectorConfig) DBOpenerProd(
    _ struct{ ConditionalOnProfile `Name:"prod"` },
) DBOpener {
    return &PostgresDB{}
}

// Fallback to SQLite if no other DBOpener has been registered
func (this *DialectorConfig) DBOpenerFallback(
    _ struct{ ConditionalOnMissingBeam },
) DBOpener {
    return &SQLiteDB{}
}

var _, _ = any(&DialectorConfig{}).(Stereotype)
```

## 3. The Unified Database Provider

Finally, create your actual database config. Inject the resolved properties and the conditionally loaded `DBOpener`.

```go
package config

import (
    "gorm.io/gorm"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

// Bind the YAML properties
type DBConfigProperties struct {
    ConfigurationProperties `Name:"app.database"`
    dsn string
}
var _, _ = any(&DBConfigProperties{}).(Stereotype)

// The Database Provider
type DbConfig struct {
    Configuration
    properties *DBConfigProperties `@:"Inject"`
    dbOpener   DBOpener            `@:"Inject"`
}

func (this *DbConfig) DB() (*gorm.DB, Error) {
    // Open the DB using the conditionally injected dialector and properties!
    if db, err := gorm.Open(this.dbOpener.Open(this.properties.dsn), &gorm.Config{}); err == nil {
        return db, nil
    } else {
        return nil, NewErrorFrom(err)
    }
}

var _, _ = any(&DbConfig{}).(Stereotype)
```

## Zero-Config Testing

With this architecture, testing becomes incredibly frictionless. When you run your Ginkgo test suites using the Rays
`TestingContainer`, the `prod` profile is **not** active by default.

The container will automatically fall back to the `SQLiteDB` opener and the `file::memory:` DSN from your base `application.yaml`,
giving you a pristine, isolated database for every test run without changing a single line of test code!
