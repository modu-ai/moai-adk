# progress — SPEC-DEPLOY-RESULT-WIRE-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (spec.md + plan.md + acceptance.md). 근거: 예상 변경 파일 6-8개 · 예상 LOC 150-350 — Tier S 상한(5 files / 300 LOC)의 경계에 걸치고, 세 호출부가 두 패키지(`internal/cli`, `internal/core/project`)에 흩어져 판정 팔이 3개로 갈라진다.
- 요구사항 **9개** / 판정 기준 **8개** — Tier M 상한 16/16 이내.
- SPEC ID 정규식 검사: `SPEC-DEPLOY-RESULT-WIRE-001` → **`PASS`** (Bash `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` 실행, verbatim 출력 `PASS`).
- 중복 검사: `ls -d .moai/specs/SPEC-DEPLOY*` → `no matches found`; `grep -rl "SPEC-DEPLOY-RESULT-WIRE-001" .moai/specs/` → 무출력. 신규 ID 확정.
- 착수 시점 실측(작성 시점, 워크트리 `t176` / 브랜치 `WT-deploy-result-wire`):
  - `deployer.Deploy(` 프로덕션 호출부 **3곳** — `internal/cli/update_template_sync.go:323`, `internal/cli/update_clean_install.go:439`, `internal/core/project/initializer.go:356`.
  - `ResultDeployer` 프로덕션 소비자 **0건**(sync-audit F2 재확인).
  - `InitResult.Warnings []string` 존재 — `internal/core/project/initializer.go:97`, 기존 append 6곳, 표시부 `internal/cli/init.go:706`.
  - 출력 스트림 독트린 — `internal/cli/CLAUDE.md:14` "stderr = human progress messages, warnings, errors. Never mix."
- 카드 전제 정정 2건: (1) 범위가 "internal/cli (update 경로)" 로 적혀 세 번째 호출부를 덮지 않았다 — 세 곳 전부를 범위에 넣었다. (2) init 배선을 "materially larger change" 로 적었으나 **되돌릴 통로(`InitResult.Warnings`)가 이미 있고 `deployTemplates` 가 `result` 를 이미 인자로 받는다** — 비용이 update 호출부와 같은 급이다.
- 설계 결정 기록: `skipped` 모드 경고(sync-audit F4 오귀속)는 사용자에게 **올리지 않는다**(REQ-DRW-009 / AC-DRW-008). 문구 수정은 승계 카드 `t173` 소관이며, 이 SPEC 은 문구를 고치지 않는다.
- 잔존 고지: 폴백 플랫폼 2회차 이후 실행은 모드가 `skipped` 라 통지가 나가지 않는다(sync-audit F1). spec §D 본문에 명시.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
