// Package curator_test — L2 shadow-apply canary tests (SPEC-HARNESS-EVOLVE-003 M4).
// REQ-HEV3-012/013/014/034: shadow-apply, regression gate, veto->A7, L2 reachability.
package curator_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/curator"
)

// makeBullets builds n bullets of approximately perBulletChars each.
func makeBullets(n int, perBulletChars int) []curator.Bullet {
	bullets := make([]curator.Bullet, n)
	for i := range n {
		bullets[i] = curator.Bullet{
			LedgerKey: "key-" + string(rune('A'+i%26)),
			Text:      strings.Repeat("x", perBulletChars),
		}
	}
	return bullets
}

// TestShadowApplyCanary_WithinBudget verifies a valid Tier-4 block passes the
// shadow-apply regression gate (REQ-HEV3-012).
func TestShadowApplyCanary_WithinBudget(t *testing.T) {
	t.Parallel()

	content := curator.BlockContent{
		Tier:    4,
		Bullets: makeBullets(3, 100),
	}
	result := curator.ShadowApplyCanary(curator.BlockTypeLearnedWorkflow, content)
	if result.Rejected {
		t.Errorf("Rejected = true, want false (within budget). Reason: %s", result.Reason)
	}
}

// TestShadowApplyCanary_ExceedsBudget verifies AC-HEV3-013: a Tier-4 block that
// exceeds the 3K digest budget is rejected by the shadow-apply regression gate.
func TestShadowApplyCanary_ExceedsBudget(t *testing.T) {
	t.Parallel()

	content := curator.BlockContent{
		Tier:    4,
		Bullets: makeBullets(10, 400), // ~4000 chars > 3000 MaxDigestBlockChars
	}
	result := curator.ShadowApplyCanary(curator.BlockTypeLearnedWorkflow, content)
	if !result.Rejected {
		t.Fatal("Rejected = false, want true (exceeds 3K budget — AC-HEV3-013)")
	}
}

// TestShadowApplyCanary_ExceedsBulletCap verifies the 20-bullet cap signal.
func TestShadowApplyCanary_ExceedsBulletCap(t *testing.T) {
	t.Parallel()

	content := curator.BlockContent{
		Tier:    4,
		Bullets: makeBullets(25, 10), // 25 > 20 cap
	}
	result := curator.ShadowApplyCanary(curator.BlockTypeLearnedWorkflow, content)
	if !result.Rejected {
		t.Fatal("Rejected = false, want true (exceeds 20-bullet cap)")
	}
}

// TestShadowApplyCanary_ForbiddenContent verifies anti-fabrication is a held-out signal.
func TestShadowApplyCanary_ForbiddenContent(t *testing.T) {
	t.Parallel()

	content := curator.BlockContent{
		Tier: 4,
		Bullets: []curator.Bullet{
			{LedgerKey: "key-1", Text: "Clean bullet about workflow patterns."},
			{LedgerKey: "key-2", Text: "SPEC-HARNESS-EVOLVE-003 leaked into the digest"},
		},
	}
	result := curator.ShadowApplyCanary(curator.BlockTypeLearnedWorkflow, content)
	if !result.Rejected {
		t.Fatal("Rejected = false, want true (forbidden SPEC-ID content)")
	}
}

// TestShadowApplyCanary_Tier3Local verifies the Tier-3 append surface passes
// (the budget/cap signals are Tier-4-specific; Tier-3 has no budget enforcement).
func TestShadowApplyCanary_Tier3Local(t *testing.T) {
	t.Parallel()

	content := curator.BlockContent{
		Tier:    3,
		Bullets: makeBullets(5, 500),
	}
	result := curator.ShadowApplyCanary(curator.BlockTypeLearnedLocal, content)
	if result.Rejected {
		t.Errorf("Rejected = true for Tier-3, want false (no budget gate). Reason: %s", result.Reason)
	}
}
