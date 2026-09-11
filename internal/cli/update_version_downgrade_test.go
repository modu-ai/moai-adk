package cli

// AC-ITI-012 (REQ-ITI-011): the `moai update --version` downgrade confirmation
// resolves its language project → active profile → English and renders title,
// description, both buttons, and help action labels in that language. Four
// cases, each a golden (ANSI-stripped View() at 80x40) plus property checks
// that do not read the table under test.

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
)

var updateDowngradeGolden = flag.Bool("update-golden", false, "rewrite the downgrade confirm goldens under testdata/downgrade-confirm")

const downgradeGoldenDir = "testdata/downgrade-confirm"

// downgradeKanbanVars is every MOAI_KANBAN* variable in internal/config/envkeys.go.
var downgradeKanbanVars = []string{
	config.EnvMoaiKanban, config.EnvMoaiKanbanSpec, config.EnvMoaiKanbanID,
	config.EnvMoaiKanbanLabel, config.EnvMoaiKanbanSettingsInjected, config.EnvMoaiKanbanLeadAddr,
	config.EnvMoaiKanbanBackend, config.EnvMoaiKanbanCard, config.EnvMoaiKanbanLeadName,
}

// downgradeExpect is the independent oracle per locale: the rendered title,
// description, and help action labels (design.md §7 values for toggle/submit).
var downgradeExpect = map[string]struct{ title, desc, toggle, submit string }{
	"en": {"Downgrade v9.9.9 → v1.0.0?", "The requested tag is older than the running version.", "toggle", "submit"},
	"ko": {"다운그레이드할까요? v9.9.9 → v1.0.0", "요청한 태그가 지금 실행 중인 버전보다 오래되었습니다.", "전환", "제출"},
	"ja": {"ダウングレードしますか？ v9.9.9 → v1.0.0", "指定したタグは実行中のバージョンより古いバージョンです。", "切替", "送信"},
	"zh": {"要降级吗？v9.9.9 → v1.0.0", "请求的标签比当前运行的版本旧。", "切换", "提交"},
}

// renderV2FormView drives a huh v2 form with bubbletea messages (no TTY) at
// 80x40 and returns its ANSI-stripped View().
func renderV2FormView(t *testing.T, f *huh.Form) string {
	t.Helper()
	var m huh.Model = f
	var drain func(cmd tea.Cmd)
	drain = func(cmd tea.Cmd) {
		if cmd == nil {
			return
		}
		ch := make(chan tea.Msg, 1)
		go func() { ch <- cmd() }()
		var msg tea.Msg
		select {
		case msg = <-ch:
		case <-time.After(50 * time.Millisecond):
			return
		}
		if msg == nil {
			return
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				drain(c)
			}
			return
		}
		var next tea.Cmd
		m, next = m.Update(msg)
		drain(next)
	}
	drain(f.Init())
	var next tea.Cmd
	m, next = m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	drain(next)
	return ptycaptest.StripANSI(m.(*huh.Form).View())
}

// seedLanguageYAML writes <dir>/.moai/config/sections/language.yaml.
func seedLanguageYAML(t *testing.T, dir, body string) {
	t.Helper()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "language.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestUpdateVersionDowngradeConfirm_Localized — AC-ITI-012 four cases. Not
// parallel: each case sets the process environment.
func TestUpdateVersionDowngradeConfirm_Localized(t *testing.T) {
	cases := []struct {
		name        string
		project     string // "" = no .moai; otherwise the language.yaml body
		profileName string // "" = no active profile
		profileLang string
		want        string
	}{
		{"a-project-ja-profile-ko", "language:\n    conversation_language: ja\n", "prof-ko", "ko", "ja"},
		{"b-no-project-profile-ko", "", "prof-ko", "ko", "ko"},
		{"c-no-project-no-profile", "", "", "", "en"},
		{"d-project-without-lang-profile-zh", "language:\n    code_comments: en\n", "prof-zh", "zh", "zh"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			moaiHome := t.TempDir()
			orig := profile.BaseDirOverride
			profile.BaseDirOverride = base
			t.Cleanup(func() { profile.BaseDirOverride = orig })

			configDir := ""
			if tc.profileName != "" {
				configDir = filepath.Join(base, tc.profileName)
				if err := profile.WritePreferences(tc.profileName, profile.ProfilePreferences{ConversationLang: tc.profileLang}); err != nil {
					t.Fatalf("seed profile: %v", err)
				}
			}
			prepared := map[string]string{config.EnvHome: moaiHome, config.EnvClaudeConfigDir: configDir}
			for _, k := range downgradeKanbanVars {
				prepared[k] = ""
			}
			for k, v := range prepared {
				t.Setenv(k, v)
			}
			for k, v := range prepared {
				if got := os.Getenv(k); got != v {
					t.Fatalf("env %s = %q at case start, want %q", k, got, v)
				}
			}

			cwd := t.TempDir()
			if tc.project != "" {
				seedLanguageYAML(t, cwd, tc.project)
			}

			locale := resolveDowngradeLocale(cwd)
			if locale != tc.want {
				t.Errorf("resolved locale %q, want %q", locale, tc.want)
			}
			var v bool
			view := renderV2FormView(t, wizard.NewDowngradeConfirmForm(locale, "v9.9.9", "v1.0.0", &v))
			ptycaptest.RequireLines(t, view, "v9.9.9 → v1.0.0")

			exp := downgradeExpect[tc.want]
			ui := wizard.GetUIStrings(tc.want)
			for label, want := range map[string]string{
				"title": exp.title, "description": exp.desc,
				"yes button": ui.ConfirmYes, "no button": ui.ConfirmNo,
				"toggle help": exp.toggle, "submit help": exp.submit,
			} {
				if !strings.Contains(view, want) {
					t.Errorf("%s %q not rendered; view:\n%s", label, want, view)
				}
			}
			if tc.want != "en" {
				for _, en := range []string{"toggle", "submit", "Downgrade", "older than the running version"} {
					if strings.Contains(view, en) {
						t.Errorf("%s view still carries the English %q", tc.want, en)
					}
				}
			}
			if err := ptycaptest.CompareGolden(downgradeGoldenDir, tc.name, view, *updateDowngradeGolden); err != nil {
				t.Error(err)
			}
		})
	}
}
