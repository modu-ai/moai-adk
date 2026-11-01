package cli

// Reports-archive action tests — SPEC-REPORTS-LIFECYCLE-001 REQ-RLC-007
// (AC-RLC-009 / AC-RLC-010).
//
// The action is move-only (no code path deletes reports content) and runs on
// a default-deny predicate: a top-level entry under .moai/reports/ archives
// only when ALL hold — evidence-shaped name (t<digits> or
// SPEC-<DOMAIN>-<NNN>), mtime older than the retention window, and zero
// git-tracked files. Explicitly protected entries (historical/, plan-audit/,
// worktrees/, archive/) are out of scope by construction, and any entry
// containing tracked files is auto-protected regardless of its name.
//
// The tracked-files probe is injected per test (reportsTrackedFilesFunc) so
// the predicate is exercised deterministically without a real git index; the
// default implementation shells out to `git ls-files`.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
)

// stubTrackedProbe swaps reportsTrackedFilesFunc for the test.
func stubTrackedProbe(t *testing.T, tracked func(rel string) bool) {
	t.Helper()
	orig := reportsTrackedFilesFunc
	reportsTrackedFilesFunc = func(root, rel string) (bool, error) {
		return tracked(rel), nil
	}
	t.Cleanup(func() { reportsTrackedFilesFunc = orig })
}

// setupReportsScaffold lays out the AC-RLC-009/010 fixture tree under a temp
// root and returns the root. Entries are backdated to be archive-eligible.
func setupReportsScaffold(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	reports := filepath.Join(root, ".moai", "reports")
	old := time.Now().AddDate(0, 0, -config.DefaultReportsArchiveRetentionDays-1)
	backdate := func(rel string) {
		path := filepath.Join(reports, filepath.FromSlash(rel))
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatalf("backdate %s: %v", rel, err)
		}
	}
	mustWrite := func(rel, content string) {
		path := filepath.Join(reports, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	// Eligible candidate: old, evidence-shaped (t<digits>), no tracked files.
	// AC-RLC-009 names this fixture t-old illustratively; the predicate's
	// name shape is load-bearing, so the conforming realization is t100.
	mustWrite("t100/evidence.md", "old card evidence\n")
	backdate("t100")
	// Same shape but fresh: mtime inside the window (t-new in the AC).
	mustWrite("t101/evidence.md", "fresh evidence\n")
	// Wrong name shape: not evidence-shaped at all.
	mustWrite("random-notes/note.md", "notes\n")
	backdate("random-notes")
	// Explicitly protected entries (AC-RLC-010).
	mustWrite("historical/kept.md", "historical evidence\n")
	mustWrite("plan-audit/.gitkeep", "")
	mustWrite("worktrees/wt-slug/evidence.md", "hoisted evidence\n")
	mustWrite("archive/2026-08/t-older/evidence.md", "already archived\n")
	// Tracked fixture: evidence-shaped, old — protected by the tracked rule.
	mustWrite("t338/ac-count-baseline.txt", "tracked fixture\n")
	backdate("t338")

	return root
}

func reportsPath(root, rel string) string {
	return filepath.Join(root, ".moai", "reports", filepath.FromSlash(rel))
}

// TestCleanReportsArchive_MovesOnlyEligibleCandidates (AC-RLC-009): the old
// evidence-shaped candidate moves into archive/<YYYY-MM>/, the fresh and
// wrong-named entries stay put, and the output reports count + bytes.
func TestCleanReportsArchive_MovesOnlyEligibleCandidates(t *testing.T) {
	root := setupReportsScaffold(t)
	stubTrackedProbe(t, func(rel string) bool { return false })

	p := printer.New(printer.WithWriters(&strings.Builder{}, &strings.Builder{}))
	if err := runCleanReportsArchiveWithRoot(p, true, config.DefaultReportsArchiveRetentionDays, root); err != nil {
		t.Fatalf("reports-archive: %v", err)
	}

	// t100 moved (original gone; t-old in the AC's illustrative naming).
	if _, err := os.Stat(reportsPath(root, "t100")); !os.IsNotExist(err) {
		t.Errorf("eligible candidate t100 must have moved out of .moai/reports/ (stat err=%v)", err)
		entries, _ := os.ReadDir(reportsPath(root, ""))
		for _, e := range entries {
			t.Errorf("  reports/ entry: %s", e.Name())
		}
	}
	// A NEW shard (the candidate's backdated mtime month) holds the moved
	// content; the scaffold's pre-existing 2026-08 shard is fixture noise.
	moved, err := os.ReadDir(reportsPath(root, "archive"))
	if err != nil {
		t.Fatalf("archive dir: %v", err)
	}
	var shard string
	for _, e := range moved {
		if e.IsDir() && e.Name() != "2026-08" {
			shard = e.Name()
		}
	}
	if shard == "" {
		t.Fatal("no new YYYY-MM shard created under archive/")
	}
	if !strings.Contains(shard, "20") {
		t.Errorf("shard %q is not a YYYY-MM directory", shard)
	}
	if _, err := os.Stat(filepath.Join(reportsPath(root, "archive"), shard, "t100", "evidence.md")); err != nil {
		t.Errorf("moved candidate content missing under archive/%s: %v", shard, err)
	}

	// Non-candidates untouched.
	for _, rel := range []string{"t101", "random-notes"} {
		if _, statErr := os.Stat(reportsPath(root, rel)); statErr != nil {
			t.Errorf("non-candidate %s was affected", rel)
		}
	}
}

// TestCleanReportsArchive_DryRunMovesNothing mirrors `moai clean`'s
// report-only default: without --force nothing moves and the candidate count
// is still reported.
func TestCleanReportsArchive_DryRunMovesNothing(t *testing.T) {
	root := setupReportsScaffold(t)
	stubTrackedProbe(t, func(rel string) bool { return false })

	p := printer.New(printer.WithWriters(&strings.Builder{}, &strings.Builder{}))
	if err := runCleanReportsArchiveWithRoot(p, false, config.DefaultReportsArchiveRetentionDays, root); err != nil {
		t.Fatalf("reports-archive dry-run: %v", err)
	}
	if _, err := os.Stat(reportsPath(root, "t100")); err != nil {
		t.Error("dry-run must not move the candidate")
	}
	// No NEW shard may appear (the scaffold's pre-existing archive/2026-08 is
	// part of the fixture, not an action output).
	if _, err := os.Stat(reportsPath(root, "archive/2026-06")); !os.IsNotExist(err) {
		t.Error("dry-run must not create an archive shard")
	}
}

// TestCleanReportsArchive_ProtectedEntriesUntouched (AC-RLC-010): the four
// explicitly protected entries are untouched even when every other predicate
// condition would admit them.
func TestCleanReportsArchive_ProtectedEntriesUntouched(t *testing.T) {
	root := setupReportsScaffold(t)
	// Probe says nothing is tracked — protection must come from the explicit
	// set, not from the tracked rule.
	stubTrackedProbe(t, func(rel string) bool { return false })

	p := printer.New(printer.WithWriters(&strings.Builder{}, &strings.Builder{}))
	if err := runCleanReportsArchiveWithRoot(p, true, config.DefaultReportsArchiveRetentionDays, root); err != nil {
		t.Fatalf("reports-archive: %v", err)
	}
	for _, rel := range []string{
		"historical/kept.md",
		"plan-audit/.gitkeep",
		"worktrees/wt-slug/evidence.md",
		"archive/2026-08/t-older/evidence.md",
	} {
		if _, statErr := os.Stat(reportsPath(root, rel)); statErr != nil {
			t.Errorf("protected entry %s was affected", rel)
		}
	}
}

// TestCleanReportsArchive_TrackedEntriesAutoProtected (AC-RLC-010 tracked
// rule): an evidence-shaped, window-exceeded entry holding tracked files is
// excluded by the tracked condition alone.
func TestCleanReportsArchive_TrackedEntriesAutoProtected(t *testing.T) {
	root := setupReportsScaffold(t)
	stubTrackedProbe(t, func(rel string) bool {
		// rel arrives slash-separated under the root: .moai/reports/t338/...
		return strings.HasPrefix(rel, ".moai/reports/t338")
	})

	p := printer.New(printer.WithWriters(&strings.Builder{}, &strings.Builder{}))
	if err := runCleanReportsArchiveWithRoot(p, true, config.DefaultReportsArchiveRetentionDays, root); err != nil {
		t.Fatalf("reports-archive: %v", err)
	}
	if _, statErr := os.Stat(reportsPath(root, "t338/ac-count-baseline.txt")); statErr != nil {
		t.Error("tracked entry t338 must stay in place regardless of name and age")
	}
}

// TestCleanReportsArchive_WarnAboveThreshold exercises the 1GiB advisory
// branch without writing a gigabyte: the threshold constant is injected via
// a package var? No — the threshold is read from config; this test asserts
// the warn path through a small dedicated helper instead. See
// reportsArchiveWarnNeeded.
func TestCleanReportsArchive_WarnAboveThreshold(t *testing.T) {
	if !reportsArchiveWarnNeeded(config.DefaultReportsArchiveWarnBytes+1, config.DefaultReportsArchiveWarnBytes) {
		t.Error("total above the warn threshold must trip the advisory")
	}
	if reportsArchiveWarnNeeded(config.DefaultReportsArchiveWarnBytes-1, config.DefaultReportsArchiveWarnBytes) {
		t.Error("total below the warn threshold must not trip the advisory")
	}
}
