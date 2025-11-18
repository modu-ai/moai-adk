"""
Automatic announcement translation system for companyAnnouncements in .claude/settings.json

This module copies and translates the 23 reference English announcements into the user's
selected language during moai-adk init and update operations.

Supported Languages:
- Hardcoded: English (en), Korean (ko), Japanese (ja), Chinese (zh)
- Other languages: Default to English fallback
"""

import json
from pathlib import Path
from typing import List, Optional

# Reference English announcements (23 items - synced with settings.json template)
REFERENCE_ANNOUNCEMENTS_EN = [
    "SPEC-First: Use /moai:1-plan to define requirements first and everything connects together",
    "TDD Cycle: RED (write tests) → GREEN (minimal code) → REFACTOR (improve) ensures quality",
    "4-Phase Workflow: /moai:1-plan plan → /moai:2-run implement → /moai:3-sync validate → git deploy",
    "TRUST 5 Principles: Test(≥85%) + Readable + Unified + Secured(OWASP) + Trackable = Quality",
    "moai Agents: 19 expert team members automatically handle planning, implementation, testing, validation",
    "Conversation Style: /output-style r2d2 (hands-on partner) or /output-style yoda (deep learning) available",
    "Task Tracking: TodoWrite keeps progress visible and prevents any tasks from being missed",
    "55+ Skill Library: Verified patterns and best practices enable hallucination-free implementation",
    "Auto Validation: After coding, TRUST 5, test coverage, and security checks run automatically",
    "GitFlow Strategy: feature/SPEC-XXX → develop → main structure for safe deployment pipeline",
    "Context7 Updates: All library versions stay current with real-time API information",
    "Parallel Processing: Independent tasks (tests, docs, deployment) run simultaneously to save time",
    "Health Check: moai-adk doctor diagnoses configuration, version, and dependency issues in seconds",
    "Doc Sync: /moai:3-sync auto synchronizes tests, code, and documentation automatically",
    "Language Separation: Communicate in your language, write code in project language, package in English",
    "Multi-language Support: Korean, English, Japanese, Spanish, and 25+ languages for conversation",
    "Safe Updates: moai-adk update brings new features automatically while preserving existing settings",
    "Auto Cleanup: Session end automatically cleans .moai/temp/ cache and logs to save space",
    "Error Recovery: Failed commits and merge conflicts are auto-analyzed with solutions provided",
    "Security First: Environment variables, API keys, and credentials auto-added to .gitignore for protection",
    "Context Optimization: Efficient use of 200K token window enables handling of large projects",
    "Quick Feedback: Questions? Ask immediately - moai interprets intent and clarifies automatically",
    "Clean Exit: End sessions with /clear to reset context and prepare for the next task"
]

# Hardcoded Korean translations (23 items)
ANNOUNCEMENTS_KO = [
    "SPEC 우선: /moai:1-plan으로 요구사항을 먼저 정의하면 모든 것이 연결됩니다",
    "TDD Cycle: RED (테스트 작성) → GREEN (최소 코드) → REFACTOR (개선)로 품질을 보장합니다",
    "4단계 워크플로우: /moai:1-plan 계획 → /moai:2-run 구현 → /moai:3-sync 검증 → git 배포",
    "TRUST 5 원칙: Test(≥85%) + Readable + Unified + Secured(OWASP) + Trackable = 품질",
    "moai Agents: 19명의 전문가 팀이 계획, 구현, 테스트, 검증을 자동으로 처리합니다",
    "대화 스타일: /output-style r2d2 (실습 파트너) 또는 /output-style yoda (심화 학습) 사용 가능",
    "작업 추적: TodoWrite로 진행 상황을 계속 표시하여 누락되는 작업이 없습니다",
    "55+ Skill 라이브러리: 검증된 패턴과 모범 사례로 환각 없는 구현을 가능하게 합니다",
    "자동 검증: 코딩 후 TRUST 5, 테스트 커버리지, 보안 검사가 자동으로 실행됩니다",
    "GitFlow 전략: feature/SPEC-XXX → develop → main 구조로 안전한 배포 파이프라인",
    "Context7 업데이트: 모든 라이브러리 버전이 실시간 API 정보로 최신 상태를 유지합니다",
    "병렬 처리: 독립적인 작업(테스트, 문서, 배포)을 동시에 실행하여 시간을 절약합니다",
    "상태 확인: moai-adk doctor로 구성, 버전, 종속성 문제를 몇 초 만에 진단합니다",
    "문서 동기화: /moai:3-sync로 테스트, 코드, 문서를 자동으로 동기화합니다",
    "언어 분리: 당신의 언어로 소통하고, 프로젝트 언어로 코드를 작성하며, 영어로 패키징합니다",
    "다국어 지원: 한국어, 영어, 일본어, 스페인어 등 25개 이상 언어로 대화할 수 있습니다",
    "안전한 업데이트: moai-adk update로 기존 설정을 유지하면서 새 기능을 자동으로 추가합니다",
    "자동 정리: 세션 종료 시 .moai/temp/ 캐시와 로그를 자동으로 정리하여 공간을 절약합니다",
    "오류 복구: 실패한 커밋과 병합 충돌을 자동으로 분석하고 해결책을 제공합니다",
    "보안 우선: 환경 변수, API 키, 자격 증명을 자동으로 .gitignore에 추가하여 보호합니다",
    "컨텍스트 최적화: 200K 토큰 윈도우를 효율적으로 사용하여 대규모 프로젝트를 처리합니다",
    "빠른 피드백: 질문이 있으신가요? 즉시 물어보세요 - moai가 의도를 파악하고 명확히 합니다",
    "깔끔한 종료: /clear로 세션을 종료하여 컨텍스트를 재설정하고 다음 작업을 준비합니다"
]

# Hardcoded Japanese translations (23 items)
ANNOUNCEMENTS_JA = [
    "SPEC優先: /moai:1-planで要件を先に定義し、すべてが連携します",
    "TDD Cycle: RED (テスト作成) → GREEN (最小コード) → REFACTOR (改善)で品質を保証",
    "4段階ワークフロー: /moai:1-plan計画 → /moai:2-run実装 → /moai:3-sync検証 → gitデプロイ",
    "TRUST 5原則: Test(≥85%) + Readable + Unified + Secured(OWASP) + Trackable = 品質",
    "moaiエージェント: 19名の専門家チームが計画、実装、テスト、検証を自動処理",
    "対話スタイル: /output-style r2d2 (実習パートナー) または /output-style yoda (深い学習) 利用可能",
    "タスク追跡: TodoWriteで進捗を表示し、タスク漏れを防止",
    "55+ Skillライブラリ: 検証済みパターンとベストプラクティスで幻想のない実装を実現",
    "自動検証: コード作成後、TRUST 5、テストカバレッジ、セキュリティチェックが自動実行",
    "GitFlow戦略: feature/SPEC-XXX → develop → main構造で安全なデプロイパイプライン",
    "Context7更新: すべてのライブラリバージョンがリアルタイムAPI情報で最新を維持",
    "並列処理: 独立したタスク(テスト、ドキュメント、デプロイ)を同時実行し時間節約",
    "ヘルスチェック: moai-adk doctorで設定、バージョン、依存関係の問題を数秒で診断",
    "ドキュメント同期: /moai:3-syncでテスト、コード、ドキュメントを自動同期",
    "言語の分離: あなたの言語でコミュニケーション、プロジェクト言語でコード、英語でパッケージ",
    "多言語対応: 韓国語、英語、日本語、スペイン語など25言語以上で会話可能",
    "安全な更新: moai-adk updateで既存設定を保持しながら新機能を自動追加",
    "自動クリーンアップ: セッション終了時に.moai/temp/キャッシュとログを自動削除してスペース節約",
    "エラー復旧: 失敗したコミットとマージ衝突を自動分析し解決策を提供",
    "セキュリティ優先: 環境変数、APIキー、認証情報を自動的に.gitignoreに追加して保護",
    "コンテキスト最適化: 200Kトークンウィンドウを効率的に活用して大規模プロジェクト処理",
    "迅速なフィードバック: 質問がありますか？即座に聞いてください - moaiが意図を理解し明確にします",
    "きれいな終了: /clearでセッションを終了し、コンテキストをリセットして次のタスク準備"
]

# Hardcoded Chinese translations (23 items)
ANNOUNCEMENTS_ZH = [
    "SPEC优先: 使用 /moai:1-plan 先定义需求，一切都会相互连接",
    "TDD循环: RED (编写测试) → GREEN (最小代码) → REFACTOR (改进) 确保质量",
    "4阶段工作流: /moai:1-plan计划 → /moai:2-run实现 → /moai:3-sync验证 → git部署",
    "TRUST 5原则: Test(≥85%) + Readable + Unified + Secured(OWASP) + Trackable = 质量",
    "moai代理: 19名专家团队自动处理计划、实现、测试和验证",
    "对话风格: /output-style r2d2 (实践伙伴) 或 /output-style yoda (深度学习) 可用",
    "任务跟踪: TodoWrite持续显示进度，确保不遗漏任何任务",
    "55+ Skill库: 经过验证的模式和最佳实践实现无幻觉实现",
    "自动验证: 编码后，TRUST 5、测试覆盖率、安全检查自动运行",
    "GitFlow策略: feature/SPEC-XXX → develop → main结构保证安全部署",
    "Context7更新: 所有库版本通过实时API信息保持最新",
    "并行处理: 独立任务 (测试、文档、部署) 同时运行节省时间",
    "健康检查: moai-adk doctor在几秒内诊断配置、版本和依赖问题",
    "文档同步: /moai:3-sync auto自动同步测试、代码和文档",
    "语言分离: 用您的语言沟通，用项目语言编写代码，用英语打包",
    "多语言支持: 韩语、英语、日语、西班牙语等25种以上语言交流",
    "安全更新: moai-adk update在保留现有设置的同时自动添加新功能",
    "自动清理: 会话结束时自动清理 .moai/temp/ 缓存和日志节省空间",
    "错误恢复: 自动分析失败的提交和合并冲突并提供解决方案",
    "安全第一: 环境变量、API密钥、凭据自动添加到 .gitignore保护",
    "上下文优化: 有效利用200K令牌窗口处理大型项目",
    "快速反馈: 有问题吗？立即提问 - moai理解意图并澄清",
    "干净退出: 使用 /clear结束会话，重置上下文并为下一个任务做准备"
]

# Hardcoded translations dictionary (only en, ko, ja, zh supported)
HARDCODED_TRANSLATIONS = {
    "en": REFERENCE_ANNOUNCEMENTS_EN,
    "ko": ANNOUNCEMENTS_KO,
    "ja": ANNOUNCEMENTS_JA,
    "zh": ANNOUNCEMENTS_ZH
}


def get_language_from_config(project_root: Optional[Path] = None) -> str:
    """
    Retrieve conversation_language from .moai/config/config.json

    Args:
        project_root: Project root directory (defaults to current working directory)

    Returns:
        Language code (e.g., "ko", "en", "ja", "es")
    """
    if project_root is None:
        project_root = Path.cwd()

    # Try both possible paths: .moai/config/config.json and .moai/config.json
    config_paths = [
        project_root / ".moai" / "config" / "config.json",  # New structure
        project_root / ".moai" / "config.json",  # Legacy structure
    ]

    for config_path in config_paths:
        if config_path.exists():
            try:
                with open(config_path, "r", encoding="utf-8") as f:
                    config = json.load(f)
                return config.get("language", {}).get("conversation_language", "en")
            except Exception:
                pass

    return "en"  # Default to English if no config found


def copy_settings_to_local(
    language_code: str,
    announcements: List[str],
    project_root: Optional[Path] = None
) -> None:
    """
    Copy settings.json to settings.local.json with translated announcements

    Only supports hardcoded languages: en, ko, ja, zh
    Other languages default to English

    Args:
        language_code: Language code (en, ko, ja, zh, etc.)
        announcements: List of translated announcement strings
        project_root: Project root directory (defaults to current working directory)
    """
    if project_root is None:
        project_root = Path.cwd()

    settings_path = project_root / ".claude" / "settings.json"
    settings_local_path = project_root / ".claude" / "settings.local.json"
    claude_dir = settings_local_path.parent

    # Create .claude directory if needed
    claude_dir.mkdir(parents=True, exist_ok=True)

    try:
        # Load settings.json
        if settings_path.exists():
            with open(settings_path, "r", encoding="utf-8") as f:
                settings = json.load(f)
        else:
            # Create default settings if template doesn't exist
            settings = {
                "_meta": {
                    "description": "Claude Code settings (generated from template)",
                },
                "hooks": {},
                "permissions": {},
            }

        # Replace companyAnnouncements with translated version
        settings["companyAnnouncements"] = announcements

        # Write to settings.local.json
        with open(settings_local_path, "w", encoding="utf-8") as f:
            json.dump(settings, f, indent=2, ensure_ascii=False)

        print(f"[announcement_translator] Copied announcements to settings.local.json ({language_code})")

    except Exception as e:
        print(f"[announcement_translator] ERROR copying settings: {e}")


def translate_announcements(language_code: str, project_root: Optional[Path] = None) -> List[str]:
    """
    Get announcements in specified language from hardcoded translations

    Args:
        language_code: Target language code (e.g., "ko", "en", "ja", "zh")
        project_root: Project root directory (optional, not used in this simplified version)

    Returns:
        List of 23 announcement strings in the specified language.
        Returns English (REFERENCE_ANNOUNCEMENTS_EN) if language not found.
    """
    # Check if language has hardcoded translation
    if language_code in HARDCODED_TRANSLATIONS:
        return HARDCODED_TRANSLATIONS[language_code]

    # For unknown languages, default to English
    print(f"[announcement_translator] Language '{language_code}' not supported, using English")
    return REFERENCE_ANNOUNCEMENTS_EN


def auto_translate_and_update(project_root: Optional[Path] = None) -> None:
    """
    Auto-copy announcements to settings.local.json

    Workflow:
    1. Read language from .moai/config/config.json
    2. Get 23 announcements (hardcoded: en, ko, ja, zh)
    3. Copy to settings.local.json with language-specific announcements

    Supports: en (English), ko (Korean), ja (Japanese), zh (Chinese)
    Other languages default to English (23 announcements)

    This is the main function called by init and update commands.

    Args:
        project_root: Project root directory (defaults to current working directory)
    """
    if project_root is None:
        project_root = Path.cwd()

    # Step 1: Get language from config
    language = get_language_from_config(project_root)
    print(f"[announcement_translator] Detected language: {language}")

    # Step 2: Get announcements (hardcoded only, no dynamic translation)
    announcements = translate_announcements(language, project_root)

    # Step 3: Copy settings.json to settings.local.json
    copy_settings_to_local(language, announcements, project_root)


if __name__ == "__main__":
    """
    CLI entry point for direct execution:

    Usage:
        python announcement_translator.py [language_code] [project_root]

    If language_code is not provided, reads from .moai/config/config.json
    If project_root is not provided, uses current directory
    """
    import sys

    if len(sys.argv) > 1:
        # Manual language and project root override
        lang = sys.argv[1]
        project_root = Path(sys.argv[2]) if len(sys.argv) > 2 else Path.cwd()
        announcements = translate_announcements(lang, project_root)
        copy_settings_to_local(lang, announcements, project_root)
    else:
        # Auto-detect from config and update
        auto_translate_and_update()
