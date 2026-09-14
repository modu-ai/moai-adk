package wizard

// QuestionTranslation holds translated strings for a question.
type QuestionTranslation struct {
	Title       string
	Description string
	Options     []OptionTranslation
}

// OptionTranslation holds translated strings for an option.
type OptionTranslation struct {
	Label string
	Desc  string
}

// UIStrings holds translated UI strings.
type UIStrings struct {
	ErrorRequired string
	// ConfirmYes / ConfirmNo localize the huh Confirm affirmative/negative
	// button labels. huh v2's Confirm exposes only static Affirmative/Negative
	// setters (no *Func variant), so these are applied once at field-build time
	// from the locale then in effect — they do NOT re-render when the user
	// changes the conversation language mid-wizard (see buildConfirmField).
	ConfirmYes string
	ConfirmNo  string
}

// translations maps language code -> question ID -> translation.
var translations = map[string]map[string]QuestionTranslation{
	"ko": {
		"conversation_language": {
			Title:       "대화 언어 선택",
			Description: "MoAI가 대화할 때 사용하는 언어입니다. 마법사가 즉시 해당 언어로 전환됩니다.",
			// Options intentionally omitted: language labels stay in their native form.
		},
		"user_name": {
			Title:       "이름 입력",
			Description: "MoAI가 부를 이름입니다. user.yaml(user.name)에 저장됩니다. 비워두면 건너뜁니다.",
		},
		"project_name": {
			Title:       "프로젝트 이름 입력",
			Description: "프로젝트의 이름입니다.",
		},
		"report_format": {
			Title:       "리포트 형식 선택",
			Description: "리포트를 HTML+마크다운으로 생성할지, 마크다운만 생성할지 설정합니다.",
			Options: []OptionTranslation{
				{Label: "HTML + 마크다운 (권장)", Desc: "브라우저에서 볼 수 있는 HTML 리포트와 마크다운을 모두 생성"},
				{Label: "마크다운만", Desc: "마크다운 리포트만 생성 (가볍고 diff 친화적)"},
			},
		},
		"git_mode": {
			Title:       "Git 자동화 모드 선택",
			Description: "Claude가 수행할 수 있는 Git 작업 범위를 설정합니다.",
			Options: []OptionTranslation{
				{Label: "Manual", Desc: "AI가 커밋이나 푸시를 하지 않음"},
				{Label: "Personal", Desc: "AI가 브랜치 생성 및 커밋 가능"},
				{Label: "Team", Desc: "AI가 브랜치 생성, 커밋, PR 생성 가능"},
			},
		},
		"git_provider": {
			Title:       "Git 프로바이더 선택",
			Description: "프로젝트의 Git 호스팅 플랫폼을 선택합니다.",
			Options: []OptionTranslation{
				{Label: "GitHub", Desc: "GitHub.com"},
				{Label: "GitLab", Desc: "GitLab.com 또는 자체 호스팅 GitLab"},
			},
		},
		"gitlab_instance_url": {
			Title:       "GitLab 인스턴스 URL 입력",
			Description: "GitLab.com은 https://gitlab.com을 사용합니다. 자체 호스팅인 경우 인스턴스 URL을 입력하세요.",
		},
		"github_username": {
			Title:       "GitHub 사용자명 입력",
			Description: "Git 자동화 기능에 필요합니다.",
		},
		"github_token": {
			Title:       "GitHub 개인 액세스 토큰 입력 (선택)",
			Description: "PR 생성 및 푸시에 필요합니다. 비워두어 건너거나 gh CLI를 사용하세요.",
		},
		"gitlab_username": {
			Title:       "GitLab 사용자명 입력",
			Description: "GitLab Git 자동화 기능에 필요합니다.",
		},
		"gitlab_token": {
			Title:       "GitLab 개인 액세스 토큰 입력 (선택사항)",
			Description: "MR 생성 및 푸시에 필요합니다. 비워두거나 glab CLI를 사용할 수 있습니다.",
		},
		"model_policy": {
			Title:       "모델 정책 선택",
			Description: "각 에이전트에 할당되는 Claude 모델 등급을 제어합니다. Claude 플랜에 맞추세요.",
			Options: []OptionTranslation{
				{Label: "Max", Desc: "Opus 5 (high~medium) + Sonnet (low, 문서/단발성 작업) — Max $200 플랜"},
				{Label: "Medium (권장)", Desc: "Opus 5 (high~low) + Sonnet (low, 문서/단발성 작업) — Max $100 플랜"},
				{Label: "Low", Desc: "Opus 5 (high~low) + Sonnet (low, 문서/E2E/단발성 작업) — Plus $20 플랜"},
			},
		},
		"autonomy_tier": {
			Title:       "세션 권한 모드 선택",
			Description: "Claude Code 권한 모드를 사용자 설정에 기록합니다. '편집 자동 수락'이 권장 기본값입니다.",
			Options: []OptionTranslation{
				{Label: "편집 자동 수락 (권장)", Desc: "파일 편집은 자동 수락; 다른 도구는 확인"},
				{Label: "자동 모드", Desc: "분류기 안전 검사 하에 도구 호출 자동 승인"},
				{Label: "권한 우회", Desc: "모든 프롬프트 생략; 샌드박스 증명 필요 (Docker/gVisor 등)"},
			},
		},
		"agent_wiring": {
			Title:       "배포하고 연결할 에이전트 하니스 선택",
			Description: "이 프로젝트에 MoAI가 배포하고 연결할 LLM 하니스입니다. 'claude'가 권장 기본값이며, --llm 플래그가 이 답변보다 우선합니다.",
			Options: []OptionTranslation{
				{Label: "Claude 단독 (권장)", Desc: ".claude/ 표면과 AGENTS.md를 배포합니다 (지금까지의 기본 동작)"},
				{Label: "Codex 단독", Desc: "AGENTS.md와 Codex 표면만 배포합니다 — .claude/ 디렉터, CLAUDE.md, .mcp.json이 생기지 않습니다"},
				{Label: "Claude + Codex", Desc: "동일한 .claude/ 배포에 .codex/ 연결을 더하고 .mcp.json 프로비저닝을 강제로 켭니다"},
			},
		},
	},
	"ja": {
		"conversation_language": {
			Title:       "会話言語を選択",
			Description: "MoAIが会話に使用する言語です。ウィザードは直ちにその言語に切り替わります。",
		},
		"user_name": {
			Title:       "お名前を入力",
			Description: "MoAIがあなたを呼ぶ名前です。user.yaml（user.name）に保存されます。空欄でスキップできます。",
		},
		"project_name": {
			Title:       "プロジェクト名を入力",
			Description: "プロジェクトの名前です。",
		},
		"report_format": {
			Title:       "レポート形式を選択",
			Description: "レポートをHTML+Markdownで生成するか、Markdownのみで生成するかを設定します。",
			Options: []OptionTranslation{
				{Label: "HTML + Markdown (推奨)", Desc: "ブラウザで表示可能なHTMLレポートとMarkdownの両方を生成"},
				{Label: "Markdownのみ", Desc: "Markdownレポートのみ生成（軽量でdiffに優しい）"},
			},
		},
		"git_mode": {
			Title:       "Git自動化モードを選択",
			Description: "Claudeが実行できるGit操作の範囲を設定します。",
			Options: []OptionTranslation{
				{Label: "Manual", Desc: "AIはコミットやプッシュを行わない"},
				{Label: "Personal", Desc: "AIがブランチ作成とコミットが可能"},
				{Label: "Team", Desc: "AIがブランチ作成、コミット、PR作成が可能"},
			},
		},
		"git_provider": {
			Title:       "Gitプロバイダーを選択",
			Description: "プロジェクトのGitホスティングプラットフォームを選択します。",
			Options: []OptionTranslation{
				{Label: "GitHub", Desc: "GitHub.com"},
				{Label: "GitLab", Desc: "GitLab.comまたはセルフホストGitLab"},
			},
		},
		"gitlab_instance_url": {
			Title:       "GitLabインスタンスURLを入力",
			Description: "GitLab.comはhttps://gitlab.comを使用します。セルフホストの場合はインスタンスURLを入力してください。",
		},
		"github_username": {
			Title:       "GitHubユーザー名を入力",
			Description: "Git自動化機能に必要です。",
		},
		"github_token": {
			Title:       "GitHubパーソナルアクセストークンを入力（省略可）",
			Description: "PR作成とプッシュに必要です。空欄のままスキップまたはgh CLIを使用してください。",
		},
		"gitlab_username": {
			Title:       "GitLabユーザー名を入力",
			Description: "GitLab Git自動化機能に必要です。",
		},
		"gitlab_token": {
			Title:       "GitLabパーソナルアクセストークンを入力（省略可）",
			Description: "MR作成とプッシュに必要です。空欄のままスキップまたはglab CLIを使用してください。",
		},
		"model_policy": {
			Title:       "モデルポリシーを選択",
			Description: "各エージェントに割り当てる Claude モデルのティアを制御します。ご利用の Claude プランに合わせてください。",
			Options: []OptionTranslation{
				{Label: "Max", Desc: "Opus 5 (high~medium) + Sonnet (low, ドキュメント/単発タスク) — Max $200 プラン"},
				{Label: "Medium (推奨)", Desc: "Opus 5 (high~low) + Sonnet (low, ドキュメント/単発タスク) — Max $100 プラン"},
				{Label: "Low", Desc: "Opus 5 (high~low) + Sonnet (low, ドキュメント/E2E/単発タスク) — Plus $20 プラン"},
			},
		},
		"autonomy_tier": {
			Title:       "セッションの権限モードを選択",
			Description: "Claude Code の権限モードをユーザー設定に書き込みます。「編集を自動承認」が推奨デフォルトです。",
			Options: []OptionTranslation{
				{Label: "編集を自動承認 (推奨)", Desc: "ファイル編集は自動承認; その他のツールは確認"},
				{Label: "自動モード", Desc: "分類器の安全検査のもとでツール呼び出しを自動承認"},
				{Label: "権限をバイパス", Desc: "すべてのプロンプトを省略; サンドボックス証明が必要 (Docker/gVisor 等)"},
			},
		},
		"agent_wiring": {
			Title:       "배포하고 접속할 에이전트 하니스를 선택",
			Description: "このプロジェクトに MoAI がデプロイ・接続する LLM ハーネスです。'claude' が推奨デフォルトで、--llm フラグがこの回答より優先されます。",
			Options: []OptionTranslation{
				{Label: "Claude のみ (推奨)", Desc: ".claude/ サーフェスと AGENTS.md をデプロイします (従来のデフォルト動作)"},
				{Label: "Codex のみ", Desc: "AGENTS.md と Codex サーフェスのみデプロイ — .claude/ ツリー、CLAUDE.md、.mcp.json は作成されません"},
				{Label: "Claude + Codex", Desc: "同じ .claude/ デプロイに .codex/ 接続を追加し、.mcp.json のプロビジョニングを強制有効化"},
			},
		},
	},
	"zh": {
		"conversation_language": {
			Title:       "选择对话语言",
			Description: "MoAI 与您交流时使用的语言。向导会立即切换到该语言。",
		},
		"user_name": {
			Title:       "输入您的姓名",
			Description: "MoAI 对您的称呼。保存到 user.yaml（user.name）。留空则跳过。",
		},
		"project_name": {
			Title:       "输入项目名称",
			Description: "项目的名称。",
		},
		"report_format": {
			Title:       "选择报告格式",
			Description: "控制报告生成为HTML+Markdown还是仅Markdown。",
			Options: []OptionTranslation{
				{Label: "HTML + Markdown (推荐)", Desc: "生成可在浏览器查看的HTML报告和Markdown"},
				{Label: "仅Markdown", Desc: "仅生成Markdown报告（更轻量，利于diff）"},
			},
		},
		"git_mode": {
			Title:       "选择Git自动化模式",
			Description: "设置Claude可以执行的Git操作范围。",
			Options: []OptionTranslation{
				{Label: "Manual", Desc: "AI不进行提交或推送"},
				{Label: "Personal", Desc: "AI可以创建分支和提交"},
				{Label: "Team", Desc: "AI可以创建分支、提交和创建PR"},
			},
		},
		"git_provider": {
			Title:       "选择Git提供商",
			Description: "选择项目的Git托管平台。",
			Options: []OptionTranslation{
				{Label: "GitHub", Desc: "GitHub.com"},
				{Label: "GitLab", Desc: "GitLab.com或自托管GitLab"},
			},
		},
		"gitlab_instance_url": {
			Title:       "输入GitLab实例URL",
			Description: "GitLab.com请使用https://gitlab.com。自托管请输入实例URL。",
		},
		"github_username": {
			Title:       "输入GitHub用户名",
			Description: "Git自动化功能所需。",
		},
		"github_token": {
			Title:       "输入GitHub个人访问令牌（可选）",
			Description: "PR创建和推送所需。留空以跳过或使用gh CLI。",
		},
		"gitlab_username": {
			Title:       "输入GitLab用户名",
			Description: "GitLab Git自动化功能所需。",
		},
		"gitlab_token": {
			Title:       "输入GitLab个人访问令牌（可选）",
			Description: "MR创建和推送所需。留空以跳过或使用glab CLI。",
		},
		"model_policy": {
			Title:       "选择模型策略",
			Description: "控制为每个智能体分配的 Claude 模型等级。请与您的 Claude 套餐匹配。",
			Options: []OptionTranslation{
				{Label: "Max", Desc: "Opus 5 (high~medium) + Sonnet (low, 文档/一次性任务) — Max $200 套餐"},
				{Label: "Medium (推荐)", Desc: "Opus 5 (high~low) + Sonnet (low, 文档/一次性任务) — Max $100 套餐"},
				{Label: "Low", Desc: "Opus 5 (high~low) + Sonnet (low, 文档/E2E/一次性任务) — Plus $20 套餐"},
			},
		},
		"autonomy_tier": {
			Title:       "选择会话权限模式",
			Description: "将 Claude Code 权限模式写入用户设置。「自动接受编辑」是推荐默认值。",
			Options: []OptionTranslation{
				{Label: "自动接受编辑 (推荐)", Desc: "自动接受文件编辑;其他工具仍需确认"},
				{Label: "自动模式", Desc: "在分类器安全检查下自动批准工具调用"},
				{Label: "跳过权限检查", Desc: "跳过所有提示;需要沙箱证明 (Docker/gVisor 等)"},
			},
		},
		"agent_wiring": {
			Title:       "选择要部署并接入的代理框架",
			Description: "MoAI 为本项目部署并接入的 LLM 框架。'claude' 是推荐默认值，--llm 参数优先于此答案。",
			Options: []OptionTranslation{
				{Label: "仅 Claude (推荐)", Desc: "部署 .claude/ 表面与 AGENTS.md（沿用至今的默认行为）"},
				{Label: "仅 Codex", Desc: "仅部署 AGENTS.md 与 Codex 表面 — 不会生成 .claude/ 目录、CLAUDE.md 和 .mcp.json"},
				{Label: "Claude + Codex", Desc: "在相同的 .claude/ 部署之上追加 .codex/ 接入，并强制开启 .mcp.json 供应"},
			},
		},
	},
}

// uiStrings maps language code to UI strings.
var uiStrings = map[string]UIStrings{
	"en": {
		ErrorRequired: "This field is required",
		ConfirmYes:    "Yes",
		ConfirmNo:     "No",
	},
	"ko": {
		ErrorRequired: "필수 입력 항목입니다",
		ConfirmYes:    "예",
		ConfirmNo:     "아니오",
	},
	"ja": {
		ErrorRequired: "この項目は必須です",
		ConfirmYes:    "はい",
		ConfirmNo:     "いいえ",
	},
	"zh": {
		ErrorRequired: "此字段为必填项",
		ConfirmYes:    "是",
		ConfirmNo:     "否",
	},
}

// GetLocalizedQuestion returns a localized copy of the question.
// If no translation exists for the locale, returns the original question.
func GetLocalizedQuestion(q *Question, locale string) Question {
	// English is the default, no translation needed
	if locale == "en" || locale == "" {
		return *q
	}

	langTranslations, ok := translations[locale]
	if !ok {
		return *q
	}

	trans, ok := langTranslations[q.ID]
	if !ok {
		return *q
	}

	// Create a copy with translated strings
	localized := *q
	if trans.Title != "" {
		localized.Title = trans.Title
	}
	if trans.Description != "" {
		localized.Description = trans.Description
	}

	// Translate options if available
	if len(trans.Options) > 0 && len(q.Options) == len(trans.Options) {
		localized.Options = make([]Option, len(q.Options))
		for i, opt := range q.Options {
			localized.Options[i] = Option{
				Label: trans.Options[i].Label,
				Value: opt.Value, // Keep original value
				Desc:  trans.Options[i].Desc,
			}
			// Use original if translation is empty
			if localized.Options[i].Label == "" {
				localized.Options[i].Label = opt.Label
			}
			if localized.Options[i].Desc == "" {
				localized.Options[i].Desc = opt.Desc
			}
		}
	}

	return localized
}

// GetUIStrings returns UI strings for the given locale.
// Returns English strings if locale is not found.
func GetUIStrings(locale string) UIStrings {
	if strings, ok := uiStrings[locale]; ok {
		return strings
	}
	return uiStrings["en"]
}

// ============================================================================
// M1/M2 absorbed wizard strings (plan.md M7: folded into the translations.go
// scheme so every wizard UI string table lives in one file).
// ============================================================================

// helpActionLabels maps the English help action strings of huh's default key
// map to each locale (design.md §7). The en column is the identity so every
// locale carries the same action set. Key names (enter, ↑, …) are never
// translated; only the action after them is.
var helpActionLabels = map[string]map[string]string{
	"en": {"next": "next", "submit": "submit", "back": "back", "select": "select", "up": "up", "down": "down",
		"filter": "filter", "set filter": "set filter", "clear filter": "clear filter", "toggle": "toggle", "complete": "complete"},
	"ko": {"next": "다음", "submit": "제출", "back": "이전", "select": "선택", "up": "위", "down": "아래",
		"filter": "검색", "set filter": "검색 적용", "clear filter": "검색 해제", "toggle": "전환", "complete": "자동 완성"},
	"ja": {"next": "次へ", "submit": "送信", "back": "戻る", "select": "選択", "up": "上", "down": "下",
		"filter": "絞り込み", "set filter": "絞り込み確定", "clear filter": "絞り込み解除", "toggle": "切替", "complete": "補完"},
	"zh": {"next": "下一步", "submit": "提交", "back": "返回", "select": "选择", "up": "上", "down": "下",
		"filter": "筛选", "set filter": "应用筛选", "clear filter": "清除筛选", "toggle": "切换", "complete": "补全"},
}

// downgradeConfirmText is the localized title format (current, target) and
// description of the `moai update --version` downgrade confirmation.
type downgradeConfirmText struct {
	TitleFormat string
	Description string
}

// downgradeConfirmTexts holds the downgrade confirmation strings for the four
// locales (design.md §6).
var downgradeConfirmTexts = map[string]downgradeConfirmText{
	"en": {TitleFormat: "Downgrade %s → %s?", Description: "The requested tag is older than the running version."},
	"ko": {TitleFormat: "다운그레이드할까요? %s → %s", Description: "요청한 태그가 지금 실행 중인 버전보다 오래되었습니다."},
	"ja": {TitleFormat: "ダウングレードしますか？ %s → %s", Description: "指定したタグは実行中のバージョンより古いバージョンです。"},
	"zh": {TitleFormat: "要降级吗？%s → %s", Description: "请求的标签比当前运行的版本旧。"},
}

// profileQuestionTexts maps locale -> profile question id -> translation
// for the profile wizard (`moai profile setup`); moved here from
// profile_translations.go in the M7 strings fold (plan.md M7).
// profileQuestionTexts maps locale -> profile question id -> translation for
// the profile wizard (`moai profile setup`). The strings are the form strings
// the v1 profile wizard carried in the cli package (profileSetupText), moved
// here unchanged so the absorbed v2 form renders the same text in every
// locale.
//
// The table is keyed separately from the init wizard's `translations` because
// four profile ids (conversation_language, user_name, model_policy,
// development_mode) also exist in the init set with different wording; one
// shared id-keyed table cannot hold both. Option labels are NOT carried here:
// the caller resolves them (schema label bridge) and passes them in through
// ProfileOptions.
var profileQuestionTexts = map[string]map[string]QuestionTranslation{
	"en": {
		"conversation_language": {Title: "Select your language", Description: "Language for this wizard and Claude's responses."},
		"user_name":             {Title: "User name", Description: "Your display name for configuration files. Press Enter to skip."},
		"git_commit_lang":       {Title: "Git commit message language", Description: "Language for commit messages."},
		"code_comment_lang":     {Title: "Code comment language", Description: "Language for code comments."},
		"doc_lang":              {Title: "Documentation language", Description: "Language for documentation files."},
		"model":                 {Title: "Default model override", Description: "Override the model when launching with this profile."},
		"model_policy":          {Title: "Agent model policy", Description: "Controls token consumption by assigning optimal models to each agent."},
		"effort_level":          {Title: "Session effort level", Description: "Reasoning depth for the Claude session launched with this profile. xhigh/max need a model that supports them (Opus 5, Sonnet 5, Opus 4.7+). Per-agent effort comes from the agent model policy instead."},
		"permission_mode":       {Title: "Permission mode", Description: "Controls how Claude asks for permission before taking actions."},
		"development_mode":      {Title: "Development mode", Description: "Project methodology written to quality.yaml. Empty keeps the project default."},
	},
	"ko": {
		"conversation_language": {Title: "언어를 선택하세요", Description: "이 설정 마법사와 Claude 응답에 사용할 언어입니다."},
		"user_name":             {Title: "사용자 이름", Description: "설정 파일에 표시될 이름입니다. Enter를 눌러 건너뛰세요."},
		"git_commit_lang":       {Title: "Git 커밋 메시지 언어", Description: "커밋 메시지에 사용할 언어입니다."},
		"code_comment_lang":     {Title: "코드 주석 언어", Description: "코드 주석에 사용할 언어입니다."},
		"doc_lang":              {Title: "문서 언어", Description: "문서 파일에 사용할 언어입니다."},
		"model":                 {Title: "기본 모델 오버라이드", Description: "이 프로필로 실행할 때 모델을 오버라이드합니다."},
		"model_policy":          {Title: "에이전트 모델 정책", Description: "각 에이전트에 최적 모델을 할당하여 토큰 소비를 제어합니다."},
		"effort_level":          {Title: "세션 추론 강도", Description: "이 프로필로 실행하는 Claude 세션의 추론 깊이입니다. xhigh/max는 이를 지원하는 모델(Opus 5, Sonnet 5, Opus 4.7 이상)이 필요합니다. 에이전트별 추론 강도는 에이전트 모델 정책에서 정해집니다."},
		"permission_mode":       {Title: "권한 모드", Description: "Claude가 작업 수행 전 권한을 요청하는 방식을 제어합니다."},
		"development_mode":      {Title: "개발 방법론", Description: "quality.yaml에 기록되는 프로젝트 개발 방법론. 비워두면 프로젝트 기본값을 유지합니다."},
	},
	"ja": {
		"conversation_language": {Title: "言語を選択してください", Description: "このウィザードとClaudeの応答に使用する言語です。"},
		"user_name":             {Title: "ユーザー名", Description: "設定ファイルに表示される名前です。Enterでスキップできます。"},
		"git_commit_lang":       {Title: "Gitコミットメッセージ言語", Description: "コミットメッセージに使用する言語です。"},
		"code_comment_lang":     {Title: "コードコメント言語", Description: "コードコメントに使用する言語です。"},
		"doc_lang":              {Title: "ドキュメント言語", Description: "ドキュメントファイルに使用する言語です。"},
		"model":                 {Title: "デフォルトモデルオーバーライド", Description: "このプロファイルで起動する際のモデルをオーバーライドします。"},
		"model_policy":          {Title: "エージェントモデルポリシー", Description: "各エージェントに最適なモデルを割り当て、トークン消費を制御します。"},
		"effort_level":          {Title: "セッション推論レベル", Description: "このプロファイルで起動する Claude セッションの推論深度です。xhigh/max は対応モデル（Opus 5、Sonnet 5、Opus 4.7 以降）が必要です。エージェントごとの推論強度はエージェントモデルポリシーで決まります。"},
		"permission_mode":       {Title: "権限モード", Description: "Claudeがアクション実行前に権限を要求する方法を制御します。"},
		"development_mode":      {Title: "開発方法論", Description: "quality.yaml に記録されるプロジェクトの開発方法論。空欄の場合はプロジェクトのデフォルトを維持します。"},
	},
	"zh": {
		"conversation_language": {Title: "请选择语言", Description: "用于此向导和Claude响应的语言。"},
		"user_name":             {Title: "用户名", Description: "配置文件中显示的名称。按Enter跳过。"},
		"git_commit_lang":       {Title: "Git提交消息语言", Description: "提交消息使用的语言。"},
		"code_comment_lang":     {Title: "代码注释语言", Description: "代码注释使用的语言。"},
		"doc_lang":              {Title: "文档语言", Description: "文档文件使用的语言。"},
		"model":                 {Title: "默认模型覆盖", Description: "使用此配置文件启动时覆盖模型。"},
		"model_policy":          {Title: "代理模型策略", Description: "通过为每个代理分配最优模型来控制token消耗。"},
		"effort_level":          {Title: "会话推理强度", Description: "使用此配置文件启动的 Claude 会话的推理深度。xhigh/max 需要支持它们的模型（Opus 5、Sonnet 5、Opus 4.7 及以上）。各代理的推理强度由代理模型策略决定。"},
		"permission_mode":       {Title: "权限模式", Description: "控制Claude在执行操作前如何请求权限。"},
		"development_mode":      {Title: "开发方法论", Description: "写入 quality.yaml 的项目开发方法论。留空则保留项目默认值。"},
	},
}
