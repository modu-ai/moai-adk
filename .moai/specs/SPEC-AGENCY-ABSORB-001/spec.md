---
id: SPEC-AGENCY-ABSORB-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-20
author: GOOS
priority: High
labels: [agency, migration, design, hybrid, absorption]
issue_number: null
---

# SPEC-AGENCY-ABSORB-001: Agency → MoAI-ADK 흡수 및 Claude Design 통합

## HISTORY

- 2026-04-20 v0.1.0: 최초 작성. Option B(흡수) + Option A(하이브리드 디자인) + Option A(전체 마이그레이션) 확정 반영.
- 2026-04-20 v0.2.0: plan-auditor iteration 1 FAIL (0.62) 후속 수정. 시간 예측 제거(D1), YAML frontmatter 필수 필드 추가(D2), REQ-FALLBACK-003 EARS 재작성(D3), REQ-REMOVE-002 EARS 재작성(D4), REQ-ROUTE-003 vs 006 모순 해소(D5), REQ-SKILL-015 bundle 버전 불일치 신규(D6), REQ-MIGRATE-013 SIGINT/SIGTERM 신규(D7), REQ-FALLBACK-003 DoD 일관성(D8), REQ-MIGRATE-012 플랫폼 분리 → REQ-MIGRATE-012a/b(D9), REQ-SKILL-008/012 분할(D10, D11), Constitution 섹션 정렬(D12), REQ-MIGRATE-004 Step 6 스코프 명확화(D13), 5개 "(암시)" AC를 명시적 시나리오로 전환(D14).

---

## 1. Goal (목적)

`/agency` 자가진화 창작 제작 시스템을 `/moai` 오케스트레이터로 무손실 흡수하여 중복 에이전트를 제거하고, Claude Design(2026-04-17 출시, Opus 4.7 기반)을 디자인 작업의 권장 경로로 통합한다. 기존 `.agency/` 프로젝트 데이터는 `moai migrate agency` 커맨드로 `.moai/` 네임스페이스로 원자적으로 이전한다.

### 1.1 배경

현재 저장소는 두 시스템을 병행 운영:
- `/moai`: SPEC-First DDD/TDD 오케스트레이터
- `/agency`: 창작 제작 GAN 루프 (6개 에이전트)

`.agency/fork-manifest.yaml`에 따라 planner/builder/evaluator는 moai 에이전트에서 포크됨이 명시되어 있어 구조적 중복. Claude Design 출시로 시각 디자인 작업이 Claude.ai 내부 도구로 이관될 수 있는 경로가 열림.

### 1.2 비목표(Non-Goals)

- Claude Design API/CLI/MCP 통합 구현 (공개되지 않음)
- `/agency` 커맨드의 즉시 완전 제거 (deprecation 라이프사이클 준수)
- `.agency/` 디렉터리의 하드 삭제 (`.agency.archived/`로 백업만)
- agency 자가진화 학습 데이터의 의미론적 변환 (단순 파일 이동)

---

## 2. Scope (범위)

### 2.1 In Scope

- `/moai design` 서브커맨드 신규 추가 (하이브리드 분기)
- `moai migrate agency` CLI 커맨드 신규 추가 (Go 구현)
- 신규 스킬 4개: `moai-domain-copywriting`, `moai-domain-brand-design`, `moai-workflow-design-import`, `moai-workflow-gan-loop`
- `.moai/project/brand/` 템플릿 3개: `brand-voice.md`, `visual-identity.md`, `target-audience.md`
- `.moai/config/sections/design.yaml` 신규 설정 파일
- `.claude/rules/moai/design/constitution.md` 재배치 (agency/constitution.md 기반)
- `/agency` 커맨드의 deprecation wrapper 변환
- `manager-spec` 에이전트에 BRIEF 섹션(Goal/Audience/Brand) 확장
- agency 에이전트 4개 제거: planner, designer, builder, evaluator

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Claude Design과의 프로그래매틱 양방향 통합 (API 미공개로 불가)
- Claude Design 미보유(Pro 이하) 사용자용 내장 비주얼 에디터
- agency 자가진화 루프(Learner의 자동 Graduation)의 Go 포팅 — 현 단계에서는 Markdown 기반 `moai-workflow-research` 통합만 수행
- `fork-manifest.yaml`의 동적 upstream sync 자동화 (아카이브 후 참조만 유지)
- 기존 `.agency/learnings/` 엔트리의 구조 변환 (원본 보존, 네임스페이스만 변경)
- 마이그레이션 롤백 후 원복된 `.agency/`에서 시스템을 다시 활성화하는 기능 — 롤백은 데이터 복구 용도이지 agency 모드 복원이 아님

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.23+)
- Claude Code v2.1.111+
- Opus 4.7
- 템플릿 시스템: `go:embed` 기반 (`internal/template/templates/`)
- 대상 OS: macOS, Linux, Windows (동등)
- 영향 디렉터리: `.agency/`, `.moai/`, `.claude/agents/agency/`, `.claude/skills/agency-*/`, `.claude/commands/agency.md`, `.claude/rules/agency/`

---

## 4. Assumptions (가정)

- 사용자는 `.agency/` 데이터의 무손실 이전을 가장 중요한 수용 기준으로 삼는다.
- Claude Design은 공식 문서에 기록된 기능 범위(ZIP/HTML/PDF/PPTX/Canva/handoff bundle 내보내기)를 2026-04-20 기준 안정적으로 제공한다.
- 사용자는 수동으로 Claude.ai UI에서 생성한 handoff bundle을 로컬 파일 시스템 경로로 전달할 수 있다.
- `moai migrate agency` 실행 시 `.agency/`는 읽기 가능, `.moai/`는 쓰기 가능 상태다.
- 기존 `/agency` 커맨드 사용자 이탈 허용 기간은 최소 1 마이너 버전 주기 동안 유지된다.
- Pro 이하 사용자(Claude Design 미가용)는 코드 기반 폴백(경로 B)으로 동등한 결과를 얻을 수 있다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 REQ-ROUTE: `/moai design` 라우팅 요구사항

**REQ-ROUTE-001 (Event-Driven)**  
**When** 사용자가 `/moai design "<brief>"`를 호출하면, the 시스템 **shall** 자동으로 브랜드 컨텍스트(`.moai/project/brand/`)의 존재 여부를 확인하고, 없을 경우 브랜드 인터뷰 플로우를 제안한다.

**REQ-ROUTE-002 (Event-Driven)**  
**When** `/moai design`이 호출되면, the 시스템 **shall** AskUserQuestion을 통해 두 경로 중 하나를 사용자에게 선택하도록 요청한다:
- 경로 A: Claude Design 위임 (권장, Claude.ai Pro/Max/Team/Enterprise 필요)
- 경로 B: 코드 기반 스킬(`moai-domain-brand-design`) 폴백

**REQ-ROUTE-003 (Ubiquitous)**  
The `/moai design` 커맨드 **shall** 기본 구독 플랜(Claude.ai Pro 이상) 가정 하에 첫 번째 옵션을 "Claude Design 위임(권장)"으로 표시하며, 각 옵션에 Claude 구독 요구사항·산출물 형식·예상 소요 단계를 상세 설명으로 포함한다. 이 순서는 REQ-ROUTE-006의 조건이 활성화될 때 override된다.

**REQ-ROUTE-004 (Event-Driven)**  
**When** 사용자가 경로 A(Claude Design)를 선택하면, the 시스템 **shall**:
- Claude.ai 디자인 툴 접근 URL(`https://claude.ai/design`)과 수동 전달 단계(텍스트 브리프 입력 → handoff bundle 다운로드)를 출력하고,
- handoff bundle 파일 경로를 AskUserQuestion으로 재수집하여 `moai-workflow-design-import` 스킬에 전달한다.

**REQ-ROUTE-005 (Event-Driven)**  
**When** 사용자가 경로 B(코드 폴백)를 선택하면, the 시스템 **shall** `moai-domain-copywriting` + `moai-domain-brand-design` + `moai-workflow-gan-loop` 스킬을 로드하여 `expert-frontend` 에이전트에 브리프와 브랜드 컨텍스트를 함께 위임한다.

**REQ-ROUTE-006 (State-Driven, OVERRIDES REQ-ROUTE-003 순서)**  
**While** Claude.ai Pro 이하 구독 플랜 감지 상태(사용자가 명시적으로 신고)에서, the 시스템 **shall** REQ-ROUTE-003의 기본 순서를 뒤집어 경로 B(코드 폴백)를 첫 번째 옵션으로 표시하고 경로 A에 "구독 업그레이드 필요" 경고를 옵션 설명에 포함하며, 경로 A 옵션은 비활성화하지 않는다.

**REQ-ROUTE-007 (Unwanted Behavior)**  
**If** 사용자가 경로 선택 없이 회답하거나 회답 시간이 초과되면, **then** the 시스템 **shall** 자동으로 경로 선택을 진행하지 않고 세션을 유지한 채 재질문한다 (최대 3회, 초과 시 중단).

**REQ-ROUTE-008 (Event-Driven)**  
**When** `expert-frontend` 위임이 완료되면, the 시스템 **shall** `moai-workflow-gan-loop`를 통해 evaluator-active에 결과를 전달하고 GAN 루프(최대 5회 반복, pass threshold 0.75)를 실행한다.

---

### 5.2 REQ-SKILL: 신규 스킬 요구사항

**REQ-SKILL-001 (Ubiquitous) — `moai-domain-copywriting`**  
The `moai-domain-copywriting` 스킬 **shall** 다음 트리거 조건으로 자동 활성화된다:
- 파일 경로가 `.moai/project/brand/brand-voice.md`를 참조할 때
- `/moai design` 경로 B 실행 중
- `copywriting`, `copy`, `microcopy`, `cta`, `headline` 키워드 감지 시

**REQ-SKILL-002 (Ubiquitous) — `moai-domain-copywriting`**  
The `moai-domain-copywriting` 스킬 **shall** 다음 진입(Entry) 조건을 만족할 때 실행을 시작한다:
- 브랜드 보이스 정의(`brand-voice.md`) 존재 또는 인라인 전달됨
- 대상 페이지/섹션 범위 명시됨
- anti-AI-slop 체크리스트(감탄사 과용 금지, 상투어 금지, 구체성 요구) 로드됨

**REQ-SKILL-003 (Event-Driven) — `moai-domain-copywriting`**  
**When** 카피 생성이 완료되면, the 스킬 **shall** 출력물을 JSON 섹션 구조(hero, features, social_proof, cta, footer)로 반환하고, 각 섹션에 대체 변형(A/B variants) 최소 1개를 포함한다.

**REQ-SKILL-004 (Ubiquitous) — `moai-domain-brand-design`**  
The `moai-domain-brand-design` 스킬 **shall** 다음 트리거 조건으로 자동 활성화된다:
- `/moai design` 경로 B 실행 중
- `.moai/project/brand/visual-identity.md` 참조 시
- `design-tokens`, `color-palette`, `typography`, `hero-section`, `wcag` 키워드 감지 시

**REQ-SKILL-005 (Ubiquitous) — `moai-domain-brand-design`**  
The `moai-domain-brand-design` 스킬 **shall** 다음 원칙을 출력 산출물에 강제한다:
- hero-first 레이아웃 (첫 화면 CTA 명확)
- 디자인 토큰 추출 (colors, spacing, typography, radii)
- WCAG AA 명도 대비(최소 4.5:1 본문, 3:1 대형 텍스트)

**REQ-SKILL-006 (Event-Driven) — `moai-domain-brand-design`**  
**When** visual-identity.md에 명시된 컬러 팔레트와 생성된 디자인 토큰이 불일치하면, the 스킬 **shall** 실행을 중단하고 충돌 리포트를 반환한다.

**REQ-SKILL-007 (Ubiquitous) — `moai-workflow-design-import`**  
The `moai-workflow-design-import` 스킬 **shall** Claude Design handoff bundle을 파싱하여 다음 산출물을 생성한다:
- 디자인 토큰 JSON (`.moai/design/tokens.json`)
- 컴포넌트 매니페스트 (`.moai/design/components.json`)
- 정적 에셋 목록 (`.moai/design/assets/`)

**REQ-SKILL-008 (Ubiquitous) — `moai-workflow-design-import`**  
The `moai-workflow-design-import` 스킬 **shall** 다음 입력 형식을 1차 지원한다: ZIP (우선), HTML (우선). 지원되지 않는 형식 입력 시 `DESIGN_IMPORT_UNSUPPORTED_FORMAT` 오류를 반환한다.

> Roadmap(비정상 REQ, §2.2 Exclusions 참조): DOCX/PPTX/PDF/Canva 링크는 Phase 2 후속 릴리스 대상.

**REQ-SKILL-009 (Unwanted Behavior) — `moai-workflow-design-import`**  
**If** handoff bundle 경로가 존재하지 않거나 읽을 수 없으면, **then** the 스킬 **shall** 즉시 오류 코드 `DESIGN_IMPORT_NOT_FOUND`를 반환하고 폴백으로 수동 경로 안내를 출력한다.

**REQ-SKILL-010 (Unwanted Behavior) — `moai-workflow-design-import`**  
**If** bundle 내부에 악성으로 의심되는 파일(실행 가능 바이너리, 스크립트 포함 ZIP 엔트리)이 감지되면, **then** the 스킬 **shall** 파싱을 거부하고 `DESIGN_IMPORT_SECURITY_REJECT` 오류를 반환한다.

**REQ-SKILL-011 (Ubiquitous) — `moai-workflow-gan-loop`**  
The `moai-workflow-gan-loop` 스킬 **shall** 다음 4차원 스코어링을 구현한다:
- Design Quality (시각적 완성도)
- Originality (독창성)
- Completeness (완결성)
- Functionality (기능 작동성)

**REQ-SKILL-012 (State-Driven) — `moai-workflow-gan-loop`**  
**While** harness level이 `thorough`인 상태에서, the 스킬 **shall** Sprint Contract 협상을 필수로 요구한다.

**REQ-SKILL-012a (Optional) — `moai-workflow-gan-loop`**  
**Where** harness level이 `standard`인 환경에서, the 스킬 **shall** Sprint Contract 협상을 선택적으로 지원한다(사용자 명시 opt-in 시 활성화).

**REQ-SKILL-013 (Event-Driven) — `moai-workflow-gan-loop`**  
**When** 반복 개선 폭이 `improvement_threshold`(0.05) 미만인 상태가 연속 2회 발생하면, the 스킬 **shall** 정체 경고를 발생시키고 사용자에게 에스컬레이션 옵션을 제시한다.

**REQ-SKILL-014 (Ubiquitous) — `moai-workflow-gan-loop`**  
The `moai-workflow-gan-loop` 스킬 **shall** `.moai/config/sections/design.yaml`에서 다음 파라미터를 읽는다: `max_iterations`, `pass_threshold`, `escalation_after`, `improvement_threshold`, `strict_mode`, `sprint_contract.*`.

**REQ-SKILL-015 (Unwanted Behavior) — `moai-workflow-design-import` bundle 버전 불일치 (NEW, D6)**  
**If** 감지된 Claude Design handoff bundle의 포맷 버전이 `.moai/config/sections/design.yaml`의 `supported_bundle_versions` 화이트리스트에 포함되지 않으면, **then** the 스킬 **shall** `DESIGN_IMPORT_UNSUPPORTED_VERSION` 오류를 반환하고 stderr에 다음 3가지를 출력한다: (1) 감지된 버전, (2) 지원되는 버전 목록, (3) 경로 B(코드 폴백) 전환 안내.

> 참고 (비규범): 사용자는 오류 수신 후 경로 B로 폴백하거나 다운그레이드된 bundle을 재전달하는 두 가지 해결책을 가진다.

---

### 5.3 REQ-MIGRATE: `moai migrate agency` 커맨드 요구사항

**REQ-MIGRATE-001 (Ubiquitous)**  
The `moai migrate agency` 커맨드 **shall** 실행 전 프로젝트 루트에 `.agency/` 디렉터리가 존재하는지 검증한다.

**REQ-MIGRATE-002 (Unwanted Behavior)**  
**If** `.agency/` 디렉터리가 존재하지 않으면, **then** the 커맨드 **shall** 오류 코드 `MIGRATE_NO_SOURCE`를 반환하고 종료 코드 2로 종료한다.

**REQ-MIGRATE-003 (Unwanted Behavior)**  
**If** `.moai/project/brand/`, `.moai/config/sections/design.yaml`, `.moai/research/observations/` 중 하나라도 이미 존재하면, **then** the 커맨드 **shall** `--force` 플래그 없이는 진행하지 않고 `MIGRATE_TARGET_EXISTS` 오류를 반환한다.

**REQ-MIGRATE-004 (Event-Driven)**  
**When** 마이그레이션이 시작되면, the 커맨드 **shall** 다음 순서로 원자적 작업을 수행한다 (하나라도 실패 시 전체 롤백):
1. `.agency/`를 `.agency.archived/`로 복사 (원본 보존)
2. `.agency/context/` → `.moai/project/brand/` 이전 (파일명 매핑: `brand-voice.md`, `visual-identity.md`, `target-audience.md`)
3. `.agency/config.yaml` → `.moai/config/sections/design.yaml` 변환
4. `.agency/learnings/` → `.moai/research/observations/` 이전
5. `.agency/fork-manifest.yaml` → `.agency.archived/fork-manifest.yaml`로만 보존
6. **조건부 실행**: 사용자 프로젝트 루트의 `.claude/rules/agency/constitution.md`(사용자 복사본)이 존재하는 경우에만 `.claude/rules/moai/design/constitution.md`로 재배치한다. 템플릿 계층(`internal/template/templates/.claude/rules/moai/design/constitution.md`)은 M1 템플릿 배포로 이미 제공되므로 이 Step에서 건드리지 않는다. 사용자 복사본이 없으면 Step 6은 no-op로 처리하고 trace 로그에만 기록한다.

**REQ-MIGRATE-005 (State-Driven)**  
**While** 마이그레이션이 진행 중인 상태에서, the 커맨드 **shall** 각 단계 완료 시 진행률(N/6)과 이전된 파일 경로를 stdout에 기록한다.

**REQ-MIGRATE-006 (Unwanted Behavior)**  
**If** 마이그레이션 중 임의 단계에서 오류가 발생하면, **then** the 커맨드 **shall**:
- `.agency.archived/`에서 `.agency/`로 복원 시도
- 이미 생성된 `.moai/` 쪽 대상 파일 삭제 (트랜잭션 로그 기반)
- 종료 코드 1과 `MIGRATE_ROLLBACK_OK` 또는 `MIGRATE_ROLLBACK_FAILED` 반환

**REQ-MIGRATE-007 (Event-Driven)**  
**When** 마이그레이션이 성공적으로 완료되면, the 커맨드 **shall** 다음을 stdout에 출력한다:
- 이전된 파일 수
- `.agency.archived/` 절대 경로 (복구 참조용)
- 후속 가이드(`/moai design` 실행 방법)

**REQ-MIGRATE-008 (Ubiquitous)**  
The `moai migrate agency` 커맨드 **shall** `--dry-run` 플래그를 지원하여 실제 파일 변경 없이 예상 작업 내역만 출력한다.

**REQ-MIGRATE-009 (Ubiquitous)**  
The `moai migrate agency` 커맨드 **shall** `--force` 플래그를 지원하여 REQ-MIGRATE-003의 대상 파일 존재 검증을 건너뛴다. 이 경우에도 `.agency.archived/` 백업은 필수다.

**REQ-MIGRATE-010 (Unwanted Behavior)**  
**If** `--force` 없이 `.agency.archived/`가 이미 존재하면, **then** the 커맨드 **shall** 덮어쓰지 않고 `MIGRATE_ARCHIVE_EXISTS` 오류를 반환한다.

**REQ-MIGRATE-011 (Ubiquitous)**  
The `moai migrate agency` 커맨드 **shall** 디스크 공간 가용성을 사전에 확인하고(최소 `.agency/` 크기의 2배), 부족 시 `MIGRATE_DISK_FULL` 오류로 중단한다.

**REQ-MIGRATE-012a (State-Driven) — POSIX 플랫폼**  
**While** 실행 환경이 POSIX 호환(macOS 또는 Linux)인 상태에서, the `moai migrate agency` 커맨드 **shall** 모든 파일 복사 시 Unix 권한 비트(0o7777 마스크)를 `os.Chmod`로 원본과 일치하게 보존한다.

**REQ-MIGRATE-012b (State-Driven) — Windows 플랫폼**  
**While** 실행 환경이 Windows인 상태에서, the `moai migrate agency` 커맨드 **shall** Unix 권한 비트 보존을 no-op로 처리하고, Windows ACL은 수정하지 않으며, stderr에 일회성 안내 메시지 "Windows: Unix permission bits not applicable, ACL preserved as-is"를 출력한다(CLAUDE.local.md lessons.md #7 참조).

**REQ-MIGRATE-013 (Unwanted Behavior) — SIGINT/SIGTERM 처리 (NEW, D7)**  
**If** 마이그레이션 실행 중 SIGINT 또는 SIGTERM 시그널이 수신되면, **then** the 커맨드 **shall** 현재 진행 중인 Phase를 완료한 후 트랜잭션 로그(`~/.moai/.migrate-tx-<timestamp>.json`)에 체크포인트를 flush하고, 롤백을 시도하며, 종료 코드 130(SIGINT) 또는 143(SIGTERM)으로 종료한다. 재실행 시 `moai migrate agency --resume <tx-id>` 플래그로 체크포인트에서 복구 가능하다.

---

### 5.4 REQ-DIR: 디렉터리 매핑 요구사항

**REQ-DIR-001 (Ubiquitous)**  
The 마이그레이션 매핑표 **shall** 다음 규칙을 준수한다:

| Source (`.agency/`) | Target (`.moai/`) | 변환 유형 |
|---|---|---|
| `context/brand-voice.md` | `project/brand/brand-voice.md` | 이동 |
| `context/visual-identity.md` | `project/brand/visual-identity.md` | 이동 |
| `context/target-audience.md` | `project/brand/target-audience.md` | 이동 |
| `context/tech-preferences.md` | `project/tech.md` (append) | 병합 |
| `config.yaml` | `config/sections/design.yaml` | 구조 변환 |
| `learnings/*.md` | `research/observations/*.md` | 이동 |
| `sprints/*` | `research/sprints/*` | 이동 |
| `evolution/changelog.md` | `research/evolution-log.md` | 이동 |
| `fork-manifest.yaml` | (아카이브만 유지) | 보존 |

**REQ-DIR-002 (Unwanted Behavior)**  
**If** `tech-preferences.md`와 `tech.md`가 동시에 존재하고 내용 충돌이 감지되면, **then** the 커맨드 **shall** 병합을 보류하고 `MIGRATE_MERGE_CONFLICT` 경고를 출력하며, 두 파일을 모두 보존한 상태로 사용자 수동 병합을 유도한다.

**REQ-DIR-003 (Ubiquitous)**  
The `.moai/project/brand/` 디렉터리 **shall** 첫 실행 시 3개 템플릿(`brand-voice.md`, `visual-identity.md`, `target-audience.md`)의 빈 스켈레톤을 `internal/template/templates/.moai/project/brand/`에서 복사한다.

---

### 5.5 REQ-DETECT: 기존 `.agency/` 감지 UX

**REQ-DETECT-001 (Event-Driven)**  
**When** `moai doctor` 또는 `moai update`가 실행되고 `.agency/` 디렉터리가 감지되면, the 시스템 **shall** 마이그레이션 권장 메시지를 출력하고 `moai migrate agency` 커맨드 예시를 제공한다.

**REQ-DETECT-002 (Event-Driven)**  
**When** Claude Code 세션 시작 시 `.agency/` 디렉터리가 감지되고 `.moai/project/brand/`가 부재하면, the SessionStart 훅 **shall** 일회성 공지(최대 1회/세션)를 출력한다.

**REQ-DETECT-003 (Optional)**  
**Where** `.agency/` 디렉터리가 감지된 환경에서, the `/moai design` 커맨드 **shall** 실행 전 마이그레이션 필요성을 알리는 경고를 출력한다.

---

### 5.6 REQ-DEPRECATE: `/agency` 커맨드 라이프사이클

**REQ-DEPRECATE-001 (Ubiquitous)**  
The `/agency` 커맨드 **shall** 흡수 릴리스(vN)부터 deprecation 단계 1에 진입한다:
- 호출 시 표준 경고 메시지 출력
- `/moai design` 래퍼로 동작 위임
- 기능 유지

**REQ-DEPRECATE-002 (Event-Driven)**  
**When** `/agency <subcommand>`가 호출되면, the 래퍼 **shall**:
- 정확히 1개의 경고 라인을 stderr에 출력("`/agency` is deprecated, use `/moai design` instead")
- 지원되는 서브커맨드의 경우 `/moai design`의 대응 동작으로 리매핑
- 미지원 서브커맨드(예: `learn`, `evolve`)의 경우 `AGENCY_SUBCOMMAND_UNSUPPORTED` 오류를 반환하고 마이그레이션 가이드 URL 출력

**REQ-DEPRECATE-003 (Ubiquitous)**  
The `/agency` 커맨드 **shall** 흡수 릴리스 이후 최소 2 마이너 버전 주기 동안 deprecation 단계를 유지한 후 완전 제거된다.

**REQ-DEPRECATE-004 (Unwanted Behavior)**  
**If** 릴리스 노트에 deprecation 경고가 명시되지 않으면, **then** CI **shall** 릴리스 머지를 차단한다.

---

### 5.7 REQ-CONST: Constitution 및 FROZEN/EVOLVABLE 보존

**REQ-CONST-001 (Ubiquitous)**  
The `.claude/rules/agency/constitution.md` **shall** `.claude/rules/moai/design/constitution.md`로 재배치되며, FROZEN zone 및 EVOLVABLE zone 정의는 변경 없이 유지된다.

**REQ-CONST-002 (Ubiquitous)**  
The 재배치된 constitution 문서 **shall** 첫 줄에 재배치 히스토리(원본 경로, 이전 버전, 이전 일자)를 HISTORY 섹션으로 명시한다.

**REQ-CONST-003 (Unwanted Behavior)**  
**If** 재배치 후 Section 5 Safety Architecture 전체(특히 Layer 5 Human Oversight 포함)의 FROZEN zone 내용이 수정되었음이 감지되면, **then** CI **shall** 해당 변경을 차단하고 수동 구성법 승인 프로세스를 요구한다.

**REQ-CONST-004 (Ubiquitous)**  
The `moai-workflow-gan-loop` 스킬 **shall** GAN Loop Contract(Section 11), Sprint Contract Protocol, Evaluator Leniency Prevention(Section 12)을 원본 그대로 구현한다.

---

### 5.8 REQ-BRIEF: manager-spec BRIEF 확장

**REQ-BRIEF-001 (Ubiquitous)**  
The `manager-spec` 에이전트 **shall** `/moai design` 호출 시 BRIEF 섹션을 생성하고, 다음 3개 하위 항목을 필수로 포함한다: Goal, Audience, Brand.

**REQ-BRIEF-002 (Event-Driven)**  
**When** BRIEF 섹션의 Brand 항목이 비어 있으면, the 에이전트 **shall** `.moai/project/brand/` 3개 파일의 핵심 내용을 자동 주입한다.

**REQ-BRIEF-003 (Unwanted Behavior)**  
**If** `.moai/project/brand/` 디렉터리가 부재하면, **then** the 에이전트 **shall** BRIEF 생성을 중단하고 브랜드 인터뷰 수행을 요구한다.

---

### 5.9 REQ-FALLBACK: Pro 이하 사용자 폴백 UX

**REQ-FALLBACK-001 (State-Driven)**  
**While** 사용자가 Claude Design 미가용 상태라고 명시적으로 신고한 상태에서, the 시스템 **shall** 경로 B(코드 폴백)를 기본 경로로 사용하며 경로 A를 회색 처리한다.

**REQ-FALLBACK-002 (Ubiquitous)**  
The 경로 B **shall** 경로 A와 동등한 산출물 범주(디자인 토큰, 컴포넌트 구조, 정적 에셋 목록)를 생성해야 한다. 품질은 GAN 루프 pass threshold로 검증한다.

**REQ-FALLBACK-003 (Optional)**  
**Where** 사용자가 Figma 계정을 연결하고 `.moai/config/sections/design.yaml`에 `figma.enabled: true`를 명시한 환경에서, the `moai-domain-brand-design` 스킬 **shall** 공개 Figma 파일 URL에서 디자인 토큰을 추출하는 보조 모드를 제공한다.

> 본 REQ는 Phase 2 릴리스 대상이며, 현 릴리스의 DoD에는 `figma.enabled` 기본값이 `false`인 상태가 포함된다(§2.2 Exclusions 및 acceptance.md DoD 정합).

---

### 5.10 REQ-REMOVE: agency 에이전트/스킬 제거

**REQ-REMOVE-001 (Ubiquitous)**  
The 흡수 릴리스 **shall** 다음 agency 에이전트를 `.claude/agents/agency/`에서 제거한다: `planner.md`, `designer.md`, `builder.md`, `evaluator.md`.

**REQ-REMOVE-002 (Ubiquitous)**  
The 흡수 릴리스 **shall** `copywriter.md`(→ `moai-domain-copywriting` 스킬, REQ-SKILL-001 ~ REQ-SKILL-003 참조) 및 `learner.md`(→ `moai-workflow-research` 스킬) 두 에이전트의 내용을 신규 스킬로 흡수한 뒤 두 파일을 `.claude/agents/agency/`에서 삭제한다.

**REQ-REMOVE-003 (Unwanted Behavior)**  
**If** 제거된 에이전트 이름을 참조하는 문서·스킬·커맨드가 저장소에 남아 있으면, **then** CI의 `grep` 기반 감사 **shall** 머지를 차단한다.

**REQ-REMOVE-004 (Ubiquitous)**  
The 흡수 릴리스 **shall** `.claude/skills/agency-frontend-patterns/`를 `.claude/skills/moai-domain-frontend/`로 병합하고 fork-manifest 참조를 제거한다.

---

## 6. Acceptance Criteria (수용 기준 요약)

상세 Given-When-Then 시나리오는 `acceptance.md` 참조.

핵심 기준:
- `moai migrate agency`: 기존 `.agency/` 데이터 0% 손실, 롤백 가능
- `/moai design`: 경로 A/B 분기 정상 동작
- 제거된 에이전트 참조 0건 (저장소 전체 `grep` 검증)
- `go test ./...` 전체 통과
- `internal/template/commands_audit_test.go` 통과 (thin command pattern)
- `make build` 성공 및 embedded.go 갱신
- `/agency` deprecation 경고가 모든 서브커맨드에서 출력됨

---

## 7. Constraints (제약)

- 도구 구현 언어(Go)와 사용자 프로젝트 언어는 분리. 템플릿(`internal/template/templates/`)은 16개 언어 중립성 유지 (CLAUDE.local.md §15)
- 하드코딩 금지 (CLAUDE.local.md §14): URL, 모델명, 임계값은 `const` 또는 `.moai/config/sections/design.yaml`에서 로드
- Claude Design 수동 경로는 AskUserQuestion으로 명시적 안내 필수
- `.agency/` 직접 삭제 금지, `.agency.archived/` 백업 필수
- 마이그레이션은 원자적(all-or-nothing)이어야 하며 부분 실패 시 완전 롤백
- 새 파일 추가 시 Template-First 원칙 준수(CLAUDE.local.md §2): `internal/template/templates/`에 먼저 배치 후 `make build`

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| Claude Design 사양 변경으로 handoff bundle 구조가 바뀜 | 경로 A 실패 | REQ-SKILL-015가 `supported_bundle_versions` 화이트리스트 검증으로 사전 차단 |
| `fork-manifest.yaml` 참조 남은 코드 발견 | CI 실패 | REQ-REMOVE-003 grep 감사로 사전 차단 |
| 마이그레이션 중 디스크 가득 참 | 데이터 유실 | REQ-MIGRATE-011 사전 검사 |
| 마이그레이션 중 SIGINT/SIGTERM 수신 | 부분 이전 상태 | REQ-MIGRATE-013 체크포인트 및 `--resume` 플래그 |
| 기존 `/agency` 사용자 혼란 | UX 저하 | REQ-DEPRECATE-001 ~ 003 라이프사이클 준수 |
| Pro 이하 사용자의 경로 A 기대치 | 혼선 | REQ-ROUTE-006 순서 override + REQ-FALLBACK-001 명시적 구독 경고 |
| `.moai/project/brand/` 병합 충돌 (tech-preferences.md ↔ tech.md) | 내용 손실 | REQ-DIR-002 수동 병합 유도 |
| Windows ACL과 POSIX 권한 semantics 차이 | 플랫폼별 오동작 | REQ-MIGRATE-012a/b 플랫폼 분리 |

---

## 9. Dependencies (의존성)

- 기존 SPEC-AGENCY-001 (현 agency 시스템 정의) — 본 SPEC이 폐기 경로 제공
- 기존 SPEC-DESIGN-001 (디자인 SPEC) — 관련 스킬 정의 참고
- Claude Code v2.1.111+ (Opus 4.7, effortLevel 지원)
- `manager-spec`, `expert-frontend`, `evaluator-active` 에이전트 (기존)

---

## 10. Traceability (추적성)

- 본 SPEC의 모든 요구사항 ID는 `plan.md` 마일스톤과 `acceptance.md` 시나리오로 역참조됨
- 구현 시 각 소스 파일에 `@SPEC:SPEC-AGENCY-ABSORB-001:REQ-<CATEGORY>-<NNN>` 주석 부착
- 총 REQ 개수: 62개 (ROUTE 8, SKILL 16 [001-014 + 012a + 015], MIGRATE 14 [001-011 + 012a/012b + 013], DIR 3, DETECT 3, DEPRECATE 4, CONST 4, BRIEF 3, FALLBACK 3, REMOVE 4). 이전 iteration 1 감사 시점에는 58개였으며, 수정 과정에서 REQ-MIGRATE-012 분리(+1), REQ-MIGRATE-013 신규(+1), REQ-SKILL-012a 신규(+1), REQ-SKILL-015 신규(+1)로 +4 증가하여 총 62개가 되었다.
- 코드 구현 파일 예상 경로:
  - `internal/cli/migrate_agency.go` (REQ-MIGRATE-*)
  - `internal/cli/migrate_agency_test.go` (REQ-MIGRATE-* 검증)
  - `internal/cli/migrate_agency_windows_test.go` (REQ-MIGRATE-012b Windows 전용)
  - `internal/cli/migrate_agency_posix_test.go` (REQ-MIGRATE-012a POSIX 전용, build tag: `//go:build !windows`)
  - `.claude/skills/moai/workflows/design.md` (REQ-ROUTE-*)
  - `.claude/skills/moai-domain-copywriting/SKILL.md` (REQ-SKILL-001 ~ 003)
  - `.claude/skills/moai-domain-brand-design/SKILL.md` (REQ-SKILL-004 ~ 006)
  - `.claude/skills/moai-workflow-design-import/SKILL.md` (REQ-SKILL-007 ~ 010, 015)
  - `.claude/skills/moai-workflow-gan-loop/SKILL.md` (REQ-SKILL-011 ~ 014, 012a)
  - `.claude/rules/moai/design/constitution.md` (REQ-CONST-*)

---

End of SPEC.
