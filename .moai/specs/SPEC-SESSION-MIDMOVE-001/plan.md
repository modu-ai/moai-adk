# Plan — SPEC-SESSION-MIDMOVE-001

## §A — Context

- Card: t1279. Worktree `.claude/worktrees/t1279`, branch `WT-session-double-load`, plan authored at HEAD `2370c5b31`.
- Source of truth: `.moai/reports/t1279/verdict.md` §①–§④. All load judgments come from session transcripts; debug logs are not evidence here.
- Scope: path D (no mid-session worktree moves for card sessions) plus a measurement of the skill-listing duplicate on the move path. Options A and B are out of scope (spec.md §E).
- Tier: **M** under the recommended DP-1 (warn) or docs-only — about 10 doctrine files, 3 evidence files, and at most 2 Go files (the PostToolUse handler and its test). Under DP-1 = block the SPEC re-tiers to **L** before run (spec.md §F), because the guard adds a PreToolUse matcher to the settings template, a new handler, a config key and default, and tests, which exceeds 15 files and introduces a new deny path.

## §B — Decisions first (most likely to change)

Milestones are ordered by how likely each decision is to change. The three decision points in spec.md §F are resolved at the Implementation Kickoff Approval gate, before M1 starts.

| DP | Recommendation | What changes in this plan if the other option is chosen |
|---|---|---|
| DP-1 guard | warn | block → re-tier to L, add design.md + research.md, add M4b (PreToolUse handler, opt-in flag, fail-open, sentinel); docs-only → drop M4 |
| DP-2 integration moves | exempt | covered → M3 adds move → `/clear` → re-send to kanban dispatch lines 250–251, lane protocol lines 37/155, CLAUDE.local.md line 340 |
| DP-3 t1175 ordering | wait | proceed → Gate G1 removed; conflict resolution with `WT-rules-diet` is added to M3's risk list |

## §C — Pre-flight (before M1)

1. Record `git rev-parse --short HEAD` and `git branch --show-current` for this worktree in the evidence file.
2. Build the scratch fixture under the session scratchpad (never inside this repository):
   - `fx/` — a git repository with `CLAUDE.md` (marker A), `CLAUDE.local.md` (marker B), and `.claude/skills/fx-primary-<n>/SKILL.md` (3 skills).
   - `fx/.claude/worktrees/w1` — created with `git -C fx worktree add` (a fixture, not a MoAI card tree; this is the fixture's own git, not this repository's), carrying `.claude/skills/fx-wt-<n>/SKILL.md` (3 skills) and its own `CLAUDE.md` / `CLAUDE.local.md` markers.
   - No `.claude/settings.json` hooks in the fixture; tool permission for the move is granted per probe with `--allowedTools EnterWorktree`.
3. Positive control extract: from the parent session transcript that holds the observed duplicate (spec.md §A), extract only the `skill_listing` attachments' `timestamp`, `isInitial`, `skillCount`, byte size, and the set of scope prefixes in `names`, into `.moai/reports/t1279/m1-control.txt`. No transcript body is copied.

## §D — Milestones

### M1 — Measurement (Priority High; runs now under any DP-3 outcome)

1. **Caps commit.** Write `.moai/reports/t1279/m1-caps.md` with the REQ-SMM-001 caps and the probe list below; force-add and commit it alone (`git add -f`), before any probe runs.
2. **Probes** (each a single compound invocation; the fixture root is the cwd):
   - Headless form: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout -k 10 300 claude -p "<prompt>" --model haiku --max-turns 4 --output-format json [--allowedTools EnterWorktree] [--resume <session-id>]`.
   - P1: started with the fixture worktree as cwd; two turns.
   - P2: started at the fixture root; turn 1 calls `EnterWorktree` on the fixture worktree; turns 2–3 continue.
   - P3: `/clear` on the P2 session, then one turn. Headless `/clear` support is not verified; when it is unavailable, P3 is a Gap (REQ-SMM-005) and is re-run once as an operator-run interactive session in the fixture (`claude` started at the fixture root, operator types the move prompt, then `/clear`, then one prompt). The operator-run session counts toward the 6-session cap and is recorded in `commands.txt` as `# operator-run: <steps>`.
   - Spare budget: 2 sessions for a re-run of a probe whose transcript was unreadable.
3. **Extraction.** For each probe session id, read its transcript from the active Claude config directory's `projects/<fixture-key>/<session-id>.jsonl`:
   - skill-source trees: scope prefixes in `skill_listing.names` (no prefix → start tree);
   - `skill_listing` count and byte sizes;
   - per assistant turn: `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`.
   The same command is run on `m1-control.txt` first; it must report at least one `.claude/worktrees/`-scoped prefix, or M1 stops (the extractor is blind).
4. **Evidence file** `.moai/reports/t1279/m1-measure.md` carries one line per key:
   - `judge_source: transcript`
   - `first_probe_ts: <ISO>` (first transcript line timestamp of the first probe)
   - `listing_trees_P1|P2|P3: <tree,...>|gap`
   - `listing_count_P1|P2|P3: <n>|gap`
   - `turn_tokens_P1|P2|P3: <n,n,...>|gap`
   - `skill_dup_tokens_P2: <int>|gap` — input-side total of the first turn after the move minus the last turn before it (reference only)
   - `launcher_single_listing: yes|no|gap` — yes exactly when `listing_trees_P1` names one tree
   - `clear_restores_single_listing: yes|no|gap` — yes exactly when `listing_trees_P3` names one tree
   - `gap_reason_<key>: <text>` for each gap
   - `probes_run: <n>`, `wall_clock_min: <n>`
   - `repo_head_before|after`, `repo_branch_before|after` (this worktree; must be equal)
5. Commit `m1-measure.md`, `m1-control.txt`, and `commands.txt` (force-added) in one commit that touches no doctrine file.

Note on P1 vs the launcher: P1 starts a headless session inside the fixture worktree; it stands in for `moai cc -w <name>`, which only sets the start directory before launching `claude`. The difference (launcher environment, hooks) is recorded as residual risk, not as a result.

### M2 — Gate G1 (only when DP-3 = wait)

- Record `t1175_landed: <sha>` in `m1-measure.md` once t1175's merge commit is on develop and develop is absorbed here; `<sha>` is the t1175 branch tip that merge brought in. G1 passes when `git merge-base --is-ancestor <sha> HEAD` exits 0. At plan time the tip `3a48485af` is not an ancestor (exit 1) — this is the predicate's negative control.

### M3 — Doctrine, template first (Priority High)

Edit order: template copies, then local mirrors in the same commit, then local-only files.

1. `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`
   - line 93 `wt` bullet: the card's session starts inside the named worktree through the launcher (`moai cc -w <card-id>`, `--spawn` for a new window); where a lane must move instead, the move is followed by `/clear` and the lead re-sends the pointer.
   - "The `/clear` handoff between phases": add one paragraph for card changes: move → operator `/clear` → lead re-sends the full pointer. Every existing sentence stays.
   - isolation table row "Re-enter one from the current session": qualify as not for starting card work.
   - line 179 [HARD] new-card clause: launcher start; exit-first kept; unavoidable move → `/clear` → re-send.
   - lines 250–251: DP-2 wording.
   - REQ-SMM-010 wording: pick the sentence from the two branches below by reading `m1-measure.md`.
2. `kanban-dispatch-detail.md` template: lines 27, 28 (example becomes `wt: moai cc -w <card-id>` form), 188 (factory next card).
3. `worktree-integration.md` template: § `EnterWorktree` / `ExitWorktree` Tools — one card-session caveat sentence after line 224.
4. `AGENTS.md.tmpl` §3, then `AGENTS.md` §3 with identical text: entry line and "Start a new card in a new worktree".
5. Copy each edited template file over its local mirror (byte-identical pairs).
6. Local-only: `.claude/rules/local/gitflow-lane-protocol.md` lines 20, 22, 37, 83, 155; `CLAUDE.local.md` line 340 (DP-2 only).

REQ-SMM-010 wording branches (template text is generic — no card ids, SPEC ids, dates, or hashes):

| Evidence | Sentence the doctrine carries |
|---|---|
| `clear_restores_single_listing: yes` | `/clear` after the move leaves the session with the new worktree's skill listing only |
| `no` or `gap` | `/clear` is still the next step after a move, but it is not shown to remove the listing carried from the start tree; ending the session and relaunching it through the launcher is the way to a single listing |
| `launcher_single_listing: yes` | starting through the launcher gives the session one skill listing |
| `no` or `gap` | the launcher is the entry form; no single-listing claim is made |

### M4 — Guard (only when DP-1 ≠ docs-only; Priority Medium)

- **M4a warn:** extend the PostToolUse handler for `EnterWorktree` so that a move into a card worktree (a path under `.claude/worktrees/` other than an exempt integration tree) adds a model-visible notice: the next step is `/clear`, then the lead re-sends the pointer. Exit 0 always; no config flag needed; a move into an exempt tree gets no notice.
- **M4b block:** a PreToolUse handler for `EnterWorktree`, opt-in flag default false, deny reason prefixed by a fixed sentinel, fail-open on any uncertainty, exempt list read from config. Requires the Tier L artifacts first.

### M5 — Verification and mechanical steps (Priority Low)

- `cmp` the three byte-identical pairs; diff the extracted `AGENTS.md` §3 sections.
- `go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v`.
- Under M4: `go test ./internal/hook/...` for the handler package.
- `moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` and `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json`.
- `make build` after template edits (embedded templates).

## §E — Constraints

- Probes run only in the scratch fixture; this repository's primary checkout and other card worktrees are never a probe cwd or a move target.
- No `--setting-sources` on any probe (plan-audit finding on the held sibling SPEC: it suppresses the CLAUDE.local.md load).
- Raw transcripts and debug output are never committed; the committed evidence is the extracted lines plus `commands.txt`.
- Every commit message names t1279 and ends with the MoAI trailer; staging is by explicit path; nothing is pushed by this card's lane.
- Template text stays neutral: no card ids, SPEC ids, dates, or commit hashes in added template lines.

## §F — Risks

| Risk | Effect | Mitigation |
|---|---|---|
| Headless cannot run `EnterWorktree` or `/clear` | P2/P3 unmeasured | REQ-SMM-005 Gap + one operator-run session within the cap |
| The skill listing in the fixture differs from a real project (3 vs ~40 skills) | Token figure not representative | Record per-skill bytes; state the scale difference as residual risk; the positive control uses the real transcript |
| t1175 rewrites the same clauses | Merge conflict on `[HARD]` text | DP-3 wait; REQ-SMM-014 `[HARD]` count check after absorption |
| t1243/t1259 move CLAUDE.local.md content | CLAUDE.local.md §4.1 edit lands on a moved section | Edit only line 340 (DP-2); re-grep before M3 |
| The always-loaded budget has little headroom after t1175 | Budget guard fails | Keep the added kanban text to a few sentences; run the guard in M5 |
| Probe sessions write to the user-level transcript store | Side effect outside the repo | Accepted and declared; the store is per-fixture-key and holds only fixture sessions |

## §G — Anti-patterns to avoid

- Reading a debug log and reporting it as a load result.
- Reporting a zero-tree result without the positive control.
- Counting lines rather than distinct keys in `m1-measure.md`.
- Writing `/clear removes the duplicate` into the doctrine before P3 says so.
- Editing the local mirror first and copying it to the template.

## §H — Cross-references

- `.moai/reports/t1279/verdict.md` — cause attribution, option comparison, A hold and revival conditions
- `.moai/specs/SPEC-SESSION-DOUBLELOAD-001/` — held sibling; its plan-audit findings (`.moai/reports/t1219/plan-audit-iter2.md`) shaped the probe isolation here
- `.claude/rules/moai/workflow/kanban-dispatch.md`, `worktree-integration.md`, `AGENTS.md` §3
- `internal/hook/post_tool_worktree.go` — existing PostToolUse handler for `EnterWorktree`/`ExitWorktree`
