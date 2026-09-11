package spec

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// buildReport is a fixture helper: a Report carrying exactly the findings given.
func buildReport(findings ...Finding) *Report {
	return &Report{Findings: findings}
}

func warn(code string, advisory bool) Finding {
	return Finding{
		File:     "x.md",
		Line:     1,
		Severity: SeverityWarning,
		Code:     code,
		Advisory: advisory,
	}
}

func errFinding(code string) Finding {
	return Finding{File: "x.md", Line: 1, Severity: SeverityError, Code: code}
}

// AC-SLGS-005 / REQ-SLGS-004 — the gating dimension is per-rule NON-ADVISORY
// warning counts. Advisory findings never enter the baseline (they are t518's
// population and would guarantee immediate rebaseline churn).
func TestComputeBaselineRules_CountsNonAdvisoryWarningsOnly(t *testing.T) {
	r := buildReport(
		warn("Alpha", false),
		warn("Alpha", false),
		warn("Alpha", true), // advisory — excluded
		warn("Beta", false),
		warn("Gamma", true), // advisory-only rule — absent from baseline entirely
		errFinding("Delta"), // error — not a warning, excluded
	)

	got := ComputeBaselineRules(r)

	want := map[string]int{"Alpha": 2, "Beta": 1}
	if len(got) != len(want) {
		t.Fatalf("rule count: got %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("rule %q: got %d, want %d", k, got[k], v)
		}
	}
	if _, ok := got["Gamma"]; ok {
		t.Errorf("advisory-only rule Gamma must not enter the baseline, got %v", got)
	}
}

// §D.6 determinism — the serialized bytes must not depend on map insertion order.
func TestMarshalBaseline_DeterministicAcrossMapInsertionOrder(t *testing.T) {
	a := &Baseline{Version: 1, UpdatedAt: "2026-09-08", TreeSHA: "abc1234", Reason: "r"}
	a.Rules = map[string]int{}
	for _, k := range []string{"Zeta", "Alpha", "Mu"} {
		a.Rules[k] = len(k)
	}

	b := &Baseline{Version: 1, UpdatedAt: "2026-09-08", TreeSHA: "abc1234", Reason: "r"}
	b.Rules = map[string]int{}
	for _, k := range []string{"Mu", "Zeta", "Alpha"} {
		b.Rules[k] = len(k)
	}

	ba, err := MarshalBaseline(a)
	if err != nil {
		t.Fatalf("MarshalBaseline(a): %v", err)
	}
	bb, err := MarshalBaseline(b)
	if err != nil {
		t.Fatalf("MarshalBaseline(b): %v", err)
	}
	// Guard against a vacuous pass: two empty outputs are trivially "equal".
	// Observed during RED — the stub returned (nil, nil) and this test went
	// green while asserting nothing.
	if len(ba) == 0 {
		t.Fatalf("MarshalBaseline produced no bytes — equality below would assert nothing")
	}
	if !bytes.Equal(ba, bb) {
		t.Fatalf("serialization is insertion-order dependent:\nA=%s\nB=%s", ba, bb)
	}

	// Same input twice must also be byte-identical (no wall-clock, no nonce).
	ba2, _ := MarshalBaseline(a)
	if !bytes.Equal(ba, ba2) {
		t.Fatalf("serialization is not idempotent:\n1=%s\n2=%s", ba, ba2)
	}
}

func TestMarshalBaseline_SortedKeysIndentAndTrailingNewline(t *testing.T) {
	b := &Baseline{
		Version:   1,
		UpdatedAt: "2026-09-08",
		TreeSHA:   "deadbee",
		Reason:    "initial",
		Rules:     map[string]int{"Zeta": 1, "Alpha": 2},
	}
	out, err := MarshalBaseline(b)
	if err != nil {
		t.Fatalf("MarshalBaseline: %v", err)
	}
	if !bytes.HasSuffix(out, []byte("\n")) {
		t.Errorf("missing trailing newline: %q", out)
	}
	if !bytes.Contains(out, []byte("\n  \"version\"")) {
		t.Errorf("expected 2-space indent, got:\n%s", out)
	}
	ai := bytes.Index(out, []byte(`"Alpha"`))
	zi := bytes.Index(out, []byte(`"Zeta"`))
	if ai < 0 || zi < 0 || ai > zi {
		t.Errorf("rule keys not sorted (Alpha at %d, Zeta at %d):\n%s", ai, zi, out)
	}
	// It must round-trip through the standard decoder.
	var back Baseline
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("round-trip decode: %v", err)
	}
	if back.Rules["Alpha"] != 2 || back.TreeSHA != "deadbee" {
		t.Errorf("round-trip lost data: %+v", back)
	}
}

// REQ-SLGS-006 — a rule above its recorded count is an increase, and the delta
// names the rule. Boundary: a rule ABSENT from the baseline has recorded 0, so
// any occurrence is an increase (the "new rule fires for the first time" case).
func TestCompareBaseline_IncreaseNamesTheRule(t *testing.T) {
	base := &Baseline{Rules: map[string]int{"Alpha": 1}}
	r := buildReport(
		warn("Alpha", false),
		warn("Alpha", false), // 1 -> 2
		warn("Brandnew", false),
	)

	cmp := CompareBaseline(r, base)

	if !cmp.Failed() {
		t.Fatalf("expected failure on increase, got %+v", cmp)
	}
	if len(cmp.Increases) != 2 {
		t.Fatalf("want 2 increases, got %d: %+v", len(cmp.Increases), cmp.Increases)
	}
	byCode := map[string]RuleDelta{}
	for _, d := range cmp.Increases {
		byCode[d.Code] = d
	}
	if d := byCode["Alpha"]; d.Recorded != 1 || d.Current != 2 {
		t.Errorf("Alpha delta: got recorded=%d current=%d, want 1/2", d.Recorded, d.Current)
	}
	if d, ok := byCode["Brandnew"]; !ok || d.Recorded != 0 || d.Current != 1 {
		t.Errorf("absent-from-baseline rule must read recorded=0 current=1, got %+v (present=%v)", d, ok)
	}
	// Increases must be deterministically ordered for stable output.
	if cmp.Increases[0].Code > cmp.Increases[1].Code {
		t.Errorf("increases not sorted by code: %+v", cmp.Increases)
	}
}

// REQ-SLGS-007 — decreases pass and are reported.
func TestCompareBaseline_DecreasePassesAndIsReported(t *testing.T) {
	base := &Baseline{Rules: map[string]int{"Alpha": 3, "Gone": 2}}
	r := buildReport(warn("Alpha", false))

	cmp := CompareBaseline(r, base)

	if cmp.Failed() {
		t.Fatalf("decrease must pass, got %+v", cmp)
	}
	if len(cmp.Decreases) != 2 {
		t.Fatalf("want 2 decreases (Alpha 3->1, Gone 2->0), got %+v", cmp.Decreases)
	}
}

// REQ-SLGS-009 — an error-severity finding fails regardless of the baseline,
// and the comparison reports it as the error cause, not as a baseline increase.
func TestCompareBaseline_ErrorFailsRegardlessOfBaseline(t *testing.T) {
	base := &Baseline{Rules: map[string]int{"Alpha": 5}}
	r := buildReport(warn("Alpha", false), errFinding("Boom"))

	cmp := CompareBaseline(r, base)

	if cmp.ErrorCount != 1 {
		t.Errorf("ErrorCount: got %d, want 1", cmp.ErrorCount)
	}
	if len(cmp.Increases) != 0 {
		t.Errorf("no increase expected (Alpha 5 -> 1), got %+v", cmp.Increases)
	}
	if !cmp.Failed() {
		t.Fatalf("error must fail the gate even with a passing baseline: %+v", cmp)
	}
	if !cmp.ErrorGated() {
		t.Fatalf("the failure cause must be reported as error-gated, not baseline-exceeded: %+v", cmp)
	}
}

// REQ-SLGS-004 — the standing inventory stays observable while green: the total
// warning count includes advisory findings.
func TestCompareBaseline_TotalWarningsIncludesAdvisory(t *testing.T) {
	base := &Baseline{Rules: map[string]int{"Alpha": 1}}
	r := buildReport(warn("Alpha", false), warn("Adv", true), warn("Adv2", true))

	cmp := CompareBaseline(r, base)

	if cmp.TotalWarnings != 3 {
		t.Errorf("TotalWarnings: got %d, want 3 (advisory included)", cmp.TotalWarnings)
	}
	if cmp.Failed() {
		t.Errorf("advisory findings must not gate: %+v", cmp)
	}
}

func TestLoadBaseline_MissingFileIsDistinguishable(t *testing.T) {
	_, err := LoadBaseline(filepath.Join(t.TempDir(), "nope.json"))
	if err == nil {
		t.Fatal("expected an error for a missing baseline file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("missing-file error must satisfy os.IsNotExist so the caller can name the path; got %T: %v", err, err)
	}
}

func TestWriteAndLoadBaseline_RoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "baseline.json")
	in := &Baseline{
		Version:   BaselineSchemaVersion,
		UpdatedAt: "2026-09-08",
		TreeSHA:   "cafebab",
		Reason:    "initial capture",
		Rules:     map[string]int{"Alpha": 2},
	}
	if err := WriteBaseline(p, in); err != nil {
		t.Fatalf("WriteBaseline: %v", err)
	}
	out, err := LoadBaseline(p)
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	if out.Reason != in.Reason || out.TreeSHA != in.TreeSHA || out.Rules["Alpha"] != 2 {
		t.Errorf("round-trip mismatch: got %+v want %+v", out, in)
	}
}
