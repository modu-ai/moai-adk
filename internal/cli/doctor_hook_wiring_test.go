package cli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/template"
)

// embeddedTemplateFS returns the production template source.
func embeddedTemplateFS(t *testing.T) fs.FS {
	t.Helper()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	return fsys
}

// renderedSettings renders the settings template exactly as a project with no
// system.yaml would receive it (hook opt-in disabled — the t.TempDir default).
// Every M2 fixture is built from this render, never by copying the project
// (acceptance.md §M2 [HARD] fixture construction note).
func renderedSettings(t *testing.T) map[string]any {
	t.Helper()
	ctx := template.NewTemplateContext(
		template.WithPlatform(runtime.GOOS),
		template.WithHookOptIn(false),
	)
	b, err := template.NewRenderer(embeddedTemplateFS(t)).Render(template.SettingsTemplateName, ctx)
	if err != nil {
		t.Fatalf("render settings template: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("rendered settings is not valid JSON: %v", err)
	}
	return doc
}

// writeFixtureProject writes doc as <dir>/.claude/settings.json.
func writeFixtureProject(t *testing.T, doc map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), b, 0o644); err != nil {
		t.Fatalf("write fixture settings: %v", err)
	}
	return dir
}

// eventHooks returns the hooks array of the first group under event.
func eventHooks(t *testing.T, doc map[string]any, event string) (group map[string]any, hooks []any) {
	t.Helper()
	hooksMap, ok := doc["hooks"].(map[string]any)
	if !ok {
		t.Fatal("rendered settings has no hooks map")
	}
	groups, ok := hooksMap[event].([]any)
	if !ok || len(groups) == 0 {
		t.Fatalf("rendered settings has no %s group", event)
	}
	group, ok = groups[0].(map[string]any)
	if !ok {
		t.Fatalf("%s group is not an object", event)
	}
	hooks, ok = group["hooks"].([]any)
	if !ok {
		t.Fatalf("%s group carries no hooks array", event)
	}
	return group, hooks
}

// AC-HWD-005 — a template-only entry (missing registration) is reported and
// the affected script is named.
func TestCheckHookWiringDriftReportsTemplateOnly(t *testing.T) {
	doc := renderedSettings(t)
	group, hooks := eventHooks(t, doc, "SubagentStop")
	kept := make([]any, 0, len(hooks))
	removed := 0
	for _, h := range hooks {
		if strings.Contains(toJSON(t, h), "chain-event.sh") {
			removed++
			continue
		}
		kept = append(kept, h)
	}
	if removed != 1 {
		t.Fatalf("fixture setup: removed %d chain-event entries, want 1", removed)
	}
	group["hooks"] = kept

	check := checkHookWiringDrift(writeFixtureProject(t, doc), embeddedTemplateFS(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn", check.Status)
	}
	if !strings.Contains(check.Message, "chain-event.sh") {
		t.Errorf("message does not name chain-event.sh: %q", check.Message)
	}
	if !strings.Contains(check.Message, "template-only") {
		t.Errorf("message does not identify the direction as template-only: %q", check.Message)
	}
}

// AC-HWD-006 — a project-only entry (extra registration) is reported.
func TestCheckHookWiringDriftReportsProjectOnly(t *testing.T) {
	doc := renderedSettings(t)
	group, hooks := eventHooks(t, doc, "PostToolUse")
	group["hooks"] = append(hooks, map[string]any{
		"command": "bash",
		"args": []any{
			"-c", "fallback",
			"${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/zz-project-only.sh",
		},
		"timeout": 5,
		"type":    "command",
	})

	check := checkHookWiringDrift(writeFixtureProject(t, doc), embeddedTemplateFS(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn", check.Status)
	}
	if !strings.Contains(check.Message, "zz-project-only.sh") {
		t.Errorf("message does not name the extra script: %q", check.Message)
	}
	if !strings.Contains(check.Message, "project-only") {
		t.Errorf("message does not identify the direction as project-only: %q", check.Message)
	}
}

// The drift-free fixture must report no drift — the negative run AC-HWD-017
// requires, so the check is not satisfied by naming every script always.
func TestCheckHookWiringDriftCleanFixtureReportsNoDrift(t *testing.T) {
	check := checkHookWiringDrift(writeFixtureProject(t, renderedSettings(t)), embeddedTemplateFS(t), false)
	if check.Status != uikit.CheckOK {
		t.Fatalf("status = %v (%s), want ok on a drift-free fixture", check.Status, check.Message)
	}
	if strings.Contains(check.Message, "zz-fixture-only.sh") || strings.Contains(check.Message, "drift") {
		t.Errorf("clean fixture message names drift: %q", check.Message)
	}
}

// AC-HWD-017 — the check actually renders the template it is handed. A
// hardcoded expected-entry list cannot name a script it has never seen.
func TestCheckHookWiringDriftRendersInjectedTemplate(t *testing.T) {
	fixtureTmpl := fstest.MapFS{
		template.SettingsTemplateName: &fstest.MapFile{Data: []byte(`{
  "hooks": {
    "SubagentStop": [
      {"hooks": [{"command": "bash", "args": ["-c", "fallback", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/zz-fixture-only.sh"], "timeout": 5, "type": "command"}]}
    ]
  }
}`)},
	}
	dir := writeFixtureProject(t, map[string]any{"hooks": map[string]any{}})

	check := checkHookWiringDrift(dir, fixtureTmpl, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn", check.Status)
	}
	if !strings.Contains(check.Message, "zz-fixture-only.sh") {
		t.Fatalf("message does not name the injected fixture script — the check is not rendering its template source: %q", check.Message)
	}
	if !strings.Contains(check.Message, "template-only") {
		t.Errorf("message does not classify the fixture script as template-only: %q", check.Message)
	}
}

// AC-HWD-008 — absent settings file: warn naming the cause, never fail.
func TestCheckHookWiringDriftMissingSettingsWarns(t *testing.T) {
	check := checkHookWiringDrift(t.TempDir(), embeddedTemplateFS(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn", check.Status)
	}
	if !strings.Contains(check.Message, ".claude/settings.json") || !strings.Contains(check.Message, "not found") {
		t.Errorf("message does not name the cause: %q", check.Message)
	}
}

// AC-HWD-008 — unparseable settings file: warn naming the cause.
func TestCheckHookWiringDriftCorruptSettingsWarns(t *testing.T) {
	dir := writeFixtureProject(t, renderedSettings(t))
	path := filepath.Join(dir, ".claude", "settings.json")
	if err := os.WriteFile(path, []byte(`{"hooks": {"SubagentStop": [`), 0o644); err != nil {
		t.Fatalf("truncate fixture: %v", err)
	}
	check := checkHookWiringDrift(dir, embeddedTemplateFS(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn", check.Status)
	}
	if !strings.Contains(check.Message, "parse") {
		t.Errorf("message does not name the parse failure as the cause: %q", check.Message)
	}
}

// AC-HWD-008 — an unusable template source is a warn, not a fail.
func TestCheckHookWiringDriftRenderFailureWarns(t *testing.T) {
	check := checkHookWiringDrift(writeFixtureProject(t, renderedSettings(t)), fstest.MapFS{}, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn", check.Status)
	}
	if !strings.Contains(strings.ToLower(check.Message), "template") {
		t.Errorf("message does not name the render failure as the cause: %q", check.Message)
	}
}

// AC-HWD-004 / REQ-HWD-002 — the message must not claim the chain ledger works.
// AC-HWD-007 — the check writes nothing under the project root.
func TestCheckHookWiringDriftWritesNothing(t *testing.T) {
	doc := renderedSettings(t)
	group, hooks := eventHooks(t, doc, "SubagentStop")
	group["hooks"] = hooks[:len(hooks)-1]
	dir := writeFixtureProject(t, doc)

	before := treeSnapshot(t, dir)
	first := checkHookWiringDrift(dir, embeddedTemplateFS(t), false)
	after := treeSnapshot(t, dir)
	if strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Errorf("project tree changed:\nbefore:\n%s\nafter:\n%s",
			strings.Join(before, "\n"), strings.Join(after, "\n"))
	}

	// Twice-run harness correction: a self-repairing check reports the same
	// drift only on its first run.
	second := checkHookWiringDrift(dir, embeddedTemplateFS(t), false)
	if first.Message != second.Message || first.Status != second.Status {
		t.Errorf("second run diverged — the check is not read-only:\nfirst:  %v %q\nsecond: %v %q",
			first.Status, first.Message, second.Status, second.Message)
	}
}

// The check is registered in the doctor check list.
func TestHookWiringCheckIsRegistered(t *testing.T) {
	found := false
	for _, c := range runDiagnosticChecks(false, "") {
		if c.Name == "Hook Wiring" {
			found = true
		}
	}
	if !found {
		t.Fatal("doctor check list carries no \"Hook Wiring\" check")
	}
}

func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}

// treeSnapshot lists every file under root as name|size|mtime.
func treeSnapshot(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		out = append(out, fmt.Sprintf("%s|%d|%s", rel, info.Size(), info.ModTime().UTC()))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(out)
	return out
}
