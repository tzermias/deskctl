# AGENTS.md

deskctl is a Go CLI (Cobra) for controlling Jiecang-controlled standing desks over Bluetooth LE.

## Commands

```bash
make fmt     # gofmt -s -w .
make vet     # go vet ./...
make test    # go mod download/tidy, then go test -v ./...
make build   # fmt -> vet -> test -> goreleaser snapshot build into ./dist
```

Code must pass `make fmt`, `make vet` and `make test` before every commit.

## Hard rules

- Never commit to `main`. Work on a `feature/<task-name-kebab-case>` branch.

## Further reading

Read the relevant file before working in that area:

- [Git workflow & commit messages](docs/agents/git-workflow.md)
- [Go conventions (naming, errors, linting)](docs/agents/go-conventions.md)
- [Testing patterns](docs/agents/testing.md)
- [CLI commands (Cobra)](docs/agents/cli-commands.md)
- [BLE protocol & `pkg/jiecang` internals](docs/agents/ble-protocol.md)
- [Build, release & CI](docs/agents/build-release.md)
