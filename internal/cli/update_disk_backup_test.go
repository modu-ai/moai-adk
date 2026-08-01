package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// SPEC-UPDATE-DATA-SURVIVAL-001 M3 — on-disk backup for the three files whose
// only backup used to live in process memory.
//
// These tests are NOT parallel: they replace the package-level
// diskBackupWriteFile seam, which is process-wide state.

// plantInMemoryOnlyFixture writes the three in-memory-only targets plus a
// minimal .moai/config tree into a fresh temp project root and returns it.
func plantInMemoryOnlyFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	files := map[string]string{
		".claude/settings.json":           `{"outputStyle":"MoAI-Easy"}`,
		".moai/status_line.sh":            "#!/bin/bash\necho moai\n",
		".gitignore":                      ".moai/cache/\n.moai/logs/\n",
		".moai/config/sections/user.yaml": "user:\n  name: tester\n",
		".claude/skills/moai/SKILL.md":    "# managed skill\n",
	}
	for rel, body := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return root
}

// hashUserTree hashes every regular file under root EXCEPT the run's own
// backup root (.moai/backups/). The backup root is the run's scratch space —
// files appearing there are additions, never destruction — so including it
// would make "nothing was destroyed" untestable.
func hashUserTree(t *testing.T, root string) string {
	t.Helper()

	var entries []string
	backupRoot := filepath.Join(root, filepath.FromSlash(defs.BackupsDir))
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == backupRoot {
				return filepath.SkipDir
			}
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		sum := sha256.Sum256(data)
		entries = append(entries, filepath.ToSlash(rel)+":"+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		t.Fatalf("hash user tree %s: %v", root, err)
	}
	sort.Strings(entries)
	total := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(total[:])
}

// AC-UDS-008 (REQ-UDS-001, REQ-UDS-002, NFR-UDS-001).
//
// The backup step and the destructive step are invoked as SEPARATE calls, and
// the assertion happens BETWEEN them. That separation is the deterministic
// reach into the crash window the SPEC requires — no crash is raced.
func TestBackup_OnDiskBeforeFirstDestructiveStep(t *testing.T) {
	root := plantInMemoryOnlyFixture(t)

	before := map[string][]byte{}
	for _, rel := range inMemoryOnlyBackupTargets() {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("fixture read %s: %v", rel, err)
		}
		before[rel] = data
	}

	// --- backup step (call 1) ---
	backupDir, written, err := backupInMemoryOnlyFiles(root, "")
	if err != nil {
		t.Fatalf("backupInMemoryOnlyFiles: %v", err)
	}
	if len(written) != len(inMemoryOnlyBackupTargets()) {
		t.Fatalf("expected %d files backed up, got %d (%v)",
			len(inMemoryOnlyBackupTargets()), len(written), written)
	}

	// --- assertion BETWEEN the two calls ---
	for rel, want := range before {
		copyPath := filepath.Join(backupDir, diskBackupSubdir, filepath.FromSlash(rel))
		got, readErr := os.ReadFile(copyPath)
		if readErr != nil {
			t.Fatalf("on-disk backup missing for %s at %s: %v", rel, copyPath, readErr)
		}
		if string(got) != string(want) {
			t.Fatalf("on-disk backup for %s is not byte-identical\n want: %q\n  got: %q",
				rel, string(want), string(got))
		}
	}

	// --- destructive step (call 2) ---
	if err := os.RemoveAll(filepath.Join(root, ".claude", "settings.json")); err != nil {
		t.Fatalf("simulated destructive step: %v", err)
	}

	// The live file is gone; the on-disk copy survives it.
	if _, statErr := os.Stat(filepath.Join(root, ".claude", "settings.json")); !os.IsNotExist(statErr) {
		t.Fatalf("expected live settings.json to be gone after the destructive step")
	}
	if _, statErr := os.Stat(filepath.Join(backupDir, diskBackupSubdir, ".claude", "settings.json")); statErr != nil {
		t.Fatalf("on-disk backup did not survive the destructive step: %v", statErr)
	}
}

// AC-UDS-009 (REQ-UDS-005).
//
// Parity has two halves, because a filesystem comparison alone cannot prove the
// two production paths still share the mechanism:
//
//	(1) the on-disk path SETS produced on two identical fixtures are equal and
//	    non-empty;
//	(2) BOTH production paths route their first destructive step through
//	    guardFirstDestructiveStep — verified by scanning the two sources, so
//	    deleting or bypassing either call site fails this test.
func TestBackup_OnDiskCoverageParityAcrossPaths(t *testing.T) {
	collect := func(root, backupDir string) []string {
		var got []string
		diskRoot := filepath.Join(backupDir, diskBackupSubdir)
		err := filepath.WalkDir(diskRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(diskRoot, path)
			if relErr != nil {
				return relErr
			}
			got = append(got, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", diskRoot, err)
		}
		sort.Strings(got)
		return got
	}

	normalRoot := plantInMemoryOnlyFixture(t)
	normalBackupDir, _, err := backupInMemoryOnlyFiles(normalRoot, "")
	if err != nil {
		t.Fatalf("normal path backup: %v", err)
	}
	normalSet := collect(normalRoot, normalBackupDir)

	cleanRoot := plantInMemoryOnlyFixture(t)
	cleanBackupDir, _, err := backupInMemoryOnlyFiles(cleanRoot, "")
	if err != nil {
		t.Fatalf("clean-reinstall path backup: %v", err)
	}
	cleanSet := collect(cleanRoot, cleanBackupDir)

	if len(normalSet) == 0 {
		t.Fatalf("normal path produced an empty backup set — parity would be vacuous")
	}
	if strings.Join(normalSet, "|") != strings.Join(cleanSet, "|") {
		t.Fatalf("on-disk backup sets differ\n normal: %v\n  clean: %v", normalSet, cleanSet)
	}

	// (2) both production call sites must exist.
	for _, src := range []string{"update_template_sync.go", "update_clean_install.go"} {
		data, readErr := os.ReadFile(src)
		if readErr != nil {
			t.Fatalf("read %s: %v", src, readErr)
		}
		if !strings.Contains(string(data), "guardFirstDestructiveStep(") {
			t.Fatalf("%s does not route its first destructive step through guardFirstDestructiveStep — "+
				"the two paths can drift into one being unprotected (REQ-UDS-005)", src)
		}
	}
}

// AC-UDS-020 (REQ-UDS-003).
//
// A failing on-disk backup write must abort BEFORE the first destructive step.
// The abort is observed through a call-recording spy, not inferred from an
// unchanged tree: an unchanged tree alone would also hold if the destructive
// step ran and happened to remove nothing.
func TestBackup_AbortsBeforeDestructionOnWriteFailure(t *testing.T) {
	root := plantInMemoryOnlyFixture(t)
	const failing = ".moai/status_line.sh"

	original := diskBackupWriteFile
	t.Cleanup(func() { diskBackupWriteFile = original })
	diskBackupWriteFile = func(name string, data []byte, perm fs.FileMode) error {
		if strings.HasSuffix(filepath.ToSlash(name), filepath.ToSlash(failing)) {
			return errors.New("injected write failure")
		}
		return original(name, data, perm)
	}

	treeBefore := hashUserTree(t, root)

	destructiveCalls := 0
	spy := func() error {
		destructiveCalls++
		return nil
	}

	err := guardFirstDestructiveStep(root, "", spy)

	// (1) grep-able sentinel naming the file that could not be backed up.
	if err == nil {
		t.Fatalf("expected an error when the on-disk backup write fails")
	}
	wantSentinel := fmt.Sprintf("%s %s", backupWriteFailedSentinel, failing)
	if !strings.Contains(err.Error(), wantSentinel) {
		t.Fatalf("error must contain %q, got %q", wantSentinel, err.Error())
	}

	// (2) the destructive step was never invoked — observed, not inferred.
	if destructiveCalls != 0 {
		t.Fatalf("destructive step was invoked %d time(s); the abort must happen BEFORE it", destructiveCalls)
	}

	// (3) the fixture tree is byte-identical to its pre-run state.
	if treeAfter := hashUserTree(t, root); treeAfter != treeBefore {
		t.Fatalf("fixture tree changed: %s -> %s", treeBefore, treeAfter)
	}
}
