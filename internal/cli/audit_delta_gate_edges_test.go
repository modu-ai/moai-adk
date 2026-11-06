package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M1 — edge cases of the closure confirmation
// and of the evidence-directory reader (REQ-AEC-004 (ii); card t2017).
//
// The gate reads every .md file of the card evidence directory, so these rows pin
// what it ignores (files that cannot be tied to the chain) and what it refuses
// (a file it cannot read in full: an unreadable closure set must never read as an
// empty one).

import (
	"path/filepath"
	"strings"
	"testing"
)

// A closing commit that is missing, not a commit of this tree, or not a SHA is
// not a closure: the call is reviewed at the full scope.
func TestAEC001d_ClosingCommitEdges_ReviewFullScope(t *testing.T) {
	rows := map[string]string{
		"an empty closing commit for an id new since the previous link": "",
		"a well-formed hex that is no commit of the tree":               "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
	}
	for name, commit := range rows {
		t.Run(name, func(t *testing.T) {
			w := newAECWorld(t)
			c := w.happyChain(aecChainOpts{})
			w.verdictFile("sync-audit-1.md", c.h1, aecN(1), "B3")
			out := w.call(aecCallOpts{deltaBase: c.h1, confirm: aecClosure("B1", c.k1, "B3", commit)})
			wantScope(t, out, "")
			if !strings.Contains(out.result.ResidualRiskNote, "B3") {
				t.Errorf("the refusal must name the id it refused: %s", out.result.ResidualRiskNote)
			}
		})
	}
}

// closingCommitRefusal is reachable with a non-SHA only below the handler (which
// validates the argument); the gate must still refuse it on its own.
func TestClosingCommitRefusal_NonSHAIsRefused(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	for _, bad := range []string{"--output=x", "HEAD", "B1", "xyz"} {
		if why := closingCommitRefusal(w.tree, c.h2, c.h1, c.h1, "B3", bad); why == "" {
			t.Errorf("closingCommitRefusal(%q) admitted a non-SHA", bad)
		}
	}
}

// Evidence files that cannot be tied to the chain decide nothing: a non-.md
// file, a directory named like a verdict, a PASS file, a FAIL file with no
// audited_sha, a FAIL file with two conflicting audited_sha lines, and a FAIL
// file audited at a commit that heads no chain record. None of their ids joins
// the closure set, so the confirmation naming only B1 is still complete.
func TestAEC001d_EvidenceFilesNotTiedToTheChain_AreIgnored(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	other := w.parent(c.h0) // a commit that heads no chain record
	dir := w.f.EvidenceDir

	w.f.Write(filepath.Join(dir, "notes.txt"), "verdict: FAIL\naudited_sha: "+c.h0+"\nblocking_count: 1\nblocking_defect: B9\n")
	w.f.Write(filepath.Join(dir, "subdir.md", "inner.md"), "x\n")
	w.f.Write(filepath.Join(dir, "pass-1.md"), "verdict: PASS\naudited_sha: "+c.h0+"\nblocking_count: 1\nblocking_defect: B9\n")
	w.f.Write(filepath.Join(dir, "no-sha-1.md"), "verdict: FAIL\nblocking_count: 1\nblocking_defect: B9\n")
	w.f.Write(filepath.Join(dir, "two-sha-1.md"), "verdict: FAIL\naudited_sha: "+c.h0+"\naudited_sha: "+c.h1+"\nblocking_count: 1\nblocking_defect: B9\n")
	w.verdictFile("unrelated-1.md", other, aecN(1), "B9")

	out := w.call(c.call()) // confirms B1 only
	wantScope(t, out, c.h1)
}

// A file the gate cannot read in full is not an empty closure set: it refuses.
// Today's reader skips it, which would let an oversized FAIL file hide its ids.
func TestAEC001d_UnreadableEvidenceFile_RefusesTheDelta(t *testing.T) {
	w := newAECWorld(t)
	c := w.happyChain(aecChainOpts{})
	body := "verdict: FAIL\naudited_sha: " + c.h1 + "\nblocking_count: 1\nblocking_defect: B7\n" +
		strings.Repeat("x", maxVerdictFileBytes+16)
	w.f.Write(filepath.Join(w.f.EvidenceDir, "sync-audit-1.md"), body)

	out := w.call(c.call())
	wantScope(t, out, "")
	if !strings.Contains(out.result.ResidualRiskNote, "sync-audit-1.md") {
		t.Errorf("the refusal must name the unreadable file: %s", out.result.ResidualRiskNote)
	}
}
