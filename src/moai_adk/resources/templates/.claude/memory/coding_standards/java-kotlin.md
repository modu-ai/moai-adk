# Java/Kotlin 규칙

## ✅ 필수
- Java: Spotless + Checkstyle로 포맷팅/정적분석, JUnit 5 + AssertJ 사용, Optional 남용 금지
- Kotlin: detekt + ktlint, null-safety/Sealed class 적극 활용, coroutine 구조화(`SupervisorJob` 등)
- Spring/SpringBoot: @Configuration 과사용 금지, Profile/Property로 환경 분리, 빈 주입은 생성자 기반
- 테스트: Spring slice(MockMvc/WebTestClient), 통합/컨트랙트 테스트 병행, Testcontainers로 외부 의존성 격리

## 👍 권장
- 불변 객체/Value Object 활용, Lombok 보단 record/sealed interface 선호
- Kotlin DSL(Gradle)로 빌드 스크립트 구성, BOM/Version catalog 관리
- REST/GraphQL API는 @Validated/Bean Validation 적용, DTO-Entity 변환 분리

## 🚀 확장/고급
- Reactor/Coroutines 기반 비동기 흐름, observability(Micrometer/OpenTelemetry) 통합
- ArchUnit/JDepend로 아키텍처 규칙 검증, ErrorProne/NullAway 등 추가 분석
- Hexagonal/Clean Architecture 패턴 도입, 모듈화/멀티 모듈 프로젝트(Gradle composite) 운영
