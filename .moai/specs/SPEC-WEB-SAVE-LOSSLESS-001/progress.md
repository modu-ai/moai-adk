# SPEC-WEB-SAVE-LOSSLESS-001 Progress

> 카드 t1314 · GitHub issue #1731 · branch `WT-web-save-lossless`

## Phase Log

- 2026-09-29 plan-phase: manager-spec이 4종 산출(spec/plan/acceptance/progress) 초안 작성. 코드 근거 12곳 워크트리 직접 확인 (develop `2b1233b13`). 설계 결정 A+B+백스톱 확정 (plan.md §A). 다음: plan-audit.
- 2026-09-29 plan-audit iter-1: **FAIL 0.85** (차단 3건) → v0.1.1 정정 (커밋 `8feb4bcdc` 대상). F1 — D2 수리 공허성(UserConfig `name` 전용, `saveSection` 무병합 재마샬) → user.yaml `name:` 행 seam 스플라이스로 재설계. F2 — quality_extras 강제 폐기 결정 확정(§A.4 Q1). F3 — AC-WSL-002 술어 변이 분리. F4 — M1 seam 허용 확장(`sectionwrite.go:56-64`) 명시. Q2/Q3 판정 승인 기록. 정정 근거 신규 실측: `manager.go:462-470`, `schema_sections_test.go:285-290`, `sectionwrite.go:56-64`, `config.go:32-37`. 다음: delta re-audit (iter-2/2).
- 2026-09-29 plan-audit iter-2: **CONCERNS 0.90 (PASS-with-debt)** — F1-F4 해소 확인, 처분 요구 F5 + 선택 F6 → v0.1.2 반영. F5 — 잔여 재마샬 3경로를 M1 공동 범위로 편입(REQ-WSL-002/003 명시적 포괄 + AC-WSL-009 신설). **인용 경로 정정**: 감사자 인용 `internal/cli/projectconfig.go`·`internal/cli/nested.go`는 본 트리에 부재(ls 실측) — 행 번호가 일치하는 실제 경로 `internal/web/projectconfig.go:231-238/240-247`, `internal/settings/nested.go:112-156`(웹/TUI 공유 seam, TUI 호출점 `internal/cli/profile_setup.go:226`)로 검증 후 반영. F6 — AC-WSL-005 name-부재 변이에 C3 허용 1행. 다음: run 진행 (Implementation Kickoff Approval 경유).

## §E.1 Plan-phase Audit-Ready Signal

- artifacts: spec.md 0.1.0 / plan.md / acceptance.md / progress.md (본 파일)
- SPEC ID 사전-검증: `SPEC-WEB-SAVE-LOSSLESS-001` regex PASS (Bash 실행, verbatim 출력 인용 완료)
- ID 유일성: `.moai/specs/` 983개 중 SPEC-WEB-SAVE-LOSSLESS-001 부재 확인 (ls 실측)
- 관련 SPEC 존재 확인: SPEC-WEB-WRITE-SAFETY-001, SPEC-GITSTRATEGY-SAVE-ISOLATION-001, SPEC-SEAM-GREENFIELD-001/002, SPEC-WEB-CONSOLE-011 (ls 실측)
- Tier: M — 영향 파일 추정 8-12 (settings 4 + config 1 + profile 1 + web 2 + 테스트) — >15 파일 아님
- 향후 진입: plan-audit → Implementation Kickoff Approval → `/moai run SPEC-WEB-SAVE-LOSSLESS-001`

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
