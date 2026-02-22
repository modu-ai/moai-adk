package project

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/foundation"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// Language represents a detected programming language with confidence scoring.
type Language struct {
	Name       string  // Human-readable language name (e.g., "Go", "Python").
	Confidence float64 // 0.0–1.0 ratio based on file count.
	FileCount  int     // Number of source files for this language.
}

// Framework represents a detected development framework.
type Framework struct {
	Name       string // Human-readable framework name (e.g., "React", "Gin").
	Version    string // Detected version string, may be empty.
	ConfigFile string // Config file where the framework was detected.
}

// Detector identifies project characteristics from the filesystem.
type Detector interface {
	// DetectLanguages scans the project root and returns detected languages
	// sorted by confidence descending.
	DetectLanguages(root string) ([]Language, error)

	// DetectFrameworks scans for known framework configuration files.
	DetectFrameworks(root string) ([]Framework, error)

	// DetectProjectType classifies the project based on structure and files.
	DetectProjectType(root string) (models.ProjectType, error)
}

// projectDetector is the concrete implementation of Detector.
type projectDetector struct {
	registry *foundation.LanguageRegistry
	logger   *slog.Logger
}

// NewDetector creates a Detector backed by the given language registry.
func NewDetector(registry *foundation.LanguageRegistry, logger *slog.Logger) Detector {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &projectDetector{
		registry: registry,
		logger:   logger,
	}
}

// configFileLanguageMap maps config file names to the languages they indicate.
var configFileLanguageMap = map[string]string{
	"package.json":     "JavaScript",
	"go.mod":           "Go",
	"pyproject.toml":   "Python",
	"requirements.txt": "Python",
	"Cargo.toml":       "Rust",
	"pom.xml":          "Java",
	"build.gradle":     "Java",
	"build.gradle.kts": "Kotlin",
	"Gemfile":          "Ruby",
	"composer.json":    "PHP",
	"Package.swift":    "Swift",
	"pubspec.yaml":     "Dart",
	"mix.exs":          "Elixir",
	"build.sbt":        "Scala",
	"stack.yaml":       "Haskell",
	"build.zig":        "Zig",
}

// skipDirs lists directories to skip during filesystem walks.
var skipDirs = map[string]bool{
	".git":         true,
	".moai":        true,
	".claude":      true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"__pycache__":  true,
	"target":       true,
	"build":        true,
	"dist":         true,
	".next":        true,
}

// DetectLanguages scans the project root and returns detected languages.
func (d *projectDetector) DetectLanguages(root string) ([]Language, error) {
	root = filepath.Clean(root)
	if err := validateRoot(root); err != nil {
		return nil, err
	}

	d.logger.Debug("detecting languages", "root", root)

	// Count files by language
	langCounts := make(map[string]int)
	totalFiles := 0

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}

		// Skip hidden and known non-source directories
		if entry.IsDir() {
			name := entry.Name()
			if skipDirs[name] || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == "" {
			return nil
		}

		info, lookupErr := d.registry.ByExtension(ext)
		if lookupErr != nil {
			return nil // skip unrecognized extensions
		}

		langCounts[info.Name]++
		totalFiles++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("detect languages in %s: %w", root, err)
	}

	// Also check config files for language hints
	for configFile, langName := range configFileLanguageMap {
		configPath := filepath.Join(root, configFile)
		if _, statErr := os.Stat(configPath); statErr == nil {
			if _, exists := langCounts[langName]; !exists {
				// Config file exists but no source files found yet;
				// add a minimal count to indicate presence.
				langCounts[langName] = max(langCounts[langName], 1)
				if totalFiles == 0 {
					totalFiles = 1
				}
			}
		}
	}

	// Check for TypeScript specifically when package.json exists
	if langCounts["JavaScript"] > 0 || langCounts["TypeScript"] > 0 {
		d.detectTypeScriptHint(root, langCounts)
	}

	if len(langCounts) == 0 {
		d.logger.Debug("no languages detected")
		return nil, nil
	}

	// Build Language slice with confidence
	languages := make([]Language, 0, len(langCounts))
	for name, count := range langCounts {
		confidence := float64(count) / float64(totalFiles)
		if confidence > 1.0 {
			confidence = 1.0
		}
		languages = append(languages, Language{
			Name:       name,
			Confidence: confidence,
			FileCount:  count,
		})
	}

	// Sort by confidence descending, then by name for stability
	sort.Slice(languages, func(i, j int) bool {
		if languages[i].Confidence != languages[j].Confidence {
			return languages[i].Confidence > languages[j].Confidence
		}
		return languages[i].Name < languages[j].Name
	})

	d.logger.Debug("languages detected", "count", len(languages))
	return languages, nil
}

// detectTypeScriptHint checks tsconfig.json to boost TypeScript detection.
func (d *projectDetector) detectTypeScriptHint(root string, langCounts map[string]int) {
	tsConfigPath := filepath.Join(root, "tsconfig.json")
	if _, err := os.Stat(tsConfigPath); err == nil {
		// tsconfig.json exists; if there are JS files, they may actually be TS
		if langCounts["TypeScript"] == 0 && langCounts["JavaScript"] > 0 {
			langCounts["TypeScript"] = langCounts["JavaScript"]
		}
	}
}

// packageJSON is used for parsing package.json dependencies.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// frameworkMapping maps dependency names to framework display names.
type frameworkMapping struct {
	Dependency string
	Framework  string
}

// jsFrameworks maps npm dependencies to framework names.
var jsFrameworks = []frameworkMapping{
	{"next", "Next.js"},
	{"react", "React"},
	{"vue", "Vue"},
	{"@angular/core", "Angular"},
	{"svelte", "Svelte"},
	{"express", "Express"},
	{"@nestjs/core", "NestJS"},
}

// goFrameworks maps Go module paths to framework names.
var goFrameworks = []frameworkMapping{
	{"github.com/gin-gonic/gin", "Gin"},
	{"github.com/labstack/echo", "Echo"},
	{"github.com/gofiber/fiber", "Fiber"},
	{"github.com/go-chi/chi", "Chi"},
}

// pythonFrameworks maps Python package names to framework names.
var pythonFrameworks = []frameworkMapping{
	{"fastapi", "FastAPI"},
	{"django", "Django"},
	{"flask", "Flask"},
}

// rustFrameworks maps Rust crate names to framework names.
var rustFrameworks = []frameworkMapping{
	{"actix-web", "Actix"},
	{"axum", "Axum"},
	{"rocket", "Rocket"},
}

// DetectFrameworks scans for known framework configuration files.
func (d *projectDetector) DetectFrameworks(root string) ([]Framework, error) {
	root = filepath.Clean(root)
	if err := validateRoot(root); err != nil {
		return nil, err
	}

	d.logger.Debug("detecting frameworks", "root", root)

	var frameworks []Framework

	// Check package.json for JS/TS frameworks
	if fws := d.detectPackageJSONFrameworks(root); len(fws) > 0 {
		frameworks = append(frameworks, fws...)
	}

	// Check go.mod for Go frameworks
	if fws := d.detectGoModFrameworks(root); len(fws) > 0 {
		frameworks = append(frameworks, fws...)
	}

	// Check pyproject.toml for Python frameworks
	if fws := d.detectPythonFrameworks(root); len(fws) > 0 {
		frameworks = append(frameworks, fws...)
	}

	// Check Cargo.toml for Rust frameworks
	if fws := d.detectCargoFrameworks(root); len(fws) > 0 {
		frameworks = append(frameworks, fws...)
	}

	d.logger.Debug("frameworks detected", "count", len(frameworks))
	return frameworks, nil
}

// detectPackageJSONFrameworks parses package.json for JS/TS frameworks.
func (d *projectDetector) detectPackageJSONFrameworks(root string) []Framework {
	pkgPath := filepath.Join(root, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		d.logger.Debug("failed to parse package.json", "error", err)
		return nil
	}

	var frameworks []Framework
	allDeps := mergeMaps(pkg.Dependencies, pkg.DevDependencies)

	for _, fm := range jsFrameworks {
		if version, ok := allDeps[fm.Dependency]; ok {
			frameworks = append(frameworks, Framework{
				Name:       fm.Framework,
				Version:    version,
				ConfigFile: "package.json",
			})
		}
	}

	return frameworks
}

// detectGoModFrameworks parses go.mod for Go frameworks.
func (d *projectDetector) detectGoModFrameworks(root string) []Framework {
	goModPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil
	}

	content := string(data)
	var frameworks []Framework

	for _, fm := range goFrameworks {
		if strings.Contains(content, fm.Dependency) {
			version := extractGoModVersion(content, fm.Dependency)
			frameworks = append(frameworks, Framework{
				Name:       fm.Framework,
				Version:    version,
				ConfigFile: "go.mod",
			})
		}
	}

	return frameworks
}

// detectPythonFrameworks checks pyproject.toml and requirements.txt for Python frameworks.
func (d *projectDetector) detectPythonFrameworks(root string) []Framework {
	var frameworks []Framework

	// Check pyproject.toml
	pyprojectPath := filepath.Join(root, "pyproject.toml")
	if data, err := os.ReadFile(pyprojectPath); err == nil {
		content := strings.ToLower(string(data))
		for _, fm := range pythonFrameworks {
			if strings.Contains(content, fm.Dependency) {
				frameworks = append(frameworks, Framework{
					Name:       fm.Framework,
					Version:    "",
					ConfigFile: "pyproject.toml",
				})
			}
		}
	}

	// Check requirements.txt if no frameworks found yet
	if len(frameworks) == 0 {
		reqPath := filepath.Join(root, "requirements.txt")
		if data, err := os.ReadFile(reqPath); err == nil {
			content := strings.ToLower(string(data))
			for _, fm := range pythonFrameworks {
				if strings.Contains(content, fm.Dependency) {
					frameworks = append(frameworks, Framework{
						Name:       fm.Framework,
						Version:    "",
						ConfigFile: "requirements.txt",
					})
				}
			}
		}
	}

	return frameworks
}

// detectCargoFrameworks checks Cargo.toml for Rust frameworks.
func (d *projectDetector) detectCargoFrameworks(root string) []Framework {
	cargoPath := filepath.Join(root, "Cargo.toml")
	data, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil
	}

	content := string(data)
	var frameworks []Framework

	for _, fm := range rustFrameworks {
		if strings.Contains(content, fm.Dependency) {
			frameworks = append(frameworks, Framework{
				Name:       fm.Framework,
				Version:    "",
				ConfigFile: "Cargo.toml",
			})
		}
	}

	return frameworks
}

// DetectProjectType classifies the project based on structure and files.
func (d *projectDetector) DetectProjectType(root string) (models.ProjectType, error) {
	root = filepath.Clean(root)
	if err := validateRoot(root); err != nil {
		return "", err
	}

	d.logger.Debug("detecting project type", "root", root)

	// IF cmd/ or main.go exists -> cli
	if dirExists(filepath.Join(root, "cmd")) || fileExists(filepath.Join(root, "main.go")) {
		return models.ProjectTypeCLI, nil
	}

	// ELSE IF public/ or src/pages/ exists -> web-app
	if dirExists(filepath.Join(root, "public")) || dirExists(filepath.Join(root, "src", "pages")) {
		return models.ProjectTypeWebApp, nil
	}

	// ELSE IF api/ or routes/ exists -> api
	if dirExists(filepath.Join(root, "api")) || dirExists(filepath.Join(root, "routes")) {
		return models.ProjectTypeAPI, nil
	}

	// ELSE -> library
	return models.ProjectTypeLibrary, nil
}

// validateRoot checks that the root path is a valid, accessible directory.
func validateRoot(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidRoot, root)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s is not a directory", ErrInvalidRoot, root)
	}
	return nil
}

// extractGoModVersion extracts the version of a module from go.mod content.
func extractGoModVersion(content, module string) string {
	lines := strings.SplitSeq(content, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, module) {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}

// mergeMaps merges two string maps, with the second taking precedence.
func mergeMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string, len(a)+len(b))
	maps.Copy(result, a)
	maps.Copy(result, b)
	return result
}

// dirExists checks if a path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// fileExists checks if a path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
