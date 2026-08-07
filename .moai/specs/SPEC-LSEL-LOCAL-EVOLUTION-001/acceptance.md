# Acceptance — SPEC-LSEL-LOCAL-EVOLUTION-001

> **Format note:** per the SPEC artifact ownership contract, this file is the VERIFICATION layer — every entry is `AC-XXX` labeled `Given … When … Then …` and binary-testable. The GEARS obligation lives in `spec.md` §C (the requirement layer). Do not restate GEARS requirements here; do not present Given-When-Then as GEARS.

## §A. Verification Lens

Every AC below binds to one of the two evidence surfaces from `.claude/rules/moai/core/verification-claim-integrity.md`:
- **Surface 3** (defect/budget claim): a mechanical tool confirms the state (grep, test, file-existence, exit-code). The audit's own greps (report §2) are the evidence baseline.
- **Surface 1** (orchestrator/manager self-report): the `§E` self-verification matrix cites the verbatim command + observed output.

"No-claim-without-observation" binds both directions: a PASS requires an observed mechanical check; a FAIL requires the same.

## §B. Severity Conventions

- **MUST** — blocks milestone ship.
- **SHOULD** — strong expectation; a deviation requires a documented rationale in plan.md.
- **MAY** — optional / conditional (used for M5).

## §C. Traceability (which report section each AC grounds in)

| AC | Report section grounded in |
|----|---------------------------|
| AC-LSEL-001 | §7 guardrail 1 (allowlist), §11 mustFix A#1 |
| AC-LSEL-002 | §7 "core invariant — allowlist outside write region", §11 mustFix A#1 |
| AC-LSEL-003 | §2 audit finding (Applier.Apply never ran), §7 "frozen Go applier stays frozen" |
| AC-LSEL-004 | §7 "immutable version ID + immediate rollback (LangWatch)" |
| AC-LSEL-005 | §7 "CSA capability-conditional escalation", §11 mustFix A#2/A#3, D3 mechanical enforcement |
| AC-LSEL-006 | §7 "namespace separation (§24)", §11 mustFix A#4 |
| AC-LSEL-007 | §7 mechanical trigger, §11 mustFix B#1 |
| AC-LSEL-008 | §7 "hook applier consumes only approved queue", §11 mustFix A#5 |
| AC-LSEL-009 | §6 stage 2 (CLUSTER), §10 P1 |
| AC-LSEL-010 | §6 stage 2 severity filter, §11 mustFix B#3 |
| AC-LSEL-011 | §6 stage 3 (PROPOSE), §10 P2 |
| AC-LSEL-012 | §10 P2 ("verify Tier-4 fires before wiring"), §11 mustFix B#4 cautionary |
| AC-LSEL-013 | §6 stage 5 (APPLY), §10 P3 |
| AC-LSEL-014 | §11 mustFix A#8 (rollback rehearsal AC) |
| AC-LSEL-015 | §6 stage 7 (VERIFY), §11 mustFix B#6 |
| AC-LSEL-016 | §6 stage 7 REFLECTION, §10 P4 |

## §D. AC Matrix

| AC ID | REQ | Severity | Summary | Milestone |
|-------|-----|----------|---------|-----------|
| AC-LSEL-001 | REQ-LSEL-001 | MUST | allowlist hard-rejects FROZEN paths; reject log records fixture violations | M3 |
| AC-LSEL-002 | REQ-LSEL-002 | MUST | allowlist file stored OUTSIDE 6 evolvable surfaces; proposal-to-amend triggers forced gate | M3 |
| AC-LSEL-003 | REQ-LSEL-003 | MUST | `enableTriggerInjectionWrites` stays `false` post-M3; zero LSEL code mutates it | M3 |
| AC-LSEL-004 | REQ-LSEL-004 | MUST | every apply = one `lsel-*` tagged commit on feature branch; `git log --grep` finds it | M3 |
| AC-LSEL-005 | REQ-LSEL-005 | MUST | CSA forced-gate categories enumerated; bother-cost exemption clause explicit; applier mechanically refuses execution-meta proposals without synchronous-approval marker (D3) | M2 |
| AC-LSEL-006 | REQ-LSEL-006 | MUST | `split_namespace_test.go` + extended `internal_content_leak_test.go` green; zero template leak | M2 |
| AC-LSEL-007 | REQ-LSEL-007 | MUST | SessionStart backlog check fires on fixture overflow; default `/loop` recipe registered | M2 |
| AC-LSEL-008 | REQ-LSEL-008 | MUST | `lsel-apply.sh` is playback-only (no new-apply primitive in source) | M3 |
| AC-LSEL-009 | REQ-LSEL-009 | MUST | 569-stub backlog drained; `drain-offset.json` advances; zero `memory/` writes | M1 |
| AC-LSEL-010 | REQ-LSEL-009 | MUST | drain severity filter excludes Bash-timeout/sandbox noise BEFORE clustering | M1 |
| AC-LSEL-011 | REQ-LSEL-010 | MUST | shadow proposal payload schema present; self-critique blocks; retrieval-before-propose evidenced | M2 |
| AC-LSEL-012 | REQ-LSEL-010 | MUST | `moai-harness-learner` Tier-4 AskUserQuestion flow verified-firing before M2 wiring | M2 |
| AC-LSEL-013 | REQ-LSEL-011 | MUST | first `lsel-*` tagged commit lands; apply-ledger row appended; allowlist validates path | M3 |
| AC-LSEL-014 | REQ-LSEL-012 | MUST | rollback-rehearsal: `git revert <lsel-tag>` lands clean on mixed-history fixture | M3 |
| AC-LSEL-015 | REQ-LSEL-013 | MUST | VERIFY runs `/moai gate` superset; timeout retry-once; auto-revert on second fail | M4 |
| AC-LSEL-016 | REQ-LSEL-014 | MUST | reflection synthesizes principle file; originals archived (not deleted); decay-weighted retrieval | M4 |

## §D.1 AC Definitions (Given-When-Then)

### AC-LSEL-001 — allowlist hard-rejects FROZEN paths
**Given** the frozen allowlist regex at `.claude/lsel/frozen-allowlist.json` and a fixture proposal targeting a FROZEN path (`.claude/rules/moai/test.md`),
**When** `hns-lsel-applier` attempts to apply the fixture proposal,
**Then** the apply is hard-rejected AND a row is appended to `.moai/logs/lsel-reject.log` containing the rejected path AND no file is written to the FROZEN target.

### AC-LSEL-002 — allowlist file lives OUTSIDE the 6 evolvable surfaces
**Given** the 6 evolvable surfaces enumerated in spec.md §B.3,
**When** a filesystem search runs for the allowlist file under any of those 6 surfaces,
**Then** the allowlist file is NOT found under any of them AND is found at `.claude/lsel/frozen-allowlist.json` (outside all 6);
**And when** a fixture proposal attempts to amend the allowlist file,
**Then** the proposal is routed to a synchronous `AskUserQuestion` (CSA forced gate) regardless of proposer confidence — and the forced-gate is mechanically enforced by the applier per REQ-LSEL-005 (see AC-LSEL-005 for the refusal-behavior test).

### AC-LSEL-003 — frozen Go applier stays frozen
**Given** the post-M3 source tree,
**When** `grep -n "enableTriggerInjectionWrites" internal/harness/applier.go` runs,
**Then** the value at line ~22 is `false`;
**And when** `grep -rn "enableTriggerInjectionWrites" .claude/skills/hns-lsel-* .moai/hooks/lsel-* .moai/state/lsel/ .claude/lsel/ 2>/dev/null` runs,
**Then** zero matches (LSEL code does not mutate the frozen flag).

### AC-LSEL-004 — every apply is a `lsel-*` commit on a feature branch
**Given** a completed M3 apply,
**When** `git log --grep "lsel-" --oneline` runs,
**Then** at least one commit matches the `lsel-<proposal-id>` pattern AND `git branch --show-current` (or the PR's base ref) shows a feature branch (NOT `main`).

### AC-LSEL-005 — CSA forced-gate categories + bother-cost exemption + mechanical refusal
**Given** the `hns-lsel-applier` / `hns-lsel-curator` skills' documentation,
**When** a grep runs for the CSA forced-gate category enumeration (INVARANTS kernel, security/validation exceptions, HIGH-fan-in refs, Bash risk paths, `permissions.allow`, execution-meta files),
**Then** all six categories are named AND a bother-cost-exemption clause for forced gates is present;
**And given** a fixture proposal whose `target_surface` or `blast_radius` matches one of the four execution-meta categories enumerated in REQ-LSEL-005's enforcement clause (the frozen allowlist meta file at `.claude/lsel/frozen-allowlist.json`, an applier or curator skill body, the `lsel-apply.sh` apply hook script, or the `settings.local.json` hook-registration subblock),
**When** the applier (`hns-lsel-applier` driving `lsel-apply.sh`) attempts to apply that proposal WITHOUT an explicit synchronous-approval marker in its `decision.json` record,
**Then** the apply is REFUSED — no write primitive executes — AND a rejection row is appended to `.moai/logs/lsel-reject.log` naming the matched execution-meta category AND no file is written to the matched execution-meta target;
**And when** the same fixture proposal is re-run WITH an explicit synchronous-approval marker (an `AskUserQuestion`-produced approval artifact) recorded in its `decision.json`,
**Then** the apply proceeds (the refusal was keyed on the absent approval marker, not on the category match alone).

### AC-LSEL-006 — namespace separation + template leak guard
**Given** the EXTENDED `internal/template/internal_content_leak_test.go` and existing `internal/template/split_namespace_test.go`,
**When** `go test ./internal/template/...` runs,
**Then** both tests PASS (exit 0) AND a fixture containing `lsel` / `hns-lsel` / `SPEC-LSEL` / an internal SHA under `internal/template/templates/` is flagged by the leak test (positive-control fixture).

### AC-LSEL-007 — mechanical trigger (SessionStart + default `/loop` recipe)
**Given** a fixture `.moai/lessons-inbox.jsonl` whose line count exceeds `drain-offset.json + N` (configured threshold),
**When** a SessionStart hook fires (or the scheduled `/loop` recipe runs),
**Then** a system-reminder referencing the LSEL drain is emitted AND the default `/loop` recipe is registered (a named workflow file exists under `.claude/workflows/` or the recipe is wired in `settings.local.json`).

### AC-LSEL-008 — `lsel-apply.sh` is playback-only
**Given** the source of `.moai/hooks/lsel-apply.sh`,
**When** `grep -nE "decision\.json|approve" .moai/hooks/lsel-apply.sh` runs,
**Then** the script reads an approved `decision.json` AND a grep for any new-apply-creation primitive (e.g. a function that emits a fresh proposal or self-approves) returns zero matches.

### AC-LSEL-009 — 569-stub backlog drained (M1)
**Given** the 569-stub baseline in `.moai/lessons-inbox.jsonl` (re-measured at M1 start; the figure is a moving count) and the initial `drain-offset.json` value,
**When** `/moai:harness lsel drain` (or the mechanical trigger) completes one drain pass,
**Then** `drain-offset.json` advances by ≥1 stub (the consumed-stub marker) AND ≥1 candidate topic entry appears in `.moai/state/lsel/clusters.json` AND `find memory -newer <drain-start-timestamp> -name "feedback_*"` returns zero files (M1 does NOT write to `memory/`).

### AC-LSEL-010 — drain noise filter
**Given** a fixture inbox where ≥60% of stubs are `tool_failure:Bash:UnknownFailure` (timeout / sandbox noise),
**When** the drain pass completes,
**Then** the cluster output's accepted-topic set contains FEWER than 30% `tool_failure:Bash:*` entries (the write-time severity hint or drain-side filter excluded noise BEFORE clustering — report §11 mustFix B#3).

### AC-LSEL-011 — PROPOSE shadow payload + self-critique + retrieval
**Given** a post-M2 shadow proposal at `.moai/state/lsel/proposals/<id>/`,
**When** the proposal directory is inspected,
**Then** all of `{proposal.md, diff.patch, self-critique.md}` exist AND `proposal.md` carries the full payload schema (`proposal_id`, `target_surface`, `rationale`, `WHY-not-just-WHAT`, `prediction`, `verify_command`, `blast_radius`, `memory_type`) AND evidence exists that relevant `feedback_*.md` files were retrieved BEFORE drafting (retrieval log or cited files in `proposal.md`) AND a fixture self-critique with an unresolved objection blocks the proposal from APPROVE.

### AC-LSEL-012 — Tier-4 AskUserQuestion flow verified-firing before M2 wiring
**Given** the existing `moai-harness-learner` skill,
**When** a behavioral test OR a structured grep verifies the Tier-4 AskUserQuestion flow has a live production invocation path (NOT just type definitions — the audit's "CuratorDispatch 0 callers" is the cautionary precedent),
**Then** the test/grep confirms the flow fires AND the M2 wiring that depends on it is gated on this confirmation (the wiring commit MUST cite the verification).

### AC-LSEL-013 — first `lsel-*` apply lands + ledger row appended
**Given** an approved `decision.json` at `.moai/state/lsel/proposals/<id>/decision.json`,
**When** `lsel-apply.sh` plays it back,
**Then** a `lsel-<proposal-id>` commit lands on a feature branch AND a row is appended to `.moai/state/lsel/apply-ledger.jsonl` with `{proposal_id, target_surface, ts, result: "applied"}` AND the frozen allowlist validated the target path (no reject-log row for this proposal).

### AC-LSEL-014 — rollback rehearsal on mixed history
**Given** a fixture `CLAUDE.local.md` history with at least one `lsel-*` commit interleaved with ≥2 manual edits,
**When** `git revert <lsel-tag>` runs,
**Then** the revert completes without conflict (exit 0) AND the resulting `CLAUDE.local.md` state matches the pre-`lsel-*` state for the affected lines.

### AC-LSEL-015 — VERIFY independence + flaky-retry policy
**Given** a completed apply,
**When** VERIFY runs,
**Then** BOTH the proposal's `verify_command` AND `/moai gate` (lint+format+type+test) execute (the gate is MANDATORY, not optional);
**And when** a timeout-class failure occurs on the first VERIFY run,
**Then** VERIFY retries exactly once;
**And when** a second non-timeout failure occurs,
**Then** `git revert lsel-<proposal-id>` auto-fires AND the proposal's `feedback_*.md` is marked `verified:false`.

### AC-LSEL-016 — REFLECTION consolidation + decay-weighted retrieval
**Given** a fixture set of ≥3 concrete topic files whose accumulated importance exceeds the reflection threshold,
**When** the periodic REFLECTION pass runs,
**Then** an abstract-principle `feedback_*.md` is synthesized AND the originals are moved to `memory/_archive/` (NOT deleted) AND each new topic carries a `memory_type` label AND a retrieval probe for a related cue returns the new principle file ranked ABOVE the archived originals (decay-weighted retrieval).

## §E. Edge Cases

- **Empty inbox at drain time:** drain is a no-op, exits 0, advances no offset, no candidate topics. Not a failure.
- **Self-critique that never converges:** proposal stays blocked. Curator returns a blocker report; orchestrator surfaces via `AskUserQuestion`. Not a ship-blocker for M2 (it proves the gate fires).
- **Tier-4 flow found DEAD at M2 entry:** M2 does NOT wire the dependency; report finding logged; M2 downgrades to "PROPOSE shadow only, APPROVE via fresh path" and returns a blocker for orchestrator decision.
- **`git revert` conflict on mixed history (AC-LSEL-014 FAIL):** M3 does NOT ship. The rollback rehearsal is a precondition. Resolution requires re-designing the commit shape (smaller scope per `lsel-*` commit, or a squashed-apply strategy).
- **REFLECTION on a single-topic cohort:** insufficient material; the pass is a no-op with a log entry. Not a failure.

## §F. Quality Gate Criteria

- **TRUST 5 — Tested:** AC-LSEL-015 mandates `/moai gate` (test) on every apply; M3 rollback-rehearsal (AC-LSEL-014) is a behavior test.
- **TRUST 5 — Readable:** the mechanism skills are `hns-*` (English body per `.claude/rules/moai/development/agent-authoring.md`); `CLAUDE.local.md §28` follows the local-guide register (Korean/English mix per existing file).
- **TRUST 5 — Unified:** `hns-lsel-*` skills follow the namespace's existing formatting conventions.
- **TRUST 5 — Secured:** REQ-LSEL-005 names `permissions.allow` additions as an explicit security-exception band; REQ-LSEL-001 hard-rejects writes outside the 6 surfaces.
- **TRUST 5 — Trackable:** every apply carries a `lsel-<proposal-id>` commit tag (Conventional Commits); `apply-ledger.jsonl` is an immutable audit log.

## §G. Definition of Done (per milestone)

- **M1 DoD:** AC-LSEL-009 + AC-LSEL-010 PASS; `drain-offset.json` advanced; candidate topics staged under `clusters.json`; ZERO `memory/` writes from M1.
- **M2 DoD:** AC-LSEL-005, AC-LSEL-006, AC-LSEL-007, AC-LSEL-011, AC-LSEL-012 PASS; `CLAUDE.local.md §28` + INVARANTS kernel present; mechanical triggers wired.
- **M3 DoD:** AC-LSEL-001, AC-LSEL-002, AC-LSEL-003, AC-LSEL-004, AC-LSEL-008, AC-LSEL-013, AC-LSEL-014 PASS; first real `lsel-*` apply on a feature branch; rollback rehearsal clean.
- **M4 DoD:** AC-LSEL-015 + AC-LSEL-016 PASS; reflection pass produced ≥1 principle file; originals archived.
- **M5 DoD:** conditional on REQ-LSEL-015 simulation-harness proof; no MUST AC; MAY ship reduced or not at all.

## §H. Forward-Looking Checks (post-M3, pre-M5)

- After M3 ships, run a 30-day observation window: count `lsel-*` applies, count reverts, count `verified:false` lessons. If revert-rate > 25%, M4's VERIFY hardening is incomplete — surface to plan-auditor.
- After M4 ships, audit `memory/_archive/` growth: if cold-tier growth exceeds hot-tier (active `feedback_*.md`) by >3x over 60 days, the 50-file hot-tier cap is no longer the right threshold — flag for the memory-hygiene doctrine owner.
- M5 entry requires a named harness SPEC (plan.md §B.2 [DECISION]); until that SPEC is approved, M5 is NOT armable.
