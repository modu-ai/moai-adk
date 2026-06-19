package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// acknowledgedUnloadedSections is the allowlist of YAML sections that intentionally
// have no loader in internal/config/loader.go (outside the Loader.Load() chain).
// Each entry should have a comment explaining why it's out of scope.
// REQ-MIG003-013 (OQ6 decision): maintain as sorted []string literal.
var acknowledgedUnloadedSections = []string{
	"db",             // out-of-scope: separate SPEC (database config not yet runtime-consumed)
	"gate",           // out-of-scope: GateConfig exists but loaded via separate path (not Loader.Load)
	"github-actions", // out-of-scope: CI config, not consumed at runtime
	"lsp",            // out-of-scope: LSP config not yet runtime-enforced (separate SPEC)
	"memo",           // out-of-scope: memo configuration, not runtime-consumed
	"mx",             // out-of-scope: ad-hoc parsing retained; struct neuverbalisation deferred (spec.md §2.2)
	"observability",  // out-of-scope: observability config, separate SPEC
	"project",        // out-of-scope: loaded via separate ProjectConfig loader path
	"security",       // out-of-scope: security config, partial loader via separate path
	"sunset",         // out-of-scope: DORMANT — struct defined but no runtime hot path (REQ-MIG003-006)
	"system",         // out-of-scope: SystemConfig has partial loader via template
	// NOTE: "workflow" removed — workflow.yaml now has a complete loader path via
	// Loader.Load() → loadWorkflowSection (SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 REQ-WSE-008).
}

// acknowledgedDedicatedLoaders is the list of sections loaded via dedicated entry-points
// outside Loader.Load() (e.g., LoadHarnessConfig, LoadRuntime). These count as "loaded"
// for the purposes of the completeness audit.
// REQ-MIG003-013 (plan.md OQ1 decision: harness stays outside Loader.Load() by design).
var acknowledgedDedicatedLoaders = []string{
	"cache",       // dedicated: LoadCacheConfig (SPEC-V3R6-PROMPT-CACHE-001) — consumed by the SDK wrapper cache_control injector, not the aggregate Loader.Load chain
	"harness",     // dedicated: LoadHarnessConfig (HRN-001) — stricter FROZEN validation semantics
	"runtime",     // dedicated: LoadRuntime (internal/runtime/config.go) — separate package
	"tool-policy", // dedicated: toolpolicy.Load / LoadFromProjectDir (internal/config/toolpolicy) — codegen SSOT, loaded outside Loader.Load chain
}

// fileNameToSectionKey maps YAML filename stems to their loadedSections key when
// the two differ. This handles cases like context.yaml → "context_search" and
// git-convention.yaml → "git_convention".
var fileNameToSectionKey = map[string]string{
	"context":        "context_search", // context.yaml top-level key is context_search:
	"git-convention": "git_convention", // loaded as git_convention (underscore, not hyphen)
}

// TestAuditLoaderCompleteness verifies that every YAML file in the template
// sections directory either:
//  1. Has a corresponding loader registered in Loader.Load() (via loadedSections),
//  2. Is listed in acknowledgedDedicatedLoaders (separate entry-point), OR
//  3. Is listed in acknowledgedUnloadedSections (explicitly out-of-scope).
//
// Fails with YAML_SECTION_NO_LOADER: <name> for any uncovered section.
// REQ-MIG003-013, AC-MIG003-12
func TestAuditLoaderCompleteness(t *testing.T) {
	// Find repo root via GOFILE path
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile = .../internal/config/audit_loader_completeness_test.go
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	templateSectionsDir := filepath.Join(repoRoot, "internal", "template", "templates",
		".moai", "config", "sections")

	entries, err := os.ReadDir(templateSectionsDir)
	if err != nil {
		t.Fatalf("os.ReadDir(%q): %v", templateSectionsDir, err)
	}

	// Build quick-lookup sets
	unloadedSet := make(map[string]bool, len(acknowledgedUnloadedSections))
	for _, s := range acknowledgedUnloadedSections {
		unloadedSet[s] = true
	}

	dedicatedSet := make(map[string]bool, len(acknowledgedDedicatedLoaders))
	for _, s := range acknowledgedDedicatedLoaders {
		dedicatedSet[s] = true
	}

	// Run Loader.Load() against a real sections dir to capture loadedSections map.
	// Use the template directory itself as the config dir for audit purposes.
	loader := NewLoader()
	// Loader.Load() expects configDir/.moai/... but templateSectionsDir already is
	// the sections dir. We need to pass a parent dir such that
	// filepath.Join(parent, "config", "sections") == templateSectionsDir.
	// parent = filepath.Join(repoRoot, "internal", "template", "templates", ".moai")
	moaiDir := filepath.Join(repoRoot, "internal", "template", "templates", ".moai")
	_, _ = loader.Load(moaiDir) // ignore errors — we only need loadedSections
	loaded := loader.LoadedSections()

	var uncovered []string
	for _, entry := range entries {
		name := entry.Name()
		// Only process .yaml files (not .yaml.tmpl, not subdirectories)
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yaml.tmpl") {
			continue
		}

		// Strip .yaml suffix to get the section name
		sectionName := strings.TrimSuffix(name, ".yaml")

		// Resolve alias: some YAML filenames use a different key in loadedSections.
		lookupKey := sectionName
		if alias, ok := fileNameToSectionKey[sectionName]; ok {
			lookupKey = alias
		}

		// Check if covered
		if loaded[lookupKey] {
			continue // loaded by Loader.Load()
		}
		if dedicatedSet[sectionName] {
			continue // dedicated entry-point (HRN-001, runtime)
		}
		if unloadedSet[sectionName] {
			continue // explicitly acknowledged out-of-scope
		}

		// Not covered by any mechanism → test failure
		uncovered = append(uncovered, sectionName)
	}

	for _, name := range uncovered {
		t.Errorf("YAML_SECTION_NO_LOADER: %s (add loader or acknowledge in allowlist)", name)
	}

	if len(uncovered) > 0 {
		t.Logf("To fix: add a loader for each uncovered section OR add to acknowledgedUnloadedSections with a reason comment")
		t.Logf("Uncovered sections: %v", uncovered)
	} else {
		t.Logf("All YAML sections are accounted for (loaded=%d, dedicated=%d, acknowledged=%d)",
			len(loaded), len(acknowledgedDedicatedLoaders), len(acknowledgedUnloadedSections))
	}
}

