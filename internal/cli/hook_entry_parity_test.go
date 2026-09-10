package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// Hook-entry parity between this repository's own .claude/settings.json and the
// settings template it ships.
//
// The doctor check (doctor_hook_wiring.go) reports drift for whatever project it
// is pointed at; this test pins THIS project, so a divergence here fails the
// build rather than waiting for someone to read a warn row.
//
// It reuses the same three helpers the check uses — ParseHookEntries,
// RenderHookEntries, DiffHookEntries — rather than re-deriving the comparison.
// A second implementation of the same tuple diff could pass while the shipped
// one is broken, which would make this test worse than nothing.

// parityProjectRoot walks up from the test's working directory to the tree root
// (the directory holding go.mod). Returning a wrong root would make every
// assertion below vacuous, so a failure to find it fails the test.
func parityProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod found in any ancestor; cannot locate the project root")
		}
		dir = parent
	}
}

// parityEntries returns (templateEntries, projectEntries) for the project root,
// rendered and parsed exactly as the doctor check does.
func parityEntries(t *testing.T) ([]template.HookEntry, []template.HookEntry) {
	t.Helper()
	root := parityProjectRoot(t)

	settingsPath := filepath.Join(root, ".claude", "settings.json")
	projectBytes, err := os.ReadFile(settingsPath) // #nosec G304 -- project-root-relative test read
	if err != nil {
		t.Fatalf("read %s: %v", settingsPath, err)
	}
	projectEntries, err := template.ParseHookEntries(projectBytes)
	if err != nil {
		t.Fatalf("parse %s: %v", settingsPath, err)
	}

	ctx := template.NewTemplateContext(
		template.WithPlatform(runtime.GOOS),
		template.WithVersion(version.GetVersion()),
		template.WithHookOptIn(config.LoadSystemHookOptInEnabled(root)),
	)
	// Same seam the doctor check uses in production, so the test renders the
	// template the shipped code renders.
	tmplFS := hookWiringTemplateSource()
	if tmplFS == nil {
		t.Fatal("embedded template source unavailable; the comparison would be vacuous")
	}
	templateEntries, err := template.RenderHookEntries(tmplFS, ctx)
	if err != nil {
		t.Fatalf("render settings template: %v", err)
	}

	if len(templateEntries) == 0 || len(projectEntries) == 0 {
		t.Fatalf("empty entry set makes every assertion below vacuous: template=%d project=%d",
			len(templateEntries), len(projectEntries))
	}
	return templateEntries, projectEntries
}

// TestHookEntryParity_ChainEventRegistered is AC-HWD-003 (a).
func TestHookEntryParity_ChainEventRegistered(t *testing.T) {
	_, projectEntries := parityEntries(t)

	var found []template.HookEntry
	for _, e := range projectEntries {
		if e.Script == "chain-event.sh" {
			found = append(found, e)
		}
	}
	if len(found) != 1 {
		t.Fatalf("SubagentStop must carry exactly one chain-event.sh entry, found %d", len(found))
	}
	got := found[0]
	if got.Event != "SubagentStop" {
		t.Errorf("chain-event.sh event = %q, want SubagentStop", got.Event)
	}
	if got.Timeout != 5 {
		t.Errorf("chain-event.sh timeout = %d, want 5", got.Timeout)
	}
	if got.Async {
		t.Error("chain-event.sh must carry no async key (async=false); it is synchronous")
	}
}

// TestHookEntryParity_StatusTransitionScoped is AC-HWD-003 (b).
//
// The single unscoped synchronous entry this replaced would satisfy a
// script-name-only check, which is why the assertion is on the if-predicate set
// and the async/timeout values rather than on presence.
func TestHookEntryParity_StatusTransitionScoped(t *testing.T) {
	_, projectEntries := parityEntries(t)

	wantIf := map[string]bool{
		"Write(**/.moai/specs/**)":     false,
		"Edit(**/.moai/specs/**)":      false,
		"MultiEdit(**/.moai/specs/**)": false,
	}

	var found []template.HookEntry
	for _, e := range projectEntries {
		if e.Script == "status-transition-ownership.sh" {
			found = append(found, e)
		}
	}
	if len(found) != 3 {
		t.Fatalf("PostToolUse must carry exactly 3 status-transition-ownership.sh entries, found %d: %v",
			len(found), found)
	}
	for _, e := range found {
		if e.Event != "PostToolUse" {
			t.Errorf("%s: event = %q, want PostToolUse", e, e.Event)
		}
		if !e.Async {
			t.Errorf("%s: async = false, want true", e)
		}
		if e.Timeout != 5 {
			t.Errorf("%s: timeout = %d, want 5", e, e.Timeout)
		}
		seen, known := wantIf[e.If]
		if !known {
			t.Errorf("%s: unexpected if predicate %q", e, e.If)
			continue
		}
		if seen {
			t.Errorf("%s: duplicate if predicate %q", e, e.If)
		}
		wantIf[e.If] = true
	}
	for pred, seen := range wantIf {
		if !seen {
			t.Errorf("missing status-transition-ownership.sh entry with if=%q", pred)
		}
	}
}

// TestHookEntryParity_NoDivergenceEitherDirection is AC-HWD-003 (c).
//
// Both directions are asserted: a one-directional comparison misses an extra
// project registration, which is exactly the shape the unscoped entry had.
func TestHookEntryParity_NoDivergenceEitherDirection(t *testing.T) {
	templateEntries, projectEntries := parityEntries(t)

	templateOnly, projectOnly := template.DiffHookEntries(templateEntries, projectEntries)

	for _, e := range templateOnly {
		t.Errorf("template-only (registered in the template, missing from this project): %s", e)
	}
	for _, e := range projectOnly {
		t.Errorf("project-only (registered in this project, absent from the template): %s", e)
	}
	if len(templateOnly) > 0 || len(projectOnly) > 0 {
		t.Fatalf("hook-entry parity: %d template-only, %d project-only",
			len(templateOnly), len(projectOnly))
	}
}
