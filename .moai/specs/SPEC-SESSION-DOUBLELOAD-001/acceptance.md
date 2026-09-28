---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Acceptance criteria — D scope (movement prohibition, skill measurement, N1 amendment)"
version: "0.3.0"
---

# Acceptance — SPEC-SESSION-DOUBLELOAD-001 (v0.3.0)

Every criterion is binary. Run each check from the worktree root with bash.

| Name | Path or value |
|---|---|
| `E` | `.moai/reports/t1279/m1-measure.md` |
| `CAPS` | `.moai/reports/t1279/m1-caps.md` |
| `P` | `.moai/reports/t1279/probes` |
| `T` / `L` | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` / `.claude/rules/moai/workflow/kanban-dispatch.md` |
| `G` | `.claude/rules/local/gitflow-lane-protocol.md` |
| `ANCHOR` | the `[HARD]` paragraph in the `/clear` handoff clause beginning `A companion session does not carry one card's context into the next card` |
| `BASE` | `BASE=$(git merge-base develop HEAD)`, recomputed at evaluation time |

Checks using `BASE` are pre-merge only. "N/A" means a milestone's gate made the check inapplicable; an N/A is recorded, never counted as a pass.

## §D AC Matrix

### Measurement (M1)

- **AC-SDL-001** (REQ-SDL-001) — Evidence lands before any doctrine edit.
  - **Given** the branch,
  - **When** `FIRST=$(git log --reverse --format=%H "$BASE"..HEAD -- T L G | head -1)` and `EC=$(git log --format=%H --diff-filter=A -- E | head -1)` are computed,
  - **Then:** if `FIRST` is empty the check is N/A; otherwise `git merge-base --is-ancestor "$EC" "$FIRST"` exits 0 and `$EC` ≠ `$FIRST`.
- **AC-SDL-002** (REQ-SDL-002) — The skill-duplication token cost is recorded with its method; the unusable figure is not cited.
  - **Given** `E`,
  - **Then:** `grep -cxE 'skill_dup_tokens: [0-9]+' E` prints `1`; `grep -cE '^skill_dup_method: first-turn-input-delta' E` prints `1`; and the 2,406 B figure appears only as a rejected figure: `grep -n '2,406' E` prints at least one line, each of which sits under the `## Gaps` or method-declaration heading (`awk`-verified), never as a cost claim.
- **AC-SDL-003** (REQ-SDL-003) — The `/clear` result is recorded with a set-comparison method.
  - **Given** `E`,
  - **Then:** `grep -cxE 'clear_restores_skill_set: (yes|no|gap)' E` prints `1`; `grep -cxE 'clear_method: (operator-manual-set-compare|none)' E` prints `1`; `grep -cE '^clear_method: operator-manual-set-compare' E` prints `1` whenever the first value is not `gap`.
- **AC-SDL-004** (REQ-SDL-004) — Caps committed alone before any probe output.
  - **Given** `CAPS`, `E`, `P`,
  - **When** `C=$(git log --format=%H --diff-filter=A -- CAPS)` and `F=$(git log --reverse --format=%H --diff-filter=A -- P | head -1)` are computed,
  - **Then:** `git show --name-only --format= "$C" | grep -c 'probes/'` prints `0`; `git merge-base --is-ancestor "$C" "$F"` exits 0 and `$C` ≠ `$F`; `grep -cxE 'caps: probes<=6 turns<=2 timeout>=1 wall-declared model=haiku' CAPS` prints `1`.
- **AC-SDL-005** (REQ-SDL-005) — Every Gap is named.
  - **Given** `E` and any field of AC-SDL-002 / AC-SDL-003 holding `gap`,
  - **Then:** `grep -c '^## Gaps' E` prints `1` and the `## Gaps` section names each such field at least once.
- **AC-SDL-006** (REQ-SDL-006) — Probes leave no stray writes, processes, or unscrubbed environments; zero counts carry a positive control.
  - **Given** M1 has finished,
  - **Then** all hold:
    - `grep -c '^unset MOAI_KANBAN' P/commands.txt` equals `grep -c . P/commands.txt` and is at least `2`;
    - `for g in $(grep -oE 'pgid=[0-9]+' E | cut -d= -f2); do pgrep -g "$g"; done | wc -l` prints `0`, with `grep -c 'pgid=' E` equal to the recorded probe count (non-empty sweep);
    - every `probe_session_id` appears in no `.moai/logs` or `.moai/state` file of the primary checkout or the t1279 worktree;
    - when any duplicate count of `0` is recorded, `E` names a positive-control input and re-running the recorded extractor on it prints a count ≥ `2`;
    - `grep -E '^primary_branch_(before|after): ' E | awk '{print $2}' | sort -u | wc -l` prints `1`.

### Doctrine (M2)

- **AC-SDL-007** (REQ-SDL-007) — The prohibition is stated in both copies and pointed to from the lane protocol.
  - **Given** T, L, G,
  - **Then:**
    - `grep -c 'mid-session' T | true` is not the check — the check is content: `sed -n '/^## Isolation is provisioned by MoAI/,/^## /p' T | grep -ciE 'mid-session|in-session' ` prints at least `1`, and the same section names the launcher start and the after-move `/clear` remedy (`| grep -c 'moai cc -w'` ≥ `1`, `| grep -ci '/clear'` ≥ `1`);
    - `cmp T L` exits 0;
    - `grep -c 'kanban-dispatch' G` prints at least `1` (the pointer) and G contains no restated prohibition sentence longer than one line beyond the pointer (manually verified line count recorded in `progress.md` §E.2);
    - negative control at BASE: `git show "$BASE":internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md | sed -n '/^## Isolation is provisioned by MoAI/,/^## /p' | grep -ci 'mid-session'` prints `0`.
- **AC-SDL-008** (REQ-SDL-008) — The added doctrine text makes no CLAUDE.local.md claim.
  - **Given** the diff of added lines in T, L, G,
  - **When** `git diff "$BASE"..HEAD -- T L G | grep '^+' | grep -v '^+++' | grep -c 'CLAUDE.local.md'` runs,
  - **Then** it prints `0`.
- **AC-SDL-009** (REQ-SDL-009) — The guard option follows the kickoff decision.
  - **Given** the recorded kickoff decision `decision_guard: (selected|not-selected)` in `progress.md` §E.1,
  - **Then:** if `not-selected`, `git diff "$BASE"..HEAD -- T L | grep '^+' | grep -vc '^+++'` counts zero guard-clause lines (recorded; the run report names the guard token it searched); if `selected`, the guard clause exists in both T and L (same-content check via `cmp`) and `git diff --stat "$BASE"..HEAD -- '*.go'` prints nothing.

### N1 amendment (M3, after Gate G1)

- **AC-SDL-010** (REQ-SDL-010, REQ-SDL-011) — The amendment is applied per draft, from a re-measured premise.
  - **Given** `E` records the Gate G1 measurements,
  - **Then:**
    - `grep -cxE 'g1_t1175_ancestor: yes' E` prints `1`;
    - `grep -cE '^anchor_T: ' E` and `grep -cE '^anchor_L: ' E` each print `1`, and ANCHOR is present in both T and L at the recorded lines;
    - both copies carry the once-after-the-move wording: `grep -c 'exactly once' T` ≥ `1` and `grep -ci 'after the move\|after the session has moved' T` ≥ `1`; `cmp T L` exits 0;
    - the non-card-transition phase-end wording survives: `grep -c 'asks the operator to `/clear` that session' T` prints `1`;
    - negative control at BASE: `git show "$BASE":internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md | grep -c 'exactly once'` prints `0`.
- **AC-SDL-011** (REQ-SDL-012) — Added template lines carry no internal identifiers.
  - **Given** the diff,
  - **When** `git diff "$BASE"..HEAD -- internal/template/templates/ | grep '^+' | grep -v '^+++' | grep -cE '\bt[0-9]{3,4}\b|SPEC-[A-Z][A-Z0-9-]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'` runs,
  - **Then:** it prints `0`; positive control: the same regex over `.moai/reports/t1279/verdict.md` prints at least `1`; sweep non-empty: `git diff --stat "$BASE"..HEAD -- internal/template/templates/ | tail -1` names at least one file; and `go test ./internal/template/ -run '^TestTemplateNoInternalContentLeak$|^TestLanguageNeutrality$' -count=1` exits 0 with both `--- PASS:` lines present.

### Scope and quality gate (M4)

- **AC-SDL-012** (§E Out of Scope — Go or runtime changes; quality gate)
  - **Given** the branch,
  - **Then:** `git diff --stat "$BASE"..HEAD -- '*.go'` prints nothing (positive control: `git diff --stat "$BASE"..HEAD -- .moai/specs/SPEC-SESSION-DOUBLELOAD-001` prints a non-empty stat); and `moai spec lint .moai/specs/SPEC-SESSION-DOUBLELOAD-001` exits 0.

## §D.1 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-SDL-001 | REQ-SDL-001 | M1 |
| AC-SDL-002 | REQ-SDL-002 | M1 |
| AC-SDL-003 | REQ-SDL-003 | M1 |
| AC-SDL-004 | REQ-SDL-004 | M1 |
| AC-SDL-005 | REQ-SDL-005 | M1 |
| AC-SDL-006 | REQ-SDL-006 | M1 |
| AC-SDL-007 | REQ-SDL-007 | M2 |
| AC-SDL-008 | REQ-SDL-008 | M2 |
| AC-SDL-009 | REQ-SDL-009 | M2 |
| AC-SDL-010 | REQ-SDL-010, REQ-SDL-011 | M3 (Gate G1) |
| AC-SDL-011 | REQ-SDL-012 | M2/M3 |
| AC-SDL-012 | (Out of Scope + quality gate) | M4 |

## §D.2 Edge Cases

- **`/clear` probe cannot run within caps.** `clear_restores_skill_set: gap`, `clear_method: none`; AC-SDL-003 still passes with the gap values named in `## Gaps`; M2 proceeds with the recorded Gap.
- **Token cost measures near zero.** `skill_dup_tokens: <small n>` is a valid result; M2 still runs — the prohibition is ordered by the lead (verdict §④), and the measurement informs the doctrine note, not its existence.
- **Anchor text moved again by a parallel card.** Gate G1 fails the content grep; M3 blocks and the lane reports the divergence — the substitution never proceeds from a stale anchor.
- **Kickoff does not select the guard.** AC-SDL-009 takes the `not-selected` branch; no guard clause exists; no other AC changes.
- **Copies diverge before M2.** Pre-flight records `copies_identical_at_baseline: no`; the lane stops and reports — the "identical where identical today" requirement cannot be evaluated from a diverging baseline.

## §D.3 Quality Gate

- Only anchored template tests run locally (`./internal/template/`, `-run` anchored); the full suite is CI's.
- Evidence persists under `.moai/reports/t1279/`, never only in `/tmp`.

## §D.4 Definition of Done

- AC-SDL-001..006 and AC-SDL-012 pass. AC-SDL-007..010 pass (AC-SDL-009 on its recorded branch). AC-SDL-011 passes on the combined M2+M3 template diff.
- `E` carries the five-section report, with a non-empty `## Gaps` section if any field is `gap`.
- All commits name t1279; nothing is pushed.
