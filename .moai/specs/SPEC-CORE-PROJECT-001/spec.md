---
id: CORE-PROJECT-001
version: 0.1.0
status: completed
created: 2025-10-13
updated: 2025-10-14
author: @Goos
priority: high
category: feature
labels:
  - project
  - initialization
  - language-detection
depends_on:
  - PY314-001
  - CORE-TEMPLATE-001
scope:
  packages:
    - moai-adk-py/src/moai_adk/core/project/
  files:
    - project/initializer.py
    - project/detector.py
    - project/checker.py
---

# @SPEC:CORE-PROJECT-001: 프로젝트 초기화 및 언어 감지

## HISTORY

### v0.1.0 (2025-10-14)
- **IMPLEMENTED**: TDD 구현 완료 (RED-GREEN-REFACTOR)
- **AUTHOR**: @Goos
- **MODULES**: detector.py (92 LOC), languages.py (44 LOC), checker.py (59 LOC), initializer.py (102 LOC)
- **TESTS**: 79/79 passed, 100% coverage (75/75 statements)
- **QUALITY**: ruff ✓, mypy --strict ✓, TRUST 5원칙 준수
- **COMMITS**:
  - bb60d78 🔴 RED: 테스트 작성
  - 0d10504 🟢 GREEN: 구현 완료
  - c504618 ♻️ REFACTOR: 품질 개선

### v0.0.1 (2025-10-13)
- **INITIAL**: 프로젝트 초기화, 20개 언어 감지, 시스템 체커 명세 작성
- **AUTHOR**: @Goos
- **REASON**: moai init 명령어 핵심 로직 구현

---

## 개요

프로젝트 초기화 시 20개 언어를 자동 감지하고, .moai/ 디렉토리 구조를 생성한다. 시스템 요구사항을 검증하고 환경을 준비한다.

---

## Environment (환경 및 전제조건)

### 기술 스택
- **언어 감지**: 파일 확장자 및 설정 파일 분석
- **시스템 체크**: subprocess로 외부 명령 실행
- **디렉토리 생성**: pathlib.Path

### 지원 언어 (20개)
Python, TypeScript/JavaScript, Java, Go, Rust, Dart, Swift, Kotlin, C#, PHP, Ruby, Elixir, Scala, Clojure, Haskell, C/C++, Shell, HTML/CSS, SQL, YAML/JSON

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 프로젝트 언어를 자동 감지해야 한다
- 시스템은 .moai/ 디렉토리 구조를 생성해야 한다
- 시스템은 시스템 요구사항을 검증해야 한다
- 시스템은 20개 언어를 지원해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN `moai init .` 명령이 실행되면, 시스템은 언어를 감지해야 한다
- WHEN 언어가 감지되면, 시스템은 해당 템플릿을 적용해야 한다
- WHEN 디렉토리가 이미 존재하면, 시스템은 경고를 표시해야 한다

### State-driven Requirements (상태 기반)
- WHILE 시스템 요구사항 미충족 시, 시스템은 설치 가이드를 제공해야 한다
- WHILE 언어를 감지할 수 없을 때, 시스템은 수동 선택을 요청해야 한다

### Constraints (제약사항)
- .moai/ 디렉토리는 프로젝트 루트에만 생성되어야 한다
- config.json은 UTF-8 인코딩이어야 한다
- 언어 감지는 1초 이내여야 한다

---

## Specifications (상세 명세)

### 1. LanguageDetector 클래스

```python
# moai_adk/core/project/detector.py
from pathlib import Path
from typing import Optional

class LanguageDetector:
    LANGUAGE_PATTERNS = {
        "python": ["*.py", "pyproject.toml", "requirements.txt"],
        "typescript": ["*.ts", "tsconfig.json", "package.json"],
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
        # ... 더 많은 언어
    }

    def detect(self, path: str = ".") -> Optional[str]:
        """프로젝트 언어 감지"""
        project_path = Path(path)

        for language, patterns in self.LANGUAGE_PATTERNS.items():
            for pattern in patterns:
                if "*" in pattern:
                    # Glob pattern
                    if list(project_path.rglob(pattern)):
                        return language
                else:
                    # Exact file
                    if (project_path / pattern).exists():
                        return language

        return None

    def detect_multiple(self, path: str = ".") -> list[str]:
        """여러 언어 감지 (멀티 언어 프로젝트)"""
        detected = []
        project_path = Path(path)

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
```

### 2. ProjectInitializer 클래스

```python
# moai_adk/core/project/initializer.py
from pathlib import Path
from moai_adk.core.template import TemplateProcessor, ConfigManager
from moai_adk.core.project.detector import LanguageDetector

class ProjectInitializer:
    MOAI_STRUCTURE = [
        ".moai/config.json",
        ".moai/project/product.md",
        ".moai/project/structure.md",
        ".moai/project/tech.md",
        ".moai/specs/",
        ".moai/memory/",
        ".moai/backup/",
    ]

    def __init__(self, path: str = "."):
        self.path = Path(path)
        self.template_processor = TemplateProcessor()
        self.config_manager = ConfigManager(str(self.path / ".moai/config.json"))
        self.detector = LanguageDetector()

    def initialize(self, mode: str = "personal", locale: str = "ko") -> None:
        """프로젝트 초기화"""
        # 1. 언어 감지
        language = self.detector.detect(str(self.path))
        if not language:
            language = "generic"

        # 2. 디렉토리 생성
        self._create_directories()

        # 3. config.json 생성
        context = {
            "version": "0.3.0",
            "mode": mode,
            "locale": locale,
            "project_name": self.path.name,
            "language": language,
        }
        self.template_processor.render_to_file(
            ".moai/config.json.j2",
            str(self.path / ".moai/config.json"),
            context
        )

        # 4. 언어별 템플릿 생성
        self._generate_language_templates(language, context)

    def _create_directories(self) -> None:
        """디렉토리 구조 생성"""
        for item in self.MOAI_STRUCTURE:
            full_path = self.path / item
            if item.endswith("/"):
                full_path.mkdir(parents=True, exist_ok=True)
            else:
                full_path.parent.mkdir(parents=True, exist_ok=True)

    def _generate_language_templates(self, language: str, context: dict) -> None:
        """언어별 템플릿 생성"""
        # tech.md 생성
        template_path = f".moai/project/tech/{language}.md.j2"
        output_path = str(self.path / ".moai/project/tech.md")
        self.template_processor.render_to_file(template_path, output_path, context)
```

### 3. SystemChecker 클래스

```python
# moai_adk/core/project/checker.py
import subprocess
from typing import Dict

class SystemChecker:
    REQUIRED_TOOLS = {
        "git": "git --version",
        "python": "python3 --version",
    }

    OPTIONAL_TOOLS = {
        "gh": "gh --version",  # GitHub CLI
        "docker": "docker --version",
    }

    def check_all(self) -> Dict[str, bool]:
        """모든 시스템 요구사항 검증"""
        results = {}

        for tool, command in {**self.REQUIRED_TOOLS, **self.OPTIONAL_TOOLS}.items():
            results[tool] = self._check_tool(command)

        return results

    def _check_tool(self, command: str) -> bool:
        """개별 도구 확인"""
        try:
            subprocess.run(
                command.split(),
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            return False
```

---

## Traceability (추적성)

- **SPEC ID**: @SPEC:CORE-PROJECT-001
- **Depends on**: PY314-001, CORE-TEMPLATE-001
- **TAG 체인**: @SPEC:CORE-PROJECT-001 → @TEST:CORE-PROJECT-001 → @CODE:CORE-PROJECT-001
