# SPEC Review Report: SPEC-MCP-WORKTREE-ROOT-001
Iteration: 2 (delta re-audit — scoped to iter-1's enumerated defect list + regression sweep, per the Retry Loop Contract delta rule; Tier S ceiling S=1 was exhausted at iter-1, this iteration runs on the orchestrator's explicit delta dispatch)
Verdict: **FAIL**
Overall Score: 0.875 (aggregate ≥ 0.80 threshold met; FAIL is verdict-driven by ONE new blocking finding, R1 below — same M6 routing as iter-1)

Auditor: plan-auditor (independent, card t171). Reasoning context (author's claims about
what was fixed) ignored as verdict evidence per M1 Context Isolation; every load-bearing
assertion re-measured against the worktree source at the pinned commit. Cross-backend
audit tools NOT invoked (inadmissible here by dispatch instruction — their project-root
resolution is the defect under audit).

## Fixed-State Confirmation

```
$ git -C .claude/worktrees/t171 log --oneline -3
4640fefe2 fix(SPEC-MCP-WORKTREE-ROOT-001): apply plan-audit iter-1 findings D1-D5
576f9b893 feat(SPEC-MCP-WORKTREE-ROOT-001): plan-phase artifacts (Tier S, 2 artifacts)
edcbf593c docs(t171): confirm the cause before planning — one premise falsified, one re-attributed
$ git status --porcelain   → ?? .moai/reports/t171/plan-audit-iter1.md  (untracked report artifact only)
```

HEAD `4640fefe2`, parent `576f9b893` — matches the pinned state. Tracked tree clean; the
sole untracked file is iter-1's own report (local artifact). SPEC v0.2.0, `version` and
`updated` bumped correctly. Tier S artifact set confirmed: `spec.md` + `plan.md` only.

## Code Re-Verification (new load-bearing citations only; iter-1's nine verified sites unchanged and not re-listed)

| SPEC claim | Re-measured at | Result |
|---|---|---|
| `audit_multi` registration + input set has no `project_root` today (spec.md:80, plan.md:23 "~361-400") | `mcp_server.go:367-375` — `add(auditMultiToolName, …)` with inputs exactly `claude_verdict`/`target`/`focus`/`gates`/`session_id` | **VERIFIED** |
| Fan-out seam is `backendCallFn(ctx, backend, target, focus string)`, carries no root (spec.md:90-91) | `mcp_convergence.go:353` — signature verbatim as quoted; :355-357 comment confirms tests swap `backendCall` with `t.Cleanup` (the "injected test doubles" are real) | **VERIFIED** |
| Seam invoked at `:485` (spec.md:80, :91; plan.md:22, :112-113) | `mcp_convergence.go:485` — `out := backendCall(gctx, s.name, target, focus)` | **VERIFIED** |
| AC-1b assertion surface "the `params` map at `mcp_codex.go:1167-1171`, with no live backend" (spec.md:155) | `mcp_codex.go:1167-1171` — `params := map[string]any{"target", "model", "cwd": resolveProjectDir()}`; consumed by `runCodexReviewRPC` at :1181 | **VERIFIED** (range exact; see Residual-risk for the seam-var prerequisite) |
| audit_multi forwards the parameter "through to its backends" (spec.md:80, plan.md:112) | `mcp_convergence.go:362-397` — `defaultBackendCaller` → `performCodexAudit`, whose params map (:387-394) carries target/model/prompt/focus and **NO `cwd` key at all** | **PARTIAL** → new finding R1 |

Also re-confirmed the four undecided-consumer sites (`mcp_server.go:466/:483/:539/:569`,
`mcp_convergence.go:561`) and the full `resolveProjectDir` census in the MCP files
(`mcp_server.go:105/466/483/526/539/569/583/597`, `mcp_codex.go:1170` only,
`mcp_convergence.go:561`, `mcp_glm.go:98`) — unchanged from iter-1's verification.

**Code-truth note (correction to iter-1's own characterization, in the SPEC's favor):**
iter-1's report described the fan-out's codex `cwd` as "fixed at `resolveProjectDir()`".
That was imprecise: `performCodexAudit` (the fan-out path) passes NO cwd — codex runs in
the server's spawn-time cwd. v0.2.0 does NOT repeat that imprecision; its actual claim
("the seam … carries no root", spec.md:89-91) is exactly code-true. The blindness is real
either way; the SPEC's repair design (thread the root through the widened seam) is correct
against the real shape.

## D1 Resolution Verdict (iter-1 blocking finding 1): **RESOLVED**

- REQ-1 (spec.md:140-143) now creates the path REQ-6 needs: "The **five** in-scope tools
  SHALL accept an optional string input `project_root`", and the §2 table (spec.md:74-80)
  enumerates `audit_multi` as the fifth row with handler citations `mcp_server.go:367`,
  fan-out `mcp_convergence.go:485` — both verified above. REQ-6 (spec.md:213) is now
  executable without any change REQ-1 fails to authorize.
- Cost disclosure (spec.md:82-95) matches the real code shape: the seam signature, its
  single call site, and the injected doubles are each verified to exist; the disclosure
  honestly states it is "more than the SPEC-tool edits" and lands on the tier-note file
  axis (§4.1).
- Option-(b) escape hatch stated twice: spec.md:84-87 ("Re-aiming REQ-6 at `codex_audit`
  … would have been the cheaper edit") and :94-95 ("Stated here so the alternative can be
  taken instead if the cost is judged wrong"); mirrored in plan.md §E D1b (:70-77).

## D2 Resolution Verdict (iter-1 blocking finding 2): **RESOLVED**

- AC-1 is split into AC-1a (spec.md:146-151, SPEC tools, both directions, with an explicit
  "Both directions are asserted" rationale) and AC-1b (spec.md:153-162, the codex `cwd`
  parameter-present direction + the audit_multi fan-out direction, with the
  why-this-is-separate rationale).
- The cited assertion surface `mcp_codex.go:1167-1171` exists in source exactly as
  described (params map, `"cwd": resolveProjectDir()` at :1170).
- plan.md M2's exit is corrected: "Ends with **AC-1b, AC-2, and AC-3** met there — the
  parameter-present direction included, which is what iter-1 found missing" (plan.md:114-115),
  restoring consistency with plan §D "Both directions asserted on every new test" (:56).
  plan §G (:142-144) names the mirror mistake as an anti-pattern.
- Bidirectional completeness across surfaces: SPEC tools = AC-1a both directions; codex =
  AC-1b present + AC-2 absent + AC-3 invalid; audit_multi = AC-1b present + AC-2/AC-3.
  No unasserted direction remains on any in-scope surface — except the one hop deeper
  inside the audit_multi path, which is new finding R1 below.

## Regression Sweep (structure damage from the dishevel-then-retidy edits)

Zero leftover structure damage found. All count claims re-counted against their lists:

| Claim | Location | Actual | Verdict |
|---|---|---|---|
| "Six requirements, seven criteria — REQ-1 carries two" | spec.md:137-138 | REQ-1..6; AC-1a,1b,2,3,4,5,6 = 7; REQ-1→AC-1a+1b | MATCH |
| "five in-scope tools" | spec.md:142 vs table :74-80 | 5 rows (spec_audit, spec_drift, spec_progress, codex, audit_multi) | MATCH (naming nit R3) |
| "four MCP tools and the convergence state directory" (undecided) | spec.md:113-116, plan.md:136-137 | goal_arm, goal_status, verify_snapshot, verify_trend + state dir = 4+1 | MATCH |
| "Five MCP handlers resolve … through `resolveProjectDir()`" | plan.md:10 | the five in-scope surfaces (loose but consistent with the table) | MATCH |

REQ numbering sequential, no gaps/duplicates (spec.md:140/164/172/186/199/211); AC labels
continuous (146/153/169/179/193/206/217), each inline under its REQ; no orphan ACs, no
uncovered REQs. No stale four-tool-scope references — "four" appears only at spec.md:37
(the process measurement), spec.md:113 and plan.md:137 (the corrected consumer count).
`codex_audit`/`audit_multi` references are all role-consistent (audit_multi = the carrier;
codex_audit = the cheaper alternative not taken). Zero `[NEEDS CLARIFICATION` markers.
Zero `syscall` occurrences in either artifact (both files read in full). D7: only the
SPEC's own ID is referenced — no reconciliation obligation.

## Must-Pass Results (spot-confirm, all 7 held at iter-1)

| # | Criterion | Result | Evidence |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | REQ-1…REQ-6 sequential, no gaps/dups (spec.md:140-211) |
| MP-2 | EARS/GEARS compliance (requirement layer only) | **PASS** | All six REQs SHALL-bearing with subjects/conditions (REQ-1 compound ubiquitous+event, REQ-6 event-driven). Judgment made entirely against the REQ layer in §3; the Given-When-Then ACs are verification-layer format, not penalized |
| MP-3 | YAML frontmatter validity | **PASS** | spec.md:1-16 — 12 canonical fields, correct types, `version: "0.2.0"` quoted semver, `phase: "v3.1.3"` release target, zero snake_case aliases; optional `era`/`tier` extras |
| MP-4 | Language neutrality | **N/A** | Single-language Go project (`module: internal/cli`) |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | SPEC-ID extraction over both artifacts → only own ID; no external refs, no BLOCKING finding |
| MP-6 | D8 cross-platform discipline | **PASS** | `syscall` absent from both artifacts (full read) — auto-pass |
| MP-7 | Clarification gate | **PASS** | `grep -c "NEEDS CLARIFICATION"` → 0 in plan.md and spec.md; research.md does not exist (Tier S) → N/A per MP-7's own rule |

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity in one-two spots a reasonable engineer may resolve inconsistently | AC-1b's second half (spec.md:156-157) admits a seam-boundary reading (see R1); the "`codex_task` / codex audit `cwd`" row cell (spec.md:79) is loosely named (R3). All operative citations precise |
| Completeness | 1.0 | all sections present and substantive | D3 (:113-116), D4 (:128-133 illustrative + grep mandate), D5 (§4.1 :231-243 with Tier-M flip condition) all applied; HISTORY carries the 0.2.0 entry |
| Testability | 0.75 | one AC measurable at the wrong boundary | AC-1b's audit_multi half is binary-testable but its natural letter-satisfying test (a `backendCall` double) bypasses `performCodexAudit` — under-asserts the behavior it exists to guarantee (R1). All other ACs binary with concrete observables |
| Traceability | 1.0 | full bijection | REQ-1→AC-1a/1b, REQ-2→AC-2 … REQ-6→AC-6; no orphan ACs, no uncovered REQs (under-assertion scored under Testability per iter-1 precedent) |

Aggregate: (0.75 + 1.0 + 0.75 + 1.0) / 4 = **0.875**. No score regression vs iter-1
(0.81 → 0.875); no STOP signal.

## Defects Found

**R1** (new, blocking) — spec.md:156-157 (AC-1b second half) + plan.md:112-115 (M2) vs
`mcp_convergence.go:362-397` — The audit_multi forwarding is asserted only as far as the
fan-out seam: "then the root reaching the backend fan-out is that path". The natural
letter-satisfying test swaps the `backendCall` double (the mechanism :355-357 documents),
which **replaces** `defaultBackendCaller` and therefore never exercises
`performCodexAudit` — yet `performCodexAudit` is where the root must land as `"cwd"` in
the backend's review params, and that params map (:387-394) carries **no cwd key today**.
It is also a DIFFERENT map from the one AC-1b's first half pins (`mcp_codex.go:1167-1171`
belongs to `handleCodexAudit`, the codex_audit tool — not the fan-out path). Net: an
implementation that widens the seam and stops there satisfies AC-1a/1b/2-5, and AC-6 is
outcome-independent by design ("A recurrence does not fail this criterion" — spec.md:219),
so the card can close with every AC green while the motivating symptom (lane-9's
audit_multi reading the wrong tree) remains unfixed — contradicting the table row's own
stated claim "carries the parameter through to its backends" (spec.md:80). Same failure
family as iter-1's D2 (assertion gap in the last mile), one hop deeper on the path the D1
fix opened. — Severity: major — Class: **blocking** — Required fix (one sentence):
extend AC-1b's second half to "…then the root reaching the backend fan-out is that path,
**and the codex backend's review parameters carry it as `cwd`**" (assertable on
`performCodexAudit`'s params map by the same no-live-backend mechanism as the first
half); mirror the clause in plan M2's exit. Bump version/updated.

**R2** (minor, optional) — spec.md:89-95 — The cost disclosure ("one seam, one call
site, the doubles that implement it") omits two mechanically entailed touches: the
signature of `performCodexAudit` itself (the production implementation behind the seam,
called at `mcp_convergence.go:365` — not a test double) and the two invariant comments
(:348-352, :480-483) stating the signature carries "(ctx, backend, target, focus) and
NOTHING ELSE", which must be reworded when widened. Contained, same file, entailed by
the work — name them so the cost statement is complete. — Severity: minor — Class:
optional.

**R3** (minor, optional; pre-existing, not a regression) — spec.md:79 — Row 4
"`codex_task` / codex audit `cwd`": `handleCodexTask` resolves no project dir — the only
`resolveProjectDir`-derived cwd in `mcp_codex.go` is :1170 inside `handleCodexAudit`
(the codex_audit tool). The slash naming makes the "five in-scope tools" count require
interpretation (iter-1 read this row as one surface; its operative citations are
precise everywhere else). Name the tool unambiguously (`codex_audit`, handler
`handleCodexAudit`, `mcp_codex.go:1170`), or state explicitly whether `codex_task` is
in or out. — Severity: minor — Class: optional.

## Regression Check (Iteration 2)

Iter-1 defects:
- **D1 (blocking)**: **RESOLVED** — evidence in § D1 Resolution Verdict above.
- **D2 (blocking)**: **RESOLVED** — evidence in § D2 Resolution Verdict above.
- **D3 (minor)**: **RESOLVED** — spec.md:113-116 "four MCP tools and the convergence
  state directory"; plan.md:136-137 consistent.
- **D4 (minor)**: **RESOLVED** — spec.md:128-133 marks both consumer lists illustrative
  and names REQ-4's grep mandate as the exhaustive source.
- **D5 (minor)**: **RESOLVED** — spec.md §4.1 (:231-243) admits the 8-10-path file-axis
  overrun and states the explicit Tier-M flip condition.

Structure-damage sweep: **0 regressions found** (counts, numbering, orphan-reference,
marker, syscall, and SPEC-ID checks all clean — table in § Regression Sweep).

## Residual Risk

- AC-1b's "no live backend" assertability presumes a test seam at `runCodexReviewRPC`
  (a plain `func` at `mcp_codex.go:718`, not a swappable var today). The package has
  established precedent for var-izing seams for testability (`var projectDirResolver =
  resolveProjectDir`, mcp_glm.go:98; `var backendCall backendCallFn`, mcp_convergence.go:357),
  so the M2 implementer adds one line — but it is an unstated prerequisite.
- Iter-1's verified non-findings (REQ-3 observability, AC-4 settling-evidence enforcement,
  AC-6 recurrence semantics, spec_progress veto marker, exclusions-vs-requirements
  coherence, §5 constraints) were spot-checked in the v0.2.0 text and remain intact.

## Recommendation

FAIL on exactly one new blocking finding (R1), which is a one-sentence fix to spec.md
AC-1b plus a mirrored clause in plan.md M2. Both iter-1 blocking findings are genuinely
and verifiably resolved, the non-blocking D3-D5 were all applied, and the regression
sweep found zero structure damage — the SPEC's diagnostic core and its v0.2.0 honesty
(escape hatch, cost statement, illustrative-list markers) are all code-true as written.
Route: the author applies R1 (optionally R2/R3 in the same edit), bumps `version`/
`updated`, and the delta is re-checked against THIS report's defect list — a fresh
from-scratch re-audit is not required to clear one enumerated sentence-fix. The Tier S
iteration ceiling is exhausted (S=1); this and any further delta pass run on explicit
orchestrator dispatch, within the LEAN 3-iteration cap (this is iter-2).
