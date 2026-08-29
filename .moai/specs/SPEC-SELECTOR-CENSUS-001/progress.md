---
id: SPEC-SELECTOR-CENSUS-001
title: "0-실행 테스트 판정 — 진행 기록"
version: "0.1.0"
status: draft
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/hook, .claude/rules/moai/development, internal/template/templates/.claude/rules/moai/development"
lifecycle: spec-anchored
tags: "t341, progress"
tier: M
---

# 진행 기록 — SPEC-SELECTOR-CENSUS-001

카드 **t341** · 브랜치 `WT-selector-census` · 기준 트리 `a6bbbf82b`

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md` · `plan.md` · `acceptance.md` (Tier M 3종) + 이 파일
- 요구 7건(REQ-SEC-001..007) / 수락 8건(AC-SEC-000..007), Tier M 상한 16 이내
- 모든 RED-now 칸이 트리 `a6bbbf82b` 에 못 박혀 있고, 사유를 함께 적었다
- 미검증 전제 3건은 `spec.md` §5 에 명시했고, 그중 살아 있는 payload 관측은 `plan.md` M0 + **AC-SEC-000** 으로 승격했다
- 상태: `draft` — plan-audit 및 Implementation Kickoff Approval 대기

### iter-1 감사 대응 (`.moai/reports/t341/plan-audit-iter1.md`, PASS-WITH-DEBT 0.81)

blocking 5건을 닫았다. 전부 기준 문면 수정이며 `internal/` 코드는 건드리지 않았다.

| 결함 | 닫은 방법 | 이 라운드에서 실행한 명령 |
|---|---|---|
| D1 결속 검사 공허 | DoD 를 두 검사(삼중점 `origin/develop...HEAD` + `git status --porcelain`)로 교체, 실측 표 첨부. `plan.md` E5 → E5a/E5b | 결속 파일 1바이트 편집 + `git add` 후 두 형태 비교 → 종전 형태 `old-grep-exit=1`(초록), `git status --porcelain` → `M  …`(붉음). 이후 `git restore --staged` + `git checkout --` 로 원상 복구 확인 |
| D2 비발화 방향 1축 | AC-SEC-003 을 러너 축 전부(go/cargo/pytest/jest·vitest, 표본 5 × payload 2 = 10건)로 확장, 뮤턴트 탐침 2건 명시 | `pytest -q` → `3 passed in 0.00s` (정밀 마커 없음). `printf 'Tests:       10 passed, 10 total\n' \| grep -c '0 passed'` → `1` (부분 문자열 충돌) |
| D3 미러 판단 탈출구 | AC-SEC-007 (2) 를 `diff` rc 0 단언으로 교체, 탈출구 삭제 | `diff` 로컬 ↔ 미러 → 출력 0줄, `diff-exit=0` |
| D4 corpus 미결속 | AC-SEC-006 에 조건 (2) 신설 — 같은 변수 공유 + 각 표본이 `detectZeroExecution` 을 실제로 발화 | `evidence_writer.go:79` 판독(`return true, false, false` — 신호 없음도 `isPass=false`), 그래서 `isPass=false` 단언만으로는 불충분 |
| D5 M0 산문 게이트 | **AC-SEC-000** 신설 + DoD 줄 + `plan.md` E0 행 + M1 진입 조건 | `ls .moai/reports/t341/live-payload.json` → `No such file or directory`, `ls-exit=1` |
| D6 인용 좌표 | `spec.md` `:223`→분기/반환 분리, `:296-330`→`:309-330`(대입 `:328`) | `grep -n 'func buildBashRecord\|rec.IsTestPass'` → `309` / `328` |

**닫지 않은 것**: D7(요구 문면의 구현 표면 이름, optional — §3.3 이 이미 근거를 담고 있어 그대로 둔다), D8(AC-SEC-003 에 RED-now 칸 없음 — 여전히 신고된 부채다. 확장 후에도 열 표본 전부 오늘 초록이며, 관측하지 않은 RED 칸을 지어내지 않는다).

### iter-2 감사 이후 추가 편집 (D9, `.moai/reports/t341/plan-audit-iter2.md`, PASS-WITH-DEBT 0.894)

**이 편집은 마지막 감사 회차(iter-2, Tier M 상한) 이후에 착지했으며 재감사를 받지 않았다.** 성격은 **순수 추가(additive)** 다 — 기존 기준·DoD·RED-now 칸을 하나도 약화·개작·삭제하지 않았다.

- `acceptance.md` AC-SEC-003 에 **표본 (f)** 추가 — 네 pass 마커를 하나도 담지 않는 진짜 pass(node 내장 러너, `npm test` 로 도달) + `{"exit_code": 0}` payload 1건이 `isPass=true` 로 남을 것. 이것이 **exit-code 축(`deriveFromExitCode`, `:69`→`:163`)의 비발화 방향**을 처음으로 고정한다. payload 총계 10 → **11**((f) 는 짝을 만들지 않는다 — 사유는 기준 본문).
- 같은 절에 **뮤턴트 탐침 3** 추가 — 거부권을 "텍스트에 실행 수 근거 없음" 으로 구현해 exit-code pass 경로를 좁히는 형태. (f) 하나만 이 뮤턴트를 죽인다.
- `acceptance.md:86` 의 래퍼 문장 교정 — `npm`·`pnpm`·`yarn` 이 (c)(d)(e) 에 흡수된다는 **일반 주장은 거짓**이다. 아래 러너가 무엇인지는 래퍼가 정하지 않는다.
- `plan.md` §F 에 대응 안티패턴 1줄 추가 — M1 이 그 형태로 흘러가는 것을 계획층에서도 막는다.
- (f) 의 마커 부재는 이 트리(`a6bbbf82b`, 2026-08-29)에서 **실측**했다(`node --test` → `node-exit=0`, 마커 5종 `grep -c` 전부 `0`). 축자 문자열의 러너 판번 고정은 (b)(d) 와 같은 규율으로 M1 몫이다.
- iter-2 의 나머지 결함(D10·D11·D7·D8)은 **optional** 이라 손대지 않았다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
