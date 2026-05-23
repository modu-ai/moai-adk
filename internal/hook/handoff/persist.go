// Package handoff implements SPEC-V3R6-SESSION-HANDOFF-AUTO-001 session-end paste-ready resume persistence.
//
// The package provides a best-effort persistence helper that the SessionEnd hook
// invokes after its existing cleanup steps. When the orchestrator has emitted a
// paste-ready resume message during the session and left a pending file at
//
//	<projectDir>/.moai/state/session-handoff/pending.md
//
// PersistIfPending reads that file, validates frontmatter (sprint/spec/status/
// index_line + optional supersedes), writes the verbatim Markdown body to
// <memoryDir>/project_<sprint>_<spec>_<status>.md via atomic rename, prepends
// the index_line to <memoryDir>/MEMORY.md (with optional supersede marker), and
// removes the pending file on success.
//
// All failures are best-effort: slog.Warn is emitted with the prefix
// "session_end: handoff: " and the function always returns nil. The package
// does not invoke AskUserQuestion or write to stdout/stderr in any
// user-visible way (REQ-SHA-009 / AC-SHA-009).
package handoff

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

// PersistIfPending reads <projectDir>/.moai/state/session-handoff/pending.md,
// validates and persists it to <memoryDir>/project_<sprint>_<spec>_<status>.md
// plus prepends <index_line> to <memoryDir>/MEMORY.md, then removes the pending
// file. All failure paths are log-only (slog.Warn) and return nil per the
// best-effort contract documented at the package level.
//
// Absent pending file is a no-op: returns nil without slog records and without
// creating the pending directory (REQ-SHA-002).
//
// The ctx parameter is reserved for future use (e.g., to honor session-end
// timeout). Currently only file I/O is performed.
//
// sessionID is reserved for future use (e.g., logging context tags).
//
// projectDir is the absolute path to the project root (typically input.CWD
// or input.ProjectDir from the SessionEnd hook).
//
// memoryDir is the absolute path to the Claude Code memory directory for
// this project (resolved by the caller; see session_end.go resolveMemoryDir).
// If the directory does not exist the helper logs warn and returns nil; it
// MUST NOT create the directory because the project-hash directory is owned
// by Claude Code (§B.2 "Project-hash directory creation").
func PersistIfPending(ctx context.Context, sessionID, projectDir, memoryDir string) error {
	_ = ctx
	_ = sessionID

	pendingPath := pendingFilePath(projectDir)

	// REQ-SHA-002: absent pending file is a no-op.
	if _, err := os.Stat(pendingPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		slog.Warn("session_end: handoff: could not stat pending file",
			"path", pendingPath,
			"error", err,
		)
		return nil
	}

	// M2 fills in: read, parse, validate, atomic write, prepend, cleanup.
	return nil
}

// pendingFilePath returns the canonical pending-resume file location per
// REQ-SHA-001. The path is `<projectDir>/.moai/state/session-handoff/pending.md`
// and is the only path read by PersistIfPending.
func pendingFilePath(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")
}
