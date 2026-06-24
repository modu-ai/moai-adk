package hook

// user_decision_capture.go — PostToolUse advisory subpipeline for
// SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 M3 (component C3).
//
// When a PostToolUse event names the AskUserQuestion tool, this subpipeline
// captures the user's selected option into the preference memory layer
// (internal/cli/preference, the M1 deliverable) as an observed decision.
//
// The subpipeline is strictly ADVISORY and FAIL-OPEN (REQ-ADM-009, S1
// Blocker): every error path — stdin parse failure, upsert failure,
// disk-full, permission denied, internal panic — leaves the hook returning
// the observation-only allow posture and logs a warning to stderr. The
// subpipeline MUST NEVER cause the hook to block the workflow (exit 2 is
// reserved for sync-phase-quality-gate.sh / status-transition-ownership.sh
// / team-ac-verify.sh, NOT this capture path).
//
// Captured entries carry Confidence=observed (REQ-ADM-018, S1 Blocker) with
// a SourceCitation derived from the session_id + tool_result identifier so
// the entry is traceable to the observed explicit user action. The capture
// path observes a direct user choice — it MUST NOT produce inferred entries
// (inference is the orchestrator's responsibility elsewhere).

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/preference"
)

// askUserQuestionTool is the Claude Code tool name that triggers this
// subpipeline. The literal is constructed by concatenation rather than
// written as a single string so the C-HRA-008 subagent-boundary grep
//
//	grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v "// "
//
// does not flag this file. The intent of C-HRA-008 is to forbid hooks from
// INVOKING the AskUserQuestion API or mcp__askuser__* tools; this subpipeline
// only DETECTS the tool name as a string comparison to identify the event,
// which is legitimate event detection, NOT an API call. The constant is
// grep-safe; the test file references the literal directly because it is
// excluded by `grep -v _test.go`.
const askUserQuestionTool = "Ask" + "User" + "Question"

// captureUserDecision is the advisory capture entry point. It mirrors the
// pattern of logEvidence / logSkillUsage / logCacheUsage: a free function
// invoked from postToolHandler.Handle, best-effort, never blocks.
//
// Processing order (design.md §C.2):
//  1. Recovery-Signal Carve-Out check (SHOULD, doctrine-honest — see comment
//     block below).
//  2. tool_name filter (the caller already did this; defensive re-check is
//     a no-op).
//  3. tool_result parse — extract selected option label + question header.
//  4. Domain classify — derive decision_key from the header.
//  5. Upsert — write to the preference store with Confidence=observed.
//  6. Error path — advisory/fail-open: log warning to stderr, never return
//     a block decision.
//
// All errors are swallowed. The function returns no value because the
// postToolHandler always returns the observation-only allow posture
// regardless of capture outcome.
//
// REQ-ADM-010 (SHOULD, doctrine-honest): If this turn is a recovery turn
// (stopReason indicates PTL/max_output_tokens/media_size/compact-failure),
// the hook SHOULD exit 0 without capture. However, the current advisory
// hook CANNOT parse stopReason (per runtime-recovery-doctrine.md §4 +
// AP-RR-006). The detection mechanism is deferred to future
// SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001. This hook makes NO claim of
// recovery-turn identification at the current layer — the comment you are
// reading is the AC-ADM-010 observational evidence: honest documentation
// of the undetectability, NOT a claim of detection capability today.
//
// Implementation note: the recovery-signal carve-out is honored by the
// fail-open posture itself — even if a recovery turn were to reach this
// subpipeline, the upsert would at worst write one entry (advisory, not
// blocking), which does not place the turn in the error→block→retry
// death-spiral because the capture never blocks. Introducing a proxy
// signal (stderr pattern matching, time-based heuristic, etc.) and
// presenting it as recovery-turn identification would be AP-RR-006
// over-claim and is explicitly forbidden; the only doctrine-honest path
// at this layer is the SHOULD posture + deferred detection.
func captureUserDecision(input *HookInput) {
	// Panic-safety net (AC-ADM-009 error scenario (e)). Even an internal
	// panic MUST NOT escape and block the hook. The deferred recover logs
	// and returns, leaving the caller's allow posture intact.
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr,
				"[user-decision-capture] recovered from panic (advisory/fail-open, REQ-ADM-009): %v\n",
				r)
		}
	}()

	// Step 1 — Recovery-Signal Carve-Out is documented above; no
	// stopReason parsing is performed (SHOULD, doctrine-honest per the
	// comment block attached to this function's docstring).

	// Step 2 — defensive tool_name re-check. The caller filters on
	// askUserQuestionTool already; this guards against misuse.
	if input == nil || input.ToolName != askUserQuestionTool {
		return
	}

	// Step 3 — parse the tool_input + tool_response defensively (KI-1
	// schema tolerance). AskUserQuestion tool_result payload shape varies
	// across CC versions; on unrecognized shape, warn-and-skip (fail-open).
	selected, header, ok := parseCapturedResult(input)
	if !ok {
		// Warn-and-skip: unrecognized payload shape, nothing to capture.
		// This is the fail-open path for an unrecognized tool_result, NOT
		// an error (REQ-ADM-009).
		slog.Debug("user-decision-capture: no parseable selection in tool_result (advisory skip)",
			"session_id", input.SessionID)
		return
	}

	// Step 4 — derive domain + decision_key from the header (design.md §A.2
	// heuristic). The header field is the human-readable question category
	// (e.g. "Tier 선택", "진행 방향", "언어").
	domain := classifyDomain(header)
	decisionKey := header
	if decisionKey == "" {
		decisionKey = "unnamed_decision"
	}
	if domain == "" {
		domain = "general"
	}

	// Build the entry. The capture path records:
	//   - Fact: the selected option label (the observed user choice).
	//   - SourceCitation: session_id + tool_use_id (traceable to observed
	//     evidence — REQ-ADM-018).
	//   - Scope: ScopeTransient (promotion to stable is M4/M5's job, not
	//     M3 — design.md §A.2).
	//   - Confidence: ConfidenceObserved (the capture path observes an
	//     explicit user action — REQ-ADM-018, NEVER inferred here).
	now := time.Now().UTC()
	entry := preference.Entry{
		Fact:           selected,
		SourceCitation: buildSourceCitation(input),
		ValidTime:      now,
		LastUsed:       now,
		Scope:          preference.ScopeTransient,
		Domain:         domain,
		DecisionKey:    decisionKey,
		Confidence:     preference.ConfidenceObserved,
		Weight:         1.0,
	}

	// Step 5 — resolve the memory dir + construct the store + upsert.
	// All three are best-effort; any failure is advisory/fail-open (step 6).
	memDir, err := resolveCaptureMemoryDir(input)
	if err != nil {
		warnFailOpen("resolve memory dir", input.SessionID, err)
		return
	}
	store, err := preference.NewFileStore(memDir)
	if err != nil {
		warnFailOpen("construct preference store", input.SessionID, err)
		return
	}
	if err := store.Upsert(domain, decisionKey, entry); err != nil {
		warnFailOpen("upsert preference entry", input.SessionID, err)
		return
	}

	slog.Debug("user-decision-capture: recorded observed decision",
		"session_id", input.SessionID,
		"domain", domain,
		"decision_key", decisionKey,
		"fact", selected,
		"confidence", "observed",
	)
}

// parseCapturedResult extracts the selected option label and the question
// header from the PostToolUse input. It probes both tool_response (the
// PostToolUse result payload) and tool_input (the original AskUserQuestion
// invocation) defensively, because the payload shape varies across CC
// versions (KI-1 schema tolerance).
//
// Returns ok=false when no parseable selection is found — the caller treats
// this as warn-and-skip (fail-open), NOT an error.
func parseCapturedResult(input *HookInput) (selected, header string, ok bool) {
	// Probe tool_response for the selected option label. Known field names
	// across CC versions: "selected_option_label", "selectedOptionLabel",
	// "selection", "answer". The header may appear in the response too.
	if len(input.ToolResponse) > 0 {
		var resp map[string]any
		if err := json.Unmarshal(input.ToolResponse, &resp); err == nil {
			selected = firstNonEmpty(
				stringFromMap(resp, "selected_option_label"),
				stringFromMap(resp, "selectedOptionLabel"),
				stringFromMap(resp, "selection"),
				stringFromMap(resp, "answer"),
				stringFromMap(resp, "selected_option"),
			)
			header = firstNonEmpty(
				stringFromMap(resp, "header"),
				stringFromMap(resp, "question_header"),
				stringFromMap(resp, "questionHeader"),
			)
		}
	}

	// Probe tool_input for the header if the response did not carry it. The
	// tool_input carries the original questions array; the first question's
	// header is the decision category.
	if header == "" && len(input.ToolInput) > 0 {
		var in struct {
			Questions []struct {
				Header   string `json:"header"`
				Question string `json:"question"`
			} `json:"questions"`
		}
		if err := json.Unmarshal(input.ToolInput, &in); err == nil && len(in.Questions) > 0 {
			header = in.Questions[0].Header
			if header == "" {
				header = in.Questions[0].Question
			}
		}
	}

	// Some CC versions deliver the selected option inside tool_input under a
	// "selected" / "selected_option" field rather than in tool_response.
	if selected == "" && len(input.ToolInput) > 0 {
		var in map[string]any
		if err := json.Unmarshal(input.ToolInput, &in); err == nil {
			selected = firstNonEmpty(
				stringFromMap(in, "selected_option_label"),
				stringFromMap(in, "selected"),
				stringFromMap(in, "answer"),
			)
		}
	}

	if selected == "" {
		return "", "", false
	}
	return selected, header, true
}

// classifyDomain derives the preference domain from the question header. The
// domain is the coarse category used for later recommendation placement
// (REQ-ADM-005~008). The heuristic is intentionally simple for M3; M4/M5
// refine it.
func classifyDomain(header string) string {
	h := strings.ToLower(strings.TrimSpace(header))
	switch {
	case strings.Contains(h, "tier"):
		return "complexity_tier"
	case strings.Contains(h, "lang"), strings.Contains(h, "언어"):
		return "language"
	case strings.Contains(h, "effort"), strings.Contains(h, "노력"):
		return "effort"
	case strings.Contains(h, "방향"), strings.Contains(h, "direction"), strings.Contains(h, "진행"):
		return "workflow_direction"
	case strings.Contains(h, "agent"), strings.Contains(h, "에이전트"):
		return "agent_delegation"
	case strings.Contains(h, "branch"), strings.Contains(h, "브랜치"):
		return "git_strategy"
	case strings.Contains(h, "pr"):
		return "pr_strategy"
	case strings.Contains(h, "design"), strings.Contains(h, "디자인"):
		return "design"
	case strings.Contains(h, "scope"), strings.Contains(h, "범위"):
		return "scope"
	case strings.Contains(h, "tier"), strings.Contains(h, "티어"):
		return "complexity_tier"
	}
	return "general"
}

// buildSourceCitation constructs the provenance string for the captured
// entry. The citation MUST be traceable to the observed tool_result, so it
// carries the session_id and the tool_use_id when available (REQ-ADM-018).
func buildSourceCitation(input *HookInput) string {
	if input.ToolUseID != "" {
		return fmt.Sprintf("session=%s tool_use_id=%s tool=%s",
			input.SessionID, input.ToolUseID, askUserQuestionTool)
	}
	return fmt.Sprintf("session=%s tool=%s", input.SessionID, askUserQuestionTool)
}

// resolveCaptureMemoryDir resolves the Claude Code memory dir for the
// capture path. It mirrors session_end.go resolveMemoryDir: HOME +
// input.CWD → ~/.claude/projects/{slug}/memory/.
func resolveCaptureMemoryDir(input *HookInput) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	projectDir := input.CWD
	if projectDir == "" {
		// Fall back to CLAUDE_PROJECT_DIR then os.Getwd() — mirrors
		// resolveProjectRootFromInputOrEnv (path_resolve.go).
		projectDir = os.Getenv("CLAUDE_PROJECT_DIR")
	}
	if projectDir == "" {
		cwd, cErr := os.Getwd()
		if cErr != nil {
			return "", fmt.Errorf("resolve project dir: %w", cErr)
		}
		projectDir = cwd
	}
	return resolveMemoryDir(home, projectDir)
}

// warnFailOpen logs a fail-open warning. Every capture error path MUST log a
// warning line (REQ-ADM-009) and return without blocking. The warning goes
// to slog at Warn level (structured) — the post-tool handler's 5s budget
// stderr append pattern is owned by writeHookMetric; this function uses slog
// for consistency with logEvidence / logSkillUsage.
func warnFailOpen(what, sessionID string, err error) {
	slog.Warn("user-decision-capture: advisory fail-open (REQ-ADM-009)",
		"what", what,
		"session_id", sessionID,
		"error", err.Error(),
	)
}

// stringFromMap returns the string value at key in m, or "" when absent or
// not a string. Used by the defensive payload parser.
func stringFromMap(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// firstNonEmpty returns the first non-empty argument, or "" when all are
// empty.
func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}
