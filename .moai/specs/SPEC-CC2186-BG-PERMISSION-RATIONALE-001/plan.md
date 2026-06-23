# Implementation Plan — SPEC-CC2186-BG-PERMISSION-RATIONALE-001

> Plan-phase artifact. Tier M (standard). Run-phase awaits the Implementation Kickoff Approval human gate.

## §A. Context

A single verified CC 2.1.186 upstream-alignment drift: the background-subagent permission *rationale* in MoAI doctrine ("auto-deny ... because they cannot interact with the user") is now factually outdated, while the [HARD] behavioral *conclusion* (`run_in_background: false` for writes) remains valid on a separate basis. This is a doc-prose correction across 6 surfaces (3 loci × live + template-mirror), with `make build` regeneration and mirror-parity verification.

Source research: `.moai/research/cc-update-2.1.183-to-2.1.186.md` (T1-1, the single GENUINE-DRIFT item). Sibling pattern (same shape, completed): SPEC-CC2178-TEAM-API-ALIGN-001.

## §B. Tier Justification — Tier M (standard)

Tier M is justified NOT by algorithmic complexity (the change is prose correction) but by the **multi-surface consistency burden**:

- 6 coordinated surfaces (3 loci × 2 trees) that must end semantically identical.
- Template-mirror parity (`embedded_mirror_test.go` byte-identity invariant) + `make build` regeneration of `embedded.go`.
- zone-registry CONST-V3R2-020 verbatim-clause sync (the registry stores the clause text literally).
- An edit-time official-doc re-confirmation precondition (the [HARD] rationale rewrite must not derive from the terse changelog bullet alone).
- A conclusion-retention invariant that must hold in every surface (the correction must not accidentally relax the restriction).

A single-surface 1-line edit would be Tier S; the cross-tree + CONST-clause + doc-reconfirm coordination raises it to Tier M.

## §C. Known Issues / Risks

| Risk | Mitigation |
|------|-----------|
| Rationale rewrite derived from terse changelog bullet alone → inaccurate [HARD] rationale | §G.1 run-phase precondition: re-confirm against `code.claude.com/docs/en/sub-agents` BEFORE writing (REQ-BGR-008) |
| Accidentally relaxing the [HARD] restriction while editing rationale | Conclusion-retention invariant (REQ-BGR-003); AC grep verifies `run_in_background: false` directive still present in all surfaces |
| Editing the out-of-scope CONST-V3R2-044 (pure conclusion) | spec.md §F Exclusions explicitly names it KEEP-AS-IS; AC verifies its clause is unchanged |
| Live/mirror prose divergence after edit | REQ-BGR-005 + AC parity grep (per-locus diff live vs mirror) |
| Forgetting `make build` after template edit | REQ-BGR-007 + AC verifies `embedded.go` reflects corrected mirror |
| Template-neutrality leak (SPEC-ID/date/SHA in template prose) | REQ-BGR-006 + `TestTemplateNeutralityAudit` / `internal_content_leak_test.go`; rationale text is generic mechanism description |
| Line numbers drift between plan and run | manager-develop greps for the drift token at run time rather than trusting plan-phase line numbers |

## §D. Constraints

- **Template-First (CLAUDE.local.md §2)**: edit `internal/template/templates/**` source, then `make build`. Both live and mirror copies edited in the same commit (embedded_mirror_test.go byte-parity).
- **Template neutrality (§2.1 / §15 / §25)**: corrected prose in templates carries no forbidden internal-content class.
- **Scope discipline**: touch ONLY the 6 enumerated surfaces. CONST-V3R2-044, the [HARD] rule itself, and the unrelated "3-phase close" mirror drift are explicitly excluded.
- **Verification-claim integrity**: every AC PASS in run/sync must be a real observed grep/test output, recorded in progress.md §E.2/§E.3.
- **No time estimates** (priority ordering only).

## §E. Self-Verification (plan completeness)

- [x] SPEC ID Pre-Write Self-Check Protocol decomposition printed with → PASS (in the agent response body)
- [x] Directory format `.moai/specs/SPEC-CC2186-BG-PERMISSION-RATIONALE-001/`
- [x] ID uniqueness verified (collision-free, `ls` confirmed FREE)
- [x] 4 plan-phase artifacts authored (spec/plan/acceptance/research) + progress.md §E skeleton
- [x] GEARS requirements (no legacy IF/THEN)
- [x] Exclusions section present with `### Out of Scope — <topic>` H3 sub-headings + `-` bullets
- [x] No implementation details in spec.md (WHAT/WHY only; HOW deferred to run)
- [x] Frontmatter 12-canonical-field schema validated
- [x] `created`/`updated` used (NOT `created_at`/`updated_at`); `tags` comma-separated string (NOT `labels` array)

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Official-doc re-confirm + draft the corrected rationale (Priority High, BLOCKING precondition)

- Re-confirm the exact 2.1.186 background-permission surface against `code.claude.com/docs/en/sub-agents` (REQ-BGR-008). Capture the verbatim surface wording as evidence.
- Draft the corrected rationale prose for each of the 3 loci, conveying: (a) 2.1.186 surfaces prompts to the main session rather than auto-denying; (b) MoAI retains `run_in_background: false` for writes because the background context does not fully inherit the parent allowlist AND surfacing a prompt per background write defeats parallel-execution purpose; (c) read-only stays safe in background.
- Output: a single corrected-prose draft per locus, ready to apply identically to live + mirror.
- Gate: M1 MUST complete (doc re-confirmed) before M2/M3 edits. Do NOT edit any surface from the changelog bullet alone.

### M2 — Apply to live tree (3 loci) (Priority High)

- Edit live Locus 1 (`CLAUDE.md` §14 L289 — adjust "auto-deny" descriptor, retain `run_in_background: false` directive).
- Edit live Locus 2 (`.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution — remove "because they cannot interact with the user" clause, retain allowlist-non-inheritance sentence + [HARD] line).
- Edit live Locus 3 (`.claude/rules/moai/core/zone-registry.md` CONST-V3R2-020 clause — adjust "auto-deny Write/Edit operations" descriptor, retain `run_in_background: false` directive).
- Leave CONST-V3R2-044 untouched (Exclusions).

### M3 — Apply to template-mirror tree (3 loci) + make build (Priority High)

- Edit mirror Locus 1 (`internal/template/templates/CLAUDE.md` L289) — byte-identical prose to live Locus 1.
- Edit mirror Locus 2 (`internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution) — byte-identical to live Locus 2.
- Edit mirror Locus 3 (`internal/template/templates/.claude/rules/moai/core/zone-registry.md` CONST-V3R2-020 clause) — byte-identical to live Locus 3.
- Run `make build` to regenerate `internal/template/embedded.go` (REQ-BGR-007).

### M4 — Parity + grep AC verification (Priority High)

- Grep-verify the drift clause "cannot interact with the user" is gone across all 6 surfaces (== 0).
- Grep-verify the `run_in_background: false` directive (conclusion-retention) still present in all corrected surfaces.
- Grep-verify CONST-V3R2-044 clause unchanged (both trees).
- Per-locus diff live vs mirror → byte-identical prose (REQ-BGR-005).
- Run `go test ./internal/template/... -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak'` (neutrality; the per-locus diff above is the mirror-parity authority — no Go test covers the 3 loci, `TestEmbeddedMirror` does not exist).
- Run `go build ./...` (embedded.go compiles) + `go test ./...` (no cascading failure).
- Record all evidence verbatim in progress.md §E.2 (run evidence) and §E.3 (run audit-ready signal).

## §G. Anti-Patterns to Avoid

### G.1 Run-phase precondition (LOAD-BEARING)

- Do NOT rewrite the [HARD] rationale from the changelog bullet alone — re-confirm against the official sub-agents doc first (REQ-BGR-008). The changelog wording is terse; the [HARD] rationale must rest on the actual documented surface.

### G.2 Scope-creep anti-patterns

- Do NOT touch CONST-V3R2-044 (pure conclusion, Exclusions).
- Do NOT relax the [HARD] background-write restriction — this is rationale correction, NOT rule re-evaluation.
- Do NOT "fix" the unrelated "3-phase close" mirror drift in spec-frontmatter-schema.md (separate concern, Exclusions).
- Do NOT add the optional T2-1/T2-2 awareness notes unless the maintainer explicitly requests them at run time (research recommends skip).

### G.3 Mirror / build anti-patterns

- Do NOT edit `embedded.go` directly (generated; `make build` only).
- Do NOT edit only the live copy and forget the mirror (or vice-versa) — same commit, byte-identical.
- Do NOT leak any SPEC-ID / date / SHA / REQ token into the template-tree prose.

## §H. Cross-References

- Research: `.moai/research/cc-update-2.1.183-to-2.1.186.md` (T1-1 GENUINE-DRIFT)
- Sibling pattern: SPEC-CC2178-TEAM-API-ALIGN-001 (completed CC-version doctrine-alignment SPEC, same shape)
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Template-First / neutrality: CLAUDE.local.md §2 / §2.1 / §15 / §25
- Mirror parity invariant: `internal/template/embedded_mirror_test.go`
- Official doc to re-confirm at run time: `code.claude.com/docs/en/sub-agents`
- Verification-claim integrity: `.claude/rules/moai/core/verification-claim-integrity.md`
