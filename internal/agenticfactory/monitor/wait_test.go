package monitor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func sampleWait() *WaitInfo {
	return &WaitInfo{
		Owner: "agent-3", NextAction: "re-run the failed check", RecheckAt: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC),
		Diagnosis: "CI red on the latest head", Retries: 1, MaxRetries: 3, EscalateTo: "leader",
	}
}

// The wait record is shown beside Park and Gate and moves nothing: stage, owner and the other overlays stay.
func TestTheWaitRecordIsShownAndDoesNotMoveTheStageOrTheOwner(t *testing.T) {
	r := stageRow(ClaimRow{
		Owner: "agent-3", Stage: string(core.StageReviewing),
		Park: &ParkInfo{Reason: "rate-limited", Evidence: "429", Resume: "after the reset"},
		Wait: sampleWait(),
	})
	w := r.Wait.Value
	if w == nil || w.Owner != "agent-3" || w.NextAction != "re-run the failed check" || w.Retries != 1 || w.MaxRetries != 3 ||
		w.EscalateTo != "leader" || w.Escalated || w.Diagnosis == "" || w.RecheckAt.IsZero() {
		t.Fatalf("wait = %+v", w)
	}
	if r.Stage.Value != "REVIEWING" || r.Owner.Value != "agent-3" || r.Park.Value == nil {
		t.Errorf("stage=%q owner=%q park=%v: a wait must not displace them", r.Stage.Value, r.Owner.Value, r.Park.Value)
	}
	if !r.Wait.Confirmed || r.Wait.Source != SrcClaimStore {
		t.Errorf("source=%q confirmed=%v", r.Wait.Source, r.Wait.Confirmed)
	}
	if none := stageRow(ClaimRow{Stage: string(core.StageReviewing)}); none.Wait.Value != nil {
		t.Errorf("a record with no wait shows one: %+v", none.Wait.Value)
	}
}

func TestAnEscalatedWaitIsShownAsEscalated(t *testing.T) {
	w := sampleWait()
	w.Retries, w.Escalated = 3, true
	if got := stageRow(ClaimRow{Wait: w}).Wait.Value; got == nil || !got.Escalated || got.Retries != got.MaxRetries {
		t.Errorf("wait = %+v", got)
	}
}

func TestWaitTextIsMaskedLikeOtherClaimText(t *testing.T) {
	secret := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	w := sampleWait()
	w.Diagnosis = "host answered with token " + secret
	w.NextAction = "retry with " + secret
	got := stageRow(ClaimRow{Wait: w}).Wait.Value
	for _, text := range []string{got.Diagnosis, got.NextAction} {
		if strings.Contains(text, secret) {
			t.Errorf("secret survived in %q", text)
		}
	}
}

type failingClaims struct{}

func (failingClaims) ReadClaim(context.Context, int) (ClaimRow, error) {
	return ClaimRow{}, errors.New("store unavailable")
}

func TestAnUnreadableClaimLeavesTheWaitUnconfirmed(t *testing.T) {
	r := collector(failingClaims{}, fakeRemote{}, fakeLedger{}).Collect(context.Background(), []int{7})[0]
	if r.Wait.Confirmed || r.Wait.Value != nil || r.Wait.Source != SrcClaimStore {
		t.Errorf("wait = %+v", r.Wait)
	}
}

// The key names are the contract with the web view: wait.value.{owner,next_action,recheck_at,diagnosis,
// retries,max_retries,escalate_to,escalated}.
func TestTheWaitJSONKeysAreStable(t *testing.T) {
	raw, err := json.Marshal(stageRow(ClaimRow{Wait: sampleWait()}))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	wait, _ := m["wait"].(map[string]any)
	val, ok := wait["value"].(map[string]any)
	if !ok {
		t.Fatalf("no wait.value in %s", raw)
	}
	for _, k := range []string{"owner", "next_action", "recheck_at", "diagnosis", "retries", "max_retries", "escalate_to", "escalated"} {
		if _, ok := val[k]; !ok {
			t.Errorf("wait.value has no %q: %v", k, val)
		}
	}
	if len(val) != 8 {
		t.Errorf("wait.value has %d keys, want exactly 8: %v", len(val), val)
	}
}
