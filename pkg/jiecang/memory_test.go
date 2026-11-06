package jiecang

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadMemoryPreset(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		input          []byte // Input
		expectedHeight uint8  // Expected result of function
	}{
		{
			name:           "Correct height with incorrect checksum",
			input:          []byte{0xf2, 0xf2, 0x25, 0x02, 0x04, 0x4e, 0x79, 0x7e},
			expectedHeight: 110,
		},
		{
			name:           "Preset of 750 mm",
			input:          frame(0x26, 0x02, 0xee),
			expectedHeight: 75,
		},
		{
			// TODO: integer division, 1105 mm should round to 111 cm.
			name:           "Preset is truncated, not rounded",
			input:          frame(0x27, 0x04, 0x51),
			expectedHeight: 110,
		},
		{
			name:           "Wrong data length",
			input:          frame(0x25, 0x04),
			expectedHeight: 0,
		},
	}

	for _, test := range tests {
		result := readMemoryPreset(test.input)
		assert.Equal(t, test.expectedHeight, result, test.name)
	}
}

func TestGoToMemory(t *testing.T) {
	errBLE := errors.New("ble failure")
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name           string // Name of the testcase
		setup          func(fd *fakeDesk, j *Jiecang)
		ctx            context.Context
		memoryNum      int
		expectedErrMsg string   // Expected error substring, empty for success
		expectedErr    error    // Expected wrapped error
		expectedHeight uint8    // Expected desk height in cm afterwards
		expectedCmd    []byte   // Every write must be this command
		expectedWrites int      // Exact number of writes, -1 for "at least one"
		notExpected    [][]byte // Commands that must not have been sent
	}{
		{name: "Memory 0 is invalid", memoryNum: 0, expectedErrMsg: "invalid memory number 0", expectedHeight: 80},
		{name: "Memory 4 is invalid", memoryNum: 4, expectedErrMsg: "invalid memory number 4", expectedHeight: 80},
		{
			name:           "Reaches memory 1",
			memoryNum:      1,
			expectedHeight: 75,
			expectedCmd:    []byte{0xf1, 0xf1, 0x05, 0x00, 0x05, 0x7e},
			expectedWrites: -1,
		},
		{
			name:           "Reaches memory 2",
			memoryNum:      2,
			expectedHeight: 110,
			expectedCmd:    []byte{0xf1, 0xf1, 0x06, 0x00, 0x06, 0x7e},
			expectedWrites: -1,
		},
		{
			name:           "Reaches memory 3 with async replies",
			setup:          func(fd *fakeDesk, j *Jiecang) { fd.async = true },
			memoryNum:      3,
			expectedHeight: 90,
			expectedCmd:    []byte{0xf1, 0xf1, 0x27, 0x00, 0x27, 0x7e},
			expectedWrites: -1,
		},
		{
			// TODO: GoToMemory does not send stop on cancel, unlike
			// GoToHeight, so the desk may keep moving.
			name:           "Context cancelled sends no stop",
			setup:          func(fd *fakeDesk, j *Jiecang) { j.pollInterval = time.Hour },
			ctx:            cancelled,
			memoryNum:      2,
			expectedHeight: 81, // one step from the initial command
			expectedCmd:    commands["goto_memory2"],
			expectedWrites: 1,
			notExpected:    [][]byte{commands["stop"]},
		},
		{
			name:           "Write error is wrapped",
			setup:          func(fd *fakeDesk, j *Jiecang) { fd.setWriteErr(errBLE) },
			memoryNum:      1,
			expectedErr:    errBLE,
			expectedErrMsg: "failed to send goto memory1 command",
			expectedHeight: 80,
			expectedCmd:    commands["goto_memory1"],
			expectedWrites: 1,
		},
		{
			name: "Write error while polling is wrapped",
			setup: func(fd *fakeDesk, j *Jiecang) {
				fd.mu.Lock()
				fd.writeErr, fd.failAt = errBLE, 2
				fd.mu.Unlock()
			},
			memoryNum:      2,
			expectedErr:    errBLE,
			expectedErrMsg: "failed to send goto memory2 command",
			expectedHeight: 81,
			expectedCmd:    commands["goto_memory2"],
			expectedWrites: 2,
		},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		j := newTestDesk(t, fd)
		if test.setup != nil {
			test.setup(fd, j)
		}
		ctx := test.ctx
		if ctx == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
		}

		err := j.GoToMemory(ctx, test.memoryNum)
		fd.wait()

		if test.expectedErrMsg != "" {
			assert.ErrorContains(t, err, test.expectedErrMsg, test.name)
			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr, test.name)
			}
		} else {
			assert.NoError(t, err, test.name)
		}
		assert.Equal(t, int(test.expectedHeight)*10, fd.Height(), test.name)

		writes := fd.Writes()
		switch {
		case test.expectedCmd == nil:
			assert.Empty(t, writes, test.name)
		case test.expectedWrites < 0:
			assert.NotEmpty(t, writes, test.name)
		default:
			assert.Len(t, writes, test.expectedWrites, test.name)
		}
		for _, w := range writes {
			assert.Equal(t, test.expectedCmd, w, test.name)
		}
		for _, n := range test.notExpected {
			assert.NotContains(t, writes, n, test.name)
		}
	}
}

func TestGoToMemoryDeprecatedWrappers(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		call           func(j *Jiecang, ctx context.Context) error
		expectedHeight int // Expected fake desk height in mm
	}{
		{name: "GoToMemory1", call: (*Jiecang).GoToMemory1, expectedHeight: 750},
		{name: "GoToMemory2", call: (*Jiecang).GoToMemory2, expectedHeight: 1100},
		{name: "GoToMemory3", call: (*Jiecang).GoToMemory3, expectedHeight: 900},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		j := newTestDesk(t, fd)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		assert.NoError(t, test.call(j, ctx), test.name)
		cancel()
		assert.Equal(t, test.expectedHeight, fd.Height(), test.name)
	}
}

func TestSaveMemory(t *testing.T) {
	errBLE := errors.New("ble failure")
	tests := []struct {
		name           string // Name of the testcase
		save           func(j *Jiecang) error
		writeErr       error
		slot           int    // Preset slot expected to change, 0 for none
		expectedErrMsg string // Expected error substring, empty for success
		expected       []byte // Exact command bytes, nil for no writes
	}{
		{name: "Memory 0 is invalid", save: func(j *Jiecang) error { return j.SaveMemory(0) }, expectedErrMsg: "invalid memory number 0"},
		{name: "Memory 4 is invalid", save: func(j *Jiecang) error { return j.SaveMemory(4) }, expectedErrMsg: "invalid memory number 4"},
		{name: "Save memory 1", save: (*Jiecang).SaveMemory1, slot: 1, expected: []byte{0xf1, 0xf1, 0x03, 0x00, 0x03, 0x7e}},
		{name: "Save memory 2", save: (*Jiecang).SaveMemory2, slot: 2, expected: []byte{0xf1, 0xf1, 0x04, 0x00, 0x04, 0x7e}},
		{name: "Save memory 3", save: (*Jiecang).SaveMemory3, slot: 3, expected: []byte{0xf1, 0xf1, 0x25, 0x00, 0x25, 0x7e}},
		{
			name:           "Write error is wrapped",
			save:           func(j *Jiecang) error { return j.SaveMemory(2) },
			writeErr:       errBLE,
			expectedErrMsg: "failed to save memory2",
			expected:       commands["save_memory2"],
		},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		fd.heightMM = 1020
		j := newTestDesk(t, fd)
		fd.setWriteErr(test.writeErr)

		err := test.save(j)
		if test.expectedErrMsg != "" {
			assert.ErrorContains(t, err, test.expectedErrMsg, test.name)
			if test.writeErr != nil {
				assert.ErrorIs(t, err, test.writeErr, test.name)
			}
		} else {
			assert.NoError(t, err, test.name)
		}

		if test.expected == nil {
			assert.Empty(t, fd.Writes(), test.name)
		} else {
			assert.Equal(t, [][]byte{test.expected}, fd.Writes(), test.name)
		}
		defaults := newFakeDesk().presetsMM
		for slot := 1; slot <= 3; slot++ {
			want := defaults[slot]
			if slot == test.slot {
				want = 1020
			}
			assert.Equal(t, want, fd.Preset(slot), test.name)
		}
		// The controller never acknowledges a save [JV], so the cached
		// preset is unchanged until the next FetchHeight.
		assert.Equal(t, uint8(defaults[1]/10), preset(j, "memory1"), test.name)
	}
}
