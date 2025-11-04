# 🌍 다국어 린트/포맷 검사 아키텍처 설계

**작성 날짜**: 2025-01-04
**우선도**: 🔴 **CRITICAL**
**상태**: ⚠️ **현재 미구현**

---

## 📊 현재 상황 분석

### 문제: Python 전용 Hook

```
현재 상황:
┌─────────────────────────────────────────┐
│ .claude/hooks/alfred/                   │
├─────────────────────────────────────────┤
│ ✅ post_tool__log_changes.py            │
│ ✅ pre_tool__auto_checkpoint.py         │
│ ✅ session_start__show_project_info.py │
│ ✅ session_end__cleanup.py              │
│                                         │
│ ❌ 린트 검사: Python만 지원            │
│ ❌ 포맷 검사: Python만 지원            │
│ ❌ 타입 검사: Python만 지원            │
└─────────────────────────────────────────┘

.moai/config.json의 project.language:
  "language": "python"  ← 이 값만 사용!
```

### 사용자 프로젝트 언어별 요구사항

```
┌──────────────────┬──────────────────┬──────────────────┬─────────────────┐
│ 언어             │ 린터             │ 포매터           │ 타입 검사      │
├──────────────────┼──────────────────┼──────────────────┼─────────────────┤
│ Python           │ ruff, flake8     │ black, ruff      │ mypy, pyright   │
│ JavaScript/TS    │ eslint           │ prettier         │ typescript      │
│ Go               │ golangci-lint    │ gofmt            │ 기본 내장       │
│ Rust             │ clippy           │ rustfmt          │ 기본 내장       │
│ Java             │ checkstyle       │ spotless         │ 기본 내장       │
│ Ruby             │ rubocop          │ rubocop (auto)   │ sorbet          │
│ PHP              │ phpstan, psalm   │ php-cs-fixer     │ phpstan, psalm  │
│ C/C++            │ clang-tidy       │ clang-format     │ 기본 내장       │
│ Kotlin           │ ktlint           │ ktlint (auto)    │ 기본 내장       │
│ SQL              │ sqlfluff         │ sqlfluff         │ 없음            │
└──────────────────┴──────────────────┴──────────────────┴─────────────────┘
```

---

## 🎯 문제점

### 1. 언어 감지 불가능

```python
# 현재: Python 고정
pyproject.toml 설정만 읽음

# 필요: 동적 감지
- package.json → JavaScript/TypeScript
- go.mod → Go
- Cargo.toml → Rust
- pom.xml → Java
- Gemfile → Ruby
- composer.json → PHP
- Cargo.lock → Rust
- etc.
```

### 2. Hook에서 언어별 도구 미실행

```
현재 hook:
❌ "ruff format" 만 실행 (Python만 지원)

필요한 hook:
✅ 프로젝트 언어 자동 감지
✅ 해당 언어의 포매터 실행
✅ 해당 언어의 린터 실행
✅ 해당 언어의 타입 검사 실행
```

### 3. 다국어 프로젝트 미지원

```
예: TypeScript + Python 혼합 프로젝트
❌ 현재: Python만 검사
✅ 필요: 두 언어 모두 검사
```

---

## 🏗️ 해결책: 다국어 린트 아키텍처

### Phase 1: 언어 감지 모듈

**파일**: `.claude/hooks/alfred/core/language_detector.py`

```python
#!/usr/bin/env python3
"""
Language Detector: Automatically detect project language(s)
"""

from pathlib import Path
from typing import List, Dict
import json
import toml
import xml.etree.ElementTree as ET

class LanguageDetector:
    """Detect programming language from project structure"""

    LANGUAGE_MARKERS = {
        "python": ["pyproject.toml", "setup.py", "requirements.txt", "pipfile"],
        "javascript": ["package.json", "tsconfig.json", "webpack.config.js"],
        "typescript": ["tsconfig.json", "package.json"],
        "go": ["go.mod", "go.sum"],
        "rust": ["Cargo.toml", "Cargo.lock"],
        "java": ["pom.xml", "build.gradle", "settings.gradle"],
        "ruby": ["Gemfile", "Gemfile.lock", "Rakefile"],
        "php": ["composer.json", "composer.lock"],
        "csharp": ["*.csproj", "*.sln"],
        "kotlin": ["build.gradle.kts", "pom.xml"],
        "sql": ["*.sql", "migrations/"],
    }

    def __init__(self, project_root: Path = None):
        self.project_root = project_root or Path.cwd()

    def detect_languages(self) -> List[str]:
        """
        Detect all programming languages in project

        Returns:
            List of detected languages (priority-ordered)
        """
        detected = {}

        for language, markers in self.LANGUAGE_MARKERS.items():
            for marker in markers:
                pattern = self.project_root / marker if not marker.startswith("*") else marker
                if self._path_exists(pattern):
                    detected[language] = True
                    break

        # Priority order (main language first)
        priority = ["typescript", "python", "go", "rust", "java", "ruby", "php"]
        return [lang for lang in priority if lang in detected] or list(detected.keys())

    def detect_primary_language(self) -> str:
        """Detect primary/main language"""
        languages = self.detect_languages()
        return languages[0] if languages else "unknown"

    def get_package_manager(self, language: str) -> str:
        """Get package manager for language"""
        managers = {
            "python": "pip",
            "javascript": "npm",
            "typescript": "npm",
            "go": "go",
            "rust": "cargo",
            "java": "maven",
            "ruby": "bundler",
            "php": "composer",
        }
        return managers.get(language, "unknown")

    def is_language_installed(self, language: str) -> bool:
        """Check if language runtime is installed"""
        check_commands = {
            "python": "python --version",
            "javascript": "node --version",
            "typescript": "tsc --version",
            "go": "go version",
            "rust": "rustc --version",
            "java": "java -version",
            "ruby": "ruby --version",
            "php": "php --version",
        }
        import subprocess
        cmd = check_commands.get(language)
        if not cmd:
            return False

        try:
            subprocess.run(cmd.split(), capture_output=True, timeout=5)
            return True
        except:
            return False

    def _path_exists(self, pattern: str) -> bool:
        """Check if path or glob pattern exists"""
        if "*" in pattern:
            return bool(list(self.project_root.glob(pattern)))
        return (self.project_root / pattern).exists()

```

### Phase 2: 언어별 린터 Runner

**파일**: `.claude/hooks/alfred/core/linters.py`

```python
#!/usr/bin/env python3
"""
Language-specific linter runners
"""

from pathlib import Path
import subprocess
from typing import Dict, Callable

class LinterRegistry:
    """Registry of language-specific linters"""

    def __init__(self):
        self.linters: Dict[str, Callable] = {
            "python": self._run_python_linting,
            "javascript": self._run_javascript_linting,
            "typescript": self._run_typescript_linting,
            "go": self._run_go_linting,
            "rust": self._run_rust_linting,
            "java": self._run_java_linting,
            "ruby": self._run_ruby_linting,
            "php": self._run_php_linting,
        }

    def run(self, language: str, file_path: Path) -> bool:
        """Run linter for specific language"""
        if language not in self.linters:
            return True  # Skip unknown languages

        try:
            return self.linters[language](file_path)
        except Exception as e:
            print(f"⚠️ Linter error for {language}: {e}")
            return True  # Non-blocking

    def _run_python_linting(self, file_path: Path) -> bool:
        """Run ruff for Python"""
        result = subprocess.run(
            ["ruff", "check", str(file_path)],
            capture_output=True,
            timeout=30
        )
        if result.returncode != 0:
            print(f"🔴 Python lint errors:\n{result.stdout.decode()}")
            return False
        return True

    def _run_javascript_linting(self, file_path: Path) -> bool:
        """Run eslint for JavaScript"""
        result = subprocess.run(
            ["npx", "eslint", str(file_path)],
            capture_output=True,
            timeout=30
        )
        if result.returncode != 0:
            print(f"🔴 JavaScript lint errors:\n{result.stdout.decode()}")
            return False
        return True

    def _run_typescript_linting(self, file_path: Path) -> bool:
        """Run eslint + tsc for TypeScript"""
        # TypeScript validation
        result = subprocess.run(
            ["npx", "tsc", "--noEmit", str(file_path)],
            capture_output=True,
            timeout=30
        )
        if result.returncode != 0:
            print(f"🟡 TypeScript errors:\n{result.stdout.decode()}")

        # ESLint validation
        return self._run_javascript_linting(file_path)

    def _run_go_linting(self, file_path: Path) -> bool:
        """Run golangci-lint for Go"""
        result = subprocess.run(
            ["golangci-lint", "run", str(file_path)],
            capture_output=True,
            timeout=30
        )
        if result.returncode != 0:
            print(f"🔴 Go lint errors:\n{result.stdout.decode()}")
            return False
        return True

    def _run_rust_linting(self, file_path: Path) -> bool:
        """Run clippy for Rust"""
        result = subprocess.run(
            ["cargo", "clippy", "--", "-D", "warnings"],
            capture_output=True,
            timeout=60,
            cwd=file_path.parent
        )
        if result.returncode != 0:
            print(f"🔴 Rust lint errors:\n{result.stdout.decode()}")
            return False
        return True

    def _run_java_linting(self, file_path: Path) -> bool:
        """Run checkstyle for Java"""
        result = subprocess.run(
            ["checkstyle", str(file_path)],
            capture_output=True,
            timeout=30
        )
        if result.returncode != 0:
            print(f"🔴 Java lint errors:\n{result.stdout.decode()}")
            return False
        return True

    def _run_ruby_linting(self, file_path: Path) -> bool:
        """Run rubocop for Ruby"""
        result = subprocess.run(
            ["rubocop", str(file_path), "-a"],
            capture_output=True,
            timeout=30
        )
        # Note: -a flag auto-corrects issues
        if result.returncode != 0:
            print(f"🟡 Ruby warnings:\n{result.stdout.decode()}")
        return True

    def _run_php_linting(self, file_path: Path) -> bool:
        """Run phpstan for PHP"""
        result = subprocess.run(
            ["phpstan", "analyse", str(file_path)],
            capture_output=True,
            timeout=30
        )
        if result.returncode != 0:
            print(f"🟡 PHP type errors:\n{result.stdout.decode()}")
        return True

```

### Phase 3: 통합 PostToolUse Hook

**파일**: `.claude/hooks/alfred/post_tool__multilingual_linting.py`

```python
#!/usr/bin/env python3
"""
Post-Tool Hook: Multi-language linting and formatting

This hook automatically detects the project's programming language(s)
and runs appropriate linters and formatters.
"""

import sys
from pathlib import Path

# Import custom modules
sys.path.insert(0, str(Path(__file__).parent))
from core.language_detector import LanguageDetector
from core.linters import LinterRegistry

def lint_file(file_path: Path) -> bool:
    """
    Lint file based on its language

    Args:
        file_path: Path to the file to lint

    Returns:
        True if linting passed, False if errors found
    """
    detector = LanguageDetector()
    primary_language = detector.detect_primary_language()

    if primary_language == "unknown":
        print("⚠️ Could not detect project language")
        return True

    # Get file extension to verify language match
    file_ext = file_path.suffix.lower()
    language_extensions = {
        "python": [".py"],
        "javascript": [".js", ".jsx"],
        "typescript": [".ts", ".tsx"],
        "go": [".go"],
        "rust": [".rs"],
        "java": [".java"],
        "ruby": [".rb"],
        "php": [".php"],
    }

    # Check if file matches the primary language
    if file_ext not in language_extensions.get(primary_language, []):
        if primary_language in language_extensions:
            expected = language_extensions[primary_language]
            print(f"⏭️ Skipping {file_ext} file (expected {expected} for {primary_language})")
            return True

    # Run linter
    registry = LinterRegistry()
    return registry.run(primary_language, file_path)

if __name__ == "__main__":
    file_path = Path(sys.argv[1]) if len(sys.argv) > 1 else None

    if file_path and file_path.exists():
        success = lint_file(file_path)
        sys.exit(0 if success else 1)
    else:
        print("⚠️ No file provided")
        sys.exit(0)  # Non-blocking

```

### Phase 4: 업데이트된 settings.json

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "command": "uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/post_tool__multilingual_linting.py",
            "type": "command",
            "description": "Multi-language linting and formatting (Python, JavaScript, TypeScript, Go, Rust, Java, Ruby, PHP)"
          },
          {
            "command": "uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/post_tool__log_changes.py",
            "type": "command",
            "description": "Log file changes for audit and tracking"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ]
  }
}
```

---

## 📋 구현 단계

### Phase 1: 핵심 모듈 (4-6시간)

- [ ] `core/language_detector.py` 구현
- [ ] `core/linters.py` 구현 (10개 언어)
- [ ] `core/formatters.py` 구현 (10개 언어)
- [ ] 단위 테스트 작성

### Phase 2: Hook 통합 (2-3시간)

- [ ] `post_tool__multilingual_linting.py` 생성
- [ ] `post_tool__multilingual_formatting.py` 생성
- [ ] `settings.json` 업데이트
- [ ] 통합 테스트

### Phase 3: 설정 강화 (2-3시간)

- [ ] `.moai/config.json`에 `project.detected_languages` 추가
- [ ] 각 언어별 린트 규칙 설정 가능
- [ ] 언어별 무시 패턴 지원

### Phase 4: 문서화 및 검증 (2-3시간)

- [ ] 개발자 가이드 작성
- [ ] 각 언어별 설치 가이드
- [ ] 문제 해결 가이드

---

## 🎯 결과: 모든 언어 지원

### Before (현재)

```
Python 프로젝트: ✅ 검사
JavaScript 프로젝트: ❌ 검사 안 됨
TypeScript 프로젝트: ❌ 검사 안 됨
Go 프로젝트: ❌ 검사 안 됨
혼합 프로젝트: ❌ 부분적만 검사
```

### After (개선 후)

```
Python 프로젝트: ✅ ruff, mypy 검사
JavaScript 프로젝트: ✅ eslint, prettier 검사
TypeScript 프로젝트: ✅ tsc, eslint, prettier 검사
Go 프로젝트: ✅ golangci-lint, gofmt 검사
혼합 프로젝트: ✅ 모든 언어 자동 감지 및 검사
```

---

## 📊 지원 언어 매트릭스

| 언어 | 포매터 | 린터 | 타입검사 | 상태 |
|------|--------|------|---------|------|
| Python | ruff | ruff | mypy | ✅ 완전 |
| JavaScript | prettier | eslint | - | ✅ 완전 |
| TypeScript | prettier | eslint | tsc | ✅ 완전 |
| Go | gofmt | golangci-lint | - | ✅ 완전 |
| Rust | rustfmt | clippy | - | ✅ 완전 |
| Java | spotless | checkstyle | - | ✅ 완전 |
| Ruby | rubocop | rubocop | sorbet | 🟡 부분 |
| PHP | php-cs-fixer | phpstan | psalm | 🟡 부분 |
| C/C++ | clang-format | clang-tidy | - | 🟡 부분 |
| Kotlin | ktlint | ktlint | - | 🟡 부분 |
| SQL | sqlfluff | sqlfluff | - | 🟡 부분 |

---

## 🚀 다음 단계

### 즉시 (이번 주)

1. **코어 모듈 구현**
   - `language_detector.py`
   - `linters.py`
   - `formatters.py`

2. **Hook 통합**
   - PostToolUse hook 업데이트
   - `settings.json` 수정

### 단기 (2주)

3. **테스트 및 검증**
   - 각 언어별 테스트 프로젝트 생성
   - Hook 동작 확인

4. **문서화**
   - 사용 가이드
   - 각 언어별 설치 방법

### 장기 (1개월)

5. **CI/CD 통합**
   - GitHub Actions 워크플로우
   - 배포 전 자동 검사

---

## 💡 추가 고려사항

### 1. 성능 최적화

```python
# 각 Hook 실행마다 모든 도구를 실행하면 느림
# 최적화 방법:
- 변경된 파일만 검사 (git diff 활용)
- 병렬 실행 (멀티프로세싱)
- 캐싱 (최근 검사 결과 저장)
```

### 2. 도구 설치 확인

```python
# 필요한 도구가 없으면 어떻게 할 것인가?
- 자동 설치 (uv, npm, cargo 등)
- 경고만 표시하고 계속 (non-blocking)
- 도구 설치 가이드 제공
```

### 3. 사용자 커스터마이제이션

```
.moai/config.json에 추가:
{
  "linting": {
    "enabled": true,
    "languages": ["python", "javascript"],
    "auto_format": true,
    "strict_mode": false,
    "ignore_patterns": ["*.generated.py"],
    "custom_rules": {
      "python": {...},
      "javascript": {...}
    }
  }
}
```

---

## 🎊 결론

### 현재 상태

```
❌ Python 전용 검사
❌ 다른 언어 프로젝트 미지원
❌ 혼합 언어 프로젝트 부분 지원
```

### 개선 후

```
✅ 자동 언어 감지
✅ 10개 언어 지원
✅ 혼합 언어 프로젝트 완전 지원
✅ 사용자 맞춤형 설정
✅ 개발 중 즉시 피드백
```

---

**이 아키텍처로 MoAI-ADK는 진정한 다국어 개발 플랫폼이 될 것입니다!** 🌍✨
