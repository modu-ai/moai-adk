# SPEC-SIMPLICITY-LADDER-001 — Acceptance Criteria

## §D. AC Matrix

| AC ID | REQ | Severity | Verification |
|-------|-----|----------|--------------|
| AC-SL-001 | REQ-1.1 | MUST | Constitution #4 contains the ordered 6-rung ladder |
| AC-SL-002 | REQ-1.2 | MUST | Ladder is language-neutral (no JS/Python/single-language tokens) |
| AC-SL-003 | REQ-1.3 | MUST | "Never simplify away" safety carve-out present + cross-references existing rules |
| AC-SL-004 | REQ-1.4 | MUST | karpathy-quickref carries ONE cross-ref line, does not restate the ladder |
| AC-SL-005a | REQ-1.* | MUST | `diff -q` clean: live `.../core/moai-constitution.md` vs template mirror (per-file) |
| AC-SL-005b | REQ-1.* | MUST | `diff -q` clean: live `.../development/karpathy-quickref.md` vs template mirror (per-file) |
| AC-SL-006 | REQ-2.1 | MUST | `@MX:DEBT` defined in mx-tag-protocol.md + constitution @MX list |
| AC-SL-007 | REQ-2.2 | MUST | `@MX:CEILING` + `@MX:UPGRADE` sub-lines documented; inline-comment (not JSON) |
| AC-SL-008 | REQ-2.3 | MUST | `moai mx query --kind DEBT --json` emits `"rotRisk": "no-trigger"` for a no-`@MX:UPGRADE` marker |
| AC-SL-009 | REQ-2.4 | MUST | `internal/mx/` scanner accepts `@MX:DEBT` (no "unknown tag kind: DEBT" error) |
| AC-SL-009b | REQ-2.4 | MUST | `@MX:CEILING` / `@MX:UPGRADE` sub-lines scan WITHOUT a "unknown tag kind" error (recognized-sub-line set) |
| AC-SL-009c | REQ-2.4 | MUST | **Regression guard (D-NEW-1)**: a standalone `// @MX:LEGACY: <text>` comment STILL returns 1 tag of kind `MXLegacy` after the sub-line-set change (LEGACY excluded from the set; NOT dropped as a sub-line sentinel) |
| AC-SL-010 | REQ-2.5 | MUST | DEBT lifecycle documented; @MX:TODO-vs-@MX:DEBT distinction stated |
| AC-SL-011a | REQ-2.* | MUST | `diff -q` clean: live `.../workflow/mx-tag-protocol.md` vs template mirror (per-file) |
| AC-SL-011b | REQ-2.* | MUST | `diff -q` clean: live `.../core/moai-constitution.md` mirror covers the @MX-list edit (shared with AC-SL-005a target — one mirror, both REQ edits) |
| AC-SL-011c | REQ-2.* | MUST | The skill reference `.../skills/moai/references/mx-tag.md:39` enumerates the tag-type grammar (`tag_type := "NOTE" \| "WARN" \| "ANCHOR" \| "TODO"`); the @MX:DEBT edit MUST add `DEBT` to that grammar AND mirror it. `diff -q` clean live vs template (no N/A escape — this file IS touched because DEBT joins the grammar) |
| AC-SL-012 | (cross) | SHOULD | Irony guard: no new config file / lint rule / hook / JSON ledger introduced |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — REQ-1 ladder is present, ordered, and language-neutral (AC-SL-001, AC-SL-002)

```
GIVEN the file .claude/rules/moai/core/moai-constitution.md
  AND the section "### 4. Enforce Simplicity"
WHEN a reader looks for a before-coding decision aid
THEN an ordered 6-rung ladder appears (YAGNI → stdlib → native platform → installed dep → one line → minimum code)
  AND each rung is numbered and ordered cheapest-capability-first
  AND no rung names a single language's stdlib, package manager, or platform feature
       (verify: the ladder prose contains none of "npm", "pip", "import React",
        "package.json", "requirements.txt", "go.mod", or any other single-language token)
```

### Scenario 2 — `@MX:DEBT` tag AND its sub-lines scan without error (AC-SL-009, AC-SL-009b)

```
GIVEN a source file containing:
        // @MX:DEBT: in-memory map cache, no eviction
        // @MX:CEILING: < 10k entries
        // @MX:UPGRADE: switch to LRU when entry count exceeds 10k
WHEN the internal/mx scanner parses the file
THEN the @MX:DEBT tag is recognized as a valid TagKind
  AND the scanner does NOT return the error "unknown tag kind: DEBT"   (AC-SL-009)
  AND the @MX:CEILING and @MX:UPGRADE lines do NOT produce a
       "unknown tag kind: CEILING" / "unknown tag kind: UPGRADE" error  (AC-SL-009b)
       (i.e. the scanner's GetErrors() is empty for this fixture —
        the recognized-sub-line-kind set in parseTag skips them, not errors)
  AND moai mx query --kind DEBT returns exactly one tag for this file
```

### Scenario 3 — no-trigger DEBT carries `rotRisk: "no-trigger"` in JSON (AC-SL-008)

```
GIVEN a source file containing a @MX:DEBT marker WITH @MX:CEILING but WITHOUT @MX:UPGRADE
WHEN `moai mx query --kind DEBT --json` is run
THEN the JSON object for that marker contains the exact field+value:  "rotRisk": "no-trigger"
  AND a second @MX:DEBT marker WITH an @MX:UPGRADE sub-line carries "rotRisk": ""
       (empty / absent under omitempty — NOT the "no-trigger" token)
       (binary PASS: `moai mx query --kind DEBT --json | grep -c '"rotRisk": "no-trigger"'` == 1 for the one-marker fixture)
```

### Scenario 4 — @MX:DEBT non-escalation is asserted as DOCTRINE TEXT, not scanner code (AC-SL-010)

> **D4 framing note**: the `@MX:TODO → @MX:WARN` ">3 iteration" escalation lives in DOCTRINE ONLY (`mx-tag-protocol.md` § Tag Lifecycle Rules) — it is NOT implemented in `internal/mx/` Go code (a grep of `internal/mx/` for `escalat`/`>3` returns 0 matches). This scenario therefore verifies the DOCTRINE TEXT, not a mechanical runtime regression. A future implementer must NOT hunt for non-existent escalation code. The §B Claim-2 verdict (the distinction holds) is unchanged — only the AC framing is corrected.

```
GIVEN the doctrine file .claude/rules/moai/workflow/mx-tag-protocol.md (and its template mirror)
WHEN the @MX:DEBT Tag Lifecycle Rules block is read
THEN the doctrine text explicitly states that @MX:DEBT does NOT auto-escalate to @MX:WARN
       (the TODO→WARN >3-iteration rule is documented as NOT applying to DEBT)
  AND the documented DEBT resolution condition is "the @MX:UPGRADE trigger fires", not "work completed"
  (binary PASS: grep of mx-tag-protocol.md finds the DEBT-non-escalation clause —
   e.g. `grep -i 'DEBT.*not.*escalat\|DEBT.*does not.*WARN' mx-tag-protocol.md` returns ≥1 match.
   This is a documentation assertion; no Go test is implied.)
```

### Scenario 5 — `@MX:LEGACY` is NOT dropped by the sub-line-set change (AC-SL-009c, D-NEW-1 regression guard)

```
GIVEN a source file containing a standalone:
        // @MX:LEGACY: pre-SPEC code, no characterization tests yet
WHEN the internal/mx scanner parses the file (after the recognized-sub-line-kind set is added)
THEN the scanner returns exactly 1 tag of kind MXLegacy ("LEGACY")
  AND the tag is NOT skipped as a sub-line sentinel
       (LEGACY is a real tag kind — tag.go:25 MXLegacy — and is DELIBERATELY EXCLUDED
        from the recognized-sub-line set per spec.md REQ-2.4 to prevent this regression)
  (binary PASS: ScanFile on the fixture yields len(tags) == 1 && tags[0].Kind == MXLegacy.
   This is the canary that the sub-line set does NOT accidentally include LEGACY.)
```

## §D.2 Edge Cases

- **EC-1**: A `@MX:DEBT` marker with neither `@MX:CEILING` nor `@MX:UPGRADE` — harvest flags it rot-risk "no-trigger" (the `@MX:UPGRADE` absence is the rot signal; `@MX:CEILING` absence is a quality note, not the rot gate).
- **EC-2**: `@mx:debt` lowercase in source — the scanner upper-cases the kind (`scanner.go` `strings.ToUpper`), so lowercase is accepted same as existing kinds. No special handling required.
- **EC-3**: Template neutrality — the mirrored ladder must pass the `template-neutrality-check.yaml` CI guard (no `/Users/`, no internal SPEC ID, no dates). The ladder prose carries no forbidden content class.
- **EC-4**: `@MX:DEBT` co-located with `@MX:WARN` on adjacent lines — independent kinds, both scanned; no interaction.

## §D.3 Quality Gate Criteria

- `go test ./internal/mx/... ./internal/cli/...` passes; changed-package coverage ≥ 85%.
- `go test ./internal/template/...` passes (neutrality CI guard green).
- `go test ./...` passes (no cascade regression from the new TagKind / new RotRisk field / the recognized-sub-line-kind set change in `parseTag`). Existing scanner tests (including `TestScanFileWithWarnReason`) MUST still pass — the sub-line-set change must not break the existing WARN+REASON pairing path.
- `golangci-lint run` clean on touched files.
- `moai spec lint .moai/specs/SPEC-SIMPLICITY-LADDER-001/spec.md` reports no MUST-FIX findings (frontmatter schema, OutOfScopeRule satisfied).
- `go vet ./internal/mx/...` clean.
- Per-file mirror parity (AC-SL-005a/005b/011a/011b/011c): each `diff -q <live> <template-mirror>` returns clean (exit 0). AC-SL-011c (`skills/moai/references/mx-tag.md`) IS a touched mirror (DEBT joins its tag-type grammar) — no N/A.
- Regression guard (AC-SL-009c): `// @MX:LEGACY:` standalone still yields 1 `MXLegacy` tag (LEGACY excluded from the recognized-sub-line set).

## §D.4 Definition of Done

- [ ] REQ-1: ladder + safety carve-out in constitution; cross-ref line in karpathy-quickref.
- [ ] REQ-1 mirror per-file diff clean: AC-SL-005a (constitution) + AC-SL-005b (karpathy-quickref).
- [ ] REQ-2: `@MX:DEBT` doctrine in mx-tag-protocol.md + constitution @MX list (with the DEBT-non-escalation clause per Scenario 4).
- [ ] REQ-2 mirror per-file diff clean: AC-SL-011a (mx-tag-protocol) + AC-SL-011b (constitution shared) + AC-SL-011c (skills/references/mx-tag.md — MUST be touched: DEBT added to its tag-type grammar).
- [ ] REQ-2 code: `MXDebt` registered in `tag.go`; recognized-sub-line-kind set `{CEILING, UPGRADE, REASON, SPEC, TEST, PRIORITY}` (LEGACY EXCLUDED) added to `parseTag` so `@MX:CEILING`/`@MX:UPGRADE` scan without error (AC-SL-009b); `MXDebt` accepted in `scanner.go` validity case + queryable in `resolver_query.go` + mapped in `cli/mx_query.go`; `RotRisk` field populated + emitted as `"rotRisk": "no-trigger"` in `--json` (AC-SL-008).
- [ ] D-NEW-1 regression guard (AC-SL-009c): standalone `@MX:LEGACY` still scans to 1 `MXLegacy` tag (not dropped).
- [ ] All Given-When-Then scenarios (1, 2, 3, 4, 5) pass — Scenario 4 verified as a doctrine-text grep (D4); Scenario 5 is the D-NEW-1 LEGACY regression canary.
- [ ] Existing scanner tests unbroken (`TestScanFileWithWarnReason` still passes — the REASON vacuous-test repair is explicitly OUT OF SCOPE per §J, owned by `SPEC-MX-SUBLINE-PARSE-REPAIR-001`).
- [ ] Quality gates (§D.3) green.
- [ ] Irony guard (AC-SL-012): no new config file, no new lint rule, no new hook, no separate JSON ledger (the `RotRisk` field is on the existing `Tag` struct, NOT a new ledger file).
- [ ] Status transitioned draft → in-progress (M1) → implemented → completed (sync) per ownership matrix.

## §E. Lifecycle Audit Signals

(Plan-phase emits §E.1 only; run/sync phases populate §E.2-§E.4 in progress.md.)
