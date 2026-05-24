# Acceptance Criteria — SPEC-V3R6-SPEC-ID-VALIDATION-001

## §A. AC Summary Matrix [iter-2: AC-SIV-006 + AC-SIV-007 added]

| AC | REQ mapping | Type | Binary verification command |
|-----|-------------|------|-----------------------------|
| **AC-SIV-001** | REQ-SIV-001 | Section presence (mirror parity) | `grep -c` × 2 files |
| **AC-SIV-002** | REQ-SIV-002 + REQ-SIV-006 | Regex literal canonical equality | `grep -F` (positive) + `grep -F` (negative) × 2 files |
| **AC-SIV-003** | REQ-SIV-003 | AC sub-ID convention enumerated | `grep -E "AC sub-ID\|NNNa"` × 2 files |
| **AC-SIV-004** | REQ-SIV-004 | Regex decomposition print directive present (D4 wording lock-in) | `grep -E "decomposition\|segment match trace\|→ PASS"` × 2 files |
| **AC-SIV-005** | REQ-SIV-007 + quality | Mirror parity (`diff -q` canonical) + lint clean | `diff -q` (primary) + `go vet` + `golangci-lint`; `TestLateBranchTemplateMirror` (supplementary, see AC-SIV-007 activation) |
| **AC-SIV-006** | REQ-SIV-008 [iter-2 D1] | Frontmatter 9→12 schema substitution | `grep -c "9 required fields"` = 0 AND `grep -c "12 canonical\|12 required fields"` ≥ 1 AND `grep -cE "\b(created_at:\|updated_at:\|labels:)"` = 0 in EACH file (D-NEW-1 inline fix: trailing-colon anchor — rejection-table prose like `created_at → must be created` does NOT match the colon-anchored pattern) |
| **AC-SIV-007** | REQ-SIV-009 [iter-2 D2] | Test allowlist enrollment + active subtest | `grep -c 'manager-spec.md' rule_template_mirror_test.go` ≥ 1 AND `go test -run TestLateBranchTemplateMirror -v` shows `manager-spec.md` PASS subtest |

Total ACs: **7** (AC-SIV-001..007) [iter-2: was 5].
Total REQs covered: **8** (REQ-SIV-001/002/003/004/006/007/008/009). REQ-SIV-005 is Optional MAY without AC coverage per spec.md §C.1 decision rule.

## §B. Detailed Acceptance Criteria

### AC-SIV-001 — "SPEC ID Pre-Write Self-Check Protocol" subsection present in both mirror files

**Mapping**: REQ-SIV-001

**Given** the `manager-spec` agent body files exist at:
- `.claude/agents/core/manager-spec.md` (operational source)
- `internal/template/templates/.claude/agents/core/manager-spec.md` (template mirror)

**When** running:
```bash
grep -c "SPEC ID Pre-Write Self-Check Protocol" .claude/agents/core/manager-spec.md
grep -c "SPEC ID Pre-Write Self-Check Protocol" internal/template/templates/.claude/agents/core/manager-spec.md
```

**Then** both commands return exactly `1` (the new subsection header appears exactly once in each file).

**Status**: TBD (verified during run-phase)

---

### AC-SIV-002 — Canonical regex literal verbatim + legacy regex absent

**Mapping**: REQ-SIV-002 (positive) + REQ-SIV-006 (negative regression check)

**Given** the `manager-spec` agent body files exist as above.

**When** running positive check:
```bash
grep -F '^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$' .claude/agents/core/manager-spec.md
grep -F '^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$' internal/template/templates/.claude/agents/core/manager-spec.md
```

**And** running negative check:
```bash
grep -F '^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$' .claude/agents/core/manager-spec.md
grep -F '^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$' internal/template/templates/.claude/agents/core/manager-spec.md
```

**Then** the positive check returns ≥1 match in EACH file (canonical regex present), AND the negative check returns **0 matches in both files** (legacy single-segment regex removed).

**Edge case**: If the canonical regex appears in multiple contexts (e.g., once in the new self-check section + once in the existing checklist L158 replacement), ≥1 match suffices. The requirement is presence, not exact count.

**Status**: TBD (verified during run-phase)

---

### AC-SIV-003 — AC sub-ID convention clarification present

**Mapping**: REQ-SIV-003

**Given** the `manager-spec` agent body files exist as above.

**When** running:
```bash
grep -E "AC sub-ID|NNNa|sub-criteria" .claude/agents/core/manager-spec.md
grep -E "AC sub-ID|NNNa|sub-criteria" internal/template/templates/.claude/agents/core/manager-spec.md
```

**Then** both commands return ≥1 match (the AC sub-ID convention is enumerated in the agent body), AND the matching text contrasts the SPEC ID digit-only anchor (`\d{3}$`) against the AC sub-ID alphabetic suffix (`NNNa`/`NNNb` form).

**Manual qualitative check**: A human reader of the new subsection should easily understand that `SPEC-V3R6-ID-001` is a valid SPEC ID while `AC-V3R6-001a` is a valid acceptance criterion sub-ID, and that mixing the two conventions is a category error.

**Status**: TBD (verified during run-phase)

---

### AC-SIV-004 — Regex match decomposition print directive present [iter-2 D4 wording lock-in]

**Mapping**: REQ-SIV-004

**Given** the `manager-spec` agent body files exist as above.

**When** reading the new "SPEC ID Pre-Write Self-Check Protocol" subsection:

**Then** the subsection MUST contain a directive instructing the agent to print, in its response body before any SPEC ID Write call, a per-segment decomposition match result. Per REQ-SIV-004 (D4 wording lock-in), the directive MUST use one of the literal strings `decomposition` or `segment match trace` and MUST end the output line with `→ PASS` or `→ FAIL` to enable deterministic grep verification.

Acceptable directive forms (non-exhaustive):

- A worked example: `decomposition: SPEC ✓ | V3R6 ✓ | SPEC ✓ | ID ✓ | VALIDATION ✓ | 001 ✓ → PASS`
- A numbered protocol step using one of the literal markers: "Before Write, print the `segment match trace`: `SPEC | <segment-1> | <segment-2> | ... | <NNN> → PASS|FAIL`"
- A template block showing the expected output format with placeholders, using `decomposition:` prefix and `→ PASS|FAIL` suffix

**Verification command (deterministic grep, D4 wording-locked)**:
```bash
grep -E "decomposition|segment match trace|→ PASS" .claude/agents/core/manager-spec.md
grep -E "decomposition|segment match trace|→ PASS" internal/template/templates/.claude/agents/core/manager-spec.md
```

**Then** both commands return ≥1 match indicating the decomposition print directive is present and uses the locked-in wording.

**Stability rationale (D4)**: The previous iter-1 regex (`decomposition|segment.*PASS|✓.*✓|→ PASS`) accepted any string containing `segment` followed by `PASS` (false positives possible from unrelated prose) and any line with two `✓` characters (fragile under unicode rendering or doc reformatting). The iter-2 regex is locked to the literal phrases `decomposition`, `segment match trace`, or the line-end marker `→ PASS` — all 3 alternatives are explicit instructional artifacts unlikely to appear by accident in surrounding prose.

**Status**: TBD (verified during run-phase)

---

### AC-SIV-005 — Mirror parity invariant + quality gate clean [iter-2 D3 disambiguation]

**Mapping**: REQ-SIV-007 (mirror parity) + cross-cutting quality (go vet + lint)

**Given** the `manager-spec` agent body files have been edited per M1.

**When** running mirror parity check **(canonical, primary)**:
```bash
diff -q .claude/agents/core/manager-spec.md internal/template/templates/.claude/agents/core/manager-spec.md
echo "diff exit code: $?"
```

**And** running quality gate:
```bash
go vet ./...
golangci-lint run --timeout=2m
```

**And** running template mirror test **(confirmatory, supplementary — depends on AC-SIV-007 enrollment)**:
```bash
go test ./internal/template/ -run TestLateBranchTemplateMirror -v
```

**Then**:
- `diff -q` returns empty stdout AND exit code 0 (byte-identical mirror parity — THIS IS THE CANONICAL CHECK)
- `go vet` exits with code 0
- `golangci-lint` reports `0 issues.`
- `TestLateBranchTemplateMirror` shows `manager-spec.md` subtest as PASS (active subtest, no longer "if listed" — AC-SIV-007 enrollment guarantees activation)

**D3 disambiguation rationale**: Iter-1 listed `diff` and `TestLateBranchTemplateMirror` as parallel checks with a confusing "(if listed)" qualifier that hid the vacuity gap. Iter-2 makes the verification path explicit: `diff -q` is the canonical primary check (format-agnostic, deterministic, depends only on filesystem state); `TestLateBranchTemplateMirror` is the supplementary CI guard whose activation is itself verified by AC-SIV-007 (REQ-SIV-009 enrollment). Both must pass, but `diff -q` is the SSOT signal of mirror parity.

**Status**: TBD (verified during run-phase)

---

### AC-SIV-006 — Frontmatter 9→12 canonical schema substitution [iter-2 D1 NEW]

**Mapping**: REQ-SIV-008

**Given** the `manager-spec` agent body files have been edited per M1 (iter-2 steps 7-15 of plan.md §B.1).

**When** running the 3-condition compound check on EACH file:

```bash
# Condition (a) — legacy "9 required fields" phrase removed
grep -c "9 required fields" .claude/agents/core/manager-spec.md
grep -c "9 required fields" internal/template/templates/.claude/agents/core/manager-spec.md

# Condition (b) — canonical "12 canonical fields" or "12 required fields" present
grep -cE "12 canonical fields|12 required fields" .claude/agents/core/manager-spec.md
grep -cE "12 canonical fields|12 required fields" internal/template/templates/.claude/agents/core/manager-spec.md

# Condition (c) — snake_case alias instructional text removed (D-NEW-1 inline fix: trailing-colon anchor)
# Rationale: the rejection-table prose `created_at → must be created` legitimately mentions
# the alias name but must not match. The trailing-colon anchor `\b<alias>:` matches only
# YAML key declarations (e.g., `created_at: 2026-05-24`), not prose mentions.
grep -cE "\b(created_at:|updated_at:|labels:)" .claude/agents/core/manager-spec.md
grep -cE "\b(created_at:|updated_at:|labels:)" internal/template/templates/.claude/agents/core/manager-spec.md
```

**Then** for EACH of the 2 mirror files:
- Condition (a) result = **0** (the literal phrase "9 required fields" no longer appears anywhere in the file)
- Condition (b) result ≥ **1** (the canonical 12-field phrasing is present at least once)
- Condition (c) result = **0** (no YAML key declarations using `created_at:`, `updated_at:`, or `labels:` remain — the trailing-colon anchor permits rejection-table prose that mentions the alias names for educational purposes, e.g., `created_at → must be created`, since these are not YAML key declarations)

**Manual qualitative check**: A human reader of the rewritten frontmatter schema block should see exactly the 12 canonical field names from `spec-frontmatter-schema.md`: `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`. The rejection table should invert: `created_at → must be created`, `updated_at → must be updated`, `labels → must be tags`.

**Edge case**: If the canonical SSOT changes the field count in the future (e.g., adds a 13th field), this AC must be re-evaluated. As of iter-2 (2026-05-24), the SSOT defines exactly 12 canonical fields.

**Status**: TBD (verified during run-phase)

---

### AC-SIV-007 — Test allowlist enrollment + active subtest fires [iter-2 D2 NEW]

**Mapping**: REQ-SIV-009

**Given** `internal/template/rule_template_mirror_test.go` has been edited per M1 step 16 (add `.claude/agents/core/manager-spec.md` to `lateBranchMirroredPaths` slice).

**When** running enrollment presence check:

```bash
grep -c 'manager-spec.md' internal/template/rule_template_mirror_test.go
```

**Then** the command returns ≥ **1** (the new slice entry is present).

**And when** running the test:

```bash
go test -run TestLateBranchTemplateMirror -v ./internal/template/ 2>&1 | tee /tmp/test-out.txt
grep -E "TestLateBranchTemplateMirror/manager-spec.md.*PASS" /tmp/test-out.txt
```

**Then**:
- The test output MUST contain a subtest line `--- PASS: TestLateBranchTemplateMirror/manager-spec.md` (or equivalent passing indicator)
- The `grep` count = **1** (the subtest fires AND passes — active, not vacuous)

**Vacuity gap closure (D2 rationale)**: Before iter-2, neither `workflowOptMirroredPaths` nor `lateBranchMirroredPaths` enumerated `manager-spec.md`. `TestLateBranchTemplateMirror` therefore never fired any subtest for `manager-spec.md`, so any mirror drift in the manager-spec.md pair was undetected by CI. AC-SIV-005's "if listed" qualifier hid this gap. Iter-2 D2 enrolls `manager-spec.md` in `lateBranchMirroredPaths` (the correct slice — it is a one-off late mirror addition, not a SPEC-V3R5-WORKFLOW-OPT-001 file). After enrollment, `TestLateBranchTemplateMirror/manager-spec.md` is a real subtest that fails on any future byte mismatch between the two mirror files.

**Status**: TBD (verified during run-phase)

## §C. Edge Cases

### §C.1 Edge case: Existing checklist regex literal removal vs subsection addition order

The run-phase agent MUST replace the legacy regex on L158 AND add the canonical regex within the new self-check subsection. The canonical regex MAY appear in two locations (subsection + existing checklist L158 replacement). This is acceptable per AC-SIV-002 wording (≥1 match required, exact count not specified).

### §C.2 Edge case: Whitespace / formatting differences between mirror files

The `diff` command (AC-SIV-005) is intolerant of any byte difference including trailing whitespace, line ending differences (CRLF vs LF), or BOM markers. Run-phase agent MUST ensure exact byte equality. Use `wc -c` on both files to pre-verify before commit.

### §C.3 Edge case: Manager-spec self-validates its own SPEC ID

This SPEC's own SPEC ID is `SPEC-V3R6-SPEC-ID-VALIDATION-001`. When manager-spec writes this very SPEC's frontmatter, the new self-check protocol (REQ-SIV-004) would print:

```
SPEC ✓ | V3R6 ✓ | SPEC ✓ | ID ✓ | VALIDATION ✓ | 001 ✓ → PASS
```

This is a meta-validation — the new self-check successfully recognizes its own SPEC ID as canonical. This SPEC's plan-phase orchestrator (who is writing this very file) performed the L51 self-check verification by hand and confirmed PASS before file creation.

### §C.4 Edge case: REQ-SIV-005 (Optional 5-incident footnote) absence

REQ-SIV-005 is Optional MAY. If the run-phase agent skips adding the footnote, no AC fails. The footnote is a "nice-to-have" enrichment, not a binary requirement.

## §D. Quality Gate Criteria [iter-2: 7 ACs]

All 7 ACs (AC-SIV-001..007) MUST report PASS for SPEC implementation to be considered complete.

| Criterion | Threshold | Verification |
|-----------|-----------|--------------|
| AC pass rate | 7/7 (100%) | All 7 ACs binary PASS |
| Mirror parity (canonical) | byte-identical | `diff -q` empty output, exit code 0 |
| Mirror parity (CI guard, supplementary) | active subtest passes | `TestLateBranchTemplateMirror/manager-spec.md` PASS (no longer vacuous after iter-2 D2 enrollment) |
| Frontmatter schema canonical alignment | 9-field references eliminated, 12-field schema present | AC-SIV-006 3-condition compound check passes in both mirror files |
| Quality gate | 0 vet errors, 0 lint issues | `go vet` + `golangci-lint` clean |
| Lint baseline | NEW issues = 0 | golangci-lint NEW vs pre-existing baseline check |

## §E. Definition of Done [iter-2: 7 ACs + 3 files]

The SPEC is **DONE** when:

- [ ] All 7 ACs (AC-SIV-001..007) report PASS in run-phase deliverable
- [ ] Both `manager-spec.md` files are byte-identical (mirror parity intact, verified by `diff -q` canonical + active `TestLateBranchTemplateMirror/manager-spec.md` subtest)
- [ ] No legacy single-segment regex (`^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) appears in either manager-spec.md file
- [ ] Canonical multi-segment regex (`^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`) verbatim-equals `internal/spec/lint.go:573`
- [ ] [iter-2 D1] Both manager-spec.md files reference 12 canonical frontmatter fields (no "9 required fields" phrase, no `created_at`/`updated_at`/`labels:` instructional text)
- [ ] [iter-2 D1] Rejection table inverted: `created_at → must be created`, `updated_at → must be updated`, `labels → must be tags`
- [ ] [iter-2 D2] `internal/template/rule_template_mirror_test.go` enrolls `.claude/agents/core/manager-spec.md` in `lateBranchMirroredPaths`
- [ ] [iter-2 D2] `TestLateBranchTemplateMirror/manager-spec.md` subtest fires and passes (active, not vacuous)
- [ ] [iter-2 D4] Self-check directive uses literal `decomposition` or `segment match trace` + `→ PASS|FAIL` line-end marker
- [ ] `go vet ./...` exits with code 0
- [ ] `golangci-lint run --timeout=2m` reports `0 issues.`
- [ ] Run-phase commit is atomic across all 3 files (manager-spec.md mirror pair + rule_template_mirror_test.go) and follows Conventional Commits format with `🗿 MoAI` trailer
- [ ] Sync-phase CHANGELOG.md `[Unreleased]` entry added (manager-docs sync responsibility, B12 9th self-test compliance)
- [ ] spec.md frontmatter `status: draft` → `implemented` (sync-phase responsibility)
- [ ] @MX tag delta evaluation per `mx-tag-protocol.md §a` (template-only .md edit dominates, but iter-2 also touches 1 .go test file with +1-3 LOC slice entry — Mx Step C judge during sync per scope: if 0 @MX tags added/removed/updated in the .go change, SKIP per IVB-001 precedent; otherwise EVALUATE-PASS per mx-tag-protocol §a)
