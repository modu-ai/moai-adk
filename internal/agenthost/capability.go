// Package agenthost describes how MoAI runtime events map to coding-agent
// host surfaces such as Claude Code, Codex, and OpenCode.
package agenthost

import (
	"fmt"
	"slices"
	"strings"
)

// Host is a supported coding-agent host.
type Host string

const (
	HostClaude   Host = "claude"
	HostCodex    Host = "codex"
	HostOpenCode Host = "opencode"
)

// SupportLevel describes how directly a host supports a MoAI event.
type SupportLevel string

const (
	SupportNative      SupportLevel = "native"
	SupportAdapter     SupportLevel = "adapter"
	SupportFallback    SupportLevel = "fallback"
	SupportUnsupported SupportLevel = "unsupported"
)

// Event is a host-neutral MoAI runtime event.
type Event string

const (
	EventSessionStart      Event = "SessionStart"
	EventPreToolUse        Event = "PreToolUse"
	EventPermissionRequest Event = "PermissionRequest"
	EventPostToolUse       Event = "PostToolUse"
	EventUserPromptSubmit  Event = "UserPromptSubmit"
	EventStop              Event = "Stop"
	EventSubagentStart     Event = "SubagentStart"
	EventSubagentStop      Event = "SubagentStop"
	EventPreCompact        Event = "PreCompact"
	EventPostCompact       Event = "PostCompact"
)

// Mapping describes one host's support for one MoAI runtime event.
type Mapping struct {
	Event       Event        `json:"event"`
	HostEvent   string       `json:"host_event"`
	Support     SupportLevel `json:"support"`
	Source      string       `json:"source"`
	Degradation string       `json:"degradation,omitempty"`
	Notes       string       `json:"notes,omitempty"`
}

// Matrix is the complete event mapping for one host.
type Matrix struct {
	Host     Host      `json:"host"`
	Source   string    `json:"source"`
	Mappings []Mapping `json:"mappings"`
}

var hostOrder = []Host{HostClaude, HostCodex, HostOpenCode}

// Hosts returns the supported hosts in stable display order.
func Hosts() []Host {
	return slices.Clone(hostOrder)
}

// ParseHost normalizes a host name.
func ParseHost(raw string) (Host, error) {
	host := Host(strings.ToLower(strings.TrimSpace(raw)))
	switch host {
	case HostClaude, HostCodex, HostOpenCode:
		return host, nil
	default:
		return "", fmt.Errorf("unsupported host %q: expected one of: claude, codex, opencode", raw)
	}
}

// MatrixFor returns the official-doc-grounded event compatibility matrix for a
// host. Source URLs intentionally point at host docs, not MoAI implementation
// files, because this matrix is the boundary between MoAI and external hosts.
func MatrixFor(host Host) (Matrix, error) {
	switch host {
	case HostClaude:
		return Matrix{
			Host:   HostClaude,
			Source: "https://docs.anthropic.com/en/docs/claude-code/hooks",
			Mappings: []Mapping{
				native(EventSessionStart, "SessionStart", "Claude Code hook event"),
				native(EventPreToolUse, "PreToolUse", "Claude Code hook event"),
				native(EventPermissionRequest, "PermissionRequest", "Claude Code hook event"),
				native(EventPostToolUse, "PostToolUse", "Claude Code hook event"),
				native(EventUserPromptSubmit, "UserPromptSubmit", "Claude Code hook event"),
				native(EventStop, "Stop", "Claude Code hook event"),
				native(EventSubagentStart, "SubagentStart", "Claude Code hook event"),
				native(EventSubagentStop, "SubagentStop", "Claude Code hook event"),
				native(EventPreCompact, "PreCompact", "Claude Code hook event"),
				native(EventPostCompact, "PostCompact", "Claude Code hook event"),
			},
		}, nil
	case HostCodex:
		return Matrix{
			Host:   HostCodex,
			Source: "https://developers.openai.com/codex/hooks",
			Mappings: []Mapping{
				native(EventSessionStart, "SessionStart", "Codex hook event; matcher filters startup|resume|clear|compact"),
				native(EventPreToolUse, "PreToolUse", "Codex hook event; matcher filters Bash, apply_patch/Edit/Write, and MCP tools"),
				native(EventPermissionRequest, "PermissionRequest", "Codex hook event"),
				native(EventPostToolUse, "PostToolUse", "Codex hook event"),
				native(EventUserPromptSubmit, "UserPromptSubmit", "Codex hook event; matcher is ignored"),
				native(EventStop, "Stop", "Codex hook event; matcher is ignored"),
				native(EventSubagentStart, "SubagentStart", "Codex hook event; matcher filters subagent type"),
				native(EventSubagentStop, "SubagentStop", "Codex hook event; matcher filters subagent type"),
				native(EventPreCompact, "PreCompact", "Codex hook event"),
				native(EventPostCompact, "PostCompact", "Codex hook event"),
			},
		}, nil
	case HostOpenCode:
		return Matrix{
			Host:   HostOpenCode,
			Source: "https://opencode.ai/docs/plugins/",
			Mappings: []Mapping{
				adapter(EventSessionStart, "session.created", "OpenCode plugin event", "No direct SessionStart hook payload parity; adapter must synthesize MoAI session fields."),
				adapter(EventPreToolUse, "tool.execute.before", "OpenCode plugin event", "Use plugin mutation/throw plus OpenCode permission rules instead of Claude/Codex permissionDecision JSON."),
				adapter(EventPermissionRequest, "permission.asked", "OpenCode plugin event", "Pair with permission.replied for observation; approval UI semantics differ."),
				adapter(EventPostToolUse, "tool.execute.after|file.edited", "OpenCode plugin events", "Split tool-result observation and file-edit observation."),
				fallback(EventUserPromptSubmit, "tui.prompt.append", "OpenCode TUI event", "No official prompt-submitted lifecycle event found; keep explicit /moai or $skill routing as fallback."),
				fallback(EventStop, "session.idle", "OpenCode session event", "Approximate turn completion; quality gate should also run via explicit moai gate/sync fallback."),
				fallback(EventSubagentStart, "session.created", "OpenCode session event", "Subagent child sessions exist, but no dedicated SubagentStart lifecycle hook is documented."),
				fallback(EventSubagentStop, "session.idle", "OpenCode session event", "Subagent stop must be inferred from child session idle/status metadata."),
				adapter(EventPreCompact, "experimental.session.compacting", "OpenCode experimental plugin hook", "OpenCode exposes a pre-compaction mutation point under an experimental event name."),
				adapter(EventPostCompact, "session.compacted", "OpenCode session event", "Post-compaction observation only; prompt/context mutation belongs to experimental.session.compacting."),
			},
		}, nil
	default:
		return Matrix{}, fmt.Errorf("unsupported host %q", host)
	}
}

// AllMatrices returns every host matrix in stable order.
func AllMatrices() []Matrix {
	out := make([]Matrix, 0, len(hostOrder))
	for _, host := range hostOrder {
		matrix, err := MatrixFor(host)
		if err == nil {
			out = append(out, matrix)
		}
	}
	return out
}

// Find returns the mapping for one event.
func (m Matrix) Find(event Event) (Mapping, bool) {
	for _, mapping := range m.Mappings {
		if mapping.Event == event {
			return mapping, true
		}
	}
	return Mapping{}, false
}

func native(event Event, hostEvent, notes string) Mapping {
	return Mapping{Event: event, HostEvent: hostEvent, Support: SupportNative, Notes: notes}
}

func adapter(event Event, hostEvent, notes, degradation string) Mapping {
	return Mapping{
		Event:       event,
		HostEvent:   hostEvent,
		Support:     SupportAdapter,
		Notes:       notes,
		Degradation: degradation,
	}
}

func fallback(event Event, hostEvent, notes, degradation string) Mapping {
	return Mapping{
		Event:       event,
		HostEvent:   hostEvent,
		Support:     SupportFallback,
		Notes:       notes,
		Degradation: degradation,
	}
}
