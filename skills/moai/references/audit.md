# Audit contract

When to read: before you audit a change, as `moai-audit` or as the session that starts the audit and records the verdict. The steps of a review (package, find, verify, rank) are the AUDIT workflow's section "Review steps" (`workflows/audit.md`); this file fixes what goes in, what comes out and who writes what. How the audit session is started is not part of it.

## What the auditor receives

- The exact head (40 hex) and the exact base the lane's checks ran on, as the lane's ready note names them (`references/flows/next-flow.md`).
- The card, the SPEC it names with its acceptance criteria, and the write scope of the change.
- The lane's check receipt for that head, when the lane posted one.
- The review instructions from the base commit only: `REVIEW.md`, its `@` includes and the subdirectory `REVIEW.md` files, with the project's instruction files as context. A change cannot weaken the rules it is judged by. `moai review-package` builds that package.
- An instruction to stay read-only.

## What the auditor returns

- The first line `RESULT audit <ok|blocked|not_run> <head 40 hex>`.
- Every command it ran with its exit code, and every check it did not run with the reason.
- Its findings, blocking first. Each finding names the file and the line in the head, the evidence and whether it blocks. `moai review-check` refuses a finding set that lacks a severity, a file and line in the head, or evidence, so a finding without both a file and a line stays out of the set: it counts as no finding (side rules).
- A clean `git status` in its own worktree: a change there makes its findings unusable.

The auditor never writes the verdict, never fixes what it finds and never commits.

## Who records the verdict

The session that started the audit records the verdict: never the auditor agent and never the author. A session that made a commit between the base and the head, the parent session included, hands the recording to another session, such as the lander, with the returned evidence. The recording session checks that evidence itself, then adds one `moai card note` whose first line is `AUDIT <card> (<one-line scope>) head <40 hex> base <40 hex> — verdict: PASS|FAIL`, naming the head the auditor examined. The lander reads that note before it runs `moai card land`.

The author of a change never audits it. Whoever wrote or repaired a change hands it to a separate context (`chains/index.md`); a session that made a commit between the base and the head is an author of that head.

## Rules for every audit

1. Review one exact head against one exact base.
2. Which checks the auditor runs itself and when a remote run may be cited follow the RUN workflow's section "Check-run rules" (`workflows/run.md`). A cited run is read back by the auditor itself, and the verdict names its run id, head and conclusion.
3. A first audit presents its whole threat model, boundaries and mutation list at once; a re-audit checks only the listed repairs and classifies anything new.
4. A Critical or High security finding blocks. Only an exception record the user approved — finding, severity, reason, scope, approver, expiry — lowers one; a missing or stale scan is unverified, never a pass.

## Side rules

- A local code review of the `<base>..<head>` range is allowed evidence, never with its `--fix` or `--comment` option: it reads and reports, changes no file and posts nothing. Claude Code's `/code-review` is one; it reads `CLAUDE.md` but not `REVIEW.md`, so it adds evidence and never replaces the review steps or the verdict.
- A re-audit goes to the first auditor: the session that audited the head first checks the repairs, unless the lander hands the audit on with a note on the card.
- A lane's tree-bound check receipt may be cited: the receipt its ready note quotes, made on the same tree as the audited head, with its targets, the exit code of each step and its log folder. A receipt from another tree is not cited.
- No forked session and no advisor session audits: a fork carries the author's context, and an advisor sees only what it is shown, not the head.
- A finding that does not name both a file and a line in the head counts as no finding: it neither blocks nor passes.
