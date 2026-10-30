package ledger

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func addCard(t *testing.T, s *Store, id string, prio int, class string, depends []string, paths ...string) {
	t.Helper()
	err := s.AddCard(bg, Card{
		ID: id, Title: "card " + id, Position: core.Position{Stage: core.StageReady}, Priority: prio,
		Spec: Spec{Acceptance: []string{"TestX"}, Paths: paths, Class: class, Depends: depends},
	})
	if err != nil {
		t.Fatalf("add %s: %v", id, err)
	}
}

func req(session string, pid int) ClaimRequest {
	return ClaimRequest{SessionID: session, Worktree: "/wt/" + session, PID: pid, PIDStart: "start-" + session}
}

func cardOf(t *testing.T, s *Store, id string) Card {
	t.Helper()
	cs, err := s.Cards(bg)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cs {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("no card %s", id)
	return Card{}
}

func TestAddCardRoundTripsAndTheDatabaseClockStamps(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "c0", 0, "A", nil, "internal/z/x.go")
	addCard(t, s, "c1", 5, "C", []string{"c0"}, "internal/a/x.go")
	c := cardOf(t, s, "c1")
	if c.Title != "card c1" || c.Priority != 5 || c.Position.Stage != core.StageReady || c.Position.Park != nil {
		t.Errorf("card: %+v", c)
	}
	if c.Spec.Class != "C" || len(c.Spec.Depends) != 1 || c.Spec.Depends[0] != "c0" || len(c.Spec.Paths) != 1 {
		t.Errorf("spec: %+v", c.Spec)
	}
	if c.CreatedAt.IsZero() || time.Since(c.CreatedAt) > time.Minute || c.UpdatedAt.IsZero() {
		t.Errorf("stamps: %v %v", c.CreatedAt, c.UpdatedAt)
	}
	if err := s.AddCard(bg, Card{ID: "c1", Title: "again", Position: core.Position{Stage: core.StageReady}}); !errors.Is(err, ErrCardExists) {
		t.Errorf("a duplicate id: %v", err)
	}
	if err := s.AddCard(bg, Card{ID: "bad", Title: "x", Position: core.Position{Stage: "NOWHERE"}}); err == nil {
		t.Error("an invalid position was accepted")
	}
	if err := s.AddCard(bg, Card{ID: "", Title: "x", Position: core.Position{Stage: core.StageReady}}); !errors.Is(err, ErrInvalid) {
		t.Errorf("an empty id: %v", err)
	}
}

func TestAParkedCardKeepsItsParkTextThroughTheLedger(t *testing.T) {
	s := openTemp(t)
	p := core.Park{Reason: core.ReasonDependency, Evidence: "waits on c0", Resume: "when c0 is done"}
	if err := s.AddCard(bg, Card{ID: "p1", Title: "x", Position: core.Position{Stage: core.StageReady, Park: &p}}); err != nil {
		t.Fatal(err)
	}
	got := cardOf(t, s, "p1").Position.Park
	if got == nil || *got != p {
		t.Errorf("park: %+v", got)
	}
}

func TestClaimPicksTheHighestPriorityReadyCardAndMovesIt(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "low", 1, "B", nil, "internal/a/x.go")
	addCard(t, s, "high", 9, "C", nil, "internal/b/x.go")
	addCard(t, s, "mid", 5, "A", nil, "internal/c/x.go")
	c, err := s.Claim(bg, req("s1", 101))
	if err != nil {
		t.Fatal(err)
	}
	if c.CardID != "high" || c.Epoch != 1 {
		t.Fatalf("claimed %+v", c)
	}
	if c.Position.Stage != core.StageBuilding || c.Position.Sub != core.SubSpec {
		t.Errorf("class C starts in BUILDING/SPEC, got %+v", c.Position)
	}
	if len(c.Keys) != 1 || c.Keys[0] != "pkg:internal/b" || c.Spec.Class != "C" {
		t.Errorf("keys %v spec %+v", c.Keys, c.Spec)
	}
	if got := cardOf(t, s, "high").Position; got.Stage != core.StageBuilding || got.Sub != core.SubSpec {
		t.Errorf("the stored position: %+v", got)
	}
	d, err := s.Claim(bg, req("s2", 102))
	if err != nil || d.CardID != "mid" {
		t.Fatalf("second claim: %+v %v", d, err)
	}
	if d.Position.Stage != core.StageBuilding || d.Position.Sub != core.SubImplement {
		t.Errorf("class A starts in BUILDING/IMPLEMENT, got %+v", d.Position)
	}
}

func TestClaimWritesTheTransitionsAndTheLease(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	addCard(t, s, "c", 0, "C", nil, "internal/c/x.go")
	if _, err := s.Claim(bg, req("s1", 7)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Claim(bg, req("s2", 8)); err != nil {
		t.Fatal(err)
	}
	count := func(card string) int {
		var n int
		if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM transitions WHERE card_id=? AND epoch=1 AND actor='claim'`, card).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	// A and B skip the spec audit as two recorded moves; C stops after the first.
	if count("a") != 2 || count("c") != 1 {
		t.Errorf("transitions a=%d c=%d", count("a"), count("c"))
	}
	ls, err := s.Leases(bg)
	if err != nil || len(ls) != 2 {
		t.Fatalf("leases %v %v", ls, err)
	}
	for _, l := range ls {
		if l.Epoch != 1 || l.Worktree == "" || l.PID == 0 || l.PIDStart == "" || len(l.Keys) != 1 {
			t.Errorf("lease: %+v", l)
		}
		if d := l.ExpiresAt.Sub(l.HeartbeatAt); d < LeaseDuration-time.Second || d > LeaseDuration+time.Second {
			t.Errorf("a lease lasts %v, want %v", d, LeaseDuration)
		}
	}
	r := req("s3", 9)
	r.Lease = time.Minute
	addCard(t, s, "x", 0, "A", nil, "internal/x/x.go")
	if _, err := s.Claim(bg, r); err != nil {
		t.Fatal(err)
	}
	var secs float64
	if err := s.db.QueryRowContext(bg, `SELECT (julianday(expires_at)-julianday(heartbeat_at))*86400 FROM leases WHERE card_id='x'`).Scan(&secs); err != nil || secs < 59 || secs > 61 {
		t.Errorf("an explicit lease: %v %v", secs, err)
	}
}

func TestClaimSkipsWhatIsParkedUnmetOrKeyed(t *testing.T) {
	s := openTemp(t)
	if err := s.AddCard(bg, Card{ID: "parked", Title: "x", Priority: 100, Position: core.Position{Stage: core.StageReady, Park: &core.Park{Reason: core.ReasonDependency, Evidence: "e", Resume: "r"}}, Spec: Spec{Class: "A"}}); err != nil {
		t.Fatal(err)
	}
	addCard(t, s, "dep", 10, "A", nil, "internal/d/x.go")
	addCard(t, s, "needs", 50, "A", []string{"dep"}, "internal/n/x.go")
	c, err := s.Claim(bg, req("s1", 1))
	if err != nil || c.CardID != "dep" {
		t.Fatalf("a parked card and one with an unmet dependency are skipped: %+v %v", c, err)
	}
	// dep is held, not DONE: needs still waits, and nothing else is ready
	if _, err := s.Claim(bg, req("s2", 2)); !errors.Is(err, ErrNoWork) {
		t.Errorf("want ErrNoWork, got %v", err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET stage='DONE', sub_stage=NULL WHERE id='dep'`); err != nil {
		t.Fatal(err)
	}
	if c, err := s.Claim(bg, req("s2", 2)); err != nil || c.CardID != "needs" {
		t.Errorf("once the dependency is DONE: %+v %v", c, err)
	}
}

func TestAddCardRefusesMissingCancelledAndSelfDependencies(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "base", 1, "A", nil, "internal/a/x.go")
	if _, err := s.db.ExecContext(bg, `INSERT INTO cards(id,title,stage,spec_json,created_at,updated_at) VALUES('gone','x','CANCELLED','{}','2026-10-30T00:00:00.000Z','2026-10-30T00:00:00.000Z')`); err != nil {
		t.Fatal(err)
	}
	for name, deps := range map[string][]string{"missing": {"nobody"}, "cancelled": {"gone"}, "self": {"me"}, "one bad of two": {"base", "nobody"}} {
		err := s.AddCard(bg, Card{ID: "me", Title: "x", Position: core.Position{Stage: core.StageReady}, Spec: Spec{Class: "A", Depends: deps}})
		if !errors.Is(err, ErrDependency) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if cs, _ := s.Cards(bg); len(cs) != 2 {
		t.Errorf("a refused card was stored: %d cards", len(cs))
	}
	if err := s.AddCard(bg, Card{ID: "me", Title: "x", Position: core.Position{Stage: core.StageReady}, Spec: Spec{Class: "A", Depends: []string{"base"}}}); err != nil {
		t.Errorf("an existing dependency: %v", err)
	}
	if err := s.AddCard(bg, Card{ID: "odd", Title: "x", Position: core.Position{Stage: core.StageReady}, Spec: Spec{Class: "Z"}}); !errors.Is(err, ErrInvalid) {
		t.Errorf("an unknown class: %v", err)
	}
}

func TestAConflictKeyHeldElsewhereSkipsTheCardAndParksNothing(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "first", 9, "A", nil, "internal/a/x.go")
	addCard(t, s, "clash", 8, "A", nil, "internal/a/y.go")
	addCard(t, s, "free", 1, "A", nil, "internal/b/x.go")
	if c, err := s.Claim(bg, req("s1", 1)); err != nil || c.CardID != "first" {
		t.Fatalf("%+v %v", c, err)
	}
	c, err := s.Claim(bg, req("s2", 2))
	if err != nil || c.CardID != "free" {
		t.Fatalf("the clashing card is skipped, the free one taken: %+v %v", c, err)
	}
	if _, err := s.Claim(bg, req("s3", 3)); !errors.Is(err, ErrNoWork) {
		t.Errorf("only the clashing card is left: %v", err)
	}
	if got := cardOf(t, s, "clash"); got.Position.Stage != core.StageReady || got.Position.Park != nil {
		t.Errorf("a conflict parks nothing: %+v", got.Position)
	}
}

func TestSharedKeysConflictAcrossPackages(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "m1", 9, "A", nil, "go.mod", "internal/a/x.go")
	addCard(t, s, "m2", 8, "A", nil, "go.sum")
	if _, err := s.Claim(bg, req("s1", 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Claim(bg, req("s2", 2)); !errors.Is(err, ErrNoWork) {
		t.Errorf("go.mod and go.sum are one key: %v", err)
	}
}

func TestClaimNeedsASessionAndAWorktree(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 1, "A", nil, "internal/a/x.go")
	for name, r := range map[string]ClaimRequest{"no session": {Worktree: "/w"}, "no worktree": {SessionID: "s"}} {
		if _, err := s.Claim(bg, r); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestADeadHoldersCardIsReclaimedAtTheStartOfClaim(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 9, "A", nil, "internal/a/x.go")
	if c, err := s.Claim(bg, req("s1", 111)); err != nil || c.Epoch != 1 {
		t.Fatalf("%+v %v", c, err)
	}
	// nothing is reclaimed without a verdict (I5)
	if _, err := s.Claim(bg, req("s2", 222)); !errors.Is(err, ErrNoWork) {
		t.Fatalf("no judge, no reclaim: %v", err)
	}
	r := req("s2", 222)
	r.Alive = func(pid int, start string) Liveness {
		if pid == 111 {
			return LivenessDead
		}
		return LivenessAlive
	}
	c, err := s.Claim(bg, r)
	if err != nil || c.CardID != "a" || c.Epoch != 2 {
		t.Fatalf("the dead holder's card is claimed again at epoch 2: %+v %v", c, err)
	}
	got := cardOf(t, s, "a")
	if got.ReclaimCount != 1 || got.Position.Stage != core.StageBuilding {
		t.Errorf("reclaim count %d, position %+v", got.ReclaimCount, got.Position)
	}
	ls, _ := s.Leases(bg)
	if len(ls) != 1 || ls[0].SessionID != "s2" {
		t.Errorf("leases: %+v", ls)
	}
}

func TestALiveOrUnknownHolderIsNeverReclaimedByClaim(t *testing.T) {
	for name, verdict := range map[string]Liveness{"alive": LivenessAlive, "unknown": LivenessUnknown} {
		s := openTemp(t)
		addCard(t, s, "a", 9, "A", nil, "internal/a/x.go")
		if _, err := s.Claim(bg, req("s1", 111)); err != nil {
			t.Fatal(err)
		}
		if _, err := s.db.ExecContext(bg, `UPDATE leases SET expires_at='2000-01-01T00:00:00.000Z'`); err != nil {
			t.Fatal(err)
		}
		r := req("s2", 222)
		r.Alive = func(int, string) Liveness { return verdict }
		if _, err := s.Claim(bg, r); !errors.Is(err, ErrNoWork) {
			t.Errorf("%s holder, lease expired: %v", name, err)
		}
		if ls, _ := s.Leases(bg); len(ls) != 1 {
			t.Errorf("%s: the lease was taken: %+v", name, ls)
		}
	}
}

func TestTheThirdReclaimParksTheCardForTheOperator(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 9, "A", nil, "internal/a/x.go")
	dead := func(int, string) Liveness { return LivenessDead }
	if c, err := s.Claim(bg, req("holder-0", 100)); err != nil || c.Epoch != 1 {
		t.Fatalf("first claim: %+v %v", c, err)
	}
	// each further claim finds the previous holder dead, reclaims, and claims again
	for n := 1; n < ReclaimCap; n++ {
		r := req("holder", 100+n)
		r.Alive = dead
		c, err := s.Claim(bg, r)
		if err != nil || c.Epoch != int64(n+1) {
			t.Fatalf("reclaim %d: %+v %v", n, c, err)
		}
		if got := cardOf(t, s, "a"); got.ReclaimCount != n || got.Position.Park != nil {
			t.Fatalf("after reclaim %d: %+v", n, got)
		}
	}
	r := req("last", 999)
	r.Alive = dead
	if _, err := s.Claim(bg, r); !errors.Is(err, ErrNoWork) {
		t.Fatalf("the %d-th reclaim parks the card, so nothing is left to claim: %v", ReclaimCap, err)
	}
	got := cardOf(t, s, "a")
	if got.ReclaimCount != ReclaimCap || got.Position.Park == nil || got.Position.Park.Reason != ParkReasonReclaimCap {
		t.Errorf("after %d reclaims: count %d, position %+v", ReclaimCap, got.ReclaimCount, got.Position)
	}
	if got.Position.Stage != core.StageBuilding {
		t.Errorf("a park does not move the stage: %+v", got.Position)
	}
	if ls, _ := s.Leases(bg); len(ls) != 0 {
		t.Errorf("a parked card holds no lease: %+v", ls)
	}
}

// Two processes on the same file never both hold one card.
func TestConcurrentClaimsNeverShareACard(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.db")
	const cards, workers = 8, 4
	seed, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < cards; i++ {
		addCard(t, seed, string(rune('a'+i)), i, "A", nil, "internal/"+string(rune('a'+i))+"/x.go")
	}
	_ = seed.Close()
	var mu sync.Mutex
	got := map[string]int{}
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			s, err := Open(context.Background(), path)
			if err != nil {
				t.Error(err)
				return
			}
			defer s.Close()
			for {
				c, err := s.Claim(context.Background(), req("w"+string(rune('0'+w)), 1000+w))
				if errors.Is(err, ErrNoWork) {
					return
				}
				if err != nil {
					t.Error(err)
					return
				}
				mu.Lock()
				got[c.CardID]++
				mu.Unlock()
			}
		}(w)
	}
	wg.Wait()
	if len(got) != cards {
		t.Errorf("claimed %d of %d cards: %v", len(got), cards, got)
	}
	for id, n := range got {
		if n != 1 {
			t.Errorf("card %s claimed %d times", id, n)
		}
	}
}

func coreReady() core.Position { return core.Position{Stage: core.StageReady} }

// cards.epoch is the epoch of the latest claim; a reclaim does not move it, a new
// claim moves it by one (the lander of an older epoch is fenced by it).
func TestCardsEpochIsTheLatestClaimsAndOnlyAClaimMovesIt(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 9, "A", nil, "internal/a/x.go")
	epoch := func() int64 {
		var e int64
		if err := s.db.QueryRowContext(bg, `SELECT epoch FROM cards WHERE id='a'`).Scan(&e); err != nil {
			t.Fatal(err)
		}
		return e
	}
	if epoch() != 0 || cardOf(t, s, "a").Epoch != 0 {
		t.Fatal("a new card is at epoch 0")
	}
	if _, err := s.Claim(bg, req("s1", 11)); err != nil || epoch() != 1 {
		t.Fatalf("first claim: epoch %d %v", epoch(), err)
	}
	if _, err := s.Reap(bg, alwaysDead); err != nil || epoch() != 1 || cardOf(t, s, "a").Epoch != 1 {
		t.Fatalf("a reclaim leaves the epoch: %d %v", epoch(), err)
	}
	if c, err := s.Claim(bg, req("s2", 12)); err != nil || c.Epoch != 2 || epoch() != 2 {
		t.Fatalf("second claim: %+v epoch %d %v", c, epoch(), err)
	}
}

// A card that is already parked when its holder is reclaimed for the third time: the
// count goes up, the park stays as it was, and nothing fails.
func TestAReclaimOfAnAlreadyParkedCardCountsAndKeepsThePark(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "a", 9, "A", nil, "internal/a/x.go")
	if _, err := s.Claim(bg, req("s1", 11)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE cards SET reclaim_count=?, parked=1, park_reason='operator-hold', park_evidence='held by the operator', park_resume='when told'`, ReclaimCap-1); err != nil {
		t.Fatal(err)
	}
	got, err := s.Reap(bg, alwaysDead)
	if err != nil || len(got) != 1 || got[0].ReclaimCount != ReclaimCap || got[0].Parked {
		t.Fatalf("%+v %v", got, err)
	}
	c := cardOf(t, s, "a")
	if c.ReclaimCount != ReclaimCap || c.Position.Park == nil || c.Position.Park.Reason != core.ReasonOperatorHold || c.Position.Park.Evidence != "held by the operator" {
		t.Errorf("the park stays as it was: %+v", c)
	}
}

// A card with no paths holds no keys, so only the lease itself stops a second claim.
func TestAHeldCardWithNoPathsIsNotClaimedAgain(t *testing.T) {
	s := openTemp(t)
	addCard(t, s, "free", 1, "A", nil)
	if c, err := s.Claim(bg, req("s1", 1)); err != nil || c.CardID != "free" || len(c.Keys) != 0 {
		t.Fatalf("%+v %v", c, err)
	}
	if c, err := s.Claim(bg, req("s2", 2)); !errors.Is(err, ErrNoWork) {
		t.Errorf("a held card was claimed again: %+v %v", c, err)
	}
}
