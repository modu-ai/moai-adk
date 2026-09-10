---
id: SPEC-GIT-DELIVERY-PROCEDURE-001
title: "배포 지침의 git 전달 절차 결함 3건 수리 — primary checkout 브랜치 변경 안내, fetch 병렬 배치, auto-merge 기본값 충돌"
version: "0.1.1"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/agents/moai/manager-git.md, .claude/rules/moai/workflow/spec-workflow.md, .claude/rules/moai/core/agent-common-protocol.md, .claude/skills/moai/workflows/plan/spec-assembly.md, .claude/skills/moai/workflows/sync/delivery.md (+ internal/template/templates mirrors)"
lifecycle: spec-anchored
tags: "git, manager-git, late-branch, auto-merge, merge-method, sync-check, template-mirror, instruction-audit"
era: V3R6
tier: M
related_specs: [SPEC-V3R5-LATE-BRANCH-001, SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-MERGE-METHOD-CONFIG-001]
---

# SPEC-GIT-DELIVERY-PROCEDURE-001 — 배포 지침의 git 전달 절차 결함 3건 수리

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-10 | manager-spec | 최초 작성. 칸반 카드 t622(지침 감사 G1). 결함 3건(AC-01·AC-11·SX-R04)과 카드 밖 형제를 리드가 정한 범위로 묶었다. 재현 근거는 `.moai/reports/t622/repro.md`(측정 트리 `c352330d3`), 인용 줄번호는 이 SPEC 작성 트리 `b412f8a33`에서 다시 읽어 확인했다. 설계 판단 2건(OD-1·OD-2)은 결정하지 않고 §C에 표로 올렸다. |
| 0.1.1 | 2026-09-10 | manager-spec | plan-audit 1회차(`.moai/reports/t622/plan-audit.md`, FAIL 0.67) 반영. 필수 결함 D1~D4: AC-GDP-015 SPEC-ID 검출식을 다중 세그먼트로 바꾸고 REQ 토큰·커밋 SHA 판정 줄을 더함, AC-GDP-002를 Pre-Spawn 코드 블록 안으로 한정, AC-GDP-007 검출식을 `AGENTS.md` §2 금지 집합 전체로 넓혀 기준 트리에서 다시 잼(14줄 → 20줄)과 §A.2 표 확장, AC-GDP-008의 참조 0건 공허 통과 경로를 막고 바뀐 문구 참조 검사와 전체 조회 대조를 더함. 결정 표 보완 D11(OD-2 선택지 C의 키가 폐기 예정으로 분류됨)·D12(OD-1 선택지 1·3이 manager-git 면제의 적힌 근거를 없앰). 선택 결함 D5~D10·D13~D19 반영 내역은 acceptance.md와 progress.md에 기록. REQ 15개·AC 16개 유지, OD-1·OD-2는 여전히 미결정. |

---

## §A 배경과 문제

### A.1 결함 요약

| 결함 | 내용 | 설계 판단 |
|---|---|---|
| AC-01 | 배포 지침이 스스로 금지한 primary checkout 브랜치 변경을 같은 배포물 안에서 실행하라고 안내한다 | 있음 (OD-1) |
| AC-11 | `git fetch` 와 그 결과를 읽는 `git rev-list` 를 서로 독립인 병렬 배치로 지시한다 | 없음 |
| SX-R04 | 워크트리 문맥의 auto-merge 기본값이 두 지침에서 반대로 적혀 있고, 병합 명령에 `--squash` 가 박혀 있다 | 병합 방식 치환은 없음, 기본값 충돌은 있음 (OD-2) |

### A.2 AC-01 — 배포물 내부의 모순

전제는 이 트리의 문구 대조로 확인했다.

| 근거 | 위치 | 내용 |
|---|---|---|
| 무조건 금지 | `internal/template/templates/AGENTS.md:64` (루트 `AGENTS.md:64` 도 같은 문장) | "Never change branch state in the primary checkout." 조건 없음. 금지 집합은 `git checkout`·`git switch`·`git branch`·`git reset --hard`·`git stash`·현재 브랜치 위의 `git rebase`/`git merge` |
| 무조건 금지 | `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md:16` | `[HARD]` MUST NOT, 조건 없음 |
| 조건이 붙은 유일한 자리 | 같은 파일 `:85-86` | "the shared-checkout hazard does not apply to single-developer repos" — 기계 차단 훅의 **기본값**(`false`)에만 붙은 조건이며, 원칙 자체에는 붙어 있지 않다 |

따라서 AC-01은 이 저장소의 로컬 규칙과만 부딪히는 문제가 아니라, 배포되는 지침끼리의 모순이다. `git pull` 은 현재 브랜치에 병합이나 리베이스를 수행하므로 같은 금지 집합에 든다.

실행 안내 위치(`b412f8a33` 기준, 로컬과 템플릿 사본이 같은 줄번호). 0.1.1에서 검출식을 금지 집합 전체로 넓혀 다시 쟀다(`acceptance.md` AC-GDP-007).

| 파일 | 줄 | 안내 |
|---|---|---|
| `.claude/agents/moai/manager-git.md` | 42 | Checkpoint Rollback `git reset --hard [checkpoint-tag]` |
| 〃 | 96 | Late-branch Phase A `git checkout main && git pull origin main` |
| 〃 | 110 | Phase C `git switch -c feat/SPEC-XXX` |
| 〃 | 119, 121, 122 | Phase D `git checkout main` · `git reset --hard origin/main` · `git pull origin main` |
| 〃 | 125 | Phase A/B push 거부 시 복구 `git switch -c feat/SPEC-*` |
| 〃 | 127 | Phase D 누락 시 복구 `git fetch origin && git reset --hard origin/main` |
| 〃 | 137 | Personal Mode 옵션 `main_late_branch` 설명 |
| 〃 | 160 | Synchronization 절 `git fetch origin` → `git pull origin [branch]` |
| `.claude/rules/moai/workflow/spec-workflow.md` | 50 | Route B Step 1 late-branch 전제, `git switch -c plan/SPEC-XXX` |
| 〃 | 56, 58, 59 | Step 4 late-branch 종결 `git checkout main` · `git reset --hard origin/main` · `git pull origin main` (계약 원문 — `manager-git.md:129` 가 이 절을 참조) |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | 336 | Phase 13 late-branch 사전 점검, 수동 `git switch -c feat/SPEC-*` |
| `.claude/skills/moai/workflows/sync/delivery.md` | 323, 324 | Step 3.3.5 기준 브랜치 복귀 `git checkout {main_branch} && git pull origin {main_branch}` 등 |
| 〃 | 328 | 병합 뒤 로컬 브랜치 정리를 사용자에게 맡기며 `git branch -d <branch>` 안내 |

같은 검출식에 걸리지만 실행 안내가 아닌 줄(분류 장부에서 따로 분류):

| 파일 | 줄 | 성격 |
|---|---|---|
| `spec-workflow.md` | 62 | Step 4 누락 시의 실패 양상 서술("the next `git pull`") |
| `delivery.md` | 276 | `WT-*` 경로 — 통합 워크트리에 진입한 뒤 그 안에서 `git merge --no-ff <branch>` (워크트리 내부 실행) |

`manager-git.md:42`·`160` 과 `delivery.md:323-324`·`328` 은 late-branch 블록 밖의 일반 안내라, OD-1이 어떤 선택지로 결정되더라도 따로 처리 방식이 필요하다(§C.1 표에 표기).

### A.3 AC-11 — 결과를 읽는 명령과의 병렬 배치

| 파일 | 줄 | 문제 |
|---|---|---|
| `.claude/agents/moai/manager-git.md` | 156 | `git fetch`, `git status`, `git rev-list --count --left-right`, `gh pr checks --json` 를 "independent and read-only" 로 묶어 한 턴 병렬 배치로 지시 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 292, 296, 299 | Pre-Spawn Sync Check 가 "parallel batch" 를 선언하고 `git fetch origin main`(296)과 `git rev-list`(299)를 코드 블록 안 별도 명령으로 나열 |
| 같은 파일 (정상) | 347 | Pre-Edit Sync Check 는 `git fetch origin main 2>&1; git rev-list ...` 를 한 명령으로 묶어 순서가 보장됨 — 유지 대상 |

`git fetch` 는 원격 추적 ref 를 갱신하고 `git rev-list` 는 그 ref 를 읽으므로 둘은 독립이 아니다. 병렬로 돌면 `rev-list` 가 갱신 전 ref 를 읽을 수 있고, 그러면 원격이 앞선 상태(`N 0`)가 `0 0` 으로 보고된다.

### A.4 SX-R04 — auto-merge 기본값 충돌과 병합 방식 고정

| 파일 | 줄 | 내용 |
|---|---|---|
| `delivery.md` | 335-338 | 트리거: `is_worktree_context == true` 이고 `--no-merge` 가 없으면 병합 (기본 병합) |
| `delivery.md` | 343, 355 | `gh pr merge --squash --delete-branch` 고정 |
| `manager-git.md` | 148 | Team Mode: "Auto-merge: only with the `--auto-merge` flag" |
| `manager-git.md` | 166, 170 | "Execute only with `--auto-merge` flag AND all approvals obtained" · `--<merge_method>` 해석 |
| `manager-git.md` | 114 | Phase C 예시가 `--squash` 로 고정 |
| `manager-git.md` | 32 | `merge_method` 해석 규칙 — 기본값 설명으로 `gh pr merge --squash --delete-branch` 를 인용 (유지 대상) |

team 모드이면서 워크트리 문맥이면 두 지침이 반대를 지시한다. `merge_method` 치환 자체는 SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-007이 이미 요구한 것이 `delivery.md` 에 반영되지 않은 상태다.

### A.5 설정 사실 (측정)

| 항목 | 측정 결과 |
|---|---|
| git-strategy 템플릿 경로 | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` (`internal/template/templates/.moai/config/sections/git-strategy.yaml` 은 존재하지 않음) |
| `branch_creation.auto_enabled` | 템플릿 `manual`·`personal`·`team` 모두 `false` (28, 61, 100행) · Go 기본값 `internal/config/defaults.go` 739·752·768행 모두 `AutoEnabled: false` · 필드 정의 `internal/config/types.go:75` (Go 코드는 읽지 않고 `spec-assembly.md` 가 YAML을 직접 읽는다는 주석 69-74행) |
| `main_late_branch` | **설정 키가 아니다.** `.moai/` 밖 검색에서 로컬·템플릿 `manager-git.md:137` 과 생성물 `.codex/agents/moai/manager-git.toml:131` 의 문구에만 나온다. 템플릿 YAML·Go 설정에는 없다 |
| `merge_method` | 템플릿 세 모드 모두 `squash` (23, 56, 95행) · Go 기본값 `defaults.go` 738·751·767행 `MergeMethod: "squash"` · 필드 `types.go:142` · 열거형 검증 `internal/config/validation.go:254-305` |
| `workflow.worktree.auto_merge` | 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml:58` `false` · Go 기본값 `defaults.go:872` `false` · `types.go:623` 주석 "AutoMerge has no production reader" · 사용처 분류 `internal/config/testdata/shipped_key_inventory.yaml:2940-2943` `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` |
| `workflow.branch_guard.enabled` | Go 기본값 `defaults.go:893-895` `false` · 템플릿 `workflow.yaml` 에는 키가 없음 |

### A.6 선행 SPEC 관계

| SPEC | 상태 | 이 SPEC과의 관계 |
|---|---|---|
| SPEC-V3R5-LATE-BRANCH-001 | completed | late-branch 절차의 기원. REQ-LB-001~006이 4단계 절차와 `git reset --hard origin/main` 종결을 규정. EXCL-LB-004가 "late-branch는 단일 main checkout 안에서 동작"이라고 워크트리 확장을 제외했다 |
| SPEC-WORKTREE-BRANCH-GUARD-001 | completed | primary checkout 가드. §A가 `manager-git.md` Phase D를 primary checkout에서 정당하게 실행되는 예외로 인용했고, REQ-WBG-011(spec.md 210행)은 manager-git 신원 면제의 근거를 "`manager-git.md` Phase D is the ONLY place the project authorizes branch-state changes in the primary checkout" 로 적었다. 면제 코드는 `internal/hook/branch_guard.go:455` `return input.AgentType == "manager-git"` |
| SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 | completed | 가드 훅을 기본 꺼짐으로 전환. "The guard's intent (shared-checkout multi-session protection) does not apply to single-developer repos" 문장의 기원 |
| SPEC-MERGE-METHOD-CONFIG-001 | completed | `merge_method` 도입. REQ-MMC-007/008이 `delivery.md`·`manager-git.md` 의 `--squash` 고정 제거를 요구, REQ-MMC-009가 기본값에서 명령이 바이트 동일할 것을 요구 |
| SPEC-GITFLOW-DOCTRINE-ALIGN-001 | in-progress | 겹치지 않음. 대상이 로컬 전용 문서 3종(`.moai/docs/git-workflow-doctrine.md`, `.moai/docs/git-local-workflow-doctrine.md`, `.claude/rules/local/repo-local-pr-policy.md`)이고 REQ-GDA-006이 `internal/template/templates/**` 수정을 명시적으로 막는다 |

워크트리 문맥 기본 병합 문구는 SPEC에서 나오지 않았다. `git log -S` 추적 결과 커밋 `b7447cb90`(2026-04-03, `sync.md`)에서 들어왔고, SPEC-V3R4-WORKFLOW-SPLIT-001 커밋 `980ccdc56` 이 `delivery.md` 로 옮겼다(분할 이동). 추적한 문구는 `delivery.md:349` 의 "Auto-merge is now the default for worktree contexts" 이다.

---

## §B 요구사항 (GEARS)

요구사항 식별자는 `REQ-GDP-NNN`. 결정 대기 요구사항은 제목에 **[DEFERRED: OD-n]** 을 붙였다. 결정 전에는 run-phase가 해당 요구사항을 구현하지 않는다.

### B.1 AC-11 — fetch 순서 보장 (결정 불필요)

**REQ-GDP-001** (Ubiquitous)
`manager-git` 에이전트 정의의 동기화 절은 `git fetch` 가 끝난 뒤에 그 결과를 읽는 `git rev-list --count --left-right` 가 실행되도록 순서를 지시해야 하며, `git fetch` 를 그 결과를 읽는 명령과 서로 독립인 병렬 배치 항목으로 분류해서는 안 된다. fetch 결과를 읽지 않는 명령(`git status`, `gh pr checks --json`)을 병렬로 두는 일은 이 요구사항이 제한하지 않는다.

**REQ-GDP-002** (Ubiquitous)
`agent-common-protocol.md` 의 Pre-Spawn Sync Check 코드 블록은 `git fetch origin main` 과 `git rev-list --count --left-right origin/main...HEAD` 를 순서가 보장되는 한 명령으로 지시해야 한다. 세 번째 명령(`moai session list --json --filter-spec=<SPEC-ID>`)의 병렬 실행 허용과 두 해석 표의 내용은 바뀌어서는 안 된다.

**REQ-GDP-003** (Unwanted)
`agent-common-protocol.md` 의 Pre-Edit Sync Check 절(328행 제목부터 `#### The sweep prohibition` 소절 앞까지)은 이 SPEC의 변경으로 수정되어서는 안 된다.

### B.2 SX-R04 — 병합 방식 해석 (결정 불필요)

**REQ-GDP-004** (Ubiquitous)
`sync/delivery.md` 의 auto-merge 실행 단계는 병합 명령을 활성 모드의 `git_strategy.<mode>.merge_method`(기본 `squash`)로 해석한 `gh pr merge --<merge_method> --delete-branch` 로 지시해야 한다. 기본값에서 실제 실행 명령은 지금과 바이트 동일해야 한다(SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-009 연속성).

**REQ-GDP-005** (Unwanted)
범위 파일의 어떤 실행 예시도 `--squash` 를 고정한 `gh pr merge` 명령을 실행할 명령으로 제시해서는 안 된다. `manager-git.md:32` 의 기본값 설명 문장은 남아야 한다. Late-branch 절차 블록이 OD-1 결정 뒤에도 남는 경우 그 예시는 `--<merge_method>` 로 적혀야 한다.

### B.3 SX-R04 — 기본값 단일 기준 [DEFERRED: OD-2]

**REQ-GDP-006** (Event-driven) **[DEFERRED: OD-2]**
When team 모드와 워크트리 문맥이 겹치는 sync 병합 판단이 일어날 때, `delivery.md` 와 `manager-git.md` 는 OD-2에서 정한 단 하나의 기준을 따라야 하며, 기준이 아닌 쪽은 반대 기본값을 적지 않고 기준을 이름으로 참조해야 한다.

### B.4 AC-01 — 결정과 무관한 불변식

**REQ-GDP-007** (Ubiquitous)
변경 후 §A.2 표의 각 안내 위치는 둘 중 하나여야 한다. (a) primary checkout에서 브랜치 상태를 바꾸는 명령을 더 이상 실행하라고 지시하지 않는다. (b) 그 안내와 `AGENTS.md` §2 금지문(루트·템플릿)이 같은 조건을 명시적으로 공유한다. 무조건 금지와 무조건 실행 안내가 같은 배포물 안에 공존해서는 안 된다.

**REQ-GDP-008** (Unwanted)
Late-branch 절차의 명령 예시만 지우고 대체 절차나 은퇴 선언을 남기지 않는 변경은 해서는 안 된다. `spec-assembly.md` 와 `spec-workflow.md` 가 참조하는 절차 제목은 참조 대상이 실제로 존재해야 한다.

### B.5 AC-01 — 선택지별 요구사항 [DEFERRED: OD-1]

**REQ-GDP-009** (Ubiquitous) **[DEFERRED: OD-1]**
`spec-workflow.md` 의 Step 1·Step 4 계약, `manager-git.md` 절차, `spec-assembly.md` 사전 점검, `delivery.md` 복귀 단계는 OD-1에서 정한 하나의 절차를 서로 같은 진입·종결 명령으로 서술해야 한다.

**REQ-GDP-010** (Capability-gate) **[DEFERRED: OD-1 = 선택지 1]**
Where OD-1이 런처 워크트리 흐름 재설계로 결정되면, 절차는 `moai cc -w <name>` 또는 `EnterWorktree(<path>)` 로 진입한 워크트리 안에서 커밋을 쌓아야 하고 맨손 `git worktree add` 를 지시해서는 안 된다. `spec-workflow.md` 49행 `[ZONE:Frozen]` 구역 수정은 plan-phase 산출물에 Frozen 구역 수정으로 명시하고 Implementation Kickoff Approval을 거쳐야 한다.

**REQ-GDP-011** (Capability-gate) **[DEFERRED: OD-1 = 선택지 2]**
Where OD-1이 조건부 유지로 결정되면, 그 조건은 기계적으로 판정 가능한 한 문장으로 정해져야 하고, 루트 `AGENTS.md` §2, 템플릿 `AGENTS.md` §2, `main-checkout-branch-guard.md`(로컬·템플릿) 16행 원칙, 남는 각 안내 위치에 같은 문구로 실려야 한다.

**REQ-GDP-012** (Capability-gate) **[DEFERRED: OD-1 = 선택지 3]**
Where OD-1이 `main_late_branch` 은퇴로 결정되면, 은퇴 사실이 `manager-git.md` 에 명시되어야 하고, `branch_creation.auto_enabled: false` 일 때 `/moai plan` 이 하는 일이 새로 정의되어야 하며, `--issue` opt-in 기본값(SPEC-V3R5-LATE-BRANCH-001 REQ-LB-009)은 유지되어야 한다.

### B.6 사본·생성물·중립성

**REQ-GDP-013** (Ubiquitous)
변경한 줄은 로컬 사본(`.claude/…`)과 템플릿 사본(`internal/template/templates/.claude/…`)에서 같아야 한다. `delivery.md` 두 사본의 의도된 차이(통합 워크트리 경로 표기 275·278행, 꼬리말 479-480행)는 그대로 남아야 한다.

**REQ-GDP-014** (Ubiquitous)
템플릿 `manager-git.md` 가 바뀌면 `internal/template/templates/.codex/agents/moai/manager-git.toml` 은 `make agents-emit` 으로 재생성되어 같은 카드에 커밋되어야 하며, 손으로 편집해서는 안 되고, `make agents-emit-check` 가 exit 0 이어야 한다.

**REQ-GDP-015** (Unwanted)
`internal/template/templates/` 아래에 추가되는 문구는 SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 16개 지원 프로그래밍 언어 중 특정 언어로 치우친 표현을 담아서는 안 된다.

---

## §C 결정 대기 항목

리드가 운영자에게 전달한다. 이 문서는 선택지를 사실로만 서술하고 권고하지 않는다. 표의 줄번호는 모두 `b412f8a33` 기준이며 "L·T" 는 로컬·템플릿 두 사본을 뜻한다.

### C.1 OD-1 — Late-branch 절차의 처리 (AC-01)

| 선택지 | 영향 파일 (로컬·템플릿·생성물) | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 (키 · 현재 기본값 · 변화) | 뒤집거나 확장하는 선행 SPEC |
|---|---|---|---|---|
| **1. 런처 워크트리 흐름으로 재설계** | `manager-git.md` L·T 88-129(절차 블록)·137(옵션 행) + 재생성 `.codex/agents/moai/manager-git.toml`; `spec-workflow.md` L·T 49-62(`[ZONE:Frozen]` Step 1 전제·Step 4 종결)·65(plan 단계 워크트리 금지 안티패턴); `spec-assembly.md` L·T 332-340; `delivery.md` L·T 319-326. 이름만 인용: 템플릿 `CLAUDE.md:75`, `agent-authoring.md:138`, `delegation.yaml:74`, `generic-patterns-guide.md`(템플릿 195행~, 로컬 207행~). 이 선택지가 스스로 답하지 않는 자리: `manager-git.md:42`·`160`, `delivery.md:323-324`·`328` | 부분. "브랜치를 PR 시점에 만든다"는 결과는 유지되고, 커밋을 쌓는 장소가 primary checkout의 `main` 에서 런처로 진입한 워크트리로 바뀐다. 기존 Phase A-D를 primary checkout에서 따르던 사용자는 다른 절차를 받는다. Step 1 "plan은 main checkout에서, 워크트리 없이" 규칙과 부딪혀 Frozen 구역 수정이 따라온다 | `git_strategy.{manual,personal,team}.branch_creation.auto_enabled` — 현재 `false`(템플릿 28·61·100, Go `defaults.go` 739·752·768). 값은 그대로, `false` 가 촉발하는 동작이 바뀐다. 새 키는 현재 없음(설계에 따라 추가될 수 있음) | SPEC-V3R5-LATE-BRANCH-001 REQ-LB-003·005·006을 대체, EXCL-LB-004(워크트리 확장 제외)를 뒤집음. SPEC-WORKTREE-BRANCH-GUARD-001 §A가 Phase D를 primary checkout 예외로 인용한 전제가 사라지고, REQ-WBG-011이 manager-git 신원 면제의 근거로 적은 "Phase D가 primary checkout 브랜치 변경을 허용하는 유일한 자리"(spec.md 210행)도 사라진다. `internal/hook/branch_guard.go:455` 의 면제는 적힌 근거 없이 남는다(훅 변경은 이 SPEC 범위 밖) |
| **2. 유지하되 명시적 조건으로 범위 한정** | 선택지 1의 네 지침 파일 L·T + 재생성 `.toml`; **함께 정해야 하는 금지문**: 루트 `AGENTS.md:64`, 템플릿 `internal/template/templates/AGENTS.md:64`, `main-checkout-branch-guard.md` L·T 16; 필요 시 `main-checkout-branch-guard-detail.md` L·T | 철회 없음. 대신 모든 배포 사용자에게 적용되는 금지 원칙이 조건부로 바뀌어, 조건이 성립하는 저장소에서는 primary checkout 브랜치 변경이 허용되는 것으로 읽힌다. `AGENTS.md` 는 병합 체인이 넘치면 꼬리부터 조용히 잘리는 문서라(같은 파일 Budget warning) 문장이 늘면 뒤쪽 조항이 밀려날 수 있다 | 기존 키에 조건을 묶는다면 후보는 `workflow.branch_guard.enabled` — 현재 Go 기본값 `false`(`defaults.go:893-895`), 템플릿 `workflow.yaml` 에는 키가 없다. 가드 규칙 86행이 이 기본값에만 "1인 저장소에는 해당 없음"을 붙여 둠. 값 변화 없음, 키의 의미가 훅 on/off에서 원칙 적용 범위로 넓어짐 | SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001의 "1인 저장소에는 해당 없음" 전제를 훅 기본값 층에서 원칙 층으로 확장. SPEC-WORKTREE-BRANCH-GUARD-001의 무조건 원칙(규칙 16행)을 좁힘. Phase D가 남으므로 REQ-WBG-011의 면제 근거는 유지됨. SPEC-V3R5-LATE-BRANCH-001은 유지 |
| **3. `main_late_branch` 은퇴** | `manager-git.md` L·T 88-129·137 + 재생성 `.toml`(131행 포함); `spec-workflow.md` L·T 50·53-62; `spec-assembly.md` L·T 332-340(사전 점검과 "Late-branch (main commit + late switch)" 표시); 템플릿 `git-strategy.yaml.tmpl` 82-84 주석; 이름 인용처 `CLAUDE.md:75`·`agent-authoring.md:138`·`delegation.yaml:74`·`generic-patterns-guide.md`; `--issue` 기본값을 "late-branch opt-in policy"로 부르는 `SKILL.md` 124·227, `workflows/moai.md` 45·233, `spec-assembly.md` 258. 이 선택지가 스스로 답하지 않는 자리: `manager-git.md:42`·`160`, `delivery.md:323-324`·`328` | 예. Phase A-D 4단계 절차가 배포 사용자에게서 사라진다. `main_late_branch` 는 설정 키가 아니므로 지울 설정은 없고, 철회되는 것은 `auto_enabled: false` 에 묶인 문서화된 동작이다 | `branch_creation.auto_enabled` 현재 `false`(세 모드). `false` 일 때 `/moai plan` 이 브랜치 생성을 건너뛰는 동작(`spec-assembly.md:336`)을 새로 정의해야 한다 — 기본값을 `true` 로 바꾸거나 `false` 에 새 의미를 부여. `automation.auto_branch` 현재 `false` 도 같은 검토 대상 | SPEC-V3R5-LATE-BRANCH-001 REQ-LB-001~006을 뒤집음. REQ-LB-009(`--issue` opt-in)는 명시적으로 보존하지 않으면 이름 인용을 통해 함께 흔들림. Phase D가 사라지므로 SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-011이 manager-git 신원 면제의 근거로 적은 문장(spec.md 210행)이 성립하지 않게 되고, `internal/hook/branch_guard.go:455` 의 면제는 적힌 근거 없이 남는다(훅 변경은 이 SPEC 범위 밖) |
| ~~예시 명령만 삭제~~ **(제외)** | — | 제외 사유: `auto_enabled` 기본값이 세 모드 모두 `false` 라 `/moai plan` 은 계속 브랜치 생성을 건너뛰고(`spec-assembly.md:336`) 사용자에게 절차를 참조하라고 안내하는데(`spec-assembly.md:340`, `spec-workflow.md:62`), 그 절차는 사라진 채 남는다. 기능이 공지 없이 철회되므로 REQ-GDP-008이 금지한다 | — | — |

### C.2 OD-2 — 워크트리 문맥 auto-merge 기본값의 단일 기준 (SX-R04)

| 선택지 | 영향 파일 (로컬·템플릿·생성물) | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 (키 · 현재 기본값 · 변화) | 뒤집거나 확장하는 선행 SPEC |
|---|---|---|---|---|
| **A. `delivery.md` 워크트리 기본값을 기준으로** (`--no-merge` 가 없으면 병합) | `manager-git.md` L·T 148·166 + 재생성 `.toml`(142·160행); `delivery.md` L·T 335-349는 유지(기준 명시 문구 추가 가능); `sync/doc-execution.md` L·T 36은 이미 일치 | 변경. team 모드이면서 워크트리 문맥이면 `--auto-merge` 플래그 없이 병합된다. 승인 누락 시 병합하지 않는 조건(`delivery.md:363`)은 남는다 | 설정 키 없이 플래그로 결정. `workflow.worktree.auto_merge`(현재 `false` — 템플릿 `workflow.yaml:58`, Go `defaults.go:872`)는 계속 읽는 곳이 없고, "옵트인, 기본 꺼짐"이라는 문서(`moai-workflow-worktree/modules/moai-adk-integration.md:195`)와 어긋난 채 남음 | 기본 병합 문구는 SPEC 없이 커밋 `b7447cb90` 에서 들어왔고 SPEC-V3R4-WORKFLOW-SPLIT-001이 옮겼다 — 그 동작을 확정. SPEC-MERGE-METHOD-CONFIG-001과는 축이 다름(병합 방식) |
| **B. `manager-git.md` 옵트인을 기준으로** (`--auto-merge` 와 전원 승인) | `delivery.md` L·T 335-338(트리거)·348-349(`--no-merge`/`--merge` 설명, "Auto-merge is now the default for worktree contexts"); `sync/doc-execution.md` L·T 34-36; `manager-git.md` 는 유지(재생성 불필요) | 변경. 워크트리 문맥의 sync가 기본으로 병합하지 않는다. 지금 기본 병합에 기대는 사용자는 플래그를 넘겨야 한다. `--merge` 폐기 경고 문구의 논리가 뒤집힌다 | 키 없음. 워크트리 문맥의 동작 기본값이 "병합"에서 "병합 안 함"으로 바뀜. `workflow.worktree.auto_merge` 는 여전히 읽는 곳이 없음 | `b7447cb90` 에서 들어온 동작을 뒤집음(SPEC 기원 없음). SPEC-MERGE-METHOD-CONFIG-001 무관 |
| **C. 기존 설정 키 `workflow.worktree.auto_merge` 를 기준으로** | `delivery.md` L·T 335-349; `sync/doc-execution.md` L·T 36; `manager-git.md` L·T 148·166 + 재생성 `.toml`; 읽는 곳 없음을 기록한 `internal/config/types.go:623` 주석, 사용처 분류 `internal/config/testdata/shipped_key_inventory.yaml:2940-2943`, `internal/web/dead_config_guard_test.go:72` 가 갱신 대상이 될 수 있음(지침 문구가 "읽는 곳"으로 분류되는지는 미검증) | 변경. 기본값이 `false` 라 워크트리 문맥의 기본 병합이 멈추고, 키를 켠 사용자만 자동 병합한다. team 모드 플래그는 호출 단위 재정의로 남길 수 있다 | `workflow.worktree.auto_merge` 값은 `false` 그대로(템플릿 `workflow.yaml:58`, Go `defaults.go:872`), 키에 처음으로 소비자가 생긴다. 사용처 분류 파일은 이 키를 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` 으로 기록하고 있으며(`shipped_key_inventory.yaml:2940-2943`), 이 SPEC의 목표 릴리스는 `v3.2.0` 이다 | `defaults.go:867-872` 주석에 이름이 적힌 SPEC-WORKTREE-ENTRY-STRATEGY-001의 기본값 true→false 전환을 동작으로 확장. SPEC-MERGE-METHOD-CONFIG-001의 "설정 필드 → 지침 문구" 패턴을 따름. `b7447cb90` 동작은 기본값에서 뒤집힘 |

---

## §D 제외 범위 (Out of Scope)

### Out of Scope — 결정 없이 확정하지 않는 것

- OD-1·OD-2의 선택. 결정 전에는 REQ-GDP-006, REQ-GDP-009~012를 구현하지 않는다.
- 결정되지 않은 선택지를 전제로 한 문구 초안을 지침 파일에 미리 넣는 일.

### Out of Scope — 코드와 설정 값

- Go 코드(`internal/config`, `internal/hook`, `internal/github`) 변경. OD-1 선택지 1·3이 남기는 `branch_guard.go:455` 면제의 근거 공백과, OD-2 선택지 C가 요구할 수 있는 사용처 분류 갱신은 결정 뒤 별도로 판단한다.
- `git-strategy.yaml.tmpl`·`workflow.yaml` 의 키 값 변경. OD-1 선택지 3이 기본값 변경을 요구하면 그 결정에 따라 범위를 다시 정한다.
- `main-checkout-branch-guard` 훅의 패턴·면제 로직 변경.

### Out of Scope — 카드 범위 밖에서 관측한 형제 (리드 판단 대상)

리드가 정한 범위를 넓히지 않으려고 기록만 한다. 같은 결함 계열이다.

- `manager-git.md:82` — `auto_branch: true` 일 때 "checkout from main_branch".
- `manager-git.md:171` — PR Auto-Merge 5단계 "Checkout main, pull, delete local branch".
- `delivery.md:356` — Auto-Merge Execution 4단계 "Checkout target branch, fetch latest".
- `sync/doc-execution.md:36`(로컬·템플릿 같은 문장) — "worktree contexts default to auto-merge". OD-2 선택지 B·C에서는 반대 기본값을 적은 문장이 된다.
- `generic-patterns-guide.md`(템플릿 195행~, 로컬 207행~) Late-Branch Phase D Recovery — `git stash push`, `git reset --keep`, `git checkout stash@{0} --` 를 워크트리 한정 없이 안내한다.

### Out of Scope — 금지 목록과 경고표

- 다음 위치의 브랜치 변경 명령은 금지 목록·경고표·인용 예시·권한 항목·워크트리 내부 안내라서 유지한다: 템플릿 `AGENTS.md` 64-66, `main-checkout-branch-guard.md` 20-24와 detail 파일, `coding-standards.md:141`, `moai-ref-git-workflow/SKILL.md` 144-145, `verification-claim-integrity-detail.md:161`, `settings.json.tmpl:564`, `moai-workflow-worktree/modules/troubleshooting.md` 176·179.

---

## §E 미검증 (Gaps)

- 파괴적 명령을 실제로 실행해 공유 checkout이 바뀌는 것을 재현하지 않았다. AC-01 판정은 실행 안내와 배포 금지문의 대조에 근거한다.
- `git fetch`/`git rev-list` 경합을 실제로 일으키지 않았다. 감사 보고서도 AC-11 신뢰도를 medium으로 매겼다.
- 배포 사용자가 `main_late_branch` 절차를 실제로 쓰는지는 확인할 수 없다.
- OD-2 선택지 C에서 지침 문구의 설정 읽기가 사용처 분류 테스트의 "reader"로 인정되는지, 그리고 `deprecate_after: "v3.1.0"` 이 v3.2.0 이후에 어떤 검사를 일으키는지 확인하지 않았다.
- OD-1 선택지 2가 `AGENTS.md` 바이트 예산에 주는 영향을 측정하지 않았다(현재 루트 15,277바이트, 템플릿 16,936바이트).
- `manager-git.md` 와 `agent-common-protocol.md` 는 `rule_template_mirror_test.go` 의 바이트 동일 대상에서 빠져 있다(같은 파일 주석). 두 파일의 사본 일치는 이 SPEC의 diff 기준으로만 확인된다.
- `TestTemplateNoInternalContentLeak` 의 날짜·SHA 검출 분류가 에이전트·규칙·스킬 경로까지 적용되는지는 읽지 않았다. 그래서 AC-GDP-015는 이 테스트에 기대지 않고 추가 줄을 직접 검사한다.
- 워크트리 기본 병합 문구의 `git log -S` 추적은 `delivery.md:349` 문장에만 수행했고, 트리거 문구(337행)와 `doc-execution.md:36` 에는 수행하지 않았다.

---

## §F 수용 기준

`acceptance.md` 가 기준(AC-GDP-001~016)과 검증 명령, 양성 대조, 1회차 감사 뮤턴트에 대한 재실행 결과를 담는다.
