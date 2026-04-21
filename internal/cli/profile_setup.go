package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/spf13/cobra"
)

// 위저드가 제공하는 canonical 값 슬라이스 — huh.NewOption 블록(~line 230)과 반드시 동기화.
// KEEP IN SYNC with huh.NewOption block at ~line 230
var statuslineModeCanonical = []string{defaultStatuslineMode, "full"}

// KEEP IN SYNC with huh.NewOption block at ~line 240
var statuslineThemeCanonical = []string{defaultStatuslineTheme, "catppuccin-latte"}

// wizard 기본값 상수.
const (
	defaultStatuslineMode  = "default"
	defaultStatuslineTheme = "catppuccin-mocha"
	defaultPermissionMode  = "acceptEdits"
)

// isCanonicalStatuslineMode는 s 가 wizard 옵션 슬라이스에 포함된 canonical 값인지 확인한다.
func isCanonicalStatuslineMode(s string) bool {
	for _, v := range statuslineModeCanonical {
		if v == s {
			return true
		}
	}
	return false
}

// isCanonicalStatuslineTheme는 s 가 wizard 옵션 슬라이스에 포함된 canonical 값인지 확인한다.
func isCanonicalStatuslineTheme(s string) bool {
	for _, v := range statuslineThemeCanonical {
		if v == s {
			return true
		}
	}
	return false
}

// normalizeStatuslineModeRaw는 statusline 패키지의 NormalizeMode를 호출해
// 문자열 변환 노이즈를 캡슐화한다.
func normalizeStatuslineModeRaw(s string) string {
	return string(statusline.NormalizeMode(statusline.StatuslineMode(s)))
}

// normalizeStatuslineMode는 wizard 옵션과 호환되는 mode 값을 반환한다.
// 이전 statusline 버전의 deprecated 이름을 v3 이름으로 변환하며,
// 옵션 집합에 없는 값은 "default"로 폴백한다.
func normalizeStatuslineMode(mode string) string {
	if mode == "" {
		return defaultStatuslineMode
	}
	normalized := normalizeStatuslineModeRaw(mode)
	if isCanonicalStatuslineMode(normalized) {
		return normalized
	}
	return defaultStatuslineMode
}

// normalizeStatuslineTheme는 wizard 옵션과 호환되는 theme 이름을 반환한다.
// "default" 등 레거시 값은 "catppuccin-mocha"로 변환해 Select 위젯이 올바르게 선택 항목을 강조한다.
func normalizeStatuslineTheme(theme string) string {
	if isCanonicalStatuslineTheme(theme) {
		return theme
	}
	return defaultStatuslineTheme
}

// @MX:NOTE: [AUTO] 위저드 v3 마이그레이션 — deprecated Claude 모델 ID를 canonical alias로 정규화.
// @MX:REASON: 이전 위저드의 "claude-opus-4-7" 옵션 제거 후 기존 prefs 값이 huh.Select 바인딩에서 침묵 소실되는 것을 방지.
func normalizeModel(m string) string {
	switch m {
	// canonical alias 는 그대로 통과
	case "", "opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan":
		return m
	// deprecated full-ID → canonical alias
	case "claude-opus-4-7", "claude-opus-4-6":
		return "opus"
	case "claude-opus-4-7[1m]", "claude-opus-4-6[1m]", "claude-opus-4-6 1M":
		return "opus[1m]"
	case "claude-sonnet-4-6":
		return "sonnet"
	case "claude-sonnet-4-6[1m]", "claude-sonnet-4-6 1M":
		return "sonnet[1m]"
	case "claude-haiku-4-5":
		return "haiku"
	default:
		// 알 수 없는 값은 런타임 기본값으로 초기화
		return ""
	}
}

var profileSetupCmd = &cobra.Command{
	Use:   "setup [name]",
	Short: "Interactive setup wizard for profile preferences",
	Long: `Configure per-profile preferences through an interactive wizard.

Settings are stored in:
  ~/.moai/claude-profiles/<name>/preferences.yaml  (identity, language, model, display)

Examples:
  moai profile setup          # Configure default profile
  moai profile setup work     # Configure 'work' profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: runProfileSetup,
}

func init() {
	profileCmd.AddCommand(profileSetupCmd)
}

// runProfileSetup은 인터랙티브 프로필 설정 wizard를 실행한다.
// 첫 번째 질문은 언어 선택이며, 이후 모든 UI 텍스트는 선택된 언어로 표시된다.
func runProfileSetup(cmd *cobra.Command, args []string) error {
	profileName := "default"
	if len(args) > 0 {
		profileName = args[0]
	}

	// 기존 설정을 기본값으로 로드
	existingPrefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		return fmt.Errorf("read existing preferences: %w", err)
	}

	// 기존 설정으로 폼 값 초기화
	userName := existingPrefs.UserName

	convLang := existingPrefs.ConversationLang
	if convLang == "" {
		convLang = "en"
	}
	gitCommitLang := existingPrefs.GitCommitLang
	if gitCommitLang == "" {
		gitCommitLang = "en"
	}
	codeCommentLang := existingPrefs.CodeCommentLang
	if codeCommentLang == "" {
		codeCommentLang = "en"
	}
	docLang := existingPrefs.DocLang
	if docLang == "" {
		docLang = "en"
	}

	// C-1: deprecated 모델 ID를 canonical alias로 정규화
	model := normalizeModel(existingPrefs.Model)
	effortLevel := existingPrefs.EffortLevel
	permissionMode := existingPrefs.PermissionMode
	if permissionMode == "" {
		permissionMode = defaultPermissionMode
	}

	// W-4: 마이그레이션 배너 출력을 위해 raw 값 보존
	rawStatuslineMode := existingPrefs.StatuslineMode
	rawStatuslineTheme := existingPrefs.StatuslineTheme

	statuslineMode := normalizeStatuslineMode(existingPrefs.StatuslineMode)
	statuslineTheme := normalizeStatuslineTheme(existingPrefs.StatuslineTheme)

	// ====== Step 1: 언어 선택 ======
	langOptions := []huh.Option[string]{
		huh.NewOption("English", "en"),
		huh.NewOption("Korean (한국어)", "ko"),
		huh.NewOption("Japanese (日本語)", "ja"),
		huh.NewOption("Chinese (中文)", "zh"),
	}

	langForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your language").
				Description("Language for this wizard and Claude's responses.").
				Options(langOptions...).
				Value(&convLang),
		).Title("Language"),
	)

	if err := langForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Setup cancelled.")
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// ====== Step 2: 선택된 언어로 나머지 폼 표시 ======
	t := getProfileText(convLang)

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.ConfiguringProfile+"\n\n", profileName)

	// W-4: statusline mode/theme 마이그레이션 배너 — 폼 표시 전 출력
	if rawStatuslineMode != "" && rawStatuslineMode != statuslineMode {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.MigrationNoticeStatuslineMode+"\n", rawStatuslineMode, statuslineMode)
	}
	if rawStatuslineTheme != "" && rawStatuslineTheme != statuslineTheme {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.MigrationNoticeStatuslineTheme+"\n", rawStatuslineTheme, statuslineTheme)
	}

	form := huh.NewForm(
		// Section 1: 사용자 정보
		huh.NewGroup(
			huh.NewInput().
				Title(t.UserNameTitle).
				Description(t.UserNameDesc).
				Value(&userName),
		).Title(t.IdentityTitle),

		// Section 2: 언어 (대화 언어 이후)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.GitCommitLangTitle).
				Description(t.GitCommitLangDesc).
				Options(langOptions...).
				Value(&gitCommitLang),
			huh.NewSelect[string]().
				Title(t.CodeCommentLangTitle).
				Description(t.CodeCommentLangDesc).
				Options(langOptions...).
				Value(&codeCommentLang),
			huh.NewSelect[string]().
				Title(t.DocLangTitle).
				Description(t.DocLangDesc).
				Options(langOptions...).
				Value(&docLang),
		).Title(t.LanguagesTitle),

		// Section 3: 모델 설정 (모델 오버라이드 + 권한 모드)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.ModelOverrideTitle).
				Description(t.ModelOverrideDesc).
				Options(
					huh.NewOption(t.ModelDefault, ""),
					huh.NewOption(t.ModelOpus, "opus"),
					huh.NewOption(t.ModelOpus1M, "opus[1m]"),
					huh.NewOption(t.ModelSonnet, "sonnet"),
					huh.NewOption(t.ModelSonnet1M, "sonnet[1m]"),
					huh.NewOption(t.ModelHaiku, "haiku"),
					huh.NewOption(t.ModelOpusPlan, "opusplan"),
				).
				Value(&model),
			huh.NewSelect[string]().
				Title(t.EffortLevelTitle).
				Description(t.EffortLevelDesc).
				Options(
					huh.NewOption(t.EffortLevelDefault, ""),
					huh.NewOption(t.EffortLevelLow, "low"),
					huh.NewOption(t.EffortLevelMedium, "medium"),
					huh.NewOption(t.EffortLevelHigh, "high"),
					huh.NewOption(t.EffortLevelXHigh, "xhigh"),
					huh.NewOption(t.EffortLevelMax, "max"),
				).
				Value(&effortLevel),
			// S-4: 옵션 순서 — acceptEdits, auto, default, plan, bypass, dontAsk
			huh.NewSelect[string]().
				Title(t.PermissionModeTitle).
				Description(t.PermissionModeDesc).
				Options(
					huh.NewOption(t.PermAcceptEdits, "acceptEdits"),
					huh.NewOption(t.PermAuto, "auto"),
					huh.NewOption(t.PermDefault, "default"),
					huh.NewOption(t.PermPlan, "plan"),
					huh.NewOption(t.PermBypass, "bypassPermissions"),
					huh.NewOption(t.PermDontAsk, "dontAsk"),
				).
				Value(&permissionMode),
		).Title(t.ModelSettingsTitle),

		// Section 4: 화면 표시 — 모드, 테마
		// KEEP IN SYNC with statuslineModeCanonical at top of file
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t.StatuslineModeTitle).
				Description(t.StatuslineModeDesc).
				Options(
					huh.NewOption(t.ModeDefault, "default"),
					huh.NewOption(t.ModeFull, "full"),
				).
				Value(&statuslineMode),
			// KEEP IN SYNC with statuslineThemeCanonical at top of file
			huh.NewSelect[string]().
				Title(t.StatuslineThemeTitle).
				Description(t.StatuslineThemeDesc).
				Options(
					huh.NewOption(t.ThemeMoaiDark, "catppuccin-mocha"),
					huh.NewOption(t.ThemeMoaiLight, "catppuccin-latte"),
				).
				Value(&statuslineTheme),
		).Title(t.DisplayTitle),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), t.SetupCancelled)
			return nil
		}
		return fmt.Errorf("wizard error: %w", err)
	}

	// 권한 모드 정규화: "acceptEdits"는 프로젝트 기본값이므로 불필요한 override를 피해 빈 문자열로 저장
	if permissionMode == defaultPermissionMode {
		permissionMode = ""
	}

	// 설정 저장
	prefs := profile.ProfilePreferences{
		UserName:         userName,
		ConversationLang: convLang,
		GitCommitLang:    gitCommitLang,
		CodeCommentLang:  codeCommentLang,
		DocLang:          docLang,
		Model:            model,
		EffortLevel:      effortLevel,
		PermissionMode:   permissionMode,
		StatuslineMode:   statuslineMode,
		StatuslineTheme:  statuslineTheme,
	}

	if err := profile.WritePreferences(profileName, prefs); err != nil {
		return fmt.Errorf("save preferences: %w", err)
	}

	// MoAI 프로젝트 내에 있는 경우 프로젝트 설정에 동기화.
	// syncedProjectRoot가 설정되면 최종 리포트에 statusline.yaml 경로를 표시해
	// 사용자가 변경이 적용된 위치를 확인할 수 있다.
	var syncedProjectRoot string
	if cwd, err := os.Getwd(); err == nil {
		moaiDir := filepath.Join(cwd, ".moai")
		if info, err := os.Stat(moaiDir); err == nil && info.IsDir() {
			if err := profile.SyncToProjectConfig(cwd, prefs); err != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to sync profile to project config: %v\n", err)
			} else {
				syncedProjectRoot = cwd
			}
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), t.SavedProfile,
		profileName,
		profile.GetPreferencesPath(profileName))

	// 사용자가 캡처된 모든 값을 시각적으로 확인할 수 있도록 구조화된 요약 출력.
	printProfileSummary(cmd.OutOrStdout(), &t, &prefs, syncedProjectRoot)
	return nil
}

// printProfileSummary는 적용된 설정의 다중 라인 요약을 out에 출력한다.
// 동기화가 실행된 경우 해당 값을 보유한 프로젝트 레벨 YAML 경로도 함께 출력한다.
func printProfileSummary(out io.Writer, t *profileSetupText, prefs *profile.ProfilePreferences, syncedProjectRoot string) {
	// S-7: 7개 필드를 단일 Fprintf 호출로 통합
	_, _ = fmt.Fprintf(out,
		"%s\n"+
			"  %s: %s\n"+
			"  %s: %s / %s / %s / %s\n"+
			"  %s: %s\n"+
			"  %s: %s\n"+
			"  %s: %s\n"+
			"  %s: %s\n"+
			"  %s: %s\n",
		t.SummaryHeader,
		t.SummaryUserName, valueOrDash(prefs.UserName),
		t.SummaryLanguages,
		valueOrDash(prefs.ConversationLang),
		valueOrDash(prefs.GitCommitLang),
		valueOrDash(prefs.CodeCommentLang),
		valueOrDash(prefs.DocLang),
		t.SummaryModel, valueOrDefault(prefs.Model, t.SummaryDefault),
		t.SummaryEffort, valueOrDefault(prefs.EffortLevel, t.SummaryDefault),
		// S-2: normalize 후 StatuslineMode/Theme 는 항상 non-empty이므로 valueOrDefault 불필요
		t.SummaryPermission, valueOrDefault(prefs.PermissionMode, defaultPermissionMode),
		t.SummaryStatuslineMode, prefs.StatuslineMode,
		t.SummaryStatuslineTheme, prefs.StatuslineTheme,
	)

	if syncedProjectRoot != "" {
		// S-1: 상대 경로 출력 (syncedProjectRoot == cwd 이므로 상대 경로 하드코딩)
		_, _ = fmt.Fprintf(out, "\n%s\n", t.SummarySyncedHeader)
		_, _ = fmt.Fprintf(out, "  statusline.yaml -> .moai/config/sections/statusline.yaml\n")
		_, _ = fmt.Fprintf(out, "  language.yaml   -> .moai/config/sections/language.yaml\n")
	} else {
		_, _ = fmt.Fprintf(out, "\n%s\n", t.SummarySyncSkipped)
	}
}

// valueOrDash는 값이 비어 있을 때 "-"를 반환한다.
// 사용자 이름/언어 등 빈 값이 "설정 안 됨"을 의미하는 필드에 사용한다.
func valueOrDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

// valueOrDefault는 v가 비어 있을 때 fallback을 반환한다.
// 빈 문자열이 "런타임 기본값 사용"을 의미하는 슬롯에 사용한다.
func valueOrDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
