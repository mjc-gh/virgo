# AGENTS.md

This file provides guidance for AI coding agents working in the virgo codebase.

Always check tests and linting after making changes.

## Project Overview

virgo is a Go-based tool converting website to Markdown. It uses the Chrome
DevTools Protocol via `chromedp` to automate browser interactions for security
analysis. The project produces two binaries: `virgo` (CLI) and `virgo-web` (REST API).

**Go Version**: 1.24.4
**Module**: `github.com/mjc-gh/virgo`

## Build Commands

```bash
make build.cli      # Build CLI to build/virgo
make build.web      # Build web server to build/virgo-web
make go.get         # Download dependencies
make go.tidy        # Tidy go.mod
```

## Test Commands

```bash
make test           # Run linting then all tests
go test ./...       # Run all tests without linting
go test -v ./...    # Run all tests with verbose output

# Run a single test by name
go test -v ./engine -run TestNewTask
go test -v ./... -run TestPerformTaskUnknownType

# Run tests matching a pattern
go test -v ./... -run "TestTask.*"

# Running tests with remote Chrome DevTools (required for browser tests)
CHROMEDP_REMOTE_URL="http://127.0.0.1:9222" make test
CHROMEDP_REMOTE_URL="http://127.0.0.1:9222" go test -v ./...

# To run with Docker headless-shell:
docker run -d -p 9222:9222 --rm chromedp/headless-shell
CHROMEDP_REMOTE_URL="http://127.0.0.1:9222" make test
```

## Lint Commands

```bash
make check          # Run golangci-lint
```

The project uses `golangci-lint` v2 with nearly all linters enabled. See
`.golangci.yml` for the full configuration. Code must pass `make check` formatting.

## Project Structure

```
virgo/
├── cmd/virgo/main.go         # CLI entry point
├── cmd/virgo-web/main.go     # Web server entry point
├── engine/                   # Core analysis engine (crawler, tasks, errors)
├── internal/browser/         # Browser configuration/profiles
├── internal/pagetest/        # Webpage test utilities and fixtures
├── internal/rest/            # REST API server
└── logger.go                 # Logging setup (root package)
```

## Code Style Guidelines

### Import Ordering

Organize imports in three groups separated by blank lines:
1. Standard library packages
2. Third-party packages
3. Internal project packages

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, single word | `engine`, `browser`, `rest` |
| Variables/functions | camelCase | `userAgent`, `winWidth` |
| Exported names | PascalCase | `NewCrawler`, `AnalyzeResult` |
| Constants | SCREAMING_SNAKE_CASE | `SIZE_LARGE`, `PROFILE_DESKTOP` |
| Error variables | `Err` prefix + PascalCase | `ErrNoCrawlerVisit` |
| JSON struct tags | snake_case | `json:"requested_url"` |
| YAML struct tags | camelCase | `yaml:"browserProfile"` |

### Error Handling

1. **Define sentinel errors** at package level:
```go
var ErrNoCrawlerVisit = errors.New("no visit from crawler")
```

2. **Wrap errors with context**:
```go
return fmt.Errorf("create output file: %w", err)
```

3. **Log errors with zerolog**:
```go
logger.Warn().Err(err).Msg("file close error")
```

### Functional Options Pattern

Use functional options for configurable types:

```go
type Option func(*Engine)

func WithRemoteAllocator(host string, port int) Option {
    return func(e *Engine) {
        host := net.JoinHostPort(host, strconv.Itoa(port))
        e.config.remoteURL = fmt.Sprintf("http://%s/json/version", host)
    }
}
```

### Testing Patterns

1. **Use table-driven tests**:
```go
tests := []struct {
    name           string
    action         string
    expectedAction string
}{
    {name: "basic task", action: "navigate", expectedAction: "navigate"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        task := NewTask(tt.action, tt.input)
        assert.Equal(t, tt.expectedAction, task.action)
    })
}
```

2. **Use testify assertions**:
```go
assert.Equal(t, expected, actual)
require.NoError(t, err)  // Fails test immediately
require.Error(t, err)
```

3. **Mark parallel-safe tests** with `t.Parallel()`

4. **Use test utilities** from `internal/pagetest/`:
   - `NewTestWebServer()` - HTTP test server with embedded testdata
   - `NewTestContext()` - Browser context (respects env vars for remote/headfull)
   - `FindByID[T]()` - Generic helper for finding structs by ID

### Test Fixtures

HTML test fixtures are stored in `internal/pagetest/testdata/[fixture-name]/` and loaded via `pagetest.NewTestWebServer("[fixture-name]")`.

When adding new test cases:
1. Add a new `<div>` or `<section>` to the appropriate fixture file (e.g., `index.html`)
2. Include a descriptive comment explaining what edge case it tests
3. Reference the fixture in your test with `pagetest.NewTestWebServer()`
4. Use assertions to validate the specific behavior being tested

Example: Testing inline content followed by a block element:
```html
<!-- Inline content immediately before block element -->
<h3>Inline to Block Transition</h3>
<div>
    <span>text before heading</span>
    <h1>Inline to Block Test</h1>
</div>
```

## Git Commit Guidelines

Follow semantic commit format with issue references:

**Format:** `<type>: <subject>`

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `test:` - Test additions or updates
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks
- `style:` - Formatting changes

**Issue References:**
- Fixes an issue: `Fixes #123` or `fixes #123`
- Closes an issue: `Closes #123` or `closes #123`

**Examples:**
```
fix: ensure headers always start on new lines (fixes #4)
feat: add links task type with fuzzy search capability
refactor: implement AST-based markdown conversion
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/chromedp/chromedp` | Chrome DevTools Protocol automation |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/urfave/cli/v3` | CLI framework |
| `github.com/gin-gonic/gin` | HTTP web framework |
| `github.com/stretchr/testify` | Test assertions |
