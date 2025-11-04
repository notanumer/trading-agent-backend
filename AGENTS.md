# AI Agent Instructions for Go Projects (Final)

This document provides comprehensive guidelines for AI agents working with Go codebases, covering code style, best practices, project structure, and modern tooling.

## 1. Code Style Guidelines

### Formatting and Imports
- Format code using `gofmt`
- Organize imports with `goimports`: standard library → internal packages → external packages; separate groups with blank lines
- Group similar declarations: `import(...)`, `const(...)`, `var(...)`, `type(...)`

### Consistency
- Maintain consistent style throughout the codebase
- Apply large-scale alignments at package level and above, not point changes

### Naming Conventions
- Error types: `...Error` (e.g., `CompareError`, `UserError`)
- Error values: `Err...` (e.g., `ErrUserNotFound`, `ErrServiceUnavailable`)
- Never shadow reserved names (e.g., variable named `error` is forbidden)
- Packages: lowercase, no underscores, short names, not plural, avoid `util`/`common`/`lib`, no conflicts with stdlib, no need for aliases
- Import aliases: only for external packages
- Top-level variables: use `var` without explicit type; leading underscore (`_name`). Specify type only if it differs from expression

### Line Length and Blank Lines
- Hard limit: 180 characters
- Blank lines: between logical blocks; after `case` (except immediate `return`/`break`/`continue`); after error handling; in large functions to separate unrelated parts

### Variables
- Narrow scope when possible. Prefer `if err := ...; err != nil { return err }`, but don't increase nesting
- Use raw string literals (`` `...` ``) for readability and multiline strings

### Comments
- Prefer self-documenting code; avoid comments in most cases. Add comments only when the logic is non-obvious, complex, has nuances/pitfalls, or documents intentional workarounds/hacks
- Write code comments in the project's primary language (often Russian for internal projects)

### Constants and Magic Values
- Avoid magic numbers and strings; extract them into constants with meaningful names

### Functions and Control Flow
- Function length limit: 130 lines
- Max number of return values: 3
- Max number of function parameters: 6
- Ordering: by usage/call order; group by structs; public above private; constructor right after type and before methods
- Named returns: acceptable for compactness and readability (especially with multiple similar results); avoid in large/branched functions
- Minimize nesting; avoid unnecessary `else` through default values and early `return`/`continue`

### Structs
- Embedding at top of struct, separated by blank line
- Struct initialization: use field names
- Zero initialization: use `var T`, not `T{}` in assignment

### Serialization and DB Tags
- Always specify tags for `db`/`json`/`xml`/`yaml`

### Pointers and Collections
- Pointer creation: `&T{...}`, not `new(T)` + assignments
- Empty map: `make(map[K]V)`; with known elements: `map[K]V{ k: v }`
- For small known sets: use literals; with known capacity: use `make` with capacity

## 2. Best Practices

### Interfaces
- Create only where implementation substitution is needed (usually for testing repositories/services)
- Don't create interfaces for helpers. Don't invent DI
- Implementation naming: `UserRepository` + `SqlUserRepository` (not `IUserRepository` + `UserRepository`)

### Receivers and Pointers
- Prefer pointer receivers: fewer copies, struct mutability. Value receiver for read-only
- Check pointer fields for `nil` before use
- If method requires pointer receiver, use pointer to value for interface; simple value causes interface mismatch error

### Zero Values
- `sync.Mutex`/`RWMutex` don't require explicit initialization
- Nil slices are safe: `len==0`, `append` works without panic

### Embedding
- Only in extreme cases and only for models. Don't embed `sync.Mutex`; keep as field (`mx sync.Mutex`)
- Embedding shouldn't be cosmetic, shouldn't expose internal methods, change API/copy semantics

### Error Handling
- Top-level package errors (`var ErrX = errors.New("...")`); check with `errors.Is`
- For dynamic text: custom error type; check with `errors.As`
- Wrap with context: `fmt.Errorf("...: %w", err)`. Don't lose original error
- Avoid `panic`; `Must*` only during application initialization
- Type assertions: use ok-idiom `v, ok := i.(T)`

### Global Objects
- Avoid `init()`; use lazy singletons. If `init` unavoidable: no dependency order, no IO/network/environment changes

### Channels
- Primarily buffer 0 or 1; large buffers only with justification

### Patterns
- Option parameters instead of long signatures (`Open(addr, opts ...Option)`)
- Enum: custom type + `iota`; zero value should be `Undefined`

### Time
- `time.Time` for moments; `time.Duration` for intervals
- Measurements: `time.Since(start)`; until event: `time.Until(t)`
- Accept intervals as `time.Duration` (`poll(10*time.Second)`)

### Varargs
- When passing slice, use `args...` (not `args`)

### Slices and Maps
- Copy incoming slices/maps when storing in internal state and before returning
- In-place filter: `b := a[:0]`; then `append`; remember that `a` will change

### Performance and Safety
- `strconv.Itoa` faster than `fmt.Sprint` for `int→string`
- `[]byte` from constant: allocate once and reuse
- Pre-allocate capacity for slices and size for maps: `make([]T, 0, n)`, `make(map[K]V, n)`
- Use `defer` for guaranteed `Unlock`/`Close`
- Case-insensitive string comparison: `strings.EqualFold`

### Shutdown and Logging
- Implement graceful shutdown: stop API, wait for completion, close outgoing connections
- Structured logging (prefer `zap` or similar), not `log.Println`

## 3. Modern Libraries and Tools

### Recommended Libraries
- **Logging**: `gitea.gospodaprogrammisty.ru/Go/servicelib/swlog` for structured logging
- **Database Connections**: `gitea.gospodaprogrammisty.ru/Go/servicelib/db` (priority connectors)
  - Postgres: `db.NewPgx(ctx, connStr, db.WithTraces())`
  - MongoDB, ClickHouse, Redis: corresponding helpers from `servicelib/db`
- **Postgres Helpers**: `github.com/jmoiron/sqlx` for extended SQL operations
- **Scylla/Cassandra**: `github.com/scylladb/gocqlx/v2` (built on `github.com/gocql/gocql`)
- **API Generation**: `github.com/oapi-codegen/oapi-codegen/v2` for HTTP services
- **Migrations**: `github.com/pressly/goose/v3` via `servicelib/db.Migrate` wrapper
- **Testing**: `github.com/stretchr/testify`

### Configuration Management
- Use environment variables instead of JSON configs per environment
- Use `smartway-smartchat.gitlab.yandexcloud.net/smart-chat/smartchat-lib/scconfig` for configuration management
- Describe `Settings` structs with strict types and `config:"..."` tags (env, sep, transform, etc.)
- Load configuration: `scconfig.LoadConfig(&settings)`. After loading, derive additional fields (e.g., DSN for Postgres)
- Common environment variables patterns:
  - Postgres: `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_SESSION_USERNAME`, `DB_SESSION_PASSWORD` (often with `url_escape` transform)
  - Scylla: `SCYLLA_CONTACT_POINTS` (separator `,`), `SCYLLA_USERNAME`, `SCYLLA_PASSWORD`, `SCYLLA_DC`

## 4. Service Structure

### General Repository Structure
- Root: `main.go` (entry), `app.go` (initialization), `.golangci.yaml` (linter), `Dockerfile`, `.gitignore`, `config/` (models.go, reader.go)
- `api/`: `server.go` (http.Server, mux.Router, registration), `handlers/` (endpoints)
- `database/`: migrations + репозиторные интерфейсы (см. детальный раздел ниже) и подпакеты с реализациями
- `cluster/`: clients and models for accessing other microservices (e.g., `booking/` with `models.go` and `client.go`)
- `{providerName}/`: integration with external provider; may be absent
- `service/`: business logic
- `models/`: optional general-purpose domain models

### Entry Point (root)
- Required files: `.gitignore`, `Dockerfile`, `.golangci.yaml`, `app.go`, `main.go`
- Entry point (`main.go`): create logger (`swlog.NewLogger(appName)`), load settings, initialize connections (with timeouts), assemble repositories/services, and start/stop application predictably

### API Structure
- In `api/server.go`: create/configure `http.Server` and router; register all handlers from `api/handlers`
- For HTTP services:
  - Store OpenAPI specifications in `openapi/` (separation into `components/`, `paths/` is encouraged)
  - Generate code into `gen/` directory using configuration files
  - Serve Swagger UI at `/v1/swaggerui`, base API prefix — `/v1`
  - Required endpoints: `/health`, `/metrics` (Prometheus), `/debug/pprof`
  - Use robust routers (e.g., `github.com/gorilla/mux`) with `handlers.RecoveryHandler` and `swlog` logging

### Database Structure
- Store migrations and repository layer artifacts
- Repository pattern (MANDATORY):
  - Root package `database`: only repository interfaces + shared errors/constants. No driver-specific code or SQL here.
  - One file per domain entity: `<entity>.go` with `<Entity>Repository` interface (e.g. `appeal.go` → `AppealRepository`).
  - Interface methods describe business operations (e.g. `GetByID`, `GetTags`, `GetEventCount`, `GetByParams`). Avoid transport / driver details.
  - Implementations live in subpackages: `database/<entity>` (e.g. `database/appeal`).
  - Implementation type naming: `<Backend><Entity>Repository` (e.g. `PostgresAppealRepository`, `ScyllaMessageRepository`, `MemoryAppealRepository`). Multiple backends allowed side‑by‑side.
  - Typical files: `repo.go` + query folders (`sql/`, `cql/`, etc.) loaded via `//go:embed`.
  - Organize queries by purpose; allow composable base fragments (`*_base.sql`).
  - Service layer depends ONLY on interfaces from root `database`; never import implementation subpackages inside business logic.
  - Wiring / instantiation happens in `app.go` (or dedicated factory) converting concrete impls to interfaces.
  - Implementation subpackages must not import each other.
  - Shared helpers (cross-interface errors/constants) may reside in root (`errors.go`, `constants.go`).
  - Example reference: `.qmt-service/database` (`appeal.go` + `appeal/`).
- **Data access style**: Avoid full ORMs (GORM, ent, etc.) and heavy query builders (squirrel, etc.). Prefer explicit SQL/CQL scripts stored as files (embedded) + lightweight helpers (`sqlx` scanning). Keep queries transparent and reviewable.
- **Postgres**: Connect via `servicelib/db.NewPgx`, configure connection pools (`SetMaxOpenConns`, etc.). Use `sqlx` for queries
- **Scylla**: Use connectors over `gocql` with `gocqlx.WrapSession`; explicit `Open()`/`Close()` methods
- **Redis**: Use `servicelib/db` helpers when available; alternatively `github.com/redis/go-redis/v9` with `Ping` and initialization timeouts
- **Migrations**: Use `servicelib/db.Migrate(fs, log, db, path)` with versioning table `db_version`. Store migrations in `database/migrations/`

### Cluster Structure
- Cluster clients and their models: by subpackages

### Provider Structure (`{providerName}`)
- All logic and models: strictly in provider package; subpackages only for intra-package use

### Service Structure
- Root `service` package: only service interfaces (and optional lightweight shared DTOs). No concrete logic, no external driver imports.
- Implementations: one subpackage per domain concern (`service/<domain>`). Implementation type typically named `Service` with constructor `NewService(...) *Service`.
- Keep implementation struct minimal: injected dependencies only (context, logger, repository interfaces, other services). No exported fields.
- Dedicated `mapper` subpackage for pure transformation / aggregation helpers (stateless, test-friendly). Business packages may depend on `service/mapper`; avoid lateral deps between sibling service subpackages.
- Mapping / diff steps split into small helpers (`create*`, `mapNew*`, etc.) each returning domain slices or maps; pure functions preferred for testability.
- Repositories accept context derived from a base context wrapper; avoid passing raw background contexts.
- No cross-imports between implementation subpackages (each only imports: root `service` (interfaces), `service/mapper`, repositories’ interface packages, and external libs).
- Do not leak implementation types outside their subpackage; expose only interfaces from root `service`.

### Models (optional)
- Extract domain models used in multiple places (database, service, etc.)

## 5. API Generation with oapi-codegen

### Configuration
- Configuration file (e.g., `oapi-codegen.yaml`) minimum settings:
  - Enable server generation, models, and embedded spec
  - Specify output file and options
- Build specification: bundle OpenAPI specs into distribution format
- Generate code: run oapi-codegen with configuration (integrate into build process)
- Implementation: implement server interfaces in `api/handlers/` and register via `gen.HandlerWithOptions` with `BaseURL: "/v1"` and `ErrorHandlerFunc` using `swlog`

## 6. Testing Best Practices

### Testing Framework
- Use `github.com/stretchr/testify` (`require`/`assert`)
- Table-driven tests: slice of test cases, `t.Run(name, ...)`, clear and readable test case names
- Parallelism: `t.Parallel()` where safe; avoid race conditions on shared state
- For databases: mock repositories through interfaces when possible; alternatively use schema isolation/fixtures

## 7. Advanced Practices and Style

### Logging and Monitoring
- Structured logging using `swlog` in project's primary language
- For exceptional situations, use `ErrorEntryf(...).WithTags(...).Write()`
- Implement proper metrics and monitoring endpoints

### External Operations
- All external I/O operations (DB, Redis, Scylla, HTTP) should use context and timeouts
  - Database initialization: ~15s timeout
  - Long-running queries: ~1m timeout (justify longer timeouts)
- Avoid "magic numbers": extract timeouts, pool sizes, chunk limits into constants

### Database-Specific Practices
- **Scylla**: Limit parameters in IN queries; use chunking (e.g., 100 items) and prepared/bind queries via `gocqlx`
- After destructive operations, check invariants when meaningful (e.g., recalculate message counts, update `last_message_id`)
- Store SQL/CQL conveniently near repositories and connect via `go:embed`
- Properly close resources (`Stop()`/`defer`), log all errors using `swlog`

### Application Lifecycle
- Implement proper lifecycle management: `InitDatabase` → `InitServices` → `Start/Stop`
- HTTP servers with recovery middleware, health checks, metrics, debug endpoints, and API documentation
- Repository layer: connect databases through proper abstractions; bind in service layer
- Configuration through `scconfig`; construct connection strings (DSN) after loading settings

## 8. CLI and Background Tasks

### CLI Applications
- Support task-based execution through flags (e.g., `-task` parameter)
- Use same connection patterns as HTTP services (same connectors from `servicelib/db`)
- Implement proper initialization with timeouts and error handling
- Include protective logging and post-condition checks in business flows using `swlog`

## AI Agent Instructions

When working with Go codebases:

1. **Always follow the code style guidelines above**
2. **Maintain consistency with existing code patterns**
3. **Apply best practices for error handling, performance, and safety**
4. **Follow the established project structure**
5. **Use proper naming conventions throughout**
6. **Ensure proper formatting and import organization**
7. **Write clean, readable, and maintainable code**
8. **Consider performance implications of your changes**
9. **Implement proper error handling and logging**
10. **Follow Go idioms and conventions**
11. **Use modern tooling and libraries appropriately**
12. **Implement proper configuration management with `scconfig`**
13. **Ensure robust database connections using `servicelib/db` connectors**
14. **Follow API generation and testing best practices**
15. **Use `swlog` for all logging throughout the application**

Remember: Code should be not just functional, but also maintainable, readable, and consistent with Go best practices and the existing codebase style. Always consider the broader architectural patterns and use appropriate modern tooling for the specific use case.
