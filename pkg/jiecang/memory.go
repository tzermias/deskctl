package jiecang

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// GoToMemory moves the desk to the specified memory preset (1-3).
// The operation polls the current height until the target preset height is reached
// or the context is cancelled.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - memoryNum: Memory preset number (1, 2, or 3)
//
// Returns an error if:
//   - memoryNum is not in the valid range (1-3)
//   - command transmission fails
//   - context is cancelled (operation stops gracefully, returns nil)
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
//	defer cancel()
//	if err := desk.GoToMemory(ctx, 1); err != nil {
//	    log.Fatal(err)
//	}
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
			fmt.Printf("\rHeight: %d cm\n", currentHeight)
			break
		}

		select {
		case <-ctx.Done():
			fmt.Printf("\nOperation cancelled at height %d cm\n", currentHeight)
			return nil
		case <-ticker.C:
			fmt.Printf("\rHeight: %d cm", currentHeight)
			if err := j.sendCommand(commands[commandKey]); err != nil {
				return fmt.Errorf("failed to send goto memory%d command: %w", memoryNum, err)
			}
		}
	}
	return nil
}

// GoToMemory1 moves the desk to the height saved in memory preset 1.
// The operation polls the current height until the target is reached or
// the context is cancelled.
//
// Deprecated: Use GoToMemory(ctx, 1) instead for a unified interface.
//
// Returns an error if command transmission fails or the context is cancelled.
func (j *Jiecang) GoToMemory1(ctx context.Context) error {
	return j.GoToMemory(ctx, 1)
}

// GoToMemory2 moves the desk to the height saved in memory preset 2.
// The operation polls the current height until the target is reached or
// the context is cancelled.
//
// Deprecated: Use GoToMemory(ctx, 2) instead for a unified interface.
//
// Returns an error if command transmission fails or the context is cancelled.
func (j *Jiecang) GoToMemory2(ctx context.Context) error {
	return j.GoToMemory(ctx, 2)
}

// GoToMemory3 moves the desk to the height saved in memory preset 3.
// The operation polls the current height until the target is reached or
// the context is cancelled.
//
// Deprecated: Use GoToMemory(ctx, 3) instead for a unified interface.
//
// Returns an error if command transmission fails or the context is cancelled.
func (j *Jiecang) GoToMemory3(ctx context.Context) error {
	return j.GoToMemory(ctx, 3)
}

// SaveMemory saves the current desk height to the specified memory preset (1-3).
// The current height is stored in the controller's non-volatile memory
// and can be recalled later using GoToMemory.
//
// Parameters:
//   - memoryNum: Memory preset number (1, 2, or 3)
//
// Returns an error if:
//   - memoryNum is not in the valid range (1-3)
//   - command transmission fails
//
// Example:
//
//	if err := desk.SaveMemory(1); err != nil {
//	    log.Fatal(err)
//	}
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

// SaveMemory1 saves the current desk height to memory preset 1.
// The current height is stored in the controller's non-volatile memory
// and can be recalled later using GoToMemory1 or GoToMemory(ctx, 1).
//
// Deprecated: Use SaveMemory(1) instead for a unified interface.
//
// Returns an error if the command transmission fails.
func (j *Jiecang) SaveMemory1() error {
	return j.SaveMemory(1)
}

// SaveMemory2 saves the current desk height to memory preset 2.
// The current height is stored in the controller's non-volatile memory
// and can be recalled later using GoToMemory2 or GoToMemory(ctx, 2).
//
// Deprecated: Use SaveMemory(2) instead for a unified interface.
//
// Returns an error if the command transmission fails.
func (j *Jiecang) SaveMemory2() error {
	return j.SaveMemory(2)
}

// SaveMemory3 saves the current desk height to memory preset 3.
// The current height is stored in the controller's non-volatile memory
// and can be recalled later using GoToMemory3 or GoToMemory(ctx, 3).
//
// Deprecated: Use SaveMemory(3) instead for a unified interface.
//
// Returns an error if the command transmission fails.
func (j *Jiecang) SaveMemory3() error {
	return j.SaveMemory(3)
}

// readMemoryPreset decodes a memory preset height value from the controller response.
// The function extracts the height from bytes 4-5 of the response buffer and converts
// it from millimeters (protocol format) to centimeters (application format).
//
// Parameters:
//   - buf: Response buffer from the controller. Expected format:
//     [0xf2, 0xf2, type, 0x02, highByte, lowByte, checksum, 0x7e]
//
// Returns the preset height in centimeters, or 0 if the response type is invalid.
func readMemoryPreset(buf []byte) uint8 {
	if buf[3] == 0x02 {
		preset := int(buf[4])*256 + int(buf[5])
		// Hack to round the value
		return uint8(math.Round(float64(preset / 10.0)))
	}
	return 0
}
