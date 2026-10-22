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
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/lockfile"
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
		return runMigrateVerify(cmd, root)
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
	if terr := lockfile.TryLock(f); terr != nil {
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
		report, engineErr := runMigrationEngine(root, apply, "migrate")
		if asJSON {
			body, jerr := json.MarshalIndent(report.rows, "", "  ")
			if jerr != nil {
				return jerr
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(body))
		} else {
			for _, r := range report.rows {
				line := fmt.Sprintf("%s\t%s → %s\t[%s]", r.Class, r.Source, r.Destination, r.Outcome)
				if r.SkipReason != "" {
					line += " (" + r.SkipReason + ")"
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "moved %d, skipped %d, failed %d, reversed %d\n",
				report.moved, report.skipped, report.failed, report.reversed)
		}
		if engineErr != nil {
			return engineErr
		}
		if report.reversed > 0 || report.failed > 0 {
			return fmt.Errorf("migration completed with %d reversed and %d failed trees", report.reversed, report.failed)
		}
		return nil
	}
	if !apply {
		return run()
	}
	return withMigrationLock(root, run)
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
}

// migrationReport carries the run's outcome counts.
type migrationReport struct {
	rows     []migrationRow
	moved    int
	skipped  int
	failed   int
	reversed int
}

// migratePlan is one tree of the plan: the classification drives the
// selection (moai-made moves, everything else never does).
type migratePlan struct {
	source, destination string
	head, branch        string
	class               string
	skip                string
}

// activeProcessCWDProbe is the injectable seam for the process working-
// directory skip cause. The default is the conservative lsof scan — the
// same logic planLegacyWorktreeMigrationFor uses through its package-cli
// seam (the two must agree; only the package boundary differs).
var activeProcessCWDProbe = defaultActiveProcessCWDs

// defaultActiveProcessCWDs scans lsof for the working directories of every
// active process (the conservative process-level guard for worktree moves).
func defaultActiveProcessCWDs() ([]string, error) {
	out, err := exec.Command("lsof", "-a", "-d", "cwd", "-F0n").Output()
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
func planMigrationFor(root string, now time.Time, apply bool) ([]migratePlan, error) {
	out, err := listGitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("list migration worktrees: %w", err)
	}
	records, err := ParseWorktreeList(out)
	if err != nil {
		return nil, fmt.Errorf("parse migration worktrees: %w", err)
	}
	homeBase := homeProjectsDir()
	// The classifier verdicts must agree across list, migrate, tier and
	// retirement: the integration branch is set exactly as list.go sets it
	// (card t1640 review P2b).
	ctx := ClassifyContext{
		PrimaryRoot:       root,
		IntegrationBranch: config.LoadGitFlowIntegrationConfig(root).IntegrationTarget,
		HomeProjectsDir:   homeBase,
		CodexWorktreesDir: codexWorktreesDir(),
		MarkerOwners:      homeMarkerOwners(homeBase),
	}
	if key := worktreeroot.RecordedKey(root); key != "" {
		ctx.OwnerID, _ = worktreeroot.OwnerMarkerOf(filepath.Join(homeBase, key))
	}
	cwd, _ := os.Getwd()
	var plans []migratePlan
	var processCWDs []string
	var processCWDErr error
	processCWDsRead := false
	for _, rec := range records {
		if rec.Bare {
			continue
		}
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
		legacy := underPath(rec.Path, filepath.Join(root, ".moai", "worktrees")) ||
			underPath(rec.Path, filepath.Join(root, ".claude", "worktrees"))
		if cls.Class != ClassMoaiMade || !legacy {
			p.skip = "outside the legacy roots selected for migration"
			plans = append(plans, p)
			continue
		}
		// SPEC-WORKTREE-HOME-ROOT-001 M2 (REQ-WHR-005a): the destination is
		// the resolver's home-root landing path for this tree.
		var dest string
		var derr error
		if apply {
			dest, derr = worktreeroot.PathFor(root, filepath.Base(rec.Path))
		} else {
			dest, derr = worktreeroot.Plan(root, filepath.Base(rec.Path))
		}
		if derr != nil {
			p.skip = fmt.Sprintf("resolve home-root destination: %v", derr)
		} else {
			p.destination = dest
		}
		// The same skip causes as planLegacyWorktreeMigrationFor.
		switch {
		case rec.Prunable:
			p.skip = "Git lists this tree as prunable"
		case rec.Locked:
			p.skip = "Git worktree is locked"
		case isUnderWorktreePrefix(cwd, rec.Path) || canonicalTreePath(cwd) == canonicalTreePath(rec.Path):
			p.skip = "current process is inside this worktree"
		case !processCWDsRead:
			processCWDs, processCWDErr = activeProcessCWDProbe()
			processCWDsRead = true
			fallthrough
		default:
			if processCWDErr != nil {
				p.skip = fmt.Sprintf("cannot inspect active process working directories: %v", processCWDErr)
			} else if processUsesWorktree(processCWDs, rec.Path) {
				p.skip = "an active process is inside this worktree"
			} else if worktreeRegistryProbe(root, rec.Path) != nil {
				p.skip = "session registry cannot be read"
			} else if liveAnchoredProbe(rec.Path, root, now) > 0 {
				p.skip = "active session is anchored in this worktree"
			} else if info, statErr := os.Lstat(rec.Path); statErr != nil || !info.IsDir() {
				p.skip = "source path is missing or not a directory"
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
}

// migrationJournal is one apply run's journal file: every tree is appended
// as pending before its move and settled afterwards, and the whole file is
// rewritten on every change — a kill mid-run leaves the journal naming
// every tree whose move may have started (card t1640 review F3).
type migrationJournal struct {
	path    string
	dir     string
	trigger string
	steps   []journalStep
}

func (j *migrationJournal) flush() error {
	if j == nil || j.path == "" {
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
	if j.path == "" {
		if err := os.MkdirAll(j.dir, 0o755); err != nil {
			return err
		}
		// Reserve each run's name exclusively. A clock tick alone cannot
		// distinguish serialized applies or a wall clock moving backward.
		f, err := os.CreateTemp(j.dir, time.Now().UTC().Format("20060102T150405.000000000")+"-*.json")
		if err != nil {
			return err
		}
		j.path = f.Name()
		if err := f.Close(); err != nil {
			return err
		}
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
			return j.flush()
		}
	}
	return nil
}

// runMigrationEngine plans and (when apply is set) executes the migration.
// The dry run and the apply both list every registered tree with its class;
// only moai-made trees are moved. Non-moai-made trees are reported as
// skipped with their class as the reason. The trigger names the caller in
// the move list ("migrate" for the verb, "update" for the consolidation —
// AC-WHR-019(g)).
func runMigrationEngine(root string, apply bool, trigger string) (*migrationReport, error) {
	report := &migrationReport{}
	plans, err := planMigrationFor(root, time.Now(), apply)
	if err != nil {
		return report, err
	}
	mlPath := moveListPath(root)

	var journal *migrationJournal
	if apply {
		journal = &migrationJournal{
			dir:     filepath.Dir(mlPath),
			trigger: trigger,
		}
	}

	var writeErrs []string
	stateLockLapses := 0
	for _, p := range plans {
		if stateLockLapses >= hook.StateLockLapseBudget && p.class == string(ClassMoaiMade) && p.skip == "" {
			p.skip = "state-file lock lapse budget spent; remaining trees await the next run"
		}
		if apply && p.class == string(ClassMoaiMade) && p.skip == "" {
			// A previous move can take long enough for another process to
			// enter or anchor this source. Refresh every safety input using
			// the same planner immediately before this tree's move steps.
			fresh := migratePlan{source: p.source, class: p.class, skip: "source is no longer registered"}
			currentPlans, err := planMigrationFor(root, time.Now(), true)
			if err != nil {
				return report, err
			}
			for _, current := range currentPlans {
				if worktreeroot.Same(current.source, p.source) {
					fresh = current
					break
				}
			}
			p = fresh
		}
		row := migrationRow{
			Class: p.class, Source: p.source, Destination: p.destination,
			Outcome: "planned", SkipReason: p.skip,
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
			// Inventory includes already-migrated home trees. Their existing
			// receipt is the outcome; an unchanged rerun must not append a
			// second event merely because they are no longer legacy sources.
			if apply && (p.destination != "" || !MovedBy(root, p.source)) {
				ident := ""
				if id, err := TreeIdent(p.source); err == nil && id.Token != "" {
					ident = id.Ident()
				}
				if err := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "skipped", p.source, p.destination, ident, row.SkipReason)); err != nil {
					return report, err
				}
			}
			continue
		}
		if !apply {
			report.rows = append(report.rows, row)
			continue
		}
		// The apply: the row's outcome is the engine's verdict for this tree
		// (card t1640 review P2h — the placeholder never survives the run).
		out := migrateOneTree(root, p, mlPath, trigger, journal)
		if errors.Is(out.Cause, hook.ErrStateLockWaitLapsed) {
			stateLockLapses++
		}
		row.Outcome, row.SkipReason = out.Outcome, out.Reason
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
		return report, fmt.Errorf("migration recovery or audit writes failed: %s", strings.Join(writeErrs, "; "))
	}
	if stateLockLapses >= hook.StateLockLapseBudget {
		return report, fmt.Errorf("migration stopped after %d state-lock lapse(s): %w", stateLockLapses, hook.ErrStateLockWaitLapsed)
	}
	if apply {
		legacy := filepath.Join(root, ".moai", "worktrees")
		entries, err := os.ReadDir(legacy)
		if err != nil && !os.IsNotExist(err) {
			return report, err
		}
		if err == nil && len(entries) == 0 {
			if err := RetireEmptyLegacyDir(root); err != nil {
				return report, err
			}
		}
	}
	return report, nil
}

// treeOutcome is migrateOneTree's verdict for one tree.
type treeOutcome struct {
	Outcome  string // moved | reversed | failed | skipped
	Reason   string
	Ident    string
	Manifest string // the pre-move manifest digest (the rollback's settled precondition)
	MLErr    error  // a persistence or reverse-metadata error requiring recovery
	Cause    error  // the original failure, retaining typed state-lock lapses
}

// migrateOneTree moves one tree through the journaled step order —
// manifest, journal-pending (F3: before the move), intent line (the
// instance token first), git worktree move, state re-key, card rebind,
// manifest comparison — and, on any failure, the one reverse order (move
// back, re-key back, rebind back). Every reverse move first compares the
// tree's identity with the journal's and refuses a mismatch with a
// `refused` line (REQ-WHR-017/018).
func migrateOneTree(root string, p migratePlan, mlPath, trigger string, journal *migrationJournal) treeOutcome {
	digest, err := captureManifest(p.source)
	if err != nil {
		return treeOutcome{Outcome: "failed", Reason: fmt.Sprintf("capture pre-move manifest: %v", err), Cause: err}
	}

	bindIdentity, err := homestate.CaptureWorktreeRebindIdentity(p.source)
	if err != nil {
		return treeOutcome{Outcome: "failed", Reason: fmt.Sprintf("capture pre-move binding identity: %v", err), Cause: err}
	}

	id, err := EnsureInstanceToken(p.source)
	if err != nil {
		reason := fmt.Sprintf("identity unavailable: %v", err)
		lerr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "skipped", p.source, p.destination, "", reason))
		return treeOutcome{Outcome: "skipped", Reason: reason, MLErr: lerr}
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

	if _, err := listGitOutput(root, "worktree", "move", p.source, p.destination); err != nil {
		cause := fmt.Errorf("git worktree move: %w", err)
		_, sourceErr := os.Lstat(p.source)
		if errors.Is(sourceErr, os.ErrNotExist) {
			// Git may rename the directory before failing to update its admin
			// backlink. Repair only the journal's exact tree, then reverse it.
			if repairErr := ensureMigrationRegistration(root, p.destination, p.source, id); repairErr == nil {
				return reverseAfterFailure(root, p, id, mlPath, trigger, digest, cause)
			} else {
				lineErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.source, p.destination, id.Ident(), repairErr.Error()))
				return treeOutcome{Outcome: "failed", Reason: cause.Error(), Ident: id.Ident(), Manifest: digest, Cause: cause, MLErr: errors.Join(repairErr, lineErr)}
			}
		}
		lineErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "failed", p.source, p.destination, id.Ident(), cause.Error()))
		return treeOutcome{Outcome: "failed", Reason: cause.Error(), Ident: id.Ident(), Manifest: digest, Cause: cause, MLErr: lineErr}
	}

	if rerr := hook.RekeyWorktreeEntry(root, p.source, p.destination); rerr != nil {
		if strings.Contains(rerr.Error(), "no entry keyed to") {
			// No state-file entry for the old path: the re-key is a no-op
			// (the tree was never registered in the state file).
		} else {
			return reverseAfterFailure(root, p, id, mlPath, trigger, digest, fmt.Errorf("re-key failed: %w", rerr))
		}
	}

	if rerr := rebindCardForPath(root, p.source, p.destination, bindIdentity); rerr != nil {
		return reverseAfterFailure(root, p, id, mlPath, trigger, digest, fmt.Errorf("card rebind failed: %w", rerr))
	}

	if after, err := captureManifest(p.destination); err != nil || after != digest {
		return reverseAfterFailure(root, p, id, mlPath, trigger, digest, fmt.Errorf("manifest mismatch after the move: %w", errors.Join(err, errors.New("mandatory fields must match"))))
	}

	if merr := appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "moved", p.source, p.destination, id.Ident())); merr != nil {
		// The move stands; the settling-line failure surfaces at the run's
		// exit (the intent line already names the tree, REQ-WHR-012b step 9).
		return treeOutcome{Outcome: "moved", Ident: id.Ident(), Manifest: digest, MLErr: merr}
	}
	return treeOutcome{Outcome: "moved", Ident: id.Ident(), Manifest: digest}
}

// reverseAfterFailure runs the one reverse order after a failed step and
// translates its result into the outcome (card t1640 review P1-4): the tree
// counts as reversed only when the reverse move SUCCEEDED — a failed
// reverse move leaves the tree standing at the destination with its re-key
// and rebind intact (they name the location the tree really sits at),
// records pending, and fails the tree.
func reverseAfterFailure(root string, p migratePlan, id TreeIdentity, mlPath, trigger, digest string, cause error) treeOutcome {
	// A lock acquisition lapse made no metadata changes. Reversing git is
	// sufficient in that case; waiting for the same lock again cannot undo
	// an applied step and would spend another budget unnecessarily.
	restoreMetadata := !errors.Is(cause, hook.ErrStateLockWaitLapsed)
	movedBack, reverseErr := reverseMigrationTree(root, p, id, mlPath, trigger, restoreMetadata)
	out := treeOutcome{Ident: id.Ident(), Manifest: digest, Cause: cause}
	reason := cause.Error()
	if movedBack && reverseErr == nil {
		out.Outcome, out.Reason = "reversed", reason
	} else {
		out.Outcome = "failed"
		if movedBack {
			out.Reason = fmt.Sprintf("%s; the tree returned to %s but reverse metadata or audit remains pending (%v)", reason, p.source, reverseErr)
		} else if reverseErr != nil {
			out.Reason = fmt.Sprintf("%s; reverse remains pending at %s (%v)", reason, p.destination, reverseErr)
		} else {
			out.Reason = fmt.Sprintf("%s; the reverse move failed, the tree stands at %s (pending)", reason, p.destination)
		}
	}
	out.MLErr = reverseErr
	return out
}

// reverseAfterFailure runs the one reverse order after a failed step and
// translates its result into the outcome (card t1640 review P1-4): the tree
// reverseGitMove is the injectable seam for the reverse `git worktree move`
// (the fault-injection point of AC-WHR-017a's reverse-step failure clause).
var reverseGitMove = func(root, from, to string) error {
	_, err := listGitOutput(root, "worktree", "move", from, to)
	return err
}

// journalWrite is the injectable seam for the journal file write (the
// fault-injection point of the REQ-WHR-017 journal clause).
var journalWrite = func(path string, data []byte) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".migration-journal-*.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	defer func() { _ = tmp.Close() }()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := atomicfile.Replace(tmp.Name(), path); err != nil {
		return err
	}
	// Flush the rename on platforms that support syncing directories.
	// Windows flushes the file above; its directory handles do not support Sync.
	if runtime.GOOS != "windows" {
		dir, err := os.Open(filepath.Dir(path))
		if err != nil {
			return err
		}
		defer func() { _ = dir.Close() }()
		return dir.Sync()
	}
	return nil
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
		report, engineErr := runMigrationEngine(root, !dryRun, "update")
		moved := report.moved
		// The update UI summarizes linked worktrees. Keep the primary
		// checkout's classification and durable skip audit in the shared
		// engine, without counting it as a skipped linked worktree.
		rows := make([]migrationRow, 0, len(report.rows))
		skipped := report.skipped
		for _, row := range report.rows {
			if row.Class == string(ClassPrimary) {
				if row.Outcome == "skipped" {
					skipped--
				}
				continue
			}
			rows = append(rows, row)
		}
		if dryRun {
			for _, r := range rows {
				switch r.Outcome {
				case "planned":
					_, _ = fmt.Fprintf(out, "Worktree migration planned %s -> %s\n", r.Source, r.Destination)
				case "skipped":
					_, _ = fmt.Fprintf(out, "Worktree migration skipped %s: %s\n", r.Source, r.SkipReason)
				}
			}
			return engineErr
		}
		for _, r := range rows {
			switch r.Outcome {
			case "skipped":
				_, _ = fmt.Fprintf(out, "Worktree migration skipped %s: %s\n", r.Source, r.SkipReason)
			case "failed", "reversed":
				// A reversed move is a tree that is NOT migrated: the
				// engine moved it back (AC-WHR-019(f)).
				_, _ = fmt.Fprintf(out, "Worktree migration failed %s: %s\n", r.Source, r.SkipReason)
			case "moved":
				_, _ = fmt.Fprintf(out, "Worktree migration moved %s -> %s\n", r.Source, r.Destination)
			}
		}
		_, _ = fmt.Fprintf(out, "Worktree migration: %d moved, %d skipped, %d failed\n",
			moved, skipped, report.failed+report.reversed)
		return engineErr
	}
	if dryRun {
		return run()
	}
	return withMigrationLock(root, run)
}

// m6ReverseMove runs the one reverse order — move back, re-key back, rebind
// back — after first comparing the tree's identity at the destination with
// the journal's ident and refusing a mismatch with a `refused` line
// (REQ-WHR-017/018, design.md D-12: only ident tells two trees apart). It
// reports whether the tree actually moved back (card t1640 review P1-4) and
// the move-list write error, if any.
func m6ReverseMove(root string, p migratePlan, id TreeIdentity, mlPath, trigger string) (bool, error) {
	return reverseMigrationTree(root, p, id, mlPath, trigger, true)
}

func reverseMigrationTree(root string, p migratePlan, id TreeIdentity, mlPath, trigger string, restoreMetadata bool) (bool, error) {
	observed, identErr := TreeIdent(p.destination)
	if identErr != nil || id.Token == "" || observed.Ident() != id.Ident() {
		refusal := fmt.Errorf("reverse migration identity is unavailable or differs at %s", p.destination)
		return false, errors.Join(refusal, appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "refused", p.destination, p.source, id.Ident(), refusal.Error())))
	}
	if err := ensureMigrationRegistration(root, p.destination, p.source, id); err != nil {
		lineErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(), err.Error()))
		return false, errors.Join(err, lineErr)
	}
	bindIdentity, err := homestate.CaptureWorktreeRebindIdentity(p.destination)
	if err != nil {
		return false, err
	}
	if rerr := reverseGitMove(root, p.destination, p.source); rerr != nil {
		// The tree stands at the destination: the re-key and the rebind stay
		// pointing at it — reverting them would name a location the tree
		// does not occupy. The pending line is the convergence record; it
		// names the destination as the line's OLD path (the orientation the
		// reverse order travels), and MovedBy reads both ends of a pending
		// line (card t1640 review F5).
		return false, errors.Join(rerr, appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(),
			fmt.Sprintf("reverse move failed: %v", rerr))))
	}
	if restoreMetadata {
		if rerr := reverseMigrationMetadata(root, p.destination, p.source, bindIdentity); rerr != nil {
			lineErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", p.destination, p.source, id.Ident(),
				fmt.Sprintf("reverse metadata failed: %v", rerr)))
			return true, errors.Join(rerr, lineErr)
		}
	}
	return true, appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "reversed", p.destination, p.source, id.Ident()))
}

func reverseMigrationMetadata(root, from, to string, bindIdentity homestate.WorktreeRebindIdentity) error {
	if err := hook.RekeyWorktreeEntryInverse(root, from, to); err != nil && !strings.Contains(err.Error(), "no entry keyed to") {
		return fmt.Errorf("reverse re-key failed: %w", err)
	}
	if err := rebindCardForPath(root, from, to, bindIdentity); err != nil {
		return fmt.Errorf("reverse card rebind failed: %w", err)
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
func rebindCardForPath(root, oldPath, newPath string, bindIdentity homestate.WorktreeRebindIdentity) error {
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		return err
	}
	if _, serr := os.Stat(dbPath); serr != nil {
		if errors.Is(serr, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("inspect factory store for rebind: %w", serr)
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
	if !ok {
		return nil
	}
	_, err = db.RebindCardWorktree(context.Background(), cardID, oldPath, newPath, "worktree-migrate", time.Now(), bindIdentity)
	return err
}

// runMigrateVerify reports each recorded move in the journals as intact:
// the destination exists AND carries the journal step's identity (D-12 —
// only ident tells two trees apart; card t1640 review P2e). A destination
// whose identity differs is a mismatch and fails the verification; a step
// that was never a completed move (failed, skipped, pending) is not a
// recorded move and is not verified.
func runMigrateVerify(cmd *cobra.Command, root string) error {
	mlDir := filepath.Join(root, ".moai", "state", "worktree-migrate")
	entries, err := os.ReadDir(mlDir)
	if err != nil {
		return fmt.Errorf("read the migration journals: %w", err)
	}
	intact, mismatched := 0, 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") || e.Name() == "move-list.jsonl" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(mlDir, e.Name()))
		if err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "mismatch\tjournal %s cannot be read: %v\n", e.Name(), err)
			mismatched++
			continue
		}
		steps, err := decodeMigrationJournal(data)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "mismatch\tjournal %s cannot be parsed: %v\n", e.Name(), err)
			mismatched++
			continue
		}
		for _, s := range steps {
			if s.Outcome != "" && s.Outcome != "moved" {
				continue
			}
			if _, serr := os.Stat(s.NewPath); serr != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "absent\t%s → %s\n", s.OldPath, s.NewPath)
				mismatched++
				continue
			}
			if s.Ident != "" {
				if ti, terr := TreeIdent(s.NewPath); terr != nil || ti.Ident() != s.Ident {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "mismatch\t%s → %s (the destination is not the tree the journal moved)\n", s.OldPath, s.NewPath)
					mismatched++
					continue
				}
			}
			if s.Manifest != "" {
				observed, err := captureManifest(s.NewPath)
				if err != nil || observed != s.Manifest {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "mismatch\t%s → %s (manifest mismatch; inspection error: %v)\n", s.OldPath, s.NewPath, err)
					mismatched++
					continue
				}
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "intact\t%s → %s\n", s.OldPath, s.NewPath)
			intact++
		}
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%d recorded move(s) intact, %d mismatched\n", intact, mismatched)
	if mismatched > 0 {
		return fmt.Errorf("verification found %d mismatched move(s)", mismatched)
	}
	return nil
}

// Decode the known journal envelope strictly: an unrelated JSON object,
// unsupported schema or truncated step list is never evidence of intact moves.
func decodeMigrationJournal(data []byte) ([]journalStep, error) {
	var journal struct {
		Steps   *[]journalStep `json:"steps"`
		Trigger string         `json:"trigger"`
		Created string         `json:"created"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&journal); err != nil {
		return nil, err
	}
	if journal.Steps == nil {
		return nil, fmt.Errorf("journal has no steps array")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("journal has trailing data: %v", err)
	}
	for _, step := range *journal.Steps {
		switch step.Outcome {
		case "", "pending", "moved", "failed", "skipped", "reversed":
		default:
			return nil, fmt.Errorf("unknown journal outcome %q", step.Outcome)
		}
	}
	return *journal.Steps, nil
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
		steps, err := decodeMigrationJournal(data)
		if err != nil {
			return fmt.Errorf("parse the journal: %w", err)
		}
		mlPath := moveListPath(root)
		restored, refused, failed := 0, 0, 0
		var problems []string
		for i := len(steps) - 1; i >= 0; i-- {
			s := steps[i]
			if s.Outcome == "reversed" || s.Outcome == "skipped" {
				continue
			}
			atSource := false
			if _, serr := os.Stat(s.NewPath); serr != nil {
				if !errors.Is(serr, os.ErrNotExist) {
					problems = append(problems, fmt.Sprintf("inspect %s: %v", s.NewPath, serr))
					failed++
					continue
				}
				// A previous reverse may have completed git's move and then
				// failed its metadata steps. Only the journal's exact instance
				// at the source permits resuming those steps without moving it.
				id, err := TreeIdent(s.OldPath)
				if err != nil || s.Ident == "" || id.Ident() != s.Ident {
					reason := "recorded tree is missing or its source identity differs"
					if lerr := appendMoveListLine(mlPath, m6MoveLineReason("refused", s.NewPath, s.OldPath, s.Ident, reason)); lerr != nil {
						return lerr
					}
					problems = append(problems, fmt.Sprintf("%s: %s", s.OldPath, reason))
					refused++
					continue
				}
				atSource = true
			}
			destIdent, terr := TreeIdent(s.NewPath)
			if !atSource && (terr != nil || (s.Ident != "" && destIdent.Ident() != s.Ident)) {
				if lerr := appendMoveListLine(mlPath, m6MoveLineReason("refused", s.NewPath, s.OldPath, s.Ident, "identity mismatch")); lerr != nil {
					return lerr
				}
				problems = append(problems, fmt.Sprintf("%s: identity mismatch; the tree stays where it is", s.NewPath))
				refused++
				continue
			}
			// The manifest-match precondition belongs to the ordinary
			// rollback of a settled move; a pending step reverses
			// unconditionally (REQ-WHR-018's fifth entry).
			if !atSource && s.Outcome == "moved" && s.Manifest != "" {
				observed, err := captureManifest(s.NewPath)
				if err != nil || observed != s.Manifest {
					if lerr := appendMoveListLine(mlPath, m6MoveLineReason("refused", s.NewPath, s.OldPath, s.Ident, "manifest mismatch")); lerr != nil {
						return lerr
					}
					problems = append(problems, fmt.Sprintf("%s: manifest mismatch; the step stays pending (inspection error: %v)", s.NewPath, err))
					refused++
					continue
				}
			}
			bindingPath := s.NewPath
			if atSource {
				bindingPath = s.OldPath
			}
			bindingIdent, bindErr := TreeIdent(bindingPath)
			if bindErr == nil {
				bindErr = ensureMigrationRegistration(root, bindingPath, s.OldPath, bindingIdent)
			}
			var bindIdentity homestate.WorktreeRebindIdentity
			if bindErr == nil {
				bindIdentity, bindErr = homestate.CaptureWorktreeRebindIdentity(bindingPath)
			}
			if bindErr != nil {
				if lerr := appendMoveListLine(mlPath, m6MoveLineReason("pending", s.NewPath, s.OldPath, s.Ident, bindErr.Error())); lerr != nil {
					return errors.Join(bindErr, lerr)
				}
				problems = append(problems, fmt.Sprintf("%s: recovery registration: %v", bindingPath, bindErr))
				failed++
				continue
			}
			if !atSource {
				if merr := reverseGitMove(root, s.NewPath, s.OldPath); merr != nil {
					if lerr := appendMoveListLine(mlPath, m6MoveLineReason("pending", s.NewPath, s.OldPath, s.Ident, fmt.Sprintf("reverse move failed: %v", merr))); lerr != nil {
						return lerr
					}
					problems = append(problems, fmt.Sprintf("%s: reverse move failed (%v)", s.NewPath, merr))
					failed++
					continue
				}
			}
			if rerr := reverseMigrationMetadata(root, s.NewPath, s.OldPath, bindIdentity); rerr != nil {
				if lerr := appendMoveListLine(mlPath, m6MoveLineReason("pending", s.NewPath, s.OldPath, s.Ident, rerr.Error())); lerr != nil {
					return errors.Join(rerr, lerr)
				}
				problems = append(problems, fmt.Sprintf("%s: %v", s.OldPath, rerr))
				failed++
				continue
			}
			if lerr := appendMoveListLine(mlPath, m6MoveLine("restored", s.NewPath, s.OldPath, s.Ident)); lerr != nil {
				return lerr
			}
			restored++
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "restored %d tree(s), %d refused, %d pending\n", restored, refused, failed)
		if len(problems) > 0 {
			return fmt.Errorf("rollback: %s", strings.Join(problems, "; "))
		}
		return nil
	})
}

// canonicalTreePath resolves symlinks for path comparisons (the same
// canonical form planLegacyWorktreeMigrationFor compares in).
func canonicalTreePath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

// listGitOutput runs git in dir and returns stdout.
func listGitOutput(dir string, args ...string) (string, error) {
	argv := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", argv...)
	cmd.Env = append(gitenv.Env(), "LC_ALL=C", "LANGUAGE=C")
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(exit.Stderr)))
		}
	}
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
// ignored-file count, reports file count) plus a content digest over every
// untracked and ignored file's bytes — the digest that detects a one-byte
// mutation in an ignored report file (REQ-WHR-017; card t1640 review P2f:
// a HEAD+status digest alone is blind to ignored content).
type treeManifest struct {
	HEAD          string
	Branch        string
	Unpushed      string
	Status        string
	Diff          string
	IgnoredCount  int
	ReportsCount  int
	ContentDigest string
}

// captureManifest digests the working state of the tree at path through the
// full AC-WHR-017a manifest.
func captureManifest(tree string) (string, error) {
	m, err := captureTreeManifest(tree)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// captureTreeManifest reads the AC-WHR-017a manifest fields of the tree at
// path. Every mandatory probe must succeed: two equally unreadable states
// are not evidence that a move preserved its content.
func captureTreeManifest(tree string) (treeManifest, error) {
	var m treeManifest
	var staged, unstaged, untrackedOutput, ignoredOutput string
	for _, field := range []struct {
		value *string
		args  []string
	}{
		{&m.HEAD, []string{"rev-parse", "HEAD"}},
		{&m.Branch, []string{"rev-parse", "--abbrev-ref", "HEAD"}},
		{&m.Unpushed, []string{"rev-list", "HEAD", "--not", "--remotes"}},
		{&m.Status, []string{"status", "--porcelain"}},
		{&staged, []string{"diff", "--cached"}},
		{&unstaged, []string{"diff"}},
		{&untrackedOutput, []string{"ls-files", "--others", "-z"}},
		{&ignoredOutput, []string{"ls-files", "--others", "--ignored", "--exclude-standard", "-z"}},
	} {
		out, err := listGitOutput(tree, field.args...)
		if err != nil {
			return m, fmt.Errorf("manifest probe %s: %w", strings.Join(field.args, " "), err)
		}
		*field.value = out
	}
	m.HEAD = strings.TrimSpace(m.HEAD)
	m.Branch = strings.TrimSpace(m.Branch)
	if m.HEAD == "" || m.Branch == "" {
		return m, errors.New("manifest HEAD and branch probes must not be empty")
	}
	// The tracked-change content (REQ-WHR-017's content digest over the
	// tracked-change diff — card t1640 review F7): a content mutation that
	// preserves the status shape must change the manifest. Both the
	// unstaged and the staged forms ride the manifest.
	m.Diff = staged + unstaged

	untracked, ignored := splitNul(untrackedOutput), splitNul(ignoredOutput)
	m.IgnoredCount = len(ignored)

	seen := map[string]bool{}
	var paths []string
	for _, p := range append(append([]string{}, untracked...), ignored...) {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		paths = append(paths, p)
	}
	sort.Strings(paths)
	digest := sha256.New()
	for _, p := range paths {
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
				target, rerr := os.Readlink(full)
				if rerr != nil {
					return m, fmt.Errorf("manifest link %s: %w", full, rerr)
				}
				_, _ = digest.Write([]byte("link:" + target))
			case info.Mode().IsRegular():
				file, rerr := os.Open(full)
				if rerr != nil {
					return m, fmt.Errorf("manifest file %s: %w", full, rerr)
				}
				_, copyErr := io.Copy(digest, file)
				if err := errors.Join(copyErr, file.Close()); err != nil {
					return m, fmt.Errorf("manifest file %s: %w", full, err)
				}
			default:
				_, _ = digest.Write([]byte("special:" + info.Mode().String()))
			}
		} else {
			return m, fmt.Errorf("manifest path %s: %w", full, lerr)
		}
		_, _ = digest.Write([]byte{0})
	}
	m.ContentDigest = hex.EncodeToString(digest.Sum(nil))

	reports := filepath.Join(tree, ".moai", "reports")
	err := filepath.WalkDir(reports, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == reports && errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		m.ReportsCount++
		return nil
	})
	return m, err
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
