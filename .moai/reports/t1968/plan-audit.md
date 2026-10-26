auditor-model: glm-5.3

# SPEC Review Report: SPEC-FETCH-WEB-001 (card t1968)

Iteration: 1/3
Verdict: **FAIL**
Overall Score: 0.80 (Tier M threshold 0.80 — score meets threshold; verdict FAIL on 3 blocking findings + required-backend convergence fail)
Plan Artifact Hash: `b8d746adde8f12d668f06ca33c624af15c48953a44ec76144b8072dee315b1e6` (sha256 over spec.md+plan.md+acceptance.md+progress.md concatenated in that order)
Auditor Version: plan-auditor (factory depth-1 leaf, card t1968)

```
verdict: FAIL
audited_sha: bdc9a86ca13bc323dfd917001bbd97b6b06352d0
overall_score: 0.80
must_pass_failed: 0
blocking_count: 3
plan_artifact_hash: b8d746adde8f12d668f06ca33c624af15c48953a44ec76144b8072dee315b1e6
scope: full
fix_scope: .moai/specs/SPEC-FETCH-WEB-001/spec.md#REQ-FW-006 + acceptance.md#AC-FW-004 (no_match decision order); .moai/specs/SPEC-FETCH-WEB-001/spec.md#REQ-FW-015 + acceptance.md#AC-FW-013 (enabled-true scope); .moai/specs/SPEC-FETCH-WEB-001/spec.md#REQ-FW-012 + acceptance.md#AC-FW-010 (retry x Repeatable composition)
defect_class: design (D1), ac-wording (D2), design (D3)
convergence_overall: fail
required_backend: codex fail
```

Cross-model receipt: `rcpt-4be9e80d666fee983a5eccb3` (codex, required gate; `moai verify audit-plan` convergence_check `ok: true` — the required backend answered; its verdict is fail and this auditor independently confirmed the findings, see D1-D7).

## Claim

1. All nine must-pass criteria pass locally (evidence below).
2. Three blocking defects exist in the SPEC/acceptance pair (D1-D3), four should-fix (D4-D7), two minor (D8-D9).
3. The required codex backend independently returned fail with 7 findings; 7/7 were adjudicated by this auditor against the tree, all confirmed in substance (two with corrected framing, see D5/D7 notes).
4. Design-source fidelity to the leader analysis is otherwise complete: applications 1-4, the not-porting list (Python deps, headless render with t1965 linkage), fixed naming, thresholds, and carried Gaps all map to SPEC anchors with no drift found.

## Evidence (commands run in this audit, this tree)

- Traceability verb: `COLLECTED: 16 REQ definitions (acceptance input: read)` — zero UNCOVERED, zero ORPHAN lines.
- CN-4 verb: `COLLECTED: 0 milestones in plan order (none), 0 exit bindings, 4 ordering candidates` + `GAP: 0 milestone headings collected` (plan uses bold bullets); hand-read M1→M7 order adjudicates all 4 candidates non-ordering (two are content words "first block"/"first output line", one is the header name `Retry-After` matching the `after` keyword, one is "after M7" which matches the plan's own last position). No CONFLICT.
- D8: `grep -c syscall` = 0 in spec.md, plan.md, acceptance.md (auto-pass per D8-4).
- D7: referenced SPEC-JEV-CORE-001 / SPEC-JEV-OPTIN-MEASURE-001 / SPEC-JEV-CONSUMERS-001 all exist, `status: completed` each (no REVIEW, no BLOCKING).
- MP-7: `grep -rn "NEEDS CLARIFICATION" plan.md research.md` → no match in plan.md; research.md absent (Tier M artifact set — expected).
- REQ enumeration: REQ-FW-001…016, 16 unique, sequential, no gaps/duplicates. AC enumeration: AC-FW-001…014, 14 unique.
- Tree facts (plan §A.1 spot-checks, all confirmed): `EstimateTokens` at internal/jev/jev.go:369; `New(enabled bool)`:343, `Ask`:538, `UsageLog`:352, `Result.OK`:288, `NoticeLine`:293; `WorkflowJevConfig` at internal/config/types.go:1059 with bare `enabled` gate (:694 comment confirms default-false INERT semantics); `"jev_ask": true` at internal/cli/mcp_server.go:393; `Use: "web [--port N] [--no-open]"` at internal/cli/web.go:35; internal/jevmeasure/measure.go exists; `internal/template/templates/.claude/settings.json.tmpl` exists (corroborates R7's side-observation); no `internal/cli/fetch*.go`, no `internal/fetchweb`.
- D2 evidence: `grep -rn "enabled: true" internal/template/templates/` → **127 occurrences** (e.g. `.moai/config/sections/harness.yaml` `auto_detection.enabled: true`, lsp.yaml.tmpl, interview.yaml, workflow.yaml).
- D3 evidence: internal/jev/jev.go:203-209 `Repeatable bool` (used at :575); three live tests today `TestRetry_NotIssuedForANonRepeatableRequest`, `TestRetry_BoundedForARepeatableRequestThatFailedBeforeDelivery`, `TestRetry_NotIssuedForAnAmbiguousFailure` (jev_test.go:246/260/278); SPEC-JEV-CORE-001 REQ-JEVC-008: "**Where** a call failed non-idempotently or ambiguously, **the `internal/jev` package shall not** retry it." `grep -rn "Retry-After" internal/jev` = 0 hits today (the Retry-After parsing is genuinely new — that part of the plan's tree facts holds).
- D5 evidence: SPEC-JEV-CORE-001 REQ-JEVC-007: "Every consumer of an unavailable result shall emit at most one notice line, exit 0, and continue the workflow it was performing."
- D7 evidence: t1762 commit `954adab25e` — SPEC-JEV-LEDGER-001 `module: "internal/jevledger (new) + internal/config + internal/template"`, and its line 102: "This SPEC does not edit `internal/jev/jev.go` or any other file in `internal/jev`. The recorder lives in its own package."
- Cross-model: `mcp__moai__audit_multi` (project_root = this worktree, no gates override; tree plan applied: codex required, claude/glm off) → codex verdict **fail**, 7 findings; receipt `rcpt-4be9e80d666fee983a5eccb3`; `moai verify audit-plan --result-file .moai/state/audit-plan-result.json` → `convergence_check: {ok: true, unmet: []}`.

## Baseline-attribution

All measurements above were taken in this run, in the worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1968`. HEAD at read time: `242e35b8027b4cc239204c4041ba9f0bbd5d3dbc` ("docs(t1968): record plan-stage landing and audit spawn"). The task-bound audited head is `bdc9a86ca13bc323dfd917001bbd97b6b06352d0`; `git diff --stat bdc9a86ca1..HEAD -- .moai/specs/SPEC-FETCH-WEB-001` is **empty** (the extra commit touches only `.moai/reports/t1968/progress.md`, the lane's own pre-spawn bookkeeping — not a foreign commit inside the audit window). SPEC content audited is byte-identical at both commits; `audited_sha` carries the recorded binding. RED ledger rows R1-R7 were **read, not re-executed**, per the MP-8 execution discipline; their pin `ec9091a80c090132ae4dda6c0588b9810c20de53` resolves in this repository.

## Gaps

- MP-8 boundary: reading the RED cells confirms what they record; it does not confirm the commands measure the premise their author claims (D4 records where one does not).
- The codex leg's internal claim of having "enumerated 510 selection combinations" for D1 was not re-executed by this auditor; D1 is confirmed on textual analysis alone (the branch is unreachable as written), so nothing rests on that enumeration.
- t1762 is in-flight in another lane; its shape was read from commit `954adab25e` in the shared object store, which may be superseded again before t1968 lands — the boundary fix (D7) should re-read the then-current t1762 artifact at revision time.

## Residual-risk

- B6's bench command name is left to the M6 delegation; two spellings would both satisfy the AC.
- The off-gate exit code (B1) has no existing consumer to regress (new command), but scripts written against exit 2 between run and any later change would bind — pin it once, in this revision, via D5's reconciliation sentence.
- `moai fetch web` string-grep (R2) is a starting-absence observation, not a flip-verifier (cobra `Use:` strings may never contain the literal); AC-FW-005's built-binary `--help` method is the correct green check, as already written.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: 16/16 sequential unique (grep enumeration above).
- [PASS] MP-2 EARS/GEARS compliance (requirement layer only; ACs are Given-When-Then verification layer per M3 § Scope): all 16 REQs use Ubiquitous/When/While/Where forms or compound chains (REQ-FW-003/008 While; REQ-FW-012 When; REQ-FW-015 Where; REQ-FW-010/016 negative legacy-equivalent form, within the compat window through 2026-11-22).
- [PASS] MP-3 YAML frontmatter: 12/12 canonical fields present, correct types, no snake_case aliases, `status: draft` in enum, `priority: P1`, `phase: "v3.3.0 target"` is a release label (not a prohibited stage word), `tier: M`, `depends_on` well-formed.
- [N/A] MP-4 language neutrality: single-language (Go) project scope; no multi-language tool enumeration.
- [PASS] MP-5 D7 cross-SPEC reconciliation: 3 external references, all exist, all `completed`; no REVIEW/BLOCKING.
- [PASS] MP-6 D8 cross-platform discipline: `syscall` occurrences 0 in all three artifacts (auto-pass).
- [PASS] MP-7 clarification gate: plan.md read, no `[NEEDS CLARIFICATION]`; research.md absent (Tier M — stated reason).
- [PASS] MP-8 RED-now completeness: all 14 release-blocking ACs carry command + verbatim output + exit code + tree SHA (document-level pin inherited from §G; per-criterion citation by row id; empty-stdout exit-1 grep rows are complete records). No re-execution performed. (Right-reason quality of two rows is graded as D4, not as an MP-8 completeness failure, per the MP-8 boundary clause.)
- [PASS] MP-9 cross-artifact ordering: verb GAP (bold-bullet milestones) → hand-read plan order M1→M7; 4 candidates adjudicated non-ordering; no CONFLICT; every green-path milestone binding (M1..M7) sits on the correct side.

## Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.70 | below 0.75 | D1 (no_match branch unreachable as written), D6 (cache-key mechanism unspecified) — two interpretive clusters; remainder crisp and numeric |
| Completeness | 0.90 | near 1.0 | all sections present, 5 specific Out-of-Scope H3 subsections, Gaps carried; deductions: D6 mechanism gap, D7 stale §A.2 premise |
| Testability | 0.70 | below 0.75 | D2 (criterion false-at-arrival under literal reading), D4 (non-discriminating RED basis for AC-FW-010/011), D1 (branch not binary-testable as written) |
| Traceability | 0.90 | near 1.0 | verb clean 16/16, 0 orphan/0 uncovered; deductions: AC-FW-014 does not exercise REQ-FW-016's boundary (D7 limb) |

## Defects Found

- **D1 [BLOCKING, class design]** spec.md §C.2 REQ-FW-006 vs REQ-FW-007 / acceptance.md:15 AC-FW-004 — REQ-FW-006 keeps the first block **unconditionally**, so on any non-empty page the kept set is never empty and the `no_match` status ("no block survives selection") is unreachable as written; the SPEC never states where no_match is decided relative to forced-keep/backfill. Repair: pin the decision order in REQ-FW-006 or REQ-FW-007 — e.g. "`no_match` is emitted when zero blocks become stage-1 candidates; the first-block keep and 300-token backfill apply only once at least one candidate exists" — and assert both branches in AC-FW-004 (a zero-candidate non-empty page → `no_match` exit 1; a low-score page with ≥1 candidate → first block present, status `filtered`).
- **D2 [BLOCKING, class ac-wording]** spec.md:80 REQ-FW-015 + acceptance.md:41 AC-FW-013 — "no `enabled: true` anywhere under `internal/template/templates/`" is **false at arrival**: the tree already carries 127 occurrences (harness.yaml `auto_detection.enabled: true`, lsp, interview, workflow…), none of which this SPEC may remove. Repair: scope the prohibition to what this SPEC controls — "this SPEC's shipped artifacts introduce no `enabled: true` (in particular not `workflow.jev.enabled`) and no machine-specific values; the existing template-neutrality CI guard stays green" — in both REQ-FW-015 and AC-FW-013.
- **D3 [BLOCKING, class design]** spec.md:71 REQ-FW-012 vs SPEC-JEV-CORE-001 REQ-JEVC-008 — as worded, the retry mandate ("When the endpoint responds 429/529/5xx with Retry-After … the client shall … retry") overrides the completed dependency's duplicate-prevention contract (`Repeatable` gating; `internal/jev` already ships three passing `TestRetry_*` safeguards). Repair: compose — "…shall delay the retry by the parsed amount, **within the retry conditions REQ-JEVC-008 already permits** (only calls whose repetition is provably free of side effects)" — and add negative cases to AC-FW-010 (non-repeatable request not retried; ambiguous post-delivery 5xx not retried).
- **D4 [SHOULD-FIX, class ac-wording]** acceptance.md §G R4 (cited by AC-FW-010/011) — `grep "cache/jev"` measures cache-string absence, not backoff/semaphore/usage-separation absence: a mutant adding only the path constant flips the observation while the behaviors stay missing (verification-completeness §2 wrong-reason direction). Repair: record behavior-discriminating probes at the pin SHA (e.g. `grep -rn "Retry-After" internal/jev` — 0 hits today, verified — and a semaphore-symbol grep), or re-cite those criteria to a package-absence row with the limitation stated. Also relabel R1/R6: `ls` writes its error to stderr, recorded under "Verbatim stdout" (cosmetic; elements complete).
- **D5 [SHOULD-FIX, class design]** plan.md:30 B1 vs REQ-JEVC-007 — REQ-JEVC-007 ("every consumer of an unavailable result shall … exit 0 … continue") is textually unreconciled with the off-gate exit 2. B1's design is sound for a one-shot command, but the artifacts must carry the carve-out sentence: REQ-JEVC-007 binds in-process consumers that continue a workflow; a dedicated read-only command reports its own unavailability through its exit-code contract (exit 2, the `fetch_error` class). If the lane instead reads REQ-JEVC-007 as binding CLI commands, that is an operator call — state the reading in the decision record.
- **D6 [SHOULD-FIX, class design]** plan.md:34 B5 / REQ-FW-011+013 — the cache key uses the interpretation model id, but that id is learned only from the **response** (`Request` has no model field); how a lookup is keyed before any response, how alias movement is detected, and what a restart-then-offline hit does are unspecified. Repair: pin the mechanism in B5/M5 (e.g. persist a last-observed alias→model map under the cache dir; lookup keyed by body-SHA + that id; response model recorded on every network answer; staleness warning when the two diverge) and extend AC-FW-009/011 method with restart-offline-hit and alias-moved cases.
- **D7 [SHOULD-FIX, class design]** plan.md §A.2 / REQ-FW-016 / AC-FW-014 — (a) AC-FW-014 checks test network isolation, not the t1762 boundary; the boundary's only verification point is DoD §I.4 (process check). (b) §A.2's premise "both cards touch `internal/jev/jev.go`" is stale: t1762 (`954adab25e`) builds the recorder in a **new** `internal/jevledger` package and explicitly does not edit `internal/jev`. Repair: refresh §A.2 (conflict premise removed — good news), name `internal/jevledger` (and t1762's config/template surfaces) as do-not-touch in REQ-FW-016, and make §I.4's diff check name the actual guarded paths with a positive control.
- **D8 [MINOR, class ac-wording]** REQ-FW-002's "adds no tokenizer dependency" / "documented at use site as approximate" limbs are mapped to AC-FW-001, which does not exercise them (mutant: import a tokenizer library; all 14 ACs still pass). Fold a one-line assertion into AC-FW-014's method (no new go.mod dependency / no tokenizer import in `internal/fetchweb`) at M3 — no new AC needed.
- **D9 [MINOR, class format]** plan.md milestones are bold bullets, so the CN-4 verb collected 0 milestones mechanically (GAP; hand-read instead). Make M1..M7 headings in the revision so the next audit observes the order mechanically.

## Regression Check

Iteration 1 — not applicable.

## B1 / B5 / B6 dispositions (lane asked)

- **B1 (off-gate exit code)** — lane-resolvable at kickoff; keep exit 2, and write D5's reconciliation sentence into the decision record (this is the fix D5 requires). Escalate to leader/operator only if the lane adopts the reading that REQ-JEVC-007 binds CLI commands.
- **B5 (stale cache)** — lane-resolvable; keep-but-warn is right (matches the analysis; never deletes user data), but it is not implementable until D6's mechanism is pinned — pin it in B5/M5 first, then resolve.
- **B6 (bench command)** — lane-resolvable; as written (loader + baseline.json compare, containment + token-savings, no upstream numbers) matches the analysis and the Gap discipline. Only name the concrete subcommand in the M6 delegation prompt.

None of the three needs an operator decision; the operator-fixed items (naming, default-off, no-Python, t1965 deferral) are directives and stay closed.

## Recommendation (for manager-spec, one revision)

1. Fix D1 (no_match decision order), D2 (enabled-true scope), D3 (retry × Repeatable) — the blocking set; fold D4-D7 into the same revision while editing (D6 is prerequisite to resolving B5).
2. Re-commit the plan artifacts, then re-audit: delta scope = D1-D7 anchors plus the mandatory full CN-4 verb run; D8/D9 may ride along or go to run-phase constraints.
3. Design fidelity otherwise verified complete — no scope changes are being requested; every fix above is wording/mechanism pinning inside the already-approved boundary.
