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

func TestAFVerbs_N45EveryVerbThatOpensTheLedgerSaysItIsBoundToOneRepository(t *testing.T) {
	for _, verb := range afN38Verbs() {
		flat := afHelpOf(t, verb)
		for _, want := range []string{"is bound to one repository", "CLAUDE_PROJECT_DIR", "refused with the ledger unchanged"} {
			if !strings.Contains(flat, want) {
				t.Errorf("`moai af %s --help` does not say %q", verb, want)
			}
		}
	}
}

func TestAFVerbs_N45TheLandHelpNoLongerPromisesAChoiceItDoesNotMake(t *testing.T) {
	flat := afHelpOf(t, "land")
	if strings.Contains(flat, "so the verb does not let the caller choose") {
		t.Error("`moai af land --help` still says the verb does not let the caller choose the repository; the working directory and CLAUDE_PROJECT_DIR choose it, and the ledger's binding is what refuses a wrong one")
	}
	for _, want := range []string{"working directory", "bound"} {
		if !strings.Contains(flat, want) {
			t.Errorf("`moai af land --help` does not say %q", want)
		}
	}
}

func TestAFVerbs_N45EnqueueAndMoveSayWhatTheyAskTheTrunk(t *testing.T) {
	for verb, wants := range map[string][]string{
		"enqueue":  {mergeq.DefaultTrunk, "landed", "waiting", "cannot answer"},
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
