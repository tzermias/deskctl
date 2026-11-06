# Testing

- Use [testify](https://github.com/stretchr/testify) for assertions.
- Write table-driven tests. Each case has a `name` field, and that name is
  passed as the assertion message:

  ```go
  tests := []struct {
      name           string // Name of the testcase
      input          []byte
      expectedResult bool
  }{
      {name: "Valid message for height", input: ..., expectedResult: true},
  }
  for _, test := range tests {
      result := isValidData(test.input)
      assert.Equal(t, test.expectedResult, result, test.name)
  }
  ```

- Tests can't touch real Bluetooth hardware. Test pure functions
  (message parsing and validation) with raw byte slices, see
  `pkg/jiecang/common_test.go`. Test anything that talks to the desk
  against the in-process fake desk below.
- Run tests with `-race` (`make test` does). Notification callbacks run
  on another goroutine.
- `make test` runs `go mod download` and `go mod tidy` first, so it needs
  network access. Offline, `go test -race ./...` is an accepted substitute
  (see [AGENTS.md](../../AGENTS.md#commands)).

## Fake desk (`pkg/jiecang/fakedesk_test.go`)

`fakeDesk` implements the `characteristic` and `disconnecter` seam (see
[ble-protocol.md](ble-protocol.md#test-seam)). It logs every command
written to it and answers with raw `F2F2` frames through the callback
captured by `EnableNotifications`, like the real controller: `07` returns
the presets and the height, `0C` the range, movement commands move the
desk 1 cm and report the new height, saves are not acknowledged.

- `newTestDesk(t, fd)` runs `newJiecang(fd, fd, fd)`, sets
  `pollInterval` to 1ms so movement tests finish in milliseconds, and
  clears the write log. Pass `nil` for the default desk (80 cm, range
  65-130 cm, presets 75/110/90 cm).
- `frame(cmd, data...)` builds a valid controller reply,
  `command(cmd, data...)` a valid `F1F1` command and `heightFrame(mm)` a
  `01` height report. Use them instead of hand-computing checksums.
- Knobs on `fakeDesk`: `writeErr`/`readErr`/`notifyErr`/`disconnectErr`
  inject errors (`failAt` delays `writeErr` to the Nth write), `batch`
  sends all replies in one notification, `async` sends them from a
  goroutine, and `settings` adds the RJ-45 settings replies.
- Read fake state with the locking accessors (`Writes()`, `Height()`,
  `Preset(n)`, `Stopped()`), and `currentHeight(j)` / `preset(j, name)`
  for the `Jiecang` side. Call `fd.wait()` before asserting when `async`
  is set.
- The fake drops commands with a bad checksum, like the desk does.

```go
fd := newFakeDesk()
j := newTestDesk(t, fd)
require.NoError(t, j.Up(), test.name)
assert.Equal(t, [][]byte{commands["up"]}, fd.Writes(), test.name)
assert.Equal(t, uint8(81), currentHeight(j), test.name)
```

**Rule:** new or changed code that sends commands to the desk or handles
its replies needs a fake-desk test. Extend `fakeDesk.handle` when a new
command needs a reply. When a test pins known-wrong behaviour, add a
`// TODO:` comment explaining the bug.

## Fuzzing

`FuzzIsValidData` and `FuzzCharacteristicReceiver` check that malformed
input never panics:

```bash
go test -run '^$' -fuzz=FuzzCharacteristicReceiver -fuzztime=30s ./pkg/jiecang
```
