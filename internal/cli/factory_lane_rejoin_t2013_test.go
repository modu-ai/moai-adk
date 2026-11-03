// factory_lane_rejoin_t2013_test.go — card t2013: `moai cc -l --name lane-<n>`
// is a REJOIN of that canonical lane label. It is granted only when no live
// roster row holds the label; a live label is refused and nothing changes; a
// dead label is reused by the claim (one row, this process's pid); the launcher
// never touches a lease, so the owner's `factory next --card` re-lease path
// (and nobody else's) is what the rejoined label unlocks.
//
// Every fixture is a real temp factory record under t.TempDir(); liveness is the
// factoryProcessAlive seam (this process and pid 1 are live, rejoinDeadPID is
// dead) and is installed only around the rejoin call, so the `next` verbs run
// under the real probe exactly as the sibling expired-owner tests do.
package cli

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// rejoinDeadPID is a pid the stubbed probe reports dead.
const rejoinDeadPID = 99991

// rejoinFixture is the expired-lease card t1 owned by lane-1 (reacqExpired),
// with lane-1's roster row made dead and lane-2's left live (pid 1).
func rejoinFixture(t *testing.T) reacqFixture {
	t.Helper()
	t.Setenv(config.EnvFactoryRunID, "")
	fx := reacqExpired(t)
	rejoinSetPID(t, fx.root, "lane-1", rejoinDeadPID)
	return fx
}

func rejoinSetPID(t *testing.T, root, label string, pid int) {
	t.Helper()
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE workers SET pid=? WHERE label=?`, pid, label); err != nil {
		t.Fatalf("set pid of %s: %v", label, err)
	}
	_ = db.Close()
}

// rejoinRoster renders every roster row as "label=pid" in label order.
func rejoinRoster(t *testing.T, root string) string {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT label,pid FROM workers ORDER BY label`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var parts []string
	for rows.Next() {
		var label string
		var pid int
		if err := rows.Scan(&label, &pid); err != nil {
			t.Fatal(err)
		}
		parts = append(parts, label+"="+strconv.Itoa(pid))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return strings.Join(parts, ",")
}

// rejoinCall runs rejoinFactoryLaneName for root under the stubbed probe and
// restores the probe before returning.
func rejoinCall(t *testing.T, root, label string) (string, error) {
	t.Helper()
	prev := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid == os.Getpid() || pid == 1 }
	defer func() { factoryProcessAlive = prev }()
	return rejoinFactoryLaneName(root, label, factory.BackendClaude)
}

// (parse) `-l --name lane-<n>` selects the rejoin; other names keep the
// conflict; the plain `-l` join is unchanged.
func TestT2013RejoinEntryParse(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())

	for _, args := range [][]string{
		{"-a", "--name", "lane-1"},
		{"-a", "-n", "lane-2"},
		{"-a", "--name=lane-3"},
		{"-a", "-n=lane-4", "-b"},
	} {
		p, err := parseLauncherEntry(args)
		if err != nil {
			t.Errorf("parseLauncherEntry(%v) = %v, want the rejoin accepted", args, err)
			continue
		}
		if !p.FactoryEnabled || !p.FactoryRejoin || p.FactoryAutoNumber {
			t.Errorf("parseLauncherEntry(%v) = enabled=%v rejoin=%v auto=%v, want enabled+rejoin and no auto number", args, p.FactoryEnabled, p.FactoryRejoin, p.FactoryAutoNumber)
		}
		if _, isLane := parseFactoryLaneLabel(p.Rest); !isLane {
			t.Errorf("parseLauncherEntry(%v) rest = %v, want the lane name kept for the lane branch", args, p.Rest)
		}
	}

	for _, args := range [][]string{
		{"-a", "--name", "board-watch"},
		{"-a", "--name", "lane-01"},
		{"-a", "--name", "lane-0"},
		{"-a", "--name", "lane-x"},
		{"-a", "--name", "worker-3"},
	} {
		if _, err := parseLauncherEntry(args); err == nil {
			t.Errorf("parseLauncherEntry(%v) accepted, want a refusal (not a canonical lane label)", args)
		}
	}

	p, err := parseLauncherEntry([]string{"-a"})
	if err != nil || p.FactoryRejoin || !p.FactoryAutoNumber {
		t.Errorf("parseLauncherEntry(-l) = rejoin=%v auto=%v err=%v, want the unchanged next-free join", p.FactoryRejoin, p.FactoryAutoNumber, err)
	}
}

// (i) A dead label is rejoined: one live row for it, this process's pid.
func TestT2013RejoinDeadLabelAccepted(t *testing.T) {
	fx := rejoinFixture(t)
	if got, want := rejoinRoster(t, fx.root), "lane-1=99991,lane-2=1"; got != want {
		t.Fatalf("setup roster = %q, want %q", got, want)
	}

	final, err := rejoinCall(t, fx.root, "lane-1")
	if err != nil {
		t.Fatalf("rejoin of the dead lane-1: %v", err)
	}
	if final != "lane-1" {
		t.Errorf("rejoined label = %q, want lane-1", final)
	}
	if got, want := rejoinRoster(t, fx.root), "lane-1="+strconv.Itoa(os.Getpid())+",lane-2=1"; got != want {
		t.Errorf("roster after rejoin = %q, want %q (one live row for lane-1, the other lane untouched)", got, want)
	}
}

// (j) A live label is refused and roster and leases are unchanged.
func TestT2013RejoinLiveLabelRefusedNothingChanges(t *testing.T) {
	fx := rejoinFixture(t)
	rejoinSetPID(t, fx.root, "lane-1", 1) // a live holder
	rosterBefore := rejoinRoster(t, fx.root)
	rowsBefore := fcRows(t, fx.root)

	final, err := rejoinCall(t, fx.root, "lane-1")
	if err == nil {
		t.Fatalf("rejoin of the live lane-1 = %q, want a refusal", final)
	}
	if !strings.Contains(err.Error(), "lane-1") || !strings.Contains(err.Error(), "occupied") {
		t.Errorf("error = %q, want it to name lane-1 and that it is occupied", err)
	}
	if got := rejoinRoster(t, fx.root); got != rosterBefore {
		t.Errorf("roster changed by a refused rejoin: %q → %q", rosterBefore, got)
	}
	if got := fcRows(t, fx.root); got != rowsBefore {
		t.Errorf("cards or events changed by a refused rejoin:\nbefore:\n%s\nafter:\n%s", rowsBefore, got)
	}
}

// (k) The same live process asking twice gets one row: the second call is a
// refusal, never a duplicate (the label is also the roster's primary key).
func TestT2013RejoinRepeatByLiveProcessNoDuplicate(t *testing.T) {
	fx := rejoinFixture(t)
	if _, err := rejoinCall(t, fx.root, "lane-1"); err != nil {
		t.Fatalf("first rejoin: %v", err)
	}
	after := rejoinRoster(t, fx.root)

	if final, err := rejoinCall(t, fx.root, "lane-1"); err == nil {
		t.Fatalf("second rejoin = %q, want a refusal (this process already holds the label)", final)
	}
	if got := rejoinRoster(t, fx.root); got != after {
		t.Errorf("roster after the repeat = %q, want it unchanged %q", got, after)
	}
	if n := strings.Count(rejoinRoster(t, fx.root), "lane-1="); n != 1 {
		t.Errorf("lane-1 rows = %d, want exactly 1", n)
	}
}

// Two simultaneous rejoin attempts for one dead label: exactly one wins. The
// claim serializes on the admission lock and one transaction
// (factory_slots.go claimFactoryLane), so the loser sees the winner's live row.
func TestT2013RejoinConcurrentAttemptsOneWinner(t *testing.T) {
	fx := rejoinFixture(t)
	prev := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid == os.Getpid() || pid == 1 }
	defer func() { factoryProcessAlive = prev }()

	const attempts = 4
	var wg sync.WaitGroup
	results := make(chan error, attempts)
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := rejoinFactoryLaneName(fx.root, "lane-1", factory.BackendClaude)
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	wins := 0
	for err := range results {
		if err == nil {
			wins++
		}
	}
	if wins != 1 {
		t.Errorf("successful concurrent rejoins = %d, want exactly 1", wins)
	}
	if n := strings.Count(rejoinRoster(t, fx.root), "lane-1="); n != 1 {
		t.Errorf("lane-1 rows = %d after the race, want exactly 1", n)
	}
}

// claimFactoryLaneForEntry routes the rejoin entry to the exact-label claim and
// every other entry to the next-free claim.
func TestT2013ClaimForEntryRoutesRejoin(t *testing.T) {
	fx := rejoinFixture(t)
	prev := factoryProcessAlive
	factoryProcessAlive = func(pid int) bool { return pid == os.Getpid() || pid == 1 }
	defer func() { factoryProcessAlive = prev }()

	final, err := claimFactoryLaneForEntry(fx.root, "lane-1", factory.BackendClaude, launcherEntryParse{FactoryEnabled: true, FactoryRejoin: true}, nil)
	if err != nil || final != "lane-1" {
		t.Fatalf("rejoin entry = (%q, %v), want (lane-1, nil)", final, err)
	}
}

// (l) After the rejoin another lane's `next --card` on the expired lease of
// lane-1's card is still refused, and nothing is written: the launcher moved
// no lease.
func TestT2013OtherLaneStillRefusedAfterRejoin(t *testing.T) {
	fx := rejoinFixture(t)
	if _, err := rejoinCall(t, fx.root, "lane-1"); err != nil {
		t.Fatalf("rejoin: %v", err)
	}
	cardBefore := fcCard(t, fx.root, "t1")

	nmLaneEnv(t, "lane-2", "")
	snapshot := nmSnapshot(t, fx.root, fx.store)
	rows := fcRows(t, fx.root)
	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, out, stderr, err, "owned")
	if after := nmSnapshot(t, fx.root, fx.store); after != snapshot {
		t.Errorf("a refused nomination changed the queue or the record:\nbefore:\n%s\nafter:\n%s", snapshot, after)
	}
	if after := fcRows(t, fx.root); after != rows {
		t.Errorf("a refused nomination changed rows or events:\nbefore:\n%s\nafter:\n%s", rows, after)
	}
	if c := fcCard(t, fx.root, "t1"); c.LeaseHolder != cardBefore.LeaseHolder || c.OwnerLabel != cardBefore.OwnerLabel || c.Version != cardBefore.Version {
		t.Errorf("the rejoin moved the lease: holder %q→%q owner %q→%q version %d→%d",
			cardBefore.LeaseHolder, c.LeaseHolder, cardBefore.OwnerLabel, c.OwnerLabel, cardBefore.Version, c.Version)
	}
}

// (m) Rejoin, then the owner's `next --card` re-leases the expired lease: the
// end-to-end path the rejoin exists to open.
func TestT2013RejoinThenOwnerNextReleases(t *testing.T) {
	fx := rejoinFixture(t)
	before := fcCard(t, fx.root, "t1")
	if _, err := rejoinCall(t, fx.root, "lane-1"); err != nil {
		t.Fatalf("rejoin: %v", err)
	}

	nmLaneEnv(t, "lane-1", "")
	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err != nil {
		t.Fatalf("owner next --card t1 after the rejoin: %v (stderr %q), want exit 0", err, stderr)
	}
	if nmLeasedHead(out) == "" {
		t.Errorf("stdout = %q, want the leased card's head line", out)
	}
	reacqAssertReacquired(t, fx, before)
}
