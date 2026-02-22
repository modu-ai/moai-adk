# SPEC-MX-001: Acceptance Criteria

| Field       | Value           |
|-------------|-----------------|
| SPEC ID     | SPEC-MX-001     |
| Title       | @MX TAG System Acceptance Criteria |
| Created     | 2026-02-20      |
| Author      | manager-spec    |

---

## 1. Tag Type Acceptance Criteria

### AC-NOTE-001: @MX:NOTE Tag

**Given** an agent is processing a Go source file containing a magic constant `const maxRetries = 42` without explanation
**When** the agent applies the @MX:NOTE trigger conditions
**Then** the agent adds a comment `// @MX:NOTE: [AUTO] Maximum retry count for payment gateway...` above the constant declaration
**And** the NOTE description is self-contained (understandable without referencing any external SPEC)

**Given** a NOTE-tagged function has an associated SPEC-PAY-001 document
**When** the agent applies the @MX:NOTE trigger conditions
**Then** the agent includes `// @MX:SPEC: SPEC-PAY-001` as a sub-line below the NOTE

**Given** a legacy Go function predates any SPEC document
**When** the agent is in DDD ANALYZE phase
**Then** the agent adds `// @MX:LEGACY: true` as a sub-line below the NOTE

### AC-WARN-001: @MX:WARN Tag

**Given** an agent detects a goroutine launched without context.Context in `func processQueue()`
**When** the agent applies the @MX:WARN trigger conditions
**Then** the agent adds:
```go
// @MX:WARN: [AUTO] Goroutine without context -- can leak on parent cancellation
// @MX:REASON: No context.Context propagation; goroutine runs indefinitely if parent exits
```
**And** the REASON sub-line is present and non-empty

**Given** a file already has 5 @MX:WARN tags
**When** the agent detects a 6th danger zone
**Then** the agent does NOT add a 6th WARN tag
**And** the agent applies priority ordering (P1 highest to P5 lowest) to the existing 5

### AC-ANCHOR-001: @MX:ANCHOR Tag

**Given** a function `func GetUser(id string) (*User, error)` is referenced from 7 different files
**When** the agent computes fan_in >= 5 via Grep-based reference counting
**Then** the agent adds:
```go
// @MX:ANCHOR: GetUser(id string) (*User, error) -- primary user lookup
// @MX:REASON: fan_in=7, called from 7 files; signature change breaks all callers
```
**And** the REASON sub-line is present and non-empty

**Given** a file already has 3 @MX:ANCHOR tags
**When** the agent detects a 4th high fan-in function
**Then** the agent demotes the ANCHOR with the lowest fan-in to @MX:NOTE
**And** the agent adds the new ANCHOR (maintaining the 3-per-file limit)

**Given** an ANCHOR-tagged function's fan-in drops to 2
**When** the agent runs ANALYZE phase re-validation
**Then** the agent proposes demotion to @MX:NOTE in the report
**And** the agent does NOT auto-delete the ANCHOR tag

### AC-TODO-001: @MX:TODO Tag

**Given** a public function `func ValidateOrder(o *Order) error` has no corresponding test file
**When** the agent applies the @MX:TODO trigger conditions
**Then** the agent adds:
```go
// @MX:TODO: [AUTO] Missing test coverage for ValidateOrder; create TestValidateOrder with valid/invalid order cases
// @MX:PRIORITY: P2
```

**Given** an @MX:TODO tag has been unresolved for 4 consecutive iterations
**When** the agent checks the TODO lifecycle
**Then** the agent escalates the TODO to @MX:WARN with reason "Unresolved TODO after 4 iterations"

---

## 2. SPEC-ID Optionality Acceptance Criteria

### AC-SPEC-001: Optional SPEC References

**Given** a function was implemented without any associated SPEC document
**When** the agent annotates the function
**Then** the agent omits the `@MX:SPEC:` sub-line entirely
**And** the primary tag description contains sufficient context to understand the annotation without any SPEC reference

**Given** a function is directly traceable to SPEC-AUTH-001
**When** the agent annotates the function
**Then** the agent includes `// @MX:SPEC: SPEC-AUTH-001` as a sub-line

---

## 3. Author Attribution Acceptance Criteria

### AC-AUTHOR-001: AUTO Prefix

**Given** an agent autonomously adds an @MX:WARN tag
**When** the tag is written to the file
**Then** the description starts with `[AUTO]` (e.g., `// @MX:WARN: [AUTO] Goroutine without context...`)

**Given** a developer manually writes an @MX:NOTE tag
**When** the agent reads the existing tag
**Then** the agent does NOT add `[AUTO]` to the existing human-authored tag
**And** the agent preserves the human-authored tag verbatim

---

## 4. TDD Workflow Acceptance Criteria

### AC-TDD-001: RED Phase

**Given** the agent is in TDD RED phase writing a test for `func ParseConfig(path string) (*Config, error)`
**When** fan-in analysis shows fan_in = 6 for ParseConfig
**Then** the agent adds @MX:ANCHOR before writing the failing test
**And** the ANCHOR includes `// @MX:REASON: fan_in=6, called from 6 files`

**Given** an existing @MX:TODO marks `ParseConfig` as needing tests
**When** the agent writes a failing test for ParseConfig
**Then** the TODO tag is identified as the test target
**And** the TODO is removed after the GREEN phase completes

### AC-TDD-002: GREEN Phase

**Given** the agent is in TDD GREEN phase implementing `func ProcessPayment(order *Order) (*Receipt, error)`
**When** the implementation contains a non-obvious business rule (e.g., tax calculation threshold)
**Then** the agent adds `// @MX:NOTE: [AUTO] Tax calculation uses tiered rates...` at the relevant location

**Given** the implementation introduces cyclomatic complexity >= 15
**When** the GREEN phase completes
**Then** the agent adds @MX:WARN with complexity metric in the REASON

### AC-TDD-003: REFACTOR Phase

**Given** the agent refactors a function that has an @MX:ANCHOR tag
**When** the function signature does not change but internals are restructured
**Then** the @MX:ANCHOR tag is preserved unchanged

**Given** the agent refactors and the fan-in of an ANCHOR-tagged function decreases from 5 to 2
**When** the REFACTOR phase completes
**Then** the agent's report includes a proposed demotion from ANCHOR to NOTE
**And** the fan-in recalculation is documented

---

## 5. DDD Workflow Acceptance Criteria

### AC-DDD-001: ANALYZE Phase

**Given** a legacy codebase with zero @MX tags
**When** the agent enters the DDD ANALYZE phase
**Then** the agent executes the 3-Pass scan algorithm
**And** outputs a priority queue of files to annotate

**Given** existing @MX:SPEC references where the SPEC file has been deleted
**When** the ANALYZE phase validates tags
**Then** the agent converts `@MX:SPEC: SPEC-OLD-001` to `@MX:LEGACY: true` and adds `@MX:TODO: Broken SPEC link, verify context`

### AC-DDD-002: PRESERVE Phase

**Given** the agent writes a characterization test for `func CalculateDiscount(amount float64) float64`
**When** the characterization test captures the current behavior
**Then** the agent adds @MX:ANCHOR with `@MX:TEST: TestCalculateDiscount_Characterization`
**And** any existing @MX:TODO on CalculateDiscount is removed (behavior is now tested)

### AC-DDD-003: IMPROVE Phase

**Given** a SPEC document is retroactively created for previously legacy code
**When** the agent processes the file during IMPROVE phase
**Then** the agent updates `@MX:LEGACY: true` to `@MX:SPEC: SPEC-NEW-001`
**And** the transition is logged in the improvement report

---

## 6. 3-Pass Scan Acceptance Criteria

### AC-SCAN-001: Legacy Bootstrap

**Given** a codebase of ~100 Go files with zero @MX tags
**When** the agent runs the 3-Pass fast tagging algorithm

**Then** Pass 1 completes using only Grep operations (no full file reads)
**And** Pass 1 produces a priority queue with P1 (fan_in>=5), P2 (goroutine/complexity), P3 (magic/no-godoc), P4 (no-test)

**Then** Pass 2 performs full Read only on P1-priority files
**And** Pass 2 generates accurate tag descriptions based on business context

**Then** Pass 3 uses one Edit call per file (batch insertion)
**And** no file exceeds anchor_per_file (3) or warn_per_file (5) limits after insertion

---

## 7. Configuration Acceptance Criteria

### AC-CONFIG-001: .moai/config/sections/mx.yaml

**Given** a project has no .moai/config/sections/mx.yaml file
**When** an agent applies @MX tag logic
**Then** the agent uses default values: anchor_per_file=3, warn_per_file=5, fan_in_anchor=5, complexity_warn=15, branch_warn=8, auto_tag=true

**Given** a .moai/config/sections/mx.yaml file exists with `auto_tag: false`
**When** the agent encounters a trigger condition for @MX:NOTE
**Then** the agent does NOT add any tags
**And** the agent still validates and reports on existing tags

**Given** a .moai/config/sections/mx.yaml file exists with `exclude: ["**/vendor/**", "**/*_generated.go"]`
**When** the agent processes `vendor/lib/client.go` or `models_generated.go`
**Then** the agent skips tag insertion, validation, and reporting for those files

---

## 8. Report Acceptance Criteria

### AC-REPORT-001: Tag Change Report

**Given** an agent completed a DDD IMPROVE phase that added 3 tags and removed 1
**When** the phase completes
**Then** the agent generates a report with sections: "Tags Added (3 new)", "Tags Removed (1 removed)", "Tags Updated (0 updated)", "Attention Required (0)"
**And** each entry includes FILE:LINE, tag type, and reason summary

**Given** a function has both @MX:ANCHOR and @MX:TODO coexisting
**When** the report is generated
**Then** the "Attention Required" section lists the coexistence with the function location

---

## 9. Edge Case Acceptance Criteria

### AC-EDGE-001: Over-ANCHOR Prevention

**Given** a file has 3 ANCHOR tags and a new function meets the fan_in >= 5 threshold
**When** the agent attempts to add a 4th ANCHOR
**Then** the ANCHOR with the lowest fan-in count is demoted to @MX:NOTE
**And** the new ANCHOR is added (total remains 3)

### AC-EDGE-002: Over-WARN Prevention

**Given** a file has 5 WARN tags and a new danger is detected
**When** the agent attempts to add a 6th WARN
**Then** the 6th WARN is not added
**And** the agent reports the omitted danger in the report under "Attention Required"

### AC-EDGE-003: Stale ANCHOR Re-validation

**Given** an ANCHOR tag claims `fan_in=8` but current Grep analysis shows fan_in=3
**When** the ANALYZE phase runs
**Then** the agent updates the ANCHOR description and proposes demotion to NOTE in the report

### AC-EDGE-004: ANCHOR Security Exception

**Given** a security vulnerability is discovered in an ANCHOR-tagged function
**When** the agent needs to modify the function
**Then** the agent adds `// @MX:WARN: [AUTO] ANCHOR breach for security patch` above the ANCHOR
**And** the modification proceeds
**And** the report explicitly documents the ANCHOR breach

### AC-EDGE-005: ANCHOR + TODO Coexistence

**Given** a function has both @MX:ANCHOR (high fan-in) and @MX:TODO (missing test)
**When** the agent generates the phase report
**Then** the "Attention Required" section highlights the coexistence

### AC-EDGE-006: Multi-Language Comment Syntax

**Given** a Python file requires an @MX:WARN tag
**When** the agent inserts the tag
**Then** the comment uses `#` prefix: `# @MX:WARN: [AUTO] ...`

**Given** a Go file requires an @MX:NOTE tag
**When** the agent inserts the tag
**Then** the comment uses `//` prefix: `// @MX:NOTE: [AUTO] ...`

### AC-EDGE-007: Auto-Generated File Exclusion

**Given** `.moai/config/sections/mx.yaml` excludes `**/*_generated.go`
**When** the agent processes `internal/template/embedded.go` (a generated file)
**Then** no @MX tags are added, modified, or validated in that file

### AC-EDGE-008: Team Environment File Ownership

**Given** Agent Teams mode is active with file ownership boundaries
**When** team-backend-dev owns `internal/core/**` and encounters a tag trigger in `internal/ui/wizard.go`
**Then** team-backend-dev does NOT modify tags in `internal/ui/wizard.go` (outside ownership)

### AC-EDGE-009: Broken SPEC Link Recovery

**Given** a tag reads `// @MX:SPEC: SPEC-OLD-001` but `.moai/specs/SPEC-OLD-001/` does not exist
**When** the ANALYZE phase validates the tag
**Then** the agent replaces `@MX:SPEC: SPEC-OLD-001` with `@MX:LEGACY: true`
**And** adds `// @MX:TODO: Broken SPEC link (was SPEC-OLD-001), verify context`

### AC-EDGE-010: Stale NOTE After Signature Change

**Given** a function `func GetUser(id string)` has an @MX:NOTE describing the string ID lookup
**When** the function signature changes to `func GetUser(id uuid.UUID)`
**Then** the agent re-reviews the NOTE and updates the description to reflect the UUID-based lookup

---

## 10. Definition of Done

The SPEC-MX-001 implementation is considered DONE when all of the following are true:

- [ ] `mx-tag.md` skill file exists at `.claude/skills/moai/workflows/mx-tag.md` within templates
- [ ] `mx-tag-protocol.md` rule file exists at `.claude/rules/moai/workflow/mx-tag-protocol.md` within templates
- [ ] `.moai/config/sections/mx.yaml` configuration template exists at project root within templates
- [ ] `manager-ddd.md` frontmatter `skills:` list includes `moai-workflow-mx-tag`
- [ ] `manager-tdd.md` frontmatter `skills:` list includes `moai-workflow-mx-tag`
- [ ] All 4 tag types (NOTE, WARN, ANCHOR, TODO) have complete syntax definitions
- [ ] All 4 lifecycle state machines are documented with transition rules
- [ ] TDD phase integration (RED/GREEN/REFACTOR) is fully specified
- [ ] DDD phase integration (ANALYZE/PRESERVE/IMPROVE) is fully specified
- [ ] 3-Pass fast tagging algorithm is documented with all 3 passes
- [ ] .moai/config/sections/mx.yaml defaults are functional (agent uses defaults when file is absent)
- [ ] Report format is fully specified with all 4 sections
- [ ] All 10 edge cases have documented resolution strategies
- [ ] Comment syntax mapping covers all 16+ supported languages
- [ ] `make build` succeeds after adding new template files
- [ ] `go test ./internal/template/...` passes
- [ ] Skill file follows MoAI skill authoring schema (YAML frontmatter validation)
- [ ] Skill file does not exceed 500 lines (or is properly modularized)
