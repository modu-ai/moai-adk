package feedback

import (
	"strings"
	"testing"
)

// AC-F-012 — a report that reads as a vulnerability disclosure is blocked and
// routed to the private advisory channel.
//
// The three subtests exercise the three ways the verdict can be reached, so a
// failure names which signal broke rather than reporting "classification is
// wrong".
func TestClassifyBlocksVulnerabilityReport(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		title string
		body  string
	}{
		{
			name: "vulnerability vocabulary",
			body: "The template renderer is exploitable: a crafted project name gives remote code execution on init.",
		},
		{
			name: "advisory identifier",
			body: "This looks like the same class of bug as CVE-2024-12345, reproduced on the current main.",
		},
		{
			name:  "key-file mention plus vocabulary",
			title: "hook writes to ~/.ssh/id_rsa",
			body:  "The PreToolUse guard can be walked past, which is a privilege escalation on any shared checkout.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := Scrub(Input{Title: tc.title, Body: tc.body}, testOptions())
			if err != nil {
				t.Fatalf("Scrub returned error: %v", err)
			}
			if res.Verdict != VerdictBlocked {
				t.Fatalf("expected verdict %q, got %q (reason %q)", VerdictBlocked, res.Verdict, res.Reason)
			}
			if !strings.Contains(res.Reason, advisoriesURL) {
				t.Fatalf("reason does not route to the advisories path: %q", res.Reason)
			}
		})
	}
}

// The false-positive control. A classifier that answers "blocked" to everything
// passes every subtest above; these cases are the only thing excluding it.
//
// AC-F-008's benign-prose case asserts the same property through the full
// pipeline. This test narrows the observation to the classifier itself and adds
// the near-miss inputs — one signal on its own must not be enough.
func TestClassifyDoesNotBlockOrdinaryReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		title string
		body  string
	}{
		{
			name:  "ordinary bug report",
			title: "moai init shows the wizard twice",
			body:  "Running moai init on a fresh directory prints the language question, then repeats it after the summary.",
		},
		{
			name: "single path mention",
			body: "The cleanup sweep also removes .git/config, which was not expected.",
		},
		{
			name: "single vocabulary term",
			body: "The docs example has a security issue, but the CLI itself behaves the way the guide describes.",
		},
		{
			name: "prose about a token prefix",
			body: "the " + ghTokenPrefix + " prefix is how GitHub tokens start, but the docs never say so.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			verdict, reason := classify(Input{Title: tc.title, Body: tc.body}, testOptions())
			if verdict != VerdictOK {
				t.Fatalf("expected verdict %q, got %q (reason %q)", VerdictOK, verdict, reason)
			}
			if reason != "" {
				t.Fatalf("expected an empty reason on an allowed report, got %q", reason)
			}
		})
	}
}

// AC-F-013 — the classifier reads the RAW input, before masking.
//
// The second assertion is the order guard: masking removes exactly the signal
// the classifier used, so a pipeline that classifies the masked text returns
// "ok" for the same report. That inversion is the silent false negative the
// design pins down, and reversing the two stages is the only way to reach it.
func TestClassifyReadsPreMaskBody(t *testing.T) {
	t.Parallel()

	// A credential is this body's only classification signal: no advisory
	// identifier, no key-file path, no vulnerability vocabulary.
	body := "while running the wizard the log printed " + fakeGitHubToken() + " and then exited"

	res, err := Scrub(Input{Body: body}, testOptions())
	if err != nil {
		t.Fatalf("Scrub returned error: %v", err)
	}
	if res.Verdict != VerdictBlocked {
		t.Fatalf("expected verdict %q on the raw body, got %q", VerdictBlocked, res.Verdict)
	}
	if res.Body == body {
		t.Fatalf("the body was not masked, so the second assertion observes nothing: %q", res.Body)
	}

	maskedVerdict, _ := classify(Input{Body: res.Body}, testOptions())
	if maskedVerdict != VerdictOK {
		t.Fatalf("the masked body still carries the signal, so this test cannot detect an order inversion (got %q)", maskedVerdict)
	}
}
