// t63 — retained-key advisory stream-merge tests for `moai update`.
//
// Defect (card t63, measured 2026-08-16 on mink-code): the retained-key
// advisory appended one raw "advisory: retained key ..." line per key to
// stderr (backup/node_merge.go retainedKeySink = os.Stderr) while the update
// progress line was mid-redraw on stdout (tui.ProgressLine cursor control),
// interleaving 49 stray lines into the cursor-controlled progress output.
// The 49 lines carried one line of real information: "49 user settings keys
// preserved".
//
// Fix contract (this file): the advisory renders through the SAME output
// channel as the progress line — ONE summary line by default, the full key
// list only under --verbose (the same verbose ledger recordMergeFallback
// reads via updateVerboseMode). No per-key raw stderr text on the restore
// path.

package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// sampleRetainedRefs is a 3-key fixture spanning two sections, mirroring the
// production shape (section-relative file + dotted key path).
func sampleRetainedRefs() []backup.RetainedKeyRef {
	return []backup.RetainedKeyRef{
		{Section: "design.yaml", Key: "evolution.max_active_learnings"},
		{Section: "llm.yaml", Key: "llm.agent_overrides"},
		{Section: "user.yaml", Key: "user.name"},
	}
}

// Default (no --verbose): exactly ONE line — the summary. None of the
// individual key names may appear; the 49-line spray must collapse to the
// single line of real information, plus the --verbose discovery hint.
func TestRenderRetainedKeyAdvisory_DefaultSingleSummaryLine(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderRetainedKeyAdvisory(&buf, sampleRetainedRefs(), false, tui.LightTheme())
	got := buf.String()

	if n := strings.Count(got, "\n"); n != 1 {
		t.Errorf("default render must emit exactly ONE line, got %d:\n%s", n, got)
	}
	if !strings.Contains(got, "3 user settings key(s) preserved") {
		t.Errorf("summary line must state the preserved count, got:\n%s", got)
	}
	if !strings.Contains(got, "--verbose") {
		t.Errorf("summary line must hint at --verbose for the key list, got:\n%s", got)
	}
	for _, ref := range sampleRetainedRefs() {
		if strings.Contains(got, ref.Key) {
			t.Errorf("default render must NOT list individual key %s, got:\n%s", ref.Key, got)
		}
	}
}

// --verbose: the summary line plus one dim line per key, naming every
// section:key so the full list is expandable without a second run.
func TestRenderRetainedKeyAdvisory_VerboseExpandsKeyList(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderRetainedKeyAdvisory(&buf, sampleRetainedRefs(), true, tui.LightTheme())
	got := buf.String()

	if n := strings.Count(got, "\n"); n != 4 {
		t.Errorf("verbose render must emit summary + 3 key lines (4 total), got %d:\n%s", n, got)
	}
	if !strings.Contains(got, "3 user settings key(s) preserved") {
		t.Errorf("verbose render must keep the summary line, got:\n%s", got)
	}
	for _, ref := range sampleRetainedRefs() {
		want := ref.Section + ": " + ref.Key
		if !strings.Contains(got, want) {
			t.Errorf("verbose render must list %q, got:\n%s", want, got)
		}
	}
}

// Nothing retained → nothing printed (no noise on clean runs).
func TestRenderRetainedKeyAdvisory_EmptySilent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderRetainedKeyAdvisory(&buf, nil, false, tui.LightTheme())
	if buf.Len() != 0 {
		t.Errorf("empty refs must print nothing, got:\n%s", buf.String())
	}
	buf.Reset()
	renderRetainedKeyAdvisory(&buf, nil, true, tui.LightTheme())
	if buf.Len() != 0 {
		t.Errorf("empty refs must print nothing under --verbose too, got:\n%s", buf.String())
	}
}

// NO_COLOR (MonochromeTheme): zero SGR sequences — the advisory must degrade
// the same way every other update render helper does (REQ-TUXIU-041).
func TestRenderRetainedKeyAdvisory_NoColorZeroSGR(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	renderRetainedKeyAdvisory(&buf, sampleRetainedRefs(), true, tui.MonochromeTheme())
	if n := len(sgrColor.FindAllString(buf.String(), -1)); n != 0 {
		t.Errorf("NO_COLOR advisory must emit zero SGR sequences, got %d in:\n%q", n, buf.String())
	}
}
