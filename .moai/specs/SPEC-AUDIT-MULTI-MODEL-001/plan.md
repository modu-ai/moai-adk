# Plan — SPEC-AUDIT-MULTI-MODEL-001

> Implementation plan, milestone decomposition, technical approach, risks. Ordered by decision-reversibility — the decisions most likely to change (data model, convergence policy, MCP tool shape) lead; the mechanical steps (skill authoring, CI guard, mirror) trail. The fixed user decisions (Tier L single SPEC; disagreement = advisory NOT block) are baked into M2 and C3 and are NOT re-litigated here.

## §A. Context

- **Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/audit-multi-model` (branch `feat/spec-audit-multi-model`, HEAD `f85ff4c3e` at plan-start).
- **SPEC artifacts**: `.moai/specs/SPEC-AUDIT-MULTI-MODEL-001/{spec,plan,acceptance,design,research,progress}.md`.
- **Hard dependency (completed)**: SPEC-MOAI-MCP-SERVER-001 (PR #1378) — the stdio server, `codex_audit`, GLM audit, `audit_model`/`audit_gate` enums, `VerdictInconclusive`, `ResolveAgentModelEffort` SSOT, codex-review-gate Stop hook all shipped. This plan consumes them; the only edit to that surface is the `multiConvergenceImplemented` sentinel flip (`internal/cli/mcp_audit.go:31`).
- **Cycle type**: `tdd` (this is new code — convergence engine, MCP tool, Skill; no legacy preservation needed). RED-GREEN-REFACTOR per the quality.yaml default.
- **Tier**: L → 5-artifact plan set + `progress.md`; plan-auditor PASS threshold 0.85.

## §B. Known Issues (auto-injected, per `manager-develop-prompt-template.md` §B)

- **B1 (cross-platform build tags)**: the convergence engine is pure Go (no syscall); `GOOS=windows GOARCH=amd64 go build ./...` MUST still pass after the changes. The Stop-hook handler reuses the existing `internal/cli/codex_review_gate.go` pattern which already handles cross-platform concerns.
- **B2 (cross-SPEC policy conflict pre-scan)**: `grep -r "multiConvergenceImplemented\|multi.*convergence\|SPEC-AUDIT-MULTI" internal/` — the only hit MUST be the sentinel at `mcp_audit.go:31` plus this SPEC's artifacts; no stray caller. Verify before M1.
- **B3 (C-HRA-008 / subagent boundary)**: the convergence engine + `audit_multi` handler + multi-review-gate handler MUST NOT call `AskUserQuestion` or `mcp__askuser__*` (REQ-AMM-018). Static guard: a grep-based test `TestConvergence_NoAskUserQuestion` asserting zero matches in the new files (mirrors `TestWeb_NoAskUserQuestion`).
- **B4 (frontmatter canonical schema)**: this SPEC uses the canonical 12 fields; the `tier: L` optional field is set.
- **B5 (CI 3-tier awareness)**: spec-lint, golangci-lint, Test (per OS) can each fail separately. The §25 CI guard (`template-neutrality-check.yaml` / `internal_content_leak_test.go`) is a fourth gate for the M6 Skill mirror — all four MUST be green.
- **B6 (spec-lint heading convention)**: the `### Out of Scope — <topic>` H3 sub-headings in `spec.md` are required by the `OutOfScopeRule` lint; do NOT collapse them to a bare `## Out of Scope` H2.
- **B7 (observer.go / capture path resolution)**: not directly relevant — the convergence engine does not capture agent I/O. Relevant only if the multi-review-gate handler needs to observe the previous turn (it reuses the codex-review-gate's change-detector seam — `withChangeDetector` — which already handles this).
- **B8 (working tree hygiene)**: do NOT modify `.moai/state/`, `.moai/harness/usage-log.jsonl`, or runtime-managed files. The convergence result is written to a state file ONLY if M5 (multi-review-gate) needs to read it across turns — and even then it lives under `.moai/state/audit-multi/` (session-scoped, gitignored).
- **B9 (Git commit + push)**: per CLAUDE.local.md §23 (PR-mandatory — `enforce_admins: true`); all changes route via PR. Conventional Commits (`feat(SPEC-AUDIT-MULTI-MODEL-001): M{N} <subject>`).
- **B10 (untouched paths PRESERVE)**: do NOT touch the single-backend `mcp_codex.go` / `mcp_glm.go` internals, the codex-review-gate handler logic, the `config.AuditModel*` / `AuditGate*` enum definitions, or MOAI-MCP-SERVER's SPEC artifacts. The only edit to `mcp_audit.go` is the sentinel flip (line 31) + the doc-comment update.
- **B11 (AskUserQuestion prohibited)**: same as B3 — see REQ-AMM-018.
- **B12 (CHANGELOG)**: manager-docs-owned at sync-phase; not a plan-phase concern.

## §C. Pre-flight Check List (manager-develop runs before any code change)

```bash
# 1. Branch + baseline
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/audit-multi-model branch --show-current
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/audit-multi-model rev-parse HEAD

# 2. Cross-platform build feasibility
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/audit-multi-model && go build ./...
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/audit-multi-model && GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline (distinguish NEW vs pre-existing)
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/audit-multi-model && golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Re-verify the deferral sentinel + the integration points from research.md
grep -n 'multiConvergenceImplemented' internal/cli/mcp_audit.go
grep -n 'VerdictInconclusive' internal/cli/mcp_codex.go
grep -n 'func ResolveAgentModelEffort' internal/template/profile_matrix.go
grep -n 'AuditModelMulti\|AuditGateRequired' internal/config/defaults.go

# 5. Cross-SPEC conflict pre-scan
grep -rn 'multiConvergenceImplemented\|multi.*convergence' internal/ 2>/dev/null | grep -v _test.go
```

## §D. Constraints (DO NOT VIOLATE)

- **C1–C9** as in `spec.md` §E (additive to MOAI-MCP-SERVER; fail-open identity; disagreement = advisory NOT block; super-review independence; subagent boundary; ResolveAgentModelEffort SSOT; secret hygiene; opt-in default-off; Template-First + §25 neutrality).
- **PRESERVE targets (do NOT modify body content)**:
  - `internal/cli/mcp_server.go` (the stdio server — only REGISTER the new `audit_multi` tool here)
  - `internal/cli/mcp_codex.go` (codex_audit handler internals — the convergence engine CALLS it, does not fork it)
  - `internal/cli/mcp_glm.go` (GLM audit handler internals — same)
  - `internal/cli/codex_review_gate.go` (the self-gate logic — the multi-review-gate reuses the same `withChangeDetector` seam in a sibling file)
  - `internal/config/defaults.go` lines 637-641 (the default `AuditConfig` — convergence reads it, does not change it)
  - MOAI-MCP-SERVER SPEC artifacts (`.moai/specs/SPEC-MOAI-MCP-SERVER-001/*`) — frozen.
- **Forbidden commands**: `--no-verify`, `--amend` on shared branches, force-push to main, direct push to main (PR-mandatory per §23).
- **Required**: Conventional Commits, `🗿 MoAI` trailer, full `go test ./...` (not affected-package-only — per the trust-but-verify lesson).
- **Binary constraints**:
  - `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_convergence.go internal/cli/mcp_multi_review_gate.go 2>/dev/null` → 0 matches (C-HRA-008).
  - `grep -rn 'VerdictDisagreement' internal/ 2>/dev/null` → 0 matches (REQ-AMM-008 — disagreement is a boolean flag, not an enum value).
  - `grep -n 'claude_verdict\|ClaudeVerdict' internal/cli/mcp_codex.go internal/cli/mcp_glm.go` → 0 matches in the backend-call argument paths (independence preservation — the `claude_verdict` is consumed only in the synthesis step, never passed to the secondary backends).

## §E. Self-Verification Deliverables (manager-develop reports at run-phase completion)

Each E-item per the verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).

- **E1** — AC binary PASS/FAIL matrix (AC-AMM-001..025).
- **E2** — cross-platform build result (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`).
- **E3** — coverage measurement (≥85% per package for the new `mcp_convergence.go` + `mcp_multi_review_gate.go`).
- **E4** — subagent-boundary grep (`AskUserQuestion` / `mcp__askuser` in new files → 0).
- **E5** — lint status (NEW vs baseline separated).
- **E6** — branch HEAD + push state.
- **E7** — blocker report (if any — structured, never free-form questions).
- **E8** — RED failure output (TDD — verbatim pre-GREEN evidence for the convergence algorithm tests, the independence-preservation test, and the disagreement-policy test).

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

> The data-model + convergence-policy decisions lead (M0, M1, M2) because they are the most likely to be revisited; the mechanical skill-authoring + mirror steps trail (M5, M6) because they are reversible edits. The fixed user decisions (Tier L single SPEC; disagreement = advisory) are baked in at M0/M2 and are NOT re-litigated.

### M0 — Design lock (Priority High — the data model + algorithm)

**Decisions to lock**: (a) `ConvergenceResult` shape (`per_backend_verdicts[]`, `overall_verdict`, `disagreement_flag`, `residual_risk_note`); (b) the 4-step convergence algorithm (REQ-AMM-006 #1–#4); (c) the independence-preservation contract (`claude_verdict` is synthesis-only, never secondary-backend prompt context); (d) the disagreement = advisory policy (NOT block — user decision); (e) the sentinel-flip rule (the `multiConvergenceImplemented` flip lands in the SAME commit as the engine — Definition of Done sentinel-flip same-commit invariant).

**Deliverable**: `design.md` §1 (data model), §2 (parallel-execution model), §3 (convergence algorithm + Verification Matrix integration), §4 (Stop-hook gate extension); `progress.md` §E.1 records M0-complete.

**No code change.** Output is locked design decisions referenced from `progress.md` §E.2 at run-phase.

### M1 — Convergence engine core + sentinel flip (Priority High — the critical-path code)

**Why this is decision-reversible-adjacent**: the engine's internal struct shapes and the parallel-execution model (`errgroup` vs hand-rolled goroutines) are the most likely to be revisited once we see the test fixtures.

**Deliverable**:
- NEW file `internal/cli/mcp_convergence.go` — the synthesis function + the parallel fan-out (using `errgroup` — already in go.mod).
- Flip `multiConvergenceImplemented` from `false` to `true` in `internal/cli/mcp_audit.go:31` IN THE SAME COMMIT as the engine (Definition of Done sentinel-flip same-commit invariant — no window where the sentinel lies).
- Update the doc-comment at `mcp_audit.go:27-31` to reflect that convergence is now implemented (the `multiConvergenceImplemented` constant MAY be retained as documentation, OR removed if the linter flags it as unused — run-phase decision).
- TDD: RED test first — `TestConvergence_AllRequiredPass`, `TestConvergence_AnyRequiredFail`, `TestConvergence_RequiredDisagreement`, `TestConvergence_FailOpenOnMissingBackend`, `TestConvergence_AdvisoryNeverFlipsToFail`, `TestConvergence_IndependenceClaudeVerdictNotInSecondaryPayload`, `TestConvergence_NoVerdictDisagreementEnum`.

**Critical-path status**: M1 is the critical path — M2 (algorithm tests), M3 (`audit_multi` tool), M4 (cross-model wiring), M5 (Stop hook) ALL depend on the engine existing.

### M2 — Convergence algorithm regression tests + disagreement-policy enforcement (Priority High)

**Deliverable**:
- The algorithm tests from M1 are expanded to cover EC-1..EC-7 edge cases.
- A dedicated `TestDisagreementPolicy_AdvisoryNotBlock` regression test asserting that `disagreement_flag == true` does NOT produce a hard block when the required-gate contract is otherwise satisfied (EC-3) — this test is the SPEC's enforcement of the fixed user decision (disagreement = advisory).
- A dedicated `TestNoVerdictDisagreementEnum` test asserting no new enum value exists (EC-7).

**Why it sits here**: M2 MUST land before M3 (the MCP tool wraps the algorithm) and before M5 (the Stop hook reads the algorithm output). The disagreement-policy test is the single most load-bearing regression guard for the user decision.

### M3 — `audit_multi` MCP tool surface (Priority Medium — the user-facing API)

**Deliverable**:
- NEW handler `internal/cli/mcp_multi.go` (or extend `mcp_audit.go`) — the `audit_multi` tool: reads config, fans out via the engine, returns `ConvergenceResult`.
- Register `audit_multi` in `mcp_server.go`'s tool-registration block (the only edit to `mcp_server.go` — additive).
- JSON Schema declaration for `audit_multi` inputs + output.
- TDD: RED test `TestAuditMulti_ThinWrapper` asserting it calls the existing `codex_audit` / GLM-audit handlers (no re-implementation), `TestAuditMulti_RespectsAuditGateOff` asserting a backend with `gate == off` is not invoked.

### M4 — plan-audit + sync-audit cross-model wiring (Path A — Priority Medium)

**Deliverable**:
- NEW Skill `moai-ref-cross-model-audit` under `.claude/skills/moai-ref-cross-model-audit/SKILL.md` — the skill body instructs the auditor agent to call `mcp__moai__audit_multi` with its in-session claude analysis as `claude_verdict`, and to fold the `ConvergenceResult` into its audit verdict + Verification Matrix residual-risk surface.
- The skill body states the independence-preservation rule verbatim (pass only the synthesized `claude_verdict`, not the full analysis).
- No change to the plan-auditor / sync-auditor agent files — the skill is loaded via the EXISTING `skill-routing.md` protocol (the agents already carry `Skill` in `tools:`).

**Why it sits here**: Path A depends on M3 (the tool must exist for the skill to call it). The skill is an evolvable asset; its prompt-content is out of scope for the SPEC requirements (§B Out of Scope).

### M5 — Fully-autonomous goal convergence gate (Path C — Priority Medium)

**Deliverable**:
- NEW file `internal/cli/multi_review_gate.go` — the Stop-hook handler, reusing the `withChangeDetector` seam from `codex_review_gate.go` (no new heuristic).
- NEW subcommand wiring `moai hook multi-review-gate` (registered under the `hook` command tree, sibling to `codex-review-gate`).
- opt-in via `workflow.multi_review_gate.enabled` (BranchGuard pattern — the config block is a sibling of `workflow.codex.review_gate`, no new schema shape).
- 900 s timeout override (same override mechanism as codex-review-gate).
- The self-gate logic (no-edit turn → ALLOW immediately) is REUSED verbatim from `codex_review_gate.go`.
- TDD: RED test `TestMultiReviewGate_SelfGateNoEditTurnAllow`, `TestMultiReviewGate_AllRequiredPassAllow`, `TestMultiReviewGate_AnyRequiredFailBlock`, `TestMultiReviewGate_AdvisoryDisagreementNoBlock`, `TestMultiReviewGate_FailOpenToClaude`.

### M6 — Skill authoring under template source + `make build` + §25 neutrality (Priority Low — mechanical, reversible)

**Deliverable**:
- Move the M4 skill under template source: `internal/template/templates/.claude/skills/moai-ref-cross-model-audit/`.
- `make build` regeneration (templates are embedded via `//go:embed all:templates`).
- §25 neutrality CI guard green (`template-neutrality-check.yaml` + `internal_content_leak_test.go`).
- Verify the skill content passes §25 (no SPEC-ID, no commit SHA, no internal date, no macOS-bias path).

**Why it trails**: this is the mechanical Template-First mirror; the skill content was authored in M4 and is reversible.

### M7 — Hardcoding prevention + full-suite green (Priority Low — closure)

**Deliverable**:
- New env-var names → `internal/config/envkeys.go` constants.
- New thresholds → `internal/config/defaults.go`.
- `multi_review_gate` config block → sibling of `workflow.codex.review_gate`.
- `go test ./...` + `go vet ./...` + §25 CI guard green.

## §G. Anti-Patterns

- **AP-AMM-1**: re-implementing the single-backend `codex_audit` / GLM-audit internals inside the convergence engine (violates C1 — the engine CALLS the existing handlers, it does not fork them).
- **AP-AMM-2**: passing `claude_verdict` into the codex/glm prompt context (violates REQ-AMM-003 / C4 — super-review independence destroyed).
- **AP-AMM-3**: inventing a `VerdictDisagreement` enum value (violates REQ-AMM-008 — disagreement is a boolean flag).
- **AP-AMM-4**: flipping the `multiConvergenceImplemented` sentinel in a commit WITHOUT the engine (violates the Definition of Done sentinel-flip same-commit invariant — the sentinel would lie).
- **AP-AMM-5**: promoting cross-model disagreement to a BLOCK on the autonomous flow (violates REQ-AMM-007 / C3 — the fixed user decision).
- **AP-AMM-6**: hardcoding the MCP tool name as a prose-only string in the skill body (violates REQ-AMM-017 — use the canonical `mcp__moai__audit_multi` so a future rename propagates via grep).
- **AP-AMM-7**: running the convergence backends sequentially rather than in parallel (defeats the wall-clock benefit — verifiable by timing test AC-AMM-002).
- **AP-AMM-8**: adding a new `VerdictDisagreement` or new `audit_model` enum token (the `multi` token already exists; the engine activates it — no new enum).
- **AP-AMM-9**: authoring the skill under local `.claude/skills/` without mirroring to `internal/template/templates/` (violates C9 / REQ-AMM-016 — Template-First).
- **AP-AMM-10**: calling `AskUserQuestion` from the convergence engine, the `audit_multi` handler, or the multi-review-gate handler (violates REQ-AMM-018 / C5 — subagent boundary).

## §H. Cross-References

- `spec.md` — REQ-AMM-001..019, §C (verbatim deferral quote), §E (constraints C1-C9), §G (risks R1-R5).
- `acceptance.md` — AC-AMM-001..025 + EC-1..EC-7 + Definition of Done.
- `design.md` — convergence data model (§1), parallel-execution model (§2), convergence algorithm + Verification Matrix integration (§3), Stop-hook gate extension (§4), independence-preservation mechanism (§5).
- `research.md` — super-review pattern [R3], AgentOrchestra peer verification [R5], cross-model adversarial review literature, verified integration-point inventory (re-verified at `f85ff4c3e`).
- `progress.md` — §E.1 (plan-phase audit-ready), §E.2-§E.4 (placeholder skeletons for run/sync-phase).
- Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4, §3.6 v3 extension, Q1 routing paths.
- Hard dependency: SPEC-MOAI-MCP-SERVER-001 (completed, PR #1378).
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- Delegation template: `.claude/rules/moai/development/manager-develop-prompt-template.md` (Tier L → full Section A-E required).
