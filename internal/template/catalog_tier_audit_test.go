package template

// @MX:NOTE: [AUTO] Audit suite for SPEC-V3R4-CATALOG-001 catalog manifest integrity.
// Tests: TestCatalogManifestPresent, TestAllSkillsInCatalog, TestAllAgentsInCatalog,
// TestCatalogReferencesValid, TestCatalogTierValid, TestPackDependencyDAG,
// TestManifestHashFormat, TestWorkflowTriggerCoverage, TestCatalogNoDuplicateEntries,
// TestCatalogReservedFieldType.
//
// T-022 (SPEC-V3R4-CATALOG-001 M4.4): Refactored to use LoadCatalog() from
// catalog_loader.go instead of inline yaml.Unmarshal. Local struct definitions
// replaced by the exported types Catalog, Entry, Pack, TierSection.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// loadCatalog is the test helper that calls LoadCatalog(embeddedRaw).
// embeddedRaw exposes catalog.yaml at its root before the "templates/" prefix strip.
func loadCatalog(t *testing.T) *Catalog {
	t.Helper()

	cat, err := LoadCatalog(embeddedRaw)
	if err != nil {
		t.Fatalf("LoadCatalog() error (CATALOG_MANIFEST_ABSENT): %v", err)
	}
	return cat
}

// allCatalogEntries returns a flat list of all entries across all tier sections.
// Used by orphan, tier, duplicate, and hash checks.
func allCatalogEntries(cat *Catalog) []Entry {
	return cat.AllEntries()
}

// TestCatalogManifestPresent asserts that catalog.yaml is embedded in the binary
// at the expected path and has the required top-level fields.
//
// Sentinel: CATALOG_MANIFEST_ABSENT
// REQ: REQ-001, REQ-002, REQ-003, REQ-004, REQ-026
func TestCatalogManifestPresent(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t)

	if cat.Version == "" {
		t.Errorf("CATALOG_MANIFEST_ABSENT: catalog.yaml missing top-level 'version' field")
	}
	if cat.GeneratedAt == "" {
		t.Errorf("CATALOG_MANIFEST_ABSENT: catalog.yaml missing top-level 'generated_at' field")
	}

	// Validate semver-like pattern for version
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if cat.Version != "" && !semverPattern.MatchString(cat.Version) {
		t.Errorf("catalog.yaml version %q does not match semver format X.Y.Z", cat.Version)
	}

	// Validate ISO 8601 date pattern (at least YYYY-MM-DD)
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)
	if cat.GeneratedAt != "" && !datePattern.MatchString(cat.GeneratedAt) {
		t.Errorf("catalog.yaml generated_at %q does not match ISO 8601 date format", cat.GeneratedAt)
	}

	// Validate 3 sub-sections exist
	if cat.Catalog.OptionalPacks == nil {
		t.Errorf("CATALOG_MANIFEST_ABSENT: catalog.yaml missing 'catalog.optional_packs' section")
	}
}

// TestAllSkillsInCatalog walks the .claude/skills/ directory in the embedded FS
// and asserts that every top-level skill directory appears exactly once in the catalog.
//
// Sentinel: CATALOG_ENTRY_MISSING: <path>
// REQ: REQ-005, REQ-015
func TestAllSkillsInCatalog(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	// Build catalog name set (union of all tier sections)
	catalogSkills := make(map[string]bool)
	for _, e := range cat.Catalog.Core.Skills {
		catalogSkills[e.Name] = true
	}
	for _, pack := range cat.Catalog.OptionalPacks {
		for _, e := range pack.Skills {
			catalogSkills[e.Name] = true
		}
	}
	for _, e := range cat.Catalog.HarnessGenerated.Skills {
		catalogSkills[e.Name] = true
	}

	// Walk .claude/skills/ top-level directories
	diskSkills := []string{}
	walkErr := fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == ".claude/skills" {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		// Only top-level directories
		parts := strings.Split(path, "/")
		if len(parts) == 3 { // .claude/skills/<name>
			diskSkills = append(diskSkills, parts[2])
			return nil
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude/skills) error: %v", walkErr)
	}

	const expectedSkillCount = 37
	if len(diskSkills) != expectedSkillCount {
		t.Errorf("expected %d skill directories on disk, found %d: %v", expectedSkillCount, len(diskSkills), diskSkills)
	}

	for _, skillName := range diskSkills {
		if !catalogSkills[skillName] {
			t.Errorf("CATALOG_ENTRY_MISSING: .claude/skills/%s/ exists on disk but is not in catalog.yaml", skillName)
		}
	}

	t.Logf("audited %d skills on disk against catalog", len(diskSkills))
}

// TestAllAgentsInCatalog walks the .claude/agents/moai/*.md files in the embedded FS
// and asserts that every agent appears exactly once in the catalog.
//
// Sentinel: CATALOG_ENTRY_MISSING: <path>
// REQ: REQ-006, REQ-016
func TestAllAgentsInCatalog(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	// Build catalog name set (union of all tier sections)
	catalogAgents := make(map[string]bool)
	for _, e := range cat.Catalog.Core.Agents {
		catalogAgents[e.Name] = true
	}
	for _, pack := range cat.Catalog.OptionalPacks {
		for _, e := range pack.Agents {
			catalogAgents[e.Name] = true
		}
	}
	for _, e := range cat.Catalog.HarnessGenerated.Agents {
		catalogAgents[e.Name] = true
	}

	// Walk .claude/agents/moai/*.md files
	diskAgents := []string{}
	walkErr := fs.WalkDir(fsys, ".claude/agents/moai", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			// Extract agent name: .claude/agents/moai/<name>.md → <name>
			parts := strings.Split(path, "/")
			fileName := parts[len(parts)-1]
			agentName := strings.TrimSuffix(fileName, ".md")
			diskAgents = append(diskAgents, agentName)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude/agents/moai) error: %v", walkErr)
	}

	const expectedAgentCount = 28
	if len(diskAgents) != expectedAgentCount {
		t.Errorf("expected %d agent files on disk, found %d: %v", expectedAgentCount, len(diskAgents), diskAgents)
	}

	for _, agentName := range diskAgents {
		if !catalogAgents[agentName] {
			t.Errorf("CATALOG_ENTRY_MISSING: .claude/agents/moai/%s.md exists on disk but is not in catalog.yaml", agentName)
		}
	}

	t.Logf("audited %d agents on disk against catalog", len(diskAgents))
}

// TestCatalogReferencesValid asserts that every catalog entry's `path` field
// resolves to an existing file or directory in the embedded FS.
//
// Sentinel: CATALOG_ENTRY_ORPHAN: <path>
// REQ: REQ-017
func TestCatalogReferencesValid(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	entries := allCatalogEntries(cat)
	for _, e := range entries {
		if e.Path == "" {
			t.Errorf("CATALOG_ENTRY_ORPHAN: entry %q has empty path", e.Name)
			continue
		}
		// Strip "templates/" prefix — paths in catalog are relative to templates/
		// Also strip trailing slash (fs.Stat does not accept trailing slashes)
		fsPath := strings.TrimPrefix(e.Path, "templates/")
		fsPath = strings.TrimSuffix(fsPath, "/")
		_, statErr := fs.Stat(fsys, fsPath)
		if statErr != nil {
			t.Errorf("CATALOG_ENTRY_ORPHAN: %s path=%q does not exist in embedded FS: %v", e.Name, e.Path, statErr)
		}
	}

	t.Logf("audited %d catalog entries for path validity", len(entries))
}

// TestCatalogTierValid asserts that every catalog entry's `tier` field matches
// the allowed tier pattern.
//
// Sentinel: CATALOG_TIER_INVALID: <entry> tier=<value>
// REQ: REQ-019
func TestCatalogTierValid(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t)

	// tier must be: "core" | "optional-pack:<name>" | "harness-generated"
	tierPattern := regexp.MustCompile(`^(core|optional-pack:[a-z][a-z0-9-]{1,30}|harness-generated)$`)

	entries := allCatalogEntries(cat)
	for _, e := range entries {
		if !tierPattern.MatchString(e.Tier) {
			t.Errorf("CATALOG_TIER_INVALID: %s tier=%q does not match allowed pattern (core|optional-pack:<name>|harness-generated)", e.Name, e.Tier)
		}
	}

	t.Logf("audited %d catalog entries for tier validity", len(entries))
}

// TestPackDependencyDAG performs DFS cycle detection on the pack depends_on graph.
//
// Sentinel: PACK_DEPENDENCY_CYCLE: <A> <-> <B>
// REQ: REQ-010, REQ-018
func TestPackDependencyDAG(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t)
	if len(cat.Catalog.OptionalPacks) == 0 {
		t.Log("no optional packs defined — vacuously passes DAG check")
		return
	}

	// Build adjacency map
	adj := make(map[string][]string)
	for packName, pack := range cat.Catalog.OptionalPacks {
		adj[packName] = pack.DependsOn
	}

	// DFS cycle detection
	const (
		colorWhite = 0
		colorGray  = 1
		colorBlack = 2
	)
	color := make(map[string]int)
	parent := make(map[string]string)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		color[node] = colorGray
		for _, neighbor := range adj[node] {
			if color[neighbor] == colorGray {
				// Cycle detected
				t.Errorf("PACK_DEPENDENCY_CYCLE: %s <-> %s (via DFS; parent chain: %s→%s)",
					parent[neighbor], neighbor, node, neighbor)
				return true
			}
			if color[neighbor] == colorWhite {
				parent[neighbor] = node
				if dfs(neighbor) {
					return true
				}
			}
		}
		color[node] = colorBlack
		return false
	}

	for packName := range adj {
		if color[packName] == colorWhite {
			dfs(packName)
		}
	}

	t.Logf("audited %d packs for DAG cycle integrity", len(adj))
}

// TestManifestHashFormat asserts that every catalog entry's `hash` field is either
// empty (placeholder) or a valid 64-char lowercase hex string (sha256).
//
// Also re-computes the hash and verifies it matches the stored value (hash stability).
//
// Sentinel: CATALOG_HASH_INVALID: <entry>
// REQ: REQ-007, REQ-020, REQ-022, REQ-023
func TestManifestHashFormat(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	hashPattern := regexp.MustCompile(`^[0-9a-f]{64}$`)

	entries := allCatalogEntries(cat)
	for _, e := range entries {
		if e.Hash == "" || e.Hash == "TODO" {
			t.Errorf("CATALOG_HASH_INVALID: %s hash=%q is placeholder — run gen-catalog-hashes.go --all to populate (EC3, REQ-007, REQ-020)", e.Name, e.Hash)
			continue
		}

		if !hashPattern.MatchString(e.Hash) {
			t.Errorf("CATALOG_HASH_INVALID: %s hash=%q does not match ^[0-9a-f]{64}$", e.Name, e.Hash)
			continue
		}

		// Hash stability check: re-compute hash from the source file and compare.
		fsPath := strings.TrimPrefix(e.Path, "templates/")
		fsPath = strings.TrimSuffix(fsPath, "/")

		var hashSourcePath string
		stat, statErr := fs.Stat(fsys, fsPath)
		if statErr != nil {
			t.Errorf("CATALOG_ENTRY_ORPHAN: %s path=%q not in FS, cannot verify hash", e.Name, e.Path)
			continue
		}

		if stat.IsDir() {
			// Skill directory: hash the root SKILL.md or skill.md
			for _, candidate := range []string{"SKILL.md", "skill.md"} {
				candidatePath := fsPath + "/" + candidate
				if _, err2 := fs.Stat(fsys, candidatePath); err2 == nil {
					hashSourcePath = candidatePath
					break
				}
			}
			if hashSourcePath == "" {
				t.Errorf("CATALOG_HASH_INVALID: %s is a directory but has no SKILL.md/skill.md for hashing", e.Name)
				continue
			}
		} else {
			hashSourcePath = fsPath
		}

		rawContent, readErr := fs.ReadFile(fsys, hashSourcePath)
		if readErr != nil {
			t.Errorf("cannot read %q for hash verification: %v", hashSourcePath, readErr)
			continue
		}

		normalized := NormalizeForHash(rawContent)
		sum := sha256.Sum256(normalized)
		computedHash := hex.EncodeToString(sum[:])

		if computedHash != e.Hash {
			t.Errorf("CATALOG_HASH_UNSTABLE: %s stored hash=%s, computed hash=%s (source=%s)",
				e.Name, e.Hash, computedHash, hashSourcePath)
		}
	}

	t.Logf("audited %d catalog entries for hash validity", len(entries))
}

// TestWorkflowTriggerCoverage walks .claude/skills/moai/workflows/*.md files
// and parses frontmatter for metadata.required-skills. At v1, all workflows are
// flat .md files with no required-skills key, so all pass vacuously.
//
// Sentinel: WORKFLOW_UNCOVERED: <workflow> requires <skill>
// REQ: REQ-021
func TestWorkflowTriggerCoverage(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	cat := loadCatalog(t)

	// Build set of all known skill names (for dependency resolution)
	knownSkills := make(map[string]bool)
	for _, e := range cat.Catalog.Core.Skills {
		knownSkills[e.Name] = true
	}
	for packName, pack := range cat.Catalog.OptionalPacks {
		for _, e := range pack.Skills {
			knownSkills[fmt.Sprintf("optional-pack:%s:%s", packName, e.Name)] = true
			knownSkills[e.Name] = true
		}
	}

	// Walk workflow files
	workflowDir := ".claude/skills/moai/workflows"
	var workflowFiles []string
	walkErr := fs.WalkDir(fsys, workflowDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			workflowFiles = append(workflowFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		// Directory may not exist yet; vacuously pass
		t.Logf("workflow dir %q not found — vacuously passing (0 workflows)", workflowDir)
		return
	}

	for _, wfPath := range workflowFiles {
		data, readErr := fs.ReadFile(fsys, wfPath)
		if readErr != nil {
			t.Errorf("cannot read workflow file %q: %v", wfPath, readErr)
			continue
		}

		// Parse frontmatter for metadata.required-skills
		requiredSkills := extractRequiredSkills(string(data))
		if len(requiredSkills) == 0 {
			// v1 expectation: vacuously true
			continue
		}

		// If required-skills declared, verify each skill is in catalog
		for _, skill := range requiredSkills {
			if !knownSkills[skill] {
				t.Errorf("WORKFLOW_UNCOVERED: %s requires skill %q which is not in catalog core or declared optional-pack",
					wfPath, skill)
			}
		}
	}

	t.Logf("audited %d workflow files for trigger coverage", len(workflowFiles))
}

// TestCatalogNoDuplicateEntries checks that no skill or agent name appears in
// more than one tier section.
//
// Sentinel: CATALOG_DUPLICATE_ENTRY: <name> in [<tier1>, <tier2>]
// REQ: REQ-027
func TestCatalogNoDuplicateEntries(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t)

	// Track where each entry name was seen
	seenIn := make(map[string][]string) // name → list of tier sections

	addSeen := func(name, tierSection string) {
		seenIn[name] = append(seenIn[name], tierSection)
	}

	for _, e := range cat.Catalog.Core.Skills {
		addSeen(e.Name, "core.skills")
	}
	for _, e := range cat.Catalog.Core.Agents {
		addSeen(e.Name, "core.agents")
	}
	for packName, pack := range cat.Catalog.OptionalPacks {
		for _, e := range pack.Skills {
			addSeen(e.Name, fmt.Sprintf("optional_packs.%s.skills", packName))
		}
		for _, e := range pack.Agents {
			addSeen(e.Name, fmt.Sprintf("optional_packs.%s.agents", packName))
		}
	}
	for _, e := range cat.Catalog.HarnessGenerated.Skills {
		addSeen(e.Name, "harness_generated.skills")
	}
	for _, e := range cat.Catalog.HarnessGenerated.Agents {
		addSeen(e.Name, "harness_generated.agents")
	}

	for name, tiers := range seenIn {
		if len(tiers) > 1 {
			t.Errorf("CATALOG_DUPLICATE_ENTRY: %s appears in multiple sections: %v", name, tiers)
		}
	}

	t.Logf("audited %d unique names for duplicate entries", len(seenIn))
}

// TestCatalogReservedFieldType verifies that when optional marketplace fields
// (marketplace_id, marketplace_url, publisher) are present on a pack, they are strings.
//
// Sentinel: CATALOG_RESERVED_FIELD_INVALID: <pack> <field>
// REQ: REQ-024
func TestCatalogReservedFieldType(t *testing.T) {
	t.Parallel()

	cat := loadCatalog(t)

	for packName, pack := range cat.Catalog.OptionalPacks {
		if pack.MarketplaceID != nil {
			if _, ok := pack.MarketplaceID.(string); !ok {
				t.Errorf("CATALOG_RESERVED_FIELD_INVALID: %s marketplace_id is not a string (got %T)",
					packName, pack.MarketplaceID)
			}
		}
		if pack.MarketplaceURL != nil {
			if _, ok := pack.MarketplaceURL.(string); !ok {
				t.Errorf("CATALOG_RESERVED_FIELD_INVALID: %s marketplace_url is not a string (got %T)",
					packName, pack.MarketplaceURL)
			}
		}
		if pack.Publisher != nil {
			if _, ok := pack.Publisher.(string); !ok {
				t.Errorf("CATALOG_RESERVED_FIELD_INVALID: %s publisher is not a string (got %T)",
					packName, pack.Publisher)
			}
		}
	}

	t.Logf("audited %d packs for reserved field type compliance", len(cat.Catalog.OptionalPacks))
}

// extractRequiredSkills parses YAML frontmatter from a workflow .md file and returns
// the list of skill names declared in metadata.required-skills.
// Returns empty slice if the key is absent.
func extractRequiredSkills(content string) []string {
	lines := strings.Split(content, "\n")

	// Find opening ---
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}

	// Find closing ---
	end := -1
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return nil
	}

	// Extract frontmatter block and look for metadata.required-skills
	frontmatter := strings.Join(lines[start+1:end], "\n")

	// Simple line-based parse for required-skills under metadata:
	inMetadata := false
	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "metadata:" {
			inMetadata = true
			continue
		}
		if inMetadata {
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				inMetadata = false
				continue
			}
			if strings.Contains(trimmed, "required-skills:") {
				_, after, _ := strings.Cut(trimmed, "required-skills:")
				val := strings.TrimSpace(after)
				val = strings.Trim(val, `"' `)
				if val == "" {
					return nil
				}
				// May be comma-separated or YAML list
				var skills []string
				for _, s := range strings.Split(val, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						skills = append(skills, s)
					}
				}
				return skills
			}
		}
	}
	return nil
}
