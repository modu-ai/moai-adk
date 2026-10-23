---
id: SPEC-TRUNK-BACKUP-SYNC-001
title: "Implementation plan — trunk backup sync verb + workflow-exclusion regression guard"
version: "0.3.3"
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, .github/workflows, .moai/docs"
lifecycle: spec-anchored
tags: "trunk-backup, sync-ref, git-push, workflow-triggers, regression-guard"
---

# plan.md — SPEC-TRUNK-BACKUP-SYNC-001

> Stateless on the status axis (no `status:` field). Lifecycle lives in `spec.md`.

## §A — Context

### A.1 Task shape

One small cobra verb in the existing `integration` group, one Go regression test guarding `.github/workflows` push-trigger exclusion of `sync/main`, one repository-local runbook. The regression test is the load-bearing deliverable; the verb is deliberately minimal (no flags, no config keys, no scheduler). Development mode is TDD (`quality.yaml` `constitution.development_mode: tdd`).

### A.2 Codebase anchors (measured at worktree HEAD `93f4fc75c`)

- **Verb registration site**: `internal/cli/integration.go:280` — `cmd.AddCommand(newIntegrationStatusCmd(), newIntegrationAcquireCmd(), newIntegrationReleaseCmd(), newIntegrationPolicyCmd(), newIntegrationRemeasureCmd(), newIntegrationMergeCmd(), newIntegrationPreflightCmd(), newIntegrationResyncCmd(), newIntegrationCandidateCmd())`. A `newIntegrationBackupCmd()` joins this list. No `backup` subcommand exists in the group.
- **Representative verb to copy structurally**: `internal/cli/integration_candidate.go` — `newIntegrationCandidateCmd()` at `:101`; the scrubbed runner `candidateScrubbedGit` at `:74` (env scrubbing via the `gitenv` package, `GitExitError` from `internal/factorylane`); the force-push precedent at `:333`: `git(in.CardWorktree, "push", "--force", "origin", candidateSHA+":refs/heads/"+candidateBranch)`. The backup verb's runner mirrors this shape; its refspec is a source constant, not a computed string.
- **Safety precedent for env scrubbing (card t1478)**: an inherited `GIT_DIR`/`GIT_WORK_TREE` once redirected a candidate push to a foreign repository's origin. The scrub precedes every exec; the backup verb inherits the same discipline, and AC-TBS-007 pins it as a regression test.
- **YAML parsing**: `gopkg.in/yaml.v3 v3.0.1` is a direct dependency (`go.mod:33`) — the regression test parses workflow files with the library already in the tree; no new dependency.

### A.3 Workflow push-trigger inventory (measured at `93f4fc75c`; the SPEC's §B.2 cites this table)

| Workflow file | `on.push` filter |
|---|---|
| `ci.yml` | `branches: [main, 'ci/**']` |
| `codeql.yml` | `branches: [main]` |
| `docs-i18n-check.yml` | `branches: [main]` + `paths` |
| `graph-freshness.yml` | `branches: [main]` |
| `judgment-first-consistency.yaml` | `branches: [main]` + `paths` |
| `label-sync.yml` | `branches: [main, develop]` + `paths` |
| `lsel-leak-guard.yaml` | `branches: [main]` |
| `release-drafter.yml` | `branches: [main]` |
| `release.yml` | `tags: ['v*']` (no `branches` filter) |
| `spec-lint.yml` | `branches: [main]` + `paths` |
| `template-neutrality-check.yaml` | `branches: [main]` + `paths` |
| `test-install.yml` | `branches: [main]` + `paths` |
| `workflow-parse-guard.yaml` | `branches: [main]` |

13 of 21 workflow files carry `on.push`. Every one excludes `sync/main` today; the guard test keeps it that way. The 8 files with no `push` trigger: `auto-merge.yml`, `claude.yml`, `community.yml`, `pr-multi-os-gate.yml`, `release-drafter-cleanup.yml`, `release-pr-multi-os.yml`, `review-quality-gate.yml`, `spec-status-auto-sync.yml`.

### A.4 PRESERVE list (run phase touches none of these)

- `.github/workflows/**` — read-only for the regression test; no edits (REQ-TBS-008).
- `internal/template/templates/**` — untouched (REQ-TBS-009).
- Branch-guard, hook, and branch-protection configuration (`.moai/config/sections/workflow.yaml`, `git-strategy.yaml`) — untouched.
- Other SPEC directories and `.moai/reports/**` (parallel-session artifacts).
- Runtime-managed paths (`.moai/state/**`, `.moai/harness/**`, `.moai/cache/**`).

### A.5 New-file footprint

- `internal/cli/integration_backup.go` (+ test files) — the verb.
- One regression-test file for the workflow guard (colocated in `internal/cli` or a sibling package; manager-develop picks the exact filename).
- `.github/test-input-filters.yml` — **edited, not created**: the `go_code` path filter gains `.github/workflows/**` (run-phase deliverable, D4 of plan-audit iteration 1). Without it, a diff touching only workflow files — a new file, or any of the currently unmapped 18 — classifies as not-Go and skips the Go suite, so the guard would not run exactly when its guarded surface changes (verification-completeness §1.3 continued firing). The file lives under `.github/`, NOT under `.github/workflows/`, so this edit is outside REQ-TBS-008's prohibition — that prohibition binds workflow files only. RED-now cell: acceptance.md §E E-TBS-R5 (at tree `8bff29c5f` the filter maps only the three named files at lines 54-56).
- `.moai/docs/trunk-backup-runbook.md` — the runbook.

## §B — Known Issues (filtered)

- **B1 Cross-platform**: the verb spawns `git` via `exec.Command` (precedent: `candidateScrubbedGit`); tests create bare repositories with git commands under `t.TempDir()`. Follow existing cross-platform test patterns in `internal/cli`; `GOOS=windows GOARCH=amd64 go build ./...` must pass.
- **B5 CI 3-tier awareness**: spec-lint runs on `.moai/specs/**` path filters; Go tests run in the OS matrix. The regression test reads `.github/workflows` from the module root — it must resolve the repository root robustly (`..` from the package dir) regardless of the runner's working directory.
- **B8 Working-tree hygiene**: stage by explicit pathspec only; the card id `t1832` rides every commit message on the branch (worktree traceability contract).
- **B10 Untouched paths**: §A.4 PRESERVE list is binding.
- **B11 Subagent boundary**: no `AskUserQuestion`; blockers return as structured reports to the lane.
- **Test-env isolation (memory discipline)**: temporary repositories and bare remotes live under `t.TempDir()` only — never under `$HOME` (an incident wrote the real `backlog.db`), never inside a repository checkout, and temp paths carry no card tokens.

## §C — Pre-flight (run before any code change)

```bash
git branch --show-current          # WT-10-17-class-4 (or the lane's card branch)
git rev-parse --short HEAD         # record baseline
go build ./...
GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m ./internal/cli/... 2>&1 | tail -5   # baseline before NEW issues
grep -rn "superseded\|Retired" internal/cli/integration*.go | head  # cross-SPEC conflict pre-scan
```

## §D — Constraints

- No push to any real remote, no pull request, no `.github/workflows` edit in the run phase (REQ-TBS-008). The `.github/test-input-filters.yml` edit (§A.5, D4) is not such an edit: the file lives under `.github/`, outside the prohibition's `.github/workflows/` scope.
- Refspec and force flag are source constants (REQ-TBS-002); no flag parsing, no config reads.
- Provisional values (verb surface, ref name, `--force`) hold until Q1–Q3 verdicts; a verdict flip is one commit amending the constant, the tests, and the runbook together.
- Conventional Commits with the card id: `feat(SPEC-TRUNK-BACKUP-SYNC-001): M<N> ... (card t1832)`.
- RED evidence before GREEN (TDD; E8 of the delegation template).

## §E — Self-Verification (per verification-claim-integrity 5-section format)

- E1 AC binary PASS/FAIL matrix (AC-TBS-001..010) with command + verbatim output.
- E2 Cross-platform build outputs.
- E3 Coverage for the new verb code (`go test -cover ./internal/cli/...` or the scoped run).
- E4 Subagent-boundary grep on the new files.
- E5 Lint status, NEW vs baseline split.
- E6 Commit list + branch state (no push — the branch stays local per lane protocol).
- E8 Verbatim RED outputs for each milestone's first failing run.

## §F — Milestones

Ordering by decision-reversibility: M1 lands the regression guard (the invariant that pins the ref name Q2 and the matcher semantics — the highest-change-likelihood decision surface), M2 lands the verb (the Q1/Q3 surface), M3 lands the runbook (pure documentation). A Q1/Q2/Q3 verdict flip after landing is a single follow-up commit crossing all three; the milestones are ordered so that flip rewrites constants, never structure.

- **M1 — Workflow-exclusion regression guard (REQ-TBS-005; AC-TBS-001, AC-TBS-002, AC-TBS-010) — Priority High**
  RED: write the test with its mutant fixture cases — bare `on: push`; `branches: ['sync/**']`; `branches-ignore`-only misuse; and the bidirectional set `['**']` (admits), `['*']` (does not), `['**', '!sync/main']` (does not, last-match negation), `['sync/m[a-z]in']` (admits, character class) — plus the empty-sweep guard; it fails to compile / fails until the matcher exists. GREEN: implement the branch-filter matcher with the REQ-TBS-005 grammar (`*` never crosses `/`; `**` crosses `/`; `[...]` character classes; leading-`!` negation, last match wins; fail-loud on unparseable or out-of-subset patterns — never silent non-match), the `branches` allowlist, `tags`-only, and bare-push rules, and the live sweep over `.github/workflows/*.{yml,yaml}`. M1 also lands the continued-firing deliverable: extend `.github/test-input-filters.yml`'s `go_code` filter with `.github/workflows/**` (§A.5, D4) so a workflow-only diff runs the Go suite carrying this guard. The live sweep is a regression guard by classification (the current tree passes it by design); the mutant fixtures are its observed-failure evidence (verification-completeness §1.1, §2.1).
- **M2 — The verb (REQ-TBS-001..004; AC-TBS-003..005, AC-TBS-007, AC-TBS-009) — Priority High**
  RED: verb tests against a `t.TempDir()` bare origin (first push; non-FF rewound-main force update; missing `refs/heads/main` refusal; pre-receive rejection with both-streams reporting; captured-command assertions that `--force` appears only with the constant refspec; inherited `GIT_DIR` scrub case) fail while `newIntegrationBackupCmd` does not exist. GREEN: implement the verb in `internal/cli/integration_backup.go` with the scrubbed runner and constant refspec; register at `integration.go:280`. Report pushed SHA + target ref on success.
- **M3 — Runbook + isolation audit (REQ-TBS-006, REQ-TBS-007; AC-TBS-006, AC-TBS-008) — Priority Medium**
  Author `.moai/docs/trunk-backup-runbook.md` with the four REQ-TBS-007 statements. Audit the test suite for the no-real-remote property (every remote operand resolves under `t.TempDir()`; no network URL; no reference to the development repository's remotes). Record the run-phase §E.2 evidence.

## §G — Anti-Patterns

- Adding flags, config keys, or a scheduler to the verb — each is a surface Q1–Q4 can flip; the minimal mechanism is the card's design bias.
- Editing `.github/workflows/` to make the test pass — the test guards reality; reality is not edited to satisfy the guard.
- A test that sweeps zero workflows and reports green — the empty-sweep guard exists to make that impossible.
- Computing the refspec from user input — REQ-TBS-002 requires constants.
- Pushing during the run phase "just once to verify" — REQ-TBS-008; the temp-remote tests are the verification.

## §H — Cross-References and Open Questions

- SPEC: `.moai/specs/SPEC-TRUNK-BACKUP-SYNC-001/spec.md` (§D requirements, §E constraints).
- Acceptance: `.moai/specs/SPEC-TRUNK-BACKUP-SYNC-001/acceptance.md` (AC-TBS-001..010, adoption classification + RED-now evidence ledger pinned to `8bff29c5f`).
- Open questions: all four live in `decision-index.md` (Q1–Q4) — none blocks a milestone; provisional values carry (spec §E).
- Delegation template: `.claude/rules/moai/development/manager-develop-prompt-template.md` (Tier M: full Section A–E form).
