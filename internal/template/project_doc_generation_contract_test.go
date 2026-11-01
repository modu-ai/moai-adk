// project_doc_generation_contract_test.go — card t2010. Pins two contracts of
// the project workflow's doc-generation.md that the 2026-11-01 review
// hypothesised were broken and a measurement confirmed:
//
//  1. The audit report path is one path. Step 3.1.2 reads
//     .moai/reports/PROJECT-review-1.md (where plan-auditor writes it), but the
//     retry loop pointed at .moai/reports/plan-audit/, the directory
//     audit-artifact-convention.md declares FORBIDDEN.
//  2. Phase 12 does not overwrite a development_mode the user already chose.
package template

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const projectDocGenerationDoc = ".claude/skills/moai/workflows/project/doc-generation.md"

func readProjectDocGeneration(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("templates", projectDocGenerationDoc))
	if err != nil {
		t.Fatalf("read template copy: %v", err)
	}
	return string(raw)
}

// TestProjectDocGeneration_AuditPathIsSingleAndAllowed fails when any
// PROJECT-review path names the forbidden plan-audit directory, or when the
// read path and the retry-loop paths disagree.
func TestProjectDocGeneration_AuditPathIsSingleAndAllowed(t *testing.T) {
	doc := readProjectDocGeneration(t)

	if strings.Contains(doc, "plan-audit/PROJECT-review") {
		t.Errorf("%s routes a project audit report into .moai/reports/plan-audit/, the directory audit-artifact-convention.md forbids", projectDocGenerationDoc)
	}

	paths := regexp.MustCompile("`?(\\.moai/reports/[A-Za-z0-9_./-]*PROJECT-review-[0-9]+\\.md)").FindAllStringSubmatch(doc, -1)
	if len(paths) == 0 {
		t.Fatalf("%s names no PROJECT-review report path", projectDocGenerationDoc)
	}
	for _, m := range paths {
		if !regexp.MustCompile(`^\.moai/reports/PROJECT-review-[0-9]+\.md$`).MatchString(m[1]) {
			t.Errorf("%s: audit report path %q is not .moai/reports/PROJECT-review-<N>.md", projectDocGenerationDoc, m[1])
		}
	}
}

// TestProjectDocGeneration_Phase12PreservesUserDevelopmentMode fails when
// Phase 12 gives no rule for a development_mode the user already set.
func TestProjectDocGeneration_Phase12PreservesUserDevelopmentMode(t *testing.T) {
	doc := readProjectDocGeneration(t)
	start := strings.Index(doc, "## Phase 12:")
	end := strings.Index(doc, "## Phase 14:")
	if start < 0 || end < start {
		t.Fatalf("%s: Phase 12 section not found", projectDocGenerationDoc)
	}
	section := doc[start:end]
	if !strings.Contains(section, "Preserve a user-configured value") {
		t.Errorf("Phase 12 of %s lets the auto-detected development_mode overwrite a value the user already chose", projectDocGenerationDoc)
	}
	if strings.Contains(section, "Phase 1 and Phase 1") {
		t.Errorf("Phase 12 of %s still carries the duplicated \"Phase 1 and Phase 1\" reference", projectDocGenerationDoc)
	}
}

// TestProjectWorkflow_NoDanglingPhase13 fails while project.md routes a Phase 13
// to doc-generation.md that the file does not contain (PJ-04): no section, no
// keyword reference, no producer of db-detection.json.
func TestProjectWorkflow_NoDanglingPhase13(t *testing.T) {
	const projectDoc = ".claude/skills/moai/workflows/project.md"
	raw, err := os.ReadFile(filepath.Join("templates", projectDoc))
	if err != nil {
		t.Fatalf("read %s: %v", projectDoc, err)
	}
	doc := string(raw)
	if !strings.Contains(readProjectDocGeneration(t), "## Phase 13") {
		for _, dangling := range []string{"Phase 13: DB", "db-detection.json", "Detection Keywords Reference"} {
			if strings.Contains(doc, dangling) {
				t.Errorf("%s still references %q, but doc-generation.md has no Phase 13 section", projectDoc, dangling)
			}
		}
	}
	local, err := os.ReadFile(filepath.Join("..", "..", projectDoc))
	if err != nil {
		t.Fatalf("read local %s: %v", projectDoc, err)
	}
	if strings.Contains(string(local), "db-detection.json") {
		t.Errorf("local mirror of %s still names db-detection.json", projectDoc)
	}
}

// TestProjectActivationTargetAndMirrors pins card t2010 / PJ-12: the project
// workflow names the same activation target meta-harness.md installs (AGENTS.md),
// a quoted tilde never appears in a shell example, and the dogfood copies of the
// touched files equal their template sources.
func TestProjectActivationTargetAndMirrors(t *testing.T) {
	skills := filepath.Join("templates", ".claude", "skills")
	read := func(path string) string {
		t.Helper()
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(raw)
	}

	project := read(filepath.Join(skills, "moai", "workflows", "project.md"))
	if strings.Contains(project, "Install CLAUDE.md marker") {
		t.Errorf("project.md names a CLAUDE.md marker but meta-harness.md installs the AGENTS.md marker block")
	}
	if !strings.Contains(project, "Install AGENTS.md marker") {
		t.Errorf("project.md Phase 16 row does not name the AGENTS.md marker")
	}

	navigator := read(filepath.Join(skills, "moai-workflow-project", "references", "navigator.md"))
	if strings.Contains(navigator, `"~/`) {
		t.Errorf("navigator.md quotes a tilde path; bash does not expand ~ inside double quotes")
	}

	for _, rel := range []string{
		filepath.Join("moai", "workflows", "project.md"),
		filepath.Join("moai-workflow-project", "references", "navigator.md"),
	} {
		local := read(filepath.Join("..", "..", ".claude", "skills", rel))
		if local != read(filepath.Join(skills, rel)) {
			t.Errorf("%s: the local mirror differs from the template source", rel)
		}
	}
}

// TestProjectDocGeneration_LocalMirrorMatches asserts the dogfood copy carries
// the same edit byte-for-byte.
func TestProjectDocGeneration_LocalMirrorMatches(t *testing.T) {
	local, err := os.ReadFile(filepath.Join("..", "..", projectDocGenerationDoc))
	if err != nil {
		t.Fatalf("read local copy: %v", err)
	}
	if string(local) != readProjectDocGeneration(t) {
		t.Errorf("%s: the local mirror differs from the template source", projectDocGenerationDoc)
	}
}
