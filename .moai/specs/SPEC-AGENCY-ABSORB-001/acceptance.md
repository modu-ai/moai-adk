---
id: SPEC-AGENCY-ABSORB-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-20
author: GOOS
priority: High
labels: [agency, migration, design, hybrid, absorption]
---

# SPEC-AGENCY-ABSORB-001: Acceptance Criteria

## HISTORY

- 2026-04-20 v0.1.0: 최초 작성
- 2026-04-20 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. YAML frontmatter 필수 필드 추가(D2), "(암시)" 5개 AC를 명시적 시나리오로 전환(D14), AC-MIGRATE-010(SIGINT) 신규, AC-SKILL-010(bundle 버전 불일치) 신규, AC-MIGRATE-011a/b(플랫폼 권한) 분할, AC-CONST-002 섹션 참조 정리(D12), AC-FALLBACK-003 신규로 REQ-FALLBACK-003 DoD 포함(D8), traceability 매트릭스에서 "(암시)" 0건 달성.

---

## 1. 개요

본 문서는 SPEC-AGENCY-ABSORB-001의 수용 기준을 Given-When-Then 형식으로 정의한다. 각 시나리오는 spec.md의 REQ-* ID와 추적 가능하며, 테스트 케이스로 직접 변환된다.

---

## 2. 시나리오 (Scenarios)

### 2.1 마이그레이션 커맨드 (REQ-MIGRATE-*, REQ-DIR-*)

#### AC-MIGRATE-001: 성공적인 전체 마이그레이션
**Covers**: REQ-MIGRATE-001, REQ-MIGRATE-004, REQ-MIGRATE-005, REQ-MIGRATE-007, REQ-DIR-001

- **Given** 프로젝트 루트에 `.agency/` 디렉터리가 존재하고 `.agency/context/brand-voice.md`, `.agency/context/visual-identity.md`, `.agency/context/target-audience.md`, `.agency/config.yaml`, `.agency/learnings/LEARN-001.md`가 포함되어 있으며,
- **And** `.moai/project/brand/`, `.moai/config/sections/design.yaml`, `.moai/research/observations/`는 부재하며,
- **When** 사용자가 `moai migrate agency`를 실행하면,
- **Then** 커맨드는 종료 코드 0으로 종료되고,
- **And** `.moai/project/brand/brand-voice.md`, `visual-identity.md`, `target-audience.md`가 생성되어 원본과 바이트 단위로 동일하며,
- **And** `.moai/config/sections/design.yaml`이 생성되고 `gan_loop`, `evolution`, `pipeline` 키가 존재하며,
- **And** `.moai/research/observations/LEARN-001.md`가 생성되며,
- **And** `.agency.archived/` 백업이 존재하고 원본과 동일한 파일 트리를 포함하며,
- **And** 모든 복사된 파일의 권한 비트가 원본과 일치하며,
- **And** stdout에 진행률(6/6)과 `.agency.archived/` 절대 경로가 표시된다.

#### AC-MIGRATE-002: Dry-run 모드
**Covers**: REQ-MIGRATE-008

- **Given** 마이그레이션 대상 환경이 AC-MIGRATE-001과 동일하며,
- **When** 사용자가 `moai migrate agency --dry-run`을 실행하면,
- **Then** 커맨드는 종료 코드 0으로 종료되고,
- **And** `.agency.archived/`, `.moai/project/brand/`, `.moai/config/sections/design.yaml`이 **생성되지 않으며**,
- **And** stdout에 예상 작업 내역(이동/변환/병합)이 표시되고,
- **And** 원본 `.agency/`는 변경되지 않는다.

#### AC-MIGRATE-003: 원본 부재 시 오류
**Covers**: REQ-MIGRATE-002

- **Given** `.agency/` 디렉터리가 존재하지 않으며,
- **When** 사용자가 `moai migrate agency`를 실행하면,
- **Then** 커맨드는 종료 코드 2로 종료되고,
- **And** stderr에 `MIGRATE_NO_SOURCE` 오류 코드와 사람이 읽기 쉬운 메시지가 출력된다.

#### AC-MIGRATE-004: 대상 이미 존재 시 force 없이 거부
**Covers**: REQ-MIGRATE-003

- **Given** `.agency/`가 존재하고,
- **And** `.moai/project/brand/brand-voice.md`가 이미 존재하며,
- **When** 사용자가 `moai migrate agency`를 실행하면,
- **Then** 커맨드는 종료 코드 1로 종료되고,
- **And** stderr에 `MIGRATE_TARGET_EXISTS` 오류와 `--force` 플래그 안내가 출력되며,
- **And** `.moai/`와 `.agency/`는 변경되지 않는다.

#### AC-MIGRATE-005: Force 플래그 덮어쓰기
**Covers**: REQ-MIGRATE-009, REQ-MIGRATE-010

- **Given** `.agency/`가 존재하고,
- **And** `.moai/project/brand/brand-voice.md`가 이미 존재하며,
- **And** `.agency.archived/`는 부재하며,
- **When** 사용자가 `moai migrate agency --force`를 실행하면,
- **Then** 커맨드는 종료 코드 0으로 종료되고,
- **And** 기존 `.moai/project/brand/brand-voice.md`가 `.agency/context/brand-voice.md`의 내용으로 덮어쓰여지며,
- **And** `.agency.archived/` 백업이 생성된다.

#### AC-MIGRATE-006: Archive 이미 존재 시 거부
**Covers**: REQ-MIGRATE-010

- **Given** `.agency/`가 존재하고,
- **And** `.agency.archived/`가 이미 존재하며,
- **When** 사용자가 `moai migrate agency`(force 없음)를 실행하면,
- **Then** 커맨드는 종료 코드 1로 종료되고,
- **And** stderr에 `MIGRATE_ARCHIVE_EXISTS` 오류가 출력된다.

#### AC-MIGRATE-007: 중간 실패 시 롤백
**Covers**: REQ-MIGRATE-006

- **Given** 마이그레이션 중 Phase 3(learnings 이전) 실행 도중 디스크 쓰기 오류가 발생하도록 모킹된 테스트 환경에서,
- **When** `moai migrate agency`를 실행하면,
- **Then** 커맨드는 종료 코드 1로 종료되고,
- **And** stderr에 `MIGRATE_ROLLBACK_OK` 코드가 출력되며,
- **And** `.agency/`는 원래 상태로 완전히 복원되고,
- **And** 실행 전에 생성되었던 `.moai/project/brand/`, `.moai/config/sections/design.yaml` 등 부분 산출물은 모두 제거된다.

#### AC-MIGRATE-008: 디스크 공간 부족
**Covers**: REQ-MIGRATE-011

- **Given** `.agency/` 크기가 100MB이고 가용 디스크 공간이 150MB(2배 미만)인 환경에서,
- **When** `moai migrate agency`를 실행하면,
- **Then** 커맨드는 종료 코드 1로 종료되고,
- **And** stderr에 `MIGRATE_DISK_FULL` 오류가 출력되며,
- **And** `.moai/`와 `.agency/`는 변경되지 않는다.

#### AC-MIGRATE-009: tech-preferences 병합 충돌
**Covers**: REQ-DIR-002

- **Given** `.agency/context/tech-preferences.md`가 "Framework: Next.js"를 선언하고,
- **And** `.moai/project/tech.md`가 "Framework: Remix"를 선언하며(충돌),
- **When** `moai migrate agency`를 실행하면,
- **Then** 커맨드는 종료 코드 1로 종료되고,
- **And** stderr에 `MIGRATE_MERGE_CONFLICT` 경고가 출력되며,
- **And** `tech-preferences.md`와 `tech.md` 두 파일이 모두 보존되고,
- **And** 나머지 마이그레이션 단계는 롤백된다.

#### AC-MIGRATE-010: SIGINT/SIGTERM 체크포인트 (NEW, REQ-MIGRATE-013)
**Covers**: REQ-MIGRATE-013

- **Given** `.agency/`가 100MB 분량이고 마이그레이션 Phase 3 실행 중인 상태에서,
- **When** 외부 프로세스가 SIGINT(Ctrl+C)를 커맨드에 전송하면,
- **Then** 커맨드는 현재 진행 중인 파일 복사를 완료한 후,
- **And** `~/.moai/.migrate-tx-<timestamp>.json`에 `{phase: 3, completed_files: [...], remaining: [...]}` 체크포인트를 flush하고,
- **And** 롤백을 시도하며 종료 코드 130(SIGINT)으로 종료하며,
- **And** `moai migrate agency --resume <tx-id>`로 재실행 시 체크포인트의 `remaining` 항목부터 이어서 진행된다.

#### AC-MIGRATE-011a: POSIX 권한 비트 보존 (NEW, REQ-MIGRATE-012a)
**Covers**: REQ-MIGRATE-012a

- **Given** macOS 또는 Linux 환경에서,
- **And** `.agency/context/brand-voice.md`가 `0o640` 권한을 가진 상태에서,
- **When** `moai migrate agency`가 성공적으로 실행되면,
- **Then** 이전된 `.moai/project/brand/brand-voice.md`의 권한 비트가 `0o640`과 일치한다 (`os.Stat` 검증).

#### AC-MIGRATE-011b: Windows 권한 no-op (NEW, REQ-MIGRATE-012b)
**Covers**: REQ-MIGRATE-012b

- **Given** Windows 환경에서,
- **When** `moai migrate agency`가 성공적으로 실행되면,
- **Then** stderr에 "Windows: Unix permission bits not applicable, ACL preserved as-is" 일회성 메시지가 출력되고,
- **And** 이전된 파일에 대해 `os.Chmod` 호출이 없으며(로그 확인),
- **And** 파일 내용이 원본과 바이트 단위로 동일하다.

#### AC-MIGRATE-012: 템플릿 브랜드 스켈레톤 배포 (NEW, REQ-DIR-003 명시 커버)
**Covers**: REQ-DIR-003

- **Given** `.agency/`가 존재하지 않고 `.moai/project/brand/`도 부재한 신규 프로젝트에서,
- **When** 사용자가 `moai init` 또는 `moai update`로 `.moai/project/brand/` 디렉터리 초기화를 요청하면,
- **Then** `.moai/project/brand/brand-voice.md`, `visual-identity.md`, `target-audience.md` 3개 스켈레톤 파일이 `internal/template/templates/.moai/project/brand/`로부터 복사되며,
- **And** 각 파일의 내용은 템플릿 원본과 바이트 단위로 동일하다.

---

### 2.2 `/moai design` 라우팅 (REQ-ROUTE-*, REQ-FALLBACK-*)

#### AC-ROUTE-001: 브랜드 컨텍스트 확인 및 경로 선택
**Covers**: REQ-ROUTE-001, REQ-ROUTE-002, REQ-ROUTE-003

- **Given** `.moai/project/brand/` 3개 파일이 완비된 프로젝트에서,
- **When** 사용자가 `/moai design "SaaS 랜딩 페이지"`를 호출하면,
- **Then** 시스템은 AskUserQuestion을 호출하고,
- **And** 첫 번째 옵션은 "Claude Design 위임 (권장)"로 표시되며 Claude.ai Pro 이상 구독 요구사항을 설명하고,
- **And** 두 번째 옵션은 "코드 기반 폴백"으로 표시되며 예상 산출물(토큰/컴포넌트/에셋)을 설명한다.

#### AC-ROUTE-002: 경로 A 선택 시 handoff bundle 경로 수집
**Covers**: REQ-ROUTE-004

- **Given** AC-ROUTE-001의 경로 선택 프롬프트가 표시된 상태에서,
- **When** 사용자가 경로 A(Claude Design)를 선택하면,
- **Then** 시스템은 Claude.ai `/design` 접근 URL과 수동 단계(브리프 입력 → bundle 다운로드)를 출력하고,
- **And** 이어서 AskUserQuestion으로 "다운로드한 bundle 파일 경로"를 수집하며,
- **And** 수집된 경로는 `moai-workflow-design-import` 스킬에 전달된다.

#### AC-ROUTE-003: 경로 B 선택 시 expert-frontend 위임
**Covers**: REQ-ROUTE-005, REQ-ROUTE-008

- **Given** 경로 선택 프롬프트가 표시된 상태에서,
- **When** 사용자가 경로 B(코드 폴백)를 선택하면,
- **Then** 시스템은 `moai-domain-copywriting`, `moai-domain-brand-design`, `moai-workflow-gan-loop` 스킬을 로드하고,
- **And** `expert-frontend` 에이전트를 호출하며, 브리프 + 브랜드 컨텍스트 + 3개 스킬 참조가 프롬프트에 포함되고,
- **And** 1회 이상의 GAN 루프 반복 후 결과가 반환된다(threshold 0.75 충족 또는 max_iterations 도달).

#### AC-ROUTE-004: Pro 이하 사용자 경고
**Covers**: REQ-ROUTE-006, REQ-FALLBACK-001

- **Given** 사용자가 `.moai/config/sections/user.yaml`에 `subscription.tier: "pro-or-below"`를 명시한 환경에서,
- **When** `/moai design`을 호출하면,
- **Then** AskUserQuestion에서 경로 A 옵션 설명에 "구독 업그레이드 필요" 경고 문구가 포함되며,
- **And** 경로 B가 기본(첫 번째) 선택으로 재지정된다.

#### AC-ROUTE-005: 응답 시간 초과 시 재질문
**Covers**: REQ-ROUTE-007

- **Given** AskUserQuestion이 표시된 상태에서,
- **When** 사용자가 빈 응답을 3회 연속 제출하면,
- **Then** 시스템은 3회째에 세션을 중단하고, 종료 없이 "선택이 확인되지 않음, 나중에 다시 시도" 메시지를 출력한다.

#### AC-ROUTE-006: 브랜드 컨텍스트 부재 시 인터뷰 제안
**Covers**: REQ-ROUTE-001, REQ-BRIEF-003

- **Given** `.moai/project/brand/` 디렉터리가 부재한 프로젝트에서,
- **When** 사용자가 `/moai design`을 호출하면,
- **Then** 시스템은 경로 선택 대신 먼저 브랜드 인터뷰 수행을 제안하고,
- **And** 사용자가 동의하면 `manager-spec` BRIEF 플로우로 위임되어 3개 파일이 생성된다.

#### AC-ROUTE-007: BRIEF 섹션 3개 하위 항목 필수 (NEW, REQ-BRIEF-001 명시 커버)
**Covers**: REQ-BRIEF-001

- **Given** `.moai/project/brand/` 3개 파일이 완비된 프로젝트에서,
- **When** 사용자가 `/moai design "<brief>"`를 호출하여 `manager-spec`이 BRIEF 섹션을 생성하면,
- **Then** 생성된 BRIEF 문서의 마크다운 본문은 `## Goal`, `## Audience`, `## Brand` 3개 헤더를 모두 포함하고,
- **And** 각 헤더 하위에 최소 1줄 이상의 본문이 존재하며(빈 섹션 금지),
- **And** 어느 하나라도 누락 시 `manager-spec`은 `BRIEF_SECTION_INCOMPLETE` 오류로 생성을 거부한다.

#### AC-ROUTE-008: BRIEF Brand 항목 자동 주입 (NEW, REQ-BRIEF-002 명시 커버)
**Covers**: REQ-BRIEF-002

- **Given** `.moai/project/brand/brand-voice.md`에 "tone: playful, professional"이 정의되고,
- **And** 사용자가 제공한 브리프에 Brand 섹션이 누락된 상태에서,
- **When** `manager-spec`이 BRIEF 섹션을 생성하면,
- **Then** 생성된 BRIEF의 `## Brand` 섹션은 `brand-voice.md`, `visual-identity.md`, `target-audience.md` 3개 파일의 핵심 내용을 자동 주입하여 포함하며,
- **And** 주입된 내용의 소스 파일 경로가 각 인용 아래에 주석으로 명시된다(`> source: .moai/project/brand/brand-voice.md`).

---

### 2.3 신규 스킬 (REQ-SKILL-*)

#### AC-SKILL-010: `moai-domain-brand-design` 자동 활성화 (NEW, REQ-SKILL-004 명시 커버)
**Covers**: REQ-SKILL-004

- **Given** `/moai design` 경로 B가 실행 중이고,
- **And** `.moai/project/brand/visual-identity.md`가 존재하는 프로젝트에서,
- **When** 브리프에 `color-palette`, `typography` 키워드가 포함되어 전송되면,
- **Then** `moai-domain-brand-design` 스킬 메타데이터가 context에 로드되고,
- **And** 스킬 트리거 로그에 `activation_reason: keyword_match|path_match`가 기록되며,
- **And** 스킬 미로드 상태에서 경로 B 실행 시 `BRAND_DESIGN_SKILL_NOT_ACTIVE` 오류가 발생한다.

#### AC-SKILL-011: `moai-workflow-design-import` bundle 버전 불일치 (NEW, REQ-SKILL-015)
**Covers**: REQ-SKILL-015

- **Given** `.moai/config/sections/design.yaml`에 `supported_bundle_versions: ["v1", "v1.1"]`이 정의되어 있고,
- **And** 사용자가 `v2.0` 버전 메타데이터를 가진 handoff bundle ZIP 파일을 전달하는 환경에서,
- **When** `moai-workflow-design-import` 스킬이 실행되면,
- **Then** 스킬은 `DESIGN_IMPORT_UNSUPPORTED_VERSION` 오류를 반환하고,
- **And** stderr에는 (1) 감지된 버전 "v2.0", (2) 지원되는 버전 목록 "v1, v1.1", (3) 경로 B 폴백 안내 3가지가 모두 출력되며,
- **And** `.moai/design/` 하위에 부분 생성 파일이 남지 않는다.

#### AC-SKILL-001: `moai-domain-copywriting` 자동 활성화
**Covers**: REQ-SKILL-001, REQ-SKILL-002

- **Given** `/moai design` 경로 B 실행 중이고,
- **And** 브랜드 보이스 파일이 로드된 상태에서,
- **When** 카피 생성 작업이 시작되면,
- **Then** `moai-domain-copywriting` 스킬 메타데이터가 context에 로드되고,
- **And** anti-AI-slop 체크리스트 항목(감탄사 과용 금지, 상투어 금지, 구체성 요구)이 활성 규칙으로 표시된다.

#### AC-SKILL-002: 카피 출력 JSON 구조
**Covers**: REQ-SKILL-003

- **Given** `moai-domain-copywriting`이 실행 중이고,
- **When** 카피 생성이 완료되면,
- **Then** 출력은 JSON 형식이며 최상위 키로 `hero`, `features`, `social_proof`, `cta`, `footer`를 포함하고,
- **And** 각 섹션 키에는 `primary`와 `variant_a` 두 필드가 모두 존재한다.

#### AC-SKILL-003: `moai-domain-brand-design` WCAG 강제
**Covers**: REQ-SKILL-005, REQ-SKILL-006

- **Given** visual-identity.md의 primary color = `#FFFFFF`(흰색)이고,
- **When** 스킬이 본문 텍스트 컬러로 `#EEEEEE`(흰색에 매우 근접)를 제안하면,
- **Then** 스킬은 생성을 거부하고 WCAG AA 대비(4.5:1) 위반 오류를 반환한다.

#### AC-SKILL-004: design-import ZIP 파싱
**Covers**: REQ-SKILL-007, REQ-SKILL-008

- **Given** 유효한 Claude Design handoff bundle ZIP 파일 경로가 제공되고,
- **When** `moai-workflow-design-import`가 실행되면,
- **Then** `.moai/design/tokens.json`이 생성되고 `colors`, `spacing`, `typography` 키를 포함하며,
- **And** `.moai/design/components.json`이 생성되고 컴포넌트 목록이 배열로 기록되며,
- **And** `.moai/design/assets/` 하위에 이미지/폰트 파일이 복사된다.

#### AC-SKILL-005: design-import 파일 부재 오류
**Covers**: REQ-SKILL-009

- **Given** 존재하지 않는 경로 `/tmp/missing-bundle.zip`이 제공될 때,
- **When** 스킬이 실행되면,
- **Then** 즉시 `DESIGN_IMPORT_NOT_FOUND` 오류가 반환되고,
- **And** 수동 경로 안내("경로 B로 전환하시겠습니까?")가 출력된다.

#### AC-SKILL-006: design-import 보안 거부
**Covers**: REQ-SKILL-010

- **Given** bundle ZIP 엔트리에 `.sh` 확장자 실행 파일이 포함된 경우,
- **When** 스킬이 실행되면,
- **Then** 파싱이 거부되고 `DESIGN_IMPORT_SECURITY_REJECT` 오류가 반환된다.

#### AC-SKILL-007: GAN 루프 4차원 스코어링
**Covers**: REQ-SKILL-011, REQ-SKILL-014

- **Given** `.moai/config/sections/design.yaml`에 `pass_threshold: 0.75`, `max_iterations: 5`가 설정되고,
- **When** `moai-workflow-gan-loop`가 실행되면,
- **Then** evaluator-active의 출력은 `design_quality`, `originality`, `completeness`, `functionality` 4개 점수를 포함하고,
- **And** 모든 점수의 산술 평균이 `pass_threshold` 이상이면 루프가 종료된다.

#### AC-SKILL-008: Sprint Contract 필수화
**Covers**: REQ-SKILL-012

- **Given** harness level이 `thorough`로 설정된 상태에서,
- **When** GAN 루프 1회차가 시작되면,
- **Then** Sprint Contract 문서가 `.moai/research/sprints/` 하위에 생성되며,
- **And** Contract 없이 Builder 단계로 진행하지 않는다.

#### AC-SKILL-009: 개선 정체 감지
**Covers**: REQ-SKILL-013

- **Given** 1회차 점수 0.60, 2회차 0.61, 3회차 0.62(모두 개선폭 < 0.05)가 기록된 상태에서,
- **When** 3회차 평가가 완료되면,
- **Then** 시스템은 "stagnation detected" 경고를 출력하고,
- **And** AskUserQuestion으로 에스컬레이션 옵션(계속, 기준 조정, 중단)을 사용자에게 제시한다.

---

### 2.4 기존 `.agency/` 감지 (REQ-DETECT-*)

#### AC-DETECT-001: moai doctor 감지
**Covers**: REQ-DETECT-001

- **Given** `.agency/`가 존재하고 `.moai/project/brand/`가 부재한 프로젝트에서,
- **When** 사용자가 `moai doctor`를 실행하면,
- **Then** 출력에 "agency data detected, migration recommended" 경고가 표시되고,
- **And** `moai migrate agency` 예시 커맨드 라인이 포함된다.

#### AC-DETECT-002: SessionStart 일회성 공지
**Covers**: REQ-DETECT-002

- **Given** `.agency/`가 존재하고 `.moai/project/brand/`가 부재한 프로젝트에서,
- **When** Claude Code 세션이 시작되면,
- **Then** SessionStart 훅이 1회 공지를 출력하고,
- **And** 동일 세션 내 재호출되어도 공지는 중복 출력되지 않는다.

#### AC-DETECT-003: `/moai design` 실행 전 마이그레이션 경고 (NEW, REQ-DETECT-003 명시)
**Covers**: REQ-DETECT-003

- **Given** `.agency/` 디렉터리가 존재하고 `.moai/project/brand/`가 부재한 프로젝트에서,
- **When** 사용자가 `/moai design "<brief>"`를 호출하면,
- **Then** 경로 선택 AskUserQuestion 전에 "agency data detected, migration recommended" 경고가 출력되고,
- **And** 경고 메시지에 `moai migrate agency` 커맨드 예시가 포함된다.

---

### 2.5 `/agency` Deprecation (REQ-DEPRECATE-*)

#### AC-DEPRECATE-001: 경고 출력
**Covers**: REQ-DEPRECATE-001, REQ-DEPRECATE-002

- **Given** `/agency brief`가 호출될 때,
- **When** 래퍼가 실행되면,
- **Then** stderr에 정확히 1개의 경고 라인("`/agency` is deprecated, use `/moai design` instead")이 출력되고,
- **And** `/moai design`의 brief 대응 동작이 수행된다.

#### AC-DEPRECATE-002: 미지원 서브커맨드 오류
**Covers**: REQ-DEPRECATE-002

- **When** 사용자가 `/agency evolve`를 호출하면,
- **Then** 래퍼는 `AGENCY_SUBCOMMAND_UNSUPPORTED` 오류를 반환하고,
- **And** 마이그레이션 가이드 URL이 오류 메시지에 포함된다.

#### AC-DEPRECATE-003: CI CHANGELOG 검증
**Covers**: REQ-DEPRECATE-004

- **Given** 흡수 릴리스 PR에 CHANGELOG.md 업데이트가 없는 상태에서,
- **When** CI가 실행되면,
- **Then** deprecation 검증 단계가 실패하고 머지가 차단된다.

#### AC-DEPRECATE-004: 라이프사이클 명시 검증 (NEW, N3)
**Covers**: REQ-DEPRECATE-003

- **Given** 흡수 릴리스(vN)의 CHANGELOG.md가 머지된 상태에서,
- **When** `grep "REQ-DEPRECATE-003\\|2 minor versions\\|완전 제거"` 결과를 검사하면,
- **Then** CHANGELOG.md에 (1) `/agency` 래퍼 라이프사이클 단계(경고→리다이렉트→제거) 명시, (2) 완전 제거 시점(vN+2) 버전 태그 명시, (3) REQ-DEPRECATE-003 참조 3가지가 모두 기록되어 있어야 한다.

---

### 2.6 Constitution 보존 (REQ-CONST-*)

#### AC-CONST-001: Constitution 재배치 및 무결성
**Covers**: REQ-CONST-001, REQ-CONST-002

- **Given** 흡수 릴리스 후 환경에서,
- **When** `.claude/rules/moai/design/constitution.md`를 열람하면,
- **Then** FROZEN zone 정의(Section 2)가 원본과 일치하며,
- **And** EVOLVABLE zone 정의(Section 2)가 원본과 일치하고,
- **And** 파일 최상단에 재배치 HISTORY 섹션이 명시된다.

#### AC-CONST-002: FROZEN 수정 차단
**Covers**: REQ-CONST-003

- **Given** PR에서 FROZEN zone의 Section 5 Safety Architecture(특히 Layer 5 Human Oversight 포함)의 텍스트가 변경된 상태에서,
- **When** CI가 실행되면,
- **Then** constitution 무결성 체크 단계가 실패하고 머지가 차단되며,
- **And** CI 로그에 수정된 라인 번호와 "수동 구성법 승인 프로세스(Layer 5 Human Oversight) 필요" 메시지가 출력된다.

---

### 2.7 Fallback 경로 (REQ-FALLBACK-*)

#### AC-FALLBACK-001: Figma 보조 모드 비활성 상태 (NEW, REQ-FALLBACK-003 DoD 포함)
**Covers**: REQ-FALLBACK-003

- **Given** `.moai/config/sections/design.yaml`에 `figma.enabled: false`가 기본값으로 설정된 환경에서,
- **When** `moai-domain-brand-design` 스킬이 로드되면,
- **Then** Figma 보조 모드는 비활성화 상태로 유지되며,
- **And** 스킬 로그에 `figma_mode: disabled` 상태가 기록되고,
- **And** Figma URL 입력 시 `FIGMA_MODE_DISABLED` 안내 메시지가 출력되고 코드 폴백 토큰 추출이 계속된다.

#### AC-FALLBACK-002: Figma 보조 모드 활성 상태 (Phase 2 향후 릴리스 기준)
**Covers**: REQ-FALLBACK-003

- **Given** `.moai/config/sections/design.yaml`에 `figma.enabled: true`로 명시되고,
- **And** 유효한 공개 Figma 파일 URL이 제공된 환경에서,
- **When** `moai-domain-brand-design` 스킬이 실행되면,
- **Then** 스킬은 Figma API 또는 공개 파일 파서를 통해 디자인 토큰을 추출하고,
- **And** 추출된 토큰은 `.moai/design/tokens.json`의 `source: figma` 섹션에 기록된다.

> 참고: 현 릴리스의 DoD에는 `figma.enabled: false` 기본값 검증(AC-FALLBACK-001)만 포함되며, AC-FALLBACK-002는 Phase 2 릴리스 시점에 DoD에 포함된다(plan.md §7 배포 계획 참조).

---

### 2.8 Agency 에이전트 제거 (REQ-REMOVE-*)

#### AC-REMOVE-001: 제거된 파일 확인
**Covers**: REQ-REMOVE-001, REQ-REMOVE-002

- **Given** M5 릴리스 후 환경에서,
- **When** `.claude/agents/agency/` 디렉터리를 조회하면,
- **Then** `planner.md`, `designer.md`, `builder.md`, `evaluator.md`, `copywriter.md`, `learner.md` 중 어떤 파일도 존재하지 않는다.

#### AC-REMOVE-002: 참조 감사
**Covers**: REQ-REMOVE-003

- **Given** M5 릴리스 후 전체 저장소에서,
- **When** `grep -r "agency/planner\|agency/designer\|agency/builder\|agency/evaluator"` 감사 스크립트를 실행하면,
- **Then** 감사 결과는 0건이며 CI가 통과한다.

#### AC-REMOVE-003: Skill 병합
**Covers**: REQ-REMOVE-004

- **Given** M5 릴리스 후 환경에서,
- **When** `.claude/skills/agency-frontend-patterns/` 존재 여부를 확인하면,
- **Then** 디렉터리는 존재하지 않으며,
- **And** `moai-domain-frontend/SKILL.md`에 기존 agency-frontend-patterns 내용이 병합되어 있고,
- **And** `.agency/fork-manifest.yaml`(아카이브)에서 해당 엔트리는 유지되지만 활성 참조는 없다.

---

## 3. 엣지 케이스 (Edge Cases)

### EC-001: 빈 `.agency/` 디렉터리
- **Given** `.agency/`가 존재하지만 내부가 비어 있으며,
- **When** `moai migrate agency`를 실행하면,
- **Then** 커맨드는 종료 코드 0으로 종료되고, 경고 "no data to migrate"를 출력하며, `.agency.archived/`는 생성된다(일관성).

### EC-002: 심볼릭 링크 포함
- **Given** `.agency/context/brand-voice.md`가 외부 경로를 가리키는 심볼릭 링크일 때,
- **When** 마이그레이션이 실행되면,
- **Then** 커맨드는 링크를 따라가지 않고, 경고를 출력하며, 해당 파일은 건너뛰고 나머지 작업을 계속 진행한다.

### EC-003: 매우 큰 `.agency/learnings/` (1000개 이상)
- **Given** `.agency/learnings/`에 1000개 이상의 `.md` 파일이 있을 때,
- **When** 마이그레이션이 실행되면,
- **Then** 모든 파일이 `.moai/research/observations/`로 이전되고, 진행률은 100단위로 업데이트된다.

### EC-004: Unicode 파일명
- **Given** `.agency/context/`에 한글 파일명(예: `브랜드-보이스.md`)이 존재할 때,
- **When** 마이그레이션이 실행되면,
- **Then** 파일명이 정확히 보존되어 이전된다 (NFC 정규화 적용).

### EC-005: `/moai design` 브랜드 컨텍스트 부분 완성
- **Given** `.moai/project/brand/brand-voice.md`만 존재하고 나머지 2개는 부재할 때,
- **When** `/moai design`을 호출하면,
- **Then** 시스템은 경고 "incomplete brand context"를 출력하고, 부재한 파일에 대해서만 인터뷰를 유도한다.

---

## 4. 품질 게이트 (Quality Gates)

### 4.1 자동 게이트
- [ ] `go test ./...` 통과 (전체)
- [ ] `go test -race ./...` 통과
- [ ] `go test -cover ./internal/cli/` 커버리지 90% 이상
- [ ] `go test -tags=integration ./internal/cli/` 플랫폼별 권한/시그널 테스트 통과
- [ ] `golangci-lint run` 경고 0건
- [ ] `internal/template/commands_audit_test.go` 통과
- [ ] `scripts/audit-no-agency-refs.sh` 결과 0건 (M5 이후)
- [ ] `make build` 성공, embedded.go 갱신 확인
- [ ] CI constitution 무결성 체크 통과
- [ ] `.moai/config/sections/design.yaml`의 `figma.enabled` 기본값이 `false`로 유지됨 (AC-FALLBACK-001)
- [ ] `.moai/config/sections/design.yaml`의 `supported_bundle_versions` 키가 존재함 (AC-SKILL-011)

### 4.2 수동 QA 게이트
- [ ] `.agency/` 샘플 픽스처로 dry-run → 실제 → 롤백 시연 성공
- [ ] `.agency/` 마이그레이션 중 Ctrl+C 전송 후 `--resume`으로 재개 검증 (AC-MIGRATE-010)
- [ ] `/moai design` 경로 A/B 실사용 테스트 완료
- [ ] `/agency` deprecation 경고 확인 (stderr 1회)
- [ ] Pro 이하 사용자 시뮬레이션에서 경로 B 기본 선택 확인 (REQ-ROUTE-006 override)
- [ ] macOS, Linux (REQ-MIGRATE-012a), Windows (REQ-MIGRATE-012b) 3개 플랫폼에서 마이그레이션 검증
- [ ] 지원되지 않는 bundle 버전 수동 전달 시 `DESIGN_IMPORT_UNSUPPORTED_VERSION` 확인

---

## 5. Definition of Done (DoD)

본 SPEC은 다음을 모두 만족할 때 완료된 것으로 본다:

1. spec.md의 62개 REQ-* 요구사항이 본 문서의 명시적 AC-* 시나리오로 1:1 이상 커버되어 있다 ("(암시)" 매핑 0건)
2. M1 ~ M5 모든 마일스톤의 작업 항목이 완료되었다
3. §4.1 자동 품질 게이트가 모두 통과한다
4. §4.2 수동 QA 게이트가 모두 수행되었다
5. CHANGELOG.md에 주요 변경사항, deprecation 경고, 마이그레이션 가이드 링크가 명시되어 있다 (REQ-DEPRECATE-003 릴리스 계획 포함)
6. 사용자 문서(`docs-site/`)의 4개국어(ko/en/ja/zh)에 마이그레이션 가이드가 작성되어 있다
7. 릴리스 vN의 PR이 승인되어 main 브랜치에 머지되었다
8. `/agency` 완전 제거 시점(vN+2)의 릴리스 일정이 CHANGELOG에 명시되어 있다
9. `.moai/config/sections/design.yaml`의 `figma.enabled: false` 기본값이 현 릴리스에 유지됨 (REQ-FALLBACK-003 현 릴리스 DoD)
10. AC-FALLBACK-002는 Phase 2 릴리스(vN+M)에서 DoD에 포함됨이 CHANGELOG에 명시됨

---

## 6. 추적성 매트릭스 (Traceability Matrix)

총 REQ 개수: 62개 (spec.md §10 참조). 모든 REQ는 최소 1개의 명시적 AC에 매핑된다. "(암시)" 매핑은 0건이다.

| REQ ID | Acceptance Criterion |
|---|---|
| REQ-ROUTE-001 | AC-ROUTE-001, AC-ROUTE-006 |
| REQ-ROUTE-002 | AC-ROUTE-001 |
| REQ-ROUTE-003 | AC-ROUTE-001 |
| REQ-ROUTE-004 | AC-ROUTE-002 |
| REQ-ROUTE-005 | AC-ROUTE-003 |
| REQ-ROUTE-006 | AC-ROUTE-004 |
| REQ-ROUTE-007 | AC-ROUTE-005 |
| REQ-ROUTE-008 | AC-ROUTE-003, AC-SKILL-007 |
| REQ-SKILL-001 | AC-SKILL-001 |
| REQ-SKILL-002 | AC-SKILL-001 |
| REQ-SKILL-003 | AC-SKILL-002 |
| REQ-SKILL-004 | AC-SKILL-010 |
| REQ-SKILL-005 | AC-SKILL-003 |
| REQ-SKILL-006 | AC-SKILL-003 |
| REQ-SKILL-007 | AC-SKILL-004 |
| REQ-SKILL-008 | AC-SKILL-004 |
| REQ-SKILL-009 | AC-SKILL-005 |
| REQ-SKILL-010 | AC-SKILL-006 |
| REQ-SKILL-011 | AC-SKILL-007 |
| REQ-SKILL-012 | AC-SKILL-008 |
| REQ-SKILL-012a | AC-SKILL-008 (harness level standard 분기 검증 포함) |
| REQ-SKILL-013 | AC-SKILL-009 |
| REQ-SKILL-014 | AC-SKILL-007 |
| REQ-SKILL-015 | AC-SKILL-011 |
| REQ-MIGRATE-001 | AC-MIGRATE-001 |
| REQ-MIGRATE-002 | AC-MIGRATE-003 |
| REQ-MIGRATE-003 | AC-MIGRATE-004 |
| REQ-MIGRATE-004 | AC-MIGRATE-001 |
| REQ-MIGRATE-005 | AC-MIGRATE-001 |
| REQ-MIGRATE-006 | AC-MIGRATE-007 |
| REQ-MIGRATE-007 | AC-MIGRATE-001 |
| REQ-MIGRATE-008 | AC-MIGRATE-002 |
| REQ-MIGRATE-009 | AC-MIGRATE-005 |
| REQ-MIGRATE-010 | AC-MIGRATE-006 |
| REQ-MIGRATE-011 | AC-MIGRATE-008 |
| REQ-MIGRATE-012a | AC-MIGRATE-011a |
| REQ-MIGRATE-012b | AC-MIGRATE-011b |
| REQ-MIGRATE-013 | AC-MIGRATE-010 |
| REQ-DIR-001 | AC-MIGRATE-001 |
| REQ-DIR-002 | AC-MIGRATE-009 |
| REQ-DIR-003 | AC-MIGRATE-012 |
| REQ-DETECT-001 | AC-DETECT-001 |
| REQ-DETECT-002 | AC-DETECT-002 |
| REQ-DETECT-003 | AC-DETECT-003 |
| REQ-DEPRECATE-001 | AC-DEPRECATE-001 |
| REQ-DEPRECATE-002 | AC-DEPRECATE-001, AC-DEPRECATE-002 |
| REQ-DEPRECATE-003 | AC-DEPRECATE-004 |
| REQ-DEPRECATE-004 | AC-DEPRECATE-003 |
| REQ-CONST-001 | AC-CONST-001 |
| REQ-CONST-002 | AC-CONST-001 |
| REQ-CONST-003 | AC-CONST-002 |
| REQ-CONST-004 | AC-SKILL-007, AC-SKILL-008 |
| REQ-BRIEF-001 | AC-ROUTE-007 |
| REQ-BRIEF-002 | AC-ROUTE-008 |
| REQ-BRIEF-003 | AC-ROUTE-006 |
| REQ-FALLBACK-001 | AC-ROUTE-004 |
| REQ-FALLBACK-002 | AC-ROUTE-003 |
| REQ-FALLBACK-003 | AC-FALLBACK-001 (현 릴리스 DoD 포함), AC-FALLBACK-002 (Phase 2 릴리스 DoD) |
| REQ-REMOVE-001 | AC-REMOVE-001 |
| REQ-REMOVE-002 | AC-REMOVE-001 |
| REQ-REMOVE-003 | AC-REMOVE-002 |
| REQ-REMOVE-004 | AC-REMOVE-003 |

---

End of ACCEPTANCE.
