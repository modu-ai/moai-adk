package security

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitRepo initializes a throwaway git repo with one initial commit and returns
// its root. Tests use t.TempDir so cleanup is automatic (CLAUDE.local.md §6).
func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "tester")
	run("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "init")
	return dir
}

// decodeOut unmarshals guardian JSON output into a generic map. An empty buffer
// yields a nil map (the silent-pass case).
func decodeOut(t *testing.T, out *bytes.Buffer) map[string]any {
	t.Helper()
	s := strings.TrimSpace(out.String())
	if s == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("output is not valid JSON: %q (%v)", s, err)
	}
	return m
}

// TestLayer1Surfaces + TestHookOutputSchema: Layer 1 emits additionalContext for
// a dangerous write and never a decision (AC-SG-008 surface + AC-SG-022 schema).
func TestLayer1Surfaces(t *testing.T) {
	t.Parallel()
	payload := `{"tool_name":"Write","tool_input":{"content":"data = yaml.load(req.body)","file_path":"a.py"}}`
	var out bytes.Buffer
	if err := HandleSecurityScan(nil, strings.NewReader(payload), &out, t.TempDir()); err != nil {
		t.Fatalf("handler returned error (must be fail-open): %v", err)
	}
	m := decodeOut(t, &out)
	if m == nil {
		t.Fatalf("expected additionalContext output for a dangerous write, got silent pass")
	}
	hso, ok := m["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("missing hookSpecificOutput: %v", m)
	}
	if _, ok := hso["additionalContext"]; !ok {
		t.Errorf("Layer 1 must emit additionalContext: %v", hso)
	}
	// AC-SG-005/022: PostToolUse output carries NO decision field.
	if _, bad := hso["decision"]; bad {
		t.Errorf("Layer 1 (advisory) must NOT carry a decision field: %v", hso)
	}
	if _, bad := m["decision"]; bad {
		t.Errorf("Layer 1 must NOT carry a top-level decision field: %v", m)
	}
}

// TestLayer1MultiEdit: every edit in a MultiEdit batch is scanned (§C edge case).
func TestLayer1MultiEdit(t *testing.T) {
	t.Parallel()
	payload := `{"toolName":"MultiEdit","toolInput":{"edits":[{"new_string":"x = 1"},{"new_string":"os.system(cmd)"}]}}`
	var out bytes.Buffer
	_ = HandleSecurityScan(nil, strings.NewReader(payload), &out, t.TempDir())
	m := decodeOut(t, &out)
	if m == nil {
		t.Fatalf("expected a finding from the second MultiEdit edit, got silent pass")
	}
	ctx := m["hookSpecificOutput"].(map[string]any)["additionalContext"].(string)
	if !strings.Contains(ctx, "command-injection") {
		t.Errorf("MultiEdit scan must flag the second edit's command-injection: %q", ctx)
	}
}

// TestLayer1CleanSilent: a clean write produces empty stdout (silent pass).
func TestLayer1CleanSilent(t *testing.T) {
	t.Parallel()
	payload := `{"tool_name":"Write","tool_input":{"content":"func add(a,b int) int { return a+b }"}}`
	var out bytes.Buffer
	_ = HandleSecurityScan(nil, strings.NewReader(payload), &out, t.TempDir())
	if s := strings.TrimSpace(out.String()); s != "" {
		t.Errorf("clean write must be a silent pass, got %q", s)
	}
}

// TestLayer2Surfaces + TestLayer2AdvisoryDefault: at Stop, Layer 2 scans the
// working-tree diff and surfaces a high-severity finding via systemMessage with
// NO decision (advisory default, no opt-in flag) — AC-SG-008/009.
func TestLayer2Surfaces(t *testing.T) {
	// Not parallel: mutates git working tree; no env is set here.
	root := gitRepo(t)
	// Commit a safe source file, then modify it to a dangerous form (tracked
	// change → visible to `git diff HEAD`).
	app := filepath.Join(root, "app.py")
	writeCommit(t, root, "app.py", "value = 1\n", "add app")
	if err := os.WriteFile(app, []byte("data = yaml.load(request.body)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleSecurityTurn(nil, strings.NewReader(`{"last_assistant_message":"done"}`), &out, root); err != nil {
		t.Fatalf("fail-open violated: %v", err)
	}
	m := decodeOut(t, &out)
	if m == nil {
		t.Fatalf("expected a systemMessage for the dangerous diff, got silent pass")
	}
	sm, ok := m["systemMessage"].(string)
	if !ok || !strings.Contains(sm, "unsafe-deserialization") {
		t.Errorf("Layer 2 must surface the finding via systemMessage: %v", m)
	}
	// Advisory default: NO decision block without the opt-in flag (AC-SG-009).
	if _, bad := m["hookSpecificOutput"]; bad {
		t.Errorf("advisory default must NOT emit a decision block: %v", m)
	}
}

// TestAdvisoryFirst: with the blocking opt-in flag set, Layer 2 promotes to a
// decision:block on the exit-0 stdout channel (AC-SG-015 opt-in blocking).
func TestAdvisoryFirst(t *testing.T) {
	// Not parallel: sets env + mutates git tree.
	t.Setenv("MOAI_SECURITY_BLOCKING", "1")
	root := gitRepo(t)
	app := filepath.Join(root, "app.py")
	writeCommit(t, root, "app.py", "value = 1\n", "add app")
	if err := os.WriteFile(app, []byte("data = yaml.load(request.body)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	_ = HandleSecurityTurn(nil, strings.NewReader(`{}`), &out, root)
	m := decodeOut(t, &out)
	if m == nil {
		t.Fatalf("expected a blocking output when MOAI_SECURITY_BLOCKING set")
	}
	hso, ok := m["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("blocking mode must emit hookSpecificOutput: %v", m)
	}
	if hso["decision"] != "block" {
		t.Errorf("blocking mode must carry decision=block: %v", hso)
	}
	if hso["hookEventName"] != "Stop" {
		t.Errorf("Stop decision must carry hookEventName=Stop: %v", hso)
	}
}

// TestLayer3DormantDefault: with MOAI_SECURITY_COMMIT_REVIEW unset, Layer 3 is a
// silent no-op even when a commit lands with dangerous cross-file flow
// (AC-SG-011, REQ-SG-032).
func TestLayer3DormantDefault(t *testing.T) {
	// Not parallel: uses git. Env is intentionally NOT set (dormant).
	root := gitRepo(t)
	writeCommit(t, root, "route.py", "doc_id = request.args.get('id')\nrender(doc_id)\n", "add route")
	var out bytes.Buffer
	if err := HandleSecurityCommit(nil, strings.NewReader(`{}`), &out, root); err != nil {
		t.Fatalf("fail-open violated: %v", err)
	}
	if s := strings.TrimSpace(out.String()); s != "" {
		t.Errorf("Layer 3 must be a silent no-op when dormant, got %q", s)
	}
}

// TestLayer3EnabledReviews: with MOAI_SECURITY_COMMIT_REVIEW set, Layer 3 reads
// the commit's changed + related files and surfaces the cross-file finding
// (AC-SG-010, REQ-SG-030/031).
func TestLayer3EnabledReviews(t *testing.T) {
	t.Setenv("MOAI_SECURITY_COMMIT_REVIEW", "1")
	root := gitRepo(t)
	// One commit changing two source files: a user-id source + an unauthorized sink.
	if err := os.WriteFile(filepath.Join(root, "route.py"),
		[]byte("doc_id = request.args.get('id')\nrender(doc_id)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "repo.py"),
		[]byte("def render(doc_id):\n    return Model.objects.get(id=doc_id)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", ".")
	gitRun(t, root, "commit", "-m", "add route+repo")

	var out bytes.Buffer
	_ = HandleSecurityCommit(nil, strings.NewReader(`{}`), &out, root)
	m := decodeOut(t, &out)
	if m == nil {
		t.Fatalf("expected a cross-file finding when Layer 3 enabled, got silent pass")
	}
	sm, _ := m["systemMessage"].(string)
	if !strings.Contains(sm, "cross-file-idor") {
		t.Errorf("Layer 3 must surface the cross-file-idor candidate: %v", m)
	}
}

// TestLayer3DocsOnlyCommitSkipped: a docs-only commit (0 source-file delta) is
// skipped even when Layer 3 is enabled (§C edge case).
func TestLayer3DocsOnlyCommitSkipped(t *testing.T) {
	t.Setenv("MOAI_SECURITY_COMMIT_REVIEW", "1")
	root := gitRepo(t)
	writeCommit(t, root, "README.md", "# docs only, password = \"x\" in prose\n", "docs")
	var out bytes.Buffer
	_ = HandleSecurityCommit(nil, strings.NewReader(`{}`), &out, root)
	if s := strings.TrimSpace(out.String()); s != "" {
		t.Errorf("docs-only commit must be skipped (0 code-file delta), got %q", s)
	}
}

// TestSkipHookAudit: --skip-hook appends to .moai/logs/hook-skip.log and no-ops
// (AC-SG-016, REQ-SG-043).
func TestSkipHookAudit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	var out bytes.Buffer
	if err := HandleSecurityTurn([]string{"--skip-hook"}, strings.NewReader(`{}`), &out, root); err != nil {
		t.Fatalf("fail-open violated: %v", err)
	}
	if s := strings.TrimSpace(out.String()); s != "" {
		t.Errorf("--skip-hook must produce no stdout, got %q", s)
	}
	logData, err := os.ReadFile(filepath.Join(root, ".moai", "logs", "hook-skip.log"))
	if err != nil {
		t.Fatalf("hook-skip.log not written: %v", err)
	}
	if !strings.Contains(string(logData), "security-turn") || !strings.Contains(string(logData), "--skip-hook") {
		t.Errorf("hook-skip.log missing audit line: %q", string(logData))
	}
}

// TestFailOpen: missing git / non-repo dir / empty stdin all degrade to a
// silent no-op exit 0 (AC-SG-021, REQ-SG-060).
func TestFailOpen(t *testing.T) {
	t.Parallel()
	nonRepo := t.TempDir() // not a git repo

	var out bytes.Buffer
	if err := HandleSecurityTurn(nil, strings.NewReader(``), &out, nonRepo); err != nil {
		t.Errorf("Layer 2 must fail-open in a non-git dir: %v", err)
	}
	if s := strings.TrimSpace(out.String()); s != "" {
		t.Errorf("Layer 2 in a non-git dir must be silent, got %q", s)
	}

	out.Reset()
	if err := HandleSecurityScan(nil, strings.NewReader(``), &out, nonRepo); err != nil {
		t.Errorf("Layer 1 must fail-open on empty stdin: %v", err)
	}
	if s := strings.TrimSpace(out.String()); s != "" {
		t.Errorf("Layer 1 on empty stdin must be silent, got %q", s)
	}
}

// ── test helpers ──

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeCommit(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", name)
	gitRun(t, dir, "commit", "-m", msg)
}
