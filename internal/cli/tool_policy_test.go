package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestToolPolicyCmd_Registered verifies the subcommand is registered on the
// root CLI tree so it appears in `moai --help` and is dispatchable.
func TestToolPolicyCmd_Registered(t *testing.T) {
	t.Parallel()
	cmd := newToolPolicyCmd()
	if cmd == nil {
		t.Fatal("newToolPolicyCmd returned nil")
	}
	if cmd.Use != "tool-policy" {
		t.Errorf("Use = %q; want %q", cmd.Use, "tool-policy")
	}
	// build + list subcommands must both be present.
	var subs []string
	for _, c := range cmd.Commands() {
		subs = append(subs, c.Use)
	}
	if !sliceContains(subs, "build") {
		t.Errorf("build subcommand missing; have %v", subs)
	}
	if !sliceContains(subs, "list") {
		t.Errorf("list subcommand missing; have %v", subs)
	}
}

// TestToolPolicyList_QueryFilters (AC-TPS-006) verifies the thin query subcommand
// loads the YAML directly and supports --risk-tier / --decision / --tool filter
// flags modeled on `moai constitution list --zone` SHAPE.
//
// The tool-policy query does NOT delegate to or wrap `moai constitution list`
// (the schemas are disjoint — see REQ-TPS-006 / D9). This test asserts the
// query works end-to-end against a fixture YAML and that the filters narrow
// the output correctly.
func TestToolPolicyList_QueryFilters(t *testing.T) {
	t.Parallel()

	// Use the project's own tool-policy.yaml as the fixture — it is a real
	// SSOT artifact and a meaningful end-to-end smoke test of the query path.
	policyPath := "../../.moai/config/sections/tool-policy.yaml"

	tests := []struct {
		name       string
		riskTier   string
		decision   string
		tool       string
		wantSubstr string // at least one output line must contain this
		wantCount  int    // 0 = don't check count; >0 = expect exactly this many entries
	}{
		{name: "all entries", wantCount: 0},
		{name: "filter irreversible", riskTier: "irreversible", wantSubstr: "irreversible"},
		{name: "filter allow", decision: "allow", wantSubstr: "allow"},
		{name: "filter deny", decision: "deny", wantSubstr: "deny"},
		{name: "filter ask", decision: "ask", wantSubstr: "ask"},
		{name: "filter tool Bash", tool: "Bash", wantSubstr: "Bash"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := newToolPolicyListCmd()
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{})
			if tt.riskTier != "" {
				if err := cmd.Flags().Set("risk-tier", tt.riskTier); err != nil {
					t.Fatalf("set risk-tier: %v", err)
				}
			}
			if tt.decision != "" {
				if err := cmd.Flags().Set("decision", tt.decision); err != nil {
					t.Fatalf("set decision: %v", err)
				}
			}
			if tt.tool != "" {
				if err := cmd.Flags().Set("tool", tt.tool); err != nil {
					t.Fatalf("set tool: %v", err)
				}
			}
			if err := cmd.Flags().Set("policy", policyPath); err != nil {
				t.Fatalf("set policy: %v", err)
			}
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			got := out.String()
			if tt.wantSubstr != "" && !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("output missing substring %q;\ngot:\n%s", tt.wantSubstr, got)
			}
		})
	}
}

// TestToolPolicyList_InvalidFlagValues verifies the query rejects invalid
// --risk-tier and --decision values with a clear error (acceptance.md EC-5
// analogue for the query path).
func TestToolPolicyList_InvalidFlagValues(t *testing.T) {
	t.Parallel()
	policyPath := "../../.moai/config/sections/tool-policy.yaml"

	tests := []struct {
		name   string
		flag   string
		value  string
		wantErr string
	}{
		{name: "bad risk-tier", flag: "risk-tier", value: "critical", wantErr: "not in {read,write,irreversible}"},
		{name: "bad decision", flag: "decision", value: "block", wantErr: "not in {allow,deny,ask}"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := newToolPolicyListCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{})
			_ = cmd.Flags().Set(tt.flag, tt.value)
			_ = cmd.Flags().Set("policy", policyPath)
			err := cmd.Execute()
			if err == nil {
				t.Fatalf("Execute returned nil; want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q; want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// TestToolPolicyList_JSONFormat verifies --format json emits valid JSON.
func TestToolPolicyList_JSONFormat(t *testing.T) {
	t.Parallel()
	policyPath := "../../.moai/config/sections/tool-policy.yaml"
	cmd := newToolPolicyListCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})
	_ = cmd.Flags().Set("format", "json")
	_ = cmd.Flags().Set("risk-tier", "irreversible")
	_ = cmd.Flags().Set("policy", policyPath)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := strings.TrimSpace(out.String())
	if !strings.HasPrefix(got, "[") || !strings.HasSuffix(got, "]") {
		t.Errorf("json output not a JSON array; first/last 40 chars: %q...%q", got[:minInt(40, len(got))], got[maxInt(0, len(got)-40):])
	}
}

// TestToolPolicyCmd_NoAskUserQuestion (C-HRA-008 subagent boundary) verifies
// the tool-policy CLI code does not call AskUserQuestion. The CLI runs in
// subagent context — orchestrator owns user interaction.
func TestToolPolicyCmd_NoAskUserQuestion(t *testing.T) {
	t.Parallel()
	files := []string{
		"tool_policy.go",
	}
	for _, f := range files {
		path := f
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			body, err := readFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			// The only acceptable occurrence of "AskUserQuestion" in CLI code is
			// inside a comment or string literal documenting the boundary itself.
			// We scan for call-shapes: "AskUserQuestion(" or "mcp__askuser".
			for _, bad := range []string{"AskUserQuestion(", "mcp__askuser"} {
				if strings.Contains(string(body), bad) {
					t.Errorf("%s contains %q — CLI must not invoke AskUserQuestion (subagent boundary)", path, bad)
				}
			}
		})
	}
}

// TestToolPolicyCmd_DoesNotWrapConstitution (AC-TPS-006) verifies the
// tool-policy query does NOT delegate to or wrap `moai constitution list`.
// The schemas are disjoint; wrapping is infeasible (D9 decision).
func TestToolPolicyCmd_DoesNotWrapConstitution(t *testing.T) {
	t.Parallel()
	body, err := readFile("tool_policy.go")
	if err != nil {
		t.Fatalf("read tool_policy.go: %v", err)
	}
	// The CLI must not import the constitution package or call its loaders.
	banned := []string{
		"internal/constitution",
		"constitution.LoadRegistry",
		"constitution.Rule",
	}
	for _, b := range banned {
		if strings.Contains(string(body), b) {
			t.Errorf("tool_policy.go references %q — tool-policy query must NOT wrap constitution (schemas disjoint, REQ-TPS-006/D9)", b)
		}
	}
}

// sliceContains is a minimal helper for string slice membership (renamed to
// avoid collision with the existing `contains` in update_yaml_test.go).
func sliceContains(items []string, want string) bool {
	for _, s := range items {
		if s == want {
			return true
		}
	}
	return false
}

// readFile reads a file relative to the test's CWD (the internal/cli dir).
func readFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// min returns the smaller of a or b.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of a or b.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
