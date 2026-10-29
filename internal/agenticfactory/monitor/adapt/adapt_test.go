package adapt

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

const head = "0123456789abcdef0123456789abcdef01234567"

func openStore(t *testing.T) *claim.Store {
	t.Helper()
	s, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestReadClaimMapsRecordAndDoesNotWrite(t *testing.T) {
	s := openStore(t)
	iss, _ := core.NewIssueRef(7)
	run, _ := core.ParseRunID("run-1")
	if _, err := s.Acquire(iss, "lane-18", "cmd-1", run); err != nil {
		t.Fatal(err)
	}
	before, _ := s.Load(iss)

	row, err := (ClaimSource{Store: s, Project: "p"}).ReadClaim(t.Context(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if row.Issue != 7 || row.Owner != "lane-18" || row.Project != "p" || int64(row.StoreSeq) != before.Seq {
		t.Errorf("row = %+v", row)
	}
	after, _ := s.Load(iss)
	if after.Seq != before.Seq {
		t.Errorf("read changed store seq %d -> %d", before.Seq, after.Seq)
	}
}

func TestReadClaimMissingIsNotFound(t *testing.T) {
	_, err := (ClaimSource{Store: openStore(t)}).ReadClaim(t.Context(), 9)
	if !errors.Is(err, monitor.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
	if _, err := (ClaimSource{Store: openStore(t)}).ReadClaim(t.Context(), 0); err == nil {
		t.Fatal("issue 0 accepted")
	}
}

func TestRecordedFromMapsDeliveryAndHead(t *testing.T) {
	rec := claim.Record{Issue: 7, Head: core.HeadSHA(head), Delivery: core.DeliveryExecuted}
	got := RecordedFrom(rec, true, "cmd-2")
	if got.Issue != 7 || got.HeadSHA != head || got.Delivery != effect.Executed || !got.LeaseExpired || got.PendingDispatch != "cmd-2" {
		t.Errorf("recorded = %+v", got)
	}
}

func TestLeaseExpiredUsesDurationAndZeroMeansNot(t *testing.T) {
	now := time.Date(2026, 10, 29, 7, 0, 0, 0, time.UTC)
	if !LeaseExpired(now.Add(-16*time.Minute), 15*time.Minute, now) {
		t.Error("16m old with 15m lease must be expired")
	}
	if LeaseExpired(now.Add(-5*time.Minute), 15*time.Minute, now) {
		t.Error("5m old must not be expired")
	}
	if LeaseExpired(time.Time{}, 15*time.Minute, now) {
		t.Error("unknown renewal time must not claim expiry")
	}
}

func TestReadClaimCarriesAgentNameAndSessionID(t *testing.T) {
	s := openStore(t)
	iss, _ := core.NewIssueRef(7)
	run, _ := core.ParseRunID("run-1")
	if _, err := s.Acquire(iss, "agent-3", "cmd-1", run); err != nil {
		t.Fatal(err)
	}
	const sid = "2a8642ba-c7da-448b-a256-444facf4c514"
	row, err := (ClaimSource{Store: s, Session: func(int, string) (string, error) { return sid, nil }}).ReadClaim(t.Context(), 7)
	if err != nil || row.AgentName != "agent-3" || row.SessionID != sid {
		t.Fatalf("row = %+v err = %v", row, err)
	}
	row, _ = (ClaimSource{Store: s}).ReadClaim(t.Context(), 7)
	if row.SessionID != "" {
		t.Errorf("no resolver must leave session id empty, got %q", row.SessionID)
	}
}

// legacyRecord writes a record the way the store wrote it before the S46 keys existed:
// the new keys are removed from the file, so the read has to derive the position.
func legacyRecord(t *testing.T, dir string, mutate func(m map[string]any)) {
	t.Helper()
	path := filepath.Join(dir, "issue-7.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"stage", "sub_stage", "park", "gate", "refusal"} {
		delete(m, k)
	}
	if mutate != nil {
		mutate(m)
	}
	out, _ := json.Marshal(m)
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestReadClaimDerivesTheStageFromARecordWithoutTheNewKeys(t *testing.T) {
	for name, tc := range map[string]struct {
		mutate func(map[string]any)
		stage  string
	}{
		"claimed state":  {func(m map[string]any) { m["state"] = "claimed" }, "BUILDING"},
		"implementing":   {func(m map[string]any) { m["state"] = "implementing" }, "BUILDING"},
		"landing":        {func(m map[string]any) { m["state"] = "landing" }, "BUILDING"},
		"landing merged": {func(m map[string]any) { m["state"] = "landing"; m["merged"] = true }, "MERGING"},
		"done":           {func(m map[string]any) { m["state"] = "done" }, "DONE"},
		"queued":         {func(m map[string]any) { m["state"] = "queued" }, "READY"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			s, err := claim.OpenWithLimit(dir, 100)
			if err != nil {
				t.Fatal(err)
			}
			iss, _ := core.NewIssueRef(7)
			run, _ := core.ParseRunID("run-1")
			if _, err := s.Acquire(iss, "lane-18", "cmd-1", run); err != nil {
				t.Fatal(err)
			}
			legacyRecord(t, dir, tc.mutate)
			row, err := (ClaimSource{Store: s}).ReadClaim(t.Context(), 7)
			if err != nil {
				t.Fatal(err)
			}
			if row.Stage != tc.stage {
				t.Errorf("stage = %q, want %q (state %q)", row.Stage, tc.stage, row.State)
			}
			if row.Park != nil || row.Gate != nil || row.Refusal != nil {
				t.Errorf("a legacy record shows overlays: %+v %+v %+v", row.Park, row.Gate, row.Refusal)
			}
		})
	}
}

func TestReadClaimCarriesParkGateAndRefusalFromTheRecord(t *testing.T) {
	r := claim.Record{
		Issue: 7, State: core.StateImplementing,
		Stage: core.StageReviewing, Sub: core.SubFix,
		Park:    &core.Park{Reason: core.ReasonPendingEffect, Evidence: "label apply unconfirmed", Resume: "after readback"},
		Gate:    &core.Gate{Reason: core.ReasonCheckNotPassed, Detail: "build red", Resume: "re-run on the new head", Head: core.HeadSHA(head)},
		Refusal: &core.Refusal{Reason: core.ReasonNotIssueAuthor, Detail: "edit refused"},
	}
	row := rowFrom(r, "p", "")
	if row.Stage != "REVIEWING" || row.SubStage != "FIX" {
		t.Errorf("stage=%q sub=%q", row.Stage, row.SubStage)
	}
	if row.Park == nil || row.Park.Reason != "pending-effect" || row.Park.Resume != "after readback" {
		t.Errorf("park = %+v", row.Park)
	}
	if row.Gate == nil || row.Gate.Reason != "check-not-passed" || row.Gate.Head != head {
		t.Errorf("gate = %+v", row.Gate)
	}
	if row.Refusal == nil || row.Refusal.Reason != "not-issue-author" {
		t.Errorf("refusal = %+v", row.Refusal)
	}
	if row.State != "implementing" {
		t.Errorf("legacy state must stay for display, got %q", row.State)
	}
}

func TestStageOfFallsBackToLegacyWhenTheRecordCarriesNone(t *testing.T) {
	for name, tc := range map[string]struct {
		rec  claim.Record
		want core.Stage
	}{
		"explicit stage wins": {claim.Record{State: core.StateImplementing, Stage: core.StageReviewing}, core.StageReviewing},
		"implementing":        {claim.Record{State: core.StateImplementing}, core.StageBuilding},
		"merged landing":      {claim.Record{State: core.StateLanding, Merged: true}, core.StageMerging},
		"done":                {claim.Record{State: core.StateDone}, core.StageDone},
	} {
		if got := StageOf(tc.rec); got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
	if got := StageOf(claim.Record{State: "unheard-of"}); got != "" {
		t.Errorf("an unmappable record must give no stage, got %q", got)
	}
}

func TestRecordedFromCarriesTheStage(t *testing.T) {
	got := RecordedFrom(claim.Record{Issue: 7, State: core.StateLanding, Merged: true}, false, "")
	if got.Stage != core.StageMerging {
		t.Errorf("stage = %q", got.Stage)
	}
}
