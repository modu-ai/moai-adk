package cli

// Card t1147: every user-owned key init writes into .moai/config/sections must
// survive a forced update. The survey initializes a project with every init
// input that reaches a section template set to a non-default value, then
// compares every leaf key of every section file before and after the update.
//
// Not parallel: the init and update helpers set env vars and chdir.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
)

// flattenSections reads every sections/*.yaml under root into "file:dotted.key" -> value.
func flattenSections(t *testing.T, root string) map[string]string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read sections dir: %v", err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		var doc any
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal %s: %v", e.Name(), err)
		}
		flattenInto(out, e.Name()+":", doc)
	}
	return out
}

func flattenInto(out map[string]string, prefix string, v any) {
	switch m := v.(type) {
	case map[string]any:
		for k, child := range m {
			p := prefix + k
			if !strings.HasSuffix(prefix, ":") {
				p = prefix + "." + k
			}
			flattenInto(out, p, child)
		}
	default:
		out[prefix] = fmt.Sprintf("%v", v)
	}
}

// assertSectionsUnchanged reports every leaf key of before whose value differs
// (or is missing) in the section files under root now.
func assertSectionsUnchanged(t *testing.T, root string, before map[string]string) {
	t.Helper()
	after := flattenSections(t, root)
	var changed []string
	for k, v := range before {
		if a, ok := after[k]; !ok {
			changed = append(changed, fmt.Sprintf("%s: %q -> <missing>", k, v))
		} else if a != v {
			changed = append(changed, fmt.Sprintf("%s: %q -> %q", k, v, a))
		}
	}
	sort.Strings(changed)
	for _, c := range changed {
		t.Errorf("changed by update: %s", c)
	}
}

// initUserOwnedKeysProject runs init with a non-default value for every input
// that reaches a section template: the wizard answers, the profile's output
// languages, and the git flags.
func initUserOwnedKeysProject(t *testing.T) string {
	t.Helper()
	prepareSafeInitHome(t)

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })
	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	wizResult := &wizard.WizardResult{
		ConversationLang:  "ko",
		UserName:          "surveyor",
		DevelopmentMode:   "ddd",
		ModelPolicy:       "low",
		ReportFormat:      "md",
		GitMode:           "team",
		GitProvider:       "gitlab",
		GitHubUsername:    "gh-user",
		GitLabInstanceURL: "https://gitlab.example.com",
		GitLabUsername:    "gl-user",
		AgentWiring:       "claude",
		AutonomyTier:      "full-auto",
		LSPEnabled:        false,
		EnforceQuality:    false,
	}
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) { return wizResult, nil }
	t.Cleanup(func() { runWizardFn = origWizard })

	// Output languages reach init only through the profile preferences.
	if err := profile.WritePreferences(profile.GetCurrentName(), profile.ProfilePreferences{
		GitCommitLang:   "ko",
		CodeCommentLang: "ja",
		DocLang:         "zh",
	}); err != nil {
		t.Fatalf("write preferences: %v", err)
	}

	root := filepath.Join(t.TempDir(), "survey")
	cmd := newInitTestCmd()
	// Git values reach init only through flags (or remote detection).
	for flag, val := range map[string]string{
		"name":                "survey-proj",
		"git-mode":            "team",
		"git-provider":        "gitlab",
		"github-username":     "gh-user",
		"gitlab-instance-url": "https://gitlab.example.com",
	} {
		if err := cmd.Flags().Set(flag, val); err != nil {
			t.Fatalf("set --%s: %v", flag, err)
		}
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, []string{root}); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}

	// Premise: init wrote every carried key with its non-default input, or the
	// survey would compare defaults with defaults and pass vacuously.
	got := flattenSections(t, root)
	for key, want := range map[string]string{
		"language.yaml:language.conversation_language":       "ko",
		"language.yaml:language.git_commit_messages":         "ko",
		"language.yaml:language.code_comments":               "ja",
		"language.yaml:language.documentation":               "zh",
		"quality.yaml:constitution.development_mode":         "ddd",
		"git-strategy.yaml:git_strategy.provider":            "gitlab",
		"git-strategy.yaml:git_strategy.github_username":     "gh-user",
		"git-strategy.yaml:git_strategy.gitlab.instance_url": "https://gitlab.example.com",
	} {
		if got[key] != want {
			t.Fatalf("after init: %s = %q, want %q (the fixture no longer exercises the key)", key, got[key], want)
		}
	}
	return root
}

func TestUpdateForce_UserOwnedKeysSurvive(t *testing.T) {
	root := initUserOwnedKeysProject(t)
	before := flattenSections(t, root)
	for i := 1; i <= 2; i++ {
		if _, err := runForcedTemplateSyncResult(t, root); err != nil {
			t.Fatalf("update #%d: %v", i, err)
		}
	}
	assertSectionsUnchanged(t, root, before)
}

func TestCleanReinstall_UserOwnedKeysSurvive(t *testing.T) {
	root := initUserOwnedKeysProject(t)

	writeTestFile(t, root, ".moai/config/sections/system.yaml", "moai:\n    version: v2.16.1\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired\n")
	// The v2 marker above is the fixture's own edit; the survey covers the rest.
	before := flattenSections(t, root)
	for k := range before {
		if strings.HasPrefix(k, "system.yaml:") {
			delete(before, k)
		}
	}

	var out, errOut bytes.Buffer
	result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              &out,
		ErrOut:           &errOut,
		RunMigrateAgency: (&stubMigrateRunner{}).Run,
	})
	if err != nil {
		t.Fatalf("runCleanReinstall: %v\nout: %s\nerr: %s", err, out.String(), errOut.String())
	}
	if !result.Detected.IsV2 {
		t.Fatalf("fixture was not detected as v2; the reinstall body never ran (details: %v)", result.Detected.SignalDetails)
	}
	runForcedTemplateSyncAt(t, root)
	assertSectionsUnchanged(t, root, before)
}

// TestUpdateForce_UserOwnedKeys_UnrenderableValueFallsBack is the negative
// control for the round-trip rule on a key card t1147 added: a github_username
// the render cannot carry verbatim is not put into the context (the default ""
// renders instead), while a renderable one is. The forced update still keeps
// the hand-edited value, because the snapshot BASE differs from it and the
// merge keeps it as a customization.
func TestUpdateForce_UserOwnedKeys_UnrenderableValueFallsBack(t *testing.T) {
	root := initIdentityProject(t)
	gitStrategy := sectionsFile(root, "git-strategy.yaml")
	original, err := os.ReadFile(gitStrategy)
	if err != nil {
		t.Fatalf("read git-strategy.yaml: %v", err)
	}
	const line = `github_username: ""`
	if !bytes.Contains(original, []byte(line)) {
		t.Fatalf("git-strategy.yaml has no %s line to hand-edit", line)
	}
	setUsername := func(yamlScalar string) {
		t.Helper()
		edited := bytes.Replace(original, []byte(line), []byte("github_username: "+yamlScalar), 1)
		if err := os.WriteFile(gitStrategy, edited, 0o644); err != nil {
			t.Fatalf("write git-strategy.yaml: %v", err)
		}
	}

	// Positive control: a plain value is carried, so the negative case below
	// is not vacuous.
	setUsername(`"gh-plain"`)
	if got := template.NewTemplateContext(loadUpdateUserValues(root)).GitHubUsername; got != "gh-plain" {
		t.Fatalf("renderable github_username carried as %q, want %q", got, "gh-plain")
	}

	// A `"` is no longer unrenderable (card t1162 escapes the value); the
	// renderer's unexpanded-token guard still rejects `$USER`.
	const unrenderable = `gh$USER`
	setUsername(`'gh$USER'`)
	if got := template.NewTemplateContext(loadUpdateUserValues(root)).GitHubUsername; got != "" {
		t.Fatalf("unrenderable github_username carried as %q, want the default \"\"", got)
	}
	runForcedTemplateSyncAt(t, root)
	if got := sectionValue(t, gitStrategy, "git_strategy", "github_username"); got != unrenderable {
		t.Errorf("after update: github_username = %v, want the hand edit %q kept by the merge", got, unrenderable)
	}
}
