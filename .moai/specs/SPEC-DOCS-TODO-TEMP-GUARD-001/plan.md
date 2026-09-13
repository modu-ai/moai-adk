# SPEC-DOCS-TODO-TEMP-GUARD-001 — Implementation Plan

Tier: S (4 files affected, doc-only, < 300 LOC). Artifact-set note: the Tier S
default set is spec.md + plan.md with AC inline in spec.md §3, but the
delegation prompt explicitly requires acceptance.md; the 4-file set
(spec/plan/acceptance/progress) is emitted as instructed. AC are inline in
spec.md §3 AND enumerated in acceptance.md.

## §A Context

- Card t575: the four docs-site `utility-commands/moai-todo.md` pages make a
  claim about queue location that SPEC-TODO-HOME-TEMP-GUARD-001's guard made
  false for temporary-origin bases. No tracking owner existed (the close
  record's CHANGELOG prose line was the only mention).
- The fix is a 4-locale synchronized doc edit per the §17 4-locale sync duty.
- Worktree: `.claude/worktrees/t575` (branch `WT-docs-todo-locales`). Cards
  t657/t684 are in flight elsewhere — do NOT touch
  `.moai/specs/SPEC-TODO-QUEUE-HOME-MERGE-001` or `.moai/reports/t684`.

## §B Known Issues (grounded during plan-phase)

1. Card premise drift: the card quotes the old path `~/.moai/todo/<key>/`; the
   pages now say `~/.moai/db/<key>/todo/backlog.db` (post-t621 rekey wording).
   The defect class (unconditional claim, false for temp-origin) is unchanged.
2. Both line 59 (storage statement) and line 248 (git-metadata-less claim) are
   false for the temp-origin class; both must carry the carve-out, or AC-003's
   "old claim absent" check passes while the page stays misleading.
3. The template-shipped `.moai/docs/todo-queue-storage.md` repeats the same
   claim and is out of scope (recorded as a follow-up finding).
4. `StateDirForRoot` has no git branch: a git repository physically under a
   temp root also resolves project-local (`state_dir.go:29-31` runs on the
   resolved root), while the CLI guidance (`TempOriginRefusal`) reports such a
   repository NOT refused. The doc wording must not assert "temp → no git
   metadata" as the guard's condition; the safe wording keys on the temporary
   origin of the project directory, matching `TempOriginReason`.

## §C Pre-flight

- [ ] Confirm the 4 target files exist and carry the line-59/248 claims
      (verified at plan-phase: present, identical line numbers across locales).
- [ ] Heading baseline: 29 `^#` lines per file (ko/en/ja/zh — verified equal).
- [ ] Census artifact directory `.moai/reports/t575/` created at run-phase
      start, before the doc edits.

## §D Constraints

- Docs-only: no Go, test, template, or README file changes.
- Section parity: identical heading count and order across the four locales.
- Locale register: each locale's new sentences follow the page's existing
  native written register (no translationese), per the language rules.
- The corrected wording must state the actual resolver behavior (truth table
  in spec.md §2), including that an absolute `MOAI_HOME` override wins.

## §E Self-Verification

Machine-checkable at run-phase close:

1. `ls docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md` → 4 files.
2. Corrected-claim grep per locale (project-local temp carve-out present) → 4/4.
3. Old-sentence grep (unconditional line-248 shape) → 0 hits across 4 files.
4. `grep -c '^#'` parity → identical count across 4 files.
5. `git diff --stat` → only the 4 docs files (+ census artifact + SPEC dir).

## §F Milestones (priority-ordered, most-change-likely first)

- **M1 (High) — census artifact**: write `.moai/reports/t575/docs-census.md`
  with the queue-path claim census (hits + truth status; see acceptance.md §D
  for the table to reproduce). This is the decision-bearing artifact — its
  findings could widen or narrow M2's wording.
- **M2 (High) — 4-locale edit**: correct lines 59 and 248 in all four locale
  pages, preserving heading parity and native register. Canonical-locale order
  per the i18n rules, then derive the other three.
- **M3 (Low) — verification batch**: run the §E checks, record outputs in
  progress.md §E.2, hand to sync.

## §G Anti-Patterns

- Do not "fix" only en and machine-translate to the other locales — derive per
  the canonical-locale chain with native phrasing.
- Do not restate the guard's internal function names in user-facing docs.
- Do not edit the template storage doc "while in the area" (scope discipline;
  it is a recorded follow-up).

## §H Cross-References

- SPEC-TODO-HOME-TEMP-GUARD-001 — the guard that made the claim false.
- `.moai/docs/docs-site-i18n-rules.md` — §17 4-locale sync duty.
- Census findings and open questions: see acceptance.md §D and progress.md §E.1.
