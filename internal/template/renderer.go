package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"regexp"
	"text/template"
)

// unexpandedTokenPattern detects leftover dynamic tokens in rendered output.
// Matches ${VAR}, {{VAR}}, and $VAR patterns.
var unexpandedTokenPattern = regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*\}|\{\{\.?[A-Za-z_][A-Za-z0-9_.]*\}\}|\$[A-Z_][A-Z0-9_]*`)

// Renderer renders Go text/template files with strict mode enabled.
type Renderer interface {
	// Render parses the named template from the embedded FS and executes
	// it with the given data. Returns ErrMissingTemplateKey if a key is
	// missing and ErrUnexpandedToken if tokens remain after rendering.
	Render(templateName string, data any) ([]byte, error)
}

// renderer is the concrete implementation of Renderer.
type renderer struct {
	fsys fs.FS
}

// NewRenderer creates a Renderer backed by the given filesystem.
func NewRenderer(fsys fs.FS) Renderer {
	return &renderer{fsys: fsys}
}

// Render parses and executes a template with strict mode (missingkey=error).
func (r *renderer) Render(templateName string, data any) ([]byte, error) {
	content, err := fs.ReadFile(r.fsys, templateName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, templateName)
	}

	tmpl, err := template.New(templateName).
		Option("missingkey=error").
		Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("template parse %q: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMissingTemplateKey, err)
	}

	result := buf.Bytes()

	// Verify no unexpanded tokens remain (ADR-011)
	if loc := unexpandedTokenPattern.Find(result); loc != nil {
		return nil, fmt.Errorf("%w: found %q", ErrUnexpandedToken, string(loc))
	}

	return result, nil
}
