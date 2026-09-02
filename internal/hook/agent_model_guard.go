// agent_model_guard.go — PreToolUse observation of per-agent model injection.
//
// The model-profile resolver (template.ResolveAgentModelEffort) computes a
// {model, effort} cell for every retained agent, but nothing has ever checked
// whether that computed model actually reaches a spawn. This file adds the
// missing inspector: on a PreToolUse event for the Agent (formerly Task) tool it
// extracts the agent identifier and the declared model from the spawn payload,
// asks the resolver what the model SHOULD be, and records a structured verdict.
//
// Layering (each layer is independently reversible):
//
//	observe  — always on; appends one JSONL record per spawn, never blocks.
//	advise   — always on; emits a non-blocking advisory on missing/mismatch.
//	block    — opt-in (workflow.agent_model_guard.enabled, default false);
//	           denies ONLY the mismatch verdict.
//
// The dominant real-world verdict is `missing`, not `mismatch`: a payload survey
// found the model argument present on well under 1% of observed spawns. Blocking
// `missing` would therefore refuse nearly every spawn, so `missing` stays
// advisory even when the gate is on.
//
// Fail-open is the house norm (branch_guard.go): a deny fires ONLY on positive
// evidence — the agent identifier parsed, the resolution mapped, a declared
// model present, and the two differing. Every other state (unparseable payload,
// absent subagent_type, unmapped agent, unreadable config, unresolved project
// root) allows. An enforcement bug must never wedge a session.
//
// effort is deliberately out of scope: the Agent tool exposes no effort
// parameter, so only `model` is observable at spawn time.
package hook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// SentinelAgentModelViolation prefixes every deny emitted by the agent-model
// guard, so the orchestrator can pattern-match the source without parsing the
// full reason string. Mirrors the BRANCH_GUARD_VIOLATION precedent.
const SentinelAgentModelViolation = "AGENT_MODEL_VIOLATION"

// agentModelAuditFileName is the per-spawn audit log under <root>/.moai/logs/.
// JSON Lines (not .log) because the point is aggregation — per-agent drift
// rates — matching the task-metrics.jsonl precedent rather than the
// branch-guard-audit.log one.
const agentModelAuditFileName = "agent-model-audit.jsonl"

// agentModelVerdict is the 4-valued classification of a single spawn.
type agentModelVerdict string

const (
	// verdictAgentModelOK — the declared model equals the resolved model.
	verdictAgentModelOK agentModelVerdict = "ok"
	// verdictAgentModelMissing — the resolver mapped a concrete alias but the
	// spawn declared no model at all. The dominant observed case.
	verdictAgentModelMissing agentModelVerdict = "missing"
	// verdictAgentModelMismatch — the spawn declared a model that differs from
	// the resolved one. The only blockable verdict.
	verdictAgentModelMismatch agentModelVerdict = "mismatch"
	// verdictAgentModelUnmapped — the agent is outside the retained catalog
	// (a user-authored harness specialist); the resolver returns the inherit
	// sentinel and there is nothing to compare against.
	verdictAgentModelUnmapped agentModelVerdict = "unmapped"
)

// agentSpawn carries the fields the guard reads out of an Agent spawn
// payload. The prompt body is deliberately NOT captured — it never reaches the
// audit log or a fixture.
type agentSpawn struct {
	// Agent is the subagent_type argument (the agent identifier).
	Agent string
	// DeclaredModel is the model argument, empty when the spawn omitted it.
	DeclaredModel string
	// Name is the named-teammate spawn argument, empty when the spawn carried
	// none. Read by the agent-stop guard's respawn-clear path: a fresh spawn
	// carrying a stopped name clears the stop record before the spawn
	// proceeds.
	Name string
}

// extractAgentSpawn parses an Agent/Task tool_input payload. It returns ok=false
// on any uncertainty — unparseable JSON, or an absent/empty subagent_type — so
// the caller can fall through to allow without a special-case branch.
func extractAgentSpawn(toolInput json.RawMessage) (agentSpawn, bool) {
	if len(toolInput) == 0 {
		return agentSpawn{}, false
	}
	var parsed struct {
		SubagentType string `json:"subagent_type"`
		Model        string `json:"model"`
		Name         string `json:"name"`
	}
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return agentSpawn{}, false
	}
	if parsed.SubagentType == "" {
		return agentSpawn{}, false
	}
	return agentSpawn{Agent: parsed.SubagentType, DeclaredModel: parsed.Model, Name: parsed.Name}, true
}

// resolveAgentModel asks the profile-matrix resolver what model the given agent
// should run under. It is the ONLY path to that answer — the guard never
// restates a matrix cell or a model alias (AP-2).
func resolveAgentModel(llm config.LLMConfig, agent string) (mapped bool, model string) {
	me, ok := template.ResolveAgentModelEffort(llm, agent)
	if !ok {
		return false, ""
	}
	return true, me.Model
}

// classifyAgentModel returns the verdict for a spawn plus the resolved model
// alias (empty when the agent is unmapped).
func classifyAgentModel(sp agentSpawn, llm config.LLMConfig) (agentModelVerdict, string) {
	mapped, resolved := resolveAgentModel(llm, sp.Agent)
	if !mapped || resolved == "" {
		return verdictAgentModelUnmapped, ""
	}
	if sp.DeclaredModel == "" {
		return verdictAgentModelMissing, resolved
	}
	// Alias comparison is case-insensitive: a declaration differing only in case
	// names the same model and must not be reported as drift.
	if !strings.EqualFold(sp.DeclaredModel, resolved) {
		return verdictAgentModelMismatch, resolved
	}
	return verdictAgentModelOK, resolved
}

// agentModelAdvisory renders the non-blocking advisory for a verdict. It returns
// "" for ok and unmapped — there is nothing to correct in either case.
func agentModelAdvisory(agent, resolved string, v agentModelVerdict) string {
	switch v {
	case verdictAgentModelMissing:
		return fmt.Sprintf(
			"agent-model: %s spawned without a model argument; the active profile resolves it to %s. "+
				"Pass model=%s on the spawn so the profile is actually applied.",
			agent, resolved, resolved)
	case verdictAgentModelMismatch:
		return fmt.Sprintf(
			"agent-model: %s spawned with a model that differs from the active profile, which resolves it to %s. "+
				"Pass model=%s on the spawn, or adjust the profile.",
			agent, resolved, resolved)
	default:
		return ""
	}
}

// agentModelAuditRecord is one JSONL row. It carries no prompt body and no
// spawn payload — only the identifiers needed to aggregate drift by agent.
type agentModelAuditRecord struct {
	Timestamp     string `json:"timestamp"`
	SessionID     string `json:"session_id"`
	Agent         string `json:"agent"`
	DeclaredModel string `json:"declared_model"`
	ResolvedModel string `json:"resolved_model"`
	Verdict       string `json:"verdict"`
}

// appendAgentModelAudit appends one record to <projectRoot>/.moai/logs/.
// Every failure path is silent-and-continue: an unresolved project root, an
// unwritable directory, or a marshal error must never surface to the caller,
// because an observation failure may not block a spawn.
func appendAgentModelAudit(projectRoot string, rec agentModelAuditRecord) {
	if projectRoot == "" {
		return
	}
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	line, err := json.Marshal(rec)
	if err != nil {
		slog.Warn("agent_model_guard: failed to marshal audit record", "error", err)
		return
	}

	logsDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o750); err != nil {
		slog.Warn("agent_model_guard: failed to create logs dir", "dir", logsDir, "error", err)
		return
	}
	path := filepath.Join(logsDir, agentModelAuditFileName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		slog.Warn("agent_model_guard: failed to open audit log", "path", path, "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(line, '\n')); err != nil {
		slog.Warn("agent_model_guard: failed to append audit record", "path", path, "error", err)
	}
}

// agentModelGuardEnabled reports whether the opt-in blocking gate is on
// (workflow.agent_model_guard.enabled, distributed default false). A nil
// provider or nil config returns false so a misconfigured hook can never
// accidentally reach the deny path — the same defensive shape as
// branchGuardEnabled.
func (h *preToolHandler) agentModelGuardEnabled() bool {
	if h.cfg == nil {
		return false
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return false
	}
	return cfg.Workflow.AgentModelGuard.Enabled
}

// llmConfig returns the LLM config the resolver needs, falling back to the zero
// value when no config is reachable. The zero value resolves through the
// matrix defaults, so observation still works without a loaded config.
func (h *preToolHandler) llmConfig() config.LLMConfig {
	if h.cfg == nil {
		return config.LLMConfig{}
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return config.LLMConfig{}
	}
	return cfg.LLM
}

// checkAgentModel is the PreToolUse entry point for an Agent/Task spawn. It
// returns (decision, reason, advisory): decision is DecisionDeny only when the
// gate is on AND the verdict is mismatch; otherwise it is empty and the caller
// falls through to allow, optionally surfacing the advisory.
func (h *preToolHandler) checkAgentModel(input *HookInput) (decision, reason, advisory string) {
	sp, ok := extractAgentSpawn(input.ToolInput)
	if !ok {
		// Uncertain payload — nothing to observe, nothing to enforce.
		return "", "", ""
	}

	verdict, resolved := classifyAgentModel(sp, h.llmConfig())

	appendAgentModelAudit(h.projectRoot(), agentModelAuditRecord{
		SessionID:     input.SessionID,
		Agent:         sp.Agent,
		DeclaredModel: sp.DeclaredModel,
		ResolvedModel: resolved,
		Verdict:       string(verdict),
	})

	advisory = agentModelAdvisory(sp.Agent, resolved, verdict)

	// Blocking is opt-in AND mismatch-only. `missing` stays advisory even when
	// the gate is on: blocking it would refuse nearly every spawn.
	if verdict == verdictAgentModelMismatch && h.agentModelGuardEnabled() {
		reason = fmt.Sprintf(
			"%s: %s was spawned with model %q but the active profile resolves it to %q. "+
				"Pass model=%s on the spawn, or adjust the profile.",
			SentinelAgentModelViolation, sp.Agent, sp.DeclaredModel, resolved, resolved)
		slog.Warn("agent model guard denied",
			"agent", sp.Agent,
			"declared_model", sp.DeclaredModel,
			"resolved_model", resolved,
			"session_id", input.SessionID,
		)
		return DecisionDeny, reason, advisory
	}

	return "", "", advisory
}
