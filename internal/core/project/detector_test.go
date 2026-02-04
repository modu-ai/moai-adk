package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/foundation"
	"github.com/modu-ai/moai-adk/pkg/models"
)

func newTestDetector() Detector {
	return NewDetector(foundation.DefaultRegistry, nil)
}

// --- DetectLanguages tests ---

func TestDetectLanguages(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, root string)
		wantLangs []string // expected language names in order
		wantEmpty bool
	}{
		{
			name: "go project with go.mod and go files",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "go.mod", "module example.com/test\ngo 1.22\n")
				writeFile(t, root, "main.go", "package main\nfunc main() {}\n")
				writeFile(t, root, "internal/app.go", "package internal\n")
			},
			wantLangs: []string{"Go"},
		},
		{
			name: "python project with pyproject.toml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "pyproject.toml", "[project]\nname = \"myapp\"\n")
				writeFile(t, root, "app/main.py", "print('hello')\n")
				writeFile(t, root, "app/routes.py", "# routes\n")
			},
			wantLangs: []string{"Python"},
		},
		{
			name: "javascript project with package.json",
			setup: func(t *testing.T, root string) {
				t.Helper()
				pkg := map[string]any{"name": "test", "dependencies": map[string]string{}}
				writeJSON(t, root, "package.json", pkg)
				writeFile(t, root, "src/index.js", "console.log('hi');\n")
			},
			wantLangs: []string{"JavaScript"},
		},
		{
			name: "rust project with Cargo.toml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "Cargo.toml", "[package]\nname = \"myapp\"\n")
				writeFile(t, root, "src/main.rs", "fn main() {}\n")
			},
			wantLangs: []string{"Rust"},
		},
		{
			name: "multi-language project sorts by confidence",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "go.mod", "module test\ngo 1.22\n")
				// 5 Go files
				for _, f := range []string{"a.go", "b.go", "c.go", "d.go", "e.go"} {
					writeFile(t, root, f, "package main\n")
				}
				// 2 Python files
				pkg := map[string]any{"name": "test"}
				writeJSON(t, root, "package.json", pkg)
				writeFile(t, root, "src/a.js", "//js\n")
				writeFile(t, root, "src/b.js", "//js\n")
			},
			wantLangs: []string{"Go", "JavaScript"},
		},
		{
			name:      "empty project returns nil",
			setup:     func(t *testing.T, root string) { t.Helper() },
			wantEmpty: true,
		},
		{
			name: "config file only (no source files) still detects language",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "go.mod", "module test\ngo 1.22\n")
			},
			wantLangs: []string{"Go"},
		},
		{
			name: "java project with pom.xml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "pom.xml", "<project></project>\n")
				writeFile(t, root, "src/Main.java", "class Main {}\n")
			},
			wantLangs: []string{"Java"},
		},
		{
			name: "ruby project with Gemfile",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "Gemfile", "source 'https://rubygems.org'\n")
				writeFile(t, root, "app.rb", "puts 'hello'\n")
			},
			wantLangs: []string{"Ruby"},
		},
		{
			name: "php project with composer.json",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "composer.json", "{}\n")
				writeFile(t, root, "index.php", "<?php echo 'hi'; ?>\n")
			},
			wantLangs: []string{"PHP"},
		},
		{
			name: "dart project with pubspec.yaml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "pubspec.yaml", "name: myapp\n")
				writeFile(t, root, "lib/main.dart", "void main() {}\n")
			},
			wantLangs: []string{"Dart"},
		},
		{
			name: "elixir project with mix.exs",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "mix.exs", "defmodule MyApp do\nend\n")
				writeFile(t, root, "lib/app.ex", "defmodule App do\nend\n")
			},
			wantLangs: []string{"Elixir"},
		},
		{
			name: "scala project with build.sbt",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "build.sbt", "name := \"myapp\"\n")
				writeFile(t, root, "src/Main.scala", "object Main\n")
			},
			wantLangs: []string{"Scala"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			d := newTestDetector()
			langs, err := d.DetectLanguages(root)
			if err != nil {
				t.Fatalf("DetectLanguages() error = %v", err)
			}

			if tt.wantEmpty {
				if len(langs) != 0 {
					t.Fatalf("expected empty, got %d languages", len(langs))
				}
				return
			}

			if len(langs) < len(tt.wantLangs) {
				t.Fatalf("expected at least %d languages, got %d", len(tt.wantLangs), len(langs))
			}

			for i, want := range tt.wantLangs {
				if langs[i].Name != want {
					t.Errorf("language[%d] = %q, want %q", i, langs[i].Name, want)
				}
				if langs[i].Confidence <= 0 {
					t.Errorf("language[%d] confidence = %f, want > 0", i, langs[i].Confidence)
				}
				if langs[i].FileCount <= 0 {
					t.Errorf("language[%d] file count = %d, want > 0", i, langs[i].FileCount)
				}
			}
		})
	}
}

func TestDetectLanguages_InvalidRoot(t *testing.T) {
	d := newTestDetector()
	_, err := d.DetectLanguages("/nonexistent/path/12345")
	if err == nil {
		t.Fatal("expected error for invalid root")
	}
}

// --- DetectFrameworks tests ---

func TestDetectFrameworks(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, root string)
		wantFrameworks []string
	}{
		{
			name: "react in package.json dependencies",
			setup: func(t *testing.T, root string) {
				t.Helper()
				pkg := map[string]any{
					"dependencies": map[string]string{"react": "^18.2.0", "react-dom": "^18.2.0"},
				}
				writeJSON(t, root, "package.json", pkg)
			},
			wantFrameworks: []string{"React"},
		},
		{
			name: "next.js in package.json",
			setup: func(t *testing.T, root string) {
				t.Helper()
				pkg := map[string]any{
					"dependencies": map[string]string{"next": "14.0.0", "react": "^18.2.0"},
				}
				writeJSON(t, root, "package.json", pkg)
			},
			wantFrameworks: []string{"Next.js", "React"},
		},
		{
			name: "vue in package.json",
			setup: func(t *testing.T, root string) {
				t.Helper()
				pkg := map[string]any{
					"dependencies": map[string]string{"vue": "^3.3.0"},
				}
				writeJSON(t, root, "package.json", pkg)
			},
			wantFrameworks: []string{"Vue"},
		},
		{
			name: "gin in go.mod",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "go.mod", "module test\ngo 1.22\nrequire github.com/gin-gonic/gin v1.9.1\n")
			},
			wantFrameworks: []string{"Gin"},
		},
		{
			name: "fastapi in pyproject.toml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "pyproject.toml", "[project]\ndependencies = [\"fastapi>=0.100.0\"]\n")
			},
			wantFrameworks: []string{"FastAPI"},
		},
		{
			name: "django in pyproject.toml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "pyproject.toml", "[project]\ndependencies = [\"django>=4.2\"]\n")
			},
			wantFrameworks: []string{"Django"},
		},
		{
			name: "flask in requirements.txt",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "requirements.txt", "flask==3.0.0\ngunicorn==21.2.0\n")
			},
			wantFrameworks: []string{"Flask"},
		},
		{
			name: "actix in Cargo.toml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "Cargo.toml", "[dependencies]\nactix-web = \"4\"\n")
			},
			wantFrameworks: []string{"Actix"},
		},
		{
			name: "no framework detected",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "go.mod", "module test\ngo 1.22\n")
			},
			wantFrameworks: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			d := newTestDetector()
			frameworks, err := d.DetectFrameworks(root)
			if err != nil {
				t.Fatalf("DetectFrameworks() error = %v", err)
			}

			if tt.wantFrameworks == nil {
				if len(frameworks) != 0 {
					t.Fatalf("expected no frameworks, got %d", len(frameworks))
				}
				return
			}

			if len(frameworks) != len(tt.wantFrameworks) {
				names := make([]string, len(frameworks))
				for i, f := range frameworks {
					names[i] = f.Name
				}
				t.Fatalf("expected %d frameworks %v, got %d: %v", len(tt.wantFrameworks), tt.wantFrameworks, len(frameworks), names)
			}

			for i, want := range tt.wantFrameworks {
				if frameworks[i].Name != want {
					t.Errorf("framework[%d] = %q, want %q", i, frameworks[i].Name, want)
				}
			}
		})
	}
}

func TestDetectFrameworks_PackageJSON_HasVersion(t *testing.T) {
	root := t.TempDir()
	pkg := map[string]any{
		"dependencies": map[string]string{"react": "^18.2.0"},
	}
	writeJSON(t, root, "package.json", pkg)

	d := newTestDetector()
	frameworks, err := d.DetectFrameworks(root)
	if err != nil {
		t.Fatalf("DetectFrameworks() error = %v", err)
	}

	if len(frameworks) == 0 {
		t.Fatal("expected at least one framework")
	}

	if frameworks[0].Version == "" {
		t.Error("expected non-empty version for React")
	}
	if frameworks[0].ConfigFile != "package.json" {
		t.Errorf("ConfigFile = %q, want %q", frameworks[0].ConfigFile, "package.json")
	}
}

// --- DetectProjectType tests ---

func TestDetectProjectType(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, root string)
		wantType models.ProjectType
	}{
		{
			name: "cli project with cmd directory",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, "cmd")
			},
			wantType: models.ProjectTypeCLI,
		},
		{
			name: "cli project with main.go",
			setup: func(t *testing.T, root string) {
				t.Helper()
				writeFile(t, root, "main.go", "package main\n")
			},
			wantType: models.ProjectTypeCLI,
		},
		{
			name: "web app with public directory",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, "public")
				mkDir(t, root, "src")
			},
			wantType: models.ProjectTypeWebApp,
		},
		{
			name: "web app with src/pages",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, "src/pages")
			},
			wantType: models.ProjectTypeWebApp,
		},
		{
			name: "api project with api directory",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, "api")
			},
			wantType: models.ProjectTypeAPI,
		},
		{
			name: "api project with routes directory",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, "routes")
			},
			wantType: models.ProjectTypeAPI,
		},
		{
			name: "library project (default)",
			setup: func(t *testing.T, root string) {
				t.Helper()
				mkDir(t, root, "src")
			},
			wantType: models.ProjectTypeLibrary,
		},
		{
			name:     "empty project defaults to library",
			setup:    func(t *testing.T, root string) { t.Helper() },
			wantType: models.ProjectTypeLibrary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			d := newTestDetector()
			got, err := d.DetectProjectType(root)
			if err != nil {
				t.Fatalf("DetectProjectType() error = %v", err)
			}
			if got != tt.wantType {
				t.Errorf("DetectProjectType() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

// --- Test helpers ---

func writeFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, root, relPath string, data any) {
	t.Helper()
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, relPath, string(b))
}

func mkDir(t *testing.T, root, relPath string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(fullPath, 0o755); err != nil {
		t.Fatal(err)
	}
}
