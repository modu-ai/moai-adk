# Wave 0 PoC Result — SPEC-V3R4-CI-FASTTRACK-001

## Summary

- **Date**: 2026-05-17
- **T0 PR**: #969 (test/skip-marker-poc-cift-001 → main, CLOSED without merge)
- **PR URL**: https://github.com/modu-ai/moai-adk/pull/969
- **Verdict**: **PASS** — skip-marker pattern confirmed working on this runner config

## CodeQL Check-Run Name Observed

```
gh api repos/modu-ai/moai-adk/commits/main/check-runs --jq '.check_runs[].name' | grep -i codeql
```

Output: `Analyze (Go) (go)`

This is the job name `Analyze (Go)` + matrix `language: go` → emitted check-run name `Analyze (Go) (go)`.

Branch protection `contexts` entry: bare `CodeQL` (legacy workflow-name match).

**Resolution (design.md AD-002)**: The branch protection `CodeQL` entry uses legacy workflow-name match. The skip-marker job for codeql.yml must be named such that it triggers the workflow `CodeQL` — meaning the skip-marker must be in the codeql.yml workflow itself (name: CodeQL), not a separate workflow. The `analyze-skip-marker` job running within the `CodeQL` workflow will satisfy the branch protection `CodeQL` check via the workflow-name match, regardless of what name the job emits. This is the correct interpretation.

**Implication for T2**: The `analyze-skip-marker` job in `codeql.yml` should use `name: Analyze (Go)` with `strategy.matrix.language: [go]` to emit `Analyze (Go) (go)` (matching what the real analyze job emits), which will satisfy the CodeQL required check via the workflow-name match. The key fact is that ANY job running within the `CodeQL`-named workflow satisfies the `CodeQL` branch protection entry.

## PoC Test Result

```
gh pr checks 969 --json name,state -q '.[] | select(.name=="PoC Test") | {name, state}'
```

Output: `{"name":"PoC Test","state":"SUCCESS"}`

- `test-skip-marker` job ran with `PoC Test` name → SUCCESS
- `test` job was in `skipped` state (mutually exclusive `if:` guard worked)
- Branch protection `PoC Test` is NOT a required check — no impact on PR mergeability

## Community Discussion Reference

- https://github.com/orgs/community/discussions/13690 — canonical thread on skip-marker pattern

## Wave 1 Decision Gate

**PoC Test = SUCCESS → Proceed to Wave 1 (T1..T8)**
