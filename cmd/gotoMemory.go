/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tzermias/deskcli/pkg/jiecang"
	"tinygo.org/x/bluetooth"
)

// gotoMemoryCmd represents the gotoMemory command
var gotoMemoryCmd = &cobra.Command{
	Use:   "gotoMemory",
	Short: "Go to memory (1-3)",
	Long:  `Moves the desk to the designated memory`,
	Run: func(cmd *cobra.Command, args []string) {
		adapter := bluetooth.DefaultAdapter

		err := adapter.Enable()
		if err != nil {
			panic("Failed to enable BLE adapter")
		}

		//Parse Bluetooth MAC address from argument
		mac, err := bluetooth.ParseMAC(address)
		if err != nil {
			panic("Failed to parse MAC address")
		}
		j := jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}})
		switch memory_num {
		case 1:
			j.GoToMemory1()
		case 2:
			j.GoToMemory2()
		case 3:
			j.GoToMemory3()
		default:
			fmt.Println("Memory %d is not a valid memory", memory_num)
		}

		err = j.Disconnect()
		if err != nil {
			println("error when disconnecting:", err)
			return
		}
	},
}

var address string
var memory_num int

func init() {
	rootCmd.AddCommand(gotoMemoryCmd)

	gotoMemoryCmd.Flags().StringVarP(&address, "address", "a", "", "Device address")
	gotoMemoryCmd.MarkFlagRequired("address")
	gotoMemoryCmd.Flags().IntVarP(&memory_num, "memory", "m", 1, "Memory address  of the desk (1-3)")
	gotoMemoryCmd.MarkFlagRequired("memory")
}
