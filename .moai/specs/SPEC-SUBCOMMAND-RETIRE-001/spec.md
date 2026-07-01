---
id: SPEC-SUBCOMMAND-RETIRE-001
title: "Retire 5 underused /moai subcommands and 7 dependent skills"
version: "0.1.0"
status: in-progress
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template/templates/.claude"
lifecycle: spec-anchored
tags: "cleanup, subcommand, skill-retirement, template, catalog"
---

# SPEC-SUBCOMMAND-RETIRE-001 — Retire 5 underused /moai subcommands and 7 dependent skills

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-01 | 0.1.0 | Initial plan-phase draft. User-confirmed cleanup (AskUserQuestion this session): remove `design`/`brain`/`e2e`/`coverage`/`security` from TEMPLATE SOURCE (most-aggressive option, permanent for all distributed users). Tier L. | manager-spec |

## §A. Context and Intent

`moai-adk-go` is a Go CLI tool maintained by a 1-person OSS author. Of its 18+ `/moai`
subcommands, five are peripheral to its domain (Go CLI development centered on the
plan→run→sync lifecycle) and carry measurable dead-weight signal (see research.md §A):
`design` (brand/UI), `brain` (7-phase ideation), `e2e` (browser E2E), `coverage`
(redundant with `go test -cover`), and `security` (re-routable through retained
reference skills). Commit `1ece11578` already removed `brain`·`design` from the
docs-site; this SPEC completes the half-done retirement across the workflow files,
SKILL.md router, 7 dependent skills, catalog, CI-guard tests, and remaining docs-site
pages.

**Layer**: removal targets the TEMPLATE SOURCE (`internal/template/templates/`) per
the Template-First Rule (CLAUDE.local.md §2), so the change is permanent for all
distributed users — not a local-only hide that `moai update` would revert.

## §B. Scope Summary

**In scope** — remove (both `.claude/` local tree and `internal/template/templates/.claude/`
template tree): 5 command files, 5 workflow files, 7 dependent skill directories,
their `catalog.yaml` entries; update the SKILL.md intent router; reconcile CI-guard
test count constants and dedicated tests; clean in-scope cross-reference surfaces;
remove residual docs-site pages for `coverage`/`e2e`; preserve the OWASP security-audit
capability via retained reference skills; run `make build` to regenerate embedded
templates + catalog hashes.

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### Core removal

- **REQ-SCR-001** (Ubiquitous): The template source shall not contain command files,
  workflow files, or `catalog.yaml` entries for the retired subcommands `design`,
  `brain`, `e2e`, `coverage`, `security`.
- **REQ-SCR-002** (Ubiquitous): The template source shall not contain the seven retired
  skill directories `moai-domain-ideation`, `moai-domain-research`,
  `moai-domain-design-handoff`, `moai-domain-brand-design`, `moai-domain-copywriting`,
  `moai-workflow-design`, `moai-workflow-gan-loop`.
- **REQ-SCR-006** (Ubiquitous): The local tree (`.claude/`) and the template tree
  (`internal/template/templates/.claude/`) shall both be free of the retired artifacts
  after removal — no local-only residue (which `moai update` would re-delete) and no
  template-only residue (which would re-deploy to users). Tree cleanliness is verified by
  the **absence of the seven removed skill basenames from both trees**, NOT by full-tree
  identity (see REQ-SCR-012 for the §24 count asymmetry).
- **REQ-SCR-012** (Ubiquitous / unwanted): The removal shall not delete any user-owned
  `harness-*` skill (`.claude/skills/harness-*/` — verified present: `harness-moaiadk-best-practices`,
  `harness-moaiadk-patterns`). Per CLAUDE.local.md §24 these are user-owned, exist ONLY in
  the local tree (never the template tree), and `moai update` MUST preserve them. The
  post-removal skill counts are therefore **asymmetric by design — 28 template skills vs 30
  local skills (28 + 2 user-owned `harness-*`)**, NOT "28 each". Any dual-tree verification
  shall be scoped to the seven removed basenames' absence from both trees; a raw full-tree
  `diff` expecting identity is prohibited (it would either report a false FAIL or motivate a
  §24-violating deletion of the two `harness-*` skills).

### Build + catalog integrity

- **REQ-SCR-003** (Event-driven): When `make build` is run after removal, the build
  shall regenerate `internal/template/embedded.go` and refresh `catalog.yaml` hashes
  such that the embedded catalog reflects the reduced skill set (28 skills / 35 total
  entries; agent count unchanged at 7). The exact post-removal counts shall be
  re-derived by recount during run-phase, not assumed.
- **REQ-SCR-009** (Capability gate / Where): Where the template-neutrality CI guard
  (`.github/workflows/template-neutrality-check.yaml` + `internal_content_leak_test.go`)
  runs on the modified template files, it shall pass — removal introduces no forbidden
  content classes.

### Test-suite integrity (highest risk)

- **REQ-SCR-004** (Event-driven): When `go test ./internal/template/...` is run after
  removal, the suite shall pass with zero failures. This binds the count-constant
  reconciliation (`catalog_tier_audit_test.go` 35→28, `catalog_loader_test.go` 42→35,
  `embed_catalog_test.go` 42→35), the deletion of `TestBrainCommandThinPattern`, and
  ALL THREE `agentless_audit_test.go` package-level path-list fixes: remove `coverage.md`
  from `utilitySkillPaths` (else `TestAgentlessUtilityNoLLMControlFlow` +
  `TestUtilitySkillsContainModeFlagIgnoredSentinel` go RED), remove `design.md` from
  `implementationSkillPaths` (else `TestImplementationSkillsContainPipelineRejectionSentinel`
  goes RED), and remove `design.md` from `runDesignSkillPaths` in
  `TestRunDesignSkillsContainModeUnknownSentinel`.

### Router + cross-reference integrity

- **REQ-SCR-005** (State-driven / While): While the SKILL.md intent router is loaded,
  the router shall not list or route to any of the five retired subcommands or their
  aliases (`brain`/`ideate`/`idea`, `design`/`brief`/`brand`, `coverage`/`cov`,
  `e2e`/`e2e-test`, `security`/`audit`/`sec`).
- **REQ-SCR-008** (Ubiquitous): The in-scope cross-reference surfaces — `moai.md`,
  `CLAUDE.md`, `spec-workflow.md`, `glm-web-tooling.md`, and the retained
  `moai-domain-humanize` SKILL.md — shall contain zero dangling references to the
  retired subcommands or skills.
- **REQ-SCR-010** (Event-driven): When the removal orphans a retained skill's
  cross-reference (specifically `moai-domain-humanize` → `moai-domain-copywriting`),
  the retained skill's reference shall be rewritten to remove the dependency on the
  retired skill.

### Capability preservation

- **REQ-SCR-007** (Capability gate / Where): Where a user issues a natural-language
  security-audit request after `/moai security` is removed, the orchestrator shall
  route the request to `Agent(general-purpose)` with security scope, loading the
  retained `moai-ref-owasp-checklist` / `moai-ref-secops` / `moai-ref-supply-chain` /
  `moai-ref-llm-security` reference skills — preserving the OWASP audit capability
  `/moai security` provided. No capability gap shall remain for the other four
  retirements either (see research.md §E replacement-path map).

### Documentation

- **REQ-SCR-011** (Event-driven): When the docs-site is rebuilt after removal, it shall
  not present menu entries or content pages for the retired `coverage` and `e2e`
  quality commands across all four locales (en / ko / ja / zh), per the §17 4-locale
  parity doctrine. (`brain`/`design` were already removed by commit `1ece11578`;
  `security` has no docs-site page.)

## §D. Acceptance Criteria Pointer

Full Given-When-Then scenarios, the CI-guard GREEN gate, the 0-dangling-reference gate,
and the Definition of Done live in `acceptance.md`. The binding verification command for
REQ-SCR-004 is `go test ./internal/template/...`.

## §E. Out of Scope

### Out of Scope — Design rule subsystem and brand assets
- The FROZEN design methodology rule `.claude/rules/moai/design/constitution.md` (header
  confirms "FROZEN/EVOLVABLE zone") and its `zone-registry.md` FROZEN mirror entries
  (051-149) are NOT edited — they are zone-protected and document the design methodology,
  which persists as reference independent of the `/moai design` slash command.
- `.moai/config/sections/design.yaml`, `.moai/project/brand/`, `.moai/design/`
  (incl. `internal/template/templates/.moai/design/README.md`), and the `internal/design/`
  Go package (DTCG design tokens — may serve the `moai web` console) are NOT removed.
- Only the `/moai design` command entry point and the five design-pack skills are removed.
  Wholesale design-subsystem retirement is deferred to a follow-up SPEC.

### Out of Scope — Subcommand re-implementation
- No replacement subcommand is built. Capability survives via existing paths only
  (research.md §E): `security` → `/moai review --security` + `Agent(general-purpose)`;
  `coverage` → `go test -cover` / `/moai gate`; `brain` → `/moai plan`; `design` →
  `moai web` + brand config; `e2e` → no in-tool replacement (out of Go-CLI domain).

### Out of Scope — Historical SPEC artifacts and dependency manifests
- Existing SPEC bodies that reference the retired features (e.g. `SPEC-V3R3-BRAIN-001`)
  and archived/backup paths are immutable historical artifacts and are NOT edited.
- `go.mod` / dependency manifests are unaffected (removal is markdown + catalog only).

### Out of Scope — CHANGELOG and README
- `CHANGELOG.md` and `README.md` entries for this retirement are owned by `manager-docs`
  in the sync phase, not authored here.
