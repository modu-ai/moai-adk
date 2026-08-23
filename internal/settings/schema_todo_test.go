package settings

import "testing"

// TestWorkflowTodoEnabledFieldRegistered is AC-T-009's first half (REQ-5).
//
// It exists because no pre-existing test observes a NEW field's registration:
// TestSchemaCurrentValuesReadsAllSections walks a fixed map of thirteen
// unrelated keys, so adding a field to the schema changes nothing it looks at.
// Without this test the AC would have no command that could fail.
func TestWorkflowTodoEnabledFieldRegistered(t *testing.T) {
	const name = "workflow.todo.enabled"

	var found *FieldDef
	for i, f := range AllFields() {
		if f.Name == name {
			found = &AllFields()[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("%s is not registered in AllFields()", name)
	}

	if found.Section != SectionWorkflow {
		t.Errorf("%s section = %q, want %q", name, found.Section, SectionWorkflow)
	}
	if found.Type != TypeBool {
		t.Errorf("%s type = %q, want %q", name, found.Type, TypeBool)
	}
	// PersistSeam is what routes the edit through yamlpatch, which is the only
	// writer that can upsert the nested mapping into a workflow.yaml that ships
	// without a todo block at all.
	if found.Persist.Kind != PersistSeam {
		t.Errorf("%s persist kind = %q, want %q", name, found.Persist.Kind, PersistSeam)
	}
	if found.Persist.Section != "workflow" {
		t.Errorf("%s persist section = %q, want %q", name, found.Persist.Section, "workflow")
	}
	want := []string{"workflow", "todo", "enabled"}
	if len(found.Persist.Path) != len(want) {
		t.Fatalf("%s persist path = %v, want %v", name, found.Persist.Path, want)
	}
	for i := range want {
		if found.Persist.Path[i] != want[i] {
			t.Fatalf("%s persist path = %v, want %v", name, found.Persist.Path, want)
		}
	}
}
