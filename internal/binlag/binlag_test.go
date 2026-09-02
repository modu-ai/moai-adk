package binlag

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// twoCommitRepo builds a throwaway git repository holding two commits in an
// ancestor relation and returns (dir, ancestorSHA, descendantSHA).
func twoCommitRepo(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("a.txt", "first")
	run("add", "a.txt")
	run("commit", "-q", "-m", "first")
	ancestor := revParse(t, dir, "HEAD")
	write("b.txt", "second")
	run("add", "b.txt")
	run("commit", "-q", "-m", "second")
	descendant := revParse(t, dir, "HEAD")
	return dir, ancestor, descendant
}

func revParse(t *testing.T, dir, ref string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", ref).Output()
	if err != nil {
		t.Fatalf("rev-parse %s: %v", ref, err)
	}
	return strings.TrimSpace(string(out))
}

// AC-BLV-002 control group at the comparer level: a binary built from the
// tree's own HEAD is fresh, so no advisory text is derived from it.
func TestEvaluate_BinaryAtHeadIsFresh(t *testing.T) {
	dir, _, head := twoCommitRepo(t)
	v := Evaluate(context.Background(), Request{Dir: dir, BinaryCommit: head})
	if v.Status != StatusFresh {
		t.Fatalf("status = %q, want %q", v.Status, StatusFresh)
	}
	if adv := Advisory(v); adv != "" {
		t.Errorf("Advisory on a fresh verdict = %q, want empty", adv)
	}
}

// AC-BLV-001 at the comparer level: an ancestor build is behind, and the
// advisory names BOTH the binary commit and the tree HEAD plus the remedy.
func TestEvaluate_AncestorBinaryIsBehind(t *testing.T) {
	dir, ancestor, head := twoCommitRepo(t)
	v := Evaluate(context.Background(), Request{Dir: dir, BinaryCommit: ancestor})
	if v.Status != StatusBehind {
		t.Fatalf("status = %q, want %q", v.Status, StatusBehind)
	}
	adv := Advisory(v)
	if adv == "" {
		t.Fatal("Advisory on a behind verdict is empty")
	}
	for _, want := range []string{ancestor[:9], head[:9], "make build && make install"} {
		if !strings.Contains(adv, want) {
			t.Errorf("advisory does not name %q:\n%s", want, adv)
		}
	}
}

// A build from a sibling branch is NOT behind — preserving the existing
// leniency (spec §5: reducing it is a downstream regression).
func TestEvaluate_NonAncestorIsDivergentNotBehind(t *testing.T) {
	dir, _, _ := twoCommitRepo(t)
	// A commit that exists nowhere in this tree stands in no ancestor relation.
	v := Evaluate(context.Background(), Request{Dir: dir, BinaryCommit: "0123456789abcdef0123456789abcdef01234567"})
	if v.Status == StatusBehind {
		t.Fatalf("status = %q; an unknown commit must not be reported behind", v.Status)
	}
	if adv := Advisory(v); adv != "" {
		t.Errorf("Advisory on a non-behind verdict = %q, want empty", adv)
	}
}

// AC-BLV-003: a directory that is not a git working tree is not-applicable —
// never Fail, and it derives no advisory.
func TestEvaluate_NonGitDirectoryIsNotApplicable(t *testing.T) {
	dir := t.TempDir()
	v := Evaluate(context.Background(), Request{Dir: dir, BinaryCommit: "0123456789abcdef0123456789abcdef01234567"})
	if v.Status != StatusNotApplicable {
		t.Fatalf("status = %q, want %q", v.Status, StatusNotApplicable)
	}
	if adv := Advisory(v); adv != "" {
		t.Errorf("Advisory in a non-git directory = %q, want empty", adv)
	}
}

// A binary carrying no commit metadata cannot be compared at all.
func TestEvaluate_MissingCommitMetadataIsNotApplicable(t *testing.T) {
	dir, _, _ := twoCommitRepo(t)
	for _, commit := range []string{"", "none", "unknown"} {
		v := Evaluate(context.Background(), Request{Dir: dir, BinaryCommit: commit})
		if v.Status != StatusNotApplicable {
			t.Errorf("commit %q: status = %q, want %q", commit, v.Status, StatusNotApplicable)
		}
	}
}

// AC-BLV-003 subdirectory anchor: running from below the tree root must not
// flip an applicable tree to not-applicable.
func TestEvaluate_SubdirectoryStaysApplicable(t *testing.T) {
	dir, ancestor, _ := twoCommitRepo(t)
	sub := filepath.Join(dir, "nested", "deeper")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	v := Evaluate(context.Background(), Request{Dir: sub, BinaryCommit: ancestor})
	if v.Status != StatusBehind {
		t.Fatalf("from subdirectory: status = %q, want %q", v.Status, StatusBehind)
	}
}

// AC-BLV-007: the verdict rests on commit ancestry, never on the version
// string. Reproduces the inversion measured on this card — the DESCENDANT
// build reports the LOWER semver and must still read as up to date, while the
// ANCESTOR build reporting the HIGHER semver must read as behind.
func TestEvaluate_VersionStringDoesNotDecideTheVerdict(t *testing.T) {
	dir, ancestor, descendant := twoCommitRepo(t)

	newer := Evaluate(context.Background(), Request{
		Dir: dir, BinaryCommit: descendant, BinaryVersion: "v3.1.2",
	})
	if newer.Status == StatusBehind {
		t.Errorf("descendant build with the lower semver was reported behind (status %q)", newer.Status)
	}

	older := Evaluate(context.Background(), Request{
		Dir: dir, BinaryCommit: ancestor, BinaryVersion: "v3.1.3-rc.5",
	})
	if older.Status != StatusBehind {
		t.Errorf("ancestor build with the higher semver: status = %q, want %q", older.Status, StatusBehind)
	}
}

// AC-BLV-005 (the seam half that lives here): Evaluate dispatches through the
// single substitutable Comparer, so a stub replaces the real comparison.
func TestEvaluate_DispatchesThroughTheComparerSeam(t *testing.T) {
	orig := Comparer
	t.Cleanup(func() { Comparer = orig })
	Comparer = func(context.Context, Request) Verdict {
		return Verdict{Status: StatusBehind, BinaryCommit: "aaaaaaaaa", SourceHead: "bbbbbbbbb"}
	}
	v := Evaluate(context.Background(), Request{Dir: t.TempDir()})
	if v.Status != StatusBehind || v.BinaryCommit != "aaaaaaaaa" {
		t.Fatalf("Evaluate did not route through Comparer: %+v", v)
	}
}
