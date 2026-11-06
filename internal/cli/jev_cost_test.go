// jev_cost_test.go — SPEC-JEV-LEDGER-003 (AC-JEVL-014 part, AC-JEVL-015,
// AC-JEVL-016). The `moai jev cost` report: labelled aggregates over the
// decision ledger, display-only by construction, exit 0 on every ledger
// state. Written RED-first: against a tree without jev_cost.go this file
// did not compile (undefined costCmd) — the RED evidence is captured in
// progress.md §E.2.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// jevCostFixture writes ledger lines under root's ledger path (the same
// path the recorder writes: root/.moai/state/jev/decisions.jsonl).
func jevCostFixture(t *testing.T, root string, lines []string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "state", "jev")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	path := filepath.Join(dir, "decisions.jsonl")
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}

// runJevCostCmd runs the real command path — rootCmd with "jev cost" — and
// returns the captured stdout. cobra executes every command on its ROOT
// ("Regardless of what command execute is called on, run on Root only"), so
// calling costCmd.Execute() directly would run root with its own empty args
// and print root help instead of the report; the args must ride rootCmd.
// Execute returning nil is the exit-0 proxy: the CLI maps a RunE error to
// exit 1. The --json flag value is reset before every run because pflag
// keeps values set by an earlier Execute on the same command object.
func runJevCostCmd(t *testing.T, root string, args ...string) string {
	t.Helper()
	t.Setenv(config.EnvClaudeProjectDir, root)
	wantJSON := false
	for _, a := range args {
		if a == "--json" {
			wantJSON = true
		}
	}
	if err := costCmd.Flags().Set("json", fmt.Sprintf("%t", wantJSON)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(append([]string{"jev", "cost"}, args...))
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("moai jev cost %v returned %v — the report exits 0 on every ledger state", args, err)
	}
	return out.String()
}

// jevCostRow builds one ledger row line. The report reads call_id,
// input_tokens, and outcome_label; the rest of the recorder's Row shape is
// carried for realism.
func jevCostRow(callID string, inputTokens int, outcome string) string {
	return fmt.Sprintf(`{"ts":"2026-11-06T00:00:00Z","consumer":"test","event":"jev_ask","harness":"claude","call_id":%q,"question_id":"q1","rubric_version":"v1","model_id":"","request_model":"jev","availability":"ok","probability":0.5,"input_tokens":%d,"mode":"shadow","configured_mode":"shadow","threshold":0.5,"action":"none","outcome_label":%q}`, callID, inputTokens, outcome)
}

// jevCostAggregatesFixture is the AC-JEVL-015 fixture: 3 calls (a, b, c;
// input_tokens 100, 200, 300, repeated across 2 rows each), 2 rows labelled
// correct.
func jevCostAggregatesFixture() []string {
	return []string{
		jevCostRow("a", 100, "correct"),
		jevCostRow("a", 100, "unlabeled"),
		jevCostRow("b", 200, "unlabeled"),
		jevCostRow("b", 200, "unlabeled"),
		jevCostRow("c", 300, "correct"),
		jevCostRow("c", 300, "unlabeled"),
	}
}

// TestJevCost_Aggregates closes AC-JEVL-015: the fixture's aggregates, the
// literal heading, the literal estimated-input-tokens label, and decision
// D3's pin — a clean fixture prints NO skipped line.
func TestJevCost_Aggregates(t *testing.T) {
	root := t.TempDir()
	jevCostFixture(t, root, jevCostAggregatesFixture())
	out := runJevCostCmd(t, root)

	for _, want := range []string{
		"model signal aggregate",
		"calls: 3",
		"total estimated input tokens: 600",
		"correct decisions: 2",
		"estimated input tokens per correct decision: 300",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got:\n%s", want, out)
		}
	}
	// Decision D3 (pinned): the skipped line is conditioned on malformed
	// lines existing (REQ-JEVL-016); a clean ledger must not print one.
	if strings.Contains(out, "skipped") {
		t.Errorf("clean-fixture output carries a skipped line; got:\n%s", out)
	}
	// REQ-JEVL-014: no figure is presented as charged or billed.
	for _, banned := range []string{"charged", "billed", "price"} {
		if strings.Contains(out, banned) {
			t.Errorf("output presents a figure as %q; got:\n%s", banned, out)
		}
	}
}

// TestJevCost_ZeroCorrectUndefined closes the zero-correct cell of
// AC-JEVL-015: per-correct reads `undefined` (text) / `null` (--json).
func TestJevCost_ZeroCorrectUndefined(t *testing.T) {
	root := t.TempDir()
	jevCostFixture(t, root, []string{
		jevCostRow("a", 100, "unlabeled"),
		jevCostRow("b", 200, "unlabeled"),
	})

	out := runJevCostCmd(t, root)
	if !strings.Contains(out, "estimated input tokens per correct decision: undefined") {
		t.Errorf("zero correct must read undefined in text; got:\n%s", out)
	}

	out = runJevCostCmd(t, root, "--json")
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("--json output is not one JSON object: %v\ngot:\n%s", err, out)
	}
	v, ok := m["estimated_input_tokens_per_correct_decision"]
	if !ok {
		t.Fatalf("--json missing estimated_input_tokens_per_correct_decision; got:\n%s", out)
	}
	if v != nil {
		t.Errorf("zero correct must serialize as null, got %v", v)
	}
}

// TestJevCost_AbsentAndMalformed closes AC-JEVL-016: absent / empty ledgers
// report zero calls and exit 0; a malformed line is skipped and counted
// while the valid rows around it are aggregated.
func TestJevCost_AbsentAndMalformed(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		out := runJevCostCmd(t, t.TempDir())
		if !strings.Contains(out, "calls: 0") {
			t.Errorf("absent ledger must report calls: 0; got:\n%s", out)
		}
		if strings.Contains(out, "skipped") {
			t.Errorf("absent ledger carries no malformed lines; got:\n%s", out)
		}
	})

	t.Run("empty", func(t *testing.T) {
		root := t.TempDir()
		jevCostFixture(t, root, nil)
		out := runJevCostCmd(t, root)
		if !strings.Contains(out, "calls: 0") {
			t.Errorf("empty ledger must report calls: 0; got:\n%s", out)
		}
	})

	t.Run("malformed between two valid", func(t *testing.T) {
		root := t.TempDir()
		jevCostFixture(t, root, []string{
			jevCostRow("a", 100, "correct"),
			"{not json at all",
			jevCostRow("b", 200, "unlabeled"),
		})
		out := runJevCostCmd(t, root)
		for _, want := range []string{
			"calls: 2",
			"total estimated input tokens: 300",
			"correct decisions: 1",
			"estimated input tokens per correct decision: 300",
			"skipped: 1",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q; got:\n%s", want, out)
			}
		}
	})
}

// TestJevCost_JsonMatchesTextAggregates closes the --json cell of
// AC-JEVL-015: one flat object carrying the same aggregates, with `skipped`
// always present (0 included, decision D3).
func TestJevCost_JsonMatchesTextAggregates(t *testing.T) {
	root := t.TempDir()
	jevCostFixture(t, root, jevCostAggregatesFixture())
	out := runJevCostCmd(t, root, "--json")

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("--json output is not one JSON object: %v\ngot:\n%s", err, out)
	}
	for k, want := range map[string]float64{
		"calls":                  3,
		"estimated_input_tokens": 600,
		"correct":                2,
		"skipped":                0,
	} {
		got, ok := m[k]
		if !ok {
			t.Errorf("--json missing %s; got:\n%s", k, out)
			continue
		}
		f, _ := got.(float64)
		if f != want {
			t.Errorf("--json %s = %v, want %v", k, got, want)
		}
	}
	per, ok := m["estimated_input_tokens_per_correct_decision"]
	if !ok {
		t.Errorf("--json missing estimated_input_tokens_per_correct_decision; got:\n%s", out)
	} else if f, _ := per.(float64); f != 300 {
		t.Errorf("--json per-correct = %v, want 300", per)
	}
}

// TestJevCost_Boundary_NoClientOrEndpoint closes AC-JEVL-014 (part): the
// source scan of jev_cost.go. The import clause is checked with go/parser
// exact-path matching — the import-aware form, not a substring scan (a
// substring scan false-positives on the legitimate internal/jevledger
// import; decision D1 of the plan-audit repair guidance). The package guard
// TestJevCallPath_HasExactlyTheDeclaredConsumers passing UNCHANGED — with
// jev_cost.go absent from its allow-list — is the by-name proof that the
// report imports no internal/jev; this test asserts the same boundary
// directly on the file plus the URL and credential clauses of REQ-JEVL-014.
func TestJevCost_Boundary_NoClientOrEndpoint(t *testing.T) {
	const subject = "jev_cost.go"
	src, err := os.ReadFile(subject)
	if err != nil {
		t.Fatalf("read %s: %v — the subject file must exist for the non-importer proof to prove anything", subject, err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, subject, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", subject, err)
	}
	if len(f.Decls) == 0 {
		t.Fatalf("%s carries no declarations — the scan establishes nothing", subject)
	}

	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		switch path {
		case "github.com/modu-ai/moai-adk/internal/jev":
			t.Errorf("%s imports internal/jev — the cost report must construct no Jev client and send no request (REQ-JEVL-014)", subject)
		case "github.com/modu-ai/moai-adk/internal/jevcred":
			t.Errorf("%s imports internal/jevcred — the cost report must read no credential (REQ-JEVL-014)", subject)
		}
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.BasicLit:
			if v.Kind == token.STRING && (strings.Contains(v.Value, "http://") || strings.Contains(v.Value, "https://")) {
				t.Errorf("%s carries a URL literal %s — no endpoint is permitted (REQ-JEVL-014)", subject, v.Value)
			}
		case *ast.SelectorExpr:
			if id, ok := v.X.(*ast.Ident); ok && id.Name == "os" && v.Sel.Name == "Getenv" {
				t.Errorf("%s reads the environment via os.Getenv — credential-environment access is barred (REQ-JEVL-014)", subject)
			}
		case *ast.Ident:
			if v.Name == "jevcred" {
				t.Errorf("%s references jevcred — the cost report must read no credential (REQ-JEVL-014)", subject)
			}
		}
		return true
	})
}
