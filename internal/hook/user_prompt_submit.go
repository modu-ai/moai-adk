// Resolution: KEEP — SPEC detect, session title, workflow keyword routing.
package hook

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/workflow"
)

// specFilePattern is the glob pattern for finding spec.md files in SPEC directories.
const specFilePattern = ".moai/specs/*/spec.md"

// promptPreviewMaxLen is the maximum length of prompt preview in logs.
const promptPreviewMaxLen = 100

// titleMaxRunes is the maximum number of runes in a derived session title before
// truncation. The transcript text is often Korean, so the limit is measured in
// runes (not bytes) to avoid splitting a multi-byte character.
const titleMaxRunes = 60

// titleMinRunes is the minimum rune count for a derived title to be usable;
// shorter results (empty or a single character) are discarded.
const titleMinRunes = 2

// titleEllipsis is appended to a derived title only when truncation occurs.
const titleEllipsis = "…" // U+2026 HORIZONTAL ELLIPSIS

// transcriptMaxLineBytes is the per-line scanner buffer ceiling. Transcript
// lines (tool outputs) can be very long, so the buffer is raised well above the
// bufio default of 64 KiB. A line exceeding this ceiling is skipped, not errored.
const transcriptMaxLineBytes = 8 * 1024 * 1024 // 8 MiB

// userPromptSubmitHandler handles UserPromptSubmit events.
// It generates a content-bearing session title and returns it to Claude Code.
type userPromptSubmitHandler struct {
	cfg ConfigProvider
}

// NewUserPromptSubmitHandler creates a new UserPromptSubmit event handler.
func NewUserPromptSubmitHandler(cfg ConfigProvider) Handler {
	return &userPromptSubmitHandler{cfg: cfg}
}

// EventType returns EventUserPromptSubmit.
func (h *userPromptSubmitHandler) EventType() EventType {
	return EventUserPromptSubmit
}

// workflowKeywords are prompt keywords that indicate an active MoAI workflow context.
var workflowKeywords = []string{"loop", "run", "plan"}

// detectWorkflowContext checks whether the prompt contains any workflow keywords
// and returns a non-empty additionalContext string if a match is found.
func detectWorkflowContext(prompt string) string {
	lower := strings.ToLower(prompt)
	for _, kw := range workflowKeywords {
		if strings.Contains(lower, kw) {
			return "workflow keyword '" + kw + "' detected — MoAI workflow context may be active"
		}
	}
	return ""
}

// Handle processes a UserPromptSubmit event.
// It builds a content-bearing session title (from an active SPEC or the first
// user prompt in the transcript) and returns it via hookSpecificOutput.
// Errors are silently handled; an empty title is returned without blocking the
// prompt so the hook never stalls the user's Claude Code session.
func (h *userPromptSubmitHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	// Log prompt for audit purposes (truncated to promptPreviewMaxLen chars)
	prompt := input.Prompt
	preview := prompt
	if len(preview) > promptPreviewMaxLen {
		preview = preview[:promptPreviewMaxLen] + "..."
	}
	slog.Info("user prompt submitted",
		"session_id", input.SessionID,
		"prompt_preview", preview,
	)

	// Routing-ledger seam A (SPEC-HARNESS-LEARNING-EVO-001 REQ-HLE-002).
	// Ensures a pending routing row exists for the session, derived from hook
	// input rather than from an orchestrator instruction. Fail-open and inert
	// while either harness observation gate is closed; it never affects the
	// output below.
	RoutingSeamUserPromptSubmit(input)

	// Build session title (errors are silently ignored, falls back to empty title)
	title := h.buildSessionTitle(ctx, input.CWD, input.TranscriptPath)

	// Detect workflow context
	additionalCtx := detectWorkflowContext(prompt)

	// Return empty output if no context to report
	if title == "" && additionalCtx == "" {
		return &HookOutput{}, nil
	}

	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     string(EventUserPromptSubmit),
			SessionTitle:      title,
			AdditionalContext: additionalCtx,
		},
	}, nil
}

// buildSessionTitle generates a content-bearing, session-stable title.
//
// Policy, in order:
//  1. First-wins guard: if the session already has any title — Claude Code's
//     native ai-title, a user /rename, or a custom-title this hook wrote on an
//     earlier prompt — return "". Re-emitting a title on every UserPromptSubmit
//     would re-set it each turn (so every session in a project shows the same
//     title) and silently clobber /rename.
//  2. No title yet, active SPEC → "SPEC-ID: heading" (content-bearing, stable).
//  3. No title yet, no SPEC, transcript has a first user prompt → derive a title
//     from that FIRST prompt (it defines the session topic).
//  4. No title yet, no SPEC, no usable first prompt → "" so the native ai-title
//     generator has a clear field.
//
// On any error the result is an empty string.
func (h *userPromptSubmitHandler) buildSessionTitle(ctx context.Context, cwd, transcriptPath string) string {
	facts := inspectTranscript(ctx, transcriptPath)

	// First-wins guard (issue #1198): never re-emit a title once the session has
	// one, regardless of source. Without this, an active SPEC would override an
	// existing title — including a user /rename — on every prompt.
	if facts.hasAITitle || facts.hasCustomTitle {
		return ""
	}

	// No title yet, and this is a kanban or factory LEAD: the session's own name
	// is the title, so the operator finds it in the session list under the name
	// they and every peer already address it by (issue #1596). This branch sits
	// ABOVE the SPEC branch deliberately — a lead session sitting in a project
	// with SPECs is exactly the case that produced an unrelated SPEC heading as
	// a lead's title. It sits BELOW the first-wins guard just as deliberately:
	// a /rename still wins, as it does for every other source.
	if name := leadSessionTitle(); name != "" {
		return name
	}

	// No title yet. An active SPEC yields a content-bearing, stable title.
	if cwd != "" {
		if title := detectActiveSpec(cwd); title != "" {
			return title
		}
	}

	// No SPEC. Derive a stable title from the FIRST user prompt. The derived
	// title is already in the user's language because the user's own prompt is —
	// nothing is translated here; conversation_language is read for explicitness.
	if facts.firstUserText != "" {
		lang := h.conversationLanguage()
		title := deriveTitleFromPrompt(facts.firstUserText)
		slog.Debug("derived session title from transcript", "lang", lang, "title", title)
		return title
	}

	// No SPEC and no usable first prompt: emit no title so the native ai-title
	// generator has a clear field.
	return ""
}

// leadSessionTitle returns the lead session's resolved name, or "" when this
// session is not a kanban or factory lead.
//
// The value is published by the launcher (exportLeadSessionName) because the
// launcher is the only actor that knows which name actually reached the backend
// argv — the operator's own when they supplied one, the bare-or-bumped role
// otherwise. Reading it here rather than re-deriving the role is what keeps the
// title and the messaging address the same string; a title that guessed the role
// would disagree with the session's real name on every bumped lead.
func leadSessionTitle() string {
	return strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanLeadName))
}

// conversationLanguage returns the configured conversation_language, or "" when
// the config provider (or its config) is nil. It is nil-safe by contract because
// the handler must degrade gracefully rather than panic (hooks never block).
func (h *userPromptSubmitHandler) conversationLanguage() string {
	if h.cfg == nil {
		return ""
	}
	cfg := h.cfg.Get()
	if cfg == nil {
		return ""
	}
	return cfg.Language.ConversationLanguage
}

// detectActiveSpec searches cwd/.moai/specs/*/spec.md and returns the title
// of the most recently modified SPEC.
// Returns empty string if no SPEC is found or read fails.
func detectActiveSpec(cwd string) string {
	pattern := filepath.Join(cwd, specFilePattern)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}

	// Select the most recently modified spec.md
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

	// Extract SPEC ID from directory name
	specDirName := filepath.Base(filepath.Dir(latestMatch))
	specID := workflow.SpecIDPattern.FindString(specDirName)
	if specID == "" {
		return ""
	}

	// Read the first heading (line starting with #) from spec.md
	heading := readFirstHeading(latestMatch)
	if heading == "" {
		return specID
	}

	// Strip SPEC-ID prefix from heading if already present (e.g., "SPEC-SRS-003: Dashboard...")
	if trimmed, ok := strings.CutPrefix(heading, specID+": "); ok {
		heading = trimmed
	} else if trimmed, ok := strings.CutPrefix(heading, specID); ok {
		heading = strings.TrimLeft(trimmed, ": ")
	}

	if heading == "" {
		return specID
	}

	return fmt.Sprintf("%s: %s", specID, heading)
}

// readFirstHeading returns the first # heading text from a markdown file.
func readFirstHeading(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if trimmed, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(trimmed)
		}
	}
	return ""
}

// transcriptFacts carries the three facts extracted from a session transcript
// that the session-title policy needs.
type transcriptFacts struct {
	hasAITitle     bool   // a Claude Code native "ai-title" record is present
	hasCustomTitle bool   // a "custom-title" record (written by this hook) is present
	firstUserText  string // text of the FIRST "user" record, or "" if none found
}

// transcriptTitleMarker matches both "ai-title" and "custom-title" record types
// (both contain the substring "-title"). It is a cheap byte prefilter before the
// authoritative JSON decode.
var transcriptTitleMarker = []byte("-title")

// transcriptUserMarker matches "user" record types. It is a cheap byte prefilter
// before the authoritative JSON decode.
var transcriptUserMarker = []byte(`"user"`)

// transcriptLine is the minimal decode target for a single transcript JSONL
// record. Only the fields the title policy needs are decoded.
type transcriptLine struct {
	Type    string `json:"type"`
	Message struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// inspectTranscript streams the transcript at path line-by-line and extracts the
// three facts the title policy needs. It NEVER loads the whole file into memory
// (transcripts can exceed 1.5 MB) and NEVER errors: any failure (missing file,
// unreadable, malformed, oversized line) yields whatever facts were gathered so
// far, so the hook never blocks the prompt. It stops early once a title record
// (either kind) is found, since that forces an empty result regardless.
func inspectTranscript(ctx context.Context, path string) transcriptFacts {
	var facts transcriptFacts
	if path == "" {
		return facts
	}

	f, err := os.Open(path)
	if err != nil {
		return facts
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), transcriptMaxLineBytes)

	for scanner.Scan() {
		// Honor cancellation / the ≤5s hook budget threaded via ctx.
		select {
		case <-ctx.Done():
			return facts
		default:
		}

		line := scanner.Bytes()

		// Cheap prefilters: skip the JSON decode for lines that cannot carry a
		// needed fact. Most lines are large assistant/tool records we ignore.
		needTitle := bytes.Contains(line, transcriptTitleMarker)
		needUser := facts.firstUserText == "" && bytes.Contains(line, transcriptUserMarker)
		if !needTitle && !needUser {
			continue
		}

		var rec transcriptLine
		if err := json.Unmarshal(line, &rec); err != nil {
			continue // malformed line — skip, do not error
		}

		switch rec.Type {
		case "ai-title":
			facts.hasAITitle = true
		case "custom-title":
			facts.hasCustomTitle = true
		case "user":
			if facts.firstUserText == "" {
				text := transcriptContentText(rec.Message.Content)
				// A slash-command turn is recorded as a user message wrapped in
				// <local-command-caveat> tags. It is a command artifact, not a
				// prompt — skip it so the FIRST real prompt defines the title.
				if !isLocalCommandArtifact(text) {
					facts.firstUserText = text
				}
			}
		}

		// A title record (either kind) forces an empty result — stop scanning.
		if facts.hasAITitle || facts.hasCustomTitle {
			break
		}
	}
	// scanner.Err() (including bufio.ErrTooLong on an oversized line) is
	// intentionally ignored: the hook returns the facts gathered so far.
	return facts
}

// transcriptContentText extracts plain text from a transcript record's
// message.content, which Claude Code encodes either as a plain string or as an
// array of content blocks ({"type":"text","text":"..."}). Unknown shapes yield "".
func transcriptContentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Common case: content is a plain string.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	// Alternative case: content is an array of content blocks.
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var b strings.Builder
		for _, blk := range blocks {
			if blk.Text == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(blk.Text)
		}
		return b.String()
	}

	return ""
}

// directiveKeywordRE matches standalone MoAI directive keywords anywhere in the
// prompt: "ultrathink." (with trailing period), "ultrathink", and "ultracode".
// Word boundaries keep it from matching substrings (e.g. "ultrathinking").
var directiveKeywordRE = regexp.MustCompile(`(?i)\bultrathink\.|\bultrathink\b|\bultracode\b`)

// localCommandArtifactRE matches the tags Claude Code wraps a slash-command turn
// in: <local-command-caveat>...</local-command-caveat> plus the command metadata
// tags it carries (<command-name>, <command-message>, <command-args>,
// <local-command-stdout>, <local-command-stderr>). A user message wrapped in
// these tags is a command artifact, not a prompt. (?s) lets . span newlines so a
// multi-line wrapper is consumed whole. Replaced with a space (not empty) so the
// text on either side of the tags never fuses into one token.
var localCommandArtifactRE = regexp.MustCompile(`(?s)<(?:local-command-caveat|command-name|command-message|command-args|local-command-stdout|local-command-stderr)>.*?</(?:local-command-caveat|command-name|command-message|command-args|local-command-stdout|local-command-stderr)>`)

// isLocalCommandArtifact reports whether a user-record content string carries a
// Claude Code slash-command wrapper. It is the cheapest reliable signal that a
// transcript user record is a command artifact rather than a real prompt, used
// by inspectTranscript to skip such records when selecting the first prompt.
func isLocalCommandArtifact(text string) bool {
	return strings.Contains(text, "<local-command-caveat>")
}

// segmentTerminators end the first sentence / first line of a prompt. Newline is
// included so "first line" and "first sentence" collapse to a single earliest cut.
var segmentTerminators = []rune{'.', '?', '!', '。', '？', '！', '\n'} // . ? ! 。 ？ ！ \n

// deriveTitleFromPrompt turns a raw user prompt into a compact, content-bearing
// title. It is a pure string function — no LLM, no network, no translation. The
// title is naturally in the user's language because the prompt is.
//
// Steps: strip a leading slash-command token and standalone directive keywords;
// take the first sentence or first line (whichever is shorter); collapse
// whitespace; truncate to titleMaxRunes runes (appending an ellipsis only when
// truncation occurs). Returns "" if the result is empty or shorter than
// titleMinRunes runes.
func deriveTitleFromPrompt(prompt string) string {
	s := strings.TrimSpace(prompt)

	// 0. Strip Claude Code local-command artifact tags (slash-command wrapper).
	//    A turn recorded as <local-command-caveat>...<command-name>/x</command-name>
	//    is a command artifact, not a prompt. Without this step the caveat prose
	//    leaked into the title for every slash-command-started session (e.g.
	//    /clear produced "Caveat: The messages below are genera…").
	s = localCommandArtifactRE.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	// 1a. Strip a leading slash-command token (e.g. "/moai plan ..." -> "plan ...").
	//     A bare slash command (e.g. "/clear") reduces to "".
	if strings.HasPrefix(s, "/") {
		if i := strings.IndexFunc(s, unicode.IsSpace); i >= 0 {
			s = strings.TrimLeftFunc(s[i:], unicode.IsSpace)
		} else {
			s = ""
		}
	}

	// 1b. Strip standalone directive keywords wherever they appear. This runs
	//     BEFORE sentence splitting so the trailing "." of "ultrathink." cannot
	//     prematurely terminate the first sentence. Newlines are preserved so the
	//     "first line" rule below still works.
	s = directiveKeywordRE.ReplaceAllString(s, "")

	// 2. First sentence or first line, whichever is shorter. Because the
	//    terminator set includes newline, cutting at the earliest terminator
	//    yields exactly that shorter segment.
	s = firstSegment(s)

	// 3. Collapse internal whitespace runs to a single space and trim.
	s = strings.Join(strings.Fields(s), " ")

	// 4. Truncate to at most titleMaxRunes runes (rune-safe), appending an
	//    ellipsis only when truncation actually occurred.
	s = truncateRunes(s, titleMaxRunes)

	// 5. Reject empty or too-short results.
	if utf8.RuneCountInString(s) < titleMinRunes {
		return ""
	}
	return s
}

// firstSegment returns the substring up to (excluding) the first segment
// terminator (see segmentTerminators), or the whole string if none is present.
func firstSegment(s string) string {
	for i, r := range s {
		for _, t := range segmentTerminators {
			if r == t {
				return s[:i]
			}
		}
	}
	return s
}

// truncateRunes truncates s to at most maxRunes runes, appending titleEllipsis
// only when truncation occurs. Slicing on the rune slice guarantees no multi-byte
// character is split.
func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + titleEllipsis
}
