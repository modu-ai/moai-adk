# 소프트웨어 개발 마스터 원칙 요약

> 업계 거장들의 시대를 초월한 원칙을 MoAI-ADK 운영 관점으로 요약했습니다. 각 항목은 실무 적용 시 체크리스트로 활용하세요.

## 1) Martin Fowler: 리팩터링 원칙
- 정의: 외부 동작은 유지하면서 내부 구조를 개선하는 규율 있는 기법
- Two Hats: 기능 추가와 리팩터링은 동시에 하지 않는다(모드 전환 명확히)
- 코드 스멜: 긴 함수/거대 클래스/긴 파라미터/중복/Feature Envy 등 신호에 반응
- 언제 리팩터링: Rule of Three, 준비성 리팩터링(쉽게 만들고, 쉽게 바꾼다), 이해 향상 리팩터링, 청소 리팩터링(원래보다 깨끗하게)

적용 체크
- [ ] 모든 리팩터링은 테스트 그린 상태에서 수행(Red-Green-Refactor의 Refactor)
- [ ] 기능 추가 중에는 리팩터링 금지(두 모자 동시 금지)

## 2) J. Sajaniemi: 변수 역할 11가지
- Fixed Value, Stepper, Flag, Walker, Most Recent/Most Wanted Holder, Gatherer, Container, Follower, Organizer, Temporary
- 의의: 변수의 ‘의도’를 이름/스코프/불변성으로 드러내 가독성과 유지보수성 향상

적용 체크
- [ ] 이름이 역할을 드러내는가(예: step, found, last, best, sum)
- [ ] 불변 가능한 값은 상수화/불변화(const, final)

## 3) Robert C. Martin: Clean Code/SOLID
- SRP: 한 가지 이유로만 변경되는 모듈
- OCP: 확장에 열려 있고 수정에 닫혀 있음(전략/팩토리/DI)
- LSP: 하위 타입은 상위 타입을 대체 가능(계약 준수)
- ISP: 클라이언트별 인터페이스 분리(작고 응집도 높은 인터페이스)
- DIP: 추상에 의존하고 구체에 의존하지 말 것(상위 모듈 ↔ 하위 모듈 역전)

클린 코드 요약
- 의미 있는 이름, 작은 함수(한 가지 일), 적은 인자, 주석 최소화(코드로 설명), 팀 포맷 규칙 준수

적용 체크
- [ ] 함수 ≤ 50 LOC, 한 가지 일만 수행, 의미 있는 이름
- [ ] 높은 응집/낮은 결합, 의존성은 인터페이스/추상에 연결

## 4) Kent Beck: TDD (Red → Green → Refactor)
- Red: 실패하는 테스트 먼저 작성(의도된 실패 확인)
- Green: 통과에 필요한 최소 코드만 작성(YAGNI)
- Refactor: 테스트 그린 상태에서 구조 개선(중복 제거, 가독성/성능 개선)

적용 체크
- [ ] 테스트 먼저, 작은 스텝, 매 스텝 실행(빠른 피드백)
- [ ] 리팩터링은 항상 그린에서 수행, 새 기능 추가는 그린 이후

## 5) Olaf Zimmermann: Microservice API 디자인 패턴
- 파운데이션: BFF, API Gateway, 클라이언트 합성
- 요청/응답 패턴: Request/Response, Ack, Query, Callback
- 메시지 패턴: Document/Command/Event, Request-Reply 상관관계
- 품질 패턴: 페이지네이션, 필드 선택(위시리스트), 조건부 요청(ETag), Rate Limiting, OAuth2, Circuit Breaker
- 진화 패턴: 명시적 버저닝, SemVer, Two in Production, 공격적 폐기, Tolerant Reader, CDC(소비자 주도 계약)
- 실행 가이드: Design First(OpenAPI), 계약 테스트, 모니터링, 거버넌스

적용 체크
- [ ] API는 OpenAPI 우선 설계, 계약 테스트 포함
- [ ] 버전 정책 명확(두 버전 동시 운영 + 폐기 공지)

## 참고 문헌
- Fowler, “Refactoring” (2nd ed.)
- Sajaniemi, “Roles of Variables”
- Martin, “Clean Code”
- Beck, “Test-Driven Development: By Example”
- Zimmermann et al., “Patterns for API Design”

---

MoAI-ADK 맵핑(요약)
- SPECIFY: Clean Code(이름/작은 함수)로 요구 정의 명확화, TDD의 수락 기준으로 테스트 가능성 확보
- PLAN: SOLID/DIP/패턴 선정, Zimmermann API 패턴/버전전략, ADR로 결정 기록
- TASKS: 작은 단위 작업/의존성 그래프, 역할 기반 네이밍(변수 역할)
- IMPLEMENT: TDD 사이클 + Refactor, 두 모자 원칙 준수
