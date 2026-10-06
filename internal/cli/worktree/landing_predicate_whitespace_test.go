package worktree

// landing_predicate_whitespace_test.go — SPEC-GITHUB-FLOW-CI-RESIDUE-001 M1,
// REQ-GFC-016 / AC-GFC-016. `git patch-id --stable` normalizes INTRA-LINE
// whitespace, so a card branch carrying an unlanded indentation- or
// trailing-space-only commit still reproduces the squash commit's cumulative
// patch-id — layer 2 declared the landing and the tree (the unlanded commit
// included) was disposed. Before declaring landed, the candidate commit must
// equal the card tip as a TREE, whitespace included.
//
// Cells B2/B3 reproduce the /tmp evidence of research.md §3.1 (squash then a
// whitespace-only commit on the card branch); the two guard cells pin the fix
// against over-refusal: whitespace that IS part of the squash still lands,
// and a whole added line (which patch-id already distinguishes) stays unlanded.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reindentLine changes the LEADING whitespace of "line <n>" in dir/f.txt.
func reindentLine(t *testing.T, dir string, n int, prefix string) {
	t.Helper()
	p := filepath.Join(dir, "f.txt")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	lines[n-1] = prefix + lines[n-1]
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
}

// trailSpaceLine appends trailing spaces to "line <n>" in dir/f.txt.
func trailSpaceLine(t *testing.T, dir string, n int, spaces string) {
	t.Helper()
	p := filepath.Join(dir, "f.txt")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	lines[n-1] = lines[n-1] + spaces
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
}

// addNewLine inserts a whole new line after "line <n>" in dir/f.txt — the
// B1 shape patch-id already distinguishes (research.md §3.1: not reproduced).
func addNewLine(t *testing.T, dir string, after int, text string) {
	t.Helper()
	p := filepath.Join(dir, "f.txt")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	out := append([]string{}, lines[:after]...)
	out = append(out, text)
	out = append(out, lines[after:]...)
	if err := os.WriteFile(p, []byte(strings.Join(out, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *gfdFixture) commitCard(t *testing.T, msg string) {
	t.Helper()
	landingGit(t, f.tree, "commit", "-q", "-a", "-m", msg)
}

// TestLandedByPatchIDWhitespaceOnlyCommitsAreNotLanded is AC-GFC-016: after a
// squash merge, an unlanded intra-line-whitespace commit on the card branch
// must NOT read as landed through layer 2 — the trees differ, whitespace
// included.
func TestLandedByPatchIDWhitespaceOnlyCommitsAreNotLanded(t *testing.T) {
	cells := []struct {
		name   string
		commit func(t *testing.T, f *gfdFixture)
	}{
		{"B2_indentation_commit_reproduces_the_patch_id", func(t *testing.T, f *gfdFixture) {
			// The whitespace change lands on the line the CARD ALREADY
			// modified (the research shape: world -> "   world") — only then
			// does the intra-line normalization make the cumulative diffs
			// collide.
			reindentLine(t, f.tree, 10, "   ")
			f.commitCard(t, "card: reindent the changed line")
		}},
		{"B3_trailing_space_commit_reproduces_the_patch_id", func(t *testing.T, f *gfdFixture) {
			trailSpaceLine(t, f.tree, 10, "  ")
			f.commitCard(t, "card: trail spaces on the changed line")
		}},
	}
	for _, c := range cells {
		t.Run(c.name, func(t *testing.T) {
			f := newGFDFixture(t)
			f.cardCommit(t, 10, "card")
			f.squash(t)
			f.pushMain(t)
			// Fixture precondition, measured: without the whitespace commit the
			// layer-2 answer IS landed — the defect is what the extra commit
			// fails to change.
			if ok, err := LandedByPatchID(f.repo, f.branch, "origin/main"); err != nil || !ok {
				t.Fatalf("fixture invalid: the plain squash must read landed (ok=%v err=%v)", ok, err)
			}
			c.commit(t, f)
			ok, err := LandedByPatchID(f.repo, f.branch, "origin/main")
			if err != nil {
				t.Fatalf("layer 2 must answer (nil error) on the whitespace case, got: %v", err)
			}
			if ok {
				t.Fatal("an unlanded whitespace-only commit must not declare the landing (AC-GFC-016)")
			}
		})
	}
}

// TestLandedByPatchIDWhitespaceInsideTheSquashStillLands guards the fix
// against over-refusal: when the whitespace IS part of the squash, the trees
// are equal and layer 2 still declares the landing.
func TestLandedByPatchIDWhitespaceInsideTheSquashStillLands(t *testing.T) {
	f := newGFDFixture(t)
	f.cardCommit(t, 10, "card")
	reindentLine(t, f.tree, 5, "   ")
	f.commitCard(t, "card: reindent line 5 (part of the card work)")
	f.squash(t)
	f.pushMain(t)
	ok, err := LandedByPatchID(f.repo, f.branch, "origin/main")
	if err != nil || !ok {
		t.Fatalf("a squash that carries the whitespace must still land (ok=%v err=%v)", ok, err)
	}
}

// TestLandedByPatchIDAddedLineStaysUnlanded is the B1 regression guard: a
// whole added line was never patch-id-equal (research.md §3.1) and must stay
// unlanded after the tree-equality guard.
func TestLandedByPatchIDAddedLineStaysUnlanded(t *testing.T) {
	f := newGFDFixture(t)
	f.cardCommit(t, 10, "card")
	f.squash(t)
	f.pushMain(t)
	addNewLine(t, f.tree, 10, "brand new line")
	f.commitCard(t, "card: add a line after the merge")
	ok, err := LandedByPatchID(f.repo, f.branch, "origin/main")
	if err != nil {
		t.Fatalf("layer 2 must answer (nil error) on the added-line case, got: %v", err)
	}
	if ok {
		t.Fatal("an added line must not declare the landing")
	}
}
