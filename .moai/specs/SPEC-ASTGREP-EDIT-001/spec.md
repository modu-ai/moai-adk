---
id: SPEC-ASTGREP-EDIT-001
title: "ast-grep wiring repair and moai ast-edit command"
version: "0.1.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: GOOS
priority: P1
phase: "v3.1.0 target"
module: "internal/cli, internal/astgrep, internal/hook/security"
lifecycle: spec-anchored
tags: "ast-grep, ast-edit, cli, security-hook, rules, autofix"
---

# SPEC-ASTGREP-EDIT-001 — ast-grep wiring repair and `moai ast-edit`

## A. Context

MoAI-ADK already ships a working ast-grep integration, but three of its wires are
cut and its write-side capability is implemented yet unreachable. This SPEC
repairs the wiring and exposes the existing rewrite engine as a distinct
`moai ast-edit` command.

### A.1 Verified baseline (measured against `origin/main`, not the local checkout)

| Asset | Verified state |
|---|---|
| `internal/astgrep/` | Present — `analyzer.go`, `scanner.go`, `rules.go`, `sarif.go`, `models.go` |
| `moai ast-grep` CLI (`internal/cli/astgrep.go`, registered `root.go:152`) | Working — text / json / sarif output; flags `--dry`, `--format`, `--lang`, `--rules-dir`, `--severity` |
| `SGAnalyzer.Replace` (`analyzer.go:287`) | Implemented, unit-tested, **zero CLI callers** |
| `SGAnalyzer.PatternReplace` (`analyzer.go:475`) | Implemented with `dryRun` parameter, unit-tested, **zero CLI callers** |
| `ReplaceResult` (`models.go:64`) | `MatchesFound`, `FilesModified`, `Changes []FileChange`, `DryRun` |
| Shipped ruleset | `sgconfig.yml` + `go/hardcoding.yml` + `security/{credentials,crypto,injection}.yml` |
| `fix:` fields in shipped rules | **0** — nothing for a rewrite pass to apply |
| `gate.yaml` `ast_grep_gate.enabled` | `false` (opt-in; `rules_dir` defaults to `.moai/config/astgrep-rules`) |
| `FindRulesConfig` (`internal/hook/security/rules.go`) | Searches 6 paths, **none is the shipped location** |
| `moai-tool-ast-grep` skill | Removed (guarded by `internal/template/skills_removal_test.go`) |

Evidence for the `FindRulesConfig` gap, run in a real MoAI project tree:

```
sgconfig.yml                                         absent
sgconfig.yaml                                        absent
.ast-grep/sgconfig.yml                               absent
.ast-grep/sgconfig.yaml                              absent
.claude/skills/moai-tool-ast-grep/rules/sgconfig.yml absent   <- removed skill
.moai/config/astgrep-rules/sgconfig.yml              EXISTS    <- not in search list
```

The function therefore returns the empty string for every default MoAI project,
and the security hook's rule load silently finds nothing. This is a functional
defect, not a documentation defect.

### A.2 Scope boundary

This SPEC does **not** remove the DB documentation subsystem. That removal
landed separately and is already merged (`origin/main` carries no `db.yaml`, no
`.moai/project/db/`, no `internal/hook/dbsync/`, no `db-schema-sync`
subcommand). Any residual DB references seen in a stale local checkout are
out of scope here.

## B. Requirements

### REQ-AGE-001 — `moai ast-edit` command exists as a distinct write-side entry point

**Where** the `sg` binary is available, **when** a user invokes
`moai ast-edit`, the CLI SHALL perform an ast-grep rewrite pass and report the
resulting `ReplaceResult` (matches found, files modified, per-file changes).

The command SHALL be registered separately from `moai ast-grep` so that a
permission rule granting `Bash(moai ast-grep:*)` does not thereby grant write
capability.

### REQ-AGE-002 — Dry-run is the safe default surface

**When** `moai ast-edit` runs with `--dry`, the command SHALL compute and print
the changes **without modifying any file**, delegating to the existing
`PatternReplace(..., dryRun=true)` path.

The command SHALL make the destructive nature of a non-`--dry` run explicit in
its help text.

### REQ-AGE-003 — Pattern-mode rewrite

**When** `moai ast-edit --pattern <p> --rewrite <r>` is invoked with a target
path, the command SHALL apply the pattern rewrite via
`SGAnalyzer.PatternReplace`, honouring `--lang` for language selection.

### REQ-AGE-004 — Rule-mode rewrite

**Where** a loaded rule declares a `fix:` field, **when** `moai ast-edit` runs
without an explicit `--pattern`, the command SHALL apply that rule's `fix:`
rewrite. `--rule <id>` SHALL restrict the pass to a single rule.

### REQ-AGE-005 — `FindRulesConfig` resolves the shipped ruleset

**When** `FindRulesConfig` searches for an ast-grep configuration, the search
list SHALL include `.moai/config/astgrep-rules/sgconfig.yml` (and its `.yaml`
sibling), so that a default MoAI project resolves the ruleset it actually ships.

The retired `moai-tool-ast-grep` skill paths SHALL be removed from the search
list, since that skill no longer exists.

### REQ-AGE-006 — Shipped rules carry safe `fix:` fields

**Where** an automatic rewrite is unambiguous and behaviour-preserving, the
shipped rule SHALL declare a `fix:` field. **Where** the rewrite is ambiguous or
security-sensitive, the rule SHALL remain detection-only, and the SPEC SHALL
record why.

The four shipped rules are assessed individually; blanket `fix:` addition is
prohibited.

### REQ-AGE-007 — Dead skill references removed

**When** a shipped skill body references the removed `moai-tool-ast-grep`
skill, that reference SHALL be replaced with the surviving surface
(`moai ast-grep` / `moai ast-edit` or the `.moai/config/astgrep-rules` path).

Affected: `moai-foundation-cc/SKILL.md`, `moai-workflow-ddd/SKILL.md`,
`moai-workflow-loop/SKILL.md` — plus their template mirrors.

### REQ-AGE-008 — Template-First parity and neutrality

**When** any file under `.claude/` or `.moai/` is changed, the corresponding
`internal/template/templates/` mirror SHALL be updated byte-identically and
`make build` SHALL be run, and the template neutrality guards SHALL pass.

## C. Scope Boundary

### C.1 Out of Scope

- Removing the DB documentation subsystem (already merged separately).
- Enabling `ast_grep_gate` by default — it stays opt-in (`enabled: false`).
- Promoting the local dogfood multi-language ruleset to the shipped set. The
  local tree's language directories are experimental (empty stubs, mixed-language
  messages); their promotion is a separate SPEC.
- Retiring the documentation-only `ralph.yaml` `ast_grep:` block. It is
  deliberately retained per the recorded settings-schema decision.

## D. Risks

| Risk | Mitigation |
|---|---|
| A `fix:` rewrite silently changes behaviour | Each `fix:` requires a positive and a negative fixture; ambiguous rules stay detection-only |
| `ast-edit` run without `--dry` damages a working tree | Destructive path documented in help text; dry-run path exercised by tests |
| Adding a search path changes existing hook behaviour | The new path is appended, not substituted, for the non-retired entries; retired skill paths are removed because the skill provably no longer exists |
| Local checkout is behind `origin/main` | Implementation rebases onto `origin/main` before starting; the baseline in §A.1 was measured against `origin/main` |
