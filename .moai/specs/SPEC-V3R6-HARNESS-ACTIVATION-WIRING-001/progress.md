# Progress — SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
era: V3R6
tier: M
plan_complete_at: 2026-06-03
plan_status: audit-ready
plan_audit_verdict: PASS-WITH-DEBT 0.86 (GATE-2 approved)
plan_audit_remediation: D1/D2 (SHOULD-FIX) + D3/D4 (MINOR) all addressed in this commit
artifacts:
  - spec.md          # 12-field frontmatter + era:V3R6, GEARS REQ-HAW-001..016 (+013b), Exclusions EX-1..EX-8
  - plan.md          # Tier M justified, M1..M6 milestones, template-first ordering, D1 prefix disambiguation
  - acceptance.md    # AC-HAW-001..015 + AC-HAW-PROC-1..2, all grep/test-verifiable, DoD, edge cases
  - design.md        # wiring-mechanism decision (Option A recommended), main.md additive router, 3-check smoke gate
  - progress.md      # this file
authored_by: manager-spec
plan_commit_sha: (this commit)
```

### Plan-audit remediation log (2026-06-03)

| Defect | Severity | Resolution |
|--------|----------|------------|
| D1 | SHOULD-FIX | plan.md M4 + AC-HAW-PROC-2 disambiguated: §6.4 correction target is the code-side `my-harness-*` (NOT the EX-1 `harness-*` migration); cites `meta-harness SKILL.md:168` doctrine-vs-code drift; AC PROC-2 now asserts `my-harness-` equality + no bare `harness-*` directive (impossible to read as endorsing migration) |
| D2 | SHOULD-FIX | (preferred option) REQ-HAW-013b + AC-HAW-015 added: Phase-6 smoke gate FAILs when a generated agent OMITS `skills:` — runtime enforcement of REQ-HAW-008, closing the silent auto-discovery gap; spec.md REQ-HAW-008 binding updated; plan.md M5 + design.md §C updated |
| D3 | MINOR | design.md §B corrected: `mainMD()` already emits Domain Summary + Linked Files; only the `## Task-Shape Routing` table is additive (not a from-scratch rewrite) |
| D4 | MINOR | spec.md EX-1 future SPEC-ID `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` marked **(planned — not yet created)** |

## §E.0 Ground-Truth Diagnosis Anchors (verified during plan-phase)

- `InjectMarker` (`internal/harness/layer3.go`) — **0 non-test callers** (orphaned, but works + tested)
- `ScaffoldHarnessDir` (`internal/harness/layer5.go`, emits `main.md`) — **0 non-test callers** (orphaned)
- This repo CLAUDE.md + template CLAUDE.md — **0 `moai:harness-start` markers** each
- `project/meta-harness.md` — Phase 7 (5-Layer Activation) referenced but **body absent** (file ends at Phase 6.5)
- L4 import lines (`@.moai/harness/`) present in all of plan/run/sync/design workflows — L4 intact
- `doctor harness` L3 (marker) + L5 (`main.md`) checks already exist — smoke gate reuses them
- B3 (empty agent descriptions) REFUTED per diagnosis — codified as REQ-HAW-009 for the gate to assert

## §E — Phase 0.95 Mode Selection

**Input parameters**
- tier: M (300-1000 LOC, 5-15 files)
- scope (file count): ~8-10 (2 Go source + 2 Go test + 2 template skill/workflow + 2 mirror sync)
- domain count: 2 — Go source code (`internal/cli/`, `internal/harness/`) + orchestration skill/workflow markdown (`.claude/skills/...`)
- file language mix: Go-dominant (the load-bearing wiring + smoke gate + TDD tests are Go); markdown is additive skill/workflow body
- concurrency benefit: LOW — coding-heavy, single coherent wiring chain (CLI command → InjectMarker/ScaffoldHarnessDir caller → smoke gate); milestones are sequential-dependent (M3 CLI before M4 router before M5 gate)
- Agent Teams prereqs status: not evaluated (single-domain coding work does not meet ≥3-domain gate)

**Mode evaluation**

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | not selected | Multi-file semantic change (new CLI command + smoke gate + tests) |
| 2 background | not selected | Write/Edit operations required (CONST-V3R2-020 forbids background writes) |
| 3 agent-team | not selected | Not multi-domain (≥3); Agent Teams capability gate not met |
| 4 parallel | not selected | Coding-heavy work — Finding A4 caveat (coding tasks have few parallelizable subtasks) |
| 5 sub-agent | **selected** | Default fallback for coding-heavy work; sequential milestone execution (M1→M6) with inter-milestone dependency |
| 6 workflow | not selected | Not ≥30-file mechanical-uniform transform; this is semantic new-code Go wiring |

**Decision: sub-agent**

**Justification**: This is coding-heavy Go work — the orphaned-installer wiring, the `moai harness install` CLI command, and the TDD smoke-gate extension form a single coherent dependency chain (the CLI command calls `InjectMarker`/`ScaffoldHarnessDir`; the smoke gate must follow the router restructure). Per Anthropic Finding A4, coding tasks involve fewer truly parallelizable subtasks than research, so the sequential sub-agent path (Mode 5) is the correct default. Mode 6 (workflow) is explicitly rejected: this is not a ≥30-file uniform mechanical transform but a small set of semantic new-code edits with inter-file dependency. Modes 3/4 are rejected because the scope is single-domain (Go + companion markdown) and below the multi-domain parallelism gate.

## §E.2 Run-phase Evidence

### M1 — RED baseline (ground truth captured 2026-06-03)

The four wiring breaks asserted as failing pre-conditions at run-phase start:

| Ground-truth assertion | Command | Result (RED) |
|------------------------|---------|--------------|
| `InjectMarker` orphaned | `grep -rn "InjectMarker(" --include="*.go" internal/ \| grep -v "_test.go" \| grep -v "func InjectMarker"` | **0 callers** (orphaned) |
| `ScaffoldHarnessDir` orphaned | `grep -rn "ScaffoldHarnessDir(" --include="*.go" internal/ \| grep -v "_test.go" \| grep -v "func ScaffoldHarnessDir"` | **0 callers** (orphaned) |
| CLAUDE.md markers absent | `grep -c "moai:harness-start" CLAUDE.md` | **0 markers** |
| Phase 7 body absent | `grep -n "Phase 7\|5-Layer Activation" project/meta-harness.md` | file ends at Phase 6.5 (only a forward reference at L306) |

Pre-flight baseline (Section C):
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `golangci-lint run ./internal/cli/... ./internal/harness/...` → 0 issues (clean baseline)
- existing `TestRunHarnessCheck*` / `TestInjectMarker*` / `TestScaffold*` → all PASS (L1-L5 + installers unit-tested but unwired)

_(M2-M6 evidence populated below as milestones complete)_

## §E.3 Run-phase Audit-Ready Signal
_(populated by manager-develop at run-phase)_

## §E.4 Sync-phase Audit-Ready Signal
_(populated by manager-docs at sync-phase)_

## §E.5 Mx-phase Audit-Ready Signal
_(populated at close)_
