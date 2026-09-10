// Package cli — SPEC-AUDIT-PARTICIPANT-COUNT-001 acceptance criteria (M3).
//
// These tests pin the participant-count axis of the convergence engine: the
// derived summary of a result must distinguish "the participants agreed"
// from "there were not enough participants to disagree". disagreement_flag
// narrows to *bool (null = undetermined below 2 participants, REQ-APC-003)
// and every result carries participant_count (REQ-APC-001/002).
//
// The byte-level criteria (AC-APC-003/004) marshal through the generic map
// view on purpose: the state file and the audit_multi tool result are both
// produced by marshalling this struct verbatim, so the JSON member — present,
// null, not false — is the contract every non-Go consumer actually reads.
//
// @MX:SPEC: SPEC-AUDIT-PARTICIPANT-COUNT-001
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// marshalToMap marshals v and decodes the bytes back into a generic map —
// the byte-level view of the result (member presence + value) rather than
// the Go-level field view.
func marshalToMap(t *testing.T, v any) (map[string]any, []byte) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal back to map: %v", err)
	}
	return m, b
}

// AC-APC-003: below two participants with no observed divergence, the
// serialized disagreement_flag member is JSON null — present, null-valued,
// and never the boolean false. Absence must fail this criterion exactly as
// false does (an omitempty on the pointer would satisfy the in-process
// criterion and fail here — which is why this one is separate).
func TestConverge_BelowTwo_NoDivergence_JSONNull_AC_APC_003(t *testing.T) {
	cases := []struct {
		name     string
		verdicts []PerBackendVerdict
	}{
		{"claude only (spec §A.2 case 3)", []PerBackendVerdict{pbvReq(BackendClaude, "pass")}},
		{"claude pass + two inconclusive (spec §A.2 case 1)", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"),
			pbvAdv(BackendCodex, VerdictInconclusive),
			pbvAdv(BackendGLM, VerdictInconclusive),
		}},
		{"empty verdict slice", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := converge(tc.verdicts)
			m, b := marshalToMap(t, r)

			got, present := m["disagreement_flag"]
			if !present {
				t.Fatalf("disagreement_flag member ABSENT from marshalled result; presence is required (bytes: %s)", b)
			}
			if got != nil {
				t.Errorf("disagreement_flag = %v (%T), want JSON null — a sub-2 no-divergence result is undetermined, not false (bytes: %s)", got, got, b)
			}
			if strings.Contains(string(b), `"disagreement_flag":false`) {
				t.Errorf("raw bytes contain \"disagreement_flag\":false; forbidden below 2 participants (bytes: %s)", b)
			}
		})
	}
}

// AC-APC-004: the two measured inputs of the premise probe — genuine
// three-way agreement vs claude-only — stop sharing an identical derived
// summary: they differ in BOTH positions this SPEC adds or narrows
// (participant_count 3 vs 1, disagreement_flag false vs null).
func TestConverge_TwoMeasuredInputs_DerivedSummaryDiffers_AC_APC_004(t *testing.T) {
	agreement := converge([]PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, "pass"),
		pbvAdv(BackendGLM, "pass"),
	})
	claudeOnly := converge([]PerBackendVerdict{pbvReq(BackendClaude, "pass")})

	am, ab := marshalToMap(t, agreement)
	cm, cb := marshalToMap(t, claudeOnly)

	if am["participant_count"] != float64(3) {
		t.Errorf("three-way agreement participant_count = %v, want 3 (bytes: %s)", am["participant_count"], ab)
	}
	if cm["participant_count"] != float64(1) {
		t.Errorf("claude-only participant_count = %v, want 1 (bytes: %s)", cm["participant_count"], cb)
	}
	if am["disagreement_flag"] != false {
		t.Errorf("three-way agreement disagreement_flag = %v (%T), want boolean false — 3 participants were compared and none disagreed (bytes: %s)", am["disagreement_flag"], am["disagreement_flag"], ab)
	}
	if cm["disagreement_flag"] != nil {
		t.Errorf("claude-only disagreement_flag = %v (%T), want JSON null — one participant cannot ground a `false` (bytes: %s)", cm["disagreement_flag"], cm["disagreement_flag"], cb)
	}
	// The direct regression witness: the two derived summaries differ in
	// both new-field positions (spec §A.3 — at HEAD they are identical).
	if am["participant_count"] == cm["participant_count"] {
		t.Errorf("participant_count identical across the two inputs (%v); the derived summary must distinguish them", am["participant_count"])
	}
	if am["disagreement_flag"] == cm["disagreement_flag"] {
		t.Errorf("disagreement_flag identical across the two inputs (%v); the derived summary must distinguish them", am["disagreement_flag"])
	}
}

// AC-APC-001: participant_count equals, for every row of the acceptance
// table, the number of entries whose gate is not `off` and whose verdict is
// pass or fail. Rows g (gate-off carrying pass) and h (all inconclusive) are
// the two boundaries REQ-APC-002 settles — a table omitting them does not
// pin the definition. Row c is also the witness for the second mutant of
// acceptance.md §D (counting inconclusive entries as participants).
func TestConverge_ParticipantCount_Table_AC_APC_001(t *testing.T) {
	cases := []struct {
		name      string
		verdicts  []PerBackendVerdict
		wantCount int
	}{
		{"a: no entries", nil, 0},
		{"b: claude required=pass", []PerBackendVerdict{pbvReq(BackendClaude, "pass")}, 1},
		{"c: claude required=pass + codex/glm advisory inconclusive", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"),
			pbvAdv(BackendCodex, VerdictInconclusive),
			pbvAdv(BackendGLM, VerdictInconclusive),
		}, 1},
		{"d: codex advisory=pass + glm advisory=pass", []PerBackendVerdict{
			pbvAdv(BackendCodex, "pass"),
			pbvAdv(BackendGLM, "pass"),
		}, 2},
		{"e: claude required=pass + codex required=fail", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"),
			pbvReq(BackendCodex, "fail"),
		}, 2},
		{"f: claude/codex required pass + glm advisory pass", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"),
			pbvReq(BackendCodex, "pass"),
			pbvAdv(BackendGLM, "pass"),
		}, 3},
		{"g: claude required=pass + codex gate-OFF pass", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"),
			pbv(BackendCodex, config.AuditGateOff, "pass"),
		}, 1},
		{"h: claude required=inconclusive + codex required=inconclusive", []PerBackendVerdict{
			pbvReq(BackendClaude, VerdictInconclusive),
			pbvReq(BackendCodex, VerdictInconclusive),
		}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := converge(tc.verdicts)
			if r.ParticipantCount != tc.wantCount {
				t.Errorf("participant_count = %d, want %d (REQ-APC-002: gate != off AND verdict ∈ {pass, fail})", r.ParticipantCount, tc.wantCount)
			}
		})
	}
	// §B edge — empty verdict slice: the count is a visible 0, the flag is
	// null, and the vacuous-pass overall is unchanged.
	t.Run("empty slice: null flag + vacuous pass (acceptance §B)", func(t *testing.T) {
		r := converge(nil)
		if r.DisagreementFlag != nil {
			t.Errorf("disagreement_flag non-nil (points at %t); want nil", *r.DisagreementFlag)
		}
		if r.OverallVerdict != overallVerdictPass {
			t.Errorf("overall_verdict = %q, want pass (vacuous truth, unchanged)", r.OverallVerdict)
		}
	})
}

// AC-APC-002: below two participants with no observed divergence, the
// in-process flag is nil — and specifically NOT a non-nil pointer to false.
// The assertion must distinguish those two states explicitly: a test shaped
// `flag != nil && !*flag` passes for both and is non-compliant
// (acceptance.md AC-APC-002), so this asserts nil directly.
func TestConverge_BelowTwo_NoDivergence_FlagNilNotFalse_AC_APC_002(t *testing.T) {
	cases := []struct {
		name      string
		verdicts  []PerBackendVerdict
		wantCount int
	}{
		{"claude only (spec §A.2 case 3)", []PerBackendVerdict{pbvReq(BackendClaude, "pass")}, 1},
		{"claude pass + two inconclusive (spec §A.2 case 1)", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"),
			pbvAdv(BackendCodex, VerdictInconclusive),
			pbvAdv(BackendGLM, VerdictInconclusive),
		}, 1},
		{"empty verdict slice", nil, 0},
		{"all entries inconclusive (acceptance §B)", []PerBackendVerdict{
			pbvReq(BackendClaude, VerdictInconclusive),
			pbvReq(BackendCodex, VerdictInconclusive),
		}, 0},
		{"RequiredFailWithInconclusive (measured count 1 — the one existing case whose asserted false becomes null)", []PerBackendVerdict{
			pbvReq(BackendClaude, "fail"),
			pbvReq(BackendCodex, VerdictInconclusive),
		}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := converge(tc.verdicts)
			if r.ParticipantCount != tc.wantCount {
				t.Errorf("participant_count = %d, want %d (the sub-2 premise of this criterion)", r.ParticipantCount, tc.wantCount)
			}
			if r.DisagreementFlag != nil {
				t.Errorf("disagreement_flag non-nil (points at %t); want nil — below 2 participants false is a claim one participant cannot ground (REQ-APC-003)", *r.DisagreementFlag)
			}
		})
	}

	// The DQ-2 refusal path: zero participants by construction — the engine
	// refuses before any backend contributes a verdict (acceptance §B row).
	t.Run("DQ-2 refusal (zero participants)", func(t *testing.T) {
		rc := &recordingCaller{}
		orig := backendCall
		backendCall = rc.call
		t.Cleanup(func() { backendCall = orig })

		r := runMultiAudit(context.Background(), ReviewOutput{}, "uncommittedChanges", "concurrency", MultiAuditConfig{
			Gates: config.AuditGates{
				Claude: config.AuditGateRequired,
				Codex:  config.AuditGateRequired,
				GLM:    config.AuditGateAdvisory,
			},
		}, nil)
		if r.ParticipantCount != 0 {
			t.Errorf("participant_count = %d, want 0 (a refusal compares nobody)", r.ParticipantCount)
		}
		if r.DisagreementFlag != nil {
			t.Errorf("disagreement_flag non-nil (points at %t); want nil — a refusal must not read as \"the participants agreed\"", *r.DisagreementFlag)
		}
		if r.OverallVerdict != overallVerdictFail {
			t.Errorf("overall_verdict = %q, want fail (refusal, unchanged)", r.OverallVerdict)
		}
		if r.ResidualRiskNote == "" {
			t.Error("residual_risk_note empty; want the missing-anchor note preserved")
		}
	})
}

// AC-APC-005: at two or more participants, DisagreementFlag is non-nil and
// equals exactly what the pre-change three-pass derivation produced for the
// same input. The six enumerated cases carry their measured participant
// counts (.moai/reports/t284/participant-count-probe.log); the boundary at
// exactly 2 is inclusive (REQ-APC-004).
func TestConverge_TwoOrMore_BooleanUnchanged_AC_APC_005(t *testing.T) {
	cases := []struct {
		name      string
		verdicts  []PerBackendVerdict
		wantCount int
		wantFlag  bool
	}{
		{"AllRequiredPass (measured 3, false)", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"), pbvReq(BackendCodex, "pass"), pbvReq(BackendGLM, "pass"),
		}, 3, false},
		{"AllRequiredFail (measured 2, false)", []PerBackendVerdict{
			pbvReq(BackendClaude, "fail"), pbvReq(BackendCodex, "fail"),
		}, 2, false},
		{"NoRequiredBackends_VacuousPass (measured 2, false — the inclusive boundary)", []PerBackendVerdict{
			pbvAdv(BackendCodex, "pass"), pbvAdv(BackendGLM, "pass"),
		}, 2, false},
		{"RequiredSplit (measured 3, true)", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"), pbvReq(BackendCodex, "fail"), pbvReq(BackendGLM, "pass"),
		}, 3, true},
		{"AdvisoryOnlyConflict (measured 3, true)", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"), pbvReq(BackendCodex, "pass"), pbvAdv(BackendGLM, "fail"),
		}, 3, true},
		{"DisagreementAdvisoryNotBlock (measured 2, true)", []PerBackendVerdict{
			pbvReq(BackendClaude, "pass"), pbvAdv(BackendCodex, "fail"),
		}, 2, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := converge(tc.verdicts)
			if r.ParticipantCount != tc.wantCount {
				t.Errorf("participant_count = %d, want %d (the ≥2 premise of this criterion)", r.ParticipantCount, tc.wantCount)
			}
			if r.DisagreementFlag == nil {
				t.Fatalf("disagreement_flag = nil; want non-nil — %d participants were compared, the boolean is decidable", tc.wantCount)
			}
			if *r.DisagreementFlag != tc.wantFlag {
				t.Errorf("disagreement_flag = %t, want %t (the pre-change three-pass derivation, unchanged at ≥2 — REQ-APC-004)", *r.DisagreementFlag, tc.wantFlag)
			}
		})
	}
}

// AC-APC-006: the undetermined state gates nothing. overall_verdict and
// fail_open_backends are what the pre-change engine produced for the same
// input (spec §A.2 case 3 measurement: pass, none), and the multi-review
// Stop gate ALLOWs it — the only BLOCK path stays overall_verdict == fail.
func TestConverge_Undetermined_GatesNothing_AC_APC_006(t *testing.T) {
	r := converge([]PerBackendVerdict{pbvReq(BackendClaude, "pass")})
	if r.OverallVerdict != overallVerdictPass {
		t.Errorf("overall_verdict = %q, want pass (the undetermined state must not alter it)", r.OverallVerdict)
	}
	if len(r.FailOpenBackends) != 0 {
		t.Errorf("fail_open_backends = %v, want empty (unchanged)", r.FailOpenBackends)
	}

	// Full path: the real DQ-1 writer, then the gate with the toggle enabled
	// and a change detected — the *bool round-trips through the state file.
	withMultiChangeDetector(t, true)
	root := t.TempDir()
	const sess = "sess-apc6"
	if err := persistConvergenceResult(r, sess, root); err != nil {
		t.Fatalf("persistConvergenceResult: %v", err)
	}
	out, err := HandleMultiReviewGate(gateInput(false), true, root, sess)
	if err != nil {
		t.Fatalf("HandleMultiReviewGate: %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("undetermined result must ALLOW (C3: a low count gates nothing); got %+v", out)
	}

	// The gate's ONLY block path remains overall_verdict == fail: an
	// undetermined result that carries overall fail (single required FAIL)
	// still blocks through that same path — no category added, none removed.
	failResult := converge([]PerBackendVerdict{pbvReq(BackendClaude, "fail")})
	if failResult.OverallVerdict != overallVerdictFail {
		t.Fatalf("setup: overall_verdict = %q, want fail", failResult.OverallVerdict)
	}
	if err := persistConvergenceResult(failResult, sess+"-fail", root); err != nil {
		t.Fatalf("persistConvergenceResult (fail): %v", err)
	}
	outFail, err := HandleMultiReviewGate(gateInput(false), true, root, sess+"-fail")
	if err != nil {
		t.Fatalf("HandleMultiReviewGate (fail): %v", err)
	}
	if outFail == nil || outFail.Decision != hook.DecisionBlock {
		t.Errorf("overall==fail must still BLOCK through the one existing path; got %+v", outFail)
	}
}

// AC-APC-007: a state file written by the current binary — boolean
// disagreement_flag, no participant_count member — still decodes in
// loadConvergenceResult. A decode failure here would fail the Stop gate open
// silently for any session in flight across the upgrade.
func TestLoadConvergenceResult_OldStateFile_Decodes_AC_APC_007(t *testing.T) {
	// Hand-written in the OLD shape on purpose: marshalling the current
	// struct can no longer produce it.
	const oldFormat = `{
  "per_backend_verdicts": [
    {"backend": "claude", "gate": "required", "verdict": "fail", "summary": "", "findings": [], "next_steps": []},
    {"backend": "codex", "gate": "required", "verdict": "pass", "summary": "", "findings": [], "next_steps": []}
  ],
  "overall_verdict": "fail",
  "disagreement_flag": true,
  "residual_risk_note": "required-backend FAIL: claude",
  "fail_open_backends": []
}`
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "state", "audit-multi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir audit-multi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sess-old.json"), []byte(oldFormat), 0o644); err != nil {
		t.Fatalf("write old-format state: %v", err)
	}

	r, ok := loadConvergenceResult(root, "sess-old")
	if !ok {
		t.Fatal("loadConvergenceResult ok = false; an old state file must still decode — a decode failure fails the gate open silently")
	}
	if r.DisagreementFlag == nil {
		t.Fatal("disagreement_flag = nil; want a non-nil pointer holding the recorded boolean")
	}
	if !*r.DisagreementFlag {
		t.Errorf("disagreement_flag = %t, want true (the recorded boolean)", *r.DisagreementFlag)
	}
	if r.ParticipantCount != 0 {
		t.Errorf("participant_count = %d, want the zero value (the member is absent from the old file)", r.ParticipantCount)
	}
	if r.OverallVerdict != overallVerdictFail {
		t.Errorf("overall_verdict = %q, want fail", r.OverallVerdict)
	}

	// The gate's decision for that result is unchanged: overall fail BLOCKs
	// exactly as it did before the field narrowed.
	withMultiChangeDetector(t, true)
	out, err := HandleMultiReviewGate(gateInput(false), true, root, "sess-old")
	if err != nil {
		t.Fatalf("HandleMultiReviewGate: %v", err)
	}
	if out == nil || out.Decision != hook.DecisionBlock {
		t.Errorf("old state file must yield the same gate decision (BLOCK on overall fail); got %+v", out)
	}
}

// AC-APC-008: the carve-out. A single participant's intra-backend divergence
// keeps disagreement_flag a non-nil true — the one below-2 case where null is
// forbidden — while overall_verdict, fail_open_backends, and the gate
// decision are unchanged. Guards the landed REQ-CVS-003 behaviour alongside
// TestConverge_SurfacesSignalDivergence_WithoutBlocking.
func TestConverge_SingleParticipantDivergence_CarveOut_AC_APC_008(t *testing.T) {
	withNote := []PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, VerdictInconclusive),
	}
	withNote[1].SynthesisNote = "codex signals diverged: stated verdict label=pass, scored verdict line=inconclusive; adopted inconclusive"

	baseline := converge([]PerBackendVerdict{
		pbvReq(BackendClaude, "pass"),
		pbvReq(BackendCodex, VerdictInconclusive),
	})
	r := converge(withNote)

	if r.ParticipantCount != 1 {
		t.Errorf("participant_count = %d, want 1 (the inconclusive codex entry is not a participant)", r.ParticipantCount)
	}
	if r.DisagreementFlag == nil || !*r.DisagreementFlag {
		t.Fatalf("disagreement_flag = %s; want non-nil true — the carve-out keeps a directly-observed divergence, the one sub-2 case where null is forbidden", flagState(r.DisagreementFlag))
	}
	m, b := marshalToMap(t, r)
	if m["disagreement_flag"] != true {
		t.Errorf("serialized disagreement_flag = %v (%T), want boolean true (bytes: %s)", m["disagreement_flag"], m["disagreement_flag"], b)
	}

	// Unchanged from the pre-change engine's output for the same input.
	if r.OverallVerdict != baseline.OverallVerdict {
		t.Errorf("overall_verdict = %q, want %q (unchanged)", r.OverallVerdict, baseline.OverallVerdict)
	}
	if len(r.FailOpenBackends) != len(baseline.FailOpenBackends) {
		t.Errorf("fail_open_backends = %v, want %v (unchanged)", r.FailOpenBackends, baseline.FailOpenBackends)
	}

	// Gate decision unchanged: overall pass ⇒ ALLOW through the real
	// persist → read round trip.
	withMultiChangeDetector(t, true)
	root := t.TempDir()
	const sess = "sess-apc8"
	if err := persistConvergenceResult(r, sess, root); err != nil {
		t.Fatalf("persistConvergenceResult: %v", err)
	}
	out, err := HandleMultiReviewGate(gateInput(false), true, root, sess)
	if err != nil {
		t.Fatalf("HandleMultiReviewGate: %v", err)
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("carve-out result with overall pass must ALLOW (unchanged gate decision); got %+v", out)
	}
}

// flagState renders a *bool for test failure messages without dereferencing
// a nil pointer.
func flagState(p *bool) string {
	if p == nil {
		return "nil"
	}
	return "non-nil " + fmt.Sprintf("%t", *p)
}
