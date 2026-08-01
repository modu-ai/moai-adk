// moai update Safety regression — User Area Preservation.
//
// SPEC-UPDATE-DATA-SURVIVAL-001 M5 (REQ-UDS-015..018): this guard drives the
// real production entry point deploy.CleanMoaiManagedPaths rather than a fake
// defined in this file. The predecessor test called a local simulateMoaiUpdate
// helper whose body only touched a managed path and carried the comment "we
// intentionally do NOT touch user areas" — it therefore guaranteed its own
// passing condition and would have stayed green while production code deleted
// every user-owned path.

package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
)

// snapshotDir computes a stable hash of all files under root, mapping
// relative path → SHA-256(content). Used to detect any modification.
//
// A missing root yields an empty map rather than a fatal error, so a user-area
// directory deleted by the code under test surfaces as a snapshot mismatch with
// a readable diff instead of an unrelated "no such file" failure.
func snapshotDir(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return out
	}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return out
}

// writeFixture creates parent directories and writes content at
// filepath.Join(root, rel).
func writeFixture(t *testing.T, root, rel, content string) string {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return full
}

// TestMoaiUpdate_PreservesUserArea drives deploy.CleanMoaiManagedPaths — the
// destructive stale-path removal step that runs before template deployment —
// against a fixture holding both MoAI-managed and user-owned paths, and asserts
// in both directions: user-owned areas are byte-identical afterwards, and the
// managed targets were actually removed.
//
// The managed fixture is derived from the cleanTarget list inside
// CleanMoaiManagedPaths, not from what a fake happened to create. The
// managed-removal assertion is the anti-no-op half: without it, replacing the
// production function with an empty body would still pass, making the guard
// vacuous in a new way.
func TestMoaiUpdate_PreservesUserArea(t *testing.T) {
	root := t.TempDir()

	// User-owned paths. These MUST survive (REQ-UDS-016).
	userFiles := []struct {
		path    string
		content string
	}{
		{".moai/harness/run-extension.md", "# user-customized chain\n## Chain Rules\nfoo: bar\n"},
		{".moai/harness/main.md", "# main user content\n"},
		{".claude/agents/harness/ios-architect.md", "# ios architect agent\n"},
		{".claude/skills/harness-ios-patterns/SKILL.md", "# user skill\n"},
	}
	for _, uf := range userFiles {
		writeFixture(t, root, uf.path, uf.content)
	}

	// MoAI-managed paths, drawn from the cleanTarget list in
	// CleanMoaiManagedPaths. These are expected to be removed.
	managedFiles := []string{
		".claude/settings.json",
		".claude/agents/moai/manager-develop.md",
		".claude/skills/moai-workflow-tdd/SKILL.md",
		".claude/rules/moai/core/moai-constitution.md",
		".claude/hooks/moai/handle-session-start.sh",
		".moai/config/config.yaml",
	}
	managedFull := make([]string, len(managedFiles))
	for i, mf := range managedFiles {
		managedFull[i] = writeFixture(t, root, mf, "# moai-managed\n")
	}

	// Snapshot user-area directories before.
	userPaths := []string{
		filepath.Join(root, ".moai", "harness"),
		filepath.Join(root, ".claude", "agents", "harness"),
		filepath.Join(root, ".claude", "skills", "harness-ios-patterns"),
	}
	preSnapshots := make([]map[string]string, len(userPaths))
	for i, p := range userPaths {
		preSnapshots[i] = snapshotDir(t, p)
	}

	// Drive the real production entry point (REQ-UDS-015).
	if err := deploy.CleanMoaiManagedPaths(root, io.Discard); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	// Direction 1: user areas byte-identical before and after (REQ-UDS-016).
	for i, p := range userPaths {
		post := snapshotDir(t, p)
		if !mapsEqual(preSnapshots[i], post) {
			t.Errorf("user area changed: %s\npre:  %v\npost: %v", p, preSnapshots[i], post)
		}
	}

	// Direction 2: managed targets were actually removed. Without this the
	// guard would pass against a no-op implementation.
	removed := 0
	for i, full := range managedFull {
		if _, err := os.Stat(full); os.IsNotExist(err) {
			removed++
			continue
		}
		t.Errorf("managed path still present after clean: %s", managedFiles[i])
	}
	if removed == 0 {
		t.Error("no managed path was removed — the entry point did nothing")
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	keysA := make([]string, 0, len(a))
	for k := range a {
		keysA = append(keysA, k)
	}
	sort.Strings(keysA)
	for _, k := range keysA {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}
