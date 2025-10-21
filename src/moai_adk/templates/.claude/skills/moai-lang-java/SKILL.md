---

name: moai-lang-java
description: Java best practices with JUnit, Maven/Gradle, Checkstyle, and Spring Boot patterns. Use when writing or reviewing Java code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Java Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Java code discussions, framework guidance, or file extensions such as .java. |
| Tier | 3 |

## What it does

Provides Java-specific expertise for TDD development, including JUnit testing, Maven/Gradle build tools, Checkstyle linting, and Spring Boot patterns.

## When to use

- Engages when the conversation references Java work, frameworks, or files like .java.
- “Writing Java tests”, “How to use JUnit”, “Spring Boot patterns”
- Automatically invoked when working with Java projects
- Java SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **JUnit 5**: Unit testing with annotations (@Test, @BeforeEach)
- **Mockito**: Mocking framework for dependencies
- **AssertJ**: Fluent assertion library
- Test coverage ≥85% with JaCoCo

**Build Tools**:
- **Maven**: pom.xml, dependency management
- **Gradle**: build.gradle, Kotlin DSL support
- Multi-module project structures

**Code Quality**:
- **Checkstyle**: Java style checker (Google/Sun conventions)
- **PMD**: Static code analysis
- **SpotBugs**: Bug detection

**Spring Boot Patterns**:
- Dependency Injection (@Autowired, @Component)
- REST controllers (@RestController, @RequestMapping)
- Service layer separation (@Service, @Repository)
- Configuration properties (@ConfigurationProperties)

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Interfaces for abstraction
- Builder pattern for complex objects
- Exception handling with custom exceptions

## Examples
```bash
./mvnw test && ./mvnw checkstyle:check
```

## Inputs
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- Oracle. "Java Language Specification." https://docs.oracle.com/javase/specs/ (accessed 2025-03-29).
- JUnit. "JUnit 5 User Guide." https://junit.org/junit5/docs/current/user-guide/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Java-specific review)
- database-expert (JPA/Hibernate patterns)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
