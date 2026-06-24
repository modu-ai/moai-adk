// SPEC-STOP-EVIDENCE-WRITER-001: 기록 시점 증거 writer 테스트.
// M1 분류기(classifyTestCommand / classifyPathKind) + M2 조립기
// (buildEvidenceRecord) + M3 writer(logEvidence) + M4 동작 보존 가드 +
// M5 falsifiable 게이트 활성화 end-to-end 증명을 다룬다.
package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// --- M1: classifyTestCommand ---

func TestClassifyTestCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		command  string
		result   []byte
		wantTest bool
		wantPass bool
		wantFail bool
	}{
		// go test — pass/fail/ambiguous
		{"go_test_pass", "go test ./...", []byte(`{"exit_code":0}`), true, true, false},
		{"go_test_fail", "go test ./...", []byte(`{"exit_code":1}`), true, false, true},
		{"go_test_ambiguous", "go test ./...", nil, true, false, false},
		{"go_test_pass_output", "go test ./internal/hook/", []byte("ok  \tgithub.com/x/y\t0.5s"), true, true, false},
		{"go_test_fail_output", "go test ./internal/hook/", []byte("--- FAIL: TestX\nFAIL\tgithub.com/x/y\t0.5s"), true, false, true},
		// D2 (sync-audit): the LIVE Claude Code Bash tool_response is a WRAPPED JSON
		// object whose stdout string JSON-escapes the tab to the two-char "\t" — the
		// raw-tab marker "ok  \t" then never matches the raw bytes, so a go-test PASS
		// silently degrades to ambiguous. wantPass must be true after the shape-resilient
		// JSON-decode fix (these failed before the fix: returned (true,false,false)).
		{"go_test_pass_wrapped_json", "go test ./...", mustRaw(map[string]any{"stdout": "ok  \tgithub.com/modu-ai/moai-adk/internal/hook\t0.5s\n", "interrupted": false}), true, true, false},
		{"go_test_fail_wrapped_json", "go test ./...", mustRaw(map[string]any{"stdout": "--- FAIL: TestX (0.00s)\nFAIL\tgithub.com/modu-ai/moai-adk/internal/hook\t0.5s\n", "interrupted": false}), true, false, true},
		// wrapped exit_code (e.g. {"stdout":"...","exit_code":0}) — structured signal still wins.
		{"go_test_pass_wrapped_exitcode", "go test ./...", mustRaw(map[string]any{"stdout": "ok  \tx\t0.1s\n", "exit_code": 0}), true, true, false},
		{"go_test_fail_wrapped_exitcode", "go test ./...", mustRaw(map[string]any{"stdout": "FAIL\n", "exit_code": 1}), true, false, true},
		// wrapped interrupted=true must be conservative (fail), regardless of stdout text.
		{"go_test_wrapped_interrupted", "go test ./...", mustRaw(map[string]any{"stdout": "ok  \tx\t0.1s\n", "interrupted": true}), true, false, true},
		// pytest wrapped — count-word marker survives in decoded stdout.
		{"pytest_fail_wrapped_json", "pytest tests/", mustRaw(map[string]any{"stdout": "1 failed, 2 passed in 0.3s\n", "interrupted": false}), true, false, true},
		{"pytest_pass_wrapped_json", "pytest tests/", mustRaw(map[string]any{"stdout": "3 passed in 0.3s\n", "interrupted": false}), true, true, false},
		// cargo wrapped — precise marker survives in decoded stdout.
		{"cargo_pass_wrapped_json", "cargo test", mustRaw(map[string]any{"stdout": "test result: ok. 5 passed; 0 failed\n", "interrupted": false}), true, true, false},
		// go-test PASS reported on stderr (build-cache notices on stdout, result on stderr).
		{"go_test_pass_wrapped_stderr", "go test ./...", mustRaw(map[string]any{"stdout": "", "stderr": "ok  \tx\t0.1s\n", "interrupted": false}), true, true, false},
		// nested exit_code (e.g. {"result":{"exit_code":0}}) — structured signal found 1-depth deep.
		{"go_test_pass_nested_exitcode", "go test ./...", mustRaw(map[string]any{"result": map[string]any{"exit_code": 0}}), true, true, false},
		{"go_test_fail_nested_exitcode", "go test ./...", mustRaw(map[string]any{"result": map[string]any{"exit_code": 2}}), true, false, true},
		// unknown text key (not in textKeys) — second loop in decodeToolResponse still extracts it.
		{"go_test_pass_unknown_textkey", "go test ./...", mustRaw(map[string]any{"log": "ok  \tx\t0.1s\n"}), true, true, false},
		// wrapped object with no usable text (empty stdout, no marker) → ambiguous (no signal).
		{"go_test_wrapped_empty_text", "go test ./...", mustRaw(map[string]any{"stdout": "", "interrupted": false}), true, false, false},
		{"go_test_run_flag", "go test -run TestX ./pkg", []byte(`{"exit_code":0}`), true, true, false},
		// pytest
		{"pytest_pass", "pytest", []byte(`{"exit_code":0}`), true, true, false},
		{"pytest_fail_output", "pytest tests/", []byte("1 failed, 2 passed in 0.3s"), true, false, true},
		{"pytest_pass_output", "pytest tests/", []byte("3 passed in 0.3s"), true, true, false},
		// cargo test
		{"cargo_test_pass", "cargo test", []byte("test result: ok. 5 passed; 0 failed"), true, true, false},
		{"cargo_test_fail", "cargo test", []byte("test result: FAILED. 1 passed; 1 failed"), true, false, true},
		// npm / pnpm / yarn / jest / vitest
		{"npm_test", "npm test", []byte(`{"exit_code":0}`), true, true, false},
		{"npm_run_test", "npm run test", []byte(`{"exit_code":1}`), true, false, true},
		{"pnpm_test", "pnpm test", []byte(`{"exit_code":0}`), true, true, false},
		{"yarn_test", "yarn test", []byte(`{"exit_code":0}`), true, true, false},
		{"jest", "jest --coverage", []byte(`{"exit_code":0}`), true, true, false},
		{"vitest", "vitest run", []byte(`{"exit_code":1}`), true, false, true},
		// interrupted signal → fail (non-clean exit)
		{"go_test_interrupted", "go test ./...", []byte(`{"interrupted":true}`), true, false, true},
		// non-test commands → not a test, neither flag
		{"go_build_not_test", "go build ./...", []byte(`{"exit_code":0}`), false, false, false},
		{"ls_not_test", "ls -la", []byte(`{"exit_code":0}`), false, false, false},
		{"git_status_not_test", "git status", []byte(`{"exit_code":0}`), false, false, false},
		// a non-test command whose arg happens to mention "test" must NOT match as a test
		{"cat_testfile_not_test", "cat foo_test.go", []byte(`{"exit_code":0}`), false, false, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			isTest, isPass, isFail := classifyTestCommand(tt.command, tt.result)
			if isTest != tt.wantTest {
				t.Errorf("isTest = %v, want %v", isTest, tt.wantTest)
			}
			if isPass != tt.wantPass {
				t.Errorf("isPass = %v, want %v", isPass, tt.wantPass)
			}
			if isFail != tt.wantFail {
				t.Errorf("isFail = %v, want %v", isFail, tt.wantFail)
			}
		})
	}
}

// --- M1: classifyPathKind ---

func TestClassifyPathKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"go_ext", "internal/hook/post_tool.go", telemetry.PathKindCodeChange},
		{"py_ext", "src/app.py", telemetry.PathKindCodeChange},
		{"ts_ext", "web/index.ts", telemetry.PathKindCodeChange},
		{"rs_ext", "src/main.rs", telemetry.PathKindCodeChange},
		{"swift_ext", "App/main.swift", telemetry.PathKindCodeChange},
		{"md_ext", "docs/guide.md", telemetry.PathKindDocsOnly},
		{"mdx_ext", "site/page.mdx", telemetry.PathKindDocsOnly},
		{"txt_ext", "notes.txt", telemetry.PathKindDocsOnly},
		{"rst_ext", "doc.rst", telemetry.PathKindDocsOnly},
		{"readme_base", "README", telemetry.PathKindDocsOnly},
		{"readme_md_base", "README.md", telemetry.PathKindDocsOnly},
		{"changelog_base", "CHANGELOG.md", telemetry.PathKindDocsOnly},
		{"spec_md", ".moai/specs/SPEC-X-001/spec.md", telemetry.PathKindDocsOnly},
		{"unknown_ext", "config.xyz", telemetry.PathKindUnknown},
		{"no_ext", "Makefile", telemetry.PathKindUnknown},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyPathKind(tt.filePath)
			if got != tt.want {
				t.Errorf("classifyPathKind(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

// --- M2: buildEvidenceRecord ---

// bashInput builds a PostToolUse HookInput for a Bash event.
func bashInput(sessionID, command string, result []byte) *HookInput {
	in := json.RawMessage(`{"command":` + mustJSON(command) + `}`)
	return &HookInput{
		ToolName:     "Bash",
		SessionID:    sessionID,
		ToolInput:    in,
		ToolResponse: result,
	}
}

// fileInput builds a PostToolUse HookInput for an Edit/Write event.
func fileInput(tool, sessionID, filePath string) *HookInput {
	in := json.RawMessage(`{"file_path":` + mustJSON(filePath) + `}`)
	return &HookInput{
		ToolName:  tool,
		SessionID: sessionID,
		ToolInput: in,
	}
}

func mustJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// mustRaw marshals v to JSON bytes — used to build a realistic WRAPPED Bash
// tool_response (e.g. {"stdout":"...\nok  \t...\n","interrupted":false}) where
// the encoder JSON-escapes the embedded tab to the two-char sequence "\t",
// exactly as the live Claude Code Bash hook delivers it (D2 sync-audit fix).
func mustRaw(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func TestBuildEvidenceRecord(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		input        *HookInput
		wantOK       bool
		wantOutcome  string
		wantPass     bool
		wantFail     bool
		wantPathKind string
	}{
		{
			name:        "bash_test_pass",
			input:       bashInput("s1", "go test ./...", []byte(`{"exit_code":0}`)),
			wantOK:      true,
			wantOutcome: telemetry.OutcomeSuccess,
			wantPass:    true,
			wantFail:    false,
		},
		{
			name:        "bash_test_fail",
			input:       bashInput("s1", "go test ./...", []byte(`{"exit_code":1}`)),
			wantOK:      true,
			wantOutcome: telemetry.OutcomeError,
			wantPass:    false,
			wantFail:    true,
		},
		{
			name:        "bash_test_ambiguous",
			input:       bashInput("s1", "go test ./...", nil),
			wantOK:      true,
			wantOutcome: telemetry.OutcomeUnknown,
			wantPass:    false,
			wantFail:    false,
		},
		{
			name:         "edit_go_code_change",
			input:        fileInput("Edit", "s1", "internal/hook/post_tool.go"),
			wantOK:       true,
			wantOutcome:  telemetry.OutcomeSuccess,
			wantPathKind: telemetry.PathKindCodeChange,
		},
		{
			name:         "write_md_docs_only",
			input:        fileInput("Write", "s1", "docs/guide.md"),
			wantOK:       true,
			wantOutcome:  telemetry.OutcomeUnknown,
			wantPathKind: telemetry.PathKindDocsOnly,
		},
		{
			name:   "non_test_bash_ok_false",
			input:  bashInput("s1", "go build ./...", []byte(`{"exit_code":0}`)),
			wantOK: false,
		},
		{
			name:   "unknown_ext_edit_ok_false",
			input:  fileInput("Edit", "s1", "config.xyz"),
			wantOK: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec, ok := buildEvidenceRecord(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if rec.Outcome != tt.wantOutcome {
				t.Errorf("Outcome = %q, want %q", rec.Outcome, tt.wantOutcome)
			}
			if rec.IsTestPass != tt.wantPass {
				t.Errorf("IsTestPass = %v, want %v", rec.IsTestPass, tt.wantPass)
			}
			if rec.IsTestFail != tt.wantFail {
				t.Errorf("IsTestFail = %v, want %v", rec.IsTestFail, tt.wantFail)
			}
			if rec.PathKind != tt.wantPathKind {
				t.Errorf("PathKind = %q, want %q", rec.PathKind, tt.wantPathKind)
			}
			// REQ-SEW-016: SessionID propagated.
			if rec.SessionID != tt.input.SessionID {
				t.Errorf("SessionID = %q, want %q", rec.SessionID, tt.input.SessionID)
			}
			// Timestamp set (non-zero) so daily-file rotation works.
			if rec.Timestamp.IsZero() {
				t.Error("Timestamp is zero; want set")
			}
		})
	}
}

// TestBuildEvidenceRecord_Guards covers the dispatcher default branch and the
// empty-input guards (defensive paths not reached from the wired Handle).
func TestBuildEvidenceRecord_Guards(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input *HookInput
	}{
		// Non-Bash/Edit/Write tool name → dispatcher default → ok=false.
		{"read_tool", &HookInput{ToolName: "Read", SessionID: "s", ToolInput: json.RawMessage(`{"file_path":"x.go"}`)}},
		{"skill_tool", &HookInput{ToolName: "Skill", SessionID: "s", ToolInput: json.RawMessage(`{"skill":"x"}`)}},
		// Bash with empty command → ok=false.
		{"bash_empty_command", &HookInput{ToolName: "Bash", SessionID: "s", ToolInput: json.RawMessage(`{}`)}},
		{"bash_nil_input", &HookInput{ToolName: "Bash", SessionID: "s"}},
		// Edit with empty file_path → ok=false.
		{"edit_empty_path", &HookInput{ToolName: "Edit", SessionID: "s", ToolInput: json.RawMessage(`{}`)}},
		{"write_nil_input", &HookInput{ToolName: "Write", SessionID: "s"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, ok := buildEvidenceRecord(tt.input); ok {
				t.Errorf("expected ok=false for %s, got ok=true", tt.name)
			}
		})
	}
}

// --- M3: logEvidence writer + Handle wiring ---

// newMoaiTempRoot creates a temp dir with a .moai/ subdir so resolveProjectRoot
// accepts it, and points CLAUDE_PROJECT_DIR at it for the duration of the test.
func newMoaiTempRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	return root
}

func TestLogEvidence_NoProjectRoot_Skips(t *testing.T) {
	// No CLAUDE_PROJECT_DIR, no CWD → resolveProjectRoot returns "" → no write, no panic.
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	in := bashInput("s1", "go test ./...", []byte(`{"exit_code":0}`))
	in.CWD = ""
	logEvidence(in) // must not panic
}

func TestLogEvidence_NonEvidenceEvent_NoWrite(t *testing.T) {
	root := newMoaiTempRoot(t)
	in := bashInput("sess-none", "go build ./...", []byte(`{"exit_code":0}`))
	in.CWD = root
	logEvidence(in)

	recs, err := telemetry.LoadBySession(root, "sess-none")
	if err != nil {
		t.Fatalf("LoadBySession: %v", err)
	}
	if len(recs) != 0 {
		t.Fatalf("expected 0 records for non-evidence event, got %d", len(recs))
	}
}

func TestLogEvidence_EvidenceEvent_Writes(t *testing.T) {
	root := newMoaiTempRoot(t)
	in := bashInput("sess-w", "go test ./...", []byte(`{"exit_code":0}`))
	in.CWD = root
	logEvidence(in)

	recs, err := telemetry.LoadBySession(root, "sess-w")
	if err != nil {
		t.Fatalf("LoadBySession: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if !recs[0].IsTestPass {
		t.Errorf("expected IsTestPass=true on persisted record")
	}
	if recs[0].SessionID != "sess-w" {
		t.Errorf("SessionID = %q, want sess-w", recs[0].SessionID)
	}
}

func TestLogEvidence_SessionID(t *testing.T) {
	root := newMoaiTempRoot(t)
	in := fileInput("Edit", "corr-id-42", "internal/hook/x.go")
	in.CWD = root
	logEvidence(in)

	recs, _ := telemetry.LoadBySession(root, "corr-id-42")
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if recs[0].SessionID != "corr-id-42" {
		t.Errorf("SessionID = %q, want corr-id-42", recs[0].SessionID)
	}
}

func TestLogEvidence_FailOpen_ReadOnlyParent(t *testing.T) {
	// Make the telemetry dir's parent a regular FILE so MkdirAll fails inside
	// RecordSkillUsage; logEvidence must swallow the error (slog.Warn), not panic.
	root := newMoaiTempRoot(t)
	// Create .moai/evolution as a regular file so .moai/evolution/telemetry MkdirAll fails.
	evoPath := filepath.Join(root, ".moai", "evolution")
	if err := os.WriteFile(evoPath, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}
	in := bashInput("sess-fail", "go test ./...", []byte(`{"exit_code":1}`))
	in.CWD = root
	logEvidence(in) // must not panic, must not return (void)
}

func TestLogEvidence_NeverBlocks(t *testing.T) {
	// logEvidence returns nothing (void) — structural guarantee that it cannot
	// propagate an error into Handle. This compiles only if the signature is void.
	root := newMoaiTempRoot(t)
	in := bashInput("s-x", "go test ./...", []byte(`{"exit_code":0}`))
	in.CWD = root
	logEvidence(in)
}

// TestHandle_RecordsEvidenceOnBashEditWrite — AC-SEW-002: a Bash test event
// dispatched through the production Handle path produces a loadable record.
func TestHandle_RecordsEvidenceOnBashEditWrite(t *testing.T) {
	root := newMoaiTempRoot(t)
	h := NewPostToolHandler()
	in := bashInput("handle-sess", "go test ./...", []byte(`{"exit_code":0}`))
	in.CWD = root

	out, err := h.Handle(t.Context(), in)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle returned nil output")
	}

	recs, _ := telemetry.LoadBySession(root, "handle-sess")
	if len(recs) != 1 {
		t.Fatalf("expected 1 evidence record via Handle, got %d", len(recs))
	}
	if !recs[0].IsTestPass {
		t.Errorf("expected IsTestPass=true on Handle-recorded evidence")
	}
}

// --- M4: behavior preservation ---

// TestHandle_PreservesExistingObservers — a Skill event still writes its
// logSkillUsage record AND a Bash event adds no systemMessage from the writer.
func TestHandle_PreservesExistingObservers(t *testing.T) {
	root := newMoaiTempRoot(t)
	h := NewPostToolHandler()

	// (a) Skill event still records via logSkillUsage (unchanged behavior).
	skillIn := &HookInput{
		ToolName:  "Skill",
		SessionID: "skill-sess",
		CWD:       root,
		ToolInput: json.RawMessage(`{"skill":"moai-foundation-core"}`),
	}
	if _, err := h.Handle(t.Context(), skillIn); err != nil {
		t.Fatalf("Handle(Skill) error: %v", err)
	}
	skillRecs, _ := telemetry.LoadBySession(root, "skill-sess")
	if len(skillRecs) != 1 {
		t.Fatalf("expected 1 Skill record (logSkillUsage preserved), got %d", len(skillRecs))
	}
	if skillRecs[0].SkillID != "moai-foundation-core" {
		t.Errorf("Skill record SkillID = %q, want moai-foundation-core", skillRecs[0].SkillID)
	}

	// (b) Bash event must not introduce a systemMessage from the evidence writer.
	bashIn := bashInput("bash-sess", "go test ./...", []byte(`{"exit_code":0}`))
	bashIn.CWD = root
	out, err := h.Handle(t.Context(), bashIn)
	if err != nil {
		t.Fatalf("Handle(Bash) error: %v", err)
	}
	if out.SystemMessage != "" {
		t.Errorf("evidence writer must not set SystemMessage; got %q", out.SystemMessage)
	}
}

// --- M5: falsifiable gate activation (THE value proof, AC-SEW-009) ---

// TestEvidenceGate_ActivatesInProduction builds a session purely from the
// SPEC's own buildEvidenceRecord + telemetry.RecordSkillUsage, then reads it
// back through the UNCHANGED GATE-001 chain
// (LoadBySession → buildSessionLedger → evaluateEvidence) and asserts:
//   - code-change(success) + test-FAIL(no pass) → non-nil Finding (gate fires)
//   - same code-change + test-PASS              → nil (gate stays silent)
//
// The only delta between the two cases is IsTestPass (false→true). If the
// writer were a no-op (records never written, or evidence fields never set),
// the positive subtest would not produce a finding and this AC would FAIL.
//
// (D2 sync-audit) The Bash tool_response fixtures use the REALISTIC WRAPPED JSON
// shape the live Claude Code Bash hook delivers ({"stdout":"...ok  \t...","interrupted":false}
// for PASS, "...--- FAIL..." for FAIL) — not the idealized {"exit_code":N} — so
// the activation proof "fires on real evidence," not just idealized plumbing.
func TestEvidenceGate_ActivatesInProduction(t *testing.T) {
	// persist stores a buildEvidenceRecord result into the temp store.
	persist := func(t *testing.T, root string, in *HookInput) {
		t.Helper()
		rec, ok := buildEvidenceRecord(in)
		if !ok {
			t.Fatalf("buildEvidenceRecord returned ok=false for %s %s", in.ToolName, string(in.ToolInput))
		}
		if err := telemetry.RecordSkillUsage(root, rec); err != nil {
			t.Fatalf("RecordSkillUsage: %v", err)
		}
	}

	t.Run("code_change_test_fail_no_pass_finding", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
			t.Fatal(err)
		}
		const sess = "gate-fire"
		// (1) Edit .go → {Outcome=success, PathKind=code-change}
		persist(t, root, fileInput("Edit", sess, "internal/hook/post_tool.go"))
		// (2) Bash go test FAIL → {Outcome=error, IsTestFail=true}
		//     realistic wrapped tool_response (D2): --- FAIL inside escaped stdout.
		persist(t, root, bashInput(sess, "go test ./...",
			mustRaw(map[string]any{"stdout": "--- FAIL: TestX (0.00s)\nFAIL\tgithub.com/modu-ai/moai-adk/internal/hook\t0.5s\n", "interrupted": false})))

		recs, err := telemetry.LoadBySession(root, sess)
		if err != nil {
			t.Fatalf("LoadBySession: %v", err)
		}
		finding := evaluateEvidence(buildSessionLedger(recs))
		if finding == nil {
			t.Fatal("expected non-nil Finding (code-change + test-fail + no pass), got nil — gate did NOT fire")
		}
		if got := finding.HumanReadable(); !strings.Contains(got, "path-kind=code-change") {
			t.Errorf("Finding.HumanReadable() = %q, want substring path-kind=code-change", got)
		}
	})

	t.Run("code_change_test_pass_nil", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
			t.Fatal(err)
		}
		const sess = "gate-silent"
		// (1) Edit .go → {Outcome=success, PathKind=code-change}
		persist(t, root, fileInput("Edit", sess, "internal/hook/post_tool.go"))
		// (2) Bash go test PASS → {Outcome=success, IsTestPass=true}  ← only delta
		//     realistic wrapped tool_response (D2): escaped-tab "ok  \t" inside stdout —
		//     this is precisely the shape that silently degraded to ambiguous before the fix.
		persist(t, root, bashInput(sess, "go test ./...",
			mustRaw(map[string]any{"stdout": "ok  \tgithub.com/modu-ai/moai-adk/internal/hook\t0.5s\n", "interrupted": false})))

		recs, _ := telemetry.LoadBySession(root, sess)
		finding := evaluateEvidence(buildSessionLedger(recs))
		if finding != nil {
			t.Fatalf("expected nil Finding (success backed by observed pass), got %+v — gate fired falsely", finding)
		}
	})
}
