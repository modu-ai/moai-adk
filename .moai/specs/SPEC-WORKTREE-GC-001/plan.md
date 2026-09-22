---
id: SPEC-WORKTREE-GC-001
title: "Legacy worktree comprehensive audit and safe disposal — implementation plan"
version: "0.1.0"
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
module: ".claude/worktrees"
tags: "worktree,cleanup,disposal,evidence-export,operations,gc"
---

# SPEC-WORKTREE-GC-001 — Implementation Plan

## §A — Context

- Work location: `.claude/worktrees/t1084` (branch `worktree-t1084` — 배차문이 `WT-legacy-cleanup`으로 기술했으나 2026-09-22 실측 커밋 직전 재판독값; gitflow-lane-protocol §1의 제자리 개명은 미적용 상태로 런 페이즈 착수 전 리드 확인 권장. fast-forwarded to local develop `7f86971fc`).
- Card: t1084 — 구배치 워크트리 약 530개 통합 점검 후 정리(운영자 지시 2026-09-22). OPERATIONS 카드 — 제품 코드 없음.
- Baseline (this session, 2026-09-22, measured): `git worktree list` = 544 rows (543 card trees + 1 primary); `du -sk .claude/worktrees` = 144,978,644 KB (~138.3 GB); local `develop` = `7f86971fc` ahead of `origin/develop` = `f5fff2190` (리드 일괄 push 대기).
- First-pass precedent (lead, 09-22): 9트리 제거(t1055, t1056, t1059, t1066, t1070, t1077, t1057, t1069, goal-dist) + t1055 트리에서 primary 결손 보고서 약 210 디렉터리 구출.
- `moai worktree` 동사 실측: new, sync, remove, clean, recover, done + guard 동사 snapshot, verify, restore. `moai worktree done`은 L2 전용 — L1 트리(`.claude/worktrees/`)는 `git worktree unlock` + `git worktree remove` 경로다.

### §A.1 — PRESERVE list

- 제외 집합 16트리: t1051, t1060, t1065, t1067, t1068, t1071, t1073, t1074, t1078, t1079, t1080, t1081, t1083, t1084(본 카드), `develop` 워크트리, primary 체크아웃.
- `origin/*` refs (읽기 전용 — fetch만 허용).
- 모든 측정 대상 트리의 내용(읽기 전용 — 어떤 출력도 안에 쓰지 않는다).

### §A.5 — 변경 대상

- `.moai/reports/t1084/` (본 카드 워크트리 안): 측정 기록, 분류 행렬, rescue/ 반출물, removal log, verdict.md — 전부 `WT-legacy-cleanup`에 커밋.
- 제거 대상 T1/T2 트리 자체(git 메타데이터 변화 — shared `.git`의 worktree 등록 해제).

## §B — Known Issues

- **B-GC-1 (원격 ref staleness)** — fetch 없이 `origin/develop`을 읽으면 자기 ref의 나이를 잰다(t1059). T1 판정의 push 조건은 반드시 `git fetch origin develop` 이후 측정한다. 로컬 develop이 origin보다 앞서 있으므로(7f86971fc vs f5fff2190), "로컬 develop에 병합됨"만으로 T1이 아니다 — origin 조상 판정만이 T1이다.
- **B-GC-2 (측정 오염)** — graph check 출력의 리다이렉트가 edges 층을 stale로 만든 실측(t1069). 측정 대상 트리 안에 출력을 쓰지 않는다. 모든 산출물은 본 카드 워크트리로.
- **B-GC-3 (반출 불완전)** — 판정서 반출이 그것이 인용하는 로그를 따라오지 않는 실측(t1009/t1022). 반출 단계에서 각 판정서가 인용하는 경로를 추적해 함께 반출한다.
- **B-GC-4 (무음 소실)** — 워크트리는 아무도 지우지 않아도 사라질 수 있다(t528). prune 전후 `git worktree list` diff로 무음 메타데이터 드리프트를 잡는다.
- **B-GC-5 (T2 판정 기록 부재 실측)** — 카드 예시 t154의 판정 기록은 이번 세션 탐색에서 미확인(memory grep 0행, 큐 드롭 목록 0행, `.moai/reports/t154` 양쪽 체크아웃 부재, primary 큐 파일은 09-13자로 현 큐 이전). 따라서 T2 인용 규칙은 "이름 붙은 기록으로 해소되는 것"만 인정하고, 미해소 주장은 blocker 상신으로 간다(REQ-WGC-013).
- **B-GC-6 (dirty/lock/점유)** — 미커밋 tracked 수정은 자동 폐기 금지(REAPER의 dirty guard와 동일 방향), 살아 있는 세션 점유는 제거 차단, lock은 unlock이 판정+반출 검증 뒤에만 온다.
- **B-GC-7 (L1/L2 혼동)** — `moai worktree done`은 L2 전용이다. `.claude/worktrees/` 아래 카드 트리는 L1이므로 `git worktree unlock` + `git worktree remove` 경로만 유효하다.

## §C — Pre-flight

```bash
# 1. 분기·HEAD 재확인
git branch --show-current && git rev-parse --short HEAD

# 2. 시작 측정 쌍 (REQ-WGC-001)
git worktree list | wc -l
du -sh .claude/worktrees

# 3. 원격 동기화 (REQ-WGC-004 — 어떤 origin/* 판정보다 먼저)
git fetch origin develop && git rev-parse origin/develop

# 4. 제외 집합 존재 확인
ls -d .claude/worktrees/{t1051,t1060,t1065,t1067,t1068,t1071,t1073,t1074,t1078,t1079,t1080,t1081,t1083,t1084,develop}

# 5. 세션 레지스트리 (점유 판정 입력)
moai session list --json
```

## §D — Constraints (안전 불변식 — 교훈 인용)

| # | 불변식 | 교훈/독트린 |
|---|--------|------------|
| C1 | 모든 origin/* 판정 전 `git fetch origin develop` 선행 | t1059 |
| C2 | 미커밋 tracked 수정 트리 ⇒ 자동 폐기 금지, T3 처리 또는 운영자 안건 | REAPER dirty guard (REQ-WR-021 계열) |
| C3 | 살아 있는 세션 점유 트리 ⇒ 단계 무관 제거 차단 | agent-common-protocol § Background Agent Execution |
| C4 | lock 해제는 단계 판정 + 반출 검증 둘 다 확인 후에만 | REAPER lock anchor guard |
| C5 | 반출 목적지는 본 카드 워크트리 rescue/ 디렉터리, `WT-legacy-cleanup`에 커밋 — primary·제외 트리 기록 금지 | 본 SPEC REQ-WGC-009 |
| C6 | 판정서 반출 시 인용 대상 동반 반출 | t1009/t1022 |
| C7 | 측정 중 트리에 출력 기록 금지 — 모든 산출물은 본 카드 워크트리로 | t1069 |
| C8 | `git worktree prune`은 제거 뒤에만, 전후 listing diff 기록 | t528 |
| C9 | 배치 안전 — du/반출은 소청크·순차, 백그라운드 부하 금지 | gitflow-lane-protocol §8 |
| C10 | AGENTS.md §3 — 분기가 통합·원격 착지하기 전 트리 폐기 금지. T1/T2/T3는 이것의 인코딩이다 | AGENTS.md §3 |
| C11 | 모든 판정 주장은 실행한 명령 + 관측한 출력으로 귀속 | verification-claim-integrity.md §1-§3 |
| C12 | 레인 push 금지 — develop push는 리드 일괄 | gitflow-lane-protocol §4 |

## §E — Self-Verification

런 페이즈 완료 시 다음을 §E 형식으로 보고한다(모든 항목: 명령 + verbatim 출력 + (this run, this tree) 귀속):

- E1: AC PASS/FAIL 행렬(acceptance.md §D) — 각 행 명령+출력.
- E2: 시작/종료 측정 쌍 — `git worktree list | wc -l`, `du -sh .claude/worktrees` 각 2회 관측값.
- E3: 제거 행렬 — REQ-WGC-014 형식(트리, 단계, 병합 근거, push 근거, 반출 근거, 제거 출력) 전 행.
- E4: 제외 무결성 — 제외 16트리의 `git worktree list` 존속 확인 + 집산식(시작 = 제거 + T3 유지 + 제외).
- E5: prune 전후 listing diff 기록.
- E6: Gaps — 명시적으로 관측하지 못한 것의 나열(빈 Gaps는 "아무것도 남지 않았다"는 주장이므로 그 자체가 참이어야 한다).
- E7: Blocker — REQ-WGC-013 미해소 T2 판정 목록.

## §F — Milestones

우선순위: High = 판정·안전 축(되돌리기 비쌈), Medium = 기계 축. 순서는 되돌릴 수 없는 결정(제거)을 뒤에 두는 가역성 축을 따른다.

- **M1 인벤토리·측정 (High)** — 시작 측정 쌍 기록. `git worktree list --porcelain` 전수 수집(트리 경로, 분기, lock 상태 — REAPER의 `parseWorktreeList`가 증명한 형태). 트리별 메타데이터는 shared `.git` refs에서 metadata-only 수집 — 단계 판정 조건에 트리당 git 서브프로세스 호출이 필요하지 않다. 제외 집합 필터 적용. 산출: `.moai/reports/t1084/inventory.md`.
- **M2 분류 (High)** — 트리별 T1/T2/T3 판정 + 근거 인용. fetch 선행(C1). T1 = `git merge-base --is-ancestor <branch> origin/develop` 종료 0(fetch 후). T2 = REQ-WGC-006 기록 해소. T3 = 나머지. dirty/lock/점유 예외(C2/C3)는 판정 결과를 무시하고 T3으로 눌러 내린다. 로컬 develop에만 병합된 트리는 "pending lead bulk push" 사유로 T3 유지. 산출: `.moai/reports/t1084/classification.md` (트리당 한 행: 트리, 단계, 판정 근거 명령+출력 또는 판정 기록 경로).
- **M3 증거 반출 (High)** — 제거 예정(T1/T2) 트리의 미추적 파일을 `.moai/reports/t1084/rescue/<tree-name>/`으로 반출. 증거류/폐기류 트리아지(REQ-WGC-010), 인용 대상 동반 반출(C6). 스킵 항목 건수 보고. `WT-legacy-cleanup`에 커밋. 산출: rescue/ 트리 + `.moai/reports/t1084/export-log.md`.
- **M4 처치 (High)** — T1/T2 제거(`git worktree unlock` 필요시 → `git worktree remove`), T1은 분기 ref 삭제, T2는 ref 보존. 미해소 T2 주장은 blocker 상신하고 트리 유지(REQ-WGC-013). 제거 행렬 기록(REQ-WGC-014). 제거 후 `git worktree prune` + 전후 listing diff(C8). 산출: `.moai/reports/t1084/removal-log.md`.
- **M5 검증·판정 (Medium)** — 종료 측정 쌍, 제외 16트리 존재 재확인, 집산식 검산(E4), `verdict.md` 5-섹션 형식(REQ-WGC-016). 산출: `.moai/reports/t1084/verdict.md`.

## §G — Anti-Patterns

- "병합됐을 것이다" 추론 — `git merge-base --is-ancestor` 종료 코드 관측 없이 제거 진행(관측 없는 주장).
- du 한 번에 138 GB 전체 재측정 반복 — 측정은 시작·종료 1회씩이면 충분(REQ-WGC-001).
- rescue 디렉터리에 트리 전체 블록복사 — 트리아지 규칙(REQ-WGC-010) 없이 무차별 복사하면 반출이 138 GB 규모로 부풀고 node_modules 등 폐기류가 디스크를 잡아먹는다.
- 제거를 먼저 하고 반출을 "확인차" 나중에 — 반출은 제거의 선행 조건이다(REQ-WGC-009 순서 하드).
- T2 판정 주장을 "리드가 말했다"는 구두 주장으로만 수용 — 이름 붙은 기록 경로 없이는 blocker다(REQ-WGC-013, B-GC-5 실측).

## Phase 1 SKIP Rationale

본 SPEC은 레인 세션에서 작성됐다. 카드 본문이 운영자 확정 의도(2026-09-22 운영자 지시, 칸반 리드 배차)를 그대로 담고 있어 Socratic 인터뷰 라운드를 생략한다. 잔여 unknown은 본 파일의 [NEEDS CLARIFICATION] 표식으로만 운반하고 spec.md·acceptance.md에는 넣지 않는다. 스킵이 임의 판단이 아닌 근거: 카드 텍스트가 3단 판정·제외 집합·전후 측정·반출 선행을 이미 [HARD] 절로 확정하고 있다.

## FO-PLAN-1 skip note

연구 fan-out(FO-PLAN-1)을 생략한다 — 연구 표면이 직접 탐색으로 소진됐다: REAPER spec.md 본문 정독(병합-탐지 3-상태 seam, lock guard, `clean --stale` 커버리지), `internal/cli/session_worktree_prmerge.go` 기호 나열(`parseWorktreeList` porcelain 파싱 + lock 캡처), `moai worktree` 동사 실측, 세션 기반 측정값(544/138.3GB). 발견은 본 계획의 §A/§B에 편입됐다.

## REAPER Prior-Art Relation

SPEC-WORKTREE-REAPER-001(v0.4.1, completed)과의 관계는 **재사용, 재구현 아님**:

- REAPER가 배송한 것: (1) 병합-탐지 3-상태 seam — merged / not merged / no answer, gh primary + `git branch --merged` 폴백(REQ-WR-001~004), (2) lock 기반 anchor guard(`internal/session`, 두 sweep 소비자 공유 — REQ-WR-019), (3) `moai worktree clean --stale`의 비-WT 커버리지 — 전 트리 열거, 트리당 keep-reason, 기본 preview, `--yes` 게이트(REQ-WR-012/022), (4) porcelain 파싱 + lock 상태 캡처(`parseWorktreeList`, `wtEntry.lock`).
- 본 SPEC run 페이즈가 재사용하는 것: 단계 판정 조건이 답하는 범위에서 REAPER의 병합-탐지 predicate(`git merge-base --is-ancestor` / gh 뷰 폴백)과 lock/점유 guard 논리를 그대로 적용한다. `moai worktree clean --stale`이 "merged + clean" 트리를 잡는 범위는 T1의 기계 하위집합으로 볼 수 있으나, 본 카드는 REAPER가 다루지 않는 것을 추가한다: 트리별 단계 증거 기록(병합 근거 + push 근거 명령 출력), 미추적 증거 반출 프로토콜(rescue/ + 트리아지 + 인용 추적), T2 운영자-판정 트리의 ref 보존 처치, 미해소 판정의 blocker 상신. `clean --stale`을 대신 실행하는 것은 본 카드의 [HARD] 반출-선행 조건을 우회하므로 대체재가 아니다 — 보완재다(REAPER가 판정하는 곳을 본 SPEC이 증거화한다).
- 본 SPEC이 REAPER를 변경하지 않는다: Go 코드 수정은 Out of Scope다.

## MX Plan

OPERATIONS 카드 — Go 코드 대상 없음. 런 페이즈가 보조 스크립트(분류 자동화 셸 조각 등)를 본 카드 워크트리에 두는 경우 해당 스크립트에 `@MX:NOTE`(용도·수명 명시)를 단다. 제품 코드가 아니므로 `@MX:ANCHOR` 요건(fan_in ≥ 3)은 적용되지 않는다. rescue/ 산출물은 문서이므로 MX 주석 대상이 아니다.

## [NEEDS CLARIFICATION] 항목

- [NEEDS CLARIFICATION: 트리당 반출 크기 상한] — 카드가 rescue 반출의 크기 상한을 정하지 않았다. 대형 빌드 산출물이 증거류로 오인되면 단일 트리 반출이 수 GB가 될 수 있다. 런 페이즈 진입 전 상한(예: 트리당 rescue 총량) 확정 필요.
- [NEEDS CLARIFICATION: lead-attested T2 판정의 표준 기록 경로] — REQ-WGC-006이 "배차 시 카드 진행 노트에 리드가 기록한 attested 기록 경로"를 인정하지만, 표준 경로/형식이 정해져 있지 않다(t154 실측 미확인이 여지의 근거). 리드가 어떤 경로에 무엇을 쓰면 유효 인용이 되는지 런 진입 전 확정 필요.
- [NEEDS CLARIFICATION: 제거 배치 상한] — 한 제거 창에서 몇 트리까지 진행할지(세션 점유 판정과 prune 사이의 무음 소실 창 최소화 관점) 미정. M4 실행 시 창 크기 확정 필요.

## §H — Cross-References

- spec.md §D (REQ-WGC-001~016) / acceptance.md §D (AC-WGC-001~012)
- SPEC-WORKTREE-REAPER-001 (v0.4.1) — 위 REAPER Prior-Art Relation 절
- `internal/cli/session_worktree_prmerge.go` — porcelain 파싱 참조 구현
- `.claude/rules/moai/core/verification-claim-integrity.md` §3 — verdict.md 형식
- 교훈: t1059, t1009, t1022, t1069, t528, t1055
