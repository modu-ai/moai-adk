package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// loadAgentFixture reads a committed M1 payload fixture and returns its
// tool_input. The fixtures were captured from a real Claude Code PreToolUse
// event for the Agent tool (M1 measurement gate).
func loadAgentFixture(t *testing.T, name string) *HookInput {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var in HookInput
	if err := json.Unmarshal(data, &in); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return &in
}

// TestAgentSpawnFixtureCarriesSubagentType pins AC-AME-002: the real captured
// PreToolUse payload is an Agent event whose tool_input carries a non-empty
// subagent_type.
func TestAgentSpawnFixtureCarriesSubagentType(t *testing.T) {
	t.Parallel()
	in := loadAgentFixture(t, "agent_pretool_payload.json")

	if in.HookEventName != "PreToolUse" {
		t.Errorf("hook_event_name: got %q, want PreToolUse", in.HookEventName)
	}
	if in.ToolName != "Agent" && in.ToolName != "Task" {
		t.Errorf("tool_name: got %q, want Agent or Task", in.ToolName)
	}
	sp, ok := extractAgentSpawn(in.ToolInput)
	if !ok {
		t.Fatalf("extractAgentSpawn returned ok=false for the real fixture")
	}
	if sp.Agent == "" {
		t.Errorf("subagent_type: got empty, want non-empty")
	}
}

// TestExtractAgentSpawn pins AC-AME-011 across the real fixture and the
// synthetic edge cases.
func TestExtractAgentSpawn(t *testing.T) {
	t.Parallel()

	t.Run("real fixture without model", func(t *testing.T) {
		t.Parallel()
		in := loadAgentFixture(t, "agent_pretool_payload.json")
		sp, ok := extractAgentSpawn(in.ToolInput)
		if !ok || sp.Agent != "Explore" || sp.DeclaredModel != "" {
			t.Errorf("got %+v ok=%v, want {Agent:Explore DeclaredModel:} ok=true", sp, ok)
		}
	})

	t.Run("real fixture with model", func(t *testing.T) {
		t.Parallel()
		in := loadAgentFixture(t, "agent_pretool_payload_with_model.json")
		sp, ok := extractAgentSpawn(in.ToolInput)
		if !ok || sp.Agent != "Explore" || sp.DeclaredModel != "haiku" {
			t.Errorf("got %+v ok=%v, want {Agent:Explore DeclaredModel:haiku} ok=true", sp, ok)
		}
	})

	t.Run("synthetic with model", func(t *testing.T) {
		t.Parallel()
		sp, ok := extractAgentSpawn(json.RawMessage(`{"subagent_type":"manager-git","model":"sonnet"}`))
		if !ok || sp.Agent != "manager-git" || sp.DeclaredModel != "sonnet" {
			t.Errorf("got %+v ok=%v", sp, ok)
		}
	})

	t.Run("synthetic without model", func(t *testing.T) {
		t.Parallel()
		sp, ok := extractAgentSpawn(json.RawMessage(`{"subagent_type":"manager-git"}`))
		if !ok || sp.DeclaredModel != "" {
			t.Errorf("got %+v ok=%v, want empty DeclaredModel", sp, ok)
		}
	})

	t.Run("unparseable tool_input", func(t *testing.T) {
		t.Parallel()
		if _, ok := extractAgentSpawn(json.RawMessage(`not json`)); ok {
			t.Errorf("ok=true for unparseable input, want false")
		}
	})

	t.Run("missing subagent_type", func(t *testing.T) {
		t.Parallel()
		if _, ok := extractAgentSpawn(json.RawMessage(`{"description":"x"}`)); ok {
			t.Errorf("ok=true for absent subagent_type, want false")
		}
	})

	t.Run("empty tool_input", func(t *testing.T) {
		t.Parallel()
		if _, ok := extractAgentSpawn(nil); ok {
			t.Errorf("ok=true for nil tool_input, want false")
		}
	})
}

// TestClassifyAgentModel pins the 2-valued declared-model verdict (M5): the
// record states what the spawn declared — verbatim when present, `inherit`
// otherwise — with no expected model to compare against.
func TestClassifyAgentModel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		spawn    agentSpawn
		want     agentModelVerdict
		wantResv string
	}{
		{"inherit: no declaration — the expected state", agentSpawn{Agent: "Explore"}, verdictAgentModelInherit, "inherit"},
		{"declared: explicit model recorded verbatim", agentSpawn{Agent: "Explore", DeclaredModel: "haiku"}, verdictAgentModelDeclared, "haiku"},
		{"declared: case and context suffix preserved as written", agentSpawn{Agent: "hns-some-user-specialist", DeclaredModel: "opus[1m]"}, verdictAgentModelDeclared, "opus[1m]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, recorded := classifyAgentModel(tc.spawn)
			if got != tc.want {
				t.Errorf("verdict: got %q, want %q", got, tc.want)
			}
			if recorded != tc.wantResv {
				t.Errorf("recorded: got %q, want %q", recorded, tc.wantResv)
			}
		})
	}
}

// TestAppendAgentModelAudit pins AC-AME-014.
func TestAppendAgentModelAudit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	for i := 0; i < 2; i++ {
		appendAgentModelAudit(root, agentModelAuditRecord{
			SessionID:     "sess-1",
			Agent:         "Explore",
			DeclaredModel: "",
			ResolvedModel: "inherit",
			Verdict:       string(verdictAgentModelInherit),
		})
	}

	data, err := os.ReadFile(filepath.Join(root, ".moai", "logs", agentModelAuditFileName))
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("audit lines: got %d, want 2", len(lines))
	}
	for _, ln := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(ln), &rec); err != nil {
			t.Fatalf("line is not JSON: %v (%s)", err, ln)
		}
		for _, field := range []string{"timestamp", "session_id", "agent", "declared_model", "resolved_model", "verdict"} {
			if _, ok := rec[field]; !ok {
				t.Errorf("missing required field %q in %s", field, ln)
			}
		}
		// R5: the prompt body must never reach the audit log.
		if _, ok := rec["prompt"]; ok {
			t.Errorf("audit record leaked a prompt field: %s", ln)
		}
	}
}

// TestAgentModelAuditWriteFailureFailsOpen pins AC-AME-016.
func TestAgentModelAuditWriteFailureFailsOpen(t *testing.T) {
	t.Parallel()

	t.Run("unresolved project root", func(t *testing.T) {
		t.Parallel()
		appendAgentModelAudit("", agentModelAuditRecord{Agent: "Explore"})
	})

	t.Run("read-only logs dir", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		logs := filepath.Join(root, ".moai", "logs")
		if err := os.MkdirAll(logs, 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		// Make the logs dir read-only AFTER creating it, so the append fails
		// on open rather than on mkdir.
		if err := os.Chmod(logs, 0o500); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(logs, 0o700) })
		appendAgentModelAudit(root, agentModelAuditRecord{Agent: "Explore"})
	})

	t.Run("logs path blocked by a regular file", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		// Make <root>/.moai a regular file so MkdirAll of .moai/logs fails.
		if err := os.WriteFile(filepath.Join(root, ".moai"), []byte("x"), 0o600); err != nil {
			t.Fatalf("seed: %v", err)
		}
		appendAgentModelAudit(root, agentModelAuditRecord{Agent: "Explore"})
	})

	t.Run("handler allows despite unwritable root", func(t *testing.T) {
		t.Parallel()
		h := newAgentModelTestHandler(t, "")
		out, err := h.Handle(context.Background(), agentInput("Explore", "haiku"))
		if err != nil {
			t.Fatalf("Handle returned error: %v", err)
		}
		assertNotDeny(t, out)
	})
}

// agentInput builds a PreToolUse HookInput for the Agent tool.
func agentInput(agent, model string) *HookInput {
	payload := map[string]any{"subagent_type": agent, "description": "d", "prompt": "p"}
	if model != "" {
		payload["model"] = model
	}
	raw, _ := json.Marshal(payload)
	return &HookInput{
		HookEventName: "PreToolUse",
		ToolName:      "Agent",
		SessionID:     "sess-test",
		ToolInput:     raw,
	}
}

// newAgentModelTestHandler builds a preToolHandler on the default config with
// the project root pinned to root.
func newAgentModelTestHandler(t *testing.T, root string) *preToolHandler {
	t.Helper()
	cfg := config.NewDefaultConfig()
	return &preToolHandler{
		cfg:        &auditConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: root,
	}
}

func assertNotDeny(t *testing.T, out *HookOutput) {
	t.Helper()
	if out == nil {
		t.Fatalf("nil output")
	}
	if d := decisionOf(out); d == DecisionDeny || d == DecisionAsk {
		t.Errorf("decision: got %q, want allow fall-through", d)
	}
}

func assertDeny(t *testing.T, out *HookOutput) {
	t.Helper()
	if out == nil {
		t.Fatalf("nil output")
	}
	if decisionOf(out) != DecisionDeny {
		t.Errorf("decision: got %q, want deny", decisionOf(out))
	}
}

// TestAgentModelObserveNeverBlocks pins AC-AME-015 + AC-AME-031: every verdict
// falls through to allow — the hook has no model-based deny.
func TestAgentModelObserveNeverBlocks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	h := newAgentModelTestHandler(t, root)

	cases := []struct{ name, agent, model string }{
		{"inherit", "Explore", ""},
		{"declared", "Explore", "haiku"},
		{"declared-unmapped-agent", "hns-user-specialist", "opus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := h.Handle(context.Background(), agentInput(tc.agent, tc.model))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			assertNotDeny(t, out)
		})
	}
}

// TestAgentModelNoAdvisory pins the M5 shape: with no expected model there is
// nothing to advise against — the observation record is the product. This
// replaces AC-AME-020's advisory-content contract.
func TestAgentModelNoAdvisory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	h := newAgentModelTestHandler(t, root)

	for name, model := range map[string]string{
		"inherit spawn": "",
		"declared spawn": "haiku",
	} {
		out, err := h.Handle(context.Background(), agentInput("Explore", model))
		if err != nil {
			t.Fatalf("%s: Handle: %v", name, err)
		}
		assertNotDeny(t, out)
		if out.SystemMessage != "" {
			t.Errorf("%s: advisory should be empty under declared-model logging, got %q", name, out.SystemMessage)
		}
	}
}

// TestAgentModelAdvisoryDoesNotBlock pins AC-AME-021 (M5 shape): the hook
// observes without advising and never blocks.
func TestAgentModelAdvisoryDoesNotBlock(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	h := newAgentModelTestHandler(t, root)
	out, err := h.Handle(context.Background(), agentInput("Explore", ""))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertNotDeny(t, out)
	if out.SystemMessage != "" {
		t.Errorf("advisory should be empty under declared-model logging, got %q", out.SystemMessage)
	}
}

// TestAgentModelGuardFailsOpen pins AC-AME-033: every uncertainty allows.
func TestAgentModelGuardFailsOpen(t *testing.T) {
	t.Parallel()

	unparseable := &HookInput{HookEventName: "PreToolUse", ToolName: "Agent", ToolInput: json.RawMessage(`not json`)}
	noSubagent := &HookInput{HookEventName: "PreToolUse", ToolName: "Agent", ToolInput: json.RawMessage(`{"description":"x"}`)}
	unmapped := agentInput("hns-user-specialist", "opus")
	emptyInput := &HookInput{HookEventName: "PreToolUse", ToolName: "Agent"}

	t.Run("payload uncertainties", func(t *testing.T) {
		t.Parallel()
		h := newAgentModelTestHandler(t, t.TempDir())
		for name, in := range map[string]*HookInput{
			"unparseable tool_input": unparseable,
			"absent subagent_type":   noSubagent,
			"unmapped agent":         unmapped,
			"empty tool_input":       emptyInput,
		} {
			out, err := h.Handle(context.Background(), in)
			if err != nil {
				t.Fatalf("%s: Handle: %v", name, err)
			}
			assertNotDeny(t, out)
		}
	})

	t.Run("nil ConfigProvider", func(t *testing.T) {
		t.Parallel()
		h := &preToolHandler{cfg: nil, policy: DefaultSecurityPolicy(), projectDir: t.TempDir()}
		out, err := h.Handle(context.Background(), agentInput("Explore", "haiku"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
	})

	t.Run("nil Config", func(t *testing.T) {
		t.Parallel()
		h := &preToolHandler{cfg: &auditConfigProvider{cfg: nil}, policy: DefaultSecurityPolicy(), projectDir: t.TempDir()}
		out, err := h.Handle(context.Background(), agentInput("Explore", "haiku"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
	})

	t.Run("unresolved project root", func(t *testing.T) {
		t.Parallel()
		h := newAgentModelTestHandler(t, "")
		out, err := h.Handle(context.Background(), agentInput("Explore", "haiku"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		// A deny still fires on positive evidence; the audit write is what
		// fails open. Assert only that no error escaped and a decision exists.
		if out == nil {
			t.Fatalf("nil output")
		}
	})
}

// TestAgentModelGuardDisabledNoExtraIO pins AC-AME-035: an Agent spawn
// writes only the audit log and performs no other file I/O under the root.
func TestAgentModelGuardDisabledNoExtraIO(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	h := newAgentModelTestHandler(t, root)

	if _, err := h.Handle(context.Background(), agentInput("Explore", "")); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	want := ".moai/logs/" + agentModelAuditFileName
	if len(files) != 1 || files[0] != want {
		t.Errorf("files under root: got %v, want exactly [%s]", files, want)
	}
}

// TestPruneObservationLogsIncludesAgentModelAudit pins AC-AME-017.
func TestPruneObservationLogsIncludesAgentModelAudit(t *testing.T) {
	t.Parallel()
	logs := t.TempDir()
	path := filepath.Join(logs, agentModelAuditFileName)
	if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	now := time.Now()
	old := now.AddDate(0, 0, -400)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	stats := PruneObservationLogs(logs, "sess", 30, now)
	if stats.AgentModelAuditAged != 1 {
		t.Errorf("AgentModelAuditAged: got %d, want 1", stats.AgentModelAuditAged)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("aged agent-model audit log was not pruned")
	}
}

// TestPreToolDecisionPrecedence pins AC-AME-061: inserting the Agent branch
// does not disturb the existing Bash deny precedence.
func TestPreToolDecisionPrecedence(t *testing.T) {
	t.Parallel()
	h := newAgentModelTestHandler(t, t.TempDir())

	raw, _ := json.Marshal(map[string]any{"command": "rm -rf /"})
	out, err := h.Handle(context.Background(), &HookInput{
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     raw,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertDeny(t, out)
	if strings.Contains(reasonOf(out), "AGENT_MODEL_VIOLATION") {
		t.Errorf("Bash dangerous-pattern deny was displaced by the agent-model guard: %s", reasonOf(out))
	}
}
