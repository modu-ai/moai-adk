package cli

// shelljoin_red_t1582_test.go — plan-phase probe for card t1582 item ②,
// half one (the execution side). t1576's R4-3 added
// shellJoinArgs/shellQuoteArg for EXECUTION quoting; the same-surface
// classifiers live in the factory package (internal/factory/
// integration_remeasure.go: shellSegments/shellFields/isGoTestCommand —
// the card text named the cli path, the live code moved them). This file
// pins the EXECUTION half: the joined line must preserve every argument
// boundary in its quoting, which is what sh re-splits into argv. The
// re-parse half (do the classifiers read those boundaries back?) runs as
// the round trip in internal/factory/remeasure_red_t1582_test.go — the two
// files together are the classification-vs-execution divergence probe.

import (
	"strings"
	"testing"
)

func TestRedT1582JoinPreservesArgBoundariesInQuoting(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantJoin string
	}{
		{
			name:     "regex with spaces and pipe",
			args:     []string{"go", "test", "-json", "-run", "Test A|Test B", "./pkg"},
			wantJoin: "go test -json -run 'Test A|Test B' ./pkg",
		},
		{
			name:     "regex with semicolon",
			args:     []string{"go", "test", "-json", "-run", "Test A;Test B", "./pkg"},
			wantJoin: "go test -json -run 'Test A;Test B' ./pkg",
		},
		{
			name:     "regex with ampersand",
			args:     []string{"go", "test", "-json", "-run", "Test A&Test B", "./pkg"},
			wantJoin: "go test -json -run 'Test A&Test B' ./pkg",
		},
		{
			name:     "env assignment value with spaces",
			args:     []string{"GOFLAGS=-json -v", "go", "test", "./pkg"},
			wantJoin: "GOFLAGS='-json -v' go test ./pkg",
		},
		{
			name:     "apostrophe inside argument",
			args:     []string{"go", "test", "-json", "-run", "it's fine", "./pkg"},
			wantJoin: "go test -json -run 'it'\\''s fine' ./pkg",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			line := shellJoinArgs(tc.args)
			t.Logf("joined line: %s", line)
			if line != tc.wantJoin {
				t.Fatalf("RED t1582-item2 (execution half): the join lost an argument boundary — got %q, want %q", line, tc.wantJoin)
			}
			if strings.Contains(line, "\x00") {
				t.Fatal("the join must not leak a NUL into the shell line")
			}
		})
	}
}

// The env-scrub compound single argument: one shell line, passed through
// verbatim (the t1576 R8-2 contract), and the env-scrub compound form the
// AGENTS.md §4 verification form prescribes must survive the join.
func TestRedT1582SingleArgScrubCompoundStaysVerbatim(t *testing.T) {
	line := shellJoinArgs([]string{"unset VARS && go test -json ./pkg"})
	if line != "unset VARS && go test -json ./pkg" {
		t.Fatalf("the single argument is the shell line the caller meant and passes through verbatim: %q", line)
	}
}
