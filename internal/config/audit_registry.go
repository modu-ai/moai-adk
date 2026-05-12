package config

// audit_registry.go — YAML-to-Go-struct registry for yaml↔struct parity enforcement.
//
// @MX:NOTE [AUTO] SPEC-V3R2-RT-005 M2 GREEN — REQ-008/021/043 implementation
// yaml↔Go struct parity: every .moai/config/sections/*.yaml MUST map to either
// YAMLToStructRegistry (has a loader) or YAMLAuditExceptions (deferred to a named SPEC).
// New yaml files MUST register here before merging. See audit_test.go::TestAuditParity.
//
// REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-021, REQ-V3R2-RT-005-043

import (
	"os"
	"path/filepath"
	"strings"
)

// YAMLToStructRegistry maps yaml section basename (without extension) to the
// Go struct name it corresponds to in the Config struct (or pkg/models types).
// Maintained as the single source of truth for yaml↔Go-struct audit parity.
//
// Key: yaml basename without extension (e.g., "quality" for quality.yaml).
// Value: Go struct name used in Config or a descriptive label.
var yamlToStructRegistry = map[string]string{
	"user":           "UserConfig",
	"language":       "LanguageConfig",
	"quality":        "QualityConfig",
	"project":        "ProjectConfig",
	"git-convention": "GitConventionConfig",
	"git-strategy":   "GitStrategyConfig",
	"system":         "SystemConfig",
	"llm":            "LLMConfig",
	"ralph":          "RalphConfig",
	"workflow":       "WorkflowConfig",
	"state":          "StateConfig",
	"statusline":     "StatuslineConfig",
	"gate":           "GateConfig",
	"sunset":         "SunsetConfig",
	"research":       "ResearchConfig",
	// Additional sections with partial/specialized loaders:
	"lsp":      "LSPQualityGates",
	"mx":       "MXConfig",
	"security": "SecurityConfig",
	"runtime":  "RuntimeConfig",
}

// yamlAuditExceptions registers yaml files that intentionally do NOT have Go struct
// loaders yet, along with the rationale and the SPEC that will add them.
// Entries here are excluded from orphan detection to avoid false CI failures.
//
// Key: yaml basename without extension.
// Value: reason string citing the blocking SPEC.
var yamlAuditExceptions = map[string]string{
	// 5 yaml-only artifacts deferred to SPEC-V3R2-MIG-003 loader additions:
	"constitution": "deferred to SPEC-V3R2-MIG-003 — yaml-only artifact, no Go loader",
	"context":      "deferred to SPEC-V3R2-MIG-003 — yaml-only artifact, no Go loader",
	"interview":    "deferred to SPEC-V3R2-MIG-003 — yaml-only artifact, no Go loader",
	"design":       "deferred to SPEC-V3R2-MIG-003 — yaml-only artifact, no Go loader",
	"harness":      "deferred to SPEC-V3R2-MIG-003 — yaml-only artifact, no Go loader",
}

// GetYAMLToStructRegistry returns a copy of the yaml→struct registry.
// Callers should treat the returned map as read-only.
func GetYAMLToStructRegistry() map[string]string {
	result := make(map[string]string, len(yamlToStructRegistry))
	for k, v := range yamlToStructRegistry {
		result[k] = v
	}
	return result
}

// GetYAMLAuditExceptions returns a copy of the audit exceptions map.
// Callers should treat the returned map as read-only.
func GetYAMLAuditExceptions() map[string]string {
	result := make(map[string]string, len(yamlAuditExceptions))
	for k, v := range yamlAuditExceptions {
		result[k] = v
	}
	return result
}

// IsRegisteredOrException reports whether the given yaml basename (without extension)
// is either registered in the yaml→struct registry or listed as a known exception.
//
// REQ-V3R2-RT-005-008: audit_test.go uses this to decide if a yaml file is an orphan.
func IsRegisteredOrException(name string) bool {
	if _, ok := yamlToStructRegistry[name]; ok {
		return true
	}
	if _, ok := yamlAuditExceptions[name]; ok {
		return true
	}
	return false
}

// ScanYAMLOrphans returns a list of yaml basenames (without extension) found in
// sectionsDir that are neither in the registry nor in the exceptions map.
//
// Callers pass in the registry and exceptions maps (obtained via GetYAMLToStructRegistry
// and GetYAMLAuditExceptions) to allow injection in tests.
//
// REQ-V3R2-RT-005-043, REQ-V3R2-RT-005-008
func ScanYAMLOrphans(sectionsDir string, registry map[string]string, exceptions map[string]string) []string {
	entries, err := os.ReadDir(sectionsDir)
	if err != nil {
		return nil
	}

	var orphans []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		base := strings.TrimSuffix(name, ext)
		if _, ok := registry[base]; ok {
			continue
		}
		if _, ok := exceptions[base]; ok {
			continue
		}
		orphans = append(orphans, name)
	}
	return orphans
}

// ScanOrphanStructs returns a list of registry keys whose yaml files are absent
// from the given sectionsDir. This detects registry entries that have no matching
// yaml file on disk (the "orphan struct" direction of the parity check).
//
// REQ-V3R2-RT-005-008, REQ-V3R2-RT-005-021
func ScanOrphanStructs(sectionsDir string, registry map[string]string) []string {
	// Build set of yaml basenames present on disk
	entries, err := os.ReadDir(sectionsDir)
	present := map[string]bool{}
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext != ".yaml" && ext != ".yml" {
				continue
			}
			base := strings.TrimSuffix(name, ext)
			present[base] = true
		}
	}

	var orphans []string
	for yamlName := range registry {
		if !present[yamlName] {
			orphans = append(orphans, yamlName)
		}
	}
	return orphans
}
