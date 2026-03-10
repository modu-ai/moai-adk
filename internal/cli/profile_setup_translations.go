package cli

// profileSetupText holds all translatable strings for the profile setup wizard.
type profileSetupText struct {
	// Banner
	ConfiguringProfile string

	// Language selection (first form)
	LangSelectTitle string
	LangSelectDesc  string
	LangGroupTitle  string

	// Section: Identity
	IdentityTitle string
	UserNameTitle string
	UserNameDesc  string

	// Section: Languages (remaining after conversation language)
	LanguagesTitle       string
	GitCommitLangTitle   string
	GitCommitLangDesc    string
	CodeCommentLangTitle string
	CodeCommentLangDesc  string
	DocLangTitle         string
	DocLangDesc          string

	// Section: Model Settings
	ModelSettingsTitle string
	ModelPolicyTitle   string
	ModelPolicyDesc    string
	ModelPolicyHigh    string
	ModelPolicyMedium  string
	ModelPolicyLow     string
	ModelOverrideTitle string
	ModelOverrideDesc  string
	ModelDefault       string
	ModelOpus          string
	ModelSonnet        string
	ModelHaiku         string
	ModelOpusPlan      string
	BypassTitle        string
	BypassDesc         string

	// Section: Display
	DisplayTitle string
	// Statusline mode selector (layout style)
	StatuslineModeTitle string
	StatuslineModeDesc  string
	// v3 mode labels (REQ-V3-MODE-003)
	ModeDefault string // label for mode = "default"
	ModeCompact string // label for mode = "compact"
	ModeFull    string // label for mode = "full"
	// Deprecated: v2 labels. Kept for backward compatibility.
	ModeVerbose string // label for mode = "verbose" (deprecated)
	ModeMinimal string // label for mode = "minimal" (deprecated)

	// Statusline theme selector
	StatuslineThemeTitle string
	StatuslineThemeDesc  string
	ThemeMoaiDark        string
	ThemeMoaiLight       string

	// Messages
	SetupCancelled string
	SavedProfile   string
}

// profileSetupTexts maps language code to translated UI strings.
var profileSetupTexts = map[string]profileSetupText{
	"en": {
		ConfiguringProfile:   "Configuring profile '%s'",
		LangSelectTitle:      "Select your language",
		LangSelectDesc:       "Language for this wizard and Claude's responses.",
		LangGroupTitle:       "Language",
		IdentityTitle:        "Identity",
		UserNameTitle:        "User name",
		UserNameDesc:         "Your display name for configuration files. Press Enter to skip.",
		LanguagesTitle:       "Languages",
		GitCommitLangTitle:   "Git commit message language",
		GitCommitLangDesc:    "Language for commit messages.",
		CodeCommentLangTitle: "Code comment language",
		CodeCommentLangDesc:  "Language for code comments.",
		DocLangTitle:         "Documentation language",
		DocLangDesc:          "Language for documentation files.",
		ModelSettingsTitle:   "Model Settings",
		ModelPolicyTitle:     "Agent model policy",
		ModelPolicyDesc:      "Controls token consumption by assigning optimal models to each agent.",
		ModelPolicyHigh:      "High - Primarily Opus, best quality",
		ModelPolicyMedium:    "Medium - Primarily Sonnet with some Opus, balanced",
		ModelPolicyLow:       "Low - Sonnet and Haiku only, budget-friendly",
		ModelOverrideTitle:   "Default model override",
		ModelOverrideDesc:    "Override the model when launching with this profile.",
		ModelDefault:         "Default (no override)",
		ModelOpus:            "claude-opus-4-6 (most capable)",
		ModelSonnet:          "claude-sonnet-4-6 (balanced)",
		ModelHaiku:           "claude-haiku-4-5 (fastest)",
		ModelOpusPlan:        "opusplan (Opus planning, Sonnet coding)",
		BypassTitle:          "Skip permission checks?",
		BypassDesc:           "Adds --dangerously-skip-permissions. Only use in trusted environments.",
		DisplayTitle:         "Display",
		StatuslineModeTitle:  "Statusline display mode",
		StatuslineModeDesc:   "Controls the layout style of the statusline.",
		ModeDefault:          "Default - 3-line: info, CW/5H/7D bars, dir+git",
		ModeCompact:          "Compact - 2-line: model+CW bar, git status",
		ModeFull:             "Full - 5-line: info, CW/5H/7D bars (40-block), dir+git",
		ModeVerbose:          "Verbose - 3-line detailed view with cost tracking",
		ModeMinimal:          "Minimal - Model and context only",
		StatuslineThemeTitle: "Statusline Theme",
		StatuslineThemeDesc:  "Select a color theme for the statusline.",
		ThemeMoaiDark:        "MoAI Dark",
		ThemeMoaiLight:       "MoAI Light",
		SetupCancelled:       "Setup cancelled.",
		SavedProfile:         "\nSaved profile '%s':\n  Preferences → %s\n",
	},
	"ko": {
		ConfiguringProfile:   "프로필 '%s' 설정",
		LangSelectTitle:      "언어를 선택하세요",
		LangSelectDesc:       "이 설정 마법사와 Claude 응답에 사용할 언어입니다.",
		LangGroupTitle:       "언어",
		IdentityTitle:        "사용자 정보",
		UserNameTitle:        "사용자 이름",
		UserNameDesc:         "설정 파일에 표시될 이름입니다. Enter를 눌러 건너뛰세요.",
		LanguagesTitle:       "언어 설정",
		GitCommitLangTitle:   "Git 커밋 메시지 언어",
		GitCommitLangDesc:    "커밋 메시지에 사용할 언어입니다.",
		CodeCommentLangTitle: "코드 주석 언어",
		CodeCommentLangDesc:  "코드 주석에 사용할 언어입니다.",
		DocLangTitle:         "문서 언어",
		DocLangDesc:          "문서 파일에 사용할 언어입니다.",
		ModelSettingsTitle:   "모델 설정",
		ModelPolicyTitle:     "에이전트 모델 정책",
		ModelPolicyDesc:      "각 에이전트에 최적 모델을 할당하여 토큰 소비를 제어합니다.",
		ModelPolicyHigh:      "High - Opus 중심, 최고 품질",
		ModelPolicyMedium:    "Medium - Sonnet 중심 + 일부 Opus, 균형",
		ModelPolicyLow:       "Low - Sonnet/Haiku만 사용, 경제적",
		ModelOverrideTitle:   "기본 모델 오버라이드",
		ModelOverrideDesc:    "이 프로필로 실행할 때 모델을 오버라이드합니다.",
		ModelDefault:         "기본값 (오버라이드 없음)",
		ModelOpus:            "claude-opus-4-6 (최고 성능)",
		ModelSonnet:          "claude-sonnet-4-6 (균형)",
		ModelHaiku:           "claude-haiku-4-5 (최고 속도)",
		ModelOpusPlan:        "opusplan (Opus 기획, Sonnet 코딩)",
		BypassTitle:          "권한 검사 건너뛰기?",
		BypassDesc:           "--dangerously-skip-permissions를 추가합니다. 신뢰할 수 있는 환경에서만 사용하세요.",
		DisplayTitle:         "화면 표시",
		StatuslineModeTitle:  "상태줄 표시 모드",
		StatuslineModeDesc:   "상태줄의 레이아웃 스타일을 제어합니다.",
		ModeDefault:          "Default - 3줄: 정보, CW/5H/7D 바, 디렉토리+git",
		ModeCompact:          "Compact - 2줄: 모델+CW 바, git 상태",
		ModeFull:             "Full - 5줄: 정보, CW/5H/7D 바(40블록), 디렉토리+git",
		ModeVerbose:          "Verbose - 비용 추적이 포함된 3줄 상세 뷰",
		ModeMinimal:          "Minimal - 모델과 컨텍스트만 표시",
		StatuslineThemeTitle: "Statusline 테마",
		StatuslineThemeDesc:  "상태줄 색상 테마를 선택하세요.",
		ThemeMoaiDark:        "MoAI Dark",
		ThemeMoaiLight:       "MoAI Light",
		SetupCancelled:       "설정이 취소되었습니다.",
		SavedProfile:         "\n프로필 '%s' 저장 완료:\n  환경설정 → %s\n",
	},
	"ja": {
		ConfiguringProfile:   "プロファイル '%s' を設定",
		LangSelectTitle:      "言語を選択してください",
		LangSelectDesc:       "このウィザードとClaudeの応答に使用する言語です。",
		LangGroupTitle:       "言語",
		IdentityTitle:        "ユーザー情報",
		UserNameTitle:        "ユーザー名",
		UserNameDesc:         "設定ファイルに表示される名前です。Enterでスキップできます。",
		LanguagesTitle:       "言語設定",
		GitCommitLangTitle:   "Gitコミットメッセージ言語",
		GitCommitLangDesc:    "コミットメッセージに使用する言語です。",
		CodeCommentLangTitle: "コードコメント言語",
		CodeCommentLangDesc:  "コードコメントに使用する言語です。",
		DocLangTitle:         "ドキュメント言語",
		DocLangDesc:          "ドキュメントファイルに使用する言語です。",
		ModelSettingsTitle:   "モデル設定",
		ModelPolicyTitle:     "エージェントモデルポリシー",
		ModelPolicyDesc:      "各エージェントに最適なモデルを割り当て、トークン消費を制御します。",
		ModelPolicyHigh:      "High - Opus中心、最高品質",
		ModelPolicyMedium:    "Medium - Sonnet中心 + 一部Opus、バランス",
		ModelPolicyLow:       "Low - Sonnet/Haikuのみ、予算節約",
		ModelOverrideTitle:   "デフォルトモデルオーバーライド",
		ModelOverrideDesc:    "このプロファイルで起動する際のモデルをオーバーライドします。",
		ModelDefault:         "デフォルト (オーバーライドなし)",
		ModelOpus:            "claude-opus-4-6 (最高性能)",
		ModelSonnet:          "claude-sonnet-4-6 (バランス)",
		ModelHaiku:           "claude-haiku-4-5 (最速)",
		ModelOpusPlan:        "opusplan (Opus設計、Sonnetコーディング)",
		BypassTitle:          "権限チェックをスキップしますか？",
		BypassDesc:           "--dangerously-skip-permissionsを追加します。信頼できる環境でのみ使用してください。",
		DisplayTitle:         "表示設定",
		StatuslineModeTitle:  "ステータスライン表示モード",
		StatuslineModeDesc:   "ステータスラインのレイアウトスタイルを制御します。",
		ModeDefault:          "Default - 3行: 情報、CW/5H/7Dバー、ディレクトリ+git",
		ModeCompact:          "Compact - 2行: モデル+CWバー、gitステータス",
		ModeFull:             "Full - 5行: 情報、CW/5H/7Dバー(40ブロック)、ディレクトリ+git",
		ModeVerbose:          "Verbose - コスト追跡付きの3行詳細表示",
		ModeMinimal:          "Minimal - モデルとコンテキストのみ",
		StatuslineThemeTitle: "ステータスラインテーマ",
		StatuslineThemeDesc:  "ステータスラインのカラーテーマを選択してください。",
		ThemeMoaiDark:        "MoAI Dark",
		ThemeMoaiLight:       "MoAI Light",
		SetupCancelled:       "セットアップがキャンセルされました。",
		SavedProfile:         "\nプロファイル '%s' を保存しました:\n  環境設定 → %s\n",
	},
	"zh": {
		ConfiguringProfile:   "配置文件 '%s' 设置",
		LangSelectTitle:      "请选择语言",
		LangSelectDesc:       "用于此向导和Claude响应的语言。",
		LangGroupTitle:       "语言",
		IdentityTitle:        "用户信息",
		UserNameTitle:        "用户名",
		UserNameDesc:         "配置文件中显示的名称。按Enter跳过。",
		LanguagesTitle:       "语言设置",
		GitCommitLangTitle:   "Git提交消息语言",
		GitCommitLangDesc:    "提交消息使用的语言。",
		CodeCommentLangTitle: "代码注释语言",
		CodeCommentLangDesc:  "代码注释使用的语言。",
		DocLangTitle:         "文档语言",
		DocLangDesc:          "文档文件使用的语言。",
		ModelSettingsTitle:   "模型设置",
		ModelPolicyTitle:     "代理模型策略",
		ModelPolicyDesc:      "通过为每个代理分配最优模型来控制token消耗。",
		ModelPolicyHigh:      "High - 以Opus为主，最佳质量",
		ModelPolicyMedium:    "Medium - 以Sonnet为主 + 部分Opus，均衡",
		ModelPolicyLow:       "Low - 仅Sonnet/Haiku，经济实惠",
		ModelOverrideTitle:   "默认模型覆盖",
		ModelOverrideDesc:    "使用此配置文件启动时覆盖模型。",
		ModelDefault:         "默认 (不覆盖)",
		ModelOpus:            "claude-opus-4-6 (最强性能)",
		ModelSonnet:          "claude-sonnet-4-6 (均衡)",
		ModelHaiku:           "claude-haiku-4-5 (最快)",
		ModelOpusPlan:        "opusplan (Opus规划，Sonnet编码)",
		BypassTitle:          "跳过权限检查？",
		BypassDesc:           "添加 --dangerously-skip-permissions。仅在可信环境中使用。",
		DisplayTitle:         "显示设置",
		StatuslineModeTitle:  "状态栏显示模式",
		StatuslineModeDesc:   "控制状态栏的布局样式。",
		ModeDefault:          "Default - 3行: 信息、CW/5H/7D栏、目录+git",
		ModeCompact:          "Compact - 2行: 模型+CW栏、git状态",
		ModeFull:             "Full - 5行: 信息、CW/5H/7D栏(40块)、目录+git",
		ModeVerbose:          "Verbose - 含费用追踪的3行详细视图",
		ModeMinimal:          "Minimal - 仅显示模型和上下文",
		StatuslineThemeTitle: "状态栏主题",
		StatuslineThemeDesc:  "选择状态栏的颜色主题。",
		ThemeMoaiDark:        "MoAI Dark",
		ThemeMoaiLight:       "MoAI Light",
		SetupCancelled:       "设置已取消。",
		SavedProfile:         "\n配置文件 '%s' 已保存:\n  偏好设置 → %s\n",
	},
}

// getProfileText returns translated UI strings for the given language.
// Falls back to English if the language is not supported.
func getProfileText(lang string) profileSetupText {
	if t, ok := profileSetupTexts[lang]; ok {
		return t
	}
	return profileSetupTexts["en"]
}
