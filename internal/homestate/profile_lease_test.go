package homestate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestProfileLeasesAreGlobalAndPrivate(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	store, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err = store.CreateProvisional(context.Background(), ProfileLease{ProfileName: "opus", ProfilePath: filepath.Join(home, "claude-profiles", "opus"), ProjectKey: "p1", PID: os.Getpid(), ProcessFingerprint: "start"}); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "run", "profile-leases.db")
	if store.Path != want {
		t.Fatalf("path=%q want=%q", store.Path, want)
	}
	if info, err := os.Stat(want); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%v err=%v", info.Mode().Perm(), err)
	}
}

func TestProfileLeaseReconcilePIDFingerprint(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	store, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	_, _ = store.CreateProvisional(ctx, ProfileLease{ProfileName: "live", ProfilePath: "/p/live", PID: 10, ProcessFingerprint: "same"})
	staleToken, _ := store.CreateProvisional(ctx, ProfileLease{ProfileName: "stale", ProfilePath: "/p/stale", PID: 11, ProcessFingerprint: "old"})
	if _, err := store.DB.ExecContext(ctx, `UPDATE profile_leases SET transfer_deadline='2000-01-01T00:00:00Z' WHERE token=?`, staleToken); err != nil {
		t.Fatal(err)
	}
	_, _ = store.CreateProvisional(ctx, ProfileLease{ProfileName: "unknown", ProfilePath: "/p/unknown", PID: 12, ProcessFingerprint: "x"})
	probe := func(pid int) (string, ProcessIdentityState) {
		switch pid {
		case 10:
			return "same", ProcessIdentityLive
		case 11:
			return "new", ProcessIdentityLive
		default:
			return "", ProcessIdentityIndeterminate
		}
	}
	if got, _ := store.ProfileProtection(ctx, "/p/live", probe); got != LeaseLive {
		t.Fatalf("live=%s", got)
	}
	if got, _ := store.ProfileProtection(ctx, "/p/stale", probe); got != LeaseStale {
		t.Fatalf("stale=%s", got)
	}
	if got, _ := store.ProfileProtection(ctx, "/p/unknown", probe); got != LeaseIndeterminate {
		t.Fatalf("unknown=%s", got)
	}
}

func TestProfileLeaseCASAndProtectionBranches(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	store, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	path := filepath.Join(home, "profiles", "opus")
	token, err := store.CreateProvisional(ctx, ProfileLease{Token: "token", ProfileName: "opus", ProfilePath: path, PID: 10, ProcessFingerprint: "parent"})
	if err != nil || token != "token" {
		t.Fatalf("token=%q err=%v", token, err)
	}
	if err := store.TransferToChild(ctx, token, 99, "wrong", 20, "child"); err == nil {
		t.Fatal("stale parent CAS accepted")
	}
	if err := store.TransferToChild(ctx, token, 10, "parent", 20, "child"); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.ProfileProtection(ctx, path, func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }); got != LeaseIndeterminate {
		t.Fatalf("transferring protection=%s", got)
	}
	if err := store.Enrich(ctx, token, "session", 20, "child"); err != nil {
		t.Fatal(err)
	}
	if err := store.Enrich(ctx, "missing", "session", 20, "child"); err == nil {
		t.Fatal("missing token enrich accepted")
	}
	if got, _ := store.ProfileProtection(ctx, filepath.Join(home, "profiles", "unused"), func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }); got != LeaseNone {
		t.Fatalf("unused protection=%s", got)
	}
	emptyPath := filepath.Join(home, "profiles", "empty-fingerprint")
	if _, err := store.CreateProvisional(ctx, ProfileLease{Token: "empty", ProfileName: "empty", ProfilePath: emptyPath, PID: 30}); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.ProfileProtection(ctx, emptyPath, func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }); got != LeaseIndeterminate {
		t.Fatalf("empty fingerprint protection=%s", got)
	}
	if err := store.ReleaseSession(ctx, "session"); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.ProfileProtection(ctx, path, func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }); got != LeaseNone {
		t.Fatalf("released protection=%s", got)
	}
}

func TestDeadProvisionalLeaseRemainsIndeterminateUntilTransferDeadline(t *testing.T) {
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
	store, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "profile")
	token, err := store.CreateProvisional(ctx, ProfileLease{ProfilePath: path, PID: 10, ProcessFingerprint: "parent"})
	if err != nil {
		t.Fatal(err)
	}
	dead := func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }
	if got, _ := store.ProfileProtection(ctx, path, dead); got != LeaseIndeterminate {
		t.Fatalf("fresh dead provisional=%s", got)
	}
	if err := store.TransferToChild(ctx, token, 10, "parent", 20, "child"); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.ProfileProtection(ctx, path, dead); got != LeaseIndeterminate {
		t.Fatalf("transferring=%s", got)
	}
	if _, err := store.DB.Exec(`UPDATE profile_leases SET transfer_deadline='2000-01-01T00:00:00Z' WHERE token=?`, token); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.ProfileProtection(ctx, path, dead); got != LeaseStale {
		t.Fatalf("expired transfer=%s", got)
	}
}

func TestProfileProtectionAtHomeWithoutLeaseDB(t *testing.T) {
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
	got, err := ProfileProtectionAtHome(filepath.Join(t.TempDir(), "profile"))
	if err != nil || got != LeaseNone {
		t.Fatalf("protection=%s err=%v", got, err)
	}
}

func TestProfileProtectionAtHomeReadsLiveLease(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("MOAI_HOME", home)
	profile := filepath.Join(home, "profiles", "live")
	store, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	fingerprint := CurrentProcessFingerprint()
	if fingerprint == "" {
		t.Fatal("current process fingerprint unavailable")
	}
	if _, err := store.CreateProvisional(context.Background(), ProfileLease{Token: "live-at-home", ProfilePath: profile, PID: os.Getpid(), ProcessFingerprint: fingerprint}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := ProfileProtectionAtHome(profile)
	if err != nil || got != LeaseLive {
		t.Fatalf("protection=%s err=%v", got, err)
	}
}

func TestProfileLeaseStoreFailsClosedOnUnavailableStorage(t *testing.T) {
	blockedHome := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockedHome, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", blockedHome)
	if _, err := OpenProfileLeases(); err == nil {
		t.Fatal("unavailable lease home accepted")
	}
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
	store, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := store.CreateProvisional(ctx, ProfileLease{ProfilePath: "/closed"}); err == nil {
		t.Fatal("closed create accepted")
	}
	if err := store.TransferToChild(ctx, "x", 1, "p", 2, "c"); err == nil {
		t.Fatal("closed transfer accepted")
	}
	if err := store.Enrich(ctx, "x", "s", 2, "c"); err == nil {
		t.Fatal("closed enrich accepted")
	}
	if err := store.ReleaseSession(ctx, "s"); err == nil {
		t.Fatal("closed release accepted")
	}
	if _, err := store.ProfileProtection(ctx, "/closed", func(int) (string, ProcessIdentityState) { return "", ProcessIdentityDead }); err == nil {
		t.Fatal("closed protection accepted")
	}
	brokenHome := filepath.Join(t.TempDir(), "broken-home")
	if err := os.MkdirAll(filepath.Join(brokenHome, "run"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brokenHome, "run", "profile-leases.db"), []byte("not sqlite"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", brokenHome)
	if _, err := OpenProfileLeases(); err == nil {
		t.Fatal("corrupt lease database accepted")
	}
	if got, err := ProfileProtectionAtHome("/profile"); err == nil || got != LeaseIndeterminate {
		t.Fatalf("corrupt at-home protection=%s err=%v", got, err)
	}
}

func TestProfileLeaseUnknownProbeStateFailsClosed(t *testing.T) {
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), "home"))
	store, err := OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	path := filepath.Join(t.TempDir(), "profile")
	if _, err := store.CreateProvisional(context.Background(), ProfileLease{Token: "unknown-state", ProfilePath: path, PID: 42, ProcessFingerprint: "known"}); err != nil {
		t.Fatal(err)
	}
	got, err := store.ProfileProtection(context.Background(), path, func(int) (string, ProcessIdentityState) {
		return "", ProcessIdentityState("unexpected")
	})
	if err != nil || got != LeaseIndeterminate {
		t.Fatalf("protection=%s err=%v", got, err)
	}
}
