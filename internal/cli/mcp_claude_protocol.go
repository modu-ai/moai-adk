package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
)

// claudeReviewOutputSchema constrains model-authored content. Transport,
// authentication, usage, and build identity are added by Go and cannot be
// claimed by the model.
const claudeReviewOutputSchema = `{"type":"object","additionalProperties":false,"properties":{"verdict":{"type":"string","enum":["pass","fail","inconclusive"]},"summary":{"type":"string"},"findings":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"severity":{"type":"string","enum":["P0","P1","P2","P3"]},"title":{"type":"string"},"body":{"type":"string"},"file":{"type":"string"},"line":{"type":"integer"},"confidence":{"type":"number"},"recommendation":{"type":"string"}},"required":["severity","title","body"]}},"next_steps":{"type":"array","items":{"type":"string"}}},"required":["verdict","summary","findings","next_steps"]}`

type claudeAuthStatus struct {
	LoggedIn         bool   `json:"loggedIn"`
	AuthMethod       string `json:"authMethod"`
	APIProvider      string `json:"apiProvider"`
	SubscriptionType string `json:"subscriptionType"`
}

func parseClaudeAuthStatus(data []byte) (claudeAuthStatus, error) {
	var status claudeAuthStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return claudeAuthStatus{}, err
	}
	return status, nil
}

func (s claudeAuthStatus) subscriptionReady() bool {
	return s.LoggedIn && s.AuthMethod == "claude.ai" && s.APIProvider == "firstParty" && strings.TrimSpace(s.SubscriptionType) != ""
}

type claudeCLIUsage struct {
	InputTokens          *int64 `json:"inputTokens"`
	CacheReadInputTokens *int64 `json:"cacheReadInputTokens"`
	OutputTokens         *int64 `json:"outputTokens"`
}

type claudeCLIEnvelope struct {
	IsError          bool                      `json:"is_error"`
	APIErrorStatus   int                       `json:"api_error_status"`
	Result           string                    `json:"result"`
	StructuredOutput json.RawMessage           `json:"structured_output"`
	ModelUsage       map[string]claudeCLIUsage `json:"modelUsage"`
}

func claudeAuditAPIStatus(data []byte) int {
	var envelope claudeCLIEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return 0
	}
	return envelope.APIErrorStatus
}

type claudeParsedUsage struct {
	InputTokens       *int64
	CachedInputTokens *int64
	OutputTokens      *int64
}

type claudeStructuredReview struct {
	Verdict   *string                    `json:"verdict"`
	Summary   *string                    `json:"summary"`
	Findings  *[]claudeStructuredFinding `json:"findings"`
	NextSteps *[]string                  `json:"next_steps"`
}

type claudeStructuredFinding struct {
	Severity       *string  `json:"severity"`
	Title          *string  `json:"title"`
	Body           *string  `json:"body"`
	File           *string  `json:"file"`
	Line           *int     `json:"line"`
	Confidence     *float64 `json:"confidence"`
	Recommendation *string  `json:"recommendation"`
}

func parseClaudeAuditOutput(data []byte) (ReviewOutput, string, claudeParsedUsage, error) {
	var envelope claudeCLIEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return ReviewOutput{}, "", claudeParsedUsage{}, err
	}
	if envelope.IsError {
		return ReviewOutput{}, "", claudeParsedUsage{}, errors.New("claude CLI reported is_error")
	}
	payload := envelope.StructuredOutput
	if len(payload) == 0 && strings.TrimSpace(envelope.Result) != "" {
		payload = json.RawMessage(envelope.Result)
	}
	if len(payload) == 0 {
		return ReviewOutput{}, "", claudeParsedUsage{}, errors.New("structured_output missing")
	}
	out, err := decodeClaudeStructuredReview(payload)
	if err != nil {
		return ReviewOutput{}, "", claudeParsedUsage{}, err
	}
	models := make([]string, 0, len(envelope.ModelUsage))
	for model := range envelope.ModelUsage {
		models = append(models, model)
	}
	sort.Strings(models)
	if len(models) != 1 {
		return out, "", claudeParsedUsage{}, nil
	}
	model := models[0]
	u := envelope.ModelUsage[model]
	return out, model, claudeParsedUsage{
		InputTokens:       u.InputTokens,
		CachedInputTokens: u.CacheReadInputTokens,
		OutputTokens:      u.OutputTokens,
	}, nil
}

func decodeClaudeStructuredReview(payload []byte) (ReviewOutput, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var raw claudeStructuredReview
	if err := decoder.Decode(&raw); err != nil {
		return ReviewOutput{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return ReviewOutput{}, errors.New("multiple structured review values")
		}
		return ReviewOutput{}, err
	}
	if raw.Verdict == nil || raw.Summary == nil || raw.Findings == nil || raw.NextSteps == nil {
		return ReviewOutput{}, errors.New("structured review is missing a required field")
	}
	if *raw.Verdict != "pass" && *raw.Verdict != "fail" && *raw.Verdict != VerdictInconclusive {
		return ReviewOutput{}, errors.New("invalid verdict")
	}

	findings := make([]Finding, 0, len(*raw.Findings))
	for _, finding := range *raw.Findings {
		if finding.Severity == nil || finding.Title == nil || finding.Body == nil {
			return ReviewOutput{}, errors.New("structured finding is missing a required field")
		}
		if !validClaudeFindingSeverity(*finding.Severity) {
			return ReviewOutput{}, errors.New("invalid finding severity")
		}
		converted := Finding{Severity: *finding.Severity, Title: *finding.Title, Body: *finding.Body}
		if finding.File != nil {
			converted.File = *finding.File
		}
		if finding.Line != nil {
			converted.Line = *finding.Line
		}
		if finding.Confidence != nil {
			converted.Confidence = *finding.Confidence
		}
		if finding.Recommendation != nil {
			converted.Recommendation = *finding.Recommendation
		}
		findings = append(findings, converted)
	}

	return ReviewOutput{
		Verdict:   *raw.Verdict,
		Summary:   *raw.Summary,
		Findings:  findings,
		NextSteps: append([]string(nil), (*raw.NextSteps)...),
	}, nil
}

func validClaudeFindingSeverity(severity string) bool {
	switch severity {
	case "P0", "P1", "P2", "P3":
		return true
	default:
		return false
	}
}

func isClaudeResolvedModel(model string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "claude-")
}

func claudeResolvedModelMatchesRequest(requested, resolved string) bool {
	requested = strings.ToLower(strings.TrimSpace(requested))
	resolved = strings.ToLower(strings.TrimSpace(resolved))
	if requested == "" || !isClaudeResolvedModel(resolved) {
		return false
	}
	if requested == resolved {
		return true
	}
	switch requested {
	case "sonnet", "opus", "haiku":
		return strings.HasPrefix(resolved, "claude-"+requested+"-")
	default:
		return false
	}
}

func normalizeReviewOutput(out ReviewOutput) ReviewOutput {
	if out.Findings == nil {
		out.Findings = []Finding{}
	}
	if out.NextSteps == nil {
		out.NextSteps = []string{}
	}
	return out
}
