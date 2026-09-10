---
id: SPEC-GIT-DELIVERY-PROCEDURE-001
title: "배포 지침의 git 전달 절차 기계적 수리 — fetch 순서, 병합 방식 해석, auto-merge 기본값 단일 기준"
version: "0.2.0"
status: draft
created: 2026-09-10
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/agents/moai/manager-git.md, .claude/rules/moai/core/agent-common-protocol.md, .claude/skills/moai/workflows/sync/delivery.md, .claude/skills/moai/workflows/sync/doc-execution.md (+ internal/template/templates mirrors, internal/template/templates/.codex/agents/moai/manager-git.toml)"
lifecycle: spec-anchored
tags: "git, manager-git, sync-check, fetch-ordering, merge-method, auto-merge, template-mirror, instruction-audit"
era: V3R6
tier: M
related_specs: [SPEC-MERGE-METHOD-CONFIG-001]
---

# SPEC-GIT-DELIVERY-PROCEDURE-001 — 배포 지침의 git 전달 절차 기계적 수리

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-10 | manager-spec | 최초 작성. 칸반 카드 t622(지침 감사 G1). 결함 3건(AC-01·AC-11·SX-R04)을 한 SPEC으로 묶었다. 재현 근거 `.moai/reports/t622/repro.md`(측정 트리 `c352330d3`), 인용 줄번호는 `b412f8a33` 에서 확인. |
| 0.1.1 | 2026-09-10 | manager-spec | **plan-audit 1회차**(`.moai/reports/t622/plan-audit.md`) **FAIL 0.67** 반영 — 전체 범위(AC-01 포함) 대상. |
| 0.1.2 | 2026-09-10 | manager-spec | **plan-audit 2회차**(`.moai/reports/t622/plan-audit-iter2.md`) **FAIL 0.75** 반영(결함 N1~N6) — 전체 범위 대상. 운영자 결정 OD-1 = 선택지 1, OD-2 = 선택지 B 반영. |
| 0.1.3 | 2026-09-10 | manager-spec | 운영자 결정 B1~B3 반영(단일 PR, primary feature 브랜치 경로 이동, 헌법 개정 경로). REQ 23·AC 24로 Tier M 상한 초과, amend 적용 단계 스텁(B4)·worktree-integration.md 모순(B6) 기록. 커밋 `87988e946`. 이 판에 대한 plan-audit은 실행되지 않았다. |
| 0.2.0 | 2026-09-11 | manager-spec | 운영자 결정 T1 = 분할(리드 경유). 이 SPEC은 기계적 수리 세 가지(AC-11 fetch 순서, SX-R04 병합 방식 해석, OD-2 = B auto-merge 기본값 단일 기준)와 그 파일들의 부수 의무만 남긴다. late-branch 재설계(AC-01 전체, OD-1·B1·B2·B3·B6, Step 3.3.5 퇴역)는 카드 **t658**로, amend 적용 단계 스텁(B4)은 카드 **t659**로 옮겼다(§G). 옮긴 요구사항·수용 기준은 번호를 유지한 한 줄 자리표시로 §B·acceptance.md에 남기고 §G 표로 추적한다. Frozen 비접촉 확인을 §A.5에 기록하고 REQ-GDP-024·AC-GDP-025로 run-phase에서도 지킨다. **plan-audit 회차는 이 판부터 1회차로 다시 센다** — 위 두 회차는 전체 범위에 대한 기록이다. |

---

## §A 배경과 문제

### A.1 결함 요약

| 결함 | 내용 | 처리 |
|---|---|---|
| AC-11 | `git fetch` 와 그 결과를 읽는 `git rev-list` 를 서로 독립인 병렬 배치로 지시한다 | 기계적 수리 (이 SPEC) |
| SX-R04 (병합 방식) | `delivery.md` 와 `manager-git.md` 예시가 `gh pr merge --squash` 를 고정한다 | 기계적 수리 (이 SPEC) |
| SX-R04 (기본값) | 워크트리 문맥의 auto-merge 기본값이 `delivery.md`·`doc-execution.md` 와 `manager-git.md` 에서 반대로 적혀 있다 | OD-2 = 선택지 B (이 SPEC) |
| AC-01 | 배포 지침이 스스로 금지한 primary checkout 브랜치 변경을 실행하라고 안내한다 | 카드 t658로 이동 (§G) |

인용 줄번호는 `b412f8a33` 기준이다. 0.2.0 작성 시점 HEAD `87988e946` 에서 범위 파일 네 개(로컬·템플릿), 생성물 `manager-git.toml`, 관련 테스트 파일, `Makefile` 은 `b412f8a33` 과 차이가 없다(`git diff --stat` 출력 없음, exit 0).

### A.2 AC-11 — 결과를 읽는 명령과의 병렬 배치

| 파일 | 줄 | 문제 |
|---|---|---|
| `.claude/agents/moai/manager-git.md` (생성물 `manager-git.toml:150`) | 156 | `git fetch`, `git status`, `git rev-list --count --left-right`, `gh pr checks --json` 를 "independent and read-only" 로 묶어 한 턴 병렬 배치로 지시 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 292, 296, 299 | Pre-Spawn Sync Check 가 "parallel batch" 를 선언하고 `git fetch origin main`(296)과 `git rev-list`(299)를 코드 블록 안 별도 명령으로 나열 |
| 같은 파일 (정상) | 347 | Pre-Edit Sync Check 는 `git fetch origin main 2>&1; git rev-list ...` 를 한 명령으로 묶어 순서가 보장됨 — 유지 대상 |

`git fetch` 는 원격 추적 ref 를 갱신하고 `git rev-list` 는 그 ref 를 읽으므로 둘은 독립이 아니다. 병렬로 돌면 원격이 앞선 상태(`N 0`)가 `0 0` 으로 보고될 수 있다.

### A.3 SX-R04 — 병합 방식 고정과 auto-merge 기본값 충돌

| 파일 | 줄 | 내용 |
|---|---|---|
| `delivery.md` | 335-338 | 트리거: `is_worktree_context == true` 이고 `--no-merge` 가 없으면 병합 (기본 병합) |
| `delivery.md` | 343, 355 | `gh pr merge --squash --delete-branch` 고정 |
| `delivery.md` | 348-349 | `--no-merge` 설명, `--merge` 폐기 경고 "Auto-merge is now the default for worktree contexts." |
| `sync/doc-execution.md` | 34-36 | "This affects auto-merge behavior: worktree contexts default to auto-merge." |
| `manager-git.md` (생성물 `.toml:108`) | 114 | 예시 `gh pr merge <PR> --squash --delete-branch   # squash default` |
| `manager-git.md` (생성물 `.toml:26`) | 32 | `merge_method` 해석 규칙의 기본값 설명 — **유지 대상** |
| `manager-git.md` (생성물 `.toml:142`, `:160`) | 148, 166 | Team Mode "Auto-merge: only with the `--auto-merge` flag", "Execute only with `--auto-merge` flag AND all approvals obtained" — **유지 대상** |

team 모드이면서 워크트리 문맥이면 두 지침이 반대를 지시한다. 병합 방식 치환은 SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-007/008이 이미 요구한 것이 반영되지 않은 상태다. 워크트리 기본 병합 문구는 커밋 `b7447cb90`(`sync.md`)에서 들어와 `980ccdc56` 이 `delivery.md` 로 옮겼다(`git log -S`, `delivery.md:349` 문장).

### A.4 설정·사본·테스트 사실 (측정)

| 항목 | 측정 결과 |
|---|---|
| `merge_method` | 템플릿 `git-strategy.yaml.tmpl` 세 모드 모두 `squash` · Go 기본값 `internal/config/defaults.go` 738·751·767행 `MergeMethod: "squash"` |
| 폐기 키 `workflow.worktree.auto_merge` | `internal/config/testdata/shipped_key_inventory.yaml:2940-2943` 에 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"`. 이 SPEC은 건드리지 않는다 |
| 로컬·템플릿 사본 | `manager-git.md`·`agent-common-protocol.md` diff exit 0(바이트 동일). `delivery.md` diff exit 1 — 차이 `275c275`, `278c278`, `479,480c479`. `doc-execution.md` diff exit 1 — 차이 `138,143d137` |
| 바이트 동일 미러 테스트 | `internal/template/rule_template_mirror_test.go` 의 `workflowOptMirroredPaths`·`lateBranchMirroredPaths` 는 범위 파일 네 개 중 어느 것도 담지 않는다. `agent-common-protocol.md` 와 `manager-git.md` 는 주석으로 "byte-parity 허용 목록에서 제거" 가 명시돼 있다 |
| 기타 사본 가드 | `sanitized_pair_parity_test.go:71` 의 `sanitizedPairPaths` 가 `agent-common-protocol.md` 를 담는다(`TestSanitizedPairParity`, 토큰 정규화 뒤 교리 일치). `internal_content_leak_test.go:1535` `TestTemplateNoInternalContentLeak` 는 템플릿 루트 전체를 걷는다(`delivery.md` 는 `SPEC-AUTH-001` 교육용 예시 허용 항목 `:714`). `manager-git.md`·`delivery.md`·`doc-execution.md` 의 사본 일치를 지키는 테스트는 없다 — diff만이 가드다 |
| 생성물 | 템플릿 `manager-git.md` 가 바뀌면 `internal/template/templates/.codex/agents/moai/manager-git.toml` 을 `make agents-emit` 으로 재생성하고 `make agents-emit-check` 로 확인한다 |

### A.5 Frozen 비접촉 확인 (0.2.0 작성 시점, 템플릿·로컬 사본)

| 확인 | 명령과 결과 | 판정 |
|---|---|---|
| 범위 파일의 `[ZONE:]` 태그 | `grep -n -E '\[ZONE:(Frozen|Evolvable)\]'` — `manager-git.md`·`delivery.md`·`doc-execution.md` 0줄. `agent-common-protocol.md` 는 `[ZONE:Frozen]` 1줄(17행, User Interaction Boundary)과 `[ZONE:Evolvable]` 16줄. 로컬 사본도 `[ZONE:Frozen]` 은 같은 17행 하나 | Pre-Spawn Sync Check 태그는 292행 `[ZONE:Evolvable]` |
| 레지스트리 등록 절 | `zone-registry.md` 에서 `file:` 이 `manager-git.md`·`delivery.md`·`doc-execution.md` 인 항목 0개. `agent-common-protocol.md` 항목 중 Frozen은 `CONST-V3R2-006`(13행 "`AskUserQuestion` is the **only** user-facing question channel"), `CONST-V3R2-036`·`038`(17행), `CONST-V3R2-037`(52행 "Preload `AskUserQuestion` via `ToolSearch(query:"`) — 모두 `#user-interaction-boundary`. 나머지는 Evolvable이며 Pre-Spawn·Pre-Edit 절을 가리키는 항목은 없다 | 등록 절은 11~110행 구간에만 있다 |
| 유지 편집 자리 | `agent-common-protocol.md` 290-305(Pre-Spawn), `manager-git.md` 114·156, `delivery.md` 335-338·343·348-349·355, `doc-execution.md` 34-36 | 어느 자리도 `[ZONE:Frozen]` 블록이나 등록 절 안에 있지 않다 |

---

## §B 요구사항 (GEARS)

요구사항 식별자는 `REQ-GDP-NNN`. **번호 방식**: 0.1.3의 번호를 그대로 유지한다. 카드 t658로 옮긴 번호와 이전에 철회된 번호는 원래 자리에 한 줄 자리표시로 남기고, 새 요구사항은 다음 번호(024)를 받는다. 이유는 §C.2.

### B.1 AC-11 — fetch 순서 보장

**REQ-GDP-001** (Ubiquitous)
`manager-git` 에이전트 정의의 동기화 절은 `git fetch` 가 끝난 뒤에 그 결과를 읽는 `git rev-list --count --left-right` 가 실행되도록 순서를 지시해야 하며, `git fetch` 를 그 결과를 읽는 명령과 서로 독립인 병렬 배치 항목으로 분류해서는 안 된다. fetch 결과를 읽지 않는 명령(`git status`, `gh pr checks --json`)을 병렬로 두는 일은 이 요구사항이 제한하지 않는다.

**REQ-GDP-002** (Ubiquitous)
`agent-common-protocol.md` 의 Pre-Spawn Sync Check 코드 블록은 `git fetch origin main` 과 `git rev-list --count --left-right origin/main...HEAD` 를 순서가 보장되는 한 명령으로 지시해야 한다. 세 번째 명령(`moai session list --json --filter-spec=<SPEC-ID>`)의 병렬 실행 허용과 두 해석 표의 내용은 바뀌어서는 안 된다.

**REQ-GDP-003** (Unwanted)
`agent-common-protocol.md` 의 Pre-Edit Sync Check 절(328행 제목부터 `#### The sweep prohibition` 소절 앞까지)은 이 SPEC의 변경으로 수정되어서는 안 된다.

### B.2 SX-R04 — 병합 방식 해석

**REQ-GDP-004** (Ubiquitous)
`sync/delivery.md` 의 auto-merge 실행 단계는 병합 명령을 활성 모드의 `git_strategy.<mode>.merge_method`(기본 `squash`)로 해석한 `gh pr merge --<merge_method> --delete-branch` 로 지시해야 한다. 기본값에서 실제 실행 명령은 지금과 바이트 동일해야 한다(SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-009 연속성).

**REQ-GDP-005** (Unwanted)
범위 파일(`manager-git.md`, `delivery.md`, `agent-common-protocol.md`, `doc-execution.md`)과 생성물 `manager-git.toml` 의 어떤 실행 예시도 `--squash` 를 고정한 `gh pr merge` 명령을 실행할 명령으로 제시해서는 안 된다. `manager-git.md:114` 예시는 `--<merge_method>` 로 적혀야 하고, `manager-git.md:32` 의 기본값 설명 문장은 남아야 한다.

### B.3 SX-R04 — auto-merge 기본값 단일 기준 (OD-2 = B)

**REQ-GDP-006** (Event-driven)
When sync 단계가 PR 병합 여부를 판단할 때, `delivery.md` 의 Auto-Merge 절(335-338 트리거, 348-349 플래그 설명)과 `doc-execution.md` 의 워크트리 문맥 감지 절(34-36)은 `manager-git.md` 의 옵트인(`--auto-merge` 플래그와 전원 승인)을 유일한 기준으로 따라야 하고, 워크트리 문맥이라는 이유만으로 병합하는 기본값을 적어서는 안 되며, 기준이 `manager-git.md` 임을 이름으로 밝혀야 한다. `manager-git.md` 의 옵트인 문장(148행, 166행)은 남아야 한다.

### B.4 옮긴 번호와 철회된 번호 (자리표시)

아래 번호는 이 SPEC에서 요구사항으로 판정하지 않는다. 원문은 커밋 `87988e946` 의 spec.md에 있고, 추적 표는 §G.

**REQ-GDP-007** — **[MOVED → 카드 t658]** AC-01 명령 형태 실행 안내 불변식.
**REQ-GDP-008** — **[MOVED → 카드 t658]** late-branch 대체 절차 없는 예시 삭제 금지와 절차 참조 유지.
**REQ-GDP-009** — **[MOVED → 카드 t658]** 네 지침 절의 흐름 일치와 `feat/SPEC-` 단일 접두.
**REQ-GDP-010** — **[MOVED → 카드 t658]** 런처 워크트리 흐름, `manager-git.md` 42·160, Frozen 줄 수정 기록.
**REQ-GDP-011** — **[WITHDRAWN 2026-09-10]** OD-1 선택지 2(조건부 유지) 경로. 기록은 §G.
**REQ-GDP-012** — **[WITHDRAWN 2026-09-10]** OD-1 선택지 3(`main_late_branch` 은퇴) 경로. 기록은 §G.

### B.5 사본·생성물·중립성

**REQ-GDP-013** (Ubiquitous)
이 SPEC이 바꾼 줄은 로컬 사본(`.claude/…`)과 템플릿 사본(`internal/template/templates/.claude/…`)에서 같아야 한다. `delivery.md` 두 사본의 의도된 차이(275·278행, 479-480행)와 `doc-execution.md` 두 사본의 의도된 차이(로컬 사본에만 있는 138-143행)는 그대로 남아야 한다.

**REQ-GDP-014** (Ubiquitous)
템플릿 `manager-git.md` 가 바뀌면 `internal/template/templates/.codex/agents/moai/manager-git.toml` 은 `make agents-emit` 으로 재생성되어 같은 카드에 커밋되어야 하며, 손으로 편집해서는 안 되고, `make agents-emit-check` 가 exit 0 이어야 한다.

**REQ-GDP-015** (Unwanted)
`internal/template/templates/` 아래에 추가되는 문구는 SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 16개 지원 프로그래밍 언어 중 특정 언어로 치우친 표현을 담아서는 안 된다.

### B.6 옮긴 번호 (자리표시, 계속)

**REQ-GDP-016** — **[MOVED → 카드 t658]** Route B SPEC당 PR 하나(B1).
**REQ-GDP-017** — **[MOVED → 카드 t658]** Route B Step 1~3 같은 워크트리와 서술형 안내 금지(B2).
**REQ-GDP-018** — **[MOVED → 카드 t658]** `delivery.md` Step 3.2 github-flow 전달 경로(B2).
**REQ-GDP-019** — **[MOVED → 카드 t658]** `delivery.md` Step 3.3.5 퇴역과 `ExitWorktree`(B2).
**REQ-GDP-020** — **[MOVED → 카드 t658]** `CONST-V3R5-027`·`028` amend 기록(B3).
**REQ-GDP-021** — **[MOVED → 카드 t658]** HumanOversight 승인 대행 금지(B3).
**REQ-GDP-022** — **[MOVED → 카드 t658]** 트리 빌드로 잰 헌법 검증 drift 0(B3).
**REQ-GDP-023** — **[MOVED → 카드 t658]** `zone-registry.md` 사본 일치와 새 clause(B3).

### B.7 Frozen 비접촉

**REQ-GDP-024** (Unwanted)
이 SPEC의 변경은 범위 파일 네 개(로컬·템플릿)의 `[ZONE:Frozen]` 줄과 `zone-registry.md` 에 등록된 Frozen clause 문장을 바꾸거나 지워서는 안 된다.

---

## §C 결정 기록

### C.1 결정

| 결정 | 선택 | 날짜 | 결정자 | 이력 |
|---|---|---|---|---|
| OD-2 — 워크트리 문맥 auto-merge 기본값의 단일 기준 | **선택지 B: `manager-git.md` 옵트인(`--auto-merge` 와 전원 승인)** | 2026-09-10 | 운영자(리드 경유) | 처음에는 선택지 C(설정 키 `workflow.worktree.auto_merge`)가 선택됐다. 그 키가 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` 으로 기록돼 있다는 사실이 제시된 뒤 B로 바뀌었다. 이 SPEC은 그 키를 건드리지 않는다 |
| T1 — SPEC 분할 | **분할.** 이 SPEC은 AC-11, SX-R04 병합 방식, OD-2 = B와 그 파일들의 부수 의무(사본 일치·의도된 차이 보존, `.toml` 재생성, 템플릿 중립성, 항상 로드 규칙 마지막 순서)만 남긴다. Frozen 구역 문장은 건드리지 않는다. late-branch 재설계는 카드 t658, amend 적용 단계 스텁은 카드 t659(t658이 의존) | 2026-09-11 | 운영자(리드 경유) | 0.1.3에서 REQ 23·AC 24로 Tier M 상한을 넘은 뒤 결정 |

OD-1·B1·B2·B3의 결정 기록은 카드 t658의 입력(§G)으로 옮겼다. 결정은 Implementation Kickoff Approval을 대신하지 않는다.

### C.2 번호 방식 — 원래 번호 유지

plan-auditor MP-1은 "REQ numbers must be sequential (REQ-001, REQ-002, ... REQ-N) with no gaps, no duplicates, and consistent zero-padding. Even one gap or duplicate = FAIL." 로 판정한다(`.claude/agents/moai/plan-auditor.md:138`).

- **택한 방식**: 0.1.3 번호를 유지하고, 옮긴 번호(007-010, 016-023)와 철회된 번호(011, 012)를 §B의 원래 자리에 한 줄 자리표시로 남긴다. §B에 REQ-GDP-001부터 024까지 모든 번호가 정확히 한 번씩 나타나므로 빈 번호도 중복도 없다. 새 요구사항은 다음 번호 024를 받는다.
- **다시 매기지 않은 이유**: 다시 매기면 같은 토큰이 두 뜻을 갖는다. 예를 들어 새 REQ-GDP-007이 사본 일치를 뜻하는 동안, §G 추적 표와 t658의 입력인 커밋 `87988e946` 에서는 REQ-GDP-007이 AC-01 불변식을 뜻한다. 한 문서 안에서 같은 번호가 두 요구사항을 가리키면 MP-1의 중복 판정과 추적성이 함께 흐려진다.
- **수용 기준도 같은 방식**: AC-GDP-001~024를 유지하고 새 기준은 AC-GDP-025를 받는다(acceptance.md §D).

### C.3 OD-2 검토한 선택지 (결정 당시 자료)

| 선택지 | 영향 파일 (L·T = 로컬·템플릿) | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 | 뒤집거나 확장하는 선행 기록 |
|---|---|---|---|---|
| **A. `delivery.md` 워크트리 기본값** (미채택) | `manager-git.md` L·T 148·166 + 재생성 `.toml` | team 모드 워크트리 문맥에서 플래그 없이 병합 | 플래그로 결정 | `b7447cb90` 동작을 확정 |
| **B. `manager-git.md` 옵트인** (채택) | `delivery.md` L·T 335-338·348-349; `doc-execution.md` L·T 34-36 | 워크트리 문맥의 sync가 기본으로 병합하지 않는다. `--merge` 폐기 경고의 논리가 뒤집힌다 | 워크트리 문맥 동작 기본값 "병합" → "병합 안 함" | `b7447cb90` 동작을 뒤집음(SPEC 기원 없음) |
| **C. 설정 키 `workflow.worktree.auto_merge`** (처음 선택 후 B로 교체) | `delivery.md`, `doc-execution.md`, `manager-git.md` + 사용처 분류 파일 | 기본 병합이 멈춤 | 키 값 `false` 유지, 첫 소비자. 키는 `class: D`, `deprecate_after: "v3.1.0"` | 폐기 예정 키를 되살림 |

### C.4 티어 재분류

| 신호 | 측정값 | Tier S 기준 | Tier M 기준 |
|---|---|---|---|
| 판정 대상 요구사항 수 | 10 (001-006, 013-015, 024) | 8 이하 | 16 이하 |
| 판정 대상 수용 기준 수 | 11 (001-006, 013-016, 025) | 8 이하 | 16 이하 |
| 영향 파일 수 | 9 (범위 파일 네 개 × 로컬·템플릿 8 + 생성물 `manager-git.toml` 1) | 5개 미만 | 5~15개 |
| 예상 변경량 | 수십 줄 규모 | 300줄 미만 | 300~1000줄 |

요구사항·수용 기준 수(8 초과)와 파일 수(9)가 Tier S 기준을 넘고 Tier M 안에 있다. 변경량만 Tier S 쪽이다. **`tier: M` 을 유지한다.** 자리표시 번호(옮김 12·철회 2)는 판정 대상이 아니므로 상한 계산에 넣지 않았다.

---

## §D 제외 범위 (Out of Scope)

### Out of Scope — 카드 t658·t659로 옮긴 작업

- late-branch 재설계 전체: AC-01 명령·서술형·단계별 PR 안내, OD-1 선택지 1, B1(SPEC당 PR 하나), B2(primary feature 브랜치 경로 이동), B3(`moai constitution amend` 경로), B6(`worktree-integration.md`), `delivery.md` Step 3.2 github-flow와 Step 3.3.5, `spec-workflow.md`·`spec-assembly.md`·`zone-registry.md` 편집 → 카드 t658 (§G).
- `internal/constitution/pipeline.go` 의 amend 적용 도우미 스텁(B4) → 카드 t659 (t658이 의존).

### Out of Scope — 이 SPEC의 편집 대상이 아닌 것

- `spec-workflow.md`, `spec-assembly.md`, `zone-registry.md`, `worktree-integration.md` 는 이 SPEC의 편집 대상이 아니다.
- `manager-git.md` 의 Late-Branch Invocation Pattern 절에서 이 SPEC이 고치는 줄은 114행 병합 예시 하나뿐이다. 나머지 절차 줄(96·110·119-127·137)과 42·160행은 t658 소관이다.
- `delivery.md` Step 3.4 안의 356행 "Checkout target branch, fetch latest" 와 `manager-git.md:171` "Checkout main, pull, delete local branch" 는 브랜치 변경 안내 계열이라 t658 입력으로 넘긴다(§G.2).
- 설정 키 `workflow.worktree.auto_merge` 와 사용처 분류 파일, Go 코드, `git-strategy.yaml.tmpl` 값.
- `[ZONE:Frozen]` 줄과 등록된 Frozen clause 문장(REQ-GDP-024).

### Out of Scope — 금지 목록과 경고표

- 범위 파일 안의 금지 목록·경고표·인용 예시에 들어 있는 `git` 명령 문구는 이 SPEC이 판정하지 않는다.

---

## §E 미검증과 잔여 위험

### E.1 미검증 (Gaps)

- `git fetch`/`git rev-list` 경합을 실제로 일으키지 않았다. 감사 보고서도 AC-11 신뢰도를 medium으로 매겼다.
- `manager-git.md`·`delivery.md`·`doc-execution.md` 의 사본 일치를 지키는 테스트가 없다. 이 세 파일의 사본 일치는 이 SPEC의 diff 판정으로만 확인된다(§A.4).
- `TestSanitizedPairParity` 는 토큰 정규화 뒤의 줄 차이를 허용 오차 안에서 받아들이므로, `agent-common-protocol.md` 의 한 줄짜리 사본 차이는 이 테스트만으로는 드러나지 않을 수 있다. 바이트 일치는 diff로 따로 확인한다.
- `make agents-emit-check` 기준선(재생성 전 exit 0)은 plan 작성 시점에 실행하지 않았다(리드 제약: make 금지).
- 워크트리 기본 병합 문구의 `git log -S` 추적은 `delivery.md:349` 문장에만 수행했다.

### E.2 잔여 위험

- **t658과의 겹침.** `manager-git.md:114`(이 SPEC)는 t658이 다시 쓸 Late-Branch Invocation Pattern 절 안에 있다. t658은 이 SPEC이 병합된 뒤의 줄을 입력으로 받아야 하며, 순서가 뒤바뀌면 `--<merge_method>` 치환이 되돌려질 수 있다.
- OD-2 = B는 워크트리 문맥의 기본 병합을 없앤다. 지금 이 기본값에 기대는 사용자는 `--auto-merge` 를 넘기지 않으면 병합되지 않는 PR을 보게 된다.
- 검출식은 줄 단위다. AC-GDP-001·006은 읽기 단계로 틈을 좁히지만 오독 가능성이 남는다.
- 인용 줄번호는 범위 파일이 `b412f8a33` 과 바이트 동일하다는 전제에 기댄다(HEAD `87988e946` 에서 확인). run-phase 전에 develop을 흡수하면 다시 확인해야 한다.

---

## §F 수용 기준

`acceptance.md` 가 판정 대상 기준(AC-GDP-001~006, 013~016, 025)과 검증 명령, 양성 대조, 뮤턴트 재실행 결과, 옮긴 번호의 자리표시를 담는다.

---

## §G 카드 t658로 옮긴 항목 (Moved to card t658)

원문 출처: 커밋 `87988e946` 의 `.moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/` (spec.md·acceptance.md·plan.md·progress.md). 표의 번호는 그 판의 번호이며, 이 문서의 §B·acceptance.md에는 같은 번호가 자리표시로만 남는다.

### G.1 요구사항·수용 기준

| 원래 ID | 제목 | 한 줄 내용 | 출처 커밋 |
|---|---|---|---|
| REQ-GDP-007 | AC-01 불변식 | 명령 형태 안내 위치가 primary checkout 브랜치 변경을 지시하지 않음 | `87988e946` |
| REQ-GDP-008 | 대체 절차 없는 삭제 금지 | 예시만 지우지 않고 `manager-git.md` 절차 참조와 대상 제목 유지 | `87988e946` |
| REQ-GDP-009 | 흐름 일치·단일 접두 | 네 지침 절이 같은 진입·PR·종결 명령, PR 접두 `feat/SPEC-` 하나 | `87988e946` |
| REQ-GDP-010 | 런처 워크트리 흐름 | 워크트리 안 커밋, 맨손 `git worktree add` 금지, `manager-git.md` 42·160 처리, Frozen 줄 기록 | `87988e946` |
| REQ-GDP-011 | [WITHDRAWN] 조건부 유지 | OD-1 선택지 2 경로 — 0.1.2에서 철회 | `87988e946` |
| REQ-GDP-012 | [WITHDRAWN] `main_late_branch` 은퇴 | OD-1 선택지 3 경로 — 0.1.2에서 철회 | `87988e946` |
| REQ-GDP-016 | SPEC당 PR 하나 (B1) | 단계별 PR과 그 브랜치 금지, 동기화 커밋은 같은 SPEC PR 브랜치 | `87988e946` |
| REQ-GDP-017 | 같은 워크트리·서술형 안내 금지 (B2) | Route B Step 1~3 같은 워크트리, 서술형 primary 브랜치 안내 금지 | `87988e946` |
| REQ-GDP-018 | Step 3.2 github-flow (B2) | 워크트리 안 PR 브랜치만 전달, primary 비-main 브랜치에서 멈춤 | `87988e946` |
| REQ-GDP-019 | Step 3.3.5 (B2) | 기준 브랜치 복귀 대신 `ExitWorktree`, 병합 착지 뒤 폐기 | `87988e946` |
| REQ-GDP-020 | amend 기록 (B3) | `CONST-V3R5-027`·`028` 을 `--before`·`--after`·`--evidence` 로 개정, 증거 추적 | `87988e946` |
| REQ-GDP-021 | 승인 대행 금지 (B3) | 레인은 HumanOversight 승인을 대신하지 않고 멈춰 올림 | `87988e946` |
| REQ-GDP-022 | 트리 빌드 검증 (B3) | 트리 빌드 `constitution validate` ok·drift 0, 빌드 커밋 = HEAD | `87988e946` |
| REQ-GDP-023 | 레지스트리 일치 (B3) | `zone-registry.md` L·T 동일, 새 clause 1회·이전 clause 0회 | `87988e946` |
| AC-GDP-007 | 명령 검출식 분류 장부 | 금지 집합 검출식 적중 분류, primary 실행 안내 0 (기준 트리 20줄) | `87988e946` |
| AC-GDP-008 | 절차 참조 유지 | 파일별 형식 참조 하한·느슨한 개수 일치·제목 조회 | `87988e946` |
| AC-GDP-009 | 흐름 표지·접두 합집합 | 네 절 표지 매트릭스, 접두 합집합 = {feat/SPEC-} (기준 트리 {chore, feat, plan, sync}) | `87988e946` |
| AC-GDP-010 | 워크트리 흐름·Frozen 기록 | `git worktree add` 부재, §E.2 Frozen·Kickoff 기록 | `87988e946` |
| AC-GDP-011 | [철회] | OD-1 선택지 2 경로 | `87988e946` |
| AC-GDP-012 | [철회] | OD-1 선택지 3 경로 | `87988e946` |
| AC-GDP-017 | 단계별 PR 분류 장부 | 단계별 PR 검출식 적중 분류 (기준 트리 19줄) | `87988e946` |
| AC-GDP-018 | Step 3.2 github-flow | Feature 경로 0·런처 표지·primary checkout 멈춤 경로 | `87988e946` |
| AC-GDP-019 | Step 3.3.5 결과 | `ExitWorktree`·명령 0·"return to base branch" 0·병합 착지 표지 | `87988e946` |
| AC-GDP-020 | amend 명령 기록 | 근거 문서·명령 원문·dry-run·운영자 실행 출력 | `87988e946` |
| AC-GDP-021 | 서술형 분류 장부 | 서술형 검출식 적중 분류 (기준 트리 11줄) | `87988e946` |
| AC-GDP-022 | 트리 빌드 헌법 검증 | 빌드 출처 기록과 `status: ok`·`drift_count: 0` (짝 대조 측정 포함) | `87988e946` |
| AC-GDP-023 | 레지스트리 사본·새 clause | L·T diff 0, 새 clause 1·이전 clause 0 | `87988e946` |
| AC-GDP-024 | 승인 대행 금지 | 명령 기록에 표준입력 주입 없음, 운영자 `(Y/N)`·리드 에스컬레이션 기록 | `87988e946` |

### G.2 t658 입력 메모 (이 SPEC의 위험·사실에서 옮긴 것)

- **결정 기록**: OD-1 = 선택지 1(런처 워크트리 흐름, 2026-09-10), B1 = SPEC당 PR 하나 `feat/SPEC-*`, B2 = primary feature 브랜치 경로도 워크트리 흐름으로, B3 = `moai constitution amend` 로 `CONST-V3R5-027`·`028` 개정(HumanOversight는 운영자, 레인은 멈춰 올림). 모두 운영자 결정(리드 경유). OD-1 선택지 표는 `87988e946` spec.md §C.6.
- **REQ-WBG-011 근거 소실**: Phase D가 사라지면 SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-011의 근거 문장("Phase D is the ONLY place…", spec.md 210행)이 성립하지 않고 `internal/hook/branch_guard.go:455` 의 manager-git 신원 면제가 근거 없이 남는다 — 후속 카드 후보.
- **git-flow와 B1 접두**: `delivery.md` Step 3.2 git-flow는 `feature/*` 만 PR로 받아 `feat/SPEC-*` 브랜치는 "Any other branch → stop and report" 로 떨어진다. 배포 `AGENTS.md:125` 의 `WT-<slug>` 명명과 C1 이름 변경이 겹친다.
- **`feature/SPEC-` 형제**: `manager-git.md:82`·`:144`, `spec-assembly.md:388`(`--branch` 경로), `:547` 시나리오.
- **브랜치 변경 형제**: `manager-git.md:171`, `delivery.md:356`, `generic-patterns-guide.md`(템플릿 195행~, 로컬 207행~) Phase D Recovery.
- **Frozen 레지스트리 사실**: spec-workflow.md `[ZONE:Frozen]` 23·49·166행 중 등록 clause는 49행 블록의 `CONST-V3R5-027`·`028` 뿐, 23행 블록·166행은 미등록. `validator.go` 는 `ZONE_UNREGISTERED` 를 만들지 않는다. 검증기는 프로젝트 밖 레지스트리를 거부한다("escapes project dir"). worktree-integration.md 45·543-556행이 B1·B2와 모순되며 545·556행 `[ZONE:Frozen]` 은 미등록(B6).
- **amend 동작**: `--before` 바이트 대조(`internal/cli/constitution.go:528-529`), 레지스트리 해석 두 갈래(`:143-153` 대 `internal/constitution/pipeline.go:66`), 게이트 위치(`frozen_guard.go:22`, `canary.go:16/18/38`, `contradiction.go:87/110/124`, `rate_limiter.go:10/12/92`, `human_oversight.go:32/44-45`), 적용 도우미 스텁(`pipeline.go:256-266`, `pipeline_test.go:334-356`·`406-423`) → 카드 t659.
- **기준 트리 측정(`b412f8a33`)**: 명령 검출식 20줄(10/5/1/4), 서술형 검출식 11줄, 단계별 PR 검출식 19줄, 접두 합집합 {chore, feat, plan, sync}, delivery Step 3.2 github-flow 30줄, Step 3.3.5 11줄.
- **이 SPEC과의 겹침**: `manager-git.md:114` 의 `--<merge_method>` 치환은 이 SPEC이 먼저 착지한다(§E.2).
