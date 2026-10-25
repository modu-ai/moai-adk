package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/spf13/cobra"
)

// newIntegrationRemeasureCmd — `moai integration remeasure -- <command>`
// (card t1479, REQ-MWQ-014/015/016): run the re-measure the merge gate
// requires, BEFORE joining the window queue, and store its record keyed by
// the candidate tree. The verb body is factory.RunRemeasure; the verb parses
// the command, renders the verdict, and maps the clean-check refusal to a
// non-zero exit.
func newIntegrationRemeasureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remeasure -- <command>",
		Short: "Run the re-measure and store its record keyed by the candidate tree",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The empty-command guard reads the RAW join: a single empty
			// argument quotes to '' and must still read as "no command"
			// (the verb's own tests pin the message).
			if strings.TrimSpace(strings.Join(args, " ")) == "" {
				return fmt.Errorf("integration remeasure: pass the command after -- (a re-measure without a command measures nothing)")
			}
			command := strings.TrimSpace(shellJoinArgs(args))
			root := integrationLockRoot()
			worktree, wdErr := remeasureWorktreeDir()
			if wdErr != nil {
				return fmt.Errorf("integration remeasure: resolve the working directory: %w", wdErr)
			}
			baseBranch, baseErr := remeasureBaseBranch(root, worktree)
			if baseErr != nil {
				return baseErr
			}
			rec, err := factory.RunRemeasure(root, worktree, baseBranch, command)
			if err != nil {
				return err
			}
			verdict, vErr := remeasureVerdictError(rec)
			// The record and the diagnostics print before the verdict's
			// error returns — the caller keeps the evidence AND the exit
			// code (t1576 review round 2: the INVALID print alone exited 0,
			// and a calling script read the re-measure as passed).
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "re-measure recorded for tree %s (base %s): %s\n", factory.ShortSHA(rec.Tree), factory.ShortSHA(rec.Base), verdict)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  command: %s (exit %d)\n", rec.Command, rec.ExitCode)
			return vErr
		},
	}
	return cmd
}

// remeasureBaseBranch names the integration branch the re-measure records as
// its absorbed base (card t1790, GH #1817). The configured target comes first.
// git-strategy.yaml is gitignored, so a checkout that never had the file (a
// fresh clone, a CI workspace) resolved an empty branch and the verb asked git
// for the fork point of the ref `refs/heads/` (no branch name), which died with
// exit status 128.
//
// Only an ABSENT or unreadable file lets the remote's default branch (origin/HEAD)
// stand in. A file that loads but resolves no target (git-flow outside manual
// mode, an empty develop_branch, an unknown workflow, an empty environment) is a
// misconfiguration: like every sibling consumer the verb refuses it with the
// cause-specific guidance, before the command runs, instead of recording a
// re-measure against a branch the project does not integrate into (sync-audit
// iter1 F1). When nothing resolves, the verb refuses with guidance rather than
// call git with an empty ref.
func remeasureBaseBranch(root, worktree string) (string, error) {
	cfg := config.LoadGitFlowIntegrationConfig(root)
	branch := strings.TrimSpace(cfg.IntegrationTarget)
	if branch == "" {
		if cfg.EmptyTargetCause() != config.EmptyTargetUnreadable {
			guidance := cfg.EmptyTargetGuidance(root, "")
			if guidance == "" {
				guidance = "the configured git strategy resolves no integration branch"
			}
			return "", fmt.Errorf("integration remeasure: no integration target is configured, so the base would be a guess: %s", guidance)
		}
		branch = remoteDefaultBranch(worktree)
	}
	if branch != "" {
		// The record names the base through refs/heads/<branch>, so that local
		// branch must exist before the command runs: a remote default branch
		// with no local counterpart (a clone that checked out another branch)
		// would otherwise fail when the fork point is computed, AFTER the
		// command (card-review).
		verify := exec.Command("git", "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
		verify.Dir = worktree
		verify.Env = gitenv.Env()
		if err := verify.Run(); err != nil {
			return "", fmt.Errorf("integration remeasure: the integration branch %q has no local branch refs/heads/%s; create it (git branch %s origin/%s) or fix git_strategy, then re-run", branch, branch, branch, branch)
		}
		return branch, nil
	}
	return "", fmt.Errorf("integration remeasure: no integration branch to measure against: %s/.moai/config/sections/git-strategy.yaml (gitignored; read from the primary checkout) is absent or unreadable and the remote default branch (refs/remotes/origin/HEAD) does not resolve; restore the git strategy file there, or, when the remote is named origin, run `git remote set-head origin --auto`, then re-run", root)
}

// remoteDefaultBranch reads the branch refs/remotes/origin/HEAD points at, or
// "" when it does not resolve. The FULL ref is read (no --short): a local branch
// or tag named origin/main makes the short form print remotes/origin/main, which
// no prefix strip recovers (sync-audit iter1 F2). A target outside
// refs/remotes/origin/ is not a default branch this verb can name.
func remoteDefaultBranch(worktree string) string {
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD")
	cmd.Dir = worktree
	cmd.Env = gitenv.Env()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	const prefix = "refs/remotes/origin/"
	ref := strings.TrimSpace(string(out))
	if !strings.HasPrefix(ref, prefix) {
		return ""
	}
	return strings.TrimPrefix(ref, prefix)
}

// shellJoinArgs joins the argv the caller passed after -- into the one
// command string the remeasure runs and records. A SINGLE argument is the
// shell line the caller meant — `exit 7`, and the AGENTS.md env-scrub
// compound (`unset VARS && cmd`) arrives the same way — and passes through
// verbatim (t1576 review round 8). When the command arrives as separate
// words, an argument carrying a shell metacharacter (a regex's `|`) is
// single-quoted into one word; a space-only argument stays bare — the
// shell-line reading of separate words.
func shellJoinArgs(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuoteArg(arg)
	}
	// Leading NAME=value arguments are assignments, not words: quoting the
	// whole argument made sh read it as a command name and the test never
	// started (t1576 review round 13). The assignment form survives; only
	// the value is quoted when it needs it.
	for i := 0; i < len(args); i++ {
		name, value, ok := strings.Cut(args[i], "=")
		if !ok || !isShellName(name) {
			break
		}
		quoted[i] = name + "=" + shellQuoteArg(value)
	}
	return strings.Join(quoted, " ")
}

// isShellName reports whether name is a valid shell identifier — the NAME
// half of an assignment argument.
func isShellName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

const shellMetachars = " \t\n" + "|&;<>()$`\\\"'*?[]#~"

func shellQuoteArg(arg string) string {
	if arg == "" {
		return "''"
	}
	if strings.ContainsAny(arg, shellMetachars) {
		return "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}
	return arg
}

// remeasureWorktreeDir resolves the tree the re-measure RUNS in and keys:
// the caller's working directory. A card worktree session's
// CLAUDE_PROJECT_DIR names the PRIMARY checkout — measuring there recorded
// the primary's tree while the card's tree carried no record (t1576 review
// round 3). The project dir stays the record store root only.
func remeasureWorktreeDir() (string, error) {
	return os.Getwd()
}

// remeasureVerdictError renders the record's verdict and carries an INVALID
// verdict as the returned error: the record and the diagnostics still land,
// and the verb's exit code no longer reads as success on an invalid record
// (t1576 review round 2).
func remeasureVerdictError(rec *factory.RemeasureRecord) (string, error) {
	if err := factory.ValidateRemeasureRecord(rec); err != nil {
		return "INVALID: " + err.Error(), fmt.Errorf("integration remeasure: the record for tree %s is not valid: %w", factory.ShortSHA(rec.Tree), err)
	}
	return "valid", nil
}
