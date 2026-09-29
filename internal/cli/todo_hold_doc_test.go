// todo_hold_doc_test.go — SPEC-TODO-HOLD-STATE-001 M5 (AC-THS-018): the
// todo/GTD documentation surfaces document the `hold` state and both verbs,
// and each local copy's hold section is content-identical to its template
// mirror — a mirror omission fails this AC even when the live copy passes.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// holdDocParagraph is the section every todo.md surface must carry
// verbatim — the shared content the local and template copies assert
// identical.
const holdDocParagraph = "The queue carries four states: `queued`, `picked`, `dropped`, `hold`."

func TestTodoHoldDocumentedOnEverySurface(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	gtdLocal := filepath.Join(root, ".claude", "skills", "moai", "workflows", "gtd.md")
	gtdMirror := filepath.Join(root, "internal", "template", "templates", ".claude", "skills", "moai", "workflows", "gtd.md")
	todoLocal := filepath.Join(root, ".claude", "commands", "moai", "todo.md")
	todoMirror := filepath.Join(root, "internal", "template", "templates", ".claude", "commands", "moai", "todo.md")

	for _, tc := range []struct{ name, path string }{
		{"live gtd.md", gtdLocal},
		{"template mirror gtd.md", gtdMirror},
		{"live todo.md", todoLocal},
		{"template mirror todo.md", todoMirror},
	} {
		raw, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.name, err)
		}
		body := string(raw)
		// The hold state and both verbs are documented.
		for _, want := range []string{"hold", "unhold"} {
			if !strings.Contains(body, want) {
				t.Errorf("%s does not mention %q", tc.name, want)
			}
		}
	}

	// The gtd.md verb table carries both verb rows with the recovery-verb
	// guidance, and the JSON example carries the literal hold state value.
	for _, tc := range []struct{ name, path string }{{"gtd.md", gtdLocal}, {"gtd.md mirror", gtdMirror}} {
		raw, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.name, err)
		}
		body := string(raw)
		for _, want := range []string{
			"`moai gtd hold <n> [--expect <prefix>]`",
			"`moai gtd unhold <n> [--expect <prefix>]`",
			`"state": "hold"`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("%s lacks %q", tc.name, want)
			}
		}
	}

	// The todo.md surfaces carry the shared hold paragraph verbatim, and the
	// local and template copies' hold sections are content-identical.
	section := func(raw string) string {
		at := strings.Index(raw, holdDocParagraph)
		if at < 0 {
			return ""
		}
		return raw[at:]
	}
	localRaw, err := os.ReadFile(todoLocal)
	if err != nil {
		t.Fatalf("read live todo.md: %v", err)
	}
	mirrorRaw, err := os.ReadFile(todoMirror)
	if err != nil {
		t.Fatalf("read template mirror todo.md: %v", err)
	}
	localSection, mirrorSection := section(string(localRaw)), section(string(mirrorRaw))
	if localSection == "" {
		t.Fatal("live todo.md lacks the hold-state section")
	}
	if localSection != mirrorSection {
		t.Errorf("the hold section diverged between the local todo.md and its template mirror.\nlocal:\n%s\nmirror:\n%s",
			localSection, mirrorSection)
	}

	// The gtd.md mirrors are byte-identical by prior convention.
	gtdLocalRaw, err := os.ReadFile(gtdLocal)
	if err != nil {
		t.Fatalf("read live gtd.md: %v", err)
	}
	gtdMirrorRaw, err := os.ReadFile(gtdMirror)
	if err != nil {
		t.Fatalf("read template mirror gtd.md: %v", err)
	}
	if string(gtdLocalRaw) != string(gtdMirrorRaw) {
		t.Error("gtd.md and its template mirror are not byte-identical")
	}
}
