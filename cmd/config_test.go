package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAddress(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		flag           string // Value of --address, unset if empty
		config         string // Contents of ~/.deskctl.yaml, no file if empty
		expectedResult string // Expected MAC address
		expectedError  string // Expected substring of the error, no error if empty
	}{
		{name: "Address from flag only", flag: "AA:BB:CC:DD:EE:FF", expectedResult: "AA:BB:CC:DD:EE:FF"},
		{name: "Address from config file only", config: "address: 11:22:33:44:55:66\n", expectedResult: "11:22:33:44:55:66"},
		{name: "Flag overrides config file", flag: "AA:BB:CC:DD:EE:FF", config: "address: 11:22:33:44:55:66\n", expectedResult: "AA:BB:CC:DD:EE:FF"},
		{name: "Flag ignores malformed config file", flag: "AA:BB:CC:DD:EE:FF", config: "address: [unclosed\n", expectedResult: "AA:BB:CC:DD:EE:FF"},
		{name: "No flag and no config file", expectedError: "no device address"},
		{name: "Config file without address", config: "other: value\n", expectedError: "no device address"},
		{name: "Malformed config file", config: "address: [unclosed\n", expectedError: "failed to read config"},
		{name: "Invalid MAC in config file", config: "address: not-a-mac\n", expectedError: "invalid MAC address [not-a-mac]"},
		{name: "Invalid MAC in flag", flag: "not-a-mac", expectedError: "invalid MAC address [not-a-mac]"},
	}
	for _, test := range tests {
		home := t.TempDir()
		if test.config != "" {
			err := os.WriteFile(filepath.Join(home, configFileName), []byte(test.config), 0o600)
			require.NoError(t, err, test.name)
		}

		cmd := &cobra.Command{}
		addAddressFlag(cmd)
		if test.flag != "" {
			require.NoError(t, cmd.Flags().Set("address", test.flag), test.name)
		}

		result, err := resolveAddress(cmd, home)
		if test.expectedError != "" {
			assert.ErrorContains(t, err, test.expectedError, test.name)
			continue
		}
		assert.NoError(t, err, test.name)
		assert.Equal(t, test.expectedResult, result.String(), test.name)
	}
}

func TestResolveAddressIgnoresEnv(t *testing.T) {
	t.Setenv("DESKCTL_ADDRESS", "AA:BB:CC:DD:EE:FF")

	cmd := &cobra.Command{}
	addAddressFlag(cmd)
	_, err := resolveAddress(cmd, t.TempDir())
	assert.ErrorContains(t, err, "no device address", "DESKCTL_ADDRESS must not be used")
}

func TestAddressFlag(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		cmd            *cobra.Command
		expectedResult bool // Whether the command has the --address flag
	}{
		{name: "up has --address", cmd: upCmd, expectedResult: true},
		{name: "down has --address", cmd: downCmd, expectedResult: true},
		{name: "goto-height has --address", cmd: gotoHeightCmd, expectedResult: true},
		{name: "goto-memory has --address", cmd: gotoMemoryCmd, expectedResult: true},
		//		{name: "devices has no --address", cmd: listDevicesCmd, expectedResult: false},
		//		{name: "version has no --address", cmd: versionCmd, expectedResult: false},
	}
	for _, test := range tests {
		// Flags() only holds local flags, so also check inherited ones
		hasFlag := test.cmd.Flags().Lookup("address") != nil || test.cmd.InheritedFlags().Lookup("address") != nil
		assert.Equal(t, test.expectedResult, hasFlag, test.name)
	}
}

func TestAddressFlagBeforeSubcommand(t *testing.T) {
	tests := []struct {
		name           string // Name of the testcase
		args           []string
		expectedResult string // Expected value of --address on the desk command
	}{
		{name: "Flag after subcommand", args: []string{"up", "-a", "AA:BB:CC:DD:EE:FF"}, expectedResult: "AA:BB:CC:DD:EE:FF"},
		{name: "Flag before subcommand", args: []string{"-a", "AA:BB:CC:DD:EE:FF", "up"}, expectedResult: "AA:BB:CC:DD:EE:FF"},
		{name: "Long flag before subcommand", args: []string{"--address", "AA:BB:CC:DD:EE:FF", "goto-height", "100"}, expectedResult: "AA:BB:CC:DD:EE:FF"},
	}
	for _, test := range tests {
		// Find and ParseFlags are what rootCmd.Execute does before running hooks
		cmd, flags, err := rootCmd.Find(test.args)
		require.NoError(t, err, test.name)
		require.NoError(t, cmd.ParseFlags(flags), test.name)

		address, err := cmd.Flags().GetString("address")
		assert.NoError(t, err, test.name)
		assert.Equal(t, test.expectedResult, address, test.name)

		// Restore the shared command's flag for the next testcase
		flag := cmd.Flags().Lookup("address")
		require.NoError(t, flag.Value.Set(""), test.name)
		flag.Changed = false
	}
}
