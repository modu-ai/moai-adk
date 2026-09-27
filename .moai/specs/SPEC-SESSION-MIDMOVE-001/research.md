# Research — SPEC-SESSION-MIDMOVE-001

Observations behind the plan, each with the command that produced it. They were taken at revision time (HEAD `a4da364c8`, Claude Code 2.1.283). They are references for the plan; the run-phase evidence is re-measured.

## R1. Verdict summary (`.moai/reports/t1279/verdict.md`)

- Instruction-file double load comes from upward directory traversal; an L1 worktree also loads the primary CLAUDE.local.md (§① Claims 1–2).
- The skill-listing duplicate is tied to the mid-session move (§① (2)).
- Option A is held, with the revival conditions in §④; option B is owned by t1243/t1259.

## R2. Real-session listing cost (transcript `bb145fe7`)

Command: `python3 <scratchpad>/cost.py <config-dir>/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t1279/bb145fe7-2ccf-4fe8-a9a8-0606c5c74138.jsonl`. `usage_before`/`after` are the input-side totals of the nearest assistant rows with `usage`. Output, abridged to the non-initial scoped rows:

| line | ts | skillCount | content UTF-8 B | scoped names | usage before → after | delta | other rows between (B) |
|---|---|---|---|---|---|---|---|
| 375 | 2026-09-26T11:47:09.164Z | 1 | 725 | 1 | 275,911 → 276,827 | 916 | — |
| 404 | 2026-09-27T03:57:49.175Z | 60 | 43,566 | 42 | 278,878 → 299,707 | 20,829 | 20,526 |
| 1000 | 2026-09-27T08:29:08.010Z | 3 | 2,155 | 3 | 381,727 → 404,547 | 22,820 | 150,068 |
| 1126 | 2026-09-27T08:46:19.764Z | 55 | 45,703 | 44 | 420,230 → 438,668 | 18,438 | 12,410 |

- The initial listing (line 40) is 59,733 B with 146 skills.
- Each delta includes the other rows between the two turns, so it is an upper bound on the listing's cost. Row 1000 shows how badly that bound can be confounded.
- Plan-audit iter-1 reported 46,307 B for row 404 on a different basis (attachment JSON). This SPEC's basis is `content` UTF-8 length.

## R3. Always-loaded budget

`go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v` printed:

- `always-loaded surface = 77530 tokens (budget 77600, headroom 70, 16 entries)`
- `contract document AGENTS.md = 16417 bytes (ceiling 24576, headroom 8159)`
- `contract document internal/template/templates/AGENTS.md.tmpl = 19171 bytes (ceiling 24576, headroom 5405)`
- two `--- PASS:` lines.

The constant `AlwaysLoadedTokenBudget = 77600` is at `internal/config/token_budget_guard.go:86`.

## R4. Load scope of the touched files

`head -5` on each file:

- The kanban dispatch rule has no `paths:` and states "Intentionally always-loaded".
- The detail companion has `paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md"`.
- The worktree integration rule has `paths: "**/.claude/agents/**,**/.claude/worktrees/**,**/.claude/teams/**"`.

## R5. Worktree-session guard and read-time bases

- `git merge-base develop HEAD` printed `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c`.
- `CARD_BASE=$(git merge-base develop HEAD); echo "$CARD_BASE"` was refused by the worktree-session guard ("names git in a form too complex to verify").
- A literal assignment followed by `git rev-parse --verify "$CARD_BASE^{commit}"` and `git diff --name-only "$CARD_BASE"..HEAD | wc -l` ran, printing the SHA and `7`.
- `awk -f <file>` and an `awk` program containing `|`-alternation strings were refused; `python3 -c` and a simple `awk` range program ran. The acceptance checks use those forms.

## R6. `[HARD]` preservation check dry run

At `b59a5d69c`, the `/clear` section extraction gave 3 non-blank lines. `grep -vxF -f <kanban template>` reported 0 missing. A copy with `asks the operator` changed to `tells the operator` reported 1 missing.

## R7. Move-without-`/clear` baseline

The same logic as the AC-SMM-012 Python check (an equivalent inline form, not the byte-identical command), run on the kanban template, the detail template, `AGENTS.md.tmpl`, CLAUDE.local.md, and the lane protocol, printed `5`.

## R8. t1175 overlap

- `git merge-base --is-ancestor WT-rules-diet HEAD` → exit 1.
- The tip moved from `3a48485af` (plan time) to `4989ea6b0` (revision time).
- `git ls-tree -r --name-only WT-rules-diet -- .claude/rules/moai/workflow/ | grep -c kanban-dispatch-mechanics` → `1`.

## R9. Hook output fields

`internal/hook/types.go` has:
- `HookSpecificOutput.AdditionalContext` (`additionalContext`, :333);
- `PermissionDecision` / `PermissionDecisionReason`;
- `HookOutput.SystemMessage`, commented "Warning message shown to user" (:366).

The existing `EnterWorktree` PostToolUse handler (`internal/hook/post_tool_worktree.go`) returns only `SystemMessage`.
