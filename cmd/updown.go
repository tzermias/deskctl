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
		if err := j.Up(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to move desk up: %v\n", err)
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
		if err := j.Down(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to move desk down: %v\n", err)
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
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}
