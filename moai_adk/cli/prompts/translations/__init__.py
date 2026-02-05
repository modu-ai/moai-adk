"""Multilingual prompt translations for CLI.

Provides localized messages for init prompts in multiple languages.
"""

from typing import TypedDict


class InitTranslations(TypedDict):
    """Translation strings for init prompts."""

    # Headers
    language_selection: str
    user_setup: str  # New: User name setup header
    service_selection: str
    pricing_selection: str
    api_key_input: str
    project_setup: str
    git_setup: str
    output_language: str
    claude_auth_selection: str  # New: Claude authentication method selection

    # Questions
    q_language: str
    q_user_name: str  # New: User name question
    q_service: str
    q_claude_auth_type: str  # New: Claude auth type question
    q_pricing_claude: str
    q_pricing_glm: str
    q_api_key_anthropic: str
    q_api_key_glm: str
    q_project_name: str
    q_git_mode: str
    q_github_username: str
    q_commit_lang: str
    q_comment_lang: str
    q_doc_lang: str

    # Options - Service
    opt_claude_subscription: str
    opt_claude_api: str
    opt_glm: str
    opt_hybrid: str

    # Options - Claude Auth Type (for hybrid)
    opt_claude_sub: str  # New
    opt_claude_api_key: str  # New
    desc_claude_sub: str  # New
    desc_claude_api_key: str  # New

    # Options - Pricing Claude
    opt_pro: str
    opt_max5: str
    opt_max20: str

    # Options - Pricing GLM
    opt_basic: str
    opt_glm_pro: str
    opt_enterprise: str

    # Options - Git
    opt_manual: str
    opt_personal: str
    opt_team: str

    # Descriptions
    desc_claude_subscription: str
    desc_claude_api: str
    desc_glm: str
    desc_hybrid: str
    desc_pro: str
    desc_max5: str
    desc_max20: str
    desc_basic: str
    desc_glm_pro: str
    desc_enterprise: str
    desc_manual: str
    desc_personal: str
    desc_team: str

    # Messages
    msg_api_key_stored: str
    msg_glm_key_found: str  # New: GLM key found message
    msg_glm_key_keep_prompt: str  # New: Prompt to keep or replace existing key
    msg_glm_key_skip_guidance: str  # New: Guidance when skipping GLM API key
    msg_setup_complete: str
    msg_cancelled: str
    msg_current_dir: str
    msg_skip_same_lang: str

    # Development Methodology Info
    dev_mode_ddd_info: str  # DDD methodology info message


TRANSLATIONS: dict[str, InitTranslations] = {
    "ko": {
        # Headers
        "language_selection": "🌐 대화 언어 선택",
        "user_setup": "👤 사용자 설정",
        "service_selection": "💳 서비스 선택",
        "pricing_selection": "💰 요금제 선택",
        "api_key_input": "🔑 API 키 입력",
        "project_setup": "📁 프로젝트 설정",
        "git_setup": "🔀 Git 설정",
        "output_language": "🗣️ 출력 언어 설정",
        "claude_auth_selection": "🔐 Claude 인증 방식",
        # Questions
        "q_language": "대화 언어를 선택하세요:",
        "q_user_name": "사용자 이름을 입력하세요 (선택사항):",
        "q_service": "사용할 서비스를 선택하세요:",
        "q_claude_auth_type": "Claude 인증 방식을 선택하세요:",
        "q_pricing_claude": "Claude 요금제를 선택하세요:",
        "q_pricing_glm": "GLM CodePlan 요금제를 선택하세요:",
        "q_api_key_anthropic": "Anthropic API 키를 입력하세요:",
        "q_api_key_glm": "GLM API 키를 입력하세요:",
        "q_project_name": "프로젝트 이름:",
        "q_git_mode": "Git 모드를 선택하세요:",
        "q_github_username": "GitHub 사용자명:",
        "q_commit_lang": "커밋 메시지 언어:",
        "q_comment_lang": "코드 주석 언어:",
        "q_doc_lang": "문서 언어:",
        # Options - Service
        "opt_claude_subscription": "Claude 구독",
        "opt_claude_api": "Claude API",
        "opt_glm": "GLM CodePlan",
        "opt_hybrid": "Claude + GLM (하이브리드)",
        # Options - Claude Auth Type (for hybrid)
        "opt_claude_sub": "구독",
        "opt_claude_api_key": "API 키",
        "desc_claude_sub": "Claude Code 구독 사용",
        "desc_claude_api_key": "API 키 직접 입력",
        # Options - Pricing Claude
        "opt_pro": "Pro ($20/월)",
        "opt_max5": "Max5 ($100/월)",
        "opt_max20": "Max20 ($200/월)",
        # Options - Pricing GLM
        "opt_basic": "Basic",
        "opt_glm_pro": "Pro",
        "opt_enterprise": "Enterprise",
        # Options - Git
        "opt_manual": "manual (로컬만)",
        "opt_personal": "personal (GitHub 개인)",
        "opt_team": "team (GitHub 팀)",
        # Descriptions
        "desc_claude_subscription": "Claude Code 구독 중 - API 키 불필요",
        "desc_claude_api": "직접 API 키 입력",
        "desc_glm": "GLM CodePlan 서비스 사용",
        "desc_hybrid": "비용 최적화 자동 배정",
        "desc_pro": "sonnet 중심, 기본 사용",
        "desc_max5": "opus 일부 사용 가능",
        "desc_max20": "opus 자유 사용",
        "desc_basic": "기본 기능",
        "desc_glm_pro": "고급 기능",
        "desc_enterprise": "전체 기능",
        "desc_manual": "로컬 저장소만 사용",
        "desc_personal": "GitHub 개인 계정 사용",
        "desc_team": "GitHub 팀/조직 사용",
        # Messages
        "msg_api_key_stored": "GLM API 키가 ~/.moai/.env.glm에 저장되었습니다",
        "msg_glm_key_found": "기존 GLM API 키를 찾았습니다:",
        "msg_glm_key_keep_prompt": "Enter를 누르면 기존 키 유지, 새 키를 입력하면 교체됩니다",
        "msg_glm_key_skip_guidance": "나중에 'moai glm <키>' 또는 'moai update'로 추가할 수 있습니다",
        "msg_setup_complete": "✅ 설정이 완료되었습니다!",
        "msg_cancelled": "사용자에 의해 설정이 취소되었습니다",
        "msg_current_dir": "(현재 디렉토리)",
        "msg_skip_same_lang": "대화 언어와 동일하게 설정됨",
        # Development Methodology Info
        "dev_mode_ddd_info": "DDD (Domain-Driven Development) 방법론을 사용합니다",
    },
    "en": {
        # Headers
        "language_selection": "🌐 Language Selection",
        "user_setup": "👤 User Setup",
        "service_selection": "💳 Service Selection",
        "pricing_selection": "💰 Pricing Plan",
        "api_key_input": "🔑 API Key Input",
        "project_setup": "📁 Project Setup",
        "git_setup": "🔀 Git Setup",
        "output_language": "🗣️ Output Language Settings",
        "claude_auth_selection": "🔐 Claude Authentication",
        # Questions
        "q_language": "Select your conversation language:",
        "q_user_name": "Enter your name (optional):",
        "q_service": "Select the service to use:",
        "q_claude_auth_type": "Select Claude authentication method:",
        "q_pricing_claude": "Select Claude pricing plan:",
        "q_pricing_glm": "Select GLM CodePlan pricing plan:",
        "q_api_key_anthropic": "Enter your Anthropic API key:",
        "q_api_key_glm": "Enter your GLM API key:",
        "q_project_name": "Project name:",
        "q_git_mode": "Select Git mode:",
        "q_github_username": "GitHub username:",
        "q_commit_lang": "Commit message language:",
        "q_comment_lang": "Code comment language:",
        "q_doc_lang": "Documentation language:",
        # Options - Service
        "opt_claude_subscription": "Claude Subscription",
        "opt_claude_api": "Claude API",
        "opt_glm": "GLM CodePlan",
        "opt_hybrid": "Claude + GLM (Hybrid)",
        # Options - Claude Auth Type (for hybrid)
        "opt_claude_sub": "Subscription",
        "opt_claude_api_key": "API Key",
        "desc_claude_sub": "Use Claude Code subscription",
        "desc_claude_api_key": "Enter API key directly",
        # Options - Pricing Claude
        "opt_pro": "Pro ($20/mo)",
        "opt_max5": "Max5 ($100/mo)",
        "opt_max20": "Max20 ($200/mo)",
        # Options - Pricing GLM
        "opt_basic": "Basic",
        "opt_glm_pro": "Pro",
        "opt_enterprise": "Enterprise",
        # Options - Git
        "opt_manual": "manual (local only)",
        "opt_personal": "personal (GitHub personal)",
        "opt_team": "team (GitHub team)",
        # Descriptions
        "desc_claude_subscription": "Claude Code subscriber - No API key needed",
        "desc_claude_api": "Enter API key directly",
        "desc_glm": "Use GLM CodePlan service",
        "desc_hybrid": "Cost-optimized automatic allocation",
        "desc_pro": "Sonnet-focused, basic usage",
        "desc_max5": "Opus partially available",
        "desc_max20": "Opus freely available",
        "desc_basic": "Basic features",
        "desc_glm_pro": "Advanced features",
        "desc_enterprise": "Full features",
        "desc_manual": "Local repository only",
        "desc_personal": "GitHub personal account",
        "desc_team": "GitHub team/organization",
        # Messages
        "msg_api_key_stored": "GLM API key stored in ~/.moai/.env.glm",
        "msg_glm_key_found": "Existing GLM API key found:",
        "msg_glm_key_keep_prompt": "Press Enter to keep existing key, or type new key to replace",
        "msg_glm_key_skip_guidance": "You can add it later with 'moai glm <key>' or 'moai update'",
        "msg_setup_complete": "✅ Setup complete!",
        "msg_cancelled": "Setup cancelled by user",
        "msg_current_dir": "(current directory)",
        "msg_skip_same_lang": "Set to same as conversation language",
        # Development Methodology Info
        "dev_mode_ddd_info": "Using DDD (Domain-Driven Development) methodology",
    },
    "ja": {
        # Headers
        "language_selection": "🌐 言語選択",
        "user_setup": "👤 ユーザー設定",
        "service_selection": "💳 サービス選択",
        "pricing_selection": "💰 料金プラン",
        "api_key_input": "🔑 APIキー入力",
        "project_setup": "📁 プロジェクト設定",
        "git_setup": "🔀 Git設定",
        "output_language": "🗣️ 出力言語設定",
        "claude_auth_selection": "🔐 Claude認証方式",
        # Questions
        "q_language": "会話言語を選択してください:",
        "q_user_name": "お名前を入力してください（任意）:",
        "q_service": "使用するサービスを選択してください:",
        "q_claude_auth_type": "Claude認証方式を選択してください:",
        "q_pricing_claude": "Claudeの料金プランを選択してください:",
        "q_pricing_glm": "GLM CodePlanの料金プランを選択してください:",
        "q_api_key_anthropic": "Anthropic APIキーを入力してください:",
        "q_api_key_glm": "GLM APIキーを入力してください:",
        "q_project_name": "プロジェクト名:",
        "q_git_mode": "Gitモードを選択してください:",
        "q_github_username": "GitHubユーザー名:",
        "q_commit_lang": "コミットメッセージ言語:",
        "q_comment_lang": "コードコメント言語:",
        "q_doc_lang": "ドキュメント言語:",
        # Options - Service
        "opt_claude_subscription": "Claude サブスクリプション",
        "opt_claude_api": "Claude API",
        "opt_glm": "GLM CodePlan",
        "opt_hybrid": "Claude + GLM (ハイブリッド)",
        # Options - Claude Auth Type (for hybrid)
        "opt_claude_sub": "サブスクリプション",
        "opt_claude_api_key": "APIキー",
        "desc_claude_sub": "Claude Codeサブスクリプション使用",
        "desc_claude_api_key": "APIキーを直接入力",
        # Options - Pricing Claude
        "opt_pro": "Pro ($20/月)",
        "opt_max5": "Max5 ($100/月)",
        "opt_max20": "Max20 ($200/月)",
        # Options - Pricing GLM
        "opt_basic": "Basic",
        "opt_glm_pro": "Pro",
        "opt_enterprise": "Enterprise",
        # Options - Git
        "opt_manual": "manual (ローカルのみ)",
        "opt_personal": "personal (GitHub個人)",
        "opt_team": "team (GitHubチーム)",
        # Descriptions
        "desc_claude_subscription": "Claude Code購読中 - APIキー不要",
        "desc_claude_api": "直接APIキーを入力",
        "desc_glm": "GLM CodePlanサービスを使用",
        "desc_hybrid": "コスト最適化自動割り当て",
        "desc_pro": "Sonnet中心、基本使用",
        "desc_max5": "Opus一部使用可能",
        "desc_max20": "Opus自由使用",
        "desc_basic": "基本機能",
        "desc_glm_pro": "高度な機能",
        "desc_enterprise": "全機能",
        "desc_manual": "ローカルリポジトリのみ",
        "desc_personal": "GitHub個人アカウント",
        "desc_team": "GitHubチーム/組織",
        # Messages
        "msg_api_key_stored": "GLM APIキーが~/.moai/.env.glmに保存されました",
        "msg_glm_key_found": "既存のGLM APIキーが見つかりました:",
        "msg_glm_key_keep_prompt": "Enterキーで既存のキーを保持、新しいキーを入力すると置換",
        "msg_glm_key_skip_guidance": "後で 'moai glm <キー>' または 'moai update' で追加できます",
        "msg_setup_complete": "✅ 設定完了！",
        "msg_cancelled": "ユーザーにより設定がキャンセルされました",
        "msg_current_dir": "(現在のディレクトリ)",
        "msg_skip_same_lang": "会話言語と同じに設定",
        # Development Methodology Info
        "dev_mode_ddd_info": "DDD (Domain-Driven Development) 方法論を使用します",
    },
    "zh": {
        # Headers
        "language_selection": "🌐 语言选择",
        "user_setup": "👤 用户设置",
        "service_selection": "💳 服务选择",
        "pricing_selection": "💰 定价方案",
        "api_key_input": "🔑 API密钥输入",
        "project_setup": "📁 项目设置",
        "git_setup": "🔀 Git设置",
        "output_language": "🗣️ 输出语言设置",
        "claude_auth_selection": "🔐 Claude认证方式",
        # Questions
        "q_language": "选择您的对话语言:",
        "q_user_name": "请输入您的姓名（可选）:",
        "q_service": "选择要使用的服务:",
        "q_claude_auth_type": "选择Claude认证方式:",
        "q_pricing_claude": "选择Claude定价方案:",
        "q_pricing_glm": "选择GLM CodePlan定价方案:",
        "q_api_key_anthropic": "输入您的Anthropic API密钥:",
        "q_api_key_glm": "输入您的GLM API密钥:",
        "q_project_name": "项目名称:",
        "q_git_mode": "选择Git模式:",
        "q_github_username": "GitHub用户名:",
        "q_commit_lang": "提交消息语言:",
        "q_comment_lang": "代码注释语言:",
        "q_doc_lang": "文档语言:",
        # Options - Service
        "opt_claude_subscription": "Claude 订阅",
        "opt_claude_api": "Claude API",
        "opt_glm": "GLM CodePlan",
        "opt_hybrid": "Claude + GLM (混合)",
        # Options - Claude Auth Type (for hybrid)
        "opt_claude_sub": "订阅",
        "opt_claude_api_key": "API密钥",
        "desc_claude_sub": "使用Claude Code订阅",
        "desc_claude_api_key": "直接输入API密钥",
        # Options - Pricing Claude
        "opt_pro": "Pro ($20/月)",
        "opt_max5": "Max5 ($100/月)",
        "opt_max20": "Max20 ($200/月)",
        # Options - Pricing GLM
        "opt_basic": "Basic",
        "opt_glm_pro": "Pro",
        "opt_enterprise": "Enterprise",
        # Options - Git
        "opt_manual": "manual (仅本地)",
        "opt_personal": "personal (GitHub个人)",
        "opt_team": "team (GitHub团队)",
        # Descriptions
        "desc_claude_subscription": "Claude Code订阅中 - 无需API密钥",
        "desc_claude_api": "直接输入API密钥",
        "desc_glm": "使用GLM CodePlan服务",
        "desc_hybrid": "成本优化自动分配",
        "desc_pro": "以Sonnet为主，基本使用",
        "desc_max5": "可部分使用Opus",
        "desc_max20": "自由使用Opus",
        "desc_basic": "基本功能",
        "desc_glm_pro": "高级功能",
        "desc_enterprise": "全部功能",
        "desc_manual": "仅本地仓库",
        "desc_personal": "GitHub个人账户",
        "desc_team": "GitHub团队/组织",
        # Messages
        "msg_api_key_stored": "GLM API密钥已保存到~/.moai/.env.glm",
        "msg_glm_key_found": "找到现有GLM API密钥:",
        "msg_glm_key_keep_prompt": "按Enter保留现有密钥,或输入新密钥进行替换",
        "msg_glm_key_skip_guidance": "您可以稍后使用 'moai glm <密钥>' 或 'moai update' 添加",
        "msg_setup_complete": "✅ 设置完成！",
        "msg_cancelled": "用户取消设置",
        "msg_current_dir": "(当前目录)",
        "msg_skip_same_lang": "设置为与对话语言相同",
        # Development Methodology Info
        "dev_mode_ddd_info": "使用DDD (领域驱动开发) 方法论",
    },
}


def get_translation(locale: str) -> InitTranslations:
    """Get translations for the specified locale.

    Args:
        locale: Language code (ko, en, ja, zh)

    Returns:
        Translation dictionary for the locale, defaults to English
    """
    return TRANSLATIONS.get(locale, TRANSLATIONS["en"])


__all__ = ["InitTranslations", "TRANSLATIONS", "get_translation"]
