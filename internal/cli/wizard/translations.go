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
		"gitlab_username": {
			Title:       "GitLab 사용자명 입력",
			Description: "GitLab Git 자동화 기능에 필요합니다.",
		},
		"gitlab_token": {
			Title:       "GitLab 개인 액세스 토큰 입력 (선택사항)",
			Description: "MR 생성 및 푸시에 필요합니다. 비워두거나 glab CLI를 사용할 수 있습니다.",
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
		"model_policy": {
			Title:       "에이전트 모델 정책 선택",
			Description: "Claude Code 요금제에 따라 각 에이전트에 최적 모델을 할당합니다.",
			Options: []OptionTranslation{
				{Label: "High (Max $200/mo)", Desc: "모든 에이전트가 opus 사용"},
				{Label: "Medium (Max $100/mo)", Desc: "핵심 에이전트 opus, 나머지 sonnet/haiku"},
				{Label: "Low (Plus $20/mo)", Desc: "opus 없음, sonnet + haiku만 사용"},
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
				{Label: "10", Desc: "최대 팀 (기본값)"},
				{Label: "9", Desc: "초대규모 팀"},
				{Label: "8", Desc: "초대규모 팀"},
				{Label: "7", Desc: "대규모 팀"},
				{Label: "6", Desc: "대규모 팀"},
				{Label: "5", Desc: "중대규모 팀"},
				{Label: "4", Desc: "중간 팀"},
				{Label: "3", Desc: "소규모 팀"},
				{Label: "2", Desc: "병렬 작업을 위한 최소값"},
			},
		},
		"default_model": {
			Title:       "팀원 기본 모델 선택",
			Description: "Agent Team원의 기본 Claude 모델.",
			Options: []OptionTranslation{
				{Label: "Sonnet (균형)", Desc: "성능과 비용의 균형 (기본값)"},
				{Label: "Haiku (빠름/저렴)", Desc: "가장 빠르고 저렴"},
				{Label: "Opus (고성능)", Desc: "가장 강력하지만 비용 높음"},
			},
		},
		"github_token": {
			Title:       "GitHub 개인 액세스 토큰 입력 (선택)",
			Description: "PR 생성 및 푸시에 필요합니다. 비워두어 건너거나 gh CLI를 사용하세요.",
		},
		"teammate_display": {
			Title:       "팀원 표시 모드 선택",
			Description: "Agent 팀원 표시 방법을 설정합니다. 분할 화면은 tmux가 필요합니다.",
			Options: []OptionTranslation{
				{Label: "Auto (권장)", Desc: "tmux 사용 가능 시 tmux, 없으면 in-process (기본값)"},
				{Label: "In-Process", Desc: "같은 터미널에서 실행 (어디서나 동작)"},
				{Label: "Tmux", Desc: "tmux 분할 화면 (tmux/iTerm2 필요)"},
			},
		},
		"statusline_preset": {
			Title:       "상태줄 표시 프리셋 선택",
			Description: "Claude Code 상태줄에 표시할 세그먼트를 설정합니다.",
			Options: []OptionTranslation{
				{Label: "Full", Desc: "8개 세그먼트 전부 표시"},
				{Label: "Compact", Desc: "모델, 컨텍스트, Git 상태, Git 브랜치"},
				{Label: "Minimal", Desc: "모델과 컨텍스트만"},
				{Label: "Custom", Desc: "개별 세그먼트 선택"},
			},
		},
		"statusline_seg_model": {
			Title:       "상태줄: 모델 이름 표시",
			Description: "상태줄에 현재 Claude 모델 이름을 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "모델 세그먼트 표시"},
				{Label: "비활성화", Desc: "모델 세그먼트 숨김"},
			},
		},
		"statusline_seg_context": {
			Title:       "상태줄: 컨텍스트 사용량 표시",
			Description: "상태줄에 컨텍스트 윈도우 사용률을 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "컨텍스트 세그먼트 표시"},
				{Label: "비활성화", Desc: "컨텍스트 세그먼트 숨김"},
			},
		},
		"statusline_seg_output_style": {
			Title:       "상태줄: 출력 스타일 표시",
			Description: "상태줄에 활성 출력 스타일 이름을 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "출력 스타일 세그먼트 표시"},
				{Label: "비활성화", Desc: "출력 스타일 세그먼트 숨김"},
			},
		},
		"statusline_seg_directory": {
			Title:       "상태줄: 디렉토리 이름 표시",
			Description: "상태줄에 현재 작업 디렉토리 이름을 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "디렉토리 세그먼트 표시"},
				{Label: "비활성화", Desc: "디렉토리 세그먼트 숨김"},
			},
		},
		"statusline_seg_git_status": {
			Title:       "상태줄: Git 상태 표시",
			Description: "상태줄에 Git 상태 (스테이지, 수정, 비추적 수)를 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "Git 상태 세그먼트 표시"},
				{Label: "비활성화", Desc: "Git 상태 세그먼트 숨김"},
			},
		},
		"statusline_seg_claude_version": {
			Title:       "상태줄: Claude 버전 표시",
			Description: "상태줄에 Claude Code 버전을 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "Claude 버전 세그먼트 표시"},
				{Label: "비활성화", Desc: "Claude 버전 세그먼트 숨김"},
			},
		},
		"statusline_seg_moai_version": {
			Title:       "상태줄: MoAI 버전 표시",
			Description: "상태줄에 MoAI-ADK 버전을 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "MoAI 버전 세그먼트 표시"},
				{Label: "비활성화", Desc: "MoAI 버전 세그먼트 숨김"},
			},
		},
		"statusline_seg_git_branch": {
			Title:       "상태줄: Git 브랜치 표시",
			Description: "상태줄에 현재 Git 브랜치 이름을 표시합니다.",
			Options: []OptionTranslation{
				{Label: "활성화", Desc: "Git 브랜치 세그먼트 표시"},
				{Label: "비활성화", Desc: "Git 브랜치 세그먼트 숨김"},
			},
		},
	},
	"ja": {
		"model_policy": {
			Title:       "エージェントモデルポリシーを選択",
			Description: "Claude Codeプランに基づいて各エージェントに最適なモデルを割り当てます。",
			Options: []OptionTranslation{
				{Label: "High (Max $200/mo)", Desc: "全エージェントがopusを使用"},
				{Label: "Medium (Max $100/mo)", Desc: "重要エージェントにopus、他はsonnet/haiku"},
				{Label: "Low (Plus $20/mo)", Desc: "opus無し、sonnet + haikuのみ"},
			},
		},
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
		"gitlab_username": {
			Title:       "GitLabユーザー名を入力",
			Description: "GitLab Git自動化機能に必要です。",
		},
		"gitlab_token": {
			Title:       "GitLabパーソナルアクセストークンを入力（省略可）",
			Description: "MR作成とプッシュに必要です。空欄のままスキップまたはglab CLIを使用してください。",
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
				{Label: "10", Desc: "最大チーム（デフォルト）"},
				{Label: "9", Desc: "超大規模チーム"},
				{Label: "8", Desc: "超大規模チーム"},
				{Label: "7", Desc: "大規模チーム"},
				{Label: "6", Desc: "大規模チーム"},
				{Label: "5", Desc: "中大規模チーム"},
				{Label: "4", Desc: "中規模チーム"},
				{Label: "3", Desc: "小規模チーム"},
				{Label: "2", Desc: "並列作業の最小値"},
			},
		},
		"default_model": {
			Title:       "チームメイトのデフォルトモデルを選択",
			Description: "Agent TeamメイトのデフォルトClaudeモデル。",
			Options: []OptionTranslation{
				{Label: "Sonnet (バランス)", Desc: "パフォーマンスとコストのバランス（デフォルト）"},
				{Label: "Haiku (高速/低コスト)", Desc: "最も高速で低コスト"},
				{Label: "Opus (高機能)", Desc: "最も高機能だが高コスト"},
			},
		},
		"github_token": {
			Title:       "GitHubパーソナルアクセストークンを入力（省略可）",
			Description: "PR作成とプッシュに必要です。空欄のままスキップまたはgh CLIを使用してください。",
		},
		"teammate_display": {
			Title:       "チームメイト表示モードを選択",
			Description: "Agent Teammatesの表示方法を制御します。分割ペインにはtmuxが必要です。",
			Options: []OptionTranslation{
				{Label: "Auto (推奨)", Desc: "tmux利用可能時はtmux、それ以外はin-process（デフォルト）"},
				{Label: "In-Process", Desc: "同一ターミナルで実行（どこでも動作）"},
				{Label: "Tmux", Desc: "tmuxで分割ペイン（tmux/iTerm2が必要）"},
			},
		},
		"statusline_preset": {
			Title:       "ステータスライン表示プリセットを選択",
			Description: "Claude Codeステータスラインに表示するセグメントを設定します。",
			Options: []OptionTranslation{
				{Label: "Full", Desc: "全8セグメントを表示"},
				{Label: "Compact", Desc: "モデル、コンテキスト、Git状態、Gitブランチ"},
				{Label: "Minimal", Desc: "モデルとコンテキストのみ"},
				{Label: "Custom", Desc: "個別のセグメントを選択"},
			},
		},
		"statusline_seg_model": {
			Title:       "ステータスライン: モデル名を表示",
			Description: "ステータスラインに現在のClaudeモデル名を表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "モデルセグメントを表示"},
				{Label: "無効", Desc: "モデルセグメントを非表示"},
			},
		},
		"statusline_seg_context": {
			Title:       "ステータスライン: コンテキスト使用量を表示",
			Description: "ステータスラインにコンテキストウィンドウ使用率を表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "コンテキストセグメントを表示"},
				{Label: "無効", Desc: "コンテキストセグメントを非表示"},
			},
		},
		"statusline_seg_output_style": {
			Title:       "ステータスライン: 出力スタイルを表示",
			Description: "ステータスラインにアクティブな出力スタイル名を表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "出力スタイルセグメントを表示"},
				{Label: "無効", Desc: "出力スタイルセグメントを非表示"},
			},
		},
		"statusline_seg_directory": {
			Title:       "ステータスライン: ディレクトリ名を表示",
			Description: "ステータスラインに現在の作業ディレクトリ名を表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "ディレクトリセグメントを表示"},
				{Label: "無効", Desc: "ディレクトリセグメントを非表示"},
			},
		},
		"statusline_seg_git_status": {
			Title:       "ステータスライン: Git状態を表示",
			Description: "ステータスラインにGit状態（ステージ、変更、未追跡の数）を表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "Git状態セグメントを表示"},
				{Label: "無効", Desc: "Git状態セグメントを非表示"},
			},
		},
		"statusline_seg_claude_version": {
			Title:       "ステータスライン: Claudeバージョンを表示",
			Description: "ステータスラインにClaude Codeバージョンを表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "Claudeバージョンセグメントを表示"},
				{Label: "無効", Desc: "Claudeバージョンセグメントを非表示"},
			},
		},
		"statusline_seg_moai_version": {
			Title:       "ステータスライン: MoAIバージョンを表示",
			Description: "ステータスラインにMoAI-ADKバージョンを表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "MoAIバージョンセグメントを表示"},
				{Label: "無効", Desc: "MoAIバージョンセグメントを非表示"},
			},
		},
		"statusline_seg_git_branch": {
			Title:       "ステータスライン: Gitブランチを表示",
			Description: "ステータスラインに現在のGitブランチ名を表示します。",
			Options: []OptionTranslation{
				{Label: "有効", Desc: "Gitブランチセグメントを表示"},
				{Label: "無効", Desc: "Gitブランチセグメントを非表示"},
			},
		},
	},
	"zh": {
		"model_policy": {
			Title:       "选择代理模型策略",
			Description: "根据Claude Code计划为每个代理分配最佳模型。",
			Options: []OptionTranslation{
				{Label: "High (Max $200/mo)", Desc: "所有代理使用opus"},
				{Label: "Medium (Max $100/mo)", Desc: "关键代理使用opus，其他sonnet/haiku"},
				{Label: "Low (Plus $20/mo)", Desc: "无opus，仅sonnet + haiku"},
			},
		},
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
		"gitlab_username": {
			Title:       "输入GitLab用户名",
			Description: "GitLab Git自动化功能所需。",
		},
		"gitlab_token": {
			Title:       "输入GitLab个人访问令牌（可选）",
			Description: "MR创建和推送所需。留空以跳过或使用glab CLI。",
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
				{Label: "10", Desc: "最大团队（默认）"},
				{Label: "9", Desc: "超大型团队"},
				{Label: "8", Desc: "超大型团队"},
				{Label: "7", Desc: "大型团队"},
				{Label: "6", Desc: "大型团队"},
				{Label: "5", Desc: "中大型团队"},
				{Label: "4", Desc: "中型团队"},
				{Label: "3", Desc: "小型团队"},
				{Label: "2", Desc: "并行工作的最小值"},
			},
		},
		"default_model": {
			Title:       "选择团队成员的默认模型",
			Description: "Agent Teammates的默认Claude模型。",
			Options: []OptionTranslation{
				{Label: "Sonnet (平衡)", Desc: "性能与成本的平衡（默认）"},
				{Label: "Haiku (快速/低成本)", Desc: "最快且成本最低"},
				{Label: "Opus (强大)", Desc: "最强大但成本较高"},
			},
		},
		"github_token": {
			Title:       "输入GitHub个人访问令牌（可选）",
			Description: "PR创建和推送所需。留空以跳过或使用gh CLI。",
		},
		"teammate_display": {
			Title:       "选择队友显示模式",
			Description: "控制Agent Teammates的显示方式。分割窗格需要tmux。",
			Options: []OptionTranslation{
				{Label: "Auto (推荐)", Desc: "有tmux时使用tmux，否则使用in-process（默认）"},
				{Label: "In-Process", Desc: "在同一终端中运行（任何地方都可用）"},
				{Label: "Tmux", Desc: "在tmux中使用分割窗格（需要tmux/iTerm2）"},
			},
		},
		"statusline_preset": {
			Title:       "选择状态栏显示预设",
			Description: "设置Claude Code状态栏中显示的段。",
			Options: []OptionTranslation{
				{Label: "Full", Desc: "显示全部8个段"},
				{Label: "Compact", Desc: "模型、上下文、Git状态、Git分支"},
				{Label: "Minimal", Desc: "仅模型和上下文"},
				{Label: "Custom", Desc: "选择单个段"},
			},
		},
		"statusline_seg_model": {
			Title:       "状态栏: 显示模型名称",
			Description: "在状态栏中显示当前Claude模型名称。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示模型段"},
				{Label: "禁用", Desc: "隐藏模型段"},
			},
		},
		"statusline_seg_context": {
			Title:       "状态栏: 显示上下文使用量",
			Description: "在状态栏中显示上下文窗口使用百分比。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示上下文段"},
				{Label: "禁用", Desc: "隐藏上下文段"},
			},
		},
		"statusline_seg_output_style": {
			Title:       "状态栏: 显示输出样式",
			Description: "在状态栏中显示活动输出样式名称。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示输出样式段"},
				{Label: "禁用", Desc: "隐藏输出样式段"},
			},
		},
		"statusline_seg_directory": {
			Title:       "状态栏: 显示目录名称",
			Description: "在状态栏中显示当前工作目录名称。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示目录段"},
				{Label: "禁用", Desc: "隐藏目录段"},
			},
		},
		"statusline_seg_git_status": {
			Title:       "状态栏: 显示Git状态",
			Description: "在状态栏中显示Git状态（暂存、修改、未跟踪数量）。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示Git状态段"},
				{Label: "禁用", Desc: "隐藏Git状态段"},
			},
		},
		"statusline_seg_claude_version": {
			Title:       "状态栏: 显示Claude版本",
			Description: "在状态栏中显示Claude Code版本。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示Claude版本段"},
				{Label: "禁用", Desc: "隐藏Claude版本段"},
			},
		},
		"statusline_seg_moai_version": {
			Title:       "状态栏: 显示MoAI版本",
			Description: "在状态栏中显示MoAI-ADK版本。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示MoAI版本段"},
				{Label: "禁用", Desc: "隐藏MoAI版本段"},
			},
		},
		"statusline_seg_git_branch": {
			Title:       "状态栏: 显示Git分支",
			Description: "在状态栏中显示当前Git分支名称。",
			Options: []OptionTranslation{
				{Label: "启用", Desc: "显示Git分支段"},
				{Label: "禁用", Desc: "隐藏Git分支段"},
			},
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
