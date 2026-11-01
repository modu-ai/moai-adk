package mergeq

// N44 rest, the mutant P-N31-second-cut-off (audits 46dbcce0d4 and a36ea8c368, residual classification): the last cut of maskTail
// keeps the log within its cap after the second masking. TestFailureLog_TheCapHoldsAfterTheMaskingMadeTheTextLonger makes the
// FIRST masking lengthen the text, and the mutant that removed the last cut survived. Here the masking only recognises the value
// once the control characters inside "key" are stripped, so it is the second masking that makes the text longer.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestFailureLog_ResidualN44TheCapHoldsWhenOnlyTheSecondMaskingMakesTheTextLonger(t *testing.T) {
	const capBytes = 8 << 10
	l := &FailureLog{Dir: t.TempDir(), Cap: capBytes}
	line := "ke\x01y=1\n" // the control character hides the key from the first masking; stripping it shows "key=1"
	in := strings.Repeat(line, capBytes/len(line)+1)
	if masked, _ := core.Redact(in); len(masked) != len(in) {
		t.Fatalf("fixture: the first masking already changes the text (%d -> %d)", len(in), len(masked))
	}
	if masked, _ := core.Redact(StripControls(in)); len(masked) <= len(in) {
		t.Fatalf("fixture: the masking after the stripping does not make the text longer (%d -> %d)", len(in), len(masked))
	}
	path, err := l.Write("t1", "c1", "gate", []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	fi, serr := os.Stat(path)
	if serr != nil {
		t.Fatal(serr)
	}
	if fi.Size() > capBytes+repairNoteSlack {
		t.Errorf("the log holds %d bytes, over the cap %d", fi.Size(), capBytes)
	}
}
