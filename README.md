# oficina-utils

Shared Go utility library for the **oficina** ecosystem. Provides reusable packages for messaging, observability, HTTP helpers, database utilities, value objects, and more — designed to be imported by other services in the project.

```
go get github.com/soat13/oficina-utils
```

---

## Packages

### `pkg/awsconfig`

Centralized AWS SDK v2 configuration. Supports static credentials and custom endpoints (useful for local development with LocalStack).

```go
import "github.com/soat13/oficina-utils/pkg/awsconfig"

cfg, err := awsconfig.New(ctx, awsconfig.Config{
    Region:      "us-east-1",
    EndpointURL: "http://localhost:4566", // optional, for local testing
})
```

---

### `pkg/messaging`

Pub/Sub abstraction layer for event-driven communication between services.

#### Core interfaces

| Interface        | Methods                             | Purpose                      |
| ---------------- | ----------------------------------- | ---------------------------- |
| `Consumer`       | `Subscribe`, `Listen`, `Stop`       | Consume messages from queues |
| `QueueSender`    | `Send`                              | Send messages to queues      |
| `TopicPublisher` | `Publish`                           | Publish messages to topics   |
| `QueueBroker`    | combines `Consumer` + `QueueSender` | Full queue broker            |

#### Generic helpers

```go
// Decode a message payload into any type
result, err := messaging.DecodePayload[MyEvent](msg)
```

#### `pkg/messaging/sqs` — SQS Broker

Implements `QueueBroker` using AWS SQS. Supports automatic SNS envelope unwrapping, FIFO queues, and goroutine-based consumers with graceful shutdown.

```go
broker, err := sqs.NewBroker(ctx, awsCfg, "https://sqs.us-east-1.amazonaws.com/123456789/")

// Subscribe to a queue
broker.Subscribe("order-created", func(ctx context.Context, msg messaging.Message) error {
    return nil
})

// Start listening (blocking)
broker.Listen(ctx)
```

A **sync broker** is available for in-process testing without AWS:

```go
broker := sqs.NewSyncBroker()
```

#### `pkg/messaging/sns` — SNS Publisher

Implements `TopicPublisher` using AWS SNS.

```go
pub, err := sns.NewPublisher(ctx, awsCfg, "arn:aws:sns:us-east-1:123456789:")

err = pub.Publish(ctx, messaging.TopicMessage{
    EventName: "order-created",
    Payload:   jsonBytes,
    GroupID:   "order-123",
})
```

---

### `pkg/observability`

Full observability stack powered by **Datadog** — logging (zerolog), APM tracing, StatsD metrics, health checks, and Fiber middleware. A single `Setup` call wires everything together.

```go
import "github.com/soat13/oficina-utils/pkg/observability"

components := observability.Setup(app, db) // app is *fiber.App, db implements DBPinger
defer observability.Shutdown(components)
```

If `app` is `nil`, middleware and health routes are skipped (useful for workers/CLI tools that still need logging and tracing).å

#### What `Setup` registers

- **Logger** — structured JSON logging (zerolog) with Datadog trace ID correlation
- **Tracer** — Datadog APM tracing with runtime metrics
- **Metrics** — StatsD client for custom metrics (`RecordHTTPRequest`, `RecordRepairOrderPhaseDuration`)
- **Health routes** - overall health, readiness, liveness and startup probes, when `app` is provided
- **Middleware** (when `app` is provided):
  - `RequestIDMiddleware` — generates/propagates `X-Request-ID`
  - Datadog Fiber trace middleware
  - `RequestLoggingMiddleware` — structured request/response logging
  - `MetricsMiddleware` — automatic HTTP metrics collection

---

### `pkg/http/fiber`

Utilities for the [Fiber](https://gofiber.io/) web framework.

#### Error Handler

Global error handler that maps application errors to HTTP responses. Integrates with `go-playground/validator` for structured validation error responses (422).

```go
import fiberHelper "github.com/soat13/oficina-utils/pkg/http/fiber"

errorHandler := fiberHelper.NewErrorHandler(errorResolver, validator)
app := fiber.New(fiber.Config{ErrorHandler: errorHandler.Handle})
```

#### ID Parameter

```go
id, err := fiberHelper.GetUuidParam(c, "id") // parses UUID route param
```

#### Pagination

```go
p := fiberHelper.NewPagination(c, 50, 0) // extracts limit/offset from query string
```

---

### `pkg/error`

Error registry that maps domain errors to HTTP status codes.

```go
import errorHelper "github.com/soat13/oficina-utils/pkg/error"

resolver := errorHelper.NewErrorResolver()
resolver.RegisterHTTPNotFoundError(ErrOrderNotFound)
resolver.RegisterHTTPConflictError(ErrDuplicateOrder)

info, found := resolver.Resolve(err)
```

---

### `pkg/db/bun_helper`

Error handling utilities for the [Bun ORM](https://bun.uptrace.dev/).

| Function            | Description                                                            |
| ------------------- | ---------------------------------------------------------------------- |
| `HandleDeleteError` | Converts foreign key violations (SQLSTATE 23503) to `ErrResourceInUse` |
| `IgnoreNoRows`      | Converts `sql.ErrNoRows` to `nil` for safe "not found" handling        |

---

### `pkg/entity`

Base types for domain entities.

```go
import "github.com/soat13/oficina-utils/pkg/entity"

type Order struct {
    ID uuid.UUID
    entity.Timestamps
}
```

---

### `pkg/maps`

Generic functional utilities for slices and maps.

```go
import "github.com/soat13/oficina-utils/pkg/maps"

keys := maps.Keys(myMap)
names := maps.Map(users, func(u User) string { return u.Name })
dtos := maps.MapPtr(users, func(u *User) DTO { return toDTO(u) })
```

---

### `pkg/pagination`

Standard pagination type used across services.

```go
import "github.com/soat13/oficina-utils/pkg/pagination"

p := pagination.New(limit, offset)
```

---

### `pkg/money`

Type-safe money representation in cents with JSON support.

```go
import "github.com/soat13/oficina-utils/pkg/money"

price, err := money.New(1500) // R$ 15,00
total := price.Add(fee)
```

Rejects negative values with `ErrMoneyNegative`. Marshals/unmarshals to/from JSON as `int64`.

---

### `pkg/valueobjects`

Domain-driven value objects with built-in validation. All follow the same pattern: construct with `New(value)`, which returns the value object or a validation error.

| Value Object  | Package                 | Description                                        |
| ------------- | ----------------------- | -------------------------------------------------- |
| `Document`    | `valueobjects/document` | Brazilian CPF/CNPJ with automatic type detection   |
| `Password`    | `valueobjects/password` | Bcrypt-hashed password (8-72 chars)                |
| `PhoneNumber` | `valueobjects/phone`    | Brazilian phone number (11 digits)                 |
| `Plate`       | `valueobjects/plate`    | Brazilian vehicle plate (old and Mercosul formats) |
| `Email`       | `valueobjects/email`    | Email address with RFC-compliant validation        |

---

### `pkg/utils`

Small utility functions.

- **`helpers/string`** — `OnlyNumbers(s)` strips non-digits; `StringToIntOrDefault(s, def)` parses int with fallback
- **`uuid`** — `IDOrNew(id)` returns the given UUID if non-nil, otherwise generates a new one
