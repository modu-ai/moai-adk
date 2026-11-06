package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card t1822: the statusline writes to a pipe (Claude Code captures stdout),
// so the command must force color rather than detect a TTY. A bytes.Buffer is
// the same non-TTY condition.
func runStatuslineToBuffer(t *testing.T) string {
	t.Helper()
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = nil

	buf := new(bytes.Buffer)
	StatuslineCmd.SetOut(buf)
	StatuslineCmd.SetErr(buf)
	t.Cleanup(func() {
		StatuslineCmd.SetOut(nil)
		StatuslineCmd.SetErr(nil)
	})
	if err := StatuslineCmd.RunE(StatuslineCmd, []string{}); err != nil {
		t.Fatalf("statusline should not error, got: %v", err)
	}
	return strings.TrimRight(buf.String(), "\n")
}

func TestStatuslineCmd_ColorForcedOnPipe(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv(config.EnvNoColor, "")
	t.Setenv("COLORTERM", "")

	got := runStatuslineToBuffer(t)
	if !strings.Contains(got, "\x1b[38;5;") {
		t.Fatalf("non-TTY statusline must carry 256-color foreground codes, got %q", got)
	}
	for i, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "\x1b[22m") {
			t.Errorf("line %d must start with SGR 22, got %q", i, line)
		}
	}
}

func TestStatuslineCmd_NoColorStaysPlain(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("COLORTERM", "truecolor")

	got := runStatuslineToBuffer(t)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("NO_COLOR statusline must carry no escape codes, got %q", got)
	}
}
