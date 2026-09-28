package kickoff

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/jev"
)

// TestParseJudgementShapes covers the strict judgement decoder: every
// malformed shape is a usage error, a well-formed one decodes.
func TestParseJudgementShapes(t *testing.T) {
	const good = `{"agent":"lead","model":"m","answer":"approve","confidence":0.9,"reason":"r","reason_refs":["contract.yaml:3"]}`
	if _, err := parseJudgement([]byte(good), 10); err != nil {
		t.Fatalf("well-formed judgement rejected: %v", err)
	}
	bad := map[string]string{
		"not-json":         `{`,
		"unknown-field":    `{"agent":"lead","answer":"approve","confidence":0.9,"reason":"r","reason_refs":["contract.yaml:3"],"extra":1}`,
		"trailing-data":    good + ` {}`,
		"empty-agent":      `{"agent":" ","answer":"approve","confidence":0.9,"reason":"r","reason_refs":["contract.yaml:3"]}`,
		"unknown-answer":   `{"agent":"lead","answer":"maybe","confidence":0.9,"reason":"r","reason_refs":["contract.yaml:3"]}`,
		"no-confidence":    `{"agent":"lead","answer":"approve","reason":"r","reason_refs":["contract.yaml:3"]}`,
		"confidence-high":  `{"agent":"lead","answer":"approve","confidence":1.5,"reason":"r","reason_refs":["contract.yaml:3"]}`,
		"confidence-low":   `{"agent":"lead","answer":"approve","confidence":-0.1,"reason":"r","reason_refs":["contract.yaml:3"]}`,
		"empty-reason":     `{"agent":"lead","answer":"approve","confidence":0.9,"reason":"","reason_refs":["contract.yaml:3"]}`,
		"no-refs":          `{"agent":"lead","answer":"approve","confidence":0.9,"reason":"r","reason_refs":[]}`,
		"ref-shape":        `{"agent":"lead","answer":"approve","confidence":0.9,"reason":"r","reason_refs":["spec.md:3"]}`,
		"ref-past-the-end": `{"agent":"lead","answer":"approve","confidence":0.9,"reason":"r","reason_refs":["contract.yaml:11"]}`,
	}
	for name, raw := range bad {
		t.Run(name, func(t *testing.T) {
			if _, err := parseJudgement([]byte(raw), 10); !errors.Is(err, ErrUsage) {
				t.Errorf("err = %v, want ErrUsage", err)
			}
		})
	}
}

// TestSpecTier covers the tier reader's default and its quoted and unknown
// values.
func TestSpecTier(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"absent-file", "", "L"},
		{"no-tier", "---\nid: X\n---\n", "L"},
		{"plain", "---\ntier: S\n---\n", "S"},
		{"quoted", "---\ntier: \"M\"\n---\n", "M"},
		{"unknown", "---\ntier: XL\n---\n", "L"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if c.body != "" {
				if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(c.body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := specTier(dir); got != c.want {
				t.Errorf("specTier = %q, want %q", got, c.want)
			}
		})
	}
}

// TestJevVerdictMapping covers every availability and answer shape the Jev
// verdict mapper distinguishes.
func TestJevVerdictMapping(t *testing.T) {
	on := DecideInput{Config: Config{JevEnabled: true, JevMinConfidence: 0.5}}
	start := func(choice string, p float64) jev.Result {
		return jev.Result{Availability: jev.Available, Answers: []jev.Answer{{QuestionID: "start", Choice: choice, Probability: p}}}
	}
	cases := []struct {
		name       string
		in         DecideInput
		r          jev.Result
		wantChoice string
		wantReason string
	}{
		{"config-off", DecideInput{}, start("approve", 0.9), "", "jev_disabled"},
		{"disabled", on, jev.Result{Availability: jev.Disabled}, "", "jev_disabled"},
		{"no-credential", on, jev.Result{Availability: jev.NoCredential}, "", "jev_key_missing"},
		{"unauthorized", on, jev.Result{Availability: jev.Unauthorized}, "", "jev_key_missing"},
		{"malformed", on, jev.Result{Availability: jev.Malformed}, "", "jev_malformed_response"},
		{"unreachable", on, jev.Result{Availability: jev.Unreachable}, "", "jev_call_failed"},
		{"no-start-answer", on, jev.Result{Availability: jev.Available, Answers: []jev.Answer{{QuestionID: "weakest", Choice: "a", Probability: 0.9}}}, "", "jev_malformed_response"},
		{"choice-outside-set", on, start("maybe", 0.9), "", "jev_malformed_response"},
		{"probability-out-of-range", on, start("approve", 1.2), "", "jev_malformed_response"},
		{"low-confidence", on, start("approve", 0.4), "", "jev_low_confidence"},
		{"answered", on, start("reject", 0.8), "reject", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, reason := jevVerdict(c.in, c.r)
			if reason != c.wantReason || a.Choice != c.wantChoice {
				t.Errorf("got choice %q reason %q, want %q %q", a.Choice, reason, c.wantChoice, c.wantReason)
			}
		})
	}
}

// TestDefaultJevDisabledMakesNoCall pins the production seam's gate: a
// disabled capability answers Disabled without constructing a request.
func TestDefaultJevDisabledMakesNoCall(t *testing.T) {
	r := defaultJev(false)(context.Background(), jev.Request{State: "s"})
	if r.Availability != jev.Disabled {
		t.Errorf("availability = %q, want %q", r.Availability, jev.Disabled)
	}
}
