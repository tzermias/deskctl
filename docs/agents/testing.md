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

- Tests can't touch real Bluetooth hardware. Test the pure functions
  (message parsing and validation in `pkg/jiecang`) with raw byte slices.
  See `pkg/jiecang/common_test.go` for examples.
- `make test` runs `go mod download` and `go mod tidy` first, so it needs
  network access. Offline, `go test ./...` is an accepted substitute (see
  [AGENTS.md](../../AGENTS.md#commands)).
