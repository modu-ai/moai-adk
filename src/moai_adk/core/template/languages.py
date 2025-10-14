# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_language_mapping.py
"""언어별 템플릿 매핑

20개 프로그래밍 언어의 템플릿 경로를 정의합니다.
"""

LANGUAGE_TEMPLATES: dict[str, str] = {
    "python": ".moai/project/tech/python.md.j2",
    "typescript": ".moai/project/tech/typescript.md.j2",
    "javascript": ".moai/project/tech/javascript.md.j2",
    "java": ".moai/project/tech/java.md.j2",
    "go": ".moai/project/tech/go.md.j2",
    "rust": ".moai/project/tech/rust.md.j2",
    "dart": ".moai/project/tech/dart.md.j2",
    "swift": ".moai/project/tech/swift.md.j2",
    "kotlin": ".moai/project/tech/kotlin.md.j2",
    "csharp": ".moai/project/tech/csharp.md.j2",
    "php": ".moai/project/tech/php.md.j2",
    "ruby": ".moai/project/tech/ruby.md.j2",
    "elixir": ".moai/project/tech/elixir.md.j2",
    "scala": ".moai/project/tech/scala.md.j2",
    "clojure": ".moai/project/tech/clojure.md.j2",
    "haskell": ".moai/project/tech/haskell.md.j2",
    "c": ".moai/project/tech/c.md.j2",
    "cpp": ".moai/project/tech/cpp.md.j2",
    "lua": ".moai/project/tech/lua.md.j2",
    "ocaml": ".moai/project/tech/ocaml.md.j2",
}


def get_language_template(language: str) -> str:
    """언어별 템플릿 경로 반환 (case-insensitive)

    Args:
        language: 언어 이름 (대소문자 무관)

    Returns:
        템플릿 파일 경로 (알 수 없는 언어는 default.md.j2)
    """
    if not language:
        return ".moai/project/tech/default.md.j2"

    language_lower = language.lower()
    return LANGUAGE_TEMPLATES.get(language_lower, ".moai/project/tech/default.md.j2")
