# Progress — SPEC-HARNESS-MCP-PROVISION-001

> Lifecycle progress ledger. §E is parser-load-bearing (era.go string-matches the
> literal `§E.2` / `§E.3` / `§E.4` heading tokens + `sync_commit_sha`). Do NOT rename
> the §E.N headings. Plan-phase populates §E.1 only; §E.2-§E.4 are placeholder
> headings owned by manager-develop (run) and manager-docs (sync).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-11
plan_version: 0.1.3
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 11
ac_count: 16
depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]
depends_on_state: "SATISFIED — SPEC-PROJECT-HARNESS-BRIDGE-001 status: completed (measured 2026-07-11). Phase 0.5 depends_on gate does NOT block run-phase entry."
open_clarifications: []
resolved_clarifications:
  - marker: "mcp-matrix config surface"
    decision: "standalone template DATA RESOURCE (internal/template/templates/.moai/config/sections/mcp-matrix.yaml, read as prose-context by project/doc-generation.md) — NOT a new Go config section; no typed loader / struct field. Keeps the SPEC doc/config-only."
  - marker: "doctor manifest-mcp validate-vs-tolerate"
    decision: "TOLERATE-ONLY, zero Go change. Verified: doctor.go + applier.go use plain json.Unmarshal with no DisallowUnknownFields; v4manifest/validate.go checks only required fields. AC-HMP-010 encodes documented-tolerance grep + regression guard grep -c DisallowUnknownFields internal/harness/v4manifest/*.go == 0. Active mcp-block validation deferred to a follow-up Go SPEC."
notes: >
  SPEC 2 of the 3-SPEC Project-Harness Pipeline Epic. Doc/config-only (markdown+yaml);
  no Go code. Phase 0.5 Depends_on pre-flight is SATISFIED —
  SPEC-PROJECT-HARNESS-BRIDGE-001 is status: completed (measured 2026-07-11); the gate
  does NOT block run-phase entry. (v0.1.0-v0.1.2 claimed BRIDGE-001 was "currently draft,
  so this gate WILL block" — an unobserved state claim, corrected at v0.1.3 per
  verification-claim-integrity.md §1.1.)
  Plan-audit fixes applied at v0.1.1: both clarifications resolved (markers struck from
  plan.md); AC-HMP-014 added for REQ-HMP-003 (write-on-approval coverage gap);
  AC-HMP-010 de-vacuumed (grep+guard, dropped repo-wide doctor smoke); Epic artifact
  numbering corrected (MCP fragment = artifact 7 OPTIONAL; verify skill = artifact 6
  mandatory, SPEC-HARNESS-VERIFY-PROMOTE-001); AC-HMP-015 added for the harness-builder
  "exactly 5" prose reconciliation.
  v0.1.2: acceptance.md §C hardened against the token-presence-vs-reachability failure
  mode — 7 sub-checks were measurably vacuous (passing at baseline with ZERO
  implementation: AC-002=6, AC-003=7, AC-005=2, AC-007 additive=2 / env-var=1,
  AC-009=8 (headline conditional-emission!), AC-010=1); every AC now carries a positive
  check that FAILS on the unmodified tree, compound-alternation-as-sole-evidence is
  eliminated (one grep per clause), positional clauses are verified inside
  heading-delimited section ranges (§C.0 extractors), and two broken commands are fixed
  (multi-file `grep -c` printed file:count pairs, never a scalar; parity `diff` ran
  without an existence precondition). REQ set, AC↔REQ mapping, GWT scenarios, and scope
  are UNCHANGED (11 REQ / 15 AC) — only HOW each AC is verified changed.
  v0.1.3 (plan-audit iter-2, 5 fixes): D1 MUST-FIX — the /moai project ROUTER
  (.claude/skills/moai/workflows/project.md) was entirely OUT OF SCOPE (measured:
  project.md appeared 0x in spec.md, 0x in acceptance.md, 0x in plan.md §F milestones).
  An implementation passing all 15 prior ACs would have left the router still documenting
  "Phase 3.5 -> Phase 3.7" with Phase 3.6 nowhere in it — the same cross-file
  reachability gap this SPEC exists to close, one file over. Precedent confirms it is not
  hypothetical: grep -c "Phase 3.2" project.md == 0, i.e. BRIDGE-001 already shipped this
  exact drift. Fixes: plan.md §F M2 gains the router (+ template mirror) as an edit target
  with BOTH router surfaces named (Phase Routing Table row + Invocation Flow diagram
  line); user-approved scope addition backfills the missing Phase 3.2 row + diagram line
  (description READ from doc-generation.md, not invented); new AC-HMP-016 pins router
  registration (Phase 3.6 >= 2 occurrences, mechanical 3.5<3.6<3.7 row ordering, Phase 3.2
  backfill >= 2, sub-skill pointer); AC-HMP-011 parity loop extended to the router pair
  (3 -> 4 files). D2 — AC-HMP-012's prohibition filter did not cover the parenthetical
  negation form "(NOT `.moai/specs/`)" that plan.md §F M2 step 2 itself prescribes, so an
  implementer copying plan.md's own wording would false-FAIL; filter widened with |NOT
  and re-verified against 3 controls. D3 — AC-HMP-015 checked only 1 of 6 stale
  "5 artifact types" sites, leaving the H2 heading "(the 5 artifact types)" standing above
  a 7-artifact section while the AC PASSED; added a second reverse delta on the heading
  and enumerated all 6 sites in plan.md §F M3 step 1. D4 — struck the stale
  "moai harness doctor" from acceptance.md §E Tested bullet (contradicted AC-HMP-010 and
  §E's own Doctor-tolerance bullet; the smoke was dropped at v0.1.1). D5 — corrected the
  false BRIDGE-001 "currently draft" dependency-state claim (it is completed). D6 —
  AC-HMP-004/013 row greps now guard the missing-file case (grep -c on an absent file
  emits nothing and exits 2 — not a scalar; same non-evaluable class fixed in AC-HMP-010).
  REQ set unchanged (11 REQ); AC count 15 -> 16.
```

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs; carries sync_commit_sha>_
