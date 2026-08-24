// todo_pr_test.go — SPEC-KANBAN-QUEUE-PR-SYNC-001 AC-004, AC-005, AC-009,
// AC-010, AC-014 (M3).
//
// Two of these criteria are structural rather than behavioural, and they are
// the ones worth reading first:
//
//	AC-004 digests the WHOLE queue directory, not just backlog.json. A single
//	       -file byte check passes a sidecar, a lock taken and released, and
//	       an mtime touch on a neighbour — all three of which are writes.
//	AC-014 counts gh processes at TWO queue lengths. A one-card fixture cannot
//	       distinguish one batched query from a per-card loop, and a per-card
//	       loop satisfies every other criterion while costing 0.878s per card.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoPRCall is one observed subprocess.
type todoPRCall struct {
	proc string
	argv []string
}

// spyRunner replaces the process seam for one test and records every
// invocation. The replacement is undone by t.Cleanup, so a failing test never
// leaks a stub into its neighbours.
type spyRunner struct {
	calls []todoPRCall
	// prJSON is what a `gh pr list` call returns. Empty means the call fails,
	// which is the fail-open fixture.
	prJSON string
	ghFail error
	// landedFor is the set of cards `git log` reports as landed.
	landedFor map[string]bool
}

func installSpy(t *testing.T, s *spyRunner) *spyRunner {
	t.Helper()
	prev := todoRunCommand
	t.Cleanup(func() { todoRunCommand = prev })
	todoRunCommand = func(name string, args ...string) (string, error) {
		s.calls = append(s.calls, todoPRCall{proc: name, argv: args})
		switch name {
		case "gh":
			if s.ghFail != nil {
				return "", s.ghFail
			}
			return s.prJSON, nil
		case "git":
			for card, landed := range s.landedFor {
				if landed && strings.Contains(strings.Join(args, " "), card) {
					return "d9899f437 fix: something (" + card + ")\n", nil
				}
			}
			return "", nil
		}
		return "", nil
	}
	return s
}

func (s *spyRunner) countOf(name string) int {
	n := 0
	for _, c := range s.calls {
		if c.proc == name {
			n++
		}
	}
	return n
}

// pinnedPRJSON is the `gh pr list --json number,title,body,state` payload,
// transcribed from acceptance.md's pinned block. It is NOT re-fetched.
const pinnedPRJSON = `[
 {"number":1614,"title":"fix(kanban): dispatch ordering (t203)","body":"closes t1 t151 t203 t69 t9","state":"OPEN"},
 {"number":1612,"title":"feat(cli): queue render (t200)","body":"t200 supersedes t201","state":"OPEN"},
 {"number":1611,"title":"chore: tidy the lock sweep","body":"part of t201","state":"OPEN"},
 {"number":1601,"title":"docs: worktree lifecycle","body":"card t188","state":"OPEN"}
]`

// seedQueue appends n cards and returns their ids.
func seedQueue(t *testing.T, store *kanban.BacklogStore, texts ...string) []string {
	t.Helper()
	ids := make([]string, 0, len(texts))
	for _, txt := range texts {
		item, _, err := store.Add(txt)
		if err != nil {
			t.Fatalf("seed %q: %v", txt, err)
		}
		ids = append(ids, item.ID)
	}
	return ids
}

// queueDirDigest is the recursive digest AC-004 asserts on: every path under
// the queue directory plus that file's SHA-256. A new path, a removed path,
// or a changed byte all move it.
func queueDirDigest(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "state", "kanban")
	var lines []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path) // #nosec G304 -- test fixture path
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		rel, _ := filepath.Rel(dir, path)
		lines = append(lines, rel+" "+hex.EncodeToString(sum[:]))
		return nil
	})
	if err != nil {
		t.Fatalf("digest %s: %v", dir, err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// AC-004 — nothing under the queue directory changes across an invocation
// (REQ-2.1, REQ-2.2). Asserted on the linked, ambiguous, landed, and
// fail-open paths, because the ruling has to hold on all of them.
func TestTodoPR_QueueDirUnchanged(t *testing.T) {
	cases := []struct {
		name string
		spy  *spyRunner
		args []string
	}{
		{"linked and ambiguous", &spyRunner{prJSON: pinnedPRJSON}, []string{"pr"}},
		{"json form", &spyRunner{prJSON: pinnedPRJSON}, []string{"pr", "--json"}},
		{"landed path", &spyRunner{prJSON: `[]`, landedFor: map[string]bool{"t1": true}}, []string{"pr"}},
		{"fail-open path", &spyRunner{ghFail: fmt.Errorf("gh: not found")}, []string{"pr"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, store := todoFixture(t)
			seedQueue(t, store, "first card", "second card")
			installSpy(t, tc.spy)

			before := queueDirDigest(t, root)
			beforeStat, err := os.Stat(store.Path())
			if err != nil {
				t.Fatalf("stat backlog: %v", err)
			}
			lockBefore, lockErr := os.Stat(store.LockPath())
			if lockErr != nil {
				t.Fatalf("stat backlog.lock: %v", lockErr)
			}

			if _, _, err := runTodo(t, tc.args...); err != nil {
				t.Fatalf("todo %v: %v", tc.args, err)
			}

			after := queueDirDigest(t, root)
			if before != after {
				t.Errorf("queue directory changed across the invocation\nbefore:\n%s\nafter:\n%s", before, after)
			}
			afterStat, err := os.Stat(store.Path())
			if err != nil {
				t.Fatalf("stat backlog after: %v", err)
			}
			if !beforeStat.ModTime().Equal(afterStat.ModTime()) {
				t.Errorf("backlog.json mtime moved: %v -> %v", beforeStat.ModTime(), afterStat.ModTime())
			}
			// The lock is NOT taken. The seeding writes left a lock artifact
			// behind, so its presence proves nothing — what proves the read
			// verb did not lock is that the artifact's mtime never moved.
			// (The digest covers its bytes; a lock taken and released is an
			// mtime event, not a content event.)
			if lockAfter, err := os.Stat(store.LockPath()); err == nil {
				if !lockBefore.ModTime().Equal(lockAfter.ModTime()) {
					t.Errorf("backlog.lock mtime moved: %v -> %v — the read verb took the lock",
						lockBefore.ModTime(), lockAfter.ModTime())
				}
			}
		})
	}
}

// AC-005 — fail-open when gh is unavailable (REQ-2.3).
func TestTodoPR_FailOpenNoGh(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "landed card", "untouched card")
	installSpy(t, &spyRunner{
		ghFail:    fmt.Errorf("exec: \"gh\": executable file not found in $PATH"),
		landedFor: map[string]bool{ids[0]: true},
	})

	out, errOut, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr must exit 0 when gh is absent, got: %v", err)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 2 {
		t.Fatalf("rendered %d rows, want 2:\n%s", len(lines), out)
	}
	for _, ln := range lines {
		cols := strings.Split(ln, "\t")
		if len(cols) < 5 {
			t.Fatalf("row %q has %d columns, want the full 5 — the link column is blank, not omitted", ln, len(cols))
		}
		if cols[2] != "" {
			t.Errorf("link column = %q with gh unavailable, want empty", cols[2])
		}
	}
	if !strings.Contains(errOut, "note:") {
		t.Errorf("stderr = %q, want a degradation note", errOut)
	}
	if strings.Contains(errOut, "Error:") {
		t.Errorf("stderr = %q; a degradation is a note, not an error", errOut)
	}
	// The landed check still runs (REQ-2.3): local git does not need gh.
	if !strings.Contains(lines[0], string(kanban.PRLinkLanded)) {
		t.Errorf("landed card row = %q, want the landed outcome", lines[0])
	}
}

// AC-005 (second half) — gh present but exiting non-zero.
func TestTodoPR_FailOpenGhNonZero(t *testing.T) {
	_, store := todoFixture(t)
	seedQueue(t, store, "a card")
	installSpy(t, &spyRunner{ghFail: fmt.Errorf("exit status 4: gh auth login required")})

	out, errOut, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr must exit 0 when gh fails, got: %v", err)
	}
	if !strings.Contains(errOut, "note:") {
		t.Errorf("stderr = %q, want a degradation note", errOut)
	}
	if !strings.Contains(out, string(kanban.PRLinkNoLink)) {
		t.Errorf("stdout = %q, want the no-link outcome rendered", out)
	}
}

// AC-009 — `moai todo list` stays network-free and git-free (REQ-2.4, REQ-2.5).
//
// A whole claim, not a no-flag-only one: REQ-2.5 puts no `--pr` flag in scope,
// so `todo list` spawns nothing under any invocation it accepts.
func TestTodoList_NoSubprocess(t *testing.T) {
	_, store := todoFixture(t)
	seedQueue(t, store, "first card", "second card")
	spy := installSpy(t, &spyRunner{prJSON: pinnedPRJSON})

	for _, args := range [][]string{{"list"}, {"list", "--json"}, {}} {
		spy.calls = nil
		out, _, err := runTodo(t, args...)
		if err != nil {
			t.Fatalf("todo %v: %v", args, err)
		}
		if strings.TrimSpace(out) == "" {
			t.Errorf("todo %v rendered nothing", args)
		}
		if len(spy.calls) != 0 {
			t.Errorf("todo %v spawned %d subprocess(es): %v", args, len(spy.calls), spy.calls)
		}
	}

	// And the flag genuinely does not exist — if a later change adds one,
	// this fails rather than silently widening AC-009's claim.
	if _, _, err := runTodo(t, "list", "--pr"); err == nil {
		t.Error("`todo list --pr` was accepted; REQ-2.5 puts no such flag in scope")
	}
}

// AC-010 — outcome and confidence are rendered, and the JSON form carries
// them (REQ-2.6).
func TestTodoPR_RendersOutcomeAndConfidence(t *testing.T) {
	_, store := todoFixture(t)
	// Four cards producing one of each: exact, inferred, ambiguous, landed.
	// Card ids are issued t1..t4, so the fixture PR payload is rewritten to
	// carry those tokens rather than the measured ones — the CARRIER
	// structure is what is under test here, and it is preserved exactly.
	ids := seedQueue(t, store, "exact card", "inferred card", "ambiguous card", "landed card")
	prJSON := fmt.Sprintf(`[
	 {"number":1612,"title":"feat: render (%s)","body":"%s supersedes %s","state":"OPEN"},
	 {"number":1601,"title":"docs: lifecycle","body":"card %s","state":"OPEN"},
	 {"number":1611,"title":"chore: sweep","body":"part of %s","state":"OPEN"}
	]`, ids[0], ids[0], ids[2], ids[1], ids[2])
	installSpy(t, &spyRunner{prJSON: prJSON, landedFor: map[string]bool{ids[3]: true}})

	out, _, err := runTodo(t, "pr", "--json")
	if err != nil {
		t.Fatalf("todo pr --json: %v", err)
	}
	var got []kanban.PRLinkOutcome
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", out, err)
	}
	if len(got) != 4 {
		t.Fatalf("json carried %d records, want 4", len(got))
	}
	byID := map[string]kanban.PRLinkOutcome{}
	for _, o := range got {
		byID[o.CardID] = o
		if o.CardID == "" || o.Kind == "" {
			t.Errorf("record %+v is missing card_id or outcome", o)
		}
	}
	if o := byID[ids[0]]; o.Kind != kanban.PRLinkLinked || o.Confidence != kanban.PRLinkExact || o.PRState != "OPEN" {
		t.Errorf("%s = %+v, want linked/exact/OPEN", ids[0], o)
	}
	if o := byID[ids[1]]; o.Kind != kanban.PRLinkLinked || o.Confidence != kanban.PRLinkInferred {
		t.Errorf("%s = %+v, want linked/inferred", ids[1], o)
	}
	if o := byID[ids[2]]; o.Kind != kanban.PRLinkAmbiguous || len(o.PRs) != 2 {
		t.Errorf("%s = %+v, want ambiguous with 2 candidates", ids[2], o)
	}
	if o := byID[ids[3]]; o.Kind != kanban.PRLinkLanded {
		t.Errorf("%s = %+v, want landed", ids[3], o)
	}

	// The human render displays the outcome kind for every card, and the
	// confidence label for every linked card.
	human, _, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}
	for _, ln := range nonEmptyLines(human) {
		cols := strings.Split(ln, "\t")
		if len(cols) < 5 || cols[1] == "" {
			t.Fatalf("row %q does not display an outcome kind", ln)
		}
		if cols[1] == string(kanban.PRLinkLinked) && cols[3] == "" {
			t.Errorf("linked row %q displays no confidence label", ln)
		}
	}
}

// AC-014 — exactly one gh process, regardless of queue length (NFR-1, NFR-2).
//
// Asserted at TWO lengths so the criterion observes invariance rather than a
// coincidence. A one-card fixture would pass a per-card loop.
func TestTodoPR_ExactlyOneGhInvocation(t *testing.T) {
	for _, n := range []int{3, 10} {
		t.Run(fmt.Sprintf("%d-cards", n), func(t *testing.T) {
			_, store := todoFixture(t)
			texts := make([]string, 0, n)
			for i := range n {
				texts = append(texts, fmt.Sprintf("card number %d", i))
			}
			seedQueue(t, store, texts...)
			spy := installSpy(t, &spyRunner{prJSON: pinnedPRJSON})

			if _, _, err := runTodo(t, "pr"); err != nil {
				t.Fatalf("todo pr: %v", err)
			}

			if got := spy.countOf("gh"); got != 1 {
				t.Errorf("%d gh processes for a %d-card queue, want exactly 1: %v", got, n, spy.calls)
			}
			for _, c := range spy.calls {
				if c.proc != "gh" {
					continue
				}
				joined := strings.Join(c.argv, " ")
				if !strings.HasPrefix(joined, "pr list --state open") {
					t.Errorf("gh argv = %q, want a single `pr list --state open` query", joined)
				}
				if strings.Contains(joined, "pr view") {
					t.Errorf("gh argv = %q is a per-card view, not a batched list", joined)
				}
			}
			// NFR-2: the landed path spawns git and never gh, so it consumes
			// no part of the one-query network budget.
			sawGit := false
			for _, c := range spy.calls {
				if c.proc == "git" {
					sawGit = true
					if strings.Contains(strings.Join(c.argv, " "), "gh") {
						t.Errorf("landed check argv %v routes through gh", c.argv)
					}
				}
			}
			if !sawGit {
				t.Error("no git process was spawned; the landed check did not run")
			}
		})
	}
}

// A single-card argument narrows the render without changing the query count.
func TestTodoPR_SingleCardArgument(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "first card", "second card", "third card")
	spy := installSpy(t, &spyRunner{prJSON: pinnedPRJSON})

	out, _, err := runTodo(t, "pr", ids[1])
	if err != nil {
		t.Fatalf("todo pr %s: %v", ids[1], err)
	}
	lines := nonEmptyLines(out)
	if len(lines) != 1 {
		t.Fatalf("rendered %d rows for one card:\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], ids[1]+"\t") {
		t.Errorf("row = %q, want the %s row", lines[0], ids[1])
	}
	if got := spy.countOf("gh"); got != 1 {
		t.Errorf("%d gh processes, want 1", got)
	}
}

// A saturated open-PR page is reported, not hidden. A pull request beyond the
// ceiling is invisible to the resolver, so its card reports `no-link` or
// `landed` — wrong, and silent. Paging is not the fix: a second page is a
// second `gh` process, which AC-014 forbids.
func TestTodoPR_SaturatedPageIsReported(t *testing.T) {
	_, store := todoFixture(t)
	seedQueue(t, store, "a card")

	// Exactly `todoPROpenPRLimit` records back from one query.
	recs := make([]string, 0, todoPROpenPRLimit)
	for i := range todoPROpenPRLimit {
		recs = append(recs, fmt.Sprintf(
			`{"number":%d,"title":"chore: filler","body":"","state":"OPEN"}`, 9000+i))
	}
	spy := installSpy(t, &spyRunner{prJSON: "[" + strings.Join(recs, ",") + "]"})

	_, errOut, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}
	if !strings.Contains(errOut, "ceiling") {
		t.Errorf("stderr = %q, want a saturation note", errOut)
	}
	if got := spy.countOf("gh"); got != 1 {
		t.Errorf("%d gh processes on the saturated path, want 1 — saturation is reported, never paged around", got)
	}
}

// An unsaturated page says nothing — the note must not fire on every run.
func TestTodoPR_UnsaturatedPageIsSilent(t *testing.T) {
	_, store := todoFixture(t)
	seedQueue(t, store, "a card")
	installSpy(t, &spyRunner{prJSON: pinnedPRJSON})

	_, errOut, err := runTodo(t, "pr")
	if err != nil {
		t.Fatalf("todo pr: %v", err)
	}
	if strings.Contains(errOut, "ceiling") {
		t.Errorf("stderr = %q; the saturation note fired on a 4-record page", errOut)
	}
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			out = append(out, ln)
		}
	}
	return out
}
