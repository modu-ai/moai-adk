package monitor

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestViewJSONIsStableAndCarriesSourceAndConfirmation(t *testing.T) {
	c := collector(
		fakeClaims{rows: map[int]ClaimRow{7: {Project: "p", Issue: 7, Owner: "lane-18", State: "verifying", HeadSHA: head, Epoch: 2}}},
		fakeRemote{checks: []CheckRun{{Name: "build", Conclusion: "success", HeadSHA: head}}},
		fakeLedger{next: "wait"},
	)
	b, err := MarshalView(c.View(t.Context(), []int{7, 8}))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Rows []map[string]json.RawMessage `json:"rows"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("rows = %d", len(got.Rows))
	}
	s := string(b)
	for _, want := range []string{`"source":"claim-store"`, `"confirmed":true`, `"confirmed":false`, `"read_at"`} {
		if !strings.Contains(s, want) {
			t.Errorf("json lacks %s", want)
		}
	}
	b2, _ := MarshalView(c.View(t.Context(), []int{7, 8}))
	if string(b) != string(b2) {
		t.Error("output not deterministic")
	}
}

func TestViewJSONNeverLeaksSecret(t *testing.T) {
	c := collector(fakeClaims{rows: map[int]ClaimRow{7: {Issue: 7, ResumeCondition: "GITHUB_TOKEN=abc123def456ghi789"}}}, fakeRemote{}, fakeLedger{})
	b, _ := MarshalView(c.View(t.Context(), []int{7}))
	if strings.Contains(string(b), "abc123def456ghi789") {
		t.Errorf("secret in json: %s", b)
	}
}
