package template

import (
	"embed"
	"io/fs"
)

//go:embed .claude .moai
var templates embed.FS

// ExtractTemplate reads a template file from the embedded filesystem
// and returns its contents as a byte slice.
func ExtractTemplate(path string) ([]byte, error) {
	return templates.ReadFile(path)
}

// ListTemplates returns a list of all embedded template files.
// This is useful for debugging and verification.
func ListTemplates() ([]string, error) {
	var files []string

	err := fs.WalkDir(templates, ".", func(path string, d fs.DirEntry, err error) error {
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

// GetTemplatesFS returns the embedded filesystem for advanced operations.
func GetTemplatesFS() embed.FS {
	return templates
}
