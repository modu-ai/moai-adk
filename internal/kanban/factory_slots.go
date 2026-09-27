// factory_slots.go — the factory lane registry's shared cluster
// (SPEC-FACTORY-WORKER-FANOUT-001 REQ-FF-004, moved here for the t85 lead
// loop).
//
// Through v1 the registry lived as package-private symbols in
// internal/cli/factory.go, which was fine while the launcher was its only
// reader. The leader loop needs the same read — which lane slots are FREE
// right now — from the SessionStart hook that renders the leader notice, and
// internal/hook cannot import internal/cli (the cli package imports hook for
// the `moai hook` subcommand), so the cluster moved to this package: the one
// that already owns the worker-label vocabulary (FactoryLaneLabel) and the
// backlog store the loop polls. The cli call sites keep their historical
// package-private names via thin delegates.
//
// The registry is persisted in the project-scoped factory.db. A legacy
// workers.json is imported only when the SQLite roster is empty.
//
// The liveness probe is passed in as a parameter rather than read from a
// package var so each consumer keeps its own test seam: cli overrides its
// factoryProcessAlive var and hands it through, and the hook uses
// FactoryProcessAlive directly.
package kanban

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// FactoryLaneEntry is one registered lane: the pid of the process that
// claimed the label. Because the launcher exec's into the backend without
// forking, the recorded pid IS the session's pid for the process's whole
// lifetime, which is what makes kill -0 a valid liveness probe for it.
type FactoryLaneEntry struct {
	PID          int    `json:"pid"`
	RegisteredAt string `json:"registered_at"`
}

// FactoryRegistryPath returns the project-scoped factory.db path. Separate
// projects keep separate registries and one project's runs share one roster.
func FactoryRegistryPath(root string) string {
	_ = homestate.EnsureProjectLayout(root)
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return filepath.Join(root, ".moai", "state", "factory", "factory.db")
	}
	return path
}

// LoadFactoryRegistry reads the lane registry, returning an empty map on
// any failure (missing file, unwritable dir, malformed JSON) — fail-open, so
// an unreadable registry reads as "every slot free" rather than blocking the
// leader loop's slot pick.
func LoadFactoryRegistry(path string) map[string]FactoryLaneEntry {
	reg := make(map[string]FactoryLaneEntry)
	db, err := homestate.OpenFactoryPath(path)
	if err != nil {
		return reg
	}
	defer func() { _ = db.Close() }()
	if root, rootErr := homestate.ProjectRootFromDBPath(path); rootErr == nil {
		_ = db.ImportLegacyWorkers(filepath.Join(root, ".moai", "state", "factory", "workers.json"))
	}
	rows, err := db.DB.Query(`SELECT label, pid, registered_at FROM workers`)
	if err != nil {
		return reg
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var label string
		var entry FactoryLaneEntry
		if err := rows.Scan(&label, &entry.PID, &entry.RegisteredAt); err == nil {
			reg[label] = entry
		}
	}
	return reg
}

// SaveFactoryRegistry writes the lane registry, creating its directory as
// needed. Best-effort: the error is returned for the caller to ignore.
func SaveFactoryRegistry(path string, reg map[string]FactoryLaneEntry) error {
	db, err := homestate.OpenFactoryPath(path)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM workers`); err != nil {
		return err
	}
	for label, entry := range reg {
		at := entry.RegisteredAt
		if at == "" {
			at = time.Now().UTC().Format(time.RFC3339Nano)
		}
		if _, err := tx.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES(?,?,?,?)`, label, entry.PID, at, at); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ClaimFactoryLaneName is ClaimFactoryLane for an operator-typed number,
// returning only the recorded label.
func ClaimFactoryLaneName(root, requested string, pid int, runID string, alive func(int) bool) (string, error) {
	claim, err := ClaimFactoryLane(root, requested, false, pid, runID, alive)
	return claim.Label, err
}

// FactoryClaim is the outcome of a lane-label claim.
type FactoryClaim struct {
	Label string // the canonical lane-<n> label recorded for pid
}

// FactoryLegacyRunError reports a join refused because the run already holds
// a LIVE record in the legacy vocabulary (written by a pre-rename binary).
// Adopting or numbering around it would mix two vocabularies in one run; the
// run must be retired and relaunched (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-022).
// The message names the legacy value, the run id, and the retire step.
type FactoryLegacyRunError struct {
	Label string // the live legacy label the run holds
	RunID string // the run the legacy record belongs to
}

func (e *FactoryLegacyRunError) Error() string {
	return e.Label + ": factory run " + e.RunID + " holds a live record from a binary before the leader/lane rename — " +
		"end its sessions, retire with 'moai factory runs --retire " + e.RunID + "', then relaunch"
}

// ClaimFactoryLane atomically removes dead claims, selects a lane number,
// and records pid under the canonical `lane-<n>` label, stamped with the
// joining run's id (runID — the membership datum the same-run refusal reads
// back). The selection and insert share one IMMEDIATE SQLite transaction, so
// two launchers cannot both observe the same free label and then erase each
// other's claim.
//
// Only the canonical `lane-<n>` shape is accepted (requested empty means
// auto-select the next free number); a legacy request is refused with the
// canonical name. A LIVE legacy record (`worker-<n>` / `agent-<n>`) stamped
// with a run id refuses the whole claim — *FactoryLegacyRunError names the
// legacy value, the run, and the retire step (REQ-RNC-022). A legacy row
// whose run_id is empty (a legacy import) belongs to no run: it is neither
// refused nor counted as a lane, and it is never rewritten. Dead claims of
// any shape are pruned as stale.
func ClaimFactoryLane(root, requested string, auto bool, pid int, runID string, alive func(int) bool) (FactoryClaim, error) {
	claim := FactoryClaim{Label: requested}
	admissionLock, lockErr := homestate.AcquireAdmissionLock(root)
	if lockErr != nil {
		return claim, lockErr
	}
	defer func() { _ = admissionLock.Release() }()
	if err := homestate.CheckRuntimeAdmission(root); err != nil {
		return claim, err
	}
	n := 0
	if requested != "" {
		var ok bool
		n, ok = factoryLabelNumber(requested)
		if !ok {
			if legacyN, isLegacy := SplitFactoryLegacyLabel(requested); isLegacy {
				return claim, fmt.Errorf("legacy factory label %q is not accepted; use %q", requested, FactoryLaneLabel(legacyN))
			}
			return claim, fmt.Errorf("invalid factory lane label %q", requested)
		}
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return claim, err
	}
	defer func() { _ = db.Close() }()
	if err := db.ImportLegacyWorkers(filepath.Join(root, ".moai", "state", "factory", "workers.json")); err != nil {
		return claim, err
	}

	tx, err := db.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return claim, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.Query(`SELECT label,pid,run_id FROM workers`)
	if err != nil {
		return claim, err
	}
	type row struct {
		label string
		pid   int
		runID string
	}
	var stale []row
	// Live canonical claims mark their number taken; maxCanonical is the
	// highest live canonical number (the auto path's next is one past it).
	// A live legacy row with a run id refuses the claim outright; with an
	// empty run id it belongs to no run and is simply ignored.
	taken := map[int]bool{}
	maxCanonical := 0
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.label, &r.pid, &r.runID); err != nil {
			_ = rows.Close()
			return claim, err
		}
		if r.pid <= 0 || !alive(r.pid) {
			stale = append(stale, r)
			continue
		}
		if num, isLane := factoryLabelNumber(r.label); isLane {
			taken[num] = true
			if num > maxCanonical {
				maxCanonical = num
			}
			continue
		}
		if IsLegacyFactoryLabel(r.label) && r.runID != "" {
			_ = rows.Close()
			return claim, &FactoryLegacyRunError{Label: r.label, RunID: r.runID}
		}
	}
	if err := rows.Close(); err != nil {
		return claim, err
	}
	for _, r := range stale {
		if _, err := tx.Exec(`DELETE FROM workers WHERE label=? AND pid=?`, r.label, r.pid); err != nil {
			return claim, err
		}
	}

	if auto {
		n = maxCanonical + 1
	}
	for taken[n] {
		n++
	}
	// Every surviving row is live and was counted into taken, so the
	// canonical label for n is free by construction.
	final := FactoryLaneLabel(n)
	at := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at,run_id) VALUES(?,?,?,?,?)`, final, pid, at, at, runID); err != nil {
		return claim, err
	}
	if err := tx.Commit(); err != nil {
		return claim, err
	}
	claim.Label = final
	return claim, nil
}

// PruneFactoryDeadClaims drops claims whose pid is dead (or non-positive)
// from reg and returns it. Dead claims must neither block a number nor
// accumulate — a crashed or exited lane leaves a dead pid behind, and a
// dead claim frees the name so a relaunch reuses it instead of counting up
// forever. Mutates reg in place, matching the original inline loop in
// resolveFactoryLaneName.
func PruneFactoryDeadClaims(reg map[string]FactoryLaneEntry, alive func(int) bool) map[string]FactoryLaneEntry {
	for l, e := range reg {
		if e.PID <= 0 || !alive(e.PID) {
			delete(reg, l)
		}
	}
	return reg
}

// FactoryFreeSlots returns the FREE slot numbers among 1..lanes under root
// — the leader loop's picker input. A slot is free when its number
// has no live claim: absent from the registry, mapped to a non-positive pid,
// or mapped to a pid the probe reports dead (dead claims are pruned on the
// way through, same rule as the bump path). Fail-open on registry errors via
// LoadFactoryRegistry, so an unreadable registry reads as all-free.
func FactoryFreeSlots(root string, lanes int, alive func(int) bool) []int {
	reg := PruneFactoryDeadClaims(LoadFactoryRegistry(FactoryRegistryPath(root)), alive)
	// Pruning leaves only live claims; a live canonical `lane-<n>` claim
	// occupies its number. Legacy shapes hold no number (a live legacy record
	// refuses the join instead — REQ-RNC-022).
	taken := map[int]bool{}
	for label := range reg {
		if n, ok := factoryLabelNumber(label); ok {
			taken[n] = true
		}
	}
	free := make([]int, 0, lanes)
	for i := 1; i <= lanes; i++ {
		if !taken[i] {
			free = append(free, i)
		}
	}
	return free
}

// NewFactoryLaneEntry stamps a claim for the current process — the register
// step the launcher's name resolver performs before a lane session starts.
func NewFactoryLaneEntry() FactoryLaneEntry {
	return FactoryLaneEntry{
		PID:          os.Getpid(),
		RegisteredAt: time.Now().UTC().Format(time.RFC3339),
	}
}
