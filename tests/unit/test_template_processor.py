# @TEST:CORE-TEMPLATE-001 | SPEC: SPEC-CORE-TEMPLATE-001.md
"""TemplateProcessor 테스트 스위트

Jinja2 템플릿 렌더링, 파일 저장, autoescape 보안 기능을 검증합니다.
"""

from pathlib import Path

import pytest
from jinja2.exceptions import TemplateNotFound

from moai_adk.core.template.processor import TemplateProcessor


class TestTemplateProcessor:
    """TemplateProcessor 클래스 테스트"""

    def test_render_simple_template(self, tmp_path: Path) -> None:
        """간단한 템플릿 렌더링"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "greeting.txt.j2"
        template_file.write_text("Hello, {{ name }}!", encoding="utf-8")

        processor = TemplateProcessor(templates_dir)
        result = processor.render("greeting.txt.j2", {"name": "World"})

        assert result == "Hello, World!"

    def test_render_with_multiple_variables(self, tmp_path: Path) -> None:
        """여러 변수가 있는 템플릿 렌더링"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "config.txt.j2"
        template_file.write_text(
            "mode={{ mode }}, locale={{ locale }}", encoding="utf-8"
        )

        processor = TemplateProcessor(templates_dir)
        result = processor.render(
            "config.txt.j2", {"mode": "personal", "locale": "ko"}
        )

        assert result == "mode=personal, locale=ko"

    def test_render_raises_template_not_found(self, tmp_path: Path) -> None:
        """존재하지 않는 템플릿 렌더링 시 예외 발생"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()

        processor = TemplateProcessor(templates_dir)

        with pytest.raises(TemplateNotFound):
            processor.render("nonexistent.j2", {})

    def test_render_to_file_creates_output(self, tmp_path: Path) -> None:
        """render_to_file이 파일을 생성"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "test.txt.j2"
        template_file.write_text("Value: {{ value }}", encoding="utf-8")

        output_path = tmp_path / "output.txt"
        processor = TemplateProcessor(templates_dir)
        processor.render_to_file("test.txt.j2", output_path, {"value": 42})

        assert output_path.exists()
        assert output_path.read_text(encoding="utf-8") == "Value: 42"

    def test_render_to_file_creates_parent_directories(self, tmp_path: Path) -> None:
        """render_to_file이 부모 디렉토리를 생성"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "test.txt.j2"
        template_file.write_text("Content", encoding="utf-8")

        output_path = tmp_path / "nested" / "deep" / "output.txt"
        processor = TemplateProcessor(templates_dir)
        processor.render_to_file("test.txt.j2", output_path, {})

        assert output_path.exists()
        assert output_path.read_text(encoding="utf-8") == "Content"

    def test_render_preserves_korean_text(self, tmp_path: Path) -> None:
        """한글 텍스트를 올바르게 렌더링"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "korean.txt.j2"
        template_file.write_text("프로젝트: {{ project }}", encoding="utf-8")

        processor = TemplateProcessor(templates_dir)
        result = processor.render("korean.txt.j2", {"project": "테스트"})

        assert result == "프로젝트: 테스트"

    def test_trim_blocks_removes_newlines(self, tmp_path: Path) -> None:
        """trim_blocks가 블록 태그 뒤 줄바꿈 제거"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "trim.txt.j2"
        template_file.write_text(
            "{% if True %}\nContent\n{% endif %}", encoding="utf-8"
        )

        processor = TemplateProcessor(templates_dir)
        result = processor.render("trim.txt.j2", {})

        # trim_blocks=True이면 첫 줄바꿈이 제거됨
        assert result == "Content\n"

    def test_lstrip_blocks_removes_leading_spaces(self, tmp_path: Path) -> None:
        """lstrip_blocks가 블록 태그 앞 공백 제거"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "lstrip.txt.j2"
        template_file.write_text("    {% if True %}Content{% endif %}", encoding="utf-8")

        processor = TemplateProcessor(templates_dir)
        result = processor.render("lstrip.txt.j2", {})

        # lstrip_blocks=True이면 앞 공백이 제거됨
        assert result == "Content"

    def test_autoescape_escapes_html(self, tmp_path: Path) -> None:
        """autoescape가 HTML을 이스케이프"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "html.html.j2"
        template_file.write_text("{{ content }}", encoding="utf-8")

        processor = TemplateProcessor(templates_dir)
        result = processor.render("html.html.j2", {"content": "<script>alert('xss')</script>"})

        # autoescape가 활성화되면 HTML 태그가 이스케이프됨
        assert "&lt;script&gt;" in result
        assert "<script>" not in result

    def test_render_with_filters(self, tmp_path: Path) -> None:
        """Jinja2 필터 사용"""
        templates_dir = tmp_path / "templates"
        templates_dir.mkdir()
        template_file = templates_dir / "filters.txt.j2"
        template_file.write_text("{{ text | upper }}", encoding="utf-8")

        processor = TemplateProcessor(templates_dir)
        result = processor.render("filters.txt.j2", {"text": "hello"})

        assert result == "HELLO"

    def test_render_with_nested_directories(self, tmp_path: Path) -> None:
        """중첩 디렉토리의 템플릿 렌더링"""
        templates_dir = tmp_path / "templates"
        nested_dir = templates_dir / ".moai" / "project"
        nested_dir.mkdir(parents=True)
        template_file = nested_dir / "test.txt.j2"
        template_file.write_text("Nested: {{ value }}", encoding="utf-8")

        processor = TemplateProcessor(templates_dir)
        result = processor.render(".moai/project/test.txt.j2", {"value": "OK"})

        assert result == "Nested: OK"
