---
id: SPEC-ASTGREP-EDIT-001
title: "Acceptance criteria — ast-grep wiring repair and moai ast-edit"
version: "0.3.0"
status: in-progress
created: 2026-07-25
updated: 2026-07-26
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

### Convention — `go test -run` criteria state a PASS line, never a zero exit

`go test -run <selector>` exits 0 when the selector matches **zero** tests; it
reports `[no tests to run]` and is indistinguishable from success by exit code
alone. Every AC below whose command is a `go test -run ...` invocation therefore
states its PASS observation as **the presence of a `--- PASS: <exact test name>`
line in the output**, and carries `-count=1` to defeat result caching. A run
reporting `[no tests to run]` is a **FAIL**, not a PASS.

Note that a `-run` selector is an unanchored regex matched per path segment, so a
name that is a prefix-in-spirit but not a substring (`TestFindRulesConfig` vs the
real `TestRuleManager_FindRulesConfig`) matches nothing and silently exits 0.

## REQ-AGE-005 — `FindRulesConfig` resolves the shipped ruleset

**AC-001** — In a tree whose only ast-grep config is
`.moai/config/astgrep-rules/sgconfig.yml`, `FindRulesConfig` returns that path.

```bash
go test ./internal/hook/security/ -run TestRuleManager_FindRulesConfig -v -count=1
```
PASS: output contains `--- PASS: TestRuleManager_FindRulesConfig`, covering the
shipped-location case. `[no tests to run]` is a FAIL (see Convention above).
Falsification check: reverting the `searchPaths` change makes this test fail.

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
go test ./internal/cli/ \
  -run 'TestAstEditCmd_PatternModeRewritesFile|TestAstEditCmd_DryRunLeavesFileUnchanged' \
  -v -count=1
```
PASS: output contains BOTH `--- PASS: TestAstEditCmd_PatternModeRewritesFile`
(fixture rewritten in the non-dry case) and
`--- PASS: TestAstEditCmd_DryRunLeavesFileUnchanged` (fixture untouched in the
dry case). `[no tests to run]`, or either line missing, is a FAIL.

**AC-021** — `--pattern` without `--rewrite` is rejected.

```bash
moai ast-edit --pattern 'x' ./internal/config; echo "exit=$?"
```
PASS: non-zero exit with a message naming the missing flag.

## REQ-AGE-004 — rule-mode rewrite

**AC-030** — Rules carrying `fix:` are applied; rules without are skipped and counted.

```bash
go test ./internal/cli/ -run TestAstEditCmd_RuleModeAppliesFix -v -count=1
```
PASS: output contains `--- PASS: TestAstEditCmd_RuleModeAppliesFix` — the fixture
with a `fix:` rule is rewritten and a detection-only rule produces a skip notice,
not an error. `[no tests to run]` is a FAIL.

**AC-031** — `--rule <id>` narrows the pass to one rule.

```bash
go test ./internal/cli/ -run TestAstEditCmd_RuleFilterNarrowsToOneRule -v -count=1
```
PASS: output contains `--- PASS: TestAstEditCmd_RuleFilterNarrowsToOneRule` —
only the named rule's changes appear in `ReplaceResult.Changes`.
`[no tests to run]` is a FAIL.

## REQ-AGE-006 — shipped rule `fix:` assessment

**AC-040** — Every shipped rule that declares `fix:` has both a positive and a
negative fixture.

The requirement is satisfied **vacuously**: no shipped rule declares a `fix:`
(see the disposition table below), so zero fixtures are required. The criterion
therefore observes that vacuity premise directly rather than asserting a fixture
suite. No fixture test named by this AC exists, and none is required while the
count stays 0 — this criterion deliberately names no test, because inventing a
`go test -run` selector for a test that does not exist is the F0 defect itself.

A zero-count assertion carries its own vacuity hazard, mirroring the `-run` case
in the Convention above: a `grep ... | wc -l` aimed at a path that does not exist
also reports `0` and would certify a deleted ruleset as clean. Command (a)
therefore pins the subject set to a non-zero size before command (b) asserts the
count over it. Both must hold.

```bash
# (a) subject guard — the shipped rule-file set is non-empty
find internal/template/templates/.moai/config/astgrep-rules \
  -name '*.yml' ! -name 'sgconfig.yml' | wc -l
# (b) no shipped rule declares a fix: field, at any indentation
grep -rnE '^[[:space:]]*fix:' \
  internal/template/templates/.moai/config/astgrep-rules/ | wc -l
```
PASS: (a) is `4` — the shipped rule files `go/hardcoding.yml` plus
`security/{credentials,crypto,injection}.yml`, with the `sgconfig.yml` config
excluded because it declares no rules — AND (b) is `0`, so the fixture obligation
has no subject. A count other than `4` from (a) means the ruleset moved or shrank
and the criterion is measuring the wrong thing (FAIL, not PASS). Any non-zero
count from (b) means a `fix:` was added, at which point this AC no longer holds
vacuously and a positive/negative fixture pair becomes mandatory for each
`fix:`-bearing rule (and this criterion must be rewritten to assert that suite).

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

**AC-060** — Every change this branch makes to a `.claude/` or `.moai/` file is
applied byte-identically to that file's template mirror.

Two corrections were applied here, and the second changes what the criterion
asserts — recorded explicitly rather than folded in silently.

*Correction 1 (runnability).* The committed form read
`for p in <changed paths>; do ...` — an unfilled placeholder, so it was never
runnable and therefore never observable. That is the same class of defect the
Convention above names: a check that cannot fail is not a check. The path list is
now derived mechanically from the branch diff so it cannot drift from the actual
changed set, and an empty list is an explicit FAIL — a loop that iterates zero
times reports no differences and would otherwise certify a missing mirror as
clean.

*Correction 2 (delta, not whole-file).* The committed wording asserted whole-file
byte-identity between each changed path and its mirror. REQ-AGE-008 states that
"the corresponding mirror SHALL be updated byte-identically", which reads
naturally as *the update* being mirrored byte-identically, and the whole-file
reading is not merely stricter — it is **unsatisfiable without violating a
different HARD rule**. `.claude/skills/moai-foundation-cc/SKILL.md` carries a
14-line local-only block (a `## Decision Heuristics` section plus a Provenance
line naming an internal constitution token and an internal memory reference)
that is absent from its mirror and is present on `origin/main`, so this branch
neither introduced nor widened it. Satisfying the whole-file reading would mean
copying that internal-provenance text into `internal/template/templates/`, which
the template internal-content isolation rule forbids. The criterion therefore
asserts delta identity, which is both what REQ-AGE-008 says and the only reading
compatible with template neutrality. The pre-existing whole-file divergence is
recorded as named debt in `progress.md` § Residual findings — it is out of this
SPEC's scope, not resolved by this rewording.

`.moai/specs/` is excluded because SPEC artifacts are project-local and have no
template mirror by design; the `-- '.claude' '.moai'` pathspec already excludes
the `internal/template/templates/` side.

```bash
paths=$(git diff --name-only origin/main...HEAD -- '.claude' '.moai' \
          | grep -v '^\.moai/specs/')
test -n "$paths" || { echo "FAIL: empty path list"; exit 1; }
printf '%s\n' "$paths" | while read -r p; do
  a=$(git diff origin/main...HEAD -- "$p" \
        | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)')
  b=$(git diff origin/main...HEAD -- "internal/template/templates/$p" \
        | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)')
  if [ -z "$a" ]; then echo "FAIL: no delta for $p"
  elif [ "$a" = "$b" ]; then echo "DELTA-MIRRORED  $p"
  else echo "DELTA-DIVERGED  $p"; fi
done
```
PASS: the path list is non-empty AND every line reads `DELTA-MIRRORED`. Any
`DELTA-DIVERGED` line, any `FAIL:` line, or the `empty path list` message is a
FAIL. Falsification check: mirroring a `.claude/` edit incompletely — or not at
all — yields `DELTA-DIVERGED` for that path.

**AC-061** — Build and guards are clean.

```bash
make build
go test ./internal/template/ -run 'Neutrality|InternalContent|Leak|Namespace' -v -count=1
```
PASS: `make build` exits 0; the `go test` output contains at least one
`--- PASS: Test...` line and zero `--- FAIL:` lines. `[no tests to run]` is a
FAIL — the selector matching zero guard tests would otherwise exit 0 and certify
a removed guard as passing.

**AC-062** — Whole-repo health is unchanged or better.

```bash
go build ./... && go vet ./... && go test ./... && golangci-lint run --timeout=5m
```
PASS: all four exit 0; `go test` reports 0 FAIL. The zero-test hazard in the
Convention above does not apply here — this invocation carries no `-run`
selector, so the whole suite runs and an empty selection is impossible.

## Regression guard

**AC-070** — The pre-existing `moai ast-grep` read path is unaffected.

```bash
moai ast-grep --format=json ./internal/config | head -c 200
```
PASS: findings are still produced with the same schema as before the change.
