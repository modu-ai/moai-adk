package cli

// integration_remeasure_basebranch_test.go — card t1790, GH #1817.
//
// `moai integration remeasure -- <command>` died with `merge-base ..HEAD …
// exit status 128` in a clean linked worktree. The integration branch is read
// from git-strategy.yaml, which is gitignored, so a checkout that never had the
// file resolved an empty branch and the verb called git with an empty ref. The
// rows below pin every state the configuration and the remote can be in: the
// verb resolves a base where it can, refuses with guidance where it cannot, and
// in every refusal the measured command has NOT run (a marker file proves it).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// remeasureBaseFixture builds origin.git (main), a primary clone of it and a
// clean linked worktree on a card branch. withOrigin false gives a primary with
// no remote at all. developAhead adds a local develop branch one commit ahead
// of main and cuts the card branch from it, so the fork point against develop
// differs from the one against main; developTip is develop's tip then.
func remeasureBaseFixture(t *testing.T, withOrigin, developAhead bool) (primary, linked, developTip string) {
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
	from := "main"
	if developAhead {
		gitRun(t, primary, "switch", "-q", "-c", "develop")
		if err := os.WriteFile(filepath.Join(primary, "develop.txt"), []byte("develop\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		gitRun(t, primary, "add", "develop.txt")
		gitRun(t, primary, "commit", "-q", "-m", "develop ahead of main")
		developTip = strings.TrimSpace(gitRun(t, primary, "rev-parse", "HEAD"))
		gitRun(t, primary, "switch", "-q", "main")
		from = "develop"
	}
	gitRun(t, primary, "worktree", "add", "-q", "-b", "card", linked, from)
	return primary, linked, developTip
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
}

func TestIntegrationRemeasureResolvesBaseWithoutLinkedConfig(t *testing.T) {
	const (
		githubFlow    = "git_strategy:\n    mode: manual\n    manual:\n        workflow: github-flow\n"
		gitFlowDevel  = "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n"
		gitFlowEmpty  = "git_strategy:\n    mode: manual\n    manual:\n        workflow: git-flow\n        develop_branch: \"\"\n"
		gitFlowPerson = "git_strategy:\n    mode: personal\n    personal:\n        workflow: git-flow\n        develop_branch: develop\n"
		unknownFlow   = "git_strategy:\n    mode: manual\n    manual:\n        workflow: svn-flow\n"
		brokenYAML    = "git_strategy: [unterminated\n"
		noMode        = "git_strategy:\n    manual:\n        workflow: git-flow\n        develop_branch: develop\n"
	)
	for _, tc := range []struct {
		name         string
		withOrigin   bool
		developAhead bool
		primaryCfg   string // "" = no configuration anywhere
		dropMain     bool   // delete the local main branch (origin/main stays)
		extraBranch  string // an extra local branch cut from main, e.g. "origin/main"
		emptyFile    bool   // the configuration file exists and is empty
		chmod000     bool   // the configuration file exists but cannot be read
		wantErr      string // substring of the refusal; "" = the verb must succeed
		wantOut      string
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
			// A local branch named origin/main makes the short form of the symbolic
			// ref print remotes/origin/main; the full ref is read, so the setup
			// still resolves to main (sync-audit iter1 F2).
			name: "a local branch named origin/main does not break the fallback", withOrigin: true, extraBranch: "origin/main",
			wantOut: "re-measure recorded",
		},
		{
			// A configured target beats origin/HEAD: the recorded base is develop's
			// fork point, not main's (sync-audit iter1 F4).
			name: "the configured target outranks origin/HEAD", withOrigin: true, developAhead: true, primaryCfg: gitFlowDevel,
			wantOut: "re-measure recorded",
		},
		{
			// Nothing to resolve a base from: refuse with guidance, naming the
			// missing configuration — no git call with an empty ref.
			name: "no configuration and no remote default branch", withOrigin: false,
			wantErr: "no integration branch",
		},
		{
			// card-review (codex): origin/HEAD names main but the clone has no
			// local main, so the base ref would fail AFTER the command. The verb
			// refuses first, naming the missing local branch.
			name: "origin/HEAD names a branch with no local ref", withOrigin: true, dropMain: true,
			wantErr: "has no local branch refs/heads/main",
		},
		{
			// sync-audit iter1 F1: a file that LOADS but resolves no target is a
			// misconfiguration, not an absent file — origin/HEAD must not stand in.
			name: "git-flow with an empty develop_branch is refused", withOrigin: true, primaryCfg: gitFlowEmpty,
			wantErr: "develop_branch is empty",
		},
		{
			name: "git-flow outside manual mode is refused", withOrigin: true, primaryCfg: gitFlowPerson,
			wantErr: "git-flow resolves a target only when git_strategy.mode is manual",
		},
		{
			name: "an unknown workflow is refused", withOrigin: true, primaryCfg: unknownFlow,
			wantErr: "svn-flow",
		},
		{
			// sync-audit iter2 R1: the fallback is for an ABSENT file only. A file
			// that exists but classifies as unreadable is a misconfiguration too.
			name: "a present but unparsable file is refused", withOrigin: true, primaryCfg: brokenYAML,
			wantErr: "cannot read",
		},
		{
			name: "a present but empty file is refused", withOrigin: true, emptyFile: true,
			wantErr: "cannot read",
		},
		{
			name: "a present file with no git_strategy.mode is refused", withOrigin: true, primaryCfg: noMode,
			wantErr: "cannot read",
		},
		{
			name: "a present file that cannot be read is refused", withOrigin: true, primaryCfg: githubFlow, chmod000: true,
			wantErr: "cannot read",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			primary, linked, developTip := remeasureBaseFixture(t, tc.withOrigin, tc.developAhead)
			if tc.primaryCfg != "" || tc.emptyFile {
				writeGitignoredStrategy(t, primary, tc.primaryCfg)
			}
			if tc.chmod000 {
				cfgPath := filepath.Join(primary, ".moai", "config", "sections", "git-strategy.yaml")
				if err := os.Chmod(cfgPath, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(cfgPath, 0o600) })
				// A 0000 file is only unreadable to a non-root caller: as root (or
				// where the mode is not enforced) the row would prove nothing.
				if f, openErr := os.Open(cfgPath); openErr == nil {
					_ = f.Close()
					t.Skip("this environment reads a 0000 file (running as root, or the mode is not enforced), so the row cannot show an unreadable configuration")
				}
			}
			if tc.extraBranch != "" {
				gitRun(t, primary, "branch", tc.extraBranch, "main")
			}
			if tc.dropMain {
				// The primary leaves main (the linked tree stays on its card
				// branch), then main is deleted: only origin/main remains.
				gitRun(t, primary, "switch", "-q", "-c", "elsewhere")
				gitRun(t, primary, "branch", "-q", "-D", "main")
			}
			// The fixture's own runtime state must not read as a dirty tree.
			exclude := filepath.Join(primary, ".git", "info", "exclude")
			existing, _ := os.ReadFile(exclude)
			if err := os.WriteFile(exclude, append(existing, []byte("\n.moai/\n.claude/\n")...), 0o600); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(t.TempDir(), "ran")
			t.Chdir(linked)
			out, err := runIntegration(t, linked, "remeasure", "--", "touch "+marker)
			observed := out
			if err != nil {
				observed += err.Error()
			}
			if strings.Contains(observed, "merge-base") && strings.Contains(observed, "exit status 128") {
				t.Fatalf("the verb died in the fork-point lookup (exit status 128) instead of resolving or refusing first (GH #1817): %q", observed)
			}
			_, ranErr := os.Stat(marker)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(observed, tc.wantErr) {
					t.Fatalf("want a refusal containing %q, got err=%v out=%q", tc.wantErr, err, out)
				}
				if ranErr == nil {
					t.Fatalf("the measured command ran although the verb refused: %q", observed)
				}
				return
			}
			if err != nil || !strings.Contains(out, tc.wantOut) {
				t.Fatalf("want success containing %q, got err=%v out=%q", tc.wantOut, err, out)
			}
			if ranErr != nil {
				t.Fatalf("the verb succeeded but the measured command did not run: %v", ranErr)
			}
			if tc.developAhead {
				if want := "(base " + developTip[:12] + ")"; !strings.Contains(out, want) {
					t.Fatalf("the recorded base must be develop's fork point %s, got %q", want, out)
				}
			}
		})
	}
}
