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
