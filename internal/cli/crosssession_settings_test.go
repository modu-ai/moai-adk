package cli

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// crosssession_settings_test.go — the crosssession.yaml → injected --settings
// translation, plus the two guard tests that pin the cross-machine approval
// posture (no template or launcher default ever turns isolatePeerMachines on).

// withCrossSessionConfig points the launcher's config root at a temp project
// root carrying the given crosssession.yaml body (empty body = no file).
func withCrossSessionConfig(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	if body != "" {
		sections := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(sections, 0o755); err != nil {
			t.Fatalf("mkdir sections: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sections, "crosssession.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("write crosssession.yaml: %v", err)
		}
	}
	orig := crossSessionConfigRootFn
	crossSessionConfigRootFn = func() string { return root }
	t.Cleanup(func() { crossSessionConfigRootFn = orig })
	return root
}

// TestCrossSessionSettingsPayloadDefault verifies the neutral-config contract:
// no crosssession.yaml (or an all-neutral one) produces an EMPTY payload, so
// the launcher injects nothing and Claude Code's own defaults apply.
func TestCrossSessionSettingsPayloadDefault(t *testing.T) {
	root := withCrossSessionConfig(t, "")

	if got := crossSessionSettingsPayload(root); len(got) != 0 {
		t.Errorf("payload = %v, want empty (neutral config injects nothing)", got)
	}
}

// TestCrossSessionSettingsPayloadFull verifies the translation table:
// inbound → crossSessionInbound, dialog_expiry → dialogExpiry, and that a
// false isolate_machines emits NO isolatePeerMachines key (omission, not
// false — the key must never appear uninvited).
func TestCrossSessionSettingsPayloadFull(t *testing.T) {
	root := withCrossSessionConfig(t, "crosssession:\n  inbound: hold\n  isolate_machines: false\n  dialog_expiry: 10m\n")

	got := crossSessionSettingsPayload(root)
	if got["crossSessionInbound"] != "hold" {
		t.Errorf("crossSessionInbound = %v, want \"hold\"", got["crossSessionInbound"])
	}
	if got["dialogExpiry"] != "10m" {
		t.Errorf("dialogExpiry = %v, want \"10m\"", got["dialogExpiry"])
	}
	if _, present := got["isolatePeerMachines"]; present {
		t.Errorf("payload carries isolatePeerMachines despite isolate_machines: false — the key must be omitted, not set false")
	}
	if len(got) != 2 {
		t.Errorf("payload = %v, want exactly 2 keys", got)
	}
}

// TestCrossSessionSettingsPayloadIsolateOptIn verifies the ONLY path by which
// the launcher ever emits isolatePeerMachines: the user's explicit
// isolate_machines: true in their own crosssession.yaml.
func TestCrossSessionSettingsPayloadIsolateOptIn(t *testing.T) {
	root := withCrossSessionConfig(t, "crosssession:\n  isolate_machines: true\n")

	got := crossSessionSettingsPayload(root)
	v, present := got["isolatePeerMachines"]
	if !present || v != true {
		t.Errorf("isolatePeerMachines = %v (present=%t), want explicit true on opt-in", v, present)
	}
}

// TestCrossSessionSettingsPayloadFiltersInvalidValues verifies the closed
// value sets: an inbound or dialog_expiry outside the documented Claude Code
// enum is dropped (never forwarded) rather than injected as-is.
func TestCrossSessionSettingsPayloadFiltersInvalidValues(t *testing.T) {
	root := withCrossSessionConfig(t, "crosssession:\n  inbound: maybe\n  dialog_expiry: 3h\n")

	if got := crossSessionSettingsPayload(root); len(got) != 0 {
		t.Errorf("payload = %v, want empty (invalid enum values are dropped, not forwarded)", got)
	}
}

// TestLauncherNeverInjectsIsolatePeerMachinesByDefault is the launcher half of
// the cross-machine-approval guard: without the user's explicit opt-in, NO
// launcher-produced payload — the general-launch payload or the kanban-mode
// payload — carries isolatePeerMachines. The documented default posture (no
// approval for cross-machine messages) must hold unless the user turns it on.
func TestLauncherNeverInjectsIsolatePeerMachinesByDefault(t *testing.T) {
	neutralConfigs := []struct {
		name string
		body string
	}{
		{"no config file", ""},
		{"neutral template default", "crosssession:\n  inbound: \"\"\n  isolate_machines: false\n  dialog_expiry: \"\"\n"},
		{"inbound only", "crosssession:\n  inbound: accept\n"},
		{"dialog expiry only", "crosssession:\n  dialog_expiry: never\n"},
	}
	for _, c := range neutralConfigs {
		t.Run(c.name, func(t *testing.T) {
			root := withCrossSessionConfig(t, c.body)
			if _, present := crossSessionSettingsPayload(root)["isolatePeerMachines"]; present {
				t.Errorf("general-launch payload carries isolatePeerMachines without opt-in")
			}
			// The kanban payload too: forced accept + user extras, but the
			// isolation key appears only on explicit opt-in.
			flag, cleanup := prepareKanbanSettings([]string{"-p", "dev"})
			t.Cleanup(cleanup)
			if len(flag) != 2 {
				t.Fatalf("kanban settings flag = %v, want [--settings <path>]", flag)
			}
			data, err := os.ReadFile(flag[1])
			if err != nil {
				t.Fatalf("read kanban settings file: %v", err)
			}
			if strings.Contains(string(data), "isolatePeerMachines") {
				t.Errorf("kanban settings file carries isolatePeerMachines without opt-in: %s", data)
			}
		})
	}
}

// TestTemplateNeverShipsIsolatePeerMachines is the template half of the guard:
// the distributed templates must never carry the Claude Code settings key
// `isolatePeerMachines` in any settings-bearing (non-prose) file. Rule and
// skill prose (.md) documents the key; config/json/tmpl files must not set it.
// A true from ANY Claude Code scope applies and cannot be turned off from a
// lower scope, so the shipped default must never introduce one.
func TestTemplateNeverShipsIsolatePeerMachines(t *testing.T) {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	var offenders []string
	err = fs.WalkDir(embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasSuffix(path, ".md") {
			return nil // prose surfaces document the key; only settings-bearing files set it
		}
		data, readErr := fs.ReadFile(embedded, path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(data), "isolatePeerMachines") {
			offenders = append(offenders, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded templates: %v", err)
	}
	if len(offenders) > 0 {
		t.Errorf("settings-bearing template files carry isolatePeerMachines (a shipped true can never be turned off from a lower scope): %v", offenders)
	}

	// The shipped crosssession.yaml must carry the neutral default.
	data, err := fs.ReadFile(embedded, filepath.ToSlash(".moai/config/sections/crosssession.yaml"))
	if err != nil {
		t.Fatalf("template crosssession.yaml missing from the embedded FS: %v", err)
	}
	if !strings.Contains(string(data), "isolate_machines: false") {
		t.Errorf("template crosssession.yaml does not pin isolate_machines: false:\n%s", data)
	}
}

// TestPrepareKanbanSettingsMergesUserConfig verifies the kanban merge: the
// kanban-required crossSessionInbound=accept wins (dispatch would stall
// otherwise), while the user's dialog_expiry / isolate opt-in ride along.
func TestPrepareKanbanSettingsMergesUserConfig(t *testing.T) {
	withCrossSessionConfig(t, "crosssession:\n  inbound: refuse\n  isolate_machines: true\n  dialog_expiry: never\n")

	for _, key := range []string{config.EnvMoaiKanbanSettingsInjected} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}

	flag, cleanup := prepareKanbanSettings([]string{"-p", "dev"})
	t.Cleanup(cleanup)
	if len(flag) != 2 || flag[0] != "--settings" {
		t.Fatalf("flag = %v, want [--settings <path>]", flag)
	}
	data, err := os.ReadFile(flag[1])
	if err != nil {
		t.Fatalf("read settings file: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("settings file is not valid JSON: %v (%s)", err, data)
	}
	if got["crossSessionInbound"] != "accept" {
		t.Errorf("crossSessionInbound = %v, want \"accept\" (kanban dispatch requirement)", got["crossSessionInbound"])
	}
	if got["dialogExpiry"] != "never" {
		t.Errorf("dialogExpiry = %v, want \"never\" (user preference rides along)", got["dialogExpiry"])
	}
	if got["isolatePeerMachines"] != true {
		t.Errorf("isolatePeerMachines = %v, want true (explicit opt-in honored)", got["isolatePeerMachines"])
	}
}

// TestAppendCrossSessionSettingsGeneralLaunch verifies the general-launch
// injection consumed by unifiedLaunchDefault: injects only when the operator
// did not supply --settings themselves AND the config carries at least one
// non-neutral value.
func TestAppendCrossSessionSettingsGeneralLaunch(t *testing.T) {
	t.Run("no config → unchanged args", func(t *testing.T) {
		root := withCrossSessionConfig(t, "")
		got := appendCrossSessionSettings(root, []string{"-p", "dev"})
		if len(got) != 2 || got[0] != "-p" || got[1] != "dev" {
			t.Errorf("args = %v, want [-p dev] unchanged (neutral config injects nothing)", got)
		}
	})
	t.Run("operator --settings → no injection", func(t *testing.T) {
		root := withCrossSessionConfig(t, "crosssession:\n  inbound: accept\n")
		got := appendCrossSessionSettings(root, []string{"--settings", "/tmp/operator.json"})
		if len(got) != 2 {
			t.Errorf("args = %v, want unchanged (operator-supplied --settings wins)", got)
		}
	})
	t.Run("config → --settings appended with payload", func(t *testing.T) {
		root := withCrossSessionConfig(t, "crosssession:\n  inbound: refuse\n  dialog_expiry: 60s\n")
		got := appendCrossSessionSettings(root, []string{"-p", "dev"})
		if len(got) != 4 || got[2] != "--settings" {
			t.Fatalf("args = %v, want [-p dev --settings <path>]", got)
		}
		data, err := os.ReadFile(got[3])
		if err != nil {
			t.Fatalf("read injected settings: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("injected settings not valid JSON: %v (%s)", err, data)
		}
		if payload["crossSessionInbound"] != "refuse" || payload["dialogExpiry"] != "60s" {
			t.Errorf("payload = %v, want refuse + 60s", payload)
		}
		if _, present := payload["isolatePeerMachines"]; present {
			t.Errorf("payload carries isolatePeerMachines without opt-in")
		}
		t.Cleanup(func() { _ = os.Remove(got[3]) })
	})
}
