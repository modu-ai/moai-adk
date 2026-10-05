package hook

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Recommendation modes. `push` is the default a consumer resolves to when the
// key is absent; `pull` withholds the recommendation label. An unrecognized
// value is never coerced to either — it is recorded verbatim, so a typo is
// visible in the record rather than silently reading as the default.
const (
	RecommendationModePush = "push"
	RecommendationModePull = "pull"
)

// askUserObservationFileName is the observer's JSONL log under
// <root>/.moai/logs/ — aggregation-shaped like agent-stop-audit.jsonl, one
// structured row per observed AskUserQuestion issuance.
const askUserObservationFileName = "askuser-observations.jsonl"

// interviewSectionFileName is the config section carrying recommendation_mode.
// The observer reads it directly rather than through the typed config: the
// field is not yet on InterviewConfig, and an observer may not depend on a
// schema landing before it can record.
const interviewSectionFileName = "interview.yaml"

// AskUserObservationRecord is one JSONL row per AskUserQuestion issuance.
//
// Mode, LabelPresent, OptionCount and QuestionCount carry no omitempty so
// every row exposes a uniform shape — a row whose label_present is absent
// would be indistinguishable from one whose detector reported false, and the
// difference is exactly what the positive control measures.
//
// QuestionType is the one field that IS omitted when empty, deliberately. The
// AskUserQuestion payload carries no type tag, and whether a report preceded
// the call in the same turn is not visible to a PreToolUse hook, so no
// defensible classification exists at this layer. Emitting an inferred value
// as though it were an observation would make anything selecting on it
// vacuous. The field is retained as a descriptive slot for a payload that
// someday carries a real tag — never as a selector.
type AskUserObservationRecord struct {
	Timestamp     string `json:"timestamp"`
	SessionID     string `json:"session_id"`
	Mode          string `json:"mode"`
	LabelPresent  bool   `json:"label_present"`
	OptionCount   int    `json:"option_count"`
	QuestionCount int    `json:"question_count"`
	// PayloadParsed separates "the detector looked and found no label" from
	// "the payload could not be read". Without it an unparseable payload would
	// land as label_present:false and be counted as a genuine negative
	// observation, which it is not.
	PayloadParsed bool   `json:"payload_parsed"`
	QuestionType  string `json:"question_type,omitempty"`
}

// resolveRecommendationMode reads interview.recommendation_mode from the
// project's config section. Every uncertainty — absent root, absent file,
// unreadable file, unparseable YAML, absent key — resolves to push, the
// documented default, with no error surfaced. An unrecognized non-empty value
// is returned verbatim.
func resolveRecommendationMode(projectRoot string) string {
	if projectRoot == "" {
		return RecommendationModePush
	}
	path := filepath.Join(projectRoot, ".moai", "config", "sections", interviewSectionFileName)
	data, err := os.ReadFile(path) //nolint:gosec // path derived from the resolved project root
	if err != nil {
		return RecommendationModePush
	}
	var doc struct {
		Interview struct {
			RecommendationMode string `yaml:"recommendation_mode"`
		} `yaml:"interview"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return RecommendationModePush
	}
	mode := strings.TrimSpace(doc.Interview.RecommendationMode)
	if mode == "" {
		return RecommendationModePush
	}
	return mode
}

// recommendationLabelMarkers are the label markers the convention uses. The
// scan is containment rather than a strict suffix test: a strict suffix fails
// on a trailing space or a trailing period and would under-report, and
// under-reporting is the dangerous direction for a positive control — an
// inert detector satisfies a zero-violation check on every row.
var recommendationLabelMarkers = []string{"(권장)", "(Recommended)"}

// askUserPayload is the parsed shape of an AskUserQuestion tool_input. Only
// the fields the record needs are declared; anything else in the payload is
// ignored rather than asserted.
type askUserPayload struct {
	Questions []struct {
		Options []struct {
			Label string `json:"label"`
		} `json:"options"`
	} `json:"questions"`
}

// scanAskUserPayload parses the payload and reports the structural facts it
// actually carries. ok is false on every unreadable payload, and the counts
// are then zero — the caller records that state rather than a fabricated
// negative.
func scanAskUserPayload(toolInput json.RawMessage) (labelPresent bool, optionCount, questionCount int, ok bool) {
	if len(toolInput) == 0 {
		return false, 0, 0, false
	}
	var parsed askUserPayload
	if err := json.Unmarshal(toolInput, &parsed); err != nil {
		return false, 0, 0, false
	}
	questionCount = len(parsed.Questions)
	for _, q := range parsed.Questions {
		optionCount += len(q.Options)
		for _, opt := range q.Options {
			for _, marker := range recommendationLabelMarkers {
				if strings.Contains(opt.Label, marker) {
					labelPresent = true
				}
			}
		}
	}
	return labelPresent, optionCount, questionCount, true
}

// appendAskUserObservation appends one record to <projectRoot>/.moai/logs/.
// Every failure path is silent-and-continue, matching the agent-stop-audit
// precedent: an observation failure may never fail the observed tool call.
func appendAskUserObservation(projectRoot string, rec AskUserObservationRecord) {
	if projectRoot == "" {
		return
	}
	if rec.Timestamp == "" {
		rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	line, err := json.Marshal(rec)
	if err != nil {
		slog.Warn("askuser_observer: failed to marshal observation record", "error", err)
		return
	}
	logsDir := filepath.Join(projectRoot, ".moai", "logs")
	if err := os.MkdirAll(logsDir, 0o750); err != nil {
		slog.Warn("askuser_observer: failed to create logs dir", "dir", logsDir, "error", err)
		return
	}
	path := filepath.Join(logsDir, askUserObservationFileName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		slog.Warn("askuser_observer: failed to open observation log", "path", path, "error", err)
		return
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(line, '\n')); err != nil {
		slog.Warn("askuser_observer: failed to append observation record", "path", path, "error", err)
	}
}

// observeQuestionChannel is the PreToolUse(AskUserQuestion) entry point. It is
// observation only: it returns nothing, so no deny path can exist here in any
// mode, on any input, and no established permission decision can be displaced.
// Exactly one row is appended per issuance — an unparseable payload is still
// an issuance, recorded with payload_parsed false.
func (h *preToolHandler) observeQuestionChannel(input *HookInput) {
	if input == nil {
		return
	}
	labelPresent, optionCount, questionCount, parsed := scanAskUserPayload(input.ToolInput)
	appendAskUserObservation(h.projectRoot(), AskUserObservationRecord{
		SessionID:     input.SessionID,
		Mode:          resolveRecommendationMode(h.projectRoot()),
		LabelPresent:  labelPresent,
		OptionCount:   optionCount,
		QuestionCount: questionCount,
		PayloadParsed: parsed,
	})
}
