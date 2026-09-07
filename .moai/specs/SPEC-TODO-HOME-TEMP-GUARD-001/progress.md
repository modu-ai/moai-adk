# Progress — SPEC-TODO-HOME-TEMP-GUARD-001

카드 t536 · 트리 `.claude/worktrees/t536` · 브랜치 `WT-home-fallback` · plan-phase base `412c8cb14`.

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier S — 카드 지시에 따라 `acceptance.md` 별도 작성; `design.md`/`research.md` 없음).
- SPEC ID 정규식 자가 점검: `SPEC-TODO-HOME-TEMP-GUARD-001` → `PASS` (Bash 실행, 이 트리).
- 운영자 결정 반영: 임시-디렉터 기원 **한정** 거부 (「모든 비git base」 판독은 기각 — spec.md §3, plan.md D1).
- 미결 사항: **없다.** U1(거부 시 CLI 형태)은 운영자 선택으로 **(b) 안내 후 계속**으로 결정됐다(2026-09-07, 리드 경유 — plan.md §D). run-phase는 고르지 않고 구현한다.
- 2차 감사 수리(`version: 0.1.2`): D9(대체 루트를 `base`로 정정 — 종전 값은 루트가 아니라 상태 디렉터였다), D10(AC-THG-001이 문자열 동일성이 아니라 `BacklogPathForRoot(반환 루트)`의 읽힘을 단언), D11(`/var/tmp` 근거의 미측정 단정 격하 + 재검토 트리거 교체), U1 결정 반영. spec.md §8에 기존 코드(`todo_root.go:121`)의 계층 불일치를 **관측**으로 등재 — 본 SPEC은 고치지 않는다(별도 카드 후보).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
