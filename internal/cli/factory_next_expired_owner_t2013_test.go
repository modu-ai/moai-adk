// factory_next_expired_owner_t2013_test.go — card t2013: the registered owner
// of a card that sits `leased` with an EXPIRED lease re-leases it with
// `moai factory next --card <id>` (and the factory_next MCP tool, which shares
// the same function): the expiry is applied and the lease is taken again in
// one transaction, the card keeps its stage, worktree and evidence, and the
// verb never touches the worktree. Everything else keeps its behavior: `renew`
// on an expired lease, another lane and the owner before expiry keep refusing,
// and `handoff adopt` stays out of the re-lease.
//
// Every fixture is a real temp factory record, a real temp queue and a real
// temp git repository standing in for the card worktree; time is the
// factoryCardNow seam (fcNow), never a sleep.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	reacqEvidenceSHA  = "0123456789abcdef0123456789abcdef01234567"
	reacqEvidencePath = ".moai/reports/t1/evidence.md"
)

// reacqFixture is one expired-lease card: queue item t1 picked, record row t1
// `leased` to lane-1 (also its owner) at version 3 with the lease expiry given,
// stage run, evidence recorded, and a real git repository as its worktree
// holding one modified tracked file and one untracked file.
type reacqFixture struct {
	root  string
	store *factory.BacklogStore
	wt    string
}

func reacqWorktree(t *testing.T) string {
	t.Helper()
	wt := t.TempDir()
	fcGit(t, wt, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(wt, "tracked.txt"), []byte("committed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, wt, "add", "tracked.txt")
	fcGit(t, wt, "commit", "-q", "-m", "base")
	if err := os.WriteFile(filepath.Join(wt, "tracked.txt"), []byte("edited, not committed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, "untracked.txt"), []byte("never added\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	for _, name := range []string{"tracked.txt", "untracked.txt"} {
		if err := os.Chtimes(filepath.Join(wt, name), old, old); err != nil {
			t.Fatal(err)
		}
	}
	return wt
}

func reacqSetup(t *testing.T, expiry time.Time) reacqFixture {
	t.Helper()
	root, store := nmBase(t, factory.BacklogStatePicked)
	wt := reacqWorktree(t)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		Stage: homestate.CardRun, Version: 3,
		LeaseExpiresAt: expiry.Format(time.RFC3339Nano), HeartbeatAt: expiry.Add(-15 * time.Minute).Format(time.RFC3339Nano),
		WorktreePath: wt, EvidenceSHA: reacqEvidenceSHA, EvidencePath: reacqEvidencePath,
	})
	return reacqFixture{root: root, store: store, wt: wt}
}

// reacqExpired is the fixture whose lease lapsed ten minutes before fcNow.
func reacqExpired(t *testing.T) reacqFixture { return reacqSetup(t, fcNow.Add(-10*time.Minute)) }

// reacqWorktreeSnapshot is what "the worktree is byte-identical" compares: the
// content hash and the modification time of every file, plus git's own status.
func reacqWorktreeSnapshot(t *testing.T, wt string) string {
	t.Helper()
	var b strings.Builder
	for _, name := range []string{"tracked.txt", "untracked.txt"} {
		path := filepath.Join(wt, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		fmt.Fprintf(&b, "%s sha=%s mtime=%d\n", name, fcFileHash(t, path), info.ModTime().UnixNano())
	}
	b.WriteString(fcGit(t, wt, "status", "--porcelain=v1"))
	b.WriteString("\n")
	b.WriteString(fcGit(t, wt, "rev-parse", "HEAD"))
	return b.String()
}

// reacqEventKinds lists the event kinds of the factory record in append order.
func reacqEventKinds(t *testing.T, root string) []string {
	t.Helper()
	db := fcOpen(t, root)
	rows, err := db.DB.Query(`SELECT kind FROM events ORDER BY seq`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var kinds []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, k)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return kinds
}

func reacqIndexOf(kinds []string, kind string) (first, count int) {
	first = -1
	for i, k := range kinds {
		if k == kind {
			if first < 0 {
				first = i
			}
			count++
		}
	}
	return first, count
}

// reacqOutcome renders the re-leased card without its worktree path, so the
// verb run and the tool run (separate roots) can be compared as equal strings.
func reacqOutcome(t *testing.T, root string) string {
	t.Helper()
	c := fcCard(t, root, "t1")
	return fmt.Sprintf("state=%s owner=%s holder=%s version=%d stage=%s evsha=%s evpath=%s expires=%s heartbeat=%s events=%s",
		c.State, c.OwnerLabel, c.LeaseHolder, c.Version, c.Stage, c.EvidenceSHA, c.EvidencePath,
		c.LeaseExpiresAt, c.HeartbeatAt, strings.Join(reacqEventKinds(t, root), ","))
}

// reacqAssertReacquired asserts the post-conditions of a successful owner
// re-lease against the card as it stood before the call.
func reacqAssertReacquired(t *testing.T, fx reacqFixture, before homestate.Card) {
	t.Helper()
	after := fcCard(t, fx.root, "t1")
	if after.State != homestate.CardLeased || after.LeaseHolder != "lane-1" || after.OwnerLabel != "lane-1" {
		t.Fatalf("t1 = %s owner=%q holder=%q, want leased to lane-1 (owner and holder)", after.State, after.OwnerLabel, after.LeaseHolder)
	}
	if after.Version <= before.Version {
		t.Errorf("version = %d, want it incremented past %d", after.Version, before.Version)
	}
	expires, err := time.Parse(time.RFC3339Nano, after.LeaseExpiresAt)
	if err != nil || !expires.After(fcNow) {
		t.Errorf("lease expiry = %q (%v), want a NEW expiry after the clock %s", after.LeaseExpiresAt, err, fcNow.Format(time.RFC3339Nano))
	}
	if after.LeaseExpiresAt == before.LeaseExpiresAt {
		t.Errorf("lease expiry %q is the lapsed one", after.LeaseExpiresAt)
	}
	if after.Stage != before.Stage || after.WorktreePath != before.WorktreePath ||
		after.EvidenceSHA != before.EvidenceSHA || after.EvidencePath != before.EvidencePath {
		t.Errorf("stage/worktree/evidence changed: %q %q %q %q → %q %q %q %q",
			before.Stage, before.WorktreePath, before.EvidenceSHA, before.EvidencePath,
			after.Stage, after.WorktreePath, after.EvidenceSHA, after.EvidencePath)
	}
	kinds := reacqEventKinds(t, fx.root)
	expiredAt, expiredN := reacqIndexOf(kinds, "lease.expired")
	reacqAt, reacqN := reacqIndexOf(kinds, "lease.reacquired")
	if expiredN != 1 || reacqN != 1 {
		t.Fatalf("events lease.expired x%d, lease.reacquired x%d, want exactly one each (events: %v)", expiredN, reacqN, kinds)
	}
	if expiredAt >= reacqAt {
		t.Errorf("lease.expired (event %d) must precede lease.reacquired (event %d): %v", expiredAt, reacqAt, kinds)
	}
}

// (1) The registered owner's `next --card` on its expired leased card exits 0
// and re-leases it in one step.
func TestT2013OwnerNextReleasesExpiredLeaseViaVerb(t *testing.T) {
	fx := reacqExpired(t)
	nmLaneEnv(t, "lane-1", "")
	before := fcCard(t, fx.root, "t1")

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err != nil {
		t.Fatalf("owner next --card t1 on its expired lease: %v (stderr %q), want exit 0", err, stderr)
	}
	if nmLeasedHead(out) == "" {
		t.Errorf("stdout = %q, want the leased card's head line", out)
	}
	reacqAssertReacquired(t, fx, before)
}

// (2) The factory_next MCP tool reaches the same function: identical outcome.
func TestT2013OwnerNextReleasesExpiredLeaseViaMCP(t *testing.T) {
	verb := reacqExpired(t)
	nmLaneEnv(t, "lane-1", "")
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("verb: owner next --card t1: %v (stderr %q)", err, stderr)
	}
	viaVerb := reacqOutcome(t, verb.root)

	tool := reacqExpired(t)
	nmLaneEnv(t, "lane-1", "")
	before := fcCard(t, tool.root, "t1")
	text, err := sdCallTool(t, handleFactoryNext, map[string]any{"project_root": tool.root, "run": fcRun, "card": "t1"})
	if err != nil {
		t.Fatalf("factory_next card=t1 on the owner's expired lease: %v", err)
	}
	if !strings.HasPrefix(text, "t1 stage=") {
		t.Errorf("factory_next text = %q, want the leased t1 line", text)
	}
	reacqAssertReacquired(t, tool, before)
	if viaTool := reacqOutcome(t, tool.root); viaTool != viaVerb {
		t.Errorf("tool and verb outcomes differ:\n verb: %s\n tool: %s", viaVerb, viaTool)
	}
}

// (3) `renew` on the expired lease still refuses and writes nothing.
func TestT2013RenewStillRefusesExpiredWithoutWrite(t *testing.T) {
	fx := reacqExpired(t)
	nmLaneEnv(t, "lane-1", "")
	before := fcRows(t, fx.root)

	wire, err := factoryRenewViaVerb(t)
	if err == nil {
		t.Errorf("renew on an expired lease exited 0, want a refusal")
	}
	if wire.Result != "refused-expired" {
		t.Fatalf("renew result = %q, want refused-expired", wire.Result)
	}
	if after := fcRows(t, fx.root); after != before {
		t.Errorf("a refused renewal wrote to the record:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// (4) Another lane's `next --card` on the expired leased card is refused and
// nothing is written.
func TestT2013OtherLaneNextRefusedOnExpiredLease(t *testing.T) {
	fx := reacqExpired(t)
	nmLaneEnv(t, "lane-2", "")
	before := nmSnapshot(t, fx.root, fx.store)
	rowsBefore := fcRows(t, fx.root)

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, out, stderr, err, "owned")
	if after := nmSnapshot(t, fx.root, fx.store); after != before {
		t.Errorf("a refused nomination changed the queue or the record:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if after := fcRows(t, fx.root); after != rowsBefore {
		t.Errorf("a refused nomination changed the rows or events:\nbefore:\n%s\nafter:\n%s", rowsBefore, after)
	}
}

// (5) The owner's `next --card` BEFORE the lease expires is refused as today
// ("in flight") and nothing is written.
func TestT2013OwnerNextRefusedBeforeExpiry(t *testing.T) {
	fx := reacqSetup(t, fcNow.Add(10*time.Minute))
	nmLaneEnv(t, "lane-1", "")
	before := nmSnapshot(t, fx.root, fx.store)
	rowsBefore := fcRows(t, fx.root)

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, out, stderr, err, "owned")
	if !strings.Contains(stderr, "in flight") {
		t.Errorf("stderr = %q, want the in-flight detail", stderr)
	}
	if after := nmSnapshot(t, fx.root, fx.store); after != before {
		t.Errorf("a refused nomination changed the queue or the record:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if after := fcRows(t, fx.root); after != rowsBefore {
		t.Errorf("a refused nomination changed the rows or events:\nbefore:\n%s\nafter:\n%s", rowsBefore, after)
	}
}

// (6) The uncommitted and the untracked file of the card worktree are
// byte-identical (content hash and mtime) after the owner's `next`.
func TestT2013OwnerNextLeavesWorktreeUntouched(t *testing.T) {
	fx := reacqExpired(t)
	nmLaneEnv(t, "lane-1", "")
	before := reacqWorktreeSnapshot(t, fx.wt)

	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("owner next --card t1: %v (stderr %q), want exit 0 before the worktree is compared", err, stderr)
	}
	if after := reacqWorktreeSnapshot(t, fx.wt); after != before {
		t.Errorf("the worktree changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// (7) `handoff adopt` is untouched and takes no part in the re-lease. Measured
// at the baseline of this card: adopt reads only the queue item's state (it is
// "picked" for a leased card too), so it does not refuse a leased card; what
// it must never do is write the factory record or expire/re-lease the card.
// The test pins that: the row and event snapshot is identical after the call.
func TestT2013AdoptTakesNoPartInReacquire(t *testing.T) {
	root, _, _ := adoptFixture(t)
	sdRegisterLane(t, root, "lane-1")
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		Stage: homestate.CardRun, Version: 3, LeaseExpiresAt: fcNow.Add(-10 * time.Minute).Format(time.RFC3339Nano),
		WorktreePath: t.TempDir(), EvidenceSHA: reacqEvidenceSHA, EvidencePath: reacqEvidencePath,
	})
	laneEnv(t)
	before := fcRows(t, root)

	// The call's own outcome is not asserted: adopt's gate is unchanged by this
	// card, whatever it answers for a leased card.
	_, _, _ = runFactory(t, "handoff", "adopt", "--card", "t1")
	if c := fcCard(t, root, "t1"); c.State != homestate.CardLeased || c.Version != 3 {
		t.Errorf("t1 = %s v%d after adopt, want it still leased at v3 (adopt neither expires nor re-leases)", c.State, c.Version)
	}
	if after := fcRows(t, root); after != before {
		t.Errorf("handoff adopt wrote to the factory record:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// (8) A second owner `next` right after a successful re-acquire finds a valid
// lease and is refused as in flight: no double re-acquire.
func TestT2013SecondOwnerNextAfterReacquireIsRefused(t *testing.T) {
	fx := reacqExpired(t)
	nmLaneEnv(t, "lane-1", "")
	before := fcCard(t, fx.root, "t1")

	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("first owner next --card t1: %v (stderr %q), want exit 0", err, stderr)
	}
	reacqAssertReacquired(t, fx, before)
	snapshot := fcRows(t, fx.root)

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, out, stderr, err, "owned")
	if !strings.Contains(stderr, "in flight") {
		t.Errorf("stderr = %q, want the in-flight detail", stderr)
	}
	if after := fcRows(t, fx.root); after != snapshot {
		t.Errorf("the second next wrote to the record:\nbefore:\n%s\nafter:\n%s", snapshot, after)
	}
	kinds := reacqEventKinds(t, fx.root)
	if _, n := reacqIndexOf(kinds, "lease.expired"); n != 1 {
		t.Errorf("lease.expired x%d after two nexts, want 1 (%v)", n, kinds)
	}
	if _, n := reacqIndexOf(kinds, "lease.reacquired"); n != 1 {
		t.Errorf("lease.reacquired x%d after two nexts, want 1 (%v)", n, kinds)
	}
}
