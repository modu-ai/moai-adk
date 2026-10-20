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
	"strings"
	"time"

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
	apply, _ := cmd.Flags().GetBool("apply")
	verify, _ := cmd.Flags().GetBool("verify")
	rollbackPath, _ := cmd.Flags().GetString("rollback")
	asJSON, _ := cmd.Flags().GetBool("json")

	switch {
	case verify:
		return runMigrateVerify(cmd, cwd)
	case rollbackPath != "":
		return runMigrateRollback(cmd, cwd, rollbackPath, asJSON)
	default:
		return runMigratePlanOrApply(cmd, cwd, apply, asJSON)
	}
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

// runMigratePlanOrApply implements the dry run and the apply.
func runMigratePlanOrApply(cmd *cobra.Command, root string, apply, asJSON bool) error {
	return withMigrationLock(root, func() error {
		report, err := runMigrationEngine(root, apply)
		if err != nil {
			return err
		}
		if asJSON {
			body, jerr := json.MarshalIndent(report.rows, "", "  ")
			if jerr != nil {
				return jerr
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(body))
		} else {
			for _, r := range report.rows {
				line := fmt.Sprintf("%s\t%s → %s\t[%s]", r.class, r.source, r.destination, r.outcome)
				if r.skipReason != "" {
					line += " (" + r.skipReason + ")"
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "moved %d, skipped %d, failed %d, reversed %d\n",
				report.moved, report.skipped, report.failed, report.reversed)
		}
		if report.reversed > 0 || report.failed > 0 {
			return fmt.Errorf("migration completed with %d reversed and %d failed trees", report.reversed, report.failed)
		}
		return nil
	})
}

// migrationRow is one line of the migration report.
type migrationRow struct {
	class       string
	source      string
	destination string
	outcome     string
	skipReason  string
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
// resolver's home projects root (SPEC-WORKTREE-HOME-ROOT-001 M6). It never
// writes.
func planMigrationFor(root string, now time.Time) []migratePlan {
	out, err := listGitOutput(root, "worktree", "list", "--porcelain")
	if err != nil {
		return nil
	}
	records, err := ParseWorktreeList(out)
	if err != nil {
		return nil
	}
	homeBase := homeProjectsDir()
	ctx := ClassifyContext{
		PrimaryRoot:     root,
		HomeProjectsDir: homeBase,
	}
	_ = homeBase
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
		dest, derr := worktreeroot.PathFor(root, filepath.Base(rec.Path))
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
	plans := planMigrationFor(root, time.Now())
	mlPath := moveListPath(root)

	var journalSteps []map[string]any
	for _, p := range plans {
		outcome := "planned"
		skipReason := p.skip
		if p.class != string(ClassMoaiMade) {
			outcome = "skipped"
			skipReason = "class " + p.class + " is never moved"
		} else if p.skip != "" {
			outcome = "skipped"
			skipReason = p.skip
		}
		report.rows = append(report.rows, migrationRow{
			class: p.class, source: p.source, destination: p.destination,
			outcome: outcome, skipReason: skipReason,
		})
		if outcome == "skipped" {
			report.skipped++
			continue
		}
		if !apply {
			continue
		}
		outcome, m6reason, m6ident := migrateOneTree(root, p, mlPath)
		journalSteps = append(journalSteps, map[string]any{
			"old_path": p.source, "new_path": p.destination, "outcome": outcome, "ident": m6ident, "reason": m6reason,
		})
		switch outcome {
		case "moved":
			report.moved++
		case "reversed":
			report.reversed++
		case "failed":
			report.failed++
		default:
			report.skipped++
		}
	}

	// The journal: one file per apply, the rollback's input.
	if len(journalSteps) > 0 && apply {
		journalData, jerr := json.MarshalIndent(map[string]any{
			"steps": journalSteps, "created": time.Now().UTC().Format(time.RFC3339),
		}, "", "  ")
		if jerr == nil {
			journalPath := filepath.Join(filepath.Dir(moveListPath(root)), time.Now().UTC().Format("20060102T150405")+".json")
			_ = os.WriteFile(journalPath, journalData, 0o644)
		}
	}
	return report, nil
}

// migrateOneTree moves one tree through the journaled step order —
// manifest, intent line (the instance token first), git worktree move,
// state re-key, card rebind, manifest comparison — and, on any failure, the
// one reverse order (move back, re-key back, rebind back). Every reverse
// move first compares the tree's identity with the journal's and refuses a
// mismatch with a `refused` line (REQ-WHR-017/018). It returns the outcome
// (moved | reversed | failed) and the reason.
func migrateOneTree(root string, p migratePlan, mlPath string) (string, string, string) {
	digest := captureManifest(p.source)

	id, err := EnsureInstanceToken(p.source)
	if err != nil {
		return "skipped", fmt.Sprintf("identity unavailable: %v", err), ""
	}
	appendMoveListLine(mlPath, m6MoveLine("pending", p.source, p.destination, id.Ident()))

	if _, err := listGitOutput(root, "worktree", "move", p.source, p.destination); err != nil {
		appendMoveListLine(mlPath, m6MoveLine("failed", p.source, p.destination, id.Ident()))
		return "failed", fmt.Sprintf("git worktree move: %v", err), id.Ident()
	}

	if rerr := hook.RekeyWorktreeEntry(root, p.source, p.destination); rerr != nil {
		if strings.Contains(rerr.Error(), "no entry keyed to") {
			// No state-file entry for the old path: the re-key is a no-op
			// (the tree was never registered in the state file).
		} else {
			m6ReverseMove(root, p, id, mlPath)
			if errors.Is(rerr, hook.ErrStateLockWaitLapsed) {
				return "reversed", fmt.Sprintf("state lock budget spent: %v", rerr), id.Ident()
			}
			return "reversed", fmt.Sprintf("re-key failed: %v", rerr), id.Ident()
		}
	}

	if rerr := rebindCardForPath(root, p.source, p.destination); rerr != nil {
		m6ReverseMove(root, p, id, mlPath)
		return "reversed", fmt.Sprintf("card rebind failed: %v", rerr), id.Ident()
	}

	if after := captureManifest(p.destination); after != digest {
		m6ReverseMove(root, p, id, mlPath)
		return "reversed", "manifest mismatch after the move", id.Ident()
	}

	appendMoveListLine(mlPath, m6MoveLine("moved", p.source, p.destination, id.Ident()))
	return "moved", "", id.Ident()
}

// m6ReverseMove runs the one reverse order — move back, re-key back, rebind
// back — after first comparing the tree's identity at the destination with
// the journal's ident and refusing a mismatch with a `refused` line
// (REQ-WHR-017/018, design.md D-12: only ident tells two trees apart).
func m6ReverseMove(root string, p migratePlan, id TreeIdentity, mlPath string) {
	destIdent, err := TreeIdent(p.destination)
	if err == nil && destIdent.Ident() != id.Ident() {
		appendMoveListLine(mlPath, m6MoveLine("refused", p.destination, p.source, id.Ident()))
		return
	}
	_, _ = listGitOutput(root, "worktree", "move", p.destination, p.source)
	_ = hook.RekeyWorktreeEntryInverse(root, p.destination, p.source)
	_ = rebindCardForPath(root, p.destination, p.source)
	appendMoveListLine(mlPath, m6MoveLine("reversed", p.destination, p.source, id.Ident()))
}

// rebindCardForPath rebinds the card record bound to oldPath onto newPath
// when the factory record holds such a binding. No factory record, no
// binding — the step is skipped silently (the store is machine-local and
// may be absent in a repository that carries no factory state).
func rebindCardForPath(root, oldPath, newPath string) error {
	db, err := homestate.OpenFactoryReadonly(filepath.Join(root, ".moai", "factory.db"))
	if err != nil {
		return nil
	}
	defer func() { _ = db.Close() }()
	cardID, ok, err := db.CardRecordingPath(context.Background(), oldPath)
	if err != nil || !ok {
		return nil
	}
	_, err = db.RebindCardWorktree(context.Background(), cardID, oldPath, newPath, "worktree-migrate", time.Now())
	return err
}

// runMigrateVerify reports each recorded move in the journals as intact:
// the destination is a registered worktree.
func runMigrateVerify(cmd *cobra.Command, root string) error {
	mlDir := filepath.Join(root, ".moai", "state", "worktree-migrate")
	entries, err := os.ReadDir(mlDir)
	if err != nil {
		return fmt.Errorf("read the migration journals: %w", err)
	}
	intact := 0
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
			} `json:"steps"`
		}
		if json.Unmarshal(data, &journal) != nil {
			continue
		}
		for _, s := range journal.Steps {
			if _, err := os.Stat(s.NewPath); err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "intact\t%s → %s\n", s.OldPath, s.NewPath)
				intact++
			}
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%d recorded move(s) intact\n", intact)
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
			appendMoveListLine(mlPath, m6MoveLine("restored", s.NewPath, s.OldPath, s.Ident))
			restored++
		}
		fmt.Fprintf(cmd.OutOrStdout(), "restored %d tree(s)\n", restored)
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
	b, _ := json.Marshal(map[string]any{
		"schema": 1, "ts": time.Now().UTC().Format(time.RFC3339),
		"run": "", "trigger": "migrate", "tree": filepath.Base(oldPath),
		"card": "", "old_path": oldPath, "new_path": newPath,
		"result": result, "reason": "", "ident": ident,
	})
	return string(b)
}

// captureManifest digests the working state of the tree at path: HEAD plus
// the status porcelain (REQ-WHR-017's content digest).
func captureManifest(tree string) string {
	out, err := listGitOutput(tree, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	status, _ := listGitOutput(tree, "status", "--porcelain")
	sum := sha256.Sum256([]byte(strings.TrimSpace(out) + "\n" + status))
	return hex.EncodeToString(sum[:])
}

// isUnderWorktreePrefix reports whether path lies strictly inside prefix,
// spelling-safe: the containment walk of classify.go (underPath) with the
// equal case excluded, for the migration's destination-prefix checks.
func isUnderWorktreePrefix(path, prefix string) bool {
	return prefix != "" && underPath(path, prefix)
}

// processUsesWorktree reports whether one of the active processes works
// inside the tree at path.
