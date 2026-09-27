# Design — SPEC-SESSION-MIDMOVE-001

Short design note for a Tier L SPEC that is mostly doctrine. The requirements themselves live in spec.md §C.

## 0. Amendment draft — the kanban dispatch `/clear` `[HARD]` line (lead decision (a))

Source: `verdict.md` §⑤.
- Only a card change moves the phase-end `/clear` to after the worktree move.
- A phase end that is not a card change is unchanged.
- The obligation is relocated, never weakened.
- Both the local and the template copy get the same line, so the two copies keep only their existing divergence (none today: `cmp` rc 0).

The run phase replaces the old line verbatim with the new line below, after absorbing develop with t1175 landed (REQ-SMM-020). The markers delimit exactly one line each; AC-SMM-018 extracts them.

Old line (verbatim at `2370c5b31` and at t1175 tip `4989ea6b0`):

<!-- old-153 -->
[HARD] A companion session does not carry one card's context into the next card. When a phase completes and the lead has read its evidence, the lead **asks the operator to `/clear` that session** — `/clear` is a user-typed command and cannot be sent as an instruction. The lead's message states, in order: what closed (card, phase, evidence read), which session to `/clear` (by name), and what happens next (the next column, and which session is instructed once the clear is done).
<!-- /old-153 -->

New line (draft; replaced verbatim in run):

<!-- amendment-153 -->
[HARD] A companion session does not carry one card's context into the next card. When a phase completes and the lead has read its evidence, the lead **asks the operator to `/clear` that session** — `/clear` is a user-typed command and cannot be sent as an instruction. On a card change, `/clear` exactly once, after the move: the session first moves into the next card's worktree, then the operator clears it, then the lead re-sends the full pointer (move → `/clear` → re-send); a phase end that is not a card change is cleared as before. The lead's message states, in order: what closed (card, phase, evidence read), which session to `/clear` (by name), and what happens next (the next column, and which session is instructed once the clear is done).
<!-- /amendment-153 -->

How the new line relates to the old one:
- Every sentence of the old line survives verbatim.
- One sentence is inserted. It moves the timing of the card-change `/clear` and binds the re-send to the lead.
- The `[HARD]` marker, the one-card-context rule, the user-typed-command rule, and the three-part message structure are unchanged.

## 1. Where the doctrine text goes

The always-loaded surface had 70 tokens of headroom at `2370c5b31` (research.md §R3). Always-loaded text is therefore added only against offsets taken from **non-`[HARD]` lines** of the same file. Every `[HARD]` line except the one in §0 stays verbatim (REQ-SMM-017).

| Content | File | Load | Budget effect |
|---|---|---|---|
| §0 amendment (adds one sentence to one `[HARD]` line) | kanban dispatch rule | always | Paid for by the offsets below |
| Continuation line after the new-card `[HARD]` line (same paragraph, not `[HARD]`): the fresh session starts with `moai cc -w <card-id>`; a standing session follows ``move → `/clear` → re-send`` | kanban dispatch rule | always | Paid for by the offsets below |
| DP-2 canonical exemption sentence | kanban dispatch rule, integration section | always | Paid for by the offsets below |
| Offset candidates (non-`[HARD]` only) | the `wt` bullet's parenthetical move sequence (replaced by a pointer to the detail sub-section); the isolation-table `EnterWorktree(<path>)` row wording. The non-`[HARD]` `Where the next phase reuses a just-cleared session …` sentence is **not** a candidate: it carries the re-send for non-card phase changes | kanban dispatch rule | negative |
| Full standing-session flow (four steps, who re-sends, relaunch optional), REQ-SMM-012 sentences, and the rationale for the order | detail companion `### Card change in a standing session` | paths | none |
| Card-session caveat on `EnterWorktree` | worktree integration rule | paths | none |
| Local lane duties | lane protocol; CLAUDE.local.md §4.1 (net ≤ 600 B) | local | outside the budget surface |

When the offsets cannot reach net ≤ 0 without touching a `[HARD]` line, the run stops with a blocker report to the lead. It does not trim a `[HARD]` line and does not raise the budget constant.

## 2. Standing-session flow

```
card N done → lead reads evidence → lead sends card N+1 pointer
  → session moves into card N+1 worktree
  → operator /clear   (on a card change, /clear exactly once, after the move)
  → lead re-sends the full pointer → work starts
```

- A phase end that is not a card change keeps the existing `/clear` → next-instruction handoff.
- A fresh session opened for a card skips the move and its `/clear`, because it starts inside the tree.
- The session keeps its name, so the dispatch address is unchanged. Relaunch is optional (REQ-SMM-012).
- The exit-first safety sentence of the new-card `[HARD]` line still governs the move: a lane anchored in the previous card's tree exits first, then moves.

## 3. Extractor contract (`.moai/reports/t1279/extract_listing.py`)

- Input: a transcript JSONL path and `--fixture-root <abs>`. With `--cost`, it switches to the REQ-SMM-008 row format.
- Output lines:
  - `session <id>`;
  - `cwd <value>` for each distinct `cwd`, in first-seen order;
  - `turn tokens=<input-side total>` per assistant row with `usage`;
  - per `skill_listing`: `listing ts=<ts> isInitial=<bool> skillCount=<n> content_bytes=<utf8 len of content> scoped_names=<n> trees=<primary|wt|primary+wt|none>`.
- Path attribution (REQ-SMM-002): for each fixture-family name, the source directory is `<fixture root>/<prefix path>/.claude/skills` when a `<path>:` prefix is present, and otherwise `<first cwd>/.claude/skills`. The tree is `wt` when that directory lies under `<fixture root>/.claude/worktrees/w1`, and `primary` otherwise.
- The family cross-check prints `attribution_conflict <name>` when the family disagrees with the tree. Any such line invalidates the evidence (AC-SMM-002).
- `--cost` rows: `row: ts=<ts> skillCount=<n> scoped_names=<n> content_bytes=<n> before=<n> after=<n> delta=<n> other_row_bytes=<n> bound=<upper|confounded>`.
  - `other_row_bytes` sums each in-between row's UTF-8 length without its newline.
  - `bound` follows the REQ-SMM-008 boundary rule.
- The controls (`m1-control.jsonl`, `m1-control-neg.jsonl`) have the same row shape, so the extractor under test is the real one.

## 4. Hook options (DP-1), for reference

- Shared predicate: Kanban/Factory mode (a `MOAI_KANBAN*` variable is set) AND target under `.claude/worktrees/` AND target not in the configured exemption set.
- Warn writes `hookSpecificOutput.additionalContext`.
- Block writes `hookSpecificOutput.permissionDecision: deny` with a sentinel-prefixed `permissionDecisionReason`. It sits behind the opt-in key `workflow.midmove_guard.enabled` (default false) and fails open.
- The existing handler's `systemMessage` is not relied on, because its model visibility is unverified.
- Recommendation stays docs-only (spec.md §F).
