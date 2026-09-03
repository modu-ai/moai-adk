# progress.md — SPEC-CODEX-SKILL-PATH-001

Card t468 · Tier S · worktree `.claude/worktrees/t468` (branch `WT-codex-path-parser`, base `d592b0551`).

## §E.1 Plan-phase Audit-Ready Signal

2026-09-03 plan-audit iter-2: D1-D4 해소 확인 + D6(AC-CSP-001-09를 -04(`~user`)/-08(bare `~`)로 접어 8-AC 상한 복원, fixture 행·REQ 열거는 유지)·D7(문서 카운트 정합) 적용.

2026-09-03 plan-audit iter-1 (PASS-WITH-DEBT 0.875) 터치업 적용: D1(REQ-CSP-004 `filepath.IsAbs` 우선 한정자 + AC-04 이중-OS 예외 형태로 교체 + plan §E windows 테스트실행 판정/fixture 절대경로 규칙), D2(AC-06 실제 missing + symlink-loop 쌍으로 재작성 — no-finding 자세 보존), D3(테스트 수 census 항목), D4(AC-CSP-001-09 `~`/`~user` + M1 fixture 행).

plan_status: audit-ready
plan_complete_at: 2026-09-03
plan_tree_sha: d592b0551
artifacts: spec.md (7 REQ / 8 AC, GEARS + Given-When-Then inline), plan.md (M1-M3 single-pass), progress.md (this file)
status: draft (initial `(none) → draft`, manager-spec)

Plan-phase evidence captured on tree `d592b0551`:

- RED-now probes (throwaway /tmp program, verbatim):
  - `probe1 stat-tilde err=stat ~/definitely-not-a-real-path-t468: no such file or directory isNotExist=true`
  - `probe2 stat-backslash err=stat C:\\Users\\goos\\SKILL.md: no such file or directory isNotExist=true`
- Green baseline: `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCheckCodexWiring_HealthyHomeSkillsNoFinding' -count=1 -v` → both PASS (this run, this tree).
- Plan-phase correction recorded in spec.md HISTORY: t451's error-taxonomy constraint is ALREADY implemented in `codexStaleSkillFinding` — carried as PRESERVE (REQ-CSP-005), not as new work.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
