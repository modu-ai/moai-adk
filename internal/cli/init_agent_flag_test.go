package cli

// The `moai init --llm` flag after the general codex execution surface removal
// (card t1909): the harness deployment is claude-only. The flag stays for
// script compatibility, its usage documents claude as the only value, and
// validateInitFlags rejects every other value fail-loud. An init run — flag or
// not — leaves no Codex wiring files behind.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitAgentFlag_RegisteredClaudeOnly verifies the flag exists on the
// production initCmd and its usage names claude as the only deployable value.
func TestInitAgentFlag_RegisteredClaudeOnly(t *testing.T) {
	flag := initCmd.Flags().Lookup("llm")
	if flag == nil {
		t.Fatal("--llm flag not registered on initCmd")
	}
	if !strings.Contains(flag.Usage, "claude") {
		t.Errorf("--llm usage %q does not document the claude value", flag.Usage)
	}
	for _, stale := range []string{"gpt, or both", "codex|both"} {
		if strings.Contains(flag.Usage, stale) {
			t.Errorf("--llm usage %q still advertises the removed selection %q", flag.Usage, stale)
		}
	}
}

// TestValidateInitFlags_AgentClosedSet verifies the fail-loud claude-only
// closed set: empty and claude pass; the removed gpt/both/codex values and any
// unknown value exit with a diagnostic naming claude.
func TestValidateInitFlags_AgentClosedSet(t *testing.T) {
	for _, val := range []string{"", "claude"} {
		cmd := newInitTestCmd()
		if val != "" {
			if err := cmd.Flags().Set("llm", val); err != nil {
				t.Fatalf("set --llm=%s: %v", val, err)
			}
		}
		if err := validateInitFlags(cmd, nil); err != nil {
			t.Errorf("--llm=%q must validate, got error: %v", val, err)
		}
	}

	for _, val := range []string{"gemini", "gpt", "both", "codex"} {
		cmd := newInitTestCmd()
		if err := cmd.Flags().Set("llm", val); err != nil {
			t.Fatal(err)
		}
		err := validateInitFlags(cmd, nil)
		if err == nil {
			t.Errorf("--llm %s must fail validation (claude-only closed set, card t1909)", val)
			continue
		}
		if !strings.Contains(err.Error(), "claude") {
			t.Errorf("validation error %q does not name the claude-only set", err.Error())
		}
	}
}

// TestRunInit_AgentAbsentLeavesNoCodexFiles verifies the flag-absent
// contract: no .codex WIRING is created. The .codex/ directory itself may
// exist — the template tree ships .codex/agents/** (M5 agentemit artifacts) —
// but the wiring files (hooks.json, config.toml) and the trust sidecar must
// be absent.
func TestRunInit_AgentAbsentLeavesNoCodexFiles(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, nil)

	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml", ".moai/state/codex-wiring.json"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			t.Errorf("%s created by a flag-absent init", rel)
		}
	}
}

// TestRunInit_AgentClaudeLeavesNoCodexFiles verifies --llm claude equals
// flag-absent behavior: neither the Codex wiring files nor the trust sidecar
// is created (sidecar assertion per SPEC-CODEX-SIDECAR-GUARD-001).
func TestRunInit_AgentClaudeLeavesNoCodexFiles(t *testing.T) {
	projectDir, _ := runInitForAutonomy(t, nil, map[string]string{"llm": "claude"})

	for _, rel := range []string{".codex/hooks.json", ".codex/config.toml", ".moai/state/codex-wiring.json"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err == nil {
			t.Errorf("%s created by --llm claude init (must equal flag-absent)", rel)
		}
	}
}

// TestRunInit_NeverCallsCodexWiring is the reachability guard (source
// inspection, the inverted TestRunInit_CallsCodexWiring): runInit must NOT
// call the codex wiring helper — the gpt|both selection that drove it is gone.
func TestRunInit_NeverCallsCodexWiring(t *testing.T) {
	src, err := os.ReadFile("init.go")
	if err != nil {
		t.Fatalf("read init.go: %v", err)
	}
	if strings.Contains(string(src), "wireCodexUnlessClaude(") {
		t.Error("init.go still calls wireCodexUnlessClaude — the general codex wiring was removed (card t1909)")
	}
}
