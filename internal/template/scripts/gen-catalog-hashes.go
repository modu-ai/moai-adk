// gen-catalog-hashes.go — Offline helper to compute sha256 hashes for catalog.yaml entries.
//
// Usage:
//
//	go run internal/template/scripts/gen-catalog-hashes.go [flags]
//
// Flags:
//
//	--all           Update every entry's hash field in catalog.yaml in-place
//	--entry NAME    Update a single named entry
//	--catalog PATH  Path to catalog.yaml (default: internal/template/catalog.yaml)
//	--templates DIR Path to templates directory (default: internal/template/templates)
//	--dry-run       Print computed hashes without modifying catalog.yaml
//
// Hash normalization (HARD — must match NormalizeForHash in catalog_hash_norm.go):
//  1. Read raw bytes of the target file
//  2. Convert CRLF → LF (uniform line endings)
//  3. Strip trailing whitespace from each line
//  4. Ensure exactly one trailing newline
//  5. sha256 the result, hex-encode → 64-char lowercase hex
//
// For skill entries: hash only the root SKILL.md or skill.md (not sub-files).
// For agent entries: hash the *.md file at the given path.
//
// SPEC-V3R4-CATALOG-001 T-007 (M2.4)
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// catalogYAML mirrors catalog.yaml for YAML round-trip with comment preservation.
// Using yaml.v3 node-based approach to preserve existing comments and formatting.
type catalogFile struct {
	Version     string          `yaml:"version"`
	GeneratedAt string          `yaml:"generated_at"`
	Catalog     catalogSections `yaml:"catalog"`
}

type catalogSections struct {
	Core             catalogTierSection      `yaml:"core"`
	OptionalPacks    map[string]*catalogPack `yaml:"optional_packs"`
	HarnessGenerated catalogTierSection      `yaml:"harness_generated"`
}

type catalogTierSection struct {
	Skills []catalogEntry `yaml:"skills"`
	Agents []catalogEntry `yaml:"agents"`
}

type catalogEntry struct {
	Name    string `yaml:"name"`
	Tier    string `yaml:"tier"`
	Path    string `yaml:"path"`
	Hash    string `yaml:"hash"`
	Version string `yaml:"version"`
}

type catalogPack struct {
	Description string         `yaml:"description"`
	DependsOn   []string       `yaml:"depends_on"`
	Skills      []catalogEntry `yaml:"skills"`
	Agents      []catalogEntry `yaml:"agents"`
}

// normalizeForHash normalizes raw file content for reproducible sha256 hashing.
// This is a local copy of template.NormalizeForHash to avoid import cycles
// (this script is in a separate package under scripts/).
// MUST remain byte-identical to catalog_hash_norm.go NormalizeForHash.
func normalizeForHash(raw []byte) []byte {
	// Step 1: Normalize line endings CRLF → LF, lone CR → LF
	normalized := bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

	// Step 2: Strip trailing whitespace from each line
	lines := strings.Split(string(normalized), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	// Step 3: Join and ensure exactly one trailing newline
	content := strings.Join(lines, "\n")
	content = strings.TrimRight(content, "\n")
	content += "\n"

	return []byte(content)
}

// computeHash computes the sha256 hash of a file at the given path,
// after applying normalizeForHash normalization.
func computeHash(filePath string) (string, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", filePath, err)
	}
	normalized := normalizeForHash(raw)
	sum := sha256.Sum256(normalized)
	return hex.EncodeToString(sum[:]), nil
}

// resolveHashSourcePath finds the file to hash for a given catalog entry path.
// For skill directories: look for root SKILL.md or skill.md.
// For agent .md files: use directly.
func resolveHashSourcePath(templatesDir, entryPath string) (string, error) {
	// entryPath is relative to templates/: e.g. "templates/.claude/skills/moai/"
	// Strip "templates/" prefix then join with templatesDir
	relPath := strings.TrimPrefix(entryPath, "templates/")
	relPath = strings.TrimSuffix(relPath, "/")
	absPath := filepath.Join(templatesDir, relPath)

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", absPath, err)
	}

	if info.IsDir() {
		// Skill directory: find root SKILL.md or skill.md
		for _, candidate := range []string{"SKILL.md", "skill.md"} {
			candidatePath := filepath.Join(absPath, candidate)
			if _, statErr := os.Stat(candidatePath); statErr == nil {
				return candidatePath, nil
			}
		}
		return "", fmt.Errorf("directory %q has no SKILL.md or skill.md for hashing", absPath)
	}

	// Agent or other file: hash directly
	return absPath, nil
}

// updateEntryHash computes and sets the hash for a single catalog entry.
func updateEntryHash(entry *catalogEntry, templatesDir string, dryRun bool) error {
	sourcePath, err := resolveHashSourcePath(templatesDir, entry.Path)
	if err != nil {
		return fmt.Errorf("resolve hash source for %q: %w", entry.Name, err)
	}

	hash, err := computeHash(sourcePath)
	if err != nil {
		return fmt.Errorf("compute hash for %q: %w", entry.Name, err)
	}

	if dryRun {
		fmt.Printf("  [dry-run] %s: %s (source: %s)\n", entry.Name, hash, sourcePath)
	} else {
		fmt.Printf("  %s: %s\n", entry.Name, hash)
		entry.Hash = hash
	}
	return nil
}

// allEntries returns pointers to all catalog entry structs across all tier sections.
func allEntries(cat *catalogFile) []*catalogEntry {
	var entries []*catalogEntry

	for i := range cat.Catalog.Core.Skills {
		entries = append(entries, &cat.Catalog.Core.Skills[i])
	}
	for i := range cat.Catalog.Core.Agents {
		entries = append(entries, &cat.Catalog.Core.Agents[i])
	}
	for _, pack := range cat.Catalog.OptionalPacks {
		for i := range pack.Skills {
			entries = append(entries, &pack.Skills[i])
		}
		for i := range pack.Agents {
			entries = append(entries, &pack.Agents[i])
		}
	}
	for i := range cat.Catalog.HarnessGenerated.Skills {
		entries = append(entries, &cat.Catalog.HarnessGenerated.Skills[i])
	}
	for i := range cat.Catalog.HarnessGenerated.Agents {
		entries = append(entries, &cat.Catalog.HarnessGenerated.Agents[i])
	}

	return entries
}

func main() {
	var (
		flagAll          = flag.Bool("all", false, "Update all entries in catalog.yaml")
		flagEntry        = flag.String("entry", "", "Update a single entry by name")
		flagCatalogPath  = flag.String("catalog", "internal/template/catalog.yaml", "Path to catalog.yaml")
		flagTemplatesDir = flag.String("templates", "internal/template/templates", "Path to templates directory")
		flagDryRun       = flag.Bool("dry-run", false, "Print computed hashes without modifying catalog.yaml")
	)
	flag.Parse()

	if !*flagAll && *flagEntry == "" {
		log.Fatal("specify --all or --entry NAME")
	}

	// Read catalog.yaml
	rawYAML, err := os.ReadFile(*flagCatalogPath)
	if err != nil {
		log.Fatalf("read catalog.yaml: %v", err)
	}

	var cat catalogFile
	if err := yaml.Unmarshal(rawYAML, &cat); err != nil {
		log.Fatalf("parse catalog.yaml: %v", err)
	}

	entries := allEntries(&cat)

	if *flagAll {
		fmt.Printf("Computing hashes for all %d entries...\n", len(entries))
		errCount := 0
		for _, e := range entries {
			if updateErr := updateEntryHash(e, *flagTemplatesDir, *flagDryRun); updateErr != nil {
				log.Printf("ERROR: %v", updateErr)
				errCount++
			}
		}
		if errCount > 0 {
			log.Fatalf("%d entries failed hash computation", errCount)
		}
	} else {
		// Single entry
		found := false
		for _, e := range entries {
			if e.Name == *flagEntry {
				fmt.Printf("Computing hash for entry %q...\n", *flagEntry)
				if updateErr := updateEntryHash(e, *flagTemplatesDir, *flagDryRun); updateErr != nil {
					log.Fatalf("ERROR: %v", updateErr)
				}
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("entry %q not found in catalog.yaml", *flagEntry)
		}
	}

	if *flagDryRun {
		fmt.Println("[dry-run] catalog.yaml not modified")
		return
	}

	// Write updated catalog.yaml back
	// Use yaml.v3 Marshal to produce clean YAML output
	// Note: This does NOT preserve comments. For a production tool, use node-based
	// round-trip. For v1, we accept comment loss on --all runs (comments can be
	// re-added manually or via T-025 in M5).
	outYAML, err := yaml.Marshal(&cat)
	if err != nil {
		log.Fatalf("marshal catalog.yaml: %v", err)
	}

	if err := os.WriteFile(*flagCatalogPath, outYAML, 0644); err != nil {
		log.Fatalf("write catalog.yaml: %v", err)
	}

	fmt.Printf("catalog.yaml updated successfully (%d bytes)\n", len(outYAML))
}
