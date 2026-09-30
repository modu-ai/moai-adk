package config

// template_main_commit_ban_keys_test.go — the D8 instrument
// (SPEC-MAIN-COMMIT-BAN-001 plan §A D8).
//
// The existing struct-YAML symmetry harness cannot see the two new keys: its
// symmetryCases cover seven sections (workflow and git-strategy are absent)
// and checkSymmetry walks one level deep, both by design. Deepening that
// harness would sweep every section's nested keys into enforcement at once —
// a blast radius far beyond this SPEC. This dedicated presence test enforces
// exactly the two new keys on all four surfaces:
//
//   - TEMPLATE workflow.yaml  → workflow.branch_guard.deny_commits_on present
//     (a `moai update` re-application dropping the template key goes red)
//   - TEMPLATE git-strategy.yaml.tmpl → git_strategy.manual.lead_push_threshold
//     present
//   - LOCAL (tracked) workflow.yaml → deny_commits_on == [main]
//   - LOCAL (tracked) git-strategy.yaml → lead_push_threshold == 20
//
// Template files are asserted by anchored text presence (the .tmpl carries Go
// template placeholders, so a strict parse would over-constrain); local files
// are parsed as YAML and asserted by value.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// t1337RepoRoot returns the repository root from this test file's location,
// the same idiom audit_loader_completeness_test.go uses.
func t1337RepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func t1337ReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// TestTemplateConfigCarriesMainCommitBanKeys closes the F4 gap (plan D8):
// template-key presence + the operator-given local values, one test.
func TestTemplateConfigCarriesMainCommitBanKeys(t *testing.T) {
	repoRoot := t1337RepoRoot(t)

	t.Run("template_workflow_yaml_has_branch_guard_block", func(t *testing.T) {
		body := t1337ReadFile(t, filepath.Join(repoRoot,
			"internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml"))
		lines := strings.Split(body, "\n")
		keyLine := -1
		for i, line := range lines {
			// The KEY line, not a comment mentioning it: 4-space indent, no #.
			if strings.HasPrefix(line, "    branch_guard:") {
				keyLine = i
				break
			}
		}
		if keyLine < 0 {
			t.Fatalf("template workflow.yaml carries no branch_guard: key — the commit-ban key surface is absent")
		}
		tail := strings.Join(lines[keyLine:], "\n")
		if !strings.Contains(tail, "deny_commits_on:") {
			t.Fatalf("template branch_guard block lacks deny_commits_on:")
		}
		if !strings.Contains(tail, "enabled: false") {
			t.Fatalf("template branch_guard block must ship enabled: false (template neutrality)")
		}
		if strings.Contains(body, "[main]") {
			t.Fatalf("template workflow.yaml must NOT pre-select a protected branch (neutrality)")
		}
	})

	t.Run("template_git_strategy_yaml_tmpl_has_lead_push_threshold", func(t *testing.T) {
		body := t1337ReadFile(t, filepath.Join(repoRoot,
			"internal", "template", "templates", ".moai", "config", "sections", "git-strategy.yaml.tmpl"))
		manualIdx := strings.Index(body, "  manual:")
		if manualIdx < 0 {
			t.Fatalf("template git-strategy.yaml.tmpl has no manual: block")
		}
		// Scope the assertion to the manual block (up to the personal: block).
		manual := body[manualIdx:]
		if end := strings.Index(manual, "\n  personal:"); end >= 0 {
			manual = manual[:end]
		}
		if !strings.Contains(manual, "lead_push_threshold: 0") {
			t.Fatalf("template git-strategy manual block lacks lead_push_threshold: 0 (the disabled default)\n%s", manual)
		}
	})

	t.Run("local_workflow_deny_commits_on_main", func(t *testing.T) {
		body := t1337ReadFile(t, filepath.Join(repoRoot, ".moai", "config", "sections", "workflow.yaml"))
		var doc map[string]any
		if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
			t.Fatalf("parse local workflow.yaml: %v", err)
		}
		wf, ok := doc["workflow"].(map[string]any)
		if !ok {
			t.Fatalf("local workflow.yaml has no workflow: mapping")
		}
		bg, ok := wf["branch_guard"].(map[string]any)
		if !ok {
			t.Fatalf("local workflow.branch_guard missing — deny_commits_on: [main] not applied")
		}
		list, ok := bg["deny_commits_on"].([]any)
		if !ok {
			t.Fatalf("local workflow.branch_guard.deny_commits_on missing or not a list: %v", bg["deny_commits_on"])
		}
		found := false
		for _, v := range list {
			if s, ok := v.(string); ok && s == "main" {
				found = true
			}
		}
		if !found {
			t.Fatalf("local deny_commits_on = %v, want it to contain main", list)
		}
	})

	t.Run("local_git_strategy_lead_push_threshold_20", func(t *testing.T) {
		body := t1337ReadFile(t, filepath.Join(repoRoot, ".moai", "config", "sections", "git-strategy.yaml"))
		var doc map[string]any
		if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
			t.Fatalf("parse local git-strategy.yaml: %v", err)
		}
		gs, ok := doc["git_strategy"].(map[string]any)
		if !ok {
			t.Fatalf("local git-strategy.yaml has no git_strategy: mapping")
		}
		manual, ok := gs["manual"].(map[string]any)
		if !ok {
			t.Fatalf("local git_strategy has no manual: mapping")
		}
		threshold, ok := manual["lead_push_threshold"].(int)
		if !ok {
			t.Fatalf("local git_strategy.manual.lead_push_threshold missing or not an int: %v", manual["lead_push_threshold"])
		}
		if threshold != 20 {
			t.Fatalf("local lead_push_threshold = %d, want 20 (the operator-given initial value)", threshold)
		}
	})
}
