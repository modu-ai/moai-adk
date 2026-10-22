package template

// Registration pins cover both settings surfaces. The completion-edge writer
// stays synchronous so session teardown cannot cancel its foreground append;
// the two observation handlers retain next-turn async delivery.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// The completion-edge entry remains covered alongside the async observations.
var hookSyncAsyncConverted = []struct {
	event  string
	script string
	async  bool
}{
	{event: "PostToolUseFailure", script: "handle-post-tool-failure.sh", async: true},
	{event: "ConfigChange", script: "handle-config-change.sh", async: true},
	{event: "SubagentStop", script: "chain-event.sh", async: false},
}

// TestHookSyncAsyncRegistrationFlags asserts delivery and durability policies
// in both live settings and the rendered template.
func TestHookSyncAsyncRegistrationFlags(t *testing.T) {
	project := hookSyncAsyncProjectEntries(t)
	rendered := hookSyncAsyncRenderedEntries(t)

	for surface, entries := range map[string][]HookEntry{"live settings.json": project, "rendered settings template": rendered} {
		count := 0
		for _, entry := range entries {
			if entry.Async {
				count++
			}
		}
		if count != 13 {
			t.Errorf("%s: async inventory = %d, want 13", surface, count)
		}
	}

	for _, tt := range hookSyncAsyncConverted {
		assertHookSyncAsyncEntry(t, project, "live settings.json", tt.event, tt.script, tt.async)
		assertHookSyncAsyncEntry(t, rendered, "rendered settings template", tt.event, tt.script, tt.async)
	}
}

func hookSyncAsyncProjectEntries(t *testing.T) []HookEntry {
	t.Helper()
	path := filepath.Join("..", "..", ".claude", "settings.json")
	raw, err := os.ReadFile(path) // #nosec G304 -- project-root-relative test read
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	entries, err := ParseHookEntries(raw)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return entries
}

func hookSyncAsyncRenderedEntries(t *testing.T) []HookEntry {
	t.Helper()
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	entries, err := RenderHookEntries(fsys, NewTemplateContext(WithPlatform(runtime.GOOS), WithHookOptIn(true)))
	if err != nil {
		t.Fatalf("RenderHookEntries: %v", err)
	}
	return entries
}

func assertHookSyncAsyncEntry(t *testing.T, entries []HookEntry, surface, event, script string, wantAsync bool) {
	t.Helper()
	for _, e := range entries {
		if e.Event == event && e.Script == script {
			if e.Async != wantAsync {
				t.Errorf("%s: %s %s async=%v, want %v", surface, event, script, e.Async, wantAsync)
			}
			return
		}
	}
	t.Fatalf("%s: no %s entry for %s", surface, event, script)
}
