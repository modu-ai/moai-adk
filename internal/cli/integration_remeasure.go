package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factory"
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
			rec, err := factory.RunRemeasure(root, worktree, configuredIntegrationBranch(), command)
			if err != nil {
				return err
			}
			verdict, vErr := remeasureVerdictError(rec)
			// The record and the diagnostics print before the verdict's
			// error returns — the caller keeps the evidence AND the exit
			// code (t1576 review round 2: the INVALID print alone exited 0,
			// and a calling script read the re-measure as passed).
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "re-measure recorded for tree %s (base %s): %s\n", rec.Tree[:12], rec.Base[:12], verdict)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  command: %s (exit %d)\n", rec.Command, rec.ExitCode)
			return vErr
		},
	}
	return cmd
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
		return "INVALID: " + err.Error(), fmt.Errorf("integration remeasure: the record for tree %s is not valid: %w", rec.Tree[:12], err)
	}
	return "valid", nil
}
