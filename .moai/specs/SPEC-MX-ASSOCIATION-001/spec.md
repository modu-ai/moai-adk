---
id: SPEC-MX-ASSOCIATION-001
title: "Drive @MX spec_associations from the @MX:SPEC sub-line"
version: 0.1.0
status: draft
created: 2026-08-04
updated: 2026-08-04
author: manager-spec
priority: P1
phase: "v3.1.0"
module: internal/mx
lifecycle: spec-anchored
tags: "mx, tag, spec-association, sub-line, coverage"
tier: M
---

# SPEC-MX-ASSOCIATION-001 — Drive spec_associations from the @MX:SPEC sub-line

## HISTORY

- 2026-08-04 — Created (manager-spec, plan-phase). P1-3 item from the MX activation queue (`mx-system-analysis-activation-20260804.html` §3 / §5). Confirmed gap: the `@MX:SPEC:` sub-line is silently dropped by the scanner and never reaches `SpecAssociator.Associate`; current `spec_associations` coverage measured at 9.7% (955 / 9,858 tags).

## §A. User Story

As an MoAI operator running `moai mx query --spec SPEC-XXX` (and as an agent consuming the `spec_associations` sidecar field), I want any `@MX` tag that carries an explicit `@MX:SPEC:<SPEC-ID>` sub-line to be associated with that SPEC ID — regardless of whether the tag's file path happens to sit under a SPEC `module:` path or the SPEC ID happens to appear in the tag `Body` — so that explicit, author-intended SPEC links are reflected in query results and the sidecar at the rate authors write them.

Today the `@MX:SPEC:` sub-line is a second-class citizen: it is a recognized scanner sub-line, but its content is discarded at scan time and never stored on the `Tag`. The two existing association sources (path-based, body-regex) cannot see it. Coverage is as a consequence dominated by incidental path matches, and the explicit `@MX:SPEC` signal (88 occurrences in production) contributes zero associations.

## §B. Background & Confirmed Gap

### B.1 Current association mechanism

`internal/mx/spec_association.go` `SpecAssociator.Associate(tag)` combines two sources:

1. **Path-based** — `tag.File` prefix-matches one of the SPEC frontmatter `module:` paths (`LoadSpecModules`, `spec_loader.go`).
2. **Body-based** — `ExtractSpecIDs(tag.Body)` runs the regex `SPEC-[A-Z0-9][A-Z0-9-]*` over the tag Body text.

### B.2 Confirmed gap (verified in code this date)

The prompt's gap hypothesis was partially correct. The verified mechanism is sharper than "parsed into a dedicated sub-line structure":

- `internal/mx/scanner.go:35` — `recognizedSubLineKinds` includes `"SPEC"`.
- `internal/mx/scanner.go:321-323` — `parseTag` returns the `errSubLineKind` sentinel for any recognized sub-line kind BEFORE the validity switch.
- `internal/mx/scanner.go:110-139` — the per-line loop's `errSubLineKind` branch pairs ONLY `@MX:REASON` (onto a pending WARN/ANCHOR) and `@MX:UPGRADE` (onto a pending DEBT, clearing `RotRisk`). There is NO branch for `@MX:SPEC`. The line is `continue`d and its content is **dropped entirely**.
- `internal/mx/tag.go` — the `Tag` struct has dedicated sub-line fields for `Reason` (`@MX:REASON`) and derived state for `@MX:CEILING`/`@MX:UPGRADE` (`RotRisk`), but **no field carries `@MX:SPEC` content**. It is neither in `Body` nor in a dedicated field.

Net effect: `ExtractSpecIDs(tag.Body)` cannot see the SPEC ID named in an `@MX:SPEC:` sub-line, because that content never enters `tag.Body`. Association then fires only on incidental path/body matches.

### B.3 Measured coverage (this date, real repo)

Measured by invoking `LoadSpecModules` + `Scanner.ScanDir` + `SpecAssociator.Associate` over the full project tree from the main checkout:

| Metric | Value |
|---|---|
| SPECs loaded (known-SPEC set size) | 568 |
| Tags scanned | 9,858 |
| Tags with ≥1 `spec_associations` entry | 955 |
| **Coverage** | **9.7 %** |
| `@MX:SPEC:` sub-lines in production source | 88 |
| `@MX:SPEC:` sub-lines contributing to `spec_associations` today | 0 |

The 88 `@MX:SPEC` sub-lines are a pure additive opportunity: they currently contribute nothing.

## §C. Requirements (GEARS)

The subject is generalized per GEARS (`<subject>` = the named component, not "the system").

### REQ-MX-ASSOC-001 — Capture the @MX:SPEC sub-line on the Tag

**When** the scanner encounters a line whose `@MX:` kind is `SPEC` within the sub-line window of a tag, the scanner shall capture the sub-line's SPEC-ID content onto that tag in a form readable by `SpecAssociator.Associate`, rather than discarding it.

- *Implementation note (non-normative)*: the minimal, model-faithful capture is a single additive `string` field on `Tag` parallel to the existing `Reason` field. The chosen field MUST be serialized in the sidecar JSON so the association is observable to consumers. See plan.md §C for the constraint discussion.
- *Sub-line window*: an `@MX:SPEC:` line pairs with the most recent preceding standalone tag in the same file, subject to the same proximity discipline already used for `@MX:REASON` (the scanner's existing pending-tag tracking). An `@MX:SPEC:` line with no preceding tag in the file is a scanner warning (`DanglingSpecRef`), not a hard error.

### REQ-MX-ASSOC-002 — Associate from the captured sub-line

**Where** a tag carries a captured `@MX:SPEC` SPEC-ID, the associator shall include that SPEC ID in the tag's `spec_associations` result, independently of the path-based and body-based sources.

- The sub-line source is ADDITIVE: it MUST NOT remove or alter any path-based or body-based association already produced for the same tag.
- De-duplication reuses the existing `seen` map so a SPEC ID named in both the sub-line and the Body appears once.

### REQ-MX-ASSOC-003 — Validation of unresolved sub-line SPEC IDs

**When** a captured `@MX:SPEC` SPEC-ID does not resolve to any SPEC in the known-SPEC set (the `specModules` keys loaded by `LoadSpecModules`), the associator shall flag-but-keep: emit a `UnresolvedSpecRef` scanner warning carrying the unresolved ID and the source `file:line`, AND still include the ID in `spec_associations`.

- Rationale (see plan.md §D for the full tradeoff): consistency with the existing body-based source (which also does not validate against the known set); preservation of explicit author intent (a stale ref is information); warnings are grep-able so drift surfaces for follow-up without breaking `moai mx query --spec <stale-id>`.
- The associator MUST NOT crash or abort the scan on an unresolved ID.

### REQ-MX-ASSOC-004 — Regression safety for existing associations

**While** the sub-line capture and association paths are introduced, the scanner shall preserve the existing path-based and body-based association outputs byte-for-byte for every tag that does not carry an `@MX:SPEC` sub-line.

- The 1,567+ production tags validated green today MUST remain valid green.
- No existing association (path-based or body-based) may be removed or altered as a side effect of this change.

### REQ-MX-ASSOC-005 — Coverage lift

**Where** the sub-line association path is active, the spec_associations coverage on the real repository shall rise from the measured 9.7 % baseline by a net-new contribution of at least 40 tags attributable to the sub-line source.

- The lift is measured as `(tags whose sole or additional association comes from the sub-line source)`. See acceptance.md AC-MX-ASSOC-002 for the exact measurement procedure.
- Target floor: post-change coverage ≥ 10.2 %, with the sub-line source contributing ≥ 40 net-new associated tags. Derivation: of the 88 `@MX:SPEC` sub-lines, the ≥40 floor assumes ~50 reference SPECs not already path/body-associated, allowing ≤30 % unresolved-against-known-set and ≤10 % overlap with existing body matches (88 × ~57 % effective yield ≈ 50 net-new). The floor leaves headroom for sub-lines that already overlap a path/body association and for unresolved IDs.

## §D. Constraints

- **[HARD] No ontology / model overhaul.** The report's §6 ontology proposal is explicitly deferred by the user. This SPEC works within the existing `Tag` / sub-line / sidecar model. The ONLY model touch permitted is a single additive `string` field on `Tag` to carry the captured `@MX:SPEC` content — this is parallel to the existing `Reason` field (same lifecycle, same sub-line origin) and is NOT an ontology change. This is NOT the deferred §6 ontology: it is the minimal model touch that completes the already-recognized-but-dropped `@MX:SPEC` sub-line handling (restoring consistency with `Reason`), and it honors "preserve MX simplicity" by removing an asymmetry rather than adding a new relation type. See plan.md §C Decision 1 for the accepted rationale.
- **[HARD] Additive only.** The change MUST NOT alter existing path/body associations or break the validator. Regression safety is delivered by characterization on a representative fixture set (acceptance.md AC-MX-ASSOC-004, ≥8 fixtures asserting byte-for-byte equality) PLUS a production-scale regression guard on the full corpus (AC-MX-ASSOC-005: validator-green + per-tag association count never drops below baseline).
- **[HARD] Preserve MX simplicity.** No new sub-line keys, no schema migration of existing sidecars beyond the additive field (omitted when empty via `omitempty`), no new config keys.
- Consistency reference: `.claude/rules/moai/workflow/mx-tag-protocol.md` (sub-line syntax, `@MX:SPEC` is OPTIONAL, `[AUTO]` prefix convention unchanged).

## §E. Open Questions

None. The model-touch decision is resolved — a single additive `Tag.SpecRef string` field parallel to `Reason` is accepted as in-scope (see plan.md §C Decision 1 for the rationale). The decision will be re-confirmed at the Implementation Kickoff Approval gate.

## §F. Dependencies

- `internal/mx/tag.go` — Tag model.
- `internal/mx/scanner.go` — sub-line parsing, `errSubLineKind` path.
- `internal/mx/spec_association.go` — `Associate`, `ExtractSpecIDs`.
- `internal/mx/spec_loader.go` — `LoadSpecModules` (known-SPEC set, validation basis).
- `internal/mx/resolver_query.go:363` — `specAssociator.Associate(tag)` call site.

## §G. Out of Scope

### Out of Scope — Ontology / model overhaul

- The report's §6 ontology proposal (rich tag→SPEC relation model, typed links, provenance fields). Deferred by user decision; not proposed by this SPEC.

### Out of Scope — New sub-line keys

- Introducing new `@MX:` sub-line keys (e.g. `@MX:OWNER`, `@MX:REVIEWED`). The sub-line set stays as-is.

### Out of Scope — Validating path-based / body-based associations

- This SPEC does NOT add resolution validation to the existing path-based or body-based association sources. REQ-MX-ASSOC-003 binds only the new sub-line source. Lifting validation onto the legacy sources is a separate decision and is out of scope.

### Out of Scope — Retroactive `@MX:SPEC` authoring

- Bulk-adding `@MX:SPEC` sub-lines to existing tags to inflate coverage. This SPEC captures the signal authors already write; it does not author new signal.

### Out of Scope — Sidecar migration tooling

- A migration / rewrite utility for existing sidecar JSON files. The additive field uses `omitempty` so existing sidecars deserialize unchanged.

### Out of Scope — Query/UI surfacing of `UnresolvedSpecRef`

- Dedicated `moai mx query` filters or dashboard views for unresolved refs. The warning is emitted to the scanner warning channel (`GetWarnings`) and is grep-able; richer surfacing is deferred.

## §H. Cross-References

- Diagnosis report: `.moai/reports/mx-system-analysis-activation-20260804.html` §3 (~line 207), §5 (~line 352).
- MX tag protocol: `.claude/rules/moai/workflow/mx-tag-protocol.md`.
- Related SPECs: `SPEC-MX-001`, `SPEC-MX-002` (MX subsystem precedents).
- Code surfaces: see §F Dependencies.
