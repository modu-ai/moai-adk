# Acceptance Criteria — SPEC-MODEL-MATRIX-DOCS-001

21 acceptance criteria across two milestones. Every criterion states observable evidence; no criterion is satisfied by assertion alone.

**Verification status inherited from the split**: zero of these criteria have been formally verified. M6 landed under the original SPEC ID with toolchain-level verification only (build / test / lint / docs-build / matrix cross-check), never an AC-level PASS/FAIL matrix. M7 has not started.

---

## §D AC Matrix

### §D.1 M6 — 4-locale documentation

**AC-MPMD-001** (REQ-MPMD-001)
- **Given** the S0 verification record
- **When** the README benchmark table is read in each of the 4 locale files
- **Then** every figure matches the record and each row's effort level is labelled
- **Verify**: manual comparison of all 4 tables against the record; every number traced
- **Current state — UNMET**: the record does not exist, so no figure can be traced to it. The published figures trace to an unverified reading.

**AC-MPMD-002** (REQ-MPMD-002)
- **Given** `advanced/profile-matrix.md` in each of the 4 locales
- **When** the matrix section is read
- **Then** it presents 11 agent rows × 3 profile columns matching the predecessor's authoritative table, and contains no agent-group table
- **Verify**: manual read of all 4; row count 11 in each; no group-name heading present

**AC-MPMD-003** (REQ-MPMD-003)
- **Given** `advanced/no-haiku-3tier.md` in each of the 4 locales
- **When** the benchmark table is read
- **Then** every figure matches the S0 record
- **Verify**: manual comparison of all 4 tables against the record
- **Current state — UNMET**: same reason as AC-MPMD-001.

**AC-MPMD-004** (REQ-MPMD-004)
- **Given** `multi-llm/model-policy.md` in each of the 4 locales
- **When** it is searched for the retired policy claim
- **Then** no locale asserts that every worker agent is pinned to Sonnet 5
- **Verify**: manual read of all 4 for the claim in each language

**AC-MPMD-005** (REQ-MPMD-005)
- **Given** the Chinese copy of `multi-llm/model-policy.md`
- **When** it is searched
- **Then** it contains no Haiku column header and no per-agent haiku assignment
- **Verify**:
  ```bash
  grep -in "haiku" docs-site/content/zh/multi-llm/model-policy.md
  # expect: no matches
  ```

**AC-MPMD-006** (REQ-MPMD-006)
- **Given** the 4 locale copies of `multi-llm/model-policy.md`
- **When** their structure is compared
- **Then** all four share the same heading sequence and the same table count
- **Verify**: per-locale heading extraction; sequences identical

**AC-MPMD-007** (REQ-MPMD-007)
- **Given** `.claude/rules/moai/development/model-policy.md`
- **When** it is read end to end
- **Then** it contains no claim that all workers are Sonnet-fixed and no reference to `model_routing_profiles` as a config SSOT
- **Verify**:
  ```bash
  grep -n "model_routing_profiles" .claude/rules/moai/development/model-policy.md
  # expect: no matches
  ```
  plus a manual read confirming the tier table is gone

**AC-MPMD-008** (REQ-MPMD-008)
- **Given** the rule file and its template twin
- **When** they are compared after the edit
- **Then** they are byte-identical
- **Verify**:
  ```bash
  go test ./internal/template/ -run TestRuleTemplateMirror
  ```

**AC-MPMD-009** (REQ-MPMD-009)
- **Given** `agent-authoring.md` and `dynamic-workflows.md`
- **When** their model enums are read
- **Then** both include `fable`
- **Verify**: manual read of the enum line in each file

**AC-MPMD-010** (REQ-MPMD-010)
- **Given** the documentation tree
- **When** it is searched for the retired flag name
- **Then** no user-facing surface instructs the reader to use `--model-policy`
- **Verify**:
  ```bash
  grep -rn -- "--model-policy" docs-site/content README.md README.ko.md README.ja.md README.zh.md .claude/rules/
  # expect: no matches, or only historical-note context
  ```

**AC-MPMD-011** (REQ-MPMD-011)
- **Given** `advanced/profile-matrix.md` in each of the 4 locales
- **When** the naming section is read
- **Then** each states that the profile names denote subscription-tier access rather than performance grade, and that the Max profile is both cheaper and higher-scoring than the Medium profile under the S0-confirmed data
- **Verify**: manual read of all 4; both statements present in each

**AC-MPMD-012** (REQ-MPMD-011, correction clause)
- **Given** the S0 record has landed and carries a non-empty delta against the figures already published
- **When** each affected published page is re-read
- **Then** every differing figure has been corrected in place, in all four locales, and no page still carries the superseded value
- **Verify**: for each non-empty delta row, a manual read of all 4 locale copies of every page carrying that figure; a delta row with no corresponding page edit fails this criterion
- **Note**: this criterion is vacuously satisfiable while the S0 record does not exist. It must be evaluated **after** S0 lands, not before — an early evaluation reports PASS on an empty delta set and closes nothing.

**AC-MPMD-013** (REQ-MPMD-012)
- **Given** `advanced/profile-matrix.md` in each locale
- **When** the Max column's Opus assignments are described
- **Then** each states the assignment is a deliberate quality-first choice for high-failure-cost work and not a benchmark optimum
- **Verify**: manual read of all 4

**AC-MPMD-014** (REQ-MPMD-013)
- **Given** every documentation surface edited in M6
- **When** the change set is inspected
- **Then** each edited surface has all four locale copies in the same change
- **Verify**: `git diff --name-only` grouped by surface; every surface shows 4 locale paths

**AC-MPMD-015** (REQ-MPMD-014, 015)
- **Given** the profile documentation
- **When** the environment-constraint section is read
- **Then** it names the expected fallback for an environment without Fable access (the Max profile draws on Fable for six of its eleven cells), and records that under GLM the increased Fable usage collapses to `glm-5.2`, that this model has the highest observed step count of the three, and that CG-mode profile pairings warrant review on that basis
- **Verify**: manual read; the fallback is **named**, not merely acknowledged, and all three GLM statements are present

### §D.2 M7 — Guard realignment and verification

**AC-MPMD-016** (REQ-MPMD-016, 017)
- **Given** the haiku-residual rule
- **When** a `haiku` value is reintroduced into the web-console model option set or the v4manifest tier-suggestion table
- **Then** the rule emits a finding for each
- **Verify**: Go test seeding each surface with a haiku value in a temp tree and asserting a finding per surface. The seeded-failure direction is required: a test asserting only that a clean tree produces no finding passes identically when the rule scans nothing

**AC-MPMD-017** (REQ-MPMD-018)
- **Given** the haiku-residual rule after M7
- **When** its surface list is read
- **Then** it no longer scans `model_routing_profiles` or `validRoutingModels`
- **Verify**:
  ```bash
  grep -n "model_routing_profiles\|validRoutingModels" internal/spec/lint_haiku_residual.go
  # expect: no matches
  ```

**AC-MPMD-018** (REQ-MPMD-019)
- **Given** the property test
- **When** it runs
- **Then** it asserts all seven matrix invariants and passes
- **Verify**: `go test ./internal/template/ -run TestProfileMatrixInvariants -v` shows all seven sub-assertions. A run reporting a pass with fewer than seven named sub-assertions fails this criterion

**AC-MPMD-019** (REQ-MPMD-020)
- **Given** the repository after M7
- **When** the full suite runs
- **Then** it passes with no skips introduced by this SPEC
- **Verify**:
  ```bash
  go vet ./... && go test ./...
  ```

**AC-MPMD-020** (REQ-MPMD-020)
- **Given** the repository after M7
- **When** the linter runs
- **Then** it reports no new findings relative to the pre-work baseline
- **Verify**:
  ```bash
  golangci-lint run
  ```

**AC-MPMD-021** (REQ-MPMD-021)
- **Given** the release target platforms
- **When** the build runs for each
- **Then** each succeeds
- **Verify**: cross-platform build command per the project's release configuration

---

## §E Severity classification

| Severity | ACs | Meaning |
|---|---|---|
| **Blocking** | AC-MPMD-001, 003, 012 | Depend on the predecessor's undischarged S0; a violation means unverified figures stay live for users |
| **Must-pass (user-facing)** | AC-MPMD-002, 004 … 011, 013 … 015 | A violation ships misleading information to users |
| **Must-pass** | AC-MPMD-016 … 021 | Guard coverage and no-regression; a violation is a defect |

AC-MPMD-012's Blocking placement is the split's addition. In the original SPEC the blocking tier meant "M6 cannot start"; here it means "the live pages stay wrong", because M6 already started and finished.

---

## §F Verification-method distribution

| Method | Count | ACs |
|---|---|---|
| Manual read (copy quality) | 11 | AC-MPMD-001 … 004, 006, 009, 011 … 013, 015, plus the read half of 007 |
| Shell command (absence/presence) | 4 | AC-MPMD-005, 007, 010, 017 |
| Go test (behavioral) | 3 | AC-MPMD-008, 016, 018 |
| Diff inspection | 1 | AC-MPMD-014 |
| Build / toolchain | 3 | AC-MPMD-019, 020, 021 |

The manual-read cluster is concentrated in M6 **by design**: no grep can verify that a paragraph *correctly explains* the naming inversion. Those criteria are satisfied by a reviewer reading the text against the canonical statements in `design.md` §A, not by a token match.

---

## §G Indirect-verification notes

Four criteria cannot be verified by direct observation of the target behaviour and use a proxy:

- **AC-MPMD-010** proxies "no user-facing surface instructs the reader to use the retired flag" through a grep whose expected result admits "historical-note context". A reviewer must distinguish an instruction from a note; the grep cannot.
- **AC-MPMD-012** cannot be evaluated until the predecessor's S0 record exists, and is **vacuously satisfiable** before then — an empty delta set produces a PASS that closes nothing. Its evaluation order is load-bearing.
- **AC-MPMD-016** would pass identically on a rule that scans nothing if it asserted only the clean-tree direction; the seeded-failure direction is therefore mandatory rather than a refinement.
- **AC-MPMD-018** would pass on a property test asserting fewer than seven invariants if the count were not itself asserted; naming the sub-assertions is part of the criterion.

Each is recorded here rather than silently accepted so a reviewer can weigh the residual gap.

**Residual that no criterion closes**: M6 shipped before S0 was discharged, so live documentation carried unverified benchmark figures for the whole interval between the landing and the eventual correction. No acceptance criterion can close that window retroactively; it is recorded as accepted history.

---

## §H Closure gates

This SPEC may close when:

1. All Blocking and Must-pass criteria pass with cited evidence.
2. The predecessor `SPEC-MODEL-MATRIX-CORE-001`'s S0 delta table exists, every published benchmark figure traces to it, and AC-MPMD-012 has been evaluated **after** the record landed.
3. The infographic-disposition clarification in `research.md` §F is resolved and the decision recorded.
4. The four carried-forward unverified inputs owned here — the ja/ko `multi-llm/model-policy.md` shape, the `advanced/*` line ranges, the infographic contents, and cross-platform build status — have been **re-measured** rather than inherited.
5. Both predecessor gates on M7 are confirmed: `SPEC-MODEL-MATRIX-CONFIG-001` deleted the two artifacts whose surfaces are removed, and `SPEC-MODEL-MATRIX-SURFACES-001` cleared the two surfaces that are added.
6. `go vet ./... && go test ./... && golangci-lint run` are green (AC-MPMD-019, 020).

---

## §I Forward-looking checks

Conditions that would indicate this SPEC's work has regressed after closure:

- A documentation surface gains a benchmark figure with no effort label, or one not traceable to the S0 record.
- A locale file gains a section absent from the other three.
- `multi-llm/model-policy.md` regains a per-agent table, re-opening the contradiction §B.3 closed structurally.
- The haiku-residual rule loses either newly-added surface, or regains either retired one.
- The matrix property test drops below seven asserted invariants.
- The naming-inversion disclosure or the Max-Opus rationale is removed from `advanced/profile-matrix.md` in any locale.
