package web

// merged_states_test.go — two states that arrived with the lane-18/20 merge, as
// the web shows them (Issue #1948, B11). The web adds no display for a state
// that no screen carries; these tests pin what is shown and what must not be.
//
// Defect answer: these tests catch a parked-for-sensitive-content record that
// leaks its text or digest onto the page, and a mirror whose ledger was refused
// (not sent because it holds a secret or personal data) that is shown with its
// cached position as if it were current.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// A parked intake item shows its reason and nothing else of the refused request:
// the decoder reads named keys only, so a payload, a text or a digest that rides
// along in the record never reaches the page, the view model or the JSON API.
func TestParkedSensitiveContentShowsOnlyTheReason(t *testing.T) {
	const digest = "sha256-d34db33fd34db33fd34db33fd34db33f"
	const text = "the refused text itself"
	rec := `{"key":"k","stage":"ACCEPTED","parked":"sensitive-content","digest":"` + digest + `","text":"` + text + `","payload":{"body":"` + text + `"},"last_error":"found email"}`
	a := newApp(Config{ProjectRoot: t.TempDir()})
	a.intakeRequests = func(context.Context) ([]byte, error) { return []byte("[" + rec + "]"), nil }
	vm := a.buildIntakeRequests(context.Background())
	if len(vm.Rows) != 1 || vm.Rows[0].Status != "PARKED" || vm.Rows[0].Parked != "sensitive-content" || vm.ParkedCount != 1 {
		t.Fatalf("row = %+v (parked %d)", vm.Rows, vm.ParkedCount)
	}
	m, _ := a.buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	if !strings.Contains(html, "sensitive-content") {
		t.Errorf("the reason must be shown")
	}
	for _, leak := range []string{digest, text, "d34db33f", "found email"} {
		if strings.Contains(html, leak) {
			t.Errorf("%q reached the page", leak)
		}
	}
	// The same through the monitor's intake item.
	b := prApp(t, `{"rows":[],"intake":{"value":[{"key":"k","stage":"ACCEPTED","parked":"sensitive-content","digest":"`+digest+`"}],"confirmed":true}}`, nil)
	row := b.buildPRTracking(context.Background()).Intake.Rows
	if len(row) != 1 || row[0].Status != "PARKED" || row[0].Parked != "sensitive-content" {
		t.Errorf("monitor intake item = %+v", row)
	}
}

// mirror_state "refused" (the ledger holds a secret or personal data and was not
// sent) is shown as such, and it confirms nothing the cache carries — the same
// treatment as stale and blocked(...). Identity stays shown.
func TestMirrorRefusedIsShownAndConfirmsNothing(t *testing.T) {
	sha := core.HeadSHA(strings.Repeat("a", 40))
	src := fakeRecords{recs: []claim.Record{
		{Issue: 61, Holder: "s-61", AgentName: "agent-r", State: core.StateLanding, Stage: core.StageMerging, Head: sha, Mirror: "refused",
			Receipt: &claim.Receipt{HandoffID: "h", PR: 601, Head: sha}},
		{Issue: 62, Holder: "s-62", AgentName: "agent-c", State: core.StateLanding, Stage: core.StageMerging, Head: sha, Mirror: "current",
			Receipt: &claim.Receipt{HandoffID: "h", PR: 602, Head: sha}},
	}}
	a := recordApp(t, src)
	by := map[int]PRTrackRowVM{}
	for _, r := range a.buildPRTracking(context.Background()).Rows {
		by[r.Issue] = r
	}
	r := by[61]
	if r.Mirror != "refused" {
		t.Errorf("mirror = %q, want refused", r.Mirror)
	}
	if r.State != prGap || r.Stage != prGap || r.Head != prGap || r.PR != prGap || r.Completed {
		t.Errorf("a refused mirror must not present the cached position as confirmed: %+v", r)
	}
	if r.Agent != "agent-r" {
		t.Errorf("identity stays shown: %q", r.Agent)
	}
	if c := by[62]; c.Mirror != "current" || c.PR != "#602" || c.Stage != "MERGING" {
		t.Errorf("the control row (current mirror) must keep its position: %+v", c)
	}
	m, _ := a.buildMonitor(time.Now())
	if html := renderTempl(t, Monitor(tg1ShellVM(), m)); !strings.Contains(html, "mirror refused") {
		t.Errorf("the page must show 'mirror refused'")
	}
}

// mirror_state rejected(<status>) (lane-18 7ef22d33b1: the host refused the
// Issue's comment) is not "current" either: it is shown as written and, like
// stale, blocked and refused, confirms none of the cached position.
func TestMirrorRejectedIsShownAndConfirmsNothing(t *testing.T) {
	sha := core.HeadSHA(strings.Repeat("a", 40))
	src := fakeRecords{recs: []claim.Record{
		{Issue: 63, Holder: "s-63", AgentName: "agent-j", State: core.StateLanding, Stage: core.StageMerging, Head: sha, Mirror: "rejected(404)",
			Receipt: &claim.Receipt{HandoffID: "h", PR: 603, Head: sha}},
	}}
	r := recordApp(t, src).buildPRTracking(context.Background()).Rows[0]
	if r.Mirror != "rejected(404)" {
		t.Errorf("mirror = %q, want rejected(404)", r.Mirror)
	}
	if r.State != prGap || r.Stage != prGap || r.Head != prGap || r.PR != prGap || r.Completed {
		t.Errorf("a rejected mirror must not present the cached position as confirmed: %+v", r)
	}
	if r.Agent != "agent-j" {
		t.Errorf("identity stays shown: %q", r.Agent)
	}
}

// lane-18 61e5b492c6: an Issue that answers 403 unclassified no longer holds the
// Issues behind it in Flush; its mirror_state becomes blocked(unclassified-403)
// while the Issues behind it are written and read "current". The web shows the
// blocked Issue with that reason and does not read its cached position as
// confirmed; the Issues behind it display normally.
func TestMirrorUnclassified403BlocksOnlyItsOwnIssue(t *testing.T) {
	sha := core.HeadSHA(strings.Repeat("a", 40))
	mk := func(issue int, mirror string) claim.Record {
		return claim.Record{Issue: issue, Holder: "s-" + itoa(issue), AgentName: "agent-" + itoa(issue), State: core.StateLanding,
			Stage: core.StageMerging, Head: sha, Mirror: mirror, Receipt: &claim.Receipt{HandoffID: "h", PR: 700 + issue, Head: sha}}
	}
	src := fakeRecords{recs: []claim.Record{
		mk(71, "blocked(unclassified-403)"), // the locked Issue
		mk(72, "current"),                   // behind it, written by the same Flush
		mk(73, "current"),
	}}
	a := recordApp(t, src)
	by := map[int]PRTrackRowVM{}
	for _, r := range a.buildPRTracking(context.Background()).Rows {
		by[r.Issue] = r
	}
	locked := by[71]
	if locked.Mirror != "blocked(unclassified-403)" {
		t.Errorf("the blocked Issue must show its reason: %q", locked.Mirror)
	}
	if locked.State != prGap || locked.Stage != prGap || locked.Head != prGap || locked.PR != prGap || locked.Completed {
		t.Errorf("a blocked Issue's cached position must not read as confirmed: %+v", locked)
	}
	for _, n := range []int{72, 73} {
		r := by[n]
		if r.Mirror != "current" || r.PR != "#"+itoa(700+n) || r.Stage != "MERGING" || r.State != "landing" || r.Head != "aaaaaaaa" {
			t.Errorf("Issue %d behind the locked one must display normally: %+v", n, r)
		}
	}
	m, _ := a.buildMonitor(time.Now())
	if html := renderTempl(t, Monitor(tg1ShellVM(), m)); !strings.Contains(html, "mirror blocked(unclassified-403)") {
		t.Errorf("the page must show 'mirror blocked(unclassified-403)'")
	}
}
