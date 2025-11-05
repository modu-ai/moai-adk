# 다국어 린트/포맷 아키텍처 - 개발자 가이드

## 개요

MoAI-ADK의 다국어 린트/포맷 아키텍처는 사용자 프로젝트가 어떤 언어로 작성되었든 자동으로 해당 언어의 린팅과 포맷팅을 수행합니다.

**지원 언어:**
- Python (ruff)
- JavaScript (eslint + prettier)
- TypeScript (tsc + eslint + prettier)
- Go (golangci-lint + gofmt)
- Rust (clippy + rustfmt)
- Java (checkstyle + spotless)
- Ruby (rubocop)
- PHP (phpstan + php-cs-fixer)
- C# (Roslyn, dotnet)
- Kotlin (ktlint, gradle)

## 아키텍처 구조

### 코어 모듈

```
.claude/hooks/alfred/core/
├── __init__.py                                 # 패키지 초기화
├── language_detector.py                        # 언어 자동 감지
├── linters.py                                  # 언터-특정 린터 레지스트리
├── formatters.py                               # 언어-특정 포매터 레지스트리
├── post_tool__multilingual_linting.py          # PostToolUse 린트 훅
├── post_tool__multilingual_formatting.py       # PostToolUse 포맷 훅
├── test_language_detector.py                   # 언어 감지 단위 테스트
├── test_linters.py                             # 린터 단위 테스트
├── test_formatters.py                          # 포매터 단위 테스트
└── test_multilingual_integration.py            # 통합 테스트
```

### 실행 흐름

```
파일 수정 (Write/Edit/MultiEdit)
    ↓
PostToolUse Hook 트리거
    ↓
post_tool__multilingual_linting.py
    ├─ 프로젝트 언어 감지 (LanguageDetector)
    ├─ 수정된 파일 언어 매핑
    ├─ 언어-특정 린터 실행 (LinterRegistry)
    └─ 결과 요약 출력
    ↓
post_tool__multilingual_formatting.py
    ├─ 프로젝트 언어 감지 (LanguageDetector)
    ├─ 수정된 파일 언어 매핑
    ├─ 언어-특정 포매터 실행 (FormatterRegistry)
    └─ 결과 요약 출력
```

## 각 모듈 상세 설명

### 1. LanguageDetector (language_detector.py)

프로젝트의 사용 언어를 자동으로 감지합니다.

**주요 메서드:**

```python
# 모든 감지된 언어 목록 반환 (우선순위 순)
languages = detector.detect_languages()
# ['typescript', 'python', 'go']

# 주 언어 반환
primary = detector.detect_primary_language()
# 'typescript'

# 특정 언어의 파일 확장자 반환
exts = detector.get_file_extension_for_language('typescript')
# ['.ts', '.tsx']

# 패키지 관리자 반환
manager = detector.get_package_manager('python')
# 'pip'

# 언어 설치 여부 확인
installed = detector.is_language_installed('go')
# True/False

# 언어-특정 린터/포매터/타입체커 도구 반환
tools = detector.get_linter_tools('typescript')
# {'formatter': 'prettier', 'linter': 'eslint', 'type_checker': 'tsc'}
```

**언어 감지 마커:**

| 언어 | 감지 파일 |
|------|----------|
| Python | pyproject.toml, setup.py, requirements.txt, Pipfile |
| JavaScript | package.json, webpack.config.js, babel.config.js |
| TypeScript | tsconfig.json (package.json 이전 우선) |
| Go | go.mod, go.sum |
| Rust | Cargo.toml, Cargo.lock |
| Java | pom.xml, build.gradle, settings.gradle |
| Ruby | Gemfile, Gemfile.lock, Rakefile |
| PHP | composer.json, composer.lock, phpunit.xml |
| C# | *.csproj, *.sln |
| Kotlin | build.gradle.kts, pom.xml |

### 2. LinterRegistry (linters.py)

각 언어의 린팅을 수행합니다.

**주요 메서드:**

```python
# 단일 파일 린팅
success = registry.run_linter('python', Path('src/main.py'))
# True (통과) 또는 False (오류 발견)

# 파일 포매팅
success = registry.run_formatter('python', Path('src/main.py'))
# True (성공) 또는 False (오류)
```

**각 언어의 린팅 도구:**

| 언어 | 린터 | 포매터 | 타입체커 |
|------|------|--------|---------|
| Python | ruff check | ruff format | mypy |
| JavaScript | eslint | prettier | - |
| TypeScript | eslint | prettier | tsc |
| Go | golangci-lint | gofmt | - |
| Rust | cargo clippy | rustfmt | - |
| Java | checkstyle | spotless | - |
| Ruby | rubocop | rubocop -a | - |
| PHP | phpstan | php-cs-fixer | psalm |

**특징:**

- **Non-blocking**: 도구가 없거나 오류 발생 시에도 실행을 계속 진행
- **타임아웃 처리**: 각 도구는 30-60초 타임아웃 설정
- **상세 로깅**: 각 단계의 성공/실패를 로그로 기록

### 3. FormatterRegistry (formatters.py)

각 언어의 포매팅을 수행합니다.

**주요 메서드:**

```python
# 단일 파일 포매팅
success = registry.format_file('python', Path('src/main.py'))

# 디렉토리 배치 포매팅
success = registry.format_directory('python', Path('src'), ['.py'])
```

**특징:**

- **자동 수정**: 포매터가 파일을 직접 수정
- **배치 처리**: 여러 파일을 한 번에 포매팅
- **안전성**: 원본 파일을 백업하지 않음 (Git으로 관리)

### 4. MultilingualLintingHook (post_tool__multilingual_linting.py)

PostToolUse 이벤트를 처리하는 메인 린팅 훅입니다.

**주요 메서드:**

```python
# 파일에 대한 언어 결정
language = hook.get_language_for_file(Path('test.py'))
# 'python'

# 단일 파일 린팅
success = hook.lint_file(Path('src/main.py'))

# 여러 파일 린팅 및 요약 생성
summary = hook.lint_files([Path('src/main.py'), Path('src/util.py')])
# {
#   'status': 'completed',
#   'total_files': 2,
#   'files_checked': 2,
#   'files_with_issues': 0,
#   'files_by_language': {...},
#   'languages_detected': ['python']
# }

# 요약 메시지 생성
message = hook.get_summary_message(summary)
```

**필터링:**

- 숨겨진 파일 제외 (`.gitignore`, `.hidden.py`)
- 숨겨진 디렉토리 제외 (`.git/`, `.venv/`)
- 종속성 디렉토리 제외 (`node_modules/`, `__pycache__/`)

### 5. MultilingualFormattingHook (post_tool__multilingual_formatting.py)

PostToolUse 이벤트를 처리하는 메인 포매팅 훅입니다.

**주요 메서드:**

```python
# 파일에 대한 언어 결정
language = hook.get_language_for_file(Path('test.py'))
# 'python'

# 단일 파일 포매팅
success = hook.format_file(Path('src/main.py'))

# 여러 파일 포매팅 및 요약 생성
summary = hook.format_files([Path('src/main.py'), Path('src/util.py')])

# 요약 메시지 생성
message = hook.get_summary_message(summary)
```

**필터링:**

- 숨겨진 파일/디렉토리 제외
- `node_modules/`, `dist/`, `build/` 제외
- 번들/압축 파일 제외 (`.min.js`, `.bundle.js`)

## 설치 및 설정

### 의존성 설치

#### Python 프로젝트

```bash
# ruff (린팅 + 포매팅)
uv add ruff --optional

# mypy (타입 체크)
uv add mypy --optional
```

#### JavaScript/TypeScript 프로젝트

```bash
# eslint (린팅)
npm install --save-dev eslint

# prettier (포매팅)
npm install --save-dev prettier

# TypeScript (타입 체크)
npm install --save-dev typescript
```

#### Go 프로젝트

```bash
# golangci-lint (린팅)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# gofmt는 Go 표준 도구이므로 자동 설치됨
```

#### Rust 프로젝트

```bash
# clippy는 cargo 표준 도구
cargo clippy

# rustfmt는 cargo 표준 도구
cargo fmt
```

#### Java 프로젝트

```bash
# Maven을 사용하는 경우
mvn clean install

# Gradle을 사용하는 경우
gradle build
```

#### Ruby 프로젝트

```bash
# rubocop (린팅 + 자동 수정)
gem install rubocop
```

#### PHP 프로젝트

```bash
# phpstan (타입 체크)
composer require --dev phpstan/phpstan

# php-cs-fixer (포매팅)
composer require --dev friendsofphp/php-cs-fixer
```

### 훅 활성화

`.claude/settings.json`에 이미 설정되어 있습니다:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "command": "uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/core/post_tool__multilingual_linting.py",
            "description": "Run multilingual linting checks"
          },
          {
            "command": "uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/core/post_tool__multilingual_formatting.py",
            "description": "Run multilingual code formatting"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ]
  }
}
```

## 테스트 실행

### 모든 테스트 실행

```bash
cd .claude/hooks/alfred/core

# 전체 테스트
pytest . -v

# 특정 모듈 테스트
pytest test_language_detector.py -v
pytest test_linters.py -v
pytest test_formatters.py -v
pytest test_multilingual_integration.py -v
```

### 테스트 커버리지

```bash
pytest --cov=. --cov-report=html
```

## 문제 해결

### 1. 도구를 찾을 수 없음

**증상:** `⚠️ Ruff not installed`

**해결책:**
```bash
# Python 프로젝트
uv add ruff

# 또는 pip 사용
pip install ruff
```

### 2. 타임아웃 오류

**증상:** `⏱️ Linter timeout`

**해결책:** 큰 프로젝트의 경우 시간 초과가 발생할 수 있습니다. 이 경우:

```python
# linters.py 또는 formatters.py에서 타임아웃 증가
timeout=60  # 30에서 60으로 증가
```

### 3. 실패한 린팅 오류

**증상:** `🔴 Python lint errors`

**해결책:**

```bash
# ruff로 자동 수정
ruff check --fix src/

# 또는 ruff로 포매팅
ruff format src/
```

### 4. 린팅 규칙 커스터마이징

**Python (ruff):**

`.claude/hooks/alfred/core/linters.py` 수정:
```python
# 린트 규칙 커스터마이징
["ruff", "check", str(file_path), "--select=E,F,W,I,N,D"]
```

**JavaScript (eslint):**

프로젝트 루트에 `.eslintrc.json` 추가:
```json
{
  "extends": "eslint:recommended",
  "rules": {
    "semi": ["error", "always"]
  }
}
```

**Go (golangci-lint):**

프로젝트 루트에 `.golangci.yml` 추가:
```yaml
linters:
  enable:
    - gofmt
    - govet
```

## 성능 최적화

### 1. 린팅 캐싱

일부 린터는 캐싱을 지원합니다:

```bash
# ruff는 캐싱 자동 지원
# eslint 캐싱 활성화
npx eslint --cache src/

# golangci-lint 캐싱 활성화
golangci-lint run --cache
```

### 2. 병렬 처리

대규모 프로젝트의 경우:

```python
# MultilingualLintingHook에서 멀티스레딩 추가 가능
from concurrent.futures import ThreadPoolExecutor

with ThreadPoolExecutor(max_workers=4) as executor:
    results = executor.map(hook.lint_file, file_list)
```

### 3. 선택적 린팅

`.claude/hooks/alfred/core/post_tool__multilingual_linting.py`에서:

```python
# 특정 파일 타입만 린팅
def should_lint_file(self, file_path: Path) -> bool:
    # 테스트 파일 제외
    if 'test' in file_path.name:
        return False
    # ...
```

## API 확장

### 새로운 언어 추가

#### Step 1: LanguageDetector에 언어 마커 추가

```python
# language_detector.py
LANGUAGE_MARKERS = {
    "kotlin": [
        "build.gradle.kts",
        "pom.xml",
    ],
    # ...
}
```

#### Step 2: 파일 확장자 추가

```python
def get_file_extension_for_language(self, language: str) -> List[str]:
    extensions = {
        "kotlin": [".kt", ".kts"],
        # ...
    }
```

#### Step 3: 패키지 관리자 추가

```python
def get_package_manager(self, language: str) -> str:
    managers = {
        "kotlin": "gradle",
        # ...
    }
```

#### Step 4: 린터 도구 추가

```python
def get_linter_tools(self, language: str) -> Dict[str, str]:
    tools = {
        "kotlin": {
            "formatter": "ktlint",
            "linter": "ktlint",
            "type_checker": None,
        },
        # ...
    }
```

#### Step 5: LinterRegistry에 언어 지원 추가

```python
# linters.py
def _run_kotlin_linting(self, file_path: Path) -> bool:
    """Run ktlint for Kotlin"""
    if file_path.suffix not in [".kt", ".kts"]:
        return True

    try:
        result = subprocess.run(
            ["ktlint", str(file_path)],
            capture_output=True,
            text=True,
            timeout=30
        )
        # ...
```

#### Step 6: FormatterRegistry에 포매터 추가

```python
# formatters.py
def _format_kotlin(self, file_path: Path) -> bool:
    """Format Kotlin with ktlint"""
    if file_path.suffix not in [".kt", ".kts"]:
        return True

    try:
        result = subprocess.run(
            ["ktlint", "-F", str(file_path)],
            # ...
```

#### Step 7: 테스트 추가

```python
# test_linters.py
def test_kotlin_linting(self):
    """Test Kotlin linting"""
    registry = LinterRegistry()

    with tempfile.TemporaryDirectory() as tmpdir:
        file_path = Path(tmpdir) / "Main.kt"
        file_path.write_text("fun main() {}")

        result = registry.run_linter("kotlin", file_path)
        assert isinstance(result, bool)
```

## 참고 자료

### 공식 문서

- [ruff](https://docs.astral.sh/ruff/)
- [eslint](https://eslint.org/docs/)
- [prettier](https://prettier.io/docs/)
- [golangci-lint](https://golangci-lint.run/)
- [clippy](https://doc.rust-lang.org/clippy/)

### 관련 파일

- `.claude/settings.json` - Hook 설정
- `.moai/config.json` - 프로젝트 설정
- `CLAUDE.md` - 프로젝트 지침

## 라이선스

MoAI-ADK는 MIT 라이선스입니다.
