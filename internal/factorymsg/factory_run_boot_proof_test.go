package factorymsg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// bootProofSandbox scopes the factory state to a temp MOAI_HOME without
// touching HOME. The project root is a temp directory outside any git
// worktree, so CanonicalProjectRoot keeps it as-is.
func bootProofSandbox(t *testing.T) string {
	t.Helper()
	t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
	root := filepath.Join(t.TempDir(), "sandbox-project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatalf("create sandbox project root: %v", err)
	}
	if _, ok := homestate.SystemBootTime(); !ok {
		t.Skip("host does not report a boot time; the boot proof declines here by design")
	}
	return root
}

// seedIdentitylessRuns writes the card t1168 row shape — active, lead_pid 0,
// empty lead_process_start, no broker — dated long before any plausible boot
// of the test host.
func seedIdentitylessRuns(t *testing.T, root string, runIDs ...string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	const at = "2001-09-11T10:25:37.923062Z"
	for _, id := range runIDs {
		if _, err := db.DB.Exec(`INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,created_at,updated_at,lead_pid,lead_process_start) VALUES(?,'','glm','active','{}',?,?,0,'')`, id, at, at); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
		if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.started','{}',?)`, id, at); err != nil {
			t.Fatalf("seed event %s: %v", id, err)
		}
	}
}

var t1168RunIDs = []string{"glm", "tl4rkl", "tl8stp", "tl8v4m", "tl8xsp"}

// card t1168 — the operator's reproduction: five identity-less pre-boot runs
// plus the live lead the joining lane is looking for. Before the fix the join
// failed AMBIGUOUS_FACTORY naming all five as "owner indeterminate"; it now
// retires them on the boot proof and resolves to the live lead's run, which
// stays active.
func TestResolveActiveRunReapsPreBootIdentitylessRuns(t *testing.T) {
	root := bootProofSandbox(t)
	seedIdentitylessRuns(t, root, t1168RunIDs...)
	livePID, liveStart := liveIdentity(t)
	recordRun(t, root, "run-current", livePID, liveStart)

	got, err := ResolveActiveRun(context.Background(), root, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "run-current" {
		t.Fatalf("resolved %q, want run-current", got)
	}
	for _, id := range t1168RunIDs {
		if st := statusOf(t, root, id); st != "retired" {
			t.Fatalf("%s status = %q, want \"retired\"", id, st)
		}
	}
	if st := statusOf(t, root, "run-current"); st != "active" {
		t.Fatalf("live run status = %q, want \"active\"", st)
	}
}

// With no live run left, the reap ends in NO_ACTIVE_FACTORY — the fail-closed
// sentinel for "start a lead" — never in a selection among the dead rows.
func TestResolveActiveRunPreBootRowsAloneFailClosedAsNoActive(t *testing.T) {
	root := bootProofSandbox(t)
	seedIdentitylessRuns(t, root, t1168RunIDs...)

	id, err := ResolveActiveRun(context.Background(), root, "")
	if err == nil || !strings.Contains(err.Error(), "NO_ACTIVE_FACTORY") || id != "" {
		t.Fatalf("resolve = (%q, %v), want NO_ACTIVE_FACTORY", id, err)
	}
}

// A broker file for the run means a lead record may exist that the fallback
// failed to read; the proof declines and the run stays active and ambiguous.
func TestResolveActiveRunBootProofDeclinesWhenBrokerExists(t *testing.T) {
	root := bootProofSandbox(t)
	seedIdentitylessRuns(t, root, "legacy-a", "legacy-b")
	for _, id := range []string{"legacy-a", "legacy-b"} {
		s, err := Open(root, id)
		if err != nil {
			t.Fatalf("open broker %s: %v", id, err)
		}
		_ = s.Close()
	}

	_, err := ResolveActiveRun(context.Background(), root, "")
	if err == nil || !strings.Contains(err.Error(), "AMBIGUOUS_FACTORY") {
		t.Fatalf("err = %v, want AMBIGUOUS_FACTORY (both runs keep a broker)", err)
	}
	// Both halves of a declining leg (AC-018): the run stays active AND the
	// resolver reports it indeterminate, rendered as "<run> (owner <class>)".
	for _, id := range []string{"legacy-a", "legacy-b"} {
		if st := statusOf(t, root, id); st != "active" {
			t.Fatalf("%s status = %q, want \"active\"", id, st)
		}
		want := id + " (owner " + string(homestate.OwnerIndeterminate) + ")"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("err = %v, want it to classify %q", err, want)
		}
	}
}
