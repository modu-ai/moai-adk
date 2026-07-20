// Package update hosts the decomposed moai-update pipeline (M3a-M3f,
// SPEC-CLI-TUX-V3-003). This file establishes the classification data model
// (REQ-TUX3-001/002) that is the single source of truth consumed by BOTH the
// Bubble Tea v2 preview table and the plain-text fallback.
package update

// ChangeClass is the single shared classification of a file affected by
// moai update (REQ-TUX3-001). Exactly one class applies per file; the value
// is consumed identically by the TUI preview table and the text fallback so
// the two surfaces never re-derive classification independently.
type ChangeClass int

const (
	// ClassAdd applies to a new file the update would introduce.
	ClassAdd ChangeClass = iota
	// ClassUpdate applies to an existing moai-managed file whose content changes.
	ClassUpdate
	// ClassPreserveUserOwned applies to a user-owned namespace file that the
	// update MUST preserve byte-identical (never overwrite, never delete). It is
	// derived from the SAME namespace-protection predicate the deploy stage
	// enforces (isUserOwnedNamespace family) — no parallel heuristic.
	ClassPreserveUserOwned
	// ClassConflict applies to a file the 3-way merge could not auto-resolve.
	ClassConflict
)

// String renders the human-readable label. The preserve class carries the
// "preserved (user-owned)" label surfaced in both the TUI table and the text
// fallback (REQ-TUX3-014).
func (c ChangeClass) String() string {
	switch c {
	case ClassAdd:
		return "add"
	case ClassUpdate:
		return "update"
	case ClassPreserveUserOwned:
		return "preserved (user-owned)"
	case ClassConflict:
		return "conflict"
	}
	return "unknown"
}

// FileClassification is the per-file classification result consumed by the
// preview renderer (TUI or fallback).
type FileClassification struct {
	RelPath string
	Class   ChangeClass
}

// UserOwnedPredicate is the namespace-protection predicate contract that the
// deploy stage enforces. The preview classifier consumes the SAME predicate
// (REQ-TUX3-002) — no parallel heuristic. At the M3a contract stage the
// predicate is injected by the caller (package cli passes its
// isUserOwnedNamespace); the M3d decomposition moves the predicate into this
// package directly, after which the injection becomes a same-package call.
type UserOwnedPredicate func(relPath string) bool

// Classify derives exactly one ChangeClass for a file from the deploy-stage
// inputs. Priority order (highest first — a user-owned file is ALWAYS
// preserved, regardless of whether it would otherwise add/update/conflict):
//  1. user-owned  → ClassPreserveUserOwned (byte-identical preservation)
//  2. conflict    → ClassConflict (3-way merge could not auto-resolve)
//  3. !exists     → ClassAdd (new file)
//  4. otherwise   → ClassUpdate (existing moai-managed file content change)
//
// The user-owned check delegates entirely to the injected predicate
// (REQ-TUX3-002 single source of truth). A nil predicate disables the
// preserve classification (used only by unit tests that isolate the
// add/update/conflict branches).
//
// @MX:ANCHOR: [AUTO] Classify is the single classification entry point consumed by preview table, text fallback, and deploy enforcement
// @MX:REASON: REQ-TUX3-001 single source of truth — every change-class consumer routes through this function; fan_in grows to >= 3 at M3c (preview) + M3e (fallback) + M3f (enforcement coherence)
func Classify(relPath string, exists, conflict bool, isUserOwned UserOwnedPredicate) ChangeClass {
	if isUserOwned != nil && isUserOwned(relPath) {
		return ClassPreserveUserOwned
	}
	if conflict {
		return ClassConflict
	}
	if !exists {
		return ClassAdd
	}
	return ClassUpdate
}
