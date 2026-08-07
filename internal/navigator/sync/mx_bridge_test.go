package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBridgeMxAssociations_SkipsIgnoreDirectories is the Bug-1 repro: a tag
// placed under an ignore-directory (`.claude/`) MUST NOT be surfaced by the
// bridge, because the mx-scanner is seeded with `mx.DefaultScanIgnore` before
// `ScanDir`. Before the fix, `NewScanner()` left `ignorePatterns` nil and the
// walk descended into `.claude/`, `.git/`, `.moai/`, etc.
//
// The fixture places a Go file carrying a body-SPEC tag under `.claude/`; the
// body source (b) would associate it to SPEC-FAKE-001 if the file were walked.
// After the fix the scanner skips `.claude/` entirely, so no bridge record
// references a path under `.claude/`.
func TestBridgeMxAssociations_SkipsIgnoreDirectories(t *testing.T) {
	root := t.TempDir()
	// A Go file under an ignore-directory (`.claude/`) carrying a body SPEC
	// reference that WOULD associate if the file were walked.
	writeFixture(t, filepath.Join(root, ".claude", "hidden", "x.go"),
		"package hidden\n\n"+
			"// @MX:NOTE: SPEC-FAKE-001 placeholder\n"+
			"func Hidden() {}\n")
	// A second Go file under a NON-ignored directory also carrying the tag,
	// so the bridge returns a non-empty result and we can distinguish
	// "ignored path skipped" from "bridge returned nothing at all".
	writeFixture(t, filepath.Join(root, "src", "visible", "y.go"),
		"package visible\n\n"+
			"// @MX:NOTE: SPEC-FAKE-001 placeholder\n"+
			"func Visible() {}\n")

	got, err := BridgeMxAssociations(root)
	if err != nil {
		t.Fatalf("BridgeMxAssociations error: %v", err)
	}
	for _, b := range got {
		if strings.Contains(filepath.ToSlash(b.SourcePath), ".claude/") {
			t.Errorf("bridge surfaced a tag from ignored `.claude/` directory: %+v", b)
		}
	}
	// Sanity: the visible tag WAS surfaced, so the bridge actually ran.
	sawVisible := false
	for _, b := range got {
		if strings.HasSuffix(filepath.ToSlash(b.SourcePath), "visible/y.go") {
			sawVisible = true
		}
	}
	if !sawVisible {
		t.Errorf("bridge did not surface the visible tag; got %+v", got)
	}
	// Keep the linter happy about the otherwise-unused os import when the
	// visible check above short-circuits on some platforms.
	_ = os.PathSeparator
}

// TestBridgeMxAssociations_PathBasedAssociationIsAbsoluteVsRelativeSafe is the
// Bug-3 repro: path-based SPEC association (source (a) in AssociateWithDiagnostics)
// broke when the projectRoot was absolute because the scanner stored an
// ABSOLUTE tag.File while `mx.LoadSpecModules` returns RELATIVE module paths
// verbatim from spec.md frontmatter. `isFileUnderModules` does a plain
// `strings.HasPrefix(filePath, modulePath)`, so HasPrefix("/abs/.../internal/mx/x.go", "internal/mx/")
// was always false and path-based association never fired in the bridge.
//
// Fixture: a SPEC whose frontmatter `module: internal/mx/` + a Go file under
// `internal/mx/` whose @MX tag carries NO SPEC id in body or sub-line, so the
// ONLY association source that can match is path-based (a). projectRoot is an
// absolute t.TempDir() so the abs/rel mismatch actually occurs.
func TestBridgeMxAssociations_PathBasedAssociationIsAbsoluteVsRelativeSafe(t *testing.T) {
	root := t.TempDir()
	// Register a SPEC whose module path is `internal/mx/`.
	writeFixture(t, filepath.Join(root, ".moai", "specs", "SPEC-MX-001", "spec.md"),
		"---\nid: SPEC-MX-001\nmodule: internal/mx/\n---\n# SPEC-MX-001\n")
	// A Go file under the module path. The tag body deliberately carries NO
	// SPEC id and NO @MX:SPEC sub-line so only path-based association can match.
	writeFixture(t, filepath.Join(root, "internal", "mx", "tagged.go"),
		"package mx\n\n"+
			"// @MX:NOTE: generic context note with no SPEC reference\n"+
			"func Tagged() {}\n")

	got, err := BridgeMxAssociations(root)
	if err != nil {
		t.Fatalf("BridgeMxAssociations error: %v", err)
	}
	var match *MxBridgeSpec
	for i := range got {
		if got[i].SpecID == "SPEC-MX-001" {
			match = &got[i]
			break
		}
	}
	if match == nil {
		t.Fatalf("path-based association missing: no bridge record for SPEC-MX-001; got %+v", got)
	}
	// SourcePath must be project-relative (consistent with the rest of the
	// graph) — NOT an absolute path.
	if filepath.IsAbs(match.SourcePath) {
		t.Errorf("SourcePath is absolute, want project-relative: %q", match.SourcePath)
	}
	// Compare WITHOUT normalizing: the bridge itself must emit forward slashes.
	// Normalizing here first would mask the Windows separator defect, since
	// `filepath.Rel` returns OS-native separators and the module paths this is
	// matched against come from frontmatter YAML (always forward-slash).
	if match.SourcePath != "internal/mx/tagged.go" {
		t.Errorf("SourcePath = %q, want internal/mx/tagged.go (forward slashes on every OS)", match.SourcePath)
	}
	if strings.Contains(match.SourcePath, `\`) {
		t.Errorf("SourcePath carries an OS-native separator: %q", match.SourcePath)
	}
}
