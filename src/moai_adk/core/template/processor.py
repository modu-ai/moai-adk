# @CODE:CORE-TEMPLATE-001 | SPEC: SPEC-CORE-TEMPLATE-001.md | TEST: tests/unit/test_template_processor.py
"""TemplateProcessor - Jinja2 템플릿 프로세서

Jinja2 템플릿을 렌더링하고 파일로 저장하는 기능을 제공합니다.
"""

from pathlib import Path
from typing import Any

from jinja2 import Environment, FileSystemLoader, select_autoescape


class TemplateProcessor:
    """Jinja2 템플릿 프로세서

    템플릿을 렌더링하고 파일로 저장하는 기능을 제공합니다.
    autoescape, trim_blocks, lstrip_blocks를 기본으로 활성화하여
    안전하고 깔끔한 출력을 보장합니다.

    Attributes:
        templates_dir: 템플릿 디렉토리 경로
        env: Jinja2 Environment 인스턴스

    Examples:
        >>> processor = TemplateProcessor("templates")
        >>> result = processor.render("test.txt.j2", {"name": "World"})
        >>> print(result)
        Hello, World!
    """

    def __init__(self, templates_dir: str | Path = "templates") -> None:
        """TemplateProcessor 초기화

        Args:
            templates_dir: 템플릿 디렉토리 경로 (기본값: "templates")
        """
        self.templates_dir = Path(templates_dir)
        self.env = Environment(
            loader=FileSystemLoader(str(self.templates_dir)),
            autoescape=select_autoescape(
                enabled_extensions=("html", "htm", "xml", "html.j2", "htm.j2", "xml.j2"),
                default_for_string=True,
            ),  # XSS 방지
            trim_blocks=True,  # 블록 태그 뒤 첫 줄바꿈 제거
            lstrip_blocks=True,  # 블록 태그 앞 공백 제거
        )

    def render(self, template_path: str, context: dict[str, Any]) -> str:
        """템플릿 렌더링

        주어진 템플릿 파일을 컨텍스트 변수로 렌더링합니다.

        Args:
            template_path: 템플릿 파일 경로 (templates_dir 기준 상대 경로)
            context: 템플릿 변수 딕셔너리

        Returns:
            렌더링된 문자열

        Raises:
            TemplateNotFound: 템플릿 파일을 찾을 수 없을 때

        Examples:
            >>> processor = TemplateProcessor("templates")
            >>> result = processor.render("greeting.txt.j2", {"name": "Alice"})
            >>> print(result)
            Hello, Alice!
        """
        template = self.env.get_template(template_path)
        return template.render(**context)

    def render_to_file(
        self, template_path: str, output_path: str | Path, context: dict[str, Any]
    ) -> None:
        """템플릿을 파일로 렌더링

        템플릿을 렌더링하고 결과를 파일로 저장합니다.
        출력 디렉토리가 없으면 자동으로 생성합니다.

        Args:
            template_path: 템플릿 파일 경로
            output_path: 출력 파일 경로
            context: 템플릿 변수 딕셔너리

        Examples:
            >>> processor = TemplateProcessor("templates")
            >>> processor.render_to_file(
            ...     "config.json.j2",
            ...     ".moai/config.json",
            ...     {"version": "0.3.0"}
            ... )
        """
        content = self.render(template_path, context)
        output_path = Path(output_path)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(content, encoding="utf-8")
