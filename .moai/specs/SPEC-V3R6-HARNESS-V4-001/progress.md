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

## §E.0 Phase 0.95 Mode Selection

- **Input parameters**: tier=L, scope≈15-20 files, domains=6 (CLI/template/commands/workflows/agents/doctrine), language mix=Go+markdown+shell+JSON, concurrency benefit=LOW (coding-heavy per Anthropic coding-task parallelism caveat), Agent Teams prereqs=N/A (coding-heavy)
- **Decision**: `sub-agent` (Mode 5)
- **Mode evaluation**: trivial✗ / background✗(writes) / agent-team✗(coding-heavy+prereqs) / parallel✗(sequential milestone deps) / **sub-agent✓** / workflow✗(not mechanical-uniform)
- **Justification**: Tier L coding-heavy work. Per Anthropic's coding-task parallelism caveat, Mode 5 sequential per-milestone is the correct default. Milestones M1-M6 are sequentially ordered (M1 namespace protection unblocks M2-M6) — not genuinely parallel. Implementation Kickoff Approval obtained 2026-06-19.

## §E.1 Plan-phase Audit-Ready Signal

- **plan-auditor iter 2/3 verdict (post-pivot 2026-06-19)**: PASS-WITH-DEBT 0.88 (Tier L threshold 0.85, +0.03 margin). Pivot: Builder → orchestrator-direct (Runner stays dynamic-workflow). Self-contradiction RESOLVED (C-HV4-004 Runner-only + §D.2 gate reachable); PRESERVE PASS. D1/D2/D3 ALL FIXED (D3 SHA auditor-suggested `68ca6c03f` → verified `823b20570`). Implementation Kickoff Approval (M2) obtained 2026-06-19.
- **plan-auditor iter 1/3 verdict (pre-pivot)**: PASS-WITH-DEBT 0.87 (Tier L threshold 0.85, +0.02 margin)
- **Skip-eligible**: NO (score < 0.90)
- **Must-pass**: 7/7 PASS (MP-1 REQ consistency / MP-2 GEARS+traceability / MP-3 frontmatter+lint / MP-4 absorb/remove mapping / MP-5 conditional-worktree / MP-6 namespace SSOT / MP-7 4-phase internal)
- **Defects (live-verified 2026-06-19)**: D1 acceptance arithmetic(22→24), D2 DYNAMIC-WORKFLOWS-DOC reference, D3 LIFECYCLE-REDESIGN stale status — **all 3 confirmed RESOLVED on disk** (spec.md v0.1.1 post-audit update at 05:45; the 05:41 report was stale). D4/D5 NFR AC binding gap (MINOR) — deferred to sync-phase debt per plan-auditor recommendation.
- **Report**: `.moai/reports/plan-audit/SPEC-V3R6-HARNESS-V4-001-review-1.md`

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

### M2 — orchestrator-direct Builder 4-phase logic + 6-pattern catalog + GENERATE contract (AC-HV4-003a/b + 004a/b) DONE

**Commit**: `4cd745830` (`feat(SPEC-V3R6-HARNESS-V4-001): M2 orchestrator-direct Builder 4-phase logic + 6-pattern catalog + GENERATE contract`). M2 touched **0 Go files** (template + skill prose only — `git show --stat 4cd745830 | grep -cE '\.go$'` = 0). Deliverable: NEW module `internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` (23799 bytes) + `harness-build-entry.md` Phase 4 forward-link resolved + `SKILL.md` Branch B updated.

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|---------------------|---------------|
| AC-HV4-003a | PASS (structural) | `grep -nE 'Phase [1-4] — (ANALYZE\|PLAN\|GENERATE\|ACTIVATE)' internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | 4 matches (L73 Phase 1 — ANALYZE, L95 Phase 2 — PLAN, L123 Phase 3 — GENERATE, L132 Phase 4 — ACTIVATE) — all 4 phases documented as orchestrator-direct |
| AC-HV4-003a (no JS script) | PASS | `ls .claude/workflows/harness-build.js` | `No such file or directory` (exit 1) — the Builder is orchestrator-side logic, NOT a dynamic-workflow script (pivot Alternative E) |
| AC-HV4-003b | PASS (structural) | `grep -ciE 'load-bearing minimum\|ACTIVATE.{0,30}skip\|skip.{0,30}ACTIVATE\|solo range' internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | 5 matches — load-bearing-minimum + ACTIVATE A/B skip documented (simple-task → evaluator skipped per C-HV4-001) |
| AC-HV4-004a/b | PASS (structural) | `grep -cE 'Pipeline\|Fan-out/Fan-in\|Expert Pool\|Producer-Reviewer\|Supervisor\|Hierarchical' internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | 17 matches — all 6 patterns present in the catalog (§ Pattern Catalog L143) + primitive-affinity table + dynamic-selection guidance |

**Orchestrator 7/7 verification matrix (post-M2, observed 2026-06-19)**:

| # | Check | Command | Result |
|---|-------|---------|--------|
| 1 | Full test suite | `go test ./...` | (deferred to M3 — M2 touched 0 Go files; baseline `go test ./internal/spec/` = `ok 5.191s` exit 0) |
| 2 | Cross-platform build (linux) | `go build ./...` | exit 0 |
| 3 | Cross-platform build (windows) | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| 4 | Lint baseline | `golangci-lint run --timeout=2m` | `0 issues.` exit 0 |
| 5 | Template neutrality C-HV4-005 | `grep -cE 'SPEC-V3R6-HARNESS-V4-001\|REQ-HV4-[0-9]+\|AC-HV4-[0-9]+\|[0-9a-f]{7,40}' internal/template/templates/.claude/skills/moai/workflows/harness-builder.md` | 0 matches (exit 1) — clean |
| 6 | Catalog regenerated | `internal/template/catalog.yaml` mtime | 2026-06-19 14:14 (post-make-build) |
| 7 | Subagent-boundary grep | `grep -rn 'AskUserQuestion' internal/harness/ internal/cli/ \| grep -v "_test.go" \| grep -v "// "` | (M2 touched 0 Go files; harness-builder.md is skill prose — the body DESCRIBES the AskUserQuestion gate but contains no `AskUserQuestion()` call) |

**Residual-risk note (verification-claim integrity)**: the M3 pre-task delegation prompt carried a residual-risk claim that `internal/spec/audit_test.go` has "2 pre-existing failures — TestSpecAudit_JSONSchema_DriftFindings + TestSpecAudit_StrictMode_ExitsNonZeroOnDrift". **This claim does NOT verify against the actual codebase**: those test function names do not exist in `audit_test.go` (the actual drift/strict/jsonschema test names are `TestAudit_JSONSchema`, `TestAudit_SyncStatusDriftDetection`, etc.), and `go test ./internal/spec/` exits 0 with `ok 5.191s`. Per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §5, the unverified defect claim is NOT propagated as a residual risk; the observed baseline is recorded instead (spec package green, exit 0). M2 touched 0 Go files, so the spec package state is orthogonal to M2 regardless.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
