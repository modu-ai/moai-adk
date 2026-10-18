package cli

// update_prior_manifest_markercopy_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// (card t1865, delta-audit finding N2): when claimPriorManifestDir puts a run's
// prior-manifest copy in a new sibling directory it gives the sibling the run
// marker of the intended directory, so retention counts both as one run. A marker
// copy that failed used to be ignored, leaving a sibling that carries no marker
// and whose name does not match the config-root pattern: retention neither counts
// nor prunes it, ever (an unmanaged entry).
//
// Expected behaviour, pinned here: a marker that exists but cannot be copied makes
// the claim fail — the just-created (still empty) sibling is removed, the state is
// priorManifestFailed, and the run removes nothing, like every other failure of
// the hand-off. A run that has no marker to copy at all (the intended directory
// carries none: an old layout, a standalone call, or a first line that is not a
// run id) keeps the behaviour it always had: the sibling is made, without one.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/defs"
)

func swapMarkerCopy(t *testing.T, fn func(src, dst string) error) {
	t.Helper()
	prev := priorManifestMarkerCopy
	priorManifestMarkerCopy = fn
	t.Cleanup(func() { priorManifestMarkerCopy = prev })
}

func backupNames(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, defs.BackupsDir))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestMarkerCopyFailureFailsClosedWithoutResidue(t *testing.T) {
	injected := errors.New("injected marker copy failure")

	t.Run("NoSiblingIsLeftAndTheEarlierCopyIsUntouched", func(t *testing.T) {
		root, cfg, _ := collisionRun(t)
		collisionPreserve(t, root, cfg, collisionGen("one"))
		swapMarkerCopy(t, func(string, string) error { return injected })
		dir, state, err := preservePriorManifest(root, cfg, collisionGen("two"))
		if state != priorManifestFailed || !errors.Is(err, injected) {
			t.Fatalf("got (%q, %v, %v), want a failed state carrying the injected error", dir, state, err)
		}
		if names := backupNames(t, root); !slices.Equal(names, []string{collisionStamp}) {
			t.Errorf("backup directories %v remain, want only the intended %s (the empty sibling must be removed)", names, collisionStamp)
		}
		if got := collisionCopy(t, cfg); got != string(collisionGen("one")) {
			t.Errorf("the earlier copy changed: %s", got)
		}
		rep := backup.PruneBackups(root, -1, 0) // a negative keepCount removes nothing
		if rep.Runs != 1 || rep.Unmanaged.Count != 0 {
			t.Errorf("retention sees runs=%d unmanaged=%d, want 1 and 0: a failed claim must leave no unmanaged residue", rep.Runs, rep.Unmanaged.Count)
		}
	})

	t.Run("TheRunRemovesNothing", func(t *testing.T) {
		removedWith := func(t *testing.T, fail bool) (removed []string, state priorManifestState, err error) {
			t.Helper()
			root, home := ownershipFixtureBase(t)
			runDir := filepath.Join(root, defs.BackupsDir, collisionStamp)
			if mkErr := os.MkdirAll(runDir, 0o755); mkErr != nil {
				t.Fatal(mkErr)
			}
			if wErr := backup.WriteRunMarker(runDir, "2026-10-18T15:30:45.000000007Z"); wErr != nil {
				t.Fatal(wErr)
			}
			collisionPreserve(t, root, runDir, collisionGen("earlier")) // the next claim collides
			if fail {
				swapMarkerCopy(t, func(string, string) error { return injected })
			}
			dir, state, err := preservePriorManifest(root, runDir, collisionGen("this run"))
			rc := &migrationRunContext{userAssetsInstalled: true, runBackupDir: dir, priorManifest: state, priorManifestErr: err}
			removed, migErr := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}, rc)
			if migErr != nil {
				t.Fatalf("migration: %v", migErr)
			}
			return removed, state, err
		}
		okRemoved, okState, okErr := removedWith(t, false)
		if okState != priorManifestWritten || okErr != nil || !slices.Equal(okRemoved, []string{migIdenticalSkill}) {
			t.Fatalf("control: state=%v err=%v removed=%v, want written/nil/[%s]", okState, okErr, okRemoved, migIdenticalSkill)
		}
		removed, state, err := removedWith(t, true)
		if state != priorManifestFailed || !errors.Is(err, injected) {
			t.Errorf("state=%v err=%v, want a failed state carrying the injected error", state, err)
		}
		if len(removed) != 0 {
			t.Errorf("the run removed %v although its prior manifest copy could not be placed", removed)
		}
	})

	t.Run("NoMarkerToCopyKeepsTheBehaviourItAlwaysHad", func(t *testing.T) {
		root := t.TempDir()
		cfg := filepath.Join(root, defs.BackupsDir, collisionStamp)
		if err := os.MkdirAll(cfg, defs.DirPerm); err != nil {
			t.Fatal(err)
		}
		collisionPreserve(t, root, cfg, collisionGen("one"))
		d2 := collisionPreserve(t, root, cfg, collisionGen("two")) // no marker anywhere, no error
		if d2 == cfg {
			t.Fatalf("both copies went to %s", cfg)
		}
		if _, err := os.Lstat(filepath.Join(d2, backup.RunMarkerName)); !os.IsNotExist(err) {
			t.Errorf("the sibling has a run marker although there was none to copy (%v)", err)
		}
	})

	t.Run("MarkerWithExtraLinesAndCRLFIsCopiedAsOneCleanLine", func(t *testing.T) {
		root, cfg, marker := collisionRun(t)
		if err := os.WriteFile(filepath.Join(cfg, backup.RunMarkerName), []byte(marker+"\r\nsecond line\r\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		collisionPreserve(t, root, cfg, collisionGen("one"))
		before := backup.PruneBackups(root, -1, 0)
		d2 := collisionPreserve(t, root, cfg, collisionGen("two"))
		got, err := os.ReadFile(filepath.Join(d2, backup.RunMarkerName))
		if err != nil || string(got) != marker+"\n" {
			t.Errorf("the sibling's marker is %q (%v), want %q", got, err, marker+"\n")
		}
		if after := backup.PruneBackups(root, -1, 0); after.Runs != before.Runs || after.Unmanaged.Count != 0 {
			t.Errorf("retention counts %d runs and %d unmanaged entries after the sibling, was %d runs", after.Runs, after.Unmanaged.Count, before.Runs)
		}
	})

	t.Run("SuffixNamesTakenByAFileOrADanglingLinkAreSkipped", func(t *testing.T) {
		root, cfg, _ := collisionRun(t)
		collisionPreserve(t, root, cfg, collisionGen("one"))
		if err := os.WriteFile(cfg+"-02", []byte("a file stands here\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(root, "nowhere"), cfg+"-03"); err != nil {
			t.Fatalf("os.Symlink failed: %v", err)
		}
		d := collisionPreserve(t, root, cfg, collisionGen("two"))
		if filepath.Base(d) != collisionStamp+"-04" {
			t.Errorf("the copy went to %s, want the first free name %s-04", d, collisionStamp)
		}
		if got, err := os.ReadFile(cfg + "-02"); err != nil || string(got) != "a file stands here\n" {
			t.Errorf("the file occupying -02 was disturbed (%v)", err)
		}
	})

	t.Run("ABackupPathWithATrailingSeparatorStillGetsASibling", func(t *testing.T) {
		root, cfg, _ := collisionRun(t)
		collisionPreserve(t, root, cfg, collisionGen("one"))
		d2 := collisionPreserve(t, root, cfg+string(filepath.Separator), collisionGen("two"))
		want := filepath.Join(root, defs.BackupsDir, collisionStamp+"-02")
		if filepath.Clean(d2) != want {
			t.Errorf("the sibling is %s, want %s (a sibling of the run directory, not a directory inside it)", d2, want)
		}
	})
}
