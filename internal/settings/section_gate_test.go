package settings

import (
	"strings"
	"testing"
)

// section_gate_test.go — the gate section's settings-schema contract
// (SPEC-PRECOMMIT-GATE-SCOPE-001 M2 / REQ-009 / AC-010). gate.pre_commit.enabled
// is exposed as an editable seam field persisted to exactly
// .moai/config/sections/gate.yaml. Two registrations are explicit work items —
// the route (RouteForSection) and the root-key whitelist (sectionRootKeys) —
// because a missing registration fails LOUD at save time
// ("section %q is not seam-writable") and catching that failure in a test is
// not a substitute for registering it.

// TestRouteForSectionGate pins the write-route registration: section "gate"
// routes through the yamlpatch seam.
func TestRouteForSectionGate(t *testing.T) {
	t.Parallel()
	if got := RouteForSection("gate"); got != RouteSeam {
		t.Errorf("RouteForSection(\"gate\") = %v, want RouteSeam", got)
	}
}

// TestGateFieldDefShape pins the field registration: exactly one editable
// bool field, named per the seam dot-path convention, persisted to the gate
// section file at pre_commit.enabled.
func TestGateFieldDefShape(t *testing.T) {
	t.Parallel()
	fields := SectionFields(SectionGate)
	if len(fields) != 1 {
		t.Fatalf("SectionFields(SectionGate) = %d fields, want 1", len(fields))
	}
	f := fields[0]
	if f.Name != "gate.pre_commit.enabled" {
		t.Errorf("field name = %q, want %q", f.Name, "gate.pre_commit.enabled")
	}
	if f.Type != TypeBool {
		t.Errorf("field type = %q, want bool", f.Type)
	}
	if f.Persist.Kind != PersistSeam {
		t.Errorf("persist kind = %q, want seam", f.Persist.Kind)
	}
	if f.Persist.Section != "gate" {
		t.Errorf("persist section = %q, want gate", f.Persist.Section)
	}
	wantPath := []string{"gate", "pre_commit", "enabled"}
	if len(f.Persist.Path) != len(wantPath) {
		t.Fatalf("persist path = %v, want %v", f.Persist.Path, wantPath)
	}
	for i := range wantPath {
		if f.Persist.Path[i] != wantPath[i] {
			t.Fatalf("persist path = %v, want %v", f.Persist.Path, wantPath)
		}
	}
}

// TestApplySchemaEditsGateSeamRoundTrip is the AC-010(2) save-path proof at
// the settings layer: the edit lands in gate.yaml (not any other section
// file), the pre_commit.enabled key is written, and the file's comments and
// non-modeled keys survive the yamlpatch seam.
func TestApplySchemaEditsGateSeamRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "gate")

	err := ApplySchemaEdits(root, map[string]string{
		"gate.pre_commit.enabled": "true",
	})
	if err != nil {
		t.Fatalf("ApplySchemaEdits(gate): %v", err)
	}
	after := readSection(t, root, "gate")
	if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Error("comments not preserved by the gate seam write")
	}
	if !strings.Contains(after, "user_only_key: keep-me") {
		t.Error("non-modeled key lost by the gate seam write")
	}

	values, err := SchemaCurrentValues(root)
	if err != nil {
		t.Fatalf("SchemaCurrentValues: %v", err)
	}
	if got := values["gate.pre_commit.enabled"]; got != "true" {
		t.Errorf("SchemaCurrentValues[gate.pre_commit.enabled] = %q, want %q", got, "true")
	}
}
