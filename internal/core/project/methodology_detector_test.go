package project

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectMethodology(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(t *testing.T, root string)
		languages       []Language
		wantRecommended string
		wantProjectType string
		wantMinConf     float64
		wantMinCoverage float64
		wantMaxCoverage float64
		wantHasWarning  bool
	}{
		{
			name:            "greenfield project recommends tdd",
			setup:           func(t *testing.T, root string) { t.Helper() },
			languages:       []Language{{Name: "Go", Confidence: 1.0, FileCount: 0}},
			wantRecommended: "tdd",
			wantProjectType: "greenfield",
			wantMinConf:     0.7,
		},
		{
			name: "brownfield no tests recommends ddd",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "go.mod", "module test\n")
				for _, f := range []string{
					"cmd/main.go", "internal/handler/user.go",
					"internal/handler/auth.go", "internal/service/user.go",
					"internal/service/auth.go", "internal/repo/db.go",
				} {
					writeFile(t, root, f, "package x\n")
				}
			},
			languages:       []Language{{Name: "Go", Confidence: 1.0, FileCount: 6}},
			wantRecommended: "ddd",
			wantProjectType: "brownfield",
			wantMinConf:     0.8,
			wantMaxCoverage: 10.0,
			wantHasWarning:  true,
		},
		{
			name: "brownfield high coverage recommends tdd",
			setup: func(t *testing.T, root string) {
				t.Helper()
				// Create 10 source files and 15 test files -> ratio 1.5 * 100 * 0.2 = 30
				// Need ratio >= 2.5 for coverage >= 50%
				// 10 code + 13 tests -> 13/10 * 100 * 0.2 = 26 (not enough)
				// Need test/code >= 2.5 for 50%
				// Let's do 4 code files and 10 test files -> 10/4 * 100 * 0.2 = 50
				for i := range 4 {
					writeFile(t, root, "internal/"+string(rune('a'+i))+".go", "package x\n")
				}
				for i := range 10 {
					writeFile(t, root, "internal/"+string(rune('a'+i))+"_test.go", "package x\n")
				}
			},
			languages:       []Language{{Name: "Go", Confidence: 1.0, FileCount: 14}},
			wantRecommended: "tdd",
			wantProjectType: "brownfield",
			wantMinConf:     0.8,
			wantMinCoverage: 50.0,
		},
		{
			name: "brownfield partial tests recommends tdd",
			setup: func(t *testing.T, root string) {
				t.Helper()
				// Create 20 source files and 5 test files
				// coverage = 5/20 * 100 * 0.2 = 5.0 -> < 10, would be ddd
				// Need 10 <= coverage < 50: test/code * 100 * 0.2 >= 10
				// test/code >= 0.5 -> test >= 10 for 20 code
				for i := range 20 {
					writeFile(t, root, "pkg/"+string(rune('a'+i%26))+string(rune('0'+i/26))+".go", "package x\n")
				}
				for i := range 12 {
					writeFile(t, root, "pkg/"+string(rune('a'+i%26))+"_test.go", "package x\n")
				}
			},
			languages:       []Language{{Name: "Go", Confidence: 1.0, FileCount: 32}},
			wantRecommended: "tdd",
			wantProjectType: "brownfield",
			wantMinConf:     0.7,
			wantMinCoverage: 10.0,
			wantMaxCoverage: 50.0,
		},
		{
			name: "python project no tests recommends ddd",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "app/main.py", "print('hello')\n")
				writeFile(t, root, "app/routes.py", "# routes\n")
				writeFile(t, root, "app/models.py", "# models\n")
			},
			languages:       []Language{{Name: "Python", Confidence: 1.0, FileCount: 3}},
			wantRecommended: "ddd",
			wantProjectType: "brownfield",
			wantMinConf:     0.8,
		},
		{
			name: "python project with test files recommends based on coverage",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "app/main.py", "print('hello')\n")
				writeFile(t, root, "app/routes.py", "# routes\n")
				writeFile(t, root, "tests/test_main.py", "def test_main(): pass\n")
				writeFile(t, root, "tests/test_routes.py", "def test_routes(): pass\n")
			},
			languages:       []Language{{Name: "Python", Confidence: 1.0, FileCount: 4}},
			wantRecommended: "tdd",
			wantProjectType: "brownfield",
			wantMinCoverage: 10.0,
		},
		{
			name: "typescript project with spec files",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "src/app.ts", "export const x = 1;\n")
				writeFile(t, root, "src/app.spec.ts", "test('x', () => {});\n")
			},
			languages:       []Language{{Name: "TypeScript", Confidence: 1.0, FileCount: 2}},
			wantRecommended: "tdd",
			wantProjectType: "brownfield",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			md := NewMethodologyDetector(nil)
			rec, err := md.DetectMethodology(root, tt.languages)
			if err != nil {
				t.Fatalf("DetectMethodology() error = %v", err)
			}

			if rec.Recommended != tt.wantRecommended {
				t.Errorf("Recommended = %q, want %q", rec.Recommended, tt.wantRecommended)
			}

			if rec.ProjectType != tt.wantProjectType {
				t.Errorf("ProjectType = %q, want %q", rec.ProjectType, tt.wantProjectType)
			}

			if tt.wantMinConf > 0 && rec.Confidence < tt.wantMinConf {
				t.Errorf("Confidence = %f, want >= %f", rec.Confidence, tt.wantMinConf)
			}

			if tt.wantMinCoverage > 0 && rec.CoverageEstimate < tt.wantMinCoverage {
				t.Errorf("CoverageEstimate = %f, want >= %f", rec.CoverageEstimate, tt.wantMinCoverage)
			}

			if tt.wantMaxCoverage > 0 && rec.CoverageEstimate > tt.wantMaxCoverage {
				t.Errorf("CoverageEstimate = %f, want <= %f", rec.CoverageEstimate, tt.wantMaxCoverage)
			}

			if tt.wantHasWarning {
				hasWarning := false
				for _, alt := range rec.Alternatives {
					if alt.Warning != "" {
						hasWarning = true
						break
					}
				}
				if !hasWarning {
					t.Error("expected at least one alternative with a warning")
				}
			}

			// Verify alternatives are populated
			if len(rec.Alternatives) == 0 {
				t.Error("expected at least one alternative methodology")
			}

			// Verify rationale is not empty
			if rec.Rationale == "" {
				t.Error("expected non-empty rationale")
			}
		})
	}
}

func TestDetectMethodology_InvalidRoot(t *testing.T) {
	md := NewMethodologyDetector(nil)
	_, err := md.DetectMethodology("/nonexistent/path/xyz", nil)
	if err == nil {
		t.Fatal("expected error for invalid root")
	}
}

func TestIsInTestDir(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"tests/test_main.py", true},
		{"src/tests/test_main.py", true},
		{"test/main_test.go", true},
		{"__tests__/app.test.js", true},
		{"spec/models/user_spec.rb", true},
		{"src/main.go", false},
		{"internal/handler.go", false},
		{"pkg/util.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isInTestDir(tt.path)
			if got != tt.want {
				t.Errorf("isInTestDir(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCountFilesWithPattern(t *testing.T) {
	root := t.TempDir()

	// Create some Go files and test files
	writeFile(t, root, "main.go", "package main\n")
	writeFile(t, root, "app.go", "package main\n")
	writeFile(t, root, "app_test.go", "package main\n")
	writeFile(t, root, "internal/handler.go", "package internal\n")
	writeFile(t, root, "internal/handler_test.go", "package internal\n")

	// Count Go test files
	testCount := countFilesWithPattern(root, func(name string) bool {
		return strings.HasSuffix(name, "_test.go")
	})
	if testCount != 2 {
		t.Errorf("test file count = %d, want 2", testCount)
	}

	// Count all .go files
	goCount := countFilesWithPattern(root, func(name string) bool {
		return strings.HasSuffix(name, ".go")
	})
	if goCount != 5 {
		t.Errorf("go file count = %d, want 5", goCount)
	}

	// Count with no matches
	zeroCount := countFilesWithPattern(root, func(name string) bool {
		return strings.HasSuffix(name, ".rs")
	})
	if zeroCount != 0 {
		t.Errorf("rust file count = %d, want 0", zeroCount)
	}
}

func TestCountFilesWithPattern_SkipsDirs(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "main.go", "package main\n")
	writeFile(t, root, "node_modules/dep/index.js", "//js\n")
	writeFile(t, root, "vendor/dep/dep.go", "package dep\n")

	goCount := countFilesWithPattern(root, func(name string) bool {
		return strings.HasSuffix(name, ".go")
	})
	if goCount != 1 {
		t.Errorf("go file count = %d, want 1 (should skip vendor)", goCount)
	}
}

func TestEnsureDir(t *testing.T) {
	root := t.TempDir()

	// Create a new directory
	newDir := filepath.Join(root, "a", "b", "c")
	if err := ensureDir(newDir); err != nil {
		t.Fatalf("ensureDir(%q) error = %v", newDir, err)
	}

	if !dirExists(newDir) {
		t.Errorf("directory %q should exist after ensureDir", newDir)
	}

	// Calling again should be idempotent
	if err := ensureDir(newDir); err != nil {
		t.Fatalf("ensureDir(%q) second call error = %v", newDir, err)
	}
}

func TestDetectMethodology_EmptyLanguages(t *testing.T) {
	root := t.TempDir()

	// Create files that don't match any known language
	writeFile(t, root, "data.csv", "a,b,c\n")
	writeFile(t, root, "readme.txt", "hello\n")

	md := NewMethodologyDetector(nil)
	rec, err := md.DetectMethodology(root, nil)
	if err != nil {
		t.Fatalf("DetectMethodology() error = %v", err)
	}

	// Should use all patterns as fallback and detect as greenfield
	if rec.Recommended == "" {
		t.Error("expected non-empty recommendation")
	}
}

func TestEstimateCoverage(t *testing.T) {
	tests := []struct {
		name      string
		testFiles int
		codeFiles int
		want      float64
	}{
		{"no code files", 0, 0, 0.0},
		{"no test files", 0, 10, 0.0},
		{"equal files", 10, 10, 20.0},  // 10/10 * 100 * 0.2 = 20
		{"many tests", 100, 10, 100.0}, // capped at 100
		{"few tests", 1, 100, 0.2},     // 1/100 * 100 * 0.2 = 0.2
		{"half tests", 5, 10, 10.0},    // 5/10 * 100 * 0.2 = 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateCoverage(tt.testFiles, tt.codeFiles)
			if got != tt.want {
				t.Errorf("estimateCoverage(%d, %d) = %f, want %f", tt.testFiles, tt.codeFiles, got, tt.want)
			}
		})
	}
}
