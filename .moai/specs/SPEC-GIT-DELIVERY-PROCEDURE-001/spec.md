---
id: SPEC-GIT-DELIVERY-PROCEDURE-001
title: "배포 지침의 git 전달 절차 결함 3건 수리 — primary checkout 브랜치 변경 안내, fetch 병렬 배치, auto-merge 기본값 충돌"
version: "0.1.3"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/agents/moai/manager-git.md, .claude/rules/moai/workflow/spec-workflow.md, .claude/rules/moai/core/agent-common-protocol.md, .claude/rules/moai/core/zone-registry.md, .claude/skills/moai/workflows/plan/spec-assembly.md, .claude/skills/moai/workflows/sync/delivery.md, .claude/skills/moai/workflows/sync/doc-execution.md (+ internal/template/templates mirrors)"
lifecycle: spec-anchored
tags: "git, manager-git, late-branch, launcher-worktree, single-pr, auto-merge, merge-method, sync-check, constitution-amend, template-mirror, instruction-audit"
era: V3R6
tier: M
related_specs: [SPEC-V3R5-LATE-BRANCH-001, SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-MERGE-METHOD-CONFIG-001]
---

# SPEC-GIT-DELIVERY-PROCEDURE-001 — 배포 지침의 git 전달 절차 결함 3건 수리

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-10 | manager-spec | 최초 작성. 칸반 카드 t622(지침 감사 G1). 결함 3건(AC-01·AC-11·SX-R04)과 카드 밖 형제를 리드가 정한 범위로 묶었다. 재현 근거는 `.moai/reports/t622/repro.md`(측정 트리 `c352330d3`), 인용 줄번호는 이 SPEC 작성 트리 `b412f8a33`에서 다시 읽어 확인했다. 설계 판단 2건(OD-1·OD-2)은 결정하지 않고 §C에 표로 올렸다. |
| 0.1.1 | 2026-09-10 | manager-spec | plan-audit 1회차(`.moai/reports/t622/plan-audit.md`, FAIL 0.67) 반영. 필수 결함 D1~D4와 결정 표 보완 D11·D12, 선택 결함 D5~D10·D13~D19. REQ 15개·AC 16개 유지. |
| 0.1.2 | 2026-09-10 | manager-spec | 운영자 결정 OD-1 = 선택지 1(런처 워크트리 흐름), OD-2 = 선택지 B(`manager-git.md` 옵트인) 반영. §C를 결정 기록으로 바꾸고 흐름의 진입·PR·종결 단계와 블록 밖 네 자리의 처리를 정의했다. REQ-GDP-011·012와 AC-GDP-011·012 철회. `sync/doc-execution.md` 34-36행을 범위에 넣었다. plan-audit 2회차(`.moai/reports/t622/plan-audit-iter2.md`, FAIL 0.75) 결함 N1~N6 반영. 설계 항목 B1~B3을 미결로 기록. |
| 0.1.3 | 2026-09-10 | manager-spec | 운영자 결정 B1~B3(리드 경유) 반영: B1 = SPEC당 PR 하나(`feat/SPEC-*`), B2 = primary checkout의 feature 브랜치 경로도 워크트리 흐름으로 이동, B3 = `moai constitution amend` 5겹 게이트로 `CONST-V3R5-027`·`028` 개정. `zone-registry.md` 로컬·템플릿과 `delivery.md` Step 3.1·3.2 github-flow를 범위에 넣었다. REQ-GDP-016~023, AC-GDP-017~024 추가(REQ-GDP-009·010, AC-GDP-009·010 개편). amend 명령의 소스를 읽은 결과 적용 단계가 스텁이라 개정이 기록되지 않는다는 사실(§C.4)과 worktree-integration.md 형제 모순을 남은 차단 항목으로 기록했다. REQ 23개(활성 21·철회 2)·AC 24개(활성 22·철회 2)로 Tier M 상한(16/16)을 넘는다 — `tier:` 는 바꾸지 않았다. |

---

## §A 배경과 문제

### A.1 결함 요약

| 결함 | 내용 | 처리 |
|---|---|---|
| AC-01 | 배포 지침이 스스로 금지한 primary checkout 브랜치 변경을 같은 배포물 안에서 실행하라고 안내한다 | OD-1 = 선택지 1(런처 워크트리 흐름), B1(SPEC당 PR 하나), B2(primary feature 브랜치 경로도 이동), B3(Frozen 절 개정) — §C |
| AC-11 | `git fetch` 와 그 결과를 읽는 `git rev-list` 를 서로 독립인 병렬 배치로 지시한다 | 기계적 수리 (M1) |
| SX-R04 | 워크트리 문맥의 auto-merge 기본값이 두 지침에서 반대로 적혀 있고, 병합 명령에 `--squash` 가 박혀 있다 | 병합 방식 치환(M2), 기본값은 OD-2 = 선택지 B(M3) |

### A.2 AC-01 — 배포물 내부의 모순

전제는 이 트리의 문구 대조로 확인했다.

| 근거 | 위치 | 내용 |
|---|---|---|
| 무조건 금지 | `internal/template/templates/AGENTS.md:64` (루트 `AGENTS.md:64` 도 같은 문장) | "Never change branch state in the primary checkout." 조건 없음. 금지 집합은 `git checkout`·`git switch`·`git branch`·`git reset --hard`·`git stash`·현재 브랜치 위의 `git rebase`/`git merge` |
| 무조건 금지 | `internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md:16` | `[HARD]` MUST NOT, 조건 없음 |
| 조건이 붙은 유일한 자리 | 같은 파일 `:85-86` | "the shared-checkout hazard does not apply to single-developer repos" — 기계 차단 훅의 **기본값**(`false`)에만 붙은 조건 |

AC-01은 배포되는 지침끼리의 모순이다. `git pull` 은 현재 브랜치에 병합이나 리베이스를 수행하므로 같은 금지 집합에 든다. 금지는 primary checkout에만 걸린다 — 런처로 진입한 워크트리 안의 같은 명령은 금지 대상이 아니다.

**A.2.1 명령 형태의 실행 안내** (`b412f8a33` 기준, 로컬·템플릿 같은 줄번호; 검출식은 `acceptance.md` AC-GDP-007)

| 파일 | 줄 | 안내 |
|---|---|---|
| `.claude/agents/moai/manager-git.md` | 42 | Checkpoint Rollback `git reset --hard [checkpoint-tag]` |
| 〃 | 96 | Late-branch Phase A `git checkout main && git pull origin main` |
| 〃 | 110 | Phase C `git switch -c feat/SPEC-XXX` |
| 〃 | 119, 121, 122 | Phase D `git checkout main` · `git reset --hard origin/main` · `git pull origin main` |
| 〃 | 125 | push 거부 시 복구 `git switch -c feat/SPEC-*` |
| 〃 | 127 | Phase D 누락 시 복구 `git fetch origin && git reset --hard origin/main` |
| 〃 | 137 | Personal Mode 옵션 `main_late_branch` 설명 |
| 〃 | 160 | Synchronization 절 `git fetch origin` → `git pull origin [branch]` |
| `.claude/rules/moai/workflow/spec-workflow.md` | 50 | Route B Step 1 late-branch 전제, `git switch -c plan/SPEC-XXX` |
| 〃 | 56, 58, 59 | Step 4 late-branch 종결 `git checkout main` · `git reset --hard origin/main` · `git pull origin main` |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | 336 | late-branch 사전 점검, 수동 `git switch -c feat/SPEC-*` |
| `.claude/skills/moai/workflows/sync/delivery.md` | 323, 324 | Step 3.3.5 기준 브랜치 복귀 `git checkout {main_branch} && git pull origin {main_branch}` 등 |
| 〃 | 328 | 로컬 브랜치 정리 `git branch -d <branch>` 안내 |

같은 검출식에 걸리지만 실행 안내가 아닌 줄: `spec-workflow.md:62`(실패 양상 서술), `delivery.md:276`(통합 워크트리 안의 `git merge --no-ff`).

**A.2.2 서술형 실행 안내** — 명령 문구가 없어 AC-GDP-007 검출식이 보지 못하는 줄 (서술형 검출식은 AC-GDP-021)

| 파일 | 줄 | 안내 |
|---|---|---|
| `spec-workflow.md` | 21 | "Default flow executes all phases on main checkout with a feature branch." |
| 〃 | 42, 43 | Route B 표 Step 1 위치 "main checkout", Step 2 "`feat/SPEC-XXX` branch in main checkout" |
| 〃 | 50, 51, 52 | Step 1 "MUST execute in main checkout on BOTH routes", Step 2·3 "continue on the (same) feature branch in main checkout" |
| 〃 | 166 | Plan Phase `[ZONE:Frozen] [HARD] Execute in main checkout. NO worktree at this step.` (서술형 검출식에는 브랜치 낱말이 없어 걸리지 않음 — AC-GDP-010 기록 대상) |
| 〃 | 192, 286 | Run Phase·Sync Phase `[SHOULD]` 문장의 main checkout feature 브랜치 기본 경로 |
| 〃 | 321, 429 | Phase Transitions의 Route B 실행 위치 |
| `delivery.md` | 238-257 | Step 3.2 github-flow "Feature branch (any branch other than main)" 경로 — 현재 트리가 primary checkout인지 묻지 않고 push·PR |
| 〃 | 321 | Step 3.3.5 "return to the base branch" |

**A.2.3 단계별 PR 안내** — B1과 충돌하는 줄 (검출식은 AC-GDP-017)

| 파일 | 줄 | 안내 |
|---|---|---|
| `spec-workflow.md` | 26 | Route B 정의 "opens a PR per phase" (`[ZONE:Frozen]` 23행 블록 안) |
| 〃 | 42-44, 47 | Route B 표의 `plan/SPEC-XXX`·`feat/SPEC-XXX`·`sync/SPEC-XXX`(`chore/SPEC-XXX-sync`) 단계별 PR, "one squash commit per phase" |
| 〃 | 53, 65, 66 | Step 4 "BOTH run AND sync PRs", 안티패턴의 plan PR 병합 |
| 〃 | 319, 320, 336, 360, 361 | Plan to Run 트리거·전제의 plan PR, skip 정책의 plan-PR, 동시 plan-run 파이프라인 |
| 〃 | 428, 433, 438, 439 | Run to Sync·Sync close·Cleanup의 run PR·sync PR |
| `delivery.md` | 50 | Step 3.1 Route B 동기화 커밋이 `sync/SPEC-XXX` 또는 `chore/SPEC-XXX-sync` 브랜치에 쌓임 |

### A.3 AC-11 — 결과를 읽는 명령과의 병렬 배치

| 파일 | 줄 | 문제 |
|---|---|---|
| `.claude/agents/moai/manager-git.md` | 156 | `git fetch`, `git status`, `git rev-list --count --left-right`, `gh pr checks --json` 를 "independent and read-only" 로 묶어 한 턴 병렬 배치로 지시 |
| `.claude/rules/moai/core/agent-common-protocol.md` | 292, 296, 299 | Pre-Spawn Sync Check 가 "parallel batch" 를 선언하고 `git fetch origin main`(296)과 `git rev-list`(299)를 별도 명령으로 나열 |
| 같은 파일 (정상) | 347 | Pre-Edit Sync Check 는 `git fetch origin main 2>&1; git rev-list ...` 를 한 명령으로 묶음 — 유지 대상 |

`git fetch` 는 원격 추적 ref 를 갱신하고 `git rev-list` 는 그 ref 를 읽으므로 둘은 독립이 아니다. 병렬로 돌면 원격이 앞선 상태(`N 0`)가 `0 0` 으로 보고될 수 있다.

### A.4 SX-R04 — auto-merge 기본값 충돌과 병합 방식 고정

| 파일 | 줄 | 내용 |
|---|---|---|
| `delivery.md` | 335-338 | 트리거: `is_worktree_context == true` 이고 `--no-merge` 가 없으면 병합 |
| `delivery.md` | 343, 355 | `gh pr merge --squash --delete-branch` 고정 |
| `delivery.md` | 348-349 | `--merge` 폐기 경고 "Auto-merge is now the default for worktree contexts." |
| `sync/doc-execution.md` | 34-36 | "This affects auto-merge behavior: worktree contexts default to auto-merge." |
| `manager-git.md` | 148, 166, 170 | Team Mode "Auto-merge: only with the `--auto-merge` flag", "Execute only with `--auto-merge` flag AND all approvals obtained", `--<merge_method>` 해석 |
| `manager-git.md` | 114 | Phase C 예시가 `--squash` 로 고정 |
| `manager-git.md` | 32 | `merge_method` 해석 규칙의 기본값 설명 (유지 대상) |

### A.5 설정·레지스트리 사실 (측정)

| 항목 | 측정 결과 |
|---|---|
| git-strategy 템플릿 경로 | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` |
| `branch_creation.auto_enabled` | 템플릿 세 모드 모두 `false` (28, 61, 100행) · Go 기본값 `internal/config/defaults.go` 739·752·768행 `AutoEnabled: false` |
| `main_late_branch` | 설정 키가 아니다. `manager-git.md:137` 과 생성물 `.codex/agents/moai/manager-git.toml:131` 의 문구에만 나온다 |
| `merge_method` | 템플릿 세 모드 모두 `squash` · Go 기본값 `defaults.go` 738·751·767행 `MergeMethod: "squash"` |
| 워크트리 기준 브랜치 | `worktree-integration.md` § Worktree Base Branch: 네이티브 워크트리는 기본으로 `origin/HEAD`(`worktree.baseRef` 기본값 `fresh`)에서 잘린다 |
| 헌법 레지스트리 | `zone-registry.md` 가 spec-workflow.md를 세 항목으로 등록: `CONST-V3R2-001`("Create comprehensive specification using EARS format.", `#plan-phase`), `CONST-V3R5-027`(Frozen, `canary_gate: true`, "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step"), `CONST-V3R5-028`(Frozen, `canary_gate: true`, "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged"). 로컬·템플릿 사본 바이트 동일(diff exit 0). spec-workflow.md의 `[ZONE:Frozen]` 태그는 23·49·166행 세 곳이며, 23행 블록(Route 정의)과 166행(Plan Phase)은 레지스트리에 항목이 없다 |
| 헌법 검증기 | `internal/constitution/validator.go`: 등록된 clause 가 원본에 공백 정규화 뒤 부분 문자열로 없으면 `DRIFT`. `ZONE_UNREGISTERED` 센티널은 정의돼 있으나 `Validate` 가 그 항목을 만들지 않는다 — 등록되지 않은 `[ZONE:Frozen]` 줄의 변경은 검증기에 드러나지 않는다. 레지스트리 경로가 프로젝트 디렉터리 밖이면 "escapes project dir" 오류로 exit 1, `status: ""` |

### A.6 선행 SPEC 관계

| SPEC | 상태 | 이 SPEC과의 관계 |
|---|---|---|
| SPEC-V3R5-LATE-BRANCH-001 | completed | late-branch 절차의 기원. OD-1 = 선택지 1은 REQ-LB-003·005·006을 대체하고 EXCL-LB-004("late-branch는 단일 main checkout 안에서 동작")를 뒤집는다 |
| SPEC-WORKTREE-BRANCH-GUARD-001 | completed | REQ-WBG-011(spec.md 210행)이 manager-git 신원 면제의 근거를 "`manager-git.md` Phase D is the ONLY place the project authorizes branch-state changes in the primary checkout" 로 적었다. 면제 코드 `internal/hook/branch_guard.go:455`. 재설계로 Phase D가 사라지면 근거가 사라진다(§E.2) |
| SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 | completed | 가드 훅 기본 꺼짐 전환 |
| SPEC-MERGE-METHOD-CONFIG-001 | completed | REQ-MMC-007/008이 `--squash` 고정 제거를 요구, REQ-MMC-009가 기본값에서 명령 바이트 동일을 요구 |
| SPEC-V3R2-CON-002 | (amend 구현의 기원) | `moai constitution amend` 5겹 게이트. 적용 단계의 두 도우미가 스텁으로 남아 있다(§A.7) |

워크트리 문맥 기본 병합 문구는 커밋 `b7447cb90`(2026-04-03, `sync.md`)에서 들어왔고 `980ccdc56` 이 `delivery.md` 로 옮겼다(`git log -S` 추적, `delivery.md:349` 문장).

### A.7 헌법 개정 명령의 실제 동작 (소스 확인 — HEAD와 `b412f8a33` 사이 `internal/constitution`·`internal/cli/constitution.go` 차이 없음)

| 확인 대상 | 위치 | 내용 |
|---|---|---|
| 도움말 | `moai constitution amend --help` (설치 빌드 `84fa4ece4`) | 플래그 `--rule`(필수) · `--before`(필수) · `--after`(필수) · `--evidence`(Frozen 필수) · `--dry-run` |
| `--before` 대조 | `internal/cli/constitution.go:528-529` | 등록된 clause 와 **바이트 동일**하지 않으면 "clause mismatch" 로 중단 |
| 레지스트리 해석 두 갈래 | `constitution.go:143-153` 대 `internal/constitution/pipeline.go:66` | CLI의 `--before` 대조는 `MOAI_CONSTITUTION_REGISTRY` → `CLAUDE_PROJECT_DIR` → 작업 디렉터리 순으로 레지스트리를 찾고, 파이프라인은 `<작업 디렉터리>/.claude/rules/moai/core/zone-registry.md` 를 고정으로 읽는다. 템플릿 사본은 어느 쪽도 건드리지 않는다 |
| dry-run | `constitution.go:470`, `human_oversight.go:32` | `--dry-run` 또는 `MOAI_CONSTITUTION_DRY_RUN=true`. 1~4층을 평가하고 5층은 자동 승인, 파일 변경 없음 |
| 1층 FrozenGuard | `frozen_guard.go:22` | Frozen 항목에 `--evidence` 가 비면 거부. 구역 강등은 실제로 일어나지 않는다(`pipeline.go` `createLogEntry` 의 `ZoneAfter: originalZone`) |
| 2층 Canary | `canary.go:16`, `:18`, `:38` | `^SPEC-[A-Z0-9]+$` 형태의 SPEC 디렉터리(`progress.md` 보유)가 3개 미만이면 사용 불가 — 파이프라인은 사용 불가를 치명으로 보지 않고 계속한다. 이 트리의 해당 디렉터리 수는 0 |
| 3층 ContradictionDetector | `contradiction.go:87`, `:110`, `:124` | `--after` 가 연속 낱말 `MUST NOT` 을 담으면 다른 MUST 항목과 모순으로 차단. `--before` 의 MUST 가 사라지고 MAY/SHOULD/OPTIONAL 이 생기면 차단 |
| 4층 RateLimiter | `rate_limiter.go:10`, `:12`, `:92` | 7일에 3건, 개정 사이 24시간. 기록 파일은 `.moai/research/evolution-log.md` (`pipeline.go:115`) |
| 5층 HumanOversight | `human_oversight.go:44-45` | 터미널 표준입력 Y/N 질문. 비대화형 실행에서는 입력이 없어 오류로 끝난다 |
| 적용 단계 | `pipeline.go:140`, `:190`, `:256-259`, `:264-266` | 승인 뒤 `applyAmendment` 가 원본 파일 → 레지스트리 → 기록 순으로 쓰는데, `updateSourceFile` 과 `updateRegistryClause` 가 둘 다 "not yet implemented" 를 돌려주는 스텁이다. 첫 단계에서 실패하므로 아무 파일도 바뀌지 않는다. `internal/constitution/pipeline_test.go:334-356`, `:406-423` 이 이 스텁 동작을 고정한다 |
| 레지스트리의 변경 절차 | `zone-registry.md` 39-58행 § Retiring an Entry | 레지스트리 안에 "amend" 절은 없다(`grep` 0건). 적힌 변경 절차는 `[SUPERSEDED by …]` 은퇴뿐이다 |

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
`sync/delivery.md` 의 auto-merge 실행 단계는 병합 명령을 활성 모드의 `git_strategy.<mode>.merge_method`(기본 `squash`)로 해석한 `gh pr merge --<merge_method> --delete-branch` 로 지시해야 한다. 기본값에서 실제 실행 명령은 지금과 바이트 동일해야 한다.

**REQ-GDP-005** (Unwanted)
범위 파일의 어떤 실행 예시도 `--squash` 를 고정한 `gh pr merge` 명령을 실행할 명령으로 제시해서는 안 된다. `manager-git.md:32` 의 기본값 설명 문장은 남아야 한다.

### B.3 SX-R04 — auto-merge 기본값 단일 기준 (OD-2 = B)

**REQ-GDP-006** (Event-driven)
When sync 단계가 PR 병합 여부를 판단할 때, `delivery.md` 의 Auto-Merge 절과 `doc-execution.md` 의 워크트리 문맥 감지 절은 `manager-git.md` 의 옵트인(`--auto-merge` 플래그와 전원 승인)을 유일한 기준으로 따라야 하고, 워크트리 문맥이라는 이유만으로 병합하는 기본값을 적어서는 안 되며, 기준이 `manager-git.md` 임을 이름으로 밝혀야 한다. `manager-git.md` 의 옵트인 문장(148행, 166행)은 남아야 한다.

### B.4 AC-01 — 불변식

**REQ-GDP-007** (Ubiquitous)
변경 후 §A.2.1 표의 각 안내 위치는 primary checkout에서 브랜치 상태를 바꾸는 명령을 실행하라고 지시해서는 안 된다. 같은 명령을 런처로 진입한 워크트리 안에서 실행하라는 안내는 이 요구사항이 제한하지 않는다.

**REQ-GDP-008** (Unwanted)
Late-branch 절차의 명령 예시만 지우고 대체 절차를 남기지 않는 변경은 해서는 안 된다. `spec-assembly.md` 와 `spec-workflow.md` 는 각각 `manager-git.md` 의 절차 절을 형식을 갖춘 참조로 계속 가리켜야 하며, 그 참조의 대상 제목은 실제로 존재해야 한다.

### B.5 AC-01 — 재설계 흐름 (OD-1 = 선택지 1, B1, B2)

**REQ-GDP-009** (Ubiquitous)
`spec-workflow.md` 의 Step 규칙, `manager-git.md` 절차, `spec-assembly.md` 사전 점검, `delivery.md` Step 3.3.5는 §C.2 흐름을 같은 진입·PR·종결 명령으로 서술해야 하며, 이 네 곳과 `delivery.md` 가 쓰는 SPEC PR 브랜치 접두는 `feat/SPEC-` 하나여야 한다.

**REQ-GDP-010** (Ubiquitous)
Route B 절차는 `moai cc -w <name>` 또는 `EnterWorktree(<name>)` 로 진입한 워크트리 안에서 커밋을 쌓아야 하고, 맨손 `git worktree add` 를 지시해서는 안 되며, `manager-git.md:42`·`:160` 은 §C.3의 처리를 따라야 한다. `spec-workflow.md` 의 `[ZONE:Frozen]` 줄(등록 여부와 무관하게 23행 블록·49행 블록·166행)을 바꾸는 변경은 Frozen 구역 수정으로 progress 기록에 남기고 Implementation Kickoff Approval을 거쳐야 한다.

**REQ-GDP-011** — **[WITHDRAWN 2026-09-10 — OD-1 = 선택지 1 채택으로 선택지 2(조건부 유지) 경로 철회]**
원래 내용: 조건부 유지로 결정되면 조건 문장을 금지문과 안내 위치에 같은 문구로 싣는다. 번호는 추적을 위해 유지한다.

**REQ-GDP-012** — **[WITHDRAWN 2026-09-10 — OD-1 = 선택지 1 채택으로 선택지 3(`main_late_branch` 은퇴) 경로 철회]**
원래 내용: 은퇴로 결정되면 은퇴 사실을 명시하고 `auto_enabled: false` 동작을 새로 정의한다. 번호는 추적을 위해 유지한다.

**REQ-GDP-016** (Ubiquitous) — B1
Route B는 SPEC 하나에 PR 하나를 내야 한다. `spec-workflow.md`(Route 정의·Route B 표·Step 규칙·안티패턴·Phase Transitions·Plan Audit Gate skip 정책·Cleanup)와 `delivery.md` Step 3.1은 plan PR·run PR·sync PR 같은 단계별 PR과 그 브랜치(`plan/SPEC-`·`sync/SPEC-`·`chore/SPEC-…-sync`)를 지시해서는 안 되며, 동기화 커밋은 같은 SPEC PR 브랜치에 쌓여야 한다.

**REQ-GDP-017** (Ubiquitous) — B2
Route B의 Step 1~3(plan·run·sync)은 런처로 진입한 같은 워크트리 안에서 실행되어야 하며, 범위 지침 파일은 명령 문구가 없는 서술형 문장으로도 primary checkout의 feature 브랜치(또는 SPEC PR 브랜치) 위에서 작업하라고 지시해서는 안 된다. Route A(primary checkout의 `main` 에 직접 커밋·push)는 이 요구사항이 바꾸지 않는다.

**REQ-GDP-018** (Event-driven) — B2
When github-flow 전달이 PR 브랜치를 만나면, `delivery.md` Step 3.2는 그 브랜치가 런처로 진입한 워크트리 안에 있을 때만 push와 PR 생성을 지시해야 하고, primary checkout에서 `main` 이 아닌 브랜치를 만나면 push와 PR 생성을 하지 않고 멈춰 워크트리 진입을 안내하는 보고를 해야 한다. 두 사본의 의도된 차이(275·278·479-480행)는 남아야 한다.

**REQ-GDP-019** (Event-driven) — B2
When PR 생성과 선택적 ready 전환이 끝나면, `delivery.md` Step 3.3.5는 기준 브랜치로 전환하거나 pull하라는 대신 `ExitWorktree` 로 워크트리를 떠나라고 지시해야 하고, 원격 병합 착지 전에는 워크트리를 폐기하지 않는다고 적어야 하며, 어떤 전략에서도 primary checkout의 전환·pull·로컬 브랜치 삭제 명령을 두어서는 안 된다. `WT-*` 문장은 남아야 한다.

### B.6 헌법 레지스트리 개정 (B3)

**REQ-GDP-020** (Ubiquitous)
`CONST-V3R5-027`·`CONST-V3R5-028` 의 문구 변경은 이 트리에서 빌드한 `moai constitution amend` 의 5겹 게이트를 거쳐야 하며, 각 호출은 등록된 현재 clause 와 바이트 동일한 `--before`, §C.4에 적힌 `--after`, 비어 있지 않은 `--evidence` 를 담아야 하고, 근거 문서와 실행한 명령·출력은 `.moai/reports/t622/run/` 아래 추적 경로에 남아야 한다.

**REQ-GDP-021** (Unwanted)
레인은 HumanOversight 층의 승인을 운영자 대신 해서는 안 된다. 표준입력 주입(파이프, here-string, 파일 리다이렉트)으로 승인을 넘겨서는 안 되며, 게이트가 승인을 요청하는 단계에 이르면 멈추고 리드에게 올려야 한다.

**REQ-GDP-022** (Ubiquitous)
개정 뒤 이 트리에서 빌드해 경로로 호출한 `moai` 의 `constitution validate --format json` 은 `status: ok` 와 `drift_count: 0` 을 보고해야 하고, 증거는 그 빌드의 커밋이 트리 HEAD와 같다는 기록을 함께 담아야 한다.

**REQ-GDP-023** (Ubiquitous)
개정 뒤 `zone-registry.md` 로컬·템플릿 사본은 바이트 동일해야 하며, 두 사본 모두 `CONST-V3R5-027`·`028` 의 clause 로 §C.4의 `--after` 문장을 담고 이전 문장을 담지 않아야 한다.

### B.7 사본·생성물·중립성

**REQ-GDP-013** (Ubiquitous)
변경한 줄은 로컬 사본(`.claude/…`)과 템플릿 사본(`internal/template/templates/.claude/…`)에서 같아야 한다. `delivery.md` 두 사본의 의도된 차이(275·278행, 479-480행)와 `doc-execution.md` 두 사본의 의도된 차이(로컬 사본에만 있는 138-143행)는 그대로 남아야 한다.

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
| OD-2 — 워크트리 문맥 auto-merge 기본값의 단일 기준 | **선택지 B: `manager-git.md` 옵트인** | 2026-09-10 | 운영자(리드 경유) | 처음에는 선택지 C(설정 키 `workflow.worktree.auto_merge`)가 선택됐다. 그 키가 `class: D`, `evidence: none`, `deprecate_after: "v3.1.0"` 으로 기록돼 있다는 사실(`internal/config/testdata/shipped_key_inventory.yaml:2940-2943`)이 제시된 뒤 B로 바뀌었다. 이 SPEC은 그 키를 건드리지 않는다 |
| B1 — PR 수와 PR 브랜치 접두 | **SPEC당 PR 하나, `feat/SPEC-*`** (`manager-git.md` Phase C·`spec-assembly.md` 기준). spec-workflow.md Route B 표와 Step 규칙을 PR 하나로 고치고 `CONST-V3R5-028` 을 개정 대상으로 삼는다 | 2026-09-10 | 운영자(리드 경유) | — |
| B2 — primary checkout feature 브랜치 경로 | **워크트리 흐름으로 이동.** spec-workflow.md Route B 표와 Step 2·3 문구, `delivery.md` Step 3.2 github-flow(로컬·템플릿, 줄 단위, 의도된 사본 차이 유지)를 범위에 넣는다. Step 3.3.5의 결과를 명시한다(§C.3) | 2026-09-10 | 운영자(리드 경유) | — |
| B3 — Frozen 절 개정 경로 | **`moai constitution amend` 5겹 게이트로 `CONST-V3R5-027`(과 B1에 따라 `028`)을 개정.** `zone-registry.md` 로컬·템플릿을 범위에 넣는다. Frozen 개정에는 `--evidence` 필수. 마지막 HumanOversight 승인은 운영자의 것이며, run-phase에서 게이트가 사람의 승인을 요청하면 레인은 멈추고 리드에게 올린다 | 2026-09-10 | 운영자(리드 경유) | — |

결정은 plan 산출물의 방향을 정할 뿐 Implementation Kickoff Approval을 대신하지 않는다.

### C.2 재설계 흐름 — 진입·PR·종결 단계

primary checkout의 브랜치 상태는 Route B 흐름 전체에서 한 번도 바뀌지 않는다. Route B의 plan·run·sync는 모두 같은 워크트리 안에서 일어나고, SPEC마다 PR은 하나다(B1·B2). Route A(Tier S/M 기본, `main` 직접 커밋)는 바뀌지 않는다.

**진입 (Route B Step 1 시작)**

| 단계 | 명령·동작 | 설명 |
|---|---|---|
| E1 | 새 세션은 `moai cc -w <name>`, 현재 세션은 `EnterWorktree(<name>)` | `.claude/worktrees/<name>/` 워크트리에 들어간다. 맨손 `git worktree add` 는 쓰지 않는다 |
| E2 | (명령 없음) | 워크트리는 원격 기본 브랜치에서 잘린다. primary checkout의 로컬 `main` 을 먼저 갱신하지 않는다 |
| E3 | 워크트리 안에서 plan·run·sync 커밋 | `branch_creation.auto_enabled: false` 일 때 `/moai plan` 은 브랜치를 새로 만들지 않고, 세 단계의 커밋이 모두 워크트리 브랜치에 쌓인다 |

**PR (구 Phase C, SPEC당 한 번)**

| 단계 | 명령·동작 | 설명 |
|---|---|---|
| C1 | 워크트리 안에서 `git branch -m feat/SPEC-<ID>` | PR 시점에 SPEC PR 브랜치 이름을 붙인다(B1). 워크트리 안의 브랜치 이름 변경은 primary checkout 금지 대상이 아니다 |
| C2 | `git push -u origin feat/SPEC-<ID>` → `gh pr create --base {main_branch}` | 푸시와 PR 생성. 동기화 커밋도 같은 브랜치에 쌓인 뒤 전달된다 |
| C3 | 병합 조건 충족 시 `gh pr merge <PR> --<merge_method> --delete-branch` | 병합 방식은 `git_strategy.<mode>.merge_method`, 자동 병합 여부는 OD-2 = B |

**종결 (구 Phase D, Route B Step 4)**

| 단계 | 명령·동작 | 설명 |
|---|---|---|
| X1 | `gh pr view <PR> --json state` 가 `MERGED` | 원격 병합 착지 확인. 착지 전에는 X3를 하지 않는다 |
| X2 | `ExitWorktree` | primary checkout으로 돌아간다 |
| X3 | 세션 종료 keep/remove 질문, 또는 `git worktree unlock <path>` + `git worktree remove <path>` | 워크트리 폐기. 짧은 이름으로 들어간 트리는 L1이므로 `moai worktree done` 의 대상이 아니다 |
| X4 | (primary checkout에서 브랜치 명령 없음; 필요하면 `git fetch` 만) | `git checkout main`·`git reset --hard origin/main`·`git pull` 을 하지 않는다. primary checkout의 `main` 은 워크트리 커밋을 받은 적이 없고, 다음 SPEC은 E1의 새 워크트리에서 원격 기본 브랜치부터 시작한다 |

**구 절차에서 사라지는 것**: Phase A의 `git checkout main && git pull origin main`, Phase C의 `git switch -c`, Phase D의 `git checkout main` · `git reset --hard origin/main` · `git pull origin main`, Phase D 누락 복구, Step 1의 "main checkout에서 워크트리 없이", Route B의 plan PR·run PR·sync PR과 `plan/SPEC-`·`sync/SPEC-`·`chore/SPEC-…-sync` 브랜치, Run·Sync Phase의 "main checkout의 feature 브랜치" 기본 경로, Cleanup의 "BOTH run AND sync PRs" 전제와 `moai worktree done` 안내. push 거부 시 복구(구 125행)는 "워크트리 안에서 C1부터 진행" 으로 바뀐다.

### C.3 흐름이 스스로 답하지 않는 자리의 처리

| 자리 | 현재 | 처리 | 분류 장부 분류 |
|---|---|---|---|
| `manager-git.md:42` Checkpoint Rollback | `git reset --hard [checkpoint-tag]` | 되돌리기는 작업을 담은 워크트리 안에서만, 사용자 명시 확인 뒤 실행한다고 적는다. primary checkout에서는 실행하지 않는다 | `worktree-internal` |
| `manager-git.md:160` Synchronization | `git fetch origin` → `git pull origin [branch]` | `git fetch origin` 은 어느 트리에서나, `git pull` 은 해당 브랜치를 담은 워크트리 안에서만 실행한다고 적는다 | `worktree-internal` |
| `delivery.md:50` Step 3.1 Route B | 동기화 커밋이 `sync/SPEC-XXX`(`chore/SPEC-XXX-sync`) 브랜치에 쌓임 | 동기화 커밋은 워크트리 안의 같은 SPEC PR 브랜치(`feat/SPEC-<ID>`)에 쌓인다고 적는다(B1) | 명령 없음 |
| `delivery.md:238-257` Step 3.2 github-flow | "Feature branch (any branch other than main)" 가 트리를 묻지 않고 push·PR, "Worktree context" 가 따로 있음 | 두 경로를 "런처로 진입한 워크트리 안의 PR 브랜치" 하나로 합친다. primary checkout에서 `main` 이 아닌 브랜치를 만나면 push·PR 없이 멈추고 워크트리 진입을 안내하는 보고를 한다. "Main branch (direct commit)" 경로(Route A)와 "Any other branch state" 경로는 유지 | 명령 없음 |

**Step 3.3.5의 결과 (B2가 요구한 명시).** "Return to Base Branch (Post-PR Cleanup)" 안내는 없어진다. 대신: PR 생성(Step 3.2)과 선택적 ready 전환(Step 3.3)이 끝나면 세션은 `ExitWorktree` 로 워크트리를 떠나되 워크트리는 남겨 두고, primary checkout은 어떤 전략에서도 전환하거나 pull 하지 않는다. 워크트리 폐기는 원격 병합이 착지한 뒤(§C.2 X1·X3)에만 한다. 원격 브랜치 정리는 호스팅 쪽 자동 삭제 설정에 맡기는 기존 문장을 유지하고, 로컬 브랜치 삭제 명령(`git branch -d`)은 지운다 — 워크트리 폐기는 브랜치를 지우지 않으며, 삭제를 원하면 primary checkout이 아닌 곳에서 한다. `WT-*` 브랜치 문장은 유지한다. 절 제목의 `Step 3.3.5` 번호는 유지한다(수용 기준 추출 표지).

### C.4 헌법 개정 절차 (B3)

**개정 대상과 문구.** `--before` 는 등록된 clause 와 바이트 동일해야 한다(§A.7). `--after` 는 3층 ContradictionDetector를 통과하도록 `MUST` 를 유지하고 연속 낱말 `MUST NOT`·큰따옴표·MAY/SHOULD/OPTIONAL 을 담지 않으며, 템플릿 중립성(REQ-GDP-015)을 지키도록 SPEC ID·날짜를 담지 않는다. 같은 문장이 spec-workflow.md 로컬·템플릿 원본에 들어가야 검증기가 `DRIFT` 를 보고하지 않는다.

| 규칙 | `--before` (현재 등록 clause) | `--after` |
|---|---|---|
| `CONST-V3R5-027` | `Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step` | `Step 1 (plan) MUST execute in main checkout on Route A and inside a launcher-entered worktree on Route B` |
| `CONST-V3R5-028` | `Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged` | `Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after the single SPEC PR is merged` |

**명령 (run-phase, 작업 디렉터리 = 워크트리 루트, 트리 빌드 `/tmp/t622-moai-tree` 를 경로로 호출).** 레지스트리 해석이 두 갈래이므로(§A.7) `CLAUDE_PROJECT_DIR`·`MOAI_CONSTITUTION_REGISTRY` 를 같은 호출 안에서 해제한다. `--evidence` 값은 근거 문서 경로를 담고, 파이프 문자와 `<` 를 담지 않는다(AC-GDP-024 검출식).

```bash
# 레인: dry-run (1~4층 평가, 5층 자동 승인, 파일 변경 없음)
unset CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_REGISTRY && /tmp/t622-moai-tree constitution amend --dry-run --rule CONST-V3R5-027 --before 'Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step' --after 'Step 1 (plan) MUST execute in main checkout on Route A and inside a launcher-entered worktree on Route B' --evidence 'Operator decision 2026-09-10: Route B runs plan, run and sync inside one launcher-entered worktree and delivers one PR per SPEC; evidence .moai/reports/t622/run/const-amend-evidence.md' > .moai/reports/t622/run/const-amend-027-dryrun.txt 2>&1
unset CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_REGISTRY && /tmp/t622-moai-tree constitution amend --dry-run --rule CONST-V3R5-028 --before 'Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged' --after 'Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after the single SPEC PR is merged' --evidence 'Operator decision 2026-09-10: one PR per SPEC replaces the run PR and sync PR pair; evidence .moai/reports/t622/run/const-amend-evidence.md' > .moai/reports/t622/run/const-amend-028-dryrun.txt 2>&1

# 운영자: 대화형 터미널에서 같은 명령을 --dry-run 없이 실행하고 Y/N 에 직접 답한다 (레인은 실행하지 않는다)
```

**게이트 예측 (소스 읽기 기반, 실행 관측 아님 — §E.1).** 1층: `--evidence` 가 있어 통과. 2층: 해당 형태의 SPEC 디렉터리 0개라 사용 불가, 파이프라인은 계속. 3층: 위 `--after` 는 차단 조건에 걸리지 않음. 4층: 기존 기록 파일의 항목이 이 파서 형식으로 읽히는지 확인하지 않았음; 첫 개정이 기록되면 두 번째 개정은 24시간 뒤에만 통과. 5층: 표준입력 Y/N — 운영자 몫.

**사람 승인 경계 (HARD).** 레인은 dry-run까지만 실행한다. `--dry-run` 없는 호출은 운영자가 대화형 터미널에서 실행하고 승인한다. 레인은 `Y` 를 파이프·here-string·파일 리다이렉트로 넘기지 않고, 게이트가 승인을 요청하는 단계에 이르면 멈추고 dry-run 출력과 명령 원문을 담아 리드에게 올린다.

**적용 단계의 현재 상태.** 운영자가 승인해도 `applyAmendment` 의 첫 도우미가 스텁이라 "amendment application error: source file update error: updateSourceFile: not yet implemented" 로 끝나고 원본·레지스트리·기록 파일은 바뀌지 않는다(§A.7). 또 파이프라인은 로컬 레지스트리만 다루므로 적용 단계가 구현되더라도 템플릿 레지스트리는 줄 단위로 따로 맞춰야 한다(REQ-GDP-023). 이 상태에서 B3을 끝낼 방법은 §C.5 B4의 결정이 필요하다.

**등록되지 않은 Frozen 줄.** spec-workflow.md 23행 블록(26행 Route B 정의)과 166행(Plan Phase)은 `[ZONE:Frozen]` 이지만 레지스트리 항목이 없어 `amend --rule` 로 지정할 수 없고, 검증기도 이 줄을 검사하지 않는다(§A.5). 이 줄의 변경은 REQ-GDP-010에 따라 Frozen 구역 수정으로 기록하고 Implementation Kickoff Approval에서 명시한다.

### C.5 남은 차단 항목

| ID | 항목 | 측정한 사실 | 선택지 (결정하지 않음) |
|---|---|---|---|
| B4 | B3의 `amend` 경로가 이 트리에서 개정을 적용하지 못함 | `pipeline.go:256-259`·`:264-266` 두 도우미가 스텁, `pipeline_test.go` 가 스텁 동작을 고정. 승인 뒤 첫 단계에서 실패, 파일 변경 없음 | (a) 적용 단계를 먼저 구현하는 별도 카드·SPEC을 선행 조건으로 두고(Go 코드, 이 SPEC의 §D 제외 범위) 그 뒤 amend 실행; (b) 게이트 기록은 amend로 남기고(레인 dry-run + 운영자 대화형 승인), 승인된 문장은 spec-workflow.md와 레지스트리 L·T에 줄 단위로 적용해 Frozen 수정으로 기록; (c) 레지스트리의 문서화된 절차대로 `027`·`028` 을 `[SUPERSEDED by …]` 로 은퇴시키고 새 항목을 등록 |
| B6 | `worktree-integration.md` 가 B1·B2 뒤 spec-workflow.md와 모순 | 45행 L2 행 "after both run + sync PRs merge", 543-556행 SPEC-to-Worktree Mapping 표(plan PR·run PR·sync PR, "the default flow runs all phases on a `feat/SPEC-XXX` branch in the main checkout"), 545행 `[ZONE:Frozen]` "on conflict, spec-workflow.md wins", 556행 `[ZONE:Frozen]` Disposal contract "after BOTH run PR AND sync PR are merged". 두 Frozen 줄 모두 레지스트리 항목 없음. 로컬·템플릿 바이트 동일 | (a) 45·543-556행 L·T를 범위에 넣고 556행을 Frozen 수정으로 기록; (b) 범위 밖 형제로 남기고 545행의 우선순위 문장에 기대기 |
| T1 | 요구사항·수용 기준 수가 Tier M 상한을 넘음 | REQ 23개(활성 21·철회 2), AC 24개(활성 22·철회 2) — 상한 16/16, 각각 독립 | Tier 재분류 여부는 리드 결정. `tier:` 필드와 design.md·research.md는 바꾸거나 만들지 않았다 |

### C.6 검토한 선택지 (결정 당시 자료)

결정 전에 운영자에게 전달한 표를 기록으로 남긴다. 줄번호는 `b412f8a33` 기준이며 "L·T" 는 로컬·템플릿 두 사본이다.

#### OD-1 — Late-branch 절차의 처리

| 선택지 | 영향 파일 (로컬·템플릿·생성물) | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 (키 · 현재 기본값 · 변화) | 뒤집거나 확장하는 선행 SPEC |
|---|---|---|---|---|
| **1. 런처 워크트리 흐름으로 재설계** (채택) | `manager-git.md` L·T 88-129·137 + 재생성 `.toml`; `spec-workflow.md` L·T 49-62·65; `spec-assembly.md` L·T 332-340; `delivery.md` L·T 319-326. 스스로 답하지 않는 자리: `manager-git.md:42`·`160`, `delivery.md:323-324`·`328` | 부분. 커밋을 쌓는 장소가 primary checkout의 `main` 에서 런처로 진입한 워크트리로 바뀐다. Frozen 구역 수정이 따라온다 | `branch_creation.auto_enabled` — 현재 `false`. 값은 그대로, `false` 가 촉발하는 동작이 바뀐다 | SPEC-V3R5-LATE-BRANCH-001 REQ-LB-003·005·006 대체, EXCL-LB-004 뒤집음. REQ-WBG-011 근거 소실 |
| **2. 유지하되 명시적 조건으로 범위 한정** (미채택) | 선택지 1의 네 지침 파일 L·T + `.toml`; `AGENTS.md:64` 루트·템플릿, `main-checkout-branch-guard.md` L·T 16 | 철회 없음. 모든 배포 사용자의 금지 원칙이 조건부로 바뀐다 | 조건 후보 `workflow.branch_guard.enabled` — Go 기본값 `false` | GUARD-OPTIN-001 전제를 원칙 층으로 확장, GUARD-001 원칙을 좁힘 |
| **3. `main_late_branch` 은퇴** (미채택) | `manager-git.md` L·T 88-129·137 + `.toml`; `spec-workflow.md` L·T 50·53-62; `spec-assembly.md` L·T 332-340; `git-strategy.yaml.tmpl` 82-84 주석; `--issue` opt-in 정책 문구 | 예. Phase A-D 절차가 사라진다 | `auto_enabled: false` 동작을 새로 정의해야 함 | LATE-BRANCH-001 REQ-LB-001~006 뒤집음 |
| ~~예시 명령만 삭제~~ (제외) | — | 절차가 사라진 채 참조만 남는다. REQ-GDP-008이 금지 | — | — |

#### OD-2 — 워크트리 문맥 auto-merge 기본값의 단일 기준

| 선택지 | 영향 파일 | 배포 사용자에게서 철회·변경되는 기능 | 기본값 영향 | 뒤집거나 확장하는 선행 SPEC |
|---|---|---|---|---|
| **A. `delivery.md` 워크트리 기본값** (미채택) | `manager-git.md` L·T 148·166 + `.toml` | team 모드 워크트리 문맥에서 플래그 없이 병합 | 플래그로 결정 | `b7447cb90` 동작을 확정 |
| **B. `manager-git.md` 옵트인** (채택) | `delivery.md` L·T 335-338·348-349; `doc-execution.md` L·T 34-36 | 워크트리 문맥의 sync가 기본으로 병합하지 않는다 | 워크트리 문맥 동작 기본값 "병합" → "병합 안 함" | `b7447cb90` 동작을 뒤집음 |
| **C. 설정 키 `workflow.worktree.auto_merge`** (처음 선택 후 B로 교체) | `delivery.md`, `doc-execution.md`, `manager-git.md` + 사용처 분류 파일 | 기본 병합이 멈춤 | 키 값 `false` 유지, 첫 소비자. 키는 `class: D`, `deprecate_after: "v3.1.0"` | WORKTREE-ENTRY-STRATEGY-001 확장 |

B1~B3은 리드가 운영자에게 선택을 올렸고, 0.1.2 §C.5가 그 자료였다(측정 사실은 §A.2.2·§A.2.3·§A.5·§A.7에 옮겼다).

---

## §D 제외 범위 (Out of Scope)

### Out of Scope — 채택되지 않은 선택지의 경로

- OD-1 선택지 2의 조건부 원칙: 루트·템플릿 `AGENTS.md` §2 금지문과 `main-checkout-branch-guard.md` 원칙은 바꾸지 않는다.
- OD-1 선택지 3의 은퇴: `main_late_branch` 옵션과 `--issue` opt-in 정책 문구를 없애지 않는다.
- OD-2 선택지 C: 설정 키 `workflow.worktree.auto_merge` 와 사용처 분류 파일은 건드리지 않는다.

### Out of Scope — 코드와 설정 값

- Go 코드 변경. 특히 `internal/hook/branch_guard.go:455` 의 manager-git 신원 면제(§E.2, 후속 카드 후보)와 `internal/constitution/pipeline.go` 의 적용 단계 스텁(§C.5 B4 선택지 (a)는 별도 카드로만 가능)은 그대로 둔다.
- `git-strategy.yaml.tmpl`·`workflow.yaml` 의 키 값 변경, `main-checkout-branch-guard` 훅 변경.
- Route A(Tier S/M 기본, primary checkout의 `main` 직접 커밋)의 절차.

### Out of Scope — 카드 범위 밖에서 관측한 형제 (리드 판단 대상)

- `manager-git.md:82` — `auto_branch: true` 일 때 "Create `feature/SPEC-{ID}`, checkout from main_branch".
- `manager-git.md:144` — "GitHub Flow: main + feature/SPEC-* branches" (B1의 `feat/SPEC-*` 와 접두 다름).
- `manager-git.md:171` — PR Auto-Merge 5단계 "Checkout main, pull, delete local branch".
- `spec-assembly.md:388` — `--branch` 경로 "Create branch: feature/SPEC-{ID}-{description}", `:547` 시나리오의 `feat/SPEC-AUTH-001-jwt-auth`.
- `delivery.md:356` — Auto-Merge Execution 4단계 "Checkout target branch, fetch latest".
- `delivery.md` Step 3.2 git-flow 경로 — `feature/*` 만 PR 경로로 받으므로 B1의 `feat/SPEC-*` 브랜치는 "Any other branch → stop and report" 로 떨어진다(§E.2).
- `generic-patterns-guide.md`(템플릿 195행~, 로컬 207행~) Late-Branch Phase D Recovery.
- `worktree-integration.md` 45·543-556행 — §C.5 B6에서 범위 결정을 기다린다.

### Out of Scope — 금지 목록과 경고표

- 다음 위치의 브랜치 변경 명령은 금지 목록·경고표·인용 예시·권한 항목·워크트리 내부 안내라서 유지한다: 템플릿 `AGENTS.md` 64-66, `main-checkout-branch-guard.md` 20-24와 detail 파일, `coding-standards.md:141`, `moai-ref-git-workflow/SKILL.md` 144-145, `verification-claim-integrity-detail.md:161`, `settings.json.tmpl:564`, `moai-workflow-worktree/modules/troubleshooting.md` 176·179.

---

## §E 미검증과 잔여 위험

### E.1 미검증 (Gaps)

- 파괴적 명령과 `git fetch`/`git rev-list` 경합을 실제로 재현하지 않았다. 재설계 흐름을 실제 저장소에서 끝까지 실행하지 않았다.
- `moai constitution amend` 는 도움말만 읽고 실행하지 않았다(리드 제약). §C.4의 게이트 예측은 소스 읽기이며 관측이 아니다. 특히 4층이 기존 `.moai/research/evolution-log.md` 항목(`---` 로 나눈 YAML 파싱)을 읽는지 확인하지 않았다.
- 헌법 검증 측정은 설치 빌드(`v3.2.0-rc.5`, `84fa4ece4`)로 했고 이 트리에서 빌드한 바이너리로 하지 않았다. 설치 빌드가 이 트리의 조상인지 확인하지 않았다. 대신 `internal/constitution`·`internal/cli/constitution.go` 가 `b412f8a33`..HEAD 사이에 바뀌지 않았음을 `git diff --stat` 로 확인했다.
- `go version -m` 의 `vcs.revision` 이 이 워크트리 빌드에서 채워지는지 확인하지 않았다(AC-GDP-022의 출처 기록이 이것에 기댄다).
- `manager-git.md`·`agent-common-protocol.md`·`doc-execution.md`·`zone-registry.md` 는 `rule_template_mirror_test.go` 의 바이트 동일 대상이 아니다. 사본 일치는 이 SPEC의 diff 기준으로만 확인된다.
- `TestTemplateNoInternalContentLeak` 의 분류가 에이전트·규칙·스킬 경로에 적용되는지는 읽지 않았다.

### E.2 잔여 위험

- **REQ-WBG-011 근거 소실 (운영자 지시로 기록).** Phase D가 사라지면 SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-011의 근거 문장이 성립하지 않고, `internal/hook/branch_guard.go:455` 의 면제는 적힌 근거 없이 남는다. 훅 변경은 범위 밖이며 **후속 카드 후보**로 기록한다(발행 여부는 리드 결정).
- **git-flow 전략과 B1 접두.** `delivery.md` Step 3.2 git-flow는 `feature/*` 만 PR로 받는다. git-flow 사용자가 Route B를 쓰면 `feat/SPEC-*` 브랜치는 전달 단계에서 멈춘다. 이 SPEC은 github-flow만 바꾼다.
- **카드 워크트리 명명과 C1.** 배포 `AGENTS.md:125` 는 카드 워크트리 브랜치를 만든 즉시 `WT-<slug>` 로 바꾸라고 한다. 칸반 카드가 Route B를 함께 쓰면 PR 시점에 `feat/SPEC-<ID>` 로 한 번 더 바꾸게 되고, git-flow에서는 `WT-*` 가 통합 워크트리 경로를 탄다.
- **검출식 한계.** 명령·서술형·단계별 PR 검출식은 모두 줄 단위이며, 예상하지 않은 표현(예: "repository root" 로 primary checkout을 가리키는 문장)은 통과한다. 읽기 단계로 틈을 좁히지만 오독 가능성이 남는다.
- **등록되지 않은 Frozen 줄.** spec-workflow.md 23행 블록·166행과 worktree-integration.md 545·556행은 검증기가 보지 않아 변경·미변경 모두 기계적으로 드러나지 않는다.
- OD-2 = B는 워크트리 문맥의 기본 병합을 없앤다. 지금 이 기본값에 기대는 사용자는 `--auto-merge` 를 넘기지 않으면 병합되지 않는 PR을 보게 된다.
- §A와 §C의 줄번호는 범위 파일이 `b412f8a33` 과 바이트 동일하다는 전제에 기댄다(HEAD `880c0c702` 에서 확인). run-phase 전에 develop을 흡수하면 다시 확인해야 한다.

---

## §F 수용 기준

`acceptance.md` 가 기준(AC-GDP-001~024; 011·012는 철회)과 검증 명령, 양성 대조, 감사 뮤턴트 재실행 결과를 담는다.
