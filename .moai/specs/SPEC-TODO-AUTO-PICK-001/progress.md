# SPEC-TODO-AUTO-PICK-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02T15:12:45Z   # iteration-3 correction (UTC; spec.md dates it 2026-10-03 local); iteration-2 repair 2026-10-02T14:51:20Z; iteration-1 repair 2026-10-02T12:22:07Z; first plan completion 2026-10-02T11:55:40Z
card: t1448
tier: M
plan_head: 625f01718   # iteration-3 correction base HEAD (the correction commit follows it); iteration-2 repair base 63daaf6a7; iteration-1 repair base b3646de10; first plan authored at 4bf547bca
plan_audit: FAIL iteration 3 (0.80, MP-1..9 PASS, one blocker F1); F1 and notes corrected in this commit WITHOUT a re-audit — leader decision pending   # iteration 1: FAIL 0.73 (b3646de10); iteration 2: FAIL 0.77 (63daaf6a7); iteration 3: FAIL 0.80 (625f01718; .moai/reports/t1448/plan-audit-iter3.md, not committed); Tier M threshold 0.80; the plan-artifact hashes the iteration-3 verdict recorded are invalidated by this correction
plan_audit_iter4: PASS 0.85 (delta, operator-commissioned; audited hash 724841520)   # Clarity 0.80, Completeness 0.88, Testability 0.82, Traceability 0.88; Tier M threshold 0.80, margin +0.045; auditor-model claude-sonnet-5-5[1m]; verdict file .moai/reports/t1448/plan-audit-iter4-delta.md (card tree only, not committed); the iteration-3 line above is history and stays as written
artifact_set: spec.md, plan.md, acceptance.md, research.md, spec-compact.md, progress.md
requirements: 16       # Tier M ceiling 16
acceptance_criteria: 14  # Tier M ceiling 16
needs_clarification: 0
```

Plan-phase notes (what a reader of this record needs, nothing populated for later phases):

- Baselines the criteria measure against were observed in this tree before any run commit and land
  in the plan commits (`verification-claim-integrity.md` §2.3): `acceptance.md` ledger rows
  L1-L47, C1-C5, S1-S2, context rows G1-G8.
- **Iteration-3 correction (F1, F2, S1-S8) — not audited.** The audit (`625f01718`) returned FAIL
  0.80 on one blocker: `internal/template/catalog.yaml` stores hashes for the `moai` and
  `moai-kanban-foreman` skill directories and the template `manager-todo.md`, and the plan did not
  regenerate them. The plan now does (plan §5, M5, M6, AC-TAU-010/-012/-013, DoD 7), and the red was
  **observed** on a reversible perturbation (ledger G7: exit 1, `CATALOG_HASH_UNSTABLE` ×3,
  `CATALOG_HASH_SKINNY` ×2; restored, `git status --short` empty, `cmp` against backups empty). The
  correction changes the plan artifacts, so the iteration-3 hashes no longer describe them and a
  verdict on the corrected plan does not exist; this record does not claim a PASS. REQ 16, AC 14.
- Iteration-2 repair: N1..N7 and the leader's additional requirements are resolved in
  `plan.md` §10 (Audit resolution map). The two deliberate deltas the leader may veto without ripple
  remain isolated: D-DEF (second clause of REQ-TAU-007) and D-LANE (REQ-TAU-008).
- **AC baseline snapshot guard — observed first, as the leader required.**
  `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1`
  on `63daaf6a7`, **before** this repair edited `acceptance.md`: exit 0, `ok  …/internal/spec  6.230s`
  (**already green** — the two earlier SPEC commits had not moved a recorded count). After the
  repair, with the anchored selector and `-v`: exit 0,
  `--- PASS: TestACCounterFullCorpusMatchesBaseline (6.84s)`; the test lists this SPEC's
  `acceptance.md` as `absent-from-snapshot … COUNT 14` (reported, not failed — a SPEC not in the
  snapshot is not a baseline movement). **No regeneration was needed**, so
  `.moai/reports/t338/ac-count-baseline.txt` is untouched and not staged.
- The compensating-control claim for the decision record stays retracted: no party executes a re-read
  of the line today (spec §B.5, §G); auto-semantics §9.1/§10 wording stands for the other gates.
- Unverified items are listed in `research.md` R10 and R12.
- The throwaway probe file `internal/cli/zz_t1448_probe_test.go` was created and deleted during
  research; it never entered a commit. Its four observations are `research.md` R2 (O1-O4). The
  scratch binary and queue of rows S1/S2 live outside the tree; nothing of them is committed.

### Iteration-4 delta plan-audit result (appended; iterations 1-3 above are unchanged)

- Iteration 4 is an operator-commissioned delta audit, outside the Tier M ceiling of 2 audits,
  commissioned after the iteration-3 FAIL (0.80). Verdict **PASS**, overall score **0.85**
  (Clarity 0.80, Completeness 0.88, Testability 0.82, Traceability 0.88); margin +0.045 over the
  Tier M threshold 0.80 (iteration 3 was 0.80). auditor-model `claude-sonnet-5-5[1m]`. Audited
  commit `724841520` (full `72484152041c8e894947fb356835678b4d9b0023`). Verdict file:
  `.moai/reports/t1448/plan-audit-iter4-delta.md` (local, card tree only).
- Scope: closure of F1 (catalog.yaml row, generator step, two guard tests, AC-TAU-010/-012/-013)
  plus new contradictions from the F2 and S1..S8 edits. F1 RESOLVED; no MUST-FIX open; MP-1..MP-9
  PASS or N/A. No audit receipts (no cross-model tool invoked).
- Optional notes D1..D5: D1 serial-slot re-adoption is order-conditional; D2 two more English
  user-facing doc hits (`docs-site/content/en/advanced/factory-mode.md:73`,
  `docs-site/content/en/cli-reference/launchers.md:35`) fall in sync-phase scope; D3 the M6
  selector literal is not printed in M6 itself but is in M5 step 5 and AC-TAU-010; D4 G7
  abbreviates sha256 values; D5 this progress record was stale until this append.
- Per-file sha256 at `724841520` (the audited hashes; re-computed at the same HEAD and equal):

  ```text
  spec.md        3d68cb79b5f2e09f8e98223f35d2d2a8e283670b263ac6c2715d272bb744d9ce
  plan.md        d251637af3f4b29f76b873dbc1abccb621888975c98e5ea1ab04b58d6d5931f9
  acceptance.md  e7242951b8b0fe7bdd835194aca41bed31400f702f0796d6beaaf2ea6d8377db
  research.md    6dc5361d386e9318a72e7402358b3013213f6164fabe69afd57334085356a26a
  spec-compact.md b0a1b4f78eda903e3320b6e587b303fe5947ca8993e516b28c50aff1d9182e76
  progress.md (at 724841520, before this append)
                 76d81e719d7c6ed79c97752f3dcb43d98f00395d429516ff925f4dbb6dd92a80
  ```

  `progress.md` is not a member of the plan-artifact hash set
  `{acceptance.md, design.md, plan.md, research.md, spec.md, tasks.md}`
  (`.claude/rules/moai/workflow/spec-workflow.md` § Report Persistence), so this append does not
  invalidate the verdict.
- **Post-audit re-pin (develop absorption) — mechanical, no re-audit run, leader to be told.**
  After the iteration-4 verdict (audited hash `724841520`) develop `7109e0900` was absorbed (merge
  `095ac6c3e`) and changed files the SPEC cites. Re-pinned, at tree `b03619b29`: AC-TAU-011 now measures
  the always-loaded `kanban-dispatch.md` relative to the blob at `git merge-base develop HEAD`
  (re-derived at reading time, never pinned) instead of the fixed 26,959 / 26,637 B and 26,754 / 26,433
  chars that develop's own +1,349 B per copy made unsatisfiable; cells L14 and L21 re-based to the
  merge base; L11, L22, C3 re-measured (28,308 / 27,986 B, 28,099 / 27,778 chars, `differ: char 25120,
  line 181`); line numbers moved by develop's insertions given with their text (the live-only
  sentence 177 → 181; `auto-semantics.md` L169 → L199, L186 → L216); new ledger rows L48-L52;
  unchanged rows tabulated with their exit codes. **Semantics unchanged:** REQ 16, AC 14, no criterion
  changed meaning, `status:` untouched, `spec.md` `version` 0.4.1. Plan artifacts differ from the
  audited hash only by this mechanical re-pin: `spec.md`, `plan.md`, `acceptance.md`,
  `spec-compact.md` (plus `spec.md`'s run-phase `status:` transition, recorded in §E.2); resolution
  map: `plan.md` § 10 "Post-audit re-pin". No re-audit was run; the leader decides whether the
  re-pinned artifacts need one before M5.

### Plan->run Kickoff record (autonomous form, `auto-semantics.md` §9.1)

The Kickoff conditions are met: independent plan-audit verdict PASS (iteration 4, by plan-auditor);
`plan_status: audit-ready` recorded above; plan-artifact hashes unchanged since that verdict
(sha256 values above re-computed at `724841520`); no open blocker (no MUST-FIX, no MP violation).
Run Phase 1 skip contract (`spec-workflow.md` § Phase Transitions) holds on all three conditions:
verdict PASS; score 0.85 >= Tier M threshold 0.80; artifact hash unchanged.

```text
decision record: decided_by=operator-decision(2026-10-03, AskUserQuestion answer relayed by the factory leader: 'correct, then re-audit only F1') + lane-4 orchestrator evidence_refs=.moai/reports/t1448/plan-audit-iter4-delta.md;.moai/reports/t1448/plan-audit-iter3.md;audited-hash=724841520 ladder_path=gate-row plan->run Kickoff (AUTONOMOUS, auto-semantics §9.1; operator decision 'correct then re-audit F1 only', audited hash 724841520)
```

Forward notes for the run phase:

- develop (`7109e0900` at the time of writing) already carries t1451's edits to `kanban-dispatch.md`
  (card-review stage + standing recheck cron paragraphs), `kanban-dispatch-detail.md`,
  `auto-semantics.md` and `catalog.yaml`. The SPEC's pinned baselines (RED-now cells at `4bf547bca`,
  `kanban-dispatch.md` byte baselines 26959 live / 26637 mirror) must be RE-MEASURED after the develop
  absorption, and the new values recorded in §E.2 by the run phase.
- The five old-authority literals L5/L6/L7/L8/L25 were re-checked on develop and are all still present
  in live and mirror, so the SPEC's replacement targets exist.
- develop `kanban-dispatch.md` is 28308 B live / 27986 B mirror; their 322 B difference (the
  pre-existing line-177 drift) is unchanged.
- A later card (t1399, lane-3) renames `kanban-dispatch*.md` -> `factory-dispatch*.md`, the foreman
  skill and `internal/kanban` -> `internal/factory`. If it lands first, the run/sync phase re-greps the
  old paths and runs that card's rename script (`.moai/reports/t1399/rename/`) before re-measuring.

## §E.2 Run-phase Evidence

Run-phase milestones M1-M4 (card t1448, cycle_type=tdd). M5 and M6 are a separate later delegation.

**Ownership-exception record.** The run-phase for M1-M4 was executed by a general-purpose carrier of the
manager-develop role; reason: the typed `manager-develop` spawn auto-isolated into its own agent worktree
(a blocker report whose pre-flight showed toplevel `.claude/worktrees/agent-a1738e79aac1cb93a`, branch
`worktree-agent-a1738e79aac1cb93a`, HEAD `7109e0900`) and returned without writing; no ownership boundary
was crossed — no manager-spec-owned body (`spec.md` body, `plan.md`, `acceptance.md`, `research.md`) was
edited. The one `spec.md` change is the sanctioned frontmatter transition `status: draft` ->
`status: in-progress` (`updated:` already read 2026-10-03). The authored-by trailer on the run commits names
the role performing the transition, not a typed agent.

### Pre-flight and baseline re-measured after develop absorption

All measured in this run, in the card tree `.moai/worktrees/t1448`, branch `WT-todo-auto-pick-autonomy`,
HEAD `095ac6c3e` (clean at start); toolchain `go1.26.8 darwin/arm64`, `golangci-lint v2.1.6`.

| Command | Verbatim output | Exit |
|---|---|---|
| `go build ./...` | (empty) | 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | (empty) | 0 |
| `wc -c .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | `28308 .claude/rules/moai/workflow/kanban-dispatch.md` · `27986 internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` · `56294 total` | 0 |
| `wc -m` of the same two files | `28099 .claude/rules/moai/workflow/kanban-dispatch.md` · `27778 internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` · `55877 total` | 0 |
| `go run ./cmd/moai factory next --help` (FLAGS block) | `-h --help` · `--run         Factory run id (default: the single active run)` · `--wait        Keep re-checking at a fixed interval until a card is leased or the wait bound elapses` · `--wait-bound  How long --wait re-checks before reporting no card (15m0s)` — no `--card` | 0 |
| `git grep -n "factoryNominateBeforeRecord\|factoryNextRefusedExit" -- internal/cli` | (empty) — neither the seam nor the constant exists at `095ac6c3e` | 1 |

The develop absorption moved the always-loaded stub: live 28,308 B / 28,099 chars, mirror 27,986 B / 27,778
chars (the SPEC's pinned `4bf547bca` values were 26,959 / 26,754 and 26,637 / 26,433). M5's non-growth
bound must be measured against THESE values, re-measured again right before M5 edits.

Baseline of the existing pinned lease tests (anchored selector, lane variables scrubbed in one compound
invocation, `-v`, HEAD `095ac6c3e`): `TestFactoryNextSerialMutualExclusivity`,
`TestFactoryNextSkipsClassificationBlocked`, `TestFactoryNextParallelizableConcurrentLeases`,
`TestFactoryNextRecordAndClaimRaceOnLeasedRow` — 4 of 4 `--- PASS`, `ok  github.com/modu-ai/moai-adk/internal/cli  10.602s`, exit 0.

### M1 — tests first (RED), the inert seam, the golden

Tree measured: HEAD `095ac6c3e` + the one inert seam declaration `factoryNominateBeforeRecord` in
`internal/cli/factory_card.go` (uncalled) + the new test file `internal/cli/factory_nominate_test.go`
(uncommitted at measurement time; committed in the M1 commit). `go vet ./internal/cli` exit 0 — every new
test compiles.

**RED (E8), verbatim, captured BEFORE any GREEN code.** Command (one compound invocation, anchored,
19 names): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER && go test ./internal/cli -run '^(TestFactoryNextNominateLeasesNominee|TestFactoryNextNominateUnknownCard|TestFactoryNextFlagSet|TestFactoryNextNominateMCPParity|TestFactoryNextNominateConcurrentLanes|TestFactoryNextNominateSameCardExactlyOne|TestFactoryNextNominateRefusesKeepSet|TestFactoryNextArmCSkipsHoldMarker|TestFactoryNextNominateQuotaHold|TestFactoryNextNominateBackendSkip|TestFactoryNextNominateRecordStateTokens|TestFactoryNextNominateRefusalLeavesStateUnchanged|TestFactoryNextNominatePromoteThenLose|TestFactoryNextNominateClaimRefusedRollsBack|TestFactoryNextNominateCompensationFailure|TestFactoryNextAllMarkerQueueExitsNoCard|TestTodoLaneRefusesAutoCycle|TestTodoLaneAutoRefusalText|TestFactoryFallbackDeclarePrintsLeasePath)$' -count=1 -v`
→ exit 1, `FAIL github.com/modu-ai/moai-adk/internal/cli 140.021s`, 84 `=== RUN` lines, 19 of 19 top-level
tests `--- FAIL`. The failure reasons (distinct verbatim assertion lines; none is a compile error or a tool failure):

```text
factory_nominate_test.go:330: next --card t2: unknown flag: --card (stderr "")
factory_nominate_test.go:355: exit code = -1, want 4 for refused unknown-card; err=unknown flag: --card stderr=""
factory_nominate_test.go:383: factory next flags = [run wait wait-bound], want [card run wait wait-bound] (the baseline --run/--wait/--wait-bound plus --card)
factory_nominate_test.go:415: factory_next inputs = [project_root run], want exactly card,project_root,run
factory_nominate_test.go:426: factory_next card=t2 text = "t1 stage=- worktree=t1\nt1\tunknown\t\t\tpicked\t\tfactory card 1", want the leased t2 line
factory_nominate_test.go:453: factory_next card=t1 error = <nil>, want refused quota-hold
factory_nominate_test.go:453: factory_next card=t1 error = <nil>, want refused backend-skip
factory_nominate_test.go:542: lane lane-1's nomination of t1 ended before reaching the seam: err=unknown flag: --card stderr=""
factory_nominate_test.go:599: exit code = -1, want 4 for refused held; err=unknown flag: --card stderr=""
factory_nominate_test.go:641: bare next leased "t1 stage=- worktree=t1", want t3 (a marker card must be skipped)
factory_nominate_test.go:798: exit code = -1, want 4 for refused backend-skip; err=unknown flag: --card stderr=""
factory_nominate_test.go:842: exit code = -1, want 4 for refused raced; err=unknown flag: --card stderr=""
factory_nominate_test.go:871: the error = unknown flag: --card, want the injected failure reported
factory_nominate_test.go:921: all-marker queue: expected an error, got nil
factory_nominate_test.go:1100: a lane session ran `moai todo --auto`: output "jev: unavailable (no local scripts) — labelled non-finding; proceeding on the operator session's own judgment\nselection: source=fallback reason=jev-disabled\nselection: ranked t1 …
factory_nominate_test.go:1171: declare output = "fallback declared: trigger=channel-unavailable card=- at 2026-10-02T15:43:01Z\nswitch to /moai:todo --auto self-service pickup (REQ-FLA-001)\n", want the lease path `moai factory next [--card <id>]`
```

(Line numbers are those of the test file as measured; the arm-(c) line was re-measured after one helper fix: first
run printed `bare next leased "moai: worktree commits under global git identity …"` because the assertion read the
materializer's line — a test defect, fixed with `nmLeasedHead`, then re-run: `factory_nominate_test.go:641: bare next
leased "t1 stage=- worktree=t1", want t3 …`, exit 1.) One subtest passes by design:
`TestTodoLaneRefusesAutoCycle/read-only_forms_still_run` (the third Given of AC-TAU-005, a guard the lane
guard already satisfies).

Reasons per criterion: AC-TAU-001/-002/-004/-014 nominate through a flag that does not exist (`unknown flag:
--card`), the MCP parity row fails on the missing input and on the ignored `card`; AC-TAU-004 arm (c) and AC-TAU-006
all-marker fail because the marker card is leased today; AC-TAU-005 fails because a label-only, role-only and
role-and-label lane session RUNS the serial cycle (queue-mutating) and the declare line still routes to
`/moai:todo --auto`.

**GREEN-on-arrival guards (by design), HEAD `095ac6c3e` + seam.**
`unset … && go test ./internal/cli -run '^(TestFactoryNextBareUnchanged)$' -count=1 -v` → exit 0,
`--- PASS: TestFactoryNextBareUnchanged (11.30s)` with the five subtests `arm-order`, `wait-bound`,
`quota-hold`, `codex-skip`, `serial-slot` each `--- PASS`; the golden transcript was OBSERVED on this tree (the
first attempt printed the worktree materializer and the machine-dependent pull-request notes into the head and
stderr cells, so the transcript was narrowed to the leased-card line and the verb's own stderr lines — a test
fix, not a behavior change — and then matched the expected arm order `t5, t4, t3, t1, t2`, the bounded wait
`exit=3 sleeps=2`, the quota hold with the assigned card still leasing, the Codex skip of the merge-ready card,
and the serial slot `t1, t3, exit 3`). `TestTodoNonLaneGPTSessionNotRefused` → `--- PASS`
(`MOAI_KANBAN_BACKEND=gpt`, no role marker, no lane label: the cycle ran, `accept t1`).

**Seeded perturbation 1 — the golden is not vacuous (MU-25).** One-line mutation of
`internal/cli/factory_card.go`: `if noNewCards {` → `if noNewCards && false {` (the quota hold's early return
disabled), tree HEAD `095ac6c3e` + seam + this mutation. Command: `unset … && go test ./internal/cli -run
'^(TestFactoryNextBareUnchanged)$' -count=1` → exit 1:

```text
--- FAIL: TestFactoryNextBareUnchanged (15.68s)
    --- FAIL: TestFactoryNextBareUnchanged/quota-hold (2.31s)
        factory_nominate_test.go:1033: bare `factory next` transcript "quota-hold" changed:
            --- got ---
            next#1 exit=0 head="t4 stage=run worktree=t4" stderr=""
            next#2 exit=0 head="t2 stage=- worktree=t2" stderr=""
            --- want (golden) ---
            next#1 exit=0 head="t4 stage=run worktree=t4" stderr=""
            next#2 exit=3 head="" stderr="quota hold: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z"
FAIL	github.com/modu-ai/moai-adk/internal/cli	16.669s
```

Reverted by the inverse edit.

**Seeded perturbation 2 — the GPT guard is not vacuous (MU-22).** Over-broad predicate temporarily added at the top of
`todoRefuseLaneMutation` in `internal/cli/todo.go`: `if todoAutoFlag && factoryLaneRefusal() { return
fmt.Errorf("mutant: over-broad lane predicate refuses --auto") }`, tree HEAD `095ac6c3e` + seam + this mutation.
Command: `unset … && go test ./internal/cli -run '^(TestTodoNonLaneGPTSessionNotRefused)$' -count=1` → exit 1:

```text
--- FAIL: TestTodoNonLaneGPTSessionNotRefused (1.32s)
    factory_nominate_test.go:1164: a non-lane Codex-backend session was refused `moai todo --auto`: mutant: over-broad lane predicate refuses --auto
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.313s
```

Reverted by the inverse edit; `git diff --stat` afterwards: `spec.md` frontmatter (1 line) and `factory_card.go` (+7, the seam
declaration) only — `todo.go` untouched.

**M1 exit (d) — scoped baselines re-run after the seam (HEAD `095ac6c3e` + seam, lane variables scrubbed, one
compound invocation, `-v`).** The five existing lease pins plus the duplicate-dispatch and claim-refused pins, and the
six doc pins of G2: `TestFactoryNextSerialMutualExclusivity`, `TestFactoryNextSkipsClassificationBlocked`,
`TestFactoryNextParallelizableConcurrentLeases`, `TestFactoryNextRecordAndClaimRaceOnLeasedRow`,
`TestFactoryNextDuplicateDispatchGuard`, `TestFactoryNextClaimRefusedMapsRace`, `TestAutoRankDoctrineAmendment`,
`TestAutoRankMirrorParity`, `TestAutoRankMarkerDisclosure`, `TestAutoRankAgentDoctrine`,
`TestAutoHelpAndRefusalDoNotAssertPickOrder`, `TestTodoSkillDocumentsClassification` → exit 0, 12 of 12 top-level tests
`--- PASS` (73 `=== RUN` lines with subtests), `ok  github.com/modu-ai/moai-adk/internal/cli  12.575s`.

### M2 — nomination and the shared keep-set predicate (GREEN)

Files: `internal/cli/factory_card.go` only (plan §5). New: the named constant `factoryNextRefusedExit = 4`
(`git grep -n "factoryNextRefusedExit" -- internal/cli` was empty at M1; `moai slot` uses 4 as `slotExitBusy` in a
different verb family — no collision inside `factory*.go`), the twelve `factoryToken*` constants, the
`factoryNominateRefusal` error (carries the status itself), the shared predicate `factoryKeepSetRefusal`, the
read-only `factoryNextValidate`, `factoryNextNominate` (validate / promote inside `Mutate` with re-validation / seam then
claim through `factoryNextRecordAndClaim` and `factoryNextClaim` / compensate), `factorySerialInFlightExcluding`
(the serial-slot closure of the unnominated arms, extracted so both read one function), `factoryQueuedHoldMarked`
(arm (c) skip, counted as seen), the `--card` flag, and the nominated branch of the `next` loop (`--wait` waits
through `raced`, `serial-slot`, `quota-hold`, ends at once on any other refusal; a blank `--card` is an exit-1 error,
never a silent fall-back to the priority-order lease).

**GREEN, HEAD `095ac6c3e` + M1 commit `0f127e624` + the M2 working tree.** Command (one compound invocation, anchored,
16 names): `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED
MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER && go test ./internal/cli -run '^(TestFactoryNextNominateLeasesNominee|TestFactoryNextNominateUnknownCard|TestFactoryNextFlagSet|TestFactoryNextNominateConcurrentLanes|TestFactoryNextNominateSameCardExactlyOne|TestFactoryNextNominateRefusesKeepSet|TestFactoryNextArmCSkipsHoldMarker|TestFactoryNextNominateQuotaHold|TestFactoryNextNominateBackendSkip|TestFactoryNextNominateRecordStateTokens|TestFactoryNextNominateRefusalLeavesStateUnchanged|TestFactoryNextNominatePromoteThenLose|TestFactoryNextNominateClaimRefusedRollsBack|TestFactoryNextNominateCompensationFailure|TestFactoryNextAllMarkerQueueExitsNoCard|TestFactoryNextBareUnchanged)$' -count=1 -v`
→ exit 0, 75 `=== RUN` lines, 16 of 16 top-level tests `--- PASS` (59 subtests `--- PASS`), `ok
github.com/modu-ai/moai-adk/internal/cli  134.150s` (a loaded machine: other lanes were running). The golden
`TestFactoryNextBareUnchanged` stayed GREEN through the arm (c) change, the closure extraction and the new flag.
The two compensation-sensitive tests were re-run after a lint-driven rewrite of one condition (below): 4 of 4
`--- PASS`, `ok … 14.277s`.

**The existing lease and next-surface pins stay green** (anchored, scrubbed, `-v`): the six lease pins of
`factory_classify_test.go` (`TestFactoryNextSerialMutualExclusivity`, `TestFactoryNextSkipsClassificationBlocked`,
`TestFactoryNextParallelizableConcurrentLeases`, `TestFactoryNextRecordAndClaimRaceOnLeasedRow`,
`TestFactoryNextDuplicateDispatchGuard`, `TestFactoryNextClaimRefusedMapsRace`) plus `TestSD_AC008_NextSelectionOrderAndOutput`,
`TestSD_AC009_NextWaitLeasesOrTimesOut`, `TestSD_AC010_NextRefusedOutsideParent`, `TestSD_AC010_MCPNextParentCheck`,
`TestSD_AC011_CardWorktreeCreateReuseRefuse`, `TestSD_AC014_MCPMatchesCLIWithProjectRoot`,
`TestSD_AC014_ProjectRootRequired`, `TestQAS_AC008_ClaudeLaneHeldAtThreshold`, `TestQAS_AC008b_MCPFactoryNextHeld`,
`TestQAS_AC009_HoldLineCarriesResetTime` → exit 0, 16 of 16 `--- PASS`, 46 `=== RUN` lines, `ok … 148.281s`.

**Mutant check (MU-33, compensation steals the winner's queue state).** The restore guard disabled
(`if false && row != nil && …` in `factoryNominateCompensate`) → `go test … -run '^(TestFactoryNextNominatePromoteThenLose)$'` exit 1:
`factory_nominate_test.go:859: t1 queue state = queued, want picked (the competitor's state is left alone)`; reverted.

**Race detector.** `go test -race ./internal/cli -run '^(TestFactoryNextNominateConcurrentLanes|TestFactoryNextNominateSameCardExactlyOne|TestFactoryNextNominatePromoteThenLose)$' -count=1`
→ exit 0, no `DATA RACE`, `ok … 26.005s`. (The concurrency tests hold both invocations at the seam until both have
arrived, so the two claims contend for real; a lane label is read once at the start of an invocation.)

**Lint (CI version, stated: `golangci-lint v2.1.6`).** `golangci-lint run --timeout=5m ./internal/cli/` after the first M2
draft: 1 NEW issue, `factory_card.go:942:19: QF1001: could apply De Morgan's law (staticcheck)` — fixed in place (the
negated conjunction in `factoryNominateCompensate` rewritten as the disjunction); re-run: `0 issues.`, exit 0.
`GOOS=windows GOARCH=amd64 go build ./...` exit 0.

**Design notes and deviations from plan.md (none contradicts a requirement).**
1. `blocked` is applied to a nominee in any queue state (REQ-TAU-009 states it unqualified; only `hold-marker` is
   qualified `queued`), so an operator-picked card whose classification is `blocked` is refused when NOMINATED while
   the bare arms (b)/(b2) still lease it as before. The bare path is untouched by design.
2. A queue state outside `{queued, picked, dropped, hold}` (a state added later) is refused with the token `held` and a
   detail naming the state — positive enumeration of the leasable states, so a new state is never leased by accident.
3. Check order inside the read-only validation: `unknown-card`, then the keep-set predicate (`dropped`, `held`, `owned`,
   `recorded`, `hold-marker`, `blocked`, `serial-slot`), then `foreign-worktree`, `quota-hold`, `backend-skip`. spec § C.2 is a
   set, not an order.
4. The refusal line is printed by the verb itself (the real root never prints an exit-coded error, as for the quota
   hold line); the cobra error carries status 4 through `ExitCoder`.
5. `compensation failed` (the restoring write errors) and a post-`RecordPicked` failure remain specified-and-untested,
   as spec § B.8 and § G state; the single seam sits before `RecordPicked`.

### M3 — a lane is refused `moai todo --auto`; the fallback declare line routes to the lease (GREEN)

Files: `internal/cli/todo.go` (`todoRefuseLaneMutation` gains one early clause; new `todoLaneSession` and
`todoLaneAutoRefusalText`) and `internal/cli/factory_messaging.go` (the printed declare instruction and its comment).
The lane definition is exactly REQ-TAU-008's: the role marker equals `lane`, or the lane label variable is non-empty
(`factoryLaneAdmission() || os.Getenv(config.EnvMoaiFactoryWorker) != ""`); `factoryLaneRefusal()` is NOT used here
and is unchanged (its Codex-backend clause would refuse the non-lane Codex-backend session, row C4 / MU-22). The
refusal text is the dedicated function, it names `moai factory next --card <id>` and contains `the --auto
authorization is exercised through`, and it does not reuse the queue-mutation wording.

**GREEN, HEAD `efec50cdc` (M2) + the M3 working tree.** `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL
MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER && go test ./internal/cli -run
'^(TestTodoLaneRefusesAutoCycle|TestTodoLaneAutoRefusalText|TestTodoNonLaneGPTSessionNotRefused|TestFactoryFallbackDeclarePrintsLeasePath|TestSD_AC015_LaneQueueAllowlistWalk|TestSD_AC015_LabelOnlyIsNotALane|TestSD_AC015_MCPTodoAddRefused|TestFactoryFallbackDeclareRestoreLifecycle|TestTodoAutoEntryPointFlag)$'
-count=1 -v` → exit 0, 16 `=== RUN` lines, 9 of 9 top-level tests `--- PASS` (the four new M3 tests, the three existing
lane-allowlist pins, the declare/restore lifecycle and the `--auto` entry-point test), `ok
github.com/modu-ai/moai-adk/internal/cli  50.476s`. The three lane variants (`label-only`, `role-only`, `role-and-label`)
now refuse in about 2 s each (they ran the 11 s serial cycle at M1); the non-lane Codex-backend session still runs the cycle
(`TestTodoNonLaneGPTSessionNotRefused` `--- PASS`, 13.52 s). The doc pins of G2 re-run after the change: 6 of 6 `--- PASS`, exit 0,
`ok … 1.793s`. `golangci-lint run ./internal/cli/` (v2.1.6) `0 issues.`, exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

Superseding note (REQ-TAU-008, spec § B.4): in lane sessions REQ-FLA-001's reference to `/moai:todo --auto` is superseded
by the lease path; the declare confirmation now prints `switch to self-service pickup through the lease: moai factory next [--card <id>]`
and the comment above `newFactoryFallbackDeclareCommand` says so. The completed SPEC's body is not edited.

### M4 — the MCP `card` parameter (GREEN)

File: `internal/cli/mcp_factory_card.go` only. `factory_next` gains an optional `card` string; with it the handler calls
the SAME `factoryNextNominate` the cobra verb calls (the quota hold is passed through the same shared evaluation) and a
refusal returns the same one-line `factory next: refused <token>: <detail>` as an error result (the tool name prefixes it, as
for every tool error); without it the handler is the unnominated lease exactly as before; a `card` argument that is
present but blank is an error, not a silent fall-back. The tool's inputs are now exactly `run`, `project_root`, `card`
(two inputs before; the CLI's `--wait`/`--wait-bound` have no MCP counterpart). `internal/mcp/catalog.go` carries only name
and write-capability — no change was needed, confirmed by the catalog guards below.

**GREEN, HEAD `f70fe64c7` (M3) + the M4 working tree.** `unset … && go test ./internal/cli -run
'^(TestFactoryNextNominateMCPParity|TestSD_AC014_MCPMatchesCLIWithProjectRoot|TestSD_AC014_ProjectRootRequired|TestSD_AC010_MCPNextParentCheck|TestQAS_AC008b_MCPFactoryNextHeld|TestMoaiMCPServer_AnnotationsMatchCatalog|TestMoaiMCPServer_RegistrationMatchesCatalog)$'
-count=1 -v` → exit 0, 21 `=== RUN` lines, 7 of 7 top-level tests `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/cli  90.843s`;
`TestFactoryNextNominateMCPParity` subtests `inputs`, `leases_the_nominee_and_refuses_an_unknown_card`, `refuses_quota-hold`,
`refuses_backend-skip` each `--- PASS`. The M1 RED of the same test (recorded above): `factory_nominate_test.go:415:
factory_next inputs = [project_root run], want exactly card,project_root,run` and `:426: factory_next card=t2 text = "t1 stage=- worktree=t1 …",
want the leased t2 line` — the card input was ignored (MU-19). The internal/mcp catalog guards, anchored: `go test ./internal/mcp -run
'^(TestMoaiMCPTools_CatalogSize|TestMoaiMCPTools_WriteCapableSet|TestMoaiMCPTools_NoDuplicateNames|TestMoaiMCPToolNames_MatchesCatalog)$' -count=1 -v`
→ exit 0, 4 of 4 `--- PASS`, `ok github.com/modu-ai/moai-adk/internal/mcp  0.655s`.

### Run-phase M1-M4 verification summary (this run, this tree; E1-E8 index)

Tree measured for the closing evidence: HEAD `f70fe64c7` (the M3 commit) + the M4 working tree (committed as the M4 commit).

- **E1 AC matrix for the criteria M1-M4 flip.** Command for all 21 tests: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL
  MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER && go test ./internal/cli -coverprofile=<scratch>/cover.out
  -run '^(TestFactoryNextNominateLeasesNominee|TestFactoryNextNominateUnknownCard|TestFactoryNextFlagSet|TestFactoryNextNominateMCPParity|TestFactoryNextNominateConcurrentLanes|TestFactoryNextNominateSameCardExactlyOne|TestFactoryNextNominateRefusesKeepSet|TestFactoryNextArmCSkipsHoldMarker|TestFactoryNextNominateQuotaHold|TestFactoryNextNominateBackendSkip|TestFactoryNextNominateRecordStateTokens|TestFactoryNextNominateRefusalLeavesStateUnchanged|TestFactoryNextNominatePromoteThenLose|TestFactoryNextNominateClaimRefusedRollsBack|TestFactoryNextNominateCompensationFailure|TestFactoryNextAllMarkerQueueExitsNoCard|TestFactoryNextBareUnchanged|TestTodoLaneRefusesAutoCycle|TestTodoLaneAutoRefusalText|TestTodoNonLaneGPTSessionNotRefused|TestFactoryFallbackDeclarePrintsLeasePath)$'
  -count=1 -v` → exit 0, 91 `=== RUN` lines, **21 of 21 top-level tests `--- PASS`**, `ok github.com/modu-ai/moai-adk/internal/cli  396.072s  coverage: 8.1% of statements`
  (a scoped run; the figure is the whole package's statements, so it is low by construction — see E3).

  | AC | Status | Evidence (all `--- PASS` in the run above) |
  |---|---|---|
  | AC-TAU-001 | PASS | `TestFactoryNextNominateLeasesNominee`, `TestFactoryNextNominateUnknownCard`, `TestFactoryNextFlagSet`, `TestFactoryNextNominateMCPParity` |
  | AC-TAU-002 | PASS | `TestFactoryNextNominateConcurrentLanes`, `TestFactoryNextNominateSameCardExactlyOne` (also clean under `-race`, M2 note) |
  | AC-TAU-004 | PASS | `TestFactoryNextNominateRefusesKeepSet` (subtests `held`, `hold-marker`, `marker-leading-space`, `marker-mid-text`, `blocked`, `serial-slot`, `dropped`, `owned`), `TestFactoryNextArmCSkipsHoldMarker` |
  | AC-TAU-005 (Go rows) | PASS | `TestTodoLaneRefusesAutoCycle`, `TestTodoLaneAutoRefusalText`, `TestTodoNonLaneGPTSessionNotRefused`, `TestFactoryFallbackDeclarePrintsLeasePath`; the routing-sentence row is M5 |
  | AC-TAU-006 | PASS | `TestFactoryNextBareUnchanged` (golden, five subtests), `TestFactoryNextAllMarkerQueueExitsNoCard` |
  | AC-TAU-014 | PASS | `TestFactoryNextNominateQuotaHold`, `TestFactoryNextNominateBackendSkip`, `TestFactoryNextNominateRecordStateTokens`, `TestFactoryNextNominateRefusalLeavesStateUnchanged`, `TestFactoryNextNominatePromoteThenLose`, `TestFactoryNextNominateClaimRefusedRollsBack`, `TestFactoryNextNominateCompensationFailure` |

  Not flipped by M1-M4 and not claimed: AC-TAU-003, -007, -008, -010, -011, -013 (M5), -009 (the first lease taken under the doctrine), -012 (M6 reads the finished diff).
- **E2 builds.** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (both re-run after M4; both also run before any change and after M2 and M3).
- **E3 coverage of the changed code.** Measured with the profile of the run above, `go tool cover -func`: `factoryKeepSetRefusal` 92.9%,
  `factoryNextValidate` 93.9%, `factoryRecordRefusal` 88.9%, `factoryNominateCompensate` 87.5%, `factoryNextNominatedClaim` 100%,
  `factoryQueuedHoldMarked` 100%, `factoryRefusal`/`waitable`/`Error`/`ExitCode` 100%, `factorySerialInFlightExcluding` 83.3%,
  `factoryNextSelectAndLease` 83.1%, `newFactoryNextCommand` 87.1%, `factoryNextNominate` 72.5%, `todoRefuseLaneMutation` 90.0%,
  `todoLaneSession` 100%, `todoLaneAutoRefusalText` 100%, `handleFactoryNext` 70.5%, `factoryRowHolder` 66.7%. What the profile shows
  uncovered: store/open error branches, the `compensation failed` branch (specified and untested, spec § G), the in-lock
  `raced`/re-validation `lost` branches of the promotion (no seam sits between the validation and the promotion, so that window
  cannot be forced from a test), and `handleFactoryNext`'s unchanged unnominated tail paths. Package-wide coverage (all of `internal/cli`) was NOT measured — a
  full-package run on a loaded machine is outside the lane-local verification budget — so the 85% package threshold is neither claimed nor refuted here;
  the changed functions are the unit measured, and the lowest (66.7-72.5%) are error-only branches.
- **E4 subagent-boundary grep.** `grep -n 'AskUserQuestion\|mcp__askuser' internal/cli/factory_card.go internal/cli/todo.go internal/cli/factory_messaging.go internal/cli/mcp_factory_card.go internal/cli/factory_nominate_test.go`
  → no output, exit 1 (no match in any file changed). Package-wide, `grep -rn 'AskUserQuestion\|mcp__askuser' --include='*.go' internal/cli | grep -v _test.go | grep -v '// ' | wc -l`
  → `18` (pre-existing help-text strings in files this run did not touch, e.g. `harness.go`, `pr_watch_cmd.go`, `harness_mute.go`).
- **E5 lint.** `golangci-lint v2.1.6`, `golangci-lint run --timeout=5m ./internal/cli/` → `0 issues.`, exit 0, after M4 (the one NEW finding of the M2 draft, `QF1001`, was
  fixed in place; no baseline finding exists in the package). `gofmt -l` on the changed files is empty (`internal/cli/mcp_claude.go` is listed by `gofmt -l internal/cli/`; it predates this work and is untouched).
- **E6 commits (no push; the repository's lanes never push):** `0f127e624` M1, `efec50cdc` M2, `f70fe64c7` M3, M4 (this commit, SHA in the completion report).
- **E7 blocker report:** none.
- **E8 RED output verbatim before GREEN:** recorded under M1 above (19 of 19 failing for the stated reasons; the two guards green with their seeded perturbations).

Preserved and untouched (spec § D / plan §5): `internal/kanban/**`, `internal/graph/**`, every queue/gtd schema file, `.claude/`, `internal/template/`,
`internal/template/catalog.yaml`, docs, and the bodies of `spec.md` / `plan.md` / `acceptance.md` / `research.md` (the one `spec.md`
change is the frontmatter transition recorded at the top of this section). No M5/M6 work was started.

### M5 — doctrine amendment, mirrors, generated artifacts, doc pins (one commit `3af5605ab`; carrier: general-purpose in the manager-develop role, same OWNERSHIP-EXCEPTION as M1-M4)

The line "No M5/M6 work was started" above is a statement as of M4 and is left as written; this section supersedes it.

- **Pins first, RED before any doc edit** (tree `d017336a4` plus the two new/moved test files, uncommitted; lane env scrubbed in one compound `unset … && go test`, slot `go-test-cli-t1448` held):
  `go test ./internal/cli -run '^(TestAutoPickDocDoctrine|TestAutoPickMirrorParity|TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity)$' -count=1 -v` → exit 1, 145 `=== RUN` lines: `--- FAIL: TestAutoRankDoctrineAmendment`, `--- FAIL: TestAutoRankMirrorParity`, `--- FAIL: TestAutoPickDocDoctrine`, `--- PASS: TestAutoPickMirrorParity` (a guard, green on arrival by design); 108 failing subtests, 33 passing. Stated reasons, verbatim excerpts:
  `todo_auto_doc_test.go:136: live kanban-dispatch.md lost the prohibition "Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item."`;
  `todo_auto_doc_test.go:206: live kanban-dispatch.md promotion and --auto reconciliation clauses: passage start marker "[HARD] **Promotion is the operator's act, in person or in advance.**" not found`;
  `todo_auto_pick_doc_test.go:189: live gtd.md does not state "demotes a card in the serial cycle and excludes it on the lease path"`;
  `todo_auto_pick_doc_test.go:196: live gtd.md still carries the replaced sentence "consumption of the queue and nothing else"`;
  `todo_auto_pick_doc_test.go:196: template manager-todo.md still carries the replaced sentence "process cards in queue order"`;
  `todo_auto_pick_doc_test.go:237: live auto-semantics.md 9.3: passage start marker "### 9.3 " not found`.
  An earlier first run had shown the drift subtest of `TestAutoPickMirrorParity` red on the unmodified tree for a wrong reason (the live-only sentence is the tail of one paragraph line, not a separate line); the subtest was corrected to "equal line count, exactly one differing line, the mirror line a strict prefix of the live line" before the RED above was taken.
- **AC-TAU-010 seeded probes (step 2, before regeneration; tree `d017336a4` plus the uncommitted doc edits):**
  (a) `AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v` → exit 1, `--- FAIL: TestGoldenCommittedArtifactsMatchEmission`, `golden_test.go:101: .codex/agents/moai/manager-todo.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing`.
  (b) `go test ./internal/template -run '^(TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles)$' -count=1 -v` → exit 1, both `--- FAIL`; `CATALOG_HASH_UNSTABLE: moai`, `CATALOG_HASH_UNSTABLE: moai-kanban-foreman`, `CATALOG_HASH_UNSTABLE: manager-todo`, `CATALOG_HASH_SKINNY: moai`, `CATALOG_HASH_SKINNY: moai-kanban-foreman`.
  (c) One-sided mirror edit: `printf 'X' >> internal/template/templates/.claude/skills/moai/workflows/gtd.md`, then `cmp .claude/skills/moai/workflows/gtd.md internal/template/templates/.claude/skills/moai/workflows/gtd.md` → stdout `cmp: EOF on .claude/skills/moai/workflows/gtd.md`, exit 1; `go test ./internal/cli -run '^(TestAutoPickMirrorParity)$' -count=1 -v` → `todo_auto_pick_doc_test.go:292: .claude/skills/moai/workflows/gtd.md differs from its template mirror (40075 vs 40076 bytes)`, `--- FAIL: TestAutoPickMirrorParity`. Reverted by restoring the saved good copy; `cmp` of the pair → exit 0 afterwards; the pair is byte-identical in `3af5605ab`.
- **Old pins against the new docs (step 3, expected breakages)**, old `todo_auto_doc_test.go` temporarily restored from `HEAD`, the moved version put back afterwards: `go test ./internal/cli -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure|TestAutoRankAgentDoctrine|TestAutoHelpAndRefusalDoNotAssertPickOrder|TestTodoSkillDocumentsClassification)$' -count=1 -v` → exit 1: `--- FAIL: TestAutoRankDoctrineAmendment` (`prohibition_kept/live_kanban-dispatch.md` and `prohibition_kept/template_kanban-dispatch.md`: `lost the prohibition "The leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item."`) and `--- FAIL: TestAutoRankMirrorParity` (`passage start marker "[HARD] **Promotion is the operator's act, always.**" not found`); `TestAutoRankMarkerDisclosure`, `TestAutoRankAgentDoctrine` (no stale phrase reintroduced), `TestAutoHelpAndRefusalDoNotAssertPickOrder`, `TestTodoSkillDocumentsClassification` PASS. Exactly the two predicted breakages; both pins moved in the same commit.
- **Regeneration, after the last template edit:** `make agents-emit` → `AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission` → `ok`, exit 0; only `internal/template/templates/.codex/agents/moai/manager-todo.toml` changed (numstat `6 4`). `go run ./internal/template/scripts/gen-catalog-hashes.go --all` (run from the worktree root; rewrites every catalog entry's `hash:` in place, default catalog and templates paths) → `catalog.yaml updated successfully (13900 bytes)`, exit 0; `git diff --numstat -- internal/template/catalog.yaml` → `3	3` (the `moai`, `moai-kanban-foreman`, `manager-todo` hash lines).
- **GREEN after the edits:** the eight-name scoped set (the two new tests, the two moved pins, `TestAutoRankMarkerDisclosure`, `TestAutoRankAgentDoctrine`, `TestAutoHelpAndRefusalDoNotAssertPickOrder`, `TestTodoSkillDocumentsClassification`) → exit 0, 200 `=== RUN` lines, 8 of 8 `--- PASS`, `ok … internal/cli 1.677s`.
- **Replacements, literals raw-clean (committed tree, `git grep -c -F`):** the old sentence `consumption of the queue and nothing else` over the six docs plus the Codex TOML (ledger L5, L23) → no output, exit 1; L6, L7, L8+L24, L25 → no output, exit 1 each; L9 `factory next --card` → `manager-todo.md:1`, `auto-semantics.md:1`, `kanban-dispatch.md:1`, `moai-kanban-foreman/SKILL.md:1`, `gtd.md:1`, exit 0; L10 `ladder_path=gate-row card pick` → `auto-semantics.md:1`, `gtd.md:1`, exit 0; L19 → both `gtd.md` copies, exit 0; L20 → both `kanban-dispatch-detail.md` copies, exit 0.
- **Diff bound (AC-TAU-013):** `git merge-base develop HEAD` → `7109e0900a060cda33269ebff071272ba64e840e`; `git diff --numstat <that sha>..HEAD -- <paths>` per file, added/deleted (cap), live equal to mirror in every pair: `kanban-dispatch.md 3/3` (3/3) · `kanban-dispatch-detail.md 2/0` (12/0) · `auto-semantics.md 36/2` (45/2) · `gtd.md 43/9` (80/24) · `manager-todo.md 6/4` (12/12) · `manager-todo.toml 6/4` (12/12) · `moai-kanban-foreman/SKILL.md 5/3` (12/12) · `catalog.yaml 3/3` (3/3) · `moai-mcp-tools-catalogue.md 1/1` (1/1). Every file meets its cap and has at least one added line (floor).
- **Always-loaded stub (AC-TAU-011), `kanban-dispatch.md`, merge-base blob to working copy** (`git cat-file -s <sha>:<path>`; `git show <sha>:<path>` to a scratch file then `wc -m`; `wc -c` and `wc -m` on the working copies): live 28308 B / 28099 chars → 28297 B / 28088 chars (−11 / −11); mirror 27986 B / 27778 chars → 27975 B / 27767 chars (−11 / −11). Conjunct (a): the old sentence is absent from both (exit 1). The feasibility draft of plan §3 predicted −12 B / −16 chars; the measured −11 / −11 differs because the committed wording also carries the raw-cell literal `factory next --card` in the second paragraph (so ledger cell L9 lists `kanban-dispatch.md`) and a shorter heading. `kanban-dispatch-detail.md` (lazy): 42940 B / 42675 chars → 43403 B / 43138 chars (+463 / +463).
- **Mirrors (AC-TAU-010):** `cmp` of the six pairs `kanban-dispatch-detail.md`, `auto-semantics.md`, `gtd.md`, `manager-todo.md`, `moai-kanban-foreman/SKILL.md`, `moai-mcp-tools-catalogue.md` → exit 0 each; `kanban-dispatch.md` → `differ: char 25109, line 181`, exit 1; `git diff --no-index --numstat` of that pair → `1	1`, exit 1 (the one preserved live-only `moai worktree sweep …` tail of line 181, unabsorbed in either direction and enforced by `TestAutoPickMirrorParity`).

### M6 — verification and measurement only (this run, this tree, HEAD `3af5605ab`; no code or doc edits; lane env scrubbed; slot `go-test-cli-t1448` held for the `internal/cli` run and released after it)

- `internal/cli`: `go test ./internal/cli -run '^(TestFactoryNext|TestAutoRank|TestAutoPick|TestAutoHelp|TestTodoAuto|TestTodoLane|TestTodoNonLane|TestTodoSkill|TestFactoryFallback)' -count=1 -v` → exit 0, 432 `=== RUN` lines, 94 top-level `--- PASS`, 0 `--- FAIL`, 0 `--- SKIP`, `ok … internal/cli 216.821s`. The M1-M4 nomination, keep-set, lane-refusal, fallback-string and golden tests, `TestAutoPickDocDoctrine` and `TestAutoPickMirrorParity` are among the passes.
- `internal/template`: `go test ./internal/template -run '^(TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles|TestJevAutoExceptionLinkage|TestJevAutoExceptionWording|TestDeclaredRuleMirrorForks|TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestContractModeGuidedPreservation|TestContractModeEmitterSites)$' -count=1 -v` → exit 0, 64 `=== RUN` lines; `catalog_tier_audit_test.go:507: audited 37 directory entries for whole-tree hash coverage`, `catalog_tier_audit_test.go:464: audited 49 catalog entries for hash validity` (equal to the planned 37 / 49), both catalog guards `--- PASS`; `TestContractModeEmitterSites` and the `TestContractModeGuidedPreservation/tree` subtest print `--- SKIP` (they need `MOAI_GR_BASE`; see Gaps).
- `AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v` → `--- PASS: TestGoldenCommittedArtifactsMatchEmission (0.00s)`, `ok`.
- `go test ./internal/spec -run '^TestACCounterFullCorpusMatchesBaseline$' -count=1 -v` → `--- PASS: TestACCounterFullCorpusMatchesBaseline (10.93s)`, `ok` (confirmation; no regeneration needed).
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli` exit 0; `golangci-lint` v2.1.6 `run --timeout=5m ./internal/cli/` → `0 issues.`; `gofmt -l` on the two touched test files → empty.
- Boundary (AC-TAU-012): `git diff --name-only 7109e0900a060cda33269ebff071272ba64e840e..HEAD` lists 29 paths (control non-empty); the same range limited to `internal/kanban internal/graph` → empty, exit 0; the non-test `internal/cli` files in the list are `factory_card.go`, `factory_messaging.go`, `mcp_factory_card.go`, `todo.go`; `manager-todo.toml` and `catalog.yaml` are the two allowed generated additions. `grep -n 'AskUserQuestion\|mcp__askuser'` over those four Go files → no output, exit 1.

### Gaps and residual risk (M5/M6)

- AC-TAU-009 (the decision record as a lane writes it) is a regression-guard with no executing party; no lease has been taken under the new doctrine, so it is not recorded as a pass.
- The base-ref guards `TestContractModeInheritedDivergence`, `TestContractModeConstitutionDriftNotIncreased`, `TestContractModeAlwaysLoadedBudget`, `TestContractModeEmitterSites` and the `/tree` subtests skip without `MOAI_GR_BASE`. One probe run with `MOAI_GR_BASE=7109e0900a060cda33269ebff071272ba64e840e` (this card's merge base, not necessarily the base those guards were written for) failed `TestContractModeGuidedPreservation/tree`, `TestContractModeChangeSetAllowlist/tree` and `TestContractModeEmitterSites` for reasons that read as unrelated to this card's wording (contract-mode blocks differ at that base, changed paths outside that guard's own allowlist, `auto-semantics.md`, `contract-autonomy.md` and `moai-mcp-tools-catalogue.md` listed as unclassified Kickoff documents at that base). The added doc lines carry no `Kickoff` word (case-insensitive count of added lines over the changed docs: 0). The guards are therefore neither claimed green nor red for this card.
- Of the acceptance mutants only the one-sided mirror edit (MU-12 form) was executed; the others were reasoned from the test bodies, not run.
- The user-facing pages that still say the pick is always the operator's (README, docs-site pages) are sync-phase scope and untouched, as the plan requires.
- Findings for the leader: (1) `kanban-dispatch-detail.md` was already over the 40,000-character instruction budget on develop (42,675 chars at the merge base; 43,138 after this card's +463-char paragraph); a follow-up card should split it. (2) The pre-existing live-only `moai worktree sweep …` sentence of `kanban-dispatch.md` is preserved and still differs between the copies; a follow-up card decides which is right. (3) The operator or leader must apply `moai gtd hold` or a leading `[보류` marker to t810 (`picked`), t1294 and t1383 (`queued`) before lanes exercise the doctrine (spec §B.3 handoff sentence; this card does not touch the queue).

### sync-audit iteration 1 repair (findings F1, F2, F4, F3 part b, F5; this run, this tree)

Carrier: a general-purpose agent carrying the manager-develop role (same ownership exception as M1-M6: a typed
`manager-develop` spawn auto-isolates into its own agent worktree and cannot write the card tree). No spec.md,
plan.md, acceptance.md or research.md body was edited; `docs-site/`, `CHANGELOG.md` and §E.4 are untouched (the sync
phase owns them). Pre-flight: toplevel `…/.moai/worktrees/t1448`, branch `WT-todo-auto-pick-autonomy`, HEAD
`8de769d81`, `git status --short` empty. Lane env scrubbed in the same compound invocation as every `internal/cli`
run (`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED
MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER && go test …`); heavy-run lease `go-test-cli-t1448` held for the runs and
released after them. Code and test commit `e55aaeb1b`; doctrine commit recorded by its own subject.

- **R1 (F1, blocking) — `TestFactoryNextNominateConcurrentLanes` made deterministic.** RED (before any edit):
  `go test ./internal/cli -run '^TestFactoryNextNominateConcurrentLanes$' -count=10 -v` → exit 1, 10 `=== RUN`, 1 PASS /
  9 FAIL, failure line `factory_nominate_test.go:559: lane lane-1 nominating t1 errored: factory next: rename the card
  worktree branch: exit status 128: error: unable to move logfile …/.git/logs/refs/.tmp-renamed…`. Cause (as the audit
  found): both in-process lanes run the pre-existing `factoryEnsureCardWorktree` (`git branch -m`) at the same instant
  in ONE repository and collide on git's reflog temp file. Fix (test only, no production change for the collision):
  `nmIsolatedWorktrees` stubs `worktree.WorktreeCreator` for this test so each card gets its own independent git
  repository (built up front on the test goroutine, restored by `t.Cleanup`); the queue promotion and the
  version-checked record edges under test are untouched, and `TestFactoryNextNominateSameCardExactlyOne` is unchanged
  (its loser is refused before any worktree step, so it never raced). The only other concurrent tests
  (`factory_classify_test.go` ×2, `factory_self_dispatch_test.go`) call `factoryNextLeaseOnce`, which has no worktree
  step, so they cannot hit it. GREEN: `… -count=30 -v` → exit 0, 30 `=== RUN`, 30 PASS, 0 FAIL, `ok … 68.290s`.
  `go test -race ./internal/cli -run '^(TestFactoryNextNominateConcurrentLanes|TestFactoryNextNominateSameCardExactlyOne|TestFactoryNextNominatePromoteThenLose|TestFactoryNextNominateCompensateRechecksRecord)$' -count=5 -v`
  → exit 0, 55 `=== RUN` (subtests included), 20 top-level PASS (5 each), 0 `DATA RACE`, `ok … 95.203s`.
- **R3 (F3 part b) — compensation re-reads the record inside the queue lock.** RED: new
  `TestFactoryNextNominateCompensateRechecksRecord/a_lease_landing_while_the_compensation_waits_for_the_queue_lock`
  (the test holds the queue lock through `todoStoreAt(root).Mutate`, starts `factoryNominateCompensate`, lets lane-2
  lease the card while the compensation waits, then releases the lock) against the unchanged code: `-count=3` → exit 1,
  3 of 3 runs FAIL with `t1 queue state = queued, want picked`; the five direct-call subtests pass on the old code.
  Fix: `factoryNominateCompensate` now performs the `LoadCard` read inside the single `Mutate` closure and restores
  `queued` only when no row exists or the row is `picked` with no owner; otherwise it leaves the queue item untouched
  and reports `otherHolder` exactly as before (`raced` refusal for another lane, the claim error for this lane's own
  row). A read or write error is still the non-token `compensation failed: card <id> stays picked and unowned: …`.
  GREEN: the same test passes `-count=5` under `-race` (above) and in the full selector below. Not closed, by design:
  F3 (a) (serial slot snapshot reused inside the lock) and F3 (c) (operator `unpick`→`hold` between promotion and
  claim) — same two-store non-atomicity class as the unnominated arm (c); a lease landing after the in-lock read and
  before the restoring write is the same accepted window (spec §B.8). The sync phase records the three as residual
  risk.
- **R2 (F4) — mutants M1, M2, M5 now killed.** New tests: `TestFactoryNextNominateBlankCardIsAnError` (CLI `--card=`
  and `--card=   `, MCP `card: ""` and `"   "`: error `--card needs a card id` / `card needs a card id`, not exit 4,
  stores byte-identical, no lease), subtests `owned/picked-with-owner-lane-2` and `…-lane-1` in
  `TestFactoryNextNominateRecordStateTokens` (a `picked` row with an owner is `owned` for any lane), and case
  `blocked-picked` in `nmCases` (a `picked` nominee classified `blocked` is `blocked`; it also runs through
  `TestFactoryNextNominateRefusesKeepSet` and `…RefusalLeavesStateUnchanged`). Mutation checks (production file copied
  to the scratchpad first, one line reverted by `Edit`, test run, file restored with `cp` and compared with `cmp`):
  M1a CLI guard disabled → `TestFactoryNextNominateBlankCardIsAnError/cli_""` and `/cli_"___"` FAIL (the bare lease
  ran: output `t1 stage=- worktree=t1`); M1b MCP guard disabled → `/mcp_""` and `/mcp_"___"` FAIL (`err = <nil>`);
  M2 `factoryRecordRefusal` made to lease an owned `picked` row → both `owned/picked-with-owner-*` subtests FAIL
  (`want exit 4, token "owned"`); M5 `blocked` limited to `queued` → `RefusesKeepSet/blocked-picked` and
  `RefusalLeavesStateUnchanged/blocked/blocked-picked` FAIL. After each restore `cmp` reported identical files and
  `git diff -- internal/cli/mcp_factory_card.go` was empty (that file ends unchanged; `factory_card.go` differs from
  HEAD only by the R3 and R5 edits). M3, M4, M6 were not pursued (the brief marked them lower value; M6 gained direct
  coverage as a side effect of the R3 table: another lane's lease/assignment returns `true`, this lane's own lease
  returns `false`).
- **R5 (F5).** The `blocked` refusal detail now reads `the card's classification is blocked; no lane lease takes it,
  the operator decides it`; no test pinned the old words (a `grep` for the phrase over `internal`, `.claude` and `.moai`
  found it in `factory_card.go` only as a code comment and the detail line, and in an older SPEC's text).
- **Anchored selector after R1-R5 (23 tests, 108 `=== RUN` lines):** `go test ./internal/cli -run
  '^(TestFactoryNextNominateLeasesNominee|TestFactoryNextNominateUnknownCard|TestFactoryNextFlagSet|TestFactoryNextNominateMCPParity|TestFactoryNextNominateConcurrentLanes|TestFactoryNextNominateSameCardExactlyOne|TestFactoryNextNominateRefusesKeepSet|TestFactoryNextArmCSkipsHoldMarker|TestFactoryNextNominateQuotaHold|TestFactoryNextNominateBackendSkip|TestFactoryNextNominateRecordStateTokens|TestFactoryNextNominateRefusalLeavesStateUnchanged|TestFactoryNextNominatePromoteThenLose|TestFactoryNextNominateClaimRefusedRollsBack|TestFactoryNextNominateCompensationFailure|TestFactoryNextNominateCompensateRechecksRecord|TestFactoryNextNominateBlankCardIsAnError|TestFactoryNextAllMarkerQueueExitsNoCard|TestFactoryNextBareUnchanged|TestTodoLaneRefusesAutoCycle|TestTodoLaneAutoRefusalText|TestTodoNonLaneGPTSessionNotRefused|TestFactoryFallbackDeclarePrintsLeasePath)$'
  -count=1 -v` → exit 0, 23 top-level PASS, 0 FAIL, `ok … 145.424s`. Static checks on the code commit tree:
  `gofmt -l` on the three touched Go files empty; `go vet ./internal/cli` exit 0; `golangci-lint` v2.1.6 `run
  --timeout=8m ./internal/cli/` → `0 issues.`, exit 0; `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build
  ./...` exit 0.
- **R4 (F2) — doctrine scoped to a lane session.** The two phrases (`each only through a lease`, `never a keep-set
  card`) are now stated of a lane session in `kanban-dispatch.md` (the one-reconciliation paragraph), `gtd.md` (the
  `--auto` authority clause) and `auto-semantics.md` (§9.2 sentence and §9.3 first paragraph), each in the live file
  and its template mirror. The kanban-dispatch edit also trims `invocation event` to `invocation` (unpinned) to pay for
  the added words. No doc pin asserted the over-broad wording, so no pin moved. `manager-todo.md` (and the generated
  TOML) already scope the lease to a lane session and were not edited; the `moai-kanban-foreman` skill says only
  "within the keep-set" and was not edited either. Measurements (this tree): `kanban-dispatch.md` live 28304 B / 28095
  chars vs merge-base blob `7109e0900` 28308 B / 28099 chars (−4 / −4); mirror 27982 B / 27774 chars vs 27986 B /
  27778 chars (−4 / −4). `cmp` of the seven pairs: `kanban-dispatch-detail.md`, `auto-semantics.md`, `gtd.md`,
  `manager-todo.md`, `moai-kanban-foreman/SKILL.md`, `moai-mcp-tools-catalogue.md` → exit 0 each;
  `kanban-dispatch.md` → `differ: char 25116, line 181` (the one pre-existing live-only `moai worktree sweep …`
  sentence), `git diff --no-index --numstat` → `1 1`. `git diff --numstat 7109e0900` per file (cap in brackets):
  kanban-dispatch 3/3 [3/3], detail 2/0 [12/0], auto-semantics 36/2 [45/2], gtd 43/9 [80/24], manager-todo 6/4
  [12/12], foreman 5/3 [12/12], `catalog.yaml` 3/3 [3/3], manager-todo TOML 6/4 [12/12]; each mirror equals its live
  numstat. `go run ./internal/template/scripts/gen-catalog-hashes.go --all` ran after the last template edit (exit 0;
  catalog staged in the same commit). Guards: `TestAutoPickDocDoctrine`, `TestAutoPickMirrorParity`,
  `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity` → 4 top-level PASS, 165 `=== RUN` lines, 0 FAIL, `ok`;
  `go test ./internal/template -run '^(TestManifestHashFormat|TestCatalogHashCoversSkillSubfiles)$' -count=1 -v` → 2
  PASS, `ok`; `go test ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -count=1 -v`
  → PASS, `ok`; `go test ./internal/template -run '^(TestDeclaredRuleMirrorForks|TestRuleTemplateMirrorDrift)$' -count=1 -v`
  → 2 PASS, `ok`.
- **Gaps and residual risk of this repair.** The R3 race test is timing-assisted: it waits 400 ms for the compensation
  to reach the queue lock before landing the lease, so on a heavily loaded machine it could pass on the OLD code (it
  cannot fail on the new code, whose in-lock read is independent of the wait); the RED above was observed 3 of 3. The
  `-count=30` evidence is for `TestFactoryNextNominateConcurrentLanes` only; the other tests were judged by
  one run plus the `-race -count=5` above. Not re-measured: `moai spec lint`, the docs-site build, the heading-parity
  ratchet, package-wide `internal/cli` coverage. The F8 base-ref guards were not re-run.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-02T17:06:53Z   # UTC; the sync commit date is 2026-10-03 local
sync_commit_sha: pending-backfill        # a commit cannot cite its own hash; backfilled in a following commit
sync_status: docs-synced                 # docs, CHANGELOG and the corrected wording are committed; the audit disposition below is the operator's, not an audit result
card: t1448
tier: M
tree_head_at_sync_start: 842369c77       # branch WT-todo-auto-pick-autonomy, clean at start
re_close_complete_at: 2026-10-02T18:07:17Z   # UTC; sync re-close after sync-audit iteration 1 repairs; local date 2026-10-03
tree_head_at_re_close_start: 9a5cb0dcb   # branch WT-todo-auto-pick-autonomy, clean at start of the re-close
sync_audit_iteration_1: "FAIL 77.2/100 (cut 85) at audited commit 8de769d81; claude fail, codex fail (3 P1), glm inconclusive (HTTP 401); report .moai/reports/t1448/sync-audit.md (local, uncommitted)"
sync_audit_iteration_2: "delta audit of 01ccc8b33: four dimensions 88/90/85/85, harmonic mean 87.0/100 PASS-WITH-DEBT, no must-fix; cross-model overall fail (claude pass, codex fail, glm inconclusive HTTP 401); report .moai/reports/t1448/sync-audit-iter2.md (local, uncommitted) keeps its first verdict line FAIL"
sync_audit_status: "iteration 2 four-dimension PASS-WITH-DEBT 87.0, required cross-model gate unmet (codex fail; audit server predates the plan surface), accepted as residual risk by operator decision 2026-10-03 - not a passed audit"
post_audit_correction: "86281f836 (13 files, no re-audit, operator-permitted); the audit did not judge it"
reclose_commits: "01ccc8b33 (re-close after iteration 1 repairs); the disposition commit that records this block (backfills sync_commit_sha in the commit after it)"
b12_self_test_a: "grep -c 'SPEC-TODO-AUTO-PICK-001' CHANGELOG.md -> 0 before emission (exit 1), proceed"
b12_self_test_b: "ac_source=.moai/specs/SPEC-TODO-AUTO-PICK-001/acceptance.md tier=M; canonical counter -> 14 (live=14 excluded=0 ambiguous=0, exit 0); live requirements REQ-TAU-001..016 = 16; CHANGELOG entry states 16 / 14"
b12_self_test_c: "every path cited in the CHANGELOG entry verified with ls (exit 0): 4 changed Go files, 3 test files, 7 doctrine files, the Codex TOML, catalog.yaml, progress.md, 4 factory-mode pages; the 9 cited commit SHAs read from git log"
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added, first bullet"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (merged close on the single sync commit); updated: 2026-10-03"
  plan_md_acceptance_md: "no frontmatter block (stateless artifacts) - nothing to transition"
  progress_md: "no frontmatter block; this section is the record"
canary_compliance_check: "not applicable - this SPEC defines no forward-looking policy that its own sync tests"
```

Sync-phase notes (what a reader needs; every figure below was measured in this run on HEAD `842369c77`):

- **User-facing documentation hits judged line by line** (fresh grep across README.md/.ko/.ja/.zh and
  `docs-site/content/{en,ko,ja,zh}`; doctrine judged against `.claude/skills/moai/workflows/gtd.md` `--auto`
  section and `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board):
  - **EDITED** — `docs-site/content/{ko,en,ja,zh}/advanced/factory-mode.md` line 65: "the actor that picks a card
    is always the operator" is no longer true, because a `/moai:todo --auto` invocation lets the invoked session
    take cards on its own judgment through the lease. Rewritten at sentence level in all four locales (ko
    canonical, en/ja/zh derived in the same commit): the operator picks in person (`moai todo next <n>`) or in
    advance (`/moai:todo --auto`); the invoked session takes cards only through `moai factory next --card <id>`
    and never a card the operator held or parked; queue admission stays the operator's; the factory leader still
    does not scan the queue. No heading added or removed (9 per locale).
  - **KEPT** — `README{,.ko,.ja,.zh}.md:161` (bare `/loop` foreman "picking them stay the operator's job"): a bare
    `/loop` is not an `--auto` invocation, so the foreman never picks; still true.
  - **KEPT** — `docs-site/content/{en,ko,ja,zh}/advanced/kanban-mode.md:287` ("picking the next one remain the
    operator's acts. The foreman only moves an already-picked card"): same foreman-boundary reason; the amended
    foreman skill still dispatches only `picked` cards outside a batch authorization.
  - **KEPT** — `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md:73` (mermaid label "the operator picks
    the card"): the diagram depicts the leader routing operator-picked cards to a free lane, which is unchanged;
    the lane lease path is not drawn there.
  - **KEPT** — `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md:117` ("never picks, only routes" for the
    foreman) and `:115` (link blurb "the operator is the one who picks" for `/moai todo`): the foreman boundary is
    unchanged; the blurb is a shorthand for the operator-picks-in-person path and the `--auto` nuance now sits in
    line 65 of the same page.
  - **KEPT** — `docs-site/content/{en,ko,ja,zh}/cli-reference/launchers.md:35` ("the leader deals the cards the
    operator picks to free lanes"): describes leader routing of operator-picked cards; unchanged.
  - **KEPT** — `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-todo.md` (lines 57, 140-146: the pick is
    made by a human through the leader session; approving several cards at once): documents the interactive
    leader-session path and the leader-admits-in-approved-order batch, whose source paragraph in `gtd.md`
    ("Picking the next card") is not in this card's diff; the page does not describe `--auto`.
  - **KEPT** — `docs-site/content/{en,ko,ja,zh}/advanced/agent-guide.md:93` and `README{,.ko,.ja,.zh}.md` manager-todo
    row (`/moai:todo --auto` serial cycle): still true for a non-lane session; the lane refusal is not stated there.
  - **KEPT** — `docs-site/content/{ko,zh}/advanced/manager-lead.md:35` (factory leader hands the operator-picked
    card to an idle lane) and `docs-site/content/en/core-concepts/kanban-board-terms.md:66` (an illustrative walk-through
    sentence unrelated to `--auto`): unchanged by this card.
  - No page in the four locales lists `moai factory next` flags or describes `moai todo --auto` beyond the
    manager-todo rows above (grep for `factory next`, `moai factory`, `fallback declare`, `self-service`,
    `self-dispatch` found nothing), so the new `--card` flag and the lane refusal have no other documented fact to
    update.
- **Quality reads (read-only):** `moai spec lint .moai/specs/SPEC-TODO-AUTO-PICK-001` exit 0, "No findings";
  with `--strict` exit 0, "No findings". Judging build: `moai-adk v3.2.0-rc.26`, `archive/t1401-293-g45600e4ee`,
  built 2026-10-02T14:42:32Z; `git merge-base --is-ancestor 45600e4ee HEAD` exit 0 (a strict ancestor), and
  `git diff --name-only 45600e4ee HEAD -- internal/spec` lists 0 files, so the lint rules are identical to the
  tree's. Docs recipe: `hugo --minify --gc --source docs-site --destination <scratchpad>` exit 0, 0 `WARN`/`ERROR`
  lines, `sitemap.xml` present (hugo found at `/opt/homebrew/bin/hugo`); URL-blacklist grep exit 1 (no match);
  Mermaid `LR`/`RL` grep exit 1 (no match); the four edited pages exist in all locales with 9 headings each;
  README `^## ` counts 12/12/12/12; the tree-wide per-page heading-count ratchet printed no NEW divergence (51
  baselined pages, none of them factory-mode); body-emoji scan of the four edited pages printed nothing.
- **Drift audit after the transition:** MCP `spec_audit` (filter `SPEC-TODO-AUTO-PICK-001`, `project_root` this
  worktree) returned `modern_era_clean: 1`, `total_specs: 1`, one `EraAutoDetected` INFO finding
  (`H-4 (§E.2 + §E.4 + sync_commit_sha)`), no drift finding; `moai spec lint` re-run after the `status: completed`
  edit exits 0 plain and `--strict`.
- **MX check (read-only, nothing edited):** the four changed Go files carry the new tags `@MX:NOTE` on
  `factoryKeepSetRefusal` and `factoryNextNominate` (with `@MX:SPEC`), `[AUTO]` prefixed and in English;
  `factory_card.go` holds 3 `@MX:ANCHOR` (the per-file limit, as its own note states) and 3 `@MX:NOTE` (limit 10).
  Non-test caller counts: `factoryNextNominate` 2, `factoryKeepSetRefusal` 2, `factoryQueuedHoldMarked` 2,
  `factorySerialInFlightExcluding` 2, `todoLaneSession` 1 — none reaches the fan-in >= 3 anchor threshold; no
  exported function and no goroutine was added. Complexity: `gocyclo` is not installed, so no cyclomatic figure was
  measured; a keyword count (`if`/`for`/`case`/`&&`/`||`) over `factoryNextNominate` plus the head of its neighbour
  gives 15, an over-wide upper bound — reported as a review item, not an established WARN trigger.

Gaps (explicitly not observed in this sync run):

- (First sync, as written at `8de769d81`; superseded by the re-close record below.) The independent sync audit had
  not run at that point; nothing here claimed a verdict. It has since run once and FAILED (see the re-close record).
- No Go test, build or lint was re-run in this sync phase (documentation-only scope); the run-phase results in §E.2
  are cited, not re-measured here.
- Cyclomatic complexity of the new functions was not measured (`gocyclo` is absent).
- The `--minify` hugo build was measured on the four-locale site as a whole; per-page rendering of the edited
  paragraph was not inspected visually.
- The `§6` version-string sync of the verify recipe was not re-run: this card touches no version display.
- Native-idiom review of the three derived-locale sentences was by the author only; the `moai-domain-humanize`
  pass was not run (single-sentence edit).

Residual risk: the KEPT judgments rest on reading the doctrine and the page context; a reader who takes the
line-115 shorthand blurb or the diagram label as a statement about `--auto` would still read them as the operator
being the only picker. The decision record a lane writes under the new doctrine has not yet been observed (AC-TAU-009
is a regression-guard with no executing party, per §E.2).

### Sync re-close after sync-audit iteration 1 (this record asserts NO audit verdict for the repaired tree)

`sync_status` stays `docs-synced-audit-pending`. The independent sync audit ran once on `8de769d81` and returned
**FAIL 77.2/100** against the 85 cut (Functionality 65 must-pass, Security 90, Craft 78, Consistency 80; cross-model
gate: claude fail, codex fail with three P1 reports, glm inconclusive — HTTP 401). Report: `.moai/reports/t1448/sync-audit.md`
(local, uncommitted). The repairs landed as `e55aaeb1b` (code and tests) and `9a5cb0dcb` (doctrine wording); their
run-side evidence is in §E.2 'sync-audit iteration 1 repair' (owned by the run phase, not restated here). The audit has
NOT been re-run on the repaired tree, so no PASS, no FAIL and no score is claimed for it; the iteration-2 audit and the
leader's disposition of F3 (a)/(c) are outstanding.

Finding dispositions (ids from the audit report; one line each):

| Id | Disposition |
|---|---|
| F1 (blocking) | Fixed in `e55aaeb1b`: the AC-TAU-002 concurrency test is deterministic (per-card worktree repositories through a stubbed creator). Builder-measured 30/30 and 20/20 under `-race -count=5`; orchestrator-measured 20/20 and 12/12 under `-race -count=3`, 0 DATA RACE. Not independently re-audited |
| F2 | Fixed: doctrine scoped to a lane session (`9a5cb0dcb`: stub, `gtd.md`, `auto-semantics.md` §9.2/§9.3, live and mirror, catalog) and user docs scoped in this re-close (`docs-site/content/{ko,en,ja,zh}/advanced/factory-mode.md` line 65) |
| F3 (b) | Fixed in `e55aaeb1b`: compensation re-reads the factory record inside the queue lock |
| F3 (a), (c) | NOT fixed — accepted residual risk (disposition: operator decision 2026-10-03, recorded under 'Sync-audit iteration 2' below): (a) two different serial cards nominated at once can both lease; (c) an operator `unpick` then `hold` between promotion and claim can leave a held card leased. Both are two-store non-atomicity windows, the class of the pre-existing unnominated arm (c) |
| F4 | Tests added in `e55aaeb1b` for mutants M1 (blank `--card`, CLI and MCP), M2 (`picked` with owner) and M5 (`picked` plus `blocked`); M3, M4, M6 judged lower value and not pursued |
| F5 | Reworded in `e55aaeb1b` (the `blocked` refusal detail) |
| F6 | Follow-up owed: `kanban-dispatch-detail.md` is over the 40,000-character budget (43,138 characters now, 42,675 on develop) — split card |
| F7 | Holds owed: t810 (`picked`), t1294 and t1383 (`queued`) carry neither the `hold` state nor a leading `[보류` — operator or leader action; this card does not touch the queue |
| F8 | Not attributable to this card (base-ref contract-mode guards fail on files the card did not touch; 0 added lines contain "kickoff") |
| F9 | Naming nit (`factoryToken*` in the record vs `factoryRefuse*` plus one `factoryTokenForeignWorktree` in code); trailer disclosure already recorded; the AC-TAU-011 re-pin without re-audit is disclosed in the plan artifacts — no action in this re-close |
| F10 | Low security note, not changed: the `unknown-card` refusal echoes the unvalidated `--card` text and `foreign-worktree` prints an absolute landing path; both reach only the caller |
| F11 | Help text: `moai todo --help` `--auto` still says the batch approval is 'of the queue and nothing else' — follow-up |
| F12 | CHANGELOG corrected in place (no second entry): AC-TAU-002 no longer stated as unqualified PASS, the audit FAIL and repairs recorded, stub bytes 28,308 → 28,304, limitations group added |

Follow-up card candidates (updated after iteration 2; none created by this docs commit — the leader issues them, and
the CHANGELOG says only 'a follow-up card is owed'): (1) an atomic lease across the queue and the factory record,
closing F3 (a), F3 (c) and F14 (the operator-accepted windows; a promotion token in the compensation is the cheap
part, an atomic serial slot plus a final queue re-read held across the claim is the rest); (2) the pre-existing
`git branch -m` reflog-temp-file collision when two lanes lease at the same instant in one repository
(`factoryEnsureCardWorktree`; surfaced by the new concurrency test; the verb errors after the lease succeeded) —
carried by the same follow-up card per the operator decision; (3) split `kanban-dispatch-detail.md` under the
40,000-character budget (card issued separately by the leader); (4) the `moai todo --help` `--auto` text (F11);
(5) the F10 low security note (unvalidated `--card` text echoed on `unknown-card`; absolute landing path in
`foreign-worktree`); (6) decide the live-only `moai worktree sweep …` sentence of `kanban-dispatch.md`; (7) holds for
t810, t1294 and t1383 (operator or leader action, see the merge-report line below).

`sync_commit_sha` stays `pending-backfill` in the disposition commit (a commit cannot cite its own hash) and is
backfilled in the commit after it.

Measured in this re-close (tree HEAD `9a5cb0dcb`, plain commands, output to the scratchpad where long):

- User docs: the four `factory-mode.md` line-65 sentences rewritten at sentence level, one line changed per file
  (`git diff --stat` 4 files, 4 insertions, 4 deletions before the CHANGELOG edit). URL-blacklist grep over
  `docs-site/content` and `docs-site/hugo.toml` exit 1 (no match); Mermaid `LR`/`RL` grep exit 1; heading parity
  `^## ` 8/8/8/8 and `^#{1,6} ` 14/14/14/14 across ko/en/ja/zh (the first sync's '9 per locale' used a filter not
  reproduced here; parity holds); body-emoji scan exit 1 (no match); `hugo --minify --gc --source docs-site
  --destination <scratchpad>` exit 0, 0 WARN/ERROR lines, `sitemap.xml` present.
- CHANGELOG (B12): `grep -c 'SPEC-TODO-AUTO-PICK-001' CHANGELOG.md` was 1 before editing, edited in place, still 1
  after; canonical counter on `.moai/specs/SPEC-TODO-AUTO-PICK-001/acceptance.md` (tier M) `live=14 excluded=0
  ambiguous=0`, exit 0; live requirements `REQ-TAU-001..016` = 16; every path and every commit SHA cited in the entry
  verified with `ls` / `git cat-file -e` (exit 0).
- Quality reads (read-only): `moai spec lint .moai/specs/SPEC-TODO-AUTO-PICK-001` exit 0 and with `--strict` exit 0,
  both "No findings". Judging build `moai-adk v3.2.0-rc.26`, `archive/t1401-293-g45600e4ee`, built
  2026-10-02T14:42:32Z; `git merge-base --is-ancestor 45600e4ee HEAD` exit 0 (a strict ancestor) and
  `git diff --name-only 45600e4ee HEAD -- internal/spec` lists 0 files, so its lint rules equal the tree's. MCP
  `spec_audit` (filter `SPEC-TODO-AUTO-PICK-001`, `project_root` this worktree): `modern_era_clean: 1`, one
  `EraAutoDetected` INFO finding (`H-4 (§E.2 + §E.4 + sync_commit_sha)`), no drift finding.

Gaps of the re-close (explicitly not observed): the sync audit was not re-run, so nothing here is a verdict; no Go
test, build or lint was run in this documentation pass — the figures for the repaired code are builder- and
orchestrator-measured and cited, not re-measured; the derived-locale sentences (en/ja/zh) were checked by the author
only and the `moai-domain-humanize` pass was not run (single-sentence edit); pre-existing `**term (note)**` emphasis
spans on lines 90-91 of the en/ja/ko pages (parenthetical inside the markers) were found by the emphasis scan and left
untouched, being outside this re-close's scope.

### Sync-audit iteration 2 and the operator's disposition (no passed audit is claimed)

운영자 결정 2026-10-03: sync-audit iter2 네 차원 87.0 PASS-WITH-DEBT, 필수 codex fail(F3 a/c·F14)은 잔여 위험으로 수용, 후속 카드로 종결
English gloss: operator decision of 2026-10-03 — the iteration 2 sync audit scored the four dimensions 87.0
PASS-WITH-DEBT; the required codex failure (F3 a/c and F14) is accepted as residual risk and closed by a follow-up card.

This decision is a **disposition, not an audit result**. The audit's own report, `.moai/reports/t1448/sync-audit-iter2.md`
(local, uncommitted), keeps its first verdict line `FAIL`; nothing in this record rewrites it, and no audit has passed.

Iteration 2 (delta audit, audited commit `01ccc8b33`; judged this tree's Go sources through the toolchain, the
fan-out through the `d194083fb` MCP server, a strict ancestor of the tree):

- Four dimensions: Functionality 88, Security 90, Craft 85, Consistency 85 → harmonic mean 87.0/100 against the 85 cut
  (margin +2.0), **PASS-WITH-DEBT**, no must-fix finding, both must-pass dimensions above their thresholds.
- Cross-model (`audit_multi`): overall `fail`; claude pass (required, the auditor's own verdict), codex fail (required),
  glm inconclusive (advisory, HTTP 401). No receipt was issued.
- The required gate was unmet for **two independent reasons**: (1) the codex backend returned fail — F3 (a) and F3 (c)
  again, plus the new F14; (2) the orchestrator's `moai verify audit-plan --result-file` check read
  `convergence_check.ok=false`, reason `plan_source: absent or null`, because the MCP server that produced the
  `audit_multi` result (build `d194083fb`) predates the plan surface.
- Dispositions that changed since iteration 1 (F1..F17):
  - F1 **resolved** — `TestFactoryNextNominateConcurrentLanes` 20/20 PASS and 12/12 under `-race -count=3`, 0 DATA RACE
    (delta-auditor measurement; the earlier builder and orchestrator counts are in the iteration-1 record).
  - F3 (b) **resolved** — RED observed against the pre-repair file, GREEN on the repaired tree.
  - F3 (a), F3 (c) and F14 — **accepted as residual risk by the operator decision above**; F14 (new, codex P2): the
    compensation after a failed nominated claim cannot tell its own promotion from an operator's fresh pick of the same
    card, so unpick, pick and a failing claim within milliseconds can revert the operator's pick; the auditor
    reproduced it with direct queue writes. F14 is now also in the CHANGELOG limitations.
  - F4 — four of the six mutants are killed; M3 (the set of refusals `--wait` waits through) survives; M4 (the in-lock
    re-validation) cannot be forced.
  - F13 (keep-set wording residue of F2) and F15 (the `blocked` detail over-claims on the bare path) — **corrected in
    `86281f836`**, not re-audited (see the next record).
  - F16 (internal audit ids in the CHANGELOG) — **fixed**: the ids are removed from the entry in this docs commit.
  - F2 was partial at iteration 2 (the lease claim scoped, the keep-set clause not); its remainder is F13, corrected in
    `86281f836`. F5 resolved, F12 resolved, F17 (400 ms sleep in a regression test) is a note with no action.
  - F6, F7, F8, F10, F11 — unchanged (F9 unchanged as well; F6/F7 re-observed).

For the merge report, one line: the SPEC identifies the keep-set by (a) the structural `hold` state, written only by
`moai gtd hold`, and (b) a card body that opens with the `[보류` marker — a lane may write neither; t810 (`picked`),
t1294 and t1383 (`queued`) currently carry neither, so the operator or leader must apply one before lanes exercise the
doctrine.

### Post-audit correction `86281f836` (no re-audit, operator-permitted)

After iteration 2 the operator permitted a low-cost correction without a further audit. `86281f836` (one commit, 13
files, 47 insertions, 30 deletions): `.claude/skills/moai/workflows/gtd.md`, `.claude/rules/moai/workflow/kanban-dispatch.md`,
`.claude/rules/moai/workflow/auto-semantics.md`, `.claude/agents/moai/manager-todo.md`,
`.claude/skills/moai-kanban-foreman/SKILL.md`, their five `internal/template/templates/` mirrors, the regenerated
`internal/template/templates/.codex/agents/moai/manager-todo.toml`, `internal/template/catalog.yaml`, and the one-line
refusal string in `internal/cli/factory_card.go`. It scopes the keep-set and lease wording to a lane session and names
the operator session's serial cycle as the contrast (F13), and rewords the `blocked` refusal detail to 'the nominated
lease refuses it, the operator decides it' (F15). No sync audit judged this tree; this record asserts none. After it
the always-loaded stub `kanban-dispatch.md` measures 28301 B / 28092 chars live and 27979 B / 27771 chars mirror (this
documentation pass, `wc -c` / `wc -m`; merge-base blob 28308 / 28099 live, 27986 / 27778 mirror).

This docs commit (the disposition commit) touches only `CHANGELOG.md` (the existing entry, edited in place) and this
§E.4; the commit after it only backfills `sync_commit_sha`. No Go code, `.claude/` rule, template, catalog, spec/plan/
acceptance body or untracked report is touched by either.

Measured in this pass (tree HEAD `86281f836`): B12 counter on `acceptance.md` `live=14 excluded=0 ambiguous=0`
(exit 0), live requirements `REQ-TAU-001..016` = 16, `grep -c 'SPEC-TODO-AUTO-PICK-001' CHANGELOG.md` = 1 before and
after the edit, every cited commit SHA passes `git cat-file -e`. Gaps: no Go test, build, lint or audit was run in
this documentation pass; the repaired-code figures are cited from the audit and the earlier records, not re-measured.
