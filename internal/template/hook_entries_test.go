package template

import (
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

// fixtureSettings is a minimal settings document carrying two hook entries in
// one event plus a statusLine block, which is `"type": "command"` but is NOT a
// hook (spec.md §A.2 correction 4).
const fixtureSettings = `{
  "hooks": {
    "SubagentStop": [
      {
        "hooks": [
          {"command": "bash", "args": ["-c", "fallback", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-subagent-stop.sh"], "timeout": 10, "type": "command", "async": true},
          {"command": "bash", "args": ["-c", "fallback", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/chain-event.sh"], "timeout": 5, "type": "command"}
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {"command": "bash", "args": ["-c", "fallback", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/status-transition-ownership.sh"], "if": "Write(**/.moai/specs/**)", "timeout": 5, "type": "command", "async": true}
        ]
      }
    ]
  },
  "statusLine": {"command": "bash", "args": ["-c", "echo hi"], "type": "command"}
}`

func TestParseHookEntriesKeysOnFullTuple(t *testing.T) {
	entries, err := ParseHookEntries([]byte(fixtureSettings))
	if err != nil {
		t.Fatalf("ParseHookEntries: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("entry count = %d, want 3 (statusLine must not be counted as a hook)", len(entries))
	}
	var chain, sto *HookEntry
	for i := range entries {
		switch entries[i].Script {
		case "chain-event.sh":
			chain = &entries[i]
		case "status-transition-ownership.sh":
			sto = &entries[i]
		}
	}
	if chain == nil {
		t.Fatal("chain-event.sh entry not parsed")
	}
	if chain.Event != "SubagentStop" || chain.Timeout != 5 || chain.Async {
		t.Errorf("chain-event entry = %+v, want SubagentStop timeout=5 async=false", *chain)
	}
	if sto == nil {
		t.Fatal("status-transition-ownership.sh entry not parsed")
	}
	if sto.Matcher != "Write|Edit" || sto.If != "Write(**/.moai/specs/**)" || !sto.Async {
		t.Errorf("status-transition entry = %+v, want matcher/if/async populated", *sto)
	}
}

func TestParseHookEntriesRejectsInvalidJSON(t *testing.T) {
	if _, err := ParseHookEntries([]byte(`{"hooks": `)); err == nil {
		t.Fatal("want error on truncated JSON, got nil")
	}
}

func TestDiffHookEntriesReportsBothDirections(t *testing.T) {
	tmpl := []HookEntry{
		{Event: "SubagentStop", Script: "chain-event.sh", Timeout: 5},
		{Event: "PostToolUse", Script: "shared.sh", Timeout: 5},
	}
	project := []HookEntry{
		{Event: "PostToolUse", Script: "shared.sh", Timeout: 5},
		{Event: "PostToolUse", Script: "extra-local.sh", Timeout: 5},
	}
	tmplOnly, projOnly := DiffHookEntries(tmpl, project)
	if len(tmplOnly) != 1 || tmplOnly[0].Script != "chain-event.sh" {
		t.Errorf("templateOnly = %v, want [chain-event.sh]", tmplOnly)
	}
	if len(projOnly) != 1 || projOnly[0].Script != "extra-local.sh" {
		t.Errorf("projectOnly = %v, want [extra-local.sh]", projOnly)
	}
}

// A same-name entry whose tuple differs (the live unscoped synchronous
// status-transition entry vs the template's three if-scoped async ones) MUST
// diff in both directions — a script-name-only comparison would report none.
func TestDiffHookEntriesTupleNotNameOnly(t *testing.T) {
	tmpl := []HookEntry{{Event: "PostToolUse", Script: "sto.sh", If: "Write(**)", Timeout: 5, Async: true}}
	project := []HookEntry{{Event: "PostToolUse", Script: "sto.sh", Timeout: 5}}
	tmplOnly, projOnly := DiffHookEntries(tmpl, project)
	if len(tmplOnly) != 1 || len(projOnly) != 1 {
		t.Fatalf("templateOnly=%v projectOnly=%v, want one each (tuple-keyed, not name-keyed)", tmplOnly, projOnly)
	}
}

func TestRenderHookEntriesFromEmbeddedTemplate(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	ctx := NewTemplateContext(WithPlatform(runtime.GOOS), WithHookOptIn(true))
	entries, err := RenderHookEntries(fsys, ctx)
	if err != nil {
		t.Fatalf("RenderHookEntries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no hook entries rendered from the embedded template")
	}
	found := false
	for _, e := range entries {
		if e.Script == "chain-event.sh" && e.Event == "SubagentStop" {
			found = true
		}
	}
	if !found {
		t.Error("rendered template carries no SubagentStop/chain-event.sh entry")
	}
}

// The template source is an injectable parameter: a fixture FS must be
// rendered instead of the embedded one (plan.md §F M2 [HARD]).
func TestRenderHookEntriesUsesInjectedSource(t *testing.T) {
	fixture := fstest.MapFS{
		SettingsTemplateName: &fstest.MapFile{Data: []byte(`{
  "hooks": {
    "SubagentStop": [
      {"hooks": [{"command": "bash", "args": ["-c", "fallback", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/zz-fixture-only.sh"], "timeout": 5, "type": "command"}]}
    ]
  }
}`)},
	}
	entries, err := RenderHookEntries(fixture, NewTemplateContext(WithPlatform(runtime.GOOS)))
	if err != nil {
		t.Fatalf("RenderHookEntries(fixture): %v", err)
	}
	if len(entries) != 1 || entries[0].Script != "zz-fixture-only.sh" {
		t.Fatalf("entries = %v, want the fixture's zz-fixture-only.sh entry", entries)
	}
}

func TestRenderHookEntriesNilSourceErrors(t *testing.T) {
	if _, err := RenderHookEntries(nil, NewTemplateContext()); err == nil {
		t.Fatal("want error on nil template source, got nil")
	}
}

func TestHookEntryScriptsDedupesWithCount(t *testing.T) {
	got := HookEntryScripts([]HookEntry{
		{Script: "b.sh"}, {Script: "a.sh", If: "x"}, {Script: "a.sh", If: "y"},
	})
	want := "a.sh x2, b.sh"
	if strings.Join(got, ", ") != want {
		t.Errorf("HookEntryScripts = %q, want %q", strings.Join(got, ", "), want)
	}
}
