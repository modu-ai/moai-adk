---
spec_id: SPEC-V3R6-GEARS-MIGRATION-001
created: 2026-05-22
updated: 2026-05-22
---

# Acceptance Criteria — SPEC-V3R6-GEARS-MIGRATION-001

## Traceability Matrix (REQ ↔ AC, 100% coverage)

| REQ | AC | Binary check | Verification owner |
|-----|-----|--------------|-------------------|
| REQ-GM-001 (Ubiquitous: dual-notation acceptance) | AC-GM-001 | Lint passes on EARS + GEARS sample REQs (non-strict) | manager-develop |
| REQ-GM-002 (WHEN: IF/THEN warning emission) | AC-GM-002 | `LegacyEARSKeyword` finding count == IF/THEN REQ count | manager-develop |
| REQ-GM-003 (WHEN: GEARS well-formed acceptance) | AC-GM-003 | Zero findings on GEARS sample SPEC | manager-develop |
| REQ-GM-004 (WHILE: 4-locale migration guide) | AC-GM-004 | 4 locale files contain GEARS section + Hugo PASS | manager-develop |
| REQ-GM-005 (WHERE: paper validation gate) | AC-GM-005 | `.moai/research/gears-paper-validation.md` exists pre-edit | orchestrator |
| REQ-GM-006 (WHEN: future-author warning linkage) | AC-GM-006 (merged with AC-GM-002) | Warning message text contains docs-site URL | manager-develop |
| REQ-GM-007 (Ubiquitous: 88-SPEC preservation) | AC-GM-007 | `git diff .moai/specs/` shows zero changes outside SPEC-V3R6-GEARS-MIGRATION-001/ | orchestrator |
| REQ-GM-008 (WHEN: --strict escalation) | AC-GM-008 | `moai spec lint --strict` exits 1 on IF/THEN SPEC | manager-develop |
| REQ-GM-009 (WHILE: 6-month window) | AC-GM-009 | `internal/spec/lint.go` source comment + 4-locale guide document the window | manager-develop |

Coverage: 9/9 REQs → 9 ACs (AC-GM-002 covers both REQ-GM-002 and REQ-GM-006). 100% traceability.

---

## AC-GM-001 — Legacy EARS REQs Continue to Pass `moai spec lint` (Non-Strict)

**REQ coverage**: REQ-GM-001, REQ-GM-007
**Severity**: BLOCKING

### Given

- A sample SPEC file `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md` containing one REQ for each EARS pattern (Ubiquitous, WHEN/THEN, WHILE/SHALL, WHERE/SHALL, IF/THEN), all well-formed per the legacy lint rule.
- The repo at the head of the branch implementing this SPEC.

### When

- The orchestrator runs `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md`

### Then (binary PASS criteria)

- **Exit code** == `0` (non-strict mode, warnings do not affect exit).
- **`ModalityMalformed` finding count** == `0` (all 5 REQs are well-formed per legacy rule).
- **`LegacyEARSKeyword` finding count** == `1` (one finding for the IF/THEN REQ — see AC-GM-002).
- **Existing 88-SPEC lint baseline** unchanged: `go run ./cmd/moai spec lint .moai/specs/` exit code matches the pre-SPEC baseline.

### Verification command

```bash
# 1. Lint the fixture
go run ./cmd/moai spec lint \
  .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md
# Expected exit: 0

# 2. Capture baseline before any code change
git stash
go run ./cmd/moai spec lint .moai/specs/ ; echo "baseline exit: $?"
git stash pop

# 3. Verify 88-SPEC baseline after change
go run ./cmd/moai spec lint .moai/specs/ ; echo "post-change exit: $?"
# Expected: same as baseline exit
```

---

## AC-GM-002 — IF/THEN REQs Emit `LegacyEARSKeyword` Warning + Docs URL Linkage

**REQ coverage**: REQ-GM-002, REQ-GM-006
**Severity**: BLOCKING

### Given

- The fixture `legacy-ears-sample.md` from AC-GM-001 with exactly one REQ using `IF <condition> THEN the system shall <action>`.

### When

- The orchestrator runs `go run ./cmd/moai spec lint -f json <fixture>` and parses the JSON output.

### Then (binary PASS criteria)

- **Finding count** with `Code == "LegacyEARSKeyword"` == `1`.
- **Finding `Severity`** == `"warning"`.
- **Finding `Message`** contains the substring `"GEARS migration"`.
- **Finding `Message`** contains a URL substring matching `adk.mo.ai.kr` OR `docs-site/content/.*GEARS` OR a `concepts/spec-authoring` path.
- **The IF/THEN REQ is NOT additionally flagged** with `Code == "ModalityMalformed"` (the GEARS warning replaces, not duplicates, the legacy "missing SHALL" check when SHALL is present).

### Verification command

```bash
go run ./cmd/moai spec lint -f json \
  .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md \
  | jq '[.[] | select(.code == "LegacyEARSKeyword")] | length'
# Expected: 1

go run ./cmd/moai spec lint -f json \
  .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md \
  | jq -r '.[] | select(.code == "LegacyEARSKeyword") | .message'
# Expected substring: "GEARS migration"
# Expected substring: a docs-site URL or path
```

---

## AC-GM-003 — GEARS Well-Formed REQs Pass Lint with Zero Findings

**REQ coverage**: REQ-GM-003
**Severity**: BLOCKING

### Given

- A fixture `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/gears-sample.md` with REQs in canonical GEARS notation: Ubiquitous (`The system shall`), Event-driven (`WHEN ... shall`), State-driven (`WHILE ... shall`), Optional-precondition (`WHERE ... shall`). No IF/THEN.

### When

- The orchestrator runs `go run ./cmd/moai spec lint -f json <fixture>`.

### Then (binary PASS criteria)

- **Exit code** == `0`.
- **Total finding count** == `0` (no `ModalityMalformed`, no `LegacyEARSKeyword`, no `MissingExclusions` — fixture includes `### Out of Scope` to satisfy `OutOfScopeRule`).

### Verification command

```bash
go run ./cmd/moai spec lint -f json \
  .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/gears-sample.md \
  | jq 'length'
# Expected: 0
```

---

## AC-GM-004 — 4-Locale docs-site GEARS Migration Guide Present + Hugo PASS

**REQ coverage**: REQ-GM-004
**Severity**: BLOCKING

### Given

- The target docs-site file (resolved via Open Question Q4 at run-phase entry; default: `workflow-commands/moai-plan.md`).
- 4 locale variants: `docs-site/content/{ko,en,ja,zh}/<resolved-path>`.

### When

- The orchestrator runs `grep` on each locale + `hugo --gc --minify` from `docs-site/`.

### Then (binary PASS criteria)

- **Each of the 4 locale files** contains a section header at H2 level whose text translates to "GEARS notation" or equivalent in the target locale (`grep -l '^## .*GEARS' docs-site/content/{ko,en,ja,zh}/<path>` returns 4 paths).
- **Each locale section** contains a 5-pattern table (Ubiquitous / WHEN / WHILE / WHERE / IF→deprecated) — verified by `grep -c '|'` returning >= 7 (1 header row + 5 data rows + 1 spacer minimum).
- **Hugo build** exits 0 with no missing-page errors.
- **Per-locale line count ratio** (max/min) <= 1.20 (translation parity check, mirrors AC-DUD-008 from SPEC-V3R6-DOCS-USER-DRIFT-001 precedent).

### Verification command

```bash
# Replace <PATH> with the run-phase-resolved path (Open Question Q4)
PATH_REL="workflow-commands/moai-plan.md"   # default; may change at run-phase

# Per-locale GEARS H2 presence
for L in ko en ja zh; do
  grep -l '^## .*GEARS\|^## .*GEARS notation' "docs-site/content/$L/$PATH_REL" \
    || echo "MISSING: $L"
done
# Expected: 4 path lines, 0 MISSING

# Hugo build
cd docs-site && hugo --gc --minify ; echo "hugo exit: $?"
# Expected: exit 0

# Line-count parity
for L in ko en ja zh; do wc -l "docs-site/content/$L/$PATH_REL" ; done | \
  awk '{print $1}' | sort -n | awk 'NR==1{min=$1} END{print "ratio:", $1/min}'
# Expected: ratio <= 1.20
```

---

## AC-GM-005 — Run-Phase Web-Research Validation Pre-Gate

**REQ coverage**: REQ-GM-005
**Severity**: BLOCKING (run-phase gate)

### Given

- This SPEC has been merged + run-phase is about to begin.

### When

- The orchestrator transitions from plan-phase to run-phase entry.

### Then (binary PASS criteria — orchestrator-owned, before any code-touch)

- **File exists**: `.moai/research/gears-paper-validation.md`.
- **File content** includes:
  - Canonical GEARS paper URL (verified via WebFetch).
  - Verbatim quote of GEARS keyword definitions from the paper.
  - Side-by-side diff against this SPEC's §1 GEARS keyword table.
  - Explicit verdict: `MATCH` (proceed) OR `MISMATCH` (re-open SPEC for amendment).
- **If verdict == MISMATCH**: a new commit on plan-branch amends spec.md §1 + plan.md M2 + acceptance.md AC-GM-003 fixture BEFORE M2 (lint.go edit) executes.

### Verification command

```bash
test -f .moai/research/gears-paper-validation.md && echo "PASS: file exists"
grep -q "Verdict:" .moai/research/gears-paper-validation.md && echo "PASS: verdict present"

VERDICT=$(grep "Verdict:" .moai/research/gears-paper-validation.md | head -1)
echo "$VERDICT"
# Expected: "Verdict: MATCH" OR (if MISMATCH) subsequent SPEC amendment commit exists.
# Note: pattern does NOT anchor to line start to tolerate Markdown bold (`**Verdict:**`)
# rendering in the validation report file. Pattern fix landed in M4 chore (B7-2).
```

---

## AC-GM-006 — Reserved (merged into AC-GM-002 above)

**REQ coverage**: REQ-GM-006 (covered by AC-GM-002).

This AC slot is intentionally vacant; REQ-GM-006's docs-URL-linkage check is performed inside AC-GM-002 to avoid splitting a single grep into two ACs.

---

## AC-GM-007 — 88 Existing SPECs Preserved (Backward-Compat Diff Verification)

**REQ coverage**: REQ-GM-007
**Severity**: BLOCKING

### Given

- The 88 existing SPEC files under `.moai/specs/SPEC-*/`(spec.md|plan.md|acceptance.md) at the head of the plan-merge commit.

### When

- The orchestrator runs `git diff <plan-merge-commit>...HEAD -- .moai/specs/` after run-phase completes.

### Then (binary PASS criteria)

- **Diff output** shows changes ONLY under `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/` (this SPEC's own dir + its fixtures dir).
- **No other SPEC file** under `.moai/specs/SPEC-*/` is modified or deleted.
- **Cross-Wave compatibility check**: if any Wave 2-5 SPEC reached `status: in-progress` after this SPEC's plan-merge, the orchestrator MUST verify (via plan-auditor) that this SPEC's lint.go changes do not invalidate the in-flight SPEC's REQ patterns. If invalidated: re-open this SPEC for amendment.

### Verification command

```bash
# Scope diff
git diff <plan-merge-commit>...HEAD -- .moai/specs/ \
  | grep '^diff --git' \
  | grep -v 'SPEC-V3R6-GEARS-MIGRATION-001' \
  ; echo "above lines should be empty"
# Expected: empty output (only this SPEC's dir modified)

# Cross-Wave check
ls .moai/specs/ | grep -E 'V3R6-(HARNESS|AGENT|RULES|SKILL|SESSION|PROJECT)' \
  | while read spec; do
    STATUS=$(grep '^status:' .moai/specs/$spec/spec.md 2>/dev/null | head -1)
    echo "$spec: $STATUS"
  done
# Manual review: any in-progress SPECs must be cross-checked
```

---

## AC-GM-008 — `--strict` Mode Escalates `LegacyEARSKeyword` to Error

**REQ coverage**: REQ-GM-008
**Severity**: BLOCKING

### Given

- The fixture `legacy-ears-sample.md` from AC-GM-001 (contains one IF/THEN REQ producing `LegacyEARSKeyword` warning in non-strict).

### When

- The orchestrator runs `go run ./cmd/moai spec lint --strict <fixture>`.

### Then (binary PASS criteria)

- **Exit code** == `1` (strict mode escalates warnings to errors).
- **JSON output** still contains the `LegacyEARSKeyword` finding with `Severity: "warning"` (severity field unchanged; only `Report.HasErrors()` exit logic escalates).

### Verification command

```bash
go run ./cmd/moai spec lint --strict \
  .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md
EXIT=$?
echo "strict exit: $EXIT"
# Expected: 1

# Verify severity in JSON is unchanged
go run ./cmd/moai spec lint --strict -f json \
  .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/test-fixtures/legacy-ears-sample.md \
  | jq '.[] | select(.code == "LegacyEARSKeyword") | .severity'
# Expected: "warning"
```

---

## AC-GM-009 — 6-Month Backward-Compatibility Window Documented

**REQ coverage**: REQ-GM-009
**Severity**: BLOCKING

### Given

- The implementation of REQ-GM-009 is policy-only (no behavioral check at runtime).

### When

- The orchestrator inspects `internal/spec/lint.go` source comment + 4-locale migration guide.

### Then (binary PASS criteria)

- **`internal/spec/lint.go`** source contains a comment block (≤ 15 lines) explaining the 6-month window OR until `SPEC-V3R6-GEARS-SWEEP-001` completes, whichever comes first.
- **Each of 4 locale guides** contains a section explaining the window in the locale's language. Translation parity ratio <= 1.20 per AC-GM-004 standard.

### Verification command

```bash
grep -A 15 '6 months\|6-month\|GEARS.*window\|Backward-compat.*window' internal/spec/lint.go
# Expected: non-empty match block.
# Note: pattern accepts both "6 months" (space) and "6-month" (hyphen) spellings
# to tolerate either documentation style. Lint.go uses "6 months" (space) as of
# M2 0bdbae7c2; this AC accepts both forms going forward. Pattern fix landed in
# M4 chore (B7-1).

for L in ko en ja zh; do
  grep -c '6.*month\|6.*개월\|6.*ヶ月\|6.*个月' \
    docs-site/content/$L/workflow-commands/moai-plan.md
done
# Expected: each locale returns >= 1
```

---

## Edge Cases (Out of binary AC scope — manual run-phase verification)

| Edge Case | Handling |
|-----------|----------|
| User SPEC mixes EARS and GEARS in the same file (some REQs IF/THEN, some WHEN) | Lint emits `LegacyEARSKeyword` per IF/THEN REQ; other REQs pass cleanly. No conflict. |
| User SPEC uses GEARS WHERE in the new "precondition / capability gate" sense | Lint accepts (well-formed GEARS) — no special semantic check at lint layer. Semantic correctness is SPEC reviewer's responsibility, not the linter's. |
| User SPEC uses Korean/Japanese/Chinese REQ text (locale-specific authoring) | Lint regex is English-only (existing behavior). Non-English REQ text passes `EARSModalityRule` silently and is not GEARS-checked. Existing behavior preserved. |
| `moai spec lint` invoked on a file that does not exist | Existing `discoverSPECs` error path; no change. |
| Fixture file in `test-fixtures/` accidentally picked up by `go run ./cmd/moai spec lint .moai/specs/` recursive scan | Lint `BaseDir` defaults to `.moai/specs/`; fixtures live under SPEC dir so they ARE scanned. **Mitigation**: fixtures must include valid frontmatter + Out of Scope to not produce spurious findings; OR fixtures live under `_fixtures/` prefix which `discoverSPECs` should skip (verify at M2). |

## Quality Gate Criteria (Tier M Summary)

For Tier M SPECs per `spec-workflow.md` § SPEC Complexity Tier:

- 9/9 REQs traced to ACs (100% coverage) ✓ (see Traceability Matrix)
- ≥ 7 binary ACs (AC-GM-001..009 minus AC-GM-006 vacancy = 8 active ACs) ✓
- Edge cases enumerated (4 above) ✓
- Plan-auditor PASS threshold: 0.80 (Tier M)

## Definition of Done

- All 8 active ACs verified PASS (commands return expected results)
- `go test ./internal/spec/...` exits 0 with new test cases for `LegacyEARSKeyword` covering IF/THEN warning path
- Coverage on `internal/spec/lint.go` is >= 85% (project baseline; new code does not lower)
- Cross-platform PASS: `GOOS=windows GOARCH=amd64 go build ./...` exits 0
- C-HRA-008 boundary grep: `grep -rn 'AskUserQuestion' internal/spec/` returns zero non-test, non-comment matches (lint engine is pure CLI, no user interaction)
- 4-locale Hugo build PASS
- progress.md created with verification command outputs preserved as evidence
