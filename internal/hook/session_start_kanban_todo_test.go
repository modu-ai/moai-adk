package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// writeTodoWorkflowConfig writes a workflow.yaml carrying body into a fresh
// temp project root and returns that root. An empty body writes no file at
// all, which is the "key absent" control condition.
func writeTodoWorkflowConfig(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("write workflow.yaml: %v", err)
		}
	}
	return root
}

// seedBacklog writes a backlog queue holding one queued card under root, so
// the summary line has a non-trivial count to print. The suppression assertion
// must not be able to pass merely because the queue was empty.
func seedBacklog(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "state", "kanban")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir kanban state: %v", err)
	}
	const queue = `{"version":1,"last_seq":1,"items":[{"id":"t1","text":"a queued card","added_at":"2026-08-22T00:00:00Z","spec_id":null,"state":"queued"}]}`
	if err := os.WriteFile(filepath.Join(dir, "backlog.json"), []byte(queue), 0o644); err != nil {
		t.Fatalf("write backlog.json: %v", err)
	}
}

// backlogLineFor renders the locale's backlog-summary line for n queued cards
// — the exact string the suppression must remove.
func backlogLineFor(lang string, n int) string {
	return fmt.Sprintf(kanbanMessagesFor(lang).backlogSummary, n)
}

// TestSessionStartKanbanRespectsTodoDisabled is AC-T-002 (REQ-2 surface 1).
//
// Two halves, and the second is the load-bearing one: a suppression test that
// only asserts absence passes just as happily on a build that removed the line
// outright, or on a run where the notice never contained it. The control case
// pins that the line IS emitted when the key is absent, so the first half is
// evidence about the flag rather than about the fixture.
func TestSessionStartKanbanRespectsTodoDisabled(t *testing.T) {
	langs := []string{langEnglish, "ko", "ja", "zh"}

	t.Run("disabled suppresses the backlog line in every locale", func(t *testing.T) {
		root := writeTodoWorkflowConfig(t, "workflow:\n    todo:\n        enabled: false\n")
		seedBacklog(t, root)

		for _, lang := range langs {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanID, "run42")

			got := kanbanLeadNotice("run42", root, lang)
			if got == "" {
				t.Fatalf("[%s] notice empty — the fixture stopped exercising the lead branch", lang)
			}
			if strings.Contains(got, backlogLineFor(lang, 1)) {
				t.Errorf("[%s] backlog line present with todo disabled:\n%s", lang, got)
			}
			// The count-independent stem must be gone too: a notice that
			// printed "0 waiting" instead of nothing would still be guidance.
			if strings.Contains(got, "moai todo") {
				t.Errorf("[%s] notice still mentions `moai todo` with todo disabled:\n%s", lang, got)
			}
		}
	})

	t.Run("control: key absent still emits the backlog line", func(t *testing.T) {
		root := writeTodoWorkflowConfig(t, "")
		seedBacklog(t, root)

		for _, lang := range langs {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanID, "run42")

			got := kanbanLeadNotice("run42", root, lang)
			want := backlogLineFor(lang, 1)
			if !strings.Contains(got, want) {
				t.Errorf("[%s] control case lost the backlog line\nwant: %s\ngot:\n%s", lang, want, got)
			}
		}
	})

	t.Run("control: explicit true still emits the backlog line", func(t *testing.T) {
		root := writeTodoWorkflowConfig(t, "workflow:\n    todo:\n        enabled: true\n")
		seedBacklog(t, root)

		clearKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanban, "1")
		t.Setenv(config.EnvMoaiKanbanID, "run42")

		got := kanbanLeadNotice("run42", root, langEnglish)
		if want := backlogLineFor(langEnglish, 1); !strings.Contains(got, want) {
			t.Errorf("explicit true lost the backlog line\nwant: %s\ngot:\n%s", want, got)
		}
	})
}
