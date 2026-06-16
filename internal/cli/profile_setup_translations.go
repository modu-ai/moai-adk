package cli

// profileSetupText holds all translatable strings for the profile setup wizard.
//
// The StatuslineModeTitle/Desc, ModeDefault/Compact/Full/Verbose/Minimal,
// StatuslinePresetTitle/Desc, PresetFull/Compact/Minimal/Custom,
// SummaryStatuslineMode, and MigrationNoticeStatuslineMode fields were removed
// by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 (runtime mode was inert; named
// presets were redundant with the segment map). Theme keys survive.
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
	ModelOpus1M        string
	ModelSonnet        string
	ModelSonnet1M      string
	ModelHaiku         string
	ModelOpusPlan      string
	// Effort level selector
	EffortLevelTitle   string
	EffortLevelDesc    string
	EffortLevelDefault string
	EffortLevelLow     string
	EffortLevelMedium  string
	EffortLevelHigh    string
	EffortLevelXHigh   string
	EffortLevelMax     string
	// Permission mode (replaces legacy bypass)
	PermissionModeTitle string
	PermissionModeDesc  string
	PermDefault         string
	PermAcceptEdits     string
	PermPlan            string
	PermAuto            string
	PermBypass          string
	PermDontAsk         string

	// Section: Display (theme only — mode + preset removed by
	// SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001)
	DisplayTitle string

	// Statusline segments multi-select — now unconditional (the preset==custom
	// gate was removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001).
	// Order matches statuslineAllSegments slice in profile_setup.go (15 segments).
	StatuslineSegmentsTitle string
	StatuslineSegmentsDesc  string
	SegmentClaudeVersion    string
	SegmentContext          string
	SegmentDirectory        string
	SegmentEffortThinking   string
	SegmentGitBranch        string
	SegmentGitStatus        string
	SegmentMoaiVersion      string
	SegmentModel            string
	SegmentOutputStyle      string
	SegmentPR               string
	SegmentSessionTime      string
	SegmentTask             string
	SegmentUsage5h          string
	SegmentUsage7d          string
	SegmentWorktree         string

	// Statusline theme selector
	StatuslineThemeTitle string
	StatuslineThemeDesc  string
	ThemeMoaiDark        string
	ThemeMoaiLight       string

	// Messages
	SetupCancelled string
	SavedProfile   string

	// Final summary block (rendered after SavedProfile). SummaryStatuslineMode
	// was removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001.
	SummaryHeader          string
	SummaryUserName        string
	SummaryLanguages       string
	SummaryModel           string
	SummaryEffort          string
	SummaryPermission      string
	SummaryStatuslineTheme string
	SummaryDefault         string
	SummarySyncedHeader    string
	SummarySyncSkipped     string

	// W-4: statusline theme migration banner (previous value → new value). The
	// mode migration banner was removed by SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001.
	MigrationNoticeStatuslineTheme string

	// SPEC-WEB-CONSOLE-003: project-config selects (development_mode + git_convention).
	// These persist to project config (quality.yaml / git-convention.yaml), NOT the
	// profile store — parity with the web console "Project" fieldset.
	DevelopmentModeTitle string
	DevelopmentModeDesc  string
	DevelopmentModeDDD   string
	DevelopmentModeTDD   string
	GitConventionTitle   string
	GitConventionDesc    string
	// ProjectDefaultOption labels the empty "(project default)" option shared by
	// both project-config selects.
	ProjectDefaultOption string
}

// profileSetupTexts maps language code to translated UI strings.
var profileSetupTexts = map[string]profileSetupText{
	"en": {
		ConfiguringProfile:      "Configuring profile '%s'",
		LangSelectTitle:         "Select your language",
		LangSelectDesc:          "Language for this wizard and Claude's responses.",
		LangGroupTitle:          "Language",
		IdentityTitle:           "Identity",
		UserNameTitle:           "User name",
		UserNameDesc:            "Your display name for configuration files. Press Enter to skip.",
		LanguagesTitle:          "Languages",
		GitCommitLangTitle:      "Git commit message language",
		GitCommitLangDesc:       "Language for commit messages.",
		CodeCommentLangTitle:    "Code comment language",
		CodeCommentLangDesc:     "Language for code comments.",
		DocLangTitle:            "Documentation language",
		DocLangDesc:             "Language for documentation files.",
		ModelSettingsTitle:      "Model Settings",
		ModelPolicyTitle:        "Agent model policy",
		ModelPolicyDesc:         "Controls token consumption by assigning optimal models to each agent.",
		ModelPolicyHigh:         "High - Primarily Opus, best quality",
		ModelPolicyMedium:       "Medium - Primarily Sonnet with some Opus, balanced",
		ModelPolicyLow:          "Low - Sonnet and Haiku only, budget-friendly",
		ModelOverrideTitle:      "Default model override",
		ModelOverrideDesc:       "Override the model when launching with this profile.",
		ModelDefault:            "Default (no override)",
		ModelOpus:               "opus (Opus 4.7, adaptive thinking)",
		ModelOpus1M:             "opus[1m] (Opus 4.7 + 1M context)",
		ModelSonnet:             "sonnet (Sonnet 4.6, balanced)",
		ModelSonnet1M:           "sonnet[1m] (Sonnet 4.6 + 1M context)",
		ModelHaiku:              "haiku (Haiku 4.5, fastest)",
		ModelOpusPlan:           "opusplan (Opus planning, Sonnet coding)",
		EffortLevelTitle:        "Session effort level",
		EffortLevelDesc:         "Sets reasoning depth for this profile. xhigh/max require Opus 4.7.",
		EffortLevelDefault:      "Default (runtime default, xhigh for Opus 4.7)",
		EffortLevelLow:          "low - fastest, least thorough",
		EffortLevelMedium:       "medium - balanced",
		EffortLevelHigh:         "high - deep reasoning",
		EffortLevelXHigh:        "xhigh - extended reasoning (Opus 4.7+)",
		EffortLevelMax:          "max - maximum effort (Opus 4.7+)",
		PermissionModeTitle:     "Permission mode",
		PermissionModeDesc:      "Controls how Claude asks for permission before taking actions.",
		PermAcceptEdits:         "Auto accept edits - Auto-accept file edits, ask for commands",
		PermDefault:             "Ask permissions - Prompt for file edits and commands",
		PermPlan:                "Plan mode - Read-only exploration and planning",
		PermAuto:                "Auto mode (auto) - Classifier-gated approvals. REQUIRES Max/Team/Enterprise/API plan + Sonnet 4.6+. Session errors at runtime if unsupported.",
		PermBypass:              "Bypass permissions - Skip all checks (isolated environments only)",
		PermDontAsk:             "Don't ask - Only pre-approved tools (CI/locked-down environments)",
		DisplayTitle:            "Display",
		StatuslineSegmentsTitle: "Statusline segments",
		StatuslineSegmentsDesc:  "Toggle which segments appear in the status bar.",
		SegmentClaudeVersion:    "Claude version",
		SegmentContext:          "Context usage",
		SegmentDirectory:        "Current directory",
		SegmentEffortThinking:   "Effort + thinking mode",
		SegmentGitBranch:        "Git branch",
		SegmentGitStatus:        "Git status (porcelain)",
		SegmentMoaiVersion:      "MoAI version",
		SegmentModel:            "Model name",
		SegmentOutputStyle:      "Output style",
		SegmentPR:               "Open PR number",
		SegmentSessionTime:      "Session elapsed time",
		SegmentTask:             "Current task (/moai run XXX)",
		SegmentUsage5h:          "Usage 5h window bar",
		SegmentUsage7d:          "Usage 7d window bar",
		SegmentWorktree:         "Worktree path / identifier",
		StatuslineThemeTitle:    "Statusline Theme",
		StatuslineThemeDesc:     "Select a color theme for the statusline.",
		ThemeMoaiDark:           "MoAI Dark",
		ThemeMoaiLight:          "MoAI Light",
		SetupCancelled:          "Setup cancelled.",
		SavedProfile:            "\nSaved profile '%s':\n  Preferences → %s\n",

		SummaryHeader:          "Captured values:",
		SummaryUserName:        "User name",
		SummaryLanguages:       "Languages (conv/git/code/doc)",
		SummaryModel:           "Model",
		SummaryEffort:          "Effort level",
		SummaryPermission:      "Permission mode",
		SummaryStatuslineTheme: "Statusline theme",
		SummaryDefault:         "(runtime default)",
		SummarySyncedHeader:    "Synced to project config:",
		SummarySyncSkipped:     "No project-level sync (profile saved globally).",

		MigrationNoticeStatuslineTheme: "Notice: your previous statusline theme %q was migrated to %q in v3.",

		DevelopmentModeTitle: "Development mode",
		DevelopmentModeDesc:  "Project methodology written to quality.yaml. Empty keeps the project default.",
		DevelopmentModeDDD:   "ddd - Domain-Driven Development (ANALYZE-PRESERVE-IMPROVE)",
		DevelopmentModeTDD:   "tdd - Test-Driven Development (RED-GREEN-REFACTOR)",
		GitConventionTitle:   "Git commit convention",
		GitConventionDesc:    "Commit message convention written to git-convention.yaml. Empty keeps the project default.",
		ProjectDefaultOption: "(project default)",
	},
	"ko": {
		ConfiguringProfile:      "프로필 '%s' 설정",
		LangSelectTitle:         "언어를 선택하세요",
		LangSelectDesc:          "이 설정 마법사와 Claude 응답에 사용할 언어입니다.",
		LangGroupTitle:          "언어",
		IdentityTitle:           "사용자 정보",
		UserNameTitle:           "사용자 이름",
		UserNameDesc:            "설정 파일에 표시될 이름입니다. Enter를 눌러 건너뛰세요.",
		LanguagesTitle:          "언어 설정",
		GitCommitLangTitle:      "Git 커밋 메시지 언어",
		GitCommitLangDesc:       "커밋 메시지에 사용할 언어입니다.",
		CodeCommentLangTitle:    "코드 주석 언어",
		CodeCommentLangDesc:     "코드 주석에 사용할 언어입니다.",
		DocLangTitle:            "문서 언어",
		DocLangDesc:             "문서 파일에 사용할 언어입니다.",
		ModelSettingsTitle:      "모델 설정",
		ModelPolicyTitle:        "에이전트 모델 정책",
		ModelPolicyDesc:         "각 에이전트에 최적 모델을 할당하여 토큰 소비를 제어합니다.",
		ModelPolicyHigh:         "High - Opus 중심, 최고 품질",
		ModelPolicyMedium:       "Medium - Sonnet 중심 + 일부 Opus, 균형",
		ModelPolicyLow:          "Low - Sonnet/Haiku만 사용, 경제적",
		ModelOverrideTitle:      "기본 모델 오버라이드",
		ModelOverrideDesc:       "이 프로필로 실행할 때 모델을 오버라이드합니다.",
		ModelDefault:            "기본값 (오버라이드 없음)",
		ModelOpus:               "opus (Opus 4.7, 적응형 사고)",
		ModelOpus1M:             "opus[1m] (Opus 4.7 + 1M 컨텍스트)",
		ModelSonnet:             "sonnet (Sonnet 4.6, 균형)",
		ModelSonnet1M:           "sonnet[1m] (Sonnet 4.6 + 1M 컨텍스트)",
		ModelHaiku:              "haiku (Haiku 4.5, 최고 속도)",
		ModelOpusPlan:           "opusplan (Opus 기획, Sonnet 코딩)",
		EffortLevelTitle:        "세션 추론 강도",
		EffortLevelDesc:         "이 프로필의 추론 깊이를 설정합니다. xhigh/max는 Opus 4.7 필요.",
		EffortLevelDefault:      "기본값 (런타임 기본값, Opus 4.7은 xhigh)",
		EffortLevelLow:          "low - 가장 빠름, 간략한 추론",
		EffortLevelMedium:       "medium - 균형",
		EffortLevelHigh:         "high - 심층 추론",
		EffortLevelXHigh:        "xhigh - 확장 추론 (Opus 4.7+)",
		EffortLevelMax:          "max - 최대 추론 (Opus 4.7+)",
		PermissionModeTitle:     "권한 모드",
		PermissionModeDesc:      "Claude가 작업 수행 전 권한을 요청하는 방식을 제어합니다.",
		PermAcceptEdits:         "자동 편집 수락 (acceptEdits) - 파일 편집 자동 수락, 명령어만 확인",
		PermDefault:             "권한 요청 (default) - 파일 편집과 명령어에 대해 매번 확인",
		PermPlan:                "계획 모드 (plan) - 읽기 전용 탐색 및 계획",
		PermAuto:                "자동 모드 (auto) - 분류기 기반 자동 승인. Max/Team/Enterprise/API 플랜 + Sonnet 4.6+ 필수. 미지원 환경에서는 런타임 오류 발생.",
		PermBypass:              "권한 건너뛰기 (bypassPermissions) - 모든 검사 생략 (격리된 환경 전용)",
		PermDontAsk:             "묻지 않기 (dontAsk) - 사전 승인된 도구만 사용 (CI/잠금 환경)",
		DisplayTitle:            "화면 표시",
		StatuslineSegmentsTitle: "상태줄 세그먼트",
		StatuslineSegmentsDesc:  "상태 표시줄에 표시할 세그먼트를 선택하세요.",
		SegmentClaudeVersion:    "Claude 버전",
		SegmentContext:          "컨텍스트 사용량",
		SegmentDirectory:        "현재 디렉토리",
		SegmentEffortThinking:   "추론 강도 + 사고 모드",
		SegmentGitBranch:        "Git 브랜치",
		SegmentGitStatus:        "Git 상태 (porcelain)",
		SegmentMoaiVersion:      "MoAI 버전",
		SegmentModel:            "모델 이름",
		SegmentOutputStyle:      "출력 스타일",
		SegmentPR:               "열린 PR 번호",
		SegmentSessionTime:      "세션 경과 시간",
		SegmentTask:             "현재 작업 (/moai run XXX)",
		SegmentUsage5h:          "사용량 5시간 바",
		SegmentUsage7d:          "사용량 7일 바",
		SegmentWorktree:         "워크트리 경로 / 식별자",
		StatuslineThemeTitle:    "Statusline 테마",
		StatuslineThemeDesc:     "상태줄 색상 테마를 선택하세요.",
		ThemeMoaiDark:           "MoAI Dark",
		ThemeMoaiLight:          "MoAI Light",
		SetupCancelled:          "설정이 취소되었습니다.",
		SavedProfile:            "\n프로필 '%s' 저장 완료:\n  환경설정 → %s\n",

		SummaryHeader:          "저장된 설정값:",
		SummaryUserName:        "사용자 이름",
		SummaryLanguages:       "언어 (대화/커밋/주석/문서)",
		SummaryModel:           "모델",
		SummaryEffort:          "추론 강도",
		SummaryPermission:      "권한 모드",
		SummaryStatuslineTheme: "상태줄 테마",
		SummaryDefault:         "(런타임 기본값)",
		SummarySyncedHeader:    "프로젝트 설정에 동기화됨:",
		SummarySyncSkipped:     "프로젝트별 동기화 없음 (프로필만 저장됨).",

		MigrationNoticeStatuslineTheme: "알림: 이전 statusline 테마 %q 가 v3에서 %q 로 마이그레이션되었습니다.",

		DevelopmentModeTitle: "개발 방법론",
		DevelopmentModeDesc:  "quality.yaml에 기록되는 프로젝트 개발 방법론. 비워두면 프로젝트 기본값을 유지합니다.",
		DevelopmentModeDDD:   "ddd - 도메인 주도 개발 (ANALYZE-PRESERVE-IMPROVE)",
		DevelopmentModeTDD:   "tdd - 테스트 주도 개발 (RED-GREEN-REFACTOR)",
		GitConventionTitle:   "Git 커밋 컨벤션",
		GitConventionDesc:    "git-convention.yaml에 기록되는 커밋 메시지 컨벤션. 비워두면 프로젝트 기본값을 유지합니다.",
		ProjectDefaultOption: "(프로젝트 기본값)",
	},
	"ja": {
		ConfiguringProfile:      "プロファイル '%s' を設定",
		LangSelectTitle:         "言語を選択してください",
		LangSelectDesc:          "このウィザードとClaudeの応答に使用する言語です。",
		LangGroupTitle:          "言語",
		IdentityTitle:           "ユーザー情報",
		UserNameTitle:           "ユーザー名",
		UserNameDesc:            "設定ファイルに表示される名前です。Enterでスキップできます。",
		LanguagesTitle:          "言語設定",
		GitCommitLangTitle:      "Gitコミットメッセージ言語",
		GitCommitLangDesc:       "コミットメッセージに使用する言語です。",
		CodeCommentLangTitle:    "コードコメント言語",
		CodeCommentLangDesc:     "コードコメントに使用する言語です。",
		DocLangTitle:            "ドキュメント言語",
		DocLangDesc:             "ドキュメントファイルに使用する言語です。",
		ModelSettingsTitle:      "モデル設定",
		ModelPolicyTitle:        "エージェントモデルポリシー",
		ModelPolicyDesc:         "各エージェントに最適なモデルを割り当て、トークン消費を制御します。",
		ModelPolicyHigh:         "High - Opus中心、最高品質",
		ModelPolicyMedium:       "Medium - Sonnet中心 + 一部Opus、バランス",
		ModelPolicyLow:          "Low - Sonnet/Haikuのみ、予算節約",
		ModelOverrideTitle:      "デフォルトモデルオーバーライド",
		ModelOverrideDesc:       "このプロファイルで起動する際のモデルをオーバーライドします。",
		ModelDefault:            "デフォルト (オーバーライドなし)",
		ModelOpus:               "opus (Opus 4.7、適応型思考)",
		ModelOpus1M:             "opus[1m] (Opus 4.7 + 1Mコンテキスト)",
		ModelSonnet:             "sonnet (Sonnet 4.6、バランス)",
		ModelSonnet1M:           "sonnet[1m] (Sonnet 4.6 + 1Mコンテキスト)",
		ModelHaiku:              "haiku (Haiku 4.5、最速)",
		ModelOpusPlan:           "opusplan (Opus設計、Sonnetコーディング)",
		EffortLevelTitle:        "セッション推論レベル",
		EffortLevelDesc:         "このプロファイルの推論深度を設定します。xhigh/maxはOpus 4.7が必要。",
		EffortLevelDefault:      "デフォルト (ランタイムデフォルト、Opus 4.7はxhigh)",
		EffortLevelLow:          "low - 最速、簡易推論",
		EffortLevelMedium:       "medium - バランス",
		EffortLevelHigh:         "high - 深い推論",
		EffortLevelXHigh:        "xhigh - 拡張推論 (Opus 4.7+)",
		EffortLevelMax:          "max - 最大推論 (Opus 4.7+)",
		PermissionModeTitle:     "権限モード",
		PermissionModeDesc:      "Claudeがアクション実行前に権限を要求する方法を制御します。",
		PermAcceptEdits:         "編集を自動承認 (acceptEdits) - ファイル編集を自動承認、コマンドのみ確認",
		PermDefault:             "権限を確認 (default) - ファイル編集とコマンドの都度確認",
		PermPlan:                "プランモード (plan) - 読み取り専用の探索と計画",
		PermAuto:                "オートモード (auto) - 分類器による自動承認。Max/Team/Enterprise/APIプラン + Sonnet 4.6+ 必須。未対応環境では実行時エラーが発生します。",
		PermBypass:              "権限スキップ (bypassPermissions) - 全チェックを省略（隔離環境専用）",
		PermDontAsk:             "確認しない (dontAsk) - 事前承認済みツールのみ（CI/制限環境）",
		DisplayTitle:            "表示設定",
		StatuslineSegmentsTitle: "ステータスラインセグメント",
		StatuslineSegmentsDesc:  "ステータスバーに表示するセグメントを選択してください。",
		SegmentClaudeVersion:    "Claude バージョン",
		SegmentContext:          "コンテキスト使用量",
		SegmentDirectory:        "現在のディレクトリ",
		SegmentEffortThinking:   "推論強度 + 思考モード",
		SegmentGitBranch:        "Git ブランチ",
		SegmentGitStatus:        "Git ステータス (porcelain)",
		SegmentMoaiVersion:      "MoAI バージョン",
		SegmentModel:            "モデル名",
		SegmentOutputStyle:      "出力スタイル",
		SegmentPR:               "オープン PR 番号",
		SegmentSessionTime:      "セッション経過時間",
		SegmentTask:             "現在のタスク (/moai run XXX)",
		SegmentUsage5h:          "使用量 5h ウィンドウバー",
		SegmentUsage7d:          "使用量 7d ウィンドウバー",
		SegmentWorktree:         "ワークツリーパス / 識別子",
		StatuslineThemeTitle:    "ステータスラインテーマ",
		StatuslineThemeDesc:     "ステータスラインのカラーテーマを選択してください。",
		ThemeMoaiDark:           "MoAI Dark",
		ThemeMoaiLight:          "MoAI Light",
		SetupCancelled:          "セットアップがキャンセルされました。",
		SavedProfile:            "\nプロファイル '%s' を保存しました:\n  環境設定 → %s\n",

		SummaryHeader:          "保存された設定値:",
		SummaryUserName:        "ユーザー名",
		SummaryLanguages:       "言語 (会話/コミット/コメント/ドキュメント)",
		SummaryModel:           "モデル",
		SummaryEffort:          "推論レベル",
		SummaryPermission:      "権限モード",
		SummaryStatuslineTheme: "ステータスラインテーマ",
		SummaryDefault:         "(ランタイムデフォルト)",
		SummarySyncedHeader:    "プロジェクト設定に同期しました:",
		SummarySyncSkipped:     "プロジェクト別同期なし (プロファイルのみ保存).",

		MigrationNoticeStatuslineTheme: "お知らせ: 以前のステータスラインテーマ %q は v3 で %q に移行されました。",

		DevelopmentModeTitle: "開発方法論",
		DevelopmentModeDesc:  "quality.yaml に記録されるプロジェクトの開発方法論。空欄の場合はプロジェクトのデフォルトを維持します。",
		DevelopmentModeDDD:   "ddd - ドメイン駆動開発 (ANALYZE-PRESERVE-IMPROVE)",
		DevelopmentModeTDD:   "tdd - テスト駆動開発 (RED-GREEN-REFACTOR)",
		GitConventionTitle:   "Git コミット規約",
		GitConventionDesc:    "git-convention.yaml に記録されるコミットメッセージ規約。空欄の場合はプロジェクトのデフォルトを維持します。",
		ProjectDefaultOption: "(プロジェクトのデフォルト)",
	},
	"zh": {
		ConfiguringProfile:      "配置文件 '%s' 设置",
		LangSelectTitle:         "请选择语言",
		LangSelectDesc:          "用于此向导和Claude响应的语言。",
		LangGroupTitle:          "语言",
		IdentityTitle:           "用户信息",
		UserNameTitle:           "用户名",
		UserNameDesc:            "配置文件中显示的名称。按Enter跳过。",
		LanguagesTitle:          "语言设置",
		GitCommitLangTitle:      "Git提交消息语言",
		GitCommitLangDesc:       "提交消息使用的语言。",
		CodeCommentLangTitle:    "代码注释语言",
		CodeCommentLangDesc:     "代码注释使用的语言。",
		DocLangTitle:            "文档语言",
		DocLangDesc:             "文档文件使用的语言。",
		ModelSettingsTitle:      "模型设置",
		ModelPolicyTitle:        "代理模型策略",
		ModelPolicyDesc:         "通过为每个代理分配最优模型来控制token消耗。",
		ModelPolicyHigh:         "High - 以Opus为主，最佳质量",
		ModelPolicyMedium:       "Medium - 以Sonnet为主 + 部分Opus，均衡",
		ModelPolicyLow:          "Low - 仅Sonnet/Haiku，经济实惠",
		ModelOverrideTitle:      "默认模型覆盖",
		ModelOverrideDesc:       "使用此配置文件启动时覆盖模型。",
		ModelDefault:            "默认 (不覆盖)",
		ModelOpus:               "opus (Opus 4.7，自适应思考)",
		ModelOpus1M:             "opus[1m] (Opus 4.7 + 1M 上下文)",
		ModelSonnet:             "sonnet (Sonnet 4.6，均衡)",
		ModelSonnet1M:           "sonnet[1m] (Sonnet 4.6 + 1M 上下文)",
		ModelHaiku:              "haiku (Haiku 4.5，最快)",
		ModelOpusPlan:           "opusplan (Opus规划，Sonnet编码)",
		EffortLevelTitle:        "会话推理强度",
		EffortLevelDesc:         "设置此配置文件的推理深度。xhigh/max需要Opus 4.7。",
		EffortLevelDefault:      "默认 (运行时默认值，Opus 4.7为xhigh)",
		EffortLevelLow:          "low - 最快，简略推理",
		EffortLevelMedium:       "medium - 均衡",
		EffortLevelHigh:         "high - 深度推理",
		EffortLevelXHigh:        "xhigh - 扩展推理 (Opus 4.7+)",
		EffortLevelMax:          "max - 最大推理 (Opus 4.7+)",
		PermissionModeTitle:     "权限模式",
		PermissionModeDesc:      "控制Claude在执行操作前如何请求权限。",
		PermAcceptEdits:         "自动接受编辑 (acceptEdits) - 自动接受文件编辑，仅确认命令",
		PermDefault:             "请求权限 (default) - 每次文件编辑和命令都需确认",
		PermPlan:                "计划模式 (plan) - 只读探索和规划",
		PermAuto:                "自动模式 (auto) - 分类器把关自动批准。需要 Max/Team/Enterprise/API 计划 + Sonnet 4.6+。不支持时会产生运行时错误。",
		PermBypass:              "跳过权限 (bypassPermissions) - 跳过所有检查（仅限隔离环境）",
		PermDontAsk:             "不询问 (dontAsk) - 仅预批准工具（CI/锁定环境）",
		DisplayTitle:            "显示设置",
		StatuslineSegmentsTitle: "状态栏段位",
		StatuslineSegmentsDesc:  "选择要在状态栏中显示的段位。",
		SegmentClaudeVersion:    "Claude 版本",
		SegmentContext:          "上下文使用量",
		SegmentDirectory:        "当前目录",
		SegmentEffortThinking:   "推理强度 + 思考模式",
		SegmentGitBranch:        "Git 分支",
		SegmentGitStatus:        "Git 状态 (porcelain)",
		SegmentMoaiVersion:      "MoAI 版本",
		SegmentModel:            "模型名称",
		SegmentOutputStyle:      "输出样式",
		SegmentPR:               "开放 PR 编号",
		SegmentSessionTime:      "会话经过时间",
		SegmentTask:             "当前任务 (/moai run XXX)",
		SegmentUsage5h:          "使用量 5 小时窗口条",
		SegmentUsage7d:          "使用量 7 天窗口条",
		SegmentWorktree:         "工作树路径 / 标识符",
		StatuslineThemeTitle:    "状态栏主题",
		StatuslineThemeDesc:     "选择状态栏的颜色主题。",
		ThemeMoaiDark:           "MoAI Dark",
		ThemeMoaiLight:          "MoAI Light",
		SetupCancelled:          "设置已取消。",
		SavedProfile:            "\n配置文件 '%s' 已保存:\n  偏好设置 → %s\n",

		SummaryHeader:          "已输入的配置值:",
		SummaryUserName:        "用户名",
		SummaryLanguages:       "语言 (对话/提交/注释/文档)",
		SummaryModel:           "模型",
		SummaryEffort:          "推理强度",
		SummaryPermission:      "权限模式",
		SummaryStatuslineTheme: "状态栏主题",
		SummaryDefault:         "(运行时默认值)",
		SummarySyncedHeader:    "已同步到项目配置:",
		SummarySyncSkipped:     "未进行项目级同步 (仅保存全局配置文件).",

		MigrationNoticeStatuslineTheme: "提示：您之前的状态栏主题 %q 已在 v3 中迁移为 %q。",

		DevelopmentModeTitle: "开发方法论",
		DevelopmentModeDesc:  "写入 quality.yaml 的项目开发方法论。留空则保留项目默认值。",
		DevelopmentModeDDD:   "ddd - 领域驱动开发 (ANALYZE-PRESERVE-IMPROVE)",
		DevelopmentModeTDD:   "tdd - 测试驱动开发 (RED-GREEN-REFACTOR)",
		GitConventionTitle:   "Git 提交规范",
		GitConventionDesc:    "写入 git-convention.yaml 的提交信息规范。留空则保留项目默认值。",
		ProjectDefaultOption: "(项目默认值)",
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
