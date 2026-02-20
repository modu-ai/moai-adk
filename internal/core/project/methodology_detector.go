package project

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// MethodologyDetector analyzes project test coverage to recommend a development methodology.
type MethodologyDetector interface {
	// DetectMethodology analyzes test coverage and recommends a development mode.
	DetectMethodology(root string, languages []Language) (*MethodologyRecommendation, error)
}

// MethodologyRecommendation provides a recommended development methodology with rationale.
type MethodologyRecommendation struct {
	Recommended      string                   // "ddd" or "tdd".
	Confidence       float64                  // 0.0–1.0.
	Rationale        string                   // Human-readable explanation.
	ProjectType      string                   // "greenfield" or "brownfield".
	TestFileCount    int                      // Number of test files found.
	CodeFileCount    int                      // Number of source code files found.
	CoverageEstimate float64                  // Estimated coverage percentage (0–100).
	Alternatives     []AlternativeMethodology // Non-recommended but available options.
}

// AlternativeMethodology represents a non-recommended but available option.
type AlternativeMethodology struct {
	Mode    string // "ddd" or "tdd".
	Reason  string // Why this is an alternative.
	Warning string // Warning message if chosen despite recommendation.
}

// testPattern defines language-specific test file detection patterns.
type testPattern struct {
	Language string
	// IsTestFile returns true if the filename matches the test pattern for this language.
	IsTestFile func(name string) bool
	// IsCodeFile returns true if the filename is a source code file for this language.
	IsCodeFile func(name string) bool
}

// knownTestPatterns maps language names to their test file patterns.
var knownTestPatterns = []testPattern{
	{
		Language:   "Go",
		IsTestFile: func(name string) bool { return strings.HasSuffix(name, "_test.go") },
		IsCodeFile: func(name string) bool { return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") },
	},
	{
		Language: "Python",
		IsTestFile: func(name string) bool {
			return (strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test.py")) && strings.HasSuffix(name, ".py")
		},
		IsCodeFile: func(name string) bool {
			return strings.HasSuffix(name, ".py") && !strings.HasPrefix(name, "test_") && !strings.HasSuffix(name, "_test.py")
		},
	},
	{
		Language: "TypeScript",
		IsTestFile: func(name string) bool {
			return (strings.HasSuffix(name, ".test.ts") || strings.HasSuffix(name, ".spec.ts") ||
				strings.HasSuffix(name, ".test.tsx") || strings.HasSuffix(name, ".spec.tsx"))
		},
		IsCodeFile: func(name string) bool {
			return (strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".tsx")) &&
				!strings.HasSuffix(name, ".test.ts") && !strings.HasSuffix(name, ".spec.ts") &&
				!strings.HasSuffix(name, ".test.tsx") && !strings.HasSuffix(name, ".spec.tsx") &&
				!strings.HasSuffix(name, ".d.ts")
		},
	},
	{
		Language: "JavaScript",
		IsTestFile: func(name string) bool {
			return (strings.HasSuffix(name, ".test.js") || strings.HasSuffix(name, ".spec.js") ||
				strings.HasSuffix(name, ".test.jsx") || strings.HasSuffix(name, ".spec.jsx"))
		},
		IsCodeFile: func(name string) bool {
			return (strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".jsx")) &&
				!strings.HasSuffix(name, ".test.js") && !strings.HasSuffix(name, ".spec.js") &&
				!strings.HasSuffix(name, ".test.jsx") && !strings.HasSuffix(name, ".spec.jsx")
		},
	},
	{
		Language:   "Java",
		IsTestFile: func(name string) bool { return strings.HasSuffix(name, "Test.java") },
		IsCodeFile: func(name string) bool {
			return strings.HasSuffix(name, ".java") && !strings.HasSuffix(name, "Test.java")
		},
	},
	{
		Language: "Rust",
		IsTestFile: func(name string) bool {
			return strings.HasSuffix(name, ".rs") && (strings.Contains(name, "test") || name == "mod.rs")
		},
		IsCodeFile: func(name string) bool { return strings.HasSuffix(name, ".rs") },
	},
	{
		Language: "Ruby",
		IsTestFile: func(name string) bool {
			return (strings.HasSuffix(name, "_spec.rb") || strings.HasSuffix(name, "_test.rb"))
		},
		IsCodeFile: func(name string) bool {
			return strings.HasSuffix(name, ".rb") && !strings.HasSuffix(name, "_spec.rb") && !strings.HasSuffix(name, "_test.rb")
		},
	},
	{
		Language:   "PHP",
		IsTestFile: func(name string) bool { return strings.HasSuffix(name, "Test.php") },
		IsCodeFile: func(name string) bool { return strings.HasSuffix(name, ".php") && !strings.HasSuffix(name, "Test.php") },
	},
	{
		Language:   "C#",
		IsTestFile: func(name string) bool { return strings.HasSuffix(name, "Tests.cs") },
		IsCodeFile: func(name string) bool { return strings.HasSuffix(name, ".cs") && !strings.HasSuffix(name, "Tests.cs") },
	},
	{
		Language:   "Kotlin",
		IsTestFile: func(name string) bool { return strings.HasSuffix(name, "Test.kt") },
		IsCodeFile: func(name string) bool { return strings.HasSuffix(name, ".kt") && !strings.HasSuffix(name, "Test.kt") },
	},
}

// methodologyDetector is the concrete implementation of MethodologyDetector.
type methodologyDetector struct {
	logger *slog.Logger
}

// NewMethodologyDetector creates a new MethodologyDetector.
func NewMethodologyDetector(logger *slog.Logger) MethodologyDetector {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &methodologyDetector{logger: logger}
}

// DetectMethodology analyzes test coverage and recommends a development mode.
func (d *methodologyDetector) DetectMethodology(root string, languages []Language) (*MethodologyRecommendation, error) {
	root = filepath.Clean(root)
	if err := validateRoot(root); err != nil {
		return nil, fmt.Errorf("detect methodology: %w", err)
	}

	d.logger.Debug("detecting methodology", "root", root, "languages", len(languages))

	// Tier 1: Scan test file patterns
	testFiles, codeFiles := d.scanTestFiles(root, languages)

	d.logger.Debug("file scan complete",
		"test_files", testFiles,
		"code_files", codeFiles,
	)

	// Tier 2: Estimate coverage
	coverageEstimate := estimateCoverage(testFiles, codeFiles)

	// Apply decision tree
	rec := d.applyDecisionTree(testFiles, codeFiles, coverageEstimate)

	d.logger.Debug("methodology recommendation",
		"recommended", rec.Recommended,
		"confidence", rec.Confidence,
		"coverage_estimate", rec.CoverageEstimate,
		"project_type", rec.ProjectType,
	)

	return rec, nil
}

// scanTestFiles counts test files and code files based on detected languages.
func (d *methodologyDetector) scanTestFiles(root string, languages []Language) (testCount, codeCount int) {
	// Build a set of detected language names
	langSet := make(map[string]bool, len(languages))
	for _, lang := range languages {
		langSet[lang.Name] = true
	}

	// Find matching test patterns for detected languages
	var patterns []testPattern
	for _, p := range knownTestPatterns {
		if langSet[p.Language] {
			patterns = append(patterns, p)
		}
	}

	// If no specific patterns matched, use all patterns as fallback
	if len(patterns) == 0 {
		patterns = knownTestPatterns
	}

	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			name := entry.Name()
			if skipDirs[name] || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			return nil
		}

		fileName := entry.Name()
		for _, p := range patterns {
			if p.IsTestFile(fileName) {
				testCount++
				return nil
			}
			if p.IsCodeFile(fileName) {
				codeCount++
				return nil
			}
		}
		return nil
	})

	return testCount, codeCount
}

// estimateCoverage calculates coverage estimate using the conservative formula.
// coverage_estimate = (test_file_count / code_file_count) * 100 * 0.2
func estimateCoverage(testFiles, codeFiles int) float64 {
	if codeFiles == 0 {
		return 0.0
	}
	estimate := (float64(testFiles) / float64(codeFiles)) * 100.0 * 0.2
	if estimate > 100.0 {
		estimate = 100.0
	}
	return estimate
}

// applyDecisionTree applies the methodology decision tree.
func (d *methodologyDetector) applyDecisionTree(testFiles, codeFiles int, coverageEstimate float64) *MethodologyRecommendation {
	rec := &MethodologyRecommendation{
		TestFileCount:    testFiles,
		CodeFileCount:    codeFiles,
		CoverageEstimate: coverageEstimate,
	}

	// Greenfield project: no code files
	if codeFiles == 0 {
		rec.Recommended = "tdd"
		rec.Confidence = 0.85
		rec.Rationale = "Greenfield project with no existing code. TDD is recommended for test-first development."
		rec.ProjectType = "greenfield"
		rec.Alternatives = []AlternativeMethodology{
			{
				Mode:   "ddd",
				Reason: "Use DDD if integrating with existing systems that need behavior preservation.",
			},
		}
		return rec
	}

	// Brownfield project
	rec.ProjectType = "brownfield"

	if coverageEstimate >= 50.0 {
		// High coverage: recommend TDD
		rec.Recommended = "tdd"
		rec.Confidence = 0.85
		rec.Rationale = fmt.Sprintf(
			"Brownfield project with strong test coverage (estimated %.0f%%). TDD is recommended to maintain quality.",
			coverageEstimate,
		)
		rec.Alternatives = []AlternativeMethodology{
			{
				Mode:   "ddd",
				Reason: "Use DDD for legacy modules that need behavior preservation.",
			},
		}
	} else if coverageEstimate >= 10.0 {
		// Partial coverage: recommend TDD
		rec.Recommended = "tdd"
		rec.Confidence = 0.75
		rec.Rationale = fmt.Sprintf(
			"Brownfield project with partial test coverage (estimated %.0f%%). TDD is recommended to expand test coverage with test-first development.",
			coverageEstimate,
		)
		rec.Alternatives = []AlternativeMethodology{
			{
				Mode:   "ddd",
				Reason: "Focus on characterization tests and behavior preservation for all code.",
			},
		}
	} else {
		// Low or no coverage: recommend DDD
		rec.Recommended = "ddd"
		rec.Confidence = 0.9
		rec.Rationale = fmt.Sprintf(
			"Brownfield project with no existing tests (estimated %.0f%% coverage). DDD with characterization tests is strongly recommended.",
			coverageEstimate,
		)
		estimatedTests := codeFiles * 5
		rec.Alternatives = []AlternativeMethodology{
			{
				Mode:   "tdd",
				Reason: "Pure test-first development.",
				Warning: fmt.Sprintf(
					"%d files x 5 tests = %d tests needed. Significant upfront investment required.",
					codeFiles, estimatedTests,
				),
			},
		}
	}

	return rec
}

// isInTestDir checks if a path is within a known test directory.
func isInTestDir(path string) bool {
	parts := strings.SplitSeq(filepath.ToSlash(path), "/")
	for part := range parts {
		lower := strings.ToLower(part)
		if lower == "tests" || lower == "test" || lower == "__tests__" || lower == "spec" {
			return true
		}
	}
	return false
}

// countFilesWithPattern counts files matching a predicate in a directory tree.
func countFilesWithPattern(root string, match func(string) bool) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			name := entry.Name()
			if skipDirs[name] || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if match(entry.Name()) {
			count++
		}
		return nil
	})
	return count
}

// ensureDir creates a directory and all parents if they don't exist.
func ensureDir(path string) error {
	return os.MkdirAll(filepath.Clean(path), 0o755)
}
