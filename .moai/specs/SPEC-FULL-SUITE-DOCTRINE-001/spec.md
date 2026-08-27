---
id: SPEC-FULL-SUITE-DOCTRINE-001
title: "manager-develop 전량-스위트 지시의 상위 계약 복귀"
version: "0.5.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: ".claude/agents/moai, .claude/rules/moai, internal/template/templates"
lifecycle: spec-anchored
tier: M
tags: "doctrine, verification, manager-develop, template-first, drift"
---

# SPEC-FULL-SUITE-DOCTRINE-001 — manager-develop 전량-스위트 지시의 상위 계약 복귀

## HISTORY

| 날짜 | 버전 | 변경 | 근거 |
|---|---|---|---|
| 2026-08-27 | 0.1.0 | 최초 작성 (plan-phase) | 카드 t301 · 조사 기록 `.moai/reports/t301/discovery.md` |
| 2026-08-27 | 0.2.0 | plan-audit iter-1 (FAIL 0.57) 대응 — 차단 결함 9건 반영 | `.moai/reports/t301/plan-audit.md` |
| 2026-08-27 | 0.3.0 | plan-audit iter-2 (FAIL 0.75) 대응 — 세 번째 배포 사본(N1) 범위 편입 + mutant 구멍 3건 마감 | `.moai/reports/t301/plan-audit-iter2.md` · 운영자 상한 연장 승인 |
| 2026-08-27 | 0.4.0 | plan-audit iter-3 (FAIL 0.81, 임계 통과) 대응 — C3가 생성 산출물임을 반영(N3 단독). 그 외 문면 불변 | `.moai/reports/t301/plan-audit-iter3.md` |
| 2026-08-27 | 0.5.0 | plan-audit final (FAIL 0.83) 대응 — AC-FSD-010 명령 순서 교정(N8) + `plan.md` 잔존 C7 제거(N9). 5줄 델타 | `.moai/reports/t301/plan-audit-final.md` |

---

## §A. 배경

`manager-develop` 에이전트 정의가 run-phase 에이전트에게 **Go 전체 테스트 스위트 실행**을 지시한다. 이 지시는 같은 리포가 이미 채택해 배포까지 하고 있는 상위 계약과 정면으로 어긋난다.

상위 계약 쪽 (둘 다 `internal/template/templates/` 에 미러가 있는 **배포 대상**이다):

- `AGENTS.md:117-118` — CLAUDE.md 0절이 `@AGENTS.md` 로 임포트하는 상시 로드 표준 계약. "변경이 영향 줄 수 있는 테스트만 돌리고, push 후 CI가 전량을 돌리게 한다. 부하 걸린 개발 머신에서의 전량 실행은 코드가 아니라 머신을 잰다."
- `.claude/rules/moai/workflow/kanban-dispatch.md:195` — `[HARD]`, 같은 규율.

즉 "로컬은 변경 범위 · 전량은 CI" 는 **이미 배포 계약 층에서, 지역 사정 서술 없이 일반 근거로 채택돼 있다.** 따라서 이 SPEC이 판정할 것은 "어느 쪽이 정본인가" 가 아니다. `manager-develop` 정의 쪽이 그 계약에서 갈라져 나온 **drift** 이고, 이 SPEC의 일은 drift를 되돌리는 것 + 사라지는 전량 판정을 무엇이 대신하는지 **명시**하는 것이다.

### A.1 측정된 위반 지점 — 사본 3개 × 지점 4개 = 12곳

`manager-develop` 정의는 **세 벌** 존재하며 셋 다 같은 네 지점을 문자 그대로 담고 있다. 측정 트리: 워크트리 `.claude/worktrees/t301`, HEAD `d29b8942e`.

| 사본 | 경로 | 배포 여부 | 로컬 쌍 |
|---|---|---|---|
| **C1** 로컬 md | `.claude/agents/moai/manager-develop.md` | 아니오(작업 사본) | — |
| **C2** 템플릿 md | `internal/template/templates/.claude/agents/moai/manager-develop.md` | 예 | C1 |
| **C3** 템플릿 toml | `internal/template/templates/.codex/agents/moai/manager-develop.toml` | 예 | **없음** (C2에서 생성) |

지점별 줄번호:

| # | C1 | C2 | C3 | 문면 |
|---|---|---|---|---|
| S1 | :92 | :93 | :80 | `Step 5 always runs the full suite regardless of scale.` |
| S2 | :126 | :127 | :114 | `run tests — targeted when ddd LARGE_SCALE, otherwise the full suite` |
| S3 | :132 | :133 | :120 | `Run the COMPLETE test suite (always full, regardless of LARGE_SCALE; ...)` |
| S4 | :135 | :136 | :123 | `Issue the independent read-only verifications (full suite, coverage, lint, boundary greps) ...` |

C1↔C2는 frontmatter `isolation: worktree` 1줄 때문에 줄번호가 **일괄 +1** 이다. C3는 TOML 포맷이라 줄번호 체계가 다르지만 **문면은 동일**하고, 네 패턴 카운트도 C1·C2와 똑같이 `3` 이다(실측).

**C3는 Codex 듀얼 하니스용 미러이며 배포된다.** `internal/template/embed.go` 의 `//go:embed all:templates` 가 `templates/` 하위 전체를 바이너리에 넣으므로 `.codex/` 도 함께 나간다. 형제 10개(`manager-lead.toml`, `manager-docs.toml` 등)와 함께 있는 정식 미러 트리다.

**그리고 C3는 손으로 관리하는 사본이 아니라 생성 산출물이다.** `internal/template/agentemit` 패키지가 중립 `.md` 층(C2)과 매니페스트(`agents-codex.yaml`)를 받아 결정적으로 `.codex/agents/moai/*.toml` 을 만든다. 재생성 진입점은 `make agents-emit` 이고(`AGENTEMIT_UPDATE=1` 로 골든 산출물을 다시 쓴다), 커밋된 C3가 현재 방출 결과와 일치하는지는 골든 테스트 `TestGoldenCommittedArtifactsMatchEmission` 이 지킨다 — 손으로 고친 `.toml` 은 이 테스트에서 `committed artifact differs from emission (sha256 mismatch)` 로 떨어진다.

**`make build` 는 `agents-emit` 을 부르지 않는다.** 두 타깃은 서로 독립이므로(`build:` 의 선행 타깃은 `templ-generate` 뿐이다), C2만 고치고 재생성을 건너뛰면 C3가 낡은 채로 남고 **다음 방출 때 수리가 조용히 되돌아간다.** 세 경로 중 하나만 옳다: C2를 고친 뒤 **재생성**(옳음) · 손으로 편집(골든 가드에 걸리거나 운으로 통과) · 재생성 생략(다음 방출이 되돌림).

**C3에는 로컬 쌍이 없다** — `.codex/agents/moai/manager-develop.toml` 는 워크트리 루트에 존재하지 않는다(실측: No such file or directory). 따라서 C3는 로컬↔템플릿 델타 판정(AC-FSD-008)의 대상이 아니며, 대신 부재·존재형 AC가 C3를 **직접** 측정하고 골든 테스트가 생성 계보를 지킨다(AC-FSD-010).

S4는 카드에도 최초 조사 기록에도 없던 발견이고, C3 사본 자체는 iter-2 감사에서 드러났다. 최초 조사가 `internal/template/templates/.claude/` 아래만 확인해 미러 계보가 하나라고 가정한 결과다.

### A.2 `LARGE_SCALE` 판별자의 운명 — 존치가 아니라 제거

`LARGE_SCALE` 는 세 사본 각각에 **정확히 3회** 등장하며(총 9회, 실측), 그 3회가 S1·S2·S3와 같은 줄이다. 유일한 귀결은 S1이 적은 대로 `switches PRESERVE/IMPROVE to targeted test execution` 이다.

새 독트린에서는 **모든** 실행이 변경 범위 스코프가 된다. 그러면 판별자가 가르는 것이 없어진다. `otherwise` 라는 단어만 지우고 판별자를 남기면 문법적 dangling은 사라져도 실질은 그대로 남는다 — 아무것도 가르지 않는 조건이다. 따라서 이 SPEC은 **판별자 자체를 제거**하는 쪽으로 확정한다(REQ-FSD-003). 이 결정은 dangling 금지를 기계 판정 가능한 형태로 바꾼다는 부수 효과도 갖는다: 토큰이 0이면 매달린 가지가 존재할 수 없다.

토큰 의존성은 없다 — 이 토큰을 읽는 Go 코드·훅·CI 워크플로가 없음이 리포 전역 검색으로 확인됐다. 메모리 가드 서술(`module-level batches when needed`)은 별개 관심사이며 이 SPEC이 존치·삭제를 강제하지 않는다.

### A.3 검증 배치 정의를 함께 고쳐야 하는 이유

S3·S4가 가리키는 배치 정의 자체가 1번 항목으로 전량 실행을 못박고 있다. 배치를 두고 에이전트 정의만 고치면 STEP 5는 배치를 통해 여전히 전량을 돈다 — 수리가 **무효(inert)** 가 된다. C3를 범위에 넣는 근거도 같은 논리다: 배포되는 사본 하나를 두면 그 표면에서 수리가 무효다.

| 파일 | 줄 | 문면 | 템플릿 미러 |
|---|---|---|---|
| `.claude/rules/moai/core/agent-common-protocol-reference.md` | :21-22 | `# 1. Full test suite (Go)` 주석 + 무조건 전량 호출 | 존재 · 로컬과 **byte 동일** |
| `.claude/rules/moai/workflow/verification-batch-pattern.md` | :45 | Group A 행 — 무조건 전량 호출 + `30-120 s` 추정치 | 존재 · 로컬과 **byte 동일** |

**이 두 규칙 파일에는 codex 미러가 없다.** `.codex/` 트리는 `agents/moai/*.toml` 11개만 담고 있고 `rules` 디렉터리 자체가 없다(실측: 전체 트리 나열로 확인). 따라서 C3 축은 에이전트 정의 한 갈래에 국한된다. C3 본문도 배치 정의를 인라인으로 품고 있지 않다.

**범위 한정 — `agent-common-protocol-reference.md` 는 배치 1번 항목만 대상이다.** 이 파일에는 무조건 전량 호출이 3회 등장하지만, 나머지 2회(`:65`, `:76`)는 처방이 아니라 **예시 산문**이다: `:65` 는 "직렬 검증 안티패턴" 블록에서 하지 말라는 형태를 보여주고, `:76` 은 "언제 직렬로 실행하는가" 의 의존 관계 예시다. 둘 다 전량 실행을 지시하지 않으므로 범위 밖이며, 이 SPEC의 어떤 변경으로도 뒤집히지 않는다. 판정은 배치 코드블록의 `# 1.` ~ `# 2.` 구간으로 좁힌다(AC-FSD-006).

### A.4 사라지는 전량 판정을 무엇이 대신하는가

run-phase STEP 5 시점에는 카드 브랜치가 **아직 push 되지 않았다.** 따라서 그 순간 CI 판정은 존재하지 않는다. 이 리포에서 통합 판정의 주체는 push 이후 통합 브랜치에서 도는 CI다(지역 사실로는 `origin/develop`, `.claude/rules/local/gitflow-lane-protocol.md` §4·§8).

그러므로 완료 보고에서 "전량 통과" 주장을 **조용히 삭제만 하면** 보고가 근거 없이 통과로 읽힌다. 위임 사실과 미결 상태를 명시하는 문구가 반드시 함께 들어가야 한다 — `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 2(manager-agent 완료 보고)의 관측되지 않은 완료 주장에 해당하기 때문이다.

**단, 그 문구는 브랜치를 명명하지 않는다.** C2·C3는 배포 템플릿이고, `origin/develop` 은 현재 배포 템플릿 전체에 **0회** 등장한다(이 트리 실측: `grep -rl 'origin/develop' internal/template/templates/` → 매치 없음, `rc=1`). `develop` 통합 브랜치 모델은 미러 없는 로컬 전용 규칙 소관이라, 하류 사용자에게는 존재하지 않는 브랜치다. 상위 계약이 이미 쓰는 방식 — 브랜치 없이 CI만 명명 — 을 따른다. `origin/develop` 이라는 지역 사실은 배포되지 않는 이 문서(§A.4)에만 남는다.

---

## §B. 요구사항 (GEARS)

> **범위 규약**: 이하 요구사항에서 `the manager-develop agent definition` 은 §A.1이 열거한 **세 사본(C1·C2·C3) 전부**를 가리킨다. 한 사본만 만족시키는 수리는 어느 요구사항도 충족하지 않는다.

### B.1 drift 되돌리기

- **REQ-FSD-001** (Ubiquitous) — The `manager-develop` agent definition shall scope run-phase test execution to the packages the change can affect, and shall carry the canonical phrase `the tests the change can affect` inside its STEP 4 change-loop block so the obligation is machine-detectable at the governing site rather than merely somewhere in the file.
- **REQ-FSD-002** (Unwanted) — The `manager-develop` agent definition shall not instruct unconditional full-suite test execution at any of the four measured sites (S1, S2, S3, S4) in any of the three copies, and shall not restate the instruction through a synonym of the whole-suite word family.
- **REQ-FSD-003** (Unwanted) — The repaired `manager-develop` agent definition shall not retain the `LARGE_SCALE` discriminant, whose single stated consequence becomes unconditional under REQ-FSD-001, because a discriminant that discriminates nothing is the dangling conditional this requirement exists to forbid.

### B.2 배치 정의

- **REQ-FSD-004** (Ubiquitous) — The canonical verification batch's first item shall prescribe a package-scoped test invocation carrying the `internal/<pkg>` placeholder, in place of the unconditional whole-repository invocation.
- **REQ-FSD-005** (Ubiquitous) — The verification-batch grouping heuristic's Group A row shall name a change-scoped invocation rather than the unconditional one, and its time-estimate cell shall not assert a duration range this SPEC did not measure.

### B.3 위임·미결 명시

- **REQ-FSD-006** (Event-driven) — When the run-phase agent generates its STEP 5 completion report, the agent definition shall require the report to name the CI run on the project's `integration branch` as the owner of the full-suite verdict **and**, in the same sentence, to state that this verdict is `PENDING at report time`; carrying only one of the two is not compliance.
- **REQ-FSD-007** (Unwanted) — The agent definition shall not name a specific branch as the verdict owner, because the file is distributed to projects whose branch model differs from this repository's.
- **REQ-FSD-008** (Unwanted) — The completion-report instruction shall not permit the full-suite claim to be silently omitted, and shall not permit the pending state to be dropped from the delegation sentence, because an omitted verdict and an unqualified delegation both read as a passed one.

### B.4 배포 중립성 · Template-First

- **REQ-FSD-009** (Capability gate) — Where a repaired file lives under `internal/template/templates/` — including the `.codex/` harness mirror, not only `.claude/` — the repair shall land in that distributed copy. Where such a copy additionally has a local twin, the two shall differ only by their known pre-existing frontmatter delta (`isolation: worktree` for `manager-develop.md`; zero delta for the two rules files); where it has no local twin, the distributed copy shall be measured directly instead.
- **REQ-FSD-010** (Unwanted) — The replacement text in distributed files shall not carry repo-local content: no incident narrative, no calendar date, no SPEC ID, no `CLAUDE.local.md` reference, no machine-specific path, and no repository-specific branch name.

### B.5 회귀 관측

- **REQ-FSD-011** (Event-driven) — When one real run-phase STEP 5 batch executes after the repair lands, the run-phase actor shall record that batch's duration and the command that produced it as the post-change observation.

---

## §C. 제약

| # | 제약 |
|---|---|
| C1 | 이 SPEC의 산출물은 **markdown · toml 문면 편집뿐**이다. Go 코드 변경 없음. |
| C2 | Template-First 순서 강제 — `internal/template/templates/` 먼저 편집 → `make build` → 로컬 반영 (`CLAUDE.local.md` §2·§2.3). 로컬만 고치면 다음 `moai update` 가 템플릿판으로 되돌린다. |
| C3 | 로컬 전용 3개 지점(`CLAUDE.local.md:246`, `:382`, `.claude/rules/local/gitflow-lane-protocol.md:92`)은 **미러가 없다.** 미러하지 않는다. |
| C4 | 구현 중 전체 스위트를 로컬에서 실행하지 않는다 — 이 SPEC이 금지 대상으로 다루는 바로 그 행위다. |
| C5 | 인용하는 모든 좌표는 편집 직전 해당 트리에서 재측정한다. 줄번호는 편집으로 이동한다. |
| C6 | REQ-FSD-001·004·006이 지정한 정본 문자열(`the tests the change can affect`, `internal/<pkg>`, `integration branch`, `PENDING at report time`)은 AC의 계측점이다. 동의어로 바꾸면 AC가 무효가 되므로 문자 그대로 쓴다. |
| C7 | **C3는 절대 손으로 편집하지 않는다.** 생성 산출물이므로 C2를 고친 뒤 `make agents-emit` 으로 재생성한다. `make build` 는 이 타깃을 부르지 않으므로 재생성은 별도 단계이며, 생략하면 다음 방출이 수리를 되돌린다. |

---

## §D. 범위 제외

### Out of Scope — 스킬 층의 같은 형태 지시

아래 지점들도 전량 스위트를 지시하지만, 같은 drift 인지 이 카드에서 판정하지 않았다. 별도 카드 소관이다.

- `.claude/skills/moai-workflow-ddd/SKILL.md` — 2개 지점
- `.claude/skills/moai-workflow-tdd/SKILL.md` — 1개 지점
- `.claude/skills/moai/workflows/sync/quality-gates-*.md`
- `.claude/skills/moai/workflows/gate.md`

### Out of Scope — `.codex/` 미러의 나머지 10개 에이전트

- `plan-auditor.toml` 이 전량 호출을 담고 있음이 관측됐으나, 이 SPEC의 대상은 `manager-develop` 정의 한 갈래다.
- 나머지 `.toml` 들이 각자의 `.md` 원본과 어떤 동기화 규율로 묶이는지도 조사하지 않았다 — 그 규율 자체가 별도 소관이다.
- **비대칭이 아니다**: `manager-develop.toml` 은 이 SPEC이 고치는 바로 그 정의의 사본이라 두면 수리가 무효가 되지만, 형제 파일들은 다른 정의라 이 SPEC의 수리와 무관하다.

### Out of Scope — 배치 참조 파일의 예시 산문

- `agent-common-protocol-reference.md:65`(직렬 안티패턴 예시)와 `:76`(의존 관계 예시)의 전량 호출은 처방이 아니라 예시다. 건드리지 않는다. §A.3 참조.

### Out of Scope — 로컬 전용 파일 내부 불일치

- `CLAUDE.local.md` §6 "Go Test Execution Rules" 는 전량 실행을 `[HARD]` 금지한 직후, 바로 다음 두 줄에서 캐시 무효화·레이스 검사용 전량 실행을 권한다. 같은 절 안에서 금지와 권장이 공존한다.
- 로컬 전용 파일이라 배포 영향이 없고, 이 카드의 대상은 배포 계약 층의 drift다. 별도 카드로 분리한다.

### Out of Scope — 배치 예시의 프로그래밍 언어 중립화

- `agent-common-protocol-reference.md` 의 배치 예시는 이미 `(Go)` 라고 명시된 Go 예시이며, 2번 항목도 Go 형태다. 이 SPEC은 그 예시를 16개 언어 중립 형태로 일반화하지 않는다 — 기존 편향이지 이 SPEC이 만드는 편향이 아니다.
- 대신 에이전트 정의 쪽 대체 문면은 언어 중립 산문(`the tests the change can affect`)으로 고정한다. 지시문 층은 중립, 예시 층은 기존대로다.

### Out of Scope — Go 코드 · 도구 변경

- 검증 배치를 실행하는 Go 코드, 훅, CI 워크플로 정의는 건드리지 않는다.
- `moai gate` / `moai verify` CLI 동작 변경 없음.
- `verification-batch-pattern.md` 의 시간 추정치를 **실측으로 대체**하는 것은 범위 밖이다(측정하려면 C4가 금지한 행위가 필요하다). 근거 없는 수치를 걷어내는 것까지가 범위다.

### Out of Scope — 리포 전반 전량-스위트 감사

- 위 4개 파일 밖의 전량-스위트 지시 전수 조사는 하지 않는다.

---

## §E. 교차 참조

- 조사 기록: `.moai/reports/t301/discovery.md` · 감사 기록: `.moai/reports/t301/plan-audit.md`, `plan-audit-iter2.md`
- 상위 계약: `AGENTS.md` §4 · `.claude/rules/moai/workflow/kanban-dispatch.md` § Verification load is lane-local
- 증거 규율: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 · §2
- Template-First: `CLAUDE.local.md` §2 · §2.3 · `.github/workflows/template-neutrality-check.yaml`
- 지역 브랜치 모델(배포되지 않음): `.claude/rules/local/gitflow-lane-protocol.md` §4 · §8
