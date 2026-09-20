# Acceptance Criteria — SPEC-MODEL-MATRIX-CORE-001

13 acceptance criteria across one blocking precondition and one milestone. Every criterion states observable evidence; no criterion is satisfied by assertion alone.

**Verification status inherited from the split**: zero of these criteria have been formally verified. M1 landed under the original SPEC ID with toolchain-level verification only (build / test / lint / matrix cross-check), never an AC-level PASS/FAIL matrix.

---

## §D AC Matrix

### §D.0 S0 — Leaderboard verification (BLOCKING)

**AC-MPMC-001** (REQ-MPMC-001, 003)
- **Given** the DeepSWE v1.1 leaderboard is reachable
- **When** the verification step reads the Fable 5, Opus 4.8, and GLM 5.2 rows
- **Then** `progress.md` contains a record naming, per model: the effort level of the row read, per-task cost, Pass@1, output tokens, agent steps, the token-column semantics, and the leaderboard version string and update date
- **Verify**: the record exists in `progress.md` and each of the 3 models has all 7 fields populated with no placeholder value

**AC-MPMC-002** (REQ-MPMC-002)
- **Given** two conflicting GLM 5.2 Pass@1 readings (42% and 45%)
- **When** the verification step resolves them
- **Then** the record states one canonical value, the effort level it belongs to, and whether the source publishes a point estimate, an interval, or both
- **Verify**: the record contains a single GLM 5.2 Pass@1 value and an explicit statement of the source's estimate form

**AC-MPMC-003** (REQ-MPMC-004, 001)
- **Given** the verification record is complete
- **When** it is compared against `spec.md` §A.1
- **Then** every differing metric appears in a delta table, and every differing metric that has already reached a shipped documentation surface is raised to `SPEC-MODEL-MATRIX-DOCS-001` as a correction
- **Verify**: the delta table exists in `progress.md`, and for each non-empty delta row a corresponding correction item is recorded against the DOCS successor

### §D.1 M1 — 33-cell matrix redesign

**AC-MPMC-004** (REQ-MPMC-005)
- **Given** the redesigned matrix
- **When** every profile column is enumerated
- **Then** each of the 3 columns contains exactly 11 agent entries, and every one of the 33 `{model, effort}` pairs equals the corresponding cell in `spec.md` §A.3
- **Verify**: a table-driven Go test enumerating all 33 cells passes

**AC-MPMC-005** (REQ-MPMC-006, 007, 008)
- **Given** the group abstraction is removed
- **When** the package is compiled
- **Then** no group constant, no `agentGroupMembership` map, and no `AgentGroup` accessor exists
- **Verify**:
  ```bash
  grep -rn "GroupSpecAuditors\|GroupDesignHarnessE2E\|agentGroupMembership\|func AgentGroup" --include="*.go" internal/
  # expect: no matches
  go build ./...
  ```
  **Current state — FAILING**: this grep returns matches in this tree (`internal/template/profile_matrix.go:209`, `:464`, consumer at `internal/web/agentfm.go:491`). The disposition is owned by card t1037; the criterion is recorded as unmet rather than relaxed.

**AC-MPMC-006** (REQ-MPMC-009)
- **Given** the profile is `max`, `medium`, or `low`
- **When** the resolver is queried for `Explore`
- **Then** it returns `{sonnet, medium}` with the mapped flag true, in all three cases
- **Verify**: Go test asserting all 3 columns for `Explore`

**AC-MPMC-007** (REQ-MPMC-010, 018)
- **Given** an agent name absent from the matrix
- **When** the resolver is queried
- **Then** it returns the `inherit` sentinel with the mapped flag false, and this assertion lives in a test separate from the `Explore` assertion
- **Verify**: two distinct test functions (or subtests) exist — one asserting `Explore` is mapped, one asserting an arbitrary name is not

**AC-MPMC-008** (REQ-MPMC-011)
- **Given** a config with `agent_overrides["manager-spec"] = {opus, xhigh}` and `profile: max`
- **When** the resolver is queried for `manager-spec` and for `plan-auditor`
- **Then** `manager-spec` returns the override and `plan-auditor` returns its own `max` cell (`fable / low`), unperturbed
- **Verify**: Go test; the chosen second agent's cell must differ from the override so the assertion cannot pass vacuously

**AC-MPMC-009** (REQ-MPMC-011)
- **Given** a config whose `profile` value is not one of `max`/`medium`/`low`
- **When** the resolver is queried for any mapped agent
- **Then** it returns that agent's `medium`-column cell
- **Verify**: Go test with an unknown profile string

**AC-MPMC-010** (REQ-MPMC-012, 013, 014, 015)
- **Given** the full matrix
- **When** every cell is inspected
- **Then** no model is `haiku`, every model is in `{fable, opus, sonnet}`, every effort is in `{low, medium, high, xhigh}`, and no model is `inherit`
- **Verify**: a single property test iterating all 33 cells and asserting all four closed-set conditions

**AC-MPMC-011** (REQ-MPMC-016)
- **Given** the matrix
- **When** `manager-docs`, `manager-git`, and `Explore` are resolved under all three profiles
- **Then** each agent yields an identical `{model, effort}` pair across the three columns
- **Verify**: Go test asserting 3 agents × 3 columns collapse to 3 distinct pairs

**AC-MPMC-012** (REQ-MPMC-017)
- **Given** the amended test suite
- **When** it is searched for retired vocabulary
- **Then** no test references a removed group constant or the `hasGroup` name
- **Verify**:
  ```bash
  grep -rn "Group\(SpecAuditors\|Develop\|Advisor\|Docs\|Git\|DesignHarnessE2E\)\|hasGroup" --include="*_test.go" internal/
  # expect: no matches
  go test ./internal/template/...
  ```

**AC-MPMC-013** (REQ-MPMC-005, design.md §A.3)
- **Given** the agent display order and the matrix key set
- **When** both are enumerated
- **Then** they contain the same 11 names
- **Verify**: Go test comparing `ProfileMatrixAgents()` against the key set of each profile column

---

## §E Severity classification

| Severity | ACs | Meaning |
|---|---|---|
| **Blocking** | AC-MPMC-001 … 003 | `SPEC-MODEL-MATRIX-DOCS-001` cannot close; a violation means unverified figures stay live for users |
| **Must-pass** | AC-MPMC-004 … 013 | Correctness and no-regression; a violation is a defect |

No Should-pass tier exists in this SPEC — every criterion it carries is either blocking or a correctness defect.

---

## §F Verification-method distribution

| Method | Count | ACs |
|---|---|---|
| Go test (behavioral) | 8 | AC-MPMC-004, 006 … 011, 013 |
| Shell command (absence/presence) | 2 | AC-MPMC-005, 012 |
| Record inspection | 3 | AC-MPMC-001, 002, 003 |

The three record-inspection criteria are S0's, and are satisfied by a reviewer reading the verification record against the live leaderboard — no grep can verify that a figure was read from the row it claims.

---

## §G Indirect-verification notes

Two criteria use a proxy rather than direct observation of the target behaviour:

- **AC-MPMC-003** proxies "no unverified figure reaches a user" through delta-table completeness plus correction hand-off. Because the documentation it guards **already shipped**, this criterion cannot be satisfied by prevention; the residual gap is the window during which live pages carried unverified figures, which no criterion can close retroactively.
- **AC-MPMC-005** is currently failing on a tree that records its milestone as landed. Recording it as unmet rather than re-scoping it is deliberate: relaxing the criterion to match the tree would erase the evidence that the milestone is incomplete.

---

## §H Closure gates

This SPEC may close when:

1. All Blocking and Must-pass criteria pass with cited evidence.
2. The `Group`-field clarification in `research.md` §F is resolved and the resolution recorded.
3. The §B.3 divergence in `plan.md` — the landed matrix distribution against `spec.md` §A.3 — is reconciled and the reconciliation recorded.
4. The S0 delta table exists, and every benchmark figure in the repository traces to it.
5. Card t1037's disposition of the unexecuted M1 remainder is recorded, and AC-MPMC-005 is either passing or carried as an explicitly-owned debt.
6. `go vet ./internal/template/... && go test ./internal/template/...` are green.

---

## §I Forward-looking checks

Conditions that would indicate this SPEC's work has regressed after closure:

- A group constant or `agentGroupMembership` reappears in `internal/template`.
- The matrix key set and the display order diverge in count.
- An agent's template frontmatter effort diverges from the matrix `medium` column.
- A documentation surface gains a benchmark figure with no effort label, or one not traceable to the S0 record.
