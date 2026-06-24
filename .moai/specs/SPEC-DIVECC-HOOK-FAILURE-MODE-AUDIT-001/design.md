# Design — SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001

> Tier M design. Defines the classification framework, the `hook-independence.md` rule structure, and the design decisions behind the audit. The WHAT/WHY lives in spec.md; this file owns the HOW-the-doctrine-is-structured.

## §A. Design intent

The paper's insight — "defense-in-depth fails when layers share failure modes" — is a *property*, not a metric. To make it actionable for moai-adk we need (1) a way to *name* a shared failure mode, (2) a way to *classify* whether a shared mode is acceptable, and (3) a *forward-looking checklist* so the next hook author does not silently add a fourth gate that shares all the existing modes. The design centers on those three artifacts.

## §B. Definition: "shared failure mode"

A **shared failure mode** is any single dependency, condition, flag, or convention whose failure correlates the degradation of **two or more** hook scripts. The unit of analysis is the *correlation*, not the dependency in isolation: a dependency that only one script has is not a shared mode; a dependency that ten scripts have but whose failure degrades them independently (not simultaneously) is a weak shared mode. The strongest shared modes are those where one condition flips many scripts at once (mode A: 31 wrappers; mode B: 3 gates).

## §C. Classification framework

Two classes, decided by a single question: **is the correlation intentional and bounded, or accidental and catastrophic?**

| Class | Decision rule | Example |
|-------|---------------|---------|
| `acceptable-by-design` | The shared mode is an *intentional* design choice, is documented, degrades *gracefully* (no-op / allow / logged skip), and a reader applying Chesterton's Fence would find a justification. | `--skip-hook` bypass (mode B): documented, audit-logged operator override. |
| `genuine-risk` | The shared mode correlates many scripts' degradation under one condition, the degradation is *silent* or *catastrophic*, and no surfaced signal warns the operator. | moai-binary chain (mode A): all 31 wrappers go silent simultaneously if moai unresolvable in all 3 tiers, with no surfaced warning. |

A `genuine-risk` classification does NOT imply "must fix in this SPEC" — it implies "must carry a mitigation *recommendation*". The mitigation may be deferred (and for N1 it always is — REQ-DIVECC-012 forbids hook edits).

## §D. `hook-independence.md` rule structure

The new rule is laid out as:

1. **Purpose** — one paragraph: why hook independence matters (the paper insight + the CLAUDE.md §17 real incident).
2. **The shared-failure-mode catalogue** — a table, one row per shared mode, columns: `mode | scope (which scripts) | evidence (command + observed) | classification | rationale`.
3. **Governance-gate cross-tab** — the per-gate dependency matrix (rows a-g of research.md §B), with the positive-signal callout (gates do not share mode A).
4. **The authoring checklist** — "Before adding a new hook, ask:" a numbered list (see §E).
5. **Cross-references** — to `hooks-system.md`, `agent-common-protocol.md` § Hook Invocation Surface, `runtime-recovery-doctrine.md` §4 (cross-reference, NOT duplicate).

The rule is **doctrine, not enforcement** — it has no lint rule or CI gate (mechanical enforcement is explicitly out of scope, spec.md §C). It changes hook-author behavior by being read, the same way `verification-claim-integrity.md` is policy-layer doctrine.

## §E. The authoring checklist (design of the checklist content)

The checklist is the forward-looking deliverable — the part that prevents recurrence. Design: each item is a yes/no the author answers before merging a new hook.

1. **Does this hook depend on the `moai` binary?** If yes, does it carry the 3-tier resolution chain + silent `exit 0` fallback (like the wrappers), or is it self-contained bash (like the governance gates)? Self-contained is more independent.
2. **Does this hook add a *new* shared condition** that, if it fails, would degrade other hooks too? (E.g. a new shared config file, a new shared binary, a new shared env var.) If yes, name it and classify it.
3. **If this hook shares the `--skip-hook` bypass**, is the bypass logged to `.moai/logs/hook-skip.log` (the audit invariant)?
4. **Does this hook degrade gracefully** (no-op / allow / logged) when its dependencies are absent, or does it crash/block? Graceful degradation keeps a shared dependency in the `acceptable-by-design` class.
5. **Is the degradation surfaced** when the shared condition fails, or silent? A silent simultaneous degradation across many hooks is the genuine-risk shape — consider a one-time probe/warning.

## §F. Design decisions + alternatives considered

- **D1 — Doctrine vs mechanical lint.** Chose doctrine (a read rule + checklist). Alternative: a meta-hook/lint that scans new hook scripts for shared-condition introduction. Rejected for N1 because (a) the paper insight is about *author judgment*, not a mechanically-decidable property (deciding "is this degradation catastrophic?" needs human classification), and (b) mechanical enforcement is a larger, separate scope (spec.md §C Out of Scope). The checklist is the cheap, high-leverage artifact; mechanical enforcement is a possible follow-up SPEC.
- **D2 — Audit lives in research.md + the rule, not a separate report.** The audit evidence is in `research.md` (SPEC-internal, may reference the SPEC) and the durable catalogue is in `hook-independence.md` (template-managed, neutral). This split honors the report-vs-spec classification (the audit of existing scripts is analysis, but its durable output is a doctrine rule, which belongs in `.claude/rules/`, not `.moai/reports/`).
- **D3 — Mode-A trigger precision.** Chose the precise "all-3-tiers-absent" trigger over the looser "not in PATH". Rationale: verification-claim-integrity.md binds risk claims as tightly as success claims; over-stating the trigger (claiming a single PATH check gates everything when there are 3 fallback tiers) would itself be an unobserved claim.
- **D4 — Cross-tab keeps per-gate granularity.** Chose to record the `jq`-dependency exception (sync-gate lacks it) rather than a layer-wide summary. Rationale: a homogenized "all gates depend on X" claim would be false for jq and would itself be a shared-mode mis-attribution — the exact error class the SPEC is auditing for.

## §G. Template-first build flow (run-phase)

```
1. author  internal/template/templates/.claude/rules/moai/development/hook-independence.md  (neutral)
2. make build                                  # regenerates internal/template/embedded.go
3. moai update (or manual copy)                # regenerates local .claude/ copy
4. verify: every new .claude/rules file has a template source
5. verify: template-neutrality CI guard passes on the new template file
```

The local `.claude/rules/.../hook-independence.md` copy MAY reference this SPEC; the template source MUST be neutral (no `SPEC-DIVECC-...`, no dates, no SHAs). The catalogue's script names + line counts are generic-mechanism content (acceptable in templates).

## §H. Cross-References

- `research.md` §B (cross-tab) / §D (classification seeds) — the input data this design organizes.
- `spec.md` §B — the requirements this design satisfies.
- `CLAUDE.local.md` §2 / §25 — the template-first + neutrality flow §G implements.
