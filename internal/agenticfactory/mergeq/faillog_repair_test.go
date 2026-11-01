package mergeq

// Regression tests for the repair after the audit of db0a69698a (card issue-1948): what the failure log does with
// hostile or merely large output, and how it is written. Values are assembled at run time, never written as literals.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// N31: nothing removed old logs, so five failed candidates left 1.3 MB of output that can hold a secret. The directory
// keeps one total cap, and the oldest files go first.
func TestFailureLog_TheDirectoryKeepsOneTotalCapAndTheOldestFilesGoFirst(t *testing.T) {
	dir := t.TempDir()
	const totalCap = 4 * DefaultFailureLogCap
	now := time.Now()
	var seeded []string
	for i := 0; i < 5; i++ {
		p := filepath.Join(dir, fmt.Sprintf("seed-%d.log", i))
		if err := os.WriteFile(p, bytes.Repeat([]byte("x"), 200<<10), 0o600); err != nil {
			t.Fatal(err)
		}
		mt := now.Add(-time.Duration(5-i) * time.Hour)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
		seeded = append(seeded, p)
	}
	path, err := (&FailureLog{Dir: dir}).Write("t1", "c", "gate", []byte(strings.Repeat("a line of output\n", DefaultFailureLogCap/16)))
	if err != nil {
		t.Fatal(err)
	}
	var sum int64
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if fi, ierr := e.Info(); ierr == nil {
			sum += fi.Size()
		}
	}
	if sum > totalCap {
		t.Errorf("the directory holds %d bytes in %d files, over the total cap %d", sum, len(entries), totalCap)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the log just written is gone: %v", err)
	}
	for i, p := range seeded {
		_, serr := os.Stat(p)
		if gone := serr != nil; gone != (i < 2) {
			t.Errorf("seed %d gone=%v, want gone=%v (the two oldest go first and no more)", i, gone, i < 2)
		}
	}
}
