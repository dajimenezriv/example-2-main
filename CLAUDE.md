# CLAUDE.md

Go DDD service (module `exercise`, Go 1.27) with two aggregates, `Actor` and `Account`, persisted in Postgres.
Single bounded context: `internal/participants`. HTTP via Huma v2 on chi. SQL via sqlc + pgx/v5.

> The README is partly outdated (`go run main.go`, `init.sql`, `actormgmt`). Trust this file and the code.

## Commands

```bash
docker compose up -d                 # Postgres on localhost:5555 (user/pass/db: ddd)
cp .env.example .env                 # DATABASE_URL, loaded by the app and Taskfile
task migrate:up                      # apply migrations/ to the local DB
go run ./cmd/server                  # API on :8080, OpenAPI docs at /docs

go test ./...                        # full suite; DB tests start Postgres via testcontainers (needs Docker)
go test -short ./...                 # fast: skips DB repository tests
go test -race ./...                  # what CI runs
golangci-lint run ./...              # strict config (.golangci.yaml), CI pins v2.14.0
golangci-lint fmt ./...              # gofumpt + goimports

task sqlc                            # regenerate + vet after changing queries/*.sql or migrations
task migrate:create -- <name>        # new sequential up/down migration pair
task squawk                          # lint migrations
task cover                           # coverage (excludes actortest/accounttest)
```

## Layout

```
cmd/server/            main.go (config, server, route registration), setup.go (manual DI wiring)
migrations/            golang-migrate SQL, also the sqlc schema source and test DB init
internal/platform/     cross-cutting: httpapi (Huma config, 500 helper), logger
internal/testhelpers/  Huma test API + response/error assertions
internal/participants/
  domain/<agg>/        aggregate, sentinel errors, Repository interface
  domain/<agg>/<agg>test/  in-memory fake repo + shared repository contract tests
  app/command/<name>/  write use cases: Command struct + Handler.Execute
  app/query/<name>/    read use cases: Query
  app/uow/             UnitOfWork interface for multi-aggregate transactions
  adapters/db/         Postgres repos, UnitOfWork, constraint-to-error mapping (errors.go)
  adapters/db/queries/ sqlc SQL; adapters/db/sqlcgen/ is GENERATED, never edit by hand
  adapters/http/<name>/ one Huma endpoint per package: Input/Output, Register, Handle
```

Dependencies point inward: `adapters → app → domain`. The domain imports nothing from app or adapters.

## Conventions

- **Aggregates**: unexported fields + getters. `New(...)` validates and returns `(*T, error)`. `Reconstitute(...)` is for repositories only and skips validation. Mutations are methods (`ChangeTaxID`) that validate.
- **Errors**: domain sentinels named `Err...` with an `"<agg>: ..."` message. Wrap with `fmt.Errorf("verb noun: %w", err)`. Repositories translate pg errors (`pgconn.PgError` code + constraint name, `pgx.ErrNoRows`) into domain sentinels. HTTP handlers map sentinels to `huma.ErrorXXX` in a `switch` with `errors.Is`, and the default case is `httpapi.Error500InternalServerError(err)`.
- **Optimistic locking**: aggregates carry `version`. Updates take the expected version, and a mismatch returns `ErrVersionConflict`, mapped to 409.
- **IDs**: stdlib `uuid` package. Inject `newID func() uuid.UUID` into handlers (`uuid.NewV7` in prod) so tests are deterministic.
- **Validation**: put rules in the domain. Mirror them in the Huma schema (pattern/enum/minimum) by reusing the domain constants (e.g. `actor.TaxIDPattern`).
- **Cross-aggregate writes** go through `uow.UnitOfWork.Do`. Single-aggregate writes use the repository directly.
- **New endpoint checklist**: domain change → command/query package → http package → wire in `cmd/server/setup.go` and register in `registerHTTPHandlers` in `main.go`.
- **Schema change checklist**: new migration pair (`task migrate:create`) → update `queries/*.sql` → `task sqlc` → update `errors.go` constraint names if needed → `task squawk`.
- Modern Go is in use (`errors.AsType`, `new(expr)`, `t.Context()`). Match it, and don't "fix" it.

## Testing

- testify: `require` for preconditions and errors, `assert` for values.
- Repository behaviour lives in `RunRepositoryContract` (`domain/<agg>/<agg>test`). It runs against both the fake and the Postgres repo. When changing a repository, add cases to the contract and keep the fake in sync.
- HTTP tests use `testhelpers.NewHumaAPI` with fake repos and assert via `AssertHumaOutput` / `AssertHumaError`. Prefer table-driven tests.
- DB tests use `testing.Short()` to skip and truncate tables per case.

## Definition of done

Before reporting a task as finished, run and make sure all of these pass:

1. `golangci-lint fmt ./... && golangci-lint run ./...` (fix the code, don't add `//nolint`; if unavoidable, it must be specific and explained)
2. `go test -race ./...`
3. If SQL or migrations changed: `task sqlc` leaves no diff, and `task squawk` passes
4. `go mod tidy` leaves no diff

Then summarise what changed and any decision that was a judgment call.
