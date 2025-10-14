# Core API Reference

> MoAI-ADK Python v0.3.0 Core Module API Documentation

MoAI-ADK의 핵심 모듈로, Git 관리, 프로젝트 초기화, 템플릿 처리를 담당합니다.

---

## 모듈 구조

```
moai_adk.core/
├── git/              # Git 저장소 관리
│   ├── manager.py    # GitManager 클래스
│   ├── branch.py     # 브랜치 네이밍 유틸리티
│   └── commit.py     # 커밋 메시지 포맷팅
├── project/          # 프로젝트 초기화 및 감지
│   ├── initializer.py # ProjectInitializer 클래스
│   ├── detector.py    # LanguageDetector 클래스
│   └── checker.py     # SystemChecker 클래스
└── template/         # 템플릿 매핑
    └── languages.py   # 언어별 템플릿 경로 매핑
```

---

## moai_adk.core.git

GitPython 기반 Git 저장소 관리 모듈입니다.

### GitManager

::: moai_adk.core.git.manager.GitManager
    options:
      show_source: true
      heading_level: 4

Git 저장소를 관리하는 핵심 클래스입니다. 브랜치 생성, 커밋, 푸시 등 Git 작업을 Python API로 제공합니다.

#### 주요 기능

- **저장소 검증**: `is_repo()` - Git 저장소 여부 확인
- **브랜치 관리**: `current_branch()`, `create_branch()` - 브랜치 조회 및 생성
- **변경사항 확인**: `is_dirty()` - 작업 디렉토리 상태 확인
- **커밋 및 푸시**: `commit()`, `push()` - 변경사항 커밋 및 원격 동기화

#### 사용 예시

```python
from moai_adk.core.git import GitManager

# GitManager 초기화
manager = GitManager("/path/to/repo")

# Git 저장소 확인
if manager.is_repo():
    print(f"현재 브랜치: {manager.current_branch()}")

# 새 브랜치 생성 및 전환
manager.create_branch("feature/SPEC-AUTH-001", from_branch="develop")

# 변경사항 커밋
manager.commit(
    message="🟢 GREEN: JWT 인증 구현",
    files=["src/auth/jwt.py", "tests/test_auth.py"]
)

# 원격 저장소에 푸시 (upstream 설정)
manager.push(set_upstream=True)

# 작업 디렉토리 상태 확인
if manager.is_dirty():
    print("커밋되지 않은 변경사항이 있습니다.")
```

#### 메서드 상세

##### `__init__(repo_path: str = ".")`

GitManager를 초기화합니다.

**Parameters:**
- `repo_path` (str): Git 저장소 경로. 기본값은 현재 디렉토리 (`"."`)

**Raises:**
- `InvalidGitRepositoryError`: 지정된 경로가 Git 저장소가 아닐 경우

**Example:**

```python
# 현재 디렉토리를 Git 저장소로 초기화
manager = GitManager()

# 특정 경로 지정
manager = GitManager("/path/to/repo")
```

---

##### `is_repo() -> bool`

현재 경로가 Git 저장소인지 확인합니다.

**Returns:**
- `bool`: Git 저장소이면 `True`, 아니면 `False`

**Example:**

```python
manager = GitManager()
if manager.is_repo():
    print("Git 저장소입니다.")
else:
    print("Git 저장소가 아닙니다.")
```

---

##### `current_branch() -> str`

현재 활성 브랜치명을 반환합니다.

**Returns:**
- `str`: 현재 브랜치명 (예: `"main"`, `"develop"`, `"feature/SPEC-AUTH-001"`)

**Example:**

```python
manager = GitManager()
branch = manager.current_branch()
print(f"현재 브랜치: {branch}")
# 출력: 현재 브랜치: main
```

---

##### `is_dirty() -> bool`

작업 디렉토리에 커밋되지 않은 변경사항이 있는지 확인합니다.

**Returns:**
- `bool`: 변경사항이 있으면 `True` (dirty), 없으면 `False` (clean)

**Example:**

```python
manager = GitManager()
if manager.is_dirty():
    print("커밋되지 않은 변경사항이 있습니다.")
else:
    print("작업 디렉토리가 깨끗합니다.")
```

---

##### `create_branch(branch_name: str, from_branch: str | None = None) -> None`

새 브랜치를 생성하고 해당 브랜치로 전환합니다.

**Parameters:**
- `branch_name` (str): 생성할 브랜치명
- `from_branch` (str | None): 기준 브랜치. `None`이면 현재 브랜치 기준

**Example:**

```python
manager = GitManager()

# 현재 브랜치에서 분기
manager.create_branch("feature/SPEC-AUTH-001")

# develop 브랜치에서 분기
manager.create_branch("feature/SPEC-AUTH-002", from_branch="develop")

# 생성된 브랜치 확인
print(manager.current_branch())
# 출력: feature/SPEC-AUTH-002
```

---

##### `commit(message: str, files: list[str] | None = None) -> None`

파일을 스테이징하고 커밋합니다.

**Parameters:**
- `message` (str): 커밋 메시지
- `files` (list[str] | None): 커밋할 파일 목록. `None`이면 모든 변경사항 커밋

**Example:**

```python
manager = GitManager()

# 특정 파일만 커밋
manager.commit(
    message="🔴 RED: JWT 인증 테스트 작성",
    files=["tests/test_auth.py"]
)

# 모든 변경사항 커밋
manager.commit(message="🟢 GREEN: JWT 인증 구현")
```

---

##### `push(branch: str | None = None, set_upstream: bool = False) -> None`

원격 저장소에 푸시합니다.

**Parameters:**
- `branch` (str | None): 푸시할 브랜치명. `None`이면 현재 브랜치
- `set_upstream` (bool): upstream 설정 여부. 첫 푸시 시 `True` 권장

**Example:**

```python
manager = GitManager()

# 현재 브랜치를 origin에 푸시 (upstream 설정)
manager.push(set_upstream=True)

# 특정 브랜치 푸시
manager.push(branch="feature/SPEC-AUTH-001")

# 일반 푸시 (upstream 이미 설정된 경우)
manager.push()
```

---

### Branch Utilities

::: moai_adk.core.git.branch.generate_branch_name
    options:
      show_source: true
      heading_level: 4

SPEC ID로부터 표준 브랜치명을 생성하는 유틸리티 함수입니다.

#### 브랜치 네이밍 규칙

MoAI-ADK는 SPEC-First TDD 방법론을 따르며, 모든 브랜치는 SPEC ID와 연결됩니다:

- **형식**: `feature/SPEC-{DOMAIN}-{NUMBER}`
- **예시**: `feature/SPEC-AUTH-001`, `feature/SPEC-CORE-GIT-001`

#### 사용 예시

```python
from moai_adk.core.git.branch import generate_branch_name
from moai_adk.core.git import GitManager

# SPEC ID로부터 브랜치명 생성
spec_id = "AUTH-001"
branch_name = generate_branch_name(spec_id)
print(branch_name)
# 출력: feature/SPEC-AUTH-001

# GitManager와 함께 사용
manager = GitManager()
manager.create_branch(generate_branch_name("CORE-GIT-001"))
print(manager.current_branch())
# 출력: feature/SPEC-CORE-GIT-001
```

#### 복합 도메인 지원

```python
# 복합 도메인 (하이픈으로 연결)
generate_branch_name("UPDATE-REFACTOR-001")
# 출력: feature/SPEC-UPDATE-REFACTOR-001

generate_branch_name("INSTALLER-SEC-001")
# 출력: feature/SPEC-INSTALLER-SEC-001
```

---

### Commit Message Formatting

::: moai_adk.core.git.commit.format_commit_message
    options:
      show_source: true
      heading_level: 4

TDD 단계별 표준 커밋 메시지를 생성하는 유틸리티 함수입니다.

#### 커밋 메시지 형식

MoAI-ADK는 TDD 단계를 명확히 표시하는 이모지 기반 커밋 메시지를 사용합니다:

| 단계 | 이모지 | 설명 | 예시 |
|------|--------|------|------|
| `red` | 🔴 | 실패하는 테스트 작성 | `🔴 RED: JWT 인증 테스트 작성` |
| `green` | 🟢 | 테스트 통과하는 최소 구현 | `🟢 GREEN: JWT 인증 구현` |
| `refactor` | ♻️ | 코드 품질 개선 | `♻️ REFACTOR: 인증 로직 모듈화` |
| `docs` | 📝 | 문서화 | `📝 DOCS: JWT 인증 API 문서 작성` |

#### 다국어 지원

4개 언어를 지원하며, `.moai/config.json`의 `locale` 설정에 따라 자동 선택됩니다:

- `ko`: 한국어 (기본값)
- `en`: English
- `ja`: 日本語
- `zh`: 中文

#### 사용 예시

```python
from moai_adk.core.git.commit import format_commit_message
from moai_adk.core.git import GitManager

# RED 단계 (한국어)
message = format_commit_message("red", "JWT 인증 테스트 작성", locale="ko")
print(message)
# 출력: 🔴 RED: JWT 인증 테스트 작성

# GREEN 단계 (영어)
message = format_commit_message("green", "Implement JWT authentication", locale="en")
print(message)
# 출력: 🟢 GREEN: Implement JWT authentication

# REFACTOR 단계 (일본어)
message = format_commit_message("refactor", "認証ロジックのモジュール化", locale="ja")
print(message)
# 출력: ♻️ REFACTOR: 認証ロジックのモジュール化

# DOCS 단계 (중국어)
message = format_commit_message("docs", "编写JWT认证API文档", locale="zh")
print(message)
# 출력: 📝 DOCS: 编写JWT认证API文档
```

#### GitManager와 통합 사용

```python
from moai_adk.core.git import GitManager
from moai_adk.core.git.commit import format_commit_message
from moai_adk.core.git.branch import generate_branch_name

# 전체 TDD 워크플로우
manager = GitManager()

# 1. SPEC 브랜치 생성
manager.create_branch(generate_branch_name("AUTH-001"), from_branch="develop")

# 2. RED: 테스트 작성
manager.commit(
    message=format_commit_message("red", "JWT 인증 테스트 작성"),
    files=["tests/test_auth.py"]
)

# 3. GREEN: 구현
manager.commit(
    message=format_commit_message("green", "JWT 인증 구현"),
    files=["src/auth/jwt.py", "tests/test_auth.py"]
)

# 4. REFACTOR: 리팩토링
manager.commit(
    message=format_commit_message("refactor", "인증 로직 모듈화"),
    files=["src/auth/jwt.py", "src/auth/utils.py"]
)

# 5. 원격 푸시
manager.push(set_upstream=True)
```

#### 에러 처리

```python
from moai_adk.core.git.commit import format_commit_message

# 잘못된 stage 입력
try:
    message = format_commit_message("invalid", "Test", locale="ko")
except ValueError as e:
    print(f"에러: {e}")
    # 출력: 에러: Invalid stage: invalid

# 지원하지 않는 locale은 영어로 폴백
message = format_commit_message("red", "Test", locale="unknown")
print(message)
# 출력: 🔴 RED: Test
```

---

## moai_adk.core.project

프로젝트 초기화, 언어 감지, 시스템 요구사항 검증을 담당하는 모듈입니다.

### ProjectInitializer

::: moai_adk.core.project.initializer.ProjectInitializer
    options:
      show_source: true
      heading_level: 4

`.moai/` 디렉토리 구조를 생성하고 프로젝트를 초기화하는 클래스입니다.

#### `.moai/` 디렉토리 구조

ProjectInitializer는 다음 구조를 자동으로 생성합니다:

```
.moai/
├── config.json              # 프로젝트 설정 (필수)
├── project/
│   ├── product.md          # 제품 개요
│   ├── structure.md        # 디렉토리 구조
│   └── tech.md             # 기술 스택
├── specs/                  # SPEC 문서
├── memory/                 # 개발 가이드 및 메모리
└── backup/                 # 백업 파일
```

#### 사용 예시

```python
from moai_adk.core.project import ProjectInitializer

# 현재 디렉토리 초기화
initializer = ProjectInitializer()

# 초기화 여부 확인
if not initializer.is_initialized():
    result = initializer.initialize(
        mode="team",         # personal | team
        locale="ko",         # ko | en | ja | zh
        language=None        # None이면 자동 감지
    )
    print(result)
    # 출력:
    # {
    #     'path': '/path/to/project',
    #     'language': 'python',
    #     'mode': 'team',
    #     'locale': 'ko'
    # }
else:
    print("이미 초기화된 프로젝트입니다.")
```

#### 특정 경로 초기화

```python
from pathlib import Path
from moai_adk.core.project import ProjectInitializer

# 특정 경로 초기화
project_path = Path("/path/to/new/project")
initializer = ProjectInitializer(project_path)

# 프로젝트 초기화 (언어 강제 지정)
result = initializer.initialize(
    mode="personal",
    locale="en",
    language="typescript"  # 자동 감지 대신 강제 지정
)

print(f"프로젝트 '{result['path']}' 초기화 완료")
print(f"감지된 언어: {result['language']}")
print(f"모드: {result['mode']}")
```

#### config.json 형식

생성되는 `config.json` 파일 구조:

```json
{
  "projectName": "my-project",
  "mode": "team",
  "locale": "ko",
  "language": "python"
}
```

---

### LanguageDetector

::: moai_adk.core.project.detector.LanguageDetector
    options:
      show_source: true
      heading_level: 4

20개 프로그래밍 언어를 자동으로 감지하는 클래스입니다.

#### 지원 언어 (20개)

| 언어 | 감지 패턴 | 우선순위 |
|------|----------|----------|
| **Python** | `*.py`, `pyproject.toml`, `requirements.txt` | 1 |
| **TypeScript** | `*.ts`, `tsconfig.json` | 2 |
| **JavaScript** | `*.js`, `package.json` | 3 |
| **Java** | `*.java`, `pom.xml`, `build.gradle` | 4 |
| **Go** | `*.go`, `go.mod` | 5 |
| **Rust** | `*.rs`, `Cargo.toml` | 6 |
| **Dart** | `*.dart`, `pubspec.yaml` | 7 |
| **Swift** | `*.swift`, `Package.swift` | 8 |
| **Kotlin** | `*.kt`, `build.gradle.kts` | 9 |
| **C#** | `*.cs`, `*.csproj` | 10 |
| **PHP** | `*.php`, `composer.json` | 11 |
| **Ruby** | `*.rb`, `Gemfile` | 12 |
| **Elixir** | `*.ex`, `mix.exs` | 13 |
| **Scala** | `*.scala`, `build.sbt` | 14 |
| **Clojure** | `*.clj`, `project.clj` | 15 |
| **Haskell** | `*.hs`, `*.cabal` | 16 |
| **C** | `*.c`, `Makefile` | 17 |
| **C++** | `*.cpp`, `CMakeLists.txt` | 18 |
| **Shell** | `*.sh`, `*.bash` | 19 |
| **Lua** | `*.lua` | 20 |

#### 사용 예시: 단일 언어 감지

```python
from moai_adk.core.project.detector import LanguageDetector

detector = LanguageDetector()

# 현재 디렉토리 감지
language = detector.detect()
if language:
    print(f"감지된 언어: {language}")
else:
    print("언어를 감지할 수 없습니다.")

# 특정 경로 감지
language = detector.detect("/path/to/project")
print(f"감지된 언어: {language}")
# 출력: 감지된 언어: python
```

#### 사용 예시: 멀티 언어 감지

```python
from moai_adk.core.project.detector import LanguageDetector

detector = LanguageDetector()

# 모든 언어 감지 (멀티 언어 프로젝트)
languages = detector.detect_multiple()
print(f"감지된 언어들: {languages}")
# 출력: 감지된 언어들: ['python', 'typescript', 'javascript']

# React Native 프로젝트 예시
languages = detector.detect_multiple("/path/to/react-native-app")
# 출력: ['javascript', 'typescript', 'java', 'swift']
```

#### 사용 예시: ProjectInitializer와 통합

```python
from moai_adk.core.project import ProjectInitializer, LanguageDetector

detector = LanguageDetector()
initializer = ProjectInitializer()

# 1. 먼저 언어 감지
detected = detector.detect()
print(f"감지된 주언어: {detected}")

# 2. 멀티 언어 프로젝트 확인
all_languages = detector.detect_multiple()
if len(all_languages) > 1:
    print(f"경고: 멀티 언어 프로젝트입니다. 감지된 언어: {all_languages}")
    print(f"주언어로 '{detected}'를 사용합니다.")

# 3. 초기화 (감지된 언어 사용)
result = initializer.initialize(language=detected)
print(f"프로젝트 초기화 완료: {result}")
```

#### 감지 로직 상세

```python
from pathlib import Path
from moai_adk.core.project.detector import LanguageDetector

detector = LanguageDetector()

# Python 프로젝트 감지 조건
# 1. *.py 파일이 존재하거나
# 2. pyproject.toml이 존재하거나
# 3. requirements.txt가 존재하면 Python으로 감지

project_path = Path("/path/to/project")

# 개별 패턴 확인 (내부 메서드 참조)
python_patterns = ["*.py", "pyproject.toml", "requirements.txt"]

for pattern in python_patterns:
    if pattern.startswith("*."):
        # 확장자 패턴
        files = list(project_path.rglob(pattern))
        if files:
            print(f"발견: {pattern} -> {len(files)}개 파일")
    else:
        # 특정 파일명
        if (project_path / pattern).exists():
            print(f"발견: {pattern}")
```

---

### SystemChecker

::: moai_adk.core.project.checker.SystemChecker
    options:
      show_source: true
      heading_level: 4

시스템 요구사항(필수/선택 도구)을 검증하는 클래스입니다.

#### 검증 도구 목록

**필수 도구 (Required)**:
- `git`: Git 버전 관리 시스템
- `python`: Python 3.9 이상

**선택 도구 (Optional)**:
- `gh`: GitHub CLI (PR 자동화)
- `docker`: Docker (컨테이너 환경)

#### 사용 예시: 전체 검증

```python
from moai_adk.core.project.checker import SystemChecker

checker = SystemChecker()

# 모든 도구 검증
result = checker.check_all()

print("=== 시스템 요구사항 검증 ===")
print(f"Git: {'✅' if result['git'] else '❌'}")
print(f"Python: {'✅' if result['python'] else '❌'}")
print(f"GitHub CLI: {'✅' if result['gh'] else '⚠️ (선택)'}")
print(f"Docker: {'✅' if result['docker'] else '⚠️ (선택)'}")

# 필수 도구 확인
required_ok = result['git'] and result['python']
if not required_ok:
    print("\n❌ 필수 도구가 설치되지 않았습니다.")
    exit(1)
```

#### 사용 예시: CLI 명령어 통합

```python
from moai_adk.core.project.checker import SystemChecker

def cmd_doctor():
    """moai doctor 명령어 구현"""
    checker = SystemChecker()
    result = checker.check_all()

    # 필수 도구
    print("필수 도구:")
    for tool in ["git", "python"]:
        status = "✅ 설치됨" if result[tool] else "❌ 미설치"
        print(f"  - {tool}: {status}")

    # 선택 도구
    print("\n선택 도구:")
    for tool in ["gh", "docker"]:
        status = "✅ 설치됨" if result[tool] else "⚠️ 미설치 (선택)"
        print(f"  - {tool}: {status}")

    # 전체 상태
    required_ok = result['git'] and result['python']
    if required_ok:
        print("\n✅ 시스템 준비 완료")
        return 0
    else:
        print("\n❌ 필수 도구를 설치해주세요:")
        if not result['git']:
            print("  - Git: https://git-scm.com/downloads")
        if not result['python']:
            print("  - Python 3.9+: https://www.python.org/downloads/")
        return 1
```

#### 사용 예시: 조건부 기능 활성화

```python
from moai_adk.core.project.checker import SystemChecker

checker = SystemChecker()
result = checker.check_all()

# GitHub CLI 사용 가능 여부에 따라 기능 분기
if result['gh']:
    print("GitHub CLI를 사용한 자동 PR 생성이 가능합니다.")
    # PR 자동화 로직
else:
    print("GitHub CLI가 없습니다. 수동으로 PR을 생성하세요.")
    print("설치: brew install gh")
    # 수동 PR 안내

# Docker 사용 가능 여부 확인
if result['docker']:
    print("Docker 컨테이너 환경에서 테스트를 실행합니다.")
    # Docker 테스트 실행
else:
    print("로컬 환경에서 테스트를 실행합니다.")
    # 로컬 테스트 실행
```

#### 개별 도구 검증

```python
from moai_adk.core.project.checker import SystemChecker

checker = SystemChecker()

# 개별 도구 확인 (내부 메서드 참조)
# _check_tool() 메서드는 private이므로 check_all() 사용 권장

# 특정 도구만 확인하고 싶은 경우
result = checker.check_all()
if not result['git']:
    print("Git이 설치되지 않았습니다.")
    print("설치 방법:")
    print("  - macOS: brew install git")
    print("  - Ubuntu: sudo apt install git")
    print("  - Windows: https://git-scm.com/download/win")
```

---

## moai_adk.core.template

템플릿 경로 매핑 및 언어별 템플릿 관리 모듈입니다.

### Language Template Mapping

::: moai_adk.core.template.languages.get_language_template
    options:
      show_source: true
      heading_level: 4

언어별 기술 스택 템플릿 경로를 반환하는 유틸리티 함수입니다.

#### 템플릿 매핑 테이블

| 언어 | 템플릿 경로 |
|------|------------|
| Python | `.moai/project/tech/python.md.j2` |
| TypeScript | `.moai/project/tech/typescript.md.j2` |
| JavaScript | `.moai/project/tech/javascript.md.j2` |
| Java | `.moai/project/tech/java.md.j2` |
| Go | `.moai/project/tech/go.md.j2` |
| Rust | `.moai/project/tech/rust.md.j2` |
| Dart | `.moai/project/tech/dart.md.j2` |
| Swift | `.moai/project/tech/swift.md.j2` |
| Kotlin | `.moai/project/tech/kotlin.md.j2` |
| C# | `.moai/project/tech/csharp.md.j2` |
| PHP | `.moai/project/tech/php.md.j2` |
| Ruby | `.moai/project/tech/ruby.md.j2` |
| Elixir | `.moai/project/tech/elixir.md.j2` |
| Scala | `.moai/project/tech/scala.md.j2` |
| Clojure | `.moai/project/tech/clojure.md.j2` |
| Haskell | `.moai/project/tech/haskell.md.j2` |
| C | `.moai/project/tech/c.md.j2` |
| C++ | `.moai/project/tech/cpp.md.j2` |
| Lua | `.moai/project/tech/lua.md.j2` |
| OCaml | `.moai/project/tech/ocaml.md.j2` |
| **기타** | `.moai/project/tech/default.md.j2` |

#### 사용 예시

```python
from moai_adk.core.template.languages import get_language_template

# Python 프로젝트
template_path = get_language_template("python")
print(template_path)
# 출력: .moai/project/tech/python.md.j2

# TypeScript 프로젝트
template_path = get_language_template("typescript")
print(template_path)
# 출력: .moai/project/tech/typescript.md.j2

# 대소문자 무관
template_path = get_language_template("PYTHON")
print(template_path)
# 출력: .moai/project/tech/python.md.j2

# 지원하지 않는 언어 (default 템플릿 반환)
template_path = get_language_template("fortran")
print(template_path)
# 출력: .moai/project/tech/default.md.j2

# None 입력 (default 템플릿 반환)
template_path = get_language_template(None)
print(template_path)
# 출력: .moai/project/tech/default.md.j2
```

#### ProjectInitializer와 통합

```python
from moai_adk.core.project import ProjectInitializer
from moai_adk.core.template.languages import get_language_template

# 프로젝트 초기화
initializer = ProjectInitializer()
result = initializer.initialize()

# 감지된 언어에 맞는 템플릿 경로 가져오기
detected_language = result['language']
template_path = get_language_template(detected_language)

print(f"프로젝트 언어: {detected_language}")
print(f"사용할 템플릿: {template_path}")

# 템플릿 복사 또는 렌더링
# (실제 템플릿 처리는 template processor에서 수행)
```

#### LANGUAGE_TEMPLATES 딕셔너리

전체 매핑 테이블에 직접 접근할 수 있습니다:

```python
from moai_adk.core.template.languages import LANGUAGE_TEMPLATES

# 지원 언어 목록 조회
supported_languages = list(LANGUAGE_TEMPLATES.keys())
print(f"지원 언어 ({len(supported_languages)}개):")
for lang in supported_languages:
    print(f"  - {lang}")

# 출력:
# 지원 언어 (20개):
#   - python
#   - typescript
#   - javascript
#   - java
#   - go
#   - rust
#   - ...
```

---

## 통합 사용 예시

### 완전한 프로젝트 초기화 워크플로우

```python
from pathlib import Path
from moai_adk.core.project import ProjectInitializer, LanguageDetector, SystemChecker
from moai_adk.core.template.languages import get_language_template
from moai_adk.core.git import GitManager
from moai_adk.core.git.branch import generate_branch_name
from moai_adk.core.git.commit import format_commit_message

def initialize_project(project_path: str = "."):
    """프로젝트 전체 초기화 워크플로우"""

    # 1. 시스템 요구사항 검증
    print("1. 시스템 요구사항 검증 중...")
    checker = SystemChecker()
    system_check = checker.check_all()

    if not (system_check['git'] and system_check['python']):
        raise RuntimeError("필수 도구가 설치되지 않았습니다.")

    print("  ✅ 시스템 준비 완료")

    # 2. 언어 감지
    print("\n2. 프로젝트 언어 감지 중...")
    detector = LanguageDetector(project_path)
    detected_language = detector.detect()
    all_languages = detector.detect_multiple()

    print(f"  주언어: {detected_language}")
    if len(all_languages) > 1:
        print(f"  보조 언어: {', '.join(all_languages[1:])}")

    # 3. 프로젝트 초기화
    print("\n3. .moai/ 디렉토리 구조 생성 중...")
    initializer = ProjectInitializer(project_path)

    if initializer.is_initialized():
        print("  ⚠️ 이미 초기화된 프로젝트입니다.")
        return

    result = initializer.initialize(
        mode="team",
        locale="ko",
        language=detected_language
    )
    print(f"  ✅ 프로젝트 '{result['path']}' 초기화 완료")

    # 4. 템플릿 경로 확인
    print("\n4. 기술 스택 템플릿 확인 중...")
    template_path = get_language_template(result['language'])
    print(f"  템플릿: {template_path}")

    # 5. Git 저장소 확인
    print("\n5. Git 저장소 확인 중...")
    git_manager = GitManager(project_path)
    if git_manager.is_repo():
        print(f"  현재 브랜치: {git_manager.current_branch()}")

        # 초기 커밋 (변경사항이 있을 경우)
        if git_manager.is_dirty():
            print("\n6. 초기화 커밋 생성 중...")
            message = format_commit_message(
                "docs",
                "프로젝트 초기화 (.moai/ 생성)",
                locale=result['locale']
            )
            git_manager.commit(message=message)
            print(f"  ✅ 커밋: {message}")
    else:
        print("  ⚠️ Git 저장소가 아닙니다. 'git init' 실행을 권장합니다.")

    print("\n✅ 프로젝트 초기화 완료!")
    print("\n다음 단계:")
    print("  1. /alfred:0-project  # 프로젝트 문서 작성")
    print("  2. /alfred:1-spec     # SPEC 작성")
    print("  3. /alfred:2-build    # TDD 구현")

# 실행
if __name__ == "__main__":
    initialize_project("/path/to/new/project")
```

### TDD 워크플로우 자동화

```python
from moai_adk.core.git import GitManager
from moai_adk.core.git.branch import generate_branch_name
from moai_adk.core.git.commit import format_commit_message

def tdd_workflow(spec_id: str, base_branch: str = "develop"):
    """SPEC ID 기반 완전한 TDD 워크플로우"""

    manager = GitManager()

    # 1. feature 브랜치 생성
    branch_name = generate_branch_name(spec_id)
    print(f"1. 브랜치 생성: {branch_name}")
    manager.create_branch(branch_name, from_branch=base_branch)

    # 2. RED: 테스트 작성
    print("\n2. RED 단계: 테스트 작성")
    input("테스트를 작성한 후 Enter를 누르세요...")

    manager.commit(
        message=format_commit_message("red", f"{spec_id} 테스트 작성"),
        files=["tests/"]
    )
    print("  ✅ RED 커밋 완료")

    # 3. GREEN: 구현
    print("\n3. GREEN 단계: 최소 구현")
    input("구현을 완료한 후 Enter를 누르세요...")

    manager.commit(
        message=format_commit_message("green", f"{spec_id} 구현"),
    )
    print("  ✅ GREEN 커밋 완료")

    # 4. REFACTOR: 리팩토링
    print("\n4. REFACTOR 단계: 코드 품질 개선")
    input("리팩토링을 완료한 후 Enter를 누르세요...")

    manager.commit(
        message=format_commit_message("refactor", f"{spec_id} 코드 품질 개선"),
    )
    print("  ✅ REFACTOR 커밋 완료")

    # 5. 원격 푸시
    print("\n5. 원격 저장소 푸시")
    manager.push(set_upstream=True)
    print("  ✅ 푸시 완료")

    print(f"\n✅ TDD 워크플로우 완료: {branch_name}")
    print(f"다음 단계: GitHub에서 PR 생성 ({branch_name} → {base_branch})")

# 실행
if __name__ == "__main__":
    tdd_workflow("AUTH-001", base_branch="develop")
```

---

## 참고 문서

- **SPEC 문서**: `.moai/specs/SPEC-CORE-GIT-001/spec.md` - Git 모듈 상세 명세
- **SPEC 문서**: `.moai/specs/SPEC-CORE-PROJECT-001/spec.md` - Project 모듈 상세 명세
- **개발 가이드**: `.moai/memory/development-guide.md` - TDD 워크플로우 및 코딩 규칙
- **테스트 코드**: `tests/unit/test_git*.py` - Git 모듈 테스트 예시
- **테스트 코드**: `tests/unit/test_project*.py` - Project 모듈 테스트 예시

---

**최종 업데이트**: 2025-10-14
**버전**: v0.3.0
**작성자**: @doc-syncer
