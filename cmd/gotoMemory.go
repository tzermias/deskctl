/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tzermias/deskctl/pkg/jiecang"
	"tinygo.org/x/bluetooth"
)

var memory_num int
var j *jiecang.Jiecang

var gotoMemoryCmd = &cobra.Command{
	Use:   "goto-memory [MEMORY]",
	Short: "Moves the desk to memory",
	Long:  `Moves the desk to the designated memory. [MEMORY] is between 1-3.`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error
		// Validate that arugment is an integer between 1 and 3
		memory_num, err = strconv.Atoi(args[0])
		if err != nil || memory_num < 1 || memory_num > 3 {
			fmt.Println("Memory number is not within boundaries (1-3)")
			os.Exit(1)
		}

		// Validate MAC address
		var mac bluetooth.MAC
		mac, err = bluetooth.ParseMAC(address)
		if err != nil {
			fmt.Printf("Invalid MAC address [%s]\n", address)
			os.Exit(1)
		}

		//Initialize device
		j = jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}})

	},
	Run: func(cmd *cobra.Command, args []string) {
		switch memory_num {
		case 1:
			j.GoToMemory1()
		case 2:
			j.GoToMemory2()
		case 3:
			j.GoToMemory3()
		default:
			// We should never reach this state as we validate this argument with PreRun hook.
			fmt.Printf("Memory %d is not within boundaries (1-3)", memory_num)
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		err := j.Disconnect()
		if err != nil {
			fmt.Printf("Error when disconnecting: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(gotoMemoryCmd)

}
