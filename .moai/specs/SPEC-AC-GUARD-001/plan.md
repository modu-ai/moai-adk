---
id: SPEC-AC-GUARD-001
title: "AC authoring convention and corpus disposition for worktree-guard-refused acceptance criteria"
version: "0.1.0"
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.1.0"
module: ".moai/specs,.claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: M
tags: "worktree-guard,acceptance-criteria,authoring-convention,census,measurement-first"
---

# Plan — SPEC-AC-GUARD-001

## §A Context

Card t1067. The "54/55 signatures" count is a proxy over a text-pattern predicate (`=$?` +
git verb), and the proxy is measured to diverge from the actual Claude Code worktree guard in
BOTH directions (false positives: the `$?`-capture skeleton P4/P6 executes fine; false
negatives: family A carries no `=$?` yet refuses — spec.md §2.4). Measurement and narrowing
therefore precede any prescription; the only repo-side actionable surface is the AC authoring
convention plus boundary documentation.

Development mode: documentation + SPEC-corpus change; no Go production code is expected to
change, so the DDD/TDD cycle degenerates to a census-and-docs workflow with acceptance.md
probes as the verification layer.

## §B Known Issues

- Proxy count drifts with develop (782/111/54 at t1059 → 794/112/55 at 0314801c2) — any figure cited later must be re-measured and tree-SHA-pinned (verification-completeness.md §4).
- The census predicate misses multi-line/adjacent-line family-A forms (spec.md §2.5) — measured this tree: 23 files carry multi-line open command substitutions (`VAR=$(…` unclosed at EOL; plan-audit-iter1) — the verified census must use block-level parsing, not line-level grep alone.
- The conjunct-selected 55 misses entire `$(git`-text populations OUTSIDE it — measured this tree: 59 corpus-wide files carry `$(git` text forms, only 18 inside the 55, 41 entirely outside (plan-audit-iter1 D2) — the census population cannot be scoped to the 55.
- The P8 exception boundary (pipeline-terminator non-git stage) is unmeasured beyond `wc` — labeled unknown, never treated as permission.
- The guard also refuses non-git-executing text that NAMES git in complex forms (spec.md §2.6) — census tooling itself must use single plain commands or it will be refused mid-census.

## §C Pre-flight

- [ ] Worktree-isolated session active (census probes and guard samples REQUIRE the worktree context to reproduce the refusal).
- [ ] Base tree pinned: record `git rev-parse --short HEAD` at census start and cite it in the disposition table.
- [ ] t1059 artifacts present as prior art (read-only): `git show 622e25d22:.moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md` lines 305-318.
- [ ] C1/C2 mirror pair confirmed for worktree-integration.md before M2 edits.

## §D Constraints

- Script-file relocation of AC commands is prohibited (spec.md §4).
- Guard refusal is upstream — do not attempt guard-side fixes.
- Template-First: C2 template mirror edit + `make build` for any `.claude/rules/moai/workflow/worktree-integration.md` change.
- rule-authoring.md statement duty fires if the always-loaded-surface growth exceeds 1,000 bytes in one edit — note worktree-integration.md carries a `paths:`-scoped loading scope per its frontmatter/loading note; verify before treating the growth as always-loaded cost.
- D-ITER2-09: no destructive cleanup tails (`rm -rf`) reintroduced into rewritten AC blocks; probes write only under /tmp.
- Census tooling commands themselves must be guard-executable (single plain invocations) — the census loop must not textually compose git in complex forms even when it executes no git (spec.md §2.6).

## §E Self-Verification

- E1: Disposition table carries, per file, the evidence command, verbatim output, exit code, and tree SHA; sweep file-list ⊆ table file-set with zero unresolved rows (AC-001/AC-002/AC-007).
- E2: Every rewritten block re-executed in a worktree-isolated session without refusal (AC-004).
- E3: C1/C2 parity check for the convention section; `make build` exit 0 (AC-003).
- E4: Grep over the SPEC diff for script-file relocation patterns (`\.sh` invocation introduced where a command previously ran inline) returns zero rows (AC-006).
- E5: Files dispositioned "leave, no rewrite" show zero content diff (AC-005).

## §F Milestones (priority order; measurement precedes prescription)

### M1 (High) — Verified per-file census over the FULL corpus

Population: the FULL acceptance.md corpus, scanned at BLOCK level — NOT scoped to the
conjunct-selected 55. The 55 is a prior snapshot/expectation only (spec.md §2.4: neither an
upper nor a lower bound; §2.5.1 records plan-audit-iter1's measured 59/18/41/23 populations
the 55 misses — 41 `$(git`-text files and 23 multi-line-substitution files lie outside it).
The disposition table must be able to receive those populations.

Named census steps, each a guard-safe single invocation (the tooling itself must not
textually compose git in complex forms, even when it executes no git — spec.md §2.6; use
`[g]it`-spelled patterns where the sweep's own text would otherwise trip the guard):

1. family-A sweep — git inside `$( )` assigned to a variable a later statement expands:
   corpus-wide `$(git`-text sweep in guard-safe spelling, then block-level parse;
2. multi-line open-substitution sweep — `VAR=$(…` unclosed at EOL (iter1 measured 23 files);
3. family-B sweep — tree-write chain mixing nested `$( )`/write + git + `rm -rf` tail;
4. exit-capture-only population (prior expectation ~34 inside the 55).

Take fresh guard samples per family (>= 2 per family, deterministic; /tmp-only writes).
Produce a per-file disposition table covering EVERY surfaced population with four-field
evidence per row (command, verbatim stdout, exit code, tree SHA). Expected result per the
coarse census inside the 55: ~21 family-B candidates (rewrite candidates) and ~34
exit-capture-only (measured-executable → leave unless the census proves otherwise); the
outside-55 and multi-line populations are classified on first contact — no prior expectation
exists for them.

**Ordering rationale**: this milestone is the SPEC's decision core — every downstream rewrite
decision derives from it; it executes first so review focuses on the highest-reversibility-cost
decisions (which files get rewritten) before mechanical doc edits.

### M2 (High) — Convention authoring at the primary convention home

Extend the t287 section (`worktree-integration.md:472` `## Refused Commands in a
Worktree-Isolated Session`) with: (a) the measured boundary map — refused families A/B with
their probe IDs, executable shapes P4/P6 (exit-capture), P8 (pipeline-terminator exception,
boundary beyond `wc` unmeasured), P9 (redirect-to-file capture), and the "names git without
executing it" corroboration; (b) a normative AC-authoring rule — plain separately-invocable
verbs, separate `echo "exit=$?"` capture lines, never git inside `$( )`, never a write/rm
composition tail in a verification command, never relocation into a script file (reduce
instead); (c) a working-example pointer, SPLIT BY TREE per template-internal-isolation-doctrine
§25.1 (forbidden class C1 / anti-pattern AP-25.2 — no SPEC-ID literals, audit citations,
internal dates, or commit SHAs in `internal/template/templates/**`): the C1 local copy carries
the SPEC-ID pointer to develop's AC-AEC-013 (`SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md:622-680`);
the C2 template mirror carries a GENERIC working example (an AC block restated as plain
separately-invocable verbs with separate exit-capture lines) — the two sections are identical
modulo internal-trace classes, and the C2 boundary-map text likewise carries none. Apply the
C1 + C2 template mirror; run `make build`.

**Convention-home rationale (candidates evaluated)**: (i) worktree-integration.md t287
section is PRIMARY because it is where a refused session already looks (the attribution table
lives there), it is path-scoped so the boundary map loads exactly when worktree refusals are
relevant, and it is the SSOT the refusal wording points to. (ii) verification-completeness.md
§2.1 is REJECTED as primary — it already pushes ACs toward the executable shape (exit code as
own field, single-invocation RED form) and adding a second normative home risks divergence; at
most a one-line cross-reference. (iii) moai-workflow-spec skill AC-authoring guidance is a
discoverability seam only (M4 pointer), not a second normative home.

### M3 (Medium) — Corpus disposition (rewrite verified-refusing blocks only)

Rewrite ONLY blocks the M1 census verified to refuse, to the M2 convention, preserving each
verification's semantics (observed stdout, exit code as its own field, pinned tree SHA).
Re-execute every rewritten block in the worktree-isolated session and record the passing
evidence. Files dispositioned "leave" get zero diff. [RESOLVED 2026-09-22, operator]: each file dispositioned "leave" carries its own recorded
per-file guard probe — a P4/P6 shape match alone never dispositions (this is the proxy fallacy
this SPEC exists to retire; acceptance.md §D.1/D.7 already bind disposition to a recorded probe
and "leave-verified").

### M4 (Low) — Discoverability seam and close

Add the one-line cross-pointer from the moai-workflow-spec skill AC-authoring guidance to the
t287 boundary map (Template-First mirror for the skill if template-shipped); retire normative
use of the proxy figure in favor of the verified census; sync-phase close. [RESOLVED 2026-09-22, operator]: the skill pointer lands in THIS SPEC's M4; the skill file is
template-shipped, so the edit rides Template-First + `make build` (C1/C2 discipline as applicable).

## §G Anti-Patterns

- Citing "55 signatures" as a census or a bound — it is a proxy, measured to diverge in both directions.
- Rewriting blocks that were never verified to refuse (guard-executable today).
- Moving AC commands into script files to dodge the guard.
- Reintroducing `rm -rf` cleanup tails into rewritten AC blocks (D-ITER2-09).
- Writing census tooling as a composed loop whose text names git in complex forms — the census itself will be refused.

## §H Cross-References

- spec.md §2 (verbatim probe matrix — calibration dataset the census predicates derive from)
- `.moai/reports/t1067/guard-probe-matrix-20260922.md` (in-tree calibration dataset; the t1059 M1-dossier Companion-finding content lives here — the sibling-worktree original is not citeable evidence, plan-audit-iter1 D6)
- `.moai/reports/t1067/plan-audit-iter1.md` (iteration-1 FAIL report; D1-D7 repair source, incl. the measured 59/18/41/23 census-scope figures)
- SPEC-AUDIT-EXPORT-CLAUSE-001 (prior art; do not modify)
- `.claude/rules/moai/workflow/worktree-integration.md:472` (t287 section; M2 target)
- `.claude/rules/moai/workflow/kanban-dispatch.md` § The env-isolated verification form
- `.claude/rules/moai/development/verification-completeness.md` §2.1, §4
- `.claude/rules/moai/development/rule-authoring.md` (statement duty)
