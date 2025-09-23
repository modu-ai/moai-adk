# MoAI-ADK Constitution (Universal)

> “명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다.”

본 헌법은 MoAI-ADK 프로젝트에서 모든 에이전트와 개발자가 따라야 할 **통합 가드레일**이다. Claude Code 세션, Headless CLI 자동화, 팀/개인 모드 작업을 모두 포함한다. 한국어를 기본 커뮤니케이션 언어로 사용한다.

---

## 0. 범위 · 기본 루프

- 적용 대상: `/moai:0-project → /moai:3-sync`, `/moai:git:*`, Headless CLI(`codex exec`, `gemini -p`)로 수행되는 모든 작업.
- 기본 행동 루프: **문제 정의 → 작고 안전한 변경 → 변경 리뷰 → 리팩터링 → 문서/TAG 동기화**.
- 모든 변경은 AGENTS 규칙, 16-Core TAG, Waiver 절차, 구조화 로깅, 보안 정책을 따른다.

---

## Article I — Mindset & Decision Loop
1. **시니어 엔지니어 마인드**: 가설 대신 사실 기반으로 판단하며, 최소 두 가지 대안을 비교(장점·단점·위험) 후 가장 단순한 해법을 선택한다.
2. **전체 문맥 파악**: 수정 전 관련 파일·정의·참조·테스트·문서·플래그를 끝까지 읽고 영향도를 1–3줄로 메모한다.
3. **한국어 소통**: 팀·AI 간 커뮤니케이션은 한국어를 기본으로 하며, 필요 시 원문 인용 후 설명을 제공한다.

---

## Article II — Workflow Guardrails
1. **사전 준비**: 코딩 전 `배경/문제/목표/비목표/제약`을 정리하고, 필요한 SPEC/TAG를 확인한다.
2. **작고 안전한 변경**: PR/커밋/파일 범위를 최소화하고, 위험 작업 전 체크포인트를 생성한다.
3. **리뷰와 문서화**: 변경 후 `/moai:3-sync`로 Living Document·TAG를 갱신하고, 요약/리뷰/리팩터링 계획을 남긴다.

---

## Article III — Simplicity & Architecture
1. **복잡도 권장치**: 기본 `simplicity_threshold = 5` 모듈. 초과가 필요하면 Waiver(사유·위험·완화책·만료 조건)를 작성한다.
2. **Fowler의 Two Hats**: “기능 추가 모드”와 “리팩터링 모드”를 명확히 구분하며, 두 모자를 동시에 쓰지 않는다.
3. **Refactoring Signals**: 긴 함수·거대한 클래스·긴 파라미터 목록·중복 코드·Feature Envy 등 코드 냄새가 감지되면 리팩터링을 계획한다.
4. **아키텍처 원칙**: Domain/Application/Infrastructure 계층 분리, DIP 준수, 인터페이스 우선 설계, API/데이터 계약 문서화.

---

## Article IV — Implementation & Code Quality
1. **Clean Code 규칙** (Robert C. Martin)
   - 명확한 이름, 작은 함수(≤50 LOC), 한 가지 작업만 수행, 인수는 최소화.
   - 중복 제거, 의도를 드러내는 구조, 주석은 최소화(의도 설명용).
   - 포맷팅은 팀 규칙을 우선하고, 관련 개념 사이에는 수직 여백을 둔다.
2. **변수 역할** (Sajaniemi): Fixed Value, Stepper, Flag, Walker, … 등 11개 역할을 명시적으로 사용한다.
3. **부수효과 격리**: I/O·네트워크·전역 상태는 경계층으로 분리하고, 가드절을 우선 사용하며 상수는 심볼화한다.
4. **보안/품질**: 입력 검증·정규화·인코딩, 구조화(JSON) 로깅, 민감 정보 마스킹, 최소 권한 원칙을 준수한다.

---

## Article V — Testing & TDD
1. **Red→Green→Refactor** (Kent Beck)
   - RED: 실패 테스트를 먼저 작성하고, 실패를 확인한다.
   - GREEN: 최소 코드로 테스트를 통과시킨 뒤 커밋한다.
   - REFACTOR: 모든 테스트가 통과한 상태에서만 리팩터링을 수행한다.
2. **테스트 정책**
   - 새 기능에는 새 테스트, 버그 수정에는 회귀 테스트를 추가한다.
   - 테스트는 결정적·독립적이어야 하며, 외부 시스템은 가짜/계약 테스트로 대체한다.
   - E2E 테스트에는 성공/실패 경로를 최소 1개씩 포함한다.
3. **커버리지 목표**: 권장 ≥ 85%. 부족 시 보완 계획 또는 Waiver를 기록한다.

---

## Article VI — Observability & Security
1. 구조화 로그(JSONL)와 핵심 메트릭(지연·오류율·처리량·자원)을 수집하고, PII/비밀은 `***redacted***`로 마스킹한다.
2. 모든 중요한 이벤트는 감사를 위해 기록하며, 훅 실패 시 서킷 브레이커를 발동해 세이프 모드(추가 승인, Bash 제한)로 전환한다.
3. 위험 명령(`rm -rf`, 비허용 네트워크 등)은 사전 차단하고, `policy_block.py`를 통해 허용 목록을 관리한다.

---

## Article VII — Versioning & GitOps
1. 시맨틱 버전(MAJOR.MINOR.BUILD)을 유지하고, Git 이력은 `@TAG`와 커밋 메시지를 통해 추적 가능하게 한다.
2. `/moai:git:*` 명령으로 체크포인트·브랜치·커밋·동기화를 수행하며, 위험 작업 전/후 자동 태그를 남긴다.
3. Draft PR → Ready 전환은 `/moai:3-sync`의 체크리스트를 따른다.

---

## Article VIII — Traceability & TAG
1. 16-Core @TAG 체인을 유지한다: Primary(@REQ→@DESIGN→@TASK→@TEST), Steering, Implementation, Quality.
2. `.moai/indexes/tags.json`과 `docs/status/sync-report.md`를 최신 상태로 유지한다.
3. HEADLESS 분석(`gemini -p`)과 구현(`codex exec`) 결과 보고 시 사용한 TAG와 체인 상태를 명시한다.

---

## Article IX — Review & Refactoring Discipline
1. **Rule of Three**: 동일한 패턴을 세 번째 발견하면 리팩터링을 계획한다.
2. **Preparatory Refactoring**: 변경이 쉬워지도록 환경을 먼저 정리한 뒤 변경을 적용한다.
3. **Litter-Pickup**: 발견한 작은 냄새는 즉시 정리하되, 범위가 커지면 별도 작업으로 분리한다.

---

## Article X — Microservice/API Patterns (Olaf Zimmermann)
1. **Foundation**: BFF, API Gateway, Client-Side Composition 중 상황에 맞는 프런트엔드 통합 전략을 선택한다.
2. **Design**: Request/Response, Request-Acknowledge, Event Message 등 패턴을 명시하고 계약 문서를 유지한다.
3. **Quality**: Pagination, Wish List, Conditional Request 등 성능 패턴과 Rate Limiting, Circuit Breaker 등 보안 패턴을 적용한다.
4. **Evolution**: 명시적 버전 식별자, “Two in Production”, Consumer-Driven Contracts, Published Language 등으로 호환성을 관리한다.

---

## Article XI — Exceptions & Waivers
- 권장 규칙을 초과하거나 위반해야 할 경우 Waiver 문서를 작성해 PR/Issue/ADR에 첨부한다.
- Waiver 내용: 사유, 대안 검토, 위험/완화책, 임시/영구 여부, 만료 조건, 승인자.

---

## Operational Annex A — Working Loop & Checklists
1. **사전 준비**
   - 배경/문제/목표/비목표/제약 작성
   - 관련 파일/테스트/문서/플래그 전체 읽기
   - 대안 비교 표 작성
2. **실행**
   - 필요한 SPEC/TAG 생성
   - 작은 단위 변경, 변경당 체크포인트 생성
   - TDD 사이클 준수, 테스트/린터 실행
3. **마무리**
   - `/moai:3-sync` 실행 → TAG 인덱스·문서 갱신
   - 분석/구현 명령(`codex exec`, `gemini -p`) 로그 기록 및 요약 보고
   - TODO/리팩터링 항목 남기고, 리뷰 요청

---

## Operational Annex B — Sajaniemi Variable Roles
| 역할 | 설명 | 예시 |
| --- | --- | --- |
| Fixed Value | 초기화 이후 변하지 않는 값 | `const MAX_SIZE = 100` |
| Stepper | 순차적으로 값을 변경 | `for (let i = 0; i < n; i++)` |
| Flag | 상태를 나타내는 불리언 | `let isValid = true` |
| Walker | 자료구조 순회 | `while (node) { node = node.next; }` |
| Most Recent Holder | 가장 최근 값 유지 | `let lastError` |
| Most Wanted Holder | 최적/최대값 유지 | `let bestScore = -Infinity` |
| Gatherer | 누적자 | `sum += value` |
| Container | 복수 값 저장 | `const list = []` |
| Follower | 다른 변수의 이전 값 | `prev = curr; curr = next;` |
| Organizer | 데이터 재구성 | `const sorted = array.sort()` |
| Temporary | 일시적 저장 | `const temp = a; a = b; b = temp;` |

---

## Operational Annex C — Refactoring Quick Reference
- **Extract Method**: 의도를 드러내고 중복 제거
- **Rename Variable**: 의미 있는 이름으로 변경
- **Move Method**: 적절한 객체로 이동
- **Replace Temp with Query**: 임시변수 대신 메서드 호출
- **Introduce Parameter Object**: 관련 파라미터 묶기
- **Matt Beck Rule**: “테스트가 실패하면 구현하지 않는다”

---

## Operational Annex D — TDD & Microservice Patterns
- **TDD 규칙**: 테스트를 먼저 작성, 실패 확인 후 최소 구현, 모든 테스트 통과 상태에서만 리팩터링.
- **Microservice 품질 패턴**: Pagination, Conditional Request, Rate Limiting, Circuit Breaker 적용.
- **API 문서화**: OpenAPI/Swagger 유지, Consumer-Driven Contracts로 양방향 검증.

---

이 헌법은 MoAI-ADK 4단계 파이프라인과 Git 자동화, Headless CLI 자동화, 팀/개인 협업을 안전하고 일관되게 수행하기 위한 기준이다. 모든 기여자는 본 문서를 세션 초기 메모리(예: CLAUDE.md)와 링크하여 항상 참조해야 한다.
