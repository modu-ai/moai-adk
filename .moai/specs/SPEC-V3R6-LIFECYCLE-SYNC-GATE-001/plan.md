---
id: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
artifact: plan
version: "0.1.2"
created: 2026-05-25
updated: 2026-05-26
---

## HISTORY

### v0.1.2 (2026-05-26, manager-spec — iter-3 narrow-scope defect resolution)
- D3 SHOULD-FIX resolved: milestone "Binds to AC" lines in F.1 (M1) / F.2 (M2) / F.6 (M6) synchronized with acceptance.md §D.3 AC→Milestone mapping (v0.1.1). F.1 appends AC-019/020/021/022; F.2 appends AC-022; F.6 appends AC-020/022. F.3/F.4/F.5 unchanged.
- D1 BLOCKING resolved (mirror): F.6 M6 milestone scope reframed from active backfill dogfood (5 modern-era violations) to no-op regression validation against 5 already-discharged SPECs. M6 commit count expectation revised from "produces 5 chore commits" to "produces 0 chore commits (no-op success path)"; E7 self-verification commit count adjusted accordingly.
- Cross-reference to spec.md v0.1.2 HISTORY for full D1/D2/D3 defect catalogue and trajectory analysis

### v0.1.1 (2026-05-25, manager-spec — iter-2 narrow-scope defect resolution)
- D2 resolved: D.1.2 HARD amended with explicit "this SPEC's plan-phase artifacts" qualifier + M3 carve-out clause (settings.json.tmpl PreCommit array registration is the SOLE allowed template change)
- D9 resolved: E7 self-verification commit count updated to ≥ 15 expected commits (6 M1-M6 + 5 M6 dogfood close + plan chore + sync chore + mx chore + atomic terminator)
- Cross-reference to spec.md v0.1.1 HISTORY for full D1-D10 defect catalogue

### v0.1.0 (2026-05-25, manager-spec)
- Initial plan-phase implementation roadmap authored
- Tier L 6-milestone (M1-M6) structure per spec-workflow.md
- Run-phase scope estimate: ~1650 LOC across 15 files
- M6 dogfood verification milestone on 5 known modern-era violations

---

## A. Context

This plan.md documents the implementation roadmap for SPEC-V3R6-LIFECYCLE-SYNC-GATE-001. The SPEC defines 5 deliverables (atomic close CLI, audit CLI, pre-commit hook, spec-lint extension, rule file) addressing the 5 remaining modern-era status drift violations and providing defensive guardrails against the L67 manager-docs scope-creep pattern.

The run-phase work is decomposed into 6 milestones (M1-M6) per Tier L lifecycle. M1-M5 implement deliverables in dependency order; M6 dogfoods the implementation on 5 known violations and validates the atomic close end-to-end.

## B. Known Issues (filtered to plan-phase relevant)

- **B1**: L60 atomic backfill chicken-and-egg pattern remains valid as backward-compat (D.1.5 HARD). `moai spec close` is additive.
- **B2**: 145 pre-V3R6 SPECs grandfather-clause-protected (D.1.1 HARD). Audit tool MUST NOT surface them as drift.
- **B3**: Pre-commit hook integration follows agent-common-protocol §Hook Invocation Surface (REQ-LSG-011 HARD). NO AskUserQuestion in hook.
- **B4**: Frontmatter schema canonical 12 fields preserved; optional `era:` field is additive per `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- **B5**: spec-lint heading h3 sub-section convention preserved (`### A.5.N Out of Scope`).
- **B6**: subagent boundary preserved — all subagent spawns return blocker reports, no direct user interaction.

## C. Pre-flight Checks (run-phase entry validation)

```bash
# 1. baseline test suite clean
go test ./...                                   # → all PASS expected (current baseline)

# 2. target SPEC plan-phase committed
git log --oneline -1                           # → feat(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): plan-phase artifacts (expected)

# 3. predecessor SPECs implemented
cat .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md | head -10  # status: implemented (needs M6 dogfood close)
cat .moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/spec.md | head -10   # status: implemented (needs M6 dogfood close)
cat .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md | head -10            # status: completed (Status Transition Ownership Matrix)

# 4. multi-session race check
git fetch origin main
git rev-list --count --left-right origin/main...HEAD   # → 0 N (clean)
```

## D. Constraints (run-phase DO NOT VIOLATE)

### D.1 Run-phase HARD Constraints

1. **[HARD]** Each milestone M1-M6 MUST commit independently with explicit attribution to SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
2. **[HARD]** `internal/template/templates/**` MUST NOT be modified by this SPEC's plan-phase artifacts (template-internal-isolation policy per CLAUDE.local.md §25), **EXCEPT** `internal/template/templates/.claude/settings.json.tmpl` PreCommit hook array entry registration per M3 scope (REQ-LSG-003 binding). The M3 carve-out is narrowly scoped: registration of the `handle-pre-commit-spec-status.sh` entry in the `hooks.PreCommit` array ONLY; no other template directory changes are permitted by any milestone of this SPEC.
3. **[HARD]** L60 atomic backfill design MUST remain backward-compatible (additive close, not replacement)
4. **[HARD]** Pre-V3R6 SPECs (145) MUST be grandfather-clause-protected — no retroactive normalization
5. **[HARD]** Pre-commit hook MUST NOT call AskUserQuestion (exit 2 + JSON only)
6. **[HARD]** Each milestone MUST achieve `go test ./...` PASS before proceeding to the next milestone
7. **[HARD]** M6 dogfood close MUST succeed on all 5 known violations OR run-phase blocks with structured blocker report

### D.2 Run-phase SHOULD Constraints

1. **[SHOULD]** Coverage for `internal/spec/closer.go` ≥ 85%
2. **[SHOULD]** Coverage for `internal/spec/audit.go` ≥ 85%
3. **[SHOULD]** Coverage for `internal/cli/spec_close.go` ≥ 80%
4. **[SHOULD]** Coverage for `internal/cli/spec_audit.go` ≥ 80%

## E. Self-Verification (run-phase per-milestone)

E1. **Test suite PASS**: `go test ./...` exits 0 after every milestone
E2. **Lint PASS**: `golangci-lint run --timeout=2m` exits 0
E3. **Coverage threshold**: per-package coverage meets D.2 SHOULD thresholds (warn-only, not blocking)
E4. **CLI smoke**: `go run ./cmd/moai spec close --help` AND `go run ./cmd/moai spec audit --help` produce expected output
E5. **Hook smoke**: `bash .claude/hooks/moai/handle-pre-commit-spec-status.sh < test-input.json` exits 0 or 2 per expected
E6. **AC binding traceability**: each AC-LSG item references the run-phase artifact that satisfies it
E7. **Commit attribution**: `git log --oneline --grep="SPEC-V3R6-LIFECYCLE-SYNC-GATE-001"` returns **≥ 11 commits** expected (revised v0.1.2 per D1 reframe): 6 milestone commits (M1-M6, where M6 produces 0 SPEC-targeted close commits per no-op reframe but is itself one M6 verification chore commit attributed to this SPEC) + plan chore (already committed `0616823dc`) + plan iter-2 chore (`2c4513930`) + plan iter-3 chore (this iter-3 amendment) + sync chore + mx chore + atomic close terminator chore. The 5 M6 close-target commits previously expected (one per violation) are eliminated by the v0.1.2 D1 reframe (no-op semantics produce 0 commits per close invocation against already-completed SPECs).

## F. Milestones (Priority Order, No Time Estimates)

### F.1 M1 — Go Primitives + Frontmatter `era` Field

**Priority**: P1 (foundational; blocks M2-M5)

**Scope**:
- `internal/spec/closer.go` (~300 LOC) — atomic close core logic: precondition matrix validation, file lock acquisition, multi-file staging (spec.md frontmatter + progress.md §E.3 + sync_commit_sha + mx_commit_sha)
- `internal/spec/audit.go` (~200 LOC) — era classification engine: heuristic-based era detection, grandfather clause exclusion, drift finding emission
- `internal/spec/lifecycle_close_test.go` (~250 LOC) — unit tests for closer.go (10+ test cases covering precondition success, precondition failure modes, file lock contention, atomic rollback on staging failure)
- `internal/spec/audit_test.go` (~200 LOC) — unit tests for audit.go (15+ test cases covering era classification by file pattern, grandfather clause for pre-V3R6, modern-era drift detection)
- Frontmatter optional `era:` field added to spec-frontmatter-schema.md (additive, non-breaking)

**Exit criteria**: M1 commit `feat(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): M1 closer.go + audit.go primitives + era field`, all unit tests PASS, coverage ≥85% for new files

**Binds to AC**: AC-LSG-001, AC-LSG-002, AC-LSG-006, AC-LSG-007, AC-LSG-009, AC-LSG-010, AC-LSG-013, AC-LSG-014, AC-LSG-019 (Cross-Platform CI matrix, NEW per v0.1.1 D1), AC-LSG-020 (Observability log emission, NEW per v0.1.1 D1), AC-LSG-021 (Concurrent close safety, formerly AC-018a renumbered per v0.1.1 D4), AC-LSG-022 (Backfill-only mode core, NEW per v0.1.1 D7)

### F.2 M2 — CLI Subcommands (`moai spec close` + `moai spec audit`)

**Priority**: P1 (depends on M1)

**Scope**:
- `internal/cli/spec_close.go` (~200 LOC) — cobra command wiring: flag parsing (`--backfill-only`, `--dry-run`, `--force`), delegate to `internal/spec/closer.Close()`, structured error output
- `internal/cli/spec_audit.go` (~150 LOC) — cobra command wiring: flag parsing (`--json`, `--filter-era=V3R6`, `--include-grandfathered`), delegate to `internal/spec/audit.Audit()`, JSON/human output formats
- `internal/cli/spec_close_test.go` (~150 LOC) — integration tests for CLI surface (flag parsing, help text, error message format)
- `internal/cli/spec_audit_test.go` (~150 LOC) — integration tests for audit CLI (JSON output schema, filter behavior)
- `cmd/moai/main.go` — register new `spec close` + `spec audit` subcommands under existing `spec` command group

**Exit criteria**: M2 commit `feat(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): M2 spec close + spec audit CLI subcommands`, `go run ./cmd/moai spec close --help` + `go run ./cmd/moai spec audit --help` produce expected output, all integration tests PASS

**Binds to AC**: AC-LSG-001, AC-LSG-002, AC-LSG-006, AC-LSG-007, AC-LSG-014, AC-LSG-016 (NFR performance), AC-LSG-022 (Backfill-only mode CLI flag wiring, NEW per v0.1.1 D7)

### F.3 M3 — Pre-Commit Hook + settings.json.tmpl Registration

**Priority**: P2 (depends on M2)

**Scope**:
- `.claude/hooks/moai/handle-pre-commit-spec-status.sh` (~80 LOC bash) — read staged diff, check for spec.md status field changes, compare with progress.md §E.3 status field, emit exit 0/2 + JSON output
- `.claude/hooks/moai/handle-pre-commit-spec-status_test.sh` (~50 LOC bash) — bats-style test harness OR Go test invoking the script via os/exec
- `internal/template/templates/.claude/settings.json.tmpl` — register hook under `hooks.PreCommit` array. This modification is **explicitly permitted** by the D.1.2 HARD carve-out (v0.1.1, D2 fix): registration is the SOLE allowed template change for this SPEC. Restrict modification to the PreCommit array entry for `handle-pre-commit-spec-status.sh` only; NO other template changes (other files, other hook arrays, settings.json keys outside hooks.PreCommit) are permitted.

**Exit criteria**: M3 commit `feat(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): M3 pre-commit hook + settings.json registration`, hook script executable, test harness PASS, `git commit` with staged spec.md status mismatch fails with exit 2 + structured JSON

**Binds to AC**: AC-LSG-003, AC-LSG-008, AC-LSG-011, AC-LSG-015

**Risk note**: M3 settings.json.tmpl modification is the SOLE allowed template directory change for this SPEC per the M3 scope clarification. All other template changes remain forbidden per D.1.2 HARD.

### F.4 M4 — spec-lint `OwnershipTransitionRule` Extension

**Priority**: P2 (depends on M2)

**Scope**:
- `internal/spec/lint_ownership.go` (~50 LOC) — new lint rule implementing `Rule` interface, parses commit subject pattern + spec.md status transition diff, emits `OwnershipTransitionInvalid` finding when manager-develop authorship + `* → implemented` transition detected
- `internal/spec/lint_ownership_test.go` (~150 LOC) — unit tests (10+ cases: manager-develop direct transition INVALID, manager-docs transition VALID, lint.skip opt-out path)
- Register new rule in `internal/spec/lint.go` rule registry

**Exit criteria**: M4 commit `feat(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): M4 spec-lint OwnershipTransitionRule extension`, all lint tests PASS, `moai spec lint` against repo emits no false positives

**Binds to AC**: AC-LSG-004, AC-LSG-012

### F.5 M5 — Rule File Authoring

**Priority**: P3 (depends on M1 era field)

**Scope**:
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (~250 LOC) — protocol SSOT: era classification heuristic table, grandfather clause policy, frontmatter `era:` field semantics + auto-detection rules, cross-reference to Status Transition Ownership Matrix, worked example demonstrating era auto-detection
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — minimal amendment adding `era` to optional fields list with cross-reference to lifecycle-sync-gate.md

**Exit criteria**: M5 commit `docs(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): M5 lifecycle-sync-gate.md rule file + schema amendment`, rule file authored, schema cross-reference added

**Binds to AC**: AC-LSG-005, AC-LSG-013, AC-LSG-017 (worked example)

### F.6 M6 — No-Op Regression Validation (5 Already-Discharged SPECs)

**Priority**: P1 (depends on M1-M5)

**Scope (v0.1.2 reframe per D1 BLOCKING)**: Execute `moai spec close SPEC-XXX --backfill-only` against each of the 5 SPECs originally identified as modern-era violations but now already at `status: completed` (transitioned via orchestrator-direct retroactive Mx chores `a1fb04625` / `8d0b1fdf9` / `d167eb08b` / `ac8ba9a99` / `adc75a33c` executed 2026-05-25 20:54-20:57, post-iter-2 verdict):
- SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (status: completed)
- SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 (status: completed)
- SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (status: completed — PROCEED-WITH-DEBT precedent)
- SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 (status: completed)
- SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 (status: completed)

**Expected behavior per close invocation**: command MUST exit code 0 as no-op (no staging change, no commit produced). Each invocation logs a single line to `.moai/logs/lifecycle-close.log` with `mode: "backfill-only"`, `result: "success"`, `transitions: {}` (empty object indicates no fields needed transition), and a structured log signal indicating "noop: already completed" pattern (stderr or log body — implementation detail in M1 closer.go).

**Rationale (v0.1.2 reframe)**: This validates the implementation handles the already-completed precondition state gracefully (success path, not failure). Active backfill dogfood is no longer applicable because the 5 target SPECs are no longer in modern-era-violation state. The no-op regression validation binds to AC-LSG-022's existing `fully-completed-noop` fixture state (acceptance.md §B.22 last fixture in parametric `TestBackfillOnlyVariants`); no new fixture or AC required.

**Exit criteria**: M6 produces **0 chore commits** (5 no-op invocations, each producing zero staging changes and zero commits). Post-M6 `moai spec audit --filter-era=V3R6 --json` returns 0 drift findings for these 5 SPECs (precondition already satisfied). `.moai/logs/lifecycle-close.log` contains ≥ 5 entries (one per no-op invocation) with `result: "success"` and `transitions: {}`. All 5 SPECs' spec.md frontmatter status remains `completed` (unchanged — no-op).

**Binds to AC**: AC-LSG-018 (no-op regression verification — all 5 invocations exit 0 with no-op semantics), AC-LSG-020 (Observability — verifies ≥ 5 log entries appended per NFR-LSG-004, NEW per v0.1.1 D1), AC-LSG-022 (Backfill-only mode `fully-completed-noop` fixture verified in production usage, NEW per v0.1.1 D7)

## G. Anti-Patterns (DO NOT)

- **AP-1**: Do not modify L60 atomic backfill code paths — `moai spec close` is purely additive (D.1.5 HARD)
- **AP-2**: Do not surface pre-V3R6 SPECs as drift findings — grandfather clause is non-negotiable (D.1.1 HARD)
- **AP-3**: Do not invoke `AskUserQuestion` from inside the pre-commit hook — exit 2 + JSON only (D.1.5 HARD)
- **AP-4**: Do not modify `internal/template/templates/**` except for the M3 settings.json.tmpl `hooks.PreCommit` array entry registration for `handle-pre-commit-spec-status.sh` (D.1.2 HARD with M3 carve-out per v0.1.1 D2 fix; carve-out is single-file + single-array narrow scope)
- **AP-5**: Do not retroactively normalize pre-V3R6 SPECs — out-of-scope per A.5.1
- **AP-6**: Do not change CHANGELOG.md in plan-phase — sync-phase responsibility (A.5.3)
- **AP-7**: Do not split `moai spec close` into multi-commit cadence — atomicity is the central invariant (D.1.3 HARD)

## H. Cross-References

- spec.md §B.1-B.5 — REQ-LSG-001..015 GEARS requirements
- acceptance.md — AC-LSG-001..018 verification criteria
- design.md §A-§D — architecture diagrams + decision rationale
- research.md §A-§D — origin / motivation / verbatim memory citations
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface
- Memory L60 (atomic backfill chicken-and-egg)
- Memory L67 (manager-docs scope-creep)
