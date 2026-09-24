package cli

// Card t1156: git-strategy.yaml.tmpl renders the GitLab instance URL inside a
// double-quoted YAML scalar without escaping it, so a URL carrying `"` makes
// init write unparseable YAML and one carrying `\` is stored as a different
// string (a YAML escape). validateHTTPSURL is the gate both the
// --gitlab-instance-url flag and the wizard result pass through.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/template"
)

// urlCandidates are well-formed https URLs with a host. The rejected ones
// include every input sync-audit found the template could not carry verbatim
// (a quote, a backslash, invalid UTF-8, C1 controls, NEL, U+2028, a
// noncharacter) and some it can carry but that are not printable (NBSP,
// zero-width space, BOM): the validator is deliberately stricter than the
// template.
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
	{"https://gitlab.example.com/a#b: c", true},
	{"https://gitlab.example.com/%41|@&*!", true},
	{`https://gitlab.example.com/a"b`, false},
	{`https://gitlab.example.com/a\b`, false},
	{`https://gitlab.example.com/a\nb`, false},
	{`https://gitlab.example.com/"`, false},
	{"https://gitlab.example.com/a\xffb", false},
	{"https://gitlab.example.com/a\u0080b", false},
	{"https://gitlab.example.com/a\u009fb", false},
	{"https://gitlab.example.com/a\u0085b", false},
	{"https://gitlab\u0085.example.com", false},
	{"https://gitlab.example.com/a\u2028 b", false},
	{"https://gitlab.example.com/a\ufffeb", false},
	{"https://gitlab.example.com/a\u00a0b", false},
	{"https://gitlab.example.com/a\u200bb", false},
	{"https://gitlab.example.com/\ufeffa", false},
}

func TestValidateHTTPSURL_RejectsValuesTheTemplateCannotCarry(t *testing.T) {
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

// TestValidateHTTPSURL_AcceptedValuesRoundTrip pins the relation to card
// t1147's update predicate (rendersVerbatim) in the one direction t1147
// needs: every candidate init accepts is rendered by git-strategy.yaml.tmpl
// and parsed back verbatim, so within this candidate set no init-accepted URL
// is written altered (BASE == OLD) and later erased by an update. The reverse
// does not hold by design: the validator also rejects some values the
// template could carry (see urlCandidates).
func TestValidateHTTPSURL_AcceptedValuesRoundTrip(t *testing.T) {
	t.Parallel()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	renderer := template.NewRenderer(fsys)
	accepted, failsRoundTrip := 0, 0
	for _, c := range urlCandidates {
		ctx := template.NewTemplateContext(template.WithGitLabInstanceURL(c.url))
		roundTrips := rendersVerbatim(renderer, gitStrategySectionTemplate, ctx, c.url, "git_strategy", "gitlab", "instance_url")
		if !roundTrips {
			failsRoundTrip++
		}
		if validateHTTPSURL(c.url) == nil {
			accepted++
			if !roundTrips {
				t.Errorf("%q: accepted by validateHTTPSURL but does not round-trip through the template", c.url)
			}
		}
	}
	// Positive controls: the set must exercise both sides, or the check above
	// measures nothing.
	if accepted == 0 || failsRoundTrip == 0 {
		t.Fatalf("candidate set is vacuous: %d accepted, %d failing the round-trip", accepted, failsRoundTrip)
	}
}

// TestValidateWizardInput_RejectsQuotedGitLabURL covers the wizard entry point.
func TestValidateWizardInput_RejectsQuotedGitLabURL(t *testing.T) {
	t.Parallel()
	err := validateWizardInput(&wizard.WizardResult{GitLabInstanceURL: `https://gitlab.example.com/a"b`})
	if err == nil || !strings.Contains(err.Error(), "gitlab_instance_url") {
		t.Fatalf("validateWizardInput = %v, want a gitlab_instance_url rejection", err)
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
