package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestDeployer_DeploysToProjectRoot tests that all files deploy to project .claude/
func TestDeployer_DeploysToProjectRoot(t *testing.T) {
	root, mgr := setupDeployProject(t)

	fs := fstest.MapFS{
		".claude/agents/moai/expert-backend.md": &fstest.MapFile{
			Data: []byte("# Expert Backend Agent"),
		},
		".claude/skills/moai/SKILL.md": &fstest.MapFile{
			Data: []byte("# MoAI Core Skill"),
		},
	}

	// Create deployer with local mode context
	tmplCtx := NewTemplateContext(
		WithProject("testproject", root),
	)
	d := NewDeployer(fs)

	err := d.Deploy(context.Background(), root, mgr, tmplCtx)
	if err != nil {
		t.Fatalf("Deploy error: %v", err)
	}

	// Files should be in project .claude/
	expectedFiles := []string{
		".claude/agents/moai/expert-backend.md",
		".claude/skills/moai/SKILL.md",
	}
	for _, f := range expectedFiles {
		absPath := filepath.Join(root, f)
		if _, err := os.Stat(absPath); err != nil {
			t.Errorf("expected file %q in project: %v", f, err)
		}
	}
}
