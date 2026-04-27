// SPEC-V3R3-PROJECT-HARNESS-001 / T-P4-01
// New Session Auto-Activation simulation (AC-PH-04 verification, fixes D9).

package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resolveImports walks @import directives inside the harness marker block of
// claudeMdPath and returns the concatenated content of all referenced files.
// Acts as a pure-Go simulation of Claude Code's session-start CLAUDE.md
// loading + @import follow. Replaces the spec's stale `moai test session-replay`
// pseudo-command (D9 fix).
func resolveImports(t *testing.T, projectRoot, claudeMdPath string) string {
	t.Helper()
	data, err := os.ReadFile(claudeMdPath)
	if err != nil {
		t.Fatalf("read claudeMd: %v", err)
	}
	content := string(data)
	startIdx := strings.Index(content, "<!-- moai:harness-start")
	endIdx := strings.Index(content, "<!-- moai:harness-end -->")
	if startIdx < 0 || endIdx < 0 || endIdx < startIdx {
		t.Fatalf("marker block missing or unpaired")
	}
	block := content[startIdx:endIdx]

	var resolved []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "See @") {
			continue
		}
		ref := strings.TrimPrefix(line, "See @")
		ref = strings.TrimSpace(ref)
		path := filepath.Join(projectRoot, ref)
		body, err := os.ReadFile(path)
		if err == nil {
			resolved = append(resolved, "[BEGIN "+ref+"]\n"+string(body)+"\n[END "+ref+"]")
		}
	}
	return strings.Join(resolved, "\n\n")
}

func TestSessionReplay_HarnessContextResolved(t *testing.T) {
	root := t.TempDir()
	// Build minimal fixture
	harnessDir := filepath.Join(root, ".moai", "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mainBody := "# Harness Main\n**Domain**: ios-mobile\nThis project uses ios-architect agent chain.\n"
	if err := os.WriteFile(filepath.Join(harnessDir, "main.md"), []byte(mainBody), 0o644); err != nil {
		t.Fatal(err)
	}

	claudeMd := `# Project

## Project-Specific Configuration (Harness-Generated)
<!-- moai:harness-start id="SPEC-PROJ-INIT-001" generated="2026-04-27T00:00:00Z" -->
**Domain**: ios-mobile

See @.moai/harness/main.md
<!-- moai:harness-end -->
`
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(claudeMd), 0o644); err != nil {
		t.Fatal(err)
	}

	resolved := resolveImports(t, root, filepath.Join(root, "CLAUDE.md"))
	if !strings.Contains(resolved, "ios-mobile") {
		t.Error("resolved context missing domain")
	}
	if !strings.Contains(resolved, "ios-architect") {
		t.Error("resolved context missing agent chain reference")
	}
	if !strings.Contains(resolved, "[BEGIN .moai/harness/main.md]") {
		t.Error("import marker missing")
	}
}

func TestSessionReplay_MissingImportFile_Skipped(t *testing.T) {
	root := t.TempDir()
	claudeMd := `## Project-Specific Configuration (Harness-Generated)
<!-- moai:harness-start id="S" -->
See @.moai/harness/missing.md
<!-- moai:harness-end -->
`
	_ = os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(claudeMd), 0o644)
	resolved := resolveImports(t, root, filepath.Join(root, "CLAUDE.md"))
	// Missing file → silently skipped (graceful), result is empty string.
	if strings.Contains(resolved, "[BEGIN") {
		t.Errorf("missing file should not produce import marker")
	}
}
