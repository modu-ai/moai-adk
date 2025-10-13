# @CODE:CORE-TEMPLATE-001 | SPEC: SPEC-CORE-TEMPLATE-001.md | TEST: tests/unit/test_language_mapping.py
"""언어별 템플릿 매핑

20개 주요 프로그래밍 언어의 템플릿 경로를 매핑합니다.
"""

LANGUAGE_TEMPLATES: dict[str, str] = {
    # 인터프리터 언어
    "python": ".moai/project/tech/python.md.j2",
    "javascript": ".moai/project/tech/javascript.md.j2",
    "typescript": ".moai/project/tech/typescript.md.j2",
    "ruby": ".moai/project/tech/ruby.md.j2",
    "php": ".moai/project/tech/php.md.j2",
    "lua": ".moai/project/tech/lua.md.j2",
    # 컴파일 언어
    "java": ".moai/project/tech/java.md.j2",
    "go": ".moai/project/tech/go.md.j2",
    "rust": ".moai/project/tech/rust.md.j2",
    "cpp": ".moai/project/tech/cpp.md.j2",
    "c": ".moai/project/tech/c.md.j2",
    "csharp": ".moai/project/tech/csharp.md.j2",
    # 모바일 언어
    "dart": ".moai/project/tech/dart.md.j2",
    "swift": ".moai/project/tech/swift.md.j2",
    "kotlin": ".moai/project/tech/kotlin.md.j2",
    # JVM 언어
    "scala": ".moai/project/tech/scala.md.j2",
    "clojure": ".moai/project/tech/clojure.md.j2",
    # 함수형 언어
    "haskell": ".moai/project/tech/haskell.md.j2",
    "elixir": ".moai/project/tech/elixir.md.j2",
    "ocaml": ".moai/project/tech/ocaml.md.j2",
}


def get_language_template(language: str) -> str:
    """언어별 템플릿 경로 반환

    주어진 언어에 해당하는 템플릿 경로를 반환합니다.
    언어를 찾을 수 없으면 기본 템플릿을 반환합니다.
    대소문자를 구분하지 않습니다.

    Args:
        language: 프로그래밍 언어명 (대소문자 무관)

    Returns:
        템플릿 파일 경로 (.moai/project/tech/로 시작하는 상대 경로)
        언어를 찾을 수 없으면 기본 템플릿 경로 반환

    Examples:
        >>> get_language_template("python")
        '.moai/project/tech/python.md.j2'
        >>> get_language_template("Python")
        '.moai/project/tech/python.md.j2'
        >>> get_language_template("UNKNOWN")
        '.moai/project/tech/default.md.j2'
    """
    return LANGUAGE_TEMPLATES.get(language.lower(), ".moai/project/tech/default.md.j2")
