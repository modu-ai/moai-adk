package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestFactoryGuideTeachesLaneFormsInEveryLocale pins the lane-axis vocabulary
// of the factory notice in all four locales, in lockstep: the lane-start
// sentence teaches the `-l` entries and the entry guide the `-f` leader
// entries (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009), and none of them still
// teaches the legacy `-f worker` / `worker-<n>` / `-f agent` / `agent-<n>`
// spellings (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-001/-003/-004; those
// spellings are refused, never advertised) or the removed `-f lane` forms.
func TestFactoryGuideTeachesLaneFormsInEveryLocale(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{langEnglish, "ko", "ja", "zh"} {
		m, ok := factoryLocales[lang]
		if !ok {
			t.Fatalf("locale %q missing from factoryLocales", lang)
		}
		for _, want := range []string{"`moai cc -l`", "`moai glm -l`", "`moai codex -l`"} {
			if !strings.Contains(m.leaderManual, want) {
				t.Errorf("%s leaderManual lacks %q:\n%s", lang, want, m.leaderManual)
			}
		}
		for _, want := range []string{"`moai cc -f`", "`moai glm -f`"} {
			if !strings.Contains(m.entryGuide, want) {
				t.Errorf("%s entryGuide lacks %q:\n%s", lang, want, m.entryGuide)
			}
		}
		for field, text := range map[string]string{"leaderManual": m.leaderManual, "entryGuide": m.entryGuide} {
			for _, banned := range []string{"-f worker", "worker-<n>", "-f agent", "agent-<n>", "-f lane", "lane-<n>"} {
				if strings.Contains(text, banned) {
					t.Errorf("%s %s still teaches %q:\n%s", lang, field, banned, text)
				}
			}
		}
		// SPEC-WORKFLOW-TASKS-001 REQ-TASKS-005: both lane rules teach the
		// TaskCreate/TaskUpdate discipline in every locale, the protocol
		// tokens verbatim the same way the MCP tool names are.
		for field, text := range map[string]string{"laneNextCardRule": m.laneNextCardRule, "laneOwnedCardRule": m.laneOwnedCardRule} {
			for _, token := range []string{"TaskCreate", "TaskUpdate"} {
				if !strings.Contains(text, token) {
					t.Errorf("%s %s lacks the tasks-discipline token %s:\n%s", lang, field, token, text)
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
