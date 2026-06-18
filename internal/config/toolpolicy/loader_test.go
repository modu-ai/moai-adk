package toolpolicy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoad_Validates6FieldSchema covers REQ-TPS-002 + acceptance.md EC-5
// (malformed YAML must return a clear error, not silently produce wrong policy).
func TestLoad_Validates6FieldSchema(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "missing risk_tier",
			yaml: `metadata: {version: "1.0.0"}
entries:
  - tool: Bash
    args_pattern: "git push:*"
    decision: allow
    owner_agent: orchestrator
    audit: test
`,
			wantErr: "risk_tier is required",
		},
		{
			name: "invalid risk_tier value",
			yaml: `metadata: {version: "1.0.0"}
entries:
  - tool: Bash
    args_pattern: "*"
    risk_tier: critical
    decision: allow
    owner_agent: orchestrator
    audit: test
`,
			wantErr: "not in {read,write,irreversible}",
		},
		{
			name: "invalid decision value",
			yaml: `metadata: {version: "1.0.0"}
entries:
  - tool: Bash
    args_pattern: "*"
    risk_tier: read
    decision: block
    owner_agent: orchestrator
    audit: test
`,
			wantErr: "not in {allow,deny,ask}",
		},
		{
			name: "missing tool",
			yaml: `metadata: {version: "1.0.0"}
entries:
  - args_pattern: "*"
    risk_tier: read
    decision: allow
    owner_agent: orchestrator
    audit: test
`,
			wantErr: "tool is required",
		},
		{
			name: "missing audit",
			yaml: `metadata: {version: "1.0.0"}
entries:
  - tool: Bash
    args_pattern: "*"
    risk_tier: read
    decision: allow
    owner_agent: orchestrator
`,
			wantErr: "audit is required",
		},
		{
			name: "missing owner_agent",
			yaml: `metadata: {version: "1.0.0"}
entries:
  - tool: Bash
    args_pattern: "*"
    risk_tier: read
    decision: allow
    audit: test
`,
			wantErr: "owner_agent is required",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(dir, tt.name+".yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			_, err := Load(path)
			if err == nil {
				t.Fatalf("Load returned nil error; want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load error = %q; want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// TestLoad_ValidMinimalEntry verifies that an entry with empty args_pattern
// (tool-level entry like "Read" with no args restriction) is accepted —
// args_pattern is intentionally NOT required to be non-empty.
func TestLoad_ValidMinimalEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")
	yaml := `metadata: {version: "1.0.0"}
entries:
  - tool: Read
    args_pattern: ""
    risk_tier: read
    decision: allow
    owner_agent: orchestrator
    audit: tool-level read
`
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	doc, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error for valid minimal entry: %v", err)
	}
	if len(doc.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(doc.Entries))
	}
	if got := doc.Entries[0].SettingsSpecifier(); got != "Read" {
		t.Errorf("SettingsSpecifier = %q; want %q", got, "Read")
	}
}

// TestSettingsSpecifier covers the entry→settings.json specifier rendering
// (the inverse direction of parsing a settings.json entry back into tool+args).
func TestSettingsSpecifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tool, args, want string
	}{
		{"Read", "", "Read"},
		{"Read", "*", "Read"},
		{"Bash", "git push:*", "Bash(git push:*)"},
		{"Bash", "rm -rf /:*", "Bash(rm -rf /:*)"},
		{"mcp__context7__resolve-library-id", "", "mcp__context7__resolve-library-id"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			e := PolicyEntry{Tool: tt.tool, ArgsPattern: tt.args}
			if got := e.SettingsSpecifier(); got != tt.want {
				t.Errorf("SettingsSpecifier(%q,%q) = %q; want %q", tt.tool, tt.args, got, tt.want)
			}
			// round-trip back via splitSpecifier
			tool, args := splitSpecifier(tt.want)
			if tool != tt.tool {
				t.Errorf("splitSpecifier tool = %q; want %q", tool, tt.tool)
			}
			// args may normalize "" ↔ "*"; both render identically
			argsNormalizeEmpty := func(s string) string {
				if s == "*" {
					return ""
				}
				return s
			}
			if argsNormalizeEmpty(args) != argsNormalizeEmpty(tt.args) {
				t.Errorf("splitSpecifier args = %q; want %q", args, tt.args)
			}
		})
	}
}

// TestFilters covers FilterByRiskTier / FilterByDecision / FilterByTool.
func TestFilters(t *testing.T) {
	t.Parallel()
	doc := &PolicyDocument{
		Entries: []PolicyEntry{
			{Tool: "Bash", ArgsPattern: "rm:*", RiskTier: RiskTierWrite, Decision: DecisionAsk},
			{Tool: "Bash", ArgsPattern: "git push:*", RiskTier: RiskTierWrite, Decision: DecisionAllow},
			{Tool: "Bash", ArgsPattern: "git push --force:*", RiskTier: RiskTierIrreversible, Decision: DecisionDeny},
			{Tool: "Read", ArgsPattern: "", RiskTier: RiskTierRead, Decision: DecisionAllow},
		},
	}
	if got := len(doc.FilterByRiskTier(RiskTierIrreversible)); got != 1 {
		t.Errorf("FilterByRiskTier(irreversible) = %d entries; want 1", got)
	}
	if got := len(doc.FilterByRiskTier(RiskTierWrite)); got != 2 {
		t.Errorf("FilterByRiskTier(write) = %d entries; want 2", got)
	}
	if got := len(doc.FilterByDecision(DecisionDeny)); got != 1 {
		t.Errorf("FilterByDecision(deny) = %d entries; want 1", got)
	}
	if got := len(doc.FilterByDecision(DecisionAllow)); got != 2 {
		t.Errorf("FilterByDecision(allow) = %d entries; want 2", got)
	}
	if got := len(doc.FilterByTool("Read")); got != 1 {
		t.Errorf("FilterByTool(Read) = %d entries; want 1", got)
	}
	if got := len(doc.FilterByTool("Bash")); got != 3 {
		t.Errorf("FilterByTool(Bash) = %d entries; want 3", got)
	}
	// empty tool filter returns all entries
	if got := len(doc.FilterByTool("")); got != len(doc.Entries) {
		t.Errorf("FilterByTool('') = %d entries; want %d (all)", got, len(doc.Entries))
	}
}
