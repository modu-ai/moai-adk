package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// gitlabInstanceURLOnDisk reads git_strategy.gitlab.instance_url, "" when absent.
func gitlabInstanceURLOnDisk(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.GitStrategyYAML))
	if err != nil {
		t.Fatalf("read git-strategy.yaml: %v", err)
	}
	var doc struct {
		GitStrategy struct {
			GitLab struct {
				InstanceURL string `yaml:"instance_url"`
			} `yaml:"gitlab"`
		} `yaml:"git_strategy"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse git-strategy.yaml: %v\n%s", err, data)
	}
	return doc.GitStrategy.GitLab.InstanceURL
}

// t1972 F5 (carried from t1966): the reconfigure wizard restores a blank GitLab
// instance URL to its default (https://gitlab.com), but the saver skipped the write
// when the URL was blank, so a result with provider gitlab and a blank URL left
// whatever was on disk. A blank answer means "the default" in both places.
func TestReconfigureBlankGitLabURLRestoresTheDefault(t *testing.T) {
	cases := []struct {
		name     string
		existing string // git-strategy.yaml before; "" = no file
		result   wizard.WizardResult
		want     string
	}{
		{
			// t2108 changed this case: a blank answer keeps a stored URL (it used to
			// overwrite it with the default), and only an absent URL becomes the default
			name:     "gitlab provider, blank URL, a stored self-hosted URL is kept",
			existing: "git_strategy:\n    mode: personal\n    gitlab:\n        instance_url: https://gitlab.corp.example\n",
			result:   wizard.WizardResult{GitMode: "personal", GitProvider: "gitlab"},
			want:     "https://gitlab.corp.example",
		},
		{
			name:   "gitlab provider, blank URL, nothing on disk",
			result: wizard.WizardResult{GitMode: "personal", GitProvider: "gitlab"},
			want:   wizard.DefaultGitLabInstanceURL,
		},
		{
			name:   "gitlab provider, an explicit URL is kept",
			result: wizard.WizardResult{GitMode: "personal", GitProvider: "gitlab", GitLabInstanceURL: "https://gitlab.corp.example"},
			want:   "https://gitlab.corp.example",
		},
		{
			// control: the question is only asked for provider gitlab, so a blank URL
			// under another provider is "not asked", never a default to write
			name:     "github provider, blank URL: nothing is written",
			existing: "git_strategy:\n    mode: personal\n",
			result:   wizard.WizardResult{GitMode: "personal", GitProvider: "github"},
			want:     "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			sections := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
			if err := os.MkdirAll(sections, 0o755); err != nil {
				t.Fatal(err)
			}
			if tc.existing != "" {
				if err := os.WriteFile(filepath.Join(sections, defs.GitStrategyYAML), []byte(tc.existing), 0o644); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(filepath.Join(sections, defs.GitStrategyYAML), []byte("git_strategy: {}\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			res := tc.result
			if err := applyWizardConfig(root, &res); err != nil {
				t.Fatalf("applyWizardConfig: %v", err)
			}
			if got := gitlabInstanceURLOnDisk(t, root); got != tc.want {
				t.Errorf("instance_url on disk = %q, want %q", got, tc.want)
			}
		})
	}
}
