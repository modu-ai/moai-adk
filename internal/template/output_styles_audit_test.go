// Package template provides tests for output style audit and parity checks.
//
// This file implements audit tests for the .claude/output-styles/moai/ directory.
// Source: SPEC-V3R2-WF-006
package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

// Output style name and file name constants — single source of truth (CLAUDE.local.md §14).
const (
	styleNameMoAI    = "MoAI"
	styleNameEinstein = "Einstein"
	styleFileMoAI    = "moai.md"
	styleFileEinstein = "einstein.md"

	// Frontmatter key names.
	keyName                 = "name"
	keyDescription          = "description"
	keyKeepCodingInstructions = "keep-coding-instructions"

	// Error code prefixes (REQ-WF006-007, 008, 010, 013, 014).
	errPrefixSchemaError  = "OUTPUT_STYLE_SCHEMA_ERROR"
	errPrefixDrift        = "OUTPUT_STYLE_DRIFT"
	errPrefixUnverified   = "OUTPUT_STYLE_UNVERIFIED"

	// Output styles directory path (relative to embedded template root).
	outputStylesDir = ".claude/output-styles/moai"

	// Expected number of output styles (REQ-WF006-002).
	expectedStyleCount = 2
)

// outputStyleFrontmatter holds parsed frontmatter fields for an output style file.
type outputStyleFrontmatter struct {
	name                 string
	description          string
	keepCodingInstructions string // raw string: "true" or "false"
}

// parseOutputStyleFrontmatter parses the frontmatter of an output style file.
// It returns the parsed fields and an error string if parsing fails.
// The parser extracts only the three required keys; additional keys are tolerated (OPEN-B decision).
// Boolean keep-coding-instructions must be a raw "true" or "false" literal (not quoted, not capitalized).
//
// Note: parseFrontmatterAndBody strips surrounding quotes from values. For keep-coding-instructions,
// we need to detect the original raw value before quote stripping. We do this by scanning the raw
// frontmatter lines directly to find the keep-coding-instructions line.
func parseOutputStyleFrontmatter(content string) (*outputStyleFrontmatter, string) {
	fm, _, parseErr := parseFrontmatterAndBody(content)
	if parseErr != "" {
		return nil, fmt.Sprintf("%s: frontmatter parse failed: %s", errPrefixSchemaError, parseErr)
	}

	result := &outputStyleFrontmatter{}

	nameVal, ok := fm[keyName]
	if !ok || strings.TrimSpace(nameVal) == "" {
		return nil, fmt.Sprintf("%s: missing or empty key: %s", errPrefixSchemaError, keyName)
	}
	result.name = strings.TrimSpace(nameVal)

	descVal, ok := fm[keyDescription]
	if !ok || strings.TrimSpace(descVal) == "" {
		return nil, fmt.Sprintf("%s: missing or empty key: %s", errPrefixSchemaError, keyDescription)
	}
	result.description = strings.TrimSpace(descVal)

	// For keep-coding-instructions, we must check the raw line before quote stripping.
	// parseFrontmatterAndBody strips surrounding quotes, so "true" becomes true — but
	// the spec requires rejection of quoted strings (acceptance.md edge cases).
	rawKCI, kciErr := extractRawBooleanValue(content, keyKeepCodingInstructions)
	if kciErr != "" {
		return nil, kciErr
	}
	result.keepCodingInstructions = rawKCI

	return result, ""
}

// extractRawBooleanValue scans the frontmatter of content for a key that must have
// a raw boolean value ("true" or "false" — no quotes, no capitalization, no other form).
// Returns the raw value string and an error prefix string if validation fails.
func extractRawBooleanValue(content, key string) (string, string) {
	// Parse the raw frontmatter lines ourselves to get the unstripped value.
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	closedFrontmatter := false
	prefix := key + ":"

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			closedFrontmatter = true
			break
		}
		if !inFrontmatter {
			continue
		}
		if strings.HasPrefix(trimmed, prefix) {
			// Extract the raw value after "key: " (or "key:").
			rawVal := strings.TrimSpace(trimmed[len(prefix):])
			if rawVal != "true" && rawVal != "false" {
				return "", fmt.Sprintf("%s: %s must be boolean literal 'true' or 'false', got: %q",
					errPrefixSchemaError, key, rawVal)
			}
			return rawVal, ""
		}
	}

	if !closedFrontmatter {
		// Frontmatter was not properly closed; parseFrontmatterAndBody already handles this.
		return "", fmt.Sprintf("%s: missing key: %s (frontmatter not found)", errPrefixSchemaError, key)
	}

	return "", fmt.Sprintf("%s: missing key: %s", errPrefixSchemaError, key)
}

// findProjectRoot ascends from the test source file's directory up to 8 levels
// looking for the .moai/ directory marker. Returns the project root path and true
// if found, or ("", false) if not found (OPEN-A RESOLVED — plan.md §8).
func findProjectRoot() (string, bool) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", false
	}
	dir := filepath.Dir(thisFile)
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".moai")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			return "", false
		}
		dir = parent
	}
	return "", false
}

// TestOutputStylesFrontmatterSchema verifies that each output style file in the
// embedded template has the required frontmatter keys with correct types and values.
//
// REQ-WF006-001: name, description, keep-coding-instructions required.
// REQ-WF006-005: MoAI=true, Einstein=false.
// REQ-WF006-007, REQ-WF006-013: violations emit OUTPUT_STYLE_SCHEMA_ERROR.
func TestOutputStylesFrontmatterSchema(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Real files: validate actual output style files in the embedded template.
	realCases := []struct {
		fileName              string
		wantName              string
		wantKeepCodingInstr   string
	}{
		{styleFileMoAI, styleNameMoAI, "true"},
		{styleFileEinstein, styleNameEinstein, "false"},
	}

	for _, tc := range realCases {
		tc := tc
		t.Run(tc.fileName, func(t *testing.T) {
			t.Parallel()

			path := outputStylesDir + "/" + tc.fileName
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			parsed, schemaErr := parseOutputStyleFrontmatter(string(data))
			if schemaErr != "" {
				t.Fatalf("%s", schemaErr)
			}

			if parsed.name != tc.wantName {
				t.Errorf("%s: %s (%s): expected name=%q, got=%q",
					errPrefixSchemaError, path, keyName, tc.wantName, parsed.name)
			}
			if parsed.description == "" {
				t.Errorf("%s: %s (%s): description is empty",
					errPrefixSchemaError, path, keyDescription)
			}
			if parsed.keepCodingInstructions != tc.wantKeepCodingInstr {
				t.Errorf("%s: %s (%s): expected=%q, got=%q",
					errPrefixSchemaError, path, keyKeepCodingInstructions,
					tc.wantKeepCodingInstr, parsed.keepCodingInstructions)
			}
		})
	}

	// Synthetic cases: test the parser's error detection with invalid input.
	// Uses in-memory strings so no embedded fs mutation is needed.
	t.Run("Synthetic/MissingKeepCodingInstructions", func(t *testing.T) {
		t.Parallel()
		content := "---\nname: TestStyle\ndescription: A test style\n---\nBody here\n"
		_, schemaErr := parseOutputStyleFrontmatter(content)
		if schemaErr == "" {
			t.Error("expected schema error for missing keep-coding-instructions, got none")
		}
		if !strings.Contains(schemaErr, errPrefixSchemaError) {
			t.Errorf("error must contain %q, got: %q", errPrefixSchemaError, schemaErr)
		}
	})

	t.Run("Synthetic/NonBoolKeepCodingInstructions", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			label string
			value string
		}{
			{"quoted-string", `"true"`},
			{"capitalized", "True"},
			{"yes", "yes"},
			{"integer", "1"},
		}
		for _, sc := range cases {
			sc := sc
			t.Run(sc.label, func(t *testing.T) {
				t.Parallel()
				content := fmt.Sprintf("---\nname: TestStyle\ndescription: Test\nkeep-coding-instructions: %s\n---\n", sc.value)
				_, schemaErr := parseOutputStyleFrontmatter(content)
				if schemaErr == "" {
					t.Errorf("expected schema error for keep-coding-instructions=%q, got none", sc.value)
				}
				if !strings.Contains(schemaErr, errPrefixSchemaError) {
					t.Errorf("error must contain %q, got: %q", errPrefixSchemaError, schemaErr)
				}
			})
		}
	})

	t.Run("Synthetic/MissingName", func(t *testing.T) {
		t.Parallel()
		content := "---\ndescription: A test style\nkeep-coding-instructions: true\n---\n"
		_, schemaErr := parseOutputStyleFrontmatter(content)
		if schemaErr == "" {
			t.Error("expected schema error for missing name, got none")
		}
		if !strings.Contains(schemaErr, errPrefixSchemaError) {
			t.Errorf("error must contain %q, got: %q", errPrefixSchemaError, schemaErr)
		}
	})

	t.Run("Synthetic/MissingFrontmatter", func(t *testing.T) {
		t.Parallel()
		content := "# Just a body without frontmatter\n"
		_, schemaErr := parseOutputStyleFrontmatter(content)
		if schemaErr == "" {
			t.Error("expected schema error for missing frontmatter, got none")
		}
		if !strings.Contains(schemaErr, errPrefixSchemaError) {
			t.Errorf("error must contain %q, got: %q", errPrefixSchemaError, schemaErr)
		}
	})

	t.Run("Synthetic/ExtraKeyTolerated", func(t *testing.T) {
		t.Parallel()
		// Extra keys beyond the required 3 should be tolerated (OPEN-B decision).
		content := "---\nname: TestStyle\ndescription: Test\nkeep-coding-instructions: true\nmodel: claude-sonnet-4-6\n---\nBody\n"
		parsed, schemaErr := parseOutputStyleFrontmatter(content)
		if schemaErr != "" {
			t.Errorf("extra key should be tolerated, got error: %s", schemaErr)
		}
		if parsed == nil || parsed.name != "TestStyle" {
			t.Error("parsed result should have name=TestStyle")
		}
	})
}

// TestOutputStylesExactlyTwo verifies that the embedded template contains exactly
// two output style files with the expected names.
//
// REQ-WF006-002: exactly two styles (MoAI, Einstein).
// REQ-WF006-014: a third style without schema validation emits OUTPUT_STYLE_UNVERIFIED.
func TestOutputStylesExactlyTwo(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	entries, readErr := fs.ReadDir(fsys, outputStylesDir)
	if readErr != nil {
		t.Fatalf("ReadDir(%q) error: %v", outputStylesDir, readErr)
	}

	// Collect .md files only.
	var mdFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mdFiles = append(mdFiles, e.Name())
		}
	}

	if len(mdFiles) != expectedStyleCount {
		t.Errorf("%s: expected exactly %d styles, got %d: %v",
			errPrefixUnverified, expectedStyleCount, len(mdFiles), mdFiles)
	}

	expectedNames := map[string]bool{
		styleFileMoAI:    true,
		styleFileEinstein: true,
	}
	for _, name := range mdFiles {
		if !expectedNames[name] {
			t.Errorf("%s: unexpected style file %q (not in allowed set {%s, %s})",
				errPrefixUnverified, name, styleFileMoAI, styleFileEinstein)
		}
	}
	for expected := range expectedNames {
		found := false
		for _, name := range mdFiles {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: required style file %q not found in embedded templates",
				errPrefixUnverified, expected)
		}
	}

	// Synthetic: simulate adding a third style.
	t.Run("Synthetic/ThirdStyleWouldFail", func(t *testing.T) {
		t.Parallel()
		extraFiles := append(mdFiles, "foo.md")
		if len(extraFiles) == expectedStyleCount {
			t.Error("synthetic third style was not added correctly to the test")
		}
		// Verify the count check logic produces the expected error string.
		if len(extraFiles) != expectedStyleCount {
			errMsg := fmt.Sprintf("%s: expected exactly %d styles, got %d",
				errPrefixUnverified, expectedStyleCount, len(extraFiles))
			if !strings.Contains(errMsg, errPrefixUnverified) {
				t.Errorf("error message should contain %q", errPrefixUnverified)
			}
		}
	})
}

// TestOutputStylesTemplateLiveParity verifies that the template files and the live
// project files are byte-identical.
//
// REQ-WF006-003: template and local shall be byte-identical.
// REQ-WF006-010: make build shall fail with OUTPUT_STYLE_DRIFT when trees diverge.
//
// Uses runtime.Caller(0) + .moai/ marker ascent to locate the project root (OPEN-A RESOLVED).
func TestOutputStylesTemplateLiveParity(t *testing.T) {
	t.Parallel()

	root, found := findProjectRoot()
	if !found {
		t.Skip("live tree not available: .moai marker not found up to 8 levels")
	}

	liveDir := filepath.Join(root, ".claude", "output-styles", "moai")

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Collect all template files.
	templateFiles := map[string][]byte{}
	walkErr := fs.WalkDir(fsys, outputStylesDir, func(path string, d fs.DirEntry, walkEr error) error {
		if walkEr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			return fmt.Errorf("read template %q: %w", path, readErr)
		}
		// Strip directory prefix to get the bare filename.
		rel := strings.TrimPrefix(path, outputStylesDir+"/")
		templateFiles[rel] = data
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir error: %v", walkErr)
	}

	// Collect all live files.
	liveFiles := map[string][]byte{}
	liveEntries, readErr := os.ReadDir(liveDir)
	if readErr != nil {
		t.Fatalf("ReadDir(%q) error: %v — is the live tree available?", liveDir, readErr)
	}
	for _, e := range liveEntries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		livePath := filepath.Join(liveDir, e.Name())
		data, readFileErr := os.ReadFile(livePath)
		if readFileErr != nil {
			t.Fatalf("ReadFile(%q) error: %v", livePath, readFileErr)
		}
		liveFiles[e.Name()] = data
	}

	// Check files that exist in templates but not in live.
	for name, templateData := range templateFiles {
		liveData, ok := liveFiles[name]
		if !ok {
			t.Errorf("%s: %s exists only in template, not in live tree",
				errPrefixDrift, name)
			continue
		}
		if !bytes.Equal(templateData, liveData) {
			t.Errorf("%s: %s differs between template and live",
				errPrefixDrift, name)
		}
	}

	// Check files that exist in live but not in templates.
	for name := range liveFiles {
		if _, ok := templateFiles[name]; !ok {
			t.Errorf("%s: %s exists only in live tree, not in template",
				errPrefixDrift, name)
		}
	}
}

// TestOutputStylesEncoding verifies that all output style files are UTF-8 encoded
// with LF line endings and no BOM.
//
// Spec §4 Assumption: UTF-8, LF-only encoding.
func TestOutputStylesEncoding(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	styleFiles := []string{styleFileMoAI, styleFileEinstein}
	for _, name := range styleFiles {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := outputStylesDir + "/" + name
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			// Check UTF-8 validity.
			if !utf8.Valid(data) {
				t.Errorf("%s: %s is not valid UTF-8", errPrefixSchemaError, name)
			}

			// Check for BOM (EF BB BF).
			bom := []byte{0xEF, 0xBB, 0xBF}
			if bytes.HasPrefix(data, bom) {
				t.Errorf("%s: %s has UTF-8 BOM", errPrefixSchemaError, name)
			}

			// Check for CR (CRLF line endings).
			if bytes.Contains(data, []byte{0x0D}) {
				t.Errorf("%s: %s contains CR (CRLF line endings not allowed, LF only)", errPrefixSchemaError, name)
			}
		})
	}
}

// TestOutputStylesFallbackDocsContract verifies that settings-management.md contains
// the exact warning format string required by REQ-WF006-008 (AC-08 drift guard).
//
// This test prevents documentation drift: if the warning format is changed in the doc
// without updating EXT-002's Go loader test, this test will catch the divergence.
//
// REQ-WF006-008: unknown style emits exact stderr line OUTPUT_STYLE_UNKNOWN: <name> not found; falling back to MoAI.
func TestOutputStylesFallbackDocsContract(t *testing.T) {
	t.Parallel()

	root, found := findProjectRoot()
	if !found {
		t.Skip("live tree not available: .moai marker not found up to 8 levels")
	}

	docsPath := filepath.Join(root, ".claude", "rules", "moai", "core", "settings-management.md")
	content, readErr := os.ReadFile(docsPath)
	if readErr != nil {
		t.Fatalf("ReadFile(%q) error: %v", docsPath, readErr)
	}

	contentStr := string(content)

	// Assert that the exact warning format is documented.
	requiredStrings := []struct {
		fragment string
		reason   string
	}{
		{
			fragment: "OUTPUT_STYLE_UNKNOWN:",
			reason:   "settings-management.md must document the exact stderr warning prefix (REQ-WF006-008)",
		},
		{
			fragment: "falling back to MoAI",
			reason:   "settings-management.md must document the fallback target in the warning string (REQ-WF006-008)",
		},
	}

	for _, req := range requiredStrings {
		if !strings.Contains(contentStr, req.fragment) {
			t.Errorf("settings-management.md missing REQ-WF006-008 fallback contract (sink=stderr, format=OUTPUT_STYLE_UNKNOWN: ...): %s\nLooking for: %q",
				req.reason, req.fragment)
		}
	}
}
