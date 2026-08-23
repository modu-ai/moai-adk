# SPEC Review Report: SPEC-MCP-WORKTREE-ROOT-001
Iteration: 3/3 (delta re-audit — scoped to iter-2's enumerated defect list [R1] + bundled self-sweep items + regression sweep, per the Retry Loop Contract delta rule; Tier S ceiling S=1 exhausted at iter-1, this pass runs on the orchestrator's explicit delta dispatch, within the LEAN 3-iteration cap)
Verdict: **PASS**
Overall Score: 0.9375 (aggregate ≥ 0.80 dispatch threshold met; also ≥ Tier S per-tier threshold 0.75)

Auditor: plan-auditor (independent, card t171). Reasoning context (author's claims about
what was fixed) ignored as verdict evidence per M1 Context Isolation; every load-bearing
assertion re-measured against the worktree source at the pinned commit. Cross-backend
audit tools NOT invoked (inadmissible here by dispatch instruction — their project-root
resolution is the defect under audit).

## Fixed-State Confirmation

```
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t171 log --oneline -3
d5b09ceb8 fix(SPEC-MCP-WORKTREE-ROOT-001): apply iter-2 R1 and the two queued self-sweep items
4640fefe2 fix(SPEC-MCP-WORKTREE-ROOT-001): apply plan-audit iter-1 findings D1-D5
576f9b893 feat(SPEC-MCP-WORKTREE-ROOT-001): plan-phase artifacts (Tier S, 2 artifacts)
$ git status --porcelain  →  ?? .moai/reports/t171/plan-audit-iter1.md, plan-audit-iter2.md (prior reports only)
```

HEAD `d5b09ceb8`, parent `4640fefe2` — matches the pinned state. Tracked tree clean.
SPEC v0.3.0, `version` bumped 0.2.0→0.3.0 and `updated` carried (same-day, correct).
Artifact set confirmed: exactly `spec.md` + `plan.md` (Tier S; `research.md` absent → MP-7
N/A component).

**Process note (measurement integrity):** one verification batch initially ran on
relative paths and resolved against a different checkout (pwd was another worktree; the
SPEC directory does not exist outside this branch). Caught on the spot; every check in
this report was re-executed against absolute t171 paths. No number below derives from
the stray-checkout run.

## Code Re-Verification (every NEW citation in this edit, against t171 source)

| SPEC/plan claim | Re-measured at | Result |
|---|---|---|
| `performCodexAudit`'s params map (`mcp_convergence.go:387-394`) carries **no `cwd` key at all** (spec.md:161-162, §4 :255-256, plan §A :23, §F :116-119) | `mcp_convergence.go:387-394` — `params := map[string]any{"target", "model": "", "prompt": …}` + conditional `"focus"`; keys = target/model/prompt/focus only, **no `cwd`**. Range 387-394 exact | **VERIFIED** |
| Single-backend path (`mcp_codex.go:1167-1171`) has `"cwd": resolveProjectDir()` (spec.md:159-160, §4 :251-252) | `mcp_codex.go:1167-1171` — `params := map[string]any{"target", "model", "cwd": resolveProjectDir()}`; the cwd line is `:1170`. Range exact | **VERIFIED** |
| R1's structural premise re-confirmed in the same read | `backendCallFn` signature at `:353` carries no root; `var backendCall` at `:357` (swappable double, comment :355-356); `defaultBackendCaller` `:362` routes to `performCodexAudit` `:365` — a swapped double replaces, not reaches, the params builder | **VERIFIED** |
| spec.md:131 "the CLI-side list is longer than the three files named — those three hold five call sites between them" | `grep -n resolveProjectDir` on the three named files: `goal.go:159`, `goal.go:188`, `memory.go:165`, `memory.go:287`, `launcher_blockcap_infinite.go:139` = **exactly 5 call sites in exactly those 3 files, at exactly the lines spec.md:117-119 cites** | **VERIFIED** |

The author's claim that the underlying fact is WORSE than R1's finding is therefore
code-true: the fan-out path passes no cwd at all, so M2's repair adds a key rather than
redirecting one. This matches iter-2's own code-truth note; v0.3.0 records it as design
input (below).

## R1 Resolution Verdict (iter-2 blocking finding): **RESOLVED**

R1's required fix was: extend AC-1b's second half to pin the codex backend's review
params to `cwd`, mirror in plan M2's exit, bump version/updated. Applied — and exceeded:

1. **spec.md AC-1b (:153-176)** now asserts on "the `params` map of **both** codex paths,
   with no live backend", naming each with its file:line — the single-backend map
   (`mcp_codex.go:1167-1171`, cwd today `resolveProjectDir()`) AND the fan-out map
   (`performCodexAudit`, `mcp_convergence.go:387-394`, "carries **no `cwd` key at all**
   today"). Both characterizations verified against source above.
2. **The bypass scenario is recorded verbatim** (:164-171): "a `backendCall` double
   asserting what it received — is satisfied while `performCodexAudit` is never entered…
   The root could stop at the seam with all seven criteria green… Pinning the assertion
   to the params map closes that gap." The loophole R1 named (implementation widens the
   seam, stops there, all ACs green) is no longer reachable: the seam-bound test no
   longer satisfies AC-1b's letter, because the letter now names the params map.
3. **plan §F M2 exit (:110-124)** is pinned at the params map, not the seam: "the
   milestone is not done when the seam carries the root — it is done when the params map
   does… Ends with AC-1b, AC-2, and AC-3 met there, asserted on both codex params maps."
4. **plan §A (:23)** adds the `:387-394` surface row ("gains a `cwd` in its params map —
   it has none today, so this adds the key rather than redirecting it").
5. **plan §G (:154-156)** records the anti-pattern ("Satisfying AC-1b with a
   `backendCall` double alone… Assert the map."), completing the mirror in all four
   artifact surfaces the finding touched.
6. `version`/`updated` bumped; HISTORY 0.3.0 entry (:284-292) describes the change
   accurately, including the divergence record.

The fix goes beyond the one-sentence minimum by converting the discovered fact (paths
have diverged) into build-shape guidance for M2 — closing the gap rather than merely
wording over it.

## Bundled Self-Sweep Resolution Verdicts: **RESOLVED**

- **§4 Evidence list** (:243-256): now carries the full citation set — the measurement
  doc, `session.go:242`, `audit.go:157`, the three SPEC-tool handlers (`:526/:583/:597`),
  `mcp_server.go:367` (audit_multi registration), `mcp_codex.go:1167-1171`,
  `mcp_convergence.go:353/:485`, and `mcp_convergence.go:387-394`. The stale pre-D1
  four-tool list is gone; every D1/D2/R1-era surface is cited, consistent with plan §A.
- **spec.md:131 wording**: tightened exactly as claimed and verified against code
  (5 call sites / 3 files, above).
- **§4.1 Tier note** (:258-271): "two Go files" → "**three** Go files" (mcp_server,
  mcp_codex, mcp_convergence — matching plan §A's Go-file set now that mcp_convergence
  entered core scope). File-axis judgment unchanged: the 8-10-path overrun is still
  disclosed, Tier S still claimed on the LOC + REQ/AC axes (6 REQ / 7 AC, both within
  the Tier S ceiling of 8), and the Tier-M flip condition (REQ-4 verdicts becoming
  in-card repairs) is preserved verbatim.

## Regression Sweep (numeric vocabulary + structure)

Zero regressions found. High-risk numerics re-counted against their referents:

| Claim | Location | Observed | Verdict |
|---|---|---|---|
| "five in-scope tools" / "Five MCP handlers" | spec.md:143, plan.md:10 vs §2 table :74-80 | 5 table rows (spec_audit, spec_drift, spec_progress, codex, audit_multi) | MATCH |
| "the two codex paths have diverged" | spec.md:175, plan.md:119-120 | code-verified: single-backend has `cwd`, fan-out has none | MATCH |
| "three Go files" | spec.md:261 vs plan §A | mcp_server.go, mcp_codex.go, mcp_convergence.go = 3 | MATCH |
| "Six requirements, seven criteria — REQ-1 carries two" | spec.md:138-139 | REQ-1..6 sequential (:141/:184/:193/:206/:219/:231); AC-1a,1b,2,3,4,5,6 (:147/:154/:189/:199/:213/:226/:237) = 7; REQ-1→AC-1a+1b | MATCH |
| "four MCP tools and the convergence state directory" (undecided) | spec.md:113-116, plan.md:146-147 | goal_arm, goal_status, verify_snapshot, verify_trend + state dir = 4+1 | MATCH |
| "those three hold five call sites between them" | spec.md:131 | grep-verified 5 sites / 3 files at the cited lines | MATCH |
| "all seven criteria green" (AC-1b rationale) | spec.md:169 | 7 ACs | MATCH |

AC numbering 1a/1b/2-6 continuous; no orphan ACs, no uncovered REQs (full bijection:
REQ-1→1a/1b, REQ-2→2 … REQ-6→6). No stale four-tool scope references. No contradiction
introduced between spec §2/§3/§4 and plan §A/§F — the `mcp_codex.go:1170` (table row) vs
`:1167-1171` (AC-1b range) citations are consistent (line ⊂ range). HISTORY 0.3.0 entry
accurate.

## Must-Pass Results (7/7 held — no breakage from this edit)

| # | Criterion | Result | Evidence (this iteration's observation) |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | REQ-1…REQ-6 sequential, no gaps/dups (headings enumerated; spec.md:141-231) |
| MP-2 | EARS/GEARS compliance (requirement layer only) | **PASS** | All six REQs SHALL-bearing with subjects/conditions (REQ-1 ubiquitous+event compound; REQ-6 event-driven). Judgment made entirely against the REQ layer in §3; inline Given-When-Then ACs are verification-layer format, not penalized |
| MP-3 | YAML frontmatter validity | **PASS** | spec.md:1-16 — all 12 canonical fields, correct types, `version: "0.3.0"` quoted semver (bumped per R1), `phase: "v3.1.3"` release target, zero snake_case aliases; optional `era: V3R6`/`tier: S` extras |
| MP-4 | Language neutrality | **N/A** | Single-language Go project (`module: internal/cli`) |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | `grep -hEo 'SPEC-…'` over both t171 artifacts → sole match `SPEC-MCP-WORKTREE-ROOT-001` (own ID); no external refs, no BLOCKING finding |
| MP-6 | D8 cross-platform discipline | **PASS** | `grep -c syscall` → 0 in spec.md, 0 in plan.md (observed against t171 paths) — auto-pass |
| MP-7 | Clarification gate | **PASS** | `grep -c 'NEEDS CLARIFICATION'` → 0 in both artifacts; `research.md` does not exist (Tier S, artifact set = spec.md + plan.md, confirmed by ls) → N/A component per MP-7's own rule |

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | minor ambiguity in one spot, consistently resolvable | Every REQ is now single-interpretation; the R1 seam-boundary reading is closed (AC-1b names both maps with file:line). One minor looseness survives: spec.md:79 row-4 slash naming ("`codex_task` / codex audit `cwd`") makes the "five tools" count require resolving via the :1170 citation (R3, optional, declined this round) |
| Completeness | 1.0 | all sections present and substantive | §4 evidence list complete including both R1-era surfaces; §4.1 tier note corrected to three files with flip condition intact; HISTORY 0.3.0 accurate; §2.1 exclusions unchanged and non-conflicting |
| Testability | 1.0 | every AC binary-testable | AC-1b now asserts at the correct boundary — `params["cwd"] == named worktree` on both maps, no live backend, no judgment calls; AC-2/3/4/5/6 unchanged and binary (iter-2-verified, intact at v0.3.0) |
| Traceability | 1.0 | full bijection | REQ-1→AC-1a/1b, REQ-2→AC-2 … REQ-6→AC-6; no orphan ACs, no uncovered REQs |

Aggregate: (0.75 + 1.0 + 1.0 + 1.0) / 4 = **0.9375**. Score progression 0.81 → 0.875 →
0.9375 — monotonically improving, no regression, no STOP signal.

## Defects Found

No blocking defects. Two previously-catalogued optional findings remain open (both
pre-existing at iter-2, both declined by the author this round — within the author's
discretion per the M6 routing rule; neither affects correctness or any criterion this
SPEC states):

R2 (optional, open) — spec.md:89-95 — cost disclosure still omits two mechanically
entailed touches (the `performCodexAudit` signature itself; the two invariant comments
`mcp_convergence.go:348-352`/`:480-483` that must be reworded when the seam widens).
— Severity: minor — Class: optional — Fix if desired: one sentence in §2's cost note.

R3 (optional, open) — spec.md:79 — row-4 slash naming still mixes `codex_task` with the
codex audit `cwd`; the operative surface is `codex_audit` (`handleCodexAudit`,
`mcp_codex.go:1170`). — Severity: minor — Class: optional — Fix if desired: name the
tool unambiguously or state `codex_task`'s disposition.

## Regression Check (Iteration 3)

Iter-2 defects:
- **R1 (blocking)**: **RESOLVED** — evidence in § R1 Resolution Verdict (code + all four artifact surfaces).
- **R2 (minor/optional)**: OPEN (author's discretion; not a regression).
- **R3 (minor/optional)**: OPEN (author's discretion; not a regression).

Iter-1 defects D1-D5: RESOLVED at iter-2; v0.3.0's edits did not disturb any of their
fix surfaces (scope table, AC split, consumer counts, illustrative-list markers, tier
note — all re-read intact, with the tier note's file count correctly updated from two to
three as part of the derived correction).

Structure-damage sweep: **0 regressions** (table in § Regression Sweep).

## Residual Risk

- AC-1b's "no live backend" assertability presumes a test seam at `runCodexReviewRPC`
  (a plain `func`, not a swappable var today). Both codex paths now route through it, so
  ONE var-ized seam covers both assertions. Package precedent exists
  (`projectDirResolver`, `backendCall`), making this a one-line M2 enabler — but it
  remains an unstated prerequisite carried from iter-2, now covering two maps instead of
  one.
- The invariant comment at `mcp_convergence.go:348-352` ("NOTHING ELSE") exists to keep
  `claude_verdict` out of the backends; widening for `project_root` must reword it
  without weakening that invariant (R2's second bullet). Run-phase reviewer should
  check the comment survives the widening.
- Iter-2's verified non-findings (REQ-3 observability, AC-4 settling-evidence
  enforcement, AC-6 recurrence semantics, spec_progress veto marker, exclusions
  coherence, §5 constraints) were spot-checked in the v0.3.0 text and remain intact.

## Recommendation

PASS. R1 is genuinely and verifiably resolved — the fix exceeds the required one-sentence
change by pinning both codex params maps in the AC, mirroring the params-map exit in
plan M2, adding the `:387-394` surface to plan §A, and recording the bypass as a §G
anti-pattern; the two bundled self-sweep items and the derived tier-note correction all
verify against source; the regression sweep is clean; all seven must-pass criteria hold.
Score 0.9375 ≥ 0.80 (and ≥ the Tier S threshold 0.75). The LEAN 3-iteration cap is now
reached with a PASS — no further iteration needed. Proceed to Implementation Kickoff
Approval (mandatory and score-independent; this PASS never auto-bypasses it). The two
open optional findings (R2/R3) may be folded into the run phase or dropped at the
orchestrator's discretion; neither blocks.
