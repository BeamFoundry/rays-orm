---
title: "5. Logging Integration"
weight: 5
---

# Logging Integration

The core Rays framework uses a centralized, `slog`-based logging mechanism to ensure structured JSON output
and trace contexts. GORM, however, has its own internal logging interface.

To maintain observability across your enterprise application, you should bridge GORM's output directly into
the Rays `Logging` system. This prevents raw database queries from leaking into standard output as unformatted text.

## Bridging GORM to `slog`

To unify your logging, we highly recommend using the `slog-gorm` adapter package.

### 📦 Installation

```bash
go get github.com/orandin/slog-gorm
```

### Example: Configured Database Beam

Simply inject the `Logging` interface into your `DBConfig` and pass its underlying `slog` handler to GORM during
initialization.

```go
package config

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    slogGorm "github.com/orandin/slog-gorm"

    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
)

type DBConfig struct {
    Configuration
    logger Logging `@:"Inject"`
}

// Provides the raw *gorm.DB Beam to the container
func (this *DBConfig) DB() (*gorm.DB, Error) {

    // Open the database connection and inject the bridged logger
    if db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
        Logger: slogGorm.New(
            slogGorm.WithHandler(this.logger.Handler()),
        ),
    }); err == nil {
        return db, nil
    } else {
        return nil, NewErrorFrom(err)
    }
}

// Register as a Stereotype
var _, _ = any(&DBConfig{}).(Stereotype)
```

With this setup, all GORM database queries, slow query warnings, and SQL errors will be formatted and routed
through your central Rays logging infrastructure!
