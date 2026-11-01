package cli

// N55, the explicit first bind (SPEC-AGENTIC-FACTORY-LOCAL-001 REQ-AFL-001, AC-AFL-001 cases (e5), (e5') and (e5'')): a verb
// that wrote nothing (a reap with nothing to reap) and carries no --bind-repo does not bind an unbound ledger. The first
// bind then comes from the existing --bind-repo declaration, or from a verb that writes and completes. The declaration
// keeps every check it had: a repository other than the one the call would bind, and a path in no repository, are refused
// with nothing changed. Real `moai af` verbs, real ledgers, real git.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// afN55Kind is one way of placing the ledger: the file inside the project's repository (--db), or the default location.
type afN55Kind struct {
	name string
	// place returns the ledger path and the arguments that name it ("" and none for the default location)
	place func(t *testing.T, repo string) (db string, dbArgs []string)
}

func afN55Kinds() []afN55Kind {
	return []afN55Kind{
		{"a ledger file inside the repository", func(t *testing.T, repo string) (string, []string) {
			db := filepath.Join(repo, ".moai", "af", "ledger.db")
			return db, []string{"--db", db}
		}},
		{"the default location", func(t *testing.T, _ string) (string, []string) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("MOAI_HOME", filepath.Join(t.TempDir(), ".moai"))
			db, err := afLedgerPath("")
			if err != nil {
				t.Fatal(err)
			}
			return db, nil
		}},
	}
}

// (e5): the verb ends, the ledger stays unbound, and a verb that writes binds it afterwards.
func TestAFVerbs_N55AnEmptyHandedVerbWithoutADeclarationLeavesTheLedgerUnbound(t *testing.T) {
	afSkipWithoutSh(t)
	for _, k := range afN55Kinds() {
		t.Run(k.name, func(t *testing.T) {
			repo, _ := afN38Other(t)
			t.Setenv("CLAUDE_PROJECT_DIR", repo)
			db, dbArgs := k.place(t, repo)
			st := afN49Ledger(t, db)
			out, errOut, err := afRunBoth(append([]string{"reap"}, dbArgs...)...)
			if err != nil {
				t.Fatalf("reap with nothing to reap: %v\n%.300s %.300s", err, out, errOut)
			}
			if got := afN49Bound(t, st); got != "" {
				t.Fatalf("a reap that wrote nothing and declared nothing bound the ledger to %q", got)
			}
			// the binding is not lost, only put off: the first verb that writes makes it
			if out, errOut, err := afRunBoth(append([]string{"claim", "--session", "s1", "--worktree", repo}, dbArgs...)...); err != nil {
				t.Fatalf("claim: %v\n%.300s %.300s", err, out, errOut)
			}
			if got, want := afN49Bound(t, st), afN38GitDir(t, repo); got != want {
				t.Errorf("after a claim bound %q, want %q", got, want)
			}
		})
	}
}

// (e5'): the same verb with --bind-repo of a path in the right repository binds it, whatever the path is (a directory
// inside it, a linked worktree of it).
func TestAFVerbs_N55AnEmptyHandedVerbThatDeclaresTheRepositoryBindsIt(t *testing.T) {
	afSkipWithoutSh(t)
	for _, k := range afN55Kinds() {
		for _, how := range []string{"a directory inside it", "a linked worktree of it"} {
			t.Run(k.name+"/"+how, func(t *testing.T) {
				repo, _ := afN38Other(t)
				t.Setenv("CLAUDE_PROJECT_DIR", repo)
				declared := filepath.Join(repo, "sub")
				if how == "a linked worktree of it" {
					declared = filepath.Join(t.TempDir(), "wt")
					afGit(t, repo, "worktree", "add", "-q", "--detach", declared)
				} else if err := os.MkdirAll(declared, 0o755); err != nil {
					t.Fatal(err)
				}
				db, dbArgs := k.place(t, repo)
				st := afN49Ledger(t, db)
				out, errOut, err := afRunBoth(append([]string{"reap", "--bind-repo", declared}, dbArgs...)...)
				if err != nil {
					t.Fatalf("reap declaring %s: %v\n%.300s %.300s", declared, err, out, errOut)
				}
				if got, want := afN49Bound(t, st), afN38GitDir(t, repo); got != want {
					t.Errorf("bound %q, want %q", got, want)
				}
			})
		}
	}
}

// (e5”): a declaration naming another repository, or a path in no repository, is refused with nothing changed, and the
// ledger stays unbound.
func TestAFVerbs_N55ADeclarationThatDoesNotFitIsRefusedAndTheLedgerStaysUnbound(t *testing.T) {
	afSkipWithoutSh(t)
	for _, k := range afN55Kinds() {
		for _, tc := range []struct {
			name string
			path func(t *testing.T) string
		}{
			{"another repository", func(t *testing.T) string { r, _ := afN38Other(t); return r }},
			{"a path in no repository", func(t *testing.T) string { return t.TempDir() }},
		} {
			t.Run(k.name+"/"+tc.name, func(t *testing.T) {
				repo, _ := afN38Other(t)
				t.Setenv("CLAUDE_PROJECT_DIR", repo)
				db, dbArgs := k.place(t, repo)
				st := afN49Ledger(t, db)
				before := afSnapshot(t, st)
				out, errOut, err := afRunBoth(append([]string{"reap", "--bind-repo", tc.path(t)}, dbArgs...)...)
				if err == nil {
					t.Fatalf("a declaration of %s was accepted: %.300s %.300s", tc.name, out, errOut)
				}
				if !strings.Contains(err.Error(), "nothing was changed") {
					t.Errorf("the refusal does not say nothing was changed: %v", err)
				}
				if got := afN49Bound(t, st); got != "" {
					t.Errorf("the refused call bound the ledger to %q", got)
				}
				if after := afSnapshot(t, st); !reflect.DeepEqual(before, after) {
					t.Errorf("the refused call changed the ledger:\nbefore %+v\nafter  %+v", before, after)
				}
			})
		}
	}
}
