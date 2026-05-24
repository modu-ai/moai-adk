---
id: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001
title: "V3R4 Self-Evolving Harness Loop Closure — Implementation Plan"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Self-Evolution Closure"
module: "internal/harness/proposalgen, internal/cli/harness, .moai/proposals/"
lifecycle: spec-anchored
tags: "harness, proposal, plan, tier-m, v3r6"
---

# Plan — SPEC-V3R6-HARNESS-PROPOSAL-GEN-001

Tier M implementation plan following the manager-develop-prompt-template.md Section A-E structure (REQUIRED for Tier M per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability).

## §A. Context

### A.1 Predecessor State

SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (HCW-001) was merged into `main` at commit `577f10308` on 2026-05-24, closing the **runtime → classifier → tier-promotions.jsonl** half of the V3R4 self-evolving harness learning loop. The classifier's output (`.moai/harness/learning-history/tier-promotions.jsonl`) is now produced live by `moai hook harness-classify` invoked from the workflow body §2.1 step 0 Bash hook.

Live data verified at plan creation (2026-05-24): 8 records, 4 unique `pattern_key` values, ALL system hook events (`agent_invocation:Bash:`, `subagent_stop:unknown:`, `session_stop::`, `user_prompt::`). See spec.md §1.2 for the critical scope clarification.

### A.2 Source Data Limitation Summary

[HARD] With current data, the proposal generator MUST produce **zero actionable SPEC proposals**. The generator's value is "future-data ready" — it materializes when richer learning sources (code-change, error-pattern, tool-failure) feed `tier-promotions.jsonl`. AC-PGN-002 and AC-PGN-004 explicitly verify the no-op behavior with current data. The mapper's actionable-pattern regex (REQ-PGN-004) excludes all 4 current system-event prefixes by design.

### A.3 User-Confirmed Decisions (Verbatim Injection)

The orchestrator confirmed the following 4 decisions via AskUserQuestion before plan-phase delegation:

1. **Delegation mode**: `auto-delegate + AskUserQuestion gate` (3 options: Approve / Modify / Reject). On Approve, the orchestrator spawns manager-spec with the proposal injected. On Modify, user edits via the Other free-form option and the gate re-emits. On Reject, the orchestrator writes `.moai/proposals/<draft-id>/REJECTED.md`. The CLI itself MUST NOT call AskUserQuestion (subagent boundary HARD).
2. **Git env**: `main direct push` (Hybrid Trunk per CLAUDE.local.md §23.7). No plan/* or feat/* branches. No worktree. pre-push hook 5s warn-only + CI 4-status-checks remain protective.
3. **Tier**: M (Medium). Estimated scope ~1200 LOC, 8-11 files, 3 artifacts (spec.md + plan.md + acceptance.md). plan-auditor PASS threshold 0.80. Section A-E delegation template REQUIRED. TDD methodology per quality.yaml. Coverage target 85%.
4. **16-language neutrality**: All proposal templates and generator output MUST be language-neutral. No Go-specific assumptions in `.moai/proposals/<draft-id>/spec.md` body.

### A.4 PRESERVE List

The run-phase manager-develop MUST NOT modify the following:

- `internal/harness/applier.go`, `internal/harness/classifier.go`, `internal/harness/layer1.go..layer5.go`, `internal/harness/evaluator_leak.go`, `internal/harness/harness_autonomy_integration_test.go` (HCW-001 / V3R4-002 / V3R4-003 territory — read-only consumption of `Promotion` struct only)
- `.moai/harness/learning-history/tier-promotions.jsonl` (read-only input)
- `.moai/harness/usage-log.jsonl`, `.moai/harness/observations.yaml` (out of SPEC scope per §7.1)
- `internal/cli/hook.go` (HCW-001 `moai hook harness-classify` subcommand — leave intact)
- `internal/cli/worktree/new_test.go` (subagent boundary precedent — read-only mirror reference)
- `.moai/config/sections/harness.yaml` (read-only configuration reference)
- `internal/cli/harness.go` (existing harness CLI verb implementations — `status` / `apply` / `rollback` / `disable`; read-only reference for the propose subcommand to mirror style)
- `internal/cli/harness_route.go` (existing `newHarnessRouterCmd()` parent registered at `internal/cli/root.go:104`; this file is MODIFIED in §A.5 with exactly one line — `cmd.AddCommand(newHarnessProposeCmd())` — inside the existing AddCommand cluster; no other modifications permitted)
- `internal/cli/root.go` (existing `rootCmd.AddCommand(newHarnessRouterCmd())` at line 104; no changes — `propose` is registered as a subcommand under the existing harness parent, NOT as a new top-level command)

### A.5 EXTEND List

The run-phase manager-develop MAY create/modify the following:

- NEW: `internal/harness/proposalgen/reader.go` + `reader_test.go` — JSONL reader binding `internal/harness.Promotion`
- NEW: `internal/harness/proposalgen/mapper.go` + `mapper_test.go` — promotion → proposal candidate mapping
- NEW: `internal/harness/proposalgen/scaffolder.go` + `scaffolder_test.go` — `.moai/proposals/<draft-id>/` writer
- NEW: `internal/harness/proposalgen/types.go` — internal struct definitions (ProposalCandidate, GeneratorResult)
- NEW: `internal/cli/harness/propose.go` + `propose_test.go` — `moai harness propose` subcommand factory (`newHarnessProposeCmd() *cobra.Command`); this is a NEW subpackage containing only the factory function, NOT a parent cobra.Command. The factory is invoked from `internal/cli/harness_route.go` (existing parent) — no duplicate parent created.
- MODIFY: `internal/cli/harness_route.go` — exactly one line added: `cmd.AddCommand(newHarnessProposeCmd())` inside the existing `newHarnessRouterCmd()` body, in the AddCommand cluster alongside the existing `newHarnessStatusCmd()` / `newHarnessApplyCmd()` / `newHarnessRollbackCmd()` / `newHarnessDisableCmd()` siblings (~lines 36-39). No other modifications to this file permitted.
- NEW (test only): `internal/cli/harness/propose_boundary_test.go` — `TestPropose_NoAskUserQuestion` static check (REQ-PGN-013)
- NEW (CHANGELOG): `CHANGELOG.md` `[Unreleased]` entry (sync-phase responsibility, NOT run-phase)
- MODIFY: `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/progress.md` — run-phase status update (manager-develop owns frontmatter `status: draft → in-progress`)

## §B. Deliverables

Estimated file list (7 NEW + 1 MODIFY = 8 files; ~1351 LOC including tests):

| Path | Role | Est. LOC |
|------|------|----------|
| `internal/harness/proposalgen/types.go` | Internal struct defs (ProposalCandidate, GeneratorResult, OutputFlags) | 60 |
| `internal/harness/proposalgen/reader.go` | JSONL line-by-line parser, malformed-line tolerance | 90 |
| `internal/harness/proposalgen/reader_test.go` | Reader unit tests + table-driven malformed-line scenarios | 180 |
| `internal/harness/proposalgen/mapper.go` | Promotion → ProposalCandidate mapping, regex-based exclusion | 110 |
| `internal/harness/proposalgen/mapper_test.go` | Mapper table-driven tests including current-data no-op fixture | 220 |
| `internal/harness/proposalgen/scaffolder.go` | `.moai/proposals/<draft-id>/` directory + spec.md/proposal.json writer | 140 |
| `internal/harness/proposalgen/scaffolder_test.go` | Scaffolder tests using `t.TempDir()` isolation | 180 |
| `internal/cli/harness/propose.go` | `moai harness propose` subcommand factory `newHarnessProposeCmd()`, flag parsing, JSON stdout | 130 |
| `internal/cli/harness/propose_test.go` | CLI integration tests via cobra test pattern | 180 |
| `internal/cli/harness/propose_boundary_test.go` | `TestPropose_NoAskUserQuestion` static check (REQ-PGN-013) | 60 |
| `internal/cli/harness_route.go` | MODIFY +1 line — `cmd.AddCommand(newHarnessProposeCmd())` in existing `newHarnessRouterCmd()` body | +1 |
| **Total** | | **~1351 LOC (target ≤1500 with margin)** |

## §C. Constraints

### C.1 Language Neutrality

All proposal output (`.moai/proposals/<draft-id>/spec.md`, `proposal.json`) MUST be language-neutral. No Go struct examples, no Python snippets, no TypeScript signatures in the generated body. EARS-style natural-language requirements only. Per CLAUDE.local.md §15 (16-language neutrality).

### C.2 Subagent Boundary HARD

[HARD] `internal/cli/harness/*.go` (excluding `*_test.go`) MUST NOT contain the substring `AskUserQuestion(` anywhere — including string literals, function calls, type references, or comment invocation patterns. Comments describing the orchestrator's AskUserQuestion behavior in prose (e.g., "the orchestrator presents AskUserQuestion after this CLI returns") are permitted as documentation but MUST NOT contain the parenthesized invocation form. Enforced by `TestPropose_NoAskUserQuestion`. See spec.md REQ-PGN-012/013.

### C.3 TDD-First

Per `.moai/config/sections/quality.yaml` `development_mode: tdd`. Every REQ has a failing test authored BEFORE implementation. RED → GREEN → REFACTOR cycle per `moai-workflow-tdd` skill.

### C.4 Coverage Target

`go test -cover ./internal/harness/proposalgen/... ./internal/cli/harness/...` ≥ 85% (per project standard). Per-file coverage ≥ 80% for non-trivial files.

### C.5 Main-Direct-Push Policy

Per CLAUDE.local.md §23.7 Hybrid Trunk 1-person OSS. No plan/* or feat/* branches. plan-phase commit lands directly on `main`. pre-push hook 5s warn-only + CI 4-status-checks (Test ubuntu / Lint / Build linux/amd64 / CodeQL) remain protective. No worktree (per user decision A.3-2).

### C.6 No External Dependencies

Generator MUST run fully offline. No `net/http`, no third-party API calls, no os.exec to external tools. Only stdlib + existing `internal/harness` package imports.

### C.7 Pre-Spawn Sync Discipline (per agent-common-protocol.md § Pre-Spawn Sync Check)

Before spawning manager-develop for run-phase, the orchestrator MUST run `git fetch origin && git rev-list --count --left-right origin/main...HEAD` and verify `0 N` or `0 0` (local clean or ahead). `N 0` (origin ahead) requires AskUserQuestion gate (rebase/inspect/abort) before run-phase begins.

## §D. Methodology

TDD RED-GREEN-REFACTOR per phase. The manager-develop subagent receives a single Milestone (M1) covering the full implementation. Within M1, the implementer follows this internal phase ordering:

### D.1 Phase 1 — Reader (RED → GREEN → REFACTOR)

1. RED: Author `reader_test.go` table-driven tests covering (a) happy path 8-record live fixture, (b) malformed-line tolerance, (c) empty/missing file path. Tests fail (no implementation).
2. GREEN: Implement `reader.go` minimally — `os.Open`, `bufio.Scanner`, `json.Unmarshal` per line into `internal/harness.Promotion`, accumulate malformed-line count.
3. REFACTOR: Extract `parseLine` helper, tighten error wrapping with `fmt.Errorf("reader: %w", err)`, add godoc.

### D.2 Phase 2 — Mapper (RED → GREEN → REFACTOR)

1. RED: Author `mapper_test.go` table-driven tests with (a) current-data fixture (4 system-event patterns → 0 candidates expected), (b) synthetic actionable fixture (`code_change:func_extract:auth_module` → 1 candidate), (c) confidence threshold boundary (0.69 excluded, 0.70 included), (d) tier filter (observation/auto_update excluded, recommendation/approval_required included).
2. GREEN: Implement `mapper.go` — regex `^(code_change|error_pattern|tool_failure|repeated_edit):[a-z_]+:[^:]+$`, confidence ≥0.70 filter, tier filter, dedup by `pattern_key`.
3. REFACTOR: Extract `isActionable(p Promotion) bool` predicate, add explanatory godoc per filter rule with cross-reference to REQ-PGN-004.

### D.3 Phase 3 — Scaffolder (RED → GREEN → REFACTOR)

1. RED: Author `scaffolder_test.go` using `t.TempDir()` — verify (a) directory `.moai/proposals/PROPOSAL-<date>-<hash>/` created with 0o755, (b) `spec.md` contains `## Origin` section with pattern_key/observation_count/confidence, (c) `proposal.json` valid JSON with schema, (d) no-op path (zero candidates) creates nothing.
2. GREEN: Implement `scaffolder.go` — derive draft-id as `PROPOSAL-<YYYYMMDD>-<sha256(pattern_key)[:8]>`, write language-neutral spec.md template, write proposal.json with `encoding/json` indented.
3. REFACTOR: Extract `renderSpecMd` and `marshalProposalJSON` helpers, ensure all path joins use `filepath.Join`.

### D.4 Phase 4 — CLI Subcommand (RED → GREEN → REFACTOR)

1. RED: Author `propose_test.go` cobra integration tests — (a) `--dry-run` exits 0 with no-op JSON on current-data fixture, (b) `--auto` sets `auto_delegate: true`, (c) `--limit 3` truncates, (d) malformed flag exits 2.
2. GREEN: Implement `propose.go` with the `newHarnessProposeCmd() *cobra.Command` factory — flag parsing via cobra, JSON stdout via `encoding/json`, exit codes per REQ-PGN-009. Add the single registration line `cmd.AddCommand(newHarnessProposeCmd())` to the existing `newHarnessRouterCmd()` body in `internal/cli/harness_route.go` (NO new parent cobra.Command is created).
3. REFACTOR: Extract `runPropose(ctx, flags) GeneratorResult` for testability, ensure deterministic ordering of proposals in JSON output (sort by `Confidence` desc).

### D.5 Phase 5 — CI Boundary Guard (RED → GREEN)

1. RED: Author `propose_boundary_test.go` mirroring `internal/cli/worktree/new_test.go` TestNew_NoAskUserQuestion — scan `propose.go` (the new subpackage's sole non-test source file) for `strings.Contains(src, "AskUserQuestion(")`. Test PASSES initially (no implementation yet contains the substring). Note: the existing `internal/cli/harness_route.go` (modified with +1 line registration) is NOT scanned by this test because it is outside the `internal/cli/harness/` subpackage; if the boundary scan needs to cover the registration site, the test file pattern can be extended in a follow-up SPEC.
2. GREEN: Confirm test catches regression — temporarily inject `// AskUserQuestion(` comment into propose.go, verify test FAILS, remove comment, verify test PASSES. (Manual regression verification only; not a committed RED state.)

### D.6 Phase 6 — Integration & Verification

1. Run full test suite: `go test ./internal/harness/proposalgen/... ./internal/cli/harness/...`
2. Run coverage: `go test -coverprofile=cover.out ./internal/harness/proposalgen/... ./internal/cli/harness/...` and verify ≥85%
3. Run subagent boundary grep manually: `grep -rn 'AskUserQuestion(' internal/cli/harness/ | grep -v '_test.go'` (expect 0 matches)
4. Run cross-platform build: `GOOS=windows GOARCH=amd64 go build ./...` exit 0
5. Run `go vet ./...` exit 0
6. Run `golangci-lint run --timeout=2m` — 0 NEW issues attributable to this SPEC
7. Run CLI smoke: `go run ./cmd/moai harness propose --dry-run` → JSON output matches REQ-PGN-014 expected shape

## §E. Verification

### E.1 Test Suite Commands

```bash
# Full test suite for new packages
go test ./internal/harness/proposalgen/...
go test ./internal/cli/harness/...

# Coverage with profile output
go test -coverprofile=cover.out ./internal/harness/proposalgen/... ./internal/cli/harness/...
go tool cover -func=cover.out | tail -1   # expect total >= 85.0%

# Static checks
go vet ./...
golangci-lint run --timeout=2m

# Subagent boundary grep (manual confirmation of REQ-PGN-013)
grep -rn 'AskUserQuestion(' internal/cli/harness/ | grep -v '_test.go'
# Expected: zero output

# Cross-platform build
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# CLI smoke against current data
go run ./cmd/moai harness propose --dry-run
# Expected JSON: {"proposals":[], "reason":"no-actionable-patterns", "malformed_lines":0, "evaluated_patterns":4, "auto_delegate":false}
```

### E.2 Parallel Verification Batch (per agent-common-protocol.md §Parallel Execution)

Orchestrator-side post-run verification SHOULD batch these 7 commands in a single response turn:

1. `go test ./internal/harness/proposalgen/... ./internal/cli/harness/...`
2. `go test -coverprofile=cover.out ./internal/harness/proposalgen/... ./internal/cli/harness/...`
3. `grep -rn 'AskUserQuestion(' internal/cli/harness/ | grep -v '_test.go'`
4. `grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ | head -20`
5. `go run ./cmd/moai harness propose --dry-run`
6. `go vet ./...`
7. `golangci-lint run --timeout=2m`

### E.3 AC-to-Verification Mapping

See `acceptance.md` for the full AC-PGN-001..008 table with per-AC verification command and expected outcome.

## §3. Trade-off Matrix (MANDATORY for Tier M)

Three architectural decisions require trade-off evaluation. Each choice evaluates 2-3 alternatives with explicit pros/cons. Decisions are recorded as the recommended option with rationale.

### §3.1 Package Placement

| Option | Path | Pros | Cons |
|--------|------|------|------|
| **A (RECOMMENDED)** | `internal/harness/proposalgen/` | Co-locates with `internal/harness/types.go` (Promotion struct); clear "harness subdomain" semantic; mirrors existing sub-package pattern (`internal/harness/applier.go` siblings) | Slightly nests deeper; import path verbose |
| B | `internal/harness/proposal/` | Shorter name; symmetric with potential future `internal/harness/observation/` | "proposal" alone is ambiguous (proposal of what?); harder to grep for "proposalgen" context |
| C | `internal/proposal/` | Decouples from harness package; reusable if future SPEC sources emerge | Loses semantic anchor to harness learning loop; breaks the "harness-specific" cognitive boundary |

**Decision**: Option A. Co-location with `internal/harness/` clarifies the V3R4 lineage and makes `import "internal/harness"` for the `Promotion` struct a sibling-package import (lower coupling cost than cross-package).

### §3.2 CLI Placement

| Option | Pattern | Pros | Cons |
|--------|---------|------|------|
| **A (RECOMMENDED)** | Add `propose` as new subcommand under existing `newHarnessRouterCmd()` in `internal/cli/harness_route.go` (mirroring the existing `status` / `apply` / `rollback` / `disable` siblings registered at lines 80-86 cluster) | Reuses the already-registered V3R4 harness command surface; zero new top-level cobra surface; semantically consistent with existing `moai harness <verb>` siblings; avoids the cobra duplicate-subcommand panic that creating a new parent would trigger at startup (`internal/cli/root.go:104` already registers `newHarnessRouterCmd()`) | Requires careful insertion in the existing AddCommand cluster (one line) instead of a self-contained new parent file |
| B | Extend `hookCmd`: `moai hook propose` | Reuses HCW-001 precedent (`moai hook harness-classify`); zero new top-level surface | Semantic mismatch — "hook" is event-driven; "propose" is on-demand human-triggered; bundling muddles intent |
| C | Standalone `proposeCmd`: `moai propose` | Shortest CLI invocation | "propose" alone is ambiguous (propose SPEC? propose PR? propose change?); no namespace reservation for future harness commands; also adds a new top-level surface |

**Decision**: Option A. Extending the existing `newHarnessRouterCmd()` parent avoids the cobra duplicate-subcommand panic that creating a new parent would cause, and aligns `propose` semantically with its V3R4 sibling verbs (`route` / `validate` / `status` / `apply` / `rollback` / `disable` / `mute` / `mute-list` / `unmute` / `verify`). The factory function follows the existing naming convention as `newHarnessProposeCmd()` and is registered with a single-line `cmd.AddCommand(newHarnessProposeCmd())` inside the existing `newHarnessRouterCmd()` body. No changes to `internal/cli/root.go` or `cmd/moai/main.go` are required.

### §3.3 Output Format

| Option | Format | Pros | Cons |
|--------|--------|------|------|
| **A (RECOMMENDED)** | Structured JSON with metadata: `{"proposals":[...], "reason":"<string>", "malformed_lines":<int>, "evaluated_patterns":<int>, "auto_delegate":<bool>}` | Single stdout payload trivially parseable by orchestrator with `jq`; explicit `reason` field surfaces no-op semantics; `auto_delegate` flag enables clean gate handoff | Slightly verbose for the common no-op case |
| B | Pure JSONL stream of proposal objects (one per line) | Streamable; conventional for log-like output | No top-level metadata fields (cannot express no-op reason without sentinel record); orchestrator must consume entire stream to know status |
| C | Markdown + JSON sidecar (write `.moai/proposals/<id>/spec.md` and `.moai/proposals/<id>/proposal.json` per candidate, CLI stdout only summary) | Separates human-readable artifact from machine-readable metadata | The scaffolder already writes these files per REQ-PGN-006; CLI stdout would duplicate. Loses single-source structured payload |

**Decision**: Option A. The orchestrator's AskUserQuestion gate (REQ-PGN-010/011) requires a structured single payload to drive the 3-option presentation. The scaffolder still writes per-proposal files (Option C is partially adopted — the disk artifacts ARE markdown+JSON), but the CLI stdout payload is structured JSON for machine consumption. Both layers coexist.

## §4. Open Questions

None. All architectural decisions are resolved in §3 trade-off matrix. All scope/policy decisions are resolved in §A.3 user-confirmed decisions. Run-phase delegation may proceed with no inline open questions.
