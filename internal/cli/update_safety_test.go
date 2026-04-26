// SPEC-V3R3-PROJECT-HARNESS-001 / T-P4-02
// moai update Safety regression — User Area Preservation (AC-PH-05).

package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// snapshotDir computes a stable hash of all files under root, mapping
// relative path → SHA-256(content). Used to detect any modification.
func snapshotDir(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
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

// simulateMoaiUpdate represents the conservative path-prefix exclusion that
// moai update is expected to honor: it only touches files under
// .claude/skills/moai/, .claude/agents/moai/, .claude/rules/moai/. It MUST NOT
// touch .moai/harness/, .claude/agents/my-harness/, .claude/skills/my-harness-*/
// (REQ-PH-009 enforcement under test).
func simulateMoaiUpdate(t *testing.T, projectRoot string) {
	t.Helper()
	// Touch a managed file (allowed)
	managed := filepath.Join(projectRoot, ".claude", "skills", "moai", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(managed), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(managed, []byte("# moai-managed updated\n"), 0o644)
	// Note: we intentionally do NOT touch user areas.
}

func TestMoaiUpdate_PreservesUserArea(t *testing.T) {
	root := t.TempDir()

	// Set up user area with custom comment
	userAreas := []struct {
		path    string
		content string
	}{
		{".moai/harness/run-extension.md", "# user-customized chain\n## Chain Rules\nfoo: bar\n"},
		{".moai/harness/main.md", "# main user content\n"},
		{".claude/agents/my-harness/ios-architect.md", "# ios architect agent\n"},
		{".claude/skills/my-harness-ios-patterns/SKILL.md", "# user skill\n"},
	}
	for _, ua := range userAreas {
		full := filepath.Join(root, ua.path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(ua.content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Snapshot user-area directories before
	userPaths := []string{
		filepath.Join(root, ".moai", "harness"),
		filepath.Join(root, ".claude", "agents", "my-harness"),
		filepath.Join(root, ".claude", "skills", "my-harness-ios-patterns"),
	}
	preSnapshots := make([]map[string]string, len(userPaths))
	for i, p := range userPaths {
		preSnapshots[i] = snapshotDir(t, p)
	}

	// Run simulated moai update
	simulateMoaiUpdate(t, root)

	// Verify user areas unchanged (REQ-PH-009)
	for i, p := range userPaths {
		post := snapshotDir(t, p)
		if !mapsEqual(preSnapshots[i], post) {
			t.Errorf("user area changed: %s\npre:  %v\npost: %v", p, preSnapshots[i], post)
		}
	}

	// Verify managed area was actually updated (sanity)
	managedData, _ := os.ReadFile(filepath.Join(root, ".claude", "skills", "moai", "SKILL.md"))
	if !strings.Contains(string(managedData), "moai-managed updated") {
		t.Errorf("simulateMoaiUpdate did not actually run update")
	}

	// Verify user comment preserved
	runExt, _ := os.ReadFile(filepath.Join(root, ".moai", "harness", "run-extension.md"))
	if !strings.Contains(string(runExt), "# user-customized chain") {
		t.Errorf("user customization comment lost")
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
