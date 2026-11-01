package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// newAFAddCmd is `moai af add`: the verb that puts a card into the ledger (N59). It hands the ledger's AddCard a READY
// card and adds no rule of its own: the ledger refuses a duplicate id, a class that is not A, B or C, a path that leaves
// the repository and a dependency that does not exist.
func newAFAddCmd() *cobra.Command {
	var f afFlags
	var title, class string
	var priority int
	var paths, depends, risk, acceptance []string
	var stop string
	cmd := &cobra.Command{
		Use:   "add <card>",
		Short: "Add a card to the ledger in READY",
		Long: `Add a card to the ledger. The card starts in READY, where "moai af claim" finds it.

Only the id is required. --title defaults to the id and --class to C (the class of a card whose class nobody chose);
--priority defaults to 0. --path (repeatable, repository-relative) names the files or directories the card edits: two
cards that share a path cannot be claimed together. --depends (repeatable) names cards that must be DONE before this one
is claimed and that must exist already. --risk (repeatable: security, permission, data, ...) marks a card that needs a
review run before it is taken to the merge queue. --acceptance (repeatable) lists the test names or commands that say
the card is done, and --stop says when to stop; both are recorded as given.

Nothing here is checked beyond what the ledger checks: it refuses an id that exists, a class that is not A, B or C, a
path that is absolute or leaves the repository, and a dependency that does not exist or is cancelled; a refusal changes
nothing.

` + afBindHelp,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				title = args[0]
			}
			card := ledger.Card{
				ID: args[0], Title: title, Position: core.Position{Stage: core.StageReady}, Priority: priority,
				Spec: ledger.Spec{Class: class, Paths: paths, Depends: depends, Risk: risk, Acceptance: acceptance, Stop: stop},
			}
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				if err := st.AddCard(ctx, card); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": card.ID, "title": card.Title, "stage": card.Position.Stage, "class": class})
			})
		},
	}
	f.bind(cmd, false)
	cmd.Flags().StringVar(&title, "title", "", "the card's title (default: the id; recorded as given)")
	cmd.Flags().StringVar(&class, "class", "C", "A, B or C")
	cmd.Flags().IntVar(&priority, "priority", 0, "the card's priority")
	cmd.Flags().StringSliceVar(&paths, "path", nil, "a repository-relative path the card edits (repeatable)")
	cmd.Flags().StringSliceVar(&depends, "depends", nil, "a card that must be DONE before this one is claimed (repeatable)")
	cmd.Flags().StringSliceVar(&risk, "risk", nil, "a risk of the card, such as security, permission or data (repeatable)")
	cmd.Flags().StringArrayVar(&acceptance, "acceptance", nil, "a test name or command that says the card is done (repeatable, recorded as given)")
	cmd.Flags().StringVar(&stop, "stop", "", "when to stop (recorded as given)")
	return cmd
}
