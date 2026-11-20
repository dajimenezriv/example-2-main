# personal

Go HTTP service (chi, OpenTelemetry, slog). Entry point: `cmd/server`.

## Commands

- `task run`: start the server
- `task lint`: `golangci-lint run`
- `task fmt`: `golangci-lint fmt`
- `go test ./...`: run tests

Run `task fmt` and `task lint` before calling Go work done.

## Go style

Follow the [Google Go Style Decisions](https://google.github.io/styleguide/go/decisions). It is subordinate to the [core guide](https://google.github.io/styleguide/go/guide), which wins on conflict; [best practices](https://google.github.io/styleguide/go/best-practices) covers the longer-form advice.

The rules below are the ones that come up most and that the linters don't reliably catch. Fetch the page when a case isn't covered here, or when reviewing style.

Skip the parts specific to Google's codebase: BUILD files, protobuf imports, `log.Exit`, and underscores in flag names.

### Naming

- Don't repeat the package in exported names: `widget.New`, not `widget.NewWidget`; `db.Load`, not `db.LoadFromDatabase`.
- No `Get` prefix: `Counts`, not `GetCounts`. Use `Compute` or `Fetch` when the call is expensive, can block, or can fail.
- Name constants for their role, not their value, in MixedCaps.
- Keep types out of variable names: `users`, not `userSlice`; `userCount`, not `numUsers`.
- Name length grows with scope. Don't drop letters to save typing: `Sandbox`, not `Sbx`.
- Avoid package names that collide with common locals: `usercount`, not `count`.

### Errors

- Return the `error` type from exported functions, never a concrete error type. A nil concrete pointer becomes a non-nil interface.
- Treat non-error results as unspecified when `err != nil`, unless documented otherwise.
- Return an extra value (`ok` or `error`) rather than an in-band sentinel like `-1`.
- Discarding an error with `_` needs a comment saying why it's safe.
- Reserve `panic` for what can't be handled as an error. `MustXYZ` helpers run only at program startup.

### Design

- Don't create an interface until there is a real need. The consumer defines it, and it stays unexported if only used inside the package.
- Prefer synchronous functions; let the caller add concurrency.
- Make goroutine lifetimes obvious. If they aren't, document when and why each goroutine exits.
- `context.Context` is the first parameter. Don't store it in a struct or define custom context types.
- Don't copy a value whose methods are on `*T`, such as `sync.Mutex` or `bytes.Buffer`.
- Don't pass pointers just to save a few bytes. When unsure about a receiver, use a pointer, and don't mix pointer and value receivers on a type.
- Don't use generics to build DSLs, or just because the code is an algorithm or data structure.
- Prefer a nil local slice over an empty one. Don't make callers distinguish nil from empty.

### Formatting gofmt won't fix

- Keep function signatures, `if` conditions, and `switch`/`case` lines on one line. Extract locals or use an option struct instead of breaking them.
- Name fields in struct literals for types from other packages.
- Use `%q` for human-facing output when a value could be empty or contain control characters.
- Don't end a `switch` clause with an unlabeled `break`.

### Tests

- Use the standard `testing` package. No assertion libraries.
- Failure messages name the function and print got before want: `YourFunc(%v) = %v, want %v`. Label diffs, e.g. `(-want +got)`.
- Prefer `t.Error` over `t.Fatal` for mismatches, so the test keeps reporting.
- Check errors with `errors.Is` or `errors.As`, never by comparing error strings.
- Table tests get named subtests, without slashes or spaces, not table indexes.
