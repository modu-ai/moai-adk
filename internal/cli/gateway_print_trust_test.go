package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/gateway/conversation"
)

func TestGatewayPrintRequiresExactWorkspaceTrustWithoutParentInheritance(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	parent := filepath.Join(t.TempDir(), "repo")
	cwd := filepath.Join(parent, ".claude", "worktrees", "repair")
	for _, approved := range []bool{false, true} {
		projects := map[string]any{parent: map[string]bool{"hasTrustDialogAccepted": true}}
		if approved {
			projects[cwd] = map[string]bool{"hasTrustDialogAccepted": true}
		}
		raw, _ := json.Marshal(map[string]any{"projects": projects})
		if err := os.WriteFile(filepath.Join(home, ".claude.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
		target := t.TempDir()
		in := gatewayLaunchRequest{CWD: cwd, Args: []string{"--print", "query"}}
		if err := seedGatewayUIState(target, in); err != nil {
			t.Fatal(err)
		}
		err := requireGatewayPrintTrust(target, in)
		if approved && err != nil {
			t.Fatal(err)
		}
		if !approved && (err == nil || !strings.Contains(err.Error(), "trust")) {
			t.Fatal("parent trust silently authorized worktree")
		}
		in.Args = nil
		if err := requireGatewayPrintTrust(target, in); err != nil {
			t.Fatal("interactive trust dialog prevented", err)
		}
		in.Args = []string{"--", "--print"}
		if err := requireGatewayPrintTrust(target, in); err != nil {
			t.Fatal("literal prompt interpreted as a headless option", err)
		}
	}
}

func TestGatewayPrintUntrustedFamilyStopsBeforeChildOrPayload(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	families, err := conversation.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := gateway.NewSessionCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	started, payload := false, false
	binding := newGatewaySessionBinding(gatewaySessionOptions{Mode: "gpt", Catalog: catalog, Family: families,
		Payload: func(conversation.Descriptor) (json.RawMessage, error) { payload = true; return nil, nil },
		Start: func(context.Context, gateway.StartOptions) (gatewayStartedChild, error) {
			started = true
			return gatewayStartedChild{}, nil
		},
	})
	_, cleanup, err := binding.Prepare(gatewayLaunchRequest{Mode: "gpt", CWD: t.TempDir(), Project: "fixture", Args: []string{"--print", "query"}})
	if cleanup != nil {
		cleanup()
	}
	if err == nil || !strings.Contains(err.Error(), "trust") || started || payload {
		t.Fatalf("untrusted request progressed: err=%v child=%v payload=%v", err, started, payload)
	}
}

func TestGatewayPrintTrustFailsClosedForMalformedOrRejectedState(t *testing.T) {
	for _, raw := range []string{`{`, `{"projects":{"/work":{"hasTrustDialogAccepted":false}}}`, `{"projects":{"/work":{"hasTrustDialogAccepted":"true"}}}`, `{}`} {
		target := t.TempDir()
		if err := os.WriteFile(filepath.Join(target, ".claude.json"), []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"--print"}, {"-p"}} {
			if err := requireGatewayPrintTrust(target, gatewayLaunchRequest{CWD: "/work", Args: args}); err == nil {
				t.Fatal("untrusted noninteractive launch allowed")
			}
		}
	}
}
