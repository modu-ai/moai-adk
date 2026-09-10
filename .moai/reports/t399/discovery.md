# t399 — plan-phase discovery (measured, not relayed)

tree: `.claude/worktrees/t399` @ `442da4f06` (= `origin/develop`, fetched 2026-09-01)
card: t399 · issue: modu-ai/moai-adk#1632 (Al-Lukyanets)

## Claim / Evidence

### C1 — the bare-string lift is the defect, and it is WIDER than baseBranch

Evidence (source, this tree) — `internal/cli/mcp_codex.go:995-1005` `coerceCodexReviewTarget`:

```go
if s, ok := v.(string); ok && s != "" {
    return map[string]any{"type": s}
}
```

Evidence (protocol ground truth, codex-cli 0.150.1, measured in this tree):

```
codex app-server generate-json-schema --out .moai/reports/t399/schema
→ schema/v2/ReviewStartParams.json  ReviewTarget oneOf:
    uncommittedChanges  required: [type]               ← bare string IS valid
    baseBranch          required: [branch, type]       ← bare string INVALID
    commit              required: [sha, type]          ← bare string INVALID
    custom              required: [instructions, type] ← bare string INVALID
```

Conclusion: `{"type": s}` is well-formed for exactly ONE of four variants. The card
states the baseBranch case; the general form is "a bare string is only well-formed
for `uncommittedChanges`".

### C2 — the caller cannot supply a branch name at all

Evidence — `internal/cli/mcp_server.go:255`:

```go
mcp.WithString("target", mcp.Enum(codexTargetUncommitted, codexTargetBaseBranch), ...)
```

and `internal/cli/mcp_codex.go:1481` `target := req.GetString("target", codexTargetUncommitted)`.

`target` is a string enum with no companion branch parameter. So the fix cannot be
"pass the object through" — the branch must be resolved server-side, or a new
parameter added.

### C3 — a base-branch SSOT already exists (simplicity ladder rung 2)

Evidence:

- `.moai/config/sections/git-strategy.yaml` → `git_strategy.worktree_base_branch: develop`
- `internal/config/loader_worktree_base.go:28` `LoadWorktreeBaseBranch(projectRoot)`
- `internal/cli/mcp_review_material.go:94` fallback chain `origin/HEAD → origin/main → main`

No new resolver needs inventing.

### C4 — no test covers `coerceCodexReviewTarget`

Evidence: `grep -rn "coerceCodexReviewTarget" --include="*_test.go" internal/` → 0 hits.
(The only `baseBranch` hits are worktree `Sync` mocks, unrelated.)

## Reporter items #1-#5 — re-measured (the card marked these UNVERIFIED)

| # | Reporter claim | Measured state in this tree |
|---|---|---|
| #1 | structured `findings[]` always empty | **Already fixed, not by this card.** `internal/cli/mcp_codex.go:1400` `codexFindingsOf` parses `- [P1] msg` bullets into `Finding{Severity,Title,Body,File,Line}`. Landed `4fe2c54c0` (card t234). |
| #2 | envelope `verdict` unreliable both ways | **Already fixed.** `410da655f` / SPEC-CODEX-VERDICT-SYNTH-001 (card t229, #1663). |
| #3 | `gates.codex: required` unenforced | **Partially addressed, still NOT enforced.** `internal/cli/mcp_codex.go:1535` `applyGateUnmet` annotates a fail-open inconclusive with `GateUnmet` when the tree's gate is `required` (`5cfefe2cc`, t234). The verdict is deliberately untouched — the gap is made visible, not blocking. The reporter's "enforce it or rename the key" is still an open policy question. |
| #4 | `auth_provider: "unknown"` while codex works | **Legacy shape addressed.** `dbca3f710` classifies a legacy `auth.json` lacking `auth_mode` (t234). Whether the reporter's specific install still classifies `unknown` is NOT measured here. |
| #5 | spurious instant timeout record | **NOT measured.** `internal/cli/codex_task.go:67` is the only writer of that string; no reproduction attempted. |

## Gaps (explicitly not observed)

- No live `review/start` call was issued from this tree. C1 rests on the schema
  contract, not on a round trip. The regression bar the card sets — "assert the
  native path actually reached codex" — is therefore UNMET so far and belongs to
  run-phase.
- #4 on the reporter's environment; #5 at all.
- t284's participant-count axis was read from the queue text only, not from its SPEC.

## Residual risk

- codex-cli here is 0.150.1; the reporter ran 0.149.0. The schema was measured on
  0.150.1, so a 0.149.0-only difference would not be visible.
