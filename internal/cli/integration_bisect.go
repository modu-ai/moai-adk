package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"slices"
	"strings"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/spf13/cobra"
)

// bisectFirstBadRe reads the line git writes into its own bisect log when the
// search converges. The log is git's state, not a stream the judged command
// writes to, so a command that prints "<sha> is the first bad commit" cannot
// move the verdict. A run that aborted or never converged leaves no such line,
// so the verb reports a failure instead of naming whatever refs/bisect/bad holds.
var bisectFirstBadRe = regexp.MustCompile(`(?m)^# first bad commit: \[([0-9a-f]{40})\]`)

// bisectRepoVars are the repository-scoped variables the pre-checks already
// scrub (internal/auditreceipt scrubbedGitVars). git bisect start/run/reset
// drop the same set, so the checks and the bisect cannot look at two trees.
var bisectRepoVars = []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE", "GIT_CEILING_DIRECTORIES"}

// bisectCardIDRe reads the card id out of a landing subject written by the
// merge step ("Merge <branch> into <target> (card <id>, integration merge)").
var bisectCardIDRe = regexp.MustCompile(`\(card (t[0-9]+)`)

// integrationBisectOpts carries one bisect invocation. Dir is the tree the
// search runs in; the verb never resolves it itself so tests can aim it.
type integrationBisectOpts struct {
	Dir, Bad, Good, Command string
	DryRun                  bool
}

// newIntegrationBisectCmd — `moai integration bisect --good <ref> -- <command>`
// (card t1833): local reproduction of a failing batch at card granularity.
// Landings are --no-ff merges that carry the card id, so a first-parent walk
// visits one commit per card and the first bad one names the culprit card.
// The verb is a thin guard around `git bisect --first-parent`: it refuses the
// primary checkout (bisect moves HEAD), a dirty tree and a bisect already in
// progress, and it resets on every path out, SIGINT and SIGTERM included (the
// judged run is stopped, then the tree is restored), failing the command when
// it cannot confirm that HEAD and the bisect state are back where they started.
func newIntegrationBisectCmd() *cobra.Command {
	var badFlag, goodFlag string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "bisect --good <ref> [--bad <ref>] [--dry-run] -- <command>",
		Short: "Find the card whose landing breaks <command> (git bisect --first-parent, card tree only)",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(strings.Join(args, " ")) == "" {
				return fmt.Errorf("integration bisect: pass the test command after -- (a bisect without a command judges nothing)")
			}
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("integration bisect: resolve the working directory: %w", err)
			}
			return runIntegrationBisect(cmd.OutOrStdout(), integrationBisectOpts{
				Dir: dir, Bad: badFlag, Good: goodFlag, Command: strings.TrimSpace(shellJoinArgs(args)), DryRun: dryRun,
			})
		},
	}
	cmd.Flags().StringVar(&badFlag, "bad", "HEAD", "A ref known to fail the command (the failing batch tip)")
	cmd.Flags().StringVar(&goodFlag, "good", "", "A ref known to pass the command (the last green landing)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the exact git commands and run nothing that moves HEAD")
	return cmd
}

// bisectGit runs git in dir. The repository-scoped GIT_* variables are removed
// (the pre-checks run without them too) and the rest of the caller's
// environment is kept: the judged command inherits it, so it must be the real
// one, not a scrubbed copy.
func bisectGit(dir string, out io.Writer, args ...string) error {
	return bisectGitCtx(context.Background(), dir, out, nil, args...)
}

// bisectGitCtx is bisectGit for a command that a context may stop; prepare, when
// set, adjusts the command before it starts.
func bisectGitCtx(ctx context.Context, dir string, out io.Writer, prepare func(*exec.Cmd), args ...string) error {
	c := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	if prepare != nil {
		prepare(c)
	}
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if !slices.Contains(bisectRepoVars, name) {
			c.Env = append(c.Env, kv)
		}
	}
	c.Stdout, c.Stderr = out, out
	return c.Run()
}

// bisectInProgress reports whether dir has a bisect in progress (BISECT_LOG).
func bisectInProgress(dir string) bool {
	logPath, _ := runScrubbedGit(dir, "rev-parse", "--path-format=absolute", "--git-path", "BISECT_LOG")
	if strings.TrimSpace(logPath) == "" {
		return false
	}
	_, statErr := os.Stat(strings.TrimSpace(logPath))
	return statErr == nil
}

// bisectRestore ends any bisect in dir and puts HEAD back on the branch or
// commit it started on, then checks that it worked. A nil result means the tree
// is exactly as the verb found it.
func bisectRestore(dir, origSHA, origBranch string) error {
	var problems []string
	if bisectInProgress(dir) {
		if err := bisectGit(dir, io.Discard, "bisect", "reset"); err != nil {
			problems = append(problems, fmt.Sprintf("git bisect reset failed (%v)", err))
		}
	}
	if head, _ := runScrubbedGit(dir, "rev-parse", "HEAD"); strings.TrimSpace(head) != origSHA {
		target := []string{"checkout", "--quiet", "--detach", origSHA}
		if origBranch != "" {
			target = []string{"checkout", "--quiet", origBranch}
		}
		if err := bisectGit(dir, io.Discard, target...); err != nil {
			problems = append(problems, fmt.Sprintf("restoring HEAD failed (%v)", err))
		}
	}
	head, _ := runScrubbedGit(dir, "rev-parse", "HEAD")
	branch, _ := runScrubbedGit(dir, "symbolic-ref", "--quiet", "--short", "HEAD")
	if bisectInProgress(dir) {
		problems = append(problems, "a bisect is still in progress")
	}
	if strings.TrimSpace(head) != origSHA || strings.TrimSpace(branch) != origBranch {
		problems = append(problems, fmt.Sprintf("HEAD is %s%s, not %s%s", factory.ShortSHA(strings.TrimSpace(head)), bisectBranchNote(strings.TrimSpace(branch)), factory.ShortSHA(origSHA), bisectBranchNote(origBranch)))
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("integration bisect: %s in %s; run git bisect reset by hand (and check out the starting commit %s)", strings.Join(problems, "; "), dir, factory.ShortSHA(origSHA))
}

// bisectIgnoredCollisions lists ignored local files whose path a commit the
// search can visit tracks. `git checkout` replaces such a file without a word
// (the dirty-tree check does not see ignored files), so the verb refuses first.
// The visited trees are the good tree plus every path a first-parent step adds.
func bisectIgnoredCollisions(top, good, bad string) ([]string, error) {
	ignored, err := runScrubbedGit(top, "ls-files", "-z", "--others", "--ignored", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("list ignored files: %w", err)
	}
	if ignored == "" {
		return nil, nil
	}
	tracked := map[string]bool{}
	for _, args := range [][]string{
		{"ls-tree", "-r", "-z", "--name-only", good},
		{"log", "--first-parent", "--diff-filter=A", "--name-only", "-z", "--format=", good + ".." + bad},
	} {
		out, listErr := runScrubbedGit(top, args...)
		if listErr != nil {
			return nil, fmt.Errorf("list tracked paths (git %s): %w", args[0], listErr)
		}
		for _, p := range strings.Split(out, "\x00") {
			tracked[strings.TrimLeft(p, "\n")] = true
		}
	}
	var hits []string
	for _, p := range strings.Split(ignored, "\x00") {
		if p != "" && tracked[p] {
			hits = append(hits, p)
		}
	}
	return hits, nil
}

func bisectBranchNote(branch string) string {
	if branch == "" {
		return " (detached)"
	}
	return " (" + branch + ")"
}

func runIntegrationBisect(out io.Writer, o integrationBisectOpts) error {
	if strings.TrimSpace(o.Good) == "" {
		return fmt.Errorf("integration bisect: --good <ref> is required (the last landing known to pass)")
	}
	top, err := runScrubbedGit(o.Dir, "rev-parse", "--path-format=absolute", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("integration bisect: %s is not inside a git work tree: %w", o.Dir, err)
	}
	top = strings.TrimSpace(top)
	primary, _, err := identifyPrimaryCheckout(top)
	if err != nil {
		return fmt.Errorf("integration bisect: cannot identify the primary checkout of %s: %w", top, err)
	}
	if factorySameTree(top, primary) {
		return fmt.Errorf("integration bisect: refused — %s is the primary checkout and bisect moves HEAD. Run it from a card worktree (moai worktree new <name>)", top)
	}
	if st, _ := runScrubbedGit(top, "status", "--porcelain"); strings.TrimSpace(st) != "" {
		return fmt.Errorf("integration bisect: refused — %s has uncommitted changes and bisect checks out other commits", top)
	}
	if bisectInProgress(top) {
		return fmt.Errorf("integration bisect: refused — a bisect is already in progress in %s (git bisect reset ends it)", top)
	}
	shas := map[string]string{}
	for _, ref := range []string{o.Bad, o.Good} {
		sha, revErr := runScrubbedGit(top, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
		if revErr != nil {
			return fmt.Errorf("integration bisect: %q does not name a commit", ref)
		}
		shas[ref] = strings.TrimSpace(sha)
	}
	hits, collideErr := bisectIgnoredCollisions(top, shas[o.Good], shas[o.Bad])
	if collideErr != nil {
		return fmt.Errorf("integration bisect: refused — cannot rule out ignored local files that the search would overwrite (%w)", collideErr)
	}
	if len(hits) > 0 {
		return fmt.Errorf("integration bisect: refused — ignored local file(s) %s in %s share a path with a commit between %s and %s; a checkout during the bisect would overwrite them without warning. Move them away, then run again", strings.Join(hits, ", "), top, factory.ShortSHA(shas[o.Good]), factory.ShortSHA(shas[o.Bad]))
	}
	start := []string{"bisect", "start", "--first-parent", o.Bad, o.Good}
	run := []string{"bisect", "run", "sh", "-c", o.Command}
	reset := []string{"bisect", "reset"}
	if o.DryRun {
		_, _ = fmt.Fprintf(out, "integration bisect (dry-run): bad %s, good %s, tree %s\n", factory.ShortSHA(shas[o.Bad]), factory.ShortSHA(shas[o.Good]), top)
		for _, args := range [][]string{start, run, reset} {
			// The judged command is one word to git; quote it so the printed line
			// pastes into a shell as the same three words.
			shown := slices.Clone(args)
			if len(shown) == 5 && shown[1] == "run" {
				shown[4] = shellQuoteArg(shown[4])
			}
			_, _ = fmt.Fprintf(out, "  git %s\n", strings.Join(shown, " "))
		}
		return nil
	}
	origSHA, headErr := runScrubbedGit(top, "rev-parse", "HEAD")
	if headErr != nil {
		return fmt.Errorf("integration bisect: refused — cannot read HEAD of %s (%w)", top, headErr)
	}
	origSHA = strings.TrimSpace(origSHA)
	origBranch, _ := runScrubbedGit(top, "symbolic-ref", "--quiet", "--short", "HEAD")
	origBranch = strings.TrimSpace(origBranch)

	// An interrupt stops the judged run (below) and still restores the tree: the
	// notification is registered before anything moves HEAD and released on return.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var startLog bytes.Buffer
	if err := bisectGit(top, &startLog, start...); err != nil {
		_, _ = fmt.Fprint(out, startLog.String())
		startErr := fmt.Errorf("integration bisect: git bisect start failed: %w", err)
		if restoreErr := bisectRestore(top, origSHA, origBranch); restoreErr != nil {
			return fmt.Errorf("%w; %w", startErr, restoreErr)
		}
		return startErr
	}
	runErr := bisectGitCtx(ctx, top, out, bisectRunPrepare, run...)
	interrupted := runErr != nil && ctx.Err() != nil
	// The culprit is read from git's own bisect log before the reset removes it.
	var culprit string
	if runErr == nil {
		blog, _ := runScrubbedGit(top, "bisect", "log")
		if m := bisectFirstBadRe.FindStringSubmatch(blog); m != nil {
			culprit = m[1]
		}
	}
	// Restoration runs on the uncancelled context: bisectRestore starts its git
	// processes without ctx, and the registered notification keeps swallowing
	// further signals until the function returns.
	restoreErr := bisectRestore(top, origSHA, origBranch)
	if interrupted {
		state := "the tree is restored (HEAD and the bisect state are back where they started)"
		if restoreErr != nil {
			state = "restoration was NOT confirmed: " + restoreErr.Error()
		}
		return fmt.Errorf("integration bisect: interrupted by a signal before the search converged, no culprit is named; %s", state)
	}
	if culprit == "" {
		notConverged := fmt.Errorf("integration bisect: the search did not converge (git bisect run aborts on exit codes 126, 127 and 128 or above, and stops short when every candidate is skipped with 125 or the endpoints do not bracket a change)")
		if restoreErr != nil {
			return fmt.Errorf("%w; %w", notConverged, restoreErr)
		}
		return notConverged
	}
	subject, _ := runScrubbedGit(top, "show", "-s", "--format=%s", culprit)
	subject = strings.TrimSpace(subject)
	card := "none in the subject (a direct commit, not a card landing)"
	if cm := bisectCardIDRe.FindStringSubmatch(subject); cm != nil {
		card = cm[1]
	}
	_, _ = fmt.Fprintf(out, "integration bisect: first bad first-parent commit %s\n  subject: %s\n  card: %s\n", factory.ShortSHA(culprit), subject, card)
	return restoreErr
}
