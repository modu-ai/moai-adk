package merge

import (
	"slices"
	"testing"
)

func TestSelectStrategy_TableDriven(t *testing.T) {
	t.Parallel()

	selector := NewStrategySelector()

	tests := []struct {
		path     string
		expected MergeStrategy
	}{
		// YAMLDeep
		{"config.yaml", YAMLDeep},
		{".moai/config/sections/user.yml", YAMLDeep},
		{"deep/nested/settings.yaml", YAMLDeep},

		// JSONMerge
		{"settings.json", JSONMerge},
		{"manifest.json", JSONMerge},
		{".moai/manifest.json", JSONMerge},

		// SectionMerge
		{"CLAUDE.md", SectionMerge},
		{"path/to/CLAUDE.md", SectionMerge},

		// EntryMerge
		{".gitignore", EntryMerge},
		{"sub/.gitignore", EntryMerge},

		// LineMerge (default text)
		{"README.md", LineMerge},
		{"agents/expert-backend.md", LineMerge},
		{"notes.txt", LineMerge},
		{"config.toml", LineMerge},

		// Overwrite (binary/unknown)
		{"unknown.bin", Overwrite},
		{"image.png", Overwrite},
		{"photo.jpg", Overwrite},
		{"archive.zip", Overwrite},
		{"font.woff", Overwrite},
		{"data.exe", Overwrite},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			got := selector.SelectStrategy(tt.path)
			if got != tt.expected {
				t.Errorf("SelectStrategy(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestLineMergeStrategy_OneSideChanged(t *testing.T) {
	t.Parallel()

	base := []byte("line1\nline2\nline3")
	current := []byte("line1\nline2\nline3")
	updated := []byte("line1\nline2_modified\nline3")

	result, err := mergeLineBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict")
	}
	if string(result.Content) != "line1\nline2_modified\nline3" {
		t.Errorf("got %q, want %q", string(result.Content), "line1\nline2_modified\nline3")
	}
}

func TestLineMergeStrategy_BothSidesNoConflict(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB\nC\nD")
	current := []byte("A\nB_user\nC\nD")
	updated := []byte("A\nB\nC\nD_template")

	result, err := mergeLineBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict")
	}
	if string(result.Content) != "A\nB_user\nC\nD_template" {
		t.Errorf("got %q, want %q", string(result.Content), "A\nB_user\nC\nD_template")
	}
}

func TestLineMergeStrategy_BothSidesConflict(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB\nC")
	current := []byte("A\nB_user\nC")
	updated := []byte("A\nB_template\nC")

	result, err := mergeLineBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasConflict {
		t.Error("expected conflict")
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}
	c := result.Conflicts[0]
	if c.Current != "B_user" {
		t.Errorf("Conflict.Current = %q, want %q", c.Current, "B_user")
	}
	if c.Updated != "B_template" {
		t.Errorf("Conflict.Updated = %q, want %q", c.Updated, "B_template")
	}
}

func TestLineMergeStrategy_BothSidesSameChange(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB\nC")
	current := []byte("A\nB_same\nC")
	updated := []byte("A\nB_same\nC")

	result, err := mergeLineBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict for identical changes")
	}
	if string(result.Content) != "A\nB_same\nC" {
		t.Errorf("got %q, want %q", string(result.Content), "A\nB_same\nC")
	}
}

func TestEntryMergeStrategy_GitignoreMerge(t *testing.T) {
	t.Parallel()

	base := []byte("*.pyc\n__pycache__/\n.env")
	current := []byte("*.pyc\n__pycache__/\n.env\nmy_secret.txt")
	updated := []byte("*.pyc\n__pycache__/\n.env\n.moai/cache/")

	result, err := mergeEntryBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict for entry merge")
	}

	content := string(result.Content)
	for _, want := range []string{"*.pyc", "__pycache__/", ".env", "my_secret.txt", ".moai/cache/"} {
		if !containsLine(content, want) {
			t.Errorf("expected content to contain %q", want)
		}
	}
}

func TestEntryMergeStrategy_UserDeletedNotRestored(t *testing.T) {
	t.Parallel()

	base := []byte("*.pyc\n*.log\n.env")
	current := []byte("*.pyc\n.env")
	updated := []byte("*.pyc\n*.log\n.env\n.cache/")

	result, err := mergeEntryBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(result.Content)
	if containsLine(content, "*.log") {
		t.Error("expected *.log to NOT be restored (user deleted it)")
	}
	if !containsLine(content, ".cache/") {
		t.Error("expected .cache/ to be present (new template entry)")
	}
}

func TestEntryMergeStrategy_NoDuplicates(t *testing.T) {
	t.Parallel()

	base := []byte("A\nB")
	current := []byte("A\nB\nC")
	updated := []byte("A\nB\nC")

	result, err := mergeEntryBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := splitLines(string(result.Content))
	seen := make(map[string]int)
	for _, line := range lines {
		seen[line]++
	}
	for entry, count := range seen {
		if count > 1 {
			t.Errorf("duplicate entry %q found %d times", entry, count)
		}
	}
}

func TestOverwriteStrategy(t *testing.T) {
	t.Parallel()

	current := []byte("old content")
	updated := []byte("new content")

	result, err := mergeOverwrite(current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict for overwrite")
	}
	if string(result.Content) != "new content" {
		t.Errorf("got %q, want %q", string(result.Content), "new content")
	}
	if result.Strategy != Overwrite {
		t.Errorf("Strategy = %q, want %q", result.Strategy, Overwrite)
	}
}

func TestJSONMergeStrategy_ObjectMerge(t *testing.T) {
	t.Parallel()

	base := []byte(`{"key1": "a", "key2": "b"}`)
	current := []byte(`{"key1": "a", "key2": "b", "user": true}`)
	updated := []byte(`{"key1": "a", "key2": "c", "key3": "d"}`)

	result, err := mergeJSON(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict")
	}

	content := string(result.Content)
	// Check that the merged result contains all expected keys
	for _, want := range []string{`"key1"`, `"key2"`, `"key3"`, `"user"`} {
		if !contains(content, want) {
			t.Errorf("expected content to contain %s, got: %s", want, content)
		}
	}
}

func TestJSONMergeStrategy_ConflictDetection(t *testing.T) {
	t.Parallel()

	base := []byte(`{"version": "1.0"}`)
	current := []byte(`{"version": "1.1"}`)
	updated := []byte(`{"version": "2.0"}`)

	result, err := mergeJSON(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasConflict {
		t.Error("expected conflict when both sides change same key differently")
	}
}

func TestYAMLDeepMerge_UserKeyPreserved(t *testing.T) {
	t.Parallel()

	base := []byte("a: 1\nb: 2\n")
	current := []byte("a: 1\nb: 2\nuser_key: custom\n")
	updated := []byte("a: 1\nb: 3\nc: 4\n")

	result, err := mergeYAML(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(result.Content)
	if !contains(content, "user_key") {
		t.Error("expected user_key to be preserved")
	}
	if !contains(content, "c:") {
		t.Error("expected new key c to be added")
	}
}

func TestYAMLDeepMerge_NestedMerge(t *testing.T) {
	t.Parallel()

	base := []byte("server:\n  host: localhost\n  port: 8080\n")
	current := []byte("server:\n  host: localhost\n  port: 9090\n")
	updated := []byte("server:\n  host: localhost\n  port: 8080\n  timeout: 30\n")

	result, err := mergeYAML(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict")
	}

	content := string(result.Content)
	if !contains(content, "9090") {
		t.Error("expected user's port change (9090) to be preserved")
	}
	if !contains(content, "timeout") {
		t.Error("expected template's timeout addition to be reflected")
	}
}

func TestYAMLDeepMerge_ConflictDetection(t *testing.T) {
	t.Parallel()

	base := []byte("version: \"1.0\"\n")
	current := []byte("version: \"1.1\"\n")
	updated := []byte("version: \"2.0\"\n")

	result, err := mergeYAML(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasConflict {
		t.Error("expected conflict when both sides change same key")
	}
}

func TestSectionMergeStrategy_UserSectionPreserved(t *testing.T) {
	t.Parallel()

	base := []byte("## Section A\ncontent_a\n## Section B\ncontent_b")
	current := []byte("## Section A\ncontent_a\n## Section B\ncontent_b\n## My Custom\nmy_content")
	updated := []byte("## Section A\ncontent_a_new\n## Section B\ncontent_b\n## Section C\ncontent_c")

	result, err := mergeSectionBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(result.Content)
	if !contains(content, "content_a_new") {
		t.Error("expected template change (content_a_new) to be reflected")
	}
	if !contains(content, "## Section C") {
		t.Error("expected new section C to be added")
	}
	if !contains(content, "## My Custom") {
		t.Error("expected user section (My Custom) to be preserved")
	}
	if result.Strategy != SectionMerge {
		t.Errorf("Strategy = %q, want %q", result.Strategy, SectionMerge)
	}
}

func TestSectionMergeStrategy_SameSectionConflict(t *testing.T) {
	t.Parallel()

	base := []byte("## Config\ndefault")
	current := []byte("## Config\ncustom_user_config")
	updated := []byte("## Config\nnew_template_config")

	result, err := mergeSectionBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasConflict {
		t.Error("expected conflict when same section is changed by both sides")
	}
}

// Helper functions.

func containsLine(content, line string) bool {
	return slices.Contains(splitLines(content), line)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
