package cli

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGPTClosedVerbs(t *testing.T) {
	for _, verb := range []string{"login", "logout", "status"} {
		t.Run(verb, func(t *testing.T) {
			called := ""
			cmd := newGPTCommand(gptCommandServices{Login: func(context.Context) error { called = "login"; return nil }, Logout: func(context.Context) error { called = "logout"; return nil }, Status: func(context.Context) error { called = "status"; return nil }})
			cmd.SetArgs([]string{verb})
			cmd.SetOut(&bytes.Buffer{})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if called != verb {
				t.Fatal(called)
			}
		})
	}
	for _, args := range [][]string{{"unknown"}, {"login", "extra"}, {"status", "--unexpected"}} {
		cmd := newGPTCommand(gptCommandServices{})
		cmd.SetArgs(args)
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		if err := cmd.Execute(); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestGPTLaunchPreservesCommonEntry(t *testing.T) {
	fakeMoaiProject(t)
	t.Setenv("MOAI_NO_PROFILE_FALLBACK", "1")
	var gotProfile, gotMode string
	var gotArgs []string
	cmd := newGPTCommand(gptCommandServices{Launch: func(profile, mode string, args []string) error {
		gotProfile, gotMode, gotArgs = profile, mode, args
		return nil
	}})
	cmd.SetArgs([]string{"-p", "default", "--model=gpt-6-astra"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if gotProfile != "default" || gotMode != "gpt" || len(gotArgs) != 1 || gotArgs[0] != "--model=gpt-6-astra" {
		t.Fatalf("%s %s %v", gotProfile, gotMode, gotArgs)
	}
}

func TestGPTLocalStatusAndLogoutUsePrivateStore(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	var output bytes.Buffer
	services := newGPTAuthServices(&output, nil)
	if err := services.Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "not logged in") {
		t.Fatal(output.String())
	}
	output.Reset()
	if err := services.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "local") {
		t.Fatal("logout omitted local scope")
	}
	if _, err := os.Stat(filepath.Join(home, "gateway-auth")); err != nil {
		t.Fatal(err)
	}
}

func TestGPTRootRegistrationHasNoGGShortcut(t *testing.T) {
	found, _, err := rootCmd.Find([]string{"gpt"})
	if err != nil || found == rootCmd {
		t.Fatal("gpt command not registered")
	}
	for _, alias := range found.Aliases {
		if alias == "gg" {
			t.Fatal("gg shortcut registered")
		}
	}
	child, _, err := rootCmd.Find([]string{"internal-gateway"})
	if err != nil || child == rootCmd || !child.Hidden {
		t.Fatal("private child command not registered")
	}
}

type cliGPTBrokerFunc func(context.Context, string, bool) error

func (f cliGPTBrokerFunc) Run(ctx context.Context, path string, refresh bool) error {
	return f(ctx, path, refresh)
}

func TestGPTLoginCommitsOnlyCompletedBrokerAndStatusIsRedacted(t *testing.T) {
	legacy := t.TempDir()
	t.Setenv("CODEX_HOME", legacy)
	legacyAuth := filepath.Join(legacy, "auth.json")
	if err := os.WriteFile(legacyAuth, []byte("preserve-existing-login"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", t.TempDir())
	var output bytes.Buffer
	broker := cliGPTBrokerFunc(func(ctx context.Context, home string, refresh bool) error {
		if filepath.Clean(home) == filepath.Clean(legacy) {
			t.Fatal("reused existing Codex home")
		}
		claims, _ := json.Marshal(map[string]any{"exp": time.Now().Add(time.Hour).Unix()})
		token := "e30." + base64.RawURLEncoding.EncodeToString(claims) + ".fixture"
		data, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]string{"access_token": token, "refresh_token": "private-refresh-fixture", "id_token": "fixture", "account_id": "fixture"}})
		return os.WriteFile(filepath.Join(home, "auth.json"), data, 0600)
	})
	services := newGPTAuthServices(&output, broker)
	if err := services.Login(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := services.Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "GPT: logged in") || strings.Contains(output.String(), "private-refresh") {
		t.Fatal(output.String())
	}
	data, readErr := os.ReadFile(legacyAuth)
	if readErr != nil || string(data) != "preserve-existing-login" {
		t.Fatal("existing Codex login changed")
	}
	if err := newGPTAuthServices(&output, nil).Login(context.Background()); err == nil {
		t.Fatal("missing broker accepted")
	}
}

func TestGPTInstalledBrokerFailsWithoutInstalledBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := (installedGPTBroker{Out: &bytes.Buffer{}}).Run(context.Background(), t.TempDir(), false); err == nil {
		t.Fatal("missing codex accepted")
	}
}

func TestGPTInstalledBrokerFailureDoesNotPublishLogin(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	var output bytes.Buffer
	services := newGPTAuthServices(&output, installedGPTBroker{Out: &output})
	if err := services.Login(context.Background()); err == nil {
		t.Fatal("failed broker committed login")
	}
	if err := services.Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "not logged in") {
		t.Fatal(output.String())
	}
}

func TestGatewayCommandsAvoidDependencyGraphInitialization(t *testing.T) {
	for _, args := range [][]string{{"gpt", "status"}, {"gpt", "--model=gpt-6-astra"}, {"internal-gateway"}} {
		if !isTrivialCommand(args) {
			t.Fatalf("gateway command initializes unrelated dependencies: %v", args)
		}
	}
}

func TestGPTUnverifiedLaunchStopsBeforeEntrySideEffects(t *testing.T) {
	cmd := newGPTCommand(gptCommandServices{})
	cmd.SetArgs([]string{"-p"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "verification") {
		t.Fatalf("unverified launch entered argument/profile processing: %v", err)
	}
}

func TestGPTStatusRejectsInaccessibleHomeAndCorruptState(t *testing.T) {
	t.Run("home is a file", func(t *testing.T) {
		home := filepath.Join(t.TempDir(), "home")
		if err := os.WriteFile(home, []byte("preserve"), 0600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("MOAI_HOME", home)
		var output bytes.Buffer
		if err := newGPTAuthServices(&output, nil).Status(context.Background()); err == nil || output.Len() != 0 {
			t.Fatalf("inaccessible home reported success: %s %v", output.String(), err)
		}
		data, _ := os.ReadFile(home)
		if string(data) != "preserve" {
			t.Fatal("home file overwritten")
		}
	})
	t.Run("corrupt state", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("MOAI_HOME", home)
		var output bytes.Buffer
		services := newGPTAuthServices(&output, nil)
		if err := services.Status(context.Background()); err != nil {
			t.Fatal(err)
		}
		output.Reset()
		if err := os.WriteFile(filepath.Join(home, "gateway-auth", "state.json"), []byte("{broken-private-fixture"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := services.Status(context.Background()); err == nil || output.Len() != 0 {
			t.Fatalf("corrupt state reported success: %s %v", output.String(), err)
		}
	})
}

func TestGPTStatusShowsLoginStateWithoutInternalGeneration(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	var output bytes.Buffer
	if err := newGPTAuthServices(&output, nil).Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "generation") || !strings.Contains(output.String(), "not logged in") {
		t.Fatalf("unexpected status %q", output.String())
	}
}
