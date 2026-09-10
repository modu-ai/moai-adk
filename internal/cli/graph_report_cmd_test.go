package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// graph_report_cmd_test.go — SPEC-GRAPH-REPORT-001 M2 CLI layer
// (REQ-GR-005/006): the `moai graph report` subcommand writes the rotating
// report to the FIXED path <root>/.moai/reports/graph-report.md (no --out
// flag exists to redirect it) and exits 0 even when code-dependent sections
// are empty with a stated reason. Hand-written edges.jsonl fixtures are the
// sanctioned fixture form (plan §F M2 test spec, B8).

// reportEdgesBody is a compact mixed-kind artifact exercising all three
// sections: an import cycle (pkg/a -> pkg/c -> pkg/b -> pkg/a), a boundary
// INFERRED code-call edge (Run in internal/cli -> Boot declared only in
// internal/hook), an intra-package INFERRED edge, and an ambiguous callee
// ("Dup" declared in two packages).
const reportEdgesBody = `{"kind":"import","source":"pkg/a","target":"pkg/c"}
{"kind":"import","source":"pkg/c","target":"pkg/b"}
{"kind":"import","source":"pkg/b","target":"pkg/a"}
{"kind":"code-call","source":"internal/hook/start.go:Boot","target":"Scan","resolution":"inferred","confidence":0.85}
{"kind":"code-call","source":"internal/graph/b.go:Beta","target":"Alpha","resolution":"inferred","confidence":0.85}
{"kind":"code-call","source":"internal/graph/a.go:Alpha","target":"Beta","resolution":"inferred","confidence":0.85}
{"kind":"code-call","source":"internal/wire/sib.go:Dup","target":"Remote","resolution":"inferred","confidence":0.85}
{"kind":"code-call","source":"internal/far/far.go:Dup","target":"Remote","resolution":"inferred","confidence":0.85}
{"kind":"code-call","source":"internal/cli/main.go:Run","target":"Boot","resolution":"inferred","confidence":0.85,"line":12}
{"kind":"mx-spec","source":"internal/demo/demo.go","target":"SPEC-GRAPH-REPORT-001","line":4}
`

// docOnlyEdgesBody is the nocgo shape (B5): a doc-only artifact with zero
// code-call edges — the code-dependent sections must still emit with the
// stated reason and the command must exit 0 (AC-GR-010).
const docOnlyEdgesBody = `{"kind":"import","source":"internal/cli","target":"internal/config"}
{"kind":"mx-spec","source":"internal/demo/demo.go","target":"SPEC-GRAPH-REPORT-001","line":4}
`

// writeReportFixture writes an edges.jsonl artifact under root's default
// graph location and returns the root.
func writeReportFixture(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "project", "graph")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "edges.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// runGraphReport executes `moai graph report --root <root>` and returns the
// captured output plus the command error.
func runGraphReport(t *testing.T, root string, extra ...string) (string, error) {
	t.Helper()
	cmd := newGraphCmd()
	cmd.SetArgs(append([]string{"report", "--root", root}, extra...))
	out := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	err := cmd.Execute()
	return out.String(), err
}

// TestGraphReportCmd_WritesFixedPathWithSections covers AC-GR-006: the
// report lands at the fixed rotating path under the fixture root and carries
// the three section headings, exit 0.
func TestGraphReportCmd_WritesFixedPathWithSections(t *testing.T) {
	root := writeReportFixture(t, reportEdgesBody)

	out, err := runGraphReport(t, root)
	if err != nil {
		t.Fatalf("graph report: %v\noutput: %s", err, out)
	}

	fixed := filepath.Join(root, ".moai", "reports", "graph-report.md")
	data, err := os.ReadFile(fixed)
	if err != nil {
		t.Fatalf("fixed-path report missing (%s): %v (cmd output: %s)", fixed, err, out)
	}
	body := string(data)
	for _, want := range []string{
		"## God Nodes",
		"## Surprising Connections",
		"## Import Cycles",
		"fan-in over: code-call, import",
		// The boundary INFERRED edge scores; the ambiguous callee never does.
		"internal/cli/main.go:Run -> internal/hook/start.go:Boot",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("report missing %q; body:\n%s", want, body)
		}
	}
	if strings.Contains(body, ":Dup") {
		t.Errorf("ambiguous callee rendered into the report; body:\n%s", body)
	}
	if !strings.Contains(out, fixed) {
		t.Errorf("command output should name the written path %s; output: %s", fixed, out)
	}
}

// TestGraphReportCmd_DoubleRunByteIdentical covers AC-GR-011: two runs on
// the same fixture produce byte-identical report files (the test-level form
// of the `cmp` exit-0 evidence).
func TestGraphReportCmd_DoubleRunByteIdentical(t *testing.T) {
	root := writeReportFixture(t, reportEdgesBody)

	if _, err := runGraphReport(t, root); err != nil {
		t.Fatalf("first run: %v", err)
	}
	first, err := os.ReadFile(filepath.Join(root, ".moai", "reports", "graph-report.md"))
	if err != nil {
		t.Fatalf("first report missing: %v", err)
	}

	if _, err := runGraphReport(t, root); err != nil {
		t.Fatalf("second run: %v", err)
	}
	second, err := os.ReadFile(filepath.Join(root, ".moai", "reports", "graph-report.md"))
	if err != nil {
		t.Fatalf("second report missing: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("two runs are not byte-identical:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

// TestGraphReportCmd_EmptyCodeLayerStillEmits covers AC-GR-010 (B5): a
// doc-only (nocgo-shaped) artifact still produces the report at the fixed
// path, code-dependent sections present-but-empty with the stated reason,
// and exit 0.
func TestGraphReportCmd_EmptyCodeLayerStillEmits(t *testing.T) {
	root := writeReportFixture(t, docOnlyEdgesBody)

	out, err := runGraphReport(t, root)
	if err != nil {
		t.Fatalf("graph report on doc-only artifact must exit 0, got %v\noutput: %s", err, out)
	}
	data, err := os.ReadFile(filepath.Join(root, ".moai", "reports", "graph-report.md"))
	if err != nil {
		t.Fatalf("report missing after doc-only run: %v (output: %s)", err, out)
	}
	body := string(data)
	for _, want := range []string{
		"## God Nodes",
		"## Surprising Connections",
		"## Import Cycles",
		"code layer absent: CGO disabled or no extraction",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("doc-only report missing %q; body:\n%s", want, body)
		}
	}
}

// TestGraphReportCmd_AbsentArtifactActionableError keeps the query-command
// convention: a missing edges.jsonl is an actionable error naming the build
// remedy, never a silently empty report.
func TestGraphReportCmd_AbsentArtifactActionableError(t *testing.T) {
	root := t.TempDir()

	_, err := runGraphReport(t, root)
	if err == nil {
		t.Fatal("absent edges artifact must error, got nil")
	}
	if !strings.Contains(err.Error(), "moai graph build") {
		t.Errorf("error must name the build remedy, got: %v", err)
	}
}

// TestGraphReportCmd_FlagsExactlyRootAndLimit pins REQ-GR-005's fixed-path
// mandate at the flag surface: --root and --limit exist, no --out exists —
// a user-selectable output path would defeat the derived-artifact contract.
func TestGraphReportCmd_FlagsExactlyRootAndLimit(t *testing.T) {
	cmd := newGraphReportCmd()
	if cmd.Flags().Lookup("root") == nil {
		t.Error("--root flag missing")
	}
	if cmd.Flags().Lookup("limit") == nil {
		t.Error("--limit flag missing")
	}
	if cmd.Flags().Lookup("out") != nil {
		t.Error("--out flag must NOT exist — the report path is fixed (REQ-GR-005 D1)")
	}
}

// TestGraphReportCmd_ErrorPathsStructured keeps the structured-error
// convention on the failure branches: an unreadable edges artifact is a
// load error (distinct from the absent-artifact remedy), and a blocked
// report directory is a mkdir error — both named, neither a panic.
func TestGraphReportCmd_ErrorPathsStructured(t *testing.T) {
	t.Run("unreadable edges artifact", func(t *testing.T) {
		root := t.TempDir()
		// edges.jsonl as a directory: present, but not readable as a file.
		if err := os.MkdirAll(filepath.Join(root, ".moai", "project", "graph", "edges.jsonl"), 0o755); err != nil {
			t.Fatal(err)
		}
		_, err := runGraphReport(t, root)
		if err == nil || !strings.Contains(err.Error(), "load edges") {
			t.Errorf("unreadable artifact must surface a load error, got: %v", err)
		}
	})

	t.Run("blocked report directory", func(t *testing.T) {
		root := writeReportFixture(t, reportEdgesBody)
		// .moai/reports exists as a regular file — MkdirAll cannot proceed.
		if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".moai", "reports"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := runGraphReport(t, root)
		if err == nil || !strings.Contains(err.Error(), "mkdir") {
			t.Errorf("blocked report dir must surface a mkdir error, got: %v", err)
		}
	})
}
