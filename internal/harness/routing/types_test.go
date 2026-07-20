package routing

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func strptr(s string) *string { return &s }

// TestRequestDigestNoVerbatim asserts the privacy contract (REQ-HEV-005,
// AC-HEV-009): the digest matches sha256:[0-9a-f]{12} and no verbatim request
// text is derivable from it.
func TestRequestDigestNoVerbatim(t *testing.T) {
	t.Parallel()

	const sentinel = "PLEASE-REMEMBER-MY-SECRET-PASSWORD-hunter2"
	digest := RequestDigest(sentinel)

	re := regexp.MustCompile(`^sha256:[0-9a-f]{12}$`)
	if !re.MatchString(digest) {
		t.Fatalf("digest %q does not match sha256:[0-9a-f]{12}", digest)
	}
	if strings.Contains(digest, "hunter2") || strings.Contains(digest, "SECRET") {
		t.Fatalf("digest leaked verbatim request text: %q", digest)
	}

	// Whitespace/case normalization -> identical digest.
	if RequestDigest("  Fix   the   BUG  ") != RequestDigest("fix the bug") {
		t.Fatal("digest not normalization-invariant")
	}
	// Different requests -> different digests.
	if RequestDigest("alpha") == RequestDigest("beta") {
		t.Fatal("distinct requests collided in the truncated digest")
	}
}

// TestClassifyRequest covers the coarse, deterministic classifier (REQ-HEV-005).
func TestClassifyRequest(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"add a new OAuth login feature": "feature",
		"implement the routing ledger":  "feature",
		"fix the broken parser":         "bugfix",
		"there is a regression in CI":   "bugfix",
		"refactor the writer package":   "refactor",
		"clean up the dead code":        "refactor",
		"update the README docs":        "docs",
		"why does the hook fail?":       "question",
		"run the /moai run pipeline":    "pipeline",
		"the sky is blue and green":     "other",
	}
	for req, want := range cases {
		if got := ClassifyRequest(req); got != want {
			t.Errorf("ClassifyRequest(%q) = %q, want %q", req, got, want)
		}
	}
}

// TestValidEvidenceKind pins the closed evidence-kind enum (REQ-HEV-006).
func TestValidEvidenceKind(t *testing.T) {
	t.Parallel()
	for _, k := range []EvidenceKind{KindGateExit, KindAuditScore, KindVerifyPath, KindAbort} {
		if !ValidEvidenceKind(k) {
			t.Errorf("kind %q should be valid", k)
		}
	}
	for _, k := range []EvidenceKind{"", "success", "prose", "outcome"} {
		if ValidEvidenceKind(k) {
			t.Errorf("kind %q should be rejected (free-text outcome would be un-fakeable violation)", k)
		}
	}
}

// TestConvergenceNullWhenNoSignal asserts evidence-or-null semantics
// (REQ-HEV-003/006, AC-HEV-008): absent convergence signals serialize as null,
// never an inferred value; empty delegations/evidence serialize as [].
func TestConvergenceNullWhenNoSignal(t *testing.T) {
	t.Parallel()

	// No convergence signal -> nil pointers -> JSON null.
	p := PendingRow{
		SchemaVersion:     SchemaVersion,
		MatchedSubcommand: "run",
		RequestDigest:     "sha256:abcdef012345",
		RequestClass:      "feature",
	}
	row := p.Finalize(OutcomeSuccess)

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	js := string(data)
	for _, want := range []string{
		`"goal_converged":null`,
		`"convergence_class":null`,
		`"mode_selected":null`,
		`"tier":null`,
		`"harness_level":null`,
		`"delegations":[]`,
		`"evidence_refs":[]`,
		`"schema_version":1`,
		`"outcome":"success"`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("serialized row missing %q\n got: %s", want, js)
		}
	}

	// Present convergence signal -> concrete value, not null.
	conv := true
	cc := "converged"
	p2 := PendingRow{GoalConverged: &conv, ConvergenceClass: &cc, Tier: strptr("M")}
	d2, _ := json.Marshal(p2.Finalize(OutcomeSuccess))
	if !strings.Contains(string(d2), `"goal_converged":true`) ||
		!strings.Contains(string(d2), `"convergence_class":"converged"`) ||
		!strings.Contains(string(d2), `"tier":"M"`) {
		t.Errorf("present convergence signal not serialized as concrete value: %s", d2)
	}
}
