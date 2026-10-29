package claim

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestRecordCarriesTheS46PositionAndKeepsTheOldState(t *testing.T) {
	s := open(t)
	i := ref(t, 1)
	rec, err := s.Acquire(i, "w", "d", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != core.StateClaimed || rec.Stage != core.StageBuilding || rec.Sub != core.SubSpec {
		t.Fatalf("claimed: %+v", rec)
	}
	step := func(fn func(*Record)) Record {
		t.Helper()
		out, err := s.Update(i, "w", rec.Epoch, func(r *Record) error { fn(r); return nil })
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	if r := step(func(r *Record) { r.State = core.StateImplementing }); r.Stage != core.StageBuilding || r.Sub != core.SubImplement {
		t.Errorf("implementing: %+v", r)
	}
	if r := step(func(r *Record) { r.State = core.StateVerifying }); r.Stage != core.StageBuilding || r.Sub != core.SubImplement {
		t.Errorf("verifying has no stage of its own: %+v", r)
	}
	r := step(func(r *Record) {
		r.Return, r.State, r.Reason, r.Resume = core.StateVerifying, core.StateBlocked, core.ReasonRateLimited, "after 15:00"
	})
	if r.Stage != core.StageBuilding || r.Park == nil || r.Park.Reason != core.ReasonRateLimited || r.Park.Resume != "after 15:00" {
		t.Errorf("blocked becomes a park on the same stage: %+v", r)
	}
	raw, err := os.ReadFile(s.path(i))
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{`"stage": "BUILDING"`, `"sub_stage": "IMPLEMENT"`, `"park": {`, `"state": "blocked"`, `"return_stage": "verifying"`, `"block_reason": "rate-limited"`} {
		if !strings.Contains(string(raw), k) {
			t.Errorf("stored record lacks %s:\n%s", k, raw)
		}
	}
	if strings.Contains(string(raw), `"parked"`) {
		t.Error("the record must not use the queue's `parked` string key")
	}
}

func TestOldRecordsAreNormalizedWhenRead(t *testing.T) {
	legacy := func(extra string) string {
		return `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,` + extra + `}`
	}
	cases := []struct {
		name  string
		body  string
		stage core.Stage
		sub   core.SubStage
		park  core.BlockReason
		gate  core.BlockReason
		refus core.BlockReason
	}{
		{"verifying", legacy(`"state":"verifying"`), core.StageBuilding, core.SubImplement, "", "", ""},
		{"blocked on a quota", legacy(`"state":"blocked","return_stage":"implementing","block_reason":"rate-limited","resume_condition":"later"`), core.StageBuilding, core.SubImplement, core.ReasonRateLimited, "", ""},
		{"blocked on a gate reason", legacy(`"state":"blocked","return_stage":"verifying","block_reason":"debt-present","resume_condition":"rerun"`), core.StageBuilding, core.SubImplement, "", core.ReasonDebtPresent, ""},
		{"blocked on a refusal reason", legacy(`"state":"blocked","return_stage":"ready","block_reason":"foreign-closing-keyword","resume_condition":"reword"`), core.StageBuilding, core.SubImplement, "", "", core.ReasonForeignClosing},
		{"on-hold", legacy(`"state":"on-hold","return_stage":"claimed","block_reason":"operator-hold","resume_condition":"operator"`), core.StageBuilding, core.SubSpec, core.ReasonOperatorHold, "", ""},
		{"landing with an accepted receipt", legacy(`"state":"landing","receipt":{"handoff_id":"h-1","leader_accepted":true,"pr":5,"head":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1","branch":"WT-b","worktree":"/wt"}`), core.StageReviewing, "", "", "", ""},
		{"landing merged", legacy(`"state":"landing","merged":true,"pending_effect":"deploy","receipt":{"handoff_id":"h-1","leader_accepted":true,"pr":5,"head":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1","branch":"WT-b","worktree":"/wt"}`), core.StageMerging, "", "", "", ""},
		{"done", legacy(`"state":"done"`), core.StageDone, "", "", "", ""},
	}
	for _, c := range cases {
		s := open(t)
		if err := writeRaw(s, 1, c.body); err != nil {
			t.Fatal(err)
		}
		rec, err := s.Read(ref(t, 1))
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if rec.Stage != c.stage || rec.Sub != c.sub {
			t.Errorf("%s: stage %s/%s want %s/%s", c.name, rec.Stage, rec.Sub, c.stage, c.sub)
		}
		if (rec.Park == nil) != (c.park == "") || (rec.Park != nil && rec.Park.Reason != c.park) {
			t.Errorf("%s: park %+v want %q", c.name, rec.Park, c.park)
		}
		if (rec.Gate == nil) != (c.gate == "") || (rec.Gate != nil && rec.Gate.Reason != c.gate) {
			t.Errorf("%s: gate %+v want %q", c.name, rec.Gate, c.gate)
		}
		if (rec.Refusal == nil) != (c.refus == "") || (rec.Refusal != nil && rec.Refusal.Reason != c.refus) {
			t.Errorf("%s: refusal %+v want %q", c.name, rec.Refusal, c.refus)
		}
		if rec.State == "" {
			t.Errorf("%s: the old state must stay", c.name)
		}
	}
}

// A record that cannot be mapped is still readable (the reader is not locked
// out); it simply has no S46 position.
func TestAnUnmappableRecordIsStillReadable(t *testing.T) {
	s := open(t)
	if err := writeRaw(s, 1, `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"blocked"}`); err != nil {
		t.Fatal(err)
	}
	rec, err := s.Read(ref(t, 1))
	if err != nil || rec.Stage != "" || rec.State != core.StateBlocked {
		t.Errorf("read: %+v %v", rec, err)
	}
}

func TestAWriteThatCannotBeMappedIsRefusedAndChangesNothing(t *testing.T) {
	s := open(t)
	i := ref(t, 1)
	rec, _ := s.Acquire(i, "w", "d", "run-1")
	_, err := s.Update(i, "w", rec.Epoch, func(r *Record) error { r.State = core.StateBlocked; return nil })
	if !errors.Is(err, core.ErrLegacyUnmappable) {
		t.Fatalf("blocked with no return stage: %v", err)
	}
	_, err = s.Update(i, "w", rec.Epoch, func(r *Record) error {
		r.Return, r.State, r.Reason, r.Resume = core.StateClaimed, core.StateBlocked, core.ReasonDebtPresent, "r"
		return nil
	})
	if err != nil {
		t.Errorf("a gate reason on a legacy pause is read as a gate, not refused: %v", err)
	}
	if cur, _ := s.Load(i); cur.State == core.StateBlocked && cur.Gate == nil {
		t.Errorf("stored record: %+v", cur)
	}
}

func TestGateAndRefusalAreStoredWithoutMovingTheStage(t *testing.T) {
	s := open(t)
	i := ref(t, 1)
	rec, _ := s.Acquire(i, "w", "d", "run-1")
	out, err := s.Update(i, "w", rec.Epoch, func(r *Record) error {
		r.State = core.StateVerifying
		r.Gate = &core.Gate{Reason: core.ReasonCheckNotPassed, Detail: "check skipped", Resume: "same-HEAD rerun passes", Head: "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"}
		r.Refusal = &core.Refusal{Reason: core.ReasonForeignClosing, Detail: "would close #3000"}
		return nil
	})
	if err != nil || out.State != core.StateVerifying || out.Stage != core.StageBuilding || out.Park != nil || out.Gate == nil || out.Refusal == nil {
		t.Fatalf("gate+refusal: %+v %v", out, err)
	}
	if _, err := s.Update(i, "w", rec.Epoch, func(r *Record) error {
		r.Gate = &core.Gate{Reason: core.ReasonRateLimited, Detail: "d", Resume: "r"}
		return nil
	}); !errors.Is(err, core.ErrReasonNotAllowed) {
		t.Errorf("a non-gate reason as a gate: %v", err)
	}
	if _, err := s.Update(i, "w", rec.Epoch, func(r *Record) error {
		r.Refusal = &core.Refusal{Reason: "made-up", Detail: "d"}
		return nil
	}); !errors.Is(err, core.ErrReasonUnknown) {
		t.Errorf("an unknown refusal reason: %v", err)
	}
}

func TestHandoffAcceptanceAndMergeMoveThePosition(t *testing.T) {
	s := open(t)
	i := handedOff(t, s, 5, "WT-b")
	rec, _ := s.Read(i)
	if rec.Stage != core.StageBuilding {
		t.Errorf("a receipt the leader has not accepted keeps the work in BUILDING: %+v", rec)
	}
	up := func(fn func(*Record)) Record {
		t.Helper()
		out, err := s.UpdateReleased(i, func(r *Record) error { fn(r); return nil })
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	if r := up(func(r *Record) { r.Receipt.LeaderAccepted = true }); r.Stage != core.StageReviewing {
		t.Errorf("accepted receipt: %+v", r)
	}
	if r := up(func(r *Record) { r.Merged, r.Pending = true, "deploy" }); r.Stage != core.StageMerging {
		t.Errorf("merge observed: %+v", r)
	}
	if r := up(func(r *Record) { r.State = core.StateDone; r.Closed = true }); r.Stage != core.StageDone {
		t.Errorf("done: %+v", r)
	}
}

// A repair claim starts the work over: a gate or refusal left on the reopened
// record belongs to the previous attempt and must not follow the new holder.
func TestRepairClaimDropsTheGateAndRefusalOfThePreviousAttempt(t *testing.T) {
	s := open(t)
	i := reopened(t, s, 5, "WT-b")
	if _, err := s.UpdateReleased(i, func(r *Record) error {
		r.Gate = &core.Gate{Reason: core.ReasonDebtPresent, Detail: "old", Resume: "rerun", Head: r.Head}
		r.Refusal = &core.Refusal{Reason: core.ReasonForeignClosing, Detail: "old"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rec, err := s.AcquireRepair(i, "r", "dr", "run-2", "WT-b")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Gate != nil || rec.Refusal != nil {
		t.Fatalf("stale gate/refusal followed the repair claim: %+v %+v", rec.Gate, rec.Refusal)
	}
	if got, _ := s.Read(i); got.Gate != nil || got.Refusal != nil {
		t.Fatalf("stored record keeps them: %+v %+v", got.Gate, got.Refusal)
	}
}
