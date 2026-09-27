package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
)

// recordBackends swaps the backend seam for the record fixtures: claude
// pass, codex fail, glm inconclusive.
func recordBackends(t *testing.T) {
	t.Helper()
	saved := backendCall
	backendCall = func(ctx context.Context, backend, target, focus, projectRoot string) (out ReviewOutput) {
		switch backend {
		case "claude":
			return claudeReview("pass")
		case "codex":
			return ReviewOutput{Verdict: "fail", Summary: "stub: codex fail"}
		default: // glm
			return ReviewOutput{Verdict: "inconclusive", Summary: "stub: glm inconclusive"}
		}
	}
	t.Cleanup(func() { backendCall = saved })
}

// readSecondReviewLines reads the record file's lines.
func readSecondReviewLines(t *testing.T, path string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("decode line: %v", err)
		}
		out = append(out, m)
	}
	return out
}

// TestAC_CLOSURE_012 — audit_multi binds the fan-out to the card.
func TestAC_CLOSURE_012(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
	recordBackends(t)

	result := runMultiAudit(context.Background(), ReviewOutput{}, "baseBranch", "", MultiAuditConfig{
		ProjectRoot: f.CardDir,
		CardID:      "c1",
	}, nil)

	records := readSecondReviewLines(t, filepath.Join(f.EvidenceDir, closure.SecondReviewFile))
	if len(records) != 1 {
		t.Fatalf("records = %d, want exactly one", len(records))
	}
	r := records[0]
	expectStr := func(key, want string) {
		t.Helper()
		if got, _ := r[key].(string); got != want {
			t.Fatalf("%s = %v, want %q", key, r[key], want)
		}
	}
	expectStr("card", "c1")
	expectStr("contract_card", "c1")
	expectStr("spec_id", closuretest.SpecID)
	if r["contract_sha256"] == "" || r["contract_sha256"] == nil {
		t.Fatalf("contract_sha256 empty")
	}
	head := f.Head()
	expectStr("head_sha", head)
	expectStr("target", "baseBranch")
	scope, _ := r["scope"].(map[string]any)
	if scope == nil {
		t.Fatalf("scope missing")
	}
	if s, _ := scope["base_branch"].(string); s == "" {
		t.Fatalf("scope.base_branch empty")
	}
	if s, _ := scope["head_sha"].(string); s != head {
		t.Fatalf("scope.head_sha = %q, want the audited commit %q", s, head)
	}
	if n, _ := scope["changed_files"].(float64); n < 1 {
		t.Fatalf("scope.changed_files = %v, want >= 1", scope["changed_files"])
	}
	backends, _ := r["backends"].([]any)
	if len(backends) != 3 {
		t.Fatalf("backends = %v, want three entries", backends)
	}
	verdicts := map[string]string{}
	for _, b := range backends {
		m, _ := b.(map[string]any)
		verdicts[m["backend"].(string)] = m["verdict"].(string)
	}
	if verdicts["claude"] != "pass" || verdicts["codex"] != "fail" || verdicts["glm"] != "inconclusive" {
		t.Fatalf("verdicts = %v", verdicts)
	}
	if n, _ := r["participant_count"].(float64); n != 2 {
		t.Fatalf("participant_count = %v, want 2", r["participant_count"])
	}
	if df, ok := r["disagreement_flag"].(bool); !ok || !df {
		t.Fatalf("disagreement_flag = %v, want true", r["disagreement_flag"])
	}
	expectStr("audit_receipt", result.AuditReceipt)
	if s, _ := r["recorded_at"].(string); s == "" {
		t.Fatalf("recorded_at empty")
	}
	if at, _ := r["recorded_at"].(string); at != "" {
		if _, perr := time.Parse(time.RFC3339, at); perr != nil {
			t.Fatalf("recorded_at %q is not RFC3339", at)
		}
	}
}

// TestAC_CLOSURE_012_UnknownCard — an unknown card writes a line with an
// empty spec_id (the reader classifies it unbound).
func TestAC_CLOSURE_012_UnknownCard(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	recordBackends(t)

	_ = runMultiAudit(context.Background(), ReviewOutput{}, "baseBranch", "", MultiAuditConfig{
		ProjectRoot: f.CardDir,
		CardID:      "c9",
	}, nil)
	// c9 has no worktree: its evidence home is the primary checkout.
	records := readSecondReviewLines(t, filepath.Join(f.Root, ".moai", "reports", "c9", closure.SecondReviewFile))
	if len(records) != 1 {
		t.Fatalf("records = %d, want one", len(records))
	}
	if s, _ := records[0]["spec_id"].(string); s != "" {
		t.Fatalf("spec_id = %q, want empty", s)
	}
}

// TestAC_CLOSURE_012_AppendFailure — an unwritable evidence directory sets
// second_review_record_error and the verdict payload equals a run without
// card_id.
func TestAC_CLOSURE_012_AppendFailure(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
	recordBackends(t)

	if err := os.MkdirAll(f.EvidenceDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(f.EvidenceDir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(f.EvidenceDir, 0o755) })

	result := runMultiAudit(context.Background(), ReviewOutput{}, "baseBranch", "", MultiAuditConfig{
		ProjectRoot: f.CardDir,
		CardID:      "c1",
	}, nil)
	if result.SecondReviewRecordError == "" {
		t.Fatalf("second_review_record_error empty, want the append failure")
	}
	// The verdict payload is unchanged by the failed append.
	if result.OverallVerdict == "" || result.ParticipantCount != 2 {
		t.Fatalf("result payload changed: verdict=%q participants=%d", result.OverallVerdict, result.ParticipantCount)
	}
}

// TestAC_CLOSURE_012_NoCardID — byte-identical output shape and no files.
func TestAC_CLOSURE_012_NoCardID(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
	recordBackends(t)

	result := runMultiAudit(context.Background(), ReviewOutput{}, "baseBranch", "", MultiAuditConfig{
		ProjectRoot: f.CardDir,
	}, nil)
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), "second_review_record_error") {
		t.Fatalf("the JSON carries the new key without card_id — not byte-identical to pre-change output")
	}
	// No file exists under any .moai/reports.
	for _, dir := range []string{f.EvidenceDir, filepath.Join(f.Root, ".moai", "reports")} {
		if _, statErr := os.Stat(dir); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("%s exists after a card-less run: %v", dir, statErr)
		}
	}
}

// TestAC_CLOSURE_012_TargetRecorded — the record's target is the REQUESTED
// target, verbatim (REQ-CLOSURE-012), so the scope-covered filter keeps its
// teeth: only a baseBranch review proves baseBranch coverage (REQ-CLOSURE-013).
func TestAC_CLOSURE_012_TargetRecorded(t *testing.T) {
	cases := []struct {
		name   string
		target string
	}{
		{"uncommittedChanges is recorded verbatim", "uncommittedChanges"},
		{"no requested target records the empty default", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := closuretest.New(t)
			fixtureSeams(t, f)
			queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
			recordBackends(t)

			runMultiAudit(context.Background(), ReviewOutput{}, tc.target, "", MultiAuditConfig{
				ProjectRoot: f.CardDir,
				CardID:      "c1",
			}, nil)

			records := readSecondReviewLines(t, filepath.Join(f.EvidenceDir, closure.SecondReviewFile))
			if len(records) != 1 {
				t.Fatalf("records = %d, want one", len(records))
			}
			if got, _ := records[0]["target"].(string); got != tc.target {
				t.Fatalf("recorded target = %q, want the requested %q", got, tc.target)
			}

			// The scope filter regains teeth: the recorded target fails the
			// baseBranch coverage rule, so the record cannot clear the
			// second review.
			line, err := json.Marshal(records[0])
			if err != nil {
				t.Fatalf("re-encode: %v", err)
			}
			var rec closure.SecondReviewRecord
			if err := json.Unmarshal(line, &rec); err != nil {
				t.Fatalf("decode record: %v", err)
			}
			st := closure.SelectSecondReview(closure.SecondReviewInput{
				Records:        []closure.SecondReviewRecord{rec},
				ContractCard:   "c1",
				ContractDigest: rec.ContractSHA256,
				EvalCommit:     f.Head(),
				Facts: closure.GitFacts{
					IsAncestor: func(a, b string) (bool, error) { return a == b, nil },
				},
				SpecID: closuretest.SpecID,
			})
			if st.State != closure.SecondReviewNotPerformed || st.Cause != closure.SecondReviewCauseScopeNotCovered {
				t.Fatalf("state = %q cause %q, want %q/%q", st.State, st.Cause,
					closure.SecondReviewNotPerformed, closure.SecondReviewCauseScopeNotCovered)
			}
		})
	}
}

// TestAC_CLOSURE_012_PathTraversalCardID — a card_id is a trust-boundary
// input (the MCP caller supplies it) that is joined into filesystem paths:
// it is validated against contract.CardPattern before any path is built, so
// a traversal id produces second_review_record_error and writes nothing
// outside the evidence home.
func TestAC_CLOSURE_012_PathTraversalCardID(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
	recordBackends(t)

	result := runMultiAudit(context.Background(), ReviewOutput{}, "baseBranch", "", MultiAuditConfig{
		ProjectRoot: f.CardDir,
		CardID:      "../../../../tmp/zzt1237",
	}, nil)
	if result.SecondReviewRecordError == "" {
		t.Fatalf("second_review_record_error empty, want the invalid card_id rejection")
	}
	escaped := filepath.Join(f.Root, ".moai", "reports", "../../../../tmp/zzt1237")
	if _, err := os.Stat(escaped); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("traversal directory %s was created: %v", escaped, err)
	}
	if _, err := os.Stat(filepath.Join(escaped, closure.SecondReviewFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("second-review.jsonl was written outside the evidence home")
	}
}
