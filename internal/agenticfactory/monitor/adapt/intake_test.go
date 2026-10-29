package adapt

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

func TestIntakeFromIntentsMapsStagesAndSkipsOtherKinds(t *testing.T) {
	st := effect.NewMemIntentStore()
	for _, in := range []effect.Intent{
		{Key: "a", Kind: effect.IssueCreate, Issue: 1, Status: effect.StatusIntended},
		{Key: "b", Kind: effect.IssueCreate, Issue: 2, Status: effect.StatusConfirmed},
		{Key: "c", Kind: effect.WorktreePrepare, Issue: 2, Status: effect.StatusConfirmed},
		{Key: "d", Kind: effect.BranchPrepare, Issue: 3, Status: effect.StatusFailed, LastError: "disk full"},
		{Key: "e", Kind: effect.IssueCreate, Issue: 4, Status: effect.StatusParked, LastError: "lookup unsure"},
		{Key: "f", Kind: effect.LedgerComment, Issue: 5, Status: effect.StatusConfirmed},
		{Key: "g", Kind: effect.CheckObserve, Issue: 6, Status: effect.StatusIntended},
	} {
		_ = st.Put(in)
	}
	items, err := (IntakeSource{Store: st}).ReadIntake(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]monitor.IntakeStage{"a": monitor.StageAck, "b": monitor.StageIssueCreated, "c": monitor.StageWTReady, "d": monitor.StageFailed, "e": monitor.StageParked}
	if len(items) != len(want) {
		t.Fatalf("items = %+v", items)
	}
	byIssueStage := map[monitor.IntakeStage]int{}
	for _, it := range items {
		byIssueStage[it.Stage]++
	}
	for _, s := range want {
		if byIssueStage[s] == 0 {
			t.Errorf("stage %s missing in %+v", s, items)
		}
	}
}
