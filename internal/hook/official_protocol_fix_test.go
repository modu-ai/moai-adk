// Package hook — regression tests for the official-spec protocol alignment
// (Claude Code hooks spec, code.claude.com/docs/en/hooks, v2.1.196):
//
//   - G4: Dispatch honors a successful handler output even when the handler
//     finishes just past the deadline (handler err inspected before ctx.Err()).
//   - G5: multi-hook merge semantics — additionalContext from EVERY hook kept;
//     updatedInput/updatedToolOutput last-non-nil wins; stopReason/retry propagate.
//   - G6: Handlers() is race-safe (returns a copy under the mutex).
//   - G7: ReadInput caps stdin at maxHookInputBytes.
//   - G13: ProjectDir populated env-first (CLAUDE_PROJECT_DIR) then cwd, so
//     SessionStart/Stop side effects fire under the official runtime (which
//     never sends a project_dir stdin field).
//   - G17: official input field names (config_source, type, task_name) decode.
package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// slowSuccessHandler ignores ctx and returns a valid output after sleeping
// past the dispatch deadline — modeling a handler that completes real work
// slightly late.
type slowSuccessHandler struct {
	event EventType
	sleep time.Duration
}

func (h *slowSuccessHandler) Handle(_ context.Context, _ *HookInput) (*HookOutput, error) {
	time.Sleep(h.sleep)
	return &HookOutput{SystemMessage: "late but valid"}, nil
}

func (h *slowSuccessHandler) EventType() EventType { return h.event }

// ctxAwareHandler returns ctx.Err() when the deadline passes (well-behaved).
type ctxAwareHandler struct {
	event EventType
	sleep time.Duration
}

func (h *ctxAwareHandler) Handle(ctx context.Context, _ *HookInput) (*HookOutput, error) {
	select {
	case <-time.After(h.sleep):
		return &HookOutput{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (h *ctxAwareHandler) EventType() EventType { return h.event }

// TestDispatch_LateSuccessOutputHonored is the G4 regression test: a handler
// that returns (output, nil) after the deadline must have its output honored,
// not discarded and replaced with ErrHookTimeout.
func TestDispatch_LateSuccessOutputHonored(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistryWithTimeout(cfg, 30*time.Millisecond)
	reg.Register(&slowSuccessHandler{event: EventNotification, sleep: 60 * time.Millisecond})

	got, err := reg.Dispatch(context.Background(), EventNotification, &HookInput{
		SessionID:     "late-success",
		CWD:           "/tmp",
		HookEventName: string(EventNotification),
	})
	if err != nil {
		t.Fatalf("Dispatch() error = %v, want nil (late success must be honored)", err)
	}
	if got == nil || got.SystemMessage != "late but valid" {
		t.Errorf("Dispatch() output = %+v, want the handler's own output", got)
	}
}

// TestDispatch_CtxErrorStillMapsToTimeout keeps the existing timeout contract:
// a handler that RETURNS the context error still maps to ErrHookTimeout.
func TestDispatch_CtxErrorStillMapsToTimeout(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistryWithTimeout(cfg, 20*time.Millisecond)
	reg.Register(&ctxAwareHandler{event: EventNotification, sleep: 200 * time.Millisecond})

	_, err := reg.Dispatch(context.Background(), EventNotification, &HookInput{
		SessionID:     "ctx-timeout",
		CWD:           "/tmp",
		HookEventName: string(EventNotification),
	})
	if !errors.Is(err, ErrHookTimeout) {
		t.Fatalf("Dispatch() error = %v, want errors.Is(ErrHookTimeout)", err)
	}
}

// TestDispatch_MergeAccumulatesAdditionalContext is the G5 regression test:
// additionalContext from EVERY non-blocking hook is kept ("\n"-joined) —
// previously only the first survived.
func TestDispatch_MergeAccumulatesAdditionalContext(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)
	reg.Register(&mockHandler{event: EventUserPromptSubmit, output: &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "UserPromptSubmit",
			AdditionalContext: "context-from-first",
		},
	}})
	reg.Register(&mockHandler{event: EventUserPromptSubmit, output: &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "UserPromptSubmit",
			AdditionalContext: "context-from-second",
		},
	}})

	got, err := reg.Dispatch(context.Background(), EventUserPromptSubmit, &HookInput{
		SessionID:     "merge-ac",
		CWD:           "/tmp",
		HookEventName: string(EventUserPromptSubmit),
	})
	if err != nil {
		t.Fatalf("Dispatch() error: %v", err)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("Dispatch() returned nil output/HSO")
	}
	ac := got.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ac, "context-from-first") || !strings.Contains(ac, "context-from-second") {
		t.Errorf("additionalContext = %q, want BOTH hooks' contexts kept", ac)
	}
	if !strings.Contains(ac, "context-from-first\ncontext-from-second") {
		t.Errorf("additionalContext = %q, want newline-joined accumulation", ac)
	}
}

// TestDispatch_MergeLastNonNilWinsAndPropagation covers the remaining G5
// merge rules: updatedInput/updatedToolOutput last-non-nil wins; stopReason,
// continue:false and retry propagate from any handler.
func TestDispatch_MergeLastNonNilWinsAndPropagation(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)
	reg.Register(&mockHandler{event: EventPostToolUse, output: &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "PostToolUse",
			UpdatedInput:      json.RawMessage(`{"v":1}`),
			UpdatedToolOutput: "first-rewrite",
		},
	}})
	halted := &HookOutput{
		StopReason: "halt requested",
		Retry:      true,
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "PostToolUse",
			UpdatedInput:      json.RawMessage(`{"v":2}`),
			UpdatedToolOutput: "second-rewrite",
		},
	}
	halted.SetContinue(false)
	reg.Register(&mockHandler{event: EventPostToolUse, output: halted})

	got, err := reg.Dispatch(context.Background(), EventPostToolUse, &HookInput{
		SessionID:     "merge-last-wins",
		CWD:           "/tmp",
		HookEventName: string(EventPostToolUse),
	})
	if err != nil {
		t.Fatalf("Dispatch() error: %v", err)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("Dispatch() returned nil output/HSO")
	}
	if !bytes.Equal(got.HookSpecificOutput.UpdatedInput, []byte(`{"v":2}`)) {
		t.Errorf("UpdatedInput = %s, want last-non-nil {\"v\":2}", got.HookSpecificOutput.UpdatedInput)
	}
	if got.HookSpecificOutput.UpdatedToolOutput != "second-rewrite" {
		t.Errorf("UpdatedToolOutput = %q, want last-non-nil second-rewrite", got.HookSpecificOutput.UpdatedToolOutput)
	}
	if got.StopReason != "halt requested" {
		t.Errorf("StopReason = %q, want propagated halt reason", got.StopReason)
	}
	if got.Continue == nil || *got.Continue != false {
		t.Errorf("Continue = %v, want explicit false propagated", got.Continue)
	}
	if !got.Retry {
		t.Error("Retry = false, want propagated true")
	}
}

// TestHandlers_RaceSafeCopy is the G6 regression test: Handlers() must return
// a copy under the mutex so concurrent Register cannot race the returned slice.
// Run with -race to exercise the guarantee.
func TestHandlers_RaceSafeCopy(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 50 {
				reg.Register(&mockHandler{event: EventNotification, output: &HookOutput{}})
			}
		}()
		go func() {
			defer wg.Done()
			for range 50 {
				_ = reg.Handlers(EventNotification)
			}
		}()
	}
	wg.Wait()

	// Mutating the returned copy must not affect the registry.
	snapshot := reg.Handlers(EventNotification)
	if len(snapshot) == 0 {
		t.Fatal("expected registered handlers")
	}
	snapshot[0] = nil
	if reg.Handlers(EventNotification)[0] == nil {
		t.Error("Handlers() must return a copy, not the internal slice")
	}
}

// TestReadInput_SizeCap is the G7 regression test: stdin larger than
// maxHookInputBytes is truncated (fails JSON parse → ErrHookInvalidInput)
// instead of being read unboundedly.
func TestReadInput_SizeCap(t *testing.T) {
	t.Parallel()

	proto := NewProtocol()
	// A single valid JSON object bigger than the cap.
	huge := `{"session_id":"s","cwd":"/tmp","hook_event_name":"PreToolUse","pad":"` +
		strings.Repeat("x", maxHookInputBytes) + `"}`
	_, err := proto.ReadInput(strings.NewReader(huge))
	if !errors.Is(err, ErrHookInvalidInput) {
		t.Fatalf("ReadInput(oversized) error = %v, want ErrHookInvalidInput (truncated at cap)", err)
	}

	// A payload under the cap still parses.
	ok := `{"session_id":"s","cwd":"/tmp","hook_event_name":"PreToolUse"}`
	input, err := proto.ReadInput(strings.NewReader(ok))
	if err != nil {
		t.Fatalf("ReadInput(normal) error = %v", err)
	}
	if input.SessionID != "s" {
		t.Errorf("SessionID = %q, want %q", input.SessionID, "s")
	}
}

// TestReadInput_ProjectDirPopulated is the G13 regression test: project_dir
// is NOT an official stdin field, so ReadInput must populate it env-first
// (CLAUDE_PROJECT_DIR) with cwd fallback — otherwise SessionStart/Stop side
// effects gated on ProjectDir never fire under the official runtime.
func TestReadInput_ProjectDirPopulated(t *testing.T) {
	proto := NewProtocol()

	t.Run("env takes precedence", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", "/env/project/root")
		input, err := proto.ReadInput(strings.NewReader(
			`{"session_id":"s","cwd":"/some/worktree","hook_event_name":"SessionStart","source":"startup"}`))
		if err != nil {
			t.Fatalf("ReadInput error: %v", err)
		}
		if input.ProjectDir != "/env/project/root" {
			t.Errorf("ProjectDir = %q, want env value /env/project/root", input.ProjectDir)
		}
	})

	t.Run("cwd fallback when env unset", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		input, err := proto.ReadInput(strings.NewReader(
			`{"session_id":"s","cwd":"/official/cwd","hook_event_name":"SessionStart","source":"startup"}`))
		if err != nil {
			t.Fatalf("ReadInput error: %v", err)
		}
		if input.ProjectDir != "/official/cwd" {
			t.Errorf("ProjectDir = %q, want cwd fallback /official/cwd", input.ProjectDir)
		}
	})

	t.Run("explicit project_dir preserved", func(t *testing.T) {
		t.Setenv("CLAUDE_PROJECT_DIR", "/env/project/root")
		input, err := proto.ReadInput(strings.NewReader(
			`{"session_id":"s","cwd":"/c","project_dir":"/explicit","hook_event_name":"SessionStart"}`))
		if err != nil {
			t.Fatalf("ReadInput error: %v", err)
		}
		if input.ProjectDir != "/explicit" {
			t.Errorf("ProjectDir = %q, want explicit value preserved", input.ProjectDir)
		}
	})
}

// TestReadInput_OfficialFieldNames is the G3/G17 regression test: the
// official stdin field names decode into HookInput —
// config_source (ConfigChange), type (Notification), task_name
// (TaskCreated/TaskCompleted), agent_type (TeammateIdle).
func TestReadInput_OfficialFieldNames(t *testing.T) {
	t.Parallel()

	proto := NewProtocol()

	t.Run("ConfigChange config_source", func(t *testing.T) {
		t.Parallel()
		input, err := proto.ReadInput(strings.NewReader(
			`{"session_id":"s","cwd":"/tmp","hook_event_name":"ConfigChange","config_source":"project_settings"}`))
		if err != nil {
			t.Fatalf("ReadInput error: %v", err)
		}
		if input.ConfigSource != "project_settings" {
			t.Errorf("ConfigSource = %q, want project_settings", input.ConfigSource)
		}
	})

	t.Run("Notification type", func(t *testing.T) {
		t.Parallel()
		input, err := proto.ReadInput(strings.NewReader(
			`{"session_id":"s","cwd":"/tmp","hook_event_name":"Notification","type":"permission_request","message":"m"}`))
		if err != nil {
			t.Fatalf("ReadInput error: %v", err)
		}
		if input.Type != "permission_request" {
			t.Errorf("Type = %q, want permission_request", input.Type)
		}
	})

	t.Run("TaskCompleted task_name", func(t *testing.T) {
		t.Parallel()
		input, err := proto.ReadInput(strings.NewReader(
			`{"session_id":"s","cwd":"/tmp","hook_event_name":"TaskCompleted","task_name":"Implement SPEC-X-001"}`))
		if err != nil {
			t.Fatalf("ReadInput error: %v", err)
		}
		if input.TaskName != "Implement SPEC-X-001" {
			t.Errorf("TaskName = %q, want Implement SPEC-X-001", input.TaskName)
		}
	})

	t.Run("TeammateIdle agent_type", func(t *testing.T) {
		t.Parallel()
		input, err := proto.ReadInput(strings.NewReader(
			`{"session_id":"s","cwd":"/tmp","hook_event_name":"TeammateIdle","agent_type":"researcher"}`))
		if err != nil {
			t.Fatalf("ReadInput error: %v", err)
		}
		if input.AgentType != "researcher" {
			t.Errorf("AgentType = %q, want researcher", input.AgentType)
		}
	})
}
