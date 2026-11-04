package template

import (
	"io/fs"
	"strings"
	"testing"
)

// Card t2008 / RUN-02. The run workflow body described the plan-artifact hash
// as whitespace-normalized and the cache lookup as a read of today's daily
// report. The implementation hashes exact bytes (internal/runtime/audit_cache.go
// ComputeHash) and reads the latest plan-phase review-N, deliberately not the
// date-stamped report (internal/runtime/audit_review.go FileAuditCache). The
// instruction consumer and the Go helper must select the same cache entry.

// planCacheDescriptionViolations lists what a phase-execution.md body gets
// wrong about the plan hash and the cache source. Empty means consistent.
func planCacheDescriptionViolations(text string) []string {
	var v []string
	check := func(label, body string, ok bool, must, mustNot []string) {
		if !ok {
			v = append(v, label+": section not found")
			return
		}
		for _, w := range must {
			if !strings.Contains(body, w) {
				v = append(v, label+": missing "+w)
			}
		}
		for _, w := range mustNot {
			if strings.Contains(body, w) {
				v = append(v, label+": still says "+w)
			}
		}
	}

	hash, ok := section(text, "### Step 1: Compute Plan Artifact Hash", "### Step 2:")
	check("Step 1 hash", hash, ok,
		[]string{"exact bytes", "ComputeHash"},
		[]string{"whitespace-normalized"})

	cache, ok := section(text, "### Step 2: Consult", "### Step 3:")
	check("Step 2 cache", cache, ok,
		[]string{"ResolveLatestPlanAudit", "review-", "history", "gate receipt"},
		[]string{"(today's date)"})

	return v
}

// TestRunPlanCacheDescription_Deployed checks the deployed template body.
func TestRunPlanCacheDescription_Deployed(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, phaseExecutionPath)
	if err != nil {
		t.Fatalf("read %s: %v", phaseExecutionPath, err)
	}
	if v := planCacheDescriptionViolations(string(data)); len(v) > 0 {
		t.Errorf("%s misdescribes the plan hash or cache source:\n  %s", phaseExecutionPath, strings.Join(v, "\n  "))
	}
}

// TestRunPlanCacheDescription_Mutants proves the checker can fail in both
// directions: the stale wording is rejected, a body with only the new words
// but the stale phrase kept is rejected, and each required word is load-bearing.
func TestRunPlanCacheDescription_Mutants(t *testing.T) {
	t.Parallel()

	const stale = "### Step 1: Compute Plan Artifact Hash\nhashes the whitespace-normalized content\n" +
		"### Step 2: Consult the Cache\nRead the report (today's date).\n### Step 3: Invoke\n"
	if v := planCacheDescriptionViolations(stale); len(v) < 5 {
		t.Fatalf("stale body must fail on every group, got %d violations: %v", len(v), v)
	}

	const good = "### Step 1: Compute Plan Artifact Hash\nComputeHash hashes the exact bytes.\n" +
		"### Step 2: Consult the Cache\nResolveLatestPlanAudit reads the latest review-N; the daily report is history; the gate receipt is separate.\n" +
		"### Step 3: Invoke\n"
	if v := planCacheDescriptionViolations(good); len(v) != 0 {
		t.Fatalf("positive control must pass, got: %v", v)
	}

	// Stale phrase kept next to the new words must still fail.
	both := strings.Replace(good, "hashes the exact bytes.", "hashes the exact bytes, whitespace-normalized.", 1)
	if v := planCacheDescriptionViolations(both); len(v) == 0 {
		t.Error("keeping the stale 'whitespace-normalized' phrase must be detected")
	}
	for _, drop := range []string{"exact bytes", "ResolveLatestPlanAudit", "review-", "history", "gate receipt"} {
		mutant := strings.ReplaceAll(good, drop, "x")
		if v := planCacheDescriptionViolations(mutant); len(v) == 0 {
			t.Errorf("dropping %q must be detected", drop)
		}
	}
}
