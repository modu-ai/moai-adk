package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestFactoryGuideTeachesWorkerFormsInEveryLocale pins the worker-axis
// vocabulary of the factory notice in all four locales, in lockstep: the
// naming sentence and the entry guide teach `worker-<n>`, `-f worker`, and
// `-f worker-<n>`, and none of them still teaches the legacy `lane-<n>` /
// `-f agent` / `agent-<n>` spellings (those still parse, with a hint, but are
// no longer advertised).
func TestFactoryGuideTeachesWorkerFormsInEveryLocale(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langEnglish, "ko", "ja", "zh"} {
		m, ok := factoryLocales[lang]
		if !ok {
			t.Fatalf("locale %q missing from factoryLocales", lang)
		}
		if !strings.Contains(m.leadManual, "worker-1..worker-%d") {
			t.Errorf("%s leadManual lacks the worker-1..worker-%%d naming:\n%s", lang, m.leadManual)
		}
		for _, want := range []string{"`moai %[2]s -f worker`", "`moai %[2]s -f worker-<n>`", "worker-<n>"} {
			if !strings.Contains(m.entryGuide, want) {
				t.Errorf("%s entryGuide lacks %q:\n%s", lang, want, m.entryGuide)
			}
		}
		for field, text := range map[string]string{"leadManual": m.leadManual, "entryGuide": m.entryGuide} {
			for _, banned := range []string{"lane-", "-f agent", "agent-<n>"} {
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
	for label, want := range map[string]int{"worker-4": 4, "lane-3": 3, "agent-2": 2} {
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, label)
		role, n, ok := kanbanRoleFromEnv()
		if !ok || role != kanban.RoleLane || n != want {
			t.Errorf("kanbanRoleFromEnv(%s) = (%q, %d, %v), want (%q, %d, true)", label, role, n, ok, kanban.RoleLane, want)
		}
	}
}

// TestKanbanNameChoicesUseWorkerNotation: the kanban notice's name-options
// line names the numbered factory shape as `worker-N` in every locale.
func TestKanbanNameChoicesUseWorkerNotation(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langEnglish, "ko", "ja", "zh"} {
		m := kanbanMessagesFor(lang)
		if !strings.Contains(m.nameChoices, "`worker-N`") || strings.Contains(m.nameChoices, "lane-N") {
			t.Errorf("%s nameChoices does not name `worker-N` (or still names lane-N): %s", lang, m.nameChoices)
		}
	}
}
