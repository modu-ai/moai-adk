package cli

import (
	"encoding/json"
	"testing"
)

func TestExtractJSONObject(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"bare", `{"a":1}`, `{"a":1}`},
		{"fencedJSON", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"fencedBare", "```\n{\"a\":1}\n```", `{"a":1}`},
		{"proseWrapped", "Here is the review:\n{\"a\":1}\nDone.", `{"a":1}`},
		{"fencedWithProse", "```json\nResult:\n{\"a\":1}\n```", `{"a":1}`},
		{"noBrace", "not json at all", "not json at all"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractJSONObject(c.in); got != c.want {
				t.Errorf("extractJSONObject(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestParseGLMReview_StripsMarkdownFence reproduces the z.ai failure mode where
// the ReviewOutput JSON is wrapped in a ```json code fence: parseGLMReview MUST
// recover the verdict rather than fail-open to inconclusive. Regression guard
// for the "invalid character '`'" parse error observed in production.
func TestParseGLMReview_StripsMarkdownFence(t *testing.T) {
	reviewJSON := `{"verdict":"pass","summary":"fenced ok"}`
	fenced := "```json\n" + reviewJSON + "\n```"
	envBytes, err := json.Marshal(struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}{Content: []struct {
		Text string `json:"text"`
	}{{Text: fenced}}})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	out := parseGLMReview(envBytes)
	if out.Verdict != "pass" {
		t.Fatalf("parseGLMReview verdict = %q (summary=%q); want pass — markdown fence not stripped",
			out.Verdict, out.Summary)
	}
}
