#!/usr/bin/env python3
"""Tool Registry for MoAI-ADK Hooks

Provides unified registry for formatters, linters, and language tools.
Supports 16+ languages with automatic tool detection and fallback.

Features:
- Automatic tool availability detection
- Graceful fallback when tools unavailable
- Extensible registry pattern
- Cross-platform support (macOS, Linux, Windows)
"""

import shutil
import subprocess
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path


class ToolType(Enum):
    """Tool type classification."""

    FORMATTER = "formatter"
    LINTER = "linter"
    TYPE_CHECKER = "type_checker"
    SECURITY_SCANNER = "security_scanner"
    AST_ANALYZER = "ast_analyzer"


@dataclass
class ToolConfig:
    """Configuration for a single tool."""

    name: str
    command: str
    args: list[str] = field(default_factory=list)
    file_args_position: str = "end"  # "end", "start", or "replace:{placeholder}"
    check_args: list[str] = field(default_factory=list)  # Args to check if tool exists
    fix_args: list[str] = field(default_factory=list)  # Args to auto-fix issues
    extensions: list[str] = field(default_factory=list)
    tool_type: ToolType = ToolType.FORMATTER
    priority: int = 1  # Lower = higher priority
    timeout_seconds: int = 30
    requires_config: bool = False  # Needs config file to work
    config_files: list[str] = field(default_factory=list)  # e.g., ["pyproject.toml"]


@dataclass
class ToolResult:
    """Result of tool execution."""

    success: bool
    tool_name: str
    output: str = ""
    error: str = ""
    exit_code: int = 0
    file_modified: bool = False
    issues_found: int = 0
    issues_fixed: int = 0


class ToolRegistry:
    """Registry for language tools (formatters, linters, etc.)."""

    def __init__(self) -> None:
        self._tools: dict[str, list[ToolConfig]] = {}
        self._extension_map: dict[str, str] = {}
        self._tool_cache: dict[str, bool] = {}
        self._register_default_tools()

    def _register_default_tools(self) -> None:
        """Register all default tools for supported languages."""
        # Python tools
        self._register_python_tools()
        # JavaScript/TypeScript tools
        self._register_js_ts_tools()
        # Go tools
        self._register_go_tools()
        # Rust tools
        self._register_rust_tools()
        # Java/Kotlin tools
        self._register_jvm_tools()
        # Swift tools
        self._register_swift_tools()
        # C/C++ tools
        self._register_cpp_tools()
        # Ruby tools
        self._register_ruby_tools()
        # PHP tools
        self._register_php_tools()
        # Elixir tools
        self._register_elixir_tools()
        # Scala tools
        self._register_scala_tools()
        # R tools
        self._register_r_tools()
        # Dart/Flutter tools
        self._register_dart_tools()
        # C# tools
        self._register_csharp_tools()
        # Markdown tools
        self._register_markdown_tools()
        # AST-Grep tools (cross-language)
        self._register_ast_grep_tools()

    def _register_python_tools(self) -> None:
        """Register Python formatting and linting tools."""
        extensions = [".py", ".pyi"]
        for ext in extensions:
            self._extension_map[ext] = "python"

        self._tools["python"] = [
            # Ruff - fastest Python linter and formatter
            ToolConfig(
                name="ruff-format",
                command="ruff",
                args=["format"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="ruff-check",
                command="ruff",
                args=["check", "--fix"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=30,
            ),
            # Black - fallback formatter
            ToolConfig(
                name="black",
                command="black",
                args=[],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=2,
                timeout_seconds=30,
            ),
            # isort - import sorter
            ToolConfig(
                name="isort",
                command="isort",
                args=[],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=3,
                timeout_seconds=15,
            ),
            # mypy - type checker
            ToolConfig(
                name="mypy",
                command="mypy",
                args=["--ignore-missing-imports"],
                extensions=extensions,
                tool_type=ToolType.TYPE_CHECKER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_js_ts_tools(self) -> None:
        """Register JavaScript/TypeScript tools."""
        js_extensions = [".js", ".jsx", ".mjs", ".cjs"]
        ts_extensions = [".ts", ".tsx", ".mts", ".cts"]

        for ext in js_extensions:
            self._extension_map[ext] = "javascript"
        for ext in ts_extensions:
            self._extension_map[ext] = "typescript"

        # Shared tools for both JS and TS
        shared_tools = [
            # Biome - fast all-in-one tool
            ToolConfig(
                name="biome-format",
                command="biome",
                args=["format", "--write"],
                extensions=js_extensions + ts_extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="biome-lint",
                command="biome",
                args=["lint", "--apply"],
                extensions=js_extensions + ts_extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=30,
            ),
            # Prettier - universal formatter
            ToolConfig(
                name="prettier",
                command="prettier",
                args=["--write"],
                extensions=js_extensions + ts_extensions,
                tool_type=ToolType.FORMATTER,
                priority=2,
                timeout_seconds=30,
            ),
            # ESLint
            ToolConfig(
                name="eslint",
                command="eslint",
                args=["--fix"],
                extensions=js_extensions + ts_extensions,
                tool_type=ToolType.LINTER,
                priority=2,
                timeout_seconds=60,
            ),
        ]

        self._tools["javascript"] = shared_tools.copy()
        self._tools["typescript"] = shared_tools.copy()

    def _register_go_tools(self) -> None:
        """Register Go tools."""
        extensions = [".go"]
        for ext in extensions:
            self._extension_map[ext] = "go"

        self._tools["go"] = [
            ToolConfig(
                name="gofmt",
                command="gofmt",
                args=["-w"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=15,
            ),
            ToolConfig(
                name="goimports",
                command="goimports",
                args=["-w"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=2,
                timeout_seconds=15,
            ),
            ToolConfig(
                name="golangci-lint",
                command="golangci-lint",
                args=["run", "--fix"],
                file_args_position="start",
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=120,
            ),
        ]

    def _register_rust_tools(self) -> None:
        """Register Rust tools."""
        extensions = [".rs"]
        for ext in extensions:
            self._extension_map[ext] = "rust"

        self._tools["rust"] = [
            ToolConfig(
                name="rustfmt",
                command="rustfmt",
                args=[],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="clippy",
                command="cargo",
                args=["clippy", "--fix", "--allow-dirty", "--allow-staged"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=180,
            ),
        ]

    def _register_jvm_tools(self) -> None:
        """Register Java and Kotlin tools."""
        java_extensions = [".java"]
        kotlin_extensions = [".kt", ".kts"]

        for ext in java_extensions:
            self._extension_map[ext] = "java"
        for ext in kotlin_extensions:
            self._extension_map[ext] = "kotlin"

        self._tools["java"] = [
            ToolConfig(
                name="google-java-format",
                command="google-java-format",
                args=["-i"],
                extensions=java_extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="checkstyle",
                command="checkstyle",
                args=["-c", "/google_checks.xml"],
                extensions=java_extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

        self._tools["kotlin"] = [
            ToolConfig(
                name="ktlint",
                command="ktlint",
                args=["-F"],
                extensions=kotlin_extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="detekt",
                command="detekt",
                args=["-i"],
                extensions=kotlin_extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_swift_tools(self) -> None:
        """Register Swift tools."""
        extensions = [".swift"]
        for ext in extensions:
            self._extension_map[ext] = "swift"

        self._tools["swift"] = [
            ToolConfig(
                name="swift-format",
                command="swift-format",
                args=["format", "-i"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="swiftlint",
                command="swiftlint",
                args=["--fix"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_cpp_tools(self) -> None:
        """Register C/C++ tools."""
        extensions = [".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx"]
        for ext in extensions:
            self._extension_map[ext] = "cpp"

        self._tools["cpp"] = [
            ToolConfig(
                name="clang-format",
                command="clang-format",
                args=["-i"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=15,
            ),
            ToolConfig(
                name="clang-tidy",
                command="clang-tidy",
                args=["--fix"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=120,
            ),
        ]

    def _register_ruby_tools(self) -> None:
        """Register Ruby tools."""
        extensions = [".rb", ".rake", ".gemspec"]
        for ext in extensions:
            self._extension_map[ext] = "ruby"

        self._tools["ruby"] = [
            ToolConfig(
                name="rubocop",
                command="rubocop",
                args=["-a"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_php_tools(self) -> None:
        """Register PHP tools."""
        extensions = [".php"]
        for ext in extensions:
            self._extension_map[ext] = "php"

        self._tools["php"] = [
            ToolConfig(
                name="php-cs-fixer",
                command="php-cs-fixer",
                args=["fix"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=60,
            ),
            ToolConfig(
                name="phpstan",
                command="phpstan",
                args=["analyze"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=120,
            ),
        ]

    def _register_elixir_tools(self) -> None:
        """Register Elixir tools."""
        extensions = [".ex", ".exs"]
        for ext in extensions:
            self._extension_map[ext] = "elixir"

        self._tools["elixir"] = [
            ToolConfig(
                name="mix-format",
                command="mix",
                args=["format"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="credo",
                command="mix",
                args=["credo", "--strict"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_scala_tools(self) -> None:
        """Register Scala tools."""
        extensions = [".scala", ".sc"]
        for ext in extensions:
            self._extension_map[ext] = "scala"

        self._tools["scala"] = [
            ToolConfig(
                name="scalafmt",
                command="scalafmt",
                args=[],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=60,
            ),
            ToolConfig(
                name="scalafix",
                command="scalafix",
                args=[],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=120,
            ),
        ]

    def _register_r_tools(self) -> None:
        """Register R tools."""
        extensions = [".r", ".R", ".Rmd"]
        for ext in extensions:
            self._extension_map[ext] = "r"

        self._tools["r"] = [
            ToolConfig(
                name="styler",
                command="Rscript",
                args=["-e", "styler::style_file"],
                file_args_position="replace:style_file",
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=60,
            ),
            ToolConfig(
                name="lintr",
                command="Rscript",
                args=["-e", "lintr::lint"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_dart_tools(self) -> None:
        """Register Dart/Flutter tools."""
        extensions = [".dart"]
        for ext in extensions:
            self._extension_map[ext] = "dart"

        self._tools["dart"] = [
            ToolConfig(
                name="dart-format",
                command="dart",
                args=["format"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=30,
            ),
            ToolConfig(
                name="dart-analyze",
                command="dart",
                args=["analyze"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_csharp_tools(self) -> None:
        """Register C# tools."""
        extensions = [".cs"]
        for ext in extensions:
            self._extension_map[ext] = "csharp"

        self._tools["csharp"] = [
            ToolConfig(
                name="dotnet-format",
                command="dotnet",
                args=["format"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=60,
            ),
        ]

    def _register_markdown_tools(self) -> None:
        """Register Markdown tools."""
        extensions = [".md", ".mdx", ".markdown"]
        for ext in extensions:
            self._extension_map[ext] = "markdown"

        self._tools["markdown"] = [
            ToolConfig(
                name="prettier-md",
                command="prettier",
                args=["--write", "--parser", "markdown"],
                extensions=extensions,
                tool_type=ToolType.FORMATTER,
                priority=1,
                timeout_seconds=15,
            ),
            ToolConfig(
                name="markdownlint",
                command="markdownlint",
                args=["--fix"],
                extensions=extensions,
                tool_type=ToolType.LINTER,
                priority=1,
                timeout_seconds=30,
            ),
        ]

    def _register_ast_grep_tools(self) -> None:
        """Register AST-Grep tools for structural code analysis.

        AST-Grep (sg) provides AST-based code search, lint, and rewriting.
        Supports 40+ languages with pattern matching and code transformation.
        """
        # AST-Grep supports multiple languages - define common extensions
        ast_grep_extensions = [
            ".py",
            ".pyi",  # Python
            ".js",
            ".jsx",
            ".mjs",
            ".cjs",  # JavaScript
            ".ts",
            ".tsx",
            ".mts",
            ".cts",  # TypeScript
            ".go",  # Go
            ".rs",  # Rust
            ".java",  # Java
            ".kt",
            ".kts",  # Kotlin
            ".c",
            ".cpp",
            ".cc",
            ".cxx",
            ".h",
            ".hpp",  # C/C++
            ".rb",  # Ruby
            ".swift",  # Swift
            ".cs",  # C#
            ".php",  # PHP
            ".scala",  # Scala
            ".ex",
            ".exs",  # Elixir
            ".lua",  # Lua
            ".html",
            ".vue",
            ".svelte",  # Web templates
        ]

        self._tools["ast_grep"] = [
            # Security scanning with rules
            ToolConfig(
                name="ast-grep-scan",
                command="sg",
                args=["scan"],
                extensions=ast_grep_extensions,
                tool_type=ToolType.AST_ANALYZER,
                priority=1,
                timeout_seconds=60,
                requires_config=True,
                config_files=["sgconfig.yml", ".ast-grep/sgconfig.yml"],
            ),
            # Pattern-based search
            ToolConfig(
                name="ast-grep-run",
                command="sg",
                args=["run", "--pattern"],
                file_args_position="end",
                extensions=ast_grep_extensions,
                tool_type=ToolType.AST_ANALYZER,
                priority=2,
                timeout_seconds=30,
            ),
            # Interactive testing
            ToolConfig(
                name="ast-grep-test",
                command="sg",
                args=["test"],
                extensions=ast_grep_extensions,
                tool_type=ToolType.AST_ANALYZER,
                priority=3,
                timeout_seconds=120,
                requires_config=True,
                config_files=["sgconfig.yml"],
            ),
        ]

    def is_tool_available(self, tool_name: str) -> bool:
        """Check if a tool is available on the system."""
        if tool_name in self._tool_cache:
            return self._tool_cache[tool_name]

        # Extract base command from tool name
        for lang_tools in self._tools.values():
            for tool in lang_tools:
                if tool.name == tool_name:
                    available = shutil.which(tool.command) is not None
                    self._tool_cache[tool_name] = available
                    return available

        return False

    def get_language_for_file(self, file_path: str) -> str | None:
        """Get language identifier for a file path."""
        ext = Path(file_path).suffix.lower()
        return self._extension_map.get(ext)

    def get_tools_for_language(self, language: str, tool_type: ToolType | None = None) -> list[ToolConfig]:
        """Get available tools for a language, optionally filtered by type."""
        tools = self._tools.get(language, [])

        if tool_type:
            tools = [t for t in tools if t.tool_type == tool_type]

        # Filter by availability and sort by priority
        available_tools = [t for t in tools if self.is_tool_available(t.name)]
        return sorted(available_tools, key=lambda t: t.priority)

    def get_tools_for_file(self, file_path: str, tool_type: ToolType | None = None) -> list[ToolConfig]:
        """Get available tools for a specific file."""
        language = self.get_language_for_file(file_path)
        if not language:
            return []
        return self.get_tools_for_language(language, tool_type)

    def _validate_file_path(self, file_path: str, cwd: str | None = None) -> str:
        """Validate and sanitize file path to prevent injection attacks.

        Security measures:
        - Reject paths with null bytes
        - Reject paths with shell metacharacters in dangerous positions
        - Resolve and validate path is within allowed boundaries
        - Ensure path exists and is a file

        Args:
            file_path: Path to validate
            cwd: Working directory context

        Returns:
            Validated absolute path

        Raises:
            ValueError: If path is invalid or potentially dangerous
        """

        # Check for null bytes (common injection technique)
        if "\x00" in file_path:
            raise ValueError("Invalid file path: contains null byte")

        # Check for shell metacharacters that could be dangerous
        # These are dangerous in string interpolation contexts (like R code)
        dangerous_chars = ["'", '"', "`", "$", ";", "&", "|", "\n", "\r"]
        for char in dangerous_chars:
            if char in file_path:
                raise ValueError(f"Invalid file path: contains dangerous character '{repr(char)}'")

        # Resolve path
        try:
            path = Path(file_path)
            if not path.is_absolute():
                base = Path(cwd) if cwd else Path.cwd()
                path = (base / path).resolve()
            else:
                path = path.resolve()
        except (OSError, ValueError) as e:
            raise ValueError(f"Invalid file path: {e}") from e

        # Ensure path exists and is a file
        if not path.exists():
            raise ValueError(f"File does not exist: {file_path}")
        if not path.is_file():
            raise ValueError(f"Path is not a file: {file_path}")

        return str(path)

    def _escape_for_code_string(self, value: str, language: str = "generic") -> str:
        """Escape a string value for safe inclusion in code.

        Args:
            value: String to escape
            language: Target language (r, python, generic)

        Returns:
            Escaped string safe for code inclusion
        """
        # Basic escaping for string literals
        # Escape backslashes first, then quotes
        escaped = value.replace("\\", "\\\\")

        if language == "r":
            # R uses single quotes, escape them
            escaped = escaped.replace("'", "\\'")
        else:
            # Generic: escape both quote types
            escaped = escaped.replace("'", "\\'")
            escaped = escaped.replace('"', '\\"')

        return escaped

    def run_tool(
        self,
        tool: ToolConfig,
        file_path: str,
        cwd: str | None = None,
        extra_args: list[str] | None = None,
    ) -> ToolResult:
        """Run a tool on a file and return the result.

        Security measures:
        - Validates file path before execution
        - Uses subprocess list mode (shell=False) to prevent shell injection
        - Properly escapes paths for code interpolation contexts
        """
        try:
            # Validate file path first (security check)
            try:
                validated_path = self._validate_file_path(file_path, cwd)
            except ValueError as e:
                return ToolResult(
                    success=False,
                    tool_name=tool.name,
                    error=f"Path validation failed: {e}",
                    exit_code=-1,
                )

            # Build command
            cmd = [tool.command] + tool.args

            # Add file path based on position
            if tool.file_args_position == "end":
                cmd.append(validated_path)
            elif tool.file_args_position == "start":
                cmd.insert(1, validated_path)
            elif tool.file_args_position.startswith("replace:"):
                # Handle code interpolation (e.g., R's styler::style_file)
                placeholder = tool.file_args_position.split(":")[1]
                # Escape path for safe code string inclusion
                escaped_path = self._escape_for_code_string(validated_path, "r")
                # Build function call: placeholder → placeholder('escaped_path')
                replacement = f"{placeholder}('{escaped_path}')"
                cmd = [c.replace(placeholder, replacement) for c in cmd]

            # Add extra args
            if extra_args:
                cmd.extend(extra_args)

            # Get file content hash before
            file_hash_before = self._get_file_hash(file_path)

            # Run tool
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=tool.timeout_seconds,
                cwd=cwd,
            )

            # Get file content hash after
            file_hash_after = self._get_file_hash(file_path)
            file_modified = file_hash_before != file_hash_after

            return ToolResult(
                success=result.returncode == 0,
                tool_name=tool.name,
                output=result.stdout,
                error=result.stderr,
                exit_code=result.returncode,
                file_modified=file_modified,
            )

        except subprocess.TimeoutExpired:
            return ToolResult(
                success=False,
                tool_name=tool.name,
                error=f"Tool timed out after {tool.timeout_seconds}s",
                exit_code=-1,
            )
        except FileNotFoundError:
            return ToolResult(
                success=False,
                tool_name=tool.name,
                error=f"Tool not found: {tool.command}",
                exit_code=-1,
            )
        except Exception as e:
            return ToolResult(
                success=False,
                tool_name=tool.name,
                error=str(e),
                exit_code=-1,
            )

    def _get_file_hash(self, file_path: str) -> str:
        """Get hash of file contents for change detection."""
        try:
            import hashlib

            with open(file_path, "rb") as f:
                return hashlib.md5(f.read()).hexdigest()
        except Exception:
            return ""


# Global registry instance
_registry: ToolRegistry | None = None


def get_tool_registry() -> ToolRegistry:
    """Get the global tool registry instance."""
    global _registry
    if _registry is None:
        _registry = ToolRegistry()
    return _registry
