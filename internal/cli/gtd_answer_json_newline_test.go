package cli

// Card t885. The `--json` line is a one-line record a consumer reads per line,
// so the trailing newline is part of the contract, not formatting. The
// staticcheck S1038 repair replaced Fprintln(Sprintf(...)) with Fprintf, where
// the newline must be written explicitly — a repair that drops it leaves a
// byte-identical payload and a broken stream. Nothing pinned that before.

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestGTDAnswerJSONOutputEndsWithNewline(t *testing.T) {
	root := t.TempDir()
	withProjectRoot(t, root)

	var out bytes.Buffer
	cmd := newGTDAnswerCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--json", "t999", "probe answer"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("gtd answer --json: %v", err)
	}

	got := out.String()
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("json output = %q, want a trailing newline — a line-oriented consumer would lose the record", got)
	}
	if strings.Count(strings.TrimSuffix(got, "\n"), "\n") != 0 {
		t.Errorf("json output = %q, want exactly one line", got)
	}

	var rec struct {
		Card       string `json:"card"`
		AnswerFile string `json:"answerFile"`
	}
	if err := json.Unmarshal([]byte(got), &rec); err != nil {
		t.Fatalf("json output %q is not valid JSON: %v", got, err)
	}
	if rec.Card != "t999" {
		t.Errorf("card = %q, want %q", rec.Card, "t999")
	}
	if filepath.Base(rec.AnswerFile) == "" {
		t.Errorf("answerFile = %q, want a path", rec.AnswerFile)
	}
}
