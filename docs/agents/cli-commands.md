# CLI Commands (Cobra)

Commands live in `cmd/`, one file per command. They use the Cobra
lifecycle hooks like this:

| Hook | Where | Responsibility |
|------|-------|----------------|
| `PersistentPreRun` | `cmd/root.go` | Enable the Bluetooth adapter (`bluetooth.DefaultAdapter`) for every subcommand |
| `PreRun` | subcommand | Validate args and the `--address` MAC, then connect with `jiecang.Init(adapter, ...)` |
| `Run` | subcommand | Do the operation. For movement, wrap `cmd.Context()` in a timeout (e.g. 60s) |
| `PostRun` | subcommand | Call `j.Disconnect()` |

Follow `cmd/gotoHeight.go` when adding a command that connects to the desk.

Errors in every hook are printed to stderr, followed by `os.Exit(1)` (see
[go-conventions.md](go-conventions.md#error-handling)).
