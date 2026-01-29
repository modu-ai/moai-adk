package initializer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/anthropics/moai-adk-go/internal/output"
	templatepkg "github.com/anthropics/moai-adk-go/pkg/templates"
)

// Extractor handles template extraction from embed.FS
type Extractor struct {
	templateFS fs.FS
}

// NewExtractor creates a new template extractor
func NewExtractor() *Extractor {
	return &Extractor{
		templateFS: templatepkg.GetTemplatesFS(),
	}
}

// ExtractTemplates extracts templates to the target directory
func (e *Extractor) ExtractTemplates(targetDir string) error {
	// Walk through the embedded filesystem and extract files
	return fs.WalkDir(e.templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing template path %s: %w", path, err)
		}

		// Skip directories (they'll be created when files are extracted)
		if d.IsDir() {
			return nil
		}

		// Construct target path
		targetPath := filepath.Join(targetDir, path)

		// Create directory structure
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", filepath.Dir(targetPath), err)
		}

		// Read file content from embedded FS
		content, err := fs.ReadFile(e.templateFS, path)
		if err != nil {
			return fmt.Errorf("error reading template file %s: %w", path, err)
		}

		// Write file to target
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("error writing file %s: %w", targetPath, err)
		}

		fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Extracted: %s", path)))
		return nil
	})
}

// ExtractFile extracts a single file from templates
func (e *Extractor) ExtractFile(templatePath, targetPath string) error {
	// Read file content from embedded FS
	content, err := fs.ReadFile(e.templateFS, templatePath)
	if err != nil {
		return fmt.Errorf("error reading template file %s: %w", templatePath, err)
	}

	// Create directory structure
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", filepath.Dir(targetPath), err)
	}

	// Write file to target
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", targetPath, err)
	}

	return nil
}

// ListTemplates returns a list of all embedded template files
func (e *Extractor) ListTemplates() ([]string, error) {
	var files []string

	err := fs.WalkDir(e.templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
