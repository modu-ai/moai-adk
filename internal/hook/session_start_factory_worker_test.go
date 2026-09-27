package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestFactoryGuideTeachesLaneFormsInEveryLocale pins the lane-axis vocabulary
// of the factory notice in all four locales, in lockstep: the naming sentence
// and the entry guide teach `lane-<n>`, `-f lane`, and `-f lane-<n>`, and none
// of them still teaches the legacy `-f worker` / `worker-<n>` / `-f agent` /
// `agent-<n>` spellings (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-001/-003/-004;
// those spellings are refused, never advertised).
func TestFactoryGuideTeachesLaneFormsInEveryLocale(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langEnglish, "ko", "ja", "zh"} {
		m, ok := factoryLocales[lang]
		if !ok {
			t.Fatalf("locale %q missing from factoryLocales", lang)
		}
		if !strings.Contains(m.leadManual, "lane-1..lane-%d") {
			t.Errorf("%s leadManual lacks the lane-1..lane-%%d naming:\n%s", lang, m.leadManual)
		}
		for _, want := range []string{"`moai %[2]s -f lane`", "`moai %[2]s -f lane-<n>`", "lane-<n>"} {
			if !strings.Contains(m.entryGuide, want) {
				t.Errorf("%s entryGuide lacks %q:\n%s", lang, want, m.entryGuide)
			}
		}
		for field, text := range map[string]string{"leadManual": m.leadManual, "entryGuide": m.entryGuide} {
			for _, banned := range []string{"-f worker", "worker-<n>", "-f agent", "agent-<n>"} {
				if strings.Contains(text, banned) {
					t.Errorf("%s %s still teaches %q:\n%s", lang, field, banned, text)
				}
			}
		}
	}
}

// TestKanbanRoleFromEnvReadsOnlyLaneLabels: the record path derives a lane
// role from the canonical `lane-<n>` label only — legacy `worker-<n>` /
// `agent-<n>` labels map to no role (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009);
// the stale-run gate above this reader handles them (REQ-RNC-025).
func TestKanbanRoleFromEnvReadsOnlyLaneLabels(t *testing.T) {
	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-3")
	role, n, ok := kanbanRoleFromEnv()
	if !ok || role != kanban.RoleLane || n != 3 {
		t.Errorf("kanbanRoleFromEnv(lane-3) = (%q, %d, %v), want (%q, 3, true)", role, n, ok, kanban.RoleLane)
	}
	for _, label := range []string{"worker-4", "agent-2"} {
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, label)
		if role, n, ok := kanbanRoleFromEnv(); ok {
			t.Errorf("kanbanRoleFromEnv(%s) = (%q, %d, true), want ok=false — legacy maps to no role", label, role, n)
		}
	}
}

// TestKanbanNameChoicesUseLaneNotation: the kanban notice's name-options line
// names the numbered factory shape as `lane-N` in every locale.
func TestKanbanNameChoicesUseLaneNotation(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langEnglish, "ko", "ja", "zh"} {
		m := kanbanMessagesFor(lang)
		if !strings.Contains(m.nameChoices, "`lane-N`") || strings.Contains(m.nameChoices, "worker-N") {
			t.Errorf("%s nameChoices does not name `lane-N` (or still names worker-N): %s", lang, m.nameChoices)
		}
	}
}
