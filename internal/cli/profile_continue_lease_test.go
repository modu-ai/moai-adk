//go:build !windows

package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestContinueLaunchCreatesSingleProvisionalLeaseBeforeChildStart(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(t.TempDir(), "home")
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	observed := filepath.Join(t.TempDir(), "observed")
	script := "#!/bin/sh\ntest -n \"$MOAI_PROFILE_LEASE_TOKEN\" || exit 91\ntest -f \"$MOAI_HOME/run/profile-leases.db\" || exit 92\nprintf '%s' \"$MOAI_PROFILE_LEASE_TOKEN\" > \"$OBSERVED\"\nexit 0\n"
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", home)
	t.Setenv("OBSERVED", observed)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Chdir(root)
	if err := launchClaudeDefault("lease-test", []string{"--continue"}); err != nil {
		t.Fatal(err)
	}
	tokenRaw, err := os.ReadFile(observed)
	if err != nil {
		t.Fatal(err)
	}
	store, err := homestate.OpenProfileLeases()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var count int
	if err := store.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM profile_leases WHERE token=? AND state='provisional'`, strings.TrimSpace(string(tokenRaw))).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("provisional rows=%d", count)
	}
}
