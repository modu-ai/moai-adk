---
id: SPEC-GIT-PROC-SAFE-001
title: "Git 절차 교정 — Late-Branch 워크트리 전환 및 auto-merge 레시피 SSOT 정리"
version: "1.0.0"
status: in-progress
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: git-procedure-doctrine
lifecycle: spec-anchored
tier: M
tags: "git, doctrine, manager-git, sync, worktree"
related_specs: [SPEC-GITFLOW-DOCTRINE-ALIGN-001, SPEC-GIT-DELIVERY-PROCEDURE-001]
---

# SPEC-GIT-PROC-SAFE-001 — Git 절차 교정 (Late-Branch 워크트리 전환 + auto-merge SSOT 정리)

## HISTORY

| 일자 | 버전 | 변경 내용 |
|------|------|-----------|
| 2026-09-14 | 1.0.0 | 최초 작성 — 카드 t782. 명령 감사(instruction-audit) 3건(AC-01, AC-11, SX-R04)을 반영. 재검증 baseline: develop `c9ceff175` (본 워크트리 기준) |

## 1. 배경 및 목적

명령 감사(baseline: main `2213871af`)가 `manager-git` git 절차 교리에서 3건의 결함을 지적했다. 3건은 본 워크트리(develop `c9ceff175`)에서 실행 재검증을 마쳤으며, 아래 요구사항은 가설이 아니라 측정된 현상을 인코딩한다.

- **AC-01 (P1, LIVE)** — `manager-git.md` § Late-Branch Invocation Pattern이 primary 공용 체크아웃에서 브랜치 상태를 바꾸는 절차를 안내한다. 이는 `AGENTS.md` §2 가 금지하는 행위다. 교리 궤적이 `spec-workflow.md`(정본 수순 보유)까지 이어져, 한쪽만 고치면 교차참조된 정본에 모순이 남는다.
- **AC-11 (P2, 이미 수리됨)** — `manager-git.md` § Synchronization의 `git fetch` ↔ `git rev-list` 배치 순서 단락이 감사가 처방한 수리 문구를 이미 갖고 있다. 본 SPEC은 편집 없이 폐쇄 사실을 검증만 한다.
- **SX-R04 (P1, 부분 수리)** — `delivery.md` Step 3.4가 `manager-git.md` § PR Auto-Merge(SSOT)의 모드 조건·실행 레시피·실패 처리를 중복 인용한다. 트리거 단일 기준 문장과 `merge_method` 설정 해석은 이미 수리됐고, 잔여 결함은 중복 레시피뿐이다.

핵심 기계적 사실(replacement model의 근거): 하나의 브랜치는 오직 하나의 워크트리에만 체크아웃될 수 있다. 따라서 "커밋이 main에 쌓인다"는 모델 자체가 워크트리 격리와 양립 불가능하며, 대체 모델은 **"커밋은 워크트리 자신의 브랜치에 쌓이고, 브랜치는 마지막에 구성된 통합 경로로 승격된다"** 이다. 승격 단계는 구성된 workflow에 조건부다: 설정이 PR로 통합하면 브랜치를 PR로 승격하고, git-flow이면 통합 창(integration window) 규율에 따라 develop 통합 워크트리로 병합한다 — 실제 운영 구성이 후자이기 때문이다(측정: git-strategy.yaml 9행 `workflow: git-flow`, 29행 `auto_enabled: false`). 이 모델은 Phase D의 `reset --hard` main 정렬도 함께 소멸시킨다 — main이 애초에 커밋을 받지 않기 때문이다.

## 2. 요구사항 (GEARS)

### REQ-GP-001 — Late-Branch 절차의 워크트리 전환 (AC-01 본체)

**While** `team.branch_creation.auto_enabled == false`(Late-branch 기본값) **인 동안**, `manager-git.md`의 Late-Branch 절차는 **모든** git 브랜치 상태 변경(체크아웃/스위치, 브랜치 생성, reset, 병합)을 런처로 진입한 워크트리(`moai cc -w <name>` 또는 `EnterWorktree(<path>)`) 내부에서만 수행하도록 규정해야 한다. primary 체크아웃은 어떤 브랜치 상태 변경도 목격해서는 안 된다.

- Phase A~D 및 실패 복구 수순 전반이 이 모델을 따라야 한다.
- 현행 위반 명령(폐기 대상, baseline `c9ceff175`에서 file:line 확인): Phase A `git checkout main && git pull origin main`(로컬 98행), Phase C `git switch -c feat/SPEC-XXX`(112행), Phase D `git checkout main` / `git fetch origin` / `git reset --hard origin/main` / `git pull origin main`(121-124행), 실패 복구 `git fetch origin && git reset --hard origin/main`(129행), Personal 모드 옵션 행 `main_late_branch` 설명의 `reset --hard origin/main`(139행).

### REQ-GP-002 — 대체 절차 텍스트의 자기규율 (AC-01 수반)

대체 절차 텍스트는 primary 체크아웃에 대해 `AGENTS.md` §2 금지 행위(체크아웃/스위치, 브랜치 생성, `reset --hard`, `checkout -- <path>`, `stash`, checked-out 브랜치 위 rebase/merge)를 가리키는 수순을 담지 않아야 한다. 허용 형태는 다음으로 한정한다: 자기 워크트리 안의 `git branch -m`, `git fetch`, checked-out 브랜치로의 커밋, 해당 브랜치의 push.

### REQ-GP-003 — SSOT 궤적의 동시 수리 (AC-01 확장 범위)

`spec-workflow.md`(로컬 `.claude/rules/moai/workflow/spec-workflow.md` + 템플릿 미러)의 Late-branch 관련 단락은 `manager-git.md`와 **같은 이슈 범위 안에서** 수리되어야 한다:

- 50행 Route B Late-branch 사전조건 — "Step 1 진입 요건 `git rev-parse --abbrev-ref HEAD == main`... plan-phase 커밋이 `main`에 직접 쌓이고 Phase C `git switch -c plan/SPEC-XXX` 시점에만 push" 모델을 워크트리 모델로 대체.
- 53-62행 Step 4 Late-branch closure — 정본 closure 명령 블록(`git checkout main` / `git fetch origin` / `git reset --hard origin/main` / `git pull origin main`)과 post-condition(`git status --porcelain` 비어 있음 + `git rev-parse main == origin/main`)을 워크트리 모델에 맞게 재작성.

근거: `manager-git.md` 131행은 spec-workflow.md에 "canonical step ordering"을 위임하는 교차참조를 갖는다. manager-git.md만 고치면 참조된 정본에 금지 수순이 그대로 남아 모순이 산다. 이 확장은 감사 인용 행 밖의 의도된 문서화된 범위 확장이며, 그 근거를 본 SPEC에 기록한다.

### REQ-GP-004 — 교차참조 무결성

수리 후 두 파일의 교차참조(`manager-git.md` → spec-workflow.md § Step 1/§ Step 4, spec-workflow.md → `manager-git.md` § Late-Branch Invocation Pattern)는 계속 해석 가능해야 한다 — 섹션 제목이 바뀌면 양쪽 참조 텍스트가 같은 이슈 안에서 갱신되어야 한다.

### REQ-SX-001 — delivery.md Step 3.4의 SSOT 위임 (SX-R04 잔여 결함)

`delivery.md` Step 3.4는 트리거 조건 + `manager-git.md` § PR Auto-Merge 에 대한 포인터로 축소되어야 한다. 제거 대상 중복: "Mode conditions (same as `manager-git.md` § PR Auto-Merge)" 뒤의 두 조건 원문 재인용, "Auto-Merge Execution"의 5단계 레시피("Checkout target branch, fetch latest" 포함), "Auto-Merge Failures"의 실패 처리 중복. 보존 대상(delivery 고유): Step 3.4의 "PR이 Step 3.2에서 생성된 경우에만 적용" 프레이밍과 `workflow.worktree.auto_cleanup` 키의 Post-Merge Automatic Cleanup — 이 둘은 delivery 계층 고유 책임이므로 delivery.md에 남는다(판단 권한은 run-phase에 위임하되 본 SPEC이 보존 의도를 명시한다).

SSOT 귀속: PR Auto-Merge의 깃발 트리거·폐기 별칭·모드 조건·merge_method 해석 5단계는 `manager-git.md` § PR Auto-Merge가 유일한 정본이다.

### REQ-AC11-001 — AC-11 검증 전용 폐쇄 기록 (편집 없음)

**When** run-phase가 AC-11 발견 건을 처리할 때, 실행 에이전트는 대상 6개 파일에 대해 어떤 편집도 수행해서는 안 되며, 양 사본에서 폐쇄 문구의 존재를 직접 관측해 폐쇄 기록으로 남겨야 한다. 관측 대상: `manager-git.md` 158행(템플릿 156행)의 fetch→rev-list 배치 순서 수리 문구 — 이것이 "이미 사라졌다면 그 사실이 산출물이다"라는 카드 지시의 이행이다.

### REQ-TF-001 — Template-First 양 사본 적용

**When** 대상 6개 파일 어디든 편집이 일어날 때, 실행 에이전트는 템플릿 사본(`internal/template/templates/...`)을 먼저 로컬 사본(`.claude/...`)을 둘째로 발견 건 범위 편집으로 적용해야 하며, 사본 간 선존재 분기 영역(manager-git.md frontmatter description ~5행, delivery.md ~94 diff 행 — lint/크로스빌드 블록, Route A/B 산문, TRACE PROBE 주석)을 접촉해서는 안 된다. 전체 파일 cp 미러링과 범위 밖 영역 정합화는 금지된다.

### REQ-TN-001 — 템플릿 중립성

템플릿 사본 콘텐츠는 내부 SPEC ID를 실어 나르지 않아야 하며, dev-only 로컬 경로(`.claude/rules/local/*` 등)를 참조하지 않아야 하고, 16개 지원 프로그래밍 언어에 중립을 유지해야 한다.

## 3. 수락 기준 요약

상세 Given-When-Then은 acceptance.md에 둔다. 요약:

- AC-GP-01a/b/c — manager-git.md(양 사본) / spec-workflow.md(양 사본) / delivery.md(양 사본)의 RED(위반 텍스트) → GREEN(대체 텍스트) 전환. 문서 전용 변경이므로 RED 셀은 baseline에서 관측된 위반 텍스트 + file:line, GREEN 셀은 대체 텍스트 검사로 표현한다.
- AC-SX-01 — delivery.md Step 3.4의 중복 제거 + delivery 고유 절 보존.
- AC-AC11-01 — 편집 0건 + 양 사본 폐쇄 텍스트 존재 관측.
- AC-XREF-01 — 교차참조 해석 가능성.
- AC-MIRROR-01 — 범위 밖 분기 무손상(전체 파일 diff에서 발견 건 범위 밖 행 불변).

## 범위 밖 (Out of Scope)

### Out of Scope — 선존재 사본 분기

- manager-git.md 로컬 vs 템플릿 frontmatter description 지연(~5행) — 후속 후보로 research.md에 기록만, 본 이슈에서 수리하지 않음
- delivery.md 로컬 vs 템플릿의 lint/크로스빌드 블록·Route A/B 산문·TRACE PROBE 주석 차이(~94 diff 행) — 후속 후보로 기록만

### Out of Scope — 런타임 코드

- `internal/hook/branch_guard.go` 등 기계 가드 코드 변경 없음 — 본 SPEC은 문서 절차 교정만 다룬다
- `git-strategy.yaml` 스키마·기본값 변경 없음 (`auto_enabled`, `merge_method` 등 설정 표면 불변)

### Out of Scope — AC-11 재수리

- AC-11에 대한 어떤 파일 편집도 하지 않는다 (이미 수리됨 — 검증 전용)

### Out of Scope — 다른 교리 파일로의 파급

- `AGENTS.md`, `kanban-dispatch.md`, `worktree-integration.md`, `main-checkout-branch-guard.md` 등 워크트리/브랜치 교리는 이미 워크트리 모델을 서술하고 있어 건드리지 않는다. 본 SPEC이 건드리는 것은 레거시 late-branch 모델을 여전히 서술하는 6개 파일 뿐이다.

## 위험

- **교리 간 잔존 모순**: 6개 파일 외에 late-branch `reset --hard` 모델을 언급하는 다른 표면이 존재할 가능성. 완화: run-phase에서 `auto_enabled`/late-branch 문맥 grep으로 2차 스윕하고, 적중은 blocker 보고로 상향(범위 확장은 운영자 판정).
- **closure post-condition 상실**: 워크트리 모델에서 "local main 정렬" post-condition은 무의미해지므로, 새 post-condition(워크트리 브랜치가 origin 기본 브랜치에서 분기했으며 main은 불변)을 명확히 정의해야 검증 가능성이 유지된다.
