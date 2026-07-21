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
	HelpSelect    string
	HelpInput     string
	ErrorRequired string
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
		"advanced_bridge": {
			Title:       "고급 설정을 구성할까요? (기본값: 아니오)",
			Description: "프로젝트 모드, 하네스 프로파일, LSP, 품질 게이트, 디자인 옵션을 이번 실행에서 표시합니다.",
		},
		"project_name": {
			Title:       "프로젝트 이름 입력",
			Description: "프로젝트의 이름입니다.",
		},
		"development_mode": {
			Title:       "개발 방법론 선택",
			Description: "구현 시 사용할 개발 워크플로우 사이클을 설정합니다.",
			Options: []OptionTranslation{
				{Label: "TDD (권장)", Desc: "테스트 주도 개발: RED-GREEN-REFACTOR"},
				{Label: "DDD", Desc: "도메인 주도 개발: ANALYZE-PRESERVE-IMPROVE"},
			},
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
		"advanced_bridge": {
			Title:       "詳細設定を構成しますか？（デフォルト: いいえ）",
			Description: "プロジェクトモード、ハーネスプロファイル、LSP、品質ゲート、デザインオプションをこの実行で表示します。",
		},
		"project_name": {
			Title:       "プロジェクト名を入力",
			Description: "プロジェクトの名前です。",
		},
		"development_mode": {
			Title:       "開発方法論を選択",
			Description: "実装時に使用する開発ワークフローサイクルを制御します。",
			Options: []OptionTranslation{
				{Label: "TDD (推奨)", Desc: "テスト駆動開発: RED-GREEN-REFACTOR"},
				{Label: "DDD", Desc: "ドメイン駆動開発: ANALYZE-PRESERVE-IMPROVE"},
			},
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
		"advanced_bridge": {
			Title:       "配置高级设置？（默认：否）",
			Description: "在本次运行中显示项目模式、评估套件配置、LSP、质量门禁和设计选项。",
		},
		"project_name": {
			Title:       "输入项目名称",
			Description: "项目的名称。",
		},
		"development_mode": {
			Title:       "选择开发方法论",
			Description: "控制实施期间使用的开发工作流程周期。",
			Options: []OptionTranslation{
				{Label: "TDD (推荐)", Desc: "测试驱动开发: RED-GREEN-REFACTOR"},
				{Label: "DDD", Desc: "领域驱动开发: ANALYZE-PRESERVE-IMPROVE"},
			},
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
	},
}

// uiStrings maps language code to UI strings.
var uiStrings = map[string]UIStrings{
	"en": {
		HelpSelect:    "Use arrow keys to navigate, Enter to select, Esc to cancel",
		HelpInput:     "Type your answer, Enter to confirm, Esc to cancel",
		ErrorRequired: "This field is required",
	},
	"ko": {
		HelpSelect:    "방향키로 이동, Enter로 선택, Esc로 취소",
		HelpInput:     "답변 입력 후 Enter로 확인, Esc로 취소",
		ErrorRequired: "필수 입력 항목입니다",
	},
	"ja": {
		HelpSelect:    "矢印キーで移動、Enterで選択、Escでキャンセル",
		HelpInput:     "入力してEnterで確定、Escでキャンセル",
		ErrorRequired: "この項目は必須です",
	},
	"zh": {
		HelpSelect:    "使用方向键导航，Enter选择，Esc取消",
		HelpInput:     "输入答案，Enter确认，Esc取消",
		ErrorRequired: "此字段为必填项",
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
