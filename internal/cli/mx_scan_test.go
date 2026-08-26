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

// TestNewMxScanCmd_RejectsExternalScanPath (CR round-3, t261 observed red):
// a --path resolving OUTSIDE the project root is REJECTED — an external scan
// would persist tags whose files the project-root-relative provenance
// inventory cannot cover.
func TestNewMxScanCmd_RejectsExternalScanPath(t *testing.T) {
	project := t.TempDir()
	outside := t.TempDir() // a second tree, by construction outside project
	if err := os.WriteFile(filepath.Join(outside, "ext.go"),
		[]byte("package ext\n\n// @MX:NOTE: external\nfunc E() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return project, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	cmd := newMxScanCmd()
	cmd.SetArgs([]string{"--path", outside})
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("external --path must be rejected, got success")
	}
	if !strings.Contains(err.Error(), "outside the project root") {
		t.Errorf("rejection must name the containment reason: %v", err)
	}
	// No sidecar was written into the project.
	if _, statErr := os.Stat(filepath.Join(project, ".moai", "state", "mx-index.json")); statErr == nil {
		t.Error("rejected scan must not write a sidecar")
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

// TestNewMxScanCmd_QuietSuppressesSummary confirms --quiet omits the per-kind summary.
func TestNewMxScanCmd_QuietSuppressesSummary(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sample.go"),
		[]byte("package main\n\n// @MX:NOTE: hi\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	cmd := newMxScanCmd()
	if err := cmd.Flags().Set("quiet", "true"); err != nil {
		t.Fatalf("set quiet: %v", err)
	}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute scan: %v", err)
	}

	if strings.Contains(out.String(), "by kind") {
		t.Fatalf("--quiet must suppress the 'by kind' summary, got: %q", out.String())
	}
}

// TestNewMxScanCmd_PathScansSubtree confirms --path scopes the scan to one subtree
// and does not pick up tags outside it.
func TestNewMxScanCmd_PathScansSubtree(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	// tag inside the scoped subtree — must appear in the index.
	if err := os.WriteFile(filepath.Join(sub, "a.go"),
		[]byte("package sub\n\n// @MX:NOTE: in-subtree\nfunc A() {}\n"), 0644); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	// tag OUTSIDE the subtree — must NOT be scanned.
	if err := os.WriteFile(filepath.Join(dir, "b.go"),
		[]byte("package main\n\n// @MX:NOTE: outside\nfunc B() {}\n"), 0644); err != nil {
		t.Fatalf("write b.go: %v", err)
	}

	orig := findProjectRootFn
	findProjectRootFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { findProjectRootFn = orig })

	cmd := newMxScanCmd()
	if err := cmd.Flags().Set("path", sub); err != nil {
		t.Fatalf("set path: %v", err)
	}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute scan: %v", err)
	}

	idx := filepath.Join(dir, ".moai", "state", "mx-index.json")
	data, err := os.ReadFile(idx)
	if err != nil {
		t.Fatalf("index not written: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "in-subtree") {
		t.Fatalf("scoped tag missing from index")
	}
	if strings.Contains(body, "outside") {
		t.Fatalf("--path scoped to sub, but outside tag leaked into index")
	}
}
