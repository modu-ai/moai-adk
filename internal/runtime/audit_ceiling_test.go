package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// input is the CeilingInput for the fixture.
func (f *ceilingFixture) input() CeilingInput {
	return CeilingInput{SpecID: f.specID, SpecDir: f.specDir, ProjectRoot: f.root, CardID: f.cardID}
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
