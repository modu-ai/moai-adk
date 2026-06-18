# Plan — SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001

> **Tier M (standard)** — implementation plan for a policy-layer doctrine SPEC.
> Status: draft (plan-phase). Run-phase begins after plan-auditor verdict + Implementation Kickoff Approval (CLAUDE.local.md §19.1).

## §A. Context

This SPEC is the entry deliverable for Sprint 15 (harness-books analysis → application cohort). It codifies the single biggest gap found when diffing book1 ("Claude Code harness engineering", `github.com/wquguru/harness-books`) against moai-adk: moai-adk has a deliverable-recovery doctrine but no runtime-recovery doctrine, and its own hooks can cause the death-spiral that book1 ch06 explicitly warns recovery must avoid.

The deliverable is **policy-layer only** — a new rule + two scoped edits. No Go runtime implementation, no hook script rewrite. The boundary is stated in spec.md §D (Non-Goals) and reinforced in §F Constraints.

## §B. Known Issues / Risks

| Risk | Mitigation |
|------|------------|
| **R1 — scope creep into Go runtime** | spec.md §D + §E explicitly exclude Go code; plan.md §C pre-flight re-verifies before M1. |
| **R2 — file-edit collision with P1a (ORCH-INTERRUPT-LEDGER)** | agent-common-protocol.md edit is scoped to §Hook Invocation Surface ONLY; the §Ledger Closure subsection is reserved for SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001 (P1a, queued for Sprint 15 plan-phase, cohort task #2 — NOT a phantom sibling). M2 acceptance enforces the boundary via grep (no `Ledger` heading added; REQ-RR-011 normative). |
| **R3 — book-grounding fidelity drift** | REQ/AC wording cites book1 named principles verbatim (withheld-recoverable, cheapest-first, death-spiral, narrative consistency); plan-auditor cross-checks against spec.md §B citations. |
| **R4 — over-broad CONST entry** | zone-registry.md gets exactly ONE `CONST-V3R6-0XX` entry; the other 4 circuit-breaker invariants stay in the doctrine rule body, not the registry. |
| **R5 — doctrine-vs-render-surface drift** | runtime-recovery-doctrine.md is SSOT; agent-common-protocol.md §Hook Invocation Surface is render surface. The two MUST stay in parity (analogous to session-handoff.md ↔ output-style render-surface parity). |

## §C. Pre-Flight Verification (before M1)

- [ ] `grep -rEl 'reactive.?compact|death.?spiral|circuit.?break|withheld.*recover' .claude/rules/moai/` returns ZERO hits (confirms the recovery-ladder gap the SPEC closes still exists at plan-phase baseline). Note: `max_output_tokens` is NOT in this grep — it already has pre-existing coverage in `hooks-system.md` as a StopFailure error type and is not part of the gap (see spec.md §A).
- [ ] `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` does NOT yet exist (M1 creates it).
- [ ] `agent-common-protocol.md` §Hook Invocation Surface is at the current line range (lines 35-63 per the plan-phase baseline).
- [ ] `zone-registry.md` last entry is `CONST-V3R5-041` (first V3R6-era entry will be `CONST-V3R6-001` unless a parallel SPEC allocates it first — re-verify at M3).
- [ ] **In-flight V3R6 SPEC check (D10 / CONST-V3R6-0XX double-allocation avoidance)**: before M3, run `git status --porcelain | grep -E 'SPEC-V3R6' && ls .moai/specs/ | grep -E 'SPEC-V3R6'` to enumerate sibling V3R6 SPECs that may allocate `CONST-V3R6-001` first. Two parallel NAMESPACE SPECs (`SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001`, `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`) are visible in `git status` at plan-phase baseline; if either reaches M3 before this SPEC, allocate the next free `CONST-V3R6-0NN` numeric. The AC is satisfied by *any* valid `CONST-V3R6-0XX`, not by a specific number (see §D.4 E2).
- [ ] No open SPEC edits `agent-common-protocol.md` §Hook Invocation Surface (race check; if ORCH-INTERRUPT-LEDGER-001 is in-flight on a different file region, sequence M2 after it).

## §D. Constraints Carried Into Run-Phase

- Policy/rule layer only — no `internal/` Go code.
- agent-common-protocol.md edit confined to §Hook Invocation Surface.
- Template neutrality preserved (no internal SPEC IDs / commit SHAs / macOS paths leak into `internal/template/templates/`).
- Book-grounding citations preserved verbatim.

## §E. Self-Verification (run-phase §E.1 placeholder)

The run-phase manager-develop will populate §E.1 of `progress.md` with the plan-phase audit-ready signal. §E.2-§E.5 remain empty placeholders until run/sync/Mx phases (see progress.md skeleton).

## §F. Milestones

> Priority-ordered; no time estimates (per agent-common-protocol.md §Time Estimation).

### M1 — Author runtime-recovery-doctrine.md (Priority High)

- Create `.claude/rules/moai/workflow/runtime-recovery-doctrine.md`.
- §1 Withheld-recoverable-error set `{PTL, max_output_tokens, media_size, compact-failure}` (REQ-RR-001).
- §2 4-rung cheapest-first ladder with moai-adk artifact cross-references (REQ-RR-003, REQ-RR-004).
- §3 Five circuit-breaker invariants as policy (REQ-RR-005).
- §4 Anti-death-spiral hook carve-out — documented as **policy guidance** (NOT mechanical enforcement) for `sync-phase-quality-gate.sh` and `status-transition-ownership.sh` exiting 0 on recovery-signal turns (REQ-RR-006, REQ-RR-007; OPTION a documentation-only per spec.md §F.4 scope binding; runtime-layer enforcement deferred to future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001).
- §5 Cross-references to session-handoff.md, verification-claim-integrity.md, book1 ch03/ch06 (REQ-RR-009).
- §6 Agent consult-the-doctrine obligation sentence (REQ-RR-012) — the agent MUST consult the doctrine + apply the cheapest-first ladder before concluding the turn failed.
- AC binding: AC-RR-001, AC-RR-002, AC-RR-003, AC-RR-006 (SHOULD, documentation-only), AC-RR-007, AC-RR-011.

### M2 — agent-common-protocol.md §Hook Invocation Surface carve-out + boundary check (Priority High)

- Edit `.claude/rules/moai/core/agent-common-protocol.md` §Hook Invocation Surface ONLY.
- Add "Recovery-Signal Carve-Out" subsection documenting that Stop/PostToolUse hooks self-gate on recovery-signal detection and exit 0 when the turn is recovering (REQ-RR-006 render surface).
- **Boundary check** (anti-collision with P1a): grep that NO `Ledger` / `Ledger Closure` heading was added outside §Hook Invocation Surface.
- AC binding: AC-RR-004, AC-RR-008.

### M3 — zone-registry.md CONST-V3R6-0XX entry + cross-ref parity (Priority Medium)

- Edit `.claude/rules/moai/core/zone-registry.md` — append ONE `CONST-V3R6-0XX` entry (likely `CONST-V3R6-001`; re-verify highest V3R6 numeric at M3 start).
- The entry's `clause` names the anti-death-spiral invariant; `file` points to `runtime-recovery-doctrine.md`; `canary_gate: true` if the invariant is a safety gate.
- Verify `moai constitution list` returns the new entry (REQ-RR-008).
- AC binding: AC-RR-005.

### M4 — Lint clean + grep reproducibility (Priority Medium)

- Run `moai spec lint` on the SPEC directory — zero findings (or documented `lint.skip` debt).
- Run the gap-closing grep: each of the RECOVERY-LADDER terms `reactive-compact`, `death-spiral`, `withheld-recoverable`, `circuit-breaker` returns ≥1 hit in `.claude/rules/moai/` (REQ-RR-010). The error-type vocabulary (`max_output_tokens`, `media_size`, `compact-failure`) is NOT asserted — `max_output_tokens` already has pre-existing coverage in `hooks-system.md`.
- AC binding: AC-RR-009, AC-RR-010.

## §G. Anti-Patterns to Avoid

- **AP-RR-001 — Go runtime creep**: introducing `internal/recovery/` or similar. OUT OF SCOPE per spec.md §D. If a run-phase need emerges for Go detection, return a blocker and spawn a separate SPEC.
- **AP-RR-002 — agent-common-protocol.md scope escape**: editing §User Interaction Boundary or adding §Ledger Closure. The first belongs to a different concern; the second belongs to P1a SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001.
- **AP-RR-003 — Hook script rewrite**: modifying `.claude/hooks/moai/*.sh` bodies. The doctrine binds AGENT BEHAVIOR and FUTURE hook authors via policy guidance; no current script change is in scope. Mechanical enforcement of the recovery-signal carve-out by the current hooks is NOT possible (they do not parse `stopReason`) and is deferred to future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001 (REQ-RR-006/007 scope binding, spec.md §F.4).
- **AP-RR-004 — Paraphrasing away book1 named principles**: "withheld-recoverable", "cheapest-first", "death-spiral", "narrative consistency" are load-bearing terms from book1 ch03/ch06. Preserve them verbatim.
- **AP-RR-005 — Batched CONST entry**: putting all 5 circuit-breaker invariants into one `CONST-V3R6-0XX` entry. One entry for the anti-death-spiral invariant only; the other 4 live in the doctrine rule body.

## §H. Cross-References

- spec.md — full requirement set (12 REQs across §F.1-§F.6; iter-2 added REQ-RR-011 scope boundary + REQ-RR-012 agent consult obligation).
- acceptance.md — 11 ACs (9 MUST + 2 SHOULD; iter-2 downgraded AC-RR-006 to SHOULD documentation-only, added AC-RR-011) with Given-When-Then + grep-reproducibility evidence patterns.
- progress.md — §E skeleton (plan-phase audit-ready signal to be populated by run-phase).
- P1a sibling: `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001` — queued for Sprint 15 plan-phase (cohort task #2); owns the `§Ledger Closure` subsection this SPEC excludes (REQ-RR-011 boundary).
- Forward-link: future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001` — runtime-layer hook that parses `stopReason` to mechanically enforce the recovery-signal carve-out (deferred per REQ-RR-006/007 scope binding).
