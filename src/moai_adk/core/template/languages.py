"""Template mapping by language.

Defines template paths for 20 programming languages.
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
    """Return the template path for a language (case-insensitive).

    Args:
        language: Language name (case-insensitive).

    Returns:
        Template path; defaults to default.md.j2 for unknown languages.
    """
    if not language:
        return ".moai/project/tech/default.md.j2"

    language_lower = language.lower()
    return LANGUAGE_TEMPLATES.get(language_lower, ".moai/project/tech/default.md.j2")
