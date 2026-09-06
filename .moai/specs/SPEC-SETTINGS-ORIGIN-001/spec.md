---
id: SPEC-SETTINGS-ORIGIN-001
title: "Dirty .claude/settings.json working-copy writer attribution — code-path inventory, external inflow, and recurrence prevention"
version: "0.2.0"
status: completed
created: 2026-09-05
updated: 2026-09-06
author: manager-spec (card t487)
priority: P2
phase: "v3.2.0 target"
module: repo-wide (read-only investigation)
lifecycle: spec-lite
tags: "forensics, settings-json, worktree, t487, t480-followup"
tier: S
related_specs: []
---

# SPEC-SETTINGS-ORIGIN-001 — Dirty settings.json Writer Attribution

## HISTORY

| Date | Event |
|---|---|
| 2026-09-05 | Created (card t487, plan-phase). Follow-up to t480 (verdict at `.moai/reports/t480/verdict.md`): t480 disproved the "four-block loss" premise and identified no writer; this SPEC owns that residual Gap 1. |
| 2026-09-05 | v0.2.0 — plan-audit PASS 0.80 conditioned fixes applied: D1 (AC-008 added), D2 (REQ-001/AC-001 scope extended to `scripts/`, `.claude/workflows/`, `.github/workflows/`), D3 (plan §B extraction pattern), D4 (t485 verdict path annotated primary-checkout-only), D7 (`--no-optional-locks` in sweep commands). Audit verdict: `.moai/reports/t487/plan-audit.md`. |

## §1 Problem Statement

During the t452 integration window (worktree existed 2026-09-03 13:40 → merge 2026-09-04 03:52 KST), the t452 worktree's tracked `.claude/settings.json` working copy was found dirty. Card t480 preserved and analyzed it: the dirty copy is an **older, thinner rendering** than every measured repository state (fewer `permissions.ask` entries, older hook wiring, one unique narrow matcher line), and t480 eliminated six origin candidates — yet **the writer of that dirty copy remains unidentified**. This SPEC is the follow-up investigation.

The preserved artifact of record (re-verified intact 2026-09-05, md5 `b669972dc738d1bf925281dcc90f152e`):

- `.claude/worktrees/t480/.moai/reports/t480/preserved-copies/settings.json.worktree-dirty` (24,667 B)
- Reference copy: `settings.json.develop` (md5 `568a3d3a32a360731d1d02f686f27d24`) — byte-identical to the develop-committed version at the time.

### Position relative to t485 ruling C4

t485 verdict C4 ruled that **config-file-STATE digging is a dead axis for WRITER attribution of transient writes** — a writer process is unrecoverable from repository/config state alone. This SPEC respects that ruling and differs from it in exactly one respect, stated here once: **the t487 artifact PERSISTED**, so state forensics *of the artifact itself* (shape fingerprinting, distribution sweep, normalized diff) is valid. What remains dead is expecting repository/config state alone to **name the writer process**. The live axes are: (a) code-path inventory — which writers *could* exist in this repository, (b) distribution analysis — is the dirty shape systematic or one-off, and (c) external-inflow enumeration. Any live-observation proposal in Q3 must be **process-level**, never config-state-level.

### Dirty-copy shape fingerprint (measured baseline — lane-8, 2026-09-05, on the preserved copy)

| Marker | Dirty copy | Current template (`internal/template/templates/.claude/settings.json.tmpl`) |
|---|---|---|
| Top-level key order | `$schema, respectGitignore, cleanupPeriodDays, skillListingBudgetFraction, env, attribution, permissions, hooks, statusLine, outputStyle, showThinkingSummaries` (11 keys) | wholly different order: `$schema, hooks@3, statusLine@395, skillListingBudgetFraction@404, showThinkingSummaries@405, cleanupPeriodDays@406, model@407, outputStyle@408, env@409, permissions@420, attribution@600, respectGitignore@605, includeGitInstructions@606, plansDirectory@607` (14 keys) |
| Keys present in template but ABSENT in dirty copy | — | `model`, `includeGitInstructions`, `plansDirectory` |
| `permissions.ask` | `["Bash(sudo:*)"]` exactly (1 entry) | 6 entries (`rm, sudo, chmod, chown, .env, .env.*`) |
| PostToolUse matcher | narrow form `"Write|Edit|MultiEdit"` (line 314) | extended `"Write|Edit|MultiEdit|EnterWorktree|ExitWorktree"` |
| `status-transition-ownership` mentions | exactly 1 (the conditional 3-entry + `async`/`if` form is absent) | present, conditional 3-entry + `async`/`if` (t216 parity) |

t480's "status-transition 배선 부재" wording is corrected by lane-8's re-measurement: wiring is not wholly absent — a **single** mention exists; the conditional 3-entry + async/if form is what is absent.

### Hypothesis tree (to be tested, not concluded)

- **H1 — Claude Code runtime accumulated/edited write**: the key-order fingerprint (env/attribution/permissions mid-file, keys appended in CC-era order) resembles a file Claude Code runtime rewrote/normalized across upgrades more than a single moai template render (current template order is wholly different). *Status: hypothesis; discriminates via H4 distribution + Q2 CC-write signature research.*
- **H2 — an old moai binary or old template render**: t480 eliminated the 5 measured release templates **on the ask-list axis only**; full-file key-order vs older releases is NOT yet measured. *Status: open axis, explicitly not re-running the ask-list comparisons.*
- **H3 — manual hand-edit**: formatting style / key-addition order signature. *Status: hypothesis.*
- **H4 — systematic writer still active**: if the worktree sweep finds the same dirty shape elsewhere, the writer is systematic (still live); if t452 is unique, one-off. *Status: the highest-information-gain measurement in this SPEC.*

## §2 Requirements (GEARS)

- **REQ-001** (Q1 — inventory): The investigation shall exhaustively inventory every code path in this repository that writes `.claude/settings.json` — including `internal/`, `pkg/`, `cmd/`, template `.json.tmpl` sources, hook scripts under `.claude/hooks/`, launcher flows (`moai cc` / `moai glm` / `moai cg`), the `Makefile`, `scripts/`, `.claude/workflows/`, and `.github/workflows/` — citing file:line for each; the cross-check search shall be bounded only by the repository root, not by any enumerated subdirectory list.
- **REQ-002** (Q1 — reconciliation): For each inventoried code path, the investigation shall record whether that path could produce the preserved dirty shape (§1 fingerprint: key order, `ask` list, narrow matcher, single status-transition mention), with per-path reasoning.
- **REQ-003** (SWEEP — distribution): The investigation shall sweep every worktree in `git worktree list` (the primary checkout is entry 1 of that listing) read-only, checking whether each working-copy `.claude/settings.json` is dirty vs its HEAD (`git --no-optional-locks -C <wt> status --porcelain -- .claude/settings.json` — the `--no-optional-locks` form is mandatory: plain `git status` takes index write locks in trees other lanes may be using), and for every dirty one shall record path + md5 + the three shape markers (ask-list length, matcher form, key order).
- **REQ-004** (Q2 — conditional external inflow): **When** REQ-001/REQ-002 find no producer of the dirty shape, the investigation shall enumerate external inflow candidates with evidence — other checkouts/projects on this machine, old binaries on disk, the TypeScript moai-adk predecessor (npm historical templates, if fetchable), Claude Code runtime writes (settings migration/normalization on CC upgrade), and manual hand-edit signature.
- **REQ-005** (Q3 — recurrence prevention): The investigation shall produce exactly one grounded recurrence-prevention recommendation, and the recommendation shall respect t485 ruling C4 (no config-file-digging axis; any live-observation proposal must be process-level).
- **REQ-006** (verdict): The investigation shall close with an evidence verdict at `.moai/reports/t487/verdict.md` in the 5-section evidence-bearing format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) per `AGENTS.md` §1.
- **REQ-007** (evidence integrity): Every claim the investigation makes shall carry the command run plus its observed output; empty output shall not be reported as zero; absence of a signal shall not be reported as evidence in either direction.
- **REQ-008** (baseline reuse): The investigation shall build on t480's measured eliminations without re-running them, and shall treat the lane-8 fingerprint observations (§1) as hypotheses in the §1 tree, not as conclusions.

## §3 Acceptance Criteria

Acceptance criteria (Given-When-Then, `AC-001`..`AC-008`) live in [`acceptance.md`](./acceptance.md) — emitted per the t487 lane deliverable shape. (Tier S's minimal set is spec.md + plan.md with AC inline; this SPEC emits the additive third file because the dispatch explicitly enumerates it.)

## §4 Constraints [HARD — binding on run-phase]

- **Byte-preserve every dirty settings.json found**: record path + md5 only. Never modify, revert, or delete one (user environment values — tokens, absolute paths, tmux pane ids — may be mixed in).
- **The worktree sweep is READ-ONLY.** No background load of any kind (many lanes run concurrently on this machine). No full-suite test runs (lane-local scoping only; and this card expects no test runs at all).
- **Verification-claim integrity**: every claim carries command + observed output; empty output is not zero; absence-of-signal is not evidence either direction.
- **Do not dispose of any worktree. Do not push.**
- Do not re-run t480's six origin-elimination measurements (tracked full history, pre-generator state `e21d85d8d^`, 5 release templates on the ask axis, global/local settings ask-null, installed-binary version) — they are the accepted baseline (REQ-008).

## §5 Scope

In scope: read-only investigation only — the three card questions (Q1 code-path inventory, Q2 external inflow, Q3 one recurrence-prevention recommendation) plus the SWEEP worktree distribution measurement; one evidence verdict artifact; SPEC artifacts themselves. **No production code changes.** Q3 produces a RECOMMENDATION ONLY — implementing it is a separate card.

Expected close shape: run-phase closes on the evidence verdict (`.moai/reports/t487/verdict.md`); **sync phase expected N/A** (evidence-only merge, t480 precedent — no docs/CHANGELOG rotation for a no-code investigation).

### Out of Scope — production code changes

- Any modification to `internal/`, `pkg/`, `cmd/`, templates, or hook scripts, including implementing the Q3 recommendation (that is a separate card, to be queued by the lead if accepted).
- Any change to any `.claude/settings.json` anywhere (see §4 byte-preservation).

### Out of Scope — re-running t480's eliminations

- Re-executing t480's six origin-elimination measurements (tracked-file history, pre-generator state, release-template ask-list comparisons, global/local ask-null check, installed-binary version check). They are the accepted baseline.

### Out of Scope — destructive or state-mutating operations

- Worktree disposal of any tree (including trees found dirty by the SWEEP), any `git push`, any branch-state change, any background load (see §4).

### Out of Scope — sync-phase documentation rotation

- API docs, codemaps, CHANGELOG, README rotation — evidence-only close per t480 precedent; the verdict file is the terminal artifact.

## §6 Dependencies and References

- t480 verdict (baseline + preserved copies): `.moai/reports/t480/verdict.md` (worktree t480 copy is canonical; preserved copies re-verified 2026-09-05).
- t485 verdict C4 (writer-attribution dead-axis ruling): `.moai/reports/t485/verdict.md` — **primary-checkout copy only**; the path does not resolve inside the t487 worktree, so a run-phase executor reads it at `/Users/goos/MoAI/moai-adk-go/.moai/reports/t485/verdict.md`.
- `AGENTS.md` §1 (evidence-bearing report format), `verification-claim-integrity.md` (no-unobserved-claim invariant).
- `CLAUDE.md` §2.3 (`moai update` deletes local-only files under managed roots — relevant as a *possible* writer class in Q1, not a premise).
