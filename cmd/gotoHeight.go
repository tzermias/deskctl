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

var height int

// gotoHeightCmd represents the gotoHeight command
var gotoHeightCmd = &cobra.Command{
	Use:   "goto-height [HEIGHT]",
	Short: "Sets height of desk to HEIGHT",
	Long: `Moves the desk up or down to reach height specified by HEIGHT.

	An error is thrown if HEIGHT exceeeds limits of the desk.`,
	Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error
		// Validate that argument is an integer
		height, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid height value [%s]: %v\n", args[0], err)
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

		if err := j.GoToHeight(opCtx, uint8(height)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to go to height: %v\n", err)
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
	rootCmd.AddCommand(gotoHeightCmd)

}
