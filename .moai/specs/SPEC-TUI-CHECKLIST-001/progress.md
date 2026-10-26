# SPEC-TUI-CHECKLIST-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-26T21:01:12Z

Plan 단계 산출물 작성 완료 (manager-spec, 카드 t1969, Tier L): spec.md v0.1.0 (REQ 20개), plan.md, acceptance.md (AC 23개, 증거 원장 L1~L14), design.md, research.md. 워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`, plan 측정 HEAD `cb30e05b3e`.

델타 1회차 (manager-spec, 측정 HEAD `d241917d17a3754541ef5ad52536d68023faf142`): 1회차 plan 감사(FAIL 0.71, `.moai/reports/t1969/plan-audit-verdict.md`)의 결함 D1~D20 을 고쳤다. spec.md v0.1.1 (REQ 23개), acceptance.md (AC 24개, 증거 원장 L1~L19 — L15~L19 는 이 HEAD 에서 측정). 레인 결정 Q1~Q3 은 plan.md §G 에 결정과 "리더가 다른 안을 택하면" 단락으로 적었고, Kickoff 기록에서 리더에게 알린다.

델타 2회차 (manager-spec, 측정 HEAD `b422a13d5377639d58fc672c986b81c678fe3318`): 2회차 plan 감사(FAIL 0.75, `.moai/reports/t1969/plan-audit-verdict-iter2.md`)의 결함 N1~N13 과 리더 codex_audit(영수증 `rcpt-bd48fb9e7212e0659ea0f6a3`)의 P2(= N8)를 고쳤다. spec.md v0.1.2 (REQ 23개 그대로, REQ-TCL-022·REQ-TCL-023 문장 수정, 결정 D-10 추가), acceptance.md (AC 24개 그대로, 증거 원장 L1~L25 — L20~L25 는 이 HEAD 에서 측정), design.md §3.2·§3.3.1·§3.4·§3.6·§4·§8, plan.md §A.0·§B·§C·§E·§F·§H·§I, research.md §7·§8 정정과 §9·§10 추가. 레인 결정 Q1·Q2·Q3, D-9, `main` 흡수 시점 규칙, 글리프 `◆`/`○` 는 다시 열지 않았다. 마지막 감사 회차(Tier L 상한 3) 대기.

델타 3회차 (manager-spec, 측정 HEAD `69358ecba8a9b8f996b3e9b0e4b39ae36728a80f`, 천장 정책의 자동 델타 1회): 3회차 plan 감사(FAIL 0.80, `scope: delta`, `.moai/reports/t1969/plan-audit-verdict-iter3.md`)의 `fix_scope` 앵커 안에서 BLOCKING M1~M3 과 SHOULD-FIX M4 를 고쳤다 — spec.md REQ-TCL-015·REQ-TCL-016(v0.1.3), acceptance.md AC-TCL-016·AC-TCL-017, design.md §3.3.1·§3.4·§4. REQ 23·AC 24 와 id 집합은 그대로다. 레인 결정 Q1~Q3, D-9, D-10, 글리프는 다시 열지 않았다. 새 측정 칸(트리 `69358ecba8`)은 AC-TCL-016·AC-TCL-017 본문 안에 적었다(원장 행은 더하지 않음).

범위 축소 v0.2.0 (manager-spec, 측정 HEAD `6f05252d6ddf62362a60cadbfc3350e0d068dbc4`, 천장 뒤 편집 — 리더 결정 기록 §15.98 SPLIT): 델타 4회차 감사(FAIL 0.80, `.moai/reports/t1969/plan-audit-verdict-iter4.md`, 차단 N1)를 받아 이 SPEC 을 렌더러 + `moai init` 연결로 좁혔다. update 연결은 카드 t1978 로 넘겼다(plan.md §K, spec.md §3.4). REQ 23·AC 24 와 id 집합은 그대로다 — 지우거나 다시 매긴 id 는 없고, 각 id 의 update 조각만 걷어내 `[DEFERRED → t1978]` 로 표시했다. Tier 는 측정 근거로 L 을 유지했다(plan.md §A.1). 새 원장 L26~L33 을 이 HEAD 에서 쟀다. 아래 미해결 부채의 처분:
- M5 → 수리(궤도에 `printer/printer.go`·`severity_line.go`, X1·X7·X9 정정, research.md §9 한계 문단).
- M6 → 수리(AC-TCL-013 O4b (i) 범위 `-- internal pkg go.mod go.sum`).
- M7 → 수리(AC-TCL-012 `init_writer_replaced`).
- M8 → 수리(검색 모양 `PillWarn`·`PillErr`, 원장 L26, 일치 9줄을 X7·X9·X11 로 분류).
- M9 → 수리(`inventory_complete` 궤도 항목 열여섯 개를 하위 테스트로 열거).
- R-M2a(4회차 N5) → research.md:127 문단에 정정·이월 표시.
- R-M2b(4회차 N3) → P22 가 update 행으로 이월되어 AC-TCL-024 (2) 의 괄호를 지움.
- R-M1(4회차 N6) → spec.md §6 에서 M5 를 지움(AC-TCL-016 은 init 만 — M4).
- R-M4, R-M4b → t1978 로 이월(plan.md §K K-12).
- 4회차 N1·N2·N4·N7 → t1978 로 이월(plan.md §K K-1~K-4).

미해결 부채 (v0.1.3 시점 기록 — 처분은 바로 위 v0.2.0 문단):
- M5 (SHOULD-FIX) — 궤도 추적이 메서드 수신자 작성기를 보지 못해 `internal/cli/printer/printer.go` 가 X9 로 분류됨(design.md §3.3.1 궤도 정의·X9 행, research.md §9 궤도 추적). 3회차 감사의 근거 위치는 design.md:131·:174 로, 앵커 §3.3.1 안이지만 감사가 fix_scope 결함으로 지정하지 않았으므로 손대지 않았다.
- M6 (SHOULD-FIX) — AC-TCL-013 O4b (i) 범위가 `-- internal` 뿐이라 `pkg/`·`go.mod`·`go.sum` 변경을 보지 못함(acceptance.md AC-TCL-013).
- M7 (MINOR) — AC-TCL-012 명령 수준에 조건 2(작성기 교체) 하위 테스트 없음.
- M8 (MINOR) — 검색 모양에 `tui.Pill(… PillWarn|PillErr …)` 없음(`update_template_sync.go:336` → `update_tux.go:120`).
- M9 (MINOR) — `inventory_complete` (다) 가 테스트 안의 궤도 목록만 기준으로 판정(acceptance.md AC-TCL-024).
- 잔여 R-M2a — research.md:127 이 "이미 최신" 경로를 여전히 본문 `update_template_sync.go:249-255` 로 적는다. 실제 명령 경로는 `:1260-1266`(design.md §3.4, AC-TCL-017 (d)). research.md 는 fix_scope 밖.
- 잔여 R-M2b — acceptance.md AC-TCL-024 (2) 의 "P22 는 REQ-TCL-023 의 `✓`" 괄호가 P22 를 항목 안 행으로 적는다. P22 는 이제 항목 밖 행이다(design.md §3.3.1). AC-TCL-017 (d) 가 우선한다고 적어 두었지만, `p22_update_up_to_date` 하위 테스트를 쓰는 run 단계가 두 문구를 함께 읽게 된다.
- 잔여 R-M1 — spec.md §6 추적표 사본은 AC-TCL-016 의 마일스톤을 M4 로만 적는다. `update` 하위 테스트는 M5(update 연결) 뒤에 녹색이 된다.
- 잔여 R-M4 — 깨끗한 update 의 정의(spec.md §3.3)가 `--yes` 를 적지 않는다. `--yes` 없는 실행의 확인 화면 줄은 줄 예산에 넣지 않았다(design.md §4). AC-TCL-011 의 고정 실행이 `--yes` 로 돈다는 전제가 문서 어디에도 고정되어 있지 않다.
- 잔여 R-M4b (문구만) — design.md §3.5 "update 의 확인은 `analyze` 항목을 닫은 뒤 열린다" 는 본문이 직접 확인을 띄우는 경우(`skipConfirm` 거짓, `update_template_sync.go:355`)를 적은 것이다. 실제 명령에서는 확인이 바깥 함수(`:1287`)에서 항목이 열리기 전에 돈다(design.md §4). 두 경우 모두 "질문 중 열린 줄 없음"은 성립하므로 결함이 아니라 문구 보강 대상이다. §3.5 는 fix_scope 밖.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
