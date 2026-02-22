package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"text/template"
)

// templateFuncMap provides custom functions available in all templates.
var templateFuncMap = template.FuncMap{
	// jsonEscape escapes a string for safe embedding in JSON values.
	// It handles backslashes, quotes, and control characters by leveraging
	// encoding/json.Marshal, then stripping the surrounding quotes.
	"jsonEscape": func(s string) string {
		b, err := json.Marshal(s)
		if err != nil {
			return s
		}
		// json.Marshal wraps in quotes: "value" → strip them
		return string(b[1 : len(b)-1])
	},
	// posixPath converts Windows backslash paths to forward-slash POSIX paths.
	"posixPath": func(s string) string {
		return strings.ReplaceAll(s, "\\", "/")
	},
}

// unexpandedTokenPattern detects leftover dynamic tokens in rendered output.
// Matches ${VAR}, {{VAR}}, and $VAR patterns.
var unexpandedTokenPattern = regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*\}|\{\{\.?[A-Za-z_][A-Za-z0-9_.]*\}\}|\$[A-Z_][A-Z0-9_]*`)

// claudeCodePassthroughTokens are environment variables resolved by Claude Code
// at runtime and must not be flagged as unexpanded tokens (ADR-011 exception).
var claudeCodePassthroughTokens = []string{
	"$CLAUDE_PROJECT_DIR",
	"$ARGUMENTS",
}

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
		Funcs(templateFuncMap).
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

	// Verify no unexpanded tokens remain (ADR-011).
	// Mask Claude Code runtime env vars before validation.
	masked := string(result)
	for _, tok := range claudeCodePassthroughTokens {
		masked = strings.ReplaceAll(masked, tok, "")
	}
	if loc := unexpandedTokenPattern.Find([]byte(masked)); loc != nil {
		return nil, fmt.Errorf("%w: found %q", ErrUnexpandedToken, string(loc))
	}

	return result, nil
}
