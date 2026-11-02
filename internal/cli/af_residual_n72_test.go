package cli

// N72 (audit a36ea8c368, the second half; residual classification): the input flags that `moai af add`, `move` and
// `move-repository` record (--title, --stop, --acceptance, --risk, --reason, ...) are stored as given and nothing masks them,
// the way `moai af land` says for its own free-input flags. The help of these three verbs says only "recorded as given" on
// some flags, not that nothing masks the text. The expectation: the help of each says that nothing masks it.

import (
	"strings"
	"testing"
)

func TestAFResidual_N72TheVerbsThatRecordFreeInputSayNothingMasksIt(t *testing.T) {
	for _, verb := range []string{"add", "move"} {
		if flat := afHelpOf(t, verb); !strings.Contains(flat, "Nothing masks") {
			t.Errorf("`moai af %s --help` does not say that nothing masks the text of its free-input flags", verb)
		}
	}
}
