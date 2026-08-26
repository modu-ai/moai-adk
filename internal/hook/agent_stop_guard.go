// agent_stop_guard.go — stop-registry + SendMessage guard against the revival
// of stopped teammates.
//
// TaskStop halts a named teammate but does not reclaim its name address: any
// later SendMessage delivered to that name resumes the agent from its
// transcript (a documented Claude Code runtime feature, v2.1.77+). This file
// adds the moai-side mechanism beneath the stopped-teammate doctrine:
//
//	record   — PostToolUse(TaskStop) rides the matcher-less harness-observe
//	           dispatch; RecordAgentStop persists a per-session stop record
//	           and appends one stop_recorded audit row.
//	observe  — PreToolUse(SendMessage) consults the same registry and appends
//	           one send_observed row per issuance; a registry hit surfaces a
//	           non-blocking advisory naming the stopped teammate.
//	deny     — enforcement layer (M2): opt-in gate, sentinel-prefixed deny.
//
// Fail-open is the house norm (branch_guard.go, agent_model_guard.go): the
// guard denies ONLY on positive evidence — a parsed recipient that matches a
// live entry in the SAME session's registry with the gate enabled. Every
// uncertain path (unparseable payload, absent recipient, unreadable or
// missing registry) observes without denying. An observation failure never
// blocks a tool call; an enforcement bug must never wedge a session.
//
// The registry is per-session by design (one team per session): entries never
// leak across sessions, so legitimate name reuse in a later session is
// unaffected. The audit trail (REQ-TRG-001) records regardless of the
// enforcement gate's state.
package hook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Audit event kinds (REQ-TRG-001). M1 observes stop_recorded and
// send_observed; send_denied, respawn_cleared, and session_cleared join with
// the M2 enforcement and lifecycle layers.
const (
	// AgentStopKindStopRecorded — one row per observed TaskStop completion
	// that resolved to an attributable target.
	AgentStopKindStopRecorded = "stop_recorded"
	// AgentStopKindSendObserved — one row per observed SendMessage issuance,
	// including every uncertain variant (fail-open observe-only).
	AgentStopKindSendObserved = "send_observed"
)

// Audit decision values. "recorded" accompanies a persisted stop entry;
// "allowed" accompanies an observed send (M1: every observed send).
const (
	agentStopDecisionRecorded = "recorded"
	agentStopDecisionAllowed  = "allowed"
)

// agentStopAuditFileName is the guard's JSONL audit log under
// <root>/.moai/logs/ — aggregation-shaped like agent-model-audit.jsonl, one
// structured row per observed event.
const agentStopAuditFileName = "agent-stop-audit.jsonl"

// AgentStopEntry is one stop record. Name is the teammate's name address;
// AgentID carries the runtime agent identifier when the TaskStop payload
// exposed one. A record keyed by identity (name when present, else agent id)
// so repeated stops of the same target upsert instead of duplicating.
type AgentStopEntry struct {
	Name      string `json:"name"`
	AgentID   string `json:"agent_id,omitempty"`
	StoppedAt string `json:"stopped_at"`
}

// identity is the upsert key: the name address when the payload carried one,
// else the agent id.
func (e AgentStopEntry) identity() string {
	return firstNonEmpty(e.Name, e.AgentID)
}

// AgentStopRegistry is the per-session stop registry persisted at
// <root>/.moai/state/agent-stops/<session-id>.json. Never a global list —
// cross-session name reuse must stay unaffected.
type AgentStopRegistry struct {
	SessionID string           `json:"session_id"`
	Entries   []AgentStopEntry `json:"entries"`
}

// AgentStopAuditRecord is one JSONL row. Name and AgentID carry no omitempty
// so every row exposes all six REQ-TRG-001 fields (empty string when
// unknown) — post-hoc correlation reads a uniform shape.
type AgentStopAuditRecord struct {
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	AgentID   string `json:"agent_id"`
	Decision  string `json:"decision"`
}

// agentStopRegistryPath returns the per-session registry file path.
func agentStopRegistryPath(projectRoot, sessionID string) string {
	return filepath.Join(projectRoot, ".moai", "state", "agent-stops", sessionID+".json")
}

// loadAgentStopRegistry reads the session's registry. It returns nil on every
// uncertainty — absent file, unreadable, or unparseable content — so callers
// fail open with no special-case branch.
func loadAgentStopRegistry(projectRoot, sessionID string) *AgentStopRegistry {
	if projectRoot == "" || sessionID == "" {
		return nil
	}
	data, err := os.ReadFile(agentStopRegistryPath(projectRoot, sessionID))
	if err != nil {
		return nil
	}
	var reg AgentStopRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil
	}
	return &reg
}

// upsertAgentStopEntry inserts or refreshes an entry keyed by identity. The
// file is written whole (read-modify-write); concurrent writers within one
// session are not expected — the lead stops one teammate at a time.
func upsertAgentStopEntry(projectRoot, sessionID string, entry AgentStopEntry) error {
	reg := loadAgentStopRegistry(projectRoot, sessionID)
	if reg == nil {
		reg = &AgentStopRegistry{SessionID: sessionID}
	}
	found := false
	for i := range reg.Entries {
		if reg.Entries[i].identity() == entry.identity() {
			reg.Entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		reg.Entries = append(reg.Entries, entry)
	}
	dir := filepath.Dir(agentStopRegistryPath(projectRoot, sessionID))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create agent-stops dir: %w", err)
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal stop registry: %w", err)
	}
	if err := os.WriteFile(agentStopRegistryPath(projectRoot, sessionID), append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write stop registry: %w", err)
	}
	return nil
}

// lookupAgentStopEntry matches an already-parsed recipient against the
// session's registry — by name address or by agent id, same session only.
// Absent registry or no match returns (nil, false); the caller allows.
func lookupAgentStopEntry(projectRoot, sessionID, recipient string) (*AgentStopEntry, bool) {
	if recipient == "" {
		return nil, false
	}
	reg := loadAgentStopRegistry(projectRoot, sessionID)
	if reg == nil {
		return nil, false
	}
	for i := range reg.Entries {
		if reg.Entries[i].Name == recipient {
			return &reg.Entries[i], true
		}
		if reg.Entries[i].AgentID != "" && reg.Entries[i].AgentID == recipient {
			return &reg.Entries[i], true
		}
	}
	return nil, false
}

// appendAgentStopAudit appends one record to <projectRoot>/.moai/logs/.
// Every failure path is silent-and-continue, matching the
// agent-model-audit precedent: an audit failure may never fail the observed
// tool call.
func appendAgentStopAudit(projectRoot string, rec AgentStopAuditRecord) {
	if projectRoot == "" {
		return
	}
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	line, err := json.Marshal(rec)
	if err != nil {
		slog.Warn("agent_stop_guard: failed to marshal audit record", "error", err)
		return
	}
	logsDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o750); err != nil {
		slog.Warn("agent_stop_guard: failed to create logs dir", "dir", logsDir, "error", err)
		return
	}
	path := filepath.Join(logsDir, agentStopAuditFileName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		slog.Warn("agent_stop_guard: failed to open audit log", "path", path, "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(line, '\n')); err != nil {
		slog.Warn("agent_stop_guard: failed to append audit record", "path", path, "error", err)
	}
}

// extractTaskStopTarget parses a TaskStop tool_input payload defensively.
// The exact wire shape is verified by the live-firing recipe (plan.md §E-P1)
// rather than assumed: both snake_case and camelCase variants of the
// plausible target fields are accepted. It returns ok=false when no target
// field resolves, so the caller can skip recording without a special-case
// branch.
func extractTaskStopTarget(toolInput json.RawMessage) (name, agentID string, ok bool) {
	if len(toolInput) == 0 {
		return "", "", false
	}
	var parsed struct {
		Name       string `json:"name"`
		AgentName  string `json:"agent_name"`
		AgentNameC string `json:"agentName"`
		AgentID    string `json:"agent_id"`
		AgentIDC   string `json:"agentId"`
		TaskID     string `json:"task_id"`
		TaskIDC    string `json:"taskId"`
	}
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return "", "", false
	}
	name = firstNonEmpty(parsed.Name, parsed.AgentName, parsed.AgentNameC)
	agentID = firstNonEmpty(parsed.AgentID, parsed.AgentIDC, parsed.TaskID, parsed.TaskIDC)
	if name == "" && agentID == "" {
		return "", "", false
	}
	return name, agentID, true
}

// parseSendRecipient extracts the addressed recipient from a SendMessage
// tool_input payload. The sanctioned "name [ref]" disambiguated addressing
// form is parsed and stripped before any registry comparison — only the
// bracketed suffix of a trailing " [" span is removed, so a malformed suffix
// leaves the raw string intact and simply fails to match (fail-open).
func parseSendRecipient(toolInput json.RawMessage) (recipient string, ok bool) {
	if len(toolInput) == 0 {
		return "", false
	}
	var parsed struct {
		To string `json:"to"`
	}
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return "", false
	}
	to := strings.TrimSpace(parsed.To)
	if to == "" {
		return "", false
	}
	if strings.HasSuffix(to, "]") {
		if i := strings.LastIndex(to, " ["); i > 0 {
			base := strings.TrimSpace(to[:i])
			if base == "" {
				return "", false
			}
			to = base
		}
	}
	return to, true
}

// RecordAgentStop is the PostToolUse(TaskStop) recorder entry point, invoked
// by the matcher-less harness-observe dispatch. It persists the stop record
// (idempotent upsert) and appends exactly one stop_recorded audit row.
// Unattributable events — an empty session id, an unparseable payload, or no
// resolvable target — record nothing and never fail the hook: recording is
// observation, and an observation failure may not surface to the caller.
func RecordAgentStop(projectRoot string, input *HookInput) {
	if input == nil || projectRoot == "" {
		return
	}
	if input.SessionID == "" {
		slog.Warn("agent_stop_guard: TaskStop completion without session_id; not recorded")
		return
	}
	name, agentID, ok := extractTaskStopTarget(input.ToolInput)
	if !ok {
		slog.Warn("agent_stop_guard: TaskStop completion without a resolvable target; not recorded",
			"session_id", input.SessionID)
		return
	}
	stoppedAt := time.Now().UTC().Format(time.RFC3339)
	if err := upsertAgentStopEntry(projectRoot, input.SessionID, AgentStopEntry{
		Name:      name,
		AgentID:   agentID,
		StoppedAt: stoppedAt,
	}); err != nil {
		slog.Warn("agent_stop_guard: stop registry write failed; not audited as recorded",
			"session_id", input.SessionID, "error", err)
		return
	}
	appendAgentStopAudit(projectRoot, AgentStopAuditRecord{
		Timestamp: stoppedAt,
		SessionID: input.SessionID,
		Kind:      AgentStopKindStopRecorded,
		Name:      name,
		AgentID:   agentID,
		Decision:  agentStopDecisionRecorded,
	})
}

// checkStopGuard is the PreToolUse(SendMessage) entry point. M1: observe and
// advise only — every issuance appends exactly one send_observed row, and a
// recipient matching a live entry in the same session's registry surfaces a
// non-blocking advisory naming the stopped teammate. No deny decision exists
// at this layer yet; the M2 enforcement layer adds the opt-in deny.
func (h *preToolHandler) checkStopGuard(input *HookInput) (decision, reason, advisory string) {
	recipient, parsed := parseSendRecipient(input.ToolInput)

	rec := AgentStopAuditRecord{
		SessionID: input.SessionID,
		Kind:      AgentStopKindSendObserved,
		Name:      recipient,
		Decision:  agentStopDecisionAllowed,
	}
	if parsed && input.SessionID != "" {
		if entry, found := lookupAgentStopEntry(h.projectRoot(), input.SessionID, recipient); found {
			rec.AgentID = entry.AgentID
			advisory = fmt.Sprintf(
				"agent-stop: SendMessage is addressed to %q, a teammate stopped in this session. "+
					"A message can revive it as an ownerless writer. Route coordination through the "+
					"owning orchestrator, or respawn the name deliberately with a fresh spawn.",
				recipient)
		}
	}
	appendAgentStopAudit(h.projectRoot(), rec)
	return "", "", advisory
}
