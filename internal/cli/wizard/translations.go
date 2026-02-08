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
		"locale": {
			Title:       "대화 언어 선택",
			Description: "Claude가 대화할 때 사용할 언어를 선택합니다.",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "한국어"},
				{Label: "English", Desc: "영어"},
				{Label: "Japanese (日本語)", Desc: "일본어"},
				{Label: "Chinese (中文)", Desc: "중국어"},
			},
		},
		"user_name": {
			Title:       "이름 입력",
			Description: "설정 파일에 사용됩니다. Enter를 눌러 건너뛸 수 있습니다.",
		},
		"project_name": {
			Title:       "프로젝트 이름 입력",
			Description: "프로젝트의 이름입니다.",
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
		"github_username": {
			Title:       "GitHub 사용자명 입력",
			Description: "Git 자동화 기능에 필요합니다.",
		},
		"git_commit_lang": {
			Title:       "Git 커밋 메시지 언어 선택",
			Description: "커밋 메시지 작성에 사용할 언어입니다.",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "한국어로 커밋"},
				{Label: "English", Desc: "영어로 커밋"},
				{Label: "Japanese (日本語)", Desc: "일본어로 커밋"},
				{Label: "Chinese (中文)", Desc: "중국어로 커밋"},
			},
		},
		"code_comment_lang": {
			Title:       "코드 주석 언어 선택",
			Description: "코드 주석에 사용할 언어입니다.",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "한국어로 주석"},
				{Label: "English", Desc: "영어로 주석"},
				{Label: "Japanese (日本語)", Desc: "일본어로 주석"},
				{Label: "Chinese (中文)", Desc: "중국어로 주석"},
			},
		},
		"doc_lang": {
			Title:       "문서 언어 선택",
			Description: "문서 파일에 사용할 언어입니다.",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "한국어로 문서"},
				{Label: "English", Desc: "영어로 문서"},
				{Label: "Japanese (日本語)", Desc: "일본어로 문서"},
				{Label: "Chinese (中文)", Desc: "중국어로 문서"},
			},
		},
		"development_mode": {
			Title:       "개발 방법론 선택",
			Description: "코드 변경 및 테스트 처리 방식을 결정합니다.",
			Options: []OptionTranslation{
				{Label: "Hybrid (TDD + DDD) (권장)", Desc: "신규 기능은 TDD, 기존 코드는 DDD"},
				{Label: "DDD (도메인 주도 개발)", Desc: "레거시 리팩토링을 위한 ANALYZE-PRESERVE-IMPROVE 사이클"},
			},
		},
		"agent_teams_mode": {
			Title:       "Agent Teams 실행 모드 선택",
			Description: "MoAI가 Agent Teams(병렬) 또는 sub-agents(순차)를 사용하도록 설정합니다.",
			Options: []OptionTranslation{
				{Label: "Auto (권장)", Desc: "작업 복잡도 기반 지능형 선택"},
				{Label: "Sub-agent (클래식)", Desc: "기존 단일 에이전트 모드"},
				{Label: "Team (실험적)", Desc: "병렬 Agent Teams (실험적 기능 필요)"},
			},
		},
		"max_teammates": {
			Title:       "최대 팀원 수 선택",
			Description: "팀의 최대 팀원 수 (2-10 권장).",
			Options: []OptionTranslation{
				{Label: "2", Desc: "병렬 작업을 위한 최소값"},
				{Label: "3", Desc: "소규 팀"},
				{Label: "4", Desc: "중간 팀"},
				{Label: "5", Desc: "중대규 팀"},
				{Label: "6", Desc: "대규 팀"},
				{Label: "7", Desc: "대규 팀"},
				{Label: "8", Desc: "초대규 팀"},
				{Label: "9", Desc: "초대규 팀"},
				{Label: "10", Desc: "최대 팀 (기본값)"},
			},
		},
		"default_model": {
			Title:       "팀원 기본 모델 선택",
			Description: "Agent Team원의 기본 Claude 모델.",
			Options: []OptionTranslation{
				{Label: "Haiku (빠름/저렴)", Desc: "가장 빠르고 저렴"},
				{Label: "Sonnet (균형)", Desc: "성능과 비용의 균형 (기본값)"},
				{Label: "Opus (고성능)", Desc: "가장 강력하지만 비용 높음"},
			},
		},
		"github_token": {
			Title:       "GitHub 개인 액세스 토큰 입력 (선택)",
			Description: "PR 생성 및 푸시에 필요합니다. 비워두어 건너거나 gh CLI를 사용하세요.",
		},
	},
	"ja": {
		"locale": {
			Title:       "会話言語を選択",
			Description: "Claudeとの会話で使用する言語を選択します。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韓国語"},
				{Label: "English", Desc: "英語"},
				{Label: "Japanese (日本語)", Desc: "日本語"},
				{Label: "Chinese (中文)", Desc: "中国語"},
			},
		},
		"user_name": {
			Title:       "名前を入力",
			Description: "設定ファイルで使用されます。Enterでスキップできます。",
		},
		"project_name": {
			Title:       "プロジェクト名を入力",
			Description: "プロジェクトの名前です。",
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
		"github_username": {
			Title:       "GitHubユーザー名を入力",
			Description: "Git自動化機能に必要です。",
		},
		"git_commit_lang": {
			Title:       "Gitコミットメッセージ言語を選択",
			Description: "コミットメッセージで使用する言語です。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韓国語でコミット"},
				{Label: "English", Desc: "英語でコミット"},
				{Label: "Japanese (日本語)", Desc: "日本語でコミット"},
				{Label: "Chinese (中文)", Desc: "中国語でコミット"},
			},
		},
		"code_comment_lang": {
			Title:       "コードコメント言語を選択",
			Description: "コードコメントで使用する言語です。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韓国語でコメント"},
				{Label: "English", Desc: "英語でコメント"},
				{Label: "Japanese (日本語)", Desc: "日本語でコメント"},
				{Label: "Chinese (中文)", Desc: "中国語でコメント"},
			},
		},
		"doc_lang": {
			Title:       "ドキュメント言語を選択",
			Description: "ドキュメントファイルで使用する言語です。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韓国語でドキュメント"},
				{Label: "English", Desc: "英語でドキュメント"},
				{Label: "Japanese (日本語)", Desc: "日本語でドキュメント"},
				{Label: "Chinese (中文)", Desc: "中国語でドキュメント"},
			},
		},
		"development_mode": {
			Title:       "開発方法論を選択",
			Description: "コード変更とテストの処理方法を決定します。",
			Options: []OptionTranslation{
				{Label: "Hybrid (TDD + DDD) (推奨)", Desc: "新機能はTDD、既存コードはDDD"},
				{Label: "DDD (ドメイン駆動開発)", Desc: "レガシーリファクタリングのためのANALYZE-PRESERVE-IMPROVEサイクル"},
			},
		},
		"agent_teams_mode": {
			Title:       "Agent Teams実行モードを選択",
			Description: "MoAIがAgent Teams（並列）かsub-agents（順次）を使用するかを制御します。",
			Options: []OptionTranslation{
				{Label: "Auto (推奨)", Desc: "タスク複雑さに基づくインテリジェント選択"},
				{Label: "Sub-agent (クラシック)", Desc: "従来の単一エージェントモード"},
				{Label: "Team (実験的)", Desc: "並列Agent Teams（実験的フラグが必要）"},
			},
		},
		"max_teammates": {
			Title:       "最大チームメイト数を選択",
			Description: "チームの最大メイト数（2-10推奨）。",
			Options: []OptionTranslation{
				{Label: "2", Desc: "並列作業の最小値"},
				{Label: "3", Desc: "小規模チーム"},
				{Label: "4", Desc: "中規模チーム"},
				{Label: "5", Desc: "中大規模チーム"},
				{Label: "6", Desc: "大規模チーム"},
				{Label: "7", Desc: "大規模チーム"},
				{Label: "8", Desc: "超大規模チーム"},
				{Label: "9", Desc: "超大規模チーム"},
				{Label: "10", Desc: "最大チーム（デフォルト）"},
			},
		},
		"default_model": {
			Title:       "チームメイトのデフォルトモデルを選択",
			Description: "Agent TeamメイトのデフォルトClaudeモデル。",
			Options: []OptionTranslation{
				{Label: "Haiku (高速/低コスト)", Desc: "最も高速で低コスト"},
				{Label: "Sonnet (バランス)", Desc: "パフォーマンスとコストのバランス（デフォルト）"},
				{Label: "Opus (高機能)", Desc: "最も高機能だが高コスト"},
			},
		},
		"github_token": {
			Title:       "GitHubパーソナルアクセストークンを入力（省略可）",
			Description: "PR作成とプッシュに必要です。空欄のままスキップまたはgh CLIを使用してください。",
		},
	},
	"zh": {
		"locale": {
			Title:       "选择对话语言",
			Description: "选择Claude与您交流时使用的语言。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韩语"},
				{Label: "English", Desc: "英语"},
				{Label: "Japanese (日本語)", Desc: "日语"},
				{Label: "Chinese (中文)", Desc: "中文"},
			},
		},
		"user_name": {
			Title:       "输入姓名",
			Description: "将用于配置文件。按Enter跳过。",
		},
		"project_name": {
			Title:       "输入项目名称",
			Description: "项目的名称。",
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
		"github_username": {
			Title:       "输入GitHub用户名",
			Description: "Git自动化功能所需。",
		},
		"git_commit_lang": {
			Title:       "选择Git提交消息语言",
			Description: "编写提交消息使用的语言。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韩语提交"},
				{Label: "English", Desc: "英语提交"},
				{Label: "Japanese (日本語)", Desc: "日语提交"},
				{Label: "Chinese (中文)", Desc: "中文提交"},
			},
		},
		"code_comment_lang": {
			Title:       "选择代码注释语言",
			Description: "代码注释使用的语言。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韩语注释"},
				{Label: "English", Desc: "英语注释"},
				{Label: "Japanese (日本語)", Desc: "日语注释"},
				{Label: "Chinese (中文)", Desc: "中文注释"},
			},
		},
		"doc_lang": {
			Title:       "选择文档语言",
			Description: "文档文件使用的语言。",
			Options: []OptionTranslation{
				{Label: "Korean (한국어)", Desc: "韩语文档"},
				{Label: "English", Desc: "英语文档"},
				{Label: "Japanese (日本語)", Desc: "日语文档"},
				{Label: "Chinese (中文)", Desc: "中文文档"},
			},
		},
		"development_mode": {
			Title:       "选择开发方法论",
			Description: "决定代码更改和测试的处理方式。",
			Options: []OptionTranslation{
				{Label: "Hybrid (TDD + DDD) (推荐)", Desc: "新功能用TDD，现有代码用DDD"},
				{Label: "DDD (领域驱动开发)", Desc: "用于遗留代码重构的ANALYZE-PRESERVE-IMPROVE循环"},
			},
		},
		"agent_teams_mode": {
			Title:       "选择Agent Teams执行模式",
			Description: "控制MoAI使用Agent Teams（并行）还是sub-agents（顺序）。",
			Options: []OptionTranslation{
				{Label: "Auto (推荐)", Desc: "基于任务复杂度的智能选择"},
				{Label: "Sub-agent (经典)", Desc: "传统单代理模式"},
				{Label: "Team (实验性)", Desc: "并行Agent Teams（需要实验性标志）"},
			},
		},
		"max_teammates": {
			Title:       "选择最大团队成员数",
			Description: "团队中最大成员数（建议2-10）。",
			Options: []OptionTranslation{
				{Label: "2", Desc: "并行工作的最小值"},
				{Label: "3", Desc: "小型团队"},
				{Label: "4", Desc: "中型团队"},
				{Label: "5", Desc: "中大型团队"},
				{Label: "6", Desc: "大型团队"},
				{Label: "7", Desc: "大型团队"},
				{Label: "8", Desc: "超大型团队"},
				{Label: "9", Desc: "超大型团队"},
				{Label: "10", Desc: "最大团队（默认）"},
			},
		},
		"default_model": {
			Title:       "选择团队成员的默认模型",
			Description: "Agent Teammates的默认Claude模型。",
			Options: []OptionTranslation{
				{Label: "Haiku (快速/低成本)", Desc: "最快且成本最低"},
				{Label: "Sonnet (平衡)", Desc: "性能与成本的平衡（默认）"},
				{Label: "Opus (强大)", Desc: "最强大但成本较高"},
			},
		},
		"github_token": {
			Title:       "输入GitHub个人访问令牌（可选）",
			Description: "PR创建和推送所需。留空以跳过或使用gh CLI。",
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
