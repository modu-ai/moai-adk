# Verification Batch Pattern

Canonical pattern for orchestrator-side read-only verification batching during run-phase completion. Motivation: reduces serial-verification round-trip latency at run-phase completion.

Cross-reference: `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution defines the HARD batching obligation; this file owns the grouping rationale and class taxonomy.

## Why Batch

When `manager-develop` reports completion, the orchestrator independently verifies seven dimensions: test suite, coverage, subagent-boundary, sentinel-key, CLI smoke, benchmark, lint. Each is read-only and independent. Serial issuance multiplies round-trip latency; multi-Bash batching collapses it to the slowest single command.

## When to Batch (Verification Class Taxonomy)

| Class | Read-only? | Mutates state? | Batch-safe? |
|-------|------------|----------------|-------------|
| Test execution | yes (output only) | no | YES |
| Coverage measurement | yes | writes `cover.out` (no side effect) | YES |
| Grep / find / sentinel scan | yes | no | YES |
| CLI smoke (--version, --help) | yes | no | YES |
| Benchmark | yes | no | YES |
| Lint (golangci-lint, ruff, etc.) | yes | no | YES |
| Build (`go build`, `npm run build`) | depends | writes artifacts | NO if downstream depends |
| Test fixture setup | yes | writes test files | NO if shared state |

All seven canonical batch items in agent-common-protocol §Parallel Execution are read-only batch-safe.

> **Re-sync sentinel**: the verbatim 7-command batch AND the file-redirect contract (redirect + bounded-tail output representation) live in `agent-common-protocol-reference.md` § Canonical 7-item example / § File-redirect contract (the detail sidecar of `agent-common-protocol.md`, which retains the binding summary). If either the 7-item list OR the file-redirect contract representation changes, re-sync this file's grouping rationale and the class taxonomy below to match. This file owns only the *why* (grouping rationale + class taxonomy + anti-patterns), not the *what* (the verbatim command list or its output representation).

## When NOT to Batch

- Explicit dependency (`make build` before tests that invoke its binary).
- Same-file writes (two `coverprofile=cover.out` runs race).
- Shared-state mutation (`git checkout` + `git status` in one tree).

Serialize dependent ops; batch independent read-only verifications by default.

## Grouping Heuristic

| Group | Members | Typical Total Time |
|-------|---------|-------------------:|
| A. Functional | `go test ./...`, coverage | 30-120 s |
| B. Boundary | subagent-boundary grep, sentinel scan, frontmatter check | 1-5 s |
| C. Quality | golangci-lint, spec-lint | 10-60 s |
| D. Smoke | CLI --version, --help | 1-3 s |
| E. Benchmark (optional) | go test -bench | 30-300 s |

Groups A-D issue as one parallel batch. Group E joins when benchmark is in AC.

## Anti-Pattern Catalogue

- **AP-VBP-001 — Serial across turns**: N turns where one suffices. Adds N × round-trip latency plus context-switch overhead.
- **AP-VBP-002 — Pseudo-batch via `&&`**: Chains sequentially in one shell, not parallel. First failure short-circuits.
- **AP-VBP-003 — Pseudo-batch via `&`**: Interleaved output is hard to parse. Orchestrator-level multi-Bash is cleaner — each call produces a separate, structured output block.

## Correct Pattern (Reference)

The orchestrator's response contains multiple Bash tool calls within a single assistant turn. The canonical 7-item example lives in `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (satisfies the canonical verification-batch acceptance criterion).

## Attributable diff-check pattern (SPEC-SYNC-PARALLEL-DOCS-001 A9)

The canonical 7-command batch defaults to RE-EXECUTION. The A9 attributable diff-check is an opt-in composition path that SUBSTITUTES re-execution with a diff-check against the shared diagnostic snapshot + the manager-develop §E evidence — without weakening the `verification-claim-integrity.md` §1.1 invariant. This section owns the pattern + the fallback contract; the doctrinal switch that selects the path lives in `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch.

### Diff-check predicate (all-three attribution match)

For each verification dimension (test / lint / vet / cover / subagent-boundary / sentinel-key / CLI-smoke), the orchestrator SHALL evaluate the following three-way match BEFORE re-executing the corresponding command:

| Match axis | Predicate | On failure |
|---|---|---|
| **(1) Snapshot key** | `moai verify check --key-current` key == §E-cited HEAD SHA | `snapshot_key_drift` → fallback |
| **(2) Command** | snapshot-recorded command == §E-cited command (a) | `command_drift` → fallback |
| **(3) Output** | snapshot-recorded output == §E-cited observed output (b) | `missing_section_e` → fallback (if §E missing) OR `output_drift` → fallback |

When all three hold, the orchestrator CONSUMES the §E evidence for that dimension and DOES NOT re-execute. The batch row is marked `PASS-attributed` with baseline-attribution = `(snapshot key, §E evidence path)` per VCI §2. When any axis fails, the orchestrator FALLS BACK to re-execution for that dimension and logs the mismatch reason.

### Fallback-to-re-execution contract (SPEC-SYNC-PARALLEL-DOCS-001 A9 — the safety boundary)

The fallback is the safety boundary A9 exists to preserve. Omitting it turns the diff-check into a verification bypass that violates `verification-claim-integrity.md` §1.1 (the named anti-pattern AP-SPD-001 in `SPEC-SYNC-PARALLEL-DOCS-001/plan.md` §G). The fallback contract binds three properties:

1. **Any-mismatch → re-execution, NOT silent skip.** A snapshot-key drift, command drift, OR missing-§E evidence MUST restore re-execution of the affected dimension. The batch NEVER marks a dimension PASS without either (a) a fresh re-executed output OR (b) a fully-matched attributable §E evidence triple.
2. **Mismatch reason logged.** The batch records the mismatch reason (`snapshot_key_drift` / `command_drift` / `missing_section_e` / `output_drift`) in the verification report so a later audit can reconstruct which path was taken per dimension.
3. **VCI §1.1 invariant holds on every path.** Both the consume-path (b) and the fallback-path (a) produce attributable evidence satisfying VCI §1.1 surface 1 (orchestrator self-report); there is NO third "silent skip" path.

The diff-check is strictly additive: it collapses wall-time on the happy path (all-three match → no re-execution) while preserving the verification invariant on every path. A regression that silently drops the fallback re-introduces AP-SPD-001 (the named anti-pattern in `SPEC-SYNC-PARALLEL-DOCS-001/plan.md` §G).

### When the diff-check does NOT apply

The diff-check predicates on the shared diagnostic snapshot (`moai verify check --key-current`), so it applies ONLY where the snapshot interface is reachable. When the snapshot CLI is absent, returns stale, or the §E evidence was recorded against a different tree, the orchestrator proceeds directly to re-execution — the diff-check is opt-in per-dimension, never mandatory. This is the SPEC-SYNC-PARALLEL-DOCS-001 A9 fallback operating at the composition layer, identical to the per-mismatch fallback but at coarser granularity.

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (HARD batching obligation) + `agent-common-protocol-reference.md` (7-item canonical example).
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch (SPEC-SYNC-PARALLEL-DOCS-001 A9 — the composition-time switch that selects consume-vs-re-execute).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E → Attribution discipline (SPEC-SYNC-PARALLEL-DOCS-001 A9 — the §E attribution triple the diff-check consults).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 + §2 (the invariant + attribution contract A9 preserves on every path).
- reduces serial CI wait.

---

Version: 1.1.0 (SPEC-SYNC-PARALLEL-DOCS-001 A9 — attributable diff-check pattern + A9 fallback contract added)
Classification: Evolvable operational rule, applies to all run-phase completion verifications
