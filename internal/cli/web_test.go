package cli

import (
	"os"
	"strings"
	"testing"
)

// TestWebCmd_FlagsRegistered verifies REQ-WC-001 / AC-WC-001:
// the `web` subcommand exposes --port (int, default 8080) and --no-open (bool).
func TestWebCmd_FlagsRegistered(t *testing.T) {
	cmd := newWebCmd()

	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("--port flag is not registered")
	}
	if portFlag.DefValue != "8080" {
		t.Errorf("--port default = %q, want 8080", portFlag.DefValue)
	}
	if portFlag.Value.Type() != "int" {
		t.Errorf("--port type = %q, want int", portFlag.Value.Type())
	}

	noOpenFlag := cmd.Flags().Lookup("no-open")
	if noOpenFlag == nil {
		t.Fatal("--no-open flag is not registered")
	}
	if noOpenFlag.DefValue != "false" {
		t.Errorf("--no-open default = %q, want false", noOpenFlag.DefValue)
	}
	if noOpenFlag.Value.Type() != "bool" {
		t.Errorf("--no-open type = %q, want bool", noOpenFlag.Value.Type())
	}
}

// TestWebCmd_HelpListsFlags verifies AC-WC-001: `moai web --help` exits 0 and the
// help text lists both flags.
func TestWebCmd_HelpListsFlags(t *testing.T) {
	cmd := newWebCmd()
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("web --help returned error: %v", err)
	}

	help := out.String()
	if !strings.Contains(help, "--port") {
		t.Errorf("help output does not mention --port:\n%s", help)
	}
	if !strings.Contains(help, "--no-open") {
		t.Errorf("help output does not mention --no-open:\n%s", help)
	}
}

// TestWebCmd_Use verifies the subcommand is named "web".
func TestWebCmd_Use(t *testing.T) {
	cmd := newWebCmd()
	if !strings.HasPrefix(cmd.Use, "web") {
		t.Errorf("web command Use = %q, want prefix \"web\"", cmd.Use)
	}
}

// TestWeb_NoAskUserQuestion is the C-HRA-008 / REQ-PGN-012 static guard:
// the CLI web subcommand source must NOT reference AskUserQuestion. The
// orchestrator owns user interaction; the CLI returns exit codes only.
// Mirrors internal/cli/worktree/new_test.go TestNew_NoAskUserQuestion.
func TestWeb_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("web.go")
	if err != nil {
		t.Fatalf("read web.go: %v", err)
	}
	if strings.Contains(string(src), "AskUserQuestion") {
		t.Error("internal/cli/web.go must NOT reference AskUserQuestion (orchestrator-only HARD)")
	}
	if strings.Contains(string(src), "mcp__askuser") {
		t.Error("internal/cli/web.go must NOT reference mcp__askuser (orchestrator-only HARD)")
	}
}
