package cli

// contract_pushcheck.go — `moai contract push-check [<remote> <refspec>...]`,
// the manual surface of the push readiness evaluation
// (SPEC-AUTONOMY-CLOSURE-001 REQ-CLOSURE-019).
//
// Exit codes: 0 every in-push contract ready (or the check inactive), 1 any
// contract not ready or the push undetermined, 2 usage error only.

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/config"
)

// pushCheckEvalFn is the evaluation seam the command runs; tests replace it.
var pushCheckEvalFn = closure.EvaluatePush

// runContractPushCheck performs the hook's evaluation for one push.
func runContractPushCheck(cmd *cobra.Command, args []string) error {
	env, err := loadContractEnv(cmd)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()

	// Mode gate: under guided (or an absent/unrecognized mode) the check is
	// inactive and exits 0 (REQ-CLOSURE-023's command-side mirror).
	if env.autonomy.Mode != "contract" {
		_, _ = fmt.Fprintf(out, "inactive (mode %s)\n", env.autonomy.Mode)
		return nil
	}
	integration := config.LoadGitFlowDevelopBranch(env.root)
	if integration == "" {
		_, _ = fmt.Fprintln(out, "inactive (no git-flow integration branch configured)")
		return nil
	}

	push, perr := parsePushArgs(env.root, integration, args)
	if perr != nil {
		return contractUsageError(cmd, "%v", perr)
	}
	if push.undetermined {
		_, _ = fmt.Fprintf(out, "push_check_undetermined: %s\n", push.cause)
		return &exitCodeError{code: contractExitInvalid, msg: "push_check_undetermined"}
	}

	result := pushCheckEvalFn(closure.GitSeam{}, closure.PushEvalInput{
		Tree:                env.root,
		Remote:              push.remote,
		Integration:         integration,
		Source:              push.source,
		SecondReview:        env.autonomy.SecondReview,
		RegistryRuleIDs:     env.ruleIDs,
		RegistryFrozenFiles: env.frozenFiles,
	})
	if result.Undetermined {
		_, _ = fmt.Fprintf(out, "push_check_undetermined: %s\n", result.Cause)
		return &exitCodeError{code: contractExitInvalid, msg: "push_check_undetermined"}
	}
	for _, cr := range result.Results {
		_, _ = fmt.Fprintf(out, "%s: %s\n", cr.SpecID, strings.Join(cr.Codes, ", "))
	}
	if len(result.Results) > 0 {
		return &exitCodeError{code: contractExitInvalid, msg: "push not ready"}
	}
	_, _ = fmt.Fprintln(out, "ready")
	return nil
}

// parsedPush is the classified push the command evaluates.
type parsedPush struct {
	remote       string
	source       string
	undetermined bool
	cause        string
}

// parsePushArgs classifies the command line per design.md §C.1: without
// arguments, a push of the local integration branch to its upstream;
// `--all`/`--mirror` are undetermined; an unknown flag is a usage error.
func parsePushArgs(root, integration string, args []string) (parsedPush, error) {
	p := parsedPush{remote: "origin", source: integration}
	var positional []string
	for _, a := range args {
		switch {
		case a == "--all" || a == "--mirror":
			return parsedPush{undetermined: true,
				cause: "--all/--mirror may include " + integration + "; it cannot be proven otherwise"}, nil
		case strings.HasPrefix(a, "-"):
			return parsedPush{}, fmt.Errorf("unknown flag %s", a)
		default:
			positional = append(positional, a)
		}
	}
	switch len(positional) {
	case 0:
		// push of the local integration branch to its upstream: the source is
		// the local integration branch head.
	case 1:
		p.remote = positional[0]
	case 2:
		p.remote = positional[0]
		refspec := positional[1]
		refspec = strings.TrimPrefix(refspec, "+")
		src := refspec
		if i := strings.Index(refspec, ":"); i >= 0 {
			src = refspec[:i]
			dst := refspec[i+1:]
			dst = strings.TrimPrefix(dst, "refs/heads/")
			if dst != integration {
				// A non-integration destination is not evaluated: ready.
				return parsedPush{remote: p.remote, source: ""}, nil
			}
		}
		p.source = src
	default:
		return parsedPush{}, fmt.Errorf("takes at most <remote> <refspec> (%d positionals)", len(positional))
	}
	if p.source == "" {
		return parsedPush{}, nil
	}
	// Resolve the source to a commit in the tree.
	seam := closure.GitSeam{}
	if _, err := seam.ResolveRef(root, p.source); err != nil {
		return parsedPush{undetermined: true, cause: "source " + p.source + " does not resolve: " + err.Error()}, nil
	}
	return p, nil
}

// newContractPushCheckCmd builds the `push-check` subcommand.
func newContractPushCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "push-check [<remote> <refspec>]",
		Short:         "Evaluate push readiness for the in-range contracts (exit 0 ready, 1 not ready or undetermined, 2 usage)",
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return runContractPushCheck(c, args)
		},
	}
}
