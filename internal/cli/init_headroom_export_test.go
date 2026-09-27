package cli

// Isolated S_init export harness (SPEC-ALWAYS-LOADED-HEADROOM-001, AC-ALH-009).
//
// TestHeadroomInitSurfaceExport runs the real init command against a
// t.TempDir() project root with the home seams redirected by
// prepareSafeInitHome, then copies the whole produced project tree (minus
// .git) to the directory named by the -headroom-export test flag, so the
// always-loaded instruction surface a default (slim) user receives can be
// measured without ever writing to the operator's real home. Without the flag
// the test skips: it is a measurement harness, not a regression gate.

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

var headroomExportDir = flag.String("headroom-export", "",
	"export the init-produced project tree to this directory (TestHeadroomInitSurfaceExport)")

// headroomCountPaths are the 18 always-loaded paths of the headroom SPEC.
var headroomCountPaths = []string{
	"CLAUDE.md",
	"AGENTS.md",
	".moai/config/sections/user.yaml",
	".moai/config/sections/language.yaml",
	".claude/rules/moai/core/agent-common-protocol.md",
	".claude/rules/moai/core/askuser-protocol.md",
	".claude/rules/moai/core/moai-constitution.md",
	".claude/rules/moai/core/moai-mcp-tools.md",
	".claude/rules/moai/core/native-idiom-and-register.md",
	".claude/rules/moai/core/verification-claim-integrity.md",
	".claude/rules/moai/workflow/cache-aware-execution.md",
	".claude/rules/moai/workflow/context-window-management.md",
	".claude/rules/moai/workflow/cross-session-messaging.md",
	".claude/rules/moai/workflow/goal-directive.md",
	".claude/rules/moai/workflow/kanban-dispatch.md",
	".claude/rules/moai/workflow/main-checkout-branch-guard.md",
	".claude/rules/moai/workflow/session-handoff.md",
	".claude/rules/moai/workflow/skill-routing.md",
}

// copyHeadroomTree copies every regular file and directory under src to dst,
// skipping the .git directory.
func copyHeadroomTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		switch {
		case d.IsDir():
			return os.MkdirAll(target, 0o755)
		case d.Type().IsRegular():
			info, err := d.Info()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, info.Mode().Perm())
		case d.Type()&fs.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		default:
			return fmt.Errorf("copyHeadroomTree: unsupported file type at %s", rel)
		}
	})
}

// headroomFileSHA256 returns the hex sha256 of path, or "" when absent.
func headroomFileSHA256(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// headroomExportMarker is written into every export so a later run can prove
// the directory it is about to delete is its own earlier output.
const headroomExportMarker = ".headroom-export-marker"

func writeHeadroomExportMarker(dir string) error {
	return os.WriteFile(filepath.Join(dir, headroomExportMarker), []byte("TestHeadroomInitSurfaceExport\n"), 0o644)
}

// clearHeadroomExportDir removes dir only when it is absent, an empty
// directory, or a directory carrying headroomExportMarker. Anything else is
// refused: the path comes from a flag and may name data the harness never
// wrote.
func clearHeadroomExportDir(dir string) error {
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat export dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("export path %s exists and is not a directory; refusing to delete", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read export dir: %w", err)
	}
	if len(entries) > 0 {
		marker, err := os.Lstat(filepath.Join(dir, headroomExportMarker))
		if err != nil || !marker.Mode().IsRegular() {
			return fmt.Errorf("export dir %s is non-empty and lacks %s; refusing to delete", dir, headroomExportMarker)
		}
	}
	return os.RemoveAll(dir)
}

func TestHeadroomInitSurfaceExport(t *testing.T) {
	if *headroomExportDir == "" {
		t.Skip("-headroom-export not set; measurement harness only")
	}
	exportDir, err := filepath.Abs(*headroomExportDir)
	if err != nil {
		t.Fatalf("resolve export dir: %v", err)
	}

	prepareSafeInitHome(t)
	// The default user gets the slim install; the env switch must not widen it.
	t.Setenv("MOAI_DISTRIBUTE_ALL", "")
	// No network update check.
	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	root := t.TempDir()
	if _, stderr, err := runInitWithFlags(t, root, map[string]string{
		"name": "headroom-probe",
		"llm":  "claude",
	}); err != nil {
		t.Fatalf("init: %v (stderr: %s)", err, stderr)
	}

	if err := clearHeadroomExportDir(exportDir); err != nil {
		t.Fatalf("clear export dir: %v", err)
	}
	if err := copyHeadroomTree(root, exportDir); err != nil {
		t.Fatalf("export: %v", err)
	}
	if err := writeHeadroomExportMarker(exportDir); err != nil {
		t.Fatalf("write export marker: %v", err)
	}

	for _, rel := range headroomCountPaths {
		src := headroomFileSHA256(t, filepath.Join(root, filepath.FromSlash(rel)))
		got := headroomFileSHA256(t, filepath.Join(exportDir, filepath.FromSlash(rel)))
		if got != src {
			t.Errorf("%s: exported sha256 %q differs from init output %q", rel, got, src)
		}
		present := "present"
		if src == "" {
			present = "absent"
		}
		t.Logf("headroom-path\t%s\t%s\t%s", rel, present, src)
	}
	if _, err := os.Stat(filepath.Join(exportDir, ".git")); err == nil {
		t.Errorf("export contains .git; it must be excluded")
	}
}

// TestHeadroomExportDirGuard proves the export-dir guard refuses to delete a
// directory the harness did not create. It needs no -headroom-export flag.
func TestHeadroomExportDirGuard(t *testing.T) {
	t.Run("refuses non-empty unmarked dir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "export")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		keep := filepath.Join(dir, "precious.txt")
		if err := os.WriteFile(keep, []byte("not harness output"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := clearHeadroomExportDir(dir); err == nil {
			t.Fatalf("guard deleted a non-empty unmarked dir; want refusal")
		}
		if _, err := os.Stat(keep); err != nil {
			t.Fatalf("unmarked content was removed: %v", err)
		}
	})
	t.Run("allows absent path", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "absent")
		if err := clearHeadroomExportDir(dir); err != nil {
			t.Fatalf("absent path: %v", err)
		}
	})
	t.Run("clears empty dir", func(t *testing.T) {
		dir := t.TempDir()
		if err := clearHeadroomExportDir(dir); err != nil {
			t.Fatalf("empty dir: %v", err)
		}
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatalf("empty dir still present: %v", err)
		}
	})
	t.Run("clears dir carrying the harness marker", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeHeadroomExportMarker(dir); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := clearHeadroomExportDir(dir); err != nil {
			t.Fatalf("marked dir: %v", err)
		}
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatalf("marked dir still present: %v", err)
		}
	})
}
