/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tzermias/deskcli/pkg/jiecang"
	"tinygo.org/x/bluetooth"
)

var gotoMemoryCmd = &cobra.Command{
	Use:   "goto-memory",
	Short: "Go to memory (1-3)",
	Long:  `Moves the desk to the designated memory`,
	Run: func(cmd *cobra.Command, args []string) {
		adapter := bluetooth.DefaultAdapter

		err := adapter.Enable()
		if err != nil {
			fmt.Println("Could not enable Bluetooth adapter.", err)
			os.Exit(-1)
		}

		//Parse Bluetooth MAC address from argument
		mac, err := bluetooth.ParseMAC(address)
		if err != nil {
			fmt.Printf("Invalid MAC address [%s]", address)
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
			fmt.Printf("Memory %d is not a valid memory", memory_num)
		}

		err = j.Disconnect()
		if err != nil {
			println("error when disconnecting:", err)
			return
		}
	},
}

var memory_num int

func init() {
	rootCmd.AddCommand(gotoMemoryCmd)

	gotoMemoryCmd.Flags().IntVarP(&memory_num, "memory", "m", 1, "Memory address  of the desk (1-3)")
	gotoMemoryCmd.MarkFlagRequired("memory")
}
