---
id: SPEC-TEMPLATE-DATE-NEUTRALITY-001
title: "Template date-leak triage, remediation, and S1 guard refinement"
version: "0.2.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P2
phase: "v3.x maintenance"
module: "internal/template"
lifecycle: spec-anchored
tags: "template, neutrality, guard, date, isolation, ci"
tier: L
---

# SPEC-TEMPLATE-DATE-NEUTRALITY-001 — Template date-leak triage, remediation, and S1 guard refinement

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | Initial plan-phase authoring. Scope from a measured strict-tier guard run (135 findings, 116 files). | manager-spec |
| 0.2.0 | 2026-07-25 | Iteration 2 after plan-audit FAIL (0.55). Four user decisions absorbed (hybrid carve-out, DC-1 preserve, mirror-stamp preserve, CI isolated target). Classification moved to occurrence-class granularity to resolve dual-shape conflicts; DC-2 split into DC-2a/DC-2b; counts restated from a committed classifier that reproduces the guard's finding set exactly; three misclassifications corrected. | manager-spec |

---

## §1 Context

`internal/template/internal_content_leak_test.go` implements a two-tier neutrality guard over the user-distributed template tree (`internal/template/templates/**`). The **narrow** tier runs in CI and is green. A **strict** tier, activated by `MOAI_TEMPLATE_LEAK_STRICT=1`, adds the `S1-internal-date` class (any `202[6-9]-MM-DD` literal) plus `S2-short-sha-sentence-final`, and is enforced nowhere.

### Measured baseline (this tree, `c7309aeb6`)

| Measurement | Observed |
|---|---|
| Strict tier | `exit=1`, `template internal-content leak detected (135 occurrences, mode=strict)` |
| Narrow tier | `exit=0` |
| Package-wide `go test ./internal/template/...` | `exit=0` — the package is **green** |
| Strict-mode CI wiring | no matches |
| Distinct files carrying a finding | 116 |
| Report cap | `limit := 50`; `... 85 more (capped)` |

### The core tension

The findings are not a homogeneous sweep target. A committed classifier (§4 REQ-TDN-020) partitions them into six categories with opposite correct treatments: internal authoring metadata that is meaningless in a distributed template, schema-coupled frontmatter fields, mirror-capture stamps that are a reader's only staleness signal for a third-party document mirror, functionally load-bearing deadline dates, attribution records, and an adjudicated residue. Deleting the wrong kind is a regression.

---

## §2 Definitions

These terms carry exactly these meanings throughout the SPEC. This section is definitional, not normative.

- **Finding** — a `(file-path, date-literal)` pair. This is the guard's own unit: `collectLeakViolations` deduplicates matches per file per distinct literal. Measured: **135**.
- **Occurrence-class row** — a `(file-path, date-literal, line-shape, line-number)` tuple. This is the *edit* unit and is finer than a finding: one finding may appear under two different line shapes in the same file. Measured: **180**.
- **Line shape** — the syntactic form of the line carrying the literal. Four shapes: `LS-FM` (a real YAML frontmatter `updated:` key), `LS-FM-FENCED` (the same key inside a fenced code block — a documentation example), `LS-PROSE-STAMP` (a standalone `Last Updated:` / `Updated:` / `# Updated:` header or footer line), `LS-OTHER` (inline prose, table cells, changelog entries).
- **Dual-shape finding** — a finding whose rows carry two line shapes that receive conflicting dispositions. Measured: **13** findings span two categories — **12** are `DC-1`/`DC-2a` conflicts and **1** is a `DC-1`/`DC-5` span.

---

## §3 Settled Decisions

Recorded so downstream readers do not re-litigate them:

- **Carve-out mechanism** = hybrid (structural gate + content-anchored allowlist). Binding in REQ-TDN-010.
- **DC-1 frontmatter `updated:`** = preserve as authored. No schema change, no render-time neutralization. Binding in REQ-TDN-009.
- **Mirror-capture stamps** = preserve. Binding in REQ-TDN-011.
- **CI enforcement** = an isolated-target step inside the existing neutrality workflow, added only after the finding count reaches zero. Binding in REQ-TDN-013 + REQ-TDN-015.

---

## §4 Requirements (GEARS)

### Classification

**REQ-TDN-001** — The classifier shall assign every occurrence-class row to exactly one of six categories, evaluating the rules in the order listed:

| # | Category | Decision rule (first match wins) | Rows | Findings | Disposition |
|---|---|---|---:|---:|---|
| DC-3 | Functional deadline | the date literal is `2026-11-22` | 13 | 9 | PRESERVE |
| DC-4 | Attribution / license | the file is `.claude/rules/moai/NOTICE.md` | 6 | 3 | PRESERVE |
| DC-1 | Frontmatter schema field | line shape is `LS-FM` | 48 | 48 | PRESERVE |
| DC-5 | Adjudicated residue | line shape is `LS-FM-FENCED` or `LS-OTHER` | 22 | 17 | PER-ROW adjudication |
| DC-2b | Mirror-capture stamp | line shape is `LS-PROSE-STAMP` and the file is under `.claude/skills/moai-foundation-cc/reference/` | 11 | 11 | PRESERVE |
| DC-2a | Prose authoring stamp | line shape is `LS-PROSE-STAMP` (all remaining) | 80 | 60 | REMOVE |

Row counts sum to 180 — the total occurrence-class row count. Finding counts sum to 148 rather than 135 because the 13 dual-shape findings are counted once per category they span.

**REQ-TDN-002** — The classifier shall operate on occurrence-class rows rather than findings, so that a dual-shape finding receives one disposition per row rather than a single disposition it cannot express.

**REQ-TDN-003** — When a single line carries two or more distinct date literals, the classifier shall bind each literal to its own row and shall not bind a literal to a leading token belonging to a different literal.

**REQ-TDN-019** — When a finding's rows fall into two categories with conflicting dispositions, the remediation shall apply each row's own disposition independently and shall not collapse the finding to a single disposition. Concretely: for the 12 measured `DC-1`/`DC-2a` conflicts, the prose stamp line is deleted and the frontmatter line is preserved, in the same file.

### Triage record

**REQ-TDN-004** — The triage record shall be a tab-separated file `triage.tsv` in this SPEC directory with exactly one header line and one data row per occurrence-class row, carrying seven columns: `file`, `date`, `line_shape`, `line_no`, `category`, `disposition`, `rationale`.

**REQ-TDN-005** — Every `DC-5` row shall carry an adjudicated disposition of `REMOVE` or `PRESERVE`; a placeholder value is a failed triage.

**REQ-TDN-006** — Both recipes shall be committed in this SPEC directory: the finding-set enumeration recipe and the classification recipe.

**REQ-TDN-020** — The committed classifier shall reproduce the guard's finding set exactly: the set of distinct `(file, date)` pairs it emits shall be identical to the set the guard reports, with no additions and no omissions.

### Remediation

**REQ-TDN-007** — When a row's disposition is `REMOVE`, the remediating edit shall delete the date-bearing construct and shall not substitute a placeholder token in its place.

**REQ-TDN-008** — The remediation shall not alter any `DC-3` or `DC-4` date literal.

**REQ-TDN-009** — The remediation shall not delete a `DC-1` frontmatter key, and shall not alter its date value. `DC-1` values are preserved as authored.

**REQ-TDN-011** — The remediation shall not delete a `DC-2b` mirror-capture stamp. These 11 stamps are the reader's only staleness signal for a third-party document mirror, so they are preserved despite sharing a line shape with the `DC-2a` REMOVE set.

### Carve-out mechanism

**REQ-TDN-010** — The carve-out shall be the hybrid mechanism: a structural gate in the guard's per-class scan for the mechanically-decidable recurring shapes (`DC-1` and `DC-4`), plus a content-anchored allowlist for the judgement-call categories (`DC-3`, `DC-2b`, and the PRESERVE subset of `DC-5`). A shape qualifies for a structural gate only where it is both mechanically decidable from the line's own syntax and expected to recur in ordinary authoring; every other preserved row is an allowlist entry.

**REQ-TDN-010b** — The carve-out shall be content-anchored: neither the structural gate nor the allowlist shall identify a preserved row by line number.

**REQ-TDN-012** — The carve-out shall not require a Go source edit when a maintainer adds a new attribution record to `NOTICE.md` or bumps a skill frontmatter `updated:` value.

### CI enforcement

**REQ-TDN-013** — Where the strict tier reports zero findings and the future-date probes pass, the CI configuration shall gain a strict-tier step; while either precondition is unmet, it shall not.

**REQ-TDN-015** — The strict-tier CI step shall invoke the guard as an isolated target by test name and shall not invoke the `internal/template` package as a whole.

**REQ-TDN-014** — The narrow tier and the existing `TestTemplateNeutralityAudit` target shall remain green throughout.

### Guard reporting

**REQ-TDN-016** — When the finding count exceeds the console report cap, the guard shall name a filesystem path holding the complete listing; it shall not truncate without one.

### Template-First discipline

**REQ-TDN-017** — Every remediating edit shall be made under `internal/template/templates/`, and `make build` shall be run afterwards.

**REQ-TDN-018** — The remediation shall not copy content between the template tree and the local `.claude/` / `.moai/` working trees in either direction. Where an edited file is enrolled in the byte-parity mirror allowlist, its local counterpart shall receive the identical edit in the same commit.

---

## §5 Out of Scope

### Out of Scope — the S2 short-sha strict class

- The `S2-short-sha-sentence-final` strict class is not triaged, remediated, or enforced by this SPEC.

### Out of Scope — the narrow-tier class set

- No narrow-tier leak class is added, removed, widened, or narrowed.

### Out of Scope — the neutrality workflow's own class set

- The C1/C2/C4/C5/C6/C8 classes owned by `template_neutrality_audit_test.go` are a disjoint pattern set and are untouched.

### Out of Scope — local-tree internal content

- The `.claude/` and `.moai/` working-tree copies are not neutralized; internal content there is permitted by the isolation doctrine.

### Out of Scope — skill frontmatter schema

- The `updated:` field is not removed from, added to, or redefined in the skill-authoring schema.

### Out of Scope — implementation in plan phase

- No template edit, guard edit, workflow edit, or `make build` occurs during plan phase.

---

## §6 Constraints

- **C-1** Go edits confined to `internal/template/internal_content_leak_test.go`; CI edits confined to `.github/workflows/template-neutrality-check.yaml`.
- **C-2** Template edits confined to `internal/template/templates/**`.
- **C-3** The guard's finding identity is a fixed input; the occurrence-class row is this SPEC's finer unit layered on top of it.
- **C-4** The 16-language neutrality policy and the internal-content isolation doctrine are binding constraints.
