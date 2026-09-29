package config

// SPEC-WEB-SAVE-LOSSLESS-001 — M3 Save dirty-gate backstop (REQ-WSL-001,
// plan.md §A layer 3). ConfigManager.Save re-marshaled user/language/quality/
// llm unconditionally, so any Save() call — even one whose only intent was a
// single dirty section — rewrote the other files and dropped their unmodeled
// keys and comments (the D1 collateral-rewrite defect, GitHub issue #1731).
// The git-strategy/git-convention dirty-or-absent precedent
// (SPEC-GITSTRATEGY-SAVE-ISOLATION-001) extends to all four.
//
// RED-first against the pre-gate tree.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// seedDirtyGateFixture writes user/quality/llm files carrying unmodeled keys
// and comments, plus a language.yaml to edit.
func seedDirtyGateFixture(t *testing.T) (root, sectionsDir string) {
	t.Helper()
	root = t.TempDir()
	sectionsDir = filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"user.yaml":     "# user comment\nuser:\n  name: original\n  github_username: example-user\n",
		"quality.yaml":  "constitution:\n  development_mode: tdd\n  # quality comment\n  session_effort_default: xhigh  # local note\n",
		"llm.yaml":      "llm:\n  mode: \"\"\n  # llm comment\n  glm:\n    models:\n      high: glm-5.2\n",
		"language.yaml": "language:\n  conversation_language: en\n  conversation_language_name: en\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sectionsDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root, sectionsDir
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// TestSaveRewritesOnlyDirtySections carries the backstop core: a
// SetSection("language") + Save() must rewrite language.yaml and leave
// user/quality/llm byte-identical — unmodeled keys and comments included.
// RED on the pre-gate tree: Save re-marshaled all four unconditionally.
func TestSaveRewritesOnlyDirtySections(t *testing.T) {
	t.Parallel()
	root, sectionsDir := seedDirtyGateFixture(t)

	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if err := m.SetSection("language", models.LanguageConfig{
		ConversationLanguage:     "ko",
		ConversationLanguageName: "ko",
		GitCommitMessages:        "en",
		CodeComments:             "en",
		Documentation:            "ko",
	}); err != nil {
		t.Fatalf("SetSection(language): %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	langAfter := readFile(t, filepath.Join(sectionsDir, "language.yaml"))
	if !strings.Contains(langAfter, "conversation_language: ko") {
		t.Fatalf("dirty section not persisted:\n%s", langAfter)
	}

	for name, want := range map[string]string{
		"user.yaml":    "github_username: example-user",
		"quality.yaml": "session_effort_default: xhigh  # local note",
		"llm.yaml":     "high: glm-5.2",
	} {
		after := readFile(t, filepath.Join(sectionsDir, name))
		if !strings.Contains(after, want) {
			t.Errorf("%s lost unmodeled content %q in a language-only Save:\n%s", name, want, after)
		}
	}
}

// TestSaveWithoutDirtySectionsTouchesNothing carries the EC-3 reset contract
// for the extended gates: after a successful Save, a SECOND Save with no
// intervening SetSection must write nothing. The sections directory is made
// read-only so any rewrite errors out — silence proves the no-write.
func TestSaveWithoutDirtySectionsTouchesNothing(t *testing.T) {
	t.Parallel()
	root, sectionsDir := seedDirtyGateFixture(t)

	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if err := m.SetSection("language", models.LanguageConfig{ConversationLanguage: "ko"}); err != nil {
		t.Fatalf("SetSection: %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	// Freeze the directory: any second-write attempt fails on the temp file.
	if err := os.Chmod(sectionsDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(sectionsDir, 0o755) })

	if err := m.Save(); err != nil {
		t.Fatalf("second Save rewrote a section despite no dirty flag: %v", err)
	}
}

// TestSaveGreenfieldSectionsStillCreated carries the absent branch: section
// files that do not exist yet are still created (init-flow regression guard).
func TestSaveGreenfieldSectionsStillCreated(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	m := NewConfigManager()
	if _, err := m.LoadRaw(root); err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if err := m.SetSection("quality", models.QualityConfig{DevelopmentMode: models.ModeTDD}); err != nil {
		t.Fatalf("SetSection: %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	for _, f := range []string{"user.yaml", "language.yaml", "quality.yaml", "llm.yaml"} {
		if _, err := os.Stat(filepath.Join(sectionsDir, f)); err != nil {
			t.Errorf("greenfield file %s not created: %v", f, err)
		}
	}
}
