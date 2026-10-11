package template

// The shipping guard for automatic-repair artifacts (SPEC participation
// REQ-023, AC-024): the repo deliberately offers NO repair automation — the
// publication contract is the only interface repair tooling gets — so no
// auto-repair artifact may ship in the deployed template tree or the plugin
// mirror. The guard walks both real roots (paths and contents) and, in the
// canary test, proves it can fail by seeding violations in a synthetic
// overlay.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// autoRepairPathTokens are the path fragments an auto-repair artifact
// carries in its file name or directory.
var autoRepairPathTokens = []string{"auto-repair", "autorepair", "auto_repair"}

// autoRepairContentToken is the marker content an auto-repair artifact
// would carry into a shipped file.
const autoRepairContentToken = "moai-bugreport"

// autoRepairViolations walks the given roots and reports every shipped file
// whose path or content violates the no-auto-repair contract. A directory
// that does not exist is not a violation (a root may be optional per
// installation shape); the caller decides whether the roots it names are
// mandatory.
func autoRepairViolations(t *testing.T, roots ...string) []string {
	t.Helper()
	var violations []string
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				t.Fatalf("walk %s: %v", path, err)
			}
			if d.IsDir() {
				return nil
			}
			// Judge the SHIPPED path — relative to the walked root — so a
			// temporary walk prefix (a test's own temp dir name) cannot
			// trip the token check.
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				t.Fatalf("rel %s: %v", path, rerr)
			}
			lower := strings.ToLower(rel)
			for _, token := range autoRepairPathTokens {
				if strings.Contains(lower, token) {
					violations = append(violations, "path carries "+token+": "+filepath.Join(root, rel))
					return nil
				}
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			if strings.Contains(string(raw), autoRepairContentToken) {
				violations = append(violations, "content carries "+autoRepairContentToken+": "+filepath.Join(root, rel))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	return violations
}

// TestNoAutoRepairArtifactsShipped: the real deployed trees — the embedded
// template tree and the plugin mirror — carry no auto-repair artifact, by
// path or by content.
func TestNoAutoRepairArtifactsShipped(t *testing.T) {
	violations := autoRepairViolations(t, "templates", filepath.Join("..", "..", "plugins", "moai"))
	if len(violations) != 0 {
		t.Fatalf("auto-repair artifacts shipped: %d violation(s):\n%s", len(violations), strings.Join(violations, "\n"))
	}
}

// TestNoAutoRepairGuardCatchesCanary: the guard is not vacuous — a seeded
// canary file (auto-repair-canary.md, plus a content violation) in a
// temporary overlay makes it fail. The red is observed here, before the
// guard is adopted.
func TestNoAutoRepairGuardCatchesCanary(t *testing.T) {
	overlay := t.TempDir()
	canary := filepath.Join(overlay, "auto-repair-canary.md")
	if err := os.WriteFile(canary, []byte("innocent prose, no marker\n"), 0o644); err != nil {
		t.Fatalf("seed canary: %v", err)
	}
	infected := filepath.Join(overlay, "clean-named.md")
	if err := os.WriteFile(infected, []byte("carries "+autoRepairContentToken+" in its content\n"), 0o644); err != nil {
		t.Fatalf("seed infected file: %v", err)
	}

	violations := autoRepairViolations(t, overlay)
	byToken := map[string]bool{}
	for _, v := range violations {
		if strings.Contains(v, "auto-repair") {
			byToken["path"] = true
		}
		if strings.Contains(v, autoRepairContentToken) {
			byToken["content"] = true
		}
	}
	if !byToken["path"] || !byToken["content"] {
		t.Fatalf("the guard did not catch both canary shapes (violations: %v)", violations)
	}
}
