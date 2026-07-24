---
id: SPEC-SUBAGENT-NESTING-DOCTRINE-001
title: "Subagent-nesting doctrine correction + auditor read-only nesting pilot — Plan"
version: "0.1.0"
status: in-progress
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: P2
phase: "v3.0.2 target"
module: ".claude"
lifecycle: spec-anchored
tags: "doctrine, subagent-nesting, claude-code, agent-authoring, sync-auditor"
tier: M
---

# Implementation Plan — SPEC-SUBAGENT-NESTING-DOCTRINE-001

> Sections are ordered by **decision-reversibility**: the decisions most likely to change (M2 pilot design, the double-guarantee framing) lead; mechanical Template-First mirror steps are last, so human review focuses on the highest-change-likelihood decisions.

## §A Context

Single SPEC, two milestones. **M1** corrects stale v2.1.172-era nesting facts across 7 always-loaded doctrine surfaces to the v2.1.217 reality (default-off + configurable depth + concurrency caps). **M2** enables a selective, opt-in, env-gated read-only nesting pilot on `sync-auditor` ONLY (`plan-auditor` is excluded and deferred to a future SPEC — see spec.md §E Out of Scope — plan-auditor nesting pilot). Both milestones ship in v3.0.2. The two are coupled: M1's Watch-note "double guarantee" wording and the M1 § Deprecated cross-reference both point at the M2 pilot exception, and M2's env-default-off design is precisely what keeps the M1 "flat by default" claim true even after `sync-auditor` gains the `Agent` tool. **Recommendation: keep as ONE SPEC** (splitting would force M1 to forward-reference an unauthored M2, or drop the exception wording).

Ground truth is orchestrator-verified (see spec.md §A) — do NOT re-derive; run-phase re-anchors live line numbers by content token before editing.

## §B Key Design Decisions (review first — reversibility-ordered)

### D1 — Env-gating placement: local/dev-only, never the shipped template (HIGHEST change likelihood)

The pilot is enabled by `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` (e.g. `1`). Per orchestrator recommendation this env is LOCAL/dev-only (`settings.local.json` or a documented opt-in note), **NOT** the distributed template `settings.json`. Consequence: the shipped distribution stays byte-identically flat; only a maintainer who explicitly sets the env exercises nesting. This is the load-bearing decision — it is what makes the "held-out default = flat" guarantee (REQ-SND-015 / AC-SND-012) true. Reversible: could later be promoted to a template opt-in, but NOT in this SPEC.

### D2 — Verdict ownership + read-only child enforcement (new behavior boundary)

Three HARD constraints define the M2 behavior surface and are the second-most-reviewable decisions:

1. The binding 4-dimension verdict stays owned by the top-level `sync-auditor` — never delegated to a child (REQ-SND-016).
2. Children are read-only: `Explore` (inherently read-only) or `general-purpose` + `mode: "plan"`. Because the parenthesized `Agent(agent_type)` allowlist is IGNORED inside a subagent (ground truth), read-only rests on the `mode: "plan"` parameter / `Explore` choice, NOT a type allowlist (REQ-SND-017).
3. No child, at any depth, invokes `AskUserQuestion` (REQ-SND-018) — the single-point-of-contact boundary is depth-independent.

### D3 — M1 "double guarantee" framing + the sync-auditor exception wording

After M2, `sync-auditor` WILL carry `Agent` in `tools`, so the blanket claim "MoAI retained agents do not list `Agent`" becomes partially false. The Watch-note rewrite (REQ-SND-002) must state: flat hierarchy holds by BOTH runtime-default-off AND `Agent`-omission for the 10 non-pilot agents; for the `sync-auditor` pilot exception the flat default rests on the **env-default-off guarantee alone**. This framing decision propagates to `agent-patterns.md` § Deprecated (REQ-SND-007) and `agent-authoring.md` § Tool Permissions (REQ-SND-006).

### D4 — zone-registry conditional re-sync (likely a no-op)

CONST-V3R2-020 mirrors the CLAUDE.md §14 background/concurrency clause; CONST-V3R2-044 mirrors the agent-common-protocol background clause. Neither is about nesting. Decision: re-sync the zone-registry entry **only if** the M1 §14 concurrency-cap sentence is authored *inside* the mirrored clause span; otherwise leave both entries untouched (REQ-SND-009). Recommended authoring: add the concurrency caps as a NEW sentence/bullet distinct from the mirrored clause, so no re-sync is triggered.

### D5 — M2 pilot scope: sync-auditor ONLY (RESOLVED)

The M2 pilot scope is **`sync-auditor` only**. `plan-auditor` is explicitly excluded from this SPEC; its read-only nesting pilot is DEFERRED to a future SPEC (spec.md §E Out of Scope — plan-auditor nesting pilot). Rationale: `plan-auditor` has `permissionMode: default` (not `plan`), so its read-only child scoping would need an explicit `mode: "plan"` — a distinct design the future SPEC will own. REQ-SND-020 is now an Unwanted requirement asserting `plan-auditor` is untouched (no `Agent` tool, no verifier-child documentation); no REQ/AC grants `plan-auditor` the `Agent` tool.

### D6 — Target release: M1 + M2 both ship in v3.0.2 (RESOLVED)

Both milestones ship in **v3.0.2** (single release). M1 is a pure doc correction; M2 is opt-in and behavior-inert at the shipped default. The shipped default stays flat: `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` is ABSENT from the distributed template `settings.json` (env-opt-in only, local/dev), so no user project receives nesting on by default.

## §C Constraints

- Plan-phase only — this SPEC DESCRIBES the edits; manager-develop performs them in run-phase.
- Template-First (CLAUDE.local.md §2): every edit to `CLAUDE.md` / `.claude/rules/**` / `.claude/agents/moai/**` is mirrored to `internal/template/templates/` + `make build`.
- Template neutrality (CLAUDE.local.md §25): only the generic v2.1.217 version facts enter template content; no SPEC ID / date / SHA / audit citation.
- Scope discipline: `model-policy.md` is NOT touched (model/effort out of scope).
- `AskUserQuestion` boundary preserved at every depth.

## §D Milestones (mechanical mirror steps last within each)

### M1 — Documentation correction (no runtime behavior change)

1. Re-anchor each stale surface by content token (line numbers drift): CLAUDE.md §4 Watch note; CLAUDE.md §14; agent-authoring §Agent(agent_type) Restrictions / §Fork Subagents / §Tool Permissions; agent-patterns §Deprecated; orchestration-mode-selection §Mode 6.
2. Rewrite each surface to the v2.1.217 facts per REQ-SND-001..008 (Watch note carries the double-guarantee + M2-exception reference; §14 gains the two concurrency caps).
3. Evaluate D4: determine whether the §14 concurrency-cap sentence lands inside the CONST-V3R2-020 mirrored clause span; re-sync zone-registry only if so (REQ-SND-009).
4. (mechanical) Mirror every edited live file to `internal/template/templates/` byte-for-byte; run `make build` (REQ-SND-010).
5. (mechanical) Verify template neutrality + mirror parity: `go test ./internal/template/...` (neutrality + mirror-parity tests) green (REQ-SND-011).

### M2 — Auditor read-only nesting pilot (opt-in, env-gated)

1. Add `Agent` to `sync-auditor` `tools`; keep `permissionMode: plan` (REQ-SND-013).
2. Document the read-only per-dimension verifier pattern in the `sync-auditor` body: one child per Functionality/Security/Craft/Consistency dimension, `Explore` or `general-purpose` + `mode: "plan"`, with HARD constraints REQ-SND-016/017/018 stated inline (REQ-SND-021).
3. Assert the env is local/dev-only: confirm the distributed template `settings.json` does NOT set `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH`; document the opt-in path (REQ-SND-019).
4. Confirm `plan-auditor` is untouched: NO `Agent` tool added, body unmodified (REQ-SND-020 Unwanted; the `plan-auditor` pilot is deferred to a future SPEC per D5).
5. (mechanical) Mirror `sync-auditor.md` to template; `make build`; `go test ./internal/template/...` green (REQ-SND-022).
6. Commit plan artifacts to `release/v3.0.2-prep` (Tier M Hybrid Trunk 1-person OSS; no PR at plan-phase).

## §E Self-Verification (plan-phase)

- [ ] SPEC ID `SPEC-SUBAGENT-NESTING-DOCTRINE-001` passes `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (self-check printed in the return).
- [ ] All 7 M1 surfaces confirmed to exist with template mirrors (verified 2026-07-24).
- [ ] `sync-auditor` `permissionMode: plan` + tools-without-`Agent` confirmed (frontmatter read).
- [ ] `plan-auditor` `permissionMode: default` confirmed (frontmatter read).
- [ ] zone-registry CONST-V3R2-020/044 clause text confirmed to be background/concurrency (NOT nesting) — re-sync is conditional.
- [ ] Out of Scope section present (6 `### Out of Scope — <topic>` H3 sub-headings, incl. plan-auditor nesting pilot).
- [x] Clarification markers RESOLVED at plan finalization — D5: M2 pilot = `sync-auditor` only (`plan-auditor` deferred, spec.md §E Out of Scope); D6: M1 + M2 both ship in v3.0.2 (shipped default flat, env opt-in only). 0 open markers remain.

## §F Risks & Anti-Patterns

- **AP-1 — blanket "no agent lists Agent" claim after M2**: adding `Agent` to `sync-auditor` silently falsifies the Watch-note blanket claim if the exception wording is omitted. Mitigation: REQ-SND-002 mandates the exception clause.
- **AP-2 — env leaking into template settings.json**: would enable nesting for all users. Mitigation: REQ-SND-019 + AC-SND-012/AC-SND-016 grep the template `settings.json` for the env → 0.
- **AP-3 — write-capable child via bare general-purpose**: a `general-purpose` child spawned WITHOUT `mode: "plan"` is write-capable. Mitigation: REQ-SND-017 mandates `mode: "plan"` or `Explore`; AC-SND-014 asserts the body constraint.
- **AP-4 — line-number-anchored edits drift**: the ~L64/L94/L100/L226/L314/L66 anchors are 2026-07-24 reads. Mitigation: M1 step 1 re-anchors by content token.
- **AP-5 — mirror drift / neutrality leak**: mechanical mirror steps can leave live/template out of parity or leak internal content. Mitigation: `go test ./internal/template/...` gate (M1 step 5, M2 step 5).
- **AP-6 — over-triggering zone-registry re-sync**: editing the CONST clauses when the nesting facts do not touch them would be an out-of-scope edit. Mitigation: D4 keeps re-sync conditional and recommends a distinct concurrency-cap sentence.

## §G Cross-References

- spec.md §A (ground truth + affected-surface table).
- acceptance.md (AC-SND-001..016).
- `code.claude.com/docs/en/sub-agents` § Spawn nested subagents.
- CLAUDE.local.md §2 / §15 / §25 (Template-First, neutrality, isolation).
