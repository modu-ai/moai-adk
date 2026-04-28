package permission

import (
	"encoding/json"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestNewPermissionResolver(t *testing.T) {
	resolver := NewPermissionResolver()
	if resolver == nil {
		t.Fatal("NewPermissionResolver() returned nil")
	}
	if len(resolver.preAllowlist) == 0 {
		t.Error("NewPermissionResolver() preAllowlist is empty")
	}
}

func TestPermissionResolver_Resolve_PreAllowlist(t *testing.T) {
	// AC-V3R2-RT-002-02: Given a Bash(go test ./...) invocation with only the pre-allowlist active,
	// When resolved, Then the result is {PermissionDecision: "allow", ResolvedBy: SrcBuiltin, Origin: "pre-allowlist"}.

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeDefault,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		RulesByTier:    make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("go test ./..."), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionAllow {
		t.Errorf("Resolve() Decision = %v, want %v", result.Decision, DecisionAllow)
	}
	if result.ResolvedBy != config.SrcBuiltin {
		t.Errorf("Resolve() ResolvedBy = %v, want %v", result.ResolvedBy, config.SrcBuiltin)
	}
	if result.Origin != "pre-allowlist" {
		t.Errorf("Resolve() Origin = %v, want 'pre-allowlist'", result.Origin)
	}
}

func TestPermissionResolver_Resolve_ProjectDeny(t *testing.T) {
	// AC-V3R2-RT-002-01: Given a Bash(rm -rf /) invocation and a project-level SrcProject rule denying Bash(rm*:*)
	// When resolved, Then the result is {PermissionDecision: "deny", ResolvedBy: SrcProject, Origin: ".claude/settings.json"}

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeDefault,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcProject: {
				{
					Pattern: "Bash(rm:*)",
					Action:  DecisionDeny,
					Source:  config.SrcProject,
					Origin:  ".claude/settings.json",
				},
			},
		},
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("rm -rf /"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionDeny {
		t.Errorf("Resolve() Decision = %v, want %v", result.Decision, DecisionDeny)
	}
	if result.ResolvedBy != config.SrcProject {
		t.Errorf("Resolve() ResolvedBy = %v, want %v", result.ResolvedBy, config.SrcProject)
	}
	if result.Origin != ".claude/settings.json" {
		t.Errorf("Resolve() Origin = %v, want '.claude/settings.json'", result.Origin)
	}
}

func TestPermissionResolver_Resolve_PolicyDenyWins(t *testing.T) {
	// AC-V3R2-RT-002-13: Given SrcPolicy contains deny Bash(curl:*) and SrcProject contains allow Bash(curl:*)
	// When resolved, Then policy tier wins and the tool is denied.

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeDefault,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcPolicy: {
				{
					Pattern: "Bash(curl:*)",
					Action:  DecisionDeny,
					Source:  config.SrcPolicy,
					Origin:  "/etc/moai/settings.json",
				},
			},
			config.SrcProject: {
				{
					Pattern: "Bash(curl:*)",
					Action:  DecisionAllow,
					Source:  config.SrcProject,
					Origin:  ".claude/settings.json",
				},
			},
		},
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("curl https://example.com"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionDeny {
		t.Errorf("Resolve() Decision = %v, want %v (policy should override project)", result.Decision, DecisionDeny)
	}
	if result.ResolvedBy != config.SrcPolicy {
		t.Errorf("Resolve() ResolvedBy = %v, want %v", result.ResolvedBy, config.SrcPolicy)
	}
}

func TestPermissionResolver_Resolve_PlanModeDeniesWrites(t *testing.T) {
	// AC-V3R2-RT-002-06: Given an agent in plan mode, When it invokes Write(path: /tmp/x)
	// Then the resolver returns deny with Origin: "plan mode denies writes"

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModePlan,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		RulesByTier:    make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("Write", json.RawMessage("/tmp/x"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionDeny {
		t.Errorf("Resolve() Decision = %v, want %v", result.Decision, DecisionDeny)
	}
	if result.Origin != "plan mode denies writes" {
		t.Errorf("Resolve() Origin = %v, want 'plan mode denies writes'", result.Origin)
	}
}

func TestPermissionResolver_Resolve_BypassPermissions(t *testing.T) {
	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeBypassPermissions,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		StrictMode:     false,
		RulesByTier:    make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("rm -rf /"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionAllow {
		t.Errorf("Resolve() Decision = %v, want %v", result.Decision, DecisionAllow)
	}
	if result.Origin != "bypassPermissions mode" {
		t.Errorf("Resolve() Origin = %v, want 'bypassPermissions mode'", result.Origin)
	}
}

func TestPermissionResolver_Resolve_BypassPermissionsInFork(t *testing.T) {
	// REQ-V3R2-RT-002-021: Fork agents with bypassPermissions are degraded to bubble/ask
	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeBypassPermissions,
		IsFork:         true,
		ParentAvailable: true,
		ForkDepth:      1,
		IsInteractive:  true,
		StrictMode:     false,
		RulesByTier:    make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("rm -rf /"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionAsk {
		t.Errorf("Resolve() Decision = %v, want %v (fork bypass should be degraded to ask)", result.Decision, DecisionAsk)
	}
	if result.SystemMessage == "" {
		t.Error("Resolve() SystemMessage should be set for fork bypass degradation")
	}
}

func TestPermissionResolver_Resolve_BubbleMode(t *testing.T) {
	// AC-V3R2-RT-002-03: Given an agent spawned with permissionMode: bubble under a parent terminal session
	// When it attempts Write(path: /tmp/test.txt), Then the AskUserQuestion prompt appears in the parent session's channel

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeBubble,
		IsFork:         true,
		ParentAvailable: true,
		ForkDepth:      1,
		IsInteractive:  true,
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcProject: {
				{
					Pattern: "Write(*)",
					Action:  DecisionAsk,
					Source:  config.SrcProject,
					Origin:  ".claude/settings.json",
				},
			},
		},
	}

	result, err := resolver.Resolve("Write", json.RawMessage("/tmp/test.txt"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionAsk {
		t.Errorf("Resolve() Decision = %v, want %v", result.Decision, DecisionAsk)
	}
	if result.SystemMessage == "" {
		t.Error("Resolve() SystemMessage should indicate bubble mode routing")
	}
}

func TestPermissionResolver_Resolve_BubbleModeParentUnavailable(t *testing.T) {
	// AC-V3R2-RT-002-08: Given a fork agent with permissionMode: bubble and parent session already closed
	// When a non-allowlisted tool fires, Then the call is denied with systemMessage "Bubble target parent unavailable"

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeBubble,
		IsFork:         true,
		ParentAvailable: false,
		ForkDepth:      1,
		IsInteractive:  true,
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcProject: {
				{
					Pattern: "Write(*)",
					Action:  DecisionAsk,
					Source:  config.SrcProject,
					Origin:  ".claude/settings.json",
				},
			},
		},
	}

	result, err := resolver.Resolve("Write", json.RawMessage("/tmp/test.txt"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionDeny {
		t.Errorf("Resolve() Decision = %v, want %v", result.Decision, DecisionDeny)
	}
	if result.SystemMessage != "Bubble target parent unavailable — decision deferred" {
		t.Errorf("Resolve() SystemMessage = %v, want 'Bubble target parent unavailable — decision deferred'", result.SystemMessage)
	}
}

func TestPermissionResolver_Resolve_NonInteractive(t *testing.T) {
	// AC-V3R2-RT-002-15: Given non-interactive mode (CI=1), When resolver would return ask
	// Then the result is instead deny with log entry in .moai/logs/permission.log

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeDefault,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  false,
		RulesByTier:    make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("Write", json.RawMessage("/tmp/test.txt"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionDeny {
		t.Errorf("Resolve() Decision = %v, want %v (non-interactive should convert ask to deny)", result.Decision, DecisionDeny)
	}
	if result.Origin != "no matching rule (default)" {
		t.Errorf("Resolve() Origin = %v, want 'no matching rule (default)'", result.Origin)
	}
}

func TestPermissionResolver_Resolve_HookOverride(t *testing.T) {
	// AC-V3R2-RT-002-04: Given a PreToolUse hook returns PermissionDecision: "allow" via HookResponse
	// When the resolver is invoked, Then the hook-contributed decision overrides any SrcSession/SrcSkill/SrcBuiltin rules

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeDefault,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		HookResponse: &hook.HookResponse{
			PermissionDecision: hook.PermissionDecisionAllow,
		},
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcSession: {
				{
					Pattern: "Bash(curl:*)",
					Action:  DecisionDeny,
					Source:  config.SrcSession,
					Origin:  "session",
				},
			},
		},
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("curl https://example.com"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionAllow {
		t.Errorf("Resolve() Decision = %v, want %v (hook should override session rules)", result.Decision, DecisionAllow)
	}
	if result.Origin != "PreToolUse hook" {
		t.Errorf("Resolve() Origin = %v, want 'PreToolUse hook'", result.Origin)
	}
}

func TestPermissionResolver_Resolve_HookUpdatedInput(t *testing.T) {
	// AC-V3R2-RT-002-10: Given a PreToolUse hook mutates input via UpdatedInput: {file_path: "/safe/path"}
	// When resolver re-matches, Then allowlist patterns against /safe/path are consulted (not the original path)

	resolver := NewPermissionResolver()
	updatedInput := json.RawMessage(`/safe/path`)

	ctx := ResolveContext{
		Mode:           ModeDefault,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		HookResponse: &hook.HookResponse{
			// No PermissionDecision - only UpdatedInput
			UpdatedInput: updatedInput,
		},
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcProject: {
				{
					Pattern: "Write(/safe/*)",
					Action:  DecisionAllow,
					Source:  config.SrcProject,
					Origin:  ".claude/settings.json",
				},
			},
		},
	}

	result, err := resolver.Resolve("Write", json.RawMessage(`/unsafe/path`), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionAllow {
		t.Errorf("Resolve() Decision = %v, want %v", result.Decision, DecisionAllow)
	}
	if len(result.UpdatedInput) == 0 {
		t.Error("Resolve() UpdatedInput should be set from hook response")
	}
}

func TestPermissionResolver_Resolve_ForkDepthExceedsLimit(t *testing.T) {
	// AC-V3R2-RT-002-14: Given a fork at depth 4, When any non-plan mode agent fires
	// Then the effective mode is treated as bubble and a warning SystemMessage is emitted

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeAcceptEdits,
		IsFork:         true,
		ParentAvailable: true,
		ForkDepth:      4,
		IsInteractive:  true,
		RulesByTier:    make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("Write", json.RawMessage("/tmp/test.txt"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.SystemMessage == "" {
		t.Error("Resolve() SystemMessage should warn about fork depth exceeding limit")
	}
}

func TestPermissionResolver_ValidateMode(t *testing.T) {
	// AC-V3R2-RT-002-07: Given security.yaml sets permission.strict_mode: true
	// When an agent with permissionMode: bypassPermissions is spawned, Then spawn fails with error PermissionModeRejected

	resolver := NewPermissionResolver()

	tests := []struct {
		name        string
		mode        PermissionMode
		isFork      bool
		strictMode  bool
		forkDepth   int
		wantErr     bool
		errContains string
	}{
		{
			name:       "bypassPermissions rejected in strict mode",
			mode:       ModeBypassPermissions,
			strictMode: true,
			isFork:     false,
			wantErr:    true,
			errContains: "not allowed in strict mode",
		},
		{
			name:       "bypassPermissions allowed in normal mode",
			mode:       ModeBypassPermissions,
			strictMode: false,
			isFork:     false,
			wantErr:    false,
		},
		{
			name:       "bubble mode always allowed",
			mode:       ModeBubble,
			strictMode: true,
			isFork:     true,
			wantErr:    false,
		},
		{
			name:        "fork depth exceeds limit",
			mode:        ModeAcceptEdits,
			isFork:      true,
			forkDepth:   5,
			wantErr:     true,
			errContains: "degraded to bubble",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resolver.ValidateMode(tt.mode, tt.isFork, tt.strictMode, tt.forkDepth)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateMode() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestPermissionResolver_Resolve_TraceGeneration(t *testing.T) {
	// AC-V3R2-RT-002-05: Given moai doctor permission --trace Bash "go build"
	// When invoked, Then stdout contains a JSON trace enumerating all 8 tiers inspected

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:           ModeDefault,
		IsFork:         false,
		ParentAvailable: true,
		IsInteractive:  true,
		RulesByTier:    make(map[config.Source][]PermissionRule), // No rules - will walk all tiers
	}

	result, err := resolver.Resolve("UnknownTool", json.RawMessage("some input"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	traceJSON, err := result.ExportTrace()
	if err != nil {
		t.Fatalf("ExportTrace() error = %v", err)
	}

	// Verify trace contains all 8 tiers
	var trace ResolutionTrace
	if err := json.Unmarshal([]byte(traceJSON), &trace); err != nil {
		t.Fatalf("Unmarshal trace error = %v", err)
	}

	if len(trace.Tries) != 8 {
		t.Errorf("Trace contains %d tiers, want 8", len(trace.Tries))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
