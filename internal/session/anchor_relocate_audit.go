// anchor_relocate_audit.go — SPEC-SESSION-ANCHOR-ATTR-001 W2: relocation
// audit + ownership plausibility (REQ-SAA-003..005).
//
// Registry.RelocateSession rewrote an entry's cwd field last-writer-wins with
// no audit and no plausibility judgment; a wrong-tree relocation (the t1339
// incident shape) left no trace. Every cwd rewrite now appends an audit row
// to <project-root>/.moai/logs/anchor-relocation-audit.jsonl, and the
// target-tree git worktree lock ownership is classified per the REQ-SAA-004
// case table:
//
//	self-owned      lock pid == entry pid            → no flag
//	other live card lock pid differs, holder alive  → advisory flag
//	unreadable      lock reason unparsable           → no flag (fail-closed
//	                                                   "treated as anchored"
//	                                                   preserved)
//	registry-only   no lock, registry anchor only    → advisory flag
//
// The advisory default is the fail-open doctrine: a flag never blocks the
// relocation. Blocking is opt-in via RefuseFlagged (the caller resolves
// workflow.anchor_relocation_guard.enabled), and a refusal is recorded in the
// audit log before anything moves.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RelocationAuditLogName is the audit log file name under the project root's
// .moai/logs directory.
const RelocationAuditLogName = "anchor-relocation-audit.jsonl"

// ErrRelocationRefused is returned when the opt-in anchor relocation guard
// refused a flagged relocation (REQ-SAA-005). The caller (hook layer) logs it
// and leaves the registry untouched; it is an intended refusal, not a fault.
var ErrRelocationRefused = errors.New("session registry: relocation refused by anchor relocation guard")

// RelocationOwner names the REQ-SAA-004 ownership case recorded on an audit
// row.
type RelocationOwner string

const (
	// OwnerSelf — the target-tree lock was written by the same process that
	// owns the relocating session entry (lock pid == entry pid).
	OwnerSelf RelocationOwner = "self"
	// OwnerOtherLive — the lock holder is a different, still-live process
	// (another live card). Advisory-flagged.
	OwnerOtherLive RelocationOwner = "other-live"
	// OwnerUnreadable — the lock reason could not be parsed. The existing
	// fail-closed reading ("treated as anchored") is preserved and the case
	// never flags.
	OwnerUnreadable RelocationOwner = "unreadable"
	// OwnerRegistryOnly — the target tree carries no git worktree lock, so
	// ownership cannot be compared; anchor evidence exists only in the
	// registry. Advisory-flagged.
	OwnerRegistryOnly RelocationOwner = "registry-only"
	// OwnerNone — the target tree carries no lock and no other registry
	// entry is anchored in it.
	OwnerNone RelocationOwner = "none"
	// OwnerDeadHolder — the lock holder was positively established dead
	// (the only negative the lock source may assert). No flag: a stale lock
	// is not another live card.
	OwnerDeadHolder RelocationOwner = "dead-holder"
)

// RelocationAudit is one JSONL audit row (REQ-SAA-003).
type RelocationAudit struct {
	// Timestamp is the row's RFC3339 UTC occurrence time.
	Timestamp string `json:"timestamp"`
	// SessionID is the relocated session.
	SessionID string `json:"session_id"`
	// FromCwd / ToCwd are the previous and new registry cwd values.
	FromCwd string `json:"from_cwd"`
	ToCwd   string `json:"to_cwd"`
	// Trigger names the hook event that drove the relocation ("CwdChanged",
	// "PostToolUse", or "unspecified" for the legacy entry point).
	Trigger string `json:"trigger"`
	// Owner is the REQ-SAA-004 ownership case.
	Owner RelocationOwner `json:"owner"`
	// Flagged marks an advisory ownership implausibility (REQ-SAA-004).
	Flagged bool `json:"flagged"`
	// Refused marks an opt-in-guard refusal (REQ-SAA-005); the row records
	// the refusal and the registry was left untouched.
	Refused bool `json:"refused"`
	// HolderCardID / HolderPID carry the target-tree lock reason's identity
	// tokens when a lock was readable.
	HolderCardID string `json:"holder_card_id,omitempty"`
	HolderPID    int    `json:"holder_pid,omitempty"`
}

// RelocationOptions carries the REQ-SAA-004/005 evaluation inputs for
// RelocateSessionWithOptions. Zero fields degrade safely: an empty trigger
// reads "unspecified", a nil TargetLock with no anchored others reads
// OwnerNone, and RefuseFlagged false keeps the advisory default.
type RelocationOptions struct {
	// Trigger names the hook event driving the relocation.
	Trigger string
	// TargetLock is the target tree's git worktree lock state; nil when the
	// caller could not resolve one (no lock observable).
	TargetLock *LockInfo
	// AnchoredOthers counts OTHER live registry entries anchored inside the
	// target tree (the relocating session excluded).
	AnchoredOthers int
	// RefuseFlagged is the opt-in anchor relocation guard
	// (workflow.anchor_relocation_guard.enabled, REQ-SAA-005).
	RefuseFlagged bool
}

// RelocationAuditPath returns the audit log path for a project root. Exported
// for the audit-log readers (tests today, operators and sweep tooling later).
func RelocationAuditPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "logs", RelocationAuditLogName)
}

// CountAnchoredOthers counts entries other than excludeSessionID whose
// recorded cwd lies at or below treePath and whose owning process is still
// alive. The relocation caller (hook layer) feeds it the entries it already
// queried from the candidate registry, so the registry-only ownership case
// needs no extra registry read. Exported for that caller.
//
// @MX:NOTE: [AUTO] SPEC-SESSION-ANCHOR-ATTR-001 — feeds the REQ-SAA-004
// registry-only flag case; liveness uses the shared isProcessAlive probe.
func CountAnchoredOthers(entries []Entry, excludeSessionID, treePath string) int {
	n := 0
	for _, e := range entries {
		if e.SessionID == excludeSessionID {
			continue
		}
		if !pathWithinTree(e.CWD, treePath) {
			continue
		}
		if e.PID > 0 && isProcessAlive(e.PID) {
			n++
		}
	}
	return n
}

// ParseLockCardID extracts the card-id token from a stored git worktree lock
// reason (`claude session <card-id> (pid <n> ...)` / `moai codex session
// <card-id> (pid <n>)`). It is the (i) leg of the REQ-SAA-004 comparison; the
// pid leg is parseLockPID.
//
// @MX:NOTE: [AUTO] Anything unrecognised reports not-ok — recorded as the
// "unreadable" owner case, never as unlocked (fail-closed preserved).
func ParseLockCardID(reason string) (string, bool) {
	idx := strings.Index(reason, "session ")
	if idx < 0 {
		return "", false
	}
	rest := reason[idx+len("session "):]
	end := strings.IndexAny(rest, " (")
	if end < 0 {
		end = len(rest)
	}
	id := strings.TrimSpace(rest[:end])
	if id == "" {
		return "", false
	}
	return id, true
}

// RelocateSessionWithOptions performs the relocation under the W2 audit and
// ownership evaluation. RelocateSession delegates here with zero options, so
// EVERY cwd rewrite produces an audit row (plan M2) regardless of entry
// point. Fail-open: an audit-write failure is logged and never blocks the
// relocation (acceptance §D.1).
//
// @MX:NOTE: [AUTO] SPEC-SESSION-ANCHOR-ATTR-001 REQ-SAA-003/004/005 — the
// audited relocation entry point; the two-pass candidate ordering upstream
// (REQ-RAR-002/003/004) is untouched, and the advisory-vs-blocking boundary
// rides RefuseFlagged.
func (r *Registry) RelocateSessionWithOptions(sessionID, newCwd string, opts RelocationOptions) (RelocationAudit, error) {
	if sessionID == "" {
		return RelocationAudit{}, errors.New("session registry: sessionID cannot be empty")
	}
	if newCwd == "" {
		return RelocationAudit{}, errors.New("session registry: newCwd cannot be empty")
	}

	audit := RelocationAudit{
		Timestamp: r.clock.Now().UTC().Format(time.RFC3339),
		SessionID: sessionID,
		ToCwd:     newCwd,
		Trigger:   opts.Trigger,
	}
	if audit.Trigger == "" {
		audit.Trigger = "unspecified"
	}

	found := false
	refused := false
	err := r.withLock(func(entries []Entry) ([]Entry, error) {
		for i := range entries {
			if entries[i].SessionID != sessionID {
				continue
			}
			found = true
			audit.FromCwd = entries[i].CWD
			audit.Owner, audit.Flagged, audit.HolderCardID, audit.HolderPID =
				evaluateRelocationOwner(entries[i], opts)
			if audit.Flagged && opts.RefuseFlagged {
				audit.Refused = true
				refused = true
				return entries, nil // the registry stays untouched
			}
			entries[i].CWD = newCwd
			return entries, nil
		}
		return entries, nil // no such entry: no-op, no audit row
	})
	if err != nil {
		return audit, err
	}
	if !found {
		return audit, nil
	}

	// A same-tree relocation (from == to) still audits — but never flags
	// (acceptance §D.1 edge case).
	if audit.FromCwd == audit.ToCwd {
		audit.Flagged = false
	}

	appendRelocationAudit(r.path, audit)
	if refused {
		return audit, ErrRelocationRefused
	}
	return audit, nil
}

// evaluateRelocationOwner implements the REQ-SAA-004 case table. Returns the
// owner case, whether the row is advisory-flagged, and the lock reason's
// identity tokens.
func evaluateRelocationOwner(entry Entry, opts RelocationOptions) (RelocationOwner, bool, string, int) {
	if opts.TargetLock == nil || !opts.TargetLock.Locked {
		if opts.AnchoredOthers > 0 {
			return OwnerRegistryOnly, true, "", 0
		}
		return OwnerNone, false, "", 0
	}
	lock := *opts.TargetLock
	pid, ok := parseLockPID(lock.Reason)
	if !ok {
		// Unreadable lock: the existing fail-closed anchor reading stands and
		// the case never flags — blocking a relocation on an unparsable lock
		// would fence off live trees behind corrupted reasons.
		return OwnerUnreadable, false, "", 0
	}
	cardID, _ := ParseLockCardID(lock.Reason)
	if pid == entry.PID {
		// The same process wrote the lock and owns the session entry.
		return OwnerSelf, false, cardID, pid
	}
	if LockHolderConfirmedDead(lock) {
		return OwnerDeadHolder, false, cardID, pid
	}
	return OwnerOtherLive, true, cardID, pid
}

// appendRelocationAudit appends one audit row to the project's relocation
// audit log, deriving the project root from the registry path. Fail-open:
// any error is logged once and swallowed — the relocation must never block
// on an audit write (acceptance §D.1).
func appendRelocationAudit(registryPath string, audit RelocationAudit) {
	root := projectRootOfRegistry(registryPath)
	if root == "" {
		return
	}
	path := RelocationAuditPath(root)
	data, err := json.Marshal(audit)
	if err != nil {
		slog.Warn("relocation audit: marshal failed", "error", err)
		return
	}
	data = append(data, '\n')
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Warn(fmt.Sprintf("relocation audit: open failed: %v", err))
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(data); err != nil {
		slog.Warn(fmt.Sprintf("relocation audit: write failed: %v", err))
	}
}

// projectRootOfRegistry recovers the project root from a registry file path
// (<root>/.moai/state/active-sessions.json → <root>). A path that does not
// end in the canonical layout yields the empty string and the audit row is
// dropped — never guessed at.
func projectRootOfRegistry(registryPath string) string {
	const suffix = string(filepath.Separator) + DefaultRegistryPath
	if !strings.HasSuffix(registryPath, suffix) {
		return ""
	}
	return strings.TrimSuffix(registryPath, suffix)
}
