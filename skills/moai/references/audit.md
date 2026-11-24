# Audit contract

When to read: before you audit a change, as `moai-audit` or as the session that starts the audit and records the verdict. The steps of a review (package, find, verify, rank) are the AUDIT workflow's section "Review steps" (`workflows/audit.md`); this file fixes what goes in, what comes out and who writes what.

## What the auditor receives

- The exact head (40 hex) and the exact base the lane's checks ran on, as the lane's ready note names them (`references/flows/next-flow.md`).
- The card, the SPEC it names with its acceptance criteria, and the write scope of the change.
- The lane's check receipt for that head, when the lane posted one.
- The review instructions from the base commit only: `REVIEW.md`, its `@` includes and the subdirectory `REVIEW.md` files, with the project's instruction files as context. A change cannot weaken the rules it is judged by. `moai review-package` builds that package.
- An instruction to stay read-only.

## What the auditor returns

- The first line `RESULT audit <ok|blocked|not_run> <head 40 hex>`.
- Every command it ran with its exit code, and every check it did not run with the reason.
- Its findings, blocking first. Each finding names the file and the line in the head, the evidence and whether it blocks. `moai review-check` refuses a set that lacks any of these, and a finding without a file and a line counts as no finding (side rules).
- A clean `git status` in its worktree.

The auditor never writes the verdict, never fixes what it finds and never commits.

## Who records the verdict

The session that started the audit records the verdict: never the auditor agent and never the author. A session that made a commit between the base and the head, the parent session included, hands the recording to another session, such as the lander, with the returned evidence. The recording session checks that evidence itself, then adds one `moai card note` whose first line is `AUDIT <card> (<one-line scope>) head <40 hex> base <40 hex> — verdict: PASS|FAIL`. The lander reads that note before it runs `moai card land`.

The author of a change never audits it. Whoever wrote or repaired a change hands it to a separate context (`chains/index.md`); a session that made a commit between the base and the head is an author of that head.

## Audit tiers

Four lines say who audits what. The audit agent, the RUN workflow and the lander's brief carry them word for word:

- docs-only (tier 0): only documents or SPEC text change; the lander reads the change and lands it, and nothing is built.
- test-only (tier 0, tier 1 when a tool changes): only tests or test tools change; the lander gives it a light audit: it reads the change and runs the changed test targets once.
- code (tier 1): product code or skill text changes; a session that wrote none of it audits the head: it cites the lane's receipt instead of running the head again, and tries at most three mutants on changed lines, each against one or two targets.
- high-risk (tier 2): a changed path starts with a prefix of the floor list below; a session that wrote none of it gives the full audit: the code steps plus the boundary, security and data-loss checks of this contract.

The tier a ready note states never goes below the floor of its changed paths (`src/gate.rs` `tier_of`). A change that mixes classes takes the highest one. A changed path that starts with one of these prefixes sets the floor to tier 2: `src/gate`, `src/ops/gate`, `src/flow.rs`, `src/hook`, `src/ops/hook`, `src/go_hooks`, `src/ops/go_hooks`, `.github/`, `install`, `scripts/install`, `src/install`, `src/agent_install`, `src/update`, `src/self_update`, `src/home`, `src/ops/home`, `src/ops/agents_home`, `src/migrat`, `src/go_state`, `src/go_settings`, `src/ops/go_settings`, `src/state_root`, `src/secret`, `src/vault`, `src/keys`, `src/key_vault`, `src/redact`, `src/forge/audit_status`, `src/forge/transport`, `src/worktree_include`, `src/audit_gate.rs`, `.worktreeinclude`, `.claude/`, the retired second host's dot folder, `REVIEW.md`, `agents/audit.md`, `skills/moai/workflows/audit.md`, `CLAUDE.md` and the older `AGENTS.md`; and, of the moai-adk repository's own tools, `tools/local_check.sh`, `tools/ci_scope.py`, `tools/check_guard.py`, `tools/protected_checks.json`, `tools/repo_scans.json`, `tools/self_check.sh` and `tools/review_checklist.json`. This is the list the tier code reads (`moai bundle check` judges a head's floor with it).

## Rules for every audit

1. Review one exact head against one exact base.
2. Which checks the auditor runs itself and when a remote run may be cited follow the RUN workflow's section "Check-run rules" (`workflows/run.md`). A cited run is read back by the auditor itself, and the verdict names its run id, head and conclusion.
3. A first audit presents its whole threat model, boundaries and mutation list at once; a re-audit checks only the listed repairs and classifies anything new.
4. A Critical or High security finding blocks. Only an exception record the user approved — finding, severity, reason, scope, approver, expiry — lowers one; a missing or stale scan is unverified, never a pass.
5. Scratch: one folder per audit, named for it; never read or write another audit's folder.

## Side rules

- A local code review of the `<base>..<head>` range is allowed evidence, never with its `--fix` or `--comment` option. Claude Code's `/code-review` is one; it does not read `REVIEW.md`, so it never replaces the review steps or the verdict.
- A re-audit goes to the first auditor: the session that audited the head first checks the repairs, unless the lander hands the audit on with a note on the card.
- A lane's tree-bound check receipt may be cited: the receipt its ready note quotes, made on the same tree as the audited head, with its targets, the exit code of each step and its log folder. A receipt from another tree is not cited.
- No forked session and no advisor session audits: a fork carries the author's context; an advisor never sees the head.
- A finding that does not name both a file and a line in the head counts as no finding: it neither blocks nor passes.
