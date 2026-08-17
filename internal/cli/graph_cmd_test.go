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

// graphQueryFixture extends graphFixtureProject with a spec.md for the
// depends_on target so the --specs-no-code universe has a known unreferenced
// member, then builds edges.jsonl under the temp root.
func graphQueryFixture(t *testing.T) string {
	t.Helper()
	root := graphFixtureProject(t)

	depDir := filepath.Join(root, ".moai", "specs", "SPEC-GRAPH-CLI-DEP-001")
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatal(err)
	}
	depBody := "---\nid: SPEC-GRAPH-CLI-DEP-001\ntitle: \"t\"\n---\n"
	if err := os.WriteFile(filepath.Join(depDir, "spec.md"), []byte(depBody), 0o644); err != nil {
		t.Fatal(err)
	}

	build := newGraphCmd()
	build.SetArgs([]string{"build", "--root", root})
	out := &strings.Builder{}
	build.SetOut(out)
	build.SetErr(out)
	if err := build.Execute(); err != nil {
		t.Fatalf("graph build: %v\noutput: %s", err, out.String())
	}
	return root
}

// runGraphQuery executes 'moai graph query' with args and returns the output.
func runGraphQuery(t *testing.T, args ...string) string {
	t.Helper()
	cmd := newGraphCmd()
	cmd.SetArgs(append([]string{"query"}, args...))
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("graph query %v: %v\noutput: %s", args, err, out.String())
	}
	return out.String()
}

func TestGraphQueryCmd_Callers(t *testing.T) {
	root := graphQueryFixture(t)
	out := runGraphQuery(t, "--root", root, "--callers", "SPEC-GRAPH-CLI-001")
	for _, want := range []string{"internal/demo/demo.go", "callers of SPEC-GRAPH-CLI-001: 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestGraphQueryCmd_Blast(t *testing.T) {
	root := graphQueryFixture(t)
	// DEP <- 001 (depends_on) <- demo.go (mx-spec): the radius spans both kinds.
	out := runGraphQuery(t, "--root", root, "--blast", "SPEC-GRAPH-CLI-DEP-001")
	for _, want := range []string{"SPEC-GRAPH-CLI-001", "internal/demo/demo.go"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestGraphQueryCmd_FanIn(t *testing.T) {
	root := graphQueryFixture(t)
	out := runGraphQuery(t, "--root", root, "--fanin", "--limit", "5")
	if !strings.Contains(out, "1\tinternal/config") {
		t.Errorf("output missing the ranked importer row:\n%s", out)
	}
}

func TestGraphQueryCmd_SpecsNoCodeIncludesCaveat(t *testing.T) {
	root := graphQueryFixture(t)
	out := runGraphQuery(t, "--root", root, "--specs-no-code")

	// The unreferenced member of the universe is the depends_on-only SPEC.
	for _, want := range []string{
		"SPECs with no code reference: 1 of 2",
		"SPEC-GRAPH-CLI-DEP-001",
		"미연결 ≠ 미구현", // [HARD] caveat must ride every --specs-no-code result
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestGraphQueryCmd_RequiresExactlyOneSelector(t *testing.T) {
	root := graphQueryFixture(t)

	for _, args := range [][]string{
		{"--root", root},
		{"--root", root, "--fanin", "--specs-no-code"},
	} {
		cmd := newGraphCmd()
		cmd.SetArgs(append([]string{"query"}, args...))
		out := &strings.Builder{}
		cmd.SetOut(out)
		cmd.SetErr(out)
		if err := cmd.Execute(); err == nil {
			t.Errorf("args %v must fail (exactly one selector required), output:\n%s", args, out.String())
		} else if !strings.Contains(err.Error(), "exactly one") {
			t.Errorf("args %v: unexpected error: %v", args, err)
		}
	}
}

func TestGraphQueryCmd_MissingArtifact(t *testing.T) {
	cmd := newGraphCmd()
	cmd.SetArgs([]string{"query", "--root", t.TempDir(), "--callers", "x"})
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("missing artifact must fail, output:\n%s", out.String())
	}
	if !strings.Contains(err.Error(), "moai graph build") {
		t.Errorf("error must point at 'moai graph build', got: %v", err)
	}
}
