# acceptance.md — SPEC-V3R6-HARNESS-V4-001 Acceptance Criteria

> One or more AC per REQ. Each AC is observable, testable, and traces to a REQ + milestone. Given-When-Then format. Severity: MUST-FIX / SHOULD-FIX / NICE-TO-HAVE.

## §D. AC Matrix

| AC ID | REQ | Milestone | Severity | Scenario |
|-------|-----|-----------|----------|----------|
| AC-HV4-001a | REQ-HV4-001 | M1 | MUST-FIX | NL analysis extracts domain/goal/constraints/scope |
| AC-HV4-001b | REQ-HV4-001 | M1 | MUST-FIX | AskUserQuestion Socratic rounds when clarity <100%; name derived |
| AC-HV4-002a | REQ-HV4-002 | M4 | MUST-FIX | `/harness:<name>` command auto-generated at GENERATE time |
| AC-HV4-002b | REQ-HV4-002 | M4 | MUST-FIX | `/harness:<name>` invokes the harness's Runner Workflow |
| AC-HV4-003a | REQ-HV4-003 | M2 | MUST-FIX | Builder runs 4 phases as dynamic-workflow |
| AC-HV4-003b | REQ-HV4-003 | M2 | MUST-FIX | ≥1 phase skipped under load-bearing-minimum (simple task) |
| AC-HV4-004a | REQ-HV4-004 | M2 | MUST-FIX | ≥2 patterns chosen for ≥2 different domain requests |
| AC-HV4-004b | REQ-HV4-004 | M2 | SHOULD-FIX | Patterns recorded in manifest.json |
| AC-HV4-005a | REQ-HV4-005 | M3 | MUST-FIX | Each specialist has primitive mapping in manifest |
| AC-HV4-005b | REQ-HV4-005 | M3 | MUST-FIX | Runner consumes primitive verbatim (no re-derivation) |
| AC-HV4-006a | REQ-HV4-006 | M3 | MUST-FIX | manifest.json matches canonical schema |
| AC-HV4-006b | REQ-HV4-006 | M3 | SHOULD-FIX | manifest is machine-readable + single source of truth |
| AC-HV4-007a | REQ-HV4-007 | M5 | MUST-FIX | 0 worktrees for read-only ANALYZE |
| AC-HV4-007b | REQ-HV4-007 | M5 | MUST-FIX | ≥1 worktree for conflict-prone GENERATE |
| AC-HV4-008a | REQ-HV4-008 | M3 | MUST-FIX | Sprint Contract present in manifest |
| AC-HV4-008b | REQ-HV4-008 | M3 | MUST-FIX | Evaluator conditional (skipped for simple tasks) |
| AC-HV4-009a | REQ-HV4-009 | M6 | MUST-FIX | revfactory 7-Phase residual grep = 0 in new v4 artifacts |
| AC-HV4-010a | REQ-HV4-010 | M1 | MUST-FIX | `moai update` preserves `.claude/commands/harness/` |
| AC-HV4-010b | REQ-HV4-010 | M1 | MUST-FIX | `moai update` preserves `.claude/workflows/harness-*.js` |
| AC-HV4-011a | REQ-HV4-011 | M4 | MUST-FIX | `/moai:harness list` enumerates harnesses |
| AC-HV4-011b | REQ-HV4-011 | M4 | MUST-FIX | `/moai:harness remove <name>` is atomic |
| AC-HV4-011c | REQ-HV4-011 | M4 | MUST-FIX | Orphan command cannot remain (fail-closed) |
| AC-HV4-012a | REQ-HV4-012 | M6 | MUST-FIX | "moai-adk-dev" harness built + runs real task |
| AC-HV4-012b | REQ-HV4-012 | M6 | MUST-FIX | A/B disclosed as author-measured (5-Section format) |
| AC-HV4-013a | REQ-HV4-013 | M6 | MUST-FIX | 4 specialists migrated to v4 manifest format |
| AC-HV4-013b | REQ-HV4-013 | M6 | MUST-FIX | Legacy `/moai:harness` 7-Phase → v4 redirect live |

## §D.1 AC Details (Given-When-Then)

### AC-HV4-001a — NL analysis extracts domain/goal/constraints/scope (MUST-FIX, M1)
**Given** a user issues `/moai:harness build a harness for moai-adk CLI template development`.
**When** the orchestrator runs Context-First Discovery on the request.
**Then** the extracted profile contains at minimum: domain ("moai-adk CLI template development"), goal, constraints (16-language neutrality, §25 template-neutrality), scope (CLI template files).
**Evidence**: orchestrator emits the extracted profile as a structured block before approval gate.

### AC-HV4-001b — AskUserQuestion when clarity <100%; name derived (MUST-FIX, M1)
**Given** the NL request has an ambiguous scope boundary (e.g., "moai-adk development" without specifying CLI vs docs vs hooks).
**When** the orchestrator assesses clarity <100%.
**Then** the orchestrator runs AskUserQuestion Socratic rounds (≤4 questions/round, `(Recommended)` first option per `.claude/rules/moai/core/askuser-protocol.md`) until clarity reaches 100%, AND derives a harness `<name>` (e.g., `moai-adk-dev`) from the confirmed intent — the name is NOT statically supplied by the user.
**Evidence**: AskUserQuestion call log + derived name surfaced in the approval gate.

### AC-HV4-002a — `/harness:<name>` command auto-generated (MUST-FIX, M4)
**Given** the Builder GENERATE phase completes successfully for harness `<name>`.
**When** the command generator runs.
**Then** a file exists at `.claude/commands/harness/<name>.md` that is a thin wrapper dispatching to the harness's Runner Workflow `harness-<name>-run.js`.
**Evidence**: `ls .claude/commands/harness/<name>.md` succeeds; file content references the Runner Workflow.

### AC-HV4-002b — `/harness:<name>` invokes the Runner (MUST-FIX, M4)
**Given** the command file `.claude/commands/harness/<name>.md` exists.
**When** the user issues `/harness:<name>` in Claude Code.
**Then** the harness's Runner Workflow `harness-<name>-run.js` is invoked (the Runner reads `manifest.json` and dispatches specialists per primitive).
**Evidence**: Runner Workflow execution trace showing manifest read + specialist dispatch.

### AC-HV4-003a — Builder runs 4 phases as dynamic-workflow (MUST-FIX, M2)
**Given** the Builder Workflow `harness-build.js` is invoked.
**When** it runs end-to-end for a non-trivial domain request.
**Then** all 4 phases execute: ANALYZE (Explore fan-out) → PLAN (opus xhigh sub-agent) → GENERATE (fan-out) → ACTIVATE (`/goal` + A/B), and the workflow script holds the plan (intermediate results in script variables, NOT Claude's context).
**Evidence**: workflow run log shows all 4 phase records; script-body inspection confirms plan-in-script structure.

### AC-HV4-003b — ≥1 phase skipped under load-bearing-minimum (MUST-FIX, M2)
**Given** a simple task that is within the model's solo reliable range (e.g., single-skill generation with no adversarial verification need).
**When** the Builder runs.
**Then** at least one phase is skipped (most commonly ACTIVATE's A/B evaluation per REQ-HV4-008), and the skip is recorded in the run log with the rationale ("task within solo range, evaluator skipped per C-HV4-001").
**Evidence**: run log shows a skipped-phase record with rationale; harness still produces correct output.

### AC-HV4-004a — ≥2 patterns for ≥2 different domain requests (MUST-FIX, M2)
**Given** two different domain requests (e.g., "build a research harness" vs "build a code-review harness").
**When** the PLAN phase selects patterns for each.
**Then** the two harnesses have different pattern selections in their manifests (e.g., research → Fan-out/Fan-in + Expert Pool; code-review → Producer-Reviewer + Pipeline). At least 2 distinct patterns are chosen across the 2 harnesses.
**Evidence**: `manifest.json.patterns` differs between the 2 harnesses; pattern selection is NOT fixed.

### AC-HV4-004b — Patterns recorded in manifest (SHOULD-FIX, M2)
**Given** the PLAN phase selects patterns.
**When** the manifest is emitted.
**Then** `manifest.json.patterns` is an array of pattern names drawn from the 6-pattern catalog {Pipeline, Fan-out/Fan-in, Expert Pool, Producer-Reviewer, Supervisor, Hierarchical}.
**Evidence**: JSON schema validation of `patterns` field.

### AC-HV4-005a — Each specialist has primitive mapping (MUST-FIX, M3)
**Given** the PLAN phase defines N specialist roles.
**When** the manifest is emitted.
**Then** each of the N `specialists[]` entries has a `primitive` field set to exactly one of {sub-agent, dynamic-workflow, worktree, `/goal`, adversarial-fan-out}, an `isolation` field (none|worktree), an `effort` field (low|medium|high|xhigh|max), and a `model` field (inherit|haiku|sonnet|opus).
**Evidence**: manifest JSON inspection; 0 specialists missing any of the 5 fields.

### AC-HV4-005b — Runner consumes primitive verbatim (MUST-FIX, M3)
**Given** a manifest declares specialist S with `primitive: "dynamic-workflow"`.
**When** the Runner dispatches S.
**Then** S is dispatched as a dynamic-workflow `agent()` call — NOT as a sub-agent or worktree. The Runner does NOT re-derive the primitive from heuristics; it reads `specialist.primitive` and dispatches accordingly.
**Evidence**: Runner Workflow code path for S traces to a dynamic-workflow dispatch; grep confirms no heuristic re-derivation of primitive in Runner.

### AC-HV4-006a — manifest.json matches canonical schema (MUST-FIX, M3)
**Given** a generated `manifest.json`.
**When** it is validated against the canonical schema (design.md §C).
**Then** all 8 top-level fields are present and well-typed: `name` (string), `domain` (string), `source_request` (string), `patterns` (array), `specialists` (array of {role,primitive,isolation,effort,model}), `sprint_contract` ({dimensions:array, thresholds:map}), `entry_command` (string `/harness:<name>`), `runner_workflow` (string `harness-<name>-run.js`).
**Evidence**: JSON schema validation passes; 0 missing fields.

### AC-HV4-006b — manifest is single source of truth (SHOULD-FIX, M3)
**Given** a harness has been built.
**When** the Runner executes.
**Then** the Runner reads ONLY `manifest.json` for its dispatch logic — it does NOT read a separate config file or hard-code specialist info.
**Evidence**: Runner Workflow source has exactly one config-read path (`manifest.json`).

### AC-HV4-007a — 0 worktrees for read-only ANALYZE (MUST-FIX, M5)
**Given** the Builder is in the ANALYZE phase (Explore fan-out, read-only).
**When** the ANALYZE fan-out runs.
**Then** ZERO worktrees are created — ANALYZE sub-agents run main-tree read-only.
**Evidence**: `git worktree list` shows no new worktrees during ANALYZE; run log records isolation=none for all ANALYZE agents.

### AC-HV4-007b — ≥1 worktree for conflict-prone GENERATE (MUST-FIX, M5)
**Given** the Builder is in the GENERATE phase with ≥2 specialists targeting overlapping file paths (conflict-prone parallel generation).
**When** the GENERATE fan-out runs.
**Then** at least 1 specialist is spawned with `Agent(isolation:"worktree")` (per its manifest `isolation:worktree` declaration); NO mandatory top-level worktree wraps the entire Builder.
**Evidence**: run log records ≥1 isolation=worktree specialist; `git worktree list` shows the spawned worktree(s); Builder top-level isolation=none.

### AC-HV4-008a — Sprint Contract present in manifest (MUST-FIX, M3)
**Given** a generated manifest.
**When** it is inspected.
**Then** `manifest.json.sprint_contract` is an object with `dimensions` (array of graded dimension names) and `thresholds` (map of dimension→threshold value), per Anthropic GAN Sprint Contract pattern.
**Evidence**: JSON inspection of `sprint_contract` field.

### AC-HV4-008b — Evaluator conditional (MUST-FIX, M3)
**Given** a task that is within the model's solo reliable range (simple, well-bounded).
**When** the Runner dispatches the evaluator.
**Then** the evaluator is SKIPPED (per REQ-HV4-008 / C-HV4-001 simplest-solution-first), and the skip is recorded with rationale. The evaluator is invoked ONLY when the task exceeds the model's solo range.
**Evidence**: run log records evaluator=skipped with rationale for a simple task; evaluator=invoked for a complex task.

### AC-HV4-009a — revfactory 7-Phase residual grep = 0 (MUST-FIX, M6)
**Given** the v4 artifacts (Builder Workflow, Runner Workflow template, generated specialists, generated skills, manifest schema doc).
**When** `grep -rE '7-Phase|Phase 7 LEARNING|Skeleton|Customization'` runs over the newly generated v4 artifacts.
**Then** the grep returns ZERO matches. (Legacy `moai-meta-harness` SKILL.md is excluded — it is the redirect source, not a v4 artifact.)
**Evidence**: grep command output showing 0 matches, with the artifact path list enumerated.

### AC-HV4-010a — `moai update` preserves `.claude/commands/harness/` (MUST-FIX, M1)
**Given** a user-owned harness command file exists at `.claude/commands/harness/dev.md`.
**When** `moai update` runs.
**Then** `.claude/commands/harness/dev.md` is preserved (backed up if needed, NOT deleted or overwritten), per the extended §24 namespace protection.
**Evidence**: file exists post-`moai update`; backup log entry if backup occurred.

### AC-HV4-010b — `moai update` preserves `.claude/workflows/harness-*.js` (MUST-FIX, M1)
**Given** a user-owned harness Runner Workflow exists at `.claude/workflows/harness-dev-run.js`.
**When** `moai update` runs.
**Then** `harness-dev-run.js` is preserved (NOT deleted or overwritten).
**Evidence**: file exists post-`moai update`.

### AC-HV4-011a — `/moai:harness list` enumerates harnesses (MUST-FIX, M4)
**Given** two harnesses `dev` and `research` exist (each has `.claude/commands/harness/<name>.md` + manifest).
**When** the user issues `/moai:harness list`.
**Then** the output enumerates both harnesses with their name, domain (from manifest), and entry command.
**Evidence**: list command output showing both harnesses.

### AC-HV4-011b — `/moai:harness remove <name>` is atomic (MUST-FIX, M4)
**Given** harness `dev` exists (command + workflow + specialists + skills + manifest).
**When** the user issues `/moai:harness remove dev`.
**Then** ALL of the following are removed: `.claude/commands/harness/dev.md`, `.claude/workflows/harness-dev-run.js`, `.claude/agents/harness/dev-*-specialist.md`, `.claude/skills/harness-dev-*/`, and the manifest. None remain.
**Evidence**: post-remove filesystem scan shows 0 `dev` artifacts.

### AC-HV4-011c — Orphan command cannot remain (fail-closed) (MUST-FIX, M4)
**Given** harness `dev` exists but its manifest has been manually deleted (simulated partial state).
**When** the user issues `/moai:harness remove dev`.
**Then** the remove operation FAILS CLOSED — it refuses to remove the command file alone (which would leave the manifest referenced by the command orphaned), surfaces the missing-manifest error, and leaves the filesystem unchanged.
**Evidence**: remove command exits non-zero; error message names the missing manifest; `.claude/commands/harness/dev.md` still exists (unchanged).

### AC-HV4-012a — "moai-adk-dev" harness built + runs real task (MUST-FIX, M6)
**Given** the v4 Builder is operational.
**When** the dogfooding validation builds a "moai-adk-dev" harness with v4.
**Then** the harness is built successfully (manifest + Runner + command emitted) AND runs a real moai-adk development task end-to-end (sample task: e.g., "add a new hook event handler" or "extend the template-neutrality CI guard" — disclosed in the validation report).
**Evidence**: dogfooding validation report (5-Section Evidence-Bearing Format) with the sample task, run trace, and output.

### AC-HV4-012b — A/B disclosed as author-measured (MUST-FIX, M6)
**Given** the dogfooding validation runs a with/without A/B comparison (harness vs no-harness baseline).
**When** the result is reported.
**Then** the report explicitly discloses: (a) sample size n, (b) "author-measured, 3rd-party replication pending", (c) the 5-Section Evidence-Bearing Report Format (Claim / Evidence with verbatim command+output / Baseline-attribution / Gaps / Residual-risk) per `.claude/rules/moai/core/verification-claim-integrity.md`. The revfactory +60% claim is cited as upstream provenance, NOT co-opted as a verified v4 claim.
**Evidence**: validation report file with the 5-section structure and the disclosure disclaimer.

### AC-HV4-013a — 4 specialists migrated to v4 manifest format (MUST-FIX, M6)
**Given** the 4 existing specialists (cli-template, quality, workflow, hook-ci) pre-v4.
**When** M6 migration completes.
**Then** each specialist has a v4 manifest entry (role / primitive / isolation / effort / model) AND the existing Layer B specialist-team regression suite passes (behavior preserved).
**Evidence**: 4 manifest entries; regression suite green.

### AC-HV4-013b — Legacy `/moai:harness` 7-Phase → v4 redirect live (MUST-FIX, M6)
**Given** the legacy `moai-meta-harness` 7-Phase path.
**When** a user invokes the legacy path.
**Then** the system surfaces a deprecation notice AND routes to `/moai:harness` v4 (the new NL-analysis entry). The legacy 7-Phase body is marked superseded (the `@MX:NOTE [AUTO] V3R4 contract` annotation is honored or explicitly superseded with a recorded rationale).
**Evidence**: invocation of legacy path shows deprecation notice + redirect; annotation handling recorded.

## §D.2 Indirect Verification (where direct AC is impractical)

- **Dynamic-workflow runtime availability** (R-HV4-001): the Builder is designed to fall back to sequential sub-agent mode if dynamic-workflow is unavailable. AC-HV4-003a assumes runtime availability; the fallback path is verified via a unit test that mocks the runtime as unavailable and asserts the fallback is triggered.
- **Worktree L1 autonomous cleanup** (R-HV4-002): the Runner emits a cleanup directive at end-of-run, but L1 worktree cleanup is runtime-autonomous. AC-HV4-007b verifies worktree creation; cleanup is verified via `git worktree list` showing no accumulation across 3 consecutive runs.

## §D.3 Closure Gates (Definition of Done for this SPEC)

- [ ] All 24 MUST-FIX ACs pass (AC-HV4-001a/b, 002a/b, 003a/b, 004a, 005a/b, 006a, 007a/b, 008a/b, 009a, 010a/b, 011a/b/c, 012a/b, 013a/b).
- [ ] Both SHOULD-FIX ACs pass or are explicitly deferred with rationale (AC-HV4-004b, 006b).
- [ ] spec-lint passes on this SPEC (12 canonical frontmatter fields, era: V3R6, OutOfScope section present, GEARS notation).
- [ ] plan-auditor verdict ≥ 0.80 (Tier L threshold) at plan-phase gate.
- [ ] Run-phase: `go test ./internal/cli/... ./internal/harness/...` green; `golangci-lint run` clean on changed packages.
- [ ] Sync-phase: frontmatter `status: completed`; CHANGELOG entry; legacy `moai-meta-harness` redirect documented.

## §D.4 Forward-Looking Checks (deferred to follow-up SPECs)

- **Learning-subsystem outer loop** — the Meta-Harness outer-loop optimization (arxiv 2603.28052) is declared cross-cutting in the Sprint Contract but its internals are out of scope (spec.md §B.2). Forward-link: candidate `SPEC-V3R6-HARNESS-LEARNING-LOOP-001`.
- **Downstream user-project migration helper** — user projects with legacy 7-Phase harnesses need a migration path (spec.md §B.2). Forward-link: candidate `SPEC-V3R6-HARNESS-V4-MIGRATION-001`.
- **Component-removal experiments** — as models improve (C-HV4-001), the v4 design should be amenable to removing Sprint Contract / evaluator / context-reset one at a time and measuring. Forward-link: candidate `SPEC-V3R6-HARNESS-COMPONENT-PRUNE-001` (deferred until model improvement warrants).
