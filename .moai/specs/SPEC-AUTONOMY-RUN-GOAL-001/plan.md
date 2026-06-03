# Implementation Plan — SPEC-AUTONOMY-RUN-GOAL-001

> Tier M. Doctrine/rules-focused, low Go-code volume. Milestones are manager-develop-sized units. No time estimates (priority/order only).

## §A — Context

- **Work location**: `/Users/goos/MoAI/moai-adk-go/`
- **SPEC artifacts**: `.moai/specs/SPEC-AUTONOMY-RUN-GOAL-001/{spec,plan,acceptance,design}.md`
- **Primary edit targets**:
  - `.claude/rules/moai/workflow/orchestration-mode-selection.md` (D1 — Mode 6 addition)
  - `.claude/skills/moai/workflows/run.md` (D2 — `/goal` `ac_converge` wiring point + GATE-2 ordering)
  - A regression test/guard (D3 — GATE-2 preservation), expressed as a grep/rule assertion against the run skill body + a doctrine cross-reference assertion
- **Template mirrors** (CLAUDE.local.md §2): `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md`, `internal/template/templates/.claude/skills/moai/workflows/run.md`
- **PRESERVE**: do NOT rewrite the legacy run sub-skill agent chain (`phase-execution.md` archived-agent references) — EX-5. Do NOT touch sibling SPEC scope (CONFIG/PATTERNS/GOAL-CONDITIONS) — EX-1/EX-2/EX-3.

## §B — Known Issues (filtered for this SPEC's domain)

- **B4 (Frontmatter schema)**: spec.md uses canonical `created`/`updated`/`tags` — verified at plan-authoring time.
- **B6 (spec-lint heading)**: `## §D — Exclusions (What NOT to Build)` satisfies the MissingExclusions check; verify no `MissingExclusions` finding at run-phase.
- **B10 (Untouched-paths PRESERVE)**: parallel sessions may touch `orchestration-mode-selection.md` or run skill — pre-spawn `git fetch origin main` + `git rev-list --count` (agent-common-protocol.md § Pre-Spawn Sync Check) before any write.
- **B11 (AskUserQuestion boundary)**: this SPEC's deliverables ADD AskUserQuestion-ordering doctrine — when editing rule/skill bodies, do not introduce a literal `AskUserQuestion(...)` call inside a subagent-domain file; the GATE-2 gate is an orchestrator responsibility described in prose.
- **B-template (mirror drift)**: every `.claude/` edit needs the `internal/template/templates/` mirror + `make build`; `TestRuleTemplateMirrorDrift` / template neutrality tests will fail otherwise. NOTE: the Mode 6 doctrine text is generic (no internal SPEC IDs in the template copy per CLAUDE.local.md §25 / template-internal-isolation) — keep SPEC-ID/REQ tokens out of the mirrored template content.

## §C — Pre-flight (run-phase, before code change)

```bash
# 1. baseline branch + HEAD
git branch --show-current && git rev-parse HEAD
# 2. pre-spawn sync (multi-session race mitigation)
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# 3. spec-lint baseline for this SPEC
moai spec lint .moai/specs/SPEC-AUTONOMY-RUN-GOAL-001/ 2>&1 | tail -10 || echo "spec lint baseline"
# 4. confirm canonical mode catalog is 5 modes (D1 precondition)
grep -c "Mode" .claude/rules/moai/workflow/orchestration-mode-selection.md
# 5. confirm run skill GATE-2 / mode wiring point exists
grep -n "GATE-2\|Phase 0.95\|AskUserQuestion" .claude/skills/moai/workflows/run.md || echo "wiring point absent — add"
```

## §D — Constraints (DO NOT VIOLATE)

- Mode 6 is appended; Modes 1–5 are NOT removed or renumbered (REQ-ARG-001).
- The `ac_converge` condition predicates are transcript-measurable, never file-path (REQ-ARG-007).
- GATE-2 `AskUserQuestion` ordering is emitted BEFORE any `/goal` set in the run skill body (REQ-ARG-006, REQ-ARG-013).
- No typed/named Workflow script API asserted anywhere (REQ-ARG-014 / EX-6) — conceptual coordinate-agents model only.
- `autonomy.enabled` stays off by default — no default-on flip (EX-7).
- Every `.claude/` edit mirrored to `internal/template/templates/` + `make build` (REQ-ARG-016).
- Template-copy content stays internal-content-neutral (no SPEC ID/REQ token leakage per CLAUDE.local.md §25).

## §E — Self-Verification Deliverables (manager-develop completion report MUST include)

- AC binary PASS/FAIL matrix (AC-ARG-001..AC-ARG-013) with verification command + actual output per AC.
- `moai spec lint` clean for this SPEC (no MissingExclusions, no FrontmatterInvalid).
- Mirror-drift test green: `go test ./internal/template/... -run 'TestRuleTemplateMirror|TestTemplateNeutralityAudit'`.
- Grep evidence for D3: GATE-2 `AskUserQuestion` appears before `/goal` token in run skill body; CLAUDE.local.md §19.1 / REQ-ATR-015 cross-reference present.
- `make build` exit 0; embedded template regenerated.

## §F — Milestones (priority-ordered)

### M1 — D1: Mode 6 (workflow) catalog addition (Priority High)
**Owner**: manager-develop (cycle_type per quality.yaml; doctrine-edit profile)
**Scope**: `.claude/rules/moai/workflow/orchestration-mode-selection.md`
- Append Mode 6 (`workflow`) row to §A Mode Catalog (concurrency = Workflow fan-out 16-concurrent/1000-total backstop; spawn surface = orchestrator-launched Workflow, NOT nested).
- Add a §B decision-tree branch: after the Mode 4 (parallel) check, a Mode 6 candidate branch gated on `scope ≥ ~30 files AND mechanical AND genuinely-parallel`; coding-heavy/multi-domain falls through to Mode 5 (Finding A4).
- Add a §C.3 capability gate for Mode 6: GATE-2-already-passed precondition + preferences-collected precondition + Mode 6 is scaling-not-nesting (Finding A1).
- Add §E anti-pattern entries: Mode 6 for coding-heavy work (AP — Finding A4); Mode 6 launch before GATE-2 (AP — C1); asserting a named Workflow script API (AP — EX-6).
- Covers: REQ-ARG-001, -002, -003, -004, -005, -014.

### M2 — D2: Run-phase `/goal` `ac_converge` wiring (Priority High)
**Owner**: manager-develop
**Scope**: `.claude/skills/moai/workflows/run.md` (router body — add a `## Run-phase Autonomy (/goal ac_converge)` section)
- Document the wiring point: AFTER GATE-2 approval, the orchestrator MAY set the `ac_converge` `/goal` with the inline condition (verbatim from design §C / strategy §5.2), bounded `max 20 turns`, transcript-measurable predicates only.
- Document the semantic-failure escape: data race / deadlock / panic / test assertion → clear goal + `AskUserQuestion` escalation (REQ-ARG-009).
- Document the `/goal` non-substitution clause: the goal never bypasses GATE-2, PR creation, or destructive ops (REQ-ARG-010).
- Cross-reference `goal-directive.md` and `orchestration-mode-selection.md` (do not restate).
- Covers: REQ-ARG-006, -007, -008, -009, -010.

### M3 — D3: GATE-2 preservation regression guard (Priority High)
**Owner**: manager-develop
**Scope**: a regression test/guard expressed as a grep/rule assertion (no heavy Go logic). Two complementary checks:
- **Check A (ordering)**: assert the run skill body emits the GATE-2 `AskUserQuestion` human-gate text BEFORE the first `/goal` token (the `ac_converge` set). Implementable as a Go test in `internal/template/` that reads the run skill body and asserts the byte-offset of the GATE-2 marker precedes the `/goal ac_converge` marker, OR a `spec-lint`/audit-style assertion.
- **Check B (score-independence + cross-ref)**: assert the run skill body contains a statement that GATE-2 is emitted regardless of plan-auditor score (incl. ≥0.90 skip-eligible), AND contains a cross-reference to CLAUDE.local.md §19.1 / REQ-ATR-015.
- Covers: REQ-ARG-011, -012, -013.

### M4 — Template mirror + build (Priority High — gating)
**Owner**: manager-develop
**Scope**: `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md`, `internal/template/templates/.claude/skills/moai/workflows/run.md`
- Mirror M1 + M2 edits into the template source (internal-content-neutral copy — strip SPEC-ID/REQ tokens per CLAUDE.local.md §25).
- `make build` to regenerate `internal/template/embedded.go`.
- Run `go test ./internal/template/... -run 'TestRuleTemplateMirror|TestTemplateNeutralityAudit|TestSettingsTemplateHookEventCount'` — green.
- Covers: REQ-ARG-016.

### M5 — Verification + spec-lint + status transition (Priority Medium)
**Owner**: manager-develop (draft → in-progress on M1 commit; the in-progress → implemented transition is manager-docs at sync-phase)
**Scope**: run-phase verification batch
- `go test ./...` exit 0; `moai spec lint` clean for this SPEC.
- AC matrix (AC-ARG-001..013) PASS evidence surfaced.
- `golangci-lint run` no NEW findings.
- Covers: all REQ verification.

## §G — Anti-Patterns (avoid during implementation)

- Rewriting the legacy run sub-skill agent chain (EX-5) — out of scope, churn risk.
- Pulling sibling SPEC scope (workflow.yaml struct, full pattern catalog, full condition registry) into this SPEC.
- Writing a file-path predicate into the `ac_converge` condition (the `/goal` evaluator cannot read files — C5/REQ-ARG-007).
- Asserting a named Workflow script API in any delivered text (EX-6).
- Leaking internal SPEC IDs / REQ tokens into the `internal/template/templates/` mirror copy (CLAUDE.local.md §25).
- Adding a literal `AskUserQuestion(...)` invocation inside a rule/skill body as if a subagent calls it — GATE-2 is the orchestrator's responsibility, described in prose.

## §H — Cross-References

- `.moai/docs/autonomous-workflow-strategy.md` §4.2-B, §5.2, §6.2, §8.1, §9 (roadmap entry SPEC-AUTONOMY-RUN-GOAL)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (D1 target — §A catalog, §B tree, §C gates, §E anti-patterns)
- `.claude/rules/moai/workflow/goal-directive.md` (D2 — `/goal` semantics, cite not restate)
- `.claude/rules/moai/workflow/dynamic-workflows.md` (Mode 6 primitive, named-API prohibition source)
- CLAUDE.local.md §19.1 (REQ-ATR-015 GATE-2 restoration — D3 cross-ref target)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (delegation template for run-phase)
- CLAUDE.local.md §2 (Template-First), §25 (template internal-content isolation)
