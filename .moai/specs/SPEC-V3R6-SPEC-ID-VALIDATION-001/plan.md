# Plan — SPEC-V3R6-SPEC-ID-VALIDATION-001

## §A. Tier S Minimal — Section A-E Variant [iter-2: envelope preserved]

Tier classification: **S (Simple)** per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.

- Scope envelope: ≤300 LOC, ≤5 files
- Estimated change [iter-2 updated]: ~80-130 LOC actual addition across 3 files (manager-spec.md mirror pair + rule_template_mirror_test.go 1-3 LOC enrollment)
- plan-auditor PASS threshold: 0.75 (Tier S)

Precedents in Sprint 2 P4 trio (IVB-001 + SARM-001 + TMC-001) and Sprint 7 entry TMD-001 used identical Tier S minimal Section A-E pattern with 4 SPEC artifacts (spec/plan/acceptance/progress) and achieved 1-pass success. This SPEC follows the same pattern; iter-2 bundle (D1+D2+D3+D4) preserves the Tier S envelope (3 files ≤ 5 cap, ~80-130 LOC ≤ 300 cap).

## §B. Implementation Plan — Single Milestone (M1) [iter-2: scope expanded for D1+D2 bundle]

### §B.1 Milestone M1: manager-spec body edit + frontmatter schema fix + test allowlist enrollment + mirror parity sync

**Goal** [iter-2]: Three coordinated edits in a single atomic milestone:

1. **(Original L51)** Insert "SPEC ID Pre-Write Self-Check Protocol" subsection + fix regex literal to canonical multi-segment form + add AC sub-ID convention clarification + add deterministic-grep decomposition print directive — applied to BOTH manager-spec.md mirror files in lockstep.
2. **(iter-2 D1 bundled)** Replace the embedded 9-field frontmatter schema in manager-spec.md with the canonical 12-field schema; invert the snake_case rejection table to match SSOT; update the Pre-write validation checklist and Verification Checklist; remove `created_at`/`updated_at`/`labels` instructional references — applied to BOTH manager-spec.md mirror files in lockstep.
3. **(iter-2 D2 bundled)** Add `.claude/agents/core/manager-spec.md` to the `lateBranchMirroredPaths` slice in `internal/template/rule_template_mirror_test.go` so `TestLateBranchTemplateMirror/manager-spec.md` becomes a real subtest.

**Files modified (3, all in single atomic commit)** [iter-2 expanded from 2 to 3]:

| File | Role | Bytes (before) | Expected bytes (after) |
|------|------|----------------|------------------------|
| `internal/template/templates/.claude/agents/core/manager-spec.md` | template SSOT | 9158 | ~10500-11200 (≈+1350-2050 bytes: new self-check section + regex fix + 12-field schema substitution + 8-line frontmatter rewrite) |
| `.claude/agents/core/manager-spec.md` | operational mirror | 9158 | ~10500-11200 (byte-identical to template) |
| `internal/template/rule_template_mirror_test.go` | test allowlist | ~3700 (lateBranchMirroredPaths block ~10 lines) | ~3750 (+1 slice entry + 1 comment line ≈ +50 bytes) |

**Operations on the 2 manager-spec.md mirror files** (apply same edit to both, atomically):

1. **Locate** the existing `#### [HARD] SPEC Frontmatter Canonical Schema` subsection (currently around L123-164) within the existing `### Step 4: Create SPEC Documents` section
2. **Insert** a new subsection `#### [HARD] SPEC ID Pre-Write Self-Check Protocol` immediately before the existing schema subsection (so the regex pre-check is the first step of validation, ahead of frontmatter schema validation)
3. **Replace** the legacy single-segment regex literal at the Pre-write validation checklist step 2 (currently `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) with the canonical multi-segment literal `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`
4. **Add** the AC sub-ID convention clarification text inside the new self-check subsection (contrast `AC-XXX-NNNa/b` against SPEC ID `\d{3}$` anchor)
5. **Add** the regex-match decomposition print directive (REQ-SIV-004) as a numbered step inside the new self-check subsection; output MUST use literal `decomposition` or `segment match trace` and end with `→ PASS|FAIL`
6. **Optionally** add the 5-incident reference footnote (REQ-SIV-005 Optional MAY)
7. **[iter-2 D1] Replace** L117 "9 required fields" → "12 canonical fields"
8. **[iter-2 D1] Replace** L125 "ALL 9 required fields" → "ALL 12 canonical fields" + rewrite the YAML schema block at L127-138 to enumerate the 12 canonical fields (`id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`) with correct types and constraints per `spec-frontmatter-schema.md`
9. **[iter-2 D1] Replace** L132-133 `created_at:` / `updated_at:` lines → `created:` / `updated:` (canonical names; remove the "NEVER use `created` (legacy, rejected)" wording entirely)
10. **[iter-2 D1] Replace** L136 `labels: [...]` line → `tags: "tag1, tag2"` (comma-separated string per canonical schema)
11. **[iter-2 D1] Relocate** L137 `issue_number: null` entry — move from Required section to the Optional Fields list below (`issue_number` is Optional per SSOT, not Required)
12. **[iter-2 D1] Invert** L150-152 rejection table: replace `- created → must be created_at` / `- updated → must be updated_at` with `- created_at → must be created` / `- updated_at → must be updated` + add `- labels → must be tags` (matches canonical `spec-frontmatter-schema.md` §Rejected Snake_Case Aliases)
13. **[iter-2 D1] Replace** L157 "All 9 required fields present" → "All 12 canonical fields present"
14. **[iter-2 D1] Remove** L162 `labels` YAML array check (canonical schema uses `tags:` comma-separated string)
15. **[iter-2 D1] Update** L175-176 Verification Checklist: replace `created_at` / `updated_at` references with `created` / `updated`; replace `labels` array check with `tags:` comma-separated string check

**Operations on `internal/template/rule_template_mirror_test.go`** [iter-2 D2]:

16. **[iter-2 D2] Add** `.claude/agents/core/manager-spec.md` as a new entry in the `lateBranchMirroredPaths` slice. Format: one slice string entry + one preceding comment line documenting the SPEC origin (`// SPEC-V3R6-SPEC-ID-VALIDATION-001 — manager-spec.md mirror parity (REQ-SIV-007 + REQ-SIV-009)`).

**Tools used**: Edit tool for sequential per-file precise edits. The 3 files are edited within the same agent invocation for atomic commit. MultiEdit MAY be used on each individual manager-spec.md file for the 9 D1 substitution operations (steps 7-15) to reduce round-trips, but is NOT used across files.

### §B.2 LOC estimate [iter-2 updated for D1+D2 bundle]

| Change type | LOC addition |
|-------------|--------------|
| **Original L51 scope (per file)** |  |
| New subsection header (`#### [HARD] SPEC ID Pre-Write Self-Check Protocol`) | 2 |
| Self-check protocol description (4-step procedure) | ~10-15 |
| Canonical regex literal display + match decomposition example with literal `decomposition:` and `→ PASS` markers | ~10-15 |
| AC sub-ID convention clarification (paired with SPEC ID anchor rule) | ~8-12 |
| Optional 5-incident footnote (REQ-SIV-005) | ~5-10 |
| Regex literal replacement at Pre-write validation step 2 | 1 (net 0 line change, swap only) |
| **iter-2 D1 frontmatter schema fix (per file)** |  |
| L117 "9 required fields" → "12 canonical fields" | 1 (net 0 line change, swap only) |
| L125 "ALL 9 required fields" → "ALL 12 canonical fields" + YAML schema block rewrite (L127-138) | ~12-18 (rewrite 12 fields; was 9, now 12) |
| L132-133 `created_at:` / `updated_at:` → `created:` / `updated:` (canonical names + remove "legacy" wording) | ~2-4 (lines shortened) |
| L136 `labels:` → `tags:` (line content swap) | 1 (net 0) |
| L137 `issue_number:` relocate to Optional Fields section | ~2-3 (move 1 line + add to Optional list) |
| L150-152 rejection table inversion (+1 entry for `labels → tags`) | ~3-4 (3 lines modified + 1 added) |
| L157 "All 9 required fields present" → "All 12 canonical fields present" | 1 (net 0) |
| L162 `labels` YAML array check → remove + add `tags:` comma-separated check | ~1-2 (net 0 to +1) |
| L175-176 Verification Checklist `created_at/updated_at/labels` → `created/updated/tags` | ~2-3 (line content swap) |
| **Subtotal per file (original L51 + D1)** | **~60-100 lines per file** |
| **Total both files (mirror)** | **~120-200 lines (counting both halves)** |
| **iter-2 D2 test allowlist enrollment** |  |
| `internal/template/rule_template_mirror_test.go` — 1 slice entry + 1 comment line | 2 |
| **GRAND TOTAL (3 files)** | **~80-130 lines net code addition (LOC measured as committed insertions, counting both mirror halves)** |

Well within Tier S envelope (≤300 LOC, ≤5 files). Margin: ~170-220 LOC headroom + 2 file headroom.

### §B.3 No new code files required

- **NO** new Go code files
- **NO** new test files (existing `rule_template_mirror_test.go` covers mirror parity invariant; the new logical assertion "manager-spec body contains canonical regex" can be implemented as either an inline grep check during run-phase verification OR a new dedicated Go test — both options described in §B.5 testing strategy)
- **NO** changes to `internal/spec/lint.go` or any package under `internal/spec/`

### §B.4 Out-of-scope (deferred to follow-up SPEC) [iter-2 updated — frontmatter drift now in scope]

Same as `spec.md` §B.2. Explicitly NOT touching:

- Other agent body audits (`manager-develop.md`, `manager-docs.md`, `expert-*.md`)
- `spec-frontmatter-schema.md` regex documentation drift (3-way docs vs lint vs agent-body — separate doc-fix SPEC)
- TEMPLATE-MIRROR-DRIFT-001 family sweep beyond the manager-spec.md addition
- Other unrelated content drift in `manager-spec.md` (Step 6, Status Matrix, etc.)

### §B.5 Testing strategy

**Existing tests (no new code, asserted to remain GREEN)**:

1. `internal/template/rule_template_mirror_test.go` — template-mirror invariant test. After M1, BOTH files must be byte-identical. AC-SIV-005 quality gate validation includes this test passing.
2. `internal/spec/lint.go` `FrontmatterSchemaRule` — existing lint enforcer. SPEC-V3R6-SPEC-ID-VALIDATION-001's own frontmatter must pass this rule (self-validation).

**Inline grep verification (run-phase deliverable, manager-develop reports)** [iter-2 V10/V11/V12 added]:

| Verification | Command | Expected output |
|--------------|---------|-----------------|
| V1 (REQ-SIV-001) | `grep -c "SPEC ID Pre-Write Self-Check Protocol" .claude/agents/core/manager-spec.md` | exactly `1` |
| V2 (REQ-SIV-001 mirror) | `grep -c "SPEC ID Pre-Write Self-Check Protocol" internal/template/templates/.claude/agents/core/manager-spec.md` | exactly `1` |
| V3 (REQ-SIV-002) | `grep -F '^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$' .claude/agents/core/manager-spec.md` | exactly 1 match |
| V4 (REQ-SIV-002 mirror) | `grep -F '^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$' internal/template/templates/.claude/agents/core/manager-spec.md` | exactly 1 match |
| V5 (REQ-SIV-006) | `grep -F '^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$' .claude/agents/core/manager-spec.md internal/template/templates/.claude/agents/core/manager-spec.md` | **0 matches** (legacy regex removed) |
| V6 (REQ-SIV-007 canonical) | `diff -q .claude/agents/core/manager-spec.md internal/template/templates/.claude/agents/core/manager-spec.md` | empty stdout, exit code 0 (byte-identical) |
| V7 (REQ-SIV-003) | `grep -E "AC sub-ID\|NNNa\|sub-criteria" .claude/agents/core/manager-spec.md` | ≥1 match |
| V7b (REQ-SIV-004 D4 wording lock-in) | `grep -E "decomposition\|segment match trace" .claude/agents/core/manager-spec.md` | ≥1 match |
| V8 (quality gate) | `go vet ./... && golangci-lint run --timeout=2m` | exit 0, 0 issues |
| V9 (REQ-SIV-007 confirmatory) | `go test ./internal/template/ -run TestLateBranchTemplateMirror -v` | PASS for `manager-spec.md` subtest (no "if listed" qualifier — must be active after V11 enrollment) |
| **V10 (REQ-SIV-008 D1 frontmatter)** | `grep -c "9 required fields" .claude/agents/core/manager-spec.md` AND `grep -c "12 canonical fields\|12 required fields" .claude/agents/core/manager-spec.md` AND `grep -cE "(created_at\|updated_at\|labels:)" .claude/agents/core/manager-spec.md` | First grep = `0` AND second grep ≥ `1` AND third grep = `0`, in EACH of the 2 mirror files |
| **V11 (REQ-SIV-009 D2 enrollment)** | `grep -c 'manager-spec.md' internal/template/rule_template_mirror_test.go` | ≥ `1` (allowlist enrollment present) |
| **V12 (REQ-SIV-009 D2 active subtest)** | `go test -run TestLateBranchTemplateMirror -v ./internal/template/ 2>&1 \| grep -c 'manager-spec.md.*PASS'` | ≥ `1` (real subtest fires + passes, not vacuous) |

**Optional dedicated Go test (NOT required for plan-phase scope)**:

A new test `TestManagerSpecRegexAlignsWithLint` could be added in run-phase to assert programmatically that `.claude/agents/core/manager-spec.md` contains the canonical regex literal verbatim. This test would:

```go
// File: internal/template/manager_spec_regex_alignment_test.go (POTENTIAL, optional)
// Reads both manager-spec.md files, greps for the canonical regex literal,
// fails if literal absent OR if legacy regex still present.
```

**Decision**: Inline grep verification (V1-V12) plus REQ-SIV-009 test allowlist enrollment is sufficient for Tier S minimal scope. Adding the additional Go test `TestManagerSpecRegexAlignsWithLint` is OPTIONAL upside (proactive regression guard) — deferred to a follow-up SPEC if desired. Run-phase agent (manager-develop) decides based on time budget; the V11/V12 enrollment via `lateBranchMirroredPaths` already provides the structural CI guard against future mirror drift.

## §C. Technical Approach

### §C.1 Edit strategy

Use `Edit` tool with exact match anchors to:

1. Insert the new subsection header + body BEFORE the existing "Pre-write validation" numbered list anchor
2. Replace the legacy regex literal on L158 with the canonical literal (exact-string replace)
3. Apply identical edits to both mirror files

Sequence within run-phase:
1. `Read` both files to confirm current state matches expected baseline (9158 bytes each)
2. `Edit` template file first (template-first principle per CLAUDE.local.md §2 [HARD] Template-First Rule)
3. `Edit` operational mirror with identical changes
4. Run V1-V9 verification commands and report all results

### §C.2 Atomic commit strategy [iter-2 updated for 3-file bundle]

All 3 files MUST be committed together in a single atomic commit. Commit message format per Conventional Commits:

```
feat(SPEC-V3R6-SPEC-ID-VALIDATION-001): M1 — manager-spec SPEC ID regex self-check + 12-field frontmatter schema fix + test allowlist enrollment

- (L51 원본) Add "SPEC ID Pre-Write Self-Check Protocol" subsection
- Replace legacy single-segment regex (`^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) with
  canonical multi-segment literal (`^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`)
- Add AC sub-ID convention clarification (XXX-NNNa/b vs digit-only \d{3})
- Add regex-match decomposition print directive (literal `decomposition` or
  `segment match trace`, must end with `→ PASS|FAIL`)
- (iter-2 D1 bundled) Replace 9-field schema with canonical 12 fields
  (id, title, version, status, created, updated, author, priority, phase,
  module, lifecycle, tags); invert snake_case rejection table to match SSOT
- (iter-2 D2 bundled) Enroll `.claude/agents/core/manager-spec.md` in
  lateBranchMirroredPaths slice — TestLateBranchTemplateMirror/manager-spec.md
  becomes a real subtest (was vacuously absent)
- Mirror parity verified via `diff -q` (canonical) + TestLateBranchTemplateMirror
  (confirmatory, now active)
- L51 lesson chain remediation: bundled fix of both regex drift + frontmatter
  schema drift (5-incident L32 chain root cause class fully closed)

🗿 MoAI <email@mo.ai.kr>
```

### §C.3 Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Mirror parity break (only one file edited) | Medium | Edit both files within same agent invocation; V6 `diff` check before commit |
| New subsection placement disrupts existing manager-spec workflow logic | Low | Insert subsection BEFORE existing checklist; existing checklist remains intact |
| Regex literal display formatted as code block uses backticks/backslashes that escape incorrectly | Low | Use raw fenced code block in markdown; grep with `-F` (fixed string) for verification |
| Manager-spec agent ignores its own new self-check in future invocations | Medium (advisory rule) | Defense-in-depth: spec-lint catches regression at PR time; orchestrator audits agent compliance during code review |
| 9-vs-12 field schema drift on L125 confuses future readers | Low | Out-of-scope per orchestrator directive; tracked as separate follow-up SPEC |

### §C.4 Dependencies

- **None blocking**: This SPEC does not depend on any other SPEC for completion
- **Implicit dependency**: `internal/spec/lint.go:573` canonical regex must remain unchanged during run-phase (REQ-SIV-002 verbatim equality assertion)

## §D. Decisions & Rationale

### §D.1 Why M1 single milestone

Tier S minimal scope (2 files, ~30-80 LOC) does not warrant decomposition into multiple milestones. A single atomic milestone with both file edits + verification + commit is the natural unit.

### §D.2 Why Edit (not MultiEdit)

MultiEdit applies multiple edits to one file in one operation. Since this SPEC edits TWO different files, MultiEdit is not the right primitive. Sequential `Edit` calls (one per file) with identical content is clearer and safer. Alternative: `Write` overwriting both files is rejected because it loses surrounding context and risks unrelated edits.

### §D.3 Why no new test FILE is required (but existing test file gains 1 enrollment) [iter-2 updated]

Existing infrastructure + 1 allowlist entry covers all REQ assertions:

- REQ-SIV-001 (subsection present) → V1/V2 grep
- REQ-SIV-002 (canonical regex literal) → V3/V4 grep
- REQ-SIV-003 (AC sub-ID convention) → V7 grep
- REQ-SIV-004 (decomposition print directive + literal wording) → V7b grep (D4 wording lock-in)
- REQ-SIV-006 (legacy regex absent) → V5 grep
- REQ-SIV-007 (mirror parity) → V6 `diff -q` (canonical) + V9 `TestLateBranchTemplateMirror` (confirmatory, activated by V11)
- **REQ-SIV-008 (D1 frontmatter schema)** → V10 grep (3-condition compound check)
- **REQ-SIV-009 (D2 test allowlist enrollment)** → V11 grep (enrollment present) + V12 active subtest fires

No new test FILE is created; iter-2 adds 1 slice entry + 1 comment line to the EXISTING `rule_template_mirror_test.go`. This minimal addition activates `TestLateBranchTemplateMirror/manager-spec.md` as a real subtest. Adding a dedicated `TestManagerSpecRegexAlignsWithLint` Go test remains OPTIONAL upside (proactive regression guard for regex literal alignment specifically) but not required for AC-SIV-001..007 coverage.

### §D.4 Why minimum change discipline

The agent body has ~9158 bytes (≈200 lines). Adding 30-80 lines is a ~15-40% growth. To avoid scope creep, the edit MUST be limited to the new self-check section + regex literal swap + AC sub-ID convention paragraph. No drive-by refactor, no formatting passes, no unrelated improvements per Karpathy Behavior #5 (Scope Discipline).

## §E. References

Same as spec.md §E (code references, documentation references, memory references, confusion case).
