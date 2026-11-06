package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M3 RED — the default delta proposal and the
// "full" opt-out (AC-AEC-001b; REQ-AEC-001, REQ-AEC-002; card t2017).
//
// An audit_multi call that names a card_id, carries no delta_base and targets
// baseBranch proposes the audited_sha of the card's newest exported FAIL verdict
// file, and that proposal goes through the M1 chain gate exactly as an explicit
// SHA does. Otherwise it reviews the full scope and says why. delta_base "full"
// reviews the resolved base.
//
// The "newest" file is the rebind-evidence selection: highest -iter<N> in the
// name, then the latest modification time, applied to the sync-audit files when
// any exist and to the plan-audit files otherwise.
//
// ASSUMED WORDING. The SPEC requires the result to NAME the verdict file and SHA it
// proposed, or STATE why it proposed none, without fixing the text. The phrases
// the rows look for are in the constants block; change them there and nowhere else.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	dpPrefix       = "default delta_base:" // leads the note of a card call without delta_base
	dpProposed     = "proposed"            // "... proposed <sha> from <file>"
	dpNone         = "none"                // "... none - <reason>"
	dpNoVerdict    = "no audit verdict file"
	dpNotFail      = "is not a FAIL"
	dpNoSHA        = "no audited_sha"
	dpConflict     = "conflicting audited_sha"
	dpNotAncestor  = "not an ancestor of HEAD"
	dpFullRequest  = "full"
	dpGateRefusal  = deltaGatePrefix // M1: the gate's reason leads a refused delta
	dpSyncAuditFmt = "sync-audit-iter%d.md"
)

// verdictRaw writes a verdict file with the given body into the card evidence dir.
func (w *aecWorld) verdictRaw(name, body string) string {
	w.t.Helper()
	p := filepath.Join(w.f.EvidenceDir, name)
	w.f.Write(p, body)
	return p
}

func failBody(sha string, count int, ids ...string) string {
	b := "verdict: FAIL\naudited_sha: " + sha + "\n"
	b += "blocking_count: " + string(rune('0'+count)) + "\n"
	for _, id := range ids {
		b += "blocking_defect: " + id + "\n"
	}
	return b
}

// sideCommit returns a commit that is not an ancestor of HEAD.
func (w *aecWorld) sideCommit() string {
	w.t.Helper()
	return w.f.GitOut(w.f.CardDir, "commit-tree", w.head()+"^{tree}", "-p", w.head(), "-m", "side")
}

func (w *aecWorld) setMTime(path string, t time.Time) {
	w.t.Helper()
	if err := os.Chtimes(path, t, t); err != nil {
		w.t.Fatalf("chtimes: %v", err)
	}
}

// wantNoteHas asserts the result's note contains every phrase.
func wantNoteHas(t *testing.T, out aecOut, phrases ...string) {
	t.Helper()
	for _, p := range phrases {
		if !strings.Contains(out.result.ResidualRiskNote, p) {
			t.Errorf("the result does not say %q; residual_risk_note = %s", p, out.result.ResidualRiskNote)
		}
	}
}

// completeConfirmation is the confirmation of the happy chain plus one new id B3,
// closed by the later change k2.
func completeConfirmation(c aecChain) map[string]any {
	return aecClosure("B1", c.k1, "B3", c.k2)
}

// AC-AEC-001b — no delta is proposed, the full scope is reviewed, and the result
// says why: a first review (no verdict file), a PASS newest verdict, an
// audited_sha that is absent, conflicting or not an ancestor of HEAD.
func TestAEC001b_NoDefaultProposal_ReviewsFullScopeAndSaysWhy(t *testing.T) {
	rows := []struct {
		name  string
		setup func(w *aecWorld)
		why   string
	}{
		{"no verdict file at all (a first review)", func(w *aecWorld) {}, dpNoVerdict},
		{"the newest verdict is a PASS", func(w *aecWorld) {
			w.verdictRaw("sync-audit-iter1.md", "verdict: PASS\naudited_sha: "+w.head()+"\n")
		}, dpNotFail},
		{"a FAIL whose audited_sha is absent", func(w *aecWorld) {
			w.verdictRaw("sync-audit-iter1.md", "verdict: FAIL\nblocking_count: 0\n")
		}, dpNoSHA},
		{"a FAIL with conflicting audited_sha lines", func(w *aecWorld) {
			h := w.head()
			w.commit("later")
			w.verdictRaw("sync-audit-iter1.md", "verdict: FAIL\naudited_sha: "+h+"\naudited_sha: "+w.head()+"\nblocking_count: 0\n")
		}, dpConflict},
		{"a FAIL whose audited_sha is not an ancestor of HEAD", func(w *aecWorld) {
			w.verdictRaw("sync-audit-iter1.md", failBody(w.sideCommit(), 0))
		}, dpNotAncestor},
		{"the newest iteration is a PASS after an older FAIL", func(w *aecWorld) {
			h := w.head()
			w.commit("later")
			w.verdictRaw("sync-audit-iter1.md", failBody(h, 0))
			w.verdictRaw("sync-audit-iter2.md", "verdict: PASS\naudited_sha: "+w.head()+"\n")
		}, dpNotFail},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			w := newAECWorld(t)
			r.setup(w)
			out := w.call(aecCallOpts{})
			wantScope(t, out, "")
			wantNoteHas(t, out, dpPrefix+" "+dpNone, r.why)
		})
	}
}

// AC-AEC-001b — a FAIL verdict whose audited_sha is an ancestor of HEAD is
// proposed, and the result names the verdict file and the SHA. The proposal goes
// through the M1 chain gate: a card whose chain is unbroken and whose confirmation
// is complete reviews the delta.
func TestAEC001b_DefaultProposal_FromTheNewestFailVerdict_ReviewsTheDelta(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	w.verdictRaw("sync-audit-iter1.md", failBody(c.h1, 1, "B3"))

	out := w.call(aecCallOpts{confirm: completeConfirmation(c)}) // no delta_base
	wantScope(t, out, c.h1)
	wantNoteHas(t, out, dpPrefix+" "+dpProposed, c.h1, "sync-audit-iter1.md")
}

// AC-AEC-001b — the proposal is still gated: with no chain on disk the proposed
// SHA is named, and the gate sends the call to the full scope and says why.
func TestAEC001b_DefaultProposal_IsGated_NoChainReviewsFullScope(t *testing.T) {
	w := newAECWorld(t)
	h0 := w.head()
	w.commit("later change")
	w.verdictRaw("sync-audit-iter1.md", failBody(h0, 0))

	out := w.call(aecCallOpts{})
	wantScope(t, out, "")
	wantNoteHas(t, out, dpPrefix+" "+dpProposed, h0, "sync-audit-iter1.md", dpGateRefusal)
}

// AC-AEC-001b — the newest verdict is chosen by the rebind-evidence rule: the
// highest -iter<N> wins over a later modification time, and a sync-audit file
// decides over any plan-audit file.
func TestAEC001b_DefaultProposal_NewestVerdictFollowsTheRebindSelection(t *testing.T) {
	t.Run("the highest iteration wins over a later mtime", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{})
		older := w.verdictRaw("sync-audit-iter1.md", failBody(w.parent(c.h1), 0)) // an ancestor, would be a different proposal
		newer := w.verdictRaw("sync-audit-iter2.md", failBody(c.h1, 1, "B3"))
		w.setMTime(older, time.Now().Add(time.Hour)) // iter1 is the later file on disk
		w.setMTime(newer, time.Now().Add(-time.Hour))

		out := w.call(aecCallOpts{confirm: completeConfirmation(c)})
		wantScope(t, out, c.h1)
		wantNoteHas(t, out, "sync-audit-iter2.md", c.h1)
	})
	t.Run("a sync-audit file decides over the plan-audit files", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{}) // plan-audit-1.md FAIL at h0 already exists
		w.verdictRaw("sync-audit-iter1.md", failBody(c.h1, 1, "B3"))

		out := w.call(aecCallOpts{confirm: completeConfirmation(c)})
		wantScope(t, out, c.h1)
		wantNoteHas(t, out, "sync-audit-iter1.md")
	})
}

// AC-AEC-001b — delta_base "full" reviews the resolved base, exactly as a call
// without delta_base does today, even when a default would have been admitted.
// Today the string is rejected as "not a commit SHA".
func TestAEC001b_DeltaBaseFull_ReviewsTheResolvedBase(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	w.verdictRaw("sync-audit-iter1.md", failBody(c.h1, 1, "B3"))

	out := w.call(aecCallOpts{deltaBase: "full", confirm: completeConfirmation(c)})
	wantScope(t, out, "")
	wantNoteHas(t, out, dpPrefix+" "+dpNone, dpFullRequest)
}

// AC-AEC-001b — an explicit SHA replaces the file's. With a card_id it still goes
// through the gate: a broken chain (the 001c gap fixture) reviews the full scope.
// Control: passes with the M1 gate in place.
func TestAEC001b_ExplicitSHA_ReplacesTheFilesAndStillGoesThroughTheGate_Control(t *testing.T) {
	t.Run("a broken chain reviews the full scope", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{r1BaseParentOfH0: true})
		w.verdictRaw("sync-audit-iter1.md", failBody(c.h1, 1, "B3"))
		out := w.call(aecCallOpts{deltaBase: c.h1, confirm: completeConfirmation(c)})
		wantScope(t, out, "")
	})
	t.Run("a valid chain reviews the explicit SHA although the newest verdict is a PASS", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{})
		w.verdictRaw("sync-audit-iter1.md", "verdict: PASS\naudited_sha: "+c.h1+"\n")
		out := w.call(aecCallOpts{deltaBase: c.h1, confirm: c.call().confirm})
		wantScope(t, out, c.h1)
	})
}

// AC-AEC-001b — without a card_id nothing is proposed (the t1708 behavior), and a
// target other than baseBranch proposes nothing either. Control: passes today.
func TestAEC001b_NoDefaultWithoutACardID_OrForAnotherTarget_Control(t *testing.T) {
	t.Run("no card_id", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{})
		w.verdictRaw("sync-audit-iter1.md", failBody(c.h1, 1, "B3"))
		out := w.call(aecCallOpts{noCard: true})
		wantScope(t, out, "")
		if strings.Contains(out.result.ResidualRiskNote, dpPrefix) {
			t.Errorf("a call without card_id must not describe a default proposal: %s", out.result.ResidualRiskNote)
		}
	})
	t.Run("a target other than baseBranch", func(t *testing.T) {
		w := newAECWorld(t)
		c := w.happyChain(aecChainOpts{})
		w.verdictRaw("sync-audit-iter1.md", failBody(c.h1, 1, "B3"))
		out := w.call(aecCallOpts{extra: map[string]any{"target": codexTargetUncommitted}})
		wantScope(t, out, "")
		if strings.Contains(out.result.ResidualRiskNote, dpPrefix) {
			t.Errorf("a non-baseBranch call must not describe a default proposal: %s", out.result.ResidualRiskNote)
		}
	})
}
