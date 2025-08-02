/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tzermias/deskctl/pkg/jiecang"
	"tinygo.org/x/bluetooth"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Move the desk up one unit",
	Long: `Moves the desk up by one unit. 

	This command is equivalent of pressing the up button in your standing desk control once.
	`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// Validate MAC address
		var mac bluetooth.MAC
		mac, err := bluetooth.ParseMAC(address)
		if err != nil {
			fmt.Printf("Invalid MAC address [%s]\n", address)
			os.Exit(1)
		}

		//Initialize device
		j = jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}})

	},
	Run: func(cmd *cobra.Command, args []string) {
		j.Up()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		err := j.Disconnect()
		if err != nil {
			fmt.Printf("Error when disconnecting: %v\n", err)
			return
		}
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Move the desk down one unit",
	Long: `Moves the desk down by one unit. 

	This command is equivalent of pressing the down button in your standing desk control once.
	`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// Validate MAC address
		var mac bluetooth.MAC
		mac, err := bluetooth.ParseMAC(address)
		if err != nil {
			fmt.Printf("Invalid MAC address [%s]\n", address)
			os.Exit(1)
		}

		//Initialize device
		j = jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}})

	},
	Run: func(cmd *cobra.Command, args []string) {
		j.Down()
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
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}
