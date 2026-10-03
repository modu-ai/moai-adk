---
id: SPEC-CODEX-GATE-SCOPING-001
title: "codex 리뷰 게이트의 비카드 검사 한계를 primary 체크아웃과 런타임 관리 표면으로 좁힌다"
version: "0.1.1"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-CODEX-GATE-SCOPE-001, SPEC-CODEX-REVIEW-OWNERSHIP-001, SPEC-CODEX-REVIEW-TARGET-001, SPEC-WORKTREE-STATE-ROOT-001]
tags: "codex, review-gate, stop-hook, primary-checkout, runtime-drift, sync-gate, wci-excludes, t1404"
---

# SPEC-CODEX-GATE-SCOPING-001 — codex 리뷰 게이트의 비카드 검사 한계를 primary 체크아웃과 런타임 관리 표면으로 좁힌다

카드: **t1404** (리더 배차 — 턴 끝 codex 리뷰 게이트의 비카드 세션 스코핑 구멍 3면)

## HISTORY

- 2026-10-03 · v0.1.0 · manager-spec · 최초 작성. 측정 원천: 본 트리 코드 좌표 직독(측정 HEAD 핀: `2de0a2cb6`, 브랜치 `WT-codex-gate-scope`, 2026-10-03) + primary 체크아웃의 카드 t1395 게이트 차단 처분 기록 9건(§A.1, 읽기 전용 인용). 인접 SPEC 충돌 사전 검사: `SPEC-CODEX-GATE-SCOPE-001` 존재 확인( completed, t1383 — 본 SPEC은 §F.5의 관계로 흡수) 및 `SPEC-CODEX-GATE-SCOPING-001` 공석 확인. Facet-1 설계 결정(배제 방향)은 §F.1에 기록.
- 2026-10-03 · v0.1.1 · manager-spec · plan-audit iteration-1 FAIL 수리(D1-D5). D1: release-blocking AC의 RED-now를 현 트리(`2de0a2cb613b04765a1554f86685a3b48e0be806`)에서 실제 관측 — plan 단계 저작 테스트(`internal/cli/codex_review_gate_primary_scope_red_test.go`, `internal/template/hook_gate_reports_exclude_test.go`) 2파일, 원문 stdout·exit 코드는 acceptance.md §D.0 관측 장부로. 구현 전 적색이 정의상 존재할 수 없는 보존 기준 4건(acceptance AC-002·003·009·012)은 §2.1 undecidable 처분에 따라 regression-guard로 재분류(acceptance §D.1). D2: `primary_scope` 판독 결과별 처분표를 §F.2에 확정하고 세 산출물을 그 표에 정합, §F.2의 REQ-CGSC-006 인용 과잉을 정정. D3: REQ-CGSC-009의 수집 경로 열거에 맞춰 plan M4 수단을 델타 삼팔 pathspec 확장으로 정합. D4: AC-CGSC-012의 계수 검사를 줄 수가 아닌 항목 멤버십 기준으로 교체(실측 baseline 반영). D5: §A.4 주석 좌표 :938/:954로 정정.

---

## §0 지배 원칙 [HARD]

> **게이트는 자기가 귀속할 수 없는 변경을 검사하지 않는다 — 그리고 자기가 검사한 트리가 살아 있는 트리임을 입증하지 못하면 검사하지 않는다.**

선퓘 SPEC-CODEX-GATE-SCOPE-001이 세운 귀속 원칙을 카드 세션까지 적용한 것이 이 SPEC이다. 측정된 실패는 그 원칙이 **비카드 경로**에서 그대로 깨지는 형태다: 카드 워크트리를 나와 primary 체크아웃에 앉은 레인 세션은 비카드 세션과 구별되지 않고, 게이트는 그 트리의 스테일 main-era 사본을 검사해 이미 develop에서 수리된 결함을 재기-flag 했다. 9번의 턴 차단, 38건의 발견, 28건의 스테일 트리 관측, develop 소스 검증 생존 0건(§A.1)이 그 대가다. 이 SPEC이 닫는 것은 셋이다 — ① primary 체크아웃의 비카드 검사 생략, ② 런타임 관리 설정 표면의 검사 결함 분류 금지, ③ sync 게이트 수집에서 `.moai/reports` 배제. 리뷰 품질·판정 문구·receipt 형식·카드 스코프 경로는 별개 축이다.

---

## §A 배경 (측정)

좌표는 모두 본 트리 기준이며 줄 번호는 2026-10-03 판독값이다(측정 HEAD 핀: `2de0a2cb6`, 브랜치 `WT-codex-gate-scope`).

### A.1 측정 요약 — t1395 게이트 차단 처분 기록 9건

primary 체크아웃의 `.moai/reports/t1395/gate-block-disposition.md` ~ `gate-block-disposition-9.md`(읽기 전용 인용; 본 카드 증거 사본은 본 트리 `.moai/reports/t1404/gate-block-evidence.md`)가 9번의 턴 끝 차단을 기록한다. 누적 행(제9건 원문): **"Cumulative: 9 blocks, 38 findings — 28 stale (0 survived), 4 REAL (reports-exclusion → t1404; gpt doc drift → t1406; triage -F → NEW candidate; .cxx detection → NEW candidate), 1 local drift (auto_cleanup → t1405), 1 plausible residual (deleted-module fallback → t1412), 1 intentional policy (t609 — operator posture)"**. 본 SPEC의 3면은 이 중 t1404에 귀속된 축이다.

### A.2 Facet 1 — primary 체크아웃 착지 세션의 스테일 트리 검사

- 세션 트리 해상: `reviewScopeSessionDir`(`codex_review_scope.go:144-154`)가 input.CWD 우선으로 세션 트리를 고른다. 카드 워크트리를 나온 레인이 primary에 앉아 있으면 CWD=primary고, 브랜치는 `main`(비 WT-)이라 스코프 클래스가 **tree**로 결정된다(`resolveReviewScope` `:84-104`).
- tree 클래스의 검사 요청은 `reviewRequestParams`(`:169-180`)가 `"target": uncommittedChanges, "cwd": primary`로 고정된다 — primary의 미커밋 전체가 검사 대상이다.
- primary는 공유 체크아웃이고 커밋사 mag인 `main`(SPEC-MAIN-COMMIT-BAN-001)에 앉아 있어, 그 사본은 develop이 수리한 결함을 여전히 품은 스테일 세대다. 처분 기록 #1-#4가 이 조합의 재현을 10건 이상 기록한다(이미 수리된 결함의 재flag, 전부 develop 소스 검증으로 기각).
- 선행 SPEC-CODEX-GATE-SCOPE-001 §F는 이 상태를 운영자 확정(2026-10-01)으로 "트리 스코프 페일오픈 — 현행 동작 유지"로 두었다. 본 SPEC의 측정이 그 팔을 반전한다(§F.5).

### A.3 Facet 2 — 런타임 관리 표면의 검사 결함 분류

- 트리 스코프의 검사 가능 판정(`reviewableFromPorcelain` `codex_review_gate.go:165-183`)과 경로 필터(`isRuntimeManagedPath` `:188-195`, 공유 리스트 `reviewGateRuntimePrefixes` `:36-43` — `.moai/state/` `.moai/cache/` `.moai/reports/` `.moai/logs/` `.moai/harness/` `.claude/agent-memory/`)는 런타임 **상태** 표면은 이미 배제하지만, 런타임이 쓰고 갱신하는 **설정** 표면은 배제하지 않는다.
- 측정된 피해: 처분 기록 #2 — 로컬 `.claude/settings.json:13`의 개인 PATH 항목(런처가 런타임에 주입; settings-drift 원장 추적 대상)이 P2 검사 결함으로 분류. 처분 기록 #4 — 로컬 `.moai/config/sections/workflow.yaml:65`의 `auto_cleanup: true` 로컬 드리프트(운영자 소유)가 P1으로 분류. 둘 다 레인이 저자도 수리자도 아니다.
- 공유 리스트 함정: `reviewGateRuntimePrefixes`는 카드 스코프 경로도 같이 쓴다(`cardChangedPaths` `codex_review_scope.go:209`, `cardScopeKeyParts` `:277`). 확장을 공유 리스트에 하면 카드 스코프가 변한다(REQ-CGSC-005가 금지).

### A.4 Facet 3 — sync 게이트 수집의 `.moai/reports` 배제 누락

- `.claude/hooks/moai/sync-phase-quality-gate.sh`의 `WCI_EXCLUDES`(:257 전후, 2026-10-03 판독)는 `.moai/state` `.moai/logs` 양 워크트리 트리, 의존성/빌드 디렉터리를 배제하지만 `.moai/reports`는 없다. 템플릿 쌍둔(`internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh`)도 동일.
- 처분 기록 #5 REAL-1: `.moai/reports/<id>/` 아래 파킹된 Go 픽스처(감사 랩 등 go.mod 소유)가 SYNC_DELTA_FILES와 모듈 탐색에 쓸려 들어 스퓨리어스 블록을 낸다. 게이트 내 "reports" 언급은 주석뿐(:938, :954). 수집 4팔 가운데 WCI_EXCLUDES pathspec을 전달하는 것은 ③ 미추적 walk(:309)와 ④ 무시된 원천(:278)뿐이고 ① 커밋 diff(:307)와 ② tracked diff(:308)는 전달하지 않는다(2026-10-03 직독) — 배제 항목 추가만으로는 REQ-CGSC-009가 요구하는 전 수집 경로 커버가 안 된다(plan M4 수단 정합 근거).
- Go 미러 `internal/cli/codex_sync_gate.go`에는 WCI_EXCLUDES 대응 수집 제외 목록 자체가 없다(2026-10-03 grep 실측, 적중 0) — 본 면의 표면은 셸+템플릿 쌍둔뿐이다.

### A.5 기존 정책과의 접점

- `treeScopeSkipApplies`(`codex_review_tree_scope.go:85-94`): tree 클래스 + 비 WT- 브랜치 + `tree_scope: skip` → 검사 전 ALLOW. 두 자동 경로(Claude Stop 훅, Codex Stop 체인 멤버 6)가 이 하나를 공유한다(REQ-CRO-006). 본 SPEC의 primary 생략은 이 정책과 **같은 자리**에서 함께 적용된다 — tree 클래스의 primary는 `tree_scope` 값과 무관하게 생략되고(REQ-CGSC-002), `tree_scope`는 계속 비-primary 트리 세션을 지배한다.
- 설정 읽기 선례: `readCodexReviewGateTreeScope`(`:54-75`) — 플랫/주석/오류 전부 review로 fail-closed. primary-scope 판독기는 이 형제다. 상수·노멀라이저는 `internal/config/defaults.go:616-635`, 기본값 `:1335`, 타입 `internal/config/types.go:958`.

### A.6 카드 t1383 관계 (record-only)

카드 지시에 따라 코드를 직독해 확인했다: "카드 세션을 카드 diff로 스코프"하는 메커니즘은 SPEC-CODEX-GATE-SCOPE-001(t1383, **completed**)이며, `reviewScopeResolver`·`cardScopeFromBranch`(WT- 접두 판별 `:109-111`)·`cardMergeBase`(`:128-137`)·`tree_scope` 정책이 본 트리에 착지돼 있다. 본 SPEC은 그 위에 얹히는 트리-클래스 확장이며, 미병합 t1383 작업에는 의존하지 않고 접촉하지도 않는다.

---

## §B 요구사항 (GEARS)

### REQ-CGSC-001 — 스코프 해상은 primary 여부를 함께 결정한다 (Where + When)

**Where** the codex review gate is enabled (`workflow.codex.review_gate.enabled`), **When** a session turn reaches the gate, the review gate shall resolve the session's scope class (card | tree) exactly as SPEC-CODEX-GATE-SCOPE-001 defines it, and shall additionally determine whether the session tree is the repository's primary working tree — the checkout whose git dir is the repository's common git dir — carrying that determination in the scope resolution output.

### REQ-CGSC-002 — primary 체크아웃의 비카드 검사는 생략된다 (Where + While + When)

**Where** the review gate is enabled and the primary-scope policy is at its default (skip), **While** the session's scope class is tree, **When** the session tree is the repository's primary working tree, the review gate shall allow the turn without invoking or consulting any review — decided before the self-gate — and shall record one structured skip row whose basis is distinguishable from the existing tree_scope skip.

### REQ-CGSC-003 — 배제 정책은 두 자동 경로가 하나를 공유한다 (Ubiquitous)

**Where** the review gate evaluates any session state, the primary-checkout exclusion shall be ONE policy consulted by both automatic gate paths — the Claude Stop-hook handler and the Codex Stop-chain member 6 receipt reader — and the two paths shall not be able to disagree on it for that state. The explicit producer (`moai verify codex-review`) shall keep its existing behavior.

### REQ-CGSC-004 — 명시 설정으로 primary 검사를 복원한다 (Where)

**Where** the primary-scope policy is explicitly configured to review, the review gate shall keep the pre-SPEC tree-scope behavior for the primary checkout — the existing `tree_scope` key's values, default, and read discipline unchanged. The restore applies only to a policy value that reads and parses as an explicit `review`; every other read outcome — key absent, unrecognized value, unreadable file, YAML parse error — leaves the default skip in force (the §F.2 policy table).

### REQ-CGSC-005 — 카드 스코프는 shape-identical이다 (Ubiquitous, 회귀)

A card-scope session's review shall remain shape-identical to SPEC-CODEX-GATE-SCOPE-001's card path — request target, self-gate, path filter set, and receipt binding unchanged. The primary-checkout exclusion shall not evaluate the card class, and the tree-only runtime-managed exclusion set shall not alter the card path's filter set.

### REQ-CGSC-006 — 판별 불가는 기존 페일오픈을 유지한다 (When, event-detected)

**When** the session tree's primary-versus-linked status cannot be established (a non-git directory, or a git error), the review gate shall not apply the primary-checkout exclusion and shall keep today's behavior for that state, reason carried in the log basis.

### REQ-CGSC-007 — 트리 스코프의 검사 가능 집합에서 런타임 관리 설정 표면을 뺀다 (While)

**While** the session's scope class is tree, the self-gate shall not count changes under the runtime-managed configuration surfaces — the local Claude settings file (`.claude/settings.json`) and the MoAI managed config tree (`.moai/config/`) — as reviewable, so a turn whose only changes are such surfaces is allowed without invoking the reviewer.

### REQ-CGSC-008 — 런타임 관리 표면만을 겨냥한 발견은 검사 결함으로 판정하지 않는다 (When, event-detected)

**When** every finding of a tree-scope review targets only paths under the runtime-managed configuration surfaces, the review gate shall not block the turn on those findings — the turn is allowed, and the reclassification (known runtime-managed drift, not a review defect) is recorded.

### REQ-CGSC-009 — sync 게이트 수집이 `.moai/reports`를 배제한다 (Ubiquitous)

The sync-phase quality gate's change-collection exclusion set shall exclude `.moai/reports/` in the same exclude-pattern style as its existing managed-tree entries, and the exclusion shall hold on every collection path that set feeds — the delta-file arms (the commit-range diff, the tracked uncommitted diff, and the untracked walk), the ignored-source collection, module discovery, and the content key — in the template mirror and the tracked hook as one change.

### REQ-CGSC-010 — 기존 sync 게이트 수집은 shape-identical이다 (While, 회귀)

**While** a collected path lies outside `.moai/reports/`, the sync gate's delta collection, language detection, and gate behavior shall be shape-identical to its pre-SPEC form, every existing exclusion entry unchanged.

### REQ-CGSC-011 — 생략과 재분류는 관측된다 (When)

**When** the gate skips a review for the primary-checkout reason or reclassifies findings as runtime-managed drift, it shall record one structured row per decision on the gate's diagnostic channel, distinguishable per reason, carrying the resolved tree and the policy values in effect.

---

## §C AC 형태에 대한 구속 [HARD]

- AC는 **관측 가능한 결과**를 관측한다 — 출력 결정(ALLOW/BLOCK), review 호출 여부, 구조화 로그 행, sync 게이트의 수집 집합. reviewer의 verdict 값 단독은 어떤 스코프·재분류 AC의 근거도 되지 못한다(스텁 codex는 스크립트된 값을 돌려줄 수 있다).
- 판별기·정책·재분류는 기존 seam 선례(`reviewScopeResolver`, `reviewGateTreeScopeReader`, `reviewGateChangeDetector`)와 같은 **주입형 순수 함수**로 검증한다 — 라이브 codex 없이 행동을 고정하고, 라이브 의존 AC는 두지 않는다.
- primary 판별 픽스처는 실제 git 저장소 + 연결 워크트리를 `t.TempDir()`에 만들어 git-dir/common-dir 쌍을 실측한다. RED 실행은 `-v`로 `=== RUN`을 함께 관측한다(셀렉터 0매칭 방지 — 검증-완결성 §1.1).
- sync 게이트 AC는 기존 `internal/template/hook_gate_*` 선례의 스크립트 픽스처 실행으로 관측한다 — `.moai/reports/lab/go.mod` + 깨진 `.go` 픽스처가 수정 전에는 수집에 들어가 블록을 내는 것(RED)을 먼저 관측한다.

---

## §D 실행 순서 구속

회귀선(REQ-CGSC-005·006·010 — 현행 경로)이 초록으로 고정된 뒤에 새 정책과 배제를 얹는다. 선행 SPEC §D와 같은 이유다: 회귀선 없는 수정은 고친 것과 깨뜨린 것을 구별하지 못한다. 각 요구는 RED-first(구현 전 실패 관측)로 채택한다.

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — `codex_sync_gate.go` Go 미러의 수집 정합

- 본 트리의 Go 미러에는 WCI_EXCLUDES 대응 수집 제외 목록이 없다(2026-10-03 grep 실측, 적중 0). 셸 게이트의 제외 확장이 Go 미러에 파장하지 않으며, 미러의 별도 결함(선행 SPEC §A.7 기록 축)은 이 카드가 묶지 않는다.

### Out of Scope — 카드 스코프 검사의 어떤 변경

- 카드 diff 정의, 카드 경로 필터, 카드 receipt 바인딩은 불변이다(REQ-CGSC-005). 런타임 관리 설정 표면의 카드 스코프 취급(예: 카드가 템플릿 설정을 고치는 경우의 원천 경로)은 본 SPEC이 다루지 않는다.

### Out of Scope — 리뷰 품질·판정 기준·게이트 기본값

- codex 리뷰가 무엇을 좋다고 보는지, verdict 문구, 900s 타임아웃은 건드리지 않는다. `workflow.codex.review_gate.enabled` 기본 OFF도 유지한다.

### Out of Scope — 런타임 관리 표면의 복원·정화

- settings-drift 독트린상 자동 복원은 금지돼 있다. 본 SPEC은 검사 범위에서 **빼는** 것일 뿐 표면 자체를 고치지 않는다. `auto_cleanup` 드리프트의 안전 판정은 t1405(검증 우선 카드)가 소유한다.

### Out of Scope — 비-primary 스테일 브랜치의 검사 판정

- primary가 아닌 연결 워크트리(릴리스 워크트리 등)의 낡은 브랜치는 계속 검사 대상이다. 측정된 실패면은 primary 착지 조합에 한정되며, 스테일성 일반 판별은 REQ-CGSC-002의 배제 방향 선택에서 기각됐다(§F.1).

### Out of Scope — 후속 카드 소관의 REAL 발견

- gpt 문서 드리프트(t1406), triage `-F` 누락, `.cxx` 탐지 누락, 삭제 모듈 루트 폴백(t1412) — 각자의 카드가 소유한다.

---

## §F 제약

### F.1 Facet-1 설계 결정 — primary 배제 (staleness 검출 기각)

배제 방향을 채택하고 staleness 검출을 기각한다. 근거 넷: (1) **귀속 정합** — 선행 SPEC §0의 원칙은 귀속이고, primary 체크아웃은 공유 공간이라 신선해도 귀속 불가다. staleness는 귀속과 다른 축이라, 신선한 primary 위의 타세션 WIP을 여전히 검사한다 — 측정되지 않은 같은 실패의 재생이다. (2) **판별식의 노후** — staleness 판정은 살아있는 develop 비교를 매 평가에 필요로 하고, 그 비교는 그 자체가 스테일해질 수 있다(t1412 cache-key 계열이 재는 축). (3) **기계성** — git-dir/common-dir 동일성은 불변이고 비교 대상 ref가 없다. (4) **측정 적합** — 9블록 전부가 primary-stale 조합이다. 배제는 실패면을 정확히 덮는다.

### F.2 fail-open 방향 불변 + primary-scope 정책 판독 처분표

reviewer 부재·오류·inconclusive는 어느 스코프에서든 ALLOW다(REQ-MCP-012). 본 SPEC의 모든 축소는 비카드/primary 방향만이고, 카드 세션 검사를 더 공격적으로 만들지 않는다(카드 제약 2).

`primary_scope` 정책 값의 판독 결과별 처분은 아래 표 하나로 정의되며, 세 산출물(spec §F.2 · acceptance §B.3 · plan §F M1)이 모두 이 표를 인용한다. 세션 트리의 git 프로브 실패(비-git 디렉터리·git 오류)는 별개 축으로 REQ-CGSC-006 그대로 생략 미적용(페일오픈)이다 — 이 표는 정책 **값**의 판독만을 다룬다.

| 판독 결과 | 처분 | 비고 |
|---|---|---|
| 명시적 `review` (파싱 성공) | 생략 미적용 — pre-SPEC 트리 동작 복원 | REQ-CGSC-004의 유일한 복원 축 |
| 명시적 `skip` (파싱 성공) | 생략 적용 | 기본값과 같은 방향 |
| 키 부재 (플랫·주석·다른 블록 포함) | 생략 적용 — 기본 skip | 기본값이 곧 정책(REQ-CGSC-002) |
| 미지/부정합 값 | 생략 적용 — 기본 skip | 파싱 성공한 명시 `review`가 아니므로 복원 없음 |
| 파일 읽기 오류 | 생략 적용 — 기본 skip | 스킵 행 basis에 판독 실패 명기(REQ-CGSC-011) |
| YAML 파스 오류 | 생략 적용 — 기본 skip | 스킵 행 basis에 판독 실패 명기(REQ-CGSC-011) |

**선택 방향과 근거(1문장)**: 판독 불가 갈래는 전부 기본 skip으로 귀결된다 — 게이트는 귀속할 수 없는 primary 변경을 검사하지 않는다(§0·§F.1)는 귀속 원칙과, 형제 판독기(`readCodexReviewGateTreeScope`)의 "명시적 비기본값만이 기본값을 이긴다" 규율을 뒤집힌 기본값(skip이 기본)에 그대로 적용한 결과이며, 형제 규율이 실은 「조용히 검사가 줄어드는 일이 없게」를 지향하므로 그 잔여 우려는 REQ-CGSC-011의 구조화 스킵 행(basis에 판독 실패 명기)이 흡수한다.

### F.3 Template-First + 미러 중성성

sync 게이트 변경은 템플릿 미러 선행 → `make build` → 추적본 동일 변경, 한 커밋. 미러에 **새로 추가되는** 주석에 카드 id·SPEC id·내부 날짜·커밋 SHA·macOS 편향 경로를 넣지 않는다(C1-C8). 미러에 이미 존재하는 카드 표기 6건(grep 실측 2026-10-03, 양 쌍둔 모두)은 pre-existing이며 본 카드가 확장하지 않는다.

### F.4 측정 규율

검증은 변경 대상 패키지 스코프로 돌린다 — `./internal/cli/... ./internal/config/... ./internal/template/...`. 전체 스위트(`go test ./...`) 로컬 금지(레인 부하 규율). 판정 테스트 출력은 tail로 자르지 않는다(실패 이름과 exit 코드가 같이 가려진다).

### F.5 선행 SPEC 관계 (record-only + 반전 선언)

- **t1383 = SPEC-CODEX-GATE-SCOPE-001 (completed)**: 카드 스코핑 메커니즘. 코드 직독으로 본 트리 착지 확인(§A.6). 본 SPEC은 그 위에 얹히며, 미병합 t1383 작업에 의존하지 않는다. 선행 SPEC 본문은 수정하지 않는다.
- **반전 선언**: 선행 SPEC §F가 운영자 확정(2026-10-01)으로 둔 "primary 체크아웃 → 트리 스코프 현행 유지" 페일오픈 팔을, 본 SPEC은 §A.1의 측정(9블록/38발견/28스테일/0생존)을 근거로 REQ-CGSC-002로 반전한다. 이 반전의 운영자 추인 축은 decision-index Q1에 남긴다.

## §G 참조

- `internal/cli/codex_review_gate.go` — 게이트 본체, 공유 런타임 prefix 리스트(:36-43), 자기게이트 파서(:165-195)
- `internal/cli/codex_review_scope.go` — 스코프 해상·카드 diff(§A.2 좌표)
- `internal/cli/codex_review_tree_scope.go` — tree_scope 정책·판독기(§A.5 좌표)
- `internal/config/types.go:958`, `internal/config/defaults.go:616-635·1335` — tree_scope 타입·상수·노멀라이저·기본값
- `.claude/hooks/moai/sync-phase-quality-gate.sh` + `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` — WCI_EXCLUDES(Facet 3 표면)
- primary 체크아웃 `/Users/goos/MoAI/moai-adk-go/.moai/reports/t1395/gate-block-disposition{,-2..9}.md` — 처분 기록 9건(읽기 전용); 본 트리 사본 `.moai/reports/t1404/gate-block-evidence.md`
- 관련 SPEC: SPEC-CODEX-GATE-SCOPE-001(카드 스코핑 — §F.5) · SPEC-CODEX-REVIEW-OWNERSHIP-001(tree_scope 정책·REQ-CRO-006 패리티) · SPEC-CODEX-REVIEW-TARGET-001(REQ-CRT-006 트리 요청 형태) · SPEC-WORKTREE-STATE-ROOT-001(설정 루트 해상) · SPEC-MAIN-COMMIT-BAN-001(primary main 커밋사 mag)
