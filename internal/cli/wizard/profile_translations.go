package wizard

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

// LocalizeProfileQuestion returns a copy of a profile question with its title
// and description in the given locale. Every other field (id, group, options,
// default) is carried over unchanged. An unknown locale, an empty locale, and
// an id absent from the profile table all return the question as given.
func LocalizeProfileQuestion(q *Question, locale string) Question {
	localized := *q
	tr, ok := profileQuestionTexts[locale][q.ID]
	if !ok {
		return localized
	}
	localized.Title = tr.Title
	localized.Description = tr.Description
	return localized
}
