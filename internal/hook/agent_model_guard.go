// agent_model_guard.go — PreToolUse observation of per-agent model injection.
//
// Subagents inherit the main session's model and effort, so the expectation for
// a spawn is DECLARATION-ONLY: what the spawn declares is what it runs. This
// file is the inspector: on a PreToolUse event for the Agent (formerly Task)
// tool it extracts the agent identifier and the declared model from the spawn
// payload, and records a structured verdict — the declared model verbatim, or
// the `inherit` sentinel when the spawn declares none.
//
// Layering (each layer is independently reversible):
//
//	observe  — always on; appends one JSONL record per spawn, never blocks.
//	advise   — always on; emits a non-blocking advisory on missing/mismatch.
//
// The former opt-in block layer (workflow.agent_model_guard.enabled) is gone:
// subagents inherit the main session's model and effort
// (SPEC-AGENT-MODEL-INHERIT-001), so no spawn is denied on the basis of its
// model. A leftover agent_model_guard key in a user's workflow.yaml is ignored.
// The observation layer stays until its reader-side rework lands.
//
// Fail-open is the house norm (branch_guard.go): every uncertainty
// (unparseable payload, absent subagent_type, unmapped agent, unreadable
// config, unresolved project root) records nothing it cannot and never blocks.
//
// effort is deliberately out of scope: the Agent tool exposes no effort
// parameter, so only `model` is observable at spawn time.
//
// v0.3.0 extension (REQ-AFR-018, SPEC-WEB-AGENTFM-RESTORE-001, card t1421):
// with the llm.agent_overrides_consume opt-in ON, the advise layer compares
// the spawn's declared model against the agent's override expectation —
// template.ResolveAgentOverrideConsumption — and the audit record gains an
// override hit/miss field. Every property above is preserved: the extension
// adds observation and advice only, never a deny and never a payload
// rewrite, and a closed gate (or an unreadable config — fail-open) degrades
// to the off mark with the pre-extension behaviour intact.
package hook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
)

// agentModelAuditFileName is the per-spawn audit log under <root>/.moai/logs/.
// JSON Lines (not .log) because the point is aggregation — per-agent drift
// rates — matching the task-metrics.jsonl precedent rather than the
// branch-guard-audit.log one.
const agentModelAuditFileName = "agent-model-audit.jsonl"

// agentModelVerdict is the 2-valued classification of a single spawn under the
// session-inherit design: what the spawn declared, nothing more.
type agentModelVerdict string

const (
	// verdictAgentModelDeclared — the spawn carried an explicit model argument,
	// recorded verbatim.
	verdictAgentModelDeclared agentModelVerdict = "declared"
	// verdictAgentModelInherit — the spawn declared no model; it inherits the
	// main session's model. The expected state.
	verdictAgentModelInherit agentModelVerdict = "inherit"
)

// agentSpawn carries the fields the guard reads out of an Agent spawn
// payload. The prompt body is captured ONLY where a consumer needs it in
// memory (the served-model gate matches the spawn's SPEC against outstanding
// refusals, card t1323) — it still never reaches the audit log or a fixture.
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
	// Prompt is the spawn prompt body, empty when the spawn carried none.
	// In-memory only: the served-model gate reads it to attribute the spawn to
	// its SPEC; it is never written to any log or fixture.
	Prompt string
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
		Prompt       string `json:"prompt"`
	}
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return agentSpawn{}, false
	}
	if parsed.SubagentType == "" {
		return agentSpawn{}, false
	}
	return agentSpawn{Agent: parsed.SubagentType, DeclaredModel: parsed.Model, Name: parsed.Name, Prompt: parsed.Prompt}, true
}

// classifyAgentModel returns the verdict for a spawn plus the model value to
// record. With the profile matrix gone (SPEC-AGENT-MODEL-INHERIT-001 M5) the
// observation layer logs what the spawn DECLARED: an explicit model argument
// is recorded verbatim, and no declaration records `inherit` — subagents
// inherit the main session's model by design, so an absent argument is the
// expected state, not a gap. There is no expected model to compare against
// and no advisory: the record is the product.
func classifyAgentModel(sp agentSpawn) (agentModelVerdict, string) {
	if sp.DeclaredModel != "" {
		return verdictAgentModelDeclared, sp.DeclaredModel
	}
	return verdictAgentModelInherit, "inherit"
}

// agentModelAdvisory is retired with the profile matrix (M5): both of its
// messages told the caller to pass a model the resolver expected — advice
// that inverts under session inheritance, where no declaration is the desired
// state. The observation log is the product; there is nothing to correct.

// agentModelAuditRecord is one JSONL row. It carries no prompt body and no
// spawn payload — only the identifiers needed to aggregate drift by agent.
type agentModelAuditRecord struct {
	Timestamp     string `json:"timestamp"`
	SessionID     string `json:"session_id"`
	Agent         string `json:"agent"`
	DeclaredModel string `json:"declared_model"`
	ResolvedModel string `json:"resolved_model"`
	Verdict       string `json:"verdict"`
	// OverrideConsumption is the v0.3.0 extension (REQ-AFR-018): how the
	// spawn's declared model related to the agent's llm.agent_overrides
	// expectation under the opt-in gate — hit (pinned and matched), miss
	// (pinned but the declaration differs from or omits the pin), inherit
	// (gate on, expectation is plain inheritance or the explicit inherit
	// no-op), or off (gate closed / config unreadable — storage-only). The
	// axis is MODEL only: effort is never compared (the file header contract).
	OverrideConsumption string `json:"override_consumption"`
}

// Override-consumption marks (the values of
// agentModelAuditRecord.OverrideConsumption).
const (
	overrideConsumptionOff     = "off"
	overrideConsumptionHit     = "hit"
	overrideConsumptionMiss    = "miss"
	overrideConsumptionInherit = "inherit"
)

// classifyOverrideConsumption compares a spawn's declared model against the
// agent's resolved override expectation (model axis only, case-insensitive —
// the same alias-comparison tolerance the pre-M5 guard used). It returns the
// record mark plus the non-blocking miss advisory ("" for every other mark —
// there is nothing to correct in a hit, a closed gate, or an explicit
// inheritance).
func classifyOverrideConsumption(sp agentSpawn, c template.AgentOverrideConsumption) (mark, advisory string) {
	if !c.ConsumeEnabled {
		return overrideConsumptionOff, ""
	}
	if !c.Pinned {
		return overrideConsumptionInherit, ""
	}
	if sp.DeclaredModel != "" && strings.EqualFold(sp.DeclaredModel, c.Model) {
		return overrideConsumptionHit, ""
	}
	declared := sp.DeclaredModel
	if declared == "" {
		declared = "no model"
	}
	advisory = fmt.Sprintf(
		"agent-model: %s carries an llm.agent_overrides pin (model: %s) but this spawn declared %s — the pin was not applied; pass model=%s on the spawn (llm.agent_overrides_consume is on).",
		sp.Agent, c.Model, declared, c.Model)
	return overrideConsumptionMiss, advisory
}

// appendAgentModelAudit appends one record to <projectRoot>/.moai/logs/.
// Every failure path is silent-and-continue: an unresolved project root, an
// unwritable directory, or a marshal error must never surface to the caller,
// because an observation failure may not block a spawn.
func appendAgentModelAudit(projectRoot string, rec agentModelAuditRecord) {
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	appendAuditJSONL(projectRoot, rec)
}

// appendAuditJSONL appends one JSON line to the agent-model audit log. It is
// the single writer shared by the PreToolUse row and the SubagentStop
// served-model row, and it is append-only: no existing row is ever rewritten.
// Every failure is silent-and-continue, for the reason given above.
func appendAuditJSONL(projectRoot string, rec any) {
	if projectRoot == "" {
		return
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

// checkAgentModel is the PreToolUse entry point for an Agent/Task spawn. It
// records the declared-model observation and — with the v0.3.0 opt-in ON —
// compares the declaration against the agent's override expectation (model
// axis), returning the non-blocking miss advisory when the pin was not
// applied. It never denies.
func (h *preToolHandler) checkAgentModel(input *HookInput) (advisory string) {
	sp, ok := extractAgentSpawn(input.ToolInput)
	if !ok {
		// Uncertain payload — nothing to observe.
		return ""
	}

	verdict, recorded := classifyAgentModel(sp)

	// Fail-open on config: an absent provider or an unreadable config
	// resolves the closed gate (storage-only), degrading the extension to the
	// off mark with the pre-extension behaviour intact.
	var consumption template.AgentOverrideConsumption
	if h.cfg != nil {
		if cfg := h.cfg.Get(); cfg != nil {
			consumption = template.ResolveAgentOverrideConsumption(cfg.LLM, sp.Agent)
		}
	}
	overrideMark, overrideAdvisory := classifyOverrideConsumption(sp, consumption)

	appendAgentModelAudit(h.projectRoot(), agentModelAuditRecord{
		SessionID:           input.SessionID,
		Agent:               sp.Agent,
		DeclaredModel:       sp.DeclaredModel,
		ResolvedModel:       recorded,
		Verdict:             string(verdict),
		OverrideConsumption: overrideMark,
	})

	return overrideAdvisory
}
