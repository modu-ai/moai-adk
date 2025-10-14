# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_template_processor.py
"""Template Processor

Jinja2 기반 템플릿 렌더링:
- 템플릿 파일 로드
- 변수 치환 및 렌더링
- 파일 저장
- autoescape, trim_blocks, lstrip_blocks 설정
"""

from pathlib import Path
from typing import Any

from jinja2 import Environment, FileSystemLoader, select_autoescape


class TemplateProcessor:
    """Template Processor 클래스

    Jinja2를 사용하여 템플릿을 렌더링합니다.
    """

    def __init__(self, templates_dir: Path) -> None:
        """TemplateProcessor 초기화

        Args:
            templates_dir: 템플릿 파일이 있는 디렉토리
        """
        self.templates_dir = templates_dir
        self.env = Environment(
            loader=FileSystemLoader(str(templates_dir)),
            autoescape=select_autoescape(
                enabled_extensions=["html", "xml", "j2"],
                default_for_string=True,
            ),
            trim_blocks=True,
            lstrip_blocks=True,
        )

    def render(self, template_name: str, context: dict[str, Any]) -> str:
        """템플릿 렌더링

        Args:
            template_name: 템플릿 파일명 (상대 경로 포함 가능)
            context: 템플릿 변수 딕셔너리

        Returns:
            렌더링된 문자열

        Raises:
            TemplateNotFound: 템플릿 파일이 존재하지 않을 때
        """
        template = self.env.get_template(template_name)
        return template.render(context)

    def render_to_file(
        self,
        template_name: str,
        output_path: Path,
        context: dict[str, Any],
    ) -> None:
        """템플릿 렌더링 후 파일 저장

        부모 디렉토리가 없으면 자동으로 생성합니다.

        Args:
            template_name: 템플릿 파일명
            output_path: 출력 파일 경로
            context: 템플릿 변수 딕셔너리

        Raises:
            TemplateNotFound: 템플릿 파일이 존재하지 않을 때
        """
        rendered = self.render(template_name, context)

        # 부모 디렉토리 생성
        output_path.parent.mkdir(parents=True, exist_ok=True)

        # 파일 저장
        output_path.write_text(rendered, encoding="utf-8")
