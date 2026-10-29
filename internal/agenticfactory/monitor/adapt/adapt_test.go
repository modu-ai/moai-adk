package adapt

import (
	"errors"
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
	s, err := claim.Open(t.TempDir())
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
