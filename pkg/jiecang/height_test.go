package jiecang

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHeight(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		input          []byte // Input
		expectedHeight uint8  // Expected result of function
	}{
		{
			name:           "Valid message for height",
			input:          []byte{0xf2, 0xf2, 0x01, 0x03, 0x03, 0x37, 0x07, 0x45, 0x7e},
			expectedHeight: 82,
		},
		{
			name:           "Correct height with incorrect checksum",
			input:          []byte{0xf2, 0xf2, 0x01, 0x03, 0x03, 0x37, 0x07, 0x46, 0x7e},
			expectedHeight: 82,
		},
		{
			name:           "Height of 800 mm",
			input:          heightFrame(800),
			expectedHeight: 80,
		},
		{
			// TODO: height / 10.0 is integer division, so math.Round never
			// rounds. 827 mm should be 83 cm.
			name:           "Height is truncated, not rounded",
			input:          heightFrame(827),
			expectedHeight: 82,
		},
		{
			// TODO: when the desk is set to inches, P0/P1 are tenths of an
			// inch [JV] (280 = 28.0 in = 71 cm), but readHeight assumes mm.
			name:           "Height in tenths of an inch is read as mm",
			input:          frame(0x01, 0x01, 0x18, 0x0f),
			expectedHeight: 28,
		},
		{
			name:           "Wrong data length",
			input:          frame(0x01, 0x03, 0x20),
			expectedHeight: 0,
		},
	}

	for _, test := range tests {
		result := readHeight(test.input)
		assert.Equal(t, test.expectedHeight, result, test.name)
	}
}

func TestReadHeightRange(t *testing.T) {
	tests := []struct {
		name            string // Name of the testcase
		input           []byte // Input
		expectedHighest uint8  // Expected highest height in cm
		expectedLowest  uint8  // Expected lowest height in cm
	}{
		{
			name:            "Range from Jarvis example 1300/650 mm",
			input:           frame(0x07, 0x05, 0x14, 0x02, 0x8a),
			expectedHighest: 130,
			expectedLowest:  65,
		},
		{
			// TODO: integer division, 1278 mm should round to 128 cm and
			// 626 mm to 63 cm.
			name:            "Range is truncated, not rounded",
			input:           frame(0x07, 0x04, 0xfe, 0x02, 0x72),
			expectedHighest: 127,
			expectedLowest:  62,
		},
		{
			name:            "Wrong data length",
			input:           frame(0x07, 0x05, 0x14, 0x02),
			expectedHighest: 0,
			expectedLowest:  0,
		},
	}

	for _, test := range tests {
		highest, lowest := readHeightRange(test.input)
		assert.Equal(t, test.expectedHighest, highest, test.name)
		assert.Equal(t, test.expectedLowest, lowest, test.name)
	}
}

func TestUpDown(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		move           func(j *Jiecang) error
		expected       []byte // Exact command bytes
		expectedHeight uint8  // Height in cm after one move from 80 cm
	}{
		{name: "Up", move: (*Jiecang).Up, expected: []byte{0xf1, 0xf1, 0x01, 0x00, 0x01, 0x7e}, expectedHeight: 81},
		{name: "Down", move: (*Jiecang).Down, expected: []byte{0xf1, 0xf1, 0x02, 0x00, 0x02, 0x7e}, expectedHeight: 79},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		j := newTestDesk(t, fd)
		require.NoError(t, test.move(j), test.name)
		assert.Equal(t, [][]byte{test.expected}, fd.Writes(), test.name)
		assert.Equal(t, test.expectedHeight, currentHeight(j), test.name)
	}
}

func TestGoToHeight(t *testing.T) {
	errBLE := errors.New("ble failure")
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name           string // Name of the testcase
		setup          func(fd *fakeDesk, j *Jiecang)
		ctx            context.Context
		height         uint8
		expectedErr    error  // Expected wrapped error, nil for success
		expectedErrMsg string // Expected error substring
		expectedHeight uint8  // Expected desk height in cm afterwards
		expectedLast   []byte // Expected last command, nil for no writes
	}{
		{
			name:           "Above range",
			height:         131,
			expectedErrMsg: "out of range",
			expectedHeight: 80,
		},
		{
			name:           "Below range",
			height:         64,
			expectedErrMsg: "out of range",
			expectedHeight: 80,
		},
		{
			name:           "Already at target",
			height:         80,
			expectedHeight: 80,
		},
		{
			name:           "Reaches target going up",
			height:         85,
			expectedHeight: 85,
			// 850 mm = 0x0352, checksum 1b+02+03+52 = 0x72
			expectedLast: []byte{0xf1, 0xf1, 0x1b, 0x02, 0x03, 0x52, 0x72, 0x7e},
		},
		{
			name:           "Reaches target going down",
			height:         70,
			expectedHeight: 70,
			expectedLast:   command(0x1b, 0x02, 0xbc),
		},
		{
			name:           "Reaches target with batched async replies",
			setup:          func(fd *fakeDesk, j *Jiecang) { fd.batch, fd.async = true, true },
			height:         100,
			expectedHeight: 100,
			expectedLast:   command(0x1b, 0x03, 0xe8),
		},
		{
			// TODO: GoToHeight documents returning ctx.Err() on cancel but
			// returns nil.
			name: "Context cancelled sends stop",
			setup: func(fd *fakeDesk, j *Jiecang) {
				j.pollInterval = time.Hour
			},
			ctx:            cancelled,
			height:         100,
			expectedHeight: 80,
			expectedLast:   []byte{0xf1, 0xf1, 0x2b, 0x00, 0x2b, 0x7e},
		},
		{
			name:           "Write error is wrapped",
			setup:          func(fd *fakeDesk, j *Jiecang) { fd.setWriteErr(errBLE) },
			height:         100,
			expectedErr:    errBLE,
			expectedErrMsg: "failed to send goto command",
			expectedHeight: 80,
			expectedLast:   command(0x1b, 0x03, 0xe8),
		},
		{
			name: "Stop write error is wrapped",
			setup: func(fd *fakeDesk, j *Jiecang) {
				j.pollInterval = time.Hour
				fd.setWriteErr(errBLE)
			},
			ctx:            cancelled,
			height:         100,
			expectedErr:    errBLE,
			expectedErrMsg: "failed to send stop command",
			expectedHeight: 80,
			expectedLast:   commands["stop"],
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

		err := j.GoToHeight(ctx, test.height)
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
		if test.expectedLast == nil {
			assert.Empty(t, writes, test.name)
			continue
		}
		if assert.NotEmpty(t, writes, test.name) {
			assert.Equal(t, test.expectedLast, writes[len(writes)-1], test.name)
		}
	}
}

func TestGoToHeightStopReachesDesk(t *testing.T) {
	fd := newFakeDesk()
	j := newTestDesk(t, fd)
	j.pollInterval = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, j.GoToHeight(ctx, 100))
	assert.True(t, fd.Stopped(), "fake desk accepted the stop command")
}
