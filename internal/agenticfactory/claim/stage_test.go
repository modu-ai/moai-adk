package claim

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

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
		r.Wait = &core.Wait{Owner: "w", NextAction: "retry", RecheckAt: time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC), Diagnosis: "429", MaxRetries: 3, Escalate: "leader"}
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
	// ...but it says so: an empty position must not read as "confirmed: no park"
	if rec.PositionUnknown == "" {
		t.Errorf("a record whose position could not be derived must carry the reason: %+v", rec)
	}
}

// One unreadable position does not fail the listing, and a healthy record never
// carries the marker. The marker is derived on every read and never stored.
func TestAnUnmappablePositionIsMarkedNotHiddenAndNotFatalToTheListing(t *testing.T) {
	s := open(t)
	if _, err := s.Acquire(ref(t, 2), "w2", "d2", "run-2"); err != nil {
		t.Fatal(err)
	}
	if err := writeRaw(s, 1, `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"blocked"}`); err != nil {
		t.Fatal(err)
	}
	recs, err := s.List()
	if err != nil || len(recs) != 2 {
		t.Fatalf("the listing must survive one unmappable record: %d %v", len(recs), err)
	}
	if recs[0].Issue != 1 || recs[0].PositionUnknown == "" {
		t.Errorf("the unmappable record is marked: %+v", recs[0])
	}
	if recs[1].PositionUnknown != "" || recs[1].Stage == "" {
		t.Errorf("a healthy record carries a position and no marker: %+v", recs[1])
	}
	if raw, err := os.ReadFile(s.path(ref(t, 1))); err != nil || strings.Contains(string(raw), "position_unknown") {
		t.Errorf("the marker must not be written: %s %v", raw, err)
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
	if r := up(func(r *Record) { finish(r) }); r.Stage != core.StageDone {
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

// Only the entry into PARKED is checked: a record stored before the rule, already
// parked and without a wait record, is read and rewritten as it is.
func TestOnlyTheEntryIntoParkedNeedsAWaitRecord(t *testing.T) {
	s := open(t)
	i := ref(t, 1)
	if _, err := s.Acquire(i, "w", "d", "run-1"); err != nil {
		t.Fatal(err)
	}
	cur, _ := s.Read(i)
	enter := func(r *Record) error {
		r.State, r.Return, r.Reason, r.Resume = core.StateBlocked, core.StateClaimed, core.ReasonRateLimited, "later"
		return nil
	}
	if _, err := s.Update(i, "w", cur.Epoch, enter); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("entering a park without a wait record: %v", err)
	}
	if now, _ := s.Read(i); now.State != core.StateClaimed || now.Park != nil {
		t.Fatalf("the refused entry changed the record: %+v", now)
	}
	// an old record, parked and without a wait, written before the rule
	if err := writeRaw(s, 2, `{"issue":2,"holder":"w","dispatch":"d2","run":"run-1","epoch":1,"seq":1,"state":"blocked","return_stage":"claimed","block_reason":"rate-limited","resume_condition":"later"}`); err != nil {
		t.Fatal(err)
	}
	two := ref(t, 2)
	old, err := s.Read(two)
	if err != nil || old.Park == nil || old.Wait != nil {
		t.Fatalf("an old parked record must still read: %+v %v", old, err)
	}
	if _, err := s.Update(two, "w", old.Epoch, func(r *Record) error { r.Dispatch = "d2b"; return nil }); err != nil {
		t.Errorf("rewriting an already parked record: %v", err)
	}
}

// A receipt stored before the spec fields existed still reads, and the record
// around it is usable.
func TestAReceiptWithoutTheSpecFieldsStillReads(t *testing.T) {
	s := open(t)
	if err := writeRaw(s, 3, `{"issue":3,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"landing","receipt":{"handoff_id":"h-1","leader_accepted":true,"pr":5,"head":"a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1","branch":"WT-x","worktree":"/wt/x"}}`); err != nil {
		t.Fatal(err)
	}
	rec, err := s.Read(ref(t, 3))
	if err != nil || rec.Receipt == nil || rec.Receipt.PR != 5 || rec.Receipt.PRURL != "" || rec.Stage != core.StageReviewing {
		t.Fatalf("an old receipt: %+v %v", rec, err)
	}
	if _, err := s.Update(ref(t, 3), "w", rec.Epoch, func(r *Record) error { r.Pending = "deploy"; return nil }); err != nil {
		t.Errorf("rewriting a record with an old receipt: %v", err)
	}
}

// An old parked record has no wait, so the write rule does not stop its hand-over
// from dropping one; the takeover itself must ask for the new holder's.
func TestTakingOverAnOldParkedRecordStillNeedsAWait(t *testing.T) {
	s := open(t)
	if err := writeRaw(s, 4, `{"issue":4,"holder":"w","dispatch":"d4","run":"run-1","epoch":1,"seq":1,"state":"blocked","return_stage":"claimed","block_reason":"rate-limited","resume_condition":"later"}`); err != nil {
		t.Fatal(err)
	}
	i := ref(t, 4)
	if _, err := s.TakeOverUnverified(i, "w", 1, "w2", "d4b", "run-2", "ref@1", nil); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("no wait for the new holder: %v", err)
	}
	w := core.Wait{Owner: "w2", NextAction: "retry", RecheckAt: time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC), Diagnosis: "429", MaxRetries: 3, Escalate: "leader"}
	rec, err := s.TakeOverUnverified(i, "w", 1, "w2", "d4b", "run-2", "ref@1", &w)
	if err != nil || rec.Holder != "w2" || rec.Wait == nil || rec.Park == nil {
		t.Errorf("with the new holder's wait: %+v %v", rec, err)
	}
}

// A write that repairs the record clears the marker on the record it returns; the
// marker is a fact about the stored position and must not outlive the fix.
func TestARepairingWriteClearsPositionUnknown(t *testing.T) {
	s := open(t)
	if err := writeRaw(s, 1, `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"blocked"}`); err != nil {
		t.Fatal(err)
	}
	cur, _ := s.Read(ref(t, 1))
	if cur.PositionUnknown == "" {
		t.Fatal("setup: the record must be unmappable")
	}
	out, err := s.Update(ref(t, 1), "w", cur.Epoch, func(r *Record) error {
		r.Return, r.Reason, r.Resume = core.StateClaimed, core.ReasonDebtPresent, "rerun" // a gate reason on a legacy pause maps
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.PositionUnknown != "" || out.Stage == "" {
		t.Errorf("the returned record still says its position is unknown: %+v", out)
	}
	if again, _ := s.Read(ref(t, 1)); again.PositionUnknown != "" {
		t.Errorf("the stored record reads as unknown: %q", again.PositionUnknown)
	}
}

// The marker holds one of a few fixed codes, not the text of an error.
func TestPositionUnknownIsAFixedReasonCode(t *testing.T) {
	good := `"issue":%d,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1`
	_ = good
	cases := []struct {
		name, raw string
		want      string
	}{
		{"unmappable legacy state", `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"blocked"}`, PositionUnmappable},
		{"gate that is not a gate", `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"claimed","gate":{"reason":"rate-limited","detail":"x","resume_condition":"y"}}`, PositionInvalidGate},
		{"refusal without detail", `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"claimed","refusal":{"reason":"foreign-closing-keyword"}}`, PositionInvalidRefusal},
		{"wait without escalation", `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"claimed","wait":{"owner":"w","next_action":"a","recheck_at":"2026-10-30T09:00:00Z","diagnosis":"d","max_retries":2}}`, PositionInvalidWait},
		{"wait owned by someone else", `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"claimed","wait":{"owner":"x","next_action":"a","recheck_at":"2026-10-30T09:00:00Z","diagnosis":"d","max_retries":2,"escalate_to":"leader"}}`, PositionWaitOwner},
		{"repair evidence incomplete", `{"issue":1,"holder":"w","dispatch":"d","run":"run-1","epoch":1,"seq":1,"state":"verifying","repair":{"decision":"d","exclusive_owner":"w"}}`, PositionInvalidRepair},
	}
	for _, c := range cases {
		s := open(t)
		if err := writeRaw(s, 1, c.raw); err != nil {
			t.Fatal(err)
		}
		rec, err := s.Read(ref(t, 1))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if rec.PositionUnknown != c.want {
			t.Errorf("%s: marker = %q, want the code %q", c.name, rec.PositionUnknown, c.want)
		}
	}
}
