# @ACCEPTANCE:LANGUAGE-DETECTION-EXTENDED-001: 11개 언어 전담 CI/CD 워크플로우 확장 인수 기준

---
spec_id: LANGUAGE-DETECTION-EXTENDED-001
version: 0.0.1
created: 2025-10-30
updated: 2025-10-30
author: @GoosLab
---

## 개요

이 문서는 SPEC-LANGUAGE-DETECTION-EXTENDED-001의 인수 기준(Acceptance Criteria)을 정의합니다. 모든 시나리오는 Given-When-Then 포맷으로 작성되며, 구현 완료 시 모든 시나리오가 통과해야 합니다.

---

## 인수 기준 카테고리

### 1. 언어 감지 시나리오 (11개)
### 2. 빌드 도구 선택 시나리오 (5개)
### 3. 우선순위 충돌 해결 시나리오 (4개)
### 4. 오류 처리 시나리오 (3개)
### 5. 하위 호환성 시나리오 (4개)
### 6. 통합 시나리오 (3개)

**총 시나리오 수**: 30개

---

## 1. 언어 감지 시나리오

### Scenario 1.1: Ruby 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 Gemfile이 존재함
When: LanguageDetector.detect()를 호출함
Then: "ruby"가 반환됨
And: ruby-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_ruby(tmp_path):
    (tmp_path / "Gemfile").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "ruby"
```

**예상 워크플로우 내용**:
- RSpec 테스트 실행
- Rubocop 린팅
- bundle install 의존성 설치

---

### Scenario 1.2: PHP 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 composer.json이 존재함
When: LanguageDetector.detect()를 호출함
Then: "php"가 반환됨
And: php-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_php(tmp_path):
    (tmp_path / "composer.json").write_text('{"name": "test/app"}')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "php"
```

**예상 워크플로우 내용**:
- PHPUnit 테스트 실행
- PHPCS 코드 스타일 검증
- composer install 실행

---

### Scenario 1.3: Java 프로젝트 감지 (Maven)

```gherkin
Given: 프로젝트 루트에 pom.xml이 존재함
When: LanguageDetector.detect()를 호출함
Then: "java"가 반환됨
And: java-tag-validation.yml 워크플로우가 선택됨
And: 빌드 도구로 "maven"이 감지됨
```

**검증 방법**:
```python
def test_detect_java_maven(tmp_path):
    (tmp_path / "pom.xml").write_text('<project></project>')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "java"
    assert detector.detect_build_tool("java") == "maven"
```

**예상 워크플로우 내용**:
- `mvn test` 실행
- JUnit 5 테스트
- Jacoco 커버리지 리포트

---

### Scenario 1.4: Java 프로젝트 감지 (Gradle)

```gherkin
Given: 프로젝트 루트에 build.gradle 또는 build.gradle.kts가 존재함
When: LanguageDetector.detect()를 호출함
Then: "java"가 반환됨
And: java-tag-validation.yml 워크플로우가 선택됨
And: 빌드 도구로 "gradle"이 감지됨
```

**검증 방법**:
```python
def test_detect_java_gradle(tmp_path):
    (tmp_path / "build.gradle").write_text('plugins { id "java" }')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "java"
    assert detector.detect_build_tool("java") == "gradle"
```

**예상 워크플로우 내용**:
- `./gradlew test` 실행
- JUnit 5 테스트
- Jacoco 플러그인 사용

---

### Scenario 1.5: Rust 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 Cargo.toml이 존재함
When: LanguageDetector.detect()를 호출함
Then: "rust"가 반환됨
And: rust-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_rust(tmp_path):
    (tmp_path / "Cargo.toml").write_text('[package]\nname = "test"')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "rust"
```

**예상 워크플로우 내용**:
- `cargo test` 실행
- `cargo clippy` 린팅 (경고를 에러로 처리)
- `cargo fmt --check` 포매팅 검증

---

### Scenario 1.6: Dart/Flutter 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 pubspec.yaml이 존재함
When: LanguageDetector.detect()를 호출함
Then: "dart"가 반환됨
And: dart-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_dart(tmp_path):
    (tmp_path / "pubspec.yaml").write_text('name: test_app')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "dart"
```

**예상 워크플로우 내용**:
- `dart analyze` 정적 분석
- `dart test` 또는 `flutter test` 실행
- Flutter 프로젝트 자동 감지 (sdk: flutter 키워드)

---

### Scenario 1.7: Swift 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 Package.swift 또는 .xcodeproj가 존재함
When: LanguageDetector.detect()를 호출함
Then: "swift"가 반환됨
And: swift-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_swift_spm(tmp_path):
    (tmp_path / "Package.swift").write_text('// swift-tools-version:5.9')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "swift"

def test_detect_swift_xcode(tmp_path):
    (tmp_path / "MyApp.xcodeproj").mkdir()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "swift"
```

**예상 워크플로우 내용**:
- macOS runner 사용
- `swift build` 실행
- `swift test` XCTest 실행
- SwiftLint 린팅

---

### Scenario 1.8: Kotlin 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 build.gradle.kts가 존재함
When: LanguageDetector.detect()를 호출함
Then: "kotlin"이 반환됨
And: kotlin-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_kotlin(tmp_path):
    (tmp_path / "build.gradle.kts").write_text('plugins { kotlin("jvm") }')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "kotlin"
```

**예상 워크플로우 내용**:
- `./gradlew test` 실행
- JUnit 5 테스트
- ktlint 린팅

---

### Scenario 1.9: C# 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 .csproj 또는 .sln 파일이 존재함
When: LanguageDetector.detect()를 호출함
Then: "csharp"가 반환됨
And: csharp-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_csharp_csproj(tmp_path):
    (tmp_path / "MyApp.csproj").write_text('<Project Sdk="Microsoft.NET.Sdk"></Project>')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "csharp"

def test_detect_csharp_sln(tmp_path):
    (tmp_path / "MySolution.sln").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "csharp"
```

**예상 워크플로우 내용**:
- `dotnet restore` 의존성 복원
- `dotnet build` 빌드
- `dotnet test` xUnit 테스트 실행

---

### Scenario 1.10: C 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 CMakeLists.txt와 .c 파일이 존재함
And: .cpp 파일은 존재하지 않음
When: LanguageDetector.detect()를 호출함
Then: "c"가 반환됨
And: c-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_c(tmp_path):
    (tmp_path / "CMakeLists.txt").write_text('project(MyApp C)')
    (tmp_path / "main.c").write_text('int main() { return 0; }')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "c"
```

**예상 워크플로우 내용**:
- `cmake . && make` 빌드
- `ctest` 테스트 실행
- `cppcheck` 정적 분석

---

### Scenario 1.11: C++ 프로젝트 감지

```gherkin
Given: 프로젝트 루트에 CMakeLists.txt와 .cpp 파일이 존재함
When: LanguageDetector.detect()를 호출함
Then: "cpp"가 반환됨
And: cpp-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_cpp(tmp_path):
    (tmp_path / "CMakeLists.txt").write_text('project(MyApp CXX)')
    (tmp_path / "main.cpp").write_text('int main() { return 0; }')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "cpp"
```

**예상 워크플로우 내용**:
- `cmake . && make` 빌드
- Google Test 실행
- `cpplint` 린팅

---

### Scenario 1.12: Shell 스크립트 감지

```gherkin
Given: 프로젝트에 .sh 파일이 존재함
When: LanguageDetector.detect()를 호출함
Then: "shell"이 반환됨
And: shell-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_detect_shell(tmp_path):
    (tmp_path / "deploy.sh").write_text('#!/bin/bash\necho "Hello"')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "shell"
```

**예상 워크플로우 내용**:
- `shellcheck` 모든 .sh 파일 검증
- `bats` 테스트 실행

---

## 2. 빌드 도구 선택 시나리오

### Scenario 2.1: Maven 빌드 도구 선택

```gherkin
Given: Java 프로젝트에 pom.xml이 존재함
And: build.gradle은 존재하지 않음
When: detect_build_tool("java")를 호출함
Then: "maven"이 반환됨
```

**검증 방법**:
```python
def test_build_tool_maven_only(tmp_path):
    (tmp_path / "pom.xml").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("java") == "maven"
```

---

### Scenario 2.2: Gradle 빌드 도구 선택

```gherkin
Given: Java 프로젝트에 build.gradle 또는 build.gradle.kts가 존재함
And: pom.xml은 존재하지 않음
When: detect_build_tool("java")를 호출함
Then: "gradle"이 반환됨
```

**검증 방법**:
```python
def test_build_tool_gradle_only(tmp_path):
    (tmp_path / "build.gradle").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("java") == "gradle"
```

---

### Scenario 2.3: Maven과 Gradle 동시 존재 시 우선순위

```gherkin
Given: Java 프로젝트에 pom.xml과 build.gradle이 모두 존재함
When: detect_build_tool("java")를 호출함
Then: "maven"이 반환됨 (Maven 우선순위 높음)
And: 경고 메시지가 출력됨: "Multiple build tools detected"
```

**검증 방법**:
```python
def test_build_tool_conflict_maven_wins(tmp_path, caplog):
    (tmp_path / "pom.xml").touch()
    (tmp_path / "build.gradle").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("java") == "maven"
    assert "Multiple build tools detected" in caplog.text
```

---

### Scenario 2.4: CMake 빌드 도구 선택 (C/C++)

```gherkin
Given: C 또는 C++ 프로젝트에 CMakeLists.txt가 존재함
When: detect_build_tool("c") 또는 detect_build_tool("cpp")를 호출함
Then: "cmake"가 반환됨
```

**검증 방법**:
```python
def test_build_tool_cmake_c(tmp_path):
    (tmp_path / "CMakeLists.txt").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("c") == "cmake"

def test_build_tool_cmake_cpp(tmp_path):
    (tmp_path / "CMakeLists.txt").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("cpp") == "cmake"
```

---

### Scenario 2.5: Swift 빌드 도구 선택 (SPM vs Xcode)

```gherkin
Given: Swift 프로젝트에 Package.swift가 존재함
When: detect_build_tool("swift")를 호출함
Then: "spm"이 반환됨 (Swift Package Manager)

Given: Swift 프로젝트에 .xcodeproj가 존재함
And: Package.swift는 존재하지 않음
When: detect_build_tool("swift")를 호출함
Then: "xcode"가 반환됨
```

**검증 방법**:
```python
def test_build_tool_spm(tmp_path):
    (tmp_path / "Package.swift").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("swift") == "spm"

def test_build_tool_xcode(tmp_path):
    (tmp_path / "MyApp.xcodeproj").mkdir()
    detector = LanguageDetector(tmp_path)
    assert detector.detect_build_tool("swift") == "xcode"
```

---

## 3. 우선순위 충돌 해결 시나리오

### Scenario 3.1: Kotlin vs Java 우선순위

```gherkin
Given: 프로젝트에 build.gradle.kts (Kotlin)와 pom.xml (Java)이 모두 존재함
When: LanguageDetector.detect()를 호출함
Then: "kotlin"이 반환됨 (Kotlin 우선순위 높음)
```

**검증 방법**:
```python
def test_priority_kotlin_over_java(tmp_path):
    (tmp_path / "build.gradle.kts").touch()
    (tmp_path / "pom.xml").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "kotlin"
```

**우선순위 근거**: Kotlin은 최신 JVM 언어이며 Java와 100% 호환 가능

---

### Scenario 3.2: C++ vs C 우선순위

```gherkin
Given: 프로젝트에 .cpp 파일과 .c 파일이 모두 존재함
And: CMakeLists.txt가 존재함
When: LanguageDetector.detect()를 호출함
Then: "cpp"가 반환됨 (C++ 우선순위 높음)
```

**검증 방법**:
```python
def test_priority_cpp_over_c(tmp_path):
    (tmp_path / "CMakeLists.txt").touch()
    (tmp_path / "main.cpp").touch()
    (tmp_path / "utils.c").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "cpp"
```

**우선순위 근거**: C++는 C의 상위 집합

---

### Scenario 3.3: Rust vs Go 우선순위

```gherkin
Given: 프로젝트에 Cargo.toml (Rust)과 go.mod (Go)가 모두 존재함
When: LanguageDetector.detect()를 호출함
Then: "rust"가 반환됨 (Rust 우선순위 높음)
```

**검증 방법**:
```python
def test_priority_rust_over_go(tmp_path):
    (tmp_path / "Cargo.toml").touch()
    (tmp_path / "go.mod").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "rust"
```

**우선순위 근거**: 설정 파일 명시도 (Cargo.toml이 더 구체적)

---

### Scenario 3.4: TypeScript vs JavaScript 우선순위 (기존)

```gherkin
Given: 프로젝트에 tsconfig.json과 package.json이 모두 존재함
When: LanguageDetector.detect()를 호출함
Then: "typescript"가 반환됨 (TypeScript 우선순위 높음)
```

**검증 방법**:
```python
def test_priority_typescript_over_javascript(tmp_path):
    (tmp_path / "tsconfig.json").touch()
    (tmp_path / "package.json").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "typescript"
```

**우선순위 근거**: TypeScript는 JavaScript의 슈퍼셋

---

## 4. 오류 처리 시나리오

### Scenario 4.1: 알 수 없는 언어 처리

```gherkin
Given: 프로젝트에 알려진 언어 설정 파일이 하나도 존재하지 않음
When: LanguageDetector.detect()를 호출함
Then: "unknown"이 반환됨
And: 명확한 에러 메시지가 출력됨: "No supported language detected"
```

**검증 방법**:
```python
def test_unknown_language(tmp_path):
    detector = LanguageDetector(tmp_path)
    result = detector.detect()
    assert result == "unknown"
```

---

### Scenario 4.2: 워크플로우 템플릿 누락 처리

```gherkin
Given: 언어는 감지되었으나 해당 워크플로우 템플릿 파일이 존재하지 않음
When: WorkflowGenerator.generate()를 호출함
Then: FileNotFoundError가 발생함
And: 에러 메시지: "Workflow template not found for language: {language}"
```

**검증 방법**:
```python
def test_missing_workflow_template(tmp_path, monkeypatch):
    # 템플릿 디렉토리를 빈 디렉토리로 설정
    monkeypatch.setattr("moai_adk.TEMPLATES_DIR", tmp_path)

    detector = LanguageDetector(tmp_path)
    with pytest.raises(FileNotFoundError, match="Workflow template not found"):
        generator = WorkflowGenerator(detector)
        generator.generate()
```

---

### Scenario 4.3: 필수 빌드 도구 누락 경고

```gherkin
Given: Java 프로젝트가 감지되었으나 Maven/Gradle 실행 파일이 시스템에 없음
When: 워크플로우가 실행됨
Then: GitHub Actions에서 빌드 도구 설치 단계가 자동 실행됨
And: 설치 실패 시 명확한 에러 메시지 출력
```

**검증 방법**:
```python
def test_build_tool_installation_required(tmp_path):
    """워크플로우 YAML에 빌드 도구 설치 단계 포함 확인"""
    (tmp_path / "pom.xml").touch()
    detector = LanguageDetector(tmp_path)
    generator = WorkflowGenerator(detector)
    workflow_content = generator.generate_yaml()

    assert "actions/setup-java" in workflow_content
    assert "maven-version" in workflow_content or "gradle-version" in workflow_content
```

---

## 5. 하위 호환성 시나리오

### Scenario 5.1: 기존 Python 지원 유지

```gherkin
Given: Python 프로젝트 (pyproject.toml 존재)
When: LanguageDetector.detect()를 호출함
Then: "python"이 반환됨
And: 기존 python-tag-validation.yml 워크플로우가 선택됨
And: 워크플로우 내용이 변경되지 않음
```

**검증 방법**:
```python
def test_backwards_compatible_python(tmp_path):
    (tmp_path / "pyproject.toml").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "python"

    # 기존 워크플로우 템플릿 무결성 검증
    template_path = Path("src/moai_adk/templates/.github/workflows/python-tag-validation.yml")
    assert template_path.exists()
```

---

### Scenario 5.2: 기존 JavaScript 지원 유지

```gherkin
Given: JavaScript 프로젝트 (package.json 존재, tsconfig.json 없음)
When: LanguageDetector.detect()를 호출함
Then: "javascript"가 반환됨
And: 기존 javascript-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_backwards_compatible_javascript(tmp_path):
    (tmp_path / "package.json").write_text('{"name": "test"}')
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "javascript"
```

---

### Scenario 5.3: 기존 TypeScript 지원 유지

```gherkin
Given: TypeScript 프로젝트 (tsconfig.json 존재)
When: LanguageDetector.detect()를 호출함
Then: "typescript"가 반환됨
And: 기존 typescript-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_backwards_compatible_typescript(tmp_path):
    (tmp_path / "tsconfig.json").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "typescript"
```

---

### Scenario 5.4: 기존 Go 지원 유지

```gherkin
Given: Go 프로젝트 (go.mod 존재)
When: LanguageDetector.detect()를 호출함
Then: "go"가 반환됨
And: 기존 go-tag-validation.yml 워크플로우가 선택됨
```

**검증 방법**:
```python
def test_backwards_compatible_go(tmp_path):
    (tmp_path / "go.mod").touch()
    detector = LanguageDetector(tmp_path)
    assert detector.detect() == "go"
```

---

## 6. 통합 시나리오

### Scenario 6.1: E2E Ruby 프로젝트 워크플로우 생성

```gherkin
Given: Ruby 샘플 프로젝트 (Gemfile, lib/, spec/ 포함)
When: WorkflowGenerator.generate()를 호출함
Then: .github/workflows/ruby-tag-validation.yml 파일이 생성됨
And: 워크플로우 파일이 유효한 YAML 형식임
And: RSpec과 Rubocop 단계가 포함됨
```

**검증 방법**:
```python
def test_e2e_ruby_workflow_generation(tmp_path):
    # 샘플 Ruby 프로젝트 설정
    (tmp_path / "Gemfile").write_text('source "https://rubygems.org"\ngem "rspec"')
    (tmp_path / "lib").mkdir()
    (tmp_path / "spec").mkdir()

    # 워크플로우 생성
    detector = LanguageDetector(tmp_path)
    generator = WorkflowGenerator(detector)
    generator.generate()

    # 검증
    workflow_path = tmp_path / ".github/workflows/ruby-tag-validation.yml"
    assert workflow_path.exists()

    with open(workflow_path) as f:
        workflow = yaml.safe_load(f)

    assert "ruby/setup-ruby" in str(workflow)
    assert "rspec" in str(workflow)
    assert "rubocop" in str(workflow)
```

---

### Scenario 6.2: E2E Java Maven 프로젝트 워크플로우 생성

```gherkin
Given: Java Maven 프로젝트 (pom.xml, src/main/java/, src/test/java/ 포함)
When: WorkflowGenerator.generate()를 호출함
Then: .github/workflows/java-tag-validation.yml 파일이 생성됨
And: Maven 빌드 명령 (mvn test)이 포함됨
And: JUnit 5 테스트 단계가 포함됨
```

**검증 방법**:
```python
def test_e2e_java_maven_workflow_generation(tmp_path):
    # 샘플 Java Maven 프로젝트 설정
    (tmp_path / "pom.xml").write_text('<project><modelVersion>4.0.0</modelVersion></project>')
    (tmp_path / "src/main/java").mkdir(parents=True)
    (tmp_path / "src/test/java").mkdir(parents=True)

    # 워크플로우 생성
    detector = LanguageDetector(tmp_path)
    generator = WorkflowGenerator(detector)
    generator.generate()

    # 검증
    workflow_path = tmp_path / ".github/workflows/java-tag-validation.yml"
    assert workflow_path.exists()

    with open(workflow_path) as f:
        workflow_content = f.read()

    assert "actions/setup-java" in workflow_content
    assert "mvn test" in workflow_content
    assert "junit" in workflow_content.lower()
```

---

### Scenario 6.3: E2E Rust 프로젝트 CI 실행

```gherkin
Given: Rust 샘플 프로젝트 (Cargo.toml, src/, tests/ 포함)
And: GitHub Actions 워크플로우가 생성됨
When: git push로 CI 트리거
Then: GitHub Actions에서 워크플로우가 성공적으로 실행됨
And: cargo test가 통과함
And: clippy 경고가 없음
And: rustfmt 검증 통과
```

**검증 방법**:
```python
@pytest.mark.integration
@pytest.mark.github_actions
def test_e2e_rust_ci_execution():
    """실제 GitHub Actions에서 Rust 워크플로우 실행 테스트"""
    # 이 테스트는 실제 GitHub 저장소에서 실행됨
    # CI 환경에서만 활성화 (로컬에서는 스킵)
    if not os.getenv("GITHUB_ACTIONS"):
        pytest.skip("Requires GitHub Actions environment")

    # GitHub API로 최근 워크플로우 실행 결과 확인
    workflow_run = get_latest_workflow_run("rust-tag-validation")
    assert workflow_run["conclusion"] == "success"
    assert "cargo test" in workflow_run["logs"]
    assert "clippy" in workflow_run["logs"]
```

---

## 최종 인수 체크리스트

이 SPEC의 완료를 위해 다음 모든 항목이 충족되어야 합니다:

### 기능적 완료 기준

- [ ] **11개 언어 워크플로우 템플릿**이 모두 생성되고 정상 작동함
- [ ] **LanguageDetector** 클래스가 15개 언어 모두 정확히 감지함
- [ ] **빌드 도구 자동 선택**이 Java, C/C++, Swift에서 작동함
- [ ] **우선순위 규칙**이 충돌 시나리오에서 정확히 적용됨
- [ ] **오류 처리**가 명확하고 실행 가능한 메시지를 제공함

### 테스트 완료 기준

- [ ] **30개 인수 시나리오**가 모두 통과함
- [ ] **테스트 커버리지** 95% 이상 달성
- [ ] **회귀 테스트** (기존 4개 언어) 모두 통과
- [ ] **E2E 테스트** 최소 3개 언어에서 성공

### 문서 완료 기준

- [ ] **README.md**에 11개 언어 지원 명시
- [ ] **language-detection-extended.md** 가이드 작성
- [ ] **CHANGELOG.md** v0.10.3 릴리스 노트 추가
- [ ] 각 워크플로우 템플릿에 주석 포함

### 품질 완료 기준

- [ ] **린팅** 0 에러 (ruff, mypy)
- [ ] **Type hints** 모든 public 메서드에 추가
- [ ] **Code review** 승인 완료
- [ ] **CI/CD** 모든 파이프라인 통과

---

## 검증 명령어

### 로컬 검증

```bash
# 전체 테스트 실행
pytest tests/test_language_detector_extended.py -v --cov

# 특정 언어 테스트
pytest tests/ -k "ruby or php or java" -v

# 회귀 테스트
pytest tests/ -m regression -v

# E2E 통합 테스트
pytest tests/integration/ --run-e2e -v
```

### CI/CD 검증

```bash
# 11개 워크플로우가 모두 존재하는지 확인
ls -1 src/moai_adk/templates/.github/workflows/*.yml | wc -l
# 예상 결과: 15 (기존 4개 + 신규 11개)

# 각 워크플로우 YAML 유효성 검증
yamllint src/moai_adk/templates/.github/workflows/*.yml

# 린팅 및 타입 체킹
ruff check src/moai_adk/
mypy src/moai_adk/
```

---

**@TAG 체인**:
- `@ACCEPTANCE:LANGUAGE-DETECTION-EXTENDED-001` (이 문서)
- `@SPEC:LANGUAGE-DETECTION-EXTENDED-001` (요구사항)
- `@PLAN:LANGUAGE-DETECTION-EXTENDED-001` (구현 계획)
- `@TEST:LANGUAGE-DETECTION-EXTENDED-001` (테스트 코드 - 구현 시 생성)
- `@CODE:LANGUAGE-DETECTION-EXTENDED-001` (구현 코드 - 구현 시 생성)

---

**구현 및 검증 시작**: `/alfred:2-run SPEC-LANGUAGE-DETECTION-EXTENDED-001`
