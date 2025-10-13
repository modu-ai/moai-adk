# @TEST:CORE-TEMPLATE-001 | SPEC: SPEC-CORE-TEMPLATE-001.md
"""언어별 템플릿 매핑 테스트 스위트

20개 언어 템플릿 경로 매핑 및 get_language_template 함수를 검증합니다.
"""

import pytest

from moai_adk.core.template.languages import LANGUAGE_TEMPLATES, get_language_template


class TestLanguageMapping:
    """언어별 템플릿 매핑 테스트"""

    def test_language_templates_has_20_languages(self) -> None:
        """LANGUAGE_TEMPLATES에 20개 언어가 정의됨"""
        assert len(LANGUAGE_TEMPLATES) == 20

    def test_all_template_paths_have_correct_format(self) -> None:
        """모든 템플릿 경로가 올바른 형식"""
        for language, path in LANGUAGE_TEMPLATES.items():
            assert path.startswith(".moai/project/tech/")
            assert path.endswith(".md.j2")

    def test_get_language_template_returns_correct_path(self) -> None:
        """get_language_template이 올바른 경로 반환"""
        assert get_language_template("python") == ".moai/project/tech/python.md.j2"
        assert get_language_template("typescript") == ".moai/project/tech/typescript.md.j2"
        assert get_language_template("java") == ".moai/project/tech/java.md.j2"

    def test_get_language_template_case_insensitive(self) -> None:
        """get_language_template이 대소문자를 구분하지 않음"""
        assert get_language_template("Python") == get_language_template("python")
        assert get_language_template("PYTHON") == get_language_template("python")
        assert get_language_template("TypeScript") == get_language_template("typescript")

    def test_get_language_template_returns_default_for_unknown(self) -> None:
        """get_language_template이 알 수 없는 언어에 기본 템플릿 반환"""
        result = get_language_template("unknown_language")
        assert result == ".moai/project/tech/default.md.j2"

    def test_all_major_languages_supported(self) -> None:
        """주요 프로그래밍 언어가 모두 지원됨"""
        major_languages = [
            "python",
            "javascript",
            "typescript",
            "java",
            "go",
            "rust",
            "cpp",
            "c",
            "csharp",
        ]
        for lang in major_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_mobile_languages_supported(self) -> None:
        """모바일 언어가 지원됨"""
        mobile_languages = ["dart", "swift", "kotlin"]
        for lang in mobile_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_functional_languages_supported(self) -> None:
        """함수형 언어가 지원됨"""
        functional_languages = ["haskell", "elixir", "scala", "clojure"]
        for lang in functional_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_scripting_languages_supported(self) -> None:
        """스크립팅 언어가 지원됨"""
        scripting_languages = ["python", "ruby", "php", "lua"]
        for lang in scripting_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_interpreter_languages_count(self) -> None:
        """인터프리터 언어 수 확인"""
        interpreter_languages = ["python", "javascript", "typescript", "ruby", "php", "lua"]
        for lang in interpreter_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_compiled_languages_count(self) -> None:
        """컴파일 언어 수 확인"""
        compiled_languages = ["java", "go", "rust", "cpp", "c", "csharp"]
        for lang in compiled_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_jvm_languages_supported(self) -> None:
        """JVM 언어가 지원됨"""
        jvm_languages = ["java", "scala", "clojure", "kotlin"]
        for lang in jvm_languages:
            assert lang in LANGUAGE_TEMPLATES

    def test_language_template_uniqueness(self) -> None:
        """각 언어가 고유한 템플릿 경로를 가짐"""
        paths = list(LANGUAGE_TEMPLATES.values())
        assert len(paths) == len(set(paths))

    def test_get_language_template_with_empty_string(self) -> None:
        """빈 문자열에 대해 기본 템플릿 반환"""
        result = get_language_template("")
        assert result == ".moai/project/tech/default.md.j2"

    def test_specific_language_paths(self) -> None:
        """특정 언어들의 템플릿 경로가 올바른지 검증"""
        expected_mappings = {
            "python": ".moai/project/tech/python.md.j2",
            "rust": ".moai/project/tech/rust.md.j2",
            "dart": ".moai/project/tech/dart.md.j2",
            "swift": ".moai/project/tech/swift.md.j2",
            "go": ".moai/project/tech/go.md.j2",
        }
        for language, expected_path in expected_mappings.items():
            assert get_language_template(language) == expected_path

    def test_ocaml_supported(self) -> None:
        """OCaml이 지원됨 (20번째 언어)"""
        assert "ocaml" in LANGUAGE_TEMPLATES
        assert get_language_template("ocaml") == ".moai/project/tech/ocaml.md.j2"
