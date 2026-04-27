package harness

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkill(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	return path
}

func TestVerifyTriggers_AllKeysPresent(t *testing.T) {
	dir := t.TempDir()
	body := `---
name: my-harness-ios-patterns
description: foo
triggers:
  paths: ["**/*.swift"]
  keywords: ["ios"]
  agents: ["manager-tdd"]
  phases: ["plan", "run"]
---

# Body
`
	path := writeSkill(t, dir, "SKILL.md", body)
	if err := VerifyTriggers(path); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestVerifyTriggers_MissingOneKey(t *testing.T) {
	dir := t.TempDir()
	body := `---
name: my-harness-foo
triggers:
  paths: ["**/*.swift"]
  keywords: ["ios"]
  agents: ["manager-tdd"]
---

# Body
`
	path := writeSkill(t, dir, "SKILL.md", body)
	err := VerifyTriggers(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "phases") {
		t.Errorf("error should mention 'phases', got: %v", err)
	}
}

func TestVerifyTriggers_MissingTriggersSection(t *testing.T) {
	dir := t.TempDir()
	body := `---
name: my-harness-foo
description: bar
---

# No triggers
`
	path := writeSkill(t, dir, "SKILL.md", body)
	err := VerifyTriggers(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrMissingTriggers) {
		t.Errorf("expected ErrMissingTriggers, got: %v", err)
	}
}

func TestVerifyTriggers_EmptyPath(t *testing.T) {
	if err := VerifyTriggers(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestVerifyTriggers_FileNotFound(t *testing.T) {
	if err := VerifyTriggers(filepath.Join(t.TempDir(), "missing.md")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestVerifyTriggers_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "SKILL.md", "# Just a body, no frontmatter\n")
	if err := VerifyTriggers(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyTriggers_UnterminatedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "SKILL.md", "---\nname: foo\nno end delimiter here")
	if err := VerifyTriggers(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyTriggers_InvalidYaml(t *testing.T) {
	dir := t.TempDir()
	body := `---
name: foo
triggers: [this is: invalid yaml
---

# Body
`
	path := writeSkill(t, dir, "SKILL.md", body)
	if err := VerifyTriggers(path); err == nil {
		t.Fatal("expected yaml error")
	}
}

func TestVerifyTriggers_ListsAllMissing(t *testing.T) {
	dir := t.TempDir()
	body := `---
name: foo
triggers:
  paths: ["x"]
---

# Body
`
	path := writeSkill(t, dir, "SKILL.md", body)
	err := VerifyTriggers(path)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, expected := range []string{"keywords", "agents", "phases"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("missing key %q not in error: %v", expected, err)
		}
	}
}
