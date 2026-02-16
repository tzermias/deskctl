package jiecang

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// GoToMemoryX functions
func (j *Jiecang) GoToMemory1(ctx context.Context) error {
	// Send the command twice the first time
	if err := j.sendCommand(commands["goto_memory1"]); err != nil {
		return fmt.Errorf("failed to send goto memory1 command: %w", err)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		j.mu.RLock()
		currentHeight := j.currentHeight
		targetHeight := j.presets["memory1"]
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
			j.sendCommand(commands["goto_memory1"])
			if err := j.sendCommand(commands["goto_memory1"]); err != nil {
				return fmt.Errorf("failed to send goto memory1 command: %w", err)
			}
		}
	}
	return nil
}

func (j *Jiecang) GoToMemory2(ctx context.Context) error {
	// Send the command twice the first time
	if err := j.sendCommand(commands["goto_memory2"]); err != nil {
		return fmt.Errorf("failed to send goto memory2 command: %w", err)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		j.mu.RLock()
		currentHeight := j.currentHeight
		targetHeight := j.presets["memory2"]
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
			if err := j.sendCommand(commands["goto_memory2"]); err != nil {
				return fmt.Errorf("failed to send goto memory2 command: %w", err)
			}
		}
	}
	return nil
}

func (j *Jiecang) GoToMemory3(ctx context.Context) error {
	// Send the command twice the first time
	if err := j.sendCommand(commands["goto_memory3"]); err != nil {
		return fmt.Errorf("failed to send goto memory3 command: %w", err)
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		j.mu.RLock()
		currentHeight := j.currentHeight
		targetHeight := j.presets["memory3"]
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
			if err := j.sendCommand(commands["goto_memory3"]); err != nil {
				return fmt.Errorf("failed to send goto memory3 command: %w", err)
			}
		}
	}
	return nil
}

// Save memory commands

func (j *Jiecang) SaveMemory1() error {
	//Save memory
	if err := j.sendCommand(commands["save_memory1"]); err != nil {
		return fmt.Errorf("failed to save memory1: %w", err)
	}

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (j *Jiecang) SaveMemory2() error {
	//Save memory
	if err := j.sendCommand(commands["save_memory2"]); err != nil {
		return fmt.Errorf("failed to save memory2: %w", err)
	}

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (j *Jiecang) SaveMemory3() error {
	//Save memory
	if err := j.sendCommand(commands["save_memory3"]); err != nil {
		return fmt.Errorf("failed to save memory3: %w", err)
	}

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
	return nil
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
