# Acceptance Criteria — SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001

> Given-When-Then scenarios for `moai spec drift` comprehensive false-positive elimination + era/grandfather/terminal alignment + close-subject doctrine. Each AC is run-phase verifiable. Mechanisms ①②③④ each carry ≥1 fixture-backed AC; mechanism ⑤doctrine carries the doctrine-content AC. Real-repo exemplars MUST be re-verified freshly at run-phase (`internal/spec/CLAUDE.md` sibling-PRESERVE discipline) before assertion.

---

## D. AC Matrix

### AC-DLC-001 — combined-scope sibling SPECs transition DRIFT → aligned (mechanism ①)

Verifies REQ-DLC-001, REQ-DLC-002. **Rebound per plan-audit D1**: the fix is a SECONDARY scope-prefix grep fallback (NOT an in-window gate — the combined-scope close is never in the per-SPEC `--grep=<full-specID>` window). The deterministic fixture is the BINDING signal; the live-repo proof confirms the motivating CCSYNC case.

**Deterministic fixture-backed unit scenario — BINDING** (`drift_combined_scope_test.go`)

- **Given** a synthetic git-log fixture (newest-first): `chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)`, then `feat(SPEC-ABC-FOO-001): M1`, then `feat(SPEC-ABC-BAR-001): M1`
- **When** the walker resolves git-implied status for `specID = SPEC-ABC-FOO-001`: the per-`SPEC-ABC-FOO-001` primary walk yields only `implemented` (the feat); the secondary scope-prefix grep (`git log --grep=SPEC-ABC`) finds the close-infix combined-scope commit, the distinguishing-segment cross-check confirms it names `FOO`
- **Then** it returns `completed`
- **AND When** the walker resolves for `specID = SPEC-ABC-BAR-001`
- **Then** it also returns `completed` (the same combined close names `BAR`) — BOTH siblings of ONE combined-scope close resolve to `completed`. This is the binding assertion.

**Live-repo integration scenario — the case the SPEC exists for**

- **Given** the repository at the post-fix state
- **When** `moai spec drift --json` is executed
- **Then** BOTH `SPEC-CCSYNC-CLAUDEMD-001` AND `SPEC-CCSYNC-TOOLCAT-001` are absent from the DRIFT set (`.Records[] | select(.Drifted==true)` does not contain them); the secondary scope-prefix grep (`git log --grep=SPEC-CCSYNC`) locates the combined-scope close commit `cf7d78a9c` (`chore(SPEC-CCSYNC): Mx-phase 4-phase close (CLAUDEMD + TOOLCAT status→completed, §E.5)`), passes the distinguishing-segment cross-check for each sibling, and resolves both to `completed`. (Empirically pre-verified: the close commit is NOT in `git log --grep=SPEC-CCSYNC-CLAUDEMD-001` but IS in `git log --grep=SPEC-CCSYNC`.)

---

### AC-DLC-012 — non-sibling partial-prefix SPEC is NOT falsely mapped (mechanism ① LSGF-001 collision guard)

Verifies REQ-DLC-002. **New per plan-audit D1**: the secondary prefix-grep is the one real collision-risk site; this AC pins that an unrelated SPEC sharing a partial prefix is NOT absorbed by the fallback.

**Fixture-backed unit scenario** (`drift_combined_scope_test.go`)

- **Given** a synthetic git-log fixture (newest-first): `chore(SPEC-ABC): Mx-phase audit-ready signal + 4-phase close (FOO + BAR)`, then `feat(SPEC-ABC-OTHER-002): M1` (under the same `SPEC-ABC` prefix but NOT named in the combined close), and a separate `SPEC-ABCD-001` (prefix merely CONTAINS `ABC` as a substring — different hyphen boundary)
- **When** the walker resolves git-implied status for `specID = SPEC-ABC-OTHER-002`
- **Then** it is NOT resolved to `completed` by the fallback — the distinguishing-segment cross-check fails (the close names only `FOO`/`BAR`, not `OTHER`), so `SPEC-ABC-OTHER-002` keeps its primary-walk status (and DRIFTs if genuinely incomplete — mechanism ⑤ preserved)
- **AND When** the walker resolves for `specID = SPEC-ABCD-001`
- **Then** the `SPEC-ABC` scope-prefix grep does not apply at all (the hyphen-delimited boundary `SPEC-ABC-` does not prefix-match `SPEC-ABCD-001`), so no false sibling mapping occurs.

---

### AC-DLC-002 — stale sync→completed rule corrected to implemented (mechanism ②)

Verifies REQ-DLC-003, REQ-DLC-004.

**Classifier unit scenario** (`transitions_test.go` extension)

- **Given** the post-fix classifier
- **When** the commit subject `sync(SPEC-EXAMPLE-001): lifecycle complete — v0.3.0 implemented` is classified by `ClassifyPRTitle`
- **Then** the resolved status is `implemented` (NOT `completed`)
- **AND When** `sync(SPEC-EXAMPLE-001): … 4-phase close` is classified
- **Then** the resolved status is `completed` (close-infix wins, checked before the sync rule)
- **AND When** `docs(sync): legacy bare prefix` is classified
- **Then** the resolved status is `implemented` (NOT `completed`).

**Real-repo integration scenario**

- **Given** the post-fix state
- **When** `moai spec drift --json` is executed
- **Then** `SPEC-V3R5-STATUSLINE-STDINFIELDS-001` (frontmatter `implemented`, closed via `438e5b214 sync(...): lifecycle complete — v0.3.0 implemented`) is absent from the DRIFT set.

**Anti-regression sub-assertion**

- **Given** a `feat(SPEC-EXAMPLE-001): M1` commit
- **When** classified
- **Then** it still resolves `implemented`; and NO commit subject lacking a close-infix resolves `completed` (REQ-DLC-004 / AP-2).

---

### AC-DLC-003 — terminal-state frontmatter is authoritative (mechanism ③)

Verifies REQ-DLC-005.

**Real-repo integration scenario**

- **Given** SPECs whose frontmatter is a terminal state
- **When** `moai spec drift --json` is executed
- **Then** `SPEC-V3R3-HARNESS-001` (superseded), `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` (superseded), and `SPEC-V3R6-AGENT-MODEL-ROUTING-001` (archived) are absent from the DRIFT set (no drift reported for terminal-state frontmatter).

**Fixture-backed unit scenario** (`drift_era_terminal_test.go`)

- **Given** a SPEC directory fixture with frontmatter `status: superseded` and a git history that the walker would infer as `in-progress`
- **When** `DetectDrift` runs against the fixture
- **Then** the returned `Records[]` CONTAINS a record for this SPEC with `Drifted == false` (D3 record-emission contract — the record is appended with `Drifted: false`, NOT dropped; `report.Count` is not incremented for it)
- **AND** the same holds for `status: archived` and `status: rejected` fixtures.

---

### AC-DLC-004 — grandfathered eras are exempt (mechanism ④)

Verifies REQ-DLC-006 (and the era-subtlety guard AP-3).

**Real-repo integration scenario**

- **Given** the post-fix drift detector
- **When** `moai spec drift --json` is executed
- **Then** grandfathered-era exemplars `SPEC-V3R2-SPC-002` and `SPEC-V3R3-PROJECT-HARNESS-001` are absent from the DRIFT set; `DetectDrift` calls `LoadEraSignalsFromDir` + `ClassifyEra` and skips any SPEC with `Era.EraFinal() == true`.

**Fixture-backed unit scenario** (`drift_era_terminal_test.go`)

- **Given** a SPEC fixture with NO `progress.md` (H-1 → V2.x, `EraFinal == true`) and a deliberate frontmatter↔git mismatch
- **When** `DetectDrift` runs
- **Then** the returned `Records[]` CONTAINS a record for this SPEC with `Drifted == false` (D3 record-emission contract — grandfather-exempt record is appended, NOT dropped)
- **AND Given** a V3R6 fixture (H-4: `progress.md` with §E.2 + §E.5 + both `sync_commit_sha`/`mx_commit_sha`) with a genuine mismatch
- **When** `DetectDrift` runs
- **Then** `Drifted == true` (modern era still detected — exemption does not over-suppress).

**Era-subtlety guard (AP-3)** — REQUIRED sub-scenario

- **Given** a SPEC fixture with frontmatter `created: 2026-04-23` (after `modernEraThreshold` 2026-04-01) AND a `progress.md` lacking any §E.* marker (H-2 → V3R2-R4, grandfathered)
- **When** `DetectDrift` runs
- **Then** the SPEC is classified as a grandfathered era and exempted — the H-1..H-4 progress.md signal chain fires BEFORE the H-5 date tie-breaker, so the late `created` date does NOT promote it to V3R6. (`Drifted == false`.)

---

### AC-DLC-005 — metadata-sweep + backfill skip regression green

Verifies REQ-DLC-007.

- **Given** the existing tests `drift_chore_skip_test.go` (`TestGetGitImpliedStatus_ChoreSkip`, `TestShouldSkipCommitTitle_ChorePattern`, AC-LSCSK-003) and the DRIFT-CONVENTION-ALIGN backfill-skip tests
- **When** `go test ./internal/spec/...` is run after the fix
- **Then** all chore-skip + narrow-backfill-skip regression tests pass with no behavioral change to `chore(spec):` / `chore(specs):` skipping or the SPEC-ID-scoped backfill-skip.

---

### AC-DLC-006 — word-boundary filter unchanged (LSGF-001)

Verifies REQ-DLC-008 (word-boundary part).

- **Given** the LSGF-001 tests `drift_specid_grep_test.go` (`TestGetGitImpliedStatus_SPECIDWordBoundary` 5 sub-cases + `TestGetGitImpliedStatus_HARNESS001Resolution`)
- **When** `go test ./internal/spec/...` is run after the fix
- **Then** the word-boundary tests pass unchanged — `commitMatchesSPECID("plan(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 …", "SPEC-V3R4-HARNESS-001") == false` (substring-collision protection preserved); the combined-scope mapping (AC-DLC-001) does NOT introduce a substring path.

---

### AC-DLC-007 — audit output unchanged

Verifies REQ-DLC-008 (audit part), REQ-DLC-010.

- **Given** `moai spec audit --json` captured to `/tmp/audit-baseline.json` before the fix
- **When** `moai spec audit --json` is run after the fix and compared
- **Then** the output is identical — same `grandfathered` count, same `modern_era_clean` count, same `drift_findings` set. The era-consumption change reuses `era.go` READ-ONLY and does not modify `audit.go`; `era.go` and `audit.go` are byte-identical to HEAD.

---

### AC-DLC-008 — no false-positive on grandfathered sibling SPECs (module PRESERVE)

Verifies REQ-DLC-006, REQ-DLC-002. (Mandatory per `internal/spec/CLAUDE.md` sibling-PRESERVE convention: "no false positive on closed SPECs" subtest.)

- **Given** a sibling-mapping + era-exemption fixture using a grandfathered-era SPEC whose `status` was set BEFORE this detector shipped (use a V3R2/V3R3 grandfathered fixture, OR the ARR-001 / COORD-001 fixtures named in `internal/spec/CLAUDE.md`)
- **When** the post-fix drift detector runs against the fixture
- **Then** the grandfathered sibling SPEC is NOT newly flagged as drift by the combined-scope mapping (M3) or the era/terminal logic (M1) — no regression introduced against pre-existing closed/grandfathered SPECs.

---

### AC-DLC-009 — drift count strictly decreases to genuine-⑤ residual

Verifies REQ-DLC-009.

- **Given** the verified baseline `moai spec drift --json` → 51 drifted records (`moai spec drift --count` reads 53 live at authoring)
- **When** `moai spec drift --count` is executed at the post-fix state
- **Then** the count is **strictly less than the baseline** and approaches the genuine-⑤ residual.
- **Directional expectation (NOT a binding band)**: removing ① (~2: CCSYNC pair), ② (~3: STATUSLINE + 2 GEARS-ALIGN), ③ (~3+: HARNESS-001 + AGENT-FOLDER-SPLIT + AGENT-MODEL-ROUTING), ④ (~14: V3R2 ×6 + V3R3 ×6 + V3R4 ×1 + DESIGN ×1) yields an expected post-fix residual of roughly the genuine-⑤ set (≈ 51 − ~22 ≈ ~29, run-phase finalizes). Only strict decrease is the binding assertion; the per-mechanism named-exemplar ACs (001–004) are the binding success signal, immune to parallel-session count fluctuation.

---

### AC-DLC-010 — full package test suite green (gate)

Verifies all REQ.

- **Given** the post-fix `internal/spec` package
- **When** `go test ./internal/spec/...` is run
- **Then** the suite is green with no skipped or failing tests attributable to this change, and `go test -cover ./internal/spec/...` shows coverage for the modified files at or above the 85% package baseline.

---

### AC-DLC-011 — close-subject doctrine amendment present + mirrored (mechanism ⑤doctrine)

Verifies REQ-DLC-011.

- **Given** the doctrine SSOT files after the fix
- **When** the **tightened oracle (D2)** is run — it requires the PROHIBITION co-located (≤400 chars, either order) with the combined/abbreviated-scope subject, NOT a bare substring that passes on incidental "full SPEC-ID" prose elsewhere:

  ```bash
  for f in .claude/rules/moai/workflow/lifecycle-sync-gate.md \
           .claude/rules/moai/development/spec-frontmatter-schema.md; do
    grep -Pzo '(?s)(combined|abbreviated)[^\n]*scope.{0,400}?(MUST use individual full-ID|prohibited|disallowed)' "$f" \
      || grep -Pzo '(?s)(MUST use individual full-ID|prohibited|disallowed).{0,400}?(combined|abbreviated)[^\n]*scope' "$f" \
      || { echo "FAIL: $f missing co-located prohibition"; exit 1; }
  done
  ```
- **Then** both files contain the amendment where the combined/abbreviated-scope phrase and the prohibition verb (`MUST use individual full-ID` / `prohibited` / `disallowed`) are co-located within the same amendment block — mandating individual full-IDs in close-commit subjects (prohibiting combined/abbreviated scope such as `chore(SPEC-CCSYNC): …` in favor of one full SPEC-ID per close commit). An incidental mention of "full SPEC-ID" anywhere else in the file does NOT satisfy the oracle.
- **AND** if `.claude/rules/moai/workflow/lifecycle-sync-gate.md` and/or `.claude/rules/moai/development/spec-frontmatter-schema.md` are template-mirrored under `internal/template/templates/.claude/rules/...`, the mirror is byte-consistent with the source (the internal SPEC-ID cross-reference generalized per §25 forbidden-class C2 in the mirror) and the mirror-drift test passes; if NOT mirrored, this clause is N/A (recorded in the M4 commit body).

---

## D.1 Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-DLC-001 | MUST-PASS | Primary mechanism ① fix (secondary prefix-grep fallback); deterministic fixture binding |
| AC-DLC-002 | MUST-PASS | Mechanism ② fix; binding 4-phase model correction |
| AC-DLC-003 | MUST-PASS | Mechanism ③ terminal authority |
| AC-DLC-004 | MUST-PASS | Mechanism ④ era alignment; AP-3 subtlety pinned |
| AC-DLC-005 | MUST-PASS (gate) | Regression — predecessor + chore-skip behavior |
| AC-DLC-006 | MUST-PASS (gate) | Regression — LSGF-001 word-boundary |
| AC-DLC-007 | MUST-PASS (gate) | Audit parity — proves era.go/audit.go untouched |
| AC-DLC-008 | MUST-PASS | Module sibling-PRESERVE convention |
| AC-DLC-009 | SHOULD-PASS (directional) | Strict-decrease; named-exemplar ACs are binding |
| AC-DLC-010 | MUST-PASS (gate) | Full suite + coverage |
| AC-DLC-011 | MUST-PASS | ⑤doctrine content (D2 co-located prohibition oracle) + mirror obligation |
| AC-DLC-012 | MUST-PASS | Mechanism ① LSGF-001 collision guard (non-sibling partial-prefix not mapped) |

## D.2 Definition of Done

- [ ] All MUST-PASS AC green (AC-DLC-001..008, 010, 011, 012)
- [ ] AC-DLC-001 BINDING deterministic fixture: BOTH siblings of one combined-scope close resolve `completed` via secondary prefix-grep fallback
- [ ] AC-DLC-012 collision guard: non-sibling partial-prefix SPEC NOT falsely mapped
- [ ] AC-DLC-009 shows strict decrease from baseline 51
- [ ] `moai spec audit --json` byte-identical to pre-fix baseline (AC-DLC-007)
- [ ] `go test ./internal/spec/...` green, coverage ≥ 85% on modified files (AC-DLC-010)
- [ ] LSGF-001 + chore-skip + backfill-skip regression tests green (AC-DLC-005, 006)
- [ ] `era.go` + `audit.go` byte-identical to HEAD (PRESERVE)
- [ ] Doctrine amendment in both SSOT files (+ mirror if mirrored) (AC-DLC-011)
- [ ] Genuine-⑤ residual classified + handed off to follow-up SPEC (informational, M5)
- [ ] No genuine-⑤ drift silently cleared (AP-1 verified — residual still reports real gaps)

## D.3 Edge Cases (run-phase must handle)

1. **Combined-scope with 3+ siblings**: `chore(SPEC-ABC): … 4-phase close (FOO + BAR + BAZ)` — each of `SPEC-ABC-FOO-001`/`SPEC-ABC-BAR-001`/`SPEC-ABC-BAZ-001` independently passes the distinguishing-segment cross-check against the close and resolves `completed`. A `SPEC-ABC-QUX-004` NOT named in the close does NOT resolve via the fallback (gate c). Verify no cross-contamination and no false attribution to unnamed siblings.
2. **Combined-scope prefix that is also a full SPEC prefix**: ensure `SPEC-ABC` (combined) does not map to `SPEC-ABCD-001` (different hyphen boundary). `deriveScopePrefix` + hyphen-delimited `HasPrefix(specID, "SPEC-ABC-")` handles this — the scope-prefix grep does not even apply to `SPEC-ABCD-001` (AC-DLC-012).
2b. **Scope-prefix derivation with multiple distinguishing segments**: `SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001` → `deriveScopePrefix` strips only the trailing `-CONVENTION-001`? NO — the scope-prefix concept applies only when a COMBINED-scope close exists (e.g. a hypothetical `chore(SPEC-V3R6-DRIFT): … close (LEGACY + ALIGN)`). When no combined-scope close exists in the scope-prefix window, the fallback simply finds nothing and the SPEC keeps its primary-walk status (no harm). The derivation strips the final `-<SEGMENT>-<NNN>` pair; run-phase must verify the derivation does not over-strip multi-segment IDs into a too-broad prefix that collides across unrelated SPEC families.
3. **Era boundary near threshold**: `created` exactly `2026-04-01` with sparse progress.md — `ClassifyEra` H-chain decides; the drift detector trusts `ClassifyEra` verbatim (does NOT re-implement date logic).
4. **Terminal state with a later non-terminal commit**: a `superseded` SPEC with a newer unrelated `feat` commit — terminal frontmatter still wins (AC-DLC-003 early-return precedes git inference).
5. **Sync commit carrying BOTH sync-phase and close-infix**: close-infix wins (checked first in `ClassifyPRTitle`); resolves `completed`, not `implemented`.

## D.4 REQ ↔ AC Traceability (mirror of spec.md §3)

| REQ | Covered by AC | Mechanism |
|-----|---------------|-----------|
| REQ-DLC-001 | AC-DLC-001 | ① combined-scope secondary prefix-grep fallback |
| REQ-DLC-002 | AC-DLC-001, AC-DLC-008, AC-DLC-012 | ① word-boundary preservation + collision guard |
| REQ-DLC-003 | AC-DLC-002 | ② stale-rule correction |
| REQ-DLC-004 | AC-DLC-002 | ② close-infix sole completed signal |
| REQ-DLC-005 | AC-DLC-003 | ③ terminal authority (+ D3 record-emission) |
| REQ-DLC-006 | AC-DLC-004, AC-DLC-008 | ④ era/grandfather alignment (+ D3 record-emission) |
| REQ-DLC-007 | AC-DLC-005 | regression: chore + backfill skip |
| REQ-DLC-008 | AC-DLC-006, AC-DLC-007 | regression: word-boundary + audit |
| REQ-DLC-009 | AC-DLC-009 | count strict-decrease |
| REQ-DLC-010 | AC-DLC-007 | audit parity |
| REQ-DLC-011 | AC-DLC-011 | ⑤doctrine + mirror (D2 oracle) |

Every REQ → ≥1 AC; every AC → ≥1 REQ. (11 REQ × 12 AC.) Mechanisms ①②③④ each have ≥1 fixture-backed AC (001/002/003/004); mechanism ① additionally carries AC-DLC-012 (LSGF-001 collision guard for the secondary prefix-grep). ⑤doctrine has the doctrine-content AC (011). The no-false-positive-on-grandfathered-sibling subtest is AC-DLC-008.
