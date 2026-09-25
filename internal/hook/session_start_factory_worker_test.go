package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestFactoryGuideTeachesWorkerFormsInEveryLocale pins the agent-axis
// vocabulary of the factory notice in all four locales, in lockstep: the
// naming sentence and the entry guide teach `agent-<n>`, `-f agent`, and
// `-f agent-<n>`, and none of them still teaches the legacy `lane-<n>` /
// `-f worker` / `worker-<n>` spellings.
func TestFactoryGuideTeachesWorkerFormsInEveryLocale(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langEnglish, "ko", "ja", "zh"} {
		m, ok := factoryLocales[lang]
		if !ok {
			t.Fatalf("locale %q missing from factoryLocales", lang)
		}
		if !strings.Contains(m.leadManual, "agent-1..agent-%d") {
			t.Errorf("%s leadManual lacks the agent-1..agent-%%d naming:\n%s", lang, m.leadManual)
		}
		for _, want := range []string{"`moai %[2]s -f agent`", "`moai %[2]s -f agent-<n>`", "agent-<n>"} {
			if !strings.Contains(m.entryGuide, want) {
				t.Errorf("%s entryGuide lacks %q:\n%s", lang, want, m.entryGuide)
			}
		}
		for field, text := range map[string]string{"leadManual": m.leadManual, "entryGuide": m.entryGuide} {
			for _, banned := range []string{"lane-", "-f worker", "worker-<n>"} {
				if strings.Contains(text, banned) {
					t.Errorf("%s %s still teaches %q:\n%s", lang, field, banned, text)
				}
			}
		}
	}
}

// TestKanbanRoleFromEnvReadsCanonicalAndLegacyWorkerLabels: a session
// launched by a pre-rename launcher still carries a legacy `lane-<n>` /
// `agent-<n>` label in its environment; the record path keeps reading its
// number, so upgrading the binary mid-run does not orphan that session.
func TestKanbanRoleFromEnvReadsCanonicalAndLegacyWorkerLabels(t *testing.T) {
	for label, want := range map[string]int{"agent-4": 4, "lane-3": 3, "worker-2": 2} {
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, label)
		role, n, ok := kanbanRoleFromEnv()
		if !ok || role != kanban.RoleLane || n != want {
			t.Errorf("kanbanRoleFromEnv(%s) = (%q, %d, %v), want (%q, %d, true)", label, role, n, ok, kanban.RoleLane, want)
		}
	}
}

// TestKanbanNameChoicesUseWorkerNotation: the kanban notice's name-options
// line names the numbered factory shape as `agent-N` in every locale.
func TestKanbanNameChoicesUseWorkerNotation(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langEnglish, "ko", "ja", "zh"} {
		m := kanbanMessagesFor(lang)
		if !strings.Contains(m.nameChoices, "`agent-N`") || strings.Contains(m.nameChoices, "lane-N") {
			t.Errorf("%s nameChoices does not name `agent-N` (or still names lane-N): %s", lang, m.nameChoices)
		}
	}
}
