// External test package: the equality pin needs both sides of the
// layering boundary — homestate's duplicated owner-label mapper and
// internal/factory's canonical one — in one binary.
package homestate_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Review round-9 P2-1 (card t1538): homestate duplicates the canonical
// owner-label mapper (the layering guard forbids importing factory), so the
// two implementations are pinned equal over the whole vocabulary the
// factory role system can produce — canonical forms, the retired leader
// spelling, the legacy worker/agent lane labels, and shapes that must pass
// through unchanged.
func TestOwnerLabelNormalizationMatchesFactory(t *testing.T) {
	corpus := []string{
		"leader", "leader-2", "leader-run-cli",
		"lead", "lead-3", "lead-run-x",
		"lane", "lane-1", "lane-12",
		"worker", "worker-1", "worker-9",
		"agent", "agent-3",
		"worker-0", "worker-x", "agent-",
		"", "operator", "factory",
	}
	for _, label := range corpus {
		want := factory.NormalizeOwnerLabel(label)
		if got := homestate.CanonicalApprovalOwnerLabel(label); got != want {
			t.Errorf("canonical(%q) = %q, want %q (factory mapper)", label, got, want)
		}
	}
}
