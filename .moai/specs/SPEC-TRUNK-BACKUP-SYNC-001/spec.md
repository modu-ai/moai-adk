---
id: SPEC-TRUNK-BACKUP-SYNC-001
title: "Trunk backup sync: push local main to the non-reacting sync/main ref via a minimal integration verb, guarded by a workflow-exclusion regression test"
version: "0.1.0"
status: draft
created: 2026-10-17
updated: 2026-10-17
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, .github/workflows, .moai/docs"
lifecycle: spec-anchored
tags: "trunk-backup, sync-ref, git-push, workflow-triggers, regression-guard, local-main"
tier: M
---

# SPEC-TRUNK-BACKUP-SYNC-001

## §A — History

- **2026-10-17 (v0.1.0)** — plan-phase authored from factory card t1832 (operator directive 10-17, Class C, P1, proposal 1). Scope: a minimal moai verb that pushes local `main` to a remote ref no workflow reacts to (`sync/main`), a Go regression test proving every `.github/workflows` push trigger excludes that ref, and a runbook stating the leader-round step and the operator-approved origin-push scope. Hard conditions carried from the card: force-push is allowed on the sync ref only; this card ships mechanism + tests + instructions and never executes an origin push; main protection and the batch-release discipline are unchanged. Tier M: 9 requirements (ceiling 16) and 9 acceptance criteria (ceiling 16). Four operator-held decisions are recorded in `decision-index.md` (Q1 verb surface, Q2 sync ref name, Q3 force semantics, Q4 leader-round integration); REQ-TBS-001 pins provisional values until verdicts land.

## §B — Problem

Lanes merge cards into local `main` (`gitflow-lane-protocol.md` §2); `origin/main` advances only through the leader's release-batch pull requests (§4, `repo-local-pr-policy.md` — required PR, `enforce_admins: true`). Between batches, local `main` is the only copy of landed work. With `git_strategy.manual.lead_push_threshold: 20` (`.moai/config/sections/git-strategy.yaml`), up to twenty merge commits can exist only on the local machine. Machine loss or local-main corruption loses landed cards whose worktrees were already disposed of.

1. **No sanctioned frequent-push path exists.** The one side-ref push precedent is `moai integration candidate` (operator approval 2026-10-03, REQ-CCI-016): it pushes `candidateSHA:refs/heads/ci/<card>` with `--force` through a repo-scoping-scrubbed git runner (`candidateScrubbedGit`, `internal/cli/integration_candidate.go:74`). Nothing pushes local `main` itself to origin outside the batch discipline.
2. **The safety of a backup ref is untested.** A push to a new ref is safe only while no workflow reacts to it. Measured at `93f4fc75c`: 13 of the 21 files under `.github/workflows/` carry `on.push` triggers — `ci.yml` (`branches: [main, 'ci/**']`), `codeql.yml`, `graph-freshness.yml`, `lsel-leak-guard.yaml`, `release-drafter.yml`, `spec-lint.yml`, `template-neutrality-check.yaml`, `test-install.yml`, `workflow-parse-guard.yaml` (all `[main]`; several with `paths`), `docs-i18n-check.yml` (`[main]` + `paths`), `judgment-first-consistency.yaml` (`[main]` + `paths`), `label-sync.yml` (`[main, develop]` + `paths`), and `release.yml` (`tags: ['v*']`, no branches filter). Today every one of them excludes `sync/main`. Nothing guards that property: a future workflow added with a bare `on: push` (no branch filter) would fire the full CI matrix on every backup push. The regression test is the load-bearing deliverable of this card.

## §C — Goal

- Ship a minimal verb that closes the single-copy window by mirroring local `main` onto a remote ref that no workflow reacts to.
- Ship the regression test that keeps that ref non-reacting as workflows are added and edited.
- State the leader-round step and the operator-approved origin-push scope in a runbook, without adding leader code.
- Change nothing about main protection, the batch-release discipline, or the distributed templates.

## §D — Requirements (GEARS)

9 requirements (Tier M ceiling 16). `<subject>` is the named component. Operator decisions Q1–Q4 gate the provisional values of REQ-TBS-001; see §E and `decision-index.md`.

### D.1 — The backup verb

- **REQ-TBS-001** (Ubiquitous) — The `moai integration backup` verb shall push the local branch `main` of the caller's repository to the ref `refs/heads/sync/main` on the remote named `origin`, issuing exactly the git arguments `push --force origin refs/heads/main:refs/heads/sync/main`, through a git runner whose environment removes repository-location variables (the `gitenv.Env()` scrubbing the `integration candidate` runner applies at `internal/cli/integration_candidate.go:74`), executed with the caller's working directory so any worktree of the repository resolves the same shared refs. The verb shall take no flags and no arguments. On success it shall print the pushed commit SHA and the target ref. The command surface (`integration backup`), the ref name (`sync/main`), and `--force` are the provisional values held while operator decisions Q1, Q2, and Q3 are undecided; a verdict selecting another value is a one-commit amendment of this requirement, its tests, and the runbook together.
- **REQ-TBS-002** (Unwanted) — While the verb executes, it shall not issue any git command that updates, creates, or deletes a remote ref other than `refs/heads/sync/main`; it shall not pass `--force` or `-f` with any refspec other than the constant refspec of REQ-TBS-001; and it shall not push, update, or delete `refs/heads/main` on any remote. The refspec and the force flag shall be constants in the verb's source, not values computed from flags, configuration, or the environment.
- **REQ-TBS-003** (Event-driven) — When `refs/heads/main` does not resolve in the caller's repository, the verb shall refuse before any push attempt, print guidance naming the missing ref and the constant target ref, and exit non-zero.
- **REQ-TBS-004** (Event-driven) — When the push command exits non-zero, the verb shall print both captured streams of the failed command (the `factorylane.GitExitError` pattern of the candidate runner) and shall perform no further git command.

### D.2 — The workflow-exclusion regression guard

- **REQ-TBS-005** (Ubiquitous) — The repository shall carry a Go regression test that (i) enumerates every workflow file under `.github/workflows/` matching `*.yml` or `*.yaml`, (ii) reads each `on.push` trigger, and (iii) asserts that no push trigger's branch filter admits the ref name `sync/main` under GitHub's branch-filter matching: a trigger with a `branches` list fires only on listed glob patterns (`*`, `**`); a trigger with `tags` and no `branches` fires only on tag pushes; a trigger with neither `branches` nor `tags` fires on every branch push. The test shall fail when it sweeps zero workflow files carrying a push trigger (the empty-sweep guard), and shall assert detection of in-test mutant fixtures — at minimum a bare `on: push` block and a `branches` list containing `sync/**` — so the guard's failure is observed on known failing inputs.

### D.3 — Test isolation

- **REQ-TBS-006** (Ubiquitous) — Every test this SPEC adds shall run the verb only inside a temporary repository whose `origin` remote is a bare repository created under `t.TempDir()`; no test shall name the development repository's configured remotes or any network URL, and the tests shall assert that every captured push command's remote operand resolves under the test's temporary directory. Temporary repositories and bare remotes shall live only under the system temporary root that `t.TempDir()` provides — never under `$HOME` and never inside a repository checkout.

### D.4 — Runbook and scope boundaries

- **REQ-TBS-007** (Ubiquitous) — `.moai/docs/trunk-backup-runbook.md` shall exist and state: (i) the leader-round step — when a card's merge has landed on local `main`, the leader (or the operator) may run the backup verb so landed work exists on a second ref at origin between batches; (ii) that every origin push of this ref is operator-approved batch scope, per the card's hard condition; (iii) that `sync/main` is a redundancy copy only — it never lands, is never merged into any branch, and never substitutes for the `release/main-batch-YYYYMMDD` pull-request path; and (iv) that the batch-release discipline and `main` branch protection are unchanged by this ref's existence.
- **REQ-TBS-008** (Unwanted) — This SPEC's run phase shall perform no push to any real remote, shall open no pull request, and shall modify no file under `.github/workflows/`. The regression test of REQ-TBS-005 reads workflow files and writes nothing.
- **REQ-TBS-009** (Ubiquitous) — This SPEC shall change no branch-protection rule, no guard or hook configuration, and no distributed template file: `main` protection on origin (required pull request, `enforce_admins`) and every pattern of the branch guard hold their current configuration.

## §E — Constraints

- **No execution of the protected act.** The card ships mechanism + tests + instructions. The run phase never pushes to any real remote (REQ-TBS-008); the first real invocation of the verb belongs to the leader's operator-approved batch scope, after this card lands.
- **Force is scoped by construction.** Force-push is allowed on `refs/heads/sync/main` only, and only as the constant refspec of REQ-TBS-001 (REQ-TBS-002). Main protection rules are untouched (REQ-TBS-009).
- **Decision gates.** Operator decisions Q1 (verb surface), Q2 (sync ref name), Q3 (force semantics), and Q4 (leader-round integration) are recorded in `decision-index.md` with empty verdicts. REQ-TBS-001 carries provisional values while Q1–Q3 are undecided — the pattern of REQ-LMF-009 in SPEC-LOCAL-MAIN-FLOW-001. No acceptance criterion requires a real origin push, so no criterion is blocked by an undecided verdict; a verdict that selects a different value is a one-commit amendment (constant + tests + runbook).
- **Development mode is TDD** (`quality.yaml` `constitution.development_mode: tdd`): each milestone writes its failing test first and records the RED output.
- **Regression guard is the load-bearing deliverable.** The verb is deliberately minimal (no flags, no config keys, no scheduler); any proposal that grows it re-opens the simplicity ladder.

## §F — Out of Scope

### Out of Scope — executing an origin push

- This card never pushes to any real remote (REQ-TBS-008). Actual execution of the backup push is the leader's operator-approved batch act, guided by the runbook after this card lands.

### Out of Scope — modifying .github/workflows

- The run phase writes nothing under `.github/workflows/`. The regression test (REQ-TBS-005) guards the current trigger set; it does not edit it.

### Out of Scope — automating the leader round

- No scheduler, cron, hook, or factory-leader code invokes the verb automatically. The leader-round step is documentation only (REQ-TBS-007); decision Q4 records the option space.

### Out of Scope — distributed template changes

- The runbook is a repository-local document under `.moai/docs/`. No file under `internal/template/templates/` changes (REQ-TBS-009); repo-specific flow never mirrors into the distributed templates.

### Out of Scope — changing the batch-release discipline

- `origin/main` still advances only through `release/main-batch-YYYYMMDD` pull requests (`gitflow-lane-protocol.md` §4, REQ-LMF-011 of SPEC-LOCAL-MAIN-FLOW-001). `sync/main` is a redundancy copy and never becomes a landing path.

### Out of Scope — force-push anywhere except the sync ref

- No force-push to `main`, to card branches, or to candidate branches is introduced or relaxed. The Bash risk-amplifier confirmation doctrine for `git push --force` in agent sessions is unchanged; the verb is a tool path whose constant refspec is the card's sanctioned exception.

### Out of Scope — redesigning the ref name or verb surface

- Q1 and Q2 in `decision-index.md` hold the option space for the verb surface and the ref name. This SPEC pins provisional values (REQ-TBS-001) and does not settle them.

## §G — Known Relations and Non-Goals

- **Non-goal: a landing path.** The verb mirrors; it does not land. Nothing reads `sync/main` back. If a recovery ever needs it, that recovery is a separate operator act outside this SPEC.
- **Relation to SPEC-CANDIDATE-CI-001.** The backup verb copies the candidate verb's runner shape (scrubbed env, `GitExitError`, force-push of an explicit refspec) but does not use the candidate record store, the `ci/**` branches, or the landing gate. The candidate exception (REQ-CCI-016) and this backup exception are siblings, not a shared mechanism.
- **Relation to SPEC-LOCAL-MAIN-FLOW-001.** That SPEC owns the landing flow and the batch release procedure (REQ-LMF-011). This SPEC adds a redundancy copy between batches and changes none of it.
- **Relation to SPEC-LEAD-AUTOPUSH-001.** That SPEC's count trigger (`lead_push_threshold`) decides when the leader closes a batch. The backup verb runs per card merge (or per round), independent of the trigger; it neither reads nor moves that threshold.

## §H — Cross-References

- `.claude/rules/local/gitflow-lane-protocol.md` §2 (local-main landing), §4 (leader batch push; lanes never push main)
- `.claude/rules/local/repo-local-pr-policy.md` (main protection, required PR, `enforce_admins`)
- `.moai/specs/SPEC-LOCAL-MAIN-FLOW-001/spec.md` (landing flow owner; batch release procedure REQ-LMF-011)
- `.moai/specs/SPEC-CANDIDATE-CI-001/spec.md` (side-ref push precedent, REQ-CCI-016)
- `internal/cli/integration.go:280` (integration command group registration)
- `internal/cli/integration_candidate.go` (representative verb: scrubbed runner, force-push refspec, GitExitError)
- `.github/workflows/` push-trigger inventory — plan.md §A.3, measured at `93f4fc75c`

## §I — Open decisions (operator-held)

- Q1 — verb surface: `moai integration backup` subcommand vs. a new top-level verb. Recorded in `decision-index.md`.
- Q2 — sync ref name: `sync/main` (card's example) vs. alternates. Recorded in `decision-index.md`.
- Q3 — force semantics: constant-refspec `--force` vs. `--force-with-lease`. Recorded in `decision-index.md`.
- Q4 — leader-round integration: documentation-only runbook vs. leader code. Recorded in `decision-index.md`.
