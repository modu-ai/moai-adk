package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

func TestWriteSectionRefusesLinkedDirectoryBeforeCreatingParents(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	link := filepath.Join(root, ".moai")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("create linked project config: %v", err)
	}
	err := WriteSectionViaSeam(root, "workflow", []yamlpatch.KeyEdit{{Path: []string{"workflow", "team_mode"}, Value: "new"}})
	if err == nil || !strings.Contains(err.Error(), "yamlpatch: stat") {
		t.Errorf("linked project section = %v; want stat refusal", err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("outside directory changed: %v, %v", entries, err)
	}
	if got, err := os.Readlink(link); err != nil || got != outside {
		t.Fatalf("project config link changed: %q, %v", got, err)
	}
}

func TestWriteSectionAcceptsTrustedProjectRootAlias(t *testing.T) {
	root := t.TempDir()
	alias := filepath.Join(t.TempDir(), "project-alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatalf("create project root alias: %v", err)
	}
	err := WriteSectionViaSeam(alias, "workflow", []yamlpatch.KeyEdit{{Path: []string{"workflow", "team_mode"}, Value: "new"}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".moai", "config", "sections", "workflow.yaml")
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "workflow:\n    team_mode: new\n" {
		t.Fatalf("canonical project section = %q, %v", got, err)
	}
	aliasInfo, err := os.Stat(filepath.Join(alias, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	rootInfo, err := os.Stat(path)
	if err != nil || !os.SameFile(aliasInfo, rootInfo) {
		t.Fatalf("alias and canonical section do not share identity: %v", err)
	}
}
