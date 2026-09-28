---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Implementation plan — D-scope re-authoring (movement prohibition, skill measurement, N1 amendment draft)"
version: "0.3.0"
---

# Plan — SPEC-SESSION-DOUBLELOAD-001 (v0.3.0, D scope)

Milestones are ordered by decision reversibility: the measurement that can still overturn the doctrine wording comes first, the movement-prohibition doctrine second (wording decisions), the gated N1 amendment third (gated + mechanical), verification last. No time estimates; priority labels only.

## §A Context

- Source of truth for every premise: `.moai/reports/t1279/verdict.md` (§①–§⑩). This re-authoring was ordered by the lead (dispatch to worker-69, 2026-09-29) with a fresh audit budget; the prior round's 3 plan-audits + 1 delta audit are history, not this round's count.
- Tree: `.claude/worktrees/t1279`, branch `WT-prose-residuals` @ `8a969dfc0` (develop `afecf81e9` absorbed; t1175 `7fe658815` is an ancestor). Probes and edits happen here; nothing is edited in the primary checkout.
- Tier M (spec, plan, acceptance, progress). No Go code changes — this SPEC is documentation + measurement.
- Kickoff decision pending: spec.md §F (guard option), resolved at Implementation Kickoff per the lane's standing rule.

## §B Known Issues

- **Line numbers drift.** The verdict's `kanban-dispatch.md:153` target now sits in the `/clear` handoff clause (observed in this tree: the `[HARD]` paragraph "A companion session does not carry one card's context into the next card" at local line 140, template line 140). Every anchor is located by content at run time (REQ-SDL-011).
- **A real `/clear` opens a new transcript file** (delta audit D2: 369/369 `/clear` rows with no preceding assistant row). The REQ-SDL-003 check therefore compares loaded skill sets across sessions — launcher-started baseline vs moved-then-cleared session in the same tree — never same-transcript continuity. This does NOT reopen candidate card #1 (that card redefines the MIDMOVE P2→/clear→P3 continuity proof; here we only compare sets).
- **Headless ≠ interactive.** Probes run headless (`--model haiku`); the launcher-start skill single-source observation (t1219 Evidence 1) was debug-log based and still owes a transcription-based re-confirmation (verdict §① Gaps). Every result names its mode; interactive behavior is not inferred from headless results.
- **Guard rejects heredoc / multi-line probe commands** (prior-round ND5/ND6). Every probe command is a single compound invocation; no heredocs, no dict literals, no `env -u` form.
- **AC grep patterns must resist bypass shapes** (delta audit D3: brace expansion, `-p <flag>` forms). Acceptance patterns are line-anchored (`grep -cxE`) wherever a shape could dodge them.
- **Clause composition (two rules, both stated in the added clause).** (a) The "A new card starts in a new worktree" clause (local line 169) is NOT amended: its content is which tree a card starts in (never reuse), and it composes with the prohibition — its exit-first requirement is satisfied by ending the session and launching in the new tree. (b) The prohibition is scoped to movement **from one card worktree into another during a card session** — the transition the verdict's duplication measurements attach to. It does NOT cover the release/integration-worktree entry the same files mandate (`kanban-dispatch.md` § Integration into the release branch :227-247; `gitflow-lane-protocol.md` §2/§11): that flow keeps its own re-entry rule and is out of scope (spec.md §E composition rule). Only the new prohibition clause and the card-transition `/clear` wording change.
- **Local and template copies are byte-identical today** (`cmp` verified at `8a969dfc0`). The N1 requirement "두 사본의 차이는 기존 분기만 남긴다" (verdict §⑤) therefore means: they stay identical after the edits.

## §C Pre-flight

1. `git -C <tree> rev-parse --show-toplevel` → the t1279 worktree; `git status --short` → empty.
2. `claude --version` recorded into E.
3. `grep -c '^\[HARD\]' T` and the same for L, recorded as `hard_clauses_before:` (re-measured at M3 after any develop absorption).
4. Anchor location recorded: the content grep for the card-transition `/clear` `[HARD]` paragraph in T and L, with line numbers, as `anchor_T:` / `anchor_L:`.
5. `cmp T L` → identical, recorded as `copies_identical_at_baseline: yes`.
6. t1175 landing pre-check recorded (this is re-measured at Gate G1 — the pre-check does not substitute for it): `git merge-base --is-ancestor 7fe658815 HEAD` exit 0.

## §D Constraints — probe isolation

- **Caps file first.** Caps (REQ-SDL-004) are written to `.moai/reports/t1279/m1-caps.md` and committed alone, before any probe output exists. The caps file carries the value-bearing fields `timeout_s: 300` and `wall_cap_minutes: 45` alongside the `caps: probes<=6 turns<=2 model=haiku` line; raising either value is a plan change, not a probe-time choice. A reached cap makes the unobserved item a Gap (REQ-SDL-005).
- **Evidence commits use force-add.** `.moai/reports/*` is gitignored (`.gitignore` `.moai/reports/*`), so committing evidence requires explicit forced paths: `git add -f .moai/reports/t1279/m1-caps.md .moai/reports/t1279/m1-measure.md .moai/reports/t1279/probes/`. A plain `git add` on these paths is silently refused and leaves the commit-based ACs (AC-SDL-001/004) with an empty `EC`.
- **Machine-readable E fields (the contract the AC greps read).** E must carry, each on its own line in the form shown in acceptance.md: `probes_run:`, `skill_dup_tokens:`, `skill_dup_method:`, `clear_restores_skill_set:`, `clear_method:`, `extractor_skills:`, `extractor_instructions:`, `probe_session_id:` (one per probe), `pgid=` (one per probe), `primary_branch_before:` / `primary_branch_after:`, `positive_control:` (when any zero count is recorded), `g1_t1175_ancestor:`, `anchor_T:`, `anchor_L:`. Assembling E without these fields fails AC-SDL-002..006 and AC-SDL-010 for a mechanical reason, not a substantive one.
- **One compound invocation per probe**, in the scrubbed form; commands are appended verbatim to `probes/commands.txt`:

  ```bash
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 300 claude -p "<probe prompt>" --model haiku --max-turns <n> --output-format stream-json --verbose --debug-file .moai/reports/t1279/probes/<X>.debug > .moai/reports/t1279/probes/<X>.jsonl
  ```

- **Scratchpad structure, not the real trees.** Probes run inside a disposable directory structure under `/tmp` (an outer git repo containing a nested inner git repo with a `.claude/worktrees/<w>` L1 worktree, mirroring the verdict probe geometry). The primary checkout and the t1279 worktree are never probe subjects. Fixture copies of instruction files are one-line markers; the token-cost probe swaps in a real-size copy only where the delta requires it (verdict Claim 4 geometry).
- **Loaded-set extraction.** Skill sources and instruction files are read from the session transcript (`Contents of <path>` lines) — transcription-based, per the verdict's own method; debug files are secondary corroboration only. The extractor commands are recorded in E as `extractor_skills:` / `extractor_instructions:`.
- **Token figure.** First assistant turn `usage`: `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`, per probe, recorded per path.
- **Process cleanup.** Each probe's process-group id recorded as `pgid=<n>`; after M1, `pgrep -g` returns nothing for every group. No process killed by name.
- **Positive control.** Any zero duplicate count is backed by a control run on a known two-source input returning a non-zero count (verdict §① Claim 3's own positive-control reasoning).
- Evidence persists under `.moai/reports/t1279/`, never left only in `/tmp`; the `/tmp` scratchpad holds probe runtime state only and is disposable.

## §E Probe Matrix (5 of 6 allowed)

| # | Subject | Start cwd | Action | Turns |
|---|---|---|---|---|
| 1 | launcher-start baseline | scratchpad `<w>` worktree | reply ok | 1 |
| 2 | moved session | scratchpad primary | `EnterWorktree` into `<w>`, then `Read` a worktree file | ≤2 |
| 3 | moved + `/clear` | scratchpad primary | `EnterWorktree` into `<w>`, then the tmux-driven interactive probe of §E.1 (one prompt, `/clear`, one further prompt) | ≤2 turns × 2 |
| 4 | control | scratchpad primary | reply ok (no move) | 1 |
| 5 | positive control | n/a | extractor on a fixture holding two skill sources | — |

- The spare probe is reserved for one retry of a timed-out probe.
- Token cost (REQ-SDL-002) = probe 2 first-turn input minus probe 1 first-turn input, same geometry. Both runs use identical fixture files except where the delta requires a real-size copy (declared in E).
- `/clear` result (REQ-SDL-003): `clear_restores_skill_set: yes` when probe 3's loaded skill set equals probe 1's; `no` when it matches probe 2's duplicated shape; `gap` when probe 3 cannot run within the caps.
- If probe 3 cannot run: `clear_method: none`, and the doctrine still lands (REQ-SDL-007) carrying the recorded Gap — the prohibition does not depend on the `/clear` measurement; only its remedy note does.

### §E.1 Probe 3 — driving the interactive `/clear` (REQ-SDL-003)

A headless `-p` session cannot issue `/clear` and a factory lane has no operator, so probe 3 is driven mechanically in tmux, bounded by an external timeout:

```bash
timeout 300 tmux new-session -d -s <probe-unique-name> -c <scratchpad worktree path>
tmux send-keys -t <probe-unique-name> "<prompt>" Enter      # first turn
tmux send-keys -t <probe-unique-name> "/clear" Enter        # the one /clear
tmux send-keys -t <probe-unique-name> "<prompt>" Enter      # turn after the clear
tmux kill-session -t <probe-unique-name>                    # exact probe-unique session name
```

- The loaded set is read from the interactive session's own transcript file (path and session id recorded in E); extraction stays transcription-based, same extractors as the headless probes. `claude` runs with default model settings in the interactive probe; the model difference vs the headless probes is recorded in E as a declared limitation, not silently ignored.
- The tmux session is ended with `tmux kill-session -t <probe-unique-name>` — an exact, probe-unique session name. Process-name kills remain prohibited (§G).
- **Declared gap route:** where tmux is unavailable, or the probe exceeds the caps, the recorded result is `clear_restores_skill_set: gap` with `clear_method: none` — this is an expected outcome, not a failure; AC-SDL-003 passes on the gap branch with the field named under `## Gaps`.

## §F Milestones

### M1 — Measure (Priority High)

Decision this milestone can overturn: the size of the skill-list cost (does it justify the doctrine change at all?) and whether `/clear` is a usable remedy note.

1. Pre-flight (§C). Write and commit `m1-caps.md` alone (`git add -f`, see §D); the commit message carries t1279.
2. Run probes #1–#5 per §D, §E and §E.1.
3. Write `.moai/reports/t1279/m1-measure.md` (E): the machine-readable E-field contract of §D, plus the five-section report (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
4. Commit E and `probes/` with explicit force-add paths (`docs(t1279): …`).

### M2 — Movement-prohibition doctrine (Priority High)

Gate G0: the Implementation Kickoff decision on the guard option (spec.md §F) is recorded before this milestone starts.

- T (template dispatch rule), edited first: in the launcher-entry section, add one `[HARD]` clause stating launcher start as the standard, the prohibition scoped to movement from one card worktree into another during a card session (REQ-SDL-007 wording — the stable phrase "card worktree into another" is what AC-SDL-007 greps), the two composition rules (new-card clause; integration-window entry exempt with its own re-entry rule), and the unavoidable-move remedy (`/clear` once, after the move). The clause names no mechanism it does not have (REQ-SDL-008) and carries no internal identifiers (REQ-SDL-012).
- Where the kickoff selected the guard: the guard clause (detection rule + denial message) is added in the same section (REQ-SDL-009). Where not: nothing.
- Copy T→L (`cmp` must stay clean — the copies are identical at baseline).
- G (`gitflow-lane-protocol.md` §1): one pointer line to the prohibition clause; no restatement.
- **Untouched:** the "A new card starts in a new worktree" clause, the `/clear` handoff clause (owned by M3), the detail companion, `AGENTS.md`.

### Gate G1 — N1 premise re-measurement (ND2 condition)

M3 does not start until all hold, measured at run time and recorded in E:

1. t1175 landed: `git log --format=%H --grep='t1175' develop` non-empty AND that commit is an ancestor of HEAD after absorbing local develop (absorb first if needed).
2. Anchor re-located by content in both copies: the `[HARD]` paragraph "A companion session does not carry one card's context into the next card" exists in T and L; its line numbers recorded fresh.
3. The value the amendment rewords is re-read at the develop tip (the T1175 value measured there, not the value this plan quotes).

### M3 — N1 amendment substitution (Priority Medium, after G1)

Apply the drafted rewording (draft text below is plan-phase output; the run-phase anchor check of REQ-SDL-011 governs):

- Draft, added to the `/clear` handoff clause of T (then copied to L):

  > [HARD] On a card transition inside a lane session, `/clear` happens exactly once, **after** the session has moved into the next card's worktree — not before the move. A phase end that is not a card transition clears at the phase boundary exactly as stated above. The count is not reduced: one card transition, one `/clear`.

- The existing paragraph's other obligations (operator-typed command, message ordering, leader-session clearing between cards) stay verbatim.
- `cmp T L` clean afterwards.

### M4 — Verification (Priority Low, mechanical)

Run acceptance.md checks; record outputs in `progress.md` §E.2 (run-phase owner). Scoped only: doc surfaces + evidence; the only test executions are the anchored template-neutrality tests of AC-SDL-011.

## §G Anti-Patterns

- Editing any doctrine file before E is committed, or M3 before G1.
- Citing the 2,406 B byte sum as the token cost.
- Writing "launcher start removes the CLAUDE.local.md load" or any equivalent false claim (REQ-SDL-008).
- Amending the "A new card starts in a new worktree" clause or the detail companion (out of scope).
- Locating the amendment anchor by a remembered line number.
- Heredoc or multi-line probe commands; `env -u` form; trailing `kill` instead of `timeout`; killing by name.
- `go test ./...` on the full suite; card ids in template lines.

## §H Cross-References

- `.moai/reports/t1279/verdict.md` — measured premises §①, comparison §②, L2 blockers §③, decision §④, N1 §⑤, audit history §⑤–§⑨, preserved conclusions §⑩.
- `.moai/reports/t1243/m1/evidence.md` @ `9f32f8077` — independent reproduction (verdict §⑧).
- `.claude/rules/moai/workflow/kanban-dispatch.md` / template mirror — § The `/clear` handoff between phases; § Isolation is provisioned by MoAI, then entered through a launcher; § A new card starts in a new worktree (untouched).
- `.claude/rules/local/gitflow-lane-protocol.md` §1 — pointer target.
- `.moai/specs/SPEC-SESSION-MIDMOVE-001` — related prior SPEC; its residual defects are candidate card #2, not this SPEC's scope.
