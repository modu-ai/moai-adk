package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// --- test helpers ---

// mkdirConfigTree creates <tmp>/.moai/config/sections and returns the project
// root (tmp). Every AC test builds its fixture underneath this root.
func mkdirConfigTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir config tree: %v", err)
	}
	return root
}

// writeFileAt writes a regular file at root/rel with the EXACT mode (chmod after
// write defeats umask), creating parent dirs as needed.
func writeFileAt(t *testing.T, root, rel string, mode os.FileMode) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), defs.DirPerm); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte("key: value\n"), mode); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	if err := os.Chmod(p, mode); err != nil {
		t.Fatalf("chmod %s to %v: %v", rel, mode, err)
	}
}

// assertMode asserts the on-disk file mode equals want.
func assertMode(t *testing.T, label string, p string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat %s (%s): %v", label, p, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s mode: got %v, want %v", label, got, want)
	}
}

// --- AC-MIG-001: dry-run default lists candidates, modes unchanged ---

func TestModeMigrateDryRun_NoOp_OnDisk(t *testing.T) {
	// A narrowed file (0600) and a canonical file (0644) both under .moai/config.
	root := mkdirConfigTree(t)
	narrowed := filepath.Join(root, ".moai", "config", "sections", "narrowed.yaml")
	canonical := filepath.Join(root, ".moai", "config", "sections", "canonical.yaml")
	writeFileAt(t, root, ".moai/config/sections/narrowed.yaml", 0o600)
	writeFileAt(t, root, ".moai/config/sections/canonical.yaml", 0o644)

	var buf bytes.Buffer
	configDir := filepath.Join(root, ".moai", "config")
	if err := runModeMigrate(&buf, configDir, false /* dry-run */); err != nil {
		t.Fatalf("runModeMigrate dry-run: %v", err)
	}

	out := buf.String()
	// Dry-run MUST list the narrowed candidate with its path + current + target.
	if !strings.Contains(out, "narrowed.yaml") {
		t.Errorf("dry-run output missing candidate path:\n%s", out)
	}
	if !strings.Contains(out, "0600") {
		t.Errorf("dry-run output missing current mode 0600:\n%s", out)
	}
	if !strings.Contains(out, "0644") {
		t.Errorf("dry-run output missing target mode 0644:\n%s", out)
	}
	if !strings.Contains(out, "--apply") {
		t.Errorf("dry-run output missing --apply footer hint:\n%s", out)
	}
	if !strings.Contains(out, "No files were modified") {
		t.Errorf("dry-run output missing no-modification announcement:\n%s", out)
	}

	// Modes on disk MUST be unchanged (0600 stays 0600; 0644 stays 0644).
	assertMode(t, "narrowed (post-dry-run)", narrowed, 0o600)
	assertMode(t, "canonical (post-dry-run)", canonical, 0o644)
}

// --- AC-MIG-002: apply widens 0600 → defs.FilePerm ---

func TestModeMigrateApply_WidensNarrowed(t *testing.T) {
	root := mkdirConfigTree(t)
	narrowed := filepath.Join(root, ".moai", "config", "sections", "llm.yaml")
	writeFileAt(t, root, ".moai/config/sections/llm.yaml", 0o600)

	var buf bytes.Buffer
	configDir := filepath.Join(root, ".moai", "config")
	if err := runModeMigrate(&buf, configDir, true /* apply */); err != nil {
		t.Fatalf("runModeMigrate apply: %v", err)
	}

	// On-disk mode MUST now be defs.FilePerm (0644).
	assertMode(t, "narrowed (post-apply)", narrowed, defs.FilePerm)
}

// --- AC-MIG-003: only-widen — 0644 unchanged, 0600 → 0644 ---

func TestModeMigrateApply_OnlyWidens(t *testing.T) {
	root := mkdirConfigTree(t)
	canonical := filepath.Join(root, ".moai", "config", "sections", "canonical.yaml")
	narrowed := filepath.Join(root, ".moai", "config", "sections", "narrowed.yaml")
	writeFileAt(t, root, ".moai/config/sections/canonical.yaml", 0o644)
	writeFileAt(t, root, ".moai/config/sections/narrowed.yaml", 0o600)

	var buf bytes.Buffer
	configDir := filepath.Join(root, ".moai", "config")
	if err := runModeMigrate(&buf, configDir, true); err != nil {
		t.Fatalf("runModeMigrate apply: %v", err)
	}

	// Canonical stays exactly at defs.FilePerm — not widened beyond, not narrowed.
	assertMode(t, "canonical (post-apply)", canonical, defs.FilePerm)
	// Narrowed widens to defs.FilePerm.
	assertMode(t, "narrowed (post-apply)", narrowed, defs.FilePerm)
}

// --- AC-MIG-004: scope — a file OUTSIDE .moai/config is never touched ---

func TestModeMigrateApply_ScopeConfigOnly(t *testing.T) {
	root := mkdirConfigTree(t)
	// An outside-config file at 0600 (simulates .claude/settings.json or /tmp file).
	outside := filepath.Join(root, ".claude", "settings.json")
	writeFileAt(t, root, ".claude/settings.json", 0o600)
	// Plus a real candidate inside .moai/config to ensure apply actually ran.
	writeFileAt(t, root, ".moai/config/sections/narrowed.yaml", 0o600)

	var buf bytes.Buffer
	configDir := filepath.Join(root, ".moai", "config")
	if err := runModeMigrate(&buf, configDir, true); err != nil {
		t.Fatalf("runModeMigrate apply: %v", err)
	}

	// Outside file MUST remain 0600 — migration never touches paths outside .moai/config.
	assertMode(t, "outside-config (post-apply)", outside, 0o600)
}

// --- AC-MIG-005: idempotent — already-canonical tree → empty list, apply no-op ---

func TestModeMigrate_Idempotent(t *testing.T) {
	root := mkdirConfigTree(t)
	// Every file already at defs.FilePerm — migration has "already been applied".
	writeFileAt(t, root, ".moai/config/sections/a.yaml", defs.FilePerm)
	writeFileAt(t, root, ".moai/config/sections/b.yaml", defs.FilePerm)

	t.Run("dry-run_empty_candidate_list", func(t *testing.T) {
		var buf bytes.Buffer
		configDir := filepath.Join(root, ".moai", "config")
		if err := runModeMigrate(&buf, configDir, false); err != nil {
			t.Fatalf("runModeMigrate dry-run: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "0 candidate") {
			t.Errorf("dry-run on already-applied tree should report 0 candidates:\n%s", out)
		}
	})

	t.Run("apply_no_op", func(t *testing.T) {
		// Capture pre-apply modes; re-run with --apply; assert no change + exit 0.
		aPath := filepath.Join(root, ".moai", "config", "sections", "a.yaml")
		assertMode(t, "a (pre)", aPath, defs.FilePerm)
		var buf bytes.Buffer
		configDir := filepath.Join(root, ".moai", "config")
		if err := runModeMigrate(&buf, configDir, true); err != nil {
			t.Fatalf("runModeMigrate apply on idempotent tree returned err: %v", err)
		}
		assertMode(t, "a (post)", aPath, defs.FilePerm)
	})
}

// --- AC-MIG-006: routes through atomicfile.Write; no bare os.WriteFile ---

func TestModeMigrate_HelperRouting(t *testing.T) {
	src, err := os.ReadFile("mode_migrate.go")
	if err != nil {
		t.Fatalf("read mode_migrate.go: %v", err)
	}
	body := string(src)

	// The apply path MUST route through the shared atomic-write helper.
	if !strings.Contains(body, "atomicfile.Write") {
		t.Errorf("mode_migrate.go does not call atomicfile.Write — AC-MIG-006 violation")
	}
	// Bare os.WriteFile on the destination path is PROHIBITED.
	// (atomicfile.Write internally uses os.CreateTemp, not os.WriteFile directly.)
	if strings.Contains(body, "os.WriteFile(") {
		t.Errorf("mode_migrate.go contains bare os.WriteFile( call — AC-MIG-006 violation")
	}
	// The single permitted os.Chmod site MUST reference the named constant, not a literal.
	if !strings.Contains(body, "os.Chmod(") {
		t.Errorf("mode_migrate.go missing the permitted os.Chmod widening site")
	}
	// A hardcoded literal mode in the chmod would violate CLAUDE.local.md §14.
	for _, lit := range []string{"0o644", "0o600", "0o750", "0644", "0600"} {
		if strings.Contains(body, lit) {
			t.Errorf("mode_migrate.go contains hardcoded mode literal %q — §14 violation", lit)
		}
	}
}

// --- AC-MIG-007: 0700 (non-subset) is excluded — never narrowed ---

func TestModeMigrateApply_NonSubsetMode_Unchanged(t *testing.T) {
	root := mkdirConfigTree(t)
	// 0700 carries owner-exec (0100) which is NOT in 0644 → not a subset → excluded.
	execFile := filepath.Join(root, ".moai", "config", "sections", "exec.yaml")
	writeFileAt(t, root, ".moai/config/sections/exec.yaml", 0o700)
	// Plus a genuine candidate to ensure apply runs.
	writeFileAt(t, root, ".moai/config/sections/narrowed.yaml", 0o600)

	t.Run("apply_leaves_0700_unchanged", func(t *testing.T) {
		var buf bytes.Buffer
		configDir := filepath.Join(root, ".moai", "config")
		if err := runModeMigrate(&buf, configDir, true); err != nil {
			t.Fatalf("runModeMigrate apply: %v", err)
		}
		assertMode(t, "0700 file (post-apply)", execFile, 0o700)
	})

	t.Run("dry-run_excludes_0700_from_candidates", func(t *testing.T) {
		// Fresh tree: the sibling apply subtest mutated the shared tree above, so
		// rebuild the (0600 + 0700) pair here to assert the dry-run view cleanly.
		r := mkdirConfigTree(t)
		writeFileAt(t, r, ".moai/config/sections/exec.yaml", 0o700)
		writeFileAt(t, r, ".moai/config/sections/narrowed.yaml", 0o600)

		var buf bytes.Buffer
		configDir := filepath.Join(r, ".moai", "config")
		if err := runModeMigrate(&buf, configDir, false); err != nil {
			t.Fatalf("runModeMigrate dry-run: %v", err)
		}
		out := buf.String()
		// The 0700 file MUST NOT appear as a candidate row. Only the genuine
		// subset file (narrowed.yaml at 0600) counts toward the candidate total.
		// The candidate count line is the load-bearing assertion: exactly 1.
		if !strings.Contains(out, "1 candidate(s) found") {
			t.Errorf("dry-run should report exactly 1 candidate (the 0600 file), "+
				"excluding the non-subset 0700 file:\n%s", out)
		}
		// Predicate unit check: 0700 is NOT a candidate.
		if IsWideningCandidate(0o700) {
			t.Errorf("IsWideningCandidate(0700) = true; want false (0700 is not a subset of 0644)")
		}
	})
}

// --- AC-MIG-008: symlink under .moai/config is Lstat-detected + skipped ---

func TestModeMigrate_SymlinkSkipped(t *testing.T) {
	root := mkdirConfigTree(t)
	// External target OUTSIDE .moai/config, held at 0600.
	outsideTarget := filepath.Join(root, "outside", "target.yaml")
	writeFileAt(t, root, "outside/target.yaml", 0o600)
	// Symlink under .moai/config pointing at the external target.
	linkPath := filepath.Join(root, ".moai", "config", "sections", "link.yaml")
	if err := os.Symlink(outsideTarget, linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	t.Run("dry-run_reports_skipped_symlink", func(t *testing.T) {
		var buf bytes.Buffer
		configDir := filepath.Join(root, ".moai", "config")
		if err := runModeMigrate(&buf, configDir, false); err != nil {
			t.Fatalf("runModeMigrate dry-run: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "symlink") {
			t.Errorf("dry-run output should report the skipped symlink:\n%s", out)
		}
	})

	t.Run("apply_does_not_chmod_external_target", func(t *testing.T) {
		var buf bytes.Buffer
		configDir := filepath.Join(root, ".moai", "config")
		if err := runModeMigrate(&buf, configDir, true); err != nil {
			t.Fatalf("runModeMigrate apply: %v", err)
		}
		// The external target MUST remain at 0600 — no os.Chmod landed on it.
		assertMode(t, "external symlink target (post-apply)", outsideTarget, 0o600)
	})
}

// --- predicate unit tests (spec.md §D.2 enumeration, verbatim) ---

func TestIsWideningCandidate_Enumeration(t *testing.T) {
	cases := []struct {
		mode os.FileMode
		want bool
		why  string
	}{
		{0o600, true, "proper subset of 0644"},
		{0o640, true, "proper subset of 0644"},
		{0o700, false, "owner-exec not in 0644 → would narrow"},
		{0o660, false, "group-write not in 0644 → would narrow"},
		{0o644, false, "already canonical"},
		{0o664, false, "superset of 0644, not a subset"},
		{0o666, false, "superset of 0644, not a subset"},
		{0o500, false, "owner-exec not in 0644 → would narrow"},
	}
	for _, c := range cases {
		if got := IsWideningCandidate(c.mode); got != c.want {
			t.Errorf("IsWideningCandidate(%o) = %v, want %v (%s)", c.mode, got, c.want, c.why)
		}
	}
}
