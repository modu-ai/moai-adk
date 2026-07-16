---
id: SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001
title: "Package the zone-registry doctrine into the embedded template tree"
version: "0.1.1"
status: draft
created: 2026-07-16
updated: 2026-07-16
author: manager-spec
priority: P1
phase: "v3.0.0-rc target"
module: "internal/template/templates/.claude/rules/moai/core"
lifecycle: spec-anchored
tags: "template, packaging, zone-registry, neutralization, cli"
issue_number: 1090
---

# SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001 — Package the zone-registry doctrine into the embedded template tree

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-16 | manager-spec | Initial plan-phase draft. Tier S packaging fix for GitHub issue #1090. |
| 0.1.1 | 2026-07-16 | manager-spec | Plan-audit iter-2 fix (iter-1 FAIL 0.62). D1: enumerated the 5th/6th template guards (`rule_provenance_audit`, `rule_date_provenance`); added REQ/AC-ZRP-007 for the file-level governance-token exemption (119 CONST occurrences preserved). D2: date guard added to inventory. D3: dropped the unobservable spec-lint sub-assertion in AC-ZRP-004. |

## §A. Context / Why

### A.1 The defect (GitHub #1090)

`.claude/rules/moai/core/zone-registry.md` is a real, git-tracked doctrine file (40337 bytes, 115 `CONST-*` entries) that the moai CLI loads at runtime, but it is **never shipped in the embedded template tree**. Consequently `moai init` / `moai update` deploy a project in which three CLI commands break:

- `moai constitution list` / `guard` / `amend` / `validate` — resolve the registry via `resolveRegistryPath(cwd)` and fail to load (`internal/cli/constitution.go:24` — `const constitutionRegistryRelPath = ".claude/rules/moai/core/zone-registry.md"`).
- `moai doctor` — emits `zone-registry.md not found at %q — run "moai constitution list" to verify` and downgrades the check to Warn (`internal/cli/doctor.go` `os.Stat(registryPath)` branch).
- `moai spec lint` — `detectRegistryPath(cwd)` returns `""` when the file is absent, so the zone-registry-backed lint dimension is silently skipped (`internal/cli/spec_lint.go:158-162`).

The consumption is grounded (not assumed): `constitution.LoadRegistry(registryPath, projectDir)` parses the file and returns `reg.Warnings`; `doctor.go` calls `LoadRegistry` after the `os.Stat` gate; `spec_lint.go` passes the resolved path into the lint request.

### A.2 Evidence — the lone omission

Every neighbor core rule IS mirrored under `internal/template/templates/.claude/rules/moai/core/` (`moai-constitution.md`, `agent-common-protocol.md`, `askuser-protocol.md`, `verification-claim-integrity.md`, `hooks-system.md`, …). `zone-registry.md` is the single omission. The mirror path `internal/template/templates/.claude/rules/moai/core/zone-registry.md` does not exist (verified via `ls`).

Corroborating: the neutrality audit test `internal/template/template_neutrality_audit_test.go` already lists `.claude/rules/moai/core/zone-registry.md` in its C2-bare-narrative allowlist — a mirror was intended but never landed.

### A.3 Deployment mechanism (grounds the fix)

The deployer walks the embedded FS generically (`internal/template/deployer.go` — `fs.WalkDir(d.fsys, ".", …)`), and the FS is embedded via `//go:embed all:templates` (`internal/template/embed.go`; the `all:` prefix includes dot-directories such as `.claude/`). `catalog.yaml` enumerates **skills only**, not individual rule files. Therefore adding the file under `internal/template/templates/.claude/rules/moai/core/` and rebuilding the binary is sufficient to deploy it — no catalog entry, no manifest edit. This is exactly how the neighbor `moai-constitution.md` deploys.

### A.4 The neutralization sub-problem (the crux)

Files under `internal/template/templates/` ship to all 16-language users and MUST NOT carry moai-adk internal development traces (`CLAUDE.local.md` §25 Template Internal-Content Isolation, §15 language neutrality). A §25 pre-scan of the 40 KB source found exactly **four** forbidden-content tokens, all §25-forbidden classes; every other forbidden class is zero:

| # | Token | Location | §25 class | CI guard that catches it |
|---|-------|----------|-----------|--------------------------|
| 1 | `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` | line 1009 (comment) | internal SPEC ID | `internal_content_leak_test.go` C1 (whole-tree, default tier) |
| 2 | `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001` | line 1017 (inside a `clause:` value) | internal SPEC ID | `internal_content_leak_test.go` C1 (whole-tree, default tier) |
| 3 | `2026-05-04` | line 629 (comment) | internal date | `internal_content_leak_test.go` S1 (strict tier, opt-in) + §25 doctrine |
| 4 | `2026-05-09` | line 630 (comment) | internal date | `internal_content_leak_test.go` S1 (strict tier, opt-in) + §25 doctrine |

Zero of the following are present: REQ/AC tokens, audit citations (`Audit N Finding`), commit SHAs, `/Users/` macOS-bias paths, `CLAUDE.local.md` references, `feedback_`/`memory.md` references, `PR #N`, `(canonical)` privilege markers. (Verified: `grep -c` returned 0 for each neutrality-audit binary class `/Users/`, `CLAUDE.local.md`, `PR #[0-9]+`, `| ... (canonical)`.)

The C1 leak class `\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b` is **whole-tree** (no skill-body scope), so tokens 1-2 WILL fail the default-tier leak test if mirrored verbatim, and `zone-registry.md` is NOT in the pedagogical allowlist. Tokens 3-4 fail only under strict mode but are §25-forbidden regardless.

### A.4.1 The SIX template CI guards that walk the mirror

The template test package (`internal/template/`) contains **six** guards that scan the `.claude/rules/` mirror (the iter-1 inventory enumerated only four; the fifth and sixth are added here after direct reads). All run under `go test ./internal/template/...`:

| # | Guard test | File | Scope | What it catches on this mirror | Sentinel |
|---|-----------|------|-------|-------------------------------|----------|
| 1 | `TestTemplateNoInternalContentLeak` | `internal_content_leak_test.go` | whole-tree (C1) + skill-body-scoped (C1b/C6/C7/S3) | C1 catches the 2 SPEC-IDs (lines 1009/1017). **C1b (which matches `CONST-V3R`) is `skillBodyScoped: true` → applies ONLY under `.claude/skills/`, so it does NOT fire on this `.claude/rules/` file.** | — |
| 2 | `TestTemplateNeutralityAudit` | `template_neutrality_audit_test.go` | whole-tree, per-class allowlists | Binary classes (`/Users/`, `CLAUDE.local.md`, `PR #N`, `(canonical)`) = 0 in source. C2-bare-narrative-v3r already **file-level-allowlists** `zone-registry.md` (line 138). | — |
| 3 | `TestRuleTemplateMirrorDrift` | `rule_template_mirror_test.go` | opt-in allowlist | `zone-registry.md` NOT enrolled in `workflowOptMirroredPaths`/`lateBranchMirroredPaths` (verified: 0 references) → byte-parity neither required nor flagged for the divergent mirror. | — |
| 4 | `TestSanitizedPairParity` | `sanitized_pair_parity_test.go` | registry-driven | REQ-ZRP-006 registration target; token-normalized structural-drift check. | `SANITIZED_PAIR_PARITY_DRIFT` |
| 5 | `TestRuleProvenanceAudit` (**D1 — was missed**) | `rule_provenance_audit_test.go` | whole `.claude/rules/` | governance-token class matches **`CONST-V3R[0-9]-[0-9]+`** whole-tree → all **119** legitimate `CONST-*` occurrences fire. Runs UNCONDITIONALLY (no strict-mode gate). | `RULE_GOVERNANCE_TOKEN_LEAK` |
| 6 | `TestRuleDateProvenance` (**D2 — was missed**) | `rule_date_provenance_audit_test.go` | whole `.claude/rules/` | date class `\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b` (BROADER than the leak-test strict `202[6-9]`). Only the 2 dates at 629/630 match (verified: source has exactly 2 ISO dates, both non-metadata prose). | `RULE_DATE_PROVENANCE_LEAK` |

### A.4.2 The governance-token guard and the CONST-token file-level exemption (D1 crux)

Guard #5 (`TestRuleProvenanceAudit`, governance-token class) is the blocking defect. Its regex `\bSPEC-V3R[0-9]-[A-Z0-9-]+\b|\bCONST-V3R[0-9]-[0-9]+\b|\bMIG-[0-9]{3}\b` walks the entire template `.claude/rules/` subtree. On this mirror it matches:

- **119 `CONST-V3R[0-9]-NNN` occurrences** (verified: `grep -coE '\bCONST-V3R[0-9]-[0-9]+\b'` = 119; the 115 `- id: CONST-` entries plus 4 in-body cross-references). These are the **legitimate content** of the constitution zone registry — §C forbids removing them. So they CANNOT be neutralized away; they must be **exempted**.
- The 2 SPEC-IDs at 1009/1017 (also matched by the `SPEC-V3R` alternation) — but neutralization removes those, so they are gone from the mirror.

The exemption hook is `isPedagogicallyAllowed(relForAllowlist, trimmed)` (`internal_content_leak_test.go:527`), and it is **strictly per-(path, token)-keyed** — it iterates `pedagogicalAllowlist` matching `entry.File == relPath && entry.SpecID == matched`. **There is NO existing file-level exemption mechanism.** Enumerating ~115 distinct CONST IDs as per-token allowlist entries would be brittle and enormous. Therefore the remediation is a **FILE-LEVEL scope carve-out for the governance-token class only, scoped to `zone-registry.md` only** — directly parallel to how `template_neutrality_audit_test.go:137-144` already holds `zone-registry.md` in a file-level `allowListSet(...)` for its C2 class.

**Concrete run-phase edit** (specified, not performed in plan-phase): in `rule_provenance_audit_test.go`, add a file-level allowlist and gate it in the governance-token branch of the walk loop:

```go
// governanceTokenFileAllowlist exempts specific rule files from the
// governance-token class ONLY. zone-registry.md IS the constitution zone
// registry — its 115 CONST-V3R* entries (119 token occurrences) are its
// legitimate content, not a leak. File-level (not per-token) because
// enumerating 119 CONST tokens would be brittle; parallel to the C2
// allowListSet in template_neutrality_audit_test.go. Single-path: the guard
// still fires on governance tokens in EVERY OTHER rules file.
var governanceTokenFileAllowlist = map[string]bool{
	".claude/rules/moai/core/zone-registry.md": true,
}
```

gated as (replacing the existing `if class.name == "governance-token" && isPedagogicallyAllowed(...)` line):

```go
if class.name == "governance-token" {
	if governanceTokenFileAllowlist[relForAllowlist] {
		continue
	}
	if isPedagogicallyAllowed(relForAllowlist, trimmed) {
		continue
	}
}
```

**Bounded blast radius**: the carve-out is a single-path map entry, scoped to the governance-token class only. It does NOT blind the guard to accidental `CONST-*`/`SPEC-V3R*` leaks in ANY other rules file, and does NOT touch the REQ/AC or lessons-W provenance classes (both = 0 in the source). The `TestRuleProvenanceRecurrenceBackstop` sibling tests the raw regex directly (not the walk-loop allowlist), so it is unaffected and stays green. AC-ZRP-007 pins the single-path property (a synthetic CONST token in a different rules file still fires).

### A.4.3 The date guard (D2) — incidentally satisfied, no exemption needed

Guard #6 (`TestRuleDateProvenance`) uses the broader `\b20[0-9]{2}-...` pattern. The source's ONLY two ISO dates are at 629/630 (both plain prose, not metadata-prefixed — verified). Neutralization **removes** both (they are also §25-forbidden token 3-4), so the mirror has zero ISO dates → the guard passes WITHOUT any file-level exemption. It is enumerated in the guard inventory for completeness and covered by AC-ZRP-005.

### A.4.4 Guard-test edits that ARE in scope

Two guard-test files gain a minimal edit in run-phase, both additive and file-scoped:

- `sanitized_pair_parity_test.go` — 1-line `sanitizedPairPaths` addition (REQ-ZRP-006).
- `rule_provenance_audit_test.go` — the file-level `governanceTokenFileAllowlist` addition above (REQ-ZRP-007).

Neither changes CLI path resolution, neither removes any `CONST-*` registry entry, and neither disables a guard class globally. These are the ONLY Go/test-file edits; everything else is markdown mirror content.

### A.5 The chosen neutralization strategy — Strategy A (sanitized-pair; neutralize the mirror only)

Two candidate strategies exist (see §B.1 / plan.md §A for the trade-off). This SPEC adopts **Strategy A**: the template mirror is a §25-neutralized derivative (the 4 tokens generalized/removed), the local source `.claude/rules/moai/core/zone-registry.md` is left untouched, and the intentional source↔mirror divergence is registered as a sanitized pair. Rationale:

1. **Direct sibling precedent.** `runtime-recovery-doctrine.md` — the very doctrine the two forbidden SPEC-IDs reference — is already handled as a sanitized pair (`sanitized_pair_parity_test.go` registry entry): "its template mirror strips internal provenance (SPEC-IDs, REQ/AC tokens, CONST registry ID, Origin line) while keeping the PUBLIC citations". `zone-registry.md` contains a `CONST-V3R6-001` entry pointing at `runtime-recovery-doctrine.md`; treating it as a sanitized pair keeps the two consistent.
2. **Provenance preservation.** The two SPEC-ID cross-references are legitimate maintainer provenance in a live CLI-loaded constitution registry. Strategy A keeps them in the working copy; Strategy B would erase them from the maintainer's own doctrine.
3. **Scope discipline.** Strategy A adds only the new mirror (plus a 1-line test-registry entry) and never edits the 40 KB live source — the smallest blast radius.
4. **§25 anticipates divergence.** The sanitized-pair machinery (`sanitized_pair_parity_test.go`) exists precisely to hold a neutralized mirror in doctrine-sync with a provenance-bearing source.

The byte-parity guard (`rule_template_mirror_test.go`) uses an explicit opt-in allowlist (`workflowOptMirroredPaths` / `lateBranchMirroredPaths`); `zone-registry.md` is NOT enrolled, so a divergent mirror does not break it.

## §B. Requirements (GEARS notation)

### REQ-ZRP-001 — Neutralized mirror content (Ubiquitous)

The template mirror `internal/template/templates/.claude/rules/moai/core/zone-registry.md` **shall** carry a §25-neutral derivative of the source `.claude/rules/moai/core/zone-registry.md` in which the four forbidden tokens (§A.4: two internal SPEC IDs, two internal dates) are generalized or removed, and **shall not** contain any `internal_content_leak_test.go` default-tier forbidden token nor any `template_neutrality_audit_test.go` binary-class token.

### REQ-ZRP-002 — Mirror file existence (Event-driven)

**When** the template tree is assembled, the mirror **shall** exist at `internal/template/templates/.claude/rules/moai/core/zone-registry.md` alongside its neighbor core rules, with doctrine content structurally equivalent to the source (all 115 `CONST-*` entries preserved; only the four tokens neutralized).

### REQ-ZRP-003 — Embed on build (Event-driven)

**When** `make build` runs, the recompiled binary **shall** embed the mirror via `//go:embed all:templates`, such that the embedded FS resolves `.claude/rules/moai/core/zone-registry.md`.

### REQ-ZRP-004 — Deployed-project consumption (Event-driven)

**When** `moai init` deploys the templates into a target project, the project **shall** contain the file at `.claude/rules/moai/core/zone-registry.md`; and **when** `moai doctor`, `moai constitution list`, and `moai spec lint` subsequently run in that project, each **shall** resolve and load the registry without the "not found" break (doctor check no longer downgraded to Warn for a missing registry; `constitution list` returns entries; `spec lint` no longer skips the registry-backed dimension for absence).

### REQ-ZRP-005 — All six template CI guards pass (Ubiquitous)

The mirror **shall** pass all six template CI guards that walk the `.claude/rules/` subtree (§A.4.1) — `TestTemplateNoInternalContentLeak`, `TestTemplateNeutralityAudit`, `TestRuleTemplateMirrorDrift`, `TestSanitizedPairParity`, `TestRuleProvenanceAudit`, and `TestRuleDateProvenance` — such that the full `go test ./internal/template/...` package (default and strict leak tiers) **shall** remain green. In particular, `TestRuleDateProvenance` (`rule_date_provenance_audit_test.go`, sentinel `RULE_DATE_PROVENANCE_LEAK`) **shall** be green because the two ISO dates (§A.4 tokens 3-4) are removed by neutralization.

### REQ-ZRP-006 — Sanitized-pair drift registration (Where capability)

**Where** the source and mirror diverge only by §25 neutralization, the pair `.claude/rules/moai/core/zone-registry.md` **shall** be registered in the `sanitizedPairPaths` registry of `internal/template/sanitized_pair_parity_test.go`, so a future doctrine change to the source that fails to propagate to the mirror is caught by `TestSanitizedPairParity` (sentinel `SANITIZED_PAIR_PARITY_DRIFT`).

### REQ-ZRP-007 — Governance-token guard green via file-level exemption (Where capability)

**Where** the mirror carries the 119 legitimate `CONST-V3R[0-9]-NNN` occurrences (§A.4.2), `internal/template/rule_provenance_audit_test.go` **shall** gain a single-path, governance-token-class-only file-level allowlist entry for `.claude/rules/moai/core/zone-registry.md`, such that `TestRuleProvenanceAudit` (sentinel `RULE_GOVERNANCE_TOKEN_LEAK`) is green AFTER the mirror lands WITH all 119 `CONST-*` occurrences present. The exemption **shall not** be a class-wide disable and **shall not** blind the governance-token guard to `CONST-*`/`SPEC-V3R*`/`MIG-*` tokens in any OTHER rules file — the guard **shall** still fire on a synthetic governance token introduced into a different `.claude/rules/` file.

## §C. Exclusions

### In scope — the two additive guard-test edits (not scope creep)

Two guard-test files gain a minimal, additive, file-scoped edit (§A.4.4): the `sanitizedPairPaths` 1-line addition (REQ-ZRP-006) and the `governanceTokenFileAllowlist` single-path addition (REQ-ZRP-007). These ARE in scope. They do not alter CLI path resolution, remove any registry entry, or disable a guard class globally. The exclusions below bound everything else.

### Out of Scope — source doctrine edits

- This SPEC does NOT edit the local source `.claude/rules/moai/core/zone-registry.md`. Strategy A leaves the working copy (including its internal provenance) byte-for-byte unchanged. Strategy B (neutralize-in-place then byte-identical mirror) is explicitly rejected in §A.5.

### Out of Scope — CLI path-resolution changes

- No change to `resolveRegistryPath` / `detectRegistryPath` / the doctor `os.Stat` gate. The CLI already hard-codes the correct deployed-project relative path; the defect is a missing file, not a wrong path. The graceful-degradation behavior (doctor Warn, spec-lint skip when the file is genuinely absent) is preserved as-is.

### Out of Scope — registry content / doctrine changes

- No `CONST-*` entry is added, removed, or semantically altered. Neutralization touches only the four provenance tokens; the 115 codified zone rules (119 `CONST-V3R` token occurrences) and their `clause`/`zone`/`canary_gate` fields are preserved (the line-1017 `clause` value is reworded only to drop the SPEC-ID, not to change the rule it codifies). The 119 CONST occurrences are kept green by the file-level governance-token exemption (REQ-ZRP-007), NOT by removing or rewriting any CONST entry.

### Out of Scope — other unmirrored rules

- Any other `.claude/rules/moai/**` file that may lack a template mirror is out of scope; this SPEC packages `zone-registry.md` only. Systematic mirror-coverage auditing is a separate concern.

### Out of Scope — strict-tier leak enforcement flip

- This SPEC does not change the default-vs-strict tiering of `internal_content_leak_test.go` (the date classes staying strict-tier opt-in is a pre-existing policy). It only ensures the mirror is clean under both tiers.

## §D. Acceptance Criteria

Full Given-When-Then enumeration lives in `acceptance.md`. Summary map:

| AC | Requirement | Gist |
|----|-------------|------|
| AC-ZRP-001 | REQ-ZRP-001 | Mirror carries zero default-tier leak tokens and zero neutrality binary-class tokens |
| AC-ZRP-002 | REQ-ZRP-002 | Mirror file exists; all 115 `CONST-*` entries present |
| AC-ZRP-003 | REQ-ZRP-003 | `make build` succeeds; embedded FS resolves the file |
| AC-ZRP-004 | REQ-ZRP-004 | `moai init` into a temp dir places the file; doctor/constitution list/spec lint operate without the "not found" break |
| AC-ZRP-005 | REQ-ZRP-005 | `go test ./internal/template/...` green (all SIX guards pass, incl. `TestRuleProvenanceAudit` + `TestRuleDateProvenance`), including strict-mode leak run |
| AC-ZRP-006 | REQ-ZRP-006 | Pair registered; `TestSanitizedPairParity` green (token-normalized doctrine identical/within tolerance) |
| AC-ZRP-007 | REQ-ZRP-007 | Governance-token guard green WITH 119 `CONST-*` present via single-path file-level exemption; guard still fires on a synthetic CONST token in a different rules file |

## §E. References

- GitHub issue #1090
- `CLAUDE.local.md` §2 (Template-First), §15 (language neutrality), §25 (Template Internal-Content Isolation)
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 (C1-C8 content-class catalogue)
- Precedent: `runtime-recovery-doctrine.md` sanitized pair (`internal/template/sanitized_pair_parity_test.go`)
- Lesson: `feedback_local_template_sync_neutralize_first` (blind copy latest-wins breaks §25/§15 — neutralize before mirroring)
