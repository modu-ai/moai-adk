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
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
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
		report, err := runMigrationEngine(root, apply)
		if err != nil {
			return err
		}
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
// readability skip cause; the default accepts (the update-time apply wires
// the same probe the old planner used).
var worktreeRegistryProbe = func(root, tree string) error {
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
func planMigrationFor(root string, now time.Time, apply bool) []migratePlan {
	out, err := listGitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil
	}
	records, err := ParseWorktreeList(out)
	if err != nil {
		return nil
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
	var plans []migratePlan
	var processCWDs []string
	var processCWDErr error
	processCWDsRead := false
	for _, rec := range records {
		if rec.Bare {
			continue
		}
		if !strings.Contains(filepath.ToSlash(rec.Path), "/.moai/worktrees/") &&
			!strings.Contains(filepath.ToSlash(rec.Path), "/.claude/worktrees/") {
			continue // primary and unrelated trees are not the migration's
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
	return plans
}

// migrationReport builders

// runMigrationEngine plans and (when apply is set) executes the migration.
// The dry run and the apply both list every registered tree with its class;
// only moai-made trees are moved. Non-moai-made trees are reported as
// skipped with their class as the reason.
func runMigrationEngine(root string, apply bool) (*migrationReport, error) {
	report := &migrationReport{}
	plans := planMigrationFor(root, time.Now(), apply)
	mlPath := moveListPath(root)

	var journalSteps []map[string]any
	var writeErrs []string
	for _, p := range plans {
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
			continue
		}
		if !apply {
			report.rows = append(report.rows, row)
			continue
		}
		// The apply: the row's outcome is the engine's verdict for this tree
		// (card t1640 review P2h — the placeholder never survives the run).
		var mlErr error
		var m6ident string
		row.Outcome, row.SkipReason, m6ident, mlErr = migrateOneTree(root, p, mlPath)
		journalSteps = append(journalSteps, map[string]any{
			"old_path": p.source, "new_path": p.destination, "outcome": row.Outcome,
			"ident": m6ident, "reason": row.SkipReason,
		})
		if mlErr != nil {
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
	}

	// The journal: one file per apply, the rollback's and --verify's input.
	// Each step carries the paths, the outcome, the tree's ident (D-12 —
	// only ident tells two trees apart) and the reason.
	if len(journalSteps) > 0 && apply {
		if jerr := os.MkdirAll(filepath.Dir(moveListPath(root)), 0o755); jerr == nil {
			journalData, jerr := json.MarshalIndent(map[string]any{
				"steps": journalSteps, "created": time.Now().UTC().Format(time.RFC3339),
			}, "", "  ")
			if jerr == nil {
				journalPath := filepath.Join(filepath.Dir(moveListPath(root)), time.Now().UTC().Format("20060102T150405")+".json")
				_ = os.WriteFile(journalPath, journalData, 0o644)
			}
		}
	}
	if len(writeErrs) > 0 {
		return report, fmt.Errorf("move-list writes failed: %s", strings.Join(writeErrs, "; "))
	}
	return report, nil
}

// migrateOneTree moves one tree through the journaled step order —
// manifest, intent line (the instance token first), git worktree move,
// state re-key, card rebind, manifest comparison — and, on any failure, the
// one reverse order (move back, re-key back, rebind back). Every reverse
// move first compares the tree's identity with the journal's and refuses a
// mismatch with a `refused` line (REQ-WHR-017/018). It returns the outcome
// (moved | reversed | failed), the reason, the tree's ident, and a
// move-list write error that outlives a successful move (card t1640 review
// P1-3: an intent-line failure aborts the move BEFORE anything moves).
func migrateOneTree(root string, p migratePlan, mlPath string) (string, string, string, error) {
	digest := captureManifest(p.source)

	id, err := EnsureInstanceToken(p.source)
	if err != nil {
		return "skipped", fmt.Sprintf("identity unavailable: %v", err), "", nil
	}
	// The intent line precedes the move (REQ-WHR-012b step 9): a write
	// failure here leaves the lane with no pointer, so the move must not
	// start (P1-3).
	if ierr := appendMoveListLine(mlPath, m6MoveLine("pending", p.source, p.destination, id.Ident())); ierr != nil {
		return "failed", fmt.Sprintf("intent line not recorded; moved nothing (%v)", ierr), id.Ident(), nil
	}

	if _, err := listGitOutput(root, "worktree", "move", p.source, p.destination); err != nil {
		_ = appendMoveListLine(mlPath, m6MoveLine("failed", p.source, p.destination, id.Ident()))
		return "failed", fmt.Sprintf("git worktree move: %v", err), id.Ident(), nil
	}

	if rerr := hook.RekeyWorktreeEntry(root, p.source, p.destination); rerr != nil {
		if strings.Contains(rerr.Error(), "no entry keyed to") {
			// No state-file entry for the old path: the re-key is a no-op
			// (the tree was never registered in the state file).
		} else {
			return reverseAfterFailure(root, p, id, mlPath, fmt.Sprintf("re-key failed: %v", rerr))
		}
	}

	if rerr := rebindCardForPath(root, p.source, p.destination); rerr != nil {
		return reverseAfterFailure(root, p, id, mlPath, fmt.Sprintf("card rebind failed: %v", rerr))
	}

	if after := captureManifest(p.destination); after != digest {
		return reverseAfterFailure(root, p, id, mlPath, "manifest mismatch after the move")
	}

	if merr := appendMoveListLine(mlPath, m6MoveLine("moved", p.source, p.destination, id.Ident())); merr != nil {
		// The move stands; the settling-line failure surfaces at the run's
		// exit (the intent line already names the tree, REQ-WHR-012b step 9).
		return "moved", "", id.Ident(), merr
	}
	return "moved", "", id.Ident(), nil
}

// reverseAfterFailure runs the one reverse order after a failed step and
// translates its result into the outcome (card t1640 review P1-4): the tree
// counts as reversed only when the reverse move SUCCEEDED — a failed
// reverse move leaves the tree standing at the destination with its re-key
// and rebind intact (they name the location the tree really sits at),
// records pending, and fails the tree.
func reverseAfterFailure(root string, p migratePlan, id TreeIdentity, mlPath, reason string) (string, string, string, error) {
	movedBack, mlErr := m6ReverseMove(root, p, id, mlPath)
	if movedBack {
		return "reversed", reason, id.Ident(), mlErr
	}
	if mlErr != nil {
		return "failed", fmt.Sprintf("%s; the reverse move failed and the pending line could not be written (%v)", reason, mlErr), id.Ident(), nil
	}
	return "failed", fmt.Sprintf("%s; the reverse move failed, the tree stands at %s (pending)", reason, p.destination), id.Ident(), nil
}

// reverseGitMove is the injectable seam for the reverse `git worktree move`
// (the fault-injection point of AC-WHR-017a's reverse-step failure clause).
var reverseGitMove = func(root, from, to string) error {
	_, err := listGitOutput(root, "worktree", "move", from, to)
	return err
}

// m6ReverseMove runs the one reverse order — move back, re-key back, rebind
// back — after first comparing the tree's identity at the destination with
// the journal's ident and refusing a mismatch with a `refused` line
// (REQ-WHR-017/018, design.md D-12: only ident tells two trees apart). It
// reports whether the tree actually moved back (card t1640 review P1-4) and
// the move-list write error, if any.
func m6ReverseMove(root string, p migratePlan, id TreeIdentity, mlPath string) (bool, error) {
	destIdent, err := TreeIdent(p.destination)
	if err == nil && destIdent.Ident() != id.Ident() {
		return false, appendMoveListLine(mlPath, m6MoveLine("refused", p.destination, p.source, id.Ident()))
	}
	if rerr := reverseGitMove(root, p.destination, p.source); rerr != nil {
		// The tree stands at the destination: the re-key and the rebind stay
		// pointing at it — reverting them would name a location the tree
		// does not occupy. The pending line is the convergence record.
		return false, appendMoveListLine(mlPath, m6MoveLineReason("pending", p.destination, p.source, id.Ident(),
			fmt.Sprintf("reverse move failed: %v", rerr)))
	}
	_ = hook.RekeyWorktreeEntryInverse(root, p.destination, p.source)
	_ = rebindCardForPath(root, p.destination, p.source)
	return true, appendMoveListLine(mlPath, m6MoveLine("reversed", p.destination, p.source, id.Ident()))
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
		return nil // no factory store in this repository: nothing to rebind
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
	_, err = db.RebindCardWorktree(context.Background(), cardID, oldPath, newPath, "worktree-migrate", time.Now())
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
			continue
		}
		var journal struct {
			Steps []struct {
				OldPath string `json:"old_path"`
				NewPath string `json:"new_path"`
				Outcome string `json:"outcome"`
				Ident   string `json:"ident"`
			} `json:"steps"`
		}
		if json.Unmarshal(data, &journal) != nil {
			continue
		}
		for _, s := range journal.Steps {
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

// runMigrateRollback reverses each journaled move: move back, re-key back,
// rebind back, restored lines appended. The identity check before every
// reverse move compares the tree's ident with the journal's (D-12).
func runMigrateRollback(cmd *cobra.Command, root, journalPath string, asJSON bool) error {
	return withMigrationLock(root, func() error {
		data, err := os.ReadFile(journalPath)
		if err != nil {
			return fmt.Errorf("read the journal: %w", err)
		}
		var journal struct {
			Steps []struct {
				OldPath string `json:"old_path"`
				NewPath string `json:"new_path"`
				Ident   string `json:"ident"`
			} `json:"steps"`
		}
		if err := json.Unmarshal(data, &journal); err != nil {
			return fmt.Errorf("parse the journal: %w", err)
		}
		mlPath := moveListPath(root)
		restored := 0
		for i := len(journal.Steps) - 1; i >= 0; i-- {
			s := journal.Steps[i]
			destIdent, terr := TreeIdent(s.NewPath)
			if terr != nil || destIdent.Ident() != s.Ident {
				continue
			}
			if _, merr := listGitOutput(root, "worktree", "move", s.NewPath, s.OldPath); merr != nil {
				return fmt.Errorf("reverse move %s → %s: %w", s.NewPath, s.OldPath, merr)
			}
			_ = hook.RekeyWorktreeEntryInverse(root, s.NewPath, s.OldPath)
			_ = rebindCardForPath(root, s.NewPath, s.OldPath)
			if lerr := appendMoveListLine(mlPath, m6MoveLine("restored", s.NewPath, s.OldPath, s.Ident)); lerr != nil {
				return lerr
			}
			restored++
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "restored %d tree(s)\n", restored)
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
	out, err := exec.Command("git", argv...).Output()
	return string(out), err
}

// m6MoveLine builds one move-list record in the D-11 field list.
func m6MoveLine(result, oldPath, newPath, ident string) string {
	return m6MoveLineReason(result, oldPath, newPath, ident, "")
}

// m6MoveLineReason is m6MoveLine with the D-11 reason field set.
func m6MoveLineReason(result, oldPath, newPath, ident, reason string) string {
	b, _ := json.Marshal(map[string]any{
		"schema": 1, "ts": time.Now().UTC().Format(time.RFC3339),
		"run": "", "trigger": "migrate", "tree": filepath.Base(oldPath),
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
	IgnoredCount  int
	ReportsCount  int
	ContentDigest string
}

// captureManifest digests the working state of the tree at path through the
// full AC-WHR-017a manifest.
func captureManifest(tree string) string {
	m := captureTreeManifest(tree)
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// captureTreeManifest reads the AC-WHR-017a manifest fields of the tree at
// path. Read failures degrade to zero values inside the manifest — the
// pre/post comparison then notices the change of shape, exactly as an
// unreadable field should fail a field-for-field comparison.
func captureTreeManifest(tree string) treeManifest {
	var m treeManifest
	m.HEAD, _ = listGitOutput(tree, "rev-parse", "HEAD")
	m.HEAD = strings.TrimSpace(m.HEAD)
	m.Branch, _ = listGitOutput(tree, "rev-parse", "--abbrev-ref", "HEAD")
	m.Branch = strings.TrimSpace(m.Branch)
	m.Unpushed, _ = listGitOutput(tree, "rev-list", "HEAD", "--not", "--remotes")
	m.Status, _ = listGitOutput(tree, "status", "--porcelain")

	var untracked, ignored []string
	if out, err := listGitOutput(tree, "ls-files", "--others", "-z"); err == nil {
		untracked = splitNul(out)
	}
	if out, err := listGitOutput(tree, "ls-files", "--others", "--ignored", "--exclude-standard", "-z"); err == nil {
		ignored = splitNul(out)
	}
	m.IgnoredCount = len(ignored)

	seen := map[string]bool{}
	var paths []string
	for _, p := range append(append([]string{}, untracked...), ignored...) {
		p = strings.TrimSpace(p)
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
		if body, err := os.ReadFile(filepath.Join(tree, filepath.FromSlash(p))); err == nil {
			_, _ = digest.Write(body)
		} else {
			_, _ = digest.Write([]byte("\x00unreadable"))
		}
		_, _ = digest.Write([]byte{0})
	}
	m.ContentDigest = hex.EncodeToString(digest.Sum(nil))

	reports := filepath.Join(tree, ".moai", "reports")
	_ = filepath.WalkDir(reports, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		m.ReportsCount++
		return nil
	})
	return m
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
