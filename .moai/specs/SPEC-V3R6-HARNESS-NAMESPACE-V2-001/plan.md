---
id: SPEC-V3R6-HARNESS-NAMESPACE-V2-001
title: "Harness namespace doctrine-code drift catch-up — implementation plan"
version: "0.1.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/harness, internal/template, .claude/skills/moai-meta-harness"
lifecycle: spec-anchored
tags: "namespace, harness, doctrine-drift, atomic-transition, generator-emission"
---

# Plan — SPEC-V3R6-HARNESS-NAMESPACE-V2-001

## §A. Context

The doctrine SSOTs declared `harness-*` as the user-owned skill namespace on 2026-05-26. The Go enforcement layer still recognizes `my-harness-*`. This is the Phase 2 catch-up SPEC that aligns code to doctrine. See `spec.md` §A/§B for the full problem statement and the rejected-predecessor rationale.

**Verbatim measurement (manager-spec, 2026-06-18):**

```bash
$ grep -rn "my-harness" internal/ | wc -l
188

$ grep -rl "my-harness" internal/ | wc -l
43
```

Breakdown:
- Non-test `.go` refs: **55** across **15** files
- Test `.go` refs: **133** across **28** files

The doctrine's "39 Go files + 30+ tests" estimate (`.moai/docs/harness-namespace-doctrine.md` line 61) and the handoff's "59 refs" figure are both STALE. The authoritative count is **188 refs / 43 files** as measured above. `research.md` §C reproduces the verbatim grep output for every file.

## §B. Known Issues & Risks

### RISK-1 — Atomic-transition partial application (HARD)

If the Go enforcement switches to `harness-*` but the generator emission does not (or vice versa), a generated `harness-*` skill would lack protection, OR an enforced `harness-*` pattern would have no producer. Mitigation: NFR-HNS-004 + AC-HNS-001 mandate a single-PR atomic commit group. Milestones M1-M5 are sequenced so all surfaces are edit-ready in one PR; the PR is not merged until all surfaces pass.

### RISK-2 — Legacy `my-harness-*` user data left unprotected (HARD)

Existing user projects may have `my-harness-*` skill directories. If the enforcement switches cold-turkey to `harness-*`, a pre-migration user skill becomes moai-managed-classified (prefix is not `moai-*`, but no longer matches the user-owned predicate) — actually, on inspection, `cleanMoaiManagedPaths` deletes via `.claude/skills/moai*` glob (update.go:1706-1708), which does NOT match `my-harness-*` at all. So the deletion path is safe. BUT the **template overlay write loop** and **backup logic** consult `isUserOwnedNamespace`; if a `my-harness-*` path is not recognized, it would not be backed up before a clean-reinstall. Mitigation: REQ-HNS-005 + design.md §D dual-recognition deprecation window.

### RISK-3 — Substring false-match between `harness-*` and `moai-harness-*`

A naive `strings.Contains(name, "harness-")` or `*harness-*` glob would misclassify `moai-harness-foo` as user-owned. Mitigation: REQ-HNS-004 mandates exact `strings.HasPrefix(name, "harness-")`. The math: `HasPrefix("moai-harness-foo", "harness-")` = false; `HasPrefix("harness-foo", "moai-harness-")` = false. Both directions safe. `design.md` §C proves this.

### RISK-4 — Test fixture churn obscures logic changes

133 test refs across 28 files is large; a reviewer may miss a logic change hidden in fixture churn. Mitigation: M2 (logic) and M3 (fixtures) are separate milestones; logic changes land first with characterization tests, then fixtures follow.

### RISK-5 — `moai-meta-harness` emission boundary misread as new generator authoring

The emission prefix lives in template Markdown, not a Go constant. A run-phase agent may over-read the constraint "edit emission prefix" as "rewrite the generator." Mitigation: NFR-HNS-003 + AC-HNS-003 + design.md §E make the boundary explicit: edit the prefix string in `meta-harness.md` § 6.4.1 and `moai-meta-harness/SKILL.md` lines 164-177; do NOT touch the 7-Phase workflow body or layer architecture.

## §C. Pre-flight (Before M1)

- [ ] `moai spec lint .moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-V2-001/` passes (0 findings)
- [ ] plan-auditor verdict ≥ 0.80 (this is a doctrine-aligned catch-up; the rejected predecessor's self-sealing failure mode does not apply because this SPEC honors the doctrine's declared target)
- [ ] Implementation Kickoff Approval (§19.1) obtained from user
- [ ] `git fetch origin main` + divergence check (pre-spawn sync discipline)

## §D. Constraints Carried Into Run-Phase

| Constraint | Source | Run-phase implication |
|-----------|--------|----------------------|
| No policy reversal | NFR-HNS-001 | Do not reclassify `moai-harness-*`. Do not edit `isUserOwnedNamespace` for `moai-harness-*` paths. |
| No `moai-builder-*` | NFR-HNS-002 | Do not introduce the prefix. |
| Emission-prefix edit only | NFR-HNS-003 | Edit prefix strings in template Markdown; do not author generator phases or skill-skeleton templates. |
| Atomic transition | NFR-HNS-004 | All 5 surfaces in one PR. Do not merge partial. |
| Cross-platform normalization | NFR-HNS-005 | Preserve `strings.ReplaceAll(rel, "\\", "/")` before every `HasPrefix`. |
| No user-data loss | NFR-HNS-006 | Keep `my-harness-*` recognition during deprecation window. |

## §E. Milestones (Tier M, atomic single-PR sequencing)

### M1 — Go enforcement: skills-side predicate migration

Edit the skills-side `HasPrefix` literals in:
- `internal/cli/update.go:1199` (`isUserAreaPath` skills check) — add `harness-*`, keep `my-harness-*` legacy
- `internal/cli/update.go:1236` (`isUserOwnedNamespace` skills check, REQ-UNP-001) — same
- `internal/cli/doctor_harness.go:122, 271` — enumerate `harness-*`
- `internal/cli/doctor_skills.go:52` — classify `harness-*` as INFO
- `internal/harness/prefix_conflict.go:52, 60` — collect `harness-*`, strip `harness-`
- `internal/harness/frozen_guard.go:19-20` — allowedPrefixes `harness-*` / `harness/`
- `internal/cli/update_preserve_inventory.go:21` (comment) — doc update

Characterization test first (DDD): capture current `my-harness-*` behavior, then migrate. REQ-HNS-001/004/005/008/009.

### M2 — Generator emission prefix migration (template Markdown edit)

Edit the emission prefix in:
- `internal/template/templates/.claude/skills/moai/workflows/project/meta-harness.md` (lines 292-293, 298, 301-304, 316, 326-327, 333-336, 344, 376, 417 — ~15 occurrences)
- `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` (lines 164, 168, 174, 177 — emission contract + drift note)

Run `make build` to regenerate embedded template. REQ-HNS-003. NFR-HNS-003 boundary: edit prefix strings only.

### M3 — CI sentinel pattern migration

Edit:
- `internal/template/namespace_protection_audit_test.go:40` — `HasPrefix(parts[2], "harness-")`
- `internal/template/namespace_protection_audit_test.go:9-20, 50` — test name + error message rename to `TestNamespaceLeakHarnessSkills` / `NAMESPACE_LEAK_HARNESS_SKILL`
- `internal/template/skills_removal_test.go:94, 108` — `HasPrefix(entry.Name(), "harness-")` + `entry.Name() == "harness"`

REQ-HNS-007.

### M4 — Test fixture migration (28 test files)

Mechanical rename `my-harness-*` → `harness-*` across the 28 test files enumerated in `research.md` §C.2. No logic changes; fixtures only. REQ-HNS-001..009 indirect coverage.

### M5 — Atomic integration verification + backward-compat sunset guard

- Run full `go test ./...` — all green
- Run `golangci-lint run` — clean
- Run `go test -run TestNamespace ./internal/template/...` — sentinel passes with new pattern
- Verify backward-compat: a fixture `my-harness-*` dir is still recognized by `isUserOwnedNamespace` (REQ-HNS-005)
- Verify substring separation: a `moai-harness-foo` fixture is NOT recognized as user-owned (REQ-HNS-004)
- Single PR; do not merge until all surfaces pass.

## §F. Anti-Patterns (Run-Phase)

- **AP-1**: Splitting the migration across multiple PRs (Go first, generator later) — violates NFR-HNS-004 atomicity.
- **AP-2**: Using `strings.Contains(name, "harness-")` instead of `HasPrefix` — violates REQ-HNS-004, risks `moai-harness-*` misclassification.
- **AP-3**: Authoring new generator phases or skill-skeleton templates during M2 — violates NFR-HNS-003.
- **AP-4**: Removing `my-harness-*` recognition cold-turkey (no deprecation window) — violates REQ-HNS-005, risks user-data loss.
- **AP-5**: Reclassifying `moai-harness-*` as user-owned — violates NFR-HNS-001 (the rejected predecessor's failure mode).
- **AP-6**: Touching `.moai/harness/` or `.claude/agents/harness/` code (already migrated) — out of scope (EXCL-6, EXCL-7).

## §G. Cross-References

- `spec.md` — requirements, scope, exclusions
- `acceptance.md` — Given-When-Then ACs incl. HARD ACs (atomic, substring, generator boundary, no data loss, no reversal)
- `research.md` — verbatim ref classification (188 refs / 43 files), file-by-file evidence
- `design.md` — atomic-transition sequencing, substring-conflict math proof, generator emission boundary, deprecation window
- `.moai/docs/harness-namespace-doctrine.md` §24.5 — drift entry-condition this SPEC resolves
- Rejected predecessor: `.moai/specs/SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001/` (status: superseded)
