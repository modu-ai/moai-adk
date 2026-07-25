---
id: SPEC-ASTGREP-EDIT-001
title: "Acceptance criteria — ast-grep wiring repair and moai ast-edit"
version: "0.1.0"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: GOOS
priority: P1
phase: "v3.1.0 target"
module: "internal/cli, internal/astgrep, internal/hook/security"
lifecycle: spec-anchored
tags: "ast-grep, ast-edit, acceptance"
---

# Acceptance Criteria — SPEC-ASTGREP-EDIT-001

Each criterion states a command and the observation that constitutes a PASS.
A criterion whose check would pass even if the implementation were reverted is
not a valid criterion; every AC below was chosen so that removing its change
makes it fail.

## REQ-AGE-005 — `FindRulesConfig` resolves the shipped ruleset

**AC-001** — In a tree whose only ast-grep config is
`.moai/config/astgrep-rules/sgconfig.yml`, `FindRulesConfig` returns that path.

```bash
go test ./internal/hook/security/ -run TestFindRulesConfig -v
```
PASS: the shipped-location case is present and passes. Falsification check:
reverting the `searchPaths` change makes this test fail.

**AC-002** — The retired skill paths are gone from the resolver.

```bash
grep -c "moai-tool-ast-grep" internal/hook/security/rules.go
```
PASS: `0`.

**AC-003** — The doc comment's numbered search order matches the code list.

```bash
sed -n '/Search order:/,/^func /p' internal/hook/security/rules.go
```
PASS: each numbered comment line corresponds 1:1, in order, to a `searchPaths`
entry.

## REQ-AGE-001 / 002 — command exists, dry-run is safe

**AC-010** — `moai ast-edit` is registered and self-documents its write nature.

```bash
moai ast-edit --help
```
PASS: help renders; the text states that a non-`--dry` run modifies files.

**AC-011** — `ast-edit` is a separate command from `ast-grep`.

```bash
moai --help | grep -cE '^\s+(ast-grep|ast-edit)'
```
PASS: `2`.

**AC-012** — A dry run modifies nothing.

```bash
git stash list > /tmp/pre.txt; git diff --stat > /tmp/pre-diff.txt
moai ast-edit --dry --pattern '<p>' --rewrite '<r>' ./internal/config
git diff --stat > /tmp/post-diff.txt; diff /tmp/pre-diff.txt /tmp/post-diff.txt
```
PASS: `diff` is empty (no working-tree change), and the command's output reports
`dry_run: true`.

## REQ-AGE-003 — pattern-mode rewrite

**AC-020** — Pattern mode reports matches and applies the rewrite.

```bash
go test ./internal/cli/ -run TestAstEdit_PatternMode -v
```
PASS: a fixture file is rewritten as expected in the non-dry case, and left
untouched in the dry case.

**AC-021** — `--pattern` without `--rewrite` is rejected.

```bash
moai ast-edit --pattern 'x' ./internal/config; echo "exit=$?"
```
PASS: non-zero exit with a message naming the missing flag.

## REQ-AGE-004 — rule-mode rewrite

**AC-030** — Rules carrying `fix:` are applied; rules without are skipped and counted.

```bash
go test ./internal/cli/ -run TestAstEdit_RuleMode -v
```
PASS: the fixture with a `fix:` rule is rewritten; a detection-only rule produces
a skip notice, not an error.

**AC-031** — `--rule <id>` narrows the pass to one rule.

```bash
go test ./internal/cli/ -run TestAstEdit_RuleFilter -v
```
PASS: only the named rule's changes appear in `ReplaceResult.Changes`.

## REQ-AGE-006 — shipped rule `fix:` assessment

**AC-040** — Every shipped rule that declares `fix:` has both a positive and a
negative fixture.

```bash
go test ./internal/astgrep/ -run 'TestRuleFixtures' -v
```
PASS: for each `fix:`-bearing rule, the positive fixture is rewritten and the
negative fixture yields zero findings.

**AC-041** — Rules deliberately left detection-only are recorded with a reason.

```bash
grep -n "detection-only" .moai/specs/SPEC-ASTGREP-EDIT-001/acceptance.md
```
PASS: the disposition table below is filled in during the run with one row per
shipped rule, each carrying an evidence-backed verdict.

### Disposition table (filled during run phase — 2026-07-25)

Outcome: **all four shipped rule files remain detection-only; zero `fix:` fields
were added.** REQ-AGE-006 conditions a `fix:` on the rewrite being unambiguous
and behaviour-preserving; none of the shipped rules meets that bar. This is a
verdict, not a deferral — each row states why the rewrite cannot be mechanical.

| Rule (id) | Verdict | Evidence |
|---|---|---|
| `go-no-hardcoded-api-url` | detection-only | The fix is "extract to a constant". A rewrite must invent a constant name and add a declaration elsewhere in the file — not expressible as an AST pattern rewrite of the literal. |
| `go-no-duplicate-coverage-threshold` | detection-only | The fix references a shared threshold constant whose name and import path are project-specific. No single rewrite is correct across projects. |
| `sec-hardcoded-credential` (go/python/javascript/typescript) | detection-only | Rewriting the literal to an env lookup requires inventing a variable name, and — more importantly — it removes the visible marker while the secret remains in git history. The rewrite would hide the exposure rather than fix it. |
| `sec-weak-hash-md5` | detection-only | `md5.New()` -> `sha256.New()` is pattern-expressible, but the call-site rewrite alone leaves the `crypto/md5` import in place and the `crypto/sha256` import absent, producing code that does not compile. It also changes digest length, breaking any persisted hashes. |
| `sec-command-injection-shell` / `sec-command-injection-exec` | detection-only | The fix is to drop the shell and validate input against an allowlist — a semantic change requiring call-site context an AST pattern does not carry. |

Because no `fix:` was added, AC-040's fixture requirement is vacuously satisfied
(zero `fix:`-bearing shipped rules, so zero fixtures required). Rule mode is
exercised instead against flat-form fixture rules in
`TestAstEditCmd_RuleModeAppliesFix` / `TestAstEditCmd_RuleFilterNarrowsToOneRule`.

### Known gap — nested `rule:` block form (recorded, not fixed)

The shipped rules express their matcher as a nested `rule:` block
(`rule: {kind, regex}` / `rule: {pattern}` / `rule: {any: [...]}`), but
`astgrep.Rule` reads only the flat top-level `pattern:` field. Measured: all 11
shipped rules load with `Pattern == ""` and `Fix == ""`.

Consequence for this SPEC: rule mode skips them with a counted
detection-only notice rather than passing an empty pattern to `sg`. That is the
correct conservative behaviour and is guarded in `applyRuleEdits`.

Consequence beyond this SPEC: a user-authored rule that uses the nested form
**and** declares a `fix:` would also be skipped. Teaching the loader the nested
form affects the scanner path as well and is therefore out of scope here.

## REQ-AGE-007 — dead skill references removed

**AC-050** — No shipped skill references the removed skill.

```bash
git ls-files '.claude/skills/*' 'internal/template/templates/.claude/skills/*' \
  | xargs grep -l "moai-tool-ast-grep"
```
PASS: no output.

**AC-051** — The `related-skills` frontmatter no longer names it.

```bash
grep -n "related-skills" .claude/skills/moai-workflow-ddd/SKILL.md
```
PASS: the value omits `moai-tool-ast-grep`.

## REQ-AGE-008 — template parity and neutrality

**AC-060** — Every changed `.claude/` or `.moai/` file is byte-identical to its
template mirror.

```bash
for p in <changed paths>; do diff -q "$p" "internal/template/templates/$p"; done
```
PASS: no differences reported.

**AC-061** — Build and guards are clean.

```bash
make build
go test ./internal/template/ -run 'Neutrality|InternalContent|Leak|Namespace'
```
PASS: `make build` exits 0; all guard tests PASS.

**AC-062** — Whole-repo health is unchanged or better.

```bash
go build ./... && go vet ./... && go test ./... && golangci-lint run --timeout=5m
```
PASS: all four exit 0; `go test` reports 0 FAIL.

## Regression guard

**AC-070** — The pre-existing `moai ast-grep` read path is unaffected.

```bash
moai ast-grep --format=json ./internal/config | head -c 200
```
PASS: findings are still produced with the same schema as before the change.
