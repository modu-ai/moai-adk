// todo_triage_test.go — `moai todo triage` (card t943).
//
// The criteria worth reading first are the ones that carry the verb's whole
// value, because each of them guards a sentence rather than a number:
//
//	TestTodoTriage_EmptyExtractionSaysItIsALimit   — 54% of real cards extract
//	      nothing, and that silence was being read as a finding. The framing IS
//	      the fix, so its absence is a regression even when every count is right.
//	TestTodoTriage_ZeroPresenceRendersBothReadings — nothing in the shipped tool
//	      classifies which reading applies, so both must ship unconditionally.
//	TestTodoTriage_ControlPrecedesPresence         — a presence result printed
//	      before its control cannot be told from a broken probe.
//	TestTodoTriage_QueueUnchanged                  — the [HARD] no-mutation
//	      guarantee, asserted rather than assumed.
//	TestTodoTriage_SubprocessBound                 — the documented ceiling, at
//	      two queue lengths: one card cannot distinguish a cached control from a
//	      per-card one.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// triageSpy installs a runner that answers every git probe from a plan keyed
// by the subcommand, and records every invocation. The replacement is undone
// by t.Cleanup, so a failing test never leaks a stub into its neighbours.
type triageSpy struct {
	calls []todoPRCall
	// answers maps a probe key to its reply, in the order the probes are
	// issued for that key. A key with no plan answers empty-and-successful.
	grep   []spyLogAnswer
	lsTree spyLogAnswer
	logs   []spyLogAnswer

	grepN, logN int
}

func installTriageSpy(t *testing.T, s *triageSpy) *triageSpy {
	t.Helper()
	prev := todoRunCommand
	t.Cleanup(func() { todoRunCommand = prev })
	todoRunCommand = func(name string, args ...string) (string, error) {
		s.calls = append(s.calls, todoPRCall{proc: name, argv: args})
		if name != "git" || len(args) == 0 {
			return "", nil
		}
		switch args[0] {
		case "grep":
			i := s.grepN
			s.grepN++
			if i < len(s.grep) {
				return s.grep[i].out, s.grep[i].err
			}
		case "ls-tree":
			return s.lsTree.out, s.lsTree.err
		case "log":
			i := s.logN
			s.logN++
			if i < len(s.logs) {
				return s.logs[i].out, s.logs[i].err
			}
		}
		return "", nil
	}
	return s
}

// gitProbes counts the git subprocesses that went through the seam.
func (s *triageSpy) gitProbes() int {
	n := 0
	for _, c := range s.calls {
		if c.proc == "git" {
			n++
		}
	}
	return n
}

// healthyControl is a control reply above the floor: enough lines that the
// probe reads as working.
func healthyControl() string {
	var b strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "develop:internal/pkg/file%d.go:3\n", i)
	}
	return b.String()
}

// --- symbol extraction -------------------------------------------------------

func TestTodoTriageSymbols(t *testing.T) {
	cases := []struct {
		name string
		text string
		want []string
	}{
		{"empty text", "", nil},
		{"no extractable shape", "the queue feels slow when many cards are open", nil},
		{
			"backtick identifier",
			"the `resolveProjectDir` helper reads the wrong tree",
			[]string{"resolveProjectDir"},
		},
		{
			"path-like literal, unquoted",
			"internal/cli/todo_pr.go loses the ref",
			[]string{"internal/cli/todo_pr.go"},
		},
		{
			"call-shaped token",
			"the render calls formatLanding( with a nil evidence",
			[]string{"formatLanding"},
		},
		{
			"korean card text with a backticked path",
			"`internal/kanban/prlink.go` 의 판별식이 브랜치마다 다르게 읽힌다",
			[]string{"internal/kanban/prlink.go"},
		},
		{
			"korean prose with no extractable shape",
			"이 카드는 배차 전에 전제가 살아 있는지 확인해야 한다",
			nil,
		},
		{
			"pattern order decides which four survive",
			"`aaaa` `bbbb` `cccc` `dddd` `eeee` and internal/x/y.go",
			[]string{"aaaa", "bbbb", "cccc", "dddd"},
		},
		{
			"duplicates collapse",
			"`sameName` and `sameName` again, plus `other`",
			[]string{"sameName", "other"},
		},
		{
			"mixed sources keep the documented order",
			"`quoted` then internal/a/b.go then longEnoughCall(",
			[]string{"quoted", "internal/a/b.go", "longEnoughCall"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := todoTriageSymbols(tc.text)
			if len(got) != len(tc.want) {
				t.Fatalf("symbols = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("symbols = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestTodoTriage_PathLikeClassification(t *testing.T) {
	cases := map[string]bool{
		"internal/cli/todo.go": true,
		"todo_pr.go":           true,
		"docs/a.md":            true,
		"resolveProjectDir":    false,
		"formatLanding":        false,
	}
	for sym, want := range cases {
		if got := todoTriageIsPathLike(sym); got != want {
			t.Errorf("todoTriageIsPathLike(%q) = %v, want %v", sym, got, want)
		}
	}
}

// --- the framing sentences ---------------------------------------------------

// The empty-extraction block must say the silence is a LIMIT OF THE EXTRACTOR,
// and must refuse both readings a bare "none found" invites.
func TestTodoTriage_EmptyExtractionSaysItIsALimit(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "이 카드는 배차 전에 전제를 다시 재야 한다")
	installTriageSpy(t, &triageSpy{})

	out, _, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("triage: %v", err)
	}
	for _, want := range []string{
		"none extracted",
		"LIMIT OF THE EXTRACTOR",
		"backtick-quoted",
		"path-like literals",
		"call-shaped tokens",
		"NOT evidence",
		"the defect is",
		"Weigh the card's own prose",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("empty-extraction block missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// A zero presence hit renders BOTH readings, unconditionally — nothing in the
// shipped tool classifies which one applies.
func TestTodoTriage_ZeroPresenceRendersBothReadings(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `neverBuiltThing` is not wired up")
	installTriageSpy(t, &triageSpy{
		// call 1: control (healthy). call 2: presence (empty).
		// call 3: identifier neighborhood (empty).
		grep: []spyLogAnswer{{out: healthyControl()}, {}, {}},
	})

	out, _, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("triage: %v", err)
	}
	if !strings.Contains(out, "presence: 0 files") {
		t.Fatalf("expected a zero presence line\n--- got ---\n%s", out)
	}
	for _, want := range []string{
		"Two readings",
		"MISSING",
		"being ALIVE",
		"not evidence it is dead",
		"EXISTS BUT IS WRONG",
		"renamed or removed",
		"does not decide which case applies",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("zero-presence block missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// The positive control is printed BEFORE any presence result — a presence
// number read ahead of its control cannot be told from a broken probe.
func TestTodoTriage_ControlPrecedesPresence(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `resolveThing` helper drifted")
	installTriageSpy(t, &triageSpy{
		grep: []spyLogAnswer{{out: healthyControl()}, {out: "develop:internal/a.go:2\n"}, {}},
	})

	out, _, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("triage: %v", err)
	}
	ctrl := strings.Index(out, "control:")
	pres := strings.Index(out, "presence:")
	if ctrl < 0 || pres < 0 {
		t.Fatalf("expected both a control and a presence line\n--- got ---\n%s", out)
	}
	if ctrl > pres {
		t.Errorf("control (at %d) must precede presence (at %d)\n--- got ---\n%s", ctrl, pres, out)
	}
	if !strings.Contains(out, "the probe works") {
		t.Errorf("a healthy control must say so\n--- got ---\n%s", out)
	}
}

// A near-zero control says the probe looks broken, and withdraws trust from
// the presence results rather than letting them read as absences.
func TestTodoTriage_BrokenControlWithdrawsTrust(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `resolveThing` helper drifted")
	installTriageSpy(t, &triageSpy{
		grep: []spyLogAnswer{{out: "develop:only/one.go:1\n"}, {}, {}},
	})

	out, _, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("triage: %v", err)
	}
	if !strings.Contains(out, "the probe looks BROKEN") {
		t.Errorf("a near-zero control must say the probe looks broken\n--- got ---\n%s", out)
	}
	if !strings.Contains(out, "cannot be trusted") {
		t.Errorf("a broken control must withdraw trust from the presence results\n--- got ---\n%s", out)
	}
}

// --- unmeasured is not zero --------------------------------------------------

// A section that could not be measured renders as unmeasured, DISTINCTLY from
// a measured zero, in both text and JSON. Exit code stays 0; a note goes to
// stderr.
func TestTodoTriage_UnmeasuredIsNotZero(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `resolveThing` helper drifted")
	// Exit 128 is git's "bad revision" — unmeasured, not empty.
	badRef := fmt.Errorf("git: fatal: bad revision")
	installTriageSpy(t, &triageSpy{
		grep: []spyLogAnswer{{err: badRef}, {err: badRef}, {err: badRef}},
		logs: []spyLogAnswer{{err: badRef}, {err: badRef}},
	})

	out, errOut, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("an unmeasurable probe must not fail the command: %v", err)
	}
	for _, want := range []string{
		"control: could not measure",
		"presence: could not measure",
		"this is NOT a zero",
		"neighborhood: could not measure",
		"could not measure",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("unmeasured render missing %q\n--- got ---\n%s", want, out)
		}
	}
	if strings.Contains(out, "presence: 0 files") {
		t.Errorf("an unmeasured presence must never render as a zero\n--- got ---\n%s", out)
	}
	if !strings.Contains(errOut, "positive control could not be measured") {
		t.Errorf("a degraded control must be noted on stderr\n--- got ---\n%s", errOut)
	}

	// JSON carries the same distinction. A FRESH spy: the plan above is
	// consumed in call order, and a second run against a spent plan would
	// answer empty-and-successful — which is the very state this test
	// distinguishes from unmeasured.
	installTriageSpy(t, &triageSpy{
		grep: []spyLogAnswer{{err: badRef}, {err: badRef}, {err: badRef}},
		logs: []spyLogAnswer{{err: badRef}, {err: badRef}},
	})
	jsonOut, _, err := runTodo(t, "triage", ids[0], "--json")
	if err != nil {
		t.Fatalf("triage --json: %v", err)
	}
	var cards []todoTriageCard
	if err := json.Unmarshal([]byte(jsonOut), &cards); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, jsonOut)
	}
	if len(cards) != 1 {
		t.Fatalf("cards = %d, want 1", len(cards))
	}
	c := cards[0]
	if c.Control.Measured {
		t.Errorf("control.measured = true, want false")
	}
	if c.Control.Reason == "" {
		t.Errorf("an unmeasured control must carry a reason")
	}
	if len(c.Details) != 1 {
		t.Fatalf("details = %d, want 1", len(c.Details))
	}
	d := c.Details[0]
	if d.Presence.Measured || d.Neighborhood.Measured || d.Provenance.Measured {
		t.Errorf("every section must report unmeasured: %+v", d)
	}
	if d.Presence.Files != 0 {
		t.Errorf("an unmeasured presence carries no count, and the discriminator is the measured flag")
	}
}

// A measured zero is measured: the flag separates it from the case above.
func TestTodoTriage_MeasuredZeroIsMeasuredInJSON(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `resolveThing` helper drifted")
	installTriageSpy(t, &triageSpy{grep: []spyLogAnswer{{out: healthyControl()}, {}, {}}})

	out, _, err := runTodo(t, "triage", ids[0], "--json")
	if err != nil {
		t.Fatalf("triage --json: %v", err)
	}
	var cards []todoTriageCard
	if err := json.Unmarshal([]byte(out), &cards); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, out)
	}
	d := cards[0].Details[0]
	if !d.Presence.Measured {
		t.Errorf("a measured zero must carry measured=true")
	}
	if d.Presence.Files != 0 {
		t.Errorf("presence.files = %d, want 0", d.Presence.Files)
	}
	if d.Presence.Reason != "" {
		t.Errorf("a measured section carries no reason, got %q", d.Presence.Reason)
	}
}

// --- neighborhood ------------------------------------------------------------

// The neighborhood is the discriminator: an absent path whose directory holds
// a near-name sibling is a different conclusion from one whose directory is
// empty, and the render must carry that sentence.
func TestTodoTriage_PathNeighborhoodReadsTheCachedTree(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "`internal/cli/todo_gone.go` never shipped")
	installTriageSpy(t, &triageSpy{
		grep: []spyLogAnswer{{out: healthyControl()}, {}},
		lsTree: spyLogAnswer{out: strings.Join([]string{
			"internal/cli/todo_pr.go",
			"internal/cli/todo.go",
			"internal/kanban/todo_gone_helper.go",
		}, "\n")},
	})

	out, _, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("triage: %v", err)
	}
	if !strings.Contains(out, "2 entries in internal/cli") {
		t.Errorf("expected the directory listing\n--- got ---\n%s", out)
	}
	if !strings.Contains(out, "internal/kanban/todo_gone_helper.go") {
		t.Errorf("expected the near-name sibling elsewhere\n--- got ---\n%s", out)
	}
	if !strings.Contains(out, "OPPOSITE conclusions") {
		t.Errorf("the discriminating sentence must ship\n--- got ---\n%s", out)
	}
}

// Near-name matches are restricted to the SAME EXTENSION as the named symbol.
//
// This pins a defect FOUND LIVE, not imagined: running the verb against
// origin/develop for a card naming `internal/cli/update/merge/base.go`, the
// stem `base` matched 55 paths by substring, and because dot-directories sort
// first, the ten rows the cap actually rendered were
// `codebase-analysis.md`, `supabase.md` and a run of `*-baseline.*` — ten rows
// of chaff and not one real Go neighbour. The section meant to answer "is a
// sibling sitting right there under another name" showed the reader nothing.
//
// The fixture below reproduces exactly that shape: a short common stem, real
// same-extension neighbours, and other-extension files whose names contain the
// stem. The chaff must not appear, and it must not be counted either — a cap
// applied to a padded total is the same defect one step removed.
func TestTodoTriage_NearNamesAreSameExtensionOnly(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "`internal/cli/update/merge/base.go` 의 near-name 목록이 노이즈로 가득하다")

	sameExt := []string{
		"internal/cli/base_loader.go",
		"internal/settings_snapshot_base_test.go",
	}
	otherExt := []string{
		".claude/skills/moai/workflows/project/codebase-analysis.md",
		".moai/archive/skills/v3.0/moai-platform-database-cloud/reference/supabase.md",
		".moai/plans/DOCS-SITE/phase-2-build-baseline.md",
		".moai/reports/t338/ac-count-baseline.txt",
		".moai/spec-lint-baseline.json",
	}
	listing := append([]string{
		// A direct sibling in the symbol's own directory: reported as a
		// directory entry, never as a near name.
		"internal/cli/update/merge/merge.go",
	}, otherExt...)
	listing = append(listing, sameExt...)

	installTriageSpy(t, &triageSpy{
		grep:   []spyLogAnswer{{out: healthyControl()}, {}},
		lsTree: spyLogAnswer{out: strings.Join(listing, "\n")},
	})

	out, _, err := runTodo(t, "triage", ids[0], "--json")
	if err != nil {
		t.Fatalf("triage --json: %v", err)
	}
	var cards []todoTriageCard
	if err := json.Unmarshal([]byte(out), &cards); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, out)
	}
	n := cards[0].Details[0].Neighborhood
	if n.Kind != "path" {
		t.Fatalf("kind = %q, want path", n.Kind)
	}
	got := strings.Join(n.NearNames, "\n")
	for _, want := range sameExt {
		if !strings.Contains(got, want) {
			t.Errorf("same-extension neighbour %q is missing\n--- got ---\n%s", want, got)
		}
	}
	for _, chaff := range otherExt {
		if strings.Contains(got, chaff) {
			t.Errorf("other-extension chaff %q must not appear\n--- got ---\n%s", chaff, got)
		}
	}
	// The COUNT is filtered too, not just the rendered rows.
	if n.NearCount != len(sameExt) {
		t.Errorf("near_count = %d, want %d — the filter must apply before the cap, or a padded total re-hides the neighbours",
			n.NearCount, len(sameExt))
	}
	// The symbol's own directory stays a directory entry, never a near name.
	if strings.Contains(got, "internal/cli/update/merge/merge.go") {
		t.Errorf("a direct sibling belongs in the directory listing, not in near names\n--- got ---\n%s", got)
	}
}

func TestTodoTriage_MissingDirectorySaysSo(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "`internal/nope/thing.go` is gone")
	installTriageSpy(t, &triageSpy{
		grep:   []spyLogAnswer{{out: healthyControl()}, {}},
		lsTree: spyLogAnswer{out: "internal/cli/todo.go\n"},
	})

	out, _, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("triage: %v", err)
	}
	if !strings.Contains(out, "does not exist at this ref") {
		t.Errorf("a missing directory must say so\n--- got ---\n%s", out)
	}
}

// --- provenance --------------------------------------------------------------

func TestTodoTriage_ProvenanceCarriesTheSiblingNote(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `resolveThing` helper drifted")
	installTriageSpy(t, &triageSpy{
		grep: []spyLogAnswer{{out: healthyControl()}, {out: "develop:internal/a.go:1\n"}, {}},
		logs: []spyLogAnswer{
			{out: "abc1234 fix(cli): resolve the thing (t900)\n"},
			{},
		},
	})

	out, _, err := runTodo(t, "triage", ids[0])
	if err != nil {
		t.Fatalf("triage: %v", err)
	}
	if !strings.Contains(out, "abc1234 fix(cli): resolve the thing (t900)") {
		t.Errorf("expected the commit line\n--- got ---\n%s", out)
	}
	if !strings.Contains(out, "an ancestor of") {
		t.Errorf("the ancestor statement must ship\n--- got ---\n%s", out)
	}
	for _, want := range []string{
		"SIBLING card's id",
		"NOT evidence the work was never done",
		"CURRENT integration ref",
		"this card's own",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("provenance framing missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// --- unknown id --------------------------------------------------------------

func TestTodoTriage_UnknownIDIsOneCleanLine(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `resolveThing` helper drifted")
	installTriageSpy(t, &triageSpy{grep: []spyLogAnswer{{out: healthyControl()}, {}, {}}})

	out, _, err := runTodo(t, "triage", "t999", ids[0])
	if err != nil {
		t.Fatalf("an unknown id is not an error: %v", err)
	}
	if !strings.Contains(out, "t999\tnot in the queue") {
		t.Errorf("expected the unknown-id line\n--- got ---\n%s", out)
	}
	// The remaining ids are still processed.
	if !strings.Contains(out, ids[0]) {
		t.Errorf("the next id must still be observed\n--- got ---\n%s", out)
	}

	jsonOut, _, err := runTodo(t, "triage", "t999", "--json")
	if err != nil {
		t.Fatalf("triage --json: %v", err)
	}
	var cards []todoTriageCard
	if err := json.Unmarshal([]byte(jsonOut), &cards); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, jsonOut)
	}
	if len(cards) != 1 || cards[0].Found {
		t.Errorf("an unknown id renders found=false, got %+v", cards)
	}
}

// --- [HARD] no mutation ------------------------------------------------------

// The queue is byte-identical before and after a run. Asserted, not assumed.
// The digest is drawn at the whole project root (queueDirDigest), so a sidecar,
// a cache, a lock taken and released, or a touched neighbour all move it.
func TestTodoTriage_QueueUnchanged(t *testing.T) {
	cases := []struct {
		name string
		spy  *triageSpy
		args []string
	}{
		{"healthy path", &triageSpy{
			grep:   []spyLogAnswer{{out: healthyControl()}, {out: "develop:a.go:1\n"}, {}},
			lsTree: spyLogAnswer{out: "internal/cli/todo.go\n"},
			logs:   []spyLogAnswer{{out: "abc1234 fix: a thing\n"}, {}},
		}, []string{"triage", "t1", "t2"}},
		{"json form", &triageSpy{
			grep: []spyLogAnswer{{out: healthyControl()}, {}, {}},
		}, []string{"triage", "t1", "--json"}},
		{"degraded path", &triageSpy{
			grep: []spyLogAnswer{{err: fmt.Errorf("git: not found")}},
			logs: []spyLogAnswer{{err: fmt.Errorf("git: not found")}},
		}, []string{"triage", "t1", "t2"}},
		{"empty extraction path", &triageSpy{}, []string{"triage", "t3"}},
		{"unknown id path", &triageSpy{}, []string{"triage", "t999"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, store := todoFixture(t)
			seedQueue(t, store,
				"the `resolveThing` helper drifted",
				"`internal/cli/todo_gone.go` never shipped",
				"이 카드에는 뽑아낼 기호가 없다")
			installTriageSpy(t, tc.spy)

			before := queueDirDigest(t, root)
			beforeStat, err := os.Stat(store.EnginePath())
			if err != nil {
				t.Fatalf("stat backlog: %v", err)
			}

			if _, _, err := runTodo(t, tc.args...); err != nil {
				t.Fatalf("triage: %v", err)
			}

			if after := queueDirDigest(t, root); after != before {
				t.Errorf("triage mutated the project root\nbefore:\n%s\nafter:\n%s", before, after)
			}
			afterStat, err := os.Stat(store.EnginePath())
			if err != nil {
				t.Fatalf("stat backlog after: %v", err)
			}
			if !afterStat.ModTime().Equal(beforeStat.ModTime()) {
				t.Errorf("triage touched the queue mtime: %v -> %v",
					beforeStat.ModTime(), afterStat.ModTime())
			}
		})
	}
}

// A second guard on the same invariant, drawn at the bytes of the queue file
// itself rather than at the tree digest — the two fail for different reasons.
func TestTodoTriage_QueueFileBytesIdentical(t *testing.T) {
	_, store := todoFixture(t)
	seedQueue(t, store, "the `resolveThing` helper drifted")
	installTriageSpy(t, &triageSpy{grep: []spyLogAnswer{{out: healthyControl()}, {}, {}}})

	digest := func() string {
		data, err := os.ReadFile(store.EnginePath()) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatalf("read backlog: %v", err)
		}
		sum := sha256.Sum256(data)
		return hex.EncodeToString(sum[:])
	}
	before := digest()
	if _, _, err := runTodo(t, "triage", "t1"); err != nil {
		t.Fatalf("triage: %v", err)
	}
	if after := digest(); after != before {
		t.Errorf("queue bytes changed: %s -> %s", before, after)
	}
}

// --- subprocess census -------------------------------------------------------

// Every subprocess goes through todoRunCommand, and the per-run count stays
// within the documented bound. Measured at TWO queue lengths: one card cannot
// distinguish a cached control from a per-card one.
func TestTodoTriage_SubprocessBound(t *testing.T) {
	for _, n := range []int{1, 3} {
		t.Run(fmt.Sprintf("%d cards", n), func(t *testing.T) {
			_, store := todoFixture(t)
			texts := make([]string, 0, n)
			for i := 0; i < n; i++ {
				texts = append(texts,
					fmt.Sprintf("`internal/cli/gone%d.go` and `symOne%d` and `symTwo%d` and `symThree%d` drifted", i, i, i, i))
			}
			ids := seedQueue(t, store, texts...)
			spy := installTriageSpy(t, &triageSpy{
				grep:   []spyLogAnswer{{out: healthyControl()}},
				lsTree: spyLogAnswer{out: "internal/cli/todo.go\n"},
			})

			args := append([]string{"triage"}, ids...)
			if _, _, err := runTodo(t, args...); err != nil {
				t.Fatalf("triage: %v", err)
			}

			// Documented ceiling: 2 cached probes + 13 per card.
			bound := 2 + 13*n
			got := spy.gitProbes()
			if got > bound {
				t.Errorf("%d git probes for %d cards, documented bound is %d", got, n, bound)
			}
			// Non-vacuity: a bound satisfied by spending nothing measures the
			// stub, not the verb. Each card here extracts four symbols, so the
			// run must spend at least the control plus one probe per symbol.
			if floor := 1 + 4*n; got < floor {
				t.Errorf("%d git probes for %d cards is below %d — the census is not measuring the verb", got, n, floor)
			}
			t.Logf("%d cards: %d git probes (bound %d)", n, got, bound)
			// Every recorded call is a git call — nothing else escaped.
			for _, c := range spy.calls {
				if c.proc != "git" {
					t.Errorf("unexpected subprocess %q through the seam", c.proc)
				}
			}
			// The two cached probes are taken AT MOST ONCE for the whole run.
			control, lsTree := 0, 0
			for _, c := range spy.calls {
				if len(c.argv) == 0 {
					continue
				}
				if c.argv[0] == "ls-tree" {
					lsTree++
				}
				if c.argv[0] == "grep" && len(c.argv) >= 4 && c.argv[3] == todoTriageControlPattern {
					control++
				}
			}
			if control != 1 {
				t.Errorf("positive control ran %d times, want 1 (cached across the run)", control)
			}
			if lsTree > 1 {
				t.Errorf("tree listing ran %d times, want at most 1 (cached across the run)", lsTree)
			}
		})
	}
}

// A run whose cards extract no path-like symbol never spends the tree listing.
func TestTodoTriage_TreeListingIsLazy(t *testing.T) {
	_, store := todoFixture(t)
	ids := seedQueue(t, store, "the `resolveThing` helper drifted")
	spy := installTriageSpy(t, &triageSpy{grep: []spyLogAnswer{{out: healthyControl()}, {}, {}}})

	if _, _, err := runTodo(t, "triage", ids[0]); err != nil {
		t.Fatalf("triage: %v", err)
	}
	for _, c := range spy.calls {
		if len(c.argv) > 0 && c.argv[0] == "ls-tree" {
			t.Errorf("tree listing ran for an identifier-only card")
		}
	}
}

// --- registration ------------------------------------------------------------

func TestTodoTriage_IsRegistered(t *testing.T) {
	cmd := newTodoCmd()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "triage" {
			return
		}
	}
	t.Errorf("triage is not registered on the todo command tree")
}

// The verb never prompts — the CLI-wide subagent boundary (C-HRA-008).
func TestTodoTriage_NoAskUserQuestion(t *testing.T) {
	data, err := os.ReadFile("todo_triage.go")
	if err != nil {
		t.Fatalf("read source: %v", err)
	}
	for _, banned := range []string{"AskUserQuestion", "mcp__askuser__", "bufio.NewReader(os.Stdin)"} {
		if strings.Contains(string(data), banned) {
			t.Errorf("todo_triage.go must not carry %q — the CLI never prompts", banned)
		}
	}
}
