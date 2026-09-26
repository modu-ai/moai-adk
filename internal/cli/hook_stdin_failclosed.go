package cli

// hook_stdin_failclosed.go — SPEC-HOOK-STDIN-FAILCLOSED-001. When a hook's
// stdin cannot be parsed, a decision-bearing event is answered with a
// fail-closed deny instead of the no-opinion object; observation events keep
// the default output. The seam stays in the CLI layer; nothing under
// internal/hook is modified.

import (
	"fmt"
	"io"
	"os"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

const (
	// stdinParseFailureCause is the fixed cause every fail-closed reason
	// carries (REQ-HSF-010 (b)).
	stdinParseFailureCause = "hook stdin could not be parsed as JSON"
	// stdinParseFailClosedDocID names the operator document that explains
	// recovery (REQ-HSF-010 (c)). Recovery steps live in that document, never
	// in the reason the model reads. The value is the project-relative path
	// init and update deploy the document to (the template mirror under
	// internal/template/templates/.moai/docs/), so the pointer resolves in a
	// user's project without any lookup rule.
	stdinParseFailClosedDocID = ".moai/docs/hook-stdin-fail-closed.md"

	// stdinParseFailClosedDiscardKey marks a parse-failure fail-closed record
	// in the adapter's sink, distinct from hookFaultDiscardKey.
	stdinParseFailClosedDiscardKey = "stdin-parse-fail-closed"
	// stdinParseExemptDiscardKey marks a (Codex, Stop) parse-failure
	// exemption record, distinct from both keys above.
	stdinParseExemptDiscardKey = "stdin-parse-exempt"
)

// stdinByteCounter counts the bytes a reader hands out, so the dispatcher can
// record how much stdin it saw without reading the payload twice or keeping
// any of it.
type stdinByteCounter struct {
	r io.Reader
	n int
}

func (c *stdinByteCounter) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

// harnessOf names the translation-table harness for the --harness mode.
func harnessOf(codex bool) codexadapter.Harness {
	if codex {
		return codexadapter.HarnessCodex
	}
	return codexadapter.HarnessClaude
}

// stdinParseFailClosedReason is the reason a fail-closed deny carries: the
// marker, the fixed cause, and the operator document identifier — nothing
// derived from the payload and no recovery steps (REQ-HSF-010, REQ-HSF-011).
func stdinParseFailClosedReason() string {
	return "fail-closed: " + stdinParseFailureCause + " (" + stdinParseFailClosedDocID + ")"
}

// @MX:WARN: [AUTO] a Stop that fails to parse is blocked under Claude on every turn — the loop is bounded only by the host's Stop block cap, measured on Claude Code 2.1.283 at the default cap: JSON decision:block + exit 0 behaves like exit 2 — the hook ran and blocked 9 times, then the turn ended; the launcher raises the cap to 200 for kanban/factory sessions and for sessions with an infinite goal armed at launch (launcher_blockcap_infinite.go); under Codex the same Stop is exempt because Codex was measured with no cap
// @MX:REASON: [AUTO] REQ-HSF-009 — a parse failure hides stop_hook_active, so no in-process guard can end a Stop loop; the Codex exemption is decided only by codexadapter.HostLacksStopBlockCap
// answerStdinParseFailure answers a hook invocation whose stdin could not be
// parsed, without dispatching. label names the invocation in the stderr
// warning ("<Event>" or "agent <action>"), stdinBytes is how much stdin the
// dispatcher read, and writeDefault writes the entry point's default output.
//
//   - observation event: the existing warning and default output (6a3603274);
//   - decision event where the host has no Stop block cap: the default output
//     plus an exemption line and record (REQ-HSF-012);
//   - any other decision event: a fail-closed deny rendered through the
//     translation table, a stderr line, and a record (REQ-HSF-001/003/004/008).
//
// Every path returns nil or a write error, so the process exits 0.
func answerStdinParseFailure(label string, event hook.EventType, codex bool, stdinBytes int, parseErr error, writeDefault func() error) error {
	if !codexadapter.IsDecisionBearing(event) {
		_, _ = fmt.Fprintf(os.Stderr, "moai hook %s: invalid stdin JSON (%v); emitting default output\n", label, parseErr)
		return writeDefault()
	}

	harness := harnessOf(codex)
	if codexadapter.HostLacksStopBlockCap(harness, event) {
		_, _ = fmt.Fprintf(os.Stderr, "moai hook %s: invalid stdin JSON (%v) on %s, harness %s; exempt from fail-closed: the host has no Stop block cap\n",
			label, parseErr, event, harness)
		recordStdinParseFailure(codexadapter.Discard{
			Event:         event,
			Key:           stdinParseExemptDiscardKey,
			ContentLength: stdinBytes,
			Reason:        "stdin parse failure on Stop answered with no opinion: the host has no Stop block cap",
		})
		return writeDefault()
	}

	_, _ = fmt.Fprintf(os.Stderr, "moai hook %s: invalid stdin JSON (%v) on %s, harness %s; answered fail-closed\n",
		label, parseErr, event, harness)
	return writeFailClosedDeny(harness, event, stdinParseFailClosedReason(), codexadapter.Discard{
		Event:         event,
		Key:           stdinParseFailClosedDiscardKey,
		ContentLength: stdinBytes,
		Reason:        "stdin parse failure on a decision-bearing event answered with a fail-closed deny",
	})
}

// writeFailClosedDeny renders the fatal_error translation of event for
// harness — Codex through TranslateCodex, Claude through the same table's
// HarnessClaude row — records it, and writes it to stdout. It mirrors
// writeCodexFailClosed, which stays the writer for dispatch faults with its
// own record key; this one serves stdin parse failures on both harnesses.
func writeFailClosedDeny(harness codexadapter.Harness, event hook.EventType, reason string, record codexadapter.Discard) error {
	var (
		out []byte
		err error
	)
	if harness == codexadapter.HarnessCodex {
		out, _, err = codexadapter.TranslateCodex(event, codexadapter.DecisionFatalError, reason)
	} else {
		row, ok := codexadapter.Lookup(harness, event, codexadapter.DecisionFatalError)
		if !ok {
			return fmt.Errorf("fail-closed for %s: no %s fatal_error translation row", event, harness)
		}
		out, err = codexadapter.Render(event, row.Outcome, reason)
	}
	if err != nil {
		return fmt.Errorf("fail-closed for %s on %s: %w", event, harness, err)
	}

	recordStdinParseFailure(record)

	if _, werr := os.Stdout.Write(append(out, '\n')); werr != nil {
		return fmt.Errorf("write fail-closed hook output: %w", werr)
	}
	return nil
}

// recordStdinParseFailure writes one record to the adapter's sink. A failed
// write is reported on stderr and never changes the hook's answer.
func recordStdinParseFailure(record codexadapter.Discard) {
	if err := codexadapter.RecordDiscards(resolveHookProjectRoot(), []codexadapter.Discard{record}, false, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "moai hook: record stdin parse failure: %v\n", err)
	}
}
