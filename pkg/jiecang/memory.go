package jiecang

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// GoToMemory moves the desk to the specified memory preset (1-3).
// The operation can be cancelled via the provided context.
func (j *Jiecang) GoToMemory(ctx context.Context, memoryNum int) error {
	if memoryNum < 1 || memoryNum > 3 {
		return fmt.Errorf("invalid memory number %d (must be 1-3)", memoryNum)
	}

	commandKey := fmt.Sprintf("goto_memory%d", memoryNum)
	memoryKey := fmt.Sprintf("memory%d", memoryNum)

	// Send the command twice the first time
	if err := j.sendCommand(commands[commandKey]); err != nil {
		return fmt.Errorf("failed to send goto memory%d command: %w", memoryNum, err)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		j.mu.RLock()
		currentHeight := j.currentHeight
		targetHeight := j.presets[memoryKey]
		j.mu.RUnlock()

		if currentHeight == targetHeight {
			break
		}

		select {
		case <-ctx.Done():
			// Context cancelled, return
			fmt.Printf("Operation cancelled at height %d cm\n", currentHeight)
			return nil
		case <-ticker.C:
			if err := j.sendCommand(commands[commandKey]); err != nil {
				return fmt.Errorf("failed to send goto memory%d command: %w", memoryNum, err)
			}
		}
	}
	return nil
}

// GoToMemory1 moves the desk to memory preset 1.
// Deprecated: Use GoToMemory(ctx, 1) instead.
func (j *Jiecang) GoToMemory1(ctx context.Context) error {
	return j.GoToMemory(ctx, 1)
}

// GoToMemory2 moves the desk to memory preset 2.
// Deprecated: Use GoToMemory(ctx, 2) instead.
func (j *Jiecang) GoToMemory2(ctx context.Context) error {
	return j.GoToMemory(ctx, 2)
}

// GoToMemory3 moves the desk to memory preset 3.
// Deprecated: Use GoToMemory(ctx, 3) instead.
func (j *Jiecang) GoToMemory3(ctx context.Context) error {
	return j.GoToMemory(ctx, 3)
}

// SaveMemory saves the current height to the specified memory preset (1-3).
func (j *Jiecang) SaveMemory(memoryNum int) error {
	if memoryNum < 1 || memoryNum > 3 {
		return fmt.Errorf("invalid memory number %d (must be 1-3)", memoryNum)
	}

	commandKey := fmt.Sprintf("save_memory%d", memoryNum)
	if err := j.sendCommand(commands[commandKey]); err != nil {
		return fmt.Errorf("failed to save memory%d: %w", memoryNum, err)
	}

	log.Printf("Saved height %d cm to memory %d", j.currentHeight, memoryNum)
	time.Sleep(200 * time.Millisecond)
	return nil
}

// SaveMemory1 saves the current height to memory preset 1.
// Deprecated: Use SaveMemory(1) instead.
func (j *Jiecang) SaveMemory1() error {
	return j.SaveMemory(1)
}

// SaveMemory2 saves the current height to memory preset 2.
// Deprecated: Use SaveMemory(2) instead.
func (j *Jiecang) SaveMemory2() error {
	return j.SaveMemory(2)
}

// SaveMemory3 saves the current height to memory preset 3.
// Deprecated: Use SaveMemory(3) instead.
func (j *Jiecang) SaveMemory3() error {
	return j.SaveMemory(3)
}

// Reads the response of the controller containing height settings of
// memory presets.
func readMemoryPreset(buf []byte) uint8 {
	if buf[3] == 0x02 {
		preset := int(buf[4])*256 + int(buf[5])
		// Hack to round the value
		return uint8(math.Round(float64(preset / 10.0)))
	}
	return 0
}
