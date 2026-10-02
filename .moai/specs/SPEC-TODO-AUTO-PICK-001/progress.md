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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
