---
id: SPEC-WORKTREE-GC-001
title: "Legacy worktree comprehensive audit and safe disposal — acceptance criteria"
version: "0.1.1"
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
module: ".claude/worktrees"
tags: "worktree,cleanup,disposal,evidence-export,operations,gc"
---

# SPEC-WORKTREE-GC-001 — Acceptance Criteria

## §A — 검증 원칙

- 모든 AC는 Given-When-Then 형식이고 이진 판정 가능해야 한다.
- 모든 PASS 행은 실제 실행한 명령 + verbatim 출력에 귀속된다(`verification-claim-integrity.md` §1-§3). 빈 sweep(0행 집계)은 판정으로 읽지 않는다.
- RED-now 셀: OPERATIONS 카드 특성상 사전 RED 실패 관측은 런 페이즈 M1의 시작 측정이 담당한다(측정 전 인벤토리는 아직 존재하지 않음 — M1이 그 부재를 최초 관측한다).

## §B — Quality Gate 기준

- plan-auditor PASS 임계: 0.80 (Tier M).
- 런 페이즈 게이트: E1-E7 자가검증 항목 전부 명령+출력 귀속 충족.
- 제외 무결성 위반 1건이라도 있으면 전체 FAIL(가역성 없는 파괴 축이므로 부분점수 없음).

## §C — Traceability

| AC | 커버 REQ | 마일스톤 |
|----|---------|---------|
| AC-WGC-001 | REQ-WGC-004, REQ-WGC-005, REQ-WGC-011 | M2, M4 |
| AC-WGC-002 | REQ-WGC-006, REQ-WGC-012 | M2, M4 |
| AC-WGC-003 | REQ-WGC-007 | M2 |
| AC-WGC-004 | REQ-WGC-002 | M5 |
| AC-WGC-005 | REQ-WGC-001 | M1, M5 |
| AC-WGC-006 | REQ-WGC-008(a) | M2 |
| AC-WGC-007 | REQ-WGC-008(c), REQ-WGC-011 | M2, M4 |
| AC-WGC-008 | REQ-WGC-008(b) | M2 |
| AC-WGC-009 | REQ-WGC-013 | M4 |
| AC-WGC-010 | REQ-WGC-009, REQ-WGC-010 | M3 |
| AC-WGC-011 | REQ-WGC-015 | M4 |
| AC-WGC-012 | REQ-WGC-003, REQ-WGC-014, REQ-WGC-016 | M1-M5 |
| AC-WGC-013 | REQ-WGC-011 | M4 |

## §D — AC Matrix (Given-When-Then)

- **AC-WGC-001 (T1 종단)** — Given `git fetch origin develop`이 선행 실행돼 갱신된 `origin/develop` SHA가 classification.md에 기록돼 있고, 로컬 develop에 병합됐고 fetch 후 `git merge-base --is-ancestor <branch> origin/develop`이 종료 0인 트리와, 완료된 rescue 반출 기록이 있을 때 / When 런 주체가 T1 처치를 실행하면 / Then 트리가 `git worktree list`에서 사라지고, 분기 ref가 삭제되고(`git rev-parse --verify refs/heads/<branch>` 종료 0 아님), removal-log에 병합·push·반출·제거 네 근거가 한 행에 기록되며, classification.md의 각 origin 근거 행에 fetch 시점의 `origin/develop` SHA가 기록돼 있다.
- **AC-WGC-002 (T2 종단 — ref 보존)** — Given 미병합 트리와 REQ-WGC-006을 충족하는 판정 기록 경로(큐 카드 본문 / memory 파일 / `.moai/reports/<card-id>/verdict.md` / 리드 attested 기록)가 있을 때 — 기록은 대상 분기/커밋을 명명하면서 처치-허용 방향(통합 안 함, 폐기 승인, 정리 대상)을 진술해야 한다 — / When T2 처치를 실행하면 / Then 트리는 제거되고, 분기 ref는 존속하고(`git rev-parse --verify refs/heads/<branch>` 종료 0), removal-log의 해당 행에 판정 기록 경로와 방향이 인용되며, 방향이 유지/보류인 기록을 가진 트리는 T2로 분류되지 않고 유지·목록 보고된다(REQ-WGC-006 방향 규칙).
- **AC-WGC-003 (T3 유지+보고)** — Given 미병합이고 판정 기록이 없는 트리가 있을 때 / When 분류가 완료되면 / Then 그 트리는 `git worktree list`에 존속하고, classification.md에 T3 행(사유)이 기록되며, verdict.md의 유지 목록에 나타난다.
- **AC-WGC-004 (제외 무결성)** — Given 제외 집합 16트리(t1051, t1060, t1065, t1067, t1068, t1071, t1073, t1074, t1078, t1079, t1080, t1081, t1083, t1084, develop 워크트리, primary)가 있을 때 / When 런 페이즈 전체가 끝나면 / Then 16트리 전부 `git worktree list`에 존속하고(트리당 존재 확인 명령 16회 관측), 집산식 `시작행수 = 제거수 + T3유지수 + 제외수` 가 verdict.md에 검산된다.
- **AC-WGC-005 (측정 쌍)** — Given 런 페이즈가 시작되려 할 때 / When M1과 M5가 각각 실행되면 / Then `git worktree list | wc -l`과 `du -sh .claude/worktrees`가 각각 시작 1회·종료 1회 기록되어 총 2쌍이 존재하고, 종료 행수 < 시작 행수이면 제거수와 일치한다.
- **AC-WGC-006 (더티 tracked 파일)** — Given `git status --porcelain`에 tracked 수정 행이 있는 트리가 있을 때 / When 분류가 실행되면 / Then 그 트리는 판정 조건이 T1/T2를 가리켜도 T3 처리(유지)되거나 명시적 운영자 안건으로 상신되고, removal-log에 제거 행이 없다.
- **AC-WGC-007 (lock 트리)** — Given git lock이 걸린 트리가 있을 때 / When 처치가 실행되면 / Then lock은 그 자체로 유지 판정이 아니고, `git worktree unlock`은 단계 ∈ {T1, T2}와 반출 검증 완료와 점유 부재가 모두 기록된 뒤에만 발화하며, unlock 실패 또는 세션 소유 lock 트리는 유지되고 blocker 안건으로 상신된다(REQ-WGC-008(c)·REQ-WGC-011 동일 판정식).
- **AC-WGC-008 (세션 점유)** — Given `moai session list`에 트리 경로를 cwd로 갖는 생존 PID 세션이 있을 때 / When 분류가 실행되면 / Then 그 트리는 단계 무관 유지되고, classification.md에 점유 사유가 기록된다.
- **AC-WGC-009 (미해소 T2 판정 → blocker)** — Given 어떤 트리가 "T2다"는 주장과 함께 배차됐지만 REQ-WGC-006의 어느 기록으로도 해소되지 않을 때 / When 분류가 실행되면 / Then 그 트리는 제자리에 유지되고, blocker 항목(트리명, 주장 출처, 미해소 사유)이 E7 보고에 올라가며, 조용한 제거와 조용한 T3 재분류가 모두 부정된다.
- **AC-WGC-010 (반출 트리아지·목적지)** — Given 제거 예정 트리의 미추적 파일 집합이 있을 때 / When M3 반출이 실행되면 / Then 증거류(`.moai/`, `.claude/`, `docs/`의 리포트·로그·설정)가 `.moai/reports/t1084/rescue/<tree-name>/` 아래에 존재하고, 판정서가 인용하는 경로가 같은 반출에 동반 존재하며, 폐기류(빌드 출력/캐시/node_modules)는 스킵 목록+건수로 보고되고, 상한(파일당 10 MB, 트리당 200 MB) 초과 파일마다 skipped-by-rule 행(트리, 경로, 크기)이 존재하며 증거류 초과분은 운영자 주의 목록에 나타나고, primary·제외 트리에는 새 파일이 없다.
- **AC-WGC-011 (prune 순서)** — Given 제거들이 완료됐을 때 / When `git worktree prune`이 실행되면 / Then prune은 제거 기록 이후 타임스탬프이고, prune 전후 `git worktree list` diff가 기록돼 있다.
- **AC-WGC-012 (판정 형식)** — Given 런 페이즈가 완료됐을 때 / When verdict.md가 검토되면 / Then 5-섹션(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)이 모두 존재하고, 모든 removed/PASS 행이 명령 출력을 인용하며, 측정 중 트리 내부에 기록된 런 산출물이 0개다.
- **AC-WGC-013 (배치 상한)** — Given 제거 대상이 25트리를 초과할 때 / When M4가 실행되면 / Then 각 제거 창은 25트리 이하이고, 창 사이마다 `git worktree list` 재나열 + 점유 재탐사 + listing diff 기록이 존재하며, `du` 재측정은 시작·종료 1회씩뿐이다.

## §E — Edge Cases 요약

- 로컬 develop에만 병합(origin push 대기) 트리 → T1 아님, T3 유지 + "pending lead bulk push" 사유(REQ-WGC-005).
- gh 부재/오류로 병합 미판정(no answer) → REAPER REQ-WR-002/003의 폴백·보존 논리 재사용, 판정 불가 트리는 T3 유지.
- detached-HEAD 트리 → 분기 없음, ref 판정 불가 → T3 유지 + 보고.
- 이미 사라진 트리(등록만 남음) → t528 교훈 — listing diff로 무음 소실을 기록하고, 본 런의 제거로 계산하지 않는다.
- 유지-방향 판정 기록 보유 트리(t1050, t1064, t810 — 이번 세션 존재 확인) → 처치 집합에서 완전히 제외되어 유지되고 기록 인용과 함께 목록 보고된다(REQ-WGC-006 방향 규칙; D9 완화).

## §F — Definition of Done

- AC-WGC-001~013 전부 PASS(명령+출력 귀속) 또는 명시적 blocker 상신.
- 시작/종료 측정 쌍 2쌍 기록 완료.
- 제외 무결성 16/16 존속 + 집산식 검산 일치.
- verdict.md 5-섹션 형식 완비, Gaps 섹션에 미관측 항목 명시.
- rescue/ 반출물이 본 카드 분기에 커밋됨(실측 `WT-legacy-cleanup` @ `9064d19aa`; 명칭 변천 경위는 plan.md §A 참조. 유일본 보호).
