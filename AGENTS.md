# AGENTS.md

This file provides guidance for AI coding agents working with the deskctl codebase.

## Project Overview

**deskctl** is a Go CLI tool for controlling Bluetooth-enabled standing desks with Jiecang controllers (equipped with Lierda LSD4BT-E95ASTD001 BLE modules). It supports movement commands, height control, and memory presets.

## Build Commands

```bash
make deps     # Download and tidy dependencies
make fmt      # Format code with gofmt -s
make vet      # Run go vet static analysis
make test     # Run all tests (go test -v ./...)
make build    # Full build: fmt -> vet -> test -> goreleaser
make clean    # Remove dist/ and bin/ directories
```

The binary is output to `./bin/deskctl`.

## Git Workflow & Standards

* *Feature Branches*: ALWAYS create a feature branch for changes: feature/[task-name-kebab-case].
* *No Direct Commits*: NEVER work directly on main or master.
* *No Build Artifacts*: NEVER commit build artifacts (bin/, dist/), binaries, or temporary files (.claude/).
        These are excluded in .gitignore and should never be part of commits or PRs.
* *Commit Messages*:
        Limit subject to 50 characters, capitalize, no period.
        Use imperative mood ("Add", "Fix", "Update").
        Wrap body at 72 characters.
        Separate subject from body with a blank line.


## Testing

Run tests with:
```bash
make test
```

Tests use the [testify](https://github.com/stretchr/testify) assertion library. Test files follow the `*_test.go` convention and are co-located with the code they test.

Test patterns used:
- Table-driven tests with named cases
- `assert.Equal()` for assertions

## Code Style

- **Go version:** 1.22.2
- **Formatting:** `gofmt -s` (run via `make fmt`)
- **Linting:** golangci-lint v2.1 (in CI), `go vet` locally
- Code must pass `make fmt`, `make vet`, and `make test` before commits

### Naming Conventions

- Package names: lowercase (`jiecang`, `cmd`)
- Exported functions: CamelCase (`GoToHeight`, `SaveMemory1`)
- Constants: UPPERCASE (`BLEDeviceId`, `LierdaDeviceID`)
- Unexported functions: camelCase (`isValidData`, `readHeight`)

### Error Handling

- Use simple `if err != nil` checks
- Return early on validation failures
- Log errors with the standard `log` package or `fmt.Printf`

## Project Structure

```
deskctl/
├── cmd/                  # CLI command implementations (Cobra)
│   ├── root.go          # Root command, Bluetooth adapter init
│   ├── updown.go        # Up/down movement commands
│   ├── gotoHeight.go    # Go to specific height
│   ├── gotoMemory.go    # Go to memory preset
│   ├── listDevices.go   # Scan for devices
│   └── version.go       # Version command
├── pkg/jiecang/         # Core desk control library
│   ├── jiecang.go       # Main Jiecang struct, BLE connection
│   ├── common.go        # Message validation utilities
│   ├── height.go        # Height operations
│   ├── memory.go        # Memory preset operations
│   └── *_test.go        # Unit tests
├── hack/                # Development/testing utilities
├── main.go              # Entry point
├── Makefile             # Build automation
└── .goreleaser.yaml     # Cross-platform release config
```

## Key Patterns

### CLI Commands (Cobra)

Commands are in `cmd/` and follow the Cobra pattern:
- Use `PreRun` for setup (e.g., Bluetooth adapter initialization)
- Use `Run` for main logic
- Use `PostRun` for cleanup

### BLE Message Protocol

Messages follow the Jiecang UART protocol:
- Format: `[0xf1, 0xf1, command, length, data..., checksum, 0x7e]`
- Responses start with `0xf2, 0xf2`
- Checksums are validated in `pkg/jiecang/common.go`

### Thread Safety

The `Jiecang` struct uses `sync.RWMutex` to protect shared state (currentHeight, height limits, etc.).

## Dependencies

Key dependencies:
- `github.com/spf13/cobra` - CLI framework
- `github.com/stretchr/testify` - Testing assertions
- `tinygo.org/x/bluetooth` - Bluetooth connectivity

Managed via Go modules. Run `make deps` to update.

## CI/CD

GitHub Actions workflows:
- **build.yml**: Runs on push/PR to main - fmt, lint, vet, test
- **release.yml**: Triggers on `v*` tags - creates cross-platform releases


