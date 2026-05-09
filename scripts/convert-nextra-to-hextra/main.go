// convert-nextra-to-hextra: Bulk Nextra MDX → Hextra Markdown converter
// SPEC-DOCS-SITE-001 Phase 3
//
// Conversion specification:
//   T1 — Callout JSX → Hextra shortcode (735 items)
//   T2 — _meta.ts → _meta.yaml (38 files)
//   T3 — YAML frontmatter injection (219 pages)
//   T4 — .mdx → .md extension change
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// MetaEntry: single item from _meta.ts
type MetaEntry struct {
	Key     string
	Title   string
	Type    string
	Display string
	Weight  int
}

// Stats: conversion statistics
type Stats struct {
	FilesProcessed  int
	CalloutsChanged int
	MetaConverted   int
	FrontmatterAdded int
	FilesRenamed    int
	Errors          []string
}

var (
	// nextra import pattern (multiple variations)
	reImportCallout = regexp.MustCompile(`(?m)^import\s*\{[^}]*Callout[^}]*\}\s*from\s*["']nextra/components["'];?\s*\n?`)
	reImportMeta    = regexp.MustCompile(`(?m)^import\s+type\s*\{[^}]*MetaRecord[^}]*\}\s*from\s*["']nextra["'];?\s*\n?`)

	// <Callout type="..."> pattern — tag start mapping
	reCalloutOpen  = regexp.MustCompile(`(?m)<Callout(\s+type\s*=\s*"([^"]*)")?\s*/?>`)
	reCalloutClose = regexp.MustCompile(`</Callout>`)

	reH1 = regexp.MustCompile(`(?m)^#\s+(.+)$`)

	reMetaKeyStr = regexp.MustCompile(`^\s*"([^"]+)"\s*:\s*"([^"]+)"\s*,?\s*$`)
	reMetaKeyBare = regexp.MustCompile(`^\s*([a-zA-Z][a-zA-Z0-9_-]*)\s*:\s*"([^"]+)"\s*,?\s*$`)
	reObjTitle   = regexp.MustCompile(`title\s*:\s*"([^"]+)"`)
	reObjDisplay = regexp.MustCompile(`display\s*:\s*"([^"]+)"`)
	reObjType    = regexp.MustCompile(`\btype\s*:\s*"([^"]+)"`)
	// Object start line: "key": { or key: {
	reObjStart = regexp.MustCompile(`^\s*"?([a-zA-Z][a-zA-Z0-9_-]*)"?\s*:\s*\{`)
)

func main() {
	dryRun := flag.Bool("dry-run", false, "simulation mode — print statistics without file modification")
	contentDir := flag.String("content", "docs-site/content", "target content directory path")
	flag.Parse()

	// go run ./scripts/... execution location
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "project root detection failed: %v\n", err)
		os.Exit(1)
	}
	absContentDir := filepath.Join(root, *contentDir)

	fmt.Printf("conversion start: %s\n", absContentDir)
	if *dryRun {
		fmt.Println("[DRY-RUN mode] will not modify files")
	}

	stats := &Stats{}

	// Phase 1: _meta.ts parsing (build weight map) + T2 conversion
	// Process per locale (first _meta.ts at each locale root, then by section)
	locales := []string{"ko", "en", "ja", "zh"}

	// (locale + dir) → []MetaEntry order-preserving map
	weightMaps := make(map[string]map[string]int) // key: locale+"/"+relDir, val: filename→weight

	for _, locale := range locales {
		localeDir := filepath.Join(absContentDir, locale)
		err := filepath.Walk(localeDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || info.Name() != "_meta.ts" {
				return nil
			}
			dir := filepath.Dir(path)
			relDir, _ := filepath.Rel(absContentDir, dir)

			entries, parseErr := parseMetaTS(path)
			if parseErr != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("_meta.ts parsing failure [%s]: %v", path, parseErr))
				return nil
			}

			// Register weight map
			wmap := make(map[string]int)
			for _, e := range entries {
				wmap[e.Key] = e.Weight
			}
			weightMaps[relDir] = wmap

			// T2: _meta.yaml creation
			if !*dryRun {
				yamlPath := filepath.Join(dir, "_meta.yaml")
				if writeErr := writeMetaYAML(entries, yamlPath); writeErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("_meta.yaml write failure [%s]: %v", yamlPath, writeErr))
					return nil
				}
				// meta.ts deletion
				if rmErr := os.Remove(path); rmErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("_meta.ts deletion failure [%s]: %v", path, rmErr))
				}
			}
			stats.MetaConverted++
			return nil
		})
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("locale Walk error [%s]: %v", locale, err))
		}
	}

	// Phase 2: .mdx file process (T1, T3, T4)
	for _, locale := range locales {
		localeDir := filepath.Join(absContentDir, locale)
		err := filepath.Walk(localeDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() || filepath.Ext(path) != ".mdx" {
				return nil
			}

			// Query weight map based on file's directory
			dir := filepath.Dir(path)
			relDir, _ := filepath.Rel(absContentDir, dir)
			wmap := weightMaps[relDir]

			// Filename → weight calculation
			baseName := strings.TrimSuffix(filepath.Base(path), ".mdx")
			weight := 99
			if wmap != nil {
				if w, ok := wmap[baseName]; ok {
					weight = w
				}
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("read failure [%s]: %v", path, readErr))
				return nil
			}
			original := string(content)

			// T1: import remove + Callout conversion
			converted, calloutCount := convertContent(original)
			stats.CalloutsChanged += calloutCount

			// T3: frontmatter injection
			title := extractH1Title(converted)
			if title == "" {
				title = toTitleCase(baseName)
			}
			if !hasFrontmatter(converted) {
				converted = buildFrontmatter(title, weight) + converted
				stats.FrontmatterAdded++
			}

			stats.FilesProcessed++

			if !*dryRun {
				if writeErr := os.WriteFile(path, []byte(converted), 0644); writeErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("write failure [%s]: %v", path, writeErr))
					return nil
				}

				// T4: .mdx → .md rename
				newPath := strings.TrimSuffix(path, ".mdx") + ".md"
				if renameErr := os.Rename(path, newPath); renameErr != nil {
					stats.Errors = append(stats.Errors, fmt.Sprintf("rename failure [%s]: %v", path, renameErr))
					return nil
				}
				stats.FilesRenamed++
			}

			return nil
		})
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("mdx Walk error [%s]: %v", locale, err))
		}
	}

	printReport(stats, *dryRun)

	if len(stats.Errors) > 0 {
		os.Exit(1)
	}
}

// findProjectRoot finds the root directory containing go.mod file
func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// fallback: cwd use
	return cwd, nil
}

// parseMetaTS: Parse _meta.ts file and return MetaEntry slice
// Extracts top-level object body by counting brace depth
func parseMetaTS(path string) ([]MetaEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)

	// Find export object location: const meta = { ... }; or export default { ... };
	// Safely process nested objects by counting brace depth
	body, extractErr := extractTopLevelObjectBody(content)
	if extractErr != nil {
		return nil, fmt.Errorf("cannot find meta object [%s]: %w", path, extractErr)
	}

	return parseMetaBody(body), nil
}

// extractTopLevelObjectBody: extracts the body (content) of the top-level export object from TypeScript source
// Safely processes nested objects by counting brace depth
func extractTopLevelObjectBody(content string) (string, error) {
	// Find first { after "const meta" or "export default"
	reObjEntryPoint := regexp.MustCompile(`(?:const\s+meta\s*(?::\s*\w+)?\s*=\s*|export\s+default\s+)\{`)
	loc := reObjEntryPoint.FindStringIndex(content)
	if loc == nil {
		return "", fmt.Errorf("cannot find export object")
	}

	startIdx := loc[1] - 1 // { position
	depth := 0
	for i := startIdx; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[startIdx+1 : i], nil
			}
		}
	}
	return "", fmt.Errorf("cannot find object closing brace")
}

// parseMetaBody: parses meta object body and returns order-preserved MetaEntry slice
func parseMetaBody(body string) []MetaEntry {
	var entries []MetaEntry
	weight := 10

	scanner := bufio.NewScanner(strings.NewReader(body))
	var objLines []string
	var objKey string
	inObj := false
	braceDepth := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		if inObj {
			objLines = append(objLines, line)
			for _, ch := range line {
				switch ch {
				case '{':
					braceDepth++
				case '}':
					braceDepth--
				}
			}
			if braceDepth <= 0 {
				objContent := strings.Join(objLines, "\n")
				entry := parseObjectEntry(objKey, objContent)
				if entry != nil {
					entry.Weight = weight
					weight += 10
					entries = append(entries, *entry)
				}
				inObj = false
				objLines = nil
				objKey = ""
			}
			continue
		}

		// string value pattern: "key": "value"
		if m := reMetaKeyStr.FindStringSubmatch(trimmed); m != nil {
			entries = append(entries, MetaEntry{Key: m[1], Title: m[2], Weight: weight})
			weight += 10
			continue
		}
		// bare key pattern: key: "value"
		if m := reMetaKeyBare.FindStringSubmatch(trimmed); m != nil {
			entries = append(entries, MetaEntry{Key: m[1], Title: m[2], Weight: weight})
			weight += 10
			continue
		}
		// Object start detection (single-line complete: "key": { ... } or multiline start: "key": {)
		if mo := reObjStart.FindStringSubmatch(trimmed); mo != nil {
			objKey = mo[1]
			inObj = true
			for _, ch := range trimmed {
				switch ch {
				case '{':
					braceDepth++
				case '}':
					braceDepth--
				}
			}
			objLines = []string{trimmed}
			if braceDepth <= 0 {
				objContent := strings.Join(objLines, "\n")
				entry := parseObjectEntry(objKey, objContent)
				if entry != nil {
					entry.Weight = weight
					weight += 10
					entries = append(entries, *entry)
				}
				inObj = false
				objLines = nil
				objKey = ""
				braceDepth = 0
			}
			continue
		}
	}

	return entries
}

// parseObjectEntry: extracts title/display/type from object value
func parseObjectEntry(key, objContent string) *MetaEntry {
	key = strings.Trim(key, `"`)
	e := &MetaEntry{Key: key}
	if m := reObjTitle.FindStringSubmatch(objContent); m != nil {
		e.Title = m[1]
	} else {
		e.Title = toTitleCase(key)
	}
	if m := reObjDisplay.FindStringSubmatch(objContent); m != nil {
		e.Display = m[1]
	}
	if m := reObjType.FindStringSubmatch(objContent); m != nil {
		e.Type = m[1]
	}
	return e
}

// writeMetaYAML: creates _meta.yaml file
func writeMetaYAML(entries []MetaEntry, path string) error {
	var sb strings.Builder
	for _, e := range entries {
		key := e.Key
		if strings.Contains(key, "-") {
			key = fmt.Sprintf(`"%s"`, key)
		}
		fmt.Fprintf(&sb, "%s:\n", key)
		fmt.Fprintf(&sb, "  title: %q\n", e.Title)
		if e.Display != "" {
			fmt.Fprintf(&sb, "  display: %q\n", e.Display)
		}
		if e.Type != "" {
			fmt.Fprintf(&sb, "  type: %q\n", e.Type)
		}
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// convertContent: removes Nextra imports + converts Callout JSX to Hextra shortcode
func convertContent(content string) (string, int) {
	// T1-1: remove import lines
	content = reImportCallout.ReplaceAllString(content, "")
	content = reImportMeta.ReplaceAllString(content, "")

	// T1-2: <Callout type="..."> single tag (self-closing or multiple tags)
	calloutCount := 0

	// Tag conversion: <Callout type="info"> → {{< callout type="info" >}}
	// <Callout> (no type) → {{< callout >}}
	content = reCalloutOpen.ReplaceAllStringFunc(content, func(match string) string {
		calloutCount++
		sub := reCalloutOpen.FindStringSubmatch(match)
		typeVal := ""
		if len(sub) >= 3 && sub[2] != "" {
			rawType := sub[2]
			typeVal = normalizeCalloutType(rawType)
		}
		if typeVal == "" {
			return "{{< callout >}}"
		}
		return fmt.Sprintf(`{{< callout type="%s" >}}`, typeVal)
	})

	content = reCalloutClose.ReplaceAllString(content, "{{< /callout >}}")

	reTripleBlank := regexp.MustCompile(`\n{3,}`)
	content = reTripleBlank.ReplaceAllString(content, "\n\n")

	return content, calloutCount
}

// Hextra default support: info, warning, error (default type is no type attribute)
func normalizeCalloutType(t string) string {
	switch t {
	case "info":
		return "info"
	case "warning":
		return "warning"
	case "error":
		return "error"
	case "tip":
		// tip → info (tip is not a separate type in Hextra, map to info)
		return "info"
	case "success":
		// success → info fallback
		return "info"
	default:
		// other types → default (no type attribute)
		return ""
	}
}

// extractH1Title: extracts first H1 header text from content
func extractH1Title(content string) string {
	m := reH1.FindStringSubmatch(content)
	if m == nil {
		return ""
	}
	title := m[1]
	title = regexp.MustCompile(`\*+`).ReplaceAllString(title, "")
	title = regexp.MustCompile("`"+"[^`]+`").ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
	return title
}

// hasFrontmatter: verifies if file already has frontmatter
func hasFrontmatter(content string) bool {
	trimmed := strings.TrimSpace(content)
	return strings.HasPrefix(trimmed, "---")
}

// buildFrontmatter: creates YAML frontmatter string
func buildFrontmatter(title string, weight int) string {
	titleStr := title
	if strings.ContainsAny(titleStr, `:"{}[]|>&*!`) {
		titleStr = fmt.Sprintf(`"%s"`, strings.ReplaceAll(titleStr, `"`, `\"`))
	}
	return fmt.Sprintf("---\ntitle: %s\nweight: %d\ndraft: false\n---\n", titleStr, weight)
}

// toTitleCase: converts kebab-case to Title Case
func toTitleCase(s string) string {
	words := strings.Split(s, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	result := strings.Join(words, " ")
	if result == "Index" {
		result = "Overview"
	}
	return result
}

func printReport(stats *Stats, dryRun bool) {
	mode := "execution"
	if dryRun {
		mode = "DRY-RUN"
	}
	fmt.Printf("\n========== conversion result [%s] ==========\n", mode)
	fmt.Printf("MDX files processed:    %d\n", stats.FilesProcessed)
	fmt.Printf("Callout conversions:       %d\n", stats.CalloutsChanged)
	fmt.Printf("_meta.ts conversions:      %d\n", stats.MetaConverted)
	fmt.Printf("Frontmatter injections:   %d\n", stats.FrontmatterAdded)
	if !dryRun {
		fmt.Printf("file name changes (.md):  %d\n", stats.FilesRenamed)
	}

	if stats.CalloutsChanged != 735 {
		fmt.Printf("\n[warning] Callout conversion count differs from expected (735): %d\n", stats.CalloutsChanged)
	} else {
		fmt.Printf("\n[OK] Callout 735 items converted exactly\n")
	}
	if stats.MetaConverted != 38 {
		fmt.Printf("[warning] _meta.ts conversion count differs from expected (38): %d\n", stats.MetaConverted)
	} else {
		fmt.Printf("[OK] _meta.ts 38 files converted\n")
	}
	if stats.FilesProcessed != 219 {
		fmt.Printf("[warning] processed file count differs from expected (219): %d\n", stats.FilesProcessed)
	} else {
		fmt.Printf("[OK] MDX 219 pages processed\n")
	}

	if len(stats.Errors) > 0 {
		fmt.Printf("\nerror list (%d items):\n", len(stats.Errors))
		sort.Strings(stats.Errors)
		for _, e := range stats.Errors {
			fmt.Printf("  - %s\n", e)
		}
	} else {
		fmt.Println("\n[OK] 0 errors")
	}
}
