---
id: SPEC-SYNC-GATE-VERDICT-001
title: "Sync quality gate: three-arm outcome-record execution proof on the current tree, severity-to-verdict unification in the template doc copy, and manifest-observation wording truth"
version: "0.2.1"
status: in-progress
created: 2026-09-14
updated: 2026-09-18
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/template/templates/.claude/hooks/moai; internal/template/templates/.claude/skills/moai/workflows/sync"
lifecycle: spec-anchored
tags: "sync-gate, stop-hook, outcome-record, severity-verdict, template-first, doc-truth, t783"
tier: M
related_specs: [SPEC-SYNC-GATE-FAILSTATE-001]
---

# SPEC-SYNC-GATE-VERDICT-001 — sync quality gate: outcome proof, severity-verdict unification, manifest-observation wording truth

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.2.1 | 2026-09-14 | manager-spec | iter-2 residual F7 folded in: stale-clause removal patterns' positive control re-scoped to the current pre-M3 tree (measured 0 hits on the `2213871af` baseline doc — the clauses postdate it via t624's M3); the baseline positive control applies to the H01 hook-state arms and the H03 absence pattern only (AC-SGV-005/006, acceptance §A, plan M1) |
| 0.2.0 | 2026-09-14 | manager-spec | Plan-audit findings F1-F6 applied (PASS 0.88, `.moai/reports/t783/plan-audit.md`). F1 (High): option A — the Phase 8 relationship paragraph's two stale clauses are ALIGNED OUT of both copies (freeze dropped); F2: all REQ bodies reflowed SHALL-first; F3: write-ordering clause declared consumed-from-FAILSTATE-001; F4: parity mechanical proxy named; F5: hook neutrality tightened to no-NEW-card-IDs-on-edited-lines; F6: baseline-hook gate-layout note added |
| 0.1.0 | 2026-09-14 | manager-spec | Initial plan-phase draft (card t783, branch `WT-syncgate-hook`, base `a404132e7`) |

## §A Problem Statement

Backlog card t783 (issued 2026-09-10 from instruction-audit finding set G3, audit baseline
`main` commit `2213871af`) carries three findings against the sync-phase quality gate surface:

- **H01 (P1)** — the gate recorded the execution attempt rather than the check result: on
  `2213871af`, a same-HEAD continuously failing Go check yielded a block JSON on the first
  Stop-hook call and an EMPTY stdout on the second, so the failure state was lost to the next
  call. Improvement named by the card: store success/fail/in-progress distinctly; re-deliver
  the stored block for a same-HEAD failure; tie re-execution to input change or explicit retry.
- **H03 (P1)** — the doc promised an automatic dependency vulnerability scan the hook does not
  perform (`quality-gates-quality.md:145` on the audit baseline: "the dependency vulnerability
  scan runs automatically via the Stop hook"). The hook's Go branch is vet+build; its manifest
  check is a `git diff` observation (`deps_modified`).
- **SX-R05 (P2)** — the security-decision severity-to-verdict mapping was inconsistent between
  review steps ("Only CRITICAL findings block" vs the sync-auditor rubric's Critical/High →
  FAIL), and Phase 8's relationship to that rubric was unstated.

Plan-phase measurement on the work tree (develop `a404132e7`, this run) shows the three
findings sit at DIFFERENT completion states here, and the SPEC scopes each to its actual
residual:

1. **H01 — machinery landed, execution proof owed on the current tree.**
   SPEC-SYNC-GATE-FAILSTATE-001 (card t624, `status: completed`, closed 2026-09-11) already
   delivered the outcome-record machinery (record `running`/`pass`/`fail`, payload file,
   retry marker, worktree-content identifier, same-HEAD re-delivery; hook header contract
   lines 28-66, record write at 519, payload-before-fail-record at 739-749) and executed its
   own control arms on ITS tree. Since that close, further hook commits landed on the same
   file (t603 C++ repair, t604 Kotlin detection, t663 `*.cxx`, t664 scan depth). Card t783's
   hard constraint mandates control-arm EXECUTION evidence, and none of it exists against the
   CURRENT tree. This SPEC's H01 deliverable is that execution proof — with a positive control
   against the `2213871af` hook copy demonstrating the harness catches the original defect —
   plus a minimal repair ONLY where an arm fails (REQ-SGV-003).
2. **H03 — wording already corrected on develop; verification and provenance owed.** Both doc
   copies carry the truthful "it is not a vulnerability scan" wording (template lines 116/142,
   local lines 135/158) and the hook comment says the same (line 661-662); the baseline false
   promise is gone. The residual is the develop-side verification of that absence (measured,
   not assumed) plus the scope fences: no scanner is added, no existing gate is touched.
3. **SX-R05 — genuinely unresolved in both copies; THE substantive edit.** FAILSTATE-001
   AC-012(b) was a regression-guard that deliberately preserved the CRITICAL-only gate ("no
   behavior change"). A later, local-only edit unified the LOCAL copy's decision text to the
   severity contract (local lines 154/166-177: "Critical and High block; Medium and Low
   are advisory", "Continue by approved exception", "If no blocking finding exists") without
   mirroring the template copy. Measured today: the template copy still carries the old trio
   ("Only CRITICAL findings block" 1 hit; "Continue with warning" 1 hit) and the old
   CRITICAL-only gate; and BOTH copies carry the Phase 8 relationship paragraph with two
   clauses the unified gate falsifies ("its CRITICAL-only stop gate below", "a HIGH finding
   that Phase 8 reports only as a warning" — measured byte-identical in both copies at
   template line 148 / local line 164). The local copy therefore already contradicts itself,
   and the template copy would gain the same contradiction if only its decision block were
   replaced. This SPEC aligns the severity-to-verdict statement AND those two stale clauses
   out of BOTH copies, template first (inline contract — the `security-decision-contract.md`
   rule it models is local-only and NOT template-mirrored, so the template text must not
   reference that path).

Evidence-path convention: `.moai/reports/t783/` (untracked files in the PRIMARY checkout, per
operator directive; not committed).

## §B Scope

In scope: the Stop-hook script pair (`.claude/hooks/moai/sync-phase-quality-gate.sh` and its
template mirror — byte-identical pair; no source-neutrality constraint on the hook, but no
NEW internal card IDs on edited lines, REQ-SGV-007); the sync workflow doc pair
(`.claude/skills/moai/workflows/sync/quality-gates-quality.md` and its template mirror)
LIMITED to the finding surfaces (H03 manifest-observation wording, SX-R05 severity decision
block AND the Phase 8 relationship paragraph's two stale clauses); the control-arm execution
harness and its evidence.

Out of scope (see §F): any dependency vulnerability scanner (t610 owns govulncheck); the
broader doc-pair divergence (TRACE PROBE comment format, Step 0.5.1 routing-contract language,
bounded-queue scheduling, A7 Phase 9 concurrency, `coverage_scope`, Step 0.7.4
verification-plan reuse, status-mode early exit) — each owned by other work; `make build`;
settings.json wiring (already wired at line 155 in both `settings.json` and
`settings.json.tmpl`, verified this run).

## §C User Stories

1. As an operator ending a sync turn under a failing vet/build, I want the block re-delivered
   on the next turn under the same HEAD until the input changes, so a failure cannot read as a
   pass.
2. As a reader of the sync workflow doc in a fresh project, I want the severity decision
   stated once, consistently, in the copy my project actually ships, so I do not act on a
   contradicting older paragraph.
3. As a maintainer, I want the doc to describe what the hook measures (`deps_modified`
   observation) and not promise a scan it does not run, so the documentation stays verifiable.

## §D Requirements (GEARS)

Each requirement states its SHALL on the first line; conditions and bounds follow (plan-audit
F2 — modality-first shape, per `internal/spec/lint.go` judgeModality).

- **REQ-SGV-001 (H01, Must)** — THE SYSTEM SHALL re-deliver the first call's block result on the second Stop-hook invocation — non-empty stdout carrying the identical block decision, exit 0 on both calls, no re-run of the checks on the second call — when the sync quality gate hook is invoked twice consecutively against the same HEAD, same work-tree content, in blocking mode, with a failing fast check. The positive control (the same two-call sequence against the `2213871af` hook copy) SHALL first reproduce the audit's defect shape (first call block, second call empty stdout) before any current-tree arm result is trusted.
- **REQ-SGV-002 (H01, Must)** — THE SYSTEM SHALL re-execute the checks on the next invocation when the work-tree content changes under an unchanged HEAD after a recorded failure (fresh execution evidence, decision reflecting the new input); SHALL emit empty stdout with exit 0 on re-invocation when the same HEAD has a recorded pass (no stale block); and SHALL keep the three-outcome distinction on disk (`running` before checks, `fail`/`pass` after). The payload-before-fail-record write ordering is CONSUMED FROM SPEC-SYNC-GATE-FAILSTATE-001 (its torn-write shims and mutants own that ordering's verification); this SPEC verifies the post-hoc state shape only and claims no new ordering proof (plan-audit F3).
- **REQ-SGV-003 (H01, Must)** — THE SYSTEM SHALL receive a minimal repair scoped to the failing arm if any control arm fails on the base tree, applied to BOTH hook copies in one commit (template copy first, local copy byte-identical); the hook code SHALL NOT CHANGE if all arms pass, and the recorded execution proof is the deliverable.
- **REQ-SGV-004 (H03, Must)** — THE SYSTEM SHALL describe, in either doc copy, the hook's dependency mechanism as an informational manifest-change observation (`deps_modified`, HEAD-commit-restricted `git diff`), not as a vulnerability scan; the baseline false-promise phrasing SHALL remain absent from both copies; the hook's matching scope comment SHALL remain present.
- **REQ-SGV-005 (H03 fence, Must)** — THE SYSTEM SHALL NOT introduce any dependency vulnerability scanner, and SHALL NOT delete or weaken any existing gate check in the hook (the `deps_modified` observation and the per-language fast checks survive unchanged).
- **REQ-SGV-006 (SX-R05, Must)** — THE SYSTEM SHALL carry, in each doc copy's security gate decision, a single severity-to-verdict statement — Critical and High block; Medium and Low advisory; a blocking finding proceeds only via a user-approved exception record carrying finding ID, rationale, scope, approver, expiry, and review condition. THE SYSTEM SHALL state in each copy that Phase 8 is an additional lens whose stop gate never clears an earlier sync-auditor FAIL (the Step 0.5.4 rubric canonical), and SHALL remove the superseded severity trio ("Only CRITICAL findings block" / "HIGH findings are reported as warnings" / "Continue with warning") and the two relationship-paragraph clauses the unified gate falsifies ("its CRITICAL-only stop gate below"; "a HIGH finding that Phase 8 reports only as a warning") from BOTH copies (plan-audit F1 option A).
- **REQ-SGV-007 (Template-First + neutrality, Must)** — THE SYSTEM SHALL have the template copy edited first and the local copy aligned to it on the scoped passages when the doc edit lands; the scoped passages SHALL agree semantically across the pair, verified by the mechanical proxy of AC-SGV-007(b), with exactly one documented deliberate delta (the local copy may reference the local-only severity-contract rule path; the template copy states the contract inline); the template doc copy's edited passages SHALL carry no internal SPEC IDs, no audit citations, no internal dates, and no reference to a rule file the distributed template does not ship; the hook copies SHALL carry no NEW internal card IDs on edited lines (pre-existing citations on untouched lines are preserved, plan-audit F5); no whole-file verbatim copy in either direction.
- **REQ-SGV-008 (process fence, Must)** — THE SYSTEM SHALL record all measurement evidence (commands + verbatim outputs + exit codes, file-redirect form, no pipes on verdict-bearing commands) under `.moai/reports/t783/` as UNTRACKED files in the primary checkout; the run phase SHALL NOT execute `make build`; verification of the doc pair SHALL be source-axis (template↔local diff evidence on the scoped passages); the wider divergence passages listed in §B SHALL remain untouched.

## §E Non-Goals

- Adding govulncheck or any scanner (t610).
- Aligning the two doc copies as whole files (the remaining divergence belongs to other work).
- Changing gate blocking semantics, mode resolution, or the settings wiring.
- Re-opening FAILSTATE-001's closed decisions (payload generalization, retry bound, advisory
  once-emitted) — this SPEC consumes them as fixed, including the payload-before-record write
  ordering (REQ-SGV-002).

## §F Out of Scope

### F.1 Findings-surface fences

- The scanner add (t610), per the card's H03 scope fence.
- Hook behavior changes beyond REQ-SGV-003's conditional minimal repair.
- The broader doc divergence enumerated in §B (each passage is another card's surface; this
  SPEC's diff on the doc pair stays inside the H03/SX-R05 passages plus the Phase 8
  relationship paragraph's two stale clauses).

### F.2 Explicitly out of scope

- `make build` and any binary recompilation (lead builds at batch end).
- `.agents/skills/moai-sync/SKILL.md` regeneration (verified this run: the emitted skill
  carries zero occurrences of the affected wording, so the commandemit surface is untouched).
- Any change to `.claude/settings.json` / `settings.json.tmpl` hook wiring.

## §G Acceptance Overview

Nine acceptance criteria (AC-SGV-001..009, `acceptance.md`), one per requirement plus the
copy-pairing and scope-fence checks; every headline AC is verified by an executed command
whose output is file-redirected with its exit code recorded. REQ/AC counts sit inside the
Tier M ceilings (≤16 each).
