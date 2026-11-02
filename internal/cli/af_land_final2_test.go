package cli

// lane-19 37th review (non-blocking): the failure of a gate reaches the ledger and the operator's terminal with no
// control character in its reason; and the help says what the Windows hosts and a gate that never stops printing get.

import (
	"context"
	"strings"
	"testing"
)

func TestAFLand_ControlCharactersInAFailedGatesOutputAreNotKept(t *testing.T) {
	afSkipWithoutSh(t)
	db, _, _, st, _ := afQueuedCard(t)
	// escape, bell, carriage return, backspace, a C1 control and a delete, around text that must survive
	gate := `printf 'before\n\033[31mred\033[0m\r\a\b\tkept\177\n'; printf '\302\233end\n'; exit 1`
	outcome, reason := afLandJSON(t, "land", "t1", "--db", db, "--gate", gate)
	if outcome != "failed" {
		t.Fatalf("outcome %q reason %q, want failed", outcome, reason)
	}
	rows, _ := st.Queue(context.Background())
	ledgerReason := ""
	if len(rows) == 1 {
		ledgerReason = rows[0].FailReason
	}
	// D1: the output is not in the reason; the reason carries no control character either way
	for name, text := range map[string]string{"the reported reason": reason, "the ledger's fail reason": ledgerReason} {
		for _, r := range text {
			if (r < 0x20 && r != '\n' && r != '\t') || (r >= 0x7f && r <= 0x9f) {
				t.Errorf("%s keeps the control character %U: %.200q", name, r, text)
				break
			}
		}
	}
}

func TestAFLand_HelpSaysWhatWindowsAndAnEndlessGateGet(t *testing.T) {
	for _, c := range newAFCmd().Commands() {
		if c.Name() != "land" {
			continue
		}
		for _, want := range []string{
			"WSL (or Git Bash, without process-group kill)",
			"prints without end is not an error: it runs until --gate-timeout",
		} {
			if !strings.Contains(c.Long, want) {
				t.Errorf("the help of `af land` lacks %q:\n%s", want, c.Long)
			}
		}
		return
	}
	t.Fatal("no `af land` verb")
}
