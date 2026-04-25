package taxonomy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// TestValidateType_ValidValues verifies all 4 canonical enum values are accepted.
func TestValidateType_ValidValues(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		typ  taxonomy.MemoryType
	}{
		{"user", taxonomy.TypeUser},
		{"feedback", taxonomy.TypeFeedback},
		{"project", taxonomy.TypeProject},
		{"reference", taxonomy.TypeReference},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := taxonomy.ValidateType(tt.typ); err != nil {
				t.Errorf("ValidateType(%q) = %v, want nil", tt.typ, err)
			}
		})
	}
}

// TestValidateType_Unknown verifies that unknown type values are rejected.
func TestValidateType_Unknown(t *testing.T) {
	t.Parallel()
	for _, typ := range []taxonomy.MemoryType{"unknown", "lesson", "task", ""} {
		typ := typ
		t.Run(string(typ), func(t *testing.T) {
			t.Parallel()
			if err := taxonomy.ValidateType(typ); err == nil {
				t.Errorf("ValidateType(%q) = nil, want error", typ)
			}
		})
	}
}

// TestValidTypes_ImmutableSet verifies REQ-EXT001-004: exactly 4 types, no more.
// If this test fails after a PR adds a 5th type, the failure message cites REQ-EXT001-004.
func TestValidTypes_ImmutableSet(t *testing.T) {
	t.Parallel()
	const expected = 4
	if got := len(taxonomy.ValidTypes); got != expected {
		t.Errorf("len(ValidTypes) = %d, want %d (REQ-EXT001-004: enum is fixed at 4 values in v3.0.0; adding a type requires a new SPEC)", got, expected)
	}
}

// TestParseFile_ValidUser verifies parsing a well-formed user-type memory file.
func TestParseFile_ValidUser(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("user_role.md"))
	if err != nil {
		t.Fatalf("ParseFile(user_role.md) error = %v", err)
	}
	if fm.Type != taxonomy.TypeUser {
		t.Errorf("Type = %q, want %q", fm.Type, taxonomy.TypeUser)
	}
	if fm.Name == "" {
		t.Error("Name is empty")
	}
	if fm.Description == "" {
		t.Error("Description is empty")
	}
}

// TestParseFile_ValidFeedback verifies parsing a feedback-type memory file.
func TestParseFile_ValidFeedback(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("feedback_testing.md"))
	if err != nil {
		t.Fatalf("ParseFile(feedback_testing.md) error = %v", err)
	}
	if fm.Type != taxonomy.TypeFeedback {
		t.Errorf("Type = %q, want %q", fm.Type, taxonomy.TypeFeedback)
	}
}

// TestParseFile_ValidProject verifies parsing a project-type memory file.
func TestParseFile_ValidProject(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("project_migration.md"))
	if err != nil {
		t.Fatalf("ParseFile(project_migration.md) error = %v", err)
	}
	if fm.Type != taxonomy.TypeProject {
		t.Errorf("Type = %q, want %q", fm.Type, taxonomy.TypeProject)
	}
}

// TestParseFile_ValidReference verifies parsing a reference-type memory file.
func TestParseFile_ValidReference(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("reference_grafana.md"))
	if err != nil {
		t.Fatalf("ParseFile(reference_grafana.md) error = %v", err)
	}
	if fm.Type != taxonomy.TypeReference {
		t.Errorf("Type = %q, want %q", fm.Type, taxonomy.TypeReference)
	}
}

// TestParseFile_NoFrontmatter verifies that missing frontmatter returns ErrNoFrontmatter.
func TestParseFile_NoFrontmatter(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "no_fm.md")
	if err := os.WriteFile(path, []byte("# Just a heading\n\nsome body\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, _, err := taxonomy.ParseFile(path)
	if err == nil {
		t.Error("ParseFile(no frontmatter) = nil error, want ErrNoFrontmatter")
	}
}

// TestParseFile_Empty verifies that a zero-byte file skips without panicking.
func TestParseFile_Empty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, _, err := taxonomy.ParseFile(path)
	if err == nil {
		t.Error("ParseFile(empty) = nil error, want ErrNoFrontmatter")
	}
}

// TestParseFile_MissingType verifies that a file with frontmatter but no type key
// returns a Frontmatter with empty Type (not an error — audit catches this).
func TestParseFile_MissingType(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("missing_type.md"))
	if err != nil {
		t.Fatalf("ParseFile(missing_type.md) unexpected error = %v", err)
	}
	if fm.Type != "" {
		t.Errorf("Type = %q, want empty string for missing type", fm.Type)
	}
}

// TestParseFile_InvalidType verifies that unknown type values are parsed without
// error (ValidateType is separate) but can be caught by AuditFile.
func TestParseFile_InvalidType(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("unknown_type.md"))
	if err != nil {
		t.Fatalf("ParseFile(unknown_type.md) unexpected error = %v", err)
	}
	if err := taxonomy.ValidateType(fm.Type); err == nil {
		t.Errorf("ValidateType(%q) = nil, want error for unknown type", fm.Type)
	}
}

// TestParseFile_MissingName verifies AC-EXT001-01b: name key is absent.
func TestParseFile_MissingName(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("missing_name.md"))
	if err != nil {
		t.Fatalf("ParseFile(missing_name.md) unexpected error = %v", err)
	}
	if fm.Name != "" {
		t.Errorf("Name = %q, want empty string for missing name", fm.Name)
	}
}

// TestParseFile_MissingDescription verifies AC-EXT001-01b: description key is absent.
func TestParseFile_MissingDescription(t *testing.T) {
	t.Parallel()
	fm, _, err := taxonomy.ParseFile(fixturesPath("missing_description.md"))
	if err != nil {
		t.Fatalf("ParseFile(missing_description.md) unexpected error = %v", err)
	}
	if fm.Description != "" {
		t.Errorf("Description = %q, want empty string for missing description", fm.Description)
	}
}

// TestParseFile_MissingBoth verifies AC-EXT001-01b: both name and description absent.
func TestParseFile_MissingBoth(t *testing.T) {
	t.Parallel()
	// Create a temp file that has type but no name or description
	dir := t.TempDir()
	path := filepath.Join(dir, "missing_both.md")
	content := "---\ntype: user\n---\nContent without name or description.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	fm, _, err := taxonomy.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(missing_both) unexpected error = %v", err)
	}
	if fm.Name != "" || fm.Description != "" {
		t.Errorf("Name=%q Description=%q, want both empty", fm.Name, fm.Description)
	}
}

// TestParseFile_Symlink verifies that symlinks are not followed (security).
func TestParseFile_Symlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	real := filepath.Join(dir, "real.md")
	if err := os.WriteFile(real, []byte("---\ntype: user\nname: test\ndescription: d\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	link := filepath.Join(dir, "link.md")
	if err := os.Symlink(real, link); err != nil {
		t.Skip("symlink not supported on this OS")
	}
	_, _, err := taxonomy.ParseFile(link)
	if err == nil {
		t.Error("ParseFile(symlink) = nil error, want error (symlinks must not be followed)")
	}
}

// fixturesPath returns the absolute path to a fixture file.
func fixturesPath(name string) string {
	return filepath.Join("fixtures", name)
}
