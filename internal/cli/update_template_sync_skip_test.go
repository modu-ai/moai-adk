package cli

// update_template_sync_skip_test.go — card t1619, codex review round 3 (P2). Two defects in the
// version-match skip judgement: (1) the judgement loaded the manifest through Load, which moves a
// corrupt manifest aside as a side effect, so the next judgement read the project as fresh and
// skipped the rebuild; (2) a version match whose template tree changed is a sync prompt, and a
// cancel at that prompt read as a version-matched skip, so the retired-key strip ran after it.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// writeSkipStamp writes the project's system.yaml with the template version stamp.
func writeSkipStamp(t *testing.T, root, stamp string) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "moai:\n  template_version: " + stamp + "\n"
	if err := os.WriteFile(filepath.Join(sections, "system.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeSkipManifestTree saves a manifest that records treeHash as its template tree.
func writeSkipManifestTree(t *testing.T, root, treeHash string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load fresh manifest: %v", err)
	}
	mgr.Manifest().TemplateTreeHash = treeHash
	if err := mgr.Save(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
}

// TestVersionMatchSkipKeepsCorruptManifestInPlace: a corrupt manifest must not skip the sync,
// and the judgement must not move the file. Two judgements in a row must agree.
func TestVersionMatchSkipKeepsCorruptManifestInPlace(t *testing.T) {
	root := t.TempDir()
	writeSkipStamp(t, root, version.GetVersion())
	manifestPath := filepath.Join(root, ".moai", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte("{not a manifest"), 0o644); err != nil {
		t.Fatal(err)
	}

	for call := 1; call <= 2; call++ {
		if versionMatchSkipsSync(root) {
			t.Fatalf("judgement %d: a corrupt manifest must not skip the sync; the sync rebuilds it", call)
		}
		if _, err := os.Stat(manifestPath); err != nil {
			t.Fatalf("judgement %d moved the corrupt manifest out of place (%v): a later judgement would read the project as fresh and skip the rebuild", call, err)
		}
	}
}

// TestUpdateSkippedOnVersionMatchRejectsChangedTemplateTree: the predicate that tells a
// version-matched skip from a cancelled merge must include the tree check that decides whether
// the sync prompts at all. A changed tree is a prompt, so its cancel is not a skip.
func TestUpdateSkippedOnVersionMatchRejectsChangedTemplateTree(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "")

	root := t.TempDir()
	writeSkipStamp(t, root, version.GetVersion())
	if got, err := plan.GetProjectConfigVersion(root); err != nil || got != version.GetVersion() {
		t.Fatalf("premise: the project stamp must match this binary (got %q, err %v)", got, err)
	}

	// Positive control: an unchanged tree is a normal version-matched skip.
	current, err := embeddedTemplateTreeHash()
	if err != nil {
		t.Fatalf("embedded tree hash: %v", err)
	}
	writeSkipManifestTree(t, root, current)
	if !updateSkippedOnVersionMatch(cmd, root) {
		t.Fatal("premise: an unchanged template tree on a matching version must read as a skip")
	}

	// The defect: the tree changed, so the sync prompts, and a cancel must not read as a skip.
	writeSkipManifestTree(t, root, "tree-hash-of-an-older-template")
	if updateSkippedOnVersionMatch(cmd, root) {
		t.Fatal("a version-matched project with a changed template tree is a sync prompt: a cancel there must not read as a version-matched skip, or the retired-key strip runs after the cancel")
	}
}
