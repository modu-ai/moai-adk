package worktree

// migrate.go — SPEC-WORKTREE-HOME-ROOT-001 M6: `moai worktree migrate`
// (REQ-WHR-012/015/017/018; decisions D-6/D-7; design.md D-12).
//
// The engine reuses the safety planning of planLegacyWorktreeMigrationFor
// behaviorally (the same skip causes: prunable, locked, process cwd,
// active-process cwd, unreadable registry, live anchored session, missing
// source, existing destination) but selects from the M1 classifier: every
// moai-made tree under the legacy .moai/worktrees root is eligible
// (including a non-WT- branch); agent-*, claude-native, codex-managed,
// unclassified, integration and primary trees are never selected. The
// destination is the resolver's home projects root.
//
// Per tree the engine writes a manifest (HEAD + working-state digest), an
// intent line whose instance token is created first, then runs
// `git worktree move`, the state-file re-key, the card-record rebind and a
// manifest comparison, journaling done/pending per step. On any failure the
// reverse order runs — move back, re-key back, rebind back — and every
// reverse move first compares the tree's identity (TreeIdent, M1) with the
// journal's, refusing a mismatch with a `refused` line and the reason
// `identity mismatch`. The move list carries one appended line per settled
// step and is never rewritten or truncated.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Move legacy moai-made worktrees to the home projects root",
		Long: `Plan and (with --apply) move legacy moai-made worktrees to the
home projects root (<MoaiHome>/worktrees/projects/<project-key>/<name>),
with a manifest, an intent line, a journaled move, a state re-key, a card
rebind and a manifest comparison per tree; the agent-* and unclassified
trees are never selected. --verify reports each recorded move as intact;
--rollback reverses the journal's moves.`,
		Args: cobra.NoArgs,
		RunE: runMigrate,
	}
	cmd.Flags().Bool("apply", false, "move the eligible trees")
	cmd.Flags().Bool("verify", false, "verify each recorded move as intact")
	cmd.Flags().String("rollback", "", "reverse the moves recorded in the journal file")
	cmd.Flags().Bool("json", false, "emit the report as a JSON array")
	return cmd
}

func runMigrate(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve the working directory: %w", err)
	}
	// The engine's durable state (lock, move list, journal, factory store)
	// belongs to the PRIMARY checkout: invoked from inside a linked worktree
	// (card t1640 review P2a), cwd names the linked tree and every
	// <root>/.moai path below must resolve against the primary root instead.
	root, err := primaryRootFrom(cwd)
	if err != nil {
		return err
	}
	apply, _ := cmd.Flags().GetBool("apply")
	verify, _ := cmd.Flags().GetBool("verify")
	rollbackPath, _ := cmd.Flags().GetString("rollback")
	asJSON, _ := cmd.Flags().GetBool("json")

	switch {
	case verify:
		return runMigrateVerify(cmd, root, asJSON)
	case rollbackPath != "":
		return runMigrateRollback(cmd, root, rollbackPath, asJSON)
	default:
		return runMigratePlanOrApply(cmd, root, apply, asJSON)
	}
}

// primaryRootFrom resolves the primary checkout's project root from any
// directory of the repository — the parent of the git common dir (the same
// derivation buildListRows and worktreeroot's resolver use).
func primaryRootFrom(dir string) (string, error) {
	out, err := listGitOutput(dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("resolve the repository of %s: %w", dir, err)
	}
	common := strings.TrimSpace(out)
	if common == "" {
		return "", fmt.Errorf("resolve the repository of %s: git printed no common dir", dir)
	}
	return filepath.Dir(common), nil
}

// migrationLockPath is the OS-held advisory lock of the migration run
// (design.md D-12): one non-blocking TryLock, taken once by each entry
// point before the first step, never re-taken by the engine, released with
// the holder's process.
func migrationLockPath(root string) string {
	return filepath.Join(root, ".moai", "state", "worktree-migrate", "lock")
}

// withMigrationLock runs fn while holding the migration lock; a mover that
// finds it held ends its run without moving anything (REQ-WHR-017).
func withMigrationLock(root string, fn func() error) error {
	lockPath := migrationLockPath(root)
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return fmt.Errorf("migration lock: %w", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("migration lock: open %s: %w", lockPath, err)
	}
	defer func() { _ = f.Close() }()
	if terr := tryMigrationLock(f); terr != nil {
		return fmt.Errorf("migration lock: %s is held by another mover; this run moves nothing", lockPath)
	}
	return fn()
}

// runMigratePlanOrApply implements the dry run and the apply. The dry run
// takes no lock and writes nothing (AC-WHR-015: no path under the repository
// or MOAI_HOME is created, modified or removed) — the lock is an apply-time
// writer gate (AC-WHR-017b), not a reader's.
func runMigratePlanOrApply(cmd *cobra.Command, root string, apply, asJSON bool) error {
	run := func() error {
		var progress *treeProgress
		if apply && !asJSON {
			progress = newTreeProgress(cmd.ErrOrStderr())
		}
		report, err := runMigrationEngineWith(root, apply, "migrate", progress)
		// The rows are printed BEFORE any error (card t2053, C8): a run that
		// handled trees and then failed must still say what happened to them.
		if err == nil || len(report.rows) > 0 {
			if rerr := printMigrationReport(cmd, report, asJSON); rerr != nil {
				return rerr
			}
		}
		if err != nil {
			return err
		}
		if report.reversed > 0 || report.failed > 0 || report.interrupted > 0 {
			if report.interrupted > 0 {
				return fmt.Errorf("migration completed with %d reversed, %d failed and %d interrupted trees", report.reversed, report.failed, report.interrupted)
			}
			return fmt.Errorf("migration completed with %d reversed and %d failed trees", report.reversed, report.failed)
		}
		return nil
	}
	if !apply {
		return run()
	}
	return withMigrationLock(root, run)
}

// printMigrationReport renders the verb's report: the JSON rows, or the text
// rows, the counts and — when trees failed or were reversed — the journal.
func printMigrationReport(cmd *cobra.Command, report *migrationReport, asJSON bool) error {
	if asJSON {
		body, jerr := json.MarshalIndent(report.rows, "", "  ")
		if jerr != nil {
			return jerr
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}
	for _, r := range report.rows {
		line := fmt.Sprintf("%s\t%s → %s\t[%s]", r.Class, r.Source, r.Destination, r.Outcome)
		if r.SkipReason != "" {
			line += " (" + r.SkipReason + ")"
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)
	}
	summary := fmt.Sprintf("moved %d, skipped %d, failed %d, reversed %d",
		report.moved, report.skipped, report.failed, report.reversed)
	if report.interrupted > 0 {
		summary += fmt.Sprintf(", interrupted %d", report.interrupted)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), summary)
	if report.journal != "" && report.failed+report.reversed > 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "journal: %s (undo with: moai worktree migrate --rollback %s)\n", report.journal, report.journal)
	}
	return nil
}

// migrationRow is one line of the migration report. The fields are exported
// with JSON tags: --json marshals the rows and unexported fields would emit
// empty objects (card t1640 review P2g).
type migrationRow struct {
	Class       string `json:"class"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Outcome     string `json:"outcome"`
	SkipReason  string `json:"skip_reason,omitempty"`
	// Journal names the journal that records this row's tree when the row is a
	// failure, a reversal or an interrupted migration: the file --rollback
	// reads (card t2053, C7).
	Journal string `json:"journal,omitempty"`
}

// migrationReport carries the run's outcome counts.
type migrationReport struct {
	rows     []migrationRow
	moved    int
	skipped  int
	failed   int
	reversed int
	// interrupted counts trees a previous run left unfinished (card t2053,
	// C7); the rerun neither moves nor reports them as clean.
	interrupted int
	// journal is this run's journal path ("" when the run wrote none).
	journal string
}

// migratePlan is one tree of the plan: the classification drives the
// selection (moai-made moves, everything else never does).
type migratePlan struct {
	source, destination string
	head, branch        string
	class               string
	skip                string
	// interrupted marks a tree whose last move-list line is an unsettled
	// `pending` (card t2053, C7); journal names the journal recording it.
	interrupted bool
	journal     string
}

// ErrCommandTimeout marks a subprocess the migration stopped waiting for (card
// t2053). It fails SAFE everywhere it surfaces: a tree whose manifest or move
// timed out is skipped or failed with this reason, never reported moved.
var ErrCommandTimeout = errors.New("timed out")

// The bounds are package variables so a test can shorten them; the production
// values are the named defaults in internal/config.
var (
	lsofCommandTimeout = config.DefaultWorktreeLsofTimeout
	gitCommandTimeout  = config.DefaultWorktreeGitTimeout
	gitMoveTimeout     = config.DefaultWorktreeMoveTimeout
)

// boundedOutput runs name with args under a deadline and returns its stdout.
// dir, when set, becomes the working directory. A deadline hit returns an
// error wrapping ErrCommandTimeout; the child is killed and its output pipes
// are released after config.DefaultWorktreeCommandWaitDelay.
func boundedOutput(timeout time.Duration, dir, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.WaitDelay = config.DefaultWorktreeCommandWaitDelay
	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("%w after %s", ErrCommandTimeout, timeout)
	}
	return out, err
}

// activeProcessCWDProbe is the injectable seam for the process working-
// directory skip cause. The default is the conservative lsof scan — the
// same logic planLegacyWorktreeMigrationFor uses through its package-cli
// seam (the two must agree; only the package boundary differs).
var activeProcessCWDProbe = defaultActiveProcessCWDs

// defaultActiveProcessCWDs scans lsof for the working directories of every
// active process (the conservative process-level guard for worktree moves).
func defaultActiveProcessCWDs() ([]string, error) {
	out, err := boundedOutput(lsofCommandTimeout, "", "lsof", "-a", "-d", "cwd", "-F0n")
	if err != nil {
		return nil, fmt.Errorf("lsof cwd scan: %w", err)
	}
	var directories []string
	for _, field := range bytes.Split(out, []byte{0}) {
		value := strings.TrimLeft(string(field), "\r\n")
		if strings.HasPrefix(value, "n/") {
			directories = append(directories, strings.TrimSuffix(value[1:], " (deleted)"))
		}
	}
	return directories, nil
}

// destinationResolver is the seam for a tree's home-root landing path: the
// writing resolver for an apply, the read-only planner for a dry run.
var destinationResolver = func(root, name string, apply bool) (string, error) {
	if apply {
		return worktreeroot.PathFor(root, name)
	}
	return worktreeroot.Plan(root, name)
}

// processCWDIndex answers "does an active process work inside this tree" for
// many trees against one lsof snapshot (card t2070).
//
// The direct comparison (processUsesWorktree) resolved symlinks and walked the
// ancestors of every process directory again for every tree: measured at about
// 200ms per tree against 1,513 process directories, which is the two minutes a
// project with 648 trees spent in the plan before anything was printed. The
// index does that work once per process directory and once per tree, and
// answers the same question: a directory is "inside" the tree when its resolved
// path sits below the tree's, when an ancestor of it is the very same
// directory as the tree (spelling-safe, as underPath is), or when both resolve
// to the same path.
type processCWDIndex struct{ entries []processCWDEntry }

type processCWDEntry struct {
	canon     string        // canonicalTreePath of the directory
	resolved  string        // evalOrClean of the directory ("" when the directory is "")
	ancestors []os.FileInfo // the ancestors underPath would stat, nearest first
}

func newProcessCWDIndex(dirs []string) *processCWDIndex {
	seen := make(map[string]bool, len(dirs))
	x := &processCWDIndex{}
	for _, d := range dirs {
		if seen[d] {
			continue
		}
		seen[d] = true
		e := processCWDEntry{canon: canonicalTreePath(d)}
		if d != "" {
			e.resolved = evalOrClean(d)
			// The ancestor walk of underPath, recorded once: stop at the
			// filesystem root or at the first ancestor that does not exist;
			// any other stat error skips that ancestor.
			rp := e.resolved
			for {
				parent := filepath.Dir(rp)
				if parent == rp {
					break
				}
				rp = parent
				fi, err := os.Stat(rp)
				if err != nil {
					if os.IsNotExist(err) {
						break
					}
					continue
				}
				e.ancestors = append(e.ancestors, fi)
			}
		}
		x.entries = append(x.entries, e)
	}
	return x
}

func (x *processCWDIndex) usesWorktree(tree string) bool {
	treeCanon := canonicalTreePath(tree)
	var rr string
	var stRoot os.FileInfo
	if tree != "" {
		rr = evalOrClean(tree)
		if fi, err := os.Stat(rr); err == nil {
			stRoot = fi
		}
	}
	for _, e := range x.entries {
		if e.canon == treeCanon {
			return true
		}
		if tree == "" || e.resolved == "" || e.resolved == rr {
			continue
		}
		if strings.HasPrefix(e.resolved, rr+string(filepath.Separator)) {
			return true
		}
		if stRoot == nil {
			continue
		}
		for _, a := range e.ancestors {
			if os.SameFile(a, stRoot) {
				return true
			}
		}
	}
	return false
}

// processUsesWorktree reports whether one of the active processes works
// inside the tree at path.
func processUsesWorktree(directories []string, tree string) bool {
	for _, d := range directories {
		if isUnderWorktreePrefix(d, tree) || canonicalTreePath(d) == canonicalTreePath(tree) {
			return true
		}
	}
	return false
}

// worktreeRegistryProbe is the injectable seam for the session-registry
// readability skip cause. The default reads the registry the way the
// session package does (card t1640 review F8: a registry that cannot be
// read is a skip cause — fail-closed, I-1 — never a silently-voided probe;
// an absent registry file reads as empty, the session package's own
// semantics, so repositories without a registry are unaffected).
var worktreeRegistryProbe = func(root, tree string) error {
	for _, base := range []string{root, tree} {
		registry := session.NewRegistry(filepath.Join(base, session.DefaultRegistryPath), nil)
		if _, err := registry.Query(""); err != nil {
			return err
		}
	}
	return nil
}

// liveAnchoredProbe is the injectable seam for the live-anchored-session
// skip cause; the default reads the session registry.
var liveAnchoredProbe = func(tree, root string, now time.Time) int {
	return len(session.LiveAnchoredSessionsForProject(tree, root, now))
}

// planMigrationFor plans the migration over the registered worktrees of
// root with the same skip causes as planLegacyWorktreeMigrationFor, but the
// selection and the destination come from the M1 classifier and the
// resolver's home projects root (SPEC-WORKTREE-HOME-ROOT-001 M6). The dry
// run (apply unset) never writes: destinations come from the resolver's
// read-only plan — no project folder, owner marker or repository-side
// record is created for a planned tree (AC-WHR-015). The apply resolves
// through the writing resolver, whose first-time assignment claims the key
// and writes the marker (REQ-WHR-002a) — an apply may write.
//
// A worktree listing that fails (git exits non-zero, or the call times out —
// card t2053) or does not parse is an ERROR, not an empty plan: returning no
// plans would let the caller print a clean "0 moved, 0 skipped, 0 failed" for a
// run that never saw a candidate.
func planMigrationFor(root string, now time.Time, apply bool) ([]migratePlan, error) {
	return planMigrationWith(root, now, apply, nil)
}

// inMigrationScope reports whether a registered tree sits under one of the
// legacy worktree folders the migration looks at.
func inMigrationScope(rec WorktreeRecord) bool {
	slashed := filepath.ToSlash(rec.Path)
	return strings.Contains(slashed, "/.moai/worktrees/") || strings.Contains(slashed, "/.claude/worktrees/")
}

// planMigrationWith is planMigrationFor with the plan's live announcements: the
// count of trees it is about to check and that this can take a long time, then
// a throttled "checking (i/n)" line (card t2070). A nil progress is silent.
func planMigrationWith(root string, now time.Time, apply bool, progress *treeProgress) ([]migratePlan, error) {
	out, err := listGitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list the registered worktrees: %w", err)
	}
	records, err := ParseWorktreeList(out)
	if err != nil {
		return nil, fmt.Errorf("parse the registered worktree list: %w", err)
	}
	homeBase := homeProjectsDir()
	// The classifier verdicts must agree across list, migrate, tier and
	// retirement: the integration branch is set exactly as list.go sets it
	// (card t1640 review P2b).
	ctx := ClassifyContext{
		PrimaryRoot:       root,
		IntegrationBranch: config.LoadGitFlowIntegrationConfig(root).IntegrationTarget,
		HomeProjectsDir:   homeBase,
	}
	cwd, _ := os.Getwd()
	// The move list is read once, checked: a list that exists but cannot be read
	// is not an empty one, and no tree can be judged against it (card t2053,
	// sixth review).
	moveEntries, mlErr := loadMoveList(root)
	if mlErr != nil {
		return nil, mlErr
	}
	var plans []migratePlan
	var processCWDs []string
	var procIndex *processCWDIndex
	var processCWDErr error
	processCWDsRead := false

	// The home root is the same for every tree (the resolver's answer depends
	// on the repository, never on the tree name), so it is resolved once, at
	// the first tree that needs it, and each destination is that root joined
	// with the tree's name. Resolving per tree started a git subprocess per
	// tree: about 56ms each in the t2070 measurement.
	var destRoot string
	var destErr error
	destResolved := false
	destinationFor := func(name string) (string, error) {
		if !destResolved {
			destRoot, destErr = destinationResolver(root, "", apply)
			destResolved = true
		}
		if destErr != nil {
			return "", destErr
		}
		return filepath.Join(destRoot, name), nil
	}

	candidates := 0
	for _, rec := range records {
		if !rec.Bare && inMigrationScope(rec) {
			candidates++
		}
	}
	progress.planBegin(candidates)
	defer progress.planEnd()
	scanned := 0
	for _, rec := range records {
		if rec.Bare {
			continue
		}
		if !inMigrationScope(rec) {
			continue // primary and unrelated trees are not the migration's
		}
		scanned++
		progress.scan(scanned, candidates)
		cls := Classify(WorktreeRecord{Path: rec.Path, Branch: rec.Branch}, ctx)
		// The plan lists every tree with its classification; only moai-made
		// trees are selected for moving (the classifier's selection: agent-*,
		// unclassified etc. are never moved).
		p := migratePlan{
			source: rec.Path,
			head:   rec.Head,
			branch: rec.Branch,
			class:  string(cls.Class),
		}
		// SPEC-WORKTREE-HOME-ROOT-001 M2 (REQ-WHR-005a): the destination is
		// the resolver's home-root landing path for this tree.
		dest, derr := destinationFor(filepath.Base(rec.Path))
		if derr != nil {
			p.skip = fmt.Sprintf("resolve home-root destination: %v", derr)
		} else {
			p.destination = dest
		}
		// A tree the migration left unfinished is not a candidate and not a
		// stale registration: it is reported as interrupted, naming the journal
		// that recovers it (card t2053, C7) — never as something to prune. The
		// check does not depend on where the tree stands (L1): a run killed after
		// its pending lines but before the move leaves the source intact and the
		// destination absent, and that state is interrupted too — unless it is
		// PROVABLY a step that never moved, which an apply settles explicitly.
		if line, interrupted := findInterrupted(moveEntries, rec.Path, p.destination); interrupted {
			journal := journalRecording(root, line.Ident)
			stepJournal, unmoved := provablyUnmovedStep(root, rec.Path, p.destination, line)
			switch {
			case unmoved && !apply:
				// A candidate: the dry run writes nothing (AC-WHR-015).
			case unmoved:
				if serr := settleUnmovedStep(root, line, stepJournal); serr != nil {
					p.interrupted, p.journal = true, stepJournal
					p.skip = fmt.Sprintf("%s; the earlier step could not be settled automatically: %v", interruptedReason(root, stepJournal), serr)
					plans = append(plans, p)
					continue
				}
			default:
				p.interrupted, p.journal = true, journal
				p.skip = interruptedReason(root, journal)
				plans = append(plans, p)
				continue
			}
		}
		// A tree already standing at its destination has nothing to migrate and
		// earns no row (C16).
		if p.destination != "" && worktreeroot.Same(rec.Path, p.destination) {
			continue
		}
		// Only fs.ErrNotExist makes a path absent: a source whose path cannot be
		// resolved is skipped saying so (card t2053, sixth review).
		treeCanon, canonErr := resolveTreePath(rec.Path)
		// The same skip causes as planLegacyWorktreeMigrationFor.
		switch {
		case rec.Prunable:
			p.skip = "Git lists this tree as prunable"
		case rec.Locked:
			p.skip = "Git worktree is locked"
		case canonErr != nil:
			p.skip = fmt.Sprintf("cannot resolve the source path: %v", canonErr)
		case isUnderWorktreePrefix(cwd, rec.Path) || canonicalTreePath(cwd) == treeCanon:
			p.skip = "current process is inside this worktree"
		case !processCWDsRead:
			processCWDs, processCWDErr = activeProcessCWDProbe()
			processCWDsRead = true
			procIndex = newProcessCWDIndex(processCWDs)
			fallthrough
		default:
			if processCWDErr != nil {
				p.skip = fmt.Sprintf("cannot inspect active process working directories: %v", processCWDErr)
			} else if procIndex.usesWorktree(rec.Path) {
				p.skip = "an active process is inside this worktree"
			} else if worktreeRegistryProbe(root, rec.Path) != nil {
				p.skip = "session registry cannot be read"
			} else if liveAnchoredProbe(rec.Path, root, now) > 0 {
				p.skip = "active session is anchored in this worktree"
			} else if info, statErr := os.Lstat(rec.Path); statErr != nil {
				if errors.Is(statErr, os.ErrNotExist) {
					p.skip = "source path is missing"
				} else {
					p.skip = fmt.Sprintf("source path cannot be inspected: %v", statErr)
				}
			} else if !info.IsDir() {
				p.skip = "source path is not a directory"
			} else if _, statErr := os.Lstat(p.destination); statErr == nil {
				p.skip = "destination already exists"
			} else if !errors.Is(statErr, os.ErrNotExist) {
				p.skip = "destination cannot be inspected"
			}
		}
		plans = append(plans, p)
	}
	return plans, nil
}

// findInterrupted returns the latest unsettled `pending` location line of a
// tree touching any of paths: a migration that started and was never settled,
// restored or reversed (card t2053, C7).
func findInterrupted(entries []mlRecord, paths ...string) (mlRecord, bool) {
	latest := map[string]mlRecord{}
	for _, rec := range entries {
		if rec.Ident != "" && isLocationLine(rec.Result) {
			latest[rec.Ident] = rec
		}
	}
	for _, rec := range latest {
		if rec.Result != "pending" {
			continue
		}
		for _, p := range paths {
			if p != "" && (worktreeroot.Same(rec.OldPath, p) || worktreeroot.Same(rec.NewPath, p)) {
				return rec, true
			}
		}
	}
	return mlRecord{}, false
}

// reasonInterruptedBeforeMove is the reason an earlier step is settled with when
// it is proven never to have moved its tree; a rollback treats such a step as
// nothing to undo.
const reasonInterruptedBeforeMove = "interrupted before the move"

// provablyUnmovedStep reports whether the pending forward-intent line is a step
// that provably never moved its tree, and the journal holding that step. Every
// condition is a positive observation: the line is the forward orientation for
// this source and destination, the source carries the line's identity, git
// registers the source, the destination is absent (fs.ErrNotExist — not merely
// unreadable), and a readable journal holds the step. Anything else is not
// proof (L1).
func provablyUnmovedStep(root, source, destination string, line mlRecord) (string, bool) {
	if line.Result != "pending" || line.Ident == "" || destination == "" ||
		!worktreeroot.Same(line.OldPath, source) || !worktreeroot.Same(line.NewPath, destination) {
		return "", false
	}
	id, state, _ := probeTree(source)
	if state != treeIdentified || id.Ident() != line.Ident || !registeredAt(root, source) {
		return "", false
	}
	if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
		return "", false
	}
	journals, err := orderedJournals(filepath.Dir(moveListPath(root)))
	if err != nil {
		return "", false
	}
	for i := len(journals) - 1; i >= 0; i-- {
		steps, _, rerr := readJournalFile(journals[i])
		if rerr != nil {
			continue
		}
		for _, s := range steps {
			if s.Ident == line.Ident && worktreeroot.Same(s.OldPath, line.OldPath) &&
				(s.Outcome == "pending" || (s.Outcome == "failed" && s.Reason == reasonInterruptedBeforeMove)) {
				return journals[i], true
			}
		}
		for _, s := range steps {
			if s.Ident == line.Ident {
				return "", false // the newest journal naming this tree does not hold an unsettled step
			}
		}
	}
	return "", false
}

// readJournalFile parses a journal file: its steps and its raw top-level
// fields (kept so a rewrite preserves trigger and creation time).
func readJournalFile(path string) ([]journalStep, map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, nil, err
	}
	var steps []journalStep
	if raw, ok := fields["steps"]; ok {
		if err := json.Unmarshal(raw, &steps); err != nil {
			return nil, nil, err
		}
	}
	return steps, fields, nil
}

// settleUnmovedStep settles an earlier step proven never to have moved its tree:
// the journal step becomes `failed` ("interrupted before the move") through the
// atomic journal write, and then a releasing `reversed` line ends the move-list
// protection. The journal goes first, so a kill between the two leaves a step the
// next run recognises and finishes settling (the step still passes
// provablyUnmovedStep), never one that looks unsettled forever.
func settleUnmovedStep(root string, line mlRecord, journalPath string) error {
	steps, fields, err := readJournalFile(journalPath)
	if err != nil {
		return fmt.Errorf("read the journal %s: %w", journalPath, err)
	}
	for i := range steps {
		if steps[i].Ident == line.Ident && worktreeroot.Same(steps[i].OldPath, line.OldPath) && steps[i].Outcome == "pending" {
			steps[i].Outcome, steps[i].Reason = "failed", reasonInterruptedBeforeMove
		}
	}
	raw, merr := json.Marshal(steps)
	if merr != nil {
		return merr
	}
	fields["steps"] = raw
	data, merr := json.MarshalIndent(fields, "", "  ")
	if merr != nil {
		return merr
	}
	if werr := journalWrite(journalPath, data); werr != nil {
		return fmt.Errorf("settle the journal %s: %w", journalPath, werr)
	}
	return appendMoveListLine(moveListPath(root), m6MoveLineReason("reversed", line.NewPath, line.OldPath, line.Ident, reasonInterruptedBeforeMove))
}

// resolveTreePath is the canonical spelling of p for comparisons: a path that is
// absent (fs.ErrNotExist) keeps its cleaned spelling, any other resolution
// failure is an error — never silently the unresolved spelling.
func resolveTreePath(p string) (string, error) {
	r, err := filepath.EvalSymlinks(p)
	if err == nil {
		return r, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return filepath.Clean(p), nil
	}
	return "", err
}

// journalRecording is the newest journal under the migration state directory
// holding a step for ident ("" when none does).
func journalRecording(root, ident string) string {
	journals, err := orderedJournals(filepath.Dir(moveListPath(root)))
	if err != nil {
		return ""
	}
	for i := len(journals) - 1; i >= 0; i-- {
		data, rerr := os.ReadFile(journals[i])
		if rerr != nil {
			continue
		}
		var journal struct {
			Steps []journalStep `json:"steps"`
		}
		if json.Unmarshal(data, &journal) != nil {
			continue
		}
		for _, s := range journal.Steps {
			if s.Ident == ident {
				return journals[i]
			}
		}
	}
	return ""
}

// interruptedReason is the row text of an interrupted migration: the way out is
// the journal's rollback, never `git worktree prune` (which would drop the
// registration of a tree that still exists at the destination).
func interruptedReason(root, journal string) string {
	if journal == "" {
		journal = filepath.Join(filepath.Dir(moveListPath(root)), "<journal>.json")
	}
	return fmt.Sprintf("interrupted migration — run `moai worktree migrate --rollback %s`", journal)
}

// migrationReport builders

// journalStep is one tree's entry of the migration journal: pending from
// the moment it is written (BEFORE the tree's move, REQ-WHR-017 / card
// t1640 review F3), settled with the step's verdict after it. The manifest
// digest rides the settled entry so a rollback can apply the
// settled-moves-only manifest precondition (REQ-WHR-018's fifth entry).
type journalStep struct {
	OldPath  string `json:"old_path"`
	NewPath  string `json:"new_path"`
	Outcome  string `json:"outcome"`
	Ident    string `json:"ident"`
	Reason   string `json:"reason,omitempty"`
	Manifest string `json:"manifest,omitempty"`
	// ManifestVersion names the digest algorithm behind Manifest. Absent (0)
	// means the pre-t2053 algorithm, which hashed the bytes of ignored files
	// too; a rollback verifies such a journal that way (captureLegacyManifest).
	ManifestVersion int `json:"manifest_version,omitempty"`
}

// manifestFormatVersion is the digest algorithm this build writes: 2 = ignored
// files digested by path, type and size only (card t2053).
const manifestFormatVersion = 2

// migrationJournal is one apply run's journal file: every tree is appended
// as pending before its move and settled afterwards, and the whole file is
// rewritten on every change — a kill mid-run leaves the journal naming
// every tree whose move may have started (card t1640 review F3).
type migrationJournal struct {
	path    string
	dir     string // when path is empty: the directory the file is claimed in on first flush
	stamp   string // the second-resolution stem of the file name
	trigger string
	steps   []journalStep
}

// claimJournalPath creates dir/<stamp>.json exclusively and returns its path;
// a name already taken (another run in the same second, or a leftover) gets a
// -N suffix, so a run never overwrites an earlier run's journal (card t2053,
// C12). The claimed stub is replaced by the first atomic flush.
func claimJournalPath(dir, stamp string) (string, error) {
	for i := 0; i < 1000; i++ {
		name := stamp + ".json"
		if i > 0 {
			name = fmt.Sprintf("%s-%d.json", stamp, i)
		}
		path := filepath.Join(dir, name)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_ = f.Close()
			return path, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("no free journal name under %s for stamp %s", dir, stamp)
}

func (j *migrationJournal) flush() error {
	if j == nil {
		return nil
	}
	if j.path == "" && j.dir != "" {
		path, err := claimJournalPath(j.dir, j.stamp)
		if err != nil {
			return fmt.Errorf("journal: %w", err)
		}
		j.path = path
	}
	if j.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(map[string]any{
		"steps": j.steps, "trigger": j.trigger,
		"created": time.Now().UTC().Format(time.RFC3339),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("journal: %w", err)
	}
	if err := journalWrite(j.path, data); err != nil {
		return fmt.Errorf("journal: %w", err)
	}
	return nil
}

// pending records a tree whose move is about to start.
func (j *migrationJournal) pending(oldPath, newPath, ident string) error {
	if j == nil {
		return nil
	}
	j.steps = append(j.steps, journalStep{
		OldPath: oldPath, NewPath: newPath, Outcome: "pending", Ident: ident,
	})
	return j.flush()
}

// settle records the step's verdict for the tree's pending entry.
func (j *migrationJournal) settle(ident, outcome, reason, manifest string) error {
	if j == nil {
		return nil
	}
	for i := range j.steps {
		if j.steps[i].Ident == ident && j.steps[i].Outcome == "pending" {
			j.steps[i].Outcome = outcome
			j.steps[i].Reason = reason
			j.steps[i].Manifest = manifest
			if manifest != "" {
				j.steps[i].ManifestVersion = manifestFormatVersion
			}
			break
		}
	}
	return j.flush()
}

// runMigrationEngine plans and (when apply is set) executes the migration.
// The dry run and the apply both list every registered tree with its class;
// only moai-made trees are moved. Non-moai-made trees are reported as
// skipped with their class as the reason. The trigger names the caller in
// the move list ("migrate" for the verb, "update" for the consolidation —
// AC-WHR-019(g)).
func runMigrationEngine(root string, apply bool, trigger string) (*migrationReport, error) {
	return runMigrationEngineWith(root, apply, trigger, nil)
}

// runMigrationEngineWith is runMigrationEngine with a live-progress renderer:
// the apply announces the count of trees it will move, then names each tree as
// it starts. A nil progress is silent.
func runMigrationEngineWith(root string, apply bool, trigger string, progress *treeProgress) (*migrationReport, error) {
	report := &migrationReport{}
	var planProgress *treeProgress
	if apply {
		planProgress = progress // the dry run and --json stay silent, as before
	}
	plans, planErr := planMigrationWith(root, time.Now(), apply, planProgress)
	if planErr != nil {
		return report, planErr
	}
	if apply {
		eligible := 0
		for _, p := range plans {
			if p.class == string(ClassMoaiMade) && p.skip == "" {
				eligible++
			}
		}
		progress.begin(eligible)
		defer progress.end()
	}
	var prefetch *manifestPrefetcher
	if apply {
		var sources []string
		for _, p := range plans {
			if !p.interrupted && p.class == string(ClassMoaiMade) && p.skip == "" {
				sources = append(sources, p.source)
			}
		}
		prefetch = newManifestPrefetcher(sources, config.DefaultWorktreeMigrateManifestLookahead)
		defer prefetch.stop()
	}
	treeIndex := 0
	mlPath := moveListPath(root)

	var journal *migrationJournal
	if apply {
		journal = &migrationJournal{
			dir:     filepath.Dir(moveListPath(root)),
			stamp:   journalStamp(),
			trigger: trigger,
		}
	}

	var writeErrs []string
	for _, p := range plans {
		row := migrationRow{
			Class: p.class, Source: p.source, Destination: p.destination,
			Outcome: "planned", SkipReason: p.skip,
		}
		if p.interrupted {
			row.Outcome, row.Journal = "interrupted", p.journal
			report.interrupted++
			report.rows = append(report.rows, row)
			continue
		}
		if p.class != string(ClassMoaiMade) {
			row.Outcome = "skipped"
			row.SkipReason = "class " + p.class + " is never moved"
		} else if p.skip != "" {
			row.Outcome = "skipped"
		}
		if row.Outcome == "skipped" {
			report.skipped++
			report.rows = append(report.rows, row)
			continue
		}
		if !apply {
			report.rows = append(report.rows, row)
			continue
		}
		// The apply: the row's outcome is the engine's verdict for this tree
		// (card t1640 review P2h — the placeholder never survives the run).
		treeIndex++
		progress.tree(treeIndex, p.source)
		index := treeIndex - 1
		out := migrateOneTreeWith(root, p, mlPath, trigger, journal, func() (string, error) {
			return prefetch.get(index)
		})
		row.Outcome, row.SkipReason = out.Outcome, out.Reason
		if journal != nil && journal.path != "" {
			report.journal = journal.path
			if out.Outcome == "failed" || out.Outcome == "reversed" {
				row.Journal = journal.path
			}
		}
		if mlErr := out.MLErr; mlErr != nil {
			writeErrs = append(writeErrs, fmt.Sprintf("%s: %v", p.source, mlErr))
		}
		switch row.Outcome {
		case "moved":
			report.moved++
		case "reversed":
			report.reversed++
		case "failed":
			report.failed++
		default:
			report.skipped++
		}
		report.rows = append(report.rows, row)
		if jerr := journal.settle(out.Ident, out.Outcome, out.Reason, out.Manifest); jerr != nil {
			return report, jerr
		}
	}
	if len(writeErrs) > 0 {
		return report, fmt.Errorf("move-list writes failed: %s", strings.Join(writeErrs, "; "))
	}
	return report, nil
}

// treeOutcome is migrateOneTree's verdict for one tree.
type treeOutcome struct {
	Outcome  string // moved | reversed | failed | skipped
	Reason   string
	Ident    string
	Manifest string // the pre-move manifest digest (the rollback's settled precondition)
	MLErr    error  // a move-list write error that outlives a successful move
}

// migrateOneTree moves one tree through the journaled step order —
// manifest, journal-pending (F3: before the move), intent line (the
// instance token first), git worktree move, state re-key, card rebind,
// manifest comparison — and, on any failure, the one reverse order (move
// back, re-key back, rebind back). Every reverse move first compares the
// tree's identity with the journal's and refuses a mismatch with a
// `refused` line (REQ-WHR-017/018).
func migrateOneTree(root string, p migratePlan, mlPath, trigger string, journal *migrationJournal) treeOutcome {
	return migrateOneTreeWith(root, p, mlPath, trigger, journal, func() (string, error) {
		return manifestCapture(p.source)
	})
}

// manifestCapture is the seam for the before-the-move manifest of one tree.
var manifestCapture = captureManifestChecked

// manifestPrefetcher reads the before-the-move manifests of the trees waiting
// behind the one being moved (card t2070). The read is read-only and
// independent per tree, so it overlaps with the previous tree's move; the move,
// the journal write and the registration repair stay strictly one tree at a
// time, in plan order. At most `ahead` captures are outstanding — started and
// not yet taken by their tree — so the gap between a tree's capture and its move
// stays a few trees wide; a change made in that gap is caught by the comparison
// after the move, exactly as before.
type manifestPrefetcher struct {
	futures []*manifestFuture
	slots   chan struct{}
	quit    chan struct{}
	once    sync.Once
}

type manifestFuture struct {
	done   chan struct{}
	digest string
	err    error
}

func newManifestPrefetcher(trees []string, ahead int) *manifestPrefetcher {
	if ahead < 1 {
		ahead = 1
	}
	pf := &manifestPrefetcher{slots: make(chan struct{}, ahead), quit: make(chan struct{})}
	for range trees {
		pf.futures = append(pf.futures, &manifestFuture{done: make(chan struct{})})
	}
	go func() {
		for i, tree := range trees {
			select {
			case pf.slots <- struct{}{}:
			case <-pf.quit:
				return
			}
			go func(f *manifestFuture, tree string) {
				f.digest, f.err = manifestCapture(tree)
				close(f.done)
			}(pf.futures[i], tree)
		}
	}()
	return pf
}

// get returns the i-th tree's manifest, waiting for its capture, and frees the
// slot for the next tree's. Each index is taken once, in order.
func (pf *manifestPrefetcher) get(i int) (string, error) {
	f := pf.futures[i]
	<-f.done
	<-pf.slots
	return f.digest, f.err
}

// stop ends the read-ahead; captures already running finish on their own.
func (pf *manifestPrefetcher) stop() { pf.once.Do(func() { close(pf.quit) }) }

// migrateOneTreeWith is migrateOneTree with the before-the-move manifest handed
// in: the engine reads it ahead of time for the trees waiting behind the one
// being moved (card t2070).
func migrateOneTreeWith(root string, p migratePlan, mlPath, trigger string, journal *migrationJournal, before func() (string, error)) treeOutcome {
	digest, merr := before()
	if merr != nil {
		// Fail safe (card t2053): a manifest that could not be read in full —
		// a git call timed out — proves nothing, so nothing moves.
		return treeOutcome{Outcome: "skipped", Reason: fmt.Sprintf("manifest unavailable: %v; moved nothing", merr)}
	}

	id, err := EnsureInstanceToken(p.source)
	if err != nil {
		return treeOutcome{Outcome: "skipped", Reason: fmt.Sprintf("identity unavailable: %v", err)}
	}
	// The journal's pending entry precedes the move (REQ-WHR-017, review
	// F3): a kill mid-move leaves the journal naming the tree whose move
	// may have started, and --rollback converges from it. The move-list
	// intent line does the same for the lane pointer (REQ-WHR-012b step 9);
	// either write failing aborts the move before anything moves.
	if jerr := journal.pending(p.source, p.destination, id.Ident()); jerr != nil {
		return treeOutcome{Outcome: "failed", Reason: fmt.Sprintf("journal pending entry not recorded; moved nothing (%v)", jerr), Ident: id.Ident()}
	}
	if ierr := appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "pending", p.source, p.destination, id.Ident())); ierr != nil {
		return treeOutcome{Outcome: "failed", Reason: fmt.Sprintf("intent line not recorded; moved nothing (%v)", ierr), Ident: id.Ident(), Manifest: digest}
	}

	if err := forwardGitMove(root, p.source, p.destination); err != nil {
		// ANY forward error (card t2053, C4), not only a deadline: git can fail
		// after the directory rename, so the destination decides. A destination
		// that carries the journal identity means the tree moved (or half
		// moved) and goes through the one reverse order; otherwise nothing
		// moved and the intent line is released.
		reason := fmt.Sprintf("git worktree move: %v", err)
		switch forwardFailureLocation(p, id.Ident()) {
		case atDestination:
			return reverseAfterFailure(root, p, id, mlPath, trigger, digest, reason)
		case atSource:
			// The tree never left the source: the pending intent line must not
			// keep protecting it (C13). A `reversed` line is the release;
			// failing to write it surfaces through MLErr.
			mlErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "reversed", p.destination, p.source, id.Ident(),
				"the forward move failed before the tree moved"))
			return treeOutcome{Outcome: "failed", Reason: reason, Ident: id.Ident(), Manifest: digest, MLErr: mlErr}
		default:
			// The tree's location cannot be established (K2): an unreadable
			// identity is not an absent tree. The protection stays and the
			// pending line records both ends for --rollback to converge.
			const msg = "the tree's location could not be established after the failed move"
			mlErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(), msg))
			return treeOutcome{Outcome: "failed", Reason: fmt.Sprintf("%s; %s (pending)", reason, msg), Ident: id.Ident(), Manifest: digest, MLErr: mlErr}
		}
	}

	if rerr := hook.RekeyWorktreeEntry(root, p.source, p.destination); rerr != nil {
		if strings.Contains(rerr.Error(), "no entry keyed to") {
			// No state-file entry for the old path: the re-key is a no-op
			// (the tree was never registered in the state file).
		} else {
			return reverseAfterFailure(root, p, id, mlPath, trigger, digest, fmt.Sprintf("re-key failed: %v", rerr))
		}
	}

	if rerr := rebindCardForPath(root, p.source, p.destination); rerr != nil {
		return reverseAfterFailure(root, p, id, mlPath, trigger, digest, fmt.Sprintf("card rebind failed: %v", rerr))
	}

	after, aerr := captureManifestChecked(p.destination)
	if aerr != nil {
		return reverseAfterFailure(root, p, id, mlPath, trigger, digest, fmt.Sprintf("manifest unavailable after the move: %v", aerr))
	}
	if after != digest {
		return reverseAfterFailure(root, p, id, mlPath, trigger, digest, "manifest mismatch after the move")
	}

	if merr := appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "moved", p.source, p.destination, id.Ident())); merr != nil {
		// The move stands; the settling-line failure surfaces at the run's
		// exit (the intent line already names the tree, REQ-WHR-012b step 9).
		return treeOutcome{Outcome: "moved", Ident: id.Ident(), Manifest: digest, MLErr: merr}
	}
	return treeOutcome{Outcome: "moved", Ident: id.Ident(), Manifest: digest}
}

// Where a tree stands after a failed forward move.
const (
	atDestination   = "destination"
	atSource        = "source"
	locationUnknown = "unknown"
)

// forwardFailureLocation decides where the tree stands after `git worktree
// move` returned an error. The destination wins only when it carries the
// journal identity; the source wins only when it exists, carries the identity
// and the destination is absent or provably another tree. Everything else —
// an identity that cannot be read, a path that cannot be inspected, neither
// end carrying the identity — is unknown, never "absent" (card t2053, K2).
func forwardFailureLocation(p migratePlan, ident string) string {
	_, derr := os.Lstat(p.destination)
	if derr != nil && !errors.Is(derr, os.ErrNotExist) {
		return locationUnknown
	}
	if derr == nil {
		destIdent, ierr := TreeIdent(p.destination)
		if ierr != nil {
			return locationUnknown
		}
		if destIdent.Ident() == ident {
			return atDestination
		}
	}
	if srcIdent, serr := TreeIdent(p.source); serr == nil && srcIdent.Ident() == ident {
		return atSource
	}
	return locationUnknown
}

// reverseAfterFailure runs the one reverse order after a failed step and
// translates its result into the outcome (card t1640 review P1-4): the tree
// counts as reversed only when the reverse move SUCCEEDED — a failed
// reverse move leaves the tree standing at the destination with its re-key
// and rebind intact (they name the location the tree really sits at),
// records pending, and fails the tree.
func reverseAfterFailure(root string, p migratePlan, id TreeIdentity, mlPath, trigger, digest, reason string) treeOutcome {
	movedBack, why, mlErr := m6ReverseMoveWhy(root, p, id, mlPath, trigger)
	out := treeOutcome{Ident: id.Ident(), Manifest: digest}
	if movedBack {
		out.Outcome, out.Reason = "reversed", reason
	} else {
		out.Outcome = "failed"
		if mlErr != nil {
			out.Reason = fmt.Sprintf("%s; %s and the pending line could not be written (%v)", reason, why, mlErr)
		} else {
			out.Reason = fmt.Sprintf("%s; %s", reason, why)
		}
	}
	out.MLErr = mlErr
	return out
}

// forwardGitMove is the injectable seam for the forward `git worktree move`
// (the fault-injection point for a move killed after the directory rename but
// before git updated its registration — card t2053).
var forwardGitMove = func(root, from, to string) error {
	_, err := gitOutputWithin(gitMoveTimeout, root, "worktree", "move", from, to)
	return err
}

// reverseGitMove is the injectable seam for the reverse `git worktree move`
// (the fault-injection point of AC-WHR-017a's reverse-step failure clause).
var reverseGitMove = func(root, from, to string) error {
	_, err := gitOutputWithin(gitMoveTimeout, root, "worktree", "move", from, to)
	return err
}

// journalStamp names a run's journal file; a seam so a test can force two runs
// onto the same name.
var journalStamp = func() string { return time.Now().UTC().Format("20060102T150405") }

// journalReplace is the final rename step of the atomic journal write; a seam
// so a test can stop the write exactly there.
var journalReplace = atomicfile.Replace

// journalWrite is the injectable seam for the journal file write (the
// fault-injection point of the REQ-WHR-017 journal clause).
var journalWrite = func(path string, data []byte) error {
	return writeFileAtomic(path, data, 0o644)
}

// writeFileAtomic replaces path with data: a temp file in the same directory,
// fsynced, then renamed over the target (card t2053, C1). A kill at any point
// leaves either the previous complete file or the new complete one — never a
// truncated journal, which would be the recovery handle for every tree the run
// already moved. On error the temp file is removed.
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".journal-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }() // a no-op once the rename has succeeded
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(name, mode)
	}
	if err != nil {
		return err
	}
	return journalReplace(name, path)
}

// RunUpdateConsolidation is the exported update-time entry into the
// migration engine: internal/cli's update consolidation delegates here
// (card t1640 review F1 — the route AC-WHR-019 requires).
func RunUpdateConsolidation(root string, dryRun bool, out io.Writer) error {
	return runUpdateConsolidation(root, dryRun, out)
}

// runUpdateConsolidation is the update-time entry into the migration
// engine — the route AC-WHR-019 requires (card t1640 review F1): the same
// classifier selection (agent-*, unclassified, integration, codex-managed
// and primary trees are never moved), the same home-root destination from
// the resolver, the same migration lock, journal, state re-key, card
// rebind and manifest comparison as the `migrate` verb, with the move-list
// trigger `update`. The dry run writes nothing (AC-WHR-015/AC-WHR-019(h):
// the plan only). The update passes its own lock state through: a held
// migration lock fails the consolidation and the update degrades it to a
// warning (AC-WHR-019(j)).
func runUpdateConsolidation(root string, dryRun bool, out io.Writer) error {
	run := func() error {
		var progress *treeProgress
		if !dryRun {
			progress = newTreeProgress(out)
		}
		report, err := runMigrationEngineWith(root, !dryRun, "update", progress)
		// Rows first, error after (card t2053, C8).
		if err == nil || len(report.rows) > 0 {
			printUpdateReport(out, report, dryRun)
		}
		return err
	}
	if dryRun {
		return run()
	}
	return withMigrationLock(root, run)
}

// printUpdateReport renders the update consolidation's lines.
func printUpdateReport(out io.Writer, report *migrationReport, dryRun bool) {
	if dryRun {
		for _, r := range report.rows {
			switch r.Outcome {
			case "planned":
				_, _ = fmt.Fprintf(out, "Worktree migration planned %s -> %s\n", r.Source, r.Destination)
			case "skipped":
				_, _ = fmt.Fprintf(out, "Worktree migration skipped %s: %s\n", r.Source, r.SkipReason)
			case "interrupted":
				_, _ = fmt.Fprintf(out, "Worktree migration interrupted %s: %s\n", r.Source, r.SkipReason)
			}
		}
		return
	}
	for _, r := range report.rows {
		switch r.Outcome {
		case "skipped":
			_, _ = fmt.Fprintf(out, "Worktree migration skipped %s: %s\n", r.Source, r.SkipReason)
		case "interrupted":
			_, _ = fmt.Fprintf(out, "Worktree migration interrupted %s: %s\n", r.Source, r.SkipReason)
		case "failed", "reversed":
			// A reversed move is a tree that is NOT migrated: the
			// engine moved it back (AC-WHR-019(f)).
			_, _ = fmt.Fprintf(out, "Worktree migration failed %s: %s\n", r.Source, r.SkipReason)
		case "moved":
			_, _ = fmt.Fprintf(out, "Worktree migration moved %s -> %s\n", r.Source, r.Destination)
		}
	}
	summary := fmt.Sprintf("Worktree migration: %d moved, %d skipped, %d failed",
		report.moved, report.skipped, report.failed+report.reversed)
	if report.interrupted > 0 {
		summary += fmt.Sprintf(", %d interrupted", report.interrupted)
	}
	_, _ = fmt.Fprintln(out, summary)
	if report.journal != "" && report.failed+report.reversed > 0 {
		_, _ = fmt.Fprintf(out, "Worktree migration journal: %s (undo with: moai worktree migrate --rollback %s)\n", report.journal, report.journal)
	}
}

// m6ReverseMove runs the one reverse order — move back, re-key back, rebind
// back — after first comparing the tree's identity at the destination with
// the journal's ident and refusing a mismatch with a `refused` line
// (REQ-WHR-017/018, design.md D-12: only ident tells two trees apart). It
// reports whether the tree actually moved back (card t1640 review P1-4) and
// the move-list write error, if any.
func m6ReverseMove(root string, p migratePlan, id TreeIdentity, mlPath, trigger string) (bool, error) {
	movedBack, _, mlErr := m6ReverseMoveWhy(root, p, id, mlPath, trigger)
	return movedBack, mlErr
}

// m6ReverseMoveWhy is m6ReverseMove plus the reason a reverse that did not
// complete stopped (card t2053, C16): a refusal by the identity check must not
// read as a failed reverse move.
func m6ReverseMoveWhy(root string, p migratePlan, id TreeIdentity, mlPath, trigger string) (bool, string, error) {
	// The identity rule the rollback applies (G8): an identity that cannot be
	// read, or a journal ident that is empty, is not a match.
	destIdent, dstate, derr := probeTree(p.destination)
	if dstate == treeUnreadable {
		// Unknown is not a mismatch: the tree may be exactly ours (sixth review).
		msg := fmt.Sprintf("the destination's identity cannot be read: %v", derr)
		return false, "the reverse move did not run: " + msg + " (pending)",
			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(), msg))
	}
	if dstate != treeIdentified || id.Ident() == "" || destIdent.Ident() != id.Ident() {
		return false, "refused: identity mismatch at the destination, the tree stays where it is",
			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "refused", p.destination, p.source, id.Ident(), "identity mismatch"))
	}
	if forwardMoveUnregistered(root, p.source, p.destination, id.Ident()) && !repairRegistration(root, p.destination) {
		const msg = "forward move left git's registration on the source and the repair did not hold"
		return false, "the reverse move did not run: " + msg + ", the tree stands at " + p.destination + " (pending)",
			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(), msg))
	}
	if rerr := reverseGitMove(root, p.destination, p.source); rerr != nil && !treeMovedBack(root, p.source, p.destination, id.Ident()) {
		// The tree stands at the destination: the re-key and the rebind stay
		// pointing at it — reverting them would name a location the tree
		// does not occupy. The pending line is the convergence record; it
		// names the destination as the line's OLD path (the orientation the
		// reverse order travels), and MovedBy reads both ends of a pending
		// line (card t1640 review F5).
		return false, fmt.Sprintf("the reverse move failed, the tree stands at %s (pending)", p.destination),
			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(),
				fmt.Sprintf("reverse move failed: %v", rerr)))
	}
	// The tree is back at the source. The metadata must follow it (C3): a
	// failed inverse re-key or rebind leaves state naming a location the tree
	// does not occupy, which is pending, never reversed.
	if terr := rekeyAndRebindBack(root, journalStep{OldPath: p.source, NewPath: p.destination}); terr != nil {
		return false, fmt.Sprintf("the tree is back at the source but its metadata was not restored (%v) (pending)", terr),
			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(),
				fmt.Sprintf("metadata restore failed: %v", terr)))
	}
	return true, "", appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "reversed", p.destination, p.source, id.Ident()))
}

// forwardMoveUnregistered detects the state a forward `git worktree move` leaves
// when it is killed after the directory rename but before git rewrote its
// registration (card t2053, H1): the destination carries the journal identity,
// the source is gone, and git still registers the source rather than the
// destination. A reverse move cannot run from there ("not a working tree").
// Every other state — including a listing that failed — answers false, so the
// caller takes its ordinary path.
func forwardMoveUnregistered(root, source, destination, ident string) bool {
	if ident == "" {
		return false
	}
	if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
		return false
	}
	destIdent, err := TreeIdent(destination)
	if err != nil || destIdent.Ident() != ident {
		return false
	}
	// The source is NOT required to be listed (card t2053, C5): an emptied or
	// truncated admin gitdir file makes git drop the worktree from its listing
	// altogether. What decides is that git does not register the destination,
	// and only a listing that succeeded can say so.
	paths, lerr := registeredPaths(root)
	if lerr != nil {
		return false
	}
	return !pathListed(paths, destination)
}

// repairRegistration runs one bounded `git worktree repair <destination>` and
// reports whether git now registers the destination. A repair that fails or
// times out answers false; the caller stays pending, never successful.
func repairRegistration(root, destination string) bool {
	if _, err := listGitOutput(root, "worktree", "repair", destination); err != nil {
		return false
	}
	return registeredAt(root, destination)
}

// treeMovedBack reports whether the tree is verifiably back at source after a
// reverse `git worktree move` that returned an error (card t2053): git can be
// killed by the deadline AFTER the rename completed. It answers true only when
// all three hold:
//
//   - the source carries the journal's identity (ident is never empty here — an
//     empty ident proves nothing),
//   - the destination no longer does, and
//   - git's registration names the source. A move killed between the rename
//     and the registration update leaves worktrees/<id>/gitdir naming the
//     destination, so a missing registration gets ONE bounded
//     `git worktree repair <source>` and is checked again.
//
// Every other state — the destination still carries the identity, neither path
// can be read, a registration that stays on the destination — is ambiguous and
// fails safe to the caller's pending path.
func treeMovedBack(root, source, destination, ident string) bool {
	if ident == "" {
		return false
	}
	srcIdent, err := TreeIdent(source)
	if err != nil || srcIdent.Ident() != ident {
		return false
	}
	switch destIdent, dstate, _ := probeTree(destination); dstate {
	case treeUnreadable:
		return false // an identity that cannot be read does not show the tree left
	case treeIdentified:
		if destIdent.Ident() == ident {
			return false
		}
	}
	if registeredAt(root, source) {
		return true
	}
	if _, rerr := listGitOutput(root, "worktree", "repair", source); rerr != nil {
		return false
	}
	return registeredAt(root, source)
}

// registeredAt reports whether git lists path (spelling-safe) as a registered
// worktree of root. A listing that fails or times out answers false.
func registeredAt(root, path string) bool {
	paths, err := registeredPaths(root)
	if err != nil {
		return false
	}
	return pathListed(paths, path)
}

// registeredPaths lists the worktree paths git registers for root, bounded; an
// error means the listing could not be read (not that nothing is registered).
func registeredPaths(root string) ([]string, error) {
	out, err := listGitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	records, perr := ParseWorktreeList(out)
	if perr != nil {
		return nil, perr
	}
	paths := make([]string, 0, len(records))
	for _, rec := range records {
		paths = append(paths, rec.Path)
	}
	return paths, nil
}

// pathListed reports whether path (spelling-safe) is among paths.
func pathListed(paths []string, path string) bool {
	want := canonicalTreePath(path)
	for _, p := range paths {
		if canonicalTreePath(p) == want {
			return true
		}
	}
	return false
}

// rekeyAndRebindBack undoes the state-file re-key and the card binding of a
// step (new path -> old path). A missing entry or binding is already converged.
func rekeyAndRebindBack(root string, s journalStep) error {
	if rerr := hook.RekeyWorktreeEntryInverse(root, s.NewPath, s.OldPath); rerr != nil &&
		!strings.Contains(rerr.Error(), "no entry keyed to") {
		return fmt.Errorf("%s: reverse re-key failed (%v)", s.OldPath, rerr)
	}
	if rerr := rebindCardForPath(root, s.NewPath, s.OldPath); rerr != nil {
		return fmt.Errorf("%s: reverse card rebind failed (%v)", s.OldPath, rerr)
	}
	return nil
}

// rebindCardForPath rebinds the card record bound to oldPath onto newPath
// when the factory record holds such a binding. No factory record, no
// binding — the step is skipped silently when the store is absent (the
// store is machine-local and may be missing in a repository that carries no
// factory state). The store opens at homestate.FactoryDBPath(root) — the
// layout <root>/.moai/factory/factory.db, not <root>/.moai/factory.db —
// and through a WRITABLE connection: RebindCardWorktree is an UPDATE and a
// mode=ro connection would refuse it (card t1640 review P1-1).
func rebindCardForPath(root, oldPath, newPath string) error {
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		return nil
	}
	if _, serr := os.Stat(dbPath); serr != nil {
		if errors.Is(serr, os.ErrNotExist) {
			return nil // no factory store in this repository: nothing to rebind
		}
		return fmt.Errorf("inspect the factory store %s for the card rebind: %w", dbPath, serr)
	}
	db, err := homestate.OpenFactoryPath(dbPath)
	if err != nil {
		return fmt.Errorf("open the factory store for the card rebind: %w", err)
	}
	defer func() { _ = db.Close() }()
	cardID, ok, err := db.CardRecordingPath(context.Background(), oldPath)
	if err != nil {
		return fmt.Errorf("find the card recording %s: %w", oldPath, err)
	}
	recorded := oldPath
	if !ok {
		// The binding may be recorded under another spelling of the same
		// directory (card t2053, C10): match canonically and rebind from the
		// spelling the row actually carries.
		bindings, berr := db.CardWorktreeBindings(context.Background())
		if berr != nil {
			return fmt.Errorf("list the card bindings: %w", berr)
		}
		want := worktreeroot.CanonicalSpelling(oldPath)
		for _, b := range bindings {
			if worktreeroot.CanonicalSpelling(b.Path) == want {
				cardID, recorded, ok = b.CardID, b.Path, true
				break
			}
		}
		if !ok {
			return nil
		}
	}
	if recorded == newPath {
		return nil
	}
	_, err = db.RebindCardWorktree(context.Background(), cardID, recorded, newPath, "worktree-migrate", time.Now())
	return err
}

// verifyRow is one line of the --verify report.
type verifyRow struct {
	Kind    string `json:"kind"`
	OldPath string `json:"old_path,omitempty"`
	NewPath string `json:"new_path,omitempty"`
	Journal string `json:"journal,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// latestLocationResult is the result of the latest location line of the move
// list that carries ident ("" when there is none).
func latestLocationResult(root, ident string) (string, error) {
	entries, err := loadMoveList(root)
	if err != nil {
		return "", err
	}
	result := ""
	for _, rec := range entries {
		if rec.Ident == ident && isLocationLine(rec.Result) {
			result = rec.Result
		}
	}
	return result, nil
}

// treeCarries reports whether path holds a linked worktree carrying ident. A
// path that is provably not such a tree carries nothing (false, nil); one whose
// identity cannot be read is unknown (false, error).
func treeCarries(path, ident string) (bool, error) {
	if ident == "" {
		return false, nil
	}
	id, state, err := probeTree(path)
	if state == treeUnreadable {
		return false, err
	}
	return state == treeIdentified && id.Ident() == ident, nil
}

// lstatPresence separates present, absent (fs.ErrNotExist only) and unknown.
func lstatPresence(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// verifyStep classifies one journal step (the latest per identity). The kinds
// that fail the verification are absent, mismatch, unregistered, pending,
// failed, reversed and unverifiable; intact and restored are healthy. A path that
// cannot be inspected, or an identity that cannot be read, is unverifiable with
// the error — only fs.ErrNotExist means a path is gone (card t2053, L2).
func verifyStep(root string, s journalStep) (kind, reason string, healthy bool) {
	unknown := func(what string, err error) (string, string, bool) {
		return "unverifiable", fmt.Sprintf("%s cannot be inspected: %v", what, err), false
	}
	dstExists, dstErr := lstatPresence(s.NewPath)
	if dstErr != nil {
		return unknown("the destination", dstErr)
	}
	switch s.Outcome {
	case "", "moved":
		if dstExists {
			carries, err := treeCarries(s.NewPath, s.Ident)
			if err != nil {
				return unknown("the destination's identity", err)
			}
			if !carries {
				return "mismatch", "the destination is not the tree the journal moved", false
			}
			if !registeredAt(root, s.NewPath) {
				return "unregistered", "git does not list the destination as a worktree", false
			}
			return "intact", "", true
		}
		carries, err := treeCarries(s.OldPath, s.Ident)
		if err != nil {
			return unknown("the source's identity", err)
		}
		if carries && registeredAt(root, s.OldPath) {
			r, lerr := latestLocationResult(root, s.Ident)
			if lerr != nil {
				return unknown("the move list", lerr)
			}
			if r == "restored" || r == "reversed" {
				return "restored", "rolled back to the source", true
			}
		}
		return "absent", "", false
	default: // pending, failed, reversed: a step that did not end as a completed move
		if dstExists {
			carries, err := treeCarries(s.NewPath, s.Ident)
			if err != nil {
				return unknown("the destination's identity", err)
			}
			if carries {
				return s.Outcome, "the tree stands at the destination; run --rollback on this journal", false
			}
			return "", "", true // a foreign tree at the destination: not this step's tree
		}
		srcExists, srcErr := lstatPresence(s.OldPath)
		if srcErr != nil {
			return unknown("the source", srcErr)
		}
		if !srcExists {
			return "", "", true // neither end exists: the tree is gone
		}
		carries, err := treeCarries(s.OldPath, s.Ident)
		if err != nil {
			return unknown("the source's identity", err)
		}
		if !carries || !registeredAt(root, s.OldPath) {
			return s.Outcome, "the tree is at the source but git does not register it under the journal identity; run --rollback on this journal", false
		}
		if s.Outcome == "pending" {
			r, lerr := latestLocationResult(root, s.Ident)
			if lerr != nil {
				return unknown("the move list", lerr)
			}
			if r != "restored" && r != "reversed" {
				return "pending", "an unsettled step; run --rollback on this journal to settle it", false
			}
		}
		return "", "", true
	}
}

// runMigrateVerify reports each recorded move in the journals (card t2053,
// C2). Only the latest step per identity counts, so a tree moved again by a
// later run is not judged by an older journal. It fails — with a row — on a
// journal it cannot read, on a step left unsettled or stranded, on a
// destination that git does not list, and on a step without an identity;
// a step the rollback restored reads as restored. A repository without any
// journal directory has nothing to verify. The migration lock keeps a
// running mover from being read mid-write.
func runMigrateVerify(cmd *cobra.Command, root string, asJSON bool) error {
	mlDir := filepath.Join(root, ".moai", "state", "worktree-migrate")
	if _, err := os.Stat(mlDir); errors.Is(err, os.ErrNotExist) {
		if asJSON {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[]")
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "nothing to verify: no migration journals")
		}
		return nil
	}
	return withMigrationLock(root, func() error {
		journalPaths, err := orderedJournals(mlDir)
		if err != nil {
			return fmt.Errorf("read the migration journals: %w", err)
		}
		rows := []verifyRow{}
		failures := 0
		fail := func(r verifyRow) {
			rows = append(rows, r)
			failures++
		}
		type located struct {
			step    journalStep
			journal string
		}
		latest := map[string]located{}
		var order []string
		for _, path := range journalPaths {
			data, rerr := os.ReadFile(path)
			var journal struct {
				Steps []journalStep `json:"steps"`
			}
			if rerr != nil {
				fail(verifyRow{Kind: "unreadable", Journal: path, Reason: rerr.Error()})
				continue
			}
			if jerr := json.Unmarshal(data, &journal); jerr != nil {
				fail(verifyRow{Kind: "unreadable", Journal: path, Reason: jerr.Error()})
				continue
			}
			for _, s := range journal.Steps {
				if s.Ident == "" {
					fail(verifyRow{Kind: "unverifiable", OldPath: s.OldPath, NewPath: s.NewPath, Journal: path,
						Reason: "the journal step carries no identity"})
					continue
				}
				if _, seen := latest[s.Ident]; !seen {
					order = append(order, s.Ident)
				}
				latest[s.Ident] = located{step: s, journal: path}
			}
		}
		intact := 0
		for _, ident := range order {
			l := latest[ident]
			kind, reason, healthy := verifyStep(root, l.step)
			if kind == "" {
				continue
			}
			row := verifyRow{Kind: kind, OldPath: l.step.OldPath, NewPath: l.step.NewPath, Journal: l.journal, Reason: reason}
			if healthy {
				rows = append(rows, row)
				if kind == "intact" {
					intact++
				}
			} else {
				fail(row)
			}
		}
		if asJSON {
			body, jerr := json.MarshalIndent(rows, "", "  ")
			if jerr != nil {
				return jerr
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		} else {
			for _, r := range rows {
				line := fmt.Sprintf("%s\t%s → %s", r.Kind, r.OldPath, r.NewPath)
				if r.Kind == "unreadable" {
					line = fmt.Sprintf("%s\t%s", r.Kind, r.Journal)
				}
				if r.Reason != "" {
					line += " (" + r.Reason + ")"
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%d recorded move(s) intact, %d mismatched\n", intact, failures)
		}
		if failures > 0 {
			return fmt.Errorf("verification found %d mismatched move(s)", failures)
		}
		return nil
	})
}

// runMigrateRollback reverses each journaled move: move back, re-key back,
// rebind back, restored lines appended. The identity check before every
// reverse move compares the tree's ident with the journal's (D-12).
// runMigrateRollback reverses each journaled move: move back, re-key back,
// rebind back, restored lines appended. The identity check before every
// reverse move compares the tree's ident with the journal's (D-12); a tree
// standing at the destination under a DIFFERENT identity is refused with a
// `refused` line (reason `identity mismatch`) and fails the run — never a
// silent skip (card t1640 review F2). The manifest precondition applies to
// settled moves only (the journal's `moved` steps carrying a manifest
// digest): a pending step — a kill mid-move — reverses unconditionally so
// the re-run converges (REQ-WHR-018's fifth entry). Reverse-step errors
// propagate: a failed re-key or rebind records the step pending (no
// `restored` line) and fails the run for the next rollback to converge.
func runMigrateRollback(cmd *cobra.Command, root, journalPath string, asJSON bool) error {
	return withMigrationLock(root, func() error {
		data, err := os.ReadFile(journalPath)
		if err != nil {
			return fmt.Errorf("read the journal: %w", err)
		}
		var journal struct {
			Steps []journalStep `json:"steps"`
		}
		if err := json.Unmarshal(data, &journal); err != nil {
			return fmt.Errorf("parse the journal: %w", err)
		}
		mlPath := moveListPath(root)
		restored, refused, failed := 0, 0, 0
		var problems []string
		rows := []rollbackRow{}
		note := func(s journalStep, outcome, reason string) {
			rows = append(rows, rollbackRow{OldPath: s.OldPath, NewPath: s.NewPath, Outcome: outcome, Reason: reason})
		}
		refuse := func(s journalStep, reason, problem string) error {
			if lerr := appendMoveListLine(mlPath, m6MoveLineReason("refused", s.NewPath, s.OldPath, s.Ident, reason)); lerr != nil {
				return lerr
			}
			problems = append(problems, problem)
			refused++
			note(s, "refused", reason)
			return nil
		}
		hold := func(s journalStep, msg, problem string) error {
			if lerr := appendMoveListLine(mlPath, m6MoveLineReason("pending", s.NewPath, s.OldPath, s.Ident, msg)); lerr != nil {
				return lerr
			}
			problems = append(problems, problem)
			failed++
			note(s, "pending", msg)
			return nil
		}
		fail := func(s journalStep, problem string) {
			problems = append(problems, problem)
			failed++
			note(s, "pending", problem)
		}
		for i := len(journal.Steps) - 1; i >= 0; i-- {
			s := journal.Steps[i]
			if s.Outcome == "failed" && s.Reason == reasonInterruptedBeforeMove {
				// Settled by a later run as a step that never moved its tree:
				// there is nothing to undo, and the identity now at the destination
				// belongs to a newer move (L1).
				note(s, "skipped", "the step never moved the tree ("+reasonInterruptedBeforeMove+")")
				continue
			}
			if _, serr := os.Stat(s.NewPath); serr != nil && !errors.Is(serr, os.ErrNotExist) {
				// A destination that cannot be inspected is not an absent one
				// (K2): where the tree stands is unknown, so nothing is restored.
				msg := fmt.Sprintf("the destination cannot be inspected: %v", serr)
				if herr := hold(s, msg, fmt.Sprintf("%s: %s", s.NewPath, msg)); herr != nil {
					return herr
				}
				continue
			} else if serr != nil {
				// The destination is gone. The tree either never left the source
				// or a reverse move renamed it back and was cut off before git's
				// registration, the re-key and the card binding followed (card
				// t2053, J1). Only a source that carries the step's identity AND
				// is registered there AND has nothing left to undo is "nothing to
				// do"; every other state converges or stays pending.
				if _, lerr := os.Lstat(s.OldPath); lerr != nil {
					if errors.Is(lerr, os.ErrNotExist) {
						continue // neither end exists: nothing here to act on
					}
					// A source that cannot be inspected is not a gone one (K2).
					msg := fmt.Sprintf("the source cannot be inspected: %v", lerr)
					if herr := hold(s, msg, fmt.Sprintf("%s: %s", s.OldPath, msg)); herr != nil {
						return herr
					}
					continue
				}
				if s.Ident == "" {
					// An identity-less step cannot be matched to the tree at the
					// source (G10): said out loud, never a silent skip.
					note(s, "refused", "the journal step carries no identity")
					problems = append(problems, fmt.Sprintf("%s: the journal step carries no identity; the tree stays where it is", s.OldPath))
					refused++
					continue
				}
				srcIdent, sstate, ierr := probeTree(s.OldPath)
				if sstate == treeUnreadable {
					msg := fmt.Sprintf("the source's identity cannot be read: %v", ierr)
					if herr := hold(s, msg, fmt.Sprintf("%s: %s", s.OldPath, msg)); herr != nil {
						return herr
					}
					continue
				}
				if sstate != treeIdentified || srcIdent.Ident() != s.Ident {
					if rerr := refuse(s, "identity mismatch", fmt.Sprintf("%s: identity mismatch at the source; the tree stays where it is", s.OldPath)); rerr != nil {
						return rerr
					}
					continue
				}
				alreadyRegistered := registeredAt(root, s.OldPath)
				if !alreadyRegistered && !treeMovedBack(root, s.OldPath, s.NewPath, s.Ident) {
					const msg = "the tree is back at the source but git's registration could not be repaired"
					if herr := hold(s, msg, fmt.Sprintf("%s: %s", s.OldPath, msg)); herr != nil {
						return herr
					}
					continue
				}
				if terr := rekeyAndRebindBack(root, s); terr != nil {
					fail(s, terr.Error())
					continue
				}
				// Converged or repaired: either way the tree is at the source,
				// so the line that releases its protection is written (C13).
				if lerr := appendMoveListLine(mlPath, m6MoveLine("restored", s.NewPath, s.OldPath, s.Ident)); lerr != nil {
					return lerr
				}
				if alreadyRegistered {
					note(s, "restored", "already at the source")
					continue // nothing was moved: not counted
				}
				note(s, "restored", "")
				restored++
				continue
			}
			// The identity rule, shared with the apply-time reverse (G8): an
			// unreadable identity or a step without one is not a match.
			if s.Ident == "" {
				if rerr := refuse(s, "no identity", fmt.Sprintf("%s: the journal step carries no identity; the tree stays where it is", s.NewPath)); rerr != nil {
					return rerr
				}
				continue
			}
			destIdent, dstate, terr := probeTree(s.NewPath)
			if dstate == treeUnreadable {
				msg := fmt.Sprintf("the destination's identity cannot be read: %v", terr)
				if herr := hold(s, msg, fmt.Sprintf("%s: %s", s.NewPath, msg)); herr != nil {
					return herr
				}
				continue
			}
			if dstate != treeIdentified || destIdent.Ident() != s.Ident {
				if rerr := refuse(s, "identity mismatch", fmt.Sprintf("%s: identity mismatch; the tree stays where it is", s.NewPath)); rerr != nil {
					return rerr
				}
				continue
			}
			// The manifest-match precondition belongs to every step that is
			// not pending and carries a digest (C14); a pending step reverses
			// unconditionally (REQ-WHR-018's fifth entry).
			if s.Outcome != "pending" && s.Manifest != "" {
				if s.ManifestVersion > manifestFormatVersion {
					if rerr := refuse(s, fmt.Sprintf("unknown manifest_version %d", s.ManifestVersion),
						fmt.Sprintf("%s: unknown manifest_version %d (this build reads up to %d); the step stays pending", s.NewPath, s.ManifestVersion, manifestFormatVersion)); rerr != nil {
						return rerr
					}
					continue
				}
				if manifestForJournal(s.NewPath, s.ManifestVersion) != s.Manifest {
					if rerr := refuse(s, "manifest mismatch", fmt.Sprintf("%s: manifest mismatch; the step stays pending", s.NewPath)); rerr != nil {
						return rerr
					}
					continue
				}
			}
			if forwardMoveUnregistered(root, s.OldPath, s.NewPath, s.Ident) && !repairRegistration(root, s.NewPath) {
				const msg = "forward move left git's registration on the source and the repair did not hold"
				if herr := hold(s, msg, fmt.Sprintf("%s: %s", s.NewPath, msg)); herr != nil {
					return herr
				}
				continue
			}
			if merr := reverseGitMove(root, s.NewPath, s.OldPath); merr != nil && !treeMovedBack(root, s.OldPath, s.NewPath, s.Ident) {
				if herr := hold(s, fmt.Sprintf("reverse move failed: %v", merr), fmt.Sprintf("%s: reverse move failed (%v)", s.NewPath, merr)); herr != nil {
					return herr
				}
				continue
			}
			if terr := rekeyAndRebindBack(root, s); terr != nil {
				fail(s, terr.Error())
				continue
			}
			if lerr := appendMoveListLine(mlPath, m6MoveLine("restored", s.NewPath, s.OldPath, s.Ident)); lerr != nil {
				return lerr
			}
			note(s, "restored", "")
			restored++
		}
		if asJSON {
			body, jerr := json.MarshalIndent(rows, "", "  ")
			if jerr != nil {
				return jerr
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "restored %d tree(s), %d refused, %d pending\n", restored, refused, failed)
		}
		if len(problems) > 0 {
			return fmt.Errorf("rollback: %s", strings.Join(problems, "; "))
		}
		return nil
	})
}

// rollbackRow is one line of the --rollback --json report.
type rollbackRow struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	Outcome string `json:"outcome"` // restored | refused | pending | skipped (a step that never moved)
	Reason  string `json:"reason,omitempty"`
}

// canonicalTreePath resolves symlinks for path comparisons (the same
// canonical form planLegacyWorktreeMigrationFor compares in).
func canonicalTreePath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

// listGitOutput runs git in dir and returns stdout, bounded by
// gitCommandTimeout.
func listGitOutput(dir string, args ...string) (string, error) {
	return gitOutputWithin(gitCommandTimeout, dir, args...)
}

// gitOutputWithin is listGitOutput under an explicit bound (the `worktree
// move` call carries its own, longer one).
func gitOutputWithin(timeout time.Duration, dir string, args ...string) (string, error) {
	argv := append([]string{"-C", dir}, args...)
	out, err := boundedOutput(timeout, "", "git", argv...)
	return string(out), err
}

// m6MoveLine builds one move-list record in the D-11 field list.
func m6MoveLine(result, oldPath, newPath, ident string) string {
	return m6MoveLineTriggerReason("migrate", result, oldPath, newPath, ident, "")
}

// m6MoveLineTrigger is m6MoveLine with the D-11 trigger field set (the
// engine's caller: "migrate" for the verb, "update" for the routine
// consolidation — AC-WHR-019(g)).
func m6MoveLineTrigger(trigger, result, oldPath, newPath, ident string) string {
	return m6MoveLineTriggerReason(trigger, result, oldPath, newPath, ident, "")
}

// m6MoveLineReason is m6MoveLine with the D-11 reason field set.
func m6MoveLineReason(result, oldPath, newPath, ident, reason string) string {
	return m6MoveLineTriggerReason("migrate", result, oldPath, newPath, ident, reason)
}

// m6MoveLineTriggerReason is the full D-11 record builder.
func m6MoveLineTriggerReason(trigger, result, oldPath, newPath, ident, reason string) string {
	b, _ := json.Marshal(map[string]any{
		"schema": 1, "ts": time.Now().UTC().Format(time.RFC3339),
		"run": "", "trigger": trigger, "tree": filepath.Base(oldPath),
		"card": "", "old_path": oldPath, "new_path": newPath,
		"result": result, "reason": reason, "ident": ident,
	})
	return string(b)
}

// treeManifest is the manifest of one tree at a point in time: the
// AC-WHR-017a field list (HEAD, branch, unpushed commit list, status,
// ignored-file count, reports file count) plus two digests over the files git
// does not track:
//
//   - ContentDigest covers every untracked, NON-ignored file's bytes — the
//     files a user wrote and did not ignore (symlinks hash their target).
//   - IgnoredDigest covers every IGNORED file's path, size and file type,
//     taken from Lstat alone; the content is never read (card t2053).
//
// Why the split: a digest over the bytes of every ignored file reads whatever
// the tree ignores — node_modules, build output, caches — twice per moved
// tree, and a project with hundreds of trees went silent for minutes. The
// price is a documented blind spot: a same-size content change inside an
// ignored file is NOT detected. A missing ignored file, a size change and a
// type change (file / symlink / directory / special) are still detected. The
// pre-t2053 digest (card t1640 review P2f) caught the same-size change too;
// ignored files are by definition reproducible output, which is the trade.
type treeManifest struct {
	HEAD          string
	Branch        string
	Unpushed      string
	Status        string
	Diff          string
	IgnoredCount  int
	ReportsCount  int
	ContentDigest string
	IgnoredDigest string
}

// captureManifest digests the working state of the tree at path through the
// full AC-WHR-017a manifest. It returns "" when the manifest could not be read
// in full (see captureManifestChecked); an empty digest never equals a real
// one, so every comparison against it fails safe.
func captureManifest(tree string) string {
	digest, _ := captureManifestChecked(tree)
	return digest
}

// manifestForJournal captures the digest the way the journal step was written:
// a step without a format version came from the pre-t2053 build, whose digest
// covered the bytes of ignored files as well (H2, card t2053).
func manifestForJournal(tree string, version int) string {
	if version >= manifestFormatVersion {
		return captureManifest(tree)
	}
	return captureLegacyManifest(tree)
}

// captureLegacyManifest reproduces the pre-t2053 manifest digest: the same
// fields as treeManifest minus IgnoredDigest, with ContentDigest hashing the
// bytes of every untracked file, ignored ones included. It exists only so
// --rollback can verify journals an older build wrote; new journals never use
// it. A timed-out git call returns "", which never equals a recorded digest.
func captureLegacyManifest(tree string) string {
	type legacyManifest struct {
		HEAD          string
		Branch        string
		Unpushed      string
		Status        string
		Diff          string
		IgnoredCount  int
		ReportsCount  int
		ContentDigest string
	}
	var m legacyManifest
	g := &manifestGit{tree: tree}
	git := func(args ...string) string { return g.run(false, args...) }
	gitHEAD := func(args ...string) string { return g.run(true, args...) }
	m.HEAD = strings.TrimSpace(gitHEAD("rev-parse", "HEAD"))
	m.Branch = strings.TrimSpace(gitHEAD("rev-parse", "--abbrev-ref", "HEAD"))
	m.Unpushed = gitHEAD("rev-list", "HEAD", "--not", "--remotes")
	m.Status = git("status", "--porcelain")
	m.Diff = git("diff", "--cached") + git("diff")
	untracked := splitNul(git("ls-files", "--others", "-z"))
	ignored := splitNul(git("ls-files", "--others", "--ignored", "--exclude-standard", "-z"))
	if g.err != nil {
		return ""
	}
	m.IgnoredCount = len(ignored)
	m.ContentDigest = hashUntrackedContent(tree, append(append([]string{}, untracked...), ignored...))
	reportsCount, cerr := countReportFiles(tree)
	if cerr != nil {
		return "" // an unreadable reports tree proves nothing: never matches
	}
	m.ReportsCount = reportsCount
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// manifestRead is the seam for one git read of a manifest capture.
var manifestRead = listGitOutput

// manifestGit runs the git reads of a manifest capture and remembers the first
// failure (card t2053, C9). A manifest read in which any git call exited
// non-zero — or timed out — proves nothing about the tree: two such captures
// would compare equal and certify a tree nobody inspected. The one benign
// failure is an unborn HEAD, where the HEAD-relative reads fail by nature.
type manifestGit struct {
	tree   string
	err    error
	unborn *bool
}

// run executes one git read; headRelative marks a read that fails on an unborn
// HEAD without that meaning anything.
func (g *manifestGit) run(headRelative bool, args ...string) string {
	if g.err != nil {
		return ""
	}
	out, err := manifestRead(g.tree, args...)
	return g.settle(headRelative, args, out, err)
}

// settle applies the failure rules to one finished read: the first failure ends
// the capture, and later reads settle to "". It is run's second half, split out
// so the reads of a capture can be issued together and settled in call order
// (card t2070) with exactly the verdict the one-after-the-other capture gave.
func (g *manifestGit) settle(headRelative bool, args []string, out string, err error) string {
	if g.err != nil {
		return ""
	}
	if err == nil {
		return out
	}
	if headRelative && !errors.Is(err, ErrCommandTimeout) && g.unbornHEAD() {
		return out
	}
	g.err = fmt.Errorf("git %s: %w", args[0], err)
	return out
}

// unbornHEAD reports whether HEAD names a branch that has no commit yet.
func (g *manifestGit) unbornHEAD() bool {
	if g.unborn == nil {
		_, symErr := listGitOutput(g.tree, "symbolic-ref", "-q", "HEAD")
		_, verErr := listGitOutput(g.tree, "rev-parse", "--verify", "-q", "HEAD")
		unborn := symErr == nil && verErr != nil
		g.unborn = &unborn
	}
	return *g.unborn
}

// captureManifestChecked is captureManifest plus the reason a manifest could
// not be completed: a git call that timed out. Other read failures still
// degrade to zero values inside the manifest (see captureTreeManifest).
func captureManifestChecked(tree string) (string, error) {
	m, err := captureTreeManifest(tree)
	if err != nil {
		return "", err
	}
	b, merr := json.Marshal(m)
	if merr != nil {
		return "", nil
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// captureTreeManifest reads the AC-WHR-017a manifest fields of the tree at
// path. A git call that exits non-zero degrades to a zero value — the pre/post
// comparison then notices the change of shape, exactly as an unreadable field
// should fail a field-for-field comparison. A git call that TIMES OUT does
// not: the capture stops at once and returns an error wrapping
// ErrCommandTimeout, because two timed-out captures would otherwise compare
// equal and certify a tree nobody inspected.
func captureTreeManifest(tree string) (treeManifest, error) {
	var m treeManifest
	g := &manifestGit{tree: tree}

	// The eight reads are independent of one another, so they are issued
	// together and settled afterwards in this order (card t2070): they took
	// about 440ms one after the other and the capture runs twice per tree. The
	// tracked-change content (REQ-WHR-017's content digest over the
	// tracked-change diff — card t1640 review F7): a content mutation that
	// preserves the status shape must change the manifest. Both the unstaged
	// and the staged forms ride the manifest.
	reads := []struct {
		headRelative bool
		args         []string
	}{
		{true, []string{"rev-parse", "HEAD"}},
		{true, []string{"rev-parse", "--abbrev-ref", "HEAD"}},
		{true, []string{"rev-list", "HEAD", "--not", "--remotes"}},
		{false, []string{"status", "--porcelain"}},
		{false, []string{"diff", "--cached"}},
		{false, []string{"diff"}},
		{false, []string{"ls-files", "--others", "--exclude-standard", "-z"}},
		{false, []string{"ls-files", "--others", "--ignored", "--exclude-standard", "-z"}},
	}
	outs := make([]string, len(reads))
	errs := make([]error, len(reads))
	var wg sync.WaitGroup
	for i := range reads {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			outs[i], errs[i] = manifestRead(tree, reads[i].args...)
		}(i)
	}
	wg.Wait()
	settled := make([]string, len(reads))
	for i := range reads {
		settled[i] = g.settle(reads[i].headRelative, reads[i].args, outs[i], errs[i])
	}

	m.HEAD = strings.TrimSpace(settled[0])
	m.Branch = strings.TrimSpace(settled[1])
	m.Unpushed = settled[2]
	m.Status = settled[3]
	m.Diff = settled[4] + settled[5]
	untracked := splitNul(settled[6])
	ignored := splitNul(settled[7])
	if g.err != nil {
		return m, g.err
	}
	m.IgnoredCount = len(ignored)
	m.ContentDigest = hashUntrackedContent(tree, untracked)
	m.IgnoredDigest = hashIgnoredMetadata(tree, ignored)

	reportsCount, cerr := countReportFiles(tree)
	if cerr != nil {
		return m, cerr
	}
	m.ReportsCount = reportsCount
	return m, nil
}

// countReportFiles counts the files under the tree's .moai/reports. An absent
// reports directory (fs.ErrNotExist) holds none; any other walk failure is an
// error — a count taken over a tree it could not fully read proves nothing.
func countReportFiles(tree string) (int, error) {
	reports := filepath.Join(tree, ".moai", "reports")
	count := 0
	err := filepath.WalkDir(reports, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == reports && errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("count the reports of %s: %w", tree, err)
	}
	return count, nil
}

// sortedUniquePaths trims, de-duplicates and sorts a git path listing.
func sortedUniquePaths(listing []string) []string {
	seen := map[string]bool{}
	var paths []string
	for _, p := range listing {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// hashUntrackedContent digests the bytes of the untracked, non-ignored files.
func hashUntrackedContent(tree string, listing []string) string {
	digest := sha256.New()
	for _, p := range sortedUniquePaths(listing) {
		_, _ = digest.Write([]byte(p))
		_, _ = digest.Write([]byte{0})
		full := filepath.Join(tree, filepath.FromSlash(p))
		if info, lerr := os.Lstat(full); lerr == nil {
			// Symlinks hash their TARGET (card t1640 review F11): os.ReadFile
			// would follow the link, and a FIFO or device would block the
			// apply while the migration lock is held. Only regular files are
			// read; everything else hashes its type marker.
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				if target, rerr := os.Readlink(full); rerr == nil {
					_, _ = digest.Write([]byte("link:" + target))
				} else {
					_, _ = digest.Write([]byte("\x00unreadable-link"))
				}
			case info.Mode().IsRegular():
				if body, rerr := os.ReadFile(full); rerr == nil {
					_, _ = digest.Write(body)
				} else {
					_, _ = digest.Write([]byte("\x00unreadable"))
				}
			default:
				_, _ = digest.Write([]byte("special:" + info.Mode().String()))
			}
		} else {
			_, _ = digest.Write([]byte("\x00unreadable"))
		}
		_, _ = digest.Write([]byte{0})
	}
	return hex.EncodeToString(digest.Sum(nil))
}

// @MX:NOTE: [AUTO] same-size content changes inside ignored files are NOT detected (card t2053); only path, size and file type are compared
// hashIgnoredMetadata digests the ignored files by path, file type and size
// from Lstat alone. It never opens a file: an unreadable (mode 000) or huge
// ignored file costs one stat call, and a symlink is recorded as a symlink
// (its target is not followed or read).
func hashIgnoredMetadata(tree string, listing []string) string {
	digest := sha256.New()
	for _, p := range sortedUniquePaths(listing) {
		kind, size := "missing", int64(0)
		if info, lerr := os.Lstat(filepath.Join(tree, filepath.FromSlash(p))); lerr == nil {
			kind, size = "regular", info.Size()
			if t := info.Mode().Type(); t != 0 {
				kind = t.String()
			}
		}
		_, _ = fmt.Fprintf(digest, "%s\x00%s\x00%d\x00", p, kind, size)
	}
	return hex.EncodeToString(digest.Sum(nil))
}

// splitNul splits a NUL-delimited git -z listing.
func splitNul(s string) []string {
	return strings.Split(strings.TrimSuffix(s, "\x00"), "\x00")
}

// isUnderWorktreePrefix reports whether path lies strictly inside prefix,
// spelling-safe: the containment walk of classify.go (underPath) with the
// equal case excluded, for the migration's destination-prefix checks.
func isUnderWorktreePrefix(path, prefix string) bool {
	return prefix != "" && underPath(path, prefix)
}

// processUsesWorktree reports whether one of the active processes works
// inside the tree at path.

// orderedJournals lists the journal files of dir in TRUE creation order (card
// t2053, K1): `<stamp>.json` first, then `<stamp>-1.json`, `<stamp>-2.json`...
// numerically. Plain lexical order puts `<stamp>-1` before `<stamp>` ('-' sorts
// before '.'), so every reader of the journal directory goes through this one
// helper instead of trusting os.ReadDir's order.
func orderedJournals(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	type journalName struct {
		path, stamp string
		n           int
	}
	var names []journalName
	for _, e := range entries {
		base, ok := strings.CutSuffix(e.Name(), ".json")
		if !ok {
			continue
		}
		stamp, n := base, 0
		if i := strings.LastIndex(base, "-"); i > 0 {
			if v, perr := strconv.Atoi(base[i+1:]); perr == nil && v > 0 {
				stamp, n = base[:i], v
			}
		}
		names = append(names, journalName{path: filepath.Join(dir, e.Name()), stamp: stamp, n: n})
	}
	sort.SliceStable(names, func(i, j int) bool {
		if names[i].stamp != names[j].stamp {
			return names[i].stamp < names[j].stamp
		}
		return names[i].n < names[j].n
	})
	out := make([]string, 0, len(names))
	for _, n := range names {
		out = append(out, n.path)
	}
	return out, nil
}
