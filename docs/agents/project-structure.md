# Project Layout & Dependencies

```
deskctl/
├── main.go              # Entry point, calls cmd.Execute(ctx)
├── cmd/                 # CLI commands (Cobra), one file per command
│   ├── root.go          # Root command, enables the Bluetooth adapter
│   ├── updown.go        # up / down
│   ├── gotoHeight.go    # goto-height
│   ├── gotoMemory.go    # goto-memory
│   ├── listDevices.go   # devices (BLE scan, no connection)
│   └── version.go       # version
├── pkg/jiecang/         # Desk control library
│   ├── jiecang.go       # Jiecang struct, BLE connection, commands
│   ├── common.go        # Message validation (isValidData)
│   ├── height.go        # Height operations
│   ├── memory.go        # Memory presets
│   └── *_test.go        # Unit tests, next to the code they test
├── hack/                # Dev utilities. hack/test.go is a scratch program,
│                        # but `go vet ./...` still checks it.
├── Makefile
└── .goreleaser.yaml
```

## Dependencies

- `github.com/spf13/cobra`: CLI framework
- `tinygo.org/x/bluetooth`: BLE connectivity
- `github.com/stretchr/testify`: test assertions

Dependencies are Go modules. Run `make deps` after changing imports.
