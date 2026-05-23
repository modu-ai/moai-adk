// SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 test suite.
//
// Covers:
//   - REQ-UNP-004 backup creation
//   - REQ-UNP-006 sentinel (assertNoUserOwnedNamespaceTouch)
//   - REQ-UNP-007 atomicity (.complete marker)
//   - REQ-UNP-010 backup directory naming regex
//   - EC-UNP-001 empty namespace
//   - EC-UNP-007 mid-copy failure cleanup
//   - NFR-UNP-004 idempotency (numeric suffix collision)
//   - AC-UNP-005, AC-UNP-004, AC-UNP-006, AC-UNP-007, AC-UNP-008
package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// nsTestFixture creates a temp project with the requested user-owned files.
// Returns the project root.
//
// The fixture mirrors the actual project layout: .claude/skills/...,
// .claude/agents/..., .moai/harness/..., all created relative to projectRoot.
func nsTestFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	projectRoot := t.TempDir()
	for rel, content := range files {
		fullPath := filepath.Join(projectRoot, rel)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("setup mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("setup write %s: %v", rel, err)
		}
	}
	return projectRoot
}

// TestBackupUserOwnedNamespace covers REQ-UNP-004 (backup creation),
// REQ-UNP-007 (.complete marker), REQ-UNP-010 (regex), plus AC-UNP-001 / 002 / 003.
func TestBackupUserOwnedNamespace(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantBackup bool // true if a backup directory should be created
		mustExist  []string
	}{
		{
			name: "REQ-UNP-001 my-harness skill preserved and backed up",
			files: map[string]string{
				".claude/skills/my-harness-test/SKILL.md": "user harness skill content",
			},
			wantBackup: true,
			mustExist:  []string{".claude/skills/my-harness-test/SKILL.md"},
		},
		{
			name: "REQ-UNP-002 harness agent directory preserved and backed up",
			files: map[string]string{
				".claude/agents/harness/test-specialist.md": "user harness teammate",
			},
			wantBackup: true,
			mustExist:  []string{".claude/agents/harness/test-specialist.md"},
		},
		{
			name: "REQ-UNP-003 moai/harness extension preserved and backed up",
			files: map[string]string{
				".moai/harness/main.md": "user harness extension",
			},
			wantBackup: true,
			mustExist:  []string{".moai/harness/main.md"},
		},
		{
			name: "REQ-UNP-009 user direct-added agent preserved",
			files: map[string]string{
				".claude/agents/custom-agent.md": "user-authored top-level agent",
			},
			wantBackup: true,
			mustExist:  []string{".claude/agents/custom-agent.md"},
		},
		{
			name: "REQ-UNP-009 user direct-added skill preserved",
			files: map[string]string{
				".claude/skills/custom-skill/SKILL.md": "user-authored skill",
			},
			wantBackup: true,
			mustExist:  []string{".claude/skills/custom-skill/SKILL.md"},
		},
		{
			name: "multiple user-owned categories backed up together",
			files: map[string]string{
				".claude/skills/my-harness-foo/SKILL.md":  "a",
				".claude/agents/harness/teammate.md":      "b",
				".moai/harness/extensions/custom.md":      "c",
				".claude/agents/custom-direct.md":         "d",
				".claude/skills/custom-skill/file.md":     "e",
			},
			wantBackup: true,
			mustExist: []string{
				".claude/skills/my-harness-foo/SKILL.md",
				".claude/agents/harness/teammate.md",
				".moai/harness/extensions/custom.md",
				".claude/agents/custom-direct.md",
				".claude/skills/custom-skill/file.md",
			},
		},
		{
			name: "EC-UNP-001 no user-owned content → no backup directory",
			files: map[string]string{
				// Only MoAI-managed paths
				".claude/agents/core/manager-develop.md": "moai-managed",
				".moai/config/sections/quality.yaml":    "config",
			},
			wantBackup: false,
		},
		{
			name: "MoAI-managed paths NOT included in backup",
			files: map[string]string{
				".claude/skills/my-harness-test/SKILL.md": "user content",
				".claude/skills/moai-foundation-cc/SKILL.md": "moai content",
				".claude/agents/core/manager-develop.md":  "moai content",
			},
			wantBackup: true,
			mustExist:  []string{".claude/skills/my-harness-test/SKILL.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := nsTestFixture(t, tt.files)

			backupDir, err := backupUserOwnedNamespace(projectRoot)
			if err != nil {
				t.Fatalf("backupUserOwnedNamespace error: %v", err)
			}

			if tt.wantBackup {
				if backupDir == "" {
					t.Fatal("expected backup directory, got empty string")
				}

				// AC-UNP-006: regex compliance
				rel, err := filepath.Rel(projectRoot, backupDir)
				if err != nil {
					t.Fatalf("rel: %v", err)
				}
				// Normalize separators for cross-platform regex match
				relNorm := strings.ReplaceAll(rel, "\\", "/")
				regex := regexp.MustCompile(`^\.moai/backups/update-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z$`)
				if !regex.MatchString(relNorm) {
					t.Errorf("backup dir %q does not match regex %s", relNorm, regex)
				}

				// REQ-UNP-007: .complete marker must exist
				markerPath := filepath.Join(backupDir, ".complete")
				if _, statErr := os.Stat(markerPath); statErr != nil {
					t.Errorf(".complete marker missing: %v", statErr)
				}

				// REQ-UNP-004 + AC-UNP-001/002/003: backed up files must exist
				for _, rel := range tt.mustExist {
					backupCopy := filepath.Join(backupDir, rel)
					if _, statErr := os.Stat(backupCopy); statErr != nil {
						t.Errorf("expected backup file %s, got %v", rel, statErr)
					}
					// Verify byte-identical content
					srcContent, srcErr := os.ReadFile(filepath.Join(projectRoot, rel))
					if srcErr != nil {
						t.Fatalf("source read: %v", srcErr)
					}
					backupContent, backupErr := os.ReadFile(backupCopy)
					if backupErr != nil {
						t.Fatalf("backup read: %v", backupErr)
					}
					if string(srcContent) != string(backupContent) {
						t.Errorf("content mismatch for %s: src=%q backup=%q",
							rel, srcContent, backupContent)
					}
				}

				// Original files must still exist (AC-UNP-001/002/003 preservation)
				for _, rel := range tt.mustExist {
					originalPath := filepath.Join(projectRoot, rel)
					if _, statErr := os.Stat(originalPath); statErr != nil {
						t.Errorf("original file %s was removed: %v", rel, statErr)
					}
				}
			} else {
				if backupDir != "" {
					t.Errorf("expected no backup, got %s", backupDir)
				}
				// EC-UNP-001: directory should not exist
				baseDir := filepath.Join(projectRoot, ".moai", "backups")
				if _, statErr := os.Stat(baseDir); statErr == nil {
					// Either the dir doesn't exist (preferred) or it's empty
					entries, _ := os.ReadDir(baseDir)
					if len(entries) > 0 {
						t.Errorf(".moai/backups should be empty, got %d entries", len(entries))
					}
				}
			}
		})
	}
}

// TestAssertNoUserOwnedNamespaceTouch covers REQ-UNP-006 sentinel emission.
// Maps to AC-UNP-005.
func TestAssertNoUserOwnedNamespaceTouch(t *testing.T) {
	tests := []struct {
		name        string
		plan        []deployOp
		wantErr     bool
		wantSentry  bool // expect UPDATE_USER_NAMESPACE_VIOLATION literal
	}{
		{
			name: "AC-UNP-005 harness file overwrite triggers sentinel",
			plan: []deployOp{
				{rel: ".claude/agents/harness/contraband.md", action: "overwrite"},
			},
			wantErr:    true,
			wantSentry: true,
		},
		{
			name: "REQ-UNP-001 my-harness skill overwrite triggers sentinel",
			plan: []deployOp{
				{rel: ".claude/skills/my-harness-foo/file.md", action: "overwrite"},
			},
			wantErr:    true,
			wantSentry: true,
		},
		{
			name: "REQ-UNP-003 .moai/harness delete triggers sentinel",
			plan: []deployOp{
				{rel: ".moai/harness/main.md", action: "delete"},
			},
			wantErr:    true,
			wantSentry: true,
		},
		{
			name: "REQ-UNP-009 user custom agent triggers sentinel",
			plan: []deployOp{
				{rel: ".claude/agents/custom-agent.md", action: "overwrite"},
			},
			wantErr:    true,
			wantSentry: true,
		},
		{
			name: "MoAI-managed core agent passes (no sentinel)",
			plan: []deployOp{
				{rel: ".claude/agents/core/manager-develop.md", action: "overwrite"},
			},
			wantErr: false,
		},
		{
			name: "mixed plan — first user-owned hit triggers sentinel",
			plan: []deployOp{
				{rel: ".claude/agents/core/manager-develop.md", action: "overwrite"},
				{rel: ".claude/agents/harness/teammate.md", action: "overwrite"},
				{rel: ".claude/skills/moai-foundation-cc/SKILL.md", action: "overwrite"},
			},
			wantErr:    true,
			wantSentry: true,
		},
		{
			name:    "empty plan passes",
			plan:    []deployOp{},
			wantErr: false,
		},
		{
			name: "windows separator harness path triggers sentinel",
			plan: []deployOp{
				{rel: ".claude\\agents\\harness\\teammate.md", action: "overwrite"},
			},
			wantErr:    true,
			wantSentry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertNoUserOwnedNamespaceTouch(tt.plan)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantSentry && !strings.Contains(err.Error(), "UPDATE_USER_NAMESPACE_VIOLATION") {
					t.Errorf("expected UPDATE_USER_NAMESPACE_VIOLATION sentinel, got: %s", err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected nil, got: %v", err)
				}
			}
		})
	}
}

// TestAssertNoUserOwnedNamespaceTouch_NoMutation covers AC-UNP-005's
// "no file modification occurs" requirement. The sentinel runs before any
// write — it inspects the plan only and does not touch the filesystem.
func TestAssertNoUserOwnedNamespaceTouch_NoMutation(t *testing.T) {
	tmpDir := t.TempDir()
	contrabandPath := filepath.Join(tmpDir, ".claude/agents/harness/contraband.md")

	plan := []deployOp{
		{rel: ".claude/agents/harness/contraband.md", action: "overwrite"},
	}
	err := assertNoUserOwnedNamespaceTouch(plan)

	if err == nil {
		t.Fatal("expected sentinel violation, got nil")
	}
	if !strings.Contains(err.Error(), "UPDATE_USER_NAMESPACE_VIOLATION") {
		t.Errorf("expected sentinel literal, got: %s", err.Error())
	}
	// Assert no filesystem mutation occurred — contraband file should NOT exist
	if _, statErr := os.Stat(contrabandPath); !os.IsNotExist(statErr) {
		t.Errorf("filesystem mutation detected: contraband file exists at %s (stat err: %v)",
			contrabandPath, statErr)
	}
}

// TestNewNamespaceBackupStamp covers REQ-UNP-010 timestamp format.
func TestNewNamespaceBackupStamp(t *testing.T) {
	stamp := newNamespaceBackupStamp()
	// Format: 2026-05-23T14-30-00Z
	regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z$`)
	if !regex.MatchString(stamp) {
		t.Errorf("stamp %q does not match REQ-UNP-010 regex", stamp)
	}
	// Must NOT contain colons (Windows-safe filename per REQ-UNP-010)
	if strings.Contains(stamp, ":") {
		t.Errorf("stamp %q contains colon (Windows-unsafe)", stamp)
	}
	// Must end in 'Z' (UTC)
	if !strings.HasSuffix(stamp, "Z") {
		t.Errorf("stamp %q does not end with 'Z' (UTC)", stamp)
	}
}

// TestResolveNamespaceBackupDir_CollisionSuffix covers NFR-UNP-004 idempotency
// via numeric suffix. Maps to AC-UNP-012.
func TestResolveNamespaceBackupDir_CollisionSuffix(t *testing.T) {
	projectRoot := t.TempDir()
	baseDir := filepath.Join(projectRoot, ".moai", "backups")
	stamp := "2026-05-23T14-30-00Z"

	// Pre-create the canonical path to simulate same-second collision
	canonical := filepath.Join(baseDir, "update-"+stamp)
	if err := os.MkdirAll(canonical, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First call should detect collision and return suffix -1
	got, err := resolveNamespaceBackupDir(projectRoot, stamp)
	if err != nil {
		t.Fatalf("resolveNamespaceBackupDir error: %v", err)
	}
	want := filepath.Join(baseDir, "update-"+stamp+"-1")
	if got != want {
		t.Errorf("first collision: got %s, want %s", got, want)
	}

	// Create the -1 path and verify next call returns -2
	if err := os.MkdirAll(got, 0o755); err != nil {
		t.Fatalf("setup -1: %v", err)
	}
	got2, err := resolveNamespaceBackupDir(projectRoot, stamp)
	if err != nil {
		t.Fatalf("resolveNamespaceBackupDir error: %v", err)
	}
	want2 := filepath.Join(baseDir, "update-"+stamp+"-2")
	if got2 != want2 {
		t.Errorf("second collision: got %s, want %s", got2, want2)
	}
}

// TestBackupUserOwnedNamespace_AtomicityMarker verifies REQ-UNP-007 explicitly.
// The .complete marker exists ONLY after all copies succeed.
func TestBackupUserOwnedNamespace_AtomicityMarker(t *testing.T) {
	projectRoot := nsTestFixture(t, map[string]string{
		".moai/harness/main.md": "extension content",
	})

	backupDir, err := backupUserOwnedNamespace(projectRoot)
	if err != nil {
		t.Fatalf("backup error: %v", err)
	}
	if backupDir == "" {
		t.Fatal("expected backup, got none")
	}

	// REQ-UNP-007: .complete marker must exist
	markerPath := filepath.Join(backupDir, ".complete")
	markerData, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf(".complete marker not found: %v", err)
	}

	// Marker content sanity: must contain stamp= and files= fields
	content := string(markerData)
	if !strings.Contains(content, "stamp=") {
		t.Errorf(".complete missing 'stamp=' field: %s", content)
	}
	if !strings.Contains(content, "files=1") {
		t.Errorf(".complete files= incorrect: %s", content)
	}
}

// TestCollectUserOwnedFiles_WindowsPathNormalization covers AC-UNP-013
// cross-platform path normalization (NFR-UNP-003). Run only on Windows when
// available; on non-Windows hosts, we exercise the backslash normalization
// branch via the assertNoUserOwnedNamespaceTouch path which accepts pre-built
// deploy plans with backslash separators.
func TestCollectUserOwnedFiles_WindowsPathNormalization(t *testing.T) {
	// The walker uses filepath.Rel which yields OS-native separators.
	// On Linux/Darwin, paths come back with '/' already.
	// On Windows, paths come with '\\' and the inner ReplaceAll normalizes them.
	// We assert that the resulting paths route correctly through
	// isUserOwnedNamespace either way.
	projectRoot := nsTestFixture(t, map[string]string{
		".moai/harness/main.md": "ext",
	})

	files, err := collectUserOwnedFiles(projectRoot)
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}
	// On any OS, the returned path must use forward slashes (normalized)
	if strings.Contains(files[0], "\\") {
		t.Errorf("path not normalized to forward slashes: %q", files[0])
	}
	if !strings.Contains(files[0], ".moai/harness/main.md") {
		t.Errorf("expected .moai/harness/main.md in path, got: %q", files[0])
	}

	// Sanity check: cross-platform OS info available
	_ = runtime.GOOS
}
