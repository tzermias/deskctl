/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var address string

var rootCmd = &cobra.Command{
	Use:   "deskctl",
	Short: "A CLI tool to control and manage Jiecang standing desks",
	Long: `Controls standing desks equipped with Jiecang controllers
Moves the desk up/down, manages memory presets`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.deskctl.yaml)")
	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "", "Device address")
}
