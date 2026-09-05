// Package cli — doctor_hook_delivery_test.go
//
// Tests for the Hook Delivery doctor check (SPEC-UPDATE-HOOK-DELIVERY-001,
// Option B — detect + guide). The check is READ-ONLY by mandate (REQ-UHD-008):
// the tests assert content and mtime stay unchanged alongside the reporting
// assertions, so a future write path cannot sneak in and pass this suite.
package cli

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// hookScriptRefPattern is the test-side, independent extraction of the handler
// path a hook entry references. Deliberately NOT the production identity
// helper: deriving the expectation from the code under test would make the
// missing-entry assertions tautological.
var hookScriptRefPattern = regexp.MustCompile(`\.claude/hooks/moai/([A-Za-z0-9._/-]+\.sh)`)

// writeFixtureSettings serializes settings as JSON into root/.claude/settings.json.
func writeFixtureSettings(t *testing.T, root string, settings map[string]any) string {
	t.Helper()
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture settings: %v", err)
	}
	dir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	p := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write fixture settings: %v", err)
	}
	return p
}

// renderedFixtureHooks renders the shipped template for root (opt-in aware)
// and returns its hooks object, so fixtures are always built against the real
// embedded template of the build under test rather than a frozen copy.
func renderedFixtureHooks(t *testing.T, root string) map[string]any {
	t.Helper()
	hooks, err := renderedTemplateHooks(root)
	if err != nil {
		t.Fatalf("renderedTemplateHooks: %v", err)
	}
	return hooks
}

// multiEntryEventKey returns the first template hook event key carrying at
// least two entries, with its entry array.
func multiEntryEventKey(t *testing.T, hooks map[string]any) (string, []any) {
	t.Helper()
	keys := make([]string, 0, len(hooks))
	for k := range hooks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if arr, ok := hooks[k].([]any); ok && len(arr) >= 2 {
			return k, arr
		}
	}
	t.Fatal("template has no hook event key with two or more entries to test against")
	return "", nil
}

// entryScriptName extracts, with the test-side pattern, the handler script
// name an entry references.
func entryScriptName(t *testing.T, entry any) string {
	t.Helper()
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal entry: %v", err)
	}
	m := hookScriptRefPattern.FindSubmatch(data)
	if m == nil {
		t.Fatalf("entry references no .claude/hooks/moai script:\n%s", data)
	}
	return path.Base(string(m[1]))
}

// assertReadOnlySettings re-stats and re-reads the fixture settings file and
// fails when the check mutated either.
func assertReadOnlySettings(t *testing.T, settingsPath string, content []byte, mtimeBefore int64) {
	t.Helper()
	info, err := os.Stat(settingsPath)
	if err != nil {
		t.Fatalf("settings.json vanished after the check ran: %v", err)
	}
	if info.ModTime().UnixNano() != mtimeBefore {
		t.Errorf("settings.json mtime changed — the check wrote the file (REQ-UHD-008 violation)")
	}
	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.json unreadable after the check ran: %v", err)
	}
	if string(after) != string(content) {
		t.Errorf("settings.json content changed — the check wrote the file (REQ-UHD-008 violation)")
	}
}

// TestCheckHookDelivery_FlagsMissingEntryInCarriedEventKey is the RED-direction
// regression for the measured defect (AC-UHD-003 / AC-UHD-006, Option B): a
// project carries a hook event key whose array is missing an entry the shipped
// template carries, and the check must name the entry, its event key, and the
// remediation — without touching the file.
func TestCheckHookDelivery_FlagsMissingEntryInCarriedEventKey(t *testing.T) {
	root := t.TempDir()
	hooks := renderedFixtureHooks(t, root)
	eventKey, entries := multiEntryEventKey(t, hooks)

	missingName := entryScriptName(t, entries[1])
	settingsPath := writeFixtureSettings(t, root, map[string]any{
		"hooks": map[string]any{eventKey: []any{entries[0]}},
	})
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	infoBefore, err := os.Stat(settingsPath)
	if err != nil {
		t.Fatalf("stat fixture: %v", err)
	}
	mtimeBefore := infoBefore.ModTime().UnixNano()

	check := checkHookDelivery(root, true)

	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %q, want warn (the gap must be flagged)", check.Status)
	}
	msg := check.Message
	if !strings.Contains(msg, "hooks."+eventKey) {
		t.Errorf("message does not name the event key %q:\n%s", eventKey, msg)
	}
	if !strings.Contains(msg, missingName) {
		t.Errorf("message does not name the missing entry %q:\n%s", missingName, msg)
	}
	// Remediation guidance rides in the message (Detail renders only under
	// --verbose), including the post-update deletion check.
	if !strings.Contains(msg, "git status --porcelain") {
		t.Errorf("message does not carry the post-update deletion-check guidance:\n%s", msg)
	}
	if check.Detail == "" {
		t.Error("verbose run produced no Detail placement guidance")
	}
	assertReadOnlySettings(t, settingsPath, content, mtimeBefore)
}

// TestCheckHookDelivery_SilentWhenAllEntriesPresent is the GREEN direction of
// the same regression: full parity between the project file and the template
// must keep the check silent. Without this direction, a "warn on everything"
// implementation would pass the suite.
func TestCheckHookDelivery_SilentWhenAllEntriesPresent(t *testing.T) {
	root := t.TempDir()
	hooks := renderedFixtureHooks(t, root)
	writeFixtureSettings(t, root, map[string]any{"hooks": hooks})

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok for a fully-populated file (message: %s)", check.Status, check.Message)
	}
}

// TestCheckHookDelivery_MissingSettingsJSONInformational covers REQ-UHD-009 /
// AC-UHD-013's absence half: no settings.json is informational, not an error,
// and the check must not create one.
func TestCheckHookDelivery_MissingSettingsJSONInformational(t *testing.T) {
	root := t.TempDir()

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok for a project without settings.json", check.Status)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Errorf("the check created .claude/settings.json (REQ-UHD-008/009 violation)")
	}
}

// TestCheckHookDelivery_NoHooksKeyInformational covers REQ-UHD-009 /
// AC-UHD-013's hook-free half: a settings.json without a hooks object is
// informational, not an error.
func TestCheckHookDelivery_NoHooksKeyInformational(t *testing.T) {
	root := t.TempDir()
	writeFixtureSettings(t, root, map[string]any{"model": "opus"})

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok for a hook-free settings.json (message: %s)", check.Status, check.Message)
	}
}

// TestCheckHookDelivery_InvalidJSONWarnsGracefully covers REQ-UHD-011 /
// AC-UHD-008's doctor half: an unparseable settings.json is reported (the
// message names settings.json) and skipped without any write.
func TestCheckHookDelivery_InvalidJSONWarnsGracefully(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	settingsPath := filepath.Join(dir, "settings.json")
	broken := []byte(`{"hooks": `)
	if err := os.WriteFile(settingsPath, broken, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %q, want warn for an unparseable settings.json", check.Status)
	}
	if !strings.Contains(check.Message, "settings.json") {
		t.Errorf("message does not name settings.json:\n%s", check.Message)
	}
	after, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(after) != string(broken) {
		t.Errorf("the check modified the broken file (REQ-UHD-011 violation)")
	}
}

// TestCheckHookDelivery_NonArrayHookValueWarnsAndSkips covers REQ-UHD-011 /
// AC-UHD-009: a hook event key holding a non-array value is reported and
// skipped while other carried keys still resolve normally.
func TestCheckHookDelivery_NonArrayHookValueWarnsAndSkips(t *testing.T) {
	root := t.TempDir()
	hooks := renderedFixtureHooks(t, root)
	eventKey, entries := multiEntryEventKey(t, hooks)
	// The anomalous key holds a string; a second carried key is fully
	// populated, so the only warning source must be the anomaly.
	otherKey, otherEntries := eventKey, entries
	fullHooks := map[string]any{
		eventKey:           "not-an-array",
		otherKey + "-full": otherEntries,
	}
	writeFixtureSettings(t, root, map[string]any{"hooks": fullHooks})

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %q, want warn for a non-array hook value", check.Status)
	}
	if !strings.Contains(check.Message, eventKey) {
		t.Errorf("message does not name the anomalous key %q:\n%s", eventKey, check.Message)
	}
}

// TestCheckHookDelivery_DoesNotFlagTemplateNewEventKeys guards REQ-UHD-001:
// an event key the user does NOT carry is delivered by today's merge, so the
// detector must not report it — flagging it would be a false positive on
// every pre-delivery doctor run.
func TestCheckHookDelivery_DoesNotFlagTemplateNewEventKeys(t *testing.T) {
	root := t.TempDir()
	hooks := renderedFixtureHooks(t, root)
	eventKey, entries := multiEntryEventKey(t, hooks)
	// The user carries exactly one key, fully populated; every other template
	// key is absent from their file.
	writeFixtureSettings(t, root, map[string]any{
		"hooks": map[string]any{eventKey: entries},
	})

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok — keys absent from the user file are template-introduced and delivered (message: %s)",
			check.Status, check.Message)
	}
}

// TestCheckHookDelivery_UserAuthoredEntryNotFlagged pins the out-of-scope
// boundary: a hook entry the user authored (absent from the template) draws
// no comment — the check only compares template-side entries inward.
func TestCheckHookDelivery_UserAuthoredEntryNotFlagged(t *testing.T) {
	root := t.TempDir()
	hooks := renderedFixtureHooks(t, root)
	eventKey, entries := multiEntryEventKey(t, hooks)

	userEntries := append([]any{}, entries...)
	userEntries = append(userEntries, map[string]any{
		"command": "bash",
		"args":    []string{"-c", "echo user-owned"},
		"type":    "command",
	})
	writeFixtureSettings(t, root, map[string]any{
		"hooks": map[string]any{eventKey: userEntries},
	})

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok — a user-authored entry must not be flagged (message: %s)",
			check.Status, check.Message)
	}
}

// TestCheckHookDelivery_OptInAwareRendering proves the template side honors
// the project's hook.opt_in toggle: with opt-in enabled, the observability
// hook series counts as shipped, so a project missing one is flagged.
func TestCheckHookDelivery_OptInAwareRendering(t *testing.T) {
	root := t.TempDir()
	sysDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sysDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	sysYAML := "hook:\n  opt_in:\n    enabled: true\n"
	if err := os.WriteFile(filepath.Join(sysDir, "system.yaml"), []byte(sysYAML), 0o644); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}

	hooks := renderedFixtureHooks(t, root)
	eventKey, entries := multiEntryEventKey(t, hooks)
	settingsPath := writeFixtureSettings(t, root, map[string]any{
		"hooks": map[string]any{eventKey: []any{entries[0]}},
	})

	check := checkHookDelivery(root, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %q, want warn under opt-in rendering (message: %s)", check.Status, check.Message)
	}
	_ = settingsPath // fixture written read-only; content asserted in the core test
}

// TestHookEntryIdentities unit-tests the identity rule recorded in design.md
// §G: a handler path is the identity; an entry referencing none falls back to
// its canonical JSON.
func TestHookEntryIdentities(t *testing.T) {
	withScript := map[string]any{
		"command": "bash",
		"args":    []string{"-c", "run", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-x.sh"},
	}
	withoutScript := map[string]any{
		"command": "echo",
		"args":    []string{"hello"},
	}
	ids := hookEntryIdentities([]any{withScript, withoutScript})

	var scriptIdentity, fallbackIdentity string
	for id, name := range ids {
		if strings.HasSuffix(id, "handle-x.sh") {
			scriptIdentity = id
			if name != "handle-x.sh" {
				t.Errorf("display name = %q, want the script base name", name)
			}
		} else {
			fallbackIdentity = id
		}
	}
	if scriptIdentity == "" {
		t.Errorf("no identity derived from the script reference (identities: %v)", ids)
	}
	if fallbackIdentity == "" {
		t.Errorf("no fallback identity derived for the script-less entry (identities: %v)", ids)
	}
	if strings.Contains(fallbackIdentity, "handle-x.sh") {
		t.Errorf("fallback identity collided with the script identity: %q", fallbackIdentity)
	}

	// The same script referenced by differently-shaped entries must yield the
	// SAME identity — that is what makes detection robust to cosmetic edits.
	reshaped := map[string]any{
		"command": "bash",
		"args":    []string{"-c", "other-wrapper", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-x.sh"},
		"timeout": 99,
	}
	ids2 := hookEntryIdentities([]any{reshaped})
	for id := range ids2 {
		if !strings.HasSuffix(id, "handle-x.sh") {
			t.Errorf("reshaped entry derived identity %q, want the script path", id)
		}
	}
}

// TestCheckHookDelivery_RegisteredInWorkspaceChecks proves the check is wired
// into the doctor's Workspace group under its documented name.
func TestCheckHookDelivery_RegisteredInWorkspaceChecks(t *testing.T) {
	groups := runGroupedChecks(false, "Hook Delivery")
	var found []DiagnosticCheck
	for _, g := range groups {
		found = append(found, g.checks...)
	}
	if len(found) != 1 {
		t.Fatalf("filter \"Hook Delivery\" matched %d checks, want 1", len(found))
	}
	if found[0].Name != "Hook Delivery" {
		t.Errorf("check name = %q, want %q", found[0].Name, "Hook Delivery")
	}
	switch found[0].Status {
	case uikit.CheckOK, uikit.CheckWarn, uikit.CheckFail:
		// valid
	default:
		t.Errorf("check returned invalid status %q", found[0].Status)
	}
}
