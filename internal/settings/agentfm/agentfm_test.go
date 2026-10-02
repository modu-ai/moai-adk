package agentfm

// agentfm_test.go — SPEC-WEB-AGENTFM-RESTORE-001 M3: unit tests for the
// read-only scan layer (plan §F M3: 파싱 실패 행 · 빈 디렉터 저하).

import (
	"os"
	"path/filepath"
	"testing"
)

func writeAgentFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListParsesFrontmatter(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "manager-develop.md",
		"---\nname: manager-develop\ndescription: Implementation specialist\nmodel: inherit\n---\nbody bytes\n")
	writeAgentFile(t, dir, "manager-todo.md",
		"---\nname: manager-todo\ndescription: Queue management\neffort: \"\"\n---\nbody\n")

	agents, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("List = %d agents, want 2", len(agents))
	}
	// Name-sorted.
	dev, todo := agents[0], agents[1]
	if dev.Name != "manager-develop" || todo.Name != "manager-todo" {
		t.Fatalf("order = [%s, %s], want name-sorted", dev.Name, todo.Name)
	}
	if !dev.ParseOK || dev.Model != "inherit" || dev.Description != "Implementation specialist" {
		t.Errorf("manager-develop row = %+v", dev)
	}
	if !todo.ParseOK || !todo.EffortPresent {
		t.Errorf("manager-todo row must carry EffortPresent for the explicit empty effort: %+v", todo)
	}
}

func TestListDowngradesParseFailures(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "broken.md", "no frontmatter delimiter\n")
	writeAgentFile(t, dir, "unterminated.md", "---\nname: x\n") // no closing delimiter

	agents, err := List(dir)
	if err != nil {
		t.Fatalf("List must not fail on a broken file: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("List = %d agents, want 2 (broken files still list)", len(agents))
	}
	for _, a := range agents {
		if a.ParseOK {
			t.Errorf("row %s must downgrade to ParseOK=false", a.Name)
		}
	}
}

func TestListEmptyAndAbsentDirectories(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		agents, err := List(t.TempDir())
		if err != nil || len(agents) != 0 {
			t.Errorf("empty dir = %v, %v; want empty slice, nil", agents, err)
		}
	})
	t.Run("absent dir (greenfield)", func(t *testing.T) {
		agents, err := List(filepath.Join(t.TempDir(), "does-not-exist"))
		if err != nil || agents != nil {
			t.Errorf("absent dir = %v, %v; want nil, nil (greenfield degradation)", agents, err)
		}
	})
}

func TestListSkipsNonMarkdownAndDirectories(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "real.md", "---\nname: real\n---\nbody\n")
	writeAgentFile(t, dir, "notes.txt", "not an agent\n")
	if err := os.MkdirAll(filepath.Join(dir, "subdir.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	agents, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 1 || agents[0].Name != "real" {
		t.Errorf("List = %+v, want only the .md file", agents)
	}
}
