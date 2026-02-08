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
}
