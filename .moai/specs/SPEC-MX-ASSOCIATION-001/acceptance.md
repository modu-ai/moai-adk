# acceptance.md — SPEC-MX-ASSOCIATION-001

> Verification layer. Each AC is a binary-testable Given-When-Then. GEARS obligations live in `spec.md` (REQ-MX-ASSOC-NNN); this file does NOT restate them as requirements.

## §D. AC Matrix

| AC ID | REQ | Subject | Severity | Traceability |
|-------|-----|---------|----------|--------------|
| AC-MX-ASSOC-001 | REQ-001, REQ-002 | `@MX:SPEC:SPEC-X-001` sub-line, no path/body match → `spec_associations` contains `SPEC-X-001` | MUST | plan M1+M2 |
| AC-MX-ASSOC-002 | REQ-005 | Coverage on real repo ≥ 10.2 % AND sub-line source contributes ≥ 40 net-new associated tags | MUST | plan M5 |
| AC-MX-ASSOC-003 | REQ-003 | Sub-line SPEC ID unresolved against known-SPEC set → `UnresolvedSpecRef` warning emitted, ID kept in `spec_associations`, no crash | MUST | plan M3 |
| AC-MX-ASSOC-004 | REQ-002, REQ-004 | Existing path-based + body-based associations unchanged (characterization regression) | MUST | plan M4 |
| AC-MX-ASSOC-005 | REQ-004 | Full production tag set validates green; no existing association removed | MUST | plan M4 |
| AC-MX-ASSOC-006 | REQ-001 | Dangling `@MX:SPEC:` (no preceding tag in file) → `DanglingSpecRef` warning, no crash, no spurious association | MUST | plan M1 |
| AC-MX-ASSOC-007 | REQ-002 | Sub-line ID also present in `Body` → appears once in `spec_associations` (de-dup) | SHOULD | plan M2 |

### §D.1 Severity model

AC-001 + AC-002 + AC-004 + AC-005 are the load-bearing quartet: AC-001 proves the new path works at all, AC-002 proves it moves the real-world needle, AC-004 + AC-005 prove the change is additive-only (no regression). AC-003 binds the validation decision (flag-but-keep) and prevents a stale-ref crash from aborting a 1,567+-tag validation. AC-006 guards malformed input. AC-007 is a SHOULD (de-dup correctness) — failure degrades quality but does not regress correctness.

### §D.2 AC definitions (Given-When-Then)

#### AC-MX-ASSOC-001 — Sub-line drives association with no path/body match (REQ-001, REQ-002)

**Given** a Go source fixture containing exactly one `@MX:NOTE` tag whose `Body` contains no `SPEC-*` token, whose `File` path matches no SPEC `module:` entry, and followed within the same file by a line `// @MX:SPEC: SPEC-FIXTURE-001`.

**When** the scanner scans the fixture and `SpecAssociator.Associate(tag)` runs against the captured tag with a `specModules` map that DOES contain `SPEC-FIXTURE-001`.

**Then** `Associate` returns a slice containing `SPEC-FIXTURE-001`, AND the captured `Tag.SpecRef` field equals `"SPEC-FIXTURE-001"`.

**Test shape:** Go table test in `internal/mx` — construct the fixture in `t.TempDir()`, scan, associate, assert.

#### AC-MX-ASSOC-002 — Real-repo coverage lift (REQ-005)

**Given** the post-change binary built from the worktree, and the real repository tree.

**When** the measurement harness runs `LoadSpecModules(projectRoot)` + `Scanner.ScanDir(projectRoot)` + `SpecAssociator.Associate` over every scanned tag.

**Then** (a) overall coverage ≥ 10.2 %, AND (b) the count of tags whose `spec_associations` set is non-empty AND contains at least one ID attributable ONLY to the sub-line source (i.e. absent when the sub-line source is disabled) is ≥ 40.

**Measurement procedure:**
1. Run the harness with the sub-line source ENABLED → record `coverage_on` and `associated_tags_on`.
2. Run an equivalent harness with the sub-line source DISABLED (toggle via build tag, env var, or a feature-flag argument wired for the test) → record `coverage_off` and `associated_tags_off`.
3. Assert `coverage_on >= 10.2` AND `(associated_tags_on - associated_tags_off) >= 40`.
4. Assert `coverage_off` is within ±0.2 % of the 9.7 % baseline (regression guard for the characterization baseline).

**Test shape:** Go integration test gated to run against the real repo from the module root (NOT `t.TempDir()`); skip if the project root is not the moai-adk-go checkout (the existing test convention for repo-rooted measurements). Record the measured numbers in the test output for audit.

#### AC-MX-ASSOC-003 — Unresolved sub-line SPEC ID: flag-but-keep (REQ-003)

**Given** a fixture tag carrying `// @MX:SPEC: SPEC-DOES-NOT-EXIST-001`, and a `specModules` map whose key set does NOT contain `SPEC-DOES-NOT-EXIST-001`.

**When** the scanner + associator run.

**Then** (a) `Associate` returns a slice containing `SPEC-DOES-NOT-EXIST-001`, AND (b) a warning whose kind is `UnresolvedSpecRef` is present in the scanner/resolver warning channel carrying the unresolved ID and the source `file:line`, AND (c) the process does not panic and the scan completes.

**Test shape:** Go table test; assert both the association slice contents and the warning presence via substring/kind match.

#### AC-MX-ASSOC-004 — Path-based + body-based associations unchanged (REQ-002, REQ-004)

**Given** a characterization fixture set of ≥ 8 tags covering: path-only association, body-only association, path+body overlap, no-association, WARN-with-reason, ANCHOR, DEBT, TODO.

**When** `Associate` runs against each fixture tag pre-change and post-change (same `specModules`).

**Then** for every fixture tag that does NOT carry an `@MX:SPEC` sub-line, the returned `spec_associations` slice is byte-for-byte equal pre- and post-change (same length, same order, same IDs).

**Test shape:** Go characterization test — capture the pre-change outputs as golden values (literal `[]string` literals in the test), assert post-change equality. The golden values are derived by running the CURRENT `Associate` on the fixtures and pasting the output; they are NOT computed at runtime (a runtime self-reference would mask any regression).

#### AC-MX-ASSOC-005 — Production tag set validates green (REQ-004)

**Given** the production source tree.

**When** `Scanner.ScanDir(projectRoot)` runs and the validator's error check is applied.

**Then** `GetErrors()` returns no new errors relative to the pre-change baseline (the pre-change set validates green; the post-change set MUST also validate green), AND no tag that previously had ≥1 association now has zero associations solely due to this change.

**Measurement:** run `go test ./internal/mx/...` and the existing validator test suite; additionally run the coverage harness (AC-002) once with the sub-line source DISABLED and assert the per-tag association count never drops below the baseline for any tag.

#### AC-MX-ASSOC-006 — Dangling @MX:SPEC warning (REQ-001)

**Given** a fixture file whose FIRST `@MX:` line is `// @MX:SPEC: SPEC-X-001` (no preceding standalone tag in the file).

**When** the scanner scans the file.

**Then** a `DanglingSpecRef` warning is emitted naming the `file:line`, AND no `Tag` in the result has `SpecRef == "SPEC-X-001"`, AND no panic.

**Test shape:** Go table test.

#### AC-MX-ASSOC-007 — Sub-line + Body de-dup (REQ-002)

**Given** a fixture tag whose `Body` is `"see SPEC-DUP-001 for context"` and whose following sub-line is `// @MX:SPEC: SPEC-DUP-001`.

**When** `Associate` runs.

**Then** the returned slice contains `SPEC-DUP-001` exactly once.

**Test shape:** Go table test.

### §D.3 Indirect verification

- Sidecar JSON observability: the `specRef` field appears in the sidecar for AC-001's fixture tag (round-trip marshal/unmarshal) — covered by adding a marshal assertion to AC-001's test, confirming REQ-001's observability note.
- `moai mx query --spec SPEC-FIXTURE-001` returns the AC-001 fixture tag — a CLI-level smoke assertion is OPTIONAL (Tier M permits unit-level verification); if added, it lives in `internal/mx` CLI tests, not this AC set.

### §D.4 Closure gates (Definition of Done)

- All MUST ACs (001-006) PASS with verbatim command output cited.
- AC-002's measured numbers (coverage_on, coverage_off, delta) recorded in `progress.md §E.2` at run-phase.
- No new errors in the production validator run.
- `go test ./internal/mx/...` green; `go test -race ./internal/mx/...` green (scanner is line-by-line; low race surface, but the associator is called concurrently from the resolver path).

### §D.5 Forward-looking checks (non-blocking)

- If a future SPEC adds validation to the body-source, AC-003's flag-but-keep contract should be re-evaluated for cross-source consistency. Tracked via spec.md §G "Out of Scope — Validating path-based / body-based associations".
- If `@MX:SPEC` adoption grows beyond ~150 sub-lines, re-baseline the AC-002 floor upward (the ≥ 40 floor is calibrated to today's 88 sub-lines).

### §D.6 Quality gate criteria

- Coverage on `internal/mx` package ≥ 85 % (project minimum), with the new scanner branch and associator loop covered.
- LSP / `golangci-lint` / `go vet` clean on `internal/mx`.
- No `fmt.Errorf` string concatenation in new code (use `%w` wrapping per the repo Go standard).

### §D.7 Severity / traceability summary

Every REQ in spec.md §C traces to at least one MUST AC: REQ-001 → AC-001 + AC-006; REQ-002 → AC-001 + AC-004 + AC-007; REQ-003 → AC-003; REQ-004 → AC-004 + AC-005; REQ-005 → AC-002. No orphan ACs; no orphan REQs.
