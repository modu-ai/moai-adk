// Package cli — codex audit backend (SPEC-MOAI-MCP-SERVER-001 M2).
//
// mcp_codex.go implements the two Phase-1 codex tools registered alongside
// the M1 core surface: `codex_audit` (REQ-MCP-006 / AC-MCP-007) and
// `codex_setup` (REQ-MCP-007 / AC-MCP-008). It is a pure Go reimplementation
// — there is NO Node bridge of any kind (design.md §3 M2).
//
// The codex binary is OPTIONAL and experimental (R1): it may be absent from
// PATH or its `app-server` JSON-RPC surface may differ from the documented
// `review/start` / `turn/start` methods. Every code path therefore fails
// OPEN: a missing / erroring / malformed codex yields a structured
// VerdictInconclusive result, never a hard error and never a panic
// (REQ-MCP-012 preview; design.md §4.3 fail-open state machine). The full
// 3-way claude-fallback plumbing lands in M3; M2 only guarantees that the
// codex_audit tool + the review-gate hook never hard-crash on a missing codex.
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
package cli

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// resolveCodexModelEffort resolves the model + effort for one codex request:
// an explicit caller-supplied `model` is sent verbatim; otherwise the request
// carries neither field and codex applies its own configured default. MoAI
// assigns no per-agent model, so no llm.yaml cell is consulted
// (SPEC-AGENT-MODEL-INHERIT-001 design D5).
//
// This is the shared body the task path and the Stop-hook review gate resolve
// through. The audit pin is deliberately NOT read here
// (SPEC-V3R6-AUDIT-MODEL-PIN-001 REQ-AMP-008 / plan.md §G AP-1) — a config-file
// pin is persistent project state and must not leak into delegation tasks.
func resolveCodexModelEffort(params map[string]any) config.ModelEffort {
	if explicit, ok := params["model"].(string); ok && strings.TrimSpace(explicit) != "" {
		return config.ModelEffort{Model: strings.TrimSpace(explicit)}
	}
	return config.ModelEffort{}
}

// pinFreeEffortResolver adapts the legacy pin-free resolver to the
// error-returning audit seam: the task path carries no pin read, so it can
// never fail (SPEC-AUDIT-CEILING-002).
func pinFreeEffortResolver(params map[string]any) (config.ModelEffort, error) {
	return resolveCodexModelEffort(params), nil
}

// readCodexTaskAllowWrite reads workflow.codex.task.allow_write from
// `.moai/config/sections/workflow.yaml` (REQ-CX2-007). It is the sibling of
// readCodexReviewGateEnabled above and shares its truth table verbatim
// (fail-CLOSED — the opt-in default is OFF):
//
//   - file missing / unreadable                   → false
//   - YAML parse error                            → false
//   - `workflow.codex.task` block absent          → false (Go zero-value default)
//   - `workflow.codex.task.allow_write: false`    → false
//   - `workflow.codex.task.allow_write: true`     → true
//
// The key path is NESTED under the file's `workflow:` root, matching the shape
// config.Loader expects and the shape the sibling gate reads; the flat form is
// NOT accepted as an alias. Failing closed here is the difference between a
// misconfigured project running codex read-only and a misconfigured project
// handing an external MCP host write access to its working tree (spec.md §H R3).
//
// The distributed default is false and lives in internal/config/defaults.go
// (CodexTaskConfig.AllowWrite); the typed field is internal/config/types.go.
// No template file carries this key (§25 / REQ-CX2-015).
func readCodexTaskAllowWrite(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	configPath := filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	var doc struct {
		Workflow struct {
			Codex struct {
				Task struct {
					AllowWrite bool `yaml:"allow_write"`
				} `yaml:"task"`
			} `yaml:"codex"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false
	}
	return doc.Workflow.Codex.Task.AllowWrite
}
