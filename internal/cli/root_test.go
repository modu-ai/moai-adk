package cli

import (
	"bytes"
	"strings"
	"testing"
)

// --- DDD PRESERVE: Characterization tests for root command behavior ---

func TestRootCmd_Exists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
}

func TestRootCmd_Use(t *testing.T) {
	if rootCmd.Use != "moai" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "moai")
	}
}

func TestRootCmd_Short(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
	if !strings.Contains(rootCmd.Short, "MoAI-ADK") {
		t.Errorf("rootCmd.Short should contain 'MoAI-ADK', got %q", rootCmd.Short)
	}
}

func TestRootCmd_Long(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestRootCmd_HasVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("rootCmd.Version should not be empty")
	}
}

func TestRootCmd_HelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("root --help error: %v", err)
	}

	output := buf.String()

	// M6-S5: tui-based help uses 4 Korean section groups (ScreenHelp design).
	// "런처" replaces the old cobra "Launch Commands:" group header.
	if !strings.Contains(output, "런처") {
		t.Error("root --help should show 런처 section (M6-S5 tui help)")
	}

	// Verify core subcommands are listed in help output
	requiredCommands := []string{"moai version", "moai init", "moai doctor", "moai status"}
	for _, cmd := range requiredCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("root --help should list %q subcommand", cmd)
		}
	}
}

func TestRootCmd_NoArgs_ShowsHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{})

	_ = rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "MoAI-ADK") {
		t.Error("root with no args should display help containing MoAI-ADK")
	}
}

func TestRootCmd_UnknownCommandError(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"nonexistent-command-xyz"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("unknown command should return error")
	}
}

func TestRootCmd_DroppedCommandsNotPresent(t *testing.T) {
	droppedCommands := []string{"language", "analyze", "switch"}
	for _, name := range droppedCommands {
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == name {
				t.Errorf("dropped command %q should not be registered", name)
			}
		}
	}
}

func TestRootCmd_SubcommandCount(t *testing.T) {
	// Ensure we have a reasonable number of subcommands
	count := len(rootCmd.Commands())
	if count < 4 {
		t.Errorf("rootCmd should have at least 4 subcommands, got %d", count)
	}
}

func TestRootCmd_NoRunE(t *testing.T) {
	if rootCmd.RunE != nil {
		t.Error("rootCmd should not have RunE (bare moai shows help)")
	}
}

// --- DDD CHARACTERIZATION: M6-S5 tui help behavior ---

// TestCharacterize_Help_FourSections verifies renderRootHelp emits all 4 ScreenHelp section titles.
func TestCharacterize_Help_FourSections(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()

	output := buf.String()

	sections := []string{"프로젝트", "런처", "자율 개발", "거버넌스"}
	for _, s := range sections {
		if !strings.Contains(output, s) {
			t.Errorf("root --help should contain section %q (ScreenHelp group)", s)
		}
	}
}

// TestCharacterize_Help_HelpBarPresent verifies HelpBar key hints appear in help output.
func TestCharacterize_Help_HelpBarPresent(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()

	output := buf.String()

	hints := []string{"↑↓", "스크롤", "enter", "선택"}
	for _, h := range hints {
		if !strings.Contains(output, h) {
			t.Errorf("root --help HelpBar should contain hint %q", h)
		}
	}
}

// TestCharacterize_Help_SubcommandHelpUnchanged verifies non-root subcommand help
// still uses cobra default (renderRootHelp is rootCmd-only).
func TestCharacterize_Help_SubcommandHelpUnchanged(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version", "--help"})
	_ = rootCmd.Execute()

	output := buf.String()

	// Cobra default help for subcommands includes "Usage:" header.
	if !strings.Contains(output, "Usage:") {
		t.Error("version --help should show cobra Usage: section (subcommand help unchanged)")
	}

	// Should NOT show ScreenHelp 4-group layout in subcommand help.
	if strings.Contains(output, "프로젝트") {
		t.Error("subcommand help should not show 프로젝트 group (tui help is rootCmd-only)")
	}
}

// TestCharacterize_Help_RootGroupsContent verifies each group lists at least one expected command.
func TestCharacterize_Help_RootGroupsContent(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()
	output := buf.String()

	expected := []string{
		"moai init",         // 프로젝트 group
		"moai cc",           // 런처 group
		"moai loop",         // 자율 개발 group
		"moai constitution", // 거버넌스 group
	}
	for _, cmd := range expected {
		if !strings.Contains(output, cmd) {
			t.Errorf("root --help should list %q in its group", cmd)
		}
	}
}

// TestCharacterize_Help_NoHexColor verifies no bare hex literals appear in the help output text.
func TestCharacterize_Help_NoHexColor(t *testing.T) {
	// Set NO_COLOR so ANSI escapes are stripped; we check output text has no hex patterns.
	t.Setenv("NO_COLOR", "1")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	_ = rootCmd.Execute()

	output := buf.String()

	// After NO_COLOR, the visible text should contain no #RRGGBB patterns.
	// (lipgloss suppresses ANSI; hex literals would appear as raw text if present.)
	if strings.Contains(output, "#") {
		// Only flag if the # is followed by 6 hex chars (a color literal leak).
		for line := range strings.SplitSeq(output, "\n") {
			if idx := strings.Index(line, "#"); idx >= 0 {
				suffix := line[idx:]
				if len(suffix) >= 7 {
					candidate := suffix[1:7]
					isHex := true
					for _, c := range candidate {
						if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
							isHex = false
							break
						}
					}
					if isHex {
						t.Errorf("help output contains bare hex color literal: %q", suffix[:7])
					}
				}
			}
		}
	}
}
