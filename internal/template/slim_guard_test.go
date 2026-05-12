package template

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestAssertBuilderHarnessAvailable_Present verifies that when builder-harness.md
// exists in the project FS, AssertBuilderHarnessAvailable returns nil.
//
// SPEC-V3R4-CATALOG-002 T2.6 / REQ-021.
func TestAssertBuilderHarnessAvailable_Present(t *testing.T) {
	t.Parallel()

	projectFS := fstest.MapFS{
		".claude/agents/moai/builder-harness.md": &fstest.MapFile{
			Data: []byte("# builder-harness\n"),
		},
	}

	err := AssertBuilderHarnessAvailable(projectFS)
	if err != nil {
		t.Errorf("AssertBuilderHarnessAvailable(present) = %v, want nil", err)
	}
}

// TestAssertBuilderHarnessAvailable_Missing verifies that when builder-harness.md
// is absent, AssertBuilderHarnessAvailable returns a CATALOG_SLIM_HARNESS_MISSING
// error containing all 4 required diagnostic substrings.
//
// SPEC-V3R4-CATALOG-002 T2.6 / REQ-021.
func TestAssertBuilderHarnessAvailable_Missing(t *testing.T) {
	t.Parallel()

	projectFS := fstest.MapFS{
		// Some other file exists but NOT builder-harness.md
		".claude/agents/moai/expert-backend.md": &fstest.MapFile{
			Data: []byte("# expert-backend\n"),
		},
	}

	err := AssertBuilderHarnessAvailable(projectFS)
	if err == nil {
		t.Fatal("AssertBuilderHarnessAvailable(missing) = nil, want CATALOG_SLIM_HARNESS_MISSING error")
	}

	// Verify all 4 required substrings are present in the error message.
	substrings := []string{
		"CATALOG_SLIM_HARNESS_MISSING",
		"MOAI_DISTRIBUTE_ALL=1",
		"moai init --all",
		"SPEC-V3R4-CATALOG-005",
	}
	for _, sub := range substrings {
		if !strings.Contains(err.Error(), sub) {
			t.Errorf("AssertBuilderHarnessAvailable(missing) error missing substring %q; full error: %q",
				sub, err.Error())
		}
	}
}

// TestAssertBuilderHarnessAvailable_NilFS verifies the defensive nil check:
// a nil projectFS must return nil (no panic).
//
// SPEC-V3R4-CATALOG-002 T2.6 / REQ-021.
func TestAssertBuilderHarnessAvailable_NilFS(t *testing.T) {
	t.Parallel()

	err := AssertBuilderHarnessAvailable(nil)
	if err != nil {
		t.Errorf("AssertBuilderHarnessAvailable(nil) = %v, want nil", err)
	}
}
