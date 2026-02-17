package jiecang

import (
	"context"
	"fmt"
	"math"
	"time"
)

// This file contains functions for controlling desk height.

// Up sends a command to move the desk upward by one increment.
// Equivalent to pressing the up button on the desk control panel once.
// Returns an error if the command transmission fails.
func (j *Jiecang) Up() error {
	return j.sendCommand(commands["up"])
}

// Down sends a command to move the desk downward by one increment.
// Equivalent to pressing the down button on the desk control panel once.
// Returns an error if the command transmission fails.
func (j *Jiecang) Down() error {
	return j.sendCommand(commands["down"])
}

// GoToHeight moves the desk to the specified height in centimeters.
//
// The function validates that the target height is within the desk's configured
// limits (LowestHeight and HighestHeight), then sends movement commands and polls
// the current height until the target is reached or the context is cancelled.
//
// Parameters:
//   - ctx: Context for timeout and cancellation. The operation can be interrupted
//     by cancelling the context (e.g., with Ctrl+C or timeout).
//   - height: Target height in centimeters. Must be between LowestHeight and
//     HighestHeight (typically 60-120cm).
//
// Returns an error if:
//   - The target height is out of range
//   - Command transmission fails
//   - The context is cancelled (returns ctx.Err())
//
// The function polls the height every 200ms and sends a stop command when
// the target is reached or the operation is cancelled.
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

// FetchHeight requests the desk's saved memory preset heights from the controller.
// The command is sent twice as required by the protocol for reliability.
//
// The response contains the height values for all memory presets (1-4).
// The values are processed asynchronously by the characteristicReceiver callback
// and stored in the presets map.
//
// Returns an error if the command transmission fails.
func (j *Jiecang) FetchHeight() error {
	if err := j.sendCommand(commands["fetch_height"]); err != nil {
		return err
	}
	return j.sendCommand(commands["fetch_height"])
}

// FetchHeightRange requests the desk's minimum and maximum height limits.
// The command is sent twice as required by the protocol for reliability.
//
// The response contains the highest and lowest height values that the desk
// can physically reach. The values are processed asynchronously by the
// characteristicReceiver callback and stored in HighestHeight and LowestHeight.
//
// Returns an error if the command transmission fails.
func (j *Jiecang) FetchHeightRange() error {
	if err := j.sendCommand(commands["fetch_height_range"]); err != nil {
		return err
	}
	return j.sendCommand(commands["fetch_height_range"])
}

// readHeight decodes the current height value from the controller response.
// The function extracts the height from bytes 4-5 of the response buffer and converts
// it from millimeters (protocol format) to centimeters (application format).
//
// Parameters:
//   - buf: Response buffer from the controller. Expected format:
//     [0xf2, 0xf2, type, 0x03, highByte, lowByte, ..., checksum, 0x7e]
//
// Returns the current height in centimeters, or 0 if the response type is invalid.
func readHeight(buf []byte) uint8 {
	if buf[3] == 0x03 {
		height := int(buf[4])*256 + int(buf[5])
		// Hack to round the value
		return uint8(math.Round(float64(height / 10.0)))
	}
	return 0
}

// readHeightRange decodes the height range limits from the controller response.
// This handles the response from the FetchHeightRange command, extracting both
// the maximum and minimum height limits that the desk can physically reach.
//
// Parameters:
//   - buf: Response buffer from the controller. Expected format:
//     [0xf2, 0xf2, type, 0x04, highestHighByte, highestLowByte,
//     lowestHighByte, lowestLowByte, ..., checksum, 0x7e]
//
// Returns:
//   - highestHeight: Maximum reachable height in centimeters
//   - lowestHeight: Minimum reachable height in centimeters
//   - (0, 0) if the response type is invalid
func readHeightRange(buf []byte) (uint8, uint8) {
	if buf[3] == 0x04 {
		highestHeight := int(buf[4])*256 + int(buf[5])
		lowestHeight := int(buf[6])*256 + int(buf[7])
		return uint8(math.Round(float64(highestHeight / 10.0))),
			uint8(math.Round(float64(lowestHeight / 10.0)))
	}
	return 0, 0
}
