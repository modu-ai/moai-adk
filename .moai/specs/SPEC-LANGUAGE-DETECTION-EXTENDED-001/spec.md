---
id: LANGUAGE-DETECTION-EXTENDED-001
version: 1.0.0
status: completed
created: 2025-10-30
updated: 2025-10-31
author: @GoosLab
priority: high
category: feature
depends_on: LANGUAGE-DETECTION-001
---

# @SPEC:LANGUAGE-DETECTION-EXTENDED-001: 11개 언어 전담 CI/CD 워크플로우 확장

## HISTORY

### v1.0.0 (2025-10-31) - COMPLETED

- **작성자**: @GoosLab
- **변경사항**: SPEC 구현 완료 및 PR #135 병합
- **설명**: 11개 언어 전담 CI/CD 워크플로우 지원 완료, LanguageDetector 클래스 확장, 34개 테스트 추가
- **커밋**: PR #135 병합 (commit 449e7b42)
- **상태**: 마스터 브랜치 병합 완료, 프로덕션 배포 준비됨

### v0.0.1 (2025-10-30) - INITIAL

- **작성자**: @GoosLab
- **변경사항**: 초기 SPEC 작성
- **설명**: GitHub Issue #131 확장 단계 - 11개 언어 전담 CI/CD 워크플로우 지원 (Ruby, PHP, Java, Rust, Dart, Swift, Kotlin, C#, C, C++, Shell)
- **의존성**: SPEC-LANGUAGE-DETECTION-001 (기본 4개 언어 지원)

---

## 1. Environment (환경)

### 1.1 시스템 컨텍스트

MoAI-ADK는 현재 4개 언어(Python, JavaScript, TypeScript, Go)에 대한 전담 CI/CD 워크플로우를 지원합니다. 이 SPEC은 추가 11개 언어(Ruby, PHP, Java, Rust, Dart, Swift, Kotlin, C#, C, C++, Shell)로 지원 범위를 확장하여 총 15개 언어를 지원합니다.

### 1.2 기술 환경

- **지원 언어 (확장 대상)**:
  - Ruby (RSpec, Rubocop, bundle)
  - PHP (PHPUnit, PHPCS, composer)
  - Java (JUnit 5, Jacoco, Maven/Gradle)
  - Rust (cargo test, clippy, rustfmt)
  - Dart (flutter test, dart analyze)
  - Swift (XCTest, SwiftLint)
  - Kotlin (JUnit 5, ktlint, gradle)
  - C# (xUnit, StyleCop, dotnet)
  - C (gcc/clang, cppcheck, Unity)
  - C++ (g++/clang++, gtest, cpplint)
  - Shell (shellcheck, bats-core)

- **기존 지원 언어** (SPEC-LANGUAGE-DETECTION-001):
  - Python, JavaScript, TypeScript, Go

- **CI/CD 플랫폼**: GitHub Actions
- **워크플로우 템플릿 위치**: `src/moai_adk/templates/.github/workflows/`
- **언어 감지 모듈**: `src/moai_adk/core/language_detector.py`

### 1.3 제약사항

- 각 언어별 최소 하나 이상의 테스트 프레임워크 지원
- 빌드 도구 자동 감지 (Maven vs Gradle, CMake vs Bazel 등)
- 하위 호환성 유지 (기존 4개 언어 워크플로우)
- GitHub Actions runner 호환성

---

## 2. Assumptions (가정)

### 2.1 프로젝트 구조 가정

- **Ruby**: `Gemfile` 또는 `Gemfile.lock`이 프로젝트 루트에 존재
- **PHP**: `composer.json` 또는 `composer.lock`이 존재
- **Java**: `pom.xml` (Maven) 또는 `build.gradle`/`build.gradle.kts` (Gradle)가 존재
- **Rust**: `Cargo.toml`이 존재
- **Dart**: `pubspec.yaml`이 존재 (Flutter 포함)
- **Swift**: `Package.swift` (Swift Package Manager) 또는 `.xcodeproj`/`.xcworkspace` (Xcode) 존재
- **Kotlin**: `build.gradle.kts` 또는 Kotlin 소스 파일(`.kt`) 존재
- **C#**: `.csproj`, `.sln` 또는 `Directory.Build.props` 존재
- **C**: `CMakeLists.txt` + `.c` 파일 존재
- **C++**: `CMakeLists.txt` + `.cpp` 파일 존재
- **Shell**: `.sh` 파일 또는 shebang(`#!/bin/bash`) 존재

### 2.2 도구 가용성 가정

- 각 언어별 표준 빌드/테스트 도구가 설치 가능함
- GitHub Actions runner가 각 언어 런타임을 지원함
- 오픈소스 린팅 도구가 사용 가능함

### 2.3 사용자 역량 가정

- 사용자는 자신의 프로젝트 언어에 익숙함
- 사용자는 기본적인 CI/CD 개념을 이해함
- 사용자는 필요시 워크플로우를 수동으로 커스터마이징할 수 있음

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous Requirements (보편적 요구사항)

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-U1**
시스템은 기존 4개 언어(Python, JavaScript, TypeScript, Go) 외에 추가 11개 언어(Ruby, PHP, Java, Rust, Dart, Swift, Kotlin, C#, C, C++, Shell)를 자동 감지해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-U2**
시스템은 각 언어별 최적화된 빌드, 테스트, 린팅 도구를 자동 선택해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-U3**
시스템은 15개 언어 모두에 대해 전담 CI/CD 워크플로우 템플릿을 제공해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-U4**
시스템은 하위 호환성을 유지하며 기존 4개 언어 워크플로우에 영향을 주지 않아야 한다.

### 3.2 Event-driven Requirements (이벤트 기반 요구사항)

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E1**
WHEN Ruby 프로젝트가 감지될 때 (`Gemfile` 존재), 시스템은 `ruby-tag-validation.yml` 워크플로우를 선택해야 한다 (RSpec, Rubocop 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E2**
WHEN PHP 프로젝트가 감지될 때 (`composer.json` 존재), 시스템은 `php-tag-validation.yml`를 선택해야 한다 (PHPUnit, PHPCS 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E3**
WHEN Java 프로젝트가 감지될 때 (`pom.xml` 또는 `build.gradle` 존재), 시스템은 `java-tag-validation.yml`를 선택해야 한다 (JUnit 5, Jacoco 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E4**
WHEN Rust 프로젝트가 감지될 때 (`Cargo.toml` 존재), 시스템은 `rust-tag-validation.yml`를 선택해야 한다 (cargo test, clippy 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E5**
WHEN Dart 프로젝트가 감지될 때 (`pubspec.yaml` 존재), 시스템은 `dart-tag-validation.yml`를 선택해야 한다 (flutter test, dart analyze 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E6**
WHEN Swift 프로젝트가 감지될 때 (`Package.swift` 또는 `.xcodeproj` 존재), 시스템은 `swift-tag-validation.yml`를 선택해야 한다 (XCTest, SwiftLint 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E7**
WHEN Kotlin 프로젝트가 감지될 때 (`build.gradle.kts` 존재), 시스템은 `kotlin-tag-validation.yml`를 선택해야 한다 (JUnit 5, ktlint 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E8**
WHEN C# 프로젝트가 감지될 때 (`.csproj` 또는 `.sln` 존재), 시스템은 `csharp-tag-validation.yml`를 선택해야 한다 (xUnit, StyleCop 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E9**
WHEN C 프로젝트가 감지될 때 (`CMakeLists.txt` + `.c` 파일 존재), 시스템은 `c-tag-validation.yml`를 선택해야 한다 (gcc/clang, cppcheck 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E10**
WHEN C++ 프로젝트가 감지될 때 (`CMakeLists.txt` + `.cpp` 파일 존재), 시스템은 `cpp-tag-validation.yml`를 선택해야 한다 (g++/clang++, gtest 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E11**
WHEN Shell 스크립트가 감지될 때 (`.sh` 파일 또는 shebang 존재), 시스템은 `shell-tag-validation.yml`를 선택해야 한다 (shellcheck, bats-core 사용).

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E12**
WHEN Java 프로젝트에서 `pom.xml`이 발견될 때, 시스템은 Maven 빌드 도구를 선택해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E13**
WHEN Java 프로젝트에서 `build.gradle` 또는 `build.gradle.kts`가 발견될 때, 시스템은 Gradle 빌드 도구를 선택해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-E14**
WHEN C/C++ 프로젝트에서 `CMakeLists.txt`가 발견될 때, 시스템은 CMake 빌드 도구를 선택해야 한다.

### 3.3 State-driven Requirements (상태 기반 요구사항)

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-S1**
WHILE 여러 언어의 설정 파일이 동시에 존재할 때, 시스템은 우선순위 규칙에 따라 하나의 언어만 선택해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-S2**
WHILE Java/Kotlin 프로젝트를 처리할 때, 시스템은 `pom.xml`과 `build.gradle` 간 빌드 도구 선택을 자동화해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-S3**
WHILE CI/CD 워크플로우가 실행 중일 때, 시스템은 언어별 환경 설정을 올바르게 적용해야 한다.

### 3.4 Optional Requirements (선택적 요구사항)

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-O1**
WHERE 사용자가 특정 언어 버전을 명시한 경우 (예: `ruby-version`, `java-version`), 시스템은 사용자 지정 버전을 우선 적용할 수 있다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-O2**
WHERE 사용자가 커스텀 빌드 스크립트를 제공한 경우, 시스템은 기본 워크플로우 대신 사용자 스크립트를 실행할 수 있다.

### 3.5 Unwanted Behaviors (원치 않는 동작)

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-N1**
IF 언어별 워크플로우 템플릿이 누락되면, 시스템은 명확한 에러 메시지를 표시하고 구현을 차단해야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-N2**
IF 호환되지 않는 도구 조합이 감지되면 (예: Maven + Gradle 동시 사용), 시스템은 경고를 발생시켜야 한다.

**@REQ:LANGUAGE-DETECTION-EXTENDED-001-N3**
IF 언어 감지 결과가 모호할 때 (예: `.c`와 `.cpp` 파일 혼재), 시스템은 사용자에게 명시적 선택을 요청해야 한다.

---

## 4. Specifications (상세 사양)

### 4.1 언어 감지 우선순위 (확장)

**우선순위 순서** (높음 → 낮음):

1. **Rust** (`Cargo.toml`)
2. **Dart** (`pubspec.yaml`)
3. **Swift** (`Package.swift`, `.xcodeproj`)
4. **Kotlin** (`build.gradle.kts`)
5. **C#** (`.csproj`, `.sln`)
6. **Java** (`pom.xml`, `build.gradle`)
7. **Ruby** (`Gemfile`)
8. **PHP** (`composer.json`)
9. **Go** (`go.mod`) — 기존
10. **Python** (`pyproject.toml`, `requirements.txt`) — 기존
11. **TypeScript** (`tsconfig.json`) — 기존
12. **JavaScript** (`package.json`) — 기존
13. **C++** (`CMakeLists.txt` + `.cpp`)
14. **C** (`CMakeLists.txt` + `.c`)
15. **Shell** (`.sh` 파일)

### 4.2 언어별 워크플로우 템플릿 사양

#### 4.2.1 Ruby 워크플로우 (`ruby-tag-validation.yml`)

```yaml
name: Ruby TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.2'
          bundler-cache: true
      - run: bundle exec rspec
      - run: bundle exec rubocop
```

#### 4.2.2 PHP 워크플로우 (`php-tag-validation.yml`)

```yaml
name: PHP TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: shivammathur/setup-php@v2
        with:
          php-version: '8.2'
          tools: composer, phpunit, phpcs
      - run: composer install
      - run: vendor/bin/phpunit
      - run: vendor/bin/phpcs
```

#### 4.2.3 Java 워크플로우 (`java-tag-validation.yml`)

```yaml
name: Java TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'
      # Maven 또는 Gradle 자동 감지
      - run: mvn test # pom.xml 존재 시
      - run: ./gradlew test # build.gradle 존재 시
```

#### 4.2.4 Rust 워크플로우 (`rust-tag-validation.yml`)

```yaml
name: Rust TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dtolnay/rust-toolchain@stable
      - run: cargo test
      - run: cargo clippy -- -D warnings
      - run: cargo fmt -- --check
```

#### 4.2.5 Dart/Flutter 워크플로우 (`dart-tag-validation.yml`)

```yaml
name: Dart TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dart-lang/setup-dart@v1
        with:
          sdk: 'stable'
      - run: dart pub get
      - run: dart analyze
      - run: dart test
      # Flutter 프로젝트인 경우:
      # - uses: subosito/flutter-action@v2
      # - run: flutter test
```

#### 4.2.6 Swift 워크플로우 (`swift-tag-validation.yml`)

```yaml
name: Swift TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      - run: swift build
      - run: swift test
      - run: swiftlint
```

#### 4.2.7 Kotlin 워크플로우 (`kotlin-tag-validation.yml`)

```yaml
name: Kotlin TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'
      - run: ./gradlew test
      - run: ./gradlew ktlintCheck
```

#### 4.2.8 C# 워크플로우 (`csharp-tag-validation.yml`)

```yaml
name: C# TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-dotnet@v4
        with:
          dotnet-version: '8.0.x'
      - run: dotnet restore
      - run: dotnet build
      - run: dotnet test
```

#### 4.2.9 C 워크플로우 (`c-tag-validation.yml`)

```yaml
name: C TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: sudo apt-get install -y cmake gcc cppcheck
      - run: cmake . && make
      - run: ctest
      - run: cppcheck --enable=all src/
```

#### 4.2.10 C++ 워크플로우 (`cpp-tag-validation.yml`)

```yaml
name: C++ TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: sudo apt-get install -y cmake g++ libgtest-dev
      - run: cmake . && make
      - run: ctest
      - run: cpplint src/**/*.cpp
```

#### 4.2.11 Shell 워크플로우 (`shell-tag-validation.yml`)

```yaml
name: Shell TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: sudo apt-get install -y shellcheck bats
      - run: shellcheck **/*.sh
      - run: bats tests/
```

### 4.3 LanguageDetector 클래스 확장

**파일**: `src/moai_adk/core/language_detector.py`

**추가 메서드**:

```python
def detect_ruby(self) -> bool:
    """Gemfile 또는 Gemfile.lock 존재 여부 확인"""
    return (self.project_root / "Gemfile").exists()

def detect_php(self) -> bool:
    """composer.json 존재 여부 확인"""
    return (self.project_root / "composer.json").exists()

def detect_java(self) -> bool:
    """pom.xml 또는 build.gradle 존재 여부 확인"""
    return (
        (self.project_root / "pom.xml").exists() or
        (self.project_root / "build.gradle").exists() or
        (self.project_root / "build.gradle.kts").exists()
    )

def detect_rust(self) -> bool:
    """Cargo.toml 존재 여부 확인"""
    return (self.project_root / "Cargo.toml").exists()

def detect_dart(self) -> bool:
    """pubspec.yaml 존재 여부 확인"""
    return (self.project_root / "pubspec.yaml").exists()

def detect_swift(self) -> bool:
    """Package.swift 또는 .xcodeproj 존재 여부 확인"""
    return (
        (self.project_root / "Package.swift").exists() or
        any(self.project_root.glob("*.xcodeproj"))
    )

def detect_kotlin(self) -> bool:
    """build.gradle.kts 또는 .kt 파일 존재 여부 확인"""
    return (
        (self.project_root / "build.gradle.kts").exists() or
        any(self.project_root.rglob("*.kt"))
    )

def detect_csharp(self) -> bool:
    """.csproj 또는 .sln 파일 존재 여부 확인"""
    return (
        any(self.project_root.glob("*.csproj")) or
        any(self.project_root.glob("*.sln"))
    )

def detect_c(self) -> bool:
    """CMakeLists.txt + .c 파일 존재 여부 확인"""
    return (
        (self.project_root / "CMakeLists.txt").exists() and
        any(self.project_root.rglob("*.c")) and
        not any(self.project_root.rglob("*.cpp"))
    )

def detect_cpp(self) -> bool:
    """CMakeLists.txt + .cpp 파일 존재 여부 확인"""
    return (
        (self.project_root / "CMakeLists.txt").exists() and
        any(self.project_root.rglob("*.cpp"))
    )

def detect_shell(self) -> bool:
    """.sh 파일 존재 여부 확인"""
    return any(self.project_root.rglob("*.sh"))

def detect_build_tool(self, language: str) -> str:
    """언어별 빌드 도구 자동 감지"""
    if language == "java":
        if (self.project_root / "pom.xml").exists():
            return "maven"
        elif (self.project_root / "build.gradle").exists():
            return "gradle"
    elif language in ["c", "cpp"]:
        if (self.project_root / "CMakeLists.txt").exists():
            return "cmake"
    return "default"
```

### 4.4 테스트 전략

**테스트 파일**: `tests/test_language_detector_extended.py`

**테스트 케이스**:

1. **11개 언어별 감지 테스트** (각 언어당 1개)
2. **우선순위 충돌 해결 테스트** (Java + Kotlin 동시 존재 등)
3. **빌드 도구 선택 테스트** (Maven vs Gradle)
4. **오류 처리 테스트** (필수 도구 누락 시나리오)
5. **하위 호환성 테스트** (기존 4개 언어 회귀 방지)

**총 테스트 수**: 25-30개

---

## 5. Traceability (추적성)

### 5.1 상위 의존성

- **@SPEC:LANGUAGE-DETECTION-001** - 기본 4개 언어 지원 (Python, JS, TS, Go)
- **GitHub Issue #131** - 11개 언어 확장 요청

### 5.2 관련 컴포넌트

- `src/moai_adk/core/language_detector.py` - 언어 감지 로직
- `src/moai_adk/templates/.github/workflows/` - 워크플로우 템플릿 디렉토리
- `tests/test_language_detector_extended.py` - 확장 테스트

### 5.3 향후 확장 가능성

- **@FUTURE:LANGUAGE-DETECTION-EXTENDED-002** - 추가 언어 지원 (Scala, Haskell, Elixir 등)
- **@FUTURE:CUSTOM-WORKFLOW-001** - 사용자 정의 워크플로우 템플릿 지원
- **@FUTURE:MULTI-LANGUAGE-001** - 다중 언어 프로젝트 지원 (monorepo)

---

## 6. Acceptance Criteria (인수 기준)

이 SPEC의 완료는 다음 조건을 충족해야 합니다:

1. ✅ 11개 언어별 워크플로우 템플릿이 모두 생성되고 템플릿 디렉토리에 배치됨
2. ✅ `LanguageDetector` 클래스가 11개 언어 감지 메서드를 포함함
3. ✅ 빌드 도구 자동 선택 로직이 구현됨 (Maven/Gradle, CMake)
4. ✅ 25개 이상의 테스트 케이스가 작성되고 모두 통과함
5. ✅ 기존 4개 언어 워크플로우가 영향을 받지 않음 (하위 호환성 검증)
6. ✅ 문서가 업데이트됨 (README.md, CHANGELOG.md)
7. ✅ CI/CD 파이프라인에서 새 워크플로우가 정상 실행됨

---

**@TAG 체인**:
- `@SPEC:LANGUAGE-DETECTION-EXTENDED-001` (이 문서)
- `@TEST:LANGUAGE-DETECTION-EXTENDED-001` (테스트 작성 시)
- `@CODE:LANGUAGE-DETECTION-EXTENDED-001` (구현 시)
- `@DOC:LANGUAGE-DETECTION-EXTENDED-001` (문서 업데이트 시)
status: completed
---

**다음 단계**: `/alfred:2-run SPEC-LANGUAGE-DETECTION-EXTENDED-001`로 구현을 시작하세요.
