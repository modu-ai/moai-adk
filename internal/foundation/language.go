package foundation

import (
	"fmt"
	"sort"
	"strings"
)

// SupportedLanguage represents a programming language identifier.
type SupportedLanguage string

const (
	// LangGo represents the Go programming language.
	LangGo SupportedLanguage = "go"

	// LangPython represents the Python programming language.
	LangPython SupportedLanguage = "python"

	// LangTypeScript represents the TypeScript programming language.
	LangTypeScript SupportedLanguage = "typescript"

	// LangJavaScript represents the JavaScript programming language.
	LangJavaScript SupportedLanguage = "javascript"

	// LangJava represents the Java programming language.
	LangJava SupportedLanguage = "java"

	// LangRust represents the Rust programming language.
	LangRust SupportedLanguage = "rust"

	// LangC represents the C programming language.
	LangC SupportedLanguage = "c"

	// LangCPP represents the C++ programming language.
	LangCPP SupportedLanguage = "cpp"

	// LangRuby represents the Ruby programming language.
	LangRuby SupportedLanguage = "ruby"

	// LangPHP represents the PHP programming language.
	LangPHP SupportedLanguage = "php"

	// LangKotlin represents the Kotlin programming language.
	LangKotlin SupportedLanguage = "kotlin"

	// LangSwift represents the Swift programming language.
	LangSwift SupportedLanguage = "swift"

	// LangDart represents the Dart programming language.
	LangDart SupportedLanguage = "dart"

	// LangElixir represents the Elixir programming language.
	LangElixir SupportedLanguage = "elixir"

	// LangScala represents the Scala programming language.
	LangScala SupportedLanguage = "scala"

	// LangHaskell represents the Haskell programming language.
	LangHaskell SupportedLanguage = "haskell"

	// LangZig represents the Zig programming language.
	LangZig SupportedLanguage = "zig"

	// LangR represents the R programming language.
	LangR SupportedLanguage = "r"

	// LangCSharp represents the C# programming language.
	LangCSharp SupportedLanguage = "csharp"

	// LangLua represents the Lua programming language.
	LangLua SupportedLanguage = "lua"

	// LangHTML represents HTML markup language.
	LangHTML SupportedLanguage = "html"

	// LangVue represents Vue.js single-file components.
	LangVue SupportedLanguage = "vue"

	// LangSvelte represents Svelte single-file components.
	LangSvelte SupportedLanguage = "svelte"
)

// String returns the string representation of the SupportedLanguage.
func (l SupportedLanguage) String() string {
	return string(l)
}

// AllSupportedLanguages returns all supported language identifiers.
func AllSupportedLanguages() []SupportedLanguage {
	return []SupportedLanguage{
		LangGo, LangPython, LangTypeScript, LangJavaScript,
		LangJava, LangRust, LangC, LangCPP,
		LangRuby, LangPHP, LangKotlin, LangSwift,
		LangDart, LangElixir, LangScala, LangHaskell,
		LangZig, LangR, LangCSharp, LangLua,
		LangHTML, LangVue, LangSvelte,
	}
}

// LanguageInfo holds metadata about a programming language.
type LanguageInfo struct {
	ID              SupportedLanguage `json:"id"`
	Name            string            `json:"name"`
	Extensions      []string          `json:"extensions"`
	TestPattern     string            `json:"test_pattern"`
	CoverageCommand string            `json:"coverage_command"`
	// AstGrepLang is the language identifier used by ast-grep CLI.
	// If empty, the ID is used as the ast-grep language name.
	// Some languages need special identifiers (e.g., "typescriptreact" for .tsx files).
	AstGrepLang map[string]string `json:"ast_grep_lang,omitempty"`
}

// AstGrepLanguageName returns the ast-grep CLI language identifier for the given file extension.
// If no special mapping exists for the extension, returns the language ID as a string.
func (l *LanguageInfo) AstGrepLanguageName(ext string) string {
	if l.AstGrepLang != nil {
		lower := strings.ToLower(ext)
		if name, ok := l.AstGrepLang[lower]; ok {
			return name
		}
	}
	return string(l.ID)
}

// LanguageRegistry provides lookup for supported programming languages.
type LanguageRegistry struct {
	languages map[SupportedLanguage]*LanguageInfo
	extIndex  map[string]SupportedLanguage
}

// languages holds the pre-populated language definitions.
var defaultLanguages = []*LanguageInfo{
	{
		ID:              LangGo,
		Name:            "Go",
		Extensions:      []string{".go"},
		TestPattern:     "go test ./...",
		CoverageCommand: "go test -cover ./...",
	},
	{
		ID:              LangPython,
		Name:            "Python",
		Extensions:      []string{".py", ".pyi"},
		TestPattern:     "pytest",
		CoverageCommand: "pytest --cov",
	},
	{
		ID:              LangTypeScript,
		Name:            "TypeScript",
		Extensions:      []string{".ts", ".tsx", ".mts", ".cts"},
		TestPattern:     "vitest",
		CoverageCommand: "vitest --coverage",
		AstGrepLang: map[string]string{
			".tsx": "typescriptreact",
		},
	},
	{
		ID:              LangJavaScript,
		Name:            "JavaScript",
		Extensions:      []string{".js", ".jsx", ".mjs", ".cjs"},
		TestPattern:     "vitest",
		CoverageCommand: "vitest --coverage",
		AstGrepLang: map[string]string{
			".jsx": "javascriptreact",
		},
	},
	{
		ID:              LangJava,
		Name:            "Java",
		Extensions:      []string{".java"},
		TestPattern:     "mvn test",
		CoverageCommand: "mvn test jacoco:report",
	},
	{
		ID:              LangRust,
		Name:            "Rust",
		Extensions:      []string{".rs"},
		TestPattern:     "cargo test",
		CoverageCommand: "cargo tarpaulin",
	},
	{
		ID:              LangC,
		Name:            "C",
		Extensions:      []string{".c", ".h"},
		TestPattern:     "ctest",
		CoverageCommand: "gcov",
	},
	{
		ID:              LangCPP,
		Name:            "C++",
		Extensions:      []string{".cpp", ".hpp", ".cc", ".cxx"},
		TestPattern:     "ctest",
		CoverageCommand: "gcov",
	},
	{
		ID:              LangRuby,
		Name:            "Ruby",
		Extensions:      []string{".rb"},
		TestPattern:     "rspec",
		CoverageCommand: "rspec --format documentation",
	},
	{
		ID:              LangPHP,
		Name:            "PHP",
		Extensions:      []string{".php"},
		TestPattern:     "phpunit",
		CoverageCommand: "phpunit --coverage-text",
	},
	{
		ID:              LangKotlin,
		Name:            "Kotlin",
		Extensions:      []string{".kt", ".kts"},
		TestPattern:     "gradle test",
		CoverageCommand: "gradle test jacocoTestReport",
	},
	{
		ID:              LangSwift,
		Name:            "Swift",
		Extensions:      []string{".swift"},
		TestPattern:     "swift test",
		CoverageCommand: "swift test --enable-code-coverage",
	},
	{
		ID:              LangDart,
		Name:            "Dart",
		Extensions:      []string{".dart"},
		TestPattern:     "dart test",
		CoverageCommand: "dart test --coverage",
	},
	{
		ID:              LangElixir,
		Name:            "Elixir",
		Extensions:      []string{".ex", ".exs"},
		TestPattern:     "mix test",
		CoverageCommand: "mix test --cover",
	},
	{
		ID:              LangScala,
		Name:            "Scala",
		Extensions:      []string{".scala", ".sc"},
		TestPattern:     "sbt test",
		CoverageCommand: "sbt coverage test coverageReport",
	},
	{
		ID:              LangHaskell,
		Name:            "Haskell",
		Extensions:      []string{".hs"},
		TestPattern:     "cabal test",
		CoverageCommand: "cabal test --enable-coverage",
	},
	{
		ID:              LangZig,
		Name:            "Zig",
		Extensions:      []string{".zig"},
		TestPattern:     "zig test",
		CoverageCommand: "zig test",
	},
	{
		ID:              LangR,
		Name:            "R",
		Extensions:      []string{".R", ".r", ".Rmd"},
		TestPattern:     "testthat",
		CoverageCommand: "covr::package_coverage()",
	},
	{
		ID:              LangCSharp,
		Name:            "C#",
		Extensions:      []string{".cs"},
		TestPattern:     "dotnet test",
		CoverageCommand: "dotnet test --collect:\"XPlat Code Coverage\"",
	},
	{
		ID:              LangLua,
		Name:            "Lua",
		Extensions:      []string{".lua"},
		TestPattern:     "busted",
		CoverageCommand: "busted --coverage",
	},
	{
		ID:              LangHTML,
		Name:            "HTML",
		Extensions:      []string{".html", ".htm"},
		TestPattern:     "",
		CoverageCommand: "",
	},
	{
		ID:              LangVue,
		Name:            "Vue",
		Extensions:      []string{".vue"},
		TestPattern:     "vitest",
		CoverageCommand: "vitest --coverage",
	},
	{
		ID:              LangSvelte,
		Name:            "Svelte",
		Extensions:      []string{".svelte"},
		TestPattern:     "vitest",
		CoverageCommand: "vitest --coverage",
	},
}

// DefaultRegistry is the pre-populated language registry.
var DefaultRegistry *LanguageRegistry

func init() {
	DefaultRegistry = NewLanguageRegistry()
}

// NewLanguageRegistry creates a new LanguageRegistry populated with all supported languages.
func NewLanguageRegistry() *LanguageRegistry {
	r := &LanguageRegistry{
		languages: make(map[SupportedLanguage]*LanguageInfo, len(defaultLanguages)),
		extIndex:  make(map[string]SupportedLanguage),
	}

	for _, lang := range defaultLanguages {
		r.languages[lang.ID] = lang
		for _, ext := range lang.Extensions {
			lower := strings.ToLower(ext)
			// First language registered for an extension wins.
			if _, exists := r.extIndex[lower]; !exists {
				r.extIndex[lower] = lang.ID
			}
		}
	}

	return r
}

// Get retrieves a LanguageInfo by its SupportedLanguage identifier.
// Returns ErrUnsupportedLanguage if the language is not found.
func (r *LanguageRegistry) Get(lang SupportedLanguage) (*LanguageInfo, error) {
	if lang == "" {
		return nil, &LanguageNotFoundError{Query: string(lang)}
	}
	info, ok := r.languages[lang]
	if !ok {
		return nil, &LanguageNotFoundError{Query: string(lang)}
	}
	return info, nil
}

// All returns all registered LanguageInfo entries sorted by ID.
func (r *LanguageRegistry) All() []*LanguageInfo {
	result := make([]*LanguageInfo, 0, len(r.languages))
	for _, info := range r.languages {
		result = append(result, info)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// ByExtension looks up a LanguageInfo by file extension.
// The extension should include the leading dot (e.g., ".go").
// Matching is case-insensitive.
// Returns ErrUnsupportedLanguage if the extension is not recognized.
func (r *LanguageRegistry) ByExtension(ext string) (*LanguageInfo, error) {
	if ext == "" {
		return nil, fmt.Errorf("%w: empty extension", ErrUnsupportedLanguage)
	}
	if !strings.HasPrefix(ext, ".") {
		return nil, fmt.Errorf("%w: extension must start with dot: %s", ErrUnsupportedLanguage, ext)
	}
	if ext == "." {
		return nil, fmt.Errorf("%w: extension cannot be just a dot", ErrUnsupportedLanguage)
	}

	lower := strings.ToLower(ext)
	langID, ok := r.extIndex[lower]
	if !ok {
		return nil, &LanguageNotFoundError{Query: ext}
	}
	return r.languages[langID], nil
}

// SupportedExtensions returns all registered file extensions sorted alphabetically.
// Each extension appears only once even if multiple languages share it.
func (r *LanguageRegistry) SupportedExtensions() []string {
	seen := make(map[string]bool, len(r.extIndex))
	result := make([]string, 0, len(r.extIndex))
	for ext := range r.extIndex {
		if !seen[ext] {
			seen[ext] = true
			result = append(result, ext)
		}
	}
	sort.Strings(result)
	return result
}
