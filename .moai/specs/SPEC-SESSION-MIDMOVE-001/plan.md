# Plan — SPEC-SESSION-MIDMOVE-001

## §A — Context

- Card t1279. Worktree `.claude/worktrees/t1279`, branch `WT-session-double-load`. Plan v0.1.0 at `2370c5b31`; revised after plan-audit iter-1 (`.moai/reports/t1279/plan-audit.md`, FAIL 0.55).
- Source of truth: `.moai/reports/t1279/verdict.md` §①–§④. Load judgments come from transcripts only.
- **Tier L** (reclassified in v0.2.0): 19 REQ and 21 AC exceed the Tier M ceiling of 16. Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md. The plan-auditor PASS threshold is 0.85. The reclassification is made before re-audit, so the next audit is scored against Tier L.

## §B — Decisions first (most likely to change)

The three decision points (spec.md §F) are resolved at the Implementation Kickoff Approval gate and recorded in progress.md §E.2 before M1.

| DP | Recommendation (label) | Plan change under the other option |
|---|---|---|
| DP-1 hook | docs-only | warn → add M5a (mode gate, exemption set, `additionalContext`, visibility probe from the M1 spare budget); block → add M5b (PreToolUse, opt-in flag, sentinel, fail-open) |
| DP-2 integration moves | exempt | covered → M4 adds move → `/clear` → re-send to the integration lines in the kanban rule, lane protocol, and CLAUDE.local.md |
| DP-3 t1175 ordering | wait | proceed → Gate G1 still records `CARD_BASE`, but the §D refresh and the budget "before" figure are taken now; conflicts with t1175 go to M4's risk list |

## §C — Read-time base and worktree-guard form

- `CARD_BASE` is `git merge-base develop HEAD`, read at the moment a check runs (lane protocol rule: measure from the merge-base with the absorbed ref, never from a literal plan-time SHA).
- The worktree-session guard refuses command substitution around git (`CARD_BASE=$(git merge-base develop HEAD)` was refused in this worktree, observed at revision time; research.md §R5). In a worktree session, therefore:
  1. run `git merge-base develop HEAD` alone;
  2. assign the printed SHA literally in the next command;
  3. verify it with `git rev-parse --verify "$CARD_BASE^{commit}"`.

  In an unguarded shell the one-line form is equivalent. The same two-step pattern applies to every derived SHA in acceptance.md.

## §D — Milestones

### M1 — Fixture mechanism check (Priority High; runs now)

1. **Caps commit:** `.moai/reports/t1279/m1-caps.md` holds the REQ-SMM-001 caps and their counting rules. Force-add it (`git add -f`) and commit it alone.
2. **Fixture** (built under the session scratchpad; commands recorded in `fixture.txt`):
   - `fx/` git repo with `CLAUDE.md`, `CLAUDE.local.md`, and `.claude/skills/fx-primary-{a,b,c}/SKILL.md`;
   - `fx/.claude/worktrees/w1` made with the fixture's own `git -C fx worktree add` (a fixture, not a MoAI card tree), carrying `.claude/skills/fx-wt-{a,b,c}/SKILL.md`;
   - no hooks in the fixture; `--allowedTools EnterWorktree` per probe.
3. **Probes** (`probes.txt`: one line per invocation, nothing else):
   - form: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout -k 10 300 claude -p "<prompt>" --model haiku --max-turns 4 --output-format json [--allowedTools EnterWorktree] [--resume <session-id>]`;
   - P1 starts in `w1`; P2 starts at the `fx` root, moves in turn 1, and continues in turns 2–3; P3 is `/clear` then one turn;
   - operator-run substitute (REQ-SMM-005): a line starting `# operator-run:` naming the session id and steps. Its transcript is checked exactly like a headless one;
   - spare: 2 sessions (re-run, or the DP-1 warn visibility probe).
4. **Extractor** (`.moai/reports/t1279/extract_listing.py`, committed; its invocations are recorded in `extract.txt`). For a transcript JSONL, it prints:
   - `session_id`;
   - the distinct `cwd` values;
   - per `skill_listing`: `isInitial`, `skillCount`, `content` UTF-8 bytes, the fixture-tree set (base-name rule, spec.md REQ-SMM-002), and the scoped names;
   - per assistant turn: the input-side usage total.
5. **Controls:**
   - `m1-control.jsonl` is the real line-404 `skill_listing` of transcript `bb145fe7`, reduced to `{"type":"attachment","attachment":{"type","isInitial","skillCount","names"}}` (skill identifiers only; no conversation text);
   - `m1-control-neg.jsonl` has the same shape with every scoped name stripped of its prefix;
   - the extractor must report ≥ 1 scoped name on the first and 0 on the second, or M1 stops.
6. **Evidence** `m1-measure.md` (one key per line; value formats are enforced by AC-SMM-002):
   - `judge_source: transcript`
   - `fixture_root:` (absolute scratch path)
   - `first_probe_ts:`, `first_probe_epoch:`, `last_probe_epoch:`, `wall_clock_min:`
   - `probes_run:` (distinct session ids)
   - `session_id_P1|P2|P3:`, `probe_cwd_P1|P2|P3:`
   - `fixture_trees_P1|P2|P3: primary|wt|primary,wt|gap`
   - `listing_count_P1|P2|P3:`, `listing_bytes_P1|P2|P3:` (comma list), `turn_tokens_P1|P2|P3:` (comma list)
   - `move_adds_listing: yes|no|gap` (yes ⇔ `fixture_trees_P2: primary,wt`)
   - `launcher_single_listing: yes|no|gap` (yes ⇔ `fixture_trees_P1` holds one tree)
   - `clear_restores_single_listing: yes|no|gap` (yes ⇔ `fixture_trees_P3` holds one tree)
   - `gap_reason_<key>:` for every `gap`
   - `repo_head_before|after`, `repo_branch_before|after`
   - added in M3: `t1175_absorbed:`, `always_loaded_before:`; added in M5a only: `additional_context_visible: yes|no|gap`
   - file formats: `extract.txt` lists each command on a line starting `$ ` followed by its output; `m1-cost.md` holds one `command:` line and `row:` lines (or one `cost: gap reason=` line) in the exact shape AC-SMM-009 checks
7. **Commit:** the evidence, controls, extractor, `probes.txt`, `extract.txt`, and `fixture.txt` are force-added in one commit that touches no doctrine file.

### M2 — Real-session cost (Priority High; runs now)

- Run the extractor's cost mode on transcript `bb145fe7`, and write `m1-cost.md`: one row per non-initial scoped `skill_listing` with the REQ-SMM-008 fields, the command, and the label `upper bound`. Research.md §R2 gives the plan-time reading: the row at `03:57:49.175Z` is 43,566 B, 42 scoped names, delta 20,829, other rows 20,526 B.

### M3 — Gate G1 and refresh (only when DP-3 = wait)

1. After t1175 lands on develop and develop is absorbed, record `CARD_BASE` (§C) and `t1175_absorbed: <sha>` in `m1-measure.md`.
2. Re-run the §D grep and record the refreshed table in `.moai/reports/t1279/m3-surface.md`, including any text t1175 moved into `kanban-dispatch-mechanics.md`.
3. Re-measure the `[HARD]` baselines and the always-loaded figure (`always_loaded_before: N`, from the `TestAlwaysLoadedTokenBudget` log line) before the first doctrine commit.
4. Re-check `AGENTS.md` §3 ↔ `AGENTS.md.tmpl` §3 identity at HEAD. When they differ for a reason outside this card, report it to the lead as a pre-existing defect and do not fix it here.

### M4 — Doctrine, template first (Priority High)

Offset rule (REQ-SMM-015): in each always-loaded file (the kanban dispatch rule and `AGENTS.md`), every sentence added is paid for by shortening wording in the same file. The candidates:
- the `wt` bullet's parenthetical move sequence, which moves to the detail companion;
- the "`EnterWorktree(<card-id>)` cannot run from inside a worktree session" explanation in the new-card paragraph, **except** the three sentences REQ-SMM-017 keeps verbatim;
- the isolation table row text.

The full standing-session flow, its rationale, and the `/clear` ordering explanation go to the detail companion (paths-scoped). The kanban rule keeps a one-line pointer.

1. Template kanban dispatch rule:
   - `wt` bullet;
   - one sentence in the `/clear` section: a card change is move → `/clear` → re-send, and the between-cards `/clear` is this one;
   - isolation table row;
   - new-card paragraph;
   - integration bullets per DP-2.
2. Template detail companion:
   - glossary lines 27–28 (the example becomes `wt: moai cc -w <card-id>`);
   - factory `/clear` boundary line 188;
   - a new sub-section "Card change in a standing session": the four-step flow, who re-sends (the lead), one `/clear` per card change, relaunch optional, and the REQ-SMM-012-bounded sentences.
3. Template worktree integration rule: one card-session caveat after line 224 (REQ-SMM-012 bounded; it states the move adds a listing only when `move_adds_listing: yes`).
4. `AGENTS.md.tmpl` §3, then `AGENTS.md` §3 with identical text (net ≤ 0 tokens).
5. Copy each edited template over its local mirror in the same commit.
6. Local-only:
   - lane protocol lines 20, 22, 37, 83, 155;
   - CLAUDE.local.md §4.1: one entry-form sentence (`moai cc -w <card-id>` for a fresh session), the standing flow, and the DP-2 wording on the window steps.

REQ-SMM-012 wording table:

| Evidence key | yes → doctrine may state | no / gap → doctrine states |
|---|---|---|
| `move_adds_listing` | a mid-session move adds the moved-into tree's skill listing | the move is followed by `/clear` and re-send (no claim about listings) |
| `launcher_single_listing` | starting through the launcher gives one skill listing | the launcher is the entry form (no single-listing claim) |
| `clear_restores_single_listing` | `/clear` after the move leaves only the new tree's listing | `/clear` is still the next step; relaunching through the launcher is the way to a single listing (optional) |

### M5 — Hook (only when DP-1 ≠ docs-only; Priority Medium)

- **M5a warn:** handler extension for PostToolUse `EnterWorktree`. It is gated on Kanban/Factory mode (the `MOAI_KANBAN*` environment), uses the exemption set from config, and puts the notice in `additionalContext`, exit 0. Before it lands, one spare M1 session runs the fixture with a hook that emits a unique marker in `additionalContext`, and the transcript must show the model repeating the marker.
- **M5b block:** PreToolUse handler, opt-in flag (default false), sentinel-prefixed deny, fail-open.

### M6 — Verification (Priority Low)

- `go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v`.
- `cmp` the three pairs and `diff` the two §3 sections at HEAD.
- `make build` after template edits.
- `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001`, and `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json`.
- Under M5: `go test ./internal/hook/... -count=1`.

## §E — Constraints

- Probes run only in the scratch fixture: no `--setting-sources`, no debug flags, no raw transcripts committed.
- Every commit message names t1279, carries `Authored-By-Agent: <agent>`, and ends with the MoAI trailer. Staging is by explicit path. The lane pushes nothing.
- Added template lines carry no card ids, SPEC ids, dates, or hashes.

## §F — Risks

| Risk | Mitigation |
|---|---|
| Always-loaded headroom is 70 tokens at `2370c5b31` | Offset rule; explanatory text in paths-scoped files; AC-SMM-015 compares after vs before |
| t1175 changes the same clauses; its tip moves (`3a48485af` → `4989ea6b0`) | DP-3 wait; `CARD_BASE` ranges; M3 refresh of §D and baselines |
| Headless cannot run `/clear` | Gap or operator-run substitute with the same checks |
| Fixture scale (3+3 skills) is not representative | The fixture checks the mechanism only; cost comes from the real transcript (M2) |
| The `bb145fe7` transcript is removed before M2 | M2 records `gap`; research.md §R2 keeps the plan-time reading as a reference, not as evidence |
| Standing lanes may lose the dispatch address on relaunch | Relaunch is optional (REQ-SMM-010); the standard flow keeps the session and its name |

## §G — Anti-patterns to avoid

- Literal plan-time SHAs as range bases.
- Command substitution around git inside a worktree session.
- Counting key names instead of validating key values.
- A control that does not pass through the extractor.
- Adding always-loaded text without an offset.

## §H — Cross-references

- design.md (doctrine placement and extractor design), research.md (measurements behind this plan)
- `.moai/reports/t1279/verdict.md`, `.moai/reports/t1279/plan-audit.md`
- `.moai/specs/SPEC-SESSION-DOUBLELOAD-001/` (held sibling)
