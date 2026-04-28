package mx

// commentPrefixes maps file extensions to their line comment prefixes.
// This table supports the 16 languages specified in SPEC-V3R2-SPC-002.
var commentPrefixes = map[string]string{
	// Go
	".go": "//",

	// Python
	".py": "#",

	// TypeScript / JavaScript
	".ts":  "//",
	".js":  "//",
	".mjs": "//",

	// Rust
	".rs": "//",

	// Java
	".java": "//",

	// Kotlin
	".kt": "//",

	// C#
	".cs": "//",

	// Ruby
	".rb": "#",

	// PHP (supports both // and #, prefer //)
	".php": "//",

	// Elixir
	".ex":  "#",
	".exs": "#",

	// C++
	".cpp": "//",
	".hpp": "//",
	".cc":  "//",
	".cxx": "//",

	// Scala
	".scala": "//",

	// R
	".R":  "#",
	".r":  "#",

	// Flutter/Dart
	".dart": "//",

	// Swift
	".swift": "//",
}

// GetCommentPrefix returns the comment prefix for a given file extension.
// Returns empty string if extension is not supported.
func GetCommentPrefix(ext string) string {
	return commentPrefixes[ext]
}
