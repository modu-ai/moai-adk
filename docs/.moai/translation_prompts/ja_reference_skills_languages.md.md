Translate the following Korean markdown document to Japanese.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/reference/skills/languages.md
**Target Language:** Japanese
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ja/reference/skills/languages.md

**Content to Translate:**

# Language Skills 완전 가이드

MoAI-ADK가 지원하는 20개+ 프로그래밍 언어의 스킬 참고서입니다.

## :bullseye: 개요

각 언어 스킬은 해당 언어의 최신 버전, 테스트 프레임워크, 린팅 도구, 그리고 모범 사례를 포함합니다.

## 🐍 Python (가장 인기)

### 기본 정보

- **버전**: Python 3.13+
- **패키지 관리**: uv 0.9.3
- **테스트**: pytest 8.4.2
- **타입 검사**: mypy 1.8.0
- **린팅**: ruff 0.13.1
- **프레임워크**: FastAPI, Flask, Django

### 구조

```
python_project/
├── src/
│   ├── models.py
│   ├── services.py
│   └── __init__.py
├── tests/
│   ├── unit/
│   ├── integration/
│   └── conftest.py
├── pyproject.toml
└── .python-version
```

### 핵심 도구

```bash
# 테스트
pytest tests/ -v --cov=src

# 타입 검사
mypy src/

# 린팅
ruff check src/

# 포매팅
black src/

# 패키지 설치
uv add pytest
uv sync
```

### 예시

```python
# @CODE:SPEC-001:example
def add(a: int, b: int) -> int:
    """두 수를 더함"""
    return a + b

# @TEST:SPEC-001:test_add
def test_add():
    assert add(2, 3) == 5
```

______________________________________________________________________

## 📘 TypeScript (웹 개발)

### 기본 정보

- **버전**: TypeScript 5.7+
- **런타임**: Node.js 20+
- **패키지 관리**: npm, pnpm, bun
- **테스트**: Vitest 2.1
- **린팅**: Biome 1.9
- **프레임워크**: React 19, Vue 3.5, Angular 19

### 구조

```
typescript_project/
├── src/
│   ├── components/
│   ├── types/
│   ├── utils/
│   └── main.ts
├── tests/
│   ├── unit/
│   └── integration/
├── tsconfig.json
└── package.json
```

### 핵심 도구

```bash
# 테스트
vitest run

# 타입 검사
tsc --noEmit

# 린팅
biome check src/

# 포매팅
biome format src/
```

______________________________________________________________________

## 🦀 Rust (시스템 프로그래밍)

### 기본 정보

- **버전**: Rust 1.84+
- **빌드 도구**: Cargo
- **테스트**: cargo test
- **린팅**: clippy
- **포매팅**: rustfmt
- **웹 프레임워크**: Axum, Rocket

### 구조

```
rust_project/
├── src/
│   ├── lib.rs
│   ├── main.rs
│   └── mod.rs
├── tests/
├── Cargo.toml
└── Cargo.lock
```

### 핵심 도구

```bash
# 테스트
cargo test

# 린팅
cargo clippy

# 포매팅
cargo fmt

# 빌드
cargo build --release
```

______________________________________________________________________

## 🐹 Go (백엔드)

### 기본 정보

- **버전**: Go 1.24+
- **패키지 관리**: go modules
- **테스트**: go test
- **린팅**: golangci-lint
- **포매팅**: gofmt
- **웹 프레임워크**: Gin, Beego

### 구조

```
go_project/
├── cmd/
│   └── main/main.go
├── internal/
│   ├── handlers/
│   └── services/
├── pkg/
├── tests/
└── go.mod
```

### 핵심 도구

```bash
# 테스트
go test ./...

# 린팅
golangci-lint run

# 포매팅
gofmt -w .

# 빌드
go build -o app
```

______________________________________________________________________

## 📜 기타 인기 언어

### Java (21+)

- 테스트: JUnit 5, Mockito
- 빌드: Maven, Gradle
- 린팅: Checkstyle, SpotBugs
- 프레임워크: Spring Boot, Quarkus

### C# (13+)

- 테스트: xUnit, NUnit
- 빌드: .NET SDK
- 린팅: StyleCop, FxCop
- 프레임워크: ASP.NET Core, Blazor

### PHP (8.4+)

- 테스트: PHPUnit 11
- 패키지: Composer
- 린팅: PHP_CodeSniffer, PHPStan
- 프레임워크: Laravel 11, Symfony 7

### Ruby (3.4+)

- 테스트: RSpec 4
- 패키지: Bundler, RubyGems
- 린팅: RuboCop 2
- 프레임워크: Rails 8

### Kotlin (2.1+)

- 테스트: KUnit, Mockk
- 빌드: Gradle, Maven
- 린팅: ktlint, detekt
- 프레임워크: Spring Boot, Ktor

### Dart (3.x)

- 테스트: Dart test
- 빌드: Pub, Flutter
- 린팅: analysis server
- 프레임워크: Flutter 3.27

## :card_index_dividers: 언어별 스킬 활성화

### 자동 활성화

```
SPEC에서 언어 감지
    ├─ "Python" 감지
    │   └─→ Skill("moai-lang-python") 자동 로드
    ├─ "TypeScript" 감지
    │   └─→ Skill("moai-lang-typescript") 자동 로드
    ├─ "Rust" 감지
    │   └─→ Skill("moai-lang-rust") 자동 로드
    └─ "Go" 감지
        └─→ Skill("moai-lang-go") 자동 로드
```

### 수동 호출

```python
# 특정 언어 스킬 명시적 호출
Skill("moai-lang-python")
Skill("moai-lang-typescript")
Skill("moai-lang-go")
Skill("moai-lang-rust")
Skill("moai-lang-java")
Skill("moai-lang-kotlin")
Skill("moai-lang-csharp")
Skill("moai-lang-ruby")
Skill("moai-lang-php")
Skill("moai-lang-sql")
Skill("moai-lang-shell")
```

## 📊 언어 스킬 비교

| 언어       | 버전  | 테스트     | 린팅            | 타입     | 성능       |
| ---------- | ----- | ---------- | --------------- | -------- | ---------- |
| Python     | 3.13+ | pytest     | ruff            | mypy     | ⭐⭐⭐     |
| TypeScript | 5.7+  | Vitest     | Biome           | Built-in | ⭐⭐⭐⭐   |
| Rust       | 1.84+ | cargo test | clippy          | Built-in | ⭐⭐⭐⭐⭐ |
| Go         | 1.24+ | go test    | golangci        | Built-in | ⭐⭐⭐⭐⭐ |
| Java       | 21+   | JUnit      | Checkstyle      | Built-in | ⭐⭐⭐⭐   |
| C#         | 13+   | xUnit      | StyleCop        | Built-in | ⭐⭐⭐⭐   |
| Ruby       | 3.4+  | RSpec      | RuboCop         | Optional | ⭐⭐⭐     |
| PHP        | 8.4+  | PHPUnit    | PHP_CodeSniffer | PHPStan  | ⭐⭐⭐     |
| Kotlin     | 2.1+  | KUnit      | ktlint          | Built-in | ⭐⭐⭐⭐   |

## :bullseye: 언어 선택 가이드

### API/백엔드

추천: **Python** (FastAPI) > **Go** > **Rust** > **Java** > **TypeScript** (Node.js)

### 프론트엔드

추천: **TypeScript** > **JavaScript** > **Dart** (Flutter)

### 모바일

추천: **Kotlin** (Android) > **Swift** (iOS) > **Dart** (Flutter - 크로스플랫폼)

### 시스템 프로그래밍

추천: **Rust** > **Go** > **C++** > **C**

### 데이터 과학/ML

추천: **Python** > **R** > **Julia**

### 웹 풀스택

추천: **TypeScript** (React/Node.js) > **Ruby** (Rails) > **Python** (Django)

## :link: 언어별 모범 사례

모든 언어 스킬은 다음을 포함합니다:

- ✅ 최신 버전 명시
- ✅ 프로젝트 구조 가이드
- ✅ 테스트 프레임워크 통합
- ✅ 린팅 및 포매팅 자동화
- ✅ 타입 안전성 보장
- ✅ 성능 최적화 기법
- ✅ 배포 자동화
- ✅ 모범 사례 및 안티패턴

## <span class="material-icons">library_books</span> 상세 문서

각 언어별 스킬은 다음을 제공합니다:

- **구조**: 프로젝트 디렉토리 구조
- **도구**: 테스트, 린팅, 포매팅 도구
- **예시**: 실제 코드 예시
- **성능**: 성능 최적화 가이드
- **테스트**: 테스트 작성 패턴
- **배포**: CI/CD 자동화

______________________________________________________________________

**다음**: [Alfred Skills](alfred.md) 또는 [Skills 개요](index.md)


**Instructions:**
- Translate the content above to Japanese
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
