# Internal: review

When to read: before you review or audit a change, as `moai-audit` or as the parent session that writes the verdict.

This note says where the review rules live; it does not restate them.

1. Review one exact head against one exact base. The steps — package, find, check, verdict — are the AUDIT workflow's section "Review steps" (`workflows/audit.md`).
2. Take the review instructions from the base commit only: `REVIEW.md`, its `@` includes and the subdirectory `REVIEW.md` files, with `AGENTS.md` and `CLAUDE.md` as context. A pull request cannot weaken the rules it is judged by. `moai-cli review-package` builds that package; `moai-cli review-check` refuses a finding set that lacks a severity, a file and line in the head, or evidence.
3. Which checks you run yourself and when you may cite an exact-head CI run follow the VERIFY workflow's section "Check-run rules" (`workflows/verify.md`). Read the run back yourself and cite its run id, head and conclusion in the verdict.
4. Never review a change you wrote or repaired. The agent that implemented a change hands it to a separate context (`chains/index.md`).
5. Each finding names the file and line, the evidence and whether it blocks. A first audit presents its whole threat model, boundaries and mutation list at once; a re-audit checks only the listed repairs and classifies anything new.
