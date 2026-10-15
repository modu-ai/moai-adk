// SPEC-UPDATE-REPAIR-RC30-001 / M5
// AC-RC30-14 (U-13, REQ-RC30-022): a same-version run with a different
// template tree does not skip the template sync.
// AC-RC30-15 (U-14, REQ-RC30-023): a stamp that disagrees with the deployed
// bytes produces a visible warning.
// AC-RC30-19 (U-18, REQ-RC30-011): the preview counts the files the run
// deploys for the same fixture.

package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
)

// measuredSyncCmd returns the command the template-sync entry points read their
// flags from. Its output buffer collects every line the sync prints.
func measuredSyncCmd(out *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{Use: "measured-sync"}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("force", false, "")
	return cmd
}

// chdirMeasuredProject moves into root for the rest of the test. The sync entry
// points resolve the project as "." (a relative path), so the fixture must be
// the working directory.
func chdirMeasuredProject(t *testing.T, root string) {
	t.Helper()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir fixture: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
}

// writeMeasuredProjectFiles writes system.yaml with the template_version stamp,
// and a manifest recording treeHash. An empty argument writes nothing for it.
func writeMeasuredProjectFiles(t *testing.T, root, stamp, treeHash string) {
	t.Helper()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if stamp != "" {
		systemYAML := fmt.Sprintf("moai:\n  template_version: %q\n", stamp)
		if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"), []byte(systemYAML), 0o644); err != nil {
			t.Fatalf("write system.yaml: %v", err)
		}
	}
	if treeHash != "" {
		manifestJSON := fmt.Sprintf("{\"version\":\"\",\"deployed_at\":\"\",\"files\":{},\"template_tree_hash\":%q}\n", treeHash)
		if err := os.WriteFile(filepath.Join(root, ".moai", "manifest.json"), []byte(manifestJSON), 0o644); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
	}
}

// TestSkipDecisionUsesTreeHash checks that a version-matched run skips only when
// the recorded template tree equals the embedded tree (REQ-RC30-022).
func TestSkipDecisionUsesTreeHash(t *testing.T) {
	// Uses os.Chdir and the confirmation seam — cannot run in parallel.
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}
	embeddedTree, err := template.ComputeDirTreeHash(embedded, ".")
	if err != nil {
		t.Fatalf("ComputeDirTreeHash: %v", err)
	}

	cases := []struct {
		name     string
		recorded string
		wantSkip bool
	}{
		{name: "same_version_same_tree_skips", recorded: embeddedTree, wantSkip: true},
		{name: "same_version_different_tree_does_not_skip", recorded: "recorded-tree-from-another-build", wantSkip: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeMeasuredProjectFiles(t, root, version.GetVersion(), tc.recorded)
			chdirMeasuredProject(t, root)

			confirmed := false
			prevConfirm := confirmViaPreviewFn
			confirmViaPreviewFn = func(merge.MergeAnalysis, string) (bool, error) {
				confirmed = true
				return false, nil
			}
			t.Cleanup(func() { confirmViaPreviewFn = prevConfirm })

			var buf bytes.Buffer
			if _, syncErr := runTemplateSyncWithProgress(measuredSyncCmd(&buf), false); syncErr != nil {
				t.Fatalf("runTemplateSyncWithProgress: %v", syncErr)
			}
			skipped := strings.Contains(buf.String(), "Up to date")
			if skipped != tc.wantSkip {
				t.Fatalf("skipped = %v, want %v; output:\n%s", skipped, tc.wantSkip, buf.String())
			}
			// Without a skip the run reaches the merge review; the seam cancels it,
			// so nothing is deployed.
			if confirmed == tc.wantSkip {
				t.Fatalf("merge review reached = %v, want %v; output:\n%s", confirmed, !tc.wantSkip, buf.String())
			}
		})
	}
}

// TestStampSkewWarning checks that a template_version stamp disagreeing with the
// deployed system.yaml is reported as a visible warning (REQ-RC30-023).
func TestStampSkewWarning(t *testing.T) {
	t.Parallel()

	t.Run("stale_stamp_on_disk_warns", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeMeasuredProjectFiles(t, root, "0.0.1-stale", "")
		var buf bytes.Buffer
		warnStampSkew(&buf, root, version.GetVersion())
		out := buf.String()
		if !strings.Contains(out, "template stamp") || !strings.Contains(out, "0.0.1-stale") {
			t.Fatalf("a stamp that disagrees with the deployed bytes must warn, got:\n%s", out)
		}
	})

	t.Run("matching_stamp_is_silent", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeMeasuredProjectFiles(t, root, version.GetVersion(), "")
		var buf bytes.Buffer
		warnStampSkew(&buf, root, version.GetVersion())
		if buf.Len() != 0 {
			t.Fatalf("a matching stamp must not warn, got:\n%s", buf.String())
		}
	})

	t.Run("missing_config_is_silent", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		warnStampSkew(&buf, t.TempDir(), version.GetVersion())
		if buf.Len() != 0 {
			t.Fatalf("no deployed system.yaml leaves nothing to compare, got:\n%s", buf.String())
		}
	})
}

// TestPreviewCountEqualsRunCount checks that the preview and the run analyse the
// same fixture with the same deployer, so they report the same file count
// (REQ-RC30-011).
func TestPreviewCountEqualsRunCount(t *testing.T) {
	// Uses os.Chdir and the confirmation seam — cannot run in parallel.
	chdirMeasuredProject(t, t.TempDir())

	var reviewed []int
	prevConfirm := confirmViaPreviewFn
	confirmViaPreviewFn = func(analysis merge.MergeAnalysis, _ string) (bool, error) {
		reviewed = append(reviewed, len(analysis.Files))
		return false, nil
	}
	t.Cleanup(func() { confirmViaPreviewFn = prevConfirm })

	var buf bytes.Buffer
	cmd := measuredSyncCmd(&buf)
	if _, err := runTemplateSyncWithProgress(cmd, false); err != nil {
		t.Fatalf("preview: %v", err)
	}
	if err := runTemplateSyncWithReporter(cmd, nil, false, nil); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(reviewed) != 2 {
		t.Fatalf("merge review reached %d times, want 2 (preview, then run); output:\n%s", len(reviewed), buf.String())
	}
	if reviewed[0] == 0 {
		t.Fatalf("fixture yields no files, so the comparison would be vacuous")
	}
	if reviewed[0] != reviewed[1] {
		t.Fatalf("preview count %d != run count %d for the same fixture", reviewed[0], reviewed[1])
	}
}
