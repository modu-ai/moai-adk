---
id: SPEC-GITHUB-FLOW-DEFAULT-001
title: "github-flow 기본 개발 흐름 전환 — main 단일 기준·카드 PR 전달·main 태그 릴리스"
version: "0.1.2"
status: in-progress
created: 2026-10-02
updated: 2026-10-03
author: GOOS
priority: P1
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "github-flow, git-flow, cutover, card-pr-delivery, release-from-main, sweep-guard, merge-window"
tier: L
---

# SPEC-GITHUB-FLOW-DEFAULT-001 — github-flow 기본 개발 흐름 전환

## HISTORY

| 버전 | 날짜 | 작성 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-10-02 | manager-spec (카드 t1453) | plan-phase 산출물 최초 작성 (Tier L, 5 artifacts) |
| 0.1.1 | 2026-10-02 | manager-spec (카드 t1453) | plan-audit 1회차(0.74, FAIL) 반영. 미해소 질문 7건을 결정 D-16~D-22 로 기록하고 표지 문자열을 제거. 절체 순서 충돌(AC-016·AC-019·push 트리거)을 D-17·D-25 로 해소. REQ-GFD-008·010·013·016·019·020·021·022 문언 정정, M2(d)·M3(f) 범위 삭제, AC-023(런북 내용) 추가로 AC 23개. REQ 는 22개 그대로 |
| 0.1.2 | 2026-10-02 | manager-spec (카드 t1453) | plan-audit 2회차(0.84, FAIL) 반영. REQ-GFD-002 를 `done`·`sweep` 의 세 층과 세션 종료 정리의 두 층(네트워크 호출 없음)으로 분리(N-04). AC-008 에 `release.yml` 배선 단언과 접미사 없는 태그 동치 시험 추가(N-01). Frozen 줄 분류와 미등재 줄의 일반 편집 취급을 REQ-GFD-013·AC-GFD-013·AC-GFD-017 에 명시(N-03). 스윕 가드의 약한 적중을 허용 목록 외에는 위반으로 승격하고 픽스처 21→25(N-05, D-27). CI 증거를 실행 id 로 고정하거나 Gaps 로 이동(N-07, D-28). 런북 행당 주체 하나(N-02, D-29). 선택 항목 D7~D12 도 반영. REQ 22개·AC 23개 그대로 |

## §A 배경

### A.1 무엇을 결정했는가

2026-10-02 저녁 운영자가 이 저장소의 기본 개발 흐름을 git-flow(develop 통합 브랜치·`release/vX.Y.Z` 브랜치·뒤처지는 main)에서 **github-flow(main 단일 기준·짧은 카드 브랜치·push 와 PR·CI 재측정·GitHub 병합)** 로 바꾸기로 확정했다. 타당성 조사가 아니라 결정이며, 이 SPEC 은 그 결정을 실행 가능한 변경 묶음과 절체 절차로 옮긴다. 쟁점은 "할 것인가"가 아니라 "어떤 순서로, 무엇을 절체 시점까지 묶어 두고, 무엇을 지금 병합해도 안전한가"다.

### A.2 현황 (실측 — 좌표는 `acceptance.md` 증거 원장과 `research.md`)

- **배포 기본값은 이미 github-flow 다.** 배포 템플릿은 세 프로필 모두 `workflow: github-flow` 이고 `worktree_base_branch` 는 비어 있다(원장 E-30, `git-strategy.yaml.tmpl`). 이 저장소가 추적하는 설정만 manual 프로필에서 git-flow 값을 든다(원장 E-08, 이 저장소의 파일: `worktree_base_branch: develop`, `workflow: git-flow`, `develop_branch: develop`).
- **main 과 develop 은 크게 갈라져 있다.** 두 팁 사이 발산은 `7614 1` 이다(원장 E-22). main 에만 있는 커밋은 하나(`4755c5e50`, PR #1740)이고 develop 에는 그 내용이 없다. develop 이 main 을 흡수하는 병합은 충돌 없이 끝난다(`research.md` §3).
- **develop 위의 통합은 로컬 병합 창과 리더 일괄 push 로 돌아간다.** 카드마다 PR 이 없고, 원격 CI 는 develop push 마다 한 번 돈다. 이 CI 의 최근 12회(실행 id 열두 개로 고정, 원장 M-1)는 7회 failure·5회 cancelled 였다(`research.md` §4). 운영자가 정한 절체 선행 조건은 origin/develop 팁의 CI 녹색이다(design D-17).
- **main 은 보호돼 있다.** PR 필수·force-push 불가·관리자 포함(`enforce_admins`)이고 필수 체크 다섯 개가 `strict: false` 로 걸려 있다. `v*` 태그는 불변 규칙셋으로 삭제도 이동도 막혀 있다.

### A.3 이 SPEC 의 성격

코드·설정·CI·문서에 걸친 전환이지만, **기준 브랜치를 실제로 바꾸는 단계는 이 SPEC 의 실행 범위가 아니다.** 그 단계는 외부 공유 시스템을 바꾸고 되돌리기 어려운 작업이라 리더의 확인을 받은 배치 경계에서만 수행한다. 이 SPEC 이 만드는 것은 (1) 지금 병합해도 안전한 변경, (2) 절체 시점에 한 묶음으로 적용할 변경, (3) 절체 런북과 그 리허설 증거다.

## §B 범위

### B.1 In Scope

| 항목 | 내용 | 마일스톤 |
|---|---|---|
| 카드 항목 ① | 설정·코드의 기준 브랜치 해석을 구성된 통합 목표로 옮기고 이 저장소의 값을 절체 시 github-flow 로 | M1, M5 |
| 카드 항목 ② | 카드 전달 경로 재정의(카드 브랜치 push·PR·CI 재측정·GitHub 병합), 병합 큐 판정(보류), 로컬 병합 창 처분 | M2 |
| 카드 항목 ③ | rc 와 정식 릴리스를 main 태그 기준으로, 3-OS 매트릭스 게이트의 위치(옵션 B: 태그 전 `workflow_dispatch` 와 스크립트 강제) | M3, M5(스위치 켜기) |
| 카드 항목 ④ | 지침·문서 전면 갱신(AGENTS·CLAUDE·규칙·에이전트·스킬·README 4종·docs-site 4로케일·템플릿 사본), Frozen 조항 정규 개정 | M4 |
| 카드 항목 ⑤ | 낡은 develop 서술 잔존 0 을 스윕 가드 테스트로 고정 | M4 |
| 절체 절차 | 수렴 병합·트리 항등 검증·되돌리기·레인 재기동 절차와 스크래치 클론 리허설 | M6 |
| 흡수·조정 | t810(`SPEC-LATE-BRANCH-REDESIGN-001`) 흡수(리더 결정, design D-19), t1452 조정(리더 결정, D-11), t1448 후 규칙 편집 | M4, plan.md §G |

### B.2 Out of Scope

### Out of Scope — 휴면 git-flow 코드 삭제

- git-flow 는 기본값에서 은퇴할 뿐 사용자 프로젝트가 고를 수 있는 `git_strategy` 워크플로로 남는다. `internal/mission/integration.go`, `lead_push_threshold` 처럼 github-flow 에서 휴면이 되는 항목은 **후속 인벤토리**로 `design.md` §D-1 에 목록만 남기고 이 카드에서 지우지 않는다.

### Out of Scope — 절체의 실행

- main 보호 규칙(필수 체크 목록 변경 포함)·develop 보호·develop 삭제·기준 브랜치 절체 자체(이 저장소의 `git-strategy.yaml` 값을 main 기준으로 확정하는 병합과 레인 재기동)는 리더·운영자 소관이며 이 카드가 확인 기록 없이 실행하지 않는다. 런북 문서와 리허설은 범위 안이다.
- `Release PR Multi-OS Gate` 를 main 의 필수 체크에서 빼는 일은 운영자가 직접 수행하는 런북 단계다(외부 공유 시스템, 되돌리기 어려움). 이 카드는 그 단계를 런북에 적을 뿐 실행하지 않는다(design D-22).

### Out of Scope — 병합 큐 도입

- 병합 큐의 도입은 이 SPEC 이 판정(보류)하되 구현하지 않는다(운영자 결정: 카드마다 PR 로 전달, 병합 큐는 보류 — design D-20). 미측정 입력(청구 분, CodeRabbit 한도 이력, 의미 충돌 빈도)은 후속 측정 항목으로 `research.md` §9 에 남기며 그 항목의 카드 발행은 리더 소관이다.

### Out of Scope — 계약 모드 에스컬레이션 분류기

- `internal/escalation` 의 계약 모드 에스컬레이션 분류(`pushWhy` 류)는 이 카드가 바꾸지 않는다. 계약 모드 레인은 PR 전달의 카드 브랜치 push 에서 에스컬레이션을 만나며, 이를 푸는 후속은 리더에게 권고로 남긴다(design D-23, `progress.md` §G).

### Out of Scope — 와이어 식별자의 개명

- `push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree` 같은 구조 식별자는 호환을 위해 이 카드에서 개명하지 않는다. 스윕 가드는 이들을 별도 차선으로 추적한다(`design.md` §D-12).

### Out of Scope — 인접 카드의 파일

- t1448(레인-4)이 편집하는 `kanban-dispatch*.md`·`gtd.md`·`auto-semantics.md`·`manager-todo.md` 는 t1448 이 develop 에 착지한 뒤에만 이 카드의 규칙 편집이 흡수한다.
- t1452 의 카드 본문과 `SPEC-LATE-BRANCH-REDESIGN-001` 의 파일은 이 SPEC 이 수정하지 않는다. 처분은 권고로만 남긴다(`plan.md` §G).

### B.3 변경 계층 (두 계층과 퇴역 뒤 정리)

| 계층 | 정의 | 예 | 병합 시점 |
|---|---|---|---|
| **PRE-CUTOVER-SAFE** | 현 git-flow 구성에서 동작이 바뀌지 않고 github-flow 구성에서만 새 동작을 켜는 가산·설정 게이트 변경 | 기준 브랜치 해석 이름 교체, 새 PR 전달 간선, 착지 판정의 squash 대응, rc 릴리스 스크립트 기구, 스윕 가드의 픽스처 자가 시험 | 언제든 develop 에 병합 |
| **CUTOVER-TIME** | 일찍 병합하면 거짓이 되는 문장과 설정 | 규칙·문서·에이전트·스킬·README·docs-site·템플릿 산문, 이 저장소의 `git-strategy.yaml`·`workflow.yaml` 값, `.coderabbit.yaml`, `spec-lint.yml` 의 develop 논리, `hns-release-specialist`(매트릭스 스위치 호출), `AGENTS.local.md`, 스윕 가드의 트리 단언 | 카드 브랜치에서 준비·검증하고 병합 가능 상태로 보류, 절체 경계에서 한 묶음으로 적용 |
| **퇴역 뒤 정리** | develop 이 사라진 뒤에야 거짓이 되는 설정 | 워크플로 push 트리거 목록의 `develop` | develop 퇴역 이후 런북의 정리 단계(design D-13·D-17). 그때까지 develop 팁 CI 가 계속 돌아야 한다 |

## §C 요구사항

> 요구사항은 GEARS 형식이며 한 개 이상의 인수 기준이 `acceptance.md` 에서 `maps REQ-…` 로 이를 덮는다. 상한은 Tier L 의 25개이며 이 SPEC 은 22개를 둔다.

### REQ-GFD-001 — 구성된 통합 목표로의 기준 해석

**Where** `git_strategy` 의 활성 프로필 `workflow` 가 `github-flow` 일 때, 카드 전달·워크트리 기준·착지 판정·통합 창 해석은 기준 브랜치로 develop 이 아니라 구성된 통합 목표(`main`)를 사용해야 한다(shall). **Where** `workflow` 가 `git-flow` 이면 현행 동작은 바뀌지 않아야 한다(shall).

### REQ-GFD-002 — squash 병합에 안전한 착지 판정

**When** `moai worktree done`·`moai worktree sweep` 이 카드 브랜치의 착지를 판정할 때, 판정은 조상 관계, 병합 기준점 이후 누적 변경의 patch-id 가 통합 목표에 이미 있는 커밋과 일치하는지, 해당 브랜치 PR 의 병합 상태의 세 층을 이 순서로 시도해야 하고(shall), 어느 것도 확정하지 못하면 트리를 보존해야 한다(shall). **When** 세션 종료 정리가 카드 브랜치의 착지를 판정할 때, 판정은 앞의 두 층(조상 관계, 누적 patch-id)만 시도해야 하고(shall), 네트워크 호출(`gh`)을 하지 않아야 하며(shall not), 두 층이 확정하지 못하면 트리를 보존해 다음 `sweep` 의 세 번째 층이 판정하게 해야 한다(shall).

### REQ-GFD-003 — 기본값의 리터럴 develop 제거

**When** `moai worktree new`·`moai worktree sweep`·카드 diff 범위 계산이 기준 브랜치의 기본값을 정할 때, 기본값은 구성된 통합 목표에서 유도돼야 하고(shall) 리터럴 `develop` 을 담지 않아야 한다(shall).

### REQ-GFD-004 — 카드 PR 전달 간선

**While** `workflow` 가 `github-flow` 인 동안, `moai factory complete` 는 merge-ready 카드를 카드 브랜치 push 와 통합 목표 대상 PR 개설로 진행시켜야 하고(shall), PR 이 병합 완료로 관측된 뒤에만 카드를 병합 완료로 기록해야 하며(shall), 로컬 병합이나 병합 창 획득을 요구하지 않아야 한다(shall not).

### REQ-GFD-005 — PR 개설 전 병합 준비 점검

**When** 카드가 PR 개설을 앞두고 있을 때, 병합 준비 점검(sync 감사 PASS·통합 목표 대비 병합 충돌 없음·트리 항등)은 통합 목표를 기준으로 실행돼야 하고(shall), 하나라도 실패하면 PR 을 열지 않고 사유를 보고해야 한다(shall).

### REQ-GFD-006 — 로컬 병합 창은 전달의 선행 조건이 아니다

**Where** `workflow` 가 `github-flow` 일 때, 로컬 통합 창(`moai integration acquire`·`release`)은 카드 전달의 선행 조건이 아니어야 하고(shall not), `moai slot` 자원 임대는 변함없이 동작해야 한다(shall). git-flow 프로젝트의 창 동작은 보존돼야 한다(shall).

### REQ-GFD-007 — main HEAD 에서 태그

**When** 릴리스 하네스가 rc 또는 정식 태그를 만들 때, 태그 대상 커밋은 `origin/main` 의 HEAD 와 같아야 하고(shall), 릴리스 스크립트는 그 커밋에 놓인 detached HEAD 워크트리를 허용하고 그 밖의 HEAD 는 거부해야 한다(shall).

### REQ-GFD-008 — rc 태그의 출처 검증 규칙

**When** 태그가 prerelease 접미사(`-rc.N`)를 가질 때, 릴리스 출처 검증은 검사 5(CHANGELOG 절)와 검사 6(`system.yaml` 버전)을 건너뛰어야 하고(shall) 그 자리를 채우는 대체 검사는 두지 않아야 하며(shall not), 검사 1~4(주석 태그·트레일러·트레일러 버전 일치·커밋 결속)와 검사 7(main 조상)은 유지해야 한다(shall). 접미사 없는 태그의 7개 검사는 바뀌지 않아야 한다(shall).

### REQ-GFD-009 — prerelease 표시

**When** GoReleaser 가 접미사 있는 태그로 GitHub 릴리스를 만들 때, 그 릴리스는 prerelease 로 표시돼야 한다(shall).

### REQ-GFD-010 — 태그 전 3-OS 매트릭스

**Where** 릴리스 하네스가 `scripts/release.sh` 를 `--require-matrix-run` 옵션과 함께 호출할 때, 태그 대상 SHA 에서 `workflow_dispatch` 로 돌린 워크플로 `Release PR Multi-OS Verification` 실행의 세 OS 레그 성공 기록이 있어야 하고(shall), 기록이 없으면 릴리스 스크립트는 태그를 만들지 않아야 한다(shall not). 옵션이 없으면 스크립트의 현행 동작은 바뀌지 않아야 한다(shall). 하네스 본문이 그 옵션을 넘기도록 바꾸는 일은 절체 묶음(REQ-GFD-016)이 맡는다.

### REQ-GFD-011 — 낡은 develop 기준 서술 잔존 0

**When** 절체 변경 묶음이 적용된 뒤, 범위 표면(design D-9 가 하위 트리 열 곳과 파일 수 바닥값으로 열거한다)에는 develop 을 기준·통합 브랜치로 서술하는 살아 있는 문장이 남지 않아야 한다(shall not). 역사적 서술은 날짜나 카드·SPEC 식별자를 곁들인 폐기 표지를, git-flow 옵션 서술은 줄 전체를 지목하는 검토된 허용 목록 항목을 가져야 한다(shall).

### REQ-GFD-012 — 전달 경로 서술의 일치

**Where** 지침·문서 표면이 카드 전달을 서술할 때, 서술은 "카드 브랜치, push, PR, CI 재측정, GitHub 병합(main)" 경로와 일치해야 하고(shall), 4개 로케일 문서는 문장 수의 일치가 아니라 의도의 일치를 기준으로 갱신돼야 한다(shall).

### REQ-GFD-013 — Frozen 조항의 정규 개정

**Where** 변경이 zone-registry 에 등재된 `[ZONE:Frozen]` 조항 `CONST-V3R5-027`·`-028` 의 문언을 바꿀 때, 개정은 조항마다 `moai constitution amend --rule <ID> --evidence …` 한 번씩, 정확히 두 번의 실행으로 5단 게이트를 통과해야 하고(shall — 마지막 인간 승인 층은 대화형 Y/N 이므로 운영자가 수행한다), 개정 뒤 `moai constitution validate` 가 종료 코드 0 이며 등재 문언과 `spec-workflow.md` 본문이 일치해야 한다(shall). `[ZONE:Frozen]` 표지 줄이 등재되어 있는지는 편집하는 줄마다 `moai constitution list --file <경로>` 로 판정해야 하고(shall), 등재되어 있지 않은 줄 — `worktree-integration.md` 의 표지 두 줄과 `spec-workflow.md` 의 Route A/B 문단 표지 줄·plan 단계 표지 줄 — 은 일반 편집으로 바꾸되 개정된 등재 문언과 같은 방향이어야 하며(shall) 로컬과 템플릿 사본이 같아야 한다(shall).

### REQ-GFD-014 — 상시로드 증가의 진술과 측정

**When** 한 편집이 상시로드 파일을 1,000 바이트 넘게 늘릴 때, 변경 설명은 `rule-authoring.md` 가 요구하는 진술을 담아야 하고(shall), 변경 전후의 상시로드 파일 수와 바이트가 측정돼 보고돼야 한다(shall).

### REQ-GFD-015 — 로컬과 템플릿 사본의 동반

**While** 대상 문서가 로컬과 템플릿 두 사본을 가질 때, 변경은 템플릿 우선 순환으로 두 사본에 함께 적용돼야 하고(shall) 이미 의도적으로 분기한 지점은 보존돼야 한다(shall).

### REQ-GFD-016 — 저장소 설정과 CI 의 절체 시점 동반

**When** 절체 변경 묶음이 적용될 때, 이 저장소의 `git-strategy.yaml`·`workflow.yaml` 값, `.coderabbit.yaml` 의 `base_branches`, `spec-lint.yml` 의 develop 의존, 릴리스 하네스 본문의 `--require-matrix-run` 호출은 함께 github-flow 기준으로 바뀌어야 하고(shall), 워크플로 push 트리거 목록의 `develop` 은 develop 이 퇴역할 때까지 남아 develop 팁 CI 가 계속 관측되어야 하며(shall), 이 모두는 절체 전에는 바뀌지 않아야 한다(shall not).

### REQ-GFD-017 — t810 범위의 흡수

**Where** 이 SPEC 이 `SPEC-LATE-BRANCH-REDESIGN-001` 의 범위를 흡수할 때, 그 REQ-LBR-001..006 의 의도(SPEC 당 PR 1개·plan 단계의 워크트리 진입·폐기 조건 단일화·Step 3.3.5 은퇴·Frozen 정규 개정·두 사본 동반)는 문서 마일스톤의 인수 기준으로 충족돼야 한다(shall).

### REQ-GFD-018 — main 과 develop 의 수렴

**When** main 을 단일 기준으로 만들기 위해 두 브랜치를 수렴시킬 때, 절차는 develop 이 `origin/main` 을 흡수한 뒤 develop 에서 main 으로의 PR 을 병합 커밋으로 병합해야 하고(shall), 병합 후 main 의 트리가 병합 직전 develop 팁의 트리와 같음을 확인해야 하며(shall), 병합 전에 되돌리는 경로를 기록해야 한다(shall).

### REQ-GFD-019 — 배치 경계와 레인 재기동

**While** 병합되지 않았거나 병합됐어도 push 되지 않은 picked 카드, 살아 있는 통합 창·슬롯 보유자, 활성 레인 세션 중 하나라도 남아 있는 동안, 기준 브랜치 절체는 시작되지 않아야 하고(shall not), 절체 절차는 레인 정지·`/clear`·정리, 재기동, 첫 카드 관측의 순서와 단계마다의 실행 주체(카드·리더·운영자) 및 외부 공유 시스템 여부, develop 퇴역 뒤 워크플로 push 트리거 잔존 점검을 담아야 한다(shall). 이 카드 자신(M4·M5 묶음을 런북 2단계에서 마지막으로 병합하는 절체 카드)은 사전 점검의 picked 카드 조건에서 그 병합·push 뒤에 제외돼야 하고(shall), 점검은 그 병합 뒤 수렴 병합(런북 4단계) 앞에서 실행돼야 한다(shall, design D-25).

### REQ-GFD-020 — 외부 공유 시스템 변경의 확인 경계

**When** 런북이 외부 공유 시스템을 바꾸는 단계(main 보호 규칙과 필수 체크 목록 변경, develop 보호·삭제, 기준 브랜치 절체의 실행)를 담을 때, 이 카드를 수행하는 레인은 그 단계를 리더의 확인 기록 없이 실행하지 않아야 한다(shall not).

### REQ-GFD-021 — 스윕 가드의 관측된 실패

**When** develop 기준 서술 스윕 가드가 실행될 때, 가드는 알려진 실패 입력(살아 있는 develop 문장 — 분기 말 토큰이 없는 문장 `Work starts from develop.` 도 허용 목록에 없으면 포함)에서 적색이어야 하고(shall), 폐기 표지 문장과 줄 전체 허용 항목에서는 녹색이어야 하며(shall), CJK 인접 표기(`develop에서`)와 접두 표기(`origin/develop`, `develop-based`)를 포착해야 하고(shall), 빈 스윕이 통과로 읽히지 않도록 방문 수가 하위 트리별·전체 바닥값(design D-9 가 수로 정한다) 이상임을 단언해야 한다(shall).

### REQ-GFD-022 — 병합 큐 판정의 기록과 t1452 조정

**Where** 병합 큐의 채택을 판정할 때, 결정은 다시 측정되어 재현된 CI 사용량·필수 체크 `strict` 설정·병합 PR 대상 분포를 증거 표로 기록해야 하고(shall), 재현되지 않았거나 측정하지 못한 항목은 후속 측정 항목으로 Gaps 에 남겨야 한다(shall). 로컬 병합 창의 은퇴로 무의미해지는 t1452 항목과 유효하게 남는 항목은 명시돼야 한다(shall).

## §D 결정 요약

각 결정의 근거·기각된 대안은 `design.md`, 측정은 `research.md` 에 있다. 아래는 plan-auditor 가 반박할 수 있는 입장이며, 리더 소관 표시(L)는 plan-audit 의 판정 대상이 아니라 리더가 정한다.

| ID | 입장 | 리더 소관 |
|---|---|---|
| P1 | git-flow 는 기본값에서만 은퇴한다. 선택 가능한 워크플로로 남고, 휴면 코드 삭제는 후속 인벤토리다 | |
| P2 | 변경을 PRE-CUTOVER-SAFE 와 CUTOVER-TIME 두 계층으로 나눈다 | |
| P3 | 카드는 카드마다 PR 로 전달하고 병합 큐는 보류한다(측정 후속 항목은 리더가 카드로 발행). 운영자 결정(D-20) | L |
| P4 | 카드 전달에서 로컬 병합 창은 은퇴, `moai slot` 은 유지. t1452 는 (c) 와 PR 전 병합 준비 점검으로 좁힌다. 리더 결정(D-11) | L |
| P5 | t1453 이 t810 을 흡수한다. t810 의 카드·워크트리와 `SPEC-LATE-BRANCH-REDESIGN-001` 은 건드리지 않고, t810 의 종결 처분은 이 카드가 닫힐 때 운영자에게 올린다. 리더 결정(D-19) | L |
| P6 | 릴리스는 main HEAD 에서 태그한다. rc 태그는 검사 5·6 을 건너뛰고 1~4·7 을 유지한다(R-a, 기본값, 리더가 뒤집을 수 있음, D-16). 3-OS 게이트는 옵션 B — `Release PR Multi-OS Gate` 는 main 필수 체크에서 빼고(운영자 수행) 태그 직전 `workflow_dispatch` 를 `release.sh --require-matrix-run` 이 강제한다. 운영자 결정(D-22·D-24) | L (옵션 B 와 R-a) |
| P7 | 수렴은 develop 이 main 을 흡수한 뒤 병합 커밋 PR, 트리 항등 검증, develop 퇴역은 단계화. 절체의 선행 조건은 origin/develop 팁 CI 녹색이며 push 트리거의 develop 은 퇴역 뒤 정리한다. 수렴 PR 은 하나로 먼저 시도한다(D-17·D-18) | L (퇴역 시점, 분할 폴백 선택) |
| P8 | 스윕 가드는 `internal/template` 패키지에 두고 패턴·허용 목록·돌연변이 프로브를 갖는다 | |
| P9 | 4로케일은 의도 기준으로 갱신하고 zh 추가 문장은 doc 마일스톤의 zh 담당이 해소한다 | |
| P10 | 상시로드 증가는 진술과 전후 측정을 doc 마일스톤의 인수 기준으로 둔다 | |

## §E 인접 카드와의 관계

- **t810** — `SPEC-LATE-BRANCH-REDESIGN-001` 의 B1~B6 와 REQ-LBR-001..006 이 이 SPEC 의 문서 마일스톤(M4)에 인수 기준으로 들어온다. 큐 기록은 두 카드가 서로를 흡수한다고 적고 있었으나(`research.md` §6) 리더가 방향을 정했다 — t1453 이 t810 을 흡수한다(D-19). t810 의 카드·워크트리·SPEC 파일은 건드리지 않고, 닫는 처분은 이 카드가 닫힐 때 운영자에게 올린다.
- **t1452** — 병합 창 단축 카드. (a)·(b)는 로컬 창 은퇴로 무의미, (c)는 유효. 리더가 t1452 를 (c) 와 PR 전 병합 준비 점검으로 좁히기로 했고, 카드 본문 수정은 리더 몫이다. 이 SPEC 은 그 전제로 진행한다(D-11).
- **t1448** — 같은 규칙 파일을 편집한다. M4 의 규칙 편집은 그 착지 뒤에 흡수한다.

## §F 측정 핀

모든 측정은 카드 트리 `4bf547bcad7c155b1e91485921569db709ec3ac2` 와 원격 팁 `284e09c44023598affe486f17701717ca173e6ca`(develop)·`4755c5e506225ba90b7a303c5763fa303c699492`(main)에 핀한다. 원격 팁은 2026-10-02 에 가져온 값이며 움직이는 참조다. 이 SPEC 의 어떤 판정도 팁이 움직인 뒤에 다시 인용하지 않는다 — 인용하려면 다시 측정하고 다시 핀한다. 증거 원장 E-01~E-25 는 위 카드 트리에, 개정 0.1.1 에서 더한 E-26~E-35 와 측정 행 M-n 은 계획 커밋 `855563dba79da74528e0f01560efe80f1016cc11` 에, 개정 0.1.2 에서 더하거나 다시 잰 E-36 이후와 M-1·M-3·M-6·M-9 는 계획 개정 2 커밋 `6c2277295d9ddeaa92e83c0225910445f83cc1c2` 에 핀한다. 세 트리는 SPEC 디렉터리의 6개 파일만 다르다(`acceptance.md` §B 머리말의 관측).

---

🗿 MoAI
