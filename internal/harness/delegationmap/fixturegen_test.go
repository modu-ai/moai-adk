package delegationmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/routing"
)

// Fixtures are GENERATED from Go values, never hand-authored (plan.md §G AP-4).
//
// The reason is specific: routing.PendingRow.Finalize maps 15 fields plus the
// outcome and normalizes two nil slices. A hand-written JSONL fixture encodes
// the author's belief about that transform rather than the transform itself, so
// a dropped or mis-mapped field would be invisible to every test that reads it.
// Building each row through Finalize() makes the producer's real output the
// fixture, which is what AC-HLA-001 pins.
//
// Regenerate with:
//
//	MOAI_REGEN_FIXTURES=1 go test ./internal/harness/delegationmap/ -run TestFixtures_GeneratedViaFinalize

// fixtureBase is the deterministic clock anchor for every generated row. A
// wall-clock timestamp would make the committed fixtures differ from a
// regeneration on the next second, which would defeat the byte-identity check.
var fixtureBase = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

// del builds one delegation entry with an observed outcome. Agent is recorded
// verbatim, exactly as the L1 seam records it — including a non-catalog type, a
// spawn name, or the unattributed marker.
func del(agent string) routing.Delegation {
	return routing.Delegation{Agent: agent, Outcome: routing.OutcomeUnknownDelegation}
}

// row builds one finalized ledger row THROUGH PendingRow.Finalize, so the
// fixture carries whatever that transform actually produces.
func row(n int, subcommand string, outcome routing.Outcome, agents ...string) routing.Row {
	dels := make([]routing.Delegation, 0, len(agents))
	for _, a := range agents {
		dels = append(dels, del(a))
	}
	p := routing.PendingRow{
		SchemaVersion:     routing.SchemaVersion,
		CreatedAt:         fixtureBase.Add(time.Duration(n) * time.Minute),
		TS:                fixtureBase.Add(time.Duration(n) * time.Minute).Format(time.RFC3339),
		SessionID:         "fixture-session-" + strconv.Itoa(n),
		ModelClass:        "opus",
		RequestDigest:     "sha256:fixture" + strconv.Itoa(n),
		RequestClass:      "feature",
		MatchedSubcommand: subcommand,
		ClarifyRounds:     0,
		LoopIterations:    0,
		Delegations:       dels,
	}
	return p.Finalize(outcome)
}

// jsonl serializes rows as one JSON object per line.
func jsonl(rows ...routing.Row) []byte {
	var b strings.Builder
	for _, r := range rows {
		line, err := json.Marshal(r)
		if err != nil {
			panic(err) // a fixture that cannot serialize is a programming error
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// repeat builds n rows for one subcommand, calling per(i) for each row's agents.
func repeat(start, n int, subcommand string, outcome routing.Outcome, per func(i int) []string) []routing.Row {
	out := make([]routing.Row, 0, n)
	for i := range n {
		out = append(out, row(start+i, subcommand, outcome, per(i)...))
	}
	return out
}

func concat(groups ...[]routing.Row) []routing.Row {
	var out []routing.Row
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// buildFixtures returns every generated fixture keyed by file name.
//
// The agent composition of each fixture is chosen so exactly the findings the
// corresponding acceptance criterion names are produced — no incidental extra
// finding from a designated agent that happens to be absent.
func buildFixtures() map[string][]byte {
	f := map[string][]byte{}

	// aggregate_basic — 8 qualifying `plan` rows; manager-spec in all 8,
	// Explore in the first 3.
	f["aggregate_basic.jsonl"] = jsonl(repeat(0, 8, "plan", routing.OutcomeSuccess, func(i int) []string {
		if i < 3 {
			return []string{"manager-spec", "Explore"}
		}
		return []string{"manager-spec"}
	})...)

	// reroute_abort_only — no qualifying row at all.
	f["reroute_abort_only.jsonl"] = jsonl(concat(
		repeat(0, 3, "plan", routing.OutcomeReroute, func(int) []string { return []string{"manager-spec"} }),
		repeat(3, 3, "plan", routing.OutcomeAbort, func(int) []string { return []string{"manager-spec"} }),
	)...)

	// non_catalog_agents — 10 qualifying `run` rows. Four non-catalog values
	// appear in 9 of 10 (clear of both thresholds), alongside one retained
	// catalog agent that legitimately qualifies, so a suppressed finding is
	// distinguishable from an empty result.
	f["non_catalog_agents.jsonl"] = jsonl(repeat(0, 10, "run", routing.OutcomeSuccess, func(i int) []string {
		if i < 9 {
			return []string{
				"general-purpose",
				"workflow-subagent",
				"hns-oss-docs-locale-translator-specialist",
				"audit-hle",
				"manager-docs",
			}
		}
		return []string{"manager-develop"}
	})...)

	// unattributed_share — 10 qualifying `plan` rows carrying 4 unattributed
	// delegation entries, plus one undesignated catalog agent that qualifies so
	// a proposal is actually emitted for `plan`.
	f["unattributed_share.jsonl"] = jsonl(repeat(0, 10, "plan", routing.OutcomeSuccess, func(i int) []string {
		agents := []string{"manager-spec"}
		if i == 0 {
			agents = append(agents, "Explore")
		}
		if i < 9 {
			agents = append(agents, "manager-develop")
		}
		if i < 4 {
			agents = append(agents, routing.AgentUnattributed)
		}
		return agents
	})...)

	// below_min_rows — 4 qualifying `plan` rows against MinQualifyingRows = 5.
	f["below_min_rows.jsonl"] = jsonl(repeat(0, 4, "plan", routing.OutcomeSuccess, func(int) []string {
		return []string{"manager-spec", "Explore", "manager-develop"}
	})...)

	// stray_agent — 10 qualifying `run` rows in which an undesignated catalog
	// agent appears exactly once (support 0.10, below MinSupportRatio).
	f["stray_agent.jsonl"] = jsonl(repeat(0, 10, "run", routing.OutcomeSuccess, func(i int) []string {
		if i == 0 {
			return []string{"manager-develop", "manager-docs"}
		}
		return []string{"manager-develop"}
	})...)

	// two_kinds — one finding of each kind and nothing else. `run` carries an
	// undesignated catalog agent at full support; `sync` carries only a
	// non-catalog value, leaving its designated manager-docs never spawned
	// (sync-auditor is conditionally invoked and therefore excluded).
	f["two_kinds.jsonl"] = jsonl(concat(
		repeat(0, 8, "run", routing.OutcomeSuccess, func(int) []string {
			return []string{"manager-develop", "manager-docs"}
		}),
		repeat(8, 6, "sync", routing.OutcomeSuccess, func(int) []string {
			return []string{"general-purpose"}
		}),
	)...)

	// conditional_designations — sync-auditor and plan-auditor are both absent
	// and both excluded; Explore is absent, designated for `plan`, and NOT
	// excluded, so its finding still fires.
	f["conditional_designations.jsonl"] = jsonl(concat(
		repeat(0, 12, "sync", routing.OutcomeSuccess, func(int) []string {
			return []string{"manager-docs"}
		}),
		repeat(12, 12, "plan", routing.OutcomeSuccess, func(int) []string {
			return []string{"manager-spec"}
		}),
	)...)

	// oversized — a normal small ledger. The oversize path is exercised by
	// lowering Options.MaxLedgerBytes in the test rather than committing a
	// multi-megabyte blob, which would cost the repository far more than the
	// guard it demonstrates.
	f["oversized.jsonl"] = jsonl(repeat(0, 2, "plan", routing.OutcomeSuccess, func(int) []string {
		return []string{"manager-spec"}
	})...)

	// all_malformed — deliberately not routed through Finalize: these lines are
	// not rows, which is the point.
	f["all_malformed.jsonl"] = []byte("{not json\n[[[\nnope\n")

	// empty — a present but empty ledger.
	f["empty.jsonl"] = []byte("")

	return f
}

// TestFixtures_GeneratedViaFinalize is AC-HLA-001. It re-runs the generator and
// requires the committed fixtures to be byte-identical, and requires at least
// one fixture to have been produced through PendingRow.Finalize rather than
// hand-authored JSON.
func TestFixtures_GeneratedViaFinalize(t *testing.T) {
	fixtures := buildFixtures()
	dir := "testdata"

	if os.Getenv("MOAI_REGEN_FIXTURES") == "1" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, body := range fixtures {
			if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		t.Log("fixtures regenerated")
	}

	names := make([]string, 0, len(fixtures))
	for name := range fixtures {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("committed fixture %s missing: %v (regenerate with MOAI_REGEN_FIXTURES=1)", name, err)
		}
		if string(got) != string(fixtures[name]) {
			t.Errorf("committed fixture %s has drifted from the generator output", name)
		}
	}

	// The transform pin: one row built through Finalize must round-trip every
	// field the generator set, so a dropped or mis-mapped field fails here.
	r := row(0, "plan", routing.OutcomeSuccess, "manager-spec")
	if r.SchemaVersion != routing.SchemaVersion ||
		r.SessionID != "fixture-session-0" ||
		r.ModelClass != "opus" ||
		r.RequestDigest != "sha256:fixture0" ||
		r.RequestClass != "feature" ||
		r.MatchedSubcommand != "plan" ||
		r.Outcome != routing.OutcomeSuccess ||
		len(r.Delegations) != 1 ||
		r.Delegations[0].Agent != "manager-spec" ||
		r.EvidenceRefs == nil {
		t.Fatalf("PendingRow.Finalize did not carry the fixture fields through: %+v", r)
	}
}
