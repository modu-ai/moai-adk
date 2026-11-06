//go:build af_e2e

// scenario_suite_af_e2e_test.go: AC-AF-025 (REQ-AF-006, auxiliary REQ-AF-001,
// 004, 005, 011), acceptance.md §A.1. One suite, seven ordered subtests 01_…~07_…
// over one real Issue pair, one run. Excluded from the default run by the
// af_e2e build tag; run it with:
//
//	MOAI_AF_E2E_ISSUE=<n> MOAI_AF_E2E_NEXT_ISSUE=<m> \
//	  go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenarioSuite$' -count=1 -v
//
// Contract (§A.1): the verdict is the seven `--- PASS: TestScenarioSuite/<name> `
// lines, never the suite's `ok` or exit code. A missing environment variable is
// a FAIL, not a skip. A subtest checks the state its predecessor left on the
// same Issue, so one alone FAILs. A non-product attribution source FAILs the
// suite before 01.
//
// What this file does today, stated as facts:
//   - 01 runs against the real claim store (two real claimants, no injection).
//   - 02..07 are UNMET: each fails with the missing product piece named, because
//     no code in this tree provides it (see the unmet strings). They are not
//     skipped and not faked, so the suite cannot be read as green.
//   - The attribution-source preflight is UNMET: no product API reports the
//     source kind of an attribution (product vs stand-in).
//
// Real GitHub writes (comments 01-05, the protected main merge and Issue close
// 06-07) are operator keep-set (OD-6); nothing here performs one.
package e2e

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

type suiteState struct {
	issue, next int
	store       *claim.Store
	ref         core.IssueRef
	run         core.RunID
	winner      string
	epoch       int64
	done        map[string]bool
}

func suiteEnvInt(t *testing.T, name string) int {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Fatalf("AF_SCENARIO_SUITE: %s is not set (a missing variable is a FAIL, never a skip)", name)
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		t.Fatalf("AF_SCENARIO_SUITE: %s=%q is not an Issue number", name, v)
	}
	return n
}

func TestScenarioSuite(t *testing.T) {
	st := &suiteState{done: map[string]bool{}}
	st.issue = suiteEnvInt(t, "MOAI_AF_E2E_ISSUE")
	st.next = suiteEnvInt(t, "MOAI_AF_E2E_NEXT_ISSUE")
	if st.issue == st.next {
		t.Fatalf("AF_SCENARIO_SUITE: the Issue pair must be two different Issues, got %d twice", st.issue)
	}
	store, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	st.store = store
	if st.ref, err = core.NewIssueRef(st.issue); err != nil {
		t.Fatal(err)
	}
	if st.run, err = core.ParseRunID("run-suite-1"); err != nil {
		t.Fatal(err)
	}

	// Preflight, before 01: the attribution source must be the product's.
	t.Run("00_attribution_source_is_product", func(t *testing.T) {
		t.Errorf("AF_SCENARIO_SUITE unmet: no product API reports the source kind of an attribution (product vs stand-in), so the harness cannot record or check it; owner: review/ (lane 18/20)")
	})

	subtests := []struct {
		name string
		run  func(t *testing.T, s *suiteState)
	}{
		{"01_competing_claim", suite01CompetingClaim},
		{"02_duplicate_dispatch", suiteUnmet("02", "a dispatch re-send path that reports whether a live attempt exists; the runner/effect surface has no dispatch API callable from a harness (fault-free), and the undelivered-dispatch branch needs an isolated copy")},
		{"03_ack_loss", suiteUnmet("03", "a transport layer with an ACK-dropping injection point on a real command send; effect/ has the delivery states but no injectable transport")},
		{"04_interrupted_recovery", suiteUnmet("04", "a clock injection that expires the lease of a real agent process, plus checkpoint remote-durable resume with a new epoch")},
		{"05_audit_budget_exhausted", suiteUnmet("05", "a real audit round recorder and a unlocked producer (M7); on the locked tree 05 stops at producer-not-allowed by design")},
		{"06_merge_result_unknown", suiteUnmet("06", "the 05->06 evidence production chain (PR, receipt, required checks, four distinct role sessions, reviews, debt-free audit PASS, QA PASS, E08) and a merge request whose response can be dropped; needs operator keep-set approval for a real protected-main merge")},
		{"07_remote_landing_to_done_and_next_dispatch", suiteUnmet("07", "E12, local sync, Issue close, DONE and a second-Issue dispatch left at PARKED operator-hold (OD-12 pending); needs operator keep-set approval")},
	}
	for _, sub := range subtests {
		t.Run(sub.name, func(t *testing.T) { sub.run(t, st) })
	}
}

func suiteUnmet(id, missing string) func(*testing.T, *suiteState) {
	return func(t *testing.T, s *suiteState) {
		t.Helper()
		if !s.done["01"] && id != "01" {
			t.Errorf("AF_SCENARIO_SUITE %s: its predecessor left no state on Issue #%d (a subtest alone FAILs, never skips)", id, s.issue)
		}
		t.Fatalf("AF_SCENARIO_SUITE unmet %s: %s", id, missing)
	}
}

// 01: two real claimants claim the same Issue at once. Exactly one wins; the
// other gets a retryable Held or Busy and no attempt; the winner holds epoch 1.
func suite01CompetingClaim(t *testing.T, s *suiteState) {
	const claimants = 2
	errs := make([]error, claimants)
	recs := make([]claim.Record, claimants)
	holders := []string{"suite-claimant-a", "suite-claimant-b"}
	var wg sync.WaitGroup
	for i := range claimants {
		wg.Go(func() {
			recs[i], errs[i] = s.store.Acquire(s.ref, holders[i], "dispatch-"+holders[i], s.run)
		})
	}
	wg.Wait()
	wins := 0
	for i, err := range errs {
		switch {
		case err == nil:
			wins++
			s.winner = holders[i]
		case errors.Is(err, core.ErrHeld), errors.Is(err, core.ErrBusy):
		default:
			t.Errorf("claimant %s: unexpected error %v", holders[i], err)
		}
	}
	if wins != 1 {
		t.Fatalf("winners = %d, want exactly 1", wins)
	}
	cur, err := s.store.Load(s.ref)
	if err != nil {
		t.Fatal(err)
	}
	if cur.Holder != s.winner || cur.Epoch != 1 {
		t.Fatalf("record holder %q epoch %d, want winner %q epoch 1", cur.Holder, cur.Epoch, s.winner)
	}
	s.epoch = cur.Epoch
	s.done["01"] = true
}
