// todo_landed_test.go — SPEC-TODO-LANDING-EVIDENCE-001 (card t359) M3:
// AC-TLE-004, AC-TLE-007, AC-TLE-008, AC-TLE-009, AC-TLE-010, AC-TLE-012,
// AC-TLE-020.
//
// The recording verb is the first WRITER of landing evidence, and three of
// these criteria exist because the plausible slips are all silent:
//
//	AC-TLE-004/008 — a write that also transitions or reorders a card. Both
//	       are asserted BEHAVIOURALLY over the whole queue, never by grepping
//	       for a mutating call: half A recorded a name-based sweep being
//	       satisfied by a mutant.
//	AC-TLE-012 — a delivering SHA filled from the landed grep's first match.
//	       The fixture below places THREE commits naming the card below the
//	       head precisely so that leak has something to leak.
//	AC-TLE-020 — the attribution boundary. Asserted by recording ONE `--sha`
//	       against TWO card ids, because the predicate this guards against
//	       keys on the id and a text rename varies the wrong variable. The
//	       fixture history names exactly one of the two ids, which is the
//	       premise that makes the clause able to fail at all.
package cli

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// landedFixture is a project root whose git history is PINNED for the
// attribution criteria, plus the SHAs those criteria name.
type landedFixture struct {
	root  string
	store *kanban.BacklogStore
	// mentioning are the three commits whose messages name card t1. They sit
	// BELOW the head, so none of them is the ref position — which is what
	// lets AC-TLE-012 assert a containment set that excludes ref_head by
	// construction rather than by exception.
	mentioning []string
	// head is the ref position: the tip commit, naming no card.
	head string
	// unreachable exists but is not an ancestor of head — a commit built as a
	// CHILD of the tip with commit-tree, so no checkout is needed and the
	// fixture's branch state never moves.
	unreachable string
}

// newLandedFixture builds the pinned history. `--ref HEAD` is what the tests
// pass, so the fixture never depends on the compiled-in default branch name.
func newLandedFixture(t *testing.T) *landedFixture {
	t.Helper()
	root, store := todoFixture(t)
	f := &landedFixture{root: root, store: store}

	// Three commits naming t1, then a tip naming nothing. The card token in
	// the message is the whole point: it is what the landed grep would match
	// if the verb ever consulted it.
	for _, msg := range []string{
		"fix(kanban): first step of t1",
		"feat(kanban): second step of t1",
		"docs: third mention of t1",
	} {
		runGitIn(t, root, "commit", "--allow-empty", "-q", "-m", msg)
		f.mentioning = append(f.mentioning, gitOut(t, root, "rev-parse", "HEAD"))
	}
	runGitIn(t, root, "commit", "--allow-empty", "-q", "-m", "chore: tip commit naming no card")
	f.head = gitOut(t, root, "rev-parse", "HEAD")

	// A commit that EXISTS but is unreachable from HEAD: built as a child of
	// the tip, so `merge-base --is-ancestor <it> HEAD` is false while
	// `rev-parse --verify <it>^{commit}` succeeds.
	tree := gitOut(t, root, "rev-parse", "HEAD^{tree}")
	f.unreachable = gitOut(t, root, "commit-tree", tree, "-p", f.head, "-m", "side branch, never merged")
	return f
}

// gitOut runs git in dir and returns its trimmed stdout.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", full, err)
	}
	return strings.TrimSpace(string(out))
}

// landingOf returns the decoded record for id, and whether one is present.
func landingOf(t *testing.T, store *kanban.BacklogStore, id string) (kanban.LandingEvidence, bool) {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		if it.ID == id {
			if it.Landing == nil {
				return kanban.LandingEvidence{}, false
			}
			return *it.Landing, true
		}
	}
	t.Fatalf("no card %s in the queue", id)
	return kanban.LandingEvidence{}, false
}

// openQueueDB opens the engine database read-only for the SQL-level
// assertions the criteria name (`SELECT landing IS NULL`, the stored DDL).
// Asking Go's decoded view would answer a different question: the criteria
// are about what is IN THE COLUMN, and a nil pointer can be produced by a
// read path that never looked.
func openQueueDB(t *testing.T, store *kanban.BacklogStore) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", store.EnginePath())
	if err != nil {
		t.Fatalf("open queue db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// landingIsNULL answers the criteria's own predicate.
func landingIsNULL(t *testing.T, store *kanban.BacklogStore, id string) int {
	t.Helper()
	var n int
	if err := openQueueDB(t, store).QueryRow(
		`SELECT landing IS NULL FROM items WHERE id = ?`, id).Scan(&n); err != nil {
		t.Fatalf("SELECT landing IS NULL for %s: %v", id, err)
	}
	return n
}

// storedLanding returns the raw column text, and whether it is non-NULL.
func storedLanding(t *testing.T, store *kanban.BacklogStore, id string) (string, bool) {
	t.Helper()
	var v sql.NullString
	if err := openQueueDB(t, store).QueryRow(
		`SELECT landing FROM items WHERE id = ?`, id).Scan(&v); err != nil {
		t.Fatalf("SELECT landing for %s: %v", id, err)
	}
	return v.String, v.Valid
}

// queueTuples renders the ordered (id, state, position, text, spec_id) tuple
// list AC-TLE-008 compares. Position is the 1-based render order.
func queueTuples(t *testing.T, store *kanban.BacklogStore) []string {
	t.Helper()
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	tuples := make([]string, 0, len(rec.Items))
	for i, it := range rec.Items {
		spec := "<nil>"
		if it.SpecID != nil {
			spec = *it.SpecID
		}
		tuples = append(tuples, fmt.Sprintf("%s|%s|%d|%s|%s", it.ID, it.State, i+1, it.Text, spec))
	}
	return tuples
}

// setState marks a card without going through a verb, so a criterion about
// what the LANDED verb does is not entangled with what `next` does.
func setState(t *testing.T, store *kanban.BacklogStore, id string, state kanban.BacklogState) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				rec.Items[i].State = state
				return nil
			}
		}
		return fmt.Errorf("no card %s", id)
	}); err != nil {
		t.Fatalf("set state of %s: %v", id, err)
	}
}

// attachSpec attaches a spec id to a card the same way.
func attachSpec(t *testing.T, store *kanban.BacklogStore, id, specID string) {
	t.Helper()
	if err := store.Mutate(func(rec *kanban.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == id {
				v := specID
				rec.Items[i].SpecID = &v
				return nil
			}
		}
		return fmt.Errorf("no card %s", id)
	}); err != nil {
		t.Fatalf("attach spec to %s: %v", id, err)
	}
}

// writeSpecStatus writes a fixture SPEC document carrying a frontmatter
// status.
func writeSpecStatus(t *testing.T, root, specID, status string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	doc := "---\nid: " + specID + "\nstatus: " + status + "\n---\n\n# fixture\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(doc), 0o600); err != nil {
		t.Fatalf("write fixture spec: %v", err)
	}
}

// AC-TLE-007 — the verb writes ONE record, under the lock, and asks nothing.
//
// The concurrency half is the load-bearing one. `BacklogStore.Mutate` is a
// whole-record read-modify-write, so two UNLOCKED concurrent writes on
// DIFFERENT cards genuinely drop one: the second reader loads the record the
// first had not yet written back, and its write erases the first's record.
// A write placed outside the lock therefore loses a record here and nowhere
// else in this file.
func TestTodoLanded_OneRecordUnderTheLock(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "first card", "second card", "third card")

	out, _, err := runTodo(t, "landed", "1", "--ref", "HEAD", "--sha", f.mentioning[0])
	if err != nil {
		t.Fatalf("landed t1: %v", err)
	}
	if !strings.Contains(out, "t1") {
		t.Errorf("stdout = %q, want it to name the card", out)
	}

	// Exactly one card carries a record.
	var nonNull int
	if err := openQueueDB(t, f.store).QueryRow(
		`SELECT COUNT(*) FROM items WHERE landing IS NOT NULL`).Scan(&nonNull); err != nil {
		t.Fatalf("count non-null landing: %v", err)
	}
	if nonNull != 1 {
		t.Errorf("%d cards carry a landing record, want exactly 1", nonNull)
	}

	// Two concurrent records on DIFFERENT cards: both must survive.
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, id := range []string{"2", "3"} {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			errs[i] = runTodoConcurrent("landed", id, "--ref", "HEAD")
		}(i, id)
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			t.Fatalf("concurrent landed #%d: %v", i, e)
		}
	}
	for _, id := range []string{"t1", "t2", "t3"} {
		if got := landingIsNULL(t, f.store, id); got != 0 {
			t.Errorf("%s: SELECT landing IS NULL = %d after a concurrent pair, want 0 "+
				"(a lost record means the write ran outside the queue lock)", id, got)
		}
	}
}

// runTodoConcurrent runs the verb from a goroutine. It takes no *testing.T —
// t.Setenv and t.Fatalf are not goroutine-safe — and relies on the process
// environment todoFixture already established for queue-root resolution.
func runTodoConcurrent(args ...string) error {
	cmd := newTodoCmd()
	var out, errBuf strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// AC-TLE-004 — the state CHECK is untouched and no write moves a card.
func TestTodoLanded_StateCheckAndStatesUntouched(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "queued card", "picked card")
	setState(t, f.store, "t2", kanban.BacklogStatePicked)

	before := map[string]kanban.BacklogState{}
	rec, err := f.store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, it := range rec.Items {
		before[it.ID] = it.State
	}

	for _, id := range []string{"1", "2"} {
		if _, _, err := runTodo(t, "landed", id, "--ref", "HEAD"); err != nil {
			t.Fatalf("landed %s: %v", id, err)
		}
	}

	// The stored DDL still carries the three-state CHECK, verbatim.
	var ddl string
	if err := openQueueDB(t, f.store).QueryRow(
		`SELECT sql FROM sqlite_master WHERE type='table' AND name='items'`).Scan(&ddl); err != nil {
		t.Fatalf("read items DDL: %v", err)
	}
	const wantCheck = `CHECK (state IN ('queued','picked','dropped'))`
	if !strings.Contains(ddl, wantCheck) {
		t.Errorf("items DDL lost %s\nDDL:\n%s", wantCheck, ddl)
	}

	after, err := f.store.Load()
	if err != nil {
		t.Fatalf("load after: %v", err)
	}
	for _, it := range after.Items {
		if it.State != before[it.ID] {
			t.Errorf("card %s state moved %q -> %q across a landing record; "+
				"evidence never transitions a card", it.ID, before[it.ID], it.State)
		}
	}
}

// AC-TLE-008 — the write moves nothing else in the queue.
func TestTodoLanded_WholeQueueUnmoved(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "alpha", "bravo", "charlie", "delta", "echo")
	setState(t, f.store, "t2", kanban.BacklogStatePicked)
	setState(t, f.store, "t4", kanban.BacklogStateDropped)
	attachSpec(t, f.store, "t3", "SPEC-FIXTURE-LANDED-001")
	writeSpecStatus(t, f.root, "SPEC-FIXTURE-LANDED-001", "implemented")

	before := queueTuples(t, f.store)
	if len(before) != 5 {
		t.Fatalf("fixture holds %d cards, want 5", len(before))
	}

	if _, _, err := runTodo(t, "landed", "3", "--ref", "HEAD"); err != nil {
		t.Fatalf("landed t3: %v", err)
	}

	after := queueTuples(t, f.store)
	if len(after) != 5 {
		t.Errorf("queue holds %d cards after the write, want 5", len(after))
	}
	if strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Errorf("the queue moved across a landing record\nbefore:\n%s\nafter:\n%s",
			strings.Join(before, "\n"), strings.Join(after, "\n"))
	}
}

// AC-TLE-009 — re-record replaces; --clear removes.
func TestTodoLanded_ReplaceAndClear(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "only card")
	shaA, shaB := f.mentioning[0], f.mentioning[1]

	if _, _, err := runTodo(t, "landed", "1", "--ref", "HEAD", "--sha", shaA); err != nil {
		t.Fatalf("record A: %v", err)
	}
	if _, _, err := runTodo(t, "landed", "1", "--ref", "HEAD", "--sha", shaB); err != nil {
		t.Fatalf("record B: %v", err)
	}

	raw, present := storedLanding(t, f.store, "t1")
	if !present {
		t.Fatal("no record stored after the second write")
	}
	if strings.Contains(raw, shaA) {
		t.Errorf("stored value still names the replaced SHA %s — the write appended rather than replaced\nstored: %s", shaA, raw)
	}
	if !strings.Contains(raw, shaB) {
		t.Errorf("stored value does not name the recorded SHA %s\nstored: %s", shaB, raw)
	}
	ev, ok := landingOf(t, f.store, "t1")
	if !ok || ev.SHA != shaB {
		t.Errorf("decoded record SHA = %q, want %q (exactly one record)", ev.SHA, shaB)
	}

	if _, _, err := runTodo(t, "landed", "1", "--clear"); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if got := landingIsNULL(t, f.store, "t1"); got != 1 {
		t.Errorf("after --clear, SELECT landing IS NULL = %d, want 1", got)
	}
}

// AC-TLE-009 (usage half) — --clear is mutually exclusive with --sha / --ref,
// and the refusal is a USAGE error (exit 2), not a runtime one.
func TestTodoLanded_ClearIsExclusive(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "only card")

	for _, args := range [][]string{
		{"landed", "1", "--clear", "--sha", f.head},
		{"landed", "1", "--clear", "--ref", "HEAD"},
	} {
		_, _, err := runTodo(t, args...)
		if err == nil {
			t.Fatalf("todo %v succeeded, want a usage refusal", args)
		}
		code, ok := ResolveExitCode(err)
		if !ok || code != 2 {
			t.Errorf("todo %v exit code = %d (resolved %v), want 2", args, code, ok)
		}
		if got := landingIsNULL(t, f.store, "t1"); got != 1 {
			t.Errorf("todo %v wrote a record on a refused invocation", args)
		}
	}
}

// AC-TLE-010 — the SPEC status is read, and an unreadable one is not invented.
func TestTodoLanded_SpecStatusReadNeverInvented(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "card with a spec", "card whose spec is missing")
	attachSpec(t, f.store, "t1", "SPEC-FIXTURE-LANDED-001")
	attachSpec(t, f.store, "t2", "SPEC-FIXTURE-ABSENT-001")
	writeSpecStatus(t, f.root, "SPEC-FIXTURE-LANDED-001", "implemented")

	for _, id := range []string{"1", "2"} {
		if _, _, err := runTodo(t, "landed", id, "--ref", "HEAD"); err != nil {
			t.Fatalf("landed %s: %v", id, err)
		}
	}
	ev1, _ := landingOf(t, f.store, "t1")
	if ev1.SpecStatus != "implemented" {
		t.Errorf("t1 spec_status = %q, want %q", ev1.SpecStatus, "implemented")
	}
	ev2, _ := landingOf(t, f.store, "t2")
	if ev2.SpecStatus != kanban.LandingSpecStatusUnknown {
		t.Errorf("t2 spec_status = %q, want the explicit unknown marker %q — an unreadable status is never defaulted",
			ev2.SpecStatus, kanban.LandingSpecStatusUnknown)
	}

	// Second phase: the fixture's frontmatter changes, and the read must
	// follow it. A hard-coded status passes the first phase and fails here.
	writeSpecStatus(t, f.root, "SPEC-FIXTURE-LANDED-001", "completed")
	if _, _, err := runTodo(t, "landed", "1", "--ref", "HEAD"); err != nil {
		t.Fatalf("re-record t1: %v", err)
	}
	reread, _ := landingOf(t, f.store, "t1")
	if reread.SpecStatus != "completed" {
		t.Errorf("after the frontmatter changed, t1 spec_status = %q, want %q", reread.SpecStatus, "completed")
	}
}

// AC-TLE-012 — a stored SHA is operator-supplied or absent.
//
// The fixture's three t1-naming commits sit BELOW the head, so the
// containment set excludes the ref position by construction: a record that
// legitimately carries ref_head cannot accidentally satisfy the assertion.
func TestTodoLanded_NoSHADerivedFromTheGrep(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "first card")

	if _, _, err := runTodo(t, "landed", "1", "--ref", "HEAD"); err != nil {
		t.Fatalf("landed t1 without --sha: %v", err)
	}

	ev, ok := landingOf(t, f.store, "t1")
	if !ok {
		t.Fatal("no record stored")
	}
	if strings.TrimSpace(ev.SHA) != "" {
		t.Errorf("stored delivering SHA = %q, want empty: no --sha was supplied and the machine has no lawful source", ev.SHA)
	}
	if ev.SHASource == kanban.LandingSHASourceOperator {
		t.Errorf("stored provenance = %q with no operator assertion", ev.SHASource)
	}
	raw, _ := storedLanding(t, f.store, "t1")
	for i, sha := range f.mentioning {
		if strings.Contains(raw, sha) {
			t.Errorf("stored value contains the #%d commit naming the card (%s) — the SHA was derived from the grep predicate\nstored: %s",
				i+1, sha, raw)
		}
		// SEVEN, not nine. git's own default abbreviation in a small
		// repository is 7 characters, so a nine-character probe cannot match
		// what `--oneline` actually emits — the clause was vacuous against
		// the very leak it names, and the planted mutant's stored value
		// ("a18ca73") is what showed it. Seven is git's floor, so a longer
		// abbreviation still contains it as a prefix.
		if short := sha[:7]; strings.Contains(raw, short) {
			t.Errorf("stored value contains the abbreviated #%d matching commit (%s)\nstored: %s", i+1, short, raw)
		}
	}
	// The ref position IS carried, under its own key — otherwise the
	// containment assertion above would pass on an empty record.
	if ev.RefHead != f.head {
		t.Errorf("ref_head = %q, want the fixture head %q", ev.RefHead, f.head)
	}
}

// AC-TLE-020 — a supplied SHA is validated for existence and reachability,
// and the card id reaches neither check.
func TestTodoLanded_SHAValidation(t *testing.T) {
	// (a) reachable — accepted, and the FULL resolved SHA is stored even
	// though the input was abbreviated.
	t.Run("reachable accepts and stores the full SHA", func(t *testing.T) {
		f := newLandedFixture(t)
		seedQueue(t, f.store, "first card")
		abbrev := f.mentioning[1][:9]
		if _, _, err := runTodo(t, "landed", "1", "--ref", "HEAD", "--sha", abbrev); err != nil {
			t.Fatalf("reachable --sha refused: %v", err)
		}
		ev, ok := landingOf(t, f.store, "t1")
		if !ok {
			t.Fatal("no record stored on the accepting branch")
		}
		if ev.SHA != f.mentioning[1] {
			t.Errorf("stored SHA = %q, want the resolved full SHA %q (never the abbreviated input %q)",
				ev.SHA, f.mentioning[1], abbrev)
		}
		if ev.SHASource != kanban.LandingSHASourceOperator {
			t.Errorf("stored provenance = %q, want %q", ev.SHASource, kanban.LandingSHASourceOperator)
		}
	})

	// (b) non-existent, (c) unreachable, (d) unrunnable — each exit 1, each
	// writes nothing, each names DISTINGUISHABLY which check failed.
	t.Run("refusing branches exit 1, write nothing, and are distinguishable", func(t *testing.T) {
		f := newLandedFixture(t)
		seedQueue(t, f.store, "first card")
		cases := []struct {
			name string
			args []string
			want string
		}{
			{"non-existent object", []string{"landed", "1", "--ref", "HEAD", "--sha", "0123456789abcdef0123456789abcdef01234567"}, "existence"},
			{"unreachable commit", []string{"landed", "1", "--ref", "HEAD", "--sha", f.unreachable}, "reachability"},
			{"unresolvable ref", []string{"landed", "1", "--ref", "refs/heads/no-such-ref", "--sha", f.mentioning[0]}, "unrunnable"},
		}
		seen := map[string]string{}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, errOut, err := runTodo(t, tc.args...)
				if err == nil {
					t.Fatalf("todo %v succeeded, want a refusal", tc.args)
				}
				if code, ok := ResolveExitCode(err); ok && code != 1 {
					t.Errorf("exit code = %d, want 1", code)
				}
				if got := landingIsNULL(t, f.store, "t1"); got != 1 {
					t.Errorf("a refused invocation wrote a record (SELECT landing IS NULL = %d, want 1)", got)
				}
				msg := errOut + err.Error()
				if !strings.Contains(msg, tc.want) {
					t.Errorf("refusal does not name the failing check %q\nstderr+err: %s", tc.want, msg)
				}
				for other, otherMsg := range seen {
					if otherMsg == msg {
						t.Errorf("the %q and %q refusals are byte-identical; the classifications must be distinguishable\n%s",
							tc.name, other, msg)
					}
				}
				seen[tc.name] = msg
			})
		}
	})

	// (d, second exercise) — git absent from the resolved command runner.
	t.Run("no git is unrunnable and refuses", func(t *testing.T) {
		f := newLandedFixture(t)
		seedQueue(t, f.store, "first card")
		prev := todoRunCommand
		t.Cleanup(func() { todoRunCommand = prev })
		todoRunCommand = func(name string, args ...string) (string, error) {
			return "", fmt.Errorf("exec: %q: executable file not found in $PATH", name)
		}
		_, errOut, err := runTodo(t, "landed", "1", "--ref", "HEAD", "--sha", f.mentioning[0])
		if err == nil {
			t.Fatal("recording with no git succeeded; a write that cannot validate refuses")
		}
		if !strings.Contains(errOut+err.Error(), "unrunnable") {
			t.Errorf("refusal does not classify as unrunnable\nstderr+err: %s", errOut+err.Error())
		}
		if got := landingIsNULL(t, f.store, "t1"); got != 1 {
			t.Errorf("a record was written with no git available (SELECT landing IS NULL = %d, want 1)", got)
		}
	})

	// The attribution boundary: ONE --sha, TWO card ids, identical outcome
	// and identical classification on BOTH an accepting and a refusing
	// branch. The fixture history names t1 and never t2 — exactly one of the
	// two — which is the premise that lets a card-token leak diverge them.
	t.Run("the card id reaches neither check", func(t *testing.T) {
		f := newLandedFixture(t)
		seedQueue(t, f.store, "card one", "card two")

		// Premise check, asserted rather than assumed: the fixture ref's
		// history mentions exactly one of the two ids. Both degenerate cases
		// (neither / both) make the clause unable to fail.
		mentions := func(card string) bool {
			out := gitOut(t, f.root, "log", "HEAD", "--perl-regexp", `--grep=\b`+card+`\b`, "--oneline")
			return strings.TrimSpace(out) != ""
		}
		if !mentions("t1") || mentions("t2") {
			t.Fatalf("fixture premise broken: history must name exactly one of t1/t2 (t1=%v, t2=%v)",
				mentions("t1"), mentions("t2"))
		}

		type outcome struct {
			failed bool
			msg    string
		}
		record := func(id, sha string) outcome {
			_, errOut, err := runTodo(t, "landed", id, "--ref", "HEAD", "--sha", sha)
			if err == nil {
				return outcome{failed: false}
			}
			return outcome{failed: true, msg: errOut + err.Error()}
		}
		// Accepting branch (a).
		a1, a2 := record("1", f.mentioning[0]), record("2", f.mentioning[0])
		if a1.failed || a2.failed {
			t.Errorf("the accepting branch diverged by card id: t1 failed=%v (%s), t2 failed=%v (%s)",
				a1.failed, a1.msg, a2.failed, a2.msg)
		}
		// Refusing branch (c) — unreachable, so the classification is
		// reachability for both.
		r1, r2 := record("1", f.unreachable), record("2", f.unreachable)
		if !r1.failed || !r2.failed {
			t.Errorf("the refusing branch diverged by card id: t1 failed=%v, t2 failed=%v", r1.failed, r2.failed)
		}
		strip := func(s, id string) string { return strings.ReplaceAll(s, id, "<card>") }
		if strip(r1.msg, "t1") != strip(r2.msg, "t2") {
			t.Errorf("the two cards' refusal classifications differ; the card token reached a check\nt1: %s\nt2: %s", r1.msg, r2.msg)
		}
	})
}

// TestLandedGitCallsGuardEndOfOptions — every git invocation this verb makes
// carries a user-controlled operand (`--ref`, `--sha`), so each must stop
// option parsing before that operand reaches git.
//
// [HARD] This is a SHAPE assertion, not a behavioural RED, and the difference
// is recorded rather than papered over. The verb's own `^{commit}` gate
// (buildLandingEvidence) refuses every option-shaped ref before the raw-ref
// merge-base call can see one, so no end-to-end input distinguishes the
// guarded build from the unguarded one. TestLandedOptionShapedRefIsRefused
// below pins that reachability fact. The guard is latent hardening: it is the
// call sites, not today's exploitability, that this test fixes in place.
//
// The token is `--end-of-options`, NOT `--`: in rev-parse `--` separates
// revisions from PATHS, so a rev placed after it resolves to nothing and the
// verb breaks. Measured on git 2.50.1 — see .moai/reports/t359/f4-fix-evidence.md.
func TestLandedGitCallsGuardEndOfOptions(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "first card")

	type call struct{ argv []string }
	var calls []call
	prev := todoRunCommand
	t.Cleanup(func() { todoRunCommand = prev })
	todoRunCommand = func(name string, args ...string) (string, error) {
		calls = append(calls, call{argv: append([]string(nil), args...)})
		return prev(name, args...)
	}

	if _, _, err := runTodo(t, "landed", "1", "--ref", "HEAD", "--sha", f.mentioning[0]); err != nil {
		t.Fatalf("landed on a legitimate ref/sha failed: %v", err)
	}

	// Every reachable git subcommand that takes a user-controlled operand.
	// Absence of the subcommand entirely is a fixture failure, not a pass:
	// a guard asserted over zero calls asserts nothing.
	wantSubcommands := map[string]bool{"rev-parse": false, "merge-base": false}
	for _, c := range calls {
		sub := gitSubcommandOf(c.argv)
		if _, watched := wantSubcommands[sub]; !watched {
			continue
		}
		wantSubcommands[sub] = true
		idx := indexOfArg(c.argv, "--end-of-options")
		if idx < 0 {
			t.Errorf("git %v: no --end-of-options guard; a user-supplied operand reaches git as a parsable option", c.argv)
			continue
		}
		// Present is not enough — it must precede every operand it guards.
		// The assertion is stated on the arguments BEFORE the token: no
		// operand may appear there, because an operand placed ahead of the
		// token is still parsed as an option, which is the whole defect.
		//
		// It cannot be stated on the arguments AFTER the token, and that is
		// not an oversight: the token's purpose is to make everything after
		// it an operand whatever its shape, so a `--`-prefixed argument there
		// is a legitimate operand and is indistinguishable from a flag.
		if operand, found := operandBefore(c.argv, idx); found {
			t.Errorf("git %v: --end-of-options at %d is preceded by operand %q, which git still parses as an option", c.argv, idx, operand)
		}
	}
	for sub, seen := range wantSubcommands {
		if !seen {
			t.Fatalf("fixture: no %q call was observed, so the guard was asserted over nothing", sub)
		}
	}
}

// TestLandedOptionShapedRefIsRefused pins the reachability fact that makes the
// guard above latent rather than a repair: an option-shaped ref never reaches
// the raw-ref merge-base call, because the `^{commit}` gate refuses it first.
// If this test ever fails, the option-shaped value became reachable and the
// guard above stopped being latent — read it as that signal, not as a flake.
func TestLandedOptionShapedRefIsRefused(t *testing.T) {
	f := newLandedFixture(t)
	seedQueue(t, f.store, "first card")

	_, errOut, err := runTodo(t, "landed", "1", "--ref", "--independent", "--sha", f.mentioning[0])
	if err == nil {
		t.Fatal("an option-shaped --ref was accepted; it must refuse at the ^{commit} gate")
	}
	if !strings.Contains(errOut+err.Error(), "did not resolve to a commit") {
		t.Errorf("refusal did not come from the ^{commit} gate, so the reachability claim is unverified\nstderr+err: %s", errOut+err.Error())
	}
	if got := landingIsNULL(t, f.store, "t1"); got != 1 {
		t.Errorf("a record was written for an option-shaped ref (SELECT landing IS NULL = %d, want 1)", got)
	}
}

// gitSubcommandOf returns the git subcommand in argv, skipping the leading
// `-C <dir>` the runner prepends.
func gitSubcommandOf(argv []string) string {
	for i := 0; i < len(argv); i++ {
		if argv[i] == "-C" {
			i++
			continue
		}
		if !strings.HasPrefix(argv[i], "-") {
			return argv[i]
		}
	}
	return ""
}

// indexOfArg returns the position of want in argv, or -1.
func indexOfArg(argv []string, want string) int {
	for i, a := range argv {
		if a == want {
			return i
		}
	}
	return -1
}

// operandBefore returns the first operand appearing before idx in argv, and
// whether one was found. An operand is any argument that is not a `-`-prefixed
// flag, not the `-C <dir>` pair todoGitOutput prepends, and not the git
// subcommand itself.
func operandBefore(argv []string, idx int) (string, bool) {
	if idx > len(argv) {
		idx = len(argv)
	}
	seenSubcommand := false
	for i := 0; i < idx; i++ {
		if argv[i] == "-C" {
			i++ // the directory is -C's own operand, not the command's
			continue
		}
		if strings.HasPrefix(argv[i], "-") {
			continue
		}
		if !seenSubcommand {
			seenSubcommand = true
			continue
		}
		return argv[i], true
	}
	return "", false
}
