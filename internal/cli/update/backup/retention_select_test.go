package backup

// retention_select_test.go — SPEC-UPDATE-BACKUP-RETENTION-001, selection step
// (design D2). The selection is a pure function of a census, so the budget, the
// current-version rule, the cover and the exempt-run disposition are asserted
// here without a filesystem. AC-UBR-02, AC-UBR-03 and AC-UBR-07 are bound to M1's
// Exit through these tests; the pruner-level probes of the same criteria are
// retention_probe_p*_test.go and turn green when M2 wires the selection.

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func ver(path, v string) versionKey { return versionKey{path: path, sum: v} }

func crun(id string, keys ...versionKey) censusRun { return censusRun{id: id, keys: keys} }

func exemptRun(id, detail string, keys ...versionKey) censusRun {
	return censusRun{id: id, keys: keys, exempt: true, exemptDetail: detail}
}

func retainedIDs(p retentionPlan) []string {
	out := make([]string, 0, len(p.retain))
	for _, r := range p.retain {
		out = append(out, r.id)
	}
	return out
}

func sorted(ss []string) []string {
	out := append([]string(nil), ss...)
	sort.Strings(out)
	return out
}

// oracleCheck verifies a plan against the census independently of the selection
// code: it recomputes the protected versions, then asserts coverage,
// irredundancy and the REQ-UBR-010 bound. It does not compare against a
// particular cover, because irredundant covers of one key set differ in size
// (design D2).
func oracleCheck(t *testing.T, runs []censusRun, keepCount, pastKeep int, plan retentionPlan) {
	t.Helper()
	if pastKeep < 1 {
		pastKeep = config.DefaultBackupPastVersionsKeep
	}
	window, excess := runs[:keepCount], runs[keepCount:]

	inWindow := map[versionKey]bool{}
	for _, r := range window {
		for _, k := range r.keys {
			inWindow[k] = true
		}
	}
	newest := map[versionKey]int{} // key -> index of its newest excess holder
	for j, r := range excess {
		for _, k := range r.keys {
			if inWindow[k] {
				continue
			}
			if _, ok := newest[k]; !ok {
				newest[k] = j
			}
		}
	}
	byPath := map[string][]versionKey{}
	for k := range newest {
		byPath[k.path] = append(byPath[k.path], k)
	}
	protected := map[versionKey]bool{}
	for _, ks := range byPath {
		sort.Slice(ks, func(a, b int) bool {
			if newest[ks[a]] != newest[ks[b]] {
				return newest[ks[a]] < newest[ks[b]]
			}
			return ks[a].sum < ks[b].sum
		})
		for i := 0; i < len(ks) && i < pastKeep; i++ {
			protected[ks[i]] = true
		}
	}

	// Partition: every excess run is removed or retained, exactly once.
	disposition := map[string]string{}
	for _, id := range plan.remove {
		disposition[id] += "R"
	}
	for _, id := range retainedIDs(plan) {
		disposition[id] += "K"
	}
	for _, r := range excess {
		if got := disposition[r.id]; got != "R" && got != "K" {
			t.Errorf("run %s disposition = %q, want exactly one of removed or retained", r.id, got)
		}
		if r.exempt && disposition[r.id] != "K" {
			t.Errorf("exempt run %s was not retained", r.id)
		}
	}

	retained := map[string]censusRun{}
	for _, r := range excess {
		if disposition[r.id] == "K" {
			retained[r.id] = r
		}
	}
	holdsKey := func(r censusRun, k versionKey) bool {
		for _, rk := range r.keys {
			if rk == k {
				return true
			}
		}
		return false
	}

	for k := range protected {
		held := false
		for _, r := range retained {
			if holdsKey(r, k) {
				held = true
				break
			}
		}
		if !held {
			t.Errorf("protected version %v has no retained holder", k)
		}
	}

	nonExempt := 0
	for id, r := range retained {
		if r.exempt {
			continue
		}
		nonExempt++
		needed := false
		for _, k := range r.keys {
			if !protected[k] {
				continue
			}
			sole := true
			for oid, o := range retained {
				if oid != id && holdsKey(o, k) {
					sole = false
					break
				}
			}
			if sole {
				needed = true
				break
			}
		}
		if !needed {
			t.Errorf("retained non-exempt run %s is redundant: removing it would uncover no protected version", id)
		}
	}
	if nonExempt > len(protected) {
		t.Errorf("retained non-exempt excess runs = %d, exceeds the %d protected versions", nonExempt, len(protected))
	}
	if plan.protected != len(protected) {
		t.Errorf("plan.protected = %d, oracle counted %d", plan.protected, len(protected))
	}
}

// AC-UBR-02: one file backed up as v1, v2, v3, v4, v4 (oldest first), window 1,
// budget 3 — the runs holding v1, v2 and v3 are retained, the older v4 holder is
// removed (REQ-UBR-003, REQ-UBR-004).
func TestAC_UBR_02_SelectionKeepsPastVersionsOutsideTheWindow(t *testing.T) {
	const p = ".claude/agents/x.md"
	runs := []censusRun{ // newest first
		crun("r5", ver(p, "v4")),
		crun("r4", ver(p, "v4")),
		crun("r3", ver(p, "v3")),
		crun("r2", ver(p, "v2")),
		crun("r1", ver(p, "v1")),
	}
	plan := selectRetention(runs, 1, 3)

	if got, want := sorted(plan.remove), []string{"r4"}; !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %v, want %v", got, want)
	}
	if got, want := sorted(retainedIDs(plan)), []string{"r1", "r2", "r3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("retained = %v, want %v", got, want)
	}
	oracleCheck(t, runs, 1, 3, plan)
}

// AC-UBR-03: ten runs holding v1 and a newer run holding v2, window 1 — exactly
// one v1 holder remains, the newest of the ten (REQ-UBR-005).
func TestAC_UBR_03_SelectionRetainsOneCopyOfRepeatedVersion(t *testing.T) {
	const p = ".claude/agents/x.md"
	runs := []censusRun{crun("r11", ver(p, "v2"))}
	for i := 10; i >= 1; i-- {
		runs = append(runs, crun(fmt.Sprintf("r%02d", i), ver(p, "v1")))
	}
	plan := selectRetention(runs, 1, 3)

	if got, want := retainedIDs(plan), []string{"r10"}; !reflect.DeepEqual(got, want) {
		t.Errorf("retained = %v, want the newest v1 holder %v", got, want)
	}
	if len(plan.remove) != 9 {
		t.Errorf("removed %d runs, want 9", len(plan.remove))
	}
	oracleCheck(t, runs, 1, 3, plan)
}

// AC-UBR-07: five distinct versions each held only by excess runs; the budget
// bounds how many stay protected (REQ-UBR-003, REQ-UBR-007).
func TestAC_UBR_07_BudgetBoundsProtection(t *testing.T) {
	const p = "pre-clean/.claude/skills/moai-user/SKILL.md"
	runs := []censusRun{ // newest first; the window run holds an unrelated file
		crun("w", ver("other.yaml", "o")),
		crun("e5", ver(p, "v5")),
		crun("e4", ver(p, "v4")),
		crun("e3", ver(p, "v3")),
		crun("e2", ver(p, "v2")),
		crun("e1", ver(p, "v1")),
	}
	cases := []struct {
		name     string
		pastKeep int
		want     []string
	}{
		{"budget 3", 3, []string{"e3", "e4", "e5"}},
		{"budget 1", 1, []string{"e5"}},
		{"zero reads as the default 3", 0, []string{"e3", "e4", "e5"}},
		{"negative reads as the default 3", -4, []string{"e3", "e4", "e5"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := selectRetention(runs, 1, tc.pastKeep)
			if got := sorted(retainedIDs(plan)); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("retained = %v, want %v", got, tc.want)
			}
			for _, id := range tc.want {
				if !t1733Contains(retainedIDs(plan), id) {
					t.Errorf("run %s should hold one of the newest past versions", id)
				}
			}
			oracleCheck(t, runs, 1, tc.pastKeep, plan)
		})
	}
}

// REQ-UBR-004 / finding F13: a current version that also sits in excess runs
// takes no budget slot, so the path keeps N past versions and not N-1.
func TestSelection_CurrentVersionIsNotAPastVersion(t *testing.T) {
	const p = ".moai/config/sections/quality.yaml"
	runs := []censusRun{
		crun("w", ver(p, "cur")),
		crun("e4", ver(p, "cur")), // the current version, again, in an excess run
		crun("e3", ver(p, "v3")),
		crun("e2", ver(p, "v2")),
		crun("e1", ver(p, "v1")),
	}
	plan := selectRetention(runs, 1, 3)

	if got, want := sorted(retainedIDs(plan)), []string{"e1", "e2", "e3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("retained = %v, want %v (three past versions)", got, want)
	}
	if got, want := plan.remove, []string{"e4"}; !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %v, want %v (the redundant current version)", got, want)
	}
	oracleCheck(t, runs, 1, 3, plan)
}

// Design D2 step 4: the cover is not "newest holder per key". One older run that
// holds the only copy of two versions is retained instead of the two newer runs
// that each hold one of them plus nothing else.
func TestSelection_CoverPrefersTheRunHoldingMostVersions(t *testing.T) {
	runs := []censusRun{
		crun("w", ver("w.txt", "w")),
		crun("y", ver("a.txt", "1")),
		crun("z", ver("b.txt", "1")),
		crun("x", ver("a.txt", "1"), ver("b.txt", "1")),
	}
	plan := selectRetention(runs, 1, 3)

	if got, want := retainedIDs(plan), []string{"x"}; !reflect.DeepEqual(got, want) {
		t.Errorf("retained = %v, want %v", got, want)
	}
	if got, want := sorted(plan.remove), []string{"y", "z"}; !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %v, want %v", got, want)
	}
	oracleCheck(t, runs, 1, 3, plan)
}

// Design D2 step 4, irredundancy pass: greedy picks s1 (the largest) first, then
// s2 and s3 for the keys it misses; s2 and s3 together cover s1's keys, so s1 is
// dropped from the retained set.
func TestSelection_IrredundancyPassDropsACoveredRun(t *testing.T) {
	// Six versions on six paths, so a budget of 1 per path protects each of them.
	runs := []censusRun{
		crun("w", ver("w.txt", "w")),
		crun("s3", ver("p3", "v"), ver("p4", "v"), ver("p6", "v")),
		crun("s2", ver("p1", "v"), ver("p2", "v"), ver("p5", "v")),
		crun("s1", ver("p1", "v"), ver("p2", "v"), ver("p3", "v"), ver("p4", "v")),
	}
	plan := selectRetention(runs, 1, 1)

	if got, want := sorted(retainedIDs(plan)), []string{"s2", "s3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("retained = %v, want %v (s1 is covered by s2 and s3)", got, want)
	}
	oracleCheck(t, runs, 1, 1, plan)
}

// Design D2 bound: the {abc}, {a}, {b}, {c} shape admits irredundant covers of
// size 1 and of size 3; whichever the selection returns must pass the oracle.
func TestSelection_AnIrredundantCoverOfEitherSizePassesTheOracle(t *testing.T) {
	a, b, c := ver("a", "v"), ver("b", "v"), ver("c", "v")
	runs := []censusRun{
		crun("w", ver("w.txt", "w")),
		crun("ra", a),
		crun("rb", b),
		crun("rc", c),
		crun("rabc", a, b, c),
	}
	plan := selectRetention(runs, 1, 3)
	oracleCheck(t, runs, 1, 3, plan)

	// A size-3 cover is also irredundant; the oracle must accept it.
	three := retentionPlan{
		remove:    []string{"rabc"},
		retain:    []retainedRun{{id: "ra", reason: RetentionReasonProtectsVersions, protects: 1}, {id: "rb", reason: RetentionReasonProtectsVersions, protects: 1}, {id: "rc", reason: RetentionReasonProtectsVersions, protects: 1}},
		protected: 3,
	}
	oracleCheck(t, runs, 1, 3, three)
}

// REQ-UBR-008 / AC-UBR-11 at the selection level: an exempt run is retained
// unconditionally and counts as a holder of the keys it holds.
func TestSelection_ExemptRunsAreRetainedAndCountAsHolders(t *testing.T) {
	const p = ".claude/agents/x.md"
	runs := []censusRun{
		crun("w", ver("w.txt", "w")),
		crun("redundant", ver(p, "v1")),                           // held by the exempt run below
		crun("unique", ver(p, "v2")),                              // the only holder of v2
		exemptRun("unreadable", "agents/broken.md", ver(p, "v1")), // exempt, holds v1
		exemptRun("fifo-only", "pipe"),                            // no readable key at all
	}
	plan := selectRetention(runs, 1, 3)

	byID := map[string]retainedRun{}
	for _, r := range plan.retain {
		byID[r.id] = r
	}
	for _, id := range []string{"unreadable", "fifo-only"} {
		r, ok := byID[id]
		if !ok {
			t.Fatalf("exempt run %s was not retained", id)
		}
		if r.reason != RetentionReasonUnclassifiable {
			t.Errorf("run %s reason = %q, want %q", id, r.reason, RetentionReasonUnclassifiable)
		}
	}
	if byID["fifo-only"].detail != "pipe" {
		t.Errorf("fifo-only detail = %q, want the unclassifiable entry", byID["fifo-only"].detail)
	}
	if r := byID["unique"]; r.reason != RetentionReasonProtectsVersions || r.protects != 1 {
		t.Errorf("unique run = %+v, want protects-versions with 1 protected version", r)
	}
	if got, want := plan.remove, []string{"redundant"}; !reflect.DeepEqual(got, want) {
		t.Errorf("removed = %v, want %v (v1 is held by the exempt run)", got, want)
	}
	oracleCheck(t, runs, 1, 3, plan)
}

// REQ-UBR-014: keepCount 0 is an explicit wipe that applies no protection;
// a negative keepCount removes nothing; a window as large as the history leaves
// nothing to decide.
func TestSelection_KeepCountEdges(t *testing.T) {
	const p = ".claude/agents/x.md"
	runs := []censusRun{crun("e3", ver(p, "v3")), crun("e2", ver(p, "v2")), crun("e1", ver(p, "v1"))}

	wipe := selectRetention(runs, 0, 3)
	if got, want := sorted(wipe.remove), []string{"e1", "e2", "e3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("keepCount 0: removed = %v, want every run %v", got, want)
	}
	if len(wipe.retain) != 0 {
		t.Errorf("keepCount 0: retained = %v, want none", retainedIDs(wipe))
	}

	for _, keep := range []int{-1, 3, 4} {
		plan := selectRetention(runs, keep, 3)
		if len(plan.remove) != 0 || len(plan.retain) != 0 {
			t.Errorf("keepCount %d: plan = %+v, want nothing decided", keep, plan)
		}
	}
}

// A seeded property check over random censuses: the plan always satisfies the
// oracle (coverage, irredundancy, bound, partition).
func TestSelection_RandomCensusesSatisfyTheOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(1733))
	for iter := 0; iter < 400; iter++ {
		n := 2 + rng.Intn(8)
		runs := make([]censusRun, n)
		for i := range runs {
			r := censusRun{id: fmt.Sprintf("r%d", i)}
			for k := rng.Intn(4); k > 0; k-- {
				r.keys = append(r.keys, ver(fmt.Sprintf("p%d", rng.Intn(3)), fmt.Sprintf("v%d", rng.Intn(5))))
			}
			if rng.Intn(10) == 0 {
				r.exempt = true
				r.exemptDetail = "pipe"
			}
			runs[i] = r
		}
		keepCount := 1 + rng.Intn(3)
		if keepCount >= n {
			keepCount = n - 1
		}
		pastKeep := 1 + rng.Intn(4)

		plan := selectRetention(runs, keepCount, pastKeep)
		oracleCheck(t, runs, keepCount, pastKeep, plan)
		if t.Failed() {
			t.Fatalf("failing census (iter %d, keepCount %d, pastKeep %d): %+v", iter, keepCount, pastKeep, runs)
		}
	}
}
