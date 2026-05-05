# Proposal — IDEA-001: MoAI Cockpit

> **Phase 6 산출물** · 작성일: 2026-05-04
> 입력: ideation.md (Lean Canvas + Phase 5 Evaluation Report) · research.md (외부 검증 + 인용 출처)
> [HARD] REQ-BRAIN-011: 본 문서 Solution 섹션은 **capabilities**로 기술하며, 구체 구현 기술명은 "Technical Stack Note"에서만 참조 정보로 노출 (SPEC 작성 시 implementation choice로 결정).

---

## Product Statement

MoAI Cockpit은 MoAI-ADK 솔로 개발자를 위한 **localhost-only · read-only · single-page 워크플로우 가시화 대시보드**다. SPEC 진척, worktree 상태, PR/CI, AI agent 메모리, drift 신호를 자동 폴링으로 한 화면에 surface하여 — pull 기반 CLI 왕복을 ambient awareness로 대체한다.

---

## Goal

| 지표 | 목표 |
|------|------|
| **Primary** | 사용자(솔로 개발자)의 1시간당 CLI 왕복 횟수 30%+ 감소 (대시보드 도입 전 1주 vs 후 1주, shell history grep 측정) |
| Secondary 1 | Page load p95 < 200ms (UX 가드) |
| Secondary 2 | 폴링 1회당 데이터 페치 시간 < 300ms (성능 가드) |
| Secondary 3 | `gh` API rate limit hit 횟수 = 0 (안정성 가드) |
| Anti | write 액션 클릭 횟수 = 0 (read-only invariant) |

---

## Audience

- **Primary**: MoAI-ADK 솔로 개발자 (Claude Code + 터미널 + 브라우저 동시 사용)
- **Secondary**: moai-adk-go 메인테이너 (dogfooding)
- **Out-of-scope**: 다중 사용자 팀 환경, 원격 배포, 외부 인증 사용자

---

## Brand Alignment

`.moai/project/brand/` 자산을 1차 시민으로 사용:
- `brand-voice.md`: 패널 타이틀 톤 결정 (간결·낙관·기술적 정확성)
- `visual-identity.md`: 컬러 팔레트·타이포 토큰 적용
- `target-audience.md`: 페르소나 일치성 검증 (Phase 1 Discovery 결과와 정합)

브랜드 통합 수준은 **SPEC-V3R3-WEB-DESIGN-001** (옵션) 에서 결정.

---

## Solution Capabilities (NO tech-stack assumptions)

### Capability C1 — Workflow Tracker (★ MVP 핵심)
현재 진행 중인 SPEC ID, plan/run/sync Phase 위치, progress.md 마지막 체크포인트를 1개 카드로 surface. 5초 주기 자동 갱신.

### Capability C2 — Worktree Switchboard
활성 git worktree 목록, 각 worktree의 branch · 마지막 commit · dirty file count를 표 형태로 surface. 5초 주기.

### Capability C3 — CI/PR Glance Wall
열린 PR 목록, 각 PR의 CI 상태 (pass/pending/fail), 리뷰 상태, MERGED 마커를 badge UI로 surface. 30초 주기 (rate limit 보호).

### Capability C4 — Memory Surfboard
auto-memory 인덱스(MEMORY.md 헤드 + 최근 5 lessons + 최근 5 feedback)를 카드 리스트로 surface. 30초 주기.

### Capability C5 — Drift Sentinel
미커밋 변경, 14일+ stale worktree, merged 후 미삭제 branch를 gentle 알림 패널로 surface (액션 버튼 없음, 알림만). 30초 주기.

### Capability C6 (옵션) — Quick Action Launcher
자주 쓰는 명령 텍스트 + 클립보드 복사 버튼. **실행은 사용자가 터미널에서 paste** (read-only invariant 유지).

### Capability C7 (옵션) — User Configuration
폴링 주기 사용자 설정, 패널 ON/OFF, 서버 포트 변경. 설정 파일 기반 (CLI flag 우선).

### Capability C8 (옵션) — Brand Tokenization
브랜드 자산(visual-identity.md)의 컬러·타이포 토큰을 컴포넌트 스타일에 자동 주입. Solo 사용자도 브랜드 일관성 유지.

---

## SPEC Decomposition Candidates

> SPEC ID prefix는 사용자 후속 명령(`/moai plan SPEC-V3R3-WEB-001`)에 정렬된 **V3R3-WEB** 시리즈를 사용.
> Format: `- SPEC-{DOMAIN}-{NUM}: {scope}` (REQ-BRAIN-011 capabilities-only)

### Recommended MVP Sequence (1→5 sequential, 6-8 optional)

- SPEC-V3R3-WEB-001: Cockpit Foundation + Workflow Tracker capability — CLI entry point, single-page HTTP server skeleton, Workflow Tracker panel rendering current SPEC ID/Phase/progress checkpoint with 5s auto-refresh (Capability C1; First Principles 1순위)
- SPEC-V3R3-WEB-002: Worktree Switchboard panel — multi-worktree state aggregation with branch, last commit, dirty count per row, 5s auto-refresh (Capability C2)
- SPEC-V3R3-WEB-003: CI/PR Glance Wall panel — PR list with CI/review status badges, MERGED marker, 30s auto-refresh + in-memory cache for rate limit protection (Capability C3)
- SPEC-V3R3-WEB-004: Memory Surfboard panel — auto-memory directory traversal, MEMORY.md head + recent lessons/feedback summary cards, 30s auto-refresh (Capability C4)
- SPEC-V3R3-WEB-005: Drift Sentinel panel — uncommitted/stale-worktree/orphan-branch detection, gentle notification list with no action buttons, 30s auto-refresh (Capability C5)

### Optional Extensions (post-MVP, independent)

- SPEC-V3R3-WEB-006: Quick Action Launcher — frequently-used command cards with clipboard-copy buttons, read-only invariant preserved (Capability C6)
- SPEC-V3R3-WEB-007: User configuration system — polling intervals, panel toggle, server port (Capability C7)
- SPEC-V3R3-WEB-008: Brand tokenization integration — automatic application of `.moai/project/brand/visual-identity.md` tokens to component styles (Capability C8)

### Validation Helper (별도 SPEC 또는 README)

- SPEC-V3R3-WEB-BASELINE-001 (옵션): Baseline measurement helper — shell history analyzer to compute pre/post CLI roundtrip counts for verifying the 30%+ reduction success metric (Phase 5 Critical Eval Challenge 1 mitigation)

---

## SPEC Sequencing Rationale

| 순번 | SPEC | 의존성 | 가치 발현 시점 |
|------|------|--------|---------------|
| 1 | WEB-001 | (없음) | Foundation + Workflow Tracker 단독으로도 30%+ 가치 (First Principles 검증) |
| 2 | WEB-002 | WEB-001 (server skeleton) | Worktree 가시화 → 다중 작업 mental load 감소 |
| 3 | WEB-003 | WEB-001, gh CLI | PR 컨텍스트 통합 → CI 결과 발견 시간 단축 |
| 4 | WEB-004 | WEB-001 | Agent 메모리 회고 → AI 협업 추적 |
| 5 | WEB-005 | WEB-002 | Drift 신호 → 누락 방지 (passive guard) |
| 6-8 | WEB-006/007/008 | WEB-001 only | 독립 추가 가능, 우선순위는 사용자 선택 |

**Wave 분할 권장** (lessons #9 wave-split 적용):
- Wave 1: WEB-001 (foundation 확립)
- Wave 2: WEB-002 + WEB-003 (병렬, 둘 다 read-only 데이터 페치 패턴)
- Wave 3: WEB-004 + WEB-005 (병렬)
- Wave 4 (옵션): WEB-006 / WEB-007 / WEB-008 사용자 선택 우선순위로 1-2개

---

## Risks & Mitigations (Phase 5 Eval 반영)

| Risk | 영향 | 완화 |
|------|------|------|
| R1: read-only 한계 만족도 부족 | 사용자가 도구 폐기 | MVP 후 2주 dogfooding go/no-go 게이트 명시 (release 전제) |
| R2: 폴링 누적 IO 부담 | 시스템 응답성 저하 | in-memory cache (TTL = polling interval / 2) + 사용자 폴링 주기 설정 (WEB-007) |
| R3: gh CLI 출력 포맷 변경 | 파서 유지보수 부담 | 모든 gh 호출은 `--json` 플래그로 안정 스키마 사용 (graphql 기반) |
| R4: 도메인 niche가 false negative일 가능성 | 후속 경쟁자 진입 | 본 도구의 USP는 MoAI-ADK 도메인 모델 정렬 → 인접 경쟁자 등장 시에도 차별화 유지 |
| R5: Solo dev 1-2주 추정의 부정확성 | release 지연 | panel-incremental SPEC 분해 (1-2 panel/SPEC) → 부분 출시 가능 |

---

## Definition of MVP Done

다음 5개 조건을 모두 충족하면 MVP 출시 가능:
1. SPEC-V3R3-WEB-001 ~ WEB-005 모두 sync phase 완료 (PR merged)
2. `moai cockpit` 명령으로 localhost 서버 시작 → 브라우저 자동 오픈 → 5개 panel 모두 렌더링
3. 각 panel의 폴링이 정상 동작하며 < 300ms 페치 시간 충족
4. `gh` API rate limit 1시간 사용 카운트 = 0 hit
5. 사용자 본인의 1주일 dogfooding 완료 + go decision

---

## Downstream Consumer Hints

본 proposal.md는 다음 워크플로우의 입력으로 사용된다:

- `/moai project --from-brain IDEA-001` → product.md 작성 시 본 문서를 1차 입력으로 사용 (REQ-BRAIN-007)
- `/moai plan SPEC-V3R3-WEB-001` → "SPEC Decomposition Candidates" 섹션이 자동 surface되어 첫 SPEC 후보 채택 (Phase A8.2 downstream patch)
- `/moai design --path A` (옵션) → claude-design-handoff/ 디렉토리가 입력 (Phase 7 산출물)

---

## Technical Stack Note (참고용 — SPEC 작성 시 결정)

> 이 섹션은 capabilities 기술과 분리된 **참고 정보**다. 사용자 입력 idea가 "Go templ + HTMX, localhost"를 명시했으며, research.md §2가 이 조합을 검증했다. 단, 각 SPEC은 implementation choice를 자체 결정한다.

검증된 후보 스택:
- Backend: Go (chi 또는 net/http) + `a-h/templ` (Apache 2.0)
- Frontend: HTMX (BSD) + 선택적 Tailwind 또는 templUI (MIT)
- 폴링: HTMX `hx-trigger="every Ns"` + templ Fragments (`@templ.Fragment("name")`) 부분 갱신
- 데이터 소스: `git`/`gh` CLI shell-out + filesystem walk (외부 DB 없음)
- 시작 명령: `moai cockpit [--port N] [--no-open]` (CLI integration)
- 베이스 레퍼런스: `Piszmog/go-htmx-template` (Go 1.25+ native CSRF), `josephspurrier/gohtmxapp` (`/dashboard` 즉시 실행 가능 패턴)

각 SPEC은 위 후보를 채택하거나 대안 평가 가능. 본 proposal.md는 capabilities를 절대 가치로 두며, 구현 기술은 SPEC plan/run 단계에서 manager-strategy 위임으로 결정.
