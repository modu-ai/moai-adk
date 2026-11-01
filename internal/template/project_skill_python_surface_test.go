package template

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestProjectSkill_NoRunnablePythonSurface pins card t2010 / PJ-09: the
// moai-workflow-project skill ships as Go binary plus templates (reference.md
// says no Python surface exists), so no reference may present the nonexistent
// moai_menu_project module as importable or runnable, and the schema must not
// promise atomic update/backup through a UnifiedConfigManager.
func TestProjectSkill_NoRunnablePythonSurface(t *testing.T) {
	base := filepath.Join("templates", ".claude", "skills", "moai-workflow-project")

	overview, err := os.ReadFile(filepath.Join(base, "references", "overview.md"))
	if err != nil {
		t.Fatalf("read overview.md: %v", err)
	}
	if strings.Contains(string(overview), "from moai_menu_project import") {
		t.Errorf("overview.md still shows an import of the nonexistent moai_menu_project module")
	}

	examples, err := os.ReadFile(filepath.Join(base, "references", "examples.md"))
	if err != nil {
		t.Fatalf("read examples.md: %v", err)
	}
	if strings.Contains(string(examples), "python -m moai_menu_project") {
		t.Errorf("examples.md still lists runnable python -m moai_menu_project commands")
	}

	rawSchema, err := os.ReadFile(filepath.Join(base, "schemas", "tab_schema.json"))
	if err != nil {
		t.Fatalf("read tab_schema.json: %v", err)
	}
	if !json.Valid(rawSchema) {
		t.Fatalf("tab_schema.json is not valid JSON")
	}
	if strings.Contains(string(rawSchema), "delegated to UnifiedConfigManager") {
		t.Errorf("tab_schema.json still delegates atomic updates to the nonexistent UnifiedConfigManager")
	}
}

// TestProjectSkill_PythonSurfaceMirrorsMatch asserts the dogfood copies carry
// the same edits as the template sources.
func TestProjectSkill_PythonSurfaceMirrorsMatch(t *testing.T) {
	for _, rel := range []string{
		filepath.Join("references", "overview.md"),
		filepath.Join("references", "examples.md"),
		filepath.Join("schemas", "tab_schema.json"),
	} {
		src, err := os.ReadFile(filepath.Join("templates", ".claude", "skills", "moai-workflow-project", rel))
		if err != nil {
			t.Fatalf("read template %s: %v", rel, err)
		}
		local, err := os.ReadFile(filepath.Join("..", "..", ".claude", "skills", "moai-workflow-project", rel))
		if err != nil {
			t.Fatalf("read local %s: %v", rel, err)
		}
		if string(src) != string(local) {
			t.Errorf("%s: the local mirror differs from the template source", rel)
		}
	}
}
