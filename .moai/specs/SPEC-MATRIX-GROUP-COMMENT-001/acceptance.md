# SPEC-MATRIX-GROUP-COMMENT-001 — Acceptance criteria

Card: **t1055**. Every criterion below is binary and is decided by a command's
output, not by a reading. Run them from the worktree root
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1055`.

`<BASE>` is the commit the branch's comment edit sits on top of — `d8304b49a`
unless the lane has already committed intervening work, in which case it is the
commit immediately preceding the comment commit.

## §D Acceptance criteria

### AC-MGC-001 — the corrected comment names both consumers and calls one a gate
*(REQ-MGC-001)*

**Given** `internal/template/profile_matrix.go` after the change,
**When** the `ResolveAgentModelEffort` doc comment block is extracted,
**Then** it contains all four of: the literal `internal/cli/model.go`, the
literal `internal/web/agentfm.go`, the word `gate`, and the word `display`.

```bash
BLOCK=$(awk '/^\/\/ Lookup is by agent NAME/,/^func ResolveAgentModelEffort/' \
  internal/template/profile_matrix.go)
for needle in 'internal/cli/model.go' 'internal/web/agentfm.go' 'gate' 'display'; do
  printf '%s -> ' "$needle"
  printf '%s' "$BLOCK" | grep -qF "$needle" && echo PRESENT || echo MISSING
done
```

PASS when all four print `PRESENT`.

### AC-MGC-002 — the comment no longer asserts display-only survival
*(REQ-MGC-001)*

**Given** the same doc-comment block,
**When** it is searched for the superseded claim,
**Then** the phrase `only as a display classification` is absent.

```bash
grep -c 'only as a display classification' internal/template/profile_matrix.go
```

PASS when the count is `0`.

### AC-MGC-003 — the sibling site at :261 no longer asserts "display-only"
*(REQ-MGC-002)*

**Given** `internal/template/profile_matrix.go` after the change,
**When** the file is searched for the sibling instance of the same claim,
**Then** the phrase `display-only` is absent from the whole file.

```bash
grep -n 'display-only' internal/template/profile_matrix.go; echo "exit=$?"
```

PASS when no line is printed and `exit=1`.

> **Negative control for AC-MGC-002 and AC-MGC-003.** Both are absence
> assertions, so each needs proof the probe fires. Before the edit, the same two
> commands must print a non-empty result (`1` and line `261` respectively). Run
> them against `<BASE>` and record both outputs:
> ```bash
> git show <BASE>:internal/template/profile_matrix.go | grep -c 'only as a display classification'
> git show <BASE>:internal/template/profile_matrix.go | grep -n 'display-only'
> ```
> A zero/empty result here means the probe is measuring nothing, and the
> post-edit zero is not evidence of anything.

### AC-MGC-004 — no non-comment line changed, and no other file changed
*(REQ-MGC-003)*

**Given** the diff of the comment commit against `<BASE>`,
**When** added and removed lines are filtered to those that are not Go comment
lines,
**Then** the result is empty; and the changed-file list is exactly the one file.

```bash
# (a) every changed line is a comment line
git diff -U0 <BASE> -- internal/template/profile_matrix.go \
  | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' \
  | grep -vE '^[+-][[:space:]]*//' ; echo "residual_exit=$?"

# (b) exactly one file changed
git diff --name-only <BASE>
```

PASS when (a) prints no lines with `residual_exit=1`, and (b) prints exactly
`internal/template/profile_matrix.go` (plus the `.moai/specs/` artifacts of this
SPEC, which are not Go files).

### AC-MGC-005 — line :8 is untouched
*(REQ-MGC-003, `spec.md` §D/§E)*

**Given** the report-only observation about the profile-axis token,
**When** the pre- and post-change copies of line 8 are compared,
**Then** they are byte-identical and still contain `max/medium/low`.

```bash
diff <(git show <BASE>:internal/template/profile_matrix.go | sed -n '8p') \
     <(sed -n '8p' internal/template/profile_matrix.go) && echo IDENTICAL
sed -n '8p' internal/template/profile_matrix.go | grep -c 'max/medium/low'
```

PASS when `IDENTICAL` prints and the count is `1`.

### AC-MGC-006 — the package builds and the affected packages' tests pass
*(REQ-MGC-004)*

**Given** the post-change tree,
**When** the owning package and both consumer packages are built and tested,
**Then** both commands exit 0.

```bash
go build ./internal/template/... ; echo "build_exit=$?"
go test ./internal/template/... ./internal/cli/... ./internal/web/... ; echo "test_exit=$?"
```

PASS when `build_exit=0` and `test_exit=0`.

> **Timeout caveat.** `internal/cli` has been observed to hit the default 10-minute
> package timeout with zero actual failures. If `test_exit` is non-zero, count the
> `--- FAIL` lines before calling it a regression, and re-run with `-timeout 30m`.

## §E Edge cases

- **Comment reflow.** If the rewrite rewraps lines, AC-MGC-004's filter still
  passes (all changed lines remain `//`-prefixed) — the criterion is deliberately
  insensitive to line count.
- **`agentfm_ordering_test.go` lookalikes.** No AC counts `AgentGroup` grep hits,
  precisely because four of them are name lookalikes that do not call
  `template.AgentGroup` (`spec.md` §B). Do not add a count-based AC.

## §F Definition of Done

- AC-MGC-001 through AC-MGC-006 all PASS, with verbatim command output recorded
  in `progress.md` §E.2.
- The negative controls for AC-MGC-002 and AC-MGC-003 are recorded with
  non-empty pre-edit output.
- The commit message names card **t1055**.
- `:8` remains unedited; the §E observation is carried forward unchanged for the
  operator.
