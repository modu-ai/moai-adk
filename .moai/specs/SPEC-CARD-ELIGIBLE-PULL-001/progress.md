# SPEC-CARD-ELIGIBLE-PULL-001 — progress

card: t1839
status: draft (plan-phase artifacts authored 2026-10-17, rewritten 0.2.0 same day)
tree: 93f4fc75c (branch WT-10-17-c)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-17T16:00:31Z (2026-10-18 KST)
plan_artifacts: spec.md(7 REQ, v0.2.2) · plan.md(§A-§H, M1-M5, v0.2.2) · acceptance.md(AC-CEP-001..007, v0.2.2) · decision-index.md(Q1 판정 기록 포함)
authored_at: 2026-10-17 (0.2.2 correction: 2026-10-18)
author: manager-spec
closing_audit: rule-C delta audit PASS — 판정 파일 .moai/reports/t1839/plan-audit-delta.md, correction commit 376f7df21, receipt rcpt-d75ece55c14bd82d83f5ec47, plan_artifact_hash 4ac6eca4…f3c1, convergence ok(codex pass, 0 findings). D1-D9 전부 해결.
notes: 플랜 페이스 종결 — 플랜 페이스는 문서만 작성·갱신(카드 지시 — 코드·테스트·설정 무변경). Q1(플래그 OFF 의미)이 레인 결정 사다리로 해석 (ii)(킬스위치)로 해결됨 — decision-index.md 판정 행 참조(decided_by: lane-17, 리더 통보; 운영자 반환 리뷰 대기). 감사 사슬: 1차 FAIL(0.83, rcpt-8b1c74f7dbede113c4974ecf) → D1-D6 수리(0.2.1) → 2차 FAIL(0.85, rcpt-390fce8f7ecca3d52673bc8e) → Rule C 단일 수정 0.2.2(D7/D8/D9 — standing d-20261015T094733Z-9e76, t1793 선례) → 델타 감사 PASS. 다음 단계: 킥오프 후 /moai run.

## §E.2 Run-phase Evidence

**Base pin (plan §C.1, recomputed at run-phase entry)**: branch `WT-10-17-c`, HEAD `51b6ffc98` (2026-10-18). Delegation-time HEAD was `e43b3eea2`; the two commits between them (`e43b3eea2`, `51b6ffc98`) are the coordinator's kickoff-gate docs commits (§E.1 audit_ready token + plan-audit-iter2 convention copy) — code-untouched, outside this SPEC's file scope. Kickoff confirmed by coordinator message: factory gate approved `t1839: run (v11)`, lease restored to lane-17. Plan-audit skip-eligibility for the Phase-1 gate: verdict PASS (rule-C delta audit, receipt `rcpt-d75ece55c14bd82d83f5ec47e`), Tier M threshold met, plan_artifact_hash `4ac6eca4…f3c1` unchanged since the verdict — skip taken per the three-condition contract.

**Run-phase interpretation I-1 — SUPERSEDED by amendment 0.3.0 (commit `909a89d53`, F1 ruling)**: the `Priority == ClassPriorityHigh` surrogate for REQ-CEP-003 condition 2 was removed; the condition is now fail-closed and classification-agnostic (no spec id ⇒ always `spec-missing`). The original note is preserved below for the audit trail.

**Run-phase interpretation I-1 (superseded — original record)**: REQ-CEP-003 condition 2 names "유효 분류(EffectiveCardClassification)가 Class C인 카드" — the queue classification (`internal/factory/classification.go` `CardClassification`) carries priority/blocked/mode and NO dispatch class letter (A/B/C). Implemented surrogate: `Priority == ClassPriorityHigh` is treated as the heavyweight class that must carry a spec id (fail-safe direction; the pre-amendment AC-CEP-003's "Class C 아님 + 무spec ⇒ 통과" cell forced a discriminating predicate, and priority was the only mechanical axis). The sync-audit round-1 F1 finding (codex bidirectional divergence, receipt `rcpt-b7f73391…`) rejected the surrogate; the F1 ruling (leader board d-20261018T021007Z-2620, operator delegation + Jev 0.98) made the requirement fail-closed via in-place amendment 0.3.0.

### Pre-flight (plan §C.2/C.3)

- **P1 — M1-ENTRY observation (AC-CEP-003 regression-guard pin, pre-implementation tree `51b6ffc98` clean)** — command: `go test ./internal/cli/ -run '^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1 -v` — exit 0. Verbatim output (tail):
  ```
  --- PASS: TestSD_AC008_NextSelectionOrderAndOutput (10.00s)
      --- PASS: TestSD_AC008_NextSelectionOrderAndOutput/assigned-to-this-lane_wins_over_unowned_picked (2.30s)
      --- PASS: TestSD_AC008_NextSelectionOrderAndOutput/unowned_picked_wins_over_older_queued (2.04s)
      --- PASS: TestSD_AC008_NextSelectionOrderAndOutput/card_assigned_to_another_lane_is_never_taken;_exit_3 (1.37s)
      --- PASS: TestSD_AC008_NextSelectionOrderAndOutput/oldest_queued_promoted_and_leased;_newer_stays_queued (2.18s)
      --- PASS: TestSD_AC008_NextSelectionOrderAndOutput/empty_queue_prints_no-card_and_exits_3 (0.77s)
      --- PASS: TestSD_AC008_NextSelectionOrderAndOutput/two_lanes_concurrently:_exactly_one_leases_the_picked_card (1.34s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/cli	11.241s
  ```
  Pins the current GREEN of ungated (b) leasing (subtests 2 and 6 lease an unowned picked card through the generic arms) — observable only before the gate exists (plan §F M1 entry).
- **P2 — vet baseline**: `go vet ./internal/cli/ ./internal/config/` — exit 0, no output.
- **P3 — lint baseline**: `golangci-lint run ./internal/cli/... ./internal/config/...` — 3 issues, all pre-existing errcheck in `internal/cli/factory_run_adopt.go` (outside this SPEC's file scope): `3 issues: * errcheck: 3`. NEW-issue gate for E4 = 0 new findings relative to this baseline.
- **P4 — test-file list (plan §C.3)**: (a) update-target (pins ungated (b)/(b2) leasing): `internal/cli/factory_self_dispatch_test.go` — subtests `TestSD_AC008_NextSelectionOrderAndOutput/unowned_picked_wins_over_older_queued` and `.../two_lanes_concurrently:_exactly_one_leases_the_picked_card`. (b) untouched-arm tests: every other subtest of `TestSD_AC008_NextSelectionOrderAndOutput` (arms (a)/(c)), `TestSD_AC009_NextWaitLeasesOrTimesOut`, `factory_classify_test.go` serial/mutual-exclusivity family, `factory_serial_slot_stale_test.go` slot family — all unmodified (E5 proves by scoped diff).

### M1 — config key plumbing (RED-first) + AC-CEP-007 closed in-milestone

- **RED (key-absent tree)** — command: `go test ./internal/config/ -run '^TestWorkflowFactorySelfDispatchDefault$' -count=1 -v` — tree `51b6ffc98` + defaults_test.go only. Verbatim output:
  ```
  internal/config/defaults_test.go:705:32: NewDefaultWorkflowConfig().Factory undefined (type WorkflowConfig has no field or method Factory)
  FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
  FAIL
  ```
- **Implementation**: `internal/config/types.go` (`FactoryConfig` + `WorkflowConfig.Factory` yaml `factory`), `internal/config/defaults.go` (`Factory: FactoryConfig{SelfDispatch: false}`), repo + template workflow.yaml (`factory.self_dispatch: false` with comment), `internal/config/defaults_test.go` (`TestWorkflowFactorySelfDispatchDefault`, \*bool yaml probe so key-absence fails).
- **Cascade follow-up (golden-file regeneration, attributable to REQ-CEP-007)**: `internal/config/testdata/shipped_key_inventory.yaml` +1 row (`workflow.factory.self_dispatch`, class W) — `TestShippedConfigKeysHaveReaders` (REQ-CKH-008 anti-rot) failed without it; outside the literal plan §D file list but inside the scope envelope per the test-fixture cascade precedent.
- **GREEN** — command: `go test ./internal/config/ -run '^TestWorkflowFactorySelfDispatchDefault$' -count=1 -v` — exit 0:
  ```
  --- PASS: TestWorkflowFactorySelfDispatchDefault (0.01s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/config	0.383s
  ```
- **Package-wide**: `go test ./internal/config/ -count=1` — exit 0: `ok github.com/modu-ai/moai-adk/internal/config 4.375s`.

### M2 — RED landing (new gate scenarios + expectation flips), tree `6cf3bf85e` + test edits

All runs on the pre-gate tree (M3 not yet implemented); anchored selectors per plan §C.2 (a).

- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullOffDormancy$' -count=1 -v`** — FAIL (RED, 3/3 subtests):
  ```
      factory_card_eligible_pull_test.go:141: load t1: factory card not found
      factory_card_eligible_pull_test.go:164: t1 gained a record row although arm (b2) is dormant — the skip must be traceless
      factory_card_eligible_pull_test.go:180: gate refusal: expected an error, got nil
  ```
  Why red: today's ungated tree leases the unowned picked card (arms (b)/(b2) active) — the no-lease/traceless assertions fail.
- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullOnGate$' -count=1 -v`** — FAIL (RED, 5/5 cells):
  ```
      factory_card_eligible_pull_test.go:250: gate refusal: expected an error, got nil            (cells i–iii: today the candidate LEASES)
      factory_card_eligible_pull_test.go:266: refused predecessor-unmerged lines for t1: 0 ([]), want exactly 1
      factory_card_eligible_pull_test.go:284: refused serial-slot lines for t1: 0 ([]), want exactly 1
  ```
  Why red per cell: (i) hold-state row — today no lease but no reason line; (ii)/(iii) high-no-spec and picked+marker — today the candidate leases through the ungated arm; (iv)/(v) predecessor/slot — today the skip already happens but silently (no token line).
- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullOnEligible$' -count=1 -v`** — FAIL (RED at the second call; the eligible lease itself passes today too, as expected — the fixture's RED premise was pinned at M1-entry):
  ```
      factory_card_eligible_pull_test.go:321: gate refusal: expected an error, got nil
  ```
- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullTraceless$' -count=1 -v`** — FAIL (RED, 2/2):
  ```
      factory_card_eligible_pull_test.go:349: gate refusal: expected an error, got nil   (off_quiet_skip)
      factory_card_eligible_pull_test.go:349: gate refusal: expected an error, got nil   (on_gate_refusal)
  ```
- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullRefusalGrammar$' -count=1 -v`** — FAIL (RED):
  ```
      factory_card_eligible_pull_test.go:411: gate refusal: expected an error, got nil
  ```
  Why red: today's no-card response (`factory_card.go:1775-1776`) carries no reason lines, so the parser cell cannot find them.
- **Flipped pinning tests — `go test ./internal/cli/ -run '^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1 -v`** — FAIL (RED on exactly the two flipped subtests; the four untouched subtests stay PASS in the same run — preserved-contrast evidence):
  ```
      factory_self_dispatch_test.go:174: load t1: factory card not found
      factory_self_dispatch_test.go:270: t2 = leased holder="lane-1", want still picked and unleased — arm (b) is dormant
      --- PASS: .../assigned-to-this-lane_wins_over_unowned_picked
      --- PASS: .../card_assigned_to_another_lane_is_never_taken;_exit_3
      --- PASS: .../oldest_queued_promoted_and_leased;_newer_stays_queued
      --- PASS: .../empty_queue_prints_no-card_and_exits_3
  ```
- Flip enumeration (§C.3 (a), complete): the only existing tests pinning ungated (b)/(b2) bare-next leasing are the two `TestSD_AC008` subtests above — verified by sweeping every `factoryNextLeaseOnce`/`factoryNextLeaseOnceGated` test call site and every picked-queue fixture (classify/serial-slot/drift-log/duplicate-dispatch/bundle/adopt/assign/nominate families all lease through arm (a)/(c) or the nominated path, out of gate scope).

### M3 — flag branch, dormancy, five-condition predicate — tree `3ba9cda1f` + gate implementation

`internal/cli/factory_card.go`: `factoryEligibility` (flag + refusal collector), `disqualify` (conditions 1→5, fixed order, first disqualifier only), `factorySelfDispatchEnabled` (root-scoped read, fail-safe OFF), `factoryNextLeaseOnceEligibility` (gated wrapper keeps its signature; refusals ride the last pass out), arms (b)/(b2) behind `elig.active()`/`disqualify`. Arms (a)/(b-priority)/(c) untouched; serial-slot read reused read-only (D5); hold marker via state-independent `autoRankHoldMarked` (D1).

- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullOffDormancy$' -count=1 -v`** — exit 0, 3/3 PASS (AC-CEP-001 green path M2→M3 closed):
  ```
  --- PASS: TestFactoryEligiblePullOffDormancy (4.18s)
      --- PASS: .../unowned_picked_row_is_not_leased;_the_queued_sibling_is_promoted_instead
      --- PASS: .../queue-picked_card_with_no_row_is_neither_recorded_nor_leased
      --- PASS: .../lone_unowned_picked_card_answers_no-card_with_no_reason_line
  ```
- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullTraceless$' -count=1 -v`** — exit 0, 2/2 PASS (AC-CEP-005 green path M3 closed; the refusal LINE assertion lives in the gate/grammar cells — M4):
  ```
  --- PASS: TestFactoryEligiblePullTraceless (2.24s)
      --- PASS: .../off_quiet_skip
      --- PASS: .../on_gate_refusal
  ```
- **`go test ./internal/cli/ -run '^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1 -v`** — exit 0, 6/6 PASS (the two flipped subtests GREEN through the implementation; the four untouched subtests still GREEN — regression guard):
  ```
  --- PASS: TestSD_AC008_NextSelectionOrderAndOutput (10.13s)   [all six subtests PASS]
  ```
- `go build ./internal/cli/` + `go vet ./internal/cli/` — exit 0, no output.

### M4 — structured refusal surface — tree `ea1facf53` + surface rendering

`factoryNextNoCardText` (one shared renderer: first line kept, one `refused <token>: <detail>` per candidate); the verb's no-card path prints it to stdout (exit 3, msg unchanged); the factory_next MCP handler returns it as the tool text (the fixed `"no card is available"` literal at mcp_factory_card.go:187 replaced by the shared source — AC-CEP-006(5)). OFF skips still add no line.

- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullOnGate$' -count=1 -v`** — exit 0, 5/5 cells PASS (AC-CEP-002 green path M2→M3→M4 closed):
  ```
  --- PASS: TestFactoryEligiblePullOnGate (7.18s)
      --- PASS: .../i_queue_item_held:_refused_not-admitted
      --- PASS: .../ii_class-C_high_without_spec:_refused_spec-missing
      --- PASS: .../iii_picked_plus_hold_marker:_refused_hold-marker
      --- PASS: .../iv_unmerged_after-predecessor:_refused_predecessor-unmerged
      --- PASS: .../v_serial_slot_held:_refused_serial-slot
  ```
- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullOnEligible$' -count=1 -v`** — exit 0 (AC-CEP-003 green path M4 closed, both rule-paired assertions):
  ```
  --- PASS: TestFactoryEligiblePullOnEligible (2.97s)
  ```
- **`go test ./internal/cli/ -run '^TestFactoryEligiblePullRefusalGrammar$' -count=1 -v`** — exit 0 (AC-CEP-006: first line preserved, 3 grammar-matching lines naming t1/t2/t3, cobra block == MCP block):
  ```
  --- PASS: TestFactoryEligiblePullRefusalGrammar (1.31s)
  ```
- **Scope suite (E2 form)** — `go test ./internal/cli/ -run '^TestFactoryEligiblePullOffDormancy$|^TestFactoryEligiblePullOnGate$|^TestFactoryEligiblePullOnEligible$|^TestFactoryEligiblePullTraceless$|^TestFactoryEligiblePullRefusalGrammar$' -count=1` — exit 0: `ok github.com/modu-ai/moai-adk/internal/cli 16.402s`.

### M5 — regression-guard deployment + verification — tree `ceb392ed3` + fail-safe cell

- **E5 scoped diff** — `git diff e43b3eea2..HEAD --stat -- internal/cli/ internal/config/`: exactly the planned files; among EXISTING test files only `factory_self_dispatch_test.go` (±39 lines, the two flipped subtests) changed; every untouched-arm test file byte-identical:
  ```
   internal/cli/factory_card.go                       | 249 +++++++++--
   internal/cli/factory_card_eligible_pull_test.go    | 460 +++++++++++++++++++++
   internal/cli/factory_self_dispatch_test.go         |  39 +-
   internal/cli/mcp_factory_card.go                   |   7 +-
   internal/config/defaults.go                        |  11 +
   internal/config/defaults_test.go                   |  46 +++
   .../config/testdata/shipped_key_inventory.yaml     |  3 +
   internal/config/types.go                           |  18 +
   8 files changed, 780 insertions(+), 53 deletions(-)
  ```
- **E3 vet** — `go vet ./internal/cli/ ./internal/config/` — exit 0, no output.
- **E5 lint** — `golangci-lint run ./internal/cli/... ./internal/config/...` — 3 issues, identical to the pre-implementation baseline (errcheck ×3, all in `internal/cli/factory_run_adopt.go`, outside this SPEC's file scope): `3 issues: * errcheck: 3` → NEW issues: 0.
- **E4 subagent-boundary grep** — `grep -rn 'AskUserQuestion' <4 touched non-test files>` — 0 matches.
- **E6 OFF-dormancy + preserved re-run at final tree** — `go test ./internal/cli/ -run '^TestFactoryEligiblePullOffDormancy$|^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1` — exit 0: `ok github.com/modu-ai/moai-adk/internal/cli 14.277s` (AC-CEP-001 dormancy cells + all six AC008 subtests incl. the four preserved).
- **E3 coverage (new gate paths)** — `go test ./internal/cli/ -run '^TestFactoryEligiblePull.*$|^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1 -coverprofile=…` + `go tool cover -func`:
  ```
  factory_card.go:393:  factoryNextLeaseOnceEligibility  80.8%   (uncovered: infra error paths — record-open/pass/lock errors)
  factory_card.go:1303: factoryNextNoCardText           100.0%
  factory_card.go:1377: disqualify                      100.0%   (all five conditions + the pass path)
  factory_card.go:1424: factorySelfDispatchEnabled       75.0%   (uncovered: the defensive err!=nil branch; malformed-yaml → defaults → dormant proven behaviorally)
  ```
  New-path coverage ≈ 91% weighted (target 85%). Package-wide: `go test ./internal/config/ -cover -count=1` — `ok … coverage: 85.6% of statements` (≥85).
- **Fail-safe cell added**: `unreadable config fails safe to dormancy` (malformed workflow.yaml → loader warn-and-default → arms dormant; loader contract makes this the behavioral carrier of the kill-switch default).
- **Slot wait record (leader load rule)**: `moai slot acquire --resource go-test-heavy` → exit 3 at first attempt; holder lane-11 (`t1668 M8: go test -cover ./internal/cli/`, bound ends 2026-10-17T21:19:15Z). Windows cross-build and package-wide internal/cli run wait for the lease; retried below.
- **Load-rule deviation, recorded**: two incremental single-package `go build ./internal/cli/` compile checks (M3, M4) ran without a lease — the rule names `go build` among heavy runs. Incremental single-package builds measured seconds and machine load stayed nominal (test timings unchanged across the run), but the letter of the rule was missed; subsequent builds (incl. the windows cross-build) wait for the lease.

- **Adjacent-family regression evidence (anchored selectors, slot-exempt)**:
  - classify family: `go test ./internal/cli/ -run '^TestFactoryNextParallelizableConcurrentLeases$|^TestFactoryNextRecordAndClaimRaceOnLeasedRow$|^TestFactoryNextDuplicateDispatchGuard$' -count=1` — `ok … 5.143s`.
  - serial-slot family (11 tests): anchored multi-name selector — `ok … 39.557s`.
  - bundle family: first pass found ONE test using an ungated (b) lease as its vehicle —
    ```
    --- FAIL: TestFactoryNextAfterGuardCountsOtherRunsMerges
        factory_bundle_test.go:376: lane-1's lease = "", want t2 — a predecessor merged under run-prev still gated the after guard
    ```
    The test's SUBJECT is the after-guard's cross-run merged-set scope (card t1454 r2f5); its VEHICLE was an unowned-picked arm-(b) lease. Update (RED observed above, pre-fix): fixture arms `cepSelfDispatchYAML(t, root, true)` so the subject keeps its coverage — the gate's condition 4 now carries it (a run-prev merge read as unmerged would refuse `predecessor-unmerged` instead of leasing). GREEN: `go test ./internal/cli/ -run '^TestFactoryNextAfterGuardCountsOtherRunsMerges$' -count=1 -v` — `--- PASS (1.52s)`. This test was outside the §C.3 enumeration (its file greps as bundle-family, not lease-once); the anchored sweep caught it, the package-wide run below is the completeness check.
  - remaining bundle family (8 tests anchored): `ok … 22.158s`.
- **Slot lease acquired** (after the recorded wait): `moai slot acquire --resource go-test-heavy` — success; holder `lane-17-t1839`, bound 25m.
- **E2 cross-platform build (slot-gated)**: `GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/config/` — exit 0, no output.

- **M5 corrected close (leader order, binding — 2026-10-18)**: the package-wide `internal/cli` rerun is DROPPED (rule 35e9: internal/cli runs only the tests the card affects, via anchored `-run`; the rest belongs to batch CI; no-timeout-expansion). The two local package-wide attempts are recorded as EVENTS, not verdicts, and are cited as neither pass nor fail: attempt 1 hit the go-test default 600 s timeout (`FAIL … 600.860s`, tail-only output — no per-test detail); attempt 2 was ordered dropped by the leader and died with the session (its partial output is not evidence of anything). No further slot acquisition.
- **GAP — package-wide `internal/cli` suite: NOT run locally.** Rule 35e9: batch CI owns it; slot-leak incident 2026-10-18. The card's regression coverage is the narrowed anchored runs below plus the scoped diff.
- **Narrowed M5 verification (all anchored, slot-exempt, final tree `ef7e84fe4` + parity repair)**:
  - `go test ./internal/cli/ -run '^TestFactoryEligiblePull.*$' -count=1` — `ok github.com/modu-ai/moai-adk/internal/cli 19.295s` (whole new family incl. the fail-safe cell).
  - Flipped/updated + MCP-surface + codex-next batch: `go test ./internal/cli/ -run '^TestSD_AC008_NextSelectionOrderAndOutput$|^TestFactoryNextAfterGuardCountsOtherRunsMerges$|^TestSD_AC014_MCPMatchesCLIWithProjectRoot$|^TestSD_AC023_CodexNextSkipsUnadvanceableCard$' -count=1` — first run FAIL, one more (b2)-vehicle test found and repaired (below); after repair — `ok github.com/modu-ai/moai-adk/internal/cli 39.036s`.
  - `TestSD_AC014_MCPMatchesCLIWithProjectRoot/factory_next_leases_identically` — RED observed verbatim pre-fix: `factory_m3_test.go:327: cli next: factory next: no card is available` (fixture leased a queue-picked no-row card through arm (b2); under the OFF default both surfaces answer no-card). Repair: fixture arms `cepSelfDispatchYAML(t, root, true)` — the parity SUBJECT (cli and mcp lease identically) keeps its coverage behind the gate. GREEN: `go test ./internal/cli/ -run '^TestSD_AC014_MCPMatchesCLIWithProjectRoot$' -count=1 -v` — 6/6 subtests PASS (`ok … 18.846s`).
  - `go test ./internal/config/ -run '^TestWorkflowFactorySelfDispatchDefault$' -count=1 -v` — `--- PASS (0.00s)`, `ok … 0.327s`.
  - `go test ./internal/config/ -cover -count=1` — `ok … coverage: 85.6% of statements` (≥85).
  - `go vet ./internal/cli/ ./internal/config/` — exit 0.
  - `golangci-lint run ./internal/cli/... ./internal/config/...` — `3 issues: * errcheck: 3` — identical to the pre-implementation baseline (all in `factory_run_adopt.go`, outside scope): NEW issues 0.
  - E2 cross-platform (delegation-sanctioned, slot-gated, recorded above): `GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/config/` — exit 0.
- **Updated existing tests, final ledger** (each with RED observed pre-fix): `TestSD_AC008` subtests ×2 (expectation flips), `TestFactoryNextAfterGuardCountsOtherRunsMerges` (flag armed), `TestSD_AC014_MCPMatchesCLIWithProjectRoot/factory_next_leases_identically` (flag armed). Every other existing test file byte-identical to base `e43b3eea2`.

### Sync-audit round-1 F2 repair + behavior-based re-sweep — 2026-10-18 (tree `5cc25a21d` → F2 commit)

- **F2 RED (auditor's observation, cited verbatim from `.moai/reports/t1839/sync-audit.md` — not re-derived)**: `go test ./internal/cli/ -run '^TestFactoryLeaseSectionRecordWritesPerArm$' -count=1 -v` → `--- FAIL: TestFactoryLeaseSectionRecordWritesPerArm (23.71s)` / `--- FAIL: .../arm-b` / `--- FAIL: .../arm-b2` (`factory_lease_section_test.go:201: next: factory next: no card is available`). The lease-atomicity table's two generic-arm rows lease unowned picked cards via `qasRunNext`/`nmBase` — a verb-driven vehicle the run phase's call-site-and-name sweep could not see.
- **F2 repair**: `flArmCase` gains `selfDispatch`; rows `arm-b` (picked row) and `arm-b2` (rowless picked item) arm `cepSelfDispatchYAML(t, root, true)` — the atomicity subject (record writes under the lease-section lock) keeps its coverage behind the gate. GREEN: `go test ./internal/cli/ -run '^TestFactoryLeaseSectionRecordWritesPerArm$' -count=1 -v` — 8/8 subtests PASS (`ok … 28.071s`).
- **F3 process correction (behavior-based sweep method)**: the §C.3/M2 enumeration axis (lease-once call sites + picked-fixture names) is falsified by F2 — it cannot see verb-driven vehicles. Corrected method: enumerate EVERY `next`-verb invocation in `internal/cli/*_test.go` (`qasRunNext` / `runFactory(t, "next"…)`), intersect with fixtures whose leased target is an unowned picked row or a rowless picked item, and inspect each hit's enclosing test. Sweep result: exactly one further at-risk vehicle outside the already-updated set —
- **Sweep finding RED (own observation, pre-fix)**: `go test ./internal/cli/ -run '^TestReviewFindingMergedPRPredecessor$' -count=1 -v` →
  ```
      review_observation_test.go:44: predecessor=merged-pr; factory next leased=""
      review_observation_test.go:46: merged-pr predecessor still blocks t2
  --- FAIL: TestReviewFindingMergedPRPredecessor (2.22s)
  ```
  Bare `fbLeasedCard` + t2 unowned picked row = arm-(b) vehicle. Repair: fixture arms the flag — the subject (merged-pr predecessor must not block the successor) now rides the gate's condition 4. GREEN: `--- PASS: TestReviewFindingMergedPRPredecessor (4.40s)` (`ok … 5.981s`).
- **All other sweep hits cleared by inspection** (nominated `--card` or arm-(a)/(c) vehicles, out of gate scope): quota t1516 wait shape (assigned row), lease-operator held-assigned (no-card expected), m4/launcher codex-env (launcher must NOT lease; child leases nominated), lease-bound stall rows (nominated + bare arm (c) queued), lane-join-guard (`--card` nominated), review sibling tests (`--card`), nominate family (`--card`), t1533/quota-worktrees/carry/claim-deadline/harness (no bare-next picked vehicles).
- **F2 repair verification (anchored, one selector)**: `go test ./internal/cli/ -run '^TestFactoryEligiblePull.*$|^TestFactoryLeaseSectionRecordWritesPerArm$|^TestReviewFindingMergedPRPredecessor$|^TestSD_AC008_NextSelectionOrderAndOutput$|^TestFactoryNextAfterGuardCountsOtherRunsMerges$|^TestSD_AC014_MCPMatchesCLIWithProjectRoot$' -count=1` — `ok github.com/modu-ai/moai-adk/internal/cli 168.104s`.

### Amendment 0.3.0 run — F1 repair: fail-closed condition 2 — 2026-10-18 (tree `abcddd597` → F1 commit)

Amendment `909a89d53` (F1 ruling: leader board d-20261018T021007Z-2620, operator delegation + Jev 0.98; sync-audit F1 codex divergence receipt `rcpt-b7f73391…`): REQ-CEP-003 condition 2 is fail-closed — a card with no spec id (row or queue hint) is NEVER leased through arms (b)/(b2), always refused `spec-missing`, classification/priority-agnostic. The priority==high surrogate is removed; interpretation I-1 marked SUPERSEDED above.

- **RED (amended AC-CEP-003 (a) cell on the surrogate tree, tree `abcddd597` + test edits)** — `go test ./internal/cli/ -run '^TestFactoryEligiblePullOnGate$|^TestFactoryEligiblePullOnEligible$' -count=1 -v`:
  ```
  === RUN   TestFactoryEligiblePullOnGate/ii-b_normal_priority_without_spec:_refused_spec-missing_(fail-closed,_amendment_0.3.0)
      factory_card_eligible_pull_test.go:296: gate refusal: expected an error, got nil
  --- FAIL: TestFactoryEligiblePullOnGate (18.56s)
  ```
  Why red: the surrogate leased the normal-priority no-spec candidate — exactly the audit's divergence case. (OnEligible with the spec-carrying eligible sibling passed on the surrogate tree too, as expected.)
- **Implementation**: `disqualify` condition 2 loses the `cls.Priority == ClassPriorityHigh` conjunct and the `cls` parameter; detail reworded to `card <id> carries no spec id on the row or the queue hint`. Doc comment cites amendment 0.3.0.
- **Amended tests**: OnGate gains cell `ii-b` (normal no-spec ⇒ refused, the F1 repair-verification cell); OnEligible's eligible sibling t1 carries `SPEC-ELIGIBLE-001` (amended AC-CEP-003 (c)); single-defect cells iii/iv/v and the grammar fixture's marker candidate carry spec ids so each cell's targeted condition is the FIRST disqualifier under fail-closed ordering.
- **Vehicle REDs under fail-closed (observed verbatim pre-fix)** — the four flag-armed vehicles lease spec-less unowned picked cards, so condition 2 now refuses them first:
  ```
  --- FAIL: TestFactoryNextAfterGuardCountsOtherRunsMerges — factory_bundle_test.go:383: lane-1's lease = "", want t2
  --- FAIL: TestFactoryLeaseSectionRecordWritesPerArm — factory_lease_section_test.go:210: next: factory next: no card is available  [arm-b, arm-b2 rows]
  --- FAIL: TestSD_AC014_MCPMatchesCLIWithProjectRoot — factory_m3_test.go:332: cli next: factory next: no card is available
  --- FAIL: TestReviewFindingMergedPRPredecessor — review_observation_test.go:50: predecessor=merged-pr; factory next leased=""
  ```
  Repairs: each vehicle's leased candidate carries a spec id (`SPEC-FAL-ARM-001` / `SPEC-AFTERGUARD-001` / `SPEC-PARITY-001` / `SPEC-REVIEW-001`) via `cepSetSpecID` — subjects unchanged.
- **GREEN — full anchored verification (one selector)**: `go test ./internal/cli/ -run '^TestFactoryEligiblePull.*$|^TestFactoryLeaseSectionRecordWritesPerArm$|^TestFactoryNextAfterGuardCountsOtherRunsMerges$|^TestSD_AC014_MCPMatchesCLIWithProjectRoot$|^TestReviewFindingMergedPRPredecessor$|^TestSD_AC008_NextSelectionOrderAndOutput$' -count=1` — `ok github.com/modu-ai/moai-adk/internal/cli 179.895s` (OnGate 6 cells incl. ii-b; OnEligible spec-carrying sibling leases, spec-less sibling refused; OffDormancy/Traceless/Grammar unchanged-green; all four vehicles green; flipped AC008 green).
- **Narrowed companions**: `go test ./internal/config/ -run '^TestWorkflowFactorySelfDispatchDefault$' -count=1` — `ok … 0.371s`; `go vet ./internal/cli/ ./internal/config/` — exit 0; `golangci-lint run ./internal/cli/... ./internal/config/...` — `3 issues: * errcheck: 3` (= baseline, 0 NEW).
- **GAP (unchanged)**: package-wide `internal/cli` suite not run locally (rule 35e9 — batch CI owns it).

### Round-3 repair — F6 + F9 + the third (exhaustive) sweep — 2026-10-18 (tree `e5302ce47` → round-3 commit)

**F6 REDs observed verbatim at HEAD** (matching the verdict's probes; each fix applied only after its RED):
```
factory_m3_test.go:105: next: factory next: no card is available                                   [AC011 materializer cell, arm-(b2)]
factory_m3_test.go:163: next over a foreign directory: err = factory next: no card is available, want a refusal naming …/​.moai/worktrees/t1   [AC011 foreign-dir cell, arm-(b)]
factory_t1533_test.go:239: lane-1 next: factory next: no card is available — a merged-pr predecessor must release its follower
factory_two_hub_cycle_test.go:40: normal lease="", want t3; combined hub/after order must make progress
factory_nominate_test.go:1214: bare `factory next` transcript "arm-order" changed:  [got next#2 head="t1" via arm (c), want "t4" via arm (b)]
```
**F6 repairs**: AC011 both cells, MergedPRFollower, TwoHub normal cell — flag-armed + `cepSetSpecID` the vehicles, subjects preserved. The **arm-order golden** (`NextBareUnchanged`) — guard-both-ways per the coordinator's judgment call, reasoning recorded: arm order (a)→(b)→(b2)→(c) is orthogonal to admission (the gate filters candidates inside the generic arms, it does not reorder them), so the pinned EXPECTATION is unchanged under the armed flag (verbatim golden preserved, spec ids on the (b)/(b2) vehicles); the OFF default gets a NEW transcript block `arm-order-off-dormancy` in `nmBareGolden` — that is the §D expectation flip, whose RED is the observed old-expectation failure above.

**F9 (vacuous green) — first-hand mutant reproduction**: falsifying (b2)'s `hubWaitUnmerged` result (one-line overlay) at HEAD → `TestReviewUnrecordedPickedHubWait` still `ok` (vacuous — the OFF `active()` skip precedes the guard); source restored. Fix: the fixture is armed + `t2` carries a spec id so the hub-wait subject executes behind the gate; the unarmed control `TestReviewUnrecordedPickedHubWaitOffDormancy` (same fixture, no flag) pins the quiet-dormancy outcome and asserts zero refusal lines. **Mutant recheck post-fix**: the same overlay now FAILS the armed test (`factory_t1533_test.go:515: … predecessor card not merged: t1 has not reached merged-local …`) — the guard demonstrably fires; source restored again.

**Third sweep — exhaustive enumeration (the F3 completion; cites the list, not a method)**:
- Enumeration commands: `grep -rn 'qasRunNext(t\|runFactory(t, "next"' internal/cli/*_test.go` → 92 direct sites in 24 files → enclosing-`func Test` extraction → **72 tests**; plus the helper axis the earlier sweeps missed: `fbLeasedCard` (factory_bundle_test.go:55, wraps the verb) callers in bundle/t1533/two_hub/review → +13; `flRaceLanes` callers → +4. **Union: 90 tests** (`/private/tmp/cep-union85.txt`, 90 lines).
- Empirical disposition: the whole union run under anchored `-run` selectors in three batches (batch 1 hit the 10 m go-test timeout mid-tail and was split; batches 2a/2b completed).
- **Per-test disposition (90 = 11 repaired + 2 F7-recorded + 77 cleared)**:
  - **Repaired this round (11, each RED-observed pre-fix, GREEN post-fix)**: `TestSD_AC011_CardWorktreeCreateReuseRefuse` (2 cells), `TestFactoryNextMergedPRPredecessorReleasesFollower`, `TestReviewTwoHubCombinedCycle/normal`, `TestFactoryNextBareUnchanged` (golden guard-both-ways), `TestReviewUnrecordedPickedHubWait` + new `…OffDormancy` control (F9), `TestQAS_AC008_ClaudeLaneHeldAtThreshold` (both lease cells — `factory_quota_test.go:378/380: t2 = picked holder="", want leased/lane-1`), `TestQWR_AC008_SurfacesSeeWorktreeReading/leases_when_no_record_anywhere` (`:387` same shape), `TestQWR_AC010b_CliRootWithoutGitMetadata/unregistered_directory_is_not_read` (`:559` same shape), `TestReviewNominatedCreationFollowsTheAfterRelation` (`t1533:871` lease=""), `TestReviewWaitFollowsTheAfterRelation` (both cells, `:953/:986`), `TestReviewFindingNoRecordArmSkipsBlockedCandidate` (`:788` lease="" — its READY card t3 is itself a (b2) candidate, so dormancy emptied the pass; repair note: my first edit placed `cepSetSpecID` before `fcQueue` and the ids landed on nothing — ordering corrected, then GREEN with t3 leasing past the skipped blocked t2).
  - **F7 timing tests (2, recorded NOT repaired — outside this round's envelope, per the verdict's `[Low] [optional] [flake-suspect, unattributed] … fix at next touch`)**: `TestFactoryLeaseQueueLockStallBounded` fails ISOLATED at HEAD (nominated cell 8.914 s vs the 8.75 s = B+500 ms bound; bare 8.737 s) and `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` clause (vii) (bare 548 ms vs the 500 ms post-release window). Unattributed to the card: both files last touched at t1454/t1506/t1458 (pre-card), the nominated cells bypass the gate entirely, and the error SHAPES are correct (exit 4 `refused raced` naming the lock; drift reconciled with 1 event). Own observation extends the auditor's: these are not load flakes — they fail isolated; the margins are simply tighter than this machine's overhead.
  - **Cleared (77)**: passed as-is at the final tree. Reason groups, every file named: (i) nominated-`--card` vehicles (gate out of scope) — the whole `factory_nominate_test.go` nominate family (13 tests: NominateLeasesNominee, PromoteThenLose, ClaimRefusedRollsBack, CompensationFailure, UnknownCard, RefusesKeepSet, SameCardExactlyOne, RecordStateTokens, BlankCardIsAnError, NominateMCPParity, BackendSkip, QuotaHold, CarriesWorktreeBinding), `TestFactoryNominatedBusyRefusalCarriesTheClaimError`, `TestFactoryNominateRefusesForeignBundleMember`, review nominated family (BundlePreservesHint, WaitsOnAllHubPredecessors, LeaseDoesNotFillAnEmptyHint, CreationFollowsTheAfterRelation's nominated half), lease-bound `--card` cells; (ii) arm-(a)/(c) or quota-hold or refusal-outcome vehicles (no unowned-picked lease) — ArmA family (3), ArmCSkipsHoldMarker, AllMarkerQueue, LeaseArmAExcludesHeldAssignedCard, LeaseArmCOperatorHold, LeaseCompensationKeepsOperatorPick, AC008's hold cells, QAS hold/latch family (5), quota_worktrees hold/status cells, ClaimDeadlineRefusal, SerialMutualExclusivity, SkipsClassificationBlocked (arm-(c) subject), LeaseSectionAllowedSet/ExcludesWorktreeStep/RejectsNestedMutate, SerialCrossProcessHelper + flRaceLanes serial family (3), HubChainPickedRowReleasesSerialSlot, carry family (CarriesWorktreeBinding, CreatesFreshTree, RefusesAnotherCardsPreviousTree, RefusesWhenPreviousBindingNamesAnotherPath), PostLeaseWorktreeFailure, QueuedAssignedCardPreservesHubHints, QWR surfaces/MCP-parity/render family (3), SD lifecycle family (AC006/009/010/015/018/023), MergedPR-family arm-(c) cells of PickedHubSkip/BundleMultiHubMemberWaits; (iii) repaired in EARLIER rounds and still GREEN at this tree — `TestSD_AC008_NextSelectionOrderAndOutput` (flipped pair + 4 preserved), `TestFactoryNextAfterGuardCountsOtherRunsMerges`, `TestSD_AC014_MCPMatchesCLIWithProjectRoot`, `TestReviewFindingMergedPRPredecessor`, `TestFactoryLeaseSectionRecordWritesPerArm`, `TestFactoryEligiblePullOffDormancy`/`OnEligible` (this SPEC's own family).
- **Consolidation verification (all 11 repairs, anchored, final tree)**: batch A `^TestSD_AC011…$|^TestFactoryNextMergedPRPredecessorReleasesFollower$|^TestReviewTwoHubCombinedCycle$|^TestFactoryNextBareUnchanged$|^TestReviewUnrecordedPickedHubWait$|^TestReviewUnrecordedPickedHubWaitOffDormancy$|^TestFactoryLeaseSectionRecordWritesPerArm$` → 7/7 PASS `ok … 108.106s`; batch B `^TestQAS_AC008…$|^TestQWR_AC008…$|^TestQWR_AC010b…$|^TestReviewNominatedCreationFollowsTheAfterRelation$|^TestReviewWaitFollowsTheAfterRelation$|^TestReviewFindingNoRecordArmSkipsBlockedCandidate$` → 6/6 PASS (batch-B run after the ordering fix; the blocked-candidate cell prints `lane lease: "t3"`).

### Round-4 repair — F10 (quota clobber), F11 (OFF-golden blindness), F12 (sweep record) — 2026-10-18 (tree `1a6c90fc3a` → round-4 commits `52272df8d`/`f5a4f1fd4`/F12 commit)

**F10 — the arming helper killed the quota gate (RED-first, first-hand)**:
- RED (the auditor's threshold-bump mutant, applied at HEAD `1a6c90fc3a` pre-fix): lease-cell usage 89.9/94.9 → 99 in `factory_quota_test.go:386/:388` → `--- PASS: TestQAS_AC008_ClaudeLaneHeldAtThreshold (26.42s)` with BOTH lease cells passing — impossible with a live gate (99% must hold). Vacuity proven: my helper's overwrite had disabled `quota_gate.enabled: true`. Bump restored (0 diff).
- Fix: `cepSelfDispatchYAML` is now a generic yaml map round-trip — it parses any existing workflow.yaml, sets ONLY `workflow.factory.self_dispatch`, and writes back; every key the fixture wrote (`quota_gate.enabled` included) survives. Absent file → the single-key body as before. Orphaned `cepBoolWord` removed.
- Mutant recheck post-fix: the same bump → `--- FAIL: …/five_hour_89_9_leases` + `--- FAIL: …/seven_day_94_9_leases`, each with the verbatim hold line (`factory_quota_test.go:386: next: factory next: quota hold: five_hour used=99.0% resets_at=…`) — the gate demonstrably fires through the merged yaml. Bump restored (0 diff).
- Re-verified green: `go test ./internal/cli/ -run '^TestQAS_AC008_ClaudeLaneHeldAtThreshold$|^TestQWR_AC008_SurfacesSeeWorktreeReading$|^TestQWR_AC010b_CliRootWithoutGitMetadata$' -count=1 -v` → 3/3 families PASS (`ok … 47.693s`) — all four F10-named cells (AC008 ×2 lease cells, QWR AC008 `leases_when_no_record_anywhere`, QWR AC010b `unregistered_directory_is_not_read`) plus their sibling cells.
- Sweep of my own diff for other armed quota fixtures: the complete `cepSelfDispatchYAML` call-site list is AC011 ×2, MergedPRFollower, TwoHub, arm-order goldens ×2, hub-wait pair, after-guard, MCP parity, review merged-pr, blocked-candidate, quota `leasedRun`, QWR ×2 — only the last four touch quota fixtures; all four re-verified above. No other fixture writes config keys the helper could clobber.

**F11 — the OFF golden was mutant-blind (RED-first via the auditor's mutant)**:
- RED (the auditor's active()-removal mutant, reproduced at HEAD `1a6c90fc3a` pre-fix): both `if !elig.active() { continue }` blocks removed (`factory_card.go:1002/:1028`) → `--- PASS: TestFactoryNextBareUnchanged (35.63s)` — the golden could not see the dormancy violation (its OFF candidates were un-spec-idd, and `nmStep` discarded the `refused` lines the violation expressed as). Source restored.
- Fix, both layers: the OFF golden's candidates t3/t4 now carry the SAME spec ids as the armed golden (a dormancy regression changes the transcript itself — the picked pair leases at next#2/#3), and the capture asserts refusal-line ABSENCE before normalization (`cepRefusedLines(out)` empty at every step, `factory_nominate_test.go` arm-order-off-dormancy).
- Mutant recheck post-fix: the same both-branches-removed mutant → `--- FAIL: TestFactoryNextBareUnchanged/arm-order-off-dormancy` (`factory_nominate_test.go:1273: bare 'factory next' transcript "arm-order-off-dormancy" changed:`) alongside `--- FAIL: TestFactoryEligiblePullOffDormancy` (2 cells, refusal-lines + gate-refusal errors). The golden's pin is no longer blind. Source restored (0 diff).
- Baseline green: `go test ./internal/cli/ -run '^TestFactoryNextBareUnchanged$|^TestFactoryEligiblePullOffDormancy$' -count=1` → `ok … 52.455s`.

**F12 — the sweep record regenerated, in-tree, reproducible**:
- Corrected arithmetic (every number now reproduces; commands verbatim in `sweep-union.txt`): raw `grep -rn 'qasRunNext(t\|runFactory(t, "next"' internal/cli/*_test.go` → **135 sites / 24 files** (the round-3 record's "92" was the same grep with an unstated inline `--card` filter — the filter is now explicit); enclosing-test extraction with per-file reset → **81 direct**; `fbLeasedCard` callers → **17**; `flRaceLanes` callers → **4**; union → **98**. All five numbers re-derived this round and matching the auditor's independent enumeration exactly.
- The union list lives IN THE TREE: `.moai/specs/SPEC-CARD-ELIGIBLE-PULL-001/sweep-union.txt` (header documents the generating commands and the beyond-union axes; 98 `^Test` entries verified by count). The `/private/tmp` file is superseded and no longer cited.
- **The 14 previously unrecorded dispositions**:
  - 9 union tests, all nominated `--card` vehicles (gate out of scope — vehicle verified per test by grep: all 7 `factory_lane_join_guard_test.go` sites are `--card`; launcher `:207` is `--card`; both review tests' bare lease is the nominated path), now GREEN-observed (anchored run of all nine: `ok github.com/modu-ai/moai-adk/internal/cli 33.262s`, 9/9 PASS): `TestFactoryNextLeasesBehindMergedPRPredecessor`, `TestFactoryNextIgnoresDriverlessPickedRow`, `TestFactoryNextForeignNominationStillHoldsBehindAssignments`, `TestFactoryNextOwnAssignedCardIgnoresSiblingAssignments`, `TestFactoryNextStillRefusedBehindADrivenRow`, `TestFactoryNextRefusesAgainAfterLaterAssign`, `TestFactoryCardVerbsResolveLaneFromWorkerMarker`, `TestReviewFindingNominatedLeasesAfterPredecessorMerges`, `TestReviewFindingNominatedOverwritesDependency`.
  - 5 beyond-union verb drivers, callers already dispositioned: `cepNoCard` (wraps runFactory "next"; callers = this SPEC's own gate family `TestFactoryEligiblePullOnGate` / `…RefusalGrammar` / `…Traceless` — repaired rounds 2-3, GREEN), and the `nmRunCase` / `nmRaceAtSeamTolerant` nominated drivers (callers `TestFactoryNextNominateBackendSkip` / `TestFactoryNextNominateQuotaHold` — inside the union, cleared as nominated in round 3).
- Disposition table status: 98 union entries = 13 repaired (11 round-3 + 0 new vehicles this round; F10/F11 repaired the GUARDS, not vehicles) + 2 F7 (recorded, not repaired — defensible per the round-3 verdict) + 83 cleared (the round-3 77 + the 9 above − 3 counted in both) — per-entry reasons in the round-3 groups + this addendum; every entry in `sweep-union.txt` is covered.

**Standing Gap re-ruled acceptable per the round-3 verdict**: package-wide `internal/cli` remains unrun locally (rule 35e9 — batch CI owns it); its acceptability precondition (F10+F11+F12 landed) is now met.

### Round-5 repair — F13 (template internal-content leak) — 2026-10-18 (tree `414aec43f` → commit `7847834791`)

- **RED (anchored, at HEAD `414aec43f` pre-fix)**: `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$' -count=1 -v` —
  ```
  internal_content_leak_test.go:1643: template internal-content leak detected (2 occurrences, mode=narrow):
  internal_content_leak_test.go:1653:   [1] templates/.moai/config/sections/workflow.yaml | class=C1-spec-id-prefix | match=SPEC-CARD-ELIGIBLE-PULL-001
  internal_content_leak_test.go:1653:   [2] templates/.moai/config/sections/workflow.yaml | class=C9-card-id | match=t1240
  ```
- **Fix**: the template `workflow.yaml` factory comment block de-identified — the flag key and its user-facing behavior only (no SPEC id, no card id, no internal paths; internal arm vocabulary genericized). The repo-side `.moai/config/sections/workflow.yaml` keeps its identifiers (only the template is leak-scanned).
- **GREEN**: same anchored command → `--- PASS: TestTemplateNoInternalContentLeak (2.39s)`, `ok github.com/modu-ai/moai-adk/internal/template 2.905s`.
- **Neighbors (anchored)**: `TestFactoryDispatchBodyCarriesRenewIntervalLinkage` → PASS (`ok … 0.371s`); config shipped-key pair `^TestWorkflowFactorySelfDispatchDefault$|^TestShippedConfigKeysHaveReaders$` → `ok github.com/modu-ai/moai-adk/internal/config 16.484s`. No template test asserts the self_dispatch block text (grep: 0 hits).
- **Lesson recorded**: the t1808 lesson recurs — template-tree edits require the `internal/template` package tests; three audit rounds (including round-3's sweep of my own diff) missed this file because the sweep axes covered `internal/cli`/`internal/config` only. The template yaml was IN the plan §D file scope from M1; the leak gate simply was never run against it until round 4.

### Round-6 repair — F14 (AC010 fixture), the re-sweep of the axes round 3-4 could not see, F15/F16 comparison — 2026-10-19 (tree `954e25f96` → round-6 commit)

**F14 — `TestSD_AC010_MCPNextParentCheck` (RED-first, first-hand)**
- Decision basis: REQ-CEP-001 (default OFF: arms (b)/(b2) lease nothing for an unowned card, a silent skip), D1 interpretation (ii) the kill switch, plan §D standing policy (a test whose vehicle is an ungated generic-arm lease is updated RED-first). The behavior is specified; the code stays; the fixture is repaired the way the F6 repairs armed their vehicles.
- RED (HEAD `954e25f96`, tree clean, load average 29.55 at the run): `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^TestSD_AC010_MCPNextParentCheck$' -count=1` →
  ```
  --- FAIL: TestSD_AC010_MCPNextParentCheck (2.09s)
      factory_m3_test.go:631: t1 = picked, want leased
  FAIL	github.com/modu-ai/moai-adk/internal/cli	3.624s
  ```
- Fix (`internal/cli/factory_m3_test.go`): `cepSelfDispatchYAML(t, root, true)` + `cepSetSpecID(t, store, "t1", "SPEC-AC010-MCP-001")` (set after `fcQueue`, the ordering the round-3 note records). No assertion changed; the linked-worktree refusal assertions are kept.
- GREEN: same command with `-v` → `--- PASS: TestSD_AC010_MCPNextParentCheck (2.56s)` / `ok  	github.com/modu-ai/moai-adk/internal/cli	3.696s`.
- Mutation check (scratch overlay, source tree untouched — `git status --short` showed only the test file): the MCP handler's parent-checkout assertion broken to run on the factory root instead of the caller's tree (`factoryAssertParentCheckout(factoryRootOf(launch))`, the shape the handler's own comment names) → the repaired test fails: `factory_m3_test.go:626: factory_next on the linked worktree: err = <nil>, want a refusal naming …/002`. The parent-check subject is still detected.

**Why rounds 3-4 missed it.** The sweep greps (`qasRunNext(t` / `runFactory(t, "next"` / `fbLeasedCard(` / `flRaceLanes(`) matched only the cobra-verb call shape. The lease entry `factoryNextLeaseOnceEligibility` has two more test-side forms: the MCP handler (`sdCallTool(t, handleFactoryNext, …)`) and direct Go calls (`factoryNextLeaseOnce` / `factoryNextLeaseOnceGated`). AC010 uses the first. The union now carries both axes (`sweep-union.txt` header, steps 6-7).

**Round-6 re-sweep (F14)**
- Method: `sweep-union.txt` regenerated from the commands in its header — 98 (verb axis, unchanged) + 22 new (20 from the MCP/direct-call grep, 2 from variable-arg verb calls) + 3 of this SPEC's own armed tests = **123 entries**; `go test -list '<123 anchored names>' ./internal/cli | grep -c '^Test'` → `123` (every name resolves to exactly one test; the swept set is not empty or short). Completeness argument and its residual are in the file header.
- Classification by observation, not by reading: a behavior-neutral probe (scratch `go test -overlay`; `factory_card.go` copy with a one-line append-to-scratch-file call at the `!elig.active()` skips of arms (b)/(b2) and at the gate evaluation, tagged `DORMANT-b` / `DORMANT-b2` / `GATE-b` / `GATE-b2`, with the enclosing `Test` function read from the stack; the card tree carries none of it) ran over the 123 names in six batches, every batch `ok`:
  `ok … 137.806s`, `ok … 89.311s`, `ok … 133.525s`, `ok … 129.765s`, `ok … 116.629s`, `ok … 81.115s` (the two timing tests excluded from the probe batches; see F15/F16). A test that logged `DORMANT-b*` has an unowned queue-picked candidate and no flag; one that logged `GATE-b*` is armed; one that logged neither never reaches arm (b)/(b2) with a candidate.
- Result (probe log, de-duplicated):
  - **Armed (flag ON + spec id; `GATE-*` observed)** — 23: `TestFactoryEligiblePullOnEligible`, `…OnGate`, `…RefusalGrammar`, `…Traceless` (ON half), `TestFactoryLeaseSectionRecordWritesPerArm`, `TestFactoryNextAfterGuardCountsOtherRunsMerges`, `TestFactoryNextMergedPRPredecessorReleasesFollower`, `TestFactoryNextBareUnchanged` (armed golden + the OFF golden), `TestQAS_AC008_ClaudeLaneHeldAtThreshold`, `TestQWR_AC008_SurfacesSeeWorktreeReading`, `TestQWR_AC010b_CliRootWithoutGitMetadata`, `TestReviewFindingMergedPRPredecessor`, `TestReviewFindingNoRecordArmSkipsBlockedCandidate`, `TestReviewNominatedCreationFollowsTheAfterRelation`, `TestReviewTwoHubCombinedCycle`, `TestReviewUnrecordedPickedHubWait`, `TestReviewWaitFollowsTheAfterRelation`, `TestSD_AC011_CardWorktreeCreateReuseRefuse`, `TestSD_AC014_MCPMatchesCLIWithProjectRoot`, plus the four repaired THIS round (below): `TestSD_AC010_MCPNextParentCheck`, `TestFactoryBundleKeepsSerialSlot`, `TestFactoryNextBundleSerialLane`, `TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms`.
  - **Dormancy pinned by design (OFF default; `DORMANT-*` observed, expectation is the quiet skip)** — 3 + the OFF halves above: `TestFactoryEligiblePullOffDormancy`, `TestReviewUnrecordedPickedHubWaitOffDormancy`, `TestSD_AC008_NextSelectionOrderAndOutput` (flipped pair + preserved cells).
  - **Unarmed dormant-hit, subject unaffected (recorded, not repaired)** — 1: `TestQAS_AC022_GateOffOrPressureOffOutputUnchanged` (`DORMANT-b` ×2, `DORMANT-b2` ×4 from the quota fixture's t2/t3; its subject is output equality gate-disabled vs gate-enabled-pressure-off, its control `off.exit != 0 || off.stdout == ""` still observes a lease — now through arm (c) — so the equality is still measured).
  - **Not a generic-arm lease vehicle (assignment / own-bundle / promotion / nominated `--card` / quota hold / refusal-before-selection; probe logged no candidate-bearing `DORMANT-b*`/`GATE-b*`)** — the remaining 96, including the five whose probe tag was `DORMANT-b-notqueuepicked` (a record row at `picked` whose queue item is not `picked` — condition 1 refuses it in old and new code alike, never a lease vehicle: `TestFactoryNextQueuedAssignedCardPreservesHubHints`, `TestFactoryNextSkipsHubCandidateWhosePredecessorIsUnmerged`, `TestFactoryNextPickedOwnerlessRowReleasesSlot`, `TestReviewFindingBundleMultiHubMemberWaits`, `TestReviewPickedHubSkip`) and `TestFactoryLeaseQueueLockStallBounded` / `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` (timing tests, F15/F16 below).
  - **No other red.** F14 was the only fixture whose lease vehicle was ungated and asserted a lease. Goroutine-launched selections would have logged an unnamed `?` tag; none appeared.
- **Three vacuous greens found by the probe and repaired in this commit (same family as F9)** — each asserts a NEGATIVE ("not leased / no card") that held only because arm (b)/(b2) was dormant, so the subject (the bundle skip, the serial slot as it binds the picked arms) never ran. Repair = arm the flag + spec ids on the vehicles; assertions untouched.
  - `TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms` (`factory_serial_slot_stale_test.go`, t2 spec id), `TestFactoryNextBundleSerialLane` and `TestFactoryBundleKeepsSerialSlot` (`factory_bundle_test.go`, t3/t4 spec ids; scenario C of BundleSerialLane leases through arm (b-priority), which the flag never touches).
  - Mutant evidence (scratch overlay on `factory_card.go`; the HEAD versions of the two test files extracted with `git archive` for the "before" side): mutant M-deps (arm (b)'s `disqualify` call passed `true` for depsMet instead of `!selectionSkips(c)`): HEAD versions of BundleSerialLane and BundleKeepsSerialSlot → `PASS` (vacuous); repaired versions → `FAIL` (`factory_bundle_test.go:136: lane lane-1 next: factory next: predecessor card not merged: t2 …`, `:196: lane lane-2 next: … t3 …`). Mutant M-slot (`modeEligible` always true): HEAD version of the serial-slot test → `PASS`; repaired → `FAIL` on both cells (`factory_serial_slot_stale_test.go:272: lane-2 took t2 through a picked arm while t1 is assigned and holds the serial slot`). A post-repair probe run showed the same candidate evaluations now tagged `GATE-b`/`GATE-b2` (2, 5/5, 12) where they were `DORMANT-*` before.
- Real-tree pass over the 123 names (final tree, no overlay), per-batch results: see "Results of the real-tree pass" below.

**Results of the real-tree pass (final tree = HEAD `954e25f96` + this round's test edits; `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli -run '^(<batch names>)$' -count=1 …`; batches partition `sweep-union.txt` = 20+20+20+20+20+21 + the two timing tests of F15/F16 = 123)**

| Batch | Names | Flags | Load at launch | Result |
|---|---|---|---|---|
| 1 | TestFactoryAssignBundleHubChain … TestFactoryLeaseSerialBareLanesExactlyOne (20) | `-v` | 30.08 | `ok … 96.781s` |
| 2 | TestFactoryLeaseSerialCrossProcessHelper … TestFactoryNextNominateCarriesWorktreeBindingAcrossRunReplacement (20) | `-v` | 24.59 | `ok … 83.137s` (the helper-process test prints `--- SKIP … helper process only`, its driver `TestFactoryLeaseSerialDistinctNomineesExactlyOne` passes) |
| 3 | TestFactoryNextNominateClaimRefusedRollsBack … TestFactoryNextStillRefusedBehindADrivenRow (20) | `-v` | 26.78 | `ok … 131.064s` |
| 4 | TestFactoryNominatedBusyRefusalCarriesTheClaimError … TestReviewNominatedLeaseDoesNotFillAnEmptyHint (20) | `-v` | 21.96 | `ok … 174.585s` |
| 5 | TestReviewNominatedWaitsOnAllHubPredecessors … TestQAS_AC008b_MCPFactoryNextHeld (20) | `-json` | 31.18 | `ok … 157.270s`, exit 0, `"Action":"fail"` count 0, top-level pass events 20 |
| 6 | TestMCPVerbParity_LaneLabel … TestFactoryEligiblePullRefusalGrammar (21) | `-json` | 31.53 | `ok … 124.277s`, exit 0, `"Action":"fail"` count 0, top-level pass events 21 |

Limits of this record: batches 1-4 ran as `-v` text before the lane rule to run each scope once with `-json` arrived, so their raw output is the task capture, not JSONL (each ends in the verbatim `ok` line above, exit 0, no `--- FAIL`); batches 5-6 are raw JSONL under the session scratch task directory, which is machine-local and not a citation target (the lines above carry the deciding evidence). The `-list` fix (123 names ↔ 123 entries) was taken before batches 5-6. No batch was re-run to obtain more output.

**F15 / F16 — card tree vs main under the same conditions (A/B, recorded loads)**
- Setup: `git archive --format=tar -o <scratch>/main-03faddb705.tar main` + `tar -x -C <scratch>/mainsrc` (main = `03faddb705`, an ancestor of the card HEAD, outside the repository, deleted at the end); card tree = HEAD `954e25f96` plus this round's test edits; each test run alone, `-count=1`, `-json` (`go test -json ./internal/cli -run '^<name>$' -count=1` on the card tree, `go -C <scratch>/mainsrc test …` on main), alternating card / main, each run launched only after a recorded load average below 32 (the admission rule; several launches waited for it).
- **F15 `TestFactoryLeaseQueueLockStallBounded`** (bound: B + 500 ms = 8.25 s + 0.5 s = 8.75 s on each of the two cells):

  | Run | Tree | 1-min load before launch | Package wall time | nominated cell | bare cell | Result |
  |---|---|---|---|---|---|---|
  | 1 | card | 30.77 | 21.423s | 8.427s | 8.660s | PASS |
  | 1 | main | 29.79 | 21.075s | 8.500s | 8.601s | PASS |
  | 2 | card | 27.86 | 23.368s | 8.769s | 8.681s | **FAIL** (nominated 19 ms over: `factory_lease_bound_test.go:334: the nominated lease returned after 8.769148209s, want within 8.75s`) |
  | 2 | main | 28.92 | 21.807s | 8.644s | 8.546s | PASS |
  | 3 | card | 31.48 | 19.685s | 8.501s | 8.449s | PASS |
  | 3 | main | 27.41 | 19.787s | 8.468s | 8.495s | PASS |
  | 4 | card | 21.32 | 19.737s | 8.540s | 8.460s | PASS |
  | 4 | main | 23.69 | 20.962s | 8.476s | 8.793s | **FAIL** (bare 43 ms over: `factory_lease_bound_test.go:357: the bare lease returned after 8.792984208s, want within 8.75s`) |

  Verdict: the first three pairs (the brief's 3 runs each) had the card tree fail once and main never; the failing cell was the NOMINATED one, which the card's diff does not touch (`git diff main HEAD -- internal/cli/factory_card.go` changes only the bare selection arms, the eligibility form of the lease entry, the verb's no-card rendering and the new gate helpers; `factoryNextNominate` and the lock wait are unchanged). Because 1 of 3 left the question open, a fourth pair was run (the lane rule allows a re-run as a failure A/B): there MAIN failed (bare cell, 8.793 s) and the card passed. 8 runs, 1 fail each side, both fails 19-43 ms past a bound whose total headroom is 500 ms over a fixed 8.25 s wait; observed overheads over the 8.25 s budget span 177-519 ms (card) and 218-543 ms (main). That is a load flake in the test's timing margin and not a card defect: it fails on main under the same conditions. Not repaired (a timing bound is not widened). The F7 note of round 3 (fails isolated on the card tree, never compared with main) is superseded by this comparison.
- **F16 `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits`** (bound: the verb returns within 500 ms after the lock release; figures are the two cells, nominated / bare):

  | Run | Tree | 1-min load before launch | Package wall time | nominated / bare after release | Result |
  |---|---|---|---|---|---|
  | 1 | card | 25.85 | 7.853s | 201 ms / 181 ms | PASS |
  | 1 | main | 28.92 | 8.615s | 202 ms / 207 ms | PASS |
  | 2 | card | 30.78 | 7.631s | 222 ms / 106 ms | PASS |
  | 2 | main | 29.63 | 7.765s | 179 ms / 110 ms | PASS |
  | 3 | card | 26.51 | 7.907s | 174 ms / 168 ms | PASS |
  | 3 | main | 23.55 | 7.870s | 201 ms / 231 ms | PASS |

  Verdict: 6/6 PASS, distributions overlap (card 106-222 ms, main 110-231 ms against a 500 ms bound). The round-6 audit's single failure (509/543 ms) did not reproduce on either tree at loads 23-31; the earlier 548 ms (round 3) and 509/543 ms (round 6 audit, load about 33) point to load sensitivity, consistent with no card involvement. Recorded, not repaired.

**Cross-check scope (the groups the round-6 audit could not finish; each its own invocation, `-json`, run once; lists fixed with `go test -list '<pattern>' ./internal/cli | grep -c '^Test'` first; no timeout, no shard)**

| Group | Pattern | `-list` count | 1-min load at launch | Result |
|---|---|---|---|---|
| G1 | `TestSD_` | 32 | 27.88 | `ok … 95.208s`, exit 0, 0 `fail` events, 32 top-level pass events |
| G2 | `EligiblePull\|SelfDispatch` | 6 | 29.72 | `ok … 20.771s`, exit 0, 0 `fail`, 6 pass |
| G3 | `TestFactoryCard\|TestFactoryNext\|TestFactoryBundle\|TestFactoryNominate` | 70 | 27.77 | `ok … 256.335s`, exit 0, 0 `fail`, 70 pass |
| G4 | `TestFactoryQuota\|TestFactoryLease\|TestFactoryM3\|TestFactoryT1533\|TestFactoryTwoHub\|TestReviewObservation\|TestMCPFactory` | 20 | 23.88 | `ok … 107.134s`, exit 0, 0 `fail`, 19 pass + 1 skip (`TestFactoryLeaseSerialCrossProcessHelper`, `helper process only`) |

`go build ./...` → exit 0, no output. `go vet ./internal/cli ./internal/config` → exit 0, no output. `gofmt -l internal/cli internal/config` → no output. The whole-package `internal/cli` suite was NOT run (the landing re-measure owns it under the slot).

**Round-6 Gaps / Residual-risk (what this record does not establish)**
- The completeness of the 123-name union rests on the grep axes in `sweep-union.txt` plus the probe; a test reaching `factoryNextLeaseOnceEligibility` through a helper whose call shape matches none of those greps would be missed by both (the probe only classifies tests that are run). The production-caller argument (three entry forms) bounds that risk but is not a mechanical proof.
- `TestQAS_AC022_GateOffOrPressureOffOutputUnchanged` is an unarmed dormant-hit test whose subject is unaffected; recorded, not repaired.
- The F15 timing bound is load-sensitive on BOTH trees (1 failure each in 8 paired runs); this record neither widens it nor repairs it. The F16 bound did not fail in 6 runs.
- Batches 1-4 of the sweep and the six probe batches are `-v` text, not JSONL; only batches 5-6, F15/F16 and G1-G4 are raw JSONL (machine-local scratch).
- The probe overlay and the mutant overlays are scratch instrumentation outside the repository; they are not reproducible from the tree and their results are cited, not committed.
- The package-wide `internal/cli` suite remains unrun locally; CI on the integration branch owns that verdict and it is PENDING at report time.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-18T01:55:00Z (2026-10-18 10:55 KST)
run_commit_sha: 3adc37875 (backfilled — the commit that carried the completed §E.2/§E.3; per the D3 placeholder pattern a commit cannot cite its own hash)
run_status: complete (M1-M5 per plan §F order; corrected M5 close per leader order 2026-10-18 — narrowed anchored verification, package-wide suite recorded as GAP)
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 3 (arm (a)+assigned-rows semantics — AC008/AC023 preserved cells GREEN; arm (b-priority) bundle-ownership semantics — bundle family anchored GREEN; arm (c) promotion + serial-slot read-only — serial-slot family anchored GREEN, slot semantics untouched per D5)
l44_pre_commit_fetch: done — `git fetch origin main` exit 0; divergence `10 99` (origin 10 / branch 99 — expected git-flow shape; local main ahead of origin, card branch stacks on it)
l44_post_push_fetch: n/a — no push (factory card; the leader batch-lands; §E.2 records the no-push rule)
new_warnings_or_lints_introduced: 0 (final `golangci-lint run ./internal/cli/... ./internal/config/...` = `3 issues: * errcheck: 3`, byte-identical to the pre-implementation baseline, all in factory_run_adopt.go outside this SPEC's file scope)
cross_platform_build.windows_amd64: pass (GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/config/ — exit 0, slot-gated)
total_run_phase_files: 10 changed (6 in-scope source/config/yaml + 2 in-scope test files authored + 2 existing test files updated under the RED-first update policy) + 1 cascade fixture (shipped_key_inventory.yaml) + 2 SPEC-side evidence files (spec.md frontmatter status/updated only; progress.md §E.2/§E.3)
m1_to_mN_commit_strategy: per-milestone Conventional commits M1..M5 (`feat`/`test` prefixes, `(card t1839)` in every subject, Authored-By-Agent: manager-develop + 🗿 MoAI trailers); M1 carries the draft → in-progress transition; no push, no PR
gaps: package-wide `internal/cli` suite NOT run locally (rule 35e9 — batch CI owns it); CI on the integration branch owns the repository-wide test verdict and that verdict is PENDING at report time; interpretation I-1 (priority=high as the Class-C surrogate for REQ-CEP-003 condition 2) awaits operator/auditor review; two incremental single-package `go build ./internal/cli/` compile checks ran lease-less during M3/M4 (recorded deviation, §E.2 M5)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
sync_complete_at: 2026-10-18T01:48:00Z (2026-10-18 10:48 KST)
sync_commit_sha: 438a3c96e (backfilled — the commit that carried the §E.4 signal, the CHANGELOG entry, and the status close; per the D3 placeholder pattern a commit cannot cite its own hash)
ac_count_at_sync: 7 (live AC-CEP-001..007 counted from acceptance.md SSOT by the B12 counter: live=7 excluded=0 ambiguous=0)
b12_self_test_a_duplicate_grep: pass — `grep -c 'SPEC-CARD-ELIGIBLE-PULL-001' CHANGELOG.md` → 0 (exit 1, no prior entry; emission proceeded)
b12_self_test_b_ac_count: pass — B12 counter stdout 7 == CHANGELOG-cited acceptance-criteria count 7; source acceptance.md (tier M resolver), not progress.md
b12_self_test_c_path_check: pass — every referenced path verified by ls before commit (internal/cli/factory_card.go, internal/cli/mcp_factory_card.go, internal/config/types.go, internal/config/defaults.go, .moai/config/sections/workflow.yaml, internal/template/templates/.moai/config/sections/workflow.yaml)
changelog_entry_position: [Unreleased] > ### Added, top entry
frontmatter_status_transitions.spec_md: in-progress → implemented → completed, both transitions merged into this single sync commit (3-phase close; status only — updated already 2026-10-18, no byte change)
frontmatter_status_transitions.plan_acceptance: untouched — updated already 2026-10-18 on both artifacts; no status field exists (Artifact Statelessness)
mx_tag_check: read-only pass — @MX:NOTE + @MX:SPEC present on factoryEligibility (factory_card.go:1352-1353); the sync phase added no code, so no tag edits (card instruction: Go code/tests untouched in sync)
rollout_note: t1240-shaped self-dispatch deployments must set workflow.factory.self_dispatch: true to preserve lane bare-next unowned-card leasing
known_deviation_for_audit: I-1 — REQ-CEP-003 condition 2 implemented via the priority==high surrogate (recorded §E.2; left as recorded for the sync auditor to judge, not editorialized)
notes: one sync commit carries the CHANGELOG entry + this §E.4 signal + the spec.md status close; body content of spec.md/plan.md/acceptance.md untouched; no push (factory card branch — the leader batch-lands per the lane protocol)

## §G Override and Refusal Record

- 2026-10-17T16:07:12Z SPEC-CARD-ELIGIBLE-PULL-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit.md
- 2026-10-17T16:07:34Z SPEC-CARD-ELIGIBLE-PULL-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit.md
- 2026-10-17T16:10:02Z SPEC-CARD-ELIGIBLE-PULL-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit.md
