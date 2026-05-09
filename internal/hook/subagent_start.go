package hook

import (
"bytes"
"context"
"fmt"
"log/slog"
"os"
"path/filepath"
"strings"

"gopkg.in/yaml.v3"

"github.com/modu-ai/moai-adk/internal/config"
)

// subagentStartHandler processes SubagentStart events.
// It logs subagent startup for session tracking and optionally injects project context.
type subagentStartHandler struct {
cfg ConfigProvider
projectDir string
}

// NewSubagentStartHandler creates a new SubagentStart event handler without config.
// This constructor preserves backward compatibility; no additionalContext is injected.
func NewSubagentStartHandler() Handler {
return &subagentStartHandler{}
}

// NewSubagentStartHandlerWithConfig creates a new SubagentStart event handler with
// config access for project context injection.
// projectDir is resolved from CLAUDE_PROJECT_DIR env var or os.Getwd() as fallback.
func NewSubagentStartHandlerWithConfig(cfg ConfigProvider) Handler {
dir := os.Getenv(config.EnvClaudeProjectDir)
if dir == "" {
if cwd, err := os.Getwd(); err == nil {
dir = cwd
}
}
return &subagentStartHandler{cfg: cfg, projectDir: dir}
}

// EventType returns EventSubagentStart.
func (h *subagentStartHandler) EventType() EventType {
return EventSubagentStart
}

// Handle processes a SubagentStart event.
// If config is available, it injects project context via additionalContext.
// Errors are non-blocking; an empty output is returned on failure.
func (h *subagentStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
slog.Info("subagent started",
"session_id", input.SessionID,
"agent_id", input.AgentID,
"agent_transcript_path", input.AgentTranscriptPath,
)

contextStr := h.buildContext(input)
if contextStr == "" {
return &HookOutput{}, nil
}

return &HookOutput{
HookSpecificOutput: &HookSpecificOutput{
HookEventName: string(EventSubagentStart),
AdditionalContext: contextStr,
},
}, nil
}

// buildContext constructs a concise project context string (under 200 chars).
// Returns empty string if no useful context is available.
func (h *subagentStartHandler) buildContext(input *HookInput) string {
if h.cfg == nil {
return ""
}

cfg := h.cfg.Get()
if cfg == nil {
return ""
}

var parts []string

if cfg.Project.Name != "" {
parts = append(parts, "project:"+cfg.Project.Name)
}
if string(cfg.Project.Type) != "" {
parts = append(parts, "type:"+string(cfg.Project.Type))
}
if cfg.Project.Language != "" {
parts = append(parts, "lang:"+cfg.Project.Language)
}

// Resolve project directory from input CWD, handler field, or env
dir := input.CWD
if dir == "" {
dir = h.projectDir
}

if spec := h.detectActiveSpec(dir); spec != "" {
parts = append(parts, "spec:"+spec)
}

if len(parts) == 0 {
return ""
}

result := strings.Join(parts, " | ")
// Truncate to stay under 200 chars
if len(result) > 199 {
result = result[:199]
}
return result
}

// detectActiveSpec returns the SPEC ID of the most recently modified spec.md under dir.
// Returns empty string if no SPEC is found.
func (h *subagentStartHandler) detectActiveSpec(dir string) string {
if dir == "" {
return ""
}

pattern := filepath.Join(dir, specFilePattern)
matches, err := filepath.Glob(pattern)
if err != nil || len(matches) == 0 {
return ""
}

var latestMatch string
var latestModTime int64
for _, match := range matches {
info, err := os.Stat(match)
if err != nil {
continue
}
mt := info.ModTime().UnixNano()
if mt > latestModTime {
latestModTime = mt
latestMatch = match
}
}
if latestMatch == "" {
return ""
}

// Extract SPEC ID from directory name (e.g., "SPEC-FOO-001")
specDirName := filepath.Base(filepath.Dir(latestMatch))
return specDirName
}

// ============================================================
// agentStartHandler: SubagentStart retired-rejection guard
// REQ-RA-004, REQ-RA-007, REQ-RA-008, REQ-RA-009, REQ-RA-012
// ============================================================

// agentFrontmatter parses YAML frontmatter from agent files and defines field structure.
// Rejects spawn when retired: true is present (REQ-RA-007).
type agentFrontmatter struct {
Retired bool `yaml:"retired"`
RetiredReplacement string `yaml:"retired_replacement"`
RetiredParamHint string `yaml:"retired_param_hint"`
}

// @MX:NOTE: [AUTO] agentStartHandler SPEC-V3R3-RETIRED-AGENT-001 5-layer defect chain
// Layer 1 (retired stub frontmatter invalid) is the core blocking entry point.
// When retired:true frontmatter is detected, returns block decision JSON + exit code 2
// to reject spawn before Claude Code Agent runtime allocates worktree (≤500ms response budget).
//
// agentStartHandler rejects retired agent spawns at SubagentStart events.
// REQ-RA-004: SubagentStart handler new registration
type agentStartHandler struct{}

// NewAgentStartHandler creates a SubagentStart event handler with retired-rejection guard.
// Option A: integrated into subagent_start.go (plan.md differs from agent_start.go naming).
// Decision rationale: clean integration without file duplication, EventSubagentStart not registered.
func NewAgentStartHandler() Handler {
return &agentStartHandler{}
}

// EventType returns EventSubagentStart.
func (h *agentStartHandler) EventType() EventType {
return EventSubagentStart
}

// Handle processes SubagentStart events.
// Allows when AgentName is missing or file not found (REQ-RA-008).
// Returns block decision when retired: true frontmatter is present (REQ-RA-007).
// Performance: single file read + YAML parse, no network I/O (REQ-RA-012 ≤500ms).
func (h *agentStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
agentName := input.AgentName
if agentName == "" {
// Pass-through when AgentName is not provided (REQ-RA-008)
return &HookOutput{}, nil
}

// Prevent path traversal: reject if slash or double-dot is included
if strings.Contains(agentName, "/") || strings.Contains(agentName, "..") {
slog.Warn("agentStartHandler: rejecting potential path traversal agent name",
"agent_name", agentName)
return &HookOutput{}, nil
}

// Determine projectDir: CWD priority, environment variable fallback
projectDir := input.CWD
if projectDir == "" {
projectDir = os.Getenv(config.EnvClaudeProjectDir)
}
if projectDir == "" {
if cwd, err := os.Getwd(); err == nil {
projectDir = cwd
}
}

fm, found, err := h.loadAgentFrontmatter(projectDir, agentName)
if err != nil {
// Fail-safe on parse error: allow
slog.Warn("agentStartHandler: frontmatter parse error, allowing with bypass",
"agent", agentName, "error", err)
return &HookOutput{}, nil
}
if !found {
// File not found → non-MoAI agent bypass (REQ-RA-008)
return &HookOutput{}, nil
}

if !fm.Retired {
// Active agent: allow (REQ-RA-008)
return &HookOutput{}, nil
}

// retired: true → block (REQ-RA-007)
reason := fmt.Sprintf(
"agent %s retired (SPEC-V3R3-RETIRED-AGENT-001), use %s",
agentName, fm.RetiredReplacement,
)
if fm.RetiredParamHint != "" {
reason += " with " + fm.RetiredParamHint
}

slog.Info("agentStartHandler: blocked retired agent spawn",
"agent", agentName,
"replacement", fm.RetiredReplacement,
)

return &HookOutput{
Decision: DecisionBlock,
Reason: reason,
}, nil
}

// loadAgentFrontmatter finds the agent file and parses YAML frontmatter.
// Search order: .claude/agents/moai/<name>.md → .claude/agents/<name>.md
// found=false means file does not exist (REQ-RA-008 bypass).
func (h *agentStartHandler) loadAgentFrontmatter(projectDir, agentName string) (*agentFrontmatter, bool, error) {
candidates := []string{
filepath.Join(projectDir, ".claude", "agents", "moai", agentName+".md"),
filepath.Join(projectDir, ".claude", "agents", agentName+".md"),
}

for _, path := range candidates {
data, err := os.ReadFile(path)
if os.IsNotExist(err) {
continue
}
if err != nil {
// Read error: treat as found=true but return parse error
return nil, true, fmt.Errorf("failed to read agent file %s: %w", path, err)
}

fm, err := parseAgentFrontmatter(data)
if err != nil {
return nil, true, fmt.Errorf("failed to parse frontmatter %s: %w", path, err)
}
return fm, true, nil
}

return nil, false, nil
}

// parseAgentFrontmatter extracts and parses YAML frontmatter (between --- delimiters) from Markdown files.
// Returns empty struct when frontmatter is absent (no error).
func parseAgentFrontmatter(data []byte) (*agentFrontmatter, error) {
fm := &agentFrontmatter{}

// Extract frontmatter starting with ---
if !bytes.HasPrefix(data, []byte("---")) {
// No frontmatter → return empty struct (treat as active agent)
return fm, nil
}

// Find second --- after first one
rest := data[3:]
// Skip newline
if len(rest) > 0 && rest[0] == '\n' {
rest = rest[1:]
} else if len(rest) > 0 && rest[0] == '\r' && len(rest) > 1 && rest[1] == '\n' {
rest = rest[2:]
}

yamlContent, _, ok := bytes.Cut(rest, []byte("\n---"))
if !ok {
// No closing --- → treat as no frontmatter
return fm, nil
}
if err := yaml.Unmarshal(yamlContent, fm); err != nil {
return nil, fmt.Errorf("YAML unmarshal failed: %w", err)
}

return fm, nil
}
