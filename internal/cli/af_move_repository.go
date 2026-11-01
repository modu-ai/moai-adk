package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// newAFMoveRepositoryCmd is `moai af move-repository` (N52): the operator's way to tell the ledger that the repository it
// is bound to moved on disk. It is the one verb that does not go through afRun: afRun binds the ledger to the project's
// repository or compares it with the bound one, and after a move those two differ by design, so every other verb is
// refused until this one has run. It opens the ledger itself (honouring --db), reads the current identity, checks that
// the move is one, and calls the ledger's MoveRepository. The ledger runs no git (REQ-AFL-008); the checks are the
// verb's.
func newAFMoveRepositoryCmd() *cobra.Command {
	var f afFlags
	var reason string
	cmd := &cobra.Command{
		Use:   "move-repository",
		Short: "Record that the repository this ledger is bound to moved on disk",
		Long: `Record that the repository this ledger is bound to moved on disk, so the other verbs compare with the new place.

The ledger is bound to the git directory of one repository. When that repository is moved or renamed, every other verb is
refused (it is bound to a git directory that is not there), and this verb is the way out. It records the move only when
it is one, and checks, before anything is written:

  - the old repository no longer exists on disk: the git directory the ledger is bound to is gone. If it is still there,
    this is a second repository and not a move, and the verb refuses and says so;
  - the new repository holds every commit the rows of the ledger name (the subject of every evidence row and the
    candidate of every queue row, as the first bind checks them), and the verb refuses and names the first it lacks;
  - --reason is not blank. It is recorded with the move and listed by the ledger.

The new repository is the one --bind-repo names (any path inside it, a linked worktree included), else the project
directory, which CLAUDE_PROJECT_DIR or else the working directory names. Its identity is computed as every other verb
computes it (its git directory, symbolic links resolved). A refusal leaves the ledger unchanged, and so does a ledger
that is not bound yet: it has no repository to move, and this verb does not bind it. This verb does not bind the ledger
or compare it with the project's repository; it only reads the repository the ledger is bound to now. Afterwards that
is the new one: a call that would use the old one is refused, and one from the new place passes.

A lease records the worktree its holder claimed with, and that path is not rewritten: evidence and enqueue for a card
claimed before the move need --repo <a path in the new repository>.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return afMoveRepository(cmd, f, reason)
		},
	}
	cmd.Flags().StringVar(&f.db, "db", "", "ledger file (default: <project dir>/af/ledger.db)")
	cmd.Flags().StringVar(&f.bindRepo, "bind-repo", "", "the repository the ledger belongs to now: any path inside it (a linked worktree included); "+
		"default: the project directory. It must hold every commit the rows name")
	cmd.Flags().StringVar(&reason, "reason", "", "why the repository moved (required, recorded with the move)")
	return cmd
}

// afMoveRepository is the body of the verb: three checks, then one ledger call. Every refusal comes before the call, so
// it changes nothing.
func afMoveRepository(cmd *cobra.Command, f afFlags, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("af: refused: --reason is required and must not be blank; nothing was changed")
	}
	path, err := afLedgerPath(f.db)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("af: refused: there is no ledger at %s (%v); nothing was changed", path, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), afOpTimeout)
	defer cancel()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		return fmt.Errorf("af: open the ledger %s: %w", path, err)
	}
	defer func() { _ = st.Close() }()

	current, err := st.Repository(ctx)
	if err != nil {
		return err
	}
	if current == "" {
		return errors.New("af: refused: this ledger is not bound to a repository yet, so there is no repository to move; " +
			"the first verb that writes it binds it; nothing was changed")
	}
	// (a) the old repository is gone: only a path that is positively not there counts, an error git or the file system
	// gives for another reason does not
	if _, err := os.Lstat(current); err == nil {
		return fmt.Errorf("af: refused: the repository this ledger is bound to still exists on disk (%s), so it did not move; "+
			"nothing was changed", current)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("af: refused: cannot tell whether the repository this ledger is bound to (%s) is gone: %v; nothing was changed", current, err)
	}
	// (b) the new repository is a repository and holds the commits the rows name
	dir := f.bindRepo
	if dir == "" {
		dir = resolveProjectDir()
	}
	next, err := afGitCommonPath(ctx, dir)
	if err != nil {
		if errors.Is(err, errAFNotARepository) {
			return fmt.Errorf("af: refused: %q (--bind-repo, else the project directory) is not a directory inside a git repository; nothing was changed", dir)
		}
		return fmt.Errorf("af: refused: cannot tell which repository %q belongs to (%w); nothing was changed", dir, err)
	}
	shas, err := afRowCommits(ctx, st)
	if err != nil {
		return err
	}
	if missing := afMissingCommit(ctx, dir, shas); missing != "" {
		return fmt.Errorf("af: refused: the rows of this ledger name the commit %s, which the repository %s (the git directory of %s) does not have; "+
			"the commits of a ledger live in its repository, so this is not where it moved to; nothing was changed", missing, next, dir)
	}
	if err := st.MoveRepository(ctx, current, next, reason); err != nil {
		return fmt.Errorf("af: refused: %w; nothing was changed", err)
	}
	return afPrint(cmd, map[string]any{"from": current, "to": next, "reason": reason, "commits_checked": len(shas)})
}
