package mergeq

// Regression tests for the repair after the audit of db0a69698a (card issue-1948): what the failure log does with
// hostile or merely large output, and how it is written. Values are assembled at run time, never written as literals.

import (
	"os"
	"strings"
	"testing"
)

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
