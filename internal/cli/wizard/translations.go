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
		"model_policy": {
			Title:       "모델 정책 선택",
			Description: "각 에이전트에 할당되는 Claude 모델 등급을 제어합니다. Claude 플랜에 맞추세요.",
			Options: []OptionTranslation{
				{Label: "Max", Desc: "Opus 5 (high~medium) + Sonnet (low, 문서/단발성 작업) — Max $200 플랜"},
				{Label: "Medium (권장)", Desc: "Opus 5 (high~low) + Sonnet (low, 문서/단발성 작업) — Max $100 플랜"},
				{Label: "Low", Desc: "Opus 5 (high~low) + Sonnet (low, 문서/E2E/단발성 작업) — Plus $20 플랜"},
			},
		},
		"project_mode": {
			Title:       "프로젝트 모드 선택",
			Description: "협업 설정을 제어합니다. 솔로 개발자는 'personal'이 권장 기본값입니다.",
			Options: []OptionTranslation{
				{Label: "Personal (권장)", Desc: "솔로 개발자 — 팀 조율 오버헤드 없음"},
				{Label: "Team", Desc: "다중 개발자 환경 — 팀 협업 기능 활성화"},
			},
		},
		"worktree_auto_create": {
			Title:       "워크트리 자동 생성을 활성화할까요?",
			Description: "활성화하면 moai init / moai profile / moai web이 자동으로 워크트리에 진입합니다. 기본은 비활성화입니다(솔로 개발자 권장).",
		},
		"todo_enabled": {
			Title:       "백로그 큐(todo)를 사용할까요?",
			Description: "사용하지 않으면 백로그 안내가 먼저 뜨지 않습니다 — 세션 시작 시 대기 카드 요약도, 상태줄 TODO 표시도 없습니다. `moai todo` 명령과 직접 부른 `/moai todo` 는 어느 쪽이든 그대로 동작합니다.",
		},
		"feedback_auto_submit": {
			Title:       "확인 절차 없이 피드백을 제출할까요?",
			Description: "끄면(기본값) 피드백 워크플로가 마스킹된 제목과 본문을 보여주고 공개 이슈를 열기 전에 한 번 묻습니다. 켜면 그 확인 단계를 건너뜁니다.",
		},
		"autonomy_tier": {
			Title:       "자율성 등급 선택",
			Description: "프롬프트 없이 세션이 몇 턴까지 실행될지 제어합니다. 'semi-auto'가 권장 기본값입니다.",
			Options: []OptionTranslation{
				{Label: "Semi-auto (권장)", Desc: "중요하지 않은 동작 전에 항상 확인"},
				{Label: "Automatic", Desc: "마일스톤 자율 실행; 게이트에서 확인"},
				{Label: "Fully-autonomous", Desc: "샌드박스 증명 필요 (Docker/gVisor 등)"},
			},
		},
		"audit_model": {
			Title:       "감사 모델 선택",
			Description: "활성 감사 백엔드. 'claude'가 배포 기본값입니다.",
			Options: []OptionTranslation{
				{Label: "Claude", Desc: "기본 앵커 판정"},
				{Label: "Codex", Desc: "Codex JSON-RPC 검토자"},
				{Label: "GLM", Desc: "GLM (z.ai) 검토자"},
				{Label: "Multi", Desc: "다중 검토자 수렴 (지연됨)"},
			},
		},
		"audit_gate_claude": {
			Title:       "Claude 감사 게이트",
			Description: "Claude 앵커 판정의 게이트.",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "Claude 감사자 비활성화"},
				{Label: "Advisory", Desc: "실행하되 차단 안 함"},
				{Label: "Required", Desc: "실패 시 수렴 차단"},
			},
		},
		"audit_gate_codex": {
			Title:       "Codex 감사 게이트",
			Description: "Codex 검토자의 게이트.",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "Codex 감사자 비활성화"},
				{Label: "Advisory", Desc: "실행하되 차단 안 함"},
				{Label: "Required", Desc: "실패 시 수렴 차단"},
			},
		},
		"audit_gate_glm": {
			Title:       "GLM 감사 게이트",
			Description: "GLM 검토자의 게이트. 기본 'advisory' — GLM 키가 없어도 차단하지 않습니다.",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "GLM 감사자 비활성화"},
				{Label: "Advisory", Desc: "실행하되 차단 안 함"},
				{Label: "Required", Desc: "실패 시 수렴 차단"},
			},
		},
		"codex_audit_enabled": {
			Title:       "Codex 검토 게이트 Stop 훅을 활성화할까요?",
			Description: "기본 비활성화. 활성화하면 Stop 훅이 미커밋 변경 사항에 대해 codex를 실행합니다.",
		},
		"mcp_provision": {
			Title:       "moai MCP 서버를 프로비저닝할까요?",
			Description: "기본 활성화. 건너뛰려면 아니요를 선택하세요.",
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
		"model_policy": {
			Title:       "モデルポリシーを選択",
			Description: "各エージェントに割り当てる Claude モデルのティアを制御します。ご利用の Claude プランに合わせてください。",
			Options: []OptionTranslation{
				{Label: "Max", Desc: "Opus 5 (high~medium) + Sonnet (low, ドキュメント/単発タスク) — Max $200 プラン"},
				{Label: "Medium (推奨)", Desc: "Opus 5 (high~low) + Sonnet (low, ドキュメント/単発タスク) — Max $100 プラン"},
				{Label: "Low", Desc: "Opus 5 (high~low) + Sonnet (low, ドキュメント/E2E/単発タスク) — Plus $20 プラン"},
			},
		},
		"project_mode": {
			Title:       "プロジェクトモードを選択",
			Description: "コラボレーション設定を制御します。ソロ開発者には 'personal' が推奨デフォルトです。",
			Options: []OptionTranslation{
				{Label: "Personal (推奨)", Desc: "ソロ開発者 — チーム調整のオーバーヘッドなし"},
				{Label: "Team", Desc: "複数人開発 — チームコラボレーション機能を有効化"},
			},
		},
		"worktree_auto_create": {
			Title:       "ワークツリー自動作成を有効にしますか?",
			Description: "有効にすると、moai init / moai profile / moai web が自動的にワークツリーに入ります。既定は無効です(ソロ開発者推奨)。",
		},
		"todo_enabled": {
			Title:       "バックログキュー(todo)を使いますか?",
			Description: "無効にすると、バックログの案内が出なくなります — セッション開始時の待機カード要約も、ステータスラインの TODO 表示もありません。`moai todo` コマンドと明示的な `/moai todo` はどちらの設定でもそのまま動きます。",
		},
		"feedback_auto_submit": {
			Title:       "確認なしでフィードバックを送信しますか?",
			Description: "無効(既定)の場合、フィードバックワークフローはマスク済みのタイトルと本文を表示し、公開 issue を作成する前に一度確認します。有効にするとその確認を省略します。",
		},
		"autonomy_tier": {
			Title:       "自律レベルを選択",
			Description: "プロンプトなしでセッションが何ターン実行するかを制御します。'semi-auto' が推奨デフォルトです。",
			Options: []OptionTranslation{
				{Label: "Semi-auto (推奨)", Desc: "重要でない操作の前に常に確認"},
				{Label: "Automatic", Desc: "マイルストーンを自律実行; ゲートで確認"},
				{Label: "Fully-autonomous", Desc: "サンドボックス証明が必要 (Docker/gVisor 等)"},
			},
		},
		"audit_model": {
			Title:       "監査モデルを選択",
			Description: "アクティブな監査バックエンド。'claude' が配布デフォルトです。",
			Options: []OptionTranslation{
				{Label: "Claude", Desc: "デフォルト アンカー評決"},
				{Label: "Codex", Desc: "Codex JSON-RPC レビューア"},
				{Label: "GLM", Desc: "GLM (z.ai) レビューア"},
				{Label: "Multi", Desc: "マルチレビューアー収束 (延期)"},
			},
		},
		"audit_gate_claude": {
			Title:       "Claude 監査ゲート",
			Description: "Claude アンカー評決のゲート。",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "Claude 監査を無効化"},
				{Label: "Advisory", Desc: "実行するがブロックしない"},
				{Label: "Required", Desc: "失敗時収束ブロック"},
			},
		},
		"audit_gate_codex": {
			Title:       "Codex 監査ゲート",
			Description: "Codex レビューアのゲート。",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "Codex 監査を無効化"},
				{Label: "Advisory", Desc: "実行するがブロックしない"},
				{Label: "Required", Desc: "失敗時収束ブロック"},
			},
		},
		"audit_gate_glm": {
			Title:       "GLM 監査ゲート",
			Description: "GLM レビューアのゲート。デフォルト 'advisory' — GLM キーがなくてもブロックしません。",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "GLM 監査を無効化"},
				{Label: "Advisory", Desc: "実行するがブロックしない"},
				{Label: "Required", Desc: "失敗時収束ブロック"},
			},
		},
		"codex_audit_enabled": {
			Title:       "Codex レビューゲート Stop フックを有効にしますか?",
			Description: "デフォルト無効。有効化すると Stop フックが未コミット変更に codex を実行します。",
		},
		"mcp_provision": {
			Title:       "moai MCP サーバーをプロビジョニングしますか?",
			Description: "デフォルト有効。スキップする場合はいいえを選択してください。",
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
		"model_policy": {
			Title:       "选择模型策略",
			Description: "控制为每个智能体分配的 Claude 模型等级。请与您的 Claude 套餐匹配。",
			Options: []OptionTranslation{
				{Label: "Max", Desc: "Opus 5 (high~medium) + Sonnet (low, 文档/一次性任务) — Max $200 套餐"},
				{Label: "Medium (推荐)", Desc: "Opus 5 (high~low) + Sonnet (low, 文档/一次性任务) — Max $100 套餐"},
				{Label: "Low", Desc: "Opus 5 (high~low) + Sonnet (low, 文档/E2E/一次性任务) — Plus $20 套餐"},
			},
		},
		"project_mode": {
			Title:       "选择项目模式",
			Description: "控制协作设置。单人开发者推荐使用 'personal' 默认值。",
			Options: []OptionTranslation{
				{Label: "Personal (推荐)", Desc: "单人开发者 — 无团队协调开销"},
				{Label: "Team", Desc: "多人开发 — 启用团队协作功能"},
			},
		},
		"worktree_auto_create": {
			Title:       "是否启用工作树自动创建?",
			Description: "启用后,moai init / moai profile / moai web 会自动进入工作树。默认关闭(推荐单人开发者)。",
		},
		"todo_enabled": {
			Title:       "是否使用待办队列(todo)?",
			Description: "关闭后将不再主动提示待办内容 — 会话开始时不显示等待卡片数量,状态栏也不显示 TODO。无论开关如何,`moai todo` 命令和显式调用的 `/moai todo` 都照常工作。",
		},
		"feedback_auto_submit": {
			Title:       "是否跳过确认直接提交反馈?",
			Description: "关闭时(默认),反馈流程会先展示脱敏后的标题与正文,并在创建公开 issue 前询问一次。开启后将跳过该确认步骤。",
		},
		"autonomy_tier": {
			Title:       "选择自主等级",
			Description: "控制会话在不提示的情况下运行多少轮。'semi-auto' 是推荐默认值。",
			Options: []OptionTranslation{
				{Label: "Semi-auto (推荐)", Desc: "每个重要操作前都确认"},
				{Label: "Automatic", Desc: "自主运行里程碑;在关卡确认"},
				{Label: "Fully-autonomous", Desc: "需要沙箱证明 (Docker/gVisor 等)"},
			},
		},
		"audit_model": {
			Title:       "选择审计模型",
			Description: "活跃的审计后端。'claude' 是锁定分发默认值。",
			Options: []OptionTranslation{
				{Label: "Claude", Desc: "默认锚定裁决"},
				{Label: "Codex", Desc: "Codex JSON-RPC 审查者"},
				{Label: "GLM", Desc: "GLM (z.ai) 审查者"},
				{Label: "Multi", Desc: "多审查者收敛(延期)"},
			},
		},
		"audit_gate_claude": {
			Title:       "Claude 审计关卡",
			Description: "Claude 锚定裁决的关卡。",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "禁用 Claude 审计者"},
				{Label: "Advisory", Desc: "运行但不阻塞"},
				{Label: "Required", Desc: "失败阻塞收敛"},
			},
		},
		"audit_gate_codex": {
			Title:       "Codex 审计关卡",
			Description: "Codex 审查者的关卡。",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "禁用 Codex 审计者"},
				{Label: "Advisory", Desc: "运行但不阻塞"},
				{Label: "Required", Desc: "失败阻塞收敛"},
			},
		},
		"audit_gate_glm": {
			Title:       "GLM 审计关卡",
			Description: "GLM 审查者的关卡。默认 'advisory' — 缺少 GLM key 不会阻塞(失败开放)。",
			Options: []OptionTranslation{
				{Label: "Off", Desc: "禁用 GLM 审计者"},
				{Label: "Advisory", Desc: "运行但不阻塞"},
				{Label: "Required", Desc: "失败阻塞收敛"},
			},
		},
		"codex_audit_enabled": {
			Title:       "是否启用 Codex 审查关卡 Stop 钩子?",
			Description: "默认关闭。启用后 Stop 钩子对未提交变更运行 codex。",
		},
		"mcp_provision": {
			Title:       "是否供应 moai MCP 服务器?",
			Description: "默认开启。如需跳过请选择否。",
		},
	},
}

// uiStrings maps language code to UI strings.
var uiStrings = map[string]UIStrings{
	"en": {
		HelpSelect:    "Use arrow keys to navigate, Enter to select, Esc to cancel",
		HelpInput:     "Type your answer, Enter to confirm, Esc to cancel",
		ErrorRequired: "This field is required",
		ConfirmYes:    "Yes",
		ConfirmNo:     "No",
	},
	"ko": {
		HelpSelect:    "방향키로 이동, Enter로 선택, Esc로 취소",
		HelpInput:     "답변 입력 후 Enter로 확인, Esc로 취소",
		ErrorRequired: "필수 입력 항목입니다",
		ConfirmYes:    "예",
		ConfirmNo:     "아니오",
	},
	"ja": {
		HelpSelect:    "矢印キーで移動、Enterで選択、Escでキャンセル",
		HelpInput:     "入力してEnterで確定、Escでキャンセル",
		ErrorRequired: "この項目は必須です",
		ConfirmYes:    "はい",
		ConfirmNo:     "いいえ",
	},
	"zh": {
		HelpSelect:    "使用方向键导航，Enter选择，Esc取消",
		HelpInput:     "输入答案，Enter确认，Esc取消",
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
