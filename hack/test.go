package main

import (
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
		j := jiecang.Init(adapter, device.Address)

		// Go to Memory1
		j.GoToMemory1()
		time.Sleep(5 * time.Second)
		// Go to Memory2
		j.GoToMemory2()
		time.Sleep(200 * time.Millisecond)

		log.Println("Disconnecting...")
		err := j.Disconnect()
		if err != nil {
			println("error when disconnecting:", err)
			return
		}
		log.Println("Disconnected...")
	}
}
