"""System requirements validation module.

Checks whether required and optional tools are installed.
"""

import platform
import shutil
import subprocess
import sys
from pathlib import Path


class SystemChecker:
    """Validate system requirements."""

    REQUIRED_TOOLS: dict[str, str] = {
        "git": "git --version",
        "python": "python3 --version",
    }

    OPTIONAL_TOOLS: dict[str, str] = {
        "gh": "gh --version",
        "docker": "docker --version",
    }

    LANGUAGE_TOOLS: dict[str, dict[str, list[str]]] = {
        "python": {
            "required": ["python3", "pip"],
            "recommended": ["pytest", "mypy", "ruff"],
            "optional": ["black", "pylint"],
        },
        "typescript": {
            "required": ["node", "npm"],
            "recommended": ["vitest", "biome"],
            "optional": ["typescript", "eslint"],
        },
        "javascript": {
            "required": ["node", "npm"],
            "recommended": ["jest", "eslint"],
            "optional": ["prettier", "webpack"],
        },
        "java": {
            "required": ["java", "javac"],
            "recommended": ["maven", "gradle"],
            "optional": ["junit", "checkstyle"],
        },
        "go": {
            "required": ["go"],
            "recommended": ["golangci-lint", "gofmt"],
            "optional": ["delve", "gopls"],
        },
        "rust": {
            "required": ["rustc", "cargo"],
            "recommended": ["rustfmt", "clippy"],
            "optional": ["rust-analyzer", "cargo-audit"],
        },
        "dart": {
            "required": ["dart"],
            "recommended": ["flutter", "dart_test"],
            "optional": ["dartfmt", "dartanalyzer"],
        },
        "swift": {
            "required": ["swift", "swiftc"],
            "recommended": ["xcrun", "swift-format"],
            "optional": ["swiftlint", "sourcekit-lsp"],
        },
        "kotlin": {
            "required": ["kotlin", "kotlinc"],
            "recommended": ["gradle", "ktlint"],
            "optional": ["detekt", "kotlin-language-server"],
        },
        "csharp": {
            "required": ["dotnet"],
            "recommended": ["msbuild", "nuget"],
            "optional": ["csharpier", "roslyn"],
        },
        "php": {
            "required": ["php"],
            "recommended": ["composer", "phpunit"],
            "optional": ["psalm", "phpstan"],
        },
        "ruby": {
            "required": ["ruby", "gem"],
            "recommended": ["bundler", "rspec"],
            "optional": ["rubocop", "solargraph"],
        },
        "elixir": {
            "required": ["elixir", "mix"],
            "recommended": ["hex", "dialyzer"],
            "optional": ["credo", "ex_unit"],
        },
        "scala": {
            "required": ["scala", "scalac"],
            "recommended": ["sbt", "scalatest"],
            "optional": ["scalafmt", "metals"],
        },
        "clojure": {
            "required": ["clojure", "clj"],
            "recommended": ["leiningen", "clojure.test"],
            "optional": ["cider", "clj-kondo"],
        },
        "haskell": {
            "required": ["ghc", "ghci"],
            "recommended": ["cabal", "stack"],
            "optional": ["hlint", "haskell-language-server"],
        },
        "c": {
            "required": ["gcc", "make"],
            "recommended": ["clang", "cmake"],
            "optional": ["gdb", "valgrind"],
        },
        "cpp": {
            "required": ["g++", "make"],
            "recommended": ["clang++", "cmake"],
            "optional": ["gdb", "cppcheck"],
        },
        "lua": {
            "required": ["lua"],
            "recommended": ["luarocks", "busted"],
            "optional": ["luacheck", "lua-language-server"],
        },
        "ocaml": {
            "required": ["ocaml", "opam"],
            "recommended": ["dune", "ocamlformat"],
            "optional": ["merlin", "ocp-indent"],
        },
    }

    def check_all(self) -> dict[str, bool]:
        """Validate every tool.

        Returns:
            Dictionary mapping tool names to availability.
        """
        result = {}

        # Check required tools
        for tool, command in self.REQUIRED_TOOLS.items():
            result[tool] = self._check_tool(command)

        # Check optional tools
        for tool, command in self.OPTIONAL_TOOLS.items():
            result[tool] = self._check_tool(command)

        return result

    def _check_tool(self, command: str) -> bool:
        """Check an individual tool.

        Args:
            command: Command to run (e.g., "git --version").

        Returns:
            True when the tool is available.
        """
        if not command:
            return False

        try:
            # Extract the tool name (first token)
            tool_name = command.split()[0]
            # Determine availability via shutil.which
            return shutil.which(tool_name) is not None
        except Exception:
            return False

    def check_language_tools(self, language: str | None) -> dict[str, bool]:
        """Validate toolchains by language.

        Args:
            language: Programming language name (e.g., "python", "typescript").

        Returns:
            Dictionary mapping tool names to availability.
        """
        # Guard clause: no language specified
        if not language:
            return {}

        language_lower = language.lower()

        # Guard clause: unsupported language
        if language_lower not in self.LANGUAGE_TOOLS:
            return {}

        # Retrieve tool configuration for the language
        tools_config = self.LANGUAGE_TOOLS[language_lower]

        # Evaluate tools by category and collect results
        result: dict[str, bool] = {}
        for category in ["required", "recommended", "optional"]:
            tools = tools_config.get(category, [])
            for tool in tools:
                result[tool] = self._is_tool_available(tool)

        return result

    def _is_tool_available(self, tool: str) -> bool:
        """Check whether a tool is available (helper).

        Args:
            tool: Tool name.

        Returns:
            True when the tool is available.
        """
        return shutil.which(tool) is not None

    def get_tool_version(self, tool: str | None) -> str | None:
        """Retrieve tool version information.

        Args:
            tool: Tool name (for example, "python3", "node").

        Returns:
            Version string or None when the tool is unavailable.
        """
        # Guard clause: unspecified or unavailable tool
        if not tool or not self._is_tool_available(tool):
            return None

        try:
            # Call the tool with --version to obtain the version string
            result = subprocess.run(
                [tool, "--version"],
                capture_output=True,
                text=True,
                timeout=2,  # 2-second timeout to respect performance constraints
                check=False,
            )

            # Return the version when the command succeeds
            if result.returncode == 0 and result.stdout:
                return self._extract_version_line(result.stdout)

            return None

        except (subprocess.TimeoutExpired, OSError):
            # Gracefully handle timeout and OS errors
            return None

    def _extract_version_line(self, version_output: str) -> str:
        """Extract the first line from version output (helper).

        Args:
            version_output: Output captured from the --version command.

        Returns:
            First line containing version information.
        """
        return version_output.strip().split("\n")[0]


def check_environment() -> dict[str, bool]:
    """Validate the overall environment (used by the CLI doctor command).

    Returns:
        Mapping from check description to boolean status.
    """
    return {
        "Python >= 3.11": sys.version_info >= (3, 11),
        "Git installed": shutil.which("git") is not None,
        "Project structure (.moai/)": Path(".moai").exists(),
        "Config file (.moai/config/config.yaml)": Path(".moai/config/config.yaml").exists(),
    }


def get_platform_specific_message(unix_message: str, windows_message: str) -> str:
    """Return platform-specific message.

    Args:
        unix_message: Message for Unix/Linux/macOS.
        windows_message: Message for Windows.

    Returns:
        Platform-appropriate message.

    Examples:
        >>> get_platform_specific_message("chmod 755 .moai", "Check directory permissions")
        'chmod 755 .moai'  # on Unix/Linux/macOS
        >>> get_platform_specific_message("chmod 755 .moai", "Check directory permissions")
        'Check directory permissions'  # on Windows
    """
    if platform.system() == "Windows":
        return windows_message
    return unix_message


def get_permission_fix_message(path: str) -> str:
    """Get platform-specific permission fix message.

    Args:
        path: Path to fix permissions for.

    Returns:
        Platform-specific fix instructions.
    """
    if platform.system() == "Windows":
        return f"Run with administrator privileges or verify permissions in the properties of the '{path}' directory"
    return f"Run 'chmod 755 {path}' and try again"
