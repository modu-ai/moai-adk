package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestSessionStartProfileLeaseDirectAndTokenFallback(t *testing.T) {
	registerProfileLease(context.Background(), nil)
	home := filepath.Join(t.TempDir(), "home")
	profile := filepath.Join(home, "claude-profiles", "opus")
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", profile)
	input := &HookInput{SessionID: "direct", ProjectDir: t.TempDir()}
	registerProfileLease(context.Background(), input)
	store, err := homestate.OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	var direct int
	if err := store.DB.QueryRow(`SELECT count(*) FROM profile_leases WHERE session_id='direct' AND profile_path=?`, profile).Scan(&direct); err != nil {
		t.Fatal(err)
	}
	token, err := store.CreateProvisional(context.Background(), homestate.ProfileLease{ProfileName: "opus", ProfilePath: profile, PID: os.Getpid(), ProcessFingerprint: homestate.CurrentProcessFingerprint()})
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Close()
	t.Setenv("MOAI_PROFILE_LEASE_TOKEN", token)
	input.SessionID = "enriched"
	registerProfileLease(context.Background(), input)
	store, err = homestate.OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	var enriched int
	if err := store.DB.QueryRow(`SELECT count(*) FROM profile_leases WHERE token=? AND session_id='enriched'`, token).Scan(&enriched); err != nil {
		t.Fatal(err)
	}
	_ = store.Close()
	if direct != 1 || enriched != 1 {
		t.Fatalf("direct=%d enriched=%d", direct, enriched)
	}
	t.Setenv("MOAI_PROFILE_LEASE_TOKEN", "missing-token")
	input.SessionID = "fallback"
	registerProfileLease(context.Background(), input)
	store, err = homestate.OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close profile lease store: %v", err)
		}
	})
	var fallback int
	if err := store.DB.QueryRow(`SELECT count(*) FROM profile_leases WHERE session_id='fallback'`).Scan(&fallback); err != nil {
		t.Fatal(err)
	}
	if fallback != 1 {
		t.Fatalf("fallback=%d", fallback)
	}
}

func TestSessionStartProfileLeaseUnavailableFailsClosedWithoutPanic(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", blocked)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(t.TempDir(), "profile"))
	registerProfileLease(context.Background(), &HookInput{SessionID: "s", ProjectDir: t.TempDir()})
}
