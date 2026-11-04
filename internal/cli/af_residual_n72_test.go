package cli

// N72 (audit a36ea8c368, the second half; residual classification): the input flags that `moai af move` records (--reason,
// --actor) are stored as given and nothing masks them, the way `moai af land` says for its own free-input flags. The help
// of the verb says only "recorded as given" on some flags, not that nothing masks the text. The expectation: the help
// says that nothing masks it.

import (
	"strings"
	"testing"
)

func TestAFResidual_N72TheVerbsThatRecordFreeInputSayNothingMasksIt(t *testing.T) {
	for _, verb := range []string{"move"} {
		if flat := afHelpOf(t, verb); !strings.Contains(flat, "Nothing masks") {
			t.Errorf("`moai af %s --help` does not say that nothing masks the text of its free-input flags", verb)
		}
	}
}
