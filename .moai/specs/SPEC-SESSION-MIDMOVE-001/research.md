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

- v0.2.0 rule, excused by any `/clear` substring: `5`. Iter-2 N6 showed that the new-card `[HARD]` paragraph was excused only by the unrelated phrase "reuse without a `/clear` in between".
- v0.3.0 rule, excused only by the P-FLOW phrase: the AC-SMM-012 heredoc (extracted verbatim from acceptance.md) printed `6`.
- The six counted units at `59ecb6582` are:
  - the kanban template `wt`-bullet list;
  - the kanban isolation-table `EnterWorktree(<path>)` row;
  - the kanban new-card `[HARD]` paragraph;
  - the detail glossary `dispatch` row;
  - the `AGENTS.md.tmpl` §3 entry paragraph;
  - the lane protocol launcher line.

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

## R10. t1175 keeps the line the amendment replaces

- Commands: `git show WT-rules-diet:<path>` for the local and template kanban dispatch rule. Then `cmp` of the two, `grep -cxF -f old153.txt` on each (with `old153.txt` extracted from design.md §0), and `grep -c '[HARD]'` on each.
- Observed at t1175 tip `4989ea6b0`:
  - `cmp` rc `0`;
  - old line present `1` / `1`;
  - `[HARD]` count `38` / `38`;
  - `MUST \`ExitWorktree\`` present `1`;
  - `EnterWorktree(<card-id>)` `2`.
- Consequence: the draft in design.md §0 applies unchanged after t1175 absorption, as long as the tip does not move again. M3 step 4 re-checks this.
- Transcript `bb145fe7` carries no `compact_boundary` or `isCompactSummary` row (`grep -c` → `0`). Its single `local_command` row is line 11, before every listing, so every R2 row qualifies for `bound=upper`.

## R11. Plan-time AC ledger (HEAD `59ecb6582`, CARD_BASE `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c`)

Each heredoc was extracted verbatim from acceptance.md (dedented as rendered) and executed. Synthetic evidence was used for the checks whose inputs are produced in the run phase (`fixture_root` = a scratch `fx/` directory).

| AC | Command(s) | Observed output | Reading |
|---|---|---|---|
| 021 | `git rev-parse --verify "$CARD_BASE^{commit}"`; `git diff --name-only "$CARD_BASE"..HEAD \| wc -l` | `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c` rc `0`; `10` | PASS (range non-empty) |
| 022 | `git merge-base --is-ancestor WT-rules-diet HEAD` | rc `1` | RED-now, right reason (t1175 not absorbed) |
| 010 | `grep -c 'moai cc -w <card-id>'` on 7 files | `$WT 0, $AT 0, $DT 0, $KT 0, $LP 1, $AL 0, $CL 0` | RED-now (LP is a regression guard) |
| 011 | P-FLOW on `$DT $AT $LP $CL`; P-CARD in `/clear` section; P-HEAD | `0 0 0 0`; `0`; `0` | RED-now |
| 012 | `EnterWorktree(<card-id>)` on `$KT $LP`; `moai cc -w … 또는 … EnterWorktree`; counter heredoc | `2`, `2`; `1`; `6` | RED-now |
| 013 | heredoc with synthetic `$E` (clear key `gap`); control with P-CLEAR appended to a `$DT` copy | `1`; `2` | RED-now (P-RELAUNCH missing); control fires |
| 014 | P-EXEMPT on the `$KT` integration section, `$LP`, `$CL` | `0`; `0`, `0` | RED-now |
| 015 | 3× `cmp`; §3 extracts; `diff`; per-pair `git log … > file` (8 calls, rc `0`); per-pair `comm -23 \| wc -l`; `$KT` commit count; control `comm` with a foreign SHA | `0 0 0`; `36`, `36`; `0`; `0 0 0 0`; `0`; `1` | runs under the worktree guard; RED-now (no `$KT` commit yet); control fires |
| 016 | budget test; heredoc (synthetic before = 77530, 64316 B); constant diff; constant grep; control before = 77529 | rc `0`, `always-loaded surface = 77530 tokens (budget 77600, headroom 70, 16 entries)`, `--- PASS: TestAlwaysLoadedTokenBudget (0.02s)`, `--- PASS: TestCodexContractByteCeiling (0.01s)`; `0`; `0`; `1`; `1` | PASS at plan time; control fires |
| 017 | added-line neutrality scan; 4-line control; `$KT` added-line count | `0`; `4`; `0` | RED-now on the added-line count (no edit yet) |
| 018 | `old153`/`new153` line counts; preservation heredoc (`$KT`, `$KL`); old-line count; section new-line heredoc; file new-line count; non-kanban `[HARD]` heredoc; controls: moved `MUST` line, amended copy (preservation / section new line / old line) | `1`, `1`; `0`, `0`; `1` (`$KT`), `1` (`$KL`); `0`; `0`; `0`; moved `1`; amended `0` / `1` / `0` | RED-now on old/new; controls fire |
| 019 | `test -e …/rules/local`; template grep; local grep | rc `1`; `0`; `1` | PASS |
| 020 | docs-only diff on `internal/hook/` | `0` | PASS at plan time (docs-only) |
| 001 | heredoc with first_probe_ts `2026-09-28T01:00:00Z` / epoch `1790557200`; CAPS_CT `1790550000` vs `1790560000` | `0`; `1` | check works on synthetic input |
| 002 | heredoc: valid; count 0 with tree `wt`; flipped key | `0`; `2`; `1` | check works; mutants caught |
| 003 | session-id grep; heredoc: valid; `fx-evil` cwd; wrong first cwd | `3`; `0`; `2`; `1` | check works; path-component containment rejects `fx-evil` |
| 006 | heredoc: valid; missing gap reason; wall_clock off by one | `0`; `1`; `1` | check works |
| 007 | heredoc: valid probes; trailing `; rm -rf x`; `fx-evil` cd target; `-d`; status digest differs | `2 0`; `1 1`; `1 1`; `1 1`; `2 1` | check works; unforgeable line form enforced |

Not run at plan time, because their inputs do not exist until the run phase and they need the committed extractor or probe output: AC-004, AC-005, AC-008, AC-009, the AC-020 warn/block branches, and the `$FIRST_KT` half of AC-022. Their commands stay as written and are checked in M6. None of these is recorded as a pass.

### R11 addendum — v0.4.0 delta ledger (HEAD `bce7248ff`, CARD_BASE `b59a5d69c1862b08a8a9e4a48afc0ad33c8d951c`, S = `mktemp -d` read alone)

Every command below ran under the worktree-session guard exactly as written in acceptance.md: `S`, `CARD_BASE`, and the derived SHAs are assigned as literals, and each git call sits on its own line. Heredocs were extracted verbatim from acceptance.md. Synthetic inputs came from a scratch fixture.

| AC | Observed | Reading |
|---|---|---|
| 022 (a) | control count `1`; `t1175.txt` `0` lines; `t1257.txt` `0` lines | RED-now, right reason: neither t1175 (`WT-rules-diet`, tip `8fb81c948`) nor t1257 (`WT-role-naming-docs`, tip `637513578`) has a develop merge commit. `git merge-base --is-ancestor WT-rules-diet develop` → exit `1`; `WT-role-naming-docs` → exit `1` |
| 022 derivation control | the landed-card merge `f01c7889a…` gives `f01c7889abcd… 047a924f17ad…`; that second parent equals `git rev-parse WT-statusline-landed-label` | the `%P` field-3 derivation returns the merged branch tip |
| 022 (b) on that control | merged-in-range grep for `t1175` → `0`; for `t1281` → `9`; `is-ancestor … HEAD` → `rc=1` | provenance rejects a foreign card and accepts the right one. The branch tip itself carries no card id (it is an absorb merge), so the range form is required |
| 002 first heredoc | valid `0`; gap path with `turn_tokens_P3: 1,2` → `1` | ND5 mutant caught |
| 002 second heredoc (cross-check) | valid `0`; unrelated P3 session `1`; clear row removed `5`; correctly gapped no-link evidence `0`; conflict in a non-gap path `1`; conflict path gapped `0`; `listing_bytes_P1` off by one `1` | ND4 linkage, ND3 gap route, and ND5 extract cross-check each fire |
| 007 | valid `2 0`; `; rm -rf x`, `fx-evil`, `-d`, `--max-turns 40`, duplicate `--max-turns`, `--dangerously-skip-permissions`, `--add-dir /x`, `--settings /x.json` → each `2 1` | ND8 allow-list |
| 012 | counter with the narrowed exclusions `6` | RED-now unchanged |
| 013 | clear key `gap`: RED `1`; P-RELAUNCH in `$DT` copy only `0`; in `$AT` copy only `1`; a `목록` line in added lines `2` | ND1 argv[4] fixed; ND11 Korean surface covered |
| 014 covered | `0`, `2`, `1` | RED-now for the covered branch across `$KT`/`$LP`/`$CL` |
| 015 | cmp `0 0 0` (no output); §3 `36`/`36`; diff `0`; 8 `git log` to files; four `comm -23` `0 0 0 0`; `$KT` commits `0` | runs under the guard as written; RED-now on the last count |
| 016 | `rc=0`; PASS-line greps `1`, `1`; `always-loaded surface = 77530 tokens (budget 77600, headroom 70, 16 entries)`; heredoc `0` (before 77530 / 64316 B); constant diff `0`; constant grep `1` | PASS at plan time |
| 018 `$KT` | old/new line files `1`/`1`; preservation `0`; old line `1`; section new line `0`; file new line `0`; re-send sentence `1` | RED-now on the amendment |
| 018 `$KL` | preservation `0`; old line `1`; section new line `0`; re-send sentence `1`; copy with the re-send sentence deleted `0` | both copies checked; ND11 control fires |
| 018 non-kanban `[HARD]` | `0` | PASS |
| 020 docs-only / 021 | `0` / `11` | PASS / range non-empty |

Not run in v0.4.0: AC-009 (no `m1-cost.md` and no extractor yet; the cutoff re-run runs in M2/M6), the AC-020 warn/block branches (no hook under docs-only), and AC-022 (b)/(c) on real landing commits, which cannot exist until t1175 lands. None of these is recorded as a pass.
