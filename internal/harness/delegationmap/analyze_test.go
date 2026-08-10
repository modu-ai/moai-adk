package delegationmap

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fixture returns the path of a committed testdata fixture.
func fixture(name string) string { return filepath.Join("testdata", name) }

// opts builds Options against a committed ledger fixture and the fixture map.
func opts(ledger string) Options {
	return Options{
		LedgerPath: fixture(ledger),
		MapPath:    fixture("delegation_map.yaml"),
	}
}

// statFor returns the aggregate for one subcommand, failing when absent.
func statFor(t *testing.T, res Result, subcommand string) SubcommandStat {
	t.Helper()
	for _, s := range res.Stats {
		if s.Subcommand == subcommand {
			return s
		}
	}
	t.Fatalf("no stat for subcommand %q (stats: %+v)", subcommand, res.Stats)
	return SubcommandStat{}
}

// findingsOfKind filters findings by kind.
func findingsOfKind(res Result, k Kind) []Finding {
	var out []Finding
	for _, f := range res.Findings {
		if f.Kind == k {
			out = append(out, f)
		}
	}
	return out
}

// TestAggregate_Basic is AC-HLA-003.
func TestAggregate_Basic(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("aggregate_basic.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	plan := statFor(t, res, "plan")
	if plan.QualifyingRows != 8 {
		t.Errorf("qualifying rows = %d, want 8", plan.QualifyingRows)
	}
	if plan.AgentCounts["manager-spec"] != 8 {
		t.Errorf("manager-spec count = %d, want 8", plan.AgentCounts["manager-spec"])
	}
	if plan.AgentCounts["Explore"] != 3 {
		t.Errorf("Explore count = %d, want 3", plan.AgentCounts["Explore"])
	}
}

// TestAnalyze_RerouteAbortExcluded is AC-HLA-004.
func TestAnalyze_RerouteAbortExcluded(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("reroute_abort_only.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(res.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(res.Findings))
	}
	plan := statFor(t, res, "plan")
	if plan.QualifyingRows != 0 {
		t.Errorf("qualifying rows = %d, want 0", plan.QualifyingRows)
	}
	if plan.RerouteRows != 3 || plan.AbortRows != 3 {
		t.Errorf("reroute/abort = %d/%d, want 3/3", plan.RerouteRows, plan.AbortRows)
	}
	// A reroute/abort row's delegations must not leak into the per-agent counts.
	if plan.AgentCounts["manager-spec"] != 0 {
		t.Errorf("non-qualifying rows leaked into agent counts: %v", plan.AgentCounts)
	}
}

// TestAnalyze_OversizedLedgerRefused is AC-HLA-002 (the runtime half).
//
// The bound is lowered through Options rather than committing a 32 MiB fixture:
// the guard being demonstrated is the size comparison, and a multi-megabyte blob
// in the repository would cost far more than it proves.
func TestAnalyze_OversizedLedgerRefused(t *testing.T) {
	t.Parallel()

	o := opts("oversized.jsonl")
	o.MaxLedgerBytes = 16 // any real fixture exceeds this
	res, err := Analyze(o)
	if err != nil {
		t.Fatalf("an oversized ledger must not be an error: %v", err)
	}
	if len(res.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(res.Findings))
	}
	if res.Reason != ReasonLedgerOversize {
		t.Errorf("reason = %q, want %q", res.Reason, ReasonLedgerOversize)
	}
}

// TestAnalyze_NonCatalogNeverUndesignated is AC-HLA-005.
func TestAnalyze_NonCatalogNeverUndesignated(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("non_catalog_agents.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	nonCatalog := []string{
		"general-purpose",
		"workflow-subagent",
		"hns-oss-docs-locale-translator-specialist",
		"audit-hle",
	}
	for _, f := range findingsOfKind(res, KindUndesignatedAgent) {
		for _, nc := range nonCatalog {
			if f.Agent == nc {
				t.Errorf("non-catalog value %q produced an undesignated_agent finding", nc)
			}
		}
	}

	run := statFor(t, res, "run")
	for _, nc := range nonCatalog {
		var seen bool
		for _, got := range run.NonCatalogAgents {
			if got == nc {
				seen = true
			}
		}
		if !seen {
			t.Errorf("%q was not recorded as non-catalog: %v", nc, run.NonCatalogAgents)
		}
		if run.AgentCounts[nc] == 0 {
			t.Errorf("%q was dropped rather than classified", nc)
		}
	}

	// The suppression must be by classification, not by an empty result: a
	// retained-catalog agent that qualifies still produces its finding.
	var sawCatalogFinding bool
	for _, f := range findingsOfKind(res, KindUndesignatedAgent) {
		if f.Agent == "manager-docs" && f.Subcommand == "run" {
			sawCatalogFinding = true
		}
	}
	if !sawCatalogFinding {
		t.Fatalf("expected an undesignated_agent finding for manager-docs on run; got %+v", res.Findings)
	}
}

// TestAnalyze_UnattributedShareReported is AC-HLA-006.
func TestAnalyze_UnattributedShareReported(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("unattributed_share.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	plan := statFor(t, res, "plan")
	if plan.UnattributedEntries != 4 {
		t.Errorf("unattributed entries = %d, want 4", plan.UnattributedEntries)
	}
	if _, ok := plan.AgentCounts[unattributedMarker]; ok {
		t.Errorf("the unattributed marker leaked into the per-agent counts: %v", plan.AgentCounts)
	}
	for _, ncAgent := range plan.NonCatalogAgents {
		if ncAgent == unattributedMarker {
			t.Errorf("the unattributed marker was classified as a non-catalog agent")
		}
	}

	var planFindings int
	for _, f := range res.Findings {
		if f.Agent == unattributedMarker {
			t.Errorf("the unattributed marker appeared as a finding agent")
		}
		if f.Subcommand == "plan" {
			planFindings++
			if f.UnattributedShare != 4 {
				t.Errorf("finding %+v carries unattributed share %d, want 4", f, f.UnattributedShare)
			}
		}
	}
	if planFindings == 0 {
		t.Fatalf("expected at least one plan finding to carry the share; got %+v", res.Findings)
	}
}

// TestAnalyze_BelowMinRows is the first half of AC-HLA-007.
func TestAnalyze_BelowMinRows(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("below_min_rows.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(res.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (4 rows < MinQualifyingRows %d)", len(res.Findings), MinQualifyingRows)
	}
	if res.Reason != ReasonBelowMinRows {
		t.Errorf("reason = %q, want %q", res.Reason, ReasonBelowMinRows)
	}
}

// TestAnalyze_StrayRowSuppressed is the second half of AC-HLA-007.
func TestAnalyze_StrayRowSuppressed(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("stray_agent.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	run := statFor(t, res, "run")
	if run.AgentCounts["manager-docs"] != 1 {
		t.Fatalf("fixture precondition: manager-docs count = %d, want 1", run.AgentCounts["manager-docs"])
	}
	for _, f := range res.Findings {
		if f.Agent == "manager-docs" {
			t.Errorf("a 0.10-support agent produced a finding: %+v", f)
		}
	}
}

// TestAnalyze_TwoKindsOnly is AC-HLA-008.
func TestAnalyze_TwoKindsOnly(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("two_kinds.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(res.Findings) != 2 {
		t.Fatalf("findings = %d, want exactly 2: %+v", len(res.Findings), res.Findings)
	}
	if got := findingsOfKind(res, KindUndesignatedAgent); len(got) != 1 || got[0].Subcommand != "run" || got[0].Agent != "manager-docs" {
		t.Errorf("undesignated finding = %+v, want one for run/manager-docs", got)
	}
	if got := findingsOfKind(res, KindDesignatedNeverSpawned); len(got) != 1 || got[0].Subcommand != "sync" || got[0].Agent != "manager-docs" {
		t.Errorf("never-spawned finding = %+v, want one for sync/manager-docs", got)
	}
	// The enum admits no third value.
	for _, k := range []Kind{"skill_proposal", "reroute_pattern", ""} {
		if ValidKind(k) {
			t.Errorf("ValidKind admitted %q", k)
		}
	}
	for _, f := range res.Findings {
		if !ValidKind(f.Kind) {
			t.Errorf("finding carries an invalid kind: %+v", f)
		}
	}
}

// TestAnalyze_ConditionalDesignationsExcluded is AC-HLA-009.
func TestAnalyze_ConditionalDesignationsExcluded(t *testing.T) {
	t.Parallel()

	res, err := Analyze(opts("conditional_designations.jsonl"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	for _, f := range findingsOfKind(res, KindDesignatedNeverSpawned) {
		if f.Agent == "sync-auditor" || f.Agent == "plan-auditor" {
			t.Errorf("a conditionally-invoked designation produced a finding: %+v", f)
		}
	}

	// Every exclusion entry carries a rule citation — without one the carve-out
	// is an allowlist that suppresses real findings.
	if len(ConditionalExclusions) == 0 {
		t.Fatal("the exclusion set is empty")
	}
	for _, e := range ConditionalExclusions {
		if e.Agent == "" || e.RuleCitation == "" {
			t.Errorf("exclusion entry %+v is missing an agent or a rule citation", e)
		}
	}

	// A third designated agent, absent from the exclusion set and observed zero
	// times in the same fixture, still produces its finding.
	var sawExplore bool
	for _, f := range findingsOfKind(res, KindDesignatedNeverSpawned) {
		if f.Agent == "Explore" && f.Subcommand == "plan" {
			sawExplore = true
		}
	}
	if !sawExplore {
		t.Fatalf("expected a designated_never_spawned finding for plan/Explore; got %+v", res.Findings)
	}
}

// TestMapReader_ReadOnly is AC-HLA-010.
func TestMapReader_ReadOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mapPath := filepath.Join(dir, "delegation.yaml")
	src, err := os.ReadFile(fixture("delegation_map.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mapPath, src, 0o444); err != nil { // read-only
		t.Fatal(err)
	}
	before, err := os.Stat(mapPath)
	if err != nil {
		t.Fatal(err)
	}

	designations, err := ReadDelegationMap(mapPath)
	if err != nil {
		t.Fatalf("ReadDelegationMap: %v", err)
	}
	if got := designations["sync"]; len(got) != 2 || got[0] != "manager-docs" || got[1] != "sync-auditor" {
		t.Errorf("sync designations = %v, want [manager-docs sync-auditor]", got)
	}
	if got := designations["run"]; len(got) != 1 || got[0] != "manager-develop" {
		t.Errorf("run designations = %v, want [manager-develop]", got)
	}

	// A full analyzer run against the read-only map must succeed and leave the
	// file untouched, bytes and mtime alike.
	res, err := Analyze(Options{LedgerPath: fixture("two_kinds.jsonl"), MapPath: mapPath})
	if err != nil {
		t.Fatalf("Analyze against a read-only map: %v", err)
	}
	if len(res.Findings) == 0 {
		t.Error("expected findings from the two_kinds fixture")
	}

	after, err := os.Stat(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("the delegation map's mtime changed: %v -> %v", before.ModTime(), after.ModTime())
	}
	got, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(src) {
		t.Error("the delegation map's bytes changed")
	}
}

// TestAnalyze_GracefulNoOp is AC-HLA-014.
func TestAnalyze_GracefulNoOp(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		ledger         string
		wantReason     string
		wantMalformed  int
		absoluteAbsent bool
	}{
		{name: "absent", ledger: "does-not-exist.jsonl", wantReason: ReasonLedgerAbsent, absoluteAbsent: true},
		{name: "empty", ledger: "empty.jsonl", wantReason: ReasonLedgerEmpty},
		{name: "all malformed", ledger: "all_malformed.jsonl", wantReason: ReasonAllLinesMalformed, wantMalformed: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := Analyze(opts(tc.ledger))
			if err != nil {
				t.Fatalf("Analyze must not error on a degraded input: %v", err)
			}
			if len(res.Findings) != 0 {
				t.Errorf("findings = %d, want 0", len(res.Findings))
			}
			if res.Findings == nil {
				t.Error("Findings must be an empty slice, not nil, so it serializes as []")
			}
			if res.Reason != tc.wantReason {
				t.Errorf("reason = %q, want %q", res.Reason, tc.wantReason)
			}
			if res.Reason == "" {
				t.Error("a degraded result must carry a machine-readable reason")
			}
			if res.MalformedLines != tc.wantMalformed {
				t.Errorf("malformed lines = %d, want %d", res.MalformedLines, tc.wantMalformed)
			}
		})
	}
}

// TestAnalyze_NoSkillProposals is the skill-proposal half of AC-HLA-016.
//
// The schema-v1 ledger row records no injected-skill data, so a skill amendment
// could not be grounded in observation. The check is structural — the Finding
// type must declare no skill field at all — because a declared-but-empty field
// is exactly what would invite the ungrounded proposal later.
func TestAnalyze_NoSkillProposals(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"two_kinds.jsonl", "unattributed_share.jsonl", "conditional_designations.jsonl"} {
		res, err := Analyze(opts(name))
		if err != nil {
			t.Fatalf("Analyze(%s): %v", name, err)
		}
		for _, f := range res.Findings {
			if f.Kind != KindUndesignatedAgent && f.Kind != KindDesignatedNeverSpawned {
				t.Errorf("%s produced a finding outside the two-kind enum: %+v", name, f)
			}
		}
	}
	assertNoSkillField(t)
}

// assertNoSkillField is the structural half of the skill-proposal prohibition:
// the Finding type must declare no skill-shaped field.
func assertNoSkillField(t *testing.T) {
	t.Helper()
	ft := reflect.TypeOf(Finding{})
	for i := range ft.NumField() {
		name := strings.ToLower(ft.Field(i).Name)
		if strings.Contains(name, "skill") {
			t.Errorf("Finding declares a skill-shaped field %q; the schema-v1 ledger records no skill data", ft.Field(i).Name)
		}
	}
}
