package cli

import (
	"strings"
	"testing"
)

// t1628 — with no live factory leader, `moai codex -l` refuses with the
// NO_ACTIVE_FACTORY sentinel. The error renderer shows it as "No_active_factory.",
// so the operator reads a lowercased code. The refusal must read as a sentence.
func TestCodexLaneNoActiveFactoryReadsAsSentence(t *testing.T) {
	primary, _ := t1628LaneFixture(t)
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t1628Enter(t, primary)
	t1628StubCodex(t)

	err := t1628RunLane(t)
	if err == nil {
		t.Fatal("codex -l with no live factory leader must be refused")
	}
	if strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") {
		t.Errorf("the refusal shows the raw sentinel (rendered as No_active_factory.): %v", err)
	}
}
