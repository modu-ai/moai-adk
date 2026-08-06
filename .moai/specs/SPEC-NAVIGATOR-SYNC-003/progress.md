# SPEC-NAVIGATOR-SYNC-003 — Progress

> Status: draft (plan-phase artifact set authored; plan-audit + Implementation Kickoff Approval held by orchestrator).

## §A. Phase

- [x] Plan-phase artifact set authored (spec.md, plan.md, acceptance.md, design.md, research.md, progress.md — Tier L, 6 files).
- [ ] plan-auditor audit (orchestrator-owned — runs on these working-tree files next; this agent does NOT commit).
- [ ] Implementation Kickoff Approval (orchestrator-owned human gate).

## §B. Plan-phase self-check

- [x] SPEC-ID regex check PASS — `SPEC-NAVIGATOR-SYNC-003` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` (verified as executed Bash, output `PASS: SPEC-NAVIGATOR-SYNC-003 matches specIDPattern`).
- [x] SPEC-ID collision check PASS — `.moai/specs/SPEC-NAVIGATOR-SYNC-003` did not exist (confirmed via `find` — 0 matches repo-wide); created fresh.
- [x] 12 canonical frontmatter fields present in spec.md (`id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags`); no snake_case aliases.
- [x] Optional fields: `tier: L`, `era: V3R6`, `phase: "v3.3 target"` (release target, NOT a lifecycle token — passes the `FrontmatterPhaseInvalid` guard), `related_specs: [5 predecessors]`, `depends_on: [SPEC-NAVIGATOR-SYNC-001]` (M0 `status: completed` → Depends_on Pre-flight will PASS).
- [x] 22 REQs in GEARS notation (Ubiquitous / Event-detected / Capability-gate `Where`) — within the Tier L 25 ceiling; 22 distinct REQ-NS3-001..022.
- [x] 22 ACs (AC-NS3-001..022) in Given-When-Then — within the Tier L 25 ceiling; 1:1 REQ→AC trace (§F matrix).
- [x] Out of Scope section carries 8 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (satisfies `OutOfScopeRule` / `MissingExclusions` — a bare `## Out of Scope` H2 is absent).
- [x] Artifact set matches Tier L (6 files: spec + plan + acceptance + design + research + progress).
- [x] spec.md carries no implementation detail (function names appear only as evidence anchors in design.md / acceptance.md / research.md, not as requirements).
- [x] REQ/AC ceilings independent (22 REQ ≤ 25; 22 AC ≤ 25).

## §C. Open items surfaced for the orchestrator

1. **SPEC-ID re-numbering (transparency, NOT a blocker)** — M0 projected 003=M2 (Route); this SPEC is 003=M4 per the team-lead's instruction (M4 prioritized ahead of M2/M3). Recorded in spec.md HISTORY + research.md §4. M2 (Route) and M3 (Fix) will receive 004/005 (or later) when authored. Changes no technical content.
2. **SubagentStart hook stale spec reference** — the launch context carried `spec:SPEC-NAVIGATOR-SYNC-002` (the M1 sibling, already `completed`). Confirmed benign (not a collision; the hook reads stale state); this agent authored 003 as instructed. Surfaced for transparency.
3. **Success-metric fixture novelty (AC-NS3-022)** — the "≥40% reads-reduction" metric is the most novel AC (no prior art in the repo). The fixture procedure is named in acceptance.md §D.AC-NS3-022 (a static corpus + a deterministic read-count comparator). The plan-auditor should scrutinize whether the fixture is mechanically measurable as written; if the auditor judges the fixture underspecified, the fallback is to mark this AC as `deferred-to-M5` (M5 brownfield provides a richer measurement context) without blocking M4 acceptance. NOT a blocker for plan-phase — flagged for auditor attention.
4. **All 5 pre-flight decisions (D1–D5) RESOLVED** — pointers in plan.md §C reference design.md §1.D{1..5}; no open `[NEEDS CLARIFICATION]` markers remain in plan.md or research.md (MP-7 firewall cleared).

## §D. Domain-specialist consultation (Step 6 — not triggered)

Full-Go implementation surface within `manager-develop`'s core competency (a new `internal/navigator/tiers/` package importing an existing `internal/navigator/astx/` engine). No backend / frontend / devops domain keyword triggers. No external specialist spawn required.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-06
plan_artifact_set: Tier L (6 files: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md)
req_count: 22  # REQ-NS3-001..022, within Tier L 25 ceiling
ac_count: 22   # AC-NS3-001..022, within Tier L 25 ceiling; 1:1 REQ→AC trace
preflight_decisions_resolved: 5  # D1..D5, design.md §1
non_overlap_carry_forward: 10    # predecessor REQs carried forward (research.md §3)
open_clarification_markers: 0    # MP-7 firewall cleared
spec_id_regex_check: PASS        # SPEC-NAVIGATOR-SYNC-003 matches ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$
spec_id_collision_check: PASS    # directory was free
phase_field_release_target: true # "v3.3 target" — not a lifecycle token
era: V3R6                        # explicit override (modern era; subject to drift detection)
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters**
- tier: L
- scope (file count est.): ~14 implementation files (`internal/navigator/tiers/` ~9 Go files + `internal/cli/navigator_tiers.go` + 5 template surfaces)
- domain count: 4 (navigator/tiers Go package, CLI wiring, template-first surfaces, SPEC docs)
- file language mix: Go source (primary) + markdown templates + JSON/YAML schemas
- concurrency benefit: LOW (coding-heavy, sequential dependency chain: schema → tier engines → overlay join → close)
- Agent Teams prereqs: N/A (Mode 3 retired)

**Mode evaluation**
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | substantial new code, 7 milestones |
| 2 background | no | write-capable, result needed before continuing |
| 3 agent-team | no | RETIRED |
| 4 parallel | no | coding-heavy with sequential dependency chain; Anthropic coding-task parallelism caveat |
| 5 sub-agent | YES | coding-heavy sequential, default for new-code work |
| 6 workflow | no | new-code + multi-rule, not mechanical-uniform transform; <30 files |

**Decision**: sub-agent (Mode 5), cycle_type=tdd, milestone-coherent chunks with orchestrator trust-but-verify gates between.

**Justification**: This is coding-heavy new-code work (a new `internal/navigator/tiers/` package with a sequential dependency chain: schema → three tier engines → overlay integration → template-first close). Per Anthropic's coding-task parallelism caveat, sequential sub-agent delegation is the correct default for coding work. Semi-autonomous progression: the orchestrator delegates in milestone-coherent chunks (M4.1+M4.2 foundation → M4.3+M4.4+M4.5 engines → M4.6+M4.7 integration+close), running a trust-but-verify batch at each chunk boundary; `/moai goal` is NOT armed (per-session milestone gates instead). Route B (Tier L, feature branch `feat/SPEC-NAVIGATOR-SYNC-003` + PR per repo-local `enforce_admins` policy). Implementation Kickoff Approval PASSED (user-approved). Phase 1 plan-audit re-execution skip-eligible (PASS-WITH-DEBT 0.85 ≥ Tier L 0.85 threshold, artifacts unchanged since merged plan PR #1383).
