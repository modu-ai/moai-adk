package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor/flow"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/session"
)

// afOpTimeout bounds one verb: the ledger write is short and a verb never waits on a person.
const afOpTimeout = 30 * time.Second

// newAFCmd is the `moai af` command: the verbs of the card flow over the local ledger (AC-AFL-017). One verb per
// ledger method the flow needs, and nothing else: add (a card), claim, move (stage move), evidence (record), enqueue,
// release, reap, land, reconcile (the queue's entry points), resume (the exit from a park) and move-repository (N52).
// Each verb opens the ledger, makes its one call, prints one JSON object and exits; the ledger owns every rule
// (fencing epochs, the position table, the evidence binding), so a verb adds none.
func newAFCmd() *cobra.Command {
	af := &cobra.Command{
		Use:          "af",
		Short:        "Agentic factory card flow over the local ledger",
		SilenceUsage: true,
	}
	af.AddCommand(newAFAddCmd(), newAFClaimCmd(), newAFMoveCmd(), newAFEvidenceCmd(), newAFEnqueueCmd(), newAFReleaseCmd(), newAFReapCmd(), newAFLandCmd(), newAFReconcileCmd(), newAFResumeCmd(), newAFMoveRepositoryCmd())
	return af
}

// afDBFlagHelp is what --db says on every verb that opens the ledger. The default is af/ledger.db in the project's state
// directory, which homestate.ProjectDir keys by the project (<MoAI home>/db/<project name>-<hash>/), not in the project
// directory itself (N72).
const afDBFlagHelp = "ledger file (default: <MoAI home>/db/<project name>-<hash>/af/ledger.db, keyed by the project directory; " +
	"it is not inside the project directory)"

// afDoneHelp is what the verbs that end a card say about DONE (N73, the E6 decision): the guarantee is about the commit,
// not about the work it holds.
const afDoneHelp = "DONE guarantees that the candidate commit landed in the trunk; it does not guarantee the card's work content " +
	"(a candidate that is the trunk's own commit lands as already in the trunk)."

// afLedgerPath is the ledger file of this project: af/ledger.db in the project's state directory
// (homestate.ProjectDir). --db overrides it (tests, or a ledger kept elsewhere).
func afLedgerPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	dir, err := homestate.ProjectDir(resolveProjectDir())
	if err != nil {
		return "", fmt.Errorf("af: cannot resolve the project directory: %w", err)
	}
	return filepath.Join(dir, "af", "ledger.db"), nil
}

// afPlace is what a call says about its ledger: db is the --db it was given ("" for the default location, which says
// nothing about the project, N50), and bindRepo the --bind-repo it declared (N51), used for the first bind only.
type afPlace struct{ db, bindRepo string }

// afRun opens the ledger, runs fn with a bounded context and closes the ledger.
func afRun(f afFlags, fn func(ctx context.Context, st *ledger.Store) error) error {
	db := f.db
	path, err := afLedgerPath(db)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), afOpTimeout)
	defer cancel()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		return fmt.Errorf("af: open the ledger %s: %w", path, err)
	}
	defer func() { _ = st.Close() }()
	// every verb binds the ledger to the repository of the project directory it runs in; the verbs that look commits up
	// bind the repository they use as well (afCheckCommit), and a project directory that is no repository has none to bind.
	// The ledger file is a clue only when the caller placed it with --db (N50): the default place is the tool's choice.
	if err := afBindRepository(ctx, st, f.place(), resolveProjectDir(), true); err != nil {
		return err
	}
	return fn(ctx, st)
}

// afBindRepository binds the ledger to the repository that dir belongs to (its git directory, which a linked worktree
// shares) or checks that it is the repository the ledger is already bound to. A call that would use another repository
// is refused with the ledger unchanged. A directory that is no git repository has no repository to compare:
// optionalRepo lets a verb that runs no git carry on there (a refusal for anything git could not answer).
//
// While the ledger is not bound yet the first call would decide, so before it binds it is compared with what the ledger
// already knows (afCheckKnownRepositories); place says whether the caller placed the ledger file (--db) and which
// repository it declared (--bind-repo).
func afBindRepository(ctx context.Context, st *ledger.Store, place afPlace, dir string, optionalRepo bool) error {
	id, err := afGitCommonPath(ctx, dir)
	if err != nil {
		if optionalRepo && errors.Is(err, errAFNotARepository) {
			return nil
		}
		return fmt.Errorf("af: cannot tell which repository %q belongs to (%w); nothing was changed", dir, err)
	}
	if bound, rerr := st.Repository(ctx); rerr == nil && bound == "" {
		if err := afCheckKnownRepositories(ctx, st, place, id, dir); err != nil {
			return err
		}
	}
	if err := st.BindRepository(ctx, id); err != nil {
		if !errors.Is(err, ledger.ErrWrongRepository) {
			return fmt.Errorf("af: bind the ledger to the repository %s: %w", id, err)
		}
		bound, _ := st.Repository(ctx)
		return fmt.Errorf("af: refused: this ledger is bound to the repository %s and this call would use %s (the git directory of %s); "+
			"nothing was changed", bound, id, dir)
	}
	return afCheckMarker(ctx, st, id, dir)
}

// afCheckMarker settles the repository identity marker once the identity (the git directory) is the ledger's (N61 N65
// N67). A ledger with a marker on record accepts only the repository whose git directory holds that marker: a clone, a
// mirror, a re-created repository or any stranger at the path has none or another. A ledger with none on record (bound by
// a build before markers, or just bound now) takes the marker of the repository it is bound to, created when the
// repository has none, and keeps it for good.
func afCheckMarker(ctx context.Context, st *ledger.Store, id, dir string) error {
	have, err := st.RepositoryMarker(ctx)
	if err != nil {
		return fmt.Errorf("af: read the repository marker of the ledger: %w", err)
	}
	if have == "" {
		mine, err := afEnsureMarker(id)
		if err != nil {
			return fmt.Errorf("af: refused: the repository %s (the git directory of %s) has no identity marker and one cannot be written "+
				"in its git directory (%v); the ledger is bound to it but holds no marker yet, and nothing else was changed", id, dir, err)
		}
		got, err := st.RecordRepositoryMarker(ctx, mine)
		if err != nil {
			return fmt.Errorf("af: record the repository marker: %w", err)
		}
		if got != mine {
			return afMarkerRefusal(id, dir, got, mine)
		}
		return nil
	}
	mine, err := afReadMarker(id)
	if err != nil {
		return fmt.Errorf("af: refused: cannot read the identity marker of the repository %s (the git directory of %s): %v; nothing was changed", id, dir, err)
	}
	if mine != have {
		return afMarkerRefusal(id, dir, have, mine)
	}
	return nil
}

// afMarkerRefusal is the refusal of a repository whose identity marker is not the one the ledger recorded; it names the
// repository and the first characters of the two markers, never a marker value.
func afMarkerRefusal(id, dir, recorded, found string) error {
	what := "carries the identity marker " + afMarkerPrefix(found) + "..."
	if found == "" {
		what = "carries no identity marker"
	}
	return fmt.Errorf("af: refused: the repository at %s (the git directory of %s) is not the one this ledger belongs to: it %s, and the ledger "+
		"recorded %s... for its repository. A repository that was moved or renamed carries its identity marker with it; a clone, a mirror or a "+
		"repository made afterwards does not; nothing was changed", id, dir, what, afMarkerPrefix(recorded))
}

// afCheckKnownRepositories refuses to bind id (the repository of dir) while the ledger already knows another one: the
// repository of any worktree a lease records, the repository the ledger file itself sits in (place.db, the --db the
// caller gave; the default location is no clue), and the repository that has the commits the rows name. A lease worktree
// that is gone or is no repository is no clue and is skipped; a relative one is refused (N51 d). A lease worktree that
// agrees with id is not a clue either (N65): it is the caller's own word.
//
// A ledger given with --db whose file sits in no repository and that holds no commit knows nothing (N51, the clone form):
// the first call would decide, and a call from a second clone of the project would decide wrongly with nothing ever
// refused afterwards. Such a first bind needs the caller to declare the repository (place.bindRepo). A declaration is also
// checked when it is given and a clue exists. The default location needs none.
func afCheckKnownRepositories(ctx context.Context, st *ledger.Store, place afPlace, id, dir string) error {
	clue := false
	// check refuses a worktree or file location in another repository than id. proof says whether agreeing with id is a
	// clue that the ledger belongs to it: the repository the ledger file sits in is, a lease worktree is not (N65), since
	// the caller of a claim names it and a claim from a directory that is no repository records whatever it is given.
	check := func(where, why string, proof bool) error {
		got, err := afGitCommonPath(ctx, where)
		if err != nil {
			return nil // no clue (gone, no repository, git could not answer)
		}
		if got == id {
			if proof {
				clue = true
			}
			return nil
		}
		return fmt.Errorf("af: refused: this ledger is not bound yet, but it already belongs to the repository %s (%s) and this call would bind "+
			"%s (the git directory of %s); nothing was changed", got, why, id, dir)
	}
	if place.bindRepo != "" {
		declared, err := afGitCommonPath(ctx, place.bindRepo)
		if err != nil {
			return fmt.Errorf("af: refused: --bind-repo %q is not a directory inside a git repository; nothing was changed", place.bindRepo)
		}
		if declared != id {
			return fmt.Errorf("af: refused: --bind-repo %q declares the repository %s, and this call would bind %s (the git directory of %s); "+
				"nothing was changed", place.bindRepo, declared, id, dir)
		}
	}
	if leases, err := st.Leases(ctx); err == nil {
		for _, l := range leases {
			if l.Worktree == "" {
				continue
			}
			if !filepath.IsAbs(l.Worktree) {
				// git -C <relative> runs in the caller's directory, so the comparison would always agree with the
				// caller (N51 d); a clue that cannot be read the same by everyone is a reason to stop, not to bind
				return fmt.Errorf("af: refused: this ledger is not bound yet, and the lease on %s records the worktree %q, which is not an absolute path "+
					"and names whatever directory the caller stands in; release that lease (from a directory that is no git repository, where the release needs no bind) "+
					"and claim the card again with the absolute worktree; this call would bind %s (the git directory of %s); nothing was changed", l.CardID, l.Worktree, id, dir)
			}
			if err := check(l.Worktree, "the worktree "+l.Worktree+" of the lease on "+l.CardID, false); err != nil {
				return err
			}
		}
	}
	if place.db != "" {
		if loc, err := afLedgerFileLocation(place.db); err == nil {
			if err := check(filepath.Dir(loc), "the ledger file "+loc+" sits in it", true); err != nil {
				return err
			}
		}
	}
	// The rows name commits (evidence subjects, queue candidates): the repository this call would bind must have them
	// (N51 b). A ledger with no such row is no clue. A git that cannot answer is a refusal and not a clue (N62): a
	// comparison that gave no answer says nothing, and counting it as agreement bound the ledger to whichever
	// repository had a damaged object.
	shas, rerr := afRowCommits(ctx, st)
	if rerr != nil {
		return fmt.Errorf("af: refused: this ledger is not bound yet and its rows could not be read (%v), so what they say about the repository is not known; nothing was changed", rerr)
	}
	if len(shas) > 0 {
		missing, err := afMissingCommit(ctx, dir, shas)
		if err != nil {
			return fmt.Errorf("af: refused: this ledger is not bound yet and its rows name commits, but git could not answer whether the repository %s "+
				"(the git directory of %s) has them (%v); a comparison that gave no answer is no clue, and this call would bind that repository; "+
				"nothing was changed", id, dir, err)
		}
		if missing != "" {
			return fmt.Errorf("af: refused: this ledger is not bound yet, but its rows name the commit %s, which the repository %s "+
				"(the git directory of %s) does not have; the commits of a ledger live in its repository, and this call would bind another one; "+
				"nothing was changed", missing, id, dir)
		}
		clue = true
	}
	if !clue && place.db != "" && place.bindRepo == "" {
		return fmt.Errorf("af: refused: this ledger (--db %s) is not bound yet, sits in no repository and holds no commit that says which "+
			"repository it belongs to (a lease worktree is the caller's own word and does not say), so the first call would decide, and a call "+
			"from a second clone or another repository would decide wrongly with nothing refused afterwards; declare the repository this "+
			"ledger belongs to with --bind-repo <a path in it> (this call would bind %s, the git directory of %s); nothing was changed", place.db, id, dir)
	}
	return nil
}

// afRowCommits are the commits the ledger's rows name: the subject of every evidence row and the candidate of every queue
// row, each once, complete SHAs only (the ledger takes any non-empty string as a subject, "n/a" of a resume row
// included, and a name that is not a commit says nothing about a repository).
func afRowCommits(ctx context.Context, st *ledger.Store) ([]string, error) {
	all, _, err := afRowCommitSets(ctx, st)
	return all, err
}

// afRowCommitSets are all the commits the rows name, and the part of them a move of the repository cannot do without:
// the candidates of the queue rows that are not failed (a waiting, verifying, landing or landed row is the work the move
// is for). The rest, the subjects of evidence rows and the candidates of failed rows, are the past: a commit of an attempt
// that ended in failure is thrown away and collected in time, and a move refused for it has no way out (N67).
func afRowCommitSets(ctx context.Context, st *ledger.Store) (all, required []string, err error) {
	seen, seenReq := map[string]bool{}, map[string]bool{}
	add := func(sha string, req bool) {
		if !mergeq.IsCompleteSHA(sha) {
			return
		}
		if !seen[sha] {
			seen[sha] = true
			all = append(all, sha)
		}
		if req && !seenReq[sha] {
			seenReq[sha] = true
			required = append(required, sha)
		}
	}
	cards, err := st.Cards(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range cards {
		recs, err := st.EvidenceFor(ctx, c.ID)
		if err != nil {
			return nil, nil, err
		}
		for _, r := range recs {
			add(r.SubjectSHA, false)
		}
	}
	rows, err := st.Queue(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range rows {
		add(r.CandidateSHA, r.State != ledger.QueueFailed)
	}
	return all, required, nil
}

// afMissingCommits are the shas that are no commit of the repository dir belongs to, in the order given. One git process
// answers for all of them. A git that cannot answer is an error and never "nothing is missing" (N62): a failure, a
// time-out (the context ends the process), an answer of another length than the question, and a line that is neither a
// commit nor a "missing" are all refusals to answer, since each of them used to read as agreement.
func afMissingCommits(ctx context.Context, dir string, shas []string) ([]string, error) {
	if len(shas) == 0 {
		return nil, nil
	}
	var in strings.Builder
	for _, sha := range shas {
		in.WriteString(sha + "^{commit}\n")
	}
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "cat-file", "--batch-check")
	cmd.Env = append(gitenv.Env(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdin = strings.NewReader(in.String())
	out, err := cmd.Output()
	if err != nil {
		if cerr := ctx.Err(); cerr != nil {
			return nil, fmt.Errorf("the check was cut off: %w", cerr)
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("git cat-file exited with status %d", ee.ExitCode())
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) != len(shas) {
		return nil, fmt.Errorf("git answered %d lines to %d questions", len(lines), len(shas))
	}
	var missing []string
	for i, line := range lines {
		f := strings.Fields(line)
		switch {
		case len(f) == 2 && f[1] == "missing":
			missing = append(missing, shas[i])
		case len(f) == 3 && f[1] == "commit":
		default:
			return nil, fmt.Errorf("git answered a line that is neither a commit nor a missing one (line %d)", i+1)
		}
	}
	return missing, nil
}

// afMissingCommit is the first of shas that is no commit of the repository dir belongs to, "" when all are. The error
// says git could not answer; callers refuse on it (N62).
func afMissingCommit(ctx context.Context, dir string, shas []string) (string, error) {
	missing, err := afMissingCommits(ctx, dir, shas)
	if err != nil || len(missing) == 0 {
		return "", err
	}
	return missing[0], nil
}

// afLedgerFileLocation is where the ledger file really is: absolute, with symbolic links resolved. A link outside any
// repository to a file inside one would otherwise be asked about the link's directory (N51 c). A file that does not
// exist yet has its directory resolved instead.
func afLedgerFileLocation(dbPath string) (string, error) {
	abs, err := filepath.Abs(dbPath)
	if err != nil {
		return "", err
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return real, nil
	}
	dir, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if err != nil {
		return abs, nil
	}
	return filepath.Join(dir, filepath.Base(abs)), nil
}

// afCommitRepo is the directory a verb looks a card's commit up in: --repo, else the worktree of the card's lease, else
// the project directory.
func afCommitRepo(ctx context.Context, st *ledger.Store, card, repoFlag string) string {
	if repoFlag != "" {
		return repoFlag
	}
	if leases, err := st.Leases(ctx); err == nil {
		for _, l := range leases {
			if l.CardID == card && l.Worktree != "" {
				return l.Worktree
			}
		}
	}
	return resolveProjectDir()
}

// afCheckCommit refuses a commit name that is not the complete lowercase hex SHA (40, or 64 for SHA-256) of a commit
// that exists in the repository: an abbreviation, HEAD or a branch name can name another commit tomorrow, so
// evidence or a queue entry bound to one is not bound to anything, and a value that starts with '-' reads as an
// option to git. The repository is --repo, else the worktree of the card's lease, else the project directory; it must be
// the repository the ledger is bound to, and the directory that was used is returned. f carries the verb's --db and
// --bind-repo: while the ledger is not bound yet the repository used here may be the first one to bind it, and it is
// compared with the ledger file's own location exactly as afRun does (N49).
func afCheckCommit(ctx context.Context, st *ledger.Store, f afFlags, card, repoFlag, sha, what string) (string, error) {
	if !mergeq.IsCompleteSHA(sha) {
		return "", fmt.Errorf("af: %s %q is not a complete lowercase hex commit SHA (40 or 64 characters)", what, sha)
	}
	repo := afCommitRepo(ctx, st, card, repoFlag)
	// afRun bound the ledger to the project directory's repository before this (when that is a repository); when the
	// project directory is none, this call is the first bind, and it asks the same questions
	if err := afBindRepository(ctx, st, f.place(), repo, false); err != nil {
		return "", err
	}
	if got, err := (mergeq.GitCLI{Repo: repo}).RevParse(ctx, sha); err != nil || got != sha {
		return "", fmt.Errorf("af: %s %s names no commit in %s", what, sha, repo)
	}
	return repo, nil
}

// afInTrunk asks the repository whether sha is in the dedicated trunk (exit 0 of the ancestor test: yes, exit 1: no).
// Any other answer, a missing trunk ref included, is an error and not a guess.
func afInTrunk(ctx context.Context, repo, sha string) (bool, error) {
	in, err := (mergeq.GitCLI{Repo: repo}).IsAncestor(ctx, sha, mergeq.DefaultTrunk)
	if err != nil {
		return false, fmt.Errorf("af: cannot tell whether %s is in %s of %s: %w; nothing was changed", sha, mergeq.DefaultTrunk, repo, err)
	}
	return in, nil
}

// afCheckLandedInTrunk refuses to finish a card whose queue row says landed while its commit is not in the dedicated
// trunk (the row landed on another trunk, or the trunk lost the commit since): the row alone says "merged" without
// asking where. The way out is a real landing, so the message says how to get one.
func afCheckLandedInTrunk(ctx context.Context, st *ledger.Store, card string) error {
	rows, err := st.Queue(ctx)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r.CardID != card || r.State != ledger.QueueLanded {
			continue
		}
		repo := resolveProjectDir()
		in, err := afInTrunk(ctx, repo, r.CandidateSHA)
		if err != nil {
			return err
		}
		if !in {
			return fmt.Errorf("af: %s cannot move to DONE: its queue row says %s landed, but that commit is not in %s of %s; "+
				"nothing was changed. Withdraw the card (move to REVIEWING) and release it; the next holder measures the commit again "+
				"and offers it with enqueue, the row starts again as waiting, and `moai af land` lands it for real", card, r.CandidateSHA, mergeq.DefaultTrunk, repo)
		}
	}
	return nil
}

// afRepoFlagHelp is what --repo says of the verbs that look a commit up; the land verbs say the same in their own words.
const afRepoFlagHelp = "the git repository the commit is looked up in: the project's own (default: the card's lease worktree, " +
	"else the project directory), or any path with the same git directory, a linked worktree of it included; any other path is refused, " +
	"and so is a repository other than the one the ledger is bound to"

// afBindHelp closes the help of every verb that opens the ledger: what the repository binding is and what it refuses.
const afBindHelp = `The ledger is bound to one repository: the first verb that writes it binds the git directory of the repository it
runs in (a linked worktree has the git directory of its repository, and two spellings of one path are one repository). That
is the project directory, which CLAUDE_PROJECT_DIR or else the working directory names; evidence and enqueue also use --repo
or the card's lease worktree. Afterwards a call that would use another repository is refused with the ledger unchanged, the
message naming both. While the ledger is not bound yet, the repository a call would bind is first compared with what the
ledger already knows: the repository of every lease worktree it records and, for a ledger given with --db, the repository
the ledger file itself sits in (the default ledger location is not compared; a lease worktree that is gone or is no
repository is no clue), and the repository that has the commits its rows name; a disagreement is refused the same way, and
so is a git that cannot say whether the repository has those commits. A lease worktree is compared but never proves where
the ledger belongs: the caller of a claim names it. A ledger given with --db that sits in no repository and holds no commit
knows nothing: the first call would decide, so it is refused unless it declares the repository with --bind-repo <a path in
it> (the default ledger location needs no declaration). Once bound, the ledger also records an identity marker kept in the
git directory of the repository, and a repository without that marker (a clone, a mirror, one made afterwards at the same
path) is refused. A project directory that is no git repository binds nothing, for the verbs that run no git.`

// afCheckTrunk refuses every --trunk but the dedicated one. A landing row says "merged" without asking which trunk or
// which repository it landed in, so a verb that took another ref (the candidate's own branch, say) ended the card DONE
// with the real trunk empty. The trunk of this queue is mergeq.DefaultTrunk and cannot be changed; the Merger library
// keeps its Trunk field, and the verbs are the trust boundary that fixes it.
func afCheckTrunk(trunk string) error {
	if trunk != mergeq.DefaultTrunk {
		return fmt.Errorf("af: --trunk %q refused: the trunk is %s and cannot be changed; nothing was landed and the ledger is unchanged", trunk, mergeq.DefaultTrunk)
	}
	return nil
}

// errAFNotARepository is what afGitCommonPath returns for a directory git says is in no repository.
var errAFNotARepository = errors.New("not in a git repository")

// afGitCommonPath is the identity of the repository dir belongs to: the absolute git directory shared by dir and every
// linked worktree of the repository, symlinks resolved so two spellings of one path give one identity.
func afGitCommonPath(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--git-common-dir")
	cmd.Env = append(gitenv.Env(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", errAFNotARepository // git ran and found no repository
		}
		return "", err
	}
	common := strings.TrimSpace(string(out))
	if common == "" {
		return "", errors.New("git printed no git directory")
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(dir, common) // git prints it relative to the directory it ran in
	}
	abs, err := filepath.Abs(common)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// afGitCommonDir is the git directory shared by dir and every linked worktree of its repository.
func afGitCommonDir(ctx context.Context, dir string) (os.FileInfo, error) {
	common, err := afGitCommonPath(ctx, dir)
	if err != nil {
		return nil, err
	}
	return os.Stat(common)
}

// afCheckRepoFlag refuses a --repo that is not the project's own repository: a path whose git directory is the project
// directory's (a linked worktree of the same repository has it too). Evidence, a queue entry and a landing made against
// another repository say nothing about the project's, and the card could end DONE without the commit in it. An empty
// repo (the flag left out) is accepted: the verbs then use the lease's worktree or the project directory themselves.
// It runs before the ledger is opened, so a refusal changes nothing.
func afCheckRepoFlag(ctx context.Context, repo string) error {
	if repo == "" {
		return nil
	}
	project := resolveProjectDir()
	want, err := afGitCommonDir(ctx, project)
	if err != nil {
		return fmt.Errorf("af: --repo %s refused: the project directory %q is not in a git repository, so there is no repository to compare it with", repo, project)
	}
	got, err := afGitCommonDir(ctx, repo)
	if err != nil {
		return fmt.Errorf("af: --repo %s refused: it is not a directory inside a git repository", repo)
	}
	if !os.SameFile(want, got) {
		return fmt.Errorf("af: --repo %s refused: its git directory is not the project's (the project directory is %s); "+
			"only the project's repository, or a linked worktree of it, is accepted, and nothing was changed", repo, project)
	}
	return nil
}

func afPrint(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	return enc.Encode(v)
}

// afFlags are the flags every verb carries.
type afFlags struct {
	db       string
	bindRepo string
	epoch    int64
}

func (f afFlags) place() afPlace { return afPlace{db: f.db, bindRepo: f.bindRepo} }

// afBindRepoFlagHelp is what --bind-repo says on every verb that opens the ledger.
const afBindRepoFlagHelp = "the repository this ledger belongs to, declared for its first bind: any path inside it (a linked worktree " +
	"included). Needed only for a ledger given with --db that sits in no repository and holds no commit that " +
	"says where it belongs (a lease worktree does not); a path in another repository than the one the call would bind is refused; ignored once the ledger is bound"

func (f *afFlags) bind(cmd *cobra.Command, needEpoch bool) {
	cmd.Flags().StringVar(&f.db, "db", "", afDBFlagHelp)
	cmd.Flags().StringVar(&f.bindRepo, "bind-repo", "", afBindRepoFlagHelp)
	if needEpoch {
		cmd.Flags().Int64Var(&f.epoch, "epoch", 0, "the fencing epoch the claim returned")
		_ = cmd.MarkFlagRequired("epoch")
	}
}

// afOwnerPID and afProbe are the seams of afHolder.
var (
	afOwnerPID = session.ResolveOwnerPID
	afProbe    = homestate.ProbeProcessIdentity
)

// afHolder is the process the lease records as the holder of the card. The command exits at once and is usually
// started through a shell wrapper that exits with it, so neither this process nor its parent can be the holder:
// a lease naming either reads as a dead holder at the next reap and the card is taken from a live session. The
// holder is the long-lived session that owns this process (the same resolution the session registry uses). When
// it cannot be resolved the holder is recorded as unknown (pid 0): the ledger never reclaims an unknown holder.
func afHolder() (pid int, start string) {
	owner, ok := afOwnerPID()
	if !ok || owner < 1 {
		return 0, ""
	}
	start, _ = afProbe(owner)
	if start == "" {
		// A pid without its fingerprint cannot be judged (a reused pid would pass for the session), so the
		// holder is recorded as unknown rather than leaving the verdict to the judge.
		return 0, ""
	}
	return owner, start
}

func newAFClaimCmd() *cobra.Command {
	var f afFlags
	var sessionID, worktree string
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Claim the next card and print its id and fencing epoch",
		Long: `Claim the next card and print its id and fencing epoch.

--worktree must be an absolute path (a relative one is refused: it would name a different directory for every caller) and
is recorded in the lease as given; it is not checked against a repository here: evidence and enqueue look a commit up in
it when --repo is left out, and refuse it there when it is in a repository other than the ledger's.

` + afBindHelp,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if worktree == "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				worktree = wd
			}
			// A lease worktree is looked up later from whichever directory the next verb runs in, so a relative one names
			// a different place for every caller (N51 d): it is refused here, where it would be recorded.
			if !filepath.IsAbs(worktree) {
				return fmt.Errorf("af claim: --worktree %q is not an absolute path: a lease worktree is looked up from the directory "+
					"of whoever asks next, so a relative one names a different place for each of them; give the absolute path; nothing was changed", worktree)
			}
			pid, start := afHolder()
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				c, err := st.Claim(ctx, ledger.ClaimRequest{SessionID: sessionID, Worktree: worktree, PID: pid, PIDStart: start, Alive: afAlive})
				if err != nil {
					return err
				}
				if path, perr := afLedgerPath(f.db); perr == nil {
					// the session's hooks renew the lease from this binding; a failure only costs the renewal
					if berr := hook.BindLedgerHeartbeat(resolveProjectDir(), sessionID, hook.LedgerBinding{DB: path, CardID: c.CardID, Epoch: c.Epoch}); berr != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "af: the lease will not be renewed by hooks: %v\n", berr)
					}
				}
				return afPrint(cmd, map[string]any{
					"card_id": c.CardID, "epoch": c.Epoch,
					"stage": c.Position.Stage, "sub": c.Position.Sub, "expires_at": c.ExpiresAt,
				})
			})
		},
	}
	f.bind(cmd, false)
	cmd.Flags().StringVar(&sessionID, "session", "", "the claiming session id")
	cmd.Flags().StringVar(&worktree, "worktree", "", "the worktree the card is worked in (default: the working directory)")
	_ = cmd.MarkFlagRequired("session")
	return cmd
}

func newAFMoveCmd() *cobra.Command {
	var f afFlags
	var stage, sub, actor, reason string
	var remote bool
	cmd := &cobra.Command{
		Use:   "move <card>",
		Short: "Move a claimed card to a stage",
		Long: `Move a claimed card to a stage.

Limits: the verb carries no guard or evidence arguments. A move that needs a guard key (an operator hold, a
fix-limit or reclaim-cap park, a resume guard) or evidence handed to the position table is refused by the
ledger, and this verb cannot supply either; such moves are made through the ledger API, except the resume of a
reclaim-cap park, which is the separate "resume" verb. Recording evidence is the separate "evidence" verb.

Moving to DONE: when the card's queue row says landed, the verb asks the repository whether that commit is in ` + mergeq.DefaultTrunk + `
and refuses when it is not, or when the repository cannot say; nothing is changed. The row alone says "merged" without
asking where, so a trunk that lost the commit, or a row that landed on another trunk, is not a landing the card can finish
on. The way out is a real landing: move the card to REVIEWING, release it, and the next holder measures the commit again
and offers it with enqueue; the row starts again as waiting and "moai af land" lands it.

` + afDoneHelp + `

A MERGING card whose holder is gone (it holds no lease) and whose queue row waits with a candidate whose newest green,
gate or review run failed can be moved, at the card's own --epoch, to REVIEWING without a lease: "moai af land" refuses
such a row, and this move takes the card back so that the next claim measures the commit again.

` + afBindHelp,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			to := core.Position{Stage: core.Stage(strings.ToUpper(stage)), Sub: core.SubStage(strings.ToUpper(sub)), RemoteIntegrated: remote}
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				if to.Stage == core.StageDone {
					if err := afCheckLandedInTrunk(ctx, st, args[0]); err != nil {
						return err
					}
				}
				if err := st.Move(ctx, args[0], f.epoch, to, actor, reason); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "stage": to.Stage, "sub": to.Sub})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&stage, "stage", "", "target stage: READY, BUILDING, REVIEWING, MERGING, DONE or CANCELLED")
	cmd.Flags().StringVar(&sub, "sub", "", "target sub-stage within BUILDING: SPEC, IMPLEMENT or FIX")
	cmd.Flags().BoolVar(&remote, "remote-integrated", false, "MERGING with the merge seen")
	cmd.Flags().StringVar(&actor, "actor", "lane", "who makes the move (recorded as given: nothing masks it, put no secret in it)")
	cmd.Flags().StringVar(&reason, "reason", "", "why (recorded as given: nothing masks it, put no secret in it)")
	_ = cmd.MarkFlagRequired("stage")
	return cmd
}

// newAFResumeCmd is `moai af resume`: the operator's exit from a reclaim-cap park. The third reclaim of a card parks it
// (core.ReasonReclaimCap) and the ledger resolves that park only by an in-place move at the card's own epoch carrying
// the evidence "resume-guard:reclaim-cap". The verb reads the card and hands the ledger that move; it adds no rule and
// does not touch the reclaim cap itself (the ledger clears the count when it accepts the move).
func newAFResumeCmd() *cobra.Command {
	var f afFlags
	var guard, actor, reason string
	cmd := &cobra.Command{
		Use:   "resume <card>",
		Short: "Resume a card the ledger parked at its reclaim cap (an operator decision)",
		Long: `Resume a card the ledger parked after its holders died the third time (park reason reclaim-cap).

A parked card holds no lease and cannot be claimed, moved or finished until the operator decides it may go on. This
verb is that decision: it moves the card in place (same stage, no park) at the epoch the park was made at, with the
ledger's resume guard for that park, and the ledger counts the card's reclaims from zero again. Nothing else changes:
the card is claimed afterwards like any other, its evidence of the earlier epochs does not count for the new holder,
and the reclaim cap itself is untouched.

--guard must be ` + core.GuardReclaimCap + ` (the only park this verb resolves; any other value is refused) and --reason is required and
must say what the operator looked at. A card that is not parked at reclaim-cap, or an --epoch that is not the card's
own, is refused with the ledger unchanged. --actor and --reason are recorded as given, in the card's transition
history: nothing masks them, put no secret in them.

` + afBindHelp,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if guard != core.GuardReclaimCap {
				return fmt.Errorf("af resume: --guard %q refused: this verb resolves the %s park only", guard, core.GuardReclaimCap)
			}
			if strings.TrimSpace(reason) == "" {
				return errors.New("af resume: --reason must say why the card may go on; an empty reason is refused")
			}
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				cards, err := st.Cards(ctx)
				if err != nil {
					return err
				}
				for _, c := range cards {
					if c.ID != args[0] {
						continue
					}
					if c.Position.Park == nil || c.Position.Park.Reason != core.ReasonReclaimCap {
						return fmt.Errorf("af resume: %s is not parked at %s (stage %s); nothing was changed", c.ID, core.GuardReclaimCap, c.Position.Stage)
					}
					to := core.Position{Stage: c.Position.Stage, Sub: c.Position.Sub, RemoteIntegrated: c.Position.RemoteIntegrated}
					// The ledger takes a guard key only inside an Evidence; the row it stores says who resumed, not what was measured.
					ev := ledger.Evidence{Kind: ledger.KindGate, TreeSHA: "n/a", SubjectSHA: "n/a", Result: ledger.ResultPass,
						Command: "moai af resume", Guards: []string{"resume-guard:" + guard}}
					if err := st.Move(ctx, c.ID, f.epoch, to, actor, reason, ev); err != nil {
						return err
					}
					return afPrint(cmd, map[string]any{"card_id": c.ID, "stage": to.Stage, "sub": to.Sub, "resumed": true})
				}
				return fmt.Errorf("af resume: card %s does not exist", args[0])
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&guard, "guard", "", "the park to resolve; only "+core.GuardReclaimCap+" is accepted")
	cmd.Flags().StringVar(&actor, "actor", "operator", "who resumes the card (recorded as given: nothing masks it, put no secret in it)")
	cmd.Flags().StringVar(&reason, "reason", "", "why the card may go on (required; recorded as given: nothing masks it, put no secret in it)")
	_ = cmd.MarkFlagRequired("guard")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}

func newAFEvidenceCmd() *cobra.Command {
	var f afFlags
	var ev ledger.Evidence
	var repo string
	cmd := &cobra.Command{
		Use:   "evidence <card>",
		Short: "Record evidence (red, green, gate, review, merge) bound to a commit",
		Long: `Record evidence (red, green, gate, review, merge) bound to a commit.

The commit is looked up in --repo, else in the worktree of the card's lease, else in the project directory; it must name a
complete commit SHA that exists there, and that repository must be the one the ledger is bound to.

A failing green, gate or review run about the commit of a queue row that is in landing is refused (non-zero exit, nothing
recorded): the trunk is being moved on that commit, and the landing's last look at the newest runs has been taken or is
being taken. Record the run again once "moai af land" has returned; where no lander is running (a process that is gone
left the row in landing), "moai af reconcile" settles the row first. Passing runs, runs about another commit and red or
merge runs are not refused.

` + afBindHelp,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := afCheckRepoFlag(cmd.Context(), repo); err != nil {
				return err
			}
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				if _, err := afCheckCommit(ctx, st, f, args[0], repo, ev.SubjectSHA, "--subject-sha"); err != nil {
					return err
				}
				if err := st.AddEvidence(ctx, args[0], f.epoch, ev); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "kind": ev.Kind, "subject_sha": ev.SubjectSHA, "result": ev.Result})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&ev.Kind, "kind", "", "red, green, gate, review or merge")
	cmd.Flags().StringVar(&ev.SubjectSHA, "subject-sha", "", "the complete lowercase hex SHA (40 or 64 characters) of the commit that was measured; it must exist in the repository")
	cmd.Flags().StringVar(&repo, "repo", "", afRepoFlagHelp)
	cmd.Flags().StringVar(&ev.TreeSHA, "tree-sha", "", "the tree SHA (informational)")
	cmd.Flags().StringVar(&ev.BaseSHA, "base-sha", "", "the base commit SHA")
	cmd.Flags().StringVar(&ev.Command, "command", "", "the command that was run (recorded as given: nothing masks it, put no secret in it)")
	cmd.Flags().StringVar(&ev.Result, "result", "", "pass or fail")
	cmd.Flags().StringVar(&ev.Path, "path", "", "a file holding the output (the path is recorded as given: nothing masks it)")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.MarkFlagRequired("subject-sha")
	_ = cmd.MarkFlagRequired("result")
	return cmd
}

func newAFEnqueueCmd() *cobra.Command {
	var f afFlags
	var candidate, repo string
	cmd := &cobra.Command{
		Use:   "enqueue <card>",
		Short: "Put the card's candidate commit on the merge queue",
		Long: `Put the card's candidate commit on the merge queue.

The commit is looked up as evidence does (--repo, else the card's lease worktree, else the project directory). The verb
then asks that repository whether the candidate is in ` + mergeq.DefaultTrunk + ` and tells the ledger. A queue row that landed
under an earlier holder of the card, for this same commit, is taken over only when the commit is in the trunk; otherwise
the row starts again as waiting, to be landed for real. When the repository cannot answer (the trunk ref does not exist,
say) the offer is refused and nothing is changed: the verb does not guess.

` + afBindHelp,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := afCheckRepoFlag(cmd.Context(), repo); err != nil {
				return err
			}
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				used, err := afCheckCommit(ctx, st, f, args[0], repo, candidate, "--candidate")
				if err != nil {
					return err
				}
				// A landed row of an earlier epoch is taken over only when its commit really is in the dedicated trunk.
				onTrunk, err := afInTrunk(ctx, used, candidate)
				if err != nil {
					return err
				}
				if err := st.Enqueue(ctx, args[0], f.epoch, candidate, ledger.OnTrunk(onTrunk)); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "candidate": candidate, "state": ledger.QueueWaiting})
			})
		},
	}
	f.bind(cmd, true)
	cmd.Flags().StringVar(&candidate, "candidate", "", "the complete lowercase hex SHA (40 or 64 characters) of the commit the gates passed for; it must exist in the repository")
	cmd.Flags().StringVar(&repo, "repo", "", afRepoFlagHelp)
	_ = cmd.MarkFlagRequired("candidate")
	return cmd
}

func newAFReleaseCmd() *cobra.Command {
	var f afFlags
	cmd := &cobra.Command{
		Use:   "release <card>",
		Short: "Let go of a claimed card",
		Long:  "Let go of a claimed card.\n\n" + afBindHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				if err := st.Release(ctx, args[0], f.epoch); err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"card_id": args[0], "released": true})
			})
		},
	}
	f.bind(cmd, true)
	return cmd
}

// afAlive judges a lease holder by the identity the claim recorded (pid plus start fingerprint). Only a
// positively dead process is Dead; a missing pid or fingerprint, or a probe that cannot decide, is Unknown,
// which the ledger never reclaims (I5). It is the flow read model's adapter over homestate's classification,
// the one rule `claim` presence also uses.
func afAlive(pid int, start string) ledger.Liveness { return flow.ProcessProbe(afProbe)(pid, start) }

// newAFReapCmd is `moai af reap`: it takes back the cards whose holder process is gone (the ledger's Reap, judged
// by afAlive). A live holder keeps its card; the third reclaim of a card parks it for the operator.
func newAFReapCmd() *cobra.Command {
	var f afFlags
	cmd := &cobra.Command{
		Use:   "reap",
		Short: "Take back the cards whose holder process is gone",
		Long:  "Take back the cards whose holder process is gone.\n\n" + afBindHelp,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return afRun(f, func(ctx context.Context, st *ledger.Store) error {
				got, err := st.Reap(ctx, afAlive)
				if err != nil {
					return err
				}
				if got == nil {
					got = []ledger.Reclaimed{}
				}
				leases, err := st.Leases(ctx)
				if err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{"reclaimed": got, "unknown_expired": afUnknownExpired(leases, time.Now())})
			})
		},
	}
	f.bind(cmd, false)
	return cmd
}

// afUnknownLease is a lease whose holder is unknown and whose time has run out: the ledger will not reclaim it by
// liveness, so a lane that really died while holding it stays there until the operator looks.
type afUnknownLease struct {
	CardID    string    `json:"card_id"`
	Epoch     int64     `json:"epoch"`
	ExpiresAt time.Time `json:"expires_at"`
}

// afUnknownExpired lists the expired leases that carry no usable holder identity (no pid or no fingerprint).
func afUnknownExpired(leases []ledger.Lease, now time.Time) []afUnknownLease {
	out := []afUnknownLease{}
	for _, l := range leases {
		if (l.PID < 1 || l.PIDStart == "") && !l.ExpiresAt.After(now) {
			out = append(out, afUnknownLease{CardID: l.CardID, Epoch: l.Epoch, ExpiresAt: l.ExpiresAt})
		}
	}
	return out
}
