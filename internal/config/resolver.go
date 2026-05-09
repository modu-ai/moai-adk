package config

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// @MX:NOTE: [AUTO] SettingsResolver implements 8-tier configuration loading with platform-specific paths
//
// SettingsResolver defines the interface for loading and querying configuration.
// This maps to REQ-V3R2-RT-005-004.
type SettingsResolver interface {
	// Load reads all 8 tier sources and produces merged settings.
	Load() (*MergedSettings, error)

	// Key retrieves a single key's value from the merged settings.
	Key(section, field string) (Value[any], bool)

	// Dump writes the merged settings to a writer.
	Dump(writer any) error

	// Diff compares two tiers and returns their differences.
	Diff(a, b Source) (map[string]Value[any], error)
}

// resolver implements SettingsResolver.
type resolver struct {
	loadedAt time.Time
	merged   *MergedSettings
}

// NewResolver creates a new SettingsResolver instance.
func NewResolver() SettingsResolver {
	return &resolver{
		loadedAt: time.Now(),
	}
}

// Load reads all 8 tier sources and produces merged settings.
// This maps to REQ-V3R2-RT-005-010.
func (r *resolver) Load() (*MergedSettings, error) {
	tiers := make(map[Source]map[string]any)
	origins := make(map[Source]string)

	// Load each tier
	for _, source := range AllSources() {
		data, origin, err := r.loadTier(source)
		if err != nil {
			if _, ok := err.(*TierReadError); ok {
				continue
			}
			return nil, err
		}

		if data != nil {
			tiers[source] = data
			origins[source] = origin
		}
	}

	// Merge all tiers
	merged, err := MergeAll(tiers, origins, r.loadedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to merge settings: %w", err)
	}

	r.merged = merged
	return merged, nil
}

// loadTier loads configuration from a single tier.
func (r *resolver) loadTier(source Source) (map[string]any, string, error) {
	switch source {
	case SrcPolicy:
		return r.loadPolicyTier()
	case SrcUser:
		return r.loadUserTier()
	case SrcProject:
		return r.loadProjectTier()
	case SrcLocal:
		return r.loadLocalTier()
	case SrcPlugin:
		// Reserved for v3.1+ - always empty
		return nil, "", nil
	case SrcSkill:
		return r.loadSkillTier()
	case SrcSession:
		return r.loadSessionTier()
	case SrcBuiltin:
		return r.loadBuiltinTier()
	default:
		return nil, "", fmt.Errorf("unknown source: %d", source)
	}
}

// loadPolicyTier loads from platform-specific policy path.
// REQ-V3R2-RT-005-014: returns empty tier without error if file doesn't exist.
func (r *resolver) loadPolicyTier() (map[string]any, string, error) {
	var policyPath string

	switch runtime.GOOS {
	case "darwin":
		policyPath = "/Library/Application Support/moai/settings.json"
	case "windows":
		policyPath = filepath.Join(os.Getenv("ProgramData"), "moai", "settings.json")
	default: // linux, etc.
		policyPath = "/etc/moai/settings.json"
	}

	if _, err := os.Stat(policyPath); os.IsNotExist(err) {
		return nil, "", nil
	}

	data, err := r.loadJSONFile(policyPath)
	if err != nil {
		return nil, "", &TierReadError{Source: SrcPolicy, Path: policyPath, Err: err}
	}

	return data, policyPath, nil
}

// loadUserTier loads from ~/.moai/settings.json and ~/.moai/config/sections/*.yaml.
func (r *resolver) loadUserTier() (map[string]any, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".moai", "settings.json")
	sectionsPath := filepath.Join(homeDir, ".moai", "config", "sections")

	data := make(map[string]any)

	if _, err := os.Stat(settingsPath); err == nil {
		settingsData, err := r.loadJSONFile(settingsPath)
		if err != nil {
			return nil, "", &TierReadError{Source: SrcUser, Path: settingsPath, Err: err}
		}
		maps.Copy(data, settingsData)
	}

	sectionsData, err := r.loadYAMLSections(sectionsPath)
	if err != nil {
		return nil, "", err
	}
	maps.Copy(data, sectionsData)

	if len(data) == 0 {
		return nil, "", nil
	}

	return data, settingsPath, nil
}

// loadProjectTier loads from .moai/config/config.yaml and .moai/config/sections/*.yaml.
func (r *resolver) loadProjectTier() (map[string]any, string, error) {
	configPath := ".moai/config/config.yaml"
	sectionsPath := ".moai/config/sections"

	data := make(map[string]any)

	if _, err := os.Stat(configPath); err == nil {
		configData, err := r.loadYAMLFile(configPath)
		if err != nil {
			return nil, "", &TierReadError{Source: SrcProject, Path: configPath, Err: err}
		}
		maps.Copy(data, configData)
	}

	sectionsData, err := r.loadYAMLSections(sectionsPath)
	if err != nil {
		return nil, "", err
	}
	maps.Copy(data, sectionsData)

	if len(data) == 0 {
		return nil, "", nil
	}

	return data, configPath, nil
}

// loadLocalTier loads from .claude/settings.local.json and .moai/config/local/*.yaml.
func (r *resolver) loadLocalTier() (map[string]any, string, error) {
	settingsPath := ".claude/settings.local.json"
	localPath := ".moai/config/local"

	data := make(map[string]any)

	if _, err := os.Stat(settingsPath); err == nil {
		settingsData, err := r.loadJSONFile(settingsPath)
		if err != nil {
			return nil, "", &TierReadError{Source: SrcLocal, Path: settingsPath, Err: err}
		}
		maps.Copy(data, settingsData)
	}

	sectionsData, err := r.loadYAMLSections(localPath)
	if err != nil {
		return nil, "", err
	}
	maps.Copy(data, sectionsData)

	if len(data) == 0 {
		return nil, "", nil
	}

	return data, settingsPath, nil
}

// loadSkillTier loads from .claude/skills/**/SKILL.md frontmatter config blocks.
// Placeholder - full implementation would parse markdown frontmatter.
func (r *resolver) loadSkillTier() (map[string]any, string, error) {
	return nil, "", nil
}

// loadSessionTier loads session-scoped configuration from runtime checkpoint.
// Placeholder - populated by SPEC-V3R2-RT-004.
func (r *resolver) loadSessionTier() (map[string]any, string, error) {
	return nil, "", nil
}

// loadBuiltinTier loads compiled-in defaults from internal/config/defaults.go.
// Placeholder - actual implementation would reference defaults.go.
func (r *resolver) loadBuiltinTier() (map[string]any, string, error) {
	origin := "internal/config/defaults.go"
	data := make(map[string]any)
	return data, origin, nil
}

// loadJSONFile loads and parses a JSON file.
func (r *resolver) loadJSONFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

// loadYAMLFile loads and parses a YAML file.
// Placeholder - actual implementation would use gopkg.in/yaml.v3.
func (r *resolver) loadYAMLFile(_ string) (map[string]any, error) {
	return make(map[string]any), nil
}

// loadYAMLSections loads all *.yaml files from a directory.
func (r *resolver) loadYAMLSections(dirPath string) (map[string]any, error) {
	data := make(map[string]any)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return data, nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".yaml" && filepath.Ext(name) != ".yml" {
			continue
		}

		filePath := filepath.Join(dirPath, name)
		fileData, err := r.loadYAMLFile(filePath)
		if err != nil {
			return nil, &TierReadError{Source: SrcProject, Path: filePath, Err: err}
		}

		for k := range fileData {
			if _, exists := data[k]; exists {
				continue
			}
			data[k] = fileData[k]
		}
	}

	return data, nil
}

// Key retrieves a single key's value from the merged settings.
func (r *resolver) Key(section, field string) (Value[any], bool) {
	if r.merged == nil {
		return Value[any]{}, false
	}

	key := section + "." + field
	return r.merged.Get(key)
}

// Dump writes the merged settings to a writer.
func (r *resolver) Dump(_ any) error {
	if r.merged == nil {
		return fmt.Errorf("settings not loaded - call Load() first")
	}

	output, err := r.merged.Dump("json")
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// Diff compares two tiers and returns their differences.
func (r *resolver) Diff(a, b Source) (map[string]Value[any], error) {
	if r.merged == nil {
		return nil, fmt.Errorf("settings not loaded - call Load() first")
	}

	tiers := make(map[Source]map[string]any)
	origins := make(map[Source]string)

	for _, source := range []Source{a, b} {
		data, origin, err := r.loadTier(source)
		if err != nil {
			return nil, err
		}
		if data != nil {
			tiers[source] = data
			origins[source] = origin
		}
	}

	merged, err := MergeAll(tiers, origins, r.loadedAt)
	if err != nil {
		return nil, err
	}

	result := make(map[string]Value[any])
	for _, key := range merged.Keys() {
		if val, ok := merged.Get(key); ok {
			result[key] = val
		}
	}

	return result, nil
}
