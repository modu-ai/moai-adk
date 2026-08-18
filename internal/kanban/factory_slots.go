// factory_slots.go — the factory worker registry's shared cluster
// (SPEC-FACTORY-WORKER-FANOUT-001 REQ-FF-004, moved here for the t85 lead
// loop).
//
// Through v1 the registry lived as package-private symbols in
// internal/cli/factory.go, which was fine while the launcher was its only
// reader. The lead loop needs the same read — which worker slots are FREE
// right now — from the SessionStart hook that renders the lead notice, and
// internal/hook cannot import internal/cli (the cli package imports hook for
// the `moai hook` subcommand), so the cluster moved to this package: the one
// that already owns the worker-label vocabulary (FactoryWorkerLabel) and the
// backlog store the loop polls. The cli call sites keep their historical
// package-private names via thin delegates.
//
// The liveness probe is passed in as a parameter rather than read from a
// package var so each consumer keeps its own test seam: cli overrides its
// factoryProcessAlive var and hands it through, and the hook uses
// FactoryProcessAlive directly.
package kanban

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// FactoryWorkerEntry is one registered worker: the pid of the process that
// claimed the label. Because the launcher exec's into the backend without
// forking, the recorded pid IS the session's pid for the process's whole
// lifetime, which is what makes kill -0 a valid liveness probe for it.
type FactoryWorkerEntry struct {
	PID          int    `json:"pid"`
	RegisteredAt string `json:"registered_at"`
}

// FactoryRegistryPath returns the liveness-checked worker-name registry's
// home. It lives under .moai/state/ beside the goal and kanban state, keyed
// by project root — separate projects keep separate registries, and one
// project's concurrent factory runs share one (which is the point: the bump
// exists to keep session names addressable).
func FactoryRegistryPath(root string) string {
	return filepath.Join(root, ".moai", "state", "factory", "workers.json")
}

// LoadFactoryRegistry reads the worker registry, returning an empty map on
// any failure (missing file, unwritable dir, malformed JSON) — fail-open, so
// an unreadable registry reads as "every slot free" rather than blocking the
// lead loop's slot pick.
func LoadFactoryRegistry(path string) map[string]FactoryWorkerEntry {
	reg := make(map[string]FactoryWorkerEntry)
	raw, err := os.ReadFile(path)
	if err != nil {
		return reg
	}
	_ = json.Unmarshal(raw, &reg)
	return reg
}

// SaveFactoryRegistry writes the worker registry, creating its directory as
// needed. Best-effort: the error is returned for the caller to ignore.
func SaveFactoryRegistry(path string, reg map[string]FactoryWorkerEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o600)
}

// PruneFactoryDeadClaims drops claims whose pid is dead (or non-positive)
// from reg and returns it. Dead claims must neither block a number nor
// accumulate — a crashed or exited worker leaves a dead pid behind, and a
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
// — the lead loop's picker input. A slot is free when its worker-<n> label
// has no live claim: absent from the registry, mapped to a non-positive pid,
// or mapped to a pid the probe reports dead (dead claims are pruned on the
// way through, same rule as the bump path). Fail-open on registry errors via
// LoadFactoryRegistry, so an unreadable registry reads as all-free.
func FactoryFreeSlots(root string, workers int, alive func(int) bool) []int {
	reg := PruneFactoryDeadClaims(LoadFactoryRegistry(FactoryRegistryPath(root)), alive)
	free := make([]int, 0, workers)
	for i := 1; i <= workers; i++ {
		if claim, taken := reg[FactoryWorkerLabel(i)]; !taken || claim.PID <= 0 || !alive(claim.PID) {
			free = append(free, i)
		}
	}
	return free
}

// NewFactoryWorkerEntry stamps a claim for the current process — the register
// step the launcher's name resolver performs before a worker session starts.
func NewFactoryWorkerEntry() FactoryWorkerEntry {
	return FactoryWorkerEntry{
		PID:          os.Getpid(),
		RegisteredAt: time.Now().UTC().Format(time.RFC3339),
	}
}
