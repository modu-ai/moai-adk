// todo_surface_test.go — the frozen verb-surface zero-delta guard
// (SPEC-TODO-SQLITE-001 AC-TOSQ-010, REQ-TOSQ-007).
//
// REQ-TOSQ-007 freezes the `moai todo` command-line surface across the storage
// swap: no verb renamed, removed, or re-flagged. The rest of this SPEC's CLI
// coverage proves BEHAVIOR is preserved — the same cards come back, refusals
// still refuse. This file proves the SHAPE is preserved, which is a different
// claim and fails differently: a dropped flag or a renamed verb breaks every
// script and every agent dispatch that types the command, while every
// behavioral test keeps passing because it types the new spelling.
//
// The frozen table below is the surface as it stood at 7ed6edb3e, the branch
// point. That provenance is not a recollection — it was measured:
//
//	git diff 7ed6edb3e..HEAD -- 'internal/cli/todo*.go' \
//	  | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' \
//	  | grep -E 'Use:|Flags\(\)\.|Args:|cobra\.(NoArgs|ExactArgs|...)'
//
// which returned exactly two lines, both introducing export-json. Every other
// verb and flag declaration on the branch is byte-unchanged.
//
// The table is read from the LIVE cobra tree rather than from source text, so
// it also catches what a source grep would miss: a flag registered elsewhere,
// a shorthand added, a default changed.
package cli

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// frozenTodoSurface is the pre-swap surface: verb Use string -> its flags,
// each rendered "name=type(default)". Sorted, so the comparison is order-free.
//
// export-json is deliberately ABSENT. It is this SPEC's one addition
// (REQ-TOSQ-016), and the test asserts it is the ONLY one — an additive change
// is permitted, a silent second addition is not.
var frozenTodoSurface = map[string][]string{
	"add <text>":        {"force=bool(false)", "pick=bool(false)"},
	"list":              {"json=bool(false)"},
	"done <n>":          {},
	"next [<n>]":        {"expect=string()", "spec=string()"},
	"unpick <n>":        {},
	"analyze":           {},
	"drop <n> <reason>": {"expect=string()"},
	"undrop <n>":        {"expect=string()"},
	"edit <n> <text>":   {"expect=string()"},
	"move <n> (--top | --bottom | --before <m> | --after <m>)": {
		"after=string()", "before=string()", "bottom=bool(false)", "top=bool(false)",
	},
	"pr [<id>]": {"json=bool(false)"},
	"relate <a> <b> --relation <contains|absorbs|replaces|conflicts>": {
		"note=string()", "relation=string()",
	},
	"unrelate <index>": {},
	"why <n>":          {},
}

// theOnlyPermittedAddition is this SPEC's additive verb (REQ-TOSQ-016).
const theOnlyPermittedAddition = "export-json"

// dumpTodoSurface reads the live command tree.
func dumpTodoSurface(t *testing.T) map[string][]string {
	t.Helper()
	root := newTodoCmd()
	surface := make(map[string][]string, len(root.Commands()))
	for _, sub := range root.Commands() {
		var flags []string
		sub.Flags().VisitAll(func(f *pflag.Flag) {
			entry := fmt.Sprintf("%s=%s(%s)", f.Name, f.Value.Type(), f.DefValue)
			if f.Shorthand != "" {
				// A shorthand is part of the surface: adding one is a
				// re-flagging, even though the long form still works.
				entry += "/-" + f.Shorthand
			}
			flags = append(flags, entry)
		})
		sort.Strings(flags)
		if flags == nil {
			flags = []string{}
		}
		surface[sub.Use] = flags
	}
	return surface
}

// AC-TOSQ-010: the verb x flag surface shows ZERO deltas against the frozen
// pre-swap table, with export-json as the single permitted addition.
func TestTodoVerbSurfaceZeroDelta(t *testing.T) {
	t.Parallel()
	live := dumpTodoSurface(t)

	// (1) Nothing was removed or renamed.
	for use, wantFlags := range frozenTodoSurface {
		gotFlags, ok := live[use]
		if !ok {
			t.Errorf("verb %q is GONE from the surface — renamed or removed; every script that types it breaks", use)
			continue
		}
		if strings.Join(gotFlags, ",") != strings.Join(wantFlags, ",") {
			t.Errorf("verb %q re-flagged:\n  frozen: %v\n  live:   %v", use, wantFlags, gotFlags)
		}
	}

	// (2) Nothing was added except the one addition this SPEC declares.
	for use := range live {
		if _, frozen := frozenTodoSurface[use]; frozen {
			continue
		}
		if !strings.HasPrefix(use, theOnlyPermittedAddition) {
			t.Errorf("verb %q appeared on the surface and is not this SPEC's declared addition", use)
		}
	}

	// (3) The declared addition is actually present — the table would
	// otherwise pass vacuously if export-json had never been registered.
	var foundAddition bool
	for use := range live {
		if strings.HasPrefix(use, theOnlyPermittedAddition) {
			foundAddition = true
		}
	}
	if !foundAddition {
		t.Errorf("%s is not registered; the downgrade route is unreachable", theOnlyPermittedAddition)
	}

	// (4) Count check. Without it, a verb removed AND a verb added would
	// cancel out in the two loops above.
	if got, want := len(live), len(frozenTodoSurface)+1; got != want {
		t.Errorf("surface holds %d verbs, want %d (%d frozen + 1 addition): %v",
			got, want, len(frozenTodoSurface), sortedTodoVerbs(live))
	}
}

// AC-TOSQ-010 exit-code half: representative invocations return the same
// success/refusal verdict they did before the swap. A surface whose shape is
// frozen but whose refusals became successes is not a preserved surface.
func TestTodoVerbExitCodesUnchanged(t *testing.T) {
	todoFixture(t)
	if _, _, err := runTodo(t, "add", "first card"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"list succeeds", []string{"list"}, false},
		{"list --json succeeds", []string{"list", "--json"}, false},
		{"analyze succeeds", []string{"analyze"}, false},
		{"why on an existing card succeeds", []string{"why", "t1"}, false},
		{"done on a missing id is refused", []string{"done", "t99"}, true},
		{"unpick on a queued card is refused", []string{"unpick", "t1"}, true},
		{"next out of range is refused", []string{"next", "99"}, true},
		{"edit with empty text is refused", []string{"edit", "t1", ""}, true},
		{"drop without a reason is refused", []string{"drop", "t1"}, true},
		{"relate without --relation is refused", []string{"relate", "t1", "t1"}, true},
		{"unrelate out of range is refused", []string{"unrelate", "99"}, true},
		{"unknown verb is refused", []string{"nosuchverb"}, true},
		{"unknown flag is refused", []string{"list", "--nosuchflag"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := runTodo(t, tc.args...)
			if tc.wantErr && err == nil {
				t.Errorf("todo %v succeeded, want a refusal", tc.args)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("todo %v = %v, want success", tc.args, err)
			}
		})
	}
}

// sortedTodoVerbs renders a surface map's verbs for an error message.
func sortedTodoVerbs(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
