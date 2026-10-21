package template

// SPEC-HOOK-SYNC-ASYNC-001 registration pins (plan §F M4/M5): the converted
// hook entries must carry "async": true in BOTH settings surfaces — this
// repository's live .claude/settings.json and the rendered settings template —
// so neither surface can drift back to a synchronous slot the SPEC retired.
// The M4 conversion rides the same commit set as its Go handler change
// (post_tool_failure.go hookSpecificOutput.additionalContext): a flag without
// the delivery change silently drops the message, and vice versa.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// hookSyncAsyncConverted enumerates the entries this SPEC converts to async.
// M4 pins the PostToolUseFailure row (REQ-HSA-003); the M5 commit extends this
// table with the mechanical ConfigChange and SubagentStop chain-event rows
// (REQ-HSA-001/002).
var hookSyncAsyncConverted = []struct {
	event  string
	script string
}{
	{event: "PostToolUseFailure", script: "handle-post-tool-failure.sh"},
}

// TestHookSyncAsyncRegistrationFlags asserts every converted entry parses with
// Async == true from both surfaces (live settings + rendered template).
func TestHookSyncAsyncRegistrationFlags(t *testing.T) {
	project := hookSyncAsyncProjectEntries(t)
	rendered := hookSyncAsyncRenderedEntries(t)

	for _, tt := range hookSyncAsyncConverted {
		assertHookSyncAsyncEntry(t, project, "live settings.json", tt.event, tt.script)
		assertHookSyncAsyncEntry(t, rendered, "rendered settings template", tt.event, tt.script)
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

func assertHookSyncAsyncEntry(t *testing.T, entries []HookEntry, surface, event, script string) {
	t.Helper()
	for _, e := range entries {
		if e.Event == event && e.Script == script {
			if !e.Async {
				t.Errorf("%s: %s %s carries no async key (async=false); SPEC-HOOK-SYNC-ASYNC-001 converts it to async", surface, event, script)
			}
			return
		}
	}
	t.Fatalf("%s: no %s entry for %s", surface, event, script)
}
