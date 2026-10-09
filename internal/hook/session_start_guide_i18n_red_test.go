package hook

// session_start_guide_i18n_red_test.go — RED-now pins for
// SPEC-SESSION-START-GUIDE-I18N-001 (card t1603). Each test names the AC it
// carries and is authored RED on the pre-implementation tree; the run phase
// (M1/M2) flips them green. The observed RED outputs are recorded verbatim in
// acceptance.md §D (RED-now cells, four-element form).
//
// D9 hardening (plan-audit iter 6): the assertions are positive as well as
// absence-only, so a mutant that DELETES the authority sentence or ships the
// flags without the design-surface marker fails these tests. Mutation runs
// over the three named variants are recorded in acceptance.md.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// factoryMatrixPath is the Status Transition Ownership Matrix pointer every
// locale's spawn authority sentence must carry verbatim (AC-002).
const factoryMatrixPath = ".claude/rules/moai/development/spec-frontmatter-schema.md"

// authorityBodyFragment is an English body fragment of the standing spawn
// authority sentence (D11-a): a variant that renames the prefix but keeps the
// English body still fails the ko/ja/zh absence check, so localization is
// verified against the body, not a renameable prefix alone.
const authorityBodyFragment = "you are the lane session and therefore the orchestrator for your card"

// designSurfaceMarker is the locale-independent design-surface literal the
// 2-mode guidance block must carry verbatim (D12-b): a bare "t1600" card
// number elsewhere in the notice does not satisfy it.
const designSurfaceMarker = "designed surface — t1600, not yet shipped"

// TestRedAgentChannelFollowsConversationLanguage — AC-001 (REQ-001): with
// conversation_language=ko, the agent-facing additionalContext channel of the
// SessionStart bootstrap notice must render the ko locale prose. RED on the
// pre-implementation tree: session_start.go:518 hardcodes langEnglish for the
// additionalContext copy (decision-index Q1 resolved — this channel follows
// conversation_language).
func TestRedAgentChannelFollowsConversationLanguage(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvFactoryRunID, "red001")
	t.Setenv(config.EnvFactoryLeadAddr, "/tmp/moai-socket-factory/red001")

	cfg := &config.Config{}
	cfg.Language.ConversationLanguage = "ko"
	h := NewSessionStartHandler(&mockConfigProvider{cfg: cfg})

	out, err := h.Handle(context.Background(), &HookInput{
		Source:     "startup",
		SessionID:  "red-session",
		ProjectDir: t.TempDir(),
		CWD:        t.TempDir(),
	})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	ac := ""
	if out.HookSpecificOutput != nil {
		ac = out.HookSpecificOutput.AdditionalContext
	}
	if !strings.Contains(ac, "팩토리 모드: run ") {
		t.Errorf("additionalContext (agent-facing channel) must render ko locale prose, got:\n%s", ac)
	}
}

// TestRedLaneJoinAuthorityLocalized — AC-002 (REQ-002): the standing spawn
// authority sentence riding the lane join notice must be localized, and EVERY
// locale's join must carry the matrix pointer path verbatim (the positive
// assertion that catches a mutant deleting the authority sentence). RED on
// the pre-implementation tree: laneSpawnAuthority (lane_spawn_authority.go:41)
// is an English-only constant, so ko/ja/zh joins carry the English prefix.
func TestRedLaneJoinAuthorityLocalized(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		join := factoryLaneNotice("lane-2", 5, lang)
		if !strings.Contains(join, factoryMatrixPath) {
			t.Errorf("%s lane join must carry the matrix pointer path %q verbatim", lang, factoryMatrixPath)
		}
		if lang == langEnglish {
			if !strings.Contains(join, "Standing spawn authority:") {
				t.Errorf("en lane join must carry the canonical authority sentence")
			}
			continue
		}
		// D11-a: the absence check covers BOTH the exact prefix and the
		// English body fragment — a renamed-prefix variant that keeps the
		// English body must still fail, so the ko/ja/zh joins are verified to
		// be genuinely localized, not re-prefixed.
		if strings.Contains(join, "Standing spawn authority:") || strings.Contains(join, authorityBodyFragment) {
			t.Errorf("%s lane join must carry the localized authority sentence, English prefix or body still present", lang)
		}
	}
}

// TestRedAutoModeGuidancePresent — AC-005 (REQ-004/REQ-005): the leader
// bootstrap notice must explain the /moai todo --auto-leader / --auto-lane
// two-mode surface WITH the design-surface marker naming card t1600 — the
// truthfulness guarantee that the flags are described as designed, not as
// live. RED on the pre-implementation tree: no 2-mode guidance exists in
// factoryLocales.
func TestRedAutoModeGuidancePresent(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvFactoryRunID, "red005")
	t.Setenv(config.EnvFactoryLeadAddr, "/tmp/moai-socket-factory/red005")

	notice := factoryBootstrapNotice("", "red-session", "en")
	for _, want := range []string{"--auto-leader", "--auto-lane"} {
		if !strings.Contains(notice, want) {
			t.Errorf("leader notice must explain the 2-mode auto surface, missing %q", want)
		}
	}
	// D11-b: the design-surface marker must co-occur with the flag guidance
	// in the SAME block — the notice's established block delimiter is "\n\n",
	// so a variant placing "t1600" in a separate paragraph must fail.
	var flagBlock string
	for _, block := range strings.Split(notice, "\n\n") {
		if strings.Contains(block, "--auto-leader") {
			flagBlock = block
			break
		}
	}
	if flagBlock == "" {
		t.Errorf("leader notice has no --auto-leader guidance block")
		return
	}
	if !strings.Contains(flagBlock, designSurfaceMarker) {
		t.Errorf("the --auto-leader guidance block must carry the design-surface literal %q in the same block, literal missing from that block", designSurfaceMarker)
	}
}
