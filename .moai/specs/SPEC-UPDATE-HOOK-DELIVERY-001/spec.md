---
id: SPEC-UPDATE-HOOK-DELIVERY-001
title: "moai update — resolve silent non-delivery of newly-added hook entries in .claude/settings.json for existing projects"
version: "0.2.0"
status: completed
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P1
phase: "v3.0.2 target"
module: internal/cli/update
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, update, settings, hooks, merge, 3-way-merge, design-decision, detection"
related_specs: [SPEC-UPDATE-YAML-PRESERVE-001, SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001, SPEC-HOOK-CONFIG-SAFETY-001]
---

# SPEC-UPDATE-HOOK-DELIVERY-001

## §A Problem / Motivation

When the shipped template's `.claude/settings.json` gains a new hook entry **inside a hook event key the user already carries** (e.g. a new matcher block inside the user's existing `hooks.PostToolUse` array), a subsequent `moai update` never writes that entry into the user's settings.json — and nothing detects or reports the gap. New moai-adk hook capabilities shipped in later releases silently never activate in existing projects.

### Root cause (verified on tree d592b0551, branch WT-update-hook-delivery)

The defect chain is add-blind, not delete-blind. Three measured facts compose it:

1. **Base derivation skips non-map keys** — `internal/cli/update/merge/base.go:112-131` `pruneToShared` recurses only into nested `map[string]any` (`:124`). Non-map values — hook event keys hold ARRAYS — are copied wholesale from the template side (`pruned[key] = updatedVal`, `:128`). Keys absent from the user's file are excluded from the base (`:116-121`).
2. **Array-level template additions classify as "only user changed"** — `internal/merge/strategies.go:427-429` (verbatim): `case baseChanged && !updChanged: // Only user changed. result[key] = curVal`. Consequence chain: base's array equals the template's array (from fact 1) ⇒ a template-side array addition makes `updChanged=false`, `baseChanged=true` ⇒ classified "only user changed" ⇒ the user's array is kept wholesale and the template's new entry is silently dropped. (`valuesEqual` at `strategies.go:685-692` JSON-marshals both sides, so arrays deep-compare — the chain is strict.)
3. **No detection surface exists** — `internal/cli/doctor.go:768-783` `checkHooksConfig` performs one `os.Stat` on the `.claude/hooks/` directory; `.claude/settings.json` is never opened. No doctor check or update output compares the template hook set against the user's file.

### Scope precision (the defect surface, stated exactly)

| Change in shipped template | Delivered today? | Why |
|---|---|---|
| Whole NEW event-type key absent from user's file (e.g. `hooks.SessionStart` when user has none) | YES | Key excluded from base (`base.go:116-121`) ⇒ merge sees template-introduced ⇒ added |
| User DELETED an entry from a carried event key | N/A (correctly preserved) | Base still equals template ⇒ "only user changed" ⇒ deletion preserved — this is today's protective behavior any fix MUST keep |
| Addition/mutation INSIDE a carried event key's array | **NO — silently dropped, undetected** | Facts 1+2 above |

The defect surface is precisely: **additions and mutations INSIDE a hook event key's array the user already carries.**

### Test-coverage gap

`internal/template/settings_test.go` (31 test functions on this tree) exercises templates only; no test covers update-time delivery of hook entries into an existing project's settings.json. The regression this SPEC introduces guards for is entirely untested today.

## §B History

| Date | Version | Change |
|------|---------|--------|
| 2026-09-03 | 0.1.0 | Initial plan-phase artifacts (Tier M, design-decision card t466). Root cause verified against d592b0551. Resolution option left OPEN for Implementation Kickoff Approval. |
| 2026-09-03 | 0.2.0 | Plan-audit delta fixes (PASS-WITH-DEBT 0.86): §D invariant range corrected to 001..006, 011..012 (D1); AC-UHD-013 added for REQ-UHD-009, counts synced (D2); REQ-007 ambiguity disambiguated (D4). |

## §C Requirements (GEARS)

### §C.1 Invariant requirements — hold under EVERY resolution option

**REQ-UHD-001** — The `moai update` settings merge path shall continue to deliver template-introduced hook event-type keys that are absent from the user's `.claude/settings.json` (today's delivered behavior; preserved).

**REQ-UHD-002** — The `moai update` settings merge path shall not resurrect hook entries the user deliberately deleted from a carried event key (today's protective behavior; preserved under every option).

**REQ-UHD-003** — The settings merge path shall preserve every non-hook key and every user-modified hook entry already present in the user's `.claude/settings.json` (no wholesale array overwrite by the template's version).

**REQ-UHD-004** — **When** `moai update` completes against an existing project whose settings.json hook event array lacks entries present in the shipped template's same event key, the system shall resolve the gap by exactly one operator-selected mechanism — deliver the missing entries (Option A), detect and report them (Option B), or explicitly no-op with user-facing documentation (Option C). **The resolution option is an OPEN decision gated at Implementation Kickoff Approval; this SPEC carries the decision axes (§ design.md) and records no verdict.**

**REQ-UHD-005** — **When** the resolution mechanism performs a write to the user's settings.json, the written file shall remain valid JSON and shall be byte-identical outside the resolved hook-entry gap.

**REQ-UHD-006** — **When** `moai update` runs twice in sequence against the same project with no template change in between, the resolution mechanism shall be idempotent (second run reports no gap and performs no further settings.json change).

### §C.2 Option-gated requirements — bind only the operator-selected option

**REQ-UHD-007** — **Where** the operator selects Option A (deliver), the delivery mechanism shall add only hook entries the user has not previously seen or deliberately deleted — via a per-entry identity scheme (entry-identity key, tombstone marker, or seen-registry; exact scheme selected at run-phase per design.md axes) — and shall never re-add an entry recorded as user-deleted. Disambiguation: "not previously seen" means the entry's identity was never delivered to this project by a prior update (per the identity/seen record) — NOT merely absent from the user's current file; an identity never delivered to this project is template-new and IS deliverable, while an identity once delivered and now absent from the user's file is user-deleted and MUST NOT be re-added.

**REQ-UHD-008** — **Where** the operator selects Option B (detect + guide), the detection surface (`moai doctor` check or `moai update` output) shall list every missing hook entry with its event key and remediation guidance, shall be read-only, and shall not write settings.json.

**REQ-UHD-009** — **Where** the operator selects Option B (detect + guide), the detection check shall tolerate a missing or hook-free settings.json as informational (not an error state).

**REQ-UHD-010** — **Where** the operator selects Option C (explicit no-op + docs), the shipped documentation shall state the limitation, its blast radius (which hook entry classes are not delivered), and the manual remediation path.

### §C.3 Robustness requirements — hold under every option

**REQ-UHD-011** — **When** the user's settings.json is not valid JSON, or a hook event key holds a non-array value, the resolution mechanism shall report the anomaly and skip resolution gracefully (no corruption, no hard failure of `moai update`).

**REQ-UHD-012** — The resolution mechanism shall be scoped to the `.claude/settings.json` hook-entry axis via the `moai update` merge path only; the `.git/hooks/pre-commit` direct-write path (`installPreCommitHookOptional`, card t461's axis) shall remain untouched.

## §D Success Criteria

- All invariant requirements (REQ-UHD-001..006, 011..012) pass under the selected option. (REQ-UHD-007..010 are §C.2 option-gated: exactly one binds, per the operator's selection.)
- The option-gated requirement for the selected option passes; the unselected options' requirements are marked N/A in acceptance.md at run-phase entry (the decision is recorded in progress.md before M1).
- The 13 acceptance criteria in acceptance.md hold; AC-UHD-003 (the core gap-resolution AC) flips from RED to GREEN on the selected branch.
- `go test ./internal/cli/update/... ./internal/merge/...` green; new characterization tests pin the preserved behaviors (REQ-UHD-001/002/003).

## §E Out of Scope

### Out of Scope — the .git/hooks/pre-commit direct-write axis

- `installPreCommitHookOptional` and every direct write to `.git/hooks/pre-commit` — owned by card t461 (run in progress). This SPEC touches neither the function nor its callers; run-phase MUST NOT drift into it (REQ-UHD-012).

### Out of Scope — hook script file delivery under .claude/hooks/

- Delivery of hook handler script files themselves (`.claude/hooks/moai/*.sh`, `.sh`/`.sh.tmpl` pair drift) — a separate file-deployment axis, already covered by template redeployment; this SPEC is settings.json entry delivery only.

### Out of Scope — user-local hook entries

- Hook entries the user authored themselves (not present in the template) are never touched, merged, reported, or tombstoned by this SPEC's mechanism.

### Out of Scope — settings.json schema redesign

- Any change to the settings.json key layout, hook block shape, or Claude Code's own rewrite behavior of settings.json. The mechanism must tolerate whatever Claude Code itself rewrites, not redefine it.

## §F Cross-References

- design.md — the OPEN decision: Option A (deliver, with resurrection-avoidance identity schemes), Option B (detect + guide), Option C (explicit no-op + docs); trade-offs and failure modes per axis.
- research.md — verified code evidence (this tree, d592b0551) and the test-coverage gap.
- Related: SPEC-UPDATE-YAML-PRESERVE-001 (the sibling YAML-comment preservation axis of `moai update`), SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 (base-snapshot mechanics of the update merge), SPEC-HOOK-CONFIG-SAFETY-001 (hook config safety conventions).
