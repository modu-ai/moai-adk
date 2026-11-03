package cli

// N45 (the rest, audit of 46dbcce0d4): the help of the verbs changed by N38 and N39 says what they now do. The
// sentence of `af land` that said the verb "does not let the caller choose" the repository was true only of --repo;
// the repository is also chosen by the working directory and CLAUDE_PROJECT_DIR, and since N38 the ledger refuses a
// repository other than the one it is bound to. The tests read the help text; the behaviour is pinned by the N38 and
// N39 tests of the verbs themselves.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

func afHelpOf(t *testing.T, verb string) string {
	t.Helper()
	out, _, err := afRunBoth(verb, "--help")
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Fields(out), " ") // the wrapping of a description is not the test's concern
}

func TestAFVerbs_N45MoveSaysWhatItAsksTheTrunk(t *testing.T) {
	for verb, wants := range map[string][]string{
		"move":     {mergeq.DefaultTrunk, "DONE", "landed", "REVIEWING"},
		"evidence": {"lease"},
		"claim":    {"--worktree", "evidence and enqueue"},
	} {
		flat := afHelpOf(t, verb)
		for _, want := range wants {
			if !strings.Contains(flat, want) {
				t.Errorf("`moai af %s --help` does not say %q", verb, want)
			}
		}
	}
}
