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

**Run-phase interpretation I-1 (recorded for operator/auditor review, not a SPEC edit)**: REQ-CEP-003 condition 2 names "유효 분류(EffectiveCardClassification)가 Class C인 카드" — the queue classification (`internal/factory/classification.go` `CardClassification`) carries priority/blocked/mode and NO dispatch class letter (A/B/C). Implemented surrogate: `Priority == ClassPriorityHigh` is treated as the heavyweight class that must carry a spec id (fail-safe direction; AC-CEP-003's "Class C 아님 + 무spec ⇒ 통과" cell forces a discriminating predicate, and priority is the only mechanical axis). Flagged for the sync audit; reversible in one predicate line.

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

_<M4-M5 evidence appended below as milestones land>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 2026-10-17T16:07:12Z SPEC-CARD-ELIGIBLE-PULL-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit.md
- 2026-10-17T16:07:34Z SPEC-CARD-ELIGIBLE-PULL-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit.md
- 2026-10-17T16:10:02Z SPEC-CARD-ELIGIBLE-PULL-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1839/.moai/reports/t1839/plan-audit.md
