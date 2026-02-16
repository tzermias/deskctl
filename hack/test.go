package main

import (
	"context"
	"log"
	"time"

	"github.com/tzermias/deskctl/pkg/jiecang"
	"tinygo.org/x/bluetooth"
)

const (
	LierdaDeviceID = 0xFE60
)

func main() {
	adapter := bluetooth.DefaultAdapter

	err := adapter.Enable()
	if err != nil {
		panic("Failed to enable BLE adapter")
	}

	err = adapter.Scan(onScan)
	log.Println("Scanning for Lierda devices ")
	if err != nil {
		panic("Failed to register scan callback")
	}
}

func onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	if device.HasServiceUUID(bluetooth.New16BitUUID(LierdaDeviceID)) {
		log.Println("Found Lierda device:", device.Address.String(), device.RSSI, device.LocalName())
		j, err := jiecang.Init(adapter, device.Address)
		if err != nil {
			log.Printf("Failed to initialize device: %v\n", err)
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Go to Memory1
		if err := j.GoToMemory1(ctx); err != nil {
			log.Printf("Failed to go to memory1: %v\n", err)
			return
		}
		time.Sleep(5 * time.Second)
		// Go to Memory2
		if err := j.GoToMemory2(ctx); err != nil {
			log.Printf("Failed to go to memory2: %v\n", err)
			return
		}
		time.Sleep(200 * time.Millisecond)

		log.Println("Disconnecting...")
		if err := j.Disconnect(); err != nil {
			log.Printf("Error when disconnecting: %v\n", err)
			return
		}
		log.Println("Disconnected...")
	}
}
