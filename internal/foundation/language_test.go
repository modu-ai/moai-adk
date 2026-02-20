package foundation

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func TestSupportedLanguageString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		lang SupportedLanguage
		want string
	}{
		{name: "Go", lang: LangGo, want: "go"},
		{name: "Python", lang: LangPython, want: "python"},
		{name: "TypeScript", lang: LangTypeScript, want: "typescript"},
		{name: "JavaScript", lang: LangJavaScript, want: "javascript"},
		{name: "Java", lang: LangJava, want: "java"},
		{name: "Rust", lang: LangRust, want: "rust"},
		{name: "C", lang: LangC, want: "c"},
		{name: "CPP", lang: LangCPP, want: "cpp"},
		{name: "Ruby", lang: LangRuby, want: "ruby"},
		{name: "PHP", lang: LangPHP, want: "php"},
		{name: "Kotlin", lang: LangKotlin, want: "kotlin"},
		{name: "Swift", lang: LangSwift, want: "swift"},
		{name: "Dart", lang: LangDart, want: "dart"},
		{name: "Elixir", lang: LangElixir, want: "elixir"},
		{name: "Scala", lang: LangScala, want: "scala"},
		{name: "Haskell", lang: LangHaskell, want: "haskell"},
		{name: "Zig", lang: LangZig, want: "zig"},
		{name: "R", lang: LangR, want: "r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.lang.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAllSupportedLanguages(t *testing.T) {
	t.Parallel()

	langs := AllSupportedLanguages()

	if len(langs) < 20 {
		t.Errorf("AllSupportedLanguages() returned %d, want at least 20", len(langs))
	}

	if len(langs) != 23 {
		t.Errorf("AllSupportedLanguages() returned %d, want exactly 23", len(langs))
	}
}

func TestNewLanguageRegistry(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()
	if r == nil {
		t.Fatal("NewLanguageRegistry() returned nil")
	}
}

func TestDefaultRegistryExists(t *testing.T) {
	t.Parallel()

	if DefaultRegistry == nil {
		t.Fatal("DefaultRegistry should not be nil")
	}
}

func TestRegistryGetAllLanguages(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()

	tests := []struct {
		name       string
		lang       SupportedLanguage
		wantName   string
		wantExtIn  string
		wantTestIn string
	}{
		{name: "Go", lang: LangGo, wantName: "Go", wantExtIn: ".go", wantTestIn: "go test"},
		{name: "Python", lang: LangPython, wantName: "Python", wantExtIn: ".py", wantTestIn: "pytest"},
		{name: "TypeScript", lang: LangTypeScript, wantName: "TypeScript", wantExtIn: ".ts", wantTestIn: "vitest"},
		{name: "JavaScript", lang: LangJavaScript, wantName: "JavaScript", wantExtIn: ".js", wantTestIn: "vitest"},
		{name: "Java", lang: LangJava, wantName: "Java", wantExtIn: ".java", wantTestIn: "mvn test"},
		{name: "Rust", lang: LangRust, wantName: "Rust", wantExtIn: ".rs", wantTestIn: "cargo test"},
		{name: "C", lang: LangC, wantName: "C", wantExtIn: ".c", wantTestIn: "ctest"},
		{name: "CPP", lang: LangCPP, wantName: "C++", wantExtIn: ".cpp", wantTestIn: "ctest"},
		{name: "Ruby", lang: LangRuby, wantName: "Ruby", wantExtIn: ".rb", wantTestIn: "rspec"},
		{name: "PHP", lang: LangPHP, wantName: "PHP", wantExtIn: ".php", wantTestIn: "phpunit"},
		{name: "Kotlin", lang: LangKotlin, wantName: "Kotlin", wantExtIn: ".kt", wantTestIn: "gradle test"},
		{name: "Swift", lang: LangSwift, wantName: "Swift", wantExtIn: ".swift", wantTestIn: "swift test"},
		{name: "Dart", lang: LangDart, wantName: "Dart", wantExtIn: ".dart", wantTestIn: "dart test"},
		{name: "Elixir", lang: LangElixir, wantName: "Elixir", wantExtIn: ".ex", wantTestIn: "mix test"},
		{name: "Scala", lang: LangScala, wantName: "Scala", wantExtIn: ".scala", wantTestIn: "sbt test"},
		{name: "Haskell", lang: LangHaskell, wantName: "Haskell", wantExtIn: ".hs", wantTestIn: "cabal test"},
		{name: "Zig", lang: LangZig, wantName: "Zig", wantExtIn: ".zig", wantTestIn: "zig test"},
		{name: "R", lang: LangR, wantName: "R", wantExtIn: ".R", wantTestIn: "testthat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, err := r.Get(tt.lang)
			if err != nil {
				t.Fatalf("Get(%s) error: %v", tt.lang, err)
			}

			if info.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", info.Name, tt.wantName)
			}

			if !containsString(info.Extensions, tt.wantExtIn) {
				t.Errorf("Extensions %v should contain %q", info.Extensions, tt.wantExtIn)
			}

			if info.TestPattern == "" {
				t.Error("TestPattern should not be empty")
			}

			if info.CoverageCommand == "" {
				t.Error("CoverageCommand should not be empty")
			}
		})
	}
}

func TestRegistryGetUnknown(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()

	tests := []struct {
		name string
		lang SupportedLanguage
	}{
		{name: "brainfuck", lang: SupportedLanguage("brainfuck")},
		{name: "empty", lang: SupportedLanguage("")},
		{name: "nonexistent", lang: SupportedLanguage("nonexistent")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info, err := r.Get(tt.lang)
			if err == nil {
				t.Error("expected error, got nil")
			}
			if info != nil {
				t.Error("info should be nil for unknown language")
			}
		})
	}
}

func TestRegistryAll(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()
	all := r.All()

	if len(all) != 23 {
		t.Errorf("All() returned %d languages, want 23", len(all))
	}

	// Verify sorted by ID.
	for i := 1; i < len(all); i++ {
		if all[i].ID < all[i-1].ID {
			t.Errorf("All() not sorted: %s comes after %s", all[i].ID, all[i-1].ID)
		}
	}

	// Verify all fields are populated.
	// HTML is a markup language and does not have test/coverage commands.
	markupLanguages := map[SupportedLanguage]bool{
		LangHTML: true,
	}

	for _, info := range all {
		if info.Name == "" {
			t.Errorf("language %s has empty Name", info.ID)
		}
		if len(info.Extensions) == 0 {
			t.Errorf("language %s has no Extensions", info.ID)
		}
		// Skip test/coverage checks for markup languages.
		if markupLanguages[info.ID] {
			continue
		}
		if info.TestPattern == "" {
			t.Errorf("language %s has empty TestPattern", info.ID)
		}
		if info.CoverageCommand == "" {
			t.Errorf("language %s has empty CoverageCommand", info.ID)
		}
	}
}

func TestRegistryByExtension(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()

	tests := []struct {
		name   string
		ext    string
		wantID SupportedLanguage
	}{
		{name: "dot_go", ext: ".go", wantID: LangGo},
		{name: "dot_py", ext: ".py", wantID: LangPython},
		{name: "dot_pyi", ext: ".pyi", wantID: LangPython},
		{name: "dot_ts", ext: ".ts", wantID: LangTypeScript},
		{name: "dot_tsx", ext: ".tsx", wantID: LangTypeScript},
		{name: "dot_mts", ext: ".mts", wantID: LangTypeScript},
		{name: "dot_js", ext: ".js", wantID: LangJavaScript},
		{name: "dot_jsx", ext: ".jsx", wantID: LangJavaScript},
		{name: "dot_java", ext: ".java", wantID: LangJava},
		{name: "dot_rs", ext: ".rs", wantID: LangRust},
		{name: "dot_c", ext: ".c", wantID: LangC},
		{name: "dot_cpp", ext: ".cpp", wantID: LangCPP},
		{name: "dot_hpp", ext: ".hpp", wantID: LangCPP},
		{name: "dot_rb", ext: ".rb", wantID: LangRuby},
		{name: "dot_php", ext: ".php", wantID: LangPHP},
		{name: "dot_kt", ext: ".kt", wantID: LangKotlin},
		{name: "dot_kts", ext: ".kts", wantID: LangKotlin},
		{name: "dot_swift", ext: ".swift", wantID: LangSwift},
		{name: "dot_dart", ext: ".dart", wantID: LangDart},
		{name: "dot_ex", ext: ".ex", wantID: LangElixir},
		{name: "dot_exs", ext: ".exs", wantID: LangElixir},
		{name: "dot_scala", ext: ".scala", wantID: LangScala},
		{name: "dot_hs", ext: ".hs", wantID: LangHaskell},
		{name: "dot_zig", ext: ".zig", wantID: LangZig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info, err := r.ByExtension(tt.ext)
			if err != nil {
				t.Fatalf("ByExtension(%q) error: %v", tt.ext, err)
			}
			if info.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", info.ID, tt.wantID)
			}
		})
	}
}

func TestRegistryByExtensionCaseInsensitive(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()

	// R language has both .R and .r extensions.
	tests := []struct {
		name string
		ext  string
	}{
		{name: "uppercase_R", ext: ".R"},
		{name: "lowercase_r", ext: ".r"},
		{name: "uppercase_Rmd", ext: ".Rmd"},
		{name: "lowercase_rmd", ext: ".rmd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info, err := r.ByExtension(tt.ext)
			if err != nil {
				t.Fatalf("ByExtension(%q) error: %v", tt.ext, err)
			}
			if info.ID != LangR {
				t.Errorf("ID = %q, want %q", info.ID, LangR)
			}
		})
	}
}

func TestRegistryByExtensionErrors(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()

	tests := []struct {
		name    string
		ext     string
		wantMsg string
	}{
		{name: "empty_extension", ext: "", wantMsg: "empty extension"},
		{name: "no_dot_prefix", ext: "go", wantMsg: "extension must start with dot"},
		{name: "just_dot", ext: ".", wantMsg: "cannot be just a dot"},
		{name: "unknown_extension", ext: ".xyz", wantMsg: "language not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info, err := r.ByExtension(tt.ext)
			if err == nil {
				t.Error("expected error, got nil")
			}
			if info != nil {
				t.Error("info should be nil for invalid extension")
			}
		})
	}
}

func TestRegistryByExtensionNotPanic(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()

	// These should not panic.
	_, _ = r.ByExtension("")                   //nolint:errcheck // testing panic safety
	_, _ = r.ByExtension(".")                  //nolint:errcheck // testing panic safety
	_, _ = r.ByExtension(".unknown")           //nolint:errcheck // testing panic safety
	_, _ = r.Get(SupportedLanguage(""))        //nolint:errcheck // testing panic safety
	_, _ = r.Get(SupportedLanguage("unknown")) //nolint:errcheck // testing panic safety
}

func TestSupportedExtensions(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()
	exts := r.SupportedExtensions()

	if len(exts) < 30 {
		t.Errorf("SupportedExtensions() returned %d, want at least 30", len(exts))
	}

	// Verify sorted.
	for i := 1; i < len(exts); i++ {
		if exts[i] < exts[i-1] {
			t.Errorf("SupportedExtensions() not sorted: %q comes after %q", exts[i], exts[i-1])
		}
	}

	// Verify no duplicates.
	seen := make(map[string]bool)
	for _, ext := range exts {
		if seen[ext] {
			t.Errorf("duplicate extension: %q", ext)
		}
		seen[ext] = true
	}

	// Verify key extensions are present.
	required := []string{".go", ".py", ".ts", ".rs", ".java"}
	for _, req := range required {
		if !seen[req] {
			t.Errorf("missing required extension: %q", req)
		}
	}
}

func TestRegistryByExtensionUnknownNotUnsupportedError(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()
	_, err := r.ByExtension(".xyz")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var langErr *LanguageNotFoundError
	if !errors.As(err, &langErr) {
		// It may be wrapped with ErrUnsupportedLanguage instead.
		if !errors.Is(err, ErrUnsupportedLanguage) {
			t.Errorf("error should be LanguageNotFoundError or ErrUnsupportedLanguage, got: %T", err)
		}
	}
}

func TestLanguageInfoJSONRoundTrip(t *testing.T) {
	t.Parallel()

	r := NewLanguageRegistry()
	info, err := r.Get(LangGo)
	if err != nil {
		t.Fatalf("Get(LangGo) error: %v", err)
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got LanguageInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got.ID != info.ID {
		t.Errorf("ID = %q, want %q", got.ID, info.ID)
	}
	if got.Name != info.Name {
		t.Errorf("Name = %q, want %q", got.Name, info.Name)
	}
	if len(got.Extensions) != len(info.Extensions) {
		t.Errorf("Extensions length = %d, want %d", len(got.Extensions), len(info.Extensions))
	}
	if got.TestPattern != info.TestPattern {
		t.Errorf("TestPattern = %q, want %q", got.TestPattern, info.TestPattern)
	}
	if got.CoverageCommand != info.CoverageCommand {
		t.Errorf("CoverageCommand = %q, want %q", got.CoverageCommand, info.CoverageCommand)
	}
}

// containsString checks if a string slice contains a target string.
func containsString(slice []string, target string) bool {
	return slices.Contains(slice, target)
}
