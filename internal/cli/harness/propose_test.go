// Package harness CLI propose subcommand tests.
// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 REQ-PGN-008..009 + AC-PGN-004.
package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPropose_DryRun_BaselineFixture verifies the structural-0 resolution at the
// CLI level (SPEC-HARNESS-EVO-PIPE-REPAIR-001 DoD-2): against the frozen 8-record
// current-data baseline fixture, the repaired mapper now emits 3 actionable
// candidates (the auto_update system events subagent_stop/session_stop/user_prompt;
// agent_invocation:Bash: stays excluded as observation tier). Pre-repair this
// fixture yielded 0 candidates + reason "no-actionable-patterns".
func TestPropose_DryRun_BaselineFixture(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join("..", "..", "harness", "proposalgen", "testdata", "tier-promotions-current-baseline.jsonl")
	if _, err := os.Stat(fixture); err != nil {
		t.Fatalf("baseline fixture missing: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), ".moai", "proposals")
	cmd := NewProposeCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"--dry-run",
		"--input", fixture,
		"--output-dir", outDir,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v, want nil", err)
	}

	var got struct {
		Proposals         []map[string]any `json:"proposals"`
		Reason            string           `json:"reason"`
		MalformedLines    int              `json:"malformed_lines"`
		EvaluatedPatterns int              `json:"evaluated_patterns"`
		AutoDelegate      bool             `json:"auto_delegate"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got.Proposals) != 3 {
		t.Errorf("proposals = %d, want 3 (auto_update system events now actionable)", len(got.Proposals))
	}
	if got.Reason != "ok" {
		t.Errorf("reason = %q, want %q", got.Reason, "ok")
	}
	if got.MalformedLines != 0 {
		t.Errorf("malformed_lines = %d, want 0", got.MalformedLines)
	}
	if got.EvaluatedPatterns != 4 {
		t.Errorf("evaluated_patterns = %d, want 4", got.EvaluatedPatterns)
	}
	if got.AutoDelegate {
		t.Errorf("auto_delegate = true, want false (no --auto flag set)")
	}

	// --dry-run: scaffolder must not write to disk even when candidates exist.
	if _, err := os.Stat(outDir); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf(".moai/proposals MUST NOT exist with --dry-run; stat err = %v", err)
	}
}

// TestPropose_DryRun_MissingInputFile verifies REQ-PGN-003: a missing
// tier-promotions.jsonl returns the empty/absent reason without error.
func TestPropose_DryRun_MissingInputFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	missing := filepath.Join(tmp, "no-such-file.jsonl")
	outDir := filepath.Join(tmp, ".moai", "proposals")

	cmd := NewProposeCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--dry-run", "--input", missing, "--output-dir", outDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v, want nil", err)
	}

	var got struct {
		Proposals      []map[string]any `json:"proposals"`
		Reason         string           `json:"reason"`
		MalformedLines int              `json:"malformed_lines"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	if !strings.Contains(got.Reason, "absent or empty") {
		t.Errorf("reason = %q, want substring 'absent or empty'", got.Reason)
	}
	if got.MalformedLines != 0 {
		t.Errorf("malformed_lines = %d, want 0", got.MalformedLines)
	}
}

// TestPropose_AutoFlagSetsAutoDelegateOnlyWithProposals verifies REQ-PGN-010:
// auto_delegate true only when --auto AND at least one proposal exists. This
// test exercises the zero-proposal branch: an all-observation-tier fixture
// yields 0 candidates (pre-actionable), so --auto alone must NOT set auto_delegate.
func TestPropose_AutoFlagSetsAutoDelegateOnlyWithProposals(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "tp.jsonl")
	// observation tier is pre-actionable → 0 candidates (post-repair contract).
	data := `{"ts":"2026-05-24T10:00:00Z","pattern_key":"agent_invocation:Bash:","from_tier":"","to_tier":"observation","observation_count":1,"confidence":1}
`
	if err := os.WriteFile(input, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	outDir := filepath.Join(tmp, ".moai", "proposals")

	cmd := NewProposeCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--auto", "--dry-run", "--input", input, "--output-dir", outDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		AutoDelegate bool             `json:"auto_delegate"`
		Proposals    []map[string]any `json:"proposals"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(got.Proposals) != 0 {
		t.Errorf("setup: expected 0 proposals on observation-tier fixture, got %d", len(got.Proposals))
	}
	if got.AutoDelegate {
		t.Errorf("auto_delegate = true, want false (no proposals to delegate)")
	}
}

// TestPropose_AutoFlagWithActionableData asserts that when at least one
// actionable proposal is present AND --auto is set, auto_delegate flips to
// true. Uses a synthetic fixture file in t.TempDir().
func TestPropose_AutoFlagWithActionableData(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "tp.jsonl")
	data := `{"ts":"2026-05-24T10:00:00Z","pattern_key":"moai_subcommand:/moai run:auth","from_tier":"heuristic","to_tier":"rule","observation_count":7,"confidence":0.85}
`
	if err := os.WriteFile(input, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	outDir := filepath.Join(tmp, ".moai", "proposals")

	cmd := NewProposeCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--auto", "--dry-run", "--input", input, "--output-dir", outDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		AutoDelegate      bool             `json:"auto_delegate"`
		Proposals         []map[string]any `json:"proposals"`
		EvaluatedPatterns int              `json:"evaluated_patterns"`
		Reason            string           `json:"reason"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(got.Proposals) != 1 {
		t.Fatalf("proposals = %d, want 1", len(got.Proposals))
	}
	if !got.AutoDelegate {
		t.Errorf("auto_delegate = false, want true (--auto + 1 actionable proposal)")
	}
	if got.EvaluatedPatterns != 1 {
		t.Errorf("evaluated_patterns = %d, want 1", got.EvaluatedPatterns)
	}
	if got.Reason != "ok" {
		t.Errorf("reason = %q, want %q", got.Reason, "ok")
	}

	// --dry-run: scaffolder must not write to disk.
	if _, err := os.Stat(outDir); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf(".moai/proposals MUST NOT exist with --dry-run; stat err = %v", err)
	}
}

// TestPropose_LimitTruncates verifies REQ-PGN-008 --limit semantics: when the
// candidate count exceeds --limit, the output is truncated to N entries
// sorted by Confidence descending.
func TestPropose_LimitTruncates(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "tp.jsonl")
	// 5 actionable candidates with distinct confidences and pattern keys
	// (EventType-prefixed keys + auto_update tier per the repaired mapper contract).
	const data = `{"ts":"2026-05-24T10:00:00Z","pattern_key":"moai_subcommand:a:m","from_tier":"rule","to_tier":"auto_update","observation_count":12,"confidence":0.71}
{"ts":"2026-05-24T10:00:00Z","pattern_key":"moai_subcommand:b:m","from_tier":"rule","to_tier":"auto_update","observation_count":12,"confidence":0.95}
{"ts":"2026-05-24T10:00:00Z","pattern_key":"moai_subcommand:c:m","from_tier":"rule","to_tier":"auto_update","observation_count":12,"confidence":0.80}
{"ts":"2026-05-24T10:00:00Z","pattern_key":"moai_subcommand:d:m","from_tier":"rule","to_tier":"auto_update","observation_count":12,"confidence":0.90}
{"ts":"2026-05-24T10:00:00Z","pattern_key":"moai_subcommand:e:m","from_tier":"rule","to_tier":"auto_update","observation_count":12,"confidence":0.75}
`
	if err := os.WriteFile(input, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	outDir := filepath.Join(tmp, ".moai", "proposals")

	cmd := NewProposeCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--dry-run", "--limit", "3", "--input", input, "--output-dir", outDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got struct {
		Proposals []map[string]any `json:"proposals"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(got.Proposals) != 3 {
		t.Fatalf("len(proposals) = %d, want 3", len(got.Proposals))
	}

	// Confirm sorted by Confidence desc: 0.95, 0.90, 0.80.
	want := []float64{0.95, 0.90, 0.80}
	for i, p := range got.Proposals {
		conf, ok := p["confidence"].(float64)
		if !ok {
			t.Errorf("proposals[%d].confidence is not a float64: %v", i, p["confidence"])
			continue
		}
		if conf != want[i] {
			t.Errorf("proposals[%d].confidence = %v, want %v", i, conf, want[i])
		}
	}
}

// TestPropose_WriteMode_CreatesFiles verifies the non-dry-run path: actual
// .moai/proposals/<draft-id>/ directories are created with spec.md +
// proposal.json.
func TestPropose_WriteMode_CreatesFiles(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "tp.jsonl")
	data := `{"ts":"2026-05-24T10:00:00Z","pattern_key":"moai_subcommand:/moai run:auth","from_tier":"heuristic","to_tier":"rule","observation_count":7,"confidence":0.85}
`
	if err := os.WriteFile(input, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	outDir := filepath.Join(tmp, ".moai", "proposals")

	cmd := NewProposeCmd()
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--input", input, "--output-dir", outDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Verify directory structure.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir outDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 draft dir, got %d", len(entries))
	}
	draftDir := filepath.Join(outDir, entries[0].Name())
	for _, fname := range []string{"spec.md", "proposal.json"} {
		if _, err := os.Stat(filepath.Join(draftDir, fname)); err != nil {
			t.Errorf("expected %s in %s; stat err = %v", fname, draftDir, err)
		}
	}
}

// TestPropose_FlagsRegistered ensures the CLI exposes all 4 documented flags.
func TestPropose_FlagsRegistered(t *testing.T) {
	t.Parallel()

	cmd := NewProposeCmd()
	required := []string{"auto", "dry-run", "limit", "input", "output-dir"}
	for _, name := range required {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag --%s not registered", name)
		}
	}
}
