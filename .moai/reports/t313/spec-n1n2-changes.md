# t313 — post-audit debt repair (N1, N2) · spec 0.3.0 → 0.3.1

Card: t313 · SPEC-WORKTREE-BASEREF-001 · worktree `.claude/worktrees/t313`, branch `WT-worktree-baseref`.
Input verdict: `.moai/reports/t313/plan-audit-iter2.md` — PASS-WITH-DEBT 0.92 harmonic, 7/7 must-pass clear, zero blocking defects.
Operator kickoff approval: GRANTED, conditional on repairing N1 first. Run-phase progression mode: autonomous.

Scope touched: `acceptance.md`, `plan.md`, plus the one permitted `spec.md` frontmatter/HISTORY edit and `progress.md`. No production code, no config, no hooks, no templates. Neither `tab_schema.json` touched (card t316 owns them).

---

## REPAIR 1 — N1 (AC-WBR-013 check (1) false NO-TEMPLATE-COUNTERPART)

**What was wrong.** The inner probe required the template counterpart to be a same-named **plain** file under `internal/template/templates/`. This SPEC's own plan §D write list includes `.moai/config/sections/git-strategy.yaml`, whose counterpart ships only as `git-strategy.yaml.tmpl` — a fact the SPEC already records at plan.md §B G5. A correct implementation therefore emitted one false `NO-TEMPLATE-COUNTERPART` line and the MUST criterion FAILED.

**What changed.** `acceptance.md:239` (the probe line; the block now spans `acceptance.md:227-245`). The probe passes BOTH candidate paths to one `git diff --name-only` pathspec list, so it reports only when NEITHER form was changed in the diff:

```
      git diff --name-only "$BASE"..HEAD -- \
        "internal/template/templates/$f" "internal/template/templates/$f.tmpl" | grep -q . \
        || echo "NO-TEMPLATE-COUNTERPART $f"
```

Five comment lines were added above the pipeline (`acceptance.md:229-234`) recording why the plain-only probe is wrong and citing plan.md §B G5, so a later reader does not "simplify" the `.tmpl` arm back out.

### N1 re-measurement (verbatim)

Measured against **real repository history** rather than a synthetic fixture. Commit `664cd6eae` changed exactly the pair this SPEC's §D write list produces — the local mirror plus its `.tmpl` counterpart, and no plain-file counterpart:

```
$ git diff --name-only 664cd6eae^..664cd6eae -- .moai/config/sections/git-strategy.yaml internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
.moai/config/sections/git-strategy.yaml
internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
```

**Positive control — the OLD probe emits the false line:**

```
$ git diff --name-only 664cd6eae^..664cd6eae -- .moai/config/sections/git-strategy.yaml | while read -r f; do git diff --name-only 664cd6eae^..664cd6eae -- "internal/template/templates/$f" | grep -q . || echo "NO-TEMPLATE-COUNTERPART $f"; done
NO-TEMPLATE-COUNTERPART .moai/config/sections/git-strategy.yaml
```

**The NEW probe emits nothing on the same input:**

```
$ git diff --name-only 664cd6eae^..664cd6eae -- .moai/config/sections/git-strategy.yaml | while read -r f; do git diff --name-only 664cd6eae^..664cd6eae -- "internal/template/templates/$f" "internal/template/templates/$f.tmpl" | grep -q . || echo "NO-TEMPLATE-COUNTERPART $f"; done; echo "NEW-PROBE-DONE rc=$?"
NEW-PROBE-DONE rc=0
```

**Negative control — detection was NOT weakened.** Commit `63b4628a6` changed the local mirror with NEITHER counterpart form changed (a genuine orphan). The new probe still reports it:

```
$ git diff --name-only 63b4628a6^..63b4628a6 -- .moai/config/sections/git-strategy.yaml | while read -r f; do git diff --name-only 63b4628a6^..63b4628a6 -- "internal/template/templates/$f" "internal/template/templates/$f.tmpl" | grep -q . || echo "NO-TEMPLATE-COUNTERPART $f"; done; echo DONE
NO-TEMPLATE-COUNTERPART .moai/config/sections/git-strategy.yaml
DONE
```

Both controls were run in worktree `.claude/worktrees/t313` at HEAD `48eb945df`. The positive control is what makes the green meaningful: the check's RED has been observed on a known failing input, not merely its green.

**Residual risk.** The probe now accepts a `.tmpl` counterpart for ANY changed `.claude/` / `.moai/` path, not only for paths whose counterpart is genuinely `.tmpl`-only. A diff that changed `foo.md` locally and `foo.md.tmpl` in the template tree would pass even if a plain `foo.md` counterpart also existed and was left stale. This is strictly narrower than the pre-repair false-positive it removes, and the `.sh`/`.sh.tmpl` twin-drift case is check (2)'s job — but it is a real widening and is recorded rather than hidden.

---

## REPAIR 2 — N2 (AC-WBR-016 "read seam" denotation)

**Resolution chosen: (b) — the seam is the alignment-ENTRY helper (the configured-value read).** The narrowing alternative (a) was rejected.

**Why (b), not (a).** Half 1 exists for exactly one purpose: to catch an implementation that is behaviourally perfect on every outcome AC and never wired into the SessionStart errgroup (REQ-WBR-004; plan-audit iter-1 D1). Under (a), half 1's `Given` would be narrowed to a NON-EMPTY configured value — but REQ-WBR-003 requires the **shipped default to be empty**, so the unset case is the common case, and an implementation that never registers the task would pass the criterion on every default installation. A resolution that lets the criterion lapse precisely where it is most needed is not a resolution. Under (b) the count is 1 for every configured value, empty included, so the "is it wired in at all?" assertion holds unconditionally.

This is the lead's reading and I concur; no overrule.

**Consistency check performed.** plan.md:91 gates the whole helper on the primary checkout **before** the configured-value read, so from a linked worktree the entry seam is invoked 0 times — which is exactly what half 2 asserts. Both halves therefore read the same seam under one denotation; no contradiction was introduced.

**What changed.**

- `acceptance.md:297` — a new preamble paragraph immediately under the AC-WBR-016 heading pins the denotation, names what the seam is NOT (the `origin/HEAD` read; the `git remote set-head` write seam), states the ordering + REQ-WBR-005 reason the ambiguity existed, and closes with: *"A run-phase test asserting on any other seam does NOT satisfy this criterion."* — so the run-phase test and the criterion cannot disagree.
- `acceptance.md:302` — half 1's `Then` restated to cite the pinned seam and to say explicitly that the count holds identically for a set value and an empty one.
- `acceptance.md:307` — half 2's `Then` restated to cite the same pinned seam and to name the primary-checkout gate (plan.md:91) as the reason the count is 0.
- `plan.md:54-62` — new decision entry **§A D3.2 — The read seam is the alignment-entry (configured-value) read**, recording the ambiguity, the ruling, the rejected alternative with its reason, and the consequence for the implementation shape.
- `plan.md:103` — M2's test bullet now names which seam the two firing-point tests count, and requires the `origin/HEAD` read to stay a separate seam.

The audit's explicit instruction — *"Do not resolve it by silently writing the permissive test"* — is satisfied: the resolution is recorded in the criterion itself and in the plan, before any test exists.

**Residual risk.** The entry seam must be a distinct function-variable from the `origin/HEAD` read. An implementer who collapses both reads into one seam function would satisfy half 1 trivially while losing the ability to assert the empty-path no-git-read property REQ-WBR-005 requires. plan.md:103 now states the separation requirement; nothing mechanically enforces it until M2's tests exist.

---

## Other edits

- `spec.md` frontmatter `version:` `"0.3.0"` → `"0.3.1"`; one HISTORY row added naming N1 and N2 (the single permitted spec.md edit).
- `progress.md` §E.1 rewritten: plan-audit iter-2 verdict + verdict path, operator kickoff approval with N1 as its condition and the condition's discharge, the autonomous progression mode, and the seven debt items from the verdict's § Debt Carried Into Run-Phase (N1/N2 marked REPAIRED; items 3-7 plus inherited G2/G4/G6 carried forward verbatim in substance).
- The cosmetic HISTORY row ordering noted by the auditor (0.1.0 → 0.3.0 → 0.2.0) was left untouched — out of the stated scope, and non-scoring.

## Count verification (after all edits)

```
$ grep -oh 'REQ-WBR-[0-9]\{3\}' .moai/specs/SPEC-WORKTREE-BASEREF-001/*.md | sort -u | wc -l
      16
$ grep -o 'AC-WBR-[0-9]\{3\}' .moai/specs/SPEC-WORKTREE-BASEREF-001/acceptance.md | sort -u | wc -l
      16
```

16 REQ / 16 AC — exactly at the Tier M ceiling, unchanged. No id added, renamed, or renumbered.

## Gaps (explicitly NOT observed)

- `make build` was NOT run (no template or Go source was touched; nothing to rebuild).
- No Go test was run — this change touches SPEC artifacts only.
- AC-WBR-013 check (2) was NOT re-measured: it is dormant for this SPEC (plan §D lists no hook wrapper) and its N4 path-mixing defect is knowingly carried per the verdict.
- The other five carried debt items were not re-litigated, per the dispatch.

## Commit

Subject: `feat(SPEC-WORKTREE-BASEREF-001): plan-phase artifacts (Tier M, 3 artifacts)`, on branch `WT-worktree-baseref`, parent `48eb945df` — one commit, not pushed. Staged by explicit pathspec (`.moai/specs/SPEC-WORKTREE-BASEREF-001/` + `.moai/reports/t313/`); no sweep-stage.

The commit SHA is deliberately NOT written here: a commit cannot carry its own hash, and an amend to insert it changes the very value inserted (observed — `5cb3ea707` became `f402f5c77` on the first attempt). Resolve it from the branch tip: `git -C .claude/worktrees/t313 rev-parse --short WT-worktree-baseref`.
