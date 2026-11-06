package jiecang

import (
	"errors"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandChecksums(t *testing.T) {
	for name, cmd := range commands {
		if !assert.GreaterOrEqual(t, len(cmd), 6, name) {
			continue
		}
		assert.Equal(t, []byte{0xf1, 0xf1}, cmd[:2], name+": preamble")
		assert.Equal(t, byte(0x7e), cmd[len(cmd)-1], name+": terminator")
		dataLen := int(cmd[3])
		if !assert.Equal(t, dataLen+6, len(cmd), name+": length") {
			continue
		}
		sum := 0
		for _, b := range cmd[2 : 4+dataLen] {
			sum += int(b)
		}
		assert.Equal(t, byte(sum%256), cmd[len(cmd)-2], name+": checksum")
	}
}

// receiverState is a snapshot of the fields characteristicReceiver writes.
type receiverState struct {
	currentHeight            uint8
	highestHeight            uint8
	lowestHeight             uint8
	presets                  map[string]uint8
	memoryConstantTouchMode  bool
	antiCollisionSensitivity uint8
}

func snapshot(j *Jiecang) receiverState {
	j.mu.RLock()
	defer j.mu.RUnlock()
	presets := make(map[string]uint8, len(j.presets))
	for k, v := range j.presets {
		presets[k] = v
	}
	return receiverState{
		currentHeight:            j.currentHeight,
		highestHeight:            j.HighestHeight,
		lowestHeight:             j.LowestHeight,
		presets:                  presets,
		memoryConstantTouchMode:  j.MemoryConstantTouchMode,
		antiCollisionSensitivity: j.AntiCollisionSensitivity,
	}
}

func concat(frames ...[]byte) []byte {
	var buf []byte
	for _, f := range frames {
		buf = append(buf, f...)
	}
	return buf
}

func TestCharacteristicReceiver(t *testing.T) {
	empty := map[string]uint8{}
	tests := []struct {
		name     string   // Name of the testcase
		inputs   [][]byte // Notifications delivered in order
		expected receiverState
	}{
		{
			name:     "Height report",
			inputs:   [][]byte{frame(0x01, 0x03, 0x37, 0x07)},
			expected: receiverState{currentHeight: 82, presets: empty},
		},
		{
			name:     "Height range report",
			inputs:   [][]byte{frame(0x07, 0x05, 0x14, 0x02, 0x8a)},
			expected: receiverState{highestHeight: 130, lowestHeight: 65, presets: empty},
		},
		{
			name: "Memory presets 1-4",
			inputs: [][]byte{
				frame(0x25, 0x02, 0xee),
				frame(0x26, 0x04, 0x4c),
				frame(0x27, 0x03, 0x84),
				frame(0x28, 0x00, 0x00),
			},
			expected: receiverState{presets: map[string]uint8{
				"memory1": 75, "memory2": 110, "memory3": 90, "memory4": 0,
			}},
		},
		{
			name: "Units, 0x17 and goto-height ack change nothing",
			inputs: [][]byte{
				frame(0x0e, 0x00),
				frame(0x17, 0x00),
				frame(0x1b),
			},
			expected: receiverState{presets: empty},
		},
		{
			name:     "Memory mode constant touch",
			inputs:   [][]byte{frame(0x19, 0x01)},
			expected: receiverState{memoryConstantTouchMode: true, presets: empty},
		},
		{
			// TODO: 19 01 00 means one-touch [JV], but the receiver checks
			// msg[3] (the length byte, always 0x01) instead of msg[4], so any
			// memory mode report sets constant touch. It is also never
			// cleared. Expected: memoryConstantTouchMode false.
			name:     "Memory mode one-touch after constant touch",
			inputs:   [][]byte{frame(0x19, 0x01), frame(0x19, 0x00)},
			expected: receiverState{memoryConstantTouchMode: true, presets: empty},
		},
		{
			// TODO: the receiver stores msg[3] (the length byte) instead of
			// msg[4], so sensitivity always reads as 1. Expected: 3.
			name:     "Anti-collision sensitivity",
			inputs:   [][]byte{frame(0x1d, 0x03)},
			expected: receiverState{antiCollisionSensitivity: 1, presets: empty},
		},
		{
			name:     "Unknown valid frame is ignored",
			inputs:   [][]byte{frame(0x99, 0x01, 0x02)},
			expected: receiverState{presets: empty},
		},
		{
			name:     "Height with invalid checksum is ignored",
			inputs:   [][]byte{{0xf2, 0xf2, 0x01, 0x03, 0x03, 0x37, 0x07, 0x46, 0x7e}},
			expected: receiverState{presets: empty},
		},
		{
			name: "Short and garbage input is ignored",
			inputs: [][]byte{
				nil,
				{},
				{0x7e},
				{0xf2, 0x7e},
				{0xde, 0xad, 0xbe, 0xef, 0x7e},
				{0xf2, 0xf2, 0x01, 0x03, 0x03, 0x37, 0x07, 0x45}, // no terminator
				{0xf2, 0xf2, 0xff, 0xff, 0x00, 0x7e},             // length too large
			},
			expected: receiverState{presets: empty},
		},
		{
			name: "Multiple frames in one buffer",
			inputs: [][]byte{concat(
				frame(0x25, 0x02, 0xee),
				frame(0x07, 0x05, 0x14, 0x02, 0x8a),
				heightFrame(800),
			)},
			expected: receiverState{
				currentHeight: 80, highestHeight: 130, lowestHeight: 65,
				presets: map[string]uint8{"memory1": 75},
			},
		},
		{
			name: "Valid frame after garbage in one buffer",
			inputs: [][]byte{concat(
				[]byte{0x01, 0x02, 0x7e},
				heightFrame(900),
			)},
			expected: receiverState{currentHeight: 90, presets: empty},
		},
		{
			// TODO: the receiver splits on every 0x7e byte, so a frame whose
			// payload contains 0x7e is cut in two and dropped. Height 894 mm
			// is 03 7E. Expected: currentHeight 89.
			name:     "Height with 0x7e in payload",
			inputs:   [][]byte{heightFrame(894)},
			expected: receiverState{presets: empty},
		},
	}

	for _, test := range tests {
		j := &Jiecang{presets: make(map[string]uint8)}
		for _, in := range test.inputs {
			j.characteristicReceiver(in)
		}
		assert.Equal(t, test.expected, snapshot(j), test.name)
	}
}

func TestNewJiecang(t *testing.T) {
	errBLE := errors.New("ble failure")
	fetchWrites := [][]byte{
		commands["fetch_height"], commands["fetch_height"],
		commands["fetch_height_range"], commands["fetch_height_range"],
		commands["fetch_stand_time"], commands["fetch_stand_time"],
		commands["fetch_all_time"], commands["fetch_all_time"],
	}
	tests := []struct {
		name        string // Name of the testcase
		setup       func(fd *fakeDesk)
		expectedErr string // Expected error prefix, empty for success
	}{
		{name: "Success", setup: func(fd *fakeDesk) {}},
		{name: "Success with batched replies", setup: func(fd *fakeDesk) { fd.batch = true }},
		{name: "Success with async replies", setup: func(fd *fakeDesk) { fd.async = true }},
		{
			name:        "Enable notifications fails",
			setup:       func(fd *fakeDesk) { fd.notifyErr = errBLE },
			expectedErr: "failed to enable notifications",
		},
		{
			name:        "Initial read fails",
			setup:       func(fd *fakeDesk) { fd.readErr = errBLE },
			expectedErr: "failed to read initial height",
		},
		{
			name:        "Fetch height fails",
			setup:       func(fd *fakeDesk) { fd.writeErr = errBLE },
			expectedErr: "failed to fetch height",
		},
		{
			name:        "Fetch height range fails",
			setup:       func(fd *fakeDesk) { fd.writeErr, fd.failAt = errBLE, 3 },
			expectedErr: "failed to fetch height range",
		},
		{
			name:        "Fetch stand time fails",
			setup:       func(fd *fakeDesk) { fd.writeErr, fd.failAt = errBLE, 6 },
			expectedErr: "failed to fetch stand time",
		},
		{
			name:        "Fetch all time fails",
			setup:       func(fd *fakeDesk) { fd.writeErr, fd.failAt = errBLE, 7 },
			expectedErr: "failed to fetch all time",
		},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		test.setup(fd)
		j, err := newJiecang(fd, fd, fd)
		fd.wait()
		if test.expectedErr != "" {
			assert.ErrorIs(t, err, errBLE, test.name)
			assert.ErrorContains(t, err, test.expectedErr, test.name)
			assert.Nil(t, j, test.name)
			continue
		}
		require.NoError(t, err, test.name)
		assert.Equal(t, defaultPollInterval, j.pollInterval, test.name)
		assert.Equal(t, receiverState{
			currentHeight: 80, highestHeight: 130, lowestHeight: 65,
			presets: map[string]uint8{
				"memory1": 75, "memory2": 110, "memory3": 90, "memory4": 0,
			},
		}, snapshot(j), test.name)
		assert.Equal(t, fetchWrites, fd.Writes(), test.name)
	}
}

func TestNewJiecangSettings(t *testing.T) {
	fd := newFakeDesk()
	fd.settings = true
	fd.memoryMode = 0x00
	fd.sensitivity = 3
	j := newTestDesk(t, fd)
	// TODO: should be false and 3; see TestCharacteristicReceiver.
	assert.True(t, j.MemoryConstantTouchMode, "memory mode from RJ-45 settings")
	assert.Equal(t, uint8(1), j.AntiCollisionSensitivity, "sensitivity from RJ-45 settings")
}

func TestSendCommand(t *testing.T) {
	errBLE := errors.New("ble failure")
	tests := []struct {
		name     string // Name of the testcase
		writeErr error
		input    []byte
	}{
		{name: "Writes the exact bytes", input: commands["up"]},
		{name: "Write error is wrapped", input: commands["down"], writeErr: errBLE},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		j := newTestDesk(t, fd)
		fd.setWriteErr(test.writeErr)
		err := j.sendCommand(test.input)
		if test.writeErr != nil {
			assert.ErrorIs(t, err, test.writeErr, test.name)
			assert.ErrorContains(t, err, "failed to send command", test.name)
		} else {
			assert.NoError(t, err, test.name)
		}
		assert.Equal(t, [][]byte{test.input}, fd.Writes(), test.name)
	}
}

func TestFetchCommands(t *testing.T) {
	errBLE := errors.New("ble failure")
	tests := []struct {
		name     string // Name of the testcase
		fetch    func(j *Jiecang) error
		expected []byte
	}{
		{name: "FetchHeight", fetch: (*Jiecang).FetchHeight, expected: []byte{0xf1, 0xf1, 0x07, 0x00, 0x07, 0x7e}},
		{name: "FetchHeightRange", fetch: (*Jiecang).FetchHeightRange, expected: []byte{0xf1, 0xf1, 0x0c, 0x00, 0x0c, 0x7e}},
		{name: "FetchStandTime", fetch: (*Jiecang).FetchStandTime, expected: []byte{0xf1, 0xf1, 0xa2, 0x00, 0xa2, 0x7e}},
		{name: "FetchAllTime", fetch: (*Jiecang).FetchAllTime, expected: []byte{0xf1, 0xf1, 0xaa, 0x00, 0xaa, 0x7e}},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		j := newTestDesk(t, fd)
		assert.NoError(t, test.fetch(j), test.name)
		assert.Equal(t, [][]byte{test.expected, test.expected}, fd.Writes(), test.name)

		fd.resetWrites()
		fd.setWriteErr(errBLE)
		assert.ErrorIs(t, test.fetch(j), errBLE, test.name+": write error")
		assert.Len(t, fd.Writes(), 1, test.name+": stops after first failed write")
	}
}

func TestDisconnect(t *testing.T) {
	errBLE := errors.New("ble failure")
	tests := []struct {
		name          string // Name of the testcase
		disconnectErr error
	}{
		{name: "Disconnect succeeds"},
		{name: "Disconnect error is returned", disconnectErr: errBLE},
	}

	for _, test := range tests {
		fd := newFakeDesk()
		j := newTestDesk(t, fd)
		fd.disconnectErr = test.disconnectErr
		assert.Equal(t, test.disconnectErr, j.Disconnect(), test.name)
		assert.Equal(t, 1, fd.disconnects, test.name)
	}
}

func FuzzCharacteristicReceiver(f *testing.F) {
	f.Add(heightFrame(800))
	f.Add(frame(0x07, 0x05, 0x14, 0x02, 0x8a))
	f.Add(frame(0x25, 0x02, 0xee))
	f.Add(frame(0x19, 0x01))
	f.Add(frame(0x1d, 0x02))
	f.Add(concat(frame(0x0e, 0x00), frame(0x17, 0x00), frame(0x1b)))
	f.Add([]byte{0xf2, 0xf2, 0xff, 0xff, 0x00, 0x7e})
	f.Add([]byte{0x7e, 0x7e, 0x7e})

	log.SetOutput(io.Discard)
	f.Cleanup(func() { log.SetOutput(os.Stderr) })

	f.Fuzz(func(t *testing.T, data []byte) {
		j := &Jiecang{presets: make(map[string]uint8)}
		j.characteristicReceiver(data) // must not panic
	})
}
