package spec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// archive_git_test.go — the archive capability against a REAL git repository.
//
// archive_test.go injects the git seam, so it proves the eligibility logic but
// never touches git itself. These tests exercise the production path end to end:
// the exported PlanArchive / ExecuteArchive entry points with their real
// dependencies, the single-pass `git log --name-only` activity scan, and the
// `git mv` relocation. That is the half most likely to break — a format-string
// typo or a path-prefix slip would sail past every seam-injected test.

// gitRepo is a throwaway repository under t.TempDir().
type gitRepo struct {
	t   *testing.T
	dir string
}

func newGitRepo(t *testing.T) *gitRepo {
	t.Helper()

	r := &gitRepo{t: t, dir: t.TempDir()}
	r.run("git", "init")
	r.run("git", "config", "user.email", "test@example.com")
	r.run("git", "config", "user.name", "Archive Test")
	return r
}

func (r *gitRepo) run(name string, args ...string) {
	r.t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = r.dir
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

// commitSPEC writes a SPEC and commits it with a pinned committer date, so the
// activity scan has a deterministic timestamp to find.
func (r *gitRepo) commitSPEC(specID, status string, when time.Time) {
	r.t.Helper()

	dir := filepath.Join(r.dir, ".moai", "specs", specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		r.t.Fatalf("mkdir: %v", err)
	}

	body := fmt.Sprintf(`---
id: %s
title: "Git fixture"
version: "0.1.0"
status: %s
created: 2020-01-01
updated: 2020-01-01
author: test
priority: P1
phase: "v3.0.0"
module: "internal/spec"
lifecycle: spec-anchored
tags: "fixture"
---

# %s
`, specID, status, specID)

	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
		r.t.Fatalf("write spec.md: %v", err)
	}

	r.run("git", "add", filepath.Join(".moai", "specs", specID))
	r.commit(specID, when)
}

// touchSPEC makes a SECOND, distinct commit against an existing SPEC.
//
// The content must actually differ — a byte-identical rewrite leaves the tree
// clean and `git commit` refuses with "nothing to commit", which is exactly the
// trap that makes a "same SPEC committed twice" fixture silently useless.
func (r *gitRepo) touchSPEC(specID string, when time.Time) {
	r.t.Helper()

	dir := filepath.Join(r.dir, ".moai", "specs", specID)
	note := filepath.Join(dir, "progress.md")
	body := fmt.Sprintf("# progress\n\ntouched at %s\n", when.Format(time.RFC3339))

	if err := os.WriteFile(note, []byte(body), 0o644); err != nil {
		r.t.Fatalf("write progress.md: %v", err)
	}

	r.run("git", "add", filepath.Join(".moai", "specs", specID))
	r.commit(specID, when)
}

// commit records the staged tree with a pinned author/committer date.
func (r *gitRepo) commit(specID string, when time.Time) {
	r.t.Helper()

	stamp := when.Format(time.RFC3339)
	cmd := exec.Command("git", "commit", "-m", fmt.Sprintf("chore(%s): fixture", specID))
	cmd.Dir = r.dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE="+stamp,
		"GIT_COMMITTER_DATE="+stamp,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("git commit: %v\n%s", err, out)
	}
}

// tracked reports whether a path is in the git index (i.e. survived as tracked).
//
// The argument to `git ls-files --error-unmatch` is a PATHSPEC, which git
// matches against its own forward-slash paths — and where a backslash is a glob
// escape character, not a separator. Callers build paths with filepath.Join for
// the os.Stat assertions, so convert here rather than making every call site
// remember it.
func (r *gitRepo) tracked(path string) bool {
	r.t.Helper()

	cmd := exec.Command("git", "ls-files", "--error-unmatch", filepath.ToSlash(path))
	cmd.Dir = r.dir
	return cmd.Run() == nil
}

// TestGitLastActivity_RealRepo pins the single-pass activity scan against real git
// output — the format string, the record separator, and the path-prefix parsing all
// have to line up for this to pass.
func TestGitLastActivity_RealRepo(t *testing.T) {
	r := newGitRepo(t)

	old := time.Date(2022, 4, 5, 12, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC)

	r.commitSPEC("SPEC-GITOLD-001", "completed", old)
	r.commitSPEC("SPEC-GITNEW-001", "completed", recent)

	activity := gitLastActivity(r.dir)

	if len(activity) != 2 {
		t.Fatalf("want 2 SPECs in the activity map, got %d: %v", len(activity), activity)
	}
	if got := activity["SPEC-GITOLD-001"]; !got.Equal(old) {
		t.Errorf("SPEC-GITOLD-001 = %v, want %v", got, old)
	}
	if got := activity["SPEC-GITNEW-001"]; !got.Equal(recent) {
		t.Errorf("SPEC-GITNEW-001 = %v, want %v", got, recent)
	}
}

// TestGitLastActivity_NewestCommitWins: a SPEC touched twice reports its NEWEST
// touch. Reporting the oldest would archive SPECs that are still being worked on.
func TestGitLastActivity_NewestCommitWins(t *testing.T) {
	r := newGitRepo(t)

	first := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	second := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	r.commitSPEC("SPEC-TWICE-001", "completed", first)
	// Touch the same SPEC again, later, with a real content change.
	r.touchSPEC("SPEC-TWICE-001", second)

	got := gitLastActivity(r.dir)["SPEC-TWICE-001"]
	if !got.Equal(second) {
		t.Errorf("SPEC-TWICE-001 = %v, want the NEWEST touch %v", got, second)
	}
}

func TestGitLastActivity_NonGitDirIsEmpty(t *testing.T) {
	t.Parallel()

	if got := gitLastActivity(t.TempDir()); len(got) != 0 {
		t.Fatalf("a non-git directory must yield an empty activity map, got %v", got)
	}
}

// TestPlanArchive_RealRepo exercises the exported PlanArchive with its real
// dependencies (the CLI's actual entry point).
func TestPlanArchive_RealRepo(t *testing.T) {
	r := newGitRepo(t)

	stale := time.Now().AddDate(0, 0, -300)
	fresh := time.Now().AddDate(0, 0, -5)

	r.commitSPEC("SPEC-RSTALE-001", "completed", stale)
	r.commitSPEC("SPEC-RFRESH-001", "completed", fresh)
	r.commitSPEC("SPEC-RACTIVE-001", "in-progress", stale)

	plan, err := PlanArchive(r.dir, ArchiveOptions{GraceDays: 90})
	if err != nil {
		t.Fatalf("PlanArchive: %v", err)
	}

	if len(plan.Candidates) != 1 || plan.Candidates[0].SPECID != "SPEC-RSTALE-001" {
		t.Fatalf("want only SPEC-RSTALE-001 eligible, got %v", candidateIDs(plan))
	}
	if plan.Candidates[0].ActivitySource != ActivitySourceGit {
		t.Errorf("ActivitySource = %q, want %q (a git-tracked SPEC must resolve via git)",
			plan.Candidates[0].ActivitySource, ActivitySourceGit)
	}
	if plan.Scanned != 3 {
		t.Errorf("Scanned = %d, want 3", plan.Scanned)
	}

	// Planning is observation-only, even on the real path.
	if _, err := os.Stat(filepath.Join(r.dir, ".moai", "archive")); !os.IsNotExist(err) {
		t.Error("PlanArchive must not create the archive tree")
	}
}

// TestExecuteArchive_RealRepo_StaysTracked is the git-tracked + grep-discoverable
// guarantee (REQ-SSP-007 / AC-SSP-023): `git mv` records the relocation as a
// rename, so the archived SPEC stays in the index rather than becoming an
// untracked orphan.
func TestExecuteArchive_RealRepo_StaysTracked(t *testing.T) {
	r := newGitRepo(t)

	stale := time.Date(2023, 8, 9, 0, 0, 0, 0, time.UTC)
	r.commitSPEC("SPEC-TRACKED-001", "completed", stale)

	plan, err := ExecuteArchive(r.dir, ArchiveOptions{GraceDays: 90})
	if err != nil {
		t.Fatalf("ExecuteArchive: %v", err)
	}
	if len(plan.Candidates) != 1 {
		t.Fatalf("want 1 archived SPEC, got %v", candidateIDs(plan))
	}

	src := filepath.Join(".moai", "specs", "SPEC-TRACKED-001", "spec.md")
	dst := filepath.Join(".moai", "archive", "specs", "2023", "SPEC-TRACKED-001", "spec.md")

	if _, err := os.Stat(filepath.Join(r.dir, src)); !os.IsNotExist(err) {
		t.Error("the SPEC should no longer exist under .moai/specs")
	}
	if _, err := os.Stat(filepath.Join(r.dir, dst)); err != nil {
		t.Fatalf("archived SPEC missing at %s: %v", dst, err)
	}

	// The load-bearing assertion: still in the git index at its new path.
	if !r.tracked(dst) {
		t.Error("the archived SPEC must remain git-tracked at its new path (git mv, not a bare move)")
	}
	if r.tracked(src) {
		t.Error("the old path must no longer be tracked")
	}
}

// TestGitMoveOrRename_FallsBackOutsideGit: with no repository, `git mv` cannot
// apply and the mover degrades to a plain rename rather than failing.
func TestGitMoveOrRename_FallsBackOutsideGit(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	src := filepath.Join(".moai", "specs", "SPEC-NOGIT-001")
	dst := filepath.Join(".moai", "archive", "specs", "2024", "SPEC-NOGIT-001")

	if err := os.MkdirAll(filepath.Join(base, src), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, src, "spec.md"), []byte("body"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := gitMoveOrRename(base, src, dst); err != nil {
		t.Fatalf("gitMoveOrRename outside a repo must fall back to rename, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(base, dst, "spec.md")); err != nil {
		t.Errorf("file did not land at the destination: %v", err)
	}
}

// TestGitMoveOrRename_RefusesToClobber: an occupied destination is an error, never
// a silent overwrite. Archiving must not be able to destroy an existing SPEC.
func TestGitMoveOrRename_RefusesToClobber(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	src := filepath.Join(".moai", "specs", "SPEC-CLOBBER-001")
	dst := filepath.Join(".moai", "archive", "specs", "2024", "SPEC-CLOBBER-001")

	for _, d := range []string{src, dst} {
		if err := os.MkdirAll(filepath.Join(base, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(base, dst, "spec.md"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := gitMoveOrRename(base, src, dst); err == nil {
		t.Fatal("an existing destination must be refused, not overwritten")
	}

	body, err := os.ReadFile(filepath.Join(base, dst, "spec.md"))
	if err != nil || string(body) != "existing" {
		t.Error("the pre-existing archived SPEC must survive untouched")
	}
}
