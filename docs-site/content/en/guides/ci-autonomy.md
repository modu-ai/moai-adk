---
title: Autonomous CI/CD Guide
weight: 10
draft: false
---

MoAI-ADK's autonomous CI/CD system automatically manages pull-request quality.

## Overview

The autonomous CI/CD system introduced in SPEC-V3R3-CI-AUTONOMY-001 is a quality
automation infrastructure composed of 8 tiers. From the pre-push hook through the
auto-fix loop, CI guarantees quality automatically — developers never need to verify
quality by hand.

## 8-Tier Architecture

| Tier | Name | Priority | Description |
|------|------|----------|------|
| T1 | Pre-push Hook | P0 | Automated quality validation before push |
| T2 | Branch Protection | P0 | Protection rules for the main branch |
| T3 | Auto-fix Loop | P1 | Automatic fixes when CI fails |
| T4 | Auxiliary Workflows | P2 | Cleanup of auxiliary workflows |
| T5 | Worktree State Guard | P1 | Guarantees worktree state integrity |
| T6 | i18n Validator | P2 | Validates 4-locale documentation consistency |
| T7 | BODP | P0 | Branch Origin Decision Protocol |
| T8 | Release Workflow | P1 | Release automation |

## Pre-push Hook (T1)

Runs automated quality validation locally before every push.

```bash
# Installed automatically (during moai init / moai update)
.git/hooks/pre-push → moai hook pre-push
```

Validations performed:

- `go vet` / `golangci-lint` (auto-detected based on the project's language)
- `go test ./...` (test suite)
- MX tag integrity check

## Auto-fix Loop (T3)

Automatically invokes `/moai loop` to fix errors when CI fails.

```yaml
# .github/workflows/ci.yml (auto-generated)
- name: Auto-fix on failure
  if: failure()
  run: |
    claude -p "/moai loop --max-iterations 3"
```

## BODP — Branch Origin Decision Protocol (T7)

Automatically decides the base branch when creating a new branch or worktree.

### 3-Signal Evaluation

| Signal | Source | Meaning |
|--------|------|------|
| Signal A | SPEC `depends_on` + diff path overlap | Code dependency |
| Signal B | A `.moai/specs/<NewSpecID>/` match in `git status` | Same working-tree location |
| Signal C | `gh pr list --head <branch> --state open` ≥ 1 | An open PR on the current branch |

### Decision Matrix

| Signal | Decision |
|--------|------|
| A only | `stacked` — based on the current branch |
| B present | `continue` — continue in the current context |
| C only | `stacked` — based on the current branch |
| None present | `main` — based on origin/main |

### Audit Trail

Every BODP decision is recorded in `.moai/branches/decisions/<branch-name>.md`.

## i18n Validator (T6)

Automatically validates consistency across the 4-locale documentation.

```bash
scripts/docs-i18n-check.sh
```

Items validated:

- File count/path parity across the 4 locales
- Presence of a front matter `title`
- Presence of an H1 heading
- Compliance with the MoAI glossary

## Worktree State Guard (T5)

Guarantees the integrity of a worktree's state:

- Detects uncommitted changes
- Checks whether the worktree is in sync with the main branch
- Surfaces the state in `moai status`

## Related Documentation

- [Worktree Guide](/worktree/guide) — Complete Git Worktree guide
- [/moai loop](/utility-commands/moai-loop) — The iterative fix loop
- [/moai fix](/utility-commands/moai-fix) — Automatic error fixing
- [Multi-LLM CI](/guides/multi-llm-ci) — Multi-LLM CI integration
