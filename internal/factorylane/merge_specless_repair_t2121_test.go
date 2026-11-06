package factorylane

// merge_specless_repair_t2121_test.go — card t2121 repair round: the absorb
// tolerance must refuse an absorb merge that carries an edit the audit never
// read. Four holes the sync audit reproduced: a binary (or -diff) card file
// whose content is swapped inside the merge, a card hunk moved to another
// position, an octopus merge, and a conflict-resolved merge. A clean absorb
// where main shifts line numbers must still pass. Every fixture is a real
// throwaway git repository.

import (
	"os/exec"
	"strings"
	"testing"
)

// absorbWithEdit merges main into the card branch without committing, lets edit
// rewrite the working tree, and commits the result as the absorb merge.
func (r specLessRepo) absorbWithEdit(t *testing.T, edit func()) {
	t.Helper()
	cmd := exec.Command("git", "merge", "-q", "--no-ff", "--no-commit", "main")
	cmd.Dir = r.dir
	_ = cmd.Run() // a conflict exits 1; the edit callback resolves it
	edit()
	r.git(t, "add", "-A")
	r.git(t, "commit", "-q", "-m", "absorb main with edit")
}

// newCodeRepo builds a card whose audited commit changes code.txt, which main
// already carries; it returns the repo with the verdict committed.
func newCodeRepo(t *testing.T, mainBody, cardBody string) specLessRepo {
	t.Helper()
	r, _ := newSpecLessRepo(t)
	r.mainAdvance(t, "code.txt", mainBody)
	r.absorbMain(t)
	r.write(t, "code.txt", cardBody)
	head := r.commit(t, "card edits code.txt")
	r.write(t, r.verdictRel(), verdictBody("PASS", head))
	r.commit(t, "evidence")
	return r
}

func TestT2121AbsorbRepair(t *testing.T) {
	t.Run("P1_binary_card_file_swapped_in_merge_refuses", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		r.write(t, "bin.dat", "\x00\x01\x02A")
		head := r.commit(t, "card binary")
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbWithEdit(t, func() { r.write(t, "bin.dat", "\x00\x01\x02B") })
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("P7_minus_diff_attribute_text_file_edited_in_merge_refuses", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		r.write(t, ".gitattributes", "card.txt -diff\n")
		head := r.commit(t, "mark card.txt -diff")
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbWithEdit(t, func() { r.write(t, "card.txt", "card EVIL\n") })
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("P2b_card_hunk_moved_in_merge_refuses", func(t *testing.T) {
		r := newCodeRepo(t, "a\nb\nc\nd\ne\n", "a\nguard\nb\nc\nd\ne\n")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbWithEdit(t, func() { r.write(t, "code.txt", "a\nb\nc\nd\nguard\ne\n") })
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("P10_clean_absorb_with_shifted_line_numbers_passes", func(t *testing.T) {
		r := newCodeRepo(t, "a\nb\nc\nd\ne\n", "a\nb\nc\nd\nguard\ne\n")
		r.mainAdvance(t, "code.txt", "top1\ntop2\na\nb\nc\nd\ne\n")
		r.absorbMain(t)
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "verdict: PASS")
	})
	t.Run("octopus_merge_refuses", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.git(t, "checkout", "-q", "-b", "m2", "main")
		r.write(t, "b.txt", "b\n")
		r.commit(t, "m2 work")
		r.git(t, "checkout", "-q", "main")
		r.write(t, "c.txt", "c\n")
		r.commit(t, "main work")
		r.git(t, "merge", "-q", "--no-ff", "-m", "main takes m2", "m2")
		first := r.git(t, "rev-parse", "main^1")
		r.git(t, "checkout", "-q", "WT-card")
		r.git(t, "merge", "-q", "--no-ff", "-m", "octopus", first, "m2")
		if n := len(strings.Fields(r.git(t, "rev-list", "--parents", "-n", "1", "HEAD"))) - 1; n != 3 {
			t.Fatalf("fixture is not an octopus: HEAD has %d parents", n)
		}
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "octopus")
	})
	t.Run("conflict_resolved_merge_refuses", func(t *testing.T) {
		r := newCodeRepo(t, "a\nb\nc\nd\ne\n", "a\nB\nc\nd\ne\n")
		r.mainAdvance(t, "code.txt", "a\nb\nC\nd\ne\n")
		r.absorbWithEdit(t, func() { r.write(t, "code.txt", "a\nB\nC\nd\ne\n") })
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
}

// merge-tree --write-tree needs git 2.38; a runner whose git rejects it must
// make the absorb check refuse and say why, never pass.
type noMergeTreeGit struct{ inner GitRunner }

func (g noMergeTreeGit) Git(args ...string) (string, error) {
	if len(args) > 0 && args[0] == "merge-tree" {
		return "", &GitExitError{ExitCode: 129, Stderr: "usage: git merge-tree [<options>] <branch1> <branch2>"}
	}
	return g.inner.Git(args...)
}

func TestT2121AbsorbRepairUnsupportedGitFailsClosed(t *testing.T) {
	r, head := newSpecLessRepo(t)
	r.write(t, r.verdictRel(), verdictBody("PASS", head))
	r.commit(t, "evidence")
	r.mainAdvance(t, "main_only.txt", "m\n")
	r.absorbMain(t)
	run, err := EvaluateMergeTriple(MergeTripleInput{
		Lane: "lane-1", Card: r.card, Branch: "WT-card", Develop: "main", RepoDir: r.dir,
	}, noMergeTreeGit{inner: ExecGitRunner{Dir: r.dir}})
	if err != nil {
		t.Fatal(err)
	}
	wantSyncAudit(t, run, false, "2.38")
}
