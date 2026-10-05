package cli

import (
	"os"
	"path/filepath"
	"sync"
)

// hookRuntimeLogRelPath is the `moai hook` path's log sink, relative to the
// resolved project root.
//
// The name avoids "diagnostic": that word already has an owner in this
// repository (lsphook.Diagnostic, canonicalHookDiagnostics) where it means an
// LSP diagnostic, and a second meaning on the same word costs every later
// reader a disambiguation.
//
// The .log extension is load-bearing too: PruneObservationLogs selects
// trace-*.jsonl by name, so a .jsonl sink risks colliding with that predicate.
const hookRuntimeLogRelPath = ".moai/logs/hook-runtime.log"

// hookSink is the lazy-open append writer behind the `moai hook` logging
// destination. It is the one thing this SPEC changes: the hook path used to
// resolve to io.Discard, so every record a hook emitted was lost.
//
// Three properties, each load-bearing:
//
//   - Lazy. The file is opened on the FIRST Write, never at construction. A
//     hook invocation that emits no admitted record therefore creates neither
//     the file nor .moai/logs/ (REQ-HDS-007) — hooks fire dozens to hundreds of
//     times a session, and a silent one must cost nothing.
//   - Fail-open. An unresolvable root, a failed MkdirAll, a failed OpenFile, or
//     a failed write all degrade to discarding. Write never reports an error
//     upward, so a hook cannot fail because its logging could not (REQ-HDS-003).
//   - Append-atomic per record. The file is opened O_APPEND and each record is
//     emitted in a SINGLE Write, so concurrent hook processes appending to the
//     same file do not interleave within a line (REQ-HDS-006). Line ORDER across
//     processes is deliberately not promised.
//
// Deliberately absent: any reference to the process's standard streams. A hook's
// stdout carries the JSON contract and its stderr is read by the Claude Code
// runtime, so neither is a legitimate destination on this path and this file
// names neither (REQ-HDS-002).
//
// No Close: a hook is a one-shot process, each record leaves in one write, and
// a flush barrier would only add a wait inside the hook's budget (REQ-HDS-008).
type hookSink struct {
	root string
	once sync.Once
	file *os.File
}

// newHookSink returns a sink rooted at root. An empty root is accepted and
// degrades to discarding on first write rather than failing here — resolution
// failure is a runtime condition, not a construction error.
func newHookSink(root string) *hookSink {
	return &hookSink{root: root}
}

// Write appends p to the sink, opening it on first use.
//
// It always reports the full length and a nil error, including on every failure
// path. That is the fail-open contract, not sloppiness: slog surfaces a handler
// error to nothing a hook can act on, and a short count would make slog's caller
// believe the record mattered enough to retry.
func (s *hookSink) Write(p []byte) (int, error) {
	s.once.Do(s.open)
	if s.file == nil {
		return len(p), nil
	}
	// The error is intentionally dropped: a sink that cannot be written is the
	// same degraded state as a sink that could not be opened.
	_, _ = s.file.Write(p)
	return len(p), nil
}

// open resolves the sink path and opens it for append, leaving s.file nil on
// any failure so every later Write discards.
func (s *hookSink) open() {
	if s.root == "" {
		return
	}
	path := filepath.Clean(filepath.Join(s.root, hookRuntimeLogRelPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	s.file = f
}
