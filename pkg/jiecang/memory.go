package jiecang

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// GoToMemoryX functions
func (j *Jiecang) GoToMemory1(ctx context.Context) {
	// Send the command twice the first time
	j.sendCommand(commands["goto_memory1"])

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
			return
		case <-ticker.C:
			j.sendCommand(commands["goto_memory1"])
		}
	}
}

func (j *Jiecang) GoToMemory2(ctx context.Context) {
	// Send the command twice the first time
	j.sendCommand(commands["goto_memory2"])

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
			return
		case <-ticker.C:
			j.sendCommand(commands["goto_memory2"])
		}
	}
}

func (j *Jiecang) GoToMemory3(ctx context.Context) {
	// Send the command twice the first time
	j.sendCommand(commands["goto_memory3"])

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
			return
		case <-ticker.C:
			j.sendCommand(commands["goto_memory3"])
		}
	}
}

// Save memory commands

func (j *Jiecang) SaveMemory1() {
	//Save memory
	j.sendCommand(commands["save_memory1"])

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
}

func (j *Jiecang) SaveMemory2() {
	//Save memory
	j.sendCommand(commands["save_memory2"])

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
}

func (j *Jiecang) SaveMemory3() {
	//Save memory
	j.sendCommand(commands["save_memory3"])

	log.Printf("Height: %d cm", j.currentHeight)
	time.Sleep(200 * time.Millisecond)
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
