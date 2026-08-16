package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// graphFixtureProject builds a tiny project with one edge per layer under a
// temporary root and returns the root.
func graphFixtureProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	depDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatal(err)
	}
	depBody := "```mermaid\ngraph TD\n    cli[\"internal/cli\"]\n    config[\"internal/config\"]\n    cli --> config\n```\n"
	if err := os.WriteFile(filepath.Join(depDir, "dependencies.md"), []byte(depBody), 0o644); err != nil {
		t.Fatal(err)
	}

	srcDir := filepath.Join(root, "internal", "demo")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	goBody := "package demo\n\n// @MX:NOTE: [AUTO] demo\n// @MX:SPEC:SPEC-GRAPH-CLI-001\nfunc Demo() {}\n"
	if err := os.WriteFile(filepath.Join(srcDir, "demo.go"), []byte(goBody), 0o644); err != nil {
		t.Fatal(err)
	}

	specDir := filepath.Join(root, ".moai", "specs", "SPEC-GRAPH-CLI-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specBody := "---\nid: SPEC-GRAPH-CLI-001\ntitle: \"t\"\ndepends_on: [SPEC-GRAPH-CLI-DEP-001]\n---\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specBody), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestGraphBuildCmd_WritesEdgesJSONL(t *testing.T) {
	root := graphFixtureProject(t)

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"build", "--root", root})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("graph build: %v\noutput: %s", err, out.String())
	}

	defaultOut := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	data, err := os.ReadFile(defaultOut)
	if err != nil {
		t.Fatalf("default output missing: %v (cmd output: %s)", err, out.String())
	}
	body := string(data)
	for _, want := range []string{
		`"kind":"import"`,
		`"kind":"mx-spec"`,
		`"kind":"spec-depends"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("edges.jsonl missing %s edge; body:\n%s", want, body)
		}
	}
}

func TestGraphBuildCmd_CustomOutPath(t *testing.T) {
	root := graphFixtureProject(t)
	custom := filepath.Join(root, "custom", "edges.jsonl")

	cmd := newGraphCmd()
	cmd.SetArgs([]string{"build", "--root", root, "--out", custom})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("graph build --out: %v\noutput: %s", err, out.String())
	}
	if _, err := os.Stat(custom); err != nil {
		t.Fatalf("custom output missing: %v", err)
	}
}

// TestGraphCmd_NoAskUserQuestion mirrors the canonical static guard
// (internal/cli/worktree/new_test.go TestNew_NoAskUserQuestion): CLI code must
// never prompt — the orchestrator owns user interaction (C-HRA-008).
func TestGraphCmd_NoAskUserQuestion(t *testing.T) {
	data, err := os.ReadFile("graph.go")
	if err != nil {
		t.Skipf("graph.go not present yet: %v", err)
	}
	for _, banned := range []string{"AskUserQuestion", "mcp__askuser"} {
		if strings.Contains(string(data), banned) {
			t.Errorf("graph.go references %s — CLI must not prompt", banned)
		}
	}
}
