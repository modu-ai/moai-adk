// Package cli — deprecated_paths_text_reference_test.go
//
// Companion guard to deprecated_paths_collision_test.go for GitHub issue #1377.
//
// TestDeprecatedPaths_NoTemplateCollision detects a deprecated path the
// template SHIPS (fs.Stat over the embedded FS). It cannot detect a deprecated
// path the template merely REFERENCES in prose or configuration — which is
// exactly how `.moai/project/brand` survived deprecation while six shipped
// template files (CLAUDE.md, design.yaml, zone-registry.md, the design
// constitution, the design workflow, the shadcn catalog) plus
// config/defaults.go continued to treat it as a live path.
//
// A deprecated path that shipped documentation still calls live is a
// contradiction the build should refuse: REQ-RIL2-010 excludes DeprecatedPaths
// entries from the PRESERVE inventory, so the clean reinstall deletes the
// user's content for a path the template told them to populate.

package cli

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// textReferenceScanExts enumerates the embedded-template file extensions whose
// content is scanned for deprecated-path mentions. Binary and generated assets
// are skipped — a deprecated path can only be "documented as live" in text.
var textReferenceScanExts = map[string]bool{
	".md":   true,
	".yaml": true,
	".yml":  true,
	".json": true,
	".tmpl": true,
	".sh":   true,
}

// textReferenceAllowlist enumerates (deprecated path, referencing file) pairs
// where the template legitimately mentions a deprecated path WITHOUT treating
// it as live. Each entry carries the rationale that justifies it.
//
// The bar for an entry is narrow: the mention must be historical (a migration
// table, a retirement note, a "this was removed" changelog line). A mention
// that instructs the reader to create, populate, or read the path is a
// liveness reference and MUST be fixed by un-deprecating the path or by
// rewriting the template — never by allowlisting.
var textReferenceAllowlist = map[string]map[string]string{
	".claude/rules/agency/constitution.md": {
		".claude/rules/moai/design/constitution.md": "relocation note recording the file's own former path (\"Relocated from ... Original path: ...\") — historical provenance, not an instruction to populate the old path",
	},
	".claude/agents/moai/manager-strategy.md": {
		".moai/docs/agent-lint.md": "illustrative `moai agent lint --format=json` sample output; the filename is sample data in a JSON literal, not a path the reader is told to create",
	},
	".agency": {
		"CLAUDE.md": "retirement note (\"Legacy .agency/ directories are archived via `moai migrate agency`\") — documents the migration away from the path, not its use",
	},
}

// deprecatedTextReference records one deprecated path mentioned by one
// embedded-template text file.
type deprecatedTextReference struct {
	DeprecatedPath string
	TemplateFile   string
}

// scanDeprecatedTextReferences walks the embedded template filesystem and
// returns every (deprecated path, template file) pair where the file's text
// contains the deprecated path as a literal substring, excluding allowlisted
// pairs.
func scanDeprecatedTextReferences(t *testing.T, entries []defs.DeprecatedPathEntry) []deprecatedTextReference {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}

	var refs []deprecatedTextReference
	walkErr := fs.WalkDir(embedded, ".", func(p string, d fs.DirEntry, inner error) error {
		if inner != nil {
			return inner
		}
		if d.IsDir() {
			return nil
		}
		if !textReferenceScanExts[strings.ToLower(path.Ext(p))] {
			return nil
		}
		data, readErr := fs.ReadFile(embedded, p)
		if readErr != nil {
			return readErr
		}
		content := string(data)
		for _, entry := range entries {
			dep := strings.TrimSuffix(entry.Path, "/")
			if dep == "" || !strings.Contains(content, dep) {
				continue
			}
			if _, ok := textReferenceAllowlist[dep][p]; ok {
				continue
			}
			refs = append(refs, deprecatedTextReference{DeprecatedPath: dep, TemplateFile: p})
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk embedded templates: %v", walkErr)
	}
	return refs
}

// TestDeprecatedPaths_NoTemplateTextReference asserts that no shipped template
// text file references a defs.DeprecatedPaths entry as a live path. This is the
// structural guard that would have caught the `.moai/project/brand` defect
// (issue #1377) at build time.
func TestDeprecatedPaths_NoTemplateTextReference(t *testing.T) {
	refs := scanDeprecatedTextReferences(t, defs.DeprecatedPaths)
	if len(refs) == 0 {
		return
	}
	var b strings.Builder
	for _, r := range refs {
		b.WriteString("\n  ")
		b.WriteString(r.DeprecatedPath)
		b.WriteString("  referenced by  ")
		b.WriteString(r.TemplateFile)
	}
	t.Errorf("shipped template text references %d deprecated path(s):%s\n"+
		"A path listed in defs.DeprecatedPaths but documented as live by the template "+
		"is deleted by the clean reinstall (REQ-RIL2-010 excludes DeprecatedPaths "+
		"entries from the PRESERVE inventory) while the docs still tell users to "+
		"populate it. Either un-deprecate the path, or rewrite the template. Add a "+
		"textReferenceAllowlist entry ONLY for a historical mention (migration table, "+
		"retirement note) that does not treat the path as live.",
		len(refs), b.String())
}

// TestDeprecatedPaths_TextReferenceGuardDetectsReinsertion is the negative-path
// proof that the guard above is non-vacuous: a synthetic deprecated entry
// naming a path the template demonstrably mentions in prose MUST be reported.
func TestDeprecatedPaths_TextReferenceGuardDetectsReinsertion(t *testing.T) {
	// `.moai/project/product.md` is referenced throughout the shipped
	// documentation and is NOT deprecated — deprecating it synthetically must
	// therefore produce a text reference.
	const synthetic = ".moai/project/product.md"
	poisoned := []defs.DeprecatedPathEntry{{
		Path:            synthetic,
		DeprecatedSince: "TEST-SYNTHETIC",
		DeprecatedBy:    "TEST-SYNTHETIC",
		RemovalSchedule: "never",
	}}

	refs := scanDeprecatedTextReferences(t, poisoned)
	if len(refs) == 0 {
		t.Errorf("text-reference guard reported nothing for the synthetic deprecated "+
			"path %q, which the shipped template documents extensively; the guard is vacuous",
			synthetic)
	}
}
