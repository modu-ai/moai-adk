auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 1e1d0cc84e6cd137592b8f0a9b7a0ff014f7791a

# SPEC Review Report: SPEC-MERGE-WINDOW-QUEUE-001 (card t1479)
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.75 (Tier L threshold 0.85)
Plan Artifact Hash (per-file sha256, subject set acceptance/design/plan/research/spec):
- acceptance.md 792552914aaffd61492fd7b8929b8fe06b1819f811c2e45ddd8b982abfe85945
- design.md     9381d301acb50defdd109fcf0c851e58c4eb0eecf0c0ee9e780b7be63c4a65b5
- plan.md       44f5e99f4262d4dc413d6dc041b28b8ce6110109903e8ad2ce2d4999a0bde3b6
- research.md   56d7063e63e7e034b7252a4a8ea12e1b15c8232048fe05869eaef24ddda0f752
- spec.md       a0a20e01c8b05f5a723756ab53facc400290fb5cdac25aab6fd315ceb03c8dcc
Auditor Version: plan-auditor/v1 (iteration 1, full audit)

Reasoning context ignored per M1 Context Isolation. The operator and leader decisions (Q1 FIFO queue and nomination abolished; Q2-Q6 LEADER-DECIDED) are taken as fixed scope. No finding below re-opens them. Where a finding touches Q4 or Q6, it says only that the REQ text as written does not yet deliver the property the decision row claims.

## Must-Pass Results

- [FAIL] MP-1 REQ number consistency. spec.md defines 32 REQs (`grep -c "^- \*\*REQ-MWQ-"` → 32) in bands: 001-009, 009a, 010-013, 020-024, 021a, 021b, 030-033, 040-044, 050-051. Two things break the strict sequential, no-gap rule: the gaps 014-019, 025-029, 034-039 and 045-049, and the lettered inserts (009a, 021a, 021b). `moai spec lint` reports no findings, but lint does not check sequencing, so its silence corroborates nothing on this axis. The fix overlaps D1, which needs a renumber anyway.
- [PASS] MP-2 GEARS compliance, judged on the requirement layer (spec.md §C) only. Each REQ uses a GEARS pattern or its legacy equivalent:
  - Event-driven **When**: 002-005, 012, 021b, 023, 024, 030, 040, 042, 044.
  - **While**: 011, 022, 041.
  - **Where**: 007, 009a, 021a.
  - Ubiquitous "shall": 001, 006, 009, 010, 013, 020, 021, 032, 033, 050.
  - Legacy Unwanted "shall not": 008, 031, 043, 051. These fall inside the 2026-11-22 compatibility window.
  - The Given-When-Then ACs were not graded here.
- [PASS] MP-3 frontmatter, spec.md:L2-L13. All 12 canonical fields are present, with `version: "0.2.0"` quoted, `status: draft`, `lifecycle: spec-anchored` and `tags` as a string. No rejected alias is used. The extra fields `tier`, `card` and `related_specs` are additive.
- [PASS] MP-4 language neutrality. REQ-MWQ-021b names `go test -json` as the one recognized report "in this repository". It gives a generic fallback (command plus exit code) for every other runner, and decision Q6 justifies the choice. No language is made primary in template text, and REQ-MWQ-050 requires the distributed text to stay neutral.
- [PASS] MP-5 D7. Eight related SPECs were checked, and every one has `status: completed` (grep of each spec.md). None is retired, superseded or archived, so no BLOCKING finding.
- [PASS] MP-6 D8. `grep syscall` across the SPEC directory returns no match, so D8 passes automatically.
- [PASS] MP-7. `grep -rn '[NEEDS CLARIFICATION'` over plan.md and research.md returns no match.
- [N/A] MP-8. No AC is classified release-blocking (`grep -i release-blocking` → no match), so MP-8 does not apply. As diligence I re-ran the RED cells these ACs cite; see Evidence.
- [PASS] MP-9. I applied the CN-4 verb by hand because `awk -f` was refused (see Gaps).
  - plan.md has milestones M0-M9 and no `Exit:` lines, so there are zero exit bindings.
  - Ordering candidates in acceptance.md: L14, L26, L35, L38, L53, L85-86, L118, L126, L152.
  - None of them binds a milestone that the plan schedules on the forbidden side, so there is no CONFLICT.
  - L35 (the AC-MWQ-007 "pre-change baseline captured on the plan tree") is a capture gap, not an ordering conflict. It is filed as D9.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | 0.50 | Several REQs allow divergent implementations: REQ-024 reservation semantics (spec.md:L146-150), REQ-011 against REQ-003/004 (L76-83, L109-111), REQ-005 against a cancelled waiter (L84-87), and REQ-041's "most recent completed" read (L176-178). |
| Completeness | 0.75 | 0.75 | All sections and §E Out of Scope H3s are present (L211-229). M0 (plan.md:L43-47) has no AC. The REQ/AC budget is exceeded (D1). |
| Testability | 0.75 | 0.75 | AC-013's "promotion-as-authorization sentence" grep names no pattern (acceptance.md:L61). The AC-007 baseline is never captured (L34-36). |
| Traceability | 1.00 | 1.00 | 32 REQ definitions, and 35 `maps REQ-MWQ-*` tags covering all 32. No uncovered REQ, no orphan (hand comparison of the two grep lists). |

Aggregate: (0.50 + 0.75 + 0.75 + 1.00) / 4 = 0.75, below the Tier L threshold of 0.85.

## Defects Found

D1. BUDGET — spec.md:L63-196, acceptance.md:L6-122 — The Tier L ceiling is 25 REQs and 25 ACs, applied independently (spec-workflow.md § SPEC Complexity Tier, L144-150). This SPEC carries 32 REQs and 35 ACs, so both ceilings are breached. Tier L is the top tier, so tiering up is not available; the rule says to split, not to relax the budget. — Severity: major — Class: blocking — Required fix: consolidate and/or split. One route reaches ≤25 on both counts:
  - Move the push verb (REQ-040..044 and its ACs, which are self-contained) into a sibling SPEC.
  - Fold 009a into 009 and 021a/021b into 021, giving 24 REQs.
  - Fold AC 002a, 003a, 005a, 009a, 021a and 021b into their parent ACs as extra Given/When/Then scenarios, giving 24 ACs.
  - Renumber sequentially, which also resolves D2.

D2. MP-1 — spec.md:L67-196 — Banded numbering leaves gaps (014-019, 025-029, 034-039, 045-049) and uses lettered inserts. — Severity: critical (must-pass) — Class: blocking — Required fix: renumber REQ-MWQ-001..N with no gaps and no letters, and update every AC `maps` tag, the acceptance.md §F table, the plan.md milestone AC lists and the decision-index "반영" lines in the same pass.

D3. QUEUE-HOLD CONTRADICTION — spec.md:L76-83 against L109-111, and design.md:L15-17 — REQ-003 and REQ-004 promote the first live ticket on release or staleness. REQ-011 says that while the policy is `hold`, acquire grants no new holder. Design D1 says every published record with a non-empty queue has a holder. A release under `hold` with a non-empty queue therefore cannot satisfy all three; codex independently reported the same contradiction. AC-011 tests a free window only and never tests release-during-hold. — Severity: major — Class: blocking — Required fix:
  - State that promotion (release, staleness and lease expiry alike) is suspended while the policy is `hold`.
  - Allow a holder-less queue while the policy is `hold`.
  - State that the switch to `open` promotes the head on the next observation.
  - Add an AC where A releases under `hold` with B queued: B is not promoted, and becomes holder after `open`.

D4. REQUEUE RESERVATION DEADLOCK AND STARVATION (focus 1) — spec.md:L146-150, acceptance.md:L84-86, decision-index Q4 — The REQ does not say whether a reserved front ticket can be promoted while its owner is still re-absorbing and re-measuring.
  - **If it can be promoted** (promotion checks only owning-session liveness, REQ-003): A becomes holder mid-re-measure and blocks the queue for the full re-measure, which is the very in-window wait the SPEC removes. Codex reproduced this with a rule model.
  - **If it cannot**: nothing bounds a reservation whose live owner never re-enters, so the queue head is blocked forever.
  - **If the reservation does not block others**: lanes merge during A's re-measure, A's re-entry finds develop moved, and A goes to the tail. A tail ticket is then invalidated by every merge ahead of it, so under sustained arrivals it never merges.
  - In none of the three readings does the Q4 claim "기아를 막으면서" hold.
  - AC-024 also asserts "develop moves again before A merges", which is only reachable in the third reading.
  - — Severity: critical — Class: blocking — Required fix (within the Q4 decision):
    - Define the reservation as its own state, separate from a promotable ticket.
    - State whether the queue pauses for it.
    - Add a bound: a reservation expires to the tail after a stated duration, or after its owner's lease-equivalent.
    - State the rule that guarantees eventual merge for a tail ticket. One option: a ticket requeued twice gets a pause-reservation.
    - Add ACs for "A never re-enters" and "B is or is not promoted during A's re-measure", and make AC-024 consistent with whichever semantics is chosen.

D5. ORPHANED WAITER TICKETS (focus 1) — design.md:L12, L32-34, spec.md:L84-87 — Ticket liveness follows the owning session's pid (`pid_source: session-owner`), not the waiting `acquire --wait` process.
  - A waiter killed by SIGTERM or SIGKILL leaves a ticket that reads as live. The Bash tool caps a call at 600,000 ms, below the 60-minute default bound, so this is the expected case, not a corner. Codex measured that the owner process survives the waiter's SIGTERM.
  - When the orphan is promoted, nobody merges. The window is blocked for the 30-minute lease, or forever when the lease is 0 (REQ-009a).
  - Design D2's claim "A waiter whose process died ... next observer treats it as stale" is false under session-owner pids.
  - REQ-005 also leaves the race between promotion and bound-elapse unspecified: a waiter promoted just as its bound elapses could exit non-zero while recorded as holder.
  - — Severity: major — Class: blocking — Required fix:
    - Record waiter-process liveness, such as a waiter pid or a heartbeat on the ticket, and require it for promotion; or require that re-invocation is the only way to claim a promotion within a short claim deadline.
    - Specify that on bound-elapse, a waiter found promoted inside the same mutation returns success (or releases).
    - Specify how a lane invokes a wait longer than its tool's command limit, for example a background run or repeated bounded waits that keep the ticket under REQ-008.
    - Add ACs for a killed waiter and for promotion racing the bound.

D6. HOLLOW LOCAL FORM (focus 4) — spec.md:L134-138, acceptance.md:L74-77 — `moai integration remeasure -- true` produces exit 0 and no recognized report, which is valid under REQ-021b. `go test ./...` without `-json`, or with a selector that matches zero tests (`ok ... [no tests to run]`), is also valid. The completion gate (REQ-030/032) can therefore be met by a command that tests nothing, which is the empty-sweep defect of verification-completeness.md §1.1. — Severity: major — Class: blocking — Required fix, within Q6 (count only when the tool reports one):
  - Treat the runner's own empty-sweep tokens as a tool-reported zero, and therefore invalid. Examples are `no tests to run`, `[no test files]`, pytest `no tests ran`, jest `No tests found` and a cargo 0-pass line. The marker list stays per-runner and neutral.
  - Surface the recorded command verbatim in `merge ready` and `status`.
  - Optionally require the command to match a configured project verification command.
  - Add AC scenarios for `true` (no report) and for `go test -run NONE` (empty sweep), stating the expected disposition of each.

D7. DIRTY-TREE MEASUREMENT KEYED AS HEAD (focus 4) — spec.md:L126-130, design.md:L39-40, acceptance.md:L65-66 — The verb runs in the caller's working tree but keys the record by `HEAD^{tree}`. No clean-tree requirement exists, and nothing checks that HEAD is unchanged during the run. Uncommitted edits can make the measured tree differ from the keyed tree; codex reproduced exit 0 on the edited copy and exit 1 on the keyed commit. — Severity: major — Class: blocking — Required fix: require an empty `git status --porcelain` (tracked and untracked) before and after the run, plus an unchanged HEAD, or run in an isolated checkout of the exact candidate commit. Add an AC where a dirty tree produces a refusal or an invalid record.

D8. PUSH VERB (focus 5) — spec.md:L171-187, design.md:L77-84:
  - (a) Red-tip deadlock. REQ-041 refuses while the remote tip is red, and the only usual way to turn the tip green is to push a fix. The verb therefore refuses exactly the push that repairs develop. gitflow-lane-protocol.md:L87 says "수리되어 green으로 돌아온 뒤에야" but does not say how the repair lands.
  - (b) Wrong run read. The reused reader `gh run list --commit <head> --limit 1` returns the newest run, which may still be in progress while an earlier completed run for the same tip failed. That reads as "no completed run" and bypasses REQ-041. Codex reproduced this.
  - (c) Moving branch name. Step 1 checks the window, then step 5 pushes the moving name `develop`. A lane can acquire and merge in between, so an unverified SHA gets pushed; codex reproduced this against a bare remote. REQ-043's "not while held" is a point-in-time read.
  - (d) "Leader-only" means only "not lane role". Any session without `MOAI_FACTORY_ROLE=lane` can push. Lanes are stamped by `enterFactoryLaneMode` (internal/cli/factory.go:L798) and by the codex launchers, so lane refusal is real, but the stated leader-only property is wider than the gate.
  - (e) "Concluded in failure" leaves open whether timed_out and startup_failure count. `mapGHConclusion` (ci_verdict.go:L167) maps both to failure.
  - — Severity: major — Class: blocking for (a), (b) and (c); optional for (d) and (e) — Required fix:
    - (a) Add an explicit, recorded repair path, for example `--repair <reason>`, logged and still never force.
    - (b) Specify reading the most recent completed run, and add an AC with an in-progress run above a completed failure.
    - (c) Push the verified SHA by explicit refspec (`<sha>:refs/heads/<b>`), taking the window or a push reservation in the same serialized section for the duration. Add an AC where develop advances between the check and the push.
    - (d) Narrow the text to "lane sessions are refused", or define a leader role marker.
    - (e) Say "mapped Failure class (failure, timed_out, startup_failure)".

D9. AC-007 BASELINE NEVER CAPTURED (focus 2) — acceptance.md:L34-36, research.md §R1 — AC-007 requires stdout, stderr, exit code and record to be byte-identical to a "pre-change baseline captured on the plan tree". research.md §R1 contains no such capture, and plan.md schedules none. The refusal text includes PID and timestamp fields (integration_lock.go:L281-282 `... (pid %d) since %s on %s`), so "byte-identical" needs a fixture. — Severity: minor — Class: blocking — Required fix: add a pre-M3 step, such as M0 or the pre-flight, that captures the baseline from the plan-tree build with a fixed fixture record, and commit it before the M3 change (verification-claim-integrity §2.3). Name the fixture.

D10. AC-013 UNSPECIFIED PATTERN — acceptance.md:L58-61 — "a grep for the promotion-as-authorization sentence returns one match in each file" names no pattern, so the AC is not binary. A mutant can satisfy it with any sentence. — Severity: minor — Class: blocking — Required fix: give the literal grep pattern (Korean in AGENTS.local.md, plus the protocol file), or a fixed token both files must carry.

D11. M0 IS UNMEASURABLE AS WRITTEN — plan.md:L43-47, spec.md:L97-101 — M0 measures the in-window duration "under the new order" before M1-M6 implement that order, and it has no AC. REQ-009 makes lowering the lease default depend on this measurement. — Severity: minor — Class: optional — Required fix: state that M0 times a manual simulation (identity check plus `git merge --no-ff` on a fixture), give the command and the statistic, or move the measurement after M6.

D12. CANDIDATE_CI KEY NOT NAMED (focus 7) — spec.md:L131-133, design.md:L62, plan.md:L59 — The REQ text says "the candidate-CI setting introduced by card t1478" and never names it; only decision-index Q5 says `candidate_ci`. t1478 names it as the "`candidate_ci` capability key" (SPEC-CANDIDATE-CI-001 spec.md:L126, plan.md:L45), and its full config path is not pinned there either. The two SPECs are consistent but not anchored to a shared key path. — Severity: minor — Class: optional — Required fix: cite `candidate_ci` and its full config path in REQ-021a, design D4 and AC-021a, and record a cross-SPEC note that whichever card lands second verifies the path.

D13. RECORD FORGEABILITY UNSTATED — spec.md:L128-130 — "captured by the moai verb ... rather than asserted by the caller" is true of the verb, but the verifier cannot tell a hand-written record file from a verb-written one. — Severity: minor — Class: optional — Required fix: state the threat model, which assumes cooperative actors, as Residual-risk, or bind the record to a provenance field that the verifier checks.

D14. AC-050 WORDING MUTANT — acceptance.md:L118-120 — Rewording "announcement to the lead" to "announcement to the leader" satisfies the AC while the announcement layer survives. — Severity: minor — Class: optional — Required fix: add a positive grep for the new queue/policy sentence.

Template-first and locality (focus 8) are correct as planned. plan.md:L88-92 orders the template edit, then `make build`, then the local mirror. AGENTS.local.md and gitflow-lane-protocol.md stay local. No defect found here.

## Recommendation

FAIL: one must-pass failure (MP-1) and blocking defects D1, D3-D10. Fix in this order:
1. D4 and D5, which carry queue liveness.
2. D3, the hold-versus-promotion semantics.
3. D6 and D7, which carry gate substance.
4. D8 (a), (b) and (c), push safety.
5. D1 and D2: consolidate or split, then renumber.
6. D9 and D10, AC testability.

The optional findings are D8 (d, e) and D11-D14. Iteration 2 is scoped to this defect delta, and the CN-4 verb will be re-run in full.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

**Claim:** RED cells E1, E3, E5, E6, E7, E8 and E10 still reproduce on the audited tree.

**Evidence:** each cell's grep, with an absolute path prefix:
- E5: `255:	if !strings.Contains(strings.ToLower(string(raw)), full[:12]) {`
- E6: `1418: ... factoryWriteMergeRecord(root, ...` and `1616:func factoryWriteMergeRecord`
- E8: line 120 in both the template and the local copy, carrying "The announcement to the lead rides alongside it."
- E7: `221: ... 리더의 창 지명만이 근거다.`
- E1: `0`
- E3: lines 254/274/329/457 (`integration [command]`, `status`, `acquire`, `release`)
- E10: `86:	cmd := exec.Command("git", args...)`
- Counts: `grep -c "^- \*\*REQ-MWQ-" spec.md` → 32; `grep -c "^- \*\*AC-MWQ-" acceptance.md` → 35.
- D7 statuses: all `completed`.
- `moai spec lint <spec.md>` → "✓ No findings".
- `spec_audit` → modern_era_clean 1.
- audit_multi: overall fail, with claude fail (in-session anchor), codex fail (7 findings), glm inconclusive ("z.ai response carried no text content"), participant_count 2.

**Baseline-attribution:**
- Branch `WT-merge-window-queue` tip is `1e1d0cc84e6cd137592b8f0a9b7a0ff014f7791a`, read from the gitdir ref file.
- research.md measured on `d7112d005`. The line numbers re-read on 1e1d0cc84 match.
- spec lint and spec_audit ran on the installed moai binary built at `45600e4ee`. audit_multi reports that build as an ancestor of tree HEAD 1e1d0cc84 (binary lag, verification-claim-integrity §2.2).

**Gaps:**
- The worktree guard refused three things, so this session did not observe them directly:
  - `git -C <tree>`, so I read HEAD via the gitdir ref file instead;
  - `sh <script>` and `awk -f <script>`, so I could not run the AC-4/AC-5 traceability verb or the CN-4 verb mechanically, and applied both by hand with plain greps;
  - the Grep tool, which is unavailable in this session, so I used Bash grep.
- I ran no RED cell that contains a pipe (E2, E4, E9); they are outside the single-invocation form.
- Codex's reproductions (process kill, dirty tree, gh `--limit 1`, push race) are its own observations, not re-executed by me.
- spec lint's "no findings" collected no shown REQ count on the numbering or budget axes, so it is not cited as corroboration.

**Residual-risk:**
- `residual_risk_note`: "required-backend FAIL: claude, codex".
- GLM (advisory) did not participate.
- Tool results come from a lagging binary (45600e4ee).
- No codex `audit_receipt` was returned by the tool.

## Operational Notes (unverified)

- `assumption`: after the D4 fix, measure the expected number of requeues per merge under N waiters with a rule-model test (codex reported having such a model); include it as an AC.

AUDIT-VERDICT: FAIL spec=SPEC-MERGE-WINDOW-QUEUE-001 receipts=none
