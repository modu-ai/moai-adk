---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Acceptance criteria — card sessions start inside their worktree"
version: "0.2.0"
---

# Acceptance — SPEC-SESSION-DOUBLELOAD-001

Every criterion is binary. Run each check from the worktree root with bash.

| Name | Path or value |
|---|---|
| `E` | `.moai/reports/t1219/m1-measure.md` |
| `CAPS` | `.moai/reports/t1219/m1-caps.md` |
| `P` | `.moai/reports/t1219/probes` |
| `T` / `L` | template / local `kanban-dispatch.md` |
| `DT` / `DL` | template / local `kanban-dispatch-detail.md` |
| `AT` / `A` | `internal/template/templates/AGENTS.md.tmpl` / `AGENTS.md` |
| `G` | `.claude/rules/local/gitflow-lane-protocol.md` |
| `BASE` | `BASE=$(git merge-base develop HEAD)`, recomputed at evaluation time |

Checks that use `BASE` are pre-merge only: after the card merges into develop, the range is empty.

"N/A" means the M2 decision made the check inapplicable. An N/A is recorded, never counted as a pass.

## §D AC Matrix

### Measurement (M1)

- **AC-SDL-001** (REQ-SDL-001, REQ-SDL-016) — Doctrine edits start only after the evidence and after t1175.
  - **Given** the branch,
  - **When** `FIRST=$(git log --reverse --format=%H "$BASE"..HEAD -- T L DT DL AT A G | head -1)` is computed,
  - **Then:**
    - if `FIRST` is empty, the check is N/A and `E` must carry `decision_primary: none`;
    - otherwise `git merge-base --is-ancestor "$(git log --format=%H --diff-filter=A -- E)" "$FIRST"` exits 0;
    - and, with `T1175=$(git log --format=%H -1 --grep='t1175' develop)`, `T1175` is non-empty and `git merge-base --is-ancestor "$T1175" "$FIRST"` exits 0.
- **AC-SDL-002** (REQ-SDL-003) — Caps are committed before any probe output.
  - **Given** `CAPS` and `E`,
  - **When** `C=$(git log --format=%H --diff-filter=A -- CAPS)` and `F=$(git log --reverse --format=%H --diff-filter=A -- P | head -1)` are computed,
  - **Then:**
    - `git show --name-only --format= "$C" | grep -c 'probes/'` prints `0`;
    - `git merge-base --is-ancestor "$C" "$F"` exits 0 and `$C` ≠ `$F`;
    - `grep -cxE 'caps: probes<=8 turns<=4 timeout=300s wall<=45m' CAPS` prints `1`;
    - `grep -cxE 'probes_run: [1-8]' E` prints `1`.
- **AC-SDL-003** (REQ-SDL-002) — Every path records its path sets and token figure.
  - **Given** `E`,
  - **When** `grep -cE '^(skills_paths|instruction_paths|prompt_tokens)_(A|B|C|D|Dp): (gap|/.+|[0-9]+)$' E` runs,
  - **Then:**
    - it prints `15` (`Dp` denotes path D′);
    - every non-gap `*_paths_*` value is a list of absolute paths: `grep -E '^(skills|instruction)_paths_' E | grep -vE ': gap$' | grep -cvE ': /'` prints `0`;
    - `grep -cxE 'prompt_tokens_turn: first-assistant-turn-after-entry' E` prints `1`.
- **AC-SDL-004** (REQ-SDL-004) — Every Gap is named.
  - **Given** `E`,
  - **When** any field of AC-SDL-003, AC-SDL-005 or AC-SDL-008 has value `gap`,
  - **Then** `grep -c '^## Gaps' E` prints `1`, and for each such field `sed -n '/^## Gaps/,/^## /p' E | grep -c '<field-name>'` prints at least `1`.
- **AC-SDL-005** (REQ-SDL-005) — The `/clear` result and its method are recorded.
  - **Given** `E`,
  - **When** `grep -cxE 'clear_restores_single_load: (yes|no|gap)' E` and `grep -cxE 'clear_method: (operator-manual|tmux|none)' E` run,
  - **Then:**
    - each prints `1`;
    - if the first value is `yes`, `grep -cx 'clear_method: none' E` prints `0`.
- **AC-SDL-006** (REQ-SDL-006) — Probes leave no stray writes, processes, or unscrubbed environments.
  - **Given** M1 has finished,
  - **Then** all of these hold:
    - **Environment scrub.** `grep -c '^unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout 300 claude ' P/commands.txt` equals `grep -c . P/commands.txt`, and that count is at least `1`.
    - **No surviving processes.** `for g in $(grep -oE 'pgid=[0-9]+' E | cut -d= -f2); do pgrep -g "$g"; done | wc -l` prints `0`. The sweep is non-empty: `grep -c 'pgid=' E` equals the value of `probes_run`.
    - **No writes to other worktrees.** For every session id in `grep -oE 'probe_session_id: [0-9a-f-]+' E`:
      - it appears in no `.moai/logs` or `.moai/state` file of any worktree listed by `git worktree list --porcelain` other than t1219 (`grep -rl "$SID" <tree>/.moai/logs <tree>/.moai/state` prints nothing);
      - positive control: the same grep over the t1219 or primary `.moai/logs` prints at least one file for at least one probe.
    - **Primary branch unchanged.** `grep -E '^primary_branch_(before|after): ' E | awk '{print $2}' | sort -u | wc -l` prints `1`.
    - **LSEL side effect recorded.** `grep -cxE 'lsel_side_effect: (excluded-by-setting-sources|declared)' E` prints `1`.
- **AC-SDL-007** (REQ-SDL-007) — Zero counts are backed by a two-source positive control.
  - **Given** `E` records any zero duplicate count,
  - **When** `grep -E '^positive_control: extractor=(skills|instructions) input=\.moai/reports/t1219/probes/[^ ]+ count=2$' E` runs,
  - **Then:**
    - it prints at least one line for each class with a zero count;
    - re-running the recorded extractor (`grep '^extractor_skills: ' E`, or `grep '^extractor_instructions: ' E`) on the named input prints `2`.

### Decision (M2)

- **AC-SDL-008** (REQ-SDL-008, REQ-SDL-009, REQ-SDL-010) — The decisions follow from the recorded evidence.
  - **Given** `E`,
  - **When** the fields below are read:
    - `grep -cxE 'path_a_proxy: (none|B)' E`
    - `grep -cxE 'premise_instructions_C: (dup|single|gap)' E`
    - `grep -cxE 'premise_skills_C: (dup|single|gap)' E`
    - `grep -cxE 'decision_primary: (launcher|none)' E`
    - `grep -cxE 'decision_lane_standard: (move-clear-resend|relaunch|none)' E`
  - **Then** each prints `1`, and:
    - **(a) Launcher decision is grounded.** If `decision_primary: launcher`, set `SRC=A` when `path_a_proxy: none`, else `SRC=B`. Then:
      - `grep -cE "^(skills|instruction)_paths_${SRC}: gap$" E` prints `0`;
      - `grep -E "^(skills|instruction)_paths_${SRC}: " E | grep -cE '/moai-adk-go/(CLAUDE\.local\.md|CLAUDE\.md|\.claude/skills|\.claude/rules)'` prints `0` — no primary-checkout path.
    - **(b)** `decision_lane_standard: move-clear-resend` appears only with `clear_restores_single_load: yes`.
    - **(c) Contradicted premise.** If both `premise_*_C` values are `single`, then `decision_primary: none`, `decision_lane_standard: none`, and AC-SDL-001's `FIRST` is empty.
    - **(d) Partly contradicted premise.** If exactly one `premise_*_C` is `single`, `E` § Decision names that class as not driving any doctrine edit (`grep -c 'no doctrine edit for class' E` prints at least `1`).

### Doctrine (M3, M4) — N/A when `decision_primary: none`

- **AC-SDL-009** (REQ-SDL-011) — The `wt` bullet and the new-card clause state the two tiers.
  - **Given** T,
  - **Then:**
    - `grep 'wt. names the new card' T | grep -c 'moai cc -w'` prints `1` (negative control at BASE: `git show "$BASE":internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md | grep 'wt. names the new card' | grep -c 'moai cc -w'` prints `0`);
    - if `decision_lane_standard: move-clear-resend`, `grep 'A new card starts in a new worktree' T | grep -cF 'EnterWorktree(<card-id>)` → `/clear`'` prints `1`;
    - if `relaunch`, `grep 'A new card starts in a new worktree' T | grep -ci 'relaunch'` prints `1`;
    - either way, the same count at BASE prints `0`.
- **AC-SDL-010** (REQ-SDL-011) — The `/clear` handoff clause orders a card change as move → `/clear` → re-send.
  - **Given** T and `decision_lane_standard` ≠ `none`,
  - **When** `S=$(sed -n '/^## The `\/clear` handoff between phases/,/^## [^T]/p' T)` is taken,
  - **Then:**
    - if `move-clear-resend`, `printf '%s\n' "$S" | grep -c 'EnterWorktree'` prints at least `1` (same count at BASE: `0`);
    - if `relaunch`, `printf '%s\n' "$S" | grep -ci 'relaunch'` prints at least `1` (BASE: `0`);
    - and `printf '%s\n' "$S" | grep -c 're-sends the full pointer'` prints `1`.
- **AC-SDL-011** (REQ-SDL-012) — Mirrors stay identical and the other surfaces use the same entry form.
  - **Given** the six surfaces,
  - **Then:**
    - `cmp T L` and `cmp DT DL` both exit 0;
    - `grep -c 'wt: EnterWorktree(t0)' DT` prints `0` (BASE: `1`);
    - `diff <(sed -n '/Start a new card in a new worktree/,/^$/p' A) <(sed -n '/Start a new card in a new worktree/,/^$/p' AT)` exits 0;
    - `sed -n '/Start a new card in a new worktree/,/^$/p' A | grep -c 'launcher'` prints at least `1` (BASE: `0`).
- **AC-SDL-012** (REQ-SDL-013) — Added template lines carry no internal identifiers.
  - **Given** the diff,
  - **When** `git diff "$BASE"..HEAD -- internal/template/templates/ | grep '^+' | grep -v '^+++' | grep -cE '\bt[0-9]{3,4}\b|SPEC-[A-Z][A-Z0-9-]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'` runs,
  - **Then:**
    - it prints `0`;
    - positive control: the same regex over `.moai/reports/t1219/verdict.md` prints at least `1`;
    - sweep non-empty: `git diff --stat "$BASE"..HEAD -- internal/template/templates/ | tail -1` names at least one file changed;
    - `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$|^TestLanguageNeutrality$' -count=1 -v` exits 0, and its output contains both `--- PASS: TestTemplateNoInternalContentLeak` and `--- PASS: TestLanguageNeutrality` (`grep -cE '^--- PASS: (TestTemplateNoInternalContentLeak|TestLanguageNeutrality) '` prints `2`).
- **AC-SDL-013** (REQ-SDL-014) — The local lane protocol matches, with no equal in-session alternative.
  - **Given** G,
  - **Then:**
    - `grep 'moai cc -w <card-id>' G | grep -c 'EnterWorktree(<card-id>)'` prints `0` (BASE: `1`, observed at `9d9e78f6a`);
    - `grep -A3 'moai cc -w <card-id>' G | grep -cE '/clear|relaunch'` prints at least `1` (BASE: `0`);
    - `ls internal/template/templates/.claude/rules/local/gitflow-lane-protocol.md` exits non-zero.
- **AC-SDL-014** (REQ-SDL-015) — No HARD clause is lost.
  - **Given** T,
  - **Then:**
    - `grep -c '^\[HARD\]' T` is at least `hard_clauses_before` from `E` (re-measured after Gate G1);
    - each of these returns at least `1`:
      - `grep -cF 'branch -m WT-<slug>' T`
      - `grep -c 'exit any previous one first' T`
      - `grep -c 'entered through the launcher' T`
      - `grep -c '^## The `/clear` handoff between phases' T`
      - `grep -c 'A companion session does not carry one card' T`

### Scope and quality gate

- **AC-SDL-015** (§E Out of Scope — Runtime or CLI changes)
  - **Given** the branch,
  - **When** `git diff --stat "$BASE"..HEAD -- '*.go'` runs,
  - **Then:**
    - it prints nothing;
    - positive control: `git diff --stat "$BASE"..HEAD -- .moai/specs/SPEC-SESSION-DOUBLELOAD-001` prints a non-empty stat.
- **AC-SDL-016** (quality gate)
  - **Given** the tree,
  - **Then** each of these exits 0:
    - `moai spec lint .moai/specs/SPEC-SESSION-DOUBLELOAD-001`
    - `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json`
    - after M3, `make build`

## §D.1 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-SDL-001 | REQ-SDL-001, REQ-SDL-016 | M1 / G1 |
| AC-SDL-002 | REQ-SDL-003 | M1 |
| AC-SDL-003 | REQ-SDL-002 | M1 |
| AC-SDL-004 | REQ-SDL-004 | M1 |
| AC-SDL-005 | REQ-SDL-005 | M1 |
| AC-SDL-006 | REQ-SDL-006 | M1 |
| AC-SDL-007 | REQ-SDL-007 | M1 |
| AC-SDL-008 | REQ-SDL-008, REQ-SDL-009, REQ-SDL-010 | M2 |
| AC-SDL-009 | REQ-SDL-011 | M3 |
| AC-SDL-010 | REQ-SDL-011 | M3 |
| AC-SDL-011 | REQ-SDL-012 | M3 |
| AC-SDL-012 | REQ-SDL-013 | M3 |
| AC-SDL-013 | REQ-SDL-014 | M4 |
| AC-SDL-014 | REQ-SDL-015 | M3 |
| AC-SDL-015 | (§E Out of Scope — Runtime or CLI changes) | M5 |
| AC-SDL-016 | (quality gate) | M5 |

## §D.2 Edge Cases

- **Path C shows one class loaded once.** For example, a second instruction set but a single skill set. The doctrine is edited on account of the duplicated class only; `E` records the single class with its positive control (AC-SDL-007, AC-SDL-008 (d)).
- **The launcher probe cannot run within the caps.** `*_paths_A: gap` and `path_a_proxy: B`. AC-SDL-008 (a) then reads B.
- **The `/clear` probe cannot run.** `clear_restores_single_load: gap` → `decision_lane_standard: relaunch`.
- **Path D shows primary paths.** The subagent case is not closed by the two tiers; `E` reports it to the orchestrator and no doctrine edit covers it here.
- **Premise contradicted wholesale.** Both decisions are `none`. M3/M4 do not run, AC-SDL-009..014 are N/A, and the orchestrator receives a blocker.
- **t1175 not landed.** Gate G1 holds and M3/M4 wait. AC-SDL-001's t1175 condition fails if doctrine edits appear first.

## §D.3 Quality Gate

- Only changed packages are tested (`./internal/template/`, anchored `-run`); the full suite is left to CI.
- Evidence persists under `.moai/reports/t1219/`, not `/tmp`.

## §D.4 Definition of Done

- AC-SDL-001..008 and AC-SDL-015..016 pass. AC-SDL-009..014 pass, or are N/A by the AC-SDL-008 decision with the N/A recorded.
- `E` carries the five-section report, with a non-empty Gaps section if any field is `gap`.
- All commits name t1219; nothing is pushed.
