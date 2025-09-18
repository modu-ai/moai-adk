# $PROJECT_NAME 백엔드(Spring Boot) 메모

> Spring Boot 기반 서비스 운영 지침입니다. Java/Kotlin 규칙(@.claude/memory/coding_standards/java-kotlin.md)과 함께 사용하세요.

## 1. 빌드 & 의존성 관리
- Gradle Kotlin DSL 권장, 버전 카탈로그(`libs.versions.toml`) 관리
- 코드 포맷: Spotless + ktlint/Google Java Format, 정적 분석: Checkstyle/Detekt
- 의존성 보안: OWASP Dependency-Check 또는 Gradle `dependencyCheckAnalyze`

## 2. 품질 파이프라인
- 테스트: JUnit 5 + AssertJ, Mockito/Kotlin test, Testcontainers(BDD) → 커버리지 80% 이상(JaCoCo)
- 슬라이스 테스트: `@WebMvcTest`, `@DataJpaTest`, API 통합 테스트는 MockMvc/WebTestClient 사용
- Contract 테스트: Spring Cloud Contract 또는 Pact 연동 고려

## 3. 설정 & 보안
- 설정: `@ConfigurationProperties` + `@Validated`, profile 기반 `application-{env}.yml`
- 인증/인가: Spring Security (JWT/OAuth2), 최소 권한 원칙, 감사 로그 작성
- Observability: Actuator + Micrometer (Prometheus), Logback JSON + MDC(requestId)

## 4. Build Hooks / pre-commit
- pre-commit 예시
```yaml
repos:
  - repo: local
    hooks:
      - id: gradle-spotless
        name: gradle spotless
        entry: ./gradlew spotlessCheck
        language: system
      - id: gradle-tests
        name: gradle test
        entry: ./gradlew test
        language: system
      - id: gradle-lint
        name: detekt check
        entry: ./gradlew detekt
        language: system
```
- CI: `./gradlew clean verify` + `./gradlew jacocoTestReport` + `./gradlew dependencyCheckAnalyze`

## 5. 배포 & 운영
- 컨테이너: Jib 또는 Buildpacks, `JAVA_OPTS` 튜닝, readiness/liveness probe 정의
- 마이그레이션: Flyway/Liquibase, 롤백 전략 문서화
- Blue/Green, Rolling Update 시 Actuator `/health` 기반 검증, distributed tracing(OpenTelemetry) 연동

## 6. 참고 문서
- Java/Kotlin 규칙: @.claude/memory/coding_standards/java-kotlin.md
- 공통 체크리스트/보안/TDD: @.claude/memory/shared_checklists.md / @.claude/memory/security_rules.md / @.claude/memory/tdd_guidelines.md
