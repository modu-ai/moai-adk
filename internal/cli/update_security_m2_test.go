// SPEC-INTERNAL-SECURITY-001 M2 (template-update group) regression tests.
//
// Covers:
//   - REQ-SEC-003: symlink dereference backup leak — collectUserOwnedFiles
//     must skip symlinks so copyFile never records the link's dereferenced
//     target content into .moai/backups/.
//   - REQ-SEC-005: conservative backup expansion — reserved-prefix names that
//     are ambiguous (could be user-authored) are included in the backup pass.
//   - REQ-SEC-006: assertNoUserOwnedNamespaceTouch is wired as a real
//     pre-modification abort gate via verifyNamespaceBackupCoverage.
//
// All symlink/secret fixtures live inside t.TempDir() only (NFR-SEC-004).
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// REQ-SEC-003 / AC-SEC-003a: a symlink inside a user-owned namespace must NOT
// have its dereferenced target content recorded into the backup directory.
func TestBackupUserOwnedNamespace_SymlinkTargetNotBackedUp(t *testing.T) {
	projectRoot := t.TempDir()

	// Secret lives OUTSIDE the scanned namespace roots so the only way it can
	// reach the backup is via a dereferenced symlink.
	secretPath := filepath.Join(projectRoot, "secret.pem")
	const secretContent = "PRIVATE KEY CONTENT FOR REQ-SEC-003"
	if err := os.WriteFile(secretPath, []byte(secretContent), 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	// Symlink inside .moai/harness (a user-owned namespace root) pointing at
	// the secret. A regular file is also placed alongside so a backup directory
	// IS created — the assertion then verifies the symlink target did not leak
	// into that backup.
	harnessDir := filepath.Join(projectRoot, ".moai", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("mkdir harness: %v", err)
	}
	if err := os.WriteFile(filepath.Join(harnessDir, "real.md"), []byte("regular user content"), 0o644); err != nil {
		t.Fatalf("write regular: %v", err)
	}
	linkPath := filepath.Join(harnessDir, "leak")
	if err := os.Symlink(secretPath, linkPath); err != nil {
		// Symlinks may be unsupported (e.g. Windows without privileges); skip
		// rather than fail on environments that cannot create them.
		t.Skipf("symlink unsupported on this host: %v", err)
	}

	backupDir, err := backupUserOwnedNamespace(projectRoot)
	if err != nil {
		t.Fatalf("backupUserOwnedNamespace error: %v", err)
	}
	if backupDir == "" {
		t.Fatal("expected a backup directory (regular user-owned file present), got empty")
	}

	// Walk the ENTIRE backup tree and assert the secret content never appears.
	leaked := false
	walkErr := filepath.WalkDir(backupDir, func(path string, d os.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr == nil && strings.Contains(string(data), secretContent) {
			leaked = true
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk backup dir: %v", walkErr)
	}
	if leaked {
		t.Error("REQ-SEC-003 violation: symlink dereferenced target content " +
			"leaked into .moai/backups/ — collectUserOwnedFiles must skip symlinks")
	}

	// The symlink entry itself must not appear as a backed-up file either.
	leakBackup := filepath.Join(backupDir, ".moai/harness/leak")
	if _, statErr := os.Lstat(leakBackup); statErr == nil {
		t.Errorf("symlink entry was recorded in backup at %s", leakBackup)
	}
}

// REQ-SEC-003 / AC-SEC-003b: collectUserOwnedFiles skips symlinks (Lstat guard).
func TestCollectUserOwnedFiles_SymlinkSkipped(t *testing.T) {
	projectRoot := t.TempDir()

	secretPath := filepath.Join(projectRoot, "outside-secret")
	if err := os.WriteFile(secretPath, []byte("topsecret"), 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	harnessDir := filepath.Join(projectRoot, ".moai", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("mkdir harness: %v", err)
	}
	regularPath := filepath.Join(harnessDir, "real.md")
	if err := os.WriteFile(regularPath, []byte("regular content"), 0o644); err != nil {
		t.Fatalf("write regular: %v", err)
	}
	linkPath := filepath.Join(harnessDir, "link.md")
	if err := os.Symlink(secretPath, linkPath); err != nil {
		t.Skipf("symlink unsupported on this host: %v", err)
	}

	files, err := collectUserOwnedFiles(projectRoot)
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}

	sawLink := false
	sawRegular := false
	for _, f := range files {
		if strings.HasSuffix(f, "link.md") {
			sawLink = true
		}
		if strings.HasSuffix(f, "real.md") {
			sawRegular = true
		}
	}
	if sawLink {
		t.Error("symlink was collected — collectUserOwnedFiles must skip symlinks (REQ-SEC-003)")
	}
	if !sawRegular {
		t.Error("regular file was not collected — symlink guard must not skip regular files (NFR-SEC-003)")
	}
}

// REQ-SEC-005 / AC-SEC-005a: reserved-prefix ambiguous names are included in
// the conservative backup pass so they are never overwritten/deleted without
// a backup.
func TestBackupUserOwnedNamespace_ConservativeReservedPrefix(t *testing.T) {
	projectRoot := nsTestFixture(t, map[string]string{
		".claude/skills/moai-my-notes/SKILL.md": "user-authored notes skill",
		".claude/agents/expert-mydomain.md":     "user-authored agent",
	})

	backupDir, err := backupUserOwnedNamespace(projectRoot)
	if err != nil {
		t.Fatalf("backup error: %v", err)
	}
	if backupDir == "" {
		t.Fatal("expected backup for reserved-prefix ambiguous files (REQ-SEC-005)")
	}

	must := []string{
		".claude/skills/moai-my-notes/SKILL.md",
		".claude/agents/expert-mydomain.md",
	}
	for _, rel := range must {
		backupCopy := filepath.Join(backupDir, rel)
		data, statErr := os.ReadFile(backupCopy)
		if statErr != nil {
			t.Errorf("conservative backup missing %s: %v", rel, statErr)
			continue
		}
		// Verify byte-identical content so we know it was a real copy.
		src, srcErr := os.ReadFile(filepath.Join(projectRoot, rel))
		if srcErr != nil {
			t.Fatalf("read source %s: %v", rel, srcErr)
		}
		if string(data) != string(src) {
			t.Errorf("content mismatch for %s", rel)
		}
	}
}

// REQ-SEC-005 / AC-SEC-005b: the conservative classifier's prefix-collision
// behavior is documented via explicit table-driven cases.
func TestIsUserOwnedNamespaceConservative_PrefixCollision(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		want bool
	}{
		// Strict user-owned (unchanged) — conservative is a superset.
		{"harness skill", ".claude/skills/harness-x/SKILL.md", true},
		{"custom skill", ".claude/skills/my-skill/SKILL.md", true},
		{".moai/harness", ".moai/harness/main.md", true},

		// Ambiguous reserved-prefix — conservative INCLUSION (REQ-SEC-005).
		{"moai- skill (ambiguous)", ".claude/skills/moai-my-notes/SKILL.md", true},
		{"moai skill dir (ambiguous)", ".claude/skills/moai/SKILL.md", true},
		{"expert- agent file (ambiguous)", ".claude/agents/expert-mydomain.md", true},
		{"manager- agent file (ambiguous)", ".claude/agents/manager-notes.md", true},
		{"builder- agent file (ambiguous)", ".claude/agents/builder-foo.md", true},
		{"evaluator- agent file (ambiguous)", ".claude/agents/evaluator-bar.md", true},

		// System agent directories remain clearly MoAI-managed (NOT ambiguous).
		{"core agent dir (system)", ".claude/agents/core/manager-develop.md", false},
		{"expert agent dir (system)", ".claude/agents/expert/sync-auditor.md", false},
		{"meta agent dir (system)", ".claude/agents/meta/plan-auditor.md", false},

		// Config paths are not user-owned namespace.
		{"moai config", ".moai/config/sections/quality.yaml", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUserOwnedNamespaceConservative(tt.rel)
			if got != tt.want {
				t.Errorf("isUserOwnedNamespaceConservative(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

// REQ-SEC-006 / AC-SEC-006a: verifyNamespaceBackupCoverage is the wired
// pre-modification abort gate. It must pass when user-owned content is fully
// backed up (NFR-SEC-003) and abort with UPDATE_USER_NAMESPACE_VIOLATION when
// a user-owned file on disk lacks a backup.
func TestVerifyNamespaceBackupCoverage(t *testing.T) {
	t.Run("no user-owned content passes", func(t *testing.T) {
		projectRoot := t.TempDir() // empty
		if err := verifyNamespaceBackupCoverage(projectRoot, ""); err != nil {
			t.Errorf("expected nil for empty project, got: %v", err)
		}
	})

	t.Run("user-owned fully backed up passes", func(t *testing.T) {
		projectRoot := nsTestFixture(t, map[string]string{
			".moai/harness/main.md": "content",
		})
		backupDir, err := backupUserOwnedNamespace(projectRoot)
		if err != nil {
			t.Fatalf("backup: %v", err)
		}
		if backupDir == "" {
			t.Fatal("expected backup dir")
		}
		if err := verifyNamespaceBackupCoverage(projectRoot, backupDir); err != nil {
			t.Errorf("expected nil when fully backed up (NFR-SEC-003), got: %v", err)
		}
	})

	t.Run("user-owned missing from backup aborts", func(t *testing.T) {
		projectRoot := nsTestFixture(t, map[string]string{
			".moai/harness/main.md": "content",
		})
		// Pass an empty/missing backup dir — the user-owned file is on disk but
		// not backed up, so the gate must abort before destructive ops.
		err := verifyNamespaceBackupCoverage(projectRoot, "")
		if err == nil {
			t.Fatal("expected UPDATE_USER_NAMESPACE_VIOLATION, got nil")
		}
		if !strings.Contains(err.Error(), "UPDATE_USER_NAMESPACE_VIOLATION") {
			t.Errorf("expected sentinel literal, got: %s", err.Error())
		}
	})

	t.Run("user-owned in backup dir missing one file aborts", func(t *testing.T) {
		projectRoot := nsTestFixture(t, map[string]string{
			".moai/harness/a.md": "a",
			".moai/harness/b.md": "b",
		})
		// Manually create a backup dir that contains only a.md (b.md missing).
		partialBackup := filepath.Join(projectRoot, ".moai", "backups", "partial")
		if err := os.MkdirAll(filepath.Join(partialBackup, ".moai", "harness"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(partialBackup, ".moai", "harness", "a.md"), []byte("a"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		err := verifyNamespaceBackupCoverage(projectRoot, partialBackup)
		if err == nil {
			t.Fatal("expected abort for missing b.md, got nil")
		}
		if !strings.Contains(err.Error(), "UPDATE_USER_NAMESPACE_VIOLATION") {
			t.Errorf("expected sentinel literal, got: %s", err.Error())
		}
	})
}
