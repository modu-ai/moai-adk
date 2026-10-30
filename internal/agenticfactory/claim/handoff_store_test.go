package claim

import (
	"errors"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func handedOff(t *testing.T, s *Store, n int, branch string) core.IssueRef {
	t.Helper()
	i := ref(t, n)
	rec, err := s.Acquire(i, "w", "d-"+branch, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateRelease(i, "w", rec.Epoch, func(r *Record) error {
		r.State = core.StateLanding
		r.Receipt = &Receipt{PR: n, Head: core.HeadSHA("a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"), Branch: branch, Worktree: "/wt"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return i
}

func TestUpdateReleaseClearsHolderAtomically(t *testing.T) {
	s := open(t)
	i := ref(t, 1)
	rec, _ := s.Acquire(i, "w", "d", "run-1")
	boom := errors.New("boom")
	if _, err := s.UpdateRelease(i, "w", rec.Epoch, func(*Record) error { return boom }); !errors.Is(err, boom) {
		t.Fatal(err)
	}
	if cur, err := s.Load(i); err != nil || cur.Holder != "w" {
		t.Errorf("failed release changed holder: %+v %v", cur, err)
	}
	if _, err := s.UpdateRelease(i, "x", rec.Epoch, func(*Record) error { return nil }); !errors.Is(err, core.ErrOwnershipUnverified) {
		t.Errorf("foreign release: %v", err)
	}
	out, err := s.UpdateRelease(i, "w", rec.Epoch, func(*Record) error { return nil })
	if err != nil || out.Holder != "" || out.Seq != rec.Seq+1 {
		t.Errorf("release: %+v %v", out, err)
	}
}

func TestUpdateReleasedRefusesHeldAndMissing(t *testing.T) {
	s := open(t)
	i := ref(t, 1)
	if _, err := s.UpdateReleased(i, func(*Record) error { return nil }); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("missing: %v", err)
	}
	_, _ = s.Acquire(i, "w", "d", "run-1")
	if _, err := s.UpdateReleased(i, func(*Record) error { return nil }); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("held record must not be written by the tracker: %v", err)
	}
	_ = releaseNow(t, s, i, "w")
	boom := errors.New("boom")
	if _, err := s.UpdateReleased(i, func(*Record) error { return boom }); !errors.Is(err, boom) {
		t.Errorf("fn error: %v", err)
	}
	out, err := s.UpdateReleased(i, func(r *Record) error { r.Holder = "sneaky"; return nil })
	if err != nil || out.Holder != "" {
		t.Errorf("fn must not be able to set a holder: %+v %v", out, err)
	}
}

// reopened is handedOff followed by the reopen edge landing -> verifying.
func reopened(t *testing.T, s *Store, n int, branch string) core.IssueRef {
	t.Helper()
	i := handedOff(t, s, n, branch)
	if _, err := s.UpdateReleased(i, func(r *Record) error { r.State = core.StateVerifying; return nil }); err != nil {
		t.Fatal(err)
	}
	return i
}

func TestRepairClaimNeedsReopenedIssue(t *testing.T) {
	s := open(t)
	i := handedOff(t, s, 5, "WT-b")
	if _, err := s.AcquireRepair(i, "r", "dr", "run-2", "WT-b"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("repair from landing: %v", err)
	}
}

func TestAcquireRepairRules(t *testing.T) {
	s := open(t)
	i := reopened(t, s, 5, "WT-b")
	if _, err := s.AcquireRepair(ref(t, 99), "r", "dr", "run-2", "WT-b"); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("unknown issue: %v", err)
	}
	if _, err := s.AcquireRepair(i, "", "dr", "run-2", "WT-b"); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("empty holder: %v", err)
	}
	if _, err := s.AcquireRepair(i, "r", "dr", "run-2", "WT-other"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("wrong branch: %v", err)
	}
	rec, err := s.AcquireRepair(i, "r", "dr", "run-2", "WT-b")
	if err != nil || rec.Holder != "r" || rec.Epoch != 2 || rec.State != core.StateClaimed || rec.Receipt == nil {
		t.Fatalf("repair: %+v %v", rec, err)
	}
	if _, err := s.AcquireRepair(i, "r2", "dr2", "run-3", "WT-b"); !errors.Is(err, core.ErrHeld) {
		t.Errorf("second repairer: %v", err)
	}
	if _, err := s.AcquireRepair(i, "r", "dr", "run-2", "WT-b"); !errors.Is(err, core.ErrDuplicateDispatch) {
		t.Errorf("duplicate: %v", err)
	}
}

func TestAcquireRefusesPublishedAndDoneIssues(t *testing.T) {
	s := open(t)
	i := handedOff(t, s, 5, "WT-b")
	if _, err := s.Acquire(i, "z", "dz", "run-9"); !errors.Is(err, core.ErrRepairOnly) {
		t.Errorf("published: %v", err)
	}
	if _, err := s.UpdateReleased(i, func(r *Record) error { accept(r); return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateReleased(i, func(r *Record) error { finish(r); return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Acquire(i, "z", "dz", "run-9"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("done: %v", err)
	}
	if _, err := s.AcquireRepair(i, "r", "dr", "run-2", "WT-b"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("repair on done: %v", err)
	}
}

func TestBacklogCountAndReadAndList(t *testing.T) {
	s := open(t)
	_ = s.SetMaxOpenPRs(2)
	handedOff(t, s, 1, "WT-1")
	handedOff(t, s, 2, "WT-2")
	if n, err := s.openPRs(); err != nil || n != 2 {
		t.Fatalf("open PRs = %d %v", n, err)
	}
	if _, err := s.Acquire(ref(t, 3), "w", "d3", "run-1"); !errors.Is(err, core.ErrBacklogFull) {
		t.Errorf("cap: %v", err)
	}
	_, _ = s.UpdateReleased(ref(t, 1), func(r *Record) error { accept(r); return nil })
	_, _ = s.UpdateReleased(ref(t, 1), func(r *Record) error { finish(r); return nil })
	if _, err := s.Acquire(ref(t, 3), "w", "d3", "run-1"); err != nil {
		t.Errorf("after one is done: %v", err)
	}
	recs, err := s.List()
	if err != nil || len(recs) != 3 || recs[0].Issue != 1 || recs[2].Issue != 3 {
		t.Errorf("list: %+v %v", recs, err)
	}
	if r, err := s.Read(ref(t, 1)); err != nil || r.Holder != "" {
		t.Errorf("read released: %+v %v", r, err)
	}
	if _, err := s.Read(ref(t, 77)); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("read missing: %v", err)
	}
}

func TestListFailsOnCorruptRecord(t *testing.T) {
	s := open(t)
	handedOff(t, s, 1, "WT-1")
	if err := writeRaw(s, 1, "{bad"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.List(); !errors.Is(err, core.ErrStoreUnavailable) {
		t.Errorf("corrupt list: %v", err)
	}
	// The limit always applies, so an unreadable store must refuse new claims
	// instead of guessing how many PRs are open.
	for _, n := range []int{9, 10} {
		if _, err := s.Acquire(ref(t, n), "w", "d", "run-1"); !errors.Is(err, core.ErrStoreUnavailable) {
			t.Errorf("claim %d with unreadable store must not guess: %v", n, err)
		}
	}
}

func writeRaw(s *Store, n int, body string) error {
	r, _ := core.NewIssueRef(n)
	return os.WriteFile(s.path(r), []byte(body), 0o600)
}
