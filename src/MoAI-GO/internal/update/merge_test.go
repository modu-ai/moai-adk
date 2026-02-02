package update

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- NewMergeManager ---

func TestNewMergeManager(t *testing.T) {
	mm := NewMergeManager("/project", "/backup")
	if mm == nil {
		t.Fatal("NewMergeManager returned nil")
	}
	if mm.projectDir != "/project" {
		t.Errorf("projectDir = %q, want %q", mm.projectDir, "/project")
	}
	if mm.backupDir != "/backup" {
		t.Errorf("backupDir = %q, want %q", mm.backupDir, "/backup")
	}
}

// --- mergeMaps ---

func TestMergeMaps_EmptyBase(t *testing.T) {
	mm := NewMergeManager("", "")

	base := map[string]any{}
	overlay := map[string]any{"key": "value"}

	result := mm.mergeMaps(base, overlay)

	if v, ok := result["key"]; !ok || v != "value" {
		t.Errorf("expected key=value, got %v", result)
	}
}

func TestMergeMaps_EmptyOverlay(t *testing.T) {
	mm := NewMergeManager("", "")

	base := map[string]any{"key": "base-value"}
	overlay := map[string]any{}

	result := mm.mergeMaps(base, overlay)

	if v, ok := result["key"]; !ok || v != "base-value" {
		t.Errorf("expected key=base-value, got %v", result)
	}
}

func TestMergeMaps_OverlayTakesPrecedence(t *testing.T) {
	mm := NewMergeManager("", "")

	base := map[string]any{"key": "base-value", "other": "stays"}
	overlay := map[string]any{"key": "overlay-value"}

	result := mm.mergeMaps(base, overlay)

	if v := result["key"]; v != "overlay-value" {
		t.Errorf("key = %v, want overlay-value", v)
	}
	if v := result["other"]; v != "stays" {
		t.Errorf("other = %v, want stays", v)
	}
}

func TestMergeMaps_NestedMaps(t *testing.T) {
	mm := NewMergeManager("", "")

	base := map[string]any{
		"nested": map[string]any{
			"key1": "base1",
			"key2": "base2",
		},
	}
	overlay := map[string]any{
		"nested": map[string]any{
			"key2": "overlay2",
			"key3": "overlay3",
		},
	}

	result := mm.mergeMaps(base, overlay)

	nested, ok := result["nested"].(map[string]any)
	if !ok {
		t.Fatal("nested is not a map")
	}

	if nested["key1"] != "base1" {
		t.Errorf("nested.key1 = %v, want base1", nested["key1"])
	}
	if nested["key2"] != "overlay2" {
		t.Errorf("nested.key2 = %v, want overlay2 (overlay should take precedence)", nested["key2"])
	}
	if nested["key3"] != "overlay3" {
		t.Errorf("nested.key3 = %v, want overlay3", nested["key3"])
	}
}

func TestMergeMaps_BothEmpty(t *testing.T) {
	mm := NewMergeManager("", "")

	result := mm.mergeMaps(map[string]any{}, map[string]any{})

	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestMergeMaps_NonMapOverlayReplacesMap(t *testing.T) {
	mm := NewMergeManager("", "")

	base := map[string]any{
		"key": map[string]any{"nested": "value"},
	}
	overlay := map[string]any{
		"key": "simple-string",
	}

	result := mm.mergeMaps(base, overlay)

	if v := result["key"]; v != "simple-string" {
		t.Errorf("key = %v, want simple-string", v)
	}
}

// --- MergeUserConfig ---

func TestMergeUserConfig_NoBackup(t *testing.T) {
	projectDir := t.TempDir()
	backupDir := t.TempDir() // Empty backup

	// Create target user.yaml
	targetDir := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	mm := NewMergeManager(projectDir, backupDir)
	err := mm.MergeUserConfig()
	if err != nil {
		t.Fatalf("MergeUserConfig() error = %v", err)
	}
}

func TestMergeUserConfig_MergesUserYAML(t *testing.T) {
	projectDir := t.TempDir()
	backupDir := t.TempDir()

	// Create backup user.yaml
	backupSections := filepath.Join(backupDir, ".moai", "config", "sections")
	if err := os.MkdirAll(backupSections, 0755); err != nil {
		t.Fatal(err)
	}
	backupUser := map[string]any{
		"user": map[string]any{
			"name": "BackupUser",
		},
	}
	backupData, err := yaml.Marshal(backupUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupSections, "user.yaml"), backupData, 0644); err != nil {
		t.Fatal(err)
	}

	// Create target user.yaml
	targetSections := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(targetSections, 0755); err != nil {
		t.Fatal(err)
	}
	targetUser := map[string]any{
		"user": map[string]any{
			"name": "TargetUser",
		},
	}
	targetData, err := yaml.Marshal(targetUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetSections, "user.yaml"), targetData, 0644); err != nil {
		t.Fatal(err)
	}

	mm := NewMergeManager(projectDir, backupDir)
	if err := mm.MergeUserConfig(); err != nil {
		t.Fatalf("MergeUserConfig() error = %v", err)
	}

	// Verify merged result - backup takes precedence
	mergedData, err := os.ReadFile(filepath.Join(targetSections, "user.yaml"))
	if err != nil {
		t.Fatalf("failed to read merged user.yaml: %v", err)
	}

	var mergedConfig map[string]any
	if err := yaml.Unmarshal(mergedData, &mergedConfig); err != nil {
		t.Fatalf("failed to unmarshal merged config: %v", err)
	}

	userMap, ok := mergedConfig["user"].(map[string]any)
	if !ok {
		t.Fatal("user key is not a map")
	}

	if userMap["name"] != "BackupUser" {
		t.Errorf("merged user name = %v, want BackupUser (backup should take precedence)", userMap["name"])
	}
}

func TestMergeUserConfig_MergesLanguageYAML(t *testing.T) {
	projectDir := t.TempDir()
	backupDir := t.TempDir()

	// Create backup language.yaml
	backupSections := filepath.Join(backupDir, ".moai", "config", "sections")
	if err := os.MkdirAll(backupSections, 0755); err != nil {
		t.Fatal(err)
	}
	backupLang := map[string]any{
		"language": map[string]any{
			"conversation_language": "ko",
		},
	}
	backupData, _ := yaml.Marshal(backupLang)
	if err := os.WriteFile(filepath.Join(backupSections, "language.yaml"), backupData, 0644); err != nil {
		t.Fatal(err)
	}

	// Create target language.yaml
	targetSections := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(targetSections, 0755); err != nil {
		t.Fatal(err)
	}
	targetLang := map[string]any{
		"language": map[string]any{
			"conversation_language": "en",
		},
	}
	targetData, _ := yaml.Marshal(targetLang)
	if err := os.WriteFile(filepath.Join(targetSections, "language.yaml"), targetData, 0644); err != nil {
		t.Fatal(err)
	}

	mm := NewMergeManager(projectDir, backupDir)
	if err := mm.MergeUserConfig(); err != nil {
		t.Fatalf("MergeUserConfig() error = %v", err)
	}

	// Verify merged result
	mergedData, err := os.ReadFile(filepath.Join(targetSections, "language.yaml"))
	if err != nil {
		t.Fatalf("failed to read merged language.yaml: %v", err)
	}

	var mergedConfig map[string]any
	if err := yaml.Unmarshal(mergedData, &mergedConfig); err != nil {
		t.Fatalf("failed to unmarshal merged config: %v", err)
	}

	langMap, ok := mergedConfig["language"].(map[string]any)
	if !ok {
		t.Fatal("language key is not a map")
	}

	if langMap["conversation_language"] != "ko" {
		t.Errorf("merged language = %v, want ko (backup should take precedence)", langMap["conversation_language"])
	}
}

func TestMergeUserConfig_NoTargetFile(t *testing.T) {
	projectDir := t.TempDir()
	backupDir := t.TempDir()

	// Create backup user.yaml only (no target)
	backupSections := filepath.Join(backupDir, ".moai", "config", "sections")
	if err := os.MkdirAll(backupSections, 0755); err != nil {
		t.Fatal(err)
	}
	backupUser := map[string]any{"user": map[string]any{"name": "BackupUser"}}
	backupData, _ := yaml.Marshal(backupUser)
	if err := os.WriteFile(filepath.Join(backupSections, "user.yaml"), backupData, 0644); err != nil {
		t.Fatal(err)
	}

	// Create target sections directory (but no user.yaml)
	targetSections := filepath.Join(projectDir, ".moai", "config", "sections")
	if err := os.MkdirAll(targetSections, 0755); err != nil {
		t.Fatal(err)
	}

	mm := NewMergeManager(projectDir, backupDir)
	if err := mm.MergeUserConfig(); err != nil {
		t.Fatalf("MergeUserConfig() error = %v", err)
	}

	// Verify backup was written as new target
	data, err := os.ReadFile(filepath.Join(targetSections, "user.yaml"))
	if err != nil {
		t.Fatalf("merged user.yaml not created: %v", err)
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	userMap, ok := config["user"].(map[string]any)
	if !ok {
		t.Fatal("user key is not a map")
	}
	if userMap["name"] != "BackupUser" {
		t.Errorf("name = %v, want BackupUser", userMap["name"])
	}
}
