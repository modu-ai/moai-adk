package config

// audit_loader_wiring_test.go — asserts the contract audit_registry.go actually
// documents: "every registry entry has a loader".
//
// TestAuditParity only proved the struct and the yaml file exist. Nothing
// asserted that Loader.Load reaches the section — which is why gate.yaml shipped
// with a registry entry, a struct, a yaml file, and no loader at all (issue
// #1265). This test closes that gap by running the real Loader against a tree
// carrying one yaml per registry key and asserting each key reports as loaded.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// registryToLoadedSectionKey maps a registry key (the yaml basename) to the key
// the loader records in LoadedSections, for the entries where the two differ.
var registryToLoadedSectionKey = map[string]string{
	"git-convention": "git_convention",
	"git-strategy":   "git_strategy",
	"context":        "context_search",
}

// partialLoaderExceptions enumerates registry keys whose yaml is consumed OUTSIDE
// the Loader.Load chain, so they legitimately never appear in LoadedSections.
// Each entry names where the read actually happens.
//
// This is deliberately separate from yamlAuditExceptions: that map answers "is
// this yaml file an orphan?", whereas this one answers "is this registry entry
// expected to be absent from Loader.Load?". Widening yamlAuditExceptions to
// silence this test would weaken the orphan check as a side effect.
var partialLoaderExceptions = map[string]string{
	"project":  "written/patched by internal/core/project initializer; no Loader.Load section",
	"system":   "version metadata read by internal/statusline and the project initializer, not by Loader.Load",
	"sunset":   "presence-only DORMANT notice via sunset_notice.go; no struct binding in the Load chain",
	"lsp":      "16-language LSP config consumed by the quality-gate and initializer paths, not by Loader.Load",
	"mx":       "MX thresholds read on demand by internal/cli mx_query/doctor, not by Loader.Load",
	"security": "sandbox policy read on demand by internal/sandbox and internal/cli deps, not by Loader.Load",
	"crosssession": "cross-session messaging preferences read on demand by the moai launchers " +
		"(LoadCrossSessionConfig → transient --settings injection), not by Loader.Load",
}

// minimalSectionYAML returns parseable placeholder content for a section file.
// The loader marks a section loaded when the file exists and parses, so a
// comment-only document is sufficient and keeps the fixture schema-agnostic.
func minimalSectionYAML(name string) string {
	return "# minimal " + name + ".yaml fixture for loader-wiring parity\n"
}

// TestAuditParity_EveryRegistryEntryHasLoader materialises one yaml file per
// registry key, runs the real Loader, and asserts every non-exception key is
// reported as loaded.
func TestAuditParity_EveryRegistryEntryHasLoader(t *testing.T) {
	dir := t.TempDir()
	// Loader.Load appends "config/sections" to the directory it is given.
	configDir := filepath.Join(dir, ".moai")
	sectionsDir := filepath.Join(configDir, "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	registry := GetYAMLToStructRegistry()
	for name := range registry {
		path := filepath.Join(sectionsDir, name+".yaml")
		if err := os.WriteFile(path, []byte(minimalSectionYAML(name)), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	loader := NewLoader()
	if _, err := loader.Load(configDir); err != nil {
		t.Fatalf("Loader.Load: %v", err)
	}
	loaded := loader.LoadedSections()

	var unwired []string
	for name := range registry {
		if _, exempt := partialLoaderExceptions[name]; exempt {
			continue
		}
		key := name
		if mapped, ok := registryToLoadedSectionKey[name]; ok {
			key = mapped
		}
		if !loaded[key] {
			unwired = append(unwired, name)
		}
	}
	sort.Strings(unwired)

	if len(unwired) != 0 {
		t.Errorf("registry entries with no Loader.Load wiring: %s\n"+
			"audit_registry.go documents the contract as \"every registry entry has a "+
			"loader\", but nothing asserted it — which is how gate.yaml shipped with a "+
			"registry entry and no loader (#1265). Wire the section into Loader.Load, or "+
			"record it in partialLoaderExceptions with the reader that actually consumes it.",
			strings.Join(unwired, ", "))
	}
}

// TestAuditParity_LoaderWiringGuardDetectsUnwiredEntry is the non-vacuity proof:
// a synthetic registry key with no loader must be reported.
func TestAuditParity_LoaderWiringGuardDetectsUnwiredEntry(t *testing.T) {
	dir := t.TempDir()
	// Loader.Load appends "config/sections" to the directory it is given.
	configDir := filepath.Join(dir, ".moai")
	sectionsDir := filepath.Join(configDir, "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const synthetic = "phantom"
	if err := os.WriteFile(filepath.Join(sectionsDir, synthetic+".yaml"),
		[]byte(minimalSectionYAML(synthetic)), 0o644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader()
	if _, err := loader.Load(configDir); err != nil {
		t.Fatalf("Loader.Load: %v", err)
	}
	if loader.LoadedSections()[synthetic] {
		t.Errorf("LoadedSections reported %q as loaded although no loader exists for it; "+
			"the wiring guard would pass vacuously", synthetic)
	}
}

// TestAuditParity_PartialLoaderExceptionsAreRegistryKeys keeps the exception map
// honest: an entry naming a key that is no longer in the registry is stale and
// silently exempts nothing.
func TestAuditParity_PartialLoaderExceptionsAreRegistryKeys(t *testing.T) {
	registry := GetYAMLToStructRegistry()
	for name := range partialLoaderExceptions {
		if _, ok := registry[name]; !ok {
			t.Errorf("partialLoaderExceptions entry %q is not a registry key; remove the stale exception", name)
		}
	}
}
