package jiecang

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// This file contains an in-process fake Jiecang desk controller. It plays
// the role of both BLE characteristics: commands written to it (DataIn) are
// logged and answered with raw F2F2 frames through the notification
// callback registered on it (DataOut), the way a real controller does.

// frame builds a valid controller response frame:
// F2 F2 cmd len data... checksum 7E.
func frame(cmd byte, data ...byte) []byte {
	return buildFrame(0xf2, cmd, data...)
}

// command builds a valid command frame as sent to the controller:
// F1 F1 cmd len data... checksum 7E.
func command(cmd byte, data ...byte) []byte {
	return buildFrame(0xf1, cmd, data...)
}

func buildFrame(preamble, cmd byte, data ...byte) []byte {
	sum := int(cmd) + len(data)
	buf := []byte{preamble, preamble, cmd, byte(len(data))}
	for _, b := range data {
		sum += int(b)
		buf = append(buf, b)
	}
	return append(buf, byte(sum%256), 0x7e)
}

// heightFrame builds a 0x01 height report for the given height in mm.
func heightFrame(mm int) []byte {
	return frame(0x01, byte(mm>>8), byte(mm), 0x07)
}

// fakeDesk is an in-process fake of a Jiecang controller. It implements
// both characteristic and disconnecter. All exported-by-convention state is
// guarded by mu; use the accessor methods from tests.
type fakeDesk struct {
	mu sync.Mutex

	heightMM    int    // Current height in mm
	lowMM       int    // Lowest height limit in mm
	highMM      int    // Highest height limit in mm
	presetsMM   [5]int // Memory presets 1-4 in mm (index 0 unused)
	memoryMode  byte   // 0x00 one-touch, 0x01 constant touch
	sensitivity byte   // Anti-collision sensitivity (1-3)
	stepMM      int    // Distance moved per movement command
	stopped     bool   // A valid stop command was received

	writes      [][]byte     // Every command written, valid or not
	notify      func([]byte) // Callback captured by EnableNotifications
	disconnects int          // Number of Disconnect calls

	// Errors returned by the corresponding methods, if set.
	writeErr      error
	readErr       error
	notifyErr     error
	disconnectErr error
	// failAt, if > 0, applies writeErr only from the failAt-th write on
	// (1-based), so later steps of a sequence can be made to fail.
	failAt int

	// batch sends all replies to one command in a single notification,
	// exercising the multi-frame split in characteristicReceiver.
	batch bool
	// async delivers replies from a goroutine, like a real BLE stack.
	async bool
	// settings makes the SETTINGS (0x07) reply also include the units,
	// memory mode, 0x17 and anti-collision frames sent over RJ-45 [JV].
	settings bool

	wg sync.WaitGroup // Tracks async deliveries
}

// newFakeDesk returns a desk at 80 cm with the 650-1300 mm range from the
// Jarvis protocol notes and presets at 75, 110 and 90 cm.
func newFakeDesk() *fakeDesk {
	return &fakeDesk{
		heightMM:    800,
		lowMM:       650,
		highMM:      1300,
		presetsMM:   [5]int{0, 750, 1100, 900, 0},
		sensitivity: 2,
		stepMM:      10,
	}
}

// newTestDesk wires fd in as device, DataIn and DataOut through newJiecang
// and sets a short poll interval so movement tests finish in milliseconds.
// The write log is cleared after initialisation.
func newTestDesk(t *testing.T, fd *fakeDesk) *Jiecang {
	t.Helper()
	if fd == nil {
		fd = newFakeDesk()
	}
	t.Cleanup(fd.wait)
	j, err := newJiecang(fd, fd, fd)
	require.NoError(t, err)
	fd.wait()
	j.pollInterval = time.Millisecond
	fd.resetWrites()
	return j
}

func (fd *fakeDesk) Disconnect() error {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	fd.disconnects++
	return fd.disconnectErr
}

func (fd *fakeDesk) EnableNotifications(cb func([]byte)) error {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	if fd.notifyErr != nil {
		return fd.notifyErr
	}
	fd.notify = cb
	return nil
}

func (fd *fakeDesk) Read(p []byte) (int, error) {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	if fd.readErr != nil {
		return 0, fd.readErr
	}
	return 0, nil
}

func (fd *fakeDesk) WriteWithoutResponse(p []byte) (int, error) {
	fd.mu.Lock()
	fd.writes = append(fd.writes, append([]byte(nil), p...))
	if fd.writeErr != nil && len(fd.writes) >= fd.failAt {
		err := fd.writeErr
		fd.mu.Unlock()
		return 0, err
	}
	replies := fd.handle(p)
	cb := fd.notify
	batch, async := fd.batch, fd.async
	fd.mu.Unlock()

	// Deliver without holding fd.mu: the callback takes j.mu.
	fd.deliver(cb, replies, batch, async)
	return len(p), nil
}

// handle updates desk state for a command and returns the reply frames.
// Must be called with fd.mu held.
func (fd *fakeDesk) handle(p []byte) [][]byte {
	if !validCommand(p) {
		// The real controller drops malformed frames.
		return nil
	}
	data := p[4 : len(p)-2]
	switch p[2] {
	case 0x07: // fetch_height, really SETTINGS [JV]
		replies := [][]byte{
			frame(0x25, byte(fd.presetsMM[1]>>8), byte(fd.presetsMM[1])),
			frame(0x26, byte(fd.presetsMM[2]>>8), byte(fd.presetsMM[2])),
			frame(0x27, byte(fd.presetsMM[3]>>8), byte(fd.presetsMM[3])),
			frame(0x28, byte(fd.presetsMM[4]>>8), byte(fd.presetsMM[4])),
			heightFrame(fd.heightMM),
		}
		if fd.settings {
			replies = append(replies,
				frame(0x0e, 0x00),
				frame(0x19, fd.memoryMode),
				frame(0x17, 0x00),
				frame(0x1d, fd.sensitivity),
			)
		}
		return replies
	case 0x0c: // fetch_height_range
		return [][]byte{frame(0x07,
			byte(fd.highMM>>8), byte(fd.highMM),
			byte(fd.lowMM>>8), byte(fd.lowMM))}
	case 0x01: // up
		fd.moveToward(fd.highMM)
		return [][]byte{heightFrame(fd.heightMM)}
	case 0x02: // down
		fd.moveToward(fd.lowMM)
		return [][]byte{heightFrame(fd.heightMM)}
	case 0x1b: // go to height
		if len(data) == 2 {
			fd.moveToward(int(data[0])*256 + int(data[1]))
		}
		return [][]byte{heightFrame(fd.heightMM), frame(0x1b)}
	case 0x05, 0x06, 0x27: // goto_memory1-3
		fd.moveToward(fd.presetsMM[memorySlot(p[2])])
		return [][]byte{heightFrame(fd.heightMM)}
	case 0x03, 0x04, 0x25: // save_memory1-3, never acknowledged [JV]
		fd.presetsMM[memorySlot(p[2])] = fd.heightMM
		return nil
	case 0x2b: // stop
		fd.stopped = true
		return nil
	default: // 0xa2, 0xaa and anything else: no reply
		return nil
	}
}

// moveToward moves the desk one step toward target. Must hold fd.mu.
func (fd *fakeDesk) moveToward(target int) {
	fd.stopped = false
	switch {
	case fd.heightMM < target:
		fd.heightMM = min(fd.heightMM+fd.stepMM, target)
	case fd.heightMM > target:
		fd.heightMM = max(fd.heightMM-fd.stepMM, target)
	}
}

func memorySlot(cmd byte) int {
	switch cmd {
	case 0x03, 0x05:
		return 1
	case 0x04, 0x06:
		return 2
	default:
		return 3
	}
}

// validCommand checks an F1F1 command frame the way the controller does.
func validCommand(p []byte) bool {
	if len(p) < 6 || p[0] != 0xf1 || p[1] != 0xf1 || p[len(p)-1] != 0x7e {
		return false
	}
	if int(p[3])+6 != len(p) {
		return false
	}
	sum := 0
	for _, b := range p[2 : len(p)-2] {
		sum += int(b)
	}
	return byte(sum%256) == p[len(p)-2]
}

func (fd *fakeDesk) deliver(cb func([]byte), replies [][]byte, batch, async bool) {
	if cb == nil || len(replies) == 0 {
		return
	}
	send := func() {
		if batch {
			var buf []byte
			for _, r := range replies {
				buf = append(buf, r...)
			}
			cb(buf)
			return
		}
		for _, r := range replies {
			cb(r)
		}
	}
	if !async {
		send()
		return
	}
	fd.wg.Add(1)
	go func() {
		defer fd.wg.Done()
		send()
	}()
}

// wait blocks until all async deliveries have been made.
func (fd *fakeDesk) wait() { fd.wg.Wait() }

func (fd *fakeDesk) Writes() [][]byte {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	return append([][]byte(nil), fd.writes...)
}

func (fd *fakeDesk) resetWrites() {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	fd.writes = nil
}

func (fd *fakeDesk) Height() int {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	return fd.heightMM
}

func (fd *fakeDesk) Preset(n int) int {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	return fd.presetsMM[n]
}

func (fd *fakeDesk) Stopped() bool {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	return fd.stopped
}

func (fd *fakeDesk) setWriteErr(err error) {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	fd.writeErr = err
}

// currentHeight reads j.currentHeight under j.mu.
func currentHeight(j *Jiecang) uint8 {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.currentHeight
}

// preset reads j.presets[name] under j.mu.
func preset(j *Jiecang, name string) uint8 {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.presets[name]
}
