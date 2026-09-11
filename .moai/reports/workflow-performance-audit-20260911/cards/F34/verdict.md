# F34 — Language and monorepo routing

## Claim

Language routing now distinguishes Kotlin Gradle projects from Java, detects
all manifest/source candidates for monorepos, and records the candidate list
while retaining a bounded primary fast check.

## Evidence

Commands:

```text
bash .claude/hooks/tests/test-language-routing-contract.sh
node --check .claude/workflows/sync-audit-4dim.js
```

Observed output:

```text
PASS: language routing distinguishes Kotlin and preserves monorepo candidates
```

## Baseline-attribution

The fixture test ran in `WT-workflow-audit-f34b` after fast-forwarding to the
F33 merge on local `develop`. The JavaScript syntax check was also run in this
worktree.

## Gaps

No real mixed-language repository was sent through the full sync fan-out, and
no toolchain dispatch timings were measured.

## Residual-risk

The Stop hook still executes one primary fast toolchain command; the full sync
orchestrator must honor the logged candidate list to avoid under-checking a
secondary language.
