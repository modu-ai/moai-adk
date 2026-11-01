package mergeq

// Regression tests for the repair after the audit of db0a69698a (card issue-1948): what the failure log does with
// hostile or merely large output, and how it is written. Values are assembled at run time, never written as literals.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// repairNoteSlack is what a log may hold beyond its cap: the note that says how much of the start was dropped.
const repairNoteSlack = 256

// repairMarker is the word that makes core.Redact mask what follows it.
const repairMarker = "pass" + "word"

// N36: StripControls ran before the masking, so a marker inside an escape sequence was removed from the text before
// the masking could see it, and the value that followed it was kept.
func TestFailureLog_AMarkerInsideAnEscapeSequenceDoesNotKeepItsValue(t *testing.T) {
	l := &FailureLog{Dir: t.TempDir()}
	for _, c := range []struct{ name, in string }{
		{"plain_marker", repairMarker + "=" + logSecret + "\n"},
		{"osc_marker_bel", "\x1b]0;" + repairMarker + "=\x07" + logSecret + "\n"},
		{"osc_marker_st", "\x1b]0;" + repairMarker + "=\x1b\\" + logSecret + "\n"},
	} {
		path, err := l.Write("t1", c.name, "gate", []byte(c.in))
		if err != nil {
			t.Fatal(err)
		}
		body, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Fatal(rerr)
		}
		if strings.Contains(string(body), logSecret) {
			t.Errorf("%s: the value is in the log: %q", c.name, body)
		}
	}
}

// N31: the cap was applied before the masking, and the masking makes text longer ("key=1" becomes "key=[REDACTED]"),
// so a log grew to 2.5 times its cap.
func TestFailureLog_TheCapHoldsAfterTheMaskingMadeTheTextLonger(t *testing.T) {
	const capBytes = 8 << 10
	l := &FailureLog{Dir: t.TempDir(), Cap: capBytes}
	for _, line := range []string{"key=1\n", repairMarker + "=a\n"} {
		in := strings.Repeat(line, capBytes/len(line)+1)
		if masked, _ := core.Redact(in); len(masked) <= len(in) {
			t.Fatalf("fixture %q: the masking did not make the text longer (%d -> %d)", line, len(in), len(masked))
		}
		path, err := l.Write("t1", "c-"+strings.TrimSpace(line), "gate", []byte(in))
		if err != nil {
			t.Fatal(err)
		}
		fi, serr := os.Stat(path)
		if serr != nil {
			t.Fatal(serr)
		}
		if fi.Size() > capBytes+repairNoteSlack {
			t.Errorf("input %q repeated: the log holds %d bytes, over the cap %d", line, fi.Size(), capBytes)
		}
	}
}
