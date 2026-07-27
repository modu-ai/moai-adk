---
id: SPEC-TEMPLATE-DATE-NEUTRALITY-002
title: "Template 2025 date-leak triage, remediation, and S1 year-class widening"
version: "0.3.0"
status: draft
created: 2026-07-27
updated: 2026-07-27
author: manager-spec
priority: P2
phase: "v3.x maintenance"
module: "internal/template"
lifecycle: spec-anchored
tags: "template, neutrality, guard, date, isolation, ci"
tier: L
---

# SPEC-TEMPLATE-DATE-NEUTRALITY-002 — Template 2025 date-leak triage, remediation, and S1 year-class widening

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-27 | Initial plan-phase authoring. Scope is the deferred follow-up recorded in SPEC-TEMPLATE-DATE-NEUTRALITY-001 §5 "Out of Scope — `2025-*` prose authoring stamps". All counts re-measured at `760f09f73` with a year-widened replica of the predecessor's committed classifier. | manager-spec |
| 0.3.0 | 2026-07-27 | Iteration 3 after plan-audit FAIL (0.83). AC-030 was vacuous — its `git log` pickaxe selected the oldest matching commit in *all* history (`ccd6be1f6`, 2026-02-03), making the ancestry test unconditionally true; both ranges are now bounded to the branch. AC-012's marker match was a bare substring that `unfenced` also satisfies; the marker is now the anchored uppercase literal `FENCED`. AC-011 gained the bold-stamp groups its sibling criteria carry. AC-033 added to close AC-029's blindness to an out-of-scope edit *inside* the template tree. REQ-003's sub-shape defaults reconciled with REQ-011's absolute pin. | manager-spec |
| 0.2.0 | 2026-07-27 | Iteration 2 after plan-audit FAIL (0.67). All measurements survived audit (17/17 baselines reproduced). Acceptance-layer repairs: AC-016 retargeted as a function of the M2 adjudication (was unsatisfiable under a permitted PRESERVE outcome); AC-011's dead pattern alternative replaced and its baseline corrected `0` → `9` after measurement showed 9 pre-existing ISO-format documentation lines; AC-012 re-anchored on content; four requirements gained criteria; a Known-AC-limitations section and an edit-scope criterion were added. REQ-024 brings the guard's fourth year-bearing site — a cross-SPEC test doc comment — explicitly into scope. Three `Where` requirements converted to `When`. | manager-spec |

---

## §1 Context

`internal/template/internal_content_leak_test.go` implements a two-tier neutrality guard over the user-distributed template tree (`internal/template/templates/**`). SPEC-TEMPLATE-DATE-NEUTRALITY-001 drove the **strict** tier (`MOAI_TEMPLATE_LEAK_STRICT=1`) to zero findings and adopted it as a CI step.

That guard's `S1-internal-date` class matches `\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b`. Every `2025-*` date literal in the distributed template tree is therefore **structurally invisible** to it. The predecessor recorded this as an explicit out-of-scope boundary and deferred it here, with a stated requirement that remediation and guard-widening ship **together**: remediating without widening leaves the guard permanently blind to the shape; widening without remediating turns the guard red.

### Measured baseline (this tree, `760f09f73`)

| Measurement | Observed |
|---|---:|
| `2025-*` occurrence-class rows | 74 |
| Distinct findings — `(file, date)` pairs | 48 |
| Distinct files carrying a finding | 34 |
| Dual-category findings (rows spanning two categories) | 4 |
| Strict tier today (`202[6-9]` only) | `PASS`, exit 0 |
| Residual carved `202[6-9]` rows (predecessor's PRESERVE set) | 88 |

The 88 residual `202[6-9]` rows independently reproduce the predecessor's disposition arithmetic (`carved out = 100 − k`, with `k = 12`), which validates the measurement instrument used throughout this SPEC.

### The core tension (unchanged in shape, different in composition)

The `2025-*` findings are **not** a homogeneous sweep target, and their category composition differs materially from the predecessor's `202[6-9]` set:

- The predecessor's largest category was `DC-1` (frontmatter schema field) at 48 rows. The `2025` set contains **zero** `DC-1` rows.
- The `2025` set is dominated by `DC-5` (adjudicated residue, 33 rows) and `DC-2a` (prose authoring stamps, 28 rows).
- 13 of the 33 `DC-5` rows are **documentation example values** — dates that are part of a teaching artifact, where deletion damages the example rather than neutralizing a leak.

A uniform sweep over the 74 rows would delete legitimate content. The classification is load-bearing.

---

## §2 Definitions

These terms carry exactly these meanings throughout the SPEC. This section is definitional, not normative. The first four are inherited verbatim from SPEC-TEMPLATE-DATE-NEUTRALITY-001 §2 so that the two triage records remain directly comparable.

- **Finding** — a `(file-path, date-literal)` pair. This is the guard's own unit: it deduplicates matches per file per distinct literal. Measured: **48**.
- **Occurrence-class row** — a `(file-path, date-literal, line-shape, line-number)` tuple. This is the *edit* unit and is finer than a finding. Measured: **74**.
- **Line shape** — the syntactic form of the line carrying the literal. Four shapes: `LS-FM` (an indented YAML frontmatter `updated:` key), `LS-FM-FENCED` (the same key inside a fenced block), `LS-PROSE-STAMP` (a standalone `Last Updated:` / `Updated:` header or footer line), `LS-OTHER` (inline prose, table cells, changelog entries, structured-data values).
- **Dual-category finding** — a finding whose rows fall into two categories carrying different dispositions. Measured: **4**.
- **Documentation-example value** — a date literal whose surrounding construct exists to *teach a format* rather than to record a fact: a frontmatter block shown as a syntax example, a JSON field in a sample payload, a code-sample argument. Deleting such a date degrades the example.
- **Allowlist masking** — the property that a carve-out allowlist entry is keyed on `(file, date)` and therefore suppresses **every** occurrence of that date in that file, not only the intended one. This is what makes a dual-category finding hazardous: an allowlist entry added for a PRESERVE row will also silently suppress a REMOVE row in the same file that was never actually deleted.

---

## §3 Settled Decisions

Recorded so downstream readers do not re-litigate them:

- **Category set** = the predecessor's six categories, carried over unchanged. No new classifier category is introduced. Binding in REQ-TDN2-001.
- **`DC-5` adjudication** = per-row, guided by six named sub-shapes recorded in the `rationale` column. Binding in REQ-TDN2-003.
- **Documentation-example values** = PRESERVE. Binding in REQ-TDN2-010.
- **Mirror-capture stamps (`DC-2b`)** = PRESERVE, on the identical rationale the predecessor recorded. Binding in REQ-TDN2-009.
- **Sequencing** = remediate and carve out first, widen the year class last. Binding in REQ-TDN2-016.
- **Acceptance-criterion coupling** = no criterion may depend on incidental formatting (quoting style, line numbers, whitespace). Binding in REQ-TDN2-018.

---

## §4 Requirements (GEARS)

### Classification

**REQ-TDN2-001** — The classifier shall assign every occurrence-class row to exactly one of the predecessor's six categories, evaluating the rules in the order listed. The category set and the decision rules are carried over unchanged; only the year class in the date pattern differs.

| # | Category | Decision rule (first match wins) | Rows | Findings | Disposition |
|---|---|---|---:|---:|---|
| DC-3 | Functional deadline | the date literal is a pinned deadline value | 0 | 0 | PRESERVE |
| DC-4 | Attribution / license | the file is `.claude/rules/moai/NOTICE.md` | 0 | 0 | PRESERVE |
| DC-1 | Frontmatter schema field | line shape is `LS-FM` | 0 | 0 | PRESERVE |
| DC-5 | Adjudicated residue | line shape is `LS-FM-FENCED` or `LS-OTHER` | 33 | 23 | PER-ROW adjudication |
| DC-2b | Mirror-capture stamp | line shape is `LS-PROSE-STAMP` and the file is under `.claude/skills/moai-foundation-cc/reference/` | 13 | 7 | PRESERVE |
| DC-2a | Prose authoring stamp | line shape is `LS-PROSE-STAMP` (all remaining) | 28 | 22 | REMOVE |

Row counts sum to 74 — the total occurrence-class row count. Finding counts sum to 52 rather than 48 because the 4 dual-category findings are counted once per category they span.

**Three categories are empty, and the emptiness is load-bearing.** `DC-1` is empty because every `updated: 2025-*` line in the tree sits at column 0, while both the classifier's `LS-FM` rule and the guard's `DC-1` structural gate require at least one leading whitespace character. `DC-3` is empty because its decision rule pins a specific deadline literal that no `2025` date matches. `DC-4` is empty because the attribution file carries no `2025` date. None of the three may be assumed to carve a `2025` row.

**REQ-TDN2-002** — The classifier shall operate on occurrence-class rows rather than findings, so that a dual-category finding receives one disposition per row rather than a single disposition it cannot express.

**REQ-TDN2-003** — When the classifier assigns a row to `DC-5`, the triage shall record one of six named sub-shape codes in the row's `rationale` column, so that the per-row adjudication is consistent across rows sharing a construct:

| Code | Sub-shape | Rows | Disposition | Adjudicable at M2? |
|---|---|---:|---|---|
| `EX-FM` | Frontmatter block shown as a syntax example (column-0 `updated:`) | 10 | PRESERVE | **No — pinned by REQ-TDN2-011** |
| `EX-DATA` | Structured-data or code-sample value (JSON field, function argument, path literal) | 3 | PRESERVE | **No — pinned by REQ-TDN2-011** |
| `HIST` | Version-history record (table row or bullet entry pairing a version with its release date) | 14 | per-row | Yes |
| `CREATED` | `Created:` prose stamp | 3 | per-row | Yes |
| `DEADLINE` | Forward-looking review or expiry date (`Next Review:`) | 1 | PRESERVE (default) | Yes |
| `COMPOSITE` | Stamp embedded mid-line in a composite footer rather than standing alone | 2 | per-row | Yes |

**The two columns are not the same claim, and the distinction is load-bearing.** `EX-FM` and `EX-DATA` (13 rows) are documentation-example values, which REQ-TDN2-011 forbids deleting **absolutely** — M2 records `PRESERVE` for those rows without discretion, and a contrary adjudication would violate REQ-TDN2-011 rather than merely differ from a default. For the remaining four codes (20 rows), the stated disposition is a starting position, not a substitute for adjudication.

This pin is also what makes the dual-category findings benign: the two `DC-2b`+`DC-5` findings pair a `DC-2b` row pinned PRESERVE by REQ-TDN2-010 with a `DC-5` row pinned PRESERVE by REQ-TDN2-011, so neither can become a masked deletion. Every row still receives an explicit recorded disposition per REQ-TDN2-007.

**REQ-TDN2-004** — When a finding's rows fall into two categories with conflicting dispositions, the remediation shall apply each row's own disposition independently and shall not collapse the finding to a single disposition. Four such findings are measured.

**REQ-TDN2-005** — The committed classifier shall reproduce the guard's finding set exactly: the set of distinct `(file, date)` pairs it emits shall be identical to the set the widened guard reports, with no additions and no omissions.

### Triage record

**REQ-TDN2-006** — The triage record shall be a tab-separated file `triage.tsv` in this SPEC directory with exactly one header line and one data row per occurrence-class row, carrying the same seven columns as the predecessor's record: `file`, `date`, `line_shape`, `line_no`, `category`, `disposition`, `rationale`.

**REQ-TDN2-007** — Every `DC-5` row shall carry an adjudicated disposition of `REMOVE` or `PRESERVE`; a placeholder value is a failed triage.

**REQ-TDN2-008** — The classification recipe shall be committed in this SPEC directory as a year-widened variant of the predecessor's classifier, so that the finding set is reproducible without re-deriving the classification logic.

### Remediation

**REQ-TDN2-009** — When a row's disposition is `REMOVE`, the remediating edit shall delete the date-bearing construct and shall not substitute a placeholder token in its place.

**REQ-TDN2-010** — The remediation shall not delete a `DC-2b` mirror-capture stamp. These stamps are the reader's only staleness signal for a third-party document mirror, so they are preserved despite sharing a line shape with the `DC-2a` REMOVE set.

**REQ-TDN2-011** — The remediation shall not delete a date that is a documentation-example value. Where such a date is removed, the surrounding construct stops demonstrating the format it exists to teach.

**REQ-TDN2-012** — When a `DC-2a` row sits inside a fenced block, the remediation shall adjudicate that row explicitly rather than sweeping it with the unfenced `DC-2a` set, and the triage shall record the marker as the **uppercase literal `FENCED`** in that row's `rationale` column. One such row is measured; the `DC-2a` decision rule does not inspect fence state, so a mechanical sweep would edit a fenced sample without review. The `rationale` marker exists so the row is identifiable by its own content rather than by a line number.

The literal is pinned to uppercase because the natural annotation for the other 27 `DC-2a` rows is the word "unfenced", which **contains** a lowercase `fenced` as a substring. A lowercase marker would make every such row match the criterion that is supposed to select exactly one. Uppercase plus the non-alphabetic anchor in AC-012 excludes both `unfenced` and `UNFENCED`.

### Carve-out mechanism

**REQ-TDN2-013** — The carve-out shall reuse the predecessor's hybrid mechanism unchanged in shape: the existing structural gate for the mechanically-decidable recurring shapes, plus content-anchored allowlist entries for every preserved row that the structural gate does not cover. Because all three structurally-gated categories are empty for this set, every preserved row requires an allowlist entry.

**REQ-TDN2-014** — The carve-out shall be content-anchored: no allowlist entry shall identify a preserved row by line number.

**REQ-TDN2-015** — When a preserved row and a removed row share a `(file, date)` pair, the verification shall confirm the removed row's deletion by a file-scoped check rather than by the guard's finding count alone. An allowlist entry keyed on that pair masks both rows, so a guard-clean result does not by itself establish that the removal occurred.

### Guard widening

**REQ-TDN2-016** — The `S1-internal-date` class year range shall widen from `202[6-9]` to `202[5-9]`, so that the class ceases to be structurally blind to the `2025` shape.

**REQ-TDN2-017** — The widening shall be applied only after every REMOVE row has been remediated and every PRESERVE row has an allowlist entry. Applying it earlier turns the strict tier red on the full finding set.

**REQ-TDN2-018** — The guard file holds **four** year-bearing sites, and each shall receive its stated treatment: the `S1-internal-date` class pattern is widened (REQ-TDN2-016); the descriptive doc comment naming that class's range is updated to match (REQ-TDN2-024); and the two sibling *patterns* are left unchanged, each for a stated reason — the attribution line matcher already spans the full `20XX` range and needs no edit, and the narrow-tier archive-path class matches a distinct `archive-`-prefixed construct that has zero occurrences in the template tree at either year range. No fifth site exists.

**REQ-TDN2-024** — When the `S1-internal-date` year range changes, the descriptive doc comment that names that range in prose shall be updated to match, and the test function owning that comment shall remain green.

That comment belongs to a **different test function enforcing a different SPEC's acceptance criterion**, so the edit is deliberately minimal: only the prose parenthetical naming the year range changes. No probe value, assertion, or class membership in that test is altered, and the widening is to the *strict*-tier class while that test scans the *default*-tier class set — so the two do not interact behaviourally. The comment is nonetheless in scope because leaving it would make the file describe a class shape it no longer has.

### CI enforcement

**REQ-TDN2-019** — The three test-target invocations in the neutrality workflow shall use a single consistent quoting style.

**REQ-TDN2-020** — No acceptance criterion in this SPEC shall depend on incidental formatting — quoting style, line numbers, or whitespace. A criterion that a formatting-only change can flip is measuring the formatting, not the property.

**REQ-TDN2-021** — The narrow tier, the neutrality-audit target, the strict tier, the cross-SPEC test named in REQ-TDN2-024, and `go build ./...` shall all be green on completion. This is the SPEC's single non-regression requirement.

### Template-First discipline

**REQ-TDN2-022** — Every remediating edit shall be made under `internal/template/templates/`, and the embedded-template build step shall be run afterwards.

**REQ-TDN2-023** — The remediation shall not copy content between the template tree and the local working trees in either direction.

---

## §5 Out of Scope

### Out of Scope — years outside the widened range

- Date literals earlier than `2025` are not triaged, remediated, or matched. The widening stops at `202[5-9]`; extending it further is a separate judgment with its own measurement.

### Out of Scope — the S2 short-sha strict class

- The short-sha strict class is not triaged, remediated, or re-scoped by this SPEC.

### Out of Scope — the narrow-tier class set

- No narrow-tier leak class is added, removed, widened, or narrowed. This explicitly includes the narrow-tier archive-path class, whose `archive-`-prefixed construct is measured at zero occurrences for both `202[5]` and `202[6-9]`, making a widening there a no-op today.

### Out of Scope — the neutrality workflow's own class set

- The classes owned by the separate neutrality-audit test are a disjoint pattern set and are untouched.

### Out of Scope — local-tree internal content

- The local working-tree copies are not neutralized; internal content there is permitted by the isolation doctrine.

### Out of Scope — skill frontmatter schema

- The `updated:` field is not removed from, added to, or redefined in the skill-authoring schema. Where a `2025` date sits in a frontmatter block shown as an example, the disposition concerns that example's content only.

### Out of Scope — generated artifacts

- The embedded-template catalog artifact is regenerated by the build step and is not a remediation target. It is measured at zero `2025` date literals.

### Out of Scope — the predecessor's artifacts

- SPEC-TEMPLATE-DATE-NEUTRALITY-001 is `completed`. Its requirements, acceptance criteria, and triage record are not amended, reopened, or re-scored by this SPEC.

### Out of Scope — implementation in plan phase

- No template edit, guard edit, workflow edit, or build step occurs during plan phase.

---

## §6 Constraints

- **C-1** Go edits confined to `internal/template/internal_content_leak_test.go`; CI edits confined to the neutrality workflow file.
- **C-2** Template edits confined to `internal/template/templates/**`.
- **C-3** The guard's finding identity is a fixed input; the occurrence-class row is this SPEC's finer unit layered on top of it.
- **C-4** The 16-language neutrality policy and the internal-content isolation doctrine are binding constraints.
- **C-5** The predecessor's carve-out entries and its remediated state are a fixed baseline; this SPEC adds to the allowlist and does not rewrite existing entries.
