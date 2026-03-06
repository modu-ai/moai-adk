//go:build !race

package cli

import (
	"testing"
)

// TestSearchCmd_HasSubcommand verifies that the search command is registered on the root command.
func TestSearchCmd_HasSubcommand(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "search" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'search' subcommand is not registered on rootCmd")
	}
}

// TestSearchCmd_FlagParsing verifies that all expected flags are defined on the search command.
func TestSearchCmd_FlagParsing(t *testing.T) {
	// Locate the search command.
	var cmd = searchCmd

	tests := []struct {
		flagName string
	}{
		{"branch"},
		{"since"},
		{"until"},
		{"role"},
		{"limit"},
		{"index-session"},
		{"project-path"},
		{"git-branch"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			f := cmd.Flags().Lookup(tt.flagName)
			if f == nil {
				t.Errorf("flag '--%s' is not defined on the search command", tt.flagName)
			}
		})
	}
}
