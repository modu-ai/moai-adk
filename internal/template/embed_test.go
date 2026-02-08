package template

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedTemplates_ReturnsValidFS(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	if fsys == nil {
		t.Fatal("EmbeddedTemplates() returned nil fs.FS")
	}
}

func TestEmbeddedTemplates_PathStripping(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Paths should NOT require "templates/" prefix
	_, err = fs.ReadFile(fsys, "CLAUDE.md")
	if err != nil {
		t.Errorf("expected CLAUDE.md at root (no templates/ prefix), got error: %v", err)
	}
}

func TestEmbeddedTemplates_AgentDefinitions(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	agentDir := ".claude/agents/moai"
	entries, err := fs.ReadDir(fsys, agentDir)
	if err != nil {
		t.Fatalf("ReadDir(%q) error: %v", agentDir, err)
	}

	var mdCount int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			mdCount++
		}
	}

	if mdCount < 15 {
		t.Errorf("expected at least 15 agent .md files, got %d", mdCount)
	}

	// Verify a specific agent file is readable and non-empty
	data, err := fs.ReadFile(fsys, agentDir+"/expert-backend.md")
	if err != nil {
		t.Fatalf("read expert-backend.md: %v", err)
	}
	if len(data) == 0 {
		t.Error("expert-backend.md is empty")
	}
}

func TestEmbeddedTemplates_SkillDefinitions(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var skillCount int
	_ = fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			skillCount++
		}
		return nil
	})

	if skillCount < 350 {
		t.Errorf("expected at least 350 skill .md files, got %d", skillCount)
	}
}

func TestEmbeddedTemplates_RuleFiles(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var ruleCount int
	_ = fs.WalkDir(fsys, ".claude/rules/moai", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			ruleCount++
		}
		return nil
	})

	if ruleCount < 15 {
		t.Errorf("expected at least 15 rule files, got %d", ruleCount)
	}
}

func TestEmbeddedTemplates_OutputStyles(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	entries, err := fs.ReadDir(fsys, ".claude/output-styles/moai")
	if err != nil {
		t.Fatalf("ReadDir output-styles: %v", err)
	}

	var styleCount int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			styleCount++
		}
	}

	if styleCount < 2 {
		t.Errorf("expected at least 2 output style files, got %d", styleCount)
	}
}

func TestEmbeddedTemplates_CLAUDEmd(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, "CLAUDE.md")
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}

	content := string(data)
	if len(content) < 5000 {
		t.Errorf("CLAUDE.md should be at least 5000 characters, got %d", len(content))
	}
	if !strings.Contains(content, "MoAI Execution Directive") {
		t.Error("CLAUDE.md should contain 'MoAI Execution Directive'")
	}
}

func TestEmbeddedTemplates_NoMCPConfig(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// .mcp.json is rendered from .mcp.json.tmpl at deploy time, not embedded as bare JSON.
	for _, name := range []string{".mcp.json", ".mcp.windows.json"} {
		if _, err := fs.ReadFile(fsys, name); err == nil {
			t.Errorf("found %s which should be excluded from templates (runtime-generated)", name)
		}
	}
}

func TestEmbeddedTemplates_Gitignore(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	data, err := fs.ReadFile(fsys, ".gitignore")
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}

	if !strings.Contains(string(data), ".moai") {
		t.Error(".gitignore should contain .moai pattern")
	}
}

func TestEmbeddedTemplates_Announcements(t *testing.T) {
	t.Parallel()

	// Skip: announcements directory is not used in the Go implementation.
	// This was a Python implementation feature for displaying version updates.
	// The Go implementation handles version updates differently.
	t.Skip("announcements directory not implemented in Go version")
}

func TestEmbeddedTemplates_LLMConfig(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// .moai/llm-configs/ was removed in recent refactoring
	// LLM configuration is now handled via config manager
	// This test is skipped as the directory is no longer part of templates
	entries, err := fs.ReadDir(fsys, ".moai/llm-configs")
	if err != nil {
		// Expected: directory does not exist, which is correct after refactoring
		t.Skip("llm-configs directory was removed in refactoring")
	}

	if len(entries) < 1 {
		t.Error("expected at least 1 LLM config file")
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := fs.ReadFile(fsys, ".moai/llm-configs/"+e.Name())
		if err != nil {
			t.Errorf("read %s: %v", e.Name(), err)
			continue
		}
		if !json.Valid(data) {
			t.Errorf("%s is not valid JSON", e.Name())
		}
	}
}

// --- Exclusion tests (ACC-002) ---

func TestEmbeddedTemplates_NoPythonHooks(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Skip directories (only check files)
		if d.IsDir() {
			return nil
		}
		// Allow Go hook wrapper templates in .claude/hooks/moai/
		// Exclude Python hooks and any other .claude/hooks/ files
		if strings.HasPrefix(path, ".claude/hooks/") && !strings.HasPrefix(path, ".claude/hooks/moai/") {
			t.Errorf("found hooks file that should be excluded: %s", path)
		}
		if strings.HasSuffix(path, ".py") {
			t.Errorf("found Python file that should be excluded: %s", path)
		}
		return nil
	})
}

func TestEmbeddedTemplates_NoSettingsJSON(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	for _, name := range []string{".claude/settings.json", ".claude/settings.local.json"} {
		if _, err := fs.ReadFile(fsys, name); err == nil {
			t.Errorf("found %s which should be excluded from templates", name)
		}
	}
}

func TestEmbeddedTemplates_NoCacheOrOSFiles(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if name == "__pycache__" || name == ".DS_Store" || strings.HasSuffix(name, ".pyc") {
			t.Errorf("found cache/OS file that should be excluded: %s", path)
		}
		return nil
	})
}

func TestEmbeddedTemplates_NoLSPConfig(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	if _, err := fs.ReadFile(fsys, ".lsp.json"); err == nil {
		t.Error("found .lsp.json which should be excluded from templates")
	}
}

func TestEmbeddedTemplates_NoGitHooks(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if strings.HasPrefix(path, ".git-hooks/") {
			t.Errorf("found git-hooks file that should be excluded: %s", path)
		}
		return nil
	})
}

// --- WalkDir / file count test (ACC-003) ---

func TestEmbeddedTemplates_WalkDirTotalCount(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var totalFiles int
	walkErr := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Errorf("walk error at %s: %v", path, err)
			return nil
		}
		if path != "." && !d.IsDir() {
			totalFiles++
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir error: %v", walkErr)
	}

	if totalFiles < 450 {
		t.Errorf("expected at least 450 embedded files, got %d", totalFiles)
	}
	t.Logf("total embedded files: %d", totalFiles)
}

// --- Benchmark ---

func BenchmarkEmbeddedTemplatesWalkDir(b *testing.B) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		b.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var count int
		_ = fs.WalkDir(fsys, ".", func(_ string, d fs.DirEntry, _ error) error {
			if d != nil && !d.IsDir() {
				count++
			}
			return nil
		})
	}
}
