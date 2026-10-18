package cli

// update_prior_manifest_collision_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// (card t1865, sync-audit finding F2; REQ-OWR-017, AC-OWR-18 (c)).
//
// Two update runs that start in the same clock second are handed the same config
// backup path by BackupMoaiConfig (the directory name is a timestamp to the
// second). The Backup stage used to write the later run's start-of-run manifest
// onto the earlier run's manifest.prior.json — the older generation was lost, and
// REQ-OWR-017 says nothing in this code may delete, prune or shorten the
// retention of a run backup.
//
// The fix puts each run's copy in a directory no other run writes to: when the
// intended directory already holds a manifest.prior.json, the copy goes to a new
// sibling `<name>-NN` (NN = 02..99, fixed width so name order is age order),
// created with Mkdir. The sibling carries the intended directory's run marker, so
// retention still counts both directories as one run and the census neither
// gains nor loses a run.
//
// Nothing here reads a clock. A same-second collision is exactly "the config
// backup path of the second run equals that of the first", which the tests state
// directly by passing the same path twice. The no-wait cases (an empty config
// backup path: the run directory is created from the clock inside the call) hold
// whichever second the calls land in.

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
	return []byte(fmt.Sprintf(`{"files":{%q:{"provenance":"template_managed","template_hash":"sha256:%s"}}}`, mark, mark))
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

func collisionPreserve(t *testing.T, root, cfg string, data []byte) string {
	t.Helper()
	dir, state, err := preservePriorManifest(root, cfg, data)
	if err != nil || state != priorManifestWritten {
		t.Fatalf("preservePriorManifest(%q) = (%q, %v, %v), want a written copy", cfg, dir, state, err)
	}
	return dir
}

func TestPriorManifestOfAnEarlierRunSurvivesALaterRun(t *testing.T) {
	t.Run("TwoRunsSharingTheConfigBackupPath", func(t *testing.T) {
		root, cfg, marker := collisionRun(t)
		d1 := collisionPreserve(t, root, cfg, collisionGen("one"))
		d2 := collisionPreserve(t, root, cfg, collisionGen("two"))
		if d1 != cfg {
			t.Errorf("the first run's copy went to %s, want the intended %s", d1, cfg)
		}
		if d2 == d1 {
			t.Fatalf("both runs wrote into %s: the later copy replaced the earlier one", d1)
		}
		if got := collisionCopy(t, d1); got != string(collisionGen("one")) {
			t.Errorf("the earlier run's copy changed: %s", got)
		}
		if got := collisionCopy(t, d2); got != string(collisionGen("two")) {
			t.Errorf("the later run's copy is wrong: %s", got)
		}
		if filepath.Dir(d2) != filepath.Dir(cfg) || !strings.HasPrefix(filepath.Base(d2), collisionStamp+"-") {
			t.Errorf("the later copy's directory %s is not a sibling `%s-NN` of the intended one", d2, collisionStamp)
		}
		// The sibling belongs to the same run as the intended directory.
		got, err := os.ReadFile(filepath.Join(d2, backup.RunMarkerName))
		if err != nil || strings.TrimSpace(string(got)) != marker {
			t.Errorf("the sibling's run marker is %q (%v), want %q", got, err, marker)
		}
	})

	t.Run("ThreeRunsKeepNameOrderEqualToAgeOrderAndTheReaderSeesEveryEarlierOne", func(t *testing.T) {
		root, cfg, _ := collisionRun(t)
		d1 := collisionPreserve(t, root, cfg, collisionGen("one"))
		d2 := collisionPreserve(t, root, cfg, collisionGen("two"))
		d3 := collisionPreserve(t, root, cfg, collisionGen("three"))
		names := []string{filepath.Base(d1), filepath.Base(d2), filepath.Base(d3)}
		if !slices.IsSorted(names) || len(slices.Compact(slices.Clone(names))) != 3 {
			t.Errorf("directory names %v are not strictly increasing with age", names)
		}
		gens, _ := readPriorGenerations(root, d3)
		var seen []string
		for _, g := range gens {
			for rel := range g.files {
				seen = append(seen, rel)
			}
		}
		if !slices.Equal(seen, []string{"two", "one"}) {
			t.Errorf("the reader saw %v (newest first), want [two one]", seen)
		}
	})

	t.Run("ADirectoryTheRunDidNotCreateIsNeverWrittenInto", func(t *testing.T) {
		root, cfg, _ := collisionRun(t)
		collisionPreserve(t, root, cfg, collisionGen("one"))
		foreign := cfg + "-02"
		if err := os.Mkdir(foreign, defs.DirPerm); err != nil {
			t.Fatal(err)
		}
		d2 := collisionPreserve(t, root, cfg, collisionGen("two"))
		if d2 == foreign {
			t.Fatalf("the copy went into %s, a directory this run did not create", foreign)
		}
		if _, err := os.Lstat(filepath.Join(foreign, priorManifestName)); !os.IsNotExist(err) {
			t.Errorf("the foreign directory gained a prior manifest (%v)", err)
		}
	})

	t.Run("NoFreeNameFailsClosedAndOverwritesNothing", func(t *testing.T) {
		root, cfg, _ := collisionRun(t)
		collisionPreserve(t, root, cfg, collisionGen("one"))
		for n := 2; n <= 99; n++ {
			if err := os.Mkdir(fmt.Sprintf("%s-%02d", cfg, n), defs.DirPerm); err != nil {
				t.Fatal(err)
			}
		}
		dir, state, err := preservePriorManifest(root, cfg, collisionGen("two"))
		if state != priorManifestFailed || err == nil {
			t.Errorf("got (%q, %v, %v), want a failed state with an error: the run must remove nothing", dir, state, err)
		}
		if got := collisionCopy(t, cfg); got != string(collisionGen("one")) {
			t.Errorf("the earlier run's copy changed: %s", got)
		}
	})

	t.Run("RunsWithoutAConfigBackupPathNeverShareACopy", func(t *testing.T) {
		// The directory is created from the clock inside the call, so the two
		// runs may or may not share a second. The copies must differ either way.
		root := t.TempDir()
		d1 := collisionPreserve(t, root, "", collisionGen("one"))
		d2 := collisionPreserve(t, root, "", collisionGen("two"))
		d3 := collisionPreserve(t, root, "", collisionGen("three"))
		if d1 == d2 || d2 == d3 || d1 == d3 {
			t.Fatalf("runs share a copy: %s %s %s", d1, d2, d3)
		}
		for dir, want := range map[string]string{d1: "one", d2: "two", d3: "three"} {
			if got := collisionCopy(t, dir); got != string(collisionGen(want)) {
				t.Errorf("%s holds %s, want %s", dir, got, want)
			}
		}
	})

	t.Run("RetentionStillCountsTheSiblingWithItsRun", func(t *testing.T) {
		root, cfg, _ := collisionRun(t)
		collisionPreserve(t, root, cfg, collisionGen("one"))
		before := backup.PruneBackups(root, -1, 0) // a negative keepCount removes nothing
		collisionPreserve(t, root, cfg, collisionGen("two"))
		after := backup.PruneBackups(root, -1, 0)
		if before.Runs != 1 || after.Runs != before.Runs {
			t.Errorf("run count %d then %d: the sibling directory must not count as another run", before.Runs, after.Runs)
		}
		if after.Unmanaged.Count != 0 {
			t.Errorf("the sibling counts as %d unmanaged entries", after.Unmanaged.Count)
		}
	})
}
