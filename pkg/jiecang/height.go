package jiecang

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Contains functions for height movement only.

// Moves the desk up
func (j *Jiecang) Up() error {
	return j.sendCommand(commands["up"])
}

// Moves the desk down
func (j *Jiecang) Down() error {
	return j.sendCommand(commands["down"])
}

// Moves the desk to the designated height. Height should be within the limits
// of the desk. The operation can be cancelled via the provided context.
func (j *Jiecang) GoToHeight(ctx context.Context, height uint8) error {
	//Ensure that height is within low and high limits of the desk.
	if height > j.HighestHeight || height < j.LowestHeight {
		return fmt.Errorf("height %d is out of range (low: %d, high: %d)", height, j.LowestHeight, j.HighestHeight)
	}
	data0 := byte((int(height) * 10) / 256)
	data1 := byte((int(height) * 10) % 256)
	command := []byte{
		0xf1,
		0xf1,
		0x1b,
		0x02,
		data0,
		data1,
		byte((int(0x1b) + int(0x02) + int(data0) + int(data1)) % 256),
		0x7e,
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		j.mu.RLock()
		currentHeight := j.currentHeight
		j.mu.RUnlock()

		if currentHeight == height {
			break
		}

		select {
		case <-ctx.Done():
			// Context cancelled, send stop command and return
			if err := j.sendCommand(commands["stop"]); err != nil {
				return fmt.Errorf("failed to send stop command: %w", err)
			}
			fmt.Printf("Operation cancelled at height %d cm\n", currentHeight)
			return nil
		case <-ticker.C:
			if err := j.sendCommand(command); err != nil {
				return fmt.Errorf("failed to send goto command: %w", err)
			}
		}
	}
	return nil
}

func (j *Jiecang) FetchHeight() error {
	//Implements fetch_height command

	// Returns
	/*
		f2f2 25 02 044e 79 7e //044e =1100 in decimal. Memory 1
		f2f2 25 02 044e 79 7e
		f2f2 26 02 030c 37 7e //030c = 780 in dec. Memory 2
		f2f2 26 02 030c 37 7e
		f2f2 27 02 0372 9e 7e //0372 = 882 in dec. Memory 3
		f2f2 27 02 0372 9e 7e
		f2f2 28 02 0000 2a 7e // Memory 4?
		f2f2 28 02 0000 2a 7e

	*/
	if err := j.sendCommand(commands["fetch_height"]); err != nil {
		return err
	}
	return j.sendCommand(commands["fetch_height"])
}

func (j *Jiecang) FetchHeightRange() error {
	// Implements fetch_height_range command

	//Retuns
	/*
			  f2f2 07 04 04f8 026c 75 7e
		            LEN HGH LOW  CSUM
	*/
	if err := j.sendCommand(commands["fetch_height_range"]); err != nil {
		return err
	}
	return j.sendCommand(commands["fetch_height_range"])
}

func readHeight(buf []byte) uint8 {
	if buf[3] == 0x03 {
		height := int(buf[4])*256 + int(buf[5])
		// Hack to round the value
		return uint8(math.Round(float64(height / 10.0)))
	}
	return 0
}

// Handles the response of FetchHeightRange command from the controller
func readHeightRange(buf []byte) (uint8, uint8) {
	if buf[3] == 0x04 {
		highestHeight := int(buf[4])*256 + int(buf[5])
		lowestHeight := int(buf[6])*256 + int(buf[7])
		return uint8(math.Round(float64(highestHeight / 10.0))),
			uint8(math.Round(float64(lowestHeight / 10.0)))
	}
	return 0, 0
}
