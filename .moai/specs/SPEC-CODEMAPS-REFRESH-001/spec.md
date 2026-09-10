---
id: SPEC-CODEMAPS-REFRESH-001
title: "Codemaps freshness closure: regeneration, accuracy verification, and provenance restamp"
version: "0.1.1"
status: completed
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
lifecycle: spec-anchored
tags: "codemaps, freshness, regeneration, accuracy-verification, provenance, stamp, graph-check"
era: V3R6
tier: M
related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002, SPEC-STAMP-REACHABILITY-001, SPEC-V3R6-DOCS-CODEMAPS-V3-001, SPEC-GRAPH-FRESHNESS-CADENCE-001]
---

# SPEC-CODEMAPS-REFRESH-001 — Codemaps 최신성 종결: 재생성 · 정확성 검증 · provenance 재스탬프

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-02 | Initial plan-phase authoring (card t432, queue intent: "codemaps 최신성 게이트 상시 적색 — 재생성 + 정확성 검증"). Baseline measured in worktree t432 @ ad272be20. | manager-spec |
| 0.1.1 | 2026-09-02 | plan-audit iter-1(FAIL 0.81) 수리 — D1~D6 전부 적용: D1 scope-hygiene 허용 경로에 본 SPEC 디렉터리 추가 + 측정 창을 base ad272be20 대비 합집합으로 고정(공허 녹색 차단), D2 AC-CMR-008(증거 독립성) 신설, D3 REQ-CMR-004 검증 경계 정량화 + AC-CMR-002 dedup 정의, D4 SHA 인용 오탈자, D5 lifecycle을 SSOT enum(spec-anchored)으로 정정, D6 progress.md 표기 수정. | manager-spec |

## §A. Problem Statement

`moai graph check`의 codemaps 계층이 상시 적색이다. 2026-09-02 본 워크트리(ad272be20)에서 재측정한 기준선:

```
codemaps  metric=described-source-diff value=60 threshold=40 verdict=stale
  contribution: 13 described-worthy file(s) vs first parent f3e11e113 (t413 merge)
```

원인은 단순하다. provenance 스탬프(`.moai/project/codemaps/provenance.json`, commit_sha `7fc0af324cf4...`, generated_at 2026-08-31T06:57:41Z, tree_root = t287 워크트리)가 가리키는 커밋 이후 described roots(internal, cmd, pkg)에서 60개 파일이 변경됐고, 그중 t413 머지 하나가 13개 기술 대상 파일을 직접 기여했다. 재생성·재스탬프가 없는 한 이 값은 계속 증가한다 — 카드 의도 그대로 "상시 적색".

그러나 이 SPEC의 중심은 게이트를 녹색으로 만드는 것이 아니다. **녹색 게이트 + 틀린 내용이야말로 이 카드가 금지하는 실패 형태다.** 스탬프를 갱신만 하면 `moai graph check`는 fresh를 보고하지만, 그것은 "문서가 최근 커밋과 일치하는 시점에 생성됐다"는 서술일 뿐 문서 내용이 실제 트리를 올바르게 기술한다는 뜻이 아니다. 실측된 내용 부정확의 예:

- 현재 codemaps 6문서는 존재하지 않는 6개 패키지(internal/design, internal/evaluator, internal/factory, internal/migrate, internal/research, internal/state — 전부 트리 부재 실측)를 언급한다. 일부는 "존재하지 않음" 경고 노트로 덮여 있고 일부(`modules.md` `### internal/factory` 섹션)는 서술로 남아 있다. 이 내용 수정은 카드 t304의 소관이다.
- `docs-truth.md` §1이 "정확히 11개 retained agents"라고 기술하는데, 현재 카탈로그는 12개(CLAUDE.md §4)다 — 생성 시점 이후 실제 트리가 움직인 단면.

한편 재생성 파이프라인의 기존 검증(`.claude/skills/moai/workflows/codemaps.md` Phase 4)은 "참조된 파일/모듈이 실제 존재하는지 확인"이라는 프로세스 서술일 뿐, 증거 산출물 수출 의무가 없다. 게이트는 통과하는데 정확성 판독이 어디에도 남지 않는 상태가 오늘의 구조적 빈틈이다.

## §B. Scope

### §B.1 In Scope

실행 순서 M1 → M4 (상세는 plan.md §F):

- **M1 재생성** — `/moai codemaps --force`로 6문서(overview, modules, dependencies, data-flow, entry-points, docs-truth)를 현재 트리 기준으로 재생성.
- **M2 정확성 검증** — 재생성 결과를 실제 트리와 대조하는 3개 검증 항목(경로 실존 / 패키지 구조 vs `go list ./...` / 인용 식별자 실존)을 증거 파일로 수출. 카드의 중심 요건.
- **M3 provenance 재스탬프** — `moai graph stamp codemaps --commit <merge-base>` 로 도달성 안전 재스탬프.
- **M4 게이트 종결** — `moai graph check` 전 계층 non-stale 확인 + 런 리포트.

### §B.2 Boundary decisions

- 이 SPEC은 **문서 재생성 + 검증** SPEC이다. Go 프로덕션 코드 변경은 없다(REQ-CMR-007).
- 정확성 검증은 기존 도구(grep, `go list ./...`, `test -d/-f`, `git merge-base`)로 수행 가능하며 신규 툴링을 요구하지 않는다 — 신규 툴링이 필요해지는 순간 실행을 중단하고 범위 질문으로 반환한다(REQ-CMR-002~004의 전제).
- 스탬프는 콘텐츠 해시가 아니라 커밋 SHA를 기록하므로, 재스탬프는 반드시 병합 생존 리비전을 명명해야 한다(REQ-CMR-005 — SPEC-STAMP-REACHABILITY-001이 닫은 고아화 결함 계열의 재발 방지).

### Out of Scope — t304 소관: 팬텀 패키지 6개의 내용 수정

- internal/design, internal/evaluator, internal/factory, internal/migrate, internal/research, internal/state 에 대한 codemaps 서술(경고 노트 포함)의 삭제·수정은 카드 t304(대기 중, 리드가 순서 소관)가 담당한다.
- 재생성이 이 6개를 다시 유도하면 본 SPEC의 실행자는 발견을 런 리포트의 t304 인계 섹션에 **기록만** 하고 인용 본문을 고치지 않는다. 이들의 존재는 본 SPEC의 합격 판정에 영향을 주지 않는다.
- known-6 외에 재생성이 **새로** 만들어낸 부정확 인용은 new-findings로 분류해 리포트에 남긴다(수정도 t304 후속으로 넘긴다).

### Out of Scope — 게이트 임계값 재보정

- `gate.yaml` `graph_freshness.codemaps_changed_files: 40`의 값 자체에 대한 재보정은 카드 텍스트상 명시적 제외다.
- 작업 중 임계값 적절성에 대한 증거(예: 정상 재생성 주기에도 value가 임계값을 넘는 관측)가 뜨면 리포트에 보고만 한다. 설정 파일을 만지지 않는다.

### Out of Scope — Go 프로덕션 코드 및 신규 툴링

- internal/, pkg/, cmd/ 의 Go 코드 변경은 없다. `moai graph`/`moai spec` CLI 자체의 수정도 소관 밖이다.
- 정확성 검증을 위해 새 Go 코드·새 서브커맨드를 만들지 않는다. 검증은 셸 명령 + 기존 CLI로 수행하고 그 증거를 기록한다.

## §C. Requirements (GEARS)

- REQ-CMR-001 (Ubiquitous) — **재생성 완전성.** `/moai codemaps --force` 실행 시 the regeneration shall produce the complete codemaps set — `overview.md`, `modules.md`, `dependencies.md`, `data-flow.md`, `entry-points.md`, `docs-truth.md` — from the current tree state under described_roots `[internal, cmd, pkg]`.

- REQ-CMR-002 (accuracy 항목 a) — **인용 경로 실존.** **While** the regenerated documents cite paths under `(internal|pkg|cmd)/`, the accuracy verification shall test every cited path against the working tree and export a per-path result table (path → exists/absent) to the evidence file. **When** a cited path is absent, the executor shall record the finding in the 리포트의 t304 인계 섹션(known-6) 또는 new-findings 섹션(그 외) — 인용 본문은 수정하지 않는다.

- REQ-CMR-003 (accuracy 항목 b) — **패키지 구조 대조.** The accuracy verification shall compare the regenerated package/dependency enumeration (modules.md, dependencies.md) against `go list ./...` output and record every discrepancy(누락된 실재 패키지 / 유령 패키지) with verbatim evidence in the 리포트.

- REQ-CMR-004 (accuracy 항목 c) — **인용 식별자 실존.** The accuracy verification shall resolve every entry-point and data-flow identifier cited in the regenerated documents (예: `cli.Execute`, `InitDependencies`, `EmbeddedTemplates`, `Deployer.Deploy`) against its named file/package via grep and export a hit/miss table. `docs-truth.md`의 검증 경계는 §1 에이전트 카탈로그 테이블 전수다 — 카탈로그의 모든 행을 `.claude/agents/` 트리 나열과 대조하며 표본 추출은 없다. 미적중은 기록만 한다.

- REQ-CMR-005 — **스탬프 도달성.** **While** the working branch is not the integration branch, the restamp shall name a merge-surviving revision via `moai graph stamp codemaps --commit "$(git merge-base HEAD origin/develop)"`; the executor shall not stamp the branch-local HEAD. (SPEC-STAMP-REACHABILITY-001 REQ-GFR-014 계승 — squash 머지 시 스탬프 고아화 방지.)

- REQ-CMR-006 — **게이트 종결.** **When** `moai graph check` runs after 재생성과 재스탬프, the codemaps layer shall report verdict=fresh (metric value < 40, expected 0), and no other layer shall report verdict=stale. 신규 worktree에서 mx-index/edges의 verdict=absent는 예상 상태(untracked runtime artifact)로 stale이 아니다.

- REQ-CMR-007 (Unwanted) — **범위 금지.** The executor shall not modify Go production code (`internal/`, `pkg/`, `cmd/`), the `gate.yaml` freshness threshold values, 또는 t304 소관인 팬텀 패키지 인용 본문.

- REQ-CMR-008 — **권고 게이트의 증거 독립성.** **Where** the graph freshness gate runs advisory (`gate.yaml` `graph_freshness.blocking: false`), the run shall not substitute a green gate verdict for the accuracy evidence; REQ-CMR-002~004 shall close with exported evidence regardless of the gate verdict.

## §D. Acceptance Criteria

전체 AC 열거와 Given-When-Then 시나리오는 `acceptance.md` §D AC Matrix에 있다. 요약:

| AC | 내용 | MUST |
|----|------|------|
| AC-CMR-001 | 재생성 완전성 — 6문서 전부 재생성 | ✓ |
| AC-CMR-002 | 경로 실존 표 수출 (accuracy a) | ✓ |
| AC-CMR-003 | `go list ./...` 대조 기록 (accuracy b) | ✓ |
| AC-CMR-004 | 식별자 hit/miss 표 수출 (accuracy c) | ✓ |
| AC-CMR-005 | 스탬프 도달성 (origin/develop 조상) | ✓ |
| AC-CMR-006 | 게이트 fresh (codemaps < 40, 타 계층 stale 없음) | ✓ |
| AC-CMR-007 | 범위 위생 — 변경 집합이 codemaps+리포트+본 SPEC 디렉터리에 한정 (base ad272be20 대비) | ✓ |
| AC-CMR-008 | 증거 독립성 — 정확성 증거와 게이트 판정이 별도 근거로 병존 | ✓ |

## §E. Cross-References

- SPEC-V3R6-GRAPH-FRESHNESS-001 — 신선도 게이트(`moai graph check`), 3계층 모델, 콘텐츠 어드레싱 인용을 건 SPEC.
- SPEC-V3R6-GRAPH-FRESHNESS-002 — M0 비상 재스탬프 + REQ-GFR-014("branch-local HEAD 재스탬프 금지") 원천.
- SPEC-STAMP-REACHABILITY-001 — 스탬프 고아화 CI 가드 + `--commit` 명시 모드. REQ-CMR-005의 직접 선행.
- SPEC-V3R6-DOCS-CODEMAPS-V3-001 — codemaps SSOT 문서 집합 원천 생성.
- SPEC-GRAPH-FRESHNESS-CADENCE-001 — 신선도 주기 운영.
- 카드 t304 — 팬텀 패키지 내용 수정 소관(대기 중).
