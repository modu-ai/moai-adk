package cli

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestFactoryMCPIdentityAttribution(t *testing.T) {
	root := t.TempDir()
	s, err := factorymsg.Open(root, "run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	pid := os.Getpid()
	fingerprint := "test-process-start"
	oldProbe := factoryProbeProcessIdentity
	factoryProbeProcessIdentity = func(gotPID int) (string, homestate.ProcessIdentityState) {
		if gotPID == pid {
			return fingerprint, homestate.ProcessIdentityLive
		}
		return "", homestate.ProcessIdentityDead
	}
	t.Cleanup(func() { factoryProbeProcessIdentity = oldProbe })
	peer, err := s.RegisterPeer(context.Background(), factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "run", Backend: "codex", Role: "worker", Slot: "agent-1", SessionUUID: "owner-session", Generation: 1, PID: pid, ProcessStart: fingerprint})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(pid))
	t.Setenv(config.EnvClaudeCodeSessionID, "foreign-authoritative-session")
	if got, err := currentFactoryPeer(context.Background(), s); err == nil {
		t.Fatalf("authoritative lookup failure fell back to PID peer: %+v", got)
	}
	t.Setenv(config.EnvClaudeCodeSessionID, "")
	t.Setenv(config.EnvClaudeProjectDir, t.TempDir()) // no shared side-channel attribution
	got, err := currentFactoryPeer(context.Background(), s)
	if err != nil || got.SessionUUID != peer.SessionUUID || got.Generation != peer.Generation {
		t.Fatalf("owner attribution=%+v err=%v", got, err)
	}
}

func TestFactoryMCPIdentityFallsBackToDirectParent(t *testing.T) {
	root := t.TempDir()
	s, err := factorymsg.Open(root, "run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	pid := os.Getppid()
	fingerprint := "parent-process-start"
	oldProbe := factoryProbeProcessIdentity
	factoryProbeProcessIdentity = func(gotPID int) (string, homestate.ProcessIdentityState) {
		if gotPID == pid {
			return fingerprint, homestate.ProcessIdentityLive
		}
		return "", homestate.ProcessIdentityDead
	}
	t.Cleanup(func() { factoryProbeProcessIdentity = oldProbe })
	peer, err := s.RegisterPeer(context.Background(), factorymsg.Peer{ProjectKey: homestate.ProjectKey(root), RunID: "run", Backend: "codex", Role: "worker", Slot: "agent-1", SessionUUID: "parent-owned-session", Generation: 1, PID: pid, ProcessStart: fingerprint})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvMoaiSessionPID, "")
	t.Setenv(config.EnvClaudeCodeSessionID, "")
	t.Setenv(config.EnvClaudeProjectDir, t.TempDir())
	got, err := currentFactoryPeer(context.Background(), s)
	if err != nil || got.SessionUUID != peer.SessionUUID {
		t.Fatalf("parent attribution=%+v err=%v", got, err)
	}
}
