---
title: Autonomous CI/CD Guide
weight: 10
draft: false
---

MoAI-ADK's autonomous CI/CD system manages pull request quality automatically.
It extends the "diagnose → fix → verify" loop that `/moai loop` runs in a
local session all the way into CI, so CI guarantees quality on its own without
developers verifying it manually — a case of agentic loop engineering applied
at the repository level.

## Overview

Introduced in SPEC-V3R3-CI-AUTONOMY-001, the autonomous CI/CD system is a
quality automation infrastructure composed of 8 tiers. It forms one continuous
line of defense, from local pre-push verification (pre-push hook) to automatic
fixes on CI failure (auto-fix loop).

## The 8-Tier architecture

| Tier | Name | Priority | Description |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | Automatic quality verification before push |
| T2 | Branch Protection | P0 | main branch protection rules |
| T3 | Auto-fix Loop | P1 | Automatic fixes on CI failure |
| T4 | Auxiliary Workflows | P2 | Auxiliary workflow housekeeping |
| T5 | Worktree State Guard | P1 | Guarantees worktree state integrity |
| T6 | i18n Validator | P2 | 4-locale documentation consistency checks |
| T7 | BODP | P0 | Branch Origin Decision Protocol |
| T8 | Release Workflow | P1 | Release automation |

## Pre-push Hook (T1)

Runs quality verification locally and automatically before pushing. It is the
first line of defense that cuts off, locally and in advance, the round-trip
cost of failing in CI and coming back.

```bash
# Installed automatically (on moai init / moai update)
.git/hooks/pre-push → moai hook pre-push
```

Verifications executed:

- `go vet` / `golangci-lint` (auto-detected by project language)
- `go test ./...` (the test suite)
- MX tag integrity check

## Auto-fix Loop (T3)

After `/moai sync` creates a PR, the CI watch script and the CI loop skill
together run a "diagnose → fix → re-verify" loop. It is the local diagnostic
self-fix loop extended on top of the PR pipeline.

**CI watch script (`scripts/ci-watch/run.sh`)**

```bash
sh scripts/ci-watch/run.sh <PR_NUMBER> [BRANCH]
```

- Polls `gh pr checks` at 30-second intervals, classifying required checks vs.
  auxiliary checks
- Exit codes: `0` all passed · `2` required check failed (emits a structured
  JSON handoff to stdout) · `3` 30-minute hard timeout · `1` error
- The required-check list is read from an SSoT file, and it supports
  environment-variable overrides for testing (`MOAI_CIWATCH_GH`,
  `CIWATCH_TIMEOUT_SECONDS`, etc.)

**CI loop skill (`moai-workflow-ci-loop`)**

When the watch script hands off a required failure, the `moai-workflow-ci-loop`
skill classifies the failure and attempts safe automated patches up to 3 times.
Semantic-level failures (where auto-fixing is risky) are escalated to the user.

## BODP — Branch Origin Decision Protocol (T7)

Automatically decides the base branch when creating a new branch/worktree.

### 3-Signal evaluation

| Signal | Source | Meaning |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff path overlap | Code dependency |
| Signal B | `.moai/specs/<NewSpecID>/` match in `git status` | Working tree co-location |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | Current branch PR |

### Decision matrix

| Signals | Decision |
|--------|------|
| A only | `stacked` — based on the current branch |
| B present | `continue` — continue in the current context |
| C only | `stacked` — based on the current branch |
| None | `main` — based on origin/main |

### Audit trail

Every BODP decision is recorded in
`.moai/branches/decisions/<branch-name>.md`. Leaving decisions as records
rather than guesses — the MoAI principle of evidence-based completion applies
to branch decisions as well.

## i18n Validator (T6)

Automatically verifies consistency across the 4-locale documentation.

```bash
scripts/docs-i18n-check.sh
```

Checks:

- Matching file counts/paths across the 4 locales
- Front matter `title` present
- H1 heading present
- MoAI glossary compliance

## Worktree State Guard (T5)

Guarantees the state integrity of worktrees:

- Detects uncommitted changes
- Checks sync state between the worktree and the main branch
- Shows status in `moai status`

## Related documentation

- [Worktree Guide](/en/worktree/guide) — the complete Git Worktree guide
- [/moai loop](/en/utility-commands/moai-loop) — the iterative fix loop
- [/moai fix](/en/utility-commands/moai-fix) — automatic error fixing
- [GitHub Integration Guide](/en/guides/multi-llm-ci) — issue parsing · SPEC linking
