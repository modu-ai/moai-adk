// Package cli — sync tests guaranteeing Wave 1 (SPEC-V3R3-CI-AUTONOMY-001)
// integrity invariants do not silently regress.
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPrePushTemplateMatchesConstant asserts that the pre-push hook template
// (deployed by `moai init` / `moai update` template sync) and the Go constant
// `prePushHookContent` (used by the InstallPrePushHook installer) are
// byte-identical. Divergence means a user installing via template sync vs
// via the Go installer would receive different hook behavior, which is a
// silent regression.
//
// REQ-CIAUT-001 — single canonical hook content.
func TestPrePushTemplateMatchesConstant(t *testing.T) {
	t.Parallel()

	// Walk up from the test file until we find the project root marker (go.mod).
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := wd
	for {
		if _, statErr := os.Stat(filepath.Join(root, "go.mod")); statErr == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Skip("project root not found (go.mod missing) — skipping in isolated test environments")
		}
		root = parent
	}

	templatePath := filepath.Join(root, "internal", "template", "templates", ".git_hooks", "pre-push")
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		t.Skipf("template not found at %s — skipping (acceptable in tarball test environments): %v", templatePath, err)
		return
	}

	if string(templateBytes) != prePushHookContent {
		t.Fatalf(
			"pre-push hook template diverges from prePushHookContent.\n"+
				"  template path: %s\n"+
				"  template len:  %d\n"+
				"  constant len:  %d\n"+
				"Both must be byte-identical (REQ-CIAUT-001).",
			templatePath, len(templateBytes), len(prePushHookContent),
		)
	}
}

// TestCrossCompileScriptMirrorsCanonical asserts that the canonical
// scripts/ci-mirror/cross-compile.sh and its template mirror under
// internal/template/templates/scripts/ci-mirror/cross-compile.sh are
// byte-identical. Divergence breaks `moai init`/`moai update` users who
// receive the template version.
//
// REQ-CIAUT-003 — full 6-target cross-compile matrix (linux/darwin/windows × amd64/arm64).
func TestCrossCompileScriptMirrorsCanonical(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := wd
	for {
		if _, statErr := os.Stat(filepath.Join(root, "go.mod")); statErr == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Skip("project root not found (go.mod missing)")
		}
		root = parent
	}

	canonicalPath := filepath.Join(root, "scripts", "ci-mirror", "cross-compile.sh")
	templatePath := filepath.Join(root, "internal", "template", "templates", "scripts", "ci-mirror", "cross-compile.sh")

	canonical, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Skipf("canonical script not found at %s — skipping: %v", canonicalPath, err)
		return
	}
	template, err := os.ReadFile(templatePath)
	if err != nil {
		t.Skipf("template not found at %s — skipping: %v", templatePath, err)
		return
	}

	if string(canonical) != string(template) {
		t.Fatalf(
			"cross-compile.sh template diverges from canonical.\n"+
				"  canonical path: %s\n"+
				"  template path:  %s\n"+
				"  canonical len:  %d\n"+
				"  template len:   %d\n"+
				"Both must be byte-identical (Template-First, REQ-CIAUT-003).",
			canonicalPath, templatePath, len(canonical), len(template),
		)
	}
}
