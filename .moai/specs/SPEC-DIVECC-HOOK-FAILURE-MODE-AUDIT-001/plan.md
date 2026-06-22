# Implementation Plan — SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001

> Tier M. Derived from `spec.md` §B requirements. Priority-ordered milestones; no time estimates (per `agent-common-protocol.md` § Time Estimation).

## §A. Context

N1 entry SPEC of Epic Dive-into-CC. Delivers an audit of moai-adk's hook layer for shared failure modes + a new independence doctrine rule. The shared-failure-mode premise is already VERIFIED (research.md §A) — run-phase re-confirms the evidence is current, then writes the catalogue and rule. **No hook script is modified.**

## §B. Known Issues / Constraints from research

- SPEC-ID validity (run-phase note, pre-empts a future false-flag): the multi-segment id `SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001` is VALID per the ENFORCED lint regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (`internal/spec/lint.go:578`), which admits one-or-more domain segments. The stale schema-prose regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` in `spec-frontmatter-schema.md` would APPEAR to reject the multi-segment id, but the enforced lint passed clean (`moai spec lint` → 0 findings). Do NOT re-flag this ID — the prose regex is the stale one, not the ID.
- The wrapper "single condition" framing is a 3-tier chain (research.md §A.2 refinement) — the run-phase audit MUST use the precise trigger (all-3-tiers-absent), not "not in PATH".
- The `--skip-hook` shared bypass is by-design and audit-logged — classify acceptable-by-design, do not flag as defect (Chesterton's Fence).
- The governance gates are NOT a homogeneous layer: `sync-phase-quality-gate.sh` lacks the `jq` dependency the other two carry. The catalogue must reflect per-gate, not layer-wide, dependency facts (research.md §B cross-tab).

## §C. Pre-flight (run-phase, before writing the rule)

1. Re-run the §A.1–§A.3 grep/read evidence to re-confirm currency (baseline re-attribution per verification-claim-integrity.md §2). The hook directory may have changed since plan-phase; the audit must measure the current tree.
2. Confirm the canonical SSOT surfaces the rule will cross-reference still exist at their cited paths (`hooks-system.md`, `agent-common-protocol.md` § Hook Invocation Surface, `runtime-recovery-doctrine.md` §4).

## §D. Constraints

- **Template-First (REQ-DIVECC-013)**: the new rule `hook-independence.md` MUST be authored in `internal/template/templates/.claude/rules/moai/development/hook-independence.md` FIRST, then `make build`, then the local `.claude/` copy is regenerated. Do NOT create the local copy directly without the template source (per CLAUDE.local.md §2). Verify every new file under `.claude/rules/` has a corresponding template source file before committing.
- **Template-Neutrality (REQ-DIVECC-014)**: the TEMPLATE copy of `hook-independence.md` must not leak internal SPEC IDs (`SPEC-DIVECC-...`), internal dates, or commit SHAs (CLAUDE.local.md §25 / CI guard `template-neutrality-check.yaml`). The audit's concrete evidence (script names, line counts) is generic-mechanism content and is acceptable; SPEC-internal provenance is not. The local `.claude/` copy MAY reference this SPEC.
- **No hook-script edits (REQ-DIVECC-012)**: zero modifications to any `.claude/hooks/moai/*.sh`. Audit + doctrine only.
- **SSOT-no-duplication (REQ-DIVECC-010)**: the rule cross-references canonical hook surfaces; it does not copy their content.

## §E. Self-Verification (run-phase deliverables map)

| REQ | Verified by |
|-----|-------------|
| REQ-DIVECC-001/002 | catalogue enumerates every shared mode with cited grep/read evidence (grep re-run output pasted) |
| REQ-DIVECC-003 | each mode carries `acceptable-by-design` / `genuine-risk` + rationale |
| REQ-DIVECC-004 | governance-gate cross-tab (a)-(g) present in the rule/audit |
| REQ-DIVECC-005 | wrapper fallback-branch finding recorded (3-tier chain) |
| REQ-DIVECC-006 | positive signal (gates do not share mode A) documented |
| REQ-DIVECC-007/008 | `hook-independence.md` exists with catalogue + authoring checklist |
| REQ-DIVECC-009 | acceptable-by-design modes carry design justification |
| REQ-DIVECC-011 | each genuine-risk mode carries a mitigation recommendation OR explicit "none, because…" |
| REQ-DIVECC-012 | `git diff --stat .claude/hooks/moai/` shows zero changes |
| REQ-DIVECC-013 | template source file exists; `make build` ran; local copy regenerated |
| REQ-DIVECC-014 | template-neutrality CI guard passes on the template copy |

## §F. Milestones (priority-ordered)

- **M1 — Re-confirm evidence**: re-run the research.md §A grep/read commands against the current tree; record the re-confirmed output. Diff against plan-phase evidence; note any change.
- **M2 — Build the shared-failure-mode catalogue**: enumerate every shared mode (A, B, C…) with cited evidence and the governance-gate cross-tab (REQ-DIVECC-001/002/004/005/006).
- **M3 — Classify each mode**: acceptable-by-design vs genuine-risk + rationale; design justification for the by-design ones (REQ-DIVECC-003/009).
- **M4 — Record mitigation recommendations**: for each genuine-risk mode, a recommendation (or explicit "none") — recommendation only, no hook edit (REQ-DIVECC-011/012).
- **M5 — Author the rule (template-first)**: write `internal/template/templates/.claude/rules/moai/development/hook-independence.md` (catalogue + authoring checklist + cross-references), `make build`, regenerate local copy (REQ-DIVECC-007/008/010/013/014).
- **M6 — Self-verify + close**: run §E verification matrix; confirm zero hook-script diff; confirm template-neutrality guard passes.

## §G. Anti-Patterns to avoid

- AP — re-asserting mode A as "single condition / not in PATH" instead of the 3-tier-absent trigger (over-stating the risk; violates verification-claim-integrity.md bidirectional invariant).
- AP — flagging the `--skip-hook` bypass as a defect (it is documented, audit-logged, by-design).
- AP — duplicating `hooks-system.md` content into `hook-independence.md` instead of cross-referencing (SSOT violation).
- AP — creating the local `.claude/rules/.../hook-independence.md` without the template source first (Template-First violation).
- AP — modifying a hook script "while we're in there" (scope-discipline violation; REQ-DIVECC-012).

## §H. Cross-References

- `research.md` — the verified evidence + classification seeds.
- `acceptance.md` — Given-When-Then ACs.
- `design.md` — the rule structure + classification framework.
- `CLAUDE.local.md` §2 / §25 — template-first + neutrality run-phase constraints.
