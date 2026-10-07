auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 6ab3dbcb2fc5042d5728af62379f8663ccfc5d1e

# SPEC Review Report: SPEC-MERGE-WINDOW-QUEUE-001 (card t1479)
Iteration: 2/3 (delta re-audit + CN-4 in full)
Verdict: FAIL — STOP (aggregate regressed 0.75 → 0.69; see Recommendation)
Overall Score: 0.69 (Tier L threshold 0.85)
Plan Artifact Hash (sha256, tree 6ab3dbcb2):
- acceptance.md     733991a0931992036a2ac2e7d7d779df348ef87466674ec9fd78a0dd7ee5b057
- design.md         822b2f3f60deea4b7c6b7060ab1f649244152a2aa74b32dc74cd6951dca93758
- plan.md           d6241bc55b3f5c442a7376558c602721875c600171b24d21ec571c13dc288afd
- research.md       9f2fbc15c7c1415a657b5a7dfb7f76ee7d89e33f0fcebc47cfe9aefeee5200ec
- spec.md           84d9421c9f79ad6a678ef37d13910f0c7a3fab8adbb2a50868cf2163e105273d
- decision-index.md 722d2fa7fb2d23fe941b0e1b7f4a569e409c137c2b615284a9df0baa4f8712ae
Auditor Version: plan-auditor/v1 (iteration 2)

I ignored the reasoning context, per M1 Context Isolation. The operator and leader decisions Q1-Q16 stand as decided, Q14 values and the Q15/Q16 rules included. None of the findings below reopens a decision. Each one says only that the REQ text, as written, contradicts itself or does not deliver what the decision row claims.

## Must-Pass Results

- [PASS] MP-1 REQ numbering.
  - `grep -c '^- \*\*REQ-MWQ-' spec.md` gives `25`, listing 001 to 025 contiguously with no letters.
  - `grep -c '^- \*\*AC-MWQ-' acceptance.md` gives `25`.
  - Each AC maps to the REQ with the same number: 001:001 through 025:025.
- [PASS] MP-2 GEARS. Judged on the requirement layer (spec.md §C) only.
  - Event-driven **When**: REQ-002..007, 013, 016, 017, 019, 021.
  - **While**: REQ-008, 018, 020. REQ-008 and 020 also carry compound When clauses.
  - **Where**: REQ-011.
  - Ubiquitous: REQ-001, 009, 010, 014, 015, 023, 024.
  - Legacy Unwanted "shall not": REQ-012, 022, 025. These fall inside the compatibility window, which runs to 2026-11-22.
- [PASS] MP-3 frontmatter, spec.md:L2-13. All 12 canonical fields are present. `version: "0.4.0"` is quoted, `status: draft`, and `tags` is a string. No rejected alias is used.
- [PASS] MP-4 language neutrality.
  - REQ-016 (L150-156) names `go test -json` only as "in this repository".
  - The other recognized runners are table rows (design D4).
  - REQ-024 requires the distributed text to stay neutral.
- [PASS] MP-5 D7. I ran the D7 loop with a plain grep.
  - Six referenced SPECs read `status: completed`.
  - `SPEC-CANDIDATE-CI-001 NOT FOUND` in this tree. That is SHOULD-level only: it is a sibling draft, acknowledged in research §R5.
  - No SPEC is retired, superseded or archived.
- [PASS] MP-6 D8. `grep -c syscall spec.md` gives `0`.
- [PASS] MP-7. `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` gives no output (rc=1).
- [N/A] MP-8. `grep -in release-blocking acceptance.md` gives no output (rc=1), so no AC is release-blocking. As diligence I re-ran the RED cells; see Evidence.
- [PASS] MP-9. I applied the CN-4 verb by hand because the guard refused inline awk (see Gaps).
  - COLLECTED: 9 milestones in plan order (M0, M1, …, M8), 0 exit bindings.
  - The ordering candidates in acceptance.md are L17, 20, 27, 37, 49, 53, 62, 65, 113, 118, 137 and 173.
  - Only L53 binds a milestone: "m0-window-duration.md … committed before the M1 code commit". The plan schedules M0 (plan.md:L51) before M1 (L60), so the order is consistent.
  - L173 is "Sync phase closes before the merge into develop", which is not a milestone pair.
  - No CONFLICT.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | 0.50 | Four core liveness and queue REQs contradict each other or leave the core transition undefined: REQ-001/007 holder pid (N1); REQ-004 against REQ-003/020 (N2, N3); REQ-020 readiness against its own requeue counter (N4). An engineer would implement these differently. |
| Completeness | 0.75 | 0.75 | All sections are present, and the §E Out of Scope H3s are at spec.md:L225-244. One clause from the push move (old REQ-042, fail-closed on unreadable CI) is carried by neither SPEC (N5). The fixture README points to a stale AC number (N7). |
| Testability | 0.50 | 0.50 | Under REQ-004 as written, three ACs cannot pass: AC-003 (+119 s re-entry, acceptance.md:L25-26), AC-020 scenario 3 (L113-114) and AC-020 scenario 4 (L114-116). |
| Traceability | 1.00 | 1.00 | 25 REQ definitions and 25 `maps` tags, one-to-one. No uncovered REQ, no orphan (hand comparison; see Gaps). |

Aggregate: (0.50 + 0.75 + 0.50 + 1.00) / 4 = **0.69**. That is below the 0.85 threshold and below iteration 1's 0.75.

## Defects Found (structured defect-list)

N1. HOLDER-PID-ANCHOR — spec.md:L69-73 (REQ-001), L101-105 (REQ-007), design.md:L12.
- The problem: a promoted holder record has no owning-session pid to carry.
  - Iteration 1's REQ-001 ticket carried "its owning-session process id" (`git show 1e1d0cc84:…/spec.md`).
  - v0.3.0 replaced it with the waiter pid and start time. The owner pid was silently dropped.
  - Promotion runs inside another process's mutation, for example A's release. That process cannot resolve B's owner pid.
- Why it matters (integration_lock.go:L172-180, L245-252, L322-329):
  - The holder record needs that pid: `Stale()` probes it, and `releasableBy` uses it for the post-`/clear` self-release.
  - The code comment warns against recording a process pid that exits.
- What happens under each choice:
  - **Waiter pid copied:** the holder reads as stale as soon as `acquire --wait` exits successfully. The next mutation, such as another waiter's 15-second heartbeat, displaces it mid-merge. That allows two concurrent merges, the failure this lock exists to prevent.
  - **pid 0:** the holder reads as live forever. It is bounded only by the lease, and is wedged indefinitely when the lease is 0 (REQ-009).
- Codex independently reported the same defect (P1, 97%).
- Severity: critical. Class: blocking.
- Required fix:
  - Restore the owning-session pid and `pid_source` on the ticket (REQ-001, design D1).
  - State that promotion stamps that pid as the holder pid (REQ-007).
  - Add an AC: after B's waiter exits following promotion, B reads as live, a third party's mutation does not displace B, and B's own release succeeds.

N2. RESERVED-TICKET DROPPED — spec.md:L87-92 (REQ-004) against L169-178 (REQ-020); design.md:L48-49; acceptance.md:L113-114.
- REQ-004 drops "every ticket" whose waiter process is gone or whose heartbeat is older than 60 s. Reserved tickets are not exempt.
- A reserved ticket is created by the merge path (REQ-019), and that process exits.
- The owner then re-absorbs and re-measures with no waiter running (design D3 L48-49). That is required, since REQ-015 runs the re-measure before joining the queue.
- So the reserved ticket is dropped within 60 s. The 30-minute readiness bound and the front-once grant (Q10) can never take effect.
- AC-020 scenario 3 ("When 30m pass, Then dropped with reason readiness bound") contradicts REQ-004, which drops it first with reason "waiter gone".
- Codex reported the same (P1, 99%).
- Severity: critical. Class: blocking.
- Required fix:
  - Define reserved-ticket liveness separately: the owning session's liveness plus the readiness deadline, not waiter or heartbeat.
  - Exempt `reserved` from the waiter and heartbeat clauses of REQ-004.
  - State how a reserved owner reattaches a waiter once ready. Today REQ-003 resumes only `between-slices`, and REQ-012 forbids a second ticket.
  - Add an AC where another session mutates the queue during A's waiter-less re-measure and A's ticket survives.

N3. BETWEEN-SLICES CONTRADICTIONS — spec.md:L81-86 (REQ-003) against L87-92 (REQ-004) and L101-105 (REQ-007); design.md:L18-20; acceptance.md:L23-27.
- (a) Survival. A normal slice exit ends the waiter process, and REQ-004 drops a ticket whose waiter is gone. Even with that exempted, the heartbeat clause (60 s) fires before the 120-second re-entry grace (Q14). AC-003's "+119 s, same position" cannot pass under REQ-004 as written.
- (b) Promotion.
  - REQ-007's promotable set is `waiting` or ready `reserved`. `between-slices` is not in it.
  - Design D1 permits a holderless, non-empty queue in only two cases (hold; all reserved and not ready). A queue whose head is between slices is a third case the design does not admit.
  - It is also unstated whether a `--wait` newcomer, or a no-`--wait` caller such as the session-end automerge, may take a free window ahead of a between-slices ticket.
  - Plan §D makes the slice loop the standard lane mechanism, so this is the common path, not a corner.
- Codex reported (a) and (b) (P1, 99%).
- Severity: major. Class: blocking.
- Required fix:
  - Exempt `between-slices` from the waiter and heartbeat clauses until its re-entry deadline.
  - State whether a between-slices head is skipped (with REQ-012 adjusted) or blocks promotion until its deadline.
  - Add that case to design D1's holderless list.
  - Add an AC for a release while the head is between slices.

N4. REQ-020 REQUEUE COUNTER UNREACHABLE — spec.md:L169-178, L165-168 (REQ-019); acceptance.md:L114-120; research.md:L129-133.
- The trap:
  - Readiness requires "a record whose absorbed commit equals the current integration tip".
  - A requeue event is defined only on the in-window merge path (REQ-019).
  - A reserved ticket whose own re-measure is invalidated by a base move is simply not ready, so it is never promoted. It never reaches the merge path, and no requeue is ever counted.
- Consequences:
  - AC-020 scenario 4 ("When A's merge path runs, Then A's new ticket is waiting at the tail") is unreachable by any defined transition.
  - The "three consecutive requeues" bound (scenario 6) is reachable only through out-of-window tip moves.
  - In practice, a repeatedly invalidated reserved ticket exits only through the 30-minute readiness-bound drop.
  - research.md §R7 claims "a lane loses its place only after three consecutive requeues". That is false: the readiness-bound drop loses the place with zero counted requeues. Under sustained merges with a re-measure longer than the inter-merge interval, the lane is dropped every cycle. That is the starvation Q10 states it prevents.
- Codex reported the unreachability (P1, 98%).
- Severity: major. Class: blocking.
- Required fix, within Q4/Q10/Q15:
  - Define when an own-re-measure invalidation is observed and counted. One option: when the owner submits a record whose base is no longer the tip, outside the window.
  - Rewrite AC-020 scenario 4 to that transition.
  - Correct §R7 to name the readiness-bound drop as a second route to losing the place.
  - Add an AC with repeated invalidation and no promotion.

N5. DROPPED PUSH CLAUSE — iteration 1's REQ-MWQ-042, a fail-closed refusal when the CI state cannot be read ("CI client absent, unauthenticated, or the query fails … rather than treat the state as green").
- That clause appears in neither this SPEC nor SPEC-CANDIDATE-CI-001.
- `grep -n -i "cannot be read|unauthenticated|unreadable" <t1478>/spec.md` gives no output.
- research.md §R6 hands off (a) through (d), but not this clause. decision-index Q7 itself notes that the policy does not cover "측정 못 함" (could not be measured).
- This breaks "no clause silently dropped" across the move.
- Severity: major. Class: blocking.
- Required fix: add it to §R6 as hand-off item (e), "unreadable CI → refuse, never green". Record that the t1478 owner adopts or explicitly rejects it.

N6. CROSS-SPEC OBLIGATION NOT ACKNOWLEDGED.
- SPEC-CANDIDATE-CI-001 assigns to t1479 the rule "stale-candidate rebuild only when the card reaches the head of the merge-window queue, never eagerly" (t1478 spec.md:L144 D7, L163; plan.md:L14 B2).
- This SPEC's reserved flow already rebuilds only after a promotion. But no REQ or §R5 line states or accepts the "never eagerly" obligation.
- Severity: minor. Class: optional.
- Required fix: one sentence in §R5, or in REQ-015/019, acknowledging the obligation.

N7. FIXTURE README STALE AC REFERENCE — `.moai/reports/t1479/baseline-acquire-nowait/README.md` L5 and L37.
- The README says "AC-MWQ-007 compares against these files". The comparing AC is now AC-MWQ-011.
- Severity: minor. Class: optional.
- Required fix: update the reference. The committed baseline bytes (`3bc274dac`) are untouched, so this is not a re-capture.

N8. AC-014 PARTIAL COVERAGE — spec.md:L135-139 against acceptance.md:L75-79.
- REQ-014 forbids both clauses in both files. AC-014 checks `지명만이 근거` only in AGENTS.local.md, and `리더 공지가 여전히 첫 번째 층` only in gitflow-lane-protocol.md.
- Today the unchecked pairs are already 0 (see Evidence), so the gap is only against reintroduction.
- Severity: minor. Class: optional.
- Required fix: add the two cross-file `→ 0` greps.

## Regression Check (iteration 1 defects)

- **D1 BUDGET — RESOLVED.**
  - The counts are 25/25.
  - The lettered items are folded: 009a into REQ-009 L112-113; 021a/021b into REQ-015 and REQ-016.
  - The push verb moved to t1478 REQ-CCI-012/013 (t1478 spec.md:L102-104).
  - One clause did not survive the move (N5).
- **D2 MP-1 — RESOLVED.** Numbering is contiguous (see MP-1).
- **D3 HOLD — RESOLVED.**
  - REQ-008 (L106-109): no promotion of any kind while `hold`, and resume on `open`.
  - Design D1 (L18-20).
  - AC-008 (L46-49) tests release during hold.
- **D4 RESERVATION — PARTIALLY RESOLVED.**
  - Resolved: the reserved state is defined, non-blocking, with a 30-minute bound and front-once (REQ-020).
  - Not resolved: REQ-004 drops a reserved ticket before any of that applies (N2), and the starvation counter is unreachable (N4).
- **D5 ORPHAN WAITER — PARTIALLY RESOLVED.**
  - Resolved:
    - Liveness by waiter pid plus start time and heartbeat (REQ-001/004; AC-004 covers SIGKILL and pid reuse).
    - The timeout-versus-promotion race decided inside the mutation (REQ-006, AC-006).
    - Slices for the Bash call cap (REQ-003, plan §D).
  - Not resolved: the fix dropped the owner pid the holder needs (N1), and the between-slices state contradicts the drop rule (N3).
- **D6 HOLLOW FORM — RESOLVED** within Q12/Q16.
  - REQ-016 (L150-156) refuses a zero count, runner empty-sweep markers, and `go test` without `-json`.
  - AC-016 is a table (L92-99).
  - `true` stays valid as a decided boundary, with residual risk recorded in §D (L217-221).
- **D7 DIRTY TREE — RESOLVED.** REQ-017 (L157-159) requires a clean `git status --porcelain` (tracked and untracked) before and after, plus an unchanged HEAD. AC-017 (L100-103) has three scenarios.
- **D8 PUSH — MOVED.**
  - (a), (b) and (c) are handed off in §R6, and t1478 REQ-CCI-012 pins `<sha>:refs/heads/develop`.
  - REQ-CCI-013 adds the repair exception and the per-tip read.
  - The CI-unreadable clause was lost (N5).
- **D9 BASELINE — RESOLVED.**
  - The fixture is committed in `3bc274dac`, an ancestor of 6ab3dbcb2 (`git merge-base --is-ancestor` → ANCESTOR).
  - It contains `record.json`, `human.stderr`, `json.stderr`, and a README naming two normalizations.
  - AC-011 compares byte-for-byte (L59-63).
  - Optional nit: N7.
- **D10 AC-013 PATTERN — RESOLVED.** AC-014 gives literal grep patterns and counts (L75-79). Optional nit: N8.
- **Optional D11-D14 — addressed.**
  - D11: M0 is a manual fixture simulation (plan L51-58).
  - D12: `workflow.candidate_ci.enabled` is named and matches t1478 REQ-CCI-023 (t1478 spec.md:L130).
  - D13: the trust model is stated in §D (L214-216).
  - D14: positive greps added to AC-024.
- **Cross-SPEC consistency:** consistent.
  - REQ-CCI-011 is cited as the in-window landing check (REQ-018 L161-163), and matches t1478 L100.
  - REQ-CCI-009 is cited as the green verdict (design D4), and matches t1478 L96.

## Recommendation

FAIL with a STOP signal. Iteration 1 scored 0.75 and this iteration scores 0.69, a regression. The new liveness mechanism that closed D5 opened four contradictions in the same subsystem (N1-N4).

Per the LEAN clause, the orchestrator should not iterate again unconditionally. It should put three options to the leader/operator:
1. Scope-reduce.
2. Accept with debt.
3. Explicitly override and run a third iteration.

If a third iteration runs, its fix scope is narrow and has a single owner: the ticket-state model.
1. N1: restore the owner pid on the ticket and stamp it on promotion.
2. N2 and N3: give each ticket state its own liveness rule.
   - `waiting`: waiter and heartbeat.
   - `between-slices`: the re-entry deadline only.
   - `reserved`: the owner session plus the readiness deadline.
   - Also state promotion eligibility for `between-slices`, and update design D1's holderless cases.
3. N4: define where an own-re-measure invalidation is counted, fix AC-020 scenario 4, and correct §R7.
4. N5: add the CI-unreadable hand-off to §R6.

These fixes touch REQ-001/003/004/007/020, AC-003/004/020, design D1-D3 and research §R6/§R7. The counts stay at 25/25 if they are folded into the existing REQs.

### Blocking list
- N1 holder pid anchor (critical)
- N2 reserved ticket dropped by REQ-004 (critical)
- N3 between-slices survival and promotion contradictions (major)
- N4 REQ-020 requeue counter unreachable; §R7 claim false (major)
- N5 push clause old REQ-042 dropped across the move (major)

### fix_scope list
- spec.md REQ-MWQ-001, -003, -004, -007, -020; §F
- design.md D1 (fields and holderless cases), D2, D3
- acceptance.md AC-MWQ-003, -004, -007 (new owner-pid scenario), -020 (scenarios 3, 4 and 6 restated); optional AC-014 (N8)
- research.md §R6 (item e), §R7 (correct the guarantee); optional §R5 (N6)
- `.moai/reports/t1479/baseline-acquire-nowait/README.md` L5 and L37 (N7, optional; reference only, do not re-capture)

## Evidence

**Claim:** iteration 1's structural defects are closed. Five new or residual blocking defects exist in the queue-liveness model and the push hand-off.

**Evidence** (commands run in this session, verbatim results):
- `grep -c '^- \*\*REQ-MWQ-' spec.md` gave `25`, and `grep -c '^- \*\*AC-MWQ-' acceptance.md` gave `25`.
- The AC→REQ map was `001:001 … 025:025`.
- `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` gave rc=1. `grep -c syscall spec.md` gave `0`. `grep -in release-blocking acceptance.md` gave rc=1.
- The D7 loop gave six SPECs at `status: completed` and `SPEC-CANDIDATE-CI-001 NOT FOUND`.
- RED cells re-run on the audited tree:
  - E1: `grep -n '"wait"' internal/cli/integration.go` gave rc=1 (no flag).
  - E5: line `255: if !strings.Contains(strings.ToLower(string(raw)), full[:12]) {`.
  - E7: `grep -c "지명만이 근거" AGENTS.local.md` gave `1`.
  - E8: `kanban-dispatch-mechanics.md:120 … The announcement to the lead rides alongside it.` (template and local).
  - `grep -c "리더 공지가 여전히 첫 번째 층" gitflow-lane-protocol.md` gave `1`.
  - The new sentence was found in both files with count `0`.
- `git log --oneline -8 6ab3dbcb2` shows 6ab3dbcb2, 550e7c7fd, 3bc274dac, 1e1d0cc84, and so on. `git merge-base --is-ancestor 3bc274dac 6ab3dbcb2` gave ANCESTOR. The tree's HEAD ref file reads `6ab3dbcb2fc5042d5728af62379f8663ccfc5d1e`.
- `git show 1e1d0cc84:.moai/specs/SPEC-MERGE-WINDOW-QUEUE-001/spec.md`: old REQ-001 carried "its owning-session process id", and old REQ-042 carried the CI-unreadable refusal.
- internal/kanban/integration_lock.go was read at L172-180 (`Stale`), L245-252 (pid-recording warning) and L322-329 (`releasableBy`).
- t1478 draft greps:
  - L96 REQ-CCI-009, L100 REQ-CCI-011, L102-104 REQ-CCI-012/013, L130 REQ-CCI-023 (`workflow.candidate_ci.enabled`), L144 D7 (rebuild at queue head belongs to t1479).
  - The CI-unreadable patterns returned no output.
- `audit_multi` (project_root = the audited tree, target baseBranch) returned:
  - `overall_verdict: fail`, `participant_count 2`.
  - claude: fail (in-session anchor).
  - codex: fail, with four P1 findings matching N1-N4.
  - glm: inconclusive ("z.ai response carried no text content").
  - `disagreement_flag false`.
  - No `audit_receipt` was returned.

**Baseline-attribution:** the branch `WT-merge-window-queue` at `6ab3dbcb2`, read through the shared object database and the worktree's HEAD ref. `audit_multi` ran on the installed moai build `45600e4ee`, which it reports as an ancestor of 6ab3dbcb2 (binary lag; verification-claim-integrity §2.2). Its codex and claude findings concern SPEC text, not tool output, so the lag does not bear on them.

**Gaps:**
- The worktree guard refused `git -C <audited tree>` (rev-parse, status), so the audited tree's clean or dirty working state was not observed. HEAD was read from its ref file instead.
- The guard also refused the inline awk program for the AC-4/AC-5 traceability verb and the CN-4 verb. Both were applied by hand with plain greps; the traceability `COLLECTED` count (25) comes from grep, not from the verb.
- Piped RED cells E2, E4 and E9 were not re-run.
- Codex's reproductions are its own observations.
- One `cd` into the SPEC directory was used for `shasum`; it is read-only.

**Residual-risk:**
- GLM (advisory) did not participate.
- Tool outputs come from a lagging binary.
- No receipt was issued, so a PASS would not be corroborated. This verdict is FAIL.

## Operational Notes (unverified)
- `inferred` (rule: REQ-020 readiness = base equals the current tip): measure the drop rate under the readiness bound with a rule model, using N lanes, merge interval I and re-measure duration R with R > I. Expect each reserved lane to be dropped every cycle.
