// Package template — card-id leak class guard.
//
// TestTemplateNoInternalContentLeak flagged an internal SPEC id plus an ISO
// date but let a bare card id (tNNNN) through, so a card citation could ship
// to every user project unseen. These tests pin the card-id class: it fires on
// the card-id shape in every scanned template surface, and stays silent on
// token shapes that only look like one.
package template

import "testing"

const cardIDClassName = "C9-card-id"

func TestCardIDLeakClassDetectsShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		text string
	}{
		{"card citation in an agent body", ".claude/agents/moai/example.md", "observed RED on card t1392 in a lane"},
		{"bare id in a rule", ".claude/rules/moai/workflow/example.md", "the same shape t1388 had already worked around"},
		{"bare id in a skill body", ".claude/skills/moai/workflows/example.md", "(t696):"},
		{"bare id in a hook comment", ".claude/hooks/moai/example.sh", "# exits silently (card t604)."},
		{"three-digit id", ".claude/agents/moai/example.md", "reproduced in t852 first-hand"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := collectLeakViolations(tc.path, tc.path, tc.text, leakClasses)
			if !anyViolationHasClass(got, cardIDClassName) {
				t.Errorf("class %q did not fire for %q at %q; violations: %v", cardIDClassName, tc.text, tc.path, got)
			}
		})
	}
}

func TestCardIDLeakClassIgnoresLookalikes(t *testing.T) {
	t.Parallel()

	path := ".claude/agents/moai/example.md"
	for _, text := range []string{
		"the wait is 5t0 steps",
		"sort the t0 and t1 buffers",
		"a hash t12345 is longer than a card id",
		"prefix_t1392 is an identifier, not a card",
		"at1392 is a word fragment",
		"the format is tNNNN",
	} {
		got := collectLeakViolations(path, path, text, leakClasses)
		if anyViolationHasClass(got, cardIDClassName) {
			t.Errorf("class %q fired on a lookalike %q: %v", cardIDClassName, text, got)
		}
	}
}
