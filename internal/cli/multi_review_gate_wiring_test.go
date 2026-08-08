// Package cli — tests for the `moai hook multi-review-gate` CLI wiring
// (SPEC-AUDIT-MULTI-MODEL-001 M5, REQ-AMM-013 / AC-AMM-018 / AC-AMM-025).
//
// multi_review_gate_test.go pins the PURE gate logic (HandleMultiReviewGate).
// This file pins the surrounding WIRING that makes the gate runnable:
//   - the config-gate reader truth table (fail-CLOSED opt-in),
//   - the cobra RunE fail-OPEN contract (a broken stdin never traps Stop),
//   - the subcommand registration under `moai hook`,
//   - the 900s timeout constant living in internal/config/defaults.go
//     (hardcoding prevention — AC-AMM-025).
//
// These tests are written BEFORE the wiring exists (RED); they mirror the
// codex-review-gate wiring tests one-for-one so the two gates stay symmetric.
//
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/spf13/cobra"
)

// writeWorkflowYAML writes body to <dir>/.moai/config/sections/workflow.yaml and
// returns dir, so each truth-table case gets an isolated project root.
func writeWorkflowYAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if body == "" {
		return dir // caller wants the file absent
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	return dir
}

// --- AC-AMM-018: opt-in config gate, fail-CLOSED ---

// TestReadMultiReviewGateEnabled_TruthTable pins the fail-CLOSED truth table of
// the config reader. Every non-affirmative path (missing file, malformed YAML,
// absent block, explicit false) MUST read false so the distributed default is
// OFF — the BranchGuard / codex-review-gate opt-in precedent. Both accepted key
// paths are exercised: the canonical nested `multi.review_gate.enabled` (the
// MultiConfig struct in internal/config/types.go) and the flat
// `multi_review_gate.enabled` spelled by AC-AMM-018.
func TestReadMultiReviewGateEnabled_TruthTable(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"missing file", "", false},
		{"malformed yaml", "multi:\n\treview_gate: [oops\n", false},
		{"block absent", "codex:\n  review_gate:\n    enabled: true\n", false},
		{"nested explicit false", "multi:\n  review_gate:\n    enabled: false\n", false},
		{"nested explicit true", "multi:\n  review_gate:\n    enabled: true\n", true},
		{"flat explicit false", "multi_review_gate:\n  enabled: false\n", false},
		{"flat explicit true", "multi_review_gate:\n  enabled: true\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeWorkflowYAML(t, tc.body)
			if got := readMultiReviewGateEnabled(dir); got != tc.want {
				t.Errorf("readMultiReviewGateEnabled(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
	if readMultiReviewGateEnabled("") {
		t.Errorf("empty projectDir ⇒ want false (fail-CLOSED)")
	}
}

// --- AC-AMM-018: 900s timeout constant, hardcoding prevention (AC-AMM-025) ---

// TestDefaultMultiReviewGateTimeout pins the 900s override that replaces the
// moai-default 5s hook budget for this hook only. The value MUST live in
// internal/config/defaults.go (single source of truth) — an inline literal at a
// call site would violate the hardcoding-prevention rule.
func TestDefaultMultiReviewGateTimeout(t *testing.T) {
	if got := config.DefaultMultiReviewGateTimeout; got != 900*time.Second {
		t.Errorf("DefaultMultiReviewGateTimeout = %v, want 900s", got)
	}
}

// --- REQ-AMM-013: cobra RunE fail-OPEN contract ---

// newGateCmd builds a throwaway cobra command wired to the multi-review-gate
// RunE with in-memory stdin/stdout/stderr, so the fail-open paths are testable
// without touching the process streams.
func newGateCmd(stdin string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	cmd := &cobra.Command{Use: "multi-review-gate", RunE: runMultiReviewGate, SilenceUsage: true}
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	return cmd, &out, &errOut
}

// TestRunMultiReviewGate_InvalidStdinFailsOpen proves a malformed stdin payload
// emits an empty ALLOW ({}) and exits 0. The Stop pipeline must NEVER be trapped
// by a parse failure; the diagnostic goes to stderr instead.
func TestRunMultiReviewGate_InvalidStdinFailsOpen(t *testing.T) {
	cmd, out, errOut := newGateCmd("{not json")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("invalid stdin must not error (fail-open); got %v", err)
	}
	assertAllowJSON(t, out.String())
	if !strings.Contains(errOut.String(), "multi-review-gate:") {
		t.Errorf("stderr must carry a multi-review-gate diagnostic, got %q", errOut.String())
	}
}

// TestRunMultiReviewGate_EmptyStdinFailsOpen proves an empty stdin (no hook
// payload at all) also fail-opens rather than propagating a parse error.
func TestRunMultiReviewGate_EmptyStdinFailsOpen(t *testing.T) {
	cmd, out, _ := newGateCmd("")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("empty stdin must not error (fail-open); got %v", err)
	}
	assertAllowJSON(t, out.String())
}

// TestRunMultiReviewGate_HappyPathAllow proves the well-formed default path:
// the gate is opt-in and the temp project has no config, so the reader reads
// false and the handler ALLOWs. Exercises the full stdin → project-dir resolve
// → config read → handler → stdout chain.
func TestRunMultiReviewGate_HappyPathAllow(t *testing.T) {
	dir := writeWorkflowYAML(t, "multi:\n  review_gate:\n    enabled: false\n")
	payload, err := json.Marshal(map[string]any{
		"session_id":  "sess-happy",
		"project_dir": dir,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	cmd, out, _ := newGateCmd(string(payload))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("happy path must not error; got %v", err)
	}
	assertAllowJSON(t, out.String())
}

// assertAllowJSON asserts stdout carries a decodable HookOutput with no BLOCK
// decision (the ALLOW contract: an empty object).
func assertAllowJSON(t *testing.T, stdout string) {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &decoded); err != nil {
		t.Fatalf("stdout must be valid JSON, got %q (%v)", stdout, err)
	}
	if d, ok := decoded["decision"]; ok && d == "block" {
		t.Errorf("expected ALLOW, got BLOCK: %q", stdout)
	}
}

// --- Registration ---

// TestMultiReviewGate_SubcommandRegistered proves `moai hook multi-review-gate`
// is wired into the hook command tree (mirrors the codex-review-gate
// registration test).
func TestMultiReviewGate_SubcommandRegistered(t *testing.T) {
	for _, c := range hookCmd.Commands() {
		if c.Name() == "multi-review-gate" {
			return // found
		}
	}
	t.Errorf("subcommand 'multi-review-gate' not registered under `moai hook`")
}
