# Implementation Plan — SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001

> Ordered by decision-reversibility: the highest-change-likelihood decision (the neutralization strategy) leads; mechanical steps (embed, verify) follow.

## §A. Highest-change-likelihood decision — Neutralization strategy (A vs B)

This is the crux of the SPEC and the one decision most likely to change under review. It is **resolved** in favor of Strategy A; the alternative is recorded so a reviewer can reverse it cheaply.

### Strategy A — sanitized-pair (neutralize the mirror only) — CHOSEN

- The template mirror is a §25-neutralized derivative (4 tokens generalized/removed); the local source is untouched; the divergence is registered in `sanitizedPairPaths`.
- **Pros**: preserves the live source's internal provenance; smallest blast radius (adds a mirror + 1-line test entry, edits no live doctrine); consistent with the `runtime-recovery-doctrine.md` sibling precedent (the very doctrine the 2 forbidden SPEC-IDs reference).
- **Cons**: introduces one more divergent pair to track (mitigated by `TestSanitizedPairParity` registration — REQ-ZRP-006).

### Strategy B — neutralize-in-place then byte-identical mirror — REJECTED

- Edit the source to drop the 4 tokens, then copy byte-identically and enroll in `workflowOptMirroredPaths`.
- **Pros**: single content; simplest parity story (byte-identity).
- **Cons**: erases legitimate provenance from the maintainer's own live constitution registry; larger blast radius (modifies a 40 KB CLI-loaded runtime file); diverges from the established sanitized-pair treatment of the sibling recovery-doctrine files.

**Decision rationale**: sibling precedent + provenance preservation + scope discipline (see spec.md §A.5). No `[NEEDS CLARIFICATION]` marker — all SIX guard mechanics (spec.md §A.4.1: leak scope gates, neutrality file-allowlist, byte-parity opt-in allowlist, sanitized-pair token normalization, governance-token per-(path,token) hook, date pattern breadth) were read directly and confirm Strategy A + the two additive guard-test edits pass all six guards without touching the source doctrine.

### Neutralization detail (run-phase guidance, not prescriptive wording)

| Token | Source line | Neutralization approach |
|-------|-------------|-------------------------|
| `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` | 1009 (comment) | Reword the comment to drop the ID, e.g. "(first V3R6 modern-era entry)". |
| `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001` | 1017 (`clause:` value) | Reword the clause tail to "…mechanical enforcement deferred to a future SPEC" — drop the ID, keep the rule meaning intact. |
| `2026-05-04` | 629 (comment) | Remove the date; keep the descriptive comment ("session-handoff.md HARD clauses (new workflow rules)"). |
| `2026-05-09` | 630 (comment) | Remove the date; keep "model-specific threshold revision" descriptive text. |

Exact wording is a run-phase choice; the constraint is REQ-ZRP-001 (zero leak tokens) + REQ-ZRP-006 (token-normalized doctrine stays within `TestSanitizedPairParity` tolerance — since the only divergence is the 4 normalized tokens, the normalized copies collapse to identical/near-identical, well within the 4-line reword tolerance).

## §B. Technical approach

1. Copy the source into the mirror path, then apply the four neutralization edits (per §A table).
2. Register the sanitized pair (1-line addition to `sanitizedPairPaths`).
3. Add the file-level governance-token exemption (`governanceTokenFileAllowlist` single-path entry for `zone-registry.md`) in `rule_provenance_audit_test.go`, gated in the governance-token branch — the D1 fix (spec.md §A.4.2). The 119 legitimate `CONST-V3R` occurrences are content §C forbids removing, and `isPedagogicallyAllowed` is per-(path,token) only, so a file-level carve-out is the mechanism.
4. `make build` to re-embed.
5. Verify in a real deployed project (`moai init` into a temp dir).

Template-First cycle (`CLAUDE.local.md` §2): the change is template-source + `make build`, verified in a deployed project — never a direct local-project edit.

## §C. Milestones (priority-ordered, reversibility-first)

### M1 — Neutralized mirror creation (highest change-likelihood)

- Create `internal/template/templates/.claude/rules/moai/core/zone-registry.md` as the Strategy-A neutralized derivative of the source.
- Apply the four neutralization edits (§A table).
- Self-verify: `grep -nE '\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b|\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b' <mirror>` → zero matches; all 115 `CONST-*` entries still present (`grep -cE '^- id: CONST-' <mirror>` == source count).

### M2 — Guard-test edits (sanitized-pair registration + governance-token file-level exemption)

- Add `.claude/rules/moai/core/zone-registry.md` to `sanitizedPairPaths` in `internal/template/sanitized_pair_parity_test.go`, with a comment mirroring the `runtime-recovery-doctrine.md` entry rationale (REQ-ZRP-006).
- Add the `governanceTokenFileAllowlist` single-path map + gate it in the governance-token branch of `TestRuleProvenanceAudit`'s walk loop in `internal/template/rule_provenance_audit_test.go` (REQ-ZRP-007; exact edit in spec.md §A.4.2). Parallel to the existing file-level `allowListSet(...)` for `zone-registry.md` in `template_neutrality_audit_test.go:137`.
- Self-verify (single-path constraint): `grep -A3 'governanceTokenFileAllowlist = map' <file>` shows exactly one path key; the `TestRuleProvenanceRecurrenceBackstop` raw-regex CONST-V3R6-099 case remains untouched (guard still fires elsewhere).

### M3 — Embed + deployed-project verification (mechanical)

- `make build` (re-embed).
- `moai init` into `$(mktemp -d)`; assert the file lands at `.claude/rules/moai/core/zone-registry.md`.
- Run `moai doctor`, `moai constitution list`, `moai spec lint` in the temp project; assert none emit the "not found" break (doctor check not Warn-for-missing; constitution list returns entries; spec lint does not skip the registry dimension).

### M4 — Full guard suite (mechanical)

- `go test ./internal/template/...` (all SIX guards green — leak, neutrality, byte-parity, sanitized-pair, `TestRuleProvenanceAudit`, `TestRuleDateProvenance`).
- `MOAI_TEMPLATE_LEAK_STRICT=1 go test -run TestTemplateNoInternalContentLeak ./internal/template/...` (strict-tier date classes also clean).
- Pin the two previously-missed guards explicitly: `go test -run 'TestRuleProvenanceAudit|TestRuleDateProvenance' ./internal/template/...`.
- `go build ./...` sanity.

## §D. Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Neutralization reword drifts doctrine beyond `TestSanitizedPairParity` tolerance | Low | Only 4 tokens change; normalized copies collapse to near-identical (≤4-line reword tolerance). M2/M4 verify. |
| A hidden forbidden token missed by the §A.4 pre-scan | Very Low | M4 runs the actual guards (default + strict) on the mirror — mechanical, not grep-by-eye. |
| Deployed-project verification env lacks `moai` on PATH | Low | Use the freshly-built binary path from `make build`; `moai init` uses the same binary. |
| Future source doctrine edit silently fails to reach the mirror | Addressed | REQ-ZRP-006 registration makes `TestSanitizedPairParity` catch it. |
| Governance-token file-level exemption over-broadens (blinds guard to real leaks) | Low | Single-path map entry, governance-token-class-only; AC-ZRP-007 pins single-path + still-fires-elsewhere. Does NOT touch REQ/AC or lessons-W classes (both 0 in source). |
| Missed a 5th/6th guard again | Very Low | All six guards enumerated in spec.md §A.4.1 by direct read; M4 runs the actual package, not a grep. |

## §E. Self-verification (plan-phase)

- [x] Defect + consumption sites grounded by reading `constitution.go` / `doctor.go` / `spec_lint.go` / `deployer.go` / `embed.go`.
- [x] ALL SIX CI guards read in full: `internal_content_leak_test.go` (C1 whole-tree + C1b skill-body-scoped confirmed), `template_neutrality_audit_test.go` (zone-registry already file-allowlisted at C2), `rule_template_mirror_test.go` (not enrolled), `sanitized_pair_parity_test.go`, `rule_provenance_audit_test.go` (D1 — governance-token, `isPedagogicallyAllowed` per-(path,token)), `rule_date_provenance_audit_test.go` (D2 — 2 dates only).
- [x] Governance-token exemption mechanism confirmed: no file-level hook exists; run-phase adds a single-path `governanceTokenFileAllowlist`, parallel to the neutrality C2 `allowListSet`.
- [x] Source-token counts verified: 119 `CONST-V3R` occurrences, 2 SPEC-IDs (1009/1017), 2 ISO dates (629/630), 0 REQ/AC/lessons-W/MIG.
- [x] Strategy A + the two additive guard-test edits confirmed to pass all six guards without editing the source doctrine.
- [x] Deploy mechanism confirmed generic (FS walk, no catalog entry needed).

## §F. @MX tag targets

Minimal — this is a doctrine-markdown packaging fix, not new exported Go code.

- **No @MX:ANCHOR / @MX:WARN targets** (no new high-fan-in or dangerous code path).
- **One optional @MX:NOTE candidate** at the `sanitizedPairPaths` registry addition, explaining the pair rationale (consistent with the existing prose comments there — the existing entries use plain comments, not `@MX`, so a plain comment is acceptable and `@MX:NOTE` is discretionary). Deferred to run-phase discretion.

## §G. Cross-references

- spec.md §A (context), §B (requirements), §C (exclusions)
- acceptance.md (AC-ZRP-001..006)
- `CLAUDE.local.md` §2 / §15 / §25
