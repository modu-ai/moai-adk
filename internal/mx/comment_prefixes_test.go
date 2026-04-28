package mx

import (
	"testing"
)

func TestGetCommentPrefix(t *testing.T) {
	tests := []struct {
		ext         string
		wantPrefix  string
		description string
	}{
		{".go", "//", "Go files use //"},
		{".py", "#", "Python files use #"},
		{".ts", "//", "TypeScript files use //"},
		{".js", "//", "JavaScript files use //"},
		{".rs", "//", "Rust files use //"},
		{".java", "//", "Java files use //"},
		{".kt", "//", "Kotlin files use //"},
		{".cs", "//", "C# files use //"},
		{".rb", "#", "Ruby files use #"},
		{".php", "//", "PHP files use //"},
		{".ex", "#", "Elixir .ex files use #"},
		{".exs", "#", "Elixir .exs files use #"},
		{".cpp", "//", "C++ files use //"},
		{".hpp", "//", "C++ header files use //"},
		{".cc", "//", "C++ .cc files use //"},
		{".cxx", "//", "C++ .cxx files use //"},
		{".scala", "//", "Scala files use //"},
		{".R", "#", "R .R files use #"},
		{".r", "#", "R .r files use #"},
		{".dart", "//", "Dart files use //"},
		{".swift", "//", "Swift files use //"},
		{".mjs", "//", "JavaScript module files use //"},
		{".txt", "", "Unsupported extension returns empty"},
		{"", "", "Empty extension returns empty"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got := GetCommentPrefix(tt.ext)
			if got != tt.wantPrefix {
				t.Errorf("GetCommentPrefix(%q) = %q, want %q", tt.ext, got, tt.wantPrefix)
			}
		})
	}
}

func TestAll16LanguagesSupported(t *testing.T) {
	// AC-SPC-002-15: Verify all 16 supported languages have prefixes
	supportedExtensions := []string{
		".go", ".py", ".ts", ".js", ".rs", ".java", ".kt", ".cs",
		".rb", ".php", ".ex", ".cpp", ".scala", ".R", ".dart", ".swift",
	}

	for _, ext := range supportedExtensions {
		t.Run(ext, func(t *testing.T) {
			prefix := GetCommentPrefix(ext)
			if prefix == "" {
				t.Errorf("Language %s should have a comment prefix", ext)
			}
		})
	}
}
