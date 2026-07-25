---
id: SPEC-TEMPLATE-DATE-NEUTRALITY-001
title: "Template date-leak triage, remediation, and S1 guard refinement"
version: "0.1.0"
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
| 0.1.0 | 2026-07-25 | Initial plan-phase authoring. Scope derived from a measured strict-tier guard run (135 findings, 116 files) and a full independent re-enumeration of the same finding set. | manager-spec |

---

## §1 Context

`internal/template/internal_content_leak_test.go` implements a two-tier neutrality guard over the user-distributed template tree (`internal/template/templates/**`). The **narrow** tier runs in CI and is green. A **strict** tier, activated by `MOAI_TEMPLATE_LEAK_STRICT=1`, adds two broader classes — `S1-internal-date` (any `202[6-9]-MM-DD` literal) and `S2-short-sha-sentence-final` — and is enforced nowhere.

### Measured baseline (this tree, `c7309aeb6`)

| Measurement | Command | Observed |
|---|---|---|
| Strict-tier result | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` | `exit=1`, `template internal-content leak detected (135 occurrences, mode=strict)` |
| Narrow-tier result | same command without the env var | `exit=0`, `ok github.com/modu-ai/moai-adk/internal/template` |
| Strict-mode CI wiring | `grep -rn "LEAK_STRICT" --include="*.yaml" --include="*.yml" --include="Makefile" .` excluding the test file | `exit=1` (no matches) |
| Findings by class | strict-run report body | `class=S1-internal-date` on every reported row |
| Distinct files | independent re-enumeration | 116 |

The guard's report is capped at the first 50 rows (`limit := 50`), so 85 of the 135 findings are never surfaced to a maintainer running the guard.

### Finding-identity semantics (load-bearing)

`collectLeakViolations` deduplicates matches **per (file, distinct match literal)**. A finding is therefore a `(file, date-literal)` pair, not a raw occurrence. A file containing the same date on five lines contributes one finding; a `Last Updated:` prose line and a frontmatter `updated:` field carrying the same date in the same file collapse into a single finding. Any triage inventory must reproduce this identity or its counts will not reconcile with the guard.

### The core tension

The 135 findings are **not** a homogeneous sweep target. The guard's own source comment records why S1 was left unenforced: "not enforced by default to avoid blocking on generic dates in CHANGELOG entries about external Anthropic releases, etc." A mechanical partition of all 135 (see §2 REQ-TDN-001) yields at least four kinds with opposite correct treatments — internal authoring metadata that is meaningless in a distributed template, schema-required frontmatter fields, functionally load-bearing deadline dates, and attribution records. Deleting the wrong kind is a regression, not a win.

This SPEC is therefore a **triage-and-guard-refinement** SPEC, not a sweep SPEC.

---

## §2 Requirements (GEARS)

### Classification

**REQ-TDN-001** — The classification scheme shall partition every S1 date finding into exactly one of five categories, using a decision rule evaluated in the stated order:

| # | Category | Decision rule (first match wins) | Measured count | Default disposition |
|---|---|---|---|---|
| DC-1 | Frontmatter schema field | The match sits on a YAML frontmatter line whose key is `updated`, `created`, or `version` | 48 | PRESERVE |
| DC-2 | Prose authoring stamp | The match sits on a standalone `Last Updated:` / `Updated:` / `# Updated:` document header or footer line | 58 | REMOVE, except the mirror-capture sub-case (REQ-TDN-011) |
| DC-3 | Functional deadline | The date literal is referenced as a future expiry or deadline that a reader must act on | 9 | PRESERVE |
| DC-4 | Attribution / license | The match sits in an attribution or license record (`NOTICE.md` import records) | 3 | PRESERVE |
| DC-5 | Adjudicated residue | Everything not matched above | 17 | PER-FINDING adjudication (REQ-TDN-005) |

The five counts shall sum to the guard's own reported finding count.

**REQ-TDN-002** — The `<subject>` of the decision rule is the **finding**, defined as a `(file-path, date-literal)` pair per the guard's dedup semantics; the classifier shall not operate on raw line occurrences.

**REQ-TDN-003** — Where a finding's category cannot be decided from its surrounding line alone (a line carrying two or more distinct date literals, such as a changelog line whose leading token is `Updated: <date-A>` while the finding's literal is `<date-B>`), the classifier shall route that finding to DC-5 rather than binding it to the leading token's category.

### Triage record

**REQ-TDN-004** — The triage shall be recorded in a committed inventory artifact under this SPEC directory, carrying one row per finding with: file path, date literal, category, disposition (REMOVE or PRESERVE), and a one-line rationale.

**REQ-TDN-005** — For every DC-5 finding, the inventory shall carry an explicit adjudicated disposition; a DC-5 row with an unset or placeholder disposition is a failed triage.

**REQ-TDN-006** — The inventory shall be regenerable: a committed enumeration recipe shall reproduce the finding set from the tree, so a reviewer can re-derive the inventory rather than trust it.

### Remediation

**REQ-TDN-007** — When a finding's disposition is REMOVE, the remediating edit shall delete the date-bearing construct rather than replace the date with a placeholder token, so that no new grep surface is introduced.

**REQ-TDN-008** — The remediation shall not alter any DC-3 or DC-4 date literal. (Unwanted-behavior requirement.)

**REQ-TDN-009** — The remediation shall not delete a DC-1 frontmatter key. Skill frontmatter carries `updated:` as a documented schema field; removing the key is a schema break, distinct from neutralizing its value.

### Carve-out mechanism

**REQ-TDN-010** — Where a finding's disposition is PRESERVE, the guard shall not report it once strict mode is enforced. The carve-out mechanism shall be selected in `design.md` from the enumerated options and shall be content-anchored: it shall not identify a carve-out by line number.

**REQ-TDN-011** — Where a preserved date is a mirror-capture stamp on an external-document mirror (a `Updated: <date>` line on a file that mirrors third-party documentation), the carve-out shall preserve it, because the stamp is the reader's only signal of the mirror's staleness.

**REQ-TDN-012** — The carve-out shall not require a code change for a future legitimate date whose shape is already covered by an accepted structural rule. A mechanism that forces a Go-file edit for every future `NOTICE.md` import entry is a regression and shall be rejected.

### CI enforcement

**REQ-TDN-013** — The CI enforcement decision (whether the strict tier becomes CI-enforced, and under what preconditions) shall be recorded in `design.md` with its preconditions stated as verifiable conditions, and shall be implemented only after the strict tier reaches zero unallowlisted findings.

**REQ-TDN-014** — While strict mode is not yet CI-enforced, the narrow tier shall remain green and the existing `template-neutrality-check.yaml` workflow shall remain green. (Non-regression requirement.)

**REQ-TDN-015** — When strict mode becomes CI-enforced, the workflow shall run the strict tier in isolation by test name, matching the existing workflow's isolated-target convention, so that unrelated pre-existing package failures do not gate it.

### Guard reporting

**REQ-TDN-016** — The guard's report cap shall stop silently hiding findings: when the finding count exceeds the cap, the guard shall either report all findings or emit a machine-readable full listing to a path named in the truncation message.

### Template-First discipline

**REQ-TDN-017** — Every remediating edit shall be made under `internal/template/templates/` first, followed by `make build`, per the Template-First rule.

**REQ-TDN-018** — The local working-tree counterparts under `.claude/` and `.moai/` shall not be updated by copying from the template tree, nor the reverse. 115 of the 116 affected template files have a local counterpart; the local copies legitimately retain internal-development content that the template copies must not. The local-mirror disposition shall be decided per file family in `design.md`, and byte-parity shall be verified only for the file families the mirror-parity guard actually enrols.

---

## §3 Out of Scope

### Out of Scope — the S2 short-sha strict class

- The `S2-short-sha-sentence-final` strict class is not triaged, remediated, or enforced by this SPEC. It was observed contributing zero findings in the strict run, but zero-in-this-tree is not a triage; S2 remains owned by its existing class definition.

### Out of Scope — the narrow-tier class set

- No narrow-tier leak class is added, removed, widened, or narrowed. The narrow tier is green and stays untouched.

### Out of Scope — the neutrality workflow's own class set

- The C1/C2/C4/C5/C6/C8 classes owned by `template_neutrality_audit_test.go` are a disjoint pattern set and are not in scope. Only the C3 date class, owned by the leak test's strict tier, is addressed here.

### Out of Scope — local-tree internal content

- The `.claude/` and `.moai/` working-tree copies are not neutralized. Internal SPEC IDs, dates, and audit citations in the local tree are permitted by the isolation doctrine and are not defects.

### Out of Scope — retroactive doctrine rewrite

- No prior SPEC's artifacts are rewritten, and no already-shipped template release is retro-neutralized. The remediation is forward-looking against the current tree.

### Out of Scope — implementation in plan phase

- No template edit, guard edit, workflow edit, or `make build` is performed during plan phase. This SPEC's plan-phase deliverable is the artifact set only.

---

## §4 Constraints

- **C-1** Go source edits are confined to `internal/template/internal_content_leak_test.go` and, if CI enforcement is adopted, `.github/workflows/template-neutrality-check.yaml`.
- **C-2** Template edits are confined to `internal/template/templates/**`.
- **C-3** The guard's finding identity (`(file, date-literal)` dedup) is a fixed input, not a variable this SPEC changes.
- **C-4** The 16-language template-neutrality policy and the internal-content isolation doctrine are binding constraints, not subjects of revision.

---

## §5 Open Questions

Recorded for the plan audit and for the user; each is a place where the correct treatment cannot be determined without a human call. See `plan.md` §B for the full statement of each.

- **[NEEDS CLARIFICATION: mirror-capture stamps]** — 11 of the 43 `2026-01-06` findings sit on `Updated:` lines in `moai-foundation-cc/reference/*-official.md`, which mirror third-party documentation. The date is the mirror's capture date. REQ-TDN-011 proposes preservation; the alternative (removal) is defensible if the mirror's staleness is signalled some other way.
- **[NEEDS CLARIFICATION: DC-1 disposition]** — 48 findings are frontmatter `updated:` / `created:` / `version:` fields. Preserving them means the S1 regex must be structurally narrowed or 48 allowlist entries maintained. Narrowing the regex is the larger design change.
- **[NEEDS CLARIFICATION: internal-incident prose]** — several DC-5 findings anchor a policy or an incident by date ("prevents the 2026-04-21 incident recurrence", "Advisory — 2026-05-17 Policy"). Removing the date may orphan the reference; keeping it leaks an internal work date.
