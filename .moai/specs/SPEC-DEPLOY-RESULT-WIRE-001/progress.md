# progress — SPEC-DEPLOY-RESULT-WIRE-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (spec.md + plan.md + acceptance.md). 근거: 예상 변경 파일 6-8개 · 예상 LOC 150-350 — Tier S 상한(5 files / 300 LOC)의 경계에 걸치고, 세 호출부가 두 패키지(`internal/cli`, `internal/core/project`)에 흩어져 판정 팔이 3개로 갈라진다.
- 요구사항 **10개** / 판정 기준 **9개** — Tier M 상한 16/16 이내 (iter-1 은 9/8).
- iter-2 (v0.2.0): plan-audit iteration 1 **FAIL(0.75)** 의 결함 7건 대응. **D1(critical)** — §A.4 스트림 실측이 틀렸다: clean-reinstall 의 `nil ⇒ os.Stderr` 는 프로덕션에서 실행되지 않는 분기이고, 호출부 두 곳이 stdout 을 주입한다. 표·결론 정정 + M3 을 `CleanReinstallOptions.ErrOut` 주입으로 재배선 + plan R4("기본 경로에서 판정") 은퇴 + AC-DRW-003 을 **주입된 프로덕션 배선**에 재고정. **D2** — `REQ-DRW-007` 비비례성을 통지 전체로 확장, `failed` 를 개수 요약 + 예시 3건 상한으로 고정, AC-DRW-004 에 `failed` 팔 추가. **D3** — AC-DRW-003 을 stdout 위험 두 호출부에 명시 고정. **D4** — 오귀속 문구 행 `196` → `205`(재측정). **D5** — `REQ-DRW-010` / `AC-DRW-009` 신설(주입 seam 기본값 가드). **D6** — §A.7 신설(`failed` 문구 3-of-3 실측)로 1-of-3 일반화 대체. **D7** — 전 AC 에 위증 검사 절 추가.
- iter-2 실측 기록: (a) `grep -n "Out:" internal/cli/update.go` → `425`, `627`; 두 자리의 `out` 은 각각 `runUpdate`(`:138`)의 `:154 out := cmd.OutOrStdout()` 와 `emitDryRunReinstallPlan`(`:592`)의 `out` 인자이며 후자는 `:362` 에서 같은 `out` 을 받는다 → **둘 다 stdout**. (b) `grep -n "non-symlink entry already exists" internal/template/skill_mirror.go` → **205**. (c) `MirrorModeFailed` 생산 지점 3곳 → `:217`, `:226`, `:243`.
- iter-2 자기 지적: iter-1 은 **기본값 선언을 읽고 스트림을 판정**했다. 기본값의 존재는 그 기본값이 쓰인다는 근거가 아니다. D1·D5·D7 이 같은 형태이며, 「판정은 프로덕션 배선에서, 각 AC 는 위증 검사를 본문에 적는다」로 일괄 대응했다.
- iter-2 감사 쟁점: 반박한 지적 없음 — D1·D4 는 독립 재측정으로 재현했고 나머지 5건도 전부 수용했다. init 범위 결정은 감사가 §A.3 인용 4건 + 추가 연결 2건(`PhaseExecutor` 통과 · `p.Collect` → stderr)을 확인해 유지됐다.
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
