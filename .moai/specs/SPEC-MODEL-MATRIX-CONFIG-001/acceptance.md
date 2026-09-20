# Acceptance Criteria — SPEC-MODEL-MATRIX-CONFIG-001

20 acceptance criteria across two milestones. Every criterion states observable evidence; no criterion is satisfied by assertion alone.

Nothing in this SPEC has landed, so every criterion below is unmet at authoring — that is the expected starting state, not a defect.

---

## §D AC Matrix

### §D.1 M2 — Retire the 36-cell axis and the config mirror

**AC-MPME-001** (REQ-MPME-001)
- **Given** both `workflow.yaml` copies
- **When** they are searched
- **Then** neither contains a `model_routing_profiles` block nor a comment referencing it
- **Verify**:
  ```bash
  grep -rn "model_routing_profiles" .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml
  # expect: no matches
  ```

**AC-MPME-002** (REQ-MPME-002, 003)
- **Given** the config package
- **When** it is compiled and searched
- **Then** `RouteModelFor`, `ModelRoutingProfiles`, and `validRoutingModels` do not exist, and no orphaned test references them
- **Verify**:
  ```bash
  grep -rn "RouteModelFor\|ModelRoutingProfiles\|validRoutingModels" --include="*.go" internal/
  # expect: no matches
  go test ./internal/config/...
  ```

**AC-MPME-003** (REQ-MPME-004, 005)
- **Given** both `llm.yaml` copies
- **When** they are searched for a `profiles:` key under `llm:`
- **Then** neither contains one, and the matrix literal exists in exactly two files (the Go constant and its fidelity test)
- **Verify**:
  ```bash
  grep -rn "^\s*profiles:" .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml
  # expect: no matches
  ```
  plus a manual enumeration confirming exactly 2 files carry the 33-cell literal

**AC-MPME-004** (REQ-MPME-004, K-5)
- **Given** the template `llm.yaml` before and after M2
- **When** the diff is inspected
- **Then** only the `profiles:` block and its explanatory comment were removed; the `glm.models` block is byte-identical to its pre-M2 state
- **Verify**: `git diff` on the template file shows no change inside `glm.models`

**AC-MPME-005** (REQ-MPME-008)
- **Given** a project whose `llm.yaml` still carries a `profiles:` block and whose `workflow.yaml` still carries `model_routing_profiles`
- **When** the config loader runs
- **Then** it succeeds, treating both blocks as inert
- **Verify**: Go test loading a legacy fixture containing both blocks and asserting no error

**AC-MPME-006** (REQ-MPME-006, 007, 009)
- **Given** a project whose `llm.yaml` `profiles:` block differs from the shipped default
- **When** `moai update` runs
- **Then** the customization is either migrated into `agent_overrides` or surfaced in a warning naming the affected profile column and cells — and in neither case is it discarded silently
- **Verify**: Go test with a customized fixture asserting that after the update either the override map is populated with the equivalent per-agent entries, or the emitted warning text names the changed cells

**AC-MPME-007** (REQ-MPME-005, K-2)
- **Given** the test suite after M2
- **When** it runs
- **Then** the test asserting that a config `profiles` cell overrides the Go default no longer exists, and the suite is green
- **Verify**:
  ```bash
  grep -rn "ConfigProfilesOverrideDefault" --include="*_test.go" internal/
  # expect: no matches
  go test ./internal/template/... ./internal/config/...
  ```

**AC-MPME-008** (REQ-MPME-009)
- **Given** a user project with a customized `llm.yaml profiles:` block
- **When** the full upgrade path runs
- **Then** the customization is observable afterward either as `agent_overrides` entries or as a warning in the update output
- **Verify**: end-to-end Go test over a temp project asserting one of the two outcomes; a run producing neither fails this criterion

### §D.2 M3 — Effort actualization

**AC-MPME-009** (REQ-MPME-010, 018)
- **Given** a deployed project at profile `low` whose `manager-develop.md` carries `effort: xhigh`
- **When** the effort application runs
- **Then** the file's `effort:` becomes `medium` and every other byte of the file is unchanged
- **Verify**: Go test comparing pre/post file bytes; exactly one line differs

**AC-MPME-010** (REQ-MPME-011)
- **Given** any deployed agent file with `model: inherit`
- **When** the effort application runs under any profile
- **Then** the `model:` line is unchanged
- **Verify**: Go test asserting the `model:` value is byte-identical pre/post for all 10 files, under all 3 profiles

**AC-MPME-011** (REQ-MPME-012)
- **Given** the template tree
- **When** the effort application runs
- **Then** no file under `internal/template/templates/.claude/agents/` is modified
- **Verify**: Go test snapshotting the template agent directory's file hashes pre/post and asserting equality

**AC-MPME-012** (REQ-MPME-013)
- **Given** the template tree after the predecessor's M1 step 7
- **When** each template agent's `effort:` is compared to the matrix's `medium` column
- **Then** every value matches
- **Verify**: Go test reading the embedded template agent frontmatter and comparing against the `medium` column

**AC-MPME-013** (REQ-MPME-014)
- **Given** a project at profile `max` with agent efforts already applied
- **When** `moai update` deploys templates and completes
- **Then** the deployed agent efforts equal the `max` column, not the template baseline
- **Verify**: integration-style Go test over a temp project: set profile, apply, update, re-read

**AC-MPME-014** (REQ-MPME-015)
- **Given** a config with `agent_overrides["manager-spec"] = {opus, xhigh}` and profile `low`
- **When** the effort application runs
- **Then** `manager-spec.md` retains `effort: xhigh` while `plan-auditor.md` becomes the `low` cell's `low`
- **Verify**: Go test asserting both files

**AC-MPME-015** (REQ-MPME-016)
- **Given** the four profile-application seams
- **When** each is inspected
- **Then** each invokes the effort application, and at every seam that deploys templates the invocation follows the deploy
- **Verify**: code read of `internal/cli/init.go`, both `internal/cli/update.go` sites, and `internal/web/agentfm.go`, plus AC-MPME-013's ordering test

**AC-MPME-016** (REQ-MPME-017)
- **Given** the matrix contains `Explore`, which has no file on disk
- **When** the effort application runs
- **Then** it completes without error and without creating an `Explore.md`
- **Verify**: Go test asserting no error and no new file

**AC-MPME-017** (REQ-MPME-019)
- **Given** the active profile and an agent name
- **When** a workflow script queries the machine-readable route
- **Then** it receives that agent's resolved `model` and `effort` for the active profile
- **Verify**:
  ```bash
  moai model profile --json
  ```
  output parsed for a named agent yields both fields; a Go test asserts the JSON shape

**AC-MPME-018** (REQ-MPME-020, 021)
- **Given** the documentation content produced by M3
- **When** the effort-injection section is read
- **Then** it distinguishes the frontmatter channel (Agent tool, no `effort` parameter) from the `opts.effort` channel (Workflow tool), records that `SPEC-MODEL-PROFILE-MATRIX-001` DECISION-001's wording is superseded, and notes that `Explore` has no agent file
- **Verify**: manual read of the section; all three statements present. Placement and locale coverage are verified by `SPEC-MODEL-MATRIX-DOCS-001`, not here

**AC-MPME-019** (REQ-MPME-010, K-3)
- **Given** `internal/web/agentfm.go`
- **When** the `applyPerfTierEdits` comment block is read
- **Then** it no longer claims frontmatter re-application is retired
- **Verify**:
  ```bash
  grep -n "retired" internal/web/agentfm.go
  ```
  no surviving hit asserts the retirement of frontmatter re-application

**AC-MPME-020** (REQ-MPME-022)
- **Given** this repository at the default `medium` profile, after the template baseline re-set
- **When** the effort application runs
- **Then** no agent file under `.claude/agents/moai/` changes
- **Verify**:
  ```bash
  git status --porcelain .claude/agents/moai/
  # expect: empty
  ```

---

## §E Severity classification

| Severity | ACs | Meaning |
|---|---|---|
| **Must-pass** | AC-MPME-001 … 017, 019, 020 | Correctness and no-regression; a violation is a defect |
| **Must-pass (user-facing)** | AC-MPME-008, 018 | A violation either destroys a user's configuration or ships a misleading explanation of how effort reaches an agent |

AC-MPME-008 appears in both rows deliberately: it is the end-to-end expression of the no-silent-drop prohibition, and the only criterion whose failure means data loss in a user's project rather than a defect in ours.

---

## §F Verification-method distribution

| Method | Count | ACs |
|---|---|---|
| Go test (behavioral) | 13 | AC-MPME-005, 006, 008 … 017 |
| Shell command (absence/presence) | 5 | AC-MPME-001, 002, 003, 019, 020 |
| Diff inspection | 1 | AC-MPME-004 |
| Manual read (copy quality) | 1 | AC-MPME-018 |

AC-MPME-003 and AC-MPME-015 carry a manual half alongside their mechanical half — the literal-count enumeration and the four-seam code read respectively. Neither half alone satisfies its criterion.

---

## §G Indirect-verification notes

Three criteria cannot be verified by direct observation of the target behaviour and use a proxy:

- **AC-MPME-011** proxies "the function does not write the template tree" through a pre/post hash snapshot rather than a filesystem-write interceptor. A write followed by an identical rewrite would evade it; accepted as low-likelihood.
- **AC-MPME-017** verifies the Workflow-path lookup route through the JSON contract rather than through an actual `Workflow` tool invocation, which is not reachable from a Go test. The route's correctness is verified; its consumption by a live workflow is not.
- **AC-MPME-020** proxies "the application is a no-op" through working-tree cleanliness. It passes vacuously if the application never ran — the criterion must be evaluated with evidence that the invocation occurred, not merely that the tree is clean.

Each is recorded here rather than silently accepted so a reviewer can weigh the residual gap.

---

## §H Closure gates

This SPEC may close when:

1. All Must-pass criteria pass with cited evidence.
2. The `profiles:` migration-representability clarification in `research.md` §F is resolved and the resolution recorded — including whether REQ-MPME-007's warning branch survives.
3. The predecessor `SPEC-MODEL-MATRIX-CORE-001` has landed its M1 step 7 template-baseline re-set, without which AC-MPME-012 and AC-MPME-020 cannot hold.
4. The handoff to `SPEC-MODEL-MATRIX-DOCS-001` is recorded: the haiku-residual rule's `model_routing_profiles` and `validRoutingModels` surfaces now scan deleted artifacts.
5. `go vet ./... && go test ./internal/config/... ./internal/template/... ./internal/cli/... ./internal/web/...` are green.

---

## §I Forward-looking checks

Conditions that would indicate this SPEC's work has regressed after closure:

- A `profiles:` block reappears in either `llm.yaml`.
- The matrix literal appears in a third file.
- `RouteModelFor` or a `model_routing_profiles` block reappears.
- The effort application writes a `model:` line or a template-tree file.
- A profile-application seam is added without an effort-application call, or with one placed before a template deploy.
- The config loader starts rejecting a legacy block.
