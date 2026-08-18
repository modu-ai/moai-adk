package hook

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// configWithLang returns a ConfigProvider whose conversation_language is lang.
func configWithLang(lang string) ConfigProvider {
	return &auditConfigProvider{cfg: &config.Config{
		Language: models.LanguageConfig{ConversationLanguage: lang},
	}}
}

// TestKanbanLocalesCoverEveryField guards the failure a message table invites:
// a locale added or edited with a field left at its zero value renders a line
// as empty output rather than as a visible defect. Every locale must carry
// every field.
func TestKanbanLocalesCoverEveryField(t *testing.T) {
	for lang, m := range kanbanLocales {
		fields := map[string]string{
			"leadHeader":       m.leadHeader,
			"leadIdentity":     m.leadIdentity,
			"leadManual":       m.leadManual,
			"glmSubstitute":    m.glmSubstitute,
			"backendRecommend": m.backendRecommend,
			"agentFanout":      m.agentFanout,
			"nameChoices":      m.nameChoices,
			"leaderSocket":     m.leaderSocket,
			"settingsAuto":     m.settingsAuto,
			"settingsVerify":   m.settingsVerify,
			"specLine":         m.specLine,
			"backlogSummary":   m.backlogSummary,
			"companionJoin":    m.companionJoin,
		}
		for name, value := range fields {
			if strings.TrimSpace(value) == "" {
				t.Errorf("locale %q: field %s is empty", lang, name)
			}
		}
		// The format-bearing fields must keep their verb, or the run id, the
		// lead label, and the socket path silently vanish from that locale's
		// notice.
		for name, value := range map[string]string{
			"leadHeader":    m.leadHeader,
			"leadIdentity":  m.leadIdentity,
			"leaderSocket":  m.leaderSocket,
			"specLine":      m.specLine,
			"companionJoin": m.companionJoin,
		} {
			if !strings.Contains(value, "%s") {
				t.Errorf("locale %q: field %s lost its %%s verb: %q", lang, name, value)
			}
		}
		// The backlog summary carries the queued-card count; a locale that
		// loses the verb prints a literal %d instead of the number.
		if !strings.Contains(m.backlogSummary, "%d") {
			t.Errorf("locale %q: backlogSummary lost its %%d verb: %q", lang, m.backlogSummary)
		}
	}
}

// TestKanbanLocalesAreTheConversationLanguageSet pins the table to the four
// conversation languages. This is the COMPLETE set — it is unrelated to the 16
// supported programming languages, which never touch prose.
func TestKanbanLocalesAreTheConversationLanguageSet(t *testing.T) {
	want := []string{"en", "ko", "ja", "zh"}
	if len(kanbanLocales) != len(want) {
		t.Errorf("locale count = %d, want %d", len(kanbanLocales), len(want))
	}
	for _, lang := range want {
		if _, ok := kanbanLocales[lang]; !ok {
			t.Errorf("locale %q missing from the table", lang)
		}
	}
}

// TestKanbanMessagesForFallsBackToEnglish covers the values a real config
// yields when it is unset, unreadable, or carries a locale the table does not
// know. None of them may produce an empty notice.
func TestKanbanMessagesForFallsBackToEnglish(t *testing.T) {
	english := kanbanLocales[langEnglish]
	for _, lang := range []string{"", "fr", "ko-KR", "EN", "zz"} {
		if got := kanbanMessagesFor(lang); got != english {
			t.Errorf("kanbanMessagesFor(%q) did not fall back to English", lang)
		}
	}
	if kanbanMessagesFor("ko") == english {
		t.Error("kanbanMessagesFor(\"ko\") returned the English set — the table is not being consulted")
	}
}

// TestOperatorLangFailsOpen asserts locale resolution degrades to English
// rather than failing the session start, matching the surrounding hook code.
func TestOperatorLangFailsOpen(t *testing.T) {
	cases := map[string]ConfigProvider{
		"nil provider": nil,
		"nil config":   &auditConfigProvider{cfg: nil},
		"empty lang":   configWithLang(""),
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if got := operatorLang(cfg); got != langEnglish {
				t.Errorf("operatorLang = %q, want %q", got, langEnglish)
			}
		})
	}
	if got := operatorLang(configWithLang("ja")); got != "ja" {
		t.Errorf("operatorLang = %q, want %q", got, "ja")
	}
}

// TestKanbanNoticePreservesProtocolTokensInEveryLocale is the case that keeps
// the notice usable: the operator copies these lines into a terminal, so the
// commands, run id, and socket path must be byte-identical across locales. Only
// the prose may differ.
func TestKanbanNoticePreservesProtocolTokensInEveryLocale(t *testing.T) {
	for lang := range kanbanLocales {
		t.Run(lang, func(t *testing.T) {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanID, "tjpzpl")
			t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-socket-kanban/tjpzpl")
			t.Setenv(config.EnvMoaiKanbanSpec, "SPEC-FOO-001")

			got := kanbanBootstrapNotice("", lang)
			for _, want := range []string{
				"moai cc -k --name plan",
				"moai glm -k --name run",
				"moai cc -k --name sync",
				"moai glm -k --name",
				"`judge`",
				"`lane-N`",
				"/tmp/moai-socket-kanban/tjpzpl",
				"SPEC-FOO-001",
				"`moai todo`",
			} {
				if !strings.Contains(got, want) {
					t.Errorf("locale %q dropped protocol token %q:\n%s", lang, want, got)
				}
			}
		})
	}
}

// TestKanbanLeadNoticeBlockLayout pins the five-block layout. The operator is
// scanning for the four lines to copy, so the launch commands must stand apart
// as their own blank-separated block — in every locale, since the builder owns
// the layout and the message table carries no spacing of its own.
func TestKanbanLeadNoticeBlockLayout(t *testing.T) {
	for lang := range kanbanLocales {
		t.Run(lang, func(t *testing.T) {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanID, "tjq2bd")
			t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-socket-kanban/tjq2bd")
			t.Setenv(config.EnvMoaiKanbanSettingsInjected, "1")

			blocks := strings.Split(strings.TrimRight(kanbanBootstrapNotice("", lang), "\n"), "\n\n")
			if len(blocks) != 5 {
				t.Fatalf("expected 5 blank-separated blocks, got %d:\n%q", len(blocks), blocks)
			}
			// Block 3 is the launch block: exactly the three commands — each
			// role on its recommended launcher — and nothing else.
			launch := strings.Split(blocks[2], "\n")
			if len(launch) != 3 {
				t.Errorf("launch block holds %d lines, want 3:\n%q", len(launch), launch)
			}
			commands := make(map[string]bool, len(kanban.CompanionRoles))
			for _, role := range kanban.CompanionRoles {
				commands["moai "+kanban.CompanionLauncher(role)+" -k --name "+role] = true
			}
			for _, line := range launch {
				if !commands[line] {
					t.Errorf("launch block carries a non-command line: %q", line)
				}
			}
			// No block may be empty, and no line may be blank inside a block.
			for i, b := range blocks {
				if strings.TrimSpace(b) == "" {
					t.Errorf("block %d is empty", i)
				}
			}
		})
	}
}

// TestKanbanLeadNoticeSPECKeepsItsOwnLine is the regression case for a defect the
// message table's trailing-newline convention used to invite: the SPEC line had
// no terminator, so with MOAI_KANBAN_SPEC set the next context line welded
// itself onto the same line ("SPEC: SPEC-FOO-001" + prose). Layout now belongs
// to the builder, so the two cannot share a line.
func TestKanbanLeadNoticeSPECKeepsItsOwnLine(t *testing.T) {
	for lang := range kanbanLocales {
		t.Run(lang, func(t *testing.T) {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanID, "tjq2bd")
			t.Setenv(config.EnvMoaiKanbanSpec, "SPEC-FOO-001")

			for _, line := range strings.Split(kanbanBootstrapNotice("", lang), "\n") {
				if !strings.Contains(line, "SPEC-FOO-001") {
					continue
				}
				if !strings.HasSuffix(strings.TrimSpace(line), "SPEC-FOO-001") {
					t.Errorf("locale %q: the SPEC line carries trailing content: %q", lang, line)
				}
				return
			}
			t.Errorf("locale %q: notice omits the SPEC identifier", lang)
		})
	}
}

// TestSessionStartKanbanChannelsCarryTheirOwnLanguage is the acceptance case for
// the split: the operator-facing systemMessage is rendered in
// conversation_language while the agent-facing additionalContext stays English,
// mirroring the agent_prompt_language / conversation_language split that
// language.yaml already draws.
func TestSessionStartKanbanChannelsCarryTheirOwnLanguage(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "tjpzpl")

	projectDir := newKanbanProjectDir(t)
	out, err := NewSessionStartHandler(configWithLang("ko")).Handle(context.Background(), &HookInput{
		SessionID:  "uuid-kanban-i18n-001",
		CWD:        projectDir,
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// Operator channel: Korean prose.
	if !strings.Contains(out.SystemMessage, "칸반 모드") {
		t.Errorf("SystemMessage is not rendered in conversation_language ko:\n%s", out.SystemMessage)
	}
	if strings.Contains(out.SystemMessage, "lead session") {
		t.Errorf("SystemMessage still carries English prose:\n%s", out.SystemMessage)
	}

	// Agent channel: English prose, regardless of conversation_language.
	ac := out.HookSpecificOutput.AdditionalContext
	if !strings.Contains(ac, "Kanban Mode: run tjpzpl, lead session.") {
		t.Errorf("AdditionalContext is not English — the agent-facing copy must not follow conversation_language:\n%s", ac)
	}
	if strings.Contains(ac, "칸반 모드") {
		t.Errorf("AdditionalContext leaked the operator locale:\n%s", ac)
	}

	// Both channels carry the same launch lines.
	for _, want := range []string{"moai cc -k --name plan", "moai cc -k --name sync"} {
		if !strings.Contains(out.SystemMessage, want) || !strings.Contains(ac, want) {
			t.Errorf("launch line %q missing from one of the two channels", want)
		}
	}
}

// TestKanbanRecommendationTableMatchesLaunchLines pins the agreement the
// recommendation redesign exists for: each locale's recommendation table names
// the same backend for a role that the role's launch line launches with. The
// table is prose, the lines are copyable commands, and a drift between the two
// would hand the operator two different recommendations in one notice. The
// lead appears ONLY in the table — it is launched by the operator's own entry
// command, so the notice prints no `--name lead` line to copy.
func TestKanbanRecommendationTableMatchesLaunchLines(t *testing.T) {
	backendFor := map[string]string{"cc": "Claude (Opus)", "glm": "GLM"}
	for lang := range kanbanLocales {
		t.Run(lang, func(t *testing.T) {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanID, "tjrec")

			got := kanbanLeadNotice("tjrec", "", lang)
			for _, role := range kanban.CompanionRoles {
				line := "moai " + kanban.CompanionLauncher(role) + " -k --name " + role
				if !strings.Contains(got, line) {
					t.Errorf("locale %q: notice omits launch line %q:\n%s", lang, line, got)
				}
				row := regexp.MustCompile(`(?m)^  ` + role + `\s+→ (Claude \(Opus\)|GLM)$`).FindStringSubmatch(got)
				if row == nil {
					t.Errorf("locale %q: recommendation table has no row for %q:\n%s", lang, role, got)
					continue
				}
				if want := backendFor[kanban.CompanionLauncher(role)]; row[1] != want {
					t.Errorf("locale %q: role %q recommended as %s but its launch line uses %q",
						lang, role, row[1], kanban.CompanionLauncher(role))
				}
			}
			if !regexp.MustCompile(`(?m)^  lead\s+→ GLM$`).MatchString(got) {
				t.Errorf("locale %q: recommendation table omits the lead → GLM row:\n%s", lang, got)
			}
			if strings.Contains(got, "--name lead") {
				t.Errorf("locale %q: notice prints a lead launch line — the lead has none:\n%s", lang, got)
			}
		})
	}
}
