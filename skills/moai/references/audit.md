# Audit contract

When to read: before you audit a change, as `moai-audit` or as the auditing session. The review steps are the AUDIT workflow's section "Review steps" (`workflows/audit.md`); this file fixes what goes in, what comes out and who writes what.

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

The auditing session records the verdict itself with `moai card verdict <card> --head <40 hex> --base <40 hex> --verdict pass|fail --tier <0-2> --evidence "<one line>"`, and for a PASS above tier 0 the lane's receipt (`--receipt-head <head>`, one `--receipt-step NAME=EXIT` per step). The reviewer is the session id the host sets. `moai card land` refuses a head that has no independent PASS of it (`no_pass`) in every mode, tier 0 included. A PASS is independent when its session wrote no commit of the head and held no reservation of the card (`self_verdict`); above tier 0 it also comes from a session other than the lander (`lander_review`) and cites a receipt of the head whose every step exited 0 (`receipt_missing`).

The leader never records a verdict for an auditor, and the `moai-audit` agent only returns evidence.

The base is the base the lane's ready note names when the head is the lane's own head, and the tip the head will land on when the head is a merge that joins the tip: `card verdict` accepts only a base that is an ancestor of the head, and an older base puts commits of other cards, maybe the auditing session's own, into the range read for authors; a verdict never names a head nobody audited.

The auditing session then adds one `moai card note`. Its first line is the verdict line, `AUDIT <card> (<one-line scope>) head <40-hex> base <40-hex> — verdict: PASS|FAIL`: lowercase hex, the em dash `—` with one space on each side, nothing after the verdict word; a first line that starts with `AUDIT ` in any other shape is an unreadable verdict, not a pass. A `tier <N>` in its scope is read too: `moai bundle check` reports a PASS below the floor of the head's changed paths as `audit_tier_low`. The next line names the auditing session as `Moai-Session: <value>`, the value `moai trailers --spec <SPEC id>` prints in the auditor's own worktree (`internal/commit-trailers`). For example:

```text
AUDIT C-12 (repo-scans 4/4, tier 1) head 997af563452bfb36f2ce722a66e9068dae191568 base 9c8a8c51c9c059be389deeaebb19a3e45709b273 — verdict: PASS
Moai-Session: <session id>:<worktree key>
```

The author of a change never audits it. Whoever wrote or repaired a change hands it to a separate context (`chains/index.md`); a session that made a commit between the base and the head is an author of that head.

## What blocks

A verdict is FAIL for four things only: behaviour that differs from the card's acceptance criteria, data loss, a security problem, and a weakened gate. Every other finding (missing test evidence, SPEC or document mismatch, wording, comments, language) is a non-blocking line of the verdict note and the verdict is PASS. Each finding carries its class: `behaviour`, `data-loss`, `security`, `gate` or `non-blocking`. A second FAIL of one card goes to the leader, not to a re-audit.

## Audit tiers

Four lines say who audits what. The audit agent, the RUN workflow and the auditor's brief carry them word for word:

- docs-only (tier 0): only documents or SPEC text change; the auditor records a PASS without a receipt, and nothing is built.
- test-only (tier 0, tier 1 when a tool changes): only tests or test tools change; the auditor gives it a light audit: it reads the change and the lane's receipt and runs nothing.
- code (tier 1): product code or skill text changes; a session that wrote none of it gives a light audit: it cites the lane's receipt instead of running the head again and runs no mutant.
- high-risk (tier 2): a changed path starts with a prefix of the floor list below; a session that wrote none of it gives the full audit: the code steps plus the boundary, security and data-loss checks of this contract, with the lane's guard-mutant lines read and at most two named test modules run.

The tier a ready note states never goes below the floor of its changed paths (`src/gate.rs` `tier_of`). A change that mixes classes takes the highest one. A changed path that starts with one of these prefixes sets the floor to tier 2: `src/gate`, `src/test_weakening`, `src/ops/gate`, `src/flow.rs`, `src/hook`, `src/ops/hook`, `src/go_hooks`, `src/ops/go_hooks`, `.github/`, `install`, `scripts/install`, `src/install`, `src/agent_install`, `src/update`, `src/self_update`, `src/home`, `src/ops/home`, `src/ops/agents_home`, `src/migrat`, `src/go_state`, `src/go_settings`, `src/ops/go_settings`, `src/state_root`, `src/secret`, `src/vault`, `src/keys`, `src/key_vault`, `src/redact`, `src/forge/audit_status`, `src/forge/transport`, `src/worktree_include`, `src/audit_gate.rs`, `.worktreeinclude`, `.claude/`, the retired second host's dot folder, `REVIEW.md`, `agents/audit.md`, `skills/moai/workflows/audit.md`, `CLAUDE.md` and the older `AGENTS.md`; and the repository's own check tools. This is the list the tier code reads (`moai bundle check` judges a head's floor with it). A change of the repository's scan registry sets the floor only when it does more than grow the registry: a rule file, a test name, a `--test` option or a scan added is growth (`gate::registry_only_grew`); any other change of it, or a text that cannot be read, is tier 2. `moai check affected --card` reads the same rule: growth keeps the card set, anything else runs the full set.

## Rules for every audit

1. Review one exact head against one exact base.
2. Which checks the auditor runs itself and when a remote run may be cited follow `internal/check-rules`; the verdict names a cited run's id, head and conclusion.
3. A re-audit checks only the listed repairs and classifies anything new by the four kinds.
4. A Critical or High security finding blocks. Only an exception record the user approved — finding, severity, reason, scope, approver, expiry — lowers one; a missing or stale scan is unverified, never a pass.
5. Scratch: one folder per audit, named for it; never read or write another audit's folder.
6. Read another head from your own worktree by object id (`git show <sha>:<path>`, `git diff <base> <head>`, `git merge-tree`); never `cd` or `git -C` into a worktree that is not your own.

## Side rules

- A local code review of the `<base>..<head>` range is allowed evidence, never with its `--fix` or `--comment` option; it never replaces the review steps or the verdict.
- A re-audit goes to the first auditor, unless the lander hands it on with a note on the card.
- A lane's tree-bound check receipt may be cited: the one its ready note quotes, made on the audited head's tree, with its targets, step exit codes and log folder.
- No forked session and no advisor session audits.
- A finding that does not name both a file and a line in the head counts as no finding: it neither blocks nor passes.
