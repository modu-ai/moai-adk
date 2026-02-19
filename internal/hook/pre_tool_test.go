package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/unicode/norm"
)

func TestPreToolHandler_EventType(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewPreToolHandler(cfg, DefaultSecurityPolicy())

	if got := h.EventType(); got != EventPreToolUse {
		t.Errorf("EventType() = %q, want %q", got, EventPreToolUse)
	}
}

func TestPreToolHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		policy       *SecurityPolicy
		input        *HookInput
		wantDecision string
		wantReason   bool
	}{
		{
			name:   "allowed tool passes security check",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Read",
				ToolInput:     json.RawMessage(`{"file_path": "/tmp/test.go"}`),
			},
			wantDecision: DecisionAllow,
		},
		{
			name: "blocked tool is rejected",
			policy: &SecurityPolicy{
				BlockedTools: []string{"DangerousTool"},
			},
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "DangerousTool",
			},
			wantDecision: DecisionDeny,
			wantReason:   true,
		},
		{
			name:   "Bash tool with dangerous rm -rf / command is blocked",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": "rm -rf /"}`),
			},
			wantDecision: DecisionDeny,
			wantReason:   true,
		},
		{
			name:   "Bash tool with safe command is allowed",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": "go test ./..."}`),
			},
			wantDecision: DecisionAllow,
		},
		{
			name:   "empty tool name is allowed (no policy match)",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "",
			},
			wantDecision: DecisionAllow,
		},
		{
			name:   "nil policy allows everything",
			policy: nil,
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": "rm -rf /"}`),
			},
			wantDecision: DecisionAllow,
		},
		{
			name:   "Bash tool with dangerous fork bomb is blocked",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": ":(){ :|:& };:"}`),
			},
			wantDecision: DecisionBlock,
			wantReason:   true,
		},
		{
			name:   "tool with nil tool_input is allowed",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Read",
			},
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			h := NewPreToolHandler(cfg, tt.policy)

			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// PreToolUse uses hookSpecificOutput.permissionDecision per Claude Code protocol
			if got.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput is nil")
			}
			// Map expected decision to permissionDecision value
			// DecisionBlock maps to DecisionDeny for PreToolUse
			wantPerm := tt.wantDecision
			if wantPerm == DecisionBlock {
				wantPerm = DecisionDeny
			}
			if got.HookSpecificOutput.PermissionDecision != wantPerm {
				t.Errorf("PermissionDecision = %q, want %q", got.HookSpecificOutput.PermissionDecision, wantPerm)
			}
			if tt.wantReason && got.HookSpecificOutput.PermissionDecisionReason == "" {
				t.Error("expected non-empty PermissionDecisionReason for deny decision")
			}
		})
	}
}

// TestPreToolHandler_UnicodeNFDNFCPathNormalization verifies that the path
// traversal security check correctly handles Unicode NFD/NFC mismatches.
// On macOS, HFS+/APFS stores paths in NFD form, but Claude Code sends paths
// in NFC form via stdin JSON. Without normalization, filepath.Rel produces
// ".." prefixed results for non-ASCII paths (e.g., Korean), causing false
// "Path traversal detected" errors.
func TestPreToolHandler_UnicodeNFDNFCPathNormalization(t *testing.T) {
	t.Parallel()

	// Korean text "코딩" used to test Unicode normalization.
	// NFC form: each syllable is a single codepoint (U+CF54, U+B529)
	// NFD form: each syllable is decomposed into jamo (ㅋ+ㅗ+ㄷ+ㅣ+ㅇ)
	koreanNFC := norm.NFC.String("코딩")
	koreanNFD := norm.NFD.String("코딩")

	// Verify that our test data actually produces different byte sequences
	if koreanNFC == koreanNFD {
		t.Skip("NFC and NFD forms are identical on this platform; test is not meaningful")
	}

	// Create a temp directory structure to simulate the macOS path scenario
	tmpDir := t.TempDir()

	// Simulate a project directory with NFD Korean path (as macOS would store it)
	projectDir := filepath.Join(tmpDir, koreanNFD+"_project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	tests := []struct {
		name         string
		projectDir   string
		filePath     string
		wantDecision string
		wantReason   bool
	}{
		{
			name:       "NFC file path within NFD project dir should NOT trigger path traversal",
			projectDir: filepath.Join(tmpDir, koreanNFD+"_project"),
			// Claude Code sends path in NFC form
			filePath:     filepath.Join(tmpDir, koreanNFC+"_project", "main.go"),
			wantDecision: DecisionAllow,
		},
		{
			name:       "NFD file path within NFC project dir should NOT trigger path traversal",
			projectDir: filepath.Join(tmpDir, koreanNFC+"_project"),
			// File path in NFD form (as macOS filesystem returns)
			filePath:     filepath.Join(tmpDir, koreanNFD+"_project", "main.go"),
			wantDecision: DecisionAllow,
		},
		{
			name:         "matching NFC paths should allow access",
			projectDir:   filepath.Join(tmpDir, koreanNFC+"_project"),
			filePath:     filepath.Join(tmpDir, koreanNFC+"_project", "src", "app.go"),
			wantDecision: DecisionAllow,
		},
		{
			name:         "matching NFD paths should allow access",
			projectDir:   filepath.Join(tmpDir, koreanNFD+"_project"),
			filePath:     filepath.Join(tmpDir, koreanNFD+"_project", "src", "app.go"),
			wantDecision: DecisionAllow,
		},
		{
			name:         "truly outside path should still be denied",
			projectDir:   filepath.Join(tmpDir, koreanNFC+"_project"),
			filePath:     filepath.Join(tmpDir, "other_project", "secret.go"),
			wantDecision: DecisionDeny,
			wantReason:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			// Use an empty policy to avoid deny/ask pattern matches on our test paths.
			// We only want to test the path traversal check.
			policy := &SecurityPolicy{}

			// Create handler with explicit projectDir
			handler := &preToolHandler{
				cfg:        cfg,
				policy:     policy,
				projectDir: tt.projectDir,
			}

			toolInput, err := json.Marshal(map[string]string{
				"file_path": tt.filePath,
			})
			if err != nil {
				t.Fatalf("failed to marshal tool input: %v", err)
			}

			input := &HookInput{
				SessionID:     "sess-unicode",
				CWD:           tmpDir,
				HookEventName: "PreToolUse",
				ToolName:      "Write",
				ToolInput:     json.RawMessage(toolInput),
			}

			ctx := context.Background()
			got, handleErr := handler.Handle(ctx, input)
			if handleErr != nil {
				t.Fatalf("unexpected error: %v", handleErr)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput is nil")
			}

			gotDecision := got.HookSpecificOutput.PermissionDecision
			if gotDecision != tt.wantDecision {
				t.Errorf("PermissionDecision = %q, want %q (projectDir=%q, filePath=%q)",
					gotDecision, tt.wantDecision, tt.projectDir, tt.filePath)
				// Log byte differences for debugging
				t.Logf("projectDir bytes: %x", []byte(tt.projectDir))
				t.Logf("filePath bytes:   %x", []byte(tt.filePath))
				t.Logf("NFC korean bytes: %x", []byte(koreanNFC))
				t.Logf("NFD korean bytes: %x", []byte(koreanNFD))
			}

			if tt.wantReason && got.HookSpecificOutput.PermissionDecisionReason == "" {
				t.Error("expected non-empty PermissionDecisionReason for deny decision")
			}
		})
	}
}

// TestUnicodeNFCNormalizationDirect verifies that NFC normalization makes
// NFD and NFC paths equivalent for filepath.Rel comparison.
func TestUnicodeNFCNormalizationDirect(t *testing.T) {
	t.Parallel()

	koreanNFC := norm.NFC.String("코딩")
	koreanNFD := norm.NFD.String("코딩")

	if koreanNFC == koreanNFD {
		t.Skip("NFC and NFD forms are identical; test is not meaningful")
	}

	// Simulate: project dir in NFD, file path in NFC
	projectDir := fmt.Sprintf("/Users/test/%s/project", koreanNFD)
	filePath := fmt.Sprintf("/Users/test/%s/project/main.go", koreanNFC)

	// Without normalization: filepath.Rel produces ".." prefix (BUG)
	rel, err := filepath.Rel(projectDir, filePath)
	if err == nil && !hasPathTraversalPrefix(rel) {
		t.Log("paths are already equivalent without normalization (unexpected)")
	}

	// With NFC normalization: filepath.Rel produces clean relative path (FIX)
	nfcProject := norm.NFC.String(projectDir)
	nfcFile := norm.NFC.String(filePath)

	relNorm, errNorm := filepath.Rel(nfcProject, nfcFile)
	if errNorm != nil {
		t.Fatalf("filepath.Rel failed after NFC normalization: %v", errNorm)
	}
	if hasPathTraversalPrefix(relNorm) {
		t.Errorf("after NFC normalization, rel = %q; should not start with '..'", relNorm)
	}

	want := "main.go"
	if relNorm != want {
		t.Errorf("after NFC normalization, rel = %q; want %q", relNorm, want)
	}

	_ = rel // suppress unused variable warning
}

// hasPathTraversalPrefix checks if a relative path starts with "..".
func hasPathTraversalPrefix(rel string) bool {
	return len(rel) >= 2 && rel[0] == '.' && rel[1] == '.'
}

func TestDefaultSecurityPolicy(t *testing.T) {
	t.Parallel()

	policy := DefaultSecurityPolicy()

	if policy == nil {
		t.Fatal("DefaultSecurityPolicy() returned nil")
	}
	if len(policy.DangerousBashPatterns) == 0 {
		t.Error("DangerousBashPatterns should not be empty")
	}
	if len(policy.DenyPatterns) == 0 {
		t.Error("DenyPatterns should not be empty")
	}
	if len(policy.AskPatterns) == 0 {
		t.Error("AskPatterns should not be empty")
	}
	if len(policy.SensitiveContentPatterns) == 0 {
		t.Error("SensitiveContentPatterns should not be empty")
	}
	if len(policy.AllowedExternalPaths) == 0 {
		t.Error("AllowedExternalPaths should not be empty (should include ~/.claude/plans)")
	}
}

func TestPreToolHandler_AllowedExternalPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	externalPlansDir := filepath.Join(tmpDir, "external-plans")
	if err := os.MkdirAll(externalPlansDir, 0o755); err != nil {
		t.Fatalf("create external plans dir: %v", err)
	}

	tests := []struct {
		name         string
		filePath     string
		allowedPaths []string
		wantDecision string
	}{
		{
			name:         "file inside project dir is allowed",
			filePath:     filepath.Join(projectDir, "main.go"),
			allowedPaths: nil,
			wantDecision: DecisionAllow,
		},
		{
			name:         "external path denied without allowlist",
			filePath:     filepath.Join(externalPlansDir, "plan.md"),
			allowedPaths: nil,
			wantDecision: DecisionDeny,
		},
		{
			name:         "external path allowed with allowlist",
			filePath:     filepath.Join(externalPlansDir, "plan.md"),
			allowedPaths: []string{externalPlansDir},
			wantDecision: DecisionAllow,
		},
		{
			name:         "external path subdirectory allowed with allowlist",
			filePath:     filepath.Join(externalPlansDir, "sub", "plan.md"),
			allowedPaths: []string{externalPlansDir},
			wantDecision: DecisionAllow,
		},
		{
			name:         "unrelated external path still denied with allowlist",
			filePath:     filepath.Join(tmpDir, "other", "secret.go"),
			allowedPaths: []string{externalPlansDir},
			wantDecision: DecisionDeny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			policy := &SecurityPolicy{
				AllowedExternalPaths: tt.allowedPaths,
			}
			handler := &preToolHandler{
				cfg:        cfg,
				policy:     policy,
				projectDir: projectDir,
			}

			toolInput, err := json.Marshal(map[string]string{
				"file_path": tt.filePath,
			})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			input := &HookInput{
				SessionID:     "sess-external",
				CWD:           projectDir,
				HookEventName: "PreToolUse",
				ToolName:      "Write",
				ToolInput:     json.RawMessage(toolInput),
			}

			ctx := context.Background()
			got, handleErr := handler.Handle(ctx, input)
			if handleErr != nil {
				t.Fatalf("unexpected error: %v", handleErr)
			}
			if got == nil || got.HookSpecificOutput == nil {
				t.Fatal("got nil output or nil HookSpecificOutput")
			}

			gotDecision := got.HookSpecificOutput.PermissionDecision
			if gotDecision != tt.wantDecision {
				t.Errorf("decision = %q, want %q (filePath=%q, allowed=%v)",
					gotDecision, tt.wantDecision, tt.filePath, tt.allowedPaths)
			}
		})
	}
}

func TestPreToolHandler_SensitiveContentDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		wantDecision string
	}{
		{
			name:         "RSA private key detected",
			content:      "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
			wantDecision: DecisionDeny,
		},
		{
			name:         "generic private key detected",
			content:      "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
			wantDecision: DecisionDeny,
		},
		{
			name:         "certificate detected",
			content:      "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
			wantDecision: DecisionDeny,
		},
		{
			name:         "OpenAI API key detected",
			content:      "api_key = \"sk-abcdefghijklmnopqrstuvwxyz12345678\"",
			wantDecision: DecisionDeny,
		},
		{
			name:         "GitHub personal access token detected",
			content:      "token: ghp_abcdefghijklmnopqrstuvwxyz1234567890",
			wantDecision: DecisionDeny,
		},
		{
			name:         "GitHub OAuth token detected",
			content:      "GITHUB_TOKEN=gho_abcdefghijklmnopqrstuvwxyz1234567890",
			wantDecision: DecisionDeny,
		},
		{
			name:         "AWS access key detected",
			content:      "aws_access_key_id = AKIAIOSFODNN7EXAMPLE",
			wantDecision: DecisionDeny,
		},
		{
			name:         "Slack token detected",
			content:      "SLACK_TOKEN=xoxb-some-token-value-here",
			wantDecision: DecisionDeny,
		},
		{
			name:         "safe content allowed",
			content:      "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}",
			wantDecision: DecisionAllow,
		},
		{
			name:         "empty content allowed",
			content:      "",
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectDir := t.TempDir()
			cfg := &mockConfigProvider{cfg: newTestConfig()}
			policy := DefaultSecurityPolicy()
			handler := &preToolHandler{
				cfg:        cfg,
				policy:     policy,
				projectDir: projectDir,
			}

			toolInput, err := json.Marshal(map[string]string{
				"file_path": filepath.Join(projectDir, "test.go"),
				"content":   tt.content,
			})
			if err != nil {
				t.Fatalf("marshal tool input: %v", err)
			}

			input := &HookInput{
				SessionID:     "sess-sensitive",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Write",
				ToolInput:     json.RawMessage(toolInput),
			}

			ctx := context.Background()
			got, handleErr := handler.Handle(ctx, input)
			if handleErr != nil {
				t.Fatalf("unexpected error: %v", handleErr)
			}
			if got == nil || got.HookSpecificOutput == nil {
				t.Fatal("got nil output or nil HookSpecificOutput")
			}

			gotDecision := got.HookSpecificOutput.PermissionDecision
			if gotDecision != tt.wantDecision {
				t.Errorf("PermissionDecision = %q, want %q", gotDecision, tt.wantDecision)
			}
		})
	}
}

func TestPreToolHandler_DenyPatternFileAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		filePath     string
		toolName     string
		wantDecision string
	}{
		{
			name:         "secrets.json is denied",
			filePath:     "/project/secrets.json",
			toolName:     "Write",
			wantDecision: DecisionDeny,
		},
		{
			name:         "credentials.yaml is denied",
			filePath:     "/project/credentials.yaml",
			toolName:     "Write",
			wantDecision: DecisionDeny,
		},
		{
			name:         "SSH key file is denied",
			filePath:     "/home/user/.ssh/id_rsa",
			toolName:     "Write",
			wantDecision: DecisionDeny,
		},
		{
			name:         "normal Go file is allowed",
			filePath:     "", // will be set per test
			toolName:     "Write",
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectDir := t.TempDir()
			cfg := &mockConfigProvider{cfg: newTestConfig()}
			policy := DefaultSecurityPolicy()
			handler := &preToolHandler{
				cfg:        cfg,
				policy:     policy,
				projectDir: projectDir,
			}

			filePath := tt.filePath
			if filePath == "" {
				filePath = filepath.Join(projectDir, "main.go")
			}

			toolInput, err := json.Marshal(map[string]string{
				"file_path": filePath,
			})
			if err != nil {
				t.Fatalf("marshal tool input: %v", err)
			}

			input := &HookInput{
				SessionID:     "sess-deny",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      tt.toolName,
				ToolInput:     json.RawMessage(toolInput),
			}

			ctx := context.Background()
			got, handleErr := handler.Handle(ctx, input)
			if handleErr != nil {
				t.Fatalf("unexpected error: %v", handleErr)
			}
			if got == nil || got.HookSpecificOutput == nil {
				t.Fatal("got nil output or nil HookSpecificOutput")
			}

			gotDecision := got.HookSpecificOutput.PermissionDecision
			if gotDecision != tt.wantDecision {
				t.Errorf("PermissionDecision = %q, want %q", gotDecision, tt.wantDecision)
			}
		})
	}
}

func TestPreToolHandler_AskBashPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		command      string
		wantDecision string
	}{
		{
			name:         "git reset --hard requires confirmation",
			command:      "git reset --hard HEAD~3",
			wantDecision: DecisionAsk,
		},
		{
			name:         "git clean -fd requires confirmation",
			command:      "git clean -fd",
			wantDecision: DecisionAsk,
		},
		{
			name:         "safe go test command is allowed",
			command:      "go test -race ./...",
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			policy := DefaultSecurityPolicy()
			handler := &preToolHandler{
				cfg:        cfg,
				policy:     policy,
				projectDir: t.TempDir(),
			}

			toolInput, err := json.Marshal(map[string]string{
				"command": tt.command,
			})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			input := &HookInput{
				SessionID:     "sess-ask",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(toolInput),
			}

			ctx := context.Background()
			got, handleErr := handler.Handle(ctx, input)
			if handleErr != nil {
				t.Fatalf("unexpected error: %v", handleErr)
			}
			if got == nil || got.HookSpecificOutput == nil {
				t.Fatal("got nil output or nil HookSpecificOutput")
			}

			gotDecision := got.HookSpecificOutput.PermissionDecision
			if gotDecision != tt.wantDecision {
				t.Errorf("PermissionDecision = %q, want %q", gotDecision, tt.wantDecision)
			}
		})
	}
}

func TestNewPreToolHandlerWithScanner_NilScanner(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	h := NewPreToolHandlerWithScanner(cfg, policy, nil)
	if h == nil {
		t.Fatal("NewPreToolHandlerWithScanner returned nil")
	}
	if h.EventType() != EventPreToolUse {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventPreToolUse)
	}
}

func TestPreToolHandler_InvalidJSON_ToolInput(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	handler := NewPreToolHandler(cfg, policy)

	input := &HookInput{
		SessionID:     "sess-invalid",
		CWD:           "/tmp",
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     json.RawMessage(`{invalid json}`),
	}

	ctx := context.Background()
	got, err := handler.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("got nil output or nil HookSpecificOutput")
	}
	// Invalid JSON should be treated as allow (cannot parse = no match)
	if got.HookSpecificOutput.PermissionDecision != DecisionAllow {
		t.Errorf("expected allow for invalid JSON, got %q", got.HookSpecificOutput.PermissionDecision)
	}
}

func TestPreToolHandler_DangerousBashPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
	}{
		{name: "terraform destroy", command: "terraform destroy"},
		{name: "docker system prune -a", command: "docker system prune -a"},
		{name: "git push force to main", command: "git push --force origin main"},
		{name: "dd to dev sda", command: "dd if=/dev/zero of=/dev/sda"},
		{name: "mkfs command", command: "mkfs.ext4 /dev/sda1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			policy := DefaultSecurityPolicy()
			handler := &preToolHandler{
				cfg:        cfg,
				policy:     policy,
				projectDir: t.TempDir(),
			}

			toolInput, err := json.Marshal(map[string]string{
				"command": tt.command,
			})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			input := &HookInput{
				SessionID:     "sess-danger",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(toolInput),
			}

			ctx := context.Background()
			got, handleErr := handler.Handle(ctx, input)
			if handleErr != nil {
				t.Fatalf("unexpected error: %v", handleErr)
			}
			if got == nil || got.HookSpecificOutput == nil {
				t.Fatal("got nil output or nil HookSpecificOutput")
			}

			// Dangerous patterns should be denied
			if got.HookSpecificOutput.PermissionDecision != DecisionDeny {
				t.Errorf("expected deny for %q, got %q", tt.command, got.HookSpecificOutput.PermissionDecision)
			}
		})
	}
}
