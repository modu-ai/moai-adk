// inbox_lifecycle.go owns the lessons-inbox size lifecycle
// (SPEC-INBOX-DRAIN-GAP-001): a collector-side write-time size cap with
// bounded archive rotation, armed only when the LSEL drain-ownership marker
// (.moai/state/lsel/) is absent. On a machine where the local drain exists
// (the maintainer's LSEL curator, t259 lineage) the collector stands down and
// behavior is byte-identical to pre-SPEC (NFC-4).
//
// The rotation chain is the SINGLE implementation shared by the collector's
// append path and the `moai inbox` CLI verbs (plan.md §F M3 — the CLI must not
// fork the collector's rotation logic).
//
// Fail-open discipline (NFC-3 / REQ-IBX-009): every cap-check or rotation
// error is logged via slog.Warn and swallowed — a learning-loop write must
// never block the session.
package hook

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
)

// LessonsInboxPath returns the project-local inbox path
// (.moai/lessons-inbox.jsonl). Exported for the `moai inbox` CLI verbs —
// single path construction for both the collector and the CLI.
func LessonsInboxPath(root string) string {
	return filepath.Join(root, ".moai", "lessons-inbox.jsonl")
}

// LessonsInboxArchiveGens returns how many archive generations currently exist
// for the live inbox path. Exported for `moai inbox status` so the generation
// naming is derived in exactly one place (plan.md §G — no second derivation).
func LessonsInboxArchiveGens(inboxPath string) int {
	gens := 0
	for g := 1; g <= config.DefaultInboxArchiveGenerations; g++ {
		if _, err := os.Stat(inboxGenPath(inboxPath, g)); err == nil {
			gens++
		}
	}
	return gens
}

// lselStateDir is the LSEL drain-ownership marker directory — the state dir
// the drain itself creates (plan.md §G: NOT the skill directory; ownership is
// proven by drain activity, not by skill installation).
func lselStateDir(root string) string {
	return filepath.Join(root, ".moai", "state", "lsel")
}

// LSELDrainMarkerPresent reports whether the LSEL drain-ownership marker
// exists under root (REQ-IBX-002).
//
// @MX:ANCHOR: [AUTO] LSEL drain-ownership probe — the single NFC-4 contact point
// @MX:REASON: fan_in=3 (enforceInboxCap, moai inbox status, moai inbox drain); every ownership-regime decision funnels through this one read-only stat — a second probe implementation would fork the curator stand-down semantics.
//
// NFC-4 (curator-first precedence): this probe is the cap path's ONLY contact
// with the marker — exactly one os.Stat, read-only. It never creates, writes,
// truncates, or reads anything under .moai/state/lsel/ (no drain-offset.json
// contact, no marker creation). Any existing entry — file or directory —
// counts as present: a false "absent" would rotate on a curator machine (the
// dangerous direction), while a false "present" only preserves pre-SPEC
// behavior.
func LSELDrainMarkerPresent(root string) bool {
	_, err := os.Stat(lselStateDir(root))
	return err == nil
}

// InboxRotationStats describes one completed rotation.
type InboxRotationStats struct {
	// RotatedBytes is the size of the live file moved into generation .1.
	RotatedBytes int64
	// OldestEvicted reports whether a prior oldest generation was removed by
	// this rotation (the documented eviction policy, REQ-IBX-004 — not an
	// error path).
	OldestEvicted bool
}

// inboxGenPath returns the archive path for generation gen of inboxPath
// (lessons-inbox.jsonl.1, .2, ...). Generations are bounded by
// config.DefaultInboxArchiveGenerations; no code path forms a path beyond it,
// so archive deletion stays bounded to the numbered-generation pattern.
func inboxGenPath(inboxPath string, gen int) string {
	return fmt.Sprintf("%s.%d", inboxPath, gen)
}

// RotateLessonsInboxArchive rotates the live lessons-inbox into the bounded
// archive chain (REQ-IBX-003 / REQ-IBX-004): the oldest generation is removed
// first, each remaining generation shifts down one slot, and the live file
// becomes generation .1.
//
// Delete-then-rename at every link (NFC-5): os.Rename fails over an existing
// destination on Windows, so the only destination ever occupied — the oldest
// slot — is removed before the chain shifts; every other destination is
// vacated by the shift itself. The same order makes the chain idempotent
// against pre-era leftover archives (acceptance §B): they are absorbed, not
// errored on.
//
// This is the single rotation implementation (plan.md §F M3): the write-time
// cap and `moai inbox drain` both call it.
//
// A missing live file is an error (fs.ErrNotExist); the caller decides what
// that means — "nothing to rotate" for the CLI drain, log-and-continue for
// the fail-open cap path.
func RotateLessonsInboxArchive(inboxPath string) (InboxRotationStats, error) {
	var st InboxRotationStats
	info, err := os.Stat(inboxPath)
	if err != nil {
		return st, fmt.Errorf("stat live inbox: %w", err)
	}
	st.RotatedBytes = info.Size()

	retention := config.DefaultInboxArchiveGenerations
	oldest := inboxGenPath(inboxPath, retention)
	if _, err := os.Stat(oldest); err == nil {
		st.OldestEvicted = true
	}
	// Best-effort eviction; the renames below are the real gate — if the
	// removal failed, the shift onto the occupied slot errors and the caller
	// fails open.
	_ = os.Remove(oldest)
	// Shift gen g -> g+1, highest-first: each destination slot was vacated by
	// the removal above (oldest) or is free because this is the first rotation.
	for g := retention - 1; g >= 1; g-- {
		if err := os.Rename(inboxGenPath(inboxPath, g), inboxGenPath(inboxPath, g+1)); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // no generation g yet (first-ever rotation)
			}
			return st, fmt.Errorf("shift generation %d: %w", g, err)
		}
	}
	if err := os.Rename(inboxPath, inboxGenPath(inboxPath, 1)); err != nil {
		return st, fmt.Errorf("rotate live -> generation 1: %w", err)
	}
	return st, nil
}

// enforceInboxCap applies the write-time size cap ahead of an append
// (REQ-IBX-001..004). Steady-state cost is one stat (NFC-2): the marker is
// probed only after the live file is observed at or over the cap, and the
// boundary is inclusive ("at or over", REQ-IBX-003 / acceptance §B).
//
// Stand-down (REQ-IBX-002): marker present → no rotation, no trim — the local
// drain owns the inbox lifecycle on that machine (NFC-4).
//
// Fail-open (REQ-IBX-009): a failed rotation logs a warning and the append
// proceeds best-effort on the existing file.
func enforceInboxCap(root, inboxPath string) {
	info, err := os.Stat(inboxPath)
	if err != nil {
		return // absent inbox — nothing to cap
	}
	if info.Size() < config.DefaultInboxMaxBytes {
		return // under cap (inclusive boundary handled by >= below)
	}
	if LSELDrainMarkerPresent(root) {
		return // curator stand-down (REQ-IBX-002)
	}
	if _, err := RotateLessonsInboxArchive(inboxPath); err != nil {
		slog.Warn("failure observer: lessons-inbox rotation failed; appending to existing file",
			"error", err)
	}
}
