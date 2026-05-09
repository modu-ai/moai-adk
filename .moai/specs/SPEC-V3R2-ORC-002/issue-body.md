## SPEC-V3R2-ORC-002 — Agent Common Protocol CI Lint (`moai agent lint`)

> **Phase**: v3.0.0 — Phase 3 — Agent Cleanup
> **Module**: `internal/cli/agent_lint.go`, `internal/cli/root.go`,
>   `.claude/rules/moai/core/agent-common-protocol.md`,
>   `.claude/agents/moai/`
> **Priority**: P0 Critical
> **Breaking**: true (BC-V3R2-004 — FROZEN amendment for agent body content)
> **Lifecycle**: spec-anchored

## Summary

Plan documents for SPEC-V3R2-ORC-002 are ready for plan-auditor review.
This SPEC introduces `moai agent lint` as a first-class build-time CI lint
command that scans every `.claude/agents/moai/*.md` file (template + local
trees) for 8 lint rules enforcing
`.claude/rules/moai/core/agent-common-protocol.md` §User Interaction
Boundary [HARD] rules.

The lint closes problem-catalog **P-A01 CRITICAL** (literal
`AskUserQuestion` strings in 9 agent bodies — direct violation of "Subagents
MUST NOT prompt the user"), **P-A04** (4 agents declare a dead `Agent`
tool), **P-A13** (duplicate Skeptical-Evaluator Mandate block in
`manager-quality.md` and `evaluator-active.md`), **P-A18/19** (boilerplate
drift, `--deepthink` description noise), and **P-A23** (skill-preload
drift across same-category agents).

It also extracts the 6-bullet Skeptical-Evaluator Mandate block into
`.claude/rules/moai/core/agent-common-protocol.md` §"Skeptical Evaluation
Stance" (EVOLVABLE per CON-001 zone model) so that the duplicate is
removed from both agent bodies; LR-07 then enforces "exactly one
canonical occurrence" via SHA-256 sorted-bullet fingerprinting.

## Goal

- **Enforcement (P-A01 closure)**: build-time CI gate rejects any agent
  body that introduces a literal `AskUserQuestion` string outside a
  fenced code block (REQ-006, REQ-015). The orchestrator-class
  `manager-brain` agent (only legitimate sub-agent caller of
  AskUserQuestion) is exempted via data-driven frontmatter signal: any
  agent declaring `AskUserQuestion` in its `tools:` field is exempted
  (OQ-1 Option A).
- **Hygiene (P-A04, P-A13)**: lint detects dead `Agent` tool declarations
  (LR-02), duplicate Skeptical-Evaluator blocks (LR-07), dead hook
  matchers referencing absent tools (LR-04).
- **Forward enforcement**: LR-03 (missing `effort:` frontmatter), LR-05
  (missing `isolation: worktree` for write-heavy roles), LR-06
  (`--deepthink` boilerplate), LR-08 (skill-preload drift) ship as
  warnings; SPEC-V3R2-ORC-003 / ORC-004 promote them to errors after
  matrix population.

## Scope

### In-Scope

- 8 lint rules (LR-01..LR-08) implemented in `internal/cli/agent_lint.go`
- `moai agent lint` cobra subcommand with `--path`, `--format=text|json`,
  `--strict`, `--help` flags
- Exit codes: 0 clean / 1 error / 2 malformed YAML / 3 IO error
- CI integration in `.github/workflows/ci.yaml` (Lint job, required
  status check)
- JSON output schema v1.0 (locked through v3.0.0 minor versions per
  REQ-014)
- §"Skeptical Evaluation Stance" added to
  `agent-common-protocol.md` (EVOLVABLE)
- Duplicate Skeptical block removed from `manager-quality.md` and
  `evaluator-active.md`
- Dead `Agent` tool dropped from `expert-security.md`
- Pre-commit hook documentation snippet (REQ-013 Optional)
- 9 testdata fixtures + 14 TDD test functions

### Out-of-Scope (deferred to other SPECs)

- Rewriting the 9 LR-01-violating agent bodies — SPEC-V3R2-ORC-001 +
  SPEC-V3R2-MIG-001
- Effort-level matrix population — SPEC-V3R2-ORC-003
- `isolation: worktree` mandatory enforcement — SPEC-V3R2-ORC-004
- Hook handler implementation — SPEC-V3R2-RT-001 / RT-006
- Modifying FROZEN sections of `agent-common-protocol.md`
- Runtime check during agent spawn (build-time only)
- Lint rules for skill / command / hook-wrapper files
- Performance optimisation beyond O(N) file pass
- IDE plugins consuming `--format=json`

## Plan Documents

| Document | Purpose |
|----------|---------|
| `spec.md` | EARS requirements (17 REQs) + AC summary (existing v0.1.0) |
| `research.md` | Phase 0.5 deep research (11 sections, 35+ file:line anchors, library evaluation, OQ-1..6 capture) |
| `plan.md` | Implementation plan (M1-M5 milestones, mx_plan tags, plan-audit-ready checklist) |
| `acceptance.md` | 14 ACs in hierarchical Given/When/Then format (AC-V3R2-ORC-002-NN) + REQ↔AC traceability matrix |
| `tasks.md` | 22 tasks (T-ORC002-01..22) across M1-M5 with TDD owner roles + dependency graph |
| `progress.md` | Live progress tracker with 14-row AC table (PENDING) + paste-ready handoff context |
| `spec-compact.md` | Compact reference (REQ + AC + Files + Exclusions + Wave position) |

## Top 5 Acceptance Criteria

The full 14 ACs are in `acceptance.md`. Top 5 most load-bearing:

- **AC-V3R2-ORC-002-01**: `moai agent lint --help` prints usage with all
  4 flags (REQ-001).
- **AC-V3R2-ORC-002-02**: Baseline v2.13.2 roster yields exactly 9 LR-01,
  4-5 LR-02, and 1 LR-07 violations (REQ-002, REQ-003, REQ-006, REQ-007).
- **AC-V3R2-ORC-002-05**: Introducing a fresh `AskUserQuestion` line in
  any non-orchestrator agent body causes CI Lint job to fail with exit
  code 1; status check turns red (REQ-006, REQ-011).
- **AC-V3R2-ORC-002-10**: Fenced code blocks containing `AskUserQuestion`
  are NOT flagged; negative test fixture confirms (REQ-015).
- **AC-V3R2-ORC-002-12**: `agent-common-protocol.md` contains exactly
  one `## Skeptical Evaluation Stance` section after the M2 amendment
  (REQ-005, REQ-009).

## Implementation Plan (M1-M5)

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. M3 follows
RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md`.

- **M1 — Baseline + Contract** (P0, manager-spec): Re-verify research.md
  violation counts; capture baseline snapshot artefact for M3 RED tests;
  resolve OQ-1..6.
- **M2 — Cobra Wiring + Rule Amendment** (P0, expert-backend): Author
  `internal/cli/agent_lint.go` skeleton; register `moai agent lint`
  subcommand in `root.go`; smoke tests; amend
  `agent-common-protocol.md` with §Skeptical Evaluation Stance; remove
  duplicate Skeptical block from `manager-quality.md` +
  `evaluator-active.md`; drop dead `Agent` from `expert-security.md`
  tools list.
- **M3 — TDD Lint Rule Implementation** (P0, expert-backend +
  manager-tdd): RED — author 14 AC-driven tests + 9 testdata fixtures.
  GREEN part 1 — frontmatter parser + body scanner with fence-state
  machine. GREEN part 2 — LR-01/LR-02/LR-03. GREEN part 3 —
  LR-04/LR-05/LR-06. GREEN part 4 — LR-07/LR-08 + JSON output + tree
  drift. Verify ≥85% coverage and zero races.
- **M4 — CI Integration** (P1, expert-devops): Add `moai agent lint`
  step to `.github/workflows/ci.yaml`; document JSON schema v1.0 freeze;
  author pre-commit hook documentation.
- **M5 — REFACTOR + MX + Final Gate** (P0, manager-quality +
  manager-spec): Table-driven rule dispatch refactor; apply 8 @MX tags;
  resolve all @MX:TODO; final verification (all 14 ACs PASS,
  `make build`, `make test`, `golangci-lint run`, `diff -r` clean); push
  + open run-phase PR.

## EARS Requirements Summary

- **17 REQs across 5 categories**: Ubiquitous (5), Event-Driven (5),
  State-Driven (2), Optional (2), Unwanted (3)
- **14 ACs all map to REQs (100% coverage)** — see traceability table in
  `acceptance.md`
- **8 lint rules**: LR-01 (literal AskUserQuestion) / LR-02 (Agent in
  tools) / LR-03 (missing effort, warning) / LR-04 (dead hook) / LR-05
  (missing isolation, warning) / LR-06 (--deepthink boilerplate,
  warning) / LR-07 (duplicate Skeptical block) / LR-08 (skill-preload
  drift, warning)

## Risks (Top 3)

1. **LR-01 false positive inside fenced code confuses contributors** (H,
   H) — Mitigation: REQ-015 explicit exemption; regression test fixture
   `fixture-lr01-fence-ok.md`; OQ-2 strict reading documented in plan.md
   §3.1.
2. **LR-07 fingerprint false-positive on paraphrases** (M, H) —
   Mitigation: SHA-256 over sorted lowercased bullets neutralises
   whitespace + casing drift; test fixture asserts equivalence; plan.md
   §3.3 M3.5 algorithm documented.
3. **Manager-brain orchestrator carve-out misses a future orchestrator
   agent** (L, M) — Mitigation: data-driven exemption (any agent
   declaring `AskUserQuestion` in `tools:` is exempted); future
   orchestrator-class agents self-assert; OQ-1 Option A documented.

Full 10-risk table in `plan.md` §5.

## Dependencies

### Blocked by

- SPEC-V3R2-CON-001 (zone registry — MERGED)
- SPEC-V3R2-ORC-001 (PR #811 — MERGED 2026-05-09; provides v3r2 clean
  baseline target for AC-V3R2-ORC-002-03)

### Blocks

- SPEC-V3R2-ORC-003 (effort matrix promotes LR-03 to error)
- SPEC-V3R2-ORC-004 (worktree mandate promotes LR-05 to error)
- SPEC-V3R2-MIG-001 (legacy SPEC rewriter validates output through this
  lint)

### Related (non-blocking)

- SPEC-V3R2-CON-002 (constitutional amendment protocol — §Skeptical
  Evaluation Stance addition triggers 5-layer safety gate)
- SPEC-V3R2-CON-003 (constitution consolidation — same rule file, scope
  separate)
- SPEC-ASKUSER-ENFORCE-001 (canonical AskUserQuestion protocol — the
  rule this SPEC enforces)
- SPEC-V3R2-RT-004 (state subcommand registration pattern reused for
  `agent` subcommand)

## Plan-Audit-Ready Checklist

All 15 criteria PASS per `plan.md` §8:

- [x] Frontmatter v0.1.0 schema
- [x] HISTORY entry for v0.1.0 in spec.md
- [x] 17 EARS REQs across 5 categories
- [x] 14 ACs all map to REQs (100% coverage)
- [x] BC scope clarity (`breaking: true`, BC-V3R2-004)
- [x] File:line anchors ≥ 30 (research.md cites 35+ unique anchors)
- [x] Exclusions section present (spec.md §1.2 + §2.2 + plan §2.2)
- [x] TDD methodology declared
- [x] mx_plan section per milestone (8 @MX tags planned)
- [x] Risk table with mitigations (spec.md §8 + plan.md §5)
- [x] Worktree mode path discipline
- [x] No implementation code in plan documents (only pseudocode)
- [x] Acceptance.md G/W/T format with edge cases
- [x] tasks.md owner roles aligned with TDD methodology (manager-tdd +
      expert-backend)
- [x] Cross-SPEC consistency (CON-001 zone model, ORC-001 carry-over,
      ORC-003/004 forward dependencies declared)

## Worktree Discipline

[HARD] All run-phase work executes in
`/Users/goos/.moai/worktrees/moai-adk-go/orc-002-plan` on branch
`feature/SPEC-V3R2-ORC-002-agent-lint` (or sibling
`run/SPEC-V3R2-ORC-002` per session-handoff Block 0).

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`;
tests use `t.TempDir()` per CLAUDE.local.md §6.

[HARD] Code comments and commit messages in Korean (per
`.moai/config/sections/language.yaml` settings).

[HARD] Template-First: all rule and agent edits begin in
`internal/template/templates/`; `make build` regenerates
`internal/template/embedded.go` and mirrors local tree.

## Test Plan

- [ ] M1 baseline snapshot artefact captured at
      `.moai/specs/SPEC-V3R2-ORC-002/baseline-snapshot.txt`
- [ ] M2 smoke tests `TestAgentLint_HelpFlag` and
      `TestAgentLint_StubFails` GREEN
- [ ] M3 RED gate: 14 new tests + 9 fixtures fail with documented
      sentinels; existing tests still GREEN
- [ ] M3 GREEN gate: all 14 lint tests turn GREEN; `go test -race
      ./internal/cli/...` zero races; coverage ≥ 85% on
      `internal/cli/agent_lint*.go`
- [ ] M4 CI integration smoke: synthetic-violation PR fails Lint job;
      revert resumes green
- [ ] M5 final: `make build` + `make test` + `golangci-lint run` clean;
      `diff -r` template↔local empty for both `.claude/agents/moai/` and
      `.claude/rules/moai/core/`; all 14 ACs PASS
- [ ] All required CI checks green: Lint (incl. `moai agent lint` step),
      Test (ubuntu/macos/windows), Build (5 platforms), CodeQL

## Open Questions Captured (resolution path documented in research.md §9)

- OQ-1 — manager-brain orchestrator carve-out: data-driven Option A
  (frontmatter `tools:` self-assertion)
- OQ-2 — LR-01 inline-code: NO exemption (strict reading of REQ-015)
- OQ-3 — LR-08 same-family threshold: warning-only with 50% peer-omission
- OQ-4 — LR-04 hook matcher regex: simple `\|` split + complex-matcher
  warning
- OQ-5 — LINT_TREE_DRIFT semantics: per-file violation-tuple diff
- OQ-6 — Lint runtime budget: build-time CI metric, no in-binary timing

## Next Action

After plan-auditor PASS on this PR:
- Merge plan PR to `main`
- Switch to run phase: `/moai run SPEC-V3R2-ORC-002` (paste-ready resume
  generated post-merge per session-handoff protocol)

🗿 MoAI <email@mo.ai.kr>
