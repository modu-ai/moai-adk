package config

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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
			if isConfigError(err) {
				return nil, "", err
			}
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

// loadYAMLFile loads and parses a YAML file using gopkg.in/yaml.v3.
// It extracts the top-level schema_version key (REQ-V3R2-RT-005-033) and stores it
// under the sentinel key "__schema_version__" in the returned map so MergeAll can
// propagate it to Provenance.SchemaVersion.
//
// For files with known wrapper structs (e.g., qualityFileWrapper for quality.yaml),
// a strict typed unmarshal is attempted first to detect type mismatches, returning
// ConfigTypeError when the yaml content does not match the expected Go types.
// (REQ-V3R2-RT-005-013)
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M2 GREEN — REQ-010/013/033 real yaml.v3 parsing
func (r *resolver) loadYAMLFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// First pass: parse into map[string]any to get schema_version and flat keys.
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, &ConfigTypeError{
			File:         path,
			Key:          "(parse)",
			ExpectedType: "yaml document",
			ActualValue:  err.Error(),
		}
	}
	if raw == nil {
		return make(map[string]any), nil
	}

	// Extract schema_version: N from the top-level map.
	// Store under a sentinel key so MergeAll can propagate it to Provenance.
	// REQ-V3R2-RT-005-033
	var schemaVersion int
	if sv, ok := raw["schema_version"]; ok {
		switch v := sv.(type) {
		case int:
			schemaVersion = v
		default:
			// schema_version present but not an integer → ConfigTypeError
			return nil, &ConfigTypeError{
				File:         path,
				Key:          "schema_version",
				ExpectedType: "int",
				ActualValue:  fmt.Sprintf("%v (%T)", sv, v),
			}
		}
		delete(raw, "schema_version")
	}

	// Second pass: attempt strict typed unmarshal for known wrapper types.
	// This detects type mismatches (e.g., string where int expected) per REQ-013.
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	basename := base[:len(base)-len(ext)]
	if typeErr := strictUnmarshalSection(data, basename, path); typeErr != nil {
		return nil, typeErr
	}

	// Flatten the nested yaml map to "section.field" flat keys.
	result := make(map[string]any)
	flattenMap("", raw, result)

	// Store sentinel schema_version for MergeAll to pick up.
	if schemaVersion != 0 {
		result[schemaVersionSentinel] = schemaVersion
	}

	return result, nil
}

// strictUnmarshalSection attempts a typed unmarshal of the yaml data into the
// appropriate wrapper struct for the given section basename.
// Returns ConfigTypeError if the yaml content has a type mismatch against expected Go types.
// Returns nil if no known wrapper exists for this section (graceful skip) or if the
// content is valid.
//
// Note: unknown fields in yaml are silently ignored (no KnownFields strict mode) to allow
// test yaml files and future schema evolution without false positives.
// Only genuine type mismatches (e.g., sequence where string expected, string where int expected)
// are reported.
//
// REQ-V3R2-RT-005-013
func strictUnmarshalSection(data []byte, basename string, filePath string) *ConfigTypeError {
	// Strip schema_version from data before typed unmarshal to avoid
	// "field not found" when wrapper structs don't include that field.
	// We already extracted schema_version in the caller.
	stripped := stripTopLevelKey(data, "schema_version")

	var target any
	switch basename {
	case "quality":
		target = &qualityFileWrapper{}
	case "user":
		target = &userFileWrapper{}
	case "language":
		target = &languageFileWrapper{}
	case "git-convention":
		target = &gitConventionFileWrapper{}
	case "llm":
		target = &llmFileWrapper{}
	case "state":
		target = &stateFileWrapper{}
	case "statusline":
		target = &statuslineFileWrapper{}
	case "research":
		target = &researchFileWrapper{}
	default:
		// No known wrapper — skip strict check.
		return nil
	}

	// Use non-strict mode (no KnownFields) so unknown fields are silently ignored.
	// yaml.v3 still reports genuine type mismatches (e.g., !!seq into string, !!str into int).
	if err := yaml.Unmarshal(stripped, target); err != nil {
		// Extract the offending key and expected type from the yaml.Node tree + error message.
		key, expectedType := extractYAMLTypeMismatch(stripped, err.Error())
		return &ConfigTypeError{
			File:         filePath,
			Key:          key,
			ExpectedType: expectedType,
			ActualValue:  err.Error(),
		}
	}
	return nil
}

// stripTopLevelKey returns a copy of yaml data with the given top-level key removed.
// This is used to prevent "field not found" errors for meta-keys like schema_version.
func stripTopLevelKey(data []byte, key string) []byte {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return data
	}
	delete(raw, key)
	out, err := yaml.Marshal(raw)
	if err != nil {
		return data
	}
	return out
}

// extractYAMLTypeMismatch extracts the offending key path and expected type from
// a yaml.v3 type error, using the yaml.Node tree to locate the failing line.
//
// yaml.v3 error format: "yaml: unmarshal errors:\n  line N: cannot unmarshal !!tag into type"
func extractYAMLTypeMismatch(data []byte, errMsg string) (key string, expectedType string) {
	// Parse expected type from "cannot unmarshal !!tag into <type>".
	const marker = "cannot unmarshal"
	if idx := strings.Index(errMsg, marker); idx >= 0 {
		rest := errMsg[idx+len(marker):]
		if into := strings.Index(rest, "into "); into >= 0 {
			typePart := strings.TrimSpace(rest[into+5:])
			if nl := strings.IndexAny(typePart, "\n\r "); nl >= 0 {
				typePart = typePart[:nl]
			}
			expectedType = typePart
		}
	}
	if expectedType == "" {
		expectedType = "unknown"
	}

	// Try to find the failing line number from the error message.
	failLine := 0
	for _, part := range strings.Split(errMsg, "\n") {
		if strings.Contains(part, "line ") && strings.Contains(part, "cannot unmarshal") {
			// "  line N: cannot unmarshal ..."
			after := strings.Index(part, "line ")
			if after >= 0 {
				lineStr := strings.TrimSpace(part[after+5:])
				if colon := strings.Index(lineStr, ":"); colon >= 0 {
					if n, parseErr := strconv.Atoi(lineStr[:colon]); parseErr == nil {
						failLine = n
					}
				}
			}
		}
	}

	if failLine > 0 {
		// Walk the yaml.Node tree to find the key path at the failing line.
		var doc yaml.Node
		if parseErr := yaml.Unmarshal(data, &doc); parseErr == nil && len(doc.Content) > 0 {
			if foundKey := findKeyPathAtLine(doc.Content[0], failLine, ""); foundKey != "" {
				return foundKey, expectedType
			}
		}
	}

	return "(type-check)", expectedType
}

// findKeyPathAtLine recursively walks a yaml.Node tree to find the dotted key path
// that corresponds to a value node at the given line number.
func findKeyPathAtLine(node *yaml.Node, line int, prefix string) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			keyPath := keyNode.Value
			if prefix != "" {
				keyPath = prefix + "." + keyNode.Value
			}
			// Check if the value node is at the failing line.
			if valNode.Line == line {
				return keyPath
			}
			// Recurse into children.
			if result := findKeyPathAtLine(valNode, line, keyPath); result != "" {
				return result
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if result := findKeyPathAtLine(child, line, prefix); result != "" {
				return result
			}
		}
	}
	return ""
}

// isConfigError reports whether err is a config-level error (ConfigTypeError or ConfigAmbiguous)
// that should propagate immediately without being wrapped in TierReadError.
func isConfigError(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *ConfigTypeError, *ConfigAmbiguous:
		return true
	}
	return false
}

// schemaVersionSentinel is the internal flat-key used to carry schema_version
// metadata from loadYAMLFile through MergeAll into Provenance.SchemaVersion.
// It must never appear as a real config key (double-underscore guards this).
const schemaVersionSentinel = "__schema_version__"

// flattenMap recursively flattens a nested map[string]any into dot-separated keys.
// e.g., {"quality": {"coverage_threshold": 85}} → {"quality.coverage_threshold": 85}
func flattenMap(prefix string, src map[string]any, dst map[string]any) {
	for k, v := range src {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if child, ok := v.(map[string]any); ok {
			flattenMap(key, child, dst)
		} else {
			dst[key] = v
		}
	}
}

// loadYAMLSections loads all *.yaml and *.yml files from a directory.
// Sibling detection (REQ-V3R2-RT-005-041): if the same basename appears as both
// .yaml and .yml AND they define the same key with different values,
// ConfigAmbiguous is returned naming both files.
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M2 GREEN — REQ-041 yaml/yml sibling ambiguity detection
func (r *resolver) loadYAMLSections(dirPath string) (map[string]any, error) {
	data := make(map[string]any)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return data, nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Group entries by basename (without extension) to detect siblings.
	type fileEntry struct {
		path string
		ext  string
	}
	byBase := make(map[string][]fileEntry)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		base := name[:len(name)-len(ext)]
		byBase[base] = append(byBase[base], fileEntry{
			path: filepath.Join(dirPath, name),
			ext:  ext,
		})
	}

	// Process each basename group.
	for _, files := range byBase {
		if len(files) == 1 {
			// No sibling — load normally.
			fileData, err := r.loadYAMLFile(files[0].path)
			if err != nil {
				// ConfigTypeError and ConfigAmbiguous are fatal — propagate directly.
				// TierReadError wrapping is only for I/O failures (permissions, missing file).
				if isConfigError(err) {
					return nil, err
				}
				return nil, &TierReadError{Source: SrcProject, Path: files[0].path, Err: err}
			}
			for k, v := range fileData {
				// Propagate the schema_version sentinel up so MergeAll can read it.
				// Use the first non-zero sentinel seen across all files in this tier.
				if k == schemaVersionSentinel {
					if _, exists := data[k]; !exists {
						data[k] = v
					}
					continue
				}
				if _, exists := data[k]; !exists {
					data[k] = v
				}
			}
			continue
		}

		// Sibling pair (.yaml and .yml): load both and compare keys.
		dataA, errA := r.loadYAMLFile(files[0].path)
		if errA != nil {
			if isConfigError(errA) {
				return nil, errA
			}
			return nil, &TierReadError{Source: SrcProject, Path: files[0].path, Err: errA}
		}
		dataB, errB := r.loadYAMLFile(files[1].path)
		if errB != nil {
			if isConfigError(errB) {
				return nil, errB
			}
			return nil, &TierReadError{Source: SrcProject, Path: files[1].path, Err: errB}
		}

		// Check for conflicting values on shared keys.
		for k, vA := range dataA {
			if k == schemaVersionSentinel {
				continue
			}
			vB, exists := dataB[k]
			if !exists {
				continue
			}
			if !valuesEqual(vA, vB) {
				return nil, &ConfigAmbiguous{
					Key:   k,
					File1: files[0].path,
					File2: files[1].path,
				}
			}
		}

		// No conflicts: merge dataA (preferred) then dataB for keys only in B.
		for k, v := range dataA {
			if k == schemaVersionSentinel {
				if _, exists := data[k]; !exists {
					data[k] = v
				}
				continue
			}
			if _, exists := data[k]; !exists {
				data[k] = v
			}
		}
		for k, v := range dataB {
			if k == schemaVersionSentinel {
				if _, exists := data[k]; !exists {
					data[k] = v
				}
				continue
			}
			if _, exists := data[k]; !exists {
				data[k] = v
			}
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
