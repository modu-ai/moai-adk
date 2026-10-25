package cli

// integration_remeasure_basebranch_test.go — card t1790, GH #1817.
//
// `moai integration remeasure -- <command>` died with `merge-base ..HEAD …
// exit status 128` in a clean linked worktree. The integration branch is read
// from git-strategy.yaml, which is gitignored, so a checkout that never had the
// file resolved an empty branch and the verb called git with an empty ref. The
// shapes below pin the three states the configuration can be in; the verb must
// resolve a base in the first two and refuse with guidance (never call git with
// an empty ref) in the third.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// remeasureBaseFixture builds origin.git (main), a primary clone of it and a
// clean linked worktree on a card branch whose HEAD equals main. withOrigin
// false gives a primary that has no remote at all. The returned linked path is
// the tree the verb runs from.
func remeasureBaseFixture(t *testing.T, withOrigin bool) (primary, linked string) {
	t.Helper()
	base := t.TempDir()
	seed := filepath.Join(base, "seed")
	if err := os.MkdirAll(seed, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, seed, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(seed, "seed.txt"), []byte("seed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitRun(t, seed, "add", "seed.txt")
	gitRun(t, seed, "commit", "-q", "-m", "seed")
	primary = filepath.Join(base, "primary")
	if withOrigin {
		origin := filepath.Join(base, "origin.git")
		gitRun(t, base, "clone", "-q", "--bare", seed, origin)
		gitRun(t, base, "clone", "-q", origin, primary)
	} else {
		gitRun(t, base, "clone", "-q", seed, primary)
		gitRun(t, primary, "remote", "remove", "origin")
	}
	linked = filepath.Join(base, "linked")
	gitRun(t, primary, "worktree", "add", "-q", "-b", "card", linked, "main")
	return primary, linked
}

// writeGitignoredStrategy writes git-strategy.yaml into root's gitignored
// .moai/config/sections (the primary's untracked, locally excluded copy).
func writeGitignoredStrategy(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	exclude := filepath.Join(root, ".git", "info", "exclude")
	existing, _ := os.ReadFile(exclude)
	if err := os.WriteFile(exclude, append(existing, []byte("\n.moai/\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationRemeasureResolvesBaseWithoutLinkedConfig(t *testing.T) {
	const githubFlow = "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n"
	for _, tc := range []struct {
		name       string
		withOrigin bool
		primaryCfg string // "" = no configuration anywhere
		wantErr    string // substring of the refusal; "" = the verb must succeed
		wantOut    string
	}{
		{
			// Positive control: the configuration lives in the primary (gitignored)
			// and the linked tree has none; the common root answers.
			name: "primary holds the gitignored configuration", withOrigin: true, primaryCfg: githubFlow,
			wantOut: "re-measure recorded",
		},
		{
			// The reported failure: no configuration anywhere, a remote default
			// branch exists — the base resolves from origin/HEAD.
			name: "no configuration, origin/HEAD resolves the default branch", withOrigin: true,
			wantOut: "re-measure recorded",
		},
		{
			// Nothing to resolve a base from: refuse with guidance, naming the
			// missing configuration — no git call with an empty ref.
			name: "no configuration and no remote default branch", withOrigin: false,
			wantErr: "no integration branch",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			primary, linked := remeasureBaseFixture(t, tc.withOrigin)
			if tc.primaryCfg != "" {
				writeGitignoredStrategy(t, primary, tc.primaryCfg)
			}
			// The fixture's own runtime state must not read as a dirty tree.
			exclude := filepath.Join(primary, ".git", "info", "exclude")
			existing, _ := os.ReadFile(exclude)
			if err := os.WriteFile(exclude, append(existing, []byte("\n.moai/\n.claude/\n")...), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Chdir(linked)
			out, err := runIntegration(t, linked, "remeasure", "--", "true")
			observed := out
			if err != nil {
				observed += err.Error()
			}
			if strings.Contains(observed, "merge-base") && strings.Contains(observed, "exit status 128") {
				t.Fatalf("the verb called git with an empty ref (GH #1817): %q", observed)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(observed, tc.wantErr) {
					t.Fatalf("want a refusal containing %q, got err=%v out=%q", tc.wantErr, err, out)
				}
				return
			}
			if err != nil || !strings.Contains(out, tc.wantOut) {
				t.Fatalf("want success containing %q, got err=%v out=%q", tc.wantOut, err, out)
			}
		})
	}
}
