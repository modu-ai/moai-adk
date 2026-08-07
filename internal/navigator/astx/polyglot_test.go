package astx

import (
	"path/filepath"
	"testing"
)

// polyglotCases is the AC-NT-004 fixture: one source file per language for
// each of the 14 working languages, each declaring a known function/method
// and a type. Each row asserts Supported == true AND that both expected
// names appear somewhere in the captured symbol set.
var polyglotCases = []struct {
	lang     string
	file     string
	wantFunc string // a name expected in the "function" or "method" group
	wantType string // a name expected in the "type" group
}{
	{"go", "sample.go", "NewGreeter", "Greeter"},
	{"python", "polyglot/sample.py", "build", "Widget"},
	{"typescript", "polyglot/sample.ts", "area", "Circle"},
	{"javascript", "polyglot/sample.js", "make", "Box"},
	{"rust", "polyglot/sample.rs", "origin", "Point"},
	{"java", "polyglot/sample.java", "start", "Engine"},
	{"kotlin", "polyglot/sample.kt", "main", "KotlinWidget"},
	{"csharp", "polyglot/sample.cs", "Run", "Service"},
	{"ruby", "polyglot/sample.rb", "total", "Basket"},
	{"php", "polyglot/sample.php", "helper", "Repo"},
	{"elixir", "polyglot/sample.ex", "run", "Worker"},
	{"cpp", "polyglot/sample.cpp", "add", "List"},
	{"scala", "polyglot/sample.scala", "init", "Config"},
	{"swift", "polyglot/sample.swift", "render", "Canvas"},
}

func TestPolyglot_AllFourteenGrammarsExtract(t *testing.T) {
	for _, tc := range polyglotCases {
		t.Run(tc.lang, func(t *testing.T) {
			path := tc.file
			// The Go fixture lives at testdata/sample.go (not under polyglot/).
			set, err := Extract(tc.lang, filepath.Join("testdata", path))
			if err != nil {
				t.Fatalf("Extract(%s) error: %v", tc.lang, err)
			}
			if !set.Supported {
				t.Fatalf("Extract(%s) Supported=false, want true", tc.lang)
			}
			// Check the expected function/method name across both groups.
			if !hasName(set, "function", tc.wantFunc) && !hasName(set, "method", tc.wantFunc) {
				t.Errorf("Extract(%s): %q not found in function/method; symbols=%v",
					tc.lang, tc.wantFunc, set.Symbols)
			}
			// Check the expected type name (some languages put classes under "type").
			if !hasName(set, "type", tc.wantType) {
				t.Errorf("Extract(%s): type %q not found; symbols=%v",
					tc.lang, tc.wantType, set.Symbols)
			}
		})
	}
}

// TestPolyglot_RAndFlutterFailOpen is AC-NT-005: r/flutter return
// Supported: false and the other rows still extract normally.
func TestPolyglot_RAndFlutterFailOpen(t *testing.T) {
	for _, lang := range []string{"r", "flutter"} {
		set, err := Extract(lang, filepath.Join("testdata", "polyglot", "sample.py"))
		if err != nil {
			t.Fatalf("Extract(%s) error: %v", lang, err)
		}
		if set.Supported {
			t.Errorf("Extract(%s) Supported=true, want false (scaffolded)", lang)
		}
	}
}

func hasName(set SymbolSet, kind, name string) bool {
	for _, sym := range set.Symbols[kind] {
		if sym.Name == name {
			return true
		}
	}
	return false
}
