# acceptance.md — SPEC-DEVPROT-REQUIRED-001

> 요구사항 계층(GEARS REQ)은 spec.md §2가 소유한다. 이 파일은 검증 계층 — Given-When-Then AC를 이원 셀(RED-now + green path)으로 열거한다.
> **모든 AC는 run-phase에서 기계 검증 가능해야 한다.** 보호 적용 자체(운영자 gh api PUT)만으로 검증 가능한 것은 AC가 아니라 런북 §절이다(REQ-DPR-010).
> RED-now 측정: 2026-09-02, worktree `t324`, tree `fa8ff89ba`, branch `WT-devprot-required` (모든 셀 동일 트리에서 이번 실행으로 측정).

## §D AC Matrix

| AC | 심각도 | 추적 REQ | 이원 셀 |
|---|---|---|---|
| AC-DPR-001 | release-blocking | REQ-DPR-007 | RED-now + green(M2) |
| AC-DPR-002 | release-blocking | REQ-DPR-004, 007 | RED-now + green(M1/M2) |
| AC-DPR-003 | regression-guard | (워크플로 편집 안전성) | baseline-green 유지 |
| AC-DPR-004 | release-blocking | REQ-DPR-010, 011 | RED-now + green(M1) |
| AC-DPR-005 | release-blocking | REQ-DPR-001 | RED-now + green(M1) |
| AC-DPR-006 | release-blocking | REQ-DPR-009 | RED-now + green(M1) |
| AC-DPR-007 | release-blocking | REQ-DPR-011 | RED-now + green(M1) |
| AC-DPR-008 | release-blocking | REQ-DPR-005, 008 | RED-now + green(M3) |
| AC-DPR-009 | release-blocking | REQ-DPR-007 | RED-now + green(M5) |
| AC-DPR-010 | release-blocking | REQ-DPR-013 | RED-now + green(M4) |
| AC-DPR-011 | release-blocking | REQ-DPR-002 | RED-now + green(M1) |
| AC-DPR-012 | release-blocking | REQ-DPR-003 | RED-now + green(M1) |
| AC-DPR-013 | release-blocking | REQ-DPR-006 | RED-now + green(M1) |
| AC-DPR-014 | release-blocking | REQ-DPR-012 | RED-now + green(M1) |

AC 14개 (Tier M 상한 16 이내).

## §D.1 시나리오 (Given-When-Then)

**AC-DPR-001 — ci.yml verify 트리거**
- Given 이 트리의 ci.yml이고, When `yq -o=json '.on.push.branches' .github/workflows/ci.yml`을 실행하면, Then 출력 배열에 런북이 선언한 verify 패턴 토큰이 포함된다(기존 `main`·`develop`과 함께).
- RED-now: `yq -o=json '.on.push.branches' .github/workflows/ci.yml` → `["main","develop"]` (verify 토큰 없음, exit 0) — red인 이유: 동반 변경이 아직 착지하지 않았고 이것이 M2의 산출물이다.
- Green path: M2에서 트리거 확정 후 동일 명령 출력에 verify 토큰 포함.

**AC-DPR-002 — codeql.yml verify 트리거 (DECIDED D-1: `Analyze (Go) (go)` phase-1 포함)**
- Given 런북이 `Analyze (Go) (go)`의 phase-1 포함을 DECIDED(plan.md §A.1 D-1)로 기록했고, When `yq -o=json '.on.push.branches' .github/workflows/codeql.yml`을 실행하면, Then 출력 배열에 `verify/*` 항목이 포함된다.
- RED-now: `yq -o=json '.on.push.branches' .github/workflows/codeql.yml` → `["main","develop"]` (`verify/*` 부재).
- Green path: M2 확장 후 포함. (조건 분기 소멸 — 결정 확정으로 무조건 AC.)

**AC-DPR-003 — 워크플로 회귀 가드 (regression-guard)**
- Given 편집된 워크플로 파일들이고, When `actionlint .github/workflows/ci.yml .github/workflows/codeql.yml`을 실행하면, Then exit 0을 유지한다.
- Baseline(이번 실행): 현재 파일에서 exit 0 측정됨 — 이미 초록이므로 release-blocking이 아니라 회귀 가드로 채택(편집이 깨뜨리지 않았음을 증명).

**AC-DPR-004 — 런북 존재 + apply/rollback 명령**
- Given run-phase 완료 상태이고, When `.moai/docs/develop-protection-runbook.md`을 확인하면, Then 파일이 존재하고 `gh api -X PUT /repos/modu-ai/moai-adk/branches/develop/protection` 문자열과 롤백 명령(보호 제거/원복 `gh api` 명령) 각 ≥1회 포함한다.
- RED-now: `ls .moai/docs/develop-protection-runbook.md` → `No such file or directory` (exit 1).
- Green path: M1에서 런북 작성 후 grep ≥1.

**AC-DPR-005 — 런북 apply 블록의 필수 세트 (DECIDED D-1: 4컨텍스트)**
- Given 런북이 존재하고, When apply 절의 `required_status_checks.contexts` 블록을 확인하면, Then `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, `Analyze (Go) (go)` 네 컨텍스트가 정확한 이름으로 나열된다.
- RED-now: 런북 부재(AC-DPR-004와 동일 측정)로 0회.
- Green path: M1 후 4회 grep 각 ≥1.

**AC-DPR-006 — GH006 예상 거부 문서화**
- Given 런북이 존재하고, When `grep -c "GH006" .moai/docs/develop-protection-runbook.md`을 실행하면, Then ≥1 (예상 거부 형태와 회복 경로 포함).
- RED-now: `grep -rn "GH006" .moai/docs/` → 출력 없음(0 hit).
- Green path: M1 후 ≥1.

**AC-DPR-007 — 롤아웃 순서**
- Given 런북이 존재하고, When 세 순서 마커(① 워크플로 변경, ② 창 절차 갱신, ③ 보호 적용)의 행 번호를 `grep -n`으로 읽으면, Then 번호가 오름차순이다(순서 보존).
- RED-now: 런북 부재로 마커 0개.
- Green path: M1 후 마커 3개 + 순서 단조 증가.

**AC-DPR-008 — 창 절차 B1 사전검증**
- Given 창 절차 문서(`.claude/rules/local/gitflow-lane-protocol.md`)이고, When `grep -n "verify/" <파일>` (또는 B1 절차 절 제목 토큰)을 실행하면, Then B1 절차(verify ref push → 초록 대기 → 동일 SHA develop push)와 락 홀드 연장 안내가 포함된 절이 ≥1 존재한다.
- RED-now: `grep -n "verify\|사전검증" .claude/rules/local/gitflow-lane-protocol.md` → 유일 hit가 :56의 `git rev-parse --verify MERGE_HEAD`(무관) — B1 관련 0개.
- Green path: M3 후 verify 패턴 토큰 hit ≥1 + 절차 단계 포함.

**AC-DPR-009 — verify 패턴 정준형 일관성 (DECIDED D-2)**
- Given run-phase 완료 상태이고, When (1) `yq`로 ci.yml·codeql.yml의 push branches를 읽고 (2) 런북·창 절차 문서에서 verify 표기를 grep하면, Then (1) 양 워크플로가 **동일한 `verify/*` 항목**을 포함하고 (2) 문서 표기는 카드별 형식 `verify/<card-id>`로 통일된다(파일마다 다른 패턴 없음).
- RED-now: `grep -n "verify" .github/workflows/ci.yml` → hit 없음 (exit 1); codeql.yml 동일.
- Green path: M5 스윕에서 양 워크플로 `verify/*` 동일 항목 + 문서 카드별 표기 확인.

**AC-DPR-010 — doctrine 낡은 스냅샷 고지**
- Given `.moai/docs/git-local-workflow-doctrine.md`이고, When §23.2 근처에 스냅샷 낡음 고지/live API SSOT 문구를 grep하면, Then ≥1 (spec.md §1.2의 어긋남 서술과 대응).
- RED-now: `grep -cn "stale\|낡은 스냅샷\|live API" .moai/docs/git-local-workflow-doctrine.md` → `0` (exit 1).
- Green path: M4 후 ≥1.

**AC-DPR-011 — 배제 근거 문서화**
- Given 런북이 존재하고, When `Release PR Multi-OS Gate`와 test-install 배제를 grep하면, Then 각각 근거(구조적 부재→영구 Pending / paths 필터→Pending)와 함께 ≥1회 등장한다.
- RED-now: 런북 부재로 0회.
- Green path: M1 후 각 ≥1.

**AC-DPR-012 — phase-2 후보 분류 문서화**
- Given 런북이 존재하고, When `grep -n "Race Test" .moai/docs/develop-protection-runbook.md`을 실행하고 동반 변경 사유(skip-marker 부재 → docs-only push에서 미보고) 표기를 확인하면, Then 사유와 함께 ≥1 hit이 존재한다 (REQ-DPR-003의 phase-2 분류 기록).
- RED-now: 런북 부재(AC-DPR-004와 동일 측정)로 0.
- Green path: M1 후 ≥1 + 사유 표기.

**AC-DPR-013 — B2 과도기 정직 프레이밍**
- Given 런북이 존재하고, When B2 단계 서술에서 (1) admin 자격 push 우회 표기(`admin`)와 (2) 사후 검증 한정 표기(예: `push 후`/`사후`)를 grep하면, Then 각 ≥1 — "필수 검사가 창을 사전 게이트한다"는 주장 없이 (REQ-DPR-006).
- RED-now: 런북 부재로 각 0.
- Green path: M1 후 각 ≥1.

**AC-DPR-014 — 7일 사전조건 보류 지시**
- Given 런북이 존재하고, When `grep -n "7일" .moai/docs/develop-protection-runbook.md`을 실행하고 보류 지시·읽기 전용 확인 절차(check-runs 조회) 표기를 확인하면, Then ≥1 (REQ-DPR-012).
- RED-now: 런북 부재로 0.
- Green path: M1 후 ≥1.

## §D.2 간접 검증 (운영자 게이트 영역 — AC가 아님)

- 런북의 모든 GET 명령은 dry-run 실행해 exit 0이어야 한다(예: `gh api .../branches/develop/protection` → 적용 전 404 "Branch not protected"가 **문서화된 기대값**).
- PUT 명령은 실행하지 않는다. 형식·인자(경로, `--input` JSON 구조) 검증으로 대체.
- 실제 보호 적용·GH006 거부 재현·B1 end-to-end는 운영자가 런북을 실행하는 시점의 관측으로 남는다 — 본 SPEC의 완료 판정에 포함하지 않는다.

## §D.3 종결 게이트 (Definition of Done)

- release-blocking AC 13개 전부 green (§D.1의 green path 실측).
- regression-guard AC-DPR-003 baseline 유지.
- `gh api -X PUT/PATCH/DELETE` 실행 이력 0회 (run-phase 전체).
- `.github/workflows` 외 워크플로·잡 구조 변경 0행 (트리거 배열 추가만).
