// SPEC-INTERNAL-SECURITY-001 M2 (REQ-SEC-004 / AC-SEC-004b) template embed test.
//
// Verifies the embedded template filesystem serves a .gitignore that contains
// the `.moai/backups/` slash-form entry. The template source MUST be edited
// first and re-embedded via `make build` (NFR-SEC-001 Template-First).
package template

import (
	"io/fs"
	"strings"
	"testing"
)

// AC-SEC-004b: the embedded FS serves .gitignore with .moai/backups/.
func TestEmbeddedGitignoreHasMoaiBackupsSlash(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fs.ReadFile(fsys, ".gitignore")
	if err != nil {
		t.Fatalf("read embedded .gitignore: %v", err)
	}
	content := string(data)
	// Must contain the slash-form entry as a discrete line (the drift
	// REQ-SEC-004 closes). A substring match is insufficient because the
	// template previously had only the hyphen form `.moai-backups/`.
	lines := strings.Split(content, "\n")
	found := false
	for _, ln := range lines {
		if strings.TrimSpace(ln) == ".moai/backups/" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("embedded .gitignore missing `.moai/backups/` slash-form entry (REQ-SEC-004)")
	}
}
