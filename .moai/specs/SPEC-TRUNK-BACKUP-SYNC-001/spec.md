---
id: SPEC-TRUNK-BACKUP-SYNC-001
title: "Trunk backup sync: push local main to the non-reacting sync/main ref via a minimal integration verb, guarded by a workflow-exclusion regression test"
version: "0.3.3"
status: completed
created: 2026-10-17
updated: 2026-10-21
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, .github/workflows, .moai/docs"
lifecycle: spec-anchored
tags: "trunk-backup, sync-ref, git-push, workflow-triggers, regression-guard, local-main"
tier: M
amendment_of: SPEC-TRUNK-BACKUP-SYNC-001
---

# SPEC-TRUNK-BACKUP-SYNC-001

## §A — History

- **2026-10-17 (v0.1.0)** — plan-phase authored from factory card t1832 (operator directive 10-17, Class C, P1, proposal 1). Scope: a minimal moai verb that pushes local `main` to a remote ref no workflow reacts to (`sync/main`), a Go regression test proving every `.github/workflows` push trigger excludes that ref, and a runbook stating the leader-round step and the operator-approved origin-push scope. Hard conditions carried from the card: force-push is allowed on the sync ref only; this card ships mechanism + tests + instructions and never executes an origin push; main protection and the batch-release discipline are unchanged. Tier M: 9 requirements (ceiling 16) and 9 acceptance criteria (ceiling 16). Four operator-held decisions are recorded in `decision-index.md` (Q1 verb surface, Q2 sync ref name, Q3 force semantics, Q4 leader-round integration); REQ-TBS-001 pins provisional values until verdicts land.

- **2026-10-17 (v0.2.0)** — plan-audit iteration 1 repairs (verdict FAIL 0.81, receipt `rcpt-b7c797c42df44e09e1bc9046`; blocking D1..D6 only — optional D7..D10 deferred). (1) REQ-TBS-001's constant argv gains `--no-follow-tags` (D2): under a caller's `push.followTags=true` the previous argv would push annotated tags reachable from `main`, updating a second remote ref and able to fire `release.yml`; REQ-TBS-002 names `--no-follow-tags` among the constants, and AC-TBS-004 gains the followTags fixture. (2) REQ-TBS-005 states the matcher's supported grammar subset explicitly — `*` within one path segment, `**` crossing `/`, bracket character classes, leading-`!` negation with last-match-wins — mandates fail-loud on unparseable patterns, and adds the bidirectional mutant set (D3). (3) plan.md schedules the `.github/test-input-filters.yml` `go_code` extension mapping `.github/workflows/**` (D4), with the explicit note that the file lives under `.github/`, outside REQ-TBS-008's `.github/workflows/` prohibition; new AC-TBS-010 asserts it. (4) AC-TBS-006 gains the no-workflow-edit half of REQ-TBS-008, measured as a diff against the run-phase baseline SHA (D5). (5) acceptance.md §E carries RED-now evidence-ledger cells for every release-blocking criterion, executed and pinned to tree `8bff29c5f` (D1). (6) plan.md's stale AC ranges corrected (D6). Requirements unchanged (9); acceptance criteria 9 → 10 (AC-TBS-010).

- **2026-10-18 (v0.3.0)** — plan-audit iteration 2 repairs (verdict FAIL 0.81, receipt `rcpt-160d6c8df9a299c42e91e0a4`, audited commit `72135c81f`; blocking D11..D13 only — optional D14/D15 deferred). All six round-1 defects were ruled RESOLVED in substance; the three blocking items are transcription-level defects the repair itself introduced. (1) The character-class mutant fixture is corrected to `['sync/m[a-z]in']` in REQ-TBS-005, AC-TBS-002, and plan M1 (D11): `[a-i]` consumes exactly one character, so the former `sync/m[a-i]n` could never match `sync/main` — an impossible-direction criterion; the corrected pattern matches (`sync/m` + `[a-z]`→`a` + `in`). (2) AC-TBS-010's verification prescription is replaced with YAML parsing of the `go_code` sequence entries — the former single-grep check is satisfiable by a comment-only mutant (D12). (3) RED-now cell E-TBS-R1's stdout is recaptured wholesale by file redirect and embedded raw (D13) — the prior block paraphrased one line (`store the record keyed to` for the build's `store its record keyed by`). Requirement and AC counts unchanged (9 / 10).

- **2026-10-18 (v0.3.1)** — run-phase debt disposition (plan-audit `dispose_in=run` debts D14, D15, D8, D10; D7/D9 stay deferred to sync). Implemented M1–M3 landed at `4571db121` (10/10 AC PASS) before this docs-only iteration; no implementation change is in scope here. (1) REQ-TBS-005 gains clause (iv), the continued-firing obligation — the `.github/test-input-filters.yml` `go_code` mapping of `.github/workflows/**` — giving AC-TBS-010's scheduling assertion its requirement-layer home (D14; the mapping is implemented and verified in the tree). (2) The grammar subset's exclusion of `?` and `+` is stated explicitly as out-of-subset-by-design with fail-loud as the shipped behavior, keeping REQ-TBS-005 truthful against the implementation (D15). (3) The §E one-commit-amendment clause now enumerates the full provisional-value footprint across REQ-TBS-001/002/005/007, AC-TBS-001..005, the mutant fixtures, and the runbook, and names the re-delegation path (D8). (4) AC-TBS-008's severity aligns to Must, matching its release-blocking class and its status as card deliverable 3 (D10). Requirement and AC counts unchanged (9 / 10).

- **2026-10-18 (v0.3.2)** — sync-phase debt disposition (plan-audit deferred debts D7 and D9; docs only, no implementation change). (1) The legacy negative-form requirements are reworded to the current canonical shapes ahead of the 2026-11-22 lint-acceptance window close (D7): REQ-TBS-002 becomes State-driven and states the invariant positively (every remote ref other than `refs/heads/sync/main` untouched; `--force`/`-f` only with the constant refspec; `refs/heads/main` holding its pre-invocation value), REQ-TBS-008 becomes Ubiquitous and states the run phase's untouched surfaces, REQ-TBS-009 states unchanged configuration positively, and REQ-TBS-004's "no further git command" clause becomes "the failed push shall be the verb's last git command". (2) Implementation-location anchors are removed from the REQ bodies (D9): REQ-TBS-001's parenthetical now names the repository-scoping scrub behaviorally and points to plan.md §A.2 for the runner anchor, and REQ-TBS-004 drops the `factorylane.GitExitError` symbol for the same cross-reference. Lint verified clean before and after. Requirement and AC counts unchanged (9 / 10).

- **2026-10-18 (v0.3.3)** — sync-audit round-1 REQ-level amendments (verdict FAIL, receipt `rcpt-8101894e528de70218f97c11`, audited `4304e1fe6`; this docs iteration precedes the implementation repair and changes no code). (1) REQ-TBS-001's constant argv gains `--no-recurse-submodules` (S1, P1): under a caller's `push.recurseSubmodules=only` the previous argv exits 0 as a silent no-op false success — the remote gains no `refs/heads/sync/main` — and under `on-demand` the push reaches submodule remotes, a second remote surface; REQ-TBS-002's constants list adds the same flag, and the §E amendment checklist already reaches the argv constant through REQ-TBS-001. (2) REQ-TBS-005's grammar pins `**/` segment semantics explicitly (S2, P2): a `**/` in segment position, including a leading `**/` prefix, matches zero or more path segments — `**/sync/main` and `sync/**/main` both admit `sync/main` — so the zero-segment case lives in the requirement, ahead of the matcher fixtures the implementation repair adds. S3 (stale printed SHA) is NOT amended at the REQ layer here (its report-the-actually-pushed-SHA remedy satisfies REQ-TBS-001's "shall print the pushed commit SHA" as written); sync-audit round 2 (R2-F1/R2-F2, P2 blocking, independently reproduced) then extended the amendment to REQ-TBS-001's argv form itself — the remedy's post-push `ls-remote` read queries the fetch url while the push goes to `remote.origin.pushurl` (R2-F1: a divergent pushurl prints a SHA never pushed) and races a competing writer landing after the push (R2-F2: the report names someone else's push) — so the pinned refspec `<sha>:refs/heads/sync/main` replaced the bare branch refspec and the post-push lookup was deleted (REQ-TBS-002's constants clause aligned; R2-F3's silent downgrade on read failure dies by construction). (3) Ride-along (S6): the v0.3.1 row's debt enumeration is completed — D16 (stale `updated:` frontmatter, plan-audit-3) was disposed in substance with the run-phase commits (all three artifacts read `updated: 2026-10-18`, verified by the sync audit); the v0.3.1 row's omission was enumeration-only. Requirement and AC counts unchanged (9 / 10).

## Amendments

- **2026-10-18 (amendment of the v0.3.2 close)** — in-place amendment, `completed → in-progress` per the Status Transition Ownership Matrix (manager-spec, orchestrator re-delegation). Prior completed version: v0.3.2 (the close state). `prior_completed_sha: af478d813` — equal to progress.md §E.4's `sync_commit_sha`, so the `SyncStatusDrift` check exempts this amendment. Rationale: sync-audit round 1 FAIL (receipt `rcpt-8101894e528de70218f97c11`, audited `4304e1fe6`) — blocking findings S1 (`push.recurseSubmodules` silent no-op false success / submodule remote surface), S2 (`**/` zero-segment undermatch against GitHub semantics), S3 (stale printed SHA). Scope: the REQ-TBS-001/002/005 amendments already applied in v0.3.3 (S1/S2 at the REQ layer; S3 needs no REQ change — reporting the actually-pushed SHA satisfies REQ-TBS-001 as written), plus the implementation repair for S1/S2/S3 (constant argv with `--no-recurse-submodules`, matcher `**/` segment semantics with bidirectional fixtures, config-hazard test matrix) and — scope extended by sync-audit round 2 (R2-F1/R2-F2/R2-F3) — the pushed-SHA reporting realized as the pinned refspec `<sha>:refs/heads/sync/main` (REQ-TBS-001's argv form and REQ-TBS-002's constants clause amended in place; the post-push `ls-remote` lookup deleted), all under this one amendment.

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

- **REQ-TBS-001** (Ubiquitous) — The `moai integration backup` verb shall push the local branch `main` of the caller's repository to the ref `refs/heads/sync/main` on the remote named `origin`, issuing exactly the git arguments `push --force --no-follow-tags --no-recurse-submodules origin <sha>:refs/heads/sync/main`, where `<sha>` is the pinned form — the full SHA of local `main` resolved once at invocation start — so the commit transmitted is exactly the commit the verb reports (the pin holds under any `remote.origin.pushurl` divergence from the fetch url, and no post-push remote read exists to rename the report) (the `--no-follow-tags` constant holds even where a caller's gitconfig sets `push.followTags=true`, so no annotated tag reachable from `main` rides the push — a tag ref would update a second remote ref and can fire `release.yml`; the `--no-recurse-submodules` constant holds even where a caller's gitconfig sets `push.recurseSubmodules` — under `only` the argv without it exits 0 as a silent no-op while the remote gains no `refs/heads/sync/main`, and under `on-demand` the push reaches submodule remotes, a second remote surface against REQ-TBS-002's invariant), through a git runner whose environment removes repository-location variables (the repository-scoping environment scrub the `integration candidate` verb applies; the runner's anchor is recorded in plan.md §A.2), executed with the caller's working directory so any worktree of the repository resolves the same shared refs. The verb shall take no flags and no arguments. On success it shall print the pinned transmitted commit SHA — the `<sha>` of the refspec — and the target ref. The command surface (`integration backup`), the ref name (`sync/main`), and `--force` are the provisional values held while operator decisions Q1, Q2, and Q3 are undecided; a verdict selecting another value is a one-commit amendment of this requirement, its tests, and the runbook together.
- **REQ-TBS-002** (State-driven) — While the verb executes, every git command it issues shall leave every remote ref other than `refs/heads/sync/main` untouched; `--force` and `-f` shall appear only with the pinned refspec of REQ-TBS-001; and `refs/heads/main` on every remote shall hold its pre-invocation value. The target-side ref name of the refspec, the force flag, `--no-follow-tags`, and `--no-recurse-submodules` shall be constants in the verb's source, not values computed from flags, configuration, or the environment; the refspec's source side shall be exactly the pinned SHA of REQ-TBS-001, resolved once at invocation start from the repository's own `refs/heads/main` — never a value computed from flags, configuration, or the environment, and never the bare branch refspec `refs/heads/main:refs/heads/sync/main`.
- **REQ-TBS-003** (Event-driven) — When `refs/heads/main` does not resolve in the caller's repository, the verb shall refuse before any push attempt, print guidance naming the missing ref and the constant target ref, and exit non-zero.
- **REQ-TBS-004** (Event-driven) — When the push command exits non-zero, the verb shall print both captured streams of the failed command (the both-streams failure-reporting pattern of the `integration candidate` verb; plan.md §A.2 carries the anchor), and the failed push shall be the verb's last git command.

### D.2 — The workflow-exclusion regression guard

- **REQ-TBS-005** (Ubiquitous) — The repository shall carry a Go regression test that (i) enumerates every workflow file under `.github/workflows/` matching `*.yml` or `*.yaml`, (ii) reads each `on.push` trigger, and (iii) asserts that no push trigger's branch filter admits the ref name `sync/main` under GitHub's branch-filter matching. The matcher's supported grammar is exactly: `*`, which matches within one path segment and never crosses `/`; `**`, which crosses `/` — a `**/` in segment position, including a leading `**/` prefix, matches zero or more path segments, so `**/sync/main` and `sync/**/main` both admit `sync/main`; bracket character classes `[...]`; and a leading `!`, which negates its pattern, with later patterns overriding earlier ones (last match wins). A pattern element outside this subset, or a pattern the matcher cannot parse, shall fail the test loudly rather than be treated as non-matching. Out of the subset by design at this version are the single-character wildcard `?` and the one-or-more quantifier `+` (both defined in GitHub's filter documentation): the matcher fail-louds on them today, and admitting either is a grammar-extension amendment to this requirement — never a silent behavior change. A trigger with a `branches` list fires only on patterns that match; a trigger with `tags` and no `branches` fires only on tag pushes; a trigger with neither `branches` nor `tags` fires on every branch push. The test shall fail when it sweeps zero workflow files carrying a push trigger (the empty-sweep guard), and shall assert detection of in-test mutant fixtures — at minimum a bare `on: push` block, a `branches` list containing `sync/**`, and the bidirectional set: `['**']` admits `sync/main`; `['*']` does not; `['**', '!sync/main']` does not (last-match negation wins); `['sync/m[a-z]in']` does (character-class support) — so the guard's failure is observed on known failing inputs in both directions. (iv) The repository's test-selection configuration shall schedule that test on workflow-only diffs: `.github/test-input-filters.yml` maps `.github/workflows/**` into its `go_code` path filter, so a diff touching only workflow files — a new file, or an edit to any workflow file — classifies as Go and runs the suite; the mapping file lives under `.github/`, not `.github/workflows/`, so the edit is outside REQ-TBS-008's prohibition. This is the continued-firing obligation (verification-completeness §1.3): the guard runs exactly when its guarded surface changes.

### D.3 — Test isolation

- **REQ-TBS-006** (Ubiquitous) — Every test this SPEC adds shall run the verb only inside a temporary repository whose `origin` remote is a bare repository created under `t.TempDir()`; no test shall name the development repository's configured remotes or any network URL, and the tests shall assert that every captured push command's remote operand resolves under the test's temporary directory. Temporary repositories and bare remotes shall live only under the system temporary root that `t.TempDir()` provides — never under `$HOME` and never inside a repository checkout.

### D.4 — Runbook and scope boundaries

- **REQ-TBS-007** (Ubiquitous) — `.moai/docs/trunk-backup-runbook.md` shall exist and state: (i) the leader-round step — when a card's merge has landed on local `main`, the leader (or the operator) may run the backup verb so landed work exists on a second ref at origin between batches; (ii) that every origin push of this ref is operator-approved batch scope, per the card's hard condition; (iii) that `sync/main` is a redundancy copy only — it never lands, is never merged into any branch, and never substitutes for the `release/main-batch-YYYYMMDD` pull-request path; and (iv) that the batch-release discipline and `main` branch protection are unchanged by this ref's existence.
- **REQ-TBS-008** (Ubiquitous) — This SPEC's run phase shall leave every real remote, every pull-request surface, and every file under `.github/workflows/` untouched: its entire verification runs against temporary remotes (REQ-TBS-006), and the regression test of REQ-TBS-005 reads workflow files and writes nothing.
- **REQ-TBS-009** (Ubiquitous) — This SPEC shall leave branch-protection rules, guard and hook configuration, and distributed template files unchanged: `main` protection on origin (required pull request, `enforce_admins`) and every pattern of the branch guard hold their current configuration.

## §E — Constraints

- **No execution of the protected act.** The card ships mechanism + tests + instructions. The run phase never pushes to any real remote (REQ-TBS-008); the first real invocation of the verb belongs to the leader's operator-approved batch scope, after this card lands.
- **Force is scoped by construction.** Force-push is allowed on `refs/heads/sync/main` only, and only as the constant refspec of REQ-TBS-001 (REQ-TBS-002). Main protection rules are untouched (REQ-TBS-009).
- **Decision gates.** Operator decisions Q1 (verb surface), Q2 (sync ref name), Q3 (force semantics), and Q4 (leader-round integration) are recorded in `decision-index.md` with empty verdicts. REQ-TBS-001 carries provisional values while Q1–Q3 are undecided — the pattern of REQ-LMF-009 in SPEC-LOCAL-MAIN-FLOW-001. No acceptance criterion requires a real origin push, so no criterion is blocked by an undecided verdict; a verdict that selects a different value is a one-commit amendment whose footprint is enumerated here — the checklist for a Q1–Q3 flip: the ref name and verb surface in REQ-TBS-001 and REQ-TBS-002; the ref name in REQ-TBS-005's guarded-ref clause and in its mutant fixtures; the ref name and verb invocation in REQ-TBS-007's runbook contract; AC-TBS-001..005 (guarded ref, refspec, fixtures, guidance text); the mutant fixtures named in plan.md M1; and the runbook `.moai/docs/trunk-backup-runbook.md`. The SPEC-body edits on that list are applied by manager-spec through orchestrator re-delegation per the ownership matrix — manager-develop and manager-docs may not edit `spec.md` / `plan.md` / `acceptance.md` bodies.
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
