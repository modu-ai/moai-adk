# MoAI-ADK 엔지니어링 표준

> 공통 코딩 규칙과 언어/플랫폼별 가이드를 한곳에 모았습니다. `.claude/memory` 요약본을 확인했다면 이 문서에서 전문을 참고하세요.

## 1. 공통 규칙
- 작은 단위: 파일 ≤ 300 LOC, 함수 ≤ 50 LOC, 파라미터 ≤ 5, 순환 복잡도 < 10.
- 구조: 입력 → 처리 → 반환, 가드절 우선, 부수효과(I/O·네트워크)는 경계 계층에 격리.
- 명시성: 의미 있는 네이밍, 상수 심볼화, 구조화 로그(JSON) + 상관관계 ID.
- 테스트: TDD(Red→Green→Refactor)와 공통 체크리스트(@.moai/memory/operations.md) 준수, 커버리지 ≥ 80%.
- 보안: 입력 검증/정규화/인코딩, 파라미터 바인딩, 민감정보 미노출, 최소 권한.
- 문서: 변경 시 CLAUDE.md / `.moai/memory` 갱신, `/moai:6-sync` 수행.

## 2. 언어/플랫폼 가이드

### Python
- black, ruff, mypy(strict) + pytest/pytest-cov.
- 타입 힌트 100%, Protocol/TypedDict로 인터페이스 명시.
- asyncio 및 contextmanager로 자원 누수 방지, 구조화 로깅.

### TypeScript / React / Next.js
- `tsconfig` strict, ESLint + Prettier, Jest/Vitest + Testing Library.
- 타입은 `unknown→narrow`, `any` 금지, `never` 처리 필수.
- Next.js: App Router, Server Component 기본, fetch 캐싱·ISR 전략 문서화.

### Go
- `gofmt`, `golangci-lint`, `go test -race`.
- `errors.Is/As`, `fmt.Errorf("%w")`, context 전달, goroutine 누수 방지.
- `internal/` 디렉터리로 경계 명확화, 인터페이스는 사용 지점에 배치.

### Java / Kotlin / Spring
- Spotless + Checkstyle/detekt, JUnit5 + AssertJ, Testcontainers.
- `@ConfigurationProperties` + `@Validated`, 생성자 주입, Profile 분리.
- Observability: Micrometer + Actuator, Logback JSON + MDC.

### C# / .NET
- `dotnet format`, StyleCop, xUnit + coverlet.
- `<Nullable>enable</Nullable>`, async/await + CancellationToken 전파.
- DI 컨테이너는 인터페이스 기반 등록, EF Core context per scope.

### Rust
- `rustfmt`, `clippy -D warnings`, `cargo nextest`.
- `anyhow`/`thiserror`, panic은 초기화 단계에 한정.
- tokio runtime 정책 준수, tracing + structured log.

### Swift / SwiftUI
- SwiftFormat + SwiftLint, async/await + MainActor 규칙.
- SwiftUI 상태/뷰모델 분리, XCTest + XCTExpectation.
- Swift Package Manager 모듈화, Instrumentation으로 성능 튜닝.

### SQL / 데이터
- 파라미터 바인딩, 인덱스 전략/EXPLAIN 검증.
- 마이그레이션 업/다운, 롤백 절차 문서화.
- 제약 조건/데이터 품질 관리, Observability 연동.

### Shell / CLI
- `set -euo pipefail`, 안전 IFS, `"$VAR"` quoting.
- `rg`/`fd` 우선, `rm -rf` 등 파괴 명령은 확인 절차 후 실행.
- shellcheck + shfmt, 복잡 로직은 고급 언어로 전환.

### Terraform / IaC
- `terraform fmt validate`, tflint, tfsec.
- 원격 backend, 버전 고정, drift 모니터링, Policy as Code(OPA).

### Frameworks 빠른 참조
- React/Next: Server/Client 분리, Suspense/CSR 전략 문서화, Storybook/Playwright.
- Vue/Nuxt: Composition API, Pinia, Nitro 서버 설정, Vitest + Cypress.
- Angular: Standalone 컴포넌트, OnPush, RxJS, Jest/Vitest.
- FastAPI: Pydantic v2, TestClient + pytest-asyncio, OpenAPI contract 테스트.
- Spring Boot: ConfigProperties 검증, Actuator, Micrometer, Flyway/Liquibase.

## 3. 도구 & 자동화
- pre-commit: 언어별 포맷터/린터/테스트 훅을 정의하고 로컬·CI에서 동일하게 실행.
- CI: lint → test → build → 보안 스캔 순으로 파이프라인 구성, 실패 시 빌드 차단.
- Observability: OpenTelemetry/Prometheus/Log aggregation을 기본 구성으로 채택.

## 4. 참고
- 운영 규칙 및 체크리스트: `@.moai/memory/operations.md`
- 헌법 전문: `@.moai/memory/constitution.md`
- 보안/테스트 세부: `@.claude/memory/security_rules.md`, `@.claude/memory/tdd_guidelines.md`
