package cli

import "testing"

func TestParseClaudeAuditOutput_ResultStringAndFindingMetadata_AC_CLA_007(t *testing.T) {
	out, model, usage, err := parseClaudeAuditOutput([]byte(`{
  "is_error": false,
  "result": "{\"verdict\":\"fail\",\"summary\":\"one issue\",\"findings\":[{\"severity\":\"P2\",\"title\":\"title\",\"body\":\"body\",\"file\":\"code.go\",\"line\":7,\"confidence\":0.9,\"recommendation\":\"fix it\"}],\"next_steps\":[\"test\"]}",
  "modelUsage": {"claude-sonnet-4-6": {"inputTokens":11,"cacheReadInputTokens":2,"outputTokens":5}}
}`))
	if err != nil {
		t.Fatalf("parse result-string output: %v", err)
	}
	if out.Verdict != "fail" || out.Summary != "one issue" || len(out.Findings) != 1 || len(out.NextSteps) != 1 {
		t.Fatalf("parsed review = %+v", out)
	}
	finding := out.Findings[0]
	if finding.File != "code.go" || finding.Line != 7 || finding.Confidence != 0.9 || finding.Recommendation != "fix it" {
		t.Fatalf("finding metadata = %+v", finding)
	}
	if model != "claude-sonnet-4-6" || usage.InputTokens == nil || *usage.InputTokens != 11 || usage.CachedInputTokens == nil || *usage.CachedInputTokens != 2 || usage.OutputTokens == nil || *usage.OutputTokens != 5 {
		t.Fatalf("model/usage = %q / %+v", model, usage)
	}
}

func TestParseClaudeAuditOutput_AbsentUsageCountsRemainNil_AC_CLA_007(t *testing.T) {
	out, model, usage, err := parseClaudeAuditOutput([]byte(`{
  "is_error": false,
  "structured_output": {"verdict":"pass","summary":"ok","findings":[],"next_steps":[]},
  "modelUsage": {"claude-sonnet-4-6": {}}
}`))
	if err != nil {
		t.Fatalf("parse output: %v", err)
	}
	if out.Verdict != "pass" || model != "claude-sonnet-4-6" {
		t.Fatalf("parsed output/model = %+v / %q", out, model)
	}
	if usage.InputTokens != nil || usage.CachedInputTokens != nil || usage.OutputTokens != nil {
		t.Fatalf("absent usage counts = %+v, want nil pointers", usage)
	}
}

func TestParseClaudeAuditOutput_RejectsSchemaViolations_AC_CLA_007(t *testing.T) {
	tests := map[string]string{
		"missing required summary": `{
  "is_error": false,
  "structured_output": {"verdict":"pass","findings":[],"next_steps":[]},
  "modelUsage": {"claude-sonnet-4-6": {}}
}`,
		"unknown structured field": `{
  "is_error": false,
  "structured_output": {"verdict":"pass","summary":"ok","findings":[],"next_steps":[],"unexpected":true},
  "modelUsage": {"claude-sonnet-4-6": {}}
}`,
		"finding missing required body": `{
  "is_error": false,
  "structured_output": {"verdict":"fail","summary":"issue","findings":[{"severity":"P1","title":"missing body"}],"next_steps":[]},
  "modelUsage": {"claude-sonnet-4-6": {}}
}`,
		"finding severity outside enum": `{
  "is_error": false,
  "structured_output": {"verdict":"fail","summary":"issue","findings":[{"severity":"critical","title":"bad enum","body":"details"}],"next_steps":[]},
  "modelUsage": {"claude-sonnet-4-6": {}}
}`,
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, _, err := parseClaudeAuditOutput([]byte(payload)); err == nil {
				t.Fatal("schema-invalid structured output was accepted")
			}
		})
	}
}
