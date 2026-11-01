package cli

// N7 (audit 22a0afd247 onwards, residual classification): the reason of a failed landing is clean (no output, no format
// characters), but the failure log keeps the gate's output with U+202E (right-to-left override) and U+200B (zero-width space)
// in it. The log is the file the reason sends the reader to, and a terminal shows it reordered. The expectation: the log
// does not carry the characters that reorder or hide text.

import (
	"os"
	"strings"
	"testing"
)

func TestAFResidual_N7TheFailureLogCarriesNoBidiOverrideOrZeroWidthCharacter(t *testing.T) {
	afSkipWithoutSh(t)
	// U+202E is e2 80 ae and U+200B is e2 80 8b; the gate prints them between ordinary words
	gate := `printf 'before \342\200\256reversed\342\200\213hidden after\n'; printf 'the last line\n'; exit 1`
	outcome, reason, logPath := afD1Land(t, gate)
	if outcome != "failed" || logPath == "" {
		t.Fatalf("setup: outcome %q, log %q (%.200s)", outcome, logPath, reason)
	}
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(raw)
	t.Logf("log: has U+202E=%v has U+200B=%v (%d bytes)", strings.ContainsRune(log, '‮'), strings.ContainsRune(log, '​'), len(log))
	if !strings.Contains(log, "the last line") {
		t.Fatalf("setup: the log lacks the gate's output: %.200q", log)
	}
	for name, r := range map[string]rune{"U+202E right-to-left override": '‮', "U+200B zero-width space": '​'} {
		if strings.ContainsRune(log, r) {
			t.Errorf("the failure log carries %s", name)
		}
	}
}
