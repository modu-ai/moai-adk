package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
	"gopkg.in/yaml.v3"
)

// Note: collectYAMLKeys and collectStructYAMLTags helpers are omitted because
// the symmetry check uses only top-level keys (one level deep), which avoids
// false positives from deeply nested optional fields.

// symmetryTestCase defines one struct↔YAML symmetry check.
type symmetryTestCase struct {
	structType    reflect.Type
	templateYAML  string // path relative to internal/template/templates/.moai/config/sections/
	yamlTopKey    string // top-level YAML key (e.g., "constitution", "context_search")
}

// symmetryCases lists the 4 MIG-003 sections plus StatuslineConfig
// (SPEC-WEB-CONSOLE-008 SLM-4) and GitConventionConfig (SPEC-WEB-CONSOLE-009
// GCM-3 — added LAST, after the struct trim dropped Custom/Formatting/
// validation.{enabled,enforce_on_commit} and the template was rewritten flat→nested,
// so struct↔YAML are now symmetric on {convention, auto_detection, validation}).
// REQ-MIG003-016, AC-MIG003-14, REQ-WC9-007
var symmetryCases = []symmetryTestCase{
	{
		structType:   reflect.TypeOf(ConstitutionConfig{}),
		templateYAML: "constitution.yaml",
		yamlTopKey:   "constitution",
	},
	{
		structType:   reflect.TypeOf(ContextConfig{}),
		templateYAML: "context.yaml",
		yamlTopKey:   "context_search",
	},
	{
		structType:   reflect.TypeOf(InterviewConfig{}),
		templateYAML: "interview.yaml",
		yamlTopKey:   "interview",
	},
	{
		structType:   reflect.TypeOf(DesignConfig{}),
		templateYAML: "design.yaml",
		yamlTopKey:   "design",
	},
	{
		structType:   reflect.TypeOf(models.StatuslineConfig{}),
		templateYAML: "statusline.yaml",
		yamlTopKey:   "statusline",
	},
	{
		structType:   reflect.TypeOf(models.GitConventionConfig{}),
		templateYAML: "git-convention.yaml",
		yamlTopKey:   "git_convention",
	},
}

// checkSymmetry verifies that every top-level yaml tag in the Go struct has a
// corresponding key in the YAML file and vice versa.
// Reports CONFIG_STRUCT_YAML_MISMATCH findings for any discrepancy.
// Only checks top-level keys to avoid false positives from nested optional keys.
func checkSymmetry(t *testing.T, tc symmetryTestCase, yamlPath string) {
	t.Helper()

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", yamlPath, err)
	}

	// Parse YAML into map
	var rawRoot map[string]any
	if err := yaml.Unmarshal(data, &rawRoot); err != nil {
		t.Fatalf("yaml.Unmarshal(%q): %v", yamlPath, err)
	}

	// Navigate to the top-level key
	topVal, ok := rawRoot[tc.yamlTopKey]
	if !ok {
		t.Fatalf("YAML file %q missing top-level key %q", yamlPath, tc.yamlTopKey)
	}
	topMap, ok := topVal.(map[string]any)
	if !ok {
		// Some values are nil (null YAML) — treat as empty map
		if topVal == nil {
			topMap = map[string]any{}
		} else {
			t.Fatalf("YAML key %q in %q is not a map (type: %T)", tc.yamlTopKey, yamlPath, topVal)
		}
	}

	// Collect top-level YAML keys (one level deep only for symmetry check)
	yamlTopKeys := make(map[string]bool)
	for k := range topMap {
		yamlTopKeys[k] = true
	}

	// Collect top-level Go struct yaml tags (one level deep only)
	structTopTags := make(map[string]bool)
	st := tc.structType
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag := field.Tag.Get("yaml")
		if tag == "" || tag == "-" {
			continue
		}
		tagName := strings.Split(tag, ",")[0]
		if tagName != "" && tagName != "-" {
			structTopTags[tagName] = true
		}
	}

	// Check: Go struct field NOT in YAML (go-only)
	for tag := range structTopTags {
		if !yamlTopKeys[tag] {
			t.Errorf("CONFIG_STRUCT_YAML_MISMATCH: field=%s.%s, side=go-only (key %q exists in struct but not in %s)",
				tc.structType.Name(), tag, tag, tc.templateYAML)
		}
	}

	// Check: YAML key NOT in Go struct (yaml-only)
	for k := range yamlTopKeys {
		if !structTopTags[k] {
			t.Errorf("CONFIG_STRUCT_YAML_MISMATCH: field=%s.%s, side=yaml-only (key %q exists in %s but not in struct)",
				tc.structType.Name(), k, k, tc.templateYAML)
		}
	}
}

// TestStructYAMLSymmetry_Constitution verifies Go struct ↔ YAML bijection.
// REQ-MIG003-016, AC-MIG003-14
func TestStructYAMLSymmetry_Constitution(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	yamlPath := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "constitution.yaml")
	checkSymmetry(t, symmetryCases[0], yamlPath)
}

// TestStructYAMLSymmetry_Context verifies Go struct ↔ YAML bijection.
// REQ-MIG003-016, AC-MIG003-14
func TestStructYAMLSymmetry_Context(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	yamlPath := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "context.yaml")
	checkSymmetry(t, symmetryCases[1], yamlPath)
}

// TestStructYAMLSymmetry_Interview verifies Go struct ↔ YAML bijection.
// REQ-MIG003-016, AC-MIG003-14
func TestStructYAMLSymmetry_Interview(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	yamlPath := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "interview.yaml")
	checkSymmetry(t, symmetryCases[2], yamlPath)
}

// TestStructYAMLSymmetry_Design verifies Go struct ↔ YAML bijection.
// REQ-MIG003-016, AC-MIG003-14
func TestStructYAMLSymmetry_Design(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	yamlPath := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "design.yaml")
	checkSymmetry(t, symmetryCases[3], yamlPath)
}

// TestStructYAMLSymmetry_Statusline verifies models.StatuslineConfig ↔ YAML
// bijection (SPEC-WEB-CONSOLE-008 SLM-4). The struct is {Preset, Segments, Theme}
// and the template statusline.yaml has exactly preset/segments/theme top-level
// keys (mode: and refresh_interval: were removed in M1/M3), so the top-level
// key sets are symmetric.
func TestStructYAMLSymmetry_Statusline(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	yamlPath := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "statusline.yaml")
	checkSymmetry(t, symmetryCases[4], yamlPath)
}

// TestStructYAMLSymmetry_GitConvention verifies models.GitConventionConfig ↔ YAML
// bijection (SPEC-WEB-CONSOLE-009 GCM-3). The trimmed struct is {Convention,
// AutoDetection, Validation} and the rewritten template git-convention.yaml has
// exactly convention/auto_detection/validation top-level keys (custom: and
// formatting: were removed in M1/M4), so the top-level key sets are symmetric.
func TestStructYAMLSymmetry_GitConvention(t *testing.T) {
	t.Parallel()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	yamlPath := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections", "git-convention.yaml")
	checkSymmetry(t, symmetryCases[5], yamlPath)
}
