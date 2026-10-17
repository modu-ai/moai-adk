package cli

// integration_bisect_test.go — card t1833: `moai integration bisect` finds the
// card-level culprit with git bisect --first-parent. Every case builds a real
// repository whose main carries --no-ff landings written the way the merge
// step writes them, so the verb is judged on the history shape it will meet.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bisectPostCheckoutHook installs a post-checkout hook in the repository that
// owns primary. The hook body is a shell fragment; $2 is the commit checked out.
func bisectPostCheckoutHook(t *testing.T, primary, body string) {
	t.Helper()
	hooks := filepath.Join(primary, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooks, "post-checkout"), []byte("#!/bin/sh\n"+body+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
}

// bisectRepo builds main with five card landings; landing 3 (card t1003)
// introduces bad.txt, and each card branch carries two inner commits so a
// first-parent walk and a full-history walk disagree. It returns the primary
// checkout (parked off main so main can be checked out elsewhere), a linked
// worktree detached at the main tip, the base commit and the tip.
func bisectRepo(t *testing.T) (primary, wt, base, tip string) {
	t.Helper()
	primary = filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(primary, 0o700); err != nil {
		t.Fatal(err)
	}
	fcGit(t, primary, "init", "-q", "-b", "main")
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(primary, name), []byte(name+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, primary, "add", name)
	}
	write("base.txt")
	fcGit(t, primary, "commit", "-q", "-m", "base")
	base = fcGit(t, primary, "rev-parse", "HEAD")
	for i := 1; i <= 5; i++ {
		branch := fmt.Sprintf("WT-card-%d", i)
		fcGit(t, primary, "checkout", "-q", "-b", branch, "main")
		if i == 3 {
			write("bad.txt")
		} else {
			write(fmt.Sprintf("ok%d.txt", i))
		}
		fcGit(t, primary, "commit", "-q", "-m", fmt.Sprintf("work %d a", i))
		write(fmt.Sprintf("extra%d.txt", i))
		fcGit(t, primary, "commit", "-q", "-m", fmt.Sprintf("work %d b", i))
		fcGit(t, primary, "checkout", "-q", "main")
		fcGit(t, primary, "merge", "-q", "--no-ff", "-m",
			fmt.Sprintf("Merge %s into main (card t100%d, integration merge)", branch, i), branch)
	}
	tip = fcGit(t, primary, "rev-parse", "HEAD")
	wt = filepath.Join(t.TempDir(), "wt")
	fcGit(t, primary, "worktree", "add", "-q", "--detach", wt, tip)
	fcGit(t, primary, "checkout", "-q", "-b", "parking", base)
	return primary, wt, base, tip
}

func TestBisectNamesTheCulpritCard(t *testing.T) {
	_, wt, base, tip := bisectRepo(t)
	var out strings.Builder
	err := runIntegrationBisect(&out, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "test ! -f bad.txt"})
	if err != nil {
		t.Fatalf("bisect: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "card: t1003") {
		t.Fatalf("culprit must be card t1003:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Merge WT-card-3 into main (card t1003, integration merge)") {
		t.Fatalf("the first-parent merge subject must be printed (an inner commit was named instead?):\n%s", out.String())
	}
	if head := fcGit(t, wt, "rev-parse", "HEAD"); head != tip {
		t.Fatalf("HEAD must be restored to %s, is %s", tip, head)
	}
	if _, statErr := os.Stat(filepath.Join(wt, ".git")); statErr != nil {
		t.Fatal(statErr)
	}
	if log, _ := runScrubbedGit(wt, "bisect", "log"); strings.TrimSpace(log) != "" {
		t.Fatalf("the bisect must be reset, log still reads:\n%s", log)
	}
}

func TestBisectRefusesThePrimaryCheckout(t *testing.T) {
	primary, _, base, tip := bisectRepo(t)
	before := fcGit(t, primary, "rev-parse", "HEAD")
	var out strings.Builder
	err := runIntegrationBisect(&out, integrationBisectOpts{Dir: primary, Bad: tip, Good: base, Command: "true"})
	if err == nil || !strings.Contains(err.Error(), "primary checkout") {
		t.Fatalf("the primary checkout must be refused, got %v", err)
	}
	if after := fcGit(t, primary, "rev-parse", "HEAD"); after != before {
		t.Fatalf("the refusal must not move HEAD: %s -> %s", before, after)
	}
}

func TestBisectRefusesADirtyTree(t *testing.T) {
	_, wt, base, tip := bisectRepo(t)
	if err := os.WriteFile(filepath.Join(wt, "dirt.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := runIntegrationBisect(&strings.Builder{}, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "true"})
	if err == nil || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("a dirty tree must be refused, got %v", err)
	}
}

func TestBisectRefusesWhenOneIsInProgress(t *testing.T) {
	_, wt, base, tip := bisectRepo(t)
	fcGit(t, wt, "bisect", "start", tip, base)
	err := runIntegrationBisect(&strings.Builder{}, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "true"})
	if err == nil || !strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("a bisect in progress must be refused, got %v", err)
	}
}

func TestBisectDryRunPrintsTheGitCommandsAndMovesNothing(t *testing.T) {
	_, wt, base, tip := bisectRepo(t)
	var out strings.Builder
	if err := runIntegrationBisect(&out, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "test ! -f bad.txt", DryRun: true}); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	for _, want := range []string{"git bisect start --first-parent " + tip + " " + base, "git bisect run sh -c test ! -f bad.txt", "git bisect reset"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("dry-run must print %q:\n%s", want, out.String())
		}
	}
	if head := fcGit(t, wt, "rev-parse", "HEAD"); head != tip {
		t.Fatalf("dry-run must not move HEAD: %s", head)
	}
	if log, _ := runScrubbedGit(wt, "bisect", "log"); strings.TrimSpace(log) != "" {
		t.Fatalf("dry-run must start no bisect:\n%s", log)
	}
}

func TestBisectRequiresGoodAndAResolvableRef(t *testing.T) {
	_, wt, _, tip := bisectRepo(t)
	if err := runIntegrationBisect(&strings.Builder{}, integrationBisectOpts{Dir: wt, Bad: tip, Command: "true"}); err == nil || !strings.Contains(err.Error(), "--good") {
		t.Fatalf("a missing --good must be refused, got %v", err)
	}
	if err := runIntegrationBisect(&strings.Builder{}, integrationBisectOpts{Dir: wt, Bad: tip, Good: "no-such-ref", Command: "true"}); err == nil || !strings.Contains(err.Error(), "no-such-ref") {
		t.Fatalf("an unresolvable ref must be refused by name, got %v", err)
	}
}

func TestBisectReportsAnAbortedSearchAndStillResets(t *testing.T) {
	_, wt, base, tip := bisectRepo(t)
	err := runIntegrationBisect(&strings.Builder{}, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "exit 130"})
	if err == nil || !strings.Contains(err.Error(), "did not converge") {
		t.Fatalf("an aborted search must not name a culprit, got %v", err)
	}
	if head := fcGit(t, wt, "rev-parse", "HEAD"); head != tip {
		t.Fatalf("HEAD must be restored after an abort: %s", head)
	}
	if log, _ := runScrubbedGit(wt, "bisect", "log"); strings.TrimSpace(log) != "" {
		t.Fatalf("an aborted bisect must still be reset:\n%s", log)
	}
}

func TestBisectCmdRejectsAnEmptyCommand(t *testing.T) {
	cmd := newIntegrationBisectCmd()
	cmd.SetArgs([]string{"--good", "HEAD"})
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "after --") {
		t.Fatalf("a bisect without a command must be refused, got %v", err)
	}
}

// F1 (card t1833 fix round 1): the culprit comes from git's own bisect state,
// never from text the judged command printed.
func TestBisectIgnoresACulpritClaimPrintedByTheJudgedCommand(t *testing.T) {
	_, wt, base, tip := bisectRepo(t)
	decoy := fcGit(t, wt, "log", "--first-parent", "--format=%H", "--grep", "card t1001", "-n1")
	var out strings.Builder
	cmdline := fmt.Sprintf("echo %s is the first bad commit; test ! -f bad.txt", decoy)
	if err := runIntegrationBisect(&out, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: cmdline}); err != nil {
		t.Fatalf("bisect: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "card: t1003") {
		t.Fatalf("a decoy line printed by the judged command must not change the culprit; want card t1003:\n%s", out.String())
	}
	if strings.Contains(out.String(), "card: t1001") {
		t.Fatalf("the decoy card was reported:\n%s", out.String())
	}
}

// F2: with GIT_DIR / GIT_WORK_TREE aimed at another repository, the pre-checks
// and the bisect itself must still look at the same tree.
func TestBisectActsOnTheCheckedTreeWhenGitDirIsSet(t *testing.T) {
	primary, wt, base, tip := bisectRepo(t)
	other := filepath.Join(t.TempDir(), "other")
	if err := os.MkdirAll(other, 0o700); err != nil {
		t.Fatal(err)
	}
	fcGit(t, other, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(other, "o.txt"), []byte("o\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, other, "add", "o.txt")
	fcGit(t, other, "commit", "-q", "-m", "other")
	otherHead := fcGit(t, other, "rev-parse", "HEAD")
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	if err := runIntegrationBisect(&strings.Builder{}, integrationBisectOpts{Dir: primary, Bad: tip, Good: base, Command: "true"}); err == nil || !strings.Contains(err.Error(), "primary checkout") {
		t.Fatalf("the primary checkout must still be refused with GIT_DIR set, got %v", err)
	}
	var out strings.Builder
	if err := runIntegrationBisect(&out, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "test ! -f bad.txt"}); err != nil {
		t.Fatalf("bisect with GIT_DIR set: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "card: t1003") {
		t.Fatalf("culprit must be card t1003:\n%s", out.String())
	}
	if head, _ := runScrubbedGit(other, "rev-parse", "HEAD"); strings.TrimSpace(head) != otherHead {
		t.Fatalf("the other repository must not move: %s -> %s", otherHead, head)
	}
	if log, _ := runScrubbedGit(other, "bisect", "log"); strings.TrimSpace(log) != "" {
		t.Fatalf("the other repository must not be bisected:\n%s", log)
	}
	if head, _ := runScrubbedGit(wt, "rev-parse", "HEAD"); strings.TrimSpace(head) != tip {
		t.Fatalf("the card tree must be restored to %s, is %s", tip, head)
	}
}

// F3: a reset that fails leaves the tree mid-bisect, and the verb must say so
// instead of exiting 0.
func TestBisectFailsWhenTheResetFails(t *testing.T) {
	primary, wt, base, tip := bisectRepo(t)
	bisectPostCheckoutHook(t, primary, fmt.Sprintf(`[ "$2" = "%s" ] && exit 1; exit 0`, tip))
	var out strings.Builder
	err := runIntegrationBisect(&out, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "test ! -f bad.txt"})
	if err == nil || !strings.Contains(err.Error(), "reset") {
		t.Fatalf("a failed git bisect reset must be an error that names the reset, got %v\n%s", err, out.String())
	}
	t.Chdir(wt)
	cmd := newIntegrationBisectCmd()
	cmd.SetArgs([]string{"--good", base, "--bad", tip, "--", "test", "!", "-f", "bad.txt"})
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	if cmdErr := cmd.Execute(); cmdErr == nil {
		t.Fatal("the command must return an error (non-zero exit) when the reset fails")
	}
}

// F4: a `git bisect start` that fails after moving HEAD must not leave the
// tree on another commit.
func TestBisectRestoresHeadWhenStartFailsPartway(t *testing.T) {
	primary, wt, base, tip := bisectRepo(t)
	bisectPostCheckoutHook(t, primary, fmt.Sprintf(`[ "$2" = "%s" ] && exit 0; exit 1`, tip))
	err := runIntegrationBisect(&strings.Builder{}, integrationBisectOpts{Dir: wt, Bad: tip, Good: base, Command: "true"})
	if err == nil || !strings.Contains(err.Error(), "start failed") {
		t.Fatalf("a failed start must be reported, got %v", err)
	}
	if head := fcGit(t, wt, "rev-parse", "HEAD"); head != tip {
		t.Fatalf("a failed start must leave HEAD at %s, is %s", tip, head)
	}
}
