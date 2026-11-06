package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M1 RED — the delta record and its
// persistence (AC-AEC-001f; REQ-AEC-005; card t2017).
//
// Kickoff not passed - pre-implementation RED.
//
// Today a delta call files no second-review record (mcp_convergence.go:1059) and
// persists its result only when it fails (:1046). Each RED below therefore fails
// on the record count or on the persisted verdict, never on a compile error.
// Tests marked "control" pass today and pin the behavior the change must keep
// (a full-scope call, and the card_id-less delta of card t1873).
//
// The additive record keys assumed by the field subtests are declared in
// audit_delta_chain_world_test.go.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
)

// lastRecordLine returns the raw JSON text of the newest record line.
func (w *aecWorld) lastRecordLine() string {
	w.t.Helper()
	data, err := os.ReadFile(w.recordsPath())
	if err != nil {
		w.t.Fatalf("read records: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return lines[len(lines)-1]
}

// seedFailSession writes a failing full-review result to the session file, as a
// prior full review would have.
func (w *aecWorld) seedFailSession(session string) {
	w.t.Helper()
	seed := ConvergenceResult{OverallVerdict: "fail", PerBackendVerdicts: []PerBackendVerdict{}}
	if err := persistConvergenceResult(seed, session, w.tree); err != nil {
		w.t.Fatalf("seed session state: %v", err)
	}
}

// sessionVerdicts returns the overall verdict of every result persisted for the
// session in the tree's store.
func (w *aecWorld) sessionVerdicts(session string) []string {
	var out []string
	for _, r := range loadConvergenceResults(w.tree, session) {
		out = append(out, r.OverallVerdict)
	}
	return out
}

// hasDeltaMarker reports whether a record carries a delta marker: a boolean true
// under a key that names "delta", or the string "delta" under any key. The SPEC
// requires "a delta marker" without fixing its shape.
func hasDeltaMarker(m map[string]any) bool {
	for k, v := range m {
		if b, ok := v.(bool); ok && b && strings.Contains(strings.ToLower(k), "delta") {
			return true
		}
		if s, ok := v.(string); ok && s == "delta" {
			return true
		}
		if nested, ok := v.(map[string]any); ok && hasDeltaMarker(nested) {
			return true
		}
	}
	return false
}

func scopeBase(m map[string]any) string {
	scope, _ := m["scope"].(map[string]any)
	s, _ := scope["base_sha"].(string)
	return s
}

// AC-AEC-001f — a delta call admitted by the chain gate files exactly one
// second-review record carrying the delta base as its scope base, a delta
// marker, the receipt id of the record it extends, the ids its confirmation
// named, and the SHA-256 of each relied-on FAIL verdict file.
func TestAEC001f_GatedDelta_FilesOneDeltaRecord(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	reliedOn := aecFileSHA256(t, c.f0)
	before := len(w.records())
	if before != 2 {
		t.Fatalf("setup: want R0 and R1 on disk, have %d records", before)
	}

	out := w.call(c.call())
	wantScope(t, out, c.h1) // control: the gate admits the unbroken chain

	records := w.records()
	if len(records) != before+1 {
		t.Errorf("exactly one second-review record must be appended by an admitted delta call: %d -> %d", before, len(records))
	}
	// The field checks read the newest line; with no new record they would read
	// R1 and could pass for the wrong reason, so they run only on a new record.
	fresh := len(records) == before+1
	line := w.lastRecordLine()
	var rec map[string]any
	if fresh {
		rec = records[len(records)-1]
	}

	t.Run("scope base is the delta base", func(t *testing.T) {
		if !fresh {
			t.Fatalf("no new record was filed (%d records)", len(records))
		}
		if got := scopeBase(rec); got != c.h1 {
			t.Errorf("scope.base_sha = %q, want the delta base %q", got, c.h1)
		}
		if got, _ := rec["head_sha"].(string); got != c.h2 {
			t.Errorf("head_sha = %q, want the audited HEAD %q", got, c.h2)
		}
	})
	t.Run("carries a delta marker", func(t *testing.T) {
		if !fresh {
			t.Fatalf("no new record was filed (%d records)", len(records))
		}
		if !hasDeltaMarker(rec) {
			t.Errorf("record carries no delta marker: %s", line)
		}
	})
	t.Run("names the receipt of the record it extends", func(t *testing.T) {
		if !fresh {
			t.Fatalf("no new record was filed (%d records)", len(records))
		}
		if !strings.Contains(line, c.r1.receipt) {
			t.Errorf("record does not carry R1's receipt id %s: %s", c.r1.receipt, line)
		}
	})
	t.Run("names the ids its confirmation named", func(t *testing.T) {
		if !fresh {
			t.Fatalf("no new record was filed (%d records)", len(records))
		}
		if !strings.Contains(line, "B1") {
			t.Errorf("record does not carry the confirmed id B1: %s", line)
		}
	})
	t.Run("carries the SHA-256 of the relied-on FAIL verdict file", func(t *testing.T) {
		if !fresh {
			t.Fatalf("no new record was filed (%d records)", len(records))
		}
		if !strings.Contains(line, reliedOn) {
			t.Errorf("record does not carry the SHA-256 %s of %s: %s", reliedOn, c.f0, line)
		}
	})
}

// AC-AEC-001f — with two FAIL files forming the closure set the record carries
// the digest of each. A digest of only the newest file, or of none, fails here.
func TestAEC001f_GatedDelta_RecordCarriesTheDigestOfEachReliedOnFile(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}})
	second := w.verdictFile("sync-audit-1.md", c.h1, aecN(1), "B3")
	d0, d1 := aecFileSHA256(t, c.f0), aecFileSHA256(t, second)
	before := len(w.records())

	out := w.call(aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B2", c.k1, "B3", c.k2)})
	wantScope(t, out, c.h1)

	if len(w.records()) != before+1 {
		t.Fatalf("no second-review record was filed by the admitted delta call (%d -> %d)", before, len(w.records()))
	}
	line := w.lastRecordLine()
	for name, d := range map[string]string{"plan-audit-1.md": d0, "sync-audit-1.md": d1} {
		if !strings.Contains(line, d) {
			t.Errorf("record does not carry the SHA-256 of %s (%s): %s", name, d, line)
		}
	}
}

// AC-AEC-001f — SelectSecondReview keeps the new record: it is a baseBranch
// record whose scope head is the audited commit, with at least one changed file
// and a second-model verdict. Fed ONLY the new record, so an older record cannot
// satisfy it.
func TestAEC001f_GatedDelta_RecordIsKeptBySelectSecondReview(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	before := len(w.records())

	wantScope(t, w.call(c.call()), c.h1)

	loaded, skipped, err := closure.LoadSecondReviews(w.recordsPath())
	if err != nil || len(skipped) != 0 {
		t.Fatalf("load records: err=%v skipped=%v", err, skipped)
	}
	if len(loaded) != before+1 {
		t.Fatalf("no second-review record was filed by the admitted delta call (%d -> %d)", before, len(loaded))
	}
	rec := loaded[len(loaded)-1]
	if st := w.selectState(rec); st.State != closure.SecondReviewPerformed {
		t.Errorf("SelectSecondReview dropped the delta record: state=%s cause=%s (scope=%+v)", st.State, st.Cause, rec.Scope)
	}
}

// selectState runs SelectSecondReview over exactly one record, bound to the
// fixture's contract and evaluated at the current HEAD.
func (w *aecWorld) selectState(rec closure.SecondReviewRecord) closure.SecondReviewState {
	return closure.SelectSecondReview(closure.SecondReviewInput{
		Records:        []closure.SecondReviewRecord{rec},
		ContractCard:   rec.ContractCard,
		ContractDigest: rec.ContractSHA256,
		EvalCommit:     w.head(),
		SpecID:         closuretest.SpecID,
		Facts: closure.GitFacts{
			IsAncestor: func(a, b string) (bool, error) { return gitio.IsAncestor(w.f.CardDir, a, b) },
			NonMergeCommits: func(string, string) ([]closure.CommitPaths, error) {
				return nil, nil
			},
		},
	})
}

// AC-AEC-001f — control (passes today): the same SelectSecondReview input keeps
// a full-scope record written by a real full-scope call. It proves the input
// scaffolding above, so a red there is the missing delta record and nothing else.
func TestAEC001f_FullScopeRecord_IsKeptBySelectSecondReview_Control(t *testing.T) {
	w := newAECWorld(t)
	w.happyChain(aecChainOpts{})
	wantScope(t, w.call(aecCallOpts{}), "")

	loaded, skipped, err := closure.LoadSecondReviews(w.recordsPath())
	if err != nil || len(skipped) != 0 {
		t.Fatalf("load records: err=%v skipped=%v", err, skipped)
	}
	rec := loaded[len(loaded)-1]
	if st := w.selectState(rec); st.State != closure.SecondReviewPerformed {
		t.Errorf("control: SelectSecondReview dropped a full-scope record: state=%s cause=%s (scope=%+v)", st.State, st.Cause, rec.Scope)
	}
}

// AC-AEC-001f — the session file holds the gated delta call's result even when
// its overall verdict is pass (the Stop gate must see the delta PASS that
// replaces the earlier full FAIL).
func TestAEC001f_GatedDelta_PersistsItsPassResult(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	w.seedFailSession(w.session)
	if got := w.sessionVerdicts(w.session); len(got) != 1 || got[0] != "fail" {
		t.Fatalf("control: seeded session state = %v, want [fail]", got)
	}

	out := w.call(c.call())
	wantScope(t, out, c.h1)
	if out.result.OverallVerdict != "pass" {
		t.Fatalf("setup: the delta call's overall verdict = %q, want pass", out.result.OverallVerdict)
	}

	got := w.sessionVerdicts(w.session)
	if len(got) != 1 || got[0] != "pass" {
		t.Errorf("session file after a gated delta PASS = %v, want [pass]", got)
	}
}

// AC-AEC-001f — a delta call the gate sent to full scope is a full-scope call:
// its record is a full-scope record (no delta marker, the resolved base as scope
// base) and its result persists exactly as a full call's does.
func TestAEC001f_GateRejectedDelta_IsAFullScopeCall(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{r1BaseParentOfH0: true}) // broken chain
	w.seedFailSession(w.session)
	records := w.records()
	before := len(records)
	fullBase := scopeBase(records[0]) // R0 is a full-scope record: the resolved base

	out := w.call(c.call())
	wantScope(t, out, "")

	after := w.records()
	t.Run("files a full-scope record", func(t *testing.T) {
		if len(after) != before+1 {
			t.Fatalf("a full-scope call files one record: %d -> %d", before, len(after))
		}
		rec := after[len(after)-1]
		if hasDeltaMarker(rec) {
			t.Errorf("the record of a gate-rejected delta carries a delta marker: %s", w.lastRecordLine())
		}
		if got := scopeBase(rec); got != fullBase {
			t.Errorf("scope.base_sha = %q, want the resolved base %q (not the proposed delta base)", got, fullBase)
		}
	})
	t.Run("persists as a full-scope call", func(t *testing.T) {
		got := w.sessionVerdicts(w.session)
		if len(got) != 1 || got[0] != "pass" {
			t.Errorf("session file after a full-scope pass = %v, want [pass] (a full call overwrites)", got)
		}
	})
}

// AC-AEC-001f — control (t1873, passes today): a delta PASS without a card_id
// leaves the session file unchanged and files no record.
func TestAEC001f_DeltaPassWithoutCardID_LeavesTheSessionFileAlone_Control(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	w.seedFailSession(w.session)
	before := len(w.records())

	out := w.call(aecCallOpts{deltaBase: c.h1, noCard: true})
	wantScope(t, out, c.h1)
	if out.result.OverallVerdict != "pass" {
		t.Fatalf("setup: the delta call's overall verdict = %q, want pass", out.result.OverallVerdict)
	}

	if got := w.sessionVerdicts(w.session); len(got) != 1 || got[0] != "fail" {
		t.Errorf("a card_id-less delta PASS changed the session file: %v, want [fail] (t1873)", got)
	}
	if after := len(w.records()); after != before {
		t.Errorf("a card_id-less call filed a record: %d -> %d", before, after)
	}
}

// AC-AEC-001f — control (passes today): a full-scope call's record is shaped as
// today, with no delta marker, and its result is persisted. It also proves the
// fixture: the handler, receipt store, evidence home and session file are the
// real ones.
func TestAEC001f_FullScopeCall_RecordIsShapedAsToday_Control(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	w.seedFailSession(w.session)
	before := len(w.records())

	out := w.call(aecCallOpts{}) // no delta_base: the resolved base
	wantScope(t, out, "")

	records := w.records()
	if len(records) != before+1 {
		t.Fatalf("a full-scope call files one record: %d -> %d", before, len(records))
	}
	rec := records[len(records)-1]
	if hasDeltaMarker(rec) {
		t.Errorf("a full-scope record carries a delta marker: %s", w.lastRecordLine())
	}
	for _, k := range []string{aecRecDeltaBaseKey, aecRecExtendsKey, aecRecConfirmedKey, aecRecDigestsKey} {
		if _, present := rec[k]; present {
			t.Errorf("a full-scope record carries the delta-only key %q: %s", k, w.lastRecordLine())
		}
	}
	if got, _ := rec["head_sha"].(string); got != c.h2 {
		t.Errorf("head_sha = %q, want %q", got, c.h2)
	}
	if got := w.sessionVerdicts(w.session); len(got) != 1 || got[0] != "pass" {
		t.Errorf("a full-scope call must persist its result: %v, want [pass]", got)
	}
}
