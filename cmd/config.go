/*
Copyright © 2025 Aris Tzermias
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tzermias/deskctl/pkg/jiecang"
	"tinygo.org/x/bluetooth"
)

// configFileName is the name of the deskctl config file, located in the user's home directory.
const configFileName = ".deskctl.yaml"

// addAddressFlag registers the --address flag on commands that connect to a desk.
func addAddressFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "Device MAC address (default: \"address\" from ~/"+configFileName+")")
}

// loadConfigAddress returns the address field of the config file in home.
// A missing config file is not an error and returns an empty address.
func loadConfigAddress(home string) (string, error) {
	path := filepath.Join(home, configFileName)

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read config %s: %w", path, err)
	}
	return v.GetString("address"), nil
}

// resolveAddress returns the desk MAC address. The --address flag takes
// precedence over the address field of the config file in home.
func resolveAddress(cmd *cobra.Command, home string) (bluetooth.MAC, error) {
	var address string
	var err error
	if cmd.Flags().Changed("address") {
		address, err = cmd.Flags().GetString("address")
	} else {
		address, err = loadConfigAddress(home)
	}
	if err != nil {
		return bluetooth.MAC{}, err
	}

	if address == "" {
		return bluetooth.MAC{}, fmt.Errorf("no device address: use --address or set \"address\" in ~/%s", configFileName)
	}

	mac, err := bluetooth.ParseMAC(address)
	if err != nil {
		return bluetooth.MAC{}, fmt.Errorf("invalid MAC address [%s]: %w", address, err)
	}
	return mac, nil
}

// connectDesk resolves the desk address and connects to it.
func connectDesk(cmd *cobra.Command) (*jiecang.Jiecang, error) {
	// The home directory is only needed to read the config file
	home, err := os.UserHomeDir()
	if err != nil && !cmd.Flags().Changed("address") {
		return nil, fmt.Errorf("failed to find home directory: %w", err)
	}

	mac, err := resolveAddress(cmd, home)
	if err != nil {
		return nil, err
	}

	return jiecang.Init(adapter, bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}})
}
