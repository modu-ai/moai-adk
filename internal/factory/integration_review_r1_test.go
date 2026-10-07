package factory

// integration_review_r1_test.go — card-review r1 repair tests (card t1479):
// one RED reproduction per finding before its fix. Findings P1-1..P1-5,
// P2-1..P2-9 from the codex card review at base 4964d0796.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// r1FixtureRepo is the fresh-per-case scratch repository.
type r1FixtureRepo struct {
	t    *testing.T
	dir  string
	root string
}

func newR1Repo(t *testing.T) *r1FixtureRepo {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	r := &r1FixtureRepo{t: t, dir: dir, root: root}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.email", "t@t.local")
	r.git("config", "user.name", "t")
	return r
}

func (r *r1FixtureRepo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (r *r1FixtureRepo) write(rel, content string) string {
	full := filepath.Join(r.dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
	return full
}

// TestR1_P1_1_QuotedPathCollisionDetected is the P1-1 RED reproduction: the
// candidate adds a KOREAN path (`런타임.local/payload`); the worktree holds an
// ignored regular file at `런타임.local`. With core.quotePath at its default,
// git prints the leaf as `"\355\225\234\...` — an UNQUOTING-illiterate check
// compares that printed form against the on-disk name and misses the
// collision entirely, letting the merge destroy the ignored bytes.
func TestR1_P1_1_QuotedPathCollisionDetected(t *testing.T) {
	r := newR1Repo(t)
	r.git("commit", "-q", "--allow-empty", "-m", "base")
	r.write(".gitignore", "런타임.local\n")
	r.git("add", ".gitignore")
	r.git("commit", "-q", "-m", "ignore")
	r.fromBranch("cand")
	r.write("런타임.local/payload", "p")
	r.git("add", "-f", "런타임.local/payload")
	r.git("commit", "-q", "-m", "cand adds korean path")
	// Back on main: the ignored regular file appears (the accepted fixture
	// order — the candidate commit ran while the path did not exist).
	r.git("checkout", "-q", "main")
	ignoredPath := r.write("런타임.local", "secretbytes")
	before := checksumOf(t, ignoredPath)

	got, err := FindAddedPathCollisions(r.dir, r.git("rev-parse", "main"), r.git("rev-parse", "cand"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsPath(got, "런타임.local") {
		t.Fatalf("RED (P1-1): the quoted Korean path's collision must be detected; got %v", got)
	}
	if after := checksumOf(t, ignoredPath); after != before {
		t.Fatalf("the ignored bytes must be untouched: before=%s after=%s", before, after)
	}
}

func checksumOf(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	cmd := exec.Command("shasum")
	cmd.Stdin = strings.NewReader(string(data))
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.Fields(string(out))[0]
}

func (r *r1FixtureRepo) fromBranch(branch string) { r.git("checkout", "-q", "-b", branch, "main") }

// TestR1_P1_2_HoldGatesTheWaitEnqueue is the P1-2 RED: EnqueueTicket runs
// its liveness refresh with a HARDCODED open policy, so once a ticket is
// queued, the NEXT enqueue's refresh PROMOTES it even while the policy is
// hold — hold suspends nothing on the wait path's own mutations.
func TestR1_P1_2_HoldGatesTheWaitEnqueue(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := WindowClock
	WindowClock = func() time.Time { return now }
	t.Cleanup(func() { WindowClock = prevClock })
	if err := WriteIntegrationWindowPolicy(root, IntegrationWindowPolicy{Policy: PolicyHold, Reason: "release-cut"}); err != nil {
		t.Fatal(err)
	}
	pid := os.Getpid()
	enqueue := func(session string) {
		t.Helper()
		if err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
			policy, policyErr := ReadIntegrationWindowPolicy(root)
			if policyErr != nil {
				return policyErr
			}
			return EnqueueTicket(w, IntegrationTicket{SessionID: session, OwnerPID: pid, WaiterPID: pid, WaiterStart: currentProcessFingerprint()},
				DefaultWindowProcProbe(), now, policy)
		}); err != nil {
			t.Fatal(err)
		}
	}
	enqueue("sess-b")
	enqueue("sess-c")
	lock, _ := ReadIntegrationLock(root)
	if lock.SessionID != "" {
		t.Fatalf("RED (P1-2): hold must suspend promotion on the enqueue path's refresh too; %q holds the window (policy hold)", lock.SessionID)
	}
	if len(lock.Queue) != 2 {
		t.Fatalf("both tickets must stay queued under hold: %+v", lock.Queue)
	}
}

// TestR1_P2_1_ReacquirePreservesQueue is the P2-1 RED: a holder re-acquiring
// (the refresh path) rewrites the record from its OWN want — the queue it
// did not carry — wiping every queued ticket.
func TestR1_P2_1_ReacquirePreservesQueue(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := WindowClock
	WindowClock = func() time.Time { return now }
	t.Cleanup(func() { WindowClock = prevClock })
	pid := os.Getpid()
	if _, err := AcquireIntegrationWindow(root, IntegrationLock{
		SessionID: "sess-a", PID: pid, PIDSource: PIDSourceSessionOwner, Branch: "develop",
	}, false, nil); err != nil {
		t.Fatal(err)
	}
	ticket := IntegrationTicket{SessionID: "sess-b", OwnerPID: pid, WaiterPID: pid, WaiterStart: currentProcessFingerprint()}
	if err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
		return EnqueueTicket(w, ticket, DefaultWindowProcProbe(), now, openPolicy())
	}); err != nil {
		t.Fatal(err)
	}
	// sess-a re-acquires (the holder refresh path).
	if _, err := AcquireIntegrationWindow(root, IntegrationLock{
		SessionID: "sess-a", PID: pid, PIDSource: PIDSourceSessionOwner, Branch: "develop",
	}, false, nil); err != nil {
		t.Fatal(err)
	}
	lock, _ := ReadIntegrationLock(root)
	if len(lock.Queue) != 1 || lock.Queue[0].SessionID != "sess-b" {
		t.Fatalf("RED (P2-1): the holder's re-acquire must preserve the queue, got %+v", lock.Queue)
	}
}

// TestR1_P2_8_ForceRecoversFromUnreadableRecord is the P2-8 RED: `--force`
// is the WEDGED-WINDOW recovery path, and an unreadable record is exactly
// when a leader reaches for it — but the force branch dereferences the nil
// current the failed read left behind and panics.
func TestR1_P2_8_ForceRecoversFromUnreadableRecord(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// An unreadable (corrupt) record — the wedge a force recovery targets.
	if err := os.WriteFile(filepath.Join(stateDir, IntegrationLockFileName), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RED (P2-8): force over an unreadable record panicked: %v", r)
		}
	}()
	_, err := AcquireIntegrationWindow(root, IntegrationLock{
		SessionID: "sess-recovery", PID: os.Getpid(), PIDSource: PIDSourceSessionOwner, Branch: "develop",
	}, true, nil)
	if err != nil {
		t.Fatalf("force must recover the window past an unreadable record: %v", err)
	}
	lock, _ := ReadIntegrationLock(root)
	if lock.SessionID != "sess-recovery" {
		t.Fatalf("the recovery force must take the window: %+v", lock)
	}
}

// TestR1_P2_9_StaleHolderRecordedBeforeClear is the P2-9 RED: clearing a
// stale holder with no queued successor (or under hold) erased the holder
// WITHOUT recording the displacement — today's takeover returned the
// replaced record to its caller, so the information existed; the queue
// refresh must keep it on the record.
func TestR1_P2_9_StaleHolderRecordedBeforeClear(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	lock := baseHolder(4001, now) // owner 4001 reads dead through the seam
	probe := WindowProcProbe{OwnerAlive: func(int) bool { return false }, WaiterAlive: func(int, string) bool { return true }}
	report := RefreshWindow(lock, openPolicy(), probe, now, IntegrationLeaseDefault)
	if report.Promoted != nil {
		t.Fatalf("no ticket, no promotion expected: %+v", report)
	}
	if lock.Displaced == nil || lock.Displaced.SessionID != "sess-a" {
		t.Fatalf("RED (P2-9): a cleared stale holder must be recorded as displaced: %+v", lock.Displaced)
	}
}

// TestR1_P2_5_TicketCarriesEnqueueIdentity is the P2-5 RED: a ticket
// enqueued through the wait verb's enqueue lands with EMPTY heartbeat and
// enqueue instants — the fields REQ-MWQ-001 says the ticket carries.
func TestR1_P2_5_TicketCarriesEnqueueIdentity(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	prevClock := WindowClock
	WindowClock = func() time.Time { return now }
	t.Cleanup(func() { WindowClock = prevClock })
	if err := UpdateIntegrationWindow(root, func(w *IntegrationLock) error {
		return EnqueueTicket(w, IntegrationTicket{SessionID: "sess-b", OwnerPID: os.Getpid(), WaiterPID: os.Getpid()},
			DefaultWindowProcProbe(), now, IntegrationWindowPolicy{Policy: PolicyOpen})
	}); err != nil {
		t.Fatal(err)
	}
	lock, _ := ReadIntegrationLock(root)
	if len(lock.Queue) != 1 {
		t.Fatalf("fixture: one ticket expected, got %+v", lock.Queue)
	}
	tk := lock.Queue[0]
	if tk.EnqueuedAt == "" || tk.Heartbeat == "" {
		t.Fatalf("RED (P2-5): the enqueued ticket carries empty identity instants: enqueued=%q heartbeat=%q", tk.EnqueuedAt, tk.Heartbeat)
	}
}
