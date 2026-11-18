# CLAUDE.md

- Go 1.27 DDD service, module `exercise`.
- Aggregates: `Actor`, `Account`. Postgres.
- Bounded context: `internal/participants`.
- HTTP: Huma v2 + chi. SQL: sqlc + pgx/v5.
- README outdated. Trust this file + code.

## Commands

```bash
docker compose up -d                 # Postgres on localhost:5555 (user/pass/db: ddd)
cp .env.example .env                 # backend env, loaded by the app and Taskfile
task migrate:up                      # apply migrations/ to the local DB
go run ./cmd/server                  # API on :8080, OpenAPI docs at /docs

go test ./...                        # full suite; DB tests start Postgres via testcontainers (needs Docker)
go test -short ./...                 # fast: skips DB repository tests
go test -race ./...                  # what CI runs
golangci-lint run ./...              # strict config (.golangci.yaml), CI pins v2.14.0
golangci-lint fmt ./...              # gofumpt + goimports

# Plus the ones defined in Taskfile.yaml
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

Deps inward: `adapters → app → domain`. Domain imports nothing from app/adapters.

## Conventions

- **Aggregates**: unexported fields + getters. `New(...)` validates → `(*T, error)`. `Reconstitute(...)`: repos only, no validation. Mutations = validating methods (`ChangeTaxID`).
- **Errors**:
  - Domain sentinels `Err...`, message `"<agg>: ..."`.
  - Wrap: `fmt.Errorf("verb noun: %w", err)`.
  - Repos: pg errors (`pgconn.PgError` code + constraint, `pgx.ErrNoRows`) → domain sentinels.
  - HTTP: `switch` + `errors.Is` → `huma.ErrorXXX`. Default: `httpapi.Error500InternalServerError(err)`.
- **Optimistic locking**: aggregates carry `version`. Updates take expected version. Mismatch → `ErrVersionConflict` → 409.
- **IDs**: stdlib `uuid`. Inject `newID func() uuid.UUID` into handlers (`uuid.NewV7` in prod).
- **Validation**: rules in domain. Mirror in Huma schema (pattern/enum/minimum) via domain constants (`actor.TaxIDPattern`).
- **Writes**: cross-aggregate → `uow.UnitOfWork.Do`. Single-aggregate → repo directly.
- **New endpoint**: domain → command/query pkg → http pkg → wire `cmd/server/setup.go` → register in `registerHTTPHandlers` (`main.go`).
- **Schema change**: `task migrate:create` → `queries/*.sql` → `task sqlc` → `errors.go` constraint names if needed → `task squawk`.
- Modern Go (`errors.AsType`, `new(expr)`, `t.Context()`). Match it. Don't "fix" it.

## Testing

- testify: `require` for preconditions/errors, `assert` for values.
- Tests only in `domain/` and `adapters/`. Never in `app/`.
- Commands/queries: tested via adapters (HTTP tests + fake repos).
- Repo behaviour: `RunRepositoryContract` (`domain/<agg>/<agg>test`), runs on fake + Postgres. Repo change → add contract cases, sync fake.
- HTTP: `testhelpers.NewHumaAPI` + fake repos. Assert via `AssertHumaOutput` / `AssertHumaError`. Table-driven.
- DB: skip on `testing.Short()`. Truncate tables per case.

## Definition of done

All must pass:

1. `golangci-lint fmt ./... && golangci-lint run ./...`. Fix code. No `//nolint` (if unavoidable: specific + explained).
2. `go test -race ./...`
3. SQL/migrations changed: `task sqlc` no diff, `task squawk` passes.
4. `go mod tidy` no diff.

Then: summarise changes + judgment calls.
