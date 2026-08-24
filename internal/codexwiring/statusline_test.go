package codexwiring

import (
	"strings"
	"testing"
)

// canonicalStatusLineTokens is the 29-token canonical identifier set from the
// Codex TUI source (openai/codex codex-rs/tui/src/bottom_pane/status_line_setup.rs,
// StatusLineItem enum, kebab-case serialization) — SPEC-CODEX-WIRING-001 §A.6.
var canonicalStatusLineTokens = []string{
	"model", "model-with-reasoning", "reasoning", "current-dir", "project-name",
	"hostname", "git-branch", "pull-request-number", "branch-changes", "run-state",
	"permissions", "approval-mode", "context-remaining", "context-used",
	"five-hour-limit", "weekly-limit", "codex-version", "context-window-size",
	"used-tokens", "total-input-tokens", "total-output-tokens", "thread-credits",
	"estimated-thread-cost", "thread-id", "fast-mode", "raw-output", "thread-title",
	"workspace-headline", "task-progress",
}

// parseOnlyStatusLineAliases are the 7 legacy parse-only alias tokens — they
// are accepted by Codex on parse for old-value compatibility but MUST NOT be
// emitted (SPEC §A.6: "발행에 쓰지 않는다").
var parseOnlyStatusLineAliases = []string{
	"model-name", "project", "project-root", "status", "approval",
	"context-usage", "session-id",
}

// TestStatusLineDefaultIsCanonicalFive verifies the default configuration is
// exactly the 5 canonical tokens the operator directive fixed (AC-CW-013a).
func TestStatusLineDefaultIsCanonicalFive(t *testing.T) {
	want := []string{"model-with-reasoning", "context-remaining", "git-branch", "current-dir", "thread-id"}
	got := DefaultStatusLine()
	if len(got) != len(want) {
		t.Fatalf("DefaultStatusLine() = %v (len %d), want %v (len %d)", got, len(got), want, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("DefaultStatusLine()[%d] = %q, want %q (order is part of the contract)", i, got[i], want[i])
		}
	}
}

// TestStatusLineDefaultSubsetOfAllowlist verifies default ⊆ allowlist — the
// emission source constant must never drift outside the validated set
// (AC-CW-013b, REQ-CW-013).
func TestStatusLineDefaultSubsetOfAllowlist(t *testing.T) {
	allowed := make(map[string]bool, len(StatusLineAllowlist))
	for _, tok := range StatusLineAllowlist {
		allowed[tok] = true
	}
	if len(allowed) != len(StatusLineAllowlist) {
		t.Fatalf("StatusLineAllowlist carries duplicate tokens (%d entries, %d unique)", len(StatusLineAllowlist), len(allowed))
	}
	for _, tok := range DefaultStatusLine() {
		if !allowed[tok] {
			t.Errorf("default token %q is not in statusLineAllowlist — Codex would drop or reject it", tok)
		}
	}
}

// TestStatusLineDefaultHasNoParseAliases verifies none of the 7 parse-only
// legacy aliases leaks into the emitted default (AC-CW-013c).
func TestStatusLineDefaultHasNoParseAliases(t *testing.T) {
	emitted := make(map[string]bool, len(DefaultStatusLine()))
	for _, tok := range DefaultStatusLine() {
		emitted[tok] = true
	}
	for _, alias := range parseOnlyStatusLineAliases {
		if emitted[alias] {
			t.Errorf("parse-only alias %q appears in the emitted default — canonical tokens only (SPEC §A.6)", alias)
		}
	}
}

// TestStatusLineAllowlistMatchesCanonicalSet pins the allowlist to exactly the
// 29 canonical tokens recorded in SPEC §A.6 — additions are an explicit
// judgment at documentation-refresh time, never blind upstream tracking.
func TestStatusLineAllowlistMatchesCanonicalSet(t *testing.T) {
	if len(StatusLineAllowlist) != len(canonicalStatusLineTokens) {
		t.Errorf("statusLineAllowlist has %d tokens, canonical set has %d (SPEC §A.6: 29)", len(StatusLineAllowlist), len(canonicalStatusLineTokens))
	}
	allowed := make(map[string]bool, len(StatusLineAllowlist))
	for _, tok := range StatusLineAllowlist {
		allowed[tok] = true
	}
	for _, tok := range canonicalStatusLineTokens {
		if !allowed[tok] {
			t.Errorf("canonical token %q missing from statusLineAllowlist", tok)
		}
	}
	for _, alias := range parseOnlyStatusLineAliases {
		if allowed[alias] {
			t.Errorf("parse-only alias %q must not be in statusLineAllowlist (canonical tokens only)", alias)
		}
	}
}

// TestStatusLineMergeBranchNoTuiAppendsAtEOF — merge branch (i): a config with
// no [tui] table gets a new [tui] section appended at EOF (REQ-CW-013).
func TestStatusLineMergeBranchNoTuiAppendsAtEOF(t *testing.T) {
	in := []byte("# user config\nmodel = \"gpt-5\"\n\n[mcp_servers.moai]\ncommand = \"moai\"\n")
	out := EnsureStatusLine(in)
	if !strings.HasPrefix(string(out), string(in)) {
		t.Errorf("branch (i) must append at EOF, byte-preserving everything before it:\nin:  %q\nout: %q", in, out)
	}
	if !strings.Contains(string(out), "[tui]") {
		t.Errorf("branch (i) must add a [tui] table:\n%s", out)
	}
	if !strings.Contains(string(out), statusLineDefaultTOML) {
		t.Errorf("branch (i) must add the default status_line assignment:\n%s", out)
	}
}

// TestStatusLineMergeBranchTuiWithoutKeyInsertsAfterHeader — merge branch
// (ii): a [tui] table without status_line gets the key inserted directly after
// the section header (REQ-CW-013).
func TestStatusLineMergeBranchTuiWithoutKeyInsertsAfterHeader(t *testing.T) {
	in := []byte("model = \"gpt-5\"\n\n[tui]\nnotifications = true\n")
	out := EnsureStatusLine(in)
	body := string(out)
	if !strings.Contains(body, "[tui]\n"+statusLineDefaultTOML) {
		t.Errorf("branch (ii) must insert status_line directly after the [tui] header:\n%s", out)
	}
	if !strings.Contains(body, "notifications = true") {
		t.Errorf("branch (ii) must preserve the [tui] table's other keys:\n%s", out)
	}
}

// TestStatusLineMergeBranchExistingKeyBytePreserved — merge branch (iii): an
// existing status_line key is user-owned and stays byte-identical
// (REQ-CW-013, AC-CW-013 second clause).
func TestStatusLineMergeBranchExistingKeyBytePreserved(t *testing.T) {
	in := []byte("[tui]\nstatus_line = [\"model\"]\nnotifications = true\n")
	out := EnsureStatusLine(in)
	if string(out) != string(in) {
		t.Errorf("branch (iii) must be a byte no-op:\nin:  %q\nout: %q", in, out)
	}
	if !strings.Contains(string(out), `status_line = ["model"]`) {
		t.Errorf("user status_line line altered:\n%s", out)
	}
}

// TestStatusLineMergeNestedAndForeignTablesNotFooled verifies detection
// precision: a nested table like [a.tui] or a status_line key under a foreign
// table is NOT the [tui].status_line surface (parsing-boundary hazard, §H).
func TestStatusLineMergeNestedAndForeignTablesNotFooled(t *testing.T) {
	// [profile.tui] is not the [tui] table; status_line under [other] is not
	// the tui key. The merge must still add a real [tui] table.
	in := []byte("[profile.tui]\nnotifications = true\n\n[other]\nstatus_line = [\"model\"]\n")
	out := EnsureStatusLine(in)
	body := string(out)
	if !strings.Contains(body, "\n[tui]\n"+statusLineDefaultTOML) && !strings.HasSuffix(body, "[tui]\n"+statusLineDefaultTOML) {
		t.Errorf("foreign/nested tables must not satisfy the [tui].status_line detection — a real [tui] table must be added:\n%s", out)
	}
}

// TestStatusLineMergeIdempotent verifies a second pass over merged output is a
// byte no-op (REQ-CW-006).
func TestStatusLineMergeIdempotent(t *testing.T) {
	once := EnsureStatusLine([]byte("model = \"gpt-5\"\n"))
	twice := EnsureStatusLine(once)
	if string(once) != string(twice) {
		t.Errorf("status_line merge not idempotent:\nonce:  %q\ntwice: %q", once, twice)
	}
}
