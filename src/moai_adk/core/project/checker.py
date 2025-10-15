# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md | TEST: tests/unit/test_cli_commands.py
"""시스템 요구사항 검증 모듈

필수 및 선택 도구의 설치 여부를 확인합니다.
"""

import shutil
import subprocess
import sys
from pathlib import Path


class SystemChecker:
    """시스템 요구사항 검증"""

    REQUIRED_TOOLS: dict[str, str] = {
        "git": "git --version",
        "python": "python3 --version",
    }

    OPTIONAL_TOOLS: dict[str, str] = {
        "gh": "gh --version",
        "docker": "docker --version",
    }

    # @CODE:CLI-001:DATA - 언어별 도구 체인 매핑 (20개 언어 지원)
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
        """모든 도구 검증

        Returns:
            도구명: 사용가능 여부 딕셔너리
        """
        result = {}

        # 필수 도구 확인
        for tool, command in self.REQUIRED_TOOLS.items():
            result[tool] = self._check_tool(command)

        # 선택 도구 확인
        for tool, command in self.OPTIONAL_TOOLS.items():
            result[tool] = self._check_tool(command)

        return result

    def _check_tool(self, command: str) -> bool:
        """개별 도구 확인

        Args:
            command: 확인할 명령어 (예: "git --version")

        Returns:
            도구가 사용 가능하면 True
        """
        if not command:
            return False

        try:
            # 명령어에서 도구 이름 추출 (첫 단어)
            tool_name = command.split()[0]
            # shutil.which로 도구 존재 확인
            return shutil.which(tool_name) is not None
        except Exception:
            return False

    def check_language_tools(self, language: str | None) -> dict[str, bool]:
        """언어별 도구 체인 검증

        Args:
            language: 프로그래밍 언어 이름 (예: "python", "typescript")

        Returns:
            도구명: 사용가능 여부 딕셔너리
        """
        # 가드절: 언어 미지정 시 빈 딕셔너리 반환
        if not language:
            return {}

        language_lower = language.lower()

        # 가드절: 지원하지 않는 언어인 경우 빈 딕셔너리 반환
        if language_lower not in self.LANGUAGE_TOOLS:
            return {}

        # 언어별 도구 설정 가져오기
        tools_config = self.LANGUAGE_TOOLS[language_lower]

        # 도구 카테고리별로 검증하여 결과 수집
        result: dict[str, bool] = {}
        for category in ["required", "recommended", "optional"]:
            tools = tools_config.get(category, [])
            for tool in tools:
                result[tool] = self._is_tool_available(tool)

        return result

    def _is_tool_available(self, tool: str) -> bool:
        """도구 사용 가능 여부 확인 (헬퍼 메서드)

        Args:
            tool: 도구 이름

        Returns:
            도구 사용 가능 여부
        """
        return shutil.which(tool) is not None

    def get_tool_version(self, tool: str | None) -> str | None:
        """도구 버전 정보 추출

        Args:
            tool: 도구 이름 (예: "python3", "node")

        Returns:
            버전 문자열 또는 None (도구 미설치 시)
        """
        # 가드절: 도구 미지정 또는 미설치
        if not tool or not self._is_tool_available(tool):
            return None

        try:
            # --version 옵션으로 버전 정보 추출
            result = subprocess.run(
                [tool, "--version"],
                capture_output=True,
                text=True,
                timeout=2,  # 2초 타임아웃 (성능 제약 준수)
                check=False,
            )

            # 성공적으로 버전 정보를 가져온 경우
            if result.returncode == 0 and result.stdout:
                return self._extract_version_line(result.stdout)

            return None

        except (subprocess.TimeoutExpired, OSError):
            # 타임아웃 또는 OS 에러 시 None 반환
            return None

    def _extract_version_line(self, version_output: str) -> str:
        """버전 출력에서 첫 번째 줄 추출 (헬퍼 메서드)

        Args:
            version_output: --version 명령어 출력 결과

        Returns:
            첫 번째 줄 버전 정보
        """
        return version_output.strip().split("\n")[0]


def check_environment() -> dict[str, bool]:
    """전체 환경 검증 (CLI doctor 명령어용)

    Returns:
        각 체크 항목의 결과 딕셔너리
    """
    return {
        "Python >= 3.13": sys.version_info >= (3, 13),
        "Git installed": shutil.which("git") is not None,
        "Project structure (.moai/)": Path(".moai").exists(),
        "Config file (.moai/config.json)": Path(".moai/config.json").exists(),
    }
