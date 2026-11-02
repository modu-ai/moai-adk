package cli

// Regression tests for the repair after the audit of db0a69698a (card issue-1948), through the real `moai af` verbs,
// a real ledger and real git. Values a write hook refuses are assembled at run time (afD1Words, afD1Value).

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afRepairLand runs `af land` with the gate on the standard queued card and returns the verb's JSON line.
func afRepairLand(t *testing.T, db, repo, gate string) (outcome, reason, stdout string) {
	t.Helper()
	out, _, err := afRunBoth("land", "t1", "--db", db, "--gate", gate)
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
			db, _, cand, st, epoch := afQueuedCard(t)
			gate := afD1Words.Replace(r.gate) + "\n" + tail
			out, errOut, err := afRunBoth("land", "t1", "--db", db, "--gate", gate)
			if err != nil {
				t.Fatalf("the verb failed: %v", err)
			}
			if _, _, eerr := afRunBoth("enqueue", "t1", "--db", db, "--epoch", strconv.FormatInt(epoch, 10), "--candidate", cand); eerr != nil {
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
	db, _, _, st, _ := afQueuedCard(t)
	gate := afD1Words.Replace(`pass=` + afD1Value + ` true`)
	out, errOut, err := afRunBoth("land", "t1", "--db", db, "--gate", gate)
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
