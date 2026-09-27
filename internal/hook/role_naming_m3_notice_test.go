package hook

// role_naming_m3_notice_test.go — SPEC-ROLE-NAMING-CODE-001 M3, AC-RNC-012
// (REQ-RNC-015): one table-driven test over the four conversation locales ×
// the factory leader notice, the factory lane notice, the kanban leader
// notice, and the stale-run notice (both the factory retire variant and the
// kanban relaunch variant). Each rendered notice must carry the design §3
// leader and lane terms for its locale; the three bootstrap notices must
// carry none of `worker-<n>`, `-f worker`, `-f agent`; the ko strings carry
// no 리드; the en strings carry no (?i)\blead\b match. The locale fallback
// for an unknown language resolves to the en table, which must itself pass
// the same assertions (covered here by the "en-fallback" subtest).

import (
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// m3LeaderTerm / m3LaneTerm are the design §3 terms per locale.
var (
	m3LeaderTerm = map[string]string{"en": "leader", "ko": "리더", "ja": "リーダー", "zh": "主导"}
	m3LaneTerm   = map[string]string{"en": "lane", "ko": "레인", "ja": "レーン", "zh": "泳道"}
	m3LeadWord   = regexp.MustCompile(`(?i)\blead\b`)
	m3WorkerNum  = regexp.MustCompile(`worker-\d`)
)

// m3ScrubEnv pins the notice-relevant environment so a sibling test's
// leftovers cannot steer the builders (the socket line, the settings line,
// and the stale-run discriminator all read the environment).
func m3ScrubEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiKanban, config.EnvMoaiKanbanID, config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanLeadName, config.EnvMoaiKanbanLeadAddr,
		config.EnvMoaiKanbanSpec, config.EnvMoaiKanbanSettingsInjected,
		config.EnvMoaiKanbanBackend, config.EnvMoaiFactoryWorker,
		config.EnvMoaiFactoryWorkers,
	} {
		t.Setenv(key, "")
	}
}

func TestRoleNamingM3NoticesCarryLeaderLaneTerms(t *testing.T) {
	m3ScrubEnv(t)
	// The stale-run variants need the run id; the bootstrap notices tolerate
	// it unset.
	t.Setenv(config.EnvMoaiKanbanID, "runX")

	langs := []string{"en", "ko", "ja", "zh", "en-fallback"}
	for _, lang := range langs {
		lang := lang
		t.Run(lang, func(t *testing.T) {
			tableLang := lang
			if lang == "en-fallback" {
				tableLang = "fr" // unknown language falls back to the en table
			}
			root := t.TempDir()

			leaderTerm := m3LeaderTerm["en"]
			laneTerm := m3LaneTerm["en"]
			if lang != "en-fallback" {
				leaderTerm = m3LeaderTerm[lang]
				laneTerm = m3LaneTerm[lang]
			}

			notices := map[string]string{
				// The three bootstrap notices (AC-RNC-012's named set).
				"factory-leader": factoryLeadNotice("runX", 2, root, tableLang),
				"factory-lane":   factoryWorkerNotice("lane-1", 3, tableLang),
				"kanban-leader":  kanbanLeadNotice("runX", root, tableLang),
				// The stale-run notice, both variants (plan.md §F M3).
				"stale-factory": staleRunNotice("worker-2", tableLang),
				"stale-kanban":  staleRunNotice("lead", tableLang),
			}

			for name, notice := range notices {
				if strings.TrimSpace(notice) == "" {
					t.Errorf("%s/%s: notice rendered empty", lang, name)
					continue
				}
				if !strings.Contains(notice, leaderTerm) {
					t.Errorf("%s/%s: notice missing the leader term %q:\n%s", lang, name, leaderTerm, notice)
				}
				if !strings.Contains(notice, laneTerm) {
					t.Errorf("%s/%s: notice missing the lane term %q:\n%s", lang, name, laneTerm, notice)
				}
				// The stale-run notice names the legacy value it detected by
				// design (REQ-RNC-022/025 — the notice must name `lead` /
				// `worker-<n>`), so the legacy-literal forbiddens below bind
				// only the three bootstrap notices. The stale rows still get
				// the ko-리드 check: the legacy value is quoted in the Latin
				// alphabet in every locale.
				if name == "stale-factory" || name == "stale-kanban" {
					if lang == "ko" && strings.Contains(notice, "리드") {
						t.Errorf("%s/%s: ko notice carries 리드:\n%s", lang, name, notice)
					}
					continue
				}
				if m3WorkerNum.MatchString(notice) {
					t.Errorf("%s/%s: notice carries worker-<n>:\n%s", lang, name, notice)
				}
				for _, legacyToken := range []string{"-f worker", "-f agent"} {
					if strings.Contains(notice, legacyToken) {
						t.Errorf("%s/%s: notice carries %q:\n%s", lang, name, legacyToken, notice)
					}
				}
				if m3LeadWord.MatchString(notice) {
					t.Errorf("%s/%s: en notice carries a \\blead\\b match:\n%s", lang, name, notice)
				}
				if lang == "ko" && strings.Contains(notice, "리드") {
					t.Errorf("%s/%s: ko notice carries 리드:\n%s", lang, name, notice)
				}
			}
		})
	}
}

// TestRoleNamingM3StaleRunNoticeNamesRetireStep pins the retire step on the
// factory variant and its absence on the kanban variant (a kanban run has no
// factory run to retire), per REQ-RNC-022's discriminator.
func TestRoleNamingM3StaleRunNoticeNamesRetireStep(t *testing.T) {
	m3ScrubEnv(t)
	t.Setenv(config.EnvMoaiKanbanID, "runX")
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")

	factory := staleRunNotice("worker-2", "en")
	for _, want := range []string{"worker-2", "runX", "moai factory runs --retire runX"} {
		if !strings.Contains(factory, want) {
			t.Errorf("factory stale-run notice %q missing %q", factory, want)
		}
	}

	t.Setenv(config.EnvMoaiFactoryWorkers, "")
	kanban := staleRunNotice("lead", "en")
	if strings.Contains(kanban, "factory runs --retire") {
		t.Errorf("kanban stale-run notice %q must not name a factory retire step", kanban)
	}
	if !strings.Contains(kanban, "lead") {
		t.Errorf("kanban stale-run notice %q does not name the legacy value", kanban)
	}
}
