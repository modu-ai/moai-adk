// carried_test.go — SPEC-AGENT-CORE4-001 REQ-AC4-007: from M2 on, the project
// definitions of the four carried agents are byte-identical to the template
// definitions they are rendered with, in both the Claude and the Codex form.
package agentemit_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// carriedFour names the agents whose project copies are generated in piece 1
// (spec §B). It is written here from the requirement, not read back from the
// layout under test, so a layout that drops one of them fails.
var carriedFour = []string{"builder-harness", "e2e-tester", "manager-design", "super-advisor"}

func TestCarriedFourCopiesEqual(t *testing.T) {
	for _, name := range carriedFour {
		t.Run(name, func(t *testing.T) {
			pairs := []struct{ label, project, template string }{
				{"c1_equals_c2", ".claude/agents/moai/" + name + ".md", "internal/template/templates/.claude/agents/moai/" + name + ".md"},
				{"c3p_equals_c3", ".codex/agents/moai/" + name + ".toml", "internal/template/templates/.codex/agents/moai/" + name + ".toml"},
			}
			for _, p := range pairs {
				t.Run(p.label, func(t *testing.T) {
					project, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(p.project)))
					if err != nil {
						t.Fatalf("project copy unreadable: %v", err)
					}
					tmpl, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(p.template)))
					if err != nil {
						t.Fatalf("template copy unreadable: %v", err)
					}
					if len(project) == 0 || len(tmpl) == 0 {
						t.Fatalf("an empty copy would make the comparison vacuous: project %d bytes, template %d bytes", len(project), len(tmpl))
					}
					if string(project) != string(tmpl) {
						t.Errorf("%s differs from %s", p.project, p.template)
					}
				})
			}
		})
	}
}

// TestDefaultLayoutGeneratesCarriedFourProjectSets pins the layout side of
// REQ-AC4-007: the shipped layout renders four sets, and the two project sets
// cover exactly the carried four.
func TestDefaultLayoutGeneratesCarriedFourProjectSets(t *testing.T) {
	lay := agentemit.DefaultLayout()
	want := map[string]agentemit.Target{
		"claude_template": {Dir: "internal/template/templates/.claude/agents/moai"},
		"codex_template":  {Dir: "internal/template/templates/.codex/agents/moai", Codex: true},
		"claude_project":  {Dir: ".claude/agents/moai", Only: carriedFour},
		"codex_project":   {Dir: ".codex/agents/moai", Codex: true, Only: carriedFour},
	}
	if len(lay.Targets) != len(want) {
		t.Fatalf("layout has %d sets, want %d", len(lay.Targets), len(want))
	}
	for _, tg := range lay.Targets {
		w, ok := want[tg.Name]
		if !ok {
			t.Errorf("unexpected set %q", tg.Name)
			continue
		}
		w.Name = tg.Name
		if !reflect.DeepEqual(tg, w) {
			t.Errorf("set %s = %+v, want %+v", tg.Name, tg, w)
		}
	}
}
