package codexwiring

import (
	"encoding/json"
	"io/fs"
	"path"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template"
)

// renderedClaudeStopScripts renders the distributed settings template from
// fsys with the hook opt-in set as given and returns the Claude Stop handler
// scripts in registration order. Order is kept (template.ParseHookEntries
// sorts), because the Codex chain runs the members in Claude's order.
func renderedClaudeStopScripts(t *testing.T, fsys fs.FS, optIn bool) []string {
	t.Helper()
	ctx := template.NewTemplateContext(template.WithPlatform(runtime.GOOS), template.WithHookOptIn(optIn))
	rendered, err := template.NewRenderer(fsys).Render(template.SettingsTemplateName, ctx)
	if err != nil {
		t.Fatalf("render settings template (opt-in %v): %v", optIn, err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string   `json:"command"`
				Args    []string `json:"args"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(rendered, &doc); err != nil {
		t.Fatalf("parse rendered settings (opt-in %v): %v", optIn, err)
	}
	var scripts []string
	for _, group := range doc.Hooks["Stop"] {
		for _, h := range group.Hooks {
			name := h.Command
			for i := len(h.Args) - 1; i >= 0; i-- {
				if strings.HasSuffix(h.Args[i], ".sh") {
					name = path.Base(h.Args[i])
					break
				}
			}
			scripts = append(scripts, name)
		}
	}
	return scripts
}

// TestStopChainInventoryMatchesClaudeTemplate is AC-HPR-001 (REQ-HPR-001).
// The distributed settings template is rendered twice (hook opt-in off and
// on). Every Stop handler of either render must appear exactly once in the
// Codex Stop-chain inventory with a class and a Codex placement; the opt-in-only
// member must be flagged conditional; and no inventory row may be absent from
// both renders.
func TestStopChainInventoryMatchesClaudeTemplate(t *testing.T) {
	t.Parallel()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	plain := renderedClaudeStopScripts(t, fsys, false)
	optIn := renderedClaudeStopScripts(t, fsys, true)
	if len(plain) == 0 || len(optIn) == 0 {
		t.Fatalf("rendered Stop arrays are empty (plain %v, opt-in %v); the inventory check would be vacuous", plain, optIn)
	}
	if problems := CheckStopInventory(plain, optIn); len(problems) > 0 {
		t.Fatalf("Stop-chain inventory does not match the distributed template:\n  %s", strings.Join(problems, "\n  "))
	}

	// The two mutations the AC names, applied to an in-memory copy of the
	// template so a correct check is shown to turn red on each.
	raw, err := fs.ReadFile(fsys, template.SettingsTemplateName)
	if err != nil {
		t.Fatalf("read settings template: %v", err)
	}
	t.Run("an extra Stop handler without an inventory row is named", func(t *testing.T) {
		const extra = "handle-unlisted-stop-member.sh"
		anchor := `"${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-stop.sh"],`
		if !strings.Contains(string(raw), anchor) {
			t.Fatalf("mutation anchor %q not found in the template", anchor)
		}
		mutated := strings.Replace(string(raw), anchor, anchor+`
            "timeout": 5,
            "type": "command"
          },
          {
            "command": "bash",
            "args": ["-c", "exec bash \"$0\"", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/`+extra+`"],`, 1)
		mfs := fstest.MapFS{template.SettingsTemplateName: {Data: []byte(mutated)}}
		problems := CheckStopInventory(renderedClaudeStopScripts(t, mfs, false), renderedClaudeStopScripts(t, mfs, true))
		if !strings.Contains(strings.Join(problems, "\n"), extra) {
			t.Fatalf("an unlisted Stop handler was not reported; problems = %v", problems)
		}
	})
	t.Run("an opt-in render against an inventory lacking harness-observe-stop fails", func(t *testing.T) {
		var trimmed []StopMember
		for _, m := range StopChainMembers {
			if m.ClaudeScript != "handle-harness-observe-stop.sh" {
				trimmed = append(trimmed, m)
			}
		}
		problems := checkStopInventory(trimmed, plain, optIn)
		if !strings.Contains(strings.Join(problems, "\n"), "handle-harness-observe-stop.sh") {
			t.Fatalf("the opt-in-only member was not reported missing; problems = %v", problems)
		}
	})
}
