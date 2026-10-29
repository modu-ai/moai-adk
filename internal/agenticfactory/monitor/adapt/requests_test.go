package adapt

import (
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

func TestRequestSourceShowsCanonicalStagesParkedBlockedAndUnknown(t *testing.T) {
	st := openStore(t)
	pol := claim.Policy{MaxInFlight: 1, MaxParked: 2}
	for _, k := range []string{"req-a", "req-b", "req-c"} {
		if _, _, err := st.PutRequest(k, "hash-"+k, 0, pol); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := st.UpdateRequest("req-a", func(r *claim.Request) error {
		r.Stage, r.Issue = claim.StageIssueConfirmed, 1948
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.UpdateRequest("req-a", func(r *claim.Request) error {
		r.Blocked = &claim.Blocked{Step: claim.StageIssueConfirmed, Detail: "push refused", Retryable: true, Resume: "retry"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.UpdateRequest("req-b", func(r *claim.Request) error {
		r.Blocked = &claim.Blocked{Step: claim.StageAccepted, Detail: "response lost", Unknown: true, Resume: "FindIssue decides"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	before, _ := st.ViewRequests()
	items, err := (RequestSource{Store: st}).ReadIntake(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	after, _ := st.ViewRequests()
	if !reflect.DeepEqual(before, after) {
		t.Error("reading the intake view changed the stored requests")
	}
	byKey := map[string]monitor.IntakeItem{}
	for _, it := range items {
		byKey[it.Key] = it
	}
	if len(items) != 3 {
		t.Fatalf("items = %d", len(items))
	}
	a, b, c := byKey["req-a"], byKey["req-b"], byKey["req-c"]
	if a.Stage != monitor.StageIssueConfirmed || a.State != monitor.StateBlocked || a.Issue != 1948 || a.Block == nil || !a.Block.Retryable {
		t.Errorf("a = %+v", a)
	}
	if b.Stage != monitor.StageAccepted || b.State != monitor.StateEffectUnknown || b.Block == nil || !b.Block.Unknown {
		t.Errorf("b = %+v", b)
	}
	if c.State != monitor.StateParked || c.Parked == "" {
		t.Errorf("c = %+v: a request over the in-flight limit must show as PARKED with its reason", c)
	}
	for _, it := range items {
		if it.Started {
			t.Errorf("%s: started without a claim attempt", it.Key)
		}
	}
}
