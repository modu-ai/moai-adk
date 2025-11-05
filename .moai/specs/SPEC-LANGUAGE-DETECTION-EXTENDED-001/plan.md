# @PLAN:LANGUAGE-DETECTION-EXTENDED-001: 11개 언어 전담 CI/CD 워크플로우 확장 구현 계획

---
spec_id: LANGUAGE-DETECTION-EXTENDED-001
version: 0.0.1
created: 2025-10-30
updated: 2025-10-30
author: @GoosLab
---

## 개요

이 구현 계획은 MoAI-ADK가 11개 추가 언어(Ruby, PHP, Java, Rust, Dart, Swift, Kotlin, C#, C, C++, Shell)를 자동 감지하고 각 언어별 전담 CI/CD 워크플로우를 제공하기 위한 로드맵입니다.

**목표**: 기존 4개 언어(Python, JS, TS, Go) 지원을 15개 언어로 확장하여 MoAI-ADK의 범용성을 극대화합니다.

---

## 구현 전략

### 핵심 원칙

1. **하위 호환성 우선**: 기존 4개 언어 워크플로우에 영향 없음
2. **단계적 구현**: 언어 그룹별 점진적 롤아웃
3. **테스트 주도 개발**: 각 언어별 테스트 먼저 작성
4. **자동화 우선**: 빌드 도구, 테스트 프레임워크 자동 감지

### 구현 순서

**Phase 1**: 워크플로우 템플릿 생성 (우선순위 높음)
**Phase 2**: LanguageDetector 확장 (우선순위 높음)
**Phase 3**: 빌드 도구 감지 로직 (우선순위 중간)
**Phase 4**: 포괄적 테스트 작성 (우선순위 높음)
**Phase 5**: 문서 업데이트 (우선순위 중간)
**Phase 6**: 통합 테스트 및 검증 (우선순위 중간)

---

## 상세 구현 단계

### Phase 1: 11개 언어별 워크플로우 템플릿 생성

**목표**: 각 언어별 최적화된 GitHub Actions 워크플로우 YAML 파일 생성

**작업 항목**:

#### 1.1 Ruby 워크플로우 (`ruby-tag-validation.yml`)

**위치**: `src/moai_adk/templates/.github/workflows/ruby-tag-validation.yml`

**포함 도구**:
- RSpec (테스트 프레임워크)
- Rubocop (린터)
- bundle (의존성 관리)

**핵심 설정**:
```yaml
- Ruby 버전: 3.2 (기본), 사용자 설정 가능
- Bundler 캐싱 활성화
- RSpec 커버리지 리포트 생성
```

#### 1.2 PHP 워크플로우 (`php-tag-validation.yml`)

**포함 도구**:
- PHPUnit (테스트 프레임워크)
- PHPCS (PHP CodeSniffer - 코드 스타일)
- composer (의존성 관리)

**핵심 설정**:
```yaml
- PHP 버전: 8.2 (기본)
- composer 캐싱
- PSR-12 코드 스타일 검증
```

#### 1.3 Java 워크플로우 (`java-tag-validation.yml`)

**포함 도구**:
- JUnit 5 (테스트 프레임워크)
- Jacoco (커버리지 도구)
- Maven 또는 Gradle (빌드 도구 자동 감지)

**핵심 설정**:
```yaml
- Java 버전: 17 (LTS 기본)
- Maven/Gradle 자동 감지 로직
- 커버리지 80% 임계값
```

#### 1.4 Rust 워크플로우 (`rust-tag-validation.yml`)

**포함 도구**:
- cargo test (테스트)
- clippy (린터)
- rustfmt (포매터)

**핵심 설정**:
```yaml
- Rust 툴체인: stable
- clippy 경고를 에러로 처리 (-D warnings)
- 포매팅 검증 자동화
```

#### 1.5 Dart/Flutter 워크플로우 (`dart-tag-validation.yml`)

**포함 도구**:
- dart test 또는 flutter test (테스트)
- dart analyze (정적 분석)

**핵심 설정**:
```yaml
- Dart SDK: stable
- Flutter 프로젝트 자동 감지
- Widget 테스트 지원
```

#### 1.6 Swift 워크플로우 (`swift-tag-validation.yml`)

**포함 도구**:
- XCTest (테스트 프레임워크)
- SwiftLint (린터)

**핵심 설정**:
```yaml
- Runner: macos-latest (Swift 빌드 요구사항)
- Swift Package Manager 지원
- Xcode 프로젝트 지원
```

#### 1.7 Kotlin 워크플로우 (`kotlin-tag-validation.yml`)

**포함 도구**:
- JUnit 5 (테스트)
- ktlint (린터)
- Gradle (빌드 도구)

**핵심 설정**:
```yaml
- Kotlin 컴파일러 버전: 최신 stable
- Gradle wrapper 사용
- Java 17 기반
```

#### 1.8 C# 워크플로우 (`csharp-tag-validation.yml`)

**포함 도구**:
- xUnit (테스트 프레임워크)
- StyleCop (린터)
- dotnet CLI (빌드)

**핵심 설정**:
```yaml
- .NET SDK: 8.0.x (최신 LTS)
- NuGet 패키지 캐싱
- StyleCop 규칙 강제
```

#### 1.9 C 워크플로우 (`c-tag-validation.yml`)

**포함 도구**:
- gcc/clang (컴파일러)
- cppcheck (정적 분석)
- Unity (테스트 프레임워크) 또는 ctest

**핵심 설정**:
```yaml
- CMake 빌드 시스템
- 메모리 누수 검사 (Valgrind 옵션)
- 경고를 에러로 처리 (-Werror)
```

#### 1.10 C++ 워크플로우 (`cpp-tag-validation.yml`)

**포함 도구**:
- g++/clang++ (컴파일러)
- Google Test (테스트 프레임워크)
- cpplint (린터)

**핵심 설정**:
```yaml
- C++ 표준: C++17 (기본)
- CMake + Google Test 통합
- 메모리 안전성 검사
```

#### 1.11 Shell 워크플로우 (`shell-tag-validation.yml`)

**포함 도구**:
- shellcheck (정적 분석)
- bats-core (테스트 프레임워크)

**핵심 설정**:
```yaml
- 모든 .sh 파일 검증
- POSIX 호환성 검사
- 베스트 프랙티스 강제
```

---

### Phase 2: LanguageDetector 클래스 확장

**목표**: 11개 언어 자동 감지 로직 추가

**파일**: `src/moai_adk/core/language_detector.py`

**작업 항목**:

#### 2.1 언어 감지 메서드 추가

각 언어별 `detect_{language}()` 메서드 작성:

```python
# Ruby 감지
def detect_ruby(self) -> bool:
    """Gemfile 존재 여부로 Ruby 프로젝트 감지"""
    return (self.project_root / "Gemfile").exists()

# PHP 감지
def detect_php(self) -> bool:
    """composer.json 존재 여부로 PHP 프로젝트 감지"""
    return (self.project_root / "composer.json").exists()

# Java 감지 (Maven/Gradle 모두 지원)
def detect_java(self) -> bool:
    """pom.xml 또는 build.gradle 존재 여부로 Java 감지"""
    return (
        (self.project_root / "pom.xml").exists() or
        (self.project_root / "build.gradle").exists() or
        (self.project_root / "build.gradle.kts").exists()
    )

# ... (나머지 8개 언어 동일 패턴)
```

#### 2.2 우선순위 로직 업데이트

`detect()` 메서드에 확장된 우선순위 체인 적용:

```python
def detect(self) -> str:
    """우선순위에 따라 프로젝트 언어 감지"""
    if self.detect_rust():
        return "rust"
    if self.detect_dart():
        return "dart"
    if self.detect_swift():
        return "swift"
    if self.detect_kotlin():
        return "kotlin"
    if self.detect_csharp():
        return "csharp"
    if self.detect_java():
        return "java"
    if self.detect_ruby():
        return "ruby"
    if self.detect_php():
        return "php"
    if self.detect_go():  # 기존
        return "go"
    if self.detect_python():  # 기존
        return "python"
    if self.detect_typescript():  # 기존
        return "typescript"
    if self.detect_javascript():  # 기존
        return "javascript"
    if self.detect_cpp():
        return "cpp"
    if self.detect_c():
        return "c"
    if self.detect_shell():
        return "shell"
    return "unknown"
```

#### 2.3 다중 언어 충돌 해결

**시나리오**: Java + Kotlin 동시 존재 시
**해결 방안**: Kotlin 우선 (최신 JVM 언어)

```python
# 우선순위 규칙:
# 1. Kotlin (build.gradle.kts 존재) → kotlin
# 2. Java (pom.xml 또는 build.gradle) → java
```

---

### Phase 3: 빌드 도구 감지 로직

**목표**: 언어별 빌드 도구 자동 선택

**작업 항목**:

#### 3.1 빌드 도구 감지 메서드 추가

```python
def detect_build_tool(self, language: str) -> str:
    """언어별 빌드 도구 자동 감지"""
    if language == "java":
        if (self.project_root / "pom.xml").exists():
            return "maven"
        elif (self.project_root / "build.gradle").exists():
            return "gradle"
        elif (self.project_root / "build.gradle.kts").exists():
            return "gradle"

    elif language in ["c", "cpp"]:
        if (self.project_root / "CMakeLists.txt").exists():
            return "cmake"
        # 향후 Bazel, Meson 지원 가능

    elif language == "swift":
        if (self.project_root / "Package.swift").exists():
            return "spm"  # Swift Package Manager
        elif any(self.project_root.glob("*.xcodeproj")):
            return "xcode"

    return "default"
```

#### 3.2 테스트 프레임워크 감지

```python
def get_test_framework(self, language: str) -> str:
    """언어별 테스트 프레임워크 반환"""
    frameworks = {
        "ruby": "rspec",
        "php": "phpunit",
        "java": "junit5",
        "rust": "cargo_test",
        "dart": "dart_test",
        "swift": "xctest",
        "kotlin": "junit5",
        "csharp": "xunit",
        "c": "unity",
        "cpp": "gtest",
        "shell": "bats",
        # 기존 언어들
        "python": "pytest",
        "javascript": "jest",
        "typescript": "jest",
        "go": "go_test",
    }
    return frameworks.get(language, "unknown")
```

---

### Phase 4: 포괄적 테스트 작성

**목표**: 11개 언어 감지 및 빌드 도구 선택 로직 검증

**파일**: `tests/test_language_detector_extended.py`

**테스트 카테고리**:

#### 4.1 언어별 감지 테스트 (11개)

```python
def test_detect_ruby(tmp_path):
    """Ruby 프로젝트 감지 테스트"""
    (tmp_path / "Gemfile").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "ruby"

def test_detect_php(tmp_path):
    """PHP 프로젝트 감지 테스트"""
    (tmp_path / "composer.json").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "php"

# ... (나머지 9개 언어 동일 패턴)
```

#### 4.2 우선순위 충돌 해결 테스트 (5개)

```python
def test_kotlin_over_java(tmp_path):
    """Kotlin과 Java 동시 존재 시 Kotlin 우선"""
    (tmp_path / "build.gradle.kts").touch()
    (tmp_path / "pom.xml").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "kotlin"

def test_cpp_over_c(tmp_path):
    """C++와 C 동시 존재 시 C++ 우선"""
    (tmp_path / "CMakeLists.txt").touch()
    (tmp_path / "main.cpp").touch()
    (tmp_path / "utils.c").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "cpp"
```

#### 4.3 빌드 도구 선택 테스트 (3개)

```python
def test_maven_detection(tmp_path):
    """Maven 빌드 도구 감지"""
    (tmp_path / "pom.xml").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("java") == "maven"

def test_gradle_detection(tmp_path):
    """Gradle 빌드 도구 감지"""
    (tmp_path / "build.gradle").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("java") == "gradle"
```

#### 4.4 오류 처리 테스트 (3개)

```python
def test_unknown_language(tmp_path):
    """알 수 없는 언어 처리"""
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "unknown"

def test_missing_build_tool(tmp_path):
    """빌드 도구 누락 시 에러 처리"""
    (tmp_path / "pom.xml").touch()
    detector = LanguageDetector(tmp_path)
    # Maven이 없으면 명확한 에러 메시지
    with pytest.raises(BuildToolNotFoundError):
        detector.validate_build_tool("java")
```

#### 4.5 하위 호환성 테스트 (4개)

```python
def test_existing_python_support(tmp_path):
    """기존 Python 지원 유지 검증"""
    (tmp_path / "pyproject.toml").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "python"

# JavaScript, TypeScript, Go 동일 패턴
```

**총 테스트 수**: 26개

---

### Phase 5: 문서 업데이트

**목표**: 사용자 가이드 및 변경 이력 업데이트

**작업 항목**:

#### 5.1 README.md 업데이트

**변경 사항**:
- 지원 언어 목록 확장 (4개 → 15개)
- 언어별 워크플로우 예시 추가
- 빌드 도구 자동 감지 설명

**추가 섹션**:
```markdown
## 지원 언어 (15개)

### 컴파일 언어
- **시스템 프로그래밍**: C, C++, Rust
- **JVM 계열**: Java, Kotlin
- **네이티브**: Swift, C#

### 스크립트 언어
- **웹 백엔드**: PHP, Ruby
- **범용**: Python, JavaScript, TypeScript
- **모바일/웹**: Dart (Flutter)

### 시스템 스크립트
- Go, Shell

## 빌드 도구 자동 감지
- **Java**: Maven (`pom.xml`) ↔ Gradle (`build.gradle`)
- **C/C++**: CMake (`CMakeLists.txt`)
- **Swift**: SPM (`Package.swift`) ↔ Xcode (`.xcodeproj`)
```

#### 5.2 새 가이드 문서 생성

**파일**: `.moai/docs/language-detection-extended.md`

**내용**:
- 11개 언어별 상세 설명
- 각 언어의 설정 파일 감지 규칙
- 우선순위 규칙 설명
- 빌드 도구 선택 로직
- 트러블슈팅 가이드

#### 5.3 CHANGELOG.md 업데이트

**v0.10.3 릴리스 노트 초안**:

```markdown
## [0.10.3] - 2025-10-30

### 🚀 Added
- 11개 언어 전담 CI/CD 워크플로우 지원 추가:
  - Ruby (RSpec, Rubocop)
  - PHP (PHPUnit, PHPCS)
  - Java (JUnit 5, Jacoco, Maven/Gradle)
  - Rust (cargo test, clippy)
  - Dart (flutter test, dart analyze)
  - Swift (XCTest, SwiftLint)
  - Kotlin (JUnit 5, ktlint)
  - C# (xUnit, StyleCop)
  - C (gcc/clang, cppcheck)
  - C++ (g++/clang++, gtest)
  - Shell (shellcheck, bats-core)
- 빌드 도구 자동 감지 기능 (Maven vs Gradle, CMake)
- 언어 우선순위 충돌 해결 로직

### 📚 Documentation
- `.moai/docs/language-detection-extended.md` 가이드 추가
- README.md 지원 언어 섹션 확장

### 🧪 Tests
- 26개 신규 테스트 케이스 추가
- 하위 호환성 테스트 (기존 4개 언어)

### 📦 Dependencies
- 언어별 도구 설치 가이드 추가

### 🔗 Related Issues
- Implements #131 (11개 언어 확장)
- Depends on SPEC-LANGUAGE-DETECTION-001
```

---

### Phase 6: 통합 테스트 및 검증

**목표**: 실제 프로젝트 템플릿으로 워크플로우 검증

**작업 항목**:

#### 6.1 샘플 프로젝트 생성 (11개)

각 언어별 최소 기능 프로젝트 템플릿 생성:

```
tests/fixtures/sample_projects/
├── ruby_app/
│   ├── Gemfile
│   ├── lib/calculator.rb
│   └── spec/calculator_spec.rb
├── php_app/
│   ├── composer.json
│   ├── src/Calculator.php
│   └── tests/CalculatorTest.php
├── java_maven_app/
│   ├── pom.xml
│   └── src/main/java/Calculator.java
├── java_gradle_app/
│   ├── build.gradle
│   └── src/main/java/Calculator.java
... (나머지 7개 언어)
```

#### 6.2 E2E 테스트

```python
@pytest.mark.integration
def test_ruby_workflow_execution(tmp_path):
    """Ruby 워크플로우 E2E 테스트"""
    # 1. 샘플 Ruby 프로젝트 복사
    shutil.copytree("tests/fixtures/sample_projects/ruby_app", tmp_path)

    # 2. LanguageDetector로 언어 감지
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "ruby"

    # 3. 워크플로우 파일 생성
    workflow_generator = WorkflowGenerator(tmp_path)
    workflow_generator.generate()

    # 4. 워크플로우 파일 존재 확인
    workflow_path = tmp_path / ".github/workflows/ruby-tag-validation.yml"
    assert workflow_path.exists()

    # 5. YAML 유효성 검증
    with open(workflow_path) as f:
        workflow_config = yaml.safe_load(f)
    assert "ruby/setup-ruby" in str(workflow_config)
```

#### 6.3 CI/CD 파이프라인 검증

**검증 항목**:
- GitHub Actions runner에서 11개 워크플로우 실행 성공
- 각 언어별 빌드, 테스트, 린팅 통과
- 커버리지 리포트 생성 확인
- 실패 시나리오 처리 확인

#### 6.4 하위 호환성 회귀 테스트

```python
@pytest.mark.regression
def test_existing_languages_unaffected():
    """기존 4개 언어에 영향 없음 확인"""
    for lang in ["python", "javascript", "typescript", "go"]:
        # 기존 워크플로우가 정상 작동하는지 검증
        assert validate_workflow(lang) == "success"
```

---

## 기술 스택 및 도구

### 언어별 도구 버전

| 언어       | 런타임 버전        | 테스트 프레임워크 | 린터/포매터          | 빌드 도구           |
|------------|-------------------|------------------|---------------------|-------------------|
| Ruby       | 3.2.x (stable)    | RSpec 3.12+      | Rubocop 1.50+       | bundle            |
| PHP        | 8.2.x (stable)    | PHPUnit 10.5+    | PHPCS 3.7+          | composer 2.6+     |
| Java       | 17 LTS            | JUnit 5.10+      | Checkstyle 10.12+   | Maven 3.9+ / Gradle 8.5+ |
| Rust       | stable (1.75+)    | cargo test       | clippy, rustfmt     | cargo             |
| Dart       | 3.2.x (stable)    | dart test        | dart analyze        | dart/flutter      |
| Swift      | 5.9+              | XCTest           | SwiftLint 0.54+     | SPM / Xcode       |
| Kotlin     | 1.9+              | JUnit 5.10+      | ktlint 12.0+        | Gradle 8.5+       |
| C#         | .NET 8.0 LTS      | xUnit 2.6+       | StyleCop 1.2+       | dotnet CLI        |
| C          | gcc 12+ / clang 15+ | Unity / ctest  | cppcheck 2.12+      | cmake 3.27+       |
| C++        | g++ 12+ / clang++ 15+ | gtest 1.14+  | cpplint 1.6+        | cmake 3.27+       |
| Shell      | bash 5.2+         | bats-core 1.10+  | shellcheck 0.9+     | N/A               |

### GitHub Actions 의존성

```yaml
# 공통 Actions
- actions/checkout@v4
- actions/cache@v4

# 언어별 Setup Actions
- ruby/setup-ruby@v1
- shivammathur/setup-php@v2
- actions/setup-java@v4
- dtolnay/rust-toolchain@stable
- dart-lang/setup-dart@v1
- subosito/flutter-action@v2
- actions/setup-dotnet@v4
```

---

## 테스트 전략

### 테스트 피라미드

```
        /\
       /  \  E2E (11개) - 각 언어별 샘플 프로젝트
      /____\
     /      \  Integration (10개) - 빌드 도구 감지, 우선순위
    /________\
   /          \  Unit (26개) - 언어 감지, 메서드 검증
  /__________\
```

### 커버리지 목표

- **Unit 테스트**: 95% 이상
- **Integration 테스트**: 85% 이상
- **E2E 테스트**: 주요 언어 100% (11개 모두)

### 테스트 실행 전략

```bash
# 빠른 검증 (Unit + Integration)
pytest tests/test_language_detector_extended.py -v

# 전체 E2E 테스트
pytest tests/integration/ --run-e2e -v

# 하위 호환성 회귀 테스트
pytest tests/ -m regression -v

# 특정 언어 테스트
pytest tests/ -k "ruby or php" -v
```

---

## 배포 및 통합 계획

### 배포 전략

1. **Feature Branch 생성**: `feature/language-detection-extended`
2. **Phase별 커밋**: 각 Phase 완료 시 커밋 (RED → GREEN → REFACTOR)
3. **Draft PR 생성**: 초기 리뷰 수집
4. **CI/CD 검증**: 모든 테스트 통과 확인
5. **문서 리뷰**: README, CHANGELOG 검토
6. **PR Merge**: `develop` 브랜치로 병합
7. **Release**: v0.10.3 태그 생성

### 롤백 계획

**만약 새 언어 지원에 문제가 발생하면**:

1. 기존 4개 언어 워크플로우는 영향받지 않음 (독립적 템플릿)
2. 특정 언어 워크플로우만 비활성화 가능
3. `LanguageDetector.detect()` 메서드에서 문제 언어 스킵

**롤백 코드 예시**:
```python
# 긴급 롤백: Java 지원 임시 비활성화
def detect_java(self) -> bool:
    return False  # 임시 비활성화
```

---

## 위험 요소 및 대응 방안

### 위험 1: 빌드 도구 충돌

**시나리오**: Maven과 Gradle이 동시에 존재하는 Java 프로젝트

**대응 방안**:
- 우선순위 규칙: Maven > Gradle (보수적 선택)
- 사용자 설정 파일 (`.moai/config.json`)에서 명시적 선택 지원
- 경고 메시지 출력

### 위험 2: GitHub Actions Runner 제약

**시나리오**: Swift/Xcode는 macOS runner 필수 (비용 발생)

**대응 방안**:
- 문서에 명시: "Swift 워크플로우는 macOS runner 필요"
- 사용자 선택권 부여: Linux에서 Swift Package Manager만 사용 가능
- 조건부 워크플로우: `if: runner.os == 'macOS'`

### 위험 3: 테스트 프레임워크 버전 불일치

**시나리오**: 사용자의 로컬 도구 버전 ≠ CI 도구 버전

**대응 방안**:
- 워크플로우에 버전 명시 (pinning)
- 사용자 `.moai/config.json`에서 버전 오버라이드 허용
- 경고 메시지: "로컬 버전과 CI 버전이 다를 수 있음"

### 위험 4: 하위 호환성 파괴

**시나리오**: 기존 Python/JS/TS/Go 워크플로우 오작동

**대응 방안**:
- 회귀 테스트 필수 실행 (Phase 6.4)
- 기존 워크플로우 템플릿 수정 금지
- 별도 템플릿 파일로 분리 유지

---

## 성공 기준

이 구현이 성공했다고 판단하는 기준:

### 기능적 성공 기준

1. ✅ 11개 언어 워크플로우 템플릿이 모두 GitHub Actions에서 실행됨
2. ✅ `LanguageDetector`가 15개 언어 모두 정확히 감지함
3. ✅ 빌드 도구 자동 선택이 3개 이상 언어에서 작동함
4. ✅ 26개 테스트가 모두 통과함
5. ✅ 기존 4개 언어에 회귀 버그 없음

### 품질 기준

1. ✅ 테스트 커버리지 95% 이상
2. ✅ 린팅 0 에러
3. ✅ Type checking 통과 (Python 3.11+)
4. ✅ 문서 리뷰 승인

### 사용자 경험 기준

1. ✅ 사용자가 추가 설정 없이 11개 언어 지원 활성화됨
2. ✅ 에러 메시지가 명확하고 실행 가능한 해결책 제공
3. ✅ 문서가 각 언어별 예시를 포함함

---

## 다음 단계

이 계획 완료 후:

1. **v0.10.3 릴리스** 배포
2. **사용자 피드백 수집** (GitHub Discussions)
3. **다음 SPEC 고려**:
   - `SPEC-MULTI-LANGUAGE-001`: Monorepo 다중 언어 지원
   - `SPEC-CUSTOM-WORKFLOW-001`: 사용자 정의 워크플로우 템플릿
   - `SPEC-LANGUAGE-VERSION-001`: 언어 버전 자동 감지 및 업그레이드

---

**@TAG 체인**:
- `@PLAN:LANGUAGE-DETECTION-EXTENDED-001` (이 문서)
- `@SPEC:LANGUAGE-DETECTION-EXTENDED-001` (요구사항 문서)
- `@TEST:LANGUAGE-DETECTION-EXTENDED-001` (구현 시 생성)
- `@CODE:LANGUAGE-DETECTION-EXTENDED-001` (구현 시 생성)

---

**구현 시작 명령**: `/alfred:2-run SPEC-LANGUAGE-DETECTION-EXTENDED-001`
