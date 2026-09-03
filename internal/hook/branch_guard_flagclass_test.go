package hook

// branch_guard_flagclass_test.go — SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001
// M1: the permanent synthetic measurement matrix driving matchBranchStateCommand
// with the full flag-form matrix of acceptance.md §D.1 (mutation forms M-01..M-33
// → deny, query forms Q-01..Q-17 → allow), the whole-token discrimination pairs
// (P-01/P-02), and the edge cases (§D.3 E-1..E-6).
//
// Expectations are the doctrine-based TARGET (post-M2) classifications; §D.1 is
// the normative expectation authority. On the pre-fix tree every row whose
// expectation is deny and whose pre-fix behavior was allow FAILs (RED — the
// under-match defect this SPEC fixes).
//
// @MX:SPEC: SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001
// @MX:TEST: M1 measurement matrix — classification authority for the git branch matcher

import "testing"

// TestBranchGuardFlagClassMatrix is the M1 permanent measurement matrix.
// Cell IDs trace 1:1 to acceptance.md §D.1/§D.3 rows.
func TestBranchGuardFlagClassMatrix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		id      string
		command string
		want    bool // true = deny (pattern matches); false = allow
	}{
		// ---- Mutation forms (§D.1 M-01..M-33) → deny ----
		{"M-01", "git branch feature", true},                       // bare creation
		{"M-02", "git branch newbranch oldstart", true},            // creation at start point
		{"M-03a", "git branch -d old", true},                       // delete
		{"M-03b", "git branch -D old", true},                       // force delete
		{"M-04a", "git branch -m renamed", true},                   // move
		{"M-04b", "git branch -M renamed", true},                   // force move
		{"M-05a", "git branch -c copied", true},                    // copy
		{"M-05b", "git branch -C copied", true},                    // force copy
		{"M-06", "git branch -f topic abc1234", true},              // force pointer rewrite
		{"M-07", "git branch -f topic", true},                      // force create at HEAD
		{"M-08", "git branch --force topic abc1234", true},         // long force
		{"M-09", "git branch -df old", true},                       // combined cluster
		{"M-10", "git branch -fm renamed", true},                   // combined cluster
		{"M-11", "git branch -vD old", true},                       // combined cluster (mid)
		{"M-12", "git branch -u origin/main topic", true},          // set upstream
		{"M-13", "git branch --set-upstream-to=origin/main topic", true}, // attached value
		{"M-14", "git branch --set-upstream origin/main topic", true},    // removed-in-git form, fail-closed-safe pin
		{"M-15", "git branch --unset-upstream topic", true},
		{"M-16", "git branch -t topic origin/main", true},            // track
		{"M-17", "git branch --track topic origin/main", true},
		{"M-18", "git branch --no-track topic origin/main", true},
		{"M-19", "git branch --edit-description topic", true},
		{"M-20", "git branch --delete old", true},
		{"M-21a", "git branch --move renamed", true},
		{"M-21b", "git branch --copy copied", true},
		{"M-22", "git branch --set-upstream-to origin/main topic", true}, // space-separated value
		{"M-23", "git branch --track=direct topic abc1234", true},        // attached value
		{"M-24", "git branch -umain topic", true},                        // attached short value
		{"M-25", "git branch -v vbranch", true},                          // flag + name = create (measured)
		{"M-26", "git branch -- cr2", true},                              // option-prefixed creation
		{"M-27", "git branch -q qbranch", true},                          // option-prefixed creation
		{"M-28", "git branch --no-force nfbranch", true},                 // option-prefixed creation
		{"M-29", "git branch -vt vtbranch main", true},                   // cluster containing t
		{"M-30", "git branch --create-reflog crbranch", true},            // modifier + creation operand
		{"M-31", "git branch --color colprobe", true},                    // attached-only flag; space token = positional
		{"M-32", "git branch --abbrev 12 abbranch", true},                // attached-only flag; positionals
		{"M-33", "git branch -vux main x", true},                         // mid-cluster u

		// ---- Query forms (§D.1 Q-01..Q-17) → allow ----
		{"Q-01", "git branch", false},
		{"Q-02a", "git branch --list", false},
		{"Q-02b", "git branch --list develop -v", false},
		{"Q-03a", "git branch -v", false},
		{"Q-03b", "git branch -vv", false},
		{"Q-04a", "git branch -a", false},
		{"Q-04b", "git branch -r", false},
		{"Q-05", "git branch --show-current", false},
		{"Q-06", "git branch --contains HEAD", false},
		{"Q-07", "git branch --merged main", false},
		{"Q-08", "git branch --no-merged main", false},
		{"Q-09", "git branch --points-at HEAD", false},
		{"Q-10", "git branch --format %(refname)", false},
		{"Q-11", "git branch --sort=-committerdate", false},
		{"Q-12a", "git branch -q", false},
		{"Q-12b", "git branch -i", false},
		{"Q-13", "git branch --sort committerdate", false}, // space-separated value consumed
		{"Q-14", "git branch --no-contains HEAD", false},   // space-separated value consumed
		{"Q-15a", "git branch -l lpattern", false},         // short list selector
		{"Q-15b", "git branch -l foo bar", false},          // variadic list mode
		{"Q-16", "git branch --contains HEAD main", false}, // filter selector → pattern, not creation
		{"Q-17", "git branch --color=always", false},       // attached-only value consumed, no positional

		// ---- Whole-token discrimination pairs (§D.1 P-01/P-02) ----
		// P-01: --format (query) vs --force (mutation) — no prefix matching.
		{"P-01a", "git branch --format %(refname)", false},
		{"P-01b", "git branch --force topic abc1234", true},
		// P-02: embedded letters in query long flags must not trip a character class.
		{"P-02a", "git branch --contains HEAD", false},
		{"P-02b", "git branch --merged main", false},
		{"P-02c", "git branch --no-merged main", false},

		// ---- Edge cases (§D.3 E-1..E-6) ----
		{"E-1", "git branch -d old -v", true},                        // mutating flag dominates
		{"E-2", "git branch --set-upstream-to=origin/main topic", true}, // = value not a creation operand
		{"E-3a", "git branch -F topic", true},                        // case-fold of known mutation flag
		{"E-3b", "git branch --FORCE topic abc1234", true},           // case-fold of known mutation flag
		{"E-4", "git branch -D 'old-feature'", true},                 // quoted operand still denies
		{"E-5a", "git branch --dele old", false},                     // abbreviation prefix → unknown → allow
		{"E-5b", "git branch -z foo", false},                         // unknown flag → allow (fail-open)
		{"E-6", "git branch -f x y && git status", true},             // compound command segment
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.id+" "+tc.command, func(t *testing.T) {
			t.Parallel()
			suffix, matched := matchBranchStateCommand(tc.command)
			if matched != tc.want {
				t.Fatalf("matchBranchStateCommand(%q) = (%q, %v), want matched=%v (§D.1 %s)",
					tc.command, suffix, matched, tc.want, tc.id)
			}
			if matched && suffix == "" {
				t.Fatalf("matched but suffix empty for %q", tc.command)
			}
		})
	}
}
