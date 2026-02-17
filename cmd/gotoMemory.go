/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tzermias/deskctl/pkg/jiecang"
	"tinygo.org/x/bluetooth"
)

var memoryNum int
var j *jiecang.Jiecang

var gotoMemoryCmd = &cobra.Command{
	Use:   "goto-memory [MEMORY]",
	Short: "Moves the desk to memory",
	Long:  `Moves the desk to the designated memory. [MEMORY] is between 1-3.`,
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error
		// Validate that argument is an integer between 1 and 3
		memoryNum, err = strconv.Atoi(args[0])
		if err != nil || memoryNum < 1 || memoryNum > 3 {
			fmt.Fprintf(os.Stderr, "Memory number is not within boundaries (1-3): %d\n", memoryNum)
			os.Exit(1)
		}

		// Validate MAC address
		var mac bluetooth.MAC
		mac, err = bluetooth.ParseMAC(address)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid MAC address [%s]: %v\n", address, err)
			os.Exit(1)
		}

		//Initialize device
		j, err = jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize device: %v\n", err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		// Add timeout for operation (60 seconds)
		opCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		var err error
		switch memoryNum {
		case 1:
			err = j.GoToMemory1(opCtx)
		case 2:
			err = j.GoToMemory2(opCtx)
		case 3:
			err = j.GoToMemory3(opCtx)
		default:
			// We should never reach this state as we validate this argument with PreRun hook.
			fmt.Fprintf(os.Stderr, "Memory %d is not within boundaries (1-3)\n", memoryNum)
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to go to memory %d: %v\n", memoryNum, err)
			os.Exit(1)
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if err := j.Disconnect(); err != nil {
			fmt.Fprintf(os.Stderr, "Error when disconnecting: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(gotoMemoryCmd)

}
