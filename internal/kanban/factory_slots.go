// factory_slots.go — the factory lane registry's shared cluster
// (SPEC-FACTORY-WORKER-FANOUT-001 REQ-FF-004, moved here for the t85 lead
// loop).
//
// Through v1 the registry lived as package-private symbols in
// internal/cli/factory.go, which was fine while the launcher was its only
// reader. The lead loop needs the same read — which lane slots are FREE
// right now — from the SessionStart hook that renders the lead notice, and
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

// FactoryWorkerEntry is one registered lane: the pid of the process that
// claimed the label. Because the launcher exec's into the backend without
// forking, the recorded pid IS the session's pid for the process's whole
// lifetime, which is what makes kill -0 a valid liveness probe for it.
type FactoryWorkerEntry struct {
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
// lead loop's slot pick.
func LoadFactoryRegistry(path string) map[string]FactoryWorkerEntry {
	reg := make(map[string]FactoryWorkerEntry)
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
		var entry FactoryWorkerEntry
		if err := rows.Scan(&label, &entry.PID, &entry.RegisteredAt); err == nil {
			reg[label] = entry
		}
	}
	return reg
}

// SaveFactoryRegistry writes the lane registry, creating its directory as
// needed. Best-effort: the error is returned for the caller to ignore.
func SaveFactoryRegistry(path string, reg map[string]FactoryWorkerEntry) error {
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

// ClaimFactoryWorkerName is ClaimFactoryWorker for an operator-typed number,
// returning only the recorded label.
func ClaimFactoryWorkerName(root, requested string, pid int, alive func(int) bool) (string, error) {
	claim, err := ClaimFactoryWorker(root, requested, false, pid, alive)
	return claim.Label, err
}

// FactoryClaim is the outcome of a worker-label claim.
type FactoryClaim struct {
	Label         string   // the canonical worker-<n> label recorded for pid
	SkippedLegacy []string // live legacy labels whose numbers the claim passed over
}

// FactoryLegacyCollisionError reports an explicit-number request whose
// number is held by a live legacy (`agent-<n>` / `lane-<n>`) row.
type FactoryLegacyCollisionError struct {
	Requested string // the canonical label that was asked for
	Held      string // the live legacy label holding the same number
}

func (e *FactoryLegacyCollisionError) Error() string {
	return e.Requested + " is held by legacy label " + e.Held
}

// ClaimFactoryWorker atomically removes dead claims, selects a worker number,
// and records pid under the canonical `worker-<n>` label. The selection and
// insert share one IMMEDIATE SQLite transaction, so two launchers cannot both
// observe the same free label and then erase each other's claim.
//
// Legacy `agent-<n>` / `lane-<n>` rows share the worker number space, and a
// collision with one is never silent:
//
//   - auto == false (the operator typed the number): when the requested
//     number is held by a live legacy row, the claim is refused with a
//     *FactoryLegacyCollisionError naming that row, and nothing is recorded.
//     A number held by a live canonical row is bumped as before.
//   - auto == true (the launcher chose the number, `-f worker`): the claim
//     never refuses; it reports in SkippedLegacy every live legacy row whose
//     number sits between the canonical-only next number and the final one —
//     the rows the unified numbering stepped over.
//
// On an explicit bump, legacy rows passed over during the bump are reported
// in SkippedLegacy too.
func ClaimFactoryWorker(root, requested string, auto bool, pid int, alive func(int) bool) (FactoryClaim, error) {
	claim := FactoryClaim{Label: requested}
	admissionLock, lockErr := homestate.AcquireAdmissionLock(root)
	if lockErr != nil {
		return claim, lockErr
	}
	defer func() { _ = admissionLock.Release() }()
	if err := homestate.CheckRuntimeAdmission(root); err != nil {
		return claim, err
	}
	// Any worker shape is accepted; the claim is always recorded under the
	// canonical `worker-<n>` label (a legacy `lane-<n>` / `agent-<n>`
	// request is a deprecated spelling of the same number).
	n, ok := factoryLabelNumber(requested)
	if !ok {
		return claim, fmt.Errorf("invalid factory worker label %q", requested)
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
	rows, err := tx.Query(`SELECT label,pid FROM workers`)
	if err != nil {
		return claim, err
	}
	type row struct {
		label string
		pid   int
	}
	var stale []row
	// Live claims by number: taken marks any worker shape, legacyAt names the
	// live legacy row holding a number, and maxCanonical is the highest live
	// canonical number (the auto path's "canonical-only next" is one past it).
	taken := map[int]bool{}
	legacyAt := map[int]string{}
	maxCanonical := 0
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.label, &r.pid); err != nil {
			_ = rows.Close()
			return claim, err
		}
		if r.pid <= 0 || !alive(r.pid) {
			stale = append(stale, r)
			continue
		}
		num, isWorker := factoryLabelNumber(r.label)
		if !isWorker {
			continue
		}
		taken[num] = true
		if IsLegacyFactoryLabel(r.label) {
			legacyAt[num] = r.label
		} else if num > maxCanonical {
			maxCanonical = num
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

	from := n
	if auto {
		from = maxCanonical + 1
	} else if held, isLegacy := legacyAt[n]; isLegacy {
		return claim, &FactoryLegacyCollisionError{Requested: FactoryLaneLabel(n), Held: held}
	}
	for taken[n] {
		n++
	}
	for i := from; i < n; i++ {
		if held, isLegacy := legacyAt[i]; isLegacy {
			claim.SkippedLegacy = append(claim.SkippedLegacy, held)
		}
	}
	// Every surviving row is live and was counted into taken, so the
	// canonical label for n is free by construction.
	final := FactoryLaneLabel(n)
	at := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES(?,?,?,?)`, final, pid, at, at); err != nil {
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
// resolveFactoryWorkerName.
func PruneFactoryDeadClaims(reg map[string]FactoryWorkerEntry, alive func(int) bool) map[string]FactoryWorkerEntry {
	for l, e := range reg {
		if e.PID <= 0 || !alive(e.PID) {
			delete(reg, l)
		}
	}
	return reg
}

// FactoryFreeSlots returns the FREE slot numbers among 1..workers under root
// — the lead loop's picker input. A slot is free when its number
// has no live claim: absent from the registry, mapped to a non-positive pid,
// or mapped to a pid the probe reports dead (dead claims are pruned on the
// way through, same rule as the bump path). Fail-open on registry errors via
// LoadFactoryRegistry, so an unreadable registry reads as all-free.
func FactoryFreeSlots(root string, workers int, alive func(int) bool) []int {
	reg := PruneFactoryDeadClaims(LoadFactoryRegistry(FactoryRegistryPath(root)), alive)
	// Pruning leaves only live claims; a live claim in any worker shape —
	// canonical or legacy — occupies its number.
	taken := map[int]bool{}
	for label := range reg {
		if n, ok := factoryLabelNumber(label); ok {
			taken[n] = true
		}
	}
	free := make([]int, 0, workers)
	for i := 1; i <= workers; i++ {
		if !taken[i] {
			free = append(free, i)
		}
	}
	return free
}

// NewFactoryWorkerEntry stamps a claim for the current process — the register
// step the launcher's name resolver performs before a lane session starts.
func NewFactoryWorkerEntry() FactoryWorkerEntry {
	return FactoryWorkerEntry{
		PID:          os.Getpid(),
		RegisteredAt: time.Now().UTC().Format(time.RFC3339),
	}
}
