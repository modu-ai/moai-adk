# Acceptance Criteria — SPEC-MODEL-PROFILE-MATRIX-002

64 acceptance criteria across one blocking precondition and seven milestones. Every criterion states observable evidence; no criterion is satisfied by assertion alone.

---

## §D AC Matrix

### §D.0 S0 — Leaderboard verification (BLOCKING)

**AC-MPM2-001** (REQ-MPM2-001, 003)
- **Given** the DeepSWE v1.1 leaderboard is reachable
- **When** the verification step reads the Fable 5, Opus 4.8, and GLM 5.2 rows
- **Then** `progress.md` contains a record naming, per model: the effort level of the row read, per-task cost, Pass@1, output tokens, agent steps, the token-column semantics, and the leaderboard version string and update date
- **Verify**: the record exists in `progress.md` and each of the 3 models has all 7 fields populated with no placeholder value

**AC-MPM2-002** (REQ-MPM2-002)
- **Given** two conflicting GLM 5.2 Pass@1 readings (42% and 45%)
- **When** the verification step resolves them
- **Then** the record states one canonical value, the effort level it belongs to, and whether the source publishes a point estimate, an interval, or both
- **Verify**: the record contains a single GLM 5.2 Pass@1 value and an explicit statement of the source's estimate form

**AC-MPM2-003** (REQ-MPM2-004, 001)
- **Given** the verification record is complete
- **When** it is compared against `spec.md` §A.2
- **Then** every differing metric appears in a delta table, and no documentation surface has been edited with a benchmark figure prior to the record's existence
- **Verify**: `git log` shows no commit touching a benchmark table earlier than the commit adding the record

### §D.1 M1 — 33-cell matrix redesign

**AC-MPM2-010** (REQ-MPM2-010)
- **Given** the redesigned matrix
- **When** every profile column is enumerated
- **Then** each of the 3 columns contains exactly 11 agent entries, and every one of the 33 `{model, effort}` pairs equals the corresponding cell in `spec.md` §A.4
- **Verify**: a table-driven Go test enumerating all 33 cells passes

**AC-MPM2-011** (REQ-MPM2-011, 012, 013)
- **Given** the group abstraction is removed
- **When** the package is compiled
- **Then** no group constant, no `agentGroupMembership` map, and no `AgentGroup` accessor exists
- **Verify**:
  ```bash
  grep -rn "GroupSpecAuditors\|GroupDesignHarnessE2E\|agentGroupMembership\|func AgentGroup" --include="*.go" internal/
  # expect: no matches
  go build ./...
  ```

**AC-MPM2-012** (REQ-MPM2-014)
- **Given** the profile is `max`, `medium`, or `low`
- **When** the resolver is queried for `Explore`
- **Then** it returns `{sonnet, medium}` with the mapped flag true, in all three cases
- **Verify**: Go test asserting all 3 columns for `Explore`

**AC-MPM2-013** (REQ-MPM2-015, 023)
- **Given** an agent name absent from the matrix
- **When** the resolver is queried
- **Then** it returns the `inherit` sentinel with the mapped flag false, and this assertion lives in a test separate from the `Explore` assertion
- **Verify**: two distinct test functions (or subtests) exist — one asserting `Explore` is mapped, one asserting an arbitrary name is not

**AC-MPM2-014** (REQ-MPM2-016)
- **Given** a config with `agent_overrides["manager-spec"] = {opus, xhigh}` and `profile: max`
- **When** the resolver is queried for `manager-spec` and for `plan-auditor`
- **Then** `manager-spec` returns the override and `plan-auditor` returns its own `max` cell (`fable / low`), unperturbed
- **Verify**: Go test; the chosen second agent's cell must differ from the override so the assertion cannot pass vacuously

**AC-MPM2-015** (REQ-MPM2-016)
- **Given** a config whose `profile` value is not one of `max`/`medium`/`low`
- **When** the resolver is queried for any mapped agent
- **Then** it returns that agent's `medium`-column cell
- **Verify**: Go test with an unknown profile string

**AC-MPM2-016** (REQ-MPM2-017, 018, 019, 020)
- **Given** the full matrix
- **When** every cell is inspected
- **Then** no model is `haiku`, every model is in `{fable, opus, sonnet}`, every effort is in `{low, medium, high, xhigh}`, and no model is `inherit`
- **Verify**: a single property test iterating all 33 cells and asserting all four closed-set conditions

**AC-MPM2-017** (REQ-MPM2-021)
- **Given** the matrix
- **When** `manager-docs`, `manager-git`, and `Explore` are resolved under all three profiles
- **Then** each agent yields an identical `{model, effort}` pair across the three columns
- **Verify**: Go test asserting 3 agents × 3 columns collapse to 3 distinct pairs

**AC-MPM2-018** (REQ-MPM2-022)
- **Given** the amended test suite
- **When** it is searched for retired vocabulary
- **Then** no test references a removed group constant or the `hasGroup` name
- **Verify**:
  ```bash
  grep -rn "Group\(SpecAuditors\|Develop\|Advisor\|Docs\|Git\|DesignHarnessE2E\)\|hasGroup" --include="*_test.go" internal/
  # expect: no matches
  go test ./internal/template/...
  ```

**AC-MPM2-019** (REQ-MPM2-010, design.md §A.3)
- **Given** the agent display order and the matrix key set
- **When** both are enumerated
- **Then** they contain the same 11 names
- **Verify**: Go test comparing `ProfileMatrixAgents()` against the key set of each profile column

### §D.2 M2 — Retire the 36-cell axis and the config mirror

**AC-MPM2-020** (REQ-MPM2-030)
- **Given** both `workflow.yaml` copies
- **When** they are searched
- **Then** neither contains a `model_routing_profiles` block nor a comment referencing it
- **Verify**:
  ```bash
  grep -rn "model_routing_profiles" .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml
  # expect: no matches
  ```

**AC-MPM2-021** (REQ-MPM2-031, 032)
- **Given** the config package
- **When** it is compiled and searched
- **Then** `RouteModelFor`, `ModelRoutingProfiles`, and `validRoutingModels` do not exist, and no orphaned test references them
- **Verify**:
  ```bash
  grep -rn "RouteModelFor\|ModelRoutingProfiles\|validRoutingModels" --include="*.go" internal/
  # expect: no matches
  go test ./internal/config/...
  ```

**AC-MPM2-022** (REQ-MPM2-033, 034)
- **Given** both `llm.yaml` copies
- **When** they are searched for a `profiles:` key under `llm:`
- **Then** neither contains one, and the matrix literal exists in exactly two files (the Go constant and its fidelity test)
- **Verify**:
  ```bash
  grep -rn "^\s*profiles:" .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml
  # expect: no matches
  ```
  plus a manual enumeration confirming exactly 2 files carry the 33-cell literal

**AC-MPM2-023** (REQ-MPM2-033, K-5)
- **Given** the template `llm.yaml` before and after M2
- **When** the diff is inspected
- **Then** only the `profiles:` block and its explanatory comment were removed; the `glm.models` block is byte-identical to its pre-M2 state
- **Verify**: `git diff` on the template file shows no change inside `glm.models`

**AC-MPM2-024** (REQ-MPM2-037)
- **Given** a project whose `llm.yaml` still carries a `profiles:` block and whose `workflow.yaml` still carries `model_routing_profiles`
- **When** the config loader runs
- **Then** it succeeds, treating both blocks as inert
- **Verify**: Go test loading a legacy fixture containing both blocks and asserting no error

**AC-MPM2-025** (REQ-MPM2-035, 036, 113)
- **Given** a project whose `llm.yaml` `profiles:` block differs from the shipped default
- **When** `moai update` runs
- **Then** the customization is either migrated into `agent_overrides` or surfaced in a warning naming the affected profile column and cells — and in neither case is it discarded silently
- **Verify**: Go test with a customized fixture asserting that after the update either the override map is populated with the equivalent per-agent entries, or the emitted warning text names the changed cells

**AC-MPM2-026** (REQ-MPM2-034, K-2)
- **Given** the test suite after M2
- **When** it runs
- **Then** the test asserting that a config `profiles` cell overrides the Go default no longer exists, and the suite is green
- **Verify**:
  ```bash
  grep -rn "ConfigProfilesOverrideDefault" --include="*_test.go" internal/
  # expect: no matches
  go test ./internal/template/... ./internal/config/...
  ```

### §D.3 M3 — Effort actualization

**AC-MPM2-030** (REQ-MPM2-040, 048)
- **Given** a deployed project at profile `low` whose `manager-develop.md` carries `effort: xhigh`
- **When** the effort application runs
- **Then** the file's `effort:` becomes `medium` and every other byte of the file is unchanged
- **Verify**: Go test comparing pre/post file bytes; exactly one line differs

**AC-MPM2-031** (REQ-MPM2-041)
- **Given** any deployed agent file with `model: inherit`
- **When** the effort application runs under any profile
- **Then** the `model:` line is unchanged
- **Verify**: Go test asserting the `model:` value is byte-identical pre/post for all 10 files, under all 3 profiles

**AC-MPM2-032** (REQ-MPM2-042)
- **Given** the template tree
- **When** the effort application runs
- **Then** no file under `internal/template/templates/.claude/agents/` is modified
- **Verify**: Go test snapshotting the template agent directory's file hashes pre/post and asserting equality

**AC-MPM2-033** (REQ-MPM2-043)
- **Given** the template tree after M1 step 7
- **When** each template agent's `effort:` is compared to the matrix's `medium` column
- **Then** every value matches
- **Verify**: Go test reading the embedded template agent frontmatter and comparing against the `medium` column

**AC-MPM2-034** (REQ-MPM2-044)
- **Given** a project at profile `max` with agent efforts already applied
- **When** `moai update` deploys templates and completes
- **Then** the deployed agent efforts equal the `max` column, not the template baseline
- **Verify**: integration-style Go test over a temp project: set profile, apply, update, re-read

**AC-MPM2-035** (REQ-MPM2-045)
- **Given** a config with `agent_overrides["manager-spec"] = {opus, xhigh}` and profile `low`
- **When** the effort application runs
- **Then** `manager-spec.md` retains `effort: xhigh` while `plan-auditor.md` becomes the `low` cell's `low`
- **Verify**: Go test asserting both files

**AC-MPM2-036** (REQ-MPM2-046)
- **Given** the four profile-application seams
- **When** each is inspected
- **Then** each invokes the effort application, and at every seam that deploys templates the invocation follows the deploy
- **Verify**: code read of `internal/cli/init.go`, both `internal/cli/update.go` sites, and `internal/web/agentfm.go`, plus AC-MPM2-034's ordering test

**AC-MPM2-037** (REQ-MPM2-047)
- **Given** the matrix contains `Explore`, which has no file on disk
- **When** the effort application runs
- **Then** it completes without error and without creating an `Explore.md`
- **Verify**: Go test asserting no error and no new file

**AC-MPM2-038** (REQ-MPM2-049)
- **Given** the active profile and an agent name
- **When** a workflow script queries the machine-readable route
- **Then** it receives that agent's resolved `model` and `effort` for the active profile
- **Verify**:
  ```bash
  moai model profile --json
  ```
  output parsed for a named agent yields both fields; a Go test asserts the JSON shape

**AC-MPM2-039** (REQ-MPM2-050, 051)
- **Given** the documentation after M3
- **When** the effort-injection section is read
- **Then** it distinguishes the frontmatter channel (Agent tool, no `effort` parameter) from the `opts.effort` channel (Workflow tool), records that SPEC-001 DECISION-001's wording is superseded, and notes that `Explore` has no agent file
- **Verify**: manual read of the section; all three statements present

**AC-MPM2-040** (REQ-MPM2-040, K-3)
- **Given** `internal/web/agentfm.go`
- **When** the `applyPerfTierEdits` comment block is read
- **Then** it no longer claims frontmatter re-application is retired
- **Verify**:
  ```bash
  grep -n "retired" internal/web/agentfm.go
  ```
  no surviving hit asserts the retirement of frontmatter re-application

### §D.4 M4 — Init wizard question

**AC-MPM2-050** (REQ-MPM2-060, 061)
- **Given** the `model_policy` question
- **When** its title, description, and option text are read
- **Then** they use subscription-tier framing and contain no reference to `haiku`
- **Verify**:
  ```bash
  grep -in "haiku" internal/cli/wizard/questions.go
  # expect: no match in the model_policy question
  ```

**AC-MPM2-051** (REQ-MPM2-062)
- **Given** each of the question's options
- **When** the wizard answer is persisted
- **Then** the resulting `llm.profile` is one of `max`, `medium`, `low`
- **Verify**: Go test iterating every option value through the persistence path and asserting the written profile

**AC-MPM2-052** (REQ-MPM2-062, K-7)
- **Given** a legacy stored answer of `high`
- **When** it is normalized
- **Then** it resolves to `max` without error
- **Verify**: Go test on the normalizer

**AC-MPM2-053** (REQ-MPM2-063, 064)
- **Given** `conversation_language` set to `ko`, then `ja`, then `zh`
- **When** the wizard renders the `model_policy` question
- **Then** the title, description, and every option label and description render in that language, with no English fallback
- **Verify**: Go test asserting a translation entry exists for every string of the question in all three locales

**AC-MPM2-054** (REQ-MPM2-065)
- **Given** the option descriptions
- **When** they are read
- **Then** each names the subscription tier it targets and none asserts that a higher profile produces stronger results
- **Verify**: manual read against `design.md` §D.1

### §D.5 M5 — Web console cleanup

**AC-MPM2-060** (REQ-MPM2-070)
- **Given** the web console agent-frontmatter model selector
- **When** its option set is enumerated
- **Then** `haiku` is absent
- **Verify**: Go test asserting the option slice excludes the haiku constant

**AC-MPM2-061** (REQ-MPM2-071)
- **Given** the v4manifest tier-suggestion table
- **When** the lightblue tier is looked up
- **Then** it yields `{sonnet, low}`
- **Verify**: Go test on the suggestion table

**AC-MPM2-062** (REQ-MPM2-072)
- **Given** the four locale files
- **When** `agentfm.tier.desc` is read in each
- **Then** the key exists in all four and its text describes the effort-reapplication behavior M3 restores
- **Verify**: Go test asserting key presence in all 4 locales, plus a manual read of the wording

**AC-MPM2-063** (REQ-MPM2-073)
- **Given** the four locale files
- **When** they are searched for the orphaned key family
- **Then** no `mp.` key remains, and a consumer grep confirms none was in use
- **Verify**:
  ```bash
  grep -rn '"mp\.' internal/web/
  # expect: no matches
  ```

**AC-MPM2-064** (REQ-MPM2-074)
- **Given** the four locale files after M5
- **When** their key sets are compared
- **Then** all four contain the same key set
- **Verify**: Go test computing the 4 key sets and asserting equality

### §D.6 M6 — 4-locale documentation

**AC-MPM2-070** (REQ-MPM2-080)
- **Given** the S0 verification record
- **When** the README benchmark table is read in each of the 4 locale files
- **Then** every figure matches the record and each row's effort level is labelled
- **Verify**: manual comparison of all 4 tables against the record; every number traced

**AC-MPM2-071** (REQ-MPM2-081)
- **Given** `advanced/profile-matrix.md` in each of the 4 locales
- **When** the matrix section is read
- **Then** it presents 11 agent rows × 3 profile columns matching `spec.md` §A.4, and contains no agent-group table
- **Verify**: manual read of all 4; row count 11 in each; no group-name heading present

**AC-MPM2-072** (REQ-MPM2-082)
- **Given** `advanced/no-haiku-3tier.md` in each of the 4 locales
- **When** the benchmark table is read
- **Then** every figure matches the S0 record
- **Verify**: manual comparison of all 4 tables against the record

**AC-MPM2-073** (REQ-MPM2-083)
- **Given** `multi-llm/model-policy.md` in each of the 4 locales
- **When** it is searched for the retired policy claim
- **Then** no locale asserts that every worker agent is pinned to Sonnet 5
- **Verify**: manual read of all 4 for the claim in each language

**AC-MPM2-074** (REQ-MPM2-084)
- **Given** the Chinese copy of `multi-llm/model-policy.md`
- **When** it is searched
- **Then** it contains no Haiku column header and no per-agent haiku assignment
- **Verify**:
  ```bash
  grep -in "haiku" docs-site/content/zh/multi-llm/model-policy.md
  # expect: no matches
  ```

**AC-MPM2-075** (REQ-MPM2-085)
- **Given** the 4 locale copies of `multi-llm/model-policy.md`
- **When** their structure is compared
- **Then** all four share the same heading sequence and the same table count
- **Verify**: per-locale heading extraction; sequences identical

**AC-MPM2-076** (REQ-MPM2-086)
- **Given** `.claude/rules/moai/development/model-policy.md`
- **When** it is read end to end
- **Then** it contains no claim that all workers are Sonnet-fixed and no reference to `model_routing_profiles` as a config SSOT
- **Verify**:
  ```bash
  grep -n "model_routing_profiles" .claude/rules/moai/development/model-policy.md
  # expect: no matches
  ```
  plus a manual read confirming the tier table is gone

**AC-MPM2-077** (REQ-MPM2-087)
- **Given** the rule file and its template twin
- **When** they are compared after the edit
- **Then** they are byte-identical
- **Verify**:
  ```bash
  go test ./internal/template/ -run TestRuleTemplateMirror
  ```

**AC-MPM2-078** (REQ-MPM2-088)
- **Given** `agent-authoring.md` and `dynamic-workflows.md`
- **When** their model enums are read
- **Then** both include `fable`
- **Verify**: manual read of the enum line in each file

**AC-MPM2-079** (REQ-MPM2-089)
- **Given** the documentation tree
- **When** it is searched for the retired flag name
- **Then** no user-facing surface instructs the reader to use `--model-policy`
- **Verify**:
  ```bash
  grep -rn -- "--model-policy" docs-site/content README.md README.ko.md README.ja.md README.zh.md .claude/rules/
  # expect: no matches, or only historical-note context
  ```

**AC-MPM2-080** (REQ-MPM2-090)
- **Given** `advanced/profile-matrix.md` in each of the 4 locales
- **When** the naming section is read
- **Then** each states that the profile names denote subscription-tier access rather than performance grade, and that the Max profile is both cheaper and higher-scoring than the Medium profile under the S0-confirmed data
- **Verify**: manual read of all 4; both statements present in each

**AC-MPM2-081** (REQ-MPM2-091)
- **Given** the same page in each locale
- **When** the Max column's Opus assignments are described
- **Then** each states the assignment is a deliberate quality-first choice for high-failure-cost work and not a benchmark optimum
- **Verify**: manual read of all 4

**AC-MPM2-082** (REQ-MPM2-092)
- **Given** every documentation surface edited in M6
- **When** the change set is inspected
- **Then** each edited surface has all four locale copies in the same change
- **Verify**: `git diff --name-only` grouped by surface; every surface shows 4 locale paths

### §D.7 M7 — Guard realignment and verification

**AC-MPM2-090** (REQ-MPM2-100, 101)
- **Given** the haiku-residual rule
- **When** a `haiku` value is reintroduced into the web-console model option set or the v4manifest tier-suggestion table
- **Then** the rule emits a finding for each
- **Verify**: Go test seeding each surface with a haiku value in a temp tree and asserting a finding per surface

**AC-MPM2-091** (REQ-MPM2-102)
- **Given** the haiku-residual rule after M7
- **When** its surface list is read
- **Then** it no longer scans `model_routing_profiles` or `validRoutingModels`
- **Verify**:
  ```bash
  grep -n "model_routing_profiles\|validRoutingModels" internal/spec/lint_haiku_residual.go
  # expect: no matches
  ```

**AC-MPM2-092** (REQ-MPM2-103)
- **Given** the property test
- **When** it runs
- **Then** it asserts all seven invariants of `design.md` §A.3 and passes
- **Verify**: `go test ./internal/template/ -run TestProfileMatrixInvariants -v` shows all seven sub-assertions

**AC-MPM2-093** (REQ-MPM2-104)
- **Given** the repository after M7
- **When** the full suite runs
- **Then** it passes with no skips introduced by this SPEC
- **Verify**:
  ```bash
  go vet ./... && go test ./...
  ```

**AC-MPM2-094** (REQ-MPM2-104)
- **Given** the repository after M7
- **When** the linter runs
- **Then** it reports no new findings relative to the pre-work baseline
- **Verify**:
  ```bash
  golangci-lint run
  ```

**AC-MPM2-095** (REQ-MPM2-105)
- **Given** the release target platforms
- **When** the build runs for each
- **Then** each succeeds
- **Verify**: cross-platform build command per the project's release configuration

### §D.8 Cross-cutting

**AC-MPM2-100** (REQ-MPM2-110)
- **Given** an environment without Fable access
- **When** the documentation is consulted
- **Then** it states the expected fallback for the Max profile, which draws on Fable for six of its eleven cells
- **Verify**: manual read; the fallback is named, not merely acknowledged

**AC-MPM2-101** (REQ-MPM2-111)
- **Given** the GLM-mode documentation
- **When** the profile interaction is read
- **Then** it records that increased Fable usage collapses to `glm-5.2`, that this model has the highest observed step count of the three, and that CG-mode profile pairings warrant review on that basis
- **Verify**: manual read; all three statements present

**AC-MPM2-102** (REQ-MPM2-112)
- **Given** this repository at the default `medium` profile, after the template baseline re-set
- **When** the effort application runs
- **Then** no agent file under `.claude/agents/moai/` changes
- **Verify**:
  ```bash
  git status --porcelain .claude/agents/moai/
  # expect: empty
  ```

**AC-MPM2-103** (REQ-MPM2-113)
- **Given** a user project with a customized `llm.yaml profiles:` block
- **When** the full upgrade path runs
- **Then** the customization is observable afterward either as `agent_overrides` entries or as a warning in the update output
- **Verify**: end-to-end Go test over a temp project asserting one of the two outcomes; a run producing neither fails this criterion

---

## §E Severity classification

| Severity | ACs | Meaning |
|---|---|---|
| **Blocking** | AC-MPM2-001 … 003 | M6 cannot start; a violation means unverified figures reach users |
| **Must-pass** | AC-MPM2-010 … 019, 020 … 026, 030 … 040, 090 … 095, 102, 103 | Correctness and no-regression; a violation is a defect |
| **Must-pass (user-facing)** | AC-MPM2-050 … 054, 070 … 082, 100, 101 | A violation ships misleading information to users |
| **Should-pass** | AC-MPM2-060 … 064 | Cleanup; a violation is cosmetic debt, not a defect |

---

## §F Verification-method distribution

| Method | Count | ACs |
|---|---|---|
| Go test (behavioral) | 33 | most of §D.1 … §D.5, §D.7 |
| Shell command (absence/presence) | 12 | AC-MPM2-011, 018, 020, 021, 022, 050, 063, 074, 076, 079, 091, 102 |
| Manual read (copy quality) | 15 | AC-MPM2-039, 054, 070 … 073, 075, 078, 080, 081, 100, 101, plus partial on 062, 077, 082 |
| Build / toolchain | 4 | AC-MPM2-093, 094, 095, and the build half of 011 |

The manual-read cluster is concentrated in M6 by design: no grep can verify that a paragraph *correctly explains* the naming inversion. Those criteria are satisfied by a reviewer reading the text against `design.md` §E.1 / §E.2, not by a token match.

---

## §G Indirect-verification notes

Three criteria cannot be verified by direct observation of the target behavior and use a proxy:

- **AC-MPM2-003** proxies "no figure was committed before S0" through commit ordering. A figure committed and then reverted before the record would evade it; this is accepted as a low-likelihood evasion.
- **AC-MPM2-032** proxies "the function does not write the template tree" through a pre/post hash snapshot rather than a filesystem-write interceptor. A write followed by an identical rewrite would evade it; also accepted.
- **AC-MPM2-038** verifies the Workflow-path lookup route through the JSON contract rather than through an actual `Workflow` tool invocation, which is not reachable from a Go test. The route's correctness is verified; its consumption by a live workflow is not.

Each is recorded here rather than silently accepted so a reviewer can weigh the residual gap.

---

## §H Closure gates

The SPEC may close when:

1. All Blocking and Must-pass criteria pass with cited evidence.
2. Any failing Should-pass criterion is recorded as debt with an owner.
3. `go vet ./... && go test ./... && golangci-lint run` are green (AC-MPM2-093, 094).
4. The four open clarification items enumerated in `research.md` §F are resolved and the resolutions recorded.
5. The S0 delta table exists, and every benchmark figure in the repository traces to it.
6. No item in `research.md` §G ("explicitly NOT verified") remains unresolved where a milestone depended on it — specifically the ja/ko `multi-llm/model-policy.md` shape, the `advanced/*` line ranges, the `mp.*` orphan status, and the infographic contents, all of which M5/M6 must re-measure rather than inherit.

---

## §I Forward-looking checks

Conditions that would indicate this SPEC's work has regressed after closure:

- A group constant or `agentGroupMembership` reappears in `internal/template`.
- The matrix literal appears in a third file.
- A `profiles:` block reappears in either `llm.yaml`.
- An agent's template frontmatter effort diverges from the matrix `medium` column.
- The effort application writes a `model:` line or a template-tree file.
- A documentation surface gains a benchmark figure with no effort label.
- A locale file gains a key absent from the other three.
