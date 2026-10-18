package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card t1688: `moai factory complete` with no card id must be a usage error,
// never an index-out-of-range panic. The verb used to accept zero positionals
// (Args=MaximumNArgs(2)) while the body read args[0] unchecked; a lane session
// reaches the body, so the panic needs the lane admission env below. The test
// touches no factory store: the panic site precedes every store call, and HOME
// and MOAI_HOME point at throwaway directories regardless.
func TestFactoryCompleteWithoutArgsIsAUsageError(t *testing.T) {
	isolateGitGlobalSeams(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv(config.EnvFactoryRole, config.FactoryRoleLane)
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("factory complete with no card id panicked: %v", r)
			}
		}()
		_, _, err = runFactory(t, "complete")
	}()
	if err == nil {
		t.Fatal("factory complete with no card id returned nil, want a usage error")
	}
	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("error %q does not read as an argument-count error", err)
	}
}

// Card t1688 item 1: the renew verb was once defined but never registered
// (SPEC-FACTORY-LEASE-RENEW-001 M2, commit 4750b8b57, registered it). This
// locks the registration so a future edit of the verb list cannot drop it.
func TestFactoryRenewVerbIsRegistered(t *testing.T) {
	isolateGitGlobalSeams(t)
	cmd, _, err := newFactoryCommand().Find([]string{"renew"})
	if err != nil {
		t.Fatalf("factory renew lookup: %v", err)
	}
	if cmd == nil || cmd.Name() != "renew" {
		t.Fatalf("factory renew is not registered on the factory command: got %v", cmd)
	}
}
