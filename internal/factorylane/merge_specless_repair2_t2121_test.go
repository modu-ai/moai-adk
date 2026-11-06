package factorylane

// merge_specless_repair2_t2121_test.go — card t2121 repair round 2. An absorb
// merge that DISCARDS main's change (`-X ours`, or deleting a file main added)
// leaves the HEAD tree equal to the audited tree, so the audited..HEAD path list
// is empty; the absorb check must still run because a merge commit sits in
// audited..HEAD. Also pins the `--binary` and `--no-textconv` flags of the card
// diff, which only a merge driver injecting content can exercise. Every fixture
// is a real throwaway git repository.

import (
	"runtime"
	"testing"
)

func TestT2121AbsorbRepair2(t *testing.T) {
	t.Run("N8_X_ours_absorb_dropping_main_change_refuses", func(t *testing.T) {
		r := newCodeRepo(t, "a\nb\nc\nd\ne\n", "a\nB\nc\nd\ne\n")
		r.mainAdvance(t, "code.txt", "a\nBMAIN\nc\nd\ne\n")
		r.git(t, "merge", "-q", "--no-ff", "-X", "ours", "-m", "absorb main keeping ours", "main")
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("N4c_absorb_deleting_a_file_main_added_refuses", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbWithEdit(t, func() { r.git(t, "rm", "-q", "-f", "main_only.txt") })
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("N2a_clean_absorb_still_passes", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbMain(t)
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "verdict: PASS")
	})
	t.Run("N3a_two_clean_absorbs_still_pass", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_one.txt", "1\n")
		r.absorbMain(t)
		r.mainAdvance(t, "main_two.txt", "2\n")
		r.absorbMain(t)
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "verdict: PASS")
	})
	t.Run("regression_no_merge_evidence_dir_only_passes", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.write(t, ".moai/reports/"+r.card+"/note.md", "note\n")
		r.commit(t, "more evidence")
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "verdict: PASS")
	})
}

// injectingMergeDriver wires a merge driver that overwrites the merged file with
// payload, so an absorb merge carries content neither parent holds while a
// re-merge with the same repository config reproduces it.
func (r specLessRepo) injectingMergeDriver(t *testing.T, file, payload string) {
	t.Helper()
	r.git(t, "config", "merge.evil.driver", "printf '"+payload+"' > %A")
}

// binaryInjectRepo builds a card that edits a binary file main also edits, then
// absorbs main under a merge driver that injects other binary content. Without
// --binary the card diff prints only "Binary files differ" for both the audited
// commit and HEAD, so only the content in the patch can tell them apart.
func binaryInjectRepo(t *testing.T) specLessRepo {
	t.Helper()
	r, _ := newSpecLessRepo(t)
	r.mainAdvance(t, "bin.dat", "\x00v0")
	r.mainAdvance(t, ".gitattributes", "bin.dat merge=evil\n")
	r.absorbMain(t)
	r.write(t, "bin.dat", "\x00ORIG")
	head := r.commit(t, "card edits binary")
	r.write(t, r.verdictRel(), verdictBody("PASS", head))
	r.commit(t, "evidence")
	r.mainAdvance(t, "bin.dat", "\x00MAIN")
	r.injectingMergeDriver(t, "bin.dat", `\000EVIL`)
	r.absorbMain(t)
	return r
}

func TestT2121AbsorbRepair2BinaryPinned(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("merge driver uses a POSIX shell command")
	}
	r := binaryInjectRepo(t)
	if got := r.git(t, "show", "HEAD:bin.dat"); got == "\x00MAIN" || got == "\x00ORIG" {
		t.Fatalf("fixture did not inject content: %q", got)
	}
	wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
}

// textconvInjectRepo is the same shape with a text file whose textconv driver
// prints a constant, so a diff that applies the textconv shows no change.
func textconvInjectRepo(t *testing.T) specLessRepo {
	t.Helper()
	r, _ := newSpecLessRepo(t)
	r.mainAdvance(t, "t.txt", "v0\n")
	r.mainAdvance(t, ".gitattributes", "t.txt merge=evil diff=hide\n")
	r.absorbMain(t)
	r.git(t, "config", "diff.hide.textconv", "printf hidden; :")
	r.write(t, "t.txt", "ORIG\n")
	head := r.commit(t, "card edits t.txt")
	r.write(t, r.verdictRel(), verdictBody("PASS", head))
	r.commit(t, "evidence")
	r.mainAdvance(t, "t.txt", "MAIN\n")
	r.injectingMergeDriver(t, "t.txt", `EVIL\n`)
	r.absorbMain(t)
	return r
}

func TestT2121AbsorbRepair2TextconvPinned(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("merge driver and textconv use a POSIX shell command")
	}
	r := textconvInjectRepo(t)
	if got := r.git(t, "show", "HEAD:t.txt"); got != "EVIL" {
		t.Fatalf("fixture did not inject content: %q", got)
	}
	wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
}
