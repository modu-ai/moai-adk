# SPEC-WEB-CONSOLE-018 — Progress

Card t1079. Baseline of record: 67.9% @ `WT-web-coverage 0314801c2` (2026-09-22, `go test ./internal/web/... -cover`).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-WEB-CONSOLE-018
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
baseline_of_record:
  figure: 67.9%
  tree: 0314801c2
  command: go test ./internal/web/... -cover
  measured_at: 2026-09-22
  profile: .moai/state/verify/t1079/cover-baseline.out
historical_figure_provenance: SPEC-WEB-CONSOLE-017 sync-audit F2 @ 9d4a20eae (t1051, cited for provenance only)
target: 85% (quality.yaml test_coverage_target)
arithmetic: 9414 stmts / 3023 zero / need +1611 covered
```

## §E.2 Run-phase Evidence

### M1 — baseline re-pin (card t1079, AC-001)

- Tree: `WT-web-coverage @ 853e0f5df` (plan commit included; parent `0314801c2`). Measured 2026-09-22.
- Command (env-scrubbed, single invocation, `-count=1` to defeat cache):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 ./internal/web/ -coverprofile=/tmp/t1079-baseline.out
  ```

- Verbatim output:

  ```
  ok  	github.com/modu-ai/moai-adk/internal/web	25.523s	coverage: 67.9% of statements
  ```

- Exit code: 0.
- Divergence vs pinned baseline of record (67.9% @ `0314801c2`): **0.0pt — within the 1pt gate; no re-baseline** (REQ-001 / AC-001 satisfied).
- Per-function baseline profile: `/tmp/t1079-baseline.out` (regenerable with the command above). Function-level confirmation of spec §A.2 shape observed via `go tool cover -func`: `Monitor 28.7%`, `gauge 30.4%`, `spark 0.0%`, `specDetail 0.0%`, `buildAttention 45.5%`, `loadGoals 19.0%`, `loadVerify 20.6%`, `panel 51.5%`, `saveCluster 56.2%`, `Kanban 54.7%`.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
