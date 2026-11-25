---
title: "9. OpenTelemetry Integration"
weight: 9
---

# OpenTelemetry Integration

Rays ORM integrates with `rays-otel` to trace GORM SQL queries and propagate active spans across repositories and transactions.

## Context Propagation

`GormTransactionalManager` automatically binds `TransactionManager.GetContext()` to every `*gorm.DB` operation (both outside and inside active transactions). When combined with `rays-fiber` or `rays-otel`, SQL spans created by GORM automatically become children of the active HTTP or service span.

## Automatic GORM Instrumentation (`pkg/contrib/otel`)

The optional `github.com/BeamFoundry/rays-orm/pkg/contrib/otel` sub-module provides `GormOtelAutoConfiguration`, which automatically instruments every `*gorm.DB` beam in the container with `gorm.io/plugin/opentelemetry/tracing` using the container's `trace.TracerProvider` when available.

### 📦 Installation

```bash
go get github.com/BeamFoundry/rays-orm/pkg/contrib/otel
```

### Usage

Underscore-import `github.com/BeamFoundry/rays-orm/pkg/contrib/otel` in your `main.go`:

```go
package main

import (
    "os"
    "github.com/BeamFoundry/rays"

    _ "github.com/BeamFoundry/rays-orm"
    _ "github.com/BeamFoundry/rays-orm/pkg/contrib/otel"
    _ "github.com/BeamFoundry/rays-otel"
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

### Configuration (`rays.orm.otel`)

You can configure the plugin in your `application.yaml`:

```yaml
rays:
  orm:
    otel:
      dbSystem: sqlite
      withoutMetrics: false
      withoutQueryVariables: false
      withoutServerAddress: false
      recordStackTrace: false
```

### Customizing Options (`GormOtelCustomizer`)

Any Beam implementing `GormOtelCustomizer` can append or override `tracing.Option` values:

```go
type CustomGormOtel struct {
    Service
}

func (this *CustomGormOtel) CustomizeGormOtel(options []tracing.Option) []tracing.Option {
    return append(options, tracing.WithAttributes(attribute.String("db.role", "primary")))
}
```
