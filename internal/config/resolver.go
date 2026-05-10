package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// @MX:ANCHOR: [AUTO] SPEC-V3R2-RT-005 REQ-004 — single resolver contract; methods Load/Key/Dump/Diff/Reload form the substrate. Adding methods is breaking; reordering signatures breaks RT-002/003/006 imports.
// @MX:REASON: Adding methods is breaking; reordering signatures breaks RT-002/003/006 imports.
//
// SettingsResolver defines the interface for loading and querying configuration.
// This maps to REQ-V3R2-RT-005-004.
type SettingsResolver interface {
	// Load reads all 8 tier sources and produces merged settings.
	Load() (*MergedSettings, error)

	// Key retrieves a single key's value from the merged settings.
	Key(section, field string) (Value[any], bool)

	// Dump writes the merged settings to the provided writer.
	// REQ-V3R2-RT-005-004, T-RT005-42.
	Dump(writer io.Writer) error

	// Diff returns merged-view delta: keys whose winner.Source is a or b.
	// Returns an empty map (never an error) when settings are not loaded.
	// REQ-V3R2-RT-005-051, T-RT005-41/42.
	Diff(a, b Source) map[string]Value[any]
}

// resolver implements SettingsResolver.
//
// @MX:ANCHOR: [AUTO] resolver is the central 8-tier config struct; concurrency via RWMutex
// @MX:REASON: fan_in >= 3 — Load/Reload/ClearSessionTier/Key/Diff all mutate or read merged
type resolver struct {
	mu          sync.RWMutex
	loadedAt    time.Time
	merged      *MergedSettings
	tierData    map[Source]map[string]any // per-tier raw data cache for diff-aware reload
	tierOrigins map[Source]string         // per-tier origin path cache
}

// NewResolver creates a new SettingsResolver instance.
func NewResolver() SettingsResolver {
	return &resolver{
		loadedAt:    time.Now(),
		tierData:    make(map[Source]map[string]any),
		tierOrigins: make(map[Source]string),
	}
}

// Load reads all 8 tier sources and produces merged settings.
// This maps to REQ-V3R2-RT-005-010.
// Load acquires a full write lock to protect concurrent access.
func (r *resolver) Load() (*MergedSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.loadLocked()
}

// loadLocked performs the actual tier loading; caller must hold r.mu (write lock).
func (r *resolver) loadLocked() (*MergedSettings, error) {
	tiers := make(map[Source]map[string]any)
	origins := make(map[Source]string)

	// Load each tier
	for _, source := range AllSources() {
		data, origin, err := r.loadTier(source)
		if err != nil {
			if _, ok := err.(*TierReadError); ok {
				logTierReadFailure(source, origin, err)
				continue
			}
			return nil, err
		}

		if data != nil {
			tiers[source] = data
			origins[source] = origin
		}
	}

	// Cache per-tier data for diff-aware reload.
	r.tierData = tiers
	r.tierOrigins = origins

	// Merge all tiers
	merged, err := MergeAll(tiers, origins, r.loadedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to merge settings: %w", err)
	}

	r.merged = merged
	return merged, nil
}

// Reload re-parses only the tier that contains the given file path and re-merges.
// If the path does not match any known tier, this is a no-op (returns nil).
// Reload acquires a full write lock and performs an atomic cache update.
//
// @MX:WARN: [AUTO] SPEC-V3R2-RT-005 REQ-011 — diff-aware reload. Mutex held during entire reload.
// @MX:REASON: Callbacks invoked from within Reload would deadlock; if a future feature needs post-reload notification, dispatch via channel from outside the locked region — see merge.go:MergeAll for the pattern.
//
// REQ-V3R2-RT-005-011, AC-04
func (r *resolver) Reload(path string) error {
	tier, ok := r.detectTier(path)
	if !ok {
		// Unrelated path — no-op.
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Re-parse only the affected tier.
	data, origin, err := r.loadTier(tier)
	if err != nil {
		if _, ok2 := err.(*TierReadError); ok2 {
			logTierReadFailure(tier, path, err)
			// Non-fatal: keep old tier data, still re-merge below.
			data = r.tierData[tier]
			origin = r.tierOrigins[tier]
		} else {
			return fmt.Errorf("reload tier %s: %w", tier, err)
		}
	}

	// Atomic update: replace only this tier's entry.
	newTiers := make(map[Source]map[string]any, len(r.tierData))
	newOrigins := make(map[Source]string, len(r.tierOrigins))
	for k, v := range r.tierData {
		newTiers[k] = v
	}
	for k, v := range r.tierOrigins {
		newOrigins[k] = v
	}
	if data != nil {
		newTiers[tier] = data
		newOrigins[tier] = origin
	} else {
		delete(newTiers, tier)
		delete(newOrigins, tier)
	}

	merged, err := MergeAll(newTiers, newOrigins, time.Now())
	if err != nil {
		return fmt.Errorf("reload merge: %w", err)
	}

	r.tierData = newTiers
	r.tierOrigins = newOrigins
	r.merged = merged
	return nil
}

// ClearSessionTier removes all SrcSession-tier values from the merged settings
// and re-merges the remaining tiers.
// This is called on SessionEnd (SPEC-V3R2-RT-006 wiring).
//
// @MX:NOTE: [AUTO] SPEC-V3R2-RT-005 M4 GREEN — REQ-050, AC-13: session tier cleanup
//
// REQ-V3R2-RT-005-050, AC-13
func (r *resolver) ClearSessionTier() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Build new tier map without SrcSession.
	newTiers := make(map[Source]map[string]any, len(r.tierData))
	newOrigins := make(map[Source]string, len(r.tierOrigins))
	for k, v := range r.tierData {
		if k == SrcSession {
			continue
		}
		newTiers[k] = v
	}
	for k, v := range r.tierOrigins {
		if k == SrcSession {
			continue
		}
		newOrigins[k] = v
	}

	merged, err := MergeAll(newTiers, newOrigins, time.Now())
	if err != nil {
		return fmt.Errorf("clear session tier merge: %w", err)
	}

	r.tierData = newTiers
	r.tierOrigins = newOrigins
	r.merged = merged
	return nil
}

// detectTier maps a file path to the tier it belongs to.
// Returns (tier, true) if the path belongs to a known tier, or (0, false) for unrelated paths.
func (r *resolver) detectTier(path string) (Source, bool) {
	// Normalise separators for cross-platform comparison.
	clean := filepath.ToSlash(filepath.Clean(path))

	// SrcLocal: .claude/settings.local.json or .moai/config/local/
	if strings.Contains(clean, ".claude/settings.local.json") ||
		strings.Contains(clean, ".moai/config/local/") {
		return SrcLocal, true
	}

	// SrcSkill: .claude/skills/
	if strings.Contains(clean, ".claude/skills/") {
		return SrcSkill, true
	}

	// SrcProject: .moai/config/ (sections or top-level config.yaml)
	if strings.Contains(clean, ".moai/config/") {
		return SrcProject, true
	}

	// SrcUser: ~/.moai/
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userBase := filepath.ToSlash(homeDir) + "/.moai/"
		if strings.HasPrefix(clean, userBase) {
			return SrcUser, true
		}
	}

	// SrcPolicy: platform-specific policy paths.
	switch runtime.GOOS {
	case "darwin":
		if strings.HasPrefix(clean, "/Library/Application Support/moai/") {
			return SrcPolicy, true
		}
	case "windows":
		programData := filepath.ToSlash(os.Getenv("ProgramData"))
		if programData != "" && strings.HasPrefix(clean, programData+"/moai/") {
			return SrcPolicy, true
		}
	default:
		if strings.HasPrefix(clean, "/etc/moai/") {
			return SrcPolicy, true
		}
	}

	return 0, false
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
		tierErr := &TierReadError{Source: SrcPolicy, Path: policyPath, Err: err}
		logTierReadFailure(SrcPolicy, policyPath, err)
		return nil, "", tierErr
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
			tierErr := &TierReadError{Source: SrcUser, Path: settingsPath, Err: err}
			logTierReadFailure(SrcUser, settingsPath, err)
			return nil, "", tierErr
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
// Origin paths are normalized to absolute paths via filepath.Abs (T-RT005-31).
func (r *resolver) loadProjectTier() (map[string]any, string, error) {
	configPath := ".moai/config/config.yaml"
	sectionsPath := ".moai/config/sections"

	// Normalize origin to absolute path (REQ-V3R2-RT-005, T-RT005-31).
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		absConfigPath = configPath
	}

	data := make(map[string]any)

	if _, err := os.Stat(configPath); err == nil {
		configData, err := r.loadYAMLFile(configPath)
		if err != nil {
			if isConfigError(err) {
				return nil, "", err
			}
			tierErr := &TierReadError{Source: SrcProject, Path: absConfigPath, Err: err}
			logTierReadFailure(SrcProject, absConfigPath, err)
			return nil, "", tierErr
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

	return data, absConfigPath, nil
}

// loadLocalTier loads from .claude/settings.local.json and .moai/config/local/*.yaml.
// Origin paths are normalized to absolute paths via filepath.Abs (T-RT005-31).
func (r *resolver) loadLocalTier() (map[string]any, string, error) {
	settingsPath := ".claude/settings.local.json"
	localPath := ".moai/config/local"

	// Normalize origin to absolute path (REQ-V3R2-RT-005, T-RT005-31).
	absSettingsPath, err := filepath.Abs(settingsPath)
	if err != nil {
		absSettingsPath = settingsPath
	}

	data := make(map[string]any)

	if _, err := os.Stat(settingsPath); err == nil {
		settingsData, err := r.loadJSONFile(settingsPath)
		if err != nil {
			tierErr := &TierReadError{Source: SrcLocal, Path: absSettingsPath, Err: err}
			logTierReadFailure(SrcLocal, absSettingsPath, err)
			return nil, "", tierErr
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

	return data, absSettingsPath, nil
}

// loadSkillTier loads from .claude/skills/**/SKILL.md frontmatter `config:` blocks.
// It walks .claude/skills/ recursively and extracts only the `config:` key from each
// SKILL.md frontmatter (YAML between the first pair of `---` delimiters).
// Files without frontmatter or without a `config:` key are silently skipped.
// Origin is normalized to an absolute path (T-RT005-31).
//
// @MX:NOTE: [AUTO] SPEC-V3R2-RT-005 M4 GREEN — REQ-001 (skill tier walked), REQ-015 partial
func (r *resolver) loadSkillTier() (map[string]any, string, error) {
	const skillsRoot = ".claude/skills"

	if _, err := os.Stat(skillsRoot); os.IsNotExist(err) {
		return nil, "", nil
	}

	combined := make(map[string]any)
	// Normalize origin to absolute path (T-RT005-31).
	absSkillsRoot, err := filepath.Abs(skillsRoot)
	if err != nil {
		absSkillsRoot = skillsRoot
	}
	origin := absSkillsRoot

	walkErr := filepath.WalkDir(skillsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log I/O errors on individual files; continue walking.
			logTierReadFailure(SrcSkill, path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "SKILL.md" {
			return nil
		}

		configData, parseErr := r.parseSkillFrontmatter(path)
		if parseErr != nil {
			logTierReadFailure(SrcSkill, path, parseErr)
			return nil // non-fatal; keep walking
		}
		if configData == nil {
			return nil
		}

		// Merge skill config into combined map (later files override earlier on conflict).
		for k, v := range configData {
			combined[k] = v
		}
		return nil
	})

	if walkErr != nil {
		return nil, "", &TierReadError{Source: SrcSkill, Path: skillsRoot, Err: walkErr}
	}

	if len(combined) == 0 {
		return nil, "", nil
	}

	return combined, origin, nil
}

// parseSkillFrontmatter reads a SKILL.md file, extracts the YAML frontmatter, and
// returns only the contents of the `config:` key as a flat map[string]any.
// Returns nil (no error) when the file has no frontmatter or no `config:` key.
func (r *resolver) parseSkillFrontmatter(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Frontmatter: content between first `---\n` and second `---\n`.
	content := string(raw)
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, nil // no frontmatter
	}

	// Find the closing `---`.
	rest := content[4:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, nil // unclosed frontmatter — skip
	}
	frontmatter := rest[:end]

	// Parse the frontmatter YAML.
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return nil, fmt.Errorf("frontmatter parse error: %w", err)
	}
	if fm == nil {
		return nil, nil
	}

	// Extract only the `config:` key.
	configVal, ok := fm["config"]
	if !ok {
		return nil, nil
	}

	configMap, ok := configVal.(map[string]any)
	if !ok {
		return nil, nil // config: present but not a mapping — skip
	}

	// Flatten the config map (config.key → value).
	result := make(map[string]any)
	flattenMap("", configMap, result)
	return result, nil
}

// loadSessionTier loads session-scoped configuration from runtime checkpoint.
// Placeholder - populated by SPEC-V3R2-RT-004.
func (r *resolver) loadSessionTier() (map[string]any, string, error) {
	return nil, "", nil
}

// loadBuiltinTier loads compiled-in defaults from internal/config/defaults.go.
// It reflect-walks the Config struct returned by NewDefaultConfig() and produces
// a flat map of "section.field" keys using yaml struct tags.
//
// REQ-V3R2-RT-005-020, AC-14: builtin defaults are loaded from compiled-in values.
// The origin is set to "internal/config/defaults.go" (non-filesystem path, kept as-is).
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M5 GREEN — REQ-020/014, AC-14: reflect-walk defaults
func (r *resolver) loadBuiltinTier() (map[string]any, string, error) {
	origin := "internal/config/defaults.go"
	cfg := NewDefaultConfig()
	data := flattenStruct(reflect.ValueOf(cfg).Elem())
	if len(data) == 0 {
		return nil, "", nil
	}
	return data, origin, nil
}

// flattenStruct reflect-walks a struct value and returns a flat map of
// "section.field" keys using yaml struct tags. Top-level fields become
// the section prefix; nested struct fields become the field name.
// Only primitive, bool, string, int, float, and slice/map values are emitted.
func flattenStruct(rv reflect.Value) map[string]any {
	result := make(map[string]any)
	if rv.Kind() != reflect.Struct {
		return result
	}
	rt := rv.Type()
	for i := range rt.NumField() {
		field := rt.Field(i)
		val := rv.Field(i)

		// Get the yaml tag name (fall back to lowercase field name).
		sectionName := yamlTagName(field)
		if sectionName == "" || sectionName == "-" {
			continue
		}

		// Dereference pointer.
		if val.Kind() == reflect.Pointer {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}

		// Walk nested struct as the section.
		if val.Kind() == reflect.Struct {
			flattenStructInto(sectionName, val, result)
			continue
		}

		// Top-level non-struct field: emit directly.
		if v := primitiveValue(val); v != nil {
			result[sectionName] = v
		}
	}
	return result
}

// flattenStructInto walks a struct and emits "prefix.field" keys into dst.
func flattenStructInto(prefix string, rv reflect.Value, dst map[string]any) {
	if rv.Kind() != reflect.Struct {
		return
	}
	rt := rv.Type()
	for i := range rt.NumField() {
		field := rt.Field(i)
		val := rv.Field(i)

		fieldName := yamlTagName(field)
		if fieldName == "" || fieldName == "-" {
			continue
		}
		key := prefix + "." + fieldName

		// Dereference pointer.
		if val.Kind() == reflect.Pointer {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			// Recurse into nested structs.
			flattenStructInto(key, val, dst)
			continue
		}

		if v := primitiveValue(val); v != nil {
			dst[key] = v
		}
	}
}

// yamlTagName returns the yaml tag name for a struct field, or the lowercase
// field name if no yaml tag is present. Returns "" for unexported fields.
func yamlTagName(field reflect.StructField) string {
	if !field.IsExported() {
		return ""
	}
	tag := field.Tag.Get("yaml")
	if tag == "" {
		return strings.ToLower(field.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// primitiveValue extracts a primitive value from a reflect.Value.
// Returns nil for zero/nil/struct/chan/func values that should be skipped.
func primitiveValue(v reflect.Value) any {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.String:
		s := v.String()
		if s == "" {
			return nil
		}
		return s
	case reflect.Slice:
		if v.IsNil() || v.Len() == 0 {
			return nil
		}
		return v.Interface()
	case reflect.Map:
		if v.IsNil() || v.Len() == 0 {
			return nil
		}
		return v.Interface()
	default:
		return nil
	}
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
// Key acquires a read lock for safe concurrent access.
func (r *resolver) Key(section, field string) (Value[any], bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.merged == nil {
		return Value[any]{}, false
	}

	key := section + "." + field
	return r.merged.Get(key)
}

// Dump writes the merged settings as JSON to the provided writer.
// Dump acquires a read lock for safe concurrent access.
// REQ-V3R2-RT-005-004, T-RT005-42: Dump(io.Writer) — aligned spec signature.
func (r *resolver) Dump(writer io.Writer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.merged == nil {
		return fmt.Errorf("settings not loaded - call Load() first")
	}

	output, err := r.merged.Dump("json")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(writer, output)
	return nil
}

// Diff returns merged-view delta: keys from the full 8-tier merged view
// whose winning source is either a or b.
//
// Semantics (REQ-V3R2-RT-005-051, T-RT005-41):
//   - After a full Load(), every key has exactly one winner (highest-priority tier that defines it).
//   - Diff(a, b) returns the subset of keys where winner.Source ∈ {a, b}.
//   - Keys won by neither tier are excluded.
//   - Returns an empty (non-nil) map when no settings are loaded.
//
// Diff acquires a read lock for safe concurrent access.
// T-RT005-42: returns map (no error) — error return dropped per spec interface alignment.
func (r *resolver) Diff(a, b Source) map[string]Value[any] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Value[any])
	if r.merged == nil {
		return result
	}

	for _, key := range r.merged.Keys() {
		val, ok := r.merged.Get(key)
		if !ok {
			continue
		}
		if val.P.Source == a || val.P.Source == b {
			result[key] = val
		}
	}

	return result
}
