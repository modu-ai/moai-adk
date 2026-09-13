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
		if err := json.Unmarshal(b, &d); err != nil {
			t.Fatal(err)
		}
		if len(d) != 3 || d["theme"] != "dark-ansi" || d["hasCompletedOnboarding"] != true {
			t.Fatal(string(b))
		}
		if err := os.WriteFile(p, []byte(`{"theme":"light"}`), 0600); err != nil {
			t.Fatal(err)
		}
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
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"projects":{"/workspace/new":{"hasTrustDialogAccepted":true}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(selected, ".claude.json"), []byte(`{"projects":{"/workspace/approved":{"hasTrustDialogAccepted":true,"allowedTools":["foreign"]},"/workspace/rejected":{"hasTrustDialogAccepted":false},"/other":{"hasTrustDialogAccepted":true}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, cwd := range []string{"/workspace/approved", "/workspace/rejected", "/workspace/new"} {
		target := t.TempDir()
		if err := seedGatewayUIState(target, gatewayLaunchRequest{CWD: cwd, OriginalConfig: selected, OriginalConfigSet: true}); err != nil {
			t.Fatal(err)
		}
		var state map[string]any
		// An absent file is a valid fresh state (nil map); malformed JSON is not.
		if b, readErr := os.ReadFile(filepath.Join(target, ".claude.json")); readErr == nil {
			if err := json.Unmarshal(b, &state); err != nil {
				t.Fatal(err)
			}
		}
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

func TestGatewayBypassAcceptanceFollowsSelectedConfigOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte(`{"skipDangerousModePermissionPrompt":true,"env":{"SECRET":"excluded"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	accepted, declined := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(accepted, "settings.json"), []byte(`{"skipDangerousModePermissionPrompt":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(declined, "settings.json"), []byte(`{"theme":"light"}`), 0600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		in   gatewayLaunchRequest
		want bool
	}{
		{"default config", gatewayLaunchRequest{}, true},
		{"selected accepted", gatewayLaunchRequest{OriginalConfig: accepted, OriginalConfigSet: true}, true},
		{"selected never accepted", gatewayLaunchRequest{OriginalConfig: declined, OriginalConfigSet: true}, false},
	}
	for _, tc := range cases {
		target := t.TempDir()
		if err := seedGatewayBypassAcceptance(target, tc.in); err != nil {
			t.Fatal(tc.name, err)
		}
		p := filepath.Join(target, "settings.json")
		b, err := os.ReadFile(p)
		if !tc.want {
			if !os.IsNotExist(err) {
				t.Fatalf("%s: invented acceptance: %s", tc.name, b)
			}
			continue
		}
		if string(b) != `{"skipDangerousModePermissionPrompt":true}` {
			t.Fatalf("%s: %s", tc.name, b)
		}
		if err := os.WriteFile(p, []byte(`{"theme":"light"}`), 0600); err != nil {
			t.Fatal(err)
		}
		if err := seedGatewayBypassAcceptance(target, tc.in); err != nil {
			t.Fatal(err)
		}
		if b, _ := os.ReadFile(p); string(b) != `{"theme":"light"}` {
			t.Fatalf("%s: existing settings overwritten", tc.name)
		}
	}
}
