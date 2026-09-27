package codexadapter

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestHostLacksStopBlockCap pins the exemption to exactly one (harness, event)
// pair among the decision-bearing events: (Codex, Stop). The expected pair is
// computed from the event set, not hand-listed beyond the one fact the
// predicate encodes.
func TestHostLacksStopBlockCap(t *testing.T) {
	t.Parallel()
	var got []string
	for _, h := range []Harness{HarnessClaude, HarnessCodex} {
		for _, ev := range DecisionBearingEvents() {
			if HostLacksStopBlockCap(h, ev) {
				got = append(got, string(h)+"/"+string(ev))
			}
		}
	}
	if len(got) != 1 || got[0] != string(HarnessCodex)+"/"+string(hook.EventStop) {
		t.Fatalf("exempt pairs = %v, want exactly [codex/Stop]", got)
	}
	// An observation event is never exempt: the predicate is not a second
	// classification of events.
	if HostLacksStopBlockCap(HarnessCodex, hook.EventPostToolUse) {
		t.Fatal("PostToolUse must not be reported as a Stop-cap exemption")
	}
}
