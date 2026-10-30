package web

// context_states_test.go — the two context item states lane-20 added to the
// monitor JSON (1b1aecaa7c): "inaccessible-local" (a local file only) and
// "inaccessible-unpushed" (a local commit not yet pushed). Both come with an empty
// link and availability "inaccessible". The web does not check `state` against a
// closed set — it shows it as written and decides the link only from availability,
// the remote check and the stage — so the values pass through and never open a link.
//
// Defect answer: these tests catch a new state dropped, rewritten or turned into
// a link, and a branch on the state's spelling.

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestContextItemInaccessibleLocalAndUnpushedStates(t *testing.T) {
	a := prApp(t, `{"rows":[{"Issue":9,"context":{"value":[
 {"section":"plan","link":"","state":"inaccessible-local","stage":"prepared","remote_verified":false,"availability":"inaccessible","at":"2026-10-30T01:00:00Z"},
 {"section":"audit","link":"","state":"inaccessible-unpushed","stage":"committed","remote_verified":false,"availability":"inaccessible","at":"2026-10-30T02:00:00Z"},
 {"section":"spec","link":"https://example.test/s","state":"inaccessible-unpushed","stage":"committed","remote_verified":false,"availability":"available","at":"2026-10-30T03:00:00Z"}
],"confirmed":true}}]}`, nil)
	items := a.buildPRTracking(context.Background()).Rows[0].Context.Items
	if len(items) != 3 {
		t.Fatalf("items = %+v", items)
	}
	for i, want := range []string{"inaccessible-local", "inaccessible-unpushed", "inaccessible-unpushed"} {
		if items[i].State != want {
			t.Errorf("item %d state = %q, want it shown as written (%q)", i, items[i].State, want)
		}
		if items[i].LinkOpen {
			t.Errorf("item %d must never open a link", i)
		}
	}
	if items[0].LinkStatus != "inaccessible" || items[1].LinkStatus != "inaccessible" {
		t.Errorf("link status: %q / %q, want inaccessible", items[0].LinkStatus, items[1].LinkStatus)
	}
	// A mismatched record (state says unpushed, availability says available, no remote check)
	// is still not an open link.
	if items[2].LinkStatus != "unverified" {
		t.Errorf("a mismatched item must read as unverified, got %q", items[2].LinkStatus)
	}
	m, _ := a.buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	for _, want := range []string{"inaccessible-local", "inaccessible-unpushed", "link inaccessible"} {
		if !strings.Contains(html, want) {
			t.Errorf("the page must show %q", want)
		}
	}
	if strings.Contains(html, `href="https://example.test/s"`) {
		t.Errorf("an unpushed item's link must not be openable")
	}
}
