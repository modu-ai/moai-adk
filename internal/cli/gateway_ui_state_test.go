package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGatewayUIStateSeedsEveryFreshFamilyWithoutSecrets(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	source := []byte(`{"theme":"dark-ansi","hasCompletedOnboarding":true,"lastOnboardingVersion":"2.1.269","oauthAccount":{"secret":"excluded"},"model":"foreign","projects":{"trusted":true}}`)
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), source, 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		target := t.TempDir()
		if err := seedGatewayUIState(target, gatewayLaunchRequest{}); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(target, ".claude.json")
		b, _ := os.ReadFile(p)
		var d map[string]any
		json.Unmarshal(b, &d)
		if len(d) != 3 || d["theme"] != "dark-ansi" || d["hasCompletedOnboarding"] != true {
			t.Fatal(string(b))
		}
		os.WriteFile(p, []byte(`{"theme":"light"}`), 0600)
		if err := seedGatewayUIState(target, gatewayLaunchRequest{}); err != nil {
			t.Fatal(err)
		}
		b, _ = os.ReadFile(p)
		if string(b) != `{"theme":"light"}` {
			t.Fatal("existing state overwritten")
		}
	}
}
func TestGatewayUIStateDoesNotInventOnboardingApproval(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := t.TempDir()
	if err := seedGatewayUIState(target, gatewayLaunchRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, ".claude.json")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestGatewayUIStateRetainsOnlySelectedWorkspaceTrust(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	selected := t.TempDir()
	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"projects":{"/workspace/new":{"hasTrustDialogAccepted":true}}}`), 0600)
	os.WriteFile(filepath.Join(selected, ".claude.json"), []byte(`{"projects":{"/workspace/approved":{"hasTrustDialogAccepted":true,"allowedTools":["foreign"]},"/workspace/rejected":{"hasTrustDialogAccepted":false},"/other":{"hasTrustDialogAccepted":true}}}`), 0600)
	for _, cwd := range []string{"/workspace/approved", "/workspace/rejected", "/workspace/new"} {
		target := t.TempDir()
		if err := seedGatewayUIState(target, gatewayLaunchRequest{CWD: cwd, OriginalConfig: selected, OriginalConfigSet: true}); err != nil {
			t.Fatal(err)
		}
		b, _ := os.ReadFile(filepath.Join(target, ".claude.json"))
		var state map[string]any
		json.Unmarshal(b, &state)
		if cwd == "/workspace/new" {
			if state["projects"] != nil {
				t.Fatal("borrowed global trust")
			}
			continue
		}
		projects := state["projects"].(map[string]any)
		if len(projects) != 1 {
			t.Fatal(projects)
		}
		project := projects[cwd].(map[string]any)
		if len(project) != 1 || project["hasTrustDialogAccepted"] != (cwd == "/workspace/approved") {
			t.Fatal(project)
		}
	}
}
