auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: beaedc3f4b1f3a73c71ce48f82132147387bdf0e

# SPEC Review Report: SPEC-MERGE-WINDOW-QUEUE-001 (card t1479)
Iteration: 3/3 (Tier L ceiling, final round). Delta re-audit of iter2 N1-N8, CN-4 run in full, and a fresh deadlock-hole pass over the reduced design.
Verdict: FAIL. Escalation to the leader/operator is required (iteration cap reached).
Overall Score: 0.75 (Tier L threshold 0.85). Iteration history: 0.75 → 0.69 → 0.75, so iter3 recovered and no STOP signal fires.
Plan Artifact Hash (sha256, working tree, whose blobs equal HEAD beaedc3f4):
- acceptance.md     6cde637eebb8bb69815ad7ff854cd9befe9b209a0e8a8ae182850cae109b8aca
- decision-index.md c1c79e5c679200f554cff43806a93beedf557f595015631816e3f4aa650df061
- design.md         3897fb45f7a62204caeebb939603c8342c56e673431f5f1be270e9ef622a76c8
- plan.md           93fe3e5f2ae6102351d51e4cfc773656f405a6f336809aec3460fda360ccd879
- progress.md       b1a9276622e94bbfd86c59052b058b751edf9f84c3c4a4ae7639496888040c1e
- research.md       e990b01aa425587e5e3c7ba61d9fd0c95f71e35d0fd79dfade60c2642b549fec
- spec.md           6b1fe4d69495fb2a067a5d1e667eb1d80641c79047ae51ff2ce94a15eb5c3c74
Auditor Version: plan-auditor/v1 (iteration 3)

I ignored the reasoning context, per M1 Context Isolation. Operator decisions Q1 and the leader decisions Q2-Q18, Q17's scope reduction included, are taken as decided. Nothing below asks for fairness machinery back. Every finding concerns a seam the reduced design itself creates or leaves open.

## Must-Pass Results

- [PASS] MP-1 REQ numbering.
  - `grep -o '^- \*\*REQ-MWQ-[0-9]+' spec.md` gives 001 to 023, contiguous.
  - The AC map is `001:001 … 023:023`.
- [PASS] MP-2 GEARS, requirement layer only (spec.md §C, L71-194).
  - When: REQ-002 to 006, 014, 015, 016, 018, 019.
  - While: REQ-007 (with a When), REQ-017. REQ-012 is When + While.
  - Where: REQ-010.
  - Ubiquitous: REQ-001, 008, 009, 013, 021, 022.
  - Legacy "shall not": REQ-011, 020, 023. These fall inside the window, which runs to 2026-11-22.
- [PASS] MP-3 frontmatter, spec.md:L2-13.
  - All 12 canonical fields are present, and `version: "0.5.0"` is quoted.
  - `tags` is a string, and no rejected alias is used.
  - The extra keys `tier`, `card` and `related_specs` are additive.
- [PASS] MP-4. `go test -json` appears only as "in this repository" (REQ-015). REQ-022 requires the distributed text to stay neutral.
- [PASS] MP-5 D7.
  - Six referenced SPECs read `status: completed`.
  - SPEC-CANDIDATE-CI-001 is `NOT FOUND` in this tree. That is a SHOULD only: it is the sibling draft, and §R5 states it.
- [PASS] MP-6 D8. `grep -c syscall spec.md` gives `0`.
- [PASS] MP-7. `grep -c "NEEDS CLARIFICATION"` gives plan.md `0` and research.md `0`.
- [N/A] MP-8. `grep -c -i release-blocking acceptance.md` gives `0`, so no AC is release-blocking. As diligence I re-ran the RED cells (see Evidence).
- [PASS] MP-9. I applied CN-4 by hand.
  - COLLECTED: 9 milestones in plan order (M0..M8), 0 `Exit:` bindings.
  - Candidates: acceptance.md L18, 32, 47, 51, 59, 62, 64, 110, 133, 170.
  - Only L51 binds a milestone pair ("committed before the M1 code commit"), and plan.md places M0 (L52) before M1 (L61). No conflict.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | 0.50 | Several seams need interpretation, and different engineers would implement them differently. (B1) How factory complete relates to the new merge verb. (B2) What happens to the window and develop when an in-window check or the merge fails. (B3) Which integration target a promoted holder carries. |
| Completeness | 0.75 | 0.75 | All sections are present (§E Out of Scope H3s at spec.md:L222-245). One internal inconsistency: the REQ-013 doctrine sentence contradicts the self-dispatch clause it leaves standing in the same two files (B1c). |
| Testability | 0.75 | 0.75 | The ACs are binary, but none covers a failure path of the merge verb (B2) or a factory-mode lane landing through the verb (B1). |
| Traceability | 1.00 | 1.00 | 23 REQ definitions and 23 `maps` tags, one to one, with no orphans (grep). |

Aggregate: (0.50 + 0.75 + 0.75 + 1.00) / 4 = **0.75**, below 0.85.

## Defects Found (structured defect-list)

B1. COMPLETE ↔ MERGE-VERB CONTRACT. spec.md:L157-184 (REQ-017..021), design.md:L61-63 (D3), plan.md:L85-89 (M5); AGENTS.local.md:L219 and gitflow-lane-protocol.md:L99 in the audited tree.

- (a) The factory card never reaches merged-local. `moai integration merge --card` merges and releases, but no REQ says it moves the factory card to merged-local.
  - A factory lane that lands through the verb, as the new doctrine requires, leaves its card in `merge-ready`.
  - If it then runs `moai factory complete`, D3 says complete "performs the same in-window steps". Develop has already moved, by the lane's own merge, past `record.base`. So REQ-018 sends the lane to re-measure and re-acquire, a loop with no exit.
  - Codex reproduced the base mismatch on scratch git (P1).
- (b) Complete is not bound by the base check.
  - Absorption now happens outside the window, so a stale base is the common case, not a rare one. REQ-018's base check binds only `integration merge`.
  - Complete merges first (factory_card.go:L1409) and gates afterwards (REQ-021, L1422-1428). That path lands an unverified merge on develop, after which the gate refuses and the card is stuck in `merging`.
  - AC-019 tests only the "no record" case.
- (c) The doctrine contradicts itself.
  - REQ-013 writes "레인은 develop에 손으로 git merge 하지 않고 moai integration merge --card 로만 병합한다" ("lanes never `git merge` into develop by hand; they merge only through `moai integration merge --card`") into AGENTS.local.md and the protocol file.
  - Both files keep the self-dispatch clause "`moai factory complete`의 통합 절차로 … `develop`에 병합한다" ("merge into `develop` through the `moai factory complete` integration procedure"), at L219 and L99.
- Severity: critical. Class: blocking.
- Required fix:
  - State in the REQ text which of the two paths a factory lane uses. Either:
    - complete performs its in-window step through the REQ-017/018 step (base check before any card transition or merge, then release-and-re-acquire), or
    - the verb records the landing and complete adopts it without merging again.
  - Add ACs for "lane lands via verb then complete" and "complete with moved base: no merge, no card transition".
  - Reword the REQ-013 sentence, or the self-dispatch clause, so the two do not contradict each other, and extend AC-013 to check it.

B2. IN-WINDOW FAILURE OUTCOMES AND SHA PINNING. spec.md:L157-170 (REQ-017/018), design.md:L48-56.

- REQ-017 releases the window only on success, and REQ-018 only when the base moved.
- No outcome is defined for any of these:
  - branch tip tree ≠ record tree;
  - a landing-check refusal;
  - a failing `git merge` (conflict, failing hook, `MERGE_HEAD` left behind; codex reproduced this);
  - a post-merge tree mismatch, which happens after develop has already moved;
  - a caller that does not hold the window.
- These are wedge holes:
  - A failing holder that does not release blocks the queue until the 30-minute lease expires, or indefinitely when the lease is 0 and the owner is live (REQ-008).
  - A half-done merge left in the integration worktree fails every later holder.
- The merge is also not pinned to the verified commit. "Check the branch tip … perform the merge" reads the moving name twice, so a tip move in between lands an unverified tree. The post-merge check catches it only after develop has moved (codex P1, with a positive control using the pinned SHA).
- Severity: major. Class: blocking.
- Required fix:
  - Pin one resolved commit SHA across the check, the landing check and the merge.
  - For each pre-mutation refusal, state: refuse, release the window, exit non-zero naming the reason.
  - For a failure that leaves the integration worktree mid-merge, state the outcome: abort the merge or set policy `hold` with a named recovery, never silent promotion onto a dirty tree.
  - Make a non-holder invocation refuse.
  - Add the failure scenarios to AC-017.

B3. PROMOTION DROPS THE INTEGRATION TARGET. spec.md:L71-76 (REQ-001), L96-102 (REQ-006); design.md:L12, L15-18 (D1).

- D1's promotion copy set is `session_id, session_name, card, pid, pid_source`. The ticket carries no `branch`, `branch_source` or `worktree`.
- Today's acquire records them, and readers rely on them:
  - factory complete treats a window as its own only when `lock.Branch != ""` (factory_card.go:L1335);
  - its refusal (1) reads `BranchSource` (L1351);
  - the merge verb needs the integration branch.
- A holder created by promotion inside another process's mutation therefore has no target, or inherits a guess.
- Codex reported the same issue (P2). It is blocking here because D1 enumerates the copy set explicitly.
- Severity: major. Class: blocking.
- Required fix:
  - Resolve the target at enqueue and store it on the ticket.
  - Copy it to the holder on promotion.
  - Extend AC-001 and AC-006 accordingly.

O1. `status` as a mutation. acceptance.md:L41-42 (AC-006 scenario 3) needs `status` to promote, but REQ-009 says only that status "shows". Name which verbs are mutations. Minor, optional.

O2. A live waiter whose ticket was dropped, for example because its heartbeat stalled under load 120-255, has no defined behavior: it could exit, re-enqueue at the tail, or poll forever. Minor, optional.

O3. Double ownership. t1478 REQ-CCI-024 also lands the "manual lane merge forbidden" clause in the same two files. AC-013 expects a count of exactly `1`, so the second card to land must reconcile the wording. Minor, optional.

O4. Branch resolution is by `.claude|.moai/worktrees/<card>` path only. That follows REQ-CCI-004, but the verb has no `--branch` fallback (t1478 has one) and does not use the factory record's `WorktreePath`. An agent-named card tree, like this card's own `agent-a1d44…`, does not resolve. Minor, optional.

## Regression Check (iteration 2 defects)

- N1 HOLDER-PID: RESOLVED.
  - REQ-001 (L73-75) puts the owner pid and the `session-owner` source on the ticket.
  - REQ-006 (L97-101) stamps both onto the holder on promotion.
  - D1 (L15-18) says the same.
  - AC-006 scenario 2 (L37-40) covers a promoted holder whose waiter has exited: it reads live, is not displaced by a third party, and can release itself.
  - This matches `Stale()` (integration_lock.go:L172-180) and `releasableBy` (L322-329).
- N2, N3, N4: DISSOLVED by Q17's removal.
  - A residual grep over spec, acceptance, plan, design, research and progress for slice, reserved, front-once, requeue, between-slices, readiness and re-entry grace finds hits only in the spec HISTORY rows (L26-29), §E's Out of Scope bullet (L230), research §R7 (history and residual risk, L139-147), and progress history (L18-25).
  - decision-index marks Q4, Q10 and Q15 superseded and Q11 and Q14 partly superseded (L55, 117, 129, 161, 173).
  - No live REQ, AC or milestone references the removed machinery.
- N5: RESOLVED. research §R6 (e), L124-132. The t1478 draft v0.5.0 adopted it in REQ-CCI-014 ("When the tip's CI state cannot be read … shall refuse … never be treated as green").
- N6: RESOLVED. REQ-018 (L169-170) and research §R5 (L95-98).
- N7: RESOLVED. The fixture README (L4-6) now says AC-MWQ-010. Only README.md differs between 3bc274dac and beaedc3f4 (`git diff --stat` shows 1 file); the stderr and record blobs are unchanged.
- N8: RESOLVED. AC-013 (L74-79) checks all four clauses in both files.

## Specific verifications requested

| Item | Result |
|---|---|
| Session-owner pid stamped on promotion; `Stale` and `releasableBy` unchanged | PASS (N1 above) |
| Queue liveness: owner gone, waiter id + start time, heartbeat > 60 s | PASS. REQ-003 L83-87; AC-003 covers ±1 s and pid reuse. |
| Timeout versus promotion under the lock mutation | PASS. REQ-005 L91-95, D2 L38-39, AC-005 with both orders. The first write wins; a late-observed promotion releases. |
| Hold means no promotion | PASS. REQ-007 L103-106, D1 L23-24, AC-007. |
| `integration merge --card`: branch resolution, landing check, no-op when off | Partially. The contract is consistent with t1478 REQ-CCI-004 (path resolution) and REQ-CCI-011 (call site named at t1478 spec L101). AC-017 asserts zero calls when off and one call when on. Failure outcomes and pinning are missing (B2); the factory-state seam is missing (B1). |
| Develop moved: release and re-acquire | PASS for the verb (REQ-018, AC-018). Not bound on the complete path (B1b). |
| Substantive completion gate | PASS on its own terms (REQ-019..021; AC-020/021; the E5 RED reproduces at L255). The B1 seams sit on top of it. |
| Clean tree before and after | PASS. REQ-016, AC-016 (untracked, tracked edit, commit). |
| Committed baseline fixture | PASS. `3bc274dac` is an ancestor of beaedc3f4. The working blobs equal HEAD (hash-object equals ls-tree for all 10 files). |
| AC-013 four clauses in both files | PASS as written, with the RED reproduced (see Evidence). The contradiction with the self-dispatch clause is B1c. |
| New deadlock holes | Yes: B2 (a failing holder never releases, and a half-done merge poisons later holders) and B1a (a verb-then-complete re-measure loop). Not holes: a promoted waiter dying, which is bounded by owner liveness and the lease; a leader that sets `hold` and departs, where waiters exit at their bound and any non-lane session can run `policy open`. |

## Recommendation

FAIL at the iteration cap. Under the Retry Loop Contract, the orchestrator escalates to the user with three options:
1. PASS-with-debt, carrying B1-B3 as run-phase obligations.
2. A scope fix and re-entry.
3. An explicit iter4 override.

A note for that decision: the fix scope is narrow and sits in a single place, the in-window step. It needs no new REQ, and the counts stay at 23/23 if the fixes fold into REQ-001/006/013/017/018/019.

### Blocking list
- B1 complete ↔ merge-verb contract: factory card never reaches merged-local, re-measure loop, unbounded stale-base merge on complete, doctrine self-contradiction (critical)
- B2 in-window failure outcomes undefined, which wedges the window and poisons later merges; merge not pinned to the verified SHA (major)
- B3 promotion drops the integration target (branch, branch_source, worktree) (major)

### fix_scope list
- spec.md REQ-MWQ-001, -006 (target fields); -017, -018 (pinned SHA, failure outcomes, non-holder refusal); -019 (complete bound to the in-window step or adopting the verb's landing); -013 (sentence reconciled with the self-dispatch clause)
- design.md D1 (copy set), D3 (complete path, failure branches)
- acceptance.md AC-001, -006 (target fields); AC-017 (failure scenarios and branch move); AC-018 or AC-019 (complete with moved base; verb-then-complete); AC-013 (self-dispatch clause check)
- Optional: REQ-009 or AC-006 sc3 (O1); REQ-003 or REQ-004 (O2); research §R5 (O3, O4)

## Evidence

**Claim:** iter2's N1-N8 are resolved or dissolved, and three blocking seams remain in the in-window step of the reduced design.

**Evidence** (commands run in this session, with the results observed):
- HEAD: `.git/worktrees/agent-a1d44be9332845923/HEAD` reads `ref: refs/heads/WT-merge-window-queue`.
- `git rev-parse beaedc3f4` gives `beaedc3f4b1f3a73c71ce48f82132147387bdf0e`. `git log --oneline -6 beaedc3f4` shows beaedc3f4, 6ab3dbcb2, 550e7c7fd, 3bc274dac, 1e1d0cc84, c8a2d8b5e.
- `git merge-base --is-ancestor 3bc274dac beaedc3f4` gives `ANCESTOR`.
- `git diff 3bc274dac beaedc3f4 --stat -- .moai/reports/t1479/baseline-acquire-nowait/` shows only `README.md | 7`.
- Working-tree cleanliness of the audited files: `git ls-tree beaedc3f4` blob ids equal `git hash-object --no-filters` for all 7 SPEC files and the fixture's README, human.stderr and record.json.
- REQ and AC collection: REQ 001..023; AC map `001:001 … 023:023`.
- Must-pass greps: `release-blocking` gives `0`; `NEEDS CLARIFICATION` gives `0` and `0`; `syscall` gives `0`.
- D7 loop: six `status: completed`, and `SPEC-CANDIDATE-CI-001 NOT FOUND`.
- RED cells re-run on the audited tree:
  - AGENTS.local.md:221 contains "리더의 창 지명만이 근거다." ("only the leader's window nomination counts").
  - gitflow-lane-protocol.md:56 contains "**리더 공지가 여전히 첫 번째 층**" ("the leader's announcement is still the first layer").
  - The two new sentences are absent.
  - kanban-dispatch-mechanics.md:120, in both the template and local copies, contains "The announcement to the lead rides alongside it."
  - card_evidence_readers.go:255 contains `if !strings.Contains(strings.ToLower(string(raw)), full[:12]) {`.
- Self-dispatch clauses: AGENTS.local.md:219 and gitflow-lane-protocol.md:99 contain "`moai factory complete`의 통합 절차로 … 병합한다" ("merge through the `moai factory complete` integration procedure").
- Code read:
  - factory_card.go:L1331-1432: `heldByUs` requires `lock.Branch != ""`; the merge happens at L1409 before the gate at L1422.
  - integration_lock.go:L172-180, L253-299, L322-368. Release deletes the whole record file, so the run must change it to keep the queue; D6 already says so.
  - session_pid.go:L92 `ResolveOwnerPID`.
- t1478 draft (agent-a18f82893f35ed6c4), v0.5.0:
  - REQ-CCI-004 at L83, REQ-CCI-011 at L101 (names this SPEC's verb as a call site), REQ-CCI-014 at L107 (unreadable CI refuses).
  - REQ-CCI-024 at L133 (manual-merge clause lands only after this verb).
  - D7 at L145, D9 at L147.
- `audit_multi` (project_root = the audited tree, target baseBranch) returned:
  - `overall_verdict: fail`, `participant_count 2`, `disagreement_flag false`.
  - claude (in-session anchor): fail (B1).
  - codex: fail, with P1 for pinning (B2), P1 for verb-then-complete (B1a), P2 for target fields (B3), and P2 for failure handling (B2). Each came with scratch-git reproductions.
  - glm: inconclusive ("z.ai response carried no text content").
  - No `audit_receipt` was returned.

**Baseline-attribution:** branch `WT-merge-window-queue` at `beaedc3f4`, read through the shared object store (plain `git` from the auditor's own worktree) and absolute-path file reads. `audit_multi` ran on the installed moai build `45600e4ee`, which it reports as an ancestor of beaedc3f4 (binary lag, verification-claim-integrity §2.2). Its findings concern SPEC text, not tool output.

**Gaps:**
- The worktree guard refused `git -C <audited tree>` and any composite command containing the substring "git" (including the `gitflow-…` path). Working-tree cleanliness was therefore established by blob-hash equality on the 10 audited files, not by `git status` over the whole tree.
- The Grep tool is unavailable in this session; plain `grep` was used.
- The AC-4/AC-5 and CN-4 awk verbs were applied by hand with grep, not run as scripts.
- Piped RED cells E2, E4 and E9 were not re-run.
- Codex's scratch-git reproductions are its own observations, not mine.

**Residual-risk:**
- GLM (advisory) did not participate.
- No receipt was issued.
- B1-B3 were judged from SPEC text plus a read of the current code; the run phase may find more seams in the same in-window step.

## Iteration history (escalation record)
- iter1 (1e1d0cc84): FAIL 0.75. D1-D10 blocking: budget, numbering, hold, reservation, orphan waiter, hollow form, dirty tree, push, baseline, AC pattern.
- iter2 (6ab3dbcb2): FAIL 0.69, STOP. N1-N5 blocking: holder pid, reserved drop, between-slices, requeue counter, push clause.
- iter3 (beaedc3f4): FAIL 0.75. N1-N8 resolved or dissolved by Q17. New blocking B1-B3, all in the in-window step and the complete seam. No defect persisted across all three iterations, so there is no stagnation flag.

## Operational Notes (unverified)
- `assumption`: B1 closes most cheaply if `factory complete` delegates its whole in-window step to the merge verb. To check that before choosing, measure it: `grep -n "factoryMergeNoFF" internal/cli/*.go`, which counts the merge call sites that would need to converge.
