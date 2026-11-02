package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M1 RED — the delta chain gate
// (AC-AEC-001c, AC-AEC-001d, AC-AEC-001e; REQ-AEC-004; card t2017).
//
// Kickoff not passed - pre-implementation RED.
//
// Every test drives the real audit_multi handler over a card tree whose
// second-review records, receipts and verdict files are on disk, with the leg
// seam faked. "Full scope" is what the legs observe on their context: a leg that
// sees an empty delta base measures from the resolved base; a leg that sees a
// SHA measures delta_base...HEAD. Today the handler takes any valid ancestor
// delta_base as given, so every full-scope expectation below fails on its scope
// assertion; the rows named "control" pass today and keep the gate from being
// satisfied by an always-full-scope mutant.
//
// Wire names the fixtures assume (confirmation argument, record keys) are
// declared in audit_delta_chain_world_test.go.

import (
	"strings"
	"testing"
)

// chainCall is the call every row starts from: the delta proposed from R1's
// head with the confirmation naming B1 closed by k1.
func (c aecChain) call() aecCallOpts {
	return aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1)}
}

// AC-AEC-001c — control: an unbroken, reviewed chain keeps the delta scope.
// Passes today (a delta is always taken); it guards the "always full scope"
// mutant of the gate.
func TestAEC001c_UnbrokenReviewedChain_ReviewsTheDeltaScope_Control(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	out := w.call(c.call())
	wantScope(t, out, c.h1)
}

// AC-AEC-001c — each broken-chain fixture must be reviewed at full scope. The
// gap row has a valid last link and an ancestor base; the inconclusive, absent
// glm entry and receipt rows are well-formed links that never ran a required
// backend (the required set is {codex, glm}); each blocks a different mutant of
// the gate (last-link-only, ancestry-only, any-record-that-exists,
// entries-present-only).
func TestAEC001c_BrokenChain_ReviewsFullScope(t *testing.T) {
	inconclusiveGLM := map[string]string{BackendClaude: "pass", BackendCodex: "pass", BackendGLM: "inconclusive"}
	noGLM := map[string]string{BackendClaude: "pass", BackendCodex: "pass"}

	rows := []struct {
		name  string
		opts  aecChainOpts
		prep  func(w *aecWorld, c aecChain) // optional change after the chain is built
		call  func(c aecChain) aecCallOpts  // default: aecChain.call
		named func(c aecChain) []string     // commits the result must name; nil = not asserted
	}{
		{
			name:  "a link's base differs from the previous record's head (gap)",
			opts:  aecChainOpts{r1BaseParentOfH0: true},
			named: func(c aecChain) []string { return []string{c.h1} },
		},
		{
			name: "the proposed base differs from the last record's head",
			call: func(c aecChain) aecCallOpts {
				return aecCallOpts{deltaBase: c.h0, confirm: aecClosure("B1", c.k1)}
			},
			named: func(c aecChain) []string { return []string{c.h0, c.h1} },
		},
		{
			name: "no full-scope record starts the chain",
			prep: func(w *aecWorld, _ aecChain) { w.keepRecords(1) },
		},
		{
			name:  "R0 records a required backend inconclusive",
			opts:  aecChainOpts{r0: aecRecOpts{backends: inconclusiveGLM}},
			named: func(c aecChain) []string { return []string{c.h0} },
		},
		{
			name:  "R1 records a required backend inconclusive",
			opts:  aecChainOpts{r1: aecRecOpts{backends: inconclusiveGLM}},
			named: func(c aecChain) []string { return []string{c.h1} },
		},
		{
			name:  "R0 carries a codex pass and no glm entry at all",
			opts:  aecChainOpts{r0: aecRecOpts{backends: noGLM}},
			named: func(c aecChain) []string { return []string{c.h0} },
		},
		{
			name:  "R1 carries a codex pass and no glm entry at all",
			opts:  aecChainOpts{r1: aecRecOpts{backends: noGLM}},
			named: func(c aecChain) []string { return []string{c.h1} },
		},
		{
			name:  "R0's receipt has a non-empty gate_unmet",
			opts:  aecChainOpts{r0: aecRecOpts{receipt: aecReceipt{gateUnmet: "codex gate unmet"}}},
			named: func(c aecChain) []string { return []string{c.h0} },
		},
		{
			name:  "R1's receipt has a non-empty gate_unmet",
			opts:  aecChainOpts{r1: aecRecOpts{receipt: aecReceipt{gateUnmet: "codex gate unmet"}}},
			named: func(c aecChain) []string { return []string{c.h1} },
		},
		{
			name:  "R0's receipt carries an inconclusive codex verdict",
			opts:  aecChainOpts{r0: aecRecOpts{receipt: aecReceipt{codex: "inconclusive"}}},
			named: func(c aecChain) []string { return []string{c.h0} },
		},
		{
			name:  "R1's receipt carries an inconclusive codex verdict",
			opts:  aecChainOpts{r1: aecRecOpts{receipt: aecReceipt{codex: "inconclusive"}}},
			named: func(c aecChain) []string { return []string{c.h1} },
		},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			w := newAECWorld(t)
			c := w.happyChain(row.opts)
			if row.prep != nil {
				row.prep(w, c)
			}
			call := c.call()
			if row.call != nil {
				call = row.call(c)
			}
			out := w.call(call)
			wantScope(t, out, "")
			if row.named != nil {
				wantNames(t, out, row.named(c)...)
			}
		})
	}
}

// AC-AEC-001d — the closure confirmation. Rows with want == "" must be reviewed
// at full scope; rows with a want are the complete-confirmation twins that keep a
// "reject everything" mutant from passing (they pass today).
func TestAEC001d_ClosureConfirmation(t *testing.T) {
	type row struct {
		name  string
		build func(w *aecWorld) (aecChain, aecCallOpts)
		// wantDelta: true = the delta from R1's head is reviewed (a control row
		// that passes today); false = the full scope is reviewed.
		wantDelta bool
	}
	// secondFail adds the sync-audit FAIL file tied to R1 (audited at h1).
	secondFail := func(w *aecWorld, c aecChain, count int, ids ...string) {
		w.verdictFile("sync-audit-1.md", c.h1, aecN(count), ids...)
	}
	rows := []row{
		{
			name: "the confirmation is omitted",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{})
				return c, aecCallOpts{deltaBase: c.h1}
			},
		},
		{
			name: "the confirmation names only some ids of the union",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}})
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1)}
			},
		},
		{
			name: "control: the confirmation names every id of the union",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}})
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B2", c.k1)}
			},
			wantDelta: true,
		},
		{
			name: "the closing commit of an id new since the previous link lies outside delta_base..HEAD",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{})
				secondFail(w, c, 1, "B3")
				// k1 precedes h1, so it is outside h1..HEAD.
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B3", c.k1)}
			},
		},
		{
			name: "control: the closing commit of the new id lies inside delta_base..HEAD",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{})
				secondFail(w, c, 1, "B3")
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B3", c.k2)}
			},
			wantDelta: true,
		},
		{
			name: "a FAIL file's blocking_defect count differs from its blocking_count",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}, f0Count: aecN(1)})
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B2", c.k1)}
			},
		},
		{
			name: "a FAIL file carries no blocking_count",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{f0NoCount: true})
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1)}
			},
		},
		{
			name: "the card has no FAIL verdict file tied to the chain (empty union, empty confirmation)",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{noF0: true})
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure()}
			},
		},
		{
			name: "two FAIL files (plan X,Y; sync Z): a confirmation naming only Z is incomplete",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}})
				secondFail(w, c, 1, "B3")
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B3", c.k2)}
			},
		},
		{
			name: "control: two FAIL files, a confirmation naming X, Y and Z is complete (X, Y inherited from R1)",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}})
				secondFail(w, c, 1, "B3")
				// k1 closed B1 and B2 and lies outside h1..HEAD: the inherited closure.
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B2", c.k1, "B3", c.k2)}
			},
			wantDelta: true,
		},
		{
			name: "inheritance: R1 recorded only B1 as confirmed, so the same R2 proposal reviews at full scope",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}, r1Confirmed: []string{"B1"}})
				secondFail(w, c, 1, "B3")
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B2", c.k1, "B3", c.k2)}
			},
		},
		{
			name: "re-report: R1's FAIL names B1 again and no commit after R1's head closes it",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{})
				secondFail(w, c, 1, "B1")
				// k1 is the old closing commit, before R1's head.
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1)}
			},
		},
		{
			name: "control: re-report with a closing commit inside delta_base..HEAD keeps the delta scope",
			build: func(w *aecWorld) (aecChain, aecCallOpts) {
				c := w.happyChain(aecChainOpts{})
				secondFail(w, c, 1, "B1")
				return c, aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k2)}
			},
			wantDelta: true,
		},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			w := newAECWorld(t)
			c, call := r.build(w)
			out := w.call(call)
			if r.wantDelta {
				wantScope(t, out, c.h1)
				return
			}
			wantScope(t, out, "")
		})
	}
}

// AC-AEC-001e — the chain is read from server-written state only. Each row
// differs from the happy chain in exactly one place and must be reviewed at full
// scope.
func TestAEC001e_ChainIsReadFromServerWrittenStateOnly(t *testing.T) {
	t.Run("a link's receipt is absent from the receipt store", func(t *testing.T) {
		for _, which := range []string{"R0", "R1"} {
			t.Run(which, func(t *testing.T) {
				w := newAECWorld(t)
				o := aecChainOpts{}
				if which == "R0" {
					o.r0 = aecRecOpts{receipt: aecReceipt{absent: true}}
				} else {
					o.r1 = aecRecOpts{receipt: aecReceipt{absent: true}}
				}
				c := w.happyChain(o)
				wantScope(t, w.call(c.call()), "")
			})
		}
	})
	t.Run("a receipt was recorded by another tool", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{r1: aecRecOpts{receipt: aecReceipt{tool: "codex_audit"}}})
		wantScope(t, w.call(c.call()), "")
	})
	t.Run("a receipt names another tree root", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{r1: aecRecOpts{receipt: aecReceipt{treeRoot: w.f.Root}}})
		wantScope(t, w.call(c.call()), "")
	})
	t.Run("a receipt's head commit disagrees with the record", func(t *testing.T) {
		w := newAECWorld(t)
		// R1's receipt binds h0 (the HEAD before the chain starts), not R1's head.
		h0 := w.head()
		c := w.happyChain(aecChainOpts{r1: aecRecOpts{receipt: aecReceipt{head: h0}}})
		wantScope(t, w.call(c.call()), "")
	})

	// The verdict files form a perfect chain (F0 audited h0, F1 audited h1) but
	// no server-written record exists: a chain built from verdict-file
	// audited_sha values must not be accepted.
	t.Run("chain present only in verdict files", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{})
		w.verdictFile("sync-audit-1.md", c.h1, aecN(0))
		w.keepRecords() // no records at all
		wantScope(t, w.call(c.call()), "")
	})

	// Chain data supplied as call arguments with no records on disk: the
	// arguments must not stand in for the server's own records.
	t.Run("chain data supplied as a call argument", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{})
		forged := []any{
			map[string]any{"head_sha": c.h0, "target": codexTargetBaseBranch, "audit_receipt": c.r0.receipt},
			map[string]any{"head_sha": c.h1, "target": codexTargetBaseBranch, "audit_receipt": c.r1.receipt,
				"scope": map[string]any{"base_sha": c.h0}},
		}
		w.keepRecords()
		call := c.call()
		call.extra = map[string]any{"chain": forged, "records": forged, "second_review_records": forged}
		wantScope(t, w.call(call), "")
	})
}

// AC-AEC-001d (last sentence) — when the confirmation is complete, every leg's
// request carries it. The carrier is the focus text each leg receives (the
// existing argument, observable at the leg seam), and the caller's own focus is
// kept. Ids inherited under REQ-AEC-004 (ii) may map to "": they are still
// named. Today the legs get the caller's focus unchanged, so this fails on the
// missing ids.
func TestAEC001d_AdmittedDelta_EveryLegRequestCarriesTheConfirmation(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{f0Ids: []string{"B1", "B2"}})
	w.verdictFile("sync-audit-1.md", c.h1, aecN(1), "B3")

	out := w.call(aecCallOpts{
		deltaBase: c.h1,
		confirm:   aecClosure("B1", c.k1, "B2", "", "B3", c.k2),
		focus:     "concurrency",
	})
	wantScope(t, out, c.h1) // control: the complete confirmation is admitted

	for _, b := range []string{BackendCodex, BackendGLM} {
		focus := out.focusSeen[b]
		if !strings.Contains(focus, "concurrency") {
			t.Errorf("leg %s lost the caller's own focus: %q", b, focus)
		}
		for _, want := range []string{"B1", "B2", "B3", c.k1, c.k2} {
			if !strings.Contains(focus, want) {
				t.Errorf("leg %s request does not carry %q of the confirmation; focus = %q", b, want, focus)
			}
		}
	}
}
