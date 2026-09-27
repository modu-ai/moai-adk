// served_model_stop.go — the SubagentStop half of served-model observation.
//
// When a subagent stops, its transcript holds every response it produced; the
// `.message.model` values of the assistant rows name the model that actually
// answered. This file records that served set next to the declared and
// resolved models, in the same audit log the PreToolUse guard writes, so one
// spawn's declaration and its served reality can be joined by agent and
// session.
//
// Fail-open is absolute here: the observation never returns a block decision,
// never surfaces an error, and never rewrites an existing audit row.
package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// servedObservationSource marks a SubagentStop row. PreToolUse rows carry no
// source field, which is how the two are told apart.
const servedObservationSource = "subagent_stop"

// selfReportPrefix opens the self-reported model line an auditor writes as
// the first line of its final message.
const selfReportPrefix = "auditor-model:"

// servedAuditRecord is the SubagentStop row of the agent-model audit log. Its
// first six keys share their names with agentModelAuditRecord so a reader can
// aggregate both row kinds by agent and session.
type servedAuditRecord struct {
	Timestamp         string   `json:"timestamp"`
	SessionID         string   `json:"session_id"`
	Agent             string   `json:"agent"`
	DeclaredModel     string   `json:"declared_model"`
	ResolvedModel     string   `json:"resolved_model"`
	Verdict           string   `json:"verdict"`
	Source            string   `json:"source"`
	AgentID           string   `json:"agent_id"`
	ServedModels      []string `json:"served_models"`
	SelfReportedModel string   `json:"self_reported_model"`
	ExpectedModel     string   `json:"expected_model,omitempty"`
	Cause             string   `json:"cause,omitempty"`
}

// loadedConfig returns the loaded configuration, or nil when none is reachable.
func (h *subagentStopHandler) loadedConfig() *config.Config {
	if h == nil || h.cfg == nil {
		return nil
	}
	return h.cfg.Get()
}

// subagentTranscriptPath locates the stopping subagent's own transcript: the
// hook input's agent_transcript_path, or else
// <transcript_path without .jsonl>/subagents/agent-<agent_id>.jsonl. An
// agent_id that could step outside that directory derives nothing.
func subagentTranscriptPath(input *HookInput) string {
	if p := strings.TrimSpace(input.AgentTranscriptPath); p != "" {
		return p
	}
	id := strings.TrimSpace(input.AgentID)
	parent := strings.TrimSpace(input.TranscriptPath)
	if id == "" || parent == "" || strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return ""
	}
	return filepath.Join(strings.TrimSuffix(parent, ".jsonl"), "subagents", "agent-"+id+".jsonl")
}

// selfReportedModel reads `auditor-model: <model>` from the first non-blank
// line of the final message. Anywhere else — a later line, the report file —
// is not a self-report. The value is recorded only; it never enters the
// served verdict.
func selfReportedModel(message string) string {
	for _, line := range strings.Split(message, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if v, ok := strings.CutPrefix(line, selfReportPrefix); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}
	return ""
}

// servedWarning renders the non-blocking warning for a drift or unknown
// verdict; "" for every other verdict.
func servedWarning(obs ServedObservation) string {
	expected := obs.ExpectedModel
	if expected == "" {
		expected = "-"
	}
	served := "none"
	if len(obs.ServedModels) > 0 {
		served = strings.Join(obs.ServedModels, ", ")
	}
	switch obs.Verdict {
	case ServedVerdictDrift:
		return fmt.Sprintf("served-model: %s was expected to run on %s but was served by [%s].",
			obs.AgentType, expected, served)
	case ServedVerdictUnknown:
		return fmt.Sprintf("served-model: the model that served %s could not be determined (%s); expected %s, served [%s].",
			obs.AgentType, obs.Cause, expected, served)
	default:
		return ""
	}
}

// observeServedModel classifies the stopping subagent's transcript and appends
// one row to the audit log. It returns the observation and the warning to
// surface ("" when there is nothing to warn about).
func (h *subagentStopHandler) observeServedModel(input *HookInput) (ServedObservation, string) {
	if input == nil {
		return ServedObservation{}, ""
	}
	path := subagentTranscriptPath(input)
	if path == "" {
		// @MX:NOTE: [AUTO] a payload naming no transcript at all (neither agent_transcript_path nor transcript_path + agent_id) identifies no subagent run to observe, so it yields no row and no warning; the runtime always sends transcript_path and agent_id, so every real stop still reaches the unknown-never-ok path when its file is absent
		return ServedObservation{}, ""
	}
	obs := ObserveServedModel(path, input.AgentType, h.loadedConfig())
	served := obs.ServedModels
	if served == nil {
		served = []string{}
	}
	appendAuditJSONL(servedAuditRoot(), servedAuditRecord{
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		SessionID:         input.SessionID,
		Agent:             obs.AgentType,
		DeclaredModel:     obs.DeclaredModel,
		ResolvedModel:     obs.ResolvedModel,
		Verdict:           obs.Verdict,
		Source:            servedObservationSource,
		AgentID:           input.AgentID,
		ServedModels:      served,
		SelfReportedModel: selfReportedModel(input.LastAssistantMessage),
		ExpectedModel:     obs.ExpectedModel,
		Cause:             obs.Cause,
	})
	return obs, servedWarning(obs)
}

// servedAuditRoot is the project root whose audit log receives the served row:
// the root the runtime exports as CLAUDE_PROJECT_DIR, the same root the
// PreToolUse row of the same spawn lands under. The hook input's cwd is
// deliberately not a fallback: in a worktree session it names a different
// tree, which would split one spawn's declared row and served row across two
// files and defeat the join the shared log exists for. With no exported root,
// or one that is not a MoAI project, no row is written (the warning still
// fires).
func servedAuditRoot() string {
	root := os.Getenv(config.EnvClaudeProjectDir)
	if root == "" {
		return ""
	}
	if _, err := os.Stat(filepath.Join(root, ".moai")); err != nil {
		return ""
	}
	return root
}

// appendSystemMessage adds msg to out's SystemMessage on its own line, creating
// an output when out is nil. An empty msg returns out unchanged.
func appendSystemMessage(out *HookOutput, msg string) *HookOutput {
	if strings.TrimSpace(msg) == "" {
		return out
	}
	if out == nil {
		return &HookOutput{SystemMessage: msg}
	}
	if out.SystemMessage == "" {
		out.SystemMessage = msg
		return out
	}
	out.SystemMessage += "\n" + msg
	return out
}
