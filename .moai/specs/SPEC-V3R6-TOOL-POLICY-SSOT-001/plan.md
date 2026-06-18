---
id: SPEC-V3R6-TOOL-POLICY-SSOT-001
title: "Tool/Permission Policy SSOT — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "policy, tools, ssot, harness, codex"
tier: L
era: V3R6
---

# Implementation Plan — SPEC-V3R6-TOOL-POLICY-SSOT-001

> Companion to `spec.md`. Tier L → 6 milestones (M1-M6). Run-phase executes via manager-develop (cycle_type per quality.yaml `development_mode: tdd` → RED-GREEN-REFACTOR).

---

## §A. Context

The run-phase implements a declarative tool/permission policy SSOT (`.moai/config/sections/tool-policy.yaml`) plus a Go codegen mechanism that emits the consumable artifacts (settings.json `permissions` block, hook matchers) FROM the YAML. The mechanism prevents the §24.5 doctrine/enforcement drift class by construction: both surfaces are generated from one source.

**Codegen approach chosen (one sentence)**: a Go codegen function in `internal/config` invoked via a NEW `moai tool-policy build` subcommand, which reads `tool-policy.yaml` and writes/regenerates the `permissions` block of `.claude/settings.json` (local) and `internal/template/templates/.claude/settings.json.tmpl` (template) in-place via block-region replacement (NOT full-file rewrite, to preserve the rest of settings.json).

**Key design decision**: the YAML lives in BOTH the local `.moai/config/sections/tool-policy.yaml` (dev source of truth) AND the template `internal/template/templates/.moai/config/sections/tool-policy.yaml` (distribution to consumer projects). `moai update` syncs template → local. The codegen reads the local YAML and writes settings.json. This preserves the Template-First Rule (C4) while keeping a single declarative source per project.

---

## §B. Known Issues (pre-run analysis)

- **KI-1**: settings.json is partially Go-template-rendered (`{{jsonEscape .SmartPATH}}`) and partially static JSON. The codegen must replace ONLY the `permissions` block (region-based edit), not the whole file. Risk: clobbering PATH/hooks/env. Mitigation: M2 implements block-region parsing (locate `"permissions": {` ... matching `}`), M4 verifies round-trip with a characterization test that diffs only the permissions region.
- **KI-2 (corrected iter-2)**: the background-agent write restriction (agent-common-protocol.md L162) is documented prose BACKED BY Claude Code runtime auto-deny (background subagents auto-deny non-pre-approved permission prompts per CLAUDE.md §14). Enforcement ALREADY EXISTS at the runtime layer. Seeding the rule into the YAML does NOT create a new enforcement point — it makes the existing policy machine-readable/auditable as a data object (book2 appendix B.3 "independently evaluable"). No workflow breaks because no new enforcement is added; the YAML entry is additive (a query/audit surface), not a behavioral change. Mitigation: M4 compatibility test confirms zero behavior change (the rule was already enforced; the YAML entry does not alter enforcement).
- **KI-3**: the status-transition-ownership.sh hook already encodes the Status Transition Ownership Matrix in bash. The YAML seeding (§D.2) duplicates this as data. Risk: drift between the bash case-branches and the YAML. Mitigation: M3 cross-references both; the bash hook is NOT regenerated from YAML in this SPEC (it stays as-is). The YAML entry for status-transition policy carries `audit: "mirrored from status-transition-ownership.sh L48-69; not yet codegen-fed"` so the duplication is explicit. Full hook-from-YAML generation is follow-up scope (§X.3).

---

## §C. Pre-flight (before M1)

- [ ] Confirm `internal/config` package layout (defaults.go, loaders) — codegen placement target.
- [ ] Confirm no naming collision: `grep -rni "tool-policy\|toolpolicy\|execpolicy" internal/` returns empty (verified during plan-phase: empty).
- [ ] Confirm `moai constitution` CLI surface (`internal/cli/constitution.go` L26-436) for query-infrastructure reuse (REQ-TPS-006).
- [ ] Confirm settings.json permissions block region boundaries (local L272+, template L380+) for block-region codegen.

---

## §D. Constraints (carry-forward from spec.md §E)

- C1 backward compatibility mandatory (characterization test in M4).
- C2 reuse the zone-registry PATTERN + constitution CLI SHAPE; NOT the constitution Rule schema (schemas are disjoint — D9 decision).
- C3 codegen in Go (`internal/config`).
- C4 Template-First Rule (YAML in both local + template).
- C5 Tier L (6 milestones, PR via manager-git).
- C6 §24.5 cross-reference (design.md §D).
- C7 GEARS + V3R6 era.

---

## §E. Self-Verification (manager-develop §E deliverables, run-phase)

The run-phase manager-develop will populate `progress.md` §E.2/§E.3 with the E1-E7 self-verification matrix. Plan-phase leaves §E.1 audit-ready signal populated and §E.2-§E.5 as placeholder headings (see `progress.md`).

---

## §F. Milestones (6 — Tier L)

### §F.1 M1 — Schema + seed YAML

**Scope**: Define the 6-field entry schema as a Go struct (`internal/config/toolpolicy/types.go`) + YAML loader. Seed `tool-policy.yaml` (both local + template) with entries extracted from the 4 scattered sources (full inventory in research.md §C). RED-GREEN-REFACTOR: write the loader test first (fails: no schema), then implement, then refactor.

**Deliverables**: `internal/config/toolpolicy/{types.go,loader.go,loader_test.go}`, `.moai/config/sections/tool-policy.yaml`, `internal/template/templates/.moai/config/sections/tool-policy.yaml`.

**Exit criteria**: loader test passes; YAML loads into `[]PolicyEntry`; seeded entries cover §D.1-D.4.

### §F.2 M2 — Codegen mechanism

**Scope**: Implement the Go codegen (`internal/config/toolpolicy/codegen.go`) that reads the YAML and writes the `permissions` block of settings.json via block-region replacement. Add `moai tool-policy build` subcommand (`internal/cli/tool_policy.go`). RED-GREEN-REFACTOR: codegen test first (fails: no codegen), then implement.

**Deliverables**: `internal/config/toolpolicy/{codegen.go,codegen_test.go,settings_region.go}`, `internal/cli/tool_policy.go`, `internal/cli/tool_policy_test.go`.

**Exit criteria**: `moai tool-policy build` regenerates the permissions block; round-trip test (YAML → generated → decision-equivalence) passes for every seeded entry.

### §F.3 M3 — Integration + thin tool-policy query + cross-references

**Scope**: Cross-reference `harness-namespace-doctrine.md` §24.5 in the YAML header (as the drift-class ANALOGY this SPEC prevents on its generated surfaces — see design.md §D.2 narrowed claim). Implement the thin `moai tool-policy list` query subcommand (`internal/cli/tool_policy.go`) that loads `tool-policy.yaml` directly and supports filter flags (`--risk-tier`, `--decision`, `--tool`) modeled on `moai constitution list --zone` SHAPE. The query does NOT delegate to or wrap `moai constitution list` — the tool-policy entry schema (`{tool, args_pattern, risk_tier, decision, owner_agent, audit}`) is DISJOINT from the constitution Rule schema (`{id, zone, zone_class, file, anchor, clause, canary_gate}`); wrapping is infeasible (D9 decision). Add cross-reference comments in codegen output.

**Deliverables**: YAML header cross-refs, `internal/cli/tool_policy.go` (NEW `list` subcommand + existing `build`), `internal/config/toolpolicy/query.go` (thin loader+filter, NOT a constitution wrapper), integration tests.

**Exit criteria**: `moai tool-policy list --risk-tier=irreversible` returns the irreversible-tier entries; `grep -rn "constitution" internal/cli/tool_policy.go` returns NO calls (no schema-unification attempt); cross-refs present in generated output.

### §F.4 M4 — Migration + compatibility verification

**Scope**: Capture pre-migration decision matrix as characterization test (the existing settings.json permissions behavior). Run codegen. Assert post-migration decisions EQUAL pre-migration for every entry (C1 backward compatibility). Verify the §24.5 drift class is structurally prevented: a single YAML edit propagates to both doctrine-comment and enforcement-block.

**Deliverables**: `internal/config/toolpolicy/compat_test.go` (characterization), migration commit.

**Exit criteria**: compat test passes (0 behavior changes); drift-prevention test passes.

### §F.5 M5 — Lint + single-rule-change demo

**Scope**: Run `golangci-lint`, `go test ./...`, `go vet`. Demonstrate REQ-TPS-009: flip one entry (e.g., `Bash(git push --force*)` from deny → ask) via YAML edit + regenerate, with NO Go decision-path edit. End-to-end test asserts the decision change propagates.

**Deliverables**: lint-clean run, single-rule-change demo test.

**Exit criteria**: lint clean; demo test passes; no Go decision-path edit required for the rule change.

### §F.6 M6 — PR creation (manager-git, Tier L)

**Scope**: manager-git creates the PR per Tier-based PR routing (Tier L → PR mandatory). Branch creation, `gh pr create`, Late-Branch closure.

**Deliverables**: PR with all M1-M5 commits.

**Exit criteria**: PR opened; CI green; ready for review.

---

## §G. Anti-Patterns (run-phase prohibitions)

- **AP-1 — Full-file settings.json rewrite**: the codegen MUST NOT rewrite the entire settings.json (would clobber PATH, hooks, env). MUST use block-region replacement. (KI-1)
- **AP-2 — Schema unification with constitution Rule**: attempting to wrap `moai constitution list` or unify the tool-policy schema with the constitution Rule schema violates C2/REQ-TPS-006 (the schemas are disjoint — D9). The tool-policy query loads its YAML directly with its own schema.
- **AP-3 — Prose-rule deletion without generated replacement**: deleting agent-common-protocol.md L162 background-write rule without a generated enforcement artifact is a silent regression (§X.3).
- **AP-4 — IF/THEN in new requirements**: GEARS only; residual `IF/THEN` triggers `LegacyEARSKeyword` lint (C7).
- **AP-5 — YAML in local-only (Template-First violation)**: the YAML MUST exist in both local + template (C4); local-only violates Template-First Rule.

---

## §H. Cross-References

- spec.md (this SPEC) — requirements, seed scope, exclusions.
- acceptance.md — AC matrix, Given-When-Then scenarios.
- research.md — scattered-policy inventory (file:line), book2 execpolicy survey.
- design.md — schema, codegen approach, alternatives considered, drift-prevention mechanism.
- progress.md — §E.1-§E.5 skeleton (run-phase fills §E.2-§E.5).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix (M1 commits carry `feat(SPEC-...): M1 ...`).
- `.moai/docs/harness-namespace-doctrine.md` §24.5 — drift class analogy (narrowed to YAML↔settings.json scope per design.md §D.2).

---

## Out of Scope

### Out of Scope — Canonical exclusions live in spec.md

- This plan.md is a companion artifact; the canonical exclusions (§X.1-§X.8) live in `spec.md`. This section satisfies the lint `MissingExclusions` rule for this file.
