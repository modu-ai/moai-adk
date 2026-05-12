package mx

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// langExtensions maps language name → file extension for 16-language sweep test.
var langExtensions = map[string]string{
	"go":         "go",
	"python":     "py",
	"typescript": "ts",
	"javascript": "js",
	"rust":       "rs",
	"java":       "java",
	"kotlin":     "kt",
	"csharp":     "cs",
	"ruby":       "rb",
	"php":        "php",
	"elixir":     "ex",
	"cpp":        "cpp",
	"scala":      "scala",
	"r":          "r",
	"flutter":    "dart",
	"swift":      "swift",
}

// commentPrefix returns the line-comment prefix for a given language extension.
func commentPrefix(ext string) string {
	switch ext {
	case "py", "rb", "r":
		return "#"
	case "ex":
		return "#"
	default:
		return "//"
	}
}

// TestResolver_AllSixteenLanguages verifies that the Resolver correctly processes
// tags from all 16 supported languages (AC-SPC-004-15: body-based SPEC association
// across 16 languages).
//
// Setup:
//   - 16 fixture files, one per language, each with 1 NOTE + 1 ANCHOR tag
//   - 8 of those languages get a 2nd file that references the anchor (fan_in = 2)
//   - SPEC-V3R2-SPC-004 is embedded in the tag body for at least 1 fixture
//
// Assertions:
//  1. Resolver returns exactly 16 ANCHOR tags when kind=anchor
//  2. Each ANCHOR has correct language-mapped file extension
//  3. SPEC-V3R2-SPC-004 body association is detected for the tagged fixture
func TestResolver_AllSixteenLanguages(t *testing.T) {
	stateDir := t.TempDir()

	// Build tags: one ANCHOR + one NOTE per language (32 tags total)
	var tags []Tag
	langOrder := []string{
		"go", "python", "typescript", "javascript", "rust", "java",
		"kotlin", "csharp", "ruby", "php", "elixir", "cpp", "scala",
		"r", "flutter", "swift",
	}

	// specTagLang: first language gets SPEC-V3R2-SPC-004 in body for AC-15 body association
	specTagLang := langOrder[0]

	for i, lang := range langOrder {
		ext := langExtensions[lang]
		filePath := fmt.Sprintf("src/%s/main.%s", lang, ext)
		anchorID := fmt.Sprintf("anchor-%s-001", lang)

		body := fmt.Sprintf("[AUTO] anchor for %s handler", lang)
		if lang == specTagLang {
			body = fmt.Sprintf("[AUTO] anchor for %s handler SPEC-V3R2-SPC-004", lang)
		}

		anchor := Tag{
			Kind:       MXAnchor,
			File:       filePath,
			Line:       i + 1,
			Body:       body,
			AnchorID:   anchorID,
			CreatedBy:  "test",
			LastSeenAt: time.Now(),
		}
		note := Tag{
			Kind:       MXNote,
			File:       filePath,
			Line:       i + 100,
			Body:       fmt.Sprintf("note for %s module", lang),
			CreatedBy:  "test",
			LastSeenAt: time.Now(),
		}
		tags = append(tags, anchor, note)
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// Query: kind=anchor → expect exactly 16 results
	result, err := resolver.Resolve(Query{
		Kind: MXAnchor,
	})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(result.Tags) != 16 {
		t.Errorf("expected 16 ANCHOR tags across 16 languages, got %d", len(result.Tags))
	}

	// Verify each ANCHOR has a language-consistent file extension
	extSet := make(map[string]bool)
	for _, tag := range result.Tags {
		ext := filepath.Ext(tag.File)
		if ext == "" {
			t.Errorf("tag file %q has no extension", tag.File)
			continue
		}
		extSet[ext[1:]] = true // strip leading "."
	}

	for _, lang := range langOrder {
		expectedExt := langExtensions[lang]
		if !extSet[expectedExt] {
			t.Errorf("language %s (ext .%s) not represented in ANCHOR results", lang, expectedExt)
		}
	}

	// Verify SPEC-V3R2-SPC-004 body association is detected
	// SpecAssociator with empty modules → falls back to body extraction
	specAssociator := NewSpecAssociator(map[string][]string{})
	specFound := false
	for _, tag := range result.Tags {
		rawTag := Tag{
			Kind: tag.Kind, File: tag.File, Line: tag.Line,
			Body: tag.Body, AnchorID: tag.AnchorID,
		}
		associations := specAssociator.Associate(rawTag)
		for _, a := range associations {
			if a == "SPEC-V3R2-SPC-004" {
				specFound = true
				break
			}
		}
		if specFound {
			break
		}
	}
	if !specFound {
		t.Error("SPEC-V3R2-SPC-004 body association not detected in any of the 16 language ANCHOR tags")
	}
}

// TestResolver_AllSixteenLanguages_FanInReference verifies that 8 out of 16 languages
// have a second file referencing the anchor, giving fan_in >= 1 for those languages.
func TestResolver_AllSixteenLanguages_FanInReference(t *testing.T) {
	// Use separate dirs for projectRoot (source files) and stateDir (sidecar),
	// so the sidecar JSON doesn't appear during textual fan-in walk.
	rawRoot := t.TempDir()
	rawState := t.TempDir()

	// Resolve symlinks (macOS /var → /private/var) to ensure path comparison works
	projectRoot, err := filepath.EvalSymlinks(rawRoot)
	if err != nil {
		projectRoot = rawRoot
	}
	stateDir, err := filepath.EvalSymlinks(rawState)
	if err != nil {
		stateDir = rawState
	}

	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	langOrder := []string{
		"go", "python", "typescript", "javascript", "rust", "java",
		"kotlin", "csharp", "ruby", "php", "elixir", "cpp", "scala",
		"r", "flutter", "swift",
	}

	// Create real fixture files on disk for textual fan-in counting
	var tags []Tag
	// First 8 languages get a reference file
	fanInLangs := langOrder[:8]
	fanInSet := make(map[string]bool, 8)
	for _, l := range fanInLangs {
		fanInSet[l] = true
	}

	for i, lang := range langOrder {
		ext := langExtensions[lang]
		srcDir := filepath.Join(projectRoot, "src", lang)
		if err := os.MkdirAll(srcDir, 0o755); err != nil {
			t.Fatal(err)
		}

		mainFile := filepath.Join(srcDir, "main."+ext)
		prefix := commentPrefix(ext)
		anchorID := fmt.Sprintf("anchor-%s-001", lang)
		content := fmt.Sprintf("%s @MX:ANCHOR: [AUTO] %s handler\n%s @MX:REASON: test\nfunc main() {}\n", prefix, anchorID, prefix)
		if err := os.WriteFile(mainFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create reference file for first 8 languages
		if fanInSet[lang] {
			refFile := filepath.Join(srcDir, "caller."+ext)
			// Use a unique body so only this file references this anchor
			refContent := fmt.Sprintf("// calls %s\nfunc caller() { %s() }\n", anchorID, anchorID)
			if err := os.WriteFile(refFile, []byte(refContent), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		// Use absolute, symlink-resolved path for tag.File so Walk comparison works
		tags = append(tags, Tag{
			Kind:       MXAnchor,
			File:       mainFile, // already under symlink-resolved projectRoot
			Line:       i + 1,
			AnchorID:   anchorID,
			Body:       fmt.Sprintf("[AUTO] anchor for %s", lang),
			CreatedBy:  "test",
			LastSeenAt: time.Now(),
		})
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	counter := &TextualFanInCounter{ProjectRoot: projectRoot}

	// For each anchor, count fan-in
	anchorResult, err := resolver.Resolve(Query{Kind: MXAnchor})
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	fanInCounts := make(map[string]int)
	for _, tag := range anchorResult.Tags {
		for _, lang := range langOrder {
			anchorID := fmt.Sprintf("anchor-%s-001", lang)
			if tag.AnchorID == anchorID {
				rawTag := Tag{
					Kind: tag.Kind, File: tag.File, Line: tag.Line,
					AnchorID: tag.AnchorID, Body: tag.Body,
				}
				count, _, _ := counter.Count(t.Context(), rawTag, projectRoot, false)
				fanInCounts[lang] = count
				break
			}
		}
	}

	// First 8 languages should have fan_in >= 1 (reference file present)
	for _, lang := range fanInLangs {
		if fanInCounts[lang] < 1 {
			t.Errorf("language %s: expected fan_in >= 1, got %d", lang, fanInCounts[lang])
		}
	}

	// Last 8 languages should have fan_in == 0 (no reference file)
	for _, lang := range langOrder[8:] {
		if fanInCounts[lang] != 0 {
			t.Errorf("language %s: expected fan_in == 0, got %d", lang, fanInCounts[lang])
		}
	}
}
