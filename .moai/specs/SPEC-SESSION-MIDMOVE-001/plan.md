# Plan — SPEC-SESSION-MIDMOVE-001

## §A — Context

- Card t1279, worktree `.claude/worktrees/t1279`, branch `WT-session-double-load`.
- Plan history:
  - v0.1.0 at `2370c5b31`;
  - v0.2.0 after plan-audit iter-1 (FAIL 0.55);
  - v0.3.0 after iter-2 (FAIL 0.74, `.moai/reports/t1279/plan-audit-iter2.md`). The next audit is the last one.
- Source of truth: `.moai/reports/t1279/verdict.md` §①–§⑤. Load judgments come from transcripts only.
- **Tier L:** 20 REQ and 22 AC, above the Tier M ceiling of 16. The six artifacts are spec, plan, acceptance, design, research, and progress. The PASS threshold is 0.85.

## §B — Decisions first (most likely to change)

| Decision | State | Plan change under the other option |
|---|---|---|
| N1 `/clear` placement | **settled by the lead: (a)** (verdict §⑤) | — |
| DP-1 hook | recommendation: docs-only (kickoff decides) | warn → add M5a; block → add M5b |
| DP-2 integration moves | recommendation: exempt (kickoff decides) | covered → M4 adds ``move → `/clear` → re-send`` to the integration lines instead of the P-EXEMPT sentence |
| DP-3 ordering | **settled by the lead: wait** — now Gate G1 (REQ-SMM-020) | — |

Kickoff records `dp1:` and `dp2:` in progress.md §E.2 before M1.

## §C — Read-time values and the worktree-guard form

- `CARD_BASE` is `git merge-base develop HEAD`, read when a check runs (lane protocol §8). It is never a literal plan-time SHA.
- The worktree-session guard refuses `$(git …)`, `awk -f`, and loops that pass computed values to `sed`/`git` (research.md §R5). Every acceptance command therefore uses one of three forms:
  - a plain git call with double-quoted variables;
  - a git call writing to a scratch file, followed by a separate command that reads the file;
  - a `python3 - args <<'PY'` heredoc whose body does not name git.
- A derived SHA is printed by one command and assigned literally in the next.

## §D — Milestones

### M1 — Fixture mechanism check (Priority High; runs now)

1. **Caps commit.** `.moai/reports/t1279/m1-caps.md` holds the four REQ-SMM-001 caps and their counting rules. Force-add it and commit it alone.
2. **Fixture** (under the session scratchpad; every command is recorded in `fixture.txt` as `$ <cmd>  # rc=<n>`):
   - `git init fx`, then commit `CLAUDE.md` and `CLAUDE.local.md` markers;
   - create the worktree `fx/.claude/worktrees/w1` with the fixture's own `git -C fx worktree add` **before any skill exists** (this is a fixture, not a MoAI card tree);
   - write `fx/.claude/skills/fx-primary-{a,b,c}/SKILL.md` and `fx/.claude/worktrees/w1/.claude/skills/fx-wt-{a,b,c}/SKILL.md` as untracked files, so neither tree can receive the other's skills;
   - record `test ! -e fx/.claude/worktrees/w1/.claude/skills/fx-primary-a  # rc=0` and `test ! -e fx/.claude/skills/fx-wt-a  # rc=0`;
   - no hooks in the fixture.
3. **Before and after the probes,** record in `m1-measure.md` this repository's `repo_head`, `repo_branch`, and `repo_status_sha` (sha256 of `git status --porcelain`).
4. **Probes** (`probes.txt`: one invocation per line, nothing else except `# operator-run:` lines):
   - the form is `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && cd -- <abs dir under fx> && timeout -k 10 300 claude -p "<plain-word prompt>" --model haiku --max-turns 4 --output-format json [--allowedTools EnterWorktree] [--resume <session-id>]`;
   - `CLAUDE_PROJECT_DIR` is unset so that a user-level hook cannot write into this repository (N13). The prompt text avoids `;&|<>` backquote, `$`, and parentheses;
   - P1 runs with `cd -- <fx>/.claude/worktrees/w1`. P2 runs with `cd -- <fx>`, moves in turn 1, then takes turns 2–3. P3 is `/clear` on the P2 session, then one turn;
   - operator-run substitute: `# operator-run: session=<uuid> steps=<…>`, checked like a headless one (same extractor and `cwd` check);
   - the 2 spare sessions cover one re-run, or the DP-1 warn visibility probe.
5. **Extractor** `.moai/reports/t1279/extract_listing.py` (design.md §3): path-based tree attribution, the family cross-check, and the `listing` / `session` / `cwd` / `turn` / `row:` output lines. Each invocation is recorded in `extract.txt` as `$ python3 .moai/reports/t1279/extract_listing.py …`, followed by its output.
6. **Controls:**
   - `m1-control.jsonl` is the real line-404 `skill_listing` of `bb145fe7`, reduced to `{"type":"attachment","attachment":{"type","isInitial","skillCount","names"}}`;
   - `m1-control-neg.jsonl` has the same shape with every prefix stripped;
   - the extractor must report ≥ 1 and 0 scoped names respectively, or M1 stops.
7. **Evidence** `m1-measure.md`, one key per line:
   - `judge_source`, `fixture_root`, `first_probe_ts`, `first_probe_epoch`, `last_probe_epoch`, `wall_clock_min`, `probes_run`
   - `session_id_P1..P3`, `probe_cwd_P1..P3` (comma list, first = start directory)
   - `listing_count_P1..P3`, `listing_trees_P1..P3` (`;`-separated sets from `primary|wt|primary+wt|none`), `listing_bytes_P1..P3`, `turn_tokens_P1..P3`
   - `launcher_single_listing`, `move_adds_listing`, `clear_restores_single_listing` (derived exactly as REQ-SMM-002 states), and a `gap_reason_<key>` line for every `gap`
   - `repo_head_before/after`, `repo_branch_before/after`, `repo_status_sha_before/after`
8. **Commit:** evidence, controls, extractor, `probes.txt`, `extract.txt`, and `fixture.txt`, force-added in one commit that touches no doctrine file.

### M2 — Real-session cost (Priority High; runs now)

Record the transcript's current line count as `cutoff_line: N` (`wc -l < <transcript>`). Then run `python3 .moai/reports/t1279/extract_listing.py --cost --until-line N <bb145fe7 transcript>` and write `m1-cost.md`: the `cutoff_line:` line, one `command:` line, then one `row:` per non-initial scoped listing, labelled `bound=upper` or `bound=confounded` per the REQ-SMM-008 boundary rule. `other_row_bytes` is newline-excluded. The plan-time reading (research.md §R2) found no compaction/clear boundary between any bracketing pair, so all four rows would read `bound=upper`.

### M3 — Gate G1: absorb, then refresh (REQ-SMM-020; before any doctrine edit)

1. Wait until t1175 has landed on develop; then absorb develop into this branch.
   - The t1175 landing commit is **not** written down by the lane. AC-SMM-022 derives it at check time: it is the second parent of the develop merge commit whose subject merges `WT-rules-diet` into develop.
   - At v0.4.0 no such merge existed (`git merge-base --is-ancestor WT-rules-diet develop` exit 1). **Correction 2026-09-27:** it now exists as merge commit `7fe658815` on develop. This branch has not absorbed it yet, and must not until develop CI is green again.
2. Check t1257 (`WT-role-naming-docs`, waiting on t1256) the same way.
   - When its merge commit exists on develop, absorb it too.
   - When it does not, record `t1257_status: not-landed` and `t1257_notified: yes`, after telling the lead that t1257 must rebase on this change.
3. Re-read `CARD_BASE` (§C). Re-run the §D grep and write `.moai/reports/t1279/m3-surface.md`, including text t1175 moved into `kanban-dispatch-mechanics.md`.
4. Confirm the old line (design.md §0 `old-153`) is still present verbatim in both kanban copies (`grep -cxF` = 1 each). If t1175 or t1257 changed it, stop and return a blocker. The amendment draft must be re-approved against the new text.
5. Record `always_loaded_before: N` (the `TestAlwaysLoadedTokenBudget` log line) and `claude_local_bytes_before: B` (`wc -c < CLAUDE.local.md`).
6. Re-check `AGENTS.md` §3 ↔ `AGENTS.md.tmpl` §3 at HEAD. A difference with a cause outside this card is reported to the lead as a pre-existing defect, and is not fixed here.

### M4 — Doctrine, template first (Priority High)

Rules:
- Every `[HARD]` line of the kanban dispatch rule stays verbatim in its section, except the `old-153` line.
- Offsets come only from non-`[HARD]` lines (design.md §1). The `Where the next phase reuses a just-cleared session …` sentence is not an offset.
- When net ≤ 0 is not reachable, stop with a blocker to the lead.

1. Template kanban dispatch rule:
   - replace the `old-153` line with the `amendment-153` line (design.md §0), byte for byte;
   - add, directly under the new-card `[HARD]` line (same paragraph, not `[HARD]`), one continuation line with `moai cc -w <card-id>` and ``move → `/clear` → re-send``;
   - `wt` bullet: point to the launcher form and to the detail sub-section;
   - isolation-table row: qualify the `EnterWorktree(<path>)` row with ``move → `/clear` → re-send``;
   - integration section: the P-EXEMPT sentence (exempt) or P-FLOW (covered).
2. Template detail companion:
   - glossary lines 27–28 (example `wt: moai cc -w <card-id>`);
   - factory line 188;
   - new `### Card change in a standing session` containing:
     - the four steps, P-FLOW, "the lead re-sends", and relaunch as an option;
     - the REQ-SMM-012 phrases, per the evidence keys;
     - the two ND9 clarifications (design.md §2).
3. Template worktree integration rule: one card-session caveat (P-MOVE only when `move_adds_listing: yes`).
4. `AGENTS.md.tmpl` §3, then `AGENTS.md` §3 with the identical paragraph (entry form + P-FLOW), net ≤ 0 tokens.
5. Copy each edited template over its mirror in the same commit (per pair).
6. Local-only:
   - lane protocol lines 20, 22, 37, 83, 155: launcher-only fresh entry, P-FLOW, and the P-EXEMPT sentence under DP-2 exempt;
   - CLAUDE.local.md §4.1: the entry-form sentence with `moai cc -w <card-id>`, P-FLOW, and the P-EXEMPT sentence; net growth ≤ 600 B.

REQ-SMM-012 wording table (the canonical phrases are in acceptance.md):

| Evidence key | yes → doctrine may state | no / gap → doctrine states |
|---|---|---|
| `move_adds_listing` | P-MOVE | only P-FLOW (no listing claim) |
| `launcher_single_listing` | P-LAUNCH | the launcher is the entry form (no listing claim) |
| `clear_restores_single_listing` | P-CLEAR | `/clear` stays the next step; P-RELAUNCH as an option |

### M5 — Hook (only when DP-1 ≠ docs-only; Priority Medium)

- **M5a warn:** the PostToolUse `EnterWorktree` handler, gated on Kanban/Factory mode and the exemption set, with the notice in `additionalContext`, exit 0. The visibility probe uses a spare M1 session: a fixture hook emits a marker, and the transcript must show the model repeating it (`additional_context_visible:`).
- **M5b block:** a PreToolUse handler behind `workflow.midmove_guard.enabled` (default false), with a sentinel-prefixed deny that fails open. Tier stays L.

### M6 — Verification (Priority Low)

Run every AC in acceptance.md, `make build`, `go run ./cmd/moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001`, and `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json`. Under M5, also `go test ./internal/hook/... -count=1`.

## §E — Constraints

- Probes run only in the scratch fixture: no `--setting-sources`, no debug flags, no raw transcripts committed.
- Every commit names t1279, carries `Authored-By-Agent: <agent>`, and ends with `🗿 MoAI`. Staging is by explicit path. The lane pushes nothing.
- Added template lines carry no card ids, SPEC ids, dates, or hashes.

## §F — Risks

| Risk | Mitigation |
|---|---|
| 70-token always-loaded headroom (at `2370c5b31`) | Offsets from non-`[HARD]` lines only; blocker instead of trimming `[HARD]` or raising the constant |
| t1175 or t1257 changes the `old-153` line | M3 step 4 stops and requests re-approval of the amendment against the new text |
| t1175 tip moves (`3a48485af` → `4989ea6b0`) | `t1175_absorbed` is recorded at absorption; AC-SMM-022 checks ancestry, not a name |
| Headless cannot run `/clear` | Gap, or an operator-run substitute with the same checks |
| Fixture scale (3+3 skills) | The fixture checks the mechanism only; cost comes from the real transcript (M2) |
| `bb145fe7` transcript removed | `cost: gap`; research.md §R2 stays as a reference, not as evidence |

## §G — Anti-patterns to avoid

- Literal plan-time SHAs as range bases.
- `$(git …)`, `awk -f`, or loops with computed values inside a worktree session.
- Name-based tree attribution; union-based mirror comparison.
- Weakening or moving a `[HARD]` line to buy budget.

## §H — Cross-references

- design.md (§0 amendment draft, doctrine placement, extractor), research.md (measurements)
- `.moai/reports/t1279/verdict.md` §⑤, `plan-audit.md`, `plan-audit-iter2.md`
- `.moai/specs/SPEC-SESSION-DOUBLELOAD-001/` (held sibling)
