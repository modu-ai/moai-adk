package hook

// session_start_compact.go implements the SPEC-INFINITE-GOAL-001 REQ-5
// SessionStart(matcher: compact) re-inject. On auto-compact, the post-compact
// model re-orients from a short stdout re-injection carrying: the goal
// condition, the active SPEC-id, the last-verified mechanical state, and the
// single next action. No-op when no goal is armed or source != "compact".
//
// Registered as a 4th SessionStart handler in deps.go (the registry accumulates
// additionalContext across all SessionStart handlers). It filters source ==
// "compact" internally (mirroring handoff_inject.go's source filtering). The
// wrapper .claude/hooks/moai/handle-session-start-compact.sh exposes a manual /
// isolated-invocation surface (moai hook session-start-compact) and is template-
// mirrored; production firing rides on the deps.go registration via the existing
// handle-session-start.sh → moai hook session-start chain (no separate
// settings.json entry, avoiding double-fire).
//
// Advisory-Check Discipline (coding-standards.md): constant-cost (one goal-file
// read + a bounded scan of .moai/specs/ for the in-progress SPEC), time-boxed,
// fail-open (any read error → empty output, never blocks the session).

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// sessionStartCompactHandler re-injects goal + SPEC context on SessionStart
// events whose source is "compact" (auto-compact boundary).
type sessionStartCompactHandler struct{}

// NewSessionStartCompactHandler creates the SessionStart(compact) re-inject
// handler.
func NewSessionStartCompactHandler() Handler {
	return &sessionStartCompactHandler{}
}

// EventType returns EventSessionStart.
func (h *sessionStartCompactHandler) EventType() EventType { return EventSessionStart }

// Handle dispatches the SessionStart event; the source filter (compact only)
// lives in handle.
func (h *sessionStartCompactHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	projectDir := resolveProjectDir(input)
	if projectDir == "" {
		return &HookOutput{}, nil
	}
	return h.handle(ctx, projectDir, input)
}

// handle is the testable core: it filters by source and emits the re-injection.
func (h *sessionStartCompactHandler) handle(_ context.Context, projectDir string, input *HookInput) (*HookOutput, error) {
	if input == nil || input.Source != "compact" {
		return &HookOutput{}, nil
	}
	sessionID := ""
	if input != nil {
		sessionID = input.SessionID
	}
	text := renderCompactReinject(projectDir, sessionID)
	if strings.TrimSpace(text) == "" {
		return &HookOutput{}, nil
	}
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     string(EventSessionStart),
			AdditionalContext: text,
		},
	}, nil
}

// renderCompactReinject builds the post-compact re-injection text for the given
// session. Returns "" (no-op) when no armed goal exists. Best-effort + fail-open:
// any read error returns "".
func renderCompactReinject(projectDir, sessionID string) string {
	if projectDir == "" || sessionID == "" {
		return ""
	}
	g, err := goal.LoadGoal(projectDir, sessionID)
	if err != nil || g == nil {
		return ""
	}
	if g.Status != goal.StatusArmed {
		return ""
	}

	var b strings.Builder
	b.WriteString("[moai goal re-inject after auto-compact]\n")
	// (a) goal condition
	b.WriteString("Goal condition: ")
	b.WriteString(g.Goal)
	b.WriteString("\n")

	// (b) active SPEC-id (scan for the in-progress SPEC)
	specID := findInProgressSPEC(projectDir)
	if specID != "" {
		b.WriteString("Active SPEC: ")
		b.WriteString(specID)
		b.WriteString("\n")
		// progress.md tail (last 5 non-empty lines)
		tail := readProgressTail(projectDir, specID, 5)
		if tail != "" {
			b.WriteString("SPEC progress tail:\n")
			b.WriteString(tail)
			b.WriteString("\n")
		}
	}

	// (c) last-verified mechanical state (last progress entry + condition list)
	if len(g.Progress) > 0 {
		last := g.Progress[len(g.Progress)-1]
		fmt.Fprintf(&b, "Last-verified state (turn %d): %s\n", last.Turn, last.Note)
	} else {
		b.WriteString("Last-verified state: (no progress entries yet)\n")
	}
	if len(g.Conditions) > 0 {
		var cmds []string
		for _, c := range g.Conditions {
			if c.Type == goal.ConditionMechanical {
				cmds = append(cmds, fmt.Sprintf("%q (expect exit %d)", c.Cmd, c.ExpectExit))
			}
		}
		if len(cmds) > 0 {
			b.WriteString("Mechanical conditions: " + strings.Join(cmds, "; ") + "\n")
		}
	}

	// (d) single next action — derived from the last progress note
	b.WriteString("Next action: ")
	b.WriteString(nextActionFromGoal(g))
	b.WriteString("\n")

	return b.String()
}

// nextActionFromGoal derives a single next action from the goal's last state.
// Conservative + generic: re-run the goal's mechanical conditions and advance
// toward the next unsatisfied acceptance criterion.
func nextActionFromGoal(g *goal.Goal) string {
	if len(g.Progress) > 0 {
		last := g.Progress[len(g.Progress)-1].Note
		if strings.Contains(last, "failed") || strings.Contains(last, "mechanical condition failed") {
			return "re-run the goal's mechanical conditions; the last turn reported a failure — diagnose and fix, then re-arm"
		}
	}
	return "advance toward the next unsatisfied acceptance criterion; re-run the goal's mechanical conditions to verify"
}

// findInProgressSPEC scans .moai/specs/ for the SPEC with status: in-progress.
// Returns the most recently updated one (by frontmatter `updated:`). Bounded by
// the number of SPECs; fail-open (any error → "").
func findInProgressSPEC(projectDir string) string {
	specsDir := filepath.Join(projectDir, ".moai", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return ""
	}
	type cand struct {
		id      string
		updated string
	}
	var cands []cand
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		specPath := filepath.Join(specsDir, e.Name(), "spec.md")
		data, err := os.ReadFile(specPath)
		if err != nil {
			continue
		}
		if !containsStatus(string(data), "in-progress") {
			continue
		}
		cands = append(cands, cand{id: e.Name(), updated: extractFrontmatterField(string(data), "updated")})
	}
	if len(cands) == 0 {
		return ""
	}
	sort.Slice(cands, func(i, j int) bool {
		return cands[i].updated > cands[j].updated // most recent first
	})
	return cands[0].id
}

// containsStatus is a minimal frontmatter status check (avoids a full YAML parse
// on the hot path). Frontmatter is delimited by leading and trailing `---`.
func containsStatus(content, status string) bool {
	lines := strings.Split(content, "\n")
	inFM := false
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "---" {
			if !inFM {
				inFM = true // opening delimiter
				continue
			}
			break // closing delimiter → end of frontmatter
		}
		if !inFM {
			continue
		}
		if strings.HasPrefix(l, "status:") {
			return strings.Contains(strings.ToLower(l), status)
		}
	}
	return false
}

// extractFrontmatterField returns the value of a top-level frontmatter field.
func extractFrontmatterField(frontmatter, field string) string {
	for _, line := range strings.Split(frontmatter, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, field+":") {
			return strings.TrimSpace(strings.TrimPrefix(l, field+":"))
		}
	}
	return ""
}

// readProgressTail returns the last n non-empty lines of the SPEC's progress.md.
func readProgressTail(projectDir, specID string, n int) string {
	path := filepath.Join(projectDir, ".moai", "specs", specID, "progress.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	var nonEmpty []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty = append(nonEmpty, l)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	start := len(nonEmpty) - n
	if start < 0 {
		start = 0
	}
	return strings.Join(nonEmpty[start:], "\n")
}
