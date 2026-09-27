# Design — SPEC-SESSION-MIDMOVE-001

Short design note for a Tier L SPEC that is mostly doctrine. The requirements themselves live in spec.md §C.

## 1. Where the doctrine text goes

The always-loaded surface has 70 tokens of headroom (research.md §R3), so the text is split by load scope:

| Content | File | Load | Budget effect |
|---|---|---|---|
| Entry form for a fresh card session (`moai cc -w <card-id>`) | kanban dispatch rule `wt` bullet, new-card paragraph; `AGENTS.md` §3 | always | Paid for by removing the in-session move sequence from the same sentences |
| One sentence in "The `/clear` handoff between phases": a card change is ``move → `/clear` → re-send``, and this is the between-cards `/clear` | kanban dispatch rule | always | Paid for by shortening the isolation-table row and the `wt` parenthetical |
| Full standing-session flow (four steps, who re-sends, one `/clear` per card change, relaunch optional), REQ-SMM-012 sentences, and the reason the `/clear` follows the move | detail companion, new `### Card change in a standing session` | paths | None |
| Card-session caveat on `EnterWorktree` | worktree integration rule § `EnterWorktree` / `ExitWorktree` Tools | paths | None |
| Local lane duties | lane protocol, CLAUDE.local.md §4.1 | local | None (CLAUDE.local.md is not in the budget surface; t1243/t1259 own its size) |

The kanban dispatch rule keeps a pointer to the detail sub-section. The pointer is paid for inside the same offset.

## 2. Standing-session flow

```
card N done → lead reads evidence → lead sends card N+1 pointer
  → session moves into card N+1 worktree (launcher-created tree, or EnterWorktree for an existing tree)
  → operator /clear
  → lead re-sends the full pointer → work starts
```

- Only one `/clear` per card change. The existing between-cards `/clear` moves from before the move to after it; no second `/clear` is added.
- A fresh session opened for a card skips the move and the `/clear`, because it starts inside the tree.
- The session keeps its name, so the dispatch address (role name, `worker-N`) is unchanged. Relaunching is an option, and REQ-SMM-012 names it when `/clear` is not shown to remove the carried listing.

## 3. Extractor contract (`.moai/reports/t1279/extract_listing.py`)

- Input: one transcript JSONL path. With `--cost`, it switches to the REQ-SMM-008 row format.
- Per `skill_listing` attachment it prints one line:
  `listing ts=<ts> isInitial=<bool> skillCount=<n> content_bytes=<utf8 len of content> scoped_names=<n> fixture_trees=<primary|wt|primary,wt|->`
- `scoped_names` counts names matching `^\.claude/worktrees/[^:]+:`.
- `fixture_trees` applies the base-name rule: strip any `<path>:` prefix; `fx-primary-*` is the primary tree, `fx-wt-*` is the worktree tree, and every other name is excluded.
- It also prints `session <id>`, `cwd <value>` for each distinct `cwd`, and `turn tokens=<input-side total>` per assistant row that carries `usage`.
- The controls (`m1-control.jsonl`, `m1-control-neg.jsonl`) use the same row shape, so the extractor that reads real transcripts is the one being tested.

## 4. Hook options (DP-1), for reference

- Both hook options share the same predicate: Kanban/Factory mode (a `MOAI_KANBAN*` variable is set) AND target under `.claude/worktrees/` AND target not in the configured exemption set.
- Warn writes `hookSpecificOutput.additionalContext`.
- Block writes `hookSpecificOutput.permissionDecision: deny` with a sentinel-prefixed `permissionDecisionReason` behind an opt-in flag, and fails open.
- The existing handler's `systemMessage` is not relied on (its model visibility is unverified).
