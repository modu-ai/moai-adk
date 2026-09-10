# SPEC-WORKTREE-BASEREF-001 — Acceptance Criteria

Card: t313 · Tier M · Baseline tree: worktree `.claude/worktrees/t313`, branch `WT-worktree-baseref`, HEAD `48eb945df`.
Evidence for run-phase: `.moai/reports/t313/` (verdict) and `.moai/state/verify/t313/` (verbatim command output).

Every criterion below names a command and an outcome someone can run. Run all Go commands scoped to the affected packages (`go test ./internal/<pkg>/...`), never `go test ./...` locally (CLAUDE.local.md §6).

---

## §D AC Matrix

| AC | Severity | REQ | Given / When / Then |
|---|---|---|---|
| AC-WBR-001 | MUST | 001, 002 | Schema and neutral default |
| AC-WBR-002 | MUST | 003 | Template carries no repository-specific branch |
| AC-WBR-003 | MUST | 005 | Unset setting → byte-identical behavior |
| AC-WBR-004 | MUST | 006 | Matching setting → no write, no output |
| AC-WBR-005 | MUST | 007 | Mismatching setting → write plus exactly one notice line |
| AC-WBR-006 | MUST | 008 | Hook failure does not block session start |
| AC-WBR-007 | MUST | 010 | `moai cc -w` cuts the tree from the configured base |
| AC-WBR-008 | MUST | 011 | Empty/unresolvable → today's invocation, byte-identical |
| AC-WBR-009 | MUST | 012 | Doctor reports empty, match, mismatch, and unresolvable — four distinct states |
| AC-WBR-010 | MUST | 013 | Field present in `settings.AllFields()` and in rendered HTML |
| AC-WBR-011 | MUST | 014 | The control is `TypeText`, renders a text input, and accepts any branch |
| AC-WBR-012 | MUST | 015 | The key reaches a consumer, proven by a test |
| AC-WBR-013 | MUST | 016 | Template-first parity (`make build`, mirror, `.sh`/`.sh.tmpl` twins) |
| AC-WBR-014 | MUST | 013 | Typed round trip does not drop unmodelled `manual.*` keys |
| AC-WBR-015 | MUST | 009 | Unresolvable configured value → no set-head, one diagnostic line, session proceeds |
| AC-WBR-016 | MUST | 004 | Firing point: exactly once from the primary checkout, never from a linked worktree |

---

## §D.1 Scenarios

Consumer-1 scenarios (AC-WBR-003, -004, -005, -015) are all stated **from the primary checkout** — that is their implicit `Given`, because REQ-WBR-004's second clause makes the primary checkout the only place the alignment step fires. AC-WBR-016 half 2 is the paired negative for a linked worktree. Consumer-2 scenarios (AC-WBR-007, -008) carry no such precondition and hold from any working tree.

### AC-WBR-001 — Schema and neutral default (MUST)

**Given** a project whose `git-strategy.yaml` omits `worktree_base_branch`
**When** the config is loaded
**Then** `GitStrategyConfig.WorktreeBaseBranch` is the empty string and no error is raised.

```bash
go test ./internal/config/... -run 'GitStrategy' -count=1
grep -n 'worktree_base_branch' internal/config/types.go   # expect exactly 1 hit
```

### AC-WBR-002 — Template neutrality (MUST)

**Given** the shipped template
**When** the template file is grepped for the key's line
**Then** the value is empty and names no branch.

```bash
grep -n 'worktree_base_branch' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
# expect a line whose value is "" — and, binding the VALUE side of the colon only:
grep 'worktree_base_branch' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl \
  | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main
# expect exit code 1 (no match) — a value-bearing hit FAILS this AC
```

The prohibition binds the **value**, not the comment: `sed 's/#.*//'` strips any trailing comment and `cut -d: -f2-` keeps only the value side, so a house-style comment naming the two common branches (`# e.g. main, develop; empty = no action`) is permitted and does not FAIL this criterion. plan.md §C M1 is written to match — the shipped **value** must be empty, the comment may name branches as examples.

### AC-WBR-003 — Unset setting is byte-identical to today (MUST)

**Given** an empty or absent `worktree_base_branch`
**When** SessionStart runs
**Then** no `git remote set-head` process is spawned, nothing is written to stderr by the alignment step, and `refs/remotes/origin/HEAD` is unchanged.

```bash
go test ./internal/hook/... -run 'WorktreeBaseBranch.*(Unset|Empty)' -count=1 -v
```

The test asserts on a fake git seam (spawn count 0) and on captured stderr (empty), not on the developer's real repository.

### AC-WBR-004 — Matching setting produces no write and no output (MUST)

**Given** `worktree_base_branch: develop` and `refs/remotes/origin/HEAD` already naming `refs/remotes/origin/develop`
**When** SessionStart runs
**Then** the write seam is invoked 0 times and captured stderr is empty.

```bash
go test ./internal/hook/... -run 'WorktreeBaseBranch.*Match' -count=1 -v
```

### AC-WBR-005 — Mismatch produces the write and exactly one notice line (MUST)

**Given** `worktree_base_branch: develop`, `refs/remotes/origin/develop` existing, and `refs/remotes/origin/HEAD` naming `refs/remotes/origin/main`
**When** SessionStart runs
**Then** the write seam is invoked exactly once with `origin` and `develop`, and captured stderr contains **exactly one** line that names the previous branch, the new branch, and the setting.

```bash
go test ./internal/hook/... -run 'WorktreeBaseBranch.*Mismatch' -count=1 -v
```

The line-count assertion is part of the criterion: two lines FAIL as surely as zero.

### AC-WBR-006 — Fail-open (MUST)

**Given** a git seam that returns a non-zero exit and a non-empty stderr
**When** `Handle` runs
**Then** `Handle` returns a nil error, the returned `HookOutput` is the same allow-shaped output as the no-op case, and the failure is at most logged.

```bash
go test ./internal/hook/... -run 'WorktreeBaseBranch.*(FailOpen|GitError)' -count=1 -v
```

### AC-WBR-007 — `moai cc -w` cuts from the configured base (MUST)

**Given** a real temporary repository (`t.TempDir()`) with two branches and `worktree_base_branch` set to the non-checked-out one
**When** the session worktree is materialized
**Then** the new tree's reflog names that branch as its base.

```bash
go test ./internal/cli/... -run 'SessionWorktree.*Base' -count=1 -v -timeout 600s
# in-test assertion, stated so it can be reproduced by hand:
#   git -C <tree> reflog show <branch> | grep -c 'Created from .*<configured-base>'  → 1
```

`-timeout 600s`: `internal/cli` needs the 600s floor.

### AC-WBR-008 — Empty/unresolvable degrades to today's invocation (MUST)

**Given** an empty value, and separately a value naming a ref that does not exist
**When** the worktree is materialized
**Then** in both cases the recorded argv is exactly `["git","worktree","add","-b",<branch>,<dest>]` — no base operand — and materialization succeeds.

```bash
go test ./internal/cli/... -run 'SessionWorktree.*(NoBase|Unresolvable)' -count=1 -v -timeout 600s
```

The argv assertion is literal: an extra trailing empty-string operand FAILS.

**Third assertion — the shared predicate (REQ-WBR-011).** The unresolvable half of this criterion additionally asserts that consumer 2's unresolvability determination came from the **same** helper consumer 1 uses (REQ-WBR-009), not from a second rule that happens to agree. Assert it structurally, not behaviourally: swap the shared resolvability helper for a fake through the same seam mechanism the package already uses for `sessionWorktreeGitWorktreeAdd` (`internal/cli/session_worktree.go:51-53`), and assert the fake was invoked exactly once with the configured value during materialization. A behavioural-equivalence check (both paths reject `no-such-branch`) does NOT satisfy this — it passes for a divergent second rule, which is the defect the assertion exists to catch.

```bash
go test ./internal/cli/... -run 'SessionWorktree.*(SharedPredicate|Resolver)' -count=1 -v -timeout 600s
```

### AC-WBR-009 — Doctor reports all four states (MUST)

**Given** the four states — (a) empty setting; (b) setting matching `origin/HEAD`; (c) setting differing but resolvable; (d) setting naming a branch that does not exist on the remote
**When** the diagnostic runs
**Then** it returns `uikit.CheckOK` / `CheckOK` / non-OK / non-OK respectively, the (c) `Message` contains the repair command string, the (d) `Message` contains the offending value and directs the user to correct the setting, and the (c) and (d) `Message` strings are **not equal**.

```bash
go test ./internal/cli/... -run 'Doctor.*WorktreeBaseBranch' -count=1 -v -timeout 600s
moai doctor --check 'Worktree Base Branch'   # manual confirmation the item is reachable by name
```

The distinctness assertion is part of the criterion: a single shared "base branch problem" message for (c) and (d) FAILS, because the two repairs differ (run the alignment vs fix the setting).

### AC-WBR-010 — The field is registered and rendered (MUST)

**Given** the built console
**When** the schema and the rendered HTML are inspected
**Then** `git_strategy.worktree_base_branch` appears in `settings.AllFields()` and a form control carrying `name="git_strategy.worktree_base_branch"` appears in the rendered page.

```bash
go test ./internal/web/... -run 'WorktreeBaseBranch' -count=1 -v
go test ./internal/settings/... -run 'AllFields' -count=1
```

This is the positive mirror of the `dead_config_guard_test.go` assertions (`:31-40` schema half, `:44-51` render half).

### AC-WBR-011 — The control is a free-text field naming the two common branches (MUST)

Resolved to the `TypeText` outcome by the operator ruling recorded in plan.md §A D2.1. No longer resolution-dependent.

**Given** the built console
**When** the schema entry and the rendered HTML are inspected
**Then** both of the following hold:

1. **Schema half** — the `FieldDef` whose `Name` is `git_strategy.worktree_base_branch` reports `Type == settings.TypeText` (`internal/settings/schema.go:109`), and an arbitrary third branch name (e.g. `trunk`) is accepted by the field's `Validate` predicate, or the field declares no predicate.
2. **Render half** — the rendered console HTML carries a **text input** control bearing that field's name.

```bash
go test ./internal/settings/... -run 'WorktreeBaseBranch.*Type' -count=1 -v
go test ./internal/web/... -run 'WorktreeBaseBranch.*(Text|FreeText)' -count=1 -v
```

The render-half assertion is modelled on `internal/web/dead_config_guard_test.go`, which renders the console page via `renderConsolePage(t)` and greps the HTML for `name="<field>"` (`:45-52` for the absence direction, `:87-100` for the presence direction). This criterion is the presence direction, plus the input-type check:

```go
html := renderConsolePage(t)
// presence, per dead_config_guard_test.go:89-99
if !strings.Contains(html, `name="git_strategy.worktree_base_branch"`) { t.Error(...) }
// free-text shape: the control is a text input, not a <select> or radio group
if !strings.Contains(html, `type="text" name="git_strategy.worktree_base_branch"`) &&
   !strings.Contains(html, `name="git_strategy.worktree_base_branch" type="text"`) { t.Error(...) }
```

Run-phase MUST confirm the attribute order the renderer actually emits and assert against that single form rather than shipping the either-or above — the two-branch condition is written here only because the emitted order was not measured at plan time (plan.md §B G6).

### AC-WBR-012 — The anti-dead-key regression guard exists and pins all three properties (MUST)

**Given** the key set to a value, and a regression guard modelled on `internal/web/dead_config_guard_test.go`
**When** the guard and the consumer tests run
**Then** all three of REQ-WBR-015's properties are asserted by an executing test — a conjunction, not a disjunction:

1. the key is present in `settings.AllFields()`;
2. a form control carrying `name="git_strategy.worktree_base_branch"` is present in the rendered console HTML;
3. the key reaches a consumer — the git seam receives the value.

```bash
go test ./internal/web/... ./internal/hook/... ./internal/cli/... -run 'WorktreeBaseBranch' -count=1 -v -timeout 600s
```

Two things make this criterion distinct from AC-WBR-010, which asserts properties 1 and 2 as *presence* under REQ-WBR-013:

- **It asserts the guard's existence.** A named test living beside `internal/web/dead_config_guard_test.go` (or in that file) must exist and must carry all three assertions. The consumer tests AC-WBR-005 and AC-WBR-007 already demand do NOT discharge this criterion.
- **It asserts the guard's failure mode.** The guard must FAIL if the field is later removed. Verify by mutation, per the run-phase discipline: delete the `FieldDef` from `gitStrategyFields()` (`internal/settings/schema_sections.go:160-177`), re-run the guard, observe a non-zero exit, then restore. A guard that still passes with the field removed FAILS this criterion.

A `grep` showing the key's name in source text does NOT satisfy this criterion (REQ-WBR-015).

### AC-WBR-013 — Template-first parity (MUST)

**Given** the change set
**When** parity is checked before commit
**Then** every touched file under `.claude/` / `.moai/` has a template counterpart, `make build` succeeds, and any edited hook `.sh` has an identically-edited `.sh.tmpl` twin.

Both checks below are scoped to **this SPEC's own diff**, so the criterion is decidable from this change alone and carries no judgement step. `BASE` is the merge base with the branch this card was cut from.

```bash
make build            # expect exit 0

BASE=$(git merge-base HEAD origin/develop)

# (1) template parity, this diff only: every changed .claude/ or .moai/ path must have
#     a template counterpart under internal/template/templates/ that this diff also changed.
#     The counterpart is EITHER the same-named plain file OR its Go-template `.tmpl` form —
#     some surfaces ship only the `.tmpl` (plan.md §B G5: `.moai/config/sections/git-strategy.yaml`
#     has no plain mirror), so probing only the plain name emits a FALSE
#     NO-TEMPLATE-COUNTERPART for a correct implementation of this SPEC's own §D write list.
#     Report only when NEITHER form was changed in this diff.
git diff --name-only "$BASE"..HEAD -- .claude .moai | grep -v '^.moai/specs/' \
  | while read -r f; do
      git diff --name-only "$BASE"..HEAD -- \
        "internal/template/templates/$f" "internal/template/templates/$f.tmpl" | grep -q . \
        || echo "NO-TEMPLATE-COUNTERPART $f"
    done
# expect no NO-TEMPLATE-COUNTERPART lines

# (2) .sh / .sh.tmpl twin parity, this diff only: for every hook wrapper this diff
#     touched on either side, the two files must be identical afterwards.
git diff --name-only "$BASE"..HEAD -- '*/hooks/moai/*.sh' '*/hooks/moai/*.sh.tmpl' \
  | sed 's/\.tmpl$//' | sort -u \
  | while read -r b; do
      [ -f "internal/template/templates/${b#internal/template/templates/}.tmpl" ] || continue
      diff -q "$b" "$b.tmpl" >/dev/null || echo "DRIFT $b"
    done
# expect no DRIFT lines
```

If this SPEC touches no hook wrapper, check (2) enumerates nothing and passes vacuously **by construction** — that is the intended reading, because the criterion is that this change introduces no drift, not that the repository carries none. Pre-existing drift in wrappers this SPEC did not touch is out of scope (plan.md §D lists no hook wrapper).

### AC-WBR-014 — Typed round trip preserves unmodelled keys (MUST, REQ-WBR-013)

**Given** a `git-strategy.yaml` carrying `manual.develop_branch`, `manual.release_branch_prefix`, and `manual.rc_version_format` (as `.moai/config/sections/git-strategy.yaml:15-17` does)
**When** `worktree_base_branch` is saved through the typed section path
**Then** those three keys are still present in the written file.

```bash
go test ./internal/settings/... -run 'GitStrategy.*RoundTrip' -count=1 -v
```

If this FAILS, the finding is escalated as a blocker rather than absorbed — the drop would be a pre-existing defect this SPEC's write path would newly expose. Promoted from a SHOULD tracing to `R4 / G3` — a risk id and a gap id, not a requirement — to a MUST tracing to REQ-WBR-013, which now carries the preservation clause: it is a correctness property of the write path this SPEC introduces, so it belongs at the requirement layer rather than as an orphaned precondition (plan-audit iter-1 D6). Folded into REQ-WBR-013 rather than given a new id, to stay within the Tier M 16-requirement ceiling (`spec-workflow.md:152`). The test file is named in plan.md §D.

### AC-WBR-015 — Unresolvable value writes nothing (MUST)

The load-bearing consequence of the free-text ruling (plan.md §A D2.1): free text can name a branch that does not exist, and pointing `refs/remotes/origin/HEAD` at a non-existent ref is strictly worse than the defect this SPEC repairs.

**Given** `worktree_base_branch` set to a value with no corresponding remote-tracking branch (e.g. `no-such-branch`), and `refs/remotes/origin/HEAD` naming `refs/remotes/origin/main`
**When** SessionStart runs
**Then** all four hold: the `git remote set-head` seam is invoked **0** times; captured stderr contains **exactly one** line, and that line contains the offending value; `Handle` returns a nil error; and `git symbolic-ref refs/remotes/origin/HEAD` still resolves to the ref it named before the run.

```bash
go test ./internal/hook/... -run 'WorktreeBaseBranch.*Unresolvable' -count=1 -v
```

The resolvability predicate itself is reproducible by hand, and its exit codes were measured in this worktree at HEAD `48eb945df`:

```bash
git show-ref --verify refs/remotes/origin/develop        ; echo "rc=$?"   # observed: rc=0
git show-ref --verify refs/remotes/origin/no-such-branch ; echo "rc=$?"   # observed: rc=128
```

Two assertions bind the implementation shape, not just the outcome:

- The predicate treats **only `rc == 0`** as resolvable. A test written against `rc == 1` would misclassify a missing ref (plan.md §B G7).
- The check uses `git show-ref --verify`. A `git branch --list` / `git branch -vv` implementation FAILS this criterion regardless of its runtime behaviour: BranchGuard's `\bgit\s+branch\b` pattern refuses those invocations at the tool layer (CLAUDE.local.md §4.1.4), so the porcelain form is not merely slower — it is blocked.

### AC-WBR-016 — Firing point: once from the primary checkout, never from a linked worktree (MUST, REQ-WBR-004)

This is REQ-WBR-004's firing-point criterion, and it is distinct from AC-WBR-003 / -004 / -005 / -015: each of those pins an **outcome** for a given config state, so an implementation that ran the comparison only inside `moai doctor` and never registered the errgroup task would satisfy all four while failing REQ-WBR-004 outright (plan-audit iter-1 D1). Both halves below are required; either alone leaves REQ-WBR-004 half-verified.

**The "read seam" denotation is pinned, and both halves below use it in exactly this sense** (plan-audit iter-2 N2; decision recorded at plan.md §A D3.2). The **read seam** is the **alignment-entry seam: the function-variable that reads the configured `git_strategy.worktree_base_branch` value**, and nothing else. It is NOT the `refs/remotes/origin/HEAD` read (a separate seam this criterion does not assert on), and NOT the `git remote set-head` write seam AC-WBR-003 and AC-WBR-005 assert on. The distinction is load-bearing: plan.md:89 orders the helper `read config → no-op silently on empty → read origin/HEAD`, and REQ-WBR-005 forbids any git-metadata read on the empty branch, so under an `origin/HEAD`-read denotation a compliant implementation would record **0** invocations whenever the value is empty and half 1 would fail a correct implementation. Under the pinned denotation the count is **1** for every configured value, empty included, which is what keeps this criterion asserting the thing it exists for — that the task is wired into the errgroup at all (REQ-WBR-004; plan-audit iter-1 D1). A run-phase test asserting on any other seam does NOT satisfy this criterion.

**Half 1 — it fires, exactly once, from the primary checkout.**
**Given** the session's working directory is the primary checkout (`git rev-parse --git-dir` and `git rev-parse --git-common-dir` resolve to the same path), and `worktree_base_branch` carries any value (set or empty)
**When** `Handle` (`internal/hook/session_start.go:66`) is invoked once
**Then** the alignment-entry **read** seam pinned above — the configured-value read — is invoked exactly **1** time, and this holds identically for a set value and for an empty one. Zero invocations FAIL (the task was never registered); two or more FAIL (the task is registered twice, or a caller invokes the helper outside the group).

**Half 2 — it does not fire at all from a linked worktree.**
**Given** the session's working directory is inside a **linked worktree** (the two `rev-parse` paths differ — the discriminant measured at `internal/cli/session_worktree.go:234-241`), and `worktree_base_branch` carries a value that differs from the branch `refs/remotes/origin/HEAD` names and that resolves to an existing remote-tracking branch — that is, the exact state AC-WBR-005 requires a write for
**When** `Handle` runs
**Then** all four hold: the alignment-entry read seam pinned above is invoked **0** times — the primary-checkout gate precedes the configured-value read (plan.md:91), so a compliant implementation never reaches it; the write seam is invoked **0** times; captured stderr is empty; and `Handle` returns a nil error.

```bash
go test ./internal/hook/... -run 'WorktreeBaseBranch.*(Fires|Registered|Once|LinkedWorktree|NotPrimary)' -count=1 -v
```

Three things this criterion pins that no other criterion does:

- **The seam call count is the assertion, not the output.** It must FAIL for an implementation whose behaviour is otherwise correct but whose task is never wired into `Handle`'s errgroup (`internal/hook/session_start.go:120-175`).
- **Half 2 pairs with AC-WBR-005 deliberately.** The same configuration that MUST produce a write from the primary checkout MUST produce nothing from a linked worktree. A half-2 test that only exercises the empty-value path does not discharge this criterion, because the empty value already no-ops under REQ-WBR-005.
- **Consumer 2 is unaffected.** The narrowing binds the alignment step only; AC-WBR-007 and AC-WBR-008 continue to hold from any working tree, linked worktrees included.

The discriminant is reproducible by hand:

```bash
git rev-parse --git-dir ; git rev-parse --git-common-dir   # differ inside a linked worktree, equal in the primary checkout
```

---

## §D.2 Edge cases

- Value with surrounding whitespace (`" develop "`) — trimmed before comparison, or rejected; either is acceptable, silently writing the untrimmed value is not.
- Value naming a branch that exists locally but not on the remote — treated as unresolvable (AC-WBR-008 path), not as a hard error.
- No `origin` remote configured — consumer 1 is a silent no-op; the doctor item reports informationally, not as a failure.
- Repository in a detached-HEAD state — irrelevant to consumer 1 (`origin/HEAD` is a remote ref) and covered by the fail-back for consumer 2.
- Two sessions starting concurrently with the same configured value — the second finds a match and no-ops (AC-WBR-004).
- Two lanes in different card worktrees whose checked-out `git-strategy.yaml` carries different values — neither writes, because REQ-WBR-004's second clause confines consumer 1 to the primary checkout (AC-WBR-016 half 2). This is the case plan.md §E R1 previously mis-analysed as a misconfiguration.

## §D.3 Definition of Done

- **Every `-run` invocation in this file must report at least one executed test. A `go test … -run '<regex>'` run whose output contains `[no tests to run]` FAILS the criterion it was run for, regardless of its exit code.** This binds every criterion discharged by a `-run` regex — AC-WBR-003, -004, -005, -006, -010, -011, -012, -015, -016 among them. Measured at plan time in this tree at HEAD `48eb945df`: `go test ./internal/settings/... -run 'WorktreeBaseBranch.*Type' -count=1` (AC-WBR-011's own command) exits **0** with `[no tests to run]` on three packages, so exit code alone would record a green MUST criterion having executed zero assertions. Pin it mechanically — every `-run` invocation carries `-v`, and its captured output must satisfy:

    ```bash
    go test <pkgs> -run '<regex>' -count=1 -v 2>&1 | tee out.txt
    grep -c '^=== RUN' out.txt   # must be >= 1; 0 FAILS the criterion
    grep -q '\[no tests to run\]' out.txt && echo "VACUOUS — criterion FAILS"
    ```

    The `=== RUN` count and the exit code are BOTH recorded under `.moai/state/verify/t313/`; a criterion is PASS only when both are satisfied.
- All MUST criteria PASS with verbatim command output captured under `.moai/state/verify/t313/`.
- Every criterion in §D is MUST; there is no SHOULD in this set (AC-WBR-014 was promoted to MUST under REQ-WBR-013 at plan-audit iter-1 D6). An AC-WBR-014 failure is escalated as a blocker, never absorbed.
- `go vet ./internal/config/... ./internal/hook/... ./internal/cli/... ./internal/settings/... ./internal/web/...` clean.
- `golangci-lint run` clean on the touched packages.
- `make build` succeeds and template parity holds (AC-WBR-013).
- The D2 widget-shape question is CLOSED (plan.md §A D2.1, operator ruling 2026-08-27: `TypeText`); M5 implements that ruling and does not re-open it.
- No file under `.claude/skills/moai-workflow-project/schemas/` is modified (spec.md §C, card t316 boundary).
