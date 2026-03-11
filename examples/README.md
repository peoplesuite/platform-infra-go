# Examples

Runnable examples for `platform-infra-go` packages. Each subdirectory demonstrates one or more packages with `main` programs you can run locally.

**Requirements:** Postgres, Redis, Neo4j, and/or Temporal must be running (or use the documented environment variables to point to your instances). If a service is unavailable, examples exit with a clear message.

## Run an example

From the repository root:

```bash
go run ./examples/<pkg>/<name>
```

Example:

```bash
go run ./examples/postgres/connect
go run ./examples/redis/client
go run ./examples/neo4j/connect
go run ./examples/temporal/connect
go run ./examples/temporal/worker
```

## Index

### postgres

| Example     | Description                                      |
| ----------- | ------------------------------------------------ |
| `connect`   | Connect with `postgres.New`, `Config`, defer close |
| `basic`     | Connect, health check, and `WithTx` usage         |
| `migrations`| Run migrations from a local embed FS with `RunMigrations` |

**Env:** `POSTGRES_DSN` (default: `postgres://localhost:5432/postgres?sslmode=disable`).

### redis

| Example | Description                                                |
| ------- | ---------------------------------------------------------- |
| `client`| `redis.New`, `GetJSON`/`SetJSON`/`Delete`                   |
| `cache` | `NewCache` with prefix, `GetOrLoad` (get-or-set pattern)    |
| `lock`  | `AcquireLock` / `ReleaseLock` for distributed locking       |

**Env:** `REDIS_ADDR` (default: `localhost:6379`), `REDIS_PASSWORD` (optional), `REDIS_DB` (default: `0`).

### neo4j

| Example   | Description                                           |
| --------- | ----------------------------------------------------- |
| `connect` | `NewDriver`, `Config`, defer `driver.Close`            |
| `query`   | `Query` / `QuerySingle` and record helpers (`Str`, `Int`) |
| `batch`   | `NewBatchWriter`, `Write` with UNWIND `$batch`         |

**Env:** `NEO4J_URI` (default: `bolt://localhost:7687`), `NEO4J_USERNAME` (default: `neo4j`), `NEO4J_PASSWORD` (required), `NEO4J_DATABASE` (default: `neo4j`).

### temporal

| Example   | Description                                                                 |
| --------- | --------------------------------------------------------------------------- |
| `connect` | Create a Temporal client with `temporal/client.NewFromConfig` or from a profile (temporal.toml); defer close |
| `worker`  | Run a worker with `temporal/worker`: client from env/config, `DefaultInterceptors`, one workflow + activity; stop with Ctrl+C |

**Env:** `TEMPORAL_ADDRESS` (default: `localhost:7233`), `TEMPORAL_NAMESPACE` (default: `default`), `TEMPORAL_TASK_QUEUE` (default: `example-queue` for worker). Optional: `TEMPORAL_CONFIG_FILE` to use a temporal.toml profile instead of env.

## Build all examples

```bash
make build-examples
```

Or:

```bash
go build ./examples/...
```

Examples are excluded from `./pkg/...` test coverage.
