---
id: "SPEC-CLOCAL-AUDIT-001"
title: "Exhaustive verification audit of CLAUDE.local.md against repository reality at d29b8942e"
version: "1.0.1"
status: "in-progress"
created: "2026-08-27"
updated: "2026-08-27"
author: "manager-spec"
priority: "P2"
phase: "v3.1.4"
module: "CLAUDE.local.md"
lifecycle: "spec-anchored"
tags: "documentation-audit,evidence-bearing,claude-local-md,card-t308"
tier: "M"
type: "documentation-audit"
related_specs: ["SPEC-STAMP-REACHABILITY-001"]
---

# SPEC-CLOCAL-AUDIT-001 — Exhaustive Verification Audit of CLAUDE.local.md

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-08-27 | 0.1.0 | manager-spec | Plan-phase artifact set authored for card t308 (spec / plan / acceptance / progress). Claims inventory frozen at `.moai/reports/t308/claims-inventory.md` (70 items, INV-001..INV-135). [0.1.0 provenance restored 2026-08-27 — N4] |
| 2026-08-27 | 1.0.0 | manager-spec | Status `draft` emitted; audit basis pinned to develop SHA d29b8942e (worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t308`, branch `WT-clocal-audit`). |
| 2026-08-27 | 1.0.1 | manager-spec | Plan-audit iter-1 repairs applied: (D1) content-based primary exclusion-zone anchor + mid-run self-reanchoring rule; (D2) all inventory line anchors mechanically re-derived vs d29b8942e (~18 cells re-pinned; CHK-000e/000f); (D3) six coverage-hole rows added (INV-136..INV-141, total 76); (D4) AC-CLOCAL-012 remapped REQ-CLOCAL-005 → REQ-CLOCAL-001; (D5) register-symbol typo fixed; (D6) AC-004 transcript-predicate form + GWT blocks for AC-005/012; (D7) INV-021/022 named in M7; (D8) GEARS label fixes (REQ-002 Where→When, REQ-003 relabel, REQ-004 declarative pointer clause). Inventory now spans INV-001..INV-141 (76 items). |

## A. Background

`CLAUDE.local.md` (663 lines at develop SHA d29b8942e; last content commit 20fab1786, 2026-08-27) is the developer-facing guide for moai-adk-go local development. Its authority depends on citation accuracy. Its recurring defect pattern is **stale line/symbol/path citations that still look plausible** — an incident on record: an earlier revision asserted a gate-config loader was absent while `internal/config/loader_gate.go` existed.

Card t308 directs an exhaustive claim-by-claim verification sweep of that copy against the SAME tree d29b8942e plus installed source referenced therein, correcting every CONFIRMED defect minimally in place.

**Audit basis (pinned)**:
- Subject under audit: `CLAUDE.local.md` at worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t308` @ HEAD `d29b8942e` (verified in this run: `git rev-parse HEAD` → `d29b8942e5b237ef7180749dc04273d83647c745`, branch `WT-clocal-audit`; `wc -l` → 663).
- Reality measured against: the same tree `d29b8942e` plus the installed source referenced therein (`internal/**`, `Makefile`, `scripts/**`, `.claude/**`, `.moai/docs/**`).
- One authorized PRIMARY-checkout read exception: §28 LSEL claims are measured at `/Users/goos/MoAI/moai-adk-go/.claude/settings.local.json` and `/Users/goos/MoAI/moai-adk-go/.moai/state/lsel/` (untracked runtime data absent from the worktree). Every such finding MUST name WHICH surface was checked.

## B. Requirements (GEARS)

### REQ-CLOCAL-001 — Audit-basis integrity (Ubiquitous)
The run-phase auditor shall measure every claim against the pinned audit basis — subject copy `CLAUDE.local.md` @ `d29b8942e` in the card worktree — and shall re-verify HEAD immediately before any commit (`git rev-parse --short HEAD` + `git branch --show-current`), stopping and reporting divergence instead of proceeding when either differs from the pinned basis. When a finding or anchor is measured on the PRIMARY-checkout §28 surfaces (authorized read exception), the auditor shall state in the finding WHICH surface was checked.

### REQ-CLOCAL-002 — Mechanical-evidence obligation (Event-driven)
**When** a finding is reported as DEFECT-CONFIRMED, the run-phase auditor shall have executed the domain-appropriate mechanical check (path/existence → `ls`/`test`/glob with exact scope named; symbol-or-line citation → open the file and confirm symbol name AND quoted line range hold at `d29b8942e`; default value → compare against `internal/config/defaults.go`; env-var name → compare against `internal/config/envkeys.go` and `manager.go applyEnvOverrides`) and recorded command + verbatim output + exit code in `.moai/reports/t308/checks-transcript.md`.

### REQ-CLOCAL-003 — Hypothesis labeling (Event-driven)
When a suspected defect has been established only by reading text without executing a mechanical check, the auditor shall not classify it DEFECT-CONFIRMED; it shall label it `UNVERIFIED-HYPOTHESIS` until a check executes or move it to Open Questions after effort.

### REQ-CLOCAL-004 — Exclusion zones (Ubiquitous)
The auditor shall exclude §4.1 from defect scope, where §4.1 is defined PRIMARILY by content: the `### §4.1 ...` heading through the horizontal rule immediately preceding the `## 5.` heading of CLAUDE.local.md (the rewritten git-flow integration-lane section); residual defects inside it belong to sibling cards t294/t295/t298/t303. At the pinned basis d29b8942e this zone spans lines 273–337 — ADVISORY coordinates only. While any CLAUDE.local.md edit has landed above the zone during run-phase, the auditor shall re-locate BOTH boundaries from the heading text before further edits and log the shifted live interval in the progress notes; frozen-basis line numbers stay valid solely for reading diffs against the frozen copy. Pointer validation shall be limited to verifying that pointers FROM outside sections INTO §4.1 name real subsections — this is the sole sanctioned access to the zone.

### REQ-CLOCAL-005 — Known-unresolved carry-through (Event-driven)
**When** the sweep encounters (a) the §4.1-end BranchGuard over-matching read-only `git branch` lookups or (b) the three registered CODE defects behind §2.3 (CleanMoaiManagedPaths protection-list absence; `archiveLegacySkills` invoked after the wipe; `--dry-run` not previewing deletions), the auditor shall carry them as KNOWN entries in the report without re-adjudicating their existence.

### REQ-CLOCAL-006 — Citation-vs-defect distinction (Ubiquitous)
The auditor shall distinguish citation verification (does the quoted line range hold AT `d29b8942e` — e.g. `deploy.go:29`, `deploy.go:122`, `plan.go:73`, `update.go:513`, `gate.go:160-161`, `Makefile:9`, `Makefile:38`, `pkg/version` compile defaults) from code-defect adjudication: citation drift outside §4.1 IS in defect scope even where the referenced defect itself is known-unresolved.

### REQ-CLOCAL-007 — Classification vocabulary (Ubiquitous)
Every inventory item shall terminate in exactly one verdict class: `DEFECT-CONFIRMED`, `PASS-ATTESTED` (mechanically held), `HISTORICAL-RECORD` (provenance narrative whose checkable anchors hold — verify anchors cheaply, never re-run history), `EXTERNAL-UNVERIFIABLE` (pricing/billing figures — mark, recommend keep absent contradiction, stop), `KNOWN-UNRESOLVED` (REQ-CLOCAL-005), `AMBIGUOUS-PATH`, or `OPEN-QUESTION`.

### REQ-CLOCAL-008 — Fix routing (State-driven)
**While** run-phase is active, the auditor shall apply fixes ONLY as follows: `DEFECT-CONFIRMED` → minimal in-place text correction preserving tone/format; historical recalibration → keep narrative, add dated correction marker; hypothesis unresolved after effort → move to the explicit Open Questions subsection, never silently dropped; cross-document staleness discovered in OTHER docs → new-card recommendation in the report, never editing those docs.

### REQ-CLOCAL-009 — Scope discipline (Unwanted)
The auditor shall not edit any path other than `.moai/specs/SPEC-CLOCAL-AUDIT-001/**`, `.moai/reports/t308/**`, and `CLAUDE.local.md` (run-phase only); shall not run `go test ./...` anywhere; shall not spawn background load; and shall not modify anything in the PRIMARY checkout except the single authorized §28 read.

### REQ-CLOCAL-010 — Evidence pack completeness (Event-driven)
**When** the audit completes, the auditor shall leave `.moai/reports/t308/` containing: `claims-inventory.md` (frozen INV-NNN items mapped to sections/lines, verdict column populated), `verdict-draft.md` (five-section Claim/Evidence/Baseline-attribution/Gaps/Residual-risk draft), and `checks-transcript.md` (consolidated per-check logs with exit codes) — plus pasted progress notes suitable for `SPEC-CLOCAL-AUDIT-001/progress.md`.

### REQ-CLOCAL-011 — Frozen-inventory closure (Event-driven)
**When** run-phase ends, every INV item in `claims-inventory.md` shall carry a verdict class per REQ-CLOCAL-007 — zero items left blank, zero items silently dropped.

## C. Exclusions

### Out of Scope — Code changes
- Any Go/shell/script code modification (the three §2.3 code defects, gate fallback behavior, deploy wipe logic) — these are OTHER cards' concern; this audit touches prose only.

### Out of Scope — Other documents
- Edits to `.moai/docs/git-workflow-doctrine.md`, `git-local-workflow-doctrine.md`, or any document other than `CLAUDE.local.md` — cross-doc staleness is recorded as new-card recommendations only.

### Out of Scope — Primary checkout mutation
- Writes anywhere under `/Users/goos/MoAI/moai-adk-go/` outside the card worktree (the §28 read-only probe at the primary `.claude/settings.local.json` and `.moai/state/lsel/` is the sole authorized access).

### Out of Scope — History re-execution and external economics
- Re-running historical incidents ("2026-08-15 실측" measurements, load-413, PR #1557 effects) — anchors checked cheaply, narratives never re-executed.
- Re-verifying Vercel pricing figures ($0.0035/$0.126 per min) or GitHub billing quotes — classified `external-unverifiable`.

## D. Acceptance Criteria (matrix structure; enumerated body in acceptance.md)

AC-CLOCAL-001..012 live in `acceptance.md` § D with Given-When-Then scenarios, severity, and traceability to REQ-CLOCAL-001..011.
