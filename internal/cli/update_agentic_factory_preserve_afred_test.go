//go:build afred

// update_agentic_factory_preserve_afred_test.go: OBSERVATION of the real
// template-sync update cycle (runTemplateSyncWithReporter) over a t.TempDir()
// project that holds user-owned Agentic Factory issue templates. Excluded from
// the default run by the afred build tag; run it with:
//
//	go test -tags afred -run '^TestAgenticFactoryUpdateCyclePreservesUserTemplates$' ./internal/cli/
//
// KNOWN DEFECT REPRODUCTION, BEFORE ANY REPAIR (RED). Measured on this tree:
// the update cycle overwrites both user-owned templates
//
//	.github/ISSUE_TEMPLATE/agentic-factory-01-parent-issue.yml  (user_modified)
//	.github/ISSUE_TEMPLATE/agentic-factory-02-child-issue.yml   (user_created)
//
// and leaves the user-only template untouched. This is the same defect the
// Deployer-level reproduction in internal/template records. The update code
// path is also reworked by another open PR, so re-run this after it lands.
//
// Sentinel on failure: AF_UPDATE_CYCLE.
package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/spf13/cobra"
)

// TestAgenticFactoryUpdateCyclePreservesUserTemplates: after the update cycle
// a user-modified and a user-created Agentic Factory issue template, plus a
// user-only template beside them, are byte-identical. The changed paths are
// listed in the failure.
func TestAgenticFactoryUpdateCyclePreservesUserTemplates(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"system:\n  template_version: \"0.0.0\"\n")

	edited := ".github/ISSUE_TEMPLATE/agentic-factory-01-parent-issue.yml"
	created := ".github/ISSUE_TEMPLATE/agentic-factory-02-child-issue.yml"
	userOnly := ".github/ISSUE_TEMPLATE/my-team-bug-report.yml"
	content := map[string]string{
		edited:   "name: edited by the user\n",
		created:  "name: written by the user before install\n",
		userOnly: "name: my team bug report\n",
	}
	for rel, c := range content {
		writeTestFile(t, root, rel, c)
	}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("AF_UPDATE_CYCLE: manifest load: %v", err)
	}
	if err := mgr.Track(edited, manifest.UserModified, manifest.HashBytes([]byte("name: shipped\n"))); err != nil {
		t.Fatalf("AF_UPDATE_CYCLE: track: %v", err)
	}
	if err := mgr.Track(created, manifest.UserCreated, manifest.HashBytes([]byte(content[created]))); err != nil {
		t.Fatalf("AF_UPDATE_CYCLE: track: %v", err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("AF_UPDATE_CYCLE: manifest save: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	runCycle := func(dir string) {
		t.Helper()
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("chdir to fixture: %v", err)
		}
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("force", false, "")
		cmd.Flags().Bool("yes", true, "")
		cmd.Flags().Bool("config", false, "")
		_ = cmd.Flags().Set("yes", "true")
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetContext(context.Background())
		if err := runTemplateSyncWithReporter(cmd, nil, true, nil); err != nil {
			t.Fatalf("AF_UPDATE_CYCLE: runTemplateSyncWithReporter: %v\noutput: %s", err, buf.String())
		}
	}

	// positive control: with no user file at those paths, the same cycle does
	// deploy them, so a build that ships no Agentic Factory templates cannot
	// pass this test
	control := t.TempDir()
	writeTestFile(t, control, ".moai/config/sections/system.yaml",
		"system:\n  template_version: \"0.0.0\"\n")
	runCycle(control)
	for _, rel := range []string{edited, created} {
		if _, err := os.Stat(filepath.Join(control, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("AF_UPDATE_CYCLE: positive control failed, %s was not deployed to a project without it: %v", rel, err)
		}
	}

	runCycle(root)

	var changed []string
	for rel, want := range content {
		got, rerr := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		switch {
		case rerr != nil:
			changed = append(changed, rel+" (removed)")
		case string(got) != want:
			changed = append(changed, rel+" (overwritten)")
		}
	}
	sort.Strings(changed)
	if len(changed) > 0 {
		t.Errorf("AF_UPDATE_CYCLE: the update cycle changed or removed user files: %s", strings.Join(changed, ", "))
	}
}
