---
id: SPEC-ASTGREP-EDIT-001
title: "Implementation plan — ast-grep wiring repair and moai ast-edit"
version: "0.1.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: GOOS
priority: P1
phase: "v3.1.0 target"
module: "internal/cli, internal/astgrep, internal/hook/security"
lifecycle: spec-anchored
tags: "ast-grep, ast-edit, cli, plan"
---

# Implementation Plan — SPEC-ASTGREP-EDIT-001

## A. Tier and cycle

- **Tier**: L — a new user-facing CLI command that **writes files**, plus a
  behavioural change to a security-hook resolver. The write surface is what
  raises this above Tier M.
- **cycle_type**: `tdd` — every requirement has a mechanically checkable
  assertion; the rewrite engine already has tests to extend.
- **Harness level**: thorough.

## B. Preconditions (verify before the first edit)

1. `git rev-list --count --left-right origin/main...HEAD` — rebase onto
   `origin/main` first. The current checkout is behind, and the §A.1 baseline was
   measured against `origin/main`.
2. Confirm the moot DB-related edits from the pre-SPEC session are dropped
   (`db.yaml`, `.moai/project/db/*` no longer exist upstream).
3. `sg --version` resolves — the rewrite path is skipped gracefully when absent,
   but the tests need it for the non-skipped assertions.

## C. Milestones

### M1 — `FindRulesConfig` resolves the shipped ruleset (REQ-AGE-005)

Files: `internal/hook/security/rules.go`, `internal/hook/security/rules_test.go`.

1. RED: add a test asserting that a tree containing only
   `.moai/config/astgrep-rules/sgconfig.yml` resolves to that path. It fails today.
2. GREEN: add `.moai/config/astgrep-rules/sgconfig.{yml,yaml}` to `searchPaths`;
   remove the two retired `moai-tool-ast-grep` entries.
3. Update the doc comment's numbered search order to match the new list.

Ordering note: M1 is first because it is the only change that repairs an
observable functional defect, and it is independent of the new command.

### M2 — `moai ast-edit` command skeleton (REQ-AGE-001, REQ-AGE-002)

Files: new `internal/cli/astedit.go`, `internal/cli/root.go`, new
`internal/cli/astedit_test.go`.

1. RED: test that the command is registered, that `--dry` is accepted, and that a
   dry run reports `DryRun: true` with `FilesModified == 0`.
2. GREEN: register alongside the existing `ast-grep` registration
   (`root.go:152`); reuse the `--lang` / `--rules-dir` flag definitions so the two
   commands stay consistent.
3. Help text states plainly that a non-`--dry` run modifies files in place.

### M3 — Pattern-mode rewrite (REQ-AGE-003)

Files: `internal/cli/astedit.go`, tests.

Wire `--pattern` / `--rewrite` to `SGAnalyzer.PatternReplace(ctx, pattern,
replacement, lang, path, dryRun)`. Render `ReplaceResult` in text and json.
Reject invocation when `--pattern` is given without `--rewrite`.

### M4 — Rule-mode rewrite (REQ-AGE-004)

Files: `internal/cli/astedit.go`, `internal/astgrep/rules.go` if the loader needs
to surface `Fix`, tests.

Load rules via the existing loader, select those with a non-empty `Fix`, and
apply. `--rule <id>` narrows to one rule. A rule without `fix:` is skipped with a
counted notice rather than an error.

### M5 — Shipped rule `fix:` assessment (REQ-AGE-006)

Files: `internal/template/templates/.moai/config/astgrep-rules/go/hardcoding.yml`,
`security/{credentials,crypto,injection}.yml`, plus local mirrors and fixtures.

Assess one rule at a time and record the verdict in `acceptance.md`:

| Rule | Expected disposition | Reasoning to verify during run |
|---|---|---|
| `go/hardcoding.yml` | `fix:` candidate | Mechanical extraction to a constant may still be ambiguous; verify per pattern |
| `security/credentials.yml` | likely detection-only | An automatic rewrite of a credential site can mask, not fix, the exposure |
| `security/crypto.yml` | likely detection-only | Algorithm substitution is a semantic decision |
| `security/injection.yml` | likely detection-only | Requires call-site context an AST pattern does not carry |

The table records *expectations to be tested*, not conclusions. A rule gains
`fix:` only after its positive and negative fixtures pass.

### M6 — Dead skill reference cleanup (REQ-AGE-007)

Files: `moai-foundation-cc/SKILL.md`, `moai-workflow-ddd/SKILL.md`,
`moai-workflow-loop/SKILL.md` and the three template mirrors.

Replace `moai-tool-ast-grep` mentions with the surviving surface. In
`moai-workflow-ddd/SKILL.md` this includes the `related-skills` frontmatter value.

### M7 — Template parity, docs, verification (REQ-AGE-008)

1. `make build`; confirm `catalog.yaml` regeneration is limited to changed
   SKILL.md hashes.
2. Template neutrality and namespace guards.
3. Document `moai ast-edit` in the docs-site across all four locales, and in the
   README set if the command list is enumerated there.

## D. Verification batch (single-turn, read-only)

```bash
go build ./...
go vet ./...
go test ./... 
golangci-lint run --timeout=5m
go test ./internal/template/ -run 'Neutrality|InternalContent|Leak|Namespace'
moai ast-grep --dry ./internal/config
moai ast-edit --dry --pattern '<p>' --rewrite '<r>' ./internal/config
```

## E. Rollback

Each milestone is an independent commit. M1 is separately revertible from the
new command. The `fix:` additions in M5 are per-rule commits so a single
problematic rule can be reverted without losing the rest.
