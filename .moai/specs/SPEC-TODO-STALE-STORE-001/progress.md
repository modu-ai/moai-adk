# progress.md — SPEC-TODO-STALE-STORE-001 (카드 t1307)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-TODO-STALE-STORE-001
card: t1307
tier: M
status: draft
artifacts:
  - .moai/specs/SPEC-TODO-STALE-STORE-001/spec.md
  - .moai/specs/SPEC-TODO-STALE-STORE-001/plan.md
  - .moai/specs/SPEC-TODO-STALE-STORE-001/acceptance.md
  - .moai/specs/SPEC-TODO-STALE-STORE-001/progress.md
spec_id_check: "PASS (Bash regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ — 충돌 없음 확인)"
code_changes_this_phase: none
open_clarifications: []
resolved_clarifications:
  - "M3 삭제 승인 주체 — 해결됨(2026-09-29, plan-audit iter1 D1): 운영자·리드 양쪽 확인으로 확정. 카드 t1307 본문 「파기 전 운영자·리드 확인」 근거. 확인 기록은 §E.2에 경로·sha256·확인 주체와 함께 기록."
next_phase: run
```

plan-audit iter1 수리(2026-09-29, manager-spec): D1(M3 게이트 확정 + 미해결 질의 마커
제거)·D2(존재하지 않는 AC 오인용 제거)·D3/D4(`-run` 패턴 비공허화 — anchored 패턴은
0개 적중)·D5(고지 진입점 2건 명시 — `todo_history.go:157` 직접 호출 경로 포함)·D6
(`state_dir.go` 인용 범위 :62-86 확대) 반영.

codex cross-audit 수리(2026-09-29, manager-spec, v0.1.2): REQ-TSS-001 5동사 고지의 AC
커버리지 결함 수리(GLM iter2 PASS 0.95를 codex OVERTURN 0.76으로 뒤집은 단일 차단
결함) — AC-TSS-001을 동사별 시나리오(AC-TSS-001a..e)로 분해, AC-TSS-002를 동사별
stdout 바이트 동일성·무고지로 확장, plan.md M1 판정 주체 표기 정합(AC-TSS-003 →
AC-TSS-001a..e).

plan-phase 조사 결과 요약:

- 큐 저장소 계층: `internal/kanban/state_dir.go`(홈 DB·레거시 경로 해석),
  `internal/homestate/paths.go`(project-key), `internal/kanban/backlog_sqlite.go`
  (`meta.last_seq` 키).
- 고지 표면: `internal/cli/todo_disclosure.go`(State D 전용 — 스테일 SQLite 미범위).
- doctor 등록: `internal/cli/doctor.go:187` `runGroupedChecksObserved`; binary_lag 쌍:
  `internal/cli/binary_lag_test.go` `namesAddedAfterBaseline` +
  `TestBinaryLag_AllowlistKeysAreLiveNames`; 골든: `doctor_golden_test.go`
  (`UPDATE_GOLDEN=1`).

## §E.2 Run-phase Evidence

Run phase executed 2026-09-29 in worktree `.claude/worktrees/t1307` (branch
`WT-todo-stale-store`), cycle_type=tdd. Every output below is verbatim from
THIS run against THIS tree.

### M1 — stale-local-store detector + stderr disclosure (commits b363e54b5)

RED (verbatim, pre-GREEN, tree ca2d0e032) — `go test ./internal/kanban/ -run 'TestInspectStaleLocalStores'`:

```
# github.com/modu-ai/moai-adk/internal/kanban [github.com/modu-ai/moai-adk/internal/kanban.test]
internal/kanban/todo_stale_store_test.go:85:10: undefined: InspectStaleLocalStores
FAIL	github.com/modu-ai/moai-adk/internal/kanban [build failed]
```

RED (verbatim, pre-GREEN, detector landed, disclosure unwired) — `go test ./internal/cli/ -run 'TestTodoStaleStoreDisclosure' -count=1` (excerpt; all five verbs):

```
--- FAIL: TestTodoStaleStoreDisclosure_PerVerb (4.41s)
    --- FAIL: TestTodoStaleStoreDisclosure_PerVerb/bare (0.68s)
        todo_stale_store_test.go:108: bare: stderr does not name the stale store path .../.moai/state/todo/backlog.db; stderr=""
    --- FAIL: TestTodoStaleStoreDisclosure_PerVerb/list ...
    --- FAIL: TestTodoStaleStoreDisclosure_PerVerb/why ...
    --- FAIL: TestTodoStaleStoreDisclosure_PerVerb/pr ...
    --- FAIL: TestTodoStaleStoreDisclosure_PerVerb/history ...
FAIL	github.com/modu-ai/moai-adk/internal/cli	25.611s
```

GREEN — `go test ./internal/cli/ -run 'TestTodoStaleStoreDisclosure' -count=1` →
`ok  github.com/modu-ai/moai-adk/internal/cli  33.119s`.
Regression — `go test ./internal/cli/ -run 'TestTodo|TestInspectBacklogArchiveVouch' -count=1` →
`ok ... 370.290s`; `go test ./internal/kanban/ -count=1` → `ok ... 186.260s`.

### M2 — doctor divergence check + binary_lag pair (commit 08ed325ac)

RED (verbatim): `internal/cli/doctor_todo_store_test.go:80:13: undefined:
checkTodoStoreDivergence` (+ todoStoreDivergenceCheckName) → build failed.

Non-vacuity (M2 Gap closure — both anchored enumeration patterns, ≥1 match each) —
`go test ./internal/cli/ -list '^(TestBinaryLag_OneSeamServesBothSurfaces|TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero|TestBinaryLag_AllowlistKeysAreLiveNames|TestBinaryLag_DoctorCheckNameSetIsUnchanged)$'`:

```
TestBinaryLag_OneSeamServesBothSurfaces
TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero
TestBinaryLag_AllowlistKeysAreLiveNames
TestBinaryLag_DoctorCheckNameSetIsUnchanged
```

`go test ./internal/cli/ -list '^(TestDoctorGolden_Light|TestDoctorGolden_Dark|TestDoctorGolden_NoColor)$'`:
`TestDoctorGolden_Light` / `TestDoctorGolden_Dark` / `TestDoctorGolden_NoColor`.

binary_lag suite — same anchored pattern with `-count=1 -v`: 4/4 `--- PASS`, `ok ... 0.895s`.
Doctor golden — regenerated with `UPDATE_GOLDEN=1 go test ./internal/cli/ -run '^(TestDoctorGolden_Light|TestDoctorGolden_Dark|TestDoctorGolden_NoColor)$' -count=1` → `ok`; delta = the new
`ok  Todo Store Divergence  no stale project-local queue store` row + counts (11→12 ok, Pass 24→25) in all three
`internal/cli/testdata/doctor-{light,dark,nocolor}.golden`; verify run without UPDATE_GOLDEN → `ok ... 0.633s`.
All doctor tests — `go test ./internal/cli/ -run 'TestDoctor' -count=1` → `ok ... 64.508s`.

### M3 — residual-store measurement (MEASUREMENT AND RECORD ONLY)

Per the lead's milestone boundary: no deletion, no destructive command. All
four targets verified present after measurement (AC-TSS-021 no-confirmation
→ no-deletion holds). Read-only probes only; note: `sqlite3 ?mode=ro`
queries touch the `-shm` sidecar mtime (WAL reader coordination) — main
`.db` bytes and mtimes unchanged (store-1 backlog.db mtime still
2026-09-13T21:56:50, sha256 9802d3f6… across both reads).

| Target | Size | mtime | sha256 | meta.last_seq (ro read) |
|---|---|---|---|---|
| `.moai/backlog.db` (primary) | 0 B | 2026-09-02T07:32:02 | n/a (empty) | n/a |
| `.moai/state/backlog.db` (primary) | 0 B | 2026-09-02T07:32:02 | n/a (empty) | n/a |
| `.moai/state/todo-merge-backup-20260913/store-1/` (backlog.db 811008 B; backlog.json 652478 B; backlog.json.migrated 155043 B; wal 0 B; tree 1656 KB) | see left | backlog.db 2026-09-13T21:56 | `9802d3f6d1c4e7c6d3c4d38d15e13a8d8283ff0eff5a5a0c992a8d45d26ce590` | 661 |
| `.moai/state/todo-merge-backup-20260913/store-2/` (backlog.db 950272 B; wal 0 B; tree 992 KB) | see left | backlog.db 2026-09-13T21:56 | `e190c863fca8503114f8049bdd345ca3b5d5d4c5f66a395ad79c95d97e525904` | 708 |

Ghost store (the M1/M2 subject, measured for context):
`.moai/state/todo/backlog.db` = 9802d3f6…ce590 — **byte-identical to
backup store-1's backlog.db**; mtime 2026-09-11T03:20:10; last_seq 661.

Home database (canonical queue, ro read this run):
`~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` last_seq = **1328** (the
SPEC's cited 1305 was the 09-29-morning value; the queue advanced). Live
divergence gap vs the ghost store: 1328 − 661 = **667 seq**.

Reference judgment (positive control: `grep -rln SPEC-TODO-STALE-STORE-001
.moai/specs/` → 3 files; same search shape): `grep -rn todo-merge-backup
internal/ cmd/ pkg/ .moai/docs/ .moai/project/ scripts/` → **0 matches** —
no production code or repo doc references the backup trees or the 0-byte
pair. The home DB is live and advancing (1328 > 708 > 661); both backups
hold strictly older states.

**Disposition gate (REQ-TSS-021): UNMET BY DESIGN this phase** — awaiting
lead+operator disposition decision. No deletion executed; no automatic
disposal path exists; all four targets exist as of this report.

### AC PASS/FAIL Matrix (commands + verbatim outputs this run, this tree)

| AC | Status | Verification | Observed |
|---|---|---|---|
| AC-TSS-001a (bare) | PASS | `go test ./internal/cli/ -run 'TestTodoStaleStoreDisclosure_PerVerb/bare' -count=1 -v` (within suite) | `--- PASS` inside `ok ... 33.119s`; stderr carries ghost path + `1305` + `661` |
| AC-TSS-001b (list) | PASS | same suite, subtest list | `--- PASS` |
| AC-TSS-001c (why) | PASS | same suite, subtest why | `--- PASS` |
| AC-TSS-001d (pr) | PASS | same suite, subtest pr | `--- PASS` |
| AC-TSS-001e (history) | PASS | same suite, subtest history | `--- PASS` (RED-first captured: history stderr was empty pre-wiring) |
| AC-TSS-002 | PASS | `TestTodoStaleStoreDisclosure_JSONStdoutByteIdentity` (8 arg-variants incl. `--json`) + `TestTodoStaleStoreDisclosure_NoDisclosureWithoutDivergence` | `ok ... 33.119s`; stdout byte-identical divergent-vs-clean; no disclosure at equal seqs |
| AC-TSS-003 | PASS | `TestTodoStaleStoreDisclosure_ReadPathPurity` | `--- PASS`; sha256+mtime+dir listings unchanged across all five verbs |
| AC-TSS-004 | PASS | code inspection (grep, this run) | `InspectStaleLocalStores` called from exactly todo_disclosure.go:99, todo_history.go:165, doctor_todo_store.go:37; `backlogMetaKeyLastSeq` SQL exists only in the detector + pre-existing engine/reader |
| AC-TSS-010 | PASS | `TestDoctorTodoStoreDivergence_ThreeStates` (5 subtests: divergent/home-absent/legacy-absent/unreadable/equal) | `ok ... 2.424s` |
| AC-TSS-011 | PASS | anchored binary_lag pattern, `-count=1 -v` | 4/4 `--- PASS`; allowlist carries bare key `todoStoreDivergenceCheckName`; `-list` shows 4 matches (non-vacuous) |
| AC-TSS-012 | PASS | golden regen + verify (same commit 08ed325ac) | 3 golden files modified in commit; `ok ... 0.633s` without UPDATE_GOLDEN |
| AC-TSS-013 | PASS | `TestDoctorTodoStoreDivergence_ReadOnly` | `--- PASS`; no db sha/mtime change, no new files |
| AC-TSS-020 | PASS | §E.2 M3 measurement table (this run) | 4/4 targets measured before any disposition; recorded above |
| AC-TSS-021 | PASS (gate holds) | post-measurement existence check | `EXISTS:` × 4 (verbatim above); zero deletions without operator+lead confirmation |
| AC-TSS-022 | **awaiting disposition** | n/a — disposition not executed | No deletion happened, so no disposal evidence exists yet; the measurement record above is the §E.2 input the disposition decision will consume |

### Quality gates

- Coverage (new code, `go tool cover -func`, this run): `InspectStaleLocalStores` 94.7%, `readStaleStoreLastSeq` 80.0%, `discloseStaleLocalStores` 100%, `todoQueueRootForDisclosure` 100%, `checkTodoStoreDivergence` 96.7% (doctor profile), `discloseQueueLayout` 66.7% (error-propagation branch of the PRE-EXISTING first probe uncovered — not new logic).
- `go vet ./internal/cli/ ./internal/kanban/` → clean. `gofmt -l internal/cli internal/kanban` → empty.
- `golangci-lint run ./internal/kanban/... ./internal/cli/...` (v2.1.6 = CI version) → `0 issues.`
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- Local full-suite run deliberately NOT executed (CLAUDE.local.md §4 / gitflow-lane-protocol §8); full-suite verdict is CI's.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-29
run_commit_sha: 08ed325ac
run_status: complete (AC-TSS-022 awaiting lead+operator disposition decision)
ac_pass_count: 13
ac_fail_count: 0
ac_awaiting: 1 (AC-TSS-022 — disposition not executed; gate unmet by design)
preserve_list_post_run_count: 4 (residual stores untouched, all verified present)
l44_pre_commit_fetch: not-performed (worktree card flow; develop push is lead-batched)
l44_post_push_fetch: not-performed (same)
new_warnings_or_lints_introduced: 0 (golangci-lint v2.1.6 → 0 issues)
cross_platform_build.darwin: pass
cross_platform_build.windows: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: 12 (6 M1 + 7 M2 incl. 3 goldens; spec.md frontmatter transition shared)
m1_to_mN_commit_strategy: per-milestone commits (M1 b363e54b5, M2 08ed325ac, M3 records-only)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-29
sync_commit_sha: "pending-backfill-sync"
sync_status: complete
b12_self_test_a: "PASS — grep -c 'SPEC-TODO-STALE-STORE-001' CHANGELOG.md = 0 before emission (no duplicate from parallel BATCH-SYNC)"
b12_self_test_b: "PASS — acceptance.md SSOT: 11 distinct AC identifiers (AC-TSS-001 counted once with five per-verb sub-rows a–e); executable scope 13 PASS / 0 FAIL; AC-TSS-022 awaiting-disposition BY DESIGN (disposal is an operator decision record, out of executed scope) — CHANGELOG entry states this split verbatim"
b12_self_test_c: "PASS — every cited path verified via ls before emission (internal/kanban/todo_stale_store.go, internal/cli/todo_disclosure.go, internal/cli/todo_history.go, internal/cli/doctor_todo_store.go, internal/cli/binary_lag_test.go, internal/cli/testdata/doctor-*.golden)"
changelog_entry_position: "[Unreleased] > Added"
frontmatter_status_transitions:
  implemented_completed: "merged into THIS sync commit (3-phase close; status: in-progress → completed + updated: 2026-09-29, spec.md frontmatter only)"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy requiring its own sync tests"
mx_tag_validation: "pass — 1 ANCHOR (+REASON) on InspectStaleLocalStores (fan_in=3: todo_disclosure.go:99, todo_history.go:165, doctor_todo_store.go:37); 0 TODO resolved (no RED-phase TODO survived to GREEN); no WARN required (no goroutine/complexity>=15 construct in new code)"
```

AC-TSS-022 disposition note (honest close): the SPEC completes its executable
scope with REQ-TSS-021's confirmation gate intact — no deletion was executed,
so AC-TSS-022's disposal evidence does not exist and is recorded as an operator
decision record (lead+operator confirmation per card t1307), not as a failed AC.
The §E.2 M3 measurement table is the input that decision consumes.

## §G Gate Disposition Log

- 2026-09-29 — served_model_gate 해제(이 트리 한정): 운영자 승인(리드 전달 2026-09-29) — 게이트 해제 + GLM iter2 PASS 0.95 채택, 단 codex 재감사 OVERTURN 시 채택 제외. plan-auditor 정의 선언 모델(opus)과 실제 서빙 모델(glm-5.3-flash) 불일치로 served-kind 거절이 서서 페이즈 진입 스폰이 차단되던 것에 대한 처분.
- 2026-09-29 — 제외 조항 소진: codex(GPT-6) 재감사 OVERTURN 0.76(REQ-TSS-001 5동사 고지 vs AC 커버리지 결함)으로 GLM 채택은 제외됨. GLM plan-audit 반복 상한 2/2 소진 — 수리 후 codex 델타 재감사가 유일 재심 경로이며, RECONFIRM 시 run 진입.
- 거부 영수증 파일(`.moai/state/audit-receipts/rejections/plan-auditor--unknown-spec--served.json`)은 지시에 따라 보존(삭제하지 않음). opus 클리어 감시 루프는 경로 폐쇄 확정으로 종료.
