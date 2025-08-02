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
		// Validate that arugment is an integer between 1 and 3
		height, err = strconv.Atoi(args[0])
		if err != nil {
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
		j.GoToHeight(uint8(height))
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
	rootCmd.AddCommand(gotoHeightCmd)

}
