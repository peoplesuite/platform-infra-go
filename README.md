# platform-infra-go

Infrastructure clients for the PeopleSuite platform.

This repository provides shared Go clients and helpers for core infrastructure
used across the PeopleSuite ecosystem.

It standardizes:

- **Postgres** – connection, pool config, health checks, transactions, migrations
- **Redis** – client, JSON cache helpers, key prefixes, distributed locks
- **Neo4j** – driver, sessions, typed queries, batch writes
- **Temporal** – client (profile/toml), worker builder, interceptors, workflow start options

The goal is that **every service uses the same patterns** for databases, caches, and workflows.

---

# Package Overview

```
pkg/
postgres     – SQL driver (pgx), Config, Health, WithTx, RunMigrations
postgres/migrations – embedded migration FS
redis        – Client (JSON get/set/delete), Cache (prefix + GetOrLoad), locks
neo4j        – Driver, Config, Query/QuerySingle, record helpers, BatchWriter
temporal     – client (Options, profile/toml), worker builder, interceptors
temporal/config   – TemporalProfile, LoadTemporalProfile, LoadTemporalWorkerConfig
temporal/activities – ClientProvider, ExecuteWorkflow, SignalWorkflow helpers
temporal/start     – StartWorkflowOptions from policies
temporal/policies  – WorkflowPolicy presets, activity timeout presets
```

---

# Example

Minimal Postgres connect and health check:

```go
package main

import (
	"context"
	"time"

	"github.com/peoplesuite/platform-infra-go/pkg/postgres"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	logger := zap.NewNop()
	cfg := postgres.Config{
		DSN:             "postgres://localhost:5432/postgres?sslmode=disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db, err := postgres.New(ctx, cfg, logger)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := postgres.Health(ctx, db); err != nil {
		panic(err)
	}
}
```

Runnable examples live under [examples/](examples/); see [examples/README.md](examples/README.md) for how to run them.

---

# Repository Structure

```
platform-infra-go
│
├── pkg
│   ├── postgres
│   │   └── migrations
│   ├── redis
│   ├── neo4j
│   └── temporal
│       ├── config
│       ├── client
│       ├── activities
│       ├── start
│       ├── policies
│       ├── interceptors
│       └── worker
│
├── examples
│   ├── postgres   (connect, basic, migrations)
│   ├── redis      (client, cache, lock)
│   └── neo4j      (connect, query, batch)
│
└── tools
```

---

# Related Repositories

```
platform-sdk-go       – service runtime and HTTP/gRPC
platform-contracts    – protobuf / API schemas
platform-integrations – external integrations
platform-tools        – developer tooling
```

---

# Development

| Command              | Description                              |
|----------------------|------------------------------------------|
| `make tidy`          | Update dependencies (`go mod tidy`)        |
| `make format`        | Run `go fmt ./...`                       |
| `make vet`           | Run `go vet ./...`                       |
| `make lint`          | Run golangci-lint                        |
| `make lint-examples` | Run golangci-lint on `./examples/...`    |
| `make test`          | Run all tests                            |
| `make test-cover-pkg`| Test and coverage for `./pkg/...`        |
| `make build`         | Build all packages                       |
| `make build-examples`| Build all examples                       |
| `make clean`         | Remove coverage and build artifacts      |
| `make check`         | Format, vet, lint, and test              |

---

# License

Internal PeopleSuite platform library.
