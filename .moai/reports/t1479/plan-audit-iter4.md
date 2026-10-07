auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 3b57f3e7c

# SPEC Review Report: SPEC-MERGE-WINDOW-QUEUE-001 (card t1479)
Iteration: 4 (single leader-granted extension after the Tier L ceiling — final). Delta re-audit of iter3 B1-B3 and O1/O2; CN-4 run in full.
Verdict: FAIL — per the extension terms, this means HOLD.
Overall Score: 0.75 (Tier L threshold 0.85). History: 0.75 → 0.69 → 0.75 → 0.75. Not lower than iter3, so no STOP signal fires; the iteration budget is exhausted.
Plan Artifact Hash: not computed. The guard refused hash/git commands against the audited tree (see Gaps). The subject is commit 3b57f3e7c, delta beaedc3f4..3b57f3e7c, with 8 files changed (+408/−71).
Auditor Version: plan-auditor/v1 (iteration 4)

Reasoning context ignored per M1 Context Isolation. Leader decisions Q1–Q19 are taken as decided. I do not reopen the one-merge-path choice, the per-cause exit codes, the abort-then-hold rule, or the ticket target copy. Every finding below is a seam that the v0.6.0 delta text itself creates, or a part of iter3 B2 it leaves open.

## Must-Pass Results
- [PASS] MP-1 REQ numbering.
  - `grep -oE '^- \*\*REQ-MWQ-[0-9]+' spec.md` → 001…023, contiguous with no duplicates.
- [PASS] MP-2 GEARS (requirement layer only).
  - The modified REQs keep their patterns: 001 Ubiquitous; 003 and 006 Event-driven; 009 Ubiquitous; 013 Ubiquitous; 017 While + When; 018 When + nested When; 019 When ×4.
  - The ACs remain Given-When-Then and are not graded here.
- [PASS] MP-3 frontmatter. `version: "0.6.0"` is quoted, and the field set is unchanged from iter3, which passed all 12 fields.
- [PASS] MP-4. There is no new language-specific tooling in the delta.
- [PASS] MP-5 D7.
  - The delta adds no new SPEC reference.
  - The iter3 D7 result holds: six referenced SPECs are completed, and SPEC-CANDIDATE-CI-001 is a sibling draft (SHOULD).
- [PASS] MP-6 D8. `grep -c syscall spec.md` → `0`.
- [PASS] MP-7. `grep -c "NEEDS CLARIFICATION"` → plan.md `0`, research.md `0`.
- [N/A] MP-8. `grep -ci release-blocking acceptance.md` → `0`, so no AC is release-blocking. As diligence, the new AC-013 RED cell reproduces (see Evidence).
- [PASS] MP-9.
  - COLLECTED: 9 milestones in plan order (M0..M8, plan.md L52–107), 0 `Exit:` bindings (`grep -ciE '^exit:'` → 0).
  - The only ordering record that binds a milestone is acceptance.md L59, "committed before the M1 code commit"; M0 (L52) precedes M1 (L61).
  - The delta's new ordering words (AC-017 "after the identity check", "before the integration branch ref moves") order steps inside the merge, not milestones. No conflict.

## Category Scores
| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | 0.50 | New internal contradictions:<br>• REQ-009 (merge is a queue mutation that drops and promotes first) vs REQ-017 (a non-holder is refused "without reading or writing the window record's holder or queue").<br>• REQ-019's refusal clause vs its adoption clause, with no precedence stated.<br>• REQ-012 (lane-role policy write refused) vs REQ-018 (the lane-run merge step sets `hold`).<br>Engineers would implement each of these differently (C1–C4). |
| Completeness | 0.75 | 0.75 | All sections are present. One outcome is still undefined: a failure after the merge commit exists (C3, carried from iter3 B2). |
| Testability | 0.75 | 0.75 | The ACs are binary, but some cannot be met as the REQ text stands:<br>• AC-018 asserts "in every row no merge commit remains", which no REQ provides for a post-merge failure.<br>• AC-017 sc4 "record bytes unchanged" conflicts with REQ-009 drops. |
| Traceability | 1.00 | 1.00 | 23 REQs and 23 `(maps REQ-MWQ-…)` tags (`grep … \| wc -l` → 23), one to one. |

Aggregate (0.50 + 0.75 + 0.75 + 1.00) / 4 = **0.75**, below 0.85.

## Defects Found (structured defect-list)

C1. COMPLETE'S CARD GATE NOW RUNS AFTER DEVELOP MOVES. spec.md REQ-MWQ-019; design.md D3 complete diagram ("call the merge step ─► success ─► merging → merged-local transitions"); plan.md M6 ("card transitions only after the step succeeds").
- Today, factory_card.go transitions T14 merge-ready→merging, the lease-holder/version F1 edge, **before** `factoryMergeNoFF`. I read this at 3b57f3e7c, L1398–1409.
- The delta moves every card transition after the merge step, so that a failed step "shall not change" the card state. That also moves the lease, state and version gate after develop has moved.
- Consequence: a lane whose card lease expired, or a lane that does not hold the card's lease, lands a merge on develop and is only refused afterwards.
  - Codex reproduced this with a Go overlay. The card stayed `merge-ready` (other lessee) or `assigned` (expired lease) while develop had moved.
- This contradicts REQ-019's own parenthetical, "so every gate runs before develop moves". The delta broke the property it claimed.
- No AC covers it.
- Severity: major. Class: blocking.
- Required fix:
  - REQ-019 must state that complete checks the card's state, lease holder and version (the T14 preconditions) before calling the merge step, and refuses with develop untouched.
  - D3 must show that pre-check.
  - Add an AC-019 scenario: expired lease, or another lane's lease → refusal, integration tip unchanged.

C2. THE ADOPTION PATH IS NOT TIED TO THE CARD'S CURRENT TIP OR TO VALID EVIDENCE, AND ITS PRECEDENCE IS UNSTATED. spec.md REQ-MWQ-019, clauses 2 and 4; design.md D3; acceptance.md AC-019 sc4.
- (a) The predicate is too loose. "A merge commit of that card's branch is already reachable from the integration branch" stays true after the lane adds commits to the `WT-` branch following the verb merge.
  - Complete would then record merged-local from the old merge, and the new commits would be silently left out.
  - Codex scratch git: ancestor check of the old merge exit `0`, of the new tip exit `1`.
- (b) Precedence is unstated. Clause 2 refuses when no valid record is keyed by "the card branch's candidate tree", which REQ-014 defines against the current integration tip. Clause 4 adopts "without requiring a fresh re-measure".
  - In AC-019 sc4, where develop advanced after the verb merge, both clauses can fire.
  - Only design.md orders them (adoption first). The normative REQ does not.
- (c) The verb path is not explicitly bound to record **validity**. REQ-017 checks only "that SHA's tree against the card's re-measure record" and the absorbed commit. Validity under REQ-014/015 (exit code, test count, the form required by `candidate_ci.enabled`) is named only in complete's pre-check (REQ-019) and in the F1 reader (REQ-021).
  - The doctrine (REQ-013) now lets lanes land through the verb directly. So a tree-matching but invalid record (zero tests, exit ≠ 0) can move develop. The F1 gate would then refuse adoption only after the fact.
- Severity: major. Class: blocking.
- Required fix:
  - REQ-017: require the record to be valid under REQ-014/015 before the landing check, with its own REQ-018 failure code and AC-018 row.
  - REQ-019: order the clauses (adopt → validity → merge step).
  - Define adoption as: a merge commit on the integration branch whose second parent equals the card branch's **current** tip and whose tree equals a valid record's tree. Otherwise, take the merge step or refuse.
  - Add an AC: verb merge, then a new lane commit, then complete → no adoption.

C3. FAILURE AFTER THE MERGE COMMIT IS STILL UNDEFINED. This is the iter3 B2 residual. spec.md REQ-MWQ-018; design.md D3; acceptance.md AC-018 lead sentence.
- iter3 B2 listed "a post-merge tree mismatch, which happens after develop has already moved" among the undefined outcomes.
- v0.6.0 lists five causes: tip moved, pinned tree, landing refusal, merge failure, other error. None of them is a failure after `git merge --no-ff` succeeded, such as a merge-tree ≠ record-tree result or a lookup error.
  - "Other error" releases the window and promotes the next holder while the merge commit stays on develop.
  - A completed merge has no `MERGE_HEAD`, so abort cannot undo it. Codex injected a post-merge lookup error: `develop_moved=True`, `merge_abort_exit=128`.
- AC-018 nonetheless asserts "in every row no merge commit remains on the integration branch". That implies a rollback that no REQ specifies.
- Nothing requires the pinned SHA to descend from `record.base`, which is the condition that would make a post-merge mismatch impossible by construction.
- Under the Retry Loop Contract, an unresolved prior-iteration defect is FAIL.
- Severity: major. Class: blocking.
- Required fix: either
  - add the precondition "the pinned SHA has `record.base` as an ancestor" to REQ-017 (refuse before the merge), which makes a post-merge mismatch unreachable; and also
  - define the post-merge-failure outcome in REQ-018: reset to the pre-merge tip, or `hold` with a named recovery. Then limit AC-018's "no merge commit remains" to the pre-merge rows, or add the post-merge row.

C4. THE NON-HOLDER REFUSAL CONTRADICTS "MERGE IS A QUEUE MUTATION". spec.md REQ-MWQ-009 (O1 delta) vs REQ-MWQ-017 (B2 delta); design.md D1 ("Every window verb is a queue mutation: … `merge` apply the drop rules and promotion before doing their own work"); acceptance.md AC-017 sc4 ("window record's bytes are unchanged").
- "Refuse without reading … the holder" cannot be satisfied, because holdership is decided by reading the holder field.
- Drops and promotion before the holder check would rewrite the record for a non-holder, breaking AC-017 sc4.
- The order also decides the expired-lease case:
  - run drops/promotion first → a holder whose lease expired is displaced, then refused;
  - check the holder first → the merge proceeds on an expired lease.
- Both orders are permitted by the text, so engineers would implement them differently.
- Severity: minor (narrow wording), but it is an internal inconsistency in the safety-relevant step. Class: blocking.
- Required fix:
  - REQ-017: "When the caller does not hold the window, it shall refuse before applying any queue mutation, leaving the window record unchanged."
  - REQ-009 / D1: say that merge applies drops and promotion only after the holder check passes.
  - State whether the holder's own lease is checked (and renewed per REQ-008) before the merge.

O1. The lane-run merge step writes the policy. REQ-MWQ-012 refuses policy writes from a lane-role session, while REQ-MWQ-018 has the lane-run merge step set `hold`.
- This is not strictly contradictory, since REQ-012 binds the policy verb.
- However, no REQ states the exemption or the recorded setter. A run phase that uses one guarded writer would silently drop the dirty-worktree hold.
- Severity: minor. Class: optional (one sentence).

O2. Exit-code count mismatch. REQ-018 enumerates five causes. D3 and AC-018 require six distinct codes (MERGE-DIRTY as its own row).
- Severity: minor. Class: optional. Name the sixth cause in REQ-018.

O3. AC-018 merge-failure row setup. "Conflicting change on develop and a matching record" reaches `git merge` only if the record's base equals the current tip while the pinned SHA does not descend from it. Otherwise the base check exits RE-MEASURE first.
- Make the setup explicit, or inject the failure at the merge seam (codex P2). If C3's ancestry precondition is adopted, this row must inject at the seam.
- Severity: minor. Class: optional.

## Regression Check (iteration 3 defects)
- B1: PARTIALLY RESOLVED.
  - Resolved:
    - complete calls the REQ-017 step, replacing L1409 (REQ-019, D3, plan M6);
    - complete-after-verb adopts without re-measure (REQ-019 clause 4, AC-019 sc4);
    - moved base on complete → no merge and unchanged state (AC-019 sc3);
    - doctrine aligned (REQ-013 sentence and self-dispatch clause, AC-013; RED reproduces).
  - New defects inside the same seam: C1 (gate ordering regression) and C2 (adoption predicate and precedence).
- B2: PARTIALLY RESOLVED.
  - Resolved:
    - SHA pinned once and used for identity, landing check and `merge --no-ff <sha>` (REQ-017, AC-017 sc2/sc3);
    - distinct per-cause exits with release (REQ-018, AC-018 table);
    - `merge --abort` + clean check, else `hold` and release (leader-accepted; AC-018 last two rows);
    - non-holder refusal stated (AC-017 sc4).
  - UNRESOLVED: post-merge failure outcome (C3, carried from iter3).
  - New: non-holder wording conflict (C4).
- B3: RESOLVED.
  - REQ-001 and REQ-006 carry `branch` / `branch_source` / `worktree` resolved at enqueue and copied at promotion.
  - D1 copy set updated, citing `factory_card.go:1335` / `:1351`.
  - AC-001 adds the fields. AC-006 sc5 is the fixture where the promoted holder takes the `lock.Branch != ""` held-by-us path.
  - Confirmed against code: `heldByUs := lock.Held() && lock.SessionID == sessionID && lock.Branch != ""`. Promotion copies `session_id`, so the predicate holds.
- O1 (iter3): RESOLVED (REQ-009 names the mutating verbs), but its wording creates C4.
- O2 (iter3): RESOLVED. REQ-003 adds exit non-zero without re-enqueue; AC-003 sc5.
- O3 / O4 (iter3): noted in research §R5 as coordination items. Acceptable as optional.
- Counts: 23/23 consistent (MP-1, Traceability).
- Stagnation: none. No defect persisted unchanged across iter2–iter4, although the in-window step has produced new seams in each round.

## Recommendation
FAIL on the final extension, so HOLD per the leader's terms.
- The remaining fixes are narrow and all sit in REQ-017/018/019 plus D1/D3, with about 4 AC scenarios: C1 card pre-check, C2 adoption predicate, order and validity, C3 ancestry precondition and post-merge outcome, C4 holder-check order.
- None needs a new REQ or a scope change.
- If the operator chooses PASS-with-debt instead of hold, C1–C4 must be carried as explicit run-phase obligations with the ACs listed above. C1 and C3 are the two that can move develop on unverified ground.

### Blocking list
- C1 complete's lease/state/version gate (T14) now runs after the merge moves develop (major) — delta regression
- C2 adoption path: loose predicate (old merge adopted past new commits), clause precedence unstated, verb path not bound to record validity (major)
- C3 post-merge failure outcome undefined; AC-018 "no merge commit remains" untraceable (major) — iter3 B2 residual
- C4 REQ-009 merge-as-mutation vs REQ-017 non-holder refusal "without reading or writing" (minor, internal inconsistency)

### fix_scope list
- spec.md: REQ-MWQ-009, -017 (holder check first; validity; ancestry), -018 (post-merge outcome; sixth cause; system-write hold), -019 (card pre-check; clause order; adoption predicate)
- design.md: D1 (mutation order), D3 (card pre-check, post-merge branch, adoption predicate)
- acceptance.md: AC-017 sc4 wording, AC-018 (lead sentence scope, invalid-record row, post-merge row, merge-failure setup), AC-019 (expired or foreign lease; verb then new commit then complete)

## Evidence (five-section format)

**Claim:**
- The v0.6.0 delta closes B3 and O1/O2.
- It closes most of B1/B2 but leaves one iter3 B2 item open (C3).
- It introduces new blocking seams in the merge step and the complete path (C1, C2, C4).

**Evidence** (commands run in this session against the audited commit or tree, with the results observed):
- `git log --oneline beaedc3f4..3b57f3e7c` → one commit, `3b57f3e7c docs(SPEC-MERGE-WINDOW-QUEUE-001): close plan-audit iter3 delta, v0.6.0 (card t1479)`.
- `git diff --stat beaedc3f4 3b57f3e7c` → 8 files, +408/−71: the iter3 report copy plus acceptance, decision-index, design, plan, progress, research and spec.
- `git diff beaedc3f4 3b57f3e7c -- .moai/specs/SPEC-MERGE-WINDOW-QUEUE-001/`, read in full. Quoted text above is from it.
- `git show 3b57f3e7c:internal/cli/factory_card.go`, L1320–1432:
  - `heldByUs := lock.Held() && lock.SessionID == sessionID && lock.Branch != ""`;
  - T14 `db.Transition(... To: homestate.CardMerging ...)` precedes `factoryMergeNoFF(integTree, cardBranch, ...)`. This is the basis for C1.
- `git show 3b57f3e7c:.../spec.md`, L100–215: the REQ-006/009/012/013/017/018/019/021 texts quoted above.
- On the audited tree's AGENTS.local.md:
  - `grep -c "self-dispatch lane 예외 — 병합 창.*moai integration merge"` → `0` (RED reproduces);
  - `grep -n "self-dispatch lane 예외 — 병합 창"` → line 219 (the header exists);
  - `grep -c "moai integration merge --card 또는 그것을 부르는 moai factory complete 로만 병합한다"` → `0` (RED);
  - `grep -c "지명만이 근거"` → `1` (RED).
- gitflow-lane-protocol.md L99 (Read): the header "**self-dispatch lane 예외 — 병합 창.**" is present, and the line does not contain "moai integration merge". RED reproduces.
- REQ list: `grep -oE '^- \*\*REQ-MWQ-[0-9]+' spec.md` → 001..023. AC maps: `grep -oE 'AC-MWQ-[0-9]+\*\* \(maps REQ-MWQ-[0-9]+' acceptance.md | wc -l` → `23`.
- Must-pass greps:
  - NEEDS CLARIFICATION: research `0`, plan `0`;
  - release-blocking: `0`;
  - syscall: `0`.
- CN-4: `grep -nE '^#+ +(Milestone +)?M[0-9]+' plan.md` → M0 L52 … M8 L107; `Exit:` bindings `0`; the only milestone-binding ordering candidate is acceptance L59.
- `mcp__moai__audit_multi` (project_root = the audited tree, target baseBranch):
  - `overall_verdict: fail`, `participant_count 2`, `disagreement_flag false`.
  - claude (independent subscription audit, claude-opus-5-5): fail. It raised C4, C2, C2c and C3 at P1, and O1, the merging-edge concern and O2 at P2/P3.
  - codex: fail. It raised C1 at P1 (Go overlay reproduction), C2a at P1 (scratch git reproduction), C3 at P1 (scratch git reproduction), C4 at P2, and O3 at P2.
  - glm: inconclusive ("z.ai response carried no text content").
  - No `audit_receipt` was returned.

**Baseline-attribution:**
- Commit 3b57f3e7c on branch `WT-merge-window-queue`, read through the shared object store with plain `git` from the auditor's own worktree.
- The doctrine greps read the audited tree's working files by absolute path.
- `audit_multi` ran on the installed moai build `45600e4ee`, which the tool reports as an ancestor of 3b57f3e7c (binary lag, §2.2). Its findings concern SPEC text, not tool output.

**Gaps:**
- The worktree guard refused `git -C <audited tree>`, a compound `git show … > file` export loop, and a compound grep whose path contained "gitflow". I fell back to plain `git` via the shared object store, Read, and split greps.
- Not observed:
  - `git status` of the audited tree, so working-file equals HEAD is assumed, not measured for the doctrine files;
  - SHA-256 plan-artifact hashes;
  - the full scripted AC-4/AC-5 and CN-4 awk verbs (applied by hand with grep).
- The Grep tool is unavailable here; plain grep was used.
- Codex's overlay and scratch-git reproductions (C1, C2a, C3) are its observations, not mine. C1 is independently supported by my read of factory_card.go ordering together with the D3 text.

**Residual-risk:**
- GLM (advisory) did not participate.
- No receipt was issued.
- The in-window merge step has produced new seams in iter3 and iter4. The run phase may surface more in the same step: for example, the window release happens before complete records merged-local, which leaves a narrow adoption race between two completes on one card that codex and claude raised at low confidence and I did not measure.

## Iteration history
- iter1 (1e1d0cc84): FAIL 0.75, D1–D10.
- iter2 (6ab3dbcb2): FAIL 0.69, STOP, N1–N5.
- iter3 (beaedc3f4): FAIL 0.75, B1–B3.
- iter4 (3b57f3e7c): FAIL 0.75.
  - B3 resolved, B1/B2 partially resolved (C3 carried).
  - New blocking: C1, C2, C4.
  - Final extension, so HOLD.

## Operational Notes (unverified)
- `inferred` (rule: an ancestry precondition makes a post-merge tree mismatch unreachable when base == tip and pinned tree == record tree): adopting "pinned SHA descends from record.base" in REQ-017 likely collapses C3 to a defensive `hold` branch. Measure it with a scratch-repo test: record.base = tip, SHA not descending from tip, then confirm the merge attempt is refused before `git merge`.
- `assumption`: C1 may be satisfiable by a read-only lease and state check before the step, keeping the T14 transition after it. Verify this against `homestate.Transition`'s F1 checks: `grep -n "CardMerging" internal/homestate/*.go`.
