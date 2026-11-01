package cli

import (
	"context"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// afCardIDMax is the longest card id `moai af add` accepts.
const afCardIDMax = 128

// afCardIDPattern is what a card id may be made of: it goes into the trunk commit message of a landing (merge(<id>): ...),
// into notes and into the names of log files, so it is a plain name (N71). It starts with a letter or digit, which also keeps
// it from reading as a flag or as ".." in a path.
var afCardIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// afCheckCardID refuses an id that is not a plain name, before anything is opened or written.
func afCheckCardID(id string) error {
	if len(id) > afCardIDMax || !afCardIDPattern.MatchString(id) {
		return fmt.Errorf("af add: the card id %.40q is refused: an id is 1 to %d characters of letters, digits, '.', '_' and '-', starting with a "+
			"letter or digit (it goes into commit messages, notes and file names); nothing was changed", id, afCardIDMax)
	}
	return nil
}

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

The id is a plain name: 1 to 128 characters of letters, digits, '.', '_' and '-', starting with a letter or digit (it goes
into commit messages, notes and file names), and the verb refuses any other before anything is written. Beyond that
nothing here is checked but what the ledger checks: it refuses an id that exists, a class that is not A, B or C, a
path that is absolute or leaves the repository, and a dependency that does not exist or is cancelled; a refusal changes
nothing.

` + afAsGivenHelp + `

` + afBindHelp,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := afCheckCardID(args[0]); err != nil {
				return err
			}
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
