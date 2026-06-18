# Progress — SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001

> **Tier M (standard)** — V3R6 4-phase close contract progress file.
> Plan-phase authored 2026-06-18 by manager-spec. §E.2-§E.5 commit_sha fields populate at run / sync / Mx phases.

## §E.1 Plan-phase Audit-Ready Signal

Plan-auditor verdict: PASS-WITH-DEBT 0.88 (run-ready). Implementation Kickoff Approval GRANTED by user via §19.1 AskUserQuestion on 2026-06-18 (option A — run-phase).

Plan-phase artifacts authored (iter-2 defect-fixed 2026-06-18, version 0.2.0):
- `spec.md` — 12 REQs across §F.1-§F.6 (GEARS notation), policy-layer scope, book1 ch03/ch06 grounding. iter-2 fixes: D1 §E→Out of Scope + h3 sub-section (lint clean); D2 `era: V3R6` frontmatter (EraAutoDetected suppressed); D3 REQ-RR-010 reframed to recovery-ladder vocabulary only; D4 REQ-RR-006/007 reframed as documentation-only policy recommendation (OPTION a); D5 §H Sprint 15 P1a queue annotation; D6 removed non-canonical `related_specs:`; D7 added REQ-RR-011 + REQ-RR-012; D8/D9 minor REQ fixes.
- `plan.md` — 4 milestones (M1 doctrine rule, M2 agent-common-protocol carve-out, M3 zone-registry CONST, M4 lint + grep reproducibility). iter-2 added D10 in-flight V3R6 SPEC check to §C M3 pre-flight.
- `acceptance.md` — 11 ACs (9 MUST + 2 SHOULD; iter-2 downgraded AC-RR-006 to SHOULD documentation-only, added AC-RR-011) with Given-When-Then + grep evidence commands.
- `progress.md` — this file (§E skeleton; §E.2-§E.5 empty placeholders).

Frontmatter: status `draft`, `era: V3R6` (D2 — explicit frontmatter field suppresses EraAutoDetected INFO finding per lifecycle-sync-gate.md), 12 canonical fields validated, SPEC ID self-check PASS, `related_specs:` removed (D6 — non-canonical).

SPEC ID pre-write self-check: `decomposition: SPEC ✓ | V3R6 ✓ | HARNESS ✓ | RUNTIME ✓ | RECOVERY ✓ | 001 ✓ → PASS`

## §E.2 Run-phase Evidence

**M1 — runtime-recovery-doctrine.md authored** (NEW file, 7 sections):
- §1 withheld-recoverable-error set `{PTL, max_output_tokens, media_size, compact-failure}` (AC-RR-001)
- §2 4-rung cheapest-first ladder table + §2.1 ordering rule (AC-RR-002)
- §3 five circuit-breaker invariants (AC-RR-003)
- §4 Recovery-Signal Carve-Out (documentation-only policy; AC-RR-004 SSOT, AC-RR-006 SHOULD)
- §5 cross-references (AC-RR-007 SHOULD)
- §6 agent consult-the-doctrine obligation (AC-RR-011)
- §7 anti-patterns (AP-RR-001..006)

**M2 — agent-common-protocol.md §Hook Invocation Surface carve-out** (AC-RR-004 render surface + AC-RR-008 boundary):
- Added "Recovery-Signal Carve-Out" subsection under §Hook Invocation Surface (documentation-only guidance; mirror of doctrine §4).
- Boundary grep confirms ZERO `Ledger` headings added (AC-RR-008 / REQ-RR-011).

**M3 — zone-registry.md CONST-V3R6-001 entry** (AC-RR-005):
- Appended one entry `CONST-V3R6-001` naming the anti-death-spiral invariant; `file: .claude/rules/moai/workflow/runtime-recovery-doctrine.md`; `canary_gate: true`.
- Highest V3R6 numeric at M3 start was none (first V3R6 entry → 001).

**M4 — lint + grep reproducibility** (AC-RR-009, AC-RR-010):
- `moai spec lint` clean (0 findings) — see §E self-verification E5.
- Recovery-ladder grep: each of `reactive-compact`, `death-spiral`, `withheld-recoverable`, `circuit-breaker` returns ≥1 hit in `.claude/rules/moai/` — see §E self-verification E7.

No Go code added (AC-RR-009 Non-Goals boundary — `internal/recovery/` absent).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

sync_commit_sha: ""

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_

mx_commit_sha: ""
