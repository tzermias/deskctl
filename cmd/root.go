/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tinygo.org/x/bluetooth"
)

var address string

var adapter *bluetooth.Adapter

var rootCmd = &cobra.Command{
	Use:   "deskctl",
	Short: "A CLI tool to control and manage Jiecang standing desks",
	Long: `Controls standing desks equipped with Jiecang controllers
Moves the desk up/down, manages memory presets`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize bluetooth adapter
		adapter = bluetooth.DefaultAdapter
		err := adapter.Enable()
		if err != nil {
			fmt.Printf("Could not enable Bluetooth adapter: %v\n", err)
			return
		}

	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "Device address")
}
