package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/spf13/cobra"
)

// bisectFirstBadRe reads the line `git bisect run` prints when the search
// converges. A run that aborted or never converged prints no such line, so
// the verb reports a failure instead of naming whatever refs/bisect/bad holds.
var bisectFirstBadRe = regexp.MustCompile(`(?m)^([0-9a-f]{40}) is the first bad commit`)

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
// progress, and it always resets.
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

// bisectGit runs git in dir under the caller's environment: the judged
// command inherits it, so it must be the real one, not a scrubbed copy.
func bisectGit(dir string, out io.Writer, args ...string) error {
	c := exec.Command("git", append([]string{"-C", dir}, args...)...)
	c.Stdout, c.Stderr = out, out
	return c.Run()
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
	if logPath, _ := runScrubbedGit(top, "rev-parse", "--path-format=absolute", "--git-path", "BISECT_LOG"); logPath != "" {
		if _, statErr := os.Stat(strings.TrimSpace(logPath)); statErr == nil {
			return fmt.Errorf("integration bisect: refused — a bisect is already in progress in %s (git bisect reset ends it)", top)
		}
	}
	shas := map[string]string{}
	for _, ref := range []string{o.Bad, o.Good} {
		sha, revErr := runScrubbedGit(top, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
		if revErr != nil {
			return fmt.Errorf("integration bisect: %q does not name a commit", ref)
		}
		shas[ref] = strings.TrimSpace(sha)
	}
	start := []string{"bisect", "start", "--first-parent", o.Bad, o.Good}
	run := []string{"bisect", "run", "sh", "-c", o.Command}
	reset := []string{"bisect", "reset"}
	if o.DryRun {
		_, _ = fmt.Fprintf(out, "integration bisect (dry-run): bad %s, good %s, tree %s\n", factory.ShortSHA(shas[o.Bad]), factory.ShortSHA(shas[o.Good]), top)
		for _, args := range [][]string{start, run, reset} {
			_, _ = fmt.Fprintf(out, "  git %s\n", strings.Join(args, " "))
		}
		return nil
	}
	var log bytes.Buffer
	if err := bisectGit(top, &log, start...); err != nil {
		_, _ = fmt.Fprint(out, log.String())
		return fmt.Errorf("integration bisect: git bisect start failed: %w", err)
	}
	defer func() {
		if resetErr := bisectGit(top, io.Discard, reset...); resetErr != nil {
			_, _ = fmt.Fprintf(out, "integration bisect: WARNING git bisect reset failed (%v); run it by hand in %s\n", resetErr, top)
		}
	}()
	log.Reset()
	runErr := bisectGit(top, io.MultiWriter(&log, out), run...)
	m := bisectFirstBadRe.FindStringSubmatch(log.String())
	if runErr != nil || m == nil {
		return fmt.Errorf("integration bisect: the search did not converge (the command exited >127, or the endpoints do not bracket a change)")
	}
	subject, _ := runScrubbedGit(top, "show", "-s", "--format=%s", m[1])
	subject = strings.TrimSpace(subject)
	card := "none in the subject (a direct commit, not a card landing)"
	if cm := bisectCardIDRe.FindStringSubmatch(subject); cm != nil {
		card = cm[1]
	}
	_, _ = fmt.Fprintf(out, "integration bisect: first bad first-parent commit %s\n  subject: %s\n  card: %s\n", factory.ShortSHA(m[1]), subject, card)
	return nil
}
