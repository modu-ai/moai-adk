# Ideation — IDEA-001: MoAI Cockpit

> **Phase 4 산출물 (Converge → Lean Canvas)** · 작성일: 2026-05-04
> Phase 5 Critical Evaluation 섹션은 본 문서 하단에 추가 append됨 (단일 파일).

---

## Converged Concept

**MoAI Cockpit** — Claude Code 세션 옆 브라우저 탭 하나로 MoAI-ADK SPEC 워크플로우(plan → run → sync) 전체를 한눈에 보는 **localhost-only · read-only · single-page** 대시보드.

Go `templ` 컴포넌트 + HTMX `hx-trigger="every Ns"` 폴링으로 자동 갱신. 5개 핵심 패널 + 1개 보조 패널을 단일 페이지에 배치한다.

### 5+1 Panel Composition

| # | Panel | 데이터 소스 | 폴링 주기 | 해결하는 CLI |
|---|-------|------------|----------|--------------|
| 1 | **Workflow Tracker** | `.moai/specs/SPEC-*/{spec.md,plan.md,progress.md}` | 5s | `ls .moai/specs/`, `cat progress.md` |
| 2 | **CI/PR Glance Wall** | `gh pr list --json` + `gh pr checks` | 30s | `gh pr list`, `gh pr checks N` |
| 3 | **Worktree Switchboard** | `git worktree list --porcelain` + per-worktree `git status -s` count | 5s | `git worktree list`, `git status` |
| 4 | **Memory Surfboard** | `~/.claude/projects/{hash}/memory/MEMORY.md` + 최근 5 lessons | 30s | `cat MEMORY.md`, `ls memory/` |
| 5 | **Drift Sentinel** | uncommitted/stale worktree/orphan branch 알림 (gentle, not actions) | 30s | `git status`, `git branch --merged` |
| + | **Quick Action Launcher** (보조) | 자주 쓰는 명령 클립보드 복사 (read-only, 실행 X) | static | n/a |

### 명시적 비범위 (MVP에서 제외)

- write 액션 (sync trigger, SPEC 편집, 명령 실제 실행)
- 다중 사용자/원격 호스팅
- 인증·세션·CSRF (localhost only · 사용자 자기 자신만 접근)
- 외부 DB (모두 filesystem + git + gh CLI 호출)
- 실시간 로그 stream (SSE/WebSocket — 폴링으로 충분)

---

## Lean Canvas (9 Blocks)

### 1. Problem

| 우선순위 | Problem |
|---------|---------|
| **P1** | MoAI-ADK 솔로 개발자가 SPEC 워크플로우 진행 도중 5종 CLI(`gh pr list`, `gh pr checks`, `git worktree list`, `git status`, `ls .moai/specs/`)를 1시간당 15-28회 반복 호출 → 컨텍스트 전환으로 시간당 평균 23분 15초의 복귀 비용 누적 (Mark, UC Irvine) |
| **P2** | 다중 worktree 운용 시 어느 worktree가 어떤 SPEC을 진행 중인지 mental tracking 부담 — 솔로 개발자가 동시에 인지할 수 있는 한계 초과 시 작업 누락·중복 발생 |
| **P3** | CI fail / PR blocker / stale worktree 같은 이벤트가 명시적 알림 없이 누적 → 발견 지연(mean time to detect) |

### 2. Customer Segments

| 세그먼트 | 설명 | 우선순위 |
|---------|------|---------|
| **Primary** | MoAI-ADK 솔로 개발자 (1인 SPEC 사이클 운영, Claude Code + 터미널 + 브라우저 동시 사용) | ★ MVP target |
| Secondary | MoAI-ADK 자체 개발자 (modu-ai/moai-adk 메인테이너) — dogfooding 수혜 | post-MVP |
| Future | 2-5명 팀 리드/매니저 — 멀티 사용자 진척 가시화 | 별도 idea (out-of-scope) |

### 3. Unique Value Proposition

> **"터미널을 떠나지 않고도, 브라우저 탭 하나만 열어두면 SPEC 워크플로우 전체가 자동으로 보인다."**

핵심 가치:
- **Zero command typing** — 5종 CLI를 자동 폴링이 대체
- **MoAI-ADK 도메인 1차 시민** — gh-dash가 인식 못하는 SPEC ID·Phase·MX·TRUST를 자체 객체로 표현
- **Localhost-only safety** — 외부 의존성·인증·SaaS 없음. solo dev 워크플로우에 정확히 맞춤

### 4. Solution

read-only single-page web app:
- 백엔드: Go (chi 또는 net/http) + `a-h/templ` HTML 컴포넌트
- 프론트: HTMX (`hx-trigger="every Ns"` polling + `hx-swap="outerHTML"`)
- 데이터: 모두 filesystem + `git`/`gh` CLI shell-out (외부 DB 없음)
- 배포: `moai cockpit` CLI 한 줄로 시작 → `localhost:PORT/` 자동 오픈
- 패널 갱신: templ Fragments(`@templ.Fragment("name")`) + HTMX partial swap → 부분 갱신만

핵심 기술 패턴 (research.md §2.2 참조):
```html
<div id="worktree-list" hx-get="/api/worktrees" hx-trigger="every 5s" hx-swap="outerHTML">
  Loading...
</div>
```
서버는 `templ.Handler(Page(state), templ.WithFragments("worktree-list"))` 로 부분 응답.

### 5. Channels

| 채널 | 도달 방법 |
|------|----------|
| **Primary** | `moai cockpit` CLI 명령 (moai-adk-go 자체 binary에 통합) |
| Discovery | README.md + adk.mo.ai.kr 문서 + CHANGELOG release notes |
| Word-of-mouth | dogfooding 후 modu-ai/moai-adk 메인테이너 공유 |

### 6. Revenue Streams

> **OSS · Free · MIT** — 직접 매출 없음.

간접 가치:
- moai-adk-go 채택률·retention 향상 → 생태계 확장
- 사용자 행동 메트릭 수집 (옵션, opt-in) → 향후 SPEC 분해/품질 개선 입력

### 7. Cost Structure

| 항목 | 추정 비용 |
|------|----------|
| 초기 구현 | 솔로 개발 1-2주 (6-10 SPEC × 0.5-1일/SPEC 가정) |
| 의존성 | a-h/templ (Apache 2.0), bigskysoftware/htmx (BSD), Tailwind 또는 templUI (MIT) — 모두 OSS |
| 유지보수 | git/gh CLI 인터페이스 변경 시 파서 업데이트 (낮음, gh CLI는 stable) |
| 사용자 비용 | 0원 (localhost only, 외부 SaaS 없음) |

### 8. Key Metrics

**Primary metric** (사용자 선택):
- **CLI 왕복 횟수 30%+ 감소** (대시보드 도입 전 1주 vs 후 1주, shell history grep 기준)

**Secondary metrics**:
- DAU (대시보드 일일 방문 횟수) — retention 신호
- 평균 세션 길이 (대시보드 탭이 열려있는 시간)
- Page load p95 latency < 200ms (UX 가드)
- 폴링 1회당 데이터 페치 시간 < 300ms (성능 가드)
- `gh` API rate limit hit 횟수 = 0 (안정성 가드)

**Anti-metrics** (이 수치가 올라가면 MVP 철학 위반):
- write 액션 클릭 횟수 (read-only 가드)
- 인증 시도 실패 횟수 (localhost-only 가드)

### 9. Unfair Advantage

| 우위 요소 | 설명 |
|----------|------|
| **MoAI-ADK 도메인 모델 직접 접근** | SPEC ID·Phase·MX tag·TRUST 5는 외부 도구가 모르는 1급 객체. 본 도구만 자연 표현 가능 |
| **Solo + Localhost niche** | SaaS 경쟁자 진입 매력 낮음(시장 작음, lock-in 가치 없음) → 장기적으로 유일한 옵션 가능 |
| **moai-adk-go 자체 통합** | `moai cockpit` 단일 명령으로 즉시 사용 가능 — 별도 install/config 불필요 |
| **Brand 파일 기반 디자인** | `.moai/project/brand/visual-identity.md` 토큰을 직접 활용 → MoAI 브랜드 일관성 |

---

## Critical Decisions Log (Phase 4 Converge 시 거부된 대안)

| 거부된 옵션 | 채택 안 한 이유 |
|-------------|---------------|
| 다중 페이지 4-5 패널 분리 | 사용자 명시 선택: "단일 페이지 read-only" — 페이지 전환 자체가 컨텍스트 전환 비용 |
| WebSocket / SSE 실시간 | 솔로 개발자 + 5초 polling으로 충분 · HTMX 폴링이 코드량 1/10 |
| 외부 SQLite/Postgres | filesystem + git/gh CLI shell-out으로 충분 · 의존성 0 유지 |
| Quick Action 실행 버튼 (write) | 사용자 명시 선택: "read-only" — 권한·confirmation 흐름 비용 회피 |
| 인증/CSRF | localhost only · 자기 자신만 접근 가정 → 불필요 |
| Tailwind 직접 vs templUI | Phase 6 SPEC 분해 시 결정 (둘 다 후보 유지) |

---

# Phase 5: Critical Evaluation Report

> **Phase 5 산출물** · 동일 파일에 append (workflow 명세)
> 평가 도구: moai-foundation-thinking — Critical Evaluation + First Principles

---

## 5.1 Critical Evaluation (Adversarial Review)

### Challenge 1 — "이게 정말 필요한가? `tmux + watch` 조합으로도 되지 않나?"

**반박 시도**: `watch -n 5 'gh pr list && git worktree list && git status -s'` 명령으로 동일 효과 가능.

**대응**:
- `watch`는 텍스트 스트림 — 시각적 그룹화·하이라이트·badge 부재
- 5종 CLI를 한 화면에 출력하면 80+ 라인 스크롤 — 한눈에 안 들어옴
- MoAI-ADK 도메인 객체(SPEC Phase, MX tag count, TRUST 5 진척)는 `watch`로 surface 불가 — 별도 파싱·계산 필요
- 그러나 **이 challenge는 일부 유효** — MVP 출시 전 사용자 1명(본인)이 1주일간 `watch` 변형으로 baseline 측정 권장. dashboard 도입 후 만족도 비교가 30%+ 감소 가설 검증의 정직한 baseline.

→ **수용 (보강 포함)**: research.md §3.4 측정 방법에 "baseline용 `watch` 비교" 섹션 추가 권장.

### Challenge 2 — "5초 polling × 5 panels = 초당 1 request. localhost지만 IO 부담?"

**반박 시도**: `git status` 호출은 큰 monorepo에서 100ms+ 소요 가능 → 누적 시 시스템 응답성 저하.

**대응**:
- 데이터 페치 캐싱 (in-memory map, TTL = polling interval / 2) → 동일 worktree 중복 호출 차단
- `gh` CLI 호출은 30s 폴링으로 분리 (rate limit + 비용 모두 완화)
- worktree count > 5인 경우 사용자 옵션으로 폴링 주기 늘리는 설정 노출 (Phase 6 SPEC 후보)

→ **수용**: ideation.md "Cost Structure"에 "in-memory cache + 폴링 주기 사용자 설정" 명시 보강 필요. Phase 6 SPEC-DASH-CONFIG-001 후보로 등록.

### Challenge 3 — "Solo dev니까 이 도구 만든 사람도 단 1명. 6-10 SPEC × 1-2주는 과대 추정 아닌가? 또는 과소 추정?"

**대응**:
- moai-adk-go의 기존 SPEC 평균 분량(WF-001~005, BRAIN-001 = 평균 0.5-1 days/SPEC, plan+run+sync 사이클) 기준 → 6-10 SPEC × 1-2주는 **현실 범위 내**
- 단, MVP를 5 panels 모두로 정의하면 위험 — **Phase 6에서 단일 SPEC 단위로 1-2 panel씩 incrementally** (예: SPEC-DASH-CORE → SPEC-DASH-WORKFLOW → SPEC-DASH-PR → ...)

→ **수용**: Phase 6 proposal.md SPEC 분해는 panel 단위 incremental 출시를 전제로 설계.

### Challenge 4 — "본 idea의 실패 시나리오는?"

**가장 확률 높은 실패 시나리오**:
1. 사용자(본인)가 1주일 사용 후 "역시 터미널이 더 빠르다"고 판단 → 폐기
2. `gh` CLI / `git worktree` 출력 포맷이 stable한 줄 알았으나 의외로 자주 변경 → 파서 유지보수 부담 누적
3. read-only 한계로 만족도 부족 → write 기능 요구 → "그러면 `moai sync`도 그냥 터미널에서 치자" → 가치 소멸

**완화 전략**:
- 시나리오 1: MVP 출시 후 2주 dogfooding 기간 명시 + go/no-go 결정 게이트
- 시나리오 2: gh CLI 호출은 `--json` 플래그로 안정적 출력 보장 (이미 graphql 스키마 안정)
- 시나리오 3: "Quick Action Launcher"가 명령 텍스트 + 복사 버튼으로 → 사용자가 직접 paste → 이게 충분한지 1차 검증

### Challenge 5 — "Solo + Localhost niche가 정말 빈 자리인가? 검색 결과 없음 ≠ 시장 부재일 수도"

**대응**:
- research.md §1.3 결론은 검색 1회로 도달한 결과 — false negative 가능
- 그러나 본 idea의 USP는 "MoAI-ADK 도메인 모델 + dashboard"이며, **MoAI-ADK 자체가 niche**이므로 동일 도메인 경쟁자 부재는 정합적
- 인접 도메인(GitHub PR + git worktree + 일반 OSS)에는 경쟁자 존재 — 본 도구는 그 영역 차별화에 의존하지 않음

→ **유지** (challenge 무력화 아닌 영역 재정의로 대응).

---

## 5.2 First Principles Decomposition

핵심 질문: **"MoAI Cockpit이 충족시키는 가장 근본적인 사용자 욕구는 무엇인가?"**

### Layer 1 — Surface 욕구
"SPEC 워크플로우 가시화" → 표면적 표현. 하지만 가시화 자체는 수단이지 목적이 아님.

### Layer 2 — Mid-level 욕구
"CLI 왕복 30% 감소" → 측정 가능한 지표. 하지만 왕복 자체가 문제이기보다는 그 결과인 컨텍스트 전환이 문제.

### Layer 3 — Deep 욕구
"솔로 개발자가 동시에 의식해야 하는 mental state 수를 줄여서, **deep work 시간을 늘리고 싶다**." (Mark, Newport: deep work 1시간 = 일반 작업 4시간 가치)

### Layer 4 — Root 욕구
"AI agent와 페어 프로그래밍하는 솔로 개발자가, **agent의 진척과 자기 결정 시점을 분리해서 인지하고 싶다**." MoAI-ADK는 인간 의사결정(plan)과 AI 실행(run/sync)을 분리하는데, 그 사이 "지금 agent가 어디 있나?"를 추적하는 인지 부담이 본질적 문제.

→ **본 도구의 진짜 가치는 "agent와의 페어 프로그래밍에서 인간이 lead 위치를 잃지 않게 하는 것"**.

### 재구성된 Solution Statement (First Principles 결과)

> "MoAI Cockpit은 인간이 AI agent에게 위임한 작업의 진척을 자기 페이스로 surface한다 — pull(터미널 명령) 대신 ambient awareness(브라우저 탭 폴링)로."

이 재구성은 Phase 6 proposal.md의 USP 섹션과 SPEC 분해 우선순위에 영향:
- ★ **Workflow Tracker (panel 1)** = 가장 본질적 → 첫 SPEC으로 분리 (혼자 작동해도 가치 30% 이상 제공)
- CI/PR · Worktree · Memory · Drift = 점진 추가 가능
- Quick Action Launcher = Phase 6 후보에서 격하 가능 (옵션 SPEC)

---

## 5.3 Evaluation Summary

| 평가 차원 | 결과 |
|----------|------|
| 컨셉 명확성 | ★★★★★ (페르소나 + 문제 + 해결 모두 명시) |
| 기술 실현성 | ★★★★★ (templ + HTMX 패턴 검증, 솔로 1-2주 범위) |
| 시장/도메인 niche | ★★★★ (검색 1회 결과 — false negative 가능성 존재) |
| 측정 가능성 | ★★★★ (shell history grep으로 baseline 가능, baseline 측정 보강 필요) |
| 실패 위험 완화 | ★★★ (3개 시나리오 식별, 완화책 부분만 정의 — Phase 6에서 SPEC 단위로 강화 필요) |
| First Principles 정합 | ★★★★★ (인간-AI 페어링의 본질적 문제와 정렬) |

**전체 결론**: Phase 6 SPEC 분해로 진행 가능. 단, 다음 사항을 proposal.md에 반영:

1. SPEC 분해는 **panel-incremental** (5-panel 일괄 X, 1-2 panel/SPEC)
2. **Workflow Tracker가 첫 번째 SPEC** (First Principles 결과 — 본질 가치 1순위)
3. **MVP baseline 측정 SPEC** (선택, Phase 6에서 사용자 결정)
4. **Polling 주기 사용자 설정** SPEC 후보로 등록 (성능 risk R2 완화)
5. read-only invariant 유지 — write 기능은 별도 idea로 분리
