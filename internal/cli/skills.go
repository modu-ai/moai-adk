package cli

// skills.go — the `moai skills` command tree
// (SPEC-CODEX-SKILL-DISABLE-001).
//
// The root name is deliberately broad: skill exposure is a per-LAYER concern,
// and the layer is always named by a flag rather than implied by the verb. So
// `moai skills disable X --codex` leaves room for a future `--claude` under
// the same roof. This card implements the Codex layer only.
//
// Making --codex REQUIRED is the mechanical form of the card's opt-in
// property: the moment moai writes into the user's HOME, the user has named
// that layer in the invocation. No project config key drives this verb, and
// none may — a project setting asking for a write into a user's home is a
// write the user did not request at that moment.

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
)

// skillsProjectRootFn resolves the project whose skill mirror is consulted.
// It is a seam so a test can name a fixture tree without moving the process.
var skillsProjectRootFn = defaultSkillsProjectRoot

// defaultSkillsProjectRoot is the working directory, deliberately — NOT the
// nearest enclosing MoAI project.
//
// Codex resolves `.agents/skills` against its OWN working directory, so the
// cwd is precisely the root a Codex started here would consult. Anchoring
// anywhere else would publish an entry for a root Codex never reads.
//
// Walking up for a project marker was tried and abandoned: from a fixture
// under /tmp the walk found an unrelated `.moai` at /private/tmp and reported
// that as the project, which is the same class of surprise on a user's
// machine. When the cwd carries no mirror the verb says so and names the exact
// path it looked at, which is also the answer for Codex started in that
// directory.
func defaultSkillsProjectRoot() (string, error) {
	return os.Getwd()
}

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skills",
		Short:   "Manage which skills each layer sees",
		GroupID: "tools",
	}
	cmd.AddCommand(newSkillsDisableCmd())
	return cmd
}

func newSkillsDisableCmd() *cobra.Command {
	var codex, force bool

	cmd := &cobra.Command{
		Use:   "disable <skill-name>",
		Short: "Hide one skill from Codex, leaving Claude Code untouched",
		Long: `Hide one skill from Codex by writing a single [[skills.config]] entry with
enabled = false into the user-layer config (~/.codex/config.toml, or
$CODEX_HOME/config.toml). Claude Code keeps seeing the skill.

The entry names the project's skill mirror file,
<projectRoot>/.agents/skills/<skill>/SKILL.md. Default is dry-run: pass
--force to actually write, and the config is backed up first (its sha256 is
reported) and keeps its own permission mode.

Three things worth knowing, because Codex reports none of them:

  1. Start Codex from the project root. The .agents/skills root is resolved
     against Codex's working directory, so starting it from a subdirectory
     removes the root entirely — the skill and the entry gating it both stop
     applying, silently.
  2. Run this verb again after moving the project. The entry carries an
     absolute path, so a moved project leaves it pointing at nothing and
     Codex says nothing about it. Collect the leftovers with
     'moai clean --codex-skills'.
  3. Still seeing the skill? The gate binds one FILE. Another root — a plugin
     root, for instance — can supply a same-named skill as a DIFFERENT file,
     which this entry does not cover. Remove it at that root.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
			root, err := skillsProjectRootFn()
			if err != nil {
				return fmt.Errorf("resolve project root: %w", err)
			}
			home, _ := userHomeDir() // an unresolvable home just drops that candidate root
			if !codex {
				// Unreachable while --codex is required; kept so the layer
				// gate does not depend on cobra's flag bookkeeping alone.
				return fmt.Errorf("name the layer to write: pass --codex")
			}
			return runCodexSkillDisable(p, codexSkillDisableOptions{
				Skill:       args[0],
				ProjectRoot: root,
				HomeDir:     home,
				Force:       force,
			})
		},
	}

	cmd.Flags().BoolVar(&codex, "codex", false, "Write to the Codex user-layer config (required — the write layer is always named)")
	cmd.Flags().BoolVar(&force, "force", false, "Actually write (default: dry-run)")
	_ = cmd.MarkFlagRequired("codex")

	return cmd
}
