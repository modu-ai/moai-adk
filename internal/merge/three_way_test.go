package merge

import (
	"context"
	"testing"
)

func TestEngine_ThreeWayMerge_OneSideChanged(t *testing.T) {
	t.Parallel()

	engine := NewEngine()

	base := []byte("line1\nline2\nline3")
	current := []byte("line1\nline2\nline3")
	updated := []byte("line1\nline2_modified\nline3")

	result, err := engine.ThreeWayMerge(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict")
	}
	if string(result.Content) != "line1\nline2_modified\nline3" {
		t.Errorf("got %q, want %q", string(result.Content), "line1\nline2_modified\nline3")
	}
	if result.Strategy != LineMerge {
		t.Errorf("Strategy = %q, want %q", result.Strategy, LineMerge)
	}
}

func TestEngine_ThreeWayMerge_BothSidesConflict(t *testing.T) {
	t.Parallel()

	engine := NewEngine()

	base := []byte("A\nB\nC")
	current := []byte("A\nB_user\nC")
	updated := []byte("A\nB_template\nC")

	result, err := engine.ThreeWayMerge(base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasConflict {
		t.Error("expected conflict")
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}
}

func TestEngine_ThreeWayMerge_IdenticalFiles(t *testing.T) {
	t.Parallel()

	engine := NewEngine()

	content := []byte("same\ncontent\nhere")
	result, err := engine.ThreeWayMerge(content, content, content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasConflict {
		t.Error("expected no conflict for identical files")
	}
	if string(result.Content) != "same\ncontent\nhere" {
		t.Errorf("got %q, want original content", string(result.Content))
	}
}

func TestEngine_MergeFile_YAMLStrategy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	ctx := context.Background()

	base := []byte("a: 1\nb: 2\n")
	current := []byte("a: 1\nb: 2\nuser_key: custom\n")
	updated := []byte("a: 1\nb: 3\nc: 4\n")

	result, err := engine.MergeFile(ctx, "config.yaml", base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Strategy != YAMLDeep {
		t.Errorf("Strategy = %q, want %q", result.Strategy, YAMLDeep)
	}
}

func TestEngine_MergeFile_JSONStrategy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	ctx := context.Background()

	base := []byte(`{"key1": "a"}`)
	current := []byte(`{"key1": "a", "user": true}`)
	updated := []byte(`{"key1": "b"}`)

	result, err := engine.MergeFile(ctx, "settings.json", base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Strategy != JSONMerge {
		t.Errorf("Strategy = %q, want %q", result.Strategy, JSONMerge)
	}
}

func TestEngine_MergeFile_SectionStrategy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	ctx := context.Background()

	base := []byte("## Section A\ncontent_a")
	current := []byte("## Section A\ncontent_a\n## My Custom\nmy_content")
	updated := []byte("## Section A\ncontent_a_new\n## Section B\ncontent_b")

	result, err := engine.MergeFile(ctx, "CLAUDE.md", base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Strategy != SectionMerge {
		t.Errorf("Strategy = %q, want %q", result.Strategy, SectionMerge)
	}
}

func TestEngine_MergeFile_EntryStrategy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	ctx := context.Background()

	base := []byte("*.pyc\n.env")
	current := []byte("*.pyc\n.env\nmy_file")
	updated := []byte("*.pyc\n.env\n.cache/")

	result, err := engine.MergeFile(ctx, ".gitignore", base, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Strategy != EntryMerge {
		t.Errorf("Strategy = %q, want %q", result.Strategy, EntryMerge)
	}
}

func TestEngine_MergeFile_OverwriteStrategy(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	ctx := context.Background()

	current := []byte("old binary data")
	updated := []byte("new binary data")

	result, err := engine.MergeFile(ctx, "image.png", nil, current, updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Strategy != Overwrite {
		t.Errorf("Strategy = %q, want %q", result.Strategy, Overwrite)
	}
	if string(result.Content) != "new binary data" {
		t.Errorf("got %q, want %q", string(result.Content), "new binary data")
	}
}

func TestEngine_MergeFile_ContextCancelled(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := engine.MergeFile(ctx, "test.md", []byte("a"), []byte("b"), []byte("c"))
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
