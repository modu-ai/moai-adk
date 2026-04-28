package permission

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestPermissionMode_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PermissionMode
		wantErr bool
	}{
		{
			name:  "default mode",
			input: "default",
			want:  ModeDefault,
		},
		{
			name:  "acceptEdits mode",
			input: "acceptEdits",
			want:  ModeAcceptEdits,
		},
		{
			name:  "bypassPermissions mode",
			input: "bypassPermissions",
			want:  ModeBypassPermissions,
		},
		{
			name:  "plan mode",
			input: "plan",
			want:  ModePlan,
		},
		{
			name:  "bubble mode",
			input: "bubble",
			want:  ModeBubble,
		},
		{
			name:    "invalid mode",
			input:   "ultra-bypass",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePermissionMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePermissionMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParsePermissionMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionMode_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		mode  PermissionMode
		valid bool
	}{
		{"default valid", ModeDefault, true},
		{"acceptEdits valid", ModeAcceptEdits, true},
		{"bypassPermissions valid", ModeBypassPermissions, true},
		{"plan valid", ModePlan, true},
		{"bubble valid", ModeBubble, true},
		{"invalid mode", PermissionMode("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.valid {
				t.Errorf("PermissionMode.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestPermissionRule_Matches(t *testing.T) {
	tests := []struct {
		name   string
		rule   PermissionRule
		tool   string
		input  string
		match  bool
	}{
		{
			name: "wildcard pattern matches all",
			rule: PermissionRule{Pattern: "*"},
			tool: "Bash",
			input: "go test",
			match: true,
		},
		{
			name: "exact tool match",
			rule: PermissionRule{Pattern: "Read(*)"},
			tool: "Read",
			input: "/path/to/file",
			match: true,
		},
		{
			name: "tool mismatch",
			rule: PermissionRule{Pattern: "Read(*)"},
			tool: "Write",
			input: "/path/to/file",
			match: false,
		},
		{
			name: "Bash go test pattern matches",
			rule: PermissionRule{Pattern: "Bash(go test:*)"},
			tool: "Bash",
			input: "go test ./...",
			match: true,
		},
		{
			name: "Bash go test pattern does not match other bash",
			rule: PermissionRule{Pattern: "Bash(go test:*)"},
			tool: "Bash",
			input: "rm -rf /",
			match: false,
		},
		{
			name: "prefix glob pattern",
			rule: PermissionRule{Pattern: "Write(/tmp/*)"},
			tool: "Write",
			input: "/tmp/test.txt",
			match: true,
		},
		{
			name: "prefix glob pattern no match",
			rule: PermissionRule{Pattern: "Write(/tmp/*)"},
			tool: "Write",
			input: "/home/test.txt",
			match: false,
		},
		{
			name: "suffix glob pattern",
			rule: PermissionRule{Pattern: "Read(*.go)"},
			tool: "Read",
			input: "main.go",
			match: true,
		},
		{
			name: "suffix glob pattern no match",
			rule: PermissionRule{Pattern: "Read(*.go)"},
			tool: "Read",
			input: "main.py",
			match: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.Matches(tt.tool, tt.input); got != tt.match {
				t.Errorf("PermissionRule.Matches() = %v, want %v", got, tt.match)
			}
		})
	}
}

func TestPreAllowlist(t *testing.T) {
	rules := PreAllowlist()

	// Check that pre-allowlist has the expected rules
	if len(rules) == 0 {
		t.Fatal("PreAllowlist() returned empty list")
	}

	// Verify some key rules exist
	expectedPatterns := []string{
		"Read(*)",
		"Glob(*)",
		"Grep(*)",
		"Bash(go test:*)",
		"Bash(golangci-lint run:*)",
		"Bash(ruff check:*)",
		"Bash(npm test:*)",
		"Bash(pytest:*)",
	}

	patterns := make(map[string]bool)
	for _, rule := range rules {
		patterns[rule.Pattern] = true
		if rule.Source != config.SrcBuiltin {
			t.Errorf("PreAllowlist() rule %s has source %v, want SrcBuiltin", rule.Pattern, rule.Source)
		}
		if rule.Action != DecisionAllow {
			t.Errorf("PreAllowlist() rule %s has action %v, want DecisionAllow", rule.Pattern, rule.Action)
		}
		if rule.Origin != "pre-allowlist" {
			t.Errorf("PreAllowlist() rule %s has origin %q, want 'pre-allowlist'", rule.Pattern, rule.Origin)
		}
	}

	for _, expected := range expectedPatterns {
		if !patterns[expected] {
			t.Errorf("PreAllowlist() missing expected pattern: %s", expected)
		}
	}
}

func TestIsWriteOperation(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		input    string
		isWrite  bool
	}{
		{
			name:    "Write tool is write",
			tool:    "Write",
			input:   "/tmp/test.txt",
			isWrite: true,
		},
		{
			name:    "Edit tool is write",
			tool:    "Edit",
			input:   "/tmp/test.txt",
			isWrite: true,
		},
		{
			name:    "Read tool is not write",
			tool:    "Read",
			input:   "/tmp/test.txt",
			isWrite: false,
		},
		{
			name:    "Bash rm is write",
			tool:    "Bash",
			input:   "rm -rf /tmp/test",
			isWrite: true,
		},
		{
			name:    "Bash mv is write",
			tool:    "Bash",
			input:   "mv old.txt new.txt",
			isWrite: true,
		},
		{
			name:    "Bash cp is write",
			tool:    "Bash",
			input:   "cp src.txt dst.txt",
			isWrite: true,
		},
		{
			name:    "Bash git commit is write",
			tool:    "Bash",
			input:   "git commit -m 'msg'",
			isWrite: true,
		},
		{
			name:    "Bash git push is write",
			tool:    "Bash",
			input:   "git push origin main",
			isWrite: true,
		},
		{
			name:    "Bash go test is not write",
			tool:    "Bash",
			input:   "go test ./...",
			isWrite: false,
		},
		{
			name:    "Bash cat is not write",
			tool:    "Bash",
			input:   "cat file.txt",
			isWrite: false,
		},
		{
			name:    "Bash grep is not write",
			tool:    "Bash",
			input:   "grep pattern file.txt",
			isWrite: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWriteOperation(tt.tool, tt.input); got != tt.isWrite {
				t.Errorf("IsWriteOperation() = %v, want %v", got, tt.isWrite)
			}
		})
	}
}

func TestDecision_String(t *testing.T) {
	tests := []struct {
		decision Decision
		want     string
	}{
		{DecisionAllow, "allow"},
		{DecisionAsk, "ask"},
		{DecisionDeny, "deny"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.decision.String(); got != tt.want {
				t.Errorf("Decision.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionRule_String(t *testing.T) {
	rule := PermissionRule{
		Pattern: "Bash(go test:*)",
		Action:  DecisionAllow,
		Source:  config.SrcProject,
		Origin:  ".claude/settings.json",
	}

	got := rule.String()
	expected := "project|allow|.claude/settings.json|Bash(go test:*)"
	if got != expected {
		t.Errorf("PermissionRule.String() = %v, want %v", got, expected)
	}
}
