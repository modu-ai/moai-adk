# SPEC Review Report: SPEC-CODEX-COVER-RESIDUAL-001

Card: t519 · Iteration: 2/3 · Tier M (PASS threshold 0.80)
Tree: `1e21635f1` on `WT-codex-cover-residual` (worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t519`); iteration 1 audited `faa76fa7e`
Auditor: plan-auditor (Claude anchor only — `audit_model` absent from `.moai/config/**`; no cross-model fan-out)

**Reasoning context ignored per M1 Context Isolation.** Every iteration-1 finding was re-verified against the tree and the production source, not against manager-spec's fix report. The fix report was read only to locate the edits; no claim in it was accepted as evidence.

**Verdict: PASS — Overall Score 0.9375** (threshold 0.80; blocking 2, advisory 1)

Score movement: 0.775 → 0.9375. No regression, so no STOP escalation and no scope-reduction proposal. All seven must-pass criteria PASS. All four iteration-1 blocking findings are resolved on the tree. Two new minor findings remain, one of which the fix round introduced and one of which iteration 1 under-scoped; both are single-token edits and neither gates run-phase entry.

---

## Claim

Every iteration-1 blocking defect is repaired in substance, verified against the source rather than the report: AC-CCR-004 is now hermetic, AC-CCR-003's mutant M2 can fire, AC-CCR-002's adoption basis M1b flips both of its payloads (verified by execution, not reasoning), and all twelve seam locations are correct. Two residual inconsistencies remain — a stale mutant pointer on one criterion and a file-count that contradicts its own list — neither of which changes what a correct implementation does.

## Evidence

### Regression check — iteration-1 findings

| Finding | Status | Evidence on tree `1e21635f1` |
|---|---|---|
| F1 — AC-CCR-004 omits the `codexLookPath` swap | **RESOLVED** | acceptance.md:76 now requires "**all three** codex seams swapped together under one `t.Cleanup`" and names `codexLookPath = func(string) (string, error) { return "/fake/codex", nil }` explicitly. plan.md §F M1 step 5 carries the block as a code fence. I diffed that fence against the cited precedent `TestReviewGate_FailOpenOnCodexError` (codex_review_gate_test.go:154-158) line by line: byte-identical apart from leading indentation, so the "copied verbatim" claim holds. acceptance.md:80 adds the host-dependence rationale, correctly citing `var codexLookPath = exec.LookPath` (mcp_codex.go:368) and the step-4 consult site (codex_review_gate.go:78, fail-open return at :80). |
| F2 — AC-CCR-003 omits `withChangeDetector`, M2 vacuous | **RESOLVED** | acceptance.md:64 now carries `withChangeDetector(t, true)` in the Given, and :68 states why it must not be dropped as redundant. Both cited line ranges check out: codex_review_gate.go:69 is the step-1 `return allow, nil // (1) opt-in default-off`, and :129-132 is the `git status` fail-open returning false. The ledger row M2 (acceptance.md:189) now carries the same warning inline. Under the fixed fixture, mutant M2 (`enabled := true`) advances past steps 1-3 and reaches `codexLookPath` at :78, firing the `t.Fatal` guard — M2 is detectable. |
| F3 — AC-CCR-002 adopted on vacuous M1 | **RESOLVED as specified; one adjacent surface left stale → F5** | The two edits my iteration-1 fix instruction named were both made: the §B matrix cell (acceptance.md:26) now reads `mutant M1b`, and a new ledger row M1b (acceptance.md:187) replaces codex_review_gate.go:188 with `return err`. The M1 row was also narrowed to line 186 and now states it "Does **not** fire AC-CCR-002 … hence M1b." A third surface I did not name — the §C bullet — was not updated; see F5. |
| F4 — plan.md §A.4 mis-locates four seams | **RESOLVED** | plan.md §A.4 is rewritten as a 12-row table. I re-grepped every row in this run: `withChangeDetector` codex_review_gate_test.go:28; `writeWorkflowYAML` :33 and `assertAllowJSON` :157 in multi_review_gate_wiring_test.go; `withCodexRunner` :56, `withCodexLookPath` :63, `fakeCodexSession` :74, `fakeCodexConn` :88, `withCodexSession` :108, `codexSessionScript` :123, `errFakeCodexCrash` :408 in mcp_codex_test.go; `fakeCodexConnPID` codex_jobs_test.go:31; `stubCodexRunner` codex_rpc_error_test.go:29. **12 of 12 match.** spec.md §E constraint 3 now names `codex_rpc_error_test.go`, the file neither artifact previously carried. |

Advisories: A1 honoured (threshold untouched — the correct response to a recommendation to change nothing); A2 applied as plan.md §D constraint 10; A3 applied as the third bullet of AC-CCR-010; A4 and A5 declined with reasons recorded in progress.md. **A4's decline is correct — see § A4 materiality ruling.**

### Probe 1 — seam interaction none of the sibling tests exercise

The lead asked specifically about `withCodexSession` and a manual `codexLookPath` swap coexisting in one test, and how their `t.Cleanup` restores interact.

**Restoration is safe by construction, in either order.** Both helpers snapshot the previous value at call time (`withCodexLookPath` at mcp_codex_test.go:65, `withCodexSession` at :110) and `t.Cleanup` runs last-added-first-called, so a nested pair unwinds correctly: the inner helper restores the outer helper's function, then the outer restores the production default. No leak across tests in either ordering.

**The hazard is not restoration — it is overwrite during the test body, and it is real.** `withCodexSession` swaps three seams, not one: mcp_codex_test.go:111-112 sets `codexRunner = stubCodexRunner{}` and `codexLookPath = func(string) (string, error) { return "/fake/codex", nil }`. A test that installs a `t.Fatal` lookPath guard via `withCodexLookPath` and then calls `withCodexSession` has its guard silently replaced by the stub, and the guard can never fire again.

**No planned test hits this.** I checked all six: AC-CCR-001/002 use neither; AC-CCR-003 uses `withCodexLookPath` alone; AC-CCR-004 uses the manual three-seam block and never calls `withCodexSession`; AC-CCR-005 uses `withCodexSession` alone; AC-CCR-006 uses neither. No sibling test in the package combines them either, which is why the trap is unexercised and undocumented. Recorded as advisory A6 because it is the regression path for the finding this round just fixed.

### Probe 2 — M1b compiles and flips both payloads

**Scope.** `err` is declared at codex_review_gate.go:184 (`input, err := readHookInput(cmd.InOrStdin())`) in the function block, so it is in scope inside the `if err != nil` block at :185-189. Line 188 currently reads `return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})`; replacing it with `return err` returns an `error` from a function whose signature is `error`. `input` remains used at :190, so no unused-variable error. The edit compiles.

**Both payloads reach the mutated line.** This is the linchpin of M1b — if the empty payload did not error, it would skip the whole block and M1b would be vacuous for AC-CCR-002 exactly as M1 was. I executed a standalone program mirroring `readHookInput` (codex_review_gate.go:207-217) verbatim rather than reasoning from stdlib semantics:

```
"{not json" -> invalid character 'n' looking for beginning of object key string
""          -> unexpected end of JSON input
"{\"a\":\"x\"}" -> <nil>
```

Both the malformed and the empty payload produce a non-nil error, so both take the `if err != nil` branch and both flip under M1b. AC-CCR-001 and AC-CCR-002 each assert `Execute()` returns nil; both fail. **M1b is a valid adoption basis for AC-CCR-002.**

### A4 materiality ruling (lead's question)

**Not material — the decline stands, and A4 is closed rather than carried forward.** spec.md still has no heading named WHAT / Scope / Overview, but scope is carried more precisely than a heading would carry it: §B.2 is a per-function table with an explicit `Disposition` column marking each of ten functions in-scope, covered, or out-of-scope, and §F enumerates four exclusion classes under `### Out of Scope — <topic>` H3 sub-headings with bullets. A reader determines exactly what is being built and what is not. Adding a heading would satisfy a checklist without adding information, and manager-spec's stated reason for declining — that the edit falls outside the acceptance/plan surface this round authorised — is sound scope discipline. Requiring it would be the over-engineering the M6 finding-consumption brake exists to prevent.

### Must-Pass Results (re-verified on `1e21635f1`)

- **[PASS] MP-1 REQ number consistency.** `grep -o 'REQ-CCR-[0-9]\{3\}' spec.md | sort -u` → `REQ-CCR-001` … `REQ-CCR-010`, ten ids, sequential, no gap, no duplicate.
- **[PASS] MP-2 GEARS compliance (requirement layer).** The requirements are unchanged from iteration 1 apart from one cross-reference repair at spec.md:92 (`§D.1` → `§D.2`); every `REQ-CCR-XXX` retains its GEARS modality. Judgment made against the requirement layer only; the twelve `Given/When/Then` `AC-XXX` entries are verification-layer and graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity.** All 12 canonical fields present and correctly typed; `version` bumped to `"0.2.0"` (valid quoted semver), `status: draft`, `created`/`updated` ISO, `phase: "v3.1.4 target"` a release target rather than a prohibited lifecycle token, `module`, `lifecycle`, `tags` intact; optional `era`/`tier`/`related_specs` permitted. No rejected snake_case alias. Independent confirmation: `moai spec lint .../spec.md` → `✓ No findings — all SPEC documents are valid`.
- **[N/A] MP-4 language neutrality.** Single-language SPEC (Go, `internal/cli`).
- **[PASS] MP-5 D7 cross-SPEC reconciliation.** Re-extracted this run: `SPEC-CODEX-TEST-GAPS-001` → `status: completed`; `SPEC-MOAI-MCP-SERVER-001` → `status: completed`. Neither retired, superseded, nor archived; no reconciliation clause required.
- **[PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall'` → spec.md 0, acceptance.md 0, plan.md 1 — the single hit is plan.md:74's negation inside the Section-B filter ("no syscall use"). No syscall introduced.
- **[PASS] MP-7 clarification gate.** `grep -c '\[NEEDS CLARIFICATION:'` → 0 across all four artifacts. `research.md` absent, correct for Tier M.

### Structural re-verification

- Counts unchanged and within Tier M ceilings: 10 requirements, 12 acceptance criteria (ceiling 16 each, applied independently).
- Mutant ledger grew 9 → 10 rows with M1b, matching progress.md's own claim.
- Fix-commit scope: `git show --stat 1e21635f1` touches five files — the four SPEC artifacts plus the iteration-1 report. No `.go` file.
- AC-CCR-007's union gate remains green for the stated reason: `git diff --name-only bf779ecf2..HEAD` lists seven paths, all `.md` / `.log` / `.txt`; `git status --short` is empty. Zero `.go` paths in either output.

### Defects Found

D1. **F5 — AC-CCR-002's §C bullet still names the mutant the matrix replaced.** — `acceptance.md:60` — Severity: minor — Class: **blocking** — The §B adoption matrix (acceptance.md:26) now reads `mutant M1b`, but the §C bullet under AC-CCR-002 still reads `- Adopted via mutant **M1** (§D).` The same criterion therefore names two different adoption bases on two surfaces — the shape F3 was raised about, relocated rather than eliminated. Harm is bounded, which is why this is minor rather than major: the M1 ledger row now says outright that M1 "Does **not** fire AC-CCR-002 … hence M1b", so a reader who follows the stale bullet is redirected within one hop, and plan.md §F M1 step 7 — the surface an implementer actually works the mutants from — correctly says "M1 fires AC-CCR-001 only; M1b is AC-CCR-002's adoption basis." **This is not a failure to apply the iteration-1 fix**: my iteration-1 instruction named exactly two edits (the matrix cell and the new ledger row) and both were made. The third surface was my own under-scoping. — **Required fix**: change `M1` to `M1b` at acceptance.md:60.

D2. **F6 — the seam table's file count contradicts the table itself.** — `plan.md:37` and `spec.md:147` — Severity: minor — Class: **blocking** — Both rewritten passages assert the helpers live in "**four** files". Both then list five. plan.md §A.4's twelve rows span `codex_review_gate_test.go`, `multi_review_gate_wiring_test.go`, `mcp_codex_test.go`, `codex_jobs_test.go`, and `codex_rpc_error_test.go`; spec.md constraint 3 names those same five inside the very sentence that claims four. The count is wrong, not the table — every location is verified correct — but the sentence draws attention to itself: plan.md:37 says "the count matters, because a helper sought in the wrong file invites a duplicate definition." Introduced by the F4 fix; the pre-fix text carried no count. — **Required fix**: change "four" to "five" at plan.md:37 and spec.md:147.

### Advisory finding

A6. **`withCodexSession` silently overwrites a `t.Fatal` lookPath guard.** — mcp_codex_test.go:111-112 — No planned test combines the two, so nothing is wrong today. The trap is that adding `withCodexSession` to AC-CCR-003 "for symmetry with the other gate tests" would replace the guard with the `/fake/codex` stub, making the guard unfirable and returning M2 to vacuous — silently reintroducing F2. One sentence in AC-CCR-003's existing rationale paragraph would close it permanently: the lookPath guard and `withCodexSession` are mutually exclusive in one test. Left to the orchestrator's discretion per M6; it is a hypothetical the SPEC does not currently claim to cover.

## Baseline-attribution

Every measurement was taken in this run, against tree `1e21635f1` in this worktree.

- Tree: `git rev-parse --short HEAD` → `1e21635f1`; `git branch --show-current` → `WT-codex-cover-residual`. Iteration 1's tree was `faa76fa7e`, the parent commit.
- All twelve seam locations were re-grepped in this run rather than carried over from iteration 1, per the attribution requirement — even though no `.go` file changed between the two trees, an unremeasured figure is not a baseline.
- The `readHookInput` behaviour is my own execution in this run (`go run` over a standalone mirror of codex_review_gate.go:207-217), not a citation of stdlib documentation.
- The coverage figures (80.9% package, 0.0% `runCodexReviewGate`, 66.7% `pid`) remain **the lane's measurement on tree `bf779ecf2`**, unchanged and still correctly attributed as such in spec.md §B.1. I did not re-execute the ~474s coverage run.
- The verbatim-copy claim for plan.md's code fence was checked by reading codex_review_gate_test.go:154-158 in this run and comparing line by line.

## Gaps

Explicitly NOT observed.

- **No test was executed against `internal/cli/`.** Every statement about what a planned test will do under a mutant remains a control-flow prediction. The one exception is probe 2, where I executed a mirror of `readHookInput` rather than reasoning — that specific claim is observed. The other nine mutants were reasoned about, not applied.
- **Coverage was not re-measured**, so the 12-of-13 / 92.3% ceiling remains a derivation. Iteration 1 recounted it independently and it reproduced; nothing in this round revisited it.
- **`go vet`, `golangci-lint`, `gofmt` were not run.** AC-CCR-011 is a run-phase gate.
- **The delta scope was honoured, so this was not a from-scratch re-audit.** I re-verified the four iteration-1 findings, ran the two probes the lead named, re-ran all seven must-pass checks, and swept the text the fix round introduced. I did **not** re-derive the parts of the SPEC iteration 1 already cleared (the S1 skip justification, the AC-CCR-005 BLOCK reachability chain, the AC-CCR-009 selector shape, traceability). A defect introduced into one of those by the fix commit would have been caught only if it fell inside the diff, which I did read in full.
- **No cross-model second opinion.** Claude-only verdict.

## Residual-risk

- **F6 is a defect the fix round created, which is the pattern worth watching.** Both new findings live in text the previous round wrote. A third iteration that edits prose again carries the same risk, which is an argument for making these two single-token corrections and stopping rather than re-opening the artifacts more broadly.
- **My iteration-1 fix instructions were surface-enumerations, and F5 shows that enumeration can be incomplete.** I named two surfaces for F3; there were three. The two corrections above are again stated as specific line edits, so the same incompleteness is possible. Before applying them, a `grep -n 'M1\b' acceptance.md` and a `grep -rn 'four' plan.md spec.md` would confirm no fourth surface carries the stale value — cheaper than another audit round.
- **The adoption bases are now sound but still unobserved.** All three repaired mutants are predicted to fire from control-flow reading; only run phase converts that to evidence. REQ-CCR-008 and AC-CCR-010 already require recording a non-firing mutant as a finding rather than dropping it, which is the right guard, and A6 names the one edit that would silently undo it.
- **PASS here is a plan-quality verdict, not an implementation guarantee.** It says the SPEC is fit to implement against, with adoption bases that can fire and fixtures that are hermetic. Whether the tests actually reach 12 of 13 statements is the run-phase re-measurement's question (REQ-CCR-009).

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | between the 0.75 and 1.0 anchors | The two under-specified Given clauses that drove iteration 1's deduction are now fully specified, each with a rationale paragraph carrying verified line citations (acceptance.md:68, :80). Short of 1.0 only because one criterion's adoption basis reads two ways depending on which surface the reader lands on (F5). |
| Completeness | 0.90 | between the 0.75 and 1.0 anchors | All required sections present; frontmatter complete and lint-clean; the seam table is now 12 rows, all verified. Deduction is F6's count contradicting its own list. A4 ruled immaterial and closed, so it carries no deduction. |
| Testability | 0.95 | 1.0 band, one notch withheld | Every AC is binary, weasel-free, and mechanically checkable, and all three defective adoption bases are repaired: AC-CCR-004 hermetic (F1), AC-CCR-003's M2 detectable (F2), AC-CCR-002's M1b verified by execution to flip both payloads (F3, probe 2). A2 and A3 applied. Withheld from 1.0 because the stale §C bullet could still route an implementer to the wrong mutant for AC-CCR-002. Iteration 1 scored 0.60. |
| Traceability | 1.00 | 1.0 — complete bidirectional coverage | Unchanged and re-verified: every REQ has at least one AC and every AC names a valid REQ, per the matrix at acceptance.md:23-36 and the coverage line at :38. REQ-CCR-006 is explicitly routed to the §F Definition of Done. No orphaned AC, no uncovered REQ. F5 is an adoption-basis pointer, not a traceability edge. |

**Aggregate**: (0.90 + 0.90 + 0.95 + 1.00) / 4 = **0.9375**. Tier M threshold 0.80 → PASS.

**Score-regression check**: iter1 0.775 → iter2 0.9375. Improvement, so the STOP escalation clause does not fire and no scope-reduction proposal is made.

## Recommendation

PASS. The SPEC is fit to implement against.

Two single-token corrections are recommended before Implementation Kickoff Approval. Neither gates the verdict, and neither changes what a correct implementation does:

1. **acceptance.md:60** — `M1` → `M1b`, so AC-CCR-002's §C bullet agrees with its §B matrix cell.
2. **plan.md:37 and spec.md:147** — `four` → `five`, so the file count agrees with the five files each passage lists.

Advisory A6 (the `withCodexSession` / lookPath-guard exclusivity) is surfaced for the orchestrator's discretion; one sentence in AC-CCR-003's existing rationale would close the regression path for F2 permanently.

Per `.claude/rules/moai/development/verification-completeness.md` §7, this audit rests on that rule's §1.1 (observed-failure completion and the empty-sweep hazard), §2 (two-cell adoption and the mutant probe), and §2.1 (RED-now cell content, the undecidable disposition), and on `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields and § Artifact Statelessness. The M6 finding-consumption brake was applied in ruling A4 immaterial and in holding F5 and F6 at minor rather than inflating them to justify a second FAIL.

VERDICT: PASS score=0.9375 blocking=2 advisory=1
