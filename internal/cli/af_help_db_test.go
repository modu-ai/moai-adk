package cli

// N72 (final audit of a36ea8c368): the --db help of the af verbs said `default: <project dir>/af/ledger.db`. The default is
// under the MoAI home, keyed by the project (homestate.ProjectDir). N73, the part that is help text: the verbs that end a
// card say what DONE guarantees and what it does not (the E6 decision: a candidate that is the trunk's own commit lands as
// already in the trunk).

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// afDBVerbs is every verb of `moai af` that has a --db flag, read from the command tree, with the usage of the flag.
func afDBVerbs(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, c := range newAFCmd().Commands() {
		if f := c.Flags().Lookup("db"); f != nil {
			out[c.Name()] = f.Usage
		}
	}
	if len(out) < 9 {
		t.Fatalf("only %d verbs carry --db (%v); the tree has changed or the lookup is wrong", len(out), out)
	}
	return out
}

func TestAFVerbs_N72EveryDbFlagSaysWhereTheDefaultLedgerReallyIs(t *testing.T) {
	const want = "<MoAI home>/db/<project name>-<hash>/af/ledger.db"
	seen := map[string]bool{}
	for verb, usage := range afDBVerbs(t) {
		if strings.Contains(usage, "<project dir>/af/ledger.db") {
			t.Errorf("`moai af %s --db` still says the default is in the project directory: %q", verb, usage)
		}
		if !strings.Contains(usage, want) {
			t.Errorf("`moai af %s --db` does not say %q: %q", verb, want, usage)
		}
		seen[usage] = true
	}
	if len(seen) != 1 {
		t.Errorf("the verbs say %d different things about --db, want one sentence shared by all: %v", len(seen), seen)
	}
}

// The sentence is true: the default ledger of a project is <MoAI home>/db/<name>-<hash>/af/ledger.db and is not
// <project>/af/ledger.db.
func TestAFVerbs_N72TheDefaultLedgerIsWhereTheHelpSays(t *testing.T) {
	home := filepath.Join(t.TempDir(), ".moai")
	t.Setenv("MOAI_HOME", home)
	project := filepath.Join(t.TempDir(), "my-project")
	t.Setenv("CLAUDE_PROJECT_DIR", project)
	got, err := afLedgerPath("")
	if err != nil {
		t.Fatal(err)
	}
	if got == filepath.Join(project, "af", "ledger.db") {
		t.Fatalf("the default ledger is in the project directory: %s", got)
	}
	re := regexp.MustCompile("^" + regexp.QuoteMeta(filepath.ToSlash(home)) + `/db/my-project-[0-9a-f]{8}/af/ledger\.db$`)
	if !re.MatchString(filepath.ToSlash(got)) {
		t.Errorf("the default ledger %s does not have the shape <MoAI home>/db/<project name>-<hash>/af/ledger.db", got)
	}
}

const afDoneMeaning = "DONE guarantees that the candidate commit landed in the trunk; it does not guarantee the card's work content " +
	"(a candidate that is the trunk's own commit lands as already in the trunk)"

func TestAFVerbs_N73TheHelpOfTheVerbsThatEndACardSaysWhatDoneGuarantees(t *testing.T) {
	for _, verb := range []string{"move", "land"} {
		if flat := afHelpOf(t, verb); !strings.Contains(flat, afDoneMeaning) {
			t.Errorf("`moai af %s --help` does not say: %s", verb, afDoneMeaning)
		}
	}
}
