---
id: SPEC-WORKTREE-GC-001
title: "Legacy worktree comprehensive audit and safe disposal: three-tier disposition with evidence-first untracked export (~530 trees)"
version: "0.1.2"
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/worktrees"
lifecycle: spec-anchored
tags: "worktree,cleanup,disposal,evidence-export,operations,gc"
tier: M
---

# SPEC-WORKTREE-GC-001

## §A — History

- **2026-09-22** — plan-phase v0.1.0 authored from card t1084 (kanban lead dispatch, operator instruction 2026-09-22). OPERATIONS card: run phase executes the audit/cleanup as lane work; ships no Go product code. Baseline measured this session (2026-09-22, worktree t1084): 544 worktrees registered (`git worktree list`, 543 card trees under `.claude/worktrees/` + 1 primary checkout), `du -sk .claude/worktrees` = 144,978,644 KB (~138.3 GB), local `develop` = `7f86971fc`, `origin/develop` = `f5fff2190` (lead bulk push pending — local ahead). The lead's 09-22 first pass already removed 9 trees (t1055, t1056, t1059, t1066, t1070, t1077, t1057, t1069, goal-dist) and rescued a ~210-directory report set from the t1055 tree before removing it — the precedent this SPEC generalizes. Tier M chosen: ~0 product LOC but ≥3 operational milestones at destructive scale (~530 trees); Tier M grants acceptance.md + 0.80 plan-audit threshold without Tier L's design/research overhead. Prior art SPEC-WORKTREE-REAPER-001 (v0.4.1, completed) reused, not re-implemented — relation recorded in plan.md § REAPER Prior-Art Relation.

- **2026-09-22** — plan-phase v0.1.1 repair after plan-audit iteration 1 (FAIL 0.8125, MP-7 + D2-D9). Orchestrator resolved the 3 clarification topics (export caps 10 MB/file·200 MB/tree, T2 record direction-classification, 25-tree batch ceiling); manager-spec folded them into REQ-WGC-006/008/010/011 and added AC-WGC-013. REQ numbering held stable (no renumber). Repair deltas: D2 fetch discipline AC coverage (AC-WGC-001 extended), D3 branch identity corrected to measured `WT-legacy-cleanup` @ `9064d19aa` with the honest rename sequence, D4 disposal-direction rule (keep-direction records t1050/t1064/t810 remove trees from the disposition set), D5 unified lock predicate (lock is not a verdict), D6 T2 declared an explicit operator-record-bound exception to AGENTS.md §3, D7 AC-WGC-007 citation corrected, D8 labels relabeled Event-driven, D9 keep-record pre-flight cross-check added to plan.md §C.

- **2026-09-22** — plan-phase v0.1.2 micro-repair after plan-audit iteration 2 (PASS-WITH-DEBT 0.9375; D10-D12, single micro-edit per the auditor's debt-routing note). D10 (major): plan.md §C keep-record sweep re-pointed from the dormant store glob `~/.claude/projects/*/memory/` (measured 0 matches) to the live CLAUDE_CONFIG_DIR profile store `~/.moai/claude-profiles/*/projects/*/memory/`, `head -20` truncation removed (truncation reads as absence), and a positive control added — the sweep MUST re-find the t1050/t1064/t810 keep-records, and an empty output is a measurement failure (미측정), never "no keep-records". D11 (minor): REQ-WGC-005 keep-record boundary — a T1-classified (merged+pushed) tree carrying a keep/pending-direction operator record routes to operator escalation or premise-resolution citation (records exported first regardless; hazard bounded). D12 (minor): REQ-WGC-011 + plan M4 — `git worktree remove --force` authorized only when the tree's M3 export-log row exists (no export row, no force flag), matching the C4 lock-predicate gating style. REQ ids held stable (001-016, no renumber).

## §B — Problem

`.claude/worktrees/` 아래 약 530개의 구배치 카드 워크트리가 남아 있다(2026-09-22 기준 등록 544 중 이번 배치 잔여 제외). 트리 제거는 커밋을 잃지 않는다 — 오브젝트는 공유 `.git` 소관이고, 실제 손실 축은 트리 고유의 **미추적 파일**(리포트, 로그, 설정 흔적)뿐이다. 09-22 1차 정리에서 t1055 트리에서 primary 결손 보고서 약 210 디렉터리를 구출한 실측이 이 전제를 확인했다. 그러나 트리마다 상태가 다르다: 완전 병합·원격 push까지 끝난 트리, 미병합이지만 운영자 판정(예: t154 「통합 안 함」)이 존재하는 트리, 판정도 없는 미병합 트리. 일괄 제거는 세 경우를 구분하지 않고 세 번째 경우의 유일한 사본을 파괴할 수 있다. 또한 원격 ref를 fetch 없이 읽으면 자기 staleness를 재는 것이고(t1059 교훈), 측정 대상 트리 안에 출력을 쓰면 측정 자체가 무효화되며(t1069 교훈), 판정서만 반출하고 그것이 인용하는 산출물을 따로 반출하지 않으면 반출이 불완전하다(t1009/t1022 교훈).

## §C — Goal

전체 구배치 워크트리를 트리 단위 3단(T1/T2/T3) 판정으로 통합 점검하고, 모든 제거에 앞서 트리 고유 미추적 증거를 본 카드 워크트리로 반출한 뒤, 안전한 트리만 제거한다. 제거 불가 트리는 유지하고 목록 보고한다. 전후 측정(트리 행수, 디스크 사용량)으로 결과를 정량 입증한다.

## §D — Requirements (GEARS)

요구 16건. 주어 `<subject>`는 run 페이즈 실행 주체(이하 "런 주체"). 식별자·명령어는 영문 그대로 둔다.

### D.1 — M1: 인벤토리·측정 (Inventory)

- **REQ-WGC-001** (Ubiquitous) — The run 주체 shall 런 시작 시 1회와 런 종료 시 1회, `git worktree list` 행수와 `du -sh .claude/worktrees` 결과를 각각 측정 기록으로 남긴다(기록 위치: `.moai/reports/t1084/`).
- **REQ-WGC-002** (Ubiquitous) — The run 주체 shall 제외 집합(t1051, t1060, t1065, t1067, t1068, t1071, t1073, t1074, t1078, t1079, t1080, t1081, t1083, t1084 본 카드 트리, `develop` 워크트리, primary 체크아웃 — 총 16개)에 속한 트리를 단계 분류 대상에서도, 반출 대상에서도, 제거 대상에서도 완전히 배제한다.
- **REQ-WGC-003** (Ubiquitous) — The run 주체 shall 측정 중인 트리 안에는 어떤 출력·파일도 기록하지 않는다(t1069 교훈 — 측정 출력이 측정 집합에 들어가면 집합이 움직인다). 모든 런 페이즈 산출물은 본 카드 워크트리(`.claude/worktrees/t1084`) 안에 둔다.

### D.2 — M2: 3단 판정 (Classification)

- **REQ-WGC-004** (Event-driven) — **When** 임의 트리의 단계 판정에 `origin/*` ref를 읽는 일이 필요하면, the run 주체 shall 판정에 앞서 `git fetch origin develop`을 먼저 실행한다(t1059 교훈 — fetch 없이 읽은 원격 ref는 자기 staleness를 잰다).
- **REQ-WGC-005** (State-driven) — **While** 트리 분기 팁이 `git merge-base --is-ancestor <branch> origin/develop`으로 측정해 `origin/develop`의 조상임이 확인되면(fetch 후 측정), the run 주체 shall 그 트리를 T1으로 분류한다. 로컬 `develop`에만 병합되고 `origin/develop`이 뒤처진 경우(리드 일괄 push 대기)는 T1이 아니며 T3으로 유지·보고한다(보류 사유: pending lead bulk push). D11 경계 — T1으로 분류된 트리에 유지·보류 방향의 운영자 판정 기록(`폐기 금지`/`보존` 등)이 지명 표면에서 읽히면, the run 주체 shall 그 트리를 운영자 에스컬레이션으로 상신하거나 제거 행에서 기록 경로와 해소된 전제(원격 착지 — remote landing)를 함께 근거 인용하게 한다(keep-record는 REQ-WGC-009대로 무조건 선반출되며 위험은 유계다).
- **REQ-WGC-006** (State-driven) — **While** 트리 분기가 미병합이지만 운영자 판정 기록이 (a) 지명 표면 — 큐 상태 카드 본문, memory `feedback_*`/`project_*` 파일, `.moai/reports/<card-id>/verdict.md`, 또는 배차 시 카드 진행 노트에 리드가 기록한 attested 기록 경로 — 에서 런 시점에 읽히고 (b) 대상 분기/커밋을 명명하면서 처치-허용 방향(통합 안 함, 폐기 승인, 정리 대상)을 진술하면, the run 주체 shall 그 트리를 T2로 분류하고 판정 기록의 경로와 방향을 판정 근거로 인용한다. 방향이 유지(폐기 금지, 보존)이거나 보류(게이트 대기, 승인 대기)인 기록은 트리를 처치 집합에서 완전히 제외시킨다 — 트리는 유지되고 기록 인용과 함께 목록 보고되며, 병합 상태가 운영자 유지 판정을 무시하지 못한다(이번 세션 존재 확인된 유지-방향 사례: t1050, t1064, t810). 방향이 해소되지 않거나 모호한 기록은 REQ-WGC-013 blocker 대상이다.
- **REQ-WGC-007** (State-driven) — **While** 트리 분기가 미병합이고 REQ-WGC-006의 어떤 판정 기록도 존속하지 않으면, the run 주체 shall 그 트리를 T3으로 분류해 트리를 유지하고 보고서에 목록화한다.
- **REQ-WGC-008** (Event-driven) — **When** 트리가 (a) 커밋되지 않은 tracked 수정(`git status --porcelain`의 tracked 변경 행)을 담고 있거나 (b) 살아 있는 세션이 트리에 앵커돼 있으면(`moai session list` 레지스트리 항목 + PID 생존 확인), the run 주체 shall 단계 판정 결과와 무관하게 그 트리를 T3 처리(유지+보고)하거나 명시적 운영자 안건으로 상신하며, 자동 폐기하지 않는다. (c) git lock은 그 자체로 판정이 아니다 — **While** 트리에 lock이 걸려 있으면, the run 주체 shall 그 트리를 자기 단계 판정에 따르게 한다: unlock+제거는 단계 ∈ {T1, T2}이고 반출이 검증됐으며 점유가 없을 때만 허용되고, unlock 실패 또는 세션 소유 lock은 유지 + blocker 안건으로 간다(REQ-WGC-011과 동일 판정식).

### D.3 — M3: 증거 반출 (Export)

- **REQ-WGC-009** (Ubiquitous) — The run 주체 shall 모든 트리 제거와 브랜치 ref 삭제의 선행 조건으로 해당 트리의 미추적 파일 증거 반출을 완료한다. 반출 목적지는 본 카드 워크트리의 `.moai/reports/t1084/rescue/<tree-name>/...` 이고, 반출 산출물은 `WT-legacy-cleanup` 브랜치에 커밋한다. primary 체크아웃이나 제외·타인 트리에는 절대 기록하지 않는다. 판정·리포트류 산출물을 반출할 때 그것이 인용하는 대상(첨부 로그, 인용 파일)을 함께 반출한다(t1009/t1022 교훈 — 판정서를 반출해도 그것이 인용하는 로그는 따라오지 않는다).
- **REQ-WGC-010** (Event-driven) — **When** 트리의 미추적 파일을 반출할 때, the run 주체 shall 증거류(`.moai/`, `.claude/`, `docs/` 아래의 리포트·로그·설정)와 폐기 가능 산출물(빌드 출력, 캐시, `node_modules` 등)을 분류하고, 규칙상 반출에서 제외한 항목은 건수와 함께 보고서에 나열한다. 반출 상한은 파일당 10 MB, 트리당 200 MB다 — 상한 초과 파일은 조용히 버려지지 않는다: 파일마다 skipped-by-rule 행(트리, 경로, 크기)을 보고서에 남기고, 증거류 초과 파일은 추가로 운영자 주의 항목으로 상신한다.

### D.4 — M4: 처치 (Disposition)

- **REQ-WGC-011** (Event-driven) — **When** T1 트리에 대해 단계 판정 근거(병합·push 증거 명령 출력)와 반출 완료 증거가 모두 검증되면, the run 주체 shall `git worktree unlock`(잠긴 경우에 한해) 후 `git worktree remove`로 트리를 제거하고, 분기 ref를 삭제한다. ref 삭제는 REQ-WGC-005의 완전 병합 판정이 있었기 때문에만 허용된다. lock 절차는 REQ-WGC-008(c)와 동일 판정식이다 — unlock+제거는 단계 ∈ {T1, T2}이고 반출 검증이 완료됐으며 점유가 없을 때만 발화하고, unlock 실패나 세션 소유 lock은 유지 + blocker로 간다. 제거 배치 상한은 창당 25트리다 — 창 사이마다 `git worktree list` 재나열, 점유 재탐사, listing diff 기록을 수행한다(t528 교훈 — 무음 메타데이터 드리프트 포착); `du`는 REQ-WGC-001대로 시작·종료 1회씩만 측정한다. D12: skipped-by-rule 잔여물이 남은 트리는 평문 `git worktree remove`가 거부하므로, `--force`는 그 트리의 M3 반출 로그 행(export-log.md 행)이 존재해 반출이 검증됐을 때만 허용된다 — 반출 행 없는 `--force`는 금지다.
- **REQ-WGC-012** (Event-driven) — **When** T2 트리에 대해 REQ-WGC-006의 판정 기록 인용과 반출 완료 증거가 모두 검증되면, the run 주체 shall 트리를 제거하되 분기 ref를 보존한다(통합 여부와 무관하게 판정 기록이 분기의 존속 가치를 보증한다).
- **REQ-WGC-013** (Event-driven) — **When** T2로 주장된 판정이 읽을 수 있는 기록으로 해소되지 않으면, the run 주체 shall 해당 트리를 리드에게 blocker 안건으로 상신하고 트리를 제자리에 유지한다. 조용한 제거도, 조용한 T3 재분류도 하지 않는다.

### D.5 — M5: 검증·보고 (Verification & Reporting)

- **REQ-WGC-014** (Ubiquitous) — The run 주체 shall 모든 제거를 한 행씩 기록한다: 트리명, 단계, 병합 근거 명령+출력, push 근거 명령+출력, 반출 근거 경로, 제거 명령 출력.
- **REQ-WGC-015** (Event-driven) — **When** 제거들이 완료되면, the run 주체 shall `git worktree prune`을 제거 뒤에만 실행하고, 그 전후 `git worktree list` 나열 diff를 기록한다(t528 교훈 — 워크트리는 내가 지우지 않아도 사라질 수 있다).
- **REQ-WGC-016** (Ubiquitous) — The run 주체 shall `.moai/reports/t1084/verdict.md`를 5-섹션 증거 판정 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk — `.claude/rules/moai/core/verification-claim-integrity.md` §3)으로 작성하고, 모든 PASS·removed 행이 실제 관측한 명령 출력에 대응하게 한다.

## §E — Non-Functional Constraints

- 부하 규율: 배치 안전 실행 — `du`(138 GB 규모)와 반출은 I/O 중량이라 소청크·순차로 제한하고, 백그라운드 부하를 만들지 않는다(정리 보장이 없는 무부하 스폰 금지).
- 직렬화: 제거·prune 창에서 다른 레인과의 충돌을 피하기 위해 통합 창 규율(`moai integration` 체계)과 무관하게, 본 카드 워크트리 안에서의 자기 커밋만 수행한다(레인은 push하지 않는다 — develop push는 리드 일괄).
- 판정 무결성: 모든 "병합됨/없음" 주장은 명령 실행 + 출력 관측을 근거로 한다(grep 텍스트 추론은 가설일 뿐이다).

## §F — Acceptance Overview

수용 기준 13건(AC-WGC-001~013)은 `acceptance.md` §D AC 행렬이 소유한다. 핵심 시나리오: T1 종단(반출→제거→병합 브랜치 ref 삭제), T2 종단(판정 기록 해소→반출→제거→ref 보존), T3 유지+보고, 제외 무결성(집산 일치), 측정 쌍, 더티 tracked 파일, lock, 세션 점유, 미해소 T2 판정 blocker.

## §G — Dependencies

- 선행: 없음(카드 t1084 배차 시점에 로컬 develop `7f86971fc` 기준 분기 완료).
- 재사용: SPEC-WORKTREE-REAPER-001이 배송한 병합-탐지(`git merge-base --is-ancestor` / gh 뷰 폴백) 및 lock anchor guard 메커니즘 — 판정 조건이 답하는 범위에서 재사용(관계 상세: plan.md § REAPER Prior-Art Relation).
- 운영 전제: 리드 일괄 push(레인은 push하지 않음) — T1 판정은 `origin/develop` 기준이므로 리드 push가 진행될수록 T1 판정 가능 트리가 늘어난다.

## §H — Cross-References

- 카드: t1084 (배차 2026-09-22)
- 선행 SPEC: SPEC-WORKTREE-REAPER-001 (v0.4.1, completed) — 병합-탐지 no-answer 처리, lock anchor guard, `moai worktree clean --stale`
- 구현 참조: `internal/cli/session_worktree_prmerge.go` (`parseWorktreeList` — porcelain 파싱 + lock 상태 캡처)
- 교훈 인용: t1059(원격 ref staleness), t1009/t1022(반출 인용 추적), t1069(측정 오염), t528(워크트리 무음 소실), t1055(결손 보고서 구출 전례)
- 독트린: `.claude/rules/moai/core/verification-claim-integrity.md` §3(5-섹션 판정 형식), AGENTS.md §3(worktree 폐기 — 통합·원격 착지 전 폐기 금지)
- T2 예외 선언: T2 처치는 AGENTS.md §3 / gitflow-lane-protocol §6(원격 착지 전 폐기 금지)에 대한 **운영자 판정 기록이 구속하는 명시적 예외**다 — 미병합 트리를 제거하되 분기 ref를 보존해 작업이 ref에 생존하며, 반출-선행(REQ-WGC-009)과 방향 분류(REQ-WGC-006)가 그 예외의 완화 장치다. T1은 §3의 직접 인코딩, T3는 §3의 보수적 축이다.

## Out of Scope

### Out of Scope — Go 제품 코드 변경

- 본 SPEC은 OPERATIONS 카드다. run 페이즈가 Go 소스(`internal/`, `pkg/`, `cmd/`)를 수정하지 않는다. 런 페이즈가 보조 스크립트를 두더라도 제품 코드로 취급하지 않는다.

### Out of Scope — CI·인프라 편집

- `.github/workflows/`, 빌드 매트릭스, CI 게이트를 편집하지 않는다.

### Out of Scope — primary 체크아웃·제외 트리 기록

- primary 체크아웃의 워킹 트리, 제외 집합 16개 트리, 타인 소유 트리에 어떤 파일도 기록하지 않는다.

### Out of Scope — git push

- 레인은 `develop`을 push하지 않는다(리드 일괄). 본 SPEC의 run 페이즈도 push를 수행하지 않는다.

### Out of Scope — 본 SPEC의 run 실행

- 본 SPEC은 계획이다. 제거 실행은 run 페이즈가 수행하며, plan 페이즈는 어떤 트리도 제거하지 않는다.
