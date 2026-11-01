package cli

// N38 first bind: the help of every verb that opens the ledger says what an unbound ledger is compared with before the
// first call binds it, so a refusal is not a surprise.

import (
	"strings"
	"testing"
)

func TestAFVerbs_N38FirstBindHelpSaysWhatAnUnboundLedgerIsComparedWith(t *testing.T) {
	for _, verb := range afN38Verbs() {
		flat := afHelpOf(t, verb)
		for _, want := range []string{"not bound yet", "lease worktree", "the ledger file"} {
			if !strings.Contains(flat, want) {
				t.Errorf("`moai af %s --help` does not say %q", verb, want)
			}
		}
	}
}
