package update

// agentic_trunk_test.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M4 (card
// t1831). AC-AGT-013: an operator-set agentic-trunk value survives `moai
// update` reconciling git-strategy.yaml against the template's github-flow
// (the same hazard TestUpdate_GitStrategyValuesSurvive pins for git-flow).

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestAgenticTrunkValueSurvivesUpdate is AC-AGT-013: a project whose manual
// workflow is agentic-trunk, updated against a template whose manual
// workflow is github-flow, keeps agentic-trunk afterward.
func TestAgenticTrunkValueSurvivesUpdate(t *testing.T) {
	const userConfig = `git_strategy:
  worktree_base_branch: main
  manual:
    workflow: agentic-trunk
`
	const templateConfig = `git_strategy:
  worktree_base_branch: ""
  manual:
    workflow: github-flow
`
	const rel = ".moai/config/sections/git-strategy.yaml"
	root := newClassifyFixture(t, map[string]string{rel: userConfig})
	carried := map[string]string{rel: templateConfig}
	tmplFS := recTmplFS(carried)

	base := func(baseRel string) ([]byte, bool) {
		if baseRel == rel {
			return []byte(templateConfig), true
		}
		return nil, false
	}

	summary, pending, err := ReconcileManagedPaths(root, io.Discard, tmplFS, renderWith(carried), nil, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileManagedPaths: %v", err)
	}
	recSimulateDeploy(t, root, carried)
	summary, err = ReconcileMerges(root, io.Discard, renderWith(carried), nil, ReconcileMergeOptions{Base: base}, pending, summary)
	if err != nil {
		t.Fatalf("ReconcileMerges: %v", err)
	}
	_ = summary

	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("git-strategy.yaml did not survive update: %v", err)
	}
	var doc struct {
		GitStrategy struct {
			WorktreeBaseBranch string `yaml:"worktree_base_branch"`
			Manual             struct {
				Workflow string `yaml:"workflow"`
			} `yaml:"manual"`
		} `yaml:"git_strategy"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse post-update git-strategy.yaml: %v", err)
	}
	if doc.GitStrategy.WorktreeBaseBranch != "main" {
		t.Errorf("worktree_base_branch = %q, want main (VALUE reversion)", doc.GitStrategy.WorktreeBaseBranch)
	}
	if doc.GitStrategy.Manual.Workflow != "agentic-trunk" {
		t.Errorf("manual.workflow = %q, want agentic-trunk (VALUE reversion — the update must preserve the operator's value verbatim)", doc.GitStrategy.Manual.Workflow)
	}
}
