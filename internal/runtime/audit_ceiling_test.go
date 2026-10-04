package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		{"", 3},     // absent tier → L
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

// Verdict bodies the ceiling fixtures record into their latest iteration file.
const (
	ceilingPassBody = "verdict: PASS\noverall_score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: abc123\n"
	ceilingDebtBody = "verdict: PASS-WITH-DEBT\noverall_score: 0.88\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: abc123\n" +
		"- debt: D1 dispose_in=run plan row omits an edit\n" +
		"- debt: D2 dispose_in=sync trace is indirect\n"
	ceilingFailBody = "verdict: FAIL\noverall_score: 0.70\nmust_pass_failed: 2\nblocking_count: 1\nplan_artifact_hash: abc123\n"
)

// ceilingFixture writes an evidence directory whose round count (3) sits at or
// above every shipped ceiling and whose latest iteration carries body.
func ceilingFixture(t *testing.T, body string) string {
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
	t.Chdir(project)

	outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-004", ceilingFixture(t, ceilingFailBody), "hold-and-split"))
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

	if err := RecordCeilingOutcome("SPEC-CEIL-004", outcome); err != nil {
		t.Fatalf("RecordCeilingOutcome: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(".moai", "state", "audit-ceiling", "SPEC-CEIL-004.json"))
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
	if err := RecordCeilingOutcome("SPEC-CEIL-004", outcome); err != nil {
		t.Fatalf("re-record: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(".moai", "state", "audit-ceiling"))
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

	outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-009", ceilingFixture(t, ceilingDebtBody), "hold-and-split"))
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
		outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-010", ceilingFixture(t, ceilingFailBody), policy))
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

	outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-012", ceilingFixture(t, ceilingFailBody), "split"))
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
		t.Chdir(project)

		outcome, hit, err := EvaluatePlanAuditCeiling(ceilingInput("SPEC-CEIL-005", ceilingFixture(t, ceilingPassBody), "hold-and-split"))
		if err != nil {
			t.Fatalf("EvaluatePlanAuditCeiling: %v", err)
		}
		if hit {
			t.Fatalf("a clean admitted PASS at the ceiling must not apply the ceiling, got %+v", outcome)
		}
		// The composition the CLI verb runs: record only on hit.
		if hit {
			if err := RecordCeilingOutcome("SPEC-CEIL-005", outcome); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := os.Stat(filepath.Join(".moai", "state", "audit-ceiling", "SPEC-CEIL-005.json")); !os.IsNotExist(err) {
			t.Errorf("record exists for a clean PASS at the ceiling (stat err = %v)", err)
		}
	})

	t.Run("invalid resolved ceiling propagates the error with no record", func(t *testing.T) {
		project := t.TempDir()
		t.Chdir(project)

		in := ceilingInput("SPEC-CEIL-005B", ceilingFixture(t, ceilingFailBody), "hold-and-split")
		in.Tier = "M"
		in.Ceilings = map[string]int{"S": 1, "L": 3} // M missing → configuration error
		_, hit, err := EvaluatePlanAuditCeiling(in)
		if err == nil {
			t.Fatal("a missing resolved ceiling must propagate the configuration error")
		}
		if hit {
			t.Error("no ceiling outcome applies when the ceiling never resolved")
		}
		if _, statErr := os.Stat(filepath.Join(".moai", "state", "audit-ceiling", "SPEC-CEIL-005B.json")); !os.IsNotExist(statErr) {
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
