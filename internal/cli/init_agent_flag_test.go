package cli

// SPEC-CODEX-WIRING-001 M2 — the `moai init --agent claude|codex|both` flag
// (REQ-CW-001) and its init-tail wiring (D3 semantics):
//
//   - claude / absent → today's behavior byte-for-byte (no .codex files)
//   - codex           → .mcp.json provisioning treated as declined + Codex wiring
//   - both            → .mcp.json provisioning + Codex wiring
//
// The flag beats the wizard answer (plan D3); validation is fail-loud with the
// closed set named (autonomy-tier pattern).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitAgentFlag_RegisteredWithClosedSet verifies the flag exists on the
// production initCmd and its usage names all three values (AC-CW-001's
// `moai init --help | grep -- '--agent'` contract).
func TestInitAgentFlag_RegisteredWithClosedSet(t *testing.T) {
	flag := initCmd.Flags().Lookup("agent")
	if flag == nil {
		t.Fatal("--agent flag not registered on initCmd")
	}
	for _, want := range []string{"claude", "codex", "both"} {
		if !strings.Contains(flag.Usage, want) {
			t.Errorf("--agent usage %q does not document %q (AC-CW-001)", flag.Usage, want)
		}
	}
}

// TestValidateInitFlags_AgentClosedSet verifies the fail-loud closed set
// (AC-CW-001 second clause): valid values and empty pass; an invalid value
// exits with a diagnostic naming the valid values.
func TestValidateInitFlags_AgentClosedSet(t *testing.T) {
	validCases := []string{"", "claude", "codex", "both"}
	for _, val := range validCases {
		cmd := newInitTestCmd()
		if val != "" {
			if err := cmd.Flags().Set("agent", val); err != nil {
				t.Fatalf("set --agent=%s: %v", val, err)
			}
		}
		if err := validateInitFlags(cmd, nil); err != nil {
			t.Errorf("--agent=%q must validate, got error: %v", val, err)
		}
	}

	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("agent", "gemini"); err != nil {
		t.Fatal(err)
	}
	err := validateInitFlags(cmd, nil)
	if err == nil {
		t.Fatal("--agent gemini must fail validation (fail-loud closed set)")
	}
	for _, want := range []string{"claude", "codex", "both"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("validation error %q does not list valid value %q", err.Error(), want)
		}
	}
}

// TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning verifies the codex
// semantics end to end through runInit: .codex wiring files exist, and the
// MCP provisioning call site treats codex as declined (D3).
func TestRunInit_AgentCodexWiresAndSkipsMCPProvisioning(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"agent": "codex"})

	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml", ".moai/state/codex-wiring.json"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Errorf("%s missing after --agent codex init: %v", rel, err)
		}
	}
}

// TestRunInit_AgentBothWiresBothSides verifies the both semantics: Codex
// wiring files exist (AC-CW-005).
func TestRunInit_AgentBothWiresBothSides(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"agent": "both"})

	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Errorf("%s missing after --agent both init: %v", rel, err)
		}
	}
}

// TestRunInit_AgentAbsentLeavesNoCodexFiles verifies the backward-compat
// contract (AC-CW-004): with no --agent flag, no .codex WIRING is created.
// The .codex/ directory itself may exist — the template tree ships
// .codex/agents/** (M5 agentemit artifacts) — but the wiring files
// (hooks.json, config.toml) and the trust sidecar must be absent.
func TestRunInit_AgentAbsentLeavesNoCodexFiles(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, nil)

	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml", ".moai/state/codex-wiring.json"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			t.Errorf("%s created by a flag-absent init (AC-CW-004 violation)", rel)
		}
	}
}

// TestRunInit_AgentClaudeLeavesNoCodexFiles verifies --agent claude equals
// flag-absent behavior (REQ-CW-001).
func TestRunInit_AgentClaudeLeavesNoCodexFiles(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"agent": "claude"})

	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			t.Errorf("%s created by --agent claude init (must equal flag-absent)", rel)
		}
	}
}

// TestRunInit_CodexProvisioningDeclineIsolated verifies the D3 isolation: the
// decision lives in one place — resolveAgentWiring(cmd) — so the codex decline
// of the .mcp.json provisioning call is a one-line reversible delta.
func TestRunInit_CodexProvisioningDeclineIsolated(t *testing.T) {
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("agent", "codex"); err != nil {
		t.Fatal(err)
	}
	if w := resolveAgentWiring(cmd); w != agentWiringCodex {
		t.Errorf("resolveAgentWiring(codex) = %v, want %v", w, agentWiringCodex)
	}
	if err := cmd.Flags().Set("agent", "both"); err != nil {
		t.Fatal(err)
	}
	if w := resolveAgentWiring(cmd); w != agentWiringBoth {
		t.Errorf("resolveAgentWiring(both) = %v, want %v", w, agentWiringBoth)
	}
	cmd2 := newInitTestCmd()
	if w := resolveAgentWiring(cmd2); w != agentWiringClaude {
		t.Errorf("resolveAgentWiring(absent) = %v, want %v (default claude)", w, agentWiringClaude)
	}
	if err := cmd2.Flags().Set("agent", "claude"); err != nil {
		t.Fatal(err)
	}
	if w := resolveAgentWiring(cmd2); w != agentWiringClaude {
		t.Errorf("resolveAgentWiring(claude) = %v, want %v", w, agentWiringClaude)
	}
}

// TestRunInit_CallsCodexWiring is the reachability guard (source inspection,
// TestRunInit_CallsMCPProvisioning pattern): runInit must call the codex
// wiring helper adjacent to provisionMCPEntryUnlessDeclined.
func TestRunInit_CallsCodexWiring(t *testing.T) {
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	body := string(src)
	if !strings.Contains(body, "wireCodexUnlessClaude(") {
		t.Error("runInit must call wireCodexUnlessClaude — without it the --agent codex|both selection is dropped")
	}
}
