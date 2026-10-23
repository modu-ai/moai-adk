---
id: SPEC-SPEC-LINT-REGRESSION-001
title: "Implementation plan — spec-lint baseline regression repair (status record + vacuous-assertion verification + green gate)"
version: "0.3.0"
created: 2026-10-23
updated: 2026-10-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, baseline-gate, status-transition, lifecycle-record, vacuous-assertion, regression-repair"
---

# plan.md — SPEC-SPEC-LINT-REGRESSION-001

> Stateless on the status axis (no `status:` field). Lifecycle lives in `spec.md`.

## §A — Context

### A.1 Task shape

A one-commit status-record repair on one SPEC's frontmatter plus its HISTORY record (the executed form of the leader disposition of 2026-10-23, which simplified the original two-commit design — landed at `35cccfc217`), a verification pass over three already-repaired SPEC files, and a gate re-measurement. No Go source changes; no lint-rule changes; no baseline changes. The load-bearing deliverable is the canonical recorded pair at the end state (`implemented → completed`, matrix row 3) — the mechanism that turns the main-push gate from red to green.

### A.2 Codebase anchors (measured at worktree HEAD `c3c86e795d`)

- **Canonical edge set**: `internal/spec/lint_transition.go:65-77` — `canonicalStatusTransitions` with matrix-row citations; row 7 is `{"completed", "in-progress"}`. Terminal targets (`superseded|archived|rejected`) are admitted from any status at `:80-100`; the legacy `planned` token is tolerated on either side — neither is a repair target here.
- **The walk that judges the pair**: `internal/spec/lint_ownership.go:204-247` — `git log -<window> --follow -p` per SPEC path, newest-first, first commit with a status-line delta wins (`extractStatusDelta`, `:313`). Consequence: each forward commit creates the pair the rule next judges; the repair sequence in REQ-SLR-002 is shaped by exactly this.
- **The advisory boundary**: `internal/spec/lint.go:412-430` (`applyEraDemotion` — grandfather era OR terminal status marks every warning advisory; active modern non-terminal SPECs pass through) and `:1766-1771` (`terminalStatusEnum` = superseded, archived, rejected, completed). The `"downgraded to warning"` annotation rides only the error-to-warning branch (`:420-423`); warning demotions carry no annotation — the measurement file's annotated lines are therefore NOT the advisory discriminator for this rule (the parser JSON is).
- **The ownership rule is silent on the end state**: `internal/spec/lint_ownership.go:64-95` — `expectedOwnerForTransition("completed", "in-progress")` returns ownerNone, and the ownership rule exempts ownerNone pairs, so the repair commits need no specific Authored-By-Agent trailer to stay clean; manager-spec owns row 7 per the matrix and should sign the amendment commit accordingly.
- **The defect record**: `<ID>` = `.moai/specs/SPEC-PROGRESS-RECORD-IO-001/`. spec.md frontmatter reads `status: implemented` (line 5). Close: commit `8d10007ca9` (2026-10-08, "sync-phase artifacts — 3-phase close"); revert: commit `fa1f04b9f4` (2026-10-09, "revert status to implemented pending AC-CI-007 CI evidence"). progress.md §E.4 carries `sync_commit_sha: 8d10007ca9804eeb58bcb0a34aca922989603198` and `sync_status: sync-ready-pending-CI`; the AC-CI-007 row reads PENDING (CI-only).
- **The landed VTA repair**: commit `17a3b98de2` (PR #1859, card t1914) — 5 files, 14 changed line-pairs: `.moai/specs/SPEC-FACTORY-RECORD-REPAIR-001/acceptance.md` (6), `.moai/specs/SPEC-PROGRESS-RECORD-IO-001/acceptance.md` (4) + `plan.md` (2), `.moai/specs/SPEC-GLM-JEV-KEY-001/acceptance.md` (1) + `plan.md` (1). Its commit message names the STI +1 as "not file-repairable and escalated separately" and records the selector-list regeneration (28 names; two named tests no longer exist in source).
- **The gate**: `.github/workflows/spec-lint.yml` runs the lint with `--baseline .moai/spec-lint-baseline.json`; the checked-in baseline records only `{"TierArtifactMissing": 4}` (tree `d5638ba81`, 2026-09-24).

### A.3 Measurement census (all on tree `c3c86e795d`, linter built from this tree)

- Full-corpus JSON lint: 8557 warnings, 0 errors; non-advisory warnings = TierArtifactMissing 4 (recorded 4, no delta) + StatusTransitionInvalid 1 (recorded 0, delta +1). Exactly one non-advisory StatusTransitionInvalid corpus-wide; targeted JSON reports it `advisory=false`.
- StatusTransitionInvalid warnings: 112 lines across 111 SPECs; 110 SPECs protected (terminal status or grandfather era). The 15 modern-dated non-terminal candidates: 14 lack V3R6 progress.md markers (H-1/H-2 grandfather); only `<ID>` carries §E.2 + §E.4 + sync_commit_sha (H-4 modern, V3R6).
- The 4 newest-landing SPECs (CARD-METADATA-VERBS, TRUNK-BACKUP-SYNC, INTEGRATION-STALE-BASE, CARD-ELIGIBLE-PULL) have zero StatusTransitionInvalid — canonical transitions from day one.
- Release-branch staleness: all six PR heads (#1828 #1843 #1844 #1848 #1853 #1858) show the identical 14-line-pair numstat against `c3c86e795d` for the three SPEC directories; the tree's own lint reports zero VacuousTestAssertion findings for the three SPECs.
- Preserved lane measurement: `.moai/reports/t1929/baseline-lint-measurement.txt` (8788 lines; verdict lines 8783-8785).

### A.4 PRESERVE list (the repair touches none of these)

- `.moai/spec-lint-baseline.json` — byte-identical before and after (REQ-SLR-007 asserts the empty diff).
- Everything under `internal/spec/` — the rules are correct.
- `<ID>/progress.md` — §E.4's pending-CI record stays verbatim (REQ-SLR-004).
- `<ID>/acceptance.md` and `<ID>/plan.md` — already repaired at `17a3b98de2`; verification only.
- All recorded git history — forward commits only.
- SPEC-FACTORY-RECORD-REPAIR-001 and SPEC-GLM-JEV-KEY-001 in full — already repaired; verification only.

### A.5 New-file footprint

This SPEC directory (3 plan-phase artifacts) and the card's own evidence under `.moai/reports/t1929/`. The repair adds no new files: both repair commits edit `<ID>/spec.md` frontmatter and body only, plus the progress record the lane keeps.

## §B — Known Issues (filtered)

- **B4 Frontmatter schema**: the repair edits `status:` and `updated:` in frontmatter plus the HISTORY body record only; canonical key names; no snake_case aliases.
- **B5 CI 3-tier**: spec-lint is the failing tier here; the repair must not disturb golangci-lint or the Go test tiers — it touches no Go source.
- **B6 spec-lint heading convention**: this SPEC's own artifacts use the `## §F — Out of Scope` + `### Out of Scope — X` shape; the bare h2 form is an error-severity finding.
- **B8 Working tree hygiene**: stage by explicit pathspec; the repair commits carry exactly one file each (`<ID>/spec.md`).
- **B10 Untouched paths**: PRESERVE list at §A.4 binds the whole run phase.
- **B11 Blocker reports**: the run agent returns structured blockers; no user prompts.

## §C — Pre-flight (run before any change)

```bash
git rev-parse --show-toplevel
git rev-parse --short HEAD
git branch --show-current
git status --short
```

Expected: the card worktree `.moai/worktrees/t1929`, HEAD `c3c86e795d` (or the lane's current tip — re-read, never assume), branch `WT-spec-lint-main`, clean tree. Then re-observe the defect on the current tree before repairing:

```bash
go run ./cmd/moai spec lint --json .moai/specs/SPEC-PROGRESS-RECORD-IO-001
```

Expected: the StatusTransitionInvalid finding, `advisory` absent (false) — the defect is still live. If the tree moved and the finding is gone, STOP and re-scope: the card premise changed.

## §D — Constraints

- Forward commits only; no rebase, no filter-branch, no revert of `fa1f04b9f4`, no force-push anywhere.
- Never touch `.moai/spec-lint-baseline.json`; never touch `internal/spec/`; never edit `<ID>/progress.md` in this card's window.
- The repair commits each stage exactly one path: `.moai/specs/SPEC-PROGRESS-RECORD-IO-001/spec.md`.
- No remote push, no PR, no release-branch write (REQ-SLR-006).
- The gate re-measurement runs in the real checkout with a binary built from the measured tree (REQ-SLR-008); an installed-binary shortcut without a provenance statement voids the verdict.

## §E — Self-Verification (per verification-claim-integrity 5-section format)

Each report item names (a) the exact command, (b) the verbatim observed output, and (c) the tree SHA it was captured against. The M2 exit batch:

1. AC matrix — every AC-SLR-NNN with its command and verbatim output.
2. The gate verdict — the `--baseline` run's exit code, its `baseline: OK` verdict line, and its inventory line, verbatim. (The per-rule `recorded -> current` increase lines print only on an EXCEEDED verdict, `internal/cli/spec_lint.go:255-259`; a green run asserts their absence through the OK verdict and the exit code, never by quoting an unprinted line.)
3. The baseline diff — `git diff <repair-start-sha> -- .moai/spec-lint-baseline.json` with empty output, where `<repair-start-sha>` is the repair-start commit SHA recorded at pre-flight and pinned in the report (N1 — the commit-pinned form, not a bare `git diff -- <path>`, which reads worktree-vs-index and misses a staged change).
4. The pair census — the targeted JSON re-run showing zero StatusTransitionInvalid for `<ID>` and zero VacuousTestAssertion for the three SPECs.
5. Gaps and residual risk stated explicitly — anything not re-observed on the repaired tree is a Gap, never a carry-over from the plan-phase census.

## §F — Milestones

- **M1 — the status-record repair** (REQ-SLR-001..004; AC-SLR-001..004). Pre-flight re-observation; then the single forward commit of REQ-SLR-002, as executed at `35cccfc217`: `<ID>/spec.md` frontmatter `implemented` → `completed` (the canonical row-3 edge), `updated:` refreshed, and the HISTORY record appended per REQ-SLR-003 (the `8d10007ca9` recorded close restored; the `fa1f04b9f4` mis-toked revert named; the leader disposition of 2026-10-23 산출물 실제 보완 cited with the one-commit simplification; AC-CI-007 left as the SPEC's own pending surface). Commit subject names the record-repair nature and cites `8d10007ca9` and `fa1f04b9f4`; trailer `Authored-By-Agent: orchestrator-direct` (the pair's expected owner class); progress record updated after the commit. (The original two-commit sequence — restore then row-7 amendment reopen — is retracted by the leader disposition; see spec.md §A v0.3.0.)
- **M2 — verification and the green gate** (REQ-SLR-005..008; AC-SLR-005..008). Build the linter from the repaired tree; run the targeted JSON (zero findings for `<ID>`), the three-SPEC VacuousTestAssertion sweep (zero), the full `--baseline` gate and observe the observable green form (exit 0, the `baseline: OK` verdict line, its inventory line, no increase lines anywhere in the output), and the baseline byte-identity diff in the commit-pinned form of §E item 3. Record the release-branch propagation note (REQ-SLR-006).
- **M3 — close-out evidence.** The §E report assembled from M2's verbatim outputs; progress.md phase sections populated by the owning agents; nothing pushed.

## §G — Anti-Patterns

- **G-1 — The non-canonical shortcut**: transitioning `implemented` directly to `in-progress` records a non-canonical pair and re-fires the finding. The executed repair's canonical edge (`implemented → completed`, row 3) is the only single-commit form out of the parked token; a `→ in-progress` reopen stays out of this card's window (REQ-SLR-002).
- **G-2 — The uncontextualized completion**: setting `completed` without the HISTORY record (REQ-SLR-003) would read as a fresh close claim while AC-CI-007 stands pending — the record must name the restored close (`8d10007ca9`), the mis-toked revert (`fa1f04b9f4`), and the leader disposition, so the audit context travels at the point of change. The leader disposition of 2026-10-23 (산출물 실제 보완 — substance repairs landed at `17a3b98de2`) is what makes the restored token truthful; the completion-leak prohibition (G-6) is unchanged.
- **G-3 — The baseline raise**: recording the +1 in the baseline file would green the gate by license, not by repair — operator-forbidden.
- **G-4 — The rule edit**: widening the canonical set or demoting the rule would green every future non-canonical record too.
- **G-5 — The archive-tree measurement**: running the gate in a git-archive export reads the walk as empty and "passes" vacuously. Real checkout only.
- **G-6 — The completion leak**: any edit to `<ID>/progress.md` §E.4 or any AC-CI-007 satisfaction claim during this card violates REQ-SLR-004.

## §H — Cross-References and Open Questions

- spec.md §B (the defect narrative), §D (requirements), §H (source citations). acceptance.md §E (the RED-now ledger captured at plan phase, tree `c3c86e795d`).
- Open question carried to run phase: none — the two operator-relevant judgment calls (repair shape, completion prohibition) are settled by the card constraints and REQ-SLR-002/004; the AC-CI-007 resolution remains t1598's own path, untouched here.
