package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// t2108 (found by the t1972 audit): the reconfigure wizard did not prefill the
// stored self-hosted GitLab instance URL, so its question default was the gitlab.com
// default, and pressing Enter (a blank answer, which the wizard turns into the
// question default) overwrote the stored URL. The stored value must be the default
// and a blank answer must keep it.

const selfHostedGitLab = "https://gitlab.corp.example"

func writeGitStrategy(t *testing.T, root, body string) {
	t.Helper()
	sections := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, defs.GitStrategyYAML), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReconfigurePrefillsTheStoredGitLabInstanceURL(t *testing.T) {
	cases := []struct {
		name string
		body string // git-strategy.yaml; "" = no file
		want string // the question default after the prefill
	}{
		{"a stored self-hosted URL is the default", "git_strategy:\n    mode: personal\n    gitlab:\n        instance_url: " + selfHostedGitLab + "\n", selfHostedGitLab},
		{"no stored URL: the question keeps the gitlab.com default", "git_strategy:\n    mode: personal\n", wizard.DefaultGitLabInstanceURL},
		{"no file at all: the question keeps the gitlab.com default", "", wizard.DefaultGitLabInstanceURL},
		{"a blank stored URL: the question keeps the gitlab.com default", "git_strategy:\n    gitlab:\n        instance_url: \"\"\n", wizard.DefaultGitLabInstanceURL},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.body != "" {
				writeGitStrategy(t, root, tc.body)
			}
			questions := wizard.ReconfigureQuestions(root)
			prefillReconfigureDefaults(questions, root)
			q := wizard.QuestionByID(questions, "gitlab_instance_url")
			if q == nil {
				t.Fatal("the reconfigure set has no gitlab_instance_url question")
			}
			if q.Default != tc.want {
				t.Errorf("gitlab_instance_url default = %q, want %q", q.Default, tc.want)
			}
		})
	}
}

// A blank answer reaches the saver as an empty URL (a result built without the
// wizard's own default restore). For a gitlab provider with a stored URL the stored
// value stays; with none stored the default is written (t1972 F5).
func TestReconfigureBlankGitLabURLKeepsTheStoredValue(t *testing.T) {
	root := t.TempDir()
	writeGitStrategy(t, root, "git_strategy:\n    mode: personal\n    gitlab:\n        instance_url: "+selfHostedGitLab+"\n")
	res := wizard.WizardResult{GitMode: "personal", GitProvider: "gitlab"}
	if err := applyWizardConfig(root, &res); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}
	if got := gitlabInstanceURLOnDisk(t, root); got != selfHostedGitLab {
		t.Errorf("instance_url on disk = %q, want the stored %q kept", got, selfHostedGitLab)
	}
}
