---
id: SPEC-JEV-GUARD-001
title: "Jev Consumer B withdrawal — restore the consumer-guard contract (SkillSuggest ships only after measurement)"
version: "0.1.0"
status: draft
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli + internal/jevmeasure + .claude/skills/moai"
lifecycle: spec-anchored
tags: "jev, consumer-guard, defect-repair, t1083, SkillSuggest"
depends_on: [SPEC-JEV-OPTIN-MEASURE-001, SPEC-JEV-CONSUMERS-001]
tier: S
related_specs: [SPEC-JEV-CONSUMERS-001, SPEC-JEV-OPTIN-MEASURE-001]
---

# SPEC-JEV-GUARD-001 — Jev Consumer B 철수 (consumer-guard 계약 복원)

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-09-22 | 0.1.0 | Initial plan-phase artifacts (card t1083, Class C). RED `TestNoConsumerCallPathShips` inherited from card t1066 cause chain; no re-diagnosis. | manager-spec |
| 2026-09-22 | 0.1.1 | plan-audit iter-1 FAIL (0.82, D1) — SKILL.md divergence baseline re-measured and corrected (§D.3: 21 changed lines / 4 categories, was misstated "17"); REQ-JEVG-006 modal rewritten (D4); AC-JEVG-004 made executable via normalized diff; AC-JEVG-008 added (D2); commit-guidance restoration pointer added to plan M1 (D5). | manager-spec |

## A. 배경 (Background)

카드 t1083 (Class C). `internal/jevmeasure/gate_demo_test.go:105` 의 `TestNoConsumerCallPathShips` (AC-JEVO-012의 강제 테스트, `SPEC-JEV-OPTIN-MEASURE-001` 소유)가 RED 이다:

- **Command**: `go test ./internal/jevmeasure/ -run TestNoConsumerCallPathShips`
- **Verbatim stdout** (RED-now, 관측 시점 이 트리):
  `gate_demo_test.go:137: a consumer call path is present before its measurement: [SkillSuggest in /…/internal/cli/jev_skill_suggest.go]`
- **Exit code**: 1
- **Tree SHA**: `cd99336bf` (branch `WT-jev-guard-green`, local develop head)

세 마커(`NearDuplicateMark`, `LaneQuestionRoute`, `SkillSuggest`) 중 SkillSuggest 하나만 적중한다 — Consumer A(C) 두 개는 부재를 유지 중이다.

**원인 사슬 (t1066에서 확정 — 재진단 금지)**: `SPEC-JEV-CONSUMERS-001` M6가 Consumer B (`SkillSuggest`)를 gate-unrun 상태(REQ-JEVN-016: 존재 허용, `workflow.jev.enabled=false` 게이트 뒤, Hidden 명령)로 출하했다. 그러나 측정 SPEC (`SPEC-JEV-OPTIN-MEASURE-001`)이 소유한 REQ-JEVO-009 / AC-JEVO-012는 비테스트 Go 전체에서 소비자 마커의 **리터럴 부재**를 단언하며, CI는 측정 측 계약을 강제한다. 두 SPEC 계약이 충돌했고, 측정 자체는 상류의 TypeSafe 와이어포맷 결함(t1066 F2, 별도 수리 카드)으로 현재 불가능하다. 따라서 소비자는 지금 측정될 수 없고, **가드 계약이 이긴다**: "consumers ship only after measurement."

## B. 요구사항 (GEARS)

**REQ-JEVG-001** (Unwanted) The shipped tree **shall not** contain a consumer call path for the SkillSuggest consumer **While** its measurement gate (REQ-JEVO-009 / AC-JEVO-012, owned by `SPEC-JEV-OPTIN-MEASURE-001`) has not produced a verdict that the consumer beats its constant-answer baseline. Verified by `TestNoConsumerCallPathShips` passing (exit 0) — AC-JEVG-001.

**REQ-JEVG-002** (Event-driven) **When** the withdrawal is applied, the CLI command registry **shall** register no `jev-suggest` command; invoking `moai jev-suggest` shall resolve as an unknown command. Verified by AC-JEVG-002 (grep absence) and AC-JEVG-005 (`go build ./...` ok).

**REQ-JEVG-003** (Ubiquitous) The guard test `internal/jevmeasure/gate_demo_test.go` **shall** remain byte-identical across this SPEC's entire lifecycle. No test suppression is permitted: no weakening, no re-scoping, no marker-list editing, no walk-scope relocation, no build tags, no skip logic. Verified by AC-JEVG-003 (byte-identity against tree `cd99336bf`).

**REQ-JEVG-004** (Ubiquitous) The shared Jev symbols living outside the withdrawal set — `jevNotice` (`internal/cli/todo_jev_finding.go`), `jevEnabled` (`internal/cli/doctor_jev.go`), `installJevProbe` (`internal/cli/todo_jev_finding_test.go`, Consumer C test helper) — **shall** remain present, and their owning packages' suites shall pass. Verified by AC-JEVG-006.

**REQ-JEVG-005** (Ubiquitous) The `/moai` skill document — both the local copy (`.claude/skills/moai/SKILL.md`) and the template source (`internal/template/templates/.claude/skills/moai/SKILL.md`) — **shall** carry no prose naming `moai jev-suggest`, **and** the two copies **shall not** widen their pre-existing divergence baseline (measured, see §D.3) as a result of this SPEC's edit. Verified by AC-JEVG-004.

**REQ-JEVG-006** (Event-driven — restoration contract) **When** the TypeSafe wire-format defect (t1066 F2, separate repair card) is repaired **and** the measurement gate runs **and** Consumer B beats its baseline, **then** a successor SPEC **shall** (a) restore the implementation, tests, `root.go` registration, and SKILL.md prose from git history, **and shall** (b) evolve the guard mechanism (`gate_demo_test.go` marker list) under the MEASURE SPEC's ownership before the consumer call path re-enters the tree. **Until** both acts are performed by that successor SPEC, the tree **shall not** contain the consumer call path (per REQ-JEVG-001). This SPEC authorizes neither the re-landing nor the guard edit. The contract exists because REQ-JEVN-016's "present-while-gate-unrun" state would trip the same guard again; a measured consumer that beats its baseline is a normative change to the guard's scope, owned by a SPEC — never by an ad-hoc test edit.

## C. 제약 (Constraints)

1. [HARD] `internal/jevmeasure/gate_demo_test.go` 및 가드를 단언하는 모든 `_test.go` — zero edits.
2. [HARD] 이 SPEC의 Go diff는 Consumer B (SkillSuggest) 패밀리로 한정된다. t1068 카드의 축(acbent absent-83-rows guard-AC family)과 무관 파일은 만지지 않는다.
3. [HARD] 전체 스위트(`go test ./...`) 로컬 실행 금지 — lane-local verification only. 전 판정은 CI 몫이다.
4. Template-First rule: 템플릿 사본 편집 후 `make build` 로 임베디드 FS 재생성.
5. 전체 스위트는 CI의 몫; 로컬은 영향 패키지만 (`./internal/jevmeasure/...`, `./internal/cli/...`).

## D. 범위 (Scope)

### D.1 철수 집합 (Withdrawal set — Consumer B family)

| # | Path | Action |
|---|------|--------|
| 1 | `internal/cli/jev_skill_suggest.go` | delete (~510 lines: impl + hidden `moai jev-suggest` command + `@MX:NOTE` gate-unrun anchor) |
| 2 | `internal/cli/jev_skill_suggest_test.go` | delete (its tests) |
| 3 | `internal/cli/root.go:218` | remove the one-line `rootCmd.AddCommand(newJevSuggestCmd())` registration |
| 4 | `.claude/skills/moai/SKILL.md` lines 115-123 | remove the `### Skill Suggestion (gated — default off)` subsection (heading + body + 3-item list naming `moai jev-suggest`) |
| 5 | `internal/template/templates/.claude/skills/moai/SKILL.md` lines 115-123 | same removal (keep the edited block identical across the two copies); then `make build` |

### D.2 유지 집합 (Stays — measured, not assumed)

- `jevNotice` — defined `internal/cli/todo_jev_finding.go:258`, used by Consumer A (`todo_jev_finding.go`) and Consumer B. After withdrawal, Consumer A's uses remain. STAYS.
- `jevEnabled` — defined `internal/cli/doctor_jev.go:126`, used by `doctor_jev.go`, `mcp_jev.go`, `todo_jev_finding.go`. STAYS.
- `installJevProbe` — defined only in `internal/cli/todo_jev_finding_test.go:25` (Consumer C test helper; the mention inside `jev_skill_suggest.go:405` is a comment that disappears with the file). STAYS.
- Consumer A (`todo_jev_finding.go`) and Consumer C — untouched; their markers (`NearDuplicateMark`, `LaneQuestionRoute`) already hold absence under the guard.

Withdrawal set 확장 없음 — shared-symbol 의존으로 강제되는 확장은 관측되지 않았다.

### D.3 SKILL.md 두 사본의 사전 분기 기준선 (measured baseline — corrected per plan-audit D1)

두 사본은 이 SPEC 이전부터 전체 파일 기준 바이트 동일이 **아니었다**. 본 워크트리의 커밋된 상태에 대해 plan-audit iter-1 이후 **재측정한** 값 (관측 시점 트리 `cd99336bf`):

- **Command**: `diff .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md`
- **Observed**: 20 hunks, 80 plain output lines, **41 diff content lines** (`<` 21 + `>` 20), **21 changed lines, 4 categories**:

| Category | Lines | Content |
|---|---|---|
| A | 17 (단일 hunk 17개) | `For detailed orchestration:` 줄 — 로컬 `${CLAUDE_SKILL_DIR}/workflows/*` vs 템플릿 리터럴 `.claude/skills/moai/workflows/*` |
| B | 2 (hunk `288,289c288,289`) | harness-Builder 블록 — Builder 단락 + 그 orchestration 줄, 역시 `${CLAUDE_SKILL_DIR}` 변형 |
| C | 1 (`335c335`) | 로컬 `moai cg -w <이름> --spawn` vs 템플릿 `moai cc -w <이름> --spawn ...` |
| D | 1 (`407d406`) | 로컬에만 존재하는 `Last Updated: 2026-07-07` |

> iter-1 산출물이 이 숫자를 "정확히 17줄"로 기술했던 것은 A 카테고리만을 전체로 오독한 것이었다(내부 산술도 17 vs 16+2로 상호 불일치). 재측정 결과는 plan-auditor의 21/41 판정과 정확히 일치한다 — 감산·보정이 아니라 관측의 교정이다.

따라서 정합성 기준은 전체 파일 바이트 동일이 아니라, **실행 가능한 형태**로: (i) 양쪽에서 `jev-suggest` 산문이 0이 되고, (ii) 정규화 비교(카테고리 A·B를 흡수하는 `${CLAUDE_SKILL_DIR}` → 리터럴 치환을 **비교 시점에만** 적용) 결과, 남는 분기가 C·D 두 기준 hunk 모양뿐임을 단언한다 — 새 hunk 하나라도 생기면 widening = FAIL (AC-JEVG-004). 이 치환은 비교기의 정규화일 뿐이며, 파일 자체를 고쳐 분기를 없애는 것(정규화 유혹)은 plan §G 가 금지하는 축 위반이다.

### D.4 폐기된 대안 (Rejected alternatives — recorded)

- **비순회/미빌드 위치로 코드 이동 (parking)** — 가드의 워크(`internal/` walk)를 우회하는 것은 이동에 의한 억제(suppression by relocation)다. 기각. `[HARD] no relocation games` 위반.
- **테스트를 gate-unrun 불변식으로 재범위화 (re-scoping)** — 가드를 약화시키는 억제다. 기각. `[HARD] no test suppression` 위반.
- **마커 이름 변경 (renaming the marker)** — 마커 게임. 가드의 탐지 대상만 옮긴다. 기각.
- **게이트 플래그 기본값 강제 (flip `workflow.jev.enabled`)** — 존재-게이트 상태는 유지하면서 CI 적색만 피하려는 시도; 가드는 존재 자체를 단언하므로 해결이 안 되고, 기본값 변경은 출하 계약 위반이다. 기각.

**선택된 처분**: Consumer B 패밀리를 트리에서 **삭제**한다. 작업물은 git history와 `SPEC-JEV-CONSUMERS-001` 의 커밋으로 보존되며, 복원 계약은 REQ-JEVG-006 에 기록된다.

### D.5 CONSUMERS SPEC 정정 여부 — 결정: (a) 미정정

`SPEC-JEV-CONSUMERS-001` (status: completed, v0.3.0) 은 **수정하지 않는다**. 근거: REQ-JEVN-016은 규범적으로 참인 문장("MAY be present **only while** 네 조건")이며, 철수는 그 전제 상태(gate-unrun)를 없앨 뿐 명령문과 충돌을 만들지 않는다 — 조건이 무너졌으므로 소비자가 존재하면 안 된다는 것은 REQ 자체가 이미 말하는 것이다. t1066의 CORE 정정 선례는 **본문 충돌이라는 규범적 결함**을 고친 경우로, 본 건(규범 변화 없는 상태 변화)과 다르다. 정정은 `completed → in-progress (amendment)` 전이 + HISTORY 절 + plan-audit 캐시 무효화라는 비용을 치르는 데 비해 규범적 이득이 없다. 추적 사슬은 본 SPEC (§F)이 단일 장소에서 운반한다.

## E. 제외 (Exclusions)

### Out of Scope — 다른 카드의 축

- t1068 카드가 소유하는 acbent absent-83-rows guard-AC 계열 — 본 SPEC이 만지지 않는다.
- t1066 F2 TypeSafe 와이어포맷 결함 수리 (endpoint 422: `questions` 사전형 + 질문별 `type` 판별자) — 별도 수리 카드. 본 SPEC의 복원 계약(REQ-JEVG-006)은 그 카드에 의존할 뿐 수행하지 않는다.
- Consumer A (`todo_jev_finding.go`)와 Consumer C — 측정 상태와 무관하게 현상 유지.

### Out of Scope — run-phase가 만지지 않는 문서 표면

- `.moai/project/codemaps/` 4개 파일의 `jev-suggest` 언급(`docs-truth.md:112`, `entry-points.md:7,63,70`, `data-flow.md:507`, `modules.md:59`) — 철수 후 stale 이 되는 **생성 문서**로, sync-phase에서 manager-docs가 재측정·재생성한다(`moai codemaps`). run-phase가 손대지 않는다 (축 분리).
- `.moai/specs/SPEC-JEV-CONSUMERS-001/progress.md` 의 `jev-suggest` 언급 — 역사 기록. 편집 금지.
- `moai-mcp-tools.md` 등 규칙 문서의 Jev 서술 — `jev_ask` MCP 도구는 본 건과 무관하며 Consumer B 철수의 대상이 아니다.

### Out of Scope — 가드 기구 변경

- `gate_demo_test.go` 의 마커 목록·워크 범위·긍정 대조 변경 — REQ-JEVG-006 이 두는 후속 SPEC 소관. 이 카드는 테스트 파일에 zero edits.

## F. 추적 사슬 (Traceability chain)

```
TestNoConsumerCallPathShips (internal/jevmeasure/gate_demo_test.go:105)
  ← AC-JEVO-012 / REQ-JEVO-009  (SPEC-JEV-OPTIN-MEASURE-001 — guard owner)
  ↔ REQ-JEVN-016 / AC-JEVN-016  (SPEC-JEV-CONSUMERS-001 — consumer owner, gate-unrun 상태 규정)
  ← REQ-JEVG-001..006           (SPEC-JEV-GUARD-001 — 본 SPEC: 위반 상태의 해소 + 복원 계약)
  → card t1083 (Class C) / t1066 F2 (상류 차단, 별도 카드) / t1068 (별개 축)
```

## G. @MX 태그 보고 (계획)

철수로 제거될 태그 (run-phase에서 실측):

- `@MX:NOTE` 1건 — `jev_skill_suggest.go:190` ("gate-unrun Consumer B anchor (SPEC-JEV-CONSUMERS-001 REQ-JEVN-016)") — 파일 삭제와 함께 소멸 (mx-tag-protocol: NOTE는 코드 삭제 시 제거).

추가·갱신 태그 없음. `root.go` 등록 한 줄 제거는 태그를 실지 않는다. 상세는 run-phase 보고의 `## @MX Tag Report` 참조.

## H. 교차 참조

- `SPEC-JEV-OPTIN-MEASURE-001` — 가드 소유자 (REQ-JEVO-009, AC-JEVO-012, `TestNoConsumerCallPathShips`)
- `SPEC-JEV-CONSUMERS-001` — 소비자 소유자 (REQ-JEVN-015/016, M6 = Consumer B 출하 커밋)
- `.moai/specs/SPEC-JEV-GUARD-001/research.md` — 관측·검증 근거 전문
- CLAUDE.local.md §2 (Template-First, `make build`), §4.1 (lane 의무: 완료 후 리드에게 로컬 develop 병합 요청)
