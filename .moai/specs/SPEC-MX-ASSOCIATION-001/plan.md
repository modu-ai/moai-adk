# plan.md — SPEC-MX-ASSOCIATION-001

> Implementation plan. Milestones are priority-ordered by decision-reversibility (highest-change-likelihood first), per the orchestrator's Approach-First ordering. No time estimates.

## §A. Context

The `@MX:SPEC:<SPEC-ID>` sub-line is an explicit author-intended SPEC link, but the scanner currently drops its content and the associator never sees it. Verified gap in `internal/mx/scanner.go` (sub-line sentinel path at lines 110-139 handles only `@MX:REASON` and `@MX:UPGRADE`) and `internal/mx/tag.go` (no field carries the content). Result: 9.7 % coverage (955 / 9,858), with 88 `@MX:SPEC` sub-lines contributing zero.

This SPEC is **additive only**: capture the sub-line, feed it to `Associate`, leave the two existing sources untouched.

## §B. Known Issues

1. **Constraint tension (RESOLVED).** The user constraint says "no schema/model changes beyond reading the existing sub-line field." Research shows there is NO existing sub-line field for `@MX:SPEC` — the content is dropped, not stored. RESOLVED — single additive field accepted; see Decision 1. The most faithful in-scope interpretation is: add ONE additive `string` field parallel to `Reason`.
2. **`ExtractSpecIDs` reuse vs. narrow regex.** The body-source regex `SPEC-[A-Z0-9][A-Z0-9-]*` is greedy and will match IDs that are not real SPECs. Reusing it for the sub-line source keeps the two sources consistent; the validation decision (REQ-MX-ASSOC-003) handles the unresolved case.
3. **Sub-line proximity / dangling refs.** `@MX:SPEC:` appearing with no preceding tag in the file is malformed; the scanner must warn rather than crash or silently pair it with a tag from a prior file.

## §C. Pre-flight Decisions (decision-reversibility-ordered)

### Decision 1 — Model touch: single additive `Tag.SpecRef string` field (HIGHEST reversibility)

**DECISION: YES, add the field. Accepted.**

- Add `SpecRef string \`json:"specRef,omitempty"\`` to `Tag` in `internal/mx/tag.go`, immediately after `Reason`. Rationale:
  - It is the exact structural analogue of `Reason` (a single sub-line content carrier with `omitempty`).
  - It keeps `Body` clean (does not pollute the description text), so `DetectChanges` and sidecar consumers reading `body` see no regression.
  - It is serialised in the sidecar, making the new association observable — a hard requirement for an acceptance-testable change.
- This is NOT the deferred §6 ontology — it is the minimal model touch that completes the already-recognized-but-dropped `@MX:SPEC` sub-line handling (restoring consistency with `Reason`), and it honors "preserve MX simplicity" by removing an asymmetry rather than adding a new relation type.
- Rejected alternatives:
  - **(a) Append the SPEC ID into `Body` at scan time.** Zero model change, BUT mutates `body` in the sidecar JSON (regression risk for consumers / `DetectChanges` semantics / display). Rejected.
  - **(b) Re-scan raw source lines at association time.** Pushes file I/O into the associator, which is currently a pure function of `(tag, specModules)`. Couples associator to the filesystem and breaks the existing unit-test shape. Rejected.
  - **(c) Capture into a `[]string`.** Over-engineered — authors write one `@MX:SPEC` per tag in practice (verified: the protocol describes a single optional sub-line). A scalar field matches the protocol. Rejected.

The constraint phrasing ("reading the existing sub-line field") presupposes the field exists; it does not. A single additive scalar field is the minimal model touch that still honors the constraint's intent (no ontology overhaul, no new sub-line keys, no schema migration). The orchestrator has RESOLVED the constraint-tension in favor of this single additive field; the decision will be re-confirmed at the Implementation Kickoff Approval gate.

### Decision 2 — Validation strategy: flag-but-keep (MEDIUM reversibility)

**Recommendation: flag-but-keep (`UnresolvedSpecRef` scanner warning + keep the ID in `spec_associations`).**

| Option | Behavior | Pro | Con |
|---|---|---|---|
| **flag-but-keep (recommended)** | Warning in `GetWarnings`; ID stays in `spec_associations` | Consistent with body-source (which also doesn't validate); preserves author intent; warnings grep-able; `moai mx query --spec <stale-id>` still returns the tag | Stale IDs appear in query results |
| drop-with-warning | Warning emitted; ID dropped from `spec_associations` | Cleaner query results | Silently demotes explicit author intent; diverges from body-source behavior; a stale ref becomes invisible |
| hard-error | Abort scan | Strongest correctness | Breaks the 1,567+ tag validator on any stale ref; unacceptable regression risk |

Recommendation rationale: the sub-line is the strongest possible author-intent signal (an explicit `@MX:SPEC:`). Demoting it when stale hides information; keeping it with a warning surfaces drift without breaking queries. The body-source already accepts unvalidated IDs, so flag-but-keep preserves cross-source consistency.

### Decision 3 — Sub-line pairing window (LOW reversibility, follows existing pattern)

Mirror `@MX:REASON`'s pending-tag proximity discipline. Concretely: when the scanner's `errSubLineKind` branch sees a line whose uppercase form contains `@MX:SPEC`, extract the SPEC-ID text and attach it to (in order):
1. the most recent tag in `tags` (same file), if one exists; otherwise
2. emit a `DanglingSpecRef` warning and continue.

No 3-line proximity cutoff is required (unlike `@MX:REASON` for WARN), because `@MX:SPEC` is optional metadata, not a mandatory pairing. A later standalone tag in the same file ends the window implicitly (the next `@MX:SPEC` pairs with that later tag).

## §D. Constraints (carried from spec.md §D)

- Additive only. No ontology overhaul. No new sub-line keys. No sidecar migration (additive `omitempty` field).
- Validator must stay green on the existing 1,567+ tag set.

## §E. Self-Verification (plan-phase)

- [ ] SPEC ID regex self-check: `SPEC-MX-ASSOCIATION-001` → PASS (run, verbatim output cited).
- [ ] Frontmatter: all 12 canonical fields present; `status: draft`; ISO dates; `module: internal/mx`.
- [ ] Gap verified in code (scanner.go:35/110-139/321-323; tag.go struct; spec_association.go:42) — citations in spec.md §B.2.
- [ ] Coverage baseline measured this date: 9.7 % (955/9,858) — cited in spec.md §B.3.
- [ ] Out of Scope section satisfies `OutOfScopeRule` (≥1 `### Out of Scope — <topic>` H3 with `-` bullets).

## §F. Milestones (priority-ordered)

### M1 — Model + scanner capture (Decision 1 + Decision 3)
- Add `Tag.SpecRef string \`json:"specRef,omitempty"\`` to `internal/mx/tag.go`.
- In `internal/mx/scanner.go` `errSubLineKind` branch: add an `@MX:SPEC` arm that extracts the SPEC-ID text and attaches it to the most recent tag (or emits `DanglingSpecRef`). Reuse the `extractReason` shape (new `extractSpecRef` helper or a shared helper).
- Emit `UnresolvedSpecRef` warning here is DEFERRED to M3 (validation needs the `specModules` set, which the scanner does not hold). M1 only captures.

### M2 — Associator consumes the sub-line source (REQ-MX-ASSOC-002)
- In `internal/mx/spec_association.go` `Associate`: after the body-based loop, add a third source loop over the captured sub-line ID (if non-empty), de-duplicated via the existing `seen` map.
- Order of sources preserved: path → body → sub-line (additive, deterministic, de-duped).

### M3 — Validation (REQ-MX-ASSOC-003, Decision 2)
- The associator (which holds `specModules`) is the natural place to validate, since the scanner does not have the known-SPEC set.
- **Tradeoff re-scoped at implementation time**: validation could live in the scanner (needs the set plumbed in) or in the associator (already has the set). Recommendation: associator-side — add an `UnresolvedSpecRef` return / warning channel on `Associate`, OR emit the warning from the resolver call site (`resolver_query.go:363`) which holds both the associator and the scanner. Decision deferred to M3 implementation; the SPEC binds only the observable behavior (warning emitted + ID kept), not the emission site.

### M4 — Characterization tests + new AC coverage (REQ-MX-ASSOC-004)
- Add characterization tests capturing TODAY's path-based and body-based outputs on a fixed fixture tag set, asserting byte-for-byte equality after the change.
- Add the new AC tests (acceptance.md §D.2).

### M5 — Coverage measurement + docs touch
- Re-run the measurement harness (see acceptance.md AC-MX-ASSOC-002 procedure) on the real repo; record the post-change number.
- Touch-up: godoc on the new field; one-line note in `mx-tag-protocol.md` that `@MX:SPEC` now drives association (factual, not a protocol change). NOTE: `mx-tag-protocol.md` is a distributed rule file — the note must be content-neutral and free of internal SPEC IDs per the template-isolation doctrine; coordinate via the Template-First cycle if mirrored.

## §G. Anti-Patterns

- **AP-1**: Mutating `tag.Body` to embed the SPEC ID (Decision 1 alternative (a)). Pollutes sidecar `body`, regresses `DetectChanges` and display. Do not do this.
- **AP-2**: Validating only the sub-line source while leaving the body-source unvalidated, then dropping unresolved sub-line IDs. Creates an inconsistency between two sources that should behave identically. Use flag-but-keep (Decision 2).
- **AP-3**: Adding a 3-line proximity cutoff for `@MX:SPEC` modelled on `@MX:REASON` for WARN. `@MX:SPEC` is optional metadata, not a mandatory pairing; a cutoff would silently drop legitimate sub-lines that follow a blank line. Do not add a cutoff.
- **AP-4**: Hardcoding the post-change coverage number in a test. The measurement runs against the live repo and drifts; assert the DELTA and a FLOOR, not an exact percentage.
- **AP-5**: Touching `ExtractSpecIDs` to add validation. That function is the body-source extractor and is reused; validating inside it would change body-source behavior and violate REQ-MX-ASSOC-004. Keep validation at the associator/resolver layer.

## §H. Cross-References

- spec.md: `.moai/specs/SPEC-MX-ASSOCIATION-001/spec.md`.
- acceptance.md: `.moai/specs/SPEC-MX-ASSOCIATION-001/acceptance.md`.
- Diagnosis report: `.moai/reports/mx-system-analysis-activation-20260804.html`.
- Related: `SPEC-MX-001`, `SPEC-MX-002`.
