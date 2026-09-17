# 진행 기록: SPEC-DOCTOR-TEST-CWD-ISOLATION-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13
baseline: `WT-doctor-red@dd235a66b1145922565841d33acafef0d1ded6a8` (RED ledger re-measured 2026-09-18, spec v0.3.0; raw output `.moai/reports/t675/red/`; prior baseline `74d872aafbd90235e67163a5bc233f7c8a934491`)
tier: S
artifacts: `spec.md` + `plan.md` — canonical Tier S set; inline ACs in `spec.md §3`
plan_artifact_hash (sha256, spec/plan v0.3.0):
- `spec.md`: `23fef05b32fb4a0dae34443c25cbf92261f6162cfd63f65c561b94f44769111b`
- `plan.md`: `0fd2f2514a1383185e6e6548b08c62b8fb88f7d4a24820e6f67ee634bacb2cfa`
scope_decision: Operator-directed Tier S CWD isolation supersedes the card-origin embedded-C1 hypothesis and Tier M Class B classification.
red_baseline: With `MOAI_EMBED_CHECK_BIN=/usr/bin/false`, the exact nine-test doctor selection exits 1 with nine failures on the pinned baseline; the representative single test also exits 1.
checks:
- `moai spec lint SPEC-DOCTOR-TEST-CWD-ISOLATION-001 --strict --json` → exit 0 (2026-09-18, v0.3.0), one `info` finding `OwnershipTransitionUnmeasured` (commit `49bf74a82` carries no Authored-By-Agent trailer; not clearable without history rewrite)
- `moai spec audit --base-dir /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675 --filter-spec SPEC-DOCTOR-TEST-CWD-ISOLATION-001 --strict --json` → exit 0, `modern_era_clean: 1`, only `EraAutoDetected` INFO via H-5
prior_gateway_blocker: Preserved at `.moai/reports/t675/gateway-502-20260913.md`; the previous manager-spec delegation ended in HTTP 502 three times before this recovery session.
next_action: Hand off the unchanged Tier S plan artifacts to `plan-auditor`; no implementation has started.

## §E.1a 재배차 재측정 (2026-09-18)

- 흡수 후 트리: `WT-doctor-red@ae6ad726b` (로컬 develop `f67d2193f` 병합)
- 자연 상태 9건: `go test -count=1 -v -run 'TestRunDoctor_|TestDoctorCmd_' ./internal/cli/` → 9건 전부 `--- PASS`, `ok ... 262.827s`
- 오염 주입 재현: `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^TestDoctorCmd_Execution$' ./internal/cli/` → `doctor command RunE error: doctor: 1 check(s) failed`, `--- FAIL`
- 자연 통과의 귀속: 이 트리에 `bin/moai` 가 없어 Agent Emit Embed 가 판정 대상 없음(정보성 건너뜀)으로 끝난다. 그 분기(`f6c027fa0`, SPEC-CI-DOCTOR-BIN-001)는 카드 실측 base `5ddccacc9` 에 이미 있었으므로, 카드 당시의 적색은 로컬에 빌드된 노후 `bin/moai` 가 있을 때만 나는 **환경 상태** 였다고 판단한다 (노후 바이너리 자체로 재현하지는 않았다 — 주입 대조로만 확인).
- 등급: Class B, Tier S 유지. CI 는 `bin/moai` 를 빌드하지 않으므로 배치 push 적색 요인이 아니다 → 우선순위 High 에서 강등 근거. 테스트가 주변 작업 폴더에 결과를 맡기는 결함은 남아 있어 plan 을 이어 간다.
- plan-audit(이전 세션): FAIL 0.67 — D1-D5 수리 후 재감사 필요.

## §E.2 Run-phase Evidence

- 상태: pending — plan 단계 미완료

## §E.3 Run-phase Audit-Ready Signal

- 상태: pending — plan 단계 미완료

## §E.4 Sync-phase Audit-Ready Signal

- 상태: pending — plan 단계 미완료
