package cli

// Card t1708 (defect 2) — a backend body that states `Verdict: pass` while the
// adopted/recorded verdict is fail (finding-bullet signals outweigh the stated
// pass, P-CONS) must surface as an explicit recording defect. What Verdict
// MEANS does not change: the verdict, disagreement_flag, OverallVerdict and the
// receipt are exactly what they were.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

const statedPassBulletBody = "Verdict: pass\n\ndelta findings: []\n\nPre-existing (not delta):\n- [P3] old debt at legacy.go:9"

func TestSynthesizeReviewOutput_StatedPassOutweighedByBulletsIsARecordingDefect(t *testing.T) {
	out := synthesizeReviewOutput(statedPassBulletBody, codexMethodTurnStart)

	if out.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want \"fail\" — what Verdict means is unchanged", out.Verdict)
	}
	if out.SynthesisNote == "" {
		t.Error("SynthesisNote must still record the divergence, exactly as before")
	}
	if strings.TrimSpace(out.RecordingDefect) == "" {
		t.Fatal("RecordingDefect is empty; a stated pass recorded as fail must surface as an explicit recording defect")
	}
}

func TestSynthesizeReviewOutput_NoRecordingDefectWithoutAStatedPass(t *testing.T) {
	cases := map[string]string{
		"plain fail body":          "- [P1] SQL injection at db.go:12",
		"stated fail with bullets": "Verdict: fail — merge blocked.\n\n- [P1] SQL injection at db.go:12",
		"clean pass":               "Verdict: pass — nothing to block on.",
		"no signal":                "I walked the diff and moved on.",
		"stated inconclusive":      "Verdict: inconclusive\n\n- [P2] cannot tell",
	}
	for name, body := range cases {
		if got := synthesizeReviewOutput(body, codexMethodTurnStart).RecordingDefect; got != "" {
			t.Errorf("%s: RecordingDefect = %q, want empty", name, got)
		}
	}
}

func TestRecordingDefect_StaysOffTheSingleBackendOutputSchema(t *testing.T) {
	out := synthesizeReviewOutput(statedPassBulletBody, codexMethodTurnStart)
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "recording_defect") {
		t.Errorf("ReviewOutput JSON must keep its schema (carrier is json:\"-\"); got %s", raw)
	}
}

func TestRunMultiAudit_RecordingDefectReachesPerBackendVerdictWithoutChangingTheVerdict(t *testing.T) {
	run := func(defect string) ConvergenceResult {
		prev := backendCall
		backendCall = func(_ context.Context, backend, _, _, _ string) ReviewOutput {
			out := claudeReview("pass")
			if backend == BackendCodex {
				out.Verdict = "fail"
				out.SynthesisNote = "codex signals diverged: stated verdict label=pass, severity-tagged finding bullet=fail; adopted fail"
				out.RecordingDefect = defect
			}
			return out
		}
		defer func() { backendCall = prev }()
		return runMultiAudit(context.Background(), claudeReview("pass"), "uncommittedChanges", "", MultiAuditConfig{
			Gates: config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateOff},
		}, nil)
	}

	const defect = "body stated Verdict: pass but severity-tagged finding bullets outweighed it"
	with, without := run(defect), run("")

	var codexJSON map[string]any
	for _, v := range with.PerBackendVerdicts {
		if v.Backend != BackendCodex {
			continue
		}
		b, _ := json.Marshal(v)
		_ = json.Unmarshal(b, &codexJSON)
		if v.Verdict != "fail" {
			t.Errorf("codex Verdict = %q, want fail (unchanged)", v.Verdict)
		}
	}
	if got, _ := codexJSON["recording_defect"].(string); got != defect {
		t.Errorf("codex per-backend entry recording_defect = %q, want %q", got, defect)
	}
	for _, v := range without.PerBackendVerdicts {
		b, _ := json.Marshal(v)
		if strings.Contains(string(b), "recording_defect") {
			t.Errorf("no defect ⇒ the member must be omitted; got %s", b)
		}
	}
	if with.OverallVerdict != without.OverallVerdict {
		t.Errorf("OverallVerdict %q vs %q: surfacing a recording defect must not change the verdict", with.OverallVerdict, without.OverallVerdict)
	}
	if (with.DisagreementFlag == nil) != (without.DisagreementFlag == nil) ||
		(with.DisagreementFlag != nil && *with.DisagreementFlag != *without.DisagreementFlag) {
		t.Errorf("disagreement_flag changed: %v vs %v", with.DisagreementFlag, without.DisagreementFlag)
	}
}
