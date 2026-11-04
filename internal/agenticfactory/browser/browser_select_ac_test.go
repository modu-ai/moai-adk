package browser

// browser_select_ac_test.go — AC-AF-020 (REQ-AF-021), cases (d), (e) and the
// resolver half of (f), named as acceptance.md §G.2 lists them. The saving of
// the selection belongs to the settings layer (lane 17) and the read-back from
// the config file is tested where the file is read (internal/web).
//
// Defect answer: these tests catch a missing selection that is not read as
// chrome, an uninstalled aside that falls back to another browser, and a
// resolver that rewrites the operator's choice.

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// (d) no selection → chrome, with nothing probed.
func TestBrowserSelect_NoSelectionUsesChrome(t *testing.T) {
	p := &fakeProbe{}
	got, err := Resolve(context.Background(), "", p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Use != Chrome || got.Blocked() || got.Selected != "" {
		t.Fatalf("resolution = %+v, want chrome used and no selection echoed", got)
	}
	if len(p.lookups)+len(p.ran) != 0 {
		t.Fatalf("a default must not probe anything: %v %v", p.lookups, p.ran)
	}
}

// (e) aside selected and not installed → blocked not-installed with a resume
// condition, and nothing else is called: no chrome, no ego-browser.
func TestBrowserSelect_AsideNotInstalledBlocksAndNeverFallsBack(t *testing.T) {
	p := &fakeProbe{
		installed: map[string]bool{"ego-browser": true},
		runs:      map[string]fakeRun{asideCheck: {code: 127, err: errors.New("not found")}},
	}
	got, err := Resolve(context.Background(), "aside", p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Reason != ReasonNotInstalled || !got.Blocked() || got.Use != "" || got.Resume == "" {
		t.Fatalf("resolution = %+v, want a not-installed block with a resume condition and no browser to use", got)
	}
	for _, c := range append(append([]string{}, p.ran...), p.lookups...) {
		if !strings.HasPrefix(c, "aside") {
			t.Fatalf("a block must not touch another browser; saw %q", c)
		}
	}
}

// (f), resolver half: whatever was saved comes back as the same value, for each
// of the three choices, and a block never rewrites it.
func TestBrowserSelect_SelectionIsEchoedUnchanged(t *testing.T) {
	for _, sel := range []string{"egolite", "aside", "chrome"} {
		p := &fakeProbe{} // nothing installed: egolite and aside block, chrome passes
		got, err := Resolve(context.Background(), sel, p)
		if err != nil {
			t.Fatal(err)
		}
		if got.Selected != sel {
			t.Errorf("selected %q came back as %q", sel, got.Selected)
		}
	}
}
