package backup

// optout_survival_test.go — card t771 control group.
//
// The card's claim: CleanMoaiManagedPaths wipes .moai/config wholesale, so a
// user's statusline opt-out (segments.github: false) evaporates on one
// `moai update` and the gh polling it disabled silently comes back.
//
// The card also states the [HARD] condition under which that claim may be
// asserted: measure a tree that carries the opt-out AND a tree that does not,
// across the real backup -> wipe -> redeploy -> restore sequence. Reading the
// wipe in isolation is not evidence, because the wipe is followed by a restore
// that 3-way merges the backed-up sections back over the freshly deployed
// template.
//
// These tests drive the production entry points (BackupMoaiConfig,
// RestoreMoaiConfig) rather than a reimplementation, so what they report is the
// shipped behaviour and not a model of it. They deliberately do NOT touch the
// real home: every path is under t.TempDir().
//
// Note on the key under test: the template's statusline.yaml segments map does
// NOT carry a `github` key at all (the segment defaults on in code). A user's
// `github: false` is therefore an OLD-ONLY key at merge time, which is a
// different merge branch from a user-edited shared key — and it is the branch
// the card's scenario actually exercises.

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/template"
)

// statuslineTemplatePath is the template-relative path of the file under test,
// the same spelling the deploy walk and the merge base both key on.
const statuslineTemplatePath = ".moai/config/sections/statusline.yaml"

// templateStatuslineYAML mirrors the shape the template ships: segment toggles
// with no `github` key.
const templateStatuslineYAML = `statusline:
    theme: "catppuccin-mocha"
    segments:
        model: true
        context: true
        pr: true
`

// userStatuslineYAMLOptedOut is the same file after a user disables the github
// segment — the opt-out whose survival the card is about.
const userStatuslineYAMLOptedOut = `statusline:
    theme: "catppuccin-mocha"
    segments:
        model: true
        context: true
        pr: true
        github: false
`

// seedProject writes a project tree whose .moai/config/sections/statusline.yaml
// carries the given content, plus the .template-defaults the backup step stages
// for the 3-way merge base.
func seedProject(t *testing.T, userYAML string) string {
	t.Helper()

	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "statusline.yaml"),
		[]byte(userYAML), 0o644); err != nil {
		t.Fatalf("write user statusline.yaml: %v", err)
	}
	return root
}

// wipeAndRedeploy reproduces what the update run does between backup and
// restore, using the SHIPPED wipe rather than a stand-in: the real
// deploy.CleanMoaiManagedPaths removes .moai/config, then the template's own
// statusline.yaml is laid down from the embedded FS exactly as the deploy walk
// does for a file that carries no .tmpl suffix.
//
// Driving the production wipe is what makes this a measurement of the shipped
// sequence rather than a model of it: the card names CleanMoaiManagedPaths by
// function, so a hand-rolled RemoveAll would leave the claim untested.
func wipeAndRedeploy(t *testing.T, root string) {
	t.Helper()

	tmplFS, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates: %v", err)
	}

	if err := deploy.CleanMoaiManagedPaths(root, io.Discard, tmplFS); err != nil {
		t.Fatalf("CleanMoaiManagedPaths: %v", err)
	}

	// The wipe must actually have removed the file the restore will merge
	// back; without this the arms below could pass because nothing happened.
	target := filepath.Join(root, ".moai", "config", "sections", "statusline.yaml")
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Fatalf("precondition: statusline.yaml still present after the wipe (stat err %v); "+
			"the arms below would not be measuring a wipe-then-restore cycle", statErr)
	}

	shipped, err := fs.ReadFile(tmplFS, statuslineTemplatePath)
	if err != nil {
		t.Fatalf("read shipped %s: %v", statuslineTemplatePath, err)
	}
	if strings.Contains(string(shipped), "github:") {
		t.Fatalf("premise drift: the shipped template now carries a github key; "+
			"this test's old-only-key branch no longer describes the scenario:\n%s", shipped)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir sections after wipe: %v", err)
	}
	if err := os.WriteFile(target, shipped, 0o644); err != nil {
		t.Fatalf("redeploy shipped statusline.yaml: %v", err)
	}
}

// runUpdateConfigCycle drives the production backup -> wipe -> redeploy ->
// restore sequence and returns the resulting statusline.yaml.
func runUpdateConfigCycle(t *testing.T, root string) string {
	t.Helper()

	backupDir, err := BackupMoaiConfig(root)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	if backupDir == "" {
		t.Fatalf("BackupMoaiConfig returned no backup dir; the restore step would have nothing to merge")
	}

	wipeAndRedeploy(t, root)

	if err := RestoreMoaiConfig(root, backupDir, nil); err != nil {
		t.Fatalf("RestoreMoaiConfig: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "statusline.yaml"))
	if err != nil {
		t.Fatalf("read restored statusline.yaml: %v", err)
	}
	return string(got)
}

// TestStatuslineOptOutSurvivesUpdate is the card's opted-out arm: a tree that
// disabled the github segment must still have it disabled after the cycle.
func TestStatuslineOptOutSurvivesUpdate(t *testing.T) {
	root := seedProject(t, userStatuslineYAMLOptedOut)

	got := runUpdateConfigCycle(t, root)

	if !strings.Contains(got, "github: false") {
		t.Errorf("card t771: the user's opt-out did not survive the update cycle.\n"+
			"want segments.github: false to still be present; restored file:\n%s", got)
	}
}

// TestStatuslineNoOptOutStaysAbsent is the control arm: a tree that never
// opted out must not gain a github key from the cycle. Without this arm the
// opted-out assertion could pass for the wrong reason (e.g. a merge that
// injects the key unconditionally).
func TestStatuslineNoOptOutStaysAbsent(t *testing.T) {
	root := seedProject(t, templateStatuslineYAML)

	got := runUpdateConfigCycle(t, root)

	if strings.Contains(got, "github:") {
		t.Errorf("control arm: a tree that never opted out gained a github key.\n"+
			"restored file:\n%s", got)
	}
}
