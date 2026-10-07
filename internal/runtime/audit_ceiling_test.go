package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/auditverdict"
)

// Ceiling-engine tests for SPEC-AUDIT-CEILING-001 M3: the round counter
// (REQ-ACE-001), the policy outcome ladder (REQ-ACE-003..007, REQ-ACE-013),
// the delta-eligibility check, the refusal output record, the override, and
// the audit trail.

// writeAuditFixture writes one plan-audit iteration file naming specID and
// returns its path. n<=0 keeps the name as given.
func writeAuditFixture(t *testing.T, dir, name, specID, verdict string, n int) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if name == "" {
		iter := ""
		if n > 0 {
			iter = fmt.Sprintf("-iter%d", n)
		}
		name = fmt.Sprintf("plan-audit%s.md", iter)
	}
	raw := fmt.Sprintf("# SPEC Review Report: %s\nverdict: %s\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: stale\naudited_sha: asha-%s\n", specID, verdict, name)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestCountAuditRounds covers REQ-ACE-001 (AC-ACE-001): the round count is
// the number of distinct (SPEC id, iteration number) pairs across both
// families — the same iteration in both families counts once, a legacy-only
// SPEC is counted, and an unreadable iteration number is never collapsed.
func TestCountAuditRounds(t *testing.T) {
	specID := "SPEC-ACE-COUNT-001"
	reports := t.TempDir()
	specDir := filepath.Join(reports, specID)
	cardDir := filepath.Join(reports, "t9000")
	writeAuditFixture(t, specDir, "", specID, "PASS", 1)
	// The same iteration 1 recorded under a card directory: one round, not two.
	writeAuditFixture(t, cardDir, "", specID, "PASS", 1)
	writeAuditFixture(t, cardDir, "", specID, "PASS", 2)

	ev, err := CountAuditRounds(specID, []string{specDir, cardDir})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 2 {
		t.Fatalf("count %d, want 2 (iter1 in two families collapses; iter2 counts)", ev.Count)
	}
	if len(ev.Sources) != 3 {
		t.Fatalf("sources %v, want the three evidence files", ev.Sources)
	}
	if ev.Latest == nil || ev.LatestPath == "" || !strings.Contains(ev.LatestPath, "iter2") {
		t.Fatalf("latest %v path %q, want the iter2 file", ev.Latest, ev.LatestPath)
	}

	// A legacy-only SPEC still counts (edge 6).
	legacyOnly := t.TempDir()
	legacy := filepath.Join(legacyOnly, specID+"-review-3.md")
	if err := os.WriteFile(legacy, []byte("# review\nverdict: PASS\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ev, err = CountAuditRounds(specID, []string{legacyOnly})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 1 {
		t.Fatalf("legacy-only count %d, want 1", ev.Count)
	}

	// A convention file whose iteration number is unreadable counts as its
	// own round and never collapses into another file.
	odd := t.TempDir()
	writeAuditFixture(t, odd, "plan-audit-iterX.md", specID, "PASS", 0)
	writeAuditFixture(t, odd, "", specID, "PASS", 1)
	ev, err = CountAuditRounds(specID, []string{odd})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 2 {
		t.Fatalf("unreadable-iteration count %d, want 2 (fail-counted)", ev.Count)
	}

	// A convention file naming another SPEC does not count for this one.
	other := t.TempDir()
	writeAuditFixture(t, other, "", "SPEC-ACE-OTHER-002", "PASS", 1)
	ev, err = CountAuditRounds(specID, []string{other})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 0 {
		t.Fatalf("foreign-SPEC count %d, want 0", ev.Count)
	}
}

// TestCountAuditRoundsDistinctN covers AC-ACE-020: two audits of the same
// SPEC on unchanged artifacts recorded as iteration N and N+1 count 2 —
// different iteration numbers are different rounds even when the verdict
// label, score, and audited state are identical.
func TestCountAuditRoundsDistinctN(t *testing.T) {
	specID := "SPEC-ACE-DUP-001"
	dir := t.TempDir()
	body := "# SPEC Review Report: " + specID + "\nverdict: FAIL\nOverall Score: 0.84\nmust_pass_failed: 0\nblocking_count: 1\nplan_artifact_hash: samehash\naudited_sha: samesha\n"
	for _, name := range []string{"plan-audit-iter1.md", "plan-audit-iter2.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ev, err := CountAuditRounds(specID, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 2 {
		t.Fatalf("identical-audit count %d, want 2", ev.Count)
	}
}

// ceilingFixture builds a project root whose SPEC directory, harness config,
// and report directories drive the engine. The fixture ceiling L=1 with
// auto_delta_rounds 0 keeps the arithmetic one round away.
type ceilingFixture struct {
	root    string
	specID  string
	cardID  string
	specDir string
}

func newCeilingFixture(t *testing.T) *ceilingFixture {
	f := &ceilingFixture{root: t.TempDir(), specID: "SPEC-ACE-ENG-001", cardID: "t9001"}
	f.specDir = filepath.Join(f.root, ".moai", "specs", f.specID)
	if err := os.MkdirAll(f.specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.specDir, "spec.md"), []byte("---\ntier: L\n---\n# spec\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// plan.md must exist so the plan-artifact hash can bind.
	if err := os.WriteFile(filepath.Join(f.specDir, "plan.md"), []byte("# plan\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(f.root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	harness := "harness:\n  evaluator:\n    memory_scope: per_iteration\n  plan_audit_tier_ceilings:\n    S: 1\n    M: 2\n    L: 1\n  plan_audit_ceiling_policy:\n    auto_delta_rounds: 0\n    on_final_hit: hold-and-split\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "harness.yaml"), []byte(harness), 0o600); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *ceilingFixture) writeIter(t *testing.T, n int, verdict string) string {
	t.Helper()
	return writeAuditFixture(t, filepath.Join(f.root, ".moai", "reports", f.cardID), "", f.specID, verdict, n)
}

func (f *ceilingFixture) hash(t *testing.T) string {
	t.Helper()
	h, err := NewInMemoryCache().ComputeHash(f.specDir)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// appendLine appends one raw line to a fixture verdict file.
func (f *ceilingFixture) appendLine(t *testing.T, path, line string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n"+line+"\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
}

// fieldsFromHash re-reads a fixture verdict file with its plan_artifact_hash
// line rewritten to the current plan-artifact hash, so hashOK reflects the
// engine's own binding check.
func (f *ceilingFixture) fieldsFromHash(t *testing.T, path string) (auditverdict.Fields, bool) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "plan_artifact_hash:") {
			lines[i] = "plan_artifact_hash: " + f.hash(t)
		}
	}
	fields := auditverdict.Parse([]byte(strings.Join(lines, "\n")))
	return fields, fields.PlanArtifactHash == f.hash(t)
}

// input is the VerdictCeilingInput for the fixture.
func (f *ceilingFixture) input() VerdictCeilingInput {
	return VerdictCeilingInput{SpecID: f.specID, SpecDir: f.specDir, ProjectRoot: f.root, CardID: f.cardID}
}

// TestCeilingRefusal covers AC-ACE-003: at the effective ceiling a verdict
// failing admission refuses before the consuming consumer admits it, and an
// admission-clean verdict at the same ceiling state emits no refusal (the
// V4-D1 pass-through exclusion arm; REQ-ACE-013 carries the round).
func TestCeilingRefusal(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")

	t.Run("failing_verdict_refuses_at_ceiling", func(t *testing.T) {
		fields, hashOK := f.fieldsFromHash(t, path)
		oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
		if err != nil {
			t.Fatal(err)
		}
		if oc == nil {
			t.Fatal("no ceiling outcome at the effective ceiling")
		}
		if override {
			t.Fatal("a hash-failing verdict is never debt-admitted")
		}
		if !oc.Blocked {
			t.Fatalf("outcome %s not blocked", oc.Outcome)
		}
	})

	t.Run("clean_verdict_passes_through_at_ceiling", func(t *testing.T) {
		fields, hashOK := f.fieldsFromHash(t, path)
		fields.Label = "PASS"
		oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
		if err != nil {
			t.Fatal(err)
		}
		if oc == nil || oc.Outcome != OutcomePassThrough || oc.Blocked || override {
			t.Fatalf("clean verdict at ceiling: outcome %+v override %v, want pass-through unblocked", oc, override)
		}
	})
}

// TestCeilingPolicyDebtAdmit covers AC-ACE-004's first arm: a verdict
// failing admission on the label alone (score, must-pass, blocking, hash all
// pass, at least one finding enumerated as a machine debt line) debt-admits
// — the outcome record carries PASS-WITH-DEBT with the findings as debts
// carrying dispose_in, and entry is admitted at the consuming seam.
func TestCeilingPolicyDebtAdmit(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	f.appendLine(t, path, "- debt: D1 dispose_in=run fixture debt")
	fields, hashOK := f.fieldsFromHash(t, path)
	if len(fields.Debts) != 1 {
		t.Fatalf("fixture debts %+v", fields.Debts)
	}
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomeDebtAdmit || oc.Blocked || !override {
		t.Fatalf("outcome %+v override %v, want debt-admit admitting", oc, override)
	}
	if len(oc.Debts) == 0 {
		t.Fatal("debt-admit outcome enumerates no debts")
	}
}

// TestCeilingPolicyReceiptHold covers AC-ACE-004's second arm (D20): a
// verdict a required backend refuses never converts to debt — work item 1
// cannot admit what work item 2 refuses.
func TestCeilingPolicyReceiptHold(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	fields, hashOK := f.fieldsFromHash(t, path)
	fields.Label = "PASS-WITH-DEBT" // a passing label the receipt must still refuse
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, []string{"claude"})
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil {
		t.Fatal("no ceiling outcome")
	}
	if override {
		t.Fatal("a receipt-refusing verdict was converted to debt-admit")
	}
	if !oc.Blocked {
		t.Fatalf("outcome %s not blocked", oc.Outcome)
	}
}

// TestCeilingPolicySplit covers AC-ACE-005: blocking findings each carrying
// a scoped fix anchor split — hold record plus a split proposal naming the
// anchored scope.
func TestCeilingPolicySplit(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	f.appendLine(t, path, "fix_scope: spec.md#REQ-ACE-003 plan.md#M3")
	fields, hashOK := f.fieldsFromHash(t, path)
	fields.BlockingKnown, fields.BlockingCount = true, 2
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomeSplit || !oc.Blocked || override {
		t.Fatalf("outcome %+v override %v, want split blocked", oc, override)
	}
	if !strings.Contains(strings.Join(oc.Reasons, " "), "spec.md#REQ-ACE-003") {
		t.Fatalf("split proposal does not name the anchored scope: %v", oc.Reasons)
	}
}

// TestCeilingPolicyHold covers AC-ACE-006: a verdict matching neither
// debt-admit nor split eligibility holds, and the hold names its release
// path (D30).
func TestCeilingPolicyHold(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	f.appendLine(t, path, "must_pass_failed: 1") // a must-pass failure no outcome arm admits
	fields, hashOK := f.fieldsFromHash(t, path)
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomeHold || !oc.Blocked || override {
		t.Fatalf("outcome %+v override %v, want hold blocked", oc, override)
	}
	if !strings.Contains(strings.Join(oc.Reasons, " "), "release path") {
		t.Fatalf("hold record names no release path: %v", oc.Reasons)
	}
}

// TestCeilingPolicyHashHold covers AC-ACE-017 (edge 7): a stale verdict
// whose plan_artifact_hash no longer binds holds — never debt-admission —
// and the counter does not reset on artifact edits.
func TestCeilingPolicyHashHold(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "PASS")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fields := auditverdict.Parse(raw) // the fixture's hash line is stale on purpose
	hashOK := fields.PlanArtifactHash == f.hash(t)
	if hashOK {
		t.Fatal("premise: stale hash should not bind")
	}
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomeHold || !oc.Blocked || override {
		t.Fatalf("outcome %+v override %v, want hold blocked on a stale hash", oc, override)
	}
}

// TestCeilingPolicyPassThrough covers AC-ACE-013 (D31; V4-D1's second
// face): a verdict satisfying every admission check admits at any ceiling
// state — the tier-ceiling final hit included — and the outcome record names
// that the ceiling was reached.
func TestCeilingPolicyPassThrough(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "PASS")
	fields, hashOK := f.fieldsFromHash(t, path)
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomePassThrough || oc.Blocked || override {
		t.Fatalf("outcome %+v override %v, want pass-through", oc, override)
	}
	if !strings.Contains(strings.Join(oc.Reasons, " "), "ceiling") {
		t.Fatalf("outcome record does not name the ceiling: %v", oc.Reasons)
	}
}

// TestCeilingDeltaEligibility covers AC-ACE-021: at the tier ceiling a
// verdict with no fix_scope anchors (or a STOP signal, or changed REQ/AC id
// sets, or a diff outside the anchors) refuses immediately as a final hit;
// an eligible delta is granted without refusal. D1: a label-only-failing
// final hit without eligible anchors debt-admits rather than holding.
func TestCeilingDeltaEligibility(t *testing.T) {
	// No anchors: final hit — no delta round is granted.
	if DeltaEligible(nil, false, true, true) {
		t.Fatal("a verdict with no fix_scope anchors earned a delta round")
	}
	// Anchors + in-anchor diff + unchanged REQ/AC ids + no STOP: granted.
	if !DeltaEligible([]string{"spec.md#REQ-ACE-003"}, false, true, true) {
		t.Fatal("an eligible delta was refused")
	}
	// A STOP signal is a final hit regardless of anchors.
	if DeltaEligible([]string{"spec.md#REQ-ACE-003"}, true, true, true) {
		t.Fatal("a STOP signal earned a delta round")
	}
	// Changed REQ/AC id sets are a final hit.
	if DeltaEligible([]string{"spec.md#REQ-ACE-003"}, false, true, false) {
		t.Fatal("changed REQ/AC id sets earned a delta round")
	}
	// A diff outside the anchors is a final hit.
	if DeltaEligible([]string{"spec.md#REQ-ACE-003"}, false, false, true) {
		t.Fatal("an out-of-anchor diff earned a delta round")
	}

	// The fixture's effective ceiling (L=1 + delta 0) makes iteration 1 the
	// final hit: the label-only-failing verdict debt-admits (D1's arm).
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	f.appendLine(t, path, "- debt: D1 dispose_in=run fixture debt")
	fields, hashOK := f.fieldsFromHash(t, path)
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomeDebtAdmit || !override {
		t.Fatalf("label-only final hit: outcome %+v override %v, want debt-admit (D1)", oc, override)
	}
}

// TestCeilingRefusalOutput covers AC-ACE-007: the CLI's structured output
// carries outcome + reasons + evidence paths, the outcome persists to
// progress.md §G, and the path completes without an interactive prompt.
func TestCeilingRefusalOutput(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	fields, hashOK := f.fieldsFromHash(t, path)
	oc, _, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil {
		t.Fatal("no outcome")
	}
	if oc.Outcome == "" || len(oc.Reasons) == 0 || len(oc.Evidence) == 0 {
		t.Fatalf("structured output incomplete: %+v", oc)
	}
	// The outcome persisted to the SPEC's progress.md §G.
	progressRaw, err := os.ReadFile(filepath.Join(f.specDir, "progress.md"))
	if err != nil {
		t.Fatalf("progress.md not written: %v", err)
	}
	if !strings.Contains(string(progressRaw), "Override and Refusal Record") || !strings.Contains(string(progressRaw), oc.Outcome) {
		t.Fatalf("progress.md §G does not carry the outcome: %.200s", progressRaw)
	}
	// The audit trail appended one line.
	trailRaw, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log"))
	if err != nil {
		t.Fatalf("trail not written: %v", err)
	}
	if line := string(trailRaw); !strings.Contains(line, f.specID) || !strings.Contains(line, oc.Outcome) {
		t.Fatalf("trail line incomplete: %s", line)
	}
}

// TestRequiredBackendOverride covers AC-ACE-011: the override names the
// backend and carries a mandatory note; an empty note is refused (edge 8);
// the ack lands in progress.md §G and the audit trail.
func TestRequiredBackendOverride(t *testing.T) {
	f := newCeilingFixture(t)
	if err := AcknowledgeRequiredBackend(f.input(), "claude", ""); err == nil {
		t.Fatal("empty override note accepted")
	}
	if err := AcknowledgeRequiredBackend(f.input(), "claude", "operator accepted the stale claude gate"); err != nil {
		t.Fatalf("valid override refused: %v", err)
	}
	progressRaw, err := os.ReadFile(filepath.Join(f.specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(progressRaw), "claude") || !strings.Contains(string(progressRaw), "operator accepted the stale claude gate") {
		t.Fatalf("override not recorded: %.200s", progressRaw)
	}
	trailRaw, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(trailRaw), "override") || !strings.Contains(string(trailRaw), "claude") {
		t.Fatalf("override trail line incomplete: %s", trailRaw)
	}
}

// TestAuditTrailAppend covers AC-ACE-012: every refusal or override appends
// a durable trail entry with timestamp, SPEC, and reason.
func TestAuditTrailAppend(t *testing.T) {
	f := newCeilingFixture(t)
	if err := appendAuditTrail(f.input(), "ceiling-refusal", "hold", "round count 1 reached the ceiling"); err != nil {
		t.Fatal(err)
	}
	if err := appendAuditTrail(f.input(), "override", "ack", "claude acknowledged"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 2 {
		t.Fatalf("trail has %d lines, want 2", len(lines))
	}
	if !strings.Contains(lines[0], f.specID) || !strings.Contains(lines[0], "ceiling-refusal") {
		t.Fatalf("refusal trail line incomplete: %s", lines[0])
	}
	if !strings.Contains(lines[1], f.specID) || !strings.Contains(lines[1], "override") {
		t.Fatalf("override trail line incomplete: %s", lines[1])
	}
}

// TestRoundReportDirs covers the SPEC-attribution enumeration (REQ-ACE-001):
// the SPEC-scoped directory, the card directory, and every foreign card
// directory whose plan-audit iteration files name the SPEC — and never the
// daily run-history directory.
func TestRoundReportDirs(t *testing.T) {
	specID := "SPEC-ACE-DIRS-001"
	root := t.TempDir()
	reports := filepath.Join(root, ".moai", "reports")
	naming := filepath.Join(reports, "t8001")
	foreign := filepath.Join(reports, "t8002")
	writeAuditFixture(t, naming, "", specID, "PASS", 1)
	writeAuditFixture(t, foreign, "", "SPEC-ACE-OTHER-009", "PASS", 1)
	if err := os.MkdirAll(filepath.Join(reports, "plan-audit"), 0o755); err != nil {
		t.Fatal(err)
	}

	dirs := RoundReportDirs(root, specID, "t8001")
	joined := strings.Join(dirs, "\n")
	if !strings.Contains(joined, filepath.Join(reports, specID)) {
		t.Fatalf("SPEC-scoped dir missing: %v", dirs)
	}
	if !strings.Contains(joined, naming) {
		t.Fatalf("card dir missing: %v", dirs)
	}
	if strings.Contains(joined, foreign) {
		t.Fatalf("a foreign card dir not naming the SPEC was included: %v", dirs)
	}
	if strings.Contains(joined, filepath.Join(reports, "plan-audit")) {
		t.Fatalf("the daily run-history directory was included: %v", dirs)
	}
}

// TestResolveRequiredBackendsInRuntime covers the gate-set resolver from its
// own package (the AC's TestAdmitConfigErrorRefused runs it from the
// auditverdict external test binary, which does not attribute coverage
// here): absent config resolves empty (C4), an unreadable config errors
// (D21), and a valid required gate resolves.
func TestResolveRequiredBackendsInRuntime(t *testing.T) {
	dir := t.TempDir()
	gs, err := ResolveRequiredBackends(dir)
	if err != nil || len(gs.Required) != 0 {
		t.Fatalf("absent config: %+v %v", gs, err)
	}
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cfgDir, "workflow.yaml")
	if err := os.WriteFile(path, []byte("workflow: [unclosed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveRequiredBackends(dir); err == nil {
		t.Fatal("unparseable workflow.yaml resolved as empty")
	}
	if err := os.WriteFile(path, []byte("workflow:\n  audit:\n    gates:\n      claude: required\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gs, err = ResolveRequiredBackends(dir)
	if err != nil || len(gs.Required) != 1 || gs.Required[0] != "claude" {
		t.Fatalf("valid config: %+v %v", gs, err)
	}
}

// TestDeltaGitHelpers covers the git-backed delta-eligibility helpers with a
// real repository: an in-anchor diff and unchanged REQ/AC id sets verify, an
// out-of-anchor diff and a changed id set fail, and unresolvable SHAs are
// fail-closed (D6).
func TestDeltaGitHelpers(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := auditreceipt.RunScrubbedGit(root, args...)
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(out)
	}
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-ACE-GIT-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	git("init", "-q")
	git("config", "user.email", "t@example.com")
	git("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# spec\nREQ-ACE-001 AC-ACE-001\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	first := git("rev-parse", "HEAD")

	// An in-anchor change: the fix_scope-anchored spec.md itself, with its
	// REQ/AC id set unchanged.
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# spec\nREQ-ACE-001 AC-ACE-001\nrepaired wording\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "in-anchor")
	second := git("rev-parse", "HEAD")

	if !diffInsideAnchors(root, first, second, []string{".moai/specs/SPEC-ACE-GIT-001/spec.md#REQ-ACE-001"}) {
		t.Fatal("an in-anchor diff did not verify")
	}
	if !reqACSetsUnchanged(root, "SPEC-ACE-GIT-001", first, second) {
		t.Fatal("unchanged REQ/AC id sets did not verify")
	}

	// An out-of-anchor change: an unanchored file.
	if err := os.WriteFile(filepath.Join(root, "unrelated.md"), []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "outside")
	third := git("rev-parse", "HEAD")
	if diffInsideAnchors(root, second, third, []string{".moai/specs/SPEC-ACE-GIT-001/spec.md#REQ-ACE-001"}) {
		t.Fatal("an out-of-anchor diff verified")
	}

	// A changed REQ/AC id set fails.
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# spec\nREQ-ACE-001 REQ-ACE-002 AC-ACE-001\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "ids")
	fourth := git("rev-parse", "HEAD")
	if reqACSetsUnchanged(root, "SPEC-ACE-GIT-001", third, fourth) {
		t.Fatal("a changed REQ/AC id set verified")
	}

	// Unresolvable SHAs are fail-closed.
	if diffInsideAnchors(root, "deadbeef", fourth, []string{"x#y"}) {
		t.Fatal("unresolvable SHAs verified an in-anchor diff")
	}
	if reqACSetsUnchanged(root, "SPEC-ACE-GIT-001", "deadbeef", fourth) {
		t.Fatal("unresolvable SHAs verified unchanged id sets")
	}
}

// TestAttachCeiling covers GateConfig's Step 3.5: the ceiling outcome
// attaches to the AuditResult without changing the Verdict, and a blocked
// outcome is observable via Ceiling.Blocked (design.md §7).
func TestAttachCeiling(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	c := &GateConfig{
		SpecID:     f.specID,
		SpecDir:    f.specDir,
		ProjectDir: f.root,
		Cache:      NewInMemoryCache(),
	}
	res := &AuditResult{Verdict: VerdictFail, ReportPath: path}
	c.attachCeiling(res)
	if res.Ceiling == nil {
		t.Fatal("no ceiling outcome attached")
	}
	if res.Verdict != VerdictFail {
		t.Fatalf("verdict rewritten to %s", res.Verdict)
	}
	if !res.Ceiling.Blocked {
		t.Fatalf("blocked outcome not observable: %+v", res.Ceiling)
	}
}

// writeRoundFile writes one plan-audit round file into dir.
func writeRoundFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("verdict: PASS\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// evidenceDir makes a fresh directory holding the named round files.
func evidenceDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		writeRoundFile(t, dir, name)
	}
	return dir
}

// TestCountPlanAuditRounds — AC-ACR-001: one recorded file one round, counted
// from durable evidence on disk across the SPEC-scoped directory (absent → 0)
// and the explicitly listed directories (absent → error, tested by
// TestCountPlanAuditRoundsListedDirMissing); identical listed paths count once.
func TestCountPlanAuditRounds(t *testing.T) {
	t.Parallel()

	t.Run("plan-audit.md plus iter1..3 counts 4", func(t *testing.T) {
		t.Parallel()
		dir := evidenceDir(t, "plan-audit.md", "plan-audit-iter1.md", "plan-audit-iter2.md", "plan-audit-iter3.md")
		got, err := CountPlanAuditRounds(dir, nil)
		if err != nil {
			t.Fatalf("CountPlanAuditRounds: %v", err)
		}
		if got != 4 {
			t.Fatalf("count = %d, want 4", got)
		}
	})

	t.Run("SPEC-scoped directory absent contributes 0", func(t *testing.T) {
		t.Parallel()
		absent := filepath.Join(t.TempDir(), "SPEC-UNAUDITED-001")
		got, err := CountPlanAuditRounds(absent, nil)
		if err != nil {
			t.Fatalf("absent SPEC-scoped directory must not error: %v", err)
		}
		if got != 0 {
			t.Fatalf("count = %d, want 0", got)
		}
	})

	t.Run("identical listed paths are deduplicated", func(t *testing.T) {
		t.Parallel()
		dir := evidenceDir(t, "plan-audit-iter1.md")
		dup := filepath.Join(dir, "nested")
		if err := os.MkdirAll(dup, 0o755); err != nil {
			t.Fatal(err)
		}
		// The same directory listed twice contributes once; a path equal to the
		// SPEC-scoped one does not double it either.
		got, err := CountPlanAuditRounds(dir, []string{dir, filepath.Clean(dir) + "/", dir})
		if err != nil {
			t.Fatalf("CountPlanAuditRounds: %v", err)
		}
		if got != 1 {
			t.Fatalf("count = %d, want 1 (identical listings deduplicate)", got)
		}
	})

	t.Run("non round-family files are ignored", func(t *testing.T) {
		t.Parallel()
		dir := evidenceDir(t, "plan-audit-iter1.md", "sync-audit.md", "notes.md", "plan-audit.md.bak")
		got, err := CountPlanAuditRounds(dir, nil)
		if err != nil {
			t.Fatalf("CountPlanAuditRounds: %v", err)
		}
		if got != 1 {
			t.Fatalf("count = %d, want 1 (only the iter family counts)", got)
		}
	})
}

// TestCountPlanAuditRoundsUnparseable — AC-ACR-002: a round file whose
// iteration suffix does not parse as a positive integer makes the count an
// error naming the file, never a silent skip.
func TestCountPlanAuditRoundsUnparseable(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"plan-audit-iterX.md", "plan-audit-iter0.md", "plan-audit-iter-1.md", "plan-audit-iter.md"} {
		dir := evidenceDir(t, name)
		_, err := CountPlanAuditRounds(dir, nil)
		if err == nil {
			t.Errorf("%s: expected an error, got nil", name)
			continue
		}
		if !strings.Contains(err.Error(), name) {
			t.Errorf("%s: error does not name the file: %v", name, err)
		}
	}
}

// TestCountPlanAuditRoundsListedDirMissing — AC-ACR-013: an EXPLICITLY LISTED
// evidence directory that does not exist (a typo) is an error naming the
// missing path, while the SPEC-scoped directory being absent still contributes
// 0 (an unaudited SPEC is not an error).
func TestCountPlanAuditRoundsListedDirMissing(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "typo-evidence")
	_, err := CountPlanAuditRounds(filepath.Join(t.TempDir(), "SPEC-UNAUDITED-002"), []string{missing})
	if err == nil {
		t.Fatal("a missing explicitly listed directory must error, got nil")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("error does not name the missing path: %v", err)
	}
}

// TestLatestVerdictTieIsError — AC-ACR-001's tie arm: the same highest
// iteration number in two evidence directories is a selection error rather
// than a silent pick.
func TestLatestVerdictTieIsError(t *testing.T) {
	t.Parallel()

	dirA := evidenceDir(t, "plan-audit-iter2.md")
	dirB := evidenceDir(t, "plan-audit-iter2.md")
	_, err := SelectLatestVerdict(dirA, []string{dirB})
	if err == nil {
		t.Fatal("a cross-directory iteration tie must error, got nil")
	}

	// A strict ordering across directories selects the highest.
	dirC := evidenceDir(t, "plan-audit.md", "plan-audit-iter5.md")
	got, err := SelectLatestVerdict(dirA, []string{dirC})
	if err != nil {
		t.Fatalf("SelectLatestVerdict: %v", err)
	}
	if !strings.HasSuffix(got, filepath.Join("plan-audit-iter5.md")) || !strings.Contains(got, filepath.Base(dirC)) {
		t.Fatalf("latest = %q, want dirC's plan-audit-iter5.md", got)
	}

	// plan-audit.md alone ranks as iteration 0 and is selected.
	dirD := evidenceDir(t, "plan-audit.md")
	got, err = SelectLatestVerdict(dirD, nil)
	if err != nil {
		t.Fatalf("SelectLatestVerdict: %v", err)
	}
	if !strings.HasSuffix(got, "plan-audit.md") {
		t.Fatalf("latest = %q, want plan-audit.md", got)
	}

	// No evidence at all selects nothing and does not error.
	got, err = SelectLatestVerdict(filepath.Join(t.TempDir(), "SPEC-UNAUDITED-003"), nil)
	if err != nil {
		t.Fatalf("empty evidence must not error: %v", err)
	}
	if got != "" {
		t.Fatalf("latest = %q, want empty", got)
	}
}

// TestResolvePlanAuditCeiling — AC-ACR-003's resolution arms: the shipped map
// resolves S:1 M:2 L:3, and an absent or unknown tier resolves to L
// (auditverdict.SpecTier's rule).
func TestResolvePlanAuditCeiling(t *testing.T) {
	t.Parallel()

	shipped := map[string]int{"S": 1, "M": 2, "L": 3}
	cases := []struct {
		tier string
		want int
	}{
		{"S", 1},
		{"M", 2},
		{"L", 3},
		{"", 3},      // absent tier → L
		{"bogus", 3}, // unknown tier → L
	}
	for _, c := range cases {
		got, err := ResolvePlanAuditCeiling(c.tier, shipped)
		if err != nil {
			t.Fatalf("tier %q: %v", c.tier, err)
		}
		if got != c.want {
			t.Errorf("tier %q: ceiling = %d, want %d", c.tier, got, c.want)
		}
	}
}

// TestResolvePlanAuditCeilingInvalid — AC-ACR-003's configuration-error arms:
// a ceilings map missing the resolved tier's key, or resolving ≤ 0, is a
// configuration error, never a ceiling of zero.
func TestResolvePlanAuditCeilingInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		tier     string
		ceilings map[string]int
	}{
		{"missing tier key", "M", map[string]int{"S": 1, "L": 3}},
		{"missing resolved L key", "", map[string]int{"S": 1, "M": 2}},
		{"zero ceiling", "M", map[string]int{"S": 1, "M": 0, "L": 3}},
		{"negative ceiling", "L", map[string]int{"S": 1, "M": 2, "L": -2}},
		{"nil map", "S", nil},
	}
	for _, c := range cases {
		got, err := ResolvePlanAuditCeiling(c.tier, c.ceilings)
		if err == nil {
			t.Errorf("%s: expected a configuration error, got ceiling %d", c.name, got)
		}
	}
}

// TestEvidenceRelativeAndAbsoluteSpellingsDeduplicate — CR-P2-2 (card-review
// r1): the same directory listed as a relative path and as its absolute form
// contributes exactly once. Pre-fix the two spellings dodged the dedupe, the
// count doubled, and SelectLatestVerdict misread the same directory as a
// cross-directory tie.
func TestEvidenceRelativeAndAbsoluteSpellingsDeduplicate(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "evidence")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeRoundFile(t, dir, "plan-audit-iter1.md")
	t.Chdir(base) // the relative spelling resolves against the process cwd

	count, err := CountPlanAuditRounds(dir, []string{"evidence"})
	if err != nil {
		t.Fatalf("CountPlanAuditRounds: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1 (relative + absolute spellings of one directory)", count)
	}

	latest, err := SelectLatestVerdict(dir, []string{"evidence"})
	if err != nil {
		t.Fatalf("one directory in two spellings must not read as a tie: %v", err)
	}
	if !strings.HasSuffix(latest, "plan-audit-iter1.md") {
		t.Fatalf("latest = %q, want the directory's plan-audit-iter1.md", latest)
	}
}

// Verdict bodies the ceiling fixtures record into their latest iteration file.
const (
	ceilingPassBody = "verdict: PASS\noverall_score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: abc123\n"
	ceilingDebtBody = "verdict: PASS-WITH-DEBT\noverall_score: 0.88\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: abc123\n" +
		"- debt: D1 dispose_in=run plan row omits an edit\n" +
		"- debt: D2 dispose_in=sync trace is indirect\n"
	ceilingFailBody = "verdict: FAIL\noverall_score: 0.70\nmust_pass_failed: 2\nblocking_count: 1\nplan_artifact_hash: abc123\n"
)

// ceilingEvidenceFixture writes an evidence directory whose round count (3) sits at or
// above every shipped ceiling and whose latest iteration carries body.
func ceilingEvidenceFixture(t *testing.T, body string) string {
	t.Helper()
	dir := evidenceDir(t, "plan-audit.md", "plan-audit-iter1.md")
	writeRoundFile(t, dir, "plan-audit-iter2.md")
	if err := os.WriteFile(filepath.Join(dir, "plan-audit-iter2.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

// ceilingInput builds a CeilingInput over the fixture with the shipped
// ceilings, the Tier M threshold, and the given policy value.
func ceilingInput(specID, evidenceDir, onFinalHit string) CeilingInput {
	return CeilingInput{
		SpecID:          specID,
		SpecEvidenceDir: evidenceDir,
		Tier:            "M",
		Threshold:       0.80,
		Ceilings:        map[string]int{"S": 1, "M": 2, "L": 3},
		OnFinalHit:      onFinalHit,
	}
}

// TestRecordCeilingOutcome — AC-ACR-004's runtime arm: a non-admitted verdict
// under the shipped hold-and-split policy records exactly one JSON record
// carrying disposition hold, the split-proposal reference, and the
// count/ceiling/label/evidence fields.
func TestRecordCeilingOutcome(t *testing.T) {
	project := t.TempDir()

	outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-004", ceilingEvidenceFixture(t, ceilingFailBody), "hold-and-split"))
	if err != nil {
		t.Fatalf("EvaluatePlanAuditCeiling: %v", err)
	}
	if !hit {
		t.Fatal("3 rounds against a Tier M ceiling of 2 must hit the ceiling")
	}
	if outcome.Disposition != CeilingDispositionHold || outcome.SplitProposalRef == "" {
		t.Fatalf("hold-and-split must record hold + a split-proposal reference, got %+v", outcome)
	}
	if outcome.Count != 3 || outcome.Ceiling != 2 {
		t.Errorf("count/ceiling = %d/%d, want 3/2", outcome.Count, outcome.Ceiling)
	}
	if outcome.VerdictLabel != "FAIL" {
		t.Errorf("verdict label = %q, want FAIL", outcome.VerdictLabel)
	}
	if len(outcome.EvidencePaths) == 0 {
		t.Error("record carries no evidence paths")
	}

	if err := RecordCeilingOutcome(project, "SPEC-CEIL-004", outcome); err != nil {
		t.Fatalf("RecordCeilingOutcome: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEIL-004.json"))
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	var got CeilingOutcome
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("record is not JSON: %v", err)
	}
	if got.Disposition != "hold" || got.SplitProposalRef == "" || got.Count != 3 || got.Ceiling != 2 || got.VerdictLabel != "FAIL" || len(got.EvidencePaths) == 0 {
		t.Errorf("written record = %+v", got)
	}

	// Recording again overwrites the same path — one record per SPEC, not an
	// append log.
	if err := RecordCeilingOutcome(project, "SPEC-CEIL-004", outcome); err != nil {
		t.Fatalf("re-record: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(project, ".moai", "state", "audit-ceiling"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("record directory holds %d entries, want 1", len(entries))
	}
}

// TestRecordCeilingOutcomeDebtProceed — AC-ACR-009: a PASS-WITH-DEBT verdict
// passing the full shared predicate records debt-proceed and references the
// verdict's debt ids.
func TestRecordCeilingOutcomeDebtProceed(t *testing.T) {
	t.Parallel()

	outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-009", ceilingEvidenceFixture(t, ceilingDebtBody), "hold-and-split"))
	if err != nil {
		t.Fatalf("EvaluatePlanAuditCeiling: %v", err)
	}
	if !hit {
		t.Fatal("expected the ceiling to apply")
	}
	if outcome.Disposition != CeilingDispositionDebtProceed {
		t.Fatalf("disposition = %q, want debt-proceed", outcome.Disposition)
	}
	if len(outcome.DebtIDs) != 2 || outcome.DebtIDs[0] != "D1" || outcome.DebtIDs[1] != "D2" {
		t.Errorf("debt ids = %v, want [D1 D2]", outcome.DebtIDs)
	}
	if outcome.VerdictAdmitted != true {
		t.Error("a debt-proceed record is an admitted verdict")
	}
}

// TestRecordCeilingOutcomeUnknownPolicy — AC-ACR-010: a policy value that is
// neither hold-and-split nor split — or an unreadable one — records hold with
// no split-proposal reference (fail-closed).
func TestRecordCeilingOutcomeUnknownPolicy(t *testing.T) {
	t.Parallel()

	for _, policy := range []string{"escalate", ""} {
		outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-010", ceilingEvidenceFixture(t, ceilingFailBody), policy))
		if err != nil {
			t.Fatalf("policy %q: %v", policy, err)
		}
		if !hit {
			t.Fatalf("policy %q: expected the ceiling to apply", policy)
		}
		if outcome.Disposition != CeilingDispositionHold {
			t.Errorf("policy %q: disposition = %q, want hold", policy, outcome.Disposition)
		}
		if outcome.SplitProposalRef != "" {
			t.Errorf("policy %q: carry-over split reference %q, want none", policy, outcome.SplitProposalRef)
		}
	}
}

// TestRecordCeilingOutcomeSplitValue — AC-ACR-012: a policy value of exactly
// split records the split disposition.
func TestRecordCeilingOutcomeSplitValue(t *testing.T) {
	t.Parallel()

	outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-012", ceilingEvidenceFixture(t, ceilingFailBody), "split"))
	if err != nil {
		t.Fatalf("EvaluatePlanAuditCeiling: %v", err)
	}
	if !hit {
		t.Fatal("expected the ceiling to apply")
	}
	if outcome.Disposition != CeilingDispositionSplit {
		t.Fatalf("disposition = %q, want split", outcome.Disposition)
	}
	if outcome.SplitProposalRef != "" {
		t.Errorf("split record carries a split-proposal reference %q; the reference belongs to hold-and-split", outcome.SplitProposalRef)
	}
}

// TestEvaluatePlanAuditCeiling — AC-ACR-005: a clean admitted PASS at the
// ceiling is not a ceiling outcome at all (the composed evaluate+record path
// writes nothing), and an invalid resolved ceiling propagates the
// configuration error before any comparison, writing no record (R2's
// Evaluate-level arm).
func TestEvaluatePlanAuditCeiling(t *testing.T) {
	t.Run("clean admitted PASS at the ceiling writes no record", func(t *testing.T) {
		project := t.TempDir()

		outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-005", ceilingEvidenceFixture(t, ceilingPassBody), "hold-and-split"))
		if err != nil {
			t.Fatalf("EvaluatePlanAuditCeiling: %v", err)
		}
		if hit {
			t.Fatalf("a clean admitted PASS at the ceiling must not apply the ceiling, got %+v", outcome)
		}
		// The composition the CLI verb runs: record only on hit.
		if hit {
			if err := RecordCeilingOutcome(project, "SPEC-CEIL-005", outcome); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := os.Stat(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEIL-005.json")); !os.IsNotExist(err) {
			t.Errorf("record exists for a clean PASS at the ceiling (stat err = %v)", err)
		}
	})

	t.Run("invalid resolved ceiling propagates the error with no record", func(t *testing.T) {
		project := t.TempDir()

		in := ceilingInput("SPEC-CEIL-005B", ceilingEvidenceFixture(t, ceilingFailBody), "hold-and-split")
		in.Tier = "M"
		in.Ceilings = map[string]int{"S": 1, "L": 3} // M missing → configuration error
		_, hit, err := EvaluatePlanAuditCeiling(in)
		if err == nil {
			t.Fatal("a missing resolved ceiling must propagate the configuration error")
		}
		if hit {
			t.Error("no ceiling outcome applies when the ceiling never resolved")
		}
		if _, statErr := os.Stat(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEIL-005B.json")); !os.IsNotExist(statErr) {
			t.Errorf("record written despite the configuration error (stat err = %v)", statErr)
		}
	})

	t.Run("rounds below the ceiling do not apply it", func(t *testing.T) {
		t.Parallel()
		dir := evidenceDir(t, "plan-audit.md") // 1 round
		_, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-LOW", dir, "hold-and-split"))
		if err != nil {
			t.Fatalf("EvaluatePlanAuditCeiling: %v", err)
		}
		if hit {
			t.Error("1 round against a Tier M ceiling of 2 must not hit")
		}
	})
}

// SPEC-AUDIT-CEILING-REPAIR-001 reproduction tests (card t1560). Each RED
// test below was observed failing on unmodified main 903ccd028 for its
// stated reason before its fix, per the engine family convention
// ("RED is a new test — E8 evidence required").

// AC-ACR-004 (D1 end-to-end RED) — EvaluateCeiling: a legacy-family latest
// verdict with every delta condition holding (fix_scope anchors, no STOP,
// the diff between the two audited SHAs touching only .moai/reports/
// paths, REQ/AC id sets unchanged) grants the delta round — a nil outcome.
// With the D1 defect prevSHA is empty, deltaOK is false, and the same
// fixture routes to the outcome ladder as a final hit (non-nil).
func TestEvaluateCeilingLegacyLatestDeltaGranted(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := auditreceipt.RunScrubbedGit(root, args...)
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(out)
	}
	specID := "SPEC-ACE-LEGD-001"
	specDir := filepath.Join(root, ".moai", "specs", specID)
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	reports := filepath.Join(root, ".moai", "reports", specID)
	for _, d := range []string{specDir, cfgDir, reports} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q")
	git("config", "user.email", "t@example.com")
	git("config", "user.name", "t")
	// tier M + ceiling 2 + auto_delta_rounds 1: count 2 reaches the ceiling
	// and grants the delta when the eligibility check verifies.
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("---\ntier: M\n---\n# spec\nREQ-ACR-001 AC-ACR-001\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	harness := "harness:\n  evaluator:\n    memory_scope: per_iteration\n  plan_audit_tier_ceilings:\n    S: 1\n    M: 2\n    L: 1\n  plan_audit_ceiling_policy:\n    auto_delta_rounds: 1\n    on_final_hit: hold-and-split\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "harness.yaml"), []byte(harness), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	base := git("rev-parse", "HEAD")

	// Round 1: legacy family, audited at the base commit.
	rev1Path := filepath.Join(reports, specID+"-review-1.md")
	if err := os.WriteFile(rev1Path, []byte("# review\nverdict: FAIL\naudited_sha: "+base+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "review-1")

	// Round 2 (latest): legacy family, fix_scope anchor, audited at a later
	// commit whose diff touches only .moai/reports/ paths. The audited_sha
	// line needs that commit's SHA, so the file is committed with a
	// placeholder and rewritten with the real one (the counter reads the
	// working tree; the diff machinery reads the committed trees).
	rev2Path := filepath.Join(reports, specID+"-review-2.md")
	body := "# SPEC Review Report: " + specID + "\nverdict: FAIL\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: stale\naudited_sha: PLACEHOLDER\nfix_scope: .moai/specs/" + specID + "/spec.md#REQ-ACR-001\n- debt: D1 dispose_in=run fixture debt\n"
	if err := os.WriteFile(rev2Path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "review-2")
	rev2 := git("rev-parse", "HEAD")
	if err := os.WriteFile(rev2Path, []byte(strings.Replace(body, "PLACEHOLDER", rev2, 1)), 0o600); err != nil {
		t.Fatal(err)
	}

	fields := auditverdict.Parse([]byte(strings.Replace(body, "PLACEHOLDER", rev2, 1)))
	in := VerdictCeilingInput{SpecID: specID, SpecDir: specDir, ProjectRoot: root}
	oc, override, err := EvaluateCeiling(in, fields, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc != nil {
		t.Fatalf("outcome %+v (override %v), want nil — the delta round is granted: count 2 reaches the tier ceiling 2 and the eligibility conditions hold", oc, override)
	}
}

// AC-ACR-005/006 (D2 RED) — persistOutcome: a debt-admit outcome's §G
// record line and its audit-trail line each carry the admitted debt
// inventory (the debt's ID and dispose_in recoverable), each a single
// line. With the defect both lines serialize kind/outcome/reasons/evidence
// only.
func TestPersistOutcomeDebtAdmitCarriesDebtInventory(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	f.appendLine(t, path, "- debt: D1 dispose_in=run fixture debt")
	fields, hashOK := f.fieldsFromHash(t, path)
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatal(err)
	}
	if oc == nil || oc.Outcome != OutcomeDebtAdmit || oc.Blocked || !override {
		t.Fatalf("outcome %+v override %v, want debt-admit admitting", oc, override)
	}
	progressRaw, err := os.ReadFile(filepath.Join(f.specDir, "progress.md"))
	if err != nil {
		t.Fatalf("progress.md not written: %v", err)
	}
	var gLine string
	for _, l := range strings.Split(string(progressRaw), "\n") {
		if strings.Contains(l, "outcome="+OutcomeDebtAdmit) {
			gLine = l
			break
		}
	}
	if gLine == "" {
		t.Fatalf("no debt-admit §G record line: %.400s", progressRaw)
	}
	for _, token := range []string{"debts=", `"id":"D1"`, `"dispose_in":"run"`} {
		if !strings.Contains(gLine, token) {
			t.Fatalf("§G record carries no %s token: %s", token, gLine)
		}
	}
	trailRaw, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log"))
	if err != nil {
		t.Fatalf("trail not written: %v", err)
	}
	var tLine string
	for _, l := range strings.Split(string(trailRaw), "\n") {
		if strings.Contains(l, "outcome="+OutcomeDebtAdmit) {
			tLine = l
			break
		}
	}
	if tLine == "" {
		t.Fatalf("no debt-admit trail line: %s", trailRaw)
	}
	for _, token := range []string{"debts=", `"id":"D1"`, `"dispose_in":"run"`} {
		if !strings.Contains(tLine, token) {
			t.Fatalf("trail line carries no %s token: %s", token, tLine)
		}
	}
}

// AC-ACR-013 (D3 RED) — appendProgressRecord: 24 concurrent records all
// survive exactly once, the §G heading is created exactly once, and no
// pre-existing line is lost. Judged over repeated runs with -race, never
// one green (concurrency discipline).
func TestAppendProgressRecordConcurrentSurvival(t *testing.T) {
	specDir := t.TempDir()
	pre := "# progress\n\n## §E.2 Run-phase Evidence\npre-existing line\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	const n = 24
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_ = appendProgressRecord(specDir, fmt.Sprintf("- record %02d outcome=x", i))
		}(i)
	}
	wg.Wait()
	raw, err := os.ReadFile(filepath.Join(specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	if !strings.Contains(content, "pre-existing line") {
		t.Fatalf("pre-existing progress content lost:\n%s", content)
	}
	if got := strings.Count(content, progressSectionHeading); got != 1 {
		t.Fatalf("§G heading appears %d times, want 1 (created exactly once under the race)", got)
	}
	for i := 0; i < n; i++ {
		want := fmt.Sprintf("- record %02d outcome=x", i)
		if got := strings.Count(content, want); got != 1 {
			t.Fatalf("record %q appears %d times, want 1\n%s", want, got, content)
		}
	}
}

// AC-ACR-016 (D3 §G insertion position RED) — appendProgressRecord: a
// record lands at the END of the §G block, immediately before the next
// same-level heading; §G last appends at end-of-file. With the defect the
// append writes to file end, so a following section absorbs the record.
func TestAppendProgressRecordInsertsAtSectionEnd(t *testing.T) {
	specDir := t.TempDir()
	pre := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n\n## §E.2 Run-phase Evidence\nlater section body\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendProgressRecord(specDir, "- new record"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	gIdx := strings.Index(content, progressSectionHeading)
	eIdx := strings.Index(content, "## §E.2 Run-phase Evidence")
	if gIdx < 0 || eIdx < 0 {
		t.Fatalf("fixture headings missing:\n%s", content)
	}
	if !strings.Contains(content[gIdx:eIdx], "- new record") {
		t.Fatalf("record did not land inside the §G block:\n%s", content)
	}
	if strings.Contains(content[eIdx:], "- new record") {
		t.Fatalf("record landed inside the later section:\n%s", content)
	}
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "## §E.2") {
			if lines[i-1] != "- new record" {
				t.Fatalf("the line immediately before the next same-level heading is %q, want the record", lines[i-1])
			}
		}
	}

	// §G last → end-of-file append.
	specDir2 := t.TempDir()
	pre2 := "# progress\n\n## §G Override and Refusal Record\n\n- old record\n"
	if err := os.WriteFile(filepath.Join(specDir2, "progress.md"), []byte(pre2), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendProgressRecord(specDir2, "- tail record"); err != nil {
		t.Fatal(err)
	}
	raw2, err := os.ReadFile(filepath.Join(specDir2, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(raw2), "- tail record\n") {
		t.Fatalf("record did not append at end-of-file when §G is last:\n%s", raw2)
	}
}

// AC-ACR-015 (delta-gate widening RED) — reqACSetsUnchanged derives the
// REQ/AC id sets from BOTH definition files: spec.md byte-identical at both
// audited SHAs while acceptance.md renames an AC id between them refuses
// the delta (false). With the defect the comparison reads only spec.md, so
// the acceptance-only rename verifies as unchanged (true).
func TestReqACSetsUnchangedReadsAcceptance(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := auditreceipt.RunScrubbedGit(root, args...)
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(out)
	}
	specID := "SPEC-ACC-GATE-001"
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	git("init", "-q")
	git("config", "user.email", "t@example.com")
	git("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# spec\nREQ-ACR-001\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "acceptance.md"), []byte("# acceptance\nAC-R-001 repro arm\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	first := git("rev-parse", "HEAD")

	// spec.md byte-identical; acceptance.md renames the AC id.
	if err := os.WriteFile(filepath.Join(specDir, "acceptance.md"), []byte("# acceptance\nAC-R-002 repro arm\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "rename")
	second := git("rev-parse", "HEAD")
	if reqACSetsUnchanged(root, specID, first, second) {
		t.Fatal("an acceptance-only AC id rename verified as unchanged — the delta must be refused (fail-closed)")
	}

	// Both definition files identical between SHAs verifies unchanged (the
	// spec.md-only contract is preserved).
	third := second
	if !reqACSetsUnchanged(root, specID, second, third) {
		t.Fatal("identical definition files verified as changed")
	}
}

// AC-ACR-007 (D2 escaping, security) — debt fields are verdict-supplied
// free text: double quotes, semicolons, and a raw newline in an admitted
// debt's ID or description never inject record lines into §G or the trail
// (both records stay single lines) and every field stays recoverable
// uncorrupted after escaping.
func TestPersistOutcomeDebtFieldsEscapedSingleLine(t *testing.T) {
	f := newCeilingFixture(t)
	oc := &VerdictCeilingOutcome{
		Outcome: OutcomeDebtAdmit,
		Debts: []auditverdict.Debt{
			{ID: `D"1;`, DisposeIn: "run", Description: "has \"quotes\" and; semicolons\nand a raw newline"},
			{ID: "D2", DisposeIn: "sync", Description: "plain second debt"},
		},
	}
	persistOutcome(f.input(), "ceiling-outcome", oc)

	progressRaw, err := os.ReadFile(filepath.Join(f.specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	var gLine string
	for _, l := range strings.Split(string(progressRaw), "\n") {
		if strings.Contains(l, "debts=") {
			gLine = l
			break
		}
	}
	if gLine == "" {
		t.Fatal("no §G record carries a debts field")
	}
	trailRaw, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log"))
	if err != nil {
		t.Fatal(err)
	}
	var tLine string
	for _, l := range strings.Split(string(trailRaw), "\n") {
		if strings.Contains(l, "debts=") {
			tLine = l
			break
		}
	}
	if tLine == "" {
		t.Fatal("no trail line carries a debts field")
	}

	for name, line := range map[string]string{"§G": gLine, "trail": tLine} {
		if strings.Contains(line, "\n") || strings.Contains(line, "\r") {
			t.Fatalf("%s record is not a single line: %q", name, line)
		}
		if !strings.Contains(line, `\n`) {
			t.Fatalf("%s record does not escape the raw newline: %q", name, line)
		}
		idx := strings.LastIndex(line, " debts=")
		if idx < 0 {
			t.Fatalf("%s record lost the debts field: %q", name, line)
		}
		var got []struct {
			ID          string `json:"id"`
			DisposeIn   string `json:"dispose_in"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal([]byte(line[idx+len(" debts="):]), &got); err != nil {
			t.Fatalf("%s debts field does not decode: %v — %q", name, err, line)
		}
		if len(got) != 2 {
			t.Fatalf("%s debts decode: %d entries, want 2", name, len(got))
		}
		if got[0].ID != `D"1;` || got[0].DisposeIn != "run" ||
			got[0].Description != "has \"quotes\" and; semicolons\nand a raw newline" {
			t.Fatalf("%s first debt corrupted: %+v", name, got[0])
		}
		if got[1].ID != "D2" || got[1].DisposeIn != "sync" || got[1].Description != "plain second debt" {
			t.Fatalf("%s second debt corrupted: %+v", name, got[1])
		}
	}
}

// AC-ACR-008 (D2 negative, RG) — outcomes with no debts keep the pre-repair
// record grammar with no debt field emitted, and the required-backend
// refusal / operator-override record shapes are untouched. Preserve test —
// green before and after the fix.
func TestPersistOutcomeNoDebtRecordUnchanged(t *testing.T) {
	f := newCeilingFixture(t)

	// pass-through, hold, and split outcomes through the engine.
	clean := f.writeIter(t, 1, "PASS")
	fields, hashOK := f.fieldsFromHash(t, clean)
	fields.Label = "PASS"
	oc, _, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil || oc == nil || oc.Outcome != OutcomePassThrough {
		t.Fatalf("pass-through fixture: %+v %v", oc, err)
	}
	path := f.writeIter(t, 2, "FAIL")
	f.appendLine(t, path, "must_pass_failed: 1") // a must-pass failure holds
	fields2, hashOK2 := f.fieldsFromHash(t, path)
	oc2, _, err := EvaluateCeiling(f.input(), fields2, hashOK2, nil)
	if err != nil || oc2 == nil || oc2.Outcome != OutcomeHold {
		t.Fatalf("hold fixture: %+v %v", oc2, err)
	}

	progressRaw, err := os.ReadFile(filepath.Join(f.specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(progressRaw), "debts=") {
		t.Fatalf("a no-debt record emitted a debts field:\n%s", progressRaw)
	}
	trailRaw, err := os.ReadFile(filepath.Join(f.root, ".moai", "state", "audit-enforcement.log"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(trailRaw), "debts=") {
		t.Fatalf("a no-debt trail line emitted a debts field:\n%s", trailRaw)
	}
	// The pre-repair grammar shape: one line per record, key=value fields.
	for _, l := range strings.Split(strings.TrimSpace(string(progressRaw)), "\n") {
		if !strings.HasPrefix(l, "- ") {
			continue
		}
		if !strings.Contains(l, " outcome=") || !strings.Contains(l, " reasons=") || !strings.Contains(l, " evidence=") {
			t.Fatalf("record line lost the pre-repair grammar: %s", l)
		}
	}

	// Required-backend refusal and override shapes are unchanged.
	RecordRequiredBackendRefusal(f.input(), "claude verdict unavailable")
	if err := AcknowledgeRequiredBackend(f.input(), "claude", "operator accepted the stale claude gate"); err != nil {
		t.Fatal(err)
	}
	progressRaw, err = os.ReadFile(filepath.Join(f.specDir, "progress.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(progressRaw)
	if !strings.Contains(content, "required-backend-refusal outcome=refused reasons=") ||
		strings.Contains(content, "debts=") {
		t.Fatalf("refusal record shape changed or grew a debts field:\n%s", content)
	}
	if !strings.Contains(content, "override backend=claude note=") {
		t.Fatalf("override record shape changed:\n%s", content)
	}
}

// AC-ACR-009 (best-effort invariant, RG) — a record-write failure never
// flips the admission decision: a debt-admit evaluation against an
// unwritable SPEC directory still returns the debt-admit outcome with
// override=true and no error. Preserve test — green before and after the
// fix.
func TestPersistOutcomeRecordFailureKeepsAdmission(t *testing.T) {
	f := newCeilingFixture(t)
	path := f.writeIter(t, 1, "FAIL")
	f.appendLine(t, path, "- debt: D1 dispose_in=run fixture debt")
	fields, hashOK := f.fieldsFromHash(t, path)
	if err := os.Chmod(f.specDir, 0o555); err != nil {
		t.Skipf("cannot make the SPEC directory unwritable here: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(f.specDir, 0o755) })
	oc, override, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err != nil {
		t.Fatalf("a record-write failure must not error the evaluation: %v", err)
	}
	if oc == nil || oc.Outcome != OutcomeDebtAdmit || oc.Blocked || !override {
		t.Fatalf("outcome %+v override %v, want debt-admit admitting despite the record failure", oc, override)
	}
}
