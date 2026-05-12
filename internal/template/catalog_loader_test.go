package template

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestLoadCatalog verifies the typed LoadCatalog() accessor introduced in M4 (T-019..T-021).
// This test is additive — it does NOT replace the existing audit tests in catalog_tier_audit_test.go.
func TestLoadCatalog(t *testing.T) {
	t.Parallel()

	cat, err := LoadCatalog(embeddedRaw)
	if err != nil {
		t.Fatalf("LoadCatalog() error: %v", err)
	}

	if cat.Version == "" {
		t.Error("LoadCatalog: Version is empty")
	}
	if cat.GeneratedAt == "" {
		t.Error("LoadCatalog: GeneratedAt is empty")
	}
	if cat.Catalog.OptionalPacks == nil {
		t.Error("LoadCatalog: OptionalPacks is nil")
	}

	// AllEntries should return all 65 entries
	all := cat.AllEntries()
	const expectedTotal = 65
	if len(all) != expectedTotal {
		t.Errorf("AllEntries() returned %d entries, want %d", len(all), expectedTotal)
	}

	// LookupSkill: known core skill
	e, ok := cat.LookupSkill("moai")
	if !ok {
		t.Error("LookupSkill(moai) returned false")
	} else {
		if e.Tier != TierCore {
			t.Errorf("LookupSkill(moai).Tier = %q, want %q", e.Tier, TierCore)
		}
		if e.Hash == "" || e.Hash == "TODO" {
			t.Errorf("LookupSkill(moai).Hash = %q, want 64-char hex", e.Hash)
		}
	}

	// LookupSkill: known optional-pack skill
	e, ok = cat.LookupSkill("moai-domain-backend")
	if !ok {
		t.Error("LookupSkill(moai-domain-backend) returned false")
	} else {
		if !strings.HasPrefix(e.Tier, TierOptionalPackPrefix) {
			t.Errorf("LookupSkill(moai-domain-backend).Tier = %q, want optional-pack:* prefix", e.Tier)
		}
	}

	// LookupSkill: missing returns false
	if _, ok := cat.LookupSkill("nonexistent-skill-xyz"); ok {
		t.Error("LookupSkill(nonexistent) should return false")
	}

	// LookupAgent: known core agent
	e, ok = cat.LookupAgent("evaluator-active")
	if !ok {
		t.Error("LookupAgent(evaluator-active) returned false")
	} else {
		if e.Tier != TierCore {
			t.Errorf("LookupAgent(evaluator-active).Tier = %q, want %q", e.Tier, TierCore)
		}
	}

	// LookupAgent: harness-generated
	e, ok = cat.LookupAgent("builder-harness")
	if !ok {
		t.Error("LookupAgent(builder-harness) returned false")
	} else {
		if e.Tier != TierHarnessGenerated {
			t.Errorf("LookupAgent(builder-harness).Tier = %q, want %q", e.Tier, TierHarnessGenerated)
		}
	}

	// LookupAgent: missing returns false
	if _, ok := cat.LookupAgent("nonexistent-agent-xyz"); ok {
		t.Error("LookupAgent(nonexistent) should return false")
	}

	// FormatOptionalPackTier helper
	packTier := FormatOptionalPackTier("backend")
	if packTier != "optional-pack:backend" {
		t.Errorf("FormatOptionalPackTier(backend) = %q, want optional-pack:backend", packTier)
	}

	t.Logf("LoadCatalog: version=%s, generated_at=%s, total_entries=%d",
		cat.Version, cat.GeneratedAt, len(all))
}

// TestLoadCatalog_MalformedYAML covers the yaml.Unmarshal error path of LoadCatalog.
// Required by evaluator-active (eval-1) to satisfy the 85%+ coverage gate for catalog_loader.go.
// Sentinel-style: LoadCatalog must return a non-nil error on malformed YAML input.
func TestLoadCatalog_MalformedYAML(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data []byte
	}{
		{
			name: "unbalanced_bracket",
			data: []byte("version: 1.0.0\ncatalog: [invalid yaml structure\n"),
		},
		{
			name: "tab_indentation_invalid",
			data: []byte("version: 1.0.0\n\tcatalog:\n\t  core: {}\n"),
		},
		{
			name: "incomplete_mapping",
			data: []byte("version: 1.0.0\ncatalog:\n  core:\n    skills: [\n"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fsys := fstest.MapFS{
				"catalog.yaml": &fstest.MapFile{Data: tc.data},
			}
			cat, err := LoadCatalog(fsys)
			if err == nil {
				t.Errorf("LoadCatalog(%s) returned nil error and cat=%+v; want yaml.Unmarshal error", tc.name, cat)
			}
		})
	}
}

// TestLoadCatalog_ManifestAbsent covers the fs.ReadFile error branch of LoadCatalog
// when catalog.yaml is missing from the provided FS (CATALOG_MANIFEST_ABSENT sentinel).
// REQ-CATALOG-001-026.
func TestLoadCatalog_ManifestAbsent(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{} // empty FS, no catalog.yaml
	cat, err := LoadCatalog(fsys)
	if err == nil {
		t.Errorf("LoadCatalog(empty FS) returned nil error and cat=%+v; want CATALOG_MANIFEST_ABSENT error", cat)
		return
	}
	if !strings.Contains(err.Error(), "CATALOG_MANIFEST_ABSENT") {
		t.Errorf("LoadCatalog(empty FS) error = %q, want sentinel CATALOG_MANIFEST_ABSENT", err.Error())
	}
}
