package cli

// todo_verb_leak_test.go — t555 (#1654): the mistyped-verb guard's two
// discriminants were each narrower than the surface they protect, so a
// mistyped verb addressing a real card still leaked into `add`.
//
// The matrix below is the card's exhaustive enumeration, kept as the
// regression suite. It is deliberately three-sided, because a one-sided
// "these are refused" table cannot tell a working guard from one that refuses
// everything:
//
//	REFUSED  — unregistered first token + an address the verbs accept
//	CONTROL  — a REGISTERED verb, which must never reach the fallthrough
//	CARD     — the t69 fallthrough, which must still add
//
// Every case runs against the isolated fixture (todoFixture + the
// liveTodoQueueRootReason guard inside runTodo); none of them can reach the
// operator's live queue.

import (
	"strconv"
	"strings"
	"testing"
)

// TestTodoVerbGuardRefusesWidenedMistypes — the shapes t555 measured as
// leaking. Each pairs a first token that is not a registered verb with a
// second token the verbs themselves resolve to a card (normalizeTodoRef).
func TestTodoVerbGuardRefusesWidenedMistypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		why  string
	}{
		// Pre-existing coverage (t203), restated so a regression in the
		// original shape fails here too.
		{"lowercase_explicit_id", []string{"show", "t401"}, "a card id"},

		// t555 widening 1 — the first-token shape. Each of these addressed a
		// real card id and silently became a card.
		{"capitalized", []string{"Show", "t401"}, "a card id"},
		{"uppercase", []string{"SHOW", "t401"}, "a card id"},
		{"digit_typo", []string{"show2", "t401"}, "a card id"},
		{"underscore", []string{"show_it", "t401"}, "a card id"},
		{"single_char", []string{"s", "t401"}, "a card id"},
		{"long_word", []string{"supercalifragilis", "t401"}, "a card id"},

		// t555 widening 2 — the bare-number address form. `done 401` and
		// `done t401` address the same card, so `show 401` is a mistyped verb
		// addressing a card, not card text.
		{"bare_number", []string{"show", "401"}, "a card reference"},
		{"bare_number_capitalized", []string{"Show", "401"}, "a card reference"},

		// The guard still fires past two tokens for an explicit id.
		{"explicit_id_longer_phrase", []string{"show", "t401", "please"}, "a card id"},

		// t555 widening 2b — three verbs take TWO positional arguments
		// (`relate <a> <b>`, `drop <n> <reason>`, `edit <n> <text>`), so a
		// mistyped call of any of them is three tokens or more and slipped
		// past an arity-only condition. The `t<n>` spelling of each was
		// already refused; the bare spelling was not.
		{"near_miss_relate", []string{"relat", "401", "402"}, "a card reference"},
		{"near_miss_drop", []string{"drp", "401", "stale"}, "a card reference"},
		{"near_miss_edit", []string{"edt", "401", "new", "title"}, "a card reference"},
		{"near_miss_prefix", []string{"don", "401", "now"}, "a card reference"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, store := todoFixture(t)

			_, errOut, err := runTodo(t, tc.args...)
			if err == nil {
				t.Fatalf("%v was accepted, want a refusal", tc.args)
			}
			combined := err.Error() + errOut
			// The tokens are matched in their QUOTED form. A bare
			// strings.Contains on a one-character token like "s" is satisfied
			// by any English sentence, so it would assert nothing.
			for _, want := range []string{
				strconv.Quote(tc.args[0]), strconv.Quote(tc.args[1]), tc.why, "moai todo add",
			} {
				if !strings.Contains(combined, want) {
					t.Errorf("refusal should name %q; got %q", want, combined)
				}
			}

			rec, loadErr := store.Load()
			if loadErr != nil {
				t.Fatalf("load: %v", loadErr)
			}
			if len(rec.Items) != 0 {
				t.Errorf("refused invocation still mutated the queue: %+v", rec.Items)
			}
		})
	}
}

// TestTodoVerbGuardControlRegisteredVerbs — the control group. A REGISTERED
// verb is routed to its subcommand by cobra and never reaches the parent
// fallthrough, so it creates no card whatever second token follows it. Without
// this side, a guard that refused every two-token invocation would pass the
// table above.
func TestTodoVerbGuardControlRegisteredVerbs(t *testing.T) {
	for _, args := range [][]string{
		{"list", "t401"},
		{"list", "401"},
		{"done", "t401"},
		{"done", "401"},
		{"pr", "t401"},
		{"pr", "401"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			_, store := todoFixture(t)

			// The verb may well fail (no such card in an empty queue); what
			// matters is that it is the VERB failing, not the fallthrough
			// creating a card.
			_, _, _ = runTodo(t, args...)

			rec, loadErr := store.Load()
			if loadErr != nil {
				t.Fatalf("load: %v", loadErr)
			}
			if len(rec.Items) != 0 {
				t.Errorf("registered verb reached the add fallthrough: %+v", rec.Items)
			}
		})
	}
}

// TestTodoVerbGuardStillAddsNonAddresses — the other control. The widening is
// bounded by the verbs' own address grammar: a second token the verbs do NOT
// resolve to a card is card text, and still adds. Without this side the
// widening could have swallowed the t69 fallthrough wholesale.
func TestTodoVerbGuardStillAddsNonAddresses(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		// A plain second word — the ordinary two-word card.
		{[]string{"show", "card"}, "show card"},
		// `T401` and `t401x` are not ids the store issues or the verbs
		// resolve (`done T401` looks up "tT401"), so they are text.
		{[]string{"show", "T401"}, "show T401"},
		{[]string{"show", "t401x"}, "show t401x"},
		// A non-ASCII first token is prose, never a mistyped English verb.
		{[]string{"재현", "t401"}, "재현 t401"},
		// An ordinary word that is NOT a near-miss of any verb keeps a bare
		// number as text past two tokens — the near-miss arm is what
		// separates `drp 401 stale` from these, not the number itself.
		{[]string{"fix", "3", "flaky", "tests"}, "fix 3 flaky tests"},
		{[]string{"epic", "7", "planning"}, "epic 7 planning"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			_, store := todoFixture(t)

			if _, _, err := runTodo(t, tc.args...); err != nil {
				t.Fatalf("fallthrough refused %v: %v", tc.args, err)
			}
			rec, err := store.Load()
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if len(rec.Items) != 1 || rec.Items[0].Text != tc.want {
				t.Errorf("card = %+v; want one card %q", rec.Items, tc.want)
			}
		})
	}
}
