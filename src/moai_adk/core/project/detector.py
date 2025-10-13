# @CODE:CORE-PROJECT-001 | SPEC: SPEC-CORE-PROJECT-001.md | TEST: tests/unit/test_language_detector.py
"""프로젝트 언어 자동 감지 모듈"""

from pathlib import Path


class LanguageDetector:
    """프로젝트 언어 자동 감지

    20개 주요 프로그래밍 언어를 파일 패턴으로 감지합니다.
    """

    LANGUAGE_PATTERNS: dict[str, list[str]] = {
        "python": ["pyproject.toml", "setup.py", "requirements.txt", "*.py"],
        "typescript": ["tsconfig.json", "*.ts", "*.tsx"],
        "javascript": ["package.json", "*.js", "*.jsx"],
        "java": ["pom.xml", "build.gradle", "*.java"],
        "go": ["go.mod", "go.sum", "*.go"],
        "rust": ["Cargo.toml", "Cargo.lock", "*.rs"],
        "dart": ["pubspec.yaml", "*.dart"],
        "swift": ["Package.swift", "*.swift"],
        "kotlin": ["build.gradle.kts", "*.kt", "*.kts"],
        "csharp": ["*.csproj", "*.sln", "*.cs"],
        "php": ["composer.json", "*.php"],
        "ruby": ["Gemfile", "Gemfile.lock", "*.rb"],
        "elixir": ["mix.exs", "*.ex", "*.exs"],
        "scala": ["build.sbt", "*.scala"],
        "clojure": ["project.clj", "deps.edn", "*.clj"],
        "haskell": ["stack.yaml", "*.cabal", "*.hs"],
        "cpp": ["CMakeLists.txt", "*.cpp", "*.hpp"],
        "c": ["Makefile", "*.c", "*.h"],
        "shell": ["*.sh", "*.bash"],
        "lua": ["*.lua"],
    }

    def detect(self, path: str | Path = ".") -> str | None:
        """프로젝트 주 언어 감지

        Args:
            path: 프로젝트 디렉토리 경로

        Returns:
            감지된 언어명 (없으면 None)
        """
        project_path = Path(path)

        for language, patterns in self.LANGUAGE_PATTERNS.items():
            for pattern in patterns:
                if "*" in pattern:
                    # Glob 패턴 (재귀 탐색)
                    if list(project_path.rglob(pattern)):
                        return language
                else:
                    # 정확한 파일명
                    if (project_path / pattern).exists():
                        return language

        return None

    def detect_multiple(self, path: str | Path = ".") -> list[str]:
        """여러 언어 감지 (멀티 언어 프로젝트)

        Args:
            path: 프로젝트 디렉토리 경로

        Returns:
            감지된 언어 목록
        """
        project_path = Path(path)
        detected = []

        for language, patterns in self.LANGUAGE_PATTERNS.items():
            for pattern in patterns:
                if "*" in pattern:
                    if list(project_path.rglob(pattern)):
                        detected.append(language)
                        break
                else:
                    if (project_path / pattern).exists():
                        detected.append(language)
                        break

        return detected
