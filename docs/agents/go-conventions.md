# Go Conventions

## Formatting & linting

- Format with `make fmt` (`gofmt -s`).
- CI also runs `golangci-lint` v2.1. If you have it installed, run
  `golangci-lint run` locally before pushing.
- The Go version lives in `go.mod`. Don't hard-code it anywhere else.

## Naming

- Constants use CamelCase, with initialisms in all caps: `LierdaDeviceID`,
  not `LIERDA_DEVICE_ID` or `LierdaDeviceId`.
- Known exception: `BLEDeviceId`, `BLECharDataInId` and `BLECharDataOutId`
  in `pkg/jiecang/jiecang.go` predate this rule. Don't copy the `Id`
  spelling into new code.

## Error handling

- **`pkg/jiecang`** (library): wrap errors with context and return them,
  e.g. `return fmt.Errorf("failed to save memory%d: %w", n, err)`. Never
  call `os.Exit` or `log.Fatal` in the library.
- **`cmd/`** (CLI): print to stderr, then exit non-zero:

  ```go
  fmt.Fprintf(os.Stderr, "Failed to go to height: %v\n", err)
  os.Exit(1)
  ```
