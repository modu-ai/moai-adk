# Verification Batch Pattern

Canonical pattern for orchestrator-side read-only verification batching during
the run-phase completion handshake.

> This rule was added by SPEC-V3R5-WORKFLOW-OPT-001 Layer D in response to the
> W3 HARNESS-AUTONOMY-001 meta-analysis. W3 lost approximately 10 minutes
> (≈11% of total run-phase wall-time) to serial verification across multiple
> orchestrator turns. This rule formalizes the parallel batching pattern that
> consolidates verification into a single turn.

> Cross-reference: `.claude/rules/moai/core/agent-common-protocol.md` §Parallel
> Execution defines the HARD batching obligation; this file documents the
> grouping rationale and verification-class taxonomy.

## Why Batch

When `manager-develop` (or any implementation subagent) reports completion, the
orchestrator must verify the report independently before accepting it. Typical
verification covers seven dimensions:

1. Test suite (functional correctness)
2. Coverage threshold (TRUST 5 craft gate)
3. Subagent-boundary discipline (C-HRA-008 class)
4. Sentinel-key correctness (retired SPEC, frozen zone)
5. CLI smoke check (binary still compiles, version still resolves)
6. Benchmark or performance baseline (where applicable)
7. Lint status (NEW vs pre-existing baseline)

Each verification is read-only and independent of the others. Issuing them
serially across N turns multiplies round-trip latency. Issuing them in a single
turn (multi-Bash) collapses the latency to the slowest single command.

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

The seven canonical batch items in agent-common-protocol §Parallel Execution
are all classified as read-only batch-safe.

## When NOT to Batch

- Commands with explicit dependency: `make build` must complete before
  `go test ./...` if the build produces a binary the tests invoke.
- Commands writing to the same file: two `go test -coverprofile=cover.out`
  invocations would race on `cover.out`.
- Commands mutating shared state: `git checkout` and `git status` cannot be
  parallelized within the same working tree.

For dependent operations, use serial execution explicitly. For independent
read-only verifications, batching is the default.

## Grouping Heuristic

Default groupings for run-phase verification:

| Group | Members | Typical Total Time |
|-------|---------|-------------------:|
| A. Functional | `go test ./...`, coverage | 30-120 s |
| B. Boundary | C-HRA-008 grep, sentinel scan, frontmatter check | 1-5 s |
| C. Quality | golangci-lint, spec-lint | 10-60 s |
| D. Smoke | CLI --version, --help | 1-3 s |
| E. Benchmark (optional) | go test -bench | 30-300 s |

Groups A, B, C, D issue together as one parallel batch. Group E joins the same
batch when benchmark is relevant to the SPEC's acceptance criteria.

## Anti-Pattern Catalogue

### AP-VBP-001: Serial verification across turns

```
Turn 1: orchestrator: "Let me verify"; Bash(go test ./...)        → wait
Turn 2: orchestrator: "Now lint"; Bash(golangci-lint run ./...)    → wait
Turn 3: orchestrator: "Now grep"; Bash(grep -rn 'AskUserQuestion') → wait
```

Three sequential turns where one would suffice. Adds 3 × round-trip latency
plus context-switch overhead.

### AP-VBP-002: Pseudo-batching via shell `&&`

```bash
go test ./... && golangci-lint run ./... && grep -rn 'AskUserQuestion' internal/
```

This chains commands sequentially in a single shell — not parallel. Worse,
the first failure short-circuits subsequent commands, hiding aggregate state.

### AP-VBP-003: Pseudo-batching via shell `&` background

```bash
go test ./... &
golangci-lint run ./... &
wait
```

This spawns background processes within a single Bash call. Output is
interleaved and harder to parse. The orchestrator-level multi-Bash batch is
cleaner because each call produces a separate, structured output block.

## Correct Pattern (Reference)

The orchestrator's response contains multiple Bash tool calls within a single
assistant turn. Each call has its own `description` and timeout. Outputs
return in parallel and can be aggregated for the verification report.

This is the structural pattern; the canonical 7-item example lives in
`.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution to
satisfy AC-WO-007.

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution
  (HARD batching obligation + 7-item canonical example).
- SPEC-V3R5-WORKFLOW-OPT-001 acceptance.md AC-WO-007 (verification command).
- W3 HARNESS-AUTONOMY-001 meta-analysis (`feedback_w3_metaanalysis_lessons.md`,
  source of the 10-min serial verification finding).

---

Version: 1.0.0
Classification: Evolvable operational rule, applies to all run-phase completion verifications
Origin: SPEC-V3R5-WORKFLOW-OPT-001 Layer D (2026-05-20)
