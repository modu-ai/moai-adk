package hook

// integration_lock_guard_test.go — the deny/allow boundary of the
// release-integration holder guard (card t194).
//
// The tests are written around one asymmetry: a deny needs positive evidence
// that a DIFFERENT, LIVE session holds the window, while every uncertainty
// allows. A guard that denied on uncertainty would wedge the very batch it
// protects, so the fail-open rows carry as much weight here as the deny row.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

func lockGuardInput(t *testing.T, sessionID, command string) *HookInput {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"command": command})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}
	return &HookInput{
		ToolName:  "Bash",
		SessionID: sessionID,
		ToolInput: raw,
	}
}

func seedLock(t *testing.T, root string, lock kanban.IntegrationLock) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	data, err := json.MarshalIndent(&lock, "", "  ")
	if err != nil {
		t.Fatalf("marshal lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, kanban.IntegrationLockFileName), data, 0o644); err != nil {
		t.Fatalf("write lock: %v", err)
	}
}

// A live foreign holder denies the merge, and the reason carries the sentinel
// plus enough identity for the operator to go ask that lane.
func TestCheckIntegrationLock_DeniesAForeignLiveHolder(t *testing.T) {
	root := t.TempDir()
	seedLock(t, root, kanban.IntegrationLock{
		SessionID:   "sess-holder",
		SessionName: "lane-5",
		PID:         os.Getpid(),
		Branch:      "release/v9.9.9",
		AcquiredAt:  "2026-01-01T00:00:00Z",
	})

	decision, reason := checkIntegrationLock(lockGuardInput(t, "sess-other", "git merge --no-ff WT-thing"), root)
	if decision != DecisionDeny {
		t.Fatalf("decision = %q, want deny", decision)
	}
	if !strings.HasPrefix(reason, integrationLockViolationPrefix) {
		t.Errorf("reason lacks the sentinel prefix: %s", reason)
	}
	if !strings.Contains(reason, "lane-5") {
		t.Errorf("reason does not name the holder: %s", reason)
	}
	if !strings.Contains(reason, "release/v9.9.9") {
		t.Errorf("reason does not name the branch: %s", reason)
	}
}

// The holder's own merge is not denied — the window is its to use.
func TestCheckIntegrationLock_AllowsTheHolder(t *testing.T) {
	root := t.TempDir()
	seedLock(t, root, kanban.IntegrationLock{SessionID: "sess-holder", PID: os.Getpid()})

	if decision, _ := checkIntegrationLock(lockGuardInput(t, "sess-holder", "git merge --no-ff WT-thing"), root); decision != "" {
		t.Errorf("holder was denied its own window: %q", decision)
	}
}

// Every fail-open path. Each row is a state where the guard cannot prove a
// foreign live holder, and each must allow rather than block the repository.
func TestCheckIntegrationLock_FailsOpen(t *testing.T) {
	t.Run("no record", func(t *testing.T) {
		root := t.TempDir()
		if d, _ := checkIntegrationLock(lockGuardInput(t, "sess-a", "git merge x"), root); d != "" {
			t.Errorf("denied with no lock record: %q", d)
		}
	})

	t.Run("unreadable record", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "state")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, kanban.IntegrationLockFileName), []byte("{nope"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		if d, _ := checkIntegrationLock(lockGuardInput(t, "sess-a", "git merge x"), root); d != "" {
			t.Errorf("denied on an unreadable record: %q", d)
		}
	})

	t.Run("caller has no session id", func(t *testing.T) {
		root := t.TempDir()
		seedLock(t, root, kanban.IntegrationLock{SessionID: "sess-holder", PID: os.Getpid()})
		if d, _ := checkIntegrationLock(lockGuardInput(t, "", "git merge x"), root); d != "" {
			t.Errorf("denied an unidentified caller: %q", d)
		}
	})

	t.Run("holder is gone", func(t *testing.T) {
		root := t.TempDir()
		dead := kanban.IntegrationLock{SessionID: "sess-dead", PID: 0x7FFFFFF0}
		if !dead.Stale() {
			t.Skip("seeded pid is live on this machine")
		}
		seedLock(t, root, dead)
		if d, _ := checkIntegrationLock(lockGuardInput(t, "sess-a", "git merge x"), root); d != "" {
			t.Errorf("denied on a dead holder: %q", d)
		}
	})

	t.Run("empty project root", func(t *testing.T) {
		if d, _ := checkIntegrationLock(lockGuardInput(t, "sess-a", "git merge x"), ""); d != "" {
			t.Errorf("denied with no project root: %q", d)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		if d, _ := checkIntegrationLock(nil, t.TempDir()); d != "" {
			t.Errorf("denied on nil input: %q", d)
		}
	})
}

// Only `git merge` is guarded. Widening the pattern would deny ordinary work
// in every worktree for as long as any lane held the window — the commands
// below must pass through even while a foreign holder is recorded.
func TestCheckIntegrationLock_OnlyGuardsMerge(t *testing.T) {
	root := t.TempDir()
	seedLock(t, root, kanban.IntegrationLock{SessionID: "sess-holder", PID: os.Getpid()})

	for _, command := range []string{
		"git status",
		"git commit -m x",
		"git push origin release/v9.9.9",
		"git log --oneline -1",
		"go test ./...",
		"echo 'git merge is mentioned in this string only'",
	} {
		if d, _ := checkIntegrationLock(lockGuardInput(t, "sess-other", command), root); d == DecisionDeny {
			t.Errorf("denied a non-merge command: %q", command)
		}
	}
}
