/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"fmt"
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

// listDevicesCmd represents the listDevices command
var listDevicesCmd = &cobra.Command{
	Use:   "listDevices",
	Short: "List available Bluetooth desks",
	Run: func(cmd *cobra.Command, args []string) {

		adapter := bluetooth.DefaultAdapter

		err := adapter.Enable()
		if err != nil {
			panic("Failed to enable BLE adapter")
		}

		results = make(map[string]bluetooth.ScanResult)
		err = adapter.Scan(onScan)
		if err != nil {
			panic("Failed to register scan callback")
		}
	},
}

func init() {
	rootCmd.AddCommand(listDevicesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listDevicesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listDevicesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	if device.HasServiceUUID(bluetooth.New16BitUUID(LierdaDeviceID)) {
		if _, scannedDevice := results[device.Address.String()]; scannedDevice {
			//Stop scanning if we scan the same device again.
			_ = adapter.StopScan()
			//Print results
			fmt.Printf("%-20s %-40s %-10s\n", "ADDRESS", "NAME", "RSSI")
			for device_address, device := range results {
				fmt.Printf("%-20s %-40s %-10d\n", device_address, device.LocalName(), device.RSSI)
			}
		} else {
			mu.Lock()
			results[device.Address.String()] = device
			fmt.Println(results)
			mu.Unlock()
		}
	}
}
