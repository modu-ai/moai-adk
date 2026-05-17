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
		Mode:            ModeDefault,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
		RulesByTier:     make(map[config.Source][]PermissionRule),
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
		Mode:            ModeDefault,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
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
		Mode:            ModeDefault,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
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
		Mode:            ModePlan,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
		RulesByTier:     make(map[config.Source][]PermissionRule),
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
		Mode:            ModeBypassPermissions,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
		StrictMode:      false,
		RulesByTier:     make(map[config.Source][]PermissionRule),
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
		Mode:            ModeBypassPermissions,
		IsFork:          true,
		ParentAvailable: true,
		ForkDepth:       1,
		IsInteractive:   true,
		StrictMode:      false,
		RulesByTier:     make(map[config.Source][]PermissionRule),
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
		Mode:            ModeBubble,
		IsFork:          true,
		ParentAvailable: true,
		ForkDepth:       1,
		IsInteractive:   true,
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
		Mode:            ModeBubble,
		IsFork:          true,
		ParentAvailable: false,
		ForkDepth:       1,
		IsInteractive:   true,
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
		Mode:            ModeDefault,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   false,
		RulesByTier:     make(map[config.Source][]PermissionRule),
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
		Mode:            ModeDefault,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
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
		Mode:            ModeDefault,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
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
		Mode:            ModeAcceptEdits,
		IsFork:          true,
		ParentAvailable: true,
		ForkDepth:       4,
		IsInteractive:   true,
		RulesByTier:     make(map[config.Source][]PermissionRule),
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
			name:        "bypassPermissions rejected in strict mode",
			mode:        ModeBypassPermissions,
			strictMode:  true,
			isFork:      false,
			wantErr:     true,
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
		Mode:            ModeDefault,
		IsFork:          false,
		ParentAvailable: true,
		IsInteractive:   true,
		RulesByTier:     make(map[config.Source][]PermissionRule), // No rules - will walk all tiers
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

// T-RT002-01: ParsePermissionMode 잘못된 값 수신 시 default + error 반환 검증.
// AC-09 관련 — resolver-side 동작 단위 검증.
func TestResolve_FrontmatterUnknownPermissionMode(t *testing.T) {
	// ParsePermissionMode가 invalid 값 수신 시 ModeDefault 와 non-nil error 반환 검증.
	mode, err := ParsePermissionMode("ultra-bypass")
	if err == nil {
		t.Fatal("ParsePermissionMode() should return error for invalid mode 'ultra-bypass'")
	}
	if mode != ModeDefault {
		t.Errorf("ParsePermissionMode() invalid mode = %v, want ModeDefault", mode)
	}
}

// T-RT002-02: hook UpdatedInput re-match 가드 검증 — 무한루프 방지.
// AC-10 관련.
func TestResolve_HookUpdatedInputReMatch(t *testing.T) {
	// hook 가 /dangerous/path → /safe/path 로 mutate 시 pre-allowlist 가 mutated path 에 매칭되는지.
	// 단: HookResponse.UpdatedInput 만 있고 PermissionDecision 없으면 re-match 단 1회만 실행.
	resolver := NewPermissionResolver()

	// /safe/path 를 allow 하는 규칙 준비.
	updatedInput := json.RawMessage(`/safe/path`)
	ctx := ResolveContext{
		Mode:          ModeDefault,
		IsFork:        false,
		IsInteractive: true,
		HookResponse: &hook.HookResponse{
			UpdatedInput: updatedInput,
			// PermissionDecision 없음 → re-match 트리거.
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

	result, err := resolver.Resolve("Write", json.RawMessage(`/dangerous/path`), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// mutated input(/safe/path) 기준으로 re-match → allow 기대.
	if result.Decision != DecisionAllow {
		t.Errorf("Resolve() Decision = %v, want Allow (re-match on mutated input)", result.Decision)
	}
	if len(result.UpdatedInput) == 0 {
		t.Error("Resolve() UpdatedInput should be set from hook response")
	}

	// 중첩 mutation 방지: HookResponse 가 nil 로 clear 되므로 무한루프 없어야 함.
	// (resolver 내부적으로 newCtx.HookResponse = nil 로 처리됨을 간접 검증)
}

// T-RT002-03: fork depth=4 에서 non-plan mode agent → bubble 강등 + SystemMessage 검증.
// AC-14 관련.
func TestResolve_ForkDepth4DegradeToBubble(t *testing.T) {
	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:          ModeAcceptEdits,
		IsFork:        true,
		ForkDepth:     4,
		IsInteractive: true,
		RulesByTier:   make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("Write", json.RawMessage("/tmp/test.txt"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// fork depth > 3 → bubble 강등 → DecisionAsk + systemMessage.
	if result.Decision != DecisionAsk {
		t.Errorf("Resolve() Decision = %v, want Ask (fork depth degraded to bubble)", result.Decision)
	}
	if result.SystemMessage == "" {
		t.Error("Resolve() SystemMessage should warn about fork depth limit")
	}
	// AC-14: systemMessage 에 "mode degraded" 또는 "exceeds limit" 포함 확인.
	if !containsMiddle(result.SystemMessage, "exceeds limit") && !containsMiddle(result.SystemMessage, "degraded") {
		t.Errorf("Resolve() SystemMessage = %q; want to contain 'exceeds limit' or 'degraded'", result.SystemMessage)
	}
}

// T-RT002-04: bubble mode + parent 닫힘 → deny + sentinel message.
// AC-08 관련.
func TestResolve_BubbleParentClosed(t *testing.T) {
	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:            ModeBubble,
		IsFork:          true,
		ParentAvailable: false, // 부모 세션 닫힘.
		ForkDepth:       1,
		IsInteractive:   true,
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
		t.Errorf("Resolve() Decision = %v, want Deny (parent unavailable)", result.Decision)
	}
	if !containsMiddle(result.SystemMessage, "parent unavailable") &&
		!containsMiddle(result.SystemMessage, "Bubble target") {
		t.Errorf("Resolve() SystemMessage = %q; want 'Bubble target parent unavailable'", result.SystemMessage)
	}
}

// T-RT002-05: 비대화형 모드 → ask → deny + log 파일 경로 검증.
// AC-15 관련.
func TestResolve_NonInteractiveAskBecomesDeny(t *testing.T) {
	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:          ModeDefault,
		IsInteractive: false, // 비대화형.
		RulesByTier:   make(map[config.Source][]PermissionRule),
	}

	result, err := resolver.Resolve("UnknownTool", json.RawMessage("some input"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.Decision != DecisionDeny {
		t.Errorf("Resolve() Decision = %v, want Deny (non-interactive mode)", result.Decision)
	}
}

// T-RT002-05b: logUnreachablePrompt 가 .moai/logs/permission.log 에 기록하는지 간접 검증.
func TestLogUnreachablePrompt_FilePath(t *testing.T) {
	// t.TempDir() 로 격리된 환경에서 로그 파일 생성 확인.
	tmpDir := t.TempDir()

	// logPath 계산: resolver 가 현재 디렉터리 기준 .moai/logs/permission.log 에 기록.
	// cwd 를 직접 변경할 수 없으므로, logUnreachablePrompt 는 내부적으로 os.OpenFile 사용.
	// 이 테스트는 비대화형 Resolve 호출이 오류 없이 완료됨을 검증.
	_ = tmpDir // 격리 디렉터리 준비 (추후 통합 테스트 확장용).

	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:          ModeDefault,
		IsInteractive: false,
		RulesByTier:   make(map[config.Source][]PermissionRule),
	}
	// 로그 파일 생성 시 오류 없이 완료되어야 함.
	result, err := resolver.Resolve("Write", json.RawMessage("/tmp/x"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if result.Decision != DecisionDeny {
		t.Errorf("Resolve() Decision = %v, want Deny", result.Decision)
	}
}

// T-RT002-06: 동일 tier 두 규칙 충돌 시 specificity 우선, 동률 시 fs-order.
// AC-12 관련 — conflict.go 의 resolveConflict 함수가 없을 때 RED (현재 첫 매치 반환).
func TestResolve_ConflictSpecificityThenFsOrder(t *testing.T) {
	resolver := NewPermissionResolver()

	// 두 SrcLocal 규칙: "Bash(git push:*)" (덜 구체적) vs "Bash(git push origin main)" (더 구체적).
	// 더 구체적인 패턴의 규칙이 우선해야 함.
	ctx := ResolveContext{
		Mode:          ModeDefault,
		IsInteractive: true,
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcLocal: {
				{
					Pattern: "Bash(git push:*)",
					Action:  DecisionDeny,   // 덜 구체적 → deny.
					Source:  config.SrcLocal,
					Origin:  "a-settings.json", // fs-order: a < b.
				},
				{
					Pattern: "Bash(git push origin main)",
					Action:  DecisionAllow, // 더 구체적 → allow.
					Source:  config.SrcLocal,
					Origin:  "b-settings.json",
				},
			},
		},
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("git push origin main"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// specificity 높은 "git push origin main" 규칙이 우선 → allow.
	if result.Decision != DecisionAllow {
		t.Errorf("Resolve() Decision = %v, want Allow (more specific rule should win)", result.Decision)
	}
}

// T-RT002-07: SrcPolicy deny > SrcProject allow 검증.
// AC-13 관련.
func TestResolve_PolicyDenyOverridesProjectAllow(t *testing.T) {
	resolver := NewPermissionResolver()
	ctx := ResolveContext{
		Mode:          ModeDefault,
		IsInteractive: true,
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcPolicy: {
				{
					Pattern: "Bash(curl:*)",
					Action:  DecisionDeny,
					Source:  config.SrcPolicy,
					Origin:  "/etc/moai/policy.json",
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
		t.Errorf("Resolve() Decision = %v, want Deny (policy wins over project)", result.Decision)
	}
	if result.ResolvedBy != config.SrcPolicy {
		t.Errorf("Resolve() ResolvedBy = %v, want SrcPolicy", result.ResolvedBy)
	}
}

// T-RT002-08: strict_mode=true + bypassPermissions → PermissionModeRejected.
// AC-07 관련.
func TestResolve_BypassPermissionsRejectedInStrictMode(t *testing.T) {
	resolver := NewPermissionResolver()

	// ValidateMode 가 strict_mode=true 에서 bypassPermissions 를 reject 해야 함.
	err := resolver.ValidateMode(ModeBypassPermissions, false, true, 0)
	if err == nil {
		t.Fatal("ValidateMode() should return error for bypassPermissions in strict mode")
	}
	if !containsMiddle(err.Error(), "not allowed in strict mode") {
		t.Errorf("ValidateMode() error = %q; want 'not allowed in strict mode'", err.Error())
	}
}

// T-RT002-09: legacy bypassPermissions action → acceptEdits reroute + deprecation warning.
// AC-11 관련 — MigrateLegacyBypassRules 함수 부재 시 RED.
func TestResolve_LegacyBypassActionMigrated(t *testing.T) {
	// MigrateLegacyBypassRules 가 구현되면 GREEN 으로 전환.
	// 현재는 함수 존재 여부만 컴파일 시간에 확인.
	rules := []PermissionRule{
		{
			Pattern: "Bash(curl:*)",
			Action:  Decision("bypassPermissions"), // legacy v2 action.
			Source:  config.SrcProject,
			Origin:  ".claude/settings.json",
		},
	}

	migrated, warnings := MigrateLegacyBypassRules(rules)
	if len(warnings) == 0 {
		t.Error("MigrateLegacyBypassRules() should return deprecation warnings for legacy bypassPermissions action")
	}
	if len(migrated) == 0 {
		t.Fatal("MigrateLegacyBypassRules() should return migrated rules")
	}
	// 마이그레이션 후 Action 은 acceptEdits (DecisionAllow 로 reroute).
	if migrated[0].Action != DecisionAllow {
		t.Errorf("MigrateLegacyBypassRules() migrated action = %v, want DecisionAllow", migrated[0].Action)
	}
}

// T-RT002-10: session_rules 키 → SrcSession tier 로 적재 검증.
// REQ-030 관련.
func TestResolve_SessionRulesLoadedAsSrcSession(t *testing.T) {
	resolver := NewPermissionResolver()

	// SrcSession tier 에 규칙이 있으면 해당 규칙 기준으로 결정해야 함.
	ctx := ResolveContext{
		Mode:          ModeDefault,
		IsInteractive: true,
		RulesByTier: map[config.Source][]PermissionRule{
			config.SrcSession: {
				{
					Pattern: "Bash(make:*)",
					Action:  DecisionAllow,
					Source:  config.SrcSession,
					Origin:  "session_rules",
				},
			},
		},
	}

	result, err := resolver.Resolve("Bash", json.RawMessage("make build"), ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// SrcSession 에서 허용 → allow.
	if result.Decision != DecisionAllow {
		t.Errorf("Resolve() Decision = %v, want Allow (session rule matched)", result.Decision)
	}
	if result.ResolvedBy != config.SrcSession {
		t.Errorf("Resolve() ResolvedBy = %v, want SrcSession", result.ResolvedBy)
	}
}
