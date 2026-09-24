package cli

// Card t1156: git-strategy.yaml.tmpl renders the GitLab instance URL inside a
// double-quoted YAML scalar without escaping it, so a URL carrying `"` makes
// init write unparseable YAML and one carrying `\` is stored as a different
// string (a YAML escape). validateHTTPSURL is the gate both the
// --gitlab-instance-url flag and the wizard result pass through.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// urlCandidates are well-formed https URLs with a host; only the quote and the
// backslash cannot be carried verbatim by the template's quoted scalar.
var urlCandidates = []struct {
	url      string
	accepted bool
}{
	{"https://gitlab.example.com", true},
	{"https://gitlab.example.com/group/sub-group", true},
	{"https://gitlab.example.com:8443/a?b=c&d=e#f", true},
	{"https://gitlab.example.com/it's", true},
	{"https://gitlab.example.com/a b", true},
	{"https://gitlab.example.com/한글", true},
	{`https://gitlab.example.com/a"b`, false},
	{`https://gitlab.example.com/a\b`, false},
	{`https://gitlab.example.com/a\nb`, false},
	{`https://gitlab.example.com/"`, false},
}

func TestValidateHTTPSURL_RejectsQuoteAndBackslash(t *testing.T) {
	t.Parallel()
	for _, c := range urlCandidates {
		err := validateHTTPSURL(c.url)
		if c.accepted && err != nil {
			t.Errorf("validateHTTPSURL(%q) = %v, want accepted", c.url, err)
		}
		if !c.accepted && err == nil {
			t.Errorf("validateHTTPSURL(%q) accepted, want rejected", c.url)
		}
	}
}

// TestValidateHTTPSURL_AcceptsExactlyWhatTheTemplateRoundTrips pins the
// relation to card t1147's update predicate (rendersVerbatim): for every
// candidate, init accepts the URL iff git-strategy.yaml.tmpl renders it and
// parses it back verbatim. Anything init accepts is therefore also carried by
// the update render, so no init-accepted URL can be written altered (BASE ==
// OLD) and later erased by an update.
func TestValidateHTTPSURL_AcceptsExactlyWhatTheTemplateRoundTrips(t *testing.T) {
	t.Parallel()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	renderer := template.NewRenderer(fsys)
	sawRejected := false
	for _, c := range urlCandidates {
		ctx := template.NewTemplateContext(template.WithGitLabInstanceURL(c.url))
		roundTrips := rendersVerbatim(renderer, gitStrategySectionTemplate, ctx, c.url, "git_strategy", "gitlab", "instance_url")
		accepted := validateHTTPSURL(c.url) == nil
		if accepted != roundTrips {
			t.Errorf("%q: validateHTTPSURL accepted=%v, template round-trip=%v", c.url, accepted, roundTrips)
		}
		if !roundTrips {
			sawRejected = true
		}
	}
	// Positive control: the candidate set must contain a URL the template
	// cannot carry, or the equivalence above is vacuous.
	if !sawRejected {
		t.Fatal("no candidate failed the template round-trip; the equivalence check measures nothing")
	}
}

// TestValidateInitFlags_RejectsQuotedGitLabURL covers the flag entry point.
func TestValidateInitFlags_RejectsQuotedGitLabURL(t *testing.T) {
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("gitlab-instance-url", `https://gitlab.example.com/a"b`); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	err := validateInitFlags(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--gitlab-instance-url") {
		t.Fatalf("validateInitFlags = %v, want a --gitlab-instance-url rejection", err)
	}
}
