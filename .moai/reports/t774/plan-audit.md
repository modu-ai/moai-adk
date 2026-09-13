# SPEC Review Report: SPEC-HEAVY-TEST-SLOT-001 (card t774)

Iteration: 1/1 (Tier S ceiling per `.moai/config/sections/harness.yaml:76` — single-pass audit; this defect list is the machine-consumable fix route)
Verdict: **FAIL**
Overall Score: **0.70** (mean of dimension scores; harmonic 0.65) — below the Tier S threshold 0.75 (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier) and below the dispatcher's 0.85 bar. D1 is independently verdict-forcing.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-HTS-001..004 sequential, no gaps/dups (spec.md:43-46); AC-HTS-001..005 sequential (acceptance.md:14-22).
- **[PASS] MP-2 EARS format compliance** (judged on the REQ layer of spec.md §C only) — all four REQs are EARS shall-forms: REQ-HTS-001 event-driven ("When a lane or the lead is about to run …, the system shall provide …", spec.md:43); REQ-HTS-002/003/004 ubiquitous shall / shall-not (spec.md:44-46). Legacy EARS within the backward-compat window; no informal "should/may" in normative text.
- **[PASS] MP-3 YAML frontmatter validity (spec.md)** — all 12 canonical fields present, no rejected snake_case aliases (spec.md:2-14); `phase: "batch 9"` is not a prohibited whole-value (plan/run/sync/mx); `tier: S` present. Note: the sibling-artifact `status:` violation is NOT absorbed here — reported as D2 (blocking).
- **[N/A] MP-4 language neutrality** — single-purpose dev-only protocol doc; no multi-language tooling content. Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep 'SPEC-[A-Z][A-Z0-9]*-[0-9]'` over all three artifacts: 0 matches; no referenced SPEC to reconcile. (Irony noted: the *missing* reference — SPEC-RESOURCE-SLOT-LEASE-001, which owns the pre-existing slot docs — is the root cause of D1/D4.)
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` count 0 in all three artifacts; auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over the SPEC dir: 0 matches; research.md absent (Tier S) → N/A.
- **[FAIL — mechanical, verdict-forcing] spec-lint OutOfScopeRule** — see D6: `internal/spec/lint.go:1292-1334` requires an `###` H3 heading containing "out of scope" with a `-` bullet under it; spec.md's only H3s are B.1/B.2/B.3 (spec.md:23,27,35) and the literal appears only as body text (spec.md:51). `MissingExclusions` at SeverityError fires on this modern-era (V3R6) SPEC and is not era-demoted (`eraDemotableCodes` exempts grandfathered SPECs only).

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Testability | 0.50 | Band 0.50 | AC-1/2/4 are binary-testable greps; AC-3's `= 0` predicate is unsatisfiable under the plan (D1); AC-5's literal base pin false-FAILs after a develop absorb (D3); AC-1's greps under-verify the WHEN rule (N1). 2 of 5 ACs cannot deliver a correct verdict as written. |
| Traceability | 1.00 | Band 1.00 | Every REQ has ≥1 AC and every AC traces to an existing REQ: REQ-001→AC-001, REQ-002→AC-002, REQ-003→AC-003, REQ-004→AC-004+AC-005 (acceptance.md:14-22). No orphans. |
| Completeness | 0.75 | Band 0.75 | HISTORY/Problem/Requirements/Non-Goals/Delivery present; spec frontmatter 12/12. Gaps: cited evidence path dangling (D5), pre-existing-surface survey missing (D4), Out-of-Scope H3 missing (D6). |
| Consistency | 0.50 | Band 0.50 | AC-HTS-003 (=0 whole-file) contradicts plan.md:27's add-only step; spec.md:31 "is still missing" is false against the committed tree (§8 landed 09-12, two days before authoring); `tier: S` vs 3-artifact set (N2). Cross-artifact paths/counts themselves are consistent. |
| Clarity | 0.75 | Band 0.75 | Requirements are precise and measurable except REQ-HTS-003, which a reasonable implementer can read two ways (add a pointer alongside §8's existing inline commands, vs replace them) — the two readings produce opposite AC-3 outcomes. |

**Verdict driver**: D1 (AC unsatisfiable under its own plan) + D6 (mechanical spec-lint Error). Aggregate 0.70 < both thresholds independently.

## Defects Found

- **D1. AC-HTS-003 zero-count unsatisfiable** — acceptance.md:19 vs `.claude/rules/local/gitflow-lane-protocol.md` §8 — AC verifies `grep -c 'slot acquire' <rule>` = 0, but the rule **already contains** `moai slot acquire --resource <이름> --max-duration <상한>` exactly once (§8, committed `e78fd0ee6`, 2026-09-12 09:49 KST, card t607 / SPEC-RESOURCE-SLOT-LEASE-001 M6; measured count: 1). plan.md:27 says "Add a pointer paragraph" and never edits §8, so implementing the SPEC exactly as planned **fails its own AC** — red at arrival and still red after the plan's work. Severity: critical — Class: blocking — Required fix: either (a) scope the zero-count to the added pointer paragraph (not the whole file), or (b) plan to replace §8's inline commands with the pointer (consolidation), stating that edit explicitly.
- **D2. Sibling artifacts carry `status:`** — plan.md:5, acceptance.md:5 — both carry `status: draft`, violating `.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness ("the four sibling artifacts … MUST NOT carry a `status:` field"; lifecycle state lives in exactly one place, spec.md). Severity: major — Class: blocking — Required fix: delete the `status:` line from both frontmatter blocks.
- **D3. AC-HTS-005 pins a literal base SHA** — acceptance.md:22-23 — `git diff d416f8162 <HEAD> -- internal/cli/slot.go` breaks after the lane absorbs `origin/develop` (mandatory per gitflow §4.1): an unrelated card's slot.go change in develop makes the diff non-empty → false FAIL. This is the exact defect class codified as [HARD] in §8 of the very rule this SPEC edits (merge-base, not literal pin; measured record `.moai/reports/t543/verdict.md`). Severity: major — Class: blocking — Required fix: `git diff "$(git merge-base origin/develop HEAD)" HEAD -- internal/cli/slot.go` (or pin the merge-base SHA at run time).
- **D4. Problem statement stale; deliverable not reconciled with the two existing visibility artifacts** — spec.md:31, :43 — B.2(a) claims the binding procedure text "is still missing". False on the base tree: (i) gitflow-lane-protocol.md §8 already carries the WHEN duty, all three commands, the slot-vs-integration distinction, and a pointer to the full surface doc; (ii) `.claude/rules/moai/workflow/resource-slot-lease.md` exists (3,975 B) **and is template-mirrored** (`internal/template/templates/…/resource-slot-lease.md` — it ships to deployment users). Both landed 2026-09-12 (card t607), before this SPEC's 09-14 authoring. The new doc's genuinely additive content shrinks to three items: the exit-3 fact, the named heavy-package list, and the control-group record (§8 already has WHEN, commands, and the integration distinction). The plan as written creates a **third copy** of the procedure — the exact divergence hazard REQ-HTS-003 itself warns against ("two copies diverge"), now self-inflicted. Severity: major — Class: blocking — Required fix: re-shape the deliverable against the existing pair — extend resource-slot-lease.md or §8 with the three additive items, or explicitly declare the three-artifact relationship and one canonical home — and cite SPEC-RESOURCE-SLOT-LEASE-001 in §A/§B.
- **D5. Cited evidence path does not resolve** — spec.md:19 — §A attributes the plan's measurements to `.moai/reports/t774/verdict.md` §2; the directory is absent in **both** the worktree and the primary checkout (verified). The substance is true (I re-verified it independently — see below), but per VCI §2 the citation is currently unattributable. Severity: major — Class: blocking — Required fix: cite `.moai/logs/slot-lease-audit.jsonl` directly (first entry 2026-09-12T12:28:42Z `heavy-test`; `t774-repro-demo` acquire/3×refuse/release 2026-09-13T17:07Z) or persist the investigation record at the cited path before run-phase.
- **D6. MissingExclusions Error (mechanical gate)** — spec.md:48-52 — `## §D — Non-Goals` does not satisfy `internal/spec/lint.go:1314` (`###` heading containing "out of scope" + `-` bullet beneath). Finding `MissingExclusions` at SeverityError fires; this SPEC is V3R6-modern (created 2026-09-14), so the era demotion does not apply and the lint exit code fails. Severity: major — Class: blocking — Required fix: add an `### Out of Scope — heavy-test slot discipline` H3 (with the three bullets under it) or retitle §D accordingly.

Non-blocking residues:

- **N1.** AC-HTS-001's greps under-verify the WHEN rule — a doc naming `internal/cli` once without any WHEN rule passes all four verifies (mutant writable per verification-completeness §2); `grep -c 'slot acquire --resource' = 1` is brittle (spurious fail if the command appears in an example). Suggest an anchored token (e.g., a literal `WHEN` heading, count ≥ 1) and `≥ 1` instead of `= 1`. — Class: optional
- **N2.** `tier: S` declares a 2-artifact set (AC inline in spec.md §3 per spec-workflow § SPEC Complexity Tier) but the SPEC ships 3 artifacts with acceptance.md. Either declare `tier: M` or fold ACs inline. — Class: optional
- **N3.** "19 events to date" (spec.md:19) is a point-in-time figure; the log now holds 25 (live growth: heavy-test×14, t675-cli-tests×4, gotest-gateway×2, t774-repro-demo×5 — internally consistent with the claim at measurement time). Prefer timestamped citations over bare counts. — Class: optional
- **N4.** AC-HTS-002's `재현|re-demonstrate|재시연` alternation can match incidental Korean prose; acceptable directional proxy, note only. — Class: optional
- **N5.** Timezone hazard: the demo log entries read 2026-09-13T17:07Z = 2026-09-14 02:07 KST; the SPEC's "measured 2026-09-14" is local time. Consistent, but record the timezone to prevent a future false flag. — Class: optional

## Card Faithfulness (dispatcher's judgment items)

- **(a) separate surface needed?** — Answered by measurement, direction correct: the surface exists, is registered (`slot.go:328`), enforced (`slotExitHeld = 3`, slot.go:37; refusal demonstrated live — 3 `refuse` events under two second-session ids in the audit log), and in active use (25 events, first 2026-09-12T12:28:42Z — postdating the 09-10 card, as claimed). The shrink-to-doc branch of the dispatch conditional correctly fires. **However** the "visibility doc" landed as a parallel third document instead of consolidating with the two existing visibility artifacts (D4).
- **(b) `moai integration` reuse?** — Correctly rejected; the merge-window vs execution-order distinction is real and independently corroborated (§8 states the same separation). Faithful.
- **(c) recorded fields?** — Verified in source: status renders session+pid, name, command, since, bound/ends (slot.go:211) — exactly the card's field list. Faithful.
- **[HARD] control-group obligation** — Satisfied in substance: the 09-10 incident stands as the no-surface control (recorded; not re-runnable — correctly forbidden by the repo's load discipline, which a fresh concurrent-600s reproduction would violate), and the exit-3 refusal demo is the enforcement evidence, mechanically corroborated by 5 audit-log entries. The reframing's treatment of this obligation is proportionate and honest (REQ-HTS-002 explicitly forbids the forbidden re-run).
- **Scope discipline** — The dev-only claim holds: `.moai/docs/` and `.claude/rules/local/` are outside the `CleanMoaiManagedPaths` wipe roots; plan.md §C's neutrality note is accurate; AC-HTS-004's `docs-site/` prefix is meaningful (the directory exists at repo root). `moai slot` the product verb is untouched (AC-HTS-005 intent is right; only its verification form is defective — D3).

## Independently Re-verified vs Taken from the Plan's Record

**Independently re-verified (this run, this tree — commands + observed outputs):**

1. `head -1 /Users/goos/MoAI/moai-adk-go/.moai/logs/slot-lease-audit.jsonl` → `{"ts":"2026-09-12T12:28:42Z","event":"acquire","resource":"heavy-test","session_id":"da011b34-…"}` — matches spec.md:19.
2. `wc -l` + per-resource count on the same log → 25 events (heavy-test×14, t675-cli-tests×4, gotest-gateway×2, t774-repro-demo×5).
3. `grep 't774-repro-demo'` on the log → 5 entries (acquire; refuse×2 session t774-demo-b; refuse t774-demo-c; release; holder pid 54135) at 2026-09-13T17:07Z — corroborates the enforcement demo.
4. `grep -n` on `internal/cli/slot.go` → `rootCmd.AddCommand(newSlotCmd())` (:328); acquire/status/release subcommands (:103); status fields session/pid/name/command/since/bound-ends (:211); `slotExitHeld = 3` (:37).
5. `grep -c 'slot acquire' .claude/rules/local/gitflow-lane-protocol.md` → **1** (§8) — the D1 measurement.
6. `ls` → `.claude/rules/moai/workflow/resource-slot-lease.md` exists (3,975 B) **and** its template mirror exists — the D4 measurement.
7. `git log -1 --format='%H %ad %s' e78fd0ee6` → 2026-09-12 09:49:24 +0900, "docs(SPEC-RESOURCE-SLOT-LEASE-001): M6 lane docs name the slot lease (card t607)" — the D4 chronology.
8. `ls .moai/reports/t774/` → absent in worktree **and** primary — the D5 measurement.
9. Grep sweeps → SPEC-ID refs 0 (D7), `syscall` 0 (D8), `[NEEDS CLARIFICATION` 0 (MP-7).
10. `internal/spec/lint.go:1292-1334` read → OutOfScopeRule requires `### … out of scope` H3 + bullet; spec.md has no such H3 — the D6 measurement.
11. `ls -d docs-site` → exists (AC-HTS-004's prefix check is meaningful); `.moai/docs/heavy-test-slot-protocol.md` absent (correct pre-run state).

**Taken from the plan's record (not independently re-verifiable):** the 09-10 incident observations themselves (three lanes, load 8–21, flipped verdicts — historical; re-running is correctly forbidden by load discipline); the 19-event composition at measurement time (consistent with the 25 observed now); the verbatim exit-3 code at demo time (source constant verified; log records `refuse` events, not exit codes).

## Recommendation (numbered fix route for the author)

1. Fix D1 first (it defines the deliverable's shape): decide consolidation vs parallel-doc, then make AC-HTS-003's zero-count check match that decision.
2. Fold D4's reconciliation into the same edit: cite SPEC-RESOURCE-SLOT-LEASE-001; either extend the existing resource-slot-lease.md/§8 with the three genuinely additive items (exit-3 fact, named package list, control-group record) or declare the canonical-home relationship explicitly.
3. Apply D2 (delete two `status:` lines), D3 (merge-base diff form), D5 (resolvable citations), D6 (Out-of-Scope H3).
4. Sweep N1-N5 opportunistically; none blocks on its own.

Ceiling note: Tier S ceiling is 1 (harness.yaml:76), so this is the single plan-audit pass for this SPEC. On ceiling-exit the orchestrator owns the disposition (fix consumed by the run-gate's re-check, PASS-with-debt, or explicit override) — the defect list above is the fix route; verdict authority for any confirming re-audit stays with this agent role.
