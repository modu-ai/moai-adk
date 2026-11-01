package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// newAFMoveRepositoryCmd is `moai af move-repository` (N52): the operator's way to tell the ledger that the repository it
// is bound to moved on disk. It is the one verb that does not go through afRun: afRun binds the ledger to the project's
// repository or compares it with the bound one, and after a move those two differ by design, so every other verb is
// refused until this one has run. It opens the ledger itself (honouring --db), reads the current identity, checks that
// the move is one, and calls the ledger's MoveRepositoryMarked. The ledger runs no git (REQ-AFL-008); the checks are the
// verb's.
func newAFMoveRepositoryCmd() *cobra.Command {
	var f afFlags
	var reason, from string
	cmd := &cobra.Command{
		Use:   "move-repository",
		Short: "Record that the repository this ledger is bound to moved on disk",
		Long: `Record that the repository this ledger is bound to moved on disk, so the other verbs compare with the new place.

The ledger is bound to the git directory of one repository, and records the identity marker the repository keeps in that
git directory (a random token; a rename or a move of the directory carries it, a clone, a mirror and a repository made
afterwards do not have it). When the repository is moved or renamed, every other verb is refused (it is bound to a git
directory that is not there), and this verb is the way out. It records the move only when it is one, and checks, before
anything is written:

  - the old repository no longer exists on disk: the git directory the ledger is bound to is gone. If it is still there,
    this is a second repository and not a move, and the verb refuses and says so;
  - the new repository is the same repository: it carries the identity marker the ledger recorded. A path that holds a
    clone, a mirror or another repository is refused, whatever commits it has;
  - --reason is not blank. It is recorded with the move and listed by the ledger.

A ledger bound before markers existed and not used since has no marker to compare. For such a ledger the verb falls back
to the commits: the new repository holds every commit the queue rows that are not failed name (the commits of evidence and
of failed rows may have been thrown away; they are counted as commits_unchecked), git must be able to say so, --bind-repo
must declare the new repository (the project directory is not taken for a declaration), and a ledger file given with --db
must not sit in another repository than the one named. The new repository's marker is then recorded. This fallback
cannot tell a clone from the repository when no queue row names a commit; the declaration is the operator's word.

The new repository is the one --bind-repo names (any path inside it, a linked worktree included), else the project
directory, which CLAUDE_PROJECT_DIR or else the working directory names. Its identity is computed as every other verb
computes it (its git directory, symbolic links resolved). A refusal leaves the ledger unchanged, and so does a ledger
that is not bound yet: it has no repository to move, and this verb does not bind it. This verb does not bind the ledger
or compare it with the project's repository; it only reads the repository the ledger is bound to now. Afterwards that
is the new one: a call that would use the old one is refused, and one from the new place passes. The first recorded
move raises the schema number of the ledger file, so a build that does not know the move refuses the file.

The default ledger location is keyed by the project directory, so a project directory that moved has a new key and no
ledger there. --from <the old project directory> names the old key: its ledger file (with its -wal and -shm) is moved to
the new key before the move is recorded, only when the new key has no ledger yet, and left where it was on any refusal.
Spell the old directory as it was when the ledger was made. --from cannot be combined with --db.

A lease records the worktree its holder claimed with, and that path is not rewritten: evidence and enqueue for a card
claimed before the move need --repo <a path in the new repository>.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return afMoveRepository(cmd, f, reason, from)
		},
	}
	cmd.Flags().StringVar(&f.db, "db", "", "ledger file (default: <project dir>/af/ledger.db)")
	cmd.Flags().StringVar(&f.bindRepo, "bind-repo", "", "the repository the ledger belongs to now: any path inside it (a linked worktree included); "+
		"default: the project directory. A ledger without an identity marker needs it, and the repository must hold the commits of its queue rows")
	cmd.Flags().StringVar(&reason, "reason", "", "why the repository moved (required, recorded with the move)")
	cmd.Flags().StringVar(&from, "from", "", "the project directory the default ledger was made for, before it moved: its ledger is moved to this "+
		"project's key first (not with --db)")
	return cmd
}

// afMoveDir is where the new repository is looked for: --bind-repo, else the project directory.
func afMoveDir(f afFlags) string {
	if f.bindRepo != "" {
		return f.bindRepo
	}
	return resolveProjectDir()
}

// afMoveRepository is the body of the verb: the checks, then (with --from) the ledger file's own move, then one ledger
// call. Every refusal comes before the first change, so it changes nothing.
func afMoveRepository(cmd *cobra.Command, f afFlags, reason, from string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("af: refused: --reason is required and must not be blank; nothing was changed")
	}
	if from != "" && f.db != "" {
		return errors.New("af: refused: --from names the old project directory of the default ledger location and cannot be combined with --db, which names the ledger file itself; nothing was changed")
	}
	newPath, err := afLedgerPath(f.db)
	if err != nil {
		return err
	}
	path, oldPath := newPath, ""
	if from != "" {
		oldDir, derr := homestate.ProjectDir(afResolveGone(from))
		if derr != nil {
			return fmt.Errorf("af: refused: cannot resolve the ledger location of --from %q: %v; nothing was changed", from, derr)
		}
		oldPath = filepath.Join(oldDir, "af", "ledger.db")
		switch {
		case oldPath == newPath:
			return fmt.Errorf("af: refused: --from %q has the same ledger location as this project (%s), so there is nothing to move; nothing was changed", from, newPath)
		case afFileExists(newPath):
			return fmt.Errorf("af: refused: the ledger of %q is at %s and this project already has a ledger at %s; moving one over the other would lose one, "+
				"so neither was touched. If the ledger at %s is an empty one a verb made there, move it aside and run this again; nothing was changed", from, oldPath, newPath, newPath)
		case !afFileExists(oldPath):
			return fmt.Errorf("af: refused: there is no ledger at %s (the default location of --from %q); nothing was changed", oldPath, from)
		}
		path = oldPath
	} else if _, serr := os.Stat(path); serr != nil {
		hint := ""
		if f.db == "" {
			hint = "; if the project directory moved, name its old place with --from <old project directory>"
		}
		return fmt.Errorf("af: refused: there is no ledger at %s (%v)%s; nothing was changed", path, serr, hint)
	}
	ctx, cancel := context.WithTimeout(context.Background(), afOpTimeout)
	defer cancel()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		return fmt.Errorf("af: open the ledger %s: %w", path, err)
	}
	defer func() { _ = st.Close() }() // st changes with --from: the deferred call closes whichever is open at the end

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
		return afStillThere(ctx, st, f, current)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("af: refused: cannot tell whether the repository this ledger is bound to (%s) is gone: %v; nothing was changed", current, err)
	}
	// (b) the new repository is a repository, and it is the same one
	dir := afMoveDir(f)
	next, err := afGitCommonPath(ctx, dir)
	if err != nil {
		if errors.Is(err, errAFNotARepository) {
			return fmt.Errorf("af: refused: %q (--bind-repo, else the project directory) is not a directory inside a git repository; nothing was changed", dir)
		}
		return fmt.Errorf("af: refused: cannot tell which repository %q belongs to (%w); nothing was changed", dir, err)
	}
	marker, checked, unchecked, err := afMoveProof(ctx, st, f, path, dir, next)
	if err != nil {
		return err
	}
	res := map[string]any{"from": current, "to": next, "reason": reason}
	// (c) with --from the ledger file goes to the new key first, and back if the record cannot be made
	if oldPath != "" {
		_ = st.Close()
		if err := afMoveLedgerFiles(oldPath, newPath); err != nil {
			return fmt.Errorf("af: refused: %v; the ledger at %s was not touched; nothing was changed", err, oldPath)
		}
		if st, err = ledger.Open(ctx, newPath); err != nil {
			_ = afMoveLedgerFiles(newPath, oldPath)
			return fmt.Errorf("af: open the moved ledger %s: %w; the ledger was put back at %s; nothing was changed", newPath, err, oldPath)
		}
		res["ledger_moved_from"] = oldPath
		res["ledger"] = newPath
	}
	if err := st.MoveRepositoryMarked(ctx, current, next, reason, marker); err != nil {
		if oldPath != "" {
			_ = st.Close()
			if rerr := afMoveLedgerFiles(newPath, oldPath); rerr != nil {
				return fmt.Errorf("af: refused: %w; the ledger could not be put back at %s either (%v) and is at %s", err, oldPath, rerr, newPath)
			}
		}
		return fmt.Errorf("af: refused: %w; nothing was changed", err)
	}
	res["commits_checked"], res["commits_unchecked"] = checked, unchecked
	return afPrint(cmd, res)
}

// afStillThere is the refusal of a move while the repository the ledger is bound to is still on disk. A ledger that a verb
// just made at the default location of a project directory that moved (no cards, no moves, bound to the repository this
// call names) is not "a repository that did not move": it is a fresh ledger for this place, and the ledger with the cards
// is at the old project directory's key.
func afStillThere(ctx context.Context, st *ledger.Store, f afFlags, current string) error {
	if f.db == "" {
		cards, cerr := st.Cards(ctx)
		moves, merr := st.RepositoryMoves(ctx)
		if next, nerr := afGitCommonPath(ctx, afMoveDir(f)); cerr == nil && merr == nil && nerr == nil && len(cards) == 0 && len(moves) == 0 && next == current {
			return fmt.Errorf("af: refused: the ledger at the default location of this project is a fresh one for this place (no cards, no moves, bound to %s, "+
				"the repository this call names), so there is nothing here to move. If the project directory moved and the ledger with the cards was made at its old "+
				"place, run this again with --from <old project directory>; nothing was changed", current)
		}
	}
	return fmt.Errorf("af: refused: the repository this ledger is bound to still exists on disk (%s), so it did not move; "+
		"nothing was changed", current)
}

// afMoveProof decides that the repository at next (named through dir) is the one the ledger belongs to, and returns the
// marker to record with the move and the commits counted. A ledger with a marker on record needs nothing else: the
// repository carries that marker or it is not the repository, and the commits are only counted (what git cannot confirm
// is reported as unchecked, a pruned commit of a dead attempt is no reason to refuse a real move, N67). A ledger without
// one falls back to the commits of the rows that are not failed, a git that can answer for them, a declared repository
// and the ledger file's own location, and takes the marker of the new repository.
func afMoveProof(ctx context.Context, st *ledger.Store, f afFlags, ledgerPath, dir, next string) (marker string, checked, unchecked int, err error) {
	recorded, err := st.RepositoryMarker(ctx)
	if err != nil {
		return "", 0, 0, err
	}
	all, required, err := afRowCommitSets(ctx, st)
	if err != nil {
		return "", 0, 0, err
	}
	missing, merr := afMissingCommits(ctx, dir, all)
	if merr == nil {
		checked, unchecked = len(all)-len(missing), len(missing)
	} else {
		checked, unchecked = 0, len(all)
	}
	if recorded != "" {
		found, rerr := afReadMarker(next)
		if rerr != nil {
			return "", 0, 0, fmt.Errorf("af: refused: cannot read the identity marker of the repository %s (the git directory of %s): %v; nothing was changed", next, dir, rerr)
		}
		if found != recorded {
			return "", 0, 0, afMarkerRefusal(next, dir, recorded, found)
		}
		return recorded, checked, unchecked, nil
	}
	// the ledger has no marker: the commits are the only evidence left
	if merr != nil {
		return "", 0, 0, fmt.Errorf("af: refused: this ledger holds no identity marker, so the commits of its rows are the evidence, but git could not answer whether "+
			"the repository %s (the git directory of %s) has them (%v); nothing was changed", next, dir, merr)
	}
	gone := map[string]bool{}
	for _, sha := range missing {
		gone[sha] = true
	}
	for _, sha := range required {
		if gone[sha] {
			return "", 0, 0, fmt.Errorf("af: refused: the queue rows of this ledger name the commit %s, which the repository %s (the git directory of %s) does not have; "+
				"the commits of a ledger live in its repository, so this is not where it moved to; nothing was changed", sha, next, dir)
		}
	}
	if f.bindRepo == "" {
		return "", 0, 0, fmt.Errorf("af: refused: this ledger holds no identity marker (it was bound before markers existed and not used since), so nothing but the "+
			"operator's word says that %s (the git directory of %s) is the repository that moved; declare it with --bind-repo <a path in it>, the project directory is "+
			"not taken for a declaration; nothing was changed", next, dir)
	}
	if f.db != "" {
		if loc, lerr := afLedgerFileLocation(ledgerPath); lerr == nil {
			if got, gerr := afGitCommonPath(ctx, filepath.Dir(loc)); gerr == nil && got != next {
				return "", 0, 0, fmt.Errorf("af: refused: the ledger file %s sits in the repository %s, and the repository this call names is %s (the git directory of %s); "+
					"a ledger file that moved with its repository sits in the new one; nothing was changed", loc, got, next, dir)
			}
		}
	}
	marker, err = afEnsureMarker(next)
	if err != nil {
		return "", 0, 0, fmt.Errorf("af: refused: the repository %s has no identity marker and one cannot be written in its git directory (%v); nothing was changed", next, err)
	}
	return marker, checked, unchecked, nil
}

// afResolveGone is the spelling the key of a project directory was made from, for a directory that is no longer there: the
// project key resolves symbolic links of a directory that exists, and a directory that moved away cannot be resolved, so
// the nearest ancestor that still exists is resolved and the rest is appended. A directory that exists comes back resolved.
func afResolveGone(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	rest, cur := "", abs
	for {
		if real, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(real, rest)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}

func afFileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// afMoveLedgerFiles renames the ledger file and its -wal and -shm companions from one path to another, all or none. The
// ledger was closed first, which checkpoints the log, so a companion is normally gone; one that is still there is moved
// with the file. A destination that exists, or a rename that fails (another file system, say), undoes what was done.
func afMoveLedgerFiles(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o700); err != nil {
		return err
	}
	var done []string
	undo := func() {
		for i := len(done) - 1; i >= 0; i-- {
			_ = os.Rename(to+done[i], from+done[i])
		}
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if !afFileExists(from + suffix) {
			continue
		}
		if afFileExists(to + suffix) {
			undo()
			return fmt.Errorf("%s already exists", to+suffix)
		}
		if err := os.Rename(from+suffix, to+suffix); err != nil {
			undo()
			return fmt.Errorf("cannot move %s: %v", from+suffix, err)
		}
		done = append(done, suffix)
	}
	return nil
}
