/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"tinygo.org/x/bluetooth"
)

const (
	LierdaDeviceID = 0xFE60
)

var (
	results map[string]bluetooth.ScanResult
	mu      sync.RWMutex
)

var listDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List available Bluetooth standing desks",
	Run: func(cmd *cobra.Command, args []string) {

		adapter := bluetooth.DefaultAdapter

		err := adapter.Enable()
		if err != nil {
			fmt.Println("Could not enable Bluetooth adapter.", err)
			os.Exit(-1)
		}

		results = make(map[string]bluetooth.ScanResult)
		err = adapter.Scan(onScan)
		if err != nil {
			fmt.Println("Could not scan available devices.", err)
			os.Exit(-1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listDevicesCmd)
}

func onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	if device.HasServiceUUID(bluetooth.New16BitUUID(LierdaDeviceID)) {
		if _, scannedDevice := results[device.Address.String()]; scannedDevice {
			//Stop scanning if we scan the same device again.
			_ = adapter.StopScan()

			//Print results
			fmt.Printf("%-20s %-20s %-10s\n", "ADDRESS", "NAME", "RSSI")
			for device_address, device := range results {
				fmt.Printf("%-20s %-20s %-10d\n", device_address, device.LocalName(), device.RSSI)
			}
		} else {
			mu.Lock()
			results[device.Address.String()] = device
			mu.Unlock()
		}
	}
}
