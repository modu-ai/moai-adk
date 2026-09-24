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
	// yamlEscape escapes a string for safe embedding inside a YAML
	// double-quoted scalar ("..."); the template keeps the quotes.
	"yamlEscape": yamlEscape,
}

// @MX:NOTE: [AUTO] Escapes only what a YAML double-quoted scalar cannot carry literally (quote, backslash, line breaks, non-printable controls); every other character is emitted unchanged so ordinary values render byte-identical to the unescaped form.
// yamlEscape returns s escaped for the inside of a YAML double-quoted scalar,
// so that the rendered scalar parses back to exactly s. `"` and `\` are
// backslash-escaped; characters YAML either folds (LF, CR, NEL) or rejects
// (C0 controls other than tab, DEL, C1 controls, U+FFFE, U+FFFF) are written
// as \uXXXX escapes. Tab, printable ASCII, and printable non-ASCII (Hangul,
// emoji, ...) pass through unchanged. An invalid UTF-8 byte decodes as U+FFFD
// and is emitted as that rune: YAML text must be valid UTF-8, and no escape
// denotes a raw byte.
func yamlEscape(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\t':
			b.WriteRune(r)
		case r < 0x20, r == 0x7f, r >= 0x80 && r <= 0x9f, r == 0xfffe, r == 0xffff:
			fmt.Fprintf(&b, `\u%04X`, r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// unexpandedTokenPattern detects leftover dynamic tokens in rendered output.
// Matches ${VAR}, {{VAR}}, and $VAR patterns.
var unexpandedTokenPattern = regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*\}|\{\{\.?[A-Za-z_][A-Za-z0-9_.]*\}\}|\$[A-Z_][A-Z0-9_]*`)

// claudeCodePassthroughTokens are environment variables resolved at runtime
// (by Claude Code or the shell) and must not be flagged as unexpanded tokens
// (ADR-011 exception).
var claudeCodePassthroughTokens = []string{
	"$CLAUDE_PROJECT_DIR",
	"$CLAUDE_SKILL_DIR",
	"$ARGUMENTS",
	"$HOME",
	// GitHub Actions runtime environment variables
	"$GITHUB_OUTPUT",
	"$GITHUB_ENV",
	"$GITHUB_STEP_SUMMARY",
	// Shell local variables (legitimate in GitHub Actions run scripts)
	"$LANG_COUNT",
	// Hook wrapper stderr log path (resolved at shell runtime)
	"$MOAI_HOOK_STDERR_LOG",
	// Shell builtin variable (current working directory, resolved by the shell)
	"$PWD",
	// Hook wrapper shell-local variables (lifecycle dormant guards and the
	// stop-goal shell-layer precondition — assigned and consumed inside the
	// generated wrapper scripts, resolved at shell runtime)
	"$ACTION",
	"$AUTONOMY_TIER_DORMANT",
	"$INPUT",
	"$MOAI_BIN",
	"$PROJECT_ROOT",
	"$SESSION_ID",
	"$STATE_FILE",
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

// @MX:ANCHOR: [AUTO] Go text/template renderer factory - used from 3 or more paths including deployer, init, and update
// @MX:REASON: [AUTO] fan_in=3, single entry point for missingkey=error mode and claudeCodePassthroughTokens validation logic
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
	// Handle both $VAR and ${VAR} forms.
	masked := string(result)
	for _, tok := range claudeCodePassthroughTokens {
		masked = strings.ReplaceAll(masked, tok, "")
		masked = strings.ReplaceAll(masked, "${"+tok[1:]+"}", "")
	}
	if loc := unexpandedTokenPattern.Find([]byte(masked)); loc != nil {
		return nil, fmt.Errorf("%w: found %q", ErrUnexpandedToken, string(loc))
	}

	return result, nil
}
