// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M4 swarm registry writer tests.
//
// Covers REQ-WTL-008 (registry schema + 0o600 file permission) per AC-WTL-008.
// Tests use t.TempDir() so every fixture lives under /tmp and is auto-cleaned;
// no shared state between tests, no real .moai/state/swarm/ touched in the
// project tree.
//
// The Windows file-mode model differs from POSIX (no concept of 0o600 group
// bits — Go's os package on Windows reports 0o666 for any writable file).
// Tests that assert exact permission bits use a build tag-free helper that
// checks permissions only when GOOS != "windows".

package worktree

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestSwarmRegistry_P1_Schema_PaneIDPopulated verifies the canonical 7-field
// JSON schema for a successful P1 (tmux + CG mode) launch. Pane ID must be
// non-empty for P1/P2 paths (REQ-WTL-008).
func TestSwarmRegistry_P1_Schema_PaneIDPopulated(t *testing.T) {
	tmpRoot := t.TempDir()
	want := SwarmEntry{
		SpecID:       "SPEC-WTL-DEMO-001",
		WorktreePath: "/tmp/wt-demo",
		Branch:       "feature/SPEC-WTL-DEMO-001",
		PaneID:       "%5",
		Mode:         "tmux-glm",
		CreatedAt:    time.Date(2026, 5, 23, 14, 0, 0, 0, time.UTC),
		CreatedByPID: 12345,
	}
	if err := WriteSwarmEntry(tmpRoot, want); err != nil {
		t.Fatalf("WriteSwarmEntry returned unexpected error: %v", err)
	}

	path := filepath.Join(tmpRoot, ".moai", "state", "swarm", "SPEC-WTL-DEMO-001.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry file: %v", err)
	}

	var got SwarmEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal registry JSON: %v", err)
	}
	if got.SpecID != want.SpecID {
		t.Errorf("SpecID = %q, want %q", got.SpecID, want.SpecID)
	}
	if got.WorktreePath != want.WorktreePath {
		t.Errorf("WorktreePath = %q, want %q", got.WorktreePath, want.WorktreePath)
	}
	if got.Branch != want.Branch {
		t.Errorf("Branch = %q, want %q", got.Branch, want.Branch)
	}
	if got.PaneID != want.PaneID {
		t.Errorf("PaneID = %q, want %q", got.PaneID, want.PaneID)
	}
	if got.Mode != "tmux-glm" {
		t.Errorf("Mode = %q, want tmux-glm", got.Mode)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, want.CreatedAt)
	}
	if got.CreatedByPID != want.CreatedByPID {
		t.Errorf("CreatedByPID = %d, want %d", got.CreatedByPID, want.CreatedByPID)
	}
}

// TestSwarmRegistry_P2_Schema_PaneIDPopulated verifies the P2 (tmux + CC mode)
// schema. Mode must be "tmux-cc" and pane_id non-empty.
func TestSwarmRegistry_P2_Schema_PaneIDPopulated(t *testing.T) {
	tmpRoot := t.TempDir()
	entry := SwarmEntry{
		SpecID:       "SPEC-WTL-DEMO-002",
		WorktreePath: "/tmp/wt-demo2",
		Branch:       "feature/SPEC-WTL-DEMO-002",
		PaneID:       "%7",
		Mode:         "tmux-cc",
		CreatedAt:    time.Now().UTC(),
		CreatedByPID: 67890,
	}
	if err := WriteSwarmEntry(tmpRoot, entry); err != nil {
		t.Fatalf("WriteSwarmEntry returned unexpected error: %v", err)
	}

	path := filepath.Join(tmpRoot, ".moai", "state", "swarm", "SPEC-WTL-DEMO-002.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry file: %v", err)
	}

	var got SwarmEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal registry JSON: %v", err)
	}
	if got.Mode != "tmux-cc" {
		t.Errorf("Mode = %q, want tmux-cc", got.Mode)
	}
	if got.PaneID == "" {
		t.Errorf("PaneID must be non-empty for P2, got empty")
	}
}

// TestSwarmRegistry_P3_Schema_PaneIDEmpty verifies the P3 (no-tmux,
// in-process) schema. Pane ID MUST be empty because P3 uses syscall.Exec
// instead of a tmux pane. Mode is "in-progress-cc" or "in-progress-glm".
func TestSwarmRegistry_P3_Schema_PaneIDEmpty(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		llm      string
		wantMode string
	}{
		{name: "P3 CC", mode: "in-progress-cc", llm: "cc", wantMode: "in-progress-cc"},
		{name: "P3 GLM", mode: "in-progress-glm", llm: "glm", wantMode: "in-progress-glm"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpRoot := t.TempDir()
			entry := SwarmEntry{
				SpecID:       "SPEC-WTL-P3-001",
				WorktreePath: "/tmp/wt-p3",
				Branch:       "feature/SPEC-WTL-P3-001",
				PaneID:       "", // P3 has no tmux pane
				Mode:         tt.mode,
				CreatedAt:    time.Now().UTC(),
				CreatedByPID: 99999,
			}
			if err := WriteSwarmEntry(tmpRoot, entry); err != nil {
				t.Fatalf("WriteSwarmEntry returned unexpected error: %v", err)
			}

			path := filepath.Join(tmpRoot, ".moai", "state", "swarm", "SPEC-WTL-P3-001.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read registry file: %v", err)
			}

			var got SwarmEntry
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal registry JSON: %v", err)
			}
			if got.PaneID != "" {
				t.Errorf("PaneID = %q, want empty string for P3", got.PaneID)
			}
			if got.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", got.Mode, tt.wantMode)
			}
		})
	}
}

// TestSwarmRegistry_ParentDirAutoCreated verifies that WriteSwarmEntry
// creates the .moai/state/swarm/ parent directory when it does not already
// exist. The parent dir uses 0o755 permissions (per-user state directory).
func TestSwarmRegistry_ParentDirAutoCreated(t *testing.T) {
	tmpRoot := t.TempDir()
	// Verify parent dir does NOT exist initially.
	swarmDir := filepath.Join(tmpRoot, ".moai", "state", "swarm")
	if _, err := os.Stat(swarmDir); !os.IsNotExist(err) {
		t.Fatalf("test fixture invalid: swarm dir should not exist initially, stat err=%v", err)
	}

	entry := SwarmEntry{
		SpecID:       "SPEC-WTL-AUTODIR-001",
		WorktreePath: "/tmp/wt-autodir",
		Branch:       "feature/SPEC-WTL-AUTODIR-001",
		PaneID:       "%1",
		Mode:         "tmux-cc",
		CreatedAt:    time.Now().UTC(),
		CreatedByPID: os.Getpid(),
	}
	if err := WriteSwarmEntry(tmpRoot, entry); err != nil {
		t.Fatalf("WriteSwarmEntry must auto-create parent dir; got error: %v", err)
	}

	// Parent dir must now exist.
	info, err := os.Stat(swarmDir)
	if err != nil {
		t.Fatalf("swarm dir should have been auto-created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("swarm path exists but is not a directory")
	}
}

// TestSwarmRegistry_FilePerm_0o600 verifies the registry file is written with
// 0o600 permission bits (per-user readable/writable, no group/other access).
// REQ-WTL-008 perm constraint. On Windows, file permissions follow a
// different model so this test is POSIX-only.
func TestSwarmRegistry_FilePerm_0o600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file mode 0o600 semantics differ on Windows (POSIX-only test)")
	}
	tmpRoot := t.TempDir()
	entry := SwarmEntry{
		SpecID:       "SPEC-WTL-PERM-001",
		WorktreePath: "/tmp/wt-perm",
		Branch:       "feature/SPEC-WTL-PERM-001",
		PaneID:       "%2",
		Mode:         "tmux-cc",
		CreatedAt:    time.Now().UTC(),
		CreatedByPID: os.Getpid(),
	}
	if err := WriteSwarmEntry(tmpRoot, entry); err != nil {
		t.Fatalf("WriteSwarmEntry returned unexpected error: %v", err)
	}

	path := filepath.Join(tmpRoot, ".moai", "state", "swarm", "SPEC-WTL-PERM-001.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat registry file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("file mode = %o, want 0o600", got)
	}
}

// TestSwarmRegistry_JSONFields verifies all 7 canonical fields appear in the
// raw JSON output with their canonical names (spec_id, worktree_path, branch,
// pane_id, mode, created_at, created_by_pid). This is a regression guard
// against accidental field renames or struct-tag changes.
func TestSwarmRegistry_JSONFields(t *testing.T) {
	tmpRoot := t.TempDir()
	entry := SwarmEntry{
		SpecID:       "SPEC-WTL-FIELDS-001",
		WorktreePath: "/tmp/wt-fields",
		Branch:       "feature/SPEC-WTL-FIELDS-001",
		PaneID:       "%3",
		Mode:         "tmux-cc",
		CreatedAt:    time.Date(2026, 5, 23, 16, 30, 0, 0, time.UTC),
		CreatedByPID: 4242,
	}
	if err := WriteSwarmEntry(tmpRoot, entry); err != nil {
		t.Fatalf("WriteSwarmEntry returned unexpected error: %v", err)
	}

	path := filepath.Join(tmpRoot, ".moai", "state", "swarm", "SPEC-WTL-FIELDS-001.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry file: %v", err)
	}

	// Unmarshal into a map to verify exact JSON keys.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal registry JSON: %v", err)
	}
	wantKeys := []string{
		"spec_id",
		"worktree_path",
		"branch",
		"pane_id",
		"mode",
		"created_at",
		"created_by_pid",
	}
	for _, k := range wantKeys {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing JSON key %q in registry file", k)
		}
	}
	// No extra fields should leak.
	if len(raw) != len(wantKeys) {
		t.Errorf("registry JSON has %d keys, want %d (extra keys: %v)", len(raw), len(wantKeys), raw)
	}
}

// TestSwarmRegistry_MkdirFailure verifies that WriteSwarmEntry surfaces an
// error when the parent directory cannot be created. The fixture creates a
// regular FILE at the position where ".moai" would be a directory, so
// os.MkdirAll fails with "not a directory". This exercises the error return
// branch and pushes WriteSwarmEntry coverage past the 85% threshold.
//
// POSIX-only: Windows file/directory semantics around MkdirAll differ.
func TestSwarmRegistry_MkdirFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("MkdirAll-vs-file semantics differ on Windows")
	}
	tmpRoot := t.TempDir()
	// Plant a regular file at ".moai" so MkdirAll(.moai/state/swarm) fails.
	blocker := filepath.Join(tmpRoot, ".moai")
	if err := os.WriteFile(blocker, []byte("not-a-dir"), 0o644); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}

	entry := SwarmEntry{
		SpecID:       "SPEC-WTL-MKDIR-FAIL",
		WorktreePath: "/tmp/wt-mkdir",
		Branch:       "feature/SPEC-WTL-MKDIR-FAIL",
		PaneID:       "%99",
		Mode:         "tmux-cc",
		CreatedAt:    time.Now().UTC(),
		CreatedByPID: os.Getpid(),
	}
	err := WriteSwarmEntry(tmpRoot, entry)
	if err == nil {
		t.Fatalf("WriteSwarmEntry should fail when .moai is a regular file; got nil")
	}
	if !strings.Contains(err.Error(), "mkdir") {
		t.Errorf("error must wrap mkdir; got: %v", err)
	}

	// And verify nothing was written. NOTE: when ".moai" itself is a regular
	// file, os.Stat on a deeper child path returns ENOTDIR ("not a
	// directory") rather than ENOENT ("does not exist"). Either result means
	// the registry file is unreachable — we accept both.
	registryPath := filepath.Join(tmpRoot, ".moai", "state", "swarm", entry.SpecID+".json")
	_, statErr := os.Stat(registryPath)
	if statErr == nil {
		t.Errorf("registry file MUST NOT exist on mkdir failure (Stat returned nil err)")
	}
}

// TestSwarmRegistry_WriteFailure verifies that WriteSwarmEntry surfaces an
// error when os.WriteFile fails on the target path. The fixture pre-creates
// a DIRECTORY at the expected file path (.moai/state/swarm/<SPEC>.json),
// which causes os.WriteFile to fail with "is a directory". This exercises
// the write-error branch and pushes WriteSwarmEntry coverage past 85%.
//
// POSIX-only: Windows directory-vs-file semantics on WriteFile differ.
func TestSwarmRegistry_WriteFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("WriteFile-vs-directory semantics differ on Windows")
	}
	tmpRoot := t.TempDir()
	specID := "SPEC-WTL-WRITE-FAIL"
	// Pre-create the swarm dir AND a directory at the file path.
	collisionPath := filepath.Join(tmpRoot, ".moai", "state", "swarm", specID+".json")
	if err := os.MkdirAll(collisionPath, 0o755); err != nil {
		t.Fatalf("mkdir collision path: %v", err)
	}

	entry := SwarmEntry{
		SpecID:       specID,
		WorktreePath: "/tmp/wt-writefail",
		Branch:       "feature/" + specID,
		PaneID:       "%99",
		Mode:         "tmux-cc",
		CreatedAt:    time.Now().UTC(),
		CreatedByPID: os.Getpid(),
	}
	err := WriteSwarmEntry(tmpRoot, entry)
	if err == nil {
		t.Fatalf("WriteSwarmEntry should fail when target path is a directory; got nil")
	}
	if !strings.Contains(err.Error(), "write swarm entry") {
		t.Errorf("error must wrap write step; got: %v", err)
	}
}

// TestPatternToMode verifies the mapping from Pattern + LLM into the
// canonical SwarmEntry.Mode string. P4 returns empty because P4 paths never
// write a registry entry (no spawn occurred).
func TestPatternToMode(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		llm     string
		want    string
	}{
		{"P1 tmux GLM", PatternP1TmuxGLM, "glm", "tmux-glm"},
		{"P2 tmux CC", PatternP2TmuxCC, "cc", "tmux-cc"},
		{"P3 in-process CC", PatternP3InProgress, "cc", "in-progress-cc"},
		{"P3 in-process GLM", PatternP3InProgress, "glm", "in-progress-glm"},
		{"P4 handoff (no registry)", PatternP4Handoff, "cc", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := patternToMode(tt.pattern, tt.llm); got != tt.want {
				t.Errorf("patternToMode(%v, %q) = %q, want %q", tt.pattern, tt.llm, got, tt.want)
			}
		})
	}
}
