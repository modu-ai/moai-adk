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
				{Label: "Max", Desc: "Opus 5 (max~medium) + Sonnet (low, 단발성 작업만) — Max $200 플랜"},
				{Label: "Medium (권장)", Desc: "Opus 5 (high~low) + Sonnet (low, 단발성 작업만) — Max $100 플랜"},
				{Label: "Low", Desc: "Opus 5 (medium~low) + Sonnet (low, 문서/E2E/단발성 작업) — Plus $20 플랜"},
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
		"autonomy_tier": {
			Title:       "자율성 등급 선택",
			Description: "Claude Code가 표시하는 권한 프롬프트 수를 제어합니다. 'Semi-auto'가 안전한 기본값입니다.",
			Options: []OptionTranslation{
				{Label: "Semi-auto (권장)", Desc: "도구별 프롬프트 — 현재 동작"},
				{Label: "Automatic", Desc: "도구별 자동 승인"},
				{Label: "Fully-autonomous", Desc: "모든 프롬프트 건너뜀 (bypassPermissions); 샌드박스 증명 필요, 킬스위치로 제어"},
			},
		},
		"worktree_auto_create": {
			Title:       "워크트리 자동 생성을 활성화할까요?",
			Description: "활성화하면 MoAI가 run-phase 작업을 위해 격리된 git 워크트리를 자동으로 생성합니다. 기본값은 꺼짐입니다 (L1 워크트리는 Claude Code 런타임이 자체 처리합니다).",
		},
		"audit_model": {
			Title:       "감사 모델 선택",
			Description: "병합을 게이트하는 활성 리뷰 백엔드입니다. 'claude'는 세션 모델을 사용하고, 'codex'/'glm'은 외부 리뷰어를 추가합니다 (부재 시 fail-open).",
			Options: []OptionTranslation{
				{Label: "Claude (권장)", Desc: "세션 모델 리뷰 — 외부 종속성 없음"},
				{Label: "Codex", Desc: "codex CLI 리뷰 (codex 설치 필요, fail-open)"},
				{Label: "GLM", Desc: "z.ai GLM 리뷰 (GLM 키 필요, fail-open)"},
				{Label: "Multi", Desc: "다중 감사자 선언 (수렴 로직은 후속 SPEC, 저장만)"},
			},
		},
		"audit_gate_claude": {
			Title:       "Claude 감사 게이트",
			Description: "off = 건너뜀, advisory = 경고만, required = PASS까지 병합 차단.",
			Options: []OptionTranslation{
				{Label: "Required (권장)", Desc: "Claude 리뷰 PASS까지 병합 차단"},
				{Label: "Advisory", Desc: "경고만 — 차단 없음"},
				{Label: "Off", Desc: "Claude 감사자 건너뜀"},
			},
		},
		"audit_gate_codex": {
			Title:       "Codex 감사 게이트",
			Description: "off = 건너뜀, advisory = 경고만, required = PASS까지 병합 차단 (codex 부재 시 fail-open).",
			Options: []OptionTranslation{
				{Label: "Required (권장)", Desc: "codex PASS까지 병합 차단 (부재 시 fail-open)"},
				{Label: "Advisory", Desc: "경고만 — 차단 없음"},
				{Label: "Off", Desc: "codex 감사자 건너뜀"},
			},
		},
		"audit_gate_glm": {
			Title:       "GLM 감사 게이트",
			Description: "off = 건너뜀, advisory = 경고만, required = PASS까지 병합 차단 (GLM 키 부재 시 fail-open).",
			Options: []OptionTranslation{
				{Label: "Required", Desc: "GLM PASS까지 병합 차단 (키 부재 시 fail-open)"},
				{Label: "Advisory (권장)", Desc: "경고만 — 차단 없음, glm 분산 기본값"},
				{Label: "Off", Desc: "GLM 감사자 건너뜀"},
			},
		},
		"codex_audit_enabled": {
			Title:       "codex 리뷰 게이트를 활성화할까요?",
			Description: "활성화하면 MoAI가 codex Stop-hook 리뷰 게이트를 켭니다 (workflow.codex.review_gate.enabled). 기본값은 꺼짐 — 게이트는 휴면 상태로 배포됩니다.",
		},
		"mcp_tools_opt_in": {
			Title:       "moai MCP 서버를 프로비저닝할까요 (.mcp.json)?",
			Description: "활성화하면 MoAI가 `moai mcp-server` stdio 항목을 .mcp.json에 기록합니다 (옵트인, 신규 프로젝트는 휴면). 기본값은 꺼짐입니다.",
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
				{Label: "Max", Desc: "Opus 5 (max~medium) + Sonnet (low, 単発タスクのみ) — Max $200 プラン"},
				{Label: "Medium (推奨)", Desc: "Opus 5 (high~low) + Sonnet (low, 単発タスクのみ) — Max $100 プラン"},
				{Label: "Low", Desc: "Opus 5 (medium~low) + Sonnet (low, ドキュメント/E2E/単発タスク) — Plus $20 プラン"},
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
		"autonomy_tier": {
			Title:       "自律性レベルを選択",
			Description: "Claude Codeが表示する権限プロンプトの数を制御します。'Semi-auto'が安全なデフォルトです。",
			Options: []OptionTranslation{
				{Label: "Semi-auto (推奨)", Desc: "ツールごとのプロンプト — 現在の動作"},
				{Label: "Automatic", Desc: "ツールごとの自動承認"},
				{Label: "Fully-autonomous", Desc: "すべてのプロンプトをスキップ (bypassPermissions); サンドボックス証明が必要、キルスイッチで制御"},
			},
		},
		"worktree_auto_create": {
			Title:       "ワークツリー自動生成を有効にしますか?",
			Description: "有効にすると、MoAI が run-phase 作業のために分離された git ワークツリーを自動生成します。デフォルトはオフです (L1 ワークツリーは Claude Code ランタイムが自律的に処理します)。",
		},
		"audit_model": {
			Title:       "監査モデルを選択",
			Description: "マージをゲートするアクティブなレビューバックエンドです。'claude' はセッションモデルを使用し、'codex'/'glm' は外部レビューアを追加します (不在時は fail-open)。",
			Options: []OptionTranslation{
				{Label: "Claude (推奨)", Desc: "セッションモデルレビュー — 外部依存なし"},
				{Label: "Codex", Desc: "codex CLI レビュー (codex インストール必須, fail-open)"},
				{Label: "GLM", Desc: "z.ai GLM レビュー (GLM キー必須, fail-open)"},
				{Label: "Multi", Desc: "複数監査者を宣言 (収束ロジックは後続 SPEC, 保存のみ)"},
			},
		},
		"audit_gate_claude": {
			Title:       "Claude 監査ゲート",
			Description: "off = スキップ, advisory = 警告のみ, required = PASS までマージブロック。",
			Options: []OptionTranslation{
				{Label: "Required (推奨)", Desc: "Claude レビュー PASS までマージブロック"},
				{Label: "Advisory", Desc: "警告のみ — ブロックなし"},
				{Label: "Off", Desc: "Claude 監査者をスキップ"},
			},
		},
		"audit_gate_codex": {
			Title:       "Codex 監査ゲート",
			Description: "off = スキップ, advisory = 警告のみ, required = PASS までマージブロック (codex 不在時は fail-open)。",
			Options: []OptionTranslation{
				{Label: "Required (推奨)", Desc: "codex PASS までマージブロック (不在時 fail-open)"},
				{Label: "Advisory", Desc: "警告のみ — ブロックなし"},
				{Label: "Off", Desc: "codex 監査者をスキップ"},
			},
		},
		"audit_gate_glm": {
			Title:       "GLM 監査ゲート",
			Description: "off = スキップ, advisory = 警告のみ, required = PASS までマージブロック (GLM キー不在時は fail-open)。",
			Options: []OptionTranslation{
				{Label: "Required", Desc: "GLM PASS までマージブロック (キー不在時 fail-open)"},
				{Label: "Advisory (推奨)", Desc: "警告のみ — ブロックなし, glm 配布デフォルト"},
				{Label: "Off", Desc: "GLM 監査者をスキップ"},
			},
		},
		"codex_audit_enabled": {
			Title:       "codex レビューゲートを有効にしますか?",
			Description: "有効にすると、MoAI が codex Stop-hook レビューゲートを起動します (workflow.codex.review_gate.enabled)。デフォルトはオフ — ゲートは休止状態で配布されます。",
		},
		"mcp_tools_opt_in": {
			Title:       "moai MCP サーバーをプロビジョニングしますか (.mcp.json)?",
			Description: "有効にすると、MoAI が `moai mcp-server` stdio エントリーを .mcp.json に書き込みます (オプトイン, 新規プロジェクトは休止)。デフォルトはオフです。",
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
				{Label: "Max", Desc: "Opus 5 (max~medium) + Sonnet (low, 仅一次性任务) — Max $200 套餐"},
				{Label: "Medium (推荐)", Desc: "Opus 5 (high~low) + Sonnet (low, 仅一次性任务) — Max $100 套餐"},
				{Label: "Low", Desc: "Opus 5 (medium~low) + Sonnet (low, 文档/E2E/一次性任务) — Plus $20 套餐"},
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
		"autonomy_tier": {
			Title:       "选择自主性等级",
			Description: "控制 Claude Code 显示的权限提示数量。'Semi-auto' 是安全的默认值。",
			Options: []OptionTranslation{
				{Label: "Semi-auto (推荐)", Desc: "逐工具提示 — 当前行为"},
				{Label: "Automatic", Desc: "逐工具自动批准"},
				{Label: "Fully-autonomous", Desc: "跳过所有提示 (bypassPermissions); 需要沙箱证明，受紧急开关控制"},
			},
		},
		"worktree_auto_create": {
			Title:       "是否启用工作树自动创建?",
			Description: "启用后,MoAI 会为 run-phase 工作自动创建隔离的 git 工作树。默认关闭(L1 工作树由 Claude Code 运行时自主处理)。",
		},
		"audit_model": {
			Title:       "选择审计模型",
			Description: " gating 合并的活跃审查后端。'claude' 使用会话模型;'codex'/'glm' 增加外部审查者(缺席时 fail-open)。",
			Options: []OptionTranslation{
				{Label: "Claude (推荐)", Desc: "会话模型审查 — 无外部依赖"},
				{Label: "Codex", Desc: "codex CLI 审查(需安装 codex,fail-open)"},
				{Label: "GLM", Desc: "z.ai GLM 审查(需 GLM key,fail-open)"},
				{Label: "Multi", Desc: "声明多审计者(收敛逻辑为后续 SPEC,仅存储)"},
			},
		},
		"audit_gate_claude": {
			Title:       "Claude 审计闸门",
			Description: "off = 跳过,advisory = 仅警告,required = 阻塞合并直至 PASS。",
			Options: []OptionTranslation{
				{Label: "Required (推荐)", Desc: "阻塞合并直至 Claude 审查 PASS"},
				{Label: "Advisory", Desc: "仅警告 — 不阻塞"},
				{Label: "Off", Desc: "跳过 Claude 审计者"},
			},
		},
		"audit_gate_codex": {
			Title:       "Codex 审计闸门",
			Description: "off = 跳过,advisory = 仅警告,required = 阻塞合并直至 PASS(codex 缺席时 fail-open)。",
			Options: []OptionTranslation{
				{Label: "Required (推荐)", Desc: "阻塞合并直至 codex PASS(缺席 fail-open)"},
				{Label: "Advisory", Desc: "仅警告 — 不阻塞"},
				{Label: "Off", Desc: "跳过 codex 审计者"},
			},
		},
		"audit_gate_glm": {
			Title:       "GLM 审计闸门",
			Description: "off = 跳过,advisory = 仅警告,required = 阻塞合并直至 PASS(GLM key 缺席时 fail-open)。",
			Options: []OptionTranslation{
				{Label: "Required", Desc: "阻塞合并直至 GLM PASS(key 缺席 fail-open)"},
				{Label: "Advisory (推荐)", Desc: "仅警告 — 不阻塞,glm 分发默认"},
				{Label: "Off", Desc: "跳过 GLM 审计者"},
			},
		},
		"codex_audit_enabled": {
			Title:       "是否启用 codex 审查闸门?",
			Description: "启用后,MoAI 激活 codex Stop-hook 审查闸门(workflow.codex.review_gate.enabled)。默认关闭 — 闸门以休眠状态分发。",
		},
		"mcp_tools_opt_in": {
			Title:       "是否 provisioning moai MCP 服务器(.mcp.json)?",
			Description: "启用后,MoAI 将 `moai mcp-server` stdio 条目写入 .mcp.json(opt-in,新项目休眠)。默认关闭。",
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
