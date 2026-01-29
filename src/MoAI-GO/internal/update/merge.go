package update

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/moai-adk-go/internal/output"
	"gopkg.in/yaml.v3"
)

// MergeManager handles configuration merging during updates
type MergeManager struct {
	projectDir string
	backupDir  string
}

// NewMergeManager creates a new merge manager
func NewMergeManager(projectDir, backupDir string) *MergeManager {
	return &MergeManager{
		projectDir: projectDir,
		backupDir:  backupDir,
	}
}

// MergeUserConfig merges user configuration from backup into new templates
func (mm *MergeManager) MergeUserConfig() error {
	// Merge user.yaml
	if err := mm.mergeUserYAML(); err != nil {
		return fmt.Errorf("error merging user.yaml: %w", err)
	}

	// Merge language.yaml
	if err := mm.mergeLanguageYAML(); err != nil {
		return fmt.Errorf("error merging language.yaml: %w", err)
	}

	return nil
}

// mergeUserYAML merges user configuration
func (mm *MergeManager) mergeUserYAML() error {
	backupUserPath := filepath.Join(mm.backupDir, ".moai", "config", "sections", "user.yaml")
	targetUserPath := filepath.Join(mm.projectDir, ".moai", "config", "sections", "user.yaml")

	// Check if backup has user.yaml
	if _, err := os.Stat(backupUserPath); os.IsNotExist(err) {
		// No user config to merge
		return nil
	}

	// Read backup user config
	backupData, err := os.ReadFile(backupUserPath)
	if err != nil {
		return err
	}

	var backupConfig map[string]interface{}
	if err := yaml.Unmarshal(backupData, &backupConfig); err != nil {
		return err
	}

	// Read target user config (if exists)
	var targetConfig map[string]interface{}
	targetData, err := os.ReadFile(targetUserPath)
	if err == nil {
		if err := yaml.Unmarshal(targetData, &targetConfig); err != nil {
			return err
		}
	} else {
		targetConfig = make(map[string]interface{})
	}

	// Merge: backup values take precedence for user settings
	mergedConfig := mm.mergeMaps(targetConfig, backupConfig)

	// Write merged config
	data, err := yaml.Marshal(mergedConfig)
	if err != nil {
		return err
	}

	if err := os.WriteFile(targetUserPath, data, 0644); err != nil {
		return err
	}

	fmt.Println(output.SuccessStyle.Render("✓ Merged user configuration"))
	return nil
}

// mergeLanguageYAML merges language configuration
func (mm *MergeManager) mergeLanguageYAML() error {
	backupLangPath := filepath.Join(mm.backupDir, ".moai", "config", "sections", "language.yaml")
	targetLangPath := filepath.Join(mm.projectDir, ".moai", "config", "sections", "language.yaml")

	// Check if backup has language.yaml
	if _, err := os.Stat(backupLangPath); os.IsNotExist(err) {
		// No language config to merge
		return nil
	}

	// Read backup language config
	backupData, err := os.ReadFile(backupLangPath)
	if err != nil {
		return err
	}

	var backupConfig map[string]interface{}
	if err := yaml.Unmarshal(backupData, &backupConfig); err != nil {
		return err
	}

	// Read target language config (if exists)
	var targetConfig map[string]interface{}
	targetData, err := os.ReadFile(targetLangPath)
	if err == nil {
		if err := yaml.Unmarshal(targetData, &targetConfig); err != nil {
			return err
		}
	} else {
		targetConfig = make(map[string]interface{})
	}

	// Merge: backup language settings take precedence
	mergedConfig := mm.mergeMaps(targetConfig, backupConfig)

	// Write merged config
	data, err := yaml.Marshal(mergedConfig)
	if err != nil {
		return err
	}

	if err := os.WriteFile(targetLangPath, data, 0644); err != nil {
		return err
	}

	fmt.Println(output.SuccessStyle.Render("✓ Merged language configuration"))
	return nil
}

// mergeMaps merges two maps recursively (backup values take precedence)
func (mm *MergeManager) mergeMaps(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base values
	for k, v := range base {
		result[k] = v
	}

	// Overlay values from backup
	for k, v := range overlay {
		if baseVal, exists := result[k]; exists {
			// Both maps have this key
			if baseMap, ok := baseVal.(map[string]interface{}); ok {
				if overlayMap, ok := v.(map[string]interface{}); ok {
					// Both are maps - merge recursively
					result[k] = mm.mergeMaps(baseMap, overlayMap)
					continue
				}
			}
		}
		// Use overlay value
		result[k] = v
	}

	return result
}
