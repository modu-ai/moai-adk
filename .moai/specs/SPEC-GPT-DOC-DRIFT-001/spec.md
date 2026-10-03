---
id: SPEC-GPT-DOC-DRIFT-001
title: "moai gpt launcher doc-CLI drift re-pointing"
version: "0.1.1"
status: completed
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "docs, launcher, drift, gpt-withdrawal, template-parity"
tier: S
---

# SPEC-GPT-DOC-DRIFT-001 — moai gpt launcher doc-CLI drift re-pointing

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-10-03 | 0.1.0 | manager-spec | Initial plan-phase draft (Tier S, card t1406, worktree `WT-gpt-launcher-doc-drift`). Drift inventory measured at 6 sites / 4 files — one site beyond the lane's 5-site inventory: `internal/template/templates/AGENTS.md.tmpl:324`, the rendered twin of the AGENTS.md verb-table row. |
| 2026-10-03 | 0.1.1 | manager-spec | Plan-audit iteration 1 FAIL (0.63 < 0.75) — delta applied per lane decision record (`.moai/reports/t1406/verdict.md` § Plan 감사 1차): D1 scope amendment admits one Go change (contract-token swap `"moai gpt"` → `"moai codex"` at `internal/template/goal_auto_workflow_test.go:52`); D2 positive-content AC-GDD-008; D3 verbatim EV-1 + AC-GDD-007 RED cell; D4 row-distance fix; D5 green-baseline observation duties; D6 §B check-command header fix. |

## 1. Background

### 1.1 The withdrawal is deliberate and recorded

- Commit `2d25a88eb` — "feat(cli)!: withdraw the moai gpt gateway and GPT-in-Claude-Code path (card t857)", 2026-09-17, 314 files / -46,832 lines; operator goal: "remove every moai-gateway code path".
- The CLI redirects its own users. `internal/cli/launcher.go:135` returns `moai gpt is removed — run GPT models through their native harness: moai codex` (measured 2026-10-03 on this tree; the delegation cited lines 140-142 — same message, coordinates corrected to :135 after measurement).
- The README 4-locale set documents the withdrawal as intentional, one note per locale: `README.md:310`, `README.ko.md:305`, `README.ja.md:306`, `README.zh.md:305` (measured 2026-10-03).

### 1.2 The drift

Doc-CLI drift: live guidance prose still presents `moai gpt` as a usable launcher while the CLI surface no longer has one. A reader following the drifted guidance hits the CLI's own removal error.

### 1.3 Drift inventory (measured 2026-10-03, tree `1e2151a380a5dd0d76efd8f1740a21f32c682d2f`)

Sweep: `grep -rn "moai gpt" AGENTS.md .claude/skills internal/template/templates` → 6 rows, exit 0. Exactly these six sites, no others in the guidance roots:

| # | File:line | Drift | Re-point target |
|---|-----------|-------|-----------------|
| 1 | `AGENTS.md:322` | §11 verb-table row lists `moai gpt` among the explicit session launchers | drop `moai gpt` from the row |
| 2 | `internal/template/templates/AGENTS.md.tmpl:324` | rendered template twin of row #1 (identical row text) | identical drop |
| 3 | `.claude/skills/moai/workflows/goal.md:52` | infinite-goal subsection: "Use the `moai gpt` launcher for this repository" | `moai cc` / `moai glm` |
| 4 | `.claude/skills/moai/workflows/goal.md:189` | auto-mission section: "Use `moai gpt` for the worktree session" | `moai codex` |
| 5 | `internal/template/templates/.claude/skills/moai/workflows/goal.md:52` | byte-parity mirror of #3 | same as #3 |
| 6 | `internal/template/templates/.claude/skills/moai/workflows/goal.md:189` | byte-parity mirror of #4 | same as #4 |

Site #2 was absent from the delegating lane's inventory (it checked for a template file named `AGENTS.md`; the twin is the `.tmpl` render source). It is in scope because the Template-First rule makes the template tree the deploy source: leaving it drifted re-introduces the row into user projects on the next template deploy.

### 1.4 Determined approach — Option B (doc re-pointing)

Recorded as decided by the delegating lane, evidence-based; this SPEC does not re-litigate it:

1. `moai gpt` was deliberately withdrawn (commit `2d25a88eb`, card t857).
2. The CLI redirects its own users to `moai codex` (`internal/cli/launcher.go:135`).
3. The README 4-locale set documents the withdrawal as intentional.
4. Re-implementing would reverse a recorded operator decision and merely duplicate `moai codex`.

Per-site re-point rationale:

- **Sites #3/#5 → `moai cc` / `moai glm`.** The surrounding advice is a Claude Code runtime environment variable (`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`), and the launcher's automatic cap inject for an armed infinite goal (`injectStopHookBlockCapForGoal`, called at `internal/cli/launcher.go:817`, implemented in `internal/cli/launcher_blockcap_infinite.go`, SPEC-INFINITE-GOAL-001 REQ-2) exists only on the unified cc/glm launch path. Corroborated this run: `grep -l BlockCap internal/cli/codex_stop_chain.go internal/cli/codex_launcher.go` → no match, exit 1. Re-pointing this site to `moai codex` would be technically incoherent.
- **Sites #4/#6 → `moai codex`.** The clause names a GPT-model session for the autonomous worktree mission; the goal engine has codex parity (`internal/cli/codex_stop_chain.go`, `internal/cli/codex_goal_parity_test.go` both present). No env coupling in that sentence.
- **Sites #1/#2 → drop the `moai gpt` token.** The GPT-session launcher entry is already the `moai codex` row three lines below (`AGENTS.md:325`); duplicating it in the cc/glm row would be wrong.

### 1.5 Not drift (protected surfaces)

- The README withdrawal notes themselves (they mention `moai gpt` in order to document its removal).
- `.moai/specs/*` historical records (including this SPEC's own text).
- `.agents/` emissions (zero `moai gpt` hits measured this run).
- `docs-site/` (zero `moai gpt` hits measured this run).

### 1.6 Contract-token coupling (D1 amendment)

`internal/template/goal_auto_workflow_test.go:52` requires the literal token `"moai gpt"` inside the goal workflow's auto-mission section — the slice from the "## `/moai goal --auto`" heading to the next "##" heading — and errors with `auto workflow missing contract token` when it is absent; a separate assertion at `goal_auto_workflow_test.go:61-63` errors when that section contains `"moai cc"`. The pre-work suite green is therefore fed by the drift itself: the drifted goal.md:189 line supplies the required token. Per the lane decision record, the card swaps the required token to `"moai codex"` — the :189 replacement introduces exactly that token, the :61-63 prohibition stays untouched and remains satisfied, and the :52 edit sits outside the scanned section's only launcher-token line's replacement target ("moai codex", not "moai cc").

## 2. Requirements (GEARS)

- **REQ-GDD-001** (Ubiquitous): The launcher guidance surface — `AGENTS.md`, `internal/template/templates/AGENTS.md.tmpl`, `.claude/skills/moai/workflows/goal.md`, and `internal/template/templates/.claude/skills/moai/workflows/goal.md` — shall reference only launchers present in the current CLI surface (`moai cc`, `moai glm`, `moai codex`) and shall not present the withdrawn `moai gpt` launcher as a live capability.
- **REQ-GDD-002** (Event-driven): **When** a reader consults the launcher verb-table row for explicit session launchers, the row shall read `| `moai cc` / `moai glm` | Explicit Claude or GLM session launchers |` in both `AGENTS.md` (§11) and `internal/template/templates/AGENTS.md.tmpl`, shall carry no `moai gpt` entry, and the pre-existing `moai codex` row shall remain the sole GPT-session launcher entry.
- **REQ-GDD-003** (Event-driven): **When** the infinite-goal subsection of the goal workflow (`--max-turns 0`) names a launcher for this repository, it shall reference the Claude Code launchers (`moai cc` / `moai glm`) — the launch path that injects `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` for an armed infinite goal (`internal/cli/launcher.go:817`, `internal/cli/launcher_blockcap_infinite.go`, SPEC-INFINITE-GOAL-001 REQ-2) — and shall not reference `moai gpt`.
- **REQ-GDD-004** (Event-driven): **When** the auto-mission section of the goal workflow names a launcher for a GPT-model worktree session, it shall reference `moai codex`, whose goal engine carries codex parity (`internal/cli/codex_stop_chain.go`, `internal/cli/codex_goal_parity_test.go`).
- **REQ-GDD-005** (State-driven): **While** the goal workflow exists in both the local skill tree and the template tree, `.claude/skills/moai/workflows/goal.md` and `internal/template/templates/.claude/skills/moai/workflows/goal.md` shall remain byte-identical after every edit of this card.
- **REQ-GDD-006** (Capability gate): **Where** the embedded template catalog is regenerated (`make build`, `agents-emit-check` prestep included), the build shall pass with the re-pointed guidance in place.
- **REQ-GDD-007** (Event-detected): **When** the template package test suite runs (`go test ./internal/template/...`) after the six doc edits and the contract-token swap of REQ-GDD-008 have landed together, it shall pass — `TestGoalAutoWorkflowContractAndMirrorParity` (`internal/template/goal_auto_workflow_test.go:10`) included.
- **REQ-GDD-008** (Ubiquitous): The card's change set shall remain doc-plus-contract-token: its only Go change shall be the one-line required-token swap `"moai gpt"` → `"moai codex"` at `internal/template/goal_auto_workflow_test.go:52` (inside the required-tokens slice only); every other `*.go` file shall remain untouched — the `"moai cc"` prohibition at `goal_auto_workflow_test.go:61-63` included — and the change set shall not modify the README 4-locale withdrawal notes, `.moai/specs/` historical records, `.agents/` emissions, `docs-site/`, or any launcher implementation.

## 3. Acceptance Criteria

Full scenarios, commands, and RED-now evidence cells live in `acceptance.md`. Index:

| AC id | Criterion (one line) | Verifies |
|-------|----------------------|----------|
| AC-GDD-001 | Zero live `moai gpt` references across the four guidance files; README positive control intact | REQ-GDD-001 |
| AC-GDD-002 | Re-pointed verb-table row present verbatim in `AGENTS.md` + `AGENTS.md.tmpl` | REQ-GDD-002 |
| AC-GDD-003 | Source ↔ template `goal.md` byte-identical after the edit | REQ-GDD-005 |
| AC-GDD-004 | `make build` passes (embed refresh, `agents-emit-check` prestep included) | REQ-GDD-006 |
| AC-GDD-005 | `go test ./internal/template/...` green after the doc edits + token swap land together | REQ-GDD-005, REQ-GDD-007 |
| AC-GDD-006 | Merge-base diff file list is exactly 5 files — the 4 guidance files + `goal_auto_workflow_test.go`; exactly one `*_test.go`, zero other `*.go` | REQ-GDD-008 |
| AC-GDD-007 | No `moai gpt` occurrence in the committed guidance tree at the final HEAD | REQ-GDD-001, REQ-GDD-002 |
| AC-GDD-008 | Positive replacement text present at sites #3-6 (both goal.md twins, both pinned strings) | REQ-GDD-003, REQ-GDD-004 |

Note: `acceptance.md` is carried at explicit lane direction. The Tier S default artifact set is spec.md + plan.md with AC inline; this §3 index keeps that inline convention while `acceptance.md` holds the verification layer.

## 4. Constraints

- Doc plus one contract-token: the only Go change is the one-line required-token swap at `internal/template/goal_auto_workflow_test.go:52` — a test assertion, not launcher implementation. The CLI binary's behavior is untouched; the embed refresh re-embeds prose only.
- Template-First cycle (AGENTS.local.md §2): template-tree edits land before `make build`; local-tree edits follow in lockstep.
- Minimal edits: swap the launcher token only; no surrounding-sentence rewrites.
- Byte-parity lockstep between source and template `goal.md`.
- The `AGENTS.md` row and the `AGENTS.md.tmpl` row keep identical text (the line carries no template variables).
- grep exit-code semantics: a zero-hit sweep exits 1 — the sweep AC reads empty stdout + exit 1 as PASS; exit 0 means hits remain.
- Test-token carve-out (D1): the `"moai cc"` prohibition at `goal_auto_workflow_test.go:61-63` stays untouched; it remains satisfied because the :189 replacement introduces `moai codex`, not `moai cc`, and the :52 edit sits outside the scanned auto section.

## 5. Out of Scope

### Out of Scope — README withdrawal notes
- The four locale notes (`README.md:310`, `README.ko.md:305`, `README.ja.md:306`, `README.zh.md:305`) that mention `moai gpt` while documenting its withdrawal are protected; they serve as the sweep's positive control.

### Out of Scope — historical and generated surfaces
- `.moai/specs/*` historical records (including this document), `.agents/` emissions, and `docs-site/` are not edited by this card.

### Out of Scope — template-neutrality debt in goal.md:189
- The same sentence re-pointed by REQ-GDD-004 reads "…and preserve the repository's existing manager ownership and local-develop integration rules" — dev-repo vocabulary inside a shipped template. Observed and recorded; deliberately NOT fixed by this card (it is a template-neutrality concern for a separate decision, and bundling it here would grow a doc-only card's blast radius).

### Out of Scope — Go code and re-implementation
- Launcher implementation stays out of scope: no CLI behavior change, no `moai gpt` gateway re-implementation. The single admitted Go change is the contract-token swap at `internal/template/goal_auto_workflow_test.go:52` (lane decision record); every other `*.go` file — the `"moai cc"` prohibition at :61-63 included — is forbidden to this card.

### Out of Scope — mechanical drift guard (follow-up candidate)
- A CI check that fails any future `moai gpt` occurrence in the four guidance roots is a candidate follow-up card; it is a code change and does not belong to this doc-only card.

## 6. Delivery

- This repository integrates `WT-*` card branches into local `develop` through the git-flow lane protocol (AGENTS.local.md §4.1): no PR from this card. The worktree branch merges into local `develop` in the serial integration window; the remote push of `develop` is the leader's batch duty.
- Plan-phase commit subject (canonical pattern, spec-frontmatter-schema § Status Transition Ownership Matrix): `feat(SPEC-GPT-DOC-DRIFT-001): plan-phase artifacts (S, 4 artifacts)` — 4 because `acceptance.md` rides at explicit lane direction beyond the Tier S default set of 2.
- Card traceability: card id `t1406` in the commit body; evidence lives in `.moai/specs/SPEC-GPT-DOC-DRIFT-001/` (progress.md §E.2 carries the run-phase outputs).

## 7. Cross-references

- SPEC-INFINITE-GOAL-001 — REQ-2 Stop-hook block-cap inject this card's site #3 rationale leans on
- Commit `2d25a88eb` (card t857) — the withdrawal this drift trails
- `internal/template/goal_auto_workflow_test.go` — `TestGoalAutoWorkflowContractAndMirrorParity` (byte-parity instrument)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — artifact ownership and statelessness
- AGENTS.local.md §2 (Template-First) · §4.1 (git-flow integration chain)
