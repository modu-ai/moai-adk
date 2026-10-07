package cli

// integration_review_r2_test.go — card-review ROUND 2 CLI-level RED
// reproductions (card t1479): classes B (wait-loop policy timing), D
// (primary-checkout target refusal), E (lease propagation at verb entries),
// G (status --json stdout purity + ticket fingerprint).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// TestR2_D_MergeRefusesPrimaryCheckoutTarget is the class D RED: the merge
// verb resolves the integration worktree; when the ONLY tree holding the
// integration branch is the parent checkout (never provisioned), the verb
// must refuse with the same not-provisioned standard factory complete
// applies — not proceed to merge into the parent.
func TestR2_D_MergeRefusesPrimaryCheckoutTarget(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	sdClearLaneEnv(t)
	sdLaneEnv(t, "lane-1", "")
	t.Setenv(config.EnvClaudeCodeSessionID, "sess-lane-1")
	// No integration worktree provisioned anywhere: the only tree holding
	// develop would be the parent checkout itself.
	cmd := newIntegrationMergeCmd()
	cmd.SetArgs([]string{"--card", "t1", "--run", fcRun})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("RED (class D): the merge must refuse when the integration branch is not provisioned in a dedicated worktree")
	}
	if !strings.Contains(err.Error(), "provision") {
		t.Fatalf("the refusal must name the provisioning standard: %v", err)
	}
}

// TestR2_E_StatusAndReleasePropagateLeaseConfig pins class E: every window
// verb entry (status, release) initializes the lease override from config —
// a lease_minutes: 0 config must govern the stamp the STATUS refresh and
// the RELEASE promotion write, not only acquire.
func TestR2_E_StatusAndReleasePropagateLeaseConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	// The REAL config surface: lease_minutes: 0 in the project's workflow
	// section is what the verbs read — write it, so the override the verbs
	// initialize reflects the disabled lease.
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"),
		[]byte("workflow:\n  integration_lock:\n    enabled: false\n    lease_minutes: 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	prev := factory.WindowLeaseDuration
	factory.WindowLeaseDuration = 0 // what lease_minutes: 0 resolves to
	t.Cleanup(func() { factory.WindowLeaseDuration = prev })

	// A holder + a queued successor; the status refresh (a mutation) and
	// the release (promotion) must stamp NOTHING with the lease disabled.
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := factory.WindowClock
	factory.WindowClock = func() time.Time { return now }
	t.Cleanup(func() { factory.WindowClock = prevClock })
	pid := os.Getpid()
	if _, err := factory.AcquireIntegrationWindow(root, factory.IntegrationLock{
		SessionID: "sess-a", PID: pid, PIDSource: factory.PIDSourceSessionOwner, Branch: "develop",
	}, false, &factory.AcquireWindowOptions{LeaseDuration: -1}); err != nil {
		t.Fatal(err)
	}
	if err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
		return factory.EnqueueTicket(w, factory.IntegrationTicket{SessionID: "sess-b", OwnerPID: pid, PIDSource: factory.PIDSourceSessionOwner,
			WaiterPID: pid, WaiterStart: cliProcessFingerprint()},
			factory.DefaultWindowProcProbe(), now, factory.IntegrationWindowPolicy{Policy: factory.PolicyOpen})
	}); err != nil {
		t.Fatal(err)
	}

	// The status verb entry: its refresh must not stamp a disabled lease.
	cmd := newIntegrationStatusCmd()
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}
	lock, _ := factory.ReadIntegrationLock(root)
	if lock.SessionID == "sess-a" && lock.LeaseExpiresAt != "" {
		t.Fatalf("RED (class E): the status refresh must honor the disabled lease, got stamp %q", lock.LeaseExpiresAt)
	}
	if _, err := factory.ReleaseIntegrationLock(root, "sess-a", 0, false); err != nil {
		t.Fatal(err)
	}
	lock, _ = factory.ReadIntegrationLock(root)
	if lock.SessionID != "sess-b" {
		t.Fatalf("fixture: the release must promote B, got %q", lock.SessionID)
	}
	if lock.LeaseExpiresAt != "" {
		t.Fatalf("RED (class E): the release-path promotion must honor the disabled lease, got stamp %q", lock.LeaseExpiresAt)
	}
}

// TestR2_G_StatusJSONStdoutPurity pins class G's stdout leg: with --json,
// the dropped-ticket lines and the queue rendering must NOT reach stdout —
// the JSON document is the only stdout content.
func TestR2_G_StatusJSONStdoutPurity(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := factory.WindowClock
	factory.WindowClock = func() time.Time { return now }
	t.Cleanup(func() { factory.WindowClock = prevClock })

	if err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
		*w = factory.IntegrationLock{
			SessionID: "sess-a", SessionName: "lane-a",
			PID: 4001, PIDSource: factory.PIDSourceSessionOwner,
			Branch: "develop", BranchSource: factory.BranchSourceConfig,
			AcquiredAt: now.Format(time.RFC3339),
		}
		w.Queue = append(w.Queue, factory.IntegrationTicket{
			SessionID: "sess-dead", OwnerPID: 1 << 20, PIDSource: factory.PIDSourceSessionOwner,
			WaiterPID: 1 << 21, // dead pids: the drop rules name this ticket
			Heartbeat: now.Format(time.RFC3339), EnqueuedAt: now.Format(time.RFC3339),
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := runIntegrationStatusJSON(t)
	if strings.Contains(stdout, "dropped ticket") || strings.Contains(stdout, "queue:") {
		t.Fatalf("RED (class G): human lines must not pollute --json stdout: %q", stdout)
	}
	var doc map[string]any
	if err := json.NewDecoder(strings.NewReader(stdout)).Decode(&doc); err != nil {
		t.Fatalf("RED (class G): --json stdout is not one parseable document: %v\nstdout=%q", err, stdout)
	}
	// The drop must still be NAMED somewhere (REQ-MWQ-003): stderr is the
	// human channel on the JSON path (the release-verb precedent).
	if !strings.Contains(stderr, "dropped ticket") {
		t.Fatalf("the dropped ticket must still be named (stderr), got stderr=%q", stderr)
	}
}

// runIntegrationStatusJSON runs `moai integration status --json` in-process
// and returns the raw stdout/stderr streams.
func runIntegrationStatusJSON(t *testing.T) (string, string) {
	t.Helper()
	cmd := newIntegrationStatusCmd()
	var out, errBuf strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status --json: %v", err)
	}
	return out.String(), errBuf.String()
}

// TestR2_G_TicketCarriesWaiterFingerprint pins class G's fingerprint leg:
// the wait verb's enqueued ticket records the waiter's process fingerprint
// (id AND start), not an empty start — an empty start makes every later
// live process with that pid count as the waiter.
func TestR2_G_TicketCarriesWaiterFingerprint(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := factory.WindowClock
	factory.WindowClock = func() time.Time { return now }
	t.Cleanup(func() { factory.WindowClock = prevClock })

	// The wait verb builds its ticket WITHOUT a start instant — reproduce
	// its construction exactly (owner pid + waiter pid only) and run the
	// enqueue the verb runs.
	ticket := factory.IntegrationTicket{SessionID: "sess-b", OwnerPID: os.Getpid(), WaiterPID: os.Getpid()}
	if err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
		policy, policyErr := factory.ReadIntegrationWindowPolicy(root)
		if policyErr != nil {
			return policyErr
		}
		return factory.EnqueueTicket(w, ticket, factory.DefaultWindowProcProbe(), now, policy)
	}); err != nil {
		t.Fatal(err)
	}
	lock, _ := factory.ReadIntegrationLock(root)
	if len(lock.Queue) != 1 {
		t.Fatalf("fixture: one ticket expected")
	}
	if lock.Queue[0].WaiterStart == "" {
		t.Fatalf("RED (class G): the enqueued ticket must carry the waiter's process start fingerprint (id AND start), got empty")
	}
}
