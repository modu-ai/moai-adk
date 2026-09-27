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
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// servedObservationSource marks a SubagentStop row. PreToolUse rows carry no
// source field, which is how the two are told apart.
const servedObservationSource = "subagent_stop"

// servedAuditRecord is the SubagentStop row of the agent-model audit log. Its
// first six keys share their names with agentModelAuditRecord so a reader can
// aggregate both row kinds by agent and session.
type servedAuditRecord struct {
	Timestamp     string   `json:"timestamp"`
	SessionID     string   `json:"session_id"`
	Agent         string   `json:"agent"`
	DeclaredModel string   `json:"declared_model"`
	ResolvedModel string   `json:"resolved_model"`
	Verdict       string   `json:"verdict"`
	Source        string   `json:"source"`
	ServedModels  []string `json:"served_models"`
}

// loadedConfig returns the loaded configuration, or nil when none is reachable.
func (h *subagentStopHandler) loadedConfig() *config.Config {
	if h == nil || h.cfg == nil {
		return nil
	}
	return h.cfg.Get()
}

// observeServedModel classifies the stopping subagent's transcript, appends
// one row to the audit log, and returns the warning to surface ("" when there
// is nothing to warn about).
func (h *subagentStopHandler) observeServedModel(input *HookInput) string {
	if input == nil {
		return ""
	}
	obs := ObserveServedModel(input.AgentTranscriptPath, input.AgentType, h.loadedConfig())
	served := obs.ServedModels
	if served == nil {
		served = []string{}
	}
	appendAuditJSONL(servedAuditRoot(), servedAuditRecord{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SessionID:     input.SessionID,
		Agent:         obs.AgentType,
		DeclaredModel: obs.DeclaredModel,
		ResolvedModel: obs.ResolvedModel,
		Verdict:       obs.Verdict,
		Source:        servedObservationSource,
		ServedModels:  served,
	})
	slog.Debug("served-model observation", "agent", obs.AgentType, "verdict", obs.Verdict)
	return ""
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
