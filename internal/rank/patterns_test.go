package rank

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPatternStore_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatalf("NewPatternStore() error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}

	exclude, include := store.ListPatterns()
	if len(exclude) != 0 {
		t.Errorf("expected empty exclude list, got %d", len(exclude))
	}
	if len(include) != 0 {
		t.Errorf("expected empty include list, got %d", len(include))
	}
}

func TestPatternStore_AddExclude(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("*.log"); err != nil {
		t.Fatalf("AddExclude() error: %v", err)
	}

	exclude, _ := store.ListPatterns()
	if len(exclude) != 1 || exclude[0] != "*.log" {
		t.Errorf("exclude = %v, want [*.log]", exclude)
	}
}

func TestPatternStore_AddExclude_Duplicate(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("*.log"); err != nil {
		t.Fatal(err)
	}
	if err := store.AddExclude("*.log"); err == nil {
		t.Error("expected error for duplicate exclude pattern")
	}
}

func TestPatternStore_RemoveExclude(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("*.log"); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveExclude("*.log"); err != nil {
		t.Fatalf("RemoveExclude() error: %v", err)
	}

	exclude, _ := store.ListPatterns()
	if len(exclude) != 0 {
		t.Errorf("expected empty exclude list after removal, got %d", len(exclude))
	}
}

func TestPatternStore_RemoveExclude_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveExclude("nonexistent"); err == nil {
		t.Error("expected error for removing nonexistent pattern")
	}
}

func TestPatternStore_AddInclude(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddInclude("src/*"); err != nil {
		t.Fatalf("AddInclude() error: %v", err)
	}

	_, include := store.ListPatterns()
	if len(include) != 1 || include[0] != "src/*" {
		t.Errorf("include = %v, want [src/*]", include)
	}
}

func TestPatternStore_AddInclude_Duplicate(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddInclude("src/*"); err != nil {
		t.Fatal(err)
	}
	if err := store.AddInclude("src/*"); err == nil {
		t.Error("expected error for duplicate include pattern")
	}
}

func TestPatternStore_RemoveInclude(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddInclude("src/*"); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveInclude("src/*"); err != nil {
		t.Fatalf("RemoveInclude() error: %v", err)
	}

	_, include := store.ListPatterns()
	if len(include) != 0 {
		t.Errorf("expected empty include list after removal, got %d", len(include))
	}
}

func TestPatternStore_RemoveInclude_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveInclude("nonexistent"); err == nil {
		t.Error("expected error for removing nonexistent include pattern")
	}
}

func TestPatternStore_ShouldExclude(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("debug.log"); err != nil {
		t.Fatal(err)
	}

	if !store.ShouldExclude("debug.log") {
		t.Error("expected ShouldExclude to return true for matching pattern")
	}
	if store.ShouldExclude("app.log") {
		t.Error("expected ShouldExclude to return false for non-matching pattern")
	}
}

func TestPatternStore_ShouldInclude(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddInclude("src/main.go"); err != nil {
		t.Fatal(err)
	}

	if !store.ShouldInclude("src/main.go") {
		t.Error("expected ShouldInclude to return true for matching pattern")
	}
	if store.ShouldInclude("other.go") {
		t.Error("expected ShouldInclude to return false for non-matching pattern")
	}
}

func TestPatternStore_GetConfig(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("*.tmp"); err != nil {
		t.Fatal(err)
	}
	if err := store.AddInclude("*.go"); err != nil {
		t.Fatal(err)
	}

	config := store.GetConfig()
	if len(config.ExcludePatterns) != 1 || config.ExcludePatterns[0] != "*.tmp" {
		t.Errorf("ExcludePatterns = %v, want [*.tmp]", config.ExcludePatterns)
	}
	if len(config.IncludePatterns) != 1 || config.IncludePatterns[0] != "*.go" {
		t.Errorf("IncludePatterns = %v, want [*.go]", config.IncludePatterns)
	}

	// Modify the returned config and verify original is unchanged
	config.ExcludePatterns[0] = "modified"
	originalConfig := store.GetConfig()
	if originalConfig.ExcludePatterns[0] != "*.tmp" {
		t.Error("GetConfig should return a copy, not a reference")
	}
}

func TestPatternStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	store1, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store1.AddExclude("test_pattern"); err != nil {
		t.Fatal(err)
	}

	// Reload from disk
	store2, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	exclude, _ := store2.ListPatterns()
	if len(exclude) != 1 || exclude[0] != "test_pattern" {
		t.Errorf("pattern not persisted: exclude = %v", exclude)
	}
}

func TestNewPatternStore_CorruptFile(t *testing.T) {
	dir := t.TempDir()

	// Write corrupt YAML
	rankPath := filepath.Join(dir, "rank.yaml")
	if err := os.WriteFile(rankPath, []byte(":::invalid yaml{{{"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := NewPatternStore(dir)
	if err == nil {
		t.Error("expected error for corrupt YAML file")
	}
}

func TestPatternStore_ListPatterns_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("original"); err != nil {
		t.Fatal(err)
	}

	exclude, _ := store.ListPatterns()
	exclude[0] = "modified"

	original, _ := store.ListPatterns()
	if original[0] != "original" {
		t.Error("ListPatterns should return a copy")
	}
}

func TestPatternStore_ShouldExclude_ExactMatch(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("/tmp/project"); err != nil {
		t.Fatal(err)
	}

	if !store.ShouldExclude("/tmp/project") {
		t.Error("expected ShouldExclude to match exact pattern")
	}
	if store.ShouldExclude("/tmp/other") {
		t.Error("expected ShouldExclude false for non-matching path")
	}
}

func TestPatternStore_ShouldInclude_ExactMatch(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddInclude("/home/user/project"); err != nil {
		t.Fatal(err)
	}

	if !store.ShouldInclude("/home/user/project") {
		t.Error("expected ShouldInclude to match exact pattern")
	}
	if store.ShouldInclude("/tmp/project") {
		t.Error("expected ShouldInclude false for non-matching path")
	}
}

func TestPatternStore_ShouldExclude_EmptyList(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if store.ShouldExclude("anything") {
		t.Error("ShouldExclude with empty patterns should return false")
	}
}

func TestPatternStore_ShouldInclude_EmptyList(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if store.ShouldInclude("anything") {
		t.Error("ShouldInclude with empty patterns should return false")
	}
}

func TestNewPatternStore_DefaultDir(t *testing.T) {
	// Pass empty string to trigger default-directory logic (os.UserHomeDir).
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	store, err := NewPatternStore("")
	if err != nil {
		t.Fatalf("NewPatternStore(\"\") error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	// No patterns should be loaded for a fresh home directory.
	exclude, include := store.ListPatterns()
	if len(exclude) != 0 {
		t.Errorf("expected empty exclude list, got %v", exclude)
	}
	if len(include) != 0 {
		t.Errorf("expected empty include list, got %v", include)
	}
}

func TestPatternStore_save_CreatesDirectory(t *testing.T) {
	// Verify that save() creates the config directory hierarchy if needed.
	tmpDir := t.TempDir()
	// Use a nested path that does not exist yet.
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	store, err := NewPatternStore(nestedDir)
	if err != nil {
		t.Fatalf("NewPatternStore error: %v", err)
	}

	// AddExclude triggers save() which must create the directory.
	if err := store.AddExclude("*.log"); err != nil {
		t.Fatalf("AddExclude error: %v", err)
	}

	// The rank.yaml file should now exist in the nested directory.
	rankPath := filepath.Join(nestedDir, "rank.yaml")
	if _, err := os.Stat(rankPath); err != nil {
		t.Fatalf("rank.yaml not created: %v", err)
	}
}

func TestPatternStore_MultiplePatterns(t *testing.T) {
	dir := t.TempDir()
	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.AddExclude("*.log"); err != nil {
		t.Fatal(err)
	}
	if err := store.AddExclude("*.tmp"); err != nil {
		t.Fatal(err)
	}
	if err := store.AddExclude("*.bak"); err != nil {
		t.Fatal(err)
	}

	exclude, _ := store.ListPatterns()
	if len(exclude) != 3 {
		t.Errorf("expected 3 exclude patterns, got %d", len(exclude))
	}

	if err := store.RemoveExclude("*.tmp"); err != nil {
		t.Fatal(err)
	}

	exclude, _ = store.ListPatterns()
	if len(exclude) != 2 {
		t.Errorf("expected 2 exclude patterns after removal, got %d", len(exclude))
	}
}

func TestNewPatternStore_ReadOnlyDir(t *testing.T) {
	dir := t.TempDir()

	// Create the rank.yaml so it tries to load
	rankPath := filepath.Join(dir, "rank.yaml")
	if err := os.WriteFile(rankPath, []byte("exclude_patterns: [\"*.log\"]"), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := NewPatternStore(dir)
	if err != nil {
		t.Fatalf("NewPatternStore: %v", err)
	}

	exclude, _ := store.ListPatterns()
	if len(exclude) != 1 || exclude[0] != "*.log" {
		t.Errorf("loaded exclude = %v, want [*.log]", exclude)
	}
}
