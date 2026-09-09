package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestProfileLeaseLifecycleAndNonExecCleanerRace(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	store, err := homestate.OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	token, err := store.CreateProvisional(ctx, homestate.ProfileLease{ProfileName: "opus", ProfilePath: filepath.Join(home, "claude-profiles", "opus"), PID: 100, ProcessFingerprint: "parent"})
	if err != nil {
		t.Fatal(err)
	}
	if err := transferProfileLeaseToChild([]string{"MOAI_PROFILE_LEASE_TOKEN=" + token}, 100, "parent", 101, "child"); err != nil {
		t.Fatal(err)
	}
	if err := store.Enrich(ctx, token, "session-1", 101, "child"); err != nil {
		t.Fatal(err)
	}
	if err := store.ReleaseSession(ctx, "session-1"); err != nil {
		t.Fatal(err)
	}
}

func TestCleanHomeReportsProtectedProfileReasonInDryRunAndForce(t *testing.T) {
	root := seedHomeCleanFixture(t)
	profile := filepath.Join(root, "claude-profiles", "p1")
	t.Setenv("CLAUDE_CONFIG_DIR", profile)
	protected := filepath.Join(profile, "debug", "old.log")
	for _, force := range []bool{false, true} {
		p, _, errBuf := newHomeTestPrinter()
		if err := runCleanHome(p, force); err != nil {
			t.Fatal(err)
		}
		out := errBuf.String()
		if !strings.Contains(out, "p1") || !strings.Contains(out, "live") || !strings.Contains(out, "deleted 0") {
			t.Fatalf("force=%v output=%q", force, out)
		}
		if _, err := os.Stat(protected); err != nil {
			t.Fatalf("force=%v protected deleted: %v", force, err)
		}
	}
}

func TestProfileLeaseChildTransferRejectsIndeterminateAndStaleToken(t *testing.T) {
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
	if err := transferProfileLeaseToChild(nil, 1, "parent", 2, "child"); err != nil {
		t.Fatal(err)
	}
	if err := transferProfileLeaseToChild([]string{"MOAI_PROFILE_LEASE_TOKEN=token"}, 1, "parent", 0, ""); err == nil {
		t.Fatal("indeterminate child accepted")
	}
	if err := transferProfileLeaseToChild([]string{"OTHER=x", "MOAI_PROFILE_LEASE_TOKEN=missing"}, 1, "parent", 2, "child"); err == nil {
		t.Fatal("missing lease token CAS accepted")
	}
	blockedHome := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockedHome, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", blockedHome)
	if err := transferProfileLeaseToChild([]string{"MOAI_PROFILE_LEASE_TOKEN=token"}, 1, "parent", 2, "child"); err == nil {
		t.Fatal("unavailable lease store accepted")
	}
}

func TestCleanHomeSkipsLiveAndIndeterminateProfiles(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	profile := filepath.Join(home, "claude-profiles", "opus")
	old := filepath.Join(profile, "debug", "old.log")
	if err := os.MkdirAll(filepath.Dir(old), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := homestate.OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateProvisional(context.Background(), homestate.ProfileLease{ProfileName: "opus", ProfilePath: profile, PID: os.Getpid(), ProcessFingerprint: homestate.CurrentProcessFingerprint()})
	_ = store.Close()
	if err != nil {
		t.Fatal(err)
	}
	candidates := scanHomeCleanable(home, 1, 1, "v0", testOldTime())
	for _, c := range candidates {
		if c.AbsPath == old {
			t.Fatalf("leased profile candidate: %+v", c)
		}
	}
}

func testOldTime() time.Time { return time.Now().AddDate(1, 0, 0) }
