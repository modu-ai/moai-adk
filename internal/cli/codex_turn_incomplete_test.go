// Package cli — a codex review turn yields a verdict only if it completed, and
// the codex leg of `audit_multi` has a deadline
// (SPEC-AUDIT-MODEL-CONVERGE-001 M3b, AC-ACV-020).
//
// The reader tests drive the REAL production path (runCodexReviewRPC over a
// scripted session, as the blank-review characterization suite does); the leg
// tests drive handleAuditMulti with the codex backend routed through
// performCodexAudit and a fake session that behaves like a hung codex process.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001
package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	incompleteFindingBody = "- [P1] unvalidated input reaches the shell"
	incompleteCleanBody   = "Verdict: pass\n\nThe change introduces no blocking issues."
	streamClosedCause     = "stream closed before the turn completed"
)

// withoutTurnCompleted drops the closing turn/completed line of a scripted
// session, so the stream ends while a review item is already held.
func withoutTurnCompleted(lines []string) []string {
	return lines[:len(lines)-1]
}

// requireIncompleteTurn asserts the shape every non-completed turn must have:
// an inconclusive review that names its cause, carries no finding derived from
// the partial text, and returns the cause as an error.
func requireIncompleteTurn(t *testing.T, label string, out ReviewOutput, err error, cause string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: error = nil, want the cause %q", label, cause)
	}
	if out.Verdict != VerdictInconclusive {
		t.Errorf("%s: verdict = %q, want %q (no verdict from a turn that did not complete)", label, out.Verdict, VerdictInconclusive)
	}
	if !strings.Contains(out.Summary, cause) {
		t.Errorf("%s: summary = %q, want it to name %q", label, out.Summary, cause)
	}
	if len(out.Findings) != 0 {
		t.Errorf("%s: findings = %d, want 0 (partial text must be discarded)", label, len(out.Findings))
	}
}

// TestCodexTurnReview_StreamClosedBeforeCompletion_Inconclusive (AC-ACV-020 (a)):
// a stream that closes after a review item and before turn/completed is no
// verdict, whether the partial body carries a finding or reads clean.
func TestCodexTurnReview_StreamClosedBeforeCompletion_Inconclusive(t *testing.T) {
	for name, body := range map[string]string{"finding body": incompleteFindingBody, "clean body": incompleteCleanBody} {
		t.Run(name, func(t *testing.T) {
			out, err := runCodexTurnWithLines(t, withoutTurnCompleted(codexSessionScript(body)))
			requireIncompleteTurn(t, name, out, err, streamClosedCause)
		})
	}
}

// cancelOnItemSession replays a scripted session and cancels the caller's
// context at the moment it hands out the line at cancelAt, which emulates a
// caller that gives up while a partial review is already held.
type cancelOnItemSession struct {
	lines    []string
	cancelAt int
	cancel   context.CancelFunc
	sent     []string
}

func (s *cancelOnItemSession) start(context.Context, string, []string) (codexConn, error) {
	return &cancelOnItemConn{s: s}, nil
}

type cancelOnItemConn struct {
	s   *cancelOnItemSession
	idx int
}

func (c *cancelOnItemConn) send(line string) error { c.s.sent = append(c.s.sent, line); return nil }
func (c *cancelOnItemConn) close() error           { return nil }
func (c *cancelOnItemConn) recv() (string, bool) {
	if c.idx >= len(c.s.lines) {
		return "", false
	}
	l := c.s.lines[c.idx]
	if c.idx == c.s.cancelAt {
		c.s.cancel()
	}
	c.idx++
	return l, true
}

// TestCodexTurnReview_ContextEnded_DiscardsPartial (AC-ACV-020 (b)): a context
// that has ended while partial review text is held is inconclusive and names the
// context end; the partial text yields no finding.
func TestCodexTurnReview_ContextEnded_DiscardsPartial(t *testing.T) {
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	t.Cleanup(func() { codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess })
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	// Line 3 is the item/completed carrying the finding; the turn/completed that
	// follows must never be read because the context is already over.
	codexSession = &cancelOnItemSession{lines: codexSessionScript(incompleteFindingBody), cancelAt: 3, cancel: cancel}

	out, err := runCodexReviewRPC(ctx, "/fake/codex", codexMethodReviewStart,
		map[string]any{"target": codexTargetUncommitted})
	requireIncompleteTurn(t, "caller cancellation", out, err, "ended by context")
}

// TestCodexTurnReview_CompletedTurnUnchanged (AC-ACV-020 (c), positive control):
// a turn that completes still yields exactly the verdict it yielded before.
func TestCodexTurnReview_CompletedTurnUnchanged(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		out, err := runCodexTurnWithLines(t, codexSessionScript(incompleteCleanBody))
		if err != nil || out.Verdict != "pass" {
			t.Errorf("verdict = %q, err = %v, want %q and no error", out.Verdict, err, "pass")
		}
	})
	t.Run("fail", func(t *testing.T) {
		out, err := runCodexTurnWithLines(t, codexSessionScript(incompleteFindingBody))
		if err != nil || out.Verdict != "fail" || len(out.Findings) == 0 {
			t.Errorf("verdict = %q, findings = %d, err = %v, want %q with a finding and no error", out.Verdict, len(out.Findings), err, "fail")
		}
	})
	t.Run("blank body keeps its own inconclusive", func(t *testing.T) {
		out, err := runCodexTurnWithLines(t, codexSessionScript("   "))
		if err == nil || err.Error() != codexBlankReviewSummary {
			t.Errorf("error = %v, want %q", err, codexBlankReviewSummary)
		}
		if out.Summary != blankReviewInconclusive().Summary {
			t.Errorf("summary = %q, want the blank-review summary %q", out.Summary, blankReviewInconclusive().Summary)
		}
	})
}

// ctxHungCodexSession replays the handshake and the turn ack, then blocks in recv
// until the context it was started with ends — which is what a codex process
// killed by exec.CommandContext looks like to the reader: its stdout closes. The
// safety valve only keeps a regression (a leg with no deadline) from hanging the
// suite; a correct leg never reaches it.
type ctxHungCodexSession struct {
	lines  []string
	safety time.Duration
	sent   []string
}

func (s *ctxHungCodexSession) start(ctx context.Context, _ string, _ []string) (codexConn, error) {
	return &ctxHungCodexConn{s: s, ctx: ctx}, nil
}

type ctxHungCodexConn struct {
	s   *ctxHungCodexSession
	ctx context.Context
	idx int
}

func (c *ctxHungCodexConn) send(line string) error { c.s.sent = append(c.s.sent, line); return nil }
func (c *ctxHungCodexConn) close() error           { return nil }
func (c *ctxHungCodexConn) recv() (string, bool) {
	if c.idx < len(c.s.lines) {
		l := c.s.lines[c.idx]
		c.idx++
		return l, true
	}
	select {
	case <-c.ctx.Done():
	case <-time.After(c.s.safety):
	}
	return "", false
}

// shortenCodexLegTimeout shortens the leg limit and restores it on cleanup.
func shortenCodexLegTimeout(t *testing.T, d time.Duration) {
	t.Helper()
	prev := config.DefaultCodexAuditLegTimeout
	config.DefaultCodexAuditLegTimeout = d
	t.Cleanup(func() { config.DefaultCodexAuditLegTimeout = prev })
}

// hungCodexAuditCall runs one audit_multi call over root with claude and glm
// answering `pass` at the backend seam and the codex leg running the REAL
// performCodexAudit over a hung session. It returns the decoded result and how
// long the call took.
func hungCodexAuditCall(t *testing.T, root string) (map[string]any, time.Duration) {
	t.Helper()
	t.Setenv(config.EnvMoaiLaunchProvider, "")

	prevRunner, prevLook, prevSess, prevCall := codexRunner, codexLookPath, codexSession, backendCall
	t.Cleanup(func() {
		codexRunner, codexLookPath, codexSession, backendCall = prevRunner, prevLook, prevSess, prevCall
	})
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexSession = &ctxHungCodexSession{lines: codexSessionScript("unused")[:3], safety: 3 * time.Second}
	backendCall = func(ctx context.Context, backend, target, focus, projectRoot string) ReviewOutput {
		if backend == BackendCodex {
			return performCodexAudit(ctx, target, focus, projectRoot)
		}
		return ReviewOutput{Verdict: "pass", Summary: backend + ":pass", Findings: []Finding{}, NextSteps: []string{}}
	}

	start := time.Now()
	res, err := callToolAuditMulti(t, nil, map[string]any{"project_root": root})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("handleAuditMulti returned a Go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %s", toolResultText(res))
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(toolResultText(res)), &m); err != nil {
		t.Fatalf("decode audit_multi result: %v\n%s", err, toolResultText(res))
	}
	return m, elapsed
}

// codexEntry returns the codex per_backend_verdicts entry of a decoded result.
func codexEntry(t *testing.T, m map[string]any) map[string]any {
	t.Helper()
	list, _ := m["per_backend_verdicts"].([]any)
	for _, e := range list {
		if entry, _ := e.(map[string]any); entry["backend"] == BackendCodex {
			return entry
		}
	}
	t.Fatalf("no codex entry in per_backend_verdicts: %v", m["per_backend_verdicts"])
	return nil
}

// TestAuditMulti_CodexLegDeadline_FailsClosedNamed (AC-ACV-020 (d)): with
// `model: multi` a hung codex ends at the leg's own deadline, the entry is
// inconclusive with a summary that names the timeout, and the codex gate is
// unmet by name.
func TestAuditMulti_CodexLegDeadline_FailsClosedNamed(t *testing.T) {
	shortenCodexLegTimeout(t, 40*time.Millisecond)
	root := newAuditMultiBaselineRoot(t, planWorkflowYAML(config.AuditModelMulti, nil))

	m, elapsed := hungCodexAuditCall(t, root)

	if elapsed >= time.Second {
		t.Errorf("audit_multi took %v, want well under 1s with the leg limit at 40ms", elapsed)
	}
	entry := codexEntry(t, m)
	if entry["verdict"] != VerdictInconclusive {
		t.Errorf("codex verdict = %v, want %q", entry["verdict"], VerdictInconclusive)
	}
	if summary, _ := entry["summary"].(string); !strings.Contains(summary, "timed out") {
		t.Errorf("codex summary = %q, want it to name a timeout", summary)
	}
	if m["overall_verdict"] != "fail" {
		t.Errorf("overall_verdict = %v, want fail (a required gate with no verdict)", m["overall_verdict"])
	}
	if m["gate_unmet"] != BackendCodex {
		t.Errorf("gate_unmet = %v, want %q", m["gate_unmet"], BackendCodex)
	}
	if note, _ := m["residual_risk_note"].(string); !strings.Contains(note, BackendCodex) {
		t.Errorf("residual_risk_note = %q, want it to name codex", note)
	}
}

// TestAuditMulti_CodexLegDeadline_UnconfiguredFailsOpen (AC-ACV-020 (e)): with
// no configuration the same hung codex ends at the same deadline and stays the
// fail-open result — codex inconclusive, overall not failed, no gate unmet.
func TestAuditMulti_CodexLegDeadline_UnconfiguredFailsOpen(t *testing.T) {
	shortenCodexLegTimeout(t, 40*time.Millisecond)
	root := newAuditMultiBaselineRoot(t, "")

	m, elapsed := hungCodexAuditCall(t, root)

	if elapsed >= time.Second {
		t.Errorf("audit_multi took %v, want well under 1s with the leg limit at 40ms", elapsed)
	}
	entry := codexEntry(t, m)
	if entry["verdict"] != VerdictInconclusive {
		t.Errorf("codex verdict = %v, want %q", entry["verdict"], VerdictInconclusive)
	}
	if m["overall_verdict"] != "pass" {
		t.Errorf("overall_verdict = %v, want pass (an unconfigured tree stays fail-open)", m["overall_verdict"])
	}
	if _, present := m["gate_unmet"]; present {
		t.Errorf("gate_unmet = %v, want absent", m["gate_unmet"])
	}
	if _, present := m["plan_source"]; present {
		t.Errorf("plan_source = %v, want absent (nothing came from configuration)", m["plan_source"])
	}
}

// TestAuditMulti_CodexLegDeadline_CallerCancelIsNotATimeout (design.md §D.10
// item 5, EC-6): a leg ended by the caller's own cancellation is still
// inconclusive but never reads as the leg's timeout. Which of the reader's two
// causes it names (the ended context or the closed stream) depends on which the
// reader sees first, so only the timeout wording is asserted absent.
func TestAuditMulti_CodexLegDeadline_CallerCancelIsNotATimeout(t *testing.T) {
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	t.Cleanup(func() { codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess })
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexSession = &ctxHungCodexSession{lines: codexSessionScript("unused")[:3], safety: 3 * time.Second}

	ctx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(40*time.Millisecond, cancel)
	out := performCodexAudit(ctx, codexTargetUncommitted, "", "")

	if out.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q", out.Verdict, VerdictInconclusive)
	}
	if strings.Contains(out.Summary, "timed out") {
		t.Errorf("summary = %q, a caller cancellation must not read as the leg's timeout", out.Summary)
	}
}
