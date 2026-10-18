package cli

// update_prior_manifest_collision_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// (card t1865, chunk 13, leader disposition B: the guarded surface is reduced).
//
// Two update runs that start in the same clock second are handed the same config
// backup path by BackupMoaiConfig (the directory name is a timestamp to the
// second). The Backup stage used to write the later run's start-of-run manifest
// onto the earlier run's manifest.prior.json (REQ-OWR-017: nothing may delete a run
// backup's evidence). Rounds 1-4 answered with a sibling-directory scheme; the
// scheme itself kept drawing findings (a run marker to copy, a directory to clean
// up), so it is gone.
//
// Behaviour pinned here: when the intended run backup directory already holds a
// manifest.prior.json, the later run does NOT overwrite it and does NOT create
// another directory. Its prior-manifest state is priorManifestFailed, so it removes
// nothing (the existing fail-closed path of REQ-OWR-016), the report says why, and
// the earlier copy stays byte-identical. A different directory takes its copy as
// ever.
//
// Nothing reads a clock: a same-second collision is exactly "the config backup path
// of the second run equals that of the first", which the tests state by passing the
// same path twice.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/defs"
)

const collisionStamp = "20261018_153045"

func collisionGen(mark string) []byte {
	return []byte(`{"files":{"` + mark + `":{"provenance":"template_managed","template_hash":"sha256:` + mark + `"}}}`)
}

// collisionRun builds a project root with the intended config backup directory,
// carrying the run marker the Backup stage writes before it preserves the manifest.
func collisionRun(t *testing.T) (root, cfg, marker string) {
	t.Helper()
	root = t.TempDir()
	cfg = filepath.Join(root, defs.BackupsDir, collisionStamp)
	if err := os.MkdirAll(cfg, defs.DirPerm); err != nil {
		t.Fatal(err)
	}
	marker = "2026-10-18T15:30:45.000000007Z"
	if err := backup.WriteRunMarker(cfg, marker); err != nil {
		t.Fatal(err)
	}
	return root, cfg, marker
}

func collisionCopy(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, priorManifestName))
	if err != nil {
		t.Fatalf("no prior manifest in %s: %v", dir, err)
	}
	return string(data)
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

func TestSecondClaimInTheSameDirectoryFailsClosed(t *testing.T) {
	t.Run("SameDirectory", func(t *testing.T) {
		root, cfg, marker := collisionRun(t)
		d1, s1, e1 := preservePriorManifest(root, cfg, collisionGen("one"))
		if s1 != priorManifestWritten || e1 != nil {
			t.Fatalf("the first claim: (%q, %v, %v), want a written copy", d1, s1, e1)
		}
		d2, s2, e2 := preservePriorManifest(root, cfg, collisionGen("two"))
		if s2 != priorManifestFailed || e2 == nil {
			t.Fatalf("the second claim in the same directory: (%q, %v, %v), want a failed state with an error", d2, s2, e2)
		}
		if !strings.Contains(e2.Error(), "earlier run") {
			t.Errorf("the error does not say why: %v", e2)
		}
		if got := collisionCopy(t, cfg); got != string(collisionGen("one")) {
			t.Errorf("the earlier run's copy changed: %s", got)
		}
		if names := backupNames(t, root); !slices.Equal(names, []string{collisionStamp}) {
			t.Errorf("backup directories %v, want only %s: the later run must not make another directory", names, collisionStamp)
		}
		if left := priorTmpLeftovers(t, cfg); len(left) != 0 {
			t.Errorf("temporary files remain: %v", left)
		}
		if got, err := os.ReadFile(filepath.Join(cfg, backup.RunMarkerName)); err != nil || strings.TrimSpace(string(got)) != marker {
			t.Errorf("the run marker is %q (%v), want %q", got, err, marker)
		}
	})

	t.Run("ADifferentDirectoryTakesItsCopy", func(t *testing.T) {
		root := t.TempDir()
		a := filepath.Join(root, defs.BackupsDir, "20261018_153045")
		b := filepath.Join(root, defs.BackupsDir, "20261018_153046")
		for _, d := range []string{a, b} {
			if err := os.MkdirAll(d, defs.DirPerm); err != nil {
				t.Fatal(err)
			}
		}
		if _, s, err := preservePriorManifest(root, a, collisionGen("one")); s != priorManifestWritten || err != nil {
			t.Fatalf("first: %v %v", s, err)
		}
		if _, s, err := preservePriorManifest(root, b, collisionGen("two")); s != priorManifestWritten || err != nil {
			t.Fatalf("second: %v %v", s, err)
		}
		if collisionCopy(t, a) != string(collisionGen("one")) || collisionCopy(t, b) != string(collisionGen("two")) {
			t.Errorf("each directory must hold its own run's copy")
		}
		gens, _ := readPriorGenerations(root, b)
		if len(gens) != 1 {
			t.Errorf("the reader sees %d earlier generations, want 1", len(gens))
		}
	})

	// No config backup path: the run directory is made from the clock inside the
	// call, so the calls may or may not share a second. Whichever they do, no copy
	// is ever replaced and a call either writes into a directory nobody wrote to or
	// fails closed.
	t.Run("RunsWithoutAConfigBackupPathNeverReplaceACopy", func(t *testing.T) {
		root := t.TempDir()
		first := map[string]string{} // directory -> the copy written first
		for _, mark := range []string{"one", "two", "three"} {
			dir, state, err := preservePriorManifest(root, "", collisionGen(mark))
			switch state {
			case priorManifestWritten:
				if _, seen := first[dir]; seen {
					t.Fatalf("%s: a second copy was written into %s", mark, dir)
				}
				first[dir] = string(collisionGen(mark))
			case priorManifestFailed:
				if err == nil {
					t.Errorf("%s: failed state without an error", mark)
				}
			default:
				t.Fatalf("%s: state %v", mark, state)
			}
		}
		for dir, want := range first {
			if got := collisionCopy(t, dir); got != want {
				t.Errorf("%s holds %s, want the first copy %s", dir, got, want)
			}
		}
	})

	// The hand-off: a collided claim makes the migration remove nothing and say why.
	// A control run with a free directory removes the file.
	t.Run("TheLaterRunRemovesNothingAndSaysWhy", func(t *testing.T) {
		run := func(t *testing.T, collide bool) (removed []string, report string, state priorManifestState) {
			t.Helper()
			root, home := ownershipFixtureBase(t)
			runDir := filepath.Join(root, defs.BackupsDir, collisionStamp)
			if err := os.MkdirAll(runDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if collide {
				if err := os.WriteFile(filepath.Join(runDir, priorManifestName), collisionGen("earlier"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			dir, state, err := preservePriorManifest(root, runDir, collisionGen("this run"))
			rc := &migrationRunContext{userAssetsInstalled: true, runBackupDir: dir, priorManifest: state, priorManifestErr: err}
			var lines []string
			removed, migErr := migrateProjectCommonAssets(root, home, true, nil, func(f string, a ...interface{}) { lines = append(lines, strings.TrimSpace(fmt.Sprintf(f, a...))) }, rc)
			if migErr != nil {
				t.Fatalf("migration: %v", migErr)
			}
			if collide {
				if got := collisionCopy(t, runDir); got != string(collisionGen("earlier")) {
					t.Errorf("the earlier copy changed: %s", got)
				}
			}
			return removed, strings.Join(lines, "\n"), state
		}
		okRemoved, _, okState := run(t, false)
		if okState != priorManifestWritten || !slices.Equal(okRemoved, []string{migIdenticalSkill}) {
			t.Fatalf("control: state=%v removed=%v, want written/[%s]", okState, okRemoved, migIdenticalSkill)
		}
		removed, report, state := run(t, true)
		if state != priorManifestFailed {
			t.Errorf("state=%v, want failed", state)
		}
		if len(removed) != 0 {
			t.Errorf("the run removed %v although its prior manifest copy was not kept", removed)
		}
		if !strings.Contains(report, "removes no file") || !strings.Contains(report, "earlier run") {
			t.Errorf("the report does not say why the run removed nothing:\n%s", report)
		}
	})
}
