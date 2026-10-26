package main

import (
	"context"
	"time"

	"github.com/tzermias/deskctl/pkg/jiecang"
	"github.com/tzermias/deskctl/pkg/logger"
	"tinygo.org/x/bluetooth"
)

const (
	LierdaDeviceID = 0xFE60
)

func main() {
	// Enable verbose logging for test script
	logger.SetVerbose(true)

	adapter := bluetooth.DefaultAdapter

	err := adapter.Enable()
	if err != nil {
		panic("Failed to enable BLE adapter")
	}

	err = adapter.Scan(onScan)
	logger.Println("Scanning for Lierda devices ")
	if err != nil {
		panic("Failed to register scan callback")
	}
}

func onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	if device.HasServiceUUID(bluetooth.New16BitUUID(LierdaDeviceID)) {
		logger.Println("Found Lierda device:", device.Address.String(), device.RSSI, device.LocalName())
		j, err := jiecang.Init(adapter, device.Address)
		if err != nil {
			logger.Printf("Failed to initialize device: %v\n", err)
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Go to Memory1
		if err := j.GoToMemory(ctx, 1); err != nil {
			logger.Printf("Failed to go to memory1: %v\n", err)
			return
		}
		time.Sleep(5 * time.Second)
		// Go to Memory2
		if err := j.GoToMemory(ctx, 2); err != nil {
			logger.Printf("Failed to go to memory2: %v\n", err)
			return
		}
		time.Sleep(200 * time.Millisecond)

		logger.Println("Disconnecting...")
		if err := j.Disconnect(); err != nil {
			logger.Printf("Error when disconnecting: %v\n", err)
			return
		}
		logger.Println("Disconnected...")
	}
}
