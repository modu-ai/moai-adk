---
id: SPEC-IGNORED-EVIDENCE-CITATION-001
title: "Implementation plan — ignored evidence citation repair"
version: "0.5.0"
created: 2026-08-31
---

# Implementation Plan — SPEC-IGNORED-EVIDENCE-CITATION-001

Milestones are ordered by **decision reversibility**: the decisions most likely to change lead, the
mechanical work follows. M1 is a decision milestone with no diff.

---

## §A. Context

Card t381, worktree `.claude/worktrees/t381`, branch `WT-ignored-evidence-cite`, based on
`origin/develop` `3f03d9c36`. Scope is the 5 class-C3 lines of `.moai/reports/t381/census.md`.
Parallel card t375 (lane-8) owns the guard, the C4 instruction, and the files of spec.md §C.5.

---

## §B. Known issues carried into this plan

1. **t375 has landed** (iter3). `SPEC-EVIDENCE-CITATION-CANON-001` is `status: completed` in
   `origin/develop` (read 2026-08-31 at `9328a5242` via `git show`), and its requirement line numbers
   were re-verified unchanged. The former "unmerged draft" caveat is retired. What remains is that
   the REQ-ECC-004 boundary question below now has an **owner to ask** rather than a moving document
   to wait on.
2. **One open question, routed not decided — and it stays open by lead decision.** Does REQ-ECC-004
   (a citation shall name a single file, not a directory) reach an **output-location statement** such
   as the one in `internal/cli/audit_pin_live_test.go`? spec.md §A.4 records it. Now that t375 has
   landed, the question has an **owner to ask** rather than a moving document to wait on — but it is
   NOT resolved here. M1 keeps the conservative default below (§F M1 step 2): repair only the false
   assertion, which is the narrower change.
3. **One treatment is conditional on whether a single file can be identified** — NOT on whether the
   scratch tree exists. The repair for `.moai/reports/template-skill-improvement-plan-20260710.html`
   turns on whether one file can be named that decided the claim (spec.md §C.3). It cannot: the
   citation is the report's report-wide footer and the eight `eb01063e` JSONs are per-category batch
   audits, so treatment (b) applies. Resolved at M3; see the correction note there for why a
   tree-existence test both picks the wrong branch and answers differently in different trees.
4. **The stale-coordinate hazard is recursive.** `e2e-lint-4paths.extract.txt` broke by citing a
   `.gitignore` line number. The repair should prefer removing the coordinate over updating it,
   or it breaks again the same way.

---

## §C. Pre-flight

Run as one parallel read-only batch before any edit; persist to `.moai/reports/t381/verify/`.

```bash
git rev-parse --short HEAD ; git branch --show-current ; git rev-parse --show-toplevel
git fetch origin develop 2>&1 ; git rev-list --count --left-right origin/develop...HEAD
git grep -n '\.moai/state/verify' -- . ':!*.md' | wc -l     # expect 25
git check-ignore -v .moai/state/verify/t225                  # expect .gitignore:298
ls .moai/state/verify/ 2>&1                                  # expect No such file or directory
```

A count other than 25, or an ignore rule at a line other than 298, means the tree moved since
census.md was accepted — re-attribute before proceeding rather than reusing the census as a fresh
measurement.

---

## §D. Constraints

- Do not touch any file in spec.md §C.5 (t375-owned) or §C.6 (carve-outs).
- Do not build a guard.
- Comment/header/footer edits only; no Go statement, no assertion, no runtime path changes.
- Evidence under `.moai/reports/t381/`, never `.moai/state/`.
- Priority labels only; no time estimates.

---

## §E. Self-verification

The acceptance batch of `acceptance.md` §C is the self-verification. It runs once after M4 and again
immediately before the completion report, with `HEAD` re-read between the two — a diff-based AC read
against a stale baseline is an unattributed measurement.

**[HARD] Execution form.** This session's worktree-isolation guard refuses `for … done`, `$(…)`, and
subshells. Every acceptance command is a plain single invocation (multi-path arguments and pipes are
permitted). Moving the batch into a script file is **prohibited** — the guard cannot read inside a
script, so every check would be bypassed for that payload. Where a check cannot be expressed as one
plain invocation, the criterion is made smaller. Each command in `acceptance.md` §C was executed in
this worktree during iter2 and its observed output is recorded beside it.

---

## §F. Milestones

### M1 — Settle the repair direction (Priority: High · decision, no diff)

The most reversible-in-principle and most expensive-to-redo decision, so it leads.

1. Confirm the per-file treatment table of spec.md §C.3 against the files as they actually read
   (not against the dispatch's framing — two of its five characterizations were already wrong; see
   spec.md §C.4).
2. Resolve, or escalate to the lead as a blocker report, the REQ-ECC-004 boundary question of
   spec.md §A.4. **Conservative default if unresolved**: treat `audit_pin_live_test.go`'s sentence as
   an output-location statement outside REQ-ECC-004's reach, and repair only the false assertion
   (treatment (d)) — the narrower change, and the one lane-8 independently converged on.
3. Record the settled table in `progress.md` §E.1 before any file is edited.

Exit: the five treatments are written down and the open question is either answered or explicitly
carried forward.

### M2 — Repair the two `internal/cli/` citations (Priority: High)

Both are user-visible doctrine-shaped comments; they change what a future reader believes about
resolvability, which is the highest-consequence part of the diff.

- `internal/cli/mcp_glm.go` — treatment (a): delete the `.moai/state/…` path from the comment,
  retaining the `t225` card token and every figure verbatim. Gates: AC-IEC-003, AC-IEC-001.
- `internal/cli/audit_pin_live_test.go` — treatment (d): keep "Evidence lands in …", replace the
  "so the cited paths still resolve at audit time" clause with the truthful consequence (gitignored,
  does not survive the worktree, extract to `.moai/reports/<card-id>/` before citing). Gates:
  AC-IEC-002, AC-IEC-001.

### M3 — Repair the two path-dependent origins (Priority: Medium)

- `internal/hook/evidence_writer_zeroexec_test.go` — treatment (b): demote
  `.moai/state/verify/t341/` to an explicitly non-resolving origin note, stating that t341 decided
  not to export the raw captures. Keep the runner versions untouched. Record at the site that a
  single file cannot be named (REQ-IEC-005). Gates: AC-IEC-004, AC-IEC-005 spirit, AC-IEC-001.
- `.moai/reports/template-skill-improvement-plan-20260710.html` — **branch on identifiability**, per
  spec.md §C.3 ("(c) if the file can be identified, otherwise (b)"):
  - if a **single** file can be named that decided the claim → treatment (c): export that one file
    into `.moai/reports/`, cite it, drop the glob;
  - if no single file decides it → treatment (b): record at the site that the raw data was not
    exported and cannot be named per-file. Silently keeping the glob is prohibited.
  Gates: AC-IEC-005, AC-IEC-004, AC-IEC-001.

  > **iter5 correction — the branch test was wrong, and wrong in a way this card is about.** Earlier
  > wording branched on whether the `eb01063e` scratch tree *exists on this machine*. That is not
  > §C.3's test, and substituting it selects the wrong treatment here: the tree **does** exist (8
  > files, 244K) — so a tree-existence test picks (c) — yet (c) is unavailable, because the citation
  > is the report's **report-wide raw-data footer** and the eight JSONs are **per-category batch
  > audits** (`domain`, `foundation-core`, `foundation-quality-thinking-moai`,
  > `moai-subworkflows-team`, `ref-harness`, `workflow-project-worktree`,
  > `workflow-spec-tdd-ddd-loop`, `workflow-testing`). No single one decides a report-wide claim, and
  > exporting all eight is prohibited by REQ-ECC-004. Treatment (b) is correct.
  >
  > **The existence test is additionally unstable, which is the sharper reason to drop it.** It was
  > measured in the **primary checkout** (`/Users/goos/MoAI/moai-adk-go/.moai/state/verify/eb01063e`);
  > in *this* worktree the same path does not exist (`ls .moai/state/verify` → `No such file or
  > directory`). `.moai/state/` is gitignored and therefore tree-local, so a plan branching on its
  > existence answers differently depending on which tree runs it — a milestone gated on a
  > non-resolving path, which is the exact defect class this SPEC exists to repair. Identifiability
  > is a property of the evidence and does not vary by tree.

### M4 — Correct the recursive stale coordinate (Priority: Low · mechanical)

- `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt` — this file's citation pattern is
  already correct; only `.gitignore:284` is stale (this tree measures `:298`). **Prefer removing the
  line number** over updating it, so the same failure cannot recur. Change nothing else in the file:
  its brace-glob names export sources, not verdict evidence (argued against the REQ-ECC-004+005 pair
  at spec.md §C.3). Backed by **REQ-IEC-009**, added in iter2 — REQ-IEC-001 does not reach this file,
  since its header already declares non-resolution and it passes AC-IEC-001 today, repaired or not.
  Gate: AC-IEC-012.

### M5 — Verification batch and evidence persistence (Priority: High · mechanical)

Run `acceptance.md` §C as one batch, re-read `HEAD`, re-run the diff-based criteria, then write
`progress.md` §E.2/§E.3.

**[HARD] Evidence filenames are fixed, because AC-IEC-010 names them.** Each of the ten MUST criteria
writes its verbatim output to `.moai/reports/t381/verify/ac-iec-<NNN>.txt` — `001, 002, 003, 004,
005, 006, 007, 010, 011, 012`. AC-IEC-010's own check is an `ls` over exactly those ten paths, so a
renamed or missing file fails it. The two §D structural checks (S-1, S-2) are recorded but not gating
and need no fixed filename.

---

## §G. Anti-patterns to avoid

- **Sweeping all five with one treatment.** The whole point of M1 is that they differ; a uniform
  `sed` across the five would delete origins that are the only pointer to their values.
- **Deleting `audit_pin_live_test.go`'s whole sentence.** That satisfies a naive grep and fails
  AC-IEC-002's second half.
- **Widening the probe.** B2 alone is 467 lines. Fixing "one more while I'm here" turns this into a
  different card.
- **Re-pointing to a path that was never exported.** Citing `.moai/reports/<card>/x.json` when no such
  file was written is the same defect in a tracked directory.
- **Editing a carve-out because a widened grep flagged it.** Those paths are defined by the code;
  absent-before-creation is correct.
- **Treating census.md as a fresh measurement.** It measured `3f03d9c36`; if HEAD moved, re-measure.

---

## §H. Cross-references

- `.moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/spec.md` §A.4, §C.2-§C.5 — the canon, the
  treatments, the do-not-touch list.
- `.moai/specs/SPEC-IGNORED-EVIDENCE-CITATION-001/acceptance.md` — the verification batch: ten MUST
  criteria plus two §D structural checks.
- `.moai/reports/t381/census.md` — the accepted scope source (measured tree `3f03d9c36`).
- `SPEC-EVIDENCE-CITATION-CANON-001` (t375) — `status: completed`, landed in `origin/develop`
  (read 2026-08-31 at `9328a5242`); the canon. Guard scope REQ-ECC-007.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — the doctrinal anchor.
