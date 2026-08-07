package astx

import (
	"path/filepath"
	"testing"
)

// TestSupportedLanguages_ReturnsSixteenWithScaffolded is the M1 RED test:
// the registration list contains all 16 languages; r and flutter are present.
func TestSupportedLanguages_ReturnsSixteenWithScaffolded(t *testing.T) {
	got := SupportedLanguages()
	if len(got) != 16 {
		t.Fatalf("SupportedLanguages() returned %d languages, want 16: %v", len(got), got)
	}
	want := map[string]bool{}
	for _, n := range []string{
		"go", "python", "typescript", "javascript", "rust",
		"java", "kotlin", "csharp", "ruby", "php", "elixir",
		"cpp", "scala", "swift", "r", "flutter",
	} {
		want[n] = true
	}
	for _, n := range got {
		if !want[n] {
			t.Errorf("unexpected language %q in SupportedLanguages()", n)
		}
	}
	for n := range want {
		found := false
		for _, g := range got {
			if g == n {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected language %q missing from SupportedLanguages()", n)
		}
	}
}

// TestIsScaffolded_RAndFlutter is the M1 RED test for the fail-open set.
func TestIsScaffolded_RAndFlutter(t *testing.T) {
	if !IsScaffolded("r") {
		t.Errorf("IsScaffolded(r) = false, want true")
	}
	if !IsScaffolded("flutter") {
		t.Errorf("IsScaffolded(flutter) = false, want true")
	}
	if IsScaffolded("go") {
		t.Errorf("IsScaffolded(go) = true, want false")
	}
}

// TestDetectLanguage_ByExtension is the M1 RED test for extension detection.
func TestDetectLanguage_ByExtension(t *testing.T) {
	cases := map[string]string{
		"main.go":       "go",
		"app.py":        "python",
		"index.ts":      "typescript",
		"comp.tsx":      "typescript",
		"main.js":       "javascript",
		"lib.rs":        "rust",
		"Main.java":     "java",
		"App.kt":        "kotlin",
		"Program.cs":    "csharp",
		"gem.rb":        "ruby",
		"page.php":      "php",
		"mix.exs":       "elixir",
		"node.cpp":      "cpp",
		"Job.scala":     "scala",
		"View.swift":    "swift",
		"plot.r":        "r",
		"analysis.R":    "r",
		"main.dart":     "flutter",
		"README.md":     "",
		"config.yaml":   "",
		"Dockerfile":    "",
	}
	for file, want := range cases {
		if got := DetectLanguage(file); got != want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", file, got, want)
		}
	}
}

// TestExtract_GoFixture is the M1 RED test: Extract on a Go fixture returns
// Supported: true and captures the expected function + type names.
func TestExtract_GoFixture(t *testing.T) {
	set, err := Extract("go", filepath.Join("testdata", "sample.go"))
	if err != nil {
		t.Fatalf("Extract(go) returned error: %v", err)
	}
	if !set.Supported {
		t.Fatalf("Extract(go) Supported = false, want true (CGO build required for this test)")
	}
	// Expect a type capture for "Greeter".
	hasType := false
	for _, sym := range set.Symbols["type"] {
		if sym.Name == "Greeter" {
			hasType = true
		}
	}
	if !hasType {
		t.Errorf("Extract(go) did not capture type Greeter; symbols=%v", set.Symbols)
	}
	// Expect function captures for "NewGreeter" and the method "Greet".
	hasFunc := false
	for _, sym := range set.Symbols["function"] {
		if sym.Name == "NewGreeter" {
			hasFunc = true
		}
	}
	if !hasFunc {
		t.Errorf("Extract(go) did not capture function NewGreeter; symbols=%v", set.Symbols)
	}
	hasMethod := false
	for _, sym := range set.Symbols["method"] {
		if sym.Name == "Greet" {
			hasMethod = true
		}
	}
	if !hasMethod {
		t.Errorf("Extract(go) did not capture method Greet; symbols=%v", set.Symbols)
	}
}

// TestExtract_ScaffoldedR_ReturnsUnsupported is the M1 RED test for fail-open
// on a scaffolded language.
func TestExtract_ScaffoldedR_ReturnsUnsupported(t *testing.T) {
	set, err := Extract("r", filepath.Join("testdata", "sample.go"))
	if err != nil {
		t.Fatalf("Extract(r) returned error: %v", err)
	}
	if set.Supported {
		t.Errorf("Extract(r) Supported = true, want false (scaffolded)")
	}
}

// TestExtract_UnknownLanguage_ReturnsUnsupported is the M1 RED test for an
// unregistered language.
func TestExtract_UnknownLanguage_ReturnsUnsupported(t *testing.T) {
	set, _ := Extract("klingon", filepath.Join("testdata", "sample.go"))
	if set.Supported {
		t.Errorf("Extract(klingon) Supported = true, want false")
	}
}
