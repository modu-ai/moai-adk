package backup

// retention_prune_test.go — pruner-level criteria of SPEC-UPDATE-BACKUP-RETENTION-001
// M2: AC-UBR-08 (bookkeeping by writer shape), AC-UBR-09 (template render exempt),
// AC-UBR-12 (each decision reported), AC-UBR-13 (containment), AC-UBR-14 (explicit
// wipe and negative keepCount). AC-UBR-01/02/03 are the P1-P3 probes; AC-UBR-11 is
// in retention_prune_unix_test.go.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// t1733RunWith creates namespace run n holding files (path inside the run -> body)
// and stamps it with the time its name encodes. It returns the run directory.
func t1733RunWith(t *testing.T, projectRoot string, n int, files map[string]string) string {
	t.Helper()
	name := fmt.Sprintf("update-2026-10-%02dT00-00-00Z", n)
	dir := filepath.Join(projectRoot, ".moai", "backups", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for rel, body := range files {
		t1733WriteFile(t, filepath.Join(dir, filepath.FromSlash(rel)), body)
	}
	ts := time.Date(2026, 10, n, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(dir, ts, ts); err != nil {
		t.Fatal(err)
	}
	return dir
}

// t1733Item returns the report row whose run is dir, or false.
func t1733Item(rep PruneReport, dir string) (PruneItem, bool) {
	for _, it := range rep.Items {
		if it.Run == dir {
			return it, true
		}
	}
	return PruneItem{}, false
}

func t1733Exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func t1733MetadataJSON(t *testing.T, stamp string) string {
	t.Helper()
	b, err := json.Marshal(BackupMetadata{Timestamp: stamp, Description: "config_backup", BackupType: "config"})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// AC-UBR-08: bookkeeping is what the writer wrote, not what is named like it.
func TestAC_UBR_08_BookkeepingIsRecognizedByWhatTheWriterWrote(t *testing.T) {
	root := t.TempDir()
	// Window (newest): a newer copy of the user's .complete file.
	t1733RunWith(t, root, 3, map[string]string{
		"x.md":      "shared\n",
		".complete": "user B\n",
	})
	// Excess run 2 differs from the window only by writer-produced bookkeeping.
	run2 := t1733RunWith(t, root, 2, map[string]string{
		"x.md":                 "shared\n",
		"backup_metadata.json": t1733MetadataJSON(t, "stamp-2"),
		".complete":            "stamp=2026-10-02T00-00-00Z\nfiles=1\ntimestamp=2026-10-02T00:00:00Z\n",
	})
	// Excess run 1 holds a user file named .complete whose content is user content.
	run1 := t1733RunWith(t, root, 1, map[string]string{
		"x.md":      "shared\n",
		".complete": "user A\n",
	})

	rep := PruneBackups(root, 1, 3)

	if it, ok := t1733Item(rep, run1); !ok || it.Action != RetentionActionRetained || it.Reason != RetentionReasonProtectsVersions {
		t.Errorf("AC-UBR-08: run holding the user .complete file = %+v (found=%v), want retained for %s", it, ok, RetentionReasonProtectsVersions)
	}
	if !t1733Exists(run1) {
		t.Errorf("AC-UBR-08: the run holding user A was removed from disk")
	}
	if it, ok := t1733Item(rep, run2); !ok || it.Action != RetentionActionRemoved {
		t.Errorf("AC-UBR-08: run differing only by bookkeeping = %+v (found=%v), want removed", it, ok)
	}
	if t1733Exists(run2) {
		t.Errorf("AC-UBR-08: the bookkeeping-only run remains on disk")
	}
}

// AC-UBR-09: the .template-defaults render is not content the budget protects.
func TestAC_UBR_09_TemplateDefaultsRenderIsExempt(t *testing.T) {
	root := t.TempDir()
	t1733RunWith(t, root, 2, map[string]string{"x.md": "shared\n"})
	run1 := t1733RunWith(t, root, 1, map[string]string{
		"x.md":                               "shared\n",
		".template-defaults/sections/a.yaml": "only-in-the-render\n",
	})

	rep := PruneBackups(root, 1, 3)

	if it, ok := t1733Item(rep, run1); !ok || it.Action != RetentionActionRemoved {
		t.Errorf("AC-UBR-09: run holding only a .template-defaults render = %+v (found=%v), want removed", it, ok)
	}
	if t1733Exists(run1) {
		t.Errorf("AC-UBR-09: the render-only run remains on disk")
	}
}

// AC-UBR-12: every removed run, every retained excess run with its reason, and
// the unmanaged entries as a count and a size; nothing unmanaged is deleted.
func TestAC_UBR_12_EachDecisionIsReported(t *testing.T) {
	root := t.TempDir()
	t1733RunWith(t, root, 4, map[string]string{"x.md": "v4\n"})
	removed := t1733RunWith(t, root, 3, map[string]string{"x.md": "v4\n"})
	kept := t1733RunWith(t, root, 2, map[string]string{"x.md": "v2\n"})
	t1733RunWith(t, root, 1, map[string]string{"x.md": "v1\n"})

	cfgRoot := filepath.Join(root, ".moai-backups")
	t1733WriteFile(t, filepath.Join(cfgRoot, "notes.txt"), "0123456789")                  // 10 bytes
	t1733WriteFile(t, filepath.Join(cfgRoot, "scratch", "a.bin"), "01234567890123456789") // 20 bytes

	rep := PruneBackups(root, 1, 1)

	if rep.Window != 1 || rep.Runs != 4 {
		t.Errorf("AC-UBR-12: window=%d runs=%d, want 1 and 4", rep.Window, rep.Runs)
	}
	if it, ok := t1733Item(rep, removed); !ok || it.Action != RetentionActionRemoved {
		t.Errorf("AC-UBR-12: redundant run = %+v (found=%v), want a removed row", it, ok)
	}
	if it, ok := t1733Item(rep, kept); !ok || it.Action != RetentionActionRetained || it.Reason != RetentionReasonProtectsVersions || it.Protects != 1 {
		t.Errorf("AC-UBR-12: protecting run = %+v (found=%v), want retained, %s, protects 1", it, ok, RetentionReasonProtectsVersions)
	}
	// Budget 1: v2 is the newest past version, v1 falls outside it, so v1's run is removed.
	if len(rep.Items) != 3 {
		t.Errorf("AC-UBR-12: %d report rows, want 3 (one per excess run): %+v", len(rep.Items), rep.Items)
	}
	if rep.Removed != 2 {
		t.Errorf("AC-UBR-12: Removed=%d directories, want 2", rep.Removed)
	}
	if rep.Unmanaged.Count != 2 || rep.Unmanaged.Bytes != 30 {
		t.Errorf("AC-UBR-12: unmanaged = %+v, want count 2 and 30 bytes", rep.Unmanaged)
	}
	for _, p := range []string{filepath.Join(cfgRoot, "notes.txt"), filepath.Join(cfgRoot, "scratch", "a.bin")} {
		if !t1733Exists(p) {
			t.Errorf("AC-UBR-12: unmanaged entry %s was deleted", p)
		}
	}
}

// AC-UBR-13: a root that is a symbolic link (config root, namespace root, or a
// root under a symbolic parent) is refused and nothing is removed.
func TestAC_UBR_13_RootContainmentStillRefuses(t *testing.T) {
	setup := func(t *testing.T, link func(root, outside string) string) (root string, runs []string) {
		t.Helper()
		root = t.TempDir()
		outside := t.TempDir()
		for i := 1; i <= 3; i++ {
			for _, rel := range []string{
				fmt.Sprintf("2026101%d_000000", i),
				fmt.Sprintf("update-2026-10-%02dT00-00-00Z", i),
				filepath.Join("backups", fmt.Sprintf("update-2026-10-%02dT00-00-00Z", i)),
			} {
				t1733WriteFile(t, filepath.Join(outside, rel, "x.md"), "x\n")
				runs = append(runs, filepath.Join(outside, rel))
			}
		}
		target := link(root, outside)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, target); err != nil {
			t.Skipf("symbolic links unavailable: %v", err)
		}
		return root, runs
	}
	cases := map[string]func(root, outside string) string{
		"config root symlink":    func(root, _ string) string { return filepath.Join(root, ".moai-backups") },
		"namespace root symlink": func(root, _ string) string { return filepath.Join(root, ".moai", "backups") },
		"root under symlinked parent": func(root, _ string) string {
			return filepath.Join(root, ".moai")
		},
	}
	for name, link := range cases {
		t.Run(name, func(t *testing.T) {
			root, runs := setup(t, link)
			rep := PruneBackups(root, 1, 3)
			if !rep.RootRefused {
				t.Errorf("AC-UBR-13: RootRefused=false, want true")
			}
			if rep.Removed != 0 || len(rep.Items) != 0 {
				t.Errorf("AC-UBR-13: removed=%d items=%d, want nothing pruned", rep.Removed, len(rep.Items))
			}
			for _, p := range runs {
				if !t1733Exists(p) {
					t.Errorf("AC-UBR-13: %s was removed through a refused root", p)
				}
			}
		})
	}
}

// AC-UBR-14: keepCount 0 removes every run and ignores protection; a negative
// keepCount removes nothing.
func TestAC_UBR_14_ExplicitWipeAndNegativeKeepCount(t *testing.T) {
	build := func(t *testing.T) (string, []string) {
		root := t.TempDir()
		var dirs []string
		for i, v := range []string{"v1", "v2", "v3"} {
			dirs = append(dirs, t1733RunWith(t, root, i+1, map[string]string{"x.md": v + "\n"}))
		}
		return root, dirs
	}

	t.Run("zero wipes every run", func(t *testing.T) {
		root, dirs := build(t)
		rep := PruneBackups(root, 0, 3)
		for _, d := range dirs {
			if t1733Exists(d) {
				t.Errorf("AC-UBR-14: %s remains after keepCount 0", d)
			}
		}
		if rep.Removed != 3 {
			t.Errorf("AC-UBR-14: Removed=%d, want 3", rep.Removed)
		}
		for _, it := range rep.Items {
			if it.Action != RetentionActionRemoved || it.Reason != "" {
				t.Errorf("AC-UBR-14: row %+v, want a plain removed row", it)
			}
		}
	})
	t.Run("negative removes nothing", func(t *testing.T) {
		root, dirs := build(t)
		rep := PruneBackups(root, -1, 3)
		for _, d := range dirs {
			if !t1733Exists(d) {
				t.Errorf("AC-UBR-14: %s was removed by keepCount -1", d)
			}
		}
		if rep.Removed != 0 || len(rep.Items) != 0 {
			t.Errorf("AC-UBR-14: report %+v, want nothing removed", rep)
		}
	})
}

// A marker-grouped run is one unit across both roots: it is removed whole or kept
// whole, which is what the P1 probe's pre-clean run depends on.
func TestPrune_RunSpansBothRootsAsOneUnit(t *testing.T) {
	root := t.TempDir()
	t1733RunWith(t, root, 3, map[string]string{"x.md": "w\n"})
	old := "2026-10-01T00:00:00.000000000Z"
	cfg := filepath.Join(root, ".moai-backups", "20261001_000000")
	ns := t1733RunWith(t, root, 1, map[string]string{"x.md": "w\n"})
	t1733WriteFile(t, filepath.Join(cfg, "x.yaml"), "y\n")
	for _, d := range []string{cfg, ns} {
		t1733WriteFile(t, filepath.Join(d, RunMarkerName), old+"\n")
	}

	rep := PruneBackups(root, 1, 3)

	// The run is one unit: the config directory holds the only copy of x.yaml, so
	// the whole run, namespace directory included, is retained.
	if len(rep.Items) != 1 || rep.Items[0].Action != RetentionActionRetained || len(rep.Items[0].Dirs) != 2 {
		t.Fatalf("run row = %+v, want one retained row spanning two directories", rep.Items)
	}
	for _, d := range []string{cfg, ns} {
		if !t1733Exists(d) {
			t.Errorf("%s was removed although its run is retained", d)
		}
	}
	if !strings.Contains(rep.Items[0].Dirs[0], string(filepath.Separator)) {
		t.Errorf("Dirs[0]=%q, want a path", rep.Items[0].Dirs[0])
	}
}
