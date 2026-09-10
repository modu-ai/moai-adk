---
id: SPEC-GIT-DELIVERY-PROCEDURE-001
title: "배포 지침의 git 전달 절차 결함 3건 수리 — primary checkout 브랜치 변경 안내, fetch 병렬 배치, auto-merge 기본값 충돌"
version: "0.1.2"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/agents/moai/manager-git.md, .claude/rules/moai/workflow/spec-workflow.md, .claude/rules/moai/core/agent-common-protocol.md, .claude/skills/moai/workflows/plan/spec-assembly.md, .claude/skills/moai/workflows/sync/delivery.md, .claude/skills/moai/workflows/sync/doc-execution.md (+ internal/template/templates mirrors)"
lifecycle: spec-anchored
tags: "git, manager-git, late-branch, launcher-worktree, auto-merge, merge-method, sync-check, template-mirror, instruction-audit"
era: V3R6
tier: M
related_specs: [SPEC-V3R5-LATE-BRANCH-001, SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-MERGE-METHOD-CONFIG-001]
---

# SPEC-GIT-DELIVERY-PROCEDURE-001 — 배포 지침의 git 전달 절차 결함 3건 수리

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-10 | manager-spec | 최초 작성. 칸반 카드 t622(지침 감사 G1). 결함 3건(AC-01·AC-11·SX-R04)과 카드 밖 형제를 리드가 정한 범위로 묶었다. 재현 근거는 `.moai/reports/t622/repro.md`(측정 트리 `c352330d3`), 인용 줄번호는 이 SPEC 작성 트리 `b412f8a33`에서 다시 읽어 확인했다. 설계 판단 2건(OD-1·OD-2)은 결정하지 않고 §C에 표로 올렸다. |
| 0.1.1 | 2026-09-10 | manager-spec | plan-audit 1회차(`.moai/reports/t622/plan-audit.md`, FAIL 0.67) 반영. 필수 결함 D1~D4(SPEC-ID 검출식, Pre-Spawn 블록 한정, 금지 집합 전체 검출식 14줄 → 20줄, 참조 0건 공허 경로)와 결정 표 보완 D11·D12, 선택 결함 D5~D10·D13~D19. REQ 15개·AC 16개 유지. |
| 0.1.2 | 2026-09-10 | manager-spec | 운영자 결정(리드 경유) 반영: OD-1 = 선택지 1(런처 워크트리 흐름으로 재설계), OD-2 = 선택지 B(`manager-git.md` 옵트인이 단일 기준). §C를 결정 기록으로 바꾸고, 재설계 흐름의 진입·PR·종결 단계와 흐름이 스스로 답하지 않는 네 자리의 처리를 정의했다. REQ-GDP-006·009·010 활성, REQ-GDP-011·012와 AC-GDP-011·012는 번호를 유지한 채 철회. `sync/doc-execution.md` 34-36행을 범위에 넣었다. spec-workflow.md Step 1이 등록된 Frozen 절(CONST-V3R5-027)이라 헌법 레지스트리 갱신이 따라온다는 사실과, run-phase 착수 전에 정해야 할 설계 항목 B1~B3을 기록했다. plan-audit 2회차(`.moai/reports/t622/plan-audit-iter2.md`, FAIL 0.75) 결함 N1~N6 반영. REQ 15개(활성 13·철회 2)·AC 16개(활성 14·철회 2). |

---

## §A 배경과 문제

### A.1 결함 요약

| 결함 | 내용 | 처리 |
|---|---|---|
| AC-01 | 배포 지침이 스스로 금지한 primary checkout 브랜치 변경을 같은 배포물 안에서 실행하라고 안내한다 | OD-1 = 선택지 1(런처 워크트리 흐름으로 재설계), §C.2 |
| AC-11 | `git fetch` 와 그 결과를 읽는 `git rev-list` 를 서로 독립인 병렬 배치로 지시한다 | 기계적 수리 (M1) |
| SX-R04 | 워크트리 문맥의 auto-merge 기본값이 두 지침에서 반대로 적혀 있고, 병합 명령에 `--squash` 가 박혀 있다 | 병합 방식 치환은 기계적 수리(M2), 기본값은 OD-2 = 선택지 B(M3) |

### A.2 AC-01 — 배포물 내부의 모순

전제는 이 트리의 문구 대조로 확인했다.

| 근거 | 위치 | 내용 |
|---|---|---|
| 무조건 금지 | `internal/template/templates/AGENTS.md:64` (루트 `AGENTS.md:64` 도 같은 문장) | "Never change branch state in the primary checkout." 조건 없음. 금지 집합은 `git checkout`·`git switch`·`git branch`·`git reset --hard`·`git stash`·현재 브랜치 위의 `git rebase`/`git merge` |
| 무조건 금지 | `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md:16` | `[HARD]` MUST NOT, 조건 없음 |
| 조건이 붙은 유일한 자리 | 같은 파일 `:85-86` | "the shared-checkout hazard does not apply to single-developer repos" — 기계 차단 훅의 **기본값**(`false`)에만 붙은 조건이며, 원칙 자체에는 붙어 있지 않다 |

따라서 AC-01은 이 저장소의 로컬 규칙과만 부딪히는 문제가 아니라, 배포되는 지침끼리의 모순이다. `git pull` 은 현재 브랜치에 병합이나 리베이스를 수행하므로 같은 금지 집합에 든다. 금지는 primary checkout에만 걸린다 — 런처로 진입한 워크트리 안의 같은 명령은 금지 대상이 아니다.

실행 안내 위치(`b412f8a33` 기준, 로컬과 템플릿 사본이 같은 줄번호). 검출식은 금지 집합 전체와 `git -C <경로>` 형태까지 잡는다(`acceptance.md` AC-GDP-007).

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

late-branch 블록 안의 줄은 §C.2 흐름으로 대체된다. 블록 밖의 네 자리(`manager-git.md:42`·`160`, `delivery.md:323-324`·`328`)는 §C.3에서 처리를 정한다.

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
| `delivery.md` | 348-349 | `--no-merge` 설명, `--merge` 폐기 경고 "Auto-merge is now the default for worktree contexts." |
| `sync/doc-execution.md` | 34-36 | 워크트리 문맥 감지 절이 `is_worktree_context` 를 정의하고 "This affects auto-merge behavior: worktree contexts default to auto-merge." 라고 적음 (로컬·템플릿 같은 줄) |
| `manager-git.md` | 148 | Team Mode: "Auto-merge: only with the `--auto-merge` flag" |
| `manager-git.md` | 166, 170 | "Execute only with `--auto-merge` flag AND all approvals obtained" · `--<merge_method>` 해석 |
| `manager-git.md` | 114 | Phase C 예시가 `--squash` 로 고정 |
| `manager-git.md` | 32 | `merge_method` 해석 규칙 — 기본값 설명으로 `gh pr merge --squash --delete-branch` 를 인용 (유지 대상) |

team 모드이면서 워크트리 문맥이면 두 지침이 반대를 지시한다. `merge_method` 치환 자체는 SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-007이 이미 요구한 것이 `delivery.md` 에 반영되지 않은 상태다.

### A.5 설정·레지스트리 사실 (측정)

| 항목 | 측정 결과 |
|---|---|
| git-strategy 템플릿 경로 | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` (`internal/template/templates/.moai/config/sections/git-strategy.yaml` 은 존재하지 않음) |
| `branch_creation.auto_enabled` | 템플릿 `manual`·`personal`·`team` 모두 `false` (28, 61, 100행) · Go 기본값 `internal/config/defaults.go` 739·752·768행 모두 `AutoEnabled: false` · 필드 정의 `internal/config/types.go:75` (Go 코드는 읽지 않고 `spec-assembly.md` 가 YAML을 직접 읽는다는 주석 69-74행) |
| `main_late_branch` | **설정 키가 아니다.** `.moai/` 밖 검색에서 로컬·템플릿 `manager-git.md:137` 과 생성물 `.codex/agents/moai/manager-git.toml:131` 의 문구에만 나온다 |
| `merge_method` | 템플릿 세 모드 모두 `squash` (23, 56, 95행) · Go 기본값 `defaults.go` 738·751·767행 `MergeMethod: "squash"` · 필드 `types.go:142` · 열거형 검증 `internal/config/validation.go:254-305` |
| 워크트리 기준 브랜치 | `worktree-integration.md` § Worktree Base Branch: 네이티브 워크트리는 기본으로 `origin/HEAD`(`worktree.baseRef` 기본값 `fresh`)에서 잘리고, `git_strategy.worktree_base_branch` 기본값은 빈 값(아무 동작 안 함) |
| 헌법 레지스트리 | `.claude/rules/moai/core/zone-registry.md` 가 spec-workflow.md Step 1 절을 `CONST-V3R5-027`(Frozen, `canary_gate: true`, clause "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step"), Step 4 절을 `CONST-V3R5-028`(Frozen, clause "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged")로 등록. 로컬·템플릿 레지스트리 사본은 바이트 동일(diff exit 0)이며 미러 테스트 허용 목록에는 없다 |
| 헌법 검증기 | `internal/constitution/validator.go`: 등록된 clause 가 원본 파일에 공백 정규화 뒤 부분 문자열로 없으면 `DRIFT`, Frozen 항목의 `canary_gate: false` 는 `FROZEN_WITHOUT_CANARY`. `moai constitution amend` 는 5겹 게이트(FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight)를 거치고 Frozen 항목에 `--evidence` 를 요구 (`internal/cli/constitution.go` 474-500행) |

### A.6 선행 SPEC 관계

| SPEC | 상태 | 이 SPEC과의 관계 |
|---|---|---|
| SPEC-V3R5-LATE-BRANCH-001 | completed | late-branch 절차의 기원. REQ-LB-001~006이 4단계 절차와 `git reset --hard origin/main` 종결을 규정. EXCL-LB-004가 "late-branch는 단일 main checkout 안에서 동작"이라고 워크트리 확장을 제외했다. OD-1 = 선택지 1은 REQ-LB-003·005·006을 대체하고 EXCL-LB-004를 뒤집는다 |
| SPEC-WORKTREE-BRANCH-GUARD-001 | completed | primary checkout 가드. §A가 `manager-git.md` Phase D를 primary checkout에서 정당하게 실행되는 예외로 인용했고, REQ-WBG-011(spec.md 210행)은 manager-git 신원 면제의 근거를 "`manager-git.md` Phase D is the ONLY place the project authorizes branch-state changes in the primary checkout" 로 적었다. 면제 코드는 `internal/hook/branch_guard.go:455` `return input.AgentType == "manager-git"`. 재설계로 Phase D가 사라지면 이 근거가 사라진다(§E.2) |
| SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 | completed | 가드 훅을 기본 꺼짐으로 전환. "The guard's intent (shared-checkout multi-session protection) does not apply to single-developer repos" 문장의 기원 |
| SPEC-MERGE-METHOD-CONFIG-001 | completed | `merge_method` 도입. REQ-MMC-007/008이 `delivery.md`·`manager-git.md` 의 `--squash` 고정 제거를 요구, REQ-MMC-009가 기본값에서 명령이 바이트 동일할 것을 요구 |
| SPEC-GITFLOW-DOCTRINE-ALIGN-001 | in-progress | 겹치지 않음. 대상이 로컬 전용 문서 3종이고 REQ-GDA-006이 `internal/template/templates/**` 수정을 명시적으로 막는다 |

워크트리 문맥 기본 병합 문구는 SPEC에서 나오지 않았다. `git log -S` 추적 결과 커밋 `b7447cb90`(2026-04-03, `sync.md`)에서 들어왔고, SPEC-V3R4-WORKFLOW-SPLIT-001 커밋 `980ccdc56` 이 `delivery.md` 로 옮겼다(분할 이동). 추적한 문구는 `delivery.md:349` 의 "Auto-merge is now the default for worktree contexts" 이다.

---

## §B 요구사항 (GEARS)

요구사항 식별자는 `REQ-GDP-NNN`. 철회된 요구사항은 번호를 유지하고 제목에 **[WITHDRAWN]** 과 사유를 적는다.

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
범위 파일의 어떤 실행 예시도 `--squash` 를 고정한 `gh pr merge` 명령을 실행할 명령으로 제시해서는 안 된다. `manager-git.md:32` 의 기본값 설명 문장은 남아야 한다. §C.2 흐름의 PR 단계 예시는 `--<merge_method>` 로 적혀야 한다.

### B.3 SX-R04 — auto-merge 기본값 단일 기준 (OD-2 = B)

**REQ-GDP-006** (Event-driven)
When sync 단계가 PR 병합 여부를 판단할 때, `delivery.md` 의 Auto-Merge 절과 `doc-execution.md` 의 워크트리 문맥 감지 절은 `manager-git.md` 의 옵트인(`--auto-merge` 플래그와 전원 승인)을 유일한 기준으로 따라야 하고, 워크트리 문맥이라는 이유만으로 병합하는 기본값을 적어서는 안 되며, 기준이 `manager-git.md` 임을 이름으로 밝혀야 한다. `manager-git.md` 의 옵트인 문장(148행, 166행)은 남아야 한다.

### B.4 AC-01 — 불변식

**REQ-GDP-007** (Ubiquitous)
변경 후 §A.2 표의 각 안내 위치는 primary checkout에서 브랜치 상태를 바꾸는 명령을 실행하라고 지시해서는 안 된다. 같은 명령을 런처로 진입한 워크트리 안에서 실행하라는 안내는 이 요구사항이 제한하지 않는다.

**REQ-GDP-008** (Unwanted)
Late-branch 절차의 명령 예시만 지우고 대체 절차를 남기지 않는 변경은 해서는 안 된다. `spec-assembly.md` 와 `spec-workflow.md` 는 각각 `manager-git.md` 의 절차 절을 형식을 갖춘 참조로 계속 가리켜야 하며, 그 참조의 대상 제목은 실제로 존재해야 한다.

### B.5 AC-01 — 재설계 흐름 (OD-1 = 선택지 1)

**REQ-GDP-009** (Ubiquitous)
`spec-workflow.md` 의 Step 1·Step 4 계약, `manager-git.md` 절차, `spec-assembly.md` 사전 점검, `delivery.md` 복귀 단계는 §C.2 흐름을 같은 진입·PR·종결 명령으로 서술해야 하며, 절차가 쓰는 PR 브랜치 접두의 집합은 `spec-workflow.md` 와 `manager-git.md` 에서 같아야 한다.

**REQ-GDP-010** (Ubiquitous)
Late-branch 절차는 `moai cc -w <name>` 또는 `EnterWorktree(<name>)` 로 진입한 워크트리 안에서 커밋을 쌓아야 하고, 맨손 `git worktree add` 를 지시해서는 안 되며, §C.3의 블록 밖 네 자리는 거기 적힌 처리를 따라야 한다. `spec-workflow.md` 의 등록된 Frozen 절을 바꾸는 변경은 Frozen 구역 수정으로 progress 기록에 남기고 Implementation Kickoff Approval을 거쳐야 하며, 변경 뒤 헌법 검증(`moai constitution validate`)은 해당 항목에 `DRIFT` 를 보고해서는 안 된다.

**REQ-GDP-011** — **[WITHDRAWN 2026-09-10 — OD-1 = 선택지 1 채택으로 선택지 2(조건부 유지) 경로 철회]**
원래 내용: 조건부 유지로 결정되면 조건 문장을 금지문과 안내 위치에 같은 문구로 싣는다. 번호는 추적을 위해 유지한다.

**REQ-GDP-012** — **[WITHDRAWN 2026-09-10 — OD-1 = 선택지 1 채택으로 선택지 3(`main_late_branch` 은퇴) 경로 철회]**
원래 내용: 은퇴로 결정되면 은퇴 사실을 명시하고 `auto_enabled: false` 동작을 새로 정의하며 `--issue` opt-in 기본값을 유지한다. 번호는 추적을 위해 유지한다.

### B.6 사본·생성물·중립성

**REQ-GDP-013** (Ubiquitous)
변경한 줄은 로컬 사본(`.claude/…`)과 템플릿 사본(`internal/template/templates/.claude/…`)에서 같아야 한다. `delivery.md` 두 사본의 의도된 차이(통합 워크트리 경로 표기 275·278행, 꼬리말 479-480행)와 `doc-execution.md` 두 사본의 의도된 차이(로컬 사본에만 있는 138-143행)는 그대로 남아야 한다.

**REQ-GDP-014** (Ubiquitous)
템플릿 `manager-git.md` 가 바뀌면 `internal/template/templates/.codex/agents/moai/manager-git.toml` 은 `make agents-emit` 으로 재생성되어 같은 카드에 커밋되어야 하며, 손으로 편집해서는 안 되고, `make agents-emit-check` 가 exit 0 이어야 한다.

**REQ-GDP-015** (Unwanted)
`internal/template/templates/` 아래에 추가되는 문구는 SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 16개 지원 프로그래밍 언어 중 특정 언어로 치우친 표현을 담아서는 안 된다.

---

## §C 결정 기록

### C.1 결정

| 결정 | 선택 | 날짜 | 결정자 | 이력 |
|---|---|---|---|---|
| OD-1 — Late-branch 절차의 처리 | **선택지 1: 런처 워크트리 흐름으로 재설계** | 2026-09-10 | 운영자(리드 경유) | — |
| OD-2 — 워크트리 문맥 auto-merge 기본값의 단일 기준 | **선택지 B: `manager-git.md` 옵트인(`--auto-merge` 와 전원 승인)** | 2026-09-10 | 운영자(리드 경유) | 처음에는 선택지 C(설정 키 `workflow.worktree.auto_merge`)가 선택됐다. 그 키가 사용처 분류 파일에 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` 으로 기록돼 있다는 사실(`internal/config/testdata/shipped_key_inventory.yaml:2940-2943`)이 제시된 뒤 선택지 B로 바뀌었다. 이 SPEC은 그 키를 건드리지 않는다 |

결정은 plan 산출물의 방향을 정할 뿐 Implementation Kickoff Approval을 대신하지 않는다.

### C.2 재설계 흐름 — 진입·PR·종결 단계 (OD-1 = 선택지 1)

primary checkout의 브랜치 상태는 흐름 전체에서 한 번도 바뀌지 않는다. 모든 브랜치 조작은 런처로 진입한 워크트리 안에서 일어난다.

**진입**

| 단계 | 명령·동작 | 설명 |
|---|---|---|
| E1 | 새 세션은 `moai cc -w <name>`, 현재 세션은 `EnterWorktree(<name>)` | `.claude/worktrees/<name>/` 워크트리에 들어간다. 맨손 `git worktree add` 는 쓰지 않는다 |
| E2 | (명령 없음) | 워크트리는 원격 기본 브랜치에서 잘린다(§A.5 워크트리 기준 브랜치). primary checkout의 로컬 `main` 을 먼저 갱신하지 않는다 |
| E3 | 워크트리 안에서 커밋 | `branch_creation.auto_enabled: false` 일 때 `/moai plan` 은 브랜치를 새로 만들지 않고, 커밋은 워크트리가 만든 브랜치에 쌓인다 |

**PR (구 Phase C)**

| 단계 | 명령·동작 | 설명 |
|---|---|---|
| C1 | 워크트리 안에서 `git branch -m <pr-branch>` | PR 시점에 브랜치 이름을 붙인다. 워크트리 안의 브랜치 이름 변경은 primary checkout 금지 대상이 아니다. `<pr-branch>` 접두는 §C.5 B1 확정 뒤 정한다 |
| C2 | `git push -u origin <pr-branch>` → `gh pr create --base {main_branch}` | 푸시와 PR 생성 |
| C3 | 병합 조건 충족 시 `gh pr merge <PR> --<merge_method> --delete-branch` | 병합 방식은 `git_strategy.<mode>.merge_method`, 자동 병합 여부는 OD-2 = B(`manager-git.md` 옵트인) |

**종결 (구 Phase D)**

| 단계 | 명령·동작 | 설명 |
|---|---|---|
| X1 | `gh pr view <PR> --json state` 가 `MERGED` | 원격 병합 착지 확인. 착지 전에는 X3를 하지 않는다 |
| X2 | `ExitWorktree` | primary checkout으로 돌아간다 |
| X3 | 세션 종료 keep/remove 질문, 또는 `git worktree unlock <path>` + `git worktree remove <path>` | 워크트리 폐기 |
| X4 | (primary checkout에서 브랜치 명령 없음; 필요하면 `git fetch` 만) | `git checkout main`·`git reset --hard origin/main`·`git pull` 을 하지 않는다. primary checkout의 `main` 은 워크트리 커밋을 받은 적이 없으므로 squash 병합 뒤 어긋날 이력이 없고, 다음 작업은 E1의 새 워크트리에서 원격 기본 브랜치부터 시작한다 |

**구 절차에서 사라지는 것**: Phase A의 `git checkout main && git pull origin main`, Phase C의 `git switch -c`, Phase D의 `git checkout main` · `git reset --hard origin/main` · `git pull origin main`, 그리고 Phase D 누락 시의 복구 명령과 "un-squashed history" 실패 양상 서술. push 거부 시 복구(구 125행)는 "워크트리 안에서 C1부터 진행" 으로 바뀐다.

### C.3 흐름이 스스로 답하지 않는 네 자리의 처리

| 자리 | 현재 | 처리 | 분류 장부 분류 |
|---|---|---|---|
| `manager-git.md:42` Checkpoint Rollback | `git reset --hard [checkpoint-tag]` | 되돌리기는 작업을 담은 워크트리 안에서만, 사용자 명시 확인 뒤 실행한다고 적는다(`coding-standards.md` § Bash Risk-Amplifier Doctrine의 파괴 명령 확인 규칙). primary checkout에서는 실행하지 않는다 | `worktree-internal` |
| `manager-git.md:160` Synchronization | `git fetch origin` → `git pull origin [branch]` | `git fetch origin` 은 어느 트리에서나, `git pull` 은 해당 브랜치를 담은 워크트리 안에서만 실행한다고 적는다 | `worktree-internal` |
| `delivery.md:323-324` Step 3.3.5 | `git checkout {main_branch} && git pull origin {main_branch}` 등 | 워크트리에 있는 PR 브랜치는 `ExitWorktree` 로 떠나고, primary checkout은 전환하거나 pull 하지 않는다고 적는다. 기존 `WT-*` 문장은 유지. primary checkout에 체크아웃된 브랜치에 대해서는 전환·pull 명령을 두지 않는다(§C.5 B2) | 명령이 남지 않음 |
| `delivery.md:328` 로컬 브랜치 정리 | `git branch -d <branch>` | 명령 예시를 지우고, 워크트리 폐기가 브랜치를 지우지 않으며 로컬 브랜치 삭제를 원하면 primary checkout이 아닌 워크트리 안에서 한다고 적는다 | 명령이 남지 않음 |

### C.4 Frozen 구역 수정 절차 (측정한 사실)

- `spec-workflow.md` 49행 `[ZONE:Frozen] [HARD] Step ordering rules` 의 Step 1 문장은 레지스트리 `CONST-V3R5-027` 로 등록돼 있고, 선택지 1은 "main checkout에서 실행, 워크트리 없음" 을 바꾼다. 65행 안티패턴("Creating an L2/L3 worktree for plan")도 함께 바뀐다. Step 4 문장(`CONST-V3R5-028`)은 PR 수(§C.5 B1)에 따라 바뀔 수 있다.
- 검증기는 등록된 clause 가 원본에 부분 문자열로 남지 않으면 `DRIFT` 를 보고한다. 따라서 문장을 바꾸면 **Implementation Kickoff Approval 외에** 레지스트리 항목 갱신이 필요하다. 레지스트리가 적은 경로는 둘이다: `moai constitution amend --rule --before --after --evidence`(5겹 게이트, Frozen은 `--evidence` 필수), 또는 레지스트리의 "Retiring an Entry" 절차(clause 앞에 `[SUPERSEDED by <replacement>]` 를 붙이고 새 항목을 등록). 레지스트리는 로컬·템플릿 두 사본이다.
- 기준선(plan 작성 시점, 설치 빌드 `v3.2.0-rc.5` `84fa4ece4`로 측정): `moai constitution validate --format json` → `status: ok`, `drift_count: 0`, `retired_count: 4`. 두 clause 문자열은 `spec-workflow.md` 에 각각 1회 들어 있다.
- 어느 경로를 쓰는지와 레지스트리 파일을 범위에 넣는지는 §C.5 B3.

### C.5 run-phase 착수 전에 정해야 할 설계 항목

결정이 정하지 않은 항목이다. 이 SPEC의 수용 기준은 어느 쪽으로 정해져도 판정할 수 있게 쓰였고, M4 편집은 해당 항목이 정해진 뒤 시작한다.

| ID | 항목 | 측정한 사실 | 영향 |
|---|---|---|---|
| B1 | 한 SPEC이 PR을 몇 개 내는가 — 하나(`feat/SPEC-*`)인가, 단계마다(plan/run/sync)인가 | `manager-git.md` 110-114행·`spec-assembly.md` 336행은 `feat/SPEC-*` PR 하나. `spec-workflow.md` 50행은 `plan/SPEC-XXX`, 53행은 "run-PR and sync-PR" 병합 뒤 종결, Route B 표(42-45행)는 단계별 PR. 등록된 Frozen 절 `CONST-V3R5-028` 이 "BOTH run AND sync PRs" 를 담고 있다 | C1의 `<pr-branch>` 접두, 워크트리를 PR마다 새로 만드는지, 66행 안티패턴, `CONST-V3R5-028` 변경 여부 |
| B2 | primary checkout에 체크아웃된 PR 브랜치 경로가 남는가 | `spec-workflow.md` 43·51·52행과 `delivery.md` Step 3.2 github-flow는 "main checkout의 feature 브랜치" 를 기본 경로로 둔다. 이 경로의 브랜치 생성 자체는 검출식에 걸리는 명령 문구가 아니라 범위 밖이다 | §C.3의 Step 3.3.5 처리가 이 경로 사용자에게서 "기준 브랜치로 복귀" 편의를 없앤다 |
| B3 | Frozen 절 변경의 레지스트리 경로와 파일 범위 | §C.4. 레지스트리(`zone-registry.md` 로컬·템플릿)는 리드가 정한 범위 파일 목록에 없다 | 레지스트리 편집을 범위에 넣지 않으면 REQ-GDP-010의 검증 조건(`DRIFT` 없음)을 만족할 수 없다 |

### C.6 검토한 선택지 (결정 당시 자료)

결정 전에 운영자에게 전달한 표를 기록으로 남긴다. 줄번호는 `b412f8a33` 기준이며 "L·T" 는 로컬·템플릿 두 사본이다.

#### OD-1 — Late-branch 절차의 처리

| 선택지 | 영향 파일 (로컬·템플릿·생성물) | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 (키 · 현재 기본값 · 변화) | 뒤집거나 확장하는 선행 SPEC |
|---|---|---|---|---|
| **1. 런처 워크트리 흐름으로 재설계** (채택) | `manager-git.md` L·T 88-129(절차 블록)·137(옵션 행) + 재생성 `.codex/agents/moai/manager-git.toml`; `spec-workflow.md` L·T 49-62(`[ZONE:Frozen]` Step 1 전제·Step 4 종결)·65(plan 단계 워크트리 금지 안티패턴); `spec-assembly.md` L·T 332-340; `delivery.md` L·T 319-326. 이름만 인용: 템플릿 `CLAUDE.md:75`, `agent-authoring.md:138`, `delegation.yaml:74`, `generic-patterns-guide.md`(템플릿 195행~, 로컬 207행~). 이 선택지가 스스로 답하지 않는 자리: `manager-git.md:42`·`160`, `delivery.md:323-324`·`328` | 부분. "브랜치를 PR 시점에 만든다"는 결과는 유지되고, 커밋을 쌓는 장소가 primary checkout의 `main` 에서 런처로 진입한 워크트리로 바뀐다. 기존 Phase A-D를 primary checkout에서 따르던 사용자는 다른 절차를 받는다. Step 1 "plan은 main checkout에서, 워크트리 없이" 규칙과 부딪혀 Frozen 구역 수정이 따라온다 | `git_strategy.{manual,personal,team}.branch_creation.auto_enabled` — 현재 `false`(템플릿 28·61·100, Go `defaults.go` 739·752·768). 값은 그대로, `false` 가 촉발하는 동작이 바뀐다 | SPEC-V3R5-LATE-BRANCH-001 REQ-LB-003·005·006을 대체, EXCL-LB-004를 뒤집음. SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-011이 manager-git 신원 면제의 근거로 적은 문장(spec.md 210행)이 사라지고, `internal/hook/branch_guard.go:455` 의 면제는 적힌 근거 없이 남는다 |
| **2. 유지하되 명시적 조건으로 범위 한정** (미채택) | 선택지 1의 네 지침 파일 L·T + 재생성 `.toml`; 루트 `AGENTS.md:64`, 템플릿 `AGENTS.md:64`, `main-checkout-branch-guard.md` L·T 16 | 철회 없음. 모든 배포 사용자에게 적용되는 금지 원칙이 조건부로 바뀐다. `AGENTS.md` 는 병합 체인이 넘치면 꼬리부터 조용히 잘리는 문서다 | 조건 후보 `workflow.branch_guard.enabled` — Go 기본값 `false`(`defaults.go:893-895`), 템플릿 `workflow.yaml` 에는 키 없음 | SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001의 전제를 원칙 층으로 확장, SPEC-WORKTREE-BRANCH-GUARD-001의 무조건 원칙을 좁힘. REQ-WBG-011 근거 유지 |
| **3. `main_late_branch` 은퇴** (미채택) | `manager-git.md` L·T 88-129·137 + 재생성 `.toml`; `spec-workflow.md` L·T 50·53-62; `spec-assembly.md` L·T 332-340; 템플릿 `git-strategy.yaml.tmpl` 82-84 주석; 이름 인용처; `--issue` 기본값을 "late-branch opt-in policy"로 부르는 `SKILL.md` 124·227, `workflows/moai.md` 45·233, `spec-assembly.md` 258 | 예. Phase A-D 4단계 절차가 사라진다. 지울 설정 키는 없다 | `auto_enabled: false` 일 때의 동작을 새로 정의해야 함 | SPEC-V3R5-LATE-BRANCH-001 REQ-LB-001~006을 뒤집음. REQ-WBG-011 근거가 사라짐 |
| ~~예시 명령만 삭제~~ (제외) | — | `auto_enabled` 기본값이 세 모드 모두 `false` 라 `/moai plan` 은 계속 브랜치 생성을 건너뛰고 절차를 참조하라고 안내하는데, 그 절차가 사라진 채 남는다. REQ-GDP-008이 금지한다 | — | — |

#### OD-2 — 워크트리 문맥 auto-merge 기본값의 단일 기준

| 선택지 | 영향 파일 (로컬·템플릿·생성물) | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 | 뒤집거나 확장하는 선행 SPEC |
|---|---|---|---|---|
| **A. `delivery.md` 워크트리 기본값** (미채택) | `manager-git.md` L·T 148·166 + 재생성 `.toml`; `delivery.md` L·T 335-349 유지; `sync/doc-execution.md` L·T 36 이미 일치 | team 모드이면서 워크트리 문맥이면 `--auto-merge` 플래그 없이 병합 | 플래그로 결정 | `b7447cb90` 에서 들어온 동작을 확정 |
| **B. `manager-git.md` 옵트인** (채택) | `delivery.md` L·T 335-338(트리거)·348-349(플래그 설명); `sync/doc-execution.md` L·T 34-36; `manager-git.md` 유지(재생성 불필요) | 워크트리 문맥의 sync가 기본으로 병합하지 않는다. 지금 기본 병합에 기대는 사용자는 플래그를 넘겨야 한다. `--merge` 폐기 경고 문구의 논리가 뒤집힌다 | 키 없음. 워크트리 문맥의 동작 기본값이 "병합"에서 "병합 안 함"으로 바뀜 | `b7447cb90` 에서 들어온 동작을 뒤집음(SPEC 기원 없음) |
| **C. 설정 키 `workflow.worktree.auto_merge`** (처음 선택 후 B로 교체) | `delivery.md`, `doc-execution.md`, `manager-git.md` + 사용처 분류 파일 | 기본값 `false` 라 워크트리 문맥의 기본 병합이 멈춤 | 키 값 `false` 유지, 첫 소비자가 생김. 그 키는 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` 으로 기록돼 있음 | SPEC-WORKTREE-ENTRY-STRATEGY-001의 기본값 전환을 동작으로 확장 |

---

## §D 제외 범위 (Out of Scope)

### Out of Scope — 채택되지 않은 선택지의 경로

- OD-1 선택지 2의 조건부 원칙: 루트·템플릿 `AGENTS.md` §2 금지문과 `main-checkout-branch-guard.md` 원칙은 바꾸지 않는다.
- OD-1 선택지 3의 은퇴: `main_late_branch` 옵션과 `--issue` opt-in 정책 문구를 없애지 않는다.
- OD-2 선택지 C: 설정 키 `workflow.worktree.auto_merge` 와 사용처 분류 파일은 건드리지 않는다.

### Out of Scope — 코드와 설정 값

- Go 코드(`internal/config`, `internal/hook`, `internal/github`, `internal/constitution`) 변경. 특히 `internal/hook/branch_guard.go:455` 의 manager-git 신원 면제는 그대로 둔다(§E.2, 후속 카드 후보).
- `git-strategy.yaml.tmpl`·`workflow.yaml` 의 키 값 변경.
- `main-checkout-branch-guard` 훅의 패턴·면제 로직 변경.

### Out of Scope — 카드 범위 밖에서 관측한 형제 (리드 판단 대상)

리드가 정한 범위를 넓히지 않으려고 기록만 한다. 같은 결함 계열이다.

- `manager-git.md:82` — `auto_branch: true` 일 때 "checkout from main_branch".
- `manager-git.md:171` — PR Auto-Merge 5단계 "Checkout main, pull, delete local branch".
- `delivery.md:356` — Auto-Merge Execution 4단계 "Checkout target branch, fetch latest".
- `generic-patterns-guide.md`(템플릿 195행~, 로컬 207행~) Late-Branch Phase D Recovery — `git stash push`, `git reset --keep`, `git checkout stash@{0} --` 를 워크트리 한정 없이 안내한다. 재설계로 Phase D가 사라지면 이 절의 전제도 사라진다.
- `spec-workflow.md` 43·51·52행과 `delivery.md` Step 3.2의 "main checkout의 feature 브랜치" 경로(§C.5 B2).

### Out of Scope — 금지 목록과 경고표

- 다음 위치의 브랜치 변경 명령은 금지 목록·경고표·인용 예시·권한 항목·워크트리 내부 안내라서 유지한다: 템플릿 `AGENTS.md` 64-66, `main-checkout-branch-guard.md` 20-24와 detail 파일, `coding-standards.md:141`, `moai-ref-git-workflow/SKILL.md` 144-145, `verification-claim-integrity-detail.md:161`, `settings.json.tmpl:564`, `moai-workflow-worktree/modules/troubleshooting.md` 176·179.

---

## §E 미검증과 잔여 위험

### E.1 미검증 (Gaps)

- 파괴적 명령을 실제로 실행해 공유 checkout이 바뀌는 것을 재현하지 않았다. AC-01 판정은 실행 안내와 배포 금지문의 대조에 근거한다.
- `git fetch`/`git rev-list` 경합을 실제로 일으키지 않았다. 감사 보고서도 AC-11 신뢰도를 medium으로 매겼다.
- 재설계 흐름(§C.2)을 실제 저장소에서 끝까지 실행해 보지 않았다.
- 헌법 검증 기준선은 설치 빌드(`84fa4ece4`)로 쟀고, 이 트리에서 빌드한 바이너리로 재지 않았다. 설치 빌드가 이 트리의 조상인지도 확인하지 않았다.
- `moai constitution amend` 의 Canary 층이 요구하는 SPEC 수와 FrozenGuard가 Frozen 항목의 문구 변경을 받아들이는 조건을 코드로 끝까지 읽지 않았다.
- `manager-git.md`·`agent-common-protocol.md`·`doc-execution.md` 는 `rule_template_mirror_test.go` 의 바이트 동일 대상이 아니다. 세 파일의 사본 일치는 이 SPEC의 diff 기준으로만 확인된다.
- `TestTemplateNoInternalContentLeak` 의 날짜·SHA 검출 분류가 에이전트·규칙·스킬 경로까지 적용되는지는 읽지 않았다. AC-GDP-015는 이 테스트에 기대지 않고 추가 줄을 직접 검사한다.
- 워크트리 기본 병합 문구의 `git log -S` 추적은 `delivery.md:349` 문장에만 수행했다.

### E.2 잔여 위험

- **REQ-WBG-011 근거 소실 (운영자 지시로 기록).** 재설계로 Phase D가 사라지면 SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-011이 manager-git 신원 면제의 근거로 적은 문장("`manager-git.md` Phase D is the ONLY place the project authorizes branch-state changes in the primary checkout", spec.md 210행)이 성립하지 않는다. `internal/hook/branch_guard.go:455` 의 면제(`return input.AgentType == "manager-git"`)는 적힌 근거 없이 남는다. 훅 변경은 이 SPEC 범위 밖이며, **후속 카드 후보**로 기록한다(발행 여부는 리드 결정).
- 검출식은 줄 단위다. 여러 줄에 걸쳐 흩어진 위반이나 검출식이 예상하지 않은 표현은 통과할 수 있다. AC-GDP-001·006·007·009는 읽기 단계로 이 틈을 좁히지만 오독 가능성이 남는다.
- §A와 §C의 줄번호는 범위 파일이 `b412f8a33` 과 바이트 동일하다는 전제에 기댄다. run-phase 전에 develop을 흡수하면 범위 파일 변경 여부를 다시 확인해야 한다.
- OD-2 = B는 워크트리 문맥의 기본 병합을 없앤다. 지금 이 기본값에 기대는 사용자는 `--auto-merge` 를 넘기지 않으면 병합되지 않는 PR을 보게 된다.

---

## §F 수용 기준

`acceptance.md` 가 기준(AC-GDP-001~016; 011·012는 철회)과 검증 명령, 양성 대조, 감사 뮤턴트 재실행 결과를 담는다.
