# CLI Commands (Cobra)

Commands live in `cmd/`, one file per command. Commands that connect to
the desk use the Cobra lifecycle hooks like this:

| Hook | Where | Responsibility |
|------|-------|----------------|
| `PersistentPreRun` | `cmd/root.go` | Enable the Bluetooth adapter (`bluetooth.DefaultAdapter`) for every subcommand |
| `PreRun` | subcommand | Validate args, then connect with `connectDesk(cmd)` (`cmd/config.go`) |
| `Run` | subcommand | Do the operation. For movement, wrap `cmd.Context()` in a timeout (e.g. 60s) |
| `PostRun` | subcommand | Call `j.Disconnect()` |

Follow `cmd/gotoHeight.go` when adding a command that connects to the desk.

Commands that connect to the desk register a local `--address` flag with
`addAddressFlag(cmd)` in their `init()`. `connectDesk` takes the address
from `--address` if set, otherwise from the `address` field of
`~/.deskctl.yaml` (read with Viper). No other sources such as env vars
are used.

Commands that don't connect to a desk (`devices`, `version`) only
define `Run`. They take no `--address` flag and never call `jiecang.Init`.
Follow `cmd/listDevices.go` for these.

Errors in every hook are printed to stderr, followed by `os.Exit(1)` (see
[go-conventions.md](go-conventions.md#error-handling)).
