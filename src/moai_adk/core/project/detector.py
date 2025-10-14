# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_language_detector.py
"""언어 감지기 모듈

20개 프로그래밍 언어를 자동으로 감지합니다.
"""

from pathlib import Path


class LanguageDetector:
    """20개 프로그래밍 언어 자동 감지"""

    LANGUAGE_PATTERNS = {
        "python": ["*.py", "pyproject.toml", "requirements.txt", "setup.py"],
        "typescript": ["*.ts", "tsconfig.json"],
        "javascript": ["*.js", "package.json"],
        "java": ["*.java", "pom.xml", "build.gradle"],
        "go": ["*.go", "go.mod"],
        "rust": ["*.rs", "Cargo.toml"],
        "dart": ["*.dart", "pubspec.yaml"],
        "swift": ["*.swift", "Package.swift"],
        "kotlin": ["*.kt", "build.gradle.kts"],
        "csharp": ["*.cs", "*.csproj"],
        "php": ["*.php", "composer.json"],
        "ruby": ["*.rb", "Gemfile"],
        "elixir": ["*.ex", "mix.exs"],
        "scala": ["*.scala", "build.sbt"],
        "clojure": ["*.clj", "project.clj"],
        "haskell": ["*.hs", "*.cabal"],
        "c": ["*.c", "Makefile"],
        "cpp": ["*.cpp", "CMakeLists.txt"],
        "shell": ["*.sh", "*.bash"],
        "lua": ["*.lua"],
    }

    def detect(self, path: str | Path = ".") -> str | None:
        """단일 언어 감지 (우선순위 순)

        Args:
            path: 검사할 디렉토리 경로

        Returns:
            감지된 언어 이름 (소문자) 또는 None
        """
        path = Path(path)

        # 각 언어에 대해 우선순위 순서로 검사
        for language, patterns in self.LANGUAGE_PATTERNS.items():
            if self._check_patterns(path, patterns):
                return language

        return None

    def detect_multiple(self, path: str | Path = ".") -> list[str]:
        """멀티 언어 감지

        Args:
            path: 검사할 디렉토리 경로

        Returns:
            감지된 모든 언어 이름 리스트
        """
        path = Path(path)
        detected = []

        for language, patterns in self.LANGUAGE_PATTERNS.items():
            if self._check_patterns(path, patterns):
                detected.append(language)

        return detected

    def _check_patterns(self, path: Path, patterns: list[str]) -> bool:
        """패턴 리스트 중 하나라도 매칭되는지 확인

        Args:
            path: 검사할 디렉토리
            patterns: glob 패턴 리스트

        Returns:
            하나 이상의 패턴이 매칭되면 True
        """
        for pattern in patterns:
            # 확장자 패턴 (예: *.py)
            if pattern.startswith("*."):
                if list(path.rglob(pattern)):
                    return True
            # 특정 파일명 (예: pyproject.toml)
            else:
                if (path / pattern).exists():
                    return True

        return False
