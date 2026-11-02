package cli

// Regression tests for the repair after the audit of db0a69698a (card issue-1948), through the real `moai af` verbs,
// a real ledger and real git. Values a write hook refuses are assembled at run time (afD1Words, afD1Value).

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
)

// afRepairLand runs `af land` with the gate on the standard queued card and returns the verb's JSON line.
func afRepairLand(t *testing.T, db, repo, gate string) (outcome, reason, stdout string) {
	t.Helper()
	out, _, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate)
	if err != nil {
		t.Fatalf("the verb failed: %v (stdout %.200q)", err, out)
	}
	var res struct{ Outcome, Reason string }
	if jerr := json.Unmarshal([]byte(out), &res); jerr != nil {
		t.Fatalf("stdout %.200q is not the verb's JSON line", out)
	}
	return res.Outcome, res.Reason, out
}

// afGateIDPattern is the short id the reason and the landed note carry in place of the gate text: "gate-" and 12 hex digits.
var afGateIDPattern = regexp.MustCompile(`gate-[0-9a-f]{12}`)

// N29: the gate TEXT used to be quoted in the failure reason, the stdout line, the note the next enqueue copies the
// reason into, and the landed note, with core.Redact as the only protection; a credential written into the command
// line stayed in 6 of 9 shapes. The reason, the stdout, the memo and the landed note now carry a short id of the gate
// text; the text itself is recorded nowhere.
func TestAFLand_TheGateTextIsNotInTheReasonTheStdoutOrTheLedger(t *testing.T) {
	afSkipWithoutSh(t)
	// every command is replaced by a built-in that does nothing, so no shape starts a client or reaches a network
	const tail = ": the-tail-of-the-gate-text; exit 1"
	rows := []struct{ name, gate string }{
		{"assign", `@PW@=` + afD1Value + ` true`},
		{"pass_assign", `pass=` + afD1Value + ` true`},
		{"mysql_p", `: mysql -uroot -p` + afD1Value + ` db`},
		{"flag_space", `: mysqldump --@PW@ ` + afD1Value + ` db`},
		{"token_flag", `: gh --@TK@ ` + afD1Value + ` api`},
		{"url_no_dot", `: psql postgres@UC@` + afD1Value + `@db:5432/x`},
		{"curl_u", `: curl -u user:` + afD1Value + ` https://example.test`},
		{"glpat", `CI_@TK@=@GL@` + afD1Value + ` true`},
		{"plain_env", `SECRET_FOO=` + afD1Value + ` true`},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			db, repo, cand, st, epoch := afQueuedCard(t)
			gate := afD1Words.Replace(r.gate) + "\n" + tail
			out, errOut, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate)
			if err != nil {
				t.Fatalf("the verb failed: %v", err)
			}
			if _, _, eerr := afRunBoth("enqueue", "t1", "--db", db, "--epoch", strconv.FormatInt(epoch, 10), "--candidate", cand, "--repo", repo); eerr != nil {
				t.Fatalf("enqueue after the failure: %v", eerr)
			}
			if strings.Contains(out+errOut, afD1Value) {
				t.Errorf("the value is in the stdout or stderr: %.300q", out)
			}
			if strings.Contains(string(afLedgerBytes(t, db)), afD1Value) {
				t.Error("the value is in the ledger file")
			}
			notes, nerr := st.Notes(context.Background(), 0)
			if nerr != nil {
				t.Fatal(nerr)
			}
			for _, n := range notes {
				if strings.Contains(n.Text, afD1Value) {
					t.Errorf("the value is in a note: %.200q", n.Text)
				}
			}
			var res struct{ Outcome, Reason string }
			if jerr := json.Unmarshal([]byte(out), &res); jerr != nil || res.Outcome != "failed" {
				t.Fatalf("stdout %.200q: want outcome failed", out)
			}
			if !afGateIDPattern.MatchString(res.Reason) {
				t.Errorf("the reason carries no gate id: %.300q", res.Reason)
			}
			if strings.Contains(res.Reason, "the-tail-of-the-gate-text") {
				t.Errorf("the gate text is in the reason: %.300q", res.Reason)
			}
		})
	}
}

// N29: the same on a gate that passes: the landed note carries the id, not the text.
func TestAFLand_TheLandedNoteCarriesTheGateIDAndNotTheGateText(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, _, st, _ := afQueuedCard(t)
	gate := afD1Words.Replace(`pass=` + afD1Value + ` true`)
	out, errOut, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--trunk", "refs/heads/af/trunk", "--gate", gate)
	if err != nil {
		t.Fatalf("the verb failed: %v", err)
	}
	if strings.Contains(out+errOut, afD1Value) || strings.Contains(string(afLedgerBytes(t, db)), afD1Value) {
		t.Error("the gate text's value is in the stdout or the ledger")
	}
	notes, nerr := st.Notes(context.Background(), 0)
	if nerr != nil {
		t.Fatal(nerr)
	}
	landed := false
	for _, n := range notes {
		if n.Kind == "landed" {
			landed = true
			if !afGateIDPattern.MatchString(n.Text) {
				t.Errorf("the landed note carries no gate id: %.300q", n.Text)
			}
		}
	}
	if !landed {
		t.Errorf("no landed note: %.300q", out)
	}
}

// afLedgerSnapshot is everything the verbs read back from the ledger, to compare before and after a call that must
// change nothing.
type afLedgerSnapshot struct {
	Queue    []ledger.QueueItem
	Cards    []ledger.Card
	Notes    []ledger.Note
	Evidence []ledger.EvidenceRecord
	Leases   []ledger.Lease
}

func afSnapshot(t *testing.T, st *ledger.Store) afLedgerSnapshot {
	t.Helper()
	ctx := context.Background()
	var s afLedgerSnapshot
	var err error
	if s.Queue, err = st.Queue(ctx); err != nil {
		t.Fatal(err)
	}
	if s.Cards, err = st.Cards(ctx); err != nil {
		t.Fatal(err)
	}
	if s.Notes, err = st.Notes(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if s.Evidence, err = st.EvidenceFor(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	if s.Leases, err = st.Leases(ctx); err != nil {
		t.Fatal(err)
	}
	return s
}

// N26: --trunk named another branch (the candidate's own, say) and the card ended DONE with the real trunk empty: the
// landing row says "merged" without asking which trunk. The trunk is the dedicated ref and cannot be changed; any
// other value is refused before the ledger is opened, for land and reconcile alike.
func TestAFLand_ATrunkOtherThanTheDedicatedOneIsRefusedBeforeLandingAndTheLedgerStaysAsItWas(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, cand, st, _ := afQueuedCard(t)
	before := afSnapshot(t, st)
	trunkBefore := afGit(t, repo, "rev-parse", mergeq.DefaultTrunk)
	for _, verb := range [][]string{{"land", "t1", "--gate", "true"}, {"reconcile"}} {
		for _, trunk := range []string{"refs/heads/c1", "refs/heads/af/other", "refs/heads/trunk", "af/trunk", "refs/heads/main", ""} {
			args := append(append([]string{}, verb...), "--db", db, "--repo", repo, "--trunk", trunk)
			out, _, err := afRunBoth(args...)
			if err == nil || !strings.Contains(err.Error(), "--trunk") || !strings.Contains(err.Error(), mergeq.DefaultTrunk) {
				t.Errorf("%s --trunk %q: error %v (stdout %.120q), want a refusal that names --trunk and %s", verb[0], trunk, err, out, mergeq.DefaultTrunk)
			}
			if after := afSnapshot(t, st); !reflect.DeepEqual(before, after) {
				t.Fatalf("%s --trunk %q changed the ledger:\nbefore %+v\nafter  %+v", verb[0], trunk, before, after)
			}
		}
	}
	if got := afGit(t, repo, "rev-parse", mergeq.DefaultTrunk); got != trunkBefore {
		t.Errorf("the trunk moved: %s -> %s", trunkBefore, got)
	}
	// the dedicated trunk, named or left out, still lands the card
	out, _, err := afRunBoth("land", "t1", "--db", db, "--repo", repo, "--gate", "true")
	if err != nil || !strings.Contains(out, `"outcome":"landed"`) {
		t.Errorf("land with the default trunk: %v %.200q, want landed", err, out)
	}
	if err := exec.Command("git", "-C", repo, "merge-base", "--is-ancestor", cand, mergeq.DefaultTrunk).Run(); err != nil {
		t.Errorf("the candidate is not in the dedicated trunk after the landing: %v", err)
	}
}

// N26: --repo named another repository: the evidence, the queue entry and the landing were then made against a
// repository the project does not own, and the card ended DONE with the project's repository without the commit.
// --repo accepts a path whose git common directory is the project directory's (a linked worktree of the same
// repository qualifies); anything else is refused before the ledger is opened.
func TestAFVerbs_ARepoThatIsNotTheProjectsRepositoryIsRefusedBeforeAnyChange(t *testing.T) {
	afSkipWithoutSh(t)
	db, repo, cand, st, epoch := afQueuedCard(t)
	e := strconv.FormatInt(epoch, 10)

	other := t.TempDir()
	afGit(t, other, "init", "-q", "-b", "work")
	if err := os.WriteFile(filepath.Join(other, "base.txt"), []byte("other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	afGit(t, other, "add", "-A")
	afGit(t, other, "commit", "-q", "-m", "other base")
	foreign := afGit(t, other, "rev-parse", "HEAD")
	notARepo := t.TempDir()

	before := afSnapshot(t, st)
	for _, repoFlag := range []string{other, notARepo, filepath.Join(t.TempDir(), "missing")} {
		for _, verb := range [][]string{
			{"evidence", "t1", "--epoch", e, "--kind", "gate", "--subject-sha", foreign, "--result", "pass"},
			{"enqueue", "t1", "--epoch", e, "--candidate", foreign},
			{"land", "t1", "--gate", "true"},
			{"reconcile"},
		} {
			args := append(append([]string{}, verb...), "--db", db, "--repo", repoFlag)
			out, _, err := afRunBoth(args...)
			if err == nil || !strings.Contains(err.Error(), "--repo") {
				t.Errorf("%s --repo %s: error %v (stdout %.120q), want a refusal that names --repo", verb[0], repoFlag, err, out)
			}
			if after := afSnapshot(t, st); !reflect.DeepEqual(before, after) {
				t.Fatalf("%s --repo %s changed the ledger:\nbefore %+v\nafter  %+v", verb[0], repoFlag, before, after)
			}
		}
	}

	// the project's own repository, a linked worktree of it and a directory inside it are accepted
	wt := filepath.Join(t.TempDir(), "linked")
	afGit(t, repo, "worktree", "add", "-q", "-b", "linked-work", wt, cand)
	inside := filepath.Join(repo, "sub")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, repoFlag := range []string{repo, wt, inside} {
		out, errOut, err := afRunBoth("evidence", "t1", "--db", db, "--epoch", e, "--kind", "review", "--subject-sha", cand, "--result", "pass", "--repo", repoFlag)
		if err != nil {
			t.Errorf("evidence --repo %s refused: %v %s %s", repoFlag, err, out, errOut)
		}
	}
}

// N34: the help of `moai af land` said that a failing gate fails the card "with the tail of its output", which has not
// been true since the output left the reason, and did not say that the free-input flags are recorded as given.
func TestAFLand_TheHelpSaysWhatIsRecordedAndWhatIsRecordedAsGiven(t *testing.T) {
	out, _, err := afRunBoth("land", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"gate-", // the short id of the gate text stands in the reason
		"never carry the gate text",
		"recorded as given", // the free-input flags
		"--command", "--reason", "--path", "--actor",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("`moai af land --help` does not say %q", want)
		}
	}
	for _, stale := range []string{"with the tail of its output", "recorded in the landed note"} {
		if strings.Contains(out, stale) {
			t.Errorf("`moai af land --help` still says %q", stale)
		}
	}
}

// N26: the help says that the trunk cannot be changed and which repositories --repo accepts (this replaces the
// warning about a mis-specified --trunk or --repo that the audit asked for before the verbs refused them).
func TestAFVerbs_TheHelpSaysTheTrunkCannotBeChangedAndWhichRepositoriesAreAccepted(t *testing.T) {
	for _, c := range []struct {
		verb  string
		wants []string
	}{
		{"land", []string{mergeq.DefaultTrunk, "cannot be changed", "git directory", "linked worktree"}},
		{"reconcile", []string{mergeq.DefaultTrunk, "cannot be changed", "git directory", "linked worktree"}},
		{"evidence", []string{"project", "git directory", "linked worktree"}},
		{"enqueue", []string{"project", "git directory", "linked worktree"}},
	} {
		out, _, err := afRunBoth(c.verb, "--help")
		if err != nil {
			t.Fatal(err)
		}
		flat := strings.Join(strings.Fields(out), " ") // the wrapping of a flag's description is not the test's concern
		for _, want := range c.wants {
			if !strings.Contains(flat, want) {
				t.Errorf("`moai af %s --help` does not say %q", c.verb, want)
			}
		}
	}
}

// N34: the verbs that take free text say that it is recorded as given.
func TestAFVerbs_TheFreeInputFlagsSayThatTheyAreRecordedAsGiven(t *testing.T) {
	for verb, flags := range map[string][]string{"evidence": {"--command", "--path"}, "move": {"--reason", "--actor"}} {
		out, _, err := afRunBoth(verb, "--help")
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range flags {
			i := strings.Index(out, f+" ")
			if i < 0 {
				t.Fatalf("`moai af %s --help` lists no %s", verb, f)
			}
			line := out[i:]
			if j := strings.Index(line, "\n"); j >= 0 {
				line = line[:j]
			}
			if !strings.Contains(line, "recorded as given") {
				t.Errorf("`moai af %s --help`: the line of %s does not say it is recorded as given: %q", verb, f, line)
			}
		}
	}
}
