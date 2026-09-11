# F40 — Executable workflow trace evidence

## Claim

The former TRACE PROBE comments are now explicitly non-evidence, and an
executable append-only JSONL helper records stage/tool identity, parent id,
input hash, comparison cohort, cache status, retry cause, and measured duration.
A summary command derives tool call count and p50/p95 duration from recorded
rows.

## Evidence

Command:

```text
bash .claude/hooks/tests/test-trace-ledger.sh
```

Observed output:

```text
PASS: workflow trace ledger records identity, cache, retry, duration, and p50/p95 evidence
```

## Baseline-attribution

The contract test ran in `WT-workflow-audit-f40b` after fast-forwarding to the
F39 merge on local `develop`. It used a temporary ledger and verified three
JSONL rows, one tool call, cache-miss count, retry count, cohort identity, and
p50/p95 values.

## Gaps

No live project → plan → run → sync trace was captured in this card. The
orchestrator call sites still need to invoke the helper at each enabled phase.

## Residual-risk

The ledger is opt-in and fail-open. A missing row while tracing is enabled is an
observation gap; this change does not claim a latency reduction or a live
workflow p50/p95 baseline until real cohorts are recorded.
