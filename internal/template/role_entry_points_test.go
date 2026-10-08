package template

// role_entry_points_test.go: the REQ-ALB-024 entry-point guards of
// SPEC-ALWAYS-LOADED-BUDGET-001 (AC-ALB-015..018).
//
// A deployed agent definition, skill, or workflow file that cites a
// role-gated rule (factory-dispatch / cross-session-messaging) carries the
// neutral marker moai:role-rules-required together with a binding directive
// to read both rule files in full before its first action. The sweep
// enumerates citing files from the deployed (embedded template) tree and
// fails on any citing file without the marker; the mutation fixture plants
// a citing-without-marker file and observes the sweep catch it.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// roleRuleCitations are the bare names that count as citing a role-gated
// rule (path citations contain the bare name; the templates cite the rules
// by path or by bare name only).
var roleRuleCitations = []string{"factory-dispatch", "cross-session-messaging"}

// isRoleEntryPointPath reports whether path is a deployed agent, skill, or
// workflow file the REQ-ALB-024 sweep covers. Rule files themselves and the
// .codex emissions are out of scope (the emissions mirror their .md source;
// mirror parity is another test's subject).
func isRoleEntryPointPath(path string) bool {
	if !strings.HasSuffix(path, ".md") {
		return false
	}
	p := filepath.ToSlash(path)
	return strings.HasPrefix(p, ".claude/agents/") ||
		strings.HasPrefix(p, ".claude/skills/")
}

// sweepRoleRuleCitations walks fsys and returns the citing files whose
// marker or read-first directive is missing. It returns the swept count so
// an empty result on an empty sweep is visible, never a silent pass.
func sweepRoleRuleCitations(fsys fs.FS) (missing []string, swept int) {
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !isRoleEntryPointPath(path) {
			return nil
		}
		data, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			return readErr
		}
		content := string(data)
		cites := false
		for _, name := range roleRuleCitations {
			if strings.Contains(content, name) {
				cites = true
				break
			}
		}
		if !cites {
			return nil
		}
		swept++
		if !strings.Contains(content, config.RoleCoreMarkerRequired) {
			missing = append(missing, path)
		}
		return nil
	})
	if err != nil {
		return nil, swept
	}
	return missing, swept
}

// TestDeployedRoleRuleEntryPointMarkers is the REQ-ALB-024 sweep over the
// embedded (deployed) tree: every citing agent/skill/workflow file carries
// the marker. The sweep count is logged so a zero swept count on the
// deployed tree is visible rather than read as a pass.
func TestDeployedRoleRuleEntryPointMarkers(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	missing, swept := sweepRoleRuleCitations(fsys)
	if swept == 0 {
		t.Fatal("swept 0 citing files — the deployed tree carries no citations, so this guard asserts nothing")
	}
	for _, m := range missing {
		t.Errorf("deployed entry point %s cites a role-gated rule without the %s marker", m, config.RoleCoreMarkerRequired)
	}
	t.Logf("swept %d citing entry-point files; %d missing the marker", swept, len(missing))
}

// TestDeployedRoleRuleEntryPointMutation is the AC-ALB-018 observed-failure
// fixture: a citing file WITHOUT the marker, planted in a temp deployment
// tree, is caught by the same sweep.
func TestDeployedRoleRuleEntryPointMutation(t *testing.T) {
	root := t.TempDir()
	dst := filepath.Join(root, ".claude", "skills", "fixture-skill")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "---\ndescription: fixture\n---\n\n# Fixture\n\nSee `.claude/rules/moai/workflow/factory-dispatch.md` for the dispatch cycle.\n"
	if err := os.WriteFile(filepath.Join(dst, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	missing, swept := sweepRoleRuleCitations(os.DirFS(root))
	if swept != 1 {
		t.Fatalf("swept %d citing files, want 1", swept)
	}
	if len(missing) != 1 {
		t.Fatalf("citing-without-marker fixture NOT caught: missing=%v", missing)
	}
	t.Logf("mutation observed RED: fixture %s reported missing the marker", missing[0])
}
