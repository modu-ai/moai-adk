---
id: SPEC-CODEX-GATE-SCOPE-001
title: "codex 리뷰 게이트의 검사 대상을 세션 스코프에 따라 한정한다"
version: "0.1.0"
status: completed
created: 2026-10-01
updated: 2026-10-01
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-CODEX-REVIEW-TARGET-001, SPEC-MOAI-MCP-SERVER-001, SPEC-DUAL-HARNESS-HOOK-PARITY-001, SPEC-WORKTREE-STATE-ROOT-001]
tags: "codex, review-gate, stop-hook, factory, lane, worktree, card-diff, scoping, t1383"
---

# SPEC-CODEX-GATE-SCOPE-001 — codex 리뷰 게이트의 검사 대상을 세션 스코프에 따라 한정한다

카드: **t1383** (리드 발행 2026-10-01 — 카드 본문에 worker-63 구조 관측 기재, §A.6 출처 표기)

## HISTORY

- 2026-10-01 · v0.1.0 · manager-spec · 최초 작성. 카드 t1383. 측정 원천: 본 트리 `.claude/worktrees/t1383`(브랜치 `WT-gate-scope-lane`)에서의 코드 좌표 직접 판독(spec.md §A). 라이브 큐 카드 t1373·t1378·t1383 본문 판독 — **미커밋 출처**로서 REQ 근거 인용 축을 decision-index 에 표기(Q2·Q4·Q5). 인접 SPEC 충돌 사전 검사: SPEC-CODEX-REVIEW-TARGET-001(completed)의 REQ-CRT-006 회귀선과 정합 — §F.

---

## §0 지배 원칙 [HARD]

> **게이트는 자기가 귀속할 수 없는 변경을 검사하지 않는다.**

현행 게이트의 검사 대상은 "해상된 트리의 미커밋 전체"뿐이다. 세션이 카드 세션일 때 그 트리가 공유 primary 체크아웃이면, 검사 대상에 **다른 세션의 WIP**가 섞이고 그 판정은 그 세션이 귀속하거나 고칠 수 없는 내용에 대해 내려진다. fail verdict 는 receipt 로 기록돼 같은 트리 상태 키를 공유하는 모든 세션의 턴을 재차단한다. 이 SPEC 이 닫는 것은 **검사 대상을 세션 스코프에 따라 가르는 일** 하나다. 리뷰 품질·판정 문구·receipt 형식 자체는 별개 축이다.

---

## §A 배경 (측정)

좌표는 모두 본 트리 기준이며 줄 번호는 2026-10-01 판독값이다(D2 — 측정 HEAD 핀: `f4aa9bf99f8fd343d83e1178629450d61510bc25`, 브랜치 `WT-gate-scope-lane`; plan-audit iter1 감사 트리와 동일).

### A.1 게이트의 구조 — 두 실행 경로가 하나의 대상 개념을 공유한다

- **Claude 경로** — Stop 훅 `moai hook codex-review-gate`. 셸 래퍼 `.claude/hooks/moai/handle-codex-review-gate.sh`가 stdin 을 전달하고 `workflow.codex.review_gate.enabled`(기본 OFF)를 순수 셸로 읽어 off 면 exit 0. 실 로직은 `internal/cli/codex_review_gate.go` `HandleCodexReviewGate` — 판정 순서 6단(비활성 → `stop_hook_active` → 셀프게이트 → reviewer 부재 → pass/inconclusive → fail BLOCK, `:67-111`).
- **Codex 경로** — Codex Stop 체인 멤버 6(`internal/cli/codex_stop_chain.go:613-658`)이 **receipt 를 읽는다**. 리뷰 실행은 훅 밖 `moai verify codex-review`(상수 `codexwiring.CodexReviewReceiptCommand`, `internal/codexwiring/stop_budget.go:89`)이고, `internal/cli/codex_review_receipt.go` `produceCodexReviewReceipt` 가 같은 review/start RPC 를 호출해 verdict 를 receipt 로 기록한다(SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d).

두 경로는 같은 검사 대상 개념(§A.2)을 공유하므로 스코핑은 양쪽에 동일하게 적용돼야 한다(REQ-CGS-009).

### A.2 검사 대상 — 해상된 트리의 미커밋 전체

`codex_review_gate.go:90-95`: `runCodexReviewRPC` 호출이 `"target": codexTargetUncommitted, "cwd": projectDir` 로 고정돼 있다. 검사 대상은 **projectDir 로 해상된 그 트리의 미커밋 변경 전체**이고, 세션 식별·브랜치·카드 소속은 입력으로 존재하지 않는다. receipt 생산자도 동일하다(`codex_review_receipt.go:96-99`).

### A.3 projectDir 해상 — 레인이 primary 로 귀결되는 두 경로

`resolveProjectDirFromInput`(`codex_review_gate.go:252-262`)은 input.ProjectDir → input.CWD → `CLAUDE_PROJECT_DIR` env 순으로 해상한다. 레인 세션이 다음 두 형태 중 하나면 해상값이 **공유 primary 체크아웃**이 된다:

1. 세션 cwd 가 primary 에 있는 경우(통합 창 진입 중, 워크트리 진입 전 등), 또는
2. spawn 시점에 freeze 된 `CLAUDE_PROJECT_DIR` 가 primary 를 가리키는 경우 — 세션이 자기 카드 워크트리에 앉아 있어도 env 가 이긴다.

### A.4 셀프게이트 — "검사할 것이 있는가"도 같은 전체 스코프로 잰다

`reviewGateChangeDetector = hasReviewableChanges`(`:49`, 구현 `:126-176`)는 `git status --porcelain` 을 projectDir 에 대해 돌리고 `reviewGateRuntimePrefixes`(`:36-43` — `.moai/state/` 등 6개)만 제외한다. primary 의 다른 세션 WIP(예: 이전 카드 빌드가 남긴 `internal/web/fieldsets_templ.go` 변형)도 "검사할 것"으로 센다 — 셀프게이트에는 스코프 개념이 없다.

### A.5 재차단 고리 — 한 리뷰의 fail 이 트리 상태 키로 모두를 묶는다

receipt 의 바인딩 키는 `verify.Key(root)` 가 산출하는 **HEAD + 트리 digest**다(`codex_review_receipt.go:58-72, 88`; 멤버 6 판독 `codex_stop_chain.go:631-647`). fail verdict 가 기록되면 같은 키(= 같은 트리 상태)를 공유하는 **모든 세션**의 멤버 6 이 차단된다. 트리 상태가 변하지 않으면 판정도 변하지 않으므로, 귀속 불가한 WIP 가 primary 에 머저 있는 동안 레인 턴은 반복 차단된다.

### A.6 재현 — 레인 턴이 귀속 불가 WIP 로 막히는 구조 (카드 t1383 R4)

메커니즘 사슬(§A.2-§A.5 의 조합, 본 트리 측정):

1. 레인 세션의 턴이 끝난다. Stop 페이로드의 cwd 또는 freeze 된 `CLAUDE_PROJECT_DIR` 가 primary 를 가리킨다(§A.3).
2. 게이트가 projectDir = primary 를 해상한다. 셀프게이트가 primary 의 미커밋을 센다 — 그 안에 이 세션과 무관한 WIP 가 있다(§A.4).
3. review/start 가 `cwd=primary` 로 날아가고 codex 는 **타 세션 WIP 를 포함한 diff** 를 본다(§A.2).
4. fail 이면 receipt 가 primary 트리 상태 키로 기록되고, 멤버 6 이 그 키로 **모든 세션**의 다음 턴을 재차단한다(§A.5). 카드 t1379 실측에서는 같은 codex findings 가 워크트리를 옮겨도 레인 턴을 반복 막았다.

**출처 표기**: "worker-63 의 구조 관측" — primary 더티 상태의 대가를 모든 레인 턴이 지불한다는 서술 — 는 카드 t1383 본문에 리드가 기재한 것이다. 본 SPEC 작성 레인은 그 관측 자체를 독립 재현하지 않았다. 본 절이 담보하는 것은 위 메커니즘 사슬(코드 좌표로 측정)까지이며, 관측 서술은 **리드 제공**으로 표기한다.

### A.7 인접 발견 — 이 SPEC 의 범위 밖 (§E 참조)

- 래퍼 쌍둔(추적본 `.claude/hooks/moai/` vs 템플릿 `internal/template/templates/.claude/hooks/moai/`)의 차이는 **주석·공란뿐**이다 — 본 트리 `diff` 실측, 비주석 행 0. 양 쌍둔의 마지막 공통 변경 커밋은 `86bbfde7b`(본 트리 `git log -1` 실측 — 카드가 전달한 `86bfde7b` 표기의 오탈자를 이 실측이 정정한다).
- `internal/cli/codex_sync_gate.go`(sync 게이트의 Go 미러)는 pre-t1379 결함을 그대로 갖고 있다 — `:282` `syncGateChecks[langs[0]]` 단일 언어 선택, `:167-172` kotlin/java switch-elif. t1379 의 셸 쪽 수리는 Go 미러에 착지하지 않았다.

---

## §B 요구사항 (GEARS)

### REQ-CGS-001 — 스코프 판별이 검사에 앞선다 (Where + When)

**Where** the codex review gate is enabled (`workflow.codex.review_gate.enabled`), **When** a session turn reaches the gate, the review gate shall determine the session's scope class — card-scope or tree-scope — from the discriminator criteria, before running or consulting any review.

### REQ-CGS-002 — 카드 세션의 검사 대상은 카드 diff 다 (While)

**While** the session's scope class is card-scope, the review gate shall scope the review to the card diff: the changes from `git merge-base develop HEAD` to the card branch HEAD — recomputed at each gate evaluation, never a pinned SHA — plus the uncommitted changes of the card worktree, excluding runtime-managed prefixes.

merge-base 규율의 정본은 `.claude/rules/local/gitflow-lane-protocol.md` §8 이다 — 흡수가 일어나도 판정식이 흡수 전 분기점에 머물지 않도록 매 평가 시점에 다시 구한다. 카드 diff 의 정의는 **운영자 확정**(2026-10-01, decision-index Q4 CONFIRMED)을 받았다: **합집합** — 커밋분(`git merge-base develop HEAD` 부터 카드 브랜치 HEAD 까지) ∪ 카드 워크트리의 미커밋분(runtime-managed 접두어 제외).

### REQ-CGS-003 — 카드 세션이 아닌 세션의 대상은 변하지 않는다 (While, 회귀)

**While** the session's scope class is not card-scope, the review gate shall scope the review to the whole uncommitted changes of the resolved tree, shape-identical to its form before this SPEC — the same path SPEC-CODEX-REVIEW-TARGET-001 REQ-CRT-006 pins.

리더 세션을 포함한 비카드 세션(리더·일반 개발 세션)의 동작을 이 요구가 지킨다. 현행 경로를 뜯어고치는 것이 아니라 **앞에 판별을 얹는** 구조다.

### REQ-CGS-004 — 판별 입력은 카드 브랜치뿐이다 (Ubiquitous + shall not)

The scope discriminator shall decide card-scope identity solely from card-branch detection — the session tree's current branch carrying the `WT-` prefix, the committed lane-protocol invariant. The discriminator shall not derive the scope from session environment labels, and shall not match any lane label format.

env 를 결정 입력에서 뺀 근거는 카드 t1373 실측(/clear 뒤에도 구 라벨 잔존 — retire 된 런 소속 라벨이 남는다; 라이브 큐 기재, 미커밋 출처)과 t1378 의 라벨 네임스페이스 재구성 예정이다. env 는 REQ-CGS-010 의 관측 맥락으로만 남는다. 이 결정(env 강등 — 관측 전용)은 **운영자 확정**을 받았다(2026-10-01, decision-index Q2 CONFIRMED).

### REQ-CGS-005 — 스코프는 세션 트리에서 정해진다 (When)

**When** the hook's project directory and the session's working directory resolve to different trees, the review gate shall determine the scope from the session's working directory, and shall not review changes of the project directory that lie outside the determined scope.

§A.3 의 (b) — freeze 된 `CLAUDE_PROJECT_DIR` — 가 이 요구의 표적이다.

### REQ-CGS-006 — 셀프게이트가 같은 스코프를 잰다 (When, event-detected)

**When** the scoped diff under the determined scope class is empty, the review gate shall allow the turn without invoking the reviewer; the self-gate shall measure the same scope the review would measure.

### REQ-CGS-007 — 판정은 자기 스코프 상태에 묶인다 (Ubiquitous + When)

A recorded review verdict shall be bound to the state of the scope it was produced from. **When** a turn's scope state matches a recorded verdict, the gate shall reuse it without re-running the review; a verdict recorded for one scope state shall not block a turn whose scope state differs.

§A.5 의 재차단 고리를 스코프 경계로 가른다. 같은 카드의 같은 diff 상태에 대한 재차단은 유지된다 — 그것이 이 고리의 설계 목적이다.

### REQ-CGS-008 — fail-open 은 모든 스코프에서 유지된다 (When, 회귀)

**When** the reviewer is missing, the review call errors, or the verdict is inconclusive, the review gate shall allow the turn in every scope class — the fail-open contract of SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012, unchanged.

### REQ-CGS-009 — 두 실행 경로가 같은 스코프를 본다 (Ubiquitous)

The receipt producer command (`moai verify codex-review`) shall resolve and review the same scope the turn-end path resolves for the same session state, and the two paths shall not be able to disagree on the scope for that state.

### REQ-CGS-010 — 스코프 판정은 관측된다 (When)

**When** the gate determines the scope class, it shall record the class and its basis — branch match, or none — in the gate's structured log output, distinguishable per class, with the session's factory env labels carried as context only.

---

## §C AC 형태에 대한 구속 [HARD]

- AC 는 **게이트가 조립한 검사 요청의 대상 필드**(target·cwd)와 **receipt 의 바인딩 상태**를 관측한다. verdict 값 단독은 어떤 스코프 AC 의 근거도 되지 못한다 — 스텁·픽스처 codex 는 요청과 무관하게 스크립트된 값을 돌려줄 수 있다.
- 스코프 판별기는 기존 seam(`reviewGateChangeDetector` 선례)과 같은 **주입형 순수 함수**로 검증 가능해야 한다 — 라이브 codex 없이 판별 행동을 고정하고, 라이브 의존 AC 는 두지 않는다.
- `WT-` 픽스처 워크트리는 실제 git worktree + 카드 커밋 1건 + 미커밋 1건 + primary 측 외부 WIP 로 구성한다. 셀렉터 0매칭 초록 방지: RED 실행은 `-v` 로 `=== RUN` 을 함께 관측한다.

## §D 실행 순서 구속

트리 스코프 회귀선(REQ-CGS-003·008 — 현행 경로)이 초록으로 고정된 뒤에 판별기와 카드 스코프를 얹는다. 두 경로가 같은 핸들러를 지나므로, 회귀선 없는 수정은 고친 것과 깨뜨린 것을 구별하지 못한다.

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — `codex_sync_gate.go` 의 pre-t1379 결함 (Go 미러)

- `:282` `syncGateChecks[langs[0]]` 단일 언어 선택, `:167-172` kotlin/java switch-elif — t1379 가 셸 쪽만 수리했다(§A.7).
- 이 SPEC 은 review 게이트의 **대상 스코핑**만 다룬다. sync 게이트의 언어 선택 로직은 별도 카드 소관이며, 같은 패키지라는 이유로 묶지 않는다.

### Out of Scope — 래퍼 쌍둔의 주석 차이 정리

- 추적본과 템플릿본의 주석 차이(§A.7)는 pre-existing 이다. 이 카드가 래퍼를 건드릴 계획이 없는 한 그 차이를 정리하지 않는다 — 무음 수정은 템플릿 쌍둔 관측의 기준을 흔든다.

### Out of Scope — t1373·t1378 의 자체 수리

- t1373(env 라벨 잔존 수리)·t1378(슬롯 상한 자동 성장·라벨 네임스페이스 재구성)은 각자의 카드가 소유한다. 이 SPEC 의 정합 방식은 REQ-CGS-004 하나 — 라벨 값을 소비하는 코드를 어디에도 두지 않음으로써 두 재구성과 충돌하지 않는다.

### Out of Scope — primary WIP 의 귀속 추적

- "이 uncommitted 변경이 어느 세션의 것인가"를 추적하는 귀속 메커니즘은 만들지 않는다. 이 SPEC 의 답은 귀속이 아니라 **스코프 배제**다 — 카드 세션은 자기 diff 만 보고, 귀속 불가 변경은 애초에 대상에 들어오지 않는다.

### Out of Scope — 리뷰 내용·판정 기준·게이트 기본값

- codex 리뷰가 무엇을 좋다고 보는지, verdict 문구, 900s 타임아웃 값은 건드리지 않는다. `workflow.codex.review_gate.enabled` 기본 OFF 도 유지한다 — 이 SPEC 은 켜져 있을 때의 **대상 선정**만 바꾼다.

---

## §F 제약

- **REQ-CRT-006 회귀선 정합**: 트리 스코프 경로의 `uncommittedChanges` 요청은 본 SPEC 이후에도 shape-identical 이어야 한다(SPEC-CODEX-REVIEW-TARGET-001 이 고정한 형태). 카드 스코프는 그 경로를 대체하지 않고 별도 대상으로 얹는다.
- **fail-open 불변**(REQ-MCP-012): reviewer 부재·오류·inconclusive 는 어느 스코프에서든 ALLOW 다. 스코핑은 판정 방향을 바꾸는 틈을 만들지 않는다.
- **Template-First**: 구현이 셸 래퍼를 건드리는 경우에만 — 추적본과 `internal/template/templates/` 쌍둔을 같은 커밋에 넣고 `make build` 를 통과시킨다. plan.md §D 가 래퍼 비접촉을 기본 전제로 둔다.
- **WT- 불변식이 하중을 진다**: 판별기의 1차 신호는 "카드 워크트리 브랜치는 `WT-` 접두사를 갖는다"는 커밋된 레인 규율(kanban-dispatch.md § Isolation, gitflow-lane-protocol.md §1)이다. **명시된 페일오픈(운영자 확정, 2026-10-01 — decision-index Q5 CONFIRMED)**: WT- 카드 브랜치를 검출할 수 없는 식별 불가 세션(병합 창의 release 브랜치, primary 체크아웃 등)은 **트리 스코프로 페일오픈**하며 현행 동작을 유지한다 — 이 페일오픈은 명시된 동작이지 모호함이 아니다. 레인이 카드 워크트리 밖(cwd=primary)에서 일하는 것 자체가 레인 규율 위반이므로, 그 상태에서의 트리 스코프는 현행과 같은 동작이다.
- **측정 규율**: 검증은 대상 패키지(`./internal/cli/...`) 스코프로 돌린다 — 전체 스위트 로컬 금지(레인 부하 규율). merge-base 기반 범위 판정식은 병합 전 평가 전용이다(gitflow-lane-protocol.md §8 한계).

## §G 참조

- `internal/cli/codex_review_gate.go` — 게이트 본체(§A.1-§A.4 좌표)
- `internal/cli/codex_stop_chain.go:613-658` — 멤버 6 receipt 판독(§A.5)
- `internal/cli/codex_review_receipt.go` — receipt 생산자(REQ-CGS-009 설계 표면)
- `internal/codexwiring/stop_budget.go:89` — `CodexReviewReceiptCommand`
- `internal/config/envkeys.go:299` — `EnvMoaiFactoryWorker`(관측 맥락 상수 — 문자열 리터럴 금지)
- `.claude/rules/local/gitflow-lane-protocol.md` §1·§8 — WT- 불변식·merge-base 규율
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Isolation — WT- 브랜치 명명 [HARD]
- 라이브 큐 카드 t1373·t1378·t1383 본문 — 미커밋 출처(인용 축별 권위 표기는 decision-index.md)
- 관련 SPEC(전부 completed): SPEC-CODEX-REVIEW-TARGET-001(REQ-CRT-006 회귀선) · SPEC-MOAI-MCP-SERVER-001(REQ-MCP-012 fail-open) · SPEC-DUAL-HARNESS-HOOK-PARITY-001(멤버 6 receipt 방식) · SPEC-WORKTREE-STATE-ROOT-001(설정 루트 해상)
