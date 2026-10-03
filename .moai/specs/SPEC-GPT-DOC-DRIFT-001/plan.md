# plan.md — SPEC-GPT-DOC-DRIFT-001

Implementation plan for the `moai gpt` launcher doc-CLI drift re-pointing (card t1406). Doc plus one contract-token swap (D1 lane amendment); Tier S.

## §A Context

- Card t1406, worktree `.moai/worktrees/t1406`, branch `WT-gpt-launcher-doc-drift`, cut from local `develop` at `1e2151a38` (`1e2151a380a5dd0d76efd8f1740a21f32c682d2f`).
- Approach determined by the lane (Option B — doc re-pointing; grounds and per-site rationale in spec.md §1.4). This plan executes it; it does not re-open it.
- quality.yaml `constitution.development_mode: tdd` — the RED-GREEN loop for this card is expressed through its mechanical checks rather than new Go tests: the sweep is RED-now (6 hits measured, acceptance.md EV-1) and flips green (0 hits, exit 1) when the edits land. No production Go path changes, so no new test code is written; the existing `internal/template` suite plus `make build` act as the regression guards.
- Plan-audit iteration 1 returned FAIL (0.63 < 0.75; audit: `.moai/reports/t1406/plan-audit.md`); the lane decided the D1 scope amendment (decision record: `.moai/reports/t1406/verdict.md` § Plan 감사 1차). This revision applies that delta; everything the audit passed stands.

## §B Known Issues

- **B1 — template-neutrality observation (out of scope).** The goal.md:189 sentence re-pointed by REQ-GDD-004 carries dev-repo vocabulary ("preserve the repository's existing manager ownership and local-develop integration rules") inside a shipped template. Recorded in spec.md §5; deliberately not fixed here.
- **B2 — inventory correction (resolved by measurement).** The lane's inventory said 5 sites / 3 files and "no template twin" for AGENTS.md. Measurement found `internal/template/templates/AGENTS.md.tmpl:324` — the `.tmpl` render source — carrying the identical drifted row. The card covers 6 sites / 4 files; without site #2 the next template deploy re-introduces the drift into user projects.
- **B3 — coordinate correction.** The CLI withdrawal message is at `internal/cli/launcher.go:135` on this tree (the delegation cited :140-142; content identical).
- **B4 — substring hazard.** `moai gpt` legitimately occurs in the README withdrawal notes, `.moai/specs/*` history, and commit messages. Every sweep in this SPEC is scoped to the four guidance files/roots, with the README notes as positive control.
- **B5 — contract-token coupling (D1, resolved by lane decision).** `internal/template/goal_auto_workflow_test.go:52` requires the literal `"moai gpt"` inside the goal workflow's auto section, so the :189 doc edit alone deterministically fails the parity suite (`auto workflow missing contract token`). Amended scope: swap the required token to `"moai codex"` at :52 — one line, inside the required-tokens slice only; the `"moai cc"` prohibition at :61-63 stays untouched and remains satisfied. Sequencing: the swap lands together with the goal.md:189 edits + template mirror — the suite goes green only when they land together.
- **B6 — drift-fed green baseline (D5).** The pre-work suite green is green *because of* the drift (the drifted :189 line feeds the required token). The honest green baseline for AC-GDD-005 exists only after the swap lands — a run-phase observation duty, not a claim made now.

## §C Pre-flight

All measured 2026-10-03 on this tree; baseline evidence pinned in acceptance.md §C.

- HEAD = `git merge-base develop HEAD` = `1e2151a38` — no card commits yet; the branch tip is the develop tip.
- Source ↔ template `goal.md` byte-parity holds green-now (`diff -q` silent, exit 0 — EV-6).
- Baseline sweep RED-now: 6 rows across the four guidance files (EV-1..EV-5).
- Mirror-parity instrument located: `TestGoalAutoWorkflowContractAndMirrorParity` at `internal/template/goal_auto_workflow_test.go:10`.
- Codex goal-parity evidence present: `internal/cli/codex_stop_chain.go`, `internal/cli/codex_goal_parity_test.go`.
- Contract-token RED-now: `grep -n '"moai gpt"' internal/template/goal_auto_workflow_test.go` → `52:		"moai gpt",` (tab-indented slice entry), exit 0 (EV-11) — the swap site; flips to 0 hits / exit 1 after the swap.

## §D Constraints

- Exactly 5 files / 7 line-level edits — the four guidance files (6 edits) plus the one-line contract-token swap; no other file may enter the diff (AC-GDD-006 enforces the file list).
- Stage by explicit pathspec — never `git add -A` / `git add .` (AGENTS.md §2).
- Template edits precede `make build`; `make build` runs before committing (embed regeneration).
- Source and template `goal.md` edits land in the same state — byte-identical (REQ-GDD-005).
- No time estimates anywhere in the deliverables; priority labels only.

## §E Self-Verification

Verification matrix the implementer runs; outputs recorded in progress.md §E.2. Full cells: acceptance.md.

| Check | Command | Expected | AC |
|---|---|---|---|
| Sweep | `grep -rn "moai gpt" AGENTS.md .claude/skills internal/template/templates` | stdout empty, exit 1 | AC-GDD-001 |
| Positive control | `grep -c "moai gpt" README.md README.ko.md README.ja.md README.zh.md` | four lines, each count `1`, exit 0 | AC-GDD-001 |
| New row present | `grep -n "Explicit Claude or GLM session launchers" AGENTS.md internal/template/templates/AGENTS.md.tmpl` | exactly 2 rows, exit 0 | AC-GDD-002 |
| Byte parity | `diff -q .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md` | no output, exit 0 | AC-GDD-003 |
| Build | `make build` | exit 0 (agents-emit-check prestep included) | AC-GDD-004 |
| Contract-token swap | `grep -n '"moai gpt"' internal/template/goal_auto_workflow_test.go` | RED-now: 1 hit at :52, exit 0 (EV-11); after the swap: empty, exit 1 | AC-GDD-006 |
| Package tests | `go test ./internal/template/...` | exit 0 after the doc edits + token swap land together (honest green baseline is post-swap — B6) | AC-GDD-005 |
| Positive text, sites #3/#5 | ``grep -c 'Use the `moai cc` / `moai glm` launchers for this repository' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md`` | two count-rows, each `1`, exit 0 | AC-GDD-008 |
| Positive text, sites #4/#6 | ``grep -c 'Use `moai codex` for the worktree session' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md`` | two count-rows, each `1`, exit 0 | AC-GDD-008 |
| Contribution diff | `git merge-base develop HEAD`, then `git diff --name-only <BASE>..HEAD` | exactly the 5 files — 4 guidance + `goal_auto_workflow_test.go`; exactly one `*_test.go`, zero other `*.go` | AC-GDD-006 |
| Committed-tree zero | `git grep -c "moai gpt" <final-HEAD> -- AGENTS.md .claude/skills internal/template/templates` | empty, exit 1 | AC-GDD-007 |

Affected-package scope only (AGENTS.local.md §4/§6): `go test ./internal/template/...` is the owning package set; the full suite is CI's verdict.

## §F Milestones

Ordered by decision-reversibility — the re-pointing wording decisions lead, mechanical steps close. Pinned replacement texts (lane decision record; the positive ACs bind to exactly these strings):

- goal.md:52 — "Use the `moai gpt` launcher for this repository" → "Use the `moai cc` / `moai glm` launchers for this repository" (rest of the sentence unchanged)
- goal.md:189 — "Use `moai gpt` for the worktree session" → "Use `moai codex` for the worktree session" (rest unchanged)
- the template twins of both lines — identical byte content

- **M1 (Priority High) — template-tree re-points (the decision-bearing edits):**
  - `internal/template/templates/.claude/skills/moai/workflows/goal.md:52` — pinned string swap (above)
  - `internal/template/templates/.claude/skills/moai/workflows/goal.md:189` — pinned string swap (above)
  - `internal/template/templates/AGENTS.md.tmpl:324` — row becomes `| `moai cc` / `moai glm` | Explicit Claude or GLM session launchers |`
- **M2 (Priority High) — contract-token swap (lands together with M1 and M3 — the parity suite goes green only when the doc edits and the swap land together):**
  - `internal/template/goal_auto_workflow_test.go:52` — required token `"moai gpt"` → `"moai codex"` (inside the required-tokens slice only)
  - RED cell: `grep -n '"moai gpt"' internal/template/goal_auto_workflow_test.go` → 1 hit at :52, exit 0 now; empty + exit 1 after the swap
  - the `"moai cc"` prohibition at :61-63 is not touched
- **M3 (Priority High) — local-tree lockstep:**
  - `.claude/skills/moai/workflows/goal.md:52` and `:189` — byte-identical to M1's template edits
  - `AGENTS.md:322` — the same row replacement
  - Confirm parity: `diff -q` silent, exit 0
- **M4 (Priority Medium) — embed refresh:** `make build` (agents-emit-check prestep included)
- **M5 (Priority Medium) — verification and evidence:** run the §E matrix, record verbatim outputs in progress.md §E.2, stage and commit by explicit pathspec with card id `t1406` in the commit body.

Completion: every §E row green; the diff file list is exactly the five files (four guidance + the parity test).

## §G Anti-Patterns

- Do NOT grep-replace `moai gpt` repo-wide — it would rewrite the README withdrawal notes, SPEC history, and adjacent surfaces. Every edit is scoped to the four files.
- Do NOT re-point goal.md:52 to `moai codex` — the surrounding advice is a Claude Code runtime env (`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`), and the automatic cap inject exists only on the unified cc/glm launch path (`internal/cli/launcher.go:817`); the codex path has no cap coupling (measured: `grep -l BlockCap` over the codex chain files → no match). Technically incoherent.
- Do NOT rewrite the sentences around the launcher tokens; do NOT fix B1's neutrality observation in passing (scope discipline).
- Do NOT touch the `"moai cc"` prohibition at `goal_auto_workflow_test.go:61-63` — it stays correct for the codex-shaped auto mission and is satisfied by the :189 replacement.
- Do NOT run the parity suite as a gate between the doc edits and the token swap — the intermediate state fails with `auto workflow missing contract token` by design (B5); they land together.
- Do NOT create a PR — the lane integrates via the develop window (spec.md §6).
- Do NOT read a zero-hit grep's exit 1 as failure — it is the PASS shape for the sweep cells.

## §H Cross-references

- spec.md §1.4 (per-site rationale) · §5 (out of scope) · §6 (delivery)
- acceptance.md (AC cells, evidence ledger, edge cases)
- SPEC-INFINITE-GOAL-001 REQ-2 · commit `2d25a88eb` (card t857)
- `internal/template/goal_auto_workflow_test.go` — `TestGoalAutoWorkflowContractAndMirrorParity`
- `.moai/reports/t1406/plan-audit.md` — iteration-1 FAIL report (defects D1-D7, claims, evidence)
- `.moai/reports/t1406/verdict.md` § Plan 감사 1차 — the lane decision record this revision implements
- AGENTS.local.md §2 (Template-First) · §4 (commit discipline) · §4.1 (git-flow integration chain)
