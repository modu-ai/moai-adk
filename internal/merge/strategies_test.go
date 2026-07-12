package merge

import (
	"slices"
	"strings"
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

// AC-HEV2-025 (acceptance.md §D Scenario 4): when the upstream template
// carries an EMPTY MOAI:LEARNED-WORKFLOW marker and the local copy carries a
// POPULATED block, mergeSectionBased preserves the local populated block
// verbatim (no clobber) and exits without conflict. Unrelated template section
// changes are still reflected.
func TestMergeSectionBased_PreservesPopulatedLearnedBlock(t *testing.T) {
	t.Parallel()

	base := []byte(
		"## Intro\nold intro\n" +
			"## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n<!-- moai:learned-end -->\n")
	current := []byte(
		"## Intro\nold intro\n" +
			"## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n" +
			"- Always run `make build` after editing templates. <!-- ledger_key: lw-build-001 -->\n" +
			"<!-- moai:learned-end -->\n")
	updated := []byte(
		"## Intro\nnew intro from template\n" +
			"## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n<!-- moai:learned-end -->\n")

	result, err := mergeSectionBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict for empty-upstream populated-local managed block")
	}

	content := string(result.Content)
	// The local populated bullet MUST be preserved verbatim (no clobber).
	if !strings.Contains(content, "Always run `make build` after editing templates.") {
		t.Error("expected local populated LEARNED bullet to be preserved verbatim (no clobber)")
	}
	if !strings.Contains(content, "ledger_key: lw-build-001") {
		t.Error("expected local LEARNED bullet ledger_key to be preserved")
	}
	// The template change to the unrelated Intro section MUST still apply.
	if !strings.Contains(content, "new intro from template") {
		t.Error("expected unrelated template section change to be reflected")
	}
	// The template's empty marker MUST NOT duplicate or shadow the local block.
	if c := strings.Count(content, "moai:learned-start"); c != 1 {
		t.Errorf("expected exactly 1 learned-start marker, got %d (clobber suspected)", c)
	}
	if result.Strategy != SectionMerge {
		t.Errorf("Strategy = %q, want %q", result.Strategy, SectionMerge)
	}
}

// AC-HEV2-026 (second REQ-HEV2-019 AC): the minimal empty-upstream /
// populated-local case — upstream (updated) carries the empty marker, local
// (current) carries a populated block, base carries the same empty marker as
// updated. The merge result preserves the local populated block verbatim.
func TestMergeSectionBased_EmptyUpstreamPopulatedLocal(t *testing.T) {
	t.Parallel()

	emptyMarker := "## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n<!-- moai:learned-end -->"
	populated := "## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n" +
		"- Prefer composition over inheritance in hook handlers. <!-- ledger_key: lw-comp-001 -->\n" +
		"- Characterize legacy behavior before refactoring. <!-- ledger_key: lw-char-002 -->\n" +
		"<!-- moai:learned-end -->"

	base := []byte(emptyMarker)
	current := []byte(populated)
	updated := []byte(emptyMarker)

	result, err := mergeSectionBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict when upstream is empty and local is populated")
	}

	got := string(result.Content)
	for _, want := range []string{
		"Prefer composition over inheritance in hook handlers.",
		"ledger_key: lw-comp-001",
		"Characterize legacy behavior before refactoring.",
		"ledger_key: lw-char-002",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected merged output to preserve %q verbatim; output:\n%s", want, got)
		}
	}
	// The empty upstream marker is NOT restored over the local block.
	if strings.Count(got, "moai:learned-start") != 1 {
		t.Errorf("expected exactly 1 learned-start marker; got output:\n%s", got)
	}
}

// AC-HEV2-027 (REQ-HEV2-020 no silent clobber): when both upstream and local
// carry conflicting POPULATED content inside the LEARNED-WORKFLOW marker
// boundaries, the merge surfaces a conflict. It is NOT auto-resolved.
func TestMergeSectionBased_ConflictInsideMarkers_Surfaced(t *testing.T) {
	t.Parallel()

	base := []byte("## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n<!-- moai:learned-end -->")
	current := []byte("## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n" +
		"- Local-side learned bullet X. <!-- ledger_key: lw-x -->\n" +
		"<!-- moai:learned-end -->")
	updated := []byte("## MOAI:LEARNED-WORKFLOW\n<!-- moai:learned-start -->\n" +
		"- Template-side learned bullet Y conflicts with local. <!-- ledger_key: lw-y -->\n" +
		"<!-- moai:learned-end -->")

	result, err := mergeSectionBased(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasConflict {
		t.Fatal("expected a conflict when both sides carry conflicting populated content inside markers; got none (silent clobber suspected)")
	}
	if len(result.Conflicts) == 0 {
		t.Fatal("expected at least one Conflict entry")
	}

	foundLocal, foundUpdated := false, false
	for _, c := range result.Conflicts {
		if strings.Contains(c.Current, "Local-side learned bullet X") {
			foundLocal = true
		}
		if strings.Contains(c.Updated, "Template-side learned bullet Y conflicts with local") {
			foundUpdated = true
		}
	}
	if !foundLocal || !foundUpdated {
		t.Errorf("expected conflict to carry both local and updated content; conflicts=%+v", result.Conflicts)
	}
}

// AC-HEV2-049 (reachability): the managedSectionHeadings allow-list
// (design.md §D.1 H-2 option (a)) must be declared and consulted by
// mergeSectionBased on the preservation path. This test pins the recognition
// contract: the curator-managed LEARNED block headings are recognized, and
// ordinary doc headings are not.
func TestManagedSectionHeadings_RecognizesLearnedBlocks(t *testing.T) {
	t.Parallel()

	for _, h := range []string{
		"## MOAI:LEARNED-WORKFLOW",
		"## MOAI:LEARNED-WORKFLOW-LOCAL",
	} {
		if !isManagedSection(h) {
			t.Errorf("isManagedSection(%q) = false, want true (curator-managed LEARNED block heading)", h)
		}
	}
	for _, h := range []string{
		"## Intro",
		"## Configuration",
		"### Subsection",
		"## Project-Specific Configuration (Harness-Generated)",
	} {
		if isManagedSection(h) {
			t.Errorf("isManagedSection(%q) = true, want false (non-managed heading)", h)
		}
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
