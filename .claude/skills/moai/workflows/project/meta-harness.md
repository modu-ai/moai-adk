---
name: moai-workflow-project-meta-harness
description: "Project Phase 5/6 — 16-question 4-round Socratic interview (AskUserQuestion) for harness activation and meta-harness invocation"
user-invocable: false
metadata:
  parent: moai-workflow-project
  phase: "Phase 5/6: Socratic Interview and meta-harness Invocation"
---

<!-- TRACE PROBE: per SPEC-V3R4-WORKFLOW-SPLIT-001 T0.5 baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

## Phase 5: Socratic Interview (Harness Activation)

Purpose: Conduct a 16-question / 4-round Socratic interview using `AskUserQuestion` to gather
project context required by `moai-meta-harness`. Answers are accumulated in an in-memory buffer
(no disk I/O) until Round 4 Q16 final confirmation (REQ-PH-001, REQ-PH-002, REQ-PH-010).

[HARD] Each round is exactly one `AskUserQuestion` call with up to 4 questions (C-PH-003).
[HARD] Each question's first option MUST be marked "(권장)" with a detailed description (C-PH-003).
[HARD] All question text and option labels MUST be in conversation_language (default: ko) (C-PH-004).
[HARD] No disk I/O until Round 4 Q16 "Confirm" answer is received (REQ-PH-010).

In-Memory Buffer Protocol:
- Maintain all 16 answers in memory across the 4 `AskUserQuestion` calls.
- On "Confirm" (Q16): call `Buffer.Commit()`, then proceed to write `.moai/harness/interview-results.md`.
- On "Restart" (Q16): clear the buffer and restart from Round 1.
- On "Abort" (Q16): call `Buffer.Abort()` — clears all answers, writes zero bytes to disk, and exits Phase 5.

---

### Round 1: Q1–Q4 (도메인 / 기술스택 / 규모 / 팀구성)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q1 — 도메인 (Project Domain)**

질문: 이 프로젝트의 주요 도메인은 무엇인가요?

옵션:
- (권장) 웹 (Web Application): 프론트엔드+백엔드 풀스택 또는 API 서비스. 사용자 대면 대시보드, SaaS, 이커머스 등에 최적. React/Vue/Next.js + REST/GraphQL 조합이 일반적.
- 모바일 (iOS): Swift + SwiftUI 또는 UIKit 기반 iOS 네이티브 앱. App Store 배포 대상. FaceID/HealthKit 등 iOS 전용 API 활용 가능.
- 모바일 (Android): Kotlin + Jetpack Compose 또는 XML 기반 Android 앱. Google Play 배포 대상.
- 기타 (Other): CLI 도구, 임베디드 시스템, 데스크톱 앱, 크로스플랫폼 (Flutter/React Native) 등 위 분류에 해당하지 않는 경우.

**Q2 — 기술스택 (Primary Technology Stack)**

질문: 주요 기술 스택은 무엇인가요?

옵션:
- (권장) TypeScript / JavaScript (Node.js + React/Next.js): 풀스택 JS 생태계. 프론트+백 코드 공유, 큰 npm 생태계, Vercel/AWS Lambda 배포 친화적.
- Go: 고성능 마이크로서비스, CLI, 클라우드 네이티브 바이너리. 단순 배포, 정적 컴파일, 낮은 메모리 사용.
- Python: AI/ML 워크로드, 백엔드 API (FastAPI/Django). 데이터 사이언스 라이브러리 풍부.
- 기타 (Swift / Kotlin / Rust / Java / C# 등): 위 3개에 해당하지 않는 언어. 구체적 언어를 직접 입력.

**Q3 — 규모 (Project Scale)**

질문: 프로젝트 규모는 어느 정도인가요?

옵션:
- (권장) MVP (1-3 모듈, 단기): 핵심 기능 1-3개로 빠르게 검증. 1-2주 내 첫 배포 목표. 기술 부채 최소화 우선.
- Small (4-8 모듈, 1-3개월): 안정화된 기능셋, 팀 2-4명, CI/CD 포함 구성.
- Medium (9-20 모듈, 3-12개월): 여러 도메인 레이어, 팀 5-10명, 마이크로서비스 또는 모듈 분리 고려.
- Large (20+ 모듈 또는 멀티팀): 조직 규모 제품, 복수 팀 협업, 플랫폼 엔지니어링 필요.

**Q4 — 팀구성 (Team Composition)**

질문: 팀 구성은 어떻게 되나요?

옵션:
- (권장) 솔로 개발자 (Solo developer): 1인 개발. 모든 역할 담당. 자동화와 AI 보조 도구로 생산성 보완.
- 소규모 팀 (2-4명): 풀스택 개발자 2-4명. 역할 유동적. 코드 리뷰 필수.
- 중간 팀 (5-10명): 프론트/백 분리, QA 포함. 명확한 소유권과 PR 프로세스 필요.
- 대규모 / 멀티팀: 10명 이상 또는 다수 팀. 아키텍처 가이드, API 계약, 플랫폼 레이어 필수.

---

### Round 2: Q5–Q8 (방법론 / 디자인툴 / UI복잡도 / 디자인시스템)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q5 — 방법론 (Development Methodology)**

질문: 주요 개발 방법론은 무엇인가요?

옵션:
- (권장) TDD (테스트 주도 개발): 테스트 먼저 작성 후 구현. RED-GREEN-REFACTOR 사이클. 새 기능 개발에 최적.
- DDD (도메인 주도 개발): 기존 코드베이스 리팩토링. ANALYZE-PRESERVE-IMPROVE 사이클. 레거시 코드에 최적.
- Agile / Scrum: 스프린트 기반 반복 개발. 백로그 관리, 데일리 스탠드업, 스프린트 리뷰.
- 기타 (Kanban / Waterfall / Ad-hoc): 위 방법론에 해당하지 않는 경우 직접 기술.

**Q6 — 디자인툴 (Design Tool)**

질문: UI/UX 디자인에 어떤 도구를 사용하나요?

옵션:
- (권장) Figma: 협업 디자인 도구. 디자인 토큰 추출, 컴포넌트 라이브러리, 개발자 핸드오프 지원.
- Sketch: macOS 전용 디자인 도구. 플러그인 생태계 풍부. Zeplin 핸드오프 많이 사용.
- Adobe XD: Adobe 생태계 통합. 프로토타이핑과 디자인 시스템 관리.
- 없음 / 코드 기반: 별도 디자인 툴 없이 코드로 직접 UI 구현. Storybook 등 컴포넌트 주도.

**Q7 — UI복잡도 (UI Complexity)**

질문: UI 복잡도는 어느 수준인가요?

옵션:
- (권장) 표준 (목록 + 폼 + 네비게이션): 일반적인 CRUD UI. 테이블, 폼, 모달, 내비게이션 바 수준.
- 단순 (정보성 페이지 / 랜딩): 마케팅 페이지, 대시보드 요약, 읽기 전용 뷰.
- 복잡 (데이터 시각화 / 드래그앤드롭): 차트, 그래프, 인터랙티브 에디터, 캔버스 기반 UI.
- 매우 복잡 (실시간 협업 / 3D / 게임): WebRTC, Three.js, 게임 UI 등 고도의 인터랙티비티.

**Q8 — 디자인시스템 (Design System)**

질문: 어떤 디자인 시스템을 사용할 예정인가요?

옵션:
- (권장) 기존 컴포넌트 라이브러리 (MUI / shadcn / Tailwind UI): 검증된 오픈소스 컴포넌트. 빠른 시작, 커스터마이징 가능.
- 커스텀 DTCG 토큰: W3C Design Token Community Group 표준. Figma 토큰 직접 추출, 완전 커스텀.
- 플랫폼 기본 (SwiftUI / Jetpack Compose / WinUI): 플랫폼 네이티브 UI. OS 가이드라인 자동 준수.
- 없음 / 미정: 디자인 시스템 없이 개별 스타일 적용. 추후 도입 예정.

---

### Round 3: Q9–Q12 (보안 / 성능 / 배포 / 외부통합)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q9 — 보안 (Security Requirements)**

질문: 주요 보안 요구사항은 무엇인가요?

옵션:
- (권장) 표준 인증 (JWT + OAuth2): 일반적인 웹/모바일 인증. Access/Refresh 토큰, 소셜 로그인 지원.
- 강화 보안 (OAuth + Keychain / Secure Enclave): iOS Keychain, Android Keystore, HSM 등 하드웨어 보안 요소 활용.
- 엔터프라이즈 (SSO / SAML / MFA): 기업 환경. Azure AD, Okta, LDAP 연동, 다중 인증.
- 최소 보안 (API Key 수준): 내부 도구, 프로토타입. 단순 API Key 또는 Basic Auth.

**Q10 — 성능 (Performance Target)**

질문: 성능 목표는 무엇인가요?

옵션:
- (권장) 일반 UI 반응성 (60fps, <200ms): 표준 앱 성능. 일반적인 CRUD 앱에 적합.
- 고성능 / 실시간 (<50ms): 금융, 게임, 실시간 협업. 최적화된 렌더링, 캐싱, WebSocket.
- 대용량 처리 (배치 / 스트리밍): 대규모 데이터 처리. 비동기 큐, 스트림 처리, 수평 확장.
- 저성능 환경 대응 (제한된 네트워크 / 구형 기기): 모바일 오프라인, IoT, 저사양 디바이스 지원.

**Q11 — 배포 (Deployment Target)**

질문: 어디에 배포할 예정인가요?

옵션:
- (권장) 클라우드 (AWS / GCP / Azure / Vercel): 관리형 클라우드. 오토스케일링, 관리형 DB, CDN.
- 앱 스토어 (App Store / Google Play): 모바일 앱 배포. 앱 심사, 버전 관리, 업데이트 정책 필요.
- 자체 서버 / On-premise: 자체 인프라. Docker + Kubernetes 또는 bare metal.
- 하이브리드 (클라우드 + 앱스토어): 모바일 앱 + 백엔드 API 조합.

**Q12 — 외부통합 (External Integrations)**

질문: 어떤 외부 시스템과 통합이 필요한가요?

옵션:
- (권장) 없음 / 표준 (결제 / 이메일 / SMS): Stripe, SendGrid, Twilio 등 범용 서비스 통합.
- 플랫폼 API (HealthKit / Maps / Push): iOS/Android 플랫폼 전용 API.
- 엔터프라이즈 시스템 (ERP / CRM / SAP): 기업 내부 시스템 연동. REST/SOAP/EDI.
- AI / ML 서비스 (OpenAI / Claude / Vision API): 외부 AI API 호출. 프롬프트 관리, 응답 처리.

---

### Round 4: Q13–Q16 (customization 범위 / 특수제약 / 우선순위 / 최종확인)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q13 — customization 범위 (Harness Customization Scope)**

질문: 프로젝트 전용 harness의 customization 범위는 어떻게 할까요?

옵션:
- (권장) 표준 (Standard): 도메인 특화 에이전트 2개 + 스킬 2개. 대부분의 프로젝트에 충분. moai-meta-harness가 답변 기반으로 최적 구성 자동 생성.
- 경량 (Minimal): 도메인 특화 스킬 1개만. 가장 빠른 setup. MVP 또는 소규모 프로젝트에 적합.
- 심화 (Thorough): 에이전트 3개 이상 + 스킬 3개 이상 + design-extension 포함. 복잡한 도메인에 최적.
- 전체 커스텀 (Advanced / full custom): 모든 요소를 완전 커스텀. design-extension.md 추가 생성 (REQ-PH-012). 고급 사용자용.

**Q14 — 특수제약 (Special Constraints)**

질문: 프로젝트에 특수 제약 사항이 있나요?

옵션:
- (권장) 없음 (No special constraints): 일반적인 제약만 적용. harness가 표준 패턴 사용.
- 최소 OS 버전 (iOS 17+ / Android 12+ 등): 플랫폼 최소 버전 제약. 하위 호환 API 사용 제한.
- 규정 준수 (HIPAA / GDPR / SOC2): 데이터 보호 규정. 암호화, 감사 로그, 데이터 거주지 제약.
- 기타 제약 (오프라인 필수 / 특정 하드웨어 / 정부 규격): 위에 해당하지 않는 특수 제약 사항.

**Q15 — 우선순위 (Harness Quality Level)**

질문: Harness 품질 수준(harness level)을 선택해 주세요.

옵션:
- (권장) standard: 기본 품질 게이트. 대부분의 프로젝트에 적합. 빠른 실행과 충분한 검증의 균형.
- thorough: 전체 evaluator-active + TRUST 5 검증. 복잡한 SPEC 또는 엔터프라이즈 프로젝트에 권장.
- minimal: 빠른 검증만. 단순 변경 또는 프로토타입에 적합. 일부 품질 게이트 생략.
- custom: 직접 구성. `.moai/config/sections/harness.yaml`에서 세부 설정 가능.

**Q16 — 최종확인 (Final Confirmation)**

질문: 위 16개 답변을 바탕으로 프로젝트 전용 harness를 생성할까요?

옵션:
- (권장) Confirm — 생성 진행: 모든 답변을 확인했습니다. `.moai/harness/interview-results.md`에 결과를 기록하고 Phase 6 (meta-harness 호출)으로 진행합니다.
- Restart — 처음부터 다시: Round 1부터 인터뷰를 다시 시작합니다. 이전 답변은 모두 초기화됩니다.
- Abort — 취소: 인터뷰를 중단합니다. 어떠한 파일도 생성되지 않습니다 (REQ-PH-010).

**Q16 Branch Logic:**
- "Confirm" → `Buffer.Commit()` 호출 → `.moai/harness/interview-results.md` 작성 → Phase 6 (meta-harness)으로 진행.
- "Restart" → `Buffer.Abort()` 후 `NewBuffer()` → Round 1부터 재시작.
- "Abort" → `Buffer.Abort()` 호출 → 디스크에 0 파일 작성 → Phase 5 종료 (zero disk writes, REQ-PH-010).

---

## Phase 6: meta-harness Invocation

Purpose: Call `Skill("moai-meta-harness")` with the 16 answers collected in Phase 5,
generating project-specific dynamic harness artifacts in the user area
(REQ-PH-004, T-P2-01).

[HARD] This phase MUST run the FROZEN guard (`EnsureAllowed`) as the **first check**
before any write attempt. Paths in `.claude/agents/{core,expert,meta,harness}/`, `.claude/skills/moai-*/`,
or `.claude/rules/moai/` are permanently FROZEN and must be rejected immediately.

[HARD] If meta-harness generation fails mid-way, `CleanupOnFailure` MUST remove all
partial artifacts written so far (REQ-PH-010).

### 6.1 Pre-Condition

- Phase 5 Round 4 Q16 answer is "Confirm" → `Buffer.Commit()` has been called.
- `.moai/harness/interview-results.md` has been written by `WriteResultsToFile`.

### 6.2 Answer-to-Context Schema

Convert the 16 in-memory answers to a structured prompt context before invoking
`Skill("moai-meta-harness")`. Each question maps to a named field:

```yaml
# Answer-to-context schema (YAML form)
context:
  # Round 1 — Domain & Technology
  domain:            # Q01 answer text (e.g., "모바일 (iOS)")
  tech_stack:        # Q02 answer text (e.g., "Swift + SwiftUI")
  project_scale:     # Q03 answer text (e.g., "MVP (1-3 모듈, 단기)")
  team_composition:  # Q04 answer text (e.g., "솔로 개발자")

  # Round 2 — Methodology & Design
  methodology:       # Q05 answer text (e.g., "TDD")
  design_tool:       # Q06 answer text (e.g., "Figma")
  ui_complexity:     # Q07 answer text (e.g., "표준 (목록 + 폼 + 네비게이션)")
  design_system:     # Q08 answer text (e.g., "커스텀 DTCG 토큰")

  # Round 3 — Security, Performance, Deployment
  security:          # Q09 answer text (e.g., "강화 보안 (OAuth + Keychain / Secure Enclave)")
  performance:       # Q10 answer text (e.g., "일반 UI 반응성 (60fps, <200ms)")
  deployment:        # Q11 answer text (e.g., "앱 스토어 (App Store / Google Play)")
  integrations:      # Q12 answer text (e.g., "플랫폼 API (HealthKit / Maps / Push)")

  # Round 4 — Customization & Final Confirmation
  customization_scope: # Q13 answer text (e.g., "표준 (Standard)")
  special_constraints: # Q14 answer text (e.g., "최소 OS 버전 (iOS 17+ / Android 12+ 등)")
  harness_level:       # Q15 answer text (e.g., "standard")
  final_confirmation:  # Q16 answer text — always "Confirm" at this point
```

### 6.3 Invocation Protocol

```
Skill("moai-meta-harness") with:
  - context: <structured answer map above>
  - project_root: <absolute path to project root>
  - spec_id: <SPEC-PROJ-INIT-NNN from interview-results.md>
  - conversation_language: <ko|en|ja|zh>
  - harness_level: <Q15 answer: minimal|standard|thorough>
  - design_extension: <true if Q13 == "전체 커스텀 (Advanced / full custom)", else false>
```

### 6.4 Expected Outputs

After successful meta-harness invocation, the following artifacts must exist
in the **user area** (FROZEN guard pre-verified):

| Artifact | Path | Required |
|----------|------|----------|
| Architect agent | `.claude/agents/harness/<domain>-architect.md` | Always |
| Engineer agent | `.claude/agents/harness/<domain>-engineer.md` | Always |
| Patterns skill | `.claude/skills/moai-harness-<domain>-patterns/SKILL.md` | Always |
| Best-practices skill | `.claude/skills/moai-harness-<domain>-best-practices/SKILL.md` | Always |
| Harness directory | `.moai/harness/` | Always |
| Design extension | `.moai/harness/design-extension.md` | Q13 == Advanced only |

All write paths must pass `EnsureAllowed(path)` before the file is created.
Any `FrozenViolationError` causes immediate abort + `CleanupOnFailure`.

### 6.5 Failure Handling

If `Skill("moai-meta-harness")` returns an error or partial output:

1. Call `CleanupOnFailure(tracker, err)` — removes all tracked partial files.
2. Surface the error to the user with a clear message.
3. Do NOT proceed to Phase 7 (5-Layer Activation).
