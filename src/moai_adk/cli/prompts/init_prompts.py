# @CODE:CLI-PROMPTS-001 | SPEC: SPEC-CLI-001.md
"""프로젝트 초기화 프롬프트

대화형 프로젝트 설정 수집
"""

from pathlib import Path
from typing import TypedDict

import questionary
from rich.console import Console

console = Console()


class ProjectSetupAnswers(TypedDict):
    """프로젝트 설정 답변"""

    project_name: str
    mode: str  # personal | team
    locale: str  # ko | en | ja | zh
    language: str | None
    author: str


def prompt_project_setup(
    project_name: str | None = None,
    is_current_dir: bool = False,
    project_path: Path | None = None,
) -> ProjectSetupAnswers:
    """프로젝트 설정 프롬프트

    Args:
        project_name: 프로젝트명 (None이면 질문)
        is_current_dir: 현재 디렉토리 모드 여부
        project_path: 프로젝트 경로 (경로 기반 이름 결정용)

    Returns:
        프로젝트 설정 답변
    """
    answers: ProjectSetupAnswers = {
        "project_name": "",
        "mode": "personal",
        "locale": "ko",
        "language": None,
        "author": "",
    }

    # 1. 프로젝트명 (현재 디렉토리가 아닐 때만)
    if not is_current_dir:
        if project_name:
            answers["project_name"] = project_name
            console.print(f"[cyan]📦 Project Name:[/cyan] {project_name}")
        else:
            answers["project_name"] = questionary.text(
                "📦 Project Name:",
                default="my-moai-project",
                validate=lambda text: len(text) > 0 or "Project name is required",
            ).ask()
    else:
        # 현재 디렉토리명 사용
        # 주의: Path.cwd()는 프로세스의 작업 디렉토리 (Claude Code의 cwd)
        # project_path가 있으면 우선 사용 (사용자 실행 위치)
        if project_path:
            answers["project_name"] = project_path.name
        else:
            answers["project_name"] = Path.cwd().name  # fallback
        console.print(
            f"[cyan]📦 Project Name:[/cyan] {answers['project_name']} [dim](current directory)[/dim]"
        )

    # 2. 프로젝트 모드
    answers["mode"] = questionary.select(
        "🔧 Project Mode:",
        choices=[
            questionary.Choice("Personal (single developer)", value="personal"),
            questionary.Choice("Team (collaborative)", value="team"),
        ],
        default="personal",
    ).ask()

    # 3. 로케일
    answers["locale"] = questionary.select(
        "🌐 Preferred Language:",
        choices=[
            questionary.Choice("한국어 (Korean)", value="ko"),
            questionary.Choice("English", value="en"),
            questionary.Choice("日本語 (Japanese)", value="ja"),
            questionary.Choice("中文 (Chinese)", value="zh"),
        ],
        default="ko",
    ).ask()

    # 4. 프로그래밍 언어 (자동 감지 또는 수동 선택)
    detect_language = questionary.confirm(
        "🔍 Auto-detect programming language?",
        default=True,
    ).ask()

    if not detect_language:
        answers["language"] = questionary.select(
            "💻 Select programming language:",
            choices=[
                "Python",
                "TypeScript",
                "JavaScript",
                "Java",
                "Go",
                "Rust",
                "Dart",
                "Swift",
                "Kotlin",
                "Generic",
            ],
        ).ask()

    # 5. 작성자 정보 (선택 사항)
    add_author = questionary.confirm(
        "👤 Add author information? (optional)",
        default=False,
    ).ask()

    if add_author:
        answers["author"] = questionary.text(
            "Author (GitHub ID):",
            default="",
            validate=lambda text: text.startswith("@") or "Must start with @",
        ).ask()

    return answers
