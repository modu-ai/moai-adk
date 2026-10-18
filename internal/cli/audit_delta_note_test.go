package cli

// Card t1708 (sync-audit F2) — the delta residual_risk_note states only the
// scope that was actually measured. It used to say "claude/glm reviewed
// <sha>...HEAD" unconditionally, even with those gates off, a leg that returned
// before reviewing, or a Claude result that is the caller's in-session anchor.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

const noteSHA = "0123456789abcdef0123456789abcdef01234567"

func deltaNoteFor(t *testing.T, origin string, anchor ReviewOutput, gates config.AuditGates, backend func(string) ReviewOutput) string {
	t.Helper()
	prev := backendCall
	backendCall = func(_ context.Context, name, _, _, _ string) ReviewOutput { return backend(name) }
	t.Cleanup(func() { backendCall = prev })
	res := runMultiAudit(context.Background(), anchor, codexTargetBaseBranch, "", MultiAuditConfig{
		Gates: gates, OriginProvider: origin, DeltaBase: noteSHA,
	}, nil)
	return res.ResidualRiskNote
}

func inconclusiveBackend() ReviewOutput {
	return ReviewOutput{Verdict: VerdictInconclusive, Summary: "no key", Findings: []Finding{}, NextSteps: []string{}}
}

// This repository's own plan: claude off, glm off. Only codex ran.
func TestRunMultiAudit_DeltaNote_GatesOffDoNotClaimAReview(t *testing.T) {
	note := deltaNoteFor(t, "", ReviewOutput{}, config.AuditGates{Claude: config.AuditGateOff, Codex: config.AuditGateRequired, GLM: config.AuditGateOff},
		func(string) ReviewOutput { return claudeReview("pass") })

	for _, claim := range []string{"claude/glm reviewed", "claude reviewed", "glm reviewed"} {
		if strings.Contains(note, claim) {
			t.Errorf("note claims %q although those gates were off: %q", claim, note)
		}
	}
	if !strings.Contains(note, noteSHA) || !strings.Contains(note, "codex was scoped in prose only") {
		t.Errorf("note must keep the delta base and the codex prose-only statement: %q", note)
	}
}

// A leg that returned before reviewing (missing key) produced no verdict.
func TestRunMultiAudit_DeltaNote_AnInconclusiveLegDoesNotClaimAReview(t *testing.T) {
	note := deltaNoteFor(t, "", ReviewOutput{}, config.AuditGates{Claude: config.AuditGateOff, Codex: config.AuditGateRequired, GLM: config.AuditGateAdvisory},
		func(name string) ReviewOutput {
			if name == BackendGLM {
				return inconclusiveBackend()
			}
			return claudeReview("pass")
		})

	if strings.Contains(note, "glm reviewed") || strings.Contains(note, "claude/glm reviewed") {
		t.Errorf("note claims a glm review although glm returned inconclusive: %q", note)
	}
}

// Legs that did return a verdict are named, and only those.
func TestRunMultiAudit_DeltaNote_NamesTheLegsThatReviewed(t *testing.T) {
	both := deltaNoteFor(t, "", ReviewOutput{}, config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateAdvisory},
		func(string) ReviewOutput { return claudeReview("pass") })
	if !strings.Contains(both, "claude/glm reviewed "+noteSHA+"...HEAD") {
		t.Errorf("both legs ran; note = %q, want it to name claude/glm and the range", both)
	}

	onlyClaude := deltaNoteFor(t, "", ReviewOutput{}, config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateOff},
		func(string) ReviewOutput { return claudeReview("pass") })
	if !strings.Contains(onlyClaude, "claude reviewed "+noteSHA+"...HEAD") || strings.Contains(onlyClaude, "glm") {
		t.Errorf("only claude ran; note = %q, want claude named and glm absent", onlyClaude)
	}
}

// The caller's in-session Claude result carries a scope the server cannot know.
func TestRunMultiAudit_DeltaNote_AnInSessionAnchorIsMarkedScopeNotVerified(t *testing.T) {
	note := deltaNoteFor(t, BackendClaude, claudeReview("pass"), config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateOff},
		func(string) ReviewOutput { return claudeReview("pass") })

	if !strings.Contains(note, "scope not verified") {
		t.Errorf("an in-session anchor must be marked scope not verified: %q", note)
	}
	for _, claim := range []string{"claude/glm reviewed", "claude reviewed"} {
		if strings.Contains(note, claim) {
			t.Errorf("note claims %q for a Claude result the server did not produce: %q", claim, note)
		}
	}
}

// Card t1708 (sync-audit F5) — the codex clause of the delta note follows the
// verdicts too: it described a codex scope even when codex never ran.
func TestRunMultiAudit_DeltaNote_CodexOffDoesNotClaimACodexScope(t *testing.T) {
	note := deltaNoteFor(t, "", ReviewOutput{}, config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateOff, GLM: config.AuditGateOff},
		func(string) ReviewOutput { return claudeReview("pass") })

	for _, claim := range []string{"codex was scoped", "prose only", "not mechanically bounded"} {
		if strings.Contains(note, claim) {
			t.Errorf("note claims %q although the codex gate was off: %q", claim, note)
		}
	}
	if !strings.Contains(note, "codex did not run") {
		t.Errorf("note must say codex did not run on the delta: %q", note)
	}
	if !strings.Contains(note, noteSHA) || !strings.Contains(note, "no second-review record and no session state written") {
		t.Errorf("note must keep the delta base and the tail: %q", note)
	}
}

// A codex verdict in the slice keeps the prose-only statement.
func TestRunMultiAudit_DeltaNote_CodexVerdictKeepsTheProseOnlyStatement(t *testing.T) {
	note := deltaNoteFor(t, "", ReviewOutput{}, config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateOff},
		func(string) ReviewOutput { return claudeReview("pass") })

	if !strings.Contains(note, "codex was scoped in prose only (not mechanically bounded)") {
		t.Errorf("codex ran; note must keep the prose-only statement: %q", note)
	}
	if strings.Contains(note, "codex did not run") {
		t.Errorf("codex ran; note must not say it did not: %q", note)
	}
}
