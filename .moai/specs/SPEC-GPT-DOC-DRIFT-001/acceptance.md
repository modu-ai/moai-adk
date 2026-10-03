# acceptance.md — SPEC-GPT-DOC-DRIFT-001

Verification layer for the `moai gpt` launcher doc-CLI drift re-pointing (card t1406). Stateless on the status axis (no `status:` field — lifecycle lives in spec.md frontmatter alone).

## §A Scope and Verification Discipline

- Doc plus one contract-token card (D1 lane amendment); every criterion below is binary-checkable by grep, diff, build, or test output.
- Two-cell adoption (verification-completeness §2): release-blocking criteria carry a RED-now cell — command, verbatim stdout, exit code, tree SHA — and a green-path cell naming what flips them. Preserve-type criteria (parity, build, tests) are green-now regression guards: observed green before the change, required to stay green through it.
- All RED-now cells measured 2026-10-03 on tree `1e2151a380a5dd0d76efd8f1740a21f32c682d2f` (the card worktree at the develop tip, pre-work).
- grep exit-code semantics: zero matches prints nothing and exits 1. For the sweep cells, **empty stdout + exit 1 is the PASS shape**; exit 0 means hits remain and is a FAIL.
- AC-GDD-004/005 green baselines are run-phase observation duties (audit D5): `make build` was unmeasured pre-work (it mutates the tree under the one-writer rule), and the suite's pre-work green is drift-fed — green *because* `goal.md:189` currently carries the token the parity test requires (audit D1/Claim 2). The honest green baseline for AC-GDD-005 exists only after the contract-token swap lands; both outputs are observed and recorded at run-phase entry.

## §B AC Matrix

| AC | Verifies | Check command(s) | RED-now baseline | Green path | Class |
|----|----------|-----------------------------|------------------|------------|-------|
| AC-GDD-001 | REQ-GDD-001 | `grep -rn "moai gpt" AGENTS.md .claude/skills internal/template/templates` | EV-1: 6 rows, exit 0 | stdout empty, exit 1 — with EV-9 positive control intact | release-blocking |
| AC-GDD-002 | REQ-GDD-002 | `grep -n "Explicit Claude or GLM session launchers" AGENTS.md internal/template/templates/AGENTS.md.tmpl` | EV-7: empty, exit 1 | exactly 2 rows (one per file), exit 0 | release-blocking |
| AC-GDD-003 | REQ-GDD-005 | `diff -q .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md` | EV-6: silent, exit 0 | unchanged — silent, exit 0 | regression-guard |
| AC-GDD-004 | REQ-GDD-006 | `make build` | unmeasured pre-work (build mutates the tree) — run-phase observation duty | exit 0, agents-emit-check prestep included — observed and recorded at run-phase entry | regression-guard |
| AC-GDD-005 | REQ-GDD-005, REQ-GDD-007 | `go test ./internal/template/...` | current green is drift-fed (audit D1/Claim 2) — honest baseline exists only post-swap | exit 0 after the doc edits + token swap land together — observed and recorded at run phase | regression-guard |
| AC-GDD-006 | REQ-GDD-008 | `git merge-base develop HEAD`, then `git diff --name-only <BASE>..HEAD` | EV-8: empty, exit 0 (pre-work) | exactly 5 files — the 4 guidance files + `internal/template/goal_auto_workflow_test.go`; exactly one `*_test.go`, zero other `*.go` | release-blocking |
| AC-GDD-007 | REQ-GDD-001, REQ-GDD-002 | `git grep -c "moai gpt" <final-HEAD> -- AGENTS.md .claude/skills internal/template/templates` | EV-12: four count-rows totalling 6, exit 0 (committed tree at pin) | empty, exit 1 at the card's final HEAD | release-blocking |
| AC-GDD-008 | REQ-GDD-003, REQ-GDD-004 | ``grep -c 'Use the `moai cc` / `moai glm` launchers for this repository' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md`` ; ``grep -c 'Use `moai codex` for the worktree session' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md`` | EV-13/EV-14: two `:0` count-rows each, exit 1 | four count-rows, each `1`, exit 0 — one per pattern per twin | release-blocking |

Check-form note: every cell above is a single-invocation plain command except AC-GDD-006, whose green path is a two-step plain-command sequence (`git merge-base develop HEAD`, then `git diff --name-only <BASE>..HEAD`); all RED cells are single-invocation (the form verification-completeness §2.1 binds).

## §C Evidence Ledger (RED-now, tree `1e2151a380a5dd0d76efd8f1740a21f32c682d2f`, 2026-10-03)

```
EV-1  cmd:   grep -rn "moai gpt" AGENTS.md .claude/skills internal/template/templates
      exit:  0
      out:   6 rows, raw stdout verbatim (cross-checked against plan-audit.md Claim 1):
AGENTS.md:322:| `moai cc` / `moai glm` / `moai gpt` | Explicit Claude, GLM, or GPT session launchers |
.claude/skills/moai/workflows/goal.md:52:An infinite goal armed with `moai goal arm "<condition>" --max-turns 0 --max-duration <seconds>` (the wall-clock primary bound) is bounded only by the REAL bounds (wall-clock / cost / stagnation) — but the default `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=8` silently terminates it first. Raise `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` (e.g. to 200) when arming a `--max-turns 0` goal. Use the `moai gpt` launcher for this repository; for an already-running session, set `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` in the environment before arming so the runtime cap does not pre-empt the infinite loop.
.claude/skills/moai/workflows/goal.md:189:Use `moai gpt` for the worktree session and preserve the repository's existing
internal/template/templates/AGENTS.md.tmpl:324:| `moai cc` / `moai glm` / `moai gpt` | Explicit Claude, GLM, or GPT session launchers |
internal/template/templates/.claude/skills/moai/workflows/goal.md:52:An infinite goal armed with `moai goal arm "<condition>" --max-turns 0 --max-duration <seconds>` (the wall-clock primary bound) is bounded only by the REAL bounds (wall-clock / cost / stagnation) — but the default `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=8` silently terminates it first. Raise `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` (e.g. to 200) when arming a `--max-turns 0` goal. Use the `moai gpt` launcher for this repository; for an already-running session, set `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` in the environment before arming so the runtime cap does not pre-empt the infinite loop.
internal/template/templates/.claude/skills/moai/workflows/goal.md:189:Use `moai gpt` for the worktree session and preserve the repository's existing

EV-2  cmd:   grep -c "moai gpt" AGENTS.md
      exit:  0      out: 1

EV-3  cmd:   grep -c "moai gpt" .claude/skills/moai/workflows/goal.md
      exit:  0      out: 2

EV-4  cmd:   grep -c "moai gpt" internal/template/templates/.claude/skills/moai/workflows/goal.md
      exit:  0      out: 2

EV-5  cmd:   grep -c "moai gpt" internal/template/templates/AGENTS.md.tmpl
      exit:  0      out: 1

EV-6  cmd:   diff -q .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md
      exit:  0      out: (empty — byte-identical green-now)

EV-7  cmd:   grep -n "Explicit Claude or GLM session launchers" AGENTS.md internal/template/templates/AGENTS.md.tmpl
      exit:  1      out: (empty — the re-pointed row text does not exist yet)

EV-8  cmd:   git diff --name-only 1e2151a38..HEAD
      exit:  0      out: (empty — pre-work: HEAD is the merge-base)

EV-9  cmd:   grep -c "moai gpt" README.md README.ko.md README.ja.md README.zh.md
      exit:  0      out: README.md:1 / README.ko.md:1 / README.ja.md:1 / README.zh.md:1
             (positive control — the withdrawal notes, one per locale)

EV-10 cmd:   grep -l "BlockCap" internal/cli/codex_stop_chain.go internal/cli/codex_launcher.go
      exit:  1      out: (empty — no cap coupling on the codex path; grounds for the
             site #3 re-point target)

EV-11 cmd:   grep -n '"moai gpt"' internal/template/goal_auto_workflow_test.go
      exit:  0      out: 52:		"moai gpt",
             (D1 RED cell — the required-token line the swap edits; the tab-indented
             entry sits inside the required-tokens slice)

EV-12 cmd:   git grep -c "moai gpt" HEAD -- AGENTS.md .claude/skills internal/template/templates
      exit:  0
      out:   HEAD:.claude/skills/moai/workflows/goal.md:2
             HEAD:AGENTS.md:1
             HEAD:internal/template/templates/.claude/skills/moai/workflows/goal.md:2
             HEAD:internal/template/templates/AGENTS.md.tmpl:1
             (AC-GDD-007 RED cell — committed tree at the pin; total 6 occurrences;
             row order as printed by git grep)

EV-13 cmd:   grep -c 'Use the `moai cc` / `moai glm` launchers for this repository' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md
      exit:  1
      out:   .claude/skills/moai/workflows/goal.md:0
             internal/template/templates/.claude/skills/moai/workflows/goal.md:0
             (AC-GDD-008 RED cell, sites #3/#5 — the pinned replacement string does
             not exist yet)

EV-14 cmd:   grep -c 'Use `moai codex` for the worktree session' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md
      exit:  1
      out:   .claude/skills/moai/workflows/goal.md:0
             internal/template/templates/.claude/skills/moai/workflows/goal.md:0
             (AC-GDD-008 RED cell, sites #4/#6 — the pinned replacement string does
             not exist yet)
```

## §D Given-When-Then Scenarios

**AC-GDD-001** — zero live guidance references
- Given the card's four-file edits are complete and committed, **when** `grep -rn "moai gpt" AGENTS.md .claude/skills internal/template/templates` runs at the final HEAD, **then** stdout is empty and the exit code is 1.
- Given the same tree, **when** `grep -c "moai gpt" README.md README.ko.md README.ja.md README.zh.md` runs, **then** all four locales report count `1` with exit 0 (the withdrawal notes survive as the positive control).

**AC-GDD-002** — verb-table row exactness
- Given the row edit landed in both files, **when** `grep -n "Explicit Claude or GLM session launchers" AGENTS.md internal/template/templates/AGENTS.md.tmpl` runs, **then** it prints exactly two rows (one per file) and exits 0.
- Given the same tree, **when** `grep -c "moai gpt" AGENTS.md` runs, **then** there is no match (exit 1), and the pre-existing `moai codex` row remains the sole GPT-session launcher entry (this card does not touch it).

**AC-GDD-003** — source ↔ template byte parity
- Given both `goal.md` copies were edited in lockstep, **when** `diff -q .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md` runs, **then** it prints nothing and exits 0; equivalently, `go test ./internal/template/ -run '^TestGoalAutoWorkflowContractAndMirrorParity$'` passes.

**AC-GDD-004** — embed refresh
- Given the template edits are in the tree, **when** `make build` runs at run-phase entry, **then** it exits 0 including its `agents-emit-check` prestep, and the regenerated embedded catalog carries the re-pointed text; the output is observed and recorded then (pre-work baseline unmeasured — the build mutates the tree).

**AC-GDD-005** — affected-package tests
- Given the six doc edits and the contract-token swap have landed together (the intermediate state is red by design — the suite requires the token the :189 edit removes), **when** `go test ./internal/template/...` runs, **then** it exits 0 with `TestGoalAutoWorkflowContractAndMirrorParity` among the passing tests; the honest green baseline is observed and recorded at run phase (pre-work green is drift-fed).

**AC-GDD-006** — doc-only contribution
- Given the card's commits are on the branch, **when** `git merge-base develop HEAD` yields BASE and `git diff --name-only BASE..HEAD` runs, **then** the output names exactly `AGENTS.md`, `.claude/skills/moai/workflows/goal.md`, `internal/template/templates/.claude/skills/moai/workflows/goal.md`, `internal/template/templates/AGENTS.md.tmpl`, and `internal/template/goal_auto_workflow_test.go` — exactly one `*_test.go` and zero other `*.go` entries.

**AC-GDD-007** — no new `moai gpt` string introduced
- Given the card's final HEAD, **when** `git grep -c "moai gpt" <final-HEAD> -- AGENTS.md .claude/skills internal/template/templates` runs, **then** it prints nothing and exits 1 — no occurrence, pre-existing or newly introduced, survives in the committed guidance tree. RED cell: EV-12 (four count-rows totalling 6 at the pin).

**AC-GDD-008** — positive replacement text at sites #3-6 (binds REQ-GDD-003/004; a lockstep wrong-token mutant now fails)
- Given the goal.md edits landed in both twins, **when** `grep -c 'Use the `moai cc` / `moai glm` launchers for this repository' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md` runs, **then** both twins report count `1` and the exit code is 0.
- Given the same tree, **when** `grep -c 'Use `moai codex` for the worktree session' .claude/skills/moai/workflows/goal.md internal/template/templates/.claude/skills/moai/workflows/goal.md` runs, **then** both twins report count `1` and the exit code is 0.
- RED cells: EV-13/EV-14 (both pinned strings absent pre-work — two `:0` count-rows, exit 1).

## §E Edge Cases

- **E-1 — legitimate occurrences.** `moai gpt` appears by design in the README withdrawal notes, `.moai/specs/*` history, and commit messages. Sweeps are scoped to the four guidance roots; EV-9 is the discriminator (README stays at 1 per locale while the guidance roots go to zero).
- **E-2 — line-number divergence between twins.** The drifted row sits at `AGENTS.md:322` and `AGENTS.md.tmpl:324` (the template render source carries two extra header lines). Criteria key on row text, never line numbers.
- **E-3 — grep exit-code polarity.** A zero-hit sweep exits 1. Reading exit 1 as failure flips this card's verdicts; the PASS shape is empty stdout + exit 1 (§A).
- **E-4 — deploy-path recurrence.** `moai update` renders `AGENTS.md` from `AGENTS.md.tmpl`. Editing only the local row would let the next deploy re-introduce the drift — which is why site #2 (`AGENTS.md.tmpl:324`) is in scope despite the lane's original inventory.
- **E-5 — template-render safety.** The edited row and sentences carry no Go-template directives; the swap cannot alter rendering behavior.
- **E-6 — swap sequencing.** The parity suite must not gate between the doc edits and the token swap: in the intermediate state it fails with `auto workflow missing contract token "moai gpt"` by design (audit D1). They land together (plan.md M1+M2+M3).

## §F Quality Gate Criteria (TRUST 5, doc-only mapping)

| Pillar | This card |
|--------|-----------|
| Tested | AC-GDD-005 (owning-package suite), AC-GDD-003 (parity instrument) |
| Readable | Minimal-diff launcher-token swaps; surrounding prose untouched |
| Unified | Replacement row matches the neighboring verb-table row format exactly |
| Secured | No executable surface touched; no input/validation change (n/a for prose) |
| Trackable | Conventional Commits with card `t1406` in the body; evidence in progress.md §E.2 |

## §G Traceability, Indirect Verification, Closure Gates

REQ ↔ AC map:

| REQ | ACs |
|-----|-----|
| REQ-GDD-001 | AC-GDD-001, AC-GDD-007 |
| REQ-GDD-002 | AC-GDD-002, AC-GDD-007 |
| REQ-GDD-003 | AC-GDD-001 (site #3 zero-hit), AC-GDD-003, AC-GDD-008 |
| REQ-GDD-004 | AC-GDD-001 (site #4 zero-hit), AC-GDD-003, AC-GDD-008 |
| REQ-GDD-005 | AC-GDD-003, AC-GDD-005 |
| REQ-GDD-006 | AC-GDD-004 |
| REQ-GDD-007 | AC-GDD-005 |
| REQ-GDD-008 | AC-GDD-006 |

- Indirect verification: AC-GDD-006 (zero `*.go` in the diff) indirectly proves CLI behavior is unchanged — the embed refresh re-embeds prose only. AC-GDD-003/005 verify parity through a second, independent instrument (the dedicated test) beyond raw `diff`.
- Closure gates: `draft → in-progress` — manager-develop, first run-phase commit; `implemented → completed` — manager-docs, the single sync commit, which populates `sync_commit_sha` in progress.md §E.4.

## §H Forward-looking Checks and Definition of Done

- Forward-looking (deliberately out of scope here, spec.md §5): a CI guard failing any future `moai gpt` occurrence in the four guidance roots — a follow-up candidate card, since it is a code change.
- Definition of Done: all eight ACs green at the final HEAD; the plan.md §E matrix fully observed with verbatim outputs recorded in progress.md §E.2; card reported merge-ready to the leader for the develop integration window.
