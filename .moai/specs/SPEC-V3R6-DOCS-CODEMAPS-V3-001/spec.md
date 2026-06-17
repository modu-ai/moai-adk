---
id: SPEC-V3R6-DOCS-CODEMAPS-V3-001
title: "v3.0 codemaps SSOT generation + docs-truth checklist for docs cohort"
version: "0.1.0"
status: implemented
created: 2026-06-17
updated: 2026-06-17
author: GOOS행님
priority: P1
phase: "v3.0.0 target"
module: ".moai/project/codemaps"
lifecycle: spec-anchored
tags: "docs, codemaps, architecture, ssot, v3"
era: V3R6
---

# SPEC-V3R6-DOCS-CODEMAPS-V3-001

## §A. Problem Statement

moai-adk-go is at v3.0.0-rc2. A 7-agent dynamic-workflow analysis (ultracode)
just identified 34 documentation gaps (P0×7 / P1×9 / P2×12 / P3×6) between the
README.* and docs-site content and the actual v3.0 codebase. Health verdict:
critical.

The root cause is documentation drift: README.* and docs-site describe an
earlier architecture (the 17-agent catalog, ad-hoc status values, ad-hoc
frontmatter conventions, an incomplete CLI surface, and the pre-glm-5.2
multi-LLM mapping). The docs need to be rewritten to match v3.0 reality —
but rewriting them in 4 separate later SPECs (README / DOCSITE / COVERAGE /
i18n) without a shared factual baseline would let each SPEC re-derive the
same facts independently and risk re-introducing contradictions.

This SPEC is **Phase 0** of the "Sprint 14 Docs-v3" cohort. It generates
the single source of truth that every later docs SPEC cross-checks factual
claims against:

- the existing `.moai/project/codemaps/` directory (produced by the
  `/moai codemaps` skill; already populated with 5 default files —
  `overview.md`, `modules.md`, `dependencies.md`, `entry-points.md`,
  `data-flow.md — refreshed for v3.0.0-rc2 so its coverage of the v3.0
  capability layers is current, AND
- a NEW `.moai/project/codemaps/docs-truth.md` checklist extracting the
  canonical v3.0 facts the later SPECs need (agent catalog, status enum,
  frontmatter schema, CLI surface, GLM→Claude tier mapping).

With a codemaps ground-truth document in place, the later SPECs simply
reference it instead of re-deriving facts from primary sources.

## §B. Scope

### §B.1 In Scope

1. Run `/moai codemaps` (the codemaps skill) against the v3.0 codebase to
   REFRESH the existing `.moai/project/codemaps/` files
   (`overview.md` + `modules.md` at minimum; `dependencies.md`,
   `entry-points.md`, `data-flow.md` as the skill emits them). The
   deterministic baseline is `go list -deps -json` + `go doc`; the
   codemaps skill MAY additionally invoke its per-package Explore fan-out
   for architectural-insight augmentation layered on top of the
   deterministic extraction.
2. Author `.moai/project/codemaps/docs-truth.md` (NEW file) — a compact,
   machine-greppable checklist of canonical v3.0 facts, each
   cross-referenced to its ground-truth source file:
   - 8 retained agents (exact names + class) → `ls .claude/agents/moai/`
     + CLAUDE.md §4.
   - lowercase 8-value status enum → `internal/spec/status.go`.
   - 12 required SPEC frontmatter fields → `internal/spec/lint.go`
     `FrontmatterSchemaRule` + `.claude/rules/moai/development/spec-frontmatter-schema.md`.
   - complete CLI subcommand surface (`moai` terminal verbs + `/moai`
     Claude Code skill set) → `internal/cli/` + `.claude/commands/moai/`.
   - GLM→Claude-tier model mapping (reflecting the recent glm-5.2[1m]
     activation) → `internal/config/defaults.go`.
3. Verify the refreshed codemaps cover the capability layers that later
   docs SPECs will rewrite against: CLI surface / Lifecycle / Harness /
   Hooks / Templates / Quality / Multi-LLM / Web / Governance /
   Foundation. Coverage is judged collectively across `overview.md` +
   `modules.md` (the skill does NOT emit a single `architecture.md`
   index file).
4. Neutrality-sweep the refreshed `.moai/project/codemaps/` content so
   it carries zero internal SPEC-ID / REQ-token / "Audit N Finding AX"
   leakage (codemaps must be reusable as a neutral architecture
   artifact, not a SPEC-scoped working document).

### §B.1.1 Path and filename decisions (iter-2 audit resolutions)

- **Codemaps output path = `.moai/project/codemaps/`** (NOT repo-root
  `codemaps/`). The `/moai codemaps` skill's own description and Phase 3
  output specification (`.claude/skills/moai/workflows/codemaps.md` lines
  4, 31, 104-110) canonicalize this path. The directory ALREADY EXISTS
  with 5 default files; this SPEC refreshes them rather than creating a
  new tree. `module:` frontmatter set to `.moai/project/codemaps`.
- **Architecture-index filename decision = option (a)** (minimalism): NO
  new `architecture.md` is invented. The skill emits `overview.md` +
  `modules.md` (+ 3 siblings); capability-layer coverage is verified
  COLLECTIVELY across these real outputs (REQ-CM-001 re-targeted). Option
  (b) — synthesizing a new `codemaps-architecture.md` — is rejected to
  avoid inventing a file the skill does not manage.
- **CLI-surface assertion shape**: REQ-CM-007 / AC-CM-005 verify that
  `docs-truth.md` lists the HUMAN-FACING verb names (e.g. `init`,
  `update`, `glm`, `cc`, `cg`, `web`, `session`, `spec`, `harness`, …)
  and all 17 `/moai` skill names, NOT internal Go identifiers like
  `worktree.WorktreeCmd`. The extraction spans the full
  `internal/cli/*.go` tree (`rootCmd.AddCommand` calls live across 26
  files, not only `root.go`) and is cross-checked against the rendered
  `moai --help` verb list for robustness.
- **"High-fan-in" package set (REQ-CM-002)** is renamed to "named
  capability-anchor package set" and defined by an explicit selection
  rule: the 9 packages (`spec`, `cli`, `config`, `statusline`, `hook`,
  `template`, `harness`, `session`, `web`) are the packages whose public
  surfaces are cited as ground-truth sources by the docs-truth checklist
  (REQ-CM-004..008) OR that carry a v3.0 capability layer named in
  REQ-CM-001. The set is therefore derivable from the docs-truth
  citation list, not copied verbatim.

### §B.2 Out of Scope

- Rewriting README.* (later cohort SPEC: README).
- Rewriting docs-site pages (later cohort SPEC: DOCSITE).
- Closing test-coverage gaps surfaced by the cohort analysis (later cohort SPEC: COVERAGE).
- i18n parity fixes across docs-site locales (later cohort SPEC: i18n).
- Any source-code change. `internal/`, `pkg/`, `cmd/` MUST remain untouched — `.moai/project/codemaps/` is additive documentation output.
- Formal publication / hosting of `.moai/project/codemaps/` on docs-site (that is a decision for the DOCSITE SPEC; this SPEC only produces the local artifact).
- Inventing a NEW `architecture.md` (or `codemaps-architecture.md`) index file. The codemaps skill does not manage such a file; capability-layer coverage is verified across the skill's real outputs (`overview.md` + `modules.md`).

## §C. History

- 2026-06-17: plan-phase authored. The 7-agent dynamic-workflow analysis
  identified 34 documentation gaps and recommended this codemaps-first
  SPEC as the prerequisite for the 4-SPEC docs rewrite cohort. This SPEC
  is the first SPEC of "Sprint 14 Docs-v3".

## §D. Requirements (GEARS notation)

### REQ-CM-001 — Codemaps refreshed and collectively cover capability layers (Ubiquitous)

The [codemaps skill] **shall** refresh the existing
`.moai/project/codemaps/` files (`overview.md` and `modules.md` at
minimum) so that, COLLECTIVELY across those files, the 10 v3.0
capability layers — CLI surface / Lifecycle / Harness / Hooks /
Templates / Quality / Multi-LLM / Web / Governance / Foundation — are
each represented. The deterministic baseline is `go list -deps -json`
and `go doc`. The skill does NOT emit a single `architecture.md` index;
coverage is judged across its real outputs.

### REQ-CM-002 — Per-package maps for the named capability-anchor package set (Where)

**Where** a package belongs to the named capability-anchor package set
— defined as the packages whose public surfaces are cited as ground-
truth sources by the docs-truth checklist (REQ-CM-004..008) OR that
carry a v3.0 capability layer named in REQ-CM-001 (the set resolves to
`spec`, `cli`, `config`, `statusline`, `hook`, `template`, `harness`,
`session`, `web`) — the [codemaps skill] **shall** cover that package's
public surface, direct dependencies, and reverse dependencies within
`.moai/project/codemaps/modules.md` (or a per-package file the skill
emits).

### REQ-CM-003 — docs-truth.md checklist existence (Ubiquitous)

The [docs cohort] **shall** have a single canonical facts checklist at
`.moai/project/codemaps/docs-truth.md` that every later docs SPEC
cross-checks its factual claims against before merging.

### REQ-CM-004 — docs-truth.md agent catalog (Ubiquitous)

The [docs-truth checklist] **shall** list exactly the 8 retained agents
with correct class labels — Manager×4 (`manager-spec`, `manager-develop`,
`manager-docs`, `manager-git`), Evaluator×2 (`plan-auditor`,
`sync-auditor`), Builder×1 (`builder-harness`), Anthropic built-in×1
(`Explore`) — cross-verified against the contents of
`.claude/agents/moai/` and CLAUDE.md §4.

### REQ-CM-005 — docs-truth.md status enum (Ubiquitous)

The [docs-truth checklist] **shall** record the SPEC status enum as the
exact lowercase 8-value set — `draft`, `planned`, `in-progress`,
`implemented`, `completed`, `superseded`, `archived`, `rejected` —
cross-verified against `internal/spec/status.go` `ValidStatuses`.

### REQ-CM-006 — docs-truth.md frontmatter schema (Ubiquitous)

The [docs-truth checklist] **shall** record the 12 required SPEC
frontmatter fields — `id`, `title`, `version`, `status`, `created`,
`updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags` —
cross-verified against `internal/spec/lint.go` `FrontmatterSchemaRule`
`required` slice (lines ~586-602) and
`.claude/rules/moai/development/spec-frontmatter-schema.md`.

### REQ-CM-007 — docs-truth.md CLI surface (Ubiquitous)

The [docs-truth checklist] **shall** record the complete CLI subcommand
surface in human-facing terms — the `moai` terminal verb names (as
registered via `rootCmd.AddCommand` across the `internal/cli/*.go` tree,
NOT only `root.go`) AND the full `/moai` Claude Code skill set
(`.claude/commands/moai/*.md`) — cross-verified against the rendered
`moai --help` verb list and `ls .claude/commands/moai/`. The checklist
MUST list the human-facing verb names (e.g. `init`, `update`, `glm`,
`cc`, `cg`, `web`, `session`, `spec`, `harness`, `worktree`, …), NOT
internal Go identifiers.

### REQ-CM-008 — docs-truth.md GLM→Claude tier mapping (Ubiquitous)

The [docs-truth checklist] **shall** record the GLM→Claude-tier model
mapping reflecting the glm-5.2[1m] activation — including the full
tier-models table from `internal/config/defaults.go` lines 40-57:
`DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`,
`DefaultGLMHigh = "glm-5.2[1m]"`, `DefaultGLMMedium = "glm-4.7"`,
`DefaultGLMLow = "glm-4.5-air"`, `DefaultGLMSonnet = "glm-4.7"`,
`DefaultGLMHaiku = "glm-4.5-air"`, `DefaultGLMOpus = "glm-5.2[1m]"` —
cross-verified against `internal/config/defaults.go`.

### REQ-CM-009 — Codemaps neutrality (When)

**When** the codemaps are refreshed, the [codemaps skill] **shall** emit
content that contains zero internal SPEC-ID tokens, zero REQ tokens, and
zero "Audit N Finding AX" citations, so that `.moai/project/codemaps/`
remains a neutral reusable architecture artifact rather than a SPEC-scoped
working document.

### REQ-CM-010 — Code-change zero (Unwanted)

The [docs cohort] **shall not** modify any Go source file under
`internal/`, `pkg/`, or `cmd/` in this SPEC — `.moai/project/codemaps/`
is additive documentation output only.

### REQ-CM-011 — No regression in build/vet/lint (When)

**When** the codemaps are refreshed, the [docs cohort] **shall** introduce
zero NEW failures in `go build ./...`, `go vet ./...`, or
`golangci-lint run` relative to the pre-SPEC baseline, because
`.moai/project/codemaps/` is additive Markdown output that the Go
toolchain does not compile.

### REQ-CM-012 — Ground-truth source citation (Ubiquitous)

The [docs-truth checklist] **shall** cite, for every canonical fact it
records, the ground-truth source file (and, where applicable, the symbol
or line) from which the fact was extracted, so that later SPECs can
re-verify the fact without consulting the codemaps itself.

## §E. Constraints

- **C1 (Tier)**: Tier M — three sequential milestones, no code change,
  additive docs output only.
- **C2 (Lifecycle)**: `spec-anchored` — the SPEC is maintained alongside
  the implementation and reviewed each release.
- **C3 (Neutrality)**: codemaps content MUST be reusable-neutral. The
  `internal/template/internal_content_leak_test.go` + the
  `.github/workflows/template-neutrality-check.yaml` guards do not cover
  `.moai/project/codemaps/` (it is not a template artifact), so
  neutrality is enforced via AC-CM-007 instead.
- **C4 (Determinism)**: the codemaps baseline is the deterministic
  `go list -deps -json` + `go doc` extraction. Any architectural-insight
  augmentation layered on top (Explore fan-out) MUST be clearly labeled
  as insight, not extraction, so reviewers can distinguish fact from
  judgment.
- **C5 (Era)**: `era: V3R6` explicit frontmatter override is set so
  `ClassifyEra` does not misclassify the initially-empty `progress.md`
  as V3R2-R4 (H-2 fires when `§E.2..§E.5` markers are absent).

## §F. Exclusions (What NOT to Build)

1. NO source-code changes under `internal/`, `pkg/`, `cmd/`.
2. NO README.* rewrites (deferred to later cohort SPEC).
3. NO docs-site page rewrites or Vercel redeploys (deferred to DOCSITE
   SPEC).
4. NO test-coverage gap closures surfaced by the cohort analysis (deferred
   to COVERAGE SPEC).
5. NO new SPEC lint rules, frontmatter validators, or status enum changes
   — the 8-value enum and 12-field schema are GROUND TRUTH for this SPEC,
   not modification targets.
6. NO publication of `.moai/project/codemaps/` to docs-site hosting
   (publication is a decision for the DOCSITE SPEC; this SPEC produces
   the local artifact only).
7. NO archived-agent catalog entries — the 12 archived agents
   (`manager-strategy`, `manager-quality`, `manager-brain`,
   `manager-project`, `claude-code-guide`, `researcher`, and the six
   `expert-*` agents) are out of scope for docs-truth; docs-truth lists
   ONLY the 8 retained agents.
8. NO new `architecture.md` (or `codemaps-architecture.md`) index file.
   The codemaps skill does not manage such a file; capability-layer
   coverage is verified across the skill's real outputs.

## §G. Success Criteria

- `.moai/project/codemaps/overview.md` AND `.moai/project/codemaps/modules.md`
  both exist (refreshed for v3.0.0-rc2) and, COLLECTIVELY across the two
  files, represent all 10 capability layers named in REQ-CM-001.
- `.moai/project/codemaps/docs-truth.md` exists and records every
  canonical fact in REQ-CM-004..REQ-CM-008 with a ground-truth citation
  per REQ-CM-012.
- `go build ./...`, `go vet ./...`, and `golangci-lint run` show zero NEW
  failures relative to pre-SPEC baseline.
- `.moai/project/codemaps/` carries zero internal SPEC-ID / REQ /
  Audit-citation leakage.
- No Go source file under `internal/`, `pkg/`, `cmd/` is modified by this
  SPEC.

## §H. Cross-References

- **`.moai/project/tech.md`** — project technology stack (consult for
  capability-layer naming conventions).
- **`.moai/project/structure.md`** — project directory layout (consult for
  per-package map scoping).
- **`CLAUDE.md` §4** — 8-agent retained catalog (SSOT for agent names
  and classes).
- **`.claude/rules/moai/development/spec-frontmatter-schema.md`** — 12-
  field canonical schema SSOT.
- **`.claude/rules/moai/workflow/lifecycle-sync-gate.md`** — era
  classification policy (rationale for the `era: V3R6` override).
- **`internal/spec/status.go`** — status enum SSOT.
- **`internal/spec/lint.go`** `FrontmatterSchemaRule` — frontmatter
  validation SSOT.
- **`internal/config/defaults.go`** — GLM/Claude tier mapping SSOT.
- **Sprint 14 Docs-v3 cohort** — this SPEC is Phase 0; subsequent phases
  are README / DOCSITE / COVERAGE / i18n (to be proposed as separate
  SPECs).
