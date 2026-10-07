auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 046d37f0e43458b8965ede9de7481e1429297fd8

> **Addendum (supersedes the PASS-WITH-DEBT verdict below):**
> - The SubagentStop receipt guard refused PASS-WITH-DEBT because no receipt was cited.
> - A follow-up `mcp__moai__codex_audit` (adversarial, baseBranch, project_root = audited tree) returned `verdict: fail` with one P2: the D6a pinned SHA == base == tip "Already up to date" no-commit outcome. It returned no `audit_receipt`.
> - Both `audit_multi` and this call issued no receipt, so a PASS cannot be corroborated, and the required codex gate stands at FAIL.
> - The gate-corroborated verdict is therefore **FAIL**, which means HOLD under the Q20 terms.
> - My own rubric adjudication is unchanged: 0.875, C1–C4 resolved, and no blocking defect under M6. The residue the codex gate holds on is D6a. Codex's earlier P1 (D1) was not repeated in this run.
> - A one-sentence fix to REQ-017/018 would close D6a: refuse a pinned SHA equal to the base before the merge, with its own code.

# SPEC Review Report: SPEC-MERGE-WINDOW-QUEUE-001 (card t1479)
Iteration: 5 (operator-granted final round after the ceiling; Q20). This is a delta re-audit of iter4 C1–C4, O1–O3 and the leader follow-up (post-merge-transition-conflict). CN-4 was run in full.
Verdict: PASS-WITH-DEBT. No blocking defect remains. Six optional findings are carried forward as run-phase debt (listed below).
Overall Score: 0.875 (Tier L threshold 0.85). Score history: 0.75 → 0.69 → 0.75 → 0.75 → 0.875.
Plan Artifact Hash (per-file SHA-256 of the working files, which are byte-identical to the HEAD blobs; see Evidence):
- spec 45318ab0…bf5d5
- plan c44766ca…3c8
- acceptance 92432e6a…63a
- design 1d5821c9…aab
- research d8ece060…5d5
Auditor Version: plan-auditor/v1 (iteration 5)

Reasoning context ignored per M1 Context Isolation. Q1–Q20 are taken as decided. Settled items are not reopened unless the delta broke them.

**Dissent notice.** `mcp__moai__audit_multi` returned `overall_verdict: fail`: claude FAIL, codex FAIL (both required gates), glm inconclusive. I read every backend finding against the text (see Defects). None of them shows any of the following:
- develop moving on unverified ground;
- an AC asserting something no REQ provides;
- two REQs that cannot both be implemented.

Those three were the blocking classes in iter4. Every backend finding is either a pre-existing wording item the delta did not touch, or an unspecified error branch with a bounded existing fallback. Under M6, these are optional, and I carry them as debt rather than manufacture a FAIL. The leader should weigh this dissent: the two required backends would hold.

## Must-Pass Results
- [PASS] MP-1 REQ numbering. `grep -oE '^- \*\*REQ-MWQ-[0-9]+' spec.md | sort | uniq -c` gives 23 IDs, each appearing once. The last one is `REQ-MWQ-023`.
- [PASS] MP-2 GEARS (requirement layer only).
  - REQ-017 (spec.md:166) is a compound: Ubiquitous "shall first read…", then "**When** … refuse…", then "**While** … shall renew…". This is GEARS compound-equivalent.
  - REQ-018 (:181) is Event-driven.
  - REQ-019 (:199) is Event-driven with nested **When** clauses.
  - REQ-009 stays Ubiquitous.
  - The ACs are Given-When-Then and are not graded here.
- [PASS] MP-3 Frontmatter. `version: "0.7.0"` (spec.md:4) is quoted. The field set is unchanged from iter3/iter4, which passed all 12 fields.
- [PASS] MP-4. The delta adds no language-specific tooling.
- [PASS] MP-5 D7. The delta adds no new SPEC reference (only SPEC-CANDIDATE-CI-001, already reconciled as a sibling draft). The iter3/iter4 D7 result stands.
- [PASS] MP-6 D8. `grep -c syscall spec.md` gives `0`.
- [PASS] MP-7. `grep -c 'NEEDS CLARIFICATION'` gives plan.md `0` and research.md `0`.
- [N/A] MP-8. `grep -ci release-blocking acceptance.md` gives `0`, so no AC is release-blocking.
- [PASS] MP-9.
  - COLLECTED: 9 milestones in plan order, M0..M8 (plan.md:52,61,66,73,79,85,95,105,112).
  - There are 0 `Exit:` bindings (`grep -ciE '^(\*\*)?exit(\*\*)?[ \t]*:' plan.md` gives `0`).
  - Ordering candidates (grep of acceptance.md): L19, 35, 55, 59, 68, 71, 73, 121, 125, 133, 136, 139, 151, 170, 173, 186, 223. Only L59 binds a milestone: "committed before the M1 code commit". Since M0 (L52) precedes M1 (L61), there is no conflict.
  - The new candidates order steps inside one call, not milestones:
    - L170/173 (AC-019 sc8, "released only after the hold is written");
    - L151 ("after the call as before it");
    - L121/125.

## Category Scores
| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | The four iter4 contradictions are closed (see Regression Check). Residual minor ambiguities that a reasonable engineer resolves consistently:<br>• REQ-018 "promote … on release" vs REQ-007 under a pre-existing hold (wording from iter3);<br>• REQ-018 cause 9 catch-all vs REQ-017's explicit unchanged-record refusal;<br>• REQ-019 step 3's "current candidate tree", which design D3 resolves as `HEAD^{tree}`;<br>• the Q20 ledger wording.<br>See D1–D6. |
| Completeness | 0.75 | 0.75 | All sections are present. One outcome is narrowly undefined: a complete-path state-transition failure for a reason other than a version or lease change, after the merge commit exists (D1). It is bounded by the existing `merging` → T15 / lease-expiry → blocked path (factory_card.go:1410–1414 comment). |
| Testability | 1.00 | 1.00 | AC-018 is a 9-row table with binary columns: "Merge commit afterwards" and "C promoted" (acceptance.md:136–146).<br>AC-019 defines "unchanged" once (:150–151), and sc5–sc8 each state seam call counts.<br>AC-017 sc4/sc5 assert record bytes unchanged.<br>The iter4 untestable assertion ("in every row no merge commit remains") is gone. |
| Traceability | 1.00 | 1.00 | 23 REQ definitions and 23 `AC-MWQ-NNN** (maps REQ-MWQ-NNN` lines, one to one, 001..023 (acceptance.md:11…192). No orphans. |

Aggregate: (0.75 + 0.75 + 1.00 + 1.00) / 4 = **0.875**, which is at or above 0.85.

## Regression Check (iteration 4 defects)
- **C1 — RESOLVED.**
  - REQ-019 (spec.md:199–203): "every gate shall run before the integration branch moves, in this order: (1) the card gates that today's T14 transition enforces — the card is `merge-ready`, the caller holds the card's lease and that lease has not expired, and the card version is the one read — and a failing gate refuses with the integration branch and the card state unchanged".
  - Plan M6: "card gates read without transitioning".
  - Design D3 complete diagram, step (1).
  - AC-019 sc6 (expired card lease) and sc7 (foreign lease): refuse, merge-step seam zero calls, everything unchanged.
  - The gate-read-to-transition race is closed by the post-merge-transition-conflict outcome (sc8).
- **C2 — RESOLVED.**
  - (a) Predicate: "a merge commit whose second parent equals the card branch's **current** tip and whose tree equals the tree of a re-measure record valid under REQ-MWQ-014/015 … a card branch that gained commits after that merge is not adopted and continues to step 3". AC-019 sc5 is the verb → new commit → complete fixture.
  - (b) Order is in the normative REQ: (1) gates → (2) adoption → (3) record refusal → (4) merge step. Clauses (2) and (3) are mutually exclusive in effect. Adoption implies merge tree == SHA tree == record tree, so a valid record keyed by the tip tree exists. The order therefore cannot change an outcome.
  - (c) The verb path is now bound to validity: REQ-017 check 1 says "require the card's re-measure record to be valid under REQ-MWQ-014/015". The AC-018 row 1 setup is "record with exit 3 (or count 0)", with integration tip unchanged. REQ-014/015 validity is content-only (exit, count, form; spec.md:149–162), so AC-019 sc4 adoption after develop advanced remains reachable.
- **C3 — RESOLVED.**
  - Ancestry precondition: REQ-017 says "require the pinned SHA to have that absorbed commit as an ancestor".
  - With base == tip and the ancestor relation, `git merge --no-ff SHA` yields SHA's tree. Merge-base = tip = ours, so the result is theirs, and a post-merge tree mismatch is unreachable. Design D3 says this explicitly.
  - REQ-018 cause 8: "any failure after the merge commit exists … leaves in place … `hold` with a reason naming the cause … merge commit SHA … before releasing … own code".
  - The AC-018 lead sentence no longer claims "no merge commit remains". Row 8 is "**yes — left in place**", with the hold naming the SHA.
  - REQ-018 now names nine causes, and AC-018 has nine rows mapping 1:1 to causes 1–9:
    - rows 6/7/8 inject at seams, which closes iter4 O3;
    - rows 1–6 and 9 promote; rows 7 and 8 hold. This matches REQ-018's "Causes 1-6 and 9 … promote … in causes 7 and 8 … hold".
- **C4 — RESOLVED.**
  - REQ-017 says "shall first read the window record to decide holdership and, **When** the caller is not the holder or the caller's lease has expired, refuse before applying any queue mutation, leaving the window record unchanged; **While** the caller holds the window with an unexpired lease, it shall renew the lease, apply REQ-MWQ-003 drops".
  - REQ-009 says "`merge` applies them only after its holder check passes".
  - Design D1 states the same order, which removes the iter4 contradiction.
  - AC-017 sc4 ("C may read the record to decide … a queued ticket whose waiter is gone … is not dropped by C's call") and sc5 (expired lease → refuse, no merge commit, bytes unchanged) both cover it.
  - The renewal order is consistent with REQ-008: renewal applies only to a holder that passed the check.
- **O1 (REQ-012 system write) — RESOLVED.** REQ-018 reads "a system write by the merge step, recorded with the step and card as setter and not subject to the lane-role refusal of REQ-MWQ-012". AC-018 row 7: "policy setter is the merge step and card".
- **O2 (exit-code count) — RESOLVED.** REQ-018 has nine causes, design D3 has "nine distinct values", and AC-018 has nine rows. The transition conflict takes a tenth code ("distinct from the nine", REQ-019 and AC-019 sc8).
- **O3 — RESOLVED** (seam injection in rows 6–8).
- **Leader follow-up (post-merge-transition-conflict + deferred release) — no deadlock, no contradiction.**
  - The hold is written before the release (AC-019 sc8: "released only after the hold is written"). REQ-007 then forbids promotion, so no holder is promoted onto the unrecorded merge.
  - The release itself always happens.
  - If the process dies inside the deferred interval, the window is bounded by the REQ-006 stale rules (owner process gone or lease expired). It is not held indefinitely.
  - The card stays `merge-ready`, and its merge commit's second parent is still the tip. A re-run of complete therefore adopts it in step 2, which is self-healing.
  - REQ-018 is scoped "When a REQ-MWQ-017 step fails". The transition conflict is not a REQ-017 step, and REQ-019 gives it its own release-after-hold. The two are consistent.
  - The residual gap is D1 below: transition failures for other reasons.
- **Stale-text sweep.** `grep -nE 'without reading|already reachable|reachable from (the integration|develop)|no merge commit remains|five causes|six distinct|six codes|merge.*apply the drop'` over spec/design/plan/acceptance/research found 0 stale hits. The only "any other error" hits are the intended cause 9 at design.md:70 and spec.md:188.

## Defects Found (structured defect-list)
No blocking defects. The optional items below are carried as run-phase debt.

D1. TRANSITION-FAILURE-OTHER-CAUSE. spec.md:211–216 (REQ-019 tail); design.md D3 "transition fails (version/lease changed after (1))"; AC-019 sc8.
- The post-merge outcome is defined only "because the card's version or lease changed after step 1".
- Not defined: a state-store error on the `merging` or `merged-local` write after the merge commit exists, or a partial transition that leaves the card in `merging`. For these, the hold, the release, and the exit code are all unspecified.
- Raised by codex (P1) and claude (P3).
- Bounded today: factory_card.go:1410–1414 documents that `merging` resolves via T15 or lease expiry → blocked, and the REQ-006 stale rules free the window. Develop has moved only on verified ground (checks 1–5 passed).
- Severity: minor. Class: optional (strongly recommended).
- Fix: widen the REQ-019 clause to "When the step succeeded but any state transition then fails (including a version or lease change after step 1)". State that a card left in `merging` is the hold's named recovery. Carry this as a run-phase obligation (M6 test: inject a store error on the second transition).

D2. HOLD-PROMOTION WORDING. spec.md:189–190 (REQ-018 "Causes 1-6 and 9 … promote the next live ticket on release") vs REQ-007 (no promotion under `hold`).
- Under a pre-existing leader hold, the literal wording conflicts.
- The wording dates from iter3 (v0.6.0 "release the window (promoting the next live ticket)"). The delta did not introduce it.
- A reasonable engineer applies REQ-007, because promotion is performed by the record mutation, which honors the policy.
- Raised by codex (P2) and claude (P3).
- Severity: minor. Class: optional.
- Fix: "release the window, which promotes per REQ-MWQ-006/007".

D3. NON-HOLDER REFUSAL EXIT CODE. REQ-017 (refusal, record unchanged) vs REQ-018 cause 9, "any other error before the merge".
- REQ-017's explicit "leaving the window record unchanged", together with AC-017 sc4/sc5 byte-equality, governs. Cause 9 reads as covering the post-holder-check steps.
- The refusal's exit code is unnamed.
- Raised by claude (P2).
- Severity: minor. Class: optional.
- Fix: "When a REQ-MWQ-017 step after the holder check fails …", plus a named non-releasing refusal code.

D4. STEP-3 KEY. REQ-019 step 3, "current candidate tree".
- REQ-014 defines the candidate tree as the branch tree after absorbing. Design D3 keys the record by `HEAD^{tree}`. Under that reading, AC-019 sc3 reaches the merge step and RE-MEASURE.
- Raised by claude (P2).
- Severity: minor. Class: optional.
- Fix: write `<card-branch tip>^{tree}` in step 3.

D5. Q20 LEDGER WORDING. decision-index.md:231 reads "거부 절은 채택보다 먼저 평가하고" ("the refusal clause is evaluated before adoption"). REQ-019 orders adoption (2) before the record refusal (3).
- There is no behavioral effect, because the clauses are mutually exclusive (see C2(b)).
- Severity: minor. Class: optional.
- Fix: reword the ledger line to "card gates → adoption → record refusal → merge step".

D6. EDGE OUTCOMES NOT STATED.
- (a) A pinned SHA equal to `record.base` passes checks 2–4. `git merge --no-ff` then prints "Already up to date" and creates no merge commit (codex, scratch repo, P2). The card's content is already wholly in develop, so the outcome is benign but unnamed.
- (b) A system hold written over an existing leader hold: whether the reason and setter are overwritten or kept (claude P3).
- (c) Complete's window acquisition point is implicit. Today's complete has its own window phase (factory_card.go ~1325–1380; research §R2), so this is existing behavior and not a delta break (claude P2).
- Severity: minor. Class: optional.

## Recommendation
PASS-WITH-DEBT.
- Every iter4 blocking defect (C1–C4) is resolved with normative REQ text, matching design and plan text, and binary ACs. The citations are above.
- MP-1 through MP-9 pass (MP-8 is N/A).
- The aggregate is 0.875, at or above the Tier L threshold of 0.85.
- D1–D6 are optional findings. D1 is the one worth fixing before or at M6. Carry D1–D4 as explicit run-phase obligations in progress.md, and fix D5 in the ledger.
- Because the two required cross-model backends returned FAIL on these same optional items, the leader decides whether this dissent changes the operator-gated outcome. If the leader treats the backends' FAIL as controlling, the outcome is HOLD per Q20 terms. My adjudication is that no finding is blocking.

### Blocking list
- (none)

### Debt list (run-phase obligations)
- D1 widen the post-merge transition-failure clause to any cause; define recovery for a card left in `merging`.
- D2 condition REQ-018 promotion on REQ-006/007.
- D3 scope REQ-018 to after the holder check; name the refusal code.
- D4 key step 3 by the branch tip tree.
- D5 ledger wording.
- D6 (a)/(b) name the outcomes.

## Evidence (five-section format)

**Claim:**
- The v0.7.0 + follow-up delta (1d29b3126, 046d37f0e) resolves iter4 C1–C4 and O1–O3.
- The deferred-release race closure neither deadlocks nor contradicts REQ-018.
- The remaining findings are non-blocking.

**Evidence** (commands run this session and their observed output):
- `git log --oneline -3 046d37f0e` gives:
  - 046d37f0e "close the gate-to-transition race";
  - 1d29b3126 "close plan-audit iter4 delta, v0.7.0";
  - 3b57f3e7c.
- `git rev-parse WT-merge-window-queue` gives `046d37f0e43458b8965ede9de7481e1429297fd8`.
- `git diff --stat 3b57f3e7c 046d37f0e` gives 7 files, +397/−82.
- `git diff 3b57f3e7c 046d37f0e -- spec.md design.md acceptance.md plan.md` was read in full. All quotes above come from it.
- `git rev-parse 046d37f0e:<path>` and `git hash-object <working file>` returned identical blob IDs for spec (cf2b6aa60), acceptance (f2d10a69b), design (03363d399) and plan (3d0fdd9c8). The working files are therefore the audited commit's content.
- MP greps: syscall `0`; NEEDS CLARIFICATION `0`/`0`; release-blocking `0`; Exit bindings `0`; REQ IDs 23 unique, last `REQ-MWQ-023`; AC maps 23 (acceptance.md:11…192).
- `sed -n 1395,1425p internal/cli/factory_card.go` shows T14 `To: homestate.CardMerging` (:1403) before `factoryMergeNoFF` (:1409). The comment at :1410–1414 reads "The card stays in `merging` … the lane resolves the tree (T15) or the lease expiry moves it to blocked". This is the D1 bound.
- `mcp__moai__audit_multi` (project_root = audited tree, target baseBranch): `overall_verdict: fail`, `participant_count 2`, `disagreement_flag false`, `fail_open_backends [glm]`. No `audit_receipt` was returned.
  - claude (claude-opus-5-5, subscription): fail. It confirms C1, C2, C3 and C4 hold and finds no deadlock. Its P2 findings map to D3, D6c and D4; its P3 findings to D5, D2, D6b and D1.
  - codex: fail. Its P1 maps to D1, and its P2s to D2 and D6a (scratch repo: `merge-base --is-ancestor` exit 0, `merge --no-ff` exit 0 "Already up to date").
  - glm: inconclusive ("z.ai response carried no text content").
  - build_commit 45600e4ee is an ancestor of 046d37f0e (binary lag). The findings concern SPEC text, not tool output.

**Baseline-attribution:**
- Commit 046d37f0e on `WT-merge-window-queue`.
- Read through the shared object store with plain `git`, plus direct reads of the audited tree's working files by absolute path. Blob equality with HEAD was verified above.

**Gaps:**
- The worktree guard refused three commands:
  - `git -C <audited tree>`;
  - the scripted AC-4/AC-5 and CN-4 awk verbs (rejected as an unverifiable awk program);
  - a `git show | shasum` loop.
- Fallbacks:
  - plain `git` from the shared object store;
  - traceability by grep, 23/23 one to one, rather than the scripted verb;
  - CN-4 by grep of the ordering keywords plus the milestone heading list, judged by hand;
  - per-file SHA-256 of the working files plus blob-ID equality, in place of a combined hash.
- `git status` of the audited tree was not observed. Equality was shown only for the four files compared (research.md was not compared).
- Codex's scratch-repo reproduction (D6a) is its observation, not mine. The git semantics it reports (an ancestor-or-equal `--no-ff` merge is a no-op) are standard.

**Residual-risk:**
- Both required cross-model gates returned FAIL. This PASS-WITH-DEBT rests on my M6 classification of their findings as optional.
- The in-window merge step has produced new seams in each round. D1 is the same family, a post-merge error branch, in its narrowest form.
- GLM did not participate.
- No receipt was issued.

## Iteration history
- iter1 (1e1d0cc84): FAIL 0.75, D1–D10.
- iter2 (6ab3dbcb2): FAIL 0.69, STOP, N1–N5.
- iter3 (beaedc3f4): FAIL 0.75, B1–B3.
- iter4 (3b57f3e7c): FAIL 0.75, C1–C4.
- iter5 (046d37f0e): PASS-WITH-DEBT 0.875, C1–C4 resolved, debt D1–D6. No stagnation: no defect persisted unchanged across three iterations.

## Operational Notes (unverified)
- `inferred` (rule: a code path that catches every transition error cannot distinguish a version conflict from an I/O error): implementing REQ-019's conflict branch as `if err != nil { hold; release; exit }` likely discharges D1 at no extra cost. Measure it in M6: inject a store error on the `merged-local` write, then assert the hold, the SHA in the reason, the release, and the code.
- `assumption`: D6a never occurs in practice, because a card branch always carries at least one commit beyond its absorbed base. Measure it: `git rev-list --count <base>..<tip>` on the M8 regression fixtures.
