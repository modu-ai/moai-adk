# progress.md — SPEC-V3R6-HARNESS-V4-001

> Plan-phase progress stub. The §E skeleton below carries the canonical 4 placeholder headings (§E.1 through §E.4) per the manager-spec plan-phase artifact protocol. Only §E.1 is populated at plan-phase; §E.2-§E.4 are empty placeholder headings awaiting run-phase / sync-phase population by their owning agents (manager-develop owns §E.2/§E.3; manager-docs owns §E.4).

## §A. Plan-Phase Status

- **Artifact set**: 5 artifacts authored (spec.md / plan.md / acceptance.md / design.md / research.md) + this progress.md stub.
- **Tier**: L (5-artifact set).
- **Era**: V3R6 (3-phase plan→run→sync; MX Tag cross-cutting per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`).
- **Frontmatter**: 12 canonical fields + `era: V3R6` + `tier: L` + `depends_on` + `related_specs`.
- **SPEC ID pre-write self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | HARNESS ✓ | V4 ✓ | 001 ✓ → PASS`.

## §B. Plan-Phase Audit-Ready Signal

- **Pending**: plan-auditor verdict (to be populated by plan-auditor sub-agent at plan-phase gate).
- **Target**: ≥ 0.80 (Tier L threshold).
- **GEARS compliance**: all 13 REQs use GEARS notation (Where/While/When compound clauses + generalized `<subject>`).
- **OutOfScopeRule**: §B.2 contains 5 `### Out of Scope — <topic>` H3 sub-headings each with `-` bullets.
- **Frontmatter schema**: 12 canonical fields present; `created`/`updated` (NOT `created_at`/`updated_at`); `tags` comma-separated string (NOT `labels` array); `version` quoted string.

## §C. Run-Phase Entry Preconditions (for manager-develop)

- [ ] plan-auditor verdict ≥ 0.80 at plan-phase gate.
- [ ] User Implementation Kickoff Approval obtained (CLAUDE.local.md §19.1 / REQ-ATR-015).
- [ ] BI-001 pre-flight: Claude Code subdirectory-command resolution verified.
- [ ] BI-002 pre-flight: `moai update` preserve-logic surface enumerated.
- [ ] BI-003 pre-flight: dynamic-workflow runtime version verified (v2.1.154+).
- [ ] Pre-spawn sync check: `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` clean.

## §D. Milestone Progress (run-phase)

| Milestone | Status | Commit SHA | Notes |
|-----------|--------|------------|-------|
| M1 — `/moai:harness` NL entry + §24 namespace extension | in-progress | <pending push> | namespace protection DONE (AC-010a/b); NL-analysis entry DONE via argument-branching routing (AC-001a/b); learning-subsystem preserved verbatim |
| M2 — Builder Workflow `harness-build.js` (4 phases) | pending | — | core dynamic-workflow |
| M3 — manifest.json schema + Runner primitive-mapping | pending | — | manifest SSOT |
| M4 — `/harness:<name>` + lifecycle + orphan prevention | pending | — | execution + lifecycle |
| M5 — Conditional worktree-isolation | pending | — | sub-agent-granular |
| M6 — Migrate 4 specialists + legacy redirect + dogfooding | pending | — | migration + validation |

---

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-auditor verdict — to be populated at plan-phase gate>_

## §E.2 Run-phase Evidence

### M1 partial — namespace protection (AC-HV4-010a/010b) DONE

**Extended `isUserAreaPath` + `isUserOwnedNamespace`** in `internal/cli/update.go` to recognize `.claude/commands/harness/` and `.claude/workflows/harness-*.js` as user-owned surfaces.

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|---------------------|---------------|
| AC-HV4-010a | PASS | `go test -run TestIsUserOwnedNamespace_HarnessV4CommandsAndWorkflows ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  0.552s` |
| AC-HV4-010b | PASS | `go test -run TestIsUserAreaPath_HarnessV4CommandsAndWorkflows ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  0.552s` |

- **TDD**: RED-phase failing tests in `internal/cli/update_namespace_harness_v4_test.go` → GREEN-phase `HasPrefix` checks in `update.go` → all tests pass.
- **Regression**: existing `TestIsUserOwnedNamespace_HarnessV2DualRecognition` + `TestIsUserAreaPath_HarnessV2Canonical` + `TestUpdate_PreserveHarnessV2Namespace` + `TestUpdate_AssertNoUserOwnedNamespaceTouch` all still PASS (BI-002 contract honored — no existing user-owned classification broken).
- **Coverage**: `isUserAreaPath` 92.9%, `isUserOwnedNamespace` 96.7% (both ≥ 85% threshold).
- **Doctrine**: `.moai/docs/harness-namespace-doctrine.md` §24.4 contract matrix extended with the two new user-owned rows.

### M1 — `/moai:harness` NL-analysis entry via argument-branching (AC-HV4-001a/001b) DONE

**Resolution (user decision 2026-06-19 via orchestrator AskUserQuestion)**: the `/moai:harness` path conflict is resolved by **argument-based routing** in a SINGLE command (not a new command, not M6 deferral). Reserved verbs (`status|apply|rollback|disable`) → existing learning-subsystem workflow (unchanged); natural-language request → NEW v4 NL-analysis entry. This preserves learning-subsystem users (zero impact) AND lands REQ-HV4-001 at the canonical `/moai:harness` path with NO SPEC body change.

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|---------------------|---------------|
| AC-HV4-001a | PASS | `grep -n "Branch B — harness build entry" internal/template/templates/.claude/skills/moai/SKILL.md` | `239:#### Branch B — harness build entry (natural-language request)` (NL-analysis entry declared; Context-First Discovery domain/goal/constraints/scope extraction specified in `${CLAUDE_SKILL_DIR}/workflows/harness-build-entry.md` Phase 1) |
| AC-HV4-001b | PASS | `grep -nE 'AskUserQuestion Socratic rounds|derived from the request' internal/template/templates/.claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/workflows/harness-build-entry.md` | matches in both files — Branch B description states "MUST conduct AskUserQuestion Socratic rounds when clarity <100%" + "name is derived from the request — NOT statically supplied by the user"; harness-build-entry.md Phase 1.5 + Phase 2 specify the Socratic round structure + name derivation rules |

**Deliverables**:
- `moai SKILL.md` harness section (template + local mirror): extended with argument-branching routing rule + Branch A (learning lifecycle, verbatim-preserved) + Branch B (NL-analysis entry). Learning-subsystem verbs preserved unchanged — `grep -c "status / apply / rollback / disable"` returns 2 in both SKILL.md copies (Branch A description + the dispatcher rule).
- NEW workflow `harness-build-entry.md` (template + local mirror): implements Context-First Discovery (Phase 1 + 1.5 Socratic rounds) → harness `<name>` derivation (Phase 2, derived NOT statically supplied) → orchestrator-issued approval gate (Phase 3, AskUserQuestion 4-option pattern) → Builder delegation forward-link (Phase 4, `harness-build.js` not yet implemented — explicit forward-link + graceful-degradation contract).
- `commands/moai/harness.md` (template + local mirror): thin-wrapper body UNCHANGED (`Use Skill("moai") with arguments: harness $ARGUMENTS`); frontmatter `description` widened to cover both surfaces, `argument-hint` widened to `{status|apply|rollback <YYYY-MM-DD>|disable|<natural-language request>}`. `commands_audit_test.go` (TestCommandsThinPattern + TestCommandsFrontmatterConsistency) PASSES — thin-wrapper LOC + frontmatter CSV constraints honored.
- `make build` regenerated `internal/template/embedded.go` + `internal/template/catalog.yaml` (mechanical, hash update only).

**Constraints honored**:
- C-HV4-005 (template neutrality): the NEW workflow file contains ZERO SPEC-IDs / REQ tokens / AC tokens / commit SHAs (self-checked via `grep -nE 'SPEC-V3R6-HARNESS-V4-001|REQ-HV4-[0-9]+|AC-HV4-[0-9]+|823b20570'` → CLEAN). Pre-existing `SPEC-V3R4-HARNESS-001` reference in SKILL.md Branch A is the authoritative SPEC for the learning subsystem (predates this change, unchanged).
- Learning-subsystem preservation: Branch A description is verbatim the pre-change content; the 4 reserved verbs route to `workflows/harness.md` unchanged. `workflows/harness.md` itself was NOT modified (scope discipline).
- Thin Command Pattern: harness.md body remains 1 non-empty line (`Use Skill("moai")...`); well under the 20-LOC cap.
- Conventional Commits + `🗿 MoAI` trailer (see commit below).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
