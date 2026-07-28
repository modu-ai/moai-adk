package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewMxScanCmd_NoAskUserQuestion enforces C-HRA-008 / REQ-PGN-012: CLI code
// runs in subagent context and MUST NOT call AskUserQuestion or mcp__askuser.
// Mirrors the canonical static guard in worktree/new_test.go.
func TestNewMxScanCmd_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("mx_scan.go")
	if err != nil {
		t.Fatalf("read mx_scan.go: %v", err)
	}
	for _, token := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(string(src), token) {
			t.Fatalf("mx_scan.go must not reference %s — CLI runs in subagent context (orchestrator owns user interaction)", token)
		}
	}
}

// TestNewMxScanCmd_Registered confirms the scan subcommand is wired under 'moai mx'.
func TestNewMxScanCmd_Registered(t *testing.T) {
	parent := newMxCmd()
	for _, c := range parent.Commands() {
		if c.Use == "scan" {
			return
		}
	}
	t.Fatal("moai mx must register a 'scan' subcommand")
}

// TestNewMxScanCmd_BuildsIndex drives the full command against a temp project:
// it scans a sample source file and confirms the sidecar index is written and
// reported. findProjectRootFn is overridden to isolate the scan root.
func TestNewMxScanCmd_BuildsIndex(t *testing.T) {
	dir := t.TempDir()

	// Sample source carrying one @MX tag.
	if err := os.WriteFile(filepath.Join(dir, "sample.go"),
		[]byte("package main\n\n// @MX:NOTE: hello world\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	cmd := newMxScanCmd()
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute scan: %v", err)
	}

	idx := filepath.Join(dir, ".moai", "state", "mx-index.json")
	if _, err := os.Stat(idx); err != nil {
		t.Fatalf("sidecar index not written at %s: %v", idx, err)
	}
	if !strings.Contains(out.String(), "wrote") {
		t.Fatalf("expected write confirmation in output, got: %q", out.String())
	}
}

// TestNewMxScanCmd_DryRunDoesNotWrite confirms --dry previews without persisting.
func TestNewMxScanCmd_DryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sample.go"),
		[]byte("package main\n\n// @MX:NOTE: hi\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	cmd := newMxScanCmd()
	if err := cmd.Flags().Set("dry", "true"); err != nil {
		t.Fatalf("set dry flag: %v", err)
	}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute dry scan: %v", err)
	}

	idx := filepath.Join(dir, ".moai", "state", "mx-index.json")
	if _, err := os.Stat(idx); !os.IsNotExist(err) {
		t.Fatalf("dry run must not write the index, got err=%v at %s", err, idx)
	}
	if !strings.Contains(out.String(), "DRY RUN") {
		t.Fatalf("expected DRY RUN marker, got: %q", out.String())
	}
}
