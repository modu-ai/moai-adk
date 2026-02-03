package cli

import (
	"strings"
	"testing"
)

func TestRankCmd_Exists(t *testing.T) {
	if rankCmd == nil {
		t.Fatal("rankCmd should not be nil")
	}
}

func TestRankCmd_Use(t *testing.T) {
	if rankCmd.Use != "rank" {
		t.Errorf("rankCmd.Use = %q, want %q", rankCmd.Use, "rank")
	}
}

func TestRankCmd_Short(t *testing.T) {
	if rankCmd.Short == "" {
		t.Error("rankCmd.Short should not be empty")
	}
}

func TestRankCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "rank" {
			found = true
			break
		}
	}
	if !found {
		t.Error("rank should be registered as a subcommand of root")
	}
}

func TestRankCmd_HasSubcommands(t *testing.T) {
	expected := []string{"login", "status", "logout", "sync", "exclude", "include", "register"}
	for _, name := range expected {
		found := false
		for _, cmd := range rankCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("rank should have %q subcommand", name)
		}
	}
}

func TestRankCmd_SubcommandCount(t *testing.T) {
	count := len(rankCmd.Commands())
	if count != 7 {
		t.Errorf("rank should have 7 subcommands, got %d", count)
	}
}

func TestRankCmd_SubcommandShortDescriptions(t *testing.T) {
	for _, cmd := range rankCmd.Commands() {
		if cmd.Short == "" {
			t.Errorf("rank subcommand %q should have a short description", cmd.Name())
		}
	}
}

func TestRankCmd_HelpOutput(t *testing.T) {
	// Verify help shows all subcommands
	var names []string
	for _, cmd := range rankCmd.Commands() {
		names = append(names, cmd.Name())
	}

	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "login") || !strings.Contains(joined, "logout") {
		t.Errorf("rank subcommands should include login and logout, got: %s", joined)
	}
}
