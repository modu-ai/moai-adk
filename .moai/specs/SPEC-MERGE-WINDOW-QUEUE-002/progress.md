# SPEC-MERGE-WINDOW-QUEUE-002 — Progress

> Card t1582 · created 2026-10-09 by manager-spec (plan phase) · branch `WT-p1-p2-t1576` · base develop tip `f7606c7bc`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md + plan.md + acceptance.md (Tier M set) + decision-index.md (`interview.decision_gate: on`) + progress.md.
- SPEC id regex check (Bash) → `PASS`; uniqueness: `ls .moai/specs/ | grep -i MERGE-WINDOW` → SPEC-MERGE-WINDOW-QUEUE-001 만 존재 (002 자유).
- RED 재현 (plan 단계 실측, 트리 `f7606c7bc`, env-scrub 복합 형태):
  - ④-R7-2: `TestRedT1582MixedSweepWithNoTestPackageCounts` — `runner reported [no test files] — an empty sweep cannot stand for a re-measure` (거짓 거부 관측).
  - ④-R7-3 단위: `TestRedT1582VerifierAdmitsAShortBaseSHA` — 검증기가 3바이트 Base 통과 관측.
  - ④-R7-3 e2e: `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease` — `runtime error: slice bounds out of range [:12] with length 3`, release 전 사망 + 창 held 관측.
  - item ②: round-trip 관측 — 경계 붕괴 **미재현**(4/5 일치, apostrophe는 escape 바이트 잔류·경계 유지) → 봉인 테스트 GREEN.
  - item ①: probe 관측 — non-test 명령의 실패 의미 출력이 exit 0으로 valid (spec.md §D 잔여, 수리 대상 아님).
  - item ③·R7-1: 재현 설계 확정 — 관측은 run 단계 M3/M4 첫 행위 (plan.md §F).
- RED 오버레이 테스트 파일: `internal/factory/remeasure_red_t1582_test.go`, `internal/cli/shelljoin_red_t1582_test.go` (트리에 작성, 미커밋 — run 단계 M1-M4가 GREEN으로 뒤집는 대상).
- 10 REQ / 8 AC (v0.2.0 — D5 분할: 002/008, 003/009, 006/010). Out of Scope 경계: 보호구역·todo 발행·nominate(리더 큐), non-test 명령 잔여, 호출자 리다이렉션, `merge gate` verdict 설계, complete의 release-failure 출력.
- plan-audit iteration 1: FAIL 0.69 / Tier M 0.80 (`.moai/reports/t1582/plan-audit.md`) — D1-D6 아티팩트 수리 완료(v0.2.0, plan.md·acceptance.md·spec.md).
- plan-audit iteration 2: FAIL 0.85 (동일 파일 Iteration 2 섹션) — 라운드 2 수리 완료(v0.3.0): D19(M1 렌더 12곳 실측 열거+총괄 규칙+cli 스코프 — 감사 prose "14"는 자기 목록 10+2의 오산술, 리더 독립 grep과 본인 측정 일치)·D13(§4 목록 교체)·D14(cause-7 시딩 설비 `internal/factory/mergestep_red_t1582_test.go` + merge-ready RED `internal/cli/factory_merge_ready_red_t1582_test.go` plan 단계 작성·실측, AC-005/006 셀 verbatim 기록, AC-006 실행 계수 가드 `--- PASS` ≥ 2)·D15(스크럽 변수 3종 전체 목록 명기+축약형 전면 제거)·D17(DoD tracked+untracked 수집 대조)·D20(GNU grep `-r`+exit 구분)·D18(version 0.3.0 정합). RED 총 4건 실측: R7-2·R7-3 단위/e2e·merge-ready ③.
- 429 재개 기록: 라운드 2 중 429(요청 한도) 2회 — 트랜스크립트 재개로 잔여 목록(디스크 상태 대조) 수행, 부분 상태는 위 행들이 증언.
- plan_status: audit-ready + ceiling-exception approved (§G 참조)

## §G Plan-audit Ceiling Exception Record

- 판정 궤적: iteration 1 FAIL 0.69 → 2 FAIL 0.85 → 3(천장) FAIL **0.91** — 단조 상승, STOP 신호 없음. 영수증 3건: `rcpt-a221eb42d027db89e2c7fac2`·`rcpt-33b7845356c311a4cddf2a1c`·`rcpt-7b69c79f30c6981fbbcc0aa1` (codex required 게이트 3/3 응답, 매 라운드 축소하는 형식 지적).
- 최종 잔여: acceptance.md 형식 hunk 2건(D21 DoD 허용 목록 내부 모순·D22 RED stdout 장부) — 코드 결함 0, 코드 변경 0줄.
- **운영자 결정: 천장 예외 승인·run 진입** — 근거 "0.91 단조·STOP 없음·잔여 형식 2건 수리 완료·코드 0줄". 리더 경유 전달(2026-10-09 21시경, msg), lane-26 기록.
- 판정 후 조치: D21·D22 hunk를 감사자의 required-fix 명세대로 수리(acceptance.md 20:30, §D.3 증거 장부 E-LEDGER-001~004 신설 포함) — 천장으로 재감사 없음, 본 행이 그 상태의 기록.
- decision record: decided_by=operator+leader evidence_refs=.moai/reports/t1582/plan-audit.md iter-3 ladder_path=리더-경유 운영자 결정(keep-set 인접 게이트의 정식 처분) — Kickoff 재개.

## §F Phase 4 Mode Selection

- Input parameters: tier=M, scope=4 소스 파일+4 RED 오버레이+테스트 재작성 1(internal/factory 2·internal/cli 2), domain count=1(Go CLI/factory 표면), file language mix=Go 100%, concurrency benefit=LOW(coding-heavy — 마일스킨 의존: M1 검증기→M2 판정→M3 exit→M4 원자화 순차), agent-team prereqs=미요청.
- Mode evaluation: direct 미선정(다중 파일·다중 마일스톤), fanout 미선정(coding-heavy — Anthropic 병렬화 주의), sweep 미선정(기계 균일 변환 아님), agent-team 미선정(명시 요청 없음).
- Decision: `serial`
- Justification: 단일 표면의 순차 의존 수리 패킷 — 마일스톤 M1→M4가 서로의 산출 위에 서고 같은 파일군을 건드리므로 단일 manager-develop 위임이 재위임 위험과 쓰기 경합을 최소화한다(Tier M 전체 Section A-E 템플릿 적용).

## §E.2 Run-phase Evidence

Run by manager-develop (serial, tier M). Baseline HEAD `bb245d325`; RED baseline commit `d7f4fcf0c`. Every command below ran in this worktree against the stated tree. Test runs use the env-scrub compound `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 …`. Full logs are machine-local scratch under `.moai/state/verify/ea52787a-4b79-40a9-8cf3-a5951fe197b2/` (gitignored); the lines quoted here are the deciding evidence.

### E.2.1 Commits per milestone

| Milestone | Commit | Subject | RED → GREEN (verbatim) |
|---|---|---|---|
| M1 — R7-3 SHA format + length-safe prefix (AC-001/002) | `a560b3b6f` | `fix(...): M1 …` | `TestRedT1582VerifierAdmitsAShortBaseSHA` FAIL → `--- PASS: TestRedT1582VerifierAdmitsAShortBaseSHA (0.00s)`; `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease` FAIL → `--- PASS: … (1.70s)` |
| M2 — R7-2 total per-test pass judgement (AC-003) | `eac7fab39` | `fix(...): M2 …` | `TestRedT1582MixedSweepWithNoTestPackageCounts` FAIL → `--- PASS: … (0.00s)` |
| M3 — ③ measurement-failure exit (AC-005) | `64aba35c1` | `fix(...): M3 …` | `TestRedT1582MergeReadyMeasurementFailureStillExitsZero` FAIL → `--- PASS: … (1.18s)` |
| M4 — R7-1 cause-7 hold+release in one section (AC-006) | `b50d61bb5` | `fix(...): M4 …` | `TestRedT1582Cause7HoldAndReleaseAreOneMutation` (new) FAIL → `--- PASS: … (2.93s)` |
| M5 — ② seal + same-class sweep (AC-004) | none (no code change) | — | seal tests GREEN throughout; scans in E.2.3 |
| M6 — quality chain (edge branches) | `a8f49f6ad` | `test(...): M6 …` | four new branch tests PASS (see E.2.4) |

Items ② (`TestRedT1582ClassifiersReadBackTheJoinedBoundaries`, `TestRedT1582JoinPreservesArgBoundariesInQuoting`, `TestRedT1582SingleArgScrubCompoundStaysVerbatim`) and ① (`TestRedT1582ProbeNonTestCommandFailureSemanticsRideExitZero`) stayed PASS and unchanged, as the sealed/observed items require.

### E.2.2 RED baseline (E8) — verbatim, pinned tree `bb245d325`, before any change

Log: `preflight-red-baseline.log`. Command: `unset … && go test -count=1 -v -run '^TestRedT1582' ./internal/factory/ ./internal/cli/` (exit 1).

```
--- FAIL: TestRedT1582MixedSweepWithNoTestPackageCounts (0.00s)
    remeasure_red_t1582_test.go:35: RED t1582-R7-2: the [no test files] marker anywhere in the stream refuses a mixed sweep that measured 1 passing test: runner reported [no test files] — an empty sweep cannot stand for a re-measure
--- FAIL: TestRedT1582VerifierAdmitsAShortBaseSHA (0.00s)
    remeasure_red_t1582_test.go:58: RED t1582-R7-3: the verifier admits a record whose Base is 3 bytes — it reaches the [:12] message renders downstream
--- FAIL: TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease (1.88s)
    remeasure_red_t1582_test.go:89: RED t1582-R7-3: the merge step panicked on the short base "bad" (runtime error: slice bounds out of range [:12] with length 3) and died before releasing the window — the window stays held
--- FAIL: TestRedT1582MergeReadyMeasurementFailureStillExitsZero (2.17s)
    factory_merge_ready_red_t1582_test.go:41: RED t1582-AC005: the measurement-failure REFUSED verdict still exits 0 (err nil) — out: merge-readiness: REFUSED — failing condition: re-measure-record …
--- PASS: TestRedT1582Cause7SeedingWritesHoldThenReleases (2.85s)
--- PASS: TestRedT1582ClassifiersReadBackTheJoinedBoundaries (0.00s)
--- PASS: TestRedT1582ProbeNonTestCommandFailureSemanticsRideExitZero (0.00s)
--- PASS: TestRedT1582JoinPreservesArgBoundariesInQuoting (0.00s)
--- PASS: TestRedT1582SingleArgScrubCompoundStaysVerbatim (0.00s)
```

M4 RED, observed on the pre-repair code (HEAD `64aba35c1`) with the `AfterHold` probe added (log: `m4-red-cause7.log`):

```
--- FAIL: TestRedT1582Cause7HoldAndReleaseAreOneMutation (1.71s)
    mergestep_atomic_t1582_test.go:75: RED t1582-R7-1: a status refresh landed between the cause-7 hold write and the window release, and the caller's release then failed (step error: … (releasing the window also failed: no release integration window is held — moai integration status reads it))
```

### E.2.3 Same-class sweep (M5) — scans with exit codes

Exit codes: 0 = matches found, 1 = no match, 2 = search error. No scan returned 2.

- **A** `grep -rn '\[:12\]' internal/factory internal/cli --include='*.go'` → exit 0. Non-test hits: `integration_remeasure.go:144` (comment only); `internal/cli/contract_revoke.go:76` `shortSeal` (already guarded by `if len(s) > 12`); `internal/cli/doctor_disk.go:345` `digest[:12]` (content digest, not a git SHA, outside the merge surface). No unguarded SHA prefix render remains on the merge or remeasure surfaces.
- **B** `grep -rn 'Contains.*no test\|noTestFiles\|emptySweepMarker' internal/factory internal/cli --include='*.go'` → exit 0, matches only in `*_test.go`. No marker refusal remains in non-test code.
- **C** `grep -n 'strings.Fields' internal/factory/integration_remeasure.go` → exit 0, one hit at `:428` in `requestsGoTestJSON`, splitting a GOFLAGS value (already one shell word) into flags. Intended; it is not the command-line split.
- **D** hold-then-release pairs: `writeMergeHold` survives only in `postMergeHold` (cause 8). Residual, see E.3 gaps.
- **E** `AfterHold`: exactly one production call site, in `holdThenReleaseCauseSeven`.

### E.2.4 Quality chain (final state, verbatim outcomes)

- `go build ./...` → exit 0. `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (run at the final state; constraint B1 asks for each milestone, and the milestone diff adds no platform-specific API: scan of added lines finds only `os/exec`).
- `gofmt -l internal/factory internal/cli` → `internal/factory/remeasure_red_t1582_test.go` only. That file is the RED overlay: unformatted at HEAD `d7f4fcf0c`, and this run must not edit it. No file changed by this card is listed (`integration_remeasure_test.go` was formatted in M2, one pre-existing whitespace line).
- `go vet ./internal/factory/ ./internal/cli/` → exit 0.
- `golangci-lint run ./internal/factory/... ./internal/cli/...` → `0 issues.`
- AC-MWQ2-007 family command over both packages: `ok github.com/modu-ai/moai-adk/internal/factory 64.338s`, `ok github.com/modu-ai/moai-adk/internal/cli 26.387s` (run before the M6 test additions; the factory package then ran in full, below).
- Race, scoped families, scrub form: factory `ok … 79.808s` at the M4 state with the clean config; cli `ok … 28.163s` at the M3 state. No `WARNING: DATA RACE`.
- E3 coverage, full package at the final state: `go test -count=1 -coverprofile … ./internal/factory/` → `ok … 298.623s coverage: 83.9% of statements`.
- E3 coverage, `internal/cli`, family-scoped (not a package measure): `ok … coverage: 6.7% of statements`.
- Per modified file (final profile): `integration_remeasure.go` 90.7%; `integration_merge_step.go` 83.7%; `integration_lock.go` 81.8%.
- Per modified function (final profile): `ValidateRemeasureRecord` 100.0%, `isFullSHA` 100.0%, `countGoTestJSONTests` 100.0%, `ShortSHA` 100.0%, `holdThenReleaseCauseSeven` 100.0%, `ReleaseIntegrationLock` 100.0%, `releaseIntegrationLockLocked` 85.0%, `RunMergeStep` 83.2%, `RunRemeasure` 73.8%.
- E4 boundary grep `grep -rn 'AskUserQuestion' internal/factory internal/cli --include='*.go' | grep -v _test.go`: matches exist, all in pre-existing comments of unrelated cli files (`plan.go`, `harness.go`, `hook.go`, …). None in `internal/factory`, none in a file this card changed, and no invocation anywhere. Literal "print nothing" is not met; see gaps.
- Test additions in M6 (`a8f49f6ad`): `TestRemeasureRecordRefusesMalformedTreeAndBase` (4 cases), `TestRemeasureRecordRefusesMissingIdentity`, `TestMergeStepCause7HoldWriteFailureKeepsWindowHeld`, `TestMergeStepCause7ReleaseFailureIsReportedWrapped`. All PASS.
- Scratch probe results: `moai verify check --key-current` → `Stale: no snapshot recorded for the current working-tree key` on every gate run, so no snapshot result was cited as evidence.

### E.2.5 Commands used (representative; all in the card worktree)

- `git rev-parse --show-toplevel && git branch --show-current && git rev-parse --short HEAD` (location guard, exit 0; toplevel `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1582`).
- `moai slot acquire|release --resource go-test-heavy` around each race or full-package run.
- `git add -- <explicit pathspecs>` and `git commit -F <message file>` for every commit; `git status --short` clean after each.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-10T02:58:51+09:00
run_commit_sha: a8f49f6ad
run_status: complete-with-gaps
ac_pass_count: 6
ac_pass_with_debt_count: 1          # AC-MWQ2-005 (built-binary observation not performed)
ac_fail_count: 1                    # AC-MWQ2-008 literal gofmt — blocker, see gaps
preserve_list_post_run_count: 0     # no PRESERVE target modified; all 14 changed files inside allowed scope
l44_pre_commit_fetch: not-applicable  # nothing pushed (B9); no fetch performed in this run
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues; go vet exit 0
cross_platform_build:
  native_go_build: exit-0
  windows_amd64_go_build: exit-0
total_run_phase_files: 14
m1_to_mN_commit_strategy: one commit per milestone (M1 a560b3b6f, M2 eac7fab39, M3 64aba35c1, M4 b50d61bb5, M6 a8f49f6ad as a test-only quality commit); M5 has no code change and is recorded in E.2.3
```

### E.3.1 AC matrix (command → observed result)

| AC | Status | Observed result (verbatim where short) |
|---|---|---|
| AC-MWQ2-001 — malformed Base/Tree refused as record-invalid | PASS | `--- PASS: TestRedT1582VerifierAdmitsAShortBaseSHA (0.00s)`; edge `TestRemeasureRecordRefusesMalformedTreeAndBase` PASS (short base, short tree, uppercase tree, 39-char base) |
| AC-MWQ2-002 — short base renders length-safe, window released, C promoted | PASS | `--- PASS: TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease (1.70s)`; `ok … 1.908s` |
| AC-MWQ2-003 — mixed sweep valid; fail events and total-zero stay refused | PASS | `--- PASS: TestRedT1582MixedSweepWithNoTestPackageCounts (0.00s)`; family `TestClassify*` PASS; negative `TestClassifyAllNoTestSweepStaysEmpty`, `TestRemeasureGoTestJSONRecordsCount` PASS |
| AC-MWQ2-004 — join/classify boundary seal (no repair) | PASS | the three `TestRedT1582…` seal tests PASS unchanged in every family run |
| AC-MWQ2-005 — measurement-failure REFUSED exits non-zero; waiting keeps 0 | PASS-WITH-DEBT | `--- PASS: TestRedT1582MergeReadyMeasurementFailureStillExitsZero (1.18s)`; `--- PASS: TestFactoryMergeReadyMeasurementFailureResolvesNonZeroExit`; `TestFactoryMergeReady_HeldWindowRefusedWithHolderNamed` (waiting, exit 0) PASS. Debt: built-binary observation not performed (gap 1) |
| AC-MWQ2-006 — hold+release atomic; execution-count guard ≥2 | PASS | `--- PASS: TestRedT1582Cause7HoldAndReleaseAreOneMutation (2.93s)`, `--- PASS: TestRedT1582Cause7SeedingWritesHoldThenReleases (2.20s)` (count 2) |
| AC-MWQ2-007 — affected families green | PASS | exact command, both packages: `ok … internal/factory 64.338s`, `ok … internal/cli 26.387s`; full factory package `ok … 298.623s` at the final state |
| AC-MWQ2-008 — gofmt empty, vet exit 0 | FAIL (literal) | vet: exit 0. gofmt: lists `internal/factory/remeasure_red_t1582_test.go` (pre-existing RED overlay; gap 2) |

### E.3.2 Gap list (not observed, or not met)

1. **AC-MWQ2-005 built-binary observation** (`make build` then `./bin/moai factory merge ready …` on the measurement path) was not performed. The measurement path needs a lane fixture (lane-label environment, a sync-complete SPEC, a merge-able card branch). This card's own SPEC has no `§E.4` yet, so an in-tree run would refuse at sync-audit, which is a different class. Substitutes: the in-process cobra run and the `ResolveExitCode` mapping guard.
2. **AC-MWQ2-008 literal gofmt — BLOCKER for the leader.** `remeasure_red_t1582_test.go` is unformatted at HEAD `d7f4fcf0c` and is the RED overlay this run must not edit. Decide: format the overlay in a sanctioned change, or rewrite AC-008 to exclude RED overlays.
3. **E4 literal boundary grep** prints matches, all in pre-existing comments of unrelated cli files. No code in `internal/factory` and no file changed by this card matches; there is no invocation.
4. **Repository-wide test verdict not run** (lane rule: no `go test ./...`). The CI run on the integration branch owns that verdict; **PENDING at report time**. The full `internal/cli` suite is structurally red inside card worktrees, so its verdict is a gap; the affected families are green.
5. **Coverage below 85%:** `integration_merge_step.go` 83.7%, `integration_lock.go` 81.8%, `RunMergeStep` 83.2%, `RunRemeasure` 73.8%. Their uncovered lines are error branches in untouched code paths (git-error returns, lock edge cases) and are not chased in this run.
6. **Cause 8 residual:** `postMergeHold` writes its hold and releases in two mutations, the same defect class as R7-1 but outside REQ-MWQ2-007's named cause-7 outcome. Recorded, not expanded here (scope rule). Needs a decision: a follow-up SPEC or a REQ amendment.
7. **Decision-index Q1 unresolved on the record:** the operator-verdict field in `decision-index.md` is blank. M3 implements the recorded Default (measurement-failure class only, waiting unchanged), as the spawn instructed. Operator confirmation is needed.
8. **Unexplained config change in this worktree:** `.moai/config/sections/workflow.yaml` flipped `workflow.codex.review_gate.enabled` from `true` to `false` at 02:32:09 local, during the M4 race leg. Provenance is not established: `slot status|acquire|release` and the factory cause-7 run did not reproduce it. Restored to HEAD and not committed. The writer should be identified.
9. **Plan-audit known gap** (acceptance.md D21/D22 repaired after the FAIL verdict without re-audit): the acceptance criteria were checked for consistency against the implementation. The only inconsistency found is AC-008's literal gofmt criterion (gap 2). Nothing was silently fixed.

### E.3.3 Residual risk (could still be wrong despite the observations)

- AC-006's GREEN side is deterministic (a mutation blocked on the section cannot finish during the probe). Its RED side depends on the 750 ms probe window staying below the 1.65 s mutation budget. If the budget changes, the probe must be re-derived; the test's comment says so.
- The lowercase-only SHA decision (REQ-MWQ2-001) refuses uppercase hex. Git prints lowercase, but a hand-written record in uppercase is refused, not normalized.
- The commit-tier hook ran on M1–M4 without blocking. Whether it re-runs repository-wide tests at commit time was not observed.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete-with-gaps
sync_complete_at: 2026-10-09T18:27:39Z
sync_commit_sha: 5d6c9b8c33d5901fe531f8d8373df80754edbe17
sync_tier: M
ac_source: .moai/specs/SPEC-MERGE-WINDOW-QUEUE-002/acceptance.md
ac_live_count_for_record: 8
b12_self_test_a: pass
b12_self_test_b: not-applicable-no-count-stated
b12_self_test_c: pass
changelog_entry_position: "CHANGELOG.md [Unreleased] - Fixed group, first bullet"
frontmatter_status_transitions:
  spec_md_status: in-progress -> completed
  spec_md_updated: 2026-10-10
  plan_md_status_field: absent
  acceptance_md_status_field: absent
  progress_md_status_field: absent
```

### E.4.1 Sync evidence (this run)

- Pre-emission check: `grep -c 'SPEC-MERGE-WINDOW-QUEUE-002' CHANGELOG.md` returned 0 before the entry was appended.
- AC counter on the Tier M acceptance file named by `ac_source`: live=8, excluded=0, ambiguous=0. The CHANGELOG entry states no AC count (sync instruction); the count is recorded for attribution only.
- Every path cited in the CHANGELOG entry was checked with `ls`.
- `spec.md`: `status: in-progress` changed to `status: completed`. The `updated:` line already read 2026-10-10, so it was not edited. `plan.md` and `acceptance.md` carry no status field and no updated field. `progress.md` has no status line.
- The `spec.md` body, HISTORY, and other frontmatter fields are untouched. §E.2, §E.3, and §J are untouched.
- Clock note: `sync_complete_at` is the UTC instant. The frontmatter `updated:` date is the local KST date (2026-10-10). Both are correct for their own clock.

### E.4.2 MX-tag validation (comment-only anchors added in the gap-closure commit)

Method: `grep -n '@MX'` over the seven changed production files; `git diff -U0 8673c2a95..HEAD` for changed and added functions; non-test call sites counted with `grep ... | wc -l`; distinct enclosing functions counted by mapping each call line to its nearest preceding `func` line (`xargs awk 'FNR==1{fn="(top)"} /^func /{fn=substr($0,1,70)} /NAME\(/ && !/^func NAME\(/ {print fn}' | sort -u | wc -l`); a `go func` scan over the same files. The dedicated `moai mx` scan was not run, because the installed moai build is not valid evidence for this card.

- The card removed 0 `@MX` lines. The seven files contain no goroutine launch.
- Mandatory ANCHOR (fan_in >= 3), introduced by this card: `ShortSHA` (internal/factory/integration_merge_step.go; exported; new). Added in the gap-closure commit as a comment-only `@MX:ANCHOR` block (with `@MX:REASON` and `@MX:SPEC`) above the declaration. Count: `grep -rn 'ShortSHA(' internal --include='*.go' | grep -v '_test\.go' | grep -v 'func ShortSHA(' | wc -l` → `16` non-test call-site lines across four files. Distinct enclosing functions: `7` (factoryCompleteCard, completePostMergeConflict, newIntegrationMergeCmd, newIntegrationRemeasureCmd, remeasureVerdictError, RunMergeStep, postMergeHold).
- Mandatory ANCHOR (fan_in >= 3), pre-existing (not introduced by this card): `ReleaseIntegrationLock` (internal/factory/integration_lock.go). Count: `grep -rn 'ReleaseIntegrationLock(' internal --include='*.go' | grep -v '_test\.go' | grep -v 'func ReleaseIntegrationLock(' | wc -l` → `8` call-site lines across five files, which is at least three, so the anchor is added. Distinct enclosing functions, same method: `6` (factoryCompleteCard, completePostMergeConflict, autoMergeNoticef, newFactoryMergeReadyCommand, newIntegrationReleaseCmd, releaseHeldWindow). The file carried zero `@MX` lines before this commit. Added as a comment-only `@MX:ANCHOR` block (with `@MX:REASON`) above the declaration.
- Present and consistent: `ValidateRemeasureRecord` carries `@MX:ANCHOR` (internal/factory/integration_remeasure.go; 5 call sites across 4 files, matching its REASON text). `RunMergeStep` carries `@MX:WARN` for complexity; kept.
- Below the fan_in threshold, no mandatory tag: `isFullSHA` (1 caller), `holdThenReleaseCauseSeven` (1), `releaseIntegrationLockLocked` (2), `RunRemeasure` (1), `countGoTestJSONTests` (1), `requestsGoTestJSON` (1), `writeMergeHold` (1), `postMergeHold` (1), `remeasureVerdictError` (1).
- Not measured: cyclomatic complexity of the changed functions (no dedicated tool run). The `RunMergeStep` WARN is carried from its tag text and was not re-measured.

### E.4.3 Gap list (carried; not fixed in the sync phase)

1. AC-MWQ2-008 literal gofmt: resolved by commit `8f01ecd2f` (style: gofmt the pinned RED overlay). The change is formatting-only: `git diff -w --stat 8f01ecd2f^ 8f01ecd2f` is empty, and `gofmt -l internal/factory internal/cli` is empty (both checked in this sync). The RED family re-run gave 10 top-level PASS, 10 subtests PASS, 0 FAIL, both packages ok; that re-run is the lane's, reported here and not repeated in this sync. AC-MWQ2-008's literal criterion is met (gofmt half checked in this sync; the vet half is the lane's run in §E.2.4). (Run-phase §E.3.2 item 2.)
2. Coverage: 기존 미달, 이 카드가 낮추지 않음. `internal/factory` is at 83.9% of statements at the card (`ok 297.437s`, measured at HEAD `b6b17e5ea`), below the 85% package threshold. Base at `8673c2a95`, same command, export at `$HOME/.cache/moai-t1582-base` (outside /tmp): 83.3% of statements. The base figure is approximate: one test failed for an environment reason (`TestMergeStepPreMergeCausesReleaseWithDistinctCodes`; git exit 128 because the export has no .git). Production code is identical to `8f01ecd2f`: no production file changed between `b6b17e5ea` and `280f6ecae`. This commit adds comment-only MX lines and changes no statement. Both coverage figures were supplied by the lane and not re-measured here. No pass is claimed against the 85% threshold.
3. `decision-index.md` Q1: 운영자 판정 대기; 기본값 구현 유지. The operator verdict is pending, and the implemented default (measurement-failure class only) is kept. The decision-index cell is not edited inside the audit window. The file is operator-owned and was not touched by this sync.
4. Cause-8 residual: `postMergeHold` writes its hold and releases in two mutations. This is the defect class of REQ-MWQ2-007, but outside its named cause-7 outcome. Recorded, not fixed here. 범위 밖, 리더 이관 (out of scope for this card, transferred to the leader).
5. AC-MWQ2-005 built-binary observation was not performed (§E.3.2 item 1). The in-process substitutes recorded there stand. No built-binary result is claimed.
6. The installed moai build is not a valid evidence source for this card: it is an ancestor of HEAD. No moai CLI output is cited in this block or in the sync evidence.
7. MX: both mandatory anchors are now present as comment-only `@MX:ANCHOR` blocks: `ShortSHA` (introduced by this card) and `ReleaseIntegrationLock` (pre-existing). See E.4.2.
8. Carried by reference, unchanged: run-phase §E.3.2 items 3 (E4 literal grep matches in comments of unrelated cli files), 4 (repository-wide verdict pending on CI), 8 (`workflow.yaml` review-gate flip, provenance not established), and 9 (plan-audit known gap, AC-MWQ2-008 literal).
9. Sync-commit trailer: `git log -1 --format='%(trailers:key=Authored-By-Agent,valueonly)'` on commit 5d6c9b8c3 printed nothing. Observed cause: git's trailer parser does not recognize the final paragraph, because the closing 🗿 MoAI line is a non-trailer line in the same paragraph. Control: the same trailer alone parses under `git interpret-trailers --parse`. The ownership audit's WHO reader (`internal/spec/lint_ownership.go`, a line regex over the commit body) matches that line; this is a code reading, not a run of the audit. Not amended, per the sync instruction.

### E.4.4 Scope note for the sync auditor (not a gap)

- The test reversal `TestRemeasureMixedTestAndEmptyPackageRemainsInvalid` → `TestRemeasureMixedTestAndEmptyPackageIsValid` (commit `eac7fab39`, M2) must be checked against AC-MWQ2-003. The lane judges it intended. Current name, checked in this sync with `grep -rn 'TestRemeasureMixedTestAndEmptyPackage' internal/factory --include='*_test.go'` → `internal/factory/integration_remeasure_run_test.go:189:func TestRemeasureMixedTestAndEmptyPackageIsValid(t *testing.T) {`

## §J Lane run-entry record (card t1582, run tmnboq, lane-5)

- Lease: `factory next --card t1582` was refused once with `serial-slot` (holder t1568 live until 2026-10-09T16:15:03Z, observed 16:08:21Z). Leader ruling bb3b2d04 (all run tmnboq cards parallel; withdraws the serial order 44610f49) was verified on disk at 16:33Z: `moai factory status` shows `mode=parallelizable` for the run. The lease was then granted; the card is `picked` and the tree was entered at 16:33Z.
- Watchdog first observation (16:33:03Z, no prior snapshot, fail-open): HEAD `8673c2a95`, integration window `free`, evidence mtime max 1791549653.
- RED baseline on the pinned tree `8673c2a95` (`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -v -run '^TestRedT1582' ./internal/factory/ ./internal/cli/`; raw log kept in the lane scratchpad): FAIL ×4 with the stated reasons — `TestRedT1582MixedSweepWithNoTestPackageCounts`, `TestRedT1582VerifierAdmitsAShortBaseSHA`, `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease`, `TestRedT1582MergeReadyMeasurementFailureStillExitsZero`; PASS ×5 (control and sealed items). Committed before any implementation as `d7f4fcf0c`.
- Pre-spawn sync (orchestrator rule, lane-local): `git rev-list --count --left-right origin/main...HEAD` = `139 95` (diverged). The implementation spawn is HELD.
- Absorb probe (`git merge-tree --write-tree HEAD origin/main`, tree unchanged): exit 1. Conflicts in `.claude/rules/moai/workflow/context-window-management.md`, its template mirror, `internal/bugreport/spool.go`, and `internal/bugreport/spool_bump_serialize_test.go`. origin/main has 36 commits touching `internal/factory` or `internal/cli` that this branch lacks (the t1538 factory-recovery series): a cross-card overlap with this card's target packages.

decision record: decided_by=lane-5 (card-pick, run tmnboq) evidence_refs=card=t1582;class=C;mode=parallelizable@2026-10-09T16:33Z;prior_hold=t1568(lease expired 16:25:01Z);leader_ruling=bb3b2d04(supersedes 44610f49);pr=no-link;landed=none ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9.3)

wait record: id=w-t1582-20261009T1638Z waiting_on=leader reason=cross-card base decision — pinned base vs absorbing origin/main (4 conflicts + t1538 overlap); implementation spawn held recheck=one-shot 5 min (local 01:43 KST) plus standing cron 8a51a689 (:07/:27/:47)

decision record: decided_by=lane-5-watchdog evidence_refs=board:d-20261009T163954Z-0030;resolves:w-t1582-20261009T1638Z;baseline=8673c2a95;red=d7f4fcf0c;run_record=f9e272a7c;waiver=pre-spawn-139/95-impl-spawn-only;mode=parallelizable@20261009T1643Z ladder_path=step2-board-ruling-option-B

- Sync-phase entry (lane-5, run tmnboq): the card moved `run -> sync` at version 17, evidence `b6b17e5ea` (HEAD; it carries the §E.3 run-phase audit-ready signal). Before the transition the run phase was re-observed from disk: RED family `TestRedT1582` 11 PASS (exit 0); `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/factory/ ./internal/cli/` exit 0; `golangci-lint run ./internal/factory/... ./internal/cli/...` 0 issues; `go test -count=1 ./internal/factory/` ok 297.437s, coverage 83.9% of statements; `go test -race` factory family 100 PASS / 0 FAIL / 0 DATA RACE; cli family 68 PASS / 0 FAIL; `gofmt -l` lists only `internal/factory/remeasure_red_t1582_test.go`.
- Test rename, located from disk: `TestRemeasureMixedTestAndEmptyPackageRemainsInvalid` (base `8673c2a95`, `internal/factory/integration_remeasure_run_test.go:185`) became `TestRemeasureMixedTestAndEmptyPackageIsValid` (HEAD, line 189) in commit `eac7fab39` (M2). The assertion flipped from "remains invalid" to "is valid" — the intended AC-MWQ2-003 behavior change. §E.2 names `integration_remeasure_test.go`; the file is `integration_remeasure_run_test.go`.
- Lease recovery (expiry 18:07:48Z): the stage verb returned the card to assigned (version 14). `factory next` run from the worktree is refused ("runs from the parent checkout"), so the lease was taken through the MCP verb with the parent checkout as `project_root` (card t1582, run tmnboq; row `leased ... until 18:32:05Z`). The resume transition with the existing evidence `bb245d325` renewed it (run v16).
- Pre-spawn divergence for the sync-phase manager-docs spawn: `git fetch origin main` exit 0, then `git rev-list --count --left-right origin/main...HEAD` = `139 104` at 2026-10-09T18:21:24Z. The `origin/main`-only count is unchanged since the ruling (139); the HEAD-side growth (95 → 104) is this card's own commits. Board record d-20261009T163954Z-0030 waives the check for the implementation spawn only. The lane applies the same basis (pinned baseline; absorb deferred to immediately before landing) to the sync-phase spawn, because that spawn writes only to this unpushed card branch inside an isolated worktree and `moai session list --json --filter-spec=SPEC-MERGE-WINDOW-QUEUE-002` returned `[]`. The leader can stop the spawn before its commit.
- Open items carried into §E.4 as gaps, not fixed in the sync phase: AC-MWQ2-008 gofmt on the RED overlay (leader decision); coverage 83.9% below the 85% package threshold, baseline not measured at `8673c2a95`; decision-index Q1 operator verdict blank; cause-8 residual in `postMergeHold`.

decision record: decided_by=lane-5 (sync-entry, run tmnboq) evidence_refs=board:d-20261009T163954Z-0030;session-list=[];rev-list=139 104@2026-10-09T18:21:24Z;head=b6b17e5ea ladder_path=step1-disk-probes+step2-board-scope-read (lane judgment; not keep-set; not cross-card)

- Coverage baseline (lane-5): `go -C <export> test -count=1 -coverprofile=<export>/cover.out ./internal/factory/` on the pinned base `8673c2a95`, exported with `git archive` to `$HOME/.cache/moai-t1582-base` (outside /tmp), under slot `go-test-internal-factory` → `coverage: 83.3% of statements`, exit 1, package time 314.123s, with one failure: `TestMergeStepPreMergeCausesReleaseWithDistinctCodes` (`tree: exit status 128: fatal: not a git repository`, because the export carries no `.git`). Card side, same command: `ok 297.437s coverage: 83.9% of statements`, no failures, measured at HEAD `b6b17e5ea`. Since then only docs commits and the formatting of one `_test.go` file changed, so the production code measured is identical to `8f01ecd2f`. Attribution: the shortfall below the 85% package threshold predates this card, and the card did not lower coverage (+0.6 points). The base figure is approximate because of the one environment-caused failure. Gap wording per leader ruling (b): "기존 미달, 이 카드가 낮추지 않음".
- gofmt decision (leader reply 10276f43, option (i)): commit `8f01ecd2f` is a formatting-only change to `internal/factory/remeasure_red_t1582_test.go` (8 lines, `git diff -w --stat` empty). Option (ii) was rejected because it would edit acceptance.md and change the plan-artifact hash that the passed plan verdict pins; option (i) touches no SPEC artifact. RED re-run after the change: 10 top-level `--- PASS`, 10 subtests PASS, 0 FAIL, both packages ok. `gofmt -l internal/factory internal/cli` is empty, so AC-MWQ2-008's literal criterion holds.
- Correction: the earlier lane record and status report said "RED 11 PASS". The logs count 10 top-level PASS (plus 10 subtests). The 11 figure was a lane miscount; cite 10.
- Cause-8 residual (`postMergeHold`, two mutations): not fixed in this card. Transferred to the leader as a follow-up candidate (leader ruling item d); the §E.4 gap entry names it as out of scope.
- Decision-index Q1: the operator's answer is pending (the leader is asking the operator). The default implementation (measurement-failure class only; waiting stays exit 0) is kept. The Operator-verdict cell is edited outside the audit window by manager-spec (leader ruling item c), not by this lane.
- Audit scope for sync-auditor (leader ruling item e): the reversal of `TestRemeasureMixedTestAndEmptyPackageRemainsInvalid` (base `8673c2a95`, `internal/factory/integration_remeasure_run_test.go:185`) into `TestRemeasureMixedTestAndEmptyPackageIsValid` (HEAD line 189, commit `eac7fab39`) is to be checked against AC-MWQ2-003. The lane judges the reversal intended.
- MX follow-up (sync sub-step, comment-only): `ShortSHA` (card-introduced) and `ReleaseIntegrationLock` (pre-existing, in `integration_lock.go`, a file this card changed) get `@MX:ANCHOR` in the manager-docs follow-up commit, provided the non-test call-site count is at least three. The §E.4.2 "not added" entries are superseded by that commit.
