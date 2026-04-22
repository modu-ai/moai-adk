package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// templatePath is the path to the real lsp.yaml template file relative to this test package.
// Used by template compliance tests to detect schema drift early (see issue #683).
const templateRelPath = "../../../internal/template/templates/.moai/config/sections/lsp.yaml.tmpl"

// TestTemplate_NoBinaryKey is a regression test for issue #683.
// It ensures the real template never uses "binary:" as a YAML key — the struct only
// recognises "command:" (ServerConfig.Command yaml:"command"). This test prevents
// the exact class of schema drift that caused all 16 language servers to silently
// parse as empty Command strings.
func TestTemplate_NoBinaryKey(t *testing.T) {
	t.Parallel()

	absPath, err := filepath.Abs(templateRelPath)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	raw, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v — is internal/template/templates/ present?", absPath, err)
	}

	for i, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		// Detect bare "binary:" YAML key (not inside a comment or string value).
		// A comment line starts with "#" after trimming.
		if strings.HasPrefix(trimmed, "binary:") {
			t.Errorf("line %d: found forbidden key \"binary:\" — use \"command:\" instead (ServerConfig YAML tag)", i+1)
		}
	}
}

// TestTemplate_16LangCommandNonEmpty verifies the real template file has all 16
// canonical language servers with a non-empty command field.
// Parses the template as YAML (no Go template directives in lsp.yaml.tmpl).
func TestTemplate_16LangCommandNonEmpty(t *testing.T) {
	t.Parallel()

	absPath, err := filepath.Abs(templateRelPath)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}

	cfg, err := config.Load(absPath)
	if err != nil {
		t.Fatalf("config.Load(%q): %v", absPath, err)
	}

	if len(cfg.Servers) < 16 {
		t.Errorf("template has %d language servers, want >= 16", len(cfg.Servers))
	}

	for _, lang := range canonicalLanguages {
		sc, ok := cfg.Servers[lang]
		if !ok {
			t.Errorf("template missing language %q", lang)
			continue
		}
		if sc.Command == "" {
			t.Errorf("language %q: Command is empty — binary: → command: rename missing in template", lang)
		}
	}
}

// TestTemplate_16LangFileExtensionsNonEmpty verifies each language in the real template
// has at least one file_extensions entry so that detectLanguage() can route files.
func TestTemplate_16LangFileExtensionsNonEmpty(t *testing.T) {
	t.Parallel()

	absPath, err := filepath.Abs(templateRelPath)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}

	cfg, err := config.Load(absPath)
	if err != nil {
		t.Fatalf("config.Load(%q): %v", absPath, err)
	}

	for _, lang := range canonicalLanguages {
		sc, ok := cfg.Servers[lang]
		if !ok {
			continue // already reported by TestTemplate_16LangCommandNonEmpty
		}
		if len(sc.FileExtensions) == 0 {
			t.Errorf("language %q: FileExtensions is empty — add file_extensions to template", lang)
		}
	}
}
