package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gateway"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// TestGatewayAuthModesDualPath pins the launcher-side automation half of
// AS-021/AS-014:
//
//  1. The GPT production catalog declares only the managed subscription
//     (PKCE) auth method — structurally, there is no API-billing route to
//     switch to, so a subscription failure can never fall through to one.
//  2. The launch binding path never touches the owned subscription token
//     store: credentials stay in the child-owned store and are resolved by
//     the request adapter at send time. The state-directory positive control
//     proves the isolated MOAI_HOME was actually in effect.
//
// The real-account selection and display of both modes stays a live-window
// Gap (AS-021 real PTY).
func TestGatewayAuthModesDualPath(t *testing.T) {
	// 1. Structural no-auto-switch, with a premise assertion so an empty or
	// changed catalog cannot pass vacuously.
	models := gatewayGPTModels()
	if len(models) != 4 {
		t.Fatalf("GPT catalog carries %d entries, want the four registered IDs", len(models))
	}
	for _, m := range models {
		if m.AuthMethod != gateway.AuthPKCE {
			t.Fatalf("GPT entry %s declares auth method %q; only the managed subscription method is registered", m.RouteID, m.AuthMethod)
		}
	}

	// 2. Binding creation touches no token store under an isolated home.
	home := t.TempDir()
	t.Setenv(paths.EnvHome, home)
	binding, err := newGPTGatewayBindingWithStart(func(context.Context, gateway.StartOptions) (gatewayStartedChild, error) {
		return gatewayStartedChild{Address: "127.0.0.1:1234", OverlayPath: "/fixture/settings.json", Stop: func() {}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if binding == nil || binding.Prepare == nil {
		t.Fatal("gateway launch binding unavailable")
	}
	if _, err := os.Stat(filepath.Join(home, "gateway-auth")); !os.IsNotExist(err) {
		t.Fatalf("launch binding path touched the subscription token store: %v", err)
	}
	// Positive control: the isolated home was in effect — the conversation
	// state root was created under it.
	if _, err := os.Stat(filepath.Join(home, "state", "gateway-conversations")); err != nil {
		t.Fatalf("isolated MOAI_HOME fixture was not in effect: %v", err)
	}
}
